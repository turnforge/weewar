package weewar

import (
	"fmt"
	"time"
)

// NextTurn advances to next player's turn
func (g *Game) NextTurn() error {
	if g.Status != GameStatusPlaying {
		return fmt.Errorf("cannot advance turn: game is not in playing state")
	}

	// Reset unit movement for current player
	if err := g.resetPlayerUnits(g.CurrentPlayer); err != nil {
		return fmt.Errorf("failed to reset player units: %w", err)
	}

	// Advance to next player
	g.CurrentPlayer = (g.CurrentPlayer + 1) % g.PlayerCount()

	// If we've cycled back to player 0, increment turn counter
	if g.CurrentPlayer == 0 {
		g.TurnCounter++
	}

	// Check for victory conditions
	if winner, hasWinner := g.checkVictoryConditions(); hasWinner {
		g.winner = winner
		g.hasWinner = true
		g.Status = GameStatusEnded
		g.eventManager.EmitGameEnded(winner)
		g.eventManager.EmitGameStateChanged(GameStateChangeGameEnded, winner)
	}

	// Update timestamp
	g.LastActionAt = time.Now()

	// Emit turn changed event
	g.eventManager.EmitTurnChanged(g.CurrentPlayer, g.TurnCounter)
	g.eventManager.EmitGameStateChanged(GameStateChangeTurnChanged, map[string]interface{}{
		"newPlayer":  g.CurrentPlayer,
		"turnNumber": g.TurnCounter,
	})

	return nil
}

// EndTurn completes current player's turn
func (g *Game) EndTurn() error {
	if g.Status != GameStatusPlaying {
		return fmt.Errorf("cannot end turn: game is not in playing state")
	}

	// For now, EndTurn is the same as NextTurn
	// In a full implementation, this might involve different logic
	// (e.g., checking if player has mandatory actions to complete)
	return g.NextTurn()
}

// CanEndTurn checks if current player can end their turn
func (g *Game) CanEndTurn() bool {
	if g.Status != GameStatusPlaying {
		return false
	}

	// For now, player can always end turn
	// In a full implementation, this might check:
	// - Whether player has units that must move
	// - Whether player has mandatory actions to complete
	// - Whether player has captured a base this turn
	return true
}

// FindPath calculates movement path between positions
func (g *Game) FindPath(from, to Position) ([]Tile, error) {
	if g.World.Map == nil {
		return nil, fmt.Errorf("no map loaded")
	}

	// Check if start and end positions are valid
	startTile := g.World.Map.TileAt(from)
	endTile := g.World.Map.TileAt(to)

	if startTile == nil {
		return nil, fmt.Errorf("invalid start position: %s", from)
	}
	if endTile == nil {
		return nil, fmt.Errorf("invalid end position: %s", to)
	}

	// For now, return a simple direct path
	// TODO: Implement proper A* pathfinding
	path := []Tile{*startTile, *endTile}
	return path, nil
}

// IsValidMove checks if movement is legal using cube coordinates
func (g *Game) IsValidMove(from, to AxialCoord) bool {
	// Check if both positions are valid
	startTile := g.World.Map.TileAt(from)
	endTile := g.World.Map.TileAt(to)

	if startTile == nil || endTile == nil {
		return false
	}

	// Check if there's a unit at the start position
	if startTile.Unit == nil {
		return false
	}

	// Check if the unit belongs to the current player
	if startTile.Unit.PlayerID != g.CurrentPlayer {
		return false
	}

	// Check if destination is empty
	if endTile.Unit != nil {
		return false
	}

	// Check if unit has movement left
	if startTile.Unit.DistanceLeft <= 0 {
		return false
	}

	// Check if the unit can move on the target terrain
	_, err := g.rulesEngine.GetTerrainMovementCost(startTile.Unit.UnitType, endTile.TileType)
	if err != nil {
		return false // Cannot move on this terrain type
	}

	// Check if destination is within movement range using rules engine
	cost, err := g.rulesEngine.GetMovementCost(g.World.Map, startTile.Unit.UnitType, from, to)
	if err != nil || cost > float64(startTile.Unit.DistanceLeft) {
		return false
	}

	return true
}

// GetMovementCost calculates movement points required using cube coordinates
func (g *Game) GetMovementCost(from, to AxialCoord) int {
	if from == to {
		return 0
	}

	// Get the unit at the from position to determine unit type
	fromTile := g.World.Map.TileAt(from)
	if fromTile == nil || fromTile.Unit == nil {
		return CubeDistance(from, to) // Fallback to distance if no unit
	}

	// Use rules engine for terrain-specific movement cost calculation
	if cost, err := g.rulesEngine.GetMovementCost(g.World.Map, fromTile.Unit.UnitType, from, to); err == nil {
		return int(cost + 0.5) // Round to nearest integer
	}

	// Fallback to simple distance calculation
	return CubeDistance(from, to)
}

