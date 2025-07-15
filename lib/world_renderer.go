package weewar

// =============================================================================
// WorldRenderer Interface - Platform-Agnostic Rendering
// =============================================================================

// WorldRenderer provides platform-agnostic rendering of game worlds.
// This interface abstracts away the differences between Buffer (PNG) and CanvasBuffer (HTML Canvas).
type WorldRenderer interface {
	// RenderWorld renders the complete world state to the given drawable surface
	RenderWorld(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions)

	// RenderTerrain renders only the terrain layer
	RenderTerrain(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions)

	// RenderUnits renders only the units layer
	RenderUnits(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions)

	// RenderHighlights renders selection highlights and movement indicators
	RenderHighlights(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions)

	// RenderUI renders text overlays and UI elements
	RenderUI(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions)
}

// WorldRenderOptions contains all rendering configuration parameters for the World-Renderer architecture
type WorldRenderOptions struct {
	// Canvas dimensions
	CanvasWidth  int `json:"canvasWidth"`
	CanvasHeight int `json:"canvasHeight"`

	// Hex grid parameters
	TileWidth  float64 `json:"tileWidth"`  // Width of each hex tile in pixels
	TileHeight float64 `json:"tileHeight"` // Height of each hex tile in pixels
	YIncrement float64 `json:"yIncrement"` // Vertical spacing between hex rows (typically tileHeight * 0.75)

	// Visual options
	ShowGrid        bool `json:"showGrid"`        // Whether to render hex grid lines
	ShowCoordinates bool `json:"showCoordinates"` // Whether to show coordinate labels
	ShowPaths       bool `json:"showPaths"`       // Whether to show movement paths
	ShowUI          bool `json:"showUI"`          // Whether to show UI elements (current player indicator, etc.)

	// Rendering quality
	HighQuality bool `json:"highQuality"` // Whether to use high-quality rendering (affects performance)
}

// =============================================================================
// BaseRenderer - Common Hex Grid Logic
// =============================================================================

// BaseRenderer provides common rendering utilities that work directly with World data.
// This ensures proper separation of concerns between Game (flow control) and World (pure state).
type BaseRenderer struct{}

// GetPlayerColor returns the color for a given player ID
func (br *BaseRenderer) GetPlayerColor(playerID int) Color {
	playerColors := []Color{
		{R: 255, G: 0, B: 0, A: 255},   // Player 0 - Red
		{R: 0, G: 0, B: 255, A: 255},   // Player 1 - Blue
		{R: 0, G: 255, B: 0, A: 255},   // Player 2 - Green
		{R: 255, G: 255, B: 0, A: 255}, // Player 3 - Yellow
		{R: 255, G: 0, B: 255, A: 255}, // Player 4 - Magenta
		{R: 0, G: 255, B: 255, A: 255}, // Player 5 - Cyan
	}

	if playerID >= 0 && playerID < len(playerColors) {
		return playerColors[playerID]
	}

	// Default color for invalid player IDs
	return Color{R: 128, G: 128, B: 128, A: 255}
}

// GetTerrainColor returns the color for a given terrain type
func (br *BaseRenderer) GetTerrainColor(terrainType int) Color {
	terrainColors := []Color{
		{R: 64, G: 64, B: 64, A: 255},     // 0 - Unknown (dark gray)
		{R: 34, G: 139, B: 34, A: 255},    // 1 - Grass (forest green)
		{R: 238, G: 203, B: 173, A: 255},  // 2 - Desert (sandy brown)
		{R: 65, G: 105, B: 225, A: 255},   // 3 - Water (royal blue)
		{R: 139, G: 69, B: 19, A: 255},    // 4 - Mountain (saddle brown)
		{R: 105, G: 105, B: 105, A: 255},  // 5 - Rock (dim gray)
	}

	if terrainType >= 0 && terrainType < len(terrainColors) {
		return terrainColors[terrainType]
	}

	// Default to unknown terrain color
	return terrainColors[0]
}

