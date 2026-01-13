//go:build !wasm
// +build !wasm

package fsbe

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/panyam/goutils/storage"
	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
	"github.com/turnforge/lilbattle/services"
	"github.com/turnforge/lilbattle/services/authz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

var WORLDS_STORAGE_DIR = ""

// FSWorldsService implements the FSWorldsService gRPC interface
type FSWorldsService struct {
	services.BackendWorldsService
	storage *storage.FileStorage
}

// NewFSWorldsService creates a new FSWorldsService implementation
func NewFSWorldsService(storageDir string, clientMgr *services.ClientMgr) *FSWorldsService {
	if storageDir == "" {
		if WORLDS_STORAGE_DIR == "" {
			WORLDS_STORAGE_DIR = DevDataPath("storage/worlds")
		}
		storageDir = WORLDS_STORAGE_DIR
	}
	service := &FSWorldsService{storage: storage.NewFileStorage(storageDir)}
	service.ClientMgr = clientMgr
	service.Self = service
	service.WorldDataUpdater = service // Implement WorldDataUpdater interface
	service.InitializeScreenshotIndexer()
	return service
}

// GetWorldData implements WorldDataUpdater interface
func (s *FSWorldsService) GetWorldData(ctx context.Context, id string) (int64, error) {
	worldData, err := storage.LoadFSArtifact[*v1.WorldData](s.storage, id, "data")
	if err != nil {
		return 0, err
	}
	return worldData.Version, nil
}

// UpdateWorldDataIndexInfo implements WorldDataUpdater interface
// Note: This does NOT increment version - IndexInfo is internal bookkeeping
// that shouldn't invalidate user's optimistic lock
func (s *FSWorldsService) UpdateWorldDataIndexInfo(ctx context.Context, id string, oldVersion int64, lastIndexedAt time.Time, needsIndexing bool) error {
	worldData, err := storage.LoadFSArtifact[*v1.WorldData](s.storage, id, "data")
	if err != nil {
		return err
	}

	// Check version matches - if not, content was updated and we'll re-index later
	if worldData.Version != oldVersion {
		return fmt.Errorf("version mismatch - content was updated, will re-index later")
	}

	// Update only IndexInfo fields, don't touch version
	if worldData.ScreenshotIndexInfo == nil {
		worldData.ScreenshotIndexInfo = &v1.IndexInfo{}
	}
	worldData.ScreenshotIndexInfo.LastIndexedAt = tspb.New(lastIndexedAt)
	worldData.ScreenshotIndexInfo.NeedsIndexing = needsIndexing
	// Note: NOT incrementing version - this is internal bookkeeping

	// Save updated data
	err = s.storage.SaveArtifact(id, "data", worldData)
	if err != nil {
		return fmt.Errorf("failed to save world data: %w", err)
	}
	return nil
}

// ListWorlds returns all available worlds (metadata only for performance)
func (s *FSWorldsService) ListWorlds(ctx context.Context, req *v1.ListWorldsRequest) (resp *v1.ListWorldsResponse, err error) {
	resp = &v1.ListWorldsResponse{
		Items: []*v1.World{},
		Pagination: &v1.PaginationResponse{
			HasMore:      false,
			TotalResults: 0,
		},
	}
	resp.Items, err = storage.ListFSEntities[*v1.World](s.storage, nil)
	resp.Pagination.TotalResults = int32(len(resp.Items))

	// Populate screenshot URLs for all worlds
	for _, world := range resp.Items {
		if len(world.PreviewUrls) == 0 {
			world.PreviewUrls = []string{fmt.Sprintf("/screenshots/worlds/%s/default.png", world.Id)}
		}
	}

	return resp, nil
}

// GetWorld returns a specific world with complete data including tiles and units
func (s *FSWorldsService) GetWorld(ctx context.Context, req *v1.GetWorldRequest) (resp *v1.GetWorldResponse, err error) {
	if req.Id == "" {
		return nil, fmt.Errorf("world ID is required")
	}

	world, err := storage.LoadFSArtifact[*v1.World](s.storage, req.Id, "metadata")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, status.Errorf(codes.NotFound, "world %s not found", req.Id)
		}
		return nil, fmt.Errorf("failed to load world metadata: %w", err)
	}

	// Populate screenshot URL if not set
	if len(world.PreviewUrls) == 0 {
		world.PreviewUrls = []string{fmt.Sprintf("/screenshots/worlds/%s/default.png", world.Id)}
	}

	worldData, err := storage.LoadFSArtifact[*v1.WorldData](s.storage, req.Id, "data")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, status.Errorf(codes.NotFound, "world data for %s not found", req.Id)
		}
		return nil, fmt.Errorf("failed to load world data: %w", err)
	}

	// Auto-migrate from old list-based format to new map-based format
	// This does not persist the migration - subsequent writes will save the new format
	lib.MigrateWorldData(worldData)

	resp = &v1.GetWorldResponse{
		World:     world,
		WorldData: worldData,
	}

	return resp, nil
}

