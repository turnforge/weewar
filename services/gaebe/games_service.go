//go:build !wasm
// +build !wasm

package gaebe

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"cloud.google.com/go/datastore"
	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	v1ds "github.com/turnforge/lilbattle/gen/datastore"
	v1dal "github.com/turnforge/lilbattle/gen/datastore/dal"
	"github.com/turnforge/lilbattle/lib"
	"github.com/turnforge/lilbattle/services"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// GamesService implements the GamesService gRPC interface for Datastore
type GamesService struct {
	services.BackendGamesService
	client       *datastore.Client
	namespace    string
	GameDAL      v1dal.GameDatastoreDAL
	GameStateDAL v1dal.GameStateDatastoreDAL
	GameMoveDAL  v1dal.GameMoveDatastoreDAL
}

// NewGamesService creates a new Datastore-backed GamesService
func NewGamesService(client *datastore.Client, namespace string, clientMgr *services.ClientMgr) *GamesService {
	service := &GamesService{
		client:    client,
		namespace: namespace,
	}
	service.GameDAL.Namespace = namespace
	service.GameStateDAL.Namespace = namespace
	service.GameMoveDAL.Namespace = namespace
	service.ClientMgr = clientMgr
	service.Self = service
	service.StorageProvider = service
	service.GameStateUpdater = service
	service.InitializeCache()
	service.InitializeScreenshotIndexer()
	service.InitializeSyncBroadcast()
	return service
}

// LoadGame implements GameStorageProvider
func (s *GamesService) LoadGame(ctx context.Context, id string) (*v1.Game, error) {
	key := NamespacedKey("Game", id, s.namespace)
	gameDs, err := s.GameDAL.Get(ctx, s.client, key)
	if err != nil {
		return nil, fmt.Errorf("failed to load game: %w", err)
	}
	if gameDs == nil {
		return nil, fmt.Errorf("game not found: %s", id)
	}

	game, err := v1ds.GameFromGameDatastore(nil, gameDs, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to convert game: %w", err)
	}

	// Populate screenshot URL if not set
	if len(game.PreviewUrls) == 0 {
		game.PreviewUrls = []string{fmt.Sprintf("/screenshots/games/%s/default.png", game.Id)}
	}

	return game, nil
}

// LoadGameState implements GameStorageProvider
func (s *GamesService) LoadGameState(ctx context.Context, id string) (*v1.GameState, error) {
	key := NamespacedKey("GameState", id, s.namespace)
	stateDs, err := s.GameStateDAL.Get(ctx, s.client, key)
	if err != nil {
		return nil, fmt.Errorf("failed to load game state: %w", err)
	}
	if stateDs == nil {
		return nil, fmt.Errorf("game state not found: %s", id)
	}

	state, err := v1ds.GameStateFromGameStateDatastore(nil, stateDs, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to convert game state: %w", err)
	}

	return state, nil
}

