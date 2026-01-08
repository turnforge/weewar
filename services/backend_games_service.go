//go:build !wasm
// +build !wasm

package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	"github.com/turnforge/weewar/lib"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GameStateUpdater is an interface for updating GameState with optimistic locking
type GameStateUpdater interface {
	// GetGameStateVersion retrieves GameState by ID and returns version
	GetGameStateVersion(ctx context.Context, id string) (version int64, err error)

	// UpdateGameStateScreenshotIndexInfo updates the GameState.ScreenshotIndexInfo with optimistic locking
	// Returns error if version mismatch (optimistic lock failure)
	UpdateGameStateScreenshotIndexInfo(ctx context.Context, id string, oldVersion int64, lastIndexedAt time.Time, needsIndexing bool) error
}

// GameStorageProvider is implemented by concrete backends (fsbe, gormbe) to provide
// raw storage operations. BackendGamesService wraps these with caching.
// Concrete implementations should ONLY implement this interface - caching is handled
// transparently by BackendGamesService.
type GameStorageProvider interface {
	// Read operations - load directly from storage (no caching)
	LoadGame(ctx context.Context, id string) (*v1.Game, error)
	LoadGameState(ctx context.Context, id string) (*v1.GameState, error)
	LoadGameHistory(ctx context.Context, id string) (*v1.GameMoveHistory, error)

	// Write operations - save directly to storage
	SaveGame(ctx context.Context, id string, game *v1.Game) error
	SaveGameState(ctx context.Context, id string, state *v1.GameState) error
	SaveGameHistory(ctx context.Context, id string, history *v1.GameMoveHistory) error

	// SaveMoves saves a move group to storage - backend-specific implementation
	// For FS: appends to history file
	// For GORM: saves as individual rows with orphan cleanup
	SaveMoves(ctx context.Context, gameId string, group *v1.GameMoveGroup, currentGroupNumber int64) error

	// Delete operation
	DeleteFromStorage(ctx context.Context, id string) error
}

// BackendGamesService provides shared caching and screenshot indexing logic for backend game services
// It embeds BaseGamesService and adds caching + screenshot management
type BackendGamesService struct {
	BaseGamesService
	ClientMgr         *ClientMgr
	ScreenShotIndexer *ScreenShotIndexer
	GameStateUpdater  GameStateUpdater
	StorageProvider   GameStorageProvider // Set by concrete implementations

	// Cache configuration
	CacheEnabled bool // Set to true to enable in-memory caching

	// In-memory cache for game data - shared across all backend implementations
	gameCache    map[string]*v1.Game
	stateCache   map[string]*v1.GameState
	historyCache map[string]*v1.GameMoveHistory
	runtimeCache map[string]*lib.Game
	cacheMu      sync.RWMutex
}

// InitializeCache sets up the in-memory cache maps and enables caching
func (s *BackendGamesService) InitializeCache() {
	s.CacheEnabled = true
	s.gameCache = make(map[string]*v1.Game)
	s.stateCache = make(map[string]*v1.GameState)
	s.historyCache = make(map[string]*v1.GameMoveHistory)
	s.runtimeCache = make(map[string]*lib.Game)
}

// GetGame returns game data, checking cache first then falling back to storage.
// This is the main entry point for reading game data - caching is transparent.
// If CacheEnabled is false, always loads from storage.
func (s *BackendGamesService) GetGame(ctx context.Context, req *v1.GetGameRequest) (*v1.GetGameResponse, error) {
	id := req.Id
	if id == "" {
		return nil, fmt.Errorf("game ID is required")
	}

	// Check cache first if enabled
	if s.CacheEnabled {
		s.cacheMu.RLock()
		game, gameOk := s.gameCache[id]
		state, stateOk := s.stateCache[id]
		history, historyOk := s.historyCache[id]
		s.cacheMu.RUnlock()

		if gameOk && stateOk && historyOk {
			return &v1.GetGameResponse{
				Game:    game,
				State:   state,
				History: history,
			}, nil
		}
	}

	// Cache miss or disabled - load from storage provider
	if s.StorageProvider == nil {
		return nil, fmt.Errorf("storage provider not configured")
	}

	game, err := s.StorageProvider.LoadGame(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load game: %w", err)
	}

	state, err := s.StorageProvider.LoadGameState(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load game state: %w", err)
	}

	history, err := s.StorageProvider.LoadGameHistory(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load game history: %w", err)
	}

	// Auto-migrate WorldData if needed
	if state.WorldData != nil {
		lib.MigrateWorldData(state.WorldData)
	}

	// Update cache if enabled
	if s.CacheEnabled {
		s.cacheMu.Lock()
		s.gameCache[id] = game
		s.stateCache[id] = state
		s.historyCache[id] = history
		s.cacheMu.Unlock()
	}

	return &v1.GetGameResponse{
		Game:    game,
		State:   state,
		History: history,
	}, nil
}


