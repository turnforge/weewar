package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	weewar "github.com/panyam/turnengine/games/weewar/lib"
)

// LegacyWeewarData represents the original weewar-data.json format
type LegacyWeewarData struct {
	Units    []LegacyUnit    `json:"units"`
	Terrains []LegacyTerrain `json:"terrains"`
}

// LegacyUnit represents unit data in original format
type LegacyUnit struct {
	ID              int                         `json:"id"`
	Name            string                      `json:"name"`
	TerrainMovement map[string]float64          `json:"terrainMovement"`
	AttackMatrix    map[string]LegacyDamageInfo `json:"attackMatrix"`
	Health          int                         `json:"health,omitempty"`
	MovementPoints  int                         `json:"movementPoints,omitempty"`
	AttackRange     int                         `json:"attackRange,omitempty"`
	Cost            int                         `json:"cost,omitempty"`
	Attack          int                         `json:"attack,omitempty"`
	Defense         int                         `json:"defense,omitempty"`
	SightRange      int                         `json:"sightRange,omitempty"`
	CanCapture      bool                        `json:"canCapture,omitempty"`
	BaseStats       *LegacyBaseStats            `json:"baseStats,omitempty"`
}

// LegacyBaseStats represents unit base stats in original format
type LegacyBaseStats struct {
	Cost       int  `json:"cost"`
	Health     int  `json:"health"`
	Movement   int  `json:"movement"`
	Attack     int  `json:"attack"`
	Defense    int  `json:"defense"`
	SightRange int  `json:"sightRange"`
	CanCapture bool `json:"canCapture"`
}

// LegacyTerrain represents terrain data in original format
type LegacyTerrain struct {
	ID           int                `json:"id"`
	Name         string             `json:"name"`
	MovementCost map[string]float64 `json:"movementCost"`
	DefenseBonus float64            `json:"defenseBonus,omitempty"`
}

