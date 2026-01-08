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
	service.StorageProvider = service // GamesService implements GameStorageProvider
	service.GameStateUpdater = service
	service.InitializeCache() // Enable caching (optional - can be disabled via CacheEnabled = false)
	service.InitializeScreenshotIndexer()
	service.InitializeSyncBroadcast()

	return service
}

// LoadGame implements GameStorageProvider - loads game directly from database
func (s *GamesService) LoadGame(ctx context.Context, id string) (*v1.Game, error) {
	gameGorm, err := s.GameDAL.Get(ctx, s.storage, id)
	if err != nil {
		return nil, fmt.Errorf("game not found: %w", err)
	}
	game, err := v1gorm.GameFromGameGORM(nil, gameGorm, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to convert game: %w", err)
	}
	// Populate screenshot URL if not set
	if len(game.PreviewUrls) == 0 {
		game.PreviewUrls = []string{fmt.Sprintf("/screenshots/games/%s/default.png", game.Id)}
	}
	return game, nil
}

// LoadGameState implements GameStorageProvider - loads game state directly from database
func (s *GamesService) LoadGameState(ctx context.Context, id string) (*v1.GameState, error) {
	stateGorm, err := s.GameStateDAL.Get(ctx, s.storage, id)
	if err != nil {
		return nil, fmt.Errorf("game state not found: %w", err)
	}
	state, err := v1gorm.GameStateFromGameStateGORM(nil, stateGorm, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to convert game state: %w", err)
	}
	return state, nil
}

// LoadGameHistory implements GameStorageProvider - loads game history directly from database
func (s *GamesService) LoadGameHistory(ctx context.Context, id string) (*v1.GameMoveHistory, error) {
	// Load moves and convert to history format
	moves, err := s.GameMoveDAL.List(ctx, s.storage.Where("game_id = ?", id).Order("group_number asc").Order("timestamp asc"))
	if err != nil {
		return nil, fmt.Errorf("failed to load moves: %w", err)
	}

	// Group moves into GameMoveHistory
	history := &v1.GameMoveHistory{GameId: id}
	groupMap := make(map[int64]*v1.GameMoveGroup)

	for _, moveGorm := range moves {
		move, err := v1gorm.GameMoveFromGameMoveGORM(nil, moveGorm, nil)
		if err != nil {
			continue
		}
		groupNum := move.GroupNumber
		if _, exists := groupMap[groupNum]; !exists {
			groupMap[groupNum] = &v1.GameMoveGroup{
				GroupNumber: groupNum,
				Moves:       []*v1.GameMove{},
			}
		}
		groupMap[groupNum].Moves = append(groupMap[groupNum].Moves, move)
	}

	// Convert map to sorted slice
	for _, group := range groupMap {
		history.Groups = append(history.Groups, group)
	}

	return history, nil
}

// SaveGame implements GameStorageProvider - saves game metadata to database
func (s *GamesService) SaveGame(ctx context.Context, id string, game *v1.Game) error {
	gameGorm, err := v1gorm.GameToGameGORM(game, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to convert game: %w", err)
	}
	gameGorm.Id = id
	return s.GameDAL.Save(ctx, s.storage, gameGorm)
}

// SaveGameState implements GameStorageProvider - saves game state to database
func (s *GamesService) SaveGameState(ctx context.Context, id string, state *v1.GameState) error {
	stateGorm, err := v1gorm.GameStateToGameStateGORM(state, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to convert game state: %w", err)
	}
	stateGorm.GameId = id
	return s.GameStateDAL.Save(ctx, s.storage, stateGorm)
}

// SaveGameHistory implements GameStorageProvider - saves game history to database
// For GORM backend, this saves individual moves (history is virtual, built from moves)
func (s *GamesService) SaveGameHistory(ctx context.Context, id string, history *v1.GameMoveHistory) error {
	// For GORM, we save individual moves - the history is built from moves on read
	// This is called after SaveMoveGroup has already saved the moves, so it's a no-op
	// unless we need to rebuild the entire history
	return nil
}

// DeleteFromStorage implements GameStorageProvider - deletes game from database
func (s *GamesService) DeleteFromStorage(ctx context.Context, id string) error {
	err := s.GameDAL.Delete(ctx, s.storage, id)
	err = errors.Join(err, s.GameStateDAL.Delete(ctx, s.storage, id))
	err = errors.Join(err, s.storage.Where("game_id = ?", id).Delete(&v1gorm.GameMoveGORM{}).Error)
	return err
}

// SaveMoves implements GameStorageProvider - saves moves as individual rows with orphan cleanup
func (s *GamesService) SaveMoves(ctx context.Context, gameId string, group *v1.GameMoveGroup, currentGroupNumber int64) error {
	// Delete any orphan moves from previous failed ProcessMoves calls
	// (moves with group_number > current_group_number are orphans)
	if err := s.storage.Where("game_id = ? AND group_number > ?", gameId, currentGroupNumber-1).
		Delete(&v1gorm.GameMoveGORM{}).Error; err != nil {
		return fmt.Errorf("failed to delete orphan moves: %w", err)
	}

	// Save each move in the group as individual rows
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

	return nil
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

	// Initialize player runtime state with starting coins + base income
	s.InitializePlayerStates(gs, req.Game.Config)

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
