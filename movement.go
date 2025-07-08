package weewar

import (
	"fmt"

	"github.com/panyam/turnengine/internal/turnengine"
)

type WeeWarMovementSystem struct {
	unitData    map[string]UnitData
	terrainData map[string]TerrainData
}

func NewWeeWarMovementSystem() (*WeeWarMovementSystem, error) {
	// Load WeeWar data
	data, err := loadWeeWarData()
	if err != nil {
		return nil, fmt.Errorf("failed to load WeeWar data: %w", err)
	}

	system := &WeeWarMovementSystem{
		unitData:    make(map[string]UnitData),
		terrainData: make(map[string]TerrainData),
	}

	// Index units by name
	for _, unit := range data.Units {
		system.unitData[unit.Name] = unit
	}

	// Index terrains by name
	for _, terrain := range data.Terrains {
		system.terrainData[terrain.Name] = terrain
	}

	return system, nil
}

func (wms *WeeWarMovementSystem) Name() string {
	return "WeeWarMovementSystem"
}

func (wms *WeeWarMovementSystem) Priority() int {
	return 50
}

func (wms *WeeWarMovementSystem) Update(world *turnengine.World) error {
	// This system doesn't run on every update, only when movement is initiated
	return nil
}

func (wms *WeeWarMovementSystem) CanMove(entity *turnengine.Entity, from, to turnengine.Position, board turnengine.Board) bool {
	// Get unit type
	unitType, err := wms.getUnitType(entity)
	if err != nil {
		return false
	}

	// Get movement range
	movementRange, err := wms.getMovementRange(entity)
	if err != nil {
		return false
	}

	// Check if target position is valid
	if !board.IsValidPosition(to) {
		return false
	}

	// Check if target position is occupied by another unit
	if hexBoard, ok := board.(*HexBoard); ok {
		if hexBoard.IsPositionOccupied(to) {
			return false
		}
	}

	// Get movement cost function for this unit type
	costFunc := wms.getMovementCostFunction(unitType)
	
	// Find path and calculate total cost
	intCostFunc := func(pos turnengine.Position) int {
		cost := costFunc(pos)
		if cost < 0 {
			return -1
		}
		return int(cost)
	}
	path, err := board.(*HexBoard).pathfinder.FindPath(board, from, to, intCostFunc)
	if err != nil {
		return false
	}

	// Calculate total movement cost
	totalCost := 0.0
	for i := 1; i < len(path); i++ {
		cost := costFunc(path[i])
		if cost < 0 {
			return false // Impassable terrain
		}
		totalCost += cost
	}

	return totalCost <= float64(movementRange)
}

func (wms *WeeWarMovementSystem) GetMovementRange(entity *turnengine.Entity, from turnengine.Position, board turnengine.Board) []turnengine.Position {
	// Get unit type
	unitType, err := wms.getUnitType(entity)
	if err != nil {
		return nil
	}

	// Get movement range
	movementRange, err := wms.getMovementRange(entity)
	if err != nil {
		return nil
	}

	// Get movement cost function for this unit type
	costFunc := wms.getMovementCostFunction(unitType)
	
	// Use board manager to get movement range
	if hexBoard, ok := board.(*HexBoard); ok {
		// Convert float64 cost function to int
		intCostFunc := func(pos turnengine.Position) int {
			cost := costFunc(pos)
			if cost < 0 {
				return -1 // Impassable
			}
			return int(cost)
		}
		
		// Calculate reachable positions manually
		var reachable []turnengine.Position
		
		queue := []struct {
			pos  turnengine.Position
			cost int
		}{{from, 0}}
		
		visited := make(map[string]bool)
		visited[from.Hash()] = true
		
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			
			neighbors := hexBoard.GetNeighbors(current.pos)
			for _, neighbor := range neighbors {
				if visited[neighbor.Hash()] || !hexBoard.IsValidPosition(neighbor) {
					continue
				}
				
				newCost := current.cost + intCostFunc(neighbor)
				if newCost <= movementRange && newCost > 0 {
					visited[neighbor.Hash()] = true
					reachable = append(reachable, neighbor)
					queue = append(queue, struct {
						pos  turnengine.Position
						cost int
					}{neighbor, newCost})
				}
			}
		}
		
		return reachable
	}

	return nil
}

func (wms *WeeWarMovementSystem) MoveUnit(entity *turnengine.Entity, from, to turnengine.Position, board turnengine.Board) error {
	// Check if move is valid
	if !wms.CanMove(entity, from, to, board) {
		return fmt.Errorf("invalid move from %s to %s", from.String(), to.String())
	}

	// Update entity position
	if err := wms.updateEntityPosition(entity, to); err != nil {
		return fmt.Errorf("failed to update entity position: %w", err)
	}

	// Update board entity tracking
	if hexBoard, ok := board.(*HexBoard); ok {
		if err := hexBoard.SetEntityAt(from, ""); err != nil {
			return fmt.Errorf("failed to clear old position: %w", err)
		}
		if err := hexBoard.SetEntityAt(to, entity.ID); err != nil {
			return fmt.Errorf("failed to set new position: %w", err)
		}
	}

	// Reduce movement points
	if err := wms.reduceMovementPoints(entity, from, to, board); err != nil {
		return fmt.Errorf("failed to reduce movement points: %w", err)
	}

	return nil
}