// createHexagonPath creates a hexagon path for rendering
func (br *BaseRenderer) createHexagonPath(centerX, centerY, tileWidth, tileHeight float64) []Point {
	// Create a hexagon using the tile dimensions
	// For pointy-topped hexagons, use tileWidth as the radius
	radius := tileWidth / 2.0
	points := make([]Point, 6)
	
	// Generate 6 points for a pointy-topped hexagon
	for i := 0; i < 6; i++ {
		angle := float64(i) * 60.0 * 3.14159 / 180.0 // 60 degrees in radians
		x := centerX + radius * cosApprox(angle)
		y := centerY + radius * sinApprox(angle)
		points[i] = Point{X: x, Y: y}
	}
	
	return points
}

// CalculateRenderOptions creates appropriate render options based on canvas size and map dimensions
func (br *BaseRenderer) CalculateRenderOptions(canvasWidth, canvasHeight int, world *World) WorldRenderOptions {
	if world == nil || world.Map == nil {
		// Default options for empty world
		return WorldRenderOptions{
			CanvasWidth:  canvasWidth,
			CanvasHeight: canvasHeight,
			TileWidth:    DefaultTileWidth,
			TileHeight:   DefaultTileHeight,
			YIncrement:   DefaultYIncrement,
			ShowGrid:     true,
		}
	}

	// Use standard tile dimensions from Game class, then calculate proper scaling
	baseTileWidth := DefaultTileWidth
	baseTileHeight := DefaultTileHeight
	baseYIncrement := DefaultYIncrement

	// Calculate actual map bounds using the Map's proper hex geometry
	// minX, minY, maxX, maxY := world.Map.getMapBounds(baseTileWidth, baseTileHeight, baseYIncrement)

	// Calculate the actual dimensions needed for the map
	// mapPixelWidth := maxX - minX
	// mapPixelHeight := maxY - minY

	// Calculate scaling factors to fit the map in the canvas
	scaleX := 1.0 // float64(canvasWidth) / mapPixelWidth
	scaleY := 1.0 // float64(canvasHeight) / mapPixelHeight

	// Use the smaller scale factor to ensure the entire map fits
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// Apply scaling to tile dimensions
	tileWidth := baseTileWidth * scale
	tileHeight := baseTileHeight * scale
	yIncrement := baseYIncrement * scale

	// Ensure minimum tile size for visibility
	minTileSize := 20.0
	if tileWidth < minTileSize {
		scaleFactor := minTileSize / tileWidth
		tileWidth = minTileSize
		tileHeight = tileHeight * scaleFactor
		yIncrement = yIncrement * scaleFactor
	}

	return WorldRenderOptions{
		CanvasWidth:     canvasWidth,
		CanvasHeight:    canvasHeight,
		TileWidth:       tileWidth,
		TileHeight:      tileHeight,
		YIncrement:      yIncrement,
		ShowGrid:        true,
		ShowCoordinates: false,
		ShowPaths:       true,
		ShowUI:          false, // Disable UI elements for static map renders
		HighQuality:     true,
	}
}

// =============================================================================
// BufferRenderer - PNG/File Output Implementation
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
			x, y := world.Map.CenterXYForTile(unit.Coord, options.TileWidth, options.TileHeight, options.YIncrement, world.Map.OriginX, world.Map.OriginY)

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
			x, y := world.Map.CenterXYForTile(coord, options.TileWidth, options.TileHeight, options.YIncrement, world.Map.OriginX, world.Map.OriginY)

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

// formatCoordinate formats cube coordinates for display
func formatCoordinate(coord CubeCoord) string {
	return ""
	// Simplified - return empty string to avoid text rendering complexity for now
	// Can be enhanced later: return fmt.Sprintf("%d,%d", coord.Q, coord.R)
}

// Mathematical helper functions (avoid importing math package for WASM compatibility)
func cosApprox(angle float64) float64 {
	// Normalize angle to [0, 2π]
	for angle < 0 {
		angle += 2 * 3.14159
	}
	for angle >= 2*3.14159 {
		angle -= 2 * 3.14159
	}

	// Use Taylor series approximation: cos(x) ≈ 1 - x²/2! + x⁴/4! - x⁶/6!
	x := angle
	x2 := x * x
	return 1 - x2/2 + x2*x2/24 - x2*x2*x2/720
}

func sinApprox(angle float64) float64 {
	// sin(x) = cos(x - π/2)
	return cosApprox(angle - 3.14159/2)
}
