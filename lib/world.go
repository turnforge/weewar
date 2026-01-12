package lib

import (
	"encoding/json"
	"fmt"
	"iter"
	"maps"
	"slices"
	"strconv"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// =============================================================================
// World - Pure Game State Container
// =============================================================================

type WorldBounds struct {
	MinX, MinY, MaxX, MaxY int
	MinQ, MinR, MaxQ, MaxR int
	Width                  int // MaxX - MinX
	Height                 int // MaxY - MinY
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

	// Proto WorldData - the actual storage for tiles, units, and crossings
	// This is the source of truth for spatial data
	data *v1.WorldData

	// Ways to identify various kinds of units by player
	unitsByPlayer [][]*v1.Unit `json:"-"` // All units in the game world by player ID

	// Unit shortcut tracking (A1, B12, C3, etc.)
	unitsByShortcut      map[string]*v1.Unit `json:"-"` // Quick lookup by shortcut
	unitCountersByPlayer map[int32]int32     `json:"-"` // Next unit number for each player

	tilesByShortcut      map[string]*v1.Tile `json:"-"` // Quick lookup by shortcut
	tileCountersByPlayer map[int32]int32     `json:"-"` // Next tile number for each player

	// In case we are pushed environment this will tell us
	// if a unit was "deleted" in this layer so not to recurse
	// up when looking up a missing unit
	unitDeleted map[string]bool `json:"-"`
	// Same as above but for tiles
	tileDeleted map[string]bool `json:"-"`

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
func NewWorld(name string, protoWorld *v1.WorldData) *World {
	w := &World{
		Name:                 name,
		unitsByShortcut:      map[string]*v1.Unit{},
		unitCountersByPlayer: map[int32]int32{},
		tilesByShortcut:      map[string]*v1.Tile{},
		tileCountersByPlayer: map[int32]int32{},
		tileDeleted:          map[string]bool{},
		unitDeleted:          map[string]bool{},
	}

	// Use the provided WorldData or create a new one
	if protoWorld != nil {
		// Migrate if needed (converts lists to maps, extracts crossings)
		MigrateWorldData(protoWorld)
		w.data = protoWorld
	} else {
		w.data = &v1.WorldData{
			TilesMap:  make(map[string]*v1.Tile),
			UnitsMap:  make(map[string]*v1.Unit),
			Crossings: make(map[string]*v1.Crossing),
		}
	}

	// Build supplementary indexes from the proto data
	w.buildIndexes()

	return w
}

// buildIndexes builds the shortcut and player indexes from proto data
func (w *World) buildIndexes() {
	// First pass: track existing tile shortcuts and find max counters
	for _, tile := range w.data.TilesMap {
		if tile.Player > 0 && tile.Shortcut != "" {
			w.tilesByShortcut[tile.Shortcut] = tile
			// Parse existing shortcut to update counter
			if len(tile.Shortcut) >= 2 {
				playerLetter := tile.Shortcut[0]
				if playerLetter >= 'A' && playerLetter <= 'Z' {
					if num, err := strconv.Atoi(tile.Shortcut[1:]); err == nil {
						playerID := int32(playerLetter - 'A' + 1)
						if current, ok := w.tileCountersByPlayer[playerID]; !ok || int32(num) >= current {
							w.tileCountersByPlayer[playerID] = int32(num + 1)
						}
					}
				}
			}
		}
	}

	// Second pass: generate shortcuts for tiles without them
	for _, tile := range w.data.TilesMap {
		if tile.Player > 0 && tile.Shortcut == "" {
			tile.Shortcut = w.GenerateTileShortcut(tile.Player)
			w.tilesByShortcut[tile.Shortcut] = tile
		}
	}

	// First pass: track existing unit shortcuts and find max counters
	for _, unit := range w.data.UnitsMap {
		if unit.Shortcut != "" {
			w.unitsByShortcut[unit.Shortcut] = unit
			// Parse existing shortcut to update counter
			if len(unit.Shortcut) >= 2 {
				playerLetter := unit.Shortcut[0]
				if playerLetter >= 'A' && playerLetter <= 'Z' {
					if num, err := strconv.Atoi(unit.Shortcut[1:]); err == nil {
						playerID := int32(playerLetter - 'A' + 1)
						if current, ok := w.unitCountersByPlayer[playerID]; !ok || int32(num) >= current {
							w.unitCountersByPlayer[playerID] = int32(num + 1)
						}
					}
				}
			}
		}
	}

	// Second pass: generate shortcuts for units without them and build player index
	for _, unit := range w.data.UnitsMap {
		if unit.Player > 0 && unit.Shortcut == "" {
			unit.Shortcut = w.GenerateUnitShortcut(unit.Player)
			w.unitsByShortcut[unit.Shortcut] = unit
		}

		// Build unitsByPlayer index
		playerID := int(unit.Player)
		for playerID >= len(w.unitsByPlayer) {
			w.unitsByPlayer = append(w.unitsByPlayer, nil)
		}
		w.unitsByPlayer[playerID] = append(w.unitsByPlayer[playerID], unit)
	}
}

// WorldData returns the underlying proto WorldData
func (w *World) WorldData() *v1.WorldData {
	return w.data
}

func (w *World) Push() *World {
	// Create a new WorldData for the transaction layer
	childData := &v1.WorldData{
		TilesMap:  make(map[string]*v1.Tile),
		UnitsMap:  make(map[string]*v1.Unit),
		Crossings: make(map[string]*v1.Crossing),
	}

	out := &World{
		Name:                 w.Name,
		parent:               w,
		data:                 childData,
		unitsByShortcut:      map[string]*v1.Unit{},
		unitCountersByPlayer: map[int32]int32{},
		tilesByShortcut:      map[string]*v1.Tile{},
		tileCountersByPlayer: map[int32]int32{},
		tileDeleted:          map[string]bool{},
		unitDeleted:          map[string]bool{},
	}

	// Inherit unit counters from parent
	for playerID, counter := range w.unitCountersByPlayer {
		out.unitCountersByPlayer[playerID] = counter
	}
	for playerID, counter := range w.tileCountersByPlayer {
		out.tileCountersByPlayer[playerID] = counter
	}
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
		return w.parent.PlayerCount()
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
		seen := make(map[string]bool)

		// First iterate current layer (child overrides parent)
		for key, tile := range w.data.TilesMap {
			seen[key] = true
			coord, _ := ParseCoordKey(key)
			if !yield(coord, tile) {
				return
			}
		}

		// Then iterate parent layers for unseen coordinates
		if w.parent != nil {
			for coord, tile := range w.parent.TilesByCoord() {
				key := CoordKeyFromAxial(coord)
				// Skip if already seen in child or explicitly deleted in child
				if seen[key] || w.tileDeleted[key] {
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
	return int32(len(w.data.UnitsMap))
}

// GenerateUnitShortcut creates a new shortcut for a unit of the given player
func (w *World) GenerateUnitShortcut(playerID int32) string {
	if playerID <= 0 || playerID > 26 {
		return "" // Only support players 1-26 (A-Z), player 0 is neutral (no shortcut)
	}

	// Get next counter for this player
	counter := w.unitCountersByPlayer[playerID]
	w.unitCountersByPlayer[playerID] = counter + 1

	// Generate shortcut: A1, B12, etc.
	// Player 1 -> 'A', Player 2 -> 'B', etc.
	playerLetter := string(rune('A' + playerID - 1))
	return fmt.Sprintf("%s%d", playerLetter, counter+1)
}

// GetUnitByShortcut returns a unit by its shortcut (e.g., "A1", "B12")
func (w *World) GetUnitByShortcut(shortcut string) *v1.Unit {
	// Check current layer first
	if unit, ok := w.unitsByShortcut[shortcut]; ok {
		return unit
	}

	// Check parent layer if exists
	if w.parent != nil {
		return w.parent.GetUnitByShortcut(shortcut)
	}

	return nil
}

// GenerateTileShortcut creates a new shortcut for a tile of the given player
func (w *World) GenerateTileShortcut(playerID int32) string {
	if playerID <= 0 || playerID > 26 {
		return "" // Only support players 1-26 (A-Z), player 0 is neutral (no shortcut)
	}

	// Get next counter for this player
	counter := w.tileCountersByPlayer[playerID]
	w.tileCountersByPlayer[playerID] = counter + 1

	// Generate shortcut: A1, B12, etc.
	// Player 1 -> 'A', Player 2 -> 'B', etc.
	playerLetter := string(rune('A' + playerID - 1))
	return fmt.Sprintf("%s%d", playerLetter, counter+1)
}

// GetTileByShortcut returns a tile by its shortcut (e.g., "A1", "B12")
func (w *World) GetTileByShortcut(shortcut string) *v1.Tile {
	// Check current layer first
	if tile, ok := w.tilesByShortcut[shortcut]; ok {
		return tile
	}

	// Check parent layer if exists
	if w.parent != nil {
		return w.parent.GetTileByShortcut(shortcut)
	}

	return nil
}

func (w *World) UnitsByCoord() iter.Seq2[AxialCoord, *v1.Unit] {
	// Merged iteration: child units override parent units, respect deletions
	return func(yield func(AxialCoord, *v1.Unit) bool) {
		seen := make(map[string]bool)

		// First iterate current layer (child overrides parent)
		for key, unit := range w.data.UnitsMap {
			seen[key] = true
			coord, _ := ParseCoordKey(key)
			if !yield(coord, unit) {
				return
			}
		}

		// Then iterate parent layers for unseen coordinates
		if w.parent != nil {
			for coord, unit := range w.parent.UnitsByCoord() {
				key := CoordKeyFromAxial(coord)
				// Skip if already seen in child or explicitly deleted in child
				if seen[key] || w.unitDeleted[key] {
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
	key := CoordKeyFromAxial(coord)
	out = w.data.UnitsMap[key]
	if out == nil && w.parent != nil && !w.unitDeleted[key] {
		out = w.parent.UnitAt(coord)
	}
	return
}

// TileAt returns the tile at the specified cube coordinates
func (w *World) TileAt(coord AxialCoord) (out *v1.Tile) {
	key := CoordKeyFromAxial(coord)
	out = w.data.TilesMap[key]
	if out == nil && w.parent != nil && !w.tileDeleted[key] {
		out = w.parent.TileAt(coord)
	}
	return
}

// GetPlayerUnits returns all units belonging to the specified player
func (w *World) GetPlayerUnits(playerID int) []*v1.Unit {
	// Check current layer first
	if playerID >= 0 && playerID < len(w.unitsByPlayer) && w.unitsByPlayer[playerID] != nil {
		return w.unitsByPlayer[playerID]
	}

	// Fall back to parent layer if available
	if w.parent != nil {
		return w.parent.GetPlayerUnits(playerID)
	}

	// No units found in any layer
	return nil
}

// =============================================================================
// Crossing (Road/Bridge) Access Methods
// =============================================================================

// CrossingAt returns the crossing at the specified coordinate
func (w *World) CrossingAt(coord AxialCoord) *v1.Crossing {
	key := CoordKeyFromAxial(coord)
	if crossing, ok := w.data.Crossings[key]; ok {
		return crossing
	}
	if w.parent != nil {
		return w.parent.CrossingAt(coord)
	}
	return nil
}

// CrossingTypeAt returns the crossing type at the specified coordinate
func (w *World) CrossingTypeAt(coord AxialCoord) v1.CrossingType {
	crossing := w.CrossingAt(coord)
	if crossing != nil {
		return crossing.Type
	}
	return v1.CrossingType_CROSSING_TYPE_UNSPECIFIED
}

// HasCrossing checks if there's any crossing at the given coordinate
func (w *World) HasCrossing(coord AxialCoord) bool {
	return w.CrossingAt(coord) != nil
}

// HasRoad checks if there's a road at the given coordinate
func (w *World) HasRoad(coord AxialCoord) bool {
	return w.CrossingTypeAt(coord) == v1.CrossingType_CROSSING_TYPE_ROAD
}

// HasBridge checks if there's a bridge at the given coordinate
func (w *World) HasBridge(coord AxialCoord) bool {
	return w.CrossingTypeAt(coord) == v1.CrossingType_CROSSING_TYPE_BRIDGE
}

// SetCrossing sets or removes a crossing at the given coordinate
func (w *World) SetCrossing(coord AxialCoord, crossing *v1.Crossing) {
	key := CoordKeyFromAxial(coord)
	if crossing == nil || crossing.Type == v1.CrossingType_CROSSING_TYPE_UNSPECIFIED {
		delete(w.data.Crossings, key)
	} else {
		w.data.Crossings[key] = crossing
	}
}

// SetCrossingType sets or removes a crossing with just the type (no connectivity info)
func (w *World) SetCrossingType(coord AxialCoord, crossingType v1.CrossingType) {
	if crossingType == v1.CrossingType_CROSSING_TYPE_UNSPECIFIED {
		w.SetCrossing(coord, nil)
	} else {
		w.SetCrossing(coord, &v1.Crossing{Type: crossingType, ConnectsTo: make([]bool, 6)})
	}
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
	key := CoordKeyFromAxial(coord)
	q, r := coord.Q, coord.R
	if q < w.minQ || q > w.maxQ || r < w.minR || r > w.maxR {
		w.boundsChanged = true
	}
	w.tileDeleted[key] = false
	w.data.TilesMap[key] = tile

	// Generate shortcut if not already set and tile is player-owned
	if tile.Player > 0 && tile.Shortcut == "" {
		tile.Shortcut = w.GenerateTileShortcut(tile.Player)
	}

	// Add to shortcut map (only for player-owned tiles with shortcuts)
	if tile.Player > 0 && tile.Shortcut != "" {
		w.tilesByShortcut[tile.Shortcut] = tile
	}
}

// DeleteTile removes the tile at the specified cube coordinate
func (w *World) DeleteTile(coord AxialCoord) {
	tile := w.TileAt(coord)
	if tile != nil {
		key := CoordKeyFromAxial(coord)
		w.tileDeleted[key] = true
		delete(w.data.TilesMap, key)

		// Remove from shortcut map
		if tile.Shortcut != "" {
			delete(w.tilesByShortcut, tile.Shortcut)
		}
	}
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
	key := CoordKeyFromAxial(coord)
	oldunit = w.UnitAt(coord)

	// Update transaction counters
	if w.parent != nil {
		// Transaction layer: track if this is a new unit or replacing a parent unit
		if oldunit == nil {
			w.unitsAdded++
		}
		w.unitDeleted[key] = false
	} else {
		// Root layer: clear any deletion marks
		delete(w.unitDeleted, key)
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
		// Remove old unit from shortcut map
		if oldunit.Shortcut != "" {
			delete(w.unitsByShortcut, oldunit.Shortcut)
		}
	}

	// Generate shortcut if not already set
	if unit.Shortcut == "" {
		unit.Shortcut = w.GenerateUnitShortcut(unit.Player)
	}

	// Add to shortcut map
	if unit.Shortcut != "" {
		w.unitsByShortcut[unit.Shortcut] = unit
	}

	w.unitsByPlayer[playerID] = append(w.unitsByPlayer[playerID], unit)
	w.data.UnitsMap[key] = unit

	return
}

// RemoveUnit removes a unit from the world
func (w *World) RemoveUnit(unit *v1.Unit) error {
	if unit == nil {
		return fmt.Errorf("unit is nil")
	}

	coord := UnitGetCoord(unit)
	key := CoordKeyFromAxial(coord)
	p := int(unit.Player)

	// Update transaction counters
	if w.parent != nil {
		// Transaction layer: check if we're deleting a unit from current layer or parent
		if _, existsInThisLayer := w.data.UnitsMap[key]; existsInThisLayer {
			// Deleting from current layer
			w.unitsAdded--
		} else {
			// Deleting from parent layer
			w.unitsDeleted++
		}
		w.unitDeleted[key] = true
	}

	delete(w.data.UnitsMap, key)

	// Remove from shortcut map
	if unit.Shortcut != "" {
		delete(w.unitsByShortcut, unit.Shortcut)
	}

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
		currentKey := CoordKeyFromAxial(currentCoord)
		// Check if unit exists in current layer or comes from parent
		if _, existsInCurrentLayer := w.data.UnitsMap[currentKey]; !existsInCurrentLayer {
			// Unit comes from parent layer - make a copy to avoid modifying parent objects
			unitToMove = &v1.Unit{
				Q:                unit.Q,
				R:                unit.R,
				Player:           unit.Player,
				UnitType:         unit.UnitType,
				AvailableHealth:  unit.AvailableHealth,
				DistanceLeft:     unit.DistanceLeft,
				LastActedTurn:    unit.LastActedTurn,
				LastToppedupTurn: unit.LastToppedupTurn,
				Shortcut:         unit.Shortcut,
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

	// Create new WorldData with cloned maps
	clonedData := &v1.WorldData{
		TilesMap:  make(map[string]*v1.Tile),
		UnitsMap:  make(map[string]*v1.Unit),
		Crossings: make(map[string]*v1.Crossing),
	}

	// Clone tiles
	for key, tile := range w.data.TilesMap {
		clonedData.TilesMap[key] = &v1.Tile{
			Q:        tile.Q,
			R:        tile.R,
			TileType: tile.TileType,
			Player:   tile.Player,
			Shortcut: tile.Shortcut,
		}
	}

	// Clone units
	for key, unit := range w.data.UnitsMap {
		clonedData.UnitsMap[key] = &v1.Unit{
			Q:                unit.Q,
			R:                unit.R,
			Player:           unit.Player,
			UnitType:         unit.UnitType,
			AvailableHealth:  unit.AvailableHealth,
			DistanceLeft:     unit.DistanceLeft,
			LastActedTurn:    unit.LastActedTurn,
			LastToppedupTurn: unit.LastToppedupTurn,
			Shortcut:         unit.Shortcut,
		}
	}

	// Clone crossings
	for key, crossing := range w.data.Crossings {
		clonedCrossing := &v1.Crossing{
			Type:       crossing.Type,
			ConnectsTo: make([]bool, 6),
		}
		copy(clonedCrossing.ConnectsTo, crossing.ConnectsTo)
		clonedData.Crossings[key] = clonedCrossing
	}

	return NewWorld(w.Name, clonedData)
}

// =============================================================================
// World Loading Methods
// =============================================================================

// =============================================================================
// JSON Serialization Methods
// =============================================================================

// MarshalJSON implements custom JSON marshaling for World
func (w *World) MarshalJSON() ([]byte, error) {
	// Convert to tile/unit lists for JSON compatibility
	tiles := slices.Collect(maps.Values(w.data.TilesMap))
	units := slices.Collect(maps.Values(w.data.UnitsMap))

	out := map[string]any{
		"Name":  w.Name,
		"Tiles": tiles,
		"Units": units,
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

	// Initialize data if needed
	if w.data == nil {
		w.data = &v1.WorldData{
			TilesMap:  make(map[string]*v1.Tile),
			UnitsMap:  make(map[string]*v1.Unit),
			Crossings: make(map[string]*v1.Crossing),
		}
	}

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
// Helper methods to convert row/col to and from Q/R
// Note all game/map/world methods should be PURELY USING Q/R coords.
// These helpers are only when showing debug info or info to UI to players
// =============================================================================

// NumRows returns the number of rows in the map (calculated from bounds)
func (m *World) NumRows() int {
	if m.minR > m.maxR {
		return 0 // Empty map
	}
	return m.maxR - m.minR + 1
}

// NumCols returns the number of columns in the map (calculated from bounds)
func (m *World) NumCols() int {
	if m.minQ > m.maxQ {
		return 0 // Empty map
	}
	return m.maxQ - m.minQ + 1
}
