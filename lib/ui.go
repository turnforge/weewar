package weewar

import "fmt"

// =============================================================================
// UI Helper Methods for WASM/Web Interface
// =============================================================================
// This file contains wrapper methods that combine core game functionality
// into convenient forms for UI consumption, particularly for the WASM bridge.
// These methods reuse existing types (TileOption, CombatResult, AxialCoord)
// and delegate to core Game methods for all logic and validation.

// SelectUnit returns unit at position with movement and attack options for UI
// Combines existing GetUnitAt, GetUnitMovementOptions, and GetUnitAttackOptions
// Returns data needed for UI highlighting and interaction
func (g *Game) SelectUnit(coord AxialCoord) (unit *Unit, movable []TileOption, attackable []AxialCoord, err error) {
	// Get unit at position using existing method
	unit = g.World.UnitAt(coord)
	if unit == nil {
		return nil, nil, nil, fmt.Errorf("no unit at position %v", coord)
	}

	// Check if it's the current player's unit
	if unit.PlayerID != g.CurrentPlayer {
		return nil, nil, nil, fmt.Errorf("unit belongs to player %d, current player is %d", unit.PlayerID, g.CurrentPlayer)
	}

	// Get movement options using existing method from moves.go
	movable, err = g.GetUnitMovementOptions(unit)
	if err != nil {
		return unit, nil, nil, fmt.Errorf("failed to get movement options: %w", err)
	}

	// Get attack options using existing method from moves.go
	attackable, err = g.GetUnitAttackOptions(unit)
	if err != nil {
		return unit, movable, nil, fmt.Errorf("failed to get attack options: %w", err)
	}

	return unit, movable, attackable, nil
}

// GetGameStateForUI returns complete game state for web UI consumption
// Uses existing Game fields and methods - all already JSON-tagged
// Provides everything needed for UI state management and display
func (g *Game) GetGameStateForUI() map[string]interface{} {
	return map[string]interface{}{
		"currentPlayer": g.CurrentPlayer,      // Current player's turn
		"turnCounter":   g.TurnCounter,        // Turn number
		"status":        g.Status,             // GameStatus (playing/ended/paused)
		"allUnits":      g.World.UnitsByCoord, // All units on map
		"players":       g.Players,            // Player information
		"teams":         g.Teams,              // Team information
		"mapSize": map[string]int{ // Map dimensions
			"rows": g.World.Map.NumRows(),
			"cols": g.World.Map.NumCols(),
		},
		"winner":    g.winner,    // Winner if game ended
		"hasWinner": g.hasWinner, // Whether game has ended
	}
}