// LegacyDamageInfo represents damage info in original format
type LegacyDamageInfo struct {
	MinDamage     int                `json:"minDamage"`
	MaxDamage     int                `json:"maxDamage"`
	Probabilities map[string]float64 `json:"probabilities"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: weewar-convert <input-file> [output-file]")
	}

	inputFile := os.Args[1]
	outputFile := "rules-data.json"
	if len(os.Args) >= 3 {
		outputFile = os.Args[2]
	}

	fmt.Printf("Converting %s to canonical format...\n", inputFile)

	// Read legacy data
	legacyData, err := loadLegacyData(inputFile)
	if err != nil {
		log.Fatalf("Failed to load legacy data: %v", err)
	}

	// Convert to canonical format
	rulesEngine, err := convertToCanonical(legacyData)
	if err != nil {
		log.Fatalf("Failed to convert data: %v", err)
	}

	// Save canonical format
	if err := saveCanonicalData(rulesEngine, outputFile); err != nil {
		log.Fatalf("Failed to save canonical data: %v", err)
	}

	fmt.Printf("Successfully converted to %s\n", outputFile)
	fmt.Printf("Units: %d, Terrains: %d\n", len(rulesEngine.Units), len(rulesEngine.Terrains))
}

func loadLegacyData(filename string) (*LegacyWeewarData, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var legacyData LegacyWeewarData
	if err := json.Unmarshal(data, &legacyData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &legacyData, nil
}

func convertToCanonical(legacy *LegacyWeewarData) (*weewar.RulesEngine, error) {
	rulesEngine := weewar.NewRulesEngine()

	// Create terrain name to ID mapping
	terrainNameToID := make(map[string]int)
	for _, terrain := range legacy.Terrains {
		terrainNameToID[terrain.Name] = terrain.ID
	}

	// Create unit name to ID mapping
	unitNameToID := make(map[string]int)
	for _, unit := range legacy.Units {
		unitNameToID[unit.Name] = unit.ID
	}

	// Convert units
	for _, legacyUnit := range legacy.Units {
		// Use baseStats values if top-level fields are zero
		health := legacyUnit.Health
		movement := legacyUnit.MovementPoints
		attackRange := legacyUnit.AttackRange
		
		// Check baseStats if available
		if health == 0 && legacyUnit.BaseStats != nil {
			health = legacyUnit.BaseStats.Health
		}
		if movement == 0 && legacyUnit.BaseStats != nil {
			movement = legacyUnit.BaseStats.Movement  
		}
		if movement == 0 {
			movement = 3 // Default movement points for units without data
		}
		if attackRange == 0 && legacyUnit.AttackRange == 0 {
			attackRange = 1 // Default attack range
		}

		unitData := &weewar.UnitData{
			ID:             legacyUnit.ID,
			Name:           legacyUnit.Name,
			Health:         health,
			MovementPoints: movement,
			AttackRange:    attackRange,
			BaseStats: weewar.UnitStats{
				Cost:       legacyUnit.Cost,
				Health:     health,
				Movement:   movement,
				Attack:     legacyUnit.Attack,
				Defense:    legacyUnit.Defense,
				SightRange: legacyUnit.SightRange,
				CanCapture: legacyUnit.CanCapture,
			},
		}

		rulesEngine.Units[legacyUnit.ID] = unitData

		// Convert terrain movement costs to movement matrix
		if rulesEngine.MovementMatrix.Costs[legacyUnit.ID] == nil {
			rulesEngine.MovementMatrix.Costs[legacyUnit.ID] = make(map[int]float64)
		}

		for terrainName, cost := range legacyUnit.TerrainMovement {
			if terrainID, exists := terrainNameToID[terrainName]; exists {
				rulesEngine.MovementMatrix.Costs[legacyUnit.ID][terrainID] = cost
			}
		}

		// Convert attack matrix
		if rulesEngine.AttackMatrix.Attacks[legacyUnit.ID] == nil {
			rulesEngine.AttackMatrix.Attacks[legacyUnit.ID] = make(map[int]*weewar.DamageDistribution)
		}

		for targetName, damageInfo := range legacyUnit.AttackMatrix {
			if targetID, exists := unitNameToID[targetName]; exists {
				// Convert string-keyed probabilities to damage buckets
				var damageBuckets []weewar.DamageBucket
				expectedDamage := 0.0

				for damageStr, probability := range damageInfo.Probabilities {
					if probability > 0 { // Only include non-zero probabilities
						damage, err := strconv.Atoi(damageStr)
						if err != nil {
							continue // Skip invalid damage values
						}
						damageBuckets = append(damageBuckets, weewar.DamageBucket{
							Damage: damage,
							Weight: probability,
						})
						expectedDamage += float64(damage) * probability
					}
				}

				damageDistribution := &weewar.DamageDistribution{
					MinDamage:      damageInfo.MinDamage,
					MaxDamage:      damageInfo.MaxDamage,
					DamageBuckets:  damageBuckets,
					ExpectedDamage: expectedDamage,
				}

				rulesEngine.AttackMatrix.Attacks[legacyUnit.ID][targetID] = damageDistribution
			}
		}
	}

	// Convert terrains
	for _, legacyTerrain := range legacy.Terrains {
		// Calculate base movement cost as average or use a default
		baseMoveCost := 1.0
		if len(legacyTerrain.MovementCost) > 0 {
			total := 0.0
			count := 0
			for _, cost := range legacyTerrain.MovementCost {
				total += cost
				count++
			}
			baseMoveCost = total / float64(count)
		}

		terrainData := &weewar.TerrainData{
			ID:           legacyTerrain.ID,
			Name:         legacyTerrain.Name,
			BaseMoveCost: baseMoveCost,
			DefenseBonus: legacyTerrain.DefenseBonus,
			Type:         weewar.TerrainNature, // Default to nature terrain
		}

		rulesEngine.Terrains[legacyTerrain.ID] = terrainData
	}

	return rulesEngine, nil
}

func saveCanonicalData(rulesEngine *weewar.RulesEngine, filename string) error {
	// Create output directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Serialize rules engine to JSON
	data, err := json.MarshalIndent(rulesEngine, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal rules engine: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
