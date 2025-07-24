package weewar

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// GameLog Core Data Structures
// =============================================================================

// GameAction represents a single action performed by a player
type GameAction struct {
	Type   string                 `json:"type"`   // "move", "attack", "endTurn", "gameStart"
	Params map[string]interface{} `json:"params"` // Action-specific parameters
}

// WorldChange represents a change to the world state resulting from an action
type WorldChange struct {
	Type       string      `json:"type"`       // "unitMoved", "unitKilled", "unitCreated", "playerChanged"
	EntityType string      `json:"entityType"` // "unit", "city", "player", "game"
	EntityID   string      `json:"entityId"`   // Identifier for the changed entity
	FromState  interface{} `json:"fromState"`  // Previous state (for undo)
	ToState    interface{} `json:"toState"`    // New state
	Metadata   interface{} `json:"metadata"`   // Additional context
}

// GameLogEntry represents a single entry in the game log
type GameLogEntry struct {
	ID        string        `json:"id"`
	Timestamp time.Time     `json:"timestamp"`
	Player    int           `json:"player"`
	Action    GameAction    `json:"action"`
	Changes   []WorldChange `json:"changes"`
}

// SessionMetadata contains metadata about a game session
type SessionMetadata struct {
	MapName     string                 `json:"mapName"`
	PlayerCount int                    `json:"playerCount"`
	MaxTurns    int                    `json:"maxTurns"`
	GameConfig  map[string]interface{} `json:"gameConfig"`
}

// GameSession represents a complete game session with all actions
type GameSession struct {
	SessionID     string          `json:"sessionId"`
	StartedAt     time.Time       `json:"startedAt"`
	LastUpdated   time.Time       `json:"lastUpdated"`
	WorldID       string          `json:"worldId"`
	StartingWorld []byte          `json:"startingWorld"` // Serialized initial world state
	Entries       []GameLogEntry  `json:"entries"`
	Status        string          `json:"status"` // "active", "paused", "completed", "abandoned"
	Metadata      SessionMetadata `json:"metadata"`
}

// =============================================================================
// SaveHandler Interface for Pluggable Storage
// =============================================================================

// SaveHandler defines the interface for saving and loading game sessions
type SaveHandler interface {
	// Save stores a game session
	Save(sessionData []byte) error
	
	// Load retrieves a game session by ID
	Load(sessionID string) ([]byte, error)
	
	// List returns all available session IDs
	List() ([]string, error)
	
	// Delete removes a session
	Delete(sessionID string) error
}

// =============================================================================
// GameLog Implementation
// =============================================================================

// GameLog manages the recording and persistence of game actions
type GameLog struct {
	session     *GameSession
	saveHandler SaveHandler
	autoSave    bool
	entries     []GameLogEntry // In-memory buffer for pending entries
}

// NewGameLog creates a new GameLog instance
func NewGameLog(saveHandler SaveHandler, autoSave bool) *GameLog {
	return &GameLog{
		saveHandler: saveHandler,
		autoSave:    autoSave,
		entries:     make([]GameLogEntry, 0),
	}
}

// StartNewSession begins a new game session
func (gl *GameLog) StartNewSession(worldID string, startingWorldData []byte, metadata SessionMetadata) error {
	sessionID := uuid.New().String()
	
	gl.session = &GameSession{
		SessionID:     sessionID,
		StartedAt:     time.Now(),
		LastUpdated:   time.Now(),
		WorldID:       worldID,
		StartingWorld: startingWorldData,
		Entries:       make([]GameLogEntry, 0),
		Status:        "active",
		Metadata:      metadata,
	}
	
	// Record the game start action
	gameStartAction := GameAction{
		Type: "gameStart",
		Params: map[string]interface{}{
			"worldId":    worldID,
			"mapName":    metadata.MapName,
			"playerCount": metadata.PlayerCount,
		},
	}
	
	return gl.RecordAction(0, gameStartAction, []WorldChange{})
}

