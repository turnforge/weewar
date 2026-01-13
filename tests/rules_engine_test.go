package tests

import (
	"math/rand"
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// var RULES_DATA_FILE = DevDataPath("assets/lilbattle-rules.json")
var RULES_DATA_FILE = "../assets/lilbattle-rules.json"
var DAMAGE_DATA_FILE = "../assets/lilbattle-damage.json"

func TestRulesEngineLoading(t *testing.T) {
	// Load rules from converted data
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
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
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Test getting unit data (Unit ID 1 should be Soldier Basic)
	unit, err := rulesEngine.GetUnitData(1)
	if err != nil {
		t.Fatalf("Failed to get unit data for ID 1: %v", err)
	}

	t.Logf("Unit 1: %s", unit.Name)
	t.Logf("  Health: %d, Movement: %f, Range: %d", unit.Health, unit.MovementPoints, unit.AttackRange)

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
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Test getting terrain data (ID 1 should exist)
	terrain, err := rulesEngine.GetTerrainData(1)
	if err != nil {
		t.Fatalf("Failed to get terrain data for ID 1: %v", err)
	}

	t.Logf("Terrain 1: %s", terrain.Name)

	if terrain.Name == "" {
		t.Error("Terrain name is empty")
	}

	// Test non-existent terrain
	_, err = rulesEngine.GetTerrainData(999)
	if err == nil {
		t.Error("Expected error for non-existent terrain ID 999")
	}
}

func TestRulesEngineMovementCosts(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Test terrain movement cost for unit 1 on terrain 1
	cost, err := rulesEngine.GetUnitTerrainCost(1, 1)
	if err != nil {
		t.Fatalf("Failed to get movement cost: %v", err)
	}

	t.Logf("Unit 1 movement cost on terrain 1: %.1f", cost)

	if cost <= 0 {
		t.Error("Movement cost should be positive")
	}

	// Test some other combinations
	testCases := []struct {
		unitID    int32
		terrainID int32
		desc      string
	}{
		{1, 2, "Soldier on terrain 2"},
		{2, 1, "Unit 2 on terrain 1"},
	}

	for _, tc := range testCases {
		cost, err := rulesEngine.GetUnitTerrainCost(tc.unitID, tc.terrainID)
		if err != nil {
			t.Logf("No movement cost data for %s: %v", tc.desc, err)
		} else {
			t.Logf("%s movement cost: %.1f", tc.desc, cost)
		}
	}
}

func TestRulesEngineCombatDamage(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Test combat prediction between unit 1 and unit 1 (if they can attack each other)
	damageDistribution, canAttack := rulesEngine.GetCombatPrediction(1, 1)
	if !canAttack {
		t.Logf("Units 1 vs 1 cannot attack each other")
		return // This is fine, not all units can attack all others
	}

	t.Logf("Combat 1 vs 1:")
	t.Logf("  Min/Max Damage: %f-%f", damageDistribution.MinDamage, damageDistribution.MaxDamage)
	t.Logf("  Expected Damage: %.1f", damageDistribution.ExpectedDamage)
	t.Logf("  Damage Ranges: %d", len(damageDistribution.Ranges))

	if damageDistribution.MinDamage < 0 {
		t.Error("Min damage should not be negative")
	}

	if damageDistribution.MaxDamage < damageDistribution.MinDamage {
		t.Error("Max damage should be >= min damage")
	}

	if len(damageDistribution.Ranges) == 0 {
		t.Error("Should have damage ranges")
	}

	// Test actual damage calculation with RNG
	rng := rand.New(rand.NewSource(42)) // Fixed seed for reproducible tests
	damage, canAttackCalc, err := rulesEngine.CalculateCombatDamage(1, 1, rng)
	if err != nil {
		t.Fatalf("Failed to calculate combat damage: %v", err)
	}
	if !canAttackCalc {
		t.Fatal("CalculateCombatDamage returned canAttack=false but GetCombatPrediction returned true")
	}

	t.Logf("Calculated damage: %d", damage)

	if float64(damage) < damageDistribution.MinDamage || float64(damage) > damageDistribution.MaxDamage {
		t.Errorf("Calculated damage %d outside expected range %.0f-%.0f",
			damage, damageDistribution.MinDamage, damageDistribution.MaxDamage)
	}
}

func TestRulesEngineAttackMatrix(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Count how many attack combinations we have using UnitUnitProperties
	totalAttacks := 0
	for key, props := range rulesEngine.UnitUnitProperties {
		if props.Damage != nil {
			totalAttacks++

			// Test one example in detail
			if totalAttacks == 1 {
				t.Logf("Example attack: %s has damage properties", key)
				t.Logf("  Damage range: %f-%f, Expected: %.1f",
					props.Damage.MinDamage, props.Damage.MaxDamage, props.Damage.ExpectedDamage)
			}
		}
	}

	t.Logf("Total attack combinations: %d", totalAttacks)

	if totalAttacks == 0 {
		t.Error("No attack combinations found in UnitUnitProperties")
	}
}

func TestRulesEngineMovementMatrix(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Check that TerrainUnitProperties is properly loaded (replaces MovementMatrix)
	if rulesEngine.TerrainUnitProperties == nil {
		t.Fatal("TerrainUnitProperties is nil - rules loading failed")
	}

	// Count how many movement cost entries we have using TerrainUnitProperties
	totalCosts := 0
	for key, props := range rulesEngine.TerrainUnitProperties {
		if props.MovementCost > 0 {
			totalCosts++

			// Test one example in detail
			if totalCosts == 1 {
				t.Logf("Example movement: %s has movement cost %.1f", key, props.MovementCost)
			}
		}
	}

	t.Logf("Total movement cost entries: %d", totalCosts)

	if totalCosts == 0 {
		t.Error("No movement cost entries found in TerrainUnitProperties")
	}
}

func TestRulesEngineDijkstraMovement(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create a world with tiles for the test
	protoWorld := &v1.WorldData{} // Empty world data for test
	world := NewWorld("test", protoWorld)

	// Fill with grass terrain (terrain ID 1 - should have reasonable movement cost)
	for q := range 5 {
		for r := range 5 {
			coord := AxialCoord{Q: q, R: r}
			tile := NewTile(coord, 1) // Grass terrain
			world.AddTile(tile)
		}
	}

	// Create a test unit (Soldier - unit type 1)
	startCoord := AxialCoord{Q: 2, R: 2} // Center of map
	unit := &v1.Unit{
		UnitType: 1,
		Q:        int32(startCoord.Q),
		R:        int32(startCoord.R),
		Player:   0,
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
		allPaths, err := rulesEngine.GetMovementOptions(world, unit, tc.movement, false)
		if err != nil {
			t.Fatalf("Failed to get movement options for %s: %v", tc.desc, err)
		}

		t.Logf("%s: %d tiles reachable", tc.desc, len(allPaths.Edges))

		// Verify all options are within budget and make sense
		for key, edge := range allPaths.Edges {
			if edge.TotalCost > float64(tc.movement) {
				t.Errorf("Option %s has cost %.1f > budget %d", key, edge.TotalCost, tc.movement)
			}

			if edge.TotalCost <= 0 {
				t.Errorf("Option %s has invalid cost %.1f", key, edge.TotalCost)
			}

			// Verify tile is adjacent to reachable area (basic sanity check)
			toCoord := AxialCoord{Q: int(edge.ToQ), R: int(edge.ToR)}
			distance := CubeDistance(startCoord, toCoord)
			if distance > tc.movement*2 { // Very generous upper bound
				t.Errorf("Option %v is suspiciously far (distance %d) for movement %d",
					toCoord, distance, tc.movement)
			}
		}

		// More movement should generally give more or equal options
		if tc.movement > 1 && len(allPaths.Edges) == 0 {
			t.Errorf("Expected some movement options for %s", tc.desc)
		}
	}
}

func TestRulesEngineDijkstraTerrainCosts(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create a world with different terrain costs for the test
	protoWorld := &v1.WorldData{} // Empty world data for test
	world := NewWorld("test", protoWorld)

	// Set up terrain: expensive terrain in middle, cheap around edges
	for q := range 3 {
		for r := range 3 {
			coord := AxialCoord{Q: q, R: r}
			terrainID := 1 // Default grass

			// Make center tile more expensive if we have different terrain types
			if q == 1 && r == 1 {
				// Try to find a more expensive terrain type
				for tID := range rulesEngine.Terrains {
					if _, err := rulesEngine.GetTerrainData(tID); err == nil {
						// Use a different terrain ID for center (ID 2 - Mountain perhaps?)
						if tID > 1 {
							terrainID = int(tID)
							break
						}
					}
				}
			}

			tile := NewTile(coord, terrainID)
			world.AddTile(tile)
		}
	}

	// Test unit at corner
	unit := &v1.Unit{
		UnitType: 1, // Soldier
		Q:        0,
		R:        0,
		Player:   0,
	}

	allPaths, err := rulesEngine.GetMovementOptions(world, unit, 3, false)
	if err != nil {
		t.Fatalf("Failed to get movement options: %v", err)
	}

	t.Logf("Movement options from corner: %d tiles", len(allPaths.Edges))

	// Log costs for debugging
	for key, edge := range allPaths.Edges {
		t.Logf("  Tile %s: cost %.1f", key, edge.TotalCost)
	}

	if len(allPaths.Edges) == 0 {
		t.Error("Expected some movement options")
	}
}

// TestPassThroughMovement tests that units can pass through occupied tiles when preventPassThrough=false
// Scenario: A -> B -> C where B is occupied
// With preventPassThrough=false: unit on A can reach C (passing through B)
// With preventPassThrough=true: unit on A cannot reach C (B blocks)
func TestPassThroughMovement(t *testing.T) {
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create a simple world with 3 tiles in a line: (0,0) -> (1,0) -> (2,0)
	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add grass tiles (terrain type 1) at positions (0,0), (1,0), (2,0)
	world.AddTile(&v1.Tile{Q: 0, R: 0, TileType: 1})
	world.AddTile(&v1.Tile{Q: 1, R: 0, TileType: 1})
	world.AddTile(&v1.Tile{Q: 2, R: 0, TileType: 1})

	// Unit at (0,0) - the one we're testing movement for
	movingUnit := &v1.Unit{
		UnitType:     1, // Soldier
		Q:            0,
		R:            0,
		Player:       1,
		DistanceLeft: 3,
	}
	world.AddUnit(movingUnit)

	// Blocking unit at (1,0)
	blockingUnit := &v1.Unit{
		UnitType: 1,
		Q:        1,
		R:        0,
		Player:   2,
	}
	world.AddUnit(blockingUnit)

	// Test 1: With preventPassThrough=false, unit should reach (2,0)
	allPathsPassThrough, err := rulesEngine.GetMovementOptions(world, movingUnit, 3, false)
	if err != nil {
		t.Fatalf("Failed to get movement options with pass-through: %v", err)
	}

	// Check if (2,0) is reachable
	key20 := "2,0"
	if _, exists := allPathsPassThrough.Edges[key20]; !exists {
		t.Errorf("With preventPassThrough=false, expected (2,0) to be reachable but it wasn't")
	} else {
		t.Logf("With preventPassThrough=false, (2,0) is reachable as expected")
	}

	// Check that (1,0) is marked as occupied (can pass through but not land)
	key10 := "1,0"
	if edge, exists := allPathsPassThrough.Edges[key10]; exists {
		if !edge.IsOccupied {
			t.Errorf("With preventPassThrough=false, expected (1,0) to be marked as occupied but IsOccupied=false")
		} else {
			t.Logf("With preventPassThrough=false, (1,0) correctly marked as occupied (pass-through only)")
		}
	} else {
		t.Errorf("With preventPassThrough=false, expected (1,0) to have an edge for path reconstruction")
	}

	// Test 2: With preventPassThrough=true, unit should NOT reach (2,0)
	allPathsNoPassThrough, err := rulesEngine.GetMovementOptions(world, movingUnit, 3, true)
	if err != nil {
		t.Fatalf("Failed to get movement options without pass-through: %v", err)
	}

	// Check that (2,0) is NOT reachable
	if _, exists := allPathsNoPassThrough.Edges[key20]; exists {
		t.Errorf("With preventPassThrough=true, expected (2,0) to NOT be reachable but it was")
	} else {
		t.Logf("With preventPassThrough=true, (2,0) correctly not reachable")
	}

	t.Logf("Pass-through test: with pass-through=%d destinations, without=%d destinations",
		len(allPathsPassThrough.Edges), len(allPathsNoPassThrough.Edges))
}
