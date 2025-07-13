package weewar

import (
	"fmt"
	"path/filepath"
)

// =============================================================================
// Map Editor Core
// =============================================================================

// MapEditor provides tools for creating and editing hex maps
type MapEditor struct {
	// Current map being edited
	currentMap *Map
	
	// Editor state
	filename     string
	modified     bool
	brushTerrain int  // Current terrain type for painting
	brushSize    int  // Brush radius (0 = single hex, 1 = 7 hexes, etc.)
	
	// Canvas rendering
	canvasBuffer *CanvasBuffer // Direct HTML canvas rendering
	canvasWidth  int
	canvasHeight int
	
	// Undo/redo system
	history    []*Map  // Map snapshots for undo
	historyPos int     // Current position in history
	maxHistory int     // Maximum undo steps
}

// NewMapEditor creates a new map editor instance
func NewMapEditor() *MapEditor {
	return &MapEditor{
		currentMap:   nil,
		filename:     "",
		modified:     false,
		brushTerrain: 1, // Default to grass
		brushSize:    0, // Single hex brush
		history:      make([]*Map, 0),
		historyPos:   -1,
		maxHistory:   50, // Keep last 50 operations
	}
}

// =============================================================================
// Map Management
// =============================================================================

// NewMap creates a new empty map for editing
func (e *MapEditor) NewMap(rows, cols int) error {
	if rows < 1 || rows > 100 {
		return fmt.Errorf("invalid rows: %d (must be 1-100)", rows)
	}
	if cols < 1 || cols > 100 {
		return fmt.Errorf("invalid cols: %d (must be 1-100)", cols)
	}
	
	// Create new map with cube coordinate storage
	e.currentMap = NewMap(rows, cols, false)
	e.filename = ""
	e.modified = false
	
	// Fill with default terrain (grass)
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			tile := NewTile(row, col, 1) // Grass terrain
			e.currentMap.AddTile(tile)
		}
	}
	
	// Clear history and take snapshot
	e.clearHistory()
	e.takeSnapshot()
	
	// Auto-render the new map
	e.renderFullMap()
	
	return nil
}

// LoadMap loads an existing map for editing
func (e *MapEditor) LoadMap(filename string) error {
	// TODO: Implement map loading from file
	// For now, create a placeholder implementation
	return fmt.Errorf("map loading not yet implemented")
}

// SaveMap saves the current map to file
func (e *MapEditor) SaveMap(filename string) error {
	if e.currentMap == nil {
		return fmt.Errorf("no map to save")
	}
	
	// TODO: Implement map saving to file
	// For now, just update the filename and mark as unmodified
	e.filename = filename
	e.modified = false
	
	return nil
}

// GetCurrentMap returns the map being edited (read-only access)
func (e *MapEditor) GetCurrentMap() *Map {
	return e.currentMap
}

// IsModified returns whether the map has unsaved changes
func (e *MapEditor) IsModified() bool {
	return e.modified
}

// GetFilename returns the current filename (empty if new map)
func (e *MapEditor) GetFilename() string {
	return e.filename
}

// =============================================================================
// Terrain Editing
// =============================================================================

// SetBrushTerrain sets the terrain type for painting
func (e *MapEditor) SetBrushTerrain(terrainType int) error {
	if terrainType < 0 || terrainType >= len(terrainData) {
		return fmt.Errorf("invalid terrain type: %d", terrainType)
	}
	e.brushTerrain = terrainType
	return nil
}

// SetBrushSize sets the brush radius (0 = single hex, 1 = 7 hexes, etc.)
func (e *MapEditor) SetBrushSize(size int) error {
	if size < 0 || size > 5 {
		return fmt.Errorf("invalid brush size: %d (must be 0-5)", size)
	}
	e.brushSize = size
	return nil
}

// PaintTerrain paints terrain at the specified display position
func (e *MapEditor) PaintTerrain(row, col int) error {
	if e.currentMap == nil {
		return fmt.Errorf("no map loaded")
	}
	
	// Convert to cube coordinate
	centerCoord := e.currentMap.DisplayToHex(row, col)
	
	// Get all positions to paint based on brush size
	positions := e.getBrushPositions(centerCoord)
	
	// Paint each position
	for _, coord := range positions {
		// Check if position is within map bounds
		displayRow, displayCol := e.currentMap.HexToDisplay(coord)
		if displayRow < 0 || displayRow >= e.currentMap.NumRows ||
		   displayCol < 0 || displayCol >= e.currentMap.NumCols {
			continue // Skip out-of-bounds positions
		}
		
		// Get existing tile or create new one
		tile := e.currentMap.TileAtCube(coord)
		if tile == nil {
			tile = NewTile(displayRow, displayCol, e.brushTerrain)
			e.currentMap.AddTileCube(coord, tile)
		} else {
			tile.TileType = e.brushTerrain
		}
	}
	
	// Take snapshot after making changes
	e.takeSnapshot()
	e.modified = true
	
	// Auto-render affected tiles
	e.renderTiles(positions)
	
	return nil
}

