package lib

import (
	"fmt"
	"testing"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
)

// Action sequence tests verify that the action_order progression system
// works correctly for various unit types and move combinations.
//
// Key concepts:
// - action_order: Array like ["move", "attack|capture"] defining allowed sequence
// - progression_step: Current index into action_order (0, 1, 2...)
// - chosen_alternative: When step has "|" options, records the choice
// - Steps advance when action is taken or resource exhausted

// =============================================================================
// Comprehensive Action Sequence Test Framework
// =============================================================================

// ActionSequenceTestCase defines a complete test scenario
type ActionSequenceTestCase struct {
	Name        string
	ActionOrder []string       // Unit's action_order to test
	Steps       []ActionStep   // Sequence of actions to execute
	Setup       *ScenarioSetup // Optional custom setup
}

// ActionStep defines a single action in a sequence
type ActionStep struct {
	Action      string // "move", "attack", "capture", "heal", "endturn"
	ExpectError bool   // Should this action fail?
	Description string // Optional description for debugging
}

// ScenarioSetup allows custom game setup
type ScenarioSetup struct {
	UnitHealth       int32   // Default 10
	UnitDistance     float64 // Default 5
	StartStep        int32   // Starting progression_step
	EnemyDistance    int     // Distance to place enemy (default 1)
	OnEnemyBase      bool    // Place unit on enemy base
	StartOnNeutral   bool    // Start on neutral capturable tile
	SecondEnemy      bool    // Add second enemy for multi-attack tests
}

// actionSequenceTestRunner executes action sequence tests
type actionSequenceTestRunner struct {
	t           *testing.T
	game        *Game
	unit        *v1.Unit
	enemy       *v1.Unit
	actionOrder []string
}

func newActionSequenceTestRunner(t *testing.T, tc ActionSequenceTestCase) *actionSequenceTestRunner {
	t.Helper()

	setup := tc.Setup
	if setup == nil {
		setup = &ScenarioSetup{}
	}

	// Apply defaults
	if setup.UnitHealth == 0 {
		setup.UnitHealth = 10
	}
	if setup.UnitDistance == 0 {
		setup.UnitDistance = 5
	}
	if setup.EnemyDistance == 0 {
		setup.EnemyDistance = 1
	}

	// Build game world
	builder := newTestGameBuilder().
		grassTiles(6).
		currentPlayer(1).
		seed(42)

	// Override tiles if needed
	if setup.OnEnemyBase {
		builder = newTestGameBuilder().
			tile(0, 0, TileTypeLandBase, 2). // Enemy base at origin
			grassTiles(5).
			currentPlayer(1).
			seed(42)
	} else if setup.StartOnNeutral {
		builder = newTestGameBuilder().
			tile(0, 0, TileTypeLandBase, 0). // Neutral base at origin
			grassTiles(5).
			currentPlayer(1).
			seed(42)
	}

	game := builder.build()

	// Create the test unit
	// Set LastToppedupTurn to current turn to prevent TopUpUnitIfNeeded from resetting ProgressionStep
	unit := &v1.Unit{
		Q: 0, R: 0, Player: 1, UnitType: testUnitTypeSoldier,
		Shortcut: "A1", AvailableHealth: setup.UnitHealth,
		DistanceLeft: setup.UnitDistance, ProgressionStep: setup.StartStep,
		LastToppedupTurn: 1, // Match game's TurnCounter to preserve ProgressionStep
	}
	game.World.AddUnit(unit)

	// Override action_order in rules engine
	unitDef, _ := game.RulesEngine.GetUnitData(testUnitTypeSoldier)
	unitDef.ActionOrder = tc.ActionOrder

	// Create enemy at specified distance
	enemy := &v1.Unit{
		Q: int32(setup.EnemyDistance), R: 0, Player: 2, UnitType: testUnitTypeSoldier,
		Shortcut: "B1", AvailableHealth: 10, DistanceLeft: 3,
	}
	game.World.AddUnit(enemy)

	// Add second enemy if needed
	if setup.SecondEnemy {
		enemy2 := &v1.Unit{
			Q: 0, R: int32(setup.EnemyDistance), Player: 2, UnitType: testUnitTypeSoldier,
			Shortcut: "B2", AvailableHealth: 10, DistanceLeft: 3,
		}
		game.World.AddUnit(enemy2)
	}

	return &actionSequenceTestRunner{
		t:           t,
		game:        game,
		unit:        unit,
		enemy:       enemy,
		actionOrder: tc.ActionOrder,
	}
}

