package weewar

import (
	"encoding/json"
	"fmt"
	"math"
)

type MapBounds struct {
	MinX, MinY, MaxX, MaxY float64
	MinQ, MinR, MaxQ, MaxR int
	MinXCoord, MinYCoord   AxialCoord
	MaxXCoord, MaxYCoord   AxialCoord
	StartingCoord          AxialCoord
	StartingX              float64
}

// Map represents the game map with hex grid topology
type Map struct {
	// Coordinate bounds - These can be evaluated.
	minQ int `json:"-"` // Minimum Q coordinate (inclusive)
	maxQ int `json:"-"` // Maximum Q coordinate (inclusive)
	minR int `json:"-"` // Minimum R coordinate (inclusive)
	maxR int `json:"-"` // Maximum R coordinate (inclusive)

	boundsChanged bool
	lastMapBounds MapBounds

	// Cube coordinate storage - primary data structure
	Tiles map[AxialCoord]*Tile `json:"-"` // Direct cube coordinate lookup (custom JSON handling)

	// JSON-friendly representation
	TileList []*Tile `json:"tiles"`
}

// IsWithinBounds checks if the given cube coordinates are within the map bounds
func (m *Map) IsWithinBounds(q, r int) bool {
	return q >= m.minQ && q <= m.maxQ && r >= m.minR && r <= m.maxR
}

