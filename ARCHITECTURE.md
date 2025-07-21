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

### Rules Engine Integration (January 2025) ✅
- **Data-Driven Game Mechanics**: Complete replacement of hardcoded logic with rules engine
- **Enhanced Game Constructor**: NewGame now requires RulesEngine parameter for proper initialization  
- **Movement System Integration**: Terrain passability and cost validation through rules data
- **Combat System Enhancement**: Probabilistic damage with counter-attacks using DamageDistribution
- **Attack Validation Integration**: Rules-based unit attack capability checking
- **Test System Migration**: All tests updated to AxialCoord system with proper unit initialization
- **API Consistency**: Unified pattern where game mechanics go through rules engine first with fallbacks

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

### Unified Map Architecture (v5.0)

#### Map Class Enhancement
**Purpose**: Single source of truth for all map data with Observer pattern
- **Observer Pattern**: MapObserver interface with type-safe MapEvent system
- **Batched Events**: TileChange and UnitChange arrays with setTimeout scheduling
- **Self-contained Persistence**: Map handles save/load operations directly
- **Automatic Change Tracking**: Eliminates manual change marking
- **Event Types**: TILES_CHANGED, UNITS_CHANGED, MAP_LOADED, MAP_SAVED, MAP_CLEARED, MAP_METADATA_CHANGED

### Component Lifecycle Architecture (v6.0)

#### Breadth-First Component Initialization Pattern
**Purpose**: Eliminate initialization order dependencies through synchronization barriers and multi-phase lifecycle management

The new lifecycle architecture implements a breadth-first initialization pattern that prevents race conditions and timing issues common in depth-first component construction. Instead of each component immediately initializing its children, we use synchronized phases that ensure all components at each level are ready before proceeding to the next phase.

#### ✅ COMPLETED: Simplified Lifecycle Architecture
Successfully implemented and simplified the complete lifecycle architecture:

**Final Architecture**:
- **BaseComponent Auto-Lifecycle**: All components auto-initialize AND implement ComponentLifecycle with empty defaults
- **Opt-in Coordination**: Components only override lifecycle methods they actually need for coordination
- **Zero Breaking Changes**: Existing components continue working exactly as before
- **No Boilerplate**: Components don't need to declare `implements ComponentLifecycle` anymore

**Implementation Details**:
- **ComponentLifecycle Interface**: Multi-phase initialization (initializeDOM, injectDependencies, activate, deactivate)
- **LifecycleController**: Breadth-first orchestration with synchronization barriers
- **Explicit Dependency Setters**: Parent components directly set dependencies using setters/getters
- **EventBus Communication**: Loose coupling via events instead of direct component dependencies
- **Source Filtering**: Components only handle events NOT originating from themselves to prevent loops
- **BaseComponent Integration**: Every component extends BaseComponent and gets lifecycle support automatically

**Completed Component Migrations**:
- **ReferenceImagePanel**: Full EventBus communication with PhaserEditorComponent, no direct dependencies
- **EditorToolsPanel**: Lifecycle-enabled with deferred execution and explicit page state dependency
- **TileStatsPanel**: Migrated from standalone to BaseComponent with lifecycle and Map dependency
- **MapEditorPage**: Uses LifecycleController for coordinated component initialization

#### Multi-Phase Lifecycle Design
```typescript
interface ComponentLifecycle {
    // Phase 1: Basic construction and DOM binding
    bindToDOM(): Promise<void>;
    
    // Phase 2: Dependency injection and configuration
    injectDependencies(dependencies: ComponentDependencies): Promise<void>;
    
    // Phase 3: Full activation and event subscription
    activate(): Promise<void>;
    
    // Cleanup phase
    deactivate(): Promise<void>;
}

interface ComponentDependencies {
    eventBus: EventBus;
    sharedState: any;
    parentContext: ComponentContext;
    configurationData: any;
}
```

#### LifecycleController for Breadth-First Orchestration
**Purpose**: Coordinates component initialization across multiple phases to prevent race conditions

