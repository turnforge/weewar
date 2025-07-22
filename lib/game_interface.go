package weewar

// =============================================================================
// WeeWar Core Game Interface Definitions
// =============================================================================
// This file defines the core game interface contracts for the WeeWar game system.
// These interfaces focus purely on game mechanics: game state, map operations,
// unit management, and core events.

// =============================================================================
// Core Data Types
// =============================================================================

// GameStatus represents the current state of the game
type GameStatus int

const (
	GameStatusPlaying GameStatus = iota
	GameStatusPaused
	GameStatusEnded
)

func (gs GameStatus) String() string {
	switch gs {
	case GameStatusPlaying:
		return "playing"
	case GameStatusPaused:
		return "paused"
	case GameStatusEnded:
		return "ended"
	default:
		return "unknown"
	}
}

// Position represents a coordinate position (row, col)
type Position = AxialCoord

// CombatResult represents the outcome of a combat action
type CombatResult struct {
	AttackerDamage int  `json:"attackerDamage"` // Damage dealt to attacker
	DefenderDamage int  `json:"defenderDamage"` // Damage dealt to defender
	AttackerKilled bool `json:"attackerKilled"` // Whether attacker was destroyed
	DefenderKilled bool `json:"defenderKilled"` // Whether defender was destroyed
	AttackerHealth int  `json:"attackerHealth"` // Attacker's health after combat
	DefenderHealth int  `json:"defenderHealth"` // Defender's health after combat
}
