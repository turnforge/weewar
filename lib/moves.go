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
	g.CurrentPlayer = (g.CurrentPlayer + 1) % g.PlayerCount

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
func (g *Game) FindPath(fromRow, fromCol, toRow, toCol int) ([]Tile, error) {
	if g.Map == nil {
		return nil, fmt.Errorf("no map loaded")
	}

	// Check if start and end positions are valid
	startTile := g.GetTileAt(fromRow, fromCol)
	endTile := g.GetTileAt(toRow, toCol)

	if startTile == nil {
		return nil, fmt.Errorf("invalid start position: (%d, %d)", fromRow, fromCol)
	}
	if endTile == nil {
		return nil, fmt.Errorf("invalid end position: (%d, %d)", toRow, toCol)
	}

	// For now, return a simple direct path
	// TODO: Implement proper A* pathfinding
	path := []Tile{*startTile, *endTile}
	return path, nil
}

// IsValidMove checks if movement is legal using cube coordinates
func (g *Game) IsValidMove(from, to CubeCoord) bool {
	// Check if both positions are valid
	startTile := g.Map.TileAt(from)
	endTile := g.Map.TileAt(to)

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

	// For now, allow movement to any adjacent tile
	// TODO: Implement proper movement range and pathfinding validation
	return true
}

// GetMovementCost calculates movement points required using cube coordinates
func (g *Game) GetMovementCost(from, to CubeCoord) int {
	// For now, return a simple cost based on distance
	// TODO: Implement proper terrain-based movement costs
	if from == to {
		return 0
	}

	// Calculate proper hex distance using cube coordinates
	distance := CubeDistance(from, to)

	if distance <= 1 {
		return 1 // Adjacent tiles cost 1 movement point
	}

	return distance // Use proper hex distance calculation
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

	// For now, return a simple range based on unit type
	// TODO: Get from unit data
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
func (g *Game) MoveUnit(unit *Unit, to CubeCoord) error {
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
	currentTile := g.Map.TileAt(unit.Coord)
	if currentTile != nil {
		currentTile.Unit = nil
	}

	// Move unit to new position
	unit.Coord = to
	unit.DistanceLeft -= cost

	// Place unit on new tile
	newTile := g.Map.TileAt(to)
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

	// Calculate damage (simplified combat)
	attackerDamage := 0
	defenderDamage := g.calculateDamage(attacker, defender)

	// Apply damage
	defender.AvailableHealth -= defenderDamage
	if defender.AvailableHealth < 0 {
		defender.AvailableHealth = 0
	}

	// Check if defender was killed
	defenderKilled := defender.AvailableHealth <= 0

	// Remove defender if killed
	if defenderKilled {
		g.RemoveUnit(defender)
	}

	// Create combat result
	result := &CombatResult{
		AttackerDamage: attackerDamage,
		DefenderDamage: defenderDamage,
		AttackerKilled: false,
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
func (g *Game) CanMoveUnit(unit *Unit, to CubeCoord) bool {
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

	// Check if attacker is in range
	distance := g.calculateDistance(attacker.Row, attacker.Col, defender.Row, defender.Col)
	attackRange := g.GetUnitAttackRange(attacker)

	return distance <= attackRange
}

// MoveUnitAt executes unit movement from one coordinate to another
func (g *Game) MoveUnitAt(from, to CubeCoord) error {
	// Find unit at from position
	fromTile := g.Map.TileAt(from)
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
func (g *Game) AttackUnitAt(attackerPos, targetPos CubeCoord) (*CombatResult, error) {
	// Find attacker unit
	attackerTile := g.Map.TileAt(attackerPos)
	if attackerTile == nil {
		return nil, fmt.Errorf("invalid attacker position: %v", attackerPos)
	}
	attacker := attackerTile.Unit
	if attacker == nil {
		return nil, fmt.Errorf("no unit at attacker position %v", attackerPos)
	}

	// Find target unit
	targetTile := g.Map.TileAt(targetPos)
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
func (g *Game) CanAttack(fromRow, fromCol, toRow, toCol int) (bool, error) {
	attacker := g.GetUnitAt(fromRow, fromCol)
	if attacker == nil {
		return false, fmt.Errorf("no unit at attacker position (%d, %d)", fromRow, fromCol)
	}

	defender := g.GetUnitAt(toRow, toCol)
	if defender == nil {
		return false, fmt.Errorf("no unit at target position (%d, %d)", toRow, toCol)
	}

	return g.CanAttackUnit(attacker, defender), nil
}

// CanMove validates potential movement using position coordinates
func (g *Game) CanMove(fromRow, fromCol, toRow, toCol int) (bool, error) {
	unit := g.GetUnitAt(fromRow, fromCol)
	if unit == nil {
		return false, fmt.Errorf("no unit at position (%d, %d)", fromRow, fromCol)
	}

	return g.CanMoveUnit(unit, toRow, toCol), nil
}

// calculateDamage calculates damage dealt in combat (simplified)
func (g *Game) calculateDamage(attacker, defender *Unit) int {
	// Simplified damage calculation
	// TODO: Implement proper damage calculation based on unit types, terrain, etc.

	baseDamage := 30

	// Add some randomness
	variation := g.rng.Intn(20) - 10 // -10 to +10
	damage := baseDamage + variation

	if damage < 10 {
		damage = 10 // Minimum damage
	}

	return damage
}

// calculateDistance calculates distance between two positions
func (g *Game) calculateDistance(row1, col1, row2, col2 int) int {
	// Simplified hex distance calculation
	// TODO: Implement proper hex distance calculation
	dRow := abs(row2 - row1)
	dCol := abs(col2 - col1)

	return max(dRow, dCol)
}
