package services

import (
	"fmt"
	"time"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

type MoveProcessor struct {
}

// copyUnit creates a deep copy of a unit with all fields
// This is used when recording unit states in WorldChange objects
func copyUnit(unit *v1.Unit) *v1.Unit {
	if unit == nil {
		return nil
	}
	return &v1.Unit{
		Q:                unit.Q,
		R:                unit.R,
		Player:           unit.Player,
		UnitType:         unit.UnitType,
		Shortcut:         unit.Shortcut,
		AvailableHealth:  unit.AvailableHealth,
		DistanceLeft:     unit.DistanceLeft,
		LastActedTurn:    unit.LastActedTurn,
		LastToppedupTurn: unit.LastToppedupTurn,
	}
}

// Process a set of moves in a transaction and returns a "log entry" of the changes as a result
func (m *MoveProcessor) ProcessMoves(game *Game, moves []*v1.GameMove) (results []*v1.GameMoveResult, err error) {
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
func (m *MoveProcessor) ProcessMove(game *Game, move *v1.GameMove) (results *v1.GameMoveResult, err error) {
	if move.MoveType == nil {
		return nil, fmt.Errorf("move type is nil")
	}

	switch a := move.MoveType.(type) {
	case *v1.GameMove_MoveUnit:
		return m.ProcessMoveUnit(game, move, a.MoveUnit)
	case *v1.GameMove_AttackUnit:
		return m.ProcessAttackUnit(game, move, a.AttackUnit)
	case *v1.GameMove_EndTurn:
		return m.ProcessEndTurn(game, move, a.EndTurn)
	default:
		return nil, fmt.Errorf("unknown move type: %T", move.MoveType)
	}
}

// EndTurn advances to next player's turn
// For now a player can just end turn but in other games there may be some mandatory
// moves left
func (m *MoveProcessor) ProcessEndTurn(g *Game, move *v1.GameMove, action *v1.EndTurnAction) (results *v1.GameMoveResult, err error) {
	// Initialize the result object
	results = &v1.GameMoveResult{
		IsPermanent: false,
		SequenceNum: 0, // TODO: Set proper sequence number
		Changes:     []*v1.WorldChange{},
	}

	// Store previous state for GameLog
	// TODO - use a pushed world at ProcessMoves level instead of g.World each time
	previousPlayer := g.CurrentPlayer
	previousTurn := g.TurnCounter

	// Capture the reset units AFTER reset (with refreshed movement points)
	var resetUnits []*v1.Unit
	playerUnits := g.World.GetPlayerUnits(int(previousPlayer))

	for _, unit := range playerUnits {
		fmt.Printf("ProcessEndTurn: Adding resetUnit at (%d, %d) player=%d, distanceLeft=%d\n",
			unit.Q, unit.R, unit.Player, unit.DistanceLeft)
		resetUnit := copyUnit(unit)
		resetUnits = append(resetUnits, resetUnit)
	}

	// Advance to next player (1-based player system: Player 1, Player 2, etc.)
	// Player 0 is reserved for neutral, so we cycle between 1, 2, ..., PlayerCount
	numPlayers := g.World.PlayerCount()

	if g.CurrentPlayer == numPlayers {
		// Last player completes their turn, go back to player 1 and increment turn counter
		g.CurrentPlayer = 1
		g.TurnCounter++
	} else {
		// Move to next player
		g.CurrentPlayer++
	}

	// Check for victory conditions
	if winner, hasWinner := g.checkVictoryConditions(); hasWinner {
		g.GameState.WinningPlayer = winner
		g.GameState.Finished = true
		g.GameState.Status = v1.GameStatus_GAME_STATUS_ENDED

		// Update GameLog status when game ends
		// TODO - g.SetGameLogStatus("completed")
	}

	// Update timestamp
	g.GameState.UpdatedAt = tspb.New(time.Now())
	change := &v1.WorldChange{
		ChangeType: &v1.WorldChange_PlayerChanged{
			PlayerChanged: &v1.PlayerChangedChange{
				PreviousPlayer: int32(previousPlayer),
				NewPlayer:      int32(g.CurrentPlayer),
				PreviousTurn:   int32(previousTurn),
				NewTurn:        int32(g.TurnCounter),
				ResetUnits:     resetUnits,
			},
		},
	}

	results.Changes = append(results.Changes, change)

	return
}

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

// MoveUnit executes unit movement using cube coordinates
func (m *MoveProcessor) ProcessMoveUnit(g *Game, move *v1.GameMove, action *v1.MoveUnitAction) (result *v1.GameMoveResult, err error) {
	// Initialize the result object
	result = &v1.GameMoveResult{
		IsPermanent: false,
		SequenceNum: 0, // TODO: Set proper sequence number
		Changes:     []*v1.WorldChange{},
	}

	// TODO - use a pushed world at ProcessMoves level instead of g.World each time
	from := CoordFromInt32(action.FromQ, action.FromR)
	to := CoordFromInt32(action.ToQ, action.ToR)
	unit := g.World.UnitAt(from)
	if unit == nil {
		return nil, fmt.Errorf("unit is nil")
	}

	// Apply lazy top-up pattern - ensure unit has current turn's movement points
	if err := g.topUpUnitIfNeeded(unit); err != nil {
		return nil, fmt.Errorf("failed to top-up unit: %w", err)
	}

	// Check if it's the correct player's turn
	if unit.Player != g.CurrentPlayer {
		return nil, fmt.Errorf("not player %d's turn", unit.Player)
	}

	// Check if move is valid
	unitCoord := UnitGetCoord(unit)
	if !g.IsValidMove(unitCoord, to) {
		return nil, fmt.Errorf("invalid move from %v to %v", unitCoord, to)
	}

	// Get movement cost using RulesEngine
	costFloat, err := g.rulesEngine.GetMovementCost(g.World, unit, to)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate movement cost: %w", err)
	}
	cost := int(costFloat + 0.5) // Round to nearest integer
	if cost > int(unit.DistanceLeft) {
		return nil, fmt.Errorf("insufficient movement points: need %d, have %d", cost, unit.DistanceLeft)
	}

	// Capture unit state before move
	previousUnit := copyUnit(unit)

	// Move unit using World unit management
	err = g.World.MoveUnit(unit, to)
	if err != nil {
		return nil, fmt.Errorf("failed to move unit: %w", err)
	}

	// Get the moved unit from the world (handles copy-on-write correctly)
	movedUnit := g.World.UnitAt(to)
	if movedUnit == nil {
		return nil, fmt.Errorf("moved unit not found at destination %v", to)
	}

	// Update unit stats on the moved unit
	movedUnit.DistanceLeft -= int32(cost)

	// Capture unit state after move (using the moved unit, not the original)
	updatedUnit := copyUnit(movedUnit)
	updatedUnit.LastActedTurn = unit.LastActedTurn
	updatedUnit.LastToppedupTurn = unit.LastToppedupTurn

	// Update timestamp
	g.GameState.UpdatedAt = tspb.New(time.Now())

	// Record action in GameLog
	change := &v1.WorldChange{
		ChangeType: &v1.WorldChange_UnitMoved{
			UnitMoved: &v1.UnitMovedChange{
				PreviousUnit: previousUnit,
				UpdatedUnit:  updatedUnit,
			},
		},
	}

	result.Changes = append(result.Changes, change)
	return result, nil
}

// AttackUnit executes combat between units
func (m *MoveProcessor) ProcessAttackUnit(g *Game, move *v1.GameMove, action *v1.AttackUnitAction) (result *v1.GameMoveResult, err error) {
	// Initialize the result object
	result = &v1.GameMoveResult{
		IsPermanent: true, // Attacks are permanent (cannot be undone)
		SequenceNum: 0,    // TODO: Set proper sequence number
		Changes:     []*v1.WorldChange{},
	}

	// TODO - use a pushed world at ProcessMoves level instead of g.World each time
	attacker := g.World.UnitAt(CoordFromInt32(action.AttackerQ, action.AttackerR))
	defender := g.World.UnitAt(CoordFromInt32(action.DefenderQ, action.DefenderR))
	if attacker == nil || defender == nil {
		return nil, fmt.Errorf("attacker or defender is nil")
	}

	// Apply lazy top-up pattern for both units
	if err := g.topUpUnitIfNeeded(attacker); err != nil {
		return nil, fmt.Errorf("failed to top-up attacker: %w", err)
	}
	if err := g.topUpUnitIfNeeded(defender); err != nil {
		return nil, fmt.Errorf("failed to top-up defender: %w", err)
	}

	// Check if it's the correct player's turn
	if attacker.Player != g.CurrentPlayer {
		return nil, fmt.Errorf("not player %d's turn", attacker.Player)
	}

	// Check if units can attack each other
	if !g.CanAttackUnit(attacker, defender) {
		return nil, fmt.Errorf("attacker cannot attack defender")
	}

	// Store original health for world changes
	attackerOriginalHealth := attacker.AvailableHealth
	defenderOriginalHealth := defender.AvailableHealth

	// Calculate damage using rules engine
	attackerDamage := 0
	defenderDamage := 0

	var canAttack bool
	defenderDamage, canAttack, err = g.rulesEngine.CalculateCombatDamage(attacker.UnitType, defender.UnitType, g.rng)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate combat damage: %w", err)
	}
	if !canAttack {
		return nil, fmt.Errorf("unit type %d cannot attack unit type %d", attacker.UnitType, defender.UnitType)
	}

	// Check if defender can counter-attack
	if canCounter, err := g.rulesEngine.CanUnitAttackTarget(defender, attacker); err == nil && canCounter {
		var canCounterAttack bool
		attackerDamage, canCounterAttack, err = g.rulesEngine.CalculateCombatDamage(defender.UnitType, attacker.UnitType, g.rng)
		if err != nil || !canCounterAttack {
			// If counter-attack calculation fails or is not possible, no counter damage
			attackerDamage = 0
		}
	}

	// Apply damage
	defender.AvailableHealth -= int32(defenderDamage)
	if defender.AvailableHealth < 0 {
		defender.AvailableHealth = 0
	}

	attacker.AvailableHealth -= int32(attackerDamage)
	if attacker.AvailableHealth < 0 {
		attacker.AvailableHealth = 0
	}

	// Check if units were killed
	defenderKilled := defender.AvailableHealth <= 0
	attackerKilled := attacker.AvailableHealth <= 0

	// Add damage changes to world changes
	if defenderDamage > 0 {
		// Capture defender state before damage
		defenderPreviousUnit := copyUnit(defender)
		defenderPreviousUnit.AvailableHealth = defenderOriginalHealth

		// Capture defender state after damage
		defenderUpdatedUnit := copyUnit(defender)

		change := &v1.WorldChange{
			ChangeType: &v1.WorldChange_UnitDamaged{
				UnitDamaged: &v1.UnitDamagedChange{
					PreviousUnit: defenderPreviousUnit,
					UpdatedUnit:  defenderUpdatedUnit,
				},
			},
		}
		result.Changes = append(result.Changes, change)
	}

	if attackerDamage > 0 {
		// Capture attacker state before damage
		attackerPreviousUnit := copyUnit(attacker)
		attackerPreviousUnit.AvailableHealth = attackerOriginalHealth

		// Capture attacker state after damage
		attackerUpdatedUnit := copyUnit(attacker)

		change := &v1.WorldChange{
			ChangeType: &v1.WorldChange_UnitDamaged{
				UnitDamaged: &v1.UnitDamagedChange{
					PreviousUnit: attackerPreviousUnit,
					UpdatedUnit:  attackerUpdatedUnit,
				},
			},
		}
		result.Changes = append(result.Changes, change)
	}

	// Add kill changes if units were killed
	if defenderKilled {
		// Capture defender state before being killed (use original health before damage)
		defenderPreviousUnit := copyUnit(defender)
		defenderPreviousUnit.AvailableHealth = defenderOriginalHealth

		change := &v1.WorldChange{
			ChangeType: &v1.WorldChange_UnitKilled{
				UnitKilled: &v1.UnitKilledChange{
					PreviousUnit: defenderPreviousUnit,
				},
			},
		}
		result.Changes = append(result.Changes, change)
		g.World.RemoveUnit(defender)
	}

	if attackerKilled {
		// Capture attacker state before being killed (use original health before damage)
		attackerPreviousUnit := copyUnit(attacker)
		attackerPreviousUnit.AvailableHealth = attackerOriginalHealth

		change := &v1.WorldChange{
			ChangeType: &v1.WorldChange_UnitKilled{
				UnitKilled: &v1.UnitKilledChange{
					PreviousUnit: attackerPreviousUnit,
				},
			},
		}
		result.Changes = append(result.Changes, change)
		g.World.RemoveUnit(attacker)
	}

	// Update timestamp
	g.GameState.UpdatedAt = tspb.New(time.Now())

	return result, nil
}

