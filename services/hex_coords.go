package services

import "fmt"

// Position represents a coordinate position (row, col)
type Position = AxialCoord

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

// =============================================================================
// Hex Cube Coordinate System
// =============================================================================
// This file implements cube coordinates for hexagonal grids, providing a
// mathematically clean coordinate system that is independent of array storage
// and EvenRowsOffset configurations.
type CubeCoord struct {
	X int `json:"x"`
	Y int `json:"y"`
	// S is not stored since S = -Q-R always
}

// AxialCoord represents a position in hex cube coordinate space
// Constraint: Q + R + S = 0 (S is calculated as -Q-R)
type AxialCoord struct {
	Q int `json:"q"`
	R int `json:"r"`
	// S is not stored since S = -Q-R always
}

// NewAxialCoord creates a new cube coordinate
func NewAxialCoord(q, r int) AxialCoord {
	return AxialCoord{Q: q, R: r}
}

// S returns the S coordinate (calculated as -Q-R)
func (c AxialCoord) S() int {
	return -c.Q - c.R
}

func CoordFromInt32(q, r int32) AxialCoord {
	return AxialCoord{int(q), int(r)}
}

// =============================================================================
// Hex Directions (Universal - independent of EvenRowsOffset)
// =============================================================================

// AxialCoordNeighbors defines the 6 direction vectors in cube coordinates
// Order must match NeighborDirection enum: LEFT, TOP_LEFT, TOP_RIGHT, RIGHT, BOTTOM_RIGHT, BOTTOM_LEFT
var AxialCoordNeighbors = [6]AxialCoord{
	{Q: -1, R: 0}, // LEFT
	{Q: 0, R: -1}, // TOP_LEFT
	{Q: 1, R: -1}, // TOP_RIGHT
	{Q: 1, R: 0},  // RIGHT
	{Q: 0, R: 1},  // BOTTOM_RIGHT
	{Q: -1, R: 1}, // BOTTOM_LEFT
}

// Neighbor returns the neighboring cube coordinate in the specified direction
func (c AxialCoord) Neighbor(direction NeighborDirection) AxialCoord {
	dir := AxialCoordNeighbors[int(direction)]
	return AxialCoord{
		Q: c.Q + dir.Q,
		R: c.R + dir.R,
	}
}

// Neighbors returns all 6 neighboring cube coordinates
func (c AxialCoord) Neighbors(out *[6]AxialCoord) {
	for i := 0; i < 6; i++ {
		out[i] = c.Neighbor(NeighborDirection(i))
	}
}

// =============================================================================
// Distance and Range Calculations
// =============================================================================

// Distance calculates the hex distance between two cube coordinates
func (c AxialCoord) Distance(other AxialCoord) int {
	return (abs(c.Q-other.Q) + abs(c.R-other.R) + abs(c.S()-other.S())) / 2
}

// CubeDistance calculates the hex distance between two cube coordinates (standalone function)
func CubeDistance(coord1, coord2 AxialCoord) int {
	return coord1.Distance(coord2)
}

// Range returns all cube coordinates within the specified radius
func (c AxialCoord) Range(radius int) []AxialCoord {
	var results []AxialCoord
	for q := -radius; q <= radius; q++ {
		r1 := max(-radius, -q-radius)
		r2 := min(radius, -q+radius)
		for r := r1; r <= r2; r++ {
			// s := -q - r (not needed since S is calculated)
			coord := AxialCoord{Q: c.Q + q, R: c.R + r}
			results = append(results, coord)
		}
	}
	return results
}

// Ring returns all cube coordinates at exactly the specified radius
func (c AxialCoord) Ring(radius int) []AxialCoord {
	if radius == 0 {
		return []AxialCoord{c}
	}

	var results []AxialCoord
	// Start at one direction and walk around the ring
	coord := c

	// Move to the starting point of the ring (go LEFT radius times)
	for range radius {
		coord = coord.Neighbor(LEFT)
	}

	// Walk around the ring in all 6 directions
	directions := []NeighborDirection{TOP_RIGHT, RIGHT, BOTTOM_RIGHT, BOTTOM_LEFT, LEFT, TOP_LEFT}
	for _, direction := range directions {
		for range radius {
			results = append(results, coord)
			coord = coord.Neighbor(direction)
		}
	}

	return results
}

