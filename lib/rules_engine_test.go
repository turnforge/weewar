package weewar

import (
	"math/rand"
	"testing"
)

func TestRulesEngineLoading(t *testing.T) {
	// Load rules from converted data
	rulesEngine, err := LoadRulesEngineFromFile("../data/rules-data.json")
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Test basic counts
	unitCount := rulesEngine.GetLoadedUnitsCount()
	terrainCount := rulesEngine.GetLoadedTerrainsCount()

	t.Logf("Loaded rules: %d units, %d terrains", unitCount, terrainCount)

	if unitCount == 0 {
		t.Error("No units loaded")
	}

	if terrainCount == 0 {
		t.Error("No terrains loaded")
	}

	// Test validation
	if err := rulesEngine.ValidateRules(); err != nil {
		t.Errorf("Rules validation failed: %v", err)
	}
}

func TestRulesEngineUnitData(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile("../data/rules-data.json")
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Test getting unit data (Unit ID 1 should be Soldier Basic)
	unit, err := rulesEngine.GetUnitData(1)
	if err != nil {
		t.Fatalf("Failed to get unit data for ID 1: %v", err)
	}

	t.Logf("Unit 1: %s", unit.Name)
	t.Logf("  Health: %d, Movement: %d, Range: %d", unit.Health, unit.MovementPoints, unit.AttackRange)

	if unit.Name == "" {
		t.Error("Unit name is empty")
	}

	if unit.Health <= 0 {
		t.Error("Unit health should be positive")
	}

	if unit.MovementPoints <= 0 {
		t.Error("Unit movement points should be positive")
	}

	// Test non-existent unit
	_, err = rulesEngine.GetUnitData(999)
	if err == nil {
		t.Error("Expected error for non-existent unit ID 999")
	}
}

func TestRulesEngineTerrainData(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile("../data/rules-data.json")
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Test getting terrain data (ID 1 should exist)
	terrain, err := rulesEngine.GetTerrainData(1)
	if err != nil {
		t.Fatalf("Failed to get terrain data for ID 1: %v", err)
	}

	t.Logf("Terrain 1: %s (Base Cost: %.1f, Defense: %.1f)",
		terrain.Name, terrain.BaseMoveCost, terrain.DefenseBonus)

	if terrain.Name == "" {
		t.Error("Terrain name is empty")
	}

	if terrain.BaseMoveCost <= 0 {
		t.Error("Terrain base movement cost should be positive")
	}

	// Test non-existent terrain
	_, err = rulesEngine.GetTerrainData(999)
	if err == nil {
		t.Error("Expected error for non-existent terrain ID 999")
	}
}

func TestRulesEngineMovementCosts(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile("../data/rules-data.json")
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Test terrain movement cost for unit 1 on terrain 1
	cost, err := rulesEngine.GetTerrainMovementCost(1, 1)
	if err != nil {
		t.Fatalf("Failed to get movement cost: %v", err)
	}

	t.Logf("Unit 1 movement cost on terrain 1: %.1f", cost)

	if cost <= 0 {
		t.Error("Movement cost should be positive")
	}

	// Test some other combinations
	testCases := []struct {
		unitID    int
		terrainID int
		desc      string
	}{
		{1, 2, "Soldier on terrain 2"},
		{2, 1, "Unit 2 on terrain 1"},
	}

	for _, tc := range testCases {
		cost, err := rulesEngine.GetTerrainMovementCost(tc.unitID, tc.terrainID)
		if err != nil {
			t.Logf("No movement cost data for %s: %v", tc.desc, err)
		} else {
			t.Logf("%s movement cost: %.1f", tc.desc, cost)
		}
	}
}

func TestRulesEngineCombatDamage(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile("../data/rules-data.json")
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Test combat prediction between unit 1 and unit 1 (if they can attack each other)
	damageDistribution, err := rulesEngine.GetCombatPrediction(1, 1)
	if err != nil {
		t.Logf("Units 1 vs 1 cannot attack each other: %v", err)
		return // This is fine, not all units can attack all others
	}

	t.Logf("Combat 1 vs 1:")
	t.Logf("  Min/Max Damage: %d-%d", damageDistribution.MinDamage, damageDistribution.MaxDamage)
	t.Logf("  Expected Damage: %.1f", damageDistribution.ExpectedDamage)
	t.Logf("  Damage Buckets: %d", len(damageDistribution.DamageBuckets))

	if damageDistribution.MinDamage < 0 {
		t.Error("Min damage should not be negative")
	}

	if damageDistribution.MaxDamage < damageDistribution.MinDamage {
		t.Error("Max damage should be >= min damage")
	}

	if len(damageDistribution.DamageBuckets) == 0 {
		t.Error("Should have damage buckets")
	}

	// Test actual damage calculation with RNG
	rng := rand.New(rand.NewSource(42)) // Fixed seed for reproducible tests
	damage, err := rulesEngine.CalculateCombatDamage(1, 1, rng)
	if err != nil {
		t.Fatalf("Failed to calculate combat damage: %v", err)
	}

	t.Logf("Calculated damage: %d", damage)

	if damage < damageDistribution.MinDamage || damage > damageDistribution.MaxDamage {
		t.Errorf("Calculated damage %d outside expected range %d-%d",
			damage, damageDistribution.MinDamage, damageDistribution.MaxDamage)
	}
}

