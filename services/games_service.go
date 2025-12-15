package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	lib "github.com/turnforge/weewar/lib"
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

type BaseGamesService struct {
	Self GamesService // The actual implementation
}

func (s *BaseGamesService) ListMoves(ctx context.Context, req *v1.ListMovesRequest) (resp *v1.ListMovesResponse, err error) {
	return nil, nil
}

// ProcessMoves processes moves for an existing game.
// It validates and applies moves, then delegates persistence to SaveMoveGroup.
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
	var dmp lib.MoveProcessor
	err = dmp.ProcessMoves(rtGame, req.Moves)
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

	// Delegate persistence to SaveMoveGroup - backend handles atomicity
	err = s.Self.SaveMoveGroup(ctx, req.GameId, gameresp.State, moveGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to save move group: %w", err)
	}

	return resp, err
}

// GetOptionsAt returns all available options at a specific position
func (s *BaseGamesService) GetOptionsAt(ctx context.Context, req *v1.GetOptionsAtRequest) (*v1.GetOptionsAtResponse, error) {
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

	var options []*v1.GameOption
	var allPaths *v1.AllPaths

	// Check what's at this position
	unit := rtGame.World.UnitAt(AxialCoord{Q: int(req.Q), R: int(req.R)})

	// Lazy top-up: If this is a unit, ensure it's refreshed for the current turn
	if unit != nil {
		if err := rtGame.TopUpUnitIfNeeded(unit); err != nil {
			return &v1.GetOptionsAtResponse{
				Options:         []*v1.GameOption{},
				CurrentPlayer:   gameresp.State.CurrentPlayer,
				GameInitialized: false,
			}, fmt.Errorf("failed to top-up unit: %w", err)
		}
	}

	// Check if there's a tile at this position and get its actions
	tile := rtGame.World.TileAt(AxialCoord{Q: int(req.Q), R: int(req.R)})
	if tile != nil {
		// Lazy top-up: Ensure tile is refreshed for the current turn
		if err := rtGame.TopUpTileIfNeeded(tile); err != nil {
			return &v1.GetOptionsAtResponse{
				Options:         []*v1.GameOption{},
				CurrentPlayer:   gameresp.State.CurrentPlayer,
				GameInitialized: false,
			}, fmt.Errorf("failed to top-up tile: %w", err)
		}

		// Only check tile actions if tile belongs to current player
		if tile.Player == rtGame.CurrentPlayer {
			// Get terrain definition for tile-specific actions
			terrainDef, err := rtGame.RulesEngine.GetTerrainData(tile.TileType)
			if err == nil {
				// Get current player's coins
				playerCoins := int32(0)
				for _, player := range rtGame.Config.Players {
					if player.PlayerId == rtGame.CurrentPlayer {
						playerCoins = player.Coins
						break
					}
				}

				// Get allowed actions for this tile
				tileActions := rtGame.RulesEngine.GetAllowedActionsForTile(tile, terrainDef, playerCoins)

				// Generate options based on allowed tile actions
				for _, action := range tileActions {
					switch action {
					case "build":
						// Filter buildable units by game's allowed units setting
						buildableUnits := FilterBuildOptionsByAllowedUnits(
							terrainDef.BuildableUnitIds,
							rtGame.Config.Settings.GetAllowedUnits(),
						)

						// Generate build unit options from filtered buildable units
						for _, unitTypeID := range buildableUnits {
							// Get unit definition to retrieve cost
							unitDef, err := rtGame.RulesEngine.GetUnitData(unitTypeID)
							if err != nil {
								continue // Skip if we can't get unit definition
							}

							// Only show units the player can afford
							if unitDef.Coins <= playerCoins {
								options = append(options, &v1.GameOption{
									OptionType: &v1.GameOption_Build{
										Build: &v1.BuildUnitAction{
											Q:        req.Q,
											R:        req.R,
											UnitType: unitTypeID,
											Cost:     unitDef.Coins,
										},
									},
								})
							}
						}
					}
				}
			}
		}
	}

	if unit != nil {
		// Our unit - get available options based on action progression
		var dmp lib.MoveProcessor

		// Get unit definition for progression rules
		unitDef, err := rtGame.RulesEngine.GetUnitData(unit.UnitType)
		if err != nil {
			// If we can't get unit def, default to all actions
			unitDef = &v1.UnitDefinition{
				ActionOrder: []string{"move", "attack|capture"},
			}
		}

		// Get allowed actions based on progression state
		allowedActions := rtGame.RulesEngine.GetAllowedActionsForUnit(unit, unitDef)

		// Check if "move" is allowed at current progression step
		moveAllowed := lib.ContainsAction(allowedActions, "move")

		// Get movement options if unit has movement left and move is allowed
		if unit.AvailableHealth > 0 && unit.DistanceLeft > 0 && moveAllowed {
			pathsResult, err := dmp.GetMovementOptions(rtGame, req.Q, req.R, false)
			if err == nil {
				allPaths = pathsResult

				// Create move options from AllPaths
				for key, edge := range allPaths.Edges {
					// Skip occupied tiles - can pass through but not land on them
					if edge.IsOccupied {
						continue
					}

					path, err := ReconstructPath(allPaths, edge.ToQ, edge.ToR)
					if err != nil {
						panic(err)
					}

					// Create ready-to-use MoveUnitAction
					moveAction := &v1.MoveUnitAction{
						FromQ:             req.Q,
						FromR:             req.R,
						ToQ:               edge.ToQ,
						ToR:               edge.ToR,
						MovementCost:      edge.TotalCost,
						ReconstructedPath: path,
					}

					options = append(options, &v1.GameOption{
						OptionType: &v1.GameOption_Move{Move: moveAction},
					})
					_ = key // Using key just to avoid unused variable warning
				}
			}
		}

		// Check if "attack" is allowed at current progression step
		attackAllowed := lib.ContainsAction(allowedActions, "attack")

		// Get attack options if unit can attack and attack is allowed
		if unit.AvailableHealth > 0 && attackAllowed {
			attackCoords, err := dmp.GetAttackOptions(rtGame, req.Q, req.R)
			if err == nil {
				for _, coord := range attackCoords {
					// Get target unit info for rich attack option data
					targetUnit := rtGame.World.UnitAt(coord)
					if targetUnit != nil {
						// Calculate estimated damage (simplified for now)
						damageEstimate := int32(50) // TODO: Use proper damage calculation from rules engine

						// Create ready-to-use AttackUnitAction
						attackAction := &v1.AttackUnitAction{
							AttackerQ:        req.Q,
							AttackerR:        req.R,
							DefenderQ:        int32(coord.Q),
							DefenderR:        int32(coord.R),
							TargetUnitType:   targetUnit.UnitType,
							TargetUnitHealth: targetUnit.AvailableHealth,
							CanAttack:        true,
							DamageEstimate:   damageEstimate,
						}

						options = append(options, &v1.GameOption{
							OptionType: &v1.GameOption_Attack{Attack: attackAction},
						})
					}
				}
			}
		}

		// TODO: Add capture building options if "capture" is allowed
		// TODO: Add build unit options if "build" is allowed
	}

	// Note: End turn is always available as a global action, so we don't include it in tile-specific options

	// Sort it for convinience too
	sort.Slice(options, func(i, j int) bool {
		return lib.GameOptionLess(options[i], options[j])
	})

	return &v1.GetOptionsAtResponse{
		Options:         options,
		CurrentPlayer:   rtGame.CurrentPlayer,
		GameInitialized: rtGame != nil && rtGame.World != nil,
		AllPaths:        allPaths,
	}, nil
}