// =============================================================================
// Array Coordinate Conversion
// =============================================================================

// =============================================================================
// Direction Utilities
// =============================================================================

// GetDirection determines the direction from one hex coordinate to an adjacent hex
// Returns -1 if the coordinates are not adjacent
func GetDirection(from, to AxialCoord) NeighborDirection {
	// Calculate the difference
	dq := to.Q - from.Q
	dr := to.R - from.R

	// Check each possible direction
	for dir := range 6 {
		neighbor := AxialCoordNeighbors[dir]
		if neighbor.Q == dq && neighbor.R == dr {
			return NeighborDirection(dir)
		}
	}

	// Not adjacent
	return NeighborDirection(-1)
}

// DirectionToString converts a NeighborDirection to an ASCII arrow
func DirectionToString(dir NeighborDirection) string {
	switch dir {
	case LEFT:
		return "←"
	case TOP_LEFT:
		return "↖"
	case TOP_RIGHT:
		return "↗"
	case RIGHT:
		return "→"
	case BOTTOM_RIGHT:
		return "↘"
	case BOTTOM_LEFT:
		return "↙"
	default:
		return "?"
	}
}

// DirectionToCode converts a NeighborDirection to a short code (L, LU, etc)
func DirectionToCode(dir NeighborDirection) string {
	switch dir {
	case LEFT:
		return "L"
	case TOP_LEFT:
		return "LU"
	case TOP_RIGHT:
		return "RU"
	case RIGHT:
		return "R"
	case BOTTOM_RIGHT:
		return "RD"
	case BOTTOM_LEFT:
		return "LD"
	default:
		return "?"
	}
}

// DirectionToLongString converts a NeighborDirection to a descriptive string
func DirectionToLongString(dir NeighborDirection) string {
	switch dir {
	case LEFT:
		return "Left"
	case TOP_LEFT:
		return "Top-Left"
	case TOP_RIGHT:
		return "Top-Right"
	case RIGHT:
		return "Right"
	case BOTTOM_RIGHT:
		return "Bottom-Right"
	case BOTTOM_LEFT:
		return "Bottom-Left"
	default:
		return "Unknown"
	}
}

// =============================================================================
// Debugging and Display Helpers
// =============================================================================

// String returns a string representation of the cube coordinate
func (c AxialCoord) String() string {
	return fmt.Sprintf("(%d,%d)", c.Q, c.R)
	// return fmt.Sprintf("(%d,%d,%d)", c.Q, c.R, c.S())
}

func (c AxialCoord) Plus(dQ, dR int) AxialCoord {
	return AxialCoord{c.Q + dQ, c.R + dR}
}

// Some functions to work with hex tiles

// Using this we can evaluate a lot of things
type HexTile struct {
	TileWidth      float64
	TileHeight     float64
	LeftSideHeight float64
}

func CubeToAxial(x, y, z int) (q, r int) {
	return x, z
}

func AxialToCube(q, r int) (x, y, z int) {
	return q, (-q - r), r
}

func CubeToOddR(x, y, z int) (row, col int) {
	col = x + (z-(z&1))/2
	row = z
	return
}

func OddRToCube(row, col int) (x, y, z int) {
	x = col - (row-(row&1))/2
	z = row
	y = -x - z
	return
}

// HexToRowCol converts cube coordinates to display coordinates (row, col)
// Uses a standard hex-to-array conversion (odd-row offset style)
func HexToRowCol(coord AxialCoord) (row, col int) {
	/*
		row = coord.R
		col = coord.Q + (coord.R+(coord.R&1))/2
		return row, col
	*/
	// cube_to_oddr(cube):
	x, _, z := AxialToCube(coord.Q, coord.R)
	col = x + (z-(z&1))/2
	row = z
	return row, col
}

// RowColToHex converts display coordinates (row, col) to cube coordinates
// Uses a standard array-to-hex conversion (odd-row offset style)
func RowColToHex(row, col int) AxialCoord {
	// q := col - (row+(row&1))/2 return NewAxialCoord(q, row)
	// oddr_to_cube(hex):
	x := col - (row-(row&1))/2
	z := row
	y := -x - z
	q, r := CubeToAxial(x, y, z)
	return AxialCoord{q, r}
}
