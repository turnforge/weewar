package tests

import (
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// =============================================================================
// Tests for heal mechanics
// =============================================================================

// TestHealingOnTurnStart tests that units heal at turn start based on terrain
func TestHealingOnTurnStart(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add a tile with healing capability
	tile := NewTile(AxialCoord{Q: 1, R: 0}, TileTypeGrass)
	tile.Player = 1 // Owned by player 1
	world.AddTile(tile)

	game := &v1.Game{Id: "test-game", Name: "Test Game"}
	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   2, // Turn 2 so healing check works
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 100, IsActive: true},
			2: {Coins: 100, IsActive: true},
		},
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Create a unit with reduced health that didn't act last turn
	unit := &v1.Unit{
		UnitType:         1, // Trooper
		Player:           1,
		Q:                1,
		R:                0,
		AvailableHealth:  5,              // Below max
		DistanceLeft:     0,              // Depleted
		LastActedTurn:    0,              // Didn't act
		LastToppedupTurn: 1,              // Last topped up turn 1
		Shortcut:         "A1",
	}
	world.AddUnit(unit)

	// Top up the unit (this should trigger healing)
	err = rtGame.TopUpUnitIfNeeded(unit)
	if err != nil {
		t.Fatalf("TopUpUnitIfNeeded failed: %v", err)
	}

	// Verify movement was restored
	if unit.DistanceLeft <= 0 {
		t.Errorf("DistanceLeft should be restored, got %f", unit.DistanceLeft)
	}

	// Verify unit was marked as topped up
	if unit.LastToppedupTurn != rtGame.TurnCounter {
		t.Errorf("LastToppedupTurn: got %d, want %d", unit.LastToppedupTurn, rtGame.TurnCounter)
	}

	t.Logf("Unit health after top-up: %d (was 5)", unit.AvailableHealth)
	t.Logf("Unit movement after top-up: %f", unit.DistanceLeft)
}

// TestNoHealingIfActedLastTurn tests that units don't heal if they acted last turn
func TestNoHealingIfActedLastTurn(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	tile := NewTile(AxialCoord{Q: 1, R: 0}, TileTypeGrass)
	tile.Player = 1
	world.AddTile(tile)

	game := &v1.Game{Id: "test-game", Name: "Test Game"}
	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   3, // Turn 3
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 100, IsActive: true},
			2: {Coins: 100, IsActive: true},
		},
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Create a unit that acted last turn (turn 2)
	unit := &v1.Unit{
		UnitType:         1,
		Player:           1,
		Q:                1,
		R:                0,
		AvailableHealth:  5,
		DistanceLeft:     0,
		LastActedTurn:    2, // Acted last turn
		LastToppedupTurn: 2,
		Shortcut:         "A1",
	}
	world.AddUnit(unit)

	initialHealth := unit.AvailableHealth

	err = rtGame.TopUpUnitIfNeeded(unit)
	if err != nil {
		t.Fatalf("TopUpUnitIfNeeded failed: %v", err)
	}

	// Health should remain the same (no healing because unit acted)
	if unit.AvailableHealth != initialHealth {
		t.Errorf("Health changed from %d to %d, but unit acted last turn and shouldn't heal",
			initialHealth, unit.AvailableHealth)
	}
}

// TestNoHealingOnEnemyBase tests that units can't heal on enemy-owned tiles
func TestNoHealingOnEnemyBase(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Create a tile owned by player 2 (enemy)
	tile := NewTile(AxialCoord{Q: 3, R: 4}, TileTypeGrass)
	tile.Player = 2 // Enemy owns this tile
	world.AddTile(tile)

	game := &v1.Game{Id: "test-game", Name: "Test Game"}
	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   2,
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 100, IsActive: true},
			2: {Coins: 100, IsActive: true},
		},
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Create a unit on enemy tile
	unit := &v1.Unit{
		UnitType:         1,
		Player:           1,
		Q:                3,
		R:                4,
		AvailableHealth:  5,
		DistanceLeft:     0,
		LastActedTurn:    0, // Didn't act
		LastToppedupTurn: 1,
		Shortcut:         "A1",
	}
	world.AddUnit(unit)

	initialHealth := unit.AvailableHealth

	err = rtGame.TopUpUnitIfNeeded(unit)
	if err != nil {
		t.Fatalf("TopUpUnitIfNeeded failed: %v", err)
	}

	// Health should remain the same (no healing on enemy base)
	if unit.AvailableHealth != initialHealth {
		t.Errorf("Health changed from %d to %d, but unit is on enemy base and shouldn't heal",
			initialHealth, unit.AvailableHealth)
	}
}

