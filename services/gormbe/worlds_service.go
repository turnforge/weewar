//go:build !wasm
// +build !wasm

package gormbe

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	v1gorm "github.com/turnforge/lilbattle/gen/gorm"
	v1dal "github.com/turnforge/lilbattle/gen/gorm/dal"
	"github.com/turnforge/lilbattle/lib"
	"github.com/turnforge/lilbattle/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// WorldsService implements the WorldsService gRPC interface
type WorldsService struct {
	services.BackendWorldsService
	storage      *gorm.DB
	MaxPageSize  int
	WorldDAL     v1dal.WorldGORMDAL
	WorldDataDAL v1dal.WorldDataGORMDAL
}

// NewWorldsService creates a new WorldsService implementation
func NewWorldsService(db *gorm.DB, clientMgr *services.ClientMgr) *WorldsService {
	// db.AutoMigrate(&v1gorm.IndexRecordsLROGORM{})
	db.AutoMigrate(&v1gorm.WorldGORM{})
	db.AutoMigrate(&v1gorm.WorldDataGORM{})

	service := &WorldsService{
		storage:     db,
		MaxPageSize: 1000,
	}
	service.ClientMgr = clientMgr
	service.WorldDAL.WillCreate = func(ctx context.Context, world *v1gorm.WorldGORM) error {
		world.UpdatedAt = time.Now()
		world.CreatedAt = time.Now()
		return nil
	}
	service.Self = service
	service.WorldDataUpdater = service // Implement WorldDataUpdater interface
	service.InitializeScreenshotIndexer()

	return service
}

// GetWorldData implements WorldDataUpdater interface
func (s *WorldsService) GetWorldData(ctx context.Context, id string) (int64, error) {
	worldData, err := s.WorldDataDAL.Get(ctx, s.storage, id)
	if err != nil {
		return 0, err
	}
	return worldData.Version, nil
}

// UpdateWorldDataIndexInfo implements WorldDataUpdater interface
// Note: This does NOT increment version - IndexInfo is internal bookkeeping
// that shouldn't invalidate user's optimistic lock
func (s *WorldsService) UpdateWorldDataIndexInfo(ctx context.Context, id string, oldVersion int64, lastIndexedAt time.Time, needsIndexing bool) (err error) {
	// Update only IndexInfo fields, don't touch version
	// We still check version to ensure we're updating the right state,
	// but we don't increment it since this isn't a content change
	result := s.storage.Model(&v1gorm.WorldDataGORM{}).
		Where("world_id = ? AND version = ?", id, oldVersion).
		Updates(map[string]any{
			"screenshot_index_last_indexed_at": lastIndexedAt,
			"screenshot_index_needs_indexing":  needsIndexing,
		})
	err = result.Error
	if err != nil {
		return fmt.Errorf("failed to update IndexInfo: %w", err)
	}
	if result.RowsAffected == 0 {
		// Version changed - this is fine, just means content was updated
		// and we'll re-index on the next screenshot cycle
		return fmt.Errorf("version mismatch - content was updated, will re-index later")
	}
	return nil
}

