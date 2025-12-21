package cmd

import (
	"testing"
)

func TestParseAssertion_Equals(t *testing.T) {
	tests := []struct {
		input    string
		field    string
		operator Operator
		value    string
	}{
		{"player==1", "player", OpEq, "1"},
		{"health==10", "health", OpEq, "10"},
		{"distance_left==0", "distance_left", OpEq, "0"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			a, err := parseAssertion(tc.input)
			if err != nil {
				t.Fatalf("parseAssertion(%q) error: %v", tc.input, err)
			}
			if a.Field != tc.field {
				t.Errorf("field = %q, want %q", a.Field, tc.field)
			}
			if a.Operator != tc.operator {
				t.Errorf("operator = %v, want %v", a.Operator, tc.operator)
			}
			if a.Value != tc.value {
				t.Errorf("value = %q, want %q", a.Value, tc.value)
			}
		})
	}
}

func TestParseAssertion_Comparisons(t *testing.T) {
	tests := []struct {
		input    string
		field    string
		operator Operator
		value    string
	}{
		{"health>=5", "health", OpGe, "5"},
		{"health<=10", "health", OpLe, "10"},
		{"health>0", "health", OpGt, "0"},
		{"health<100", "health", OpLt, "100"},
		{"player!=2", "player", OpNe, "2"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			a, err := parseAssertion(tc.input)
			if err != nil {
				t.Fatalf("parseAssertion(%q) error: %v", tc.input, err)
			}
			if a.Field != tc.field {
				t.Errorf("field = %q, want %q", a.Field, tc.field)
			}
			if a.Operator != tc.operator {
				t.Errorf("operator = %v, want %v", a.Operator, tc.operator)
			}
			if a.Value != tc.value {
				t.Errorf("value = %q, want %q", a.Value, tc.value)
			}
		})
	}
}

func TestParseAssertion_TextOperators(t *testing.T) {
	tests := []struct {
		input    string
		field    string
		operator Operator
		value    string
	}{
		{"health gte 5", "health", OpGe, "5"},
		{"health lte 10", "health", OpLe, "10"},
		{"health gt 0", "health", OpGt, "0"},
		{"health lt 100", "health", OpLt, "100"},
		{"player eq 1", "player", OpEq, "1"},
		{"player ne 2", "player", OpNe, "2"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			a, err := parseAssertion(tc.input)
			if err != nil {
				t.Fatalf("parseAssertion(%q) error: %v", tc.input, err)
			}
			if a.Field != tc.field {
				t.Errorf("field = %q, want %q", a.Field, tc.field)
			}
			if a.Operator != tc.operator {
				t.Errorf("operator = %v, want %v", a.Operator, tc.operator)
			}
			if a.Value != tc.value {
				t.Errorf("value = %q, want %q", a.Value, tc.value)
			}
		})
	}
}

func TestParseAssertion_Set(t *testing.T) {
	a, err := parseAssertion("health=")
	if err != nil {
		t.Fatalf("parseAssertion error: %v", err)
	}
	if a.Field != "health" {
		t.Errorf("field = %q, want %q", a.Field, "health")
	}
	if a.Operator != OpSet {
		t.Errorf("operator = %v, want %v", a.Operator, OpSet)
	}
	if a.Value != "" {
		t.Errorf("value = %q, want empty", a.Value)
	}
}

func TestParseAssertion_InOperator(t *testing.T) {
	a, err := parseAssertion("health in (5,8,10)")
	if err != nil {
		t.Fatalf("parseAssertion error: %v", err)
	}
	if a.Field != "health" {
		t.Errorf("field = %q, want %q", a.Field, "health")
	}
	if a.Operator != OpIn {
		t.Errorf("operator = %v, want %v", a.Operator, OpIn)
	}
	if len(a.Values) != 3 {
		t.Errorf("values length = %d, want 3", len(a.Values))
	}
	expected := []string{"5", "8", "10"}
	for i, v := range expected {
		if a.Values[i] != v {
			t.Errorf("values[%d] = %q, want %q", i, a.Values[i], v)
		}
	}
}

func TestParseAssertion_NotInOperator(t *testing.T) {
	a, err := parseAssertion("player notin (1,2)")
	if err != nil {
		t.Fatalf("parseAssertion error: %v", err)
	}
	if a.Field != "player" {
		t.Errorf("field = %q, want %q", a.Field, "player")
	}
	if a.Operator != OpNotIn {
		t.Errorf("operator = %v, want %v", a.Operator, OpNotIn)
	}
	if len(a.Values) != 2 {
		t.Errorf("values length = %d, want 2", len(a.Values))
	}
}

func TestSplitAssertions(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"player==1, health>=5", []string{"player==1", " health>=5"}},
		{"health in (5,8,10), player==1", []string{"health in (5,8,10)", " player==1"}},
		{"a==1, b in (1,2,3), c!=4", []string{"a==1", " b in (1,2,3)", " c!=4"}},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			parts := splitAssertions(tc.input)
			if len(parts) != len(tc.expected) {
				t.Fatalf("got %d parts, want %d: %v", len(parts), len(tc.expected), parts)
			}
			for i, p := range parts {
				if p != tc.expected[i] {
					t.Errorf("part[%d] = %q, want %q", i, p, tc.expected[i])
				}
			}
		})
	}
}

