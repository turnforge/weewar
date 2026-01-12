package lib

import (
	"fmt"
	"strings"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

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
	for i := range 6 {
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

// ParseDirection parses a direction string (L, R, UL, UR, DL, DR, etc.) to NeighborDirection
// Supports multiple naming conventions:
//   - L, R (original)
//   - TL, TR, BL, BR (consistent with optionsd output)
//   - UL, UR, DL, DR (user-friendly aliases for upper/down)
//   - LU, RU, LD, RD (alternative notation)
func ParseDirection(input string) (NeighborDirection, error) {
	input = strings.ToUpper(strings.TrimSpace(input))

	switch input {
	case "L", "LEFT":
		return LEFT, nil
	case "TL", "LU", "UL", "TOPLEFT", "UPPERLEFT":
		return TOP_LEFT, nil
	case "TR", "RU", "UR", "TOPRIGHT", "UPPERRIGHT":
		return TOP_RIGHT, nil
	case "R", "RIGHT":
		return RIGHT, nil
	case "BR", "RD", "DR", "BOTTOMRIGHT", "DOWNRIGHT":
		return BOTTOM_RIGHT, nil
	case "BL", "LD", "DL", "BOTTOMLEFT", "DOWNLEFT":
		return BOTTOM_LEFT, nil
	default:
		return -1, fmt.Errorf("invalid direction: %s (valid: L, R, TL, TR, BL, BR)", input)
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

// =============================================================================
// Coordinate Map Key Functions
// =============================================================================
// These functions provide a consistent way to use coordinates as map keys
// in the format "q,r" for use with proto map<string, T> fields.

// CoordKey returns the map key for a coordinate in "q,r" format
func CoordKey(q, r int32) string {
	return fmt.Sprintf("%d,%d", q, r)
}

// CoordKeyFromAxial returns the map key for an AxialCoord in "q,r" format
func CoordKeyFromAxial(coord AxialCoord) string {
	return fmt.Sprintf("%d,%d", coord.Q, coord.R)
}

// ParseCoordKey parses a "q,r" format string back to an AxialCoord
func ParseCoordKey(key string) (AxialCoord, error) {
	var q, r int
	_, err := fmt.Sscanf(key, "%d,%d", &q, &r)
	if err != nil {
		return AxialCoord{}, fmt.Errorf("invalid coord key %q: %w", key, err)
	}
	return AxialCoord{Q: q, R: r}, nil
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

func CubeToEvenR(x, y, z int) (row, col int) {
	col = x + (z+(z&1))/2
	row = z
	return
}

func EvenRToCube(row, col int) (x, y, z int) {
	x = col - (row+(row&1))/2
	z = row
	y = -x - z
	return
}

// HexToRowCol converts Axial coordinates to display coordinates (row, col)
// Uses a standard hex-to-array conversion (odd-row offset style)
func HexToRowCol(coord AxialCoord, evenrow bool) (row, col int) {
	/*
		row = coord.R
		col = coord.Q + (coord.R+(coord.R&1))/2
		return row, col
	*/
	// cube_to_oddr(cube):
	x, _, z := AxialToCube(coord.Q, coord.R)
	col = x + (z-(z&1))/2
	if evenrow {
		col = x + (z+(z&1))/2
	}
	row = z
	return row, col
}

// RowColToHex converts display coordinates (row, col) to cube coordinates
// Uses a standard array-to-hex conversion (odd-row offset style)
func RowColToHex(row, col int, evenrow bool) AxialCoord {
	// q := col - (row+(row&1))/2 return NewAxialCoord(q, row)
	// oddr_to_cube(hex):
	x := col - (row-(row&1))/2
	if evenrow {
		x = col - (row+(row&1))/2
	}
	z := row
	y := -x - z
	q, r := CubeToAxial(x, y, z)
	return AxialCoord{q, r}
}

// =============================================================================
// Pixel Coordinate Conversion (for rendering)
// =============================================================================

// Default tile dimensions matching the TypeScript hexUtils.ts
const (
	DefaultTileWidth  = 64
	DefaultTileHeight = 64
	DefaultYIncrement = 48
)

// RenderOptions holds tile dimension parameters for pixel calculations
type RenderOptions struct {
	TileWidth           int  // Width of each hex tile in pixels
	TileHeight          int  // Height of each hex tile in pixels
	YIncrement          int  // Vertical spacing between rows (typically 3/4 of TileHeight for pointy-top)
	ShowUnitLabels      bool // Show unit labels (Shortcut:MP/Health) below units
	ShowTileLabels      bool // Show tile labels (Shortcut) below tile
	EvenRowOffsetCoords bool
}

// DefaultRenderOptions returns standard rendering options
func DefaultRenderOptions() *RenderOptions {
	return &RenderOptions{
		TileWidth:  DefaultTileWidth,
		TileHeight: DefaultTileHeight,
		YIncrement: DefaultYIncrement,
	}
}

// HexToPixel converts hex coordinates to pixel coordinates (top-left of tile)
// This matches the Go CenterXYForTile and TypeScript hexToPixel implementations
func HexToPixel(coord AxialCoord, opts *RenderOptions) (x, y int) {
	if opts == nil {
		opts = DefaultRenderOptions()
	}
	row, col := HexToRowCol(coord, opts.EvenRowOffsetCoords)
	y = opts.YIncrement * row
	x = opts.TileWidth * col
	// Odd rows are offset by half a tile width
	if (row & 1) == 1 {
		if opts.EvenRowOffsetCoords {
		} else {
			x += opts.TileWidth / 2
		}
	}
	return x, y
}

// HexToPixelInt32 is a convenience wrapper for int32 coordinates
func HexToPixelInt32(q, r int32, opts *RenderOptions) (x, y int) {
	return HexToPixel(AxialCoord{Q: int(q), R: int(r)}, opts)
}

// ComputeWorldBounds calculates the pixel bounding box for tiles and units
// Returns bounds where (MinX, MinY) is the top-left corner of the top-left-most tile
func ComputeWorldBounds(tiles map[string]*v1.Tile, units map[string]*v1.Unit, opts *RenderOptions) WorldBounds {
	if len(tiles) == 0 && len(units) == 0 {
		return WorldBounds{}
	}
	if opts == nil {
		opts = DefaultRenderOptions()
	}

	minX, minY := int(^uint(0)>>1), int(^uint(0)>>1) // Max int
	maxX, maxY := -minX-1, -minY-1                   // Min int

	// Process tiles
	for _, tile := range tiles {
		x, y := HexToPixelInt32(tile.Q, tile.R, opts)
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x+opts.TileWidth > maxX {
			maxX = x + opts.TileWidth
		}
		if y+opts.TileHeight > maxY {
			maxY = y + opts.TileHeight
		}
	}

	// Process units (in case any are outside tile bounds)
	for _, unit := range units {
		x, y := HexToPixelInt32(unit.Q, unit.R, opts)
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x+opts.TileWidth > maxX {
			maxX = x + opts.TileWidth
		}
		if y+opts.TileHeight > maxY {
			maxY = y + opts.TileHeight
		}
	}

	return WorldBounds{
		MinX:   minX,
		MinY:   minY,
		MaxX:   maxX,
		MaxY:   maxY,
		Width:  maxX - minX,
		Height: maxY - minY,
	}
}