// CreateWorld creates a new world
func (s *WorldsService) CreateWorld(ctx context.Context, req *v1.CreateWorldRequest) (resp *v1.CreateWorldResponse, err error) {
	ctx, span := Tracer.Start(ctx, "CreateWorlds")
	defer span.End()
	resp = &v1.CreateWorldResponse{}
	if req.World == nil {
		return nil, fmt.Errorf("world data is required")
	}

	worldGorm, err := v1gorm.WorldToWorldGORM(req.World, nil, nil)
	if err != nil {
		return
	}

	// Try to assign ID (custom or generated)
	assignedId, suggestedId, err := NewIDWithSuggestion(s.storage, "worlds", worldGorm.Id)
	if err != nil {
		return nil, err
	}
	if assignedId == "" {
		// ID conflict - return suggested ID in field_errors
		resp.FieldErrors = map[string]string{
			"id": suggestedId,
		}
		return resp, nil
	}
	worldGorm.Id = assignedId
	if err = s.WorldDAL.Save(ctx, s.storage, worldGorm); err != nil {
		return
	}
	resp.World, err = v1gorm.WorldFromWorldGORM(nil, worldGorm, nil)

	// Auto-migrate from old list-based format to new map-based format before saving
	lib.MigrateWorldData(req.WorldData)

	// see if we have world data too
	worldDataGorm, err := v1gorm.WorldDataToWorldDataGORM(req.WorldData, nil, nil)
	if err != nil {
		return
	}
	if worldDataGorm == nil {
		worldDataGorm = &v1gorm.WorldDataGORM{}
	}
	worldDataGorm.WorldId = worldGorm.Id
	if err = s.WorldDataDAL.Save(ctx, s.storage, worldDataGorm); err != nil {
		return
	}
	resp.WorldData, err = v1gorm.WorldDataFromWorldDataGORM(nil, worldDataGorm, nil)
	if err == nil {
		go VerifyID(s.storage, "worlds", worldGorm.Id)
	}
	return
}

// ListWorlds returns all available worlds (metadata only for performance)
func (s *WorldsService) ListWorlds(ctx context.Context, req *v1.ListWorldsRequest) (resp *v1.ListWorldsResponse, err error) {
	ctx, span := Tracer.Start(ctx, "ListWorlds")
	defer span.End()

	// Step 0: Preamble + Auth + Validate request

	resp = &v1.ListWorldsResponse{
		Pagination: &v1.PaginationResponse{
			HasMore:      false,
			TotalResults: 0,
		},
	}
	gormWorlds, err := s.WorldDAL.List(ctx, s.storage.Order("name asc"))
	if err != nil {
		return
	}

	// Step 3: Convert query results to proto results
	for _, input := range gormWorlds {
		output, err := v1gorm.WorldFromWorldGORM(nil, input, nil)
		if err == nil {
			// Populate screenshot URLs for all worlds
			if len(output.PreviewUrls) == 0 {
				output.PreviewUrls = []string{fmt.Sprintf("/screenshots/worlds/%s/default.png", output.Id)}
			}
			resp.Items = append(resp.Items, output)
		} else {
			log.Println("Error converting world: ", err, input)
		}
	}
	resp.Pagination.TotalResults = int32(len(resp.Items))

	return resp, nil
}

// GetWorld returns a specific world with complete data including tiles and units
func (s *WorldsService) GetWorld(ctx context.Context, req *v1.GetWorldRequest) (resp *v1.GetWorldResponse, err error) {
	if req.Id == "" {
		return nil, fmt.Errorf("world ID is required")
	}

	// Step 0: Preamble + Auth + Validate request
	resp = &v1.GetWorldResponse{}
	ctx, span := Tracer.Start(ctx, "GetWorld")
	defer span.End()

	// Step 1: Build query for world
	// Step 2: Execute query for world
	world, worldData, err := s.getWorldAndData(ctx, req.Id)
	if err != nil {
		return
	} else if world == nil {
		err = status.Error(codes.NotFound, fmt.Sprintf("World with id '%s' not found", req.Id))
		return
	}

	// Step 3: Convert query results to proto results
	// Populate screenshot URL if not set
	if len(world.PreviewUrls) == 0 {
		world.PreviewUrls = []string{fmt.Sprintf("/screenshots/worlds/%s/default.png", world.Id)}
	}
	resp.World, err = v1gorm.WorldFromWorldGORM(nil, world, nil)
	if err != nil {
		return
	}
	resp.WorldData, err = v1gorm.WorldDataFromWorldDataGORM(nil, worldData, nil)
	if err != nil {
		return
	}

	// Auto-migrate from old list-based format to new map-based format
	// This does not persist the migration - subsequent writes will save the new format
	lib.MigrateWorldData(resp.WorldData)

	return
}

