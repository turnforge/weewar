# WeeWar Architecture

## System Overview

WeeWar implements a clean separation between game logic (Go), state management (WASM), coordination (TurnEngine), and presentation (TypeScript/Phaser). The architecture emphasizes transaction safety, type safety, distributed validation, and maintainable separation of concerns.

## Core Design Principles

### 1. Transaction-Safe State Management

**Copy-on-Write Semantics**: The World system implements parent-child transaction layers where child layers create copies of parent objects before modification, preventing corruption of parent state.

```go
// Transaction layer creates copies to avoid parent mutation
if w.parent != nil {
    if _, existsInCurrentLayer := w.unitsByCoord[currentCoord]; !existsInCurrentLayer {
        // Unit comes from parent layer - make a copy
        unitToMove = &v1.Unit{
            Q: unit.Q, R: unit.R, Player: unit.Player,
            // ... copy all fields
        }
    }
}
```

**Rollback Safety**: ProcessMoves creates transaction snapshots, processes moves, then rolls back to original state for ordered change application:

```go
originalWorld := rtGame.World
rtGame.World = originalWorld.Push()  // Create transaction
// Process moves on transaction layer
rtGame.World = originalWorld        // Rollback for ordered apply
```

### 2. Service Layer Abstraction

**Transport Independence**: Same service implementations work across HTTP, gRPC, and WASM through interface abstraction:

```go
type GamesServiceImpl interface {
    v1.GamesServiceServer
    GetRuntimeGame(game *v1.Game, gameState *v1.GameState) (*Game, error)
}
```

**Base Implementation**: `BaseGamesServiceImpl` provides core logic that concrete implementations extend for specific transports.

### 3. Type-Safe WASM Integration

**Generated Client**: WASM client provides compile-time type checking with protobuf integration:

```typescript
// Type-safe API calls
const response = await this.wasmService.ProcessMoves({
    gameId: this.gameId,
    moves: [moveAction]
});
```

**Direct Property Access**: Simplified protobuf handling with direct field access (`change.unitMoved`) instead of complex oneof handling.

## Component Architecture

### Game Engine Layer (`lib/`)

**World**: Pure game state container implementing hex coordinate system with transaction support.
- Immutable parent-child relationships for transactions
- Efficient merged iteration across transaction layers
- Copy-on-write semantics for state safety

**Game**: Runtime game logic integrating World with Rules Engine.
- Current player/turn state management
- Move validation and processing coordination
- Rules engine integration for game mechanics

**Move Processor**: Validates and processes moves with full transaction support.
- Transaction-aware move validation
- Change result generation for state updates
- Error handling and rollback coordination

### Service Layer (`services/`)

**ProcessMoves Pipeline**:
1. Create transaction snapshot of game state
2. Process moves on transaction layer
3. Generate change results from transaction
4. Rollback to original state
5. Apply changes in ordered sequence
6. Update persistent state

**Change Application**: `ApplyChangeResults` ensures ordered, atomic application of move results to maintain state consistency.

### Frontend Layer (`web/`)

**GameState Controller**: Lightweight wrapper managing WASM service interactions.
- Move execution coordination
- State synchronization with server
- Event emission for UI updates

**GameViewer Renderer**: Phaser-based hex map rendering with unit display.
- Event-driven updates from GameState
- Hex coordinate conversion for display
- User interaction handling (clicks, selections)

## Transaction System Deep Dive

### Problem Solved

**Unit Duplication Bug**: Units appearing at both old and new positions after moves due to shared object references between transaction and parent layers.

### Solution Architecture

**Copy-on-Write in MoveUnit**:
```go
func (w *World) MoveUnit(unit *v1.Unit, newCoord AxialCoord) error {
    unitToMove := unit
    if w.parent != nil {
        // Check if unit comes from parent layer
        if _, existsInCurrentLayer := w.unitsByCoord[currentCoord]; !existsInCurrentLayer {
            // Make copy to avoid modifying parent objects
            unitToMove = &v1.Unit{/* copy all fields */}
        }
    }
    // Safe to modify copy
    UnitSetCoord(unitToMove, newCoord)
}
```

**Transaction Counter Optimization**:
```go
func (w *World) NumUnits() int32 {
    if w.parent != nil {
        return w.parent.NumUnits() + w.unitsAdded - w.unitsDeleted
    }
    return int32(len(w.unitsByCoord))
}
```

### Benefits

- **Data Integrity**: Parent layers remain immutable during transaction processing
- **Performance**: Efficient counting without expensive iteration
- **Rollback Safety**: Clean rollback to known good state
- **Test Coverage**: Comprehensive validation of transaction semantics

## Event System Architecture

### Clean Separation

**Game Logic → State → UI**: Unidirectional data flow prevents circular dependencies.

**Event Bus Pattern**: Loose coupling between components through centralized event coordination:
```typescript
// GameState processes moves and emits server changes
this.eventBus.emit('server-changes', { changes: worldChanges }, this, this);

// World applies changes and emits specific world events
this.eventBus.emit(WorldEventType.TILES_CHANGED, { changes: tileChanges }, this, this);

// PhaserWorldScene automatically syncs display
handleTilesChanged(data) { this.setTile(change.tile); }
```

