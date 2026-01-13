package tests

import (
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
)

// TestGameSetup is a helper that creates a test game with configurable world
type TestGameSetup struct {
	World       *lib.World
	Game        *lib.Game
	RulesEngine *lib.RulesEngine
}

// NewTestGameSetup creates a minimal test game environment
func NewTestGameSetup(t *testing.T) *TestGameSetup {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	world := lib.NewWorld("test", nil)

	game := &v1.Game{
		Id:   "test-game",
		Name: "Test Game",
		Config: &v1.GameConfiguration{
			Players: []*v1.GamePlayer{
				{PlayerId: 1, StartingCoins: 1000},
				{PlayerId: 2, StartingCoins: 1000},
			},
		},
	}

	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 1000, IsActive: true},
			2: {Coins: 1000, IsActive: true},
		},
	}

	rtGame := lib.NewGame(game, gameState, world, rulesEngine, 12345)

	return &TestGameSetup{
		World:       world,
		Game:        rtGame,
		RulesEngine: rulesEngine,
	}
}

// AddTile adds a tile to the test world
func (s *TestGameSetup) AddTile(q, r int, tileType int) *v1.Tile {
	coord := lib.AxialCoord{Q: q, R: r}
	tile := lib.NewTile(coord, tileType)
	s.World.AddTile(tile)
	return tile
}

// AddPlayerTile adds a tile owned by a player
func (s *TestGameSetup) AddPlayerTile(q, r int, tileType int, player int32) *v1.Tile {
	tile := s.AddTile(q, r, tileType)
	tile.Player = player
	return tile
}

// AddUnit adds a unit to the test world
func (s *TestGameSetup) AddUnit(q, r int, player, unitType int32) *v1.Unit {
	unit := &v1.Unit{
		Q:               int32(q),
		R:               int32(r),
		Player:          player,
		UnitType:        unitType,
		AvailableHealth: 10,
		DistanceLeft:    3,
	}
	s.World.AddUnit(unit)
	return unit
}

// AddUnitWithShortcut adds a unit with a specific shortcut
// Note: Sets shortcut BEFORE adding to ensure proper indexing
func (s *TestGameSetup) AddUnitWithShortcut(q, r int, player, unitType int32, shortcut string) *v1.Unit {
	unit := &v1.Unit{
		Q:               int32(q),
		R:               int32(r),
		Player:          player,
		UnitType:        unitType,
		AvailableHealth: 10,
		DistanceLeft:    3,
		Shortcut:        shortcut,
	}
	s.World.AddUnit(unit)
	return unit
}

// AddTileWithShortcut adds a tile with a specific shortcut
// Note: Sets shortcut BEFORE adding to ensure proper indexing
func (s *TestGameSetup) AddTileWithShortcut(q, r int, tileType int, shortcut string) *v1.Tile {
	coord := lib.AxialCoord{Q: q, R: r}
	tile := lib.NewTile(coord, tileType)
	tile.Shortcut = shortcut
	s.World.AddTile(tile)
	return tile
}

// AddPlayerTileWithShortcut adds a tile owned by a player with a specific shortcut
// Note: Sets Player and Shortcut BEFORE adding to ensure proper indexing
func (s *TestGameSetup) AddPlayerTileWithShortcut(q, r int, tileType int, player int32, shortcut string) *v1.Tile {
	coord := lib.AxialCoord{Q: q, R: r}
	tile := lib.NewTile(coord, tileType)
	tile.Player = player   // Must set before AddTile for shortcut indexing
	tile.Shortcut = shortcut
	s.World.AddTile(tile)
	return tile
}

// AddGrassTiles adds a grid of grass tiles
func (s *TestGameSetup) AddGrassTiles(minQ, maxQ, minR, maxR int) {
	for q := minQ; q <= maxQ; q++ {
		for r := minR; r <= maxR; r++ {
			s.AddTile(q, r, lib.TileTypeGrass)
		}
	}
}

// ================================================================
// Controller API Tests
// ================================================================

