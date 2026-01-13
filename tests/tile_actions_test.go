package tests

import (
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// TestGetAllowedActionsForTile tests the tile action state machine
func TestGetAllowedActionsForTile(t *testing.T) {
	rulesEngine := DefaultRulesEngine()

	tests := []struct {
		name        string
		tile        *v1.Tile
		terrainDef  *v1.TerrainDefinition
		playerCoins int32
		expected    []string
		description string
	}{
		{
			name: "Tile with buildable units and sufficient money - build action",
			tile: &v1.Tile{
				Q:        0,
				R:        0,
				TileType: 1,
				Player:   1,
			},
			terrainDef: &v1.TerrainDefinition{
				Id:               1,
				Name:             "Land Base",
				BuildableUnitIds: []int32{1, 2, 3}, // Trooper (75), Heavy Trooper (150), Sniper (250)
			},
			playerCoins: 100, // Can afford Trooper (75)
			expected:    []string{"build"},
			description: "Tile with affordable units should allow build action",
		},
		{
			name: "Tile with buildable units but no money - no build action",
			tile: &v1.Tile{
				Q:        0,
				R:        0,
				TileType: 1,
				Player:   1,
			},
			terrainDef: &v1.TerrainDefinition{
				Id:               1,
				Name:             "Land Base",
				BuildableUnitIds: []int32{1, 2, 3}, // Trooper (75), Heavy Trooper (150), Sniper (250)
			},
			playerCoins: 0, // Cannot afford any units
			expected:    []string{},
			description: "Tile with no affordable units should provide no build action",
		},
		{
			name: "Tile with no buildable units - no actions",
			tile: &v1.Tile{
				Q:        0,
				R:        0,
				TileType: 10,
				Player:   1,
			},
			terrainDef: &v1.TerrainDefinition{
				Id:               10,
				Name:             "Grass",
				BuildableUnitIds: []int32{}, // No buildable units
			},
			playerCoins: 1000,
			expected:    []string{},
			description: "Tile without buildable units should provide no actions",
		},
		{
			name: "Tile with nil terrain def - no actions",
			tile: &v1.Tile{
				Q:        0,
				R:        0,
				TileType: 1,
				Player:   1,
			},
			terrainDef:  nil,
			playerCoins: 1000,
			expected:    []string{},
			description: "Tile with nil terrain definition should provide no actions",
		},
		{
			name: "Tile with expensive units only - no build action",
			tile: &v1.Tile{
				Q:        0,
				R:        0,
				TileType: 1,
				Player:   1,
			},
			terrainDef: &v1.TerrainDefinition{
				Id:               1,
				Name:             "Land Base",
				BuildableUnitIds: []int32{2, 3}, // Heavy Trooper (150), Sniper (250)
			},
			playerCoins: 100, // Cannot afford cheapest unit (150)
			expected:    []string{},
			description: "Tile with only expensive units should provide no build action when player is poor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := rulesEngine.GetAllowedActionsForTile(tt.tile, tt.terrainDef, tt.playerCoins)

			// Check length
			if len(allowed) != len(tt.expected) {
				t.Errorf("%s: expected %d actions, got %d\nExpected: %v\nGot: %v",
					tt.description, len(tt.expected), len(allowed), tt.expected, allowed)
				return
			}

			// Check each action
			for i, expectedAction := range tt.expected {
				if allowed[i] != expectedAction {
					t.Errorf("%s: action[%d] = %s, want %s",
						tt.description, i, allowed[i], expectedAction)
				}
			}

			t.Logf("✅ %s: %v", tt.description, allowed)
		})
	}
}

// TestGetAllowedActionsForTile_BuildWithMoney tests build action with player money constraints
// TODO: Implement once player balance field is added to GamePlayer proto
func TestGetAllowedActionsForTile_BuildWithMoney(t *testing.T) {
	t.Skip("Skipping until player balance field is added to GamePlayer proto")

	// Future test cases:
	// - Player has enough money for all buildable units
	// - Player has enough money for some buildable units
	// - Player has no money - no build actions
	// - Player has exactly enough money for cheapest unit
}
