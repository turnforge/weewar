package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	"github.com/turnforge/weewar/lib"
)

// assertCmd represents the assert command
var assertCmd = &cobra.Command{
	Use:   "assert",
	Short: "Assert game state conditions",
	Long: `Assert conditions about game state for testing and validation.

Syntax:
  # Unit assertions (by shortcut, Q/R, or row/col)
  ww assert unit A1 [player==1, health>=5]
  ww assert unit 0,-1 [progression_step==2]
  ww assert unit r4,5 [health>=5]

  # Tile assertions
  ww assert tile H1 [player==1, tile_type==6]
  ww assert tile 0,-1 [player==2]

  # Player assertions
  ww assert player 1 [coins>=100, unit_count==3]

  # Game assertions
  ww assert game [turn==5, current_player==2, status==1]

  # Exists checks
  ww assert exists unit A1 A2 B3
  ww assert notexists unit B3

  # Set/capture values (use = without value)
  ww assert unit A1 [health=, distance_left=]

  # Options checks (verify available actions)
  ww assert options unit A1 [attack B3, move 0,5]
  ww assert options unit A1 [attacks B1 B2 B3]  # can attack one of
  ww assert options tile H1 [build trooper, build tank]
  ww assert options unit A1 [capture L]         # capture tile at direction

Operators:
  =     Set (capture current value, always passes)
  ==    Equals (or: eq)
  !=    Not equals (or: ne)
  >     Greater than (or: gt)
  >=    Greater or equal (or: gte)
  <     Less than (or: lt)
  <=    Less or equal (or: lte)
  in    Value in set: health in (5,8,10)
  notin Value not in set

Note: Use text operators (lt, lte, gt, gte, eq, ne) to avoid shell escaping issues.

Exit codes:
  0     All assertions passed
  1     One or more assertions failed`,
	RunE: runAssert,
}

func init() {
	rootCmd.AddCommand(assertCmd)
}

// Operator represents a comparison operator
type Operator int

const (
	OpSet   Operator = iota // = (set/capture value)
	OpEq                    // ==
	OpNe                    // !=
	OpGt                    // >
	OpGe                    // >=
	OpLt                    // <
	OpLe                    // <=
	OpIn                    // in (a,b,c)
	OpNotIn                 // notin (a,b,c)
)

func (o Operator) String() string {
	switch o {
	case OpSet:
		return "="
	case OpEq:
		return "=="
	case OpNe:
		return "!="
	case OpGt:
		return ">"
	case OpGe:
		return ">="
	case OpLt:
		return "<"
	case OpLe:
		return "<="
	case OpIn:
		return "in"
	case OpNotIn:
		return "notin"
	default:
		return "?"
	}
}

// Assertion represents a single assertion
type Assertion struct {
	Field    string
	Operator Operator
	Value    string   // For single value operators
	Values   []string // For in/notin operators
}

// OptionAssertion represents an assertion about available options
// Syntax: "attack B3" (singular) or "attacks B1 B2 B3" (plural = one of)
type OptionAssertion struct {
	OptionType string   // attack, move, build, capture, retreat
	Targets    []string // Target positions/units/unit-types
	IsPlural   bool     // True if using plural form (attacks, moves, etc.) - means "one of"
}

// Valid option types (singular -> plural mapping for parsing)
var optionTypePlurals = map[string]string{
	"attacks":  "attack",
	"moves":    "move",
	"builds":   "build",
	"captures": "capture",
	"retreats": "retreat",
}

// AssertionResult holds the result of evaluating an assertion
type AssertionResult struct {
	EntityType string // unit, tile, player, game
	EntityID   string // A1, 0,-1, 1, etc.
	Field      string
	Operator   Operator
	Expected   string
	Actual     string
	Passed     bool
	IsSet      bool
}

