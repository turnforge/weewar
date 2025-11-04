**Purpose:**

This folder contains the core client-side TypeScript logic for the webapp, managing UI state, user events, API interactions, and DOM manipulation using a modern component-based architecture with strict separation of concerns and event-driven communication.

**Core Architecture Components:**

## Modern Component System (New)

*   **`WorldViewer.ts`**: Phaser-based world visualization component with proper DOM scoping and event-driven initialization  
*   **`WorldStatsPanel.ts`**: Statistics display component with safe DOM selectors and event-driven updates
*   **`WorldViewerPage.ts`**: Orchestrator page following new architecture - handles data loading and component coordination only

## Key Architecture Principles

*   **Separation of Concerns**: Clear boundaries between layout, behavior, and communication responsibilities
*   **Event-Driven**: Components communicate through EventBus events, never direct method calls  
*   **DOM Isolation**: Components only access DOM within their assigned root elements
*   **Error Resilience**: Component failures are isolated and don't affect other components
*   **Timing Awareness**: Proper handling of initialization order, race conditions, and async operations
*   **WebGL Integration**: Specialized patterns for graphics libraries like Phaser with timing considerations

## Critical Timing Patterns Learned

*   **TypeScript Field Initializers**: Avoid explicit `= null` for constructor-set fields
*   **Event Subscription Order**: Subscribe to events BEFORE creating components that emit them
*   **WebGL Context Readiness**: Use small setTimeout for graphics library initialization completion
*   **State → Subscribe → Create**: Strict three-phase initialization order
*   **Async in Handlers**: EventBus stays synchronous, handlers use `.then()/.catch()` for async operations

## Integration Capabilities

*   **Phaser.js**: WebGL-based world rendering with proper timing handling
*   **HTMX**: Component hydration support for server-driven UI updates  
*   **Canvas/WebGL**: Specialized initialization patterns for graphics contexts
*   **Toast/Modal Systems**: User feedback and interaction patterns
*   **Theme Management**: Coordinated theming across component boundaries

## Recent Session Work (2025-01-24)

### Layer System Architecture Complete ✅
*   **Generic WorldViewer**: `WorldViewer<TScene>` with template parameter for proper typing
*   **GameViewer Specialization**: `GameViewer extends WorldViewer<PhaserGameScene>` with game-specific layer access
*   **Layer-Based Interaction**: Direct layer manipulation (`getSelectionHighlightLayer()`, `getMovementHighlightLayer()`, etc.)
*   **Editor Integration**: PhaserEditorComponent uses layer callbacks for painting logic
*   **Callback Architecture**: Click handling through BaseMapLayer callbacks with validation in components
*   **Brush Size Support**: Multi-tile painting with hex distance calculations in component layer

### Architecture Improvements ✅
*   **Scene Separation**: PhaserWorldScene for rendering, components for business logic
*   **Single Source of Truth**: World model updates trigger observer pattern for visual updates
*   **Type Safety**: Proper TypeScript generics eliminate casting and improve developer experience
*   **Clean Separation**: UI logic in components, rendering logic in scenes, interaction through layers

## Recent Session Work (2025-01-22)

### Interactive Game Viewer Foundation ✅
*   **GameViewerPage Architecture**: Complete interactive game interface with lifecycle controller integration
*   **External Orchestration Pattern**: LifecycleController with breadth-first component initialization eliminates race conditions
*   **LCMComponent Interface**: Multi-phase initialization (performLocalInit, setupDependencies, activate, deactivate)
*   **WASM Bridge Architecture**: GameState component with async loading and synchronous gameplay operations
*   **Synchronous UI Pattern**: Immediate UI feedback with notification events for coordination only

### Component Communication Architecture ✅  
*   **Event-Driven Coordination**: Components communicate via EventBus without tight coupling
*   **Source Filtering**: Components ignore events they originate to prevent feedback loops
*   **Error Isolation**: Component failures don't cascade through event system
*   **Debug Support**: Comprehensive logging and lifecycle event callbacks
*   **Notification Events**: System coordination (`game-created`, `unit-moved`, `turn-ended`) for logging, animations

### Previous Session Work (2025-01-20)

