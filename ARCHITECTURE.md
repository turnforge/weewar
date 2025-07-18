# WeeWar Architecture Overview

## Current Architecture (Phaser-First v4.0)

### Major Architectural Transformation (January 2025)
- **Phaser.js Editor Integration**: Complete migration from canvas-based to Phaser.js-based map editor
- **Clean Component Architecture**: Introduced PhaserPanel class for separation of concerns
- **Improved Coordinate System**: Fixed coordinate conversion to match Go backend implementation exactly
- **Dynamic Grid System**: Infinite grid that updates based on camera viewport
- **Professional UX**: Intuitive mouse interaction with paint modes and drag behavior
- **Legacy System Removal**: Completely eliminated old canvas system and related complexity
- **UI Reorganization**: Moved view controls to logical locations with Phaser editor tools

### Core Components

#### 1. KeyboardShortcutManager (`web/frontend/components/KeyboardShortcutManager.ts`) ✅ COMPLETED
**Purpose**: Reusable keyboard shortcut system for all pages
- **Configuration-driven**: Define shortcuts declaratively with ShortcutConfig interface
- **State Machine**: Handle multi-key commands (n12, c5, u3) with timeout and visual feedback
- **Context Awareness**: Disable shortcuts in input fields, modals, and other contexts
- **Help System**: Auto-generated help overlay from shortcut configuration
- **Clean Architecture**: Pure input handling with no UI dependencies
- **Framework Agnostic**: Can be used across Map Editor, Game Play, and Detail pages

```typescript
interface ShortcutConfig {
  key: string;
  handler: (args?: string) => void;
  description: string;
  category?: string;
  requiresArgs?: boolean;
  argType?: 'number' | 'string';
  contextFilter?: (event: KeyboardEvent) => boolean;
}

interface ShortcutManagerConfig {
  shortcuts: ShortcutConfig[];
  helpContainer?: string;
  timeout?: number; // ms to return to normal state
}

enum KeyboardState {
  NORMAL = 'normal',
  AWAITING_ARGS = 'awaiting_args'
}

class KeyboardShortcutManager {
  private shortcuts: Map<string, ShortcutConfig>;
  private state: KeyboardState;
  private currentCommand: string;
  private currentArgs: string;
  private helpOverlay: HTMLElement | null;
  
  constructor(config: ShortcutManagerConfig) {
    // Pure input handling - no UI dependencies
    // Registers global keydown listener
    // Manages state machine and help overlay
  }
  
  // Key methods
  private handleKeydown(event: KeyboardEvent): void
  private executeShortcut(shortcut: ShortcutConfig, args?: string): void
  private showHelp(): void
  private updateStateIndicator(): void
  public destroy(): void
}
```

**Architecture Benefits**:
- **Separation of Concerns**: Pure input handling separate from UI updates
- **Reusability**: Generic class usable across all application pages
- **Testability**: Pure functions with no external dependencies
- **Maintainability**: Configuration-driven with clear interfaces
- **Extensibility**: Easy to add new shortcuts and categories

#### 2. Game Object (`lib/game.go`)
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
- Normalized origin management (OriginX/OriginY in tile units)
- Direct hex-to-pixel conversion using Red Blob Games formulas
- Dynamic map resizing with Add/Remove methods
- Efficient tile storage via `map[AxialCoord]*Tile`

```go
type Map struct {
    // Coordinate bounds - dynamic and expandable
    minQ, maxQ, minR, maxR int
    
    // Normalized origin for coordinate system (in tile units)
    OriginX, OriginY float64
    
    // Cube coordinate storage - primary data structure
    Tiles map[AxialCoord]*Tile
}
```

#### 4. LayeredRenderer (`lib/layered_renderer.go`)
**Purpose**: Efficient multi-layer rendering with platform abstraction
- Layer-based rendering system (TileLayer, UnitLayer, UILayer)
- Dirty tracking for efficient partial updates
- Platform-agnostic via Drawable interface
- Batched rendering with scheduler
- Viewport management for scrolling