```typescript
export class LifecycleController {
    private components: Map<string, ComponentLifecycle> = new Map();
    private currentPhase: LifecyclePhase = LifecyclePhase.IDLE;
    private phaseBarriers: Map<LifecyclePhase, Set<string>> = new Map();
    
    // Register component for lifecycle management
    public registerComponent(id: string, component: ComponentLifecycle): void;
    
    // Execute all phases in breadth-first order
    public async initializeAll(): Promise<void> {
        await this.executePhase(LifecyclePhase.BIND_TO_DOM);
        await this.executePhase(LifecyclePhase.INJECT_DEPENDENCIES);
        await this.executePhase(LifecyclePhase.ACTIVATE);
    }
    
    // Execute a single phase for all components
    private async executePhase(phase: LifecyclePhase): Promise<void> {
        const promises = Array.from(this.components.values()).map(async (component) => {
            try {
                await this.executeComponentPhase(component, phase);
                this.markPhaseComplete(component.componentId, phase);
            } catch (error) {
                this.handlePhaseError(component.componentId, phase, error);
            }
        });
        
        await Promise.allSettled(promises);
        await this.waitForPhaseBarrier(phase);
    }
    
    // Synchronization barrier - wait for all components to complete phase
    private async waitForPhaseBarrier(phase: LifecyclePhase): Promise<void>;
}

enum LifecyclePhase {
    IDLE = 'idle',
    BIND_TO_DOM = 'bind-to-dom',
    INJECT_DEPENDENCIES = 'inject-dependencies', 
    ACTIVATE = 'activate',
    DEACTIVATING = 'deactivating'
}
```

#### Benefits Over Depth-First Initialization

**Eliminates Race Conditions**: Components don't emit events until all components are ready to receive them
- **Traditional Problem**: Component A initializes and emits events before Component B has subscribed
- **Breadth-First Solution**: All components bind to DOM first, then all inject dependencies and subscribe to events, then all activate

**Prevents Initialization Order Dependencies**: Components can be created in any order
- **Traditional Problem**: Component creation order determines whether dependencies are available
- **Breadth-First Solution**: Dependencies are injected in a separate phase after all components exist

**Handles Async Initialization Gracefully**: Each phase can be async without blocking other components
- **Traditional Problem**: Async component initialization blocks dependent components indefinitely
- **Breadth-First Solution**: Phase barriers ensure all async operations complete before proceeding

**Provides Clear Error Isolation**: Failed component initialization doesn't cascade to other components
- **Traditional Problem**: One component failure can prevent entire application initialization
- **Breadth-First Solution**: Failed components are isolated, remaining components continue initialization

#### Implementation Example
```typescript
export class MapEditorPage extends BasePage {
    private lifecycleController: LifecycleController;
    
    protected async initializeComponents(): Promise<void> {
        this.lifecycleController = new LifecycleController();
        
        // Register all components first (they only create basic structure)
        const editorToolsPanel = new EditorToolsPanel(this.ensureElement('[data-component="editor-tools"]'));
        const phaserEditor = new PhaserEditorComponent(this.ensureElement('[data-component="phaser-editor"]'));
        const tileStatsPanel = new TileStatsPanel(this.ensureElement('[data-component="tile-stats"]'));
        
        this.lifecycleController.registerComponent('editor-tools', editorToolsPanel);
        this.lifecycleController.registerComponent('phaser-editor', phaserEditor);
        this.lifecycleController.registerComponent('tile-stats', tileStatsPanel);
        
        // Execute breadth-first initialization
        await this.lifecycleController.initializeAll();
        
        // All components are now fully initialized and ready
    }
}
```

#### Component Encapsulation Pattern
**Enhanced Component Base Class**: Implements ComponentLifecycle interface
```typescript
export abstract class BaseComponent implements Component, ComponentLifecycle {
    // Phase 1: Create DOM structure and find/create elements
    public async bindToDOM(): Promise<void> {
        this.findOrCreateElements();
        this.validateDOMStructure();
    }
    
    // Phase 2: Receive dependencies and configuration
    public async injectDependencies(deps: ComponentDependencies): Promise<void> {
        this.eventBus = deps.eventBus;
        this.sharedState = deps.sharedState;
        this.configureFromData(deps.configurationData);
    }
    
    // Phase 3: Subscribe to events and become fully active
    public async activate(): Promise<void> {
        this.subscribeToEvents();
        this.initializeBusinessLogic();
        this.markReady();
    }
    
    // Cleanup in reverse order
    public async deactivate(): Promise<void> {
        this.unsubscribeFromEvents();
        this.cleanupBusinessLogic();
        this.clearDependencies();
    }
}
```

