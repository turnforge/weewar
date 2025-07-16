package weewar

import "image"

// =============================================================================
// TileLayer - Terrain Rendering
// =============================================================================

// TileLayer handles rendering of terrain tiles
type TileLayer struct {
	*BaseLayer
	terrainSprites map[int]image.Image // Cached terrain sprites
}

// NewTileLayer creates a new tile layer
func NewTileLayer(width, height int, scheduler LayerScheduler) *TileLayer {
	return &TileLayer{
		BaseLayer:      NewBaseLayer("terrain", width, height, scheduler),
		terrainSprites: make(map[int]image.Image),
	}
}

// Render renders terrain tiles to the layer buffer
func (tl *TileLayer) Render(world *World, options LayerRenderOptions) {
	if world == nil || world.Map == nil {
		return
	}

	// Clear buffer if full rebuild needed
	if tl.allDirty {
		tl.buffer.Clear()

		// Render all tiles
		for coord, tile := range world.Map.Tiles {
			if tile != nil {
				tl.renderTile(world, coord, tile, options)
			}
		}

		tl.allDirty = false
	} else {
		// Render only dirty tiles
		for coord := range tl.dirtyCoords {
			tile := world.Map.TileAt(coord)
			tl.renderTile(world, coord, tile, options)
		}
	}

	// Clear dirty tracking
	tl.ClearDirty()
}

// renderTile renders a single terrain tile
func (tl *TileLayer) renderTile(world *World, coord CubeCoord, tile *Tile, options LayerRenderOptions) {
	if tile == nil {
		return
	}

	// Get pixel position using Map's coordinate system
	x, y := world.Map.CenterXYForTile(coord, options.TileWidth, options.TileHeight, options.YIncrement)

	// Apply viewport offset
	x += float64(tl.x)
	y += float64(tl.y)

	// Try to use real terrain sprite if available
	if tl.assetProvider != nil && tl.assetProvider.HasTileAsset(tile.TileType) {
		tl.renderTerrainSprite(tile.TileType, x, y, options)
	} else {
		// Fallback to colored hexagon
		color := tl.getTerrainColor(tile.TileType)
		tl.drawSimpleHexToBuffer(x, y, color, options)
	}
}

// renderTerrainSprite renders a terrain sprite
func (tl *TileLayer) renderTerrainSprite(tileType int, x, y float64, options LayerRenderOptions) {
	// Check cache first
	cachedSprite, exists := tl.terrainSprites[tileType]
	if !exists {
		// Load and cache sprite
		img, err := tl.assetProvider.GetTileImage(tileType)
		if err != nil {
			// Fallback to colored hex
			color := tl.getTerrainColor(tileType)
			tl.drawSimpleHexToBuffer(x, y, color, options)
			return
		}
		tl.terrainSprites[tileType] = img
		cachedSprite = img
	}

	// Draw sprite to buffer
	tl.drawImageToBuffer(cachedSprite, x, y, options.TileWidth, options.TileHeight)
}

// getTerrainColor returns color for terrain type
func (tl *TileLayer) getTerrainColor(terrainType int) Color {
	switch terrainType {
	case 1: // Grass
		return Color{R: 0x22, G: 0x8B, B: 0x22}
	case 2: // Desert
		return Color{R: 0xEE, G: 0xCB, B: 0xAD}
	case 3: // Water
		return Color{R: 0x41, G: 0x69, B: 0xE1}
	case 4: // Mountain
		return Color{R: 0x8B, G: 0x89, B: 0x89}
	case 5: // Rock
		return Color{R: 0x69, G: 0x69, B: 0x69}
	default:
		return Color{R: 0xC8, G: 0xC8, B: 0xC8}
	}
}