```go
type LayeredRenderer struct {
    drawable Drawable
    x, y, width, height int
    layers []Layer
    outputBuffer *Buffer
    renderOptions LayerRenderOptions
}
```

#### 5. WorldEditor (`lib/editor.go`)
**Purpose**: Map editing interface with clean architecture
- Works directly with World objects (no Game intermediates)
- Delegates all rendering to LayeredRenderer
- Viewport scrolling support (scrollX, scrollY)
- Platform-agnostic via Drawable interface
- Cube coordinate native operations

```go
type WorldEditor struct {
    currentWorld *World
    drawable Drawable
    layeredRenderer *LayeredRenderer
    scrollX, scrollY int
    tileLayer *TileLayer
    unitLayer *UnitLayer
}
```

#### 6. Observer Pattern (`lib/world_observer.go`)
**Purpose**: Reactive updates for state changes
- WorldSubject embedded in World for notifications
- WorldObserver interface for components requiring updates
- Event batching for performance optimization

### Key Design Principles

#### 1. Separation of Concerns
- **Game**: Flow control, rules, validation
- **World**: Pure state storage
- **WorldEditor**: Map editing logic
- **LayeredRenderer**: Multi-layer rendering
- **CLI**: Translation layer (chess notation ↔ cube coordinates)
- **WASM**: JS↔Go conversion only

#### 2. Cube Coordinate System (Universal)
- Primary coordinate system throughout codebase
- Proper hex mathematics using Red Blob Games formulas
- CLI preserves chess notation for user experience
- Efficient coordinate conversion and validation
- Direct cube-to-pixel conversion (no row/col intermediates)

#### 3. Clean Architecture
- No circular dependencies
- Interface-driven design (Drawable, Layer, AssetProvider)
- Dependency injection for testability
- Clear data flow: CLI/WASM → Editor → World → LayeredRenderer → Drawable

#### 4. Platform Abstraction
- **Drawable interface**: Supports Buffer (CLI/PNG) and CanvasBuffer (Web)
- **LayeredRenderer**: Works with any Drawable implementation
- **WorldEditor**: Platform-agnostic editing operations
- **AssetProvider**: Embedded assets + fetch-based assets

#### 5. Performance Optimizations
- Direct Map.Tiles access (no copying)
- Efficient cube coordinate storage
- Layer-based dirty tracking
- Event batching for multiple state changes
- Asset caching and fallback rendering
- Viewport culling for large maps

### WASM Architecture (v3.0 Refactoring)

#### Problem Statement
Original WASM code had massive boilerplate:
- ~1300 lines with repetitive validation
- Manual Game object creation for every operation
- Complex initialization requiring browser calls
- Inconsistent error handling and response formats

#### New WASM Architecture
```go
// Global state initialized in main()
var globalEditor *weewar.WorldEditor
var globalWorld *weewar.World
var globalAssetProvider weewar.AssetProvider

// Generic wrapper infrastructure
type WASMFunction func(args []js.Value) (interface{}, error)

func createWrapper(minArgs, maxArgs int, fn WASMFunction) js.Func {
    // Validates arguments and handles errors uniformly
    // No reflection - direct js.Value handling
}
```

#### Key WASM Improvements
1. **Immediate Initialization**: Editor/World/Assets ready on WASM load
2. **Zero Boilerplate**: Functions like `paintTerrain(args []js.Value)` handle own conversion
3. **No Game Objects**: Direct World manipulation via WorldEditor
4. **Generic Wrapper**: Argument validation and error handling abstracted
5. **Consistent Responses**: Standardized success/error JSON format
6. **Performance**: No reflection, direct `args[0].Int()` calls

#### WASM Function Pattern
```go
// Old (100+ lines with boilerplate)
func paintTerrain(this js.Value, args []js.Value) any {
    if globalEditor == nil {
        return createEditorResponse(false, "", "Editor not initialized", nil)
    }
    if len(args) < 2 {
        return createEditorResponse(false, "", "Missing arguments", nil)
    }
    // ... more boilerplate validation, error handling, response formatting
}

// New (clean, no boilerplate)
func paintTerrain(args []js.Value) (interface{}, error) {
    row := args[0].Int()
    col := args[1].Int()
    coord := weewar.AxialCoord{Q: col, R: row}
    return nil, globalEditor.PaintTerrain(coord)
}

// Registration (one line)
js.Global().Set("editorPaintTerrain", createWrapper(2, 2, paintTerrain))
```

