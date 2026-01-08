package lib

import (
	"container/heap"
	"fmt"

	"github.com/turnforge/weewar/assets"
	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
)

// =============================================================================
// Priority Queue for Dijkstra's Algorithm (heap-based, O(log n) operations)
// =============================================================================

// dijkstraItem represents a coordinate with its movement cost for the priority queue
type dijkstraItem struct {
	coord AxialCoord
	cost  float64
	index int // index in the heap, maintained by heap.Interface
}

// dijkstraHeap implements heap.Interface for efficient min-cost extraction
type dijkstraHeap []*dijkstraItem

func (h dijkstraHeap) Len() int           { return len(h) }
func (h dijkstraHeap) Less(i, j int) bool { return h[i].cost < h[j].cost }
func (h dijkstraHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *dijkstraHeap) Push(x any) {
	n := len(*h)
	item := x.(*dijkstraItem)
	item.index = n
	*h = append(*h, item)
}

func (h *dijkstraHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*h = old[0 : n-1]
	return item
}

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
			Units:                 make(map[int32]*v1.UnitDefinition),
			Terrains:              make(map[int32]*v1.TerrainDefinition),
			TerrainUnitProperties: make(map[string]*v1.TerrainUnitProperties),
			UnitUnitProperties:    make(map[string]*v1.UnitUnitProperties),
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
	defaultRulesEngine, err = LoadRulesEngineFromJSON(assets.RulesDataJSON, assets.RulesDamageDataJSON)
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

// =============================================================================
// Terrain Type Classification Methods
// =============================================================================

// GetTerrainType returns the terrain type classification for a terrain ID
func (re *RulesEngine) GetTerrainType(terrainID int32) v1.TerrainType {
	if t, ok := re.TerrainTypes[terrainID]; ok {
		return t
	}
	return v1.TerrainType_TERRAIN_TYPE_UNSPECIFIED
}

// IsCityTerrain checks if a terrain is a city/building tile (player-owned structures)
func (re *RulesEngine) IsCityTerrain(terrainID int32) bool {
	return re.GetTerrainType(terrainID) == v1.TerrainType_TERRAIN_TYPE_CITY
}

// IsNatureTerrain checks if a terrain is a natural tile (grass, mountains, etc.)
func (re *RulesEngine) IsNatureTerrain(terrainID int32) bool {
	return re.GetTerrainType(terrainID) == v1.TerrainType_TERRAIN_TYPE_NATURE
}

// IsBridgeTerrain checks if a terrain is a bridge
func (re *RulesEngine) IsBridgeTerrain(terrainID int32) bool {
	return re.GetTerrainType(terrainID) == v1.TerrainType_TERRAIN_TYPE_BRIDGE
}

// IsWaterTerrain checks if a terrain is water
func (re *RulesEngine) IsWaterTerrain(terrainID int32) bool {
	return re.GetTerrainType(terrainID) == v1.TerrainType_TERRAIN_TYPE_WATER
}

// IsRoadTerrain checks if a terrain is a road
func (re *RulesEngine) IsRoadTerrain(terrainID int32) bool {
	return re.GetTerrainType(terrainID) == v1.TerrainType_TERRAIN_TYPE_ROAD
}

// GetCityTerrains returns a map of terrain IDs that are city/building types.
// This is useful for themes that need to know which terrains use player colors.
func (re *RulesEngine) GetCityTerrains() map[int32]bool {
	result := make(map[int32]bool)
	for terrainID, terrainType := range re.TerrainTypes {
		if terrainType == v1.TerrainType_TERRAIN_TYPE_CITY {
			result[terrainID] = true
		}
	}
	return result
}

// =============================================================================
// Action Progression System
// =============================================================================