// TestMoveController tests the Move controller method
func TestMoveController(t *testing.T) {
	setup := NewTestGameSetup(t)

	// Create a small map with grass tiles
	setup.AddGrassTiles(-2, 2, -2, 2)

	// Add a trooper (unit type 1) for player 1 at (0,0)
	unit := setup.AddUnitWithShortcut(0, 0, 1, lib.UnitTypeSoldier, "A1")

	// Verify initial position
	initialUnit := setup.World.UnitAt(lib.AxialCoord{Q: 0, R: 0})
	if initialUnit == nil {
		t.Fatal("Unit not found at initial position")
	}

	// Move using relative direction "R" (right)
	changes, err := setup.Game.Move("A1", "R")
	if err != nil {
		t.Fatalf("Move failed: %v", err)
	}

	// Verify move change was recorded
	if len(changes) == 0 {
		t.Fatal("Expected move changes to be recorded")
	}

	// Verify unit moved (right in hex coordinates is Q+1)
	movedUnit := setup.World.UnitAt(lib.AxialCoord{Q: 1, R: 0})
	if movedUnit == nil {
		t.Error("Unit not found at new position after move")
	}

	// Verify old position is empty
	if setup.World.UnitAt(lib.AxialCoord{Q: 0, R: 0}) != nil {
		t.Error("Unit still found at old position (duplication bug)")
	}

	// Verify unit count is still 1
	if setup.World.NumUnits() != 1 {
		t.Errorf("Expected 1 unit after move, got %d", setup.World.NumUnits())
	}

	// Verify unit's coordinates were updated
	if unit.Q != 1 || unit.R != 0 {
		t.Errorf("Unit coordinates not updated: Q=%d, R=%d (expected 1,0)", unit.Q, unit.R)
	}
}

// TestMoveControllerWithCoordinates tests Move using Q,R coordinates
func TestMoveControllerWithCoordinates(t *testing.T) {
	setup := NewTestGameSetup(t)
	setup.AddGrassTiles(-2, 2, -2, 2)
	setup.AddUnitWithShortcut(0, 0, 1, lib.UnitTypeSoldier, "A1")

	// Move using Q,R coordinates - (1,0) is adjacent to (0,0)
	_, err := setup.Game.Move("A1", "1,0")
	if err != nil {
		t.Fatalf("Move with coordinates failed: %v", err)
	}

	// Verify unit is at new position
	if setup.World.UnitAt(lib.AxialCoord{Q: 1, R: 0}) == nil {
		t.Error("Unit not found at target coordinates (1,0)")
	}
}

// TestMoveControllerMultipleDirections tests Move with chained directions
func TestMoveControllerMultipleDirections(t *testing.T) {
	setup := NewTestGameSetup(t)
	setup.AddGrassTiles(-3, 3, -3, 3)
	unit := setup.AddUnitWithShortcut(0, 0, 1, lib.UnitTypeSoldier, "A1")
	// Set LastToppedupTurn to current turn to prevent TopUpUnitIfNeeded from resetting DistanceLeft
	unit.LastToppedupTurn = 1
	unit.DistanceLeft = 5 // Give more movement points

	// Move using chained directions "R,R"
	_, err := setup.Game.Move("A1", "R,R")
	if err != nil {
		t.Fatalf("Move with chained directions failed: %v", err)
	}

	// Verify unit moved two steps right
	if setup.World.UnitAt(lib.AxialCoord{Q: 2, R: 0}) == nil {
		t.Error("Unit not found at expected position (2,0) after chained move")
	}
}