#### Current WASM Implementation Highlights

**Initialization Pattern**:
```go
func main() {
    // Immediate initialization - no browser calls needed
    globalWorld = &weewar.World{
        Map:           weewar.NewMapWithBounds(0, 0, 0, 0),
        UnitsByPlayer: make([][]*weewar.Unit, 2),
        PlayerCount:   2,
    }
    
    globalEditor = weewar.NewWorldEditor()
    globalEditor.NewWorld()
    
    globalAssetProvider = assets.NewEmbeddedAssetManager()
    globalAssetProvider.PreloadCommonAssets()
    globalEditor.SetAssetProvider(globalAssetProvider)
    
    registerEditorFunctions()
    registerUtilityFunctions()
}
```

**Function Implementation Pattern**:
```go
func paintTerrain(args []js.Value) (interface{}, error) {
    row := args[0].Int()
    col := args[1].Int()  
    coord := weewar.AxialCoord{Q: col, R: row}
    return nil, globalEditor.PaintTerrain(coord)
}

func pixelToCoords(args []js.Value) (interface{}, error) {
    x := args[0].Float()
    y := args[1].Float()
    
    coord := globalWorld.Map.XYToQR(x, y, 
        weewar.DefaultTileWidth, 
        weewar.DefaultTileHeight, 
        weewar.DefaultYIncrement)
    
    return map[string]interface{}{
        "row":          coord.R,
        "col":          coord.Q,
        "cubeQ":        coord.Q,
        "cubeR":        coord.R,
        "withinBounds": globalWorld.Map.IsWithinBoundsCube(coord),
    }, nil
}
```

### Coordinate System

#### Cube Coordinates (Internal)
- Primary system: `AxialCoord{Q, R}` with `S = -Q - R`
- Bounds: `minQ/maxQ/minR/maxR` (dynamic, expandable)
- Normalized origin: `OriginX/OriginY` in tile units
- Universal throughout: Map, Editor, Renderer, WASM

#### Display Coordinates (User Interface)
- CLI: Chess notation (A1, B2, C3...)
- Internal conversion: Chess → Row/Col → Cube (CLI only)
- WASM: Direct cube coordinates (browser handles display)
- Preserved user experience with mathematical correctness

#### Pixel Coordinates (Rendering)
- Direct conversion: Cube → Pixel using `CenterXYForTile()`
- Pointy-topped hexagons with odd-r layout
- Formula: `x = originX + tileWidth * sqrt(3) * (q + r/2)`
- No scaling - DefaultTileWidth/Height constants
- Viewport via scrollX/scrollY offset

### Data Flow

#### CLI Flow
```
User Input → CLI (chess→cube) → Game (logic) → World (state) → WorldRenderer → Buffer → PNG
```

#### WASM Flow  
```
JS Call → Wrapper → Editor Method → World (state) → LayeredRenderer → CanvasBuffer → HTML Canvas
```

#### Editor Flow
```
User Action → WorldEditor → World (state) → LayerDirty → LayeredRenderer → Drawable
```

### Rendering Architecture

#### Layer System
```go
type Layer interface {
    Render(world *World, options LayerRenderOptions)
    MarkDirty(coord AxialCoord)
    MarkAllDirty()
    // ...
}

// Concrete layers
- TileLayer: Terrain rendering with assets/fallback
- UnitLayer: Unit rendering with player colors  
- UILayer: Highlights, selections, grid overlay
```

#### Rendering Pipeline
1. **Dirty Tracking**: Only render changed tiles/units
2. **Layer Composition**: Each layer renders to own Buffer
3. **Platform Output**: Composite to Drawable (Buffer or CanvasBuffer)
4. **Viewport Culling**: Skip off-screen tiles

#### Platform Support
- **CLI**: Buffer → PNG file output
- **Web**: CanvasBuffer → HTML Canvas (direct rendering)
- **Future**: Easy to add new platforms via Drawable interface

