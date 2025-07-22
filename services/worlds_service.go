package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var WORLDS_STORAGE_DIR = weewar.DevDataPath("storage/worlds")

// WorldsServiceImpl implements the WorldsService gRPC interface
type WorldsServiceImpl struct {
	v1.UnimplementedWorldsServiceServer
	storageDir string
}

// NewWorldsService creates a new WorldsService implementation
func NewWorldsService() *WorldsServiceImpl {
	service := &WorldsServiceImpl{
		storageDir: WORLDS_STORAGE_DIR,
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(service.storageDir, 0755); err != nil {
		log.Printf("Failed to create worlds storage directory: %v", err)
	}

	return service
}

// WorldMetadata represents the metadata stored in metadata.json
type WorldMetadata struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	Difficulty  string    `json:"difficulty"`
	CreatorID   string    `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WorldData represents the complete world data stored in data.json
type WorldData struct {
	// World tiles with hex coordinates as keys "q,r"
	Tiles map[string]*TileData `json:"tiles"`
	// All units on the world
	Units []*UnitData `json:"world_units"`
}

// TileData represents tile data for storage
type TileData struct {
	Q        int32 `json:"q"`
	R        int32 `json:"r"`
	TileType int32 `json:"tile_type"`
	Player   int32 `json:"player"`
}

// UnitData represents unit data for storage
type UnitData struct {
	Q        int32 `json:"q"`
	R        int32 `json:"r"`
	Player   int32 `json:"player"`
	UnitType int32 `json:"unit_type"`
}

// getWorldPath returns the directory path for a world
func (s *WorldsServiceImpl) getWorldPath(worldID string) string {
	return filepath.Join(s.storageDir, worldID)
}

// getMetadataPath returns the metadata.json file path for a world
func (s *WorldsServiceImpl) getMetadataPath(worldID string) string {
	return filepath.Join(s.getWorldPath(worldID), "metadata.json")
}

// getDataPath returns the data.json file path for a world
func (s *WorldsServiceImpl) getDataPath(worldID string) string {
	return filepath.Join(s.getWorldPath(worldID), "data.json")
}

// loadWorldMetadata loads metadata from metadata.json
func (s *WorldsServiceImpl) loadWorldMetadata(worldID string) (*WorldMetadata, error) {
	metadataPath := s.getMetadataPath(worldID)
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata for world %s: %w", worldID, err)
	}

	var metadata WorldMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata for world %s: %w", worldID, err)
	}

	return &metadata, nil
}

// loadWorldData loads complete world data from data.json
func (s *WorldsServiceImpl) loadWorldData(worldID string) (*WorldData, error) {
	dataPath := s.getDataPath(worldID)
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read world data for world %s: %w", worldID, err)
	}

	var worldData WorldData
	if err := json.Unmarshal(data, &worldData); err != nil {
		return nil, fmt.Errorf("failed to parse world data for world %s: %w", worldID, err)
	}

	return &worldData, nil
}

// saveWorldMetadata saves metadata to metadata.json
func (s *WorldsServiceImpl) saveWorldMetadata(worldID string, metadata *WorldMetadata) error {
	worldDir := s.getWorldPath(worldID)
	if err := os.MkdirAll(worldDir, 0755); err != nil {
		return fmt.Errorf("failed to create world directory %s: %w", worldDir, err)
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata for world %s: %w", worldID, err)
	}

	metadataPath := s.getMetadataPath(worldID)
	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata for world %s: %w", worldID, err)
	}

	return nil
}

// saveWorldData saves complete world data to data.json
func (s *WorldsServiceImpl) saveWorldData(worldID string, worldData *WorldData) error {
	data, err := json.MarshalIndent(worldData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal world data for world %s: %w", worldID, err)
	}

	dataPath := s.getDataPath(worldID)
	if err := os.WriteFile(dataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write world data for world %s: %w", worldID, err)
	}

	return nil
}

// checkWorldId checks if a world ID already exists in storage
func (s *WorldsServiceImpl) checkWorldId(worldID string) (bool, error) {
	metadataPath := s.getMetadataPath(worldID)
	_, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			// World doesn't exist - ID is available
			return false, nil
		}
		// Other file system error
		return false, fmt.Errorf("failed to check world ID %s: %w", worldID, err)
	}
	// World exists - ID is taken
	return true, nil
}

// newWorldId generates a new unique world ID of specified length (default 8 chars)
func (s *WorldsServiceImpl) newWorldId(numChars ...int) (string, error) {
	const maxRetries = 10

	// Default to 8 characters if not specified
	length := 8
	if len(numChars) > 0 && numChars[0] > 0 {
		length = numChars[0]
	}

	// Calculate number of bytes needed (2 hex chars per byte)
	numBytes := (length + 1) / 2

	for attempt := range maxRetries {
		// Generate random bytes
		bytes := make([]byte, numBytes)
		if _, err := rand.Read(bytes); err != nil {
			return "", fmt.Errorf("failed to generate random bytes: %w", err)
		}

		// Convert to hex string and truncate to exact length
		worldID := hex.EncodeToString(bytes)[:length]

		// Check if this ID is already taken
		exists, err := s.checkWorldId(worldID)
		if err != nil {
			return "", fmt.Errorf("failed to check world ID uniqueness: %w", err)
		}

		if !exists {
			// Found a unique ID
			return worldID, nil
		}

		// ID collision, try again
		log.Printf("World ID collision detected (attempt %d/%d): %s", attempt+1, maxRetries, worldID)
	}

	return "", fmt.Errorf("failed to generate unique world ID after %d attempts", maxRetries)
}

// convertToProtoTiles converts storage tiles to protobuf format
func (s *WorldsServiceImpl) convertToProtoTiles(tiles map[string]*TileData) map[string]*v1.Tile {
	result := make(map[string]*v1.Tile)
	for key, tile := range tiles {
		result[key] = &v1.Tile{
			Q:        tile.Q,
			R:        tile.R,
			TileType: tile.TileType,
			Player:   tile.Player,
		}
	}
	return result
}

// convertToProtoUnits converts storage units to protobuf format
func (s *WorldsServiceImpl) convertToProtoUnits(units []*UnitData) []*v1.Unit {
	result := make([]*v1.Unit, len(units))
	for i, unit := range units {
		result[i] = &v1.Unit{
			Q:        unit.Q,
			R:        unit.R,
			Player:   unit.Player,
			UnitType: unit.UnitType,
		}
	}
	return result
}

// convertFromProtoTiles converts protobuf tiles to storage format
func (s *WorldsServiceImpl) convertFromProtoTiles(tiles map[string]*v1.Tile) map[string]*TileData {
	result := make(map[string]*TileData)
	for key, tile := range tiles {
		result[key] = &TileData{
			Q:        tile.Q,
			R:        tile.R,
			TileType: tile.TileType,
			Player:   tile.Player,
		}
	}
	return result
}

// convertFromProtoUnits converts protobuf units to storage format
func (s *WorldsServiceImpl) convertFromProtoUnits(units []*v1.Unit) []*UnitData {
	result := make([]*UnitData, len(units))
	for i, unit := range units {
		result[i] = &UnitData{
			Q:        unit.Q,
			R:        unit.R,
			Player:   unit.Player,
			UnitType: unit.UnitType,
		}
	}
	return result
}

// convertMetadataToProto converts WorldMetadata to protobuf World (metadata only)
func (s *WorldsServiceImpl) convertMetadataToProto(metadata *WorldMetadata) *v1.World {
	return &v1.World{
		Id:          metadata.ID,
		Name:        metadata.Name,
		Description: metadata.Description,
		Tags:        metadata.Tags,
		Difficulty:  metadata.Difficulty,
		CreatorId:   metadata.CreatorID,
		ImageUrl:    fmt.Sprintf("/worlds/%s/preview", metadata.ID),
		CreatedAt:   timestamppb.New(metadata.CreatedAt),
		UpdatedAt:   timestamppb.New(metadata.UpdatedAt),
	}
}

// convertToFullProto converts metadata + data to complete protobuf World
func (s *WorldsServiceImpl) convertToFullProto(metadata *WorldMetadata, worldData *WorldData) *v1.World {
	protoWorld := s.convertMetadataToProto(metadata)
	protoWorld.Tiles = s.convertToProtoTiles(worldData.Tiles)
	protoWorld.Units = s.convertToProtoUnits(worldData.Units)
	return protoWorld
}

// ListWorlds returns all available worlds (metadata only for performance)
func (s *WorldsServiceImpl) ListWorlds(ctx context.Context, req *v1.ListWorldsRequest) (resp *v1.ListWorldsResponse, err error) {
	resp = &v1.ListWorldsResponse{
		Items: []*v1.World{},
		Pagination: &v1.PaginationResponse{
			HasMore:      false,
			TotalResults: 0,
		},
	}

	// Read all world directories
	entries, err := os.ReadDir(s.storageDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Storage directory doesn't exist yet, return empty list
			return resp, nil
		}
		return nil, fmt.Errorf("failed to read worlds storage directory: %w", err)
	}

	var worlds []*v1.World
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		worldID := entry.Name()
		metadata, err := s.loadWorldMetadata(worldID)
		if err != nil {
			log.Printf("Failed to load metadata for world %s: %v", worldID, err)
			continue
		}

		// Filter by owner if specified
		if req.OwnerId != "" && metadata.CreatorID != req.OwnerId {
			continue
		}

		// Only return metadata for listing (not full world data)
		worlds = append(worlds, s.convertMetadataToProto(metadata))
	}

	resp.Items = worlds
	resp.Pagination.TotalResults = int32(len(worlds))

	return resp, nil
}

// GetWorld returns a specific world with complete data including tiles and units
func (s *WorldsServiceImpl) GetWorld(ctx context.Context, req *v1.GetWorldRequest) (resp *v1.GetWorldResponse, err error) {
	if req.Id == "" {
		return nil, fmt.Errorf("world ID is required")
	}

	metadata, err := s.loadWorldMetadata(req.Id)
	if err != nil {
		return nil, fmt.Errorf("world not found: %w", err)
	}

	worldData, err := s.loadWorldData(req.Id)
	if err != nil {
		// If data.json doesn't exist, create empty world data
		log.Printf("World data not found for %s, creating empty data: %v", req.Id, err)
		worldData = &WorldData{
			Tiles: make(map[string]*TileData),
			Units: []*UnitData{},
		}
	}

	resp = &v1.GetWorldResponse{
		World: s.convertToFullProto(metadata, worldData),
	}

	return resp, nil
}

// CreateWorld creates a new world
func (s *WorldsServiceImpl) CreateWorld(ctx context.Context, req *v1.CreateWorldRequest) (resp *v1.CreateWorldResponse, err error) {
	if req.World == nil {
		return nil, fmt.Errorf("world data is required")
	}

	var worldID string

	// Determine which world ID to use
	if req.World.Id != "" {
		// World ID provided - check if it's available
		exists, err := s.checkWorldId(req.World.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to check world ID: %w", err)
		}

		if exists {
			// World ID is taken, generate a new one
			worldID, err = s.newWorldId()
			if err != nil {
				return nil, fmt.Errorf("failed to generate new world ID: %w", err)
			}
		} else {
			// World ID is available, use it
			worldID = req.World.Id
		}
	} else {
		// No world ID provided, generate a new one
		worldID, err = s.newWorldId()
		if err != nil {
			return nil, fmt.Errorf("failed to generate world ID: %w", err)
		}
	}

	now := time.Now()
	metadata := &WorldMetadata{
		ID:          worldID,
		Name:        req.World.Name,
		Description: req.World.Description,
		Tags:        req.World.Tags,
		Difficulty:  req.World.Difficulty,
		CreatorID:   req.World.CreatorId,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.saveWorldMetadata(worldID, metadata); err != nil {
		return nil, fmt.Errorf("failed to create world: %w", err)
	}

	// Create world data with tiles and units from request
	worldData := &WorldData{
		Tiles: s.convertFromProtoTiles(req.World.Tiles),
		Units: s.convertFromProtoUnits(req.World.Units),
	}

	// Initialize empty if no data provided
	if worldData.Tiles == nil {
		worldData.Tiles = make(map[string]*TileData)
	}
	if worldData.Units == nil {
		worldData.Units = []*UnitData{}
	}

	if err := s.saveWorldData(worldID, worldData); err != nil {
		log.Printf("Failed to create data.json for world %s: %v", worldID, err)
	}

	resp = &v1.CreateWorldResponse{
		World: s.convertToFullProto(metadata, worldData),
	}

	return resp, nil
}

// UpdateWorld updates an existing world
func (s *WorldsServiceImpl) UpdateWorld(ctx context.Context, req *v1.UpdateWorldRequest) (resp *v1.UpdateWorldResponse, err error) {
	if req.World == nil || req.World.Id == "" {
		return nil, fmt.Errorf("world ID is required")
	}

	// Load existing metadata
	metadata, err := s.loadWorldMetadata(req.World.Id)
	if err != nil {
		return nil, fmt.Errorf("world not found: %w", err)
	}

	// Update metadata fields
	if req.World.Name != "" {
		metadata.Name = req.World.Name
	}
	if req.World.Description != "" {
		metadata.Description = req.World.Description
	}
	if req.World.Tags != nil {
		metadata.Tags = req.World.Tags
	}
	if req.World.Difficulty != "" {
		metadata.Difficulty = req.World.Difficulty
	}
	metadata.UpdatedAt = time.Now()

	if err := s.saveWorldMetadata(req.World.Id, metadata); err != nil {
		return nil, fmt.Errorf("failed to update world metadata: %w", err)
	}

	// Update world data if provided
	if req.World.Tiles != nil || req.World.Units != nil {
		worldData := &WorldData{
			Tiles: s.convertFromProtoTiles(req.World.Tiles),
			Units: s.convertFromProtoUnits(req.World.Units),
		}

		if err := s.saveWorldData(req.World.Id, worldData); err != nil {
			return nil, fmt.Errorf("failed to update world data: %w", err)
		}

		resp = &v1.UpdateWorldResponse{
			World: s.convertToFullProto(metadata, worldData),
		}
	} else {
		resp = &v1.UpdateWorldResponse{
			World: s.convertMetadataToProto(metadata),
		}
	}

	return resp, nil
}

// DeleteWorld deletes a world
func (s *WorldsServiceImpl) DeleteWorld(ctx context.Context, req *v1.DeleteWorldRequest) (resp *v1.DeleteWorldResponse, err error) {
	if req.Id == "" {
		return nil, fmt.Errorf("world ID is required")
	}

	worldPath := s.getWorldPath(req.Id)
	if err := os.RemoveAll(worldPath); err != nil {
		return nil, fmt.Errorf("failed to delete world: %w", err)
	}

	resp = &v1.DeleteWorldResponse{}
	return resp, nil
}
