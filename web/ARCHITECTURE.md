# Web Module Architecture (v8.0)

## Overview

The WeeWar web module implements a sophisticated **Lifecycle Managed Component (LCM)** architecture with **EventSubscriber patterns** for building complex interactive web applications. This architecture evolved from solving critical timing issues, race conditions, and component coordination challenges in large-scale TypeScript applications.

## Core Architectural Patterns

### 1. LCMComponent Lifecycle Management

The architecture replaces traditional constructor-based initialization with a **four-phase, breadth-first lifecycle** that eliminates race conditions and ensures proper component coordination.

#### Four-Phase Lifecycle
1. **Phase 1: `performLocalInit()`** - DOM setup, event subscriptions, child discovery
2. **Phase 2: `setupDependencies()`** - Dependency injection and validation  
3. **Phase 3: `activate()`** - Final coordination when all dependencies ready
4. **Phase 4: `deactivate()`** - Cleanup and resource release

#### Breadth-First Initialization
```
Level 0: [Page Components]                    ← All complete Phase 1 before any start Phase 2
Level 1: [WorldViewer, GameState, Panels]     ← All complete Phase 1 before any start Phase 2  
Level 2: [PhaserEditorComponent, SubPanels]   ← All complete Phase 1 before any start Phase 2
```

**Synchronization Barriers**: No component proceeds to next phase until ALL components in ALL levels complete current phase.

#### Benefits
- **Eliminates Race Conditions**: Synchronization barriers prevent timing dependencies
- **Order Independence**: Components can be created in any sequence within a level
- **Error Isolation**: Component failures don't cascade through barriers
- **Async Safety**: Proper async initialization without blocking other components

### 2. EventSubscriber Interface Pattern

**Moved away from callback-based subscriptions** to a cleaner, type-safe interface pattern:

#### Old Pattern (Deprecated)
```typescript
// Callback-based approach - error prone
eventBus.subscribe('event-type', (data) => { /* handle */ });
```

#### New Pattern (Current)
```typescript
// Interface-based subscription - type safe
export interface EventSubscriber {
    handleBusEvent(eventType: string, data: any, subject: any, emitter: any): void;
}

// Components implement EventSubscriber
public handleBusEvent(eventType: string, data: any, subject: any, emitter: any): void {
    switch(eventType) {
        case WorldEventTypes.WORLD_VIEWER_READY:
            this.handleWorldViewerReady(data);
            break;
        case EditorEventTypes.PHASER_READY:
            this.handlePhaserReady(data);
            break;
        default:
            super.handleBusEvent(eventType, data, subject, emitter);
    }
}

// Subscribe using the interface  
this.addSubscription(eventType, target, this); // 'this' implements EventSubscriber
```

#### EventBus Features
- **At-most-once guarantees**: Prevents duplicate subscriptions
- **Error isolation**: Handler failures don't stop other handlers
- **Source exclusion**: Events not sent back to emitting component
- **Debug logging**: Comprehensive event flow tracking

### 3. World-Centric Data Architecture

The **World class** serves as the **single source of truth** for all game data:

#### Observer Pattern Implementation
```typescript
// World emits events when data changes
world.setTileAt(q, r, terrainType, playerId);
// → Emits WorldEventType.TILES_CHANGED with change details
// → All subscribed components receive event and update themselves

// Components react to World changes
handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
    switch(eventType) {
        case WorldEventType.TILES_CHANGED:
            this.updatePhaserDisplay(data.changes);
            break;
        case WorldEventType.UNITS_CHANGED:
            this.updateUnitDisplay(data.changes);
            break;
    }
}
```

#### Benefits
- **Data Consistency**: Single source eliminates scattered data copies
- **Automatic Synchronization**: All components stay in sync via events
- **Batch Operations**: Efficient bulk updates with consolidated events
- **Persistence Layer**: World handles save/load operations

### 4. Component Hierarchy and Orchestration

#### BasePage Pattern
All pages extend `BasePage` which provides:
- **Common UI Components**: Theme manager, modals, toast notifications
- **LCMComponent Interface**: Implements lifecycle management
- **EventBus Integration**: Centralized event handling
- **Utility Methods**: Element creation, error handling