// LoadGameHistory implements GameStorageProvider
func (s *GamesService) LoadGameHistory(ctx context.Context, id string) (*v1.GameMoveHistory, error) {
	query := NamespacedQuery("GameMove", s.namespace).
		FilterField("game_id", "=", id).
		Order("group_number").
		Order("move_number")

	var entities []*v1ds.GameMoveDatastore
	keys, err := s.client.GetAll(ctx, query, &entities)
	if err != nil {
		return nil, fmt.Errorf("failed to load moves: %w", err)
	}

	// Set keys on entities
	for i, key := range keys {
		entities[i].Key = key
	}

	// Group moves into GameMoveGroups
	history := &v1.GameMoveHistory{GameId: id}
	groupMap := make(map[int64]*v1.GameMoveGroup)

	for _, entity := range entities {
		move, err := v1ds.GameMoveFromGameMoveDatastore(nil, entity, nil)
		if err != nil {
			log.Printf("Warning: failed to convert move: %v", err)
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
	var groupNums []int64
	for num := range groupMap {
		groupNums = append(groupNums, num)
	}
	sort.Slice(groupNums, func(i, j int) bool { return groupNums[i] < groupNums[j] })

	for _, num := range groupNums {
		history.Groups = append(history.Groups, groupMap[num])
	}

	return history, nil
}

// SaveGame implements GameStorageProvider
func (s *GamesService) SaveGame(ctx context.Context, id string, game *v1.Game) error {
	gameDs, err := v1ds.GameToGameDatastore(game, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to convert game: %w", err)
	}

	key := NamespacedKey("Game", id, s.namespace)
	gameDs.Key = key
	gameDs.Id = id

	_, err = s.GameDAL.Put(ctx, s.client, gameDs)
	return err
}

// SaveGameState implements GameStorageProvider
func (s *GamesService) SaveGameState(ctx context.Context, id string, state *v1.GameState) error {
	stateDs, err := v1ds.GameStateToGameStateDatastore(state, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to convert game state: %w", err)
	}

	key := NamespacedKey("GameState", id, s.namespace)
	stateDs.Key = key
	stateDs.GameId = id

	_, err = s.GameStateDAL.Put(ctx, s.client, stateDs)
	return err
}

// SaveGameHistory implements GameStorageProvider
// For Datastore, history is built from individual moves on read
func (s *GamesService) SaveGameHistory(ctx context.Context, id string, history *v1.GameMoveHistory) error {
	// No-op - history is virtual, built from moves
	return nil
}

// SaveMoves implements GameStorageProvider - saves moves with cross-entity transaction
func (s *GamesService) SaveMoves(ctx context.Context, gameId string, group *v1.GameMoveGroup, currentGroupNumber int64) error {
	// Use transaction for atomicity
	_, err := s.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		// Delete any orphan moves from previous failed attempts
		orphanQuery := NamespacedQuery("GameMove", s.namespace).
			FilterField("game_id", "=", gameId).
			FilterField("group_number", ">", currentGroupNumber-1).
			KeysOnly()

		orphanKeys, err := s.client.GetAll(ctx, orphanQuery, nil)
		if err != nil {
			return fmt.Errorf("failed to query orphan moves: %w", err)
		}

		if len(orphanKeys) > 0 {
			if err := tx.DeleteMulti(orphanKeys); err != nil {
				return fmt.Errorf("failed to delete orphan moves: %w", err)
			}
		}

		// Save each move in the group
		for i, move := range group.Moves {
			move.GroupNumber = group.GroupNumber
			move.MoveNumber = int64(i)

			moveDs, err := v1ds.GameMoveToGameMoveDatastore(move, nil, func(src *v1.GameMove, dest *v1ds.GameMoveDatastore) error {
				dest.GameId = gameId
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to convert move %d: %w", i, err)
			}

			// Use composite key: gameId-groupNumber-moveNumber
			keyName := fmt.Sprintf("%s-%d-%d", gameId, group.GroupNumber, i)
			key := NamespacedKey("GameMove", keyName, s.namespace)
			moveDs.Key = key

			if _, err := tx.Put(key, moveDs); err != nil {
				return fmt.Errorf("failed to save move %d: %w", i, err)
			}
		}

		return nil
	})

	return err
}

// DeleteFromStorage implements GameStorageProvider
func (s *GamesService) DeleteFromStorage(ctx context.Context, id string) error {
	_, err := s.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		// Delete game
		gameKey := NamespacedKey("Game", id, s.namespace)
		if err := tx.Delete(gameKey); err != nil && err != datastore.ErrNoSuchEntity {
			return err
		}

		// Delete game state
		stateKey := NamespacedKey("GameState", id, s.namespace)
		if err := tx.Delete(stateKey); err != nil && err != datastore.ErrNoSuchEntity {
			return err
		}

		// Delete all moves for this game
		moveQuery := NamespacedQuery("GameMove", s.namespace).
			FilterField("game_id", "=", id).
			KeysOnly()

		moveKeys, err := s.client.GetAll(ctx, moveQuery, nil)
		if err != nil {
			return err
		}

		if len(moveKeys) > 0 {
			return tx.DeleteMulti(moveKeys)
		}

		return nil
	})

	return err
}

// GetGameStateVersion implements GameStateUpdater
func (s *GamesService) GetGameStateVersion(ctx context.Context, id string) (int64, error) {
	key := NamespacedKey("GameState", id, s.namespace)
	stateDs, err := s.GameStateDAL.Get(ctx, s.client, key)
	if err != nil {
		return 0, err
	}
	if stateDs == nil {
		return 0, fmt.Errorf("game state not found: %s", id)
	}

	return stateDs.Version, nil
}

// UpdateGameStateScreenshotIndexInfo implements GameStateUpdater
// Note: This does NOT increment version - IndexInfo is internal bookkeeping
func (s *GamesService) UpdateGameStateScreenshotIndexInfo(ctx context.Context, id string, oldVersion int64, lastIndexedAt time.Time, needsIndexing bool) error {
	key := NamespacedKey("GameState", id, s.namespace)

	_, err := s.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		var stateDs v1ds.GameStateDatastore
		if err := tx.Get(key, &stateDs); err != nil {
			return err
		}

		// Optimistic lock check
		if stateDs.Version != oldVersion {
			return VersionMismatchError
		}

		// Update only IndexInfo fields
		stateDs.WorldData.ScreenshotIndexInfo.LastIndexedAt = lastIndexedAt
		stateDs.WorldData.ScreenshotIndexInfo.NeedsIndexing = needsIndexing
		// Note: NOT incrementing version

		_, err := tx.Put(key, &stateDs)
		return err
	})

	return err
}

// ListGames returns all games
func (s *GamesService) ListGames(ctx context.Context, req *v1.ListGamesRequest) (*v1.ListGamesResponse, error) {
	ctx, span := Tracer.Start(ctx, "ListGames")
	defer span.End()

	query := NamespacedQuery("Game", s.namespace).
		Order("-updated_at")

	var entities []*v1ds.GameDatastore
	keys, err := s.client.GetAll(ctx, query, &entities)
	if err != nil {
		return nil, err
	}

	// Set keys on entities
	for i, key := range keys {
		entities[i].Key = key
	}

	resp := &v1.ListGamesResponse{
		Items: make([]*v1.Game, 0, len(entities)),
		Pagination: &v1.PaginationResponse{
			TotalResults: int32(len(entities)),
		},
	}

	for _, entity := range entities {
		game, err := v1ds.GameFromGameDatastore(nil, entity, nil)
		if err != nil {
			log.Printf("Warning: failed to convert game: %v", err)
			continue
		}
		if len(game.PreviewUrls) == 0 {
			game.PreviewUrls = []string{fmt.Sprintf("/screenshots/games/%s/default.png", game.Id)}
		}
		resp.Items = append(resp.Items, game)
	}

	return resp, nil
}

