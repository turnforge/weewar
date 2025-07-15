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
		{R: 64, G: 64, B: 64, A: 255},    // 0 - Unknown (dark gray)
		{R: 34, G: 139, B: 34, A: 255},   // 1 - Grass (forest green)
		{R: 238, G: 203, B: 173, A: 255}, // 2 - Desert (sandy brown)
		{R: 65, G: 105, B: 225, A: 255},  // 3 - Water (royal blue)
		{R: 139, G: 69, B: 19, A: 255},   // 4 - Mountain (saddle brown)
		{R: 105, G: 105, B: 105, A: 255}, // 5 - Rock (dim gray)
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
		x := centerX + radius*cosApprox(angle)
		y := centerY + radius*sinApprox(angle)
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