#### Page Orchestration Pattern
```typescript
export class GameViewerPage extends BasePage implements LCMComponent {
    // Phase 1: Initialize DOM and discover children
    performLocalInit(): LCMComponent[] {
        this.loadGameConfiguration();
        this.subscribeToGameStateEvents();  // FIRST: Subscribe before creating children
        this.createWorldViewerComponent();  // THEN: Create children (they can emit immediately)
        return [this.worldViewer, this.gameState, this.terrainStatsPanel];
    }
    
    // Phase 3: Activate when all dependencies ready
    async activate(): Promise<void> {
        this.bindGameSpecificEvents();
        // Ready to handle user interactions
    }
    
    // Handle events from EventBus
    public handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
        switch(eventType) {
            case WorldEventTypes.WORLD_VIEWER_READY:
                await this.worldViewer.loadWorld(this.world);
                break;
        }
    }
}
```

#### LifecycleController Orchestration
```typescript
// Initialize with LifecycleController for proper coordination
document.addEventListener('DOMContentLoaded', async () => {
    const page = new GameViewerPage("GameViewerPage");
    const lifecycleController = new LifecycleController(page.eventBus, {
        enableDebugLogging: true,
        phaseTimeoutMs: 15000,
        continueOnError: false
    });
    await lifecycleController.initializeFromRoot(page);
});
```

### 5. Phaser Integration Patterns (Unified Architecture)

After extensive refactoring in v8.0, we consolidated Phaser integration around a **unified scene-based architecture** that eliminates unnecessary wrapper layers while maintaining proper container management.

#### PhaserWorldScene - Core Phaser Integration
```typescript
export class PhaserWorldScene extends Phaser.Scene implements LCMComponent {
    constructor(containerElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('phaser-world-scene');
        this.containerElement = containerElement;  // Direct container reference
        this.eventBus = eventBus;
        this.debugMode = debugMode;
    }
    
    // LCMComponent lifecycle
    async performLocalInit(): Promise<LCMComponent[]> {
        // Validate container and setup
        if (!this.containerElement) {
            throw new Error('PhaserWorldScene: Container element is required');
        }
        await this.initializePhaser();
        return []; // Leaf component
    }
    
    // Phaser configuration for proper container targeting
    private async initializePhaser(): Promise<void> {
        const config: Phaser.Types.Core.GameConfig = {
            type: Phaser.AUTO,
            parent: this.containerElement.id || this.containerElement,  // Proper parent targeting
            width: width,
            height: height,
            backgroundColor: '#2c3e50',
            scene: this,
            scale: {
                mode: Phaser.Scale.FIT,           // Fixed: was RESIZE (caused infinite growth)
                width: width,
                height: height,
                autoCenter: Phaser.Scale.CENTER_BOTH
            }
        };
        this.phaserGame = new Phaser.Game(config);
    }
}
```

#### PhaserEditorScene - Editor Extension
```typescript
export class PhaserEditorScene extends PhaserWorldScene {
    // Editor-specific functionality built on PhaserWorldScene foundation
    // Includes: reference images, terrain painting, brush tools, world generation
    
    constructor(containerElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super(containerElement, eventBus, debugMode);
    }
    
    // Higher-level API methods for editor functionality
    public async setTerrain(terrain: number): Promise<void> { /* painting tools */ }
    public async setTilesData(tiles: Array<Tile>): Promise<void> { /* world loading */ }
    public fillAllTerrain(terrain: number, color: number = 0): void { /* world generation */ }
}
```

#### Container Management - Critical Fix
**Problem Identified**: Canvas being created as sibling instead of child of intended container

**Root Cause**: Pages were passing wrong container elements to PhaserWorldScene:
```typescript
// WRONG: Passing outer wrapper container
const worldViewerRoot = this.ensureElement('[data-component="world-viewer"]', 'world-viewer-root');
this.worldScene = new PhaserWorldScene(worldViewerRoot, this.eventBus, true);

// CORRECT: Passing actual Phaser canvas container  
const phaserContainer = this.ensureElement('#phaser-viewer-container', 'phaser-viewer-container');
this.worldScene = new PhaserWorldScene(phaserContainer, this.eventBus, true);
```

**Container Hierarchy**:
```html
<div data-component="world-viewer">     <!-- Outer component wrapper -->
  <div class="p-4 border-b">...</div>   <!-- Component header -->
  <div id="phaser-viewer-container">    <!-- Actual Phaser container (CORRECT TARGET) -->
    <!-- Canvas renders here -->
  </div>
</div>
```

#### Wrapper Elimination (v8.0 Major Refactor)
Eliminated unnecessary wrapper layers in editor architecture:

**Before** (v7.0):
```
PhaserEditorComponent → PhaserWorldEditor → PhaserEditorScene
                     (wrapper)        (wrapper)    (actual scene)
```