// TestProcessEndTurnTopsUpIncomingPlayer tests that ProcessEndTurn tops up the incoming player
func TestProcessEndTurnTopsUpIncomingPlayer(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add a tile for player 2's unit
	tile := NewTile(AxialCoord{Q: 3, R: 4}, TileTypeGrass)
	world.AddTile(tile)

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

	// Create a unit for player 2 with depleted movement
	unit := &v1.Unit{
		UnitType:         1,
		Player:           2,
		Q:                3,
		R:                4,
		AvailableHealth:  10,
		DistanceLeft:     0, // Depleted
		LastActedTurn:    0,
		LastToppedupTurn: 0, // Never topped up
		Shortcut:         "B1",
	}
	world.AddUnit(unit)

	// End turn for player 1
	move := &v1.GameMove{Player: 1}
	action := &v1.EndTurnAction{}
	err = rtGame.ProcessEndTurn(move, action)
	if err != nil {
		t.Fatalf("ProcessEndTurn failed: %v", err)
	}

	// Verify player advanced to player 2
	if rtGame.CurrentPlayer != 2 {
		t.Errorf("CurrentPlayer: got %d, want 2", rtGame.CurrentPlayer)
	}

	// Check the world change has reset units with topped-up values
	if len(move.Changes) < 1 {
		t.Fatalf("Expected at least 1 change, got %d", len(move.Changes))
	}

	// Find PlayerChanged in changes (might be after CoinsChanged)
	var playerChange *v1.PlayerChangedChange
	for _, change := range move.Changes {
		if pc := change.GetPlayerChanged(); pc != nil {
			playerChange = pc
			break
		}
	}

	if playerChange == nil {
		t.Fatal("No PlayerChanged in move changes")
	}

	// Find the reset unit
	if len(playerChange.ResetUnits) != 1 {
		t.Fatalf("Expected 1 reset unit, got %d", len(playerChange.ResetUnits))
	}

	resetUnit := playerChange.ResetUnits[0]
	if resetUnit.DistanceLeft <= 0 {
		t.Errorf("Reset unit should have movement restored, got %f", resetUnit.DistanceLeft)
	}

	t.Logf("Reset unit: distanceLeft=%f, lastToppedupTurn=%d",
		resetUnit.DistanceLeft, resetUnit.LastToppedupTurn)
}

// TestProcessHealUnit tests the manual heal action
func TestProcessHealUnit(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add a tile owned by player 1 (friendly base)
	tile := NewTile(AxialCoord{Q: 1, R: 0}, TileTypeGrass)
	tile.Player = 1
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

	// Create a unit with reduced health that hasn't acted
	unit := &v1.Unit{
		UnitType:         1, // Trooper
		Player:           1,
		Q:                1,
		R:                0,
		AvailableHealth:  5, // Below max
		DistanceLeft:     3, // Has movement points
		LastActedTurn:    0, // Hasn't acted this turn
		LastToppedupTurn: 1,
		Shortcut:         "A1",
	}
	world.AddUnit(unit)

	initialHealth := unit.AvailableHealth

	// Process heal action
	move := &v1.GameMove{Player: 1}
	action := &v1.HealUnitAction{
		Pos: &v1.Position{Q: 1, R: 0},
	}
	err = rtGame.ProcessHealUnit(move, action)
	if err != nil {
		t.Fatalf("ProcessHealUnit failed: %v", err)
	}

	// Verify health increased
	if unit.AvailableHealth <= initialHealth {
		t.Errorf("Health should have increased: got %d, was %d", unit.AvailableHealth, initialHealth)
	}

	// Verify unit was marked as having acted
	if unit.LastActedTurn != rtGame.TurnCounter {
		t.Errorf("LastActedTurn should be %d, got %d", rtGame.TurnCounter, unit.LastActedTurn)
	}

	// Verify a UnitHealed change was recorded
	if len(move.Changes) == 0 {
		t.Fatal("Expected at least 1 change")
	}

	var healChange *v1.UnitHealedChange
	for _, change := range move.Changes {
		if hc := change.GetUnitHealed(); hc != nil {
			healChange = hc
			break
		}
	}

	if healChange == nil {
		t.Fatal("No UnitHealedChange in move changes")
	}

	t.Logf("Unit healed: %d -> %d (+%d)", initialHealth, unit.AvailableHealth, healChange.HealAmount)
}

// TestHealOptionShownForDamagedUnit tests that heal option is returned for damaged units
func TestHealOptionShownForDamagedUnit(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add a tile owned by player 1
	tile := NewTile(AxialCoord{Q: 1, R: 0}, TileTypeGrass)
	tile.Player = 1
	world.AddTile(tile)

	game := &v1.Game{Id: "test-game", Name: "Test Game"}
	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 100, IsActive: true},
		},
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Create a damaged unit that hasn't acted
	unit := &v1.Unit{
		UnitType:         1, // Trooper
		Player:           1,
		Q:                1,
		R:                0,
		AvailableHealth:  5, // Below max (10)
		DistanceLeft:     3,
		LastActedTurn:    0, // Hasn't acted this turn
		LastToppedupTurn: 1,
		Shortcut:         "A1",
	}
	world.AddUnit(unit)

	options, _, err := rtGame.GetUnitOptions(unit)
	if err != nil {
		t.Fatalf("GetUnitOptions failed: %v", err)
	}

	// Look for heal option
	var healOption *v1.HealUnitAction
	for _, opt := range options {
		if heal := opt.GetHeal(); heal != nil {
			healOption = heal
			break
		}
	}

	if healOption == nil {
		t.Fatal("Expected heal option for damaged unit on friendly terrain")
	}

	t.Logf("Heal option found: heal amount = %d", healOption.HealAmount)
}

// TestNoHealOptionForFullHealthUnit tests that heal option is not shown for full health units
func TestNoHealOptionForFullHealthUnit(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	tile := NewTile(AxialCoord{Q: 1, R: 0}, TileTypeGrass)
	tile.Player = 1
	world.AddTile(tile)

	game := &v1.Game{Id: "test-game", Name: "Test Game"}
	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 100, IsActive: true},
		},
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Create a full health unit
	unit := &v1.Unit{
		UnitType:         1, // Trooper
		Player:           1,
		Q:                1,
		R:                0,
		AvailableHealth:  10, // Full health
		DistanceLeft:     3,
		LastActedTurn:    0,
		LastToppedupTurn: 1,
		Shortcut:         "A1",
	}
	world.AddUnit(unit)

	options, _, err := rtGame.GetUnitOptions(unit)
	if err != nil {
		t.Fatalf("GetUnitOptions failed: %v", err)
	}

	// Look for heal option - should not exist
	for _, opt := range options {
		if opt.GetHeal() != nil {
			t.Error("Heal option should not be shown for full health unit")
		}
	}
}
