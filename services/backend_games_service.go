//go:build !wasm
// +build !wasm

package services

import (
	"context"
	"log"
	"time"
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
