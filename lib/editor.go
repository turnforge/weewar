package weewar

import (
	"fmt"
	"path/filepath"
)

// =============================================================================
// World Editor Core
// =============================================================================

// WorldEditor provides tools for creating and editing game worlds (maps, units, etc.)
type WorldEditor struct {
	// Current world being edited
	currentWorld *World

	// Editor state
	filename     string
	modified     bool
	brushTerrain int // Current terrain type for painting
	brushSize    int // Brush radius (0 = single hex, 1 = 7 hexes, etc.)

	// Rendering and viewport
	drawable        Drawable         // Platform-agnostic drawing interface
	layeredRenderer *LayeredRenderer // Fast layered rendering system
	canvasWidth     int
	canvasHeight    int
	scrollX         int // Horizontal scroll offset for viewport
	scrollY         int // Vertical scroll offset for viewport

	unitLayer *UnitLayer
	tileLayer *TileLayer
}

// NewWorldEditor creates a new world editor instance
func NewWorldEditor() *WorldEditor {
	return &WorldEditor{
		currentWorld: nil,
		filename:     "",
		modified:     false,
		brushTerrain: 1, // Default to grass
		brushSize:    0, // Single hex brush
		scrollX:      0, // No initial scroll offset
		scrollY:      0, // No initial scroll offset
	}
}

// =============================================================================
// Map Management
// =============================================================================

// NewWorld creates a new 1x1 world for editing (use Add/Remove methods to expand)
func (e *WorldEditor) NewWorld() error {
	// Create new map with single tile at origin (Q=0, R=0)
	gameMap := NewMapWithBounds(0, 0, 0, 0)
	e.filename = ""
	e.modified = false

	// Create single tile at origin with default terrain (grass)
	coord := CubeCoord{Q: 0, R: 0}
	tile := NewTile(coord, 1) // Grass terrain
	gameMap.AddTile(tile)

	// Create world with map and initialize units by player
	e.currentWorld = &World{
		Map:           gameMap,
		UnitsByPlayer: make([][]*Unit, 2), // Start with 2 players
		PlayerCount:   2,
	}

	// Update layered renderer with new world
	if e.layeredRenderer != nil {
		e.layeredRenderer.SetWorld(e.currentWorld)
	}

	return nil
}

// LoadMap loads an existing map for editing
func (e *WorldEditor) LoadMap(filename string) error {
	// TODO: Implement map loading from file
	// For now, create a placeholder implementation
	return fmt.Errorf("map loading not yet implemented")
}

// SaveMap saves the current map to file
func (e *WorldEditor) SaveMap(filename string) error {
	if e.currentWorld == nil {
		return fmt.Errorf("no map to save")
	}

	// TODO: Implement map saving to file
	// For now, just update the filename and mark as unmodified
	e.filename = filename
	e.modified = false

	return nil
}

// GetCurrentMap returns the map being edited (read-only access)
func (e *WorldEditor) GetCurrentMap() *Map {
	return e.currentWorld.Map
}

// IsModified returns whether the map has unsaved changes
func (e *WorldEditor) IsModified() bool {
	return e.modified
}

// CalculateCanvasSize returns the optimal canvas size for the current map
func (e *WorldEditor) CalculateCanvasSize(rows, cols int) (width, height int) {
	if e.currentWorld == nil {
		return 400, 300 // Default size for empty editor
	}

	// Use the world renderer to calculate proper tile dimensions
	renderer := &BaseRenderer{}
	tempWorld := &World{Map: e.currentWorld.Map}

	// Use a reasonable base canvas size for calculation
	options := renderer.CalculateRenderOptions(800, 600, tempWorld)

	mapWidth, mapHeight := e.currentWorld.Map.CanvasSize(options.TileWidth, options.TileHeight, options.YIncrement)

	// Calculate canvas dimensions with minimal padding
	width = int(mapWidth + max(e.scrollX, 0))
	height = int(mapHeight + max(e.scrollY, 0))
	fmt.Println("Ok here, rows, cols, w, h: ", rows, cols, width, height, options)

	// Ensure minimum size for usability
	if width < 200 {
		width = 200
	}
	if height < 200 {
		height = 200
	}

	return width, height
}

// GetFilename returns the current filename (empty if new map)
func (e *WorldEditor) GetFilename() string {
	return e.filename
}

// GetCanvasSize returns the current canvas dimensions
func (e *WorldEditor) GetCanvasSize() (width, height int) {
	return e.canvasWidth, e.canvasHeight
}

// GetLayeredRenderer returns the layered renderer for direct access
func (e *WorldEditor) GetLayeredRenderer() *LayeredRenderer {
	return e.layeredRenderer
}

// SetAssetProvider updates the asset provider for terrain/unit sprites
func (e *WorldEditor) SetAssetProvider(provider AssetProvider) {
	if e.layeredRenderer != nil {
		e.layeredRenderer.SetAssetProvider(provider)
	}
}

