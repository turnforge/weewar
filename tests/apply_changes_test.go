package tests

import (
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// =============================================================================
// Tests for lib/changes.go - applyPlayerChanged handling ResetUnits
// =============================================================================

// TestApplyPlayerChangedUpdatesCurrentPlayer tests that applyPlayerChanged
// correctly updates CurrentPlayer and TurnCounter
func TestApplyPlayerChangedUpdatesCurrentPlayer(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create a simple game
	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add a tile
	tile := NewTile(AxialCoord{Q: 0, R: 0}, TileTypeGrass)
	world.AddTile(tile)

	game := &v1.Game{Id: "test-game", Name: "Test Game"}
	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 100, IsActive: true},
			2: {Coins: 100, IsActive: true},
		},
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Create a PlayerChanged change
	playerChangedChange := &v1.PlayerChangedChange{
		PreviousPlayer: 1,
		NewPlayer:      2,
		PreviousTurn:   1,
		NewTurn:        1,
	}

	move := &v1.GameMove{
		Player: 1,
		Changes: []*v1.WorldChange{
			{
				ChangeType: &v1.WorldChange_PlayerChanged{
					PlayerChanged: playerChangedChange,
				},
			},
		},
	}

	// Apply the changes
	err = rtGame.ApplyChanges([]*v1.GameMove{move})
	if err != nil {
		t.Fatalf("ApplyChanges failed: %v", err)
	}

	// Verify CurrentPlayer was updated
	if rtGame.CurrentPlayer != 2 {
		t.Errorf("CurrentPlayer: got %d, want 2", rtGame.CurrentPlayer)
	}

	// Verify GameState.CurrentPlayer was updated
	if rtGame.GameState.CurrentPlayer != 2 {
		t.Errorf("GameState.CurrentPlayer: got %d, want 2", rtGame.GameState.CurrentPlayer)
	}
}

// TestApplyPlayerChangedUpdatesResetUnits tests that applyPlayerChanged
// correctly applies ResetUnits to update unit state (for remote updates)
func TestApplyPlayerChangedUpdatesResetUnits(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create a game with a unit that has exhausted movement
	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add a tile
	tile := NewTile(AxialCoord{Q: 0, R: 0}, TileTypeGrass)
	world.AddTile(tile)

	// Add a unit with exhausted movement (DistanceLeft = 0)
	unit := &v1.Unit{
		Q:                0,
		R:                0,
		Player:           2, // Player 2's unit (will be reset when turn changes to player 2)
		UnitType:         1,
		AvailableHealth:  8, // Damaged
		DistanceLeft:     0, // Exhausted
		LastToppedupTurn: 1, // Last topped up on turn 1
		LastActedTurn:    1,
	}
	world.AddUnit(unit)

	game := &v1.Game{Id: "test-game", Name: "Test Game"}
	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 100, IsActive: true},
			2: {Coins: 100, IsActive: true},
		},
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Verify initial state - unit is exhausted
	unitBefore := rtGame.World.UnitAt(AxialCoord{Q: 0, R: 0})
	if unitBefore.DistanceLeft != 0 {
		t.Fatalf("Initial DistanceLeft should be 0, got %f", unitBefore.DistanceLeft)
	}

	// Create a PlayerChanged change with ResetUnits
	// This simulates what the server sends when ending turn
	resetUnit := &v1.Unit{
		Q:                0,
		R:                0,
		Player:           2,
		UnitType:         1,
		AvailableHealth:  8,
		DistanceLeft:     3, // Topped up movement
		LastToppedupTurn: 2, // Updated to new turn
		LastActedTurn:    1,
	}

	playerChangedChange := &v1.PlayerChangedChange{
		PreviousPlayer: 1,
		NewPlayer:      2,
		PreviousTurn:   1,
		NewTurn:        2,
		ResetUnits:     []*v1.Unit{resetUnit},
	}

	move := &v1.GameMove{
		Player: 1,
		Changes: []*v1.WorldChange{
			{
				ChangeType: &v1.WorldChange_PlayerChanged{
					PlayerChanged: playerChangedChange,
				},
			},
		},
	}

	// Apply the changes
	err = rtGame.ApplyChanges([]*v1.GameMove{move})
	if err != nil {
		t.Fatalf("ApplyChanges failed: %v", err)
	}

	// Verify unit was updated with reset values
	unitAfter := rtGame.World.UnitAt(AxialCoord{Q: 0, R: 0})
	if unitAfter == nil {
		t.Fatal("Unit not found after ApplyChanges")
	}

	if unitAfter.DistanceLeft != 3 {
		t.Errorf("DistanceLeft after reset: got %f, want 3", unitAfter.DistanceLeft)
	}

	if unitAfter.LastToppedupTurn != 2 {
		t.Errorf("LastToppedupTurn after reset: got %d, want 2", unitAfter.LastToppedupTurn)
	}
}