#### Component Architecture Cleanup ✅
*   **WorldEditorPage Streamlining**: Removed dead code and consolidated component management patterns
*   **Panel Integration Optimization**: Improved coordination between EditorToolsPanel, TileStatsPanel, and PhaserEditor
*   **Reference Management**: Cleaner component initialization and lifecycle patterns
*   **State Management Consolidation**: Reduced complexity in page-level state handling

### Recent Session Work (2025-11-04)

#### Page Variant Architecture Pattern ✅
*   **GameViewerPageBase Abstract Class**: Core game logic extracted to base class (WASM, presenter, panels, events, RPC methods)
*   **Layout-Specific Child Classes**: GameViewerPageDockView (flexible dockable layout) and GameViewerPageGrid (static CSS grid)
*   **Abstract Method Pattern**: Child classes implement layout-specific concerns (initializeLayout, createPanels, getGameSceneContainer)
*   **Zero Logic Duplication**: All game logic lives in base class, children only handle structural differences
*   **Flexible Timing Control**: Child classes control when game scene is created (early for Grid, late for DockView)
*   **Server-Side Template Rendering**: Each variant has its own HTML template with appropriate layout structure
*   **Planned Mobile Variant**: Foundation ready for GameViewerPageMobile with bottom sheet and context-aware panels

**Architecture Benefits:**
- Clean separation of game logic from layout concerns
- Easy to add new layout variants without code duplication
- Can serve different variants based on user agent/screen size
- Each variant optimized for its use case (performance, flexibility, touch)

### Recent Session Work (2025-11-03)

#### Responsive Bottom Sheet Implementation ✅
*   **Bottom Sheet Handler Pattern**: Reusable `initializeBottomSheet()` method pattern for mobile overlays
*   **Event Handling**: openSheet/closeSheet functions with backdrop, close button, and ESC key support
*   **Multiple Elements Support**: Using `querySelectorAll` for binding to duplicate buttons across desktop/mobile layouts
*   **Applied to Pages**: WorldViewerPage (`initializeBottomSheet` for stats), StartGamePage (`initializeConfigBottomSheet` for config)
*   **Consistent Architecture**: Same implementation pattern across pages for maintainability

**Bottom Sheet Pattern Implementation:**
```typescript
private initializeBottomSheet(): void {
    const fab = document.getElementById('fab-id');
    const overlay = document.getElementById('overlay-id');
    const panel = document.getElementById('panel-id');
    const backdrop = document.getElementById('backdrop-id');
    const closeButton = document.getElementById('close-id');

    if (!fab || !overlay || !panel || !backdrop || !closeButton) return;

    const openSheet = () => {
        overlay.classList.remove('hidden');
        overlay.offsetHeight; // Force reflow
        panel.classList.remove('translate-y-full');
    };

    const closeSheet = () => {
        panel.classList.add('translate-y-full');
        setTimeout(() => overlay.classList.add('hidden'), 300);
    };

    fab.addEventListener('click', openSheet);
    closeButton.addEventListener('click', closeSheet);
    backdrop.addEventListener('click', closeSheet);
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && !overlay.classList.contains('hidden')) {
            closeSheet();
        }
    });
}
```

**Key Fixes:**
- StartGamePage: Fixed mobile "Start Game" button by using `querySelectorAll` to bind all buttons (desktop + mobile)
- Validation: Updated `validateGameConfiguration()` to disable/enable all button instances across layouts

#### Listing Page TypeScript Classes ✅
Created TypeScript pages for listing pages that previously had no JavaScript:

**WorldListingPage.ts** - Worlds listing page
```typescript
class WorldListingPage {
    constructor() {
        ThemeManager.init();
        this.init();
    }

    private init(): void {
        SplashScreen.dismiss();
    }
}
```

**GameListingPage.ts** - Games listing page
```typescript
class GameListingPage {
    constructor() {
        ThemeManager.init();
        this.init();
    }

    private init(): void {
        SplashScreen.dismiss();
    }
}
```

**Purpose:**
- Initialize ThemeManager for dark/light mode support
- Dismiss splash screen on page load
- Consistent pattern with HomePage.ts

**Integration:**
- Added to webpack.config.js components array
- Generated bundles included in respective page templates
- Both pages now properly initialize JavaScript functionality

**Result:** Splash screens now properly dismissed on /worlds/ and /games/ listing pages
