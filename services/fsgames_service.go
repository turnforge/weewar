package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	turnengine "github.com/panyam/turnengine/engine/gen/go/turnengine/v1"
	"github.com/panyam/turnengine/engine/storage"
	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
	"google.golang.org/protobuf/proto"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

var GAMES_STORAGE_DIR = ""

// FSGamesServiceImpl implements the GamesService gRPC interface
type FSGamesServiceImpl struct {
	BaseGamesServiceImpl
	WorldsService v1.WorldsServiceServer
	storage       *storage.FileStorage // Storage area for all files
	
	// Simple caches - maps with game ID as key
	gameCache     map[string]*v1.Game
	stateCache    map[string]*v1.GameState
	historyCache  map[string]*v1.GameMoveHistory
	runtimeCache  map[string]*weewar.Game
}

// NewGamesService creates a new GamesService implementation for server mode
func NewFSGamesService() *FSGamesServiceImpl {
	if GAMES_STORAGE_DIR == "" {
		GAMES_STORAGE_DIR = weewar.DevDataPath("storage/games")
	}
	service := &FSGamesServiceImpl{
		BaseGamesServiceImpl: BaseGamesServiceImpl{},
		WorldsService:        NewFSWorldsService(),
		storage:              storage.NewFileStorage(GAMES_STORAGE_DIR),
		gameCache:            make(map[string]*v1.Game),
		stateCache:           make(map[string]*v1.GameState),
		historyCache:         make(map[string]*v1.GameMoveHistory),
		runtimeCache:         make(map[string]*weewar.Game),
	}
	service.Self = service

	return service
}

// ListGames returns all available games (metadata only for performance)
func (s *FSGamesServiceImpl) ListGames(ctx context.Context, req *v1.ListGamesRequest) (resp *v1.ListGamesResponse, err error) {
	resp = &v1.ListGamesResponse{
		Items: []*v1.Game{},
		Pagination: &v1.PaginationResponse{
			HasMore:      false,
			TotalResults: 0,
		},
	}
	resp.Items, err = storage.ListFSEntities[*v1.Game](s.storage, nil)
	resp.Pagination.TotalResults = int32(len(resp.Items))
	return resp, nil
}

// DeleteGame deletes a game
func (s *FSGamesServiceImpl) DeleteGame(ctx context.Context, req *v1.DeleteGameRequest) (resp *v1.DeleteGameResponse, err error) {
	resp = &v1.DeleteGameResponse{}
	err = s.storage.DeleteEntity(req.Id)
	return
}

// CreateWorld creates a new world
func (s *FSGamesServiceImpl) CreateGame(ctx context.Context, req *v1.CreateGameRequest) (resp *v1.CreateGameResponse, err error) {
	if req.Game == nil {
		return nil, fmt.Errorf("game data is required")
	}

	req.Game.Id, err = s.storage.CreateEntity(req.Game.Id)
	if err != nil {
		return resp, err
	}

	now := time.Now()
	req.Game.CreatedAt = tspb.New(now)
	req.Game.UpdatedAt = tspb.New(now)

	// Save game metadta
	if err := s.storage.SaveArtifact(req.Game.Id, "metadata", req.Game); err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	world, err := s.WorldsService.GetWorld(ctx, &v1.GetWorldRequest{Id: req.Game.WorldId})
	if err != nil {
		return nil, fmt.Errorf("Error loading world: %w", err)
	}

	// Save a new empty game state and a new move list
	gs := &v1.GameState{
		GameId:        req.Game.Id,
		CurrentPlayer: 1, // Game starts with player 1
		TurnCounter:   1, // First turn
		WorldData:     world.WorldData,
	}

	// Initialize units with default stats from rules engine for new games
	if gs.WorldData != nil && gs.WorldData.Units != nil {
		rulesEngine := weewar.DefaultRulesEngine()
		for _, unit := range gs.WorldData.Units {
			// Get unit defaults from rules engine
			unitData, err := rulesEngine.GetUnitData(unit.UnitType)
			if err != nil {
				log.Printf("Warning: failed to get unit data for type %d: %v", unit.UnitType, err)
				continue // Skip this unit but don't fail the entire game creation
			}

			// Set default health and movement points for new game
			unit.AvailableHealth = unitData.Health
			unit.DistanceLeft = unitData.MovementPoints
			unit.TurnCounter = gs.TurnCounter
		}
	}
	if err := s.storage.SaveArtifact(req.Game.Id, "state", gs); err != nil {
		log.Printf("Failed to create state for game %s: %v", req.Game.Id, err)
	}

	// Save a new empty game history and a new move list
	if err := s.storage.SaveArtifact(req.Game.Id, "history", &v1.GameMoveHistory{GameId: req.Game.Id}); err != nil {
		log.Printf("Failed to create state for game %s: %v", req.Game.Id, err)
	}

	resp = &v1.CreateGameResponse{
		Game:      req.Game,
		GameState: gs,
	}

	return resp, nil
}