// RemoveTerrain removes terrain at the specified position
func (e *MapEditor) RemoveTerrain(row, col int) error {
	if e.currentMap == nil {
		return fmt.Errorf("no map loaded")
	}
	
	coord := e.currentMap.DisplayToHex(row, col)
	e.currentMap.DeleteTileCube(coord)
	
	e.takeSnapshot()
	e.modified = true
	
	// Auto-render affected tile
	e.renderTiles([]CubeCoord{coord})
	
	return nil
}

// FloodFill fills a connected region with the current brush terrain
func (e *MapEditor) FloodFill(row, col int) error {
	if e.currentMap == nil {
		return fmt.Errorf("no map loaded")
	}
	
	startCoord := e.currentMap.DisplayToHex(row, col)
	startTile := e.currentMap.TileAtCube(startCoord)
	if startTile == nil {
		return fmt.Errorf("no tile at position (%d, %d)", row, col)
	}
	
	originalTerrain := startTile.TileType
	if originalTerrain == e.brushTerrain {
		return nil // Already the target terrain
	}
	
	// Use breadth-first search for flood fill
	visited := make(map[CubeCoord]bool)
	queue := []CubeCoord{startCoord}
	
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		
		if visited[current] {
			continue
		}
		visited[current] = true
		
		// Check if this position has the original terrain
		tile := e.currentMap.TileAtCube(current)
		if tile == nil || tile.TileType != originalTerrain {
			continue
		}
		
		// Change terrain
		tile.TileType = e.brushTerrain
		
		// Add neighbors to queue
		neighbors := current.Neighbors()
		for _, neighbor := range neighbors {
			if !visited[neighbor] {
				// Check if neighbor is within bounds
				nRow, nCol := e.currentMap.HexToDisplay(neighbor)
				if nRow >= 0 && nRow < e.currentMap.NumRows &&
				   nCol >= 0 && nCol < e.currentMap.NumCols {
					queue = append(queue, neighbor)
				}
			}
		}
	}
	
	e.takeSnapshot()
	e.modified = true
	
	// Auto-render entire map (flood fill can affect large areas)
	e.renderFullMap()
	
	return nil
}

// =============================================================================
// History Management (Undo/Redo)
// =============================================================================

// Undo reverts the last operation
func (e *MapEditor) Undo() error {
	if e.historyPos <= 0 {
		return fmt.Errorf("nothing to undo")
	}
	
	e.historyPos--
	e.currentMap = e.copyMap(e.history[e.historyPos])
	e.modified = true
	
	// Auto-render entire map (undo can affect entire state)
	e.renderFullMap()
	
	return nil
}

// Redo reapplies the next operation
func (e *MapEditor) Redo() error {
	if e.historyPos >= len(e.history)-1 {
		return fmt.Errorf("nothing to redo")
	}
	
	e.historyPos++
	e.currentMap = e.copyMap(e.history[e.historyPos])
	e.modified = true
	
	// Auto-render entire map (redo can affect entire state)
	e.renderFullMap()
	
	return nil
}

// CanUndo returns whether undo is available
func (e *MapEditor) CanUndo() bool {
	return e.historyPos > 0
}

// CanRedo returns whether redo is available
func (e *MapEditor) CanRedo() bool {
	return e.historyPos < len(e.history)-1
}

// =============================================================================
// Utility Methods
// =============================================================================

// getBrushPositions returns all cube coordinates affected by the current brush
func (e *MapEditor) getBrushPositions(center CubeCoord) []CubeCoord {
	if e.brushSize == 0 {
		return []CubeCoord{center}
	}
	
	// Use the Range method from cube coordinates
	return center.Range(e.brushSize)
}

