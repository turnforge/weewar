package weewar

import (
	"encoding/json"
	"fmt"
	"math"
)

// Basic terrain data
var terrainData = []TerrainData{
	{0, "Unknown", 1, 0},
	{1, "Grass", 1, 0},
	{2, "Desert", 1, 0},
	{3, "Water", 2, 0},
	{4, "Mountain", 2, 10},
	{5, "Rock", 3, 20},
}

// GetTerrainData returns terrain data for the given type
func GetTerrainData(terrainType int) *TerrainData {
	for i := range terrainData {
		if terrainData[i].ID == terrainType {
			return &terrainData[i]
		}
	}
	return &terrainData[0] // Default to unknown
}

// NeighborDirection represents the 6 directions in a hex grid
type NeighborDirection int

const (
	LEFT NeighborDirection = iota
	TOP_LEFT
	TOP_RIGHT
	RIGHT
	BOTTOM_RIGHT
	BOTTOM_LEFT
)

// Map represents the game map with hex grid topology
type Map struct {
	// Coordinate bounds - These can be evaluated.
	minQ int `json:"minQ"` // Minimum Q coordinate (inclusive)
	maxQ int `json:"maxQ"` // Maximum Q coordinate (inclusive)
	minR int `json:"minR"` // Minimum R coordinate (inclusive)
	maxR int `json:"maxR"` // Maximum R coordinate (inclusive)

	// Where X/Y of the Origin tile (Q = R = 0) are, in normalized tile units.
	// OriginX = 5 means the origin tile is 5 tile widths to the right
	// OriginY = 3 means the origin tile is 3 tile heights down
	// Initially it would be 0,0 (top left of the screen)
	// But as we add/remove rows and columns from the 4 sides we could extend
	// the map "viewport" in each of the directions.  Which means the origin
	// tile's X/Y would change.  By tracking this we can find the coord location
	// of all other tiles.
	OriginX float64 // In tile width units
	OriginY float64 // In tile height units

	// Cube coordinate storage - primary data structure
	Tiles map[CubeCoord]*Tile `json:"-"` // Direct cube coordinate lookup (custom JSON handling)

	// JSON-friendly representation
	TileList []*Tile `json:"tiles"`
}

// IsWithinBounds checks if the given cube coordinates are within the map bounds
func (m *Map) IsWithinBounds(q, r int) bool {
	return q >= m.minQ && q <= m.maxQ && r >= m.minR && r <= m.maxR
}

// IsWithinBoundsCube checks if the given cube coordinate is within the map bounds
func (m *Map) IsWithinBoundsCube(coord CubeCoord) bool {
	return m.IsWithinBounds(coord.Q, coord.R)
}

// GetBounds returns the current map bounds
func (m *Map) GetBounds() (minQ, maxQ, minR, maxR int) {
	return m.minQ, m.maxQ, m.minR, m.maxR
}

// SetBounds updates the map bounds (use carefully - may invalidate existing tiles)
func (m *Map) SetBounds(minQ, maxQ, minR, maxR int) {
	m.minQ, m.maxQ, m.minR, m.maxR = minQ, maxQ, minR, maxR
}

// NewMap creates a new empty map with the specified dimensions
// evenRowsOffset parameter is deprecated and ignored (cube coordinates are universal)
func NewMapRect(numRows, numCols int) *Map {
	return NewMapWithBounds(0, numRows, 0, numCols)
}

// NewMapWithBounds creates a new empty map with the specified coordinate bounds
func NewMapWithBounds(minQ, maxQ, minR, maxR int) *Map {
	return &Map{
		minQ:     minQ,
		maxQ:     maxQ,
		minR:     minR,
		maxR:     maxR,
		Tiles:    make(map[CubeCoord]*Tile),
		TileList: make([]*Tile, 0),
	}
}

// =============================================================================
// JSON Serialization Methods
// =============================================================================

// MarshalJSON implements custom JSON marshaling for Map
func (m *Map) MarshalJSON() ([]byte, error) {
	// Convert cube map to tile list for JSON
	m.syncTileListFromMap()

	// Create a temporary struct with the same fields
	type mapJSON Map
	return json.Marshal((*mapJSON)(m))
}

// UnmarshalJSON implements custom JSON unmarshaling for Map
func (m *Map) UnmarshalJSON(data []byte) error {
	// First try to unmarshal with new bounds format
	type mapJSON Map
	if err := json.Unmarshal(data, (*mapJSON)(m)); err != nil {
		return err
	}

	// Handle backward compatibility: if bounds are not set but we have old numRows/numCols,
	// check if the JSON contains the old fields and convert them
	if m.minQ == 0 && m.maxQ == 0 && m.minR == 0 && m.maxR == 0 {
		// Try to parse old format
		var legacy struct {
			NumRows int `json:"numRows"`
			NumCols int `json:"numCols"`
		}
		if err := json.Unmarshal(data, &legacy); err == nil {
			if legacy.NumRows > 0 && legacy.NumCols > 0 {
				// Convert old format to new bounds (assuming 0,0 origin)
				m.minQ = 0
				m.maxQ = legacy.NumCols - 1
				m.minR = 0
				m.maxR = legacy.NumRows - 1
			}
		}
	}

	// Initialize the cube map if it's nil
	if m.Tiles == nil {
		m.Tiles = make(map[CubeCoord]*Tile)
	}

	// Convert tile list back to cube map
	m.syncMapFromTileList()
	return nil
}

