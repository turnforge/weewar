package tests

import (
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// TestGetAllowedActionsForUnit tests the core progression logic
func TestGetAllowedActionsForUnit(t *testing.T) {
	rulesEngine := DefaultRulesEngine()

	tests := []struct {
		name              string
		actionOrder       []string
		progressionStep   int32
		chosenAlternative string
		distanceLeft      float64
		expected          []string
	}{
		{
			name:              "Step 0 - move allowed",
			actionOrder:       []string{"move", "attack"},
			progressionStep:   0,
			chosenAlternative: "",
			distanceLeft:      3.0,
			expected:          []string{"move"},
		},
		{
			name:              "Step 0 - no movement left",
			actionOrder:       []string{"move", "attack"},
			progressionStep:   0,
			chosenAlternative: "",
			distanceLeft:      0.0,
			expected:          []string{}, // Can't move, no actions available
		},
		{
			name:              "Step 1 - attack allowed",
			actionOrder:       []string{"move", "attack"},
			progressionStep:   1,
			chosenAlternative: "",
			distanceLeft:      0.0,
			expected:          []string{"attack"},
		},
		{
			name:              "Step 2 - all complete",
			actionOrder:       []string{"move", "attack"},
			progressionStep:   2,
			chosenAlternative: "",
			distanceLeft:      0.0,
			expected:          []string{},
		},
		{
			name:              "Pipe-separated - both allowed",
			actionOrder:       []string{"move", "attack|capture"},
			progressionStep:   1,
			chosenAlternative: "",
			distanceLeft:      0.0,
			expected:          []string{"attack", "capture"},
		},
		{
			name:              "Pipe-separated - attack chosen",
			actionOrder:       []string{"move", "attack|capture"},
			progressionStep:   1,
			chosenAlternative: "attack",
			distanceLeft:      0.0,
			expected:          []string{"attack"}, // Only attack, not capture
		},
		{
			name:              "Default action order",
			actionOrder:       []string{}, // Empty uses default
			progressionStep:   0,
			chosenAlternative: "",
			distanceLeft:      3.0,
			expected:          []string{"move"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unit := &v1.Unit{
				ProgressionStep:   tt.progressionStep,
				ChosenAlternative: tt.chosenAlternative,
				DistanceLeft:      tt.distanceLeft,
			}

			unitDef := &v1.UnitDefinition{
				ActionOrder: tt.actionOrder,
			}

			allowed := rulesEngine.GetAllowedActionsForUnit(unit, unitDef)

			if len(allowed) != len(tt.expected) {
				t.Errorf("Expected %d actions, got %d: %v", len(tt.expected), len(allowed), allowed)
				return
			}

			for _, exp := range tt.expected {
				found := false
				for _, act := range allowed {
					if act == exp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected action '%s' not found in %v", exp, allowed)
				}
			}
		})
	}
}

// TestProgressionStepAdvancement tests that progression_step advances correctly
func TestProgressionStepAdvancement(t *testing.T) {
	rulesEngine := DefaultRulesEngine()

	// Create a simple world with tiles
	worldData := &v1.WorldData{
		TilesMap: map[string]*v1.Tile{
			"0,0": {Q: 0, R: 0, TileType: 5}, // Grass
			"1,0": {Q: 1, R: 0, TileType: 5}, // Grass
			"2,0": {Q: 2, R: 0, TileType: 5}, // Grass
		},
		UnitsMap: map[string]*v1.Unit{
			"0,0": {
				Q:                0,
				R:                0,
				Player:           1,
				UnitType:         1, // Soldier
				AvailableHealth:  10,
				DistanceLeft:     1.0, // Only 1 movement point
				ProgressionStep:  0,
				LastToppedupTurn: 1, // Already topped up this turn
			},
		},
	}
	world := NewWorld("test", worldData)

	game := &Game{
		Game: &v1.Game{
			Config: &v1.GameConfiguration{
				Players: []*v1.GamePlayer{
					{PlayerId: 1},
				},
			},
		},
		GameState: &v1.GameState{
			CurrentPlayer: 1,
			TurnCounter:   1,
		},
		World:       world,
		RulesEngine: rulesEngine,
	}

	// Process a move that uses up all movement points
	moveAction := &v1.MoveUnitAction{
		From: &v1.Position{Q: 0, R: 0},
		To:   &v1.Position{Q: 1, R: 0},
	}

	err := game.ProcessMoveUnit(&v1.GameMove{Player: 1}, moveAction, false)
	if err != nil {
		t.Fatalf("Failed to process move: %v", err)
	}

	// Check that unit advanced to step 1 (since distance_left reached 0)
	movedUnit := world.UnitAt(AxialCoord{Q: 1, R: 0})
	if movedUnit == nil {
		t.Fatal("Unit not found after move")
	}

	if movedUnit.ProgressionStep != 1 {
		t.Errorf("Expected progression_step=1, got %d", movedUnit.ProgressionStep)
	}

	if movedUnit.DistanceLeft != 0 {
		t.Errorf("Expected distance_left=0, got %f", movedUnit.DistanceLeft)
	}
}

// TestTopUpResetsProgression tests that TopUpUnitIfNeeded resets progression
func TestTopUpResetsProgression(t *testing.T) {
	rulesEngine := DefaultRulesEngine()

	worldData := &v1.WorldData{
		TilesMap: map[string]*v1.Tile{
			"0,0": {Q: 0, R: 0, TileType: 5},
		},
		UnitsMap: map[string]*v1.Unit{
			"0,0": {
				Q:                 0,
				R:                 0,
				Player:            1,
				UnitType:          1, // Soldier
				AvailableHealth:   10,
				DistanceLeft:      0,
				LastToppedupTurn:  1,
				ProgressionStep:   2,        // At completion
				ChosenAlternative: "attack", // Had chosen attack
			},
		},
	}

	world := NewWorld("test", worldData)

	game := &Game{
		Game: &v1.Game{},
		GameState: &v1.GameState{
			TurnCounter: 2, // New turn
		},
		World:       world,
		RulesEngine: rulesEngine,
	}

	// Get the unit from the world
	unit := world.UnitAt(AxialCoord{Q: 0, R: 0})

	// Top up unit for new turn
	err := game.TopUpUnitIfNeeded(unit)
	if err != nil {
		t.Fatalf("TopUpUnitIfNeeded failed: %v", err)
	}

	// Check that progression was reset
	if unit.ProgressionStep != 0 {
		t.Errorf("Expected progression_step=0 after top-up, got %d", unit.ProgressionStep)
	}

	if unit.ChosenAlternative != "" {
		t.Errorf("Expected chosen_alternative=\"\" after top-up, got %q", unit.ChosenAlternative)
	}

	// Check that movement points were restored
	if unit.DistanceLeft == 0 {
		t.Error("Expected distance_left to be restored after top-up")
	}
}

// TestParseActionAlternatives tests the pipe-separated parsing
func TestParseActionAlternatives(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "attack",
			expected: []string{"attack"},
		},
		{
			input:    "attack|capture",
			expected: []string{"attack", "capture"},
		},
		{
			input:    "attack|capture|build",
			expected: []string{"attack", "capture", "build"},
		},
		{
			input:    "move",
			expected: []string{"move"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseActionAlternatives(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d alternatives, got %d: %v",
					len(tt.expected), len(result), result)
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected alternative[%d]='%s', got '%s'",
						i, expected, result[i])
				}
			}
		})
	}
}