// GetUnitMovementLeft returns remaining movement points
func (g *Game) GetUnitMovementLeft(unit *Unit) int {
	if unit == nil {
		return 0
	}
	return unit.DistanceLeft
}

// GetUnitAttackRange returns attack range in tiles
func (g *Game) GetUnitAttackRange(unit *Unit) int {
	if unit == nil {
		return 0
	}

	// Use rules engine to get unit data
	if unitData, err := g.rulesEngine.GetUnitData(unit.UnitType); err == nil {
		return unitData.AttackRange
	}

	// Fallback to simple range based on unit type
	switch unit.UnitType {
	case 1: // Infantry
		return 1
	case 2: // Artillery
		return 3
	case 3: // Tank
		return 1
	default:
		return 1
	}
}

// MoveUnit executes unit movement using cube coordinates
func (g *Game) MoveUnit(unit *Unit, to AxialCoord) error {
	if unit == nil {
		return fmt.Errorf("unit is nil")
	}

	// Check if it's the correct player's turn
	if unit.PlayerID != g.CurrentPlayer {
		return fmt.Errorf("not player %d's turn", unit.PlayerID)
	}

	// Check if move is valid
	if !g.IsValidMove(unit.Coord, to) {
		return fmt.Errorf("invalid move from %v to %v", unit.Coord, to)
	}

	// Get movement cost
	cost := g.GetMovementCost(unit.Coord, to)
	if cost > unit.DistanceLeft {
		return fmt.Errorf("insufficient movement points: need %d, have %d", cost, unit.DistanceLeft)
	}

	// Store original position for event
	fromPos := unit.Coord
	toPos := to

	// Remove unit from current tile
	currentTile := g.World.Map.TileAt(unit.Coord)
	if currentTile != nil {
		currentTile.Unit = nil
	}

	// Move unit to new position
	unit.Coord = to
	unit.DistanceLeft -= cost

	// Place unit on new tile
	newTile := g.World.Map.TileAt(to)
	if newTile != nil {
		newTile.Unit = unit
	}

	// Update timestamp
	g.LastActionAt = time.Now()

	// Emit events
	g.eventManager.EmitUnitMoved(unit, fromPos, toPos)
	g.eventManager.EmitGameStateChanged(GameStateChangeUnitMoved, map[string]interface{}{
		"unit": unit,
		"from": fromPos,
		"to":   toPos,
	})

	return nil
}

// AttackUnit executes combat between units
func (g *Game) AttackUnit(attacker, defender *Unit) (*CombatResult, error) {
	if attacker == nil || defender == nil {
		return nil, fmt.Errorf("attacker or defender is nil")
	}

	// Check if it's the correct player's turn
	if attacker.PlayerID != g.CurrentPlayer {
		return nil, fmt.Errorf("not player %d's turn", attacker.PlayerID)
	}

	// Check if units can attack each other
	if !g.CanAttackUnit(attacker, defender) {
		return nil, fmt.Errorf("attacker cannot attack defender")
	}

	// Calculate damage using rules engine
	attackerDamage := 0
	defenderDamage := 0
	
	var err error
	defenderDamage, err = g.rulesEngine.CalculateCombatDamage(attacker.UnitType, defender.UnitType, g.rng)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate combat damage: %w", err)
	}
	
	// Check if defender can counter-attack
	if canCounter, err := g.rulesEngine.CanUnitAttackTarget(defender, attacker); err == nil && canCounter {
		attackerDamage, err = g.rulesEngine.CalculateCombatDamage(defender.UnitType, attacker.UnitType, g.rng)
		if err != nil {
			// If counter-attack calculation fails, no counter damage
			attackerDamage = 0
		}
	}

	// Apply damage
	defender.AvailableHealth -= defenderDamage
	if defender.AvailableHealth < 0 {
		defender.AvailableHealth = 0
	}
	
	attacker.AvailableHealth -= attackerDamage
	if attacker.AvailableHealth < 0 {
		attacker.AvailableHealth = 0
	}

	// Check if units were killed
	defenderKilled := defender.AvailableHealth <= 0
	attackerKilled := attacker.AvailableHealth <= 0

	// Remove killed units
	if defenderKilled {
		g.RemoveUnit(defender)
	}
	if attackerKilled {
		g.RemoveUnit(attacker)
	}

	// Create combat result
	result := &CombatResult{
		AttackerDamage: attackerDamage,
		DefenderDamage: defenderDamage,
		AttackerKilled: attackerKilled,
		DefenderKilled: defenderKilled,
		AttackerHealth: attacker.AvailableHealth,
		DefenderHealth: defender.AvailableHealth,
	}

	// Update timestamp
	g.LastActionAt = time.Now()

	// Emit events
	g.eventManager.EmitUnitAttacked(attacker, defender, result)
	g.eventManager.EmitGameStateChanged(GameStateChangeUnitAttacked, map[string]interface{}{
		"attacker": attacker,
		"defender": defender,
		"result":   result,
	})

	return result, nil
}