// syncTileListFromMap converts the cube map to tile list for JSON serialization
func (m *Map) syncTileListFromMap() {
	m.TileList = make([]*Tile, 0, len(m.Tiles))
	for _, tile := range m.Tiles {
		m.TileList = append(m.TileList, tile)
	}
}

// syncMapFromTileList converts the tile list back to cube map after JSON deserialization
func (m *Map) syncMapFromTileList() {
	for _, tileWithCoord := range m.TileList {
		m.Tiles[tileWithCoord.Coord] = tileWithCoord
	}
}

// =============================================================================
// Primary Cube-Based Storage Methods
// =============================================================================

// TileAt returns the tile at the specified cube coordinate (primary method)
func (m *Map) TileAt(coord CubeCoord) *Tile {
	return m.Tiles[coord]
}

// AddTileCube adds a tile at the specified cube coordinate (primary method)
func (m *Map) AddTile(tile *Tile) {
	m.Tiles[tile.Coord] = tile
}

// DeleteTile removes the tile at the specified cube coordinate
func (m *Map) DeleteTile(coord CubeCoord) {
	delete(m.Tiles, coord)
}

// GetAllTiles returns all tiles as a map from cube coordinates to tiles
func (m *Map) CopyAllTiles() map[CubeCoord]*Tile {
	// Return a copy to prevent external modification
	result := make(map[CubeCoord]*Tile)
	for coord, tile := range m.Tiles {
		result[coord] = tile
	}
	return result
}

// =============================================================================
// Legacy Display-Based Methods (for backward compatibility)
// =============================================================================

func (m *Map) GetNeighbors(coord CubeCoord, out [6]CubeCoord) {
	// Implement this and update out with the neighbor coords
	// out[i] is coord of ith NeighborDirection
	return
}

// GetHexNeighborCoords returns the coordinates of the 6 hex neighbors
// This is no longer required.  We should get Neighbors using cubed coords
/*
func (m *Map) GetHexNeighborCoords(row, col int) [6][2]int {
	var neighbors [6][2]int

	// Hex grid neighbor calculation depends on whether we're in even or odd row
	isEvenRow := (row % 2) == 0

	if m.EvenRowsOffset() {
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
*/

// CenterXYForTile converts cube coordinates directly to pixel center x,y coordinates for rendering
// Uses normalized OriginX/OriginY (in tile units) which are then scaled by tile dimensions
// Uses odd-r layout (odd rows offset) as our fixed, consistent layout
// Based on formulas from redblobgames.com for pointy-topped hexagons
func (m *Map) CenterXYForTile(coord CubeCoord, tileWidth, tileHeight, yIncrement float64) (x, y float64) {
	// Direct cube coordinate to pixel conversion using proper hex math
	q := float64(coord.Q)
	r := float64(coord.R)

	// For pointy-topped hexagons with odd-r layout:
	// x = size * sqrt(3) * (q + r/2)
	// y = size * 3/2 * r
	// where size = tileWidth (center-to-center distance)

	// Convert normalized origin to pixel coordinates
	originPixelX := m.OriginX * tileWidth
	originPixelY := m.OriginY * tileHeight

	x = originPixelX + tileWidth*1.732050808*(q+r/2.0) // 1.732050808 ≈ sqrt(3)
	y = originPixelY + tileWidth*3.0/2.0*r

	return x, y
}

// XYToQR converts screen coordinates to cube coordinates for the map
// Given x,y screen coordinates and tile size properties, returns the CubeCoord
// Uses the Map's normalized OriginX/OriginY for proper coordinate translation
// Based on formulas from redblobgames.com for pointy-topped hexagons with odd-r layout
func (m *Map) XYToQR(x, y, tileWidth, tileHeight, yIncrement float64) CubeCoord {
	// Convert normalized origin to pixel coordinates
	originPixelX := m.OriginX * tileWidth
	originPixelY := m.OriginY * tileHeight
	
	// Translate screen coordinates to hex coordinate space by removing origin offset
	hexX := x - originPixelX
	hexY := y - originPixelY

	// For pointy-topped hexagons, convert pixel coordinates to fractional hex coordinates
	// Using inverse of the hex-to-pixel conversion formulas:
	// x = size * sqrt(3) * (q + r/2)  =>  q = (sqrt(3) * x) / (y * 3)
	// y = size * 3/2 * r             =>  r = (y * 2.0 / 3.0)

	sqrt3 := 1.732050808 // sqrt(3)

	// Calculate fractional q coordinate
	fractionalQ := (hexX * tileWidth * sqrt3) / (y * tileHeight * 3.0)

	// Calculate fractional r coordinate
	fractionalR := (hexY * tileHeight * 2.0) / 3.0

	// Round to nearest integer coordinates using cube coordinate rounding
	// This ensures we get the correct hex tile even for coordinates near boundaries
	return roundCubeCoord(fractionalQ, fractionalR)
}