#### Architectural Benefits
- **Predictable Initialization**: All components go through same phases in same order
- **Race Condition Prevention**: Events only flow when all components are ready to handle them
- **Error Resilience**: Component failures are isolated and don't prevent other components from initializing
- **Debugging Simplicity**: Clear phase boundaries make initialization issues easier to trace
- **Async-Safe**: Properly handles async operations without blocking other components
- **Testability**: Each phase can be tested independently with mocked dependencies

```typescript
export interface MapObserver {
    onMapEvent(event: MapEvent): void;
}

export interface MapEvent {
    type: MapEventType;
    data: any;
}

export class Map {
    // Core data
    private metadata: MapMetadata;
    private tiles: { [key: string]: TileData } = {};
    private units: { [key: string]: UnitData } = {};
    
    // Observer pattern
    private observers: MapObserver[] = [];
    private pendingTileChanges: TileChange[] = [];
    private pendingUnitChanges: UnitChange[] = [];
    private batchTimeout: number | null = null;
    
    // Methods for Observer pattern
    public subscribe(observer: MapObserver): void
    public unsubscribe(observer: MapObserver): void
    private emit(event: MapEvent): void
    
    // Batched change management
    private scheduleBatchEmit(): void
    private flushBatchedChanges(): void
    
    // Self-contained persistence
    public async save(): Promise<SaveResult>
    public async load(mapId: string): Promise<void>
    public loadFromElement(elementId: string): void
    public loadFromData(data: any): void
}
```

#### Component Integration Pattern
```typescript
// MapEditorPage implements MapObserver
export class MapEditorPage extends BasePage implements MapObserver {
    private map: Map;
    
    constructor() {
        // Create Map instance as single source of truth
        this.map = new Map();
        this.map.subscribe(this); // Subscribe to changes
    }
    
    // Implement Observer interface
    public onMapEvent(event: MapEvent): void {
        switch (event.type) {
            case MapEventType.TILES_CHANGED:
                this.handleTilesChanged(event.data);
                break;
            case MapEventType.MAP_SAVED:
                this.handleMapSaved(event.data);
                break;
            // Handle other events...
        }
    }
    
    // Use Map as single source of truth
    private save(): void {
        this.map.save(); // Map handles persistence
    }
}
```

#### Architecture Benefits
- **Single Source of Truth**: All map data flows through Map class
- **Event-Driven Updates**: Components automatically stay synchronized
- **Performance**: Batched events reduce UI update frequency
- **Maintainability**: Centralized map logic easier to debug and extend
- **Type Safety**: Comprehensive TypeScript interfaces prevent errors
- **Clean Separation**: Components focus on UI, Map handles data

### Future Extensions

#### Planned Features
1. **Component Integration**: Complete Observer pattern integration across all components
2. **Advanced Editor**: Multi-tile selection, copy/paste, templates via unified Map
3. **Network Play**: Real-time multiplayer with Map state synchronization
4. **Mobile Support**: Touch-friendly controls via Phaser input system
5. **Performance**: Optimized Map operations and event batching
6. **AI Integration**: Clean Map state for AI decision making

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

**Last Updated**: 2025-01-21  
**Architecture Version**: 8.0 (Game Mechanics Architecture Design)  
**Status**: Production-ready UI framework with comprehensive game engine foundation discovered. Ready for rules integration and WASM activation.

**Latest Achievement (v7.0)**: Comprehensive UI framework completion and game foundation:

## Game Mechanics Architecture (v8.0 Design)

### Discovered Foundation: Strong Game Engine Already Exists ✅

