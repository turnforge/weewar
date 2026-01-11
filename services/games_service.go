package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	lib "github.com/turnforge/weewar/lib"
	"github.com/turnforge/weewar/services/authz"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GamesService interface {
	// Create a new game
	CreateGame(context.Context, *v1.CreateGameRequest) (*v1.CreateGameResponse, error)
	// *
	// Batch get multiple games by ID
	GetGames(context.Context, *v1.GetGamesRequest) (*v1.GetGamesResponse, error)
	// ListGames returns all available games
	ListGames(context.Context, *v1.ListGamesRequest) (*v1.ListGamesResponse, error)
	// GetGame returns a specific game with metadata
	GetGame(context.Context, *v1.GetGameRequest) (*v1.GetGameResponse, error)
	// *
	// Delete a particular game
	DeleteGame(context.Context, *v1.DeleteGameRequest) (*v1.DeleteGameResponse, error)
	// GetGame returns a specific game with metadata
	UpdateGame(context.Context, *v1.UpdateGameRequest) (*v1.UpdateGameResponse, error)
	// Gets the latest game state
	GetGameState(context.Context, *v1.GetGameStateRequest) (*v1.GetGameStateResponse, error)
	// List the moves for a game
	ListMoves(context.Context, *v1.ListMovesRequest) (*v1.ListMovesResponse, error)
	ProcessMoves(context.Context, *v1.ProcessMovesRequest) (*v1.ProcessMovesResponse, error)
	GetOptionsAt(context.Context, *v1.GetOptionsAtRequest) (*v1.GetOptionsAtResponse, error)
	// *
	// Simulates combat between two units to generate damage distributions
	// This is a stateless utility method that doesn't require game state
	SimulateAttack(context.Context, *v1.SimulateAttackRequest) (*v1.SimulateAttackResponse, error)
	GetRuntimeGame(game *v1.Game, gameState *v1.GameState) (*lib.Game, error)

	// SaveMoveGroup saves a move group atomically with the game state.
	// Each backend implements this with appropriate transactionality:
	// - GORM: uses database transaction
	// - FS: writes history then state (pseudo-atomic)
	// - Non-transactional: writes moves first, then commits via state update (checkpoint pattern)
	SaveMoveGroup(ctx context.Context, gameId string, state *v1.GameState, group *v1.GameMoveGroup) error
}

// MovesSavedCallback is called after moves are saved.
// Used by BackendGamesService to broadcast to sync subscribers.
type MovesSavedCallback func(ctx context.Context, gameId string, moves []*v1.GameMove, groupNumber int64)

type BaseGamesService struct {
	Self          GamesService // The actual implementation
	OnMovesSaved  MovesSavedCallback
}

func (s *BaseGamesService) ListMoves(ctx context.Context, req *v1.ListMovesRequest) (resp *v1.ListMovesResponse, err error) {
	return nil, nil
}

