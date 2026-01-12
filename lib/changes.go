package lib

import (
	"fmt"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// ApplyChanges applies WorldChanges from processed moves to the game state.
// This handles the transaction rollback pattern: after ProcessMoves operates on
// a transaction snapshot, ApplyChanges applies the changes to the original world.
func (g *Game) ApplyChanges(moves []*v1.GameMove) error {
	// TRANSACTIONAL FIX: Temporary rollback to original world for ordered application
	if parent := g.World.Pop(); parent != nil {
		g.World = parent // Switch back to original world
	}

	// Apply each change to runtime game (now the original, not the transaction snapshot)
	for _, moveResult := range moves {
		for _, change := range moveResult.Changes {
			err := g.applyWorldChange(change)
			if err != nil {
				return fmt.Errorf("failed to apply world change: %w", err)
			}
		}
	}

	return nil
}

// applyWorldChange applies a single WorldChange to the game state
func (g *Game) applyWorldChange(change *v1.WorldChange) error {
	switch changeType := change.ChangeType.(type) {
	case *v1.WorldChange_UnitMoved:
		return g.applyUnitMoved(changeType.UnitMoved)
	case *v1.WorldChange_UnitDamaged:
		return g.applyUnitDamaged(changeType.UnitDamaged)
	case *v1.WorldChange_UnitKilled:
		return g.applyUnitKilled(changeType.UnitKilled)
	case *v1.WorldChange_PlayerChanged:
		return g.applyPlayerChanged(changeType.PlayerChanged)
	case *v1.WorldChange_UnitBuilt:
		return g.applyUnitBuilt(changeType.UnitBuilt)
	case *v1.WorldChange_CoinsChanged:
		return g.applyCoinsChanged(changeType.CoinsChanged)
	default:
		return fmt.Errorf("unknown world change type")
	}
}

// applyUnitMoved moves a unit in the runtime game
func (g *Game) applyUnitMoved(change *v1.UnitMovedChange) error {
	if change.PreviousUnit == nil || change.UpdatedUnit == nil {
		return fmt.Errorf("missing unit data in UnitMovedChange")
	}

	fromCoord := AxialCoord{Q: int(change.PreviousUnit.Q), R: int(change.PreviousUnit.R)}
	toCoord := AxialCoord{Q: int(change.UpdatedUnit.Q), R: int(change.UpdatedUnit.R)}

	// Move unit in runtime game
	unit := g.World.UnitAt(fromCoord)
	if unit == nil {
		return fmt.Errorf("unit not found at %v", fromCoord)
	}

	// Update unit with complete state from the change
	unit.AvailableHealth = change.UpdatedUnit.AvailableHealth
	unit.DistanceLeft = change.UpdatedUnit.DistanceLeft
	unit.LastActedTurn = change.UpdatedUnit.LastActedTurn
	unit.LastToppedupTurn = change.UpdatedUnit.LastToppedupTurn
	unit.ProgressionStep = change.UpdatedUnit.ProgressionStep
	unit.ChosenAlternative = change.UpdatedUnit.ChosenAlternative

	// Remove from old position and add to new position
	return g.World.MoveUnit(unit, toCoord)
}

// applyUnitDamaged updates unit health in the runtime game
func (g *Game) applyUnitDamaged(change *v1.UnitDamagedChange) error {
	if change.UpdatedUnit == nil {
		return fmt.Errorf("missing updated unit data in UnitDamagedChange")
	}

	coord := AxialCoord{Q: int(change.UpdatedUnit.Q), R: int(change.UpdatedUnit.R)}

	unit := g.World.UnitAt(coord)
	if unit == nil {
		return fmt.Errorf("unit not found at %v", coord)
	}

	// Update unit with complete state from the change
	unit.AvailableHealth = change.UpdatedUnit.AvailableHealth
	unit.DistanceLeft = change.UpdatedUnit.DistanceLeft
	unit.LastActedTurn = change.UpdatedUnit.LastActedTurn
	unit.LastToppedupTurn = change.UpdatedUnit.LastToppedupTurn
	unit.ProgressionStep = change.UpdatedUnit.ProgressionStep
	unit.ChosenAlternative = change.UpdatedUnit.ChosenAlternative
	return nil
}

// applyUnitKilled removes a unit from the runtime game
func (g *Game) applyUnitKilled(change *v1.UnitKilledChange) error {
	if change.PreviousUnit == nil {
		return fmt.Errorf("missing previous unit data in UnitKilledChange")
	}

	coord := AxialCoord{Q: int(change.PreviousUnit.Q), R: int(change.PreviousUnit.R)}
	unit := g.World.UnitAt(coord)

	err := g.World.RemoveUnit(unit)
	if err != nil {
		return fmt.Errorf("unit not found at %v", coord)
	}
	return nil
}

// applyPlayerChanged updates game state for turn/player changes
func (g *Game) applyPlayerChanged(change *v1.PlayerChangedChange) error {
	g.CurrentPlayer = change.NewPlayer
	g.TurnCounter = change.NewTurn

	// Also update the protobuf GameState
	g.GameState.CurrentPlayer = change.NewPlayer
	g.GameState.TurnCounter = change.NewTurn

	// Apply reset units (for remote updates where units need topped-up values)
	// The server has already calculated the new unit states; we apply them here
	for _, resetUnit := range change.ResetUnits {
		coord := AxialCoord{Q: int(resetUnit.Q), R: int(resetUnit.R)}
		unit := g.World.UnitAt(coord)
		if unit != nil {
			// Update unit with topped-up values from the change
			unit.DistanceLeft = resetUnit.DistanceLeft
			unit.AvailableHealth = resetUnit.AvailableHealth
			unit.LastToppedupTurn = resetUnit.LastToppedupTurn
			unit.LastActedTurn = resetUnit.LastActedTurn
		}
	}

	return nil
}

// applyUnitBuilt adds a newly built unit to the runtime game
func (g *Game) applyUnitBuilt(change *v1.UnitBuiltChange) error {
	if change.Unit == nil {
		return fmt.Errorf("missing unit data in UnitBuiltChange")
	}

	// Add the new unit to the runtime game
	g.World.AddUnit(change.Unit)

	// Update tile's last acted turn
	coord := AxialCoord{Q: int(change.TileQ), R: int(change.TileR)}
	tile := g.World.TileAt(coord)
	if tile != nil {
		tile.LastActedTurn = g.TurnCounter
	}

	return nil
}

// applyCoinsChanged updates a player's coin balance in the runtime game
func (g *Game) applyCoinsChanged(change *v1.CoinsChangedChange) error {
	// Update player's coins in GameState.PlayerStates
	if g.GameState.PlayerStates == nil {
		g.GameState.PlayerStates = make(map[int32]*v1.PlayerState)
	}
	playerState := g.GameState.PlayerStates[change.PlayerId]
	if playerState == nil {
		playerState = &v1.PlayerState{}
		g.GameState.PlayerStates[change.PlayerId] = playerState
	}
	playerState.Coins = change.NewCoins
	return nil
}
