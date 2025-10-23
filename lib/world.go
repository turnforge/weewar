package weewar

import (
	"encoding/json"
	"fmt"
	"iter"
	"maps"
	"slices"
	"strconv"

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

	// Unit shortcut tracking (A1, B12, C3, etc.)
	unitsByShortcut      map[string]*v1.Unit `json:"-"` // Quick lookup by shortcut
	unitCountersByPlayer map[int32]int32     `json:"-"` // Next unit number for each player

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
func NewWorld(name string, protoWorld *v1.WorldData) *World {
	w := &World{
		Name:                 name,
		tilesByCoord:         map[AxialCoord]*v1.Tile{},
		unitsByCoord:         map[AxialCoord]*v1.Unit{},
		unitsByShortcut:      map[string]*v1.Unit{},
		unitCountersByPlayer: map[int32]int32{},
		tileDeleted:          map[AxialCoord]bool{},
		unitDeleted:          map[AxialCoord]bool{},
	}

	// Convert protobuf tiles to runtime tiles
	if protoWorld != nil {
		for _, protoTile := range protoWorld.Tiles {
			coord := AxialCoord{Q: int(protoTile.Q), R: int(protoTile.R)}
			w.SetTileType(coord, int(protoTile.TileType))
		}

		// Convert protobuf units to runtime units
		// First pass: track existing shortcuts and find max counters
		for _, protoUnit := range protoWorld.Units {
			if protoUnit.Shortcut != "" {
				// Parse existing shortcut to update counter
				if len(protoUnit.Shortcut) >= 2 {
					playerLetter := protoUnit.Shortcut[0]
					if playerLetter >= 'A' && playerLetter <= 'Z' {
						if num, err := strconv.Atoi(protoUnit.Shortcut[1:]); err == nil {
							playerID := int32(playerLetter - 'A')
							if current, ok := w.unitCountersByPlayer[playerID]; !ok || int32(num) >= current {
								w.unitCountersByPlayer[playerID] = int32(num + 1)
							}
						}
					}
				}
			}
		}

		// Second pass: add units (AddUnit will generate shortcuts for those without)
		for _, protoUnit := range protoWorld.Units {
			coord := AxialCoord{Q: int(protoUnit.Q), R: int(protoUnit.R)}
			fmt.Printf("NewWorld: Converting unit at (%d, %d), saved DistanceLeft=%d, AvailableHealth=%d\n",
				coord.Q, coord.R, protoUnit.DistanceLeft, protoUnit.AvailableHealth)
			unit := &v1.Unit{
				UnitType:        protoUnit.UnitType,
				Q:               int32(coord.Q),
				R:               int32(coord.R),
				Player:          protoUnit.Player,
				Shortcut:        protoUnit.Shortcut, // Preserve existing shortcut
				AvailableHealth: protoUnit.AvailableHealth,
				DistanceLeft:    protoUnit.DistanceLeft, // Preserve saved movement points
				TurnCounter:     protoUnit.TurnCounter,
			}
			w.AddUnit(unit)
		}
	}

	return w
}

func (w *World) Push() *World {
	out := NewWorld(w.Name, nil)
	out.parent = w
	// Inherit unit counters from parent
	for playerID, counter := range w.unitCountersByPlayer {
		out.unitCountersByPlayer[playerID] = counter
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

// GenerateUnitShortcut creates a new shortcut for a unit of the given player
func (w *World) GenerateUnitShortcut(playerID int32) string {
	if playerID < 0 || playerID > 25 {
		return "" // Only support A-Z for now
	}

	// Get next counter for this player
	counter := w.unitCountersByPlayer[playerID]
	w.unitCountersByPlayer[playerID] = counter + 1

	// Generate shortcut: A1, B12, etc.
	playerLetter := string(rune('A' + playerID))
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
	w.unitsByCoord[coord] = unit

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

	out := NewWorld(w.Name, nil)
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
