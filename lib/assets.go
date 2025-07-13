package weewar

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"sync"
)

// =============================================================================
// Asset Management System
// =============================================================================

// AssetManager handles loading and caching of game assets
type AssetManager struct {
	dataPath      string
	tileCache     map[int]image.Image
	unitCache     map[string]image.Image // key: "unitId_playerColor"
	gameData      *GameDataAssets
	cacheMutex    sync.RWMutex
	dataLoaded    bool
}

// GameDataAssets represents the structure of weewar-data.json
type GameDataAssets struct {
	Units    []UnitData    `json:"units"`
	Terrains []TerrainData `json:"terrains"`
}

// NewAssetManager creates a new asset manager
func NewAssetManager(dataPath string) *AssetManager {
	return &AssetManager{
		dataPath:   dataPath,
		tileCache:  make(map[int]image.Image),
		unitCache:  make(map[string]image.Image),
		cacheMutex: sync.RWMutex{},
	}
}

// GetTileImage returns the tile image for a given tile type
func (am *AssetManager) GetTileImage(tileType int) (image.Image, error) {
	am.cacheMutex.RLock()
	if img, exists := am.tileCache[tileType]; exists {
		am.cacheMutex.RUnlock()
		return img, nil
	}
	am.cacheMutex.RUnlock()
	
	// Load tile image
	tilePath := filepath.Join(am.dataPath, "Tiles", fmt.Sprintf("%d_files", tileType), "0.png")
	img, err := am.loadImageFile(tilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load tile image for type %d: %w", tileType, err)
	}
	
	// Cache the image
	am.cacheMutex.Lock()
	am.tileCache[tileType] = img
	am.cacheMutex.Unlock()
	
	return img, nil
}

// GetUnitImage returns the unit image for a given unit type and player color
func (am *AssetManager) GetUnitImage(unitType int, playerColor int) (image.Image, error) {
	key := fmt.Sprintf("%d_%d", unitType, playerColor)
	
	am.cacheMutex.RLock()
	if img, exists := am.unitCache[key]; exists {
		am.cacheMutex.RUnlock()
		return img, nil
	}
	am.cacheMutex.RUnlock()
	
	// Load unit image
	unitPath := filepath.Join(am.dataPath, "Units", fmt.Sprintf("%d_files", unitType), fmt.Sprintf("%d.png", playerColor))
	img, err := am.loadImageFile(unitPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load unit image for type %d, color %d: %w", unitType, playerColor, err)
	}
	
	// Cache the image
	am.cacheMutex.Lock()
	am.unitCache[key] = img
	am.cacheMutex.Unlock()
	
	return img, nil
}

// loadImageFile loads a PNG image from the filesystem
func (am *AssetManager) loadImageFile(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	img, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG image: %w", err)
	}
	
	return img, nil
}

// HasTileAsset checks if a tile asset exists without loading it
func (am *AssetManager) HasTileAsset(tileType int) bool {
	tilePath := filepath.Join(am.dataPath, "Tiles", fmt.Sprintf("%d_files", tileType), "0.png")
	_, err := os.Stat(tilePath)
	return err == nil
}

// HasUnitAsset checks if a unit asset exists without loading it
func (am *AssetManager) HasUnitAsset(unitType int, playerColor int) bool {
	unitPath := filepath.Join(am.dataPath, "Units", fmt.Sprintf("%d_files", unitType), fmt.Sprintf("%d.png", playerColor))
	_, err := os.Stat(unitPath)
	return err == nil
}

// PreloadCommonAssets preloads commonly used assets for better performance
func (am *AssetManager) PreloadCommonAssets() error {
	// Preload common tile types (1-26)
	for i := 1; i <= 26; i++ {
		if am.HasTileAsset(i) {
			_, err := am.GetTileImage(i)
			if err != nil {
				fmt.Printf("Warning: Could not preload tile %d: %v\n", i, err)
			}
		}
	}
	
	// Preload common unit types with basic player colors (0-5)
	for unitType := 1; unitType <= 44; unitType++ {
		for playerColor := 0; playerColor <= 5; playerColor++ {
			if am.HasUnitAsset(unitType, playerColor) {
				_, err := am.GetUnitImage(unitType, playerColor)
				if err != nil {
					// Continue on error - not all combinations exist
					continue
				}
			}
		}
	}
	
	return nil
}

// ClearCache clears all cached assets
func (am *AssetManager) ClearCache() {
	am.cacheMutex.Lock()
	defer am.cacheMutex.Unlock()
	
	am.tileCache = make(map[int]image.Image)
	am.unitCache = make(map[string]image.Image)
}

// LoadGameData loads the weewar-data.json file
func (am *AssetManager) LoadGameData() error {
	if am.dataLoaded {
		return nil // Already loaded
	}
	
	dataFile := filepath.Join(am.dataPath, "weewar-data.json")
	
	file, err := os.Open(dataFile)
	if err != nil {
		return fmt.Errorf("failed to open game data file: %w", err)
	}
	defer file.Close()
	
	am.gameData = &GameDataAssets{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(am.gameData); err != nil {
		return fmt.Errorf("failed to decode game data: %w", err)
	}
	
	am.dataLoaded = true
	return nil
}

// GetUnitData returns unit data for a given unit type
func (am *AssetManager) GetUnitData(unitType int) (*UnitData, error) {
	if !am.dataLoaded {
		if err := am.LoadGameData(); err != nil {
			return nil, err
		}
	}
	
	for i := range am.gameData.Units {
		if am.gameData.Units[i].ID == unitType {
			return &am.gameData.Units[i], nil
		}
	}
	
	return nil, fmt.Errorf("unit type %d not found", unitType)
}

// GetTerrainDataAsset returns terrain data for a given terrain type
func (am *AssetManager) GetTerrainDataAsset(terrainType int) (*TerrainData, error) {
	if !am.dataLoaded {
		if err := am.LoadGameData(); err != nil {
			return nil, err
		}
	}
	
	for i := range am.gameData.Terrains {
		if am.gameData.Terrains[i].ID == terrainType {
			return &am.gameData.Terrains[i], nil
		}
	}
	
	return nil, fmt.Errorf("terrain type %d not found", terrainType)
}

// GetCacheStats returns statistics about cached assets
func (am *AssetManager) GetCacheStats() (int, int) {
	am.cacheMutex.RLock()
	defer am.cacheMutex.RUnlock()
	
	return len(am.tileCache), len(am.unitCache)
}