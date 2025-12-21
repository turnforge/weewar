package tests

import (
	"testing"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	"github.com/turnforge/weewar/lib"
)

// TestProcessCaptureBuilding tests the basic capture functionality
func TestProcessCaptureBuilding(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create a world with a capturable base
	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add a neutral land base at (0,0)
	baseTile := NewTile(AxialCoord{Q: 0, R: 0}, lib.TileTypeLandBase)
	baseTile.Player = 0 // Neutral
	world.AddTile(baseTile)

	// Add an infantry (type 1) for player 1 at the base - infantry can capture land bases
	unit := &v1.Unit{
		Q:               0,
		R:               0,
		Player:          1,
		UnitType:        1, // Infantry - can capture
		Shortcut:        "A1",
		AvailableHealth: 10,
		DistanceLeft:    3,
	}
	world.AddUnit(unit)

	// Create game
	game := &v1.Game{
		Id:   "test-game",
		Name: "Test Game",
	}

	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 500, IsActive: true},
		},
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Execute capture
	processor := &MoveProcessor{}
	move := &v1.GameMove{
		Player: 1,
		MoveType: &v1.GameMove_CaptureBuilding{
			CaptureBuilding: &v1.CaptureBuildingAction{
				Q:        0,
				R:        0,
				TileType: lib.TileTypeLandBase,
			},
		},
	}

	err = processor.ProcessCaptureBuilding(rtGame, move, move.GetCaptureBuilding())
	if err != nil {
		t.Fatalf("ProcessCaptureBuilding failed: %v", err)
	}

	// Verify unit has capture started
	capturedUnit := rtGame.World.UnitAt(AxialCoord{Q: 0, R: 0})
	if capturedUnit == nil {
		t.Fatal("Unit not found after capture started")
	}

	if capturedUnit.CaptureStartedTurn != 1 {
		t.Errorf("CaptureStartedTurn: got %d, want 1", capturedUnit.CaptureStartedTurn)
	}

	// Verify CaptureStartedChange was recorded
	hasCaptureChange := false
	for _, change := range move.Changes {
		if cs := change.GetCaptureStarted(); cs != nil {
			hasCaptureChange = true
			if cs.TileQ != 0 || cs.TileR != 0 {
				t.Errorf("CaptureStarted tile: got (%d,%d), want (0,0)", cs.TileQ, cs.TileR)
			}
			if cs.CurrentOwner != 0 {
				t.Errorf("CaptureStarted previous owner: got %d, want 0 (neutral)", cs.CurrentOwner)
			}
		}
	}

	if !hasCaptureChange {
		t.Error("Expected CaptureStartedChange in move changes")
	}

	// Tile should NOT change ownership yet (happens on next turn)
	tile := rtGame.World.TileAt(AxialCoord{Q: 0, R: 0})
	if tile.Player != 0 {
		t.Errorf("Tile owner should still be neutral (0), got %d", tile.Player)
	}
}

// TestCaptureCompletesOnNextTurn tests that capture completes when unit's turn comes
func TestCaptureCompletesOnNextTurn(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create a world with a capturable base
	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add a neutral land base at (0,0)
	baseTile := NewTile(AxialCoord{Q: 0, R: 0}, lib.TileTypeLandBase)
	baseTile.Player = 0 // Neutral
	world.AddTile(baseTile)

	// Add an infantry with capture already started from turn 1
	unit := &v1.Unit{
		Q:                  0,
		R:                  0,
		Player:             1,
		UnitType:           1, // Infantry - can capture
		Shortcut:           "A1",
		AvailableHealth:    10,
		DistanceLeft:       0, // Exhausted
		CaptureStartedTurn: 1, // Started capture on turn 1
		LastToppedupTurn:   1,
	}
	world.AddUnit(unit)

	// Create game at turn 2 (player 1's next turn)
	game := &v1.Game{
		Id:   "test-game",
		Name: "Test Game",
	}

	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   2, // Turn 2
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 500, IsActive: true},
		},
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// TopUpUnitIfNeeded should complete the capture
	err = rtGame.TopUpUnitIfNeeded(unit)
	if err != nil {
		t.Fatalf("TopUpUnitIfNeeded failed: %v", err)
	}

	// Verify capture completed
	if unit.CaptureStartedTurn != 0 {
		t.Errorf("CaptureStartedTurn should be reset to 0, got %d", unit.CaptureStartedTurn)
	}

	// Verify tile ownership changed
	tile := rtGame.World.TileAt(AxialCoord{Q: 0, R: 0})
	if tile.Player != 1 {
		t.Errorf("Tile owner should be player 1, got %d", tile.Player)
	}
}

