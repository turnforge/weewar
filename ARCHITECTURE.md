# WeeWar Game Architecture

## Overview

The WeeWar game is built on top of the TurnEngine framework, demonstrating a clear separation between the reusable game engine components and game-specific implementations. This architecture follows the principle of 80% shared engine code and 20% game-specific code.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     TurnEngine (Framework)                  │
├─────────────────────────────────────────────────────────────┤
│ • Entity Component System (ECS)                             │
│ • Game State Management                                     │
│ • Turn-based Game Loop                                      │
│ • Command Processing Pipeline                               │
│ • Abstract Board Interface                                  │
│ • Abstract Position Interface                               │
│ • Player Management                                         │
│ • World/System Management                                   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    WeeWar Implementation                    │
├─────────────────────────────────────────────────────────────┤
│ • HexBoard & HexPosition                                    │
│ • WeeWar-specific Components                                │
│ • Combat & Movement Systems                                 │
│ • Real Game Data Integration                                │
│ • Map System                                                │
│ • Command Validators & Processors                           │
└─────────────────────────────────────────────────────────────┘
```

## Core Framework Components (TurnEngine)

### 1. Entity Component System
**Location**: `internal/turnengine/`
- **Purpose**: Provides the foundation for all game objects
- **Key Files**: `entity.go`, `component.go`
- **Reusability**: 100% - Used by all games

```go
// Framework provides abstract interfaces
type Component interface {
    Name() string
    Data() map[string]interface{}
}

type Entity struct {
    ID         string
    Components map[string]map[string]interface{}
}
```

### 2. Game State Management
**Location**: `internal/turnengine/`
- **Purpose**: Manages overall game state, players, and turn progression
- **Key Files**: `game_state.go`, `player.go`
- **Reusability**: 100% - Generic turn-based game management

### 3. Abstract Board System
**Location**: `internal/turnengine/board.go`
- **Purpose**: Defines interface for different coordinate systems
- **Reusability**: 100% - Supports hex, grid, graph, and 3D boards

```go
// Abstract interfaces support multiple coordinate systems
type Position interface {
    Hash() string
    String() string
    Equals(Position) bool
}

type Board interface {
    IsValidPosition(Position) bool
    GetDistance(Position, Position) int
    GetNeighbors(Position) []Position
}
```

### 4. Command Processing Pipeline
**Location**: `internal/turnengine/command.go`
- **Purpose**: Generic command validation and processing
- **Reusability**: 100% - Works with any game's command types

## Game-Specific Components (WeeWar)

### 1. Hex Board System
**Location**: `games/weewar/board.go`
- **Purpose**: Implements hexagonal grid mechanics
- **Game-Specific**: Yes - WeeWar uses hex coordinates

```go
// WeeWar-specific hex implementation
type HexPosition struct {
    Q, R int  // Axial coordinates
}

type HexBoard struct {
    width, height int
    terrain       map[string]string
    entities      map[string]string
    pathfinder    *AStarPathfinder
}
```

### 2. WeeWar Components
**Location**: `games/weewar/components.go`
- **Purpose**: Game-specific data structures
- **Game-Specific**: Yes - WeeWar unit attributes

```go
// WeeWar-specific component implementations
type PositionComponent struct {
    X, Y, Z float64
}

type HealthComponent struct {
    Current, Max int
}

type MovementComponent struct {
    Range, MovesLeft int
}