// GetRuntimeGameCached returns a cached runtime game, creating one if needed
// If caching is disabled, always creates a new runtime game
func (s *BackendGamesService) GetRuntimeGameCached(id string, game *v1.Game, state *v1.GameState) *lib.Game {
	if s.CacheEnabled {
		s.cacheMu.RLock()
		rtGame, ok := s.runtimeCache[id]
		s.cacheMu.RUnlock()

		if ok {
			return rtGame
		}
	}

	// Create new runtime game
	rtGame := lib.ProtoToRuntimeGame(game, state)

	if s.CacheEnabled {
		s.cacheMu.Lock()
		s.runtimeCache[id] = rtGame
		s.cacheMu.Unlock()
	}

	return rtGame
}

// GetRuntimeGame implements the GamesService interface
func (s *BackendGamesService) GetRuntimeGame(game *v1.Game, gameState *v1.GameState) (*lib.Game, error) {
	return lib.ProtoToRuntimeGame(game, gameState), nil
}

// UpdateGame updates an existing game with transparent caching.
// It loads current data, merges changes, saves via StorageProvider, and updates cache.
func (s *BackendGamesService) UpdateGame(ctx context.Context, req *v1.UpdateGameRequest) (*v1.UpdateGameResponse, error) {
	if req.GameId == "" {
		return nil, fmt.Errorf("game ID is required")
	}
	if s.StorageProvider == nil {
		return nil, fmt.Errorf("storage provider not configured")
	}

	resp := &v1.UpdateGameResponse{}

	// Handle game metadata update
	if req.NewGame != nil {
		game, err := s.StorageProvider.LoadGame(ctx, req.GameId)
		if err != nil {
			return nil, fmt.Errorf("game not found: %w", err)
		}

		// Merge update fields
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
		if req.NewGame.Config != nil {
			game.Config = req.NewGame.Config
		}
		game.UpdatedAt = timestamppb.New(time.Now())

		if err := s.StorageProvider.SaveGame(ctx, req.GameId, game); err != nil {
			return nil, fmt.Errorf("failed to save game: %w", err)
		}

		s.updateCache(req.GameId, game, nil, nil)
		resp.Game = game
	}

	// Handle game state update
	if req.NewState != nil {
		// Load current game for runtime game creation
		game, err := s.StorageProvider.LoadGame(ctx, req.GameId)
		if err != nil {
			return nil, fmt.Errorf("failed to load game: %w", err)
		}

		// Load current state to get version
		currentState, err := s.StorageProvider.LoadGameState(ctx, req.GameId)
		if err != nil {
			return nil, fmt.Errorf("failed to load game state: %w", err)
		}

		// Auto-migrate WorldData
		if req.NewState.WorldData != nil {
			lib.MigrateWorldData(req.NewState.WorldData)
		}

		// Top up units
		if req.NewState.WorldData != nil {
			rg := lib.ProtoToRuntimeGame(game, req.NewState)
			for _, unit := range req.NewState.WorldData.UnitsMap {
				rg.TopUpUnitIfNeeded(unit)
			}
		}

		// Update version and index info
		oldVersion := currentState.Version
		if req.NewState.WorldData.ScreenshotIndexInfo == nil {
			req.NewState.WorldData.ScreenshotIndexInfo = &v1.IndexInfo{}
		}
		req.NewState.WorldData.ScreenshotIndexInfo.LastUpdatedAt = timestamppb.New(time.Now())
		req.NewState.WorldData.ScreenshotIndexInfo.NeedsIndexing = true
		req.NewState.Version = oldVersion + 1

		if err := s.StorageProvider.SaveGameState(ctx, req.GameId, req.NewState); err != nil {
			return nil, fmt.Errorf("failed to save game state: %w", err)
		}

		s.updateCache(req.GameId, nil, req.NewState, nil)

		// Queue for screenshot
		if s.ScreenShotIndexer != nil {
			s.ScreenShotIndexer.Send("games", req.GameId, req.NewState.Version, req.NewState.WorldData)
		}
	}

	// Handle history update
	if req.NewHistory != nil {
		if err := s.StorageProvider.SaveGameHistory(ctx, req.GameId, req.NewHistory); err != nil {
			return nil, fmt.Errorf("failed to save game history: %w", err)
		}

		s.updateCache(req.GameId, nil, nil, req.NewHistory)
	}

	return resp, nil
}

