package lib

import (
	"fmt"
	"strconv"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// Tile type constants for migration
const (
	TileTypeRoad          = 22
	TileTypeBridgeShallow = 18
	TileTypeBridgeRegular = 17
	TileTypeBridgeDeep    = 19
	TileTypePlains        = 5
	TileTypeWaterShallow  = 14
	TileTypeWaterRegular  = 10
	TileTypeWaterDeep     = 15
)

// MigrateWorldData converts old list-based WorldData to map-based storage.
// It also extracts crossings (roads, bridges) from tile types and ensures shortcuts exist.
// This function is idempotent - calling it multiple times is safe.
func MigrateWorldData(wd *v1.WorldData) {
	if wd == nil {
		return
	}

	// Initialize maps if nil
	if wd.TilesMap == nil {
		wd.TilesMap = make(map[string]*v1.Tile)
	}
	if wd.UnitsMap == nil {
		wd.UnitsMap = make(map[string]*v1.Unit)
	}
	if wd.Crossings == nil {
		wd.Crossings = make(map[string]*v1.Crossing)
	}

	// Extract crossings from tile types
	extractCrossings(wd)
}

// EnsureShortcuts generates shortcuts for tiles and units that don't have them.
// This is idempotent - tiles/units with existing shortcuts are left unchanged.
func EnsureShortcuts(wd *v1.WorldData) {
	if wd == nil {
		return
	}

	// Track counters per player to generate unique shortcuts
	tileCounters := make(map[int32]int32)
	unitCounters := make(map[int32]int32)

	// First pass: find max existing counters for tiles
	for _, tile := range wd.TilesMap {
		if tile.Player > 0 && tile.Shortcut != "" {
			if len(tile.Shortcut) >= 2 {
				playerLetter := tile.Shortcut[0]
				if playerLetter >= 'A' && playerLetter <= 'Z' {
					if num, err := strconv.Atoi(tile.Shortcut[1:]); err == nil {
						playerID := int32(playerLetter - 'A' + 1)
						if current, ok := tileCounters[playerID]; !ok || int32(num) >= current {
							tileCounters[playerID] = int32(num + 1)
						}
					}
				}
			}
		}
	}

	// First pass: find max existing counters for units
	for _, unit := range wd.UnitsMap {
		if unit.Player > 0 && unit.Shortcut != "" {
			if len(unit.Shortcut) >= 2 {
				playerLetter := unit.Shortcut[0]
				if playerLetter >= 'A' && playerLetter <= 'Z' {
					if num, err := strconv.Atoi(unit.Shortcut[1:]); err == nil {
						playerID := int32(playerLetter - 'A' + 1)
						if current, ok := unitCounters[playerID]; !ok || int32(num) >= current {
							unitCounters[playerID] = int32(num + 1)
						}
					}
				}
			}
		}
	}

	// Second pass: generate shortcuts for tiles without them
	for _, tile := range wd.TilesMap {
		if tile.Player > 0 && tile.Shortcut == "" {
			tile.Shortcut = generateShortcut(tile.Player, tileCounters)
		}
	}

	// Second pass: generate shortcuts for units without them
	for _, unit := range wd.UnitsMap {
		if unit.Player > 0 && unit.Shortcut == "" {
			unit.Shortcut = generateShortcut(unit.Player, unitCounters)
		}
	}
}

// generateShortcut generates a shortcut like "A1", "B2" for a player and increments the counter
func generateShortcut(playerID int32, counters map[int32]int32) string {
	if playerID <= 0 || playerID > 26 {
		return ""
	}
	counter := counters[playerID]
	counters[playerID] = counter + 1
	letter := byte('A' + playerID - 1)
	return fmt.Sprintf("%c%d", letter, counter+1)
}

// extractCrossings extracts roads and bridges from tile types into the crossings map
// and updates the tile types to their underlying terrain.
func extractCrossings(wd *v1.WorldData) {
	for key, tile := range wd.TilesMap {
		switch tile.TileType {
		case TileTypeRoad:
			// Road -> Plains with road crossing
			wd.Crossings[key] = &v1.Crossing{Type: v1.CrossingType_CROSSING_TYPE_ROAD, ConnectsTo: make([]bool, 6)}
			tile.TileType = TileTypePlains

		case TileTypeBridgeShallow:
			// Bridge over shallow water
			wd.Crossings[key] = &v1.Crossing{Type: v1.CrossingType_CROSSING_TYPE_BRIDGE, ConnectsTo: make([]bool, 6)}
			tile.TileType = TileTypeWaterShallow

		case TileTypeBridgeRegular:
			// Bridge over regular water
			wd.Crossings[key] = &v1.Crossing{Type: v1.CrossingType_CROSSING_TYPE_BRIDGE, ConnectsTo: make([]bool, 6)}
			tile.TileType = TileTypeWaterRegular

		case TileTypeBridgeDeep:
			// Bridge over deep water
			wd.Crossings[key] = &v1.Crossing{Type: v1.CrossingType_CROSSING_TYPE_BRIDGE, ConnectsTo: make([]bool, 6)}
			tile.TileType = TileTypeWaterDeep
		}
	}
}

// GetCrossingType returns the crossing type at the given coordinates
func GetCrossingType(wd *v1.WorldData, q, r int32) v1.CrossingType {
	if wd == nil || wd.Crossings == nil {
		return v1.CrossingType_CROSSING_TYPE_UNSPECIFIED
	}
	key := CoordKey(q, r)
	if crossing := wd.Crossings[key]; crossing != nil {
		return crossing.Type
	}
	return v1.CrossingType_CROSSING_TYPE_UNSPECIFIED
}

// HasCrossing checks if there's any crossing at the given coordinates
func HasCrossing(wd *v1.WorldData, q, r int32) bool {
	return GetCrossingType(wd, q, r) != v1.CrossingType_CROSSING_TYPE_UNSPECIFIED
}

// HasRoad checks if there's a road at the given coordinates
func HasRoad(wd *v1.WorldData, q, r int32) bool {
	return GetCrossingType(wd, q, r) == v1.CrossingType_CROSSING_TYPE_ROAD
}

// HasBridge checks if there's a bridge at the given coordinates
func HasBridge(wd *v1.WorldData, q, r int32) bool {
	return GetCrossingType(wd, q, r) == v1.CrossingType_CROSSING_TYPE_BRIDGE
}

// GetTileFromMap retrieves a tile from the map-based storage
func GetTileFromMap(wd *v1.WorldData, q, r int32) *v1.Tile {
	if wd == nil || wd.TilesMap == nil {
		return nil
	}
	key := CoordKey(q, r)
	return wd.TilesMap[key]
}

// GetUnitFromMap retrieves a unit from the map-based storage
func GetUnitFromMap(wd *v1.WorldData, q, r int32) *v1.Unit {
	if wd == nil || wd.UnitsMap == nil {
		return nil
	}
	key := CoordKey(q, r)
	return wd.UnitsMap[key]
}

// SetTileInMap adds or updates a tile in the map-based storage
func SetTileInMap(wd *v1.WorldData, tile *v1.Tile) {
	if wd == nil || tile == nil {
		return
	}
	if wd.TilesMap == nil {
		wd.TilesMap = make(map[string]*v1.Tile)
	}
	key := CoordKey(tile.Q, tile.R)
	wd.TilesMap[key] = tile
}

// SetUnitInMap adds or updates a unit in the map-based storage
func SetUnitInMap(wd *v1.WorldData, unit *v1.Unit) {
	if wd == nil || unit == nil {
		return
	}
	if wd.UnitsMap == nil {
		wd.UnitsMap = make(map[string]*v1.Unit)
	}
	key := CoordKey(unit.Q, unit.R)
	wd.UnitsMap[key] = unit
}

// RemoveUnitFromMap removes a unit from the map-based storage
func RemoveUnitFromMap(wd *v1.WorldData, q, r int32) {
	if wd == nil || wd.UnitsMap == nil {
		return
	}
	key := CoordKey(q, r)
	delete(wd.UnitsMap, key)
}

// MoveUnitInMap moves a unit from one position to another in the map
func MoveUnitInMap(wd *v1.WorldData, unit *v1.Unit, toQ, toR int32) {
	if wd == nil || unit == nil {
		return
	}
	// Remove from old position
	RemoveUnitFromMap(wd, unit.Q, unit.R)
	// Update unit coordinates
	unit.Q = toQ
	unit.R = toR
	// Add to new position
	SetUnitInMap(wd, unit)
}

// SetCrossing sets or removes a crossing at the given coordinates
func SetCrossing(wd *v1.WorldData, q, r int32, crossing *v1.Crossing) {
	if wd == nil {
		return
	}
	if wd.Crossings == nil {
		wd.Crossings = make(map[string]*v1.Crossing)
	}
	key := CoordKey(q, r)
	if crossing == nil || crossing.Type == v1.CrossingType_CROSSING_TYPE_UNSPECIFIED {
		delete(wd.Crossings, key)
	} else {
		wd.Crossings[key] = crossing
	}
}

// SetCrossingType sets or removes a crossing with just the type (no connectivity info)
func SetCrossingType(wd *v1.WorldData, q, r int32, crossingType v1.CrossingType) {
	if crossingType == v1.CrossingType_CROSSING_TYPE_UNSPECIFIED {
		SetCrossing(wd, q, r, nil)
	} else {
		SetCrossing(wd, q, r, &v1.Crossing{Type: crossingType, ConnectsTo: make([]bool, 6)})
	}
}