### Automatic Rendering Synchronization

**Unified Event Flow**: Server changes automatically propagate to all rendering components:
1. **User Action** → **GameState.processMoves()** 
2. **server-changes** → **World.applyServerChanges()**
3. **TILES_CHANGED/UNITS_CHANGED** → **PhaserWorldScene.handleTilesChanged()/handleUnitsChanged()**
4. **Automatic Rendering Updates**

**Scene Architecture**: PhaserWorldScene base class handles world synchronization for all scenes:
- **PhaserEditorScene**: Inherits automatic world sync + editor interactions
- **PhaserViewerScene**: Inherits automatic world sync + viewer interactions
- No duplication of world event handling logic

### State Synchronization

**Server as Source of Truth**: All state changes validated server-side with client updates.

**Dual Subscription Pattern**: GameState subscribes to server-changes for metadata (currentPlayer, turnCounter) while also emitting them.

**World as Data Authority**: World component handles actual tile/unit data changes and re-emits as specific world events.

## Testing Architecture

### Comprehensive Coverage

**Unit Tests**: World operations with/without transactions
- Basic move operations
- Unit replacement scenarios  
- Transaction layer isolation
- Copy-on-write semantics

**Integration Tests**: End-to-end ProcessMoves pipeline
- Real WasmGamesService usage
- Transaction flow validation
- State consistency verification

**E2E Browser Tests**: Real GameViewerPage testing with Playwright
- Tests actual production `/games/{gameId}/view` endpoints
- Real WASM + EventBus + World integration testing
- Minimal surgical API mocking (only external calls)
- Command interface for accessibility and programmatic testing
- Both headless and head-full browser modes for debugging

**Transaction Flow Tests**: Exact simulation of ProcessMoves behavior
- Transaction creation and rollback
- Change application ordering
- Unit object sharing prevention

## Performance Considerations

### Efficient Operations

**Transaction Counters**: O(1) unit counting instead of O(n) iteration
**Copy-on-Write**: Only copies objects when actually modified
**Merged Iteration**: Lazy evaluation of parent/child object combination

### Memory Management

**Shallow Copies**: Unit objects use shallow copying for performance
**Transaction Cleanup**: Automatic cleanup when transactions complete
**Object Reuse**: Minimize allocation/deallocation in hot paths

## Multiplayer Coordination Architecture

### Local-First Validation Model

**Distributed Trust**: Each player's WASM validates moves locally before submission to coordinator.

**Server as Coordinator**: Server never runs game logic - only manages consensus and storage:
```go
// CoordinatorService is game-agnostic
type CoordinatorService struct {
    storage Storage      // Opaque blob storage
    callbacks Callbacks  // Game-specific hooks
}
```

### Coordination Flow

1. **Player Makes Move**: WASM runs ProcessMoves locally
2. **Submit Proposal**: Send moves + changes + new state to coordinator  
3. **Validator Assignment**: Coordinator assigns K validators from other players
4. **Distributed Validation**: Each validator's WASM verifies the proposal
5. **Consensus**: K-of-N approval commits the new state

### Key Design Decisions

**Game-Agnostic Coordinator**: TurnEngine coordinator knows nothing about WeeWar:
- All game data stored as opaque blobs
- State transitions verified by hash comparison
- Callbacks notify game service of proposal lifecycle

**Pull-Based Synchronization**: Simple REST polling (no websockets yet):
- Players refresh before making moves
- Validators poll for pending work
- Reduces complexity for initial implementation

**File-Based Storage**: JSON files for development simplicity:
- Human-readable for debugging
- Atomic operations via file locking
- Easy migration to database later

### Service Architecture

```
TurnEngine (Generic)
├── coordination/
│   ├── service.go         # Consensus logic
│   ├── storage.go         # Storage interface
│   └── file_storage.go    # File implementation
└── storage/
    └── file_storage.go    # Generic file storage

WeeWar (Game-Specific)  
├── services/
│   ├── coordinator_games.go  # Wraps FSGamesService
│   └── base.go               # ProcessMoves logic
└── protos/
    └── models.proto          # Includes ProposalTrackingInfo
```

## Future Architecture Considerations

### Scalability

**Multi-Player Support**: Coordination system supports K-of-N validation
**Cross-Game Validators**: Can use validators from other games (v2)
**State Partitioning**: World architecture supports regional game state management
**Caching Layer**: Service layer ready for redis/memcache integration

### Extension Points

**Rules Engine**: JSON-configurable game mechanics
**Transport Layer**: Easy addition of new client protocols
**Rendering Backend**: Phaser abstraction allows for alternative renderers
**Consensus Algorithms**: Pluggable validation strategies

This architecture provides a solid foundation for turn-based strategy games with distributed validation, excellent separation of concerns, comprehensive testing, and robust state management.