// LoadSession loads an existing game session
func (gl *GameLog) LoadSession(sessionID string) error {
	if gl.saveHandler == nil {
		return fmt.Errorf("no save handler configured")
	}
	
	sessionData, err := gl.saveHandler.Load(sessionID)
	if err != nil {
		return fmt.Errorf("failed to load session %s: %w", sessionID, err)
	}
	
	var session GameSession
	if err := json.Unmarshal(sessionData, &session); err != nil {
		return fmt.Errorf("failed to unmarshal session data: %w", err)
	}
	
	gl.session = &session
	gl.entries = make([]GameLogEntry, 0) // Clear pending entries
	
	return nil
}

// RecordAction adds a new action to the game log
func (gl *GameLog) RecordAction(player int, action GameAction, changes []WorldChange) error {
	if gl.session == nil {
		return fmt.Errorf("no active session - call StartNewSession first")
	}
	
	entry := GameLogEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Player:    player,
		Action:    action,
		Changes:   changes,
	}
	
	// Add to both session and pending entries
	gl.session.Entries = append(gl.session.Entries, entry)
	gl.entries = append(gl.entries, entry)
	gl.session.LastUpdated = time.Now()
	
	// Auto-save if enabled
	if gl.autoSave && gl.saveHandler != nil {
		return gl.Save()
	}
	
	return nil
}

// Save persists the current session using the SaveHandler
func (gl *GameLog) Save() error {
	if gl.session == nil {
		return fmt.Errorf("no active session to save")
	}
	
	if gl.saveHandler == nil {
		return fmt.Errorf("no save handler configured")
	}
	
	sessionData, err := json.Marshal(gl.session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	
	if err := gl.saveHandler.Save(sessionData); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	
	// Clear pending entries after successful save
	gl.entries = make([]GameLogEntry, 0)
	
	return nil
}

// GetCurrentSession returns the current session (for debugging/export)
func (gl *GameLog) GetCurrentSession() *GameSession {
	if gl.session == nil {
		return nil
	}
	
	// Return a copy to prevent external modification
	sessionCopy := *gl.session
	sessionCopy.Entries = make([]GameLogEntry, len(gl.session.Entries))
	copy(sessionCopy.Entries, gl.session.Entries)
	
	return &sessionCopy
}

// GetSessionID returns the current session ID
func (gl *GameLog) GetSessionID() string {
	if gl.session == nil {
		return ""
	}
	return gl.session.SessionID
}

// SetStatus updates the session status
func (gl *GameLog) SetStatus(status string) error {
	if gl.session == nil {
		return fmt.Errorf("no active session")
	}
	
	gl.session.Status = status
	gl.session.LastUpdated = time.Now()
	
	// Auto-save status change
	if gl.autoSave && gl.saveHandler != nil {
		return gl.Save()
	}
	
	return nil
}

// GetEntryCount returns the number of entries in the current session
func (gl *GameLog) GetEntryCount() int {
	if gl.session == nil {
		return 0
	}
	return len(gl.session.Entries)
}

// =============================================================================
// Helper Functions for Creating Common Actions and Changes
// =============================================================================

// CreateMoveAction creates a standardized move action
func CreateMoveAction(fromQ, fromR, toQ, toR int) GameAction {
	return GameAction{
		Type: "move",
		Params: map[string]interface{}{
			"fromQ": fromQ,
			"fromR": fromR,
			"toQ":   toQ,
			"toR":   toR,
		},
	}
}

// CreateAttackAction creates a standardized attack action
func CreateAttackAction(attackerQ, attackerR, defenderQ, defenderR int) GameAction {
	return GameAction{
		Type: "attack",
		Params: map[string]interface{}{
			"attackerQ":  attackerQ,
			"attackerR":  attackerR,
			"defenderQ":  defenderQ,
			"defenderR":  defenderR,
		},
	}
}

// CreateEndTurnAction creates a standardized end turn action
func CreateEndTurnAction() GameAction {
	return GameAction{
		Type:   "endTurn",
		Params: map[string]interface{}{},
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