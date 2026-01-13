//go:build !wasm
// +build !wasm

package gaebe

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/datastore"
	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	v1ds "github.com/turnforge/lilbattle/gen/datastore"
	v1dal "github.com/turnforge/lilbattle/gen/datastore/dal"
	"github.com/turnforge/lilbattle/lib"
	"github.com/turnforge/lilbattle/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// WorldsService implements the WorldsService gRPC interface for Datastore
type WorldsService struct {
	services.BackendWorldsService
	client       *datastore.Client
	namespace    string
	MaxPageSize  int
	WorldDAL     v1dal.WorldDatastoreDAL
	WorldDataDAL v1dal.WorldDataDatastoreDAL
}

// NewWorldsService creates a new WorldsService implementation
func NewWorldsService(client *datastore.Client, namespace string, clientMgr *services.ClientMgr) *WorldsService {
	service := &WorldsService{
		client:      client,
		namespace:   namespace,
		MaxPageSize: 1000,
	}
	service.WorldDAL.Namespace = namespace
	service.WorldDataDAL.Namespace = namespace
	service.ClientMgr = clientMgr
	service.Self = service
	service.WorldDataUpdater = service
	service.InitializeScreenshotIndexer()

	return service
}

// GetWorldData implements WorldDataUpdater interface
func (s *WorldsService) GetWorldData(ctx context.Context, id string) (int64, error) {
	key := NamespacedKey("WorldData", id, s.namespace)
	worldDataDs, err := s.WorldDataDAL.Get(ctx, s.client, key)
	if err != nil {
		return 0, err
	}
	if worldDataDs == nil {
		return 0, fmt.Errorf("world data not found: %s", id)
	}
	return worldDataDs.Version, nil
}

// UpdateWorldDataIndexInfo implements WorldDataUpdater interface
// Note: This does NOT increment version - IndexInfo is internal bookkeeping
func (s *WorldsService) UpdateWorldDataIndexInfo(ctx context.Context, id string, oldVersion int64, lastIndexedAt time.Time, needsIndexing bool) error {
	key := NamespacedKey("WorldData", id, s.namespace)

	_, err := s.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		var worldDataDs v1ds.WorldDataDatastore
		if err := tx.Get(key, &worldDataDs); err != nil {
			return err
		}

		// Optimistic lock check
		if worldDataDs.Version != oldVersion {
			return VersionMismatchError
		}

		// Update only IndexInfo fields
		worldDataDs.ScreenshotIndexInfo.LastIndexedAt = lastIndexedAt
		worldDataDs.ScreenshotIndexInfo.NeedsIndexing = needsIndexing
		// Note: NOT incrementing version

		_, err := tx.Put(key, &worldDataDs)
		return err
	})

	return err
}

// CreateWorld creates a new world
func (s *WorldsService) CreateWorld(ctx context.Context, req *v1.CreateWorldRequest) (resp *v1.CreateWorldResponse, err error) {
	ctx, span := Tracer.Start(ctx, "CreateWorld")
	defer span.End()

	resp = &v1.CreateWorldResponse{}
	if req.World == nil {
		return nil, fmt.Errorf("world data is required")
	}

	// Try to assign ID (custom or generated)
	assignedId := NewID(ctx, s.client, s.namespace, "worlds", req.World.Id)
	if assignedId == "" {
		return nil, fmt.Errorf("world with ID %q already exists or failed to generate ID", req.World.Id)
	}
	req.World.Id = assignedId

	now := time.Now()
	req.World.CreatedAt = tspb.New(now)
	req.World.UpdatedAt = tspb.New(now)

	// Auto-migrate from old list-based format to new map-based format before saving
	lib.MigrateWorldData(req.WorldData)

	// Use transaction to save world + world data atomically
	_, err = s.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		// Save world metadata
		worldDs, err := v1ds.WorldToWorldDatastore(req.World, nil, nil)
		if err != nil {
			return err
		}
		worldKey := NamespacedKey("World", req.World.Id, s.namespace)
		worldDs.Key = worldKey
		if _, err := tx.Put(worldKey, worldDs); err != nil {
			return err
		}

		// Save world data
		worldDataDs, err := v1ds.WorldDataToWorldDataDatastore(req.WorldData, nil, nil)
		if err != nil {
			return err
		}
		if worldDataDs == nil {
			worldDataDs = &v1ds.WorldDataDatastore{}
		}
		worldDataKey := NamespacedKey("WorldData", req.World.Id, s.namespace)
		worldDataDs.Key = worldDataKey
		worldDataDs.WorldId = req.World.Id
		if _, err := tx.Put(worldDataKey, worldDataDs); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create world: %w", err)
	}

	resp.World = req.World
	resp.WorldData = req.WorldData
	return resp, nil
}

