package services

import (
	"fmt"
	"math"
	"math/rand"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

// CombatContext contains all information needed to calculate combat damage using the formula
type CombatContext struct {
	Attacker       *v1.Unit
	AttackerTile   *v1.Tile
	AttackerHealth int32
	Defender       *v1.Unit
	DefenderTile   *v1.Tile
	DefenderHealth int32
	WoundBonus     int32 // B in the formula
}

// CalculateHitProbability calculates the hit probability (p) using the attack formula
// Formula: p = 0.05 * ( ( ( A + Ta ) - ( D + Td ) ) + B ) + 0.5
// Where:
//
//	A = Attack value of attacking unit
//	Ta = Terrain attack bonus for attacker
//	D = Defense value of defending unit
//	Td = Terrain defense bonus for defender
//	B = Wound bonus
func (re *RulesEngine) CalculateHitProbability(ctx *CombatContext) (float64, error) {
	// Get attacker unit definition
	attackerDef, err := re.GetUnitData(ctx.Attacker.UnitType)
	if err != nil {
		return 0, fmt.Errorf("failed to get attacker data: %w", err)
	}

	// Get defender unit definition
	defenderDef, err := re.GetUnitData(ctx.Defender.UnitType)
	if err != nil {
		return 0, fmt.Errorf("failed to get defender data: %w", err)
	}

	// Get base attack value (A) from attack_vs_class table
	attackKey := fmt.Sprintf("%s:%s", defenderDef.UnitClass, defenderDef.UnitTerrain)
	baseAttack, hasAttack := attackerDef.AttackVsClass[attackKey]
	if !hasAttack {
		// Cannot attack this unit type
		return 0, fmt.Errorf("unit %s cannot attack %s:%s", attackerDef.Name, defenderDef.UnitClass, defenderDef.UnitTerrain)
	}

	// Get terrain bonuses
	attackerTerrainProps := re.GetTerrainUnitPropertiesForUnit(ctx.AttackerTile.TileType, ctx.Attacker.UnitType)
	defenderTerrainProps := re.GetTerrainUnitPropertiesForUnit(ctx.DefenderTile.TileType, ctx.Defender.UnitType)

	var Ta int32 = 0 // Terrain attack bonus for attacker
	var Td int32 = 0 // Terrain defense bonus for defender

	if attackerTerrainProps != nil {
		Ta = attackerTerrainProps.AttackBonus
	}
	if defenderTerrainProps != nil {
		Td = defenderTerrainProps.DefenseBonus
	}

	// Get base defense (D)
	D := defenderDef.Defense

	// Calculate using the formula
	// p = 0.05 * ( ( ( A + Ta ) - ( D + Td ) ) + B ) + 0.5
	A := float64(baseAttack)
	p := 0.05*(((A+float64(Ta))-(float64(D)+float64(Td)))+float64(ctx.WoundBonus)) + 0.5

	// Clamp p to [0, 1]
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}

	return p, nil
}

// SimulateCombatDamage simulates combat damage by rolling dice according to the formula
// For each health unit (Ha) of the attacker, roll 6 dice
// In WeeWar, each health unit = 10 HP, so 100 HP = 10 health units
// Each die roll that's < p counts as a hit
// Total damage = hits / 6
func (re *RulesEngine) SimulateCombatDamage(ctx *CombatContext, rng *rand.Rand) (int32, error) {
	p, err := re.CalculateHitProbability(ctx)
	if err != nil {
		return 0, err
	}

	// Roll 6 dice for each health unit of the attacker
	hits := 0.0

	for range ctx.AttackerHealth {
		for range 6 {
			roll := rng.Float64()
			if roll < p {
				hits++
			}
		}
	}

	// Damage = hits / 6
	damage := hits / 6

	// Cap damage at attacker's health (cannot deal more damage than you have health)
	if damage > float64(ctx.AttackerHealth) {
		return ctx.AttackerHealth, nil
	}

	return int32(damage), nil
}

