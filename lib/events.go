package weewar

import (
	"sync"
)

// =============================================================================
// WeeWar Event System
// =============================================================================
// This file defines the event system for the WeeWar game, providing
// observer pattern functionality for game events and state changes.

// =============================================================================
// Event System Interface
// =============================================================================

// EventInterface provides event subscription and emission
type EventInterface interface {
	// Event Subscription (Observer Pattern)
	// OnUnitMoved subscribes to unit movement events
	// Called by: Animation systems, Sound effects, Statistics tracking
	OnUnitMoved(callback func(unit *Unit, from, to Position))
	
	// OnUnitAttacked subscribes to combat events
	// Called by: Animation systems, Sound effects, Damage display
	OnUnitAttacked(callback func(attacker, defender *Unit, result *CombatResult))
	
	// OnTurnChanged subscribes to turn change events
	// Called by: UI updates, Player notifications, AI activation
	OnTurnChanged(callback func(newPlayer int, turnNumber int))
	
	// OnGameEnded subscribes to game completion events
	// Called by: Victory screens, Statistics saving, Game result logging
	OnGameEnded(callback func(winner int))
	
	// OnUnitCreated subscribes to unit creation events
	// Called by: Statistics tracking, Achievement systems
	OnUnitCreated(callback func(unit *Unit))
	
	// OnUnitDestroyed subscribes to unit destruction events
	// Called by: Statistics tracking, Achievement systems
	OnUnitDestroyed(callback func(unit *Unit))
	
	// OnGameStateChanged subscribes to general game state changes
	// Called by: Save systems, State synchronization
	OnGameStateChanged(callback func(changeType string, data interface{}))
	
	// Event Emission (Internal Use)
	// EmitUnitMoved triggers unit movement events
	// Called by: Internal game logic after successful movement
	EmitUnitMoved(unit *Unit, from, to Position)
	
	// EmitUnitAttacked triggers combat events
	// Called by: Internal game logic after combat resolution
	EmitUnitAttacked(attacker, defender *Unit, result *CombatResult)
	
	// EmitTurnChanged triggers turn change events
	// Called by: Internal game logic after turn advancement
	EmitTurnChanged(newPlayer int, turnNumber int)
	
	// EmitGameEnded triggers game completion events
	// Called by: Internal game logic when victory conditions met
	EmitGameEnded(winner int)
	
	// EmitUnitCreated triggers unit creation events
	// Called by: Internal game logic after unit creation
	EmitUnitCreated(unit *Unit)
	
	// EmitUnitDestroyed triggers unit destruction events
	// Called by: Internal game logic after unit destruction
	EmitUnitDestroyed(unit *Unit)
	
	// EmitGameStateChanged triggers general game state change events
	// Called by: Internal game logic after state changes
	EmitGameStateChanged(changeType string, data interface{})
	
	// Event Management
	// ClearAllCallbacks removes all event subscriptions
	// Called by: Game cleanup, Testing, Reset
	ClearAllCallbacks()
	
	// GetCallbackCounts returns number of callbacks for each event type
	// Called by: Debugging, Statistics, Memory management
	GetCallbackCounts() map[string]int
}

// =============================================================================
// Event Manager Implementation
// =============================================================================

// EventManager manages event subscriptions and emissions
type EventManager struct {
	// Event callbacks
	unitMovedCallbacks       []func(unit *Unit, from, to Position)
	unitAttackedCallbacks    []func(attacker, defender *Unit, result *CombatResult)
	turnChangedCallbacks     []func(newPlayer int, turnNumber int)
	gameEndedCallbacks       []func(winner int)
	unitCreatedCallbacks     []func(unit *Unit)
	unitDestroyedCallbacks   []func(unit *Unit)
	gameStateChangedCallbacks []func(changeType string, data interface{})
	
	// Mutex for thread-safe operations
	mutex sync.RWMutex
}