// GetGame returns a specific game with complete data including tiles and units
func (s *FSGamesServiceImpl) GetGame(ctx context.Context, req *v1.GetGameRequest) (resp *v1.GetGameResponse, err error) {
	if req.Id == "" {
		return nil, fmt.Errorf("game ID is required")
	}

	// Check cache first
	if game, ok := s.gameCache[req.Id]; ok {
		if state, ok := s.stateCache[req.Id]; ok {
			if history, ok := s.historyCache[req.Id]; ok {
				return &v1.GetGameResponse{
					Game:    game,
					State:   state,
					History: history,
				}, nil
			}
		}
	}

	// Load from disk
	game, err := storage.LoadFSArtifact[*v1.Game](s.storage, req.Id, "metadata")
	if err != nil {
		return nil, fmt.Errorf("game metadata not found: %w", err)
	}

	gameState, err := storage.LoadFSArtifact[*v1.GameState](s.storage, req.Id, "state")
	if err != nil {
		return nil, fmt.Errorf("game state not found: %w", err)
	}

	gameHistory, err := storage.LoadFSArtifact[*v1.GameMoveHistory](s.storage, req.Id, "history")
	if err != nil {
		return nil, fmt.Errorf("game state not found: %w", err)
	}

	// Cache everything
	s.gameCache[req.Id] = game
	s.stateCache[req.Id] = gameState
	s.historyCache[req.Id] = gameHistory

	resp = &v1.GetGameResponse{
		Game:    game,
		State:   gameState,
		History: gameHistory,
	}

	return resp, nil
}

// UpdateGame updates an existing game
func (s *FSGamesServiceImpl) UpdateGame(ctx context.Context, req *v1.UpdateGameRequest) (resp *v1.UpdateGameResponse, err error) {
	if req.GameId == "" {
		return nil, fmt.Errorf("game ID is required")
	}

	resp = &v1.UpdateGameResponse{}

	// Load existing metadata if updating
	if req.NewGame != nil {
		game, err := storage.LoadFSArtifact[*v1.Game](s.storage, req.GameId, "metadata")
		if err != nil {
			return nil, fmt.Errorf("game not found: %w", err)
		}

		// Update metadata fields
		if req.NewGame.Name != "" {
			game.Name = req.NewGame.Name
		}
		if req.NewGame.Description != "" {
			game.Description = req.NewGame.Description
		}
		if req.NewGame.Tags != nil {
			game.Tags = req.NewGame.Tags
		}
		if req.NewGame.Difficulty != "" {
			game.Difficulty = req.NewGame.Difficulty
		}
		game.UpdatedAt = tspb.New(time.Now())

		if err := s.storage.SaveArtifact(req.GameId, "metadata", game); err != nil {
			return nil, fmt.Errorf("failed to update game metadata: %w", err)
		}
		
		// Update cache
		s.gameCache[req.GameId] = game
		resp.Game = game
	}

	if req.NewState != nil {
		if err := s.storage.SaveArtifact(req.GameId, "state", req.NewState); err != nil {
			return nil, fmt.Errorf("failed to update game state: %w", err)
		}
		
		// Update cache and invalidate runtime game
		s.stateCache[req.GameId] = req.NewState
		delete(s.runtimeCache, req.GameId)
	}

	if req.NewHistory != nil {
		if err := s.storage.SaveArtifact(req.GameId, "history", req.NewHistory); err != nil {
			return nil, fmt.Errorf("failed to update game history: %w", err)
		}
		
		// Update cache
		s.historyCache[req.GameId] = req.NewHistory
	}

	return resp, err
}

// GetRuntimeGame implements the interface method (for compatibility)
func (s *FSGamesServiceImpl) GetRuntimeGame(game *v1.Game, gameState *v1.GameState) (*weewar.Game, error) {
	return ProtoToRuntimeGame(game, gameState), nil
}

// GetRuntimeGameByID returns a cached runtime game instance for the given game ID
func (s *FSGamesServiceImpl) GetRuntimeGameByID(ctx context.Context, gameID string) (*weewar.Game, error) {
	// Check runtime cache first
	if rtGame, ok := s.runtimeCache[gameID]; ok {
		return rtGame, nil
	}

	// Load proto data (will use cache if available)
	resp, err := s.GetGame(ctx, &v1.GetGameRequest{Id: gameID})
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Convert to runtime game
	rtGame := ProtoToRuntimeGame(resp.Game, resp.State)
	
	// Cache it
	s.runtimeCache[gameID] = rtGame
	
	return rtGame, nil
}

// Helper functions for serialization

// serialize converts a protobuf message to bytes
func serialize(msg proto.Message) []byte {
	if msg == nil {
		return nil
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("Failed to serialize: %v", err)
		return nil
	}
	return data
}