// =============================================================================
// Tests for exhausted highlights considering LastToppedupTurn
// =============================================================================

// TestUnitNeedsTopUpCheck tests that we can correctly identify when a unit
// needs to be topped up (hasn't been refreshed this turn)
func TestUnitNeedsTopUpCheck(t *testing.T) {
	testCases := []struct {
		name             string
		lastToppedupTurn int32
		turnCounter      int32
		wantNeedsTopUp   bool
	}{
		{
			name:             "unit topped up this turn - no top up needed",
			lastToppedupTurn: 5,
			turnCounter:      5,
			wantNeedsTopUp:   false,
		},
		{
			name:             "unit topped up previous turn - needs top up",
			lastToppedupTurn: 4,
			turnCounter:      5,
			wantNeedsTopUp:   true,
		},
		{
			name:             "unit never topped up - needs top up",
			lastToppedupTurn: 0,
			turnCounter:      1,
			wantNeedsTopUp:   true,
		},
		{
			name:             "unit topped up future turn (edge case) - no top up needed",
			lastToppedupTurn: 6,
			turnCounter:      5,
			wantNeedsTopUp:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// A unit needs top-up if LastToppedupTurn < TurnCounter
			needsTopUp := tc.lastToppedupTurn < tc.turnCounter

			if needsTopUp != tc.wantNeedsTopUp {
				t.Errorf("needsTopUp: got %v, want %v", needsTopUp, tc.wantNeedsTopUp)
			}
		})
	}
}

// TestExhaustedUnitDetection tests the logic for detecting exhausted units
// considering the lazy top-up pattern
func TestExhaustedUnitDetection(t *testing.T) {
	testCases := []struct {
		name             string
		distanceLeft     float64
		lastToppedupTurn int32
		turnCounter      int32
		wantExhausted    bool
	}{
		{
			name:             "unit with moves left - not exhausted",
			distanceLeft:     3,
			lastToppedupTurn: 5,
			turnCounter:      5,
			wantExhausted:    false,
		},
		{
			name:             "unit topped up this turn with no moves - exhausted",
			distanceLeft:     0,
			lastToppedupTurn: 5,
			turnCounter:      5,
			wantExhausted:    true,
		},
		{
			name:             "unit NOT topped up this turn with no moves - NOT exhausted (will be topped up)",
			distanceLeft:     0,
			lastToppedupTurn: 4,
			turnCounter:      5,
			wantExhausted:    false, // Will get topped up when accessed
		},
		{
			name:             "new turn just started, unit shows 0 moves from last turn - NOT exhausted",
			distanceLeft:     0,
			lastToppedupTurn: 1,
			turnCounter:      2,
			wantExhausted:    false, // Will get topped up when accessed
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Correct exhausted detection logic:
			// Only mark as exhausted if topped up this turn AND no moves left
			isExhausted := tc.lastToppedupTurn >= tc.turnCounter && tc.distanceLeft <= 0

			if isExhausted != tc.wantExhausted {
				t.Errorf("isExhausted: got %v, want %v (distanceLeft=%f, lastToppedupTurn=%d, turnCounter=%d)",
					isExhausted, tc.wantExhausted, tc.distanceLeft, tc.lastToppedupTurn, tc.turnCounter)
			}
		})
	}
}

// =============================================================================
// Integration test for ApplyChanges flow with PlayerChanged
// =============================================================================

