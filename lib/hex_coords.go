package weewar

import "fmt"

// =============================================================================
// Hex Cube Coordinate System
// =============================================================================
// This file implements cube coordinates for hexagonal grids, providing a 
// mathematically clean coordinate system that is independent of array storage
// and EvenRowsOffset configurations.

// CubeCoord represents a position in hex cube coordinate space
// Constraint: Q + R + S = 0 (S is calculated as -Q-R)
type CubeCoord struct {
	Q int `json:"q"`
	R int `json:"r"`
	// S is not stored since S = -Q-R always
}

// NewCubeCoord creates a new cube coordinate
func NewCubeCoord(q, r int) CubeCoord {
	return CubeCoord{Q: q, R: r}
}

// S returns the S coordinate (calculated as -Q-R)
func (c CubeCoord) S() int {
	return -c.Q - c.R
}

// IsValid checks if the cube coordinate is valid (always true by construction)
func (c CubeCoord) IsValid() bool {
	return true // Always valid since S is calculated
}

// =============================================================================
// Hex Directions (Universal - independent of EvenRowsOffset)
// =============================================================================

// HexDirections defines the 6 direction vectors in cube coordinates
// Order must match NeighborDirection enum: LEFT, TOP_LEFT, TOP_RIGHT, RIGHT, BOTTOM_RIGHT, BOTTOM_LEFT
var HexDirections = [6]CubeCoord{
	{Q: -1, R: 0},  // LEFT
	{Q: 0, R: -1},  // TOP_LEFT
	{Q: 1, R: -1},  // TOP_RIGHT
	{Q: 1, R: 0},   // RIGHT
	{Q: 0, R: 1},   // BOTTOM_RIGHT
	{Q: -1, R: 1},  // BOTTOM_LEFT
}

// Neighbor returns the neighboring cube coordinate in the specified direction
func (c CubeCoord) Neighbor(direction NeighborDirection) CubeCoord {
	dir := HexDirections[int(direction)]
	return CubeCoord{
		Q: c.Q + dir.Q,
		R: c.R + dir.R,
	}
}

// Neighbors returns all 6 neighboring cube coordinates
func (c CubeCoord) Neighbors() [6]CubeCoord {
	var neighbors [6]CubeCoord
	for i := 0; i < 6; i++ {
		neighbors[i] = c.Neighbor(NeighborDirection(i))
	}
	return neighbors
}

// =============================================================================
// Distance and Range Calculations
// =============================================================================

// Distance calculates the hex distance between two cube coordinates
func (c CubeCoord) Distance(other CubeCoord) int {
	return (abs(c.Q-other.Q) + abs(c.R-other.R) + abs(c.S()-other.S())) / 2
}

// CubeDistance calculates the hex distance between two cube coordinates (standalone function)
func CubeDistance(coord1, coord2 CubeCoord) int {
	return coord1.Distance(coord2)
}

// Range returns all cube coordinates within the specified radius
func (c CubeCoord) Range(radius int) []CubeCoord {
	var results []CubeCoord
	for q := -radius; q <= radius; q++ {
		r1 := max(-radius, -q-radius)
		r2 := min(radius, -q+radius)
		for r := r1; r <= r2; r++ {
			// s := -q - r (not needed since S is calculated)
			coord := CubeCoord{Q: c.Q + q, R: c.R + r}
			results = append(results, coord)
		}
	}
	return results
}

// Ring returns all cube coordinates at exactly the specified radius
func (c CubeCoord) Ring(radius int) []CubeCoord {
	if radius == 0 {
		return []CubeCoord{c}
	}
	
	var results []CubeCoord
	// Start at one direction and walk around the ring
	coord := c
	
	// Move to the starting point of the ring (go LEFT radius times)
	for i := 0; i < radius; i++ {
		coord = coord.Neighbor(LEFT)
	}
	
	// Walk around the ring in all 6 directions
	directions := []NeighborDirection{TOP_RIGHT, RIGHT, BOTTOM_RIGHT, BOTTOM_LEFT, LEFT, TOP_LEFT}
	for _, direction := range directions {
		for i := 0; i < radius; i++ {
			results = append(results, coord)
			coord = coord.Neighbor(direction)
		}
	}
	
	return results
}

// =============================================================================
// Array Coordinate Conversion
// =============================================================================

// ArrayToHex converts array coordinates (row, col) to cube coordinates
// Takes into account the map's EvenRowsOffset configuration
func (m *Map) ArrayToHex(row, col int) CubeCoord {
	var q, r int
	r = row
	
	if m.EvenRowsOffset() {
		// Even rows are offset to the right (flat-top hexes)
		q = col - (row - (row&1)) / 2
	} else {
		// Odd rows are offset to the right (flat-top hexes)
		q = col - (row + (row&1)) / 2
	}
	
	return NewCubeCoord(q, r)
}

// HexToArray converts cube coordinates to array coordinates (row, col)
// Takes into account the map's EvenRowsOffset configuration
func (m *Map) HexToArray(coord CubeCoord) (row, col int) {
	row = coord.R
	
	if m.EvenRowsOffset() {
		// Even rows are offset to the right (flat-top hexes)
		col = coord.Q + (coord.R - (coord.R&1)) / 2
	} else {
		// Odd rows are offset to the right (flat-top hexes)
		col = coord.Q + (coord.R + (coord.R&1)) / 2
	}
	
	return row, col
}

// =============================================================================
// Map Integration Methods
// =============================================================================

// Note: TileAtCube moved to game.go to use direct cube storage

// TileAtCubeQR returns the tile at the specified cube coordinate (Q, R)
func (m *Map) TileAtCubeQR(q, r int) *Tile {
	return m.TileAtCube(NewCubeCoord(q, r))
}

// GetNeighborCube returns the neighboring tile in the specified direction using cube coordinates
func (m *Map) GetNeighborCube(coord CubeCoord, direction NeighborDirection) *Tile {
	neighborCoord := coord.Neighbor(direction)
	return m.TileAtCube(neighborCoord)
}

// GetTileNeighborsCube returns all 6 neighboring tiles using cube coordinates
func (m *Map) GetTileNeighborsCube(coord CubeCoord) []*Tile {
	neighborCoords := coord.Neighbors()
	neighbors := make([]*Tile, 6)
	
	for i, neighborCoord := range neighborCoords {
		neighbors[i] = m.TileAtCube(neighborCoord)
	}
	
	return neighbors
}

// Note: Utility functions abs, max, min are defined in board.go

// =============================================================================
// Debugging and Display Helpers
// =============================================================================

// String returns a string representation of the cube coordinate
func (c CubeCoord) String() string {
	return fmt.Sprintf("(%d,%d,%d)", c.Q, c.R, c.S())
}

// ToArrayString returns the equivalent array coordinates as a string (for debugging)
func (c CubeCoord) ToArrayString(m *Map) string {
	row, col := m.HexToArray(c)
	return fmt.Sprintf("[%d,%d]", row, col)
}