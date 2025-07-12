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

// RenderWorld renders the complete world state to a CanvasBuffer using the same proven architecture as BufferRenderer
func (cr *CanvasRenderer) RenderWorld(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
	cr.RenderWorldWithAssets(world, viewState, drawable, options, nil)
}

// RenderWorldWithAssets renders the complete world state with AssetManager support (matches BufferRenderer)
func (cr *CanvasRenderer) RenderWorldWithAssets(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions, originalGame *Game) {
	// Clear the canvas
	drawable.Clear()
	
	// Create temporary Game instance to access proven rendering methods with AssetManager
	game := cr.CreateGameForRenderingWithAssets(world, originalGame)
	
	// Use the Game's proven rendering methods for terrain (SAME as BufferRenderer)
	game.RenderTerrainTo(drawable, options.TileWidth, options.TileHeight, options.YIncrement)
	
	// Render highlights if viewState is provided
	if viewState != nil {
		cr.RenderHighlights(world, viewState, drawable, options)
	}
	
	// Use Game's generic methods that work with any Drawable (SAME for all platforms)
	game.RenderUnitsTo(drawable, options.TileWidth, options.TileHeight, options.YIncrement)
	if options.ShowUI {
		game.RenderUITo(drawable, options.TileWidth, options.TileHeight, options.YIncrement)
	}
}

// RenderTerrain renders the terrain layer using Game's proven methods with asset support (matches BufferRenderer)
func (cr *CanvasRenderer) RenderTerrain(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
	if world == nil || world.Map == nil {
		return
	}
	
	// Create temporary Game instance and use its proven terrain rendering
	game := cr.CreateGameForRendering(world)
	game.RenderTerrainTo(drawable, options.TileWidth, options.TileHeight, options.YIncrement)
}

// RenderTerrainWithAssets renders terrain using AssetManager when available (matches BufferRenderer)
func (cr *CanvasRenderer) RenderTerrainWithAssets(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions, originalGame *Game) {
	if world == nil || world.Map == nil {
		return
	}
	
	// Create temporary Game instance with preserved AssetManager
	game := cr.CreateGameForRenderingWithAssets(world, originalGame)
	game.RenderTerrainTo(drawable, options.TileWidth, options.TileHeight, options.YIncrement)
}

// RenderUnits renders the units layer using Game's proven coordinate methods with asset support (matches BufferRenderer)
func (cr *CanvasRenderer) RenderUnits(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
	cr.RenderUnitsWithAssets(world, viewState, drawable, options, nil)
}