// takeSnapshot saves the current map state for undo
func (e *MapEditor) takeSnapshot() {
	if e.currentMap == nil {
		return
	}
	
	// Remove any redo history when taking a new snapshot
	if e.historyPos < len(e.history)-1 {
		e.history = e.history[:e.historyPos+1]
	}
	
	// Add current map to history
	mapCopy := e.copyMap(e.currentMap)
	e.history = append(e.history, mapCopy)
	e.historyPos = len(e.history) - 1
	
	// Limit history size
	if len(e.history) > e.maxHistory {
		e.history = e.history[1:]
		e.historyPos--
	}
}

// copyMap creates a deep copy of a map
func (e *MapEditor) copyMap(original *Map) *Map {
	if original == nil {
		return nil
	}
	
	copy := NewMap(original.NumRows, original.NumCols, false)
	
	// Copy all tiles
	for coord, tile := range original.Tiles {
		if tile != nil {
			newTile := &Tile{
				Row:      tile.Row,
				Col:      tile.Col,
				TileType: tile.TileType,
				Unit:     nil, // Don't copy units in editor
			}
			copy.AddTileCube(coord, newTile)
		}
	}
	
	return copy
}

// clearHistory clears the undo/redo history
func (e *MapEditor) clearHistory() {
	e.history = make([]*Map, 0)
	e.historyPos = -1
}

// =============================================================================
// Map Information
// =============================================================================