// TestCaptureController tests the Capture controller method
func TestCaptureController(t *testing.T) {
	setup := NewTestGameSetup(t)

	// Add a neutral land base at (0,0)
	base := setup.AddTile(0, 0, lib.TileTypeLandBase)
	base.Player = 0 // Neutral

	// Add a trooper for player 1 at the base
	unit := setup.AddUnitWithShortcut(0, 0, 1, lib.UnitTypeSoldier, "A1")

	// Start capture
	changes, err := setup.Game.Capture("A1")
	if err != nil {
		t.Fatalf("Capture failed: %v", err)
	}

	// Verify capture started change was recorded
	if len(changes) == 0 {
		t.Fatal("Expected capture changes to be recorded")
	}

	// Verify unit's capture started turn was set
	if unit.CaptureStartedTurn != 1 {
		t.Errorf("CaptureStartedTurn: got %d, want 1", unit.CaptureStartedTurn)
	}

	// Tile should NOT change ownership yet (happens on next turn)
	tile := setup.World.TileAt(lib.AxialCoord{Q: 0, R: 0})
	if tile.Player != 0 {
		t.Errorf("Tile should still be neutral after capture started, got player %d", tile.Player)
	}
}

// TestCaptureCompletesAfterTurn tests that capture completes after EndTurn cycle
func TestCaptureCompletesAfterTurn(t *testing.T) {
	setup := NewTestGameSetup(t)

	// Add a neutral land base
	base := setup.AddTile(0, 0, lib.TileTypeLandBase)
	base.Player = 0

	// Add a trooper for player 1
	unit := setup.AddUnitWithShortcut(0, 0, 1, lib.UnitTypeSoldier, "A1")

	// Start capture on turn 1
	_, err := setup.Game.Capture("A1")
	if err != nil {
		t.Fatalf("Capture failed: %v", err)
	}

	// End turn for player 1
	_, err = setup.Game.EndTurn()
	if err != nil {
		t.Fatalf("EndTurn (player 1) failed: %v", err)
	}

	// End turn for player 2 (now on turn 2)
	_, err = setup.Game.EndTurn()
	if err != nil {
		t.Fatalf("EndTurn (player 2) failed: %v", err)
	}

	// Now on turn 2, player 1's turn again
	// The capture should complete when the unit is topped up
	err = setup.Game.TopUpUnitIfNeeded(unit)
	if err != nil {
		t.Fatalf("TopUpUnitIfNeeded failed: %v", err)
	}

	// Verify tile is now owned by player 1
	tile := setup.World.TileAt(lib.AxialCoord{Q: 0, R: 0})
	if tile.Player != 1 {
		t.Errorf("Tile should be owned by player 1 after capture completes, got player %d", tile.Player)
	}

	// Verify capture started turn was cleared
	if unit.CaptureStartedTurn != 0 {
		t.Errorf("CaptureStartedTurn should be 0 after capture completes, got %d", unit.CaptureStartedTurn)
	}
}

// TestEndTurnController tests the EndTurn controller method
func TestEndTurnController(t *testing.T) {
	setup := NewTestGameSetup(t)
	setup.AddTile(0, 0, lib.TileTypeGrass)

	// Verify initial state
	if setup.Game.CurrentPlayer != 1 {
		t.Errorf("Initial player should be 1, got %d", setup.Game.CurrentPlayer)
	}
	if setup.Game.TurnCounter != 1 {
		t.Errorf("Initial turn should be 1, got %d", setup.Game.TurnCounter)
	}

	// End turn for player 1
	_, err := setup.Game.EndTurn()
	if err != nil {
		t.Fatalf("EndTurn failed: %v", err)
	}

	// Should now be player 2's turn
	if setup.Game.CurrentPlayer != 2 {
		t.Errorf("After EndTurn, current player should be 2, got %d", setup.Game.CurrentPlayer)
	}

	// Turn counter should still be 1 (increments after all players complete)
	if setup.Game.TurnCounter != 1 {
		t.Errorf("Turn counter should still be 1 after player 1's turn, got %d", setup.Game.TurnCounter)
	}

	// End turn for player 2
	_, err = setup.Game.EndTurn()
	if err != nil {
		t.Fatalf("EndTurn (player 2) failed: %v", err)
	}

	// Should cycle back to player 1, turn counter incremented
	if setup.Game.CurrentPlayer != 1 {
		t.Errorf("After full round, current player should be 1, got %d", setup.Game.CurrentPlayer)
	}
	if setup.Game.TurnCounter != 2 {
		t.Errorf("Turn counter should be 2 after full round, got %d", setup.Game.TurnCounter)
	}
}