### File Organization

#### Core Library (`lib/`)
- **Core State**: `game.go`, `world.go`, `map.go`, `hex_coords.go`
- **Rendering**: `layered_renderer.go`, `layers.go`, `world_renderer.go`
- **Interfaces**: `drawable.go`, `game_interface.go`, `asset_interface.go`
- **Platform**: `buffer.go`, `canvas_buffer.go`, `canvas_renderer.go`
- **Editor**: `editor.go`

#### Commands (`cmd/`)
- **CLI**: `cmd/weewar-cli/` (proper separation)
- **WASM**: `cmd/editor-wasm/` (clean, minimal)

#### Backup Files
- Legacy implementations moved to `.bak` extensions
- Clean separation between old and new architecture

### Migration Status (Completed Features)

#### ✅ Coordinate System Migration
- Complete cube coordinate implementation
- Eliminated all row/col coordinate confusion
- Proper hex mathematics throughout
- Dynamic map bounds with Add/Remove methods
- Normalized coordinate system with origin management
- Direct cube-to-pixel conversion using Red Blob Games formulas

#### ✅ Rendering Architecture
- LayeredRenderer with Layer abstraction
- Platform-agnostic Drawable interface  
- Proper Game-World-Renderer separation
- Efficient dirty tracking and viewport management
- Layer-based rendering system (TileLayer, UnitLayer, UILayer)
- Clean architectural separation between state and rendering

#### ✅ WASM Refactoring (Major Improvements)
- Generic wrapper infrastructure (completed)
- Immediate initialization in main() (completed)
- Direct js.Value handling (no reflection)
- Clean function implementations (major boilerplate reduction)
- Eliminated ~1300 lines of repetitive validation code
- Consistent error handling and response formatting
- Performance improvements through direct type conversion

#### ✅ Editor Architecture
- WorldEditor works directly with World
- Delegates all rendering to LayeredRenderer
- Proper cube coordinate operations
- Platform-agnostic via Drawable
- Viewport scrolling support (scrollX, scrollY)
- Clean separation from Game flow control

#### ✅ File Organization
- CLI components moved to cmd/weewar-cli/
- Clear separation between library and command implementations
- Proper architectural boundaries maintained
- Legacy implementations preserved as .bak files

### Testing Strategy

#### Unit Tests
- Coordinate conversion accuracy
- Game logic validation
- Observer pattern functionality
- Layer rendering verification

#### Integration Tests
- CLI command compatibility
- WASM function integration
- Save/load functionality
- Cross-component communication

#### Performance Tests
- Large map rendering
- Coordinate conversion benchmarks
- Memory usage optimization
- Asset loading performance

### Development Guidelines

#### Coordinate Usage
- **Always use cube coordinates** in internal APIs
- CLI translates chess notation at boundary only
- Direct AxialCoord{Q, R} in all function signatures
- Use Map bounds, not NumRows/NumCols

#### Rendering Principles
- Delegate to LayeredRenderer for all drawing
- Mark dirty for efficient updates
- Use Drawable interface for platform abstraction
- Default tile dimensions (no scaling)

#### WASM Development
- Functions take `[]js.Value` and handle own conversion
- Use createWrapper for all JS exports
- No Game objects - direct World manipulation
- Initialize everything in main()

#### Error Handling
- Return errors, don't print to console
- Use wrapper for consistent error responses
- Validate at boundaries, not internal APIs
- Clean error messages for users

### Performance Characteristics

#### Memory Usage
- Efficient cube coordinate storage
- Asset caching with fallback generation
- Layer buffers only for dirty regions
- No temporary object creation

#### Rendering Performance  
- Dirty tracking minimizes redraws
- Viewport culling for large maps
- Direct coordinate conversion (no lookups)
- Platform-optimized drawing (Buffer vs Canvas)

#### Coordinate Performance
- O(1) coordinate conversion
- Direct map access via AxialCoord keys
- No array iteration for bounds checking
- Efficient hex mathematics

### Future Extensions

#### Web Frontend Architecture (v4.0 - Phaser-First)