// deserialize converts bytes back to a protobuf message
func deserialize(data []byte, msg proto.Message) error {
	if len(data) == 0 {
		return fmt.Errorf("empty data")
	}
	return proto.Unmarshal(data, msg)
}

// computeHash generates a SHA256 hash of any protobuf message
func computeHash(msg proto.Message) string {
	if msg == nil {
		return ""
	}
	data := serialize(msg)
	if data == nil {
		return ""
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (s *FSGamesServiceImpl) ProcessMoves(ctx context.Context, req *v1.ProcessMovesRequest) (*v1.ProcessMovesResponse, error) {
	// If client didn't provide expected results, run ProcessMoves locally
	if req.ExpectedResponse == nil {
		return s.BaseGamesServiceImpl.ProcessMoves(ctx, req)
	}

	// Client provided expected results - validate through coordinator

	// Get current game state
	gameresp, err := s.Self.GetGame(ctx, &v1.GetGameRequest{Id: req.GameId})
	if err != nil || gameresp.Game == nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}
	if gameresp.State == nil {
		return nil, fmt.Errorf("game state cannot be nil")
	}

	// Apply expected changes to current state to get new state
	// This is what validators will do independently
	newState := proto.Clone(gameresp.State).(*v1.GameState)
	// TODO: Apply req.ExpectedChanges to newState

	// Compute hashes (validators will compute same hashes)
	fromStateHash := computeHash(gameresp.State)
	toStateHash := computeHash(newState)

	// Serialize for coordinator (game-agnostic blobs)
	movesBlob := serialize(&v1.ProcessMovesRequest{
		GameId: req.GameId,
		Moves:  req.Moves,
	})
	changesBlob := serialize(req.ExpectedResponse)
	newStateBlob := serialize(newState)

	// Create proposal for coordinator
	proposal := &turnengine.SubmitProposalRequest{
		SessionId:     req.GameId,
		ProposerId:    "player1", // TODO: Get from context/session
		FromStateHash: fromStateHash,
		ToStateHash:   toStateHash,
		MovesBlob:     movesBlob,
		ChangesBlob:   changesBlob,
		NewStateBlob:  newStateBlob,
	}

	// TODO: Submit to coordinator when wired up
	_ = proposal

	// For now, just apply the changes directly (simulating accepted proposal)
	return req.ExpectedResponse, nil
}

// Implement coordination.Callbacks interface

// OnProposalStarted is called when a proposal is accepted for validation
func (s *FSGamesServiceImpl) OnProposalStarted(gameID string, proposal *turnengine.ProposalInfo) error {
	// Load the game state
	gameState, err := storage.LoadFSArtifact[*v1.GameState](s.storage, gameID, "state")
	if err != nil {
		return fmt.Errorf("failed to load game state: %w", err)
	}

	// Set the proposal tracking info
	gameState.ProposalInfo = &turnengine.ProposalTrackingInfo{
		ProposalId:     proposal.ProposalId,
		ProposerId:     proposal.ProposerId,
		Phase:          turnengine.ProposalPhase_PROPOSAL_PHASE_COLLECTING,
		CreatedAt:      proposal.CreatedAt,
		ValidatorCount: int32(len(proposal.AssignedValidators)),
		VotesReceived:  0,
	}

	// Save the updated state
	return s.storage.SaveArtifact(gameID, "state", gameState)
}

// OnProposalAccepted is called when consensus approves the proposal
func (s *FSGamesServiceImpl) OnProposalAccepted(gameID string, proposal *turnengine.ProposalInfo) error {
	// The new state is in the proposal's new_state_blob
	// We need to save it as the new game state

	// Note: In a real implementation, we'd unmarshal proposal.NewStateBlob
	// For now, just clear the proposal info

	gameState, err := storage.LoadFSArtifact[*v1.GameState](s.storage, gameID, "state")
	if err != nil {
		return fmt.Errorf("failed to load game state: %w", err)
	}

	// Clear proposal info and update state hash
	gameState.ProposalInfo = nil
	gameState.StateHash = proposal.ToStateHash

	// Save the state
	return s.storage.SaveArtifact(gameID, "state", gameState)
}

// OnProposalFailed is called when proposal is rejected or times out
func (s *FSGamesServiceImpl) OnProposalFailed(gameID string, proposal *turnengine.ProposalInfo, reason string) error {
	// Clear the proposal info from game state
	gameState, err := storage.LoadFSArtifact[*v1.GameState](s.storage, gameID, "state")
	if err != nil {
		return fmt.Errorf("failed to load game state: %w", err)
	}

	// Clear proposal info
	gameState.ProposalInfo = nil

	// Save the state
	return s.storage.SaveArtifact(gameID, "state", gameState)
}