// TestAttackController tests the Attack controller method
func TestAttackController(t *testing.T) {
	setup := NewTestGameSetup(t)
	setup.AddGrassTiles(-2, 2, -2, 2)

	// Add attacker for player 1 at (0,0)
	attacker := setup.AddUnitWithShortcut(0, 0, 1, lib.UnitTypeSoldier, "A1")
	attacker.AvailableHealth = 10

	// Add defender for player 2 at (1,0) - adjacent
	defender := setup.AddUnitWithShortcut(1, 0, 2, lib.UnitTypeSoldier, "B1")
	defender.AvailableHealth = 10

	// Attack using relative direction
	changes, err := setup.Game.Attack("A1", "R")
	if err != nil {
		t.Fatalf("Attack failed: %v", err)
	}

	// Verify attack changes were recorded
	if len(changes) == 0 {
		t.Fatal("Expected attack changes to be recorded")
	}

	// Verify defender took damage (exact amount depends on combat calculation)
	if defender.AvailableHealth >= 10 {
		t.Error("Defender should have taken damage from attack")
	}
}

// TestBuildController tests the Build controller method
func TestBuildController(t *testing.T) {
	setup := NewTestGameSetup(t)

	// Add a land base for player 1 with shortcut
	setup.AddPlayerTileWithShortcut(0, 0, lib.TileTypeLandBase, 1, "A1")

	// Build a trooper
	changes, err := setup.Game.Build("t:A1", lib.UnitTypeSoldier)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify build changes were recorded
	if len(changes) == 0 {
		t.Fatal("Expected build changes to be recorded")
	}

	// Verify unit was created at the base
	unit := setup.World.UnitAt(lib.AxialCoord{Q: 0, R: 0})
	if unit == nil {
		t.Error("No unit created at base after build")
	} else {
		if unit.UnitType != lib.UnitTypeSoldier {
			t.Errorf("Built unit type: got %d, want %d", unit.UnitType, lib.UnitTypeSoldier)
		}
		if unit.Player != 1 {
			t.Errorf("Built unit player: got %d, want 1", unit.Player)
		}
	}

	// Verify coins were deducted
	playerState := setup.Game.GameState.PlayerStates[1]
	if playerState.Coins >= 1000 {
		t.Error("Player coins should have been deducted after build")
	}
}

// TestGetOptionsAtController tests the GetOptionsAt controller method
func TestGetOptionsAtController(t *testing.T) {
	setup := NewTestGameSetup(t)
	setup.AddGrassTiles(-2, 2, -2, 2)

	// Add a trooper for player 1
	setup.AddUnitWithShortcut(0, 0, 1, lib.UnitTypeSoldier, "A1")

	// Get options for the unit
	resp, err := setup.Game.GetOptionsAt("A1")
	if err != nil {
		t.Fatalf("GetOptionsAt failed: %v", err)
	}

	// Should have move options (grass tiles around it)
	if len(resp.Options) == 0 {
		t.Error("Expected options for unit with movement points")
	}

	// Verify we have move options
	hasMoveOption := false
	for _, opt := range resp.Options {
		if opt.GetMove() != nil {
			hasMoveOption = true
			break
		}
	}
	if !hasMoveOption {
		t.Error("Expected at least one move option")
	}
}

// TestGetOptionsAtTileController tests GetOptionsAt for build options
func TestGetOptionsAtTileController(t *testing.T) {
	setup := NewTestGameSetup(t)

	// Add a land base for player 1 with shortcut
	setup.AddPlayerTileWithShortcut(0, 0, lib.TileTypeLandBase, 1, "A1")

	// Get options for the tile
	resp, err := setup.Game.GetOptionsAt("t:A1")
	if err != nil {
		t.Fatalf("GetOptionsAt failed: %v", err)
	}

	// Should have build options
	if len(resp.Options) == 0 {
		t.Error("Expected build options for player-owned base")
	}

	// Verify we have build options
	hasBuildOption := false
	for _, opt := range resp.Options {
		if opt.GetBuild() != nil {
			hasBuildOption = true
			break
		}
	}
	if !hasBuildOption {
		t.Error("Expected at least one build option")
	}
}

