package lib

import (
	"fmt"
	"math"
	"math/rand"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
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

// SplashDamageTarget represents a unit that may receive splash damage
type SplashDamageTarget struct {
	Unit   *v1.Unit
	Damage int32
}

// =============================================================================
// Fix (Repair) Mechanics
// =============================================================================

// FixContext contains all information needed to calculate fix (repair) results
type FixContext struct {
	FixingUnit       *v1.Unit
	FixingUnitHealth int32 // Health of the unit performing the fix (Hf)
	FixValue         int32 // Fix value (F) of the fixing unit type
	InjuredUnit      *v1.Unit
}

// CalculateFixProbability calculates the fix probability (p) using the fix formula
// Formula: p = 0.05 * F
// Where F = fix value of the fixing unit
// The result is clamped to [0, 1]
func (re *RulesEngine) CalculateFixProbability(fixValue int32) float64 {
	p := 0.05 * float64(fixValue)

	// Clamp p to [0, 1]
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}

	return p
}

// SimulateFixHealing simulates the fix action and returns health restored
// For each health unit (Hf) of the fixing unit, 3 random numbers between 0 and 1 are generated
// Each time r < p, a fix is counted
// Total health restored = fixes / 3
func (re *RulesEngine) SimulateFixHealing(ctx *FixContext, rng *rand.Rand) int32 {
	p := re.CalculateFixProbability(ctx.FixValue)

	// Roll 3 dice for each health unit of the fixing unit
	fixes := 0.0

	for range ctx.FixingUnitHealth {
		for range 3 {
			roll := rng.Float64()
			if roll < p {
				fixes++
			}
		}
	}

	// Health restored = fixes / 3
	healthRestored := fixes / 3

	return int32(healthRestored)
}

// GenerateFixDistribution generates a distribution of possible healing outcomes
// by running many simulations
func (re *RulesEngine) GenerateFixDistribution(ctx *FixContext, numSimulations int) (*v1.DamageDistribution, error) {
	if numSimulations <= 0 {
		numSimulations = 10000 // Default number of simulations
	}

	// Run simulations
	simRng := rand.New(rand.NewSource(12345)) // Use fixed seed for reproducibility
	healingCounts := make(map[int32]int)
	totalHealing := float64(0)

	for range numSimulations {
		healing := re.SimulateFixHealing(ctx, simRng)
		healingCounts[healing]++
		totalHealing += float64(healing)
	}

	// Calculate expected healing
	expectedHealing := totalHealing / float64(numSimulations)

	// Calculate min and max healing observed
	var minHealing int32 = math.MaxInt32
	var maxHealing int32 = 0
	for healing := range healingCounts {
		if healing < minHealing {
			minHealing = healing
		}
		if healing > maxHealing {
			maxHealing = healing
		}
	}

	// Build healing ranges (reusing DamageDistribution/DamageRange structs)
	var ranges []*v1.DamageRange
	for healing := minHealing; healing <= maxHealing; healing++ {
		count := healingCounts[healing]
		if count > 0 {
			probability := float64(count) / float64(numSimulations)
			ranges = append(ranges, &v1.DamageRange{
				MinValue:    float64(healing),
				MaxValue:    float64(healing),
				Probability: probability,
			})
		}
	}

	return &v1.DamageDistribution{
		MinDamage:      float64(minHealing),
		MaxDamage:      float64(maxHealing),
		ExpectedDamage: expectedHealing,
		Ranges:         ranges,
	}, nil
}

// CalculateSplashDamage calculates splash damage for adjacent units
// Returns a list of units and the damage they receive
// Splash damage uses the same formula but without wound bonus (B = 0)
// Only deals damage if the calculated damage > 4
// Air units are immune to splash damage
func (re *RulesEngine) CalculateSplashDamage(
	attacker *v1.Unit,
	attackerTile *v1.Tile,
	defenderCoord AxialCoord,
	adjacentUnits []*v1.Unit,
	world *World,
	rng *rand.Rand,
) ([]*SplashDamageTarget, error) {
	// Get attacker definition
	attackerDef, err := re.GetUnitData(attacker.UnitType)
	if err != nil {
		return nil, fmt.Errorf("failed to get attacker data: %w", err)
	}

	// Check if attacker has splash damage
	if attackerDef.SplashDamage <= 0 {
		return nil, nil // No splash damage
	}

	var targets []*SplashDamageTarget

	for _, target := range adjacentUnits {
		// Get target unit definition
		targetDef, err := re.GetUnitData(target.UnitType)
		if err != nil {
			continue // Skip if we can't get unit data
		}

		// Air units are immune to splash damage
		if targetDef.UnitTerrain == "Air" {
			continue
		}

		// Get target tile
		targetCoord := UnitGetCoord(target)
		targetTile := world.TileAt(targetCoord)
		if targetTile == nil {
			continue
		}

		// Create combat context with NO wound bonus for splash
		ctx := &CombatContext{
			Attacker:       attacker,
			AttackerTile:   attackerTile,
			AttackerHealth: attacker.AvailableHealth,
			Defender:       target,
			DefenderTile:   targetTile,
			DefenderHealth: target.AvailableHealth,
			WoundBonus:     0, // No wound bonus for splash damage
		}

		// Run the formula splash_damage times
		totalDamage := int32(0)
		for i := int32(0); i < attackerDef.SplashDamage; i++ {
			damage, err := re.SimulateCombatDamage(ctx, rng)
			if err != nil {
				continue
			}
			totalDamage += damage
		}

		// Only apply splash damage if > 4
		if totalDamage > 4 {
			targets = append(targets, &SplashDamageTarget{
				Unit:   target,
				Damage: totalDamage,
			})
		}
	}

	return targets, nil
}