// UpdateWorld updates an existing world
func (s *WorldsService) UpdateWorld(ctx context.Context, req *v1.UpdateWorldRequest) (resp *v1.UpdateWorldResponse, err error) {
	if req.World == nil || req.World.Id == "" {
		return nil, fmt.Errorf("world ID is required")
	}
	resp = &v1.UpdateWorldResponse{}
	ctx, span := Tracer.Start(ctx, "UpdateWorld")
	defer span.End()

	// Step 1: Build query for world
	// Step 2: Execute query for world
	world, worldData, err := s.getWorldAndData(ctx, req.World.Id)
	if err != nil {
		return
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
	world.UpdatedAt = time.Now()

	// Update world data if provided
	worldDataSaved := false
	if req.ClearWorld {
		oldVersion := worldData.Version
		req.WorldData = &v1.WorldData{}
		req.WorldData.Version = oldVersion
		worldData = &v1gorm.WorldDataGORM{}
		worldData.Version = oldVersion
		worldDataSaved = true
	} else if req.WorldData != nil {
		// Auto-migrate incoming request data from old list-based format
		lib.MigrateWorldData(req.WorldData)

		worldDataSaved = true
		protoWorldData, err := v1gorm.WorldDataFromWorldDataGORM(nil, worldData, nil)
		if err != nil {
			return resp, err
		}

		// Also migrate existing data for proper comparison
		lib.MigrateWorldData(protoWorldData)

		// Optimistic lock: verify client version matches server version
		clientVersion := req.WorldData.Version
		serverVersion := protoWorldData.Version
		if clientVersion != serverVersion {
			return resp, fmt.Errorf("optimistic lock failed: client has version %d but server has version %d", clientVersion, serverVersion)
		}

		// Use client version for the update
		if req.WorldData.TilesMap == nil {
			req.WorldData.TilesMap = protoWorldData.TilesMap
		}
		if req.WorldData.UnitsMap == nil {
			req.WorldData.UnitsMap = protoWorldData.UnitsMap
		}
		if req.WorldData.Crossings == nil {
			req.WorldData.Crossings = protoWorldData.Crossings
		}
		worldData, err = v1gorm.WorldDataToWorldDataGORM(req.WorldData, nil, nil)
		worldData.WorldId = req.World.Id
	}

	err = s.WorldDAL.Save(ctx, s.storage, world)
	if err != nil {
		return
	}

	if err == nil && worldDataSaved {
		oldVersion := worldData.Version
		worldData.ScreenshotIndexInfo.LastUpdatedAt = time.Now()
		worldData.ScreenshotIndexInfo.NeedsIndexing = true
		worldData.Version = worldData.Version + 1

		// Optimistic lock: update only if version hasn't changed
		result := s.storage.Model(&v1gorm.WorldDataGORM{}).
			Where("world_id = ? AND version = ?", worldData.WorldId, oldVersion).
			Updates(worldData)

		if result.Error != nil {
			return resp, fmt.Errorf("failed to update WorldData: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return resp, fmt.Errorf("optimistic lock failed: WorldData was modified by another request")
		}

		resp.World, err = v1gorm.WorldFromWorldGORM(nil, world, nil)
		resp.WorldData, err = v1gorm.WorldDataFromWorldDataGORM(nil, worldData, nil)

		// Queue it for being screenshotted
		s.ScreenShotIndexer.Send("worlds", worldData.WorldId, worldData.Version, resp.WorldData)
	}

	return resp, err
}

// DeleteWorld deletes a world
func (s *WorldsService) DeleteWorld(ctx context.Context, req *v1.DeleteWorldRequest) (resp *v1.DeleteWorldResponse, err error) {
	err = s.WorldDAL.Delete(ctx, s.storage, req.Id)
	err = errors.Join(err, s.WorldDataDAL.Delete(ctx, s.storage, req.Id))
	resp = &v1.DeleteWorldResponse{}
	return resp, err
}

func (s *WorldsService) getWorldAndData(ctx context.Context, worldId string) (world *v1gorm.WorldGORM, worldData *v1gorm.WorldDataGORM, err error) {
	world, err = s.WorldDAL.Get(ctx, s.storage, worldId)
	if err == nil {
		worldData, err = s.WorldDataDAL.Get(ctx, s.storage, worldId)
	}
	return
}
