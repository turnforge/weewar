package rendering

import (
	"fmt"
	"image"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

// =============================================================================
// TileLayer - Terrain Rendering
// =============================================================================

// TileLayer handles rendering of terrain tiles
type TileLayer struct {
	*BaseLayer
	terrainSprites map[int32]image.Image // Cached terrain sprites
}

// NewTileLayer creates a new tile layer
func NewTileLayer(width, height int, scheduler LayerScheduler) *TileLayer {
	return &TileLayer{
		BaseLayer:      NewBaseLayer("terrain", width, height, scheduler),
		terrainSprites: make(map[int32]image.Image),
	}
}

// Render renders terrain tiles to the layer buffer
func (tl *TileLayer) Render(world *World, options LayerRenderOptions) {
	if world == nil {
		return
	}

	fmt.Println("0. Dirty, Changed Tiles: ", tl.allDirty, world.tilesByCoord)
	// Clear buffer if full rebuild needed
	if tl.allDirty {
		tl.buffer.Clear()

		// Render all tiles
		for coord, tile := range world.tilesByCoord {
			if tile != nil {
				tl.renderTile(world, coord, tile, options)
			}
		}

		tl.allDirty = false
	} else {
		// Render only dirty tiles
		for coord := range tl.dirtyCoords {
			tile := world.TileAt(coord)
			tl.renderTile(world, coord, tile, options)
		}
	}

	// Clear dirty tracking
	tl.ClearDirty()
}

// renderTile renders a single terrain tile
func (tl *TileLayer) renderTile(world *World, coord AxialCoord, tile *v1.Tile, options LayerRenderOptions) {
	if tile == nil {
		return
	}

	// Get pixel position using privateMap's coordinate system
	x, y := world.CenterXYForTile(coord, options.TileWidth, options.TileHeight, options.YIncrement)

	// Apply viewport offset
	x -= float64(tl.X)
	y -= float64(tl.Y)

	// Check if this is a player-owned tile (base, city, etc.)
	playerID := tile.Player // -1 means neutral/no player

	// Try to use real terrain sprite if available
	// log.Println("Has Tile: ", tile.TileType, playerID, tl.assetProvider.HasTileAsset(tile.TileType, playerID))
	if tl.assetProvider != nil && tl.assetProvider.HasTileAsset(tile.TileType, playerID) {
		tl.renderTerrainSprite(tile.TileType, playerID, x, y, options)
	} else {
		// Fallback to colored hexagon
		color := tl.getTerrainColor(tile.TileType, playerID)
		tl.drawSimpleHexToBuffer(x, y, color, options)
	}
}

// isPlayerOwnedTerrain checks if a terrain type is player-owned (bases, cities, etc.)
func (tl *TileLayer) isPlayerOwnedTerrain(tileType int32) bool {
	switch tileType {
	case 1: // Land Base
		return true
	case 2: // Naval Base
		return true
	case 3: // Airport Base
		return true
	case 6: // Hospital
		return true
	case 16: // Missile Silo
		return true
	case 20: // Mines
		return true
	case 21: // City
		return true
	case 25: // Guard Tower
		return true
	default:
		return false
	}
}

// renderTerrainSprite renders a terrain sprite
func (tl *TileLayer) renderTerrainSprite(tileType int32, playerID int32, x, y float64, options LayerRenderOptions) {
	// Create cache key that includes player ID for player-owned terrain
	cacheKey := tileType
	if tl.isPlayerOwnedTerrain(tileType) {
		cacheKey = tileType*1000 + playerID // Simple way to make unique key
	}

	// Check cache first
	cachedSprite, exists := tl.terrainSprites[cacheKey]
	if !exists {
		// Load and cache sprite
		img, err := tl.assetProvider.GetTileImage(tileType, playerID)
		if err != nil {
			// Fallback to colored hex
			color := tl.getTerrainColor(tileType, playerID)
			tl.drawSimpleHexToBuffer(x, y, color, options)
			return
		}
		tl.terrainSprites[cacheKey] = img
		cachedSprite = img
	}

	// Draw sprite to buffer
	tl.drawImageToBuffer(cachedSprite, x, y, options.TileWidth, options.TileHeight)
}

// getTerrainColor returns color for terrain type
func (tl *TileLayer) getTerrainColor(terrainType int32, playerID int32) Color {
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
