package lib

import (
	"testing"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
)

// TestProcessBuildUnit_BasicBuild tests successful unit building
func TestProcessBuildUnit_BasicBuild(t *testing.T) {
	game := newTestGameBuilder().
		tile(0, 0, TileTypeLandBase, 1). // Player 1's base
		coins(1, 500).
		currentPlayer(1).
		build()

	move := &v1.GameMove{
		MoveType: &v1.GameMove_BuildUnit{
			BuildUnit: &v1.BuildUnitAction{
				Pos:      &v1.Position{Q: 0, R: 0},
				UnitType: testUnitTypeSoldier, // Costs 75
			},
		},
	}

	if err := game.ProcessMove(move); err != nil {
		t.Fatalf("ProcessMove failed: %v", err)
	}

	// Verify unit was created
	unit := game.World.UnitAt(AxialCoord{Q: 0, R: 0})
	if unit == nil {
		t.Fatal("Unit should be created at base")
	}

	if unit.Player != 1 {
		t.Errorf("Unit should belong to player 1, got %d", unit.Player)
	}

	if unit.UnitType != testUnitTypeSoldier {
		t.Errorf("Unit type should be %d, got %d", testUnitTypeSoldier, unit.UnitType)
	}
}

// TestProcessBuildUnit_DeductsCoins tests coin deduction
func TestProcessBuildUnit_DeductsCoins(t *testing.T) {
	initialCoins := int32(500)
	game := newTestGameBuilder().
		tile(0, 0, TileTypeLandBase, 1).
		coins(1, initialCoins).
		currentPlayer(1).
		build()

	move := &v1.GameMove{
		MoveType: &v1.GameMove_BuildUnit{
			BuildUnit: &v1.BuildUnitAction{
				Pos:      &v1.Position{Q: 0, R: 0},
				UnitType: testUnitTypeSoldier, // Costs 75
			},
		},
	}

	if err := game.ProcessMove(move); err != nil {
		t.Fatalf("ProcessMove failed: %v", err)
	}

	// Check coins were deducted
	playerState := game.GameState.PlayerStates[1]
	if playerState.Coins >= initialCoins {
		t.Errorf("Coins should decrease: was %d, now %d", initialCoins, playerState.Coins)
	}
}

// TestProcessBuildUnit_InsufficientCoins tests coin validation
func TestProcessBuildUnit_InsufficientCoins(t *testing.T) {
	game := newTestGameBuilder().
		tile(0, 0, TileTypeLandBase, 1).
		coins(1, 10). // Too few coins for soldier (75)
		currentPlayer(1).
		build()

	move := &v1.GameMove{
		MoveType: &v1.GameMove_BuildUnit{
			BuildUnit: &v1.BuildUnitAction{
				Pos:      &v1.Position{Q: 0, R: 0},
				UnitType: testUnitTypeSoldier,
			},
		},
	}

	if err := game.ProcessMove(move); err == nil {
		t.Error("Should fail with insufficient coins")
	}
}

// TestProcessBuildUnit_NotOwnedBase tests ownership validation
func TestProcessBuildUnit_NotOwnedBase(t *testing.T) {
	game := newTestGameBuilder().
		tile(0, 0, TileTypeLandBase, 2). // Player 2's base
		coins(1, 500).
		currentPlayer(1).
		build()

	move := &v1.GameMove{
		MoveType: &v1.GameMove_BuildUnit{
			BuildUnit: &v1.BuildUnitAction{
				Pos:      &v1.Position{Q: 0, R: 0},
				UnitType: testUnitTypeSoldier,
			},
		},
	}

	if err := game.ProcessMove(move); err == nil {
		t.Error("Should not build on opponent's base")
	}
}

// TestProcessBuildUnit_OccupiedTile tests occupied tile blocking
func TestProcessBuildUnit_OccupiedTile(t *testing.T) {
	game := newTestGameBuilder().
		tile(0, 0, TileTypeLandBase, 1).
		unit(0, 0, 1, testUnitTypeSoldier). // Already occupied
		coins(1, 500).
		currentPlayer(1).
		build()

	move := &v1.GameMove{
		MoveType: &v1.GameMove_BuildUnit{
			BuildUnit: &v1.BuildUnitAction{
				Pos:      &v1.Position{Q: 0, R: 0},
				UnitType: testUnitTypeSoldier,
			},
		},
	}

	if err := game.ProcessMove(move); err == nil {
		t.Error("Should not build on occupied tile")
	}
}

// TestProcessBuildUnit_WrongTurn tests turn validation
func TestProcessBuildUnit_WrongTurn(t *testing.T) {
	game := newTestGameBuilder().
		tile(0, 0, TileTypeLandBase, 1).
		coins(1, 500).
		currentPlayer(2). // Player 2's turn
		build()

	move := &v1.GameMove{
		MoveType: &v1.GameMove_BuildUnit{
			BuildUnit: &v1.BuildUnitAction{
				Pos:      &v1.Position{Q: 0, R: 0},
				UnitType: testUnitTypeSoldier,
			},
		},
	}

	if err := game.ProcessMove(move); err == nil {
		t.Error("Should not build on opponent's turn")
	}
}

