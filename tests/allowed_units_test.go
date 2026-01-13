package tests

import (
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
)

// createTestGameWithAllowedUnits creates a minimal game for testing allowed units
func createTestGameWithAllowedUnits(allowedUnits []int32) *lib.Game {
	worldData := &v1.WorldData{
		TilesMap: map[string]*v1.Tile{
			"0,0": {Q: 0, R: 0, TileType: lib.TileTypeLandBase, Player: 1}, // Player 1's base
		},
		UnitsMap: map[string]*v1.Unit{},
	}

	game := &v1.Game{
		Id:      "test-game",
		WorldId: "test-world",
		Config: &v1.GameConfiguration{
			Players: []*v1.GamePlayer{
				{PlayerId: 1, StartingCoins: 1000}, // Enough coins to build anything
			},
			Settings: &v1.GameSettings{
				AllowedUnits: allowedUnits,
			},
		},
	}

	state := &v1.GameState{
		GameId:        "test-game",
		CurrentPlayer: 1,
		TurnCounter:   1,
		WorldData:     worldData,
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 1000, IsActive: true},
		},
	}

	// Use default rules engine
	rulesEngine := lib.DefaultRulesEngine()

	return lib.NewGame(game, state, lib.NewWorld("test-world", worldData), rulesEngine, 0)
}

func TestFilterBuildOptionsByAllowedUnits(t *testing.T) {
	tests := []struct {
		name              string
		buildableUnits    []int32
		allowedUnits      []int32
		expectedBuildable []int32
	}{
		{
			name:              "all units allowed - all buildable units shown",
			buildableUnits:    []int32{1, 2, 3, 4, 5},
			allowedUnits:      []int32{1, 2, 3, 4, 5},
			expectedBuildable: []int32{1, 2, 3, 4, 5},
		},
		{
			name:              "only units 1 and 3 allowed",
			buildableUnits:    []int32{1, 2, 3, 4, 5},
			allowedUnits:      []int32{1, 3},
			expectedBuildable: []int32{1, 3},
		},
		{
			name:              "no units allowed - empty build options",
			buildableUnits:    []int32{1, 2, 3, 4, 5},
			allowedUnits:      []int32{},
			expectedBuildable: []int32{},
		},
		{
			name:              "allowed units not in buildable list - empty build options",
			buildableUnits:    []int32{1, 2, 3, 4, 5},
			allowedUnits:      []int32{99, 100}, // Units that base can't build
			expectedBuildable: []int32{},
		},
		{
			name:              "nil allowed units - all buildable units shown (no restriction)",
			buildableUnits:    []int32{1, 2, 3, 4, 5},
			allowedUnits:      nil, // No restrictions
			expectedBuildable: []int32{1, 2, 3, 4, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildOptions := lib.FilterBuildOptionsByAllowedUnits(
				tt.buildableUnits,
				tt.allowedUnits,
			)

			if len(buildOptions) != len(tt.expectedBuildable) {
				t.Errorf("got %d build options, want %d", len(buildOptions), len(tt.expectedBuildable))
			}

			for _, expected := range tt.expectedBuildable {
				found := false
				for _, actual := range buildOptions {
					if actual == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected unit %d in build options, not found", expected)
				}
			}
		})
	}
}

func TestProcessBuildUnit_RejectsDisallowedUnits(t *testing.T) {
	tests := []struct {
		name         string
		allowedUnits []int32
		buildUnit    int32
		expectError  bool
	}{
		{
			name:         "building allowed unit succeeds",
			allowedUnits: []int32{1, 2, 3},
			buildUnit:    1,
			expectError:  false,
		},
		{
			name:         "building disallowed unit fails",
			allowedUnits: []int32{1, 2, 3},
			buildUnit:    5, // Not in allowed list
			expectError:  true,
		},
		{
			name:         "empty allowed list - all builds fail",
			allowedUnits: []int32{},
			buildUnit:    1,
			expectError:  true,
		},
		{
			name:         "nil allowed list - no restriction",
			allowedUnits: nil,
			buildUnit:    1,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := createTestGameWithAllowedUnits(tt.allowedUnits)

			move := &v1.GameMove{}
			action := &v1.BuildUnitAction{
				Pos:      &v1.Position{Q: 0, R: 0},
				UnitType: tt.buildUnit,
			}

			err := game.ProcessBuildUnit(move, action)

			if tt.expectError && err == nil {
				t.Errorf("expected error for building unit %d with allowed units %v, got nil", tt.buildUnit, tt.allowedUnits)
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error for building unit %d with allowed units %v, got: %v", tt.buildUnit, tt.allowedUnits, err)
			}
		})
	}
}
