# WeeWar Architecture Overview

## Current Architecture (Post-Refactoring)

### Core Components

#### 1. Game Object (`lib/game.go`)
**Purpose**: Flow control and game logic management
- Manages game state transitions (turns, player actions)
- Handles game rules and validation
- Contains random number generation with deterministic seeding
- Manages event system for state changes
- **Pure flow control** - no rendering or UI concerns

```go
type Game struct {
    World *World `json:"world"` // Contains pure state
    
    // Game flow control
    CurrentPlayer int        `json:"currentPlayer"`
    TurnCounter   int        `json:"turnCounter"`
    Status        GameStatus `json:"status"`
    
    // Game systems
    Seed int64 `json:"seed"`
    rng *rand.Rand `json:"-"`
    eventManager *EventManager `json:"-"`
    assetProvider AssetProvider `json:"-"`
}
```

#### 2. World Object (`lib/world.go`)
**Purpose**: Pure game state container
- Contains all game entities (Map, Units by player)
- Implements WorldSubject for observer pattern
- **Pure state** - no game logic or rendering

```go
type World struct {
    Map           *Map      `json:"map"`
    UnitsByPlayer [][]*Unit `json:"units"` // Units organized by player
    
    WorldSubject `json:"-"` // Observer pattern
    PlayerCount int `json:"playerCount"`
}
```

#### 3. Map Object (`lib/map.go`)
**Purpose**: Hex grid management with cube coordinates
- Cube coordinate system (Q/R) with bounds (MinQ/MaxQ/MinR/MaxR)
- Origin management (OriginX/OriginY for coordinate system)
- Direct hex-to-pixel conversion using Red Blob Games formulas
- Efficient tile storage via `map[CubeCoord]*Tile`

```go
type Map struct {
    // Coordinate bounds
    MinQ, MaxQ, MinR, MaxR int
    
    // Origin for coordinate system
    OriginX, OriginY float64
    
    // Cube coordinate storage
    Tiles map[CubeCoord]*Tile
}
```

#### 4. WorldRenderer (`lib/world_renderer.go`)
**Purpose**: Platform-agnostic rendering of World state
- Works directly with World data (no Game object creation)
- Supports asset rendering with fallback to simple shapes
- Uses cube coordinates throughout
- Efficient rendering with direct Map.Tiles access

```go
type WorldRenderer interface {
    RenderWorld(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions)
    RenderTerrain(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions)
    RenderUnits(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions)
    RenderHighlights(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions)
}
```

#### 5. Observer Pattern (`lib/world_observer.go`)
**Purpose**: Reactive updates for state changes
- WorldSubject embedded in World for notifications
- WorldObserver interface for components requiring updates
- Event batching for performance optimization

### Key Design Principles

#### 1. Separation of Concerns
- **Game**: Flow control, rules, validation
- **World**: Pure state storage
- **WorldRenderer**: Rendering and visualization
- **CLI**: Translation layer (chess notation ↔ cube coordinates)

#### 2. Cube Coordinate System
- Primary coordinate system throughout codebase
- Proper hex mathematics using Red Blob Games formulas
- CLI preserves chess notation for user experience
- Efficient coordinate conversion and validation

#### 3. Clean Architecture
- No circular dependencies
- Interface-driven design
- Dependency injection for testability
- Clear data flow: CLI → Game → World → WorldRenderer

#### 4. Performance Optimizations
- Direct Map.Tiles access (no copying)
- Efficient cube coordinate storage
- Event batching for multiple state changes
- Asset caching and fallback rendering

### Data Flow

```
User Input (CLI) → Game (validation/logic) → World (state update) → WorldObserver (notifications) → WorldRenderer (display)
```

### Coordinate System

#### Cube Coordinates (Internal)
- Primary system: `CubeCoord{Q, R}` with `S = -Q - R`
- Bounds: `MinQ/MaxQ/MinR/MaxR`
- Origin: `OriginX/OriginY` for pixel conversion

#### Display Coordinates (User Interface)
- CLI: Chess notation (A1, B2, C3...)
- Internal conversion: Chess → Row/Col → Cube
- Preserved user experience with mathematical correctness

#### Pixel Coordinates (Rendering)
- Direct conversion: Cube → Pixel using `CenterXYForTile()`
- Pointy-topped hexagons with odd-r layout
- Formula: `x = originX + tileWidth * sqrt(3) * (q + r/2)`

### Testing Strategy

#### Unit Tests
- Coordinate conversion accuracy
- Game logic validation
- Observer pattern functionality
- Rendering output verification

#### Integration Tests
- CLI command compatibility
- Save/load functionality
- Cross-component communication
- Performance benchmarks

### Migration Benefits

#### 1. Mathematical Correctness
- Proper hex distance calculations
- Accurate coordinate conversions
- Support for arbitrary map regions

#### 2. Code Quality
- Single source of truth for coordinates
- Eliminated hardcoded values
- Clear separation of concerns
- Improved maintainability

#### 3. User Experience
- Preserved CLI chess notation
- Consistent coordinate behavior
- Support for negative coordinates
- Efficient rendering

#### 4. Extensibility
- Clean architecture for new features
- Platform-agnostic rendering
- Observer pattern for UI updates
- Interface-driven design

### Next Steps

1. **Complete Migration**: Update remaining components (canvas_buffer.go, editor.go)
2. **Testing**: Update test suite for cube coordinates
3. **Documentation**: API documentation and user guides
4. **Performance**: Benchmark and optimize where needed
5. **Features**: Build on clean architecture foundation

---

**Last Updated**: 2025-01-15  
**Architecture Version**: 2.0 (Post-Refactoring)  
**Status**: Production-ready foundation