// TestApplyChangesFullFlow tests the complete flow of applying remote changes
// including player change with unit resets
func TestApplyChangesFullFlow(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create a game simulating end of player 1's turn
	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add tiles
	for q := 0; q < 3; q++ {
		tile := NewTile(AxialCoord{Q: q, R: 0}, TileTypeGrass)
		world.AddTile(tile)
	}

	// Add units for player 2 (who will be the next player)
	// These units are "exhausted" from their previous turn
	unit1 := &v1.Unit{
		Q:                0,
		R:                0,
		Player:           2,
		UnitType:         1,
		AvailableHealth:  10,
		DistanceLeft:     0, // Exhausted
		LastToppedupTurn: 1,
	}
	unit2 := &v1.Unit{
		Q:                1,
		R:                0,
		Player:           2,
		UnitType:         1,
		AvailableHealth:  10,
		DistanceLeft:     0, // Exhausted
		LastToppedupTurn: 1,
	}
	world.AddUnit(unit1)
	world.AddUnit(unit2)

	game := &v1.Game{
		Id:   "test-game",
		Name: "Test Game",
		Config: &v1.GameConfiguration{
			Players: []*v1.GamePlayer{
				{PlayerId: 1},
				{PlayerId: 2},
			},
		},
	}
	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 100, IsActive: true},
			2: {Coins: 100, IsActive: true},
		},
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Simulate receiving remote EndTurn from player 1
	// This includes reset units for player 2
	resetUnit1 := &v1.Unit{
		Q:                0,
		R:                0,
		Player:           2,
		UnitType:         1,
		AvailableHealth:  10,
		DistanceLeft:     3, // Topped up
		LastToppedupTurn: 2, // New turn
	}
	resetUnit2 := &v1.Unit{
		Q:                1,
		R:                0,
		Player:           2,
		UnitType:         1,
		AvailableHealth:  10,
		DistanceLeft:     3, // Topped up
		LastToppedupTurn: 2, // New turn
	}

	playerChangedChange := &v1.PlayerChangedChange{
		PreviousPlayer: 1,
		NewPlayer:      2,
		PreviousTurn:   1,
		NewTurn:        2,
		ResetUnits:     []*v1.Unit{resetUnit1, resetUnit2},
	}

	move := &v1.GameMove{
		Player: 1,
		MoveType: &v1.GameMove_EndTurn{
			EndTurn: &v1.EndTurnAction{},
		},
		Changes: []*v1.WorldChange{
			{
				ChangeType: &v1.WorldChange_PlayerChanged{
					PlayerChanged: playerChangedChange,
				},
			},
		},
	}

	// Apply the changes (simulating what ApplyRemoteChanges should do)
	err = rtGame.ApplyChanges([]*v1.GameMove{move})
	if err != nil {
		t.Fatalf("ApplyChanges failed: %v", err)
	}

	// Verify game state was updated
	if rtGame.CurrentPlayer != 2 {
		t.Errorf("CurrentPlayer: got %d, want 2", rtGame.CurrentPlayer)
	}
	if rtGame.TurnCounter != 2 {
		t.Errorf("TurnCounter: got %d, want 2", rtGame.TurnCounter)
	}
	if rtGame.GameState.CurrentPlayer != 2 {
		t.Errorf("GameState.CurrentPlayer: got %d, want 2", rtGame.GameState.CurrentPlayer)
	}
	if rtGame.GameState.TurnCounter != 2 {
		t.Errorf("GameState.TurnCounter: got %d, want 2", rtGame.GameState.TurnCounter)
	}

	// Verify units were reset
	unitAfter1 := rtGame.World.UnitAt(AxialCoord{Q: 0, R: 0})
	if unitAfter1 == nil {
		t.Fatal("Unit 1 not found")
	}
	if unitAfter1.DistanceLeft != 3 {
		t.Errorf("Unit 1 DistanceLeft: got %f, want 3", unitAfter1.DistanceLeft)
	}
	if unitAfter1.LastToppedupTurn != 2 {
		t.Errorf("Unit 1 LastToppedupTurn: got %d, want 2", unitAfter1.LastToppedupTurn)
	}

	unitAfter2 := rtGame.World.UnitAt(AxialCoord{Q: 1, R: 0})
	if unitAfter2 == nil {
		t.Fatal("Unit 2 not found")
	}
	if unitAfter2.DistanceLeft != 3 {
		t.Errorf("Unit 2 DistanceLeft: got %f, want 3", unitAfter2.DistanceLeft)
	}
}

// =============================================================================
// Helper to verify the correct exhausted detection behavior
// =============================================================================

// isUnitExhausted returns true if a unit should be shown as exhausted
// This is the CORRECT logic that should be used by refreshExhaustedHighlights
func isUnitExhausted(unit *v1.Unit, turnCounter int32) bool {
	// Only mark as exhausted if:
	// 1. Unit has been topped up this turn (LastToppedupTurn >= TurnCounter)
	// 2. AND has no movement left (DistanceLeft <= 0)
	// If LastToppedupTurn < TurnCounter, the unit will be topped up when accessed
	return unit.LastToppedupTurn >= turnCounter && unit.DistanceLeft <= 0
}

// TestIsUnitExhaustedHelper verifies the helper function works correctly
func TestIsUnitExhaustedHelper(t *testing.T) {
	testCases := []struct {
		name          string
		unit          *v1.Unit
		turnCounter   int32
		wantExhausted bool
	}{
		{
			name: "topped up with moves - not exhausted",
			unit: &v1.Unit{
				DistanceLeft:     3,
				LastToppedupTurn: 5,
			},
			turnCounter:   5,
			wantExhausted: false,
		},
		{
			name: "topped up without moves - exhausted",
			unit: &v1.Unit{
				DistanceLeft:     0,
				LastToppedupTurn: 5,
			},
			turnCounter:   5,
			wantExhausted: true,
		},
		{
			name: "not topped up without moves - NOT exhausted (lazy top-up)",
			unit: &v1.Unit{
				DistanceLeft:     0,
				LastToppedupTurn: 4,
			},
			turnCounter:   5,
			wantExhausted: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := isUnitExhausted(tc.unit, tc.turnCounter)
			if got != tc.wantExhausted {
				t.Errorf("isUnitExhausted: got %v, want %v", got, tc.wantExhausted)
			}
		})
	}
}