// GetAllowedActionsForUnit returns which actions are currently valid for a unit
// based on its progression_step index into the UnitDefinition.action_order
func (re *RulesEngine) GetAllowedActionsForUnit(unit *v1.Unit, unitDef *v1.UnitDefinition) []string {
	// Get action_order, default to ["move", "attack|capture"]
	actionOrder := unitDef.ActionOrder
	if len(actionOrder) == 0 {
		actionOrder = []string{"move", "attack|capture"}
	}

	// Check if all steps complete
	if unit.ProgressionStep >= int32(len(actionOrder)) {
		return []string{} // Only end turn available
	}

	// Get current step's actions
	stepActions := actionOrder[unit.ProgressionStep]
	alternatives := ParseActionAlternatives(stepActions)

	// If user already chose an alternative from pipe-separated options,
	// only that alternative is allowed (prevents switching mid-step)
	if unit.ChosenAlternative != "" {
		alternatives = []string{unit.ChosenAlternative}
	}

	// Filter by what can actually be performed
	var allowed []string
	for _, action := range alternatives {
		if re.canPerformAction(unit, unitDef, action) {
			allowed = append(allowed, action)
		}
	}

	return allowed
}

// GetAllowedActionsForTile returns which actions are currently valid for a tile
// for a given player with specified coin balance.
// NOTE: Caller should only call this for tiles belonging to the player being checked.
// This function does not validate current turn or player ownership - that's the caller's responsibility.
func (re *RulesEngine) GetAllowedActionsForTile(tile *v1.Tile, terrainDef *v1.TerrainDefinition, playerCoins int32) []string {
	var allowed []string

	// Check if this terrain type can build units and player can afford at least one
	if terrainDef != nil && len(terrainDef.BuildableUnitIds) > 0 {
		// Check if player can afford at least one buildable unit
		canAffordAny := false
		for _, unitTypeID := range terrainDef.BuildableUnitIds {
			unitDef, err := re.GetUnitData(unitTypeID)
			if err == nil && unitDef.Coins <= playerCoins {
				canAffordAny = true
				break
			}
		}

		if canAffordAny {
			allowed = append(allowed, "build")
		}
	}

	// TODO: Add other tile-specific actions:
	// - "repair" - if tile can repair units on it
	// - "heal" - if tile provides healing
	// - "income" - if tile generates income

	return allowed
}

// canPerformAction checks if a unit can currently perform a specific action
// based on available resources (distance_left, etc.) and action_limits
func (re *RulesEngine) canPerformAction(unit *v1.Unit, unitDef *v1.UnitDefinition, action string) bool {
	switch action {
	case "move":
		// Can move if has movement points remaining
		return unit.DistanceLeft > 0

	case "attack":
		// Can attack if hasn't reached attack limit
		// TODO: Implement attack counting at current step and check against action_limits
		return true

	case "capture":
		// Can capture if unit has capture ability
		// TODO: Add canCapture flag to UnitDefinition or infer from unit type
		return true

	case "build":
		// Can build if unit has build ability
		// TODO: Add canBuild flag to UnitDefinition or infer from unit type
		return true

	case "retreat":
		// Can retreat if has retreat points remaining (DistanceLeft is set to retreat_points after attack)
		return unit.DistanceLeft > 0

	default:
		return false
	}
}

