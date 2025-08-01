package services

import (
	"context"
	"fmt"
	"log"
	"time"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

var GAMES_STORAGE_DIR = ""

// FSGamesServiceImpl implements the GamesService gRPC interface
type FSGamesServiceImpl struct {
	BaseGamesServiceImpl
	WorldsService v1.WorldsServiceServer
	storage       *FileStorage // Storage area for all files
}

// NewGamesService creates a new GamesService implementation for server mode
func NewFSGamesService() *FSGamesServiceImpl {
	if GAMES_STORAGE_DIR == "" {
		GAMES_STORAGE_DIR = weewar.DevDataPath("storage/games")
	}
	service := &FSGamesServiceImpl{
		BaseGamesServiceImpl: BaseGamesServiceImpl{},
		WorldsService:        NewFSWorldsService(),
		storage:              NewFileStorage(GAMES_STORAGE_DIR),
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
	resp.Items, err = ListFSEntities[*v1.Game](s.storage, nil)
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

	game, err := LoadFSArtifact[*v1.Game](s.storage, req.Id, "metadata")
	if err != nil {
		return nil, fmt.Errorf("game metadata not found: %w", err)
	}

	gameState, err := LoadFSArtifact[*v1.GameState](s.storage, req.Id, "state")
	if err != nil {
		return nil, fmt.Errorf("game state not found: %w", err)
	}

	gameHistory, err := LoadFSArtifact[*v1.GameMoveHistory](s.storage, req.Id, "history")
	if err != nil {
		return nil, fmt.Errorf("game state not found: %w", err)
	}

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

	// Load existing metadata
	if req.NewGame != nil {
		game, err := LoadFSArtifact[*v1.Game](s.storage, req.GameId, "metadata")
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

		if err := s.storage.SaveArtifact(req.NewGame.Id, "metadata", game); err != nil {
			return nil, fmt.Errorf("failed to update game metadata: %w", err)
		}
	}

	if req.NewState != nil {
		if err := s.storage.SaveArtifact(req.GameId, "state", req.NewState); err != nil {
			return nil, fmt.Errorf("failed to update game state: %w", err)
		}
	}

	if req.NewHistory != nil {
		if err := s.storage.SaveArtifact(req.GameId, "history", req.NewHistory); err != nil {
			return nil, fmt.Errorf("failed to update game history: %w", err)
		}
	}

	return resp, err
}

func (w *FSGamesServiceImpl) GetRuntimeGame(game *v1.Game, gameState *v1.GameState) (out *weewar.Game, err error) {
	return ProtoToRuntimeGame(game, gameState)
}