// =============================================================================
// Terrain Editing
// =============================================================================

// SetBrushTerrain sets the terrain type for painting
func (e *WorldEditor) SetBrushTerrain(terrainType int) error {
	if terrainType < 0 || terrainType >= len(terrainData) {
		return fmt.Errorf("invalid terrain type: %d", terrainType)
	}
	e.brushTerrain = terrainType
	return nil
}

// SetBrushSize sets the brush radius (0 = single hex, 1 = 7 hexes, etc.)
func (e *WorldEditor) SetBrushSize(size int) error {
	if size < 0 || size > 5 {
		return fmt.Errorf("invalid brush size: %d (must be 0-5)", size)
	}
	e.brushSize = size
	return nil
}

// PaintTerrain paints terrain at the specified cube coordinate
func (e *WorldEditor) PaintTerrain(coord CubeCoord) error {
	if e.currentWorld == nil {
		return fmt.Errorf("no world loaded")
	}

	// Get all positions to paint based on brush size
	positions := e.getBrushPositions(coord)

	// Paint each position
	for _, paintCoord := range positions {
		// Check if position is within map bounds
		if !e.currentWorld.Map.IsWithinBoundsCube(paintCoord) {
			continue // Skip out-of-bounds positions
		}

		// Get existing tile or create new one
		tile := e.currentWorld.Map.TileAt(paintCoord)
		if tile == nil {
			tile = NewTile(paintCoord, e.brushTerrain)
			e.currentWorld.Map.AddTile(tile)
		} else {
			tile.TileType = e.brushTerrain
		}
	}

	e.modified = true

	// Mark affected tiles as dirty for efficient rendering
	for _, paintCoord := range positions {
		e.tileLayer.MarkDirty(paintCoord)
	}

	return nil
}

// RemoveTerrain removes terrain at the specified cube coordinate
func (e *WorldEditor) RemoveTerrain(coord CubeCoord) error {
	if e.currentWorld == nil {
		return fmt.Errorf("no world loaded")
	}

	e.currentWorld.Map.DeleteTile(coord)
	e.modified = true

	// Mark affected tile as dirty for efficient rendering
	e.tileLayer.MarkDirty(coord)
	return nil
}

// FloodFill fills a connected region with the current brush terrain
func (e *WorldEditor) FloodFill(coord CubeCoord) error {
	if e.currentWorld == nil {
		return fmt.Errorf("no world loaded")
	}

	startTile := e.currentWorld.Map.TileAt(coord)
	if startTile == nil {
		return fmt.Errorf("no tile at coordinate %v", coord)
	}

	originalTerrain := startTile.TileType
	if originalTerrain == e.brushTerrain {
		return nil // Already the target terrain
	}

	// Use breadth-first search for flood fill
	visited := make(map[CubeCoord]bool)
	queue := []CubeCoord{coord}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		// Check if this position has the original terrain
		tile := e.currentWorld.Map.TileAt(current)
		if tile == nil || tile.TileType != originalTerrain {
			continue
		}

		// Change terrain
		tile.TileType = e.brushTerrain

		// Add neighbors to queue
		var neighbors [6]CubeCoord
		current.Neighbors(&neighbors)
		for _, neighbor := range neighbors {
			if !visited[neighbor] {
				// Check if neighbor is within bounds
				if e.currentWorld.Map.IsWithinBoundsCube(neighbor) {
					queue = append(queue, neighbor)
				}
			}
		}
	}

	e.modified = true

	// Mark entire terrain as dirty since flood fill can affect large areas
	e.tileLayer.MarkAllDirty()

	return nil
}

// =============================================================================
// History Management (Undo/Redo) - TODO: Implement later
// =============================================================================

// TODO: Implement undo/redo system
// For now, history is disabled to simplify the editor

// =============================================================================
// Utility Methods
// =============================================================================

// getBrushPositions returns all cube coordinates affected by the current brush
func (e *WorldEditor) getBrushPositions(center CubeCoord) []CubeCoord {
	if e.brushSize == 0 {
		return []CubeCoord{center}
	}

	// Use the Range method from cube coordinates
	return center.Range(e.brushSize)
}

// TODO: Implement history methods when undo/redo is added back

// =============================================================================
// Map Information
// =============================================================================

// GetMapInfo returns information about the current map
func (e *WorldEditor) GetMapInfo() *MapInfo {
	if e.currentWorld == nil || e.currentWorld.Map == nil {
		return nil
	}

	// Count terrain types
	terrainCounts := make(map[int]int)
	totalTiles := 0

	for _, tile := range e.currentWorld.Map.Tiles {
		if tile != nil {
			terrainCounts[tile.TileType]++
			totalTiles++
		}
	}

	// Calculate map dimensions from bounds
	minQ, maxQ, minR, maxR := e.currentWorld.Map.GetBounds()
	width := maxQ - minQ + 1
	height := maxR - minR + 1

	return &MapInfo{
		Filename:      e.filename,
		Width:         width,
		Height:        height,
		TotalTiles:    totalTiles,
		TerrainCounts: terrainCounts,
		Modified:      e.modified,
	}
}

