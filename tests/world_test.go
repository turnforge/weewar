package tests

import (
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// Helper function to create a test world with units and tiles
func createTestWorld(name string, units []*v1.Unit, tiles []*v1.Tile) *World {
	world := NewWorld(name, nil)

	// Add tiles
	for _, tile := range tiles {
		world.AddTile(tile)
	}

	// Add units
	for _, unit := range units {
		world.AddUnit(unit)
	}

	return world
}

// Helper to create a simple unit
func createTestUnit(q, r int, player, unitType int32) *v1.Unit {
	return &v1.Unit{
		Q:               int32(q),
		R:               int32(r),
		Player:          player,
		UnitType:        unitType,
		AvailableHealth: 10,
		DistanceLeft:    3.0,
	}
}

// Helper to create a simple tile
func createTestTile(q, r int, tileType int32) *v1.Tile {
	return &v1.Tile{
		Q:        int32(q),
		R:        int32(r),
		TileType: tileType,
		Player:   0,
	}
}

func TestWorldBasicOperations(t *testing.T) {
	world := createTestWorld("test", nil, nil)

	// Test empty world
	if world.NumUnits() != 0 {
		t.Errorf("Expected 0 units in new world, got %d", world.NumUnits())
	}

	// Test adding a unit
	unit := createTestUnit(1, 2, 1, 1)

	oldUnit, err := world.AddUnit(unit)
	if err != nil {
		t.Fatalf("Failed to add unit: %v", err)
	}
	if oldUnit != nil {
		t.Errorf("Expected no old unit, got %v", oldUnit)
	}

	// Verify unit was added
	if world.NumUnits() != 1 {
		t.Errorf("Expected 1 unit after adding, got %d", world.NumUnits())
	}

	retrievedUnit := world.UnitAt(AxialCoord{Q: 1, R: 2})
	if retrievedUnit == nil {
		t.Fatal("Failed to retrieve added unit")
	}
	if retrievedUnit.Player != 1 || retrievedUnit.UnitType != 1 {
		t.Errorf("Retrieved unit has wrong properties: player=%d, type=%d",
			retrievedUnit.Player, retrievedUnit.UnitType)
	}
}

func TestWorldMoveUnit(t *testing.T) {
	// Create world with one unit at (1,2)
	units := []*v1.Unit{createTestUnit(1, 2, 1, 1)}
	world := createTestWorld("test", units, nil)

	unit := world.UnitAt(AxialCoord{Q: 1, R: 2})
	if unit == nil {
		t.Fatal("Unit not found at original position")
	}

	// Verify no unit at destination
	if world.UnitAt(AxialCoord{Q: 3, R: 4}) != nil {
		t.Fatal("Unit found at destination before move")
	}

	// Move unit to (3,4)
	err := world.MoveUnit(unit, AxialCoord{Q: 3, R: 4})
	if err != nil {
		t.Fatalf("Failed to move unit: %v", err)
	}

	// Verify unit is at new position and not at old position
	if world.UnitAt(AxialCoord{Q: 1, R: 2}) != nil {
		t.Error("Unit still found at old position after move")
	}
	if world.UnitAt(AxialCoord{Q: 3, R: 4}) == nil {
		t.Error("Unit not found at new position after move")
	}

	// Verify unit count is still 1
	if world.NumUnits() != 1 {
		t.Errorf("Expected 1 unit after move, got %d", world.NumUnits())
	}

	// Verify unit coordinates were updated
	if unit.Q != 3 || unit.R != 4 {
		t.Errorf("Unit coordinates not updated: Q=%d, R=%d", unit.Q, unit.R)
	}
}

func TestWorldRemoveUnit(t *testing.T) {
	// Create world with one unit
	units := []*v1.Unit{createTestUnit(1, 2, 1, 1)}
	world := createTestWorld("test", units, nil)

	unit := world.UnitAt(AxialCoord{Q: 1, R: 2})
	if unit == nil {
		t.Fatal("Unit was not added properly")
	}

	// Remove the unit
	err := world.RemoveUnit(unit)
	if err != nil {
		t.Fatalf("Failed to remove unit: %v", err)
	}

	// Verify unit is removed
	if world.NumUnits() != 0 {
		t.Errorf("Expected 0 units after removal, got %d", world.NumUnits())
	}
	if world.UnitAt(AxialCoord{Q: 1, R: 2}) != nil {
		t.Error("Unit still found after removal")
	}
}

func TestWorldPushPop(t *testing.T) {
	// Create base world with a unit
	units := []*v1.Unit{createTestUnit(1, 2, 1, 1)}
	baseWorld := createTestWorld("base", units, nil)

	// Create transaction layer
	transactionWorld := baseWorld.Push()

	// Verify transaction layer sees base unit
	if transactionWorld.UnitAt(AxialCoord{Q: 1, R: 2}) == nil {
		t.Error("Transaction layer doesn't see base unit")
	}
	if transactionWorld.NumUnits() != 1 {
		t.Errorf("Transaction layer should have 1 unit, got %d", transactionWorld.NumUnits())
	}

	// Add a unit to transaction layer
	transactionUnit := createTestUnit(3, 4, 2, 2)
	transactionWorld.AddUnit(transactionUnit)

	// Verify transaction layer has both units
	if transactionWorld.NumUnits() != 2 {
		t.Errorf("Transaction layer should have 2 units, got %d", transactionWorld.NumUnits())
	}

	// Verify base world still has only 1 unit
	if baseWorld.NumUnits() != 1 {
		t.Errorf("Base world should still have 1 unit, got %d", baseWorld.NumUnits())
	}
	if baseWorld.UnitAt(AxialCoord{Q: 3, R: 4}) != nil {
		t.Error("Base world should not see transaction unit")
	}

	// Test Pop operation
	parent := transactionWorld.Pop()
	if parent != baseWorld {
		t.Error("Pop() should return the parent world")
	}
}

func TestWorldTransactionIsolation(t *testing.T) {
	// Create base world with a unit
	units := []*v1.Unit{createTestUnit(1, 2, 1, 1)}
	baseWorld := createTestWorld("base", units, nil)

	// Create transaction layer
	transactionWorld := baseWorld.Push()

	// Move the base unit in transaction layer
	transactionUnit := transactionWorld.UnitAt(AxialCoord{Q: 1, R: 2})
	if transactionUnit == nil {
		t.Fatal("Transaction layer should see base unit")
	}

	err := transactionWorld.MoveUnit(transactionUnit, AxialCoord{Q: 5, R: 6})
	if err != nil {
		t.Fatalf("Failed to move unit in transaction: %v", err)
	}

	// Verify isolation: transaction sees moved unit, base sees original
	if transactionWorld.UnitAt(AxialCoord{Q: 1, R: 2}) != nil {
		t.Error("Transaction layer should not see unit at old position")
	}
	if transactionWorld.UnitAt(AxialCoord{Q: 5, R: 6}) == nil {
		t.Error("Transaction layer should see unit at new position")
	}

	// Base world should be unchanged
	if baseWorld.UnitAt(AxialCoord{Q: 1, R: 2}) == nil {
		t.Error("Base world should still see unit at original position")
	}
	if baseWorld.UnitAt(AxialCoord{Q: 5, R: 6}) != nil {
		t.Error("Base world should not see moved unit")
	}
}

func TestWorldUnitDeletion(t *testing.T) {
	// Create base world with a unit
	units := []*v1.Unit{createTestUnit(1, 2, 1, 1)}
	baseWorld := createTestWorld("base", units, nil)

	// Create transaction layer
	transactionWorld := baseWorld.Push()

	// Delete unit in transaction layer
	transactionUnit := transactionWorld.UnitAt(AxialCoord{Q: 1, R: 2})
	if transactionUnit == nil {
		t.Fatal("Transaction layer should see base unit")
	}

	err := transactionWorld.RemoveUnit(transactionUnit)
	if err != nil {
		t.Fatalf("Failed to remove unit in transaction: %v", err)
	}

	// Verify deletion isolation
	if transactionWorld.UnitAt(AxialCoord{Q: 1, R: 2}) != nil {
		t.Error("Transaction layer should not see deleted unit")
	}
	if transactionWorld.NumUnits() != 0 {
		t.Errorf("Transaction layer should have 0 units after deletion, got %d", transactionWorld.NumUnits())
	}

	// Base world should be unchanged
	if baseWorld.UnitAt(AxialCoord{Q: 1, R: 2}) == nil {
		t.Error("Base world should still see unit after transaction deletion")
	}
	if baseWorld.NumUnits() != 1 {
		t.Errorf("Base world should still have 1 unit, got %d", baseWorld.NumUnits())
	}
}

func TestWorldMergedIteration(t *testing.T) {
	// Create base world with units
	baseUnits := []*v1.Unit{
		createTestUnit(1, 2, 1, 1),
		createTestUnit(3, 4, 1, 1),
	}
	baseWorld := createTestWorld("base", baseUnits, nil)

	// Create transaction layer
	transactionWorld := baseWorld.Push()

	// Add a new unit in transaction
	transactionUnit := createTestUnit(5, 6, 2, 2)
	transactionWorld.AddUnit(transactionUnit)

	// Override a base unit in transaction
	overrideUnit := createTestUnit(1, 2, 1, 3)
	transactionWorld.AddUnit(overrideUnit)

	// Test merged iteration
	units := make(map[AxialCoord]*v1.Unit)
	for coord, unit := range transactionWorld.UnitsByCoord() {
		units[coord] = unit
	}

	// Should have 3 units total
	if len(units) != 3 {
		t.Errorf("Expected 3 units in merged iteration, got %d", len(units))
	}

	// Should have override unit, not base unit at (1,2)
	unit12 := units[AxialCoord{Q: 1, R: 2}]
	if unit12 == nil {
		t.Error("Missing unit at (1,2) in merged iteration")
	} else if unit12.UnitType != 3 {
		t.Errorf("Expected override unit (type 3) at (1,2), got type %d", unit12.UnitType)
	}

	// Should have base unit at (3,4)
	unit34 := units[AxialCoord{Q: 3, R: 4}]
	if unit34 == nil {
		t.Error("Missing base unit at (3,4) in merged iteration")
	}

	// Should have transaction unit at (5,6)
	unit56 := units[AxialCoord{Q: 5, R: 6}]
	if unit56 == nil {
		t.Error("Missing transaction unit at (5,6) in merged iteration")
	}
}

// Test the specific move duplication bug we found
func TestWorldMoveUnitNoDuplication(t *testing.T) {
	// Create world with one unit
	units := []*v1.Unit{createTestUnit(1, 2, 1, 1)}
	world := createTestWorld("test", units, nil)

	// Verify we have exactly 1 unit
	if world.NumUnits() != 1 {
		t.Fatalf("Expected 1 unit initially, got %d", world.NumUnits())
	}

	unit := world.UnitAt(AxialCoord{Q: 1, R: 2})
	if unit == nil {
		t.Fatal("Unit not found at initial position")
	}

	// Move the unit
	err := world.MoveUnit(unit, AxialCoord{Q: 3, R: 4})
	if err != nil {
		t.Fatalf("Failed to move unit: %v", err)
	}

	// Critical test: ensure we still have exactly 1 unit (no duplication)
	if world.NumUnits() != 1 {
		t.Errorf("Expected exactly 1 unit after move, got %d", world.NumUnits())

		// Debug: list all units if we have duplication
		for coord, u := range world.UnitsByCoord() {
			t.Logf("Found unit at (%d,%d) player=%d type=%d", coord.Q, coord.R, u.Player, u.UnitType)
		}
	}

	// Verify unit is only at new position
	if world.UnitAt(AxialCoord{Q: 1, R: 2}) != nil {
		t.Error("Unit still found at old position (duplication bug)")
	}
	if world.UnitAt(AxialCoord{Q: 3, R: 4}) == nil {
		t.Error("Unit not found at new position")
	}
}
