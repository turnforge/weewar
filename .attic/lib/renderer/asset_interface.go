package weewar

import (
	"image"

	"github.com/panyam/turnengine/games/weewar/assets"
)

// =============================================================================
// Asset Manager Interface - Platform Abstraction
// =============================================================================

// AssetProvider defines the interface for loading game assets
// This allows switching between filesystem-based (CLI) and embedded (WASM) implementations
type AssetProvider interface {
	// Image loading
	GetTileImage(tileType int32, playerID int32) (image.Image, error)
	GetUnitImage(unitType int32, playerColor int32) (image.Image, error)

	// Asset existence checks
	HasTileAsset(tileType int32, playerID int32) bool
	HasUnitAsset(unitType int32, playerColor int32) bool

	// Performance optimization
	PreloadCommonAssets() error
	ClearCache()
	GetCacheStats() (int, int)
}

// Ensure our implementations satisfy the interface
var _ AssetProvider = (*AssetManager)(nil)
var _ AssetProvider = (*assets.EmbeddedAssetManager)(nil)