**After** (v8.0):
```
PhaserEditorComponent → PhaserEditorScene
                     (component)    (unified scene)
```

**Benefits**:
- **Reduced Complexity**: Fewer abstraction layers to debug
- **Better Performance**: Direct method calls instead of wrapper forwarding  
- **Cleaner API**: Consistent interface between viewer and editor
- **Easier Maintenance**: Single point of truth for Phaser functionality

## Key Architectural Decisions

### 1. Why Breadth-First Initialization?

**Problem**: Traditional depth-first initialization creates race conditions:
```
Parent starts → Child1 starts → Grandchild starts & finishes → Child1 finishes → Child2 starts...
```
If Child2 depends on Grandchild being ready, timing issues occur.

**Solution**: Breadth-first with synchronization barriers:
```
Level 0: All parents start and finish Phase 1
Level 1: All children start and finish Phase 1  
Level 2: All grandchildren start and finish Phase 1
Then: All components start Phase 2...
```

### 2. Why EventSubscriber Interface?

**Problem**: Callback-based subscriptions were error-prone:
- Duplicate subscriptions
- Memory leaks from uncleaned callbacks  
- Type safety issues
- Error propagation between handlers

**Solution**: Interface-based pattern:
- Type-safe event handling
- Automatic subscription management
- Error isolation between components
- Clear event flow debugging

### 3. Why World-Centric Architecture?

**Problem**: Scattered data management:
- Components maintained separate data copies
- Manual synchronization between UI elements
- Inconsistent state across application
- Complex coordination logic

**Solution**: Single source of truth:
- World class manages all game data
- Components react to World events
- Automatic synchronization via Observer pattern
- Consistent state guarantees

### 4. Why Four-Phase Lifecycle?

**Traditional Constructor Problems**:
- Complex initialization logic mixed with object creation
- Race conditions between dependent components
- Difficult to handle async operations safely
- Error propagation during construction

**Four-Phase Benefits**:
- **Phase 1**: Simple DOM setup and child discovery
- **Phase 2**: Clean dependency injection point
- **Phase 3**: Safe coordination when all dependencies ready
- **Phase 4**: Proper cleanup and resource management

## Component Implementation Patterns

### Basic Component Pattern
```typescript
export class MyComponent extends BaseComponent implements LCMComponent {
    private childComponents: LCMComponent[] = [];
    
    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode = false) {
        super('my-component', rootElement, eventBus, debugMode);
    }
    
    // Phase 1: DOM setup and child discovery
    public performLocalInit(): LCMComponent[] {
        // 1. CRITICAL: Subscribe to events BEFORE creating children
        this.addSubscription('child-ready', this);
        this.addSubscription('data-loaded', this);
        
        // 2. Setup DOM elements
        this.setupDOMElements();
        
        // 3. Create child components (they can emit events immediately)
        this.childComponents = this.createChildComponents();
        
        // 4. Return children for lifecycle management
        return this.childComponents;
    }
    
    // Phase 2: Dependency injection (optional)
    public setupDependencies(): void {
        // Validate required dependencies are available
        // Set up references to other components
    }
    
    // Phase 3: Final activation
    public async activate(): Promise<void> {
        // Bind DOM event handlers
        this.bindDOMEvents();
        
        // Any coordination with other components
        this.emit('component-ready', { componentId: this.componentId }, this, this);
    }
    
    // Phase 4: Cleanup
    public deactivate(): void {
        this.destroy();
    }
    
    // Handle EventBus events
    public handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
        switch(eventType) {
            case 'child-ready':
                this.handleChildReady(data);
                break;
            case 'data-loaded':
                this.handleDataLoaded(data);
                break;
            default:
                super.handleBusEvent(eventType, data, target, emitter);
        }
    }
    
    // Component-specific cleanup
    protected destroyComponent(): void {
        // Clean up resources, timers, external library instances
    }
}
```

### Page Component Pattern
```typescript
export class MyPage extends BasePage {
    private world: World;
    private mainComponent: MyComponent;
    
    // Phase 1: Create World and child components
    performLocalInit(): LCMComponent[] {
        // Load World data first
        this.world = this.loadWorldFromDOM();
        
        // Subscribe to events before creating children
        this.addSubscription('component-ready', this);
        
        // Create main component
        const container = this.ensureElement('[data-component="main"]', 'main-container');
        this.mainComponent = new MyComponent(container, this.eventBus, true);
        
        return [this.mainComponent];
    }
    
    // Phase 3: Final page activation
    async activate(): Promise<void> {
        // Bind page-specific events
        this.bindPageEvents();
        
        // Load World into components
        if (this.mainComponent && this.world) {
            await this.mainComponent.loadWorld(this.world);
        }
    }
}

// Initialize with LifecycleController
document.addEventListener('DOMContentLoaded', async () => {
    const page = new MyPage('my-page');
    const controller = new LifecycleController(page.eventBus);
    await controller.initializeFromRoot(page);
});
```