// ListWorlds returns all available worlds (metadata only for performance)
func (s *WorldsService) ListWorlds(ctx context.Context, req *v1.ListWorldsRequest) (resp *v1.ListWorldsResponse, err error) {
	ctx, span := Tracer.Start(ctx, "ListWorlds")
	defer span.End()

	resp = &v1.ListWorldsResponse{
		Pagination: &v1.PaginationResponse{
			HasMore:      false,
			TotalResults: 0,
		},
	}

	query := NamespacedQuery("World", s.namespace).
		Order("name")

	var entities []*v1ds.WorldDatastore
	keys, err := s.client.GetAll(ctx, query, &entities)
	if err != nil {
		return nil, err
	}

	// Set keys on entities
	for i, key := range keys {
		entities[i].Key = key
	}

	// Convert to proto
	for _, entity := range entities {
		world, err := v1ds.WorldFromWorldDatastore(nil, entity, nil)
		if err != nil {
			log.Printf("Warning: failed to convert world: %v", err)
			continue
		}

		// Populate screenshot URL if not set
		if len(world.PreviewUrls) == 0 {
			world.PreviewUrls = []string{fmt.Sprintf("/screenshots/worlds/%s/default.png", world.Id)}
		}
		resp.Items = append(resp.Items, world)
	}

	resp.Pagination.TotalResults = int32(len(resp.Items))
	return resp, nil
}

// GetWorld returns a specific world with complete data including tiles and units
func (s *WorldsService) GetWorld(ctx context.Context, req *v1.GetWorldRequest) (resp *v1.GetWorldResponse, err error) {
	if req.Id == "" {
		return nil, fmt.Errorf("world ID is required")
	}

	ctx, span := Tracer.Start(ctx, "GetWorld")
	defer span.End()

	resp = &v1.GetWorldResponse{}

	world, worldData, err := s.getWorldAndData(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if world == nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("World with id '%s' not found", req.Id))
	}

	// Populate screenshot URL if not set
	if len(world.PreviewUrls) == 0 {
		world.PreviewUrls = []string{fmt.Sprintf("/screenshots/worlds/%s/default.png", world.Id)}
	}

	resp.World = world
	resp.WorldData = worldData

	// Auto-migrate from old list-based format to new map-based format
	lib.MigrateWorldData(resp.WorldData)

	return resp, nil
}

