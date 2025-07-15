//go:build js && wasm
// +build js,wasm

package weewar

// =============================================================================
// CanvasRenderer - HTML Canvas Implementation for WASM
// =============================================================================

// CanvasRenderer implements WorldRenderer for direct HTML Canvas rendering in WASM builds.
// It provides high-performance rendering by drawing directly to the canvas without PNG encoding.
type CanvasRenderer struct {
	BaseRenderer
}

// NewCanvasRenderer creates a new Canvas-based renderer for WASM
func NewCanvasRenderer() *CanvasRenderer {
	return &CanvasRenderer{}
}

// RenderWorld renders the complete world state to a CanvasBuffer
func (cr *CanvasRenderer) RenderWorld(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
	// Clear the drawable surface
	drawable.Clear()

	// Render all layers directly using World data
	cr.RenderTerrain(world, viewState, drawable, options)
	
	// Render highlights if viewState is provided
	if viewState != nil {
		cr.RenderHighlights(world, viewState, drawable, options)
	}

	cr.RenderUnits(world, viewState, drawable, options)
	if options.ShowUI {
		cr.RenderUI(world, viewState, drawable, options)
	}
}

// RenderTerrain renders the terrain layer directly using World data
func (cr *CanvasRenderer) RenderTerrain(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
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
		terrainColor := cr.GetTerrainColor(tile.TileType)
		
		// Create hex shape for the tile
		hexPath := cr.createHexagonPath(x, y, options.TileWidth, options.TileHeight)
		
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

// RenderUnits renders the units layer directly using World data
func (cr *CanvasRenderer) RenderUnits(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
	cr.RenderUnitsWithAssets(world, viewState, drawable, options, nil)
}

// RenderUnitsWithAssets renders units using AssetManager when available, falling back to simple shapes
func (cr *CanvasRenderer) RenderUnitsWithAssets(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions, assetProvider AssetProvider) {
	if world == nil {
		return
	}

	// Render units for each player
	for _, playerUnits := range world.UnitsByPlayer {
		for _, unit := range playerUnits {
			if unit == nil {
				continue
			}

			// Use Map's CenterXYForTile method (Map handles origin internally)
			x, y := world.Map.CenterXYForTile(unit.Coord, options.TileWidth, options.TileHeight, options.YIncrement)

			// Try to load real unit asset first if AssetProvider is available
			if assetProvider != nil && assetProvider.HasUnitAsset(unit.UnitType, unit.PlayerID) {
				if unitImg, err := assetProvider.GetUnitImage(unit.UnitType, unit.PlayerID); err == nil {
					// Render real unit sprite (CenterXYForTile already returns centered coordinates)
					drawable.DrawImage(x-options.TileWidth/2, y-options.TileHeight/2, options.TileWidth, options.TileHeight, unitImg)

					// Add health indicator if unit is damaged
					if unit.AvailableHealth < 100 {
						cr.renderHealthBar(drawable, x, y, options.TileWidth, options.TileHeight, unit.AvailableHealth, 100)
					}
					continue
				}
			}

			// Fallback to simple colored circle if no asset available
			unitColor := cr.GetPlayerColor(unit.PlayerID)

			// Draw unit as a circle centered at the tile position
			radius := (options.TileWidth + options.TileHeight) / 8 // Smaller than hex
			circlePoints := cr.createCirclePoints(x, y, radius, 12)

			// Fill unit circle
			drawable.FillPath(circlePoints, unitColor)

			// Draw unit border
			borderColor := Color{R: 0, G: 0, B: 0, A: 255}
			strokeProps := StrokeProperties{Width: 2.0, LineCap: "round", LineJoin: "round"}
			drawable.StrokePath(circlePoints, borderColor, strokeProps)

			// Add health indicator if unit is damaged
			if unit.AvailableHealth < 100 {
				cr.renderHealthBar(drawable, x, y, options.TileWidth, options.TileHeight, unit.AvailableHealth, 100)
			}
		}
	}
}

// RenderHighlights renders selection highlights and movement indicators directly using World data
func (cr *CanvasRenderer) RenderHighlights(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
	if viewState == nil || world == nil || world.Map == nil {
		return
	}

	// Highlight movable tiles (green overlay)
	for _, coord := range viewState.MovableTiles {
		// Use Map's CenterXYForTile method (Map handles origin internally)
		x, y := world.Map.CenterXYForTile(coord, options.TileWidth, options.TileHeight, options.YIncrement)

		// Create hex shape for highlighting
		hexPath := cr.createHexagonPath(x, y, options.TileWidth, options.TileHeight)
		highlightColor := Color{R: 0, G: 255, B: 0, A: 64} // Transparent green
		drawable.FillPath(hexPath, highlightColor)
	}

	// Highlight attackable tiles (red overlay)
	for _, coord := range viewState.AttackableTiles {
		// Use Map's CenterXYForTile method (Map handles origin internally)
		x, y := world.Map.CenterXYForTile(coord, options.TileWidth, options.TileHeight, options.YIncrement)

		// Create hex shape for highlighting
		hexPath := cr.createHexagonPath(x, y, options.TileWidth, options.TileHeight)
		highlightColor := Color{R: 255, G: 0, B: 0, A: 64} // Transparent red
		drawable.FillPath(hexPath, highlightColor)
	}

	// Highlight selected unit (yellow border)
	if viewState.SelectedUnit != nil {
		unit := viewState.SelectedUnit
		// Use Map's CenterXYForTile method (Map handles origin internally)
		x, y := world.Map.CenterXYForTile(unit.Coord, options.TileWidth, options.TileHeight, options.YIncrement)

		// Create hex shape for highlighting
		hexPath := cr.createHexagonPath(x, y, options.TileWidth, options.TileHeight)
		selectionColor := Color{R: 255, G: 255, B: 0, A: 192} // Bright yellow
		strokeProps := StrokeProperties{Width: 3.0, LineCap: "round", LineJoin: "round"}
		drawable.StrokePath(hexPath, selectionColor, strokeProps)
	}
}

// RenderUI renders text overlays and UI elements directly using World data
func (cr *CanvasRenderer) RenderUI(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
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

			// Use Map's CenterXYForTile method (Map handles origin internally)
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

	// Render brush preview in editor mode
	if viewState.HoveredTile != nil && viewState.BrushSize >= 0 {
		// Show brush preview at hovered tile - note: HoveredTile should be updated to use CubeCoord
		// For now, assume it has a Coord field that's CubeCoord
		hoveredCoord := viewState.HoveredTile.Coord // This may need updating when we update ViewState

		// Use Map's CenterXYForTile method (Map handles origin internally)
		x, y := world.Map.CenterXYForTile(hoveredCoord, options.TileWidth, options.TileHeight, options.YIncrement)

		// Create hex shape for brush preview
		hexPath := cr.createHexagonPath(x, y, options.TileWidth, options.TileHeight)
		brushColor := Color{R: 255, G: 255, B: 255, A: 128}
		strokeProps := StrokeProperties{
			Width:       2.0,
			LineCap:     "round",
			LineJoin:    "round",
			DashPattern: []float64{5.0, 5.0}, // Dotted line
		}
		drawable.StrokePath(hexPath, brushColor, strokeProps)
	}
}

// =============================================================================
// Canvas-Specific Utility Functions
// =============================================================================

// createCirclePoints creates points for a circle approximation optimized for canvas rendering
func (cr *CanvasRenderer) createCirclePoints(centerX, centerY, radius float64, segments int) []Point {
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

// renderHealthBar renders a health bar for a unit
func (cr *CanvasRenderer) renderHealthBar(drawable Drawable, x, y, tileWidth, tileHeight float64, currentHealth, maxHealth int) {
	if currentHealth >= maxHealth {
		return // Don't render health bar for full health
	}
	
	// Calculate health bar dimensions
	barWidth := tileWidth * 0.8
	barHeight := 6.0
	barX := x - barWidth/2
	barY := y + tileHeight/2 - barHeight - 2
	
	// Background bar (red)
	backgroundBar := []Point{
		{X: barX, Y: barY},
		{X: barX + barWidth, Y: barY},
		{X: barX + barWidth, Y: barY + barHeight},
		{X: barX, Y: barY + barHeight},
	}
	redColor := Color{R: 255, G: 0, B: 0, A: 255}
	drawable.FillPath(backgroundBar, redColor)
	
	// Health bar (green, proportional to health)
	healthRatio := float64(currentHealth) / float64(maxHealth)
	healthWidth := barWidth * healthRatio
	healthBar := []Point{
		{X: barX, Y: barY},
		{X: barX + healthWidth, Y: barY},
		{X: barX + healthWidth, Y: barY + barHeight},
		{X: barX, Y: barY + barHeight},
	}
	greenColor := Color{R: 0, G: 255, B: 0, A: 255}
	drawable.FillPath(healthBar, greenColor)
	
	// Health bar border
	borderColor := Color{R: 0, G: 0, B: 0, A: 255}
	strokeProps := StrokeProperties{Width: 1.0, LineCap: "round", LineJoin: "round"}
	drawable.StrokePath(backgroundBar, borderColor, strokeProps)
}