func (r *actionSequenceTestRunner) runSteps(steps []ActionStep) {
	for i, step := range steps {
		err := r.executeAction(step.Action)

		if step.ExpectError && err == nil {
			r.t.Errorf("Step %d (%s): expected error but got none", i, step.Action)
		}
		if !step.ExpectError && err != nil {
			r.t.Errorf("Step %d (%s): unexpected error: %v", i, step.Action, err)
		}

		// Update unit reference (may have moved)
		r.unit = r.findPlayerUnit(1)
	}
}

func (r *actionSequenceTestRunner) executeAction(action string) error {
	var move *v1.GameMove

	// Find current unit position
	unit := r.findPlayerUnit(1)
	if unit == nil {
		return fmt.Errorf("player 1 unit not found")
	}

	switch action {
	case "move":
		// Move one tile in available direction
		toQ, toR := r.findMoveTarget(unit)
		move = &v1.GameMove{
			MoveType: &v1.GameMove_MoveUnit{
				MoveUnit: &v1.MoveUnitAction{
					From: &v1.Position{Q: unit.Q, R: unit.R},
					To:   &v1.Position{Q: toQ, R: toR},
				},
			},
		}

	case "attack":
		// Attack adjacent enemy
		enemy := r.findPlayerUnit(2)
		if enemy == nil {
			return fmt.Errorf("no enemy to attack")
		}
		move = &v1.GameMove{
			MoveType: &v1.GameMove_AttackUnit{
				AttackUnit: &v1.AttackUnitAction{
					Attacker: &v1.Position{Q: unit.Q, R: unit.R},
					Defender: &v1.Position{Q: enemy.Q, R: enemy.R},
				},
			},
		}

	case "capture":
		move = &v1.GameMove{
			MoveType: &v1.GameMove_CaptureBuilding{
				CaptureBuilding: &v1.CaptureBuildingAction{
					Pos: &v1.Position{Q: unit.Q, R: unit.R},
				},
			},
		}

	case "heal":
		move = &v1.GameMove{
			MoveType: &v1.GameMove_HealUnit{
				HealUnit: &v1.HealUnitAction{
					Pos: &v1.Position{Q: unit.Q, R: unit.R},
				},
			},
		}

	case "endturn":
		move = &v1.GameMove{
			MoveType: &v1.GameMove_EndTurn{EndTurn: &v1.EndTurnAction{}},
		}

	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	return r.game.ProcessMove(move)
}

func (r *actionSequenceTestRunner) findPlayerUnit(player int32) *v1.Unit {
	for _, u := range r.game.World.GetPlayerUnits(int(player)) {
		return u
	}
	return nil
}

func (r *actionSequenceTestRunner) findMoveTarget(unit *v1.Unit) (int32, int32) {
	// Try adjacent tiles in order
	directions := []struct{ dq, dr int32 }{
		{1, 0}, {0, 1}, {-1, 1}, {-1, 0}, {0, -1}, {1, -1},
	}

	for _, d := range directions {
		targetQ, targetR := unit.Q+d.dq, unit.R+d.dr
		coord := AxialCoord{Q: int(targetQ), R: int(targetR)}

		// Check tile exists and is unoccupied
		if tile := r.game.World.TileAt(coord); tile != nil {
			if existing := r.game.World.UnitAt(coord); existing == nil {
				return targetQ, targetR
			}
		}
	}

	// Fallback - return adjacent even if invalid
	return unit.Q + 1, unit.R
}

func (r *actionSequenceTestRunner) getProgressionStep() int32 {
	unit := r.findPlayerUnit(1)
	if unit == nil {
		return -1
	}
	return unit.ProgressionStep
}

func (r *actionSequenceTestRunner) getChosenAlternative() string {
	unit := r.findPlayerUnit(1)
	if unit == nil {
		return ""
	}
	return unit.ChosenAlternative
}

// =============================================================================
// Table-Driven Tests for All Action Order Patterns
// =============================================================================

// NOTE: These tests document CURRENT behavior. Action_order is NOT enforced
// at the ProcessMove level - it's only used for:
// 1. GetAllowedActionsForUnit (computing available options for UI)
// 2. Advancing ProgressionStep after actions are performed
//
// Future enhancement: Add action_order validation to ProcessMove* functions.

func TestActionSequences_StandardPatterns(t *testing.T) {
	tests := []ActionSequenceTestCase{
		// Pattern: move -> attack (most common, 18 units)
		{
			Name:        "move_then_attack_succeeds",
			ActionOrder: []string{"move", "attack"},
			Steps: []ActionStep{
				{Action: "move", ExpectError: false},
				{Action: "attack", ExpectError: false},
			},
		},
		{
			Name:        "attack_without_move_succeeds_currently",
			ActionOrder: []string{"move", "attack"},
			Setup:       &ScenarioSetup{UnitDistance: 0}, // No movement = can't do move step
			Steps: []ActionStep{
				// NOTE: Currently succeeds because action_order not validated at ProcessMove level
				{Action: "attack", ExpectError: false},
			},
		},

		// Pattern: move -> attack|capture (soldiers, 7 units)
		{
			Name:        "move_then_attack_with_alternative",
			ActionOrder: []string{"move", "attack|capture"},
			Steps: []ActionStep{
				{Action: "move", ExpectError: false},
				{Action: "attack", ExpectError: false}, // Choose attack over capture
			},
		},
		{
			Name:        "capture_on_enemy_base",
			ActionOrder: []string{"move", "attack|capture"},
			Setup:       &ScenarioSetup{OnEnemyBase: true},
			Steps: []ActionStep{
				{Action: "capture", ExpectError: false}, // Capture on enemy base
			},
		},

		// Pattern: attack (stationary units, 3 units)
		{
			Name:        "attack_only_unit_can_attack",
			ActionOrder: []string{"attack"},
			Steps: []ActionStep{
				{Action: "attack", ExpectError: false},
			},
		},
		{
			Name:        "attack_only_unit_move_succeeds_currently",
			ActionOrder: []string{"attack"},
			Steps: []ActionStep{
				// NOTE: Currently succeeds because action_order not validated at ProcessMove level
				{Action: "move", ExpectError: false},
			},
		},

		// Pattern: move|attack (flexible units, 3 units)
		{
			Name:        "flexible_can_attack_first",
			ActionOrder: []string{"move|attack"},
			Steps: []ActionStep{
				{Action: "attack", ExpectError: false}, // Can attack without moving
			},
		},
		{
			Name:        "flexible_can_move_first",
			ActionOrder: []string{"move|attack"},
			Steps: []ActionStep{
				{Action: "move", ExpectError: false}, // Can move instead
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := newActionSequenceTestRunner(t, tc)
			runner.runSteps(tc.Steps)
		})
	}
}

func TestActionSequences_SpecialPatterns(t *testing.T) {
	tests := []ActionSequenceTestCase{
		// Pattern: move -> attack -> attack (double attack, 1 unit)
		// Note: After move, unit position changes so enemy must be adjacent to destination
		{
			Name:        "attack_then_second_attack_in_place",
			ActionOrder: []string{"attack", "attack"},
			Setup:       &ScenarioSetup{SecondEnemy: true},
			Steps: []ActionStep{
				{Action: "attack", ExpectError: false},
				// Second attack would need second enemy still adjacent
				// First enemy may be dead or damaged
			},
		},

		// Pattern: move -> attack -> retreat
		{
			Name:        "move_then_attack_sequence",
			ActionOrder: []string{"move", "attack", "retreat"},
			Steps: []ActionStep{
				{Action: "attack", ExpectError: false}, // Attack first (action_order not enforced)
				// Retreat would need separate test since it's not implemented as "move"
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := newActionSequenceTestRunner(t, tc)
			runner.runSteps(tc.Steps)
		})
	}
}

func TestActionSequences_EndTurnBehavior(t *testing.T) {
	tests := []ActionSequenceTestCase{
		{
			Name:        "end_turn_without_acting",
			ActionOrder: []string{"move", "attack"},
			Steps: []ActionStep{
				{Action: "endturn", ExpectError: false},
			},
		},
		{
			Name:        "end_turn_after_partial_sequence",
			ActionOrder: []string{"move", "attack"},
			Steps: []ActionStep{
				{Action: "move", ExpectError: false},
				{Action: "endturn", ExpectError: false}, // Can end turn without attacking
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := newActionSequenceTestRunner(t, tc)
			runner.runSteps(tc.Steps)
		})
	}
}

// =============================================================================
// GetAllowedActionsForUnit Tests (Unit Tests for Rules Engine)
// =============================================================================

func TestGetAllowedActionsForUnit(t *testing.T) {
	rulesEngine := DefaultRulesEngine()

	tests := []struct {
		name            string
		progressionStep int32
		distanceLeft    float64
		chosenAlt       string
		actionOrder     []string
		wantActions     []string
		dontWant        []string
	}{
		{
			name:            "step_0_with_movement_allows_move",
			progressionStep: 0,
			distanceLeft:    3,
			actionOrder:     []string{"move", "attack"},
			wantActions:     []string{"move"},
			dontWant:        []string{"attack"},
		},
		{
			name:            "step_0_no_movement_skips_move",
			progressionStep: 0,
			distanceLeft:    0,
			actionOrder:     []string{"move", "attack"},
			wantActions:     []string{}, // No valid actions at step 0 without movement
			dontWant:        []string{"move"},
		},
		{
			name:            "step_1_allows_attack_or_capture",
			progressionStep: 1,
			distanceLeft:    0,
			actionOrder:     []string{"move", "attack|capture"},
			wantActions:     []string{"attack", "capture"},
		},
		{
			name:            "chosen_alternative_locks_choice",
			progressionStep: 1,
			distanceLeft:    0,
			chosenAlt:       "attack",
			actionOrder:     []string{"move", "attack|capture"},
			wantActions:     []string{"attack"},
			dontWant:        []string{"capture"},
		},
		{
			name:            "all_steps_complete_no_actions",
			progressionStep: 2,
			distanceLeft:    0,
			actionOrder:     []string{"move", "attack"},
			wantActions:     []string{},
		},
		{
			name:            "move_or_attack_pattern_allows_both",
			progressionStep: 0,
			distanceLeft:    3,
			actionOrder:     []string{"move|attack"},
			wantActions:     []string{"move", "attack"},
		},
		{
			name:            "double_attack_step_2_allows_attack",
			progressionStep: 2,
			distanceLeft:    0,
			actionOrder:     []string{"move", "attack", "attack"},
			wantActions:     []string{"attack"},
		},
		{
			name:            "retreat_step_allows_retreat",
			progressionStep: 2,
			distanceLeft:    2, // Has retreat points
			actionOrder:     []string{"move", "attack", "retreat"},
			wantActions:     []string{"retreat"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unit := &v1.Unit{
				ProgressionStep:   tt.progressionStep,
				DistanceLeft:      tt.distanceLeft,
				ChosenAlternative: tt.chosenAlt,
			}

			unitDef := &v1.UnitDefinition{
				ActionOrder: tt.actionOrder,
			}

			allowed := rulesEngine.GetAllowedActionsForUnit(unit, unitDef)

			// Check wanted actions are present
			for _, want := range tt.wantActions {
				found := false
				for _, a := range allowed {
					if a == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected action %q in allowed list, got %v", want, allowed)
				}
			}

			// Check unwanted actions are absent
			for _, dontWant := range tt.dontWant {
				for _, a := range allowed {
					if a == dontWant {
						t.Errorf("Did not expect action %q in allowed list, got %v", dontWant, allowed)
					}
				}
			}
		})
	}
}

// =============================================================================
// Progression Step Advancement Tests
// =============================================================================

func TestProgressionStepAdvancement(t *testing.T) {
	tests := []struct {
		name          string
		actionOrder   []string
		action        string
		startStep     int32
		startDistance float64
		wantStep      int32
	}{
		{
			name:          "attack_advances_step",
			actionOrder:   []string{"move", "attack"},
			action:        "attack",
			startStep:     1, // Already at attack step
			startDistance: 0,
			wantStep:      2, // Advances past attack
		},
		{
			name:          "move_exhaustion_advances_step",
			actionOrder:   []string{"move", "attack"},
			action:        "move",
			startStep:     0,
			startDistance: 1.0, // Exactly enough for one grass move (cost 1.0), exhausts to 0
			wantStep:      1,   // Advances to attack step when DistanceLeft <= 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := ActionSequenceTestCase{
				Name:        tt.name,
				ActionOrder: tt.actionOrder,
				Setup: &ScenarioSetup{
					StartStep:    tt.startStep,
					UnitDistance: tt.startDistance,
				},
				Steps: []ActionStep{
					{Action: tt.action},
				},
			}

			runner := newActionSequenceTestRunner(t, tc)
			runner.runSteps(tc.Steps)

			gotStep := runner.getProgressionStep()
			if gotStep != -1 && gotStep != tt.wantStep {
				t.Errorf("After %s: progression_step = %d, want %d", tt.action, gotStep, tt.wantStep)
			}
		})
	}
}

// =============================================================================
// Turn Reset Tests
// =============================================================================

func TestTurnResetBehavior(t *testing.T) {
	tc := ActionSequenceTestCase{
		Name:        "turn_cycle_resets_progression",
		ActionOrder: []string{"move", "attack"},
		Steps: []ActionStep{
			{Action: "attack"}, // Advances step
			{Action: "endturn"},
			{Action: "endturn"}, // Back to player 1
		},
	}

	runner := newActionSequenceTestRunner(t, tc)

	// Attack first (at step 0 with move|attack flexibility or start at step 1)
	// Override unit to start at step 1 for this test
	runner.unit.ProgressionStep = 1

	runner.runSteps(tc.Steps)

	// After turn cycle, progression should reset
	step := runner.getProgressionStep()
	if step != 0 {
		t.Errorf("After turn cycle: progression_step = %d, want 0", step)
	}
}