// UpdateWorld updates an existing world
// Authorization: Only the world creator can update a world.
func (s *FSWorldsService) UpdateWorld(ctx context.Context, req *v1.UpdateWorldRequest) (resp *v1.UpdateWorldResponse, err error) {
	if req.World == nil || req.World.Id == "" {
		return nil, fmt.Errorf("world ID is required")
	}

	// Load existing metadata
	world, err := storage.LoadFSArtifact[*v1.World](s.storage, req.World.Id, "metadata")
	if err != nil {
		return nil, fmt.Errorf("world not found: %w", err)
	}

	// Authorization: only the world creator can update
	if err := authz.CanModifyWorld(ctx, world); err != nil {
		return nil, err
	}

	// Update metadata fields
	if req.World.Name != "" {
		world.Name = req.World.Name
	}
	if req.World.Description != "" {
		world.Description = req.World.Description
	}
	if req.World.Tags != nil {
		world.Tags = req.World.Tags
	}
	if req.World.Difficulty != "" {
		world.Difficulty = req.World.Difficulty
	}
	if req.World.DefaultGameConfig != nil {
		world.DefaultGameConfig = req.World.DefaultGameConfig
	}
	world.UpdatedAt = tspb.New(time.Now())

	if err := s.storage.SaveArtifact(req.World.Id, "metadata", world); err != nil {
		return nil, fmt.Errorf("failed to update world metadata: %w", err)
	}

	worldData, err := storage.LoadFSArtifact[*v1.WorldData](s.storage, req.World.Id, "data")
	if err != nil {
		return nil, fmt.Errorf("world not found: %w", err)
	}

	// Auto-migrate from old list-based format to new map-based format
	lib.MigrateWorldData(worldData)

	// Update world data if provided
	worldDataSaved := false
	if req.ClearWorld {
		oldVersion := worldData.Version
		req.WorldData = &v1.WorldData{}
		req.WorldData.Version = oldVersion
		worldData = &v1.WorldData{}
		worldData.Version = oldVersion
		worldDataSaved = true
	} else if req.WorldData != nil {
		// Auto-migrate incoming request data from old list-based format
		lib.MigrateWorldData(req.WorldData)

		worldDataSaved = true

		// Optimistic lock: verify client version matches server version
		clientVersion := req.WorldData.Version
		serverVersion := worldData.Version
		if clientVersion != serverVersion {
			return nil, fmt.Errorf("optimistic lock failed: client has version %d but server has version %d", clientVersion, serverVersion)
		}

		// Use client version for the update
		if req.WorldData.TilesMap == nil {
			req.WorldData.TilesMap = worldData.TilesMap
		}
		if req.WorldData.UnitsMap == nil {
			req.WorldData.UnitsMap = worldData.UnitsMap
		}
		if req.WorldData.Crossings == nil {
			req.WorldData.Crossings = worldData.Crossings
		}
		worldData = req.WorldData
	}

	if worldDataSaved {
		if worldData.ScreenshotIndexInfo == nil {
			worldData.ScreenshotIndexInfo = &v1.IndexInfo{}
		}
		worldData.ScreenshotIndexInfo.LastUpdatedAt = tspb.New(time.Now())
		worldData.ScreenshotIndexInfo.NeedsIndexing = true
		worldData.Version = worldData.Version + 1

		err = s.storage.SaveArtifact(req.World.Id, "data", worldData)
		if err != nil {
			return nil, fmt.Errorf("failed to save world data: %w", err)
		}

		resp = &v1.UpdateWorldResponse{
			World:     world,
			WorldData: worldData,
		}

		// Queue it for being screenshotted
		s.ScreenShotIndexer.Send("worlds", world.Id, worldData.Version, resp.WorldData)
	} else {
		resp = &v1.UpdateWorldResponse{
			World:     world,
			WorldData: worldData,
		}
	}

	return resp, nil
}

// DeleteWorld deletes a world
// Authorization: Only the world creator can delete a world.
func (s *FSWorldsService) DeleteWorld(ctx context.Context, req *v1.DeleteWorldRequest) (resp *v1.DeleteWorldResponse, err error) {
	if req.Id == "" {
		return nil, fmt.Errorf("world ID is required")
	}

	// Load world to check ownership
	world, err := storage.LoadFSArtifact[*v1.World](s.storage, req.Id, "metadata")
	if err != nil {
		return nil, fmt.Errorf("world not found: %w", err)
	}

	// Authorization: only the world creator can delete
	if err := authz.CanModifyWorld(ctx, world); err != nil {
		return nil, err
	}

	err = s.storage.DeleteEntity(req.Id)

	resp = &v1.DeleteWorldResponse{}
	return resp, err
}

// CreateWorld creates a new world
func (s *FSWorldsService) CreateWorld(ctx context.Context, req *v1.CreateWorldRequest) (resp *v1.CreateWorldResponse, err error) {
	resp = &v1.CreateWorldResponse{}
	if req.World == nil {
		return nil, fmt.Errorf("world data is required")
	}

	worldId, err := s.storage.CreateEntity(req.World.Id)
	if err != nil {
		// Check if this is an ID conflict (custom ID already exists)
		if req.World.Id != "" {
			// Generate a suggested ID by adding a random suffix
			suggestedId := req.World.Id + "-" + shortRandSuffix()
			resp.FieldErrors = map[string]string{
				"id": suggestedId,
			}
			return resp, nil
		}
		return nil, err
	}
	req.World.Id = worldId

	now := time.Now()
	req.World.CreatedAt = tspb.New(now)
	req.World.UpdatedAt = tspb.New(now)

	if err := s.storage.SaveArtifact(req.World.Id, "metadata", req.World); err != nil {
		return nil, fmt.Errorf("failed to create world: %w", err)
	}

	// Auto-migrate from old list-based format to new map-based format before saving
	lib.MigrateWorldData(req.WorldData)

	// Create world data with tiles and units from request
	if err := s.storage.SaveArtifact(worldId, "data", req.WorldData); err != nil {
		log.Printf("Failed to create data.json for world %s: %v", worldId, err)
	}

	resp = &v1.CreateWorldResponse{
		World:     req.World,
		WorldData: req.WorldData,
	}

	return resp, nil
}
