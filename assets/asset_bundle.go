package assets

import _ "embed"

//go:embed lilbattle-rules.json
var RulesDataJSON []byte

//go:embed lilbattle-damage.json
var RulesDamageDataJSON []byte

// =============================================================================
// WASM Asset Bundle System - Embedded Assets for Browser
// =============================================================================

// go:embed v1/Tiles v1/Units
/*
var embeddedAssets embed.FS

// EmbeddedAssetManager is a WASM-specific AssetManager that uses embedded files
type EmbeddedAssetManager struct {
	tileCache  map[string]image.Image // key: "tileType_playerID"
	unitCache  map[string]image.Image // key: "unitId_playerColor"
	cacheMutex sync.RWMutex
	loaded     bool
}

// NewEmbeddedAssetManager creates a new embedded asset manager for WASM
func NewEmbeddedAssetManager() *EmbeddedAssetManager {
	return &EmbeddedAssetManager{
		tileCache:  make(map[string]image.Image),
		unitCache:  make(map[string]image.Image),
		cacheMutex: sync.RWMutex{},
	}
}

// GetTileImage returns the tile image for a given tile type and player ID using embedded assets
func (eam *EmbeddedAssetManager) GetTileImage(tileType int32, playerID int32) (image.Image, error) {
	cacheKey := fmt.Sprintf("%d_%d", tileType, playerID)

	eam.cacheMutex.RLock()
	if img, exists := eam.tileCache[cacheKey]; exists {
		eam.cacheMutex.RUnlock()
		return img, nil
	}
	eam.cacheMutex.RUnlock()

	// Load from embedded filesystem
	tilePath := fmt.Sprintf("v1/Tiles/%d/%d.png", tileType, playerID)
	img, err := eam.loadEmbeddedImage(tilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded tile image for type %d, player %d: %w", tileType, playerID, err)
	}

	// Cache the image
	eam.cacheMutex.Lock()
	eam.tileCache[cacheKey] = img
	eam.cacheMutex.Unlock()

	return img, nil
}

// GetUnitImage returns the unit image for a given unit type and player color using embedded assets
func (eam *EmbeddedAssetManager) GetUnitImage(unitType int32, playerColor int32) (image.Image, error) {
	key := fmt.Sprintf("%d_%d", unitType, playerColor)

	eam.cacheMutex.RLock()
	if img, exists := eam.unitCache[key]; exists {
		eam.cacheMutex.RUnlock()
		return img, nil
	}
	eam.cacheMutex.RUnlock()

	// Load from embedded filesystem
	unitPath := fmt.Sprintf("v1/Units/%d/%d.png", unitType, playerColor)
	img, err := eam.loadEmbeddedImage(unitPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded unit image for type %d, color %d: %w", unitType, playerColor, err)
	}

	// Cache the image
	eam.cacheMutex.Lock()
	eam.unitCache[key] = img
	eam.cacheMutex.Unlock()

	return img, nil
}

// loadEmbeddedImage loads a PNG image from the embedded filesystem
func (eam *EmbeddedAssetManager) loadEmbeddedImage(path string) (image.Image, error) {
	data, err := embeddedAssets.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded file %s: %w", path, err)
	}

	// Use bytes.NewReader instead of strings.NewReader to avoid corrupting binary data
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode embedded PNG image: %w", err)
	}

	return img, nil
}

// HasTileAsset checks if a tile asset exists in embedded files
func (eam *EmbeddedAssetManager) HasTileAsset(tileType int32, playerID int32) bool {
	tilePath := fmt.Sprintf("v1/Tiles/%d/%d.png", tileType, playerID)
	_, err := embeddedAssets.ReadFile(tilePath)
	return err == nil
}

// HasUnitAsset checks if a unit asset exists in embedded files
func (eam *EmbeddedAssetManager) HasUnitAsset(unitType int32, playerColor int32) bool {
	unitPath := fmt.Sprintf("v1/Units/%d/%d.png", unitType, playerColor)
	_, err := embeddedAssets.ReadFile(unitPath)
	return err == nil
}

// PreloadCommonAssets preloads commonly used assets for better performance
func (eam *EmbeddedAssetManager) PreloadCommonAssets() error {
	// Preload common tile types (1-26) with basic player colors (0-5)
	for i := int32(1); i <= 26; i++ {
		for playerID := int32(0); playerID <= 5; playerID++ {
			if eam.HasTileAsset(i, playerID) {
				_, err := eam.GetTileImage(i, playerID)
				if err != nil {
					// Continue on error - not all combinations exist
					continue
				}
			}
		}
	}

	// Preload common unit types with basic player colors (0-5)
	for unitType := int32(1); unitType <= 44; unitType++ {
		for playerColor := int32(0); playerColor <= 5; playerColor++ {
			if eam.HasUnitAsset(unitType, playerColor) {
				_, err := eam.GetUnitImage(unitType, playerColor)
				if err != nil {
					// Continue on error - not all combinations exist
					continue
				}
			}
		}
	}

	eam.loaded = true
	return nil
}

// GetCacheStats returns statistics about cached assets
func (eam *EmbeddedAssetManager) GetCacheStats() (int, int) {
	eam.cacheMutex.RLock()
	defer eam.cacheMutex.RUnlock()

	return len(eam.tileCache), len(eam.unitCache)
}

// ClearCache clears all cached assets
func (eam *EmbeddedAssetManager) ClearCache() {
	eam.cacheMutex.Lock()
	defer eam.cacheMutex.Unlock()

	eam.tileCache = make(map[string]image.Image)
	eam.unitCache = make(map[string]image.Image)
}

// IsLoaded returns whether assets have been preloaded
func (eam *EmbeddedAssetManager) IsLoaded() bool {
	return eam.loaded
}
*/
