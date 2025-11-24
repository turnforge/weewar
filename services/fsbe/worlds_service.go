//go:build !wasm
// +build !wasm

package fsbe

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/turnforge/turnengine/engine/storage"
	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	"github.com/turnforge/weewar/services"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

var WORLDS_STORAGE_DIR = ""

// FSWorldsService implements the FSWorldsService gRPC interface
type FSWorldsService struct {
	services.BaseWorldsService
	storage   *storage.FileStorage
	ClientMgr *services.ClientMgr
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
	return service
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
			world.PreviewUrls = []string{fmt.Sprintf("/worlds/%s/screenshots/default", world.Id)}
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
		return nil, fmt.Errorf("world metadata not found: %w", err)
	}

	// Populate screenshot URL if not set
	if len(world.PreviewUrls) == 0 {
		world.PreviewUrls = []string{fmt.Sprintf("/worlds/%s/screenshots/default", world.Id)}
	}

	worldData, err := storage.LoadFSArtifact[*v1.WorldData](s.storage, req.Id, "data")
	if err != nil {
		return nil, fmt.Errorf("world data not found: %w", err)
	}

	resp = &v1.GetWorldResponse{
		World:     world,
		WorldData: worldData,
	}

	return resp, nil
}

// UpdateWorld updates an existing world
func (s *FSWorldsService) UpdateWorld(ctx context.Context, req *v1.UpdateWorldRequest) (resp *v1.UpdateWorldResponse, err error) {
	if req.World == nil || req.World.Id == "" {
		return nil, fmt.Errorf("world ID is required")
	}

	// Load existing metadata
	world, err := storage.LoadFSArtifact[*v1.World](s.storage, req.World.Id, "metadata")
	if err != nil {
		return nil, fmt.Errorf("world not found: %w", err)
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
	world.UpdatedAt = tspb.New(time.Now())

	if err := s.storage.SaveArtifact(req.World.Id, "metadata", world); err != nil {
		return nil, fmt.Errorf("failed to update world metadata: %w", err)
	}

	worldData, err := storage.LoadFSArtifact[*v1.WorldData](s.storage, req.World.Id, "data")
	if err != nil {
		return nil, fmt.Errorf("world not found: %w", err)
	}

	// Update world data if provided
	var updated bool
	if req.ClearWorld {
		updated = true
		req.WorldData = &v1.WorldData{}
	} else if req.WorldData != nil {
		updated = true
		if req.WorldData.Tiles != nil {
			worldData.Tiles = req.WorldData.Tiles
		}
		if req.WorldData.Units != nil {
			worldData.Units = req.WorldData.Units
		}
	}

	if updated {
		err = s.storage.SaveArtifact(req.World.Id, "data", req.WorldData)
	}
	return resp, err
}

// DeleteWorld deletes a world
func (s *FSWorldsService) DeleteWorld(ctx context.Context, req *v1.DeleteWorldRequest) (resp *v1.DeleteWorldResponse, err error) {
	if req.Id == "" {
		return nil, fmt.Errorf("world ID is required")
	}
	err = s.storage.DeleteEntity(req.Id)

	resp = &v1.DeleteWorldResponse{}
	return resp, err
}

// CreateWorld creates a new world
func (s *FSWorldsService) CreateWorld(ctx context.Context, req *v1.CreateWorldRequest) (resp *v1.CreateWorldResponse, err error) {
	if req.World == nil {
		return nil, fmt.Errorf("world data is required")
	}

	worldId, err := s.storage.CreateEntity(req.World.Id)
	if err != nil {
		return resp, err
	}
	req.World.Id = worldId

	now := time.Now()
	req.World.CreatedAt = tspb.New(now)
	req.World.UpdatedAt = tspb.New(now)

	if err := s.storage.SaveArtifact(req.World.Id, "metadata", req.World); err != nil {
		return nil, fmt.Errorf("failed to create world: %w", err)
	}

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