// TestProcessBuildUnit_NonBuildableTerrain tests terrain validation
func TestProcessBuildUnit_NonBuildableTerrain(t *testing.T) {
	game := newTestGameBuilder().
		tile(0, 0, TileTypeGrass, 1). // Grass can't build
		coins(1, 500).
		currentPlayer(1).
		build()

	move := &v1.GameMove{
		MoveType: &v1.GameMove_BuildUnit{
			BuildUnit: &v1.BuildUnitAction{
				Pos:      &v1.Position{Q: 0, R: 0},
				UnitType: testUnitTypeSoldier,
			},
		},
	}

	if err := game.ProcessMove(move); err == nil {
		t.Error("Should not build on non-buildable terrain")
	}
}

// TestProcessBuildUnit_ChangeRecorded tests change recording
func TestProcessBuildUnit_ChangeRecorded(t *testing.T) {
	game := newTestGameBuilder().
		tile(0, 0, TileTypeLandBase, 1).
		coins(1, 500).
		currentPlayer(1).
		build()

	move := &v1.GameMove{
		MoveType: &v1.GameMove_BuildUnit{
			BuildUnit: &v1.BuildUnitAction{
				Pos:      &v1.Position{Q: 0, R: 0},
				UnitType: testUnitTypeSoldier,
			},
		},
	}

	if err := game.ProcessMove(move); err != nil {
		t.Fatalf("ProcessMove failed: %v", err)
	}

	hasBuildChange := false
	hasCoinsChange := false
	for _, change := range move.Changes {
		if _, ok := change.ChangeType.(*v1.WorldChange_UnitBuilt); ok {
			hasBuildChange = true
		}
		if _, ok := change.ChangeType.(*v1.WorldChange_CoinsChanged); ok {
			hasCoinsChange = true
		}
	}

	if !hasBuildChange {
		t.Error("Build should record UnitBuilt change")
	}
	if !hasCoinsChange {
		t.Error("Build should record CoinsChanged change")
	}
}

// TestProcessBuildUnit_NewUnitCannotMove tests newly built unit mobility
func TestProcessBuildUnit_NewUnitCannotMove(t *testing.T) {
	game := newTestGameBuilder().
		tile(0, 0, TileTypeLandBase, 1).
		tile(1, 0, TileTypeGrass, 0).
		coins(1, 500).
		currentPlayer(1).
		build()

	// Build a unit
	buildMove := &v1.GameMove{
		MoveType: &v1.GameMove_BuildUnit{
			BuildUnit: &v1.BuildUnitAction{
				Pos:      &v1.Position{Q: 0, R: 0},
				UnitType: testUnitTypeSoldier,
			},
		},
	}

	if err := game.ProcessMove(buildMove); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify new unit has 0 distance
	unit := game.World.UnitAt(AxialCoord{Q: 0, R: 0})
	if unit == nil {
		t.Fatal("Unit not created")
	}

	if unit.DistanceLeft > 0 {
		t.Errorf("Newly built unit should have 0 movement, got %f", unit.DistanceLeft)
	}
}

// TestProcessBuildUnit_OneBuildPerTilePerTurn tests build limit
func TestProcessBuildUnit_OneBuildPerTilePerTurn(t *testing.T) {
	game := newTestGameBuilder().
		tile(0, 0, TileTypeLandBase, 1).
		tile(1, 0, TileTypeGrass, 0). // For unit to move to
		coins(1, 1000).
		currentPlayer(1).
		build()

	// First build
	move1 := &v1.GameMove{
		MoveType: &v1.GameMove_BuildUnit{
			BuildUnit: &v1.BuildUnitAction{
				Pos:      &v1.Position{Q: 0, R: 0},
				UnitType: testUnitTypeSoldier,
			},
		},
	}

	if err := game.ProcessMove(move1); err != nil {
		t.Fatalf("First build failed: %v", err)
	}

	// Move unit away to free tile
	unit := game.World.UnitAt(AxialCoord{Q: 0, R: 0})
	unit.DistanceLeft = 3 // Give it movement points
	game.World.MoveUnit(unit, AxialCoord{Q: 1, R: 0})

	// Try second build on same tile same turn
	move2 := &v1.GameMove{
		MoveType: &v1.GameMove_BuildUnit{
			BuildUnit: &v1.BuildUnitAction{
				Pos:      &v1.Position{Q: 0, R: 0},
				UnitType: testUnitTypeSoldier,
			},
		},
	}

	if err := game.ProcessMove(move2); err == nil {
		t.Error("Should not build twice on same tile per turn")
	}
}
