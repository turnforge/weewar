package weewar

import "image"

// =============================================================================
// Asset Manager Interface - Platform Abstraction
// =============================================================================

// AssetProvider defines the interface for loading game assets
// This allows switching between filesystem-based (CLI) and embedded (WASM) implementations
type AssetProvider interface {
	// Image loading
	GetTileImage(tileType int) (image.Image, error)
	GetUnitImage(unitType int, playerColor int) (image.Image, error)
	
	// Asset existence checks
	HasTileAsset(tileType int) bool
	HasUnitAsset(unitType int, playerColor int) bool
	
	// Performance optimization
	PreloadCommonAssets() error
	ClearCache()
	GetCacheStats() (int, int)
}

// Ensure our implementations satisfy the interface
var _ AssetProvider = (*AssetManager)(nil)
var _ AssetProvider = (*EmbeddedAssetManager)(nil)