func (wms *WeeWarMovementSystem) getMovementCostFunction(unitType string) func(turnengine.Position) float64 {
	unitData, exists := wms.unitData[unitType]
	if !exists {
		// Return default cost function
		return func(pos turnengine.Position) float64 {
			return 1.0
		}
	}

	return func(pos turnengine.Position) float64 {
		// Get terrain at position
		// Note: This is a simplified version - in a real implementation,
		// we'd need to query the board for terrain at this position
		// For now, return grass cost as default
		if cost, exists := unitData.TerrainMovement["Grass"]; exists {
			return cost
		}
		return 1.0
	}
}

func (wms *WeeWarMovementSystem) getTerrainMovementCost(unitType, terrainType string) float64 {
	unitData, exists := wms.unitData[unitType]
	if !exists {
		return 1.0
	}

	if cost, exists := unitData.TerrainMovement[terrainType]; exists {
		return cost
	}

	return 1.0
}

func (wms *WeeWarMovementSystem) getUnitType(entity *turnengine.Entity) (string, error) {
	unitType, exists := entity.GetComponent("unitType")
	if !exists {
		return "", fmt.Errorf("entity has no unitType component")
	}

	unitTypeName, ok := unitType["unitType"].(string)
	if !ok {
		return "", fmt.Errorf("invalid unitType value")
	}

	return unitTypeName, nil
}

func (wms *WeeWarMovementSystem) getMovementRange(entity *turnengine.Entity) (int, error) {
	movement, exists := entity.GetComponent("movement")
	if !exists {
		return 0, fmt.Errorf("entity has no movement component")
	}

	movesLeft, ok := movement["movesLeft"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid movesLeft value")
	}

	return int(movesLeft), nil
}

func (wms *WeeWarMovementSystem) updateEntityPosition(entity *turnengine.Entity, newPos turnengine.Position) error {
	position, exists := entity.GetComponent("position")
	if !exists {
		return fmt.Errorf("entity has no position component")
	}

	if hexPos, ok := newPos.(*HexPosition); ok {
		position["q"] = float64(hexPos.Q)
		position["r"] = float64(hexPos.R)
		entity.Components["position"] = position
		return nil
	}

	return fmt.Errorf("invalid position type")
}

func (wms *WeeWarMovementSystem) reduceMovementPoints(entity *turnengine.Entity, from, to turnengine.Position, board turnengine.Board) error {
	// Get unit type for movement cost calculation
	unitType, err := wms.getUnitType(entity)
	if err != nil {
		return err
	}

	// Get movement cost function
	costFunc := wms.getMovementCostFunction(unitType)
	
	// Calculate path cost
	path, err := board.(*HexBoard).pathfinder.FindPath(board, from, to, func(pos turnengine.Position) int {
		cost := costFunc(pos)
		if cost < 0 {
			return -1
		}
		return int(cost)
	})
	if err != nil {
		return err
	}

	// Calculate total cost
	totalCost := 0.0
	for i := 1; i < len(path); i++ {
		totalCost += costFunc(path[i])
	}

	// Reduce movement points
	movement, exists := entity.GetComponent("movement")
	if !exists {
		return fmt.Errorf("entity has no movement component")
	}

	movesLeft, ok := movement["movesLeft"].(float64)
	if !ok {
		return fmt.Errorf("invalid movesLeft value")
	}

	newMovesLeft := movesLeft - totalCost
	if newMovesLeft < 0 {
		newMovesLeft = 0
	}

	movement["movesLeft"] = newMovesLeft
	entity.Components["movement"] = movement

	return nil
}

func (wms *WeeWarMovementSystem) ResetMovementPoints(entity *turnengine.Entity) error {
	// Get unit type to determine base movement
	unitType, err := wms.getUnitType(entity)
	if err != nil {
		return err
	}

	// Get base movement from unit data
	unitData, exists := wms.unitData[unitType]
	if !exists {
		return fmt.Errorf("unknown unit type: %s", unitType)
	}

	// Reset movement points to base movement
	movement, exists := entity.GetComponent("movement")
	if !exists {
		return fmt.Errorf("entity has no movement component")
	}

	movement["movesLeft"] = float64(unitData.BaseStats.Movement)
	movement["range"] = float64(unitData.BaseStats.Movement)
	entity.Components["movement"] = movement

	return nil
}

func (wms *WeeWarMovementSystem) GetTerrainMovementCost(unitType, terrainType string) float64 {
	return wms.getTerrainMovementCost(unitType, terrainType)
}