## Current System Components

### Core Library (`web/lib/`)
- **LCMComponent.ts**: Lifecycle management interface
- **LifecycleController.ts**: Breadth-first orchestration
- **EventBus.ts**: EventSubscriber pattern implementation
- **Component.ts**: BaseComponent with lifecycle integration
- **BasePage.ts**: Page base class with common UI functionality

### Pages (`web/src/`)
- **StartGamePage.ts**: Game configuration and world selection
- **GameViewerPage.ts**: Interactive game play interface
- **WorldEditorPage.ts**: Professional world editing tools  
- **WorldViewerPage.ts**: Read-only world visualization

### Components (`web/src/`)
- **WorldViewer.ts**: Phaser-based world visualization
- **PhaserEditorComponent.ts**: Phaser-based world editor
- **GameState.ts**: WASM game engine integration
- **World.ts**: Single source of truth for world data
- **Various Panels**: UI components for stats, tools, etc.

## Data Flow Patterns

### 1. World Data Loading
```
Page loads → Deserialize World from DOM → Pass World to components → Components load into Phaser
```

### 2. User Interaction
```  
User clicks → Component handles → Updates World → World emits events → All components sync
```

### 3. Component Communication
```
Component A → EventBus.emit() → EventBus → Component B.handleBusEvent() → Component B reacts
```

### 4. Phaser Integration
```
World changes → World emits event → PhaserComponent receives → Updates Phaser scene → Visual update
```

## Benefits of This Architecture

### For Developers
- **Predictable Initialization**: Clear phases eliminate timing issues
- **Type Safety**: EventSubscriber interface prevents runtime errors
- **Easy Debugging**: Comprehensive logging and clear event flow
- **Component Isolation**: Failures don't cascade between components
- **Reusable Components**: Clear interfaces enable composition

### For Users  
- **Faster Loading**: Parallel initialization within phases
- **Reliable Behavior**: Synchronization barriers prevent race conditions
- **Better Error Handling**: Graceful degradation when components fail
- **Consistent Experience**: Predictable component behavior

### For Maintenance
- **Clear Separation**: World data vs UI components vs Phaser rendering
- **Event-Driven**: Loose coupling between components
- **Single Source of Truth**: World class manages all data consistency
- **Lifecycle Management**: Proper resource cleanup and error handling

## Migration from Previous Versions

### Key Changes from v7.0 to v8.0
1. **Wrapper Elimination**: Removed PhaserWorldEditor and PhaserPanel wrapper classes
2. **Unified Phaser Architecture**: PhaserWorldScene as base, PhaserEditorScene as extension
3. **Container Management Fix**: Pages now target correct Phaser containers instead of outer wrappers
4. **Scale Mode Fix**: Changed from RESIZE to FIT mode to prevent infinite canvas growth
5. **Method Signature Alignment**: Fixed TypeScript compatibility between components and scenes
6. **Constructor Simplification**: Streamlined LCMComponent initialization in Phaser scenes

### Breaking Changes
- **Event Subscriptions**: Must implement `handleBusEvent()` instead of callbacks
- **Initialization**: Components must implement `performLocalInit()` instead of constructor logic  
- **World Loading**: Pass World objects, not raw data, to Phaser components
- **Lifecycle Management**: Use LifecycleController for page initialization

## Future Architecture Directions

### Planned Enhancements
1. **Service Container**: Dependency registry for complex dependencies
2. **Component Registry**: Dynamic component loading and discovery
3. **Event Replay**: Debug capabilities for complex event flows
4. **Performance Metrics**: Component initialization and rendering performance
5. **Hot Reload**: Development-time component updates without page refresh

### Architectural Principles to Maintain
1. **Breadth-First Lifecycle**: Synchronization barriers prevent race conditions
2. **EventSubscriber Pattern**: Type-safe, error-isolated event handling
3. **World-Centric Data**: Single source of truth for all game data
4. **Component Isolation**: Clear boundaries and responsibilities
5. **Async Safety**: Proper handling of async operations in lifecycle phases

This architecture represents a mature, battle-tested approach to building complex interactive web applications with multiple interdependent components, real-time data synchronization, and rich user interfaces.