// GetMapInfo returns information about the current map
func (e *MapEditor) GetMapInfo() *MapInfo {
	if e.currentMap == nil {
		return nil
	}
	
	// Count terrain types
	terrainCounts := make(map[int]int)
	totalTiles := 0
	
	for _, tile := range e.currentMap.Tiles {
		if tile != nil {
			terrainCounts[tile.TileType]++
			totalTiles++
		}
	}
	
	return &MapInfo{
		Filename:      e.filename,
		Width:         e.currentMap.NumCols,
		Height:        e.currentMap.NumRows,
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
func (e *MapEditor) ValidateMap() []string {
	if e.currentMap == nil {
		return []string{"No map loaded"}
	}
	
	var issues []string
	
	// Check for missing tiles (holes in the map)
	expectedTiles := e.currentMap.NumRows * e.currentMap.NumCols
	actualTiles := len(e.currentMap.Tiles)
	
	if actualTiles < expectedTiles {
		issues = append(issues, fmt.Sprintf("Map has holes: %d tiles missing", expectedTiles-actualTiles))
	}
	
	// Check for invalid terrain types
	for _, tile := range e.currentMap.Tiles {
		if tile != nil {
			if tile.TileType < 0 || tile.TileType >= len(terrainData) {
				issues = append(issues, fmt.Sprintf("Invalid terrain type %d at (%d, %d)", 
					tile.TileType, tile.Row, tile.Col))
			}
		}
	}
	
	// Check map dimensions
	if e.currentMap.NumRows < 3 || e.currentMap.NumCols < 3 {
		issues = append(issues, "Map is very small (recommended minimum 3x3)")
	}
	
	if e.currentMap.NumRows > 50 || e.currentMap.NumCols > 50 {
		issues = append(issues, "Map is very large (may cause performance issues)")
	}
	
	return issues
}

// =============================================================================
// Export Functions
// =============================================================================

// ExportToGame converts the edited map to a Game instance for testing
func (e *MapEditor) ExportToGame(playerCount int) (*Game, error) {
	if e.currentMap == nil {
		return nil, fmt.Errorf("no map to export")
	}
	
	// Validate player count
	if playerCount < 2 || playerCount > 6 {
		return nil, fmt.Errorf("invalid player count: %d (must be 2-6)", playerCount)
	}
	
	// Create a copy of the map for the game
	gamemap := e.copyMap(e.currentMap)
	
	// Create the game
	game, err := NewGame(playerCount, gamemap, 12345) // Use fixed seed for testing
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}
	
	return game, nil
}

// RenderToFile saves the current map as a PNG image
func (e *MapEditor) RenderToFile(filename string, width, height int) error {
	if e.currentMap == nil {
		return fmt.Errorf("no map to render")
	}
	
	// Create a temporary game for rendering
	game, err := e.ExportToGame(2)
	if err != nil {
		return fmt.Errorf("failed to create game for rendering: %w", err)
	}
	
	// Create buffer and render
	buffer := NewBuffer(width, height)
	
	// Calculate tile size based on map dimensions and buffer size
	tileWidth := float64(width) / float64(e.currentMap.NumCols)
	tileHeight := float64(height) / float64(e.currentMap.NumRows)
	yIncrement := tileHeight * 0.75 // Hex grid spacing
	
	err = game.RenderToBuffer(buffer, tileWidth, tileHeight, yIncrement)
	if err != nil {
		return fmt.Errorf("failed to render map: %w", err)
	}
	
	// Ensure the filename has the correct extension
	if filepath.Ext(filename) != ".png" {
		filename += ".png"
	}
	
	return buffer.Save(filename)
}

// =============================================================================
// Canvas Management
// =============================================================================

// SetCanvas initializes the canvas for real-time rendering
func (e *MapEditor) SetCanvas(canvasID string, width, height int) error {
	// Create new canvas buffer
	e.canvasBuffer = NewCanvasBuffer(canvasID, width, height)
	if e.canvasBuffer == nil {
		return fmt.Errorf("failed to create canvas buffer for '%s'", canvasID)
	}
	
	e.canvasWidth = width
	e.canvasHeight = height
	
	// Render current map if available
	if e.currentMap != nil {
		e.renderFullMap()
	}
	
	return nil
}

// SetCanvasSize resizes the canvas
func (e *MapEditor) SetCanvasSize(width, height int) error {
	if e.canvasBuffer == nil {
		return fmt.Errorf("no canvas initialized")
	}
	
	e.canvasWidth = width
	e.canvasHeight = height
	
	// Recreate canvas buffer with new size
	canvasID := e.canvasBuffer.canvasID
	e.canvasBuffer = NewCanvasBuffer(canvasID, width, height)
	if e.canvasBuffer == nil {
		return fmt.Errorf("failed to resize canvas buffer")
	}
	
	// Re-render current map
	if e.currentMap != nil {
		e.renderFullMap()
	}
	
	return nil
}

// renderFullMap renders the entire current map to the canvas
func (e *MapEditor) renderFullMap() error {
	if e.canvasBuffer == nil || e.currentMap == nil {
		return nil // No canvas or map to render
	}
	
	// Simplified rendering directly using FillPath for each tile
	tileWidth := float64(e.canvasWidth) / float64(e.currentMap.NumCols)
	tileHeight := float64(e.canvasHeight) / float64(e.currentMap.NumRows)
	
	// Render each tile as a hexagon
	for coord, tile := range e.currentMap.Tiles {
		if tile == nil {
			continue
		}
		
		// Convert hex coordinates to display coordinates
		displayRow, displayCol := e.currentMap.HexToDisplay(coord)
		
		// Calculate tile position
		x := float64(displayCol) * tileWidth
		y := float64(displayRow) * (tileHeight * 0.75) // Hex grid spacing
		
		// Offset even rows for hex grid
		if displayRow%2 == 0 {
			x += tileWidth * 0.5
		}
		
		// Create hexagon points
		hexPoints := createHexPoints(x+tileWidth/2, y+tileHeight/2, tileWidth*0.4)
		
		// Get terrain color
		var fillColor Color
		switch tile.TileType {
		case 1: // Grass
			fillColor = Color{R: 34, G: 139, B: 34, A: 255}
		case 2: // Desert
			fillColor = Color{R: 238, G: 203, B: 173, A: 255}
		case 3: // Water
			fillColor = Color{R: 65, G: 105, B: 225, A: 255}
		case 4: // Mountain
			fillColor = Color{R: 139, G: 137, B: 137, A: 255}
		case 5: // Rock
			fillColor = Color{R: 105, G: 105, B: 105, A: 255}
		default:
			fillColor = Color{R: 200, G: 200, B: 200, A: 255}
		}
		
		// Fill the hexagon
		e.canvasBuffer.FillPath(hexPoints, fillColor)
		
		// Draw border
		borderColor := Color{R: 0, G: 0, B: 0, A: 100}
		strokeProps := StrokeProperties{Width: 1.0, LineCap: "round", LineJoin: "round"}
		e.canvasBuffer.StrokePath(hexPoints, borderColor, strokeProps)
	}
	
	return nil
}

// renderTiles renders specific tiles to the canvas (for partial updates)
func (e *MapEditor) renderTiles(coords []CubeCoord) error {
	if e.canvasBuffer == nil || e.currentMap == nil || len(coords) == 0 {
		return nil // No canvas, map, or tiles to render
	}
	
	// For now, do a full render - we'll optimize for partial rendering later
	return e.renderFullMap()
}