#### Comprehensive Game Class (lib/game.go)
**Purpose**: Complete turn-based game state management with multiplayer support
- **Game State Management**: CurrentPlayer, TurnCounter, Status (playing/paused/ended)
- **Turn System**: NextTurn(), EndTurn(), CanEndTurn() with player cycling
- **Victory Conditions**: checkVictoryConditions() with last-player-standing logic
- **Save/Load System**: JSON serialization with complete state persistence
- **Event System**: EventManager with game state change notifications
- **Deterministic Gameplay**: RNG with seed for reproducible game sessions
- **Asset Management**: AssetProvider interface for platform flexibility

```go
type Game struct {
    World *World           // Pure state (map, units, entities)
    CurrentPlayer int      // 0-based player index
    TurnCounter int        // 1-based turn number
    Status GameStatus      // playing/paused/ended
    Seed int64             // Random seed for deterministic gameplay
    rng *rand.Rand         // RNG for combat calculations
    eventManager *EventManager // Observer pattern for state changes
    assetProvider AssetProvider // Platform-agnostic asset management
}
```

#### Movement & Combat System (lib/moves.go)
**Purpose**: Complete unit movement and combat mechanics with validation
- **Movement Validation**: IsValidMove(), CanMoveUnit() with player turn checking
- **Combat System**: AttackUnit(), CanAttackUnit() with damage calculation
- **Pathfinding**: FindPath() with A* foundation (currently simplified)
- **Movement Costs**: GetMovementCost() using proper hex distance calculation
- **Attack Range**: GetUnitAttackRange() with unit-type specific ranges
- **Coordinate System**: Full AxialCoord (cube coordinates) throughout

```go
// Current movement system
func (g *Game) MoveUnit(unit *Unit, to AxialCoord) error
func (g *Game) AttackUnit(attacker, defender *Unit) (*CombatResult, error)
func (g *Game) IsValidMove(from, to AxialCoord) bool
func (g *Game) GetMovementCost(from, to AxialCoord) int
func (g *Game) calculateDamage(attacker, defender *Unit) int
```

#### Professional CLI Interface (cmd/weewar-cli/)
**Purpose**: Complete command-line interface for gameplay and testing
- **Game Commands**: move, attack, status, map, units, player, help, save, load
- **Advanced Commands**: predict, attackoptions, moveoptions, autorender
- **REPL Mode**: Interactive gameplay with dynamic prompts `weewar[T1:P0]>`
- **Batch Processing**: Execute commands from files for automated testing
- **Session Recording**: Record and replay game sessions for analysis
- **Multiple Display Modes**: compact, detailed, ASCII, JSON output formats
- **Chess Notation**: A1, B2, C3 coordinate system for user-friendly interaction

### Architecture Gaps Analysis

#### What's Missing for Production Game Mechanics

**1. Rules Engine Integration**
```go
// Current: Hardcoded values in lib/moves.go
unit.DistanceLeft = 3 // TODO: Get from unit data
baseDamage := 30      // TODO: Use attack matrices

// Needed: Data-driven rules from weewar-data.json
type RulesEngine struct {
    unitData map[int]UnitData
    terrainMovement map[string]map[string]float64
    attackMatrices map[string]map[string]AttackData
}
```

**2. Map-to-Game Integration**
```go
// Current: Hardcoded test map creation
func (g *Game) initializeStartingUnits() // Fixed positions

// Needed: Initialize from Map editor data
func NewGameFromMap(mapData *Map, playerCount int) (*Game, error)
func (g *Game) ConvertPlayerToNeutral(playerID int) error
func (g *Game) RemovePlayer(playerID int) error
```

**3. WASM Game Module**
```go
// Current: cmd/weewar-wasm/main.go commented out
// Needed: Active WASM APIs for web interface
weewarCreateGameFromMap(mapData, playerCount)
weewarSelectUnit(q, r) → validMoves, validAttacks
weewarMoveUnit(fromQ, fromR, toQ, toR) → moveResult
weewarAttackUnit(attackerQ, attackerR, defenderQ, defenderR) → combatResult
```

**4. Move Recording System**
```go
// Current: CLI command recording only
// Needed: Structured game move logging
type GameMove struct {
    Turn int, Player int, Action string
    From/To AxialCoord, Timestamp time.Time
    Result interface{}, Valid bool
}
```

### v8.0 Game Mechanics Architecture Design

