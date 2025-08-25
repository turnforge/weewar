package weewar

import (
	"encoding/json"
	"fmt"
	"iter"
	"maps"
	"math"
	"slices"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

// =============================================================================
// World - Pure Game State Container
// =============================================================================

type WorldBounds struct {
	MinX, MinY, MaxX, MaxY float64
	MinQ, MinR, MaxQ, MaxR int
	MinXCoord, MinYCoord   AxialCoord
	MaxXCoord, MaxYCoord   AxialCoord
	StartingCoord          AxialCoord
	StartingX              float64
}

// World represents the pure game state without any rendering or UI concerns.
// This is the single source of truth for all game data.
type World struct {
	// By having a parent world - we are able to offer "pushed" environments - to test things out like starting
	// transactions etc
	parent *World

	// JSON-friendly representation
	Name string
	// PlayerCount int `json:"playerCount"` // Number of players in the game

	// Ways to identify various kinds of units and tiles
	unitsByPlayer [][]*v1.Unit            `json:"-"` // All units in the game world by player ID
	unitsByCoord  map[AxialCoord]*v1.Unit `json:"-"` // All units in the game world by player ID
	tilesByCoord  map[AxialCoord]*v1.Tile `json:"-"` // Direct cube coordinate lookup (custom JSON handling)

	// In case we are pushed environment this will tell us
	// if a unit was "deleted" in this layer so not to recurse
	//up when looking up a missing unit
	unitDeleted map[AxialCoord]bool `json:"-"`
	// Same as above but for tiles
	tileDeleted map[AxialCoord]bool `json:"-"` // Direct cube coordinate lookup (custom JSON handling)

	// Transaction layer counters for efficient NumUnits calculation
	unitsAdded   int32 `json:"-"` // Number of units added in this layer
	unitsDeleted int32 `json:"-"` // Number of units deleted from parent in this layer

	// Coordinate bounds - These can be evaluated.
	minQ int `json:"-"` // Minimum Q coordinate (inclusive)
	maxQ int `json:"-"` // Maximum Q coordinate (inclusive)
	minR int `json:"-"` // Minimum R coordinate (inclusive)
	maxR int `json:"-"` // Maximum R coordinate (inclusive)

	boundsChanged   bool
	lastWorldBounds WorldBounds

	// Observer pattern for state changes
	WorldSubject `json:"-"`
}

// Note: Position type is already defined in game_interface.go

// =============================================================================
// World Creation and Management
// =============================================================================

// NewWorld creates a new game world with the specified parameters
func NewWorld(name string) *World {
	w := &World{
		Name:         name,
		tilesByCoord: map[AxialCoord]*v1.Tile{},
		unitsByCoord: map[AxialCoord]*v1.Unit{},
		tileDeleted:  map[AxialCoord]bool{},
		unitDeleted:  map[AxialCoord]bool{},
	}

	return w
}

func (w *World) Push() *World {
	out := NewWorld(w.Name)
	out.parent = w
	return out
}

// Pop returns the parent world (for transaction rollback)
func (w *World) Pop() *World {
	return w.parent
}

// =============================================================================
// World State Access Methods
// =============================================================================

func (w *World) PlayerCount() int32 {
	if w.parent != nil {
		return w.parent.PlayerCount() // FIX: Call parent.PlayerCount(), not w.PlayerCount()
	}
	return int32(len(w.unitsByPlayer) - 1)
}

// Iterate a given coord's neighbors (if they are also in the map)
func (w *World) Neighbors(coord AxialCoord) iter.Seq2[AxialCoord, *v1.Tile] {
	var neighbors [6]AxialCoord
	coord.Neighbors(&neighbors)
	return func(yield func(AxialCoord, *v1.Tile) bool) {
		for _, neigh := range neighbors {
			// Check if neighbor tile exists and is passable
			tile := w.TileAt(neigh)
			if tile == nil {
				continue // Invalid tile
			}
			if !yield(neigh, tile) {
				return
			}
		}
	}
}

func (w *World) TilesByCoord() iter.Seq2[AxialCoord, *v1.Tile] {
	// Merged iteration: child tiles override parent tiles, respect deletions
	return func(yield func(AxialCoord, *v1.Tile) bool) {
		seen := make(map[AxialCoord]bool)

		// First iterate current layer (child overrides parent)
		for coord, tile := range w.tilesByCoord {
			seen[coord] = true
			if !yield(coord, tile) {
				return
			}
		}

		// Then iterate parent layers for unseen coordinates
		if w.parent != nil {
			for coord, tile := range w.parent.TilesByCoord() {
				// Skip if already seen in child or explicitly deleted in child
				if seen[coord] || w.tileDeleted[coord] {
					continue
				}
				if !yield(coord, tile) {
					return
				}
			}
		}
	}
}

func (w *World) NumUnits() int32 {
	if w.parent != nil {
		// Transaction layer: parent count + added - deleted
		return w.parent.NumUnits() + w.unitsAdded - w.unitsDeleted
	}
	// Root layer: just count the units in this layer
	return int32(len(w.unitsByCoord))
}

func (w *World) UnitsByCoord() iter.Seq2[AxialCoord, *v1.Unit] {
	// Merged iteration: child units override parent units, respect deletions
	return func(yield func(AxialCoord, *v1.Unit) bool) {
		seen := make(map[AxialCoord]bool)

		// First iterate current layer (child overrides parent)
		for coord, unit := range w.unitsByCoord {
			seen[coord] = true
			if !yield(coord, unit) {
				return
			}
		}

		// Then iterate parent layers for unseen coordinates
		if w.parent != nil {
			for coord, unit := range w.parent.UnitsByCoord() {
				// Skip if already seen in child or explicitly deleted in child
				if seen[coord] || w.unitDeleted[coord] {
					continue
				}
				if !yield(coord, unit) {
					return
				}
			}
		}
	}
}

// UnitAt returns the unit at the specified coordinate, respecting transaction deletions
func (w *World) UnitAt(coord AxialCoord) (out *v1.Unit) {
	out = w.unitsByCoord[coord]
	if out == nil && w.parent != nil && !w.unitDeleted[coord] {
		out = w.parent.UnitAt(coord)
	}
	return
}

// TileAt returns the tile at the specified cube coordinates
func (w *World) TileAt(coord AxialCoord) (out *v1.Tile) {
	out = w.tilesByCoord[coord]
	if out == nil && w.parent != nil {
		out = w.parent.TileAt(coord)
	}
	return
}

// GetPlayerUnits returns all units belonging to the specified player
func (w *World) GetPlayerUnits(playerID int) []*v1.Unit {
	// TODO - handle the case of doing a "merged" iteration with parents if anything is missing here
	// or conversely iterate parent and only return parent's K,V value if it is not in this layer
	return w.unitsByPlayer[playerID]
}

// =============================================================================
// World State Mutation Methods
// =============================================================================

// SetTileTypeCube changes the terrain type at the specified cube coordinates
func (w *World) SetTileType(coord AxialCoord, terrainType int) bool {
	// Get or create tile at position
	tile := w.TileAt(coord)
	if tile == nil {
		// Create new tile
		tile = NewTile(coord, terrainType)
		w.AddTile(tile)
	} else {
		// Update existing tile
		tile.TileType = int32(terrainType)
	}

	return true
}

// AddTileCube adds a tile at the specified cube coordinate (primary method)
func (w *World) AddTile(tile *v1.Tile) {
	coord := TileGetCoord(tile)
	q, r := coord.Q, coord.R
	if q < w.minQ || q > w.maxQ || r < w.minR || r > w.maxR {
		w.boundsChanged = true
	}
	w.tileDeleted[coord] = false
	w.tilesByCoord[coord] = tile
}

// DeleteTile removes the tile at the specified cube coordinate
func (w *World) DeleteTile(coord AxialCoord) {
	w.tileDeleted[coord] = true
	delete(w.tilesByCoord, coord)
}

// AddUnit adds a new unit to the world at the specified position
func (w *World) AddUnit(unit *v1.Unit) (oldunit *v1.Unit, err error) {
	if unit == nil {
		return nil, fmt.Errorf("unit is nil")
	}

	playerID := int(unit.Player)
	if playerID < 0 {
		return nil, fmt.Errorf("invalid player ID: %d", playerID)
	}
	for playerID >= len(w.unitsByPlayer) {
		w.unitsByPlayer = append(w.unitsByPlayer, nil)
	}

	coord := UnitGetCoord(unit)
	oldunit = w.UnitAt(coord)

	// Update transaction counters
	if w.parent != nil {
		// Transaction layer: track if this is a new unit or replacing a parent unit
		if oldunit == nil {
			w.unitsAdded++
		}
		w.unitDeleted[coord] = false
	} else {
		// Root layer: clear any deletion marks
		delete(w.unitDeleted, coord)
	}

	// Remove old unit from player's unit list if replacing
	if oldunit != nil {
		oldPlayerID := int(oldunit.Player)
		if oldPlayerID < len(w.unitsByPlayer) && w.unitsByPlayer[oldPlayerID] != nil {
			for i, u := range w.unitsByPlayer[oldPlayerID] {
				if u == oldunit {
					// Remove old unit from slice
					w.unitsByPlayer[oldPlayerID] = append(w.unitsByPlayer[oldPlayerID][:i], w.unitsByPlayer[oldPlayerID][i+1:]...)
					break
				}
			}
		}
	}

	w.unitsByPlayer[playerID] = append(w.unitsByPlayer[playerID], unit)
	w.unitsByCoord[coord] = unit

	// Now give this unit a unique ID
	return
}

// RemoveUnit removes a unit from the world
func (w *World) RemoveUnit(unit *v1.Unit) error {
	if unit == nil {
		return fmt.Errorf("unit is nil")
	}

	coord := UnitGetCoord(unit)
	p := int(unit.Player)

	// Update transaction counters
	if w.parent != nil {
		// Transaction layer: check if we're deleting a unit from current layer or parent
		if _, existsInThisLayer := w.unitsByCoord[coord]; existsInThisLayer {
			// Deleting from current layer
			w.unitsAdded--
		} else {
			// Deleting from parent layer
			w.unitsDeleted++
		}
		w.unitDeleted[coord] = true
	}

	delete(w.unitsByCoord, coord)

	// Remove from player's unit list if it exists
	if p < len(w.unitsByPlayer) && w.unitsByPlayer[p] != nil {
		for i, u := range w.unitsByPlayer[p] {
			if u == unit {
				// Remove unit from slice
				w.unitsByPlayer[p] = append(w.unitsByPlayer[p][:i], w.unitsByPlayer[p][i+1:]...)
				break
			}
		}
	}
	return nil
}

// MoveUnit moves a unit to a new position
func (w *World) MoveUnit(unit *v1.Unit, newCoord AxialCoord) error {
	if unit == nil {
		return fmt.Errorf("unit is nil")
	}

	// For transaction layers: ensure copy-on-write semantics
	// If we're in a transaction and the unit comes from parent layer, make a copy
	unitToMove := unit
	if w.parent != nil {
		currentCoord := UnitGetCoord(unit)
		// Check if unit exists in current layer or comes from parent
		if _, existsInCurrentLayer := w.unitsByCoord[currentCoord]; !existsInCurrentLayer {
			// Unit comes from parent layer - make a copy to avoid modifying parent objects
			unitToMove = &v1.Unit{
				Q:               unit.Q,
				R:               unit.R,
				Player:          unit.Player,
				UnitType:        unit.UnitType,
				AvailableHealth: unit.AvailableHealth,
				DistanceLeft:    unit.DistanceLeft,
				TurnCounter:     unit.TurnCounter,
			}
		}
	}

	// Remove unit from current position (handles transaction deletion flags)
	if err := w.RemoveUnit(unit); err != nil {
		return fmt.Errorf("failed to remove unit: %w", err)
	}

	// Update unit position (now safe to modify copy)
	UnitSetCoord(unitToMove, newCoord)

	// Add unit at new position (handles transaction addition flags)
	_, err := w.AddUnit(unitToMove)
	if err != nil {
		return fmt.Errorf("failed to add unit at new position: %w", err)
	}

	return nil
}

// =============================================================================
// World Validation and Utilities
// =============================================================================

// Clone creates a deep copy of the world state (useful for undo/redo systems)
func (w *World) Clone() *World {
	if w == nil {
		return nil
	}

	out := NewWorld(w.Name)
	// Clone map
	for _, tile := range w.tilesByCoord {
		if tile != nil {
			// Create a copy of the proto tile
			clonedTile := &v1.Tile{
				Q:        tile.Q,
				R:        tile.R,
				TileType: tile.TileType,
				Player:   tile.Player,
			}
			out.AddTile(clonedTile)
		}
	}
	for _, unit := range w.unitsByCoord {
		if unit != nil {
			// Create a copy of the proto unit
			clonedUnit := &v1.Unit{
				Q:               unit.Q,
				R:               unit.R,
				Player:          unit.Player,
				UnitType:        unit.UnitType,
				AvailableHealth: unit.AvailableHealth,
				DistanceLeft:    unit.DistanceLeft,
				TurnCounter:     unit.TurnCounter,
			}
			out.AddUnit(clonedUnit)
		}
	}
	return out
}

// =============================================================================
// World Loading Methods
// =============================================================================

// GetAllTiles returns all tiles as a map from cube coordinates to tiles
func (w *World) CopyAllTiles() map[AxialCoord]*v1.Tile {
	// Return a copy to prevent external modification
	result := make(map[AxialCoord]*v1.Tile)
	for coord, tile := range w.tilesByCoord {
		result[coord] = tile
	}
	return result
}

/*
// ViewState represents UI-specific state that doesn't affect game logic.
// This includes visual concerns like selections, highlights, and camera position.
type ViewState struct {
	// Selection and highlighting
	SelectedUnit    *v1.Unit   `json:"selectedUnit"`    // Currently selected unit
	HoveredTile     *v1.Tile   `json:"hoveredTile"`     // Tile under cursor
	MovableTiles    []Position `json:"movableTiles"`    // Highlighted movement tiles
	AttackableTiles []Position `json:"attackableTiles"` // Highlighted attack tiles

	// Visual settings
	ShowGrid        bool `json:"showGrid"`        // Whether to show hex grid lines
	ShowCoordinates bool `json:"showCoordinates"` // Whether to show coordinate labels
	ShowPaths       bool `json:"showPaths"`       // Whether to show movement paths

	// Camera and viewport
	CameraX   float64 `json:"cameraX"`   // Camera X position
	CameraY   float64 `json:"cameraY"`   // Camera Y position
	ZoomLevel float64 `json:"zoomLevel"` // Zoom level (1.0 = normal)

	// Editor-specific state
	BrushTerrain int `json:"brushTerrain"` // Current terrain type for painting
	BrushSize    int `json:"brushSize"`    // Brush radius (0 = single hex)
}

// NewViewState creates a new view state with default settings
func NewViewState() *ViewState {
	return &ViewState{
		SelectedUnit:    nil,
		HoveredTile:     nil,
		MovableTiles:    make([]Position, 0),
		AttackableTiles: make([]Position, 0),
		ShowGrid:        true,
		ShowCoordinates: false,
		ShowPaths:       true,
		CameraX:         0.0,
		CameraY:         0.0,
		ZoomLevel:       1.0,
		BrushTerrain:    1, // Default to grass
		BrushSize:       0, // Single hex brush
	}
}
*/

// =============================================================================
// JSON Serialization Methods
// =============================================================================

// MarshalJSON implements custom JSON marshaling for World
func (w *World) MarshalJSON() ([]byte, error) {
	// Convert cube map to tile list for JSON
	out := map[string]any{
		"Name":  w.Name,
		"Tiles": slices.Collect(maps.Values(w.tilesByCoord)),
		"Units": slices.Collect(maps.Values(w.unitsByCoord)),
	}
	return json.Marshal(out)
}

// UnmarshalJSON implements custom JSON unmarshaling for privateMap
func (w *World) UnmarshalJSON(data []byte) error {
	// First try to unmarshal with new bounds format
	type mapJSON struct {
		Name  string
		Tiles []*v1.Tile
		Units []*v1.Unit
	}

	var dict mapJSON

	if err := json.Unmarshal(data, &dict); err != nil {
		return err
	}

	w.Name = dict.Name
	// w.PlayerCount = dict.PlayerCount
	for _, tile := range dict.Tiles {
		w.AddTile(tile)
	}

	for _, unit := range dict.Units {
		w.AddUnit(unit)
	}
	w.boundsChanged = true
	return nil
}

// =============================================================================
// ViewState Management
// =============================================================================

// ClearSelection clears the current unit selection and highlights
func (vs *ViewState) ClearSelection() {
	vs.SelectedUnit = nil
	vs.MovableTiles = make([]Position, 0)
	vs.AttackableTiles = make([]Position, 0)
}

// SetSelection sets the selected unit and updates related highlights
func (vs *ViewState) SetSelection(unit *v1.Unit, movableTiles, attackableTiles []Position) {
	vs.SelectedUnit = unit
	vs.MovableTiles = movableTiles
	vs.AttackableTiles = attackableTiles
}

// SetCamera updates the camera position and zoom
func (vs *ViewState) SetCamera(x, y, zoom float64) {
	vs.CameraX = x
	vs.CameraY = y
	vs.ZoomLevel = zoom
}

// SetBrush updates the brush settings for terrain editing
func (vs *ViewState) SetBrush(terrainType, brushSize int) {
	vs.BrushTerrain = terrainType
	vs.BrushSize = brushSize
}

// ViewState represents UI-specific state that doesn't affect game logic.
// This includes visual concerns like selections, highlights, and camera position.
type ViewState struct {
	// Selection and highlighting
	SelectedUnit    *v1.Unit   `json:"selectedUnit"`    // Currently selected unit
	HoveredTile     *v1.Tile   `json:"hoveredTile"`     // Tile under cursor
	MovableTiles    []Position `json:"movableTiles"`    // Highlighted movement tiles
	AttackableTiles []Position `json:"attackableTiles"` // Highlighted attack tiles

	// Visual settings
	ShowGrid        bool `json:"showGrid"`        // Whether to show hex grid lines
	ShowCoordinates bool `json:"showCoordinates"` // Whether to show coordinate labels
	ShowPaths       bool `json:"showPaths"`       // Whether to show movement paths

	// Camera and viewport
	CameraX   float64 `json:"cameraX"`   // Camera X position
	CameraY   float64 `json:"cameraY"`   // Camera Y position
	ZoomLevel float64 `json:"zoomLevel"` // Zoom level (1.0 = normal)

	// Editor-specific state
	BrushTerrain int `json:"brushTerrain"` // Current terrain type for painting
	BrushSize    int `json:"brushSize"`    // Brush radius (0 = single hex)
}

// NewViewState creates a new view state with default settings
func NewViewState() *ViewState {
	return &ViewState{
		SelectedUnit:    nil,
		HoveredTile:     nil,
		MovableTiles:    make([]Position, 0),
		AttackableTiles: make([]Position, 0),
		ShowGrid:        true,
		ShowCoordinates: false,
		ShowPaths:       true,
		CameraX:         0.0,
		CameraY:         0.0,
		ZoomLevel:       1.0,
		BrushTerrain:    1, // Default to grass
		BrushSize:       0, // Single hex brush
	}
}

// CenterXYForTile converts cube coordinates directly to pixel center x,y coordinates for rendering
// Uses odd-r layout (odd rows offset) as our fixed, consistent layout
// Based on formulas from redblobgames.com for pointy-topped hexagons
func (m *World) CenterXYForTile(coord AxialCoord, tileWidth, tileHeight, yIncrement float64) (x, y float64) {
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
// Uses the privateMap's normalized OriginX/OriginY for proper coordinate translation
// Based on formulas from redblobgames.com for pointy-topped hexagons with odd-r layout
func (m *World) XYToQR(x, y, tileWidth, tileHeight, yIncrement float64) (coord AxialCoord) {
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
	} else { // given we can have non "equal" side length hexagons, easier to do this by converting to row/col first
		row := int((y + tileHeight/2) / yIncrement)

		halfDists := int(1 + math.Abs(x*2/tileWidth))
		if (row & 1) != 0 {
			halfDists = int(1 + math.Abs((x-tileWidth/2)*2/tileWidth))
		}
		col := halfDists / 2
		if x < 0 {
			col = -col
		}
		// col := int((x + tileWidth/2) / tileWidth)
		coord = RowColToHex(row, col)
	}
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

// getprivateMapBounds calculates the pixel bounds of the entire map
// TODO - cache this and only update when bounds changed beyond min/max Q/R
func (m *World) GetWorldBounds(tileWidth, tileHeight, yIncrement float64) WorldBounds {
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

		for coord := range m.tilesByCoord {
			// Use origin at (0,0) for bounds calculation
			x, y := m.CenterXYForTile(coord, tileWidth, tileHeight, yIncrement)

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

		m.lastWorldBounds.MinX = minX
		m.lastWorldBounds.MinY = minY
		m.lastWorldBounds.MaxX = maxX
		m.lastWorldBounds.MaxY = maxY
		m.lastWorldBounds.MinQ = minQ
		m.lastWorldBounds.MinR = minR
		m.lastWorldBounds.MaxQ = maxQ
		m.lastWorldBounds.MaxR = maxR
		m.lastWorldBounds.StartingX = startingX
		m.lastWorldBounds.MinXCoord = minXCoord
		m.lastWorldBounds.MinYCoord = minYCoord
		m.lastWorldBounds.MaxXCoord = maxXCoord
		m.lastWorldBounds.MaxYCoord = maxYCoord
		m.lastWorldBounds.StartingCoord = startingCoord
	}
	return m.lastWorldBounds
}
