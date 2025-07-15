package weewar

// =============================================================================
// BufferRenderer - PNG/File Output Implementation
// Moved from world_renderer.go
// =============================================================================

// BufferRenderer implements WorldRenderer for PNG file output and CLI usage.
// It maintains compatibility with the existing Buffer-based rendering system.
type BufferRenderer struct {
	BaseRenderer
}

// NewBufferRenderer creates a new Buffer-based renderer
func NewBufferRenderer() *BufferRenderer {
	return &BufferRenderer{}
}

// RenderWorld renders the complete world state to a drawable surface
func (br *BufferRenderer) RenderWorld(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
	// Clear the drawable surface
	drawable.Clear()

	// Render all layers directly using World data
	br.RenderTerrain(world, viewState, drawable, options)

	// Render highlights if viewState is provided
	if viewState != nil {
		br.RenderHighlights(world, viewState, drawable, options)
	}

	br.RenderUnits(world, viewState, drawable, options)
	if options.ShowUI {
		br.RenderUI(world, viewState, drawable, options)
	}
}

// RenderTerrain renders the terrain layer directly using World data
func (br *BufferRenderer) RenderTerrain(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
	if world == nil || world.Map == nil {
		return
	}

	// Render each terrain tile directly using Map's coordinate system
	for coord, tile := range world.Map.Tiles {
		if tile == nil {
			continue
		}

		// Use Map's CenterXYForTile method (Map handles origin internally)
		x, y := world.Map.CenterXYForTile(coord, options.TileWidth, options.TileHeight, options.YIncrement)

		// Get terrain color based on tile type
		terrainColor := br.GetTerrainColor(tile.TileType)

		// Create hex shape for the tile
		hexPath := br.createHexagonPath(x, y, options.TileWidth, options.TileHeight)

		// Fill the hex with terrain color
		drawable.FillPath(hexPath, terrainColor)

		// Add border if grid is enabled
		if options.ShowGrid {
			borderColor := Color{R: 64, G: 64, B: 64, A: 255} // Dark gray border
			strokeProps := StrokeProperties{Width: 1.0, LineCap: "round", LineJoin: "round"}
			drawable.StrokePath(hexPath, borderColor, strokeProps)
		}
	}
}

// RenderUnits renders the units layer directly using World data with asset support
func (br *BufferRenderer) RenderUnits(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
	br.RenderUnitsWithAssets(world, viewState, drawable, options, nil)
}

// RenderUnitsWithAssets renders units using AssetManager when available, falling back to simple shapes
func (br *BufferRenderer) RenderUnitsWithAssets(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions, assetProvider AssetProvider) {
	if world == nil {
		return
	}

	// Render units for each player
	for _, playerUnits := range world.UnitsByPlayer {
		for _, unit := range playerUnits {
			if unit == nil {
				continue
			}

			// Use Map's CenterXYForTile method with the Map's origin
			x, y := world.Map.CenterXYForTile(unit.Coord, options.TileWidth, options.TileHeight, options.YIncrement)

			// Try to load real unit asset first if AssetProvider is available
			if assetProvider != nil && assetProvider.HasUnitAsset(unit.UnitType, unit.PlayerID) {
				if unitImg, err := assetProvider.GetUnitImage(unit.UnitType, unit.PlayerID); err == nil {
					// Render real unit sprite (CenterXYForTile already returns centered coordinates)
					drawable.DrawImage(x-options.TileWidth/2, y-options.TileHeight/2, options.TileWidth, options.TileHeight, unitImg)

					// Add health indicator if unit is damaged
					if unit.AvailableHealth < 100 {
						br.renderHealthBar(drawable, x, y, options.TileWidth, options.TileHeight, unit.AvailableHealth, 100)
					}
					continue
				}
			}

			// Fallback to simple colored circle if no asset available
			unitColor := br.GetPlayerColor(unit.PlayerID)

			// Draw unit as a circle centered at the tile position
			radius := (options.TileWidth + options.TileHeight) / 8 // Smaller than hex
			circlePoints := br.createCirclePoints(x, y, radius, 12)

			// Fill unit circle
			drawable.FillPath(circlePoints, unitColor)

			// Draw unit border
			borderColor := Color{R: 0, G: 0, B: 0, A: 255}
			strokeProps := StrokeProperties{Width: 2.0, LineCap: "round", LineJoin: "round"}
			drawable.StrokePath(circlePoints, borderColor, strokeProps)

			// Add health indicator if unit is damaged
			if unit.AvailableHealth < 100 {
				br.renderHealthBar(drawable, x, y, options.TileWidth, options.TileHeight, unit.AvailableHealth, 100)
			}
		}
	}
}