// CanMoveUnit validates potential movement using cube coordinates
func (g *Game) CanMoveUnit(unit *v1.Unit, to AxialCoord) bool {
	if unit == nil {
		return false
	}

	// Check if it's the correct player's turn
	if unit.Player != g.CurrentPlayer {
		return false
	}

	// Check if move is valid
	unitCoord := UnitGetCoord(unit)
	return g.IsValidMove(unitCoord, to)
}

// CanAttackUnit validates potential attack
func (g *Game) CanAttackUnit(attacker, defender *v1.Unit) bool {
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
/* TODO -
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
*/

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

// GetMovementOptions returns movement options for unit at given coordinates with full validation
func (m *MoveProcessor) GetMovementOptions(game *Game, q, r int32) (*v1.AllPaths, error) {
	unit := game.World.UnitAt(AxialCoord{Q: int(q), R: int(r)})
	if unit == nil {
		return nil, fmt.Errorf("no unit found at position (%d, %d)", q, r)
	}
	if unit.Player != game.CurrentPlayer {
		return nil, fmt.Errorf("unit belongs to player %d, but it's player %d's turn", unit.Player, game.CurrentPlayer)
	}
	if unit.AvailableHealth <= 0 {
		return nil, fmt.Errorf("unit has no health remaining")
	}
	if unit.DistanceLeft <= 0 {
		return nil, fmt.Errorf("unit has no movement points remaining")
	}
	return game.rulesEngine.GetMovementOptions(game.World, unit, int(unit.DistanceLeft))
}

// GetAttackOptions returns attack options for unit at given coordinates with full validation
func (m *MoveProcessor) GetAttackOptions(game *Game, q, r int32) ([]AxialCoord, error) {
	unit := game.World.UnitAt(AxialCoord{Q: int(q), R: int(r)})
	if unit == nil {
		return nil, fmt.Errorf("no unit found at position (%d, %d)", q, r)
	}
	if unit.Player != game.CurrentPlayer {
		return nil, fmt.Errorf("unit belongs to player %d, but it's player %d's turn", unit.Player, game.CurrentPlayer)
	}
	if unit.AvailableHealth <= 0 {
		return nil, fmt.Errorf("unit has no health remaining")
	}
	return game.rulesEngine.GetAttackOptions(game.World, unit)
}

// CanSelectUnit validates if unit at given coordinates can be selected by current player
func (m *MoveProcessor) CanSelectUnit(game *Game, q, r int32) (bool, string) {
	unit := game.World.UnitAt(AxialCoord{Q: int(q), R: int(r)})
	if unit == nil {
		return false, fmt.Sprintf("no unit found at position (%d, %d)", q, r)
	}
	if unit.Player != game.CurrentPlayer {
		return false, fmt.Sprintf("unit belongs to player %d, but it's player %d's turn", unit.Player, game.CurrentPlayer)
	}
	if unit.AvailableHealth <= 0 {
		return false, "unit has no health remaining"
	}
	return true, ""
}

// CanMove validates potential movement using position coordinates
func (g *Game) CanMove(from, to Position) (bool, error) {
	unit := g.World.UnitAt(from)
	return g.CanMoveUnit(unit, to), nil
}

// GetUnitAttackOptions returns all positions a unit can attack using rules engine
func (g *Game) GetUnitAttackOptionsFrom(q, r int) ([]AxialCoord, error) {
	return g.GetUnitAttackOptions(g.World.UnitAt(AxialCoord{q, r}))
}
func (g *Game) GetUnitAttackOptions(unit *v1.Unit) ([]AxialCoord, error) {
	return g.rulesEngine.GetAttackOptions(g.World, unit)
}
