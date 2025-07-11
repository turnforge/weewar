package weewar

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"
)

// Unit represents a runtime unit instance in the game
type Unit struct {
	UnitType int // Reference to UnitData by ID
	
	// Runtime state
	DistanceLeft    int // Movement points remaining this turn
	AvailableHealth int // Current health points
	TurnCounter     int // Which turn this unit was created/last acted
	
	// Position on the map
	Row int
	Col int
	
	// Player ownership
	PlayerID int
}

// Tile represents a single hex tile on the map
type Tile struct {
	Row int
	Col int
	
	// Hex neighbors - clockwise from LEFT
	// [0] = LEFT, [1] = TOP_LEFT, [2] = TOP_RIGHT, [3] = RIGHT, [4] = BOTTOM_RIGHT, [5] = BOTTOM_LEFT
	Neighbours [6]*Tile
	
	TileType int // Reference to TerrainData by ID
	
	// Optional: Unit occupying this tile
	Unit *Unit
}

// Map represents the game map with hex grid topology
type Map struct {
	NumRows int
	NumCols int
	
	// Hex offset configuration
	EvenRowsOffset bool // If true, even rows start at x = hex_width/2, odd rows at x = 0
	
	// Tile storage - sparse representation
	Tiles map[int]map[int]*Tile // Tiles[row][col] = *Tile
}

// Game represents the complete game state
type Game struct {
	Map *Map
	
	// Units organized by player
	Units [][]*Unit // Units[player_id][unit_index] = *Unit
	
	// Game state
	CurrentPlayer int // Index of current player
	TurnCounter   int // Current turn number
	
	// Random number generator for deterministic gameplay
	rng *rand.Rand
}


// NewMap creates a new empty map with the specified dimensions
func NewMap(numRows, numCols int, evenRowsOffset bool) *Map {
	return &Map{
		NumRows:        numRows,
		NumCols:        numCols,
		EvenRowsOffset: evenRowsOffset,
		Tiles:          make(map[int]map[int]*Tile),
	}
}

// TileAt returns the tile at the specified position, or nil if none exists
func (m *Map) TileAt(row, col int) *Tile {
	if rowMap, exists := m.Tiles[row]; exists {
		return rowMap[col]
	}
	return nil
}

// AddTile adds a tile to the map at its specified row/column position
// If a tile already exists at that position, it will be replaced
func (m *Map) AddTile(t *Tile) {
	if m.Tiles[t.Row] == nil {
		m.Tiles[t.Row] = make(map[int]*Tile)
	}
	m.Tiles[t.Row][t.Col] = t
}

// DeleteTile removes the tile at the specified position
func (m *Map) DeleteTile(row, col int) {
	if rowMap, exists := m.Tiles[row]; exists {
		delete(rowMap, col)
		// Clean up empty row maps
		if len(rowMap) == 0 {
			delete(m.Tiles, row)
		}
	}
}

// NewTile creates a new tile at the specified position
func NewTile(row, col, tileType int) *Tile {
	return &Tile{
		Row:      row,
		Col:      col,
		TileType: tileType,
	}
}

// NewUnit creates a new unit instance
func NewUnit(unitType, playerID int) *Unit {
	return &Unit{
		UnitType:        unitType,
		PlayerID:        playerID,
		DistanceLeft:    0, // Will be set based on UnitData
		AvailableHealth: 0, // Will be set based on UnitData
		TurnCounter:     0,
	}
}

// NewGame creates a new game instance with the specified number of players
func NewGame(numPlayers int, mapInstance *Map, seed int64) *Game {
	// Initialize units slice with one slice per player
	units := make([][]*Unit, numPlayers)
	for i := range units {
		units[i] = make([]*Unit, 0)
	}
	
	return &Game{
		Map:           mapInstance,
		Units:         units,
		CurrentPlayer: 0,
		TurnCounter:   1,
		rng:           rand.New(rand.NewSource(seed)),
	}
}


// AddUnit adds a unit to the game for the specified player
func (g *Game) AddUnit(unit *Unit, playerID int) {
	if playerID >= 0 && playerID < len(g.Units) {
		unit.PlayerID = playerID
		g.Units[playerID] = append(g.Units[playerID], unit)
		
		// Place unit on the map if it has a valid position
		if tile := g.Map.TileAt(unit.Row, unit.Col); tile != nil {
			tile.Unit = unit
		}
	}
}

// GetUnitsForPlayer returns all units belonging to the specified player
func (g *Game) GetUnitsForPlayer(playerID int) []*Unit {
	if playerID >= 0 && playerID < len(g.Units) {
		return g.Units[playerID]
	}
	return nil
}

// NextTurn advances the game to the next player's turn
func (g *Game) NextTurn() {
	g.CurrentPlayer = (g.CurrentPlayer + 1) % len(g.Units)
	if g.CurrentPlayer == 0 {
		g.TurnCounter++
	}
}

