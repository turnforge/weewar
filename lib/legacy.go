package weewar

import (
	"encoding/json"
	"os"
	"sort"
	"strconv"
)

// =============================================================================
// Legacy Import Structures (for weewar-data.json import only)
// =============================================================================

// LegacyGameData represents the weewar-data.json format for import
type LegacyGameData struct {
	Units    []LegacyUnitData    `json:"units"`
	Terrains []LegacyTerrainData `json:"terrains"`
}

// LegacyUnitData represents unit data from weewar-data.json
type LegacyUnitData struct {
	ID              int                         `json:"id"`
	Name            string                      `json:"name"`
	TerrainMovement map[string]float64          `json:"terrainMovement"`
	AttackMatrix    map[string]LegacyAttackData `json:"attackMatrix"`
}

// LegacyTerrainData represents terrain data from weewar-data.json
type LegacyTerrainData struct {
	ID           int                `json:"id"`
	Name         string             `json:"name"`
	MovementCost map[string]float64 `json:"movementCost"`
	DefenseBonus float64            `json:"defenseBonus"`
	Properties   []string           `json:"properties"`
}

// LegacyAttackData represents attack data from weewar-data.json
type LegacyAttackData struct {
	MinDamage     int                `json:"minDamage"`
	MaxDamage     int                `json:"maxDamage"`
	Probabilities map[string]float64 `json:"probabilities"`
}

// importFromLegacy imports and converts weewar-data.json
func (re *RulesEngine) importFromLegacy(legacyPath string) error {
	data, err := os.ReadFile(legacyPath)
	if err != nil {
		return err
	}

	legacyData := &LegacyGameData{}
	if err := json.Unmarshal(data, legacyData); err != nil {
		return err
	}

	return re.convertLegacyData(legacyData)
}

// =============================================================================
// Legacy Data Conversion
// =============================================================================

// convertLegacyData converts weewar-data.json to canonical format
func (re *RulesEngine) convertLegacyData(legacyData *LegacyGameData) error {
	// Create name-to-ID mappings for lookup
	terrainNameToID := make(map[string]int)
	unitNameToID := make(map[string]int)

	// Convert units to existing UnitData format
	for _, legacyUnit := range legacyData.Units {
		unit := &UnitData{
			ID:             legacyUnit.ID,
			Name:           legacyUnit.Name,
			MovementPoints: 3,   // Default, will be enhanced later
			AttackRange:    1,   // Default, will be enhanced later
			Health:         100, // Default, will be enhanced later
			Properties:     []string{},
		}
		re.Units[legacyUnit.ID] = unit
		unitNameToID[legacyUnit.Name] = legacyUnit.ID
	}

	// Convert terrains to existing TerrainData format
	for _, legacyTerrain := range legacyData.Terrains {
		terrainType := TerrainNature // Default
		if contains(legacyTerrain.Properties, "base") || contains(legacyTerrain.Properties, "city") {
			terrainType = TerrainPlayer
		}

		terrain := &TerrainData{
			ID:           legacyTerrain.ID,
			Name:         legacyTerrain.Name,
			DefenseBonus: legacyTerrain.DefenseBonus,
			Type:         terrainType,
			Properties:   legacyTerrain.Properties,
		}
		re.Terrains[legacyTerrain.ID] = terrain
		terrainNameToID[legacyTerrain.Name] = legacyTerrain.ID
	}

	// Convert movement matrix using IDs
	for _, legacyUnit := range legacyData.Units {
		unitID := legacyUnit.ID
		re.MovementMatrix.Costs[unitID] = make(map[int]float64)

		for terrainName, cost := range legacyUnit.TerrainMovement {
			if terrainID, exists := terrainNameToID[terrainName]; exists {
				re.MovementMatrix.Costs[unitID][terrainID] = cost
			}
		}
	}

	// Convert attack matrix using IDs and canonical DamageDistribution
	for _, legacyUnit := range legacyData.Units {
		attackerID := legacyUnit.ID
		re.AttackMatrix.Attacks[attackerID] = make(map[int]*DamageDistribution)

		for defenderName, legacyAttack := range legacyUnit.AttackMatrix {
			if defenderID, exists := unitNameToID[defenderName]; exists {
				damageDist := re.convertLegacyAttackData(legacyAttack)
				re.AttackMatrix.Attacks[attackerID][defenderID] = damageDist
			}
		}
	}
	return nil
}

// convertLegacyAttackData converts legacy probability format to canonical DamageDistribution
func (re *RulesEngine) convertLegacyAttackData(legacyAttack LegacyAttackData) *DamageDistribution {
	// Convert string-keyed probabilities to sorted damage buckets
	var buckets []DamageBucket
	expectedDamage := 0.0

	// Sort damage values for consistent ordering
	var damages []int
	for damageStr := range legacyAttack.Probabilities {
		damage, err := strconv.Atoi(damageStr)
		if err == nil && legacyAttack.Probabilities[damageStr] > 0 {
			damages = append(damages, damage)
		}
	}
	sort.Ints(damages)

	// Create buckets in sorted order
	for _, damage := range damages {
		weight := legacyAttack.Probabilities[strconv.Itoa(damage)]
		if weight > 0 {
			buckets = append(buckets, DamageBucket{
				Damage: damage,
				Weight: weight,
			})
			expectedDamage += float64(damage) * weight
		}
	}

	return &DamageDistribution{
		MinDamage:      legacyAttack.MinDamage,
		MaxDamage:      legacyAttack.MaxDamage,
		DamageBuckets:  buckets,
		ExpectedDamage: expectedDamage,
	}
}
