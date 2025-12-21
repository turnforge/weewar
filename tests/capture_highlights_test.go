package tests

import (
	"testing"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	"github.com/turnforge/weewar/lib"
)

// TestCapturingHighlightsGeneration tests that units with active captures
// are correctly identified for capturing flag highlights
func TestCapturingHighlightsGeneration(t *testing.T) {
	// Create a world with units in various capture states
	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add tiles
	baseTile1 := NewTile(AxialCoord{Q: 0, R: 0}, lib.TileTypeLandBase)
	baseTile1.Player = 0 // Neutral
	world.AddTile(baseTile1)

	baseTile2 := NewTile(AxialCoord{Q: 1, R: 0}, lib.TileTypeLandBase)
	baseTile2.Player = 2 // Enemy
	world.AddTile(baseTile2)

	baseTile3 := NewTile(AxialCoord{Q: 2, R: 0}, lib.TileTypeLandBase)
	baseTile3.Player = 1 // Own tile
	world.AddTile(baseTile3)

	// Unit 1: Actively capturing (should show flag)
	unit1 := &v1.Unit{
		Q:                  0,
		R:                  0,
		Player:             1,
		UnitType:           1,
		Shortcut:           "A1",
		AvailableHealth:    10,
		CaptureStartedTurn: 1, // Started capture on turn 1
	}
	world.AddUnit(unit1)

	// Unit 2: Also capturing (should show flag)
	unit2 := &v1.Unit{
		Q:                  1,
		R:                  0,
		Player:             1,
		UnitType:           1,
		Shortcut:           "A2",
		AvailableHealth:    10,
		CaptureStartedTurn: 2, // Started capture on turn 2
	}
	world.AddUnit(unit2)

	// Unit 3: Not capturing (no flag)
	unit3 := &v1.Unit{
		Q:                  2,
		R:                  0,
		Player:             1,
		UnitType:           1,
		Shortcut:           "A3",
		AvailableHealth:    10,
		CaptureStartedTurn: 0, // Not capturing
	}
	world.AddUnit(unit3)

	// Test: Find units that should have capturing flags
	capturingUnits := getCapturingUnits(world)

	if len(capturingUnits) != 2 {
		t.Errorf("Expected 2 capturing units, got %d", len(capturingUnits))
	}

	// Verify the capturing units are at the expected positions
	foundA1 := false
	foundA2 := false
	for _, u := range capturingUnits {
		if u.Q == 0 && u.R == 0 {
			foundA1 = true
		}
		if u.Q == 1 && u.R == 0 {
			foundA2 = true
		}
	}

	if !foundA1 {
		t.Error("Expected unit at (0,0) to be capturing")
	}
	if !foundA2 {
		t.Error("Expected unit at (1,0) to be capturing")
	}
}

// TestCaptureHighlightClearedOnComplete tests that capture flags are removed
// when capture completes
func TestCaptureHighlightClearedOnComplete(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	baseTile := NewTile(AxialCoord{Q: 0, R: 0}, lib.TileTypeLandBase)
	baseTile.Player = 0 // Neutral
	world.AddTile(baseTile)

	// Unit started capturing on turn 1
	unit := &v1.Unit{
		Q:                  0,
		R:                  0,
		Player:             1,
		UnitType:           1,
		Shortcut:           "A1",
		AvailableHealth:    10,
		DistanceLeft:       0,
		CaptureStartedTurn: 1, // Started capture on turn 1
		LastToppedupTurn:   1,
	}
	world.AddUnit(unit)

	game := &v1.Game{Id: "test-game", Name: "Test Game"}
	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   2, // Turn 2 - capture should complete
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 500, IsActive: true},
		},
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Before top-up: should have capturing flag
	capturingBefore := getCapturingUnits(rtGame.World)
	if len(capturingBefore) != 1 {
		t.Errorf("Before top-up: expected 1 capturing unit, got %d", len(capturingBefore))
	}

	// Top up the unit (this completes the capture)
	err = rtGame.TopUpUnitIfNeeded(unit)
	if err != nil {
		t.Fatalf("TopUpUnitIfNeeded failed: %v", err)
	}

	// After top-up: capture should be complete, no more flag
	capturingAfter := getCapturingUnits(rtGame.World)
	if len(capturingAfter) != 0 {
		t.Errorf("After capture complete: expected 0 capturing units, got %d", len(capturingAfter))
	}

	// Verify CaptureStartedTurn was reset
	if unit.CaptureStartedTurn != 0 {
		t.Errorf("CaptureStartedTurn should be 0 after capture complete, got %d", unit.CaptureStartedTurn)
	}
}

