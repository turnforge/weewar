package lib

import (
	"testing"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
)

func TestCalculatePlayerBaseIncome(t *testing.T) {
	tests := []struct {
		name         string
		playerId     int32
		worldData    *v1.WorldData
		incomeConfig *v1.IncomeConfig
		expected     int32
	}{
		{
			name:     "empty world returns zero",
			playerId: 1,
			worldData: &v1.WorldData{
				TilesMap: map[string]*v1.Tile{},
			},
			incomeConfig: nil,
			expected:     0,
		},
		{
			name:     "player with one landbase gets default landbase income",
			playerId: 1,
			worldData: &v1.WorldData{
				TilesMap: map[string]*v1.Tile{
					"0,0": {Q: 0, R: 0, TileType: TileTypeLandBase, Player: 1},
				},
			},
			incomeConfig: nil,
			expected:     DefaultLandbaseIncome, // 100
		},
		{
			name:     "player with multiple bases gets sum of income",
			playerId: 1,
			worldData: &v1.WorldData{
				TilesMap: map[string]*v1.Tile{
					"0,0": {Q: 0, R: 0, TileType: TileTypeLandBase, Player: 1},
					"1,0": {Q: 1, R: 0, TileType: TileTypeNavalBase, Player: 1},
					"2,0": {Q: 2, R: 0, TileType: TileTypeAirport, Player: 1},
				},
			},
			incomeConfig: nil,
			expected:     DefaultLandbaseIncome + DefaultNavalbaseIncome + DefaultAirportbaseIncome, // 100 + 150 + 200 = 450
		},
		{
			name:     "tiles owned by other players are not counted",
			playerId: 1,
			worldData: &v1.WorldData{
				TilesMap: map[string]*v1.Tile{
					"0,0": {Q: 0, R: 0, TileType: TileTypeLandBase, Player: 1},
					"1,0": {Q: 1, R: 0, TileType: TileTypeLandBase, Player: 2},
				},
			},
			incomeConfig: nil,
			expected:     DefaultLandbaseIncome, // 100 (only player 1's base)
		},
		{
			name:     "custom income config overrides defaults",
			playerId: 1,
			worldData: &v1.WorldData{
				TilesMap: map[string]*v1.Tile{
					"0,0": {Q: 0, R: 0, TileType: TileTypeLandBase, Player: 1},
				},
			},
			incomeConfig: &v1.IncomeConfig{
				LandbaseIncome: 200,
			},
			expected: 200,
		},
		{
			name:     "game income is added to total",
			playerId: 1,
			worldData: &v1.WorldData{
				TilesMap: map[string]*v1.Tile{
					"0,0": {Q: 0, R: 0, TileType: TileTypeLandBase, Player: 1},
				},
			},
			incomeConfig: &v1.IncomeConfig{
				LandbaseIncome: 100,
				GameIncome:     50,
			},
			expected: 150, // 100 from base + 50 game income
		},
		{
			name:     "mines income with custom config",
			playerId: 1,
			worldData: &v1.WorldData{
				TilesMap: map[string]*v1.Tile{
					"0,0": {Q: 0, R: 0, TileType: TileTypeMines, Player: 1},
				},
			},
			incomeConfig: &v1.IncomeConfig{
				MinesIncome: 1000,
			},
			expected: 1000,
		},
		{
			name:     "missile silo income",
			playerId: 1,
			worldData: &v1.WorldData{
				TilesMap: map[string]*v1.Tile{
					"0,0": {Q: 0, R: 0, TileType: TileTypeMissileSilo, Player: 1},
				},
			},
			incomeConfig: nil,
			expected:     DefaultMissilesiloIncome, // 300
		},
		{
			name:     "non-income tiles are not counted",
			playerId: 1,
			worldData: &v1.WorldData{
				TilesMap: map[string]*v1.Tile{
					"0,0": {Q: 0, R: 0, TileType: 5, Player: 1},  // Plains or other non-income tile
					"1,0": {Q: 1, R: 0, TileType: 6, Player: 1},  // Another non-income tile
					"2,0": {Q: 2, R: 0, TileType: TileTypeLandBase, Player: 1},
				},
			},
			incomeConfig: nil,
			expected:     DefaultLandbaseIncome, // Only the landbase counts
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePlayerBaseIncome(tt.playerId, tt.worldData, tt.incomeConfig)
			if result != tt.expected {
				t.Errorf("CalculatePlayerBaseIncome() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestGetTileIncomeFromConfig(t *testing.T) {
	tests := []struct {
		name         string
		tileType     int32
		incomeConfig *v1.IncomeConfig
		expected     int32
	}{
		{
			name:         "nil config uses default for landbase",
			tileType:     TileTypeLandBase,
			incomeConfig: nil,
			expected:     DefaultLandbaseIncome,
		},
		{
			name:         "nil config uses default for naval base",
			tileType:     TileTypeNavalBase,
			incomeConfig: nil,
			expected:     DefaultNavalbaseIncome,
		},
		{
			name:         "nil config uses default for airport",
			tileType:     TileTypeAirport,
			incomeConfig: nil,
			expected:     DefaultAirportbaseIncome,
		},
		{
			name:         "nil config uses default for missile silo",
			tileType:     TileTypeMissileSilo,
			incomeConfig: nil,
			expected:     DefaultMissilesiloIncome,
		},
		{
			name:         "nil config uses default for mines",
			tileType:     TileTypeMines,
			incomeConfig: nil,
			expected:     DefaultMinesIncome,
		},
		{
			name:     "custom landbase income",
			tileType: TileTypeLandBase,
			incomeConfig: &v1.IncomeConfig{
				LandbaseIncome: 250,
			},
			expected: 250,
		},
		{
			name:     "zero income config value falls back to default",
			tileType: TileTypeLandBase,
			incomeConfig: &v1.IncomeConfig{
				LandbaseIncome: 0,
			},
			expected: DefaultLandbaseIncome,
		},
		{
			name:         "unknown tile type returns zero",
			tileType:     99,
			incomeConfig: nil,
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTileIncomeFromConfig(tt.tileType, tt.incomeConfig)
			if result != tt.expected {
				t.Errorf("GetTileIncomeFromConfig() = %d, want %d", result, tt.expected)
			}
		})
	}
}
