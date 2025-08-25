package weewar

import (
	"fmt"
	"math/rand"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

// AttackMatrix defines combat outcomes between unit types using IDs
type AttackMatrix struct {
	// attacks[attackerID][defenderID] = damage distribution
	Attacks map[int32]map[int32]*DamageDistribution `json:"attacks"`
}

// CalculateCombatDamage calculates damage using canonical DamageDistribution
func (re *RulesEngine) CalculateCombatDamage(attackerID, defenderID int32, rng *rand.Rand) (int, error) {
	attackerAttacks, exists := re.AttackMatrix.Attacks[attackerID]
	if !exists {
		return 0, fmt.Errorf("unit ID %d cannot attack", attackerID)
	}

	damageDist, exists := attackerAttacks[defenderID]
	if !exists {
		return 0, fmt.Errorf("unit ID %d cannot attack unit ID %d", attackerID, defenderID)
	}

	return re.rollDamageFromBuckets(damageDist.DamageBuckets, rng), nil
}

// GetCombatPrediction provides combat prediction using existing types
func (re *RulesEngine) GetCombatPrediction(attackerID, defenderID int32) (*DamageDistribution, error) {
	attackerAttacks, exists := re.AttackMatrix.Attacks[attackerID]
	if !exists {
		return nil, fmt.Errorf("unit ID %d cannot attack", attackerID)
	}

	damageDist, exists := attackerAttacks[defenderID]
	if !exists {
		return nil, fmt.Errorf("unit ID %d cannot attack unit ID %d", attackerID, defenderID)
	}

	return damageDist, nil
}

// rollDamageFromBuckets uses weighted random selection
func (re *RulesEngine) rollDamageFromBuckets(buckets []DamageBucket, rng *rand.Rand) int {
	if len(buckets) == 0 {
		return 0
	}

	// Calculate total weight
	totalWeight := 0.0
	for _, bucket := range buckets {
		totalWeight += bucket.Weight
	}

	if totalWeight <= 0 {
		return buckets[0].Damage
	}

	// Generate random value and find bucket
	random := rng.Float64() * totalWeight
	cumulative := 0.0
	for _, bucket := range buckets {
		cumulative += bucket.Weight
		if random <= cumulative {
			return bucket.Damage
		}
	}

	return buckets[len(buckets)-1].Damage
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
			if _, err := re.GetCombatPrediction(unit.UnitType, targetUnit.UnitType); err == nil {
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
	_, err := re.GetCombatPrediction(attacker.UnitType, target.UnitType)
	if err != nil {
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