// CanMoveUnit validates potential movement using cube coordinates
func (g *Game) CanMoveUnit(unit *Unit, to AxialCoord) bool {
	if unit == nil {
		return false
	}

	// Check if it's the correct player's turn
	if unit.PlayerID != g.CurrentPlayer {
		return false
	}

	// Check if move is valid
	return g.IsValidMove(unit.Coord, to)
}

// CanAttackUnit validates potential attack
func (g *Game) CanAttackUnit(attacker, defender *Unit) bool {
	if attacker == nil || defender == nil {
		return false
	}

	// Check if it's the correct player's turn
	if attacker.PlayerID != g.CurrentPlayer {
		return false
	}

	// Check if units are enemies
	if attacker.PlayerID == defender.PlayerID {
		return false
	}

	// Use rules engine for attack validation
	canAttack, err := g.rulesEngine.CanUnitAttackTarget(attacker, defender)
	if err != nil {
		return false
	}
	return canAttack
}

// MoveUnitAt executes unit movement from one coordinate to another
func (g *Game) MoveUnitAt(from, to AxialCoord) error {
	// Find unit at from position
	fromTile := g.World.Map.TileAt(from)
	if fromTile == nil {
		return fmt.Errorf("invalid from position: %v", from)
	}
	unit := fromTile.Unit
	if unit == nil {
		return fmt.Errorf("no unit at position %v", from)
	}
	// Use existing MoveUnit method
	return g.MoveUnit(unit, to)
}

// AttackUnitAt executes combat between units at the given coordinates
func (g *Game) AttackUnitAt(attackerPos, targetPos AxialCoord) (*CombatResult, error) {
	// Find attacker unit
	attackerTile := g.World.Map.TileAt(attackerPos)
	if attackerTile == nil {
		return nil, fmt.Errorf("invalid attacker position: %v", attackerPos)
	}
	attacker := attackerTile.Unit
	if attacker == nil {
		return nil, fmt.Errorf("no unit at attacker position %v", attackerPos)
	}

	// Find target unit
	targetTile := g.World.Map.TileAt(targetPos)
	if targetTile == nil {
		return nil, fmt.Errorf("invalid target position: %v", targetPos)
	}
	target := targetTile.Unit
	if target == nil {
		return nil, fmt.Errorf("no unit at target position %v", targetPos)
	}

	// Use existing AttackUnit method
	return g.AttackUnit(attacker, target)
}

// CanAttack validates potential attack using position coordinates
func (g *Game) CanAttack(from, to AxialCoord) (bool, error) {
	attacker := g.GetUnitAt(from)
	if attacker == nil {
		return false, fmt.Errorf("no unit at attacker position (%d, %d)", from.Q, from.R)
	}

	defender := g.GetUnitAt(to)
	if defender == nil {
		return false, fmt.Errorf("no unit at target position (%d, %d)", to.Q, to.R)
	}

	return g.CanAttackUnit(attacker, defender), nil
}

// CanMove validates potential movement using position coordinates
func (g *Game) CanMove(from, to Position) (bool, error) {
	unit := g.GetUnitAt(from)
	if unit == nil {
		return false, fmt.Errorf("no unit at position (%d, %d)", from.Q, from.R)
	}

	return g.CanMoveUnit(unit, to), nil
}


// calculateDistance calculates distance between two positions
// Source: https://www.redblobgames.com/grids/hexagons-v1/#distances
func (g *Game) calculateDistance(a, b AxialCoord) int {
	// Simplified hex distance calculation
	return (abs(a.Q-b.Q) + abs(a.Q+a.R-b.Q-b.R) + abs(a.R-b.R)) / 2
}

// GetUnitMovementOptions returns all tiles a unit can move to using rules engine
func (g *Game) GetUnitMovementOptions(unit *Unit) ([]TileOption, error) {
	if unit == nil {
		return nil, fmt.Errorf("unit is nil")
	}

	return g.rulesEngine.GetMovementOptions(g.World.Map, unit, unit.DistanceLeft)
}

// GetUnitAttackOptions returns all positions a unit can attack using rules engine  
func (g *Game) GetUnitAttackOptions(unit *Unit) ([]AxialCoord, error) {
	if unit == nil {
		return nil, fmt.Errorf("unit is nil")
	}

	return g.rulesEngine.GetAttackOptions(g.World.Map, unit)
}