// roundCubeCoord rounds fractional cube coordinates to the nearest integer cube coordinate
// Uses the cube coordinate constraint (q + r + s = 0) to ensure valid hex coordinates
// Reference: https://www.redblobgames.com/grids/hexagons-v1/#rounding
func roundCubeCoord(fractionalQ, fractionalR float64) CubeCoord {
	// Calculate s from the cube coordinate constraint: s = -q - r
	fractionalS := -fractionalQ - fractionalR

	// Round each coordinate to nearest integer
	roundedQ := int(fractionalQ + 0.5)
	roundedR := int(fractionalR + 0.5)
	roundedS := int(fractionalS + 0.5)

	// Calculate rounding deltas
	deltaQ := math.Abs(float64(roundedQ) - fractionalQ)
	deltaR := math.Abs(float64(roundedR) - fractionalR)
	deltaS := math.Abs(float64(roundedS) - fractionalS)

	// Fix the coordinate with the largest rounding error to maintain constraint
	if deltaQ > deltaR && deltaQ > deltaS {
		roundedQ = -roundedR - roundedS
	} else if deltaR > deltaS {
		roundedR = -roundedQ - roundedS
	} else {
		roundedS = -roundedQ - roundedR
	}

	// Return the rounded cube coordinate (s is implicit)
	return CubeCoord{Q: roundedQ, R: roundedR}
}

// getMapBounds calculates the pixel bounds of the entire map
func (m *Map) getMapBounds(tileWidth, tileHeight, yIncrement float64) (minX, minY, maxX, maxY float64) {
	minX = math.Inf(1)
	minY = math.Inf(1)
	maxX = math.Inf(-1)
	maxY = math.Inf(-1)

	for coord := range m.Tiles {
		// Use origin at (0,0) for bounds calculation
		x, y := m.CenterXYForTile(coord, tileWidth, tileHeight, yIncrement)

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

	return minX, minY, maxX, maxY
}

// AddLeftCols adds n columns to the left of the map, adjusting bounds and origin
func (m *Map) AddLeftCols(n int) error {
	if n <= 0 {
		return fmt.Errorf("n must be positive, got %d", n)
	}
	
	// Decrease minQ to expand leftward
	m.minQ -= n
	
	// Adjust OriginX to maintain existing tile positions
	// In odd-r layout, horizontal distance between adjacent tiles is 1 tile width unit
	// Since OriginX is in normalized tile units, we add n directly
	m.OriginX += float64(n)
	
	return nil
}

// RemoveLeftCols removes n columns from the left of the map, adjusting bounds and origin
func (m *Map) RemoveLeftCols(n int) error {
	if n <= 0 {
		return fmt.Errorf("n must be positive, got %d", n)
	}
	
	// Check if we have enough columns to remove
	currentCols := m.maxQ - m.minQ + 1
	if n >= currentCols {
		return fmt.Errorf("cannot remove %d columns from map with only %d columns", n, currentCols)
	}
	
	// Increase minQ to shrink from left
	m.minQ += n
	
	// Adjust OriginX to maintain existing tile positions
	// In odd-r layout, horizontal distance between adjacent tiles is 1 tile width unit
	// Since OriginX is in normalized tile units, we subtract n directly
	m.OriginX -= float64(n)
	
	return nil
}

// =============================================================================
// Unified Game Implementation
// =============================================================================
// This file implements the unified Game struct that combines the best parts
// of the existing core.go with the new interface architecture. It provides
// a single, coherent implementation of all core game interfaces.

// =============================================================================
// Game Creation and Initialization
// =============================================================================

func approximateCos(angle float64) float64 {
	// Simple approximation - in a real implementation, use math.Cos
	return 1.0 - angle*angle/2.0 + angle*angle*angle*angle/24.0
}

func approximateSin(angle float64) float64 {
	// Simple approximation - in a real implementation, use math.Sin
	return angle - angle*angle*angle/6.0 + angle*angle*angle*angle*angle/120.0
}

// createTestMap creates a simple test map for development
func CreateTestMap(rows, cols int) (*Map, error) {
	// Create a small test map
	gameMap := NewMapRect(rows, cols)

	// Add some test tiles
	for q := 0; q < rows; q++ {
		for r := 0; r < cols; r++ {
			// Create varied terrain
			tileType := 1 // Default to grass
			if (q+r)%4 == 0 {
				tileType = 2 // Some desert
			} else if (q+r)%7 == 0 {
				tileType = 3 // Some water
			}

			tile := NewTile(CubeCoord{q, r}, tileType)
			gameMap.AddTile(tile)
		}
	}

	// Note: Neighbor connections calculated on-demand via GetNeighbor()

	return gameMap, nil
}