func (b *BaseGamesService) ApplyChangeResults(changes []*v1.GameMove, rtGame *lib.Game, game *v1.Game, state *v1.GameState) error {

	// TRANSACTIONAL FIX: Temporary rollback to original world for ordered application
	if parent := rtGame.World.Pop(); parent != nil {
		rtGame.World = parent // Switch back to original world
	}

	// Apply each change to runtime game (now the original, not the transaction snapshot)
	for _, moveResult := range changes {
		for _, change := range moveResult.Changes {
			err := b.applyWorldChange(change, rtGame, state)
			if err != nil {
				return fmt.Errorf("failed to apply world change: %w", err)
			}
		}
	}

	state.WorldData = b.convertRuntimeWorldToProto(rtGame.World)
	state.UpdatedAt = timestamppb.New(time.Now())

	return nil
}

// applyWorldChange applies a single WorldChange to both runtime game and protobuf state
func (b *BaseGamesService) applyWorldChange(change *v1.WorldChange, rtGame *lib.Game, state *v1.GameState) error {
	switch changeType := change.ChangeType.(type) {
	case *v1.WorldChange_UnitMoved:
		return b.applyUnitMoved(changeType.UnitMoved, rtGame)
	case *v1.WorldChange_UnitDamaged:
		return b.applyUnitDamaged(changeType.UnitDamaged, rtGame)
	case *v1.WorldChange_UnitKilled:
		return b.applyUnitKilled(changeType.UnitKilled, rtGame)
	case *v1.WorldChange_PlayerChanged:
		return b.applyPlayerChanged(changeType.PlayerChanged, rtGame, state)
	case *v1.WorldChange_UnitBuilt:
		return b.applyUnitBuilt(changeType.UnitBuilt, rtGame)
	case *v1.WorldChange_CoinsChanged:
		return b.applyCoinsChanged(changeType.CoinsChanged, rtGame)
	default:
		return fmt.Errorf("unknown world change type")
	}
}