func (r AssertionResult) String() string {
	prefix := "PASS"
	if !r.Passed {
		prefix = "FAIL"
	}
	if r.IsSet {
		prefix = "SET"
	}

	entityStr := r.EntityType
	if r.EntityID != "" {
		entityStr = fmt.Sprintf("%s.%s", r.EntityType, r.EntityID)
	}

	// For exists/notexists, no field is specified
	if r.Field == "" {
		return fmt.Sprintf("%s - %s %s", prefix, entityStr, r.Actual)
	}

	if r.IsSet {
		return fmt.Sprintf("%s - %s.%s = %s", prefix, entityStr, r.Field, r.Actual)
	}

	if r.Operator == OpIn || r.Operator == OpNotIn {
		return fmt.Sprintf("%s - %s.%s %s (%s) (actual: %s)", prefix, entityStr, r.Field, r.Operator, r.Expected, r.Actual)
	}

	// Show actual value when it differs from expected (for comparisons) or on failure
	if r.Expected != r.Actual || !r.Passed {
		return fmt.Sprintf("%s - %s.%s %s %s (actual: %s)", prefix, entityStr, r.Field, r.Operator, r.Expected, r.Actual)
	}
	return fmt.Sprintf("%s - %s.%s %s %s", prefix, entityStr, r.Field, r.Operator, r.Expected)
}

func runAssert(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no assertions provided")
	}

	// Get game ID
	gameID, err := getGameID()
	if err != nil {
		return err
	}

	// Create presenter
	pc, err := createPresenter(gameID)
	if err != nil {
		return err
	}

	if pc.GameState.State == nil {
		return fmt.Errorf("game state not initialized")
	}

	// Parse and evaluate assertions
	results, err := parseAndEvaluate(args, pc)
	if err != nil {
		return err
	}

	// Print results
	passed := 0
	failed := 0
	for _, r := range results {
		fmt.Println(r.String())
		if r.IsSet || r.Passed {
			passed++
		} else {
			failed++
		}
	}

	fmt.Println()
	if failed == 0 {
		fmt.Printf("All %d assertions passed\n", passed)
		return nil
	}
	fmt.Printf("%d of %d assertions failed\n", failed, passed+failed)
	// Return error to trigger non-zero exit code
	return fmt.Errorf("%d assertions failed", failed)
}

func parseAndEvaluate(args []string, pc *PresenterContext) ([]AssertionResult, error) {
	// Join args and re-parse to handle spaces within brackets
	input := strings.Join(args, " ")

	// Check for exists/notexists first
	if strings.HasPrefix(input, "exists ") || strings.HasPrefix(input, "notexists ") {
		return parseExistsAssertions(input, pc)
	}

	// Check for options assertions
	if strings.HasPrefix(input, "options ") {
		return parseOptionsAssertionsFromArgs(args, pc)
	}

	// Parse entity assertions: entity id [assertions]
	// Regex to match: (unit|tile|player|game) (id)? [assertions]
	// The brackets may contain spaces, so we need careful parsing
	return parseEntityAssertions(input, pc)
}

func parseExistsAssertions(input string, pc *PresenterContext) ([]AssertionResult, error) {
	var results []AssertionResult
	expectExists := strings.HasPrefix(input, "exists ")

	// Remove prefix
	if expectExists {
		input = strings.TrimPrefix(input, "exists ")
	} else {
		input = strings.TrimPrefix(input, "notexists ")
	}
	input = strings.TrimSpace(input)

	// Expect: unit A1 A2 B3 or tile H1 0,-1
	parts := strings.Fields(input)
	if len(parts) < 2 {
		return nil, fmt.Errorf("exists requires entity type and at least one identifier")
	}

	entityType := parts[0]
	identifiers := parts[1:]

	for _, id := range identifiers {
		var exists bool
		var err error

		switch entityType {
		case "unit":
			_, exists, err = findUnit(id, pc)
		case "tile":
			_, exists, err = findTile(id, pc)
		default:
			return nil, fmt.Errorf("exists only supports 'unit' and 'tile', got %q", entityType)
		}

		if err != nil {
			return nil, err
		}

		passed := exists == expectExists
		actual := "exists"
		if !exists {
			actual = "does not exist"
		}
		expected := "exists"
		if !expectExists {
			expected = "does not exist"
		}

		results = append(results, AssertionResult{
			EntityType: entityType,
			EntityID:   id,
			Field:      "",
			Operator:   OpEq,
			Expected:   expected,
			Actual:     actual,
			Passed:     passed,
		})
	}

	return results, nil
}

