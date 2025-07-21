package weewar

import (
	"fmt"
	"math/rand"
)

// =============================================================================
// Rules Engine - Extends existing types with data-driven rules
// =============================================================================

// RulesEngine provides data-driven game rules working with existing types
type RulesEngine struct {
	// Core data maps (extends existing unitDataMap pattern)
	Units    map[int]*UnitData    `json:"units"`
	Terrains map[int]*TerrainData `json:"terrains"`

	// Canonical rule matrices
	MovementMatrix *MovementMatrix `json:"movementMatrix"`
	AttackMatrix   *AttackMatrix   `json:"attackMatrix"`
}

// MovementMatrix defines movement costs between unit types and terrain types using IDs
type MovementMatrix struct {
	// costs[unitID][terrainID] = movement cost
	Costs map[int]map[int]float64 `json:"costs"`
}

// AttackMatrix defines combat outcomes between unit types using IDs
type AttackMatrix struct {
	// attacks[attackerID][defenderID] = damage distribution
	Attacks map[int]map[int]*DamageDistribution `json:"attacks"`
}

// =============================================================================
// Global Rules Engine Instance
// =============================================================================

var globalRulesEngine *RulesEngine

// GetRulesEngine returns the global rules engine instance
func GetRulesEngine() *RulesEngine {
	if globalRulesEngine == nil {
		globalRulesEngine = NewRulesEngine()
	}
	return globalRulesEngine
}

// =============================================================================
// Constructor and Initialization
// =============================================================================

// NewRulesEngine creates a new rules engine instance
func NewRulesEngine() *RulesEngine {
	return &RulesEngine{
		Units:          make(map[int]*UnitData),
		Terrains:       make(map[int]*TerrainData),
		MovementMatrix: &MovementMatrix{Costs: make(map[int]map[int]float64)},
		AttackMatrix:   &AttackMatrix{Attacks: make(map[int]map[int]*DamageDistribution)},
	}
}

// =============================================================================
// Enhanced Data Access API (extends existing GetUnitData pattern)
// =============================================================================

// GetUnitData returns unit data by ID (enhanced version of existing function)
func (re *RulesEngine) GetUnitData(unitID int) (*UnitData, error) {
	unit, exists := re.Units[unitID]
	if !exists {
		return nil, fmt.Errorf("unit ID %d not found", unitID)
	}

	return unit, nil
}

// GetTerrainData returns terrain data by ID
func (re *RulesEngine) GetTerrainData(terrainID int) (*TerrainData, error) {
	terrain, exists := re.Terrains[terrainID]
	if !exists {
		return nil, fmt.Errorf("terrain ID %d not found", terrainID)
	}

	return terrain, nil
}

// GetTerrainMovementCost returns movement cost for unit type on terrain type
// First checks unit-specific matrix, then falls back to terrain's base cost
func (re *RulesEngine) GetTerrainMovementCost(unitID, terrainID int) (float64, error) {
	// First, try unit-specific movement cost from matrix
	if unitCosts, exists := re.MovementMatrix.Costs[unitID]; exists {
		if cost, exists := unitCosts[terrainID]; exists {
			return cost, nil
		}
	}

	// Fall back to terrain's base movement cost
	if terrain, err := re.GetTerrainData(terrainID); err == nil {
		if terrain.BaseMoveCost > 0 {
			return terrain.BaseMoveCost, nil
		}
	}

	// Final fallback to 1.0
	return 1.0, nil
}

// GetMovementCost calculates total movement cost from one position to another
// Uses Dijkstra's algorithm for multi-tile pathfinding with terrain costs
func (re *RulesEngine) GetMovementCost(gameMap *Map, unitType int, from, to AxialCoord) (float64, error) {
	if from == to {
		return 0, nil
	}

	// For single adjacent moves, just return terrain cost
	distance := CubeDistance(from, to)
	if distance == 1 {
		toTile := gameMap.TileAt(to)
		if toTile == nil {
			return 0, fmt.Errorf("invalid destination tile")
		}
		return re.GetTerrainMovementCost(unitType, toTile.TileType)
	}

	// For multi-tile moves, use Dijkstra pathfinding
	return re.calculatePathCost(gameMap, unitType, from, to)
}

// calculatePathCost uses Dijkstra's algorithm to find minimum cost path
func (re *RulesEngine) calculatePathCost(gameMap *Map, unitType int, from, to AxialCoord) (float64, error) {
	// Simple implementation - for now return distance * average terrain cost
	// TODO: Implement full Dijkstra's algorithm
	distance := float64(CubeDistance(from, to))

	// Get average terrain cost as approximation
	averageCost := 1.5                                    // Default average
	if terrain, err := re.GetTerrainData(1); err == nil { // Use grass as reference
		averageCost = terrain.BaseMoveCost
	}

	return distance * averageCost, nil
}

