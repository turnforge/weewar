package services

import (
	"fmt"
	"math/rand"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

// CalculateCombatDamage calculates damage using the new proto-based system
// Returns (damage, canAttack, error) where canAttack indicates if the attack is possible
func (re *RulesEngine) CalculateCombatDamage(attackerID, defenderID int32, rng *rand.Rand) (int, bool, error) {
	// Create key for unit-unit combat properties
	key := fmt.Sprintf("%d:%d", attackerID, defenderID)

	props, exists := re.UnitUnitProperties[key]
	if !exists || props.Damage == nil {
		// Attack is not possible between these unit types
		return 0, false, nil
	}

	damage := re.rollDamageFromDistribution(props.Damage, rng)
	return damage, true, nil
}

// GetCombatPrediction provides combat prediction using the new proto-based system
// Returns (damage_distribution, canAttack) where canAttack indicates if the attack is possible
func (re *RulesEngine) GetCombatPrediction(attackerID, defenderID int32) (*v1.DamageDistribution, bool) {
	// Create key for unit-unit combat properties
	key := fmt.Sprintf("%d:%d", attackerID, defenderID)

	props, exists := re.UnitUnitProperties[key]
	if !exists || props.Damage == nil {
		// Attack is not possible between these unit types
		return nil, false
	}

	return props.Damage, true
}

// rollDamageFromDistribution uses the proto damage distribution with ranges
func (re *RulesEngine) rollDamageFromDistribution(dist *v1.DamageDistribution, rng *rand.Rand) int {
	if dist == nil || len(dist.Ranges) == 0 {
		// Fall back to expected damage if no ranges defined
		return int(dist.ExpectedDamage)
	}

	// Use weighted random selection from damage ranges
	totalWeight := 0.0
	for _, damageRange := range dist.Ranges {
		totalWeight += damageRange.Probability
	}

	if totalWeight <= 0 {
		return int(dist.ExpectedDamage)
	}

	// Random selection based on probability weights
	randomValue := rng.Float64() * totalWeight
	cumulative := 0.0

	for _, damageRange := range dist.Ranges {
		cumulative += damageRange.Probability
		if randomValue <= cumulative {
			// Random damage within the range
			minDmg := damageRange.MinValue
			maxDmg := damageRange.MaxValue
			return int(minDmg + rng.Float64()*(maxDmg-minDmg))
		}
	}

	// Fallback (should not reach here)
	return int(dist.ExpectedDamage)
}

// GetAttackOptions returns all positions a unit can attack from its current position
// Only returns tiles with ENEMY units that are within attack range
func (re *RulesEngine) GetAttackOptions(world *World, unit *v1.Unit) ([]AxialCoord, error) {
	if unit == nil {
		return nil, fmt.Errorf("unit is nil")
	}

	unitData, err := re.GetUnitData(unit.UnitType)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit data: %w", err)
	}

	var attackPositions []AxialCoord
	attackRange := unitData.AttackRange

	// Check all positions within attack range
	unitCoord := UnitGetCoord(unit)
	for dQ := -attackRange; dQ <= attackRange; dQ++ {
		for dR := -attackRange; dR <= attackRange; dR++ {
			if dQ == 0 && dR == 0 {
				continue // Skip self
			}

			targetCoord := AxialCoord{Q: unitCoord.Q + int(dQ), R: unitCoord.R + int(dR)}

			// Check if there's an enemy unit at this position (attack rule: only enemy units)
			tile := world.TileAt(targetCoord)
			targetUnit := world.UnitAt(targetCoord)
			if tile == nil || targetUnit == nil {
				continue // No unit to attack
			}

			// Check if it's an enemy unit (different player)
			if targetUnit.Player == unit.Player {
				continue // Same player, can't attack
			}

			// Check if this unit can attack the target unit type
			if _, canAttack := re.GetCombatPrediction(unit.UnitType, targetUnit.UnitType); canAttack {
				attackPositions = append(attackPositions, targetCoord)
			}
		}
	}

	return attackPositions, nil
}

// CanUnitAttackTarget checks if a unit can attack a specific target
func (re *RulesEngine) CanUnitAttackTarget(attacker *v1.Unit, target *v1.Unit) (bool, error) {
	if attacker == nil || target == nil {
		return false, fmt.Errorf("attacker or target is nil")
	}

	// Check if units are enemies
	if attacker.Player == target.Player {
		return false, nil // Same team
	}

	// Check if attacker can attack this unit type
	_, canAttack := re.GetCombatPrediction(attacker.UnitType, target.UnitType)
	if !canAttack {
		return false, nil // Cannot attack this unit type
	}

	// Check range (using simple distance for now)
	attackerCoord := UnitGetCoord(attacker)
	targetCoord := UnitGetCoord(target)
	distance := CubeDistance(attackerCoord, targetCoord)
	unitData, err := re.GetUnitData(attacker.UnitType)
	if err != nil {
		return false, err
	}

	return distance <= int(unitData.AttackRange), nil
}