#### Frontend Component Structure
```
MapEditorPage.ts (Main Controller)
├── PhaserPanel.ts (Editor Logic)
│   ├── PhaserMapEditor.ts (Phaser Game Management)
│   └── PhaserMapScene.ts (Scene Logic & Rendering)
├── ToolsPanel (Terrain & Brush Controls)
├── AdvancedToolsPanel (Phaser Controls & View Options)
└── ConsolePanel (Logging & Debug)
```

#### Phaser.js Integration
- **PhaserMapEditor**: High-level API for tile management and event handling
- **PhaserMapScene**: Core Phaser scene with WebGL rendering and input handling
- **Coordinate Accuracy**: Matches Go backend implementation (`lib/map.go`) exactly
- **Dynamic Grid**: Infinite grid system that renders only visible hexes
- **Interactive Controls**: Mouse wheel zoom, drag pan, modifier key painting

#### Mouse Interaction System
```javascript
// Normal Click: Paint tile on mouse up (no accidental painting)
click → release → paint tile

// Drag without modifiers: Pan camera view
mousedown → drag → pan camera

// Paint Mode: Hold Alt/Cmd/Ctrl + drag to paint continuously
Alt/Cmd + mousedown → immediate paint → drag → continuous painting
```

#### Component Communication
```
UI Controls → MapEditorPage → PhaserPanel → PhaserMapEditor → PhaserMapScene
                ↓                                            ↓
            Logging & State                              Phaser Rendering
```

#### Key Frontend Features
- **Professional UX**: No accidental tile painting during camera movement
- **Efficient Rendering**: Only renders visible grid hexes based on camera bounds
- **Clean Architecture**: Each component has single responsibility
- **Type Safety**: Full TypeScript integration with proper type definitions
- **Event-Driven**: Clean callback system for tile clicks and map changes

### Coordinate System Accuracy

#### Fixed Implementation (v4.0)
The coordinate conversion now exactly matches the Go backend:
- **Backend**: `lib/map.go` CenterXYForTile() and XYToQR() functions
- **Frontend**: PhaserMapScene with matching tileWidth=64, tileHeight=64, yIncrement=48
- **Conversion**: Uses row/col intermediate step with odd-row offset layout
- **Accuracy**: Pixel-perfect coordinate matching between frontend and backend

#### Benefits of Accurate Coordinates
- **No Coordinate Drift**: Frontend and backend always agree on tile positions
- **Precise Interaction**: Click coordinates map exactly to intended hexes  
- **Seamless Integration**: Easy integration with WASM and backend APIs
- **Mathematical Correctness**: Proper hex geometry throughout the system

### Future Extensions

#### Planned Features
1. **Advanced Editor**: Multi-tile selection, copy/paste, templates via Phaser
2. **Network Play**: Real-time multiplayer with WebSocket integration
3. **Mobile Support**: Touch-friendly controls via Phaser input system
4. **Performance**: Further Phaser optimizations (sprite batching, culling)
5. **AI Integration**: Clean World state for AI decision making
6. **Advanced UI**: Animation systems, visual effects, improved UX

#### Architecture Benefits for Extensions
- **Phaser.js Foundation**: Professional game engine enables advanced features
- **Clean Component Structure**: Easy to extend with new panels and tools
- **Accurate Coordinates**: Reliable foundation for complex spatial features
- **Event-Driven Design**: Simple to add new interactions and behaviors
- **TypeScript Safety**: Prevents runtime errors and improves development experience

#### Web Technology Stack
- **Phaser.js 3.x**: WebGL-accelerated rendering engine
- **TypeScript**: Type-safe frontend development
- **Tailwind CSS**: Utility-first styling system
- **DockView**: Professional panel management system
- **Webpack**: Module bundling and hot reload development

---

**Last Updated**: 2025-01-17  
**Architecture Version**: 4.0 (Phaser-First Frontend)  
**Status**: Production-ready Phaser.js editor with accurate coordinate system

**Key Achievement**: Complete migration to modern web technologies with professional UX and pixel-perfect coordinate accuracy matching the Go backend implementation.