// GetRNG returns the game's random number generator
func (g *Game) GetRNG() *rand.Rand {
	return g.rng
}

// RenderToBuffer renders the complete game state to a buffer
func (g *Game) RenderToBuffer(buffer *Buffer, tileWidth, tileHeight, yIncrement float64) {
	// Clear buffer first
	buffer.Clear()
	
	// Render terrain layer
	g.RenderTerrain(buffer, tileWidth, tileHeight, yIncrement)
	
	// Render units layer
	g.RenderUnits(buffer, tileWidth, tileHeight, yIncrement)
	
	// Render UI layer
	g.RenderUI(buffer, tileWidth, tileHeight, yIncrement)
}

// RenderTerrain renders the terrain tiles to a buffer
func (g *Game) RenderTerrain(buffer *Buffer, tileWidth, tileHeight, yIncrement float64) {
	if g.Map == nil {
		return
	}
	
	// Define colors for different tile types
	tileColors := map[int]color.RGBA{
		0: {200, 200, 200, 255}, // Unknown/default - light gray
		1: {50, 150, 50, 255},   // Grass - green
		2: {200, 180, 100, 255}, // Desert - sandy
		3: {100, 100, 200, 255}, // Water - blue
		4: {150, 100, 50, 255},  // Mountain - brown
		5: {180, 180, 180, 255}, // Rock - gray
	}
	
	// Render each tile
	for row := range g.Map.Tiles {
		for col := range g.Map.Tiles[row] {
			tile := g.Map.Tiles[row][col]
			if tile != nil {
				// Calculate tile position
				x, y := g.Map.XYForTile(row, col, tileWidth, tileHeight, yIncrement)
				
				// Get tile color
				tileColor, exists := tileColors[tile.TileType]
				if !exists {
					tileColor = tileColors[0] // Default color
				}
				
				// Create a simple colored rectangle image for the tile
				tileImg := image.NewRGBA(image.Rect(0, 0, 1, 1))
				tileImg.Set(0, 0, tileColor)
				
				// Draw the tile as a scaled colored rectangle
				// Note: This is a simplified version - you could create proper hex-shaped tile images
				buffer.DrawImage(x, y, tileWidth, tileHeight, tileImg)
			}
		}
	}
}

// RenderUnits renders the units to a buffer
func (g *Game) RenderUnits(buffer *Buffer, tileWidth, tileHeight, yIncrement float64) {
	if g.Map == nil {
		return
	}
	
	// Define colors for different players
	playerColors := map[int]color.RGBA{
		0: {255, 0, 0, 255},   // Player 0 - red
		1: {0, 0, 255, 255},   // Player 1 - blue
		2: {0, 255, 0, 255},   // Player 2 - green
		3: {255, 255, 0, 255}, // Player 3 - yellow
		4: {255, 0, 255, 255}, // Player 4 - magenta
		5: {0, 255, 255, 255}, // Player 5 - cyan
	}
	
	// Render units for each player
	for playerID, units := range g.Units {
		for _, unit := range units {
			if unit != nil {
				// Calculate unit position (same as tile position)
				x, y := g.Map.XYForTile(unit.Row, unit.Col, tileWidth, tileHeight, yIncrement)
				
				// Get player color
				unitColor, exists := playerColors[playerID]
				if !exists {
					unitColor = color.RGBA{128, 128, 128, 255} // Default gray
				}
				
				// Create a simple colored circle/square for the unit
				unitImg := image.NewRGBA(image.Rect(0, 0, 1, 1))
				unitImg.Set(0, 0, unitColor)
				
				// Draw the unit smaller than the tile, centered
				unitSize := math.Min(tileWidth, tileHeight) * 0.6
				unitOffsetX := (tileWidth - unitSize) / 2
				unitOffsetY := (tileHeight - unitSize) / 2
				
				buffer.DrawImage(x+unitOffsetX, y+unitOffsetY, unitSize, unitSize, unitImg)
			}
		}
	}
}

// RenderUI renders UI elements to a buffer
func (g *Game) RenderUI(buffer *Buffer, tileWidth, tileHeight, yIncrement float64) {
	// For now, just render basic game info
	// In a full implementation, you'd render text, current player indicator, etc.
	
	// Create a simple indicator for current player
	currentPlayerColor := color.RGBA{255, 255, 255, 200} // Semi-transparent white
	if g.CurrentPlayer < len(g.Units) {
		// You could render player info, turn counter, etc.
		// For now, just create a small indicator
		indicatorImg := image.NewRGBA(image.Rect(0, 0, 1, 1))
		indicatorImg.Set(0, 0, currentPlayerColor)
		
		// Draw a small indicator in the top-left corner
		buffer.DrawImage(5, 5, 20, 20, indicatorImg)
	}
}

