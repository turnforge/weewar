//go:build !wasm
// +build !wasm

package gormbe

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	gfn "github.com/panyam/goutils/fn"
	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	v1gorm "github.com/turnforge/weewar/gen/gorm"
	v1dal "github.com/turnforge/weewar/gen/gorm/dal"
	"github.com/turnforge/weewar/lib"
	"github.com/turnforge/weewar/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// GamesService implements the GamesService gRPC interface
type GamesService struct {
	services.BackendGamesService
	storage      *gorm.DB
	MaxPageSize  int
	GameDAL      v1dal.GameGORMDAL
	GameStateDAL v1dal.GameStateGORMDAL
	GameMoveDAL  v1dal.GameMoveGORMDAL
}

// NewGamesService creates a new GamesService implementation
func NewGamesService(db *gorm.DB, clientMgr *services.ClientMgr) *GamesService {
	// db.AutoMigrate(&v1gorm.IndexRecordsLROGORM{})
	db.AutoMigrate(&v1gorm.GameGORM{})
	db.AutoMigrate(&v1gorm.GameStateGORM{})
	db.AutoMigrate(&v1gorm.GameMoveGORM{})

	service := &GamesService{
		storage:     db,
		MaxPageSize: 1000,
	}
	service.ClientMgr = clientMgr
	service.GameDAL.WillCreate = func(ctx context.Context, game *v1gorm.GameGORM) error {
		game.UpdatedAt = time.Now()
		game.CreatedAt = time.Now()
		return nil
	}
	service.Self = service
	service.GameStateUpdater = service // Implement GameStateUpdater interface
	service.InitializeScreenshotIndexer()

	return service
}

// GetGameStateVersion implements GameStateUpdater interface
func (s *GamesService) GetGameStateVersion(ctx context.Context, id string) (int64, error) {
	gameState, err := s.GameStateDAL.Get(ctx, s.storage, id)
	if err != nil {
		return 0, err
	}
	return gameState.Version, nil
}

// UpdateGameStateScreenshotIndexInfo implements GameStateUpdater interface
// Note: This does NOT increment version - IndexInfo is internal bookkeeping
// that shouldn't invalidate user's optimistic lock
func (s *GamesService) UpdateGameStateScreenshotIndexInfo(ctx context.Context, id string, oldVersion int64, lastIndexedAt time.Time, needsIndexing bool) error {
	gameState, err := s.GameStateDAL.Get(ctx, s.storage, id)
	if err != nil {
		return err
	}

	// Check version matches - if not, content was updated and we'll re-index later
	if gameState.Version != oldVersion {
		return fmt.Errorf("version mismatch - content was updated, will re-index later")
	}

	// Update only IndexInfo fields, don't touch version
	gameState.WorldData.ScreenshotIndexInfo.LastIndexedAt = lastIndexedAt
	gameState.WorldData.ScreenshotIndexInfo.NeedsIndexing = needsIndexing
	// Note: NOT incrementing version - this is internal bookkeeping

	// Save - use version check to ensure we're updating the right state
	err = s.GameStateDAL.Save(ctx, s.storage.Where("game_id = ? AND version = ?", id, oldVersion), gameState)
	if err != nil {
		return fmt.Errorf("failed to update IndexInfo: %w", err)
	}
	return nil
}

// ListGames returns all available games (metadata only for performance)
func (s *GamesService) ListGames(ctx context.Context, req *v1.ListGamesRequest) (resp *v1.ListGamesResponse, err error) {
	ctx, span := Tracer.Start(ctx, "ListGames")
	defer span.End()
	resp = &v1.ListGamesResponse{
		Items: []*v1.Game{},
		Pagination: &v1.PaginationResponse{
			HasMore:      false,
			TotalResults: 0,
		},
	}
	games, err := s.GameDAL.List(ctx, s.storage.Order("updated_at desc").Order("name asc"))
	if err != nil {
		return
	}
	resp.Items = gfn.Map(games, func(g *v1gorm.GameGORM) *v1.Game {
		out, _ := v1gorm.GameFromGameGORM(nil, g, nil)
		if len(out.PreviewUrls) == 0 {
			out.PreviewUrls = []string{fmt.Sprintf("/screenshots/games/%s/default.png", out.Id)}
		}
		return out
	})
	resp.Pagination.TotalResults = int32(len(resp.Items))

	return resp, nil
}