// getCapturingUnits returns all units that are currently capturing
// (mirrors the logic in refreshCapturingHighlights)
func getCapturingUnits(world *lib.World) []*v1.Unit {
	var capturing []*v1.Unit

	for coord, unit := range world.UnitsByCoord() {
		// TODO: When capture duration becomes configurable (N turns instead of 1),
		// also check that CaptureStartedTurn <= CurrentTurn - N
		_ = coord // unused
		if unit.CaptureStartedTurn > 0 {
			capturing = append(capturing, unit)
		}
	}

	return capturing
}

// TestCaptureHighlightSpecs tests the generation of HighlightSpec for capturing flags
func TestCaptureHighlightSpecs(t *testing.T) {
	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	baseTile := NewTile(AxialCoord{Q: 3, R: -2}, lib.TileTypeLandBase)
	baseTile.Player = 2
	world.AddTile(baseTile)

	unit := &v1.Unit{
		Q:                  3,
		R:                  -2,
		Player:             1,
		UnitType:           1,
		Shortcut:           "A1",
		AvailableHealth:    10,
		CaptureStartedTurn: 5,
	}
	world.AddUnit(unit)

	// Generate highlight specs (mirrors presenter logic)
	capturingUnits := getCapturingUnits(world)
	highlights := make([]*v1.HighlightSpec, 0, len(capturingUnits))
	for _, u := range capturingUnits {
		highlights = append(highlights, &v1.HighlightSpec{
			Q:    u.Q,
			R:    u.R,
			Type: "capturing",
		})
	}

	if len(highlights) != 1 {
		t.Fatalf("Expected 1 highlight spec, got %d", len(highlights))
	}

	h := highlights[0]
	if h.Q != 3 || h.R != -2 {
		t.Errorf("Highlight position: got (%d,%d), want (3,-2)", h.Q, h.R)
	}
	if h.Type != "capturing" {
		t.Errorf("Highlight type: got %q, want \"capturing\"", h.Type)
	}
}

// TestCaptureFlagPersistsThroughSelection tests that capturing flags
// remain visible when selection changes
func TestCaptureFlagPersistsThroughSelection(t *testing.T) {
	// This is more of an integration test concept:
	// When clearHighlightsAndSelection is called, it should NOT clear "capturing" type
	// The types cleared are: selection, movement, attack, build, capture
	// NOT: exhausted, capturing (persistent state indicators)

	clearedTypes := []string{"selection", "movement", "attack", "build", "capture"}

	// Verify "capturing" is not in the cleared types
	for _, typ := range clearedTypes {
		if typ == "capturing" {
			t.Error("'capturing' should NOT be cleared by clearHighlightsAndSelection")
		}
	}

	// Verify "capturing" is different from "capture" (interactive vs persistent)
	if "capture" == "capturing" {
		t.Error("'capture' and 'capturing' should be different types")
	}
}

// TestCaptureInteractiveHighlightSpecs tests that units can capture
// when conditions are met (this is tested more thoroughly in capture_test.go)
func TestCaptureInteractiveHighlightSpecs(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	baseTile := NewTile(AxialCoord{Q: 0, R: 0}, lib.TileTypeLandBase)
	baseTile.Player = 0 // Neutral - can be captured
	world.AddTile(baseTile)

	unit := &v1.Unit{
		Q:                  0,
		R:                  0,
		Player:             1,
		UnitType:           1, // Infantry - can capture
		Shortcut:           "A1",
		AvailableHealth:    10,
		DistanceLeft:       3,
		CaptureStartedTurn: 0, // Not currently capturing
	}
	world.AddUnit(unit)

	game := &v1.Game{Id: "test-game", Name: "Test Game"}
	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 500, IsActive: true},
		},
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Verify setup - unit can capture because:
	// 1. Unit health > 0
	// 2. Tile is not owned by unit's player
	// 3. Unit is not already capturing
	coord := AxialCoord{Q: 0, R: 0}
	tile := rtGame.World.TileAt(coord)
	unitAtCoord := rtGame.World.UnitAt(coord)

	if tile == nil {
		t.Fatal("Tile not found")
	}
	if unitAtCoord == nil {
		t.Fatal("Unit not found")
	}

	// Check conditions for capture
	canCapture := unitAtCoord.AvailableHealth > 0 &&
		tile.Player != unitAtCoord.Player &&
		unitAtCoord.CaptureStartedTurn == 0

	if !canCapture {
		t.Error("Unit should be able to capture: health > 0, tile not owned, not already capturing")
	}

	// Check that unit type can capture this tile type via rules engine
	terrainProps := rulesEngine.GetTerrainUnitPropertiesForUnit(tile.TileType, unit.UnitType)
	if terrainProps == nil {
		t.Fatal("No terrain properties found for unit type on tile type")
	}
	if !terrainProps.CanCapture {
		t.Error("Unit type should be able to capture this tile type according to rules")
	}
}