#### Rules Engine Pattern
**Design**: Pluggable rules system driven by weewar-data.json
```go
type RulesEngine struct {
    gameData *GameData
    unitCache map[int]*UnitData
    terrainCache map[string]*TerrainData
}

func (re *RulesEngine) GetMovementCost(unitType int, terrainType string) float64
func (re *RulesEngine) GetAttackDamage(attackerType, defenderType int) CombatResult
func (re *RulesEngine) GetUnitStats(unitType int) UnitStats
```

#### Game Initialization Flow
```
Map Editor → Map Data → NewGameFromMap() → Game Instance
    ↓
Rules Engine Loading → Unit Stats → Starting Positions
    ↓
CLI/WASM Interface → Player Actions → Move Validation
    ↓
EventBus Notifications → Web UI Updates → Phaser Rendering
```

#### WASM API Architecture
**Design**: Multiplayer-first validation APIs for current player actions
```go
// Game lifecycle management
weewarCreateGameFromMap(mapData, playerCount) → gameId
weewarGetGameState() → {currentPlayer, turn, units, status}

// Current player actions (validation-focused)
weewarSelectUnit(q, r) → {unitInfo, validMoves, validAttacks}
weewarMoveUnit(fromQ, fromR, toQ, toR) → {success, newPosition, movementLeft}
weewarAttackUnit(attackerQ, attackerR, defenderQ, defenderR) → {damage, health, killed}
weewarEndTurn() → {nextPlayer, turnNumber}

// Query methods for UI updates
weewarGetValidMoves(q, r) → [AxialCoord...]
weewarGetValidAttacks(q, r) → [AxialCoord...]
weewarGetUnitInfo(q, r) → {type, health, movement, canAct}
```

#### Move Recording Architecture
**Design**: Complete game session logging for testing and replay
```go
type GameSession struct {
    GameID string
    InitialState GameState
    Moves []GameMove
    Created time.Time
}

type GameMove struct {
    Turn int           `json:"turn"`
    Player int         `json:"player"`
    Action string      `json:"action"` // "move", "attack", "build", "end_turn"
    From *AxialCoord   `json:"from,omitempty"`
    To *AxialCoord     `json:"to,omitempty"`
    Timestamp time.Time `json:"timestamp"`
    Result interface{} `json:"result"`
    Valid bool         `json:"valid"`
}
```

#### Web Interface Integration
**Design**: Game mode enhancement to existing map editor
```typescript
// New components following existing BaseComponent pattern
class GameState extends BaseComponent {
    async createGameFromMap(mapData: any, playerCount: number): Promise<void>
    async selectUnit(q: number, r: number): Promise<UnitSelection>
    async moveUnit(from: Coord, to: Coord): Promise<MoveResult>
    async attackUnit(attacker: Coord, target: Coord): Promise<CombatResult>
    async endTurn(): Promise<TurnResult>
}

class GameController extends BaseComponent {
    // Orchestrates game flow using EventBus
    // Integrates with existing PhaserEditorComponent
    // Handles mode switching: Edit Mode ↔ Game Mode
}
```

### Integration with Existing Architecture

#### EventBus Communication Enhancement
**Extension**: Add game-specific events to existing EventBus
```typescript
// New game events
UNIT_SELECTED, MOVE_COMPLETED, ATTACK_EXECUTED, TURN_ENDED
GAME_STATE_CHANGED, PLAYER_SWITCHED, VICTORY_ACHIEVED

// Integration with existing Map events
TILES_CHANGED, UNITS_CHANGED, MAP_LOADED
```

#### Phaser Integration
**Enhancement**: Add game mode to existing PhaserEditorComponent
```typescript
class PhaserEditorComponent {
    private gameMode: boolean = false
    
    switchToGameMode(): void {
        this.gameMode = true
        this.disableEditing()
        this.enableUnitSelection()
    }
    
    private handleTileClick(q: number, r: number): void {
        if (this.gameMode) {
            this.handleUnitSelection(q, r)
        } else {
            this.handleTilePainting(q, r)
        }
    }
}
```

