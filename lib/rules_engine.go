package weewar

import (
	"fmt"

	"github.com/panyam/turnengine/games/weewar/assets"
	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

// =============================================================================
// Rules Engine - Extends existing types with data-driven rules
// =============================================================================

// RulesEngine embeds the proto-based rules engine
type RulesEngine struct {
	*v1.RulesEngine
}

// =============================================================================
// Constructor and Initialization
// =============================================================================

// NewRulesEngine creates a new rules engine instance
func NewRulesEngine() *RulesEngine {
	return &RulesEngine{
		RulesEngine: &v1.RulesEngine{
			Units:                  make(map[int32]*v1.UnitDefinition),
			Terrains:               make(map[int32]*v1.TerrainDefinition),
			TerrainUnitProperties:  make(map[string]*v1.TerrainUnitProperties),
			UnitUnitProperties:     make(map[string]*v1.UnitUnitProperties),
		},
	}
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

// PopulateReferenceMaps populates the terrain/unit property reference maps for fast lookup
// This should be called after loading the centralized properties
func (re *RulesEngine) PopulateReferenceMaps() {
	// Initialize reference maps in units and terrains
	for _, unit := range re.Units {
		if unit.TerrainProperties == nil {
			unit.TerrainProperties = make(map[int32]*v1.TerrainUnitProperties)
		}
	}
	for _, terrain := range re.Terrains {
		if terrain.UnitProperties == nil {
			terrain.UnitProperties = make(map[int32]*v1.TerrainUnitProperties)
		}
	}

	// Populate reference maps from centralized properties
	for _, props := range re.TerrainUnitProperties {
		// Add to unit's terrain map
		if unit := re.Units[props.UnitId]; unit != nil {
			unit.TerrainProperties[props.TerrainId] = props
		}
		
		// Add to terrain's unit map  
		if terrain := re.Terrains[props.TerrainId]; terrain != nil {
			terrain.UnitProperties[props.UnitId] = props
		}
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

// GetMovementOptions returns all EMPTY tiles a unit can move to using Dijkstra's algorithm
// Only returns tiles without units (movement destinations)
func (re *RulesEngine) GetMovementOptions(world *World, unit *v1.Unit, remainingMovement int) (distances map[AxialCoord]float64, parents map[AxialCoord]AxialCoord, err error) {
	if unit == nil {
		return nil, nil, fmt.Errorf("unit is nil")
	}

	_, err = re.GetUnitData(unit.UnitType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get unit data: %w", err)
	}

	unitCoord := UnitGetCoord(unit)
	distances, parents = re.dijkstraMovement(world, unit.UnitType, unitCoord, float64(remainingMovement))
	return
}

// GetMovementCost calculates movement cost for a unit to move to a specific destination
// Uses dijkstraMovement for accurate pathfinding costs
func (re *RulesEngine) GetMovementCost(world *World, unit *v1.Unit, to AxialCoord) (float64, error) {
	if unit == nil {
		return 0, fmt.Errorf("unit is nil")
	}

	from := UnitGetCoord(unit)
	if from == to {
		return 0, nil
	}

	// Use dijkstraMovement to get accurate costs
	distances, _ := re.dijkstraMovement(world, unit.UnitType, from, float64(unit.DistanceLeft))
	
	cost, exists := distances[to]
	if !exists {
		return 0, fmt.Errorf("destination %v is not reachable from %v", to, from)
	}
	
	return cost, nil
}

// calculatePathCost uses dijkstraMovement to find minimum cost path
func (re *RulesEngine) calculatePathCost(world *World, unitType int32, from, to AxialCoord) (float64, error) {
	// Get unit data to determine movement points
	unitData, err := re.GetUnitData(unitType)
	if err != nil {
		return 0, fmt.Errorf("failed to get unit data: %w", err)
	}
	
	// Use the unit's maximum movement points as limit
	maxMovement := float64(unitData.MovementPoints)
	
	distances, _ := re.dijkstraMovement(world, unitType, from, maxMovement)
	
	cost, exists := distances[to]
	if !exists {
		return 0, fmt.Errorf("destination %v is not reachable from %v with %d movement points", to, from, unitData.MovementPoints)
	}
	
	return cost, nil
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

// dijkstraMovement implements Dijkstra's algorithm to find all reachable EMPTY tiles with minimum cost
func (re *RulesEngine) dijkstraMovement(world *World, unitType int32, startCoord AxialCoord, maxMovement float64) (distances map[AxialCoord]float64, parents map[AxialCoord]AxialCoord) {
	// Distance map: coord -> minimum cost to reach
	distances = map[AxialCoord]float64{}
	parents = map[AxialCoord]AxialCoord{}

	// Priority queue for Dijkstra (simple implementation)
	type queueItem struct {
		coord AxialCoord
		cost  float64
	}

	queue := []queueItem{{coord: startCoord, cost: 0}}
	distances[startCoord] = 0

	popMinCoord := func() queueItem {
		minIdx := 0
		for i := 1; i < len(queue); i++ {
			if queue[i].cost < queue[minIdx].cost {
				minIdx = i
			}
		}

		current := queue[minIdx]
		// Remove from queue
		queue = append(queue[:minIdx], queue[minIdx+1:]...)
		return current
	}

	// Dijkstra's algorithm
	for len(queue) > 0 {
		// Find minimum cost item (simple O(n) for now, could use heap)
		current := popMinCoord()
		fmt.Println("111111 - Here???", len(queue), "Current: ", current)

		// Skip if we've already processed this with lower cost
		if cost, exists := distances[current.coord]; exists && current.cost > cost {
			continue
		}

		// Explore neighbors
		for neighborCoord, tile := range world.Neighbors(current.coord) {
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

			terrain, _ := re.GetTerrainData(tile.TileType)
			fmt.Printf("From: %s, To: %s, TileType: %s, Cost: %f, Error: %v\n", current.coord, neighborCoord, terrain.Name, moveCost, err)
			if newCost <= maxMovement {
				// Check if this is a better path to the neighbor
				if existingCost, exists := distances[neighborCoord]; !exists || newCost < existingCost {
					distances[neighborCoord] = newCost
					parents[neighborCoord] = current.coord
					queue = append(queue, queueItem{coord: neighborCoord, cost: newCost})
				}
			}
		}
	}

	// Convert distances map to TileOption slice (excluding start position)
	return distances, parents
}

// getUnitTerrainCost returns movement cost for unit type on terrain type (internal helper)
// Uses the new centralized TerrainUnitProperties system
func (re *RulesEngine) getUnitTerrainCost(unitID, terrainID int32) (float64, error) {
	// Create key for centralized properties lookup
	key := fmt.Sprintf("%d:%d", terrainID, unitID)
	
	// First, try centralized properties (source of truth)
	if props, exists := re.TerrainUnitProperties[key]; exists {
		if props.MovementCost > 0 {
			return props.MovementCost, nil
		}
	}

	// Fall back to unit's terrain properties map (populated reference)
	if unit, err := re.GetUnitData(unitID); err == nil {
		if props, exists := unit.TerrainProperties[terrainID]; exists {
			if props.MovementCost > 0 {
				return props.MovementCost, nil
			}
		}
	}

	// Final fallback to default movement cost of 1.0
	return 1.0, nil
}