func TestParseAssertions_Multiple(t *testing.T) {
	assertions, err := parseAssertions("player==1, health>=5, distance_left==0")
	if err != nil {
		t.Fatalf("parseAssertions error: %v", err)
	}
	if len(assertions) != 3 {
		t.Fatalf("got %d assertions, want 3", len(assertions))
	}

	// Check first assertion
	if assertions[0].Field != "player" || assertions[0].Operator != OpEq || assertions[0].Value != "1" {
		t.Errorf("assertion[0] = %+v, want player==1", assertions[0])
	}

	// Check second assertion
	if assertions[1].Field != "health" || assertions[1].Operator != OpGe || assertions[1].Value != "5" {
		t.Errorf("assertion[1] = %+v, want health>=5", assertions[1])
	}

	// Check third assertion
	if assertions[2].Field != "distance_left" || assertions[2].Operator != OpEq || assertions[2].Value != "0" {
		t.Errorf("assertion[2] = %+v, want distance_left==0", assertions[2])
	}
}

func TestEvaluateComparison_Equals(t *testing.T) {
	a := Assertion{Field: "health", Operator: OpEq, Value: "10"}

	// Test pass
	result, err := evaluateComparison("unit", "A1", a, "10")
	if err != nil {
		t.Fatalf("evaluateComparison error: %v", err)
	}
	if !result.Passed {
		t.Error("expected pass for 10 == 10")
	}

	// Test fail
	result, err = evaluateComparison("unit", "A1", a, "5")
	if err != nil {
		t.Fatalf("evaluateComparison error: %v", err)
	}
	if result.Passed {
		t.Error("expected fail for 5 == 10")
	}
}

func TestEvaluateComparison_GreaterOrEqual(t *testing.T) {
	a := Assertion{Field: "health", Operator: OpGe, Value: "5"}

	tests := []struct {
		actual string
		pass   bool
	}{
		{"10", true},
		{"5", true},
		{"4", false},
		{"0", false},
	}

	for _, tc := range tests {
		result, err := evaluateComparison("unit", "A1", a, tc.actual)
		if err != nil {
			t.Fatalf("evaluateComparison error: %v", err)
		}
		if result.Passed != tc.pass {
			t.Errorf("health >= 5 with actual %s: got passed=%v, want %v", tc.actual, result.Passed, tc.pass)
		}
	}
}

func TestEvaluateComparison_In(t *testing.T) {
	a := Assertion{Field: "health", Operator: OpIn, Values: []string{"5", "8", "10"}}

	tests := []struct {
		actual string
		pass   bool
	}{
		{"5", true},
		{"8", true},
		{"10", true},
		{"7", false},
		{"0", false},
	}

	for _, tc := range tests {
		result, err := evaluateComparison("unit", "A1", a, tc.actual)
		if err != nil {
			t.Fatalf("evaluateComparison error: %v", err)
		}
		if result.Passed != tc.pass {
			t.Errorf("health in (5,8,10) with actual %s: got passed=%v, want %v", tc.actual, result.Passed, tc.pass)
		}
	}
}

func TestEvaluateComparison_Set(t *testing.T) {
	a := Assertion{Field: "health", Operator: OpSet, Value: ""}

	result, err := evaluateComparison("unit", "A1", a, "10")
	if err != nil {
		t.Fatalf("evaluateComparison error: %v", err)
	}
	if !result.Passed {
		t.Error("set should always pass")
	}
	if !result.IsSet {
		t.Error("expected IsSet to be true")
	}
	if result.Actual != "10" {
		t.Errorf("actual = %q, want %q", result.Actual, "10")
	}
}

func TestAssertionResult_String(t *testing.T) {
	tests := []struct {
		result   AssertionResult
		expected string
	}{
		{
			AssertionResult{EntityType: "unit", EntityID: "A1", Field: "health", Operator: OpEq, Expected: "10", Actual: "10", Passed: true},
			"PASS - unit.A1.health == 10",
		},
		{
			AssertionResult{EntityType: "unit", EntityID: "A1", Field: "health", Operator: OpGe, Expected: "5", Actual: "10", Passed: true},
			"PASS - unit.A1.health >= 5 (actual: 10)",
		},
		{
			AssertionResult{EntityType: "unit", EntityID: "A1", Field: "player", Operator: OpEq, Expected: "1", Actual: "2", Passed: false},
			"FAIL - unit.A1.player == 1 (actual: 2)",
		},
		{
			AssertionResult{EntityType: "game", EntityID: "", Field: "turn", Operator: OpEq, Expected: "5", Actual: "5", Passed: true},
			"PASS - game.turn == 5",
		},
		{
			AssertionResult{EntityType: "unit", EntityID: "A1", Field: "health", Operator: OpSet, Expected: "10", Actual: "10", Passed: true, IsSet: true},
			"SET - unit.A1.health = 10",
		},
		{
			// Exists check (no field)
			AssertionResult{EntityType: "unit", EntityID: "A1", Field: "", Operator: OpEq, Expected: "exists", Actual: "exists", Passed: true},
			"PASS - unit.A1 exists",
		},
		{
			// Not exists check (no field)
			AssertionResult{EntityType: "unit", EntityID: "B99", Field: "", Operator: OpEq, Expected: "does not exist", Actual: "does not exist", Passed: true},
			"PASS - unit.B99 does not exist",
		},
	}

	for _, tc := range tests {
		got := tc.result.String()
		if got != tc.expected {
			t.Errorf("got %q, want %q", got, tc.expected)
		}
	}
}