// NewEventManager creates a new event manager
func NewEventManager() *EventManager {
	return &EventManager{
		unitMovedCallbacks:       make([]func(unit *Unit, from, to Position), 0),
		unitAttackedCallbacks:    make([]func(attacker, defender *Unit, result *CombatResult), 0),
		turnChangedCallbacks:     make([]func(newPlayer int, turnNumber int), 0),
		gameEndedCallbacks:       make([]func(winner int), 0),
		unitCreatedCallbacks:     make([]func(unit *Unit), 0),
		unitDestroyedCallbacks:   make([]func(unit *Unit), 0),
		gameStateChangedCallbacks: make([]func(changeType string, data interface{}), 0),
	}
}

// =============================================================================
// Event Subscription Methods
// =============================================================================

// OnUnitMoved subscribes to unit movement events
func (em *EventManager) OnUnitMoved(callback func(unit *Unit, from, to Position)) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	
	em.unitMovedCallbacks = append(em.unitMovedCallbacks, callback)
}

// OnUnitAttacked subscribes to combat events
func (em *EventManager) OnUnitAttacked(callback func(attacker, defender *Unit, result *CombatResult)) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	
	em.unitAttackedCallbacks = append(em.unitAttackedCallbacks, callback)
}

// OnTurnChanged subscribes to turn change events
func (em *EventManager) OnTurnChanged(callback func(newPlayer int, turnNumber int)) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	
	em.turnChangedCallbacks = append(em.turnChangedCallbacks, callback)
}

// OnGameEnded subscribes to game completion events
func (em *EventManager) OnGameEnded(callback func(winner int)) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	
	em.gameEndedCallbacks = append(em.gameEndedCallbacks, callback)
}

// OnUnitCreated subscribes to unit creation events
func (em *EventManager) OnUnitCreated(callback func(unit *Unit)) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	
	em.unitCreatedCallbacks = append(em.unitCreatedCallbacks, callback)
}

// OnUnitDestroyed subscribes to unit destruction events
func (em *EventManager) OnUnitDestroyed(callback func(unit *Unit)) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	
	em.unitDestroyedCallbacks = append(em.unitDestroyedCallbacks, callback)
}

// OnGameStateChanged subscribes to general game state changes
func (em *EventManager) OnGameStateChanged(callback func(changeType string, data interface{})) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	
	em.gameStateChangedCallbacks = append(em.gameStateChangedCallbacks, callback)
}

// =============================================================================
// Event Emission Methods
// =============================================================================

// EmitUnitMoved triggers unit movement events
func (em *EventManager) EmitUnitMoved(unit *Unit, from, to Position) {
	em.mutex.RLock()
	callbacks := make([]func(unit *Unit, from, to Position), len(em.unitMovedCallbacks))
	copy(callbacks, em.unitMovedCallbacks)
	em.mutex.RUnlock()
	
	// Call all callbacks without holding the lock
	for _, callback := range callbacks {
		callback(unit, from, to)
	}
}

// EmitUnitAttacked triggers combat events
func (em *EventManager) EmitUnitAttacked(attacker, defender *Unit, result *CombatResult) {
	em.mutex.RLock()
	callbacks := make([]func(attacker, defender *Unit, result *CombatResult), len(em.unitAttackedCallbacks))
	copy(callbacks, em.unitAttackedCallbacks)
	em.mutex.RUnlock()
	
	// Call all callbacks without holding the lock
	for _, callback := range callbacks {
		callback(attacker, defender, result)
	}
}

// EmitTurnChanged triggers turn change events
func (em *EventManager) EmitTurnChanged(newPlayer int, turnNumber int) {
	em.mutex.RLock()
	callbacks := make([]func(newPlayer int, turnNumber int), len(em.turnChangedCallbacks))
	copy(callbacks, em.turnChangedCallbacks)
	em.mutex.RUnlock()
	
	// Call all callbacks without holding the lock
	for _, callback := range callbacks {
		callback(newPlayer, turnNumber)
	}
}