// GetHexNeighborCoords returns the coordinates of the 6 hex neighbors
// Returns array of [row, col] pairs in clockwise order from LEFT
// [0] = LEFT, [1] = TOP_LEFT, [2] = TOP_RIGHT, [3] = RIGHT, [4] = BOTTOM_RIGHT, [5] = BOTTOM_LEFT
func (m *Map) GetHexNeighborCoords(row, col int) [6][2]int {
	var neighbors [6][2]int
	
	// Hex grid neighbor calculation depends on whether we're in even or odd row
	isEvenRow := (row % 2) == 0
	
	if m.EvenRowsOffset {
		// Even rows are offset to the right
		if isEvenRow {
			// Even row neighbors
			neighbors[0] = [2]int{row, col - 1}     // LEFT
			neighbors[1] = [2]int{row - 1, col}     // TOP_LEFT
			neighbors[2] = [2]int{row - 1, col + 1} // TOP_RIGHT
			neighbors[3] = [2]int{row, col + 1}     // RIGHT
			neighbors[4] = [2]int{row + 1, col + 1} // BOTTOM_RIGHT
			neighbors[5] = [2]int{row + 1, col}     // BOTTOM_LEFT
		} else {
			// Odd row neighbors
			neighbors[0] = [2]int{row, col - 1}     // LEFT
			neighbors[1] = [2]int{row - 1, col - 1} // TOP_LEFT
			neighbors[2] = [2]int{row - 1, col}     // TOP_RIGHT
			neighbors[3] = [2]int{row, col + 1}     // RIGHT
			neighbors[4] = [2]int{row + 1, col}     // BOTTOM_RIGHT
			neighbors[5] = [2]int{row + 1, col - 1} // BOTTOM_LEFT
		}
	} else {
		// Odd rows are offset to the right
		if isEvenRow {
			// Even row neighbors
			neighbors[0] = [2]int{row, col - 1}     // LEFT
			neighbors[1] = [2]int{row - 1, col - 1} // TOP_LEFT
			neighbors[2] = [2]int{row - 1, col}     // TOP_RIGHT
			neighbors[3] = [2]int{row, col + 1}     // RIGHT
			neighbors[4] = [2]int{row + 1, col}     // BOTTOM_RIGHT
			neighbors[5] = [2]int{row + 1, col - 1} // BOTTOM_LEFT
		} else {
			// Odd row neighbors
			neighbors[0] = [2]int{row, col - 1}     // LEFT
			neighbors[1] = [2]int{row - 1, col}     // TOP_LEFT
			neighbors[2] = [2]int{row - 1, col + 1} // TOP_RIGHT
			neighbors[3] = [2]int{row, col + 1}     // RIGHT
			neighbors[4] = [2]int{row + 1, col + 1} // BOTTOM_RIGHT
			neighbors[5] = [2]int{row + 1, col}     // BOTTOM_LEFT
		}
	}
	
	return neighbors
}

// ConnectHexNeighbors automatically connects all tiles in the map as hex neighbors
// This should be called after all tiles have been added to the map
func (m *Map) ConnectHexNeighbors() {
	for row := range m.Tiles {
		for col := range m.Tiles[row] {
			tile := m.Tiles[row][col]
			if tile != nil {
				neighborCoords := m.GetHexNeighborCoords(row, col)
				
				for i, coord := range neighborCoords {
					neighborTile := m.TileAt(coord[0], coord[1])
					tile.Neighbours[i] = neighborTile
				}
			}
		}
	}
}

// XYForTile converts tile row/col coordinates to pixel x,y coordinates for rendering
// Given the hex tile's width and height, return the x and y coordinate of the given row and col.
// yIncrement is the vertical spacing between rows (typically less than tileHeight for overlapping hex rows)
//
// Notes:
// 1. Takes into account the map's EvenRowOffset setting
// 2. Tiles are laid out horizontally so they are "pointing up" (flat-top hexagons)
// 3. Supports custom scaling in x and y axes via tileWidth and tileHeight
//
// The resulting hexagon clip path would be:
// [(x + tileWidth/2, y), (x+tileWidth, y + tileHeight - yIncrement), (x + tileWidth, y + yIncrement),
//  (x + tileWidth/2, y + tileHeight), (x, y + yIncrement), (x, y + tileHeight - yIncrement)]
func (m *Map) XYForTile(row, col int, tileWidth, tileHeight, yIncrement float64) (x, y float64) {
	// Calculate base x position
	x = float64(col) * tileWidth
	
	// Apply offset for alternating rows (hex grid staggering)
	isEvenRow := (row % 2) == 0
	if m.EvenRowsOffset {
		// Even rows are offset to the right
		if isEvenRow {
			x += tileWidth / 2
		}
	} else {
		// Odd rows are offset to the right
		if !isEvenRow {
			x += tileWidth / 2
		}
	}
	
	// Calculate y position
	y = float64(row) * yIncrement
	
	return x, y
}

