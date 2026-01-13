package lib

import (
	"fmt"
	"strconv"
	"strings"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

const UseEvenRowOffsetCoords = false

// ParseTarget represents either a unit ID or a coordinate position
type ParseTarget struct {
	IsShortcut  bool       // true if this represents a unit, false if coordinate
	ForceTile   bool       // true if "t:" prefix was used to force tile lookup
	Unit        *v1.Unit   // the unit if IsShortcut is true
	Tile        *v1.Tile   // the tile at this position
	Coordinate  AxialCoord // the coordinate if IsShortcut is false
	Raw         string     // original input string
	RawNoPrefix string     // input string without the "t:" prefix
}

func (p *ParseTarget) Position() *v1.Position {
	return &v1.Position{
		Q:     int32(p.Coordinate.Q),
		R:     int32(p.Coordinate.R),
		Label: p.Raw,
	}
}

// ParsePositionOrUnit parses a string that can be either:
// - Unit ID: A1, B12, C2 (PlayerLetter + UnitNumber)
// - Q/R Coordinate: 3,4 or 5,-2
// - Row/Col Coordinate: r4,5 (prefixed with 'r')
func ParsePositionOrUnit(game *Game, input string) (target *ParseTarget, err error) {
	return ParsePositionOrUnitWithContext(game, input, nil)
}

// ParsePositionOrUnitWithContext parses position formats with optional base coordinate for relative directions
// Supports all formats from ParsePositionOrUnit, plus:
// - Direction: L, R, TL, TR, BL, BR (when baseCoord is provided)
// - Multiple directions: TL,TL,TR (when baseCoord is provided, applies sequentially)
// - Tile prefix: t:A1, t:3,4, t:r4,5 (forces tile lookup instead of unit)
func ParsePositionOrUnitWithContext(game *Game, input string, baseCoord *AxialCoord) (target *ParseTarget, err error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	// Check for "t:" prefix to force tile lookup
	forceTile := false
	originalInput := input
	if strings.HasPrefix(strings.ToLower(input), "t:") {
		forceTile = true
		input = strings.TrimSpace(input[2:]) // Remove "t:" prefix
		if input == "" {
			return nil, fmt.Errorf("empty input after 't:' prefix")
		}
	}

	// Check if it's a direction sequence (only if baseCoord is provided)
	if baseCoord != nil {
		parts := strings.Split(input, ",")
		directions := make([]NeighborDirection, 0, len(parts))
		allDirections := true

		// Try to parse all parts as directions
		for _, part := range parts {
			dir, dirErr := ParseDirection(strings.TrimSpace(part))
			if dirErr != nil {
				allDirections = false
				break
			}
			directions = append(directions, dir)
		}

		// If all parts are valid directions, apply them sequentially
		if allDirections {
			currentCoord := *baseCoord
			for _, dir := range directions {
				currentCoord = currentCoord.Neighbor(dir)
			}
			tile := game.World.TileAt(currentCoord)
			unit := game.World.UnitAt(currentCoord)
			if forceTile {
				unit = nil // Don't return unit when tile is explicitly requested
			}
			return &ParseTarget{
				IsShortcut:  false,
				ForceTile:   forceTile,
				Unit:        unit,
				Tile:        tile,
				Coordinate:  currentCoord,
				Raw:         originalInput,
				RawNoPrefix: input,
			}, nil
		}
		// If not all directions, fall through to coordinate parsing
	}

	// Check if it's a row/col coordinate (starts with 'r')
	if strings.HasPrefix(strings.ToLower(input), "r") {
		target, err = parseRowColCoordinate(input[1:])
	} else if strings.Contains(input, ",") {
		// Check if it's a coordinate (contains comma)
		target, err = parseQRCoordinate(input)
	} else {
		// Try to parse as unit ID or tile shortcut
		if forceTile {
			target, err = parseTileID(game, input)
		} else {
			target, err = parseUnitID(game, input)
		}
	}

	if target == nil {
		return
	}

	// Set the forceTile flag and original input
	target.ForceTile = forceTile
	target.Raw = originalInput
	target.RawNoPrefix = input

	// Get the tile and unit at this coordinate
	target.Tile = game.World.TileAt(target.Coordinate)
	if !target.IsShortcut && !forceTile {
		target.Unit = game.World.UnitAt(target.Coordinate)
	}

	return
}

// parseUnitID parses unit ID format: A1, B12, C2, etc.
func parseUnitID(game *Game, input string) (*ParseTarget, error) {
	input = strings.ToUpper(strings.TrimSpace(input))
	if len(input) < 2 {
		return nil, fmt.Errorf("unit ID too short")
	}

	// Use the shortcut lookup directly
	unit := game.World.GetUnitByShortcut(input)
	if unit == nil {
		return nil, fmt.Errorf("unit %s does not exist", input)
	}

	return &ParseTarget{
		IsShortcut: true,
		Unit:       unit,
		Coordinate: CoordFromInt32(unit.Q, unit.R), // Also provide the coordinate for convenience
	}, nil
}

// parseTileID parses tile shortcut format: A1, B12, C2, etc.
func parseTileID(game *Game, input string) (*ParseTarget, error) {
	input = strings.ToUpper(strings.TrimSpace(input))
	if len(input) < 2 {
		return nil, fmt.Errorf("tile ID too short")
	}

	// Use the shortcut lookup directly
	tile := game.World.GetTileByShortcut(input)
	if tile == nil {
		return nil, fmt.Errorf("tile %s does not exist", input)
	}

	return &ParseTarget{
		IsShortcut: true, // tile shortcuts use the same format as unit shortcuts
		Tile:       tile,
		Coordinate: CoordFromInt32(tile.Q, tile.R), // Also provide the coordinate for convenience
	}, nil
}

// parseQRCoordinate parses Q,R coordinate format: 3,4 or -2,5
func parseQRCoordinate(input string) (*ParseTarget, error) {
	parts := strings.Split(input, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("coordinate must have exactly 2 parts separated by comma")
	}

	q, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid Q coordinate: %s", parts[0])
	}

	r, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid R coordinate: %s", parts[1])
	}

	coord := AxialCoord{Q: q, R: r}

	return &ParseTarget{
		IsShortcut: false,
		Coordinate: coord,
	}, nil
}