// EmitGameEnded triggers game completion events
func (em *EventManager) EmitGameEnded(winner int) {
	em.mutex.RLock()
	callbacks := make([]func(winner int), len(em.gameEndedCallbacks))
	copy(callbacks, em.gameEndedCallbacks)
	em.mutex.RUnlock()
	
	// Call all callbacks without holding the lock
	for _, callback := range callbacks {
		callback(winner)
	}
}

// EmitUnitCreated triggers unit creation events
func (em *EventManager) EmitUnitCreated(unit *Unit) {
	em.mutex.RLock()
	callbacks := make([]func(unit *Unit), len(em.unitCreatedCallbacks))
	copy(callbacks, em.unitCreatedCallbacks)
	em.mutex.RUnlock()
	
	// Call all callbacks without holding the lock
	for _, callback := range callbacks {
		callback(unit)
	}
}

// EmitUnitDestroyed triggers unit destruction events
func (em *EventManager) EmitUnitDestroyed(unit *Unit) {
	em.mutex.RLock()
	callbacks := make([]func(unit *Unit), len(em.unitDestroyedCallbacks))
	copy(callbacks, em.unitDestroyedCallbacks)
	em.mutex.RUnlock()
	
	// Call all callbacks without holding the lock
	for _, callback := range callbacks {
		callback(unit)
	}
}

// EmitGameStateChanged triggers general game state change events
func (em *EventManager) EmitGameStateChanged(changeType string, data interface{}) {
	em.mutex.RLock()
	callbacks := make([]func(changeType string, data interface{}), len(em.gameStateChangedCallbacks))
	copy(callbacks, em.gameStateChangedCallbacks)
	em.mutex.RUnlock()
	
	// Call all callbacks without holding the lock
	for _, callback := range callbacks {
		callback(changeType, data)
	}
}

// =============================================================================
// Event Management Methods
// =============================================================================

// ClearAllCallbacks removes all event subscriptions
func (em *EventManager) ClearAllCallbacks() {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	
	em.unitMovedCallbacks = make([]func(unit *Unit, from, to Position), 0)
	em.unitAttackedCallbacks = make([]func(attacker, defender *Unit, result *CombatResult), 0)
	em.turnChangedCallbacks = make([]func(newPlayer int, turnNumber int), 0)
	em.gameEndedCallbacks = make([]func(winner int), 0)
	em.unitCreatedCallbacks = make([]func(unit *Unit), 0)
	em.unitDestroyedCallbacks = make([]func(unit *Unit), 0)
	em.gameStateChangedCallbacks = make([]func(changeType string, data interface{}), 0)
}

// GetCallbackCounts returns the number of callbacks for each event type
func (em *EventManager) GetCallbackCounts() map[string]int {
	em.mutex.RLock()
	defer em.mutex.RUnlock()
	
	return map[string]int{
		"unitMoved":         len(em.unitMovedCallbacks),
		"unitAttacked":      len(em.unitAttackedCallbacks),
		"turnChanged":       len(em.turnChangedCallbacks),
		"gameEnded":         len(em.gameEndedCallbacks),
		"unitCreated":       len(em.unitCreatedCallbacks),
		"unitDestroyed":     len(em.unitDestroyedCallbacks),
		"gameStateChanged":  len(em.gameStateChangedCallbacks),
	}
}

// =============================================================================
// Event Type Constants
// =============================================================================

// Game state change types
const (
	GameStateChangeGameStarted = "game_started"
	GameStateChangeGameEnded   = "game_ended"
	GameStateChangeGamePaused  = "game_paused"
	GameStateChangeGameResumed = "game_resumed"
	GameStateChangeMapLoaded   = "map_loaded"
	GameStateChangeUnitMoved   = "unit_moved"
	GameStateChangeUnitAttacked = "unit_attacked"
	GameStateChangeUnitCreated = "unit_created"
	GameStateChangeUnitDestroyed = "unit_destroyed"
	GameStateChangeTurnChanged = "turn_changed"
	GameStateChangePlayerJoined = "player_joined"
	GameStateChangePlayerLeft  = "player_left"
)