// DeleteGame deletes a game with transparent cache invalidation.
func (s *BackendGamesService) DeleteGame(ctx context.Context, req *v1.DeleteGameRequest) (*v1.DeleteGameResponse, error) {
	if s.StorageProvider == nil {
		return nil, fmt.Errorf("storage provider not configured")
	}

	err := s.StorageProvider.DeleteFromStorage(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	s.invalidateCache(req.Id)
	return &v1.DeleteGameResponse{}, nil
}

// SaveMoveGroup saves a move group with transparent cache update.
// It delegates move persistence to backend-specific SaveMoves, saves state, and updates cache.
func (s *BackendGamesService) SaveMoveGroup(ctx context.Context, gameId string, state *v1.GameState, group *v1.GameMoveGroup) error {
	if s.StorageProvider == nil {
		return fmt.Errorf("storage provider not configured")
	}

	// Save moves using backend-specific implementation
	if err := s.StorageProvider.SaveMoves(ctx, gameId, group, state.CurrentGroupNumber); err != nil {
		return fmt.Errorf("failed to save moves: %w", err)
	}

	// Save state (this is the "commit point")
	if err := s.StorageProvider.SaveGameState(ctx, gameId, state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// Load updated history for cache
	history, _ := s.StorageProvider.LoadGameHistory(ctx, gameId)

	// Update cache transparently
	s.updateCache(gameId, nil, state, history)

	// Queue for screenshot
	if s.ScreenShotIndexer != nil {
		s.ScreenShotIndexer.Send("games", gameId, state.Version, state.WorldData)
	}

	return nil
}

// updateCache is a private method to update cache after mutations
func (s *BackendGamesService) updateCache(id string, game *v1.Game, state *v1.GameState, history *v1.GameMoveHistory) {
	if !s.CacheEnabled {
		return
	}

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	if game != nil {
		s.gameCache[id] = game
	}
	if state != nil {
		s.stateCache[id] = state
	}
	if history != nil {
		s.historyCache[id] = history
	}
	// Invalidate runtime cache when state changes
	if state != nil {
		delete(s.runtimeCache, id)
	}
}

// invalidateCache is a private method to remove a game from all caches
func (s *BackendGamesService) invalidateCache(id string) {
	if !s.CacheEnabled {
		return
	}

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	delete(s.gameCache, id)
	delete(s.stateCache, id)
	delete(s.historyCache, id)
	delete(s.runtimeCache, id)
}

// InitializeScreenshotIndexer sets up the screenshot indexer with completion callback
func (s *BackendGamesService) InitializeScreenshotIndexer() {
	s.ScreenShotIndexer = NewScreenShotIndexer(s.ClientMgr)
	s.ScreenShotIndexer.OnComplete = s.handleScreenshotCompletion
}

// InitializeSyncBroadcast sets up the callback to broadcast moves to sync subscribers.
// Called by backend game services (fsbe, gormbe) after initialization.
func (s *BackendGamesService) InitializeSyncBroadcast() {
	s.OnMovesSaved = func(ctx context.Context, gameId string, moves []*v1.GameMove, groupNumber int64) {
		// Skip if ClientMgr is not available (e.g., in tests)
		if s.ClientMgr == nil {
			return
		}
		syncClient := s.ClientMgr.GetGameSyncSvcClient()
		if syncClient == nil {
			log.Println("Sync Client not found...")
			return
		}

		// Get current player from moves (all moves in a group are from same player)
		var player int32
		if len(moves) > 0 {
			player = moves[0].Player
		}

		// Broadcast moves to all subscribers
		_, err := syncClient.Broadcast(ctx, &v1.BroadcastRequest{
			GameId: gameId,
			Update: &v1.GameUpdate{
				UpdateType: &v1.GameUpdate_MovesPublished{
					MovesPublished: &v1.MovesPublished{
						Player:      player,
						Moves:       moves,
						GroupNumber: groupNumber,
					},
				},
			},
		})
		if err != nil {
			log.Printf("Failed to broadcast moves for game %s: %v", gameId, err)
		}
	}
}

// ValidateCreateGameRequest validates a CreateGameRequest for common errors
// that apply to all backend implementations (fsbe, gormbe, etc.)
func (s *BackendGamesService) ValidateCreateGameRequest(game *v1.Game, worldData *v1.WorldData) error {
	if game == nil {
		return fmt.Errorf("game data is required")
	}

	// Check for duplicate player IDs
	if game.Config != nil && len(game.Config.Players) > 0 {
		seenPlayerIds := make(map[int32]bool)
		for _, player := range game.Config.Players {
			if seenPlayerIds[player.PlayerId] {
				return fmt.Errorf("duplicate player ID: %d", player.PlayerId)
			}
			seenPlayerIds[player.PlayerId] = true
		}

		// Check that each player has at least one unit or tile in the world
		if worldData != nil {
			for _, player := range game.Config.Players {
				hasUnitOrTile := false

				// Check tiles owned by this player
				for _, tile := range worldData.TilesMap {
					if tile.Player == player.PlayerId {
						hasUnitOrTile = true
						break
					}
				}

				// Check units owned by this player
				if !hasUnitOrTile {
					for _, unit := range worldData.UnitsMap {
						if unit.Player == player.PlayerId {
							hasUnitOrTile = true
							break
						}
					}
				}

				if !hasUnitOrTile {
					return fmt.Errorf("player %d has no units or tiles in the world", player.PlayerId)
				}
			}
		}
	}

	return nil
}

// InitializePlayerStates initializes the PlayerStates map in GameState from game config.
// This sets up initial coins (starting_coins + base income) for each player.
// Called during game creation by both fsbe and gormbe.
func (s *BackendGamesService) InitializePlayerStates(gameState *v1.GameState, config *v1.GameConfiguration) {
	if config == nil {
		return
	}

	if gameState.PlayerStates == nil {
		gameState.PlayerStates = make(map[int32]*v1.PlayerState)
	}

	var incomeConfig *v1.IncomeConfig
	if config.IncomeConfigs != nil {
		incomeConfig = config.IncomeConfigs
	}

	for _, player := range config.Players {
		baseIncome := lib.CalculatePlayerBaseIncome(player.PlayerId, gameState.WorldData, incomeConfig)
		initialCoins := player.StartingCoins + baseIncome
		gameState.PlayerStates[player.PlayerId] = &v1.PlayerState{
			Coins:    initialCoins,
			IsActive: true,
		}
	}
}

// handleScreenshotCompletion updates IndexInfo after screenshots are generated
func (s *BackendGamesService) handleScreenshotCompletion(items []ScreenShotItem) error {
	for _, item := range items {
		// Only process games (not worlds)
		if item.Kind != "games" {
			continue
		}

		ctx := context.Background()

		// Get current version from storage
		currentVersion, err := s.GameStateUpdater.GetGameStateVersion(ctx, item.Id)
		if err != nil {
			log.Printf("Failed to get GameState for %s: %v", item.Id, err)
			continue
		}

		// Check version matches - if not, this gameState has been updated since we started
		if currentVersion != item.Version {
			log.Printf("Version mismatch for game %s: expected %d, got %d - skipping IndexInfo update",
				item.Id, item.Version, currentVersion)
			continue
		}

		// Update IndexInfo
		lastIndexedAt := time.Now()

		// Check if there were any errors
		hasErrors := len(item.ThemeErrors) > 0
		needsIndexing := hasErrors
		if hasErrors {
			log.Printf("Screenshot errors for game %s: %v", item.Id, item.ThemeErrors)
		}

		// Update IndexInfo (does not increment version - internal bookkeeping only)
		err = s.GameStateUpdater.UpdateGameStateScreenshotIndexInfo(ctx, item.Id, item.Version, lastIndexedAt, needsIndexing)
		if err != nil {
			log.Printf("Failed to update GameState IndexInfo for %s: %v", item.Id, err)
		} else {
			log.Printf("Successfully updated IndexInfo for game %s (version %d)",
				item.Id, item.Version)
		}
	}
	return nil
}
