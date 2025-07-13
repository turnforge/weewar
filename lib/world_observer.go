package weewar

// =============================================================================
// Observer Pattern for World State Changes
// =============================================================================

// WorldObserver interface allows components to be notified when the world state changes.
// This enables reactive updates where views automatically re-render when game state is modified.
type WorldObserver interface {
	// OnWorldChanged is called whenever the world state is modified
	OnWorldChanged(world *World)
}

// WorldSubject manages a list of observers and notifies them of world changes.
// This is typically embedded in the Game class to provide observer functionality.
type WorldSubject struct {
	observers []WorldObserver
}

// AddObserver registers a new observer to receive world change notifications
func (ws *WorldSubject) AddObserver(observer WorldObserver) {
	ws.observers = append(ws.observers, observer)
}

// RemoveObserver unregisters an observer from receiving notifications
func (ws *WorldSubject) RemoveObserver(observer WorldObserver) {
	for i, obs := range ws.observers {
		if obs == observer {
			// Remove observer from slice
			ws.observers = append(ws.observers[:i], ws.observers[i+1:]...)
			return
		}
	}
}

// NotifyWorldChanged sends a world change notification to all registered observers
func (ws *WorldSubject) NotifyWorldChanged(world *World) {
	for _, observer := range ws.observers {
		observer.OnWorldChanged(world)
	}
}

// GetObserverCount returns the number of registered observers (useful for testing)
func (ws *WorldSubject) GetObserverCount() int {
	return len(ws.observers)
}

// ClearObservers removes all registered observers
func (ws *WorldSubject) ClearObservers() {
	ws.observers = make([]WorldObserver, 0)
}

// =============================================================================
// Future Event Types (For Extension)
// =============================================================================

// These interfaces are defined for future expansion of the observer pattern
// to support more fine-grained event notifications.

// UnitObserver provides notifications for unit-specific events
type UnitObserver interface {
	OnUnitMoved(unit *Unit, fromRow, fromCol, toRow, toCol int)
	OnUnitDestroyed(unit *Unit)
	OnUnitSpawned(unit *Unit)
	OnUnitDamaged(unit *Unit, damage int)
}

// TerrainObserver provides notifications for terrain changes
type TerrainObserver interface {
	OnTerrainChanged(row, col int, oldType, newType int)
	OnTileAdded(row, col int, terrainType int)
	OnTileRemoved(row, col int)
}

// GameObserver provides notifications for game-level events
type GameObserver interface {
	OnPlayerTurnStarted(playerID int)
	OnPlayerTurnEnded(playerID int)
	OnGameStarted()
	OnGameEnded(winnerID int)
}

// DetailedWorldObserver combines all observer types for components that need comprehensive notifications
type DetailedWorldObserver interface {
	WorldObserver
	UnitObserver
	TerrainObserver
	GameObserver
}

// =============================================================================
// Observer Utility Functions
// =============================================================================

// NewWorldSubject creates a new WorldSubject with an empty observer list
func NewWorldSubject() *WorldSubject {
	return &WorldSubject{
		observers: make([]WorldObserver, 0),
	}
}

// ObserverGroup allows managing multiple observers as a single unit
type ObserverGroup struct {
	observers []WorldObserver
}

// NewObserverGroup creates a new group of observers
func NewObserverGroup() *ObserverGroup {
	return &ObserverGroup{
		observers: make([]WorldObserver, 0),
	}
}

// Add adds an observer to the group
func (og *ObserverGroup) Add(observer WorldObserver) {
	og.observers = append(og.observers, observer)
}

// OnWorldChanged notifies all observers in the group
func (og *ObserverGroup) OnWorldChanged(world *World) {
	for _, observer := range og.observers {
		observer.OnWorldChanged(world)
	}
}

// =============================================================================
// Event Batching (For Performance)
// =============================================================================

// EventBatch allows batching multiple world changes into a single notification.
// This is useful for operations that make multiple modifications to the world state.
type EventBatch struct {
	subject *WorldSubject
	world   *World
	active  bool
}

// NewEventBatch creates a new event batch for the given subject and world
func NewEventBatch(subject *WorldSubject, world *World) *EventBatch {
	return &EventBatch{
		subject: subject,
		world:   world,
		active:  false,
	}
}

// Begin starts the event batch, suppressing individual notifications
func (eb *EventBatch) Begin() {
	eb.active = true
}

// End completes the event batch and sends a single notification to all observers
func (eb *EventBatch) End() {
	if eb.active {
		eb.active = false
		eb.subject.NotifyWorldChanged(eb.world)
	}
}

// IsActive returns whether the batch is currently active
func (eb *EventBatch) IsActive() bool {
	return eb.active
}