#### Component Lifecycle Integration
**Compatibility**: Game components follow existing lifecycle patterns
```typescript
// GameState component follows existing BaseComponent pattern
export class GameState extends BaseComponent implements ComponentLifecycle {
    public async bindToDOM(): Promise<void> { /* Initialize WASM */ }
    public async injectDependencies(deps: ComponentDependencies): Promise<void> { /* EventBus setup */ }
    public async activate(): Promise<void> { /* Subscribe to game events */ }
}
```

### Performance & Scalability Considerations

#### Rules Engine Optimization
- **Caching Strategy**: Cache frequently accessed unit/terrain data
- **Lazy Loading**: Load rule data on demand, not at startup
- **Batch Operations**: Validate multiple moves in single API call

#### WASM Performance
- **Minimal Data Transfer**: Only send essential data across WASM boundary
- **Batch Updates**: Group multiple game state changes
- **Memory Management**: Efficient handling of large game states

#### Web Interface Responsiveness
- **Progressive Updates**: Update UI incrementally during long operations
- **Background Processing**: Use Web Workers for complex calculations
- **State Synchronization**: Efficient diff-based updates via EventBus

### Testing Strategy

#### CLI-Driven Testing
- **Recorded Sessions**: Complete games recorded via CLI for regression testing
- **Rule Validation**: Compare with original WeeWar mechanics
- **Performance Benchmarks**: Large maps with many units

#### Integration Testing
- **WASM API Testing**: Browser console validation of all game APIs
- **Web UI Testing**: End-to-end gameplay through web interface
- **Cross-Platform**: Ensure CLI and web produce identical results

#### Automated Testing
- **Unit Tests**: Rules engine with weewar-data.json validation
- **Replay Tests**: Recorded sessions for regression detection
- **Performance Tests**: Response time and memory usage validation

**Previous Achievement (v7.0)**: Comprehensive UI framework completion and game foundation:

### EventBus Architecture ✅ COMPLETED
- **Lifecycle-Based Component System**: Template-scoped event binding for dynamic UI components
- **EventBus Communication**: Type-safe, loosely-coupled component interaction with source filtering
- **Defensive Programming**: Robust state management with graceful error handling and automatic recovery
- **Observer Pattern Integration**: Unified Map and PageState architecture for reactive updates

### Map Editor Polish ✅ COMPLETED  
- **Unit Toggle Behavior**: Same unit+player removes unit, different unit/player replaces unit
- **City Tile Player Ownership**: Fixed city terrain rendering with proper player colors and ownership
- **Reference Image Controls**: Complete scale and position controls with horizontal switch UI
- **Per-Tab Number Overlays**: N/C/U keys toggle overlays per tab with persistent state management
- **Map Details Layout**: Fixed-width 250px right panel with responsive map preview

### Backend Integration ✅ COMPLETED
- **Maps Delete Endpoint**: Complete DELETE /maps/{mapId} with proper error handling and redirects
- **Web Route Architecture**: Clean HTTP method handling with proper REST semantics
- **Service Layer Integration**: Full integration with existing MapsService and file storage
- **Frontend Error Resolution**: Fixed HTMX delete button integration with backend endpoints

### Reference Image System ✅ COMPLETED
- **Horizontal Switch UI**: Replaced dropdown with professional switch-style radio buttons
- **Scale Controls**: Fixed scale state corruption with proper property mapping (scaleX/scaleY)
- **Position Controls**: Complete X/Y position translation with input fields and +/- buttons
- **Mode Visibility**: Scale/position controls visible in both background and overlay modes
- **State Management**: Proper toggle state tracking with visual feedback and EventBus communication

### Technical Architecture ✅ COMPLETED
- **Pure Observer Pattern**: All map changes go through Map class with Phaser updates via EventBus
- **Template-Scoped Event Binding**: Dynamic UI components work properly in dockview containers
- **Event Delegation Pattern**: Robust button handling that works within layout systems
- **Auto-Tile Placement**: Units automatically place grass tiles when no terrain exists
- **Error Recovery**: Comprehensive error handling with user feedback and graceful degradation

**Previous Achievement**: Fixed keyboard input interference and created shared DOM utilities for better input handling in complex UI layouts.

**Foundation Achievement**: Successfully implemented breadth-first component lifecycle architecture with zero-boilerplate lifecycle support and full backward compatibility.
