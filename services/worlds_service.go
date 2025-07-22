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

var MAPS_STORAGE_DIR = weewar.DevDataPath("storage/maps")

// MapsServiceImpl implements the MapsService gRPC interface
type MapsServiceImpl struct {
	v1.UnimplementedMapsServiceServer
	storageDir string
}

// NewMapsService creates a new MapsService implementation
func NewMapsService() *MapsServiceImpl {
	service := &MapsServiceImpl{
		storageDir: MAPS_STORAGE_DIR,
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(service.storageDir, 0755); err != nil {
		log.Printf("Failed to create maps storage directory: %v", err)
	}

	return service
}

// MapMetadata represents the metadata stored in metadata.json
type MapMetadata struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	Difficulty  string    `json:"difficulty"`
	CreatorID   string    `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MapData represents the complete map data stored in data.json
type MapData struct {
	// Map tiles with hex coordinates as keys "q,r"
	Tiles map[string]*MapTileData `json:"tiles"`
	// All units on the map
	MapUnits []*MapUnitData `json:"map_units"`
}

// MapTileData represents tile data for storage
type MapTileData struct {
	Q        int32 `json:"q"`
	R        int32 `json:"r"`
	TileType int32 `json:"tile_type"`
	Player   int32 `json:"player"`
}

// MapUnitData represents unit data for storage
type MapUnitData struct {
	Q        int32 `json:"q"`
	R        int32 `json:"r"`
	Player   int32 `json:"player"`
	UnitType int32 `json:"unit_type"`
}

// getMapPath returns the directory path for a map
func (s *MapsServiceImpl) getMapPath(mapID string) string {
	return filepath.Join(s.storageDir, mapID)
}

// getMetadataPath returns the metadata.json file path for a map
func (s *MapsServiceImpl) getMetadataPath(mapID string) string {
	return filepath.Join(s.getMapPath(mapID), "metadata.json")
}

// getDataPath returns the data.json file path for a map
func (s *MapsServiceImpl) getDataPath(mapID string) string {
	return filepath.Join(s.getMapPath(mapID), "data.json")
}

// loadMapMetadata loads metadata from metadata.json
func (s *MapsServiceImpl) loadMapMetadata(mapID string) (*MapMetadata, error) {
	metadataPath := s.getMetadataPath(mapID)
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata for map %s: %w", mapID, err)
	}

	var metadata MapMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata for map %s: %w", mapID, err)
	}

	return &metadata, nil
}

// loadMapData loads complete map data from data.json
func (s *MapsServiceImpl) loadMapData(mapID string) (*MapData, error) {
	dataPath := s.getDataPath(mapID)
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read map data for map %s: %w", mapID, err)
	}

	var mapData MapData
	if err := json.Unmarshal(data, &mapData); err != nil {
		return nil, fmt.Errorf("failed to parse map data for map %s: %w", mapID, err)
	}

	return &mapData, nil
}

// saveMapMetadata saves metadata to metadata.json
func (s *MapsServiceImpl) saveMapMetadata(mapID string, metadata *MapMetadata) error {
	mapDir := s.getMapPath(mapID)
	if err := os.MkdirAll(mapDir, 0755); err != nil {
		return fmt.Errorf("failed to create map directory %s: %w", mapDir, err)
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata for map %s: %w", mapID, err)
	}

	metadataPath := s.getMetadataPath(mapID)
	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata for map %s: %w", mapID, err)
	}

	return nil
}

// saveMapData saves complete map data to data.json
func (s *MapsServiceImpl) saveMapData(mapID string, mapData *MapData) error {
	data, err := json.MarshalIndent(mapData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal map data for map %s: %w", mapID, err)
	}

	dataPath := s.getDataPath(mapID)
	if err := os.WriteFile(dataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write map data for map %s: %w", mapID, err)
	}

	return nil
}

// checkMapId checks if a map ID already exists in storage
func (s *MapsServiceImpl) checkMapId(mapID string) (bool, error) {
	metadataPath := s.getMetadataPath(mapID)
	_, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Map doesn't exist - ID is available
			return false, nil
		}
		// Other file system error
		return false, fmt.Errorf("failed to check map ID %s: %w", mapID, err)
	}
	// Map exists - ID is taken
	return true, nil
}

// newMapId generates a new unique map ID of specified length (default 8 chars)
func (s *MapsServiceImpl) newMapId(numChars ...int) (string, error) {
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
		mapID := hex.EncodeToString(bytes)[:length]

		// Check if this ID is already taken
		exists, err := s.checkMapId(mapID)
		if err != nil {
			return "", fmt.Errorf("failed to check map ID uniqueness: %w", err)
		}

		if !exists {
			// Found a unique ID
			return mapID, nil
		}

		// ID collision, try again
		log.Printf("Map ID collision detected (attempt %d/%d): %s", attempt+1, maxRetries, mapID)
	}

	return "", fmt.Errorf("failed to generate unique map ID after %d attempts", maxRetries)
}

// convertToProtoTiles converts storage tiles to protobuf format
func (s *MapsServiceImpl) convertToProtoTiles(tiles map[string]*MapTileData) map[string]*v1.MapTile {
	result := make(map[string]*v1.MapTile)
	for key, tile := range tiles {
		result[key] = &v1.MapTile{
			Q:        tile.Q,
			R:        tile.R,
			TileType: tile.TileType,
			Player:   tile.Player,
		}
	}
	return result
}

// convertToProtoUnits converts storage units to protobuf format
func (s *MapsServiceImpl) convertToProtoUnits(units []*MapUnitData) []*v1.MapUnit {
	result := make([]*v1.MapUnit, len(units))
	for i, unit := range units {
		result[i] = &v1.MapUnit{
			Q:        unit.Q,
			R:        unit.R,
			Player:   unit.Player,
			UnitType: unit.UnitType,
		}
	}
	return result
}

// convertFromProtoTiles converts protobuf tiles to storage format
func (s *MapsServiceImpl) convertFromProtoTiles(tiles map[string]*v1.MapTile) map[string]*MapTileData {
	result := make(map[string]*MapTileData)
	for key, tile := range tiles {
		result[key] = &MapTileData{
			Q:        tile.Q,
			R:        tile.R,
			TileType: tile.TileType,
			Player:   tile.Player,
		}
	}
	return result
}

// convertFromProtoUnits converts protobuf units to storage format
func (s *MapsServiceImpl) convertFromProtoUnits(units []*v1.MapUnit) []*MapUnitData {
	result := make([]*MapUnitData, len(units))
	for i, unit := range units {
		result[i] = &MapUnitData{
			Q:        unit.Q,
			R:        unit.R,
			Player:   unit.Player,
			UnitType: unit.UnitType,
		}
	}
	return result
}

// convertMetadataToProto converts MapMetadata to protobuf Map (metadata only)
func (s *MapsServiceImpl) convertMetadataToProto(metadata *MapMetadata) *v1.Map {
	return &v1.Map{
		Id:          metadata.ID,
		Name:        metadata.Name,
		Description: metadata.Description,
		Tags:        metadata.Tags,
		Difficulty:  metadata.Difficulty,
		CreatorId:   metadata.CreatorID,
		ImageUrl:    fmt.Sprintf("/maps/%s/preview", metadata.ID),
		CreatedAt:   timestamppb.New(metadata.CreatedAt),
		UpdatedAt:   timestamppb.New(metadata.UpdatedAt),
	}
}

// convertToFullProto converts metadata + data to complete protobuf Map
func (s *MapsServiceImpl) convertToFullProto(metadata *MapMetadata, mapData *MapData) *v1.Map {
	protoMap := s.convertMetadataToProto(metadata)
	protoMap.Tiles = s.convertToProtoTiles(mapData.Tiles)
	protoMap.MapUnits = s.convertToProtoUnits(mapData.MapUnits)
	return protoMap
}

// ListMaps returns all available maps (metadata only for performance)
func (s *MapsServiceImpl) ListMaps(ctx context.Context, req *v1.ListMapsRequest) (resp *v1.ListMapsResponse, err error) {
	resp = &v1.ListMapsResponse{
		Items: []*v1.Map{},
		Pagination: &v1.PaginationResponse{
			HasMore:      false,
			TotalResults: 0,
		},
	}

	// Read all map directories
	entries, err := os.ReadDir(s.storageDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Storage directory doesn't exist yet, return empty list
			return resp, nil
		}
		return nil, fmt.Errorf("failed to read maps storage directory: %w", err)
	}

	var maps []*v1.Map
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		mapID := entry.Name()
		metadata, err := s.loadMapMetadata(mapID)
		if err != nil {
			log.Printf("Failed to load metadata for map %s: %v", mapID, err)
			continue
		}

		// Filter by owner if specified
		if req.OwnerId != "" && metadata.CreatorID != req.OwnerId {
			continue
		}

		// Only return metadata for listing (not full map data)
		maps = append(maps, s.convertMetadataToProto(metadata))
	}

	resp.Items = maps
	resp.Pagination.TotalResults = int32(len(maps))

	return resp, nil
}

// GetMap returns a specific map with complete data including tiles and units
func (s *MapsServiceImpl) GetMap(ctx context.Context, req *v1.GetMapRequest) (resp *v1.GetMapResponse, err error) {
	if req.Id == "" {
		return nil, fmt.Errorf("map ID is required")
	}

	metadata, err := s.loadMapMetadata(req.Id)
	if err != nil {
		return nil, fmt.Errorf("map not found: %w", err)
	}

	mapData, err := s.loadMapData(req.Id)
	if err != nil {
		// If data.json doesn't exist, create empty map data
		log.Printf("Map data not found for %s, creating empty data: %v", req.Id, err)
		mapData = &MapData{
			Tiles:    make(map[string]*MapTileData),
			MapUnits: []*MapUnitData{},
		}
	}

	resp = &v1.GetMapResponse{
		Map: s.convertToFullProto(metadata, mapData),
	}

	return resp, nil
}

// CreateMap creates a new map
func (s *MapsServiceImpl) CreateMap(ctx context.Context, req *v1.CreateMapRequest) (resp *v1.CreateMapResponse, err error) {
	if req.Map == nil {
		return nil, fmt.Errorf("map data is required")
	}

	var mapID string

	// Determine which map ID to use
	if req.Map.Id != "" {
		// Map ID provided - check if it's available
		exists, err := s.checkMapId(req.Map.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to check map ID: %w", err)
		}

		if exists {
			// Map ID is taken, generate a new one
			mapID, err = s.newMapId()
			if err != nil {
				return nil, fmt.Errorf("failed to generate new map ID: %w", err)
			}
		} else {
			// Map ID is available, use it
			mapID = req.Map.Id
		}
	} else {
		// No map ID provided, generate a new one
		mapID, err = s.newMapId()
		if err != nil {
			return nil, fmt.Errorf("failed to generate map ID: %w", err)
		}
	}

	now := time.Now()
	metadata := &MapMetadata{
		ID:          mapID,
		Name:        req.Map.Name,
		Description: req.Map.Description,
		Tags:        req.Map.Tags,
		Difficulty:  req.Map.Difficulty,
		CreatorID:   req.Map.CreatorId,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.saveMapMetadata(mapID, metadata); err != nil {
		return nil, fmt.Errorf("failed to create map: %w", err)
	}

	// Create map data with tiles and units from request
	mapData := &MapData{
		Tiles:    s.convertFromProtoTiles(req.Map.Tiles),
		MapUnits: s.convertFromProtoUnits(req.Map.MapUnits),
	}

	// Initialize empty if no data provided
	if mapData.Tiles == nil {
		mapData.Tiles = make(map[string]*MapTileData)
	}
	if mapData.MapUnits == nil {
		mapData.MapUnits = []*MapUnitData{}
	}

	if err := s.saveMapData(mapID, mapData); err != nil {
		log.Printf("Failed to create data.json for map %s: %v", mapID, err)
	}

	resp = &v1.CreateMapResponse{
		Map: s.convertToFullProto(metadata, mapData),
	}

	return resp, nil
}

// UpdateMap updates an existing map
func (s *MapsServiceImpl) UpdateMap(ctx context.Context, req *v1.UpdateMapRequest) (resp *v1.UpdateMapResponse, err error) {
	if req.Map == nil || req.Map.Id == "" {
		return nil, fmt.Errorf("map ID is required")
	}

	// Load existing metadata
	metadata, err := s.loadMapMetadata(req.Map.Id)
	if err != nil {
		return nil, fmt.Errorf("map not found: %w", err)
	}

	// Update metadata fields
	if req.Map.Name != "" {
		metadata.Name = req.Map.Name
	}
	if req.Map.Description != "" {
		metadata.Description = req.Map.Description
	}
	if req.Map.Tags != nil {
		metadata.Tags = req.Map.Tags
	}
	if req.Map.Difficulty != "" {
		metadata.Difficulty = req.Map.Difficulty
	}
	metadata.UpdatedAt = time.Now()

	if err := s.saveMapMetadata(req.Map.Id, metadata); err != nil {
		return nil, fmt.Errorf("failed to update map metadata: %w", err)
	}

	// Update map data if provided
	if req.Map.Tiles != nil || req.Map.MapUnits != nil {
		mapData := &MapData{
			Tiles:    s.convertFromProtoTiles(req.Map.Tiles),
			MapUnits: s.convertFromProtoUnits(req.Map.MapUnits),
		}

		if err := s.saveMapData(req.Map.Id, mapData); err != nil {
			return nil, fmt.Errorf("failed to update map data: %w", err)
		}

		resp = &v1.UpdateMapResponse{
			Map: s.convertToFullProto(metadata, mapData),
		}
	} else {
		resp = &v1.UpdateMapResponse{
			Map: s.convertMetadataToProto(metadata),
		}
	}

	return resp, nil
}

// DeleteMap deletes a map
func (s *MapsServiceImpl) DeleteMap(ctx context.Context, req *v1.DeleteMapRequest) (resp *v1.DeleteMapResponse, err error) {
	if req.Id == "" {
		return nil, fmt.Errorf("map ID is required")
	}

	mapPath := s.getMapPath(req.Id)
	if err := os.RemoveAll(mapPath); err != nil {
		return nil, fmt.Errorf("failed to delete map: %w", err)
	}

	resp = &v1.DeleteMapResponse{}
	return resp, nil
}
