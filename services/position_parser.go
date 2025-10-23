package services

import (
	"fmt"
	"strconv"
	"strings"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

// ParseTarget represents either a unit ID or a coordinate position
type ParseTarget struct {
	IsUnit     bool       // true if this represents a unit, false if coordinate
	Unit       *v1.Unit   // the unit if IsUnit is true
	Coordinate AxialCoord // the coordinate if IsUnit is false
	Raw        string     // original input string
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
func ParsePositionOrUnitWithContext(game *Game, input string, baseCoord *AxialCoord) (target *ParseTarget, err error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	// Check if it's a direction (only if baseCoord is provided)
	if baseCoord != nil {
		dir, dirErr := ParseDirection(input)
		if dirErr == nil {
			// It's a valid direction, calculate neighbor
			neighborCoord := baseCoord.Neighbor(dir)
			return &ParseTarget{
				IsUnit:     false,
				Unit:       game.World.UnitAt(neighborCoord),
				Coordinate: neighborCoord,
				Raw:        input,
			}, nil
		}
	}

	// Check if it's a row/col coordinate (starts with 'r')
	if strings.HasPrefix(strings.ToLower(input), "r") {
		target, err = parseRowColCoordinate(input[1:])
	} else if strings.Contains(input, ",") {
		// Check if it's a coordinate (contains comma)
		target, err = parseQRCoordinate(input)
	} else {
		// Try to parse as unit ID
		target, err = parseUnitID(game, input)
	}

	if target == nil {
		return
	}

	// Get the unit if you can
	if !target.IsUnit {
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
		IsUnit:     true,
		Unit:       unit,
		Coordinate: CoordFromInt32(unit.Q, unit.R), // Also provide the coordinate for convenience
		Raw:        input,
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
		IsUnit:     false,
		Unit:       nil,
		Coordinate: coord,
		Raw:        input,
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
	coord := RowColToHex(row, col)

	return &ParseTarget{
		IsUnit:     false,
		Unit:       nil,
		Coordinate: coord,
		Raw:        fmt.Sprintf("r%s", input),
	}, nil
}

// String returns a human-readable representation of the target
func (t *ParseTarget) String() string {
	if t.IsUnit {
		return fmt.Sprintf("Unit %s at %s", t.Raw, t.Coordinate.String())
	}
	return fmt.Sprintf("Position %s", t.Coordinate.String())
}

// GetCoordinate returns the coordinate for this target (works for both units and positions)
func (t *ParseTarget) GetCoordinate() AxialCoord {
	return t.Coordinate
}

// GetUnit returns the unit if this target represents a unit, nil otherwise
func (t *ParseTarget) GetUnit() *v1.Unit {
	return t.Unit
}
