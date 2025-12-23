//go:build !wasm
// +build !wasm

package services

import (
	"context"
	"fmt"
	"log"
	"time"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	"github.com/turnforge/weewar/lib"
)

// GameStateUpdater is an interface for updating GameState with optimistic locking
type GameStateUpdater interface {
	// GetGameStateVersion retrieves GameState by ID and returns version
	GetGameStateVersion(ctx context.Context, id string) (version int64, err error)

	// UpdateGameStateScreenshotIndexInfo updates the GameState.ScreenshotIndexInfo with optimistic locking
	// Returns error if version mismatch (optimistic lock failure)
	UpdateGameStateScreenshotIndexInfo(ctx context.Context, id string, oldVersion int64, lastIndexedAt time.Time, needsIndexing bool) error
}

// BackendGamesService provides shared screenshot indexing logic for backend game services
// It embeds BaseGamesService and adds screenshot management
type BackendGamesService struct {
	BaseGamesService
	ClientMgr         *ClientMgr
	ScreenShotIndexer *ScreenShotIndexer
	GameStateUpdater  GameStateUpdater
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
