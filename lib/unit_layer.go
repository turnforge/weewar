package weewar

import (
	"fmt"
	"image"
	"image/color"
)

// =============================================================================
// UnitLayer - Unit Rendering
// =============================================================================

// UnitLayer handles rendering of units
type UnitLayer struct {
	*BaseLayer
	unitSprites map[string]image.Image // Cached unit sprites by "type_player"
}

// GridLayer handles rendering of hex grid lines and coordinates
type GridLayer struct {
	*BaseLayer
}

// NewUnitLayer creates a new unit layer
func NewUnitLayer(width, height int, scheduler LayerScheduler) *UnitLayer {
	return &UnitLayer{
		BaseLayer:   NewBaseLayer("units", width, height, scheduler),
		unitSprites: make(map[string]image.Image),
	}
}

// Render renders units to the layer buffer
func (ul *UnitLayer) Render(world *World, options LayerRenderOptions) {
	if world == nil {
		return
	}

	// Clear buffer if full rebuild needed
	if ul.allDirty {
		ul.buffer.Clear()

		// Render all units from all players
		for _, playerUnits := range world.UnitsByPlayer {
			for _, unit := range playerUnits {
				if unit != nil {
					ul.renderUnit(world, unit, options)
				}
			}
		}

		ul.allDirty = false
	} else {
		// Clear and render only dirty unit positions
		for coord := range ul.dirtyCoords {
			ul.clearHexArea(coord, options)

			// Find unit at this position
			unit := ul.findUnitAt(world, coord)
			if unit != nil {
				ul.renderUnit(world, unit, options)
			}
		}
	}

	// Clear dirty tracking
	ul.ClearDirty()
}

// renderUnit renders a single unit
func (ul *UnitLayer) renderUnit(world *World, unit *Unit, options LayerRenderOptions) {
	// Get pixel position using Map's coordinate system
	x, y := world.Map.CenterXYForTile(unit.Coord, options.TileWidth, options.TileHeight, options.YIncrement)

	// Apply viewport offset
	x += float64(ul.x)
	y += float64(ul.y)

	// Try to use real unit sprite if available
	if ul.assetProvider != nil && ul.assetProvider.HasUnitAsset(unit.UnitType, unit.PlayerID) {
		ul.renderUnitSprite(unit.UnitType, unit.PlayerID, x, y, options)
	} else {
		// Fallback to colored circle
		ul.drawSimpleUnitToBuffer(x, y, unit.PlayerID, options)
	}
}

// renderUnitSprite renders a unit sprite
func (ul *UnitLayer) renderUnitSprite(unitType, playerID int, x, y float64, options LayerRenderOptions) {
	// Check cache first
	spriteKey := fmt.Sprintf("%d_%d", unitType, playerID)
	cachedSprite, exists := ul.unitSprites[spriteKey]
	if !exists {
		// Load and cache sprite
		img, err := ul.assetProvider.GetUnitImage(unitType, playerID)
		if err != nil {
			// Fallback to colored circle
			ul.drawSimpleUnitToBuffer(x, y, playerID, options)
			return
		}
		ul.unitSprites[spriteKey] = img
		cachedSprite = img
	}

	// Draw sprite to buffer
	ul.drawImageToBuffer(cachedSprite, x, y, options.TileWidth, options.TileHeight)
}

// drawSimpleUnitToBuffer draws a colored circle for a unit
func (ul *UnitLayer) drawSimpleUnitToBuffer(x, y float64, playerID int, options LayerRenderOptions) {
	// Get player color
	var unitColor Color
	switch playerID {
	case 0:
		unitColor = Color{R: 255, G: 0, B: 0, A: 255} // Red
	case 1:
		unitColor = Color{R: 0, G: 0, B: 255, A: 255} // Blue
	case 2:
		unitColor = Color{R: 0, G: 255, B: 0, A: 255} // Green
	case 3:
		unitColor = Color{R: 255, G: 255, B: 0, A: 255} // Yellow
	default:
		unitColor = Color{R: 128, G: 128, B: 128, A: 255} // Gray
	}

	bufferImg := ul.buffer.GetImageData()

	// Draw smaller ellipse for units (60% of tile size)
	radiusX := int(options.TileWidth * 0.3)
	radiusY := int(options.TileHeight * 0.3)
	centerX, centerY := int(x), int(y)

	for dy := -radiusY; dy <= radiusY; dy++ {
		for dx := -radiusX; dx <= radiusX; dx++ {
			if float64(dx*dx)/float64(radiusX*radiusX)+float64(dy*dy)/float64(radiusY*radiusY) <= 1.0 {
				px, py := centerX+dx, centerY+dy
				if px >= 0 && py >= 0 && px < ul.width && py < ul.height {
					rgba := color.RGBA{R: unitColor.R, G: unitColor.G, B: unitColor.B, A: unitColor.A}
					bufferImg.Set(px, py, rgba)
				}
			}
		}
	}
}

// clearHexArea clears a hexagonal area at the given coordinate
func (ul *UnitLayer) clearHexArea(coord CubeCoord, options LayerRenderOptions) {
	// For now, just clear the entire buffer - can optimize later
	ul.buffer.Clear()
}

// findUnitAt finds a unit at the given coordinate
func (ul *UnitLayer) findUnitAt(world *World, coord CubeCoord) *Unit {
	for _, playerUnits := range world.UnitsByPlayer {
		for _, unit := range playerUnits {
			if unit != nil && unit.Coord == coord {
				return unit
			}
		}
	}
	return nil
}