// CreateGame creates a new game
func (s *GamesService) CreateGame(ctx context.Context, req *v1.CreateGameRequest) (*v1.CreateGameResponse, error) {
	ctx, span := Tracer.Start(ctx, "CreateGame")
	defer span.End()

	// Load world data for validation
	worldsSvcClient := s.ClientMgr.GetWorldsSvcClient()
	world, err := worldsSvcClient.GetWorld(ctx, &v1.GetWorldRequest{Id: req.Game.WorldId})
	if err != nil {
		return nil, fmt.Errorf("error loading world: %w", err)
	}

	// Validate request
	if err := s.ValidateCreateGameRequest(req.Game, world.WorldData); err != nil {
		return nil, err
	}

	// Try to assign ID (custom or generated)
	assignedId := NewID(ctx, s.client, s.namespace, "games", req.Game.Id)
	if assignedId == "" {
		return nil, fmt.Errorf("game with ID %q already exists or failed to generate ID", req.Game.Id)
	}
	req.Game.Id = assignedId

	now := time.Now()
	req.Game.CreatedAt = tspb.New(now)
	req.Game.UpdatedAt = tspb.New(now)

	// Create game state
	gs := &v1.GameState{
		GameId:        req.Game.Id,
		CurrentPlayer: 1,
		TurnCounter:   1,
		WorldData:     world.WorldData,
	}

	lib.MigrateWorldData(gs.WorldData)
	lib.EnsureShortcuts(gs.WorldData)
	s.InitializePlayerStates(gs, req.Game.Config)

	// Use transaction to save game + state atomically
	_, err = s.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		// Save game
		gameDs, err := v1ds.GameToGameDatastore(req.Game, nil, nil)
		if err != nil {
			return err
		}
		gameKey := NamespacedKey("Game", req.Game.Id, s.namespace)
		gameDs.Key = gameKey
		if _, err := tx.Put(gameKey, gameDs); err != nil {
			return err
		}

		// Save game state
		stateDs, err := v1ds.GameStateToGameStateDatastore(gs, nil, nil)
		if err != nil {
			return err
		}
		stateKey := NamespacedKey("GameState", req.Game.Id, s.namespace)
		stateDs.Key = stateKey
		if _, err := tx.Put(stateKey, stateDs); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	return &v1.CreateGameResponse{
		Game:      req.Game,
		GameState: gs,
	}, nil
}

// ListMoves returns moves filtered by group range
func (s *GamesService) ListMoves(ctx context.Context, req *v1.ListMovesRequest) (*v1.ListMovesResponse, error) {
	if req.GameId == "" {
		return nil, fmt.Errorf("game ID is required")
	}

	ctx, span := Tracer.Start(ctx, "ListMoves")
	defer span.End()

	query := NamespacedQuery("GameMove", s.namespace).
		FilterField("game_id", "=", req.GameId).
		Order("group_number").
		Order("move_number")

	if req.FromGroup > 0 {
		query = query.FilterField("group_number", ">=", req.FromGroup)
	}
	if req.ToGroup > 0 {
		query = query.FilterField("group_number", "<=", req.ToGroup)
	}

	var entities []*v1ds.GameMoveDatastore
	keys, err := s.client.GetAll(ctx, query, &entities)
	if err != nil {
		return nil, fmt.Errorf("failed to list moves: %w", err)
	}

	// Set keys on entities
	for i, key := range keys {
		entities[i].Key = key
	}

	// Group moves by group_number
	groupMap := make(map[int64]*v1.GameMoveGroup)
	var groupNumbers []int64

	for _, entity := range entities {
		move, err := v1ds.GameMoveFromGameMoveDatastore(nil, entity, nil)
		if err != nil {
			log.Printf("Warning: failed to convert move: %v", err)
			continue
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

	// Sort group numbers
	sort.Slice(groupNumbers, func(i, j int) bool { return groupNumbers[i] < groupNumbers[j] })

	var groups []*v1.GameMoveGroup
	for _, num := range groupNumbers {
		groups = append(groups, groupMap[num])
	}

	// Check if there are earlier moves
	hasMore := false
	if req.FromGroup > 0 {
		countQuery := NamespacedQuery("GameMove", s.namespace).
			FilterField("game_id", "=", req.GameId).
			FilterField("group_number", "<", req.FromGroup).
			KeysOnly()

		keys, _ := s.client.GetAll(ctx, countQuery, nil)
		hasMore = len(keys) > 0
	}

	return &v1.ListMovesResponse{
		MoveGroups: groups,
		HasMore:    hasMore,
	}, nil
}