func parseEntityAssertions(input string, pc *PresenterContext) ([]AssertionResult, error) {
	var results []AssertionResult

	// Find all entity blocks: entity id [...] or game [...]
	// Pattern: (unit|tile|player|game) (id)? \[...\]
	re := regexp.MustCompile(`(unit|tile|player|game)\s+([^\[\]]+)?\s*\[([^\]]*)\]`)
	matches := re.FindAllStringSubmatch(input, -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("no valid assertions found in: %s", input)
	}

	for _, match := range matches {
		entityType := match[1]
		entityID := strings.TrimSpace(match[2])
		assertionsStr := match[3]

		// Parse assertions within brackets
		assertions, err := parseAssertions(assertionsStr)
		if err != nil {
			return nil, fmt.Errorf("parsing assertions for %s %s: %w", entityType, entityID, err)
		}

		// Evaluate assertions
		entityResults, err := evaluateAssertions(entityType, entityID, assertions, pc)
		if err != nil {
			return nil, err
		}
		results = append(results, entityResults...)
	}

	return results, nil
}

func parseAssertions(input string) ([]Assertion, error) {
	var assertions []Assertion

	// Split by comma, but be careful with in (a,b,c) syntax
	parts := splitAssertions(input)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		assertion, err := parseAssertion(part)
		if err != nil {
			return nil, err
		}
		assertions = append(assertions, assertion)
	}

	return assertions, nil
}