// MapInfo contains information about a map
type MapInfo struct {
	Filename      string
	Width         int
	Height        int
	TotalTiles    int
	TerrainCounts map[int]int
	Modified      bool
}

// =============================================================================
// Map Validation
// =============================================================================

// ValidateMap checks the map for common issues
func (e *WorldEditor) ValidateMap() []string {
	if e.currentWorld == nil || e.currentWorld.Map == nil {
		return []string{"No world loaded"}
	}

	var issues []string

	// Check for invalid terrain types
	for coord, tile := range e.currentWorld.Map.Tiles {
		if tile != nil {
			if tile.TileType < 0 || tile.TileType >= len(terrainData) {
				issues = append(issues, fmt.Sprintf("Invalid terrain type %d at %v",
					tile.TileType, coord))
			}
		}
	}

	// Check map dimensions
	minQ, maxQ, minR, maxR := e.currentWorld.Map.GetBounds()
	width := maxQ - minQ + 1
	height := maxR - minR + 1

	if width < 3 || height < 3 {
		issues = append(issues, "Map is very small (recommended minimum 3x3)")
	}

	if width > 50 || height > 50 {
		issues = append(issues, "Map is very large (may cause performance issues)")
	}

	return issues
}

// =============================================================================
// Export Functions
// =============================================================================

// RenderToFile saves the current map as a PNG image
func (e *WorldEditor) RenderToFile(filename string, width, height int) error {
	if e.currentWorld == nil {
		return fmt.Errorf("no map to render")
	}

	// Create buffer for rendering (Buffer implements Drawable interface)
	buffer := NewBuffer(width, height)

	// Create temporary layered renderer for file output
	renderer, err := NewLayeredRenderer(buffer, width, height)
	if err != nil {
		return fmt.Errorf("failed to create renderer: %w", err)
	}

	renderer.layers = []Layer{
		NewTileLayer(width, height, renderer), // Terrain tiles (bottom layer)
		NewUnitLayer(width, height, renderer), // Units (middle layer)
	}

	// Set the world and render
	renderer.SetWorld(e.currentWorld)

	// Ensure the filename has the correct extension
	if filepath.Ext(filename) != ".png" {
		filename += ".png"
	}

	return buffer.Save(filename)
}

// =============================================================================
// Canvas Management
// =============================================================================

// SetDrawable initializes the drawable for real-time rendering
func (e *WorldEditor) SetDrawable(drawable Drawable, width, height int) error {
	// Create new layered renderer for fast prototyping
	var err error
	e.layeredRenderer, err = NewLayeredRenderer(drawable, width, height)
	if err != nil {
		return fmt.Errorf("failed to create layered renderer: %v", err)
	}

	e.tileLayer = NewTileLayer(width, height, e.layeredRenderer)
	e.unitLayer = NewUnitLayer(width, height, e.layeredRenderer)
	e.layeredRenderer.layers = []Layer{
		e.tileLayer,
		e.unitLayer,
	}

	return nil
}

// SetCanvasSize resizes the canvas
func (e *WorldEditor) SetCanvasSize(width, height int) error {
	if e.layeredRenderer == nil {
		return fmt.Errorf("no layered renderer initialized")
	}

	e.canvasWidth = width
	e.canvasHeight = height

	// Resize the layered renderer (this will mark everything as dirty)
	err := e.layeredRenderer.SetViewPort(e.scrollX, e.scrollY, width, height)
	if err != nil {
		return fmt.Errorf("failed to resize layered renderer: %v", err)
	}

	return nil
}

// RenderFull renders the entire current map using the layered renderer
func (e *WorldEditor) RenderFull() error {
	if e.layeredRenderer == nil || e.currentWorld == nil {
		return nil // No renderer or map to render
	}

	// Update scroll offset in render options
	e.layeredRenderer.SetViewPort(e.scrollX, e.scrollY, e.canvasWidth, e.canvasHeight)

	return nil
}

// RenderTiles renders specific tiles using the layered renderer (for partial updates)
func (e *WorldEditor) RenderTiles(coords []CubeCoord) error {
	if e.layeredRenderer == nil || e.currentWorld == nil || len(coords) == 0 {
		return nil // No renderer, map, or tiles to render
	}

	// Mark specific tiles as dirty
	for _, coord := range coords {
		e.tileLayer.MarkDirty(coord)
	}

	// Update scroll offset in render options
	e.layeredRenderer.SetViewPort(e.scrollX, e.scrollY, e.canvasWidth, e.canvasHeight)

	return nil
}