// IsWithinBoundsCube checks if the given cube coordinate is within the map bounds
func (m *Map) IsWithinBoundsCube(coord AxialCoord) bool {
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
		Tiles:    make(map[AxialCoord]*Tile),
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
		m.Tiles = make(map[AxialCoord]*Tile)
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
func (m *Map) TileAt(coord AxialCoord) *Tile {
	return m.Tiles[coord]
}

// AddTileCube adds a tile at the specified cube coordinate (primary method)
func (m *Map) AddTile(tile *Tile) {
	q, r := tile.Coord.Q, tile.Coord.R
	if q < m.minQ || q > m.maxQ || r < m.minR || r > m.maxR {
		m.boundsChanged = true
	}
	m.Tiles[tile.Coord] = tile
}

// DeleteTile removes the tile at the specified cube coordinate
func (m *Map) DeleteTile(coord AxialCoord) {
	delete(m.Tiles, coord)
}

// GetAllTiles returns all tiles as a map from cube coordinates to tiles
func (m *Map) CopyAllTiles() map[AxialCoord]*Tile {
	// Return a copy to prevent external modification
	result := make(map[AxialCoord]*Tile)
	for coord, tile := range m.Tiles {
		result[coord] = tile
	}
	return result
}

// =============================================================================
// Legacy Display-Based Methods (for backward compatibility)
// =============================================================================

// CenterXYForTile converts cube coordinates directly to pixel center x,y coordinates for rendering
// Uses odd-r layout (odd rows offset) as our fixed, consistent layout
// Based on formulas from redblobgames.com for pointy-topped hexagons
func (m *Map) CenterXYForTile(coord AxialCoord, tileWidth, tileHeight, yIncrement float64) (x, y float64) {
	// Direct cube coordinate to pixel conversion using proper hex math
	if false {
		q := float64(coord.Q)
		r := float64(coord.R)

		// For pointy-topped hexagons with odd-r layout:
		// x = size * sqrt(3) * (q + r/2)
		// y = size * 3/2 * r
		size := tileWidth / SQRT3

		// Convert normalized origin to pixel coordinates
		// Note: Both OriginX and OriginY are in tile width units for consistency with hex geometry

		// tileWidth = size * SQRT3
		x = tileWidth * (q + r/2.0) // 1.732050808 ≈ sqrt(3)
		y = size * 1.5 * r
	} else {
		row, col := HexToRowCol(coord)
		// fmt.Printf("HexToRow, QR: %s, RowCol: (%d, %d)\n", coord, row, col)
		y = yIncrement * float64(row)  // + (tileHeight / 2)
		x = tileWidth * (float64(col)) //  + 0.5)
		if (row & 1) == 1 {
			x += tileWidth / 2
		}

		// x = tileWidth * (float64(col) + 0.5*float64(row&1))
	}

	return x, y
}

// XYToQR converts screen coordinates to cube coordinates for the map
// Given x,y screen coordinates and tile size properties, returns the AxialCoord
// Uses the Map's normalized OriginX/OriginY for proper coordinate translation
// Based on formulas from redblobgames.com for pointy-topped hexagons with odd-r layout
func (m *Map) XYToQR(x, y, tileWidth, tileHeight, yIncrement float64) (coord AxialCoord) {
	if false {
		// Convert normalized origin to pixel coordinates
		// Note: Both OriginX and OriginY are in tile width units for consistency with hex geometry
		originPixelX := 0.0 // m.OriginX * tileWidth
		originPixelY := 0.0 // m.OriginY * tilHeight

		// Translate screen coordinates to hex coordinate space by removing origin offset
		hexX := x - originPixelX
		hexY := y - originPixelY

		// For pointy-topped hexagons, convert pixel coordinates to fractional hex coordinates
		// Using inverse of the hex-to-pixel conversion formulas:
		// x = size * sqrt(3) * (q + r/2)  =>  q = (sqrt(3) * x) / (y * 3)
		// y = size * 3/2 * r             =>  r = (y * 2.0 / 3.0)
		size := tileWidth / SQRT3

		// Calculate fractional q coordinate
		fractionalQ := (hexX*SQRT3 - y) / (size * 3.0)

		// Calculate fractional r coordinate
		fractionalR := (hexY * 2.0) / (3.0 * size)

		// Round to nearest integer coordinates using cube coordinate rounding
		// This ensures we get the correct hex tile even for coordinates near boundaries
		coord = roundAxialCoord(fractionalQ, fractionalR)

		fmt.Println("X,Y: ", x, y)
		fmt.Println("FQ, FR, FQ+FR: ", fractionalQ, fractionalR, fractionalQ+fractionalR)
	} else { // given we can have non "equal" side length hexagons, easier to do this by converting to row/col first
		row := int((y + tileHeight/2) / yIncrement)

		halfDists := int(1 + math.Abs(x*2/tileWidth))
		if (row & 1) != 0 {
			halfDists = int(1 + math.Abs((x-tileWidth/2)*2/tileWidth))
		}
		// log.Println("Half Dists: ", halfDists)
		col := halfDists / 2
		if x < 0 {
			col = -col
		}
		// col := int((x + tileWidth/2) / tileWidth)
		coord = RowColToHex(row, col)
		// fmt.Println("X,Y: ", x, y)
		// fmt.Println("Row, Col: ", row, col)
	}
	// fmt.Println("Final Coord: ", coord)
	// fmt.Println("======")
	return
}

// roundAxialCoord rounds fractional cube coordinates to the nearest integer cube coordinate
// Uses the cube coordinate constraint (q + r + s = 0) to ensure valid hex coordinates
// Reference: https://www.redblobgames.com/grids/hexagons-v1/#rounding
func roundAxialCoord(fractionalQ, fractionalR float64) AxialCoord {
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
	return AxialCoord{Q: roundedQ, R: roundedR}
}

// getMapBounds calculates the pixel bounds of the entire map
// TODO - cache this and only update when bounds changed beyond min/max Q/R
func (m *Map) GetMapBounds(tileWidth, tileHeight, yIncrement float64) MapBounds {
	if true || !m.boundsChanged {
		// TODO - return last avlues
		// m.boundsChanged = false
		minX := math.Inf(1)
		minY := math.Inf(1)
		maxX := math.Inf(-1)
		maxY := math.Inf(-1)
		minQ := int(math.Inf(1))
		minR := int(math.Inf(1))
		maxQ := int(math.Inf(-1))
		maxR := int(math.Inf(-1))
		startingX := 0.0
		var minXCoord, minYCoord, maxXCoord, maxYCoord, startingCoord AxialCoord

		for coord := range m.Tiles {
			// Use origin at (0,0) for bounds calculation
			x, y := m.CenterXYForTile(coord, tileWidth, tileHeight, yIncrement)
			// fmt.Printf("Tile Coords: QR: %s, XY: (%f, %f)\n", coord, x, y)

			if coord.Q < minQ {
				minQ = coord.Q
			}
			if coord.Q > maxQ {
				maxQ = coord.Q
			}
			if coord.R < minR {
				minR = coord.R
			}
			if coord.R > maxR {
				maxR = coord.R
			}
			if x < minX {
				minX = x
				minXCoord = coord
			}
			if y < minY {
				minY = y
				minYCoord = coord
			}
			if x+tileWidth > maxX {
				maxX = x + tileWidth
				maxXCoord = coord
			}
			if y+tileHeight > maxY {
				maxY = y + tileHeight
				maxYCoord = coord
			}
		}

		// Now that we have minY and minX coords, we can findout starting by walking "left" from minYCoord and "up" from
		// minXcoord and see where they meet
		// NOTE - the rows "decrease" as we go up vertically
		minYRow := minYCoord.R // S coord is same in a row for pointy-top hexes
		minXRow := minXCoord.R // S coord is same in a row for pointy-top hexes

		// if minx == miny or both minXCoord and minYCoord are in the same row then easy
		startingCoord = minXCoord
		startingX = minX

		if minXCoord != minYCoord || minXRow != minYRow {
			// The hard case
			if minXRow < minYRow {
				// because X should be "below" Y so it should have a higher row number than minYCoord
				panic(fmt.Sprintf("minXRow (%d, %f) cannot be less than minYRow (%d, %f)??", minXRow, minX, minYRow, minY))
			}
			startingCoord = minXCoord
			for i := minXRow; i >= minYRow; i-- {
				if (i & 1) == 0 {
					// Always take the "Right" path first so we are guaranteed
					// to always be on a tile whose X Coordinate is >= minX
					startingCoord = startingCoord.Neighbor(TOP_RIGHT)
				} else {
					startingCoord = startingCoord.Neighbor(TOP_LEFT)
				}
			}
		}

		// If distance was odd then we would have a half tile width shift to the right
		if ((minXRow - minYRow) & 1) == 0 {
			startingX += tileWidth / 2.0
		}
		// startingX, _ = m.CenterXYForTile(startingCoord, tileWidth, tileHeight, yIncrement)
		// fmt.Printf("StartingX, StartingCoord: ", startingX, startingCoord)

		m.lastMapBounds.MinX = minX
		m.lastMapBounds.MinY = minY
		m.lastMapBounds.MaxX = maxX
		m.lastMapBounds.MaxY = maxY
		m.lastMapBounds.MinQ = minQ
		m.lastMapBounds.MinR = minR
		m.lastMapBounds.MaxQ = maxQ
		m.lastMapBounds.MaxR = maxR
		m.lastMapBounds.StartingX = startingX
		m.lastMapBounds.MinXCoord = minXCoord
		m.lastMapBounds.MinYCoord = minYCoord
		m.lastMapBounds.MaxXCoord = maxXCoord
		m.lastMapBounds.MaxYCoord = maxYCoord
		m.lastMapBounds.StartingCoord = startingCoord
	}
	return m.lastMapBounds
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

			tile := NewTile(AxialCoord{q, r}, tileType)
			gameMap.AddTile(tile)
		}
	}

	// Note: Neighbor connections calculated on-demand via GetNeighbor()

	return gameMap, nil
}