// GetGame returns a specific game with complete data including tiles and units
func (s *GamesService) GetGame(ctx context.Context, req *v1.GetGameRequest) (resp *v1.GetGameResponse, err error) {
	if req.Id == "" {
		return nil, fmt.Errorf("game ID is required")
	}

	resp = &v1.GetGameResponse{}
	ctx, span := Tracer.Start(ctx, "GetGame")
	defer span.End()

	// Load from disk
	game, state, _ /*moves*/, err := s.getGameStateAndMoves(ctx, req.Id)
	if err != nil {
		return
	} else if game == nil {
		err = status.Error(codes.NotFound, fmt.Sprintf("Game with id '%s' not found", req.Id))
		return
	}

	// Populate screenshot URL if not set
	if len(game.PreviewUrls) == 0 {
		game.PreviewUrls = []string{fmt.Sprintf("/screenshots/games/%s/default.png", game.Id)}
	}

	// Cache everything
	resp.Game, _ = v1gorm.GameFromGameGORM(nil, game, nil)
	resp.State, _ = v1gorm.GameStateFromGameStateGORM(nil, state, nil)
	// TODO - convert move list to groups of moves and GroupMoveHistory
	// resp.History, _ = v1gorm.GameStateFromGameStateGORM(nil, state, nil)

	// Auto-migrate WorldData from old list-based format to new map-based format
	// This does not persist the migration - subsequent writes will save the new format
	if resp.State != nil && resp.State.WorldData != nil {
		lib.MigrateWorldData(resp.State.WorldData)
	}

	return resp, nil
}

// DeleteGame deletes a game
func (s *GamesService) DeleteGame(ctx context.Context, req *v1.DeleteGameRequest) (resp *v1.DeleteGameResponse, err error) {
	err = s.GameDAL.Delete(ctx, s.storage, req.Id)
	err = errors.Join(err, s.GameStateDAL.Delete(ctx, s.storage, req.Id))
	err = errors.Join(s.storage.Where("game_id = ?", req.Id).Delete(&v1gorm.GameMoveGORM{}).Error)
	resp = &v1.DeleteGameResponse{}
	return resp, err
}