// TestCaptureFailsWhenAlreadyOwned tests that you can't capture your own tile
func TestCaptureFailsWhenAlreadyOwned(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add a base owned by player 1
	baseTile := NewTile(AxialCoord{Q: 0, R: 0}, lib.TileTypeLandBase)
	baseTile.Player = 1 // Owned by player 1
	world.AddTile(baseTile)

	// Add an infantry for player 1
	unit := &v1.Unit{
		Q:               0,
		R:               0,
		Player:          1,
		UnitType:        1, // Infantry - can capture
		Shortcut:        "A1",
		AvailableHealth: 10,
		DistanceLeft:    3,
	}
	world.AddUnit(unit)

	game := &v1.Game{Id: "test-game", Name: "Test Game"}
	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Try to capture own tile - should fail
	processor := &MoveProcessor{}
	move := &v1.GameMove{
		Player: 1,
		MoveType: &v1.GameMove_CaptureBuilding{
			CaptureBuilding: &v1.CaptureBuildingAction{
				Q:        0,
				R:        0,
				TileType: lib.TileTypeLandBase,
			},
		},
	}

	err = processor.ProcessCaptureBuilding(rtGame, move, move.GetCaptureBuilding())
	if err == nil {
		t.Error("Expected error when capturing own tile, got nil")
	}
}

// TestCaptureFailsWhenUnitCantCapture tests that non-capturing units fail
func TestCaptureFailsWhenUnitCantCapture(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add a neutral base
	baseTile := NewTile(AxialCoord{Q: 0, R: 0}, lib.TileTypeLandBase)
	baseTile.Player = 0
	world.AddTile(baseTile)

	// Add a tank (type 6) - cannot capture
	unit := &v1.Unit{
		Q:               0,
		R:               0,
		Player:          1,
		UnitType:        6, // Heavy Tank - cannot capture
		Shortcut:        "A1",
		AvailableHealth: 10,
		DistanceLeft:    3,
	}
	world.AddUnit(unit)

	game := &v1.Game{Id: "test-game", Name: "Test Game"}
	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	processor := &MoveProcessor{}
	move := &v1.GameMove{
		Player: 1,
		MoveType: &v1.GameMove_CaptureBuilding{
			CaptureBuilding: &v1.CaptureBuildingAction{
				Q:        0,
				R:        0,
				TileType: lib.TileTypeLandBase,
			},
		},
	}

	err = processor.ProcessCaptureBuilding(rtGame, move, move.GetCaptureBuilding())
	if err == nil {
		t.Error("Expected error when unit cannot capture, got nil")
	}
}

// TestCaptureFailsWhenAlreadyCapturing tests double capture prevention
func TestCaptureFailsWhenAlreadyCapturing(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	baseTile := NewTile(AxialCoord{Q: 0, R: 0}, lib.TileTypeLandBase)
	baseTile.Player = 0
	world.AddTile(baseTile)

	// Unit already capturing
	unit := &v1.Unit{
		Q:                  0,
		R:                  0,
		Player:             1,
		UnitType:           1, // Infantry - can capture
		Shortcut:           "A1",
		AvailableHealth:    10,
		DistanceLeft:       3,
		CaptureStartedTurn: 1, // Already capturing
	}
	world.AddUnit(unit)

	game := &v1.Game{Id: "test-game", Name: "Test Game"}
	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	processor := &MoveProcessor{}
	move := &v1.GameMove{
		Player: 1,
		MoveType: &v1.GameMove_CaptureBuilding{
			CaptureBuilding: &v1.CaptureBuildingAction{
				Q:        0,
				R:        0,
				TileType: lib.TileTypeLandBase,
			},
		},
	}

	err = processor.ProcessCaptureBuilding(rtGame, move, move.GetCaptureBuilding())
	if err == nil {
		t.Error("Expected error when unit already capturing, got nil")
	}
}

// TestCaptureEnemyBase tests capturing an enemy-owned base
func TestCaptureEnemyBase(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add a base owned by player 2
	baseTile := NewTile(AxialCoord{Q: 0, R: 0}, lib.TileTypeLandBase)
	baseTile.Player = 2 // Enemy owned
	world.AddTile(baseTile)

	// Add an infantry for player 1
	unit := &v1.Unit{
		Q:               0,
		R:               0,
		Player:          1,
		UnitType:        1, // Infantry - can capture
		Shortcut:        "A1",
		AvailableHealth: 10,
		DistanceLeft:    3,
	}
	world.AddUnit(unit)

	game := &v1.Game{Id: "test-game", Name: "Test Game"}
	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Start capture
	processor := &MoveProcessor{}
	move := &v1.GameMove{
		Player: 1,
		MoveType: &v1.GameMove_CaptureBuilding{
			CaptureBuilding: &v1.CaptureBuildingAction{
				Q:        0,
				R:        0,
				TileType: lib.TileTypeLandBase,
			},
		},
	}

	err = processor.ProcessCaptureBuilding(rtGame, move, move.GetCaptureBuilding())
	if err != nil {
		t.Fatalf("ProcessCaptureBuilding failed: %v", err)
	}

	// Verify capture started
	capturedUnit := rtGame.World.UnitAt(AxialCoord{Q: 0, R: 0})
	if capturedUnit.CaptureStartedTurn != 1 {
		t.Errorf("CaptureStartedTurn: got %d, want 1", capturedUnit.CaptureStartedTurn)
	}

	// Verify CaptureStartedChange records previous owner
	for _, change := range move.Changes {
		if cs := change.GetCaptureStarted(); cs != nil {
			if cs.CurrentOwner != 2 {
				t.Errorf("CaptureStarted previous owner: got %d, want 2", cs.CurrentOwner)
			}
		}
	}
}