// parseRowColCoordinate parses row/col coordinate format: 4,5 (after 'r' prefix)
func parseRowColCoordinate(input string) (*ParseTarget, error) {
	parts := strings.Split(input, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("row/col coordinate must have exactly 2 parts separated by comma")
	}

	row, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid row coordinate: %s", parts[0])
	}

	col, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid col coordinate: %s", parts[1])
	}

	// Convert row/col to Q/R using the hex coordinate system
	coord := RowColToHex(row, col, UseEvenRowOffsetCoords)

	return &ParseTarget{
		IsShortcut: false,
		Coordinate: coord,
	}, nil
}

// String returns a human-readable representation of the target
func (t *ParseTarget) String() string {
	prefix := ""
	if t.ForceTile {
		prefix = "Tile "
	}
	if t.IsShortcut {
		if t.ForceTile {
			return fmt.Sprintf("%s%s at %s", prefix, t.RawNoPrefix, t.Coordinate.String())
		}
		return fmt.Sprintf("Unit %s at %s", t.RawNoPrefix, t.Coordinate.String())
	}
	return fmt.Sprintf("%sPosition %s", prefix, t.Coordinate.String())
}

// GetCoordinate returns the coordinate for this target (works for both units and positions)
func (t *ParseTarget) GetCoordinate() AxialCoord {
	return t.Coordinate
}

// GetUnit returns the unit if this target represents a unit, nil otherwise
func (t *ParseTarget) GetUnit() *v1.Unit {
	return t.Unit
}

// GetTile returns the tile at this target's position
func (t *ParseTarget) GetTile() *v1.Tile {
	return t.Tile
}