// ProcessMoves processes moves for an existing game.
// It validates and applies moves, then delegates persistence to SaveMoveGroup.
// Authorization: User must be a player in the game AND it must be their turn.
func (s *BaseGamesService) ProcessMoves(ctx context.Context, req *v1.ProcessMovesRequest) (resp *v1.ProcessMovesResponse, err error) {
	if len(req.Moves) == 0 {
		return nil, fmt.Errorf("at least one move is required")
	}

	gameresp, err := s.Self.GetGame(ctx, &v1.GetGameRequest{Id: req.GameId})
	if err != nil || gameresp.Game == nil {
		return nil, err
	}
	if gameresp.State == nil {
		return nil, fmt.Errorf("game state cannot be nil")
	}

	// Authorization: user must be a player in the game AND it must be their turn
	if err := authz.CanSubmitMoves(ctx, gameresp.Game, gameresp.State.CurrentPlayer); err != nil {
		return nil, err
	}

	// Get the runtime game corresponding to this game Id
	rtGame, err := s.Self.GetRuntimeGame(gameresp.Game, gameresp.State)
	if err != nil {
		return nil, err
	}

	// TRANSACTIONAL FIX: Create transaction snapshot for move processing
	// ProcessMoves will operate on the snapshot, ApplyChangeResults will apply to original
	originalWorld := rtGame.World
	rtGame.World = originalWorld.Push() // Create transaction layer

	// Validate and process moves in transaction layer
	err = rtGame.ProcessMoves(req.Moves)
	if err != nil {
		return nil, err
	}
	resp = &v1.ProcessMovesResponse{Moves: req.Moves}

	// Increment group number for this batch
	nextGroupNumber := gameresp.State.CurrentGroupNumber + 1

	// Create a new move group to track this batch of processed moves
	startTime := time.Now()
	moveGroup := &v1.GameMoveGroup{
		StartedAt:   timestamppb.New(startTime),
		EndedAt:     timestamppb.New(startTime),
		Moves:       req.Moves,
		GroupNumber: nextGroupNumber,
	}

	// Apply the changes to update gamestate
	s.ApplyChangeResults(req.Moves, rtGame, gameresp.Game, gameresp.State)

	// Update state with new group number (this is the "commit marker")
	gameresp.State.CurrentGroupNumber = nextGroupNumber

	// Update the end time after processing is complete
	moveGroup.EndedAt = timestamppb.New(time.Now())

	// Skip persistence in dry run mode
	if req.DryRun {
		return resp, nil
	}

	// Delegate persistence to SaveMoveGroup - backend handles atomicity
	err = s.Self.SaveMoveGroup(ctx, req.GameId, gameresp.State, moveGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to save move group: %w", err)
	}

	// Broadcast to sync subscribers (multiplayer)
	if s.OnMovesSaved != nil {
		s.OnMovesSaved(ctx, req.GameId, req.Moves, nextGroupNumber)
	}

	return resp, err
}

// GetOptionsAt returns all available options at a specific position
func (s *BaseGamesService) GetOptionsAt(ctx context.Context, req *v1.GetOptionsAtRequest) (out *v1.GetOptionsAtResponse, err error) {
	// Load game data using the service implementation
	gameresp, err := s.Self.GetGame(ctx, &v1.GetGameRequest{Id: req.GameId})
	if err != nil || gameresp.Game == nil {
		return &v1.GetOptionsAtResponse{
			Options:         []*v1.GameOption{},
			CurrentPlayer:   0,
			GameInitialized: false,
		}, nil
	}
	if gameresp.State == nil {
		return &v1.GetOptionsAtResponse{
			Options:         []*v1.GameOption{},
			CurrentPlayer:   0,
			GameInitialized: false,
		}, nil
	}

	// Get the runtime game
	rtGame, err := s.Self.GetRuntimeGame(gameresp.Game, gameresp.State)
	if err != nil {
		return &v1.GetOptionsAtResponse{
			Options:         []*v1.GameOption{},
			CurrentPlayer:   gameresp.State.CurrentPlayer,
			GameInitialized: false,
		}, nil
	}

	// Delegate to lib.Game.GetOptionsAt
	posLabel := ""
	if req.Pos != nil {
		if req.Pos.Label != "" {
			posLabel = req.Pos.Label
		} else {
			posLabel = fmt.Sprintf("%d,%d", req.Pos.Q, req.Pos.R)
		}
	}

	out, err = rtGame.GetOptionsAt(posLabel)
	if err != nil {
		return out, err
	}

	// Sort options for convenience
	sort.Slice(out.Options, func(i, j int) bool {
		return lib.GameOptionLess(out.Options[i], out.Options[j])
	})

	return
}

func (b *BaseGamesService) ApplyChangeResults(changes []*v1.GameMove, rtGame *lib.Game, game *v1.Game, state *v1.GameState) error {
	// Apply changes to the runtime game
	if err := rtGame.ApplyChanges(changes); err != nil {
		return err
	}

	// Update proto state with new world data and timestamp
	state.WorldData = rtGame.World.WorldData() // b.convertRuntimeWorldToProto(rtGame.World)
	state.UpdatedAt = timestamppb.New(time.Now())

	return nil
}