// CreateGame creates a new game
func (s *GamesService) CreateGame(ctx context.Context, req *v1.CreateGameRequest) (resp *v1.CreateGameResponse, err error) {
	ctx, span := Tracer.Start(ctx, "CreateGames")
	defer span.End()
	resp = &v1.CreateGameResponse{}

	// Make sure world exists - for now we must be given a worldId to create game from
	world, err := s.ClientMgr.GetWorldsSvcClient().GetWorld(ctx, &v1.GetWorldRequest{Id: req.Game.WorldId})
	if err != nil {
		return nil, fmt.Errorf("Error loading world: %w", err)
	}

	// Validate the request (duplicate players, players with units/tiles, etc.)
	if err := s.ValidateCreateGameRequest(req.Game, world.WorldData); err != nil {
		return nil, err
	}

	now := time.Now()
	req.Game.CreatedAt = tspb.New(now)
	req.Game.UpdatedAt = tspb.New(now)

	gameGorm, err := v1gorm.GameToGameGORM(req.Game, nil, nil)
	if err != nil {
		return
	}
	existingId := gameGorm.Id
	gameGorm.Id = NewID(s.storage, "games", gameGorm.Id)
	if gameGorm.Id == "" {
		return nil, fmt.Errorf("game with ID %q already exists", existingId)
	}
	if err = s.GameDAL.Save(ctx, s.storage, gameGorm); err != nil {
		return
	}
	resp.Game, err = v1gorm.GameFromGameGORM(nil, gameGorm, nil)
	// TODO - investigate why keys arent copied in protoc-gen-dal
	req.Game.Id = gameGorm.Id

	// Save a new empty game state and a new move list
	gs := &v1.GameState{
		GameId:        req.Game.Id,
		CurrentPlayer: 1, // Game starts with player 1
		TurnCounter:   1, // First turn starts at 1 for lazy top-up pattern
		WorldData:     world.WorldData,
	}

	// Auto-migrate WorldData from old list-based format to new map-based format
	lib.MigrateWorldData(gs.WorldData)

	// Generate shortcuts for tiles and units
	lib.EnsureShortcuts(gs.WorldData)

	// Add initial base income to each player's starting coins
	// This ensures players start with their configured coins PLUS income from their starting bases
	if req.Game.Config != nil {
		var incomeConfig *v1.IncomeConfig
		if req.Game.Config.IncomeConfigs != nil {
			incomeConfig = req.Game.Config.IncomeConfigs
		}
		for i, player := range req.Game.Config.Players {
			baseIncome := lib.CalculatePlayerBaseIncome(player.PlayerId, gs.WorldData, incomeConfig)
			req.Game.Config.Players[i].Coins += baseIncome
		}
		// Update the game record with new coin values
		gameGorm, err = v1gorm.GameToGameGORM(req.Game, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to convert game: %w", err)
		}
		gameGorm.Id = req.Game.Id
		if err = s.GameDAL.Save(ctx, s.storage, gameGorm); err != nil {
			return nil, fmt.Errorf("failed to update game with initial coins: %w", err)
		}
		resp.Game = req.Game
	}

	gameStateGorm, err := v1gorm.GameStateToGameStateGORM(gs, nil, nil)
	if err != nil {
		log.Println("Here 1 ????: ", err)
		return
	}
	gameStateGorm.GameId = gameGorm.Id
	if err = s.GameStateDAL.Save(ctx, s.storage, gameStateGorm); err != nil {
		log.Println("Here2 ????: ", err)
		return
	}

	// Units start with default zero values (current_turn=0, distance_left=0, available_health=0)
	// They will be lazily topped-up when accessed if unit.current_turn < game.turn_counter
	// This eliminates the need to initialize all units at game creation

	resp = &v1.CreateGameResponse{
		Game:      req.Game,
		GameState: gs,
	}

	return resp, nil
}