// GetHexPolygonPoints returns the 6 points of a hexagon at the given x,y position
// The hexagon is flat-top (pointing up) with the specified dimensions
// Returns points in clockwise order starting from the top point
func GetHexPolygonPoints(x, y, tileWidth, tileHeight, yIncrement float64) []image.Point {
	points := make([]image.Point, 6)
	
	// Top point
	points[0] = image.Point{X: int(x + tileWidth/2), Y: int(y)}
	
	// Top-right point
	points[1] = image.Point{X: int(x + tileWidth), Y: int(y + tileHeight - yIncrement)}
	
	// Bottom-right point
	points[2] = image.Point{X: int(x + tileWidth), Y: int(y + yIncrement)}
	
	// Bottom point
	points[3] = image.Point{X: int(x + tileWidth/2), Y: int(y + tileHeight)}
	
	// Bottom-left point
	points[4] = image.Point{X: int(x), Y: int(y + yIncrement)}
	
	// Top-left point
	points[5] = image.Point{X: int(x), Y: int(y + tileHeight - yIncrement)}
	
	return points
}

// fillPolygon fills a polygon defined by points with the given color
func fillPolygon(img draw.Image, points []image.Point, fillColor color.Color) {
	// Simple polygon fill using scanline algorithm
	if len(points) < 3 {
		return
	}
	
	bounds := img.Bounds()
	minY := bounds.Max.Y
	maxY := bounds.Min.Y
	
	// Find Y bounds
	for _, p := range points {
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	
	// Ensure bounds are within image
	if minY < bounds.Min.Y {
		minY = bounds.Min.Y
	}
	if maxY > bounds.Max.Y {
		maxY = bounds.Max.Y
	}
	
	// For each scanline
	for y := minY; y <= maxY; y++ {
		intersections := []int{}
		
		// Find intersections with polygon edges
		for i := 0; i < len(points); i++ {
			j := (i + 1) % len(points)
			p1, p2 := points[i], points[j]
			
			if (p1.Y <= y && y < p2.Y) || (p2.Y <= y && y < p1.Y) {
				// Calculate intersection x
				if p2.Y != p1.Y {
					x := p1.X + (y-p1.Y)*(p2.X-p1.X)/(p2.Y-p1.Y)
					intersections = append(intersections, x)
				}
			}
		}
		
		// Sort intersections
		for i := 0; i < len(intersections)-1; i++ {
			for j := i + 1; j < len(intersections); j++ {
				if intersections[i] > intersections[j] {
					intersections[i], intersections[j] = intersections[j], intersections[i]
				}
			}
		}
		
		// Fill between pairs of intersections
		for i := 0; i < len(intersections)-1; i += 2 {
			x1 := intersections[i]
			x2 := intersections[i+1]
			
			// Ensure bounds are within image
			if x1 < bounds.Min.X {
				x1 = bounds.Min.X
			}
			if x2 > bounds.Max.X {
				x2 = bounds.Max.X
			}
			
			for x := x1; x <= x2; x++ {
				img.Set(x, y, fillColor)
			}
		}
	}
}

// getMapBounds calculates the pixel bounds of the entire map
func (m *Map) getMapBounds(tileWidth, tileHeight, yIncrement float64) (minX, minY, maxX, maxY float64) {
	minX = math.Inf(1)
	minY = math.Inf(1)
	maxX = math.Inf(-1)
	maxY = math.Inf(-1)
	
	for row := range m.Tiles {
		for col := range m.Tiles[row] {
			x, y := m.XYForTile(row, col, tileWidth, tileHeight, yIncrement)
			
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x+tileWidth > maxX {
				maxX = x + tileWidth
			}
			if y+tileHeight > maxY {
				maxY = y + tileHeight
			}
		}
	}
	
	return minX, minY, maxX, maxY
}


// drawLine draws a line between two points
func drawLine(img draw.Image, p1, p2 image.Point, lineColor color.Color) {
	dx := abs(p2.X - p1.X)
	dy := abs(p2.Y - p1.Y)
	
	var sx, sy int
	if p1.X < p2.X {
		sx = 1
	} else {
		sx = -1
	}
	if p1.Y < p2.Y {
		sy = 1
	} else {
		sy = -1
	}
	
	err := dx - dy
	x, y := p1.X, p1.Y
	
	for {
		if x >= img.Bounds().Min.X && x < img.Bounds().Max.X &&
			y >= img.Bounds().Min.Y && y < img.Bounds().Max.Y {
			img.Set(x, y, lineColor)
		}
		
		if x == p2.X && y == p2.Y {
			break
		}
		
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