// TestPosMethodBasic tests the Pos method for parsing positions
func TestPosMethodBasic(t *testing.T) {
	setup := NewTestGameSetup(t)
	setup.AddGrassTiles(-2, 2, -2, 2)
	setup.AddUnitWithShortcut(0, 0, 1, lib.UnitTypeSoldier, "A1")

	// Test unit shortcut parsing
	target, err := setup.Game.Pos("A1")
	if err != nil {
		t.Fatalf("Pos('A1') failed: %v", err)
	}
	if target.Unit == nil {
		t.Error("Expected unit for shortcut 'A1'")
	}
	if target.Coordinate.Q != 0 || target.Coordinate.R != 0 {
		t.Errorf("Expected coord (0,0), got (%d,%d)", target.Coordinate.Q, target.Coordinate.R)
	}

	// Test Q,R coordinate parsing
	target, err = setup.Game.Pos("1,1")
	if err != nil {
		t.Fatalf("Pos('1,1') failed: %v", err)
	}
	if target.Coordinate.Q != 1 || target.Coordinate.R != 1 {
		t.Errorf("Expected coord (1,1), got (%d,%d)", target.Coordinate.Q, target.Coordinate.R)
	}

	// Test relative direction parsing
	target, err = setup.Game.Pos("R", "A1") // Right from A1
	if err != nil {
		t.Fatalf("Pos('R', 'A1') failed: %v", err)
	}
	if target.Coordinate.Q != 1 || target.Coordinate.R != 0 {
		t.Errorf("Expected coord (1,0) for 'R' from (0,0), got (%d,%d)", target.Coordinate.Q, target.Coordinate.R)
	}
}

// TestMoveAndCaptureSequence tests a realistic game sequence
func TestMoveAndCaptureSequence(t *testing.T) {
	setup := NewTestGameSetup(t)

	// Create a map:
	// - Player 1 base at (0,0) with a trooper
	// - Neutral base at (1,0) - adjacent
	setup.AddPlayerTile(0, 0, lib.TileTypeLandBase, 1)
	neutralBase := setup.AddTile(1, 0, lib.TileTypeLandBase)
	neutralBase.Player = 0

	unit := setup.AddUnitWithShortcut(0, 0, 1, lib.UnitTypeSoldier, "A1")

	// Move trooper to the neutral base
	_, err := setup.Game.Move("A1", "R")
	if err != nil {
		t.Fatalf("Move to neutral base failed: %v", err)
	}

	// Verify unit is at (1,0)
	if unit.Q != 1 || unit.R != 0 {
		t.Errorf("Unit should be at (1,0), got (%d,%d)", unit.Q, unit.R)
	}

	// Start capturing the neutral base
	_, err = setup.Game.Capture("A1")
	if err != nil {
		t.Fatalf("Capture failed: %v", err)
	}

	// Verify capture started
	if unit.CaptureStartedTurn != 1 {
		t.Errorf("CaptureStartedTurn should be 1, got %d", unit.CaptureStartedTurn)
	}

	// End turn for player 1
	_, err = setup.Game.EndTurn()
	if err != nil {
		t.Fatalf("EndTurn (player 1) failed: %v", err)
	}

	// End turn for player 2
	_, err = setup.Game.EndTurn()
	if err != nil {
		t.Fatalf("EndTurn (player 2) failed: %v", err)
	}

	// Back to player 1, turn 2 - top up should complete capture
	err = setup.Game.TopUpUnitIfNeeded(unit)
	if err != nil {
		t.Fatalf("TopUpUnitIfNeeded failed: %v", err)
	}

	// Verify base is now owned by player 1
	tile := setup.World.TileAt(lib.AxialCoord{Q: 1, R: 0})
	if tile.Player != 1 {
		t.Errorf("Neutral base should be owned by player 1 after capture, got player %d", tile.Player)
	}
}
