//go:build !wasm
// +build !wasm

package gormbe

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	v1gorm "github.com/turnforge/weewar/gen/gorm"
	v1dal "github.com/turnforge/weewar/gen/gorm/dal"
	"github.com/turnforge/weewar/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// WorldsService implements the WorldsService gRPC interface
type WorldsService struct {
	services.BaseWorldsService
	ClientMgr    *services.ClientMgr
	storage      *gorm.DB
	MaxPageSize  int
	WorldDAL     v1dal.WorldGORMDAL
	WorldDataDAL v1dal.WorldDataGORMDAL

	// The chan where we write world event updates to
	updateEventChan chan *v1gorm.WorldDataGORM
}

// NewWorldsService creates a new WorldsService implementation
func NewWorldsService(db *gorm.DB) *WorldsService {
	// db.AutoMigrate(&v1gorm.IndexRecordsLROGORM{})
	db.AutoMigrate(&v1gorm.WorldGORM{})
	db.AutoMigrate(&v1gorm.WorldDataGORM{})

	service := &WorldsService{
		storage:         db,
		MaxPageSize:     1000,
		updateEventChan: make(chan *v1gorm.WorldDataGORM),
	}
	service.WorldDAL.WillCreate = func(ctx context.Context, world *v1gorm.WorldGORM) error {
		world.UpdatedAt = time.Now()
		world.CreatedAt = time.Now()
		return nil
	}
	service.Self = service

	go service.StartIndexer()
	return service
}

func (b *WorldsService) StartIndexer() {
	pickedInWindow := map[string]*v1gorm.WorldDataGORM{}
	windowTimer := time.NewTimer(10 * time.Second)  // does the actual actioning the index update
	checkerTimer := time.NewTimer(10 * time.Second) // we also see which items have "changed" but missed the indexing through some loss
	startedBatchProcessing := false

	// TODO - Put this into a worker pool later
	startBatchProcessing := func(batch map[string]*v1gorm.WorldDataGORM) {
		for worldId, worldData := range batch {
			log.Println("Creating screenthots for: ", worldId, worldData)
			// imageBytes := generateScreenshot(worldId, worldData)
		}
		startedBatchProcessing = false
	}

	for {
		select {
		case <-windowTimer.C:
			log.Println("Time to update screenshots on batch")
			if !startedBatchProcessing {
				// if a batch process is already running then wait for next time to do this
				startedBatchProcessing = true
				go startBatchProcessing(pickedInWindow)
				pickedInWindow = map[string]*v1gorm.WorldDataGORM{}
			}
		case newWorldData := <-b.updateEventChan:
			curr, ok := pickedInWindow[newWorldData.WorldId]
			if curr == nil || !ok || curr.ScreenshotIndexInfo.LastUpdatedAt.Before(newWorldData.ScreenshotIndexInfo.LastUpdatedAt) {
				pickedInWindow[newWorldData.WorldId] = newWorldData
			}
		case <-checkerTimer.C:
			log.Println("Time to proactively find items that need to be re-indexed")
		}
	}
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
	existingId := worldGorm.Id
	worldGorm.Id = NewID(s.storage, "worlds", worldGorm.Id)
	if worldGorm.Id == "" {
		return nil, fmt.Errorf("world with ID %q already exists", existingId)
	}
	if err = s.WorldDAL.Save(ctx, s.storage, worldGorm); err != nil {
		return
	}
	resp.World, err = v1gorm.WorldFromWorldGORM(nil, worldGorm, nil)

	// see if we have world data too
	worldDataGorm, err := v1gorm.WorldDataToWorldDataGORM(req.WorldData, nil, nil)
	if err != nil {
		return
	}
	if worldDataGorm == nil {
		worldDataGorm = &v1gorm.WorldDataGORM{}
	}
	log.Println("Req WorldData: ", req.WorldData)
	log.Println("WorldData: ", worldDataGorm)
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
				output.PreviewUrls = []string{fmt.Sprintf("/worlds/%s/screenshots/default", output.Id)}
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
		world.PreviewUrls = []string{fmt.Sprintf("/worlds/%s/screenshots/default", world.Id)}
	}
	resp.World, err = v1gorm.WorldFromWorldGORM(nil, world, nil)
	if err != nil {
		return
	}
	resp.WorldData, err = v1gorm.WorldDataFromWorldDataGORM(nil, worldData, nil)
	if err != nil {
		return
	}
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
		req.WorldData = &v1.WorldData{}
		worldData = &v1gorm.WorldDataGORM{}
		worldDataSaved = true
	} else if req.WorldData != nil {
		worldDataSaved = true
		protoWorldData, err := v1gorm.WorldDataFromWorldDataGORM(nil, worldData, nil)
		if err != nil {
			return resp, err
		}
		if req.WorldData.Tiles == nil {
			req.WorldData.Tiles = protoWorldData.Tiles
		}
		if req.WorldData.Units == nil {
			req.WorldData.Units = protoWorldData.Units
		}
		log.Println("Req world data: ", req.WorldData)
		log.Println("New Proto world data: ", protoWorldData)
		worldData, err = v1gorm.WorldDataToWorldDataGORM(req.WorldData, nil, nil)
		worldData.WorldId = req.World.Id
	}

	err = s.WorldDAL.Save(ctx, s.storage, world)
	if err != nil {
		return
	}

	log.Println("On Update, err: ", err)
	log.Println("On Update, wds: ", worldDataSaved)
	if err == nil && worldDataSaved {
		worldData.ScreenshotIndexInfo.LastUpdatedAt = time.Now()
		worldData.ScreenshotIndexInfo.NeedsIndexing = true
		err = s.WorldDataDAL.Save(ctx, s.storage, worldData)
		log.Println("New world data: ", worldData, err)
		resp.World, err = v1gorm.WorldFromWorldGORM(nil, world, nil)
		resp.WorldData, err = v1gorm.WorldDataFromWorldDataGORM(nil, worldData, nil)

		// Queue it for being screenshotted
		s.updateEventChan <- worldData
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
