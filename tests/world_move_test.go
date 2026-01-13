package tests

import (
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// Test basic MoveUnit behavior on root world (no transactions)
func TestMoveUnitBasic(t *testing.T) {
	world := NewWorld("test", nil)

	// Add a unit at (1,2)
	unit := CreateTestUnit(1, 2, 1, 1)
	_, err := world.AddUnit(unit)
	if err != nil {
		t.Fatalf("Failed to add unit: %v", err)
	}

	// Verify initial state
	if world.NumUnits() != 1 {
		t.Fatalf("Expected 1 unit initially, got %d", world.NumUnits())
	}
	if world.UnitAt(AxialCoord{Q: 1, R: 2}) == nil {
		t.Fatal("Unit not found at initial position (1,2)")
	}

	// Move unit to (3,4)
	newCoord := AxialCoord{Q: 3, R: 4}
	err = world.MoveUnit(unit, newCoord)
	if err != nil {
		t.Fatalf("Failed to move unit: %v", err)
	}

	// CRITICAL TESTS: Verify no duplication
	if world.NumUnits() != 1 {
		t.Errorf("DUPLICATION BUG: Expected 1 unit after move, got %d", world.NumUnits())

		// Debug: list all units
		for coord, u := range world.UnitsByCoord() {
			t.Logf("  Unit at (%d,%d) player=%d", coord.Q, coord.R, u.Player)
		}
	}

	// Verify unit is at new position only
	if world.UnitAt(AxialCoord{Q: 1, R: 2}) != nil {
		t.Error("Unit still found at old position (1,2)")
	}
	if world.UnitAt(AxialCoord{Q: 3, R: 4}) == nil {
		t.Error("Unit not found at new position (3,4)")
	}

	// Verify unit coordinates were updated
	if unit.Q != 3 || unit.R != 4 {
		t.Errorf("Unit coordinates not updated: Q=%d, R=%d", unit.Q, unit.R)
	}
}

// Test MoveUnit with existing unit at destination (replacement)
func TestMoveUnitReplacement(t *testing.T) {
	world := NewWorld("test", nil)

	// Add two units
	unit1 := CreateTestUnit(1, 2, 1, 1)
	unit2 := CreateTestUnit(3, 4, 2, 1)

	world.AddUnit(unit1)
	world.AddUnit(unit2)

	// Verify initial state: 2 units
	if world.NumUnits() != 2 {
		t.Fatalf("Expected 2 units initially, got %d", world.NumUnits())
	}

	// Move unit1 to unit2's position (should replace unit2)
	err := world.MoveUnit(unit1, AxialCoord{Q: 3, R: 4})
	if err != nil {
		t.Fatalf("Failed to move unit: %v", err)
	}

	// CRITICAL TEST: Should have 1 unit (unit1 moved to replace unit2)
	finalCount := world.NumUnits()
	if finalCount != 1 {
		t.Errorf("Expected 1 unit after replacement move, got %d", finalCount)

		// Debug: list all units
		for coord, u := range world.UnitsByCoord() {
			t.Logf("  Unit at (%d,%d) player=%d", coord.Q, coord.R, u.Player)
		}
	}

	// Verify positions
	if world.UnitAt(AxialCoord{Q: 1, R: 2}) != nil {
		t.Error("Unit still found at old position (1,2)")
	}
	movedUnit := world.UnitAt(AxialCoord{Q: 3, R: 4})
	if movedUnit == nil {
		t.Error("No unit found at destination (3,4)")
	} else if movedUnit.Player != 1 {
		t.Errorf("Wrong unit at destination: expected player 1, got player %d", movedUnit.Player)
	}
}

// Test AddUnit behavior when replacing existing unit
func TestAddUnitReplacement(t *testing.T) {
	world := NewWorld("test", nil)

	// Add initial unit
	unit1 := CreateTestUnit(1, 2, 1, 1)
	oldUnit, err := world.AddUnit(unit1)
	if err != nil {
		t.Fatalf("Failed to add first unit: %v", err)
	}
	if oldUnit != nil {
		t.Error("Expected no old unit for first add")
	}

	// Verify initial state
	if world.NumUnits() != 1 {
		t.Fatalf("Expected 1 unit after first add, got %d", world.NumUnits())
	}

	// Add second unit at same position (should replace)
	unit2 := CreateTestUnit(1, 2, 2, 1)
	oldUnit, err = world.AddUnit(unit2)
	if err != nil {
		t.Fatalf("Failed to add replacement unit: %v", err)
	}
	if oldUnit != unit1 {
		t.Error("Expected old unit to be returned")
	}

	// CRITICAL TEST: Should still have exactly 1 unit
	if world.NumUnits() != 1 {
		t.Errorf("DUPLICATION BUG: Expected 1 unit after replacement, got %d", world.NumUnits())

		// Debug: list all units
		for coord, u := range world.UnitsByCoord() {
			t.Logf("  Unit at (%d,%d) player=%d", coord.Q, coord.R, u.Player)
		}
	}

	// Verify the correct unit is there
	finalUnit := world.UnitAt(AxialCoord{Q: 1, R: 2})
	if finalUnit == nil {
		t.Fatal("No unit found at position")
	}
	if finalUnit.Player != 2 {
		t.Errorf("Wrong unit: expected player 2, got player %d", finalUnit.Player)
	}
}

// Test MoveUnit in transaction layer
func TestMoveUnitTransaction(t *testing.T) {
	// Create base world with unit
	baseWorld := NewWorld("base", nil)
	unit := CreateTestUnit(1, 2, 1, 1)
	baseWorld.AddUnit(unit)

	// Create transaction layer
	transactionWorld := baseWorld.Push()

	// Verify transaction sees base unit
	if transactionWorld.NumUnits() != 1 {
		t.Fatalf("Transaction should see 1 unit from base, got %d", transactionWorld.NumUnits())
	}

	// Move unit in transaction
	transactionUnit := transactionWorld.UnitAt(AxialCoord{Q: 1, R: 2})
	if transactionUnit == nil {
		t.Fatal("Transaction layer should see base unit")
	}

	err := transactionWorld.MoveUnit(transactionUnit, AxialCoord{Q: 3, R: 4})
	if err != nil {
		t.Fatalf("Failed to move unit in transaction: %v", err)
	}

	// CRITICAL TESTS: Transaction layer
	if transactionWorld.NumUnits() != 1 {
		t.Errorf("Transaction should have 1 unit after move, got %d", transactionWorld.NumUnits())
	}
	if transactionWorld.UnitAt(AxialCoord{Q: 1, R: 2}) != nil {
		t.Error("Transaction should not see unit at old position")
	}
	if transactionWorld.UnitAt(AxialCoord{Q: 3, R: 4}) == nil {
		t.Error("Transaction should see unit at new position")
	}

	// CRITICAL TESTS: Base world should be unchanged
	if baseWorld.NumUnits() != 1 {
		t.Errorf("Base world should still have 1 unit, got %d", baseWorld.NumUnits())
	}
	if baseWorld.UnitAt(AxialCoord{Q: 1, R: 2}) == nil {
		t.Error("Base world should still see unit at original position")
	}
	if baseWorld.UnitAt(AxialCoord{Q: 3, R: 4}) != nil {
		t.Error("Base world should not see moved unit")
	}
}

// Test RemoveUnit and AddUnit sequence (what MoveUnit does internally)
func TestRemoveAddSequence(t *testing.T) {
	world := NewWorld("test", nil)

	// Add unit
	unit := CreateTestUnit(1, 2, 1, 1)
	world.AddUnit(unit)

	// Verify initial state
	if world.NumUnits() != 1 {
		t.Fatalf("Expected 1 unit initially, got %d", world.NumUnits())
	}

	// Remove unit
	err := world.RemoveUnit(unit)
	if err != nil {
		t.Fatalf("Failed to remove unit: %v", err)
	}

	// Verify removal
	if world.NumUnits() != 0 {
		t.Errorf("Expected 0 units after removal, got %d", world.NumUnits())
	}
	if world.UnitAt(AxialCoord{Q: 1, R: 2}) != nil {
		t.Error("Unit still found after removal")
	}

	// Update unit position and re-add
	UnitSetCoord(unit, AxialCoord{Q: 3, R: 4})
	_, err = world.AddUnit(unit)
	if err != nil {
		t.Fatalf("Failed to re-add unit: %v", err)
	}

	// CRITICAL TEST: Should have exactly 1 unit at new position
	if world.NumUnits() != 1 {
		t.Errorf("DUPLICATION BUG: Expected 1 unit after re-add, got %d", world.NumUnits())

		// Debug: list all units
		for coord, u := range world.UnitsByCoord() {
			t.Logf("  Unit at (%d,%d) player=%d", coord.Q, coord.R, u.Player)
		}
	}

	// Verify unit is at new position only
	if world.UnitAt(AxialCoord{Q: 1, R: 2}) != nil {
		t.Error("Unit found at old position after move sequence")
	}
	if world.UnitAt(AxialCoord{Q: 3, R: 4}) == nil {
		t.Error("Unit not found at new position after move sequence")
	}
}

// Test the exact scenario from ProcessMoves integration test
func TestMoveUnitExactProcessMovesScenario(t *testing.T) {
	world := NewWorld("test", nil)

	// Create unit at (1,2) exactly like ProcessMoves test
	unit := &v1.Unit{
		Q:               1,
		R:               2,
		Player:          1,
		UnitType:        1,
		AvailableHealth: 10,
		DistanceLeft:    3,
		// TurnCounter removed: Units will be lazily topped-up when accessed
	}

	// Add unit to world
	world.AddUnit(unit)

	t.Logf("Before MoveUnit:")
	t.Logf("  NumUnits: %d", world.NumUnits())
	for coord, u := range world.UnitsByCoord() {
		t.Logf("  Unit at (%d,%d) player=%d", coord.Q, coord.R, u.Player)
	}

	// Move unit from (1,2) to (1,1) - exactly like ProcessMoves test
	newCoord := AxialCoord{Q: 1, R: 1}
	err := world.MoveUnit(unit, newCoord)
	if err != nil {
		t.Fatalf("MoveUnit failed: %v", err)
	}

	t.Logf("After MoveUnit:")
	t.Logf("  NumUnits: %d", world.NumUnits())
	for coord, u := range world.UnitsByCoord() {
		t.Logf("  Unit at (%d,%d) player=%d", coord.Q, coord.R, u.Player)
	}

	// CRITICAL TESTS: Same as ProcessMoves test expects
	if world.NumUnits() != 1 {
		t.Errorf("DUPLICATION BUG: Expected 1 unit after move, got %d", world.NumUnits())
	}

	// Check specific positions
	oldPos := world.UnitAt(AxialCoord{Q: 1, R: 2})
	newPos := world.UnitAt(AxialCoord{Q: 1, R: 1})

	if oldPos != nil {
		t.Error("Unit still found at old position (1,2)")
	}
	if newPos == nil {
		t.Error("Unit not found at new position (1,1)")
	}

	// Verify unit coordinates were updated
	if unit.Q != 1 || unit.R != 1 {
		t.Errorf("Unit coordinates not updated: Q=%d, R=%d", unit.Q, unit.R)
	}
}

// Test the exact transaction flow from ProcessMoves
func TestProcessMovesTransactionFlow(t *testing.T) {
	// Step 1: Create original world with unit (simulates runtime game state)
	originalWorld := NewWorld("test", nil)
	unit := &v1.Unit{
		Q:               1,
		R:               2,
		Player:          1,
		UnitType:        1,
		AvailableHealth: 10,
		DistanceLeft:    3,
		// TurnCounter removed: Units will be lazily topped-up when accessed
	}
	originalWorld.AddUnit(unit)

	t.Logf("  NumUnits: %d", originalWorld.NumUnits())
	for coord, u := range originalWorld.UnitsByCoord() {
		t.Logf("  Unit at (%d,%d) player=%d", coord.Q, coord.R, u.Player)
	}

	// Step 2: Create transaction snapshot (simulates ProcessMoves transaction)
	transactionWorld := originalWorld.Push()

	t.Logf("Transaction world created:")
	t.Logf("  NumUnits: %d", transactionWorld.NumUnits())

	// Step 3: Do move processing on transaction layer (simulates move processor)
	transactionUnit := transactionWorld.UnitAt(AxialCoord{Q: 1, R: 2})
	if transactionUnit == nil {
		t.Fatal("Unit not found in transaction layer")
	}

	// Update unit state (simulates move processor changes)
	transactionUnit.DistanceLeft = 2
	transactionUnit.LastActedTurn = 2

	err := transactionWorld.MoveUnit(transactionUnit, AxialCoord{Q: 1, R: 1})
	if err != nil {
		t.Fatalf("Transaction move failed: %v", err)
	}

	t.Logf("After transaction move:")
	t.Logf("  Transaction NumUnits: %d", transactionWorld.NumUnits())
	t.Logf("  Original NumUnits: %d", originalWorld.NumUnits())

	// Step 4: Roll back to original world (simulates ApplyChangeResults rollback)
	currentWorld := originalWorld // Switch back to original world

	t.Logf("After rollback to original:")
	t.Logf("  NumUnits: %d", currentWorld.NumUnits())
	for coord, u := range currentWorld.UnitsByCoord() {
		t.Logf("  Unit at (%d,%d) player=%d distanceLeft=%f unitCoords=(%d,%d)", coord.Q, coord.R, u.Player, u.DistanceLeft, u.Q, u.R)
	}

	// Step 5: Apply changes to original world (simulates applyUnitMoved)
	unitToMove := currentWorld.UnitAt(AxialCoord{Q: 1, R: 2})
	if unitToMove == nil {
		t.Fatal("Unit not found in original world after rollback")
	}

	// Update unit state from change (simulates applyUnitMoved updates)
	unitToMove.DistanceLeft = 2
	unitToMove.LastActedTurn = 2

	err = currentWorld.MoveUnit(unitToMove, AxialCoord{Q: 1, R: 1})
	if err != nil {
		t.Fatalf("Final move failed: %v", err)
	}

	t.Logf("After final move application:")
	t.Logf("  NumUnits: %d", currentWorld.NumUnits())
	for coord, u := range currentWorld.UnitsByCoord() {
		t.Logf("  Unit at (%d,%d) player=%d distanceLeft=%f", coord.Q, coord.R, u.Player, u.DistanceLeft)
	}

	// CRITICAL TESTS: Should have 1 unit at new position
	if currentWorld.NumUnits() != 1 {
		t.Errorf("DUPLICATION BUG: Expected 1 unit after transaction flow, got %d", currentWorld.NumUnits())
	}

	if currentWorld.UnitAt(AxialCoord{Q: 1, R: 2}) != nil {
		t.Error("Unit still found at old position (1,2) after transaction flow")
	}
	if currentWorld.UnitAt(AxialCoord{Q: 1, R: 1}) == nil {
		t.Error("Unit not found at new position (1,1) after transaction flow")
	}
}