func TestParseOptionAssertion(t *testing.T) {
	tests := []struct {
		input      string
		optionType string
		targets    []string
		isPlural   bool
	}{
		// Singular - must have exactly this option
		{`"attack B3"`, "attack", []string{"B3"}, false},
		{`"move 0,5"`, "move", []string{"0,5"}, false},
		{`"build trooper"`, "build", []string{"trooper"}, false},
		{`"capture L"`, "capture", []string{"L"}, false},
		{`"retreat 0,5"`, "retreat", []string{"0,5"}, false},
		// Plural - can do one of these targets
		{`"attacks A1 3,2 TR,TL,L r3,4"`, "attack", []string{"A1", "3,2", "TR,TL,L", "r3,4"}, true},
		{`"moves 0,5 1,5"`, "move", []string{"0,5", "1,5"}, true},
		{`"builds trooper tank"`, "build", []string{"trooper", "tank"}, true},
		{`"captures L R"`, "capture", []string{"L", "R"}, true},
		{`"retreats 0,5 1,5"`, "retreat", []string{"0,5", "1,5"}, true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			oa, err := parseOptionAssertion(tc.input)
			if err != nil {
				t.Fatalf("parseOptionAssertion(%q) error: %v", tc.input, err)
			}
			if oa.OptionType != tc.optionType {
				t.Errorf("optionType = %q, want %q", oa.OptionType, tc.optionType)
			}
			if oa.IsPlural != tc.isPlural {
				t.Errorf("isPlural = %v, want %v", oa.IsPlural, tc.isPlural)
			}
			if len(oa.Targets) != len(tc.targets) {
				t.Fatalf("targets length = %d, want %d: %v", len(oa.Targets), len(tc.targets), oa.Targets)
			}
			for i, target := range tc.targets {
				if oa.Targets[i] != target {
					t.Errorf("targets[%d] = %q, want %q", i, oa.Targets[i], target)
				}
			}
		})
	}
}

func TestExtractQuotedStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{`"attack B3" "move 0,5"`, []string{"attack B3", "move 0,5"}},
		{`"attacks B1 B2 B3"`, []string{"attacks B1 B2 B3"}},
		{`"build trooper" "build tank"`, []string{"build trooper", "build tank"}},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := extractQuotedStrings(tc.input)
			if len(result) != len(tc.expected) {
				t.Fatalf("got %d strings, want %d: %v", len(result), len(tc.expected), result)
			}
			for i, s := range tc.expected {
				if result[i] != s {
					t.Errorf("result[%d] = %q, want %q", i, result[i], s)
				}
			}
		})
	}
}

func TestParseCoordinate(t *testing.T) {
	tests := []struct {
		input string
		q     int
		r     int
	}{
		{"0,0", 0, 0},
		{"1,-1", 1, -1},
		{"-2,3", -2, 3},
		{"5,5", 5, 5},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			coord, err := parseCoordinate(tc.input)
			if err != nil {
				t.Fatalf("parseCoordinate(%q) error: %v", tc.input, err)
			}
			if coord.Q != tc.q || coord.R != tc.r {
				t.Errorf("got (%d,%d), want (%d,%d)", coord.Q, coord.R, tc.q, tc.r)
			}
		})
	}
}

func TestParseRowColCoordinate(t *testing.T) {
	// r4,5 should be parsed as row=4, col=5
	coord, err := parseCoordinate("r4,5")
	if err != nil {
		t.Fatalf("parseCoordinate(r4,5) error: %v", err)
	}
	// The coordinate should be converted to Q,R
	// Based on RowColToHex(4, 5) - row=4 is even, so q = 5 - (4-0)/2 = 5 - 2 = 3
	// Actually let me verify: row=4, col=5
	// x = col - (row-(row&1))/2 = 5 - (4-0)/2 = 5 - 2 = 3
	// z = row = 4
	// y = -x - z = -3 - 4 = -7
	// q, r = CubeToAxial(x, y, z) = x, z = 3, 4
	if coord.Q != 3 || coord.R != 4 {
		t.Errorf("r4,5 got (%d,%d), want (3,4)", coord.Q, coord.R)
	}
}