// RenderHighlights renders selection highlights and movement indicators directly using World data
func (br *BufferRenderer) RenderHighlights(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
	if viewState == nil || world == nil || world.Map == nil {
		return
	}

	// Highlight movable tiles (green overlay)
	for _, coord := range viewState.MovableTiles {
		// Use Map's CenterXYForTile method (Map handles origin internally)
		x, y := world.Map.CenterXYForTile(coord, options.TileWidth, options.TileHeight, options.YIncrement)

		// Create hex shape for highlighting
		hexPath := br.createHexagonPath(x, y, options.TileWidth, options.TileHeight)
		highlightColor := Color{R: 0, G: 255, B: 0, A: 64} // Transparent green
		drawable.FillPath(hexPath, highlightColor)
	}

	// Highlight attackable tiles (red overlay)
	for _, coord := range viewState.AttackableTiles {
		// Use Map's CenterXYForTile method (Map handles origin internally)
		x, y := world.Map.CenterXYForTile(coord, options.TileWidth, options.TileHeight, options.YIncrement)

		// Create hex shape for highlighting
		hexPath := br.createHexagonPath(x, y, options.TileWidth, options.TileHeight)
		highlightColor := Color{R: 255, G: 0, B: 0, A: 64} // Transparent red
		drawable.FillPath(hexPath, highlightColor)
	}
}

// RenderUI renders text overlays and UI elements directly using World data
func (br *BufferRenderer) RenderUI(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
	if world == nil {
		return
	}

	// Render coordinate labels if enabled
	if options.ShowCoordinates && world.Map != nil {
		textColor := Color{R: 255, G: 255, B: 255, A: 255}
		backgroundColor := Color{R: 0, G: 0, B: 0, A: 128}

		for coord, tile := range world.Map.Tiles {
			if tile == nil {
				continue
			}

			// Use Map's CenterXYForTile method with the Map's origin
			x, y := world.Map.CenterXYForTile(coord, options.TileWidth, options.TileHeight, options.YIncrement)

			// Draw coordinate text
			coordText := formatCoordinate(coord)
			fontSize := options.TileWidth / 8 // Scale font with tile size
			if fontSize < 8 {
				fontSize = 8
			}

			drawable.DrawTextWithStyle(x-10, y, coordText, fontSize, textColor, false, backgroundColor)
		}
	}
}

// renderHealthBar renders a health bar for a unit using Game's proven method
func (br *BufferRenderer) renderHealthBar(drawable Drawable, x, y, tileWidth, tileHeight float64, currentHealth, maxHealth int) {
	// For now, only render health bars on Buffer (since Game's method expects Buffer)
	if buffer, ok := drawable.(*Buffer); ok {
		// Create a temporary game to access the renderHealthBar method
		tempGame := &Game{}
		tempGame.renderHealthBar(buffer, x, y, tileWidth, tileHeight, currentHealth, maxHealth)
	}
	// For non-Buffer drawables, we could implement a simplified health bar here if needed
}

// =============================================================================
// Utility Functions
// =============================================================================

// createCirclePoints creates points for a circle approximation (used for units)
func (br *BufferRenderer) createCirclePoints(centerX, centerY, radius float64, segments int) []Point {
	points := make([]Point, segments)
	for i := 0; i < segments; i++ {
		angle := float64(i) * 360.0 / float64(segments)
		angleRad := angle * 3.14159 / 180.0

		x := centerX + radius*cosApprox(angleRad)
		y := centerY + radius*sinApprox(angleRad)

		points[i] = Point{X: x, Y: y}
	}
	return points
}
