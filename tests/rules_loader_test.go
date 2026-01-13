package tests

import (
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
)

// TestSetDefaultIncomeValues verifies that default income values are set correctly for terrain types
func TestSetDefaultIncomeValues(t *testing.T) {
	// Create a minimal RulesEngine with test terrains
	re := &lib.RulesEngine{
		RulesEngine: &v1.RulesEngine{
			Terrains: map[int32]*v1.TerrainDefinition{
				1:  {Id: 1, Name: "Base", BuildableUnitIds: []int32{1, 2}},
				2:  {Id: 2, Name: "Harbor", BuildableUnitIds: []int32{3}},
				3:  {Id: 3, Name: "Airport", BuildableUnitIds: []int32{4}},
				16: {Id: 16, Name: "Missile Silo", BuildableUnitIds: []int32{5}},
				20: {Id: 20, Name: "Mines", BuildableUnitIds: []int32{6}},
				// Non-income generating terrain
				10: {Id: 10, Name: "Grass", BuildableUnitIds: []int32{}},
			},
		},
	}

	// Call setDefaultIncomeValues
	lib.SetDefaultIncomeValues(re)

	// Test cases
	testCases := []struct {
		tileID         int32
		expectedIncome int32
		description    string
	}{
		{1, 100, "Land Base should have income of 100"},
		{2, 150, "Naval Base should have income of 150"},
		{3, 200, "Airport should have income of 200"},
		{16, 300, "Missile Silo should have income of 300"},
		{20, 500, "Mines should have income of 500"},
		{10, 0, "Grass should have no income (not in DefaultIncomeMap)"},
	}

	for _, tc := range testCases {
		terrain := re.Terrains[tc.tileID]
		if terrain.IncomePerTurn != tc.expectedIncome {
			t.Errorf("%s: got %d, want %d", tc.description, terrain.IncomePerTurn, tc.expectedIncome)
		}
	}
}

// TestLoadRulesEngineIncomeValues verifies that loaded rules have income values set
func TestLoadRulesEngineIncomeValues(t *testing.T) {
	rulesEngine, err := lib.LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Verify income values are set for terrains in DefaultIncomeMap
	for tileID, expectedIncome := range lib.DefaultIncomeMap {
		terrain, err := rulesEngine.GetTerrainData(tileID)
		if err != nil {
			t.Errorf("Failed to get terrain %d: %v", tileID, err)
			continue
		}

		if terrain.IncomePerTurn != expectedIncome {
			t.Errorf("Terrain %d (%s) has income %d, expected %d",
				tileID, terrain.Name, terrain.IncomePerTurn, expectedIncome)
		}
	}

	// Verify non-income terrains have zero income
	// Test a few known non-income terrain IDs
	nonIncomeTiles := []int32{4, 5, 6, 7, 8, 9, 10} // Grass, water, etc.
	for _, tileID := range nonIncomeTiles {
		terrain, err := rulesEngine.GetTerrainData(tileID)
		if err != nil {
			// Skip if terrain doesn't exist
			continue
		}

		if terrain.IncomePerTurn != 0 {
			t.Errorf("Non-income terrain %d (%s) should have income 0, got %d",
				tileID, terrain.Name, terrain.IncomePerTurn)
		}
	}
}
