package lib

import (
	"fmt"
	"math/rand"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
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
// Uses proper hex distance calculation and checks unit-to-unit combat compatibility
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

	// Get all coordinates within attack range using proper hex distance
	unitCoord := UnitGetCoord(unit)
	coordsInRange := unitCoord.Range(int(attackRange))

	// Check each coordinate for valid attack targets
	for _, targetCoord := range coordsInRange {
		// Skip self
		if targetCoord.Q == unitCoord.Q && targetCoord.R == unitCoord.R {
			continue
		}

		// Check if there's an enemy unit at this position
		targetUnit := world.UnitAt(targetCoord)
		if targetUnit == nil {
			continue // No unit to attack
		}

		// Check if it's an enemy unit (different player)
		if targetUnit.Player == unit.Player {
			continue // Same player, can't attack
		}

		// Check if this unit can attack the target unit type (handles compatibility)
		if _, canAttack := re.GetCombatPrediction(unit.UnitType, targetUnit.UnitType); canAttack {
			attackPositions = append(attackPositions, targetCoord)
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

// GetFixOptions returns all adjacent friendly units that can be fixed by this unit
func (re *RulesEngine) GetFixOptions(world *World, fixer *v1.Unit) ([]AxialCoord, error) {
	if fixer == nil {
		return nil, fmt.Errorf("fixer unit is nil")
	}

	// Check if this unit can fix
	fixerData, err := re.GetUnitData(fixer.UnitType)
	if err != nil {
		return nil, fmt.Errorf("failed to get fixer data: %w", err)
	}

	if fixerData.FixValue <= 0 {
		return nil, nil // Unit cannot fix
	}

	var fixPositions []AxialCoord
	fixerCoord := UnitGetCoord(fixer)

	// Get all adjacent hexes (distance 1)
	var neighbors [6]AxialCoord
	fixerCoord.Neighbors(&neighbors)

	for _, neighborCoord := range neighbors {
		// Check if there's a friendly unit at this position
		targetUnit := world.UnitAt(neighborCoord)
		if targetUnit == nil {
			continue // No unit to fix
		}

		// Check if it's a friendly unit (same player)
		if targetUnit.Player != fixer.Player {
			continue // Enemy unit, can't fix
		}

		// Check if target is damaged
		targetData, err := re.GetUnitData(targetUnit.UnitType)
		if err != nil {
			continue // Can't get target data
		}
		if targetUnit.AvailableHealth >= targetData.Health {
			continue // Already at max health
		}

		// Check if fixer can fix this target type (terrain compatibility)
		canFix, _ := re.CanUnitFixTarget(fixer, targetUnit)
		if canFix {
			fixPositions = append(fixPositions, neighborCoord)
		}
	}

	return fixPositions, nil
}

// CanUnitFixTarget checks if a fixer unit can fix a specific target unit
// based on unit terrain type compatibility
func (re *RulesEngine) CanUnitFixTarget(fixer *v1.Unit, target *v1.Unit) (bool, error) {
	if fixer == nil || target == nil {
		return false, fmt.Errorf("fixer or target is nil")
	}

	// Check if fixer has fix ability
	fixerData, err := re.GetUnitData(fixer.UnitType)
	if err != nil {
		return false, err
	}
	if fixerData.FixValue <= 0 {
		return false, nil // Cannot fix
	}

	// Check if they are on the same team
	if fixer.Player != target.Player {
		return false, nil // Can only fix friendly units
	}

	// Check terrain compatibility using DefaultFixTargets
	allowedTerrains, hasRestriction := DefaultFixTargets[fixer.UnitType]
	if !hasRestriction {
		// No specific restriction, check if fixer terrain matches target terrain
		targetData, err := re.GetUnitData(target.UnitType)
		if err != nil {
			return false, err
		}
		return fixerData.UnitTerrain == targetData.UnitTerrain, nil
	}

	// Check if target's terrain is in the allowed list
	targetData, err := re.GetUnitData(target.UnitType)
	if err != nil {
		return false, err
	}

	for _, terrain := range allowedTerrains {
		if targetData.UnitTerrain == terrain {
			return true, nil
		}
	}

	return false, nil
}
