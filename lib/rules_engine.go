package weewar

import (
	"fmt"
	"math/rand"

	"github.com/panyam/turnengine/games/weewar/assets"
	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

// =============================================================================
// Rules Engine - Extends existing types with data-driven rules
// =============================================================================

// RulesEngine provides data-driven game rules using proto definitions
type RulesEngine struct {
	// Core data maps using proto types for consistency
	Units    map[int32]*v1.UnitDefinition    `json:"units"`
	Terrains map[int32]*v1.TerrainDefinition `json:"terrains"`

	// Canonical rule matrices using proto types
	MovementMatrix *v1.MovementMatrix `json:"movementMatrix"`
	AttackMatrix   *AttackMatrix      `json:"attackMatrix"` // TODO: Convert to proto type
}

// MovementMatrix is now defined in protos/weewar/v1/models.proto

// AttackMatrix defines combat outcomes between unit types using IDs
type AttackMatrix struct {
	// attacks[attackerID][defenderID] = damage distribution
	Attacks map[int32]map[int32]*DamageDistribution `json:"attacks"`
}

// =============================================================================
// Constructor and Initialization
// =============================================================================

// NewRulesEngine creates a new rules engine instance
func NewRulesEngine() *RulesEngine {
	return &RulesEngine{
		Units:          make(map[int32]*v1.UnitDefinition),
		Terrains:       make(map[int32]*v1.TerrainDefinition),
		MovementMatrix: &v1.MovementMatrix{Costs: make(map[int32]*v1.TerrainCostMap)},
		AttackMatrix:   &AttackMatrix{Attacks: make(map[int32]map[int32]*DamageDistribution)},
	}
}

// =============================================================================
// Enhanced Data Access API (extends existing GetUnitData pattern)
// =============================================================================

// GetUnitData returns unit data by ID (enhanced version of existing function)
func (re *RulesEngine) GetUnitData(unitID int32) (*v1.UnitDefinition, error) {
	unit, exists := re.Units[unitID]
	if !exists {
		return nil, fmt.Errorf("unit ID %d not found", unitID)
	}

	return unit, nil
}

// GetTerrainData returns terrain data by ID
func (re *RulesEngine) GetTerrainData(terrainID int32) (*v1.TerrainDefinition, error) {
	terrain, exists := re.Terrains[terrainID]
	if !exists {
		return nil, fmt.Errorf("terrain ID %d not found", terrainID)
	}

	return terrain, nil
}

// getUnitTerrainCost returns movement cost for unit type on terrain type (internal helper)
// First checks unit-specific matrix, then falls back to terrain's base cost
func (re *RulesEngine) getUnitTerrainCost(unitID, terrainID int32) (float64, error) {
	// First, try unit-specific movement cost from matrix
	if unitCosts, exists := re.MovementMatrix.Costs[unitID]; exists {
		if cost, exists := unitCosts.TerrainCosts[terrainID]; exists {
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

// GetMovementCost calculates movement cost for a unit to move to a specific destination
// Uses the unit's current position as starting point and recalculates based on current world state
func (re *RulesEngine) GetMovementCost(world *World, unit *v1.Unit, to AxialCoord) (float64, error) {
	if unit == nil {
		return 0, fmt.Errorf("unit is nil")
	}

	from := UnitGetCoord(unit)
	if from == to {
		return 0, nil
	}

	// For single adjacent moves, just return terrain cost
	distance := CubeDistance(from, to)
	if distance == 1 {
		toTile := world.TileAt(to)
		if toTile == nil {
			return 0, fmt.Errorf("invalid destination tile")
		}
		return re.getUnitTerrainCost(unit.UnitType, toTile.TileType)
	}

	// For multi-tile moves, use Dijkstra pathfinding
	return re.calculatePathCost(world, unit.UnitType, from, to)
}

// calculatePathCost uses Dijkstra's algorithm to find minimum cost path
func (re *RulesEngine) calculatePathCost(world *World, unitType int32, from, to AxialCoord) (float64, error) {
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

// =============================================================================
// Path Validation Methods
// =============================================================================

// IsValidPath validates if a unit can traverse a specific path
// This method performs comprehensive validation of any path, including:
// - Path structure (adjacent tiles, no jumps)
// - Terrain traversability (unit type vs terrain rules)
// - Movement cost feasibility (enough movement points)
// - Game state validity (no units blocking, correct start position)
func (re *RulesEngine) IsValidPath(unit *v1.Unit, path []AxialCoord, world *World) (bool, error) {
	if unit == nil {
		return false, fmt.Errorf("unit is nil")
	}

	if len(path) == 0 {
		return false, fmt.Errorf("path is empty")
	}

	// Path must start at unit's current position
	unitCoord := UnitGetCoord(unit)
	if path[0] != unitCoord {
		return false, fmt.Errorf("path does not start at unit position: expected %v, got %v", unitCoord, path[0])
	}

	// Empty movement (staying in place) is always valid
	if len(path) == 1 {
		return true, nil
	}

	totalCost := 0.0

	// Validate each step in the path
	for i := 1; i < len(path); i++ {
		fromCoord := path[i-1]
		toCoord := path[i]

		// 1. Check adjacency - tiles must be adjacent (no jumping)
		distance := CubeDistance(fromCoord, toCoord)
		if distance != 1 {
			return false, fmt.Errorf("path step %d->%d: tiles are not adjacent (distance=%d)", i-1, i, distance)
		}

		// 2. Check destination tile exists
		toTile := world.TileAt(toCoord)
		if toTile == nil {
			return false, fmt.Errorf("path step %d: destination tile %v does not exist", i, toCoord)
		}

		// 3. Check terrain traversability
		stepCost, err := re.getUnitTerrainCost(unit.UnitType, toTile.TileType)
		if err != nil {
			return false, fmt.Errorf("path step %d: unit type %d cannot traverse terrain %d: %w",
				i, unit.UnitType, toTile.TileType, err)
		}

		// 4. Check for blocking units
		blockingUnit := world.UnitAt(toCoord)
		if blockingUnit != nil && blockingUnit != unit {
			return false, fmt.Errorf("path step %d: tile %v is blocked by unit", i, toCoord)
		}

		// 5. Accumulate movement cost
		totalCost += stepCost
	}

	// 6. Check total movement cost against unit's remaining movement
	if totalCost > float64(unit.DistanceLeft) {
		return false, fmt.Errorf("path requires %.2f movement points, unit has %d remaining",
			totalCost, unit.DistanceLeft)
	}

	return true, nil
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

// GetMovementOptions returns all EMPTY tiles a unit can move to using Dijkstra's algorithm
// Only returns tiles without units (movement destinations)
func (re *RulesEngine) GetMovementOptions(world *World, unit *v1.Unit, remainingMovement int) ([]TileOption, error) {
	if unit == nil {
		return nil, fmt.Errorf("unit is nil")
	}

	_, err := re.GetUnitData(unit.UnitType)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit data: %w", err)
	}

	unitCoord := UnitGetCoord(unit)
	return re.dijkstraMovement(world, unit.UnitType, unitCoord, float64(remainingMovement))
}

// dijkstraMovement implements Dijkstra's algorithm to find all reachable EMPTY tiles with minimum cost
func (re *RulesEngine) dijkstraMovement(world *World, unitType int32, startCoord AxialCoord, maxMovement float64) ([]TileOption, error) {
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
			tile := world.TileAt(neighborCoord)
			if tile == nil {
				continue // Invalid tile
			}

			// Skip if occupied by another unit (movement rule: only empty tiles)
			if world.UnitAt(neighborCoord) != nil {
				continue // Occupied tile
			}

			// Get movement cost to this terrain
			moveCost, err := re.getUnitTerrainCost(unitType, tile.TileType)
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

// Note: Default terrain data has been migrated to proto definitions.
// Use LoadRulesEngineFromJSON to load terrain definitions from proto-based data.

var (
	defaultRulesEngine *RulesEngine
)

func init() {
	var err error
	defaultRulesEngine, err = LoadRulesEngineFromJSON(assets.RulesDataJSON)
	if err != nil {
		panic(err)
	}
}

// GetDefaultRulesEngine returns a font family that works in WASM environments
func DefaultRulesEngine() *RulesEngine {
	return defaultRulesEngine
}