// splitAssertions splits by comma, but respects parentheses
func splitAssertions(input string) []string {
	var parts []string
	var current strings.Builder
	depth := 0

	for _, ch := range input {
		switch ch {
		case '(':
			depth++
			current.WriteRune(ch)
		case ')':
			depth--
			current.WriteRune(ch)
		case ',':
			if depth == 0 {
				parts = append(parts, current.String())
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

func parseAssertion(input string) (Assertion, error) {
	input = strings.TrimSpace(input)

	// Check for in/notin operators first (they contain spaces)
	if idx := strings.Index(input, " notin "); idx > 0 {
		field := strings.TrimSpace(input[:idx])
		valuesStr := strings.TrimSpace(input[idx+7:])
		values, err := parseValueSet(valuesStr)
		if err != nil {
			return Assertion{}, err
		}
		return Assertion{Field: field, Operator: OpNotIn, Values: values}, nil
	}
	if idx := strings.Index(input, " in "); idx > 0 {
		field := strings.TrimSpace(input[:idx])
		valuesStr := strings.TrimSpace(input[idx+4:])
		values, err := parseValueSet(valuesStr)
		if err != nil {
			return Assertion{}, err
		}
		return Assertion{Field: field, Operator: OpIn, Values: values}, nil
	}

	// Check for text-based comparison operators first (shell-safe alternatives)
	// These use space-delimited format like "health lt 5"
	textOperators := []struct {
		str string
		op  Operator
	}{
		{" lte ", OpLe},
		{" gte ", OpGe},
		{" lt ", OpLt},
		{" gt ", OpGt},
		{" eq ", OpEq},
		{" ne ", OpNe},
	}

	for _, op := range textOperators {
		if idx := strings.Index(input, op.str); idx > 0 {
			field := strings.TrimSpace(input[:idx])
			value := strings.TrimSpace(input[idx+len(op.str):])
			return Assertion{Field: field, Operator: op.op, Value: value}, nil
		}
	}

	// Check for symbol-based comparison operators (order matters: >= before >, etc.)
	operators := []struct {
		str string
		op  Operator
	}{
		{"==", OpEq},
		{"!=", OpNe},
		{">=", OpGe},
		{"<=", OpLe},
		{">", OpGt},
		{"<", OpLt},
		{"=", OpSet},
	}

	for _, op := range operators {
		if idx := strings.Index(input, op.str); idx > 0 {
			field := strings.TrimSpace(input[:idx])
			value := strings.TrimSpace(input[idx+len(op.str):])
			return Assertion{Field: field, Operator: op.op, Value: value}, nil
		}
	}

	return Assertion{}, fmt.Errorf("invalid assertion syntax: %s", input)
}

func parseValueSet(input string) ([]string, error) {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "(") || !strings.HasSuffix(input, ")") {
		return nil, fmt.Errorf("value set must be in parentheses: %s", input)
	}
	inner := input[1 : len(input)-1]
	parts := strings.Split(inner, ",")
	var values []string
	for _, p := range parts {
		values = append(values, strings.TrimSpace(p))
	}
	return values, nil
}

func evaluateAssertions(entityType, entityID string, assertions []Assertion, pc *PresenterContext) ([]AssertionResult, error) {
	var results []AssertionResult

	for _, a := range assertions {
		var result AssertionResult
		var err error

		switch entityType {
		case "unit":
			result, err = evaluateUnitAssertion(entityID, a, pc)
		case "tile":
			result, err = evaluateTileAssertion(entityID, a, pc)
		case "player":
			result, err = evaluatePlayerAssertion(entityID, a, pc)
		case "game":
			result, err = evaluateGameAssertion(a, pc)
		default:
			return nil, fmt.Errorf("unknown entity type: %s", entityType)
		}

		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// parseCoordinate parses a coordinate string in Q,R or rRow,Col format
func parseCoordinate(input string) (lib.AxialCoord, error) {
	input = strings.TrimSpace(input)

	// Check for row/col format (starts with 'r')
	if strings.HasPrefix(strings.ToLower(input), "r") {
		parts := strings.Split(input[1:], ",")
		if len(parts) != 2 {
			return lib.AxialCoord{}, fmt.Errorf("row/col coordinate must have format rRow,Col")
		}
		row, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return lib.AxialCoord{}, fmt.Errorf("invalid row: %s", parts[0])
		}
		col, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return lib.AxialCoord{}, fmt.Errorf("invalid col: %s", parts[1])
		}
		return lib.RowColToHex(row, col, lib.UseEvenRowOffsetCoords), nil
	}

	// Parse Q,R format
	parts := strings.Split(input, ",")
	if len(parts) != 2 {
		return lib.AxialCoord{}, fmt.Errorf("coordinate must have format Q,R or rRow,Col")
	}
	q, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return lib.AxialCoord{}, fmt.Errorf("invalid Q coordinate: %s", parts[0])
	}
	r, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return lib.AxialCoord{}, fmt.Errorf("invalid R coordinate: %s", parts[1])
	}
	return lib.AxialCoord{Q: q, R: r}, nil
}

func findUnit(id string, pc *PresenterContext) (*v1.Unit, bool, error) {
	if pc.GameState.State.WorldData == nil {
		return nil, false, nil
	}

	// Try shortcut lookup first
	for _, unit := range pc.GameState.State.WorldData.UnitsMap {
		if unit != nil && unit.Shortcut == id {
			return unit, true, nil
		}
	}

	// Try to parse as coordinate
	coord, err := parseCoordinate(id)
	if err == nil {
		key := lib.CoordKey(int32(coord.Q), int32(coord.R))
		unit := pc.GameState.State.WorldData.UnitsMap[key]
		return unit, unit != nil, nil
	}

	return nil, false, nil
}

func findTile(id string, pc *PresenterContext) (*v1.Tile, bool, error) {
	if pc.GameState.State.WorldData == nil {
		return nil, false, nil
	}

	// Try shortcut lookup first
	for _, tile := range pc.GameState.State.WorldData.TilesMap {
		if tile != nil && tile.Shortcut == id {
			return tile, true, nil
		}
	}

	// Try to parse as coordinate
	coord, err := parseCoordinate(id)
	if err == nil {
		key := lib.CoordKey(int32(coord.Q), int32(coord.R))
		tile := pc.GameState.State.WorldData.TilesMap[key]
		return tile, tile != nil, nil
	}

	return nil, false, nil
}

func evaluateUnitAssertion(id string, a Assertion, pc *PresenterContext) (AssertionResult, error) {
	unit, exists, err := findUnit(id, pc)
	if err != nil {
		return AssertionResult{}, err
	}
	if !exists {
		return AssertionResult{}, fmt.Errorf("unit %s not found", id)
	}

	// Get field value
	actual, err := getUnitFieldValue(unit, a.Field)
	if err != nil {
		return AssertionResult{}, err
	}

	return evaluateComparison("unit", id, a, actual)
}

func getUnitFieldValue(unit *v1.Unit, field string) (string, error) {
	switch field {
	case "player":
		return fmt.Sprintf("%d", unit.Player), nil
	case "unit_type", "type":
		return fmt.Sprintf("%d", unit.UnitType), nil
	case "health", "available_health":
		return fmt.Sprintf("%d", unit.AvailableHealth), nil
	case "distance_left", "moves":
		return fmt.Sprintf("%.0f", unit.DistanceLeft), nil
	case "progression_step", "step":
		return fmt.Sprintf("%d", unit.ProgressionStep), nil
	case "chosen_alternative":
		return unit.ChosenAlternative, nil
	case "q":
		return fmt.Sprintf("%d", unit.Q), nil
	case "r":
		return fmt.Sprintf("%d", unit.R), nil
	case "shortcut":
		return unit.Shortcut, nil
	default:
		return "", fmt.Errorf("unknown unit field: %s", field)
	}
}

func evaluateTileAssertion(id string, a Assertion, pc *PresenterContext) (AssertionResult, error) {
	tile, exists, err := findTile(id, pc)
	if err != nil {
		return AssertionResult{}, err
	}
	if !exists {
		return AssertionResult{}, fmt.Errorf("tile %s not found", id)
	}

	// Get field value
	actual, err := getTileFieldValue(tile, a.Field)
	if err != nil {
		return AssertionResult{}, err
	}

	return evaluateComparison("tile", id, a, actual)
}

func getTileFieldValue(tile *v1.Tile, field string) (string, error) {
	switch field {
	case "player":
		return fmt.Sprintf("%d", tile.Player), nil
	case "tile_type", "type":
		return fmt.Sprintf("%d", tile.TileType), nil
	case "q":
		return fmt.Sprintf("%d", tile.Q), nil
	case "r":
		return fmt.Sprintf("%d", tile.R), nil
	case "shortcut":
		return tile.Shortcut, nil
	default:
		return "", fmt.Errorf("unknown tile field: %s", field)
	}
}

func evaluatePlayerAssertion(id string, a Assertion, pc *PresenterContext) (AssertionResult, error) {
	playerID, err := strconv.Atoi(id)
	if err != nil {
		return AssertionResult{}, fmt.Errorf("invalid player ID: %s", id)
	}

	// Get field value
	actual, err := getPlayerFieldValue(int32(playerID), a.Field, pc)
	if err != nil {
		return AssertionResult{}, err
	}

	return evaluateComparison("player", id, a, actual)
}

func getPlayerFieldValue(playerID int32, field string, pc *PresenterContext) (string, error) {
	switch field {
	case "coins":
		if ps := pc.GameState.State.PlayerStates[playerID]; ps != nil {
			return fmt.Sprintf("%d", ps.Coins), nil
		}
		return "0", nil
	case "unit_count":
		count := 0
		if pc.GameState.State.WorldData != nil {
			for _, unit := range pc.GameState.State.WorldData.UnitsMap {
				if unit != nil && unit.Player == playerID {
					count++
				}
			}
		}
		return fmt.Sprintf("%d", count), nil
	case "tile_count":
		count := 0
		if pc.GameState.State.WorldData != nil {
			for _, tile := range pc.GameState.State.WorldData.TilesMap {
				if tile != nil && tile.Player == playerID {
					count++
				}
			}
		}
		return fmt.Sprintf("%d", count), nil
	case "is_active":
		if pc.GameState.Game != nil && pc.GameState.Game.Config != nil {
			for _, p := range pc.GameState.Game.Config.Players {
				if p.PlayerId == playerID {
					return fmt.Sprintf("%t", p.IsActive), nil
				}
			}
		}
		return "false", nil
	default:
		return "", fmt.Errorf("unknown player field: %s", field)
	}
}

func evaluateGameAssertion(a Assertion, pc *PresenterContext) (AssertionResult, error) {
	// Get field value
	actual, err := getGameFieldValue(a.Field, pc)
	if err != nil {
		return AssertionResult{}, err
	}

	return evaluateComparison("game", "", a, actual)
}

func getGameFieldValue(field string, pc *PresenterContext) (string, error) {
	state := pc.GameState.State
	switch field {
	case "turn", "turn_counter":
		return fmt.Sprintf("%d", state.TurnCounter), nil
	case "current_player", "player":
		return fmt.Sprintf("%d", state.CurrentPlayer), nil
	case "status":
		return fmt.Sprintf("%d", int32(state.Status)), nil
	case "finished":
		return fmt.Sprintf("%t", state.Finished), nil
	case "winning_player":
		return fmt.Sprintf("%d", state.WinningPlayer), nil
	case "winning_team":
		return fmt.Sprintf("%d", state.WinningTeam), nil
	default:
		return "", fmt.Errorf("unknown game field: %s", field)
	}
}

func evaluateComparison(entityType, entityID string, a Assertion, actual string) (AssertionResult, error) {
	result := AssertionResult{
		EntityType: entityType,
		EntityID:   entityID,
		Field:      a.Field,
		Operator:   a.Operator,
		Actual:     actual,
	}

	// Set operator - capture current value
	if a.Operator == OpSet {
		result.IsSet = true
		result.Passed = true
		result.Expected = actual
		return result, nil
	}

	// Set expected value(s)
	if a.Operator == OpIn || a.Operator == OpNotIn {
		result.Expected = strings.Join(a.Values, ",")
	} else {
		result.Expected = a.Value
	}

	// Evaluate based on operator
	switch a.Operator {
	case OpEq:
		result.Passed = actual == a.Value
	case OpNe:
		result.Passed = actual != a.Value
	case OpGt, OpGe, OpLt, OpLe:
		passed, err := compareNumeric(actual, a.Value, a.Operator)
		if err != nil {
			return AssertionResult{}, err
		}
		result.Passed = passed
	case OpIn:
		result.Passed = contains(a.Values, actual)
	case OpNotIn:
		result.Passed = !contains(a.Values, actual)
	}

	return result, nil
}

func compareNumeric(actual, expected string, op Operator) (bool, error) {
	// Try as float first (handles both int and float)
	actualF, err := strconv.ParseFloat(actual, 64)
	if err != nil {
		return false, fmt.Errorf("cannot compare %q as number", actual)
	}
	expectedF, err := strconv.ParseFloat(expected, 64)
	if err != nil {
		return false, fmt.Errorf("cannot compare %q as number", expected)
	}

	switch op {
	case OpGt:
		return actualF > expectedF, nil
	case OpGe:
		return actualF >= expectedF, nil
	case OpLt:
		return actualF < expectedF, nil
	case OpLe:
		return actualF <= expectedF, nil
	default:
		return false, fmt.Errorf("unexpected operator for numeric comparison: %v", op)
	}
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

// =============================================================================
// Options Assertions
// =============================================================================

// parseOptionsAssertionsFromArgs parses args like ["options", "unit", "A1", "attack B3", "move 0,5"]
// Note: Shell removes quotes, so "attack B3" becomes a single arg without quotes
func parseOptionsAssertionsFromArgs(args []string, pc *PresenterContext) ([]AssertionResult, error) {
	// args[0] = "options"
	// args[1] = entity type (unit/tile)
	// args[2] = entity id
	// args[3:] = option assertions (each was a quoted string on command line)
	if len(args) < 4 {
		return nil, fmt.Errorf("options requires: options unit/tile id \"assertion\" ...")
	}

	entityType := args[1]
	if entityType != "unit" && entityType != "tile" {
		return nil, fmt.Errorf("options requires 'unit' or 'tile', got %q", entityType)
	}
	entityID := args[2]

	// Each remaining arg is an option assertion (shell already unquoted them)
	var optionAssertions []OptionAssertion
	for _, arg := range args[3:] {
		oa, err := parseOptionAssertion(arg)
		if err != nil {
			return nil, fmt.Errorf("parsing option assertion %q: %w", arg, err)
		}
		optionAssertions = append(optionAssertions, oa)
	}

	// Get actual options for the entity
	actualOptions, err := getOptionsForEntity(entityType, entityID, pc)
	if err != nil {
		return nil, err
	}

	// Evaluate each option assertion
	var results []AssertionResult
	for _, oa := range optionAssertions {
		result := evaluateOptionAssertion(entityType, entityID, oa, actualOptions, pc)
		results = append(results, result)
	}

	return results, nil
}

// parseOptionAssertions splits the bracket contents into individual option assertions
func parseOptionAssertions(input string) ([]OptionAssertion, error) {
	var assertions []OptionAssertion

	// Split by comma, respecting quotes
	parts := splitQuotedAssertions(input)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		oa, err := parseOptionAssertion(part)
		if err != nil {
			return nil, err
		}
		assertions = append(assertions, oa)
	}

	return assertions, nil
}

// splitQuotedAssertions splits by comma but respects quoted strings
func splitQuotedAssertions(input string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false

	for _, ch := range input {
		switch ch {
		case '"':
			inQuote = !inQuote
			current.WriteRune(ch)
		case ',':
			if inQuote {
				current.WriteRune(ch)
			} else {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(ch)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

// extractQuotedStrings extracts all quoted strings from input
func extractQuotedStrings(input string) []string {
	var results []string
	re := regexp.MustCompile(`"([^"]*)"`)
	matches := re.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		if len(match) > 1 {
			results = append(results, match[1])
		}
	}
	return results
}

// parseOptionAssertion parses a single quoted option like "attack B3" or "attacks B1 B2"
func parseOptionAssertion(input string) (OptionAssertion, error) {
	input = strings.TrimSpace(input)

	// Remove surrounding quotes
	if strings.HasPrefix(input, `"`) && strings.HasSuffix(input, `"`) {
		input = input[1 : len(input)-1]
	}

	// Split into words
	parts := strings.Fields(input)
	if len(parts) < 2 {
		return OptionAssertion{}, fmt.Errorf("option assertion must have type and target: %s", input)
	}

	verb := strings.ToLower(parts[0])
	targets := parts[1:]

	// Check if plural form
	isPlural := false
	optionType := verb

	if singularForm, ok := optionTypePlurals[verb]; ok {
		optionType = singularForm
		isPlural = true
	}

	// Validate option type
	validTypes := map[string]bool{"attack": true, "move": true, "build": true, "capture": true, "retreat": true}
	if !validTypes[optionType] {
		return OptionAssertion{}, fmt.Errorf("invalid option type: %s", verb)
	}

	return OptionAssertion{
		OptionType: optionType,
		Targets:    targets,
		IsPlural:   isPlural,
	}, nil
}

// getOptionsForEntity fetches available options for a unit or tile
func getOptionsForEntity(entityType, entityID string, pc *PresenterContext) (*v1.GetOptionsAtResponse, error) {
	// Find the coordinate
	var coord lib.AxialCoord
	var err error

	switch entityType {
	case "unit":
		unit, exists, findErr := findUnit(entityID, pc)
		if findErr != nil {
			return nil, findErr
		}
		if !exists {
			return nil, fmt.Errorf("unit %s not found", entityID)
		}
		coord = lib.AxialCoord{Q: int(unit.Q), R: int(unit.R)}
	case "tile":
		tile, exists, findErr := findTile(entityID, pc)
		if findErr != nil {
			return nil, findErr
		}
		if !exists {
			return nil, fmt.Errorf("tile %s not found", entityID)
		}
		coord = lib.AxialCoord{Q: int(tile.Q), R: int(tile.R)}
	default:
		return nil, fmt.Errorf("options only supports 'unit' and 'tile', got %q", entityType)
	}

	// Simulate click to get options
	ctx := context.Background()
	_, err = pc.Presenter.SceneClicked(ctx, &v1.SceneClickedRequest{
		GameId: pc.GameID,
		Q:      int32(coord.Q),
		R:      int32(coord.R),
		Layer:  "base-map",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}

	return pc.TurnOptions.Options, nil
}

// evaluateOptionAssertion checks if the option assertion is satisfied by actual options
func evaluateOptionAssertion(entityType, entityID string, oa OptionAssertion, options *v1.GetOptionsAtResponse, pc *PresenterContext) AssertionResult {
	result := AssertionResult{
		EntityType: entityType,
		EntityID:   entityID,
		Field:      oa.OptionType,
	}

	// Build description
	if oa.IsPlural {
		result.Expected = fmt.Sprintf("%ss %s", oa.OptionType, strings.Join(oa.Targets, " "))
	} else {
		result.Expected = fmt.Sprintf("%s %s", oa.OptionType, oa.Targets[0])
	}

	if options == nil {
		result.Actual = "no options"
		result.Passed = false
		return result
	}

	// Check based on option type
	switch oa.OptionType {
	case "attack":
		result.Passed, result.Actual = checkAttackOptions(oa, options, pc)
	case "move", "retreat":
		result.Passed, result.Actual = checkMoveOptions(oa, options, pc)
	case "build":
		result.Passed, result.Actual = checkBuildOptions(oa, options, pc)
	case "capture":
		result.Passed, result.Actual = checkCaptureOptions(oa, options, pc)
	}

	return result
}

func checkAttackOptions(oa OptionAssertion, options *v1.GetOptionsAtResponse, pc *PresenterContext) (bool, string) {
	// Collect all attack targets from options
	var attackTargets []string
	for _, opt := range options.Options {
		if attack, ok := opt.OptionType.(*v1.GameOption_Attack); ok {
			key := lib.CoordKey(attack.Attack.DefenderQ, attack.Attack.DefenderR)
			attackTargets = append(attackTargets, key)

			// Also check by unit shortcut if there's a unit there
			if unit := pc.GameState.State.WorldData.UnitsMap[key]; unit != nil && unit.Shortcut != "" {
				attackTargets = append(attackTargets, unit.Shortcut)
			}
		}
	}

	return matchTargets(oa, attackTargets, pc)
}

func checkMoveOptions(oa OptionAssertion, options *v1.GetOptionsAtResponse, pc *PresenterContext) (bool, string) {
	// Collect all move targets from options
	var moveTargets []string
	for _, opt := range options.Options {
		if move, ok := opt.OptionType.(*v1.GameOption_Move); ok {
			key := lib.CoordKey(move.Move.ToQ, move.Move.ToR)
			moveTargets = append(moveTargets, key)
		}
	}

	return matchTargets(oa, moveTargets, pc)
}

func checkBuildOptions(oa OptionAssertion, options *v1.GetOptionsAtResponse, pc *PresenterContext) (bool, string) {
	// Collect all build unit types from options
	var buildTypes []string
	for _, opt := range options.Options {
		if build, ok := opt.OptionType.(*v1.GameOption_Build); ok {
			// Add unit type as string
			buildTypes = append(buildTypes, fmt.Sprintf("%d", build.Build.UnitType))

			// Also add unit name if we can resolve it
			rulesEngine := &lib.RulesEngine{RulesEngine: pc.Presenter.RulesEngine}
			if unitDef, err := rulesEngine.GetUnitData(build.Build.UnitType); err == nil {
				buildTypes = append(buildTypes, strings.ToLower(unitDef.Name))
			}
		}
	}

	return matchTargets(oa, buildTypes, pc)
}

func checkCaptureOptions(oa OptionAssertion, options *v1.GetOptionsAtResponse, pc *PresenterContext) (bool, string) {
	// Collect all capture targets from options
	var captureTargets []string
	for _, opt := range options.Options {
		if capture, ok := opt.OptionType.(*v1.GameOption_Capture); ok {
			key := lib.CoordKey(capture.Capture.Q, capture.Capture.R)
			captureTargets = append(captureTargets, key)
		}
	}

	return matchTargets(oa, captureTargets, pc)
}

// matchTargets checks if the assertion targets match actual targets
func matchTargets(oa OptionAssertion, actualTargets []string, pc *PresenterContext) (bool, string) {
	if len(actualTargets) == 0 {
		return false, "none available"
	}

	actualStr := strings.Join(actualTargets, ", ")

	if oa.IsPlural {
		// Plural: need at least one target to match
		for _, target := range oa.Targets {
			normalizedTarget := normalizeTarget(target, pc)
			for _, actual := range actualTargets {
				if strings.EqualFold(normalizedTarget, actual) || strings.EqualFold(target, actual) {
					return true, fmt.Sprintf("found %s in [%s]", target, actualStr)
				}
			}
		}
		return false, fmt.Sprintf("none of targets in [%s]", actualStr)
	}

	// Singular: the exact target must exist
	target := oa.Targets[0]
	normalizedTarget := normalizeTarget(target, pc)
	for _, actual := range actualTargets {
		if strings.EqualFold(normalizedTarget, actual) || strings.EqualFold(target, actual) {
			return true, fmt.Sprintf("found in [%s]", actualStr)
		}
	}
	return false, fmt.Sprintf("not in [%s]", actualStr)
}

// normalizeTarget converts a target to coordinate key format if possible
func normalizeTarget(target string, pc *PresenterContext) string {
	// Try to parse as coordinate
	coord, err := parseCoordinate(target)
	if err == nil {
		return lib.CoordKey(int32(coord.Q), int32(coord.R))
	}

	// Try relative directions (for capture assertions)
	// TODO: implement direction-based targets when baseCoord is available

	return target
}