func TestRulesEngineAttackMatrix(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile("../data/rules-data.json")
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Count how many attack combinations we have
	totalAttacks := 0
	for attackerID, attacks := range rulesEngine.AttackMatrix.Attacks {
		for targetID := range attacks {
			totalAttacks++

			// Test one example in detail
			if totalAttacks == 1 {
				t.Logf("Example attack: Unit %d can attack Unit %d", attackerID, targetID)

				dist, err := rulesEngine.GetCombatPrediction(attackerID, targetID)
				if err != nil {
					t.Errorf("Failed to get prediction for valid attack: %v", err)
				} else {
					t.Logf("  Damage range: %d-%d, Expected: %.1f",
						dist.MinDamage, dist.MaxDamage, dist.ExpectedDamage)
				}
			}
		}
	}

	t.Logf("Total attack combinations: %d", totalAttacks)

	if totalAttacks == 0 {
		t.Error("No attack combinations found in attack matrix")
	}
}

func TestRulesEngineMovementMatrix(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile("../data/rules-data.json")
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Count how many movement cost entries we have
	totalCosts := 0
	for unitID, costs := range rulesEngine.MovementMatrix.Costs {
		for terrainID, cost := range costs {
			totalCosts++

			// Test one example in detail
			if totalCosts == 1 {
				t.Logf("Example movement: Unit %d on terrain %d costs %.1f", unitID, terrainID, cost)

				if cost <= 0 {
					t.Errorf("Movement cost should be positive, got %.1f", cost)
				}
			}
		}
	}

	t.Logf("Total movement cost entries: %d", totalCosts)

	if totalCosts == 0 {
		t.Error("No movement cost entries found in movement matrix")
	}
}

func TestRulesEngineDijkstraMovement(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile("../data/rules-data.json")
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create a simple test map
	gameMap := NewMapRect(5, 5)

	// Fill with grass terrain (terrain ID 1 - should have reasonable movement cost)
	for q := range 5 {
		for r := range 5 {
			coord := AxialCoord{Q: q, R: r}
			tile := NewTile(coord, 1) // Grass terrain
			gameMap.AddTile(tile)
		}
	}

	// Create a test unit (Soldier - unit type 1)
	startCoord := AxialCoord{Q: 2, R: 2} // Center of map
	unit := &Unit{
		UnitType: 1,
		Coord:    startCoord,
		PlayerID: 0,
	}

	// Test movement options with different movement budgets
	testCases := []struct {
		movement int
		desc     string
	}{
		{1, "1 movement point"},
		{2, "2 movement points"},
		{3, "3 movement points"},
		{5, "5 movement points"},
	}

	for _, tc := range testCases {
		options, err := rulesEngine.GetMovementOptions(gameMap, unit, tc.movement)
		if err != nil {
			t.Fatalf("Failed to get movement options for %s: %v", tc.desc, err)
		}

		t.Logf("%s: %d tiles reachable", tc.desc, len(options))

		// Verify all options are within budget and make sense
		for _, option := range options {
			if option.Cost > float64(tc.movement) {
				t.Errorf("Option %v has cost %.1f > budget %d", option.Coord, option.Cost, tc.movement)
			}

			if option.Cost <= 0 {
				t.Errorf("Option %v has invalid cost %.1f", option.Coord, option.Cost)
			}

			// Verify tile is adjacent to reachable area (basic sanity check)
			distance := CubeDistance(startCoord, option.Coord)
			if distance > tc.movement*2 { // Very generous upper bound
				t.Errorf("Option %v is suspiciously far (distance %d) for movement %d",
					option.Coord, distance, tc.movement)
			}
		}

		// More movement should generally give more or equal options
		if tc.movement > 1 && len(options) == 0 {
			t.Errorf("Expected some movement options for %s", tc.desc)
		}
	}
}

func TestRulesEngineDijkstraTerrainCosts(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile("../data/rules-data.json")
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create a map with different terrain costs
	gameMap := NewMapRect(3, 3)

	// Set up terrain: expensive terrain in middle, cheap around edges
	for q := 0; q < 3; q++ {
		for r := 0; r < 3; r++ {
			coord := AxialCoord{Q: q, R: r}
			terrainID := 1 // Default grass

			// Make center tile more expensive if we have different terrain types
			if q == 1 && r == 1 {
				// Try to find a more expensive terrain type
				for tID := range rulesEngine.Terrains {
					if terrain, err := rulesEngine.GetTerrainData(tID); err == nil {
						if terrain.BaseMoveCost > 1.5 { // More expensive than grass
							terrainID = tID
							break
						}
					}
				}
			}

			tile := NewTile(coord, terrainID)
			gameMap.AddTile(tile)
		}
	}

	// Test unit at corner
	unit := &Unit{
		UnitType: 1, // Soldier
		Coord:    AxialCoord{Q: 0, R: 0},
		PlayerID: 0,
	}

	options, err := rulesEngine.GetMovementOptions(gameMap, unit, 3)
	if err != nil {
		t.Fatalf("Failed to get movement options: %v", err)
	}

	t.Logf("Movement options from corner: %d tiles", len(options))

	// Log costs for debugging
	for _, option := range options {
		t.Logf("  Tile (%d,%d): cost %.1f", option.Coord.Q, option.Coord.R, option.Cost)
	}

	if len(options) == 0 {
		t.Error("Expected some movement options")
	}
}