type CombatComponent struct {
    Attack, Defense int
}
```

### 3. Combat System
**Location**: `games/weewar/combat.go`
- **Purpose**: Handles WeeWar-specific combat mechanics
- **Game-Specific**: Yes - Uses real WeeWar damage matrices
- **Data Source**: `weewar-data.json` (extracted from real game)

```go
type WeeWarCombatSystem struct {
    unitData     map[string]UnitData
    damageMatrix map[string]map[string]DamageDistribution
}
```

### 4. Movement System
**Location**: `games/weewar/movement.go`
- **Purpose**: Terrain-specific movement costs and pathfinding
- **Game-Specific**: Yes - WeeWar terrain movement costs
- **Data Source**: `weewar-data.json`

### 5. Map System
**Location**: `games/weewar/map.go`
- **Purpose**: Manages map configurations and layouts
- **Game-Specific**: Yes - WeeWar map formats
- **Data Source**: `weewar-maps.json` (extracted from HTML)

### 6. Data Integration
**Location**: `games/weewar/combat.go`, `games/weewar/cmd/extract-data/`
- **Purpose**: Integrates real WeeWar game data
- **Game-Specific**: Yes - WeeWar-specific data formats
- **Sources**: HTML files from tinyattack.com

## Integration Points

### 1. Component Registration
```go
// Framework provides registration mechanism
func RegisterWeeWarComponents(registry *ComponentRegistry) {
    registry.Register("position", &PositionComponent{})
    registry.Register("health", &HealthComponent{})
    registry.Register("movement", &MovementComponent{})
    // ... other WeeWar components
}
```

### 2. System Integration
```go
// WeeWar systems implement framework interfaces
func (wcs *WeeWarCombatSystem) Name() string { return "WeeWarCombatSystem" }
func (wcs *WeeWarCombatSystem) Priority() int { return 100 }
func (wcs *WeeWarCombatSystem) Update(world *World) error { /* ... */ }
```

### 3. Command Processing
```go
// Framework provides command pipeline
type MoveCommandValidator struct {
    movementSystem *WeeWarMovementSystem
}

func (mcv *MoveCommandValidator) ValidateCommand(gameState *GameState, command *Command) error {
    // WeeWar-specific validation logic
}
```

## Data Flow

### 1. Game Initialization
1. **Framework**: Creates base game state and world
2. **WeeWar**: Registers components and systems
3. **WeeWar**: Loads real game data from JSON files
4. **WeeWar**: Initializes map and starting units

### 2. Turn Processing
1. **Framework**: Manages turn progression
2. **Framework**: Processes commands through pipeline
3. **WeeWar**: Validates commands using game rules
4. **WeeWar**: Executes commands using systems
5. **Framework**: Updates game state

### 3. Game Logic
1. **Framework**: Provides ECS foundation
2. **WeeWar**: Implements specific game mechanics
3. **WeeWar**: Uses real WeeWar data for calculations
4. **Framework**: Manages entity lifecycle

## Key Design Principles

### 1. Separation of Concerns
- **Framework**: Provides infrastructure and abstractions
- **WeeWar**: Implements specific game mechanics and rules

### 2. Data-Driven Design
- **Real Data**: WeeWar uses authentic game data extracted from HTML
- **Configurability**: Maps and units defined in JSON files
- **Validation**: Game calculations match original WeeWar

### 3. Extensibility
- **Abstract Interfaces**: Easy to add new board types (Grid, Graph, 3D)
- **Component System**: Easy to add new unit types and abilities
- **System Architecture**: Easy to add new game mechanics

### 4. Reusability
- **Framework**: 80% of code is reusable across games
- **Game-Specific**: 20% of code is WeeWar-specific
- **Future Games**: Neptune's Pride can reuse most framework code

## File Organization

```
turnengine/
├── internal/turnengine/          # Framework (80% - Reusable)
│   ├── entity.go
│   ├── component.go
│   ├── game_state.go
│   ├── board.go
│   ├── command.go
│   └── world.go
└── games/weewar/                 # WeeWar Implementation (20% - Game-specific)
    ├── board.go                  # Hex board implementation
    ├── components.go             # WeeWar components
    ├── combat.go                 # Combat system
    ├── movement.go               # Movement system
    ├── map.go                    # Map system
    ├── game.go                   # Game initialization
    ├── weewar-data.json          # Real game data
    ├── weewar-maps.json          # Map configurations
    └── cmd/
        ├── extract-data/         # Data extraction tools
        └── extract-map-data/
```

## Summary

The WeeWar implementation demonstrates a clean separation between framework and game-specific code. The TurnEngine framework provides all the infrastructure for turn-based games, while WeeWar adds the specific mechanics, data, and rules that make it authentic to the original game. This architecture enables rapid development of new games while maintaining code reusability and authentic gameplay mechanics.