// CalculateCombatDamage calculates damage using canonical DamageDistribution
func (re *RulesEngine) CalculateCombatDamage(attackerID, defenderID int, rng *rand.Rand) (int, error) {
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
func (re *RulesEngine) GetCombatPrediction(attackerID, defenderID int) (*DamageDistribution, error) {
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

// =============================================================================
// Helper Functions
// =============================================================================

// GetLoadedUnitsCount returns number of loaded units
func (re *RulesEngine) GetLoadedUnitsCount() int {
	return len(re.Units)
}

// GetLoadedTerrainsCount returns number of loaded terrains
func (re *RulesEngine) GetLoadedTerrainsCount() int {
	return len(re.Terrains)
}

// ValidateRules performs basic validation
func (re *RulesEngine) ValidateRules() error {
	if len(re.Units) == 0 {
		return fmt.Errorf("no units loaded")
	}

	if len(re.Terrains) == 0 {
		return fmt.Errorf("no terrains loaded")
	}

	return nil
}

// =============================================================================
// Spatial Query Methods for UI/Gameplay
// =============================================================================

// TileOption represents a tile that a unit can move to with its cost
type TileOption struct {
	Coord AxialCoord `json:"coord"`
	Cost  float64    `json:"cost"`
}

// GetMovementOptions returns all tiles a unit can move to using Dijkstra's algorithm
func (re *RulesEngine) GetMovementOptions(gameMap *Map, unit *Unit, remainingMovement int) ([]TileOption, error) {
	if unit == nil {
		return nil, fmt.Errorf("unit is nil")
	}

	_, err := re.GetUnitData(unit.UnitType)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit data: %w", err)
	}

	return re.dijkstraMovement(gameMap, unit.UnitType, unit.Coord, float64(remainingMovement))
}

// dijkstraMovement implements Dijkstra's algorithm to find all reachable tiles with minimum cost
func (re *RulesEngine) dijkstraMovement(gameMap *Map, unitType int, startCoord AxialCoord, maxMovement float64) ([]TileOption, error) {
	// Distance map: coord -> minimum cost to reach
	distances := make(map[AxialCoord]float64)
	
	// Priority queue for Dijkstra (simple implementation)
	type queueItem struct {
		coord AxialCoord
		cost  float64
	}
	
	queue := []queueItem{{coord: startCoord, cost: 0}}
	distances[startCoord] = 0
	
	// Dijkstra's algorithm
	for len(queue) > 0 {
		// Find minimum cost item (simple O(n) for now, could use heap)
		minIdx := 0
		for i := 1; i < len(queue); i++ {
			if queue[i].cost < queue[minIdx].cost {
				minIdx = i
			}
		}
		
		current := queue[minIdx]
		// Remove from queue
		queue = append(queue[:minIdx], queue[minIdx+1:]...)
		
		// Skip if we've already processed this with lower cost
		if cost, exists := distances[current.coord]; exists && current.cost > cost {
			continue
		}
		
		// Get all 6 hex neighbors using existing helper
		var neighbors [6]AxialCoord
		current.coord.Neighbors(&neighbors)
		
		// Explore neighbors
		for _, neighborCoord := range neighbors {
			// Check if neighbor tile exists and is passable
			tile := gameMap.TileAt(neighborCoord)
			if tile == nil {
				continue // Invalid tile
			}
			
			// Skip if occupied by another unit
			if tile.Unit != nil {
				continue // Occupied tile
			}
			
			// Get movement cost to this terrain
			moveCost, err := re.GetTerrainMovementCost(unitType, tile.TileType)
			if err != nil {
				continue // Cannot move on this terrain
			}
			
			newCost := current.cost + moveCost
			
			// Skip if exceeds movement budget
			if newCost > maxMovement {
				continue
			}
			
			// Check if this is a better path to the neighbor
			if existingCost, exists := distances[neighborCoord]; !exists || newCost < existingCost {
				distances[neighborCoord] = newCost
				queue = append(queue, queueItem{coord: neighborCoord, cost: newCost})
			}
		}
	}
	
	// Convert distances map to TileOption slice (excluding start position)
	var options []TileOption
	for coord, cost := range distances {
		if coord != startCoord { // Exclude starting position
			options = append(options, TileOption{
				Coord: coord,
				Cost:  cost,
			})
		}
	}
	
	return options, nil
}

// GetAttackOptions returns all positions a unit can attack from its current position
func (re *RulesEngine) GetAttackOptions(gameMap *Map, unit *Unit) ([]AxialCoord, error) {
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
	for dQ := -attackRange; dQ <= attackRange; dQ++ {
		for dR := -attackRange; dR <= attackRange; dR++ {
			if dQ == 0 && dR == 0 {
				continue // Skip self
			}

			targetCoord := AxialCoord{Q: unit.Coord.Q + dQ, R: unit.Coord.R + dR}

			// Check if there's an enemy unit at this position
			tile := gameMap.TileAt(targetCoord)
			if tile == nil || tile.Unit == nil {
				continue // No unit to attack
			}

			// Check if it's an enemy unit (different player)
			if tile.Unit.PlayerID == unit.PlayerID {
				continue // Same player, can't attack
			}

			// Check if this unit can attack the target unit type
			if _, err := re.GetCombatPrediction(unit.UnitType, tile.Unit.UnitType); err == nil {
				attackPositions = append(attackPositions, targetCoord)
			}
		}
	}

	return attackPositions, nil
}

// CanUnitAttackTarget checks if a unit can attack a specific target
func (re *RulesEngine) CanUnitAttackTarget(attacker *Unit, target *Unit) (bool, error) {
	if attacker == nil || target == nil {
		return false, fmt.Errorf("attacker or target is nil")
	}

	// Check if units are enemies
	if attacker.PlayerID == target.PlayerID {
		return false, nil // Same team
	}

	// Check if attacker can attack this unit type
	_, err := re.GetCombatPrediction(attacker.UnitType, target.UnitType)
	if err != nil {
		return false, nil // Cannot attack this unit type
	}

	// Check range (using simple distance for now)
	distance := CubeDistance(attacker.Coord, target.Coord)
	unitData, err := re.GetUnitData(attacker.UnitType)
	if err != nil {
		return false, err
	}

	return distance <= unitData.AttackRange, nil
}