// UpdateGame updates an existing game
func (s *GamesService) UpdateGame(ctx context.Context, req *v1.UpdateGameRequest) (resp *v1.UpdateGameResponse, err error) {
	if req.GameId == "" {
		return nil, fmt.Errorf("game ID is required")
	}
	resp = &v1.UpdateGameResponse{}
	ctx, span := Tracer.Start(ctx, "UpdateGame")
	defer span.End()

	gameGORM, _ /*state*/, _ /*moves*/, err := s.getGameStateAndMoves(ctx, req.GameId)
	if err != nil {
		return
	} else if gameGORM == nil {
		err = status.Error(codes.NotFound, fmt.Sprintf("Game with id '%s' not found", req.GameId))
		return
	}
	currGame, _ := v1gorm.GameFromGameGORM(nil, gameGORM, nil)

	// Load existing metadata if updating
	if req.NewGame != nil {
		// Update metadata fields
		if req.NewGame.Name != "" {
			currGame.Name = req.NewGame.Name
		}
		if req.NewGame.Description != "" {
			currGame.Description = req.NewGame.Description
		}
		if req.NewGame.Tags != nil {
			currGame.Tags = req.NewGame.Tags
		}
		if req.NewGame.Difficulty != "" {
			currGame.Difficulty = req.NewGame.Difficulty
		}
		if req.NewGame.Config != nil {
			currGame.Config = req.NewGame.Config
		}
		currGame.UpdatedAt = tspb.New(time.Now())
		gameGORM, _ = v1gorm.GameToGameGORM(currGame, nil, nil)

		if err = s.GameDAL.Save(ctx, s.storage, gameGORM); err != nil {
			return
		}

		resp.Game = currGame
	}

	if req.NewState != nil {
		// Load current game state to get version
		gameStateGorm, err := s.GameStateDAL.Get(ctx, s.storage, req.GameId)
		if err != nil {
			return resp, fmt.Errorf("failed to get game state: %w", err)
		}

		// Auto-migrate WorldData from old list-based format to new map-based format
		if req.NewState.WorldData != nil {
			lib.MigrateWorldData(req.NewState.WorldData)
		}

		// Make sure to topup units
		if req.NewState.WorldData != nil {
			rg, err := s.GetRuntimeGame(currGame, req.NewState)
			if err != nil {
				panic(err)
			}
			for _, unit := range req.NewState.WorldData.UnitsMap {
				rg.TopUpUnitIfNeeded(unit)
			}
		}

		oldVersion := gameStateGorm.Version

		// Server controls version - don't trust client
		req.NewState.Version = oldVersion

		// Update the gameStateGorm with new state data
		newGameStateGorm, _ := v1gorm.GameStateToGameStateGORM(req.NewState, nil, nil)
		newGameStateGorm.GameId = req.GameId
		newGameStateGorm.WorldData.ScreenshotIndexInfo.LastUpdatedAt = time.Now()
		newGameStateGorm.WorldData.ScreenshotIndexInfo.NeedsIndexing = true
		newGameStateGorm.Version = oldVersion + 1

		// Optimistic lock: update GameState with version check
		result := s.storage.Model(&v1gorm.GameStateGORM{}).
			Where("game_id = ? AND version = ?", req.GameId, oldVersion).
			Updates(newGameStateGorm)

		if result.Error != nil {
			return resp, fmt.Errorf("failed to update GameState: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return resp, fmt.Errorf("optimistic lock failed: GameState was modified by another request")
		}

		// Queue it for being screenshotted
		s.ScreenShotIndexer.Send("games", req.GameId, newGameStateGorm.Version, req.NewState.WorldData)
	}

	// Ignore history for now
	if req.NewHistory != nil {
		/*
			if err := s.storage.SaveArtifact(req.GameId, "history", req.NewHistory); err != nil {
				return nil, fmt.Errorf("failed to update game history: %w", err)
			}
		*/
	}

	return resp, err
}

// GetRuntimeGame implements the interface method (for compatibility)
func (s *GamesService) GetRuntimeGame(game *v1.Game, gameState *v1.GameState) (*lib.Game, error) {
	return lib.ProtoToRuntimeGame(game, gameState), nil
}

// GetRuntimeGameByID returns a cached runtime game instance for the given game ID
func (s *GamesService) GetRuntimeGameByID(ctx context.Context, gameID string) (*lib.Game, error) {
	// Load proto data (will use cache if available)
	resp, err := s.GetGame(ctx, &v1.GetGameRequest{Id: gameID})
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Convert to runtime game
	rtGame := lib.ProtoToRuntimeGame(resp.Game, resp.State)

	return rtGame, nil
}

func (s *GamesService) getGameStateAndMoves(ctx context.Context, gameId string) (game *v1gorm.GameGORM, state *v1gorm.GameStateGORM, moves []*v1gorm.GameMoveGORM, err error) {
	game, err = s.GameDAL.Get(ctx, s.storage, gameId)
	if err == nil {
		state, err = s.GameStateDAL.Get(ctx, s.storage, gameId)
	}
	if err == nil {
		moves, err = s.GameMoveDAL.List(ctx, s.storage.Where("game_id = ?", gameId).Order("group_number asc").Order("timestamp asc"))
	}
	// get the moves
	return
}

// SaveMoveGroup saves a move group atomically with the game state using checkpoint pattern.
// Moves are written first (orphans OK), then GameState is updated as the "commit point".
// Only moves with groupNumber <= currentGroupNumber are considered valid.
func (s *GamesService) SaveMoveGroup(ctx context.Context, gameId string, state *v1.GameState, group *v1.GameMoveGroup) error {
	ctx, span := Tracer.Start(ctx, "SaveMoveGroup")
	defer span.End()

	// 0. Delete any orphan moves from previous failed ProcessMoves calls
	// (moves with group_number > current_group_number are orphans)
	if err := s.storage.Where("game_id = ? AND group_number > ?", gameId, state.CurrentGroupNumber-1).
		Delete(&v1gorm.GameMoveGORM{}).Error; err != nil {
		return fmt.Errorf("failed to delete orphan moves: %w", err)
	}

	// 1. Save each move in the group as individual rows (orphans OK if state save fails)
	for i, move := range group.Moves {
		move.GroupNumber = group.GroupNumber
		move.MoveNumber = int64(i)

		moveGorm, err := v1gorm.GameMoveToGameMoveGORM(move, nil, func(src *v1.GameMove, dest *v1gorm.GameMoveGORM) error {
			dest.GameId = gameId
			// Handle oneof move_type by serializing to JSON bytes
			if src.GetMoveType() != nil {
				moveTypeBytes, err := protojson.Marshal(src)
				if err != nil {
					return fmt.Errorf("failed to serialize move_type: %w", err)
				}
				dest.MoveType = moveTypeBytes
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to convert move %d: %w", i, err)
		}

		if err := s.storage.Create(moveGorm).Error; err != nil {
			return fmt.Errorf("failed to save move %d: %w", i, err)
		}
	}

	// 2. Save GameState as the "commit point" - this makes the moves valid
	stateGorm, err := v1gorm.GameStateToGameStateGORM(state, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to convert game state: %w", err)
	}
	stateGorm.GameId = gameId
	stateGorm.Version = state.Version + 1

	if err := s.GameStateDAL.Save(ctx, s.storage, stateGorm); err != nil {
		return fmt.Errorf("failed to save game state: %w", err)
	}

	// Queue for screenshot indexing
	s.ScreenShotIndexer.Send("games", gameId, stateGorm.Version, state.WorldData)

	return nil
}

// ListMoves returns moves from game history, optionally filtered by group range
func (s *GamesService) ListMoves(ctx context.Context, req *v1.ListMovesRequest) (*v1.ListMovesResponse, error) {
	if req.GameId == "" {
		return nil, fmt.Errorf("game ID is required")
	}

	ctx, span := Tracer.Start(ctx, "ListMoves")
	defer span.End()

	// Build query with optional group range filters
	query := s.storage.Where("game_id = ?", req.GameId)
	if req.FromGroup > 0 {
		query = query.Where("group_number >= ?", req.FromGroup)
	}
	if req.ToGroup > 0 {
		query = query.Where("group_number <= ?", req.ToGroup)
	}
	query = query.Order("group_number asc").Order("move_number asc")

	moves, err := s.GameMoveDAL.List(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list moves: %w", err)
	}

	// Group moves by group_number
	groupMap := make(map[int64]*v1.GameMoveGroup)
	var groupNumbers []int64

	for _, moveGorm := range moves {
		move, err := v1gorm.GameMoveFromGameMoveGORM(nil, moveGorm, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to convert move: %w", err)
		}

		groupNum := move.GroupNumber
		if _, exists := groupMap[groupNum]; !exists {
			groupMap[groupNum] = &v1.GameMoveGroup{
				GroupNumber: groupNum,
				Moves:       []*v1.GameMove{},
			}
			groupNumbers = append(groupNumbers, groupNum)
		}
		groupMap[groupNum].Moves = append(groupMap[groupNum].Moves, move)
	}

	// Build ordered list of groups
	var groups []*v1.GameMoveGroup
	for _, num := range groupNumbers {
		groups = append(groups, groupMap[num])
	}

	// Check if there are earlier moves
	hasMore := false
	if req.FromGroup > 0 {
		var count int64
		s.storage.Model(&v1gorm.GameMoveGORM{}).
			Where("game_id = ? AND group_number < ?", req.GameId, req.FromGroup).
			Count(&count)
		hasMore = count > 0
	}

	return &v1.ListMovesResponse{
		MoveGroups: groups,
		HasMore:    hasMore,
	}, nil
}