// applyUnitMoved moves a unit in the runtime game
func (b *BaseGamesService) applyUnitMoved(change *v1.UnitMovedChange, rtGame *lib.Game) error {
	if change.PreviousUnit == nil || change.UpdatedUnit == nil {
		return fmt.Errorf("missing unit data in UnitMovedChange")
	}

	fromCoord := AxialCoord{Q: int(change.PreviousUnit.Q), R: int(change.PreviousUnit.R)}
	toCoord := AxialCoord{Q: int(change.UpdatedUnit.Q), R: int(change.UpdatedUnit.R)}

	// Move unit in runtime game
	unit := rtGame.World.UnitAt(fromCoord)
	if unit == nil {
		return fmt.Errorf("unit not found at %v", fromCoord)
	}

	// Update unit with complete state from the change
	unit.AvailableHealth = change.UpdatedUnit.AvailableHealth
	unit.DistanceLeft = change.UpdatedUnit.DistanceLeft
	unit.LastActedTurn = change.UpdatedUnit.LastActedTurn
	unit.LastToppedupTurn = change.UpdatedUnit.LastToppedupTurn

	// Remove from old position and add to new position
	return rtGame.World.MoveUnit(unit, toCoord)
}

// applyUnitDamaged updates unit health in the runtime game
func (b *BaseGamesService) applyUnitDamaged(change *v1.UnitDamagedChange, rtGame *lib.Game) error {
	if change.UpdatedUnit == nil {
		return fmt.Errorf("missing updated unit data in UnitDamagedChange")
	}

	coord := AxialCoord{Q: int(change.UpdatedUnit.Q), R: int(change.UpdatedUnit.R)}

	unit := rtGame.World.UnitAt(coord)
	if unit == nil {
		return fmt.Errorf("unit not found at %v", coord)
	}

	// Update unit with complete state from the change
	unit.AvailableHealth = change.UpdatedUnit.AvailableHealth
	unit.DistanceLeft = change.UpdatedUnit.DistanceLeft
	unit.LastActedTurn = change.UpdatedUnit.LastActedTurn
	unit.LastToppedupTurn = change.UpdatedUnit.LastToppedupTurn
	return nil
}

// applyUnitKilled removes a unit from the runtime game
func (b *BaseGamesService) applyUnitKilled(change *v1.UnitKilledChange, rtGame *lib.Game) error {
	if change.PreviousUnit == nil {
		return fmt.Errorf("missing previous unit data in UnitKilledChange")
	}

	coord := AxialCoord{Q: int(change.PreviousUnit.Q), R: int(change.PreviousUnit.R)}
	unit := rtGame.World.UnitAt(coord)

	err := rtGame.World.RemoveUnit(unit)
	if err != nil {
		return fmt.Errorf("unit not found at %v", coord)
	}
	return nil
}

// applyPlayerChanged updates game state for turn/player changes
func (b *BaseGamesService) applyPlayerChanged(change *v1.PlayerChangedChange, rtGame *lib.Game, state *v1.GameState) error {
	rtGame.CurrentPlayer = change.NewPlayer
	rtGame.TurnCounter = change.NewTurn

	// Also update the protobuf GameState
	state.CurrentPlayer = change.NewPlayer
	state.TurnCounter = change.NewTurn

	return nil
}

// applyUnitBuilt adds a newly built unit to the runtime game
func (b *BaseGamesService) applyUnitBuilt(change *v1.UnitBuiltChange, rtGame *lib.Game) error {
	if change.Unit == nil {
		return fmt.Errorf("missing unit data in UnitBuiltChange")
	}

	// Add the new unit to the runtime game
	rtGame.World.AddUnit(change.Unit)

	// Update tile's last acted turn
	coord := AxialCoord{Q: int(change.TileQ), R: int(change.TileR)}
	tile := rtGame.World.TileAt(coord)
	if tile != nil {
		tile.LastActedTurn = rtGame.TurnCounter
	}

	return nil
}

// applyCoinsChanged updates a player's coin balance in the runtime game
func (b *BaseGamesService) applyCoinsChanged(change *v1.CoinsChangedChange, rtGame *lib.Game) error {
	// Update player's coins in game config
	for i, player := range rtGame.Config.Players {
		if player.PlayerId == change.PlayerId {
			rtGame.Config.Players[i].Coins = change.NewCoins
			return nil
		}
	}
	return fmt.Errorf("player %d not found", change.PlayerId)
}

// convertRuntimeWorldToProto converts runtime world state to protobuf WorldData
// Since World now holds proto data directly, this just returns the underlying WorldData
func (b *BaseGamesService) convertRuntimeWorldToProto(world *lib.World) *v1.WorldData {
	return world.WorldData()
}

// FilterBuildOptionsByAllowedUnits filters buildable units by allowed units.
// If allowedUnits is nil, no filtering is applied (all units allowed).
// If allowedUnits is empty slice, no units are allowed.
func FilterBuildOptionsByAllowedUnits(buildableUnits, allowedUnits []int32) []int32 {
	// If allowedUnits is nil, no restriction - return all buildable units
	if allowedUnits == nil {
		return buildableUnits
	}

	// If allowedUnits is empty, nothing is allowed
	if len(allowedUnits) == 0 {
		return []int32{}
	}

	allowedSet := make(map[int32]bool)
	for _, u := range allowedUnits {
		allowedSet[u] = true
	}

	var filtered []int32
	for _, u := range buildableUnits {
		if allowedSet[u] {
			filtered = append(filtered, u)
		}
	}
	return filtered
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