// ParseActionAlternatives parses pipe-separated action alternatives
// e.g., "attack|capture" -> ["attack", "capture"]
func ParseActionAlternatives(stepActions string) []string {
	// Simple split on pipe character
	alternatives := []string{}
	current := ""

	for _, ch := range stepActions {
		if ch == '|' {
			if current != "" {
				alternatives = append(alternatives, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		alternatives = append(alternatives, current)
	}

	return alternatives
}

// =============================================================================
// Movement Options
// =============================================================================

// GetMovementOptions returns all tiles a unit can move to using Dijkstra's algorithm
// Returns AllPaths structure containing all reachable tiles and path information
// When preventPassThrough is false (default), units can traverse through occupied tiles but cannot land on them
func (re *RulesEngine) GetMovementOptions(world *World, unit *v1.Unit, remainingMovement int, preventPassThrough bool) (*v1.AllPaths, error) {
	if unit == nil {
		return nil, fmt.Errorf("unit is nil")
	}

	_, err := re.GetUnitData(unit.UnitType)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit data: %w", err)
	}

	unitCoord := UnitGetCoord(unit)
	allPaths := re.dijkstraMovement(world, unit.UnitType, unitCoord, float64(remainingMovement), preventPassThrough)
	return allPaths, nil
}

// GetMovementCost calculates movement cost for a unit to move to a specific destination
// Uses dijkstraMovement for accurate pathfinding costs
func (re *RulesEngine) GetMovementCost(world *World, unit *v1.Unit, to AxialCoord, preventPassThrough bool) (float64, error) {
	if unit == nil {
		return 0, fmt.Errorf("unit is nil")
	}

	from := UnitGetCoord(unit)
	if from == to {
		return 0, nil
	}

	// Use dijkstraMovement to get accurate costs
	allPaths := re.dijkstraMovement(world, unit.UnitType, from, float64(unit.DistanceLeft), preventPassThrough)

	// Look up the destination in AllPaths
	key := fmt.Sprintf("%d,%d", to.Q, to.R)
	edge, exists := allPaths.Edges[key]
	if !exists {
		return 0, fmt.Errorf("destination %v is not reachable from %v", to, from)
	}

	return float64(edge.TotalCost), nil
}

// calculatePathCost uses dijkstraMovement to find minimum cost path
func (re *RulesEngine) calculatePathCost(world *World, unitType int32, from, to AxialCoord, preventPassThrough bool) (float64, error) {
	// Get unit data to determine movement points
	unitData, err := re.GetUnitData(unitType)
	if err != nil {
		return 0, fmt.Errorf("failed to get unit data: %w", err)
	}

	// Use the unit's maximum movement points as limit
	maxMovement := float64(unitData.MovementPoints)

	allPaths := re.dijkstraMovement(world, unitType, from, maxMovement, preventPassThrough)

	// Look up the destination in AllPaths
	key := fmt.Sprintf("%d,%d", to.Q, to.R)
	edge, exists := allPaths.Edges[key]
	if !exists {
		return 0, fmt.Errorf("destination %v is not reachable from %v with %f movement points", to, from, unitData.MovementPoints)
	}

	return float64(edge.TotalCost), nil
}

// =============================================================================
// Path Finding Methods
// =============================================================================

// FindPathTo finds the shortest path from unit's position to destination using Dijkstra.
// Stops as soon as destination is reached for efficiency.
// Returns the path and total cost, or an error if destination is unreachable.
func (re *RulesEngine) FindPathTo(unit *v1.Unit, dest AxialCoord, world *World, preventPassThrough bool) (*v1.Path, float64, error) {
	if unit == nil {
		return nil, 0, fmt.Errorf("unit is nil")
	}

	startCoord := UnitGetCoord(unit)

	// Same position is always valid (no movement)
	if startCoord == dest {
		return &v1.Path{Edges: []*v1.PathEdge{}, TotalCost: 0}, 0, nil
	}

	maxMovement := unit.DistanceLeft

	// Track visited nodes, their costs, and parent info for path reconstruction
	type nodeInfo struct {
		cost       float64
		parentQ    int32
		parentR    int32
		moveCost   float64
		isOccupied bool
	}
	visited := make(map[AxialCoord]*nodeInfo)

	// Priority queue for Dijkstra
	type queueItem struct {
		coord AxialCoord
		cost  float64
	}

	queue := []queueItem{{coord: startCoord, cost: 0}}
	visited[startCoord] = &nodeInfo{cost: 0, parentQ: int32(startCoord.Q), parentR: int32(startCoord.R)}

	popMinCoord := func() queueItem {
		minIdx := 0
		for i := 1; i < len(queue); i++ {
			if queue[i].cost < queue[minIdx].cost {
				minIdx = i
			}
		}
		current := queue[minIdx]
		queue = append(queue[:minIdx], queue[minIdx+1:]...)
		return current
	}

	// Dijkstra's algorithm with early exit
	for len(queue) > 0 {
		current := popMinCoord()

		// Early exit: reached destination
		if current.coord == dest {
			break
		}

		// Skip if we've already processed this with lower cost
		if info, exists := visited[current.coord]; exists && current.cost > info.cost {
			continue
		}

		// Explore neighbors
		for neighborCoord := range world.Neighbors(current.coord) {
			isOccupied := world.UnitAt(neighborCoord) != nil

			if preventPassThrough && isOccupied {
				continue
			}

			effectiveTileType := re.GetEffectiveTileType(world, neighborCoord)
			moveCost, err := re.GetUnitTerrainCost(unit.UnitType, effectiveTileType)
			if err != nil {
				continue
			}

			newCost := current.cost + moveCost

			if newCost <= maxMovement {
				if existingInfo, exists := visited[neighborCoord]; !exists || newCost < existingInfo.cost {
					visited[neighborCoord] = &nodeInfo{
						cost:       newCost,
						parentQ:    int32(current.coord.Q),
						parentR:    int32(current.coord.R),
						moveCost:   moveCost,
						isOccupied: isOccupied,
					}
					queue = append(queue, queueItem{coord: neighborCoord, cost: newCost})
				}
			}
		}
	}

	// Check if destination was reached
	destInfo, reached := visited[dest]
	if !reached {
		return nil, 0, fmt.Errorf("destination (%d,%d) not reachable from (%d,%d)",
			dest.Q, dest.R, startCoord.Q, startCoord.R)
	}

	// Cannot land on occupied tile
	if destInfo.isOccupied {
		return nil, 0, fmt.Errorf("destination (%d,%d) is occupied", dest.Q, dest.R)
	}

	// Reconstruct path by walking backwards from destination
	var pathEdges []*v1.PathEdge
	currentQ, currentR := int32(dest.Q), int32(dest.R)

	for {
		info := visited[AxialCoord{Q: int(currentQ), R: int(currentR)}]
		if info == nil || (currentQ == int32(startCoord.Q) && currentR == int32(startCoord.R)) {
			break
		}

		pathEdges = append(pathEdges, &v1.PathEdge{
			FromQ:        info.parentQ,
			FromR:        info.parentR,
			ToQ:          currentQ,
			ToR:          currentR,
			MovementCost: info.moveCost,
			TotalCost:    info.cost,
		})

		currentQ, currentR = info.parentQ, info.parentR
	}

	// Reverse to get source->destination order
	for i := 0; i < len(pathEdges)/2; i++ {
		j := len(pathEdges) - 1 - i
		pathEdges[i], pathEdges[j] = pathEdges[j], pathEdges[i]
	}

	return &v1.Path{
		Edges:     pathEdges,
		TotalCost: destInfo.cost,
	}, destInfo.cost, nil
}

// IsValidPath validates if a unit can traverse a specific path (legacy compatibility)
// Prefer FindPathTo for new code as it finds the optimal path.
func (re *RulesEngine) IsValidPath(unit *v1.Unit, path []AxialCoord, world *World) (bool, error) {
	if len(path) < 2 {
		return len(path) == 1, nil // Single position is valid (no movement)
	}

	dest := path[len(path)-1]
	_, _, err := re.FindPathTo(unit, dest, world, false)
	return err == nil, err
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

// dijkstraMovement implements Dijkstra's algorithm to find all reachable tiles with minimum cost
// When preventPassThrough is false (default), units can traverse through occupied tiles but cannot land on them
func (re *RulesEngine) dijkstraMovement(world *World, unitType int32, startCoord AxialCoord, maxMovement float64, preventPassThrough bool) *v1.AllPaths {
	// Initialize AllPaths
	allPaths := &v1.AllPaths{
		SourceQ: int32(startCoord.Q),
		SourceR: int32(startCoord.R),
		Edges:   make(map[string]*v1.PathEdge),
	}

	// Track visited nodes and their costs
	visited := make(map[AxialCoord]float64)

	// Get unit data for explanations
	unitData, _ := re.GetUnitData(unitType)

	// Priority queue using heap for O(log n) operations instead of O(n)
	pq := &dijkstraHeap{}
	heap.Init(pq)
	heap.Push(pq, &dijkstraItem{coord: startCoord, cost: 0})
	visited[startCoord] = 0

	// Dijkstra's algorithm
	for pq.Len() > 0 {
		// Extract minimum cost item in O(log n)
		current := heap.Pop(pq).(*dijkstraItem)

		// Skip if we've already processed this with lower cost
		if cost, exists := visited[current.coord]; exists && current.cost > cost {
			continue
		}

		// Explore neighbors
		for neighborCoord := range world.Neighbors(current.coord) {
			// Check if tile is occupied by another unit
			isOccupied := world.UnitAt(neighborCoord) != nil

			// If preventPassThrough is true, skip occupied tiles entirely
			if preventPassThrough && isOccupied {
				continue // Occupied tile blocks traversal
			}

			// Get effective tile type (considers crossings like roads/bridges)
			effectiveTileType := re.GetEffectiveTileType(world, neighborCoord)

			// Get movement cost to this terrain (using effective type)
			moveCost, err := re.GetUnitTerrainCost(unitType, effectiveTileType)
			if err != nil {
				continue // Cannot move on this terrain
			}

			newCost := current.cost + moveCost

			if newCost <= maxMovement {
				// Check if this is a better path to the neighbor
				if existingCost, exists := visited[neighborCoord]; !exists || newCost < existingCost {
					visited[neighborCoord] = newCost

					// Add to heap for further exploration (pass-through)
					heap.Push(pq, &dijkstraItem{coord: neighborCoord, cost: newCost})

					// Get terrain data for explanation (use effective type for display)
					terrainData, _ := re.GetTerrainData(effectiveTileType)
					terrainName := "unknown"
					if terrainData != nil {
						terrainName = terrainData.Name
					}

					// Create explanation
					unitName := "Unit"
					if unitData != nil {
						unitName = unitData.Name
					}
					explanation := fmt.Sprintf("%s costs %s %.0f movement points", terrainName, unitName, moveCost)

					// Always add edges to AllPaths for path reconstruction
					// Mark occupied tiles with IsOccupied=true to indicate pass-through only
					// (GetOptionsAt will filter these out as invalid landing spots)
					key := fmt.Sprintf("%d,%d", neighborCoord.Q, neighborCoord.R)
					allPaths.Edges[key] = &v1.PathEdge{
						FromQ:        int32(current.coord.Q),
						FromR:        int32(current.coord.R),
						ToQ:          int32(neighborCoord.Q),
						ToR:          int32(neighborCoord.R),
						MovementCost: moveCost,
						TotalCost:    newCost,
						TerrainType:  terrainName,
						Explanation:  explanation,
						IsOccupied:   isOccupied,
					}
				}
			}
		}
	}

	return allPaths
}

// GetUnitTerrainCost returns movement cost for unit type on terrain type (internal helper)
// Uses the new centralized TerrainUnitProperties system
func (re *RulesEngine) GetUnitTerrainCost(unitID, terrainID int32) (float64, error) {
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

// =============================================================================
// Crossing (Road/Bridge) Support
// =============================================================================

// GetEffectiveTileType returns the tile type to use for rules lookups,
// considering crossings (roads/bridges) that override the base terrain.
// Roads return the Road tile type (22), bridges return the appropriate bridge type
// based on the underlying water depth.
func (re *RulesEngine) GetEffectiveTileType(world *World, coord AxialCoord) int32 {
	tile := world.TileAt(coord)
	if tile == nil {
		return 0
	}

	// Check for crossing - crossings override terrain for movement
	crossingType := world.CrossingTypeAt(coord)
	if crossingType == v1.CrossingType_CROSSING_TYPE_ROAD {
		return TileTypeRoad // Road tile type ID (22)
	}
	if crossingType == v1.CrossingType_CROSSING_TYPE_BRIDGE {
		// Bridge type depends on underlying water terrain
		switch tile.TileType {
		case TileTypeWaterShallow:
			return TileTypeBridgeShallow // Bridge over shallow water (18)
		case TileTypeWaterRegular:
			return TileTypeBridgeRegular // Bridge over regular water (17)
		case TileTypeWaterDeep:
			return TileTypeBridgeDeep // Bridge over deep water (19)
		default:
			// Default to regular bridge if terrain doesn't match expected water types
			return TileTypeBridgeRegular
		}
	}

	return tile.TileType
}

// GetUnitTerrainCostAt returns movement cost for a unit at a specific coordinate,
// considering crossings (roads/bridges) that may override the terrain type.
func (re *RulesEngine) GetUnitTerrainCostAt(world *World, unitID int32, coord AxialCoord) (float64, error) {
	effectiveTileType := re.GetEffectiveTileType(world, coord)
	return re.GetUnitTerrainCost(unitID, effectiveTileType)
}