// UpdateWorld updates an existing world
func (s *WorldsService) UpdateWorld(ctx context.Context, req *v1.UpdateWorldRequest) (resp *v1.UpdateWorldResponse, err error) {
	if req.World == nil || req.World.Id == "" {
		return nil, fmt.Errorf("world ID is required")
	}

	ctx, span := Tracer.Start(ctx, "UpdateWorld")
	defer span.End()

	resp = &v1.UpdateWorldResponse{}

	worldKey := NamespacedKey("World", req.World.Id, s.namespace)
	worldDataKey := NamespacedKey("WorldData", req.World.Id, s.namespace)

	_, err = s.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		// Load existing world
		var worldDs v1ds.WorldDatastore
		if err := tx.Get(worldKey, &worldDs); err != nil {
			if err == datastore.ErrNoSuchEntity {
				return status.Error(codes.NotFound, fmt.Sprintf("World with id '%s' not found", req.World.Id))
			}
			return err
		}
		worldDs.Key = worldKey

		// Load existing world data
		var worldDataDs v1ds.WorldDataDatastore
		if err := tx.Get(worldDataKey, &worldDataDs); err != nil {
			if err == datastore.ErrNoSuchEntity {
				worldDataDs = v1ds.WorldDataDatastore{WorldId: req.World.Id}
			} else {
				return err
			}
		}
		worldDataDs.Key = worldDataKey

		// Update metadata fields
		if req.World.Name != "" {
			worldDs.Name = req.World.Name
		}
		if req.World.Description != "" {
			worldDs.Description = req.World.Description
		}
		if req.World.Tags != nil {
			worldDs.Tags = req.World.Tags
		}
		if req.World.Difficulty != "" {
			worldDs.Difficulty = req.World.Difficulty
		}
		worldDs.UpdatedAt = time.Now()

		// Update world data if provided
		worldDataSaved := false
		if req.ClearWorld {
			oldVersion := worldDataDs.Version
			worldDataDs = v1ds.WorldDataDatastore{
				WorldId: req.World.Id,
				Version: oldVersion,
			}
			worldDataDs.Key = worldDataKey
			worldDataSaved = true
		} else if req.WorldData != nil {
			// Auto-migrate incoming request data from old list-based format
			lib.MigrateWorldData(req.WorldData)

			// Optimistic lock: verify client version matches server version
			clientVersion := req.WorldData.Version
			serverVersion := worldDataDs.Version
			if clientVersion != serverVersion {
				return fmt.Errorf("optimistic lock failed: client has version %d but server has version %d", clientVersion, serverVersion)
			}

			// Convert existing to proto for comparison
			existingData, err := v1ds.WorldDataFromWorldDataDatastore(nil, &worldDataDs, nil)
			if err != nil {
				return err
			}
			lib.MigrateWorldData(existingData)

			// Use client version for the update
			if req.WorldData.TilesMap == nil {
				req.WorldData.TilesMap = existingData.TilesMap
			}
			if req.WorldData.UnitsMap == nil {
				req.WorldData.UnitsMap = existingData.UnitsMap
			}
			if req.WorldData.Crossings == nil {
				req.WorldData.Crossings = existingData.Crossings
			}

			newWorldDataDs, err := v1ds.WorldDataToWorldDataDatastore(req.WorldData, nil, nil)
			if err != nil {
				return err
			}
			worldDataDs = *newWorldDataDs
			worldDataDs.WorldId = req.World.Id
			worldDataDs.Key = worldDataKey
			worldDataSaved = true
		}

		// Save world
		if _, err := tx.Put(worldKey, &worldDs); err != nil {
			return err
		}

		// Save world data if modified
		if worldDataSaved {
			worldDataDs.ScreenshotIndexInfo.LastUpdatedAt = time.Now()
			worldDataDs.ScreenshotIndexInfo.NeedsIndexing = true
			worldDataDs.Version = worldDataDs.Version + 1

			if _, err := tx.Put(worldDataKey, &worldDataDs); err != nil {
				return err
			}

			// Convert for response
			resp.WorldData, err = v1ds.WorldDataFromWorldDataDatastore(nil, &worldDataDs, nil)
			if err != nil {
				return err
			}

			// Queue for screenshot
			s.ScreenShotIndexer.Send("worlds", worldDataDs.WorldId, worldDataDs.Version, resp.WorldData)
		}

		// Convert for response
		resp.World, err = v1ds.WorldFromWorldDatastore(nil, &worldDs, nil)
		return err
	})

	return resp, err
}

// DeleteWorld deletes a world
func (s *WorldsService) DeleteWorld(ctx context.Context, req *v1.DeleteWorldRequest) (resp *v1.DeleteWorldResponse, err error) {
	ctx, span := Tracer.Start(ctx, "DeleteWorld")
	defer span.End()

	resp = &v1.DeleteWorldResponse{}

	_, err = s.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		// Delete world
		worldKey := NamespacedKey("World", req.Id, s.namespace)
		if err := tx.Delete(worldKey); err != nil && err != datastore.ErrNoSuchEntity {
			return err
		}

		// Delete world data
		worldDataKey := NamespacedKey("WorldData", req.Id, s.namespace)
		if err := tx.Delete(worldDataKey); err != nil && err != datastore.ErrNoSuchEntity {
			return err
		}

		return nil
	})

	return resp, err
}

func (s *WorldsService) getWorldAndData(ctx context.Context, worldId string) (*v1.World, *v1.WorldData, error) {
	worldKey := NamespacedKey("World", worldId, s.namespace)
	worldDataKey := NamespacedKey("WorldData", worldId, s.namespace)

	worldDs, err := s.WorldDAL.Get(ctx, s.client, worldKey)
	if err != nil {
		return nil, nil, err
	}
	if worldDs == nil {
		return nil, nil, nil
	}

	worldDataDs, err := s.WorldDataDAL.Get(ctx, s.client, worldDataKey)
	if err != nil {
		return nil, nil, err
	}

	world, err := v1ds.WorldFromWorldDatastore(nil, worldDs, nil)
	if err != nil {
		return nil, nil, err
	}

	var worldData *v1.WorldData
	if worldDataDs != nil {
		worldData, err = v1ds.WorldDataFromWorldDataDatastore(nil, worldDataDs, nil)
		if err != nil {
			return nil, nil, err
		}
	}

	return world, worldData, nil
}