// GenerateDamageDistribution generates a damage distribution by running many simulations
// This is useful for UI tooltips showing expected damage ranges
func (re *RulesEngine) GenerateDamageDistribution(ctx *CombatContext, numSimulations int) (*v1.DamageDistribution, error) {
	if numSimulations <= 0 {
		numSimulations = 10000 // Default number of simulations
	}

	// Validate hit probability calculation works
	_, err := re.CalculateHitProbability(ctx)
	if err != nil {
		return nil, err
	}

	// Run simulations
	simRng := rand.New(rand.NewSource(12345)) // Use fixed seed for reproducibility
	damageCounts := make(map[int32]int)
	totalDamage := float64(0)

	for range numSimulations {
		damage, err := re.SimulateCombatDamage(ctx, simRng)
		if err != nil {
			return nil, err
		}
		damageCounts[damage]++
		totalDamage += float64(damage)
	}

	// Calculate expected damage
	expectedDamage := totalDamage / float64(numSimulations)

	// Calculate min and max damage observed
	var minDamage int32 = math.MaxInt32
	var maxDamage int32 = 0
	for damage := range damageCounts {
		if damage < minDamage {
			minDamage = damage
		}
		if damage > maxDamage {
			maxDamage = damage
		}
	}

	// Build damage ranges
	var ranges []*v1.DamageRange
	for damage := minDamage; damage <= maxDamage; damage++ {
		count := damageCounts[damage]
		if count > 0 {
			probability := float64(count) / float64(numSimulations)
			ranges = append(ranges, &v1.DamageRange{
				MinValue:    float64(damage),
				MaxValue:    float64(damage),
				Probability: probability,
			})
		}
	}

	return &v1.DamageDistribution{
		MinDamage:      float64(minDamage),
		MaxDamage:      float64(maxDamage),
		ExpectedDamage: expectedDamage,
		Ranges:         ranges,
	}, nil
}

// GetTerrainUnitPropertiesForUnit is a helper to get terrain-unit properties
func (re *RulesEngine) GetTerrainUnitPropertiesForUnit(terrainID, unitID int32) *v1.TerrainUnitProperties {
	key := fmt.Sprintf("%d:%d", terrainID, unitID)
	return re.TerrainUnitProperties[key]
}

// CalculateWoundBonus calculates the wound bonus (B) based on attack history
// Returns the wound bonus value to add to the attack formula
func (re *RulesEngine) CalculateWoundBonus(defender *v1.Unit, attackerCoord AxialCoord) int32 {
	if len(defender.AttackHistory) == 0 {
		return 0
	}

	defenderCoord := UnitGetCoord(defender)
	woundBonus := int32(0)

	// Check if current attacker is ranged (2+ tiles away)
	currentDistance := CubeDistance(attackerCoord, defenderCoord)
	currentIsRanged := currentDistance >= 2

	for _, attack := range defender.AttackHistory {
		prevAttackerCoord := AxialCoord{Q: int(attack.Q), R: int(attack.R)}

		if currentIsRanged {
			// If current attacker is ranged: +1 for each previous attack
			woundBonus++
		} else {
			// Current attacker is adjacent (distance 1)
			if attack.IsRanged {
				// +1 for each previous ranged attack
				woundBonus++
			} else {
				// Previous attack was also adjacent
				// Calculate geometric relationship
				// +1: from hex adjacent to both attacker and defender
				// +2: from any other adjacent hex
				// +3: from opposite side

				// Check if previous attacker is adjacent to current attacker
				prevToCurrent := CubeDistance(prevAttackerCoord, attackerCoord)

				if prevToCurrent == 1 {
					// Adjacent to current attacker
					woundBonus += 1
				} else if re.isOppositeSide(defenderCoord, prevAttackerCoord, attackerCoord) {
					// Opposite side
					woundBonus += 3
				} else {
					// Any other adjacent hex
					woundBonus += 2
				}
			}
		}
	}

	return woundBonus
}

// isOppositeSide checks if two positions are on opposite sides of a defender
func (re *RulesEngine) isOppositeSide(defender, pos1, pos2 AxialCoord) bool {
	// Calculate vectors from defender to each position
	vec1Q := pos1.Q - defender.Q
	vec1R := pos1.R - defender.R

	vec2Q := pos2.Q - defender.Q
	vec2R := pos2.R - defender.R

	// In hex coordinates, opposite sides have opposite direction vectors
	// Check if vectors point in opposite directions
	return vec1Q == -vec2Q && vec1R == -vec2R
}
