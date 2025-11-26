//go:build !wasm
// +build !wasm

package services

import (
	"context"
	"log"
	"time"
)

// WorldDataUpdater is an interface for updating WorldData with optimistic locking
type WorldDataUpdater interface {
	// GetWorldData retrieves WorldData by ID and returns version
	GetWorldData(ctx context.Context, id string) (version int64, err error)

	// UpdateWorldDataIndexInfo updates the IndexInfo with optimistic locking
	// Returns error if version mismatch (optimistic lock failure)
	UpdateWorldDataIndexInfo(ctx context.Context, id string, oldVersion int64, lastIndexedAt time.Time, needsIndexing bool) error
}

// BackendWorldsService provides shared screenshot indexing logic for backend world services
// It embeds BaseWorldsService and adds screenshot management
type BackendWorldsService struct {
	BaseWorldsService
	ClientMgr         *ClientMgr
	ScreenShotIndexer *ScreenShotIndexer
	WorldDataUpdater  WorldDataUpdater
}

// InitializeScreenshotIndexer sets up the screenshot indexer with completion callback
func (s *BackendWorldsService) InitializeScreenshotIndexer() {
	s.ScreenShotIndexer = NewScreenShotIndexer(s.ClientMgr)
	s.ScreenShotIndexer.OnComplete = s.handleScreenshotCompletion
}

// handleScreenshotCompletion updates IndexInfo after screenshots are generated
func (s *BackendWorldsService) handleScreenshotCompletion(items []ScreenShotItem) error {
	for _, item := range items {
		// Only process worlds (not games)
		if item.Kind != "worlds" {
			continue
		}

		ctx := context.Background()

		// Get current version from storage
		currentVersion, err := s.WorldDataUpdater.GetWorldData(ctx, item.Id)
		if err != nil {
			log.Printf("Failed to get WorldData for %s: %v", item.Id, err)
			continue
		}

		// Check version matches - if not, this worldData has been updated since we started
		if currentVersion != item.Version {
			log.Printf("Version mismatch for world %s: expected %d, got %d - skipping IndexInfo update",
				item.Id, item.Version, currentVersion)
			continue
		}

		// Update IndexInfo
		lastIndexedAt := time.Now()

		// Check if there were any errors
		hasErrors := len(item.ThemeErrors) > 0
		needsIndexing := hasErrors
		if hasErrors {
			log.Printf("Screenshot errors for world %s: %v", item.Id, item.ThemeErrors)
		}

		// Update IndexInfo (does not increment version - internal bookkeeping only)
		err = s.WorldDataUpdater.UpdateWorldDataIndexInfo(ctx, item.Id, item.Version, lastIndexedAt, needsIndexing)
		if err != nil {
			log.Printf("Failed to update WorldData IndexInfo for %s: %v", item.Id, err)
		} else {
			log.Printf("Successfully updated IndexInfo for world %s (version %d)",
				item.Id, item.Version)
		}
	}
	return nil
}
