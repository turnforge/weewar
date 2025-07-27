package weewar

import (
	"fmt"
	"time"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

type MoveProcessor interface {
	ProcessMoves(game *Game, moves []*v1.GameMove) (results *v1.GameMoveResult, err error)
}

type DefaultMoveProcessor struct {
}

// Process a set of moves in a transaction and returns a "log entry" of the changes as a result
func (m *DefaultMoveProcessor) ProcessMoves(game *Game, moves []*v1.GameMove) (results []*v1.GameMoveResult, err error) {
	results = []*v1.GameMoveResult{}
	for _, move := range moves {
		result, err := m.ProcessMove(game, move)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return
}

// This is the dispatcher for a move
// The moves work is we submit a move to the game, it calls the correct move handler
// Moves in a game are "known" so we can have a simple static dispatcher here
// The move handler/processor update the Game state and also updates the action object
// indicating changes that were incurred as part of running the move.  Note that
// since we are looking at "transactionality" in games we want to make sure all moves
// are first valid and ATOMIC and only then finally commit the changes for all the moves.
// For example we may have 3 moves where first two units are moved to a common location
// and then they attack another unit.  Here If we treat it as a single unit attacking it
// will have different outcomes than a "combined" attack.
func (m *DefaultMoveProcessor) ProcessMove(game *Game, move *v1.GameMove) (results *v1.GameMoveResult, err error) {
	switch a := move.MoveType.(type) {
	case *v1.GameMove_MoveUnit:
		return m.ProcessMoveUnit(game, move, a.MoveUnit)
	case *v1.GameMove_AttackUnit:
		return m.ProcessAttackUnit(game, move, a.AttackUnit)
	case *v1.GameMove_EndTurn:
		return m.ProcessEndTurn(game, move, a.EndTurn)
	}
	return nil, nil
}

// EndTurn advances to next player's turn
// For now a player can just end turn but in other games there may be some mandatory
// moves left
func (m *DefaultMoveProcessor) ProcessEndTurn(g *Game, move *v1.GameMove, action *v1.EndTurnAction) (results *v1.GameMoveResult, err error) {
	// Store previous state for GameLog
	// TODO - use a pushed world at ProcessMoves level instead of g.World each time
	previousPlayer := g.CurrentPlayer
	previousTurn := g.TurnCounter

	// Reset unit movement for current player
	if err := g.resetPlayerUnits(g.CurrentPlayer); err != nil {
		return nil, fmt.Errorf("failed to reset player units: %w", err)
	}

	// Advance to next player
	g.CurrentPlayer = (g.CurrentPlayer + 1) % g.World.PlayerCount()

	// If we've cycled back to player 0, increment turn counter
	if g.CurrentPlayer == 0 {
		g.TurnCounter++
	}

	// Check for victory conditions
	if winner, hasWinner := g.checkVictoryConditions(); hasWinner {
		g.winner = winner
		g.hasWinner = true
		g.Status = GameStatusEnded

		// Update GameLog status when game ends
		// TODO - g.SetGameLogStatus("completed")
	}

	// Update timestamp
	g.LastActionAt = time.Now()
	change := &v1.WorldChange{
		ChangeType: &v1.WorldChange_PlayerChanged{
			PlayerChanged: &v1.PlayerChangedChange{
				PreviousPlayer: int32(previousPlayer),
				NewPlayer:      int32(g.CurrentPlayer),
				PreviousTurn:   int32(previousTurn),
				NewTurn:        int32(g.TurnCounter),
			},
		},
	}

	results.Changes = append(results.Changes, change)

	return
}

// CanEndTurn checks if current player can end their turn
/*
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
*/

// IsValidMove checks if movement is legal using cube coordinates
func (g *Game) IsValidMove(from, to AxialCoord) bool {
	// Get the unit at the starting position
	unit := g.World.UnitAt(from)
	if unit == nil {
		return false
	}

	// Check if the unit belongs to the current player
	if unit.Player != g.CurrentPlayer {
		return false
	}

	// Create simple two-tile path and validate using RulesEngine
	path := []AxialCoord{from, to}
	valid, err := g.rulesEngine.IsValidPath(unit, path, g.World)
	if err != nil {
		return false
	}

	return valid
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
func (m *DefaultMoveProcessor) ProcessMoveUnit(g *Game, move *v1.GameMove, action *v1.MoveUnitAction) (result *v1.GameMoveResult, err error) {
	// TODO - use a pushed world at ProcessMoves level instead of g.World each time
	from := CoordFromInt32(action.FromQ, action.FromR)
	to := CoordFromInt32(action.ToQ, action.ToR)
	unit := g.World.UnitAt(from)
	if unit == nil {
		return nil, fmt.Errorf("unit is nil")
	}

	// Check if it's the correct player's turn
	if unit.Player != g.CurrentPlayer {
		return nil, fmt.Errorf("not player %d's turn", unit.Player)
	}

	// Check if move is valid
	if !g.IsValidMove(unit.Coord, to) {
		return nil, fmt.Errorf("invalid move from %v to %v", unit.Coord, to)
	}

	// Get movement cost using RulesEngine
	costFloat, err := g.rulesEngine.GetMovementCost(g.World, unit, to)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate movement cost: %w", err)
	}
	cost := int(costFloat + 0.5) // Round to nearest integer
	if cost > unit.DistanceLeft {
		return nil, fmt.Errorf("insufficient movement points: need %d, have %d", cost, unit.DistanceLeft)
	}

	// Store original position for event
	fromPos := unit.Coord
	toPos := to

	// Move unit using World unit management
	err = g.World.MoveUnit(unit, to)
	if err != nil {
		return nil, fmt.Errorf("failed to move unit: %w", err)
	}

	// Update unit stats
	unit.DistanceLeft -= cost

	// Update timestamp
	g.LastActionAt = time.Now()

	// Record action in GameLog
	change := &v1.WorldChange{
		ChangeType: &v1.WorldChange_UnitMoved{
			UnitMoved: &v1.UnitMovedChange{
				FromQ: int32(fromPos.Q),
				FromR: int32(fromPos.R),
				ToQ:   int32(toPos.Q),
				ToR:   int32(toPos.R),
			},
		},
	}

	result.Changes = append(result.Changes, change)
	return result, nil
}

// AttackUnit executes combat between units
func (m *DefaultMoveProcessor) ProcessAttackUnit(g *Game, move *v1.GameMove, action *v1.AttackUnitAction) (result *v1.GameMoveResult, err error) {
	// TODO - use a pushed world at ProcessMoves level instead of g.World each time
	attacker := g.World.UnitAt(CoordFromInt32(action.AttackerQ, action.AttackerR))
	defender := g.World.UnitAt(CoordFromInt32(action.DefenderQ, action.DefenderR))
	if attacker == nil || defender == nil {
		return nil, fmt.Errorf("attacker or defender is nil")
	}

	// Check if it's the correct player's turn
	if attacker.Player != g.CurrentPlayer {
		return nil, fmt.Errorf("not player %d's turn", attacker.Player)
	}

	// Check if units can attack each other
	if !g.CanAttackUnit(attacker, defender) {
		return nil, fmt.Errorf("attacker cannot attack defender")
	}

	// Calculate damage using rules engine
	attackerDamage := 0
	defenderDamage := 0

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
		g.World.RemoveUnit(defender)
	}
	if attackerKilled {
		g.World.RemoveUnit(attacker)
	}
	/*
		change := &v1.WorldChange{
			ChangeType: &v1.WorldChange_PlayerChanged{
				PlayerChanged: &v1.PlayerChangedChange{
					PreviousPlayer: int32(previousPlayer),
					NewPlayer:      int32(g.CurrentPlayer),
					PreviousTurn:   int32(previousTurn),
					NewTurn:        int32(g.TurnCounter),
				},
			},
		}

			results.Changes = append(results.Changes, change)

			// Create combat result
			result := &CombatResult{
				AttackerDamage: attackerDamage,
				DefenderDamage: defenderDamage,
				AttackerKilled: attackerKilled,
				DefenderKilled: defenderKilled,
				AttackerHealth: attacker.AvailableHealth,
				DefenderHealth: defender.AvailableHealth,
			}
	*/

	// Update timestamp
	g.LastActionAt = time.Now()

	/*
		// Add damage changes
		if defenderDamage > 0 {
			changes = append(changes, WorldChange{
				Type:       "unitDamaged",
				EntityType: "unit",
				EntityID:   fmt.Sprintf("unit_%d_%d", defender.Player, defender.UnitType),
				FromState:  map[string]interface{}{"health": defender.AvailableHealth + defenderDamage},
				ToState:    map[string]interface{}{"health": defender.AvailableHealth},
			})
		}

		if attackerDamage > 0 {
			changes = append(changes, WorldChange{
				Type:       "unitDamaged",
				EntityType: "unit",
				EntityID:   fmt.Sprintf("unit_%d_%d", attacker.Player, attacker.UnitType),
				FromState:  map[string]interface{}{"health": attacker.AvailableHealth + attackerDamage},
				ToState:    map[string]interface{}{"health": attacker.AvailableHealth},
			})
		}

		// Add kill changes
		if defenderKilled {
			changes = append(changes, CreateUnitKilledChange(
				fmt.Sprintf("unit_%d_%d", defender.Player, defender.UnitType),
				map[string]interface{}{
					"player":   defender.Player,
					"unitType": defender.UnitType,
					"position": defender.Coord,
					"health":   defender.AvailableHealth + defenderDamage,
				},
			))
		}

		if attackerKilled {
			changes = append(changes, CreateUnitKilledChange(
				fmt.Sprintf("unit_%d_%d", attacker.Player, attacker.UnitType),
				map[string]interface{}{
					"player":   attacker.Player,
					"unitType": attacker.UnitType,
					"position": attacker.Coord,
					"health":   attacker.AvailableHealth + attackerDamage,
				},
			))
		}
	*/

	return nil, nil
}

// CanMoveUnit validates potential movement using cube coordinates
func (g *Game) CanMoveUnit(unit *Unit, to AxialCoord) bool {
	if unit == nil {
		return false
	}

	// Check if it's the correct player's turn
	if unit.Player != g.CurrentPlayer {
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
	if attacker.Player != g.CurrentPlayer {
		return false
	}

	// Check if units are enemies
	if attacker.Player == defender.Player {
		return false
	}

	// Use rules engine for attack validation
	canAttack, err := g.rulesEngine.CanUnitAttackTarget(attacker, defender)
	if err != nil {
		return false
	}
	return canAttack
}

// AttackUnitAt executes combat between units at the given coordinates
func (g *Game) AttackUnitAt(attackerPos, targetPos AxialCoord) (*CombatResult, error) {
	// Find attacker unit using World
	attacker := g.World.UnitAt(attackerPos)
	if attacker == nil {
		return nil, fmt.Errorf("no unit at attacker position %v", attackerPos)
	}

	// Find target unit using World
	target := g.World.UnitAt(targetPos)
	if target == nil {
		return nil, fmt.Errorf("no unit at target position %v", targetPos)
	}

	// Use existing AttackUnit method
	return nil, nil // g.AttackUnit(attacker, target)
}

// CanAttack validates potential attack using position coordinates
func (g *Game) CanAttack(from, to AxialCoord) (bool, error) {
	attacker := g.World.UnitAt(from)
	if attacker == nil {
		return false, fmt.Errorf("no unit at attacker position (%d, %d)", from.Q, from.R)
	}

	defender := g.World.UnitAt(to)
	if defender == nil {
		return false, fmt.Errorf("no unit at target position (%d, %d)", to.Q, to.R)
	}

	return g.CanAttackUnit(attacker, defender), nil
}

// CanMove validates potential movement using position coordinates
func (g *Game) CanMove(from, to Position) (bool, error) {
	unit := g.World.UnitAt(from)
	return g.CanMoveUnit(unit, to), nil
}

// calculateDistance calculates distance between two positions
// Source: https://www.redblobgames.com/grids/hexagons-v1/#distances
func (g *Game) calculateDistance(a, b AxialCoord) int {
	// Simplified hex distance calculation
	return (abs(a.Q-b.Q) + abs(a.Q+a.R-b.Q-b.R) + abs(a.R-b.R)) / 2
}

// GetUnitMovementOptions returns all tiles a unit can move to using rules engine
func (g *Game) GetUnitMovementOptionsFrom(q, r int) ([]TileOption, error) {
	return g.GetUnitMovementOptions(g.World.UnitAt(AxialCoord{q, r}))
}

// GetUnitMovementOptions returns all tiles a unit can move to using rules engine
func (g *Game) GetUnitMovementOptions(unit *Unit) ([]TileOption, error) {
	dl := 0
	if unit != nil {
		dl = unit.DistanceLeft
	}
	return g.rulesEngine.GetMovementOptions(g.World, unit, dl)
}

// GetUnitAttackOptions returns all positions a unit can attack using rules engine
func (g *Game) GetUnitAttackOptionsFrom(q, r int) ([]AxialCoord, error) {
	return g.GetUnitAttackOptions(g.World.UnitAt(AxialCoord{q, r}))
}
func (g *Game) GetUnitAttackOptions(unit *Unit) ([]AxialCoord, error) {
	return g.rulesEngine.GetAttackOptions(g.World, unit)
}

/*
// CreateAttackAction creates a standardized attack action
func CreateAttackAction(attackerQ, attackerR, defenderQ, defenderR int) GameAction {
	return GameAction{
		Type: "attack",
		Params: map[string]interface{}{
			"attackerQ": attackerQ,
			"attackerR": attackerR,
			"defenderQ": defenderQ,
			"defenderR": defenderR,
		},
	}
}

// CreateUnitMovedChange creates a standardized unit moved change
func CreateUnitMovedChange(unitID string, fromQ, fromR, toQ, toR int) WorldChange {
	return WorldChange{
		Type:       "unitMoved",
		EntityType: "unit",
		EntityID:   unitID,
		FromState: map[string]interface{}{
			"q": fromQ,
			"r": fromR,
		},
		ToState: map[string]interface{}{
			"q": toQ,
			"r": toR,
		},
	}
}

// CreateUnitKilledChange creates a standardized unit killed change
func CreateUnitKilledChange(unitID string, unitData interface{}) WorldChange {
	return WorldChange{
		Type:       "unitKilled",
		EntityType: "unit",
		EntityID:   unitID,
		FromState:  unitData,
		ToState:    nil,
	}
}

// CreatePlayerChangedChange creates a standardized player changed change
func CreatePlayerChangedChange(fromPlayer, toPlayer int) WorldChange {
	return WorldChange{
		Type:       "playerChanged",
		EntityType: "game",
		EntityID:   "currentPlayer",
		FromState: map[string]interface{}{
			"player": fromPlayer,
		},
		ToState: map[string]interface{}{
			"player": toPlayer,
		},
	}
}

// CreateTurnAdvancedChange creates a standardized turn advanced change
func CreateTurnAdvancedChange(fromTurn, toTurn int) WorldChange {
	return WorldChange{
		Type:       "turnAdvanced",
		EntityType: "game",
		EntityID:   "turnCounter",
		FromState: map[string]interface{}{
			"turn": fromTurn,
		},
		ToState: map[string]interface{}{
			"turn": toTurn,
		},
	}
}
*/