// RenderUnitsWithAssets renders units using AssetManager when available, falling back to simple shapes (matches BufferRenderer)
func (cr *CanvasRenderer) RenderUnitsWithAssets(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions, originalGame *Game) {
	if world == nil {
		return
	}
	
	// Create temporary Game instance to access proven coordinate methods and AssetManager
	game := cr.CreateGameForRenderingWithAssets(world, originalGame)
	
	for _, unit := range world.Units {
		if unit == nil {
			continue
		}
		
		// Use Game's proven XYForTile method for coordinate calculation
		x, y := game.Map.XYForTile(unit.Row, unit.Col, options.TileWidth, options.TileHeight, options.YIncrement)
		
		// Try to load real unit asset first (like the original Game.RenderUnits method)
		assetManager := game.GetAssetManager()
		if assetManager != nil && assetManager.HasUnitAsset(unit.UnitType, unit.PlayerID) {
			if unitImg, err := assetManager.GetUnitImage(unit.UnitType, unit.PlayerID); err == nil {
				// Render real unit sprite (XYForTile already returns centered coordinates)
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
	}
}

// RenderHighlights renders selection highlights and movement indicators using Game's proven hex methods (matches BufferRenderer)
func (cr *CanvasRenderer) RenderHighlights(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
	if viewState == nil || world == nil || world.Map == nil {
		return
	}
	
	// Create temporary Game instance to access proven coordinate and hex methods
	game := cr.CreateGameForRendering(world)
	
	// Highlight movable tiles (green overlay)
	for _, pos := range viewState.MovableTiles {
		// Use Game's proven XYForTile method for coordinate calculation
		x, y := game.Map.XYForTile(pos.Row, pos.Col, options.TileWidth, options.TileHeight, options.YIncrement)
		
		// Use Game's proven createHexagonPath method
		hexPath := game.createHexagonPath(x, y, options.TileWidth, options.TileHeight, options.YIncrement)
		highlightColor := Color{R: 0, G: 255, B: 0, A: 64} // Transparent green
		drawable.FillPath(hexPath, highlightColor)
	}
	
	// Highlight attackable tiles (red overlay)
	for _, pos := range viewState.AttackableTiles {
		// Use Game's proven XYForTile method for coordinate calculation
		x, y := game.Map.XYForTile(pos.Row, pos.Col, options.TileWidth, options.TileHeight, options.YIncrement)
		
		// Use Game's proven createHexagonPath method
		hexPath := game.createHexagonPath(x, y, options.TileWidth, options.TileHeight, options.YIncrement)
		highlightColor := Color{R: 255, G: 0, B: 0, A: 64} // Transparent red
		drawable.FillPath(hexPath, highlightColor)
	}
	
	// Highlight selected unit (yellow border)
	if viewState.SelectedUnit != nil {
		unit := viewState.SelectedUnit
		// Use Game's proven XYForTile method for coordinate calculation
		x, y := game.Map.XYForTile(unit.Row, unit.Col, options.TileWidth, options.TileHeight, options.YIncrement)
		
		// Use Game's proven createHexagonPath method
		hexPath := game.createHexagonPath(x, y, options.TileWidth, options.TileHeight, options.YIncrement)
		selectionColor := Color{R: 255, G: 255, B: 0, A: 192} // Bright yellow
		strokeProps := StrokeProperties{Width: 3.0, LineCap: "round", LineJoin: "round"}
		drawable.StrokePath(hexPath, selectionColor, strokeProps)
	}
}

// RenderUI renders text overlays and UI elements using Game's proven coordinate methods (matches BufferRenderer)
func (cr *CanvasRenderer) RenderUI(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
	if world == nil {
		return
	}
	
	// Create temporary Game instance to access proven coordinate methods
	game := cr.CreateGameForRendering(world)
	
	// Render coordinate labels if enabled
	if options.ShowCoordinates && world.Map != nil {
		textColor := Color{R: 255, G: 255, B: 255, A: 255}
		backgroundColor := Color{R: 0, G: 0, B: 0, A: 128}
		
		for coord, tile := range world.Map.Tiles {
			if tile == nil {
				continue
			}
			
			displayRow, displayCol := world.Map.HexToDisplay(coord)
			// Use Game's proven XYForTile method for coordinate calculation
			x, y := game.Map.XYForTile(displayRow, displayCol, options.TileWidth, options.TileHeight, options.YIncrement)
			
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
		// Show brush preview at hovered tile
		hoveredRow := viewState.HoveredTile.Row
		hoveredCol := viewState.HoveredTile.Col
		
		// Use Game's proven XYForTile method for coordinate calculation
		x, y := game.Map.XYForTile(hoveredRow, hoveredCol, options.TileWidth, options.TileHeight, options.YIncrement)
		
		// Use Game's proven createHexagonPath method
		hexPath := game.createHexagonPath(x, y, options.TileWidth, options.TileHeight, options.YIncrement)
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

// renderHealthBar renders a health bar for a unit (matches BufferRenderer)
func (cr *CanvasRenderer) renderHealthBar(drawable Drawable, x, y, tileWidth, tileHeight float64, currentHealth, maxHealth int) {
	if currentHealth >= maxHealth {
		return // Don't render health bar for full health
	}
	
	// Calculate health bar dimensions (matches Game's renderHealthBar)
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

// Note: formatCoordinate function is defined in world_renderer.go and shared