// SimulateAttack simulates combat between two units and returns damage distributions
func (s *BaseGamesService) SimulateAttack(ctx context.Context, req *v1.SimulateAttackRequest) (resp *v1.SimulateAttackResponse, err error) {
	resp = &v1.SimulateAttackResponse{}

	// Set default number of simulations if not provided
	numSims := req.NumSimulations
	if numSims <= 0 {
		numSims = 1000
	}

	// Create mock units and tiles for simulation
	attackerUnit := &v1.Unit{
		Q:               0,
		R:               0,
		Player:          1,
		UnitType:        req.AttackerUnitType,
		AvailableHealth: req.AttackerHealth,
	}
	attackerTile := &v1.Tile{
		Q:        0,
		R:        0,
		TileType: req.AttackerTerrain,
	}

	defenderUnit := &v1.Unit{
		Q:               1,
		R:               0,
		Player:          2,
		UnitType:        req.DefenderUnitType,
		AvailableHealth: req.DefenderHealth,
	}
	defenderTile := &v1.Tile{
		Q:        1,
		R:        0,
		TileType: req.DefenderTerrain,
	}

	// Get rules engine
	rulesEngine := lib.DefaultRulesEngine()

	// Simulate attacker -> defender
	attackerCtx := &lib.CombatContext{
		Attacker:       attackerUnit,
		AttackerTile:   attackerTile,
		AttackerHealth: req.AttackerHealth,
		Defender:       defenderUnit,
		DefenderTile:   defenderTile,
		DefenderHealth: req.DefenderHealth,
		WoundBonus:     req.WoundBonus,
	}

	attackerDist, err := rulesEngine.GenerateDamageDistribution(attackerCtx, int(numSims))
	if err != nil {
		return nil, fmt.Errorf("failed to generate attacker damage distribution: %w", err)
	}

	// Convert attacker distribution to map
	attackerDamageMap := make(map[int32]int32)
	attackerMeanDamage := 0.0
	attackerKillCount := int32(0)
	for _, dmgRange := range attackerDist.Ranges {
		damage := int32(dmgRange.MinValue)
		count := int32(float64(numSims) * dmgRange.Probability)
		attackerDamageMap[damage] = count
		attackerMeanDamage += dmgRange.MinValue * dmgRange.Probability
		if damage >= req.DefenderHealth {
			attackerKillCount += count
		}
	}

	// Simulate defender -> attacker (counter-attack)
	defenderCtx := &lib.CombatContext{
		Attacker:       defenderUnit,
		AttackerTile:   defenderTile,
		AttackerHealth: req.DefenderHealth,
		Defender:       attackerUnit,
		DefenderTile:   attackerTile,
		DefenderHealth: req.AttackerHealth,
		WoundBonus:     0, // No wound bonus for counter-attack
	}

	defenderDist, err := rulesEngine.GenerateDamageDistribution(defenderCtx, int(numSims))
	defenderDamageMap := make(map[int32]int32)
	defenderMeanDamage := 0.0
	defenderKillCount := int32(0)

	if err == nil && defenderDist != nil {
		for _, dmgRange := range defenderDist.Ranges {
			damage := int32(dmgRange.MinValue)
			count := int32(float64(numSims) * dmgRange.Probability)
			defenderDamageMap[damage] = count
			defenderMeanDamage += dmgRange.MinValue * dmgRange.Probability
			if damage >= req.AttackerHealth {
				defenderKillCount += count
			}
		}
	}

	resp.AttackerDamageDistribution = attackerDamageMap
	resp.DefenderDamageDistribution = defenderDamageMap
	resp.AttackerMeanDamage = attackerMeanDamage
	resp.DefenderMeanDamage = defenderMeanDamage
	resp.AttackerKillProbability = float64(attackerKillCount) / float64(numSims)
	resp.DefenderKillProbability = float64(defenderKillCount) / float64(numSims)

	return resp, nil
}
