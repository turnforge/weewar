# Web Module Summary

## Purpose
The web module provides a modern web interface for the LilBattle turn-based strategy game, featuring a professional world editor, readonly world viewer, and comprehensive game management system.

## Current Architecture (v8.0)

### LCMComponent Lifecycle Architecture with EventSubscriber Pattern
- **4-Phase Lifecycle Management**: performLocalInit() → setupDependencies() → activate() → deactivate()
- **Breadth-First Initialization**: LifecycleController orchestrates component coordination with synchronization barriers
- **EventSubscriber Interface**: Type-safe event handling via handleBusEvent() method, replacing callback-based subscriptions
- **World-Centric Data Management**: Enhanced World class serves as single source of truth with observer pattern
- **Race Condition Elimination**: Synchronization barriers prevent timing issues and component coordination problems

### Core Components

#### Frontend Components (`web/pages/`)
- **common/World.ts**: Enhanced with Observer pattern, batched events, and self-contained persistence
- **WorldEditorPage/index.ts**: Refactored to use World as single source of truth, implements WorldObserver
- **WorldEditorPage/PhaserEditorComponent.ts**: Phaser.js-based world editor with WebGL rendering

#### Frontend Library Components (`web/lib/`)
- **LCMComponent.ts**: Lifecycle Managed Component interface with 4-phase initialization
- **LifecycleController.ts**: Breadth-first component orchestration with synchronization barriers
- **EventBus.ts**: EventSubscriber pattern with type-safe event handling and error isolation
- **Component.ts**: BaseComponent with LCMComponent integration and DOM scoping
- **BasePage.ts**: Page base class implementing the LCMComponent pattern with common UI functionality

#### Backend Services (`pkg/services/`)
- **WorldsService**: gRPC service for world CRUD operations
- **File-based Storage**: Worlds stored in `./storage/worlds/<worldId>/` structure
- **Connect Bindings**: HTTP API integration with frontend

### Key Features

#### World Editor
- **Phaser.js Integration**: WebGL-accelerated rendering with professional UX
- **Coordinate Accuracy**: Pixel-perfect matching with Go backend implementation
- **Observer Pattern**: Real-time component synchronization via World events
- **Keyboard Shortcuts**: Comprehensive shortcut system for rapid world building
- **Professional Tools**: Terrain painting, unit placement, brush sizes, player management

#### Component Architecture
- **LCMComponent Lifecycle**: 4-phase initialization with breadth-first coordination via LifecycleController
- **EventSubscriber Pattern**: Type-safe event handling via handleBusEvent() interface method
- **World-Centric Data Flow**: Single source of truth with automatic component synchronization
- **Error Isolation**: Component failures don't cascade through synchronization barriers
- **Race Condition Prevention**: Synchronization barriers eliminate timing dependencies

### Recent Achievements (Session 2025-08-20)

#### World Event Architecture Refactoring & Input System Fixes (Complete)
- **common/PhaserWorldScene Event Integration**: Moved world synchronization logic to base scene class
  - All Phaser scenes now automatically sync with World changes through EventBus
  - Eliminated duplicate world event handling between editor and viewer scenes
  - Cleaner separation: WorldEditorPage/PhaserEditorComponent handles editor events, common/PhaserWorldScene handles world sync
- **Event Flow Optimization**: Clarified dual subscription pattern in GameViewerPage/GameState
  - common/World applies actual world changes and re-emits as specific world events
  - common/PhaserWorldScene base class handles TILES_CHANGED/UNITS_CHANGED/WORLD_LOADED/WORLD_CLEARED
- **Input System Refactoring**: Replaced manual drag detection with proper Phaser patterns
  - Separated tap detection (tile painting) from drag detection (camera panning)
  - Used time/distance thresholds for robust tap detection instead of manual mouse tracking
  - Eliminated camera panning during normal tile clicks
- **Critical Bug Fix**: Fixed tile clearing visual issue due to sprite key mismatch
  - setTile() was storing sprites with texture keys, removeTile() was looking up by coordinate keys
  - Unified sprite storage to use coordinate keys (q,r) for consistent lookup

#### Previous Session: Command Interface and Advanced E2E Testing (Complete)
- **Command Interface Implementation**: High-level game action API for testing and accessibility
  - Unified methods used by both UI interactions and programmatic testing
  - Structured ActionResult responses with success/failure reporting
  - Full game state queries and selection management
  - Exposed via window.gameViewerPage for e2e access
- **Advanced E2E Infrastructure**: Production-ready testing with persistent test worlds
  - GameActions class for high-level test operations (selectUnit, moveUnit, endTurn)
  - GameTestUtils for debugging and enhanced failure reporting  
  - Test world setup/cleanup scripts with reusable world management
  - Real production endpoint testing with minimal surgical API mocking
- **Consolidated Architecture**: Eliminated duplicate methods, single implementation for UI and tests

#### Previous Session Achievements (Session 2025-08-05)

#### Phaser Architecture Unification (v8.0)
- **Wrapper Elimination**: Removed PhaserWorldEditor and PhaserPanel unnecessary wrapper classes
- **Unified Scene Architecture**: common/PhaserWorldScene as base class, WorldEditorPage/PhaserEditorScene as editor extension
- **Container Management Fix**: Resolved canvas placement issues - canvas now properly renders inside target containers
- **Scale Mode Optimization**: Fixed infinite canvas growth by switching from RESIZE to FIT scale mode
- **Method Signature Alignment**: Resolved TypeScript compatibility issues between components and scenes
- **Constructor Cleanup**: Streamlined LCMComponent initialization in Phaser scene classes

#### Critical Bug Fixes
- **Canvas Container Issue**: Fixed canvas being created as sibling instead of child of intended container
- **Root Cause Analysis**: Pages were passing outer wrapper containers instead of actual Phaser containers
- **Proper Container Targeting**: Updated WorldViewerPage to target `#phaser-viewer-container` directly
- **Phaser Configuration**: Improved parent element targeting with `containerElement.id || containerElement`
- **Scale Mode Stability**: Changed from problematic RESIZE mode to stable FIT mode with autoCenter

#### Architecture Simplification Benefits
- **Reduced Complexity**: Eliminated multiple abstraction layers (PhaserEditorComponent → PhaserEditorScene direct)
- **Better Performance**: Direct method calls instead of wrapper forwarding through eliminated layers
- **Cleaner API**: Consistent interface patterns between viewer and editor components
- **Easier Debugging**: Fewer layers to trace through when investigating issues
- **Maintainability**: Single point of truth for Phaser functionality instead of scattered wrappers

#### Component Architecture Cleanup and Technical Debt Reduction (v5.2)
- Comprehensive cleanup of WorldEditorPage with dead code elimination
- Component reference streamlining and initialization pattern improvements
- Panel integration optimization between EditorToolsPanel, TileStatsPanel, and PhaserEditor
- Import cleanup and removal of unnecessary dependencies throughout components
- Method consolidation and code organization improvements for better maintainability
- State management simplification and complexity reduction

#### Previous Session: Unified World Architecture Implementation
- Enhanced common/World class with comprehensive Observer pattern support
- Implemented WorldObserver interface with type-safe event handling
- Added batched event system for performance optimization
- Created self-contained persistence methods in common/World class
- Refactored WorldEditorPage/index.ts to use World as single source of truth
- Removed redundant state properties and manual change tracking
- Fixed all compilation errors and achieved clean build

#### Previous Session: Component State Management Architecture
- Created WorldEditorPage/PageState class for centralized page-level state management
- Established proper component encapsulation with DOM ownership principles
- Refactored WorldEditorPage/ToolsPanel to be state generator and exclusive DOM owner
- Eliminated cross-component DOM manipulation violations
- Implemented clean state flow: User clicks → Component updates state → State emits events → Components observe
- Separated state levels: Page-level (tools), Application-level (theme), Component-level (local UI)

#### Architecture Benefits
- **LCMComponent Lifecycle**: 4-phase initialization eliminates race conditions and timing dependencies
- **EventSubscriber Pattern**: Type-safe, error-isolated event handling across all components
- **Breadth-First Coordination**: LifecycleController prevents component initialization conflicts
- **World-Centric Architecture**: Single source of truth with automatic component synchronization
- **Error Isolation**: Component failures contained through synchronization barriers
- **Consistent APIs**: All Phaser components use same loadWorld(world) pattern
- **Type Safety**: Interface-based event handling prevents runtime errors
- **Component Boundaries**: Proper encapsulation with clear lifecycle phases
- **Async Safety**: Proper async initialization without blocking other components
- **Maintainability**: Clear separation of phases makes debugging and extension easier
- **Code Organization**: Consistent patterns across all pages and components
- **Performance**: Batched events and efficient component coordination

### Technical Specifications

#### Coordinate System
- **Backend**: Cube coordinates (Q/R) with proper hex mathematics
- **Frontend**: Matches backend exactly with tileWidth=64, tileHeight=64, yIncrement=48
- **Conversion**: Row/col intermediate step using odd-row offset layout
- **Accuracy**: Pixel-perfect coordinate worldping between frontend and backend

#### Rendering Pipeline
- **Phaser.js**: WebGL-accelerated rendering engine
- **Dynamic Grid**: Infinite grid system rendering only visible hexes
- **Professional Interaction**: Paint-on-release, drag-to-pan, modifier key painting
- **Asset Integration**: Direct static URLs for tile/unit graphics

#### Data Persistence
- **World Class**: Handles own save/load operations with server API calls
- **Server Format**: Compatible with protobuf definitions for CreateWorld/UpdateWorld
- **Client Loading**: Supports both server data and HTML element parsing
- **Change Tracking**: Automatic via Observer pattern, eliminates manual marking

### Development Guidelines

#### LCMComponent Lifecycle Pattern
- Implement all 4 phases: performLocalInit() → setupDependencies() → activate() → deactivate()
- Subscribe to events FIRST in performLocalInit() before creating child components
- Use LifecycleController for page initialization with proper configuration
- Return child components from performLocalInit() for lifecycle management

#### EventSubscriber Pattern Usage
- Implement EventSubscriber interface with handleBusEvent() method
- Use addSubscription() instead of callback-based subscriptions
- Handle events via switch statement in handleBusEvent()
- Call super.handleBusEvent() for unhandled events

#### Component Development
- Extend BaseComponent and implement LCMComponent interface
- Use EventBus with EventSubscriber pattern for inter-component communication
- Implement proper cleanup in destroyComponent() and deactivate()
- Scope DOM queries to component containers with findElement()

#### World Operations
- Use common/World class methods for all tile/unit operations (single source of truth)
- Pass World objects (not raw data) to Phaser components via loadWorld() method
- Subscribe to World events for automatic UI synchronization
- Let common/World handle persistence and change tracking automatically

### Next Development Priorities

#### Component Integration Completion
- Update WorldEditorPage/PhaserEditorComponent to subscribe to World events
- Update WorldEditorPage/TileStatsPanel to read from World instead of Phaser
- Remove redundant getTilesData/setTilesData methods
- Test complete component synchronization via World events

#### Performance Optimization
- Performance testing with large worlds
- Memory usage optimization for Observer pattern
- Event debouncing for rapid interactions
- Benchmarking unified vs scattered approach

### Technology Stack
- **Backend**: Go with gRPC services and Connect bindings
- **Frontend**: TypeScript with Phaser.js, EventBus, and Observer patterns
- **Styling**: Tailwind CSS with dark/light theme support
- **Build**: Webpack with hot reload development
- **Layout**: DockView for professional panel management

### Recent Achievements (Session 2025-09-05)

#### SVG Asset Loading System (Complete)
- **AssetProvider Architecture**: Created interface-based system for swappable asset loading strategies
- **Theme Support**: Assets organized under `web/assets/themes/<themeName>/` with mapping.json configuration
- **SVG Template Processing**: Dynamic player color replacement using template variables in SVGs
- **Phaser Integration**: Fixed async loading issues using Phaser's JSON loader for mapping files
- **Memory Optimization**: 160x160 rasterization for rendering 1000+ tiles efficiently
- **Provider Independence**: common/PhaserWorldScene agnostic to specific provider implementations
- **Provider-Specific Display Sizing**: SVG and PNG providers use different display dimensions for proper hex overlap
- **Unit Label Repositioning**: Moved health/movement labels below units with smaller font to prevent row overlap
- **Show Health Toggle**: Added checkbox control in editor toolbar to toggle unit health/movement display visibility

### Recent Achievements (Session 2025-09-08)

#### Damage Distribution Panel Refactoring (Complete)
- **Component Extraction**: Separated damage distribution visualization from UnitStatsPanel into dedicated GameViewerPage/DamageDistributionPanel
- **Independent Dockview Panel**: Damage distribution now appears as its own dockable panel in GameViewerPage
- **Code Organization**: Removed createDamageHistogram() and generateUnitCombatTable() from GameViewerPage/UnitStatsPanel
- **Template Separation**: Moved combat table templates to DamageDistributionPanel.html
- **Event Integration**: Both panels receive unit selection events independently for synchronization
- **UI Flexibility**: Users can now show/hide damage distribution independently from unit stats

### Recent Achievements (Session 2025-10-30)

#### Animation Framework Implementation (Complete)
- **Presenter-Driven Architecture**: Animations orchestrated by presenter, scene remains a dumb renderer
- **Promise-Based API**: All animations return Promises for easy sequencing and chaining
- **Smart Batching**: Splash damage explosions play simultaneously, sequential chains for causality
- **Configurable Timing**: Single config file (common/animations/AnimationConfig.ts) controls all animation speeds with instant mode support
- **Effect Classes**: Modular, self-contained effect implementations in common/animations/effects/ (ProjectileEffect, ExplosionEffect, HealBubblesEffect, CaptureEffect)
- **Scene API Enhancements**: Enhanced common/PhaserWorldScene with moveUnit(), showAttackEffect(), showHealEffect(), showCaptureEffect()
- **Unit Lifecycle Animations**: setUnit with flash/appear options, removeUnit with fade-out animation
- **Particle System**: Runtime-generated particle texture with Phaser particle emitters for effects
- **Path Animation**: Units smoothly slide along hex paths instead of teleporting
- **Attack Sequences**: Complete attack animations with attacker flash, projectile arc, and impact explosions

### Recent Achievements (Session 2025-11-03)

#### Responsive Layout System with Bottom Sheets (Complete)
- **Mobile-First Design**: Full-screen canvas with overlay panels on mobile devices
- **Bottom Sheet Pattern**: Consistent slide-up overlay UI with backdrop, handle bar, and smooth transitions
- **FAB (Floating Action Button)**: Icon-based triggers for accessing secondary panels on mobile
- **Responsive Header Buttons**: Desktop buttons collapse into three-dot dropdown menu on mobile
- **Media Query Control**: Explicit CSS media queries (768px, 1024px breakpoints) with display controls
- **Applied to Multiple Pages**: WorldViewerPage (stats panel) and StartGamePage (config panel)
- **Consistent Implementation**: Reusable patterns for openSheet/closeSheet handlers with ESC key support
- **Layout Preservation**: Desktop two-column layouts unchanged, mobile gets optimized single-column with overlays

#### Configurable World Listing System (Complete)
- **WorldFilterPanel Component**: Reusable search, filter, and sort controls with optional HTMX support
- **WorldGrid Component**: Responsive grid view (1-4 columns) with card-based world display
- **WorldList Component**: Unified component supporting both table and grid view modes
- **View Modes**: Toggle between table (traditional list) and grid (card gallery) views with query params
- **Action Modes**: "manage" mode (edit/delete/start game) vs "select" mode (large Play buttons)
- **SelectWorldPage**: Dedicated world selection page for game creation workflow (/worlds/select)
- **Pagination**: Full pagination support in both table and grid views
- **Smart Defaults**: Grid view for selection, table view for management
- **User Flow Optimization**: /games/new without worldId auto-redirects to /worlds/select

#### Splash Screen System Completion (Complete)
- **WorldListingPage.ts**: Created TypeScript page class for worlds listing with splash screen dismissal
- **GameListingPage.ts**: Created TypeScript page class for games listing with splash screen dismissal
- **Webpack Integration**: Added both pages to build configuration for bundling
- **Consistent Pattern**: All pages now properly dismiss splash screens on load

#### StartGamePage Robustness (Complete)
- **Graceful Missing WorldId**: No error toast when accessing without worldId parameter
- **Auto-Redirect**: Automatically redirects to /worlds/select for world selection
- **Conditional Loading**: Only loads world data and Phaser components when worldId present
- **Safe Activation**: Component initialization checks prevent crashes

### Recent Achievements (Session 2025-01-05)

#### Reference Image Layer System for World Editor (Complete)
- **Layer System Integration**: Integrated WorldEditorPage/ReferenceImageLayer into WorldEditorPage/PhaserEditorScene for proper overlay/background handling
- **Independent Drag Handling**: Overlay mode allows dragging reference image without moving tiles or camera
- **Independent Scroll Handling**: Overlay mode allows scaling reference image with mouse wheel without zooming camera
- **Event-Driven UI Sync**: Position and scale changes emit events to keep UI controls in sync during drag/scroll
- **Circular Reference Prevention**: Value comparison guards prevent infinite update loops between layer and UI
- **LayerManager Extensions**: Added processClick(), processDrag(), processScroll() for unified event routing
- **Background vs Overlay Modes**: Background (depth -1) moves with world, Overlay (depth 1000) interactive and independent
- **World Coordinate Persistence**: Reference image stays in world coordinates when switching modes (no position jumping)

**Architecture**:
- WorldEditorPage/ReferenceImageLayer implements full Layer interface with hitTest(), handleClick(), handleDrag(), handleScroll()
- common/PhaserWorldScene routes pointer events through LayerManager before camera operations
- Layers can block camera pan/zoom by returning true from event handlers
- Scene events bridge layer changes to EventBus for UI synchronization
- Guard conditions in WorldEditorPage/ReferenceImagePanel prevent circular updates from programmatic input.value changes

### Recent Achievements (Session 2025-01-06)

#### UI Fixes and Panel Scrolling (Complete)
- **WorldListingPage Navigation Fix**: Disabled HTMX in-place updates for search filters to fix splash screen issue
  - Changed WorldFilterPanel HtmxEnabled parameter from true to false
  - Search and sort controls now perform full page redirects instead of partial updates
  - Added JavaScript handlers (handleSearchChange, handleSortChange) for non-HTMX navigation
  - Splash screen now properly dismisses after view switching
- **Panel Scrolling Fix**: Fixed vertical scrolling in WorldEditorPage dockview panels
  - Added `style="height: 100%"` to all panel wrapper divs in WorldEditorPage.html
  - Fixed: ToolsPanel, GameConfigPanel, ReferenceImagePanel, WorldStatsPanel, TileStatsPanel, TerrainStatsPanel
  - Added `overflow-y-auto` to WorldStatsPanel and TileStatsPanel content
  - Panels now properly constrain to parent height and scroll when content exceeds available space

**Architecture**:
- Panel wrappers must have `height: 100%` to properly constrain children in dockview
- Panel content divs need `h-full max-h-full overflow-y-auto` for proper scrolling
- Following same pattern as GameViewerPage panels (TurnOptionsPanel, etc.)

### Recent Achievements (Session 2025-11-06)

#### Multi-Click Shape Tool System (Complete)
- **Extensible Shape Tool Architecture**: Created WorldEditorPage/tools/ShapeTool interface for multi-click shape drawing workflow
  - WorldEditorPage/tools/RectangleTool: First click sets corner, mouse move shows preview, second click completes
  - WorldEditorPage/tools/CircleTool: First click sets center, second click sets radius (2 clicks total)
  - WorldEditorPage/tools/OvalTool: First click sets center, second click sets radiusX, third click sets radiusY (3 clicks total, axis-aligned)
  - WorldEditorPage/tools/LineTool: Multi-click path tool - N clicks to add vertices, Enter to finish, Escape to cancel
  - Replaced fatiguing drag-based system with ergonomic multi-click approach
  - Camera panning disabled during shape mode to prevent accidental interruptions
  - Escape key cancels current shape but stays in shape mode
- **World Helper Methods**: Added reusable shape generation methods to common/World.ts
  - World.circleFrom(): Generate circle tiles using hex distance formula
  - World.ovalFrom(): Generate axis-aligned ellipse tiles in row/col space
  - World.lineFrom(): Generate line/path tiles using Bresenham algorithm
  - hexDistance() utility in common/hexUtils.ts for hex coordinate distance calculation
- **Fill/Outline Toggle**: Added UI checkbox for switching between filled and outline shapes
  - Toggle appears for Rectangle, Circle, and Oval modes
  - Hidden for Line mode (lines always stroke only)
  - Updates shape preview in real-time when toggled
  - Integrated into WorldEditorPageToolbar template
- **Event Handler Refactoring**: Replaced drag-based rectangle logic with click-based system
  - pointerdown: Collects shape anchor points
  - pointermove: Shows live preview without requiring button hold
  - keyboard: Escape to cancel current shape, Enter for shapes requiring confirmation (line/path)
- **Unified Shape API**: Single setShapeMode(shapeType) method instead of separate methods per shape
  - Type-safe string literals: 'rectangle' | 'circle' | 'oval' | 'line' | null
  - Cleaner UI handler with shape mode mapping
- **UI Organization**: Created "Shapes" optgroup in brush selector dropdown
  - Rectangle (2 clicks)
  - Circle (2 clicks)
  - Oval (3 clicks)
  - Line/Path (N clicks, Enter to finish)

**Architecture**:
- WorldEditorPage/tools/ShapeTool interface defines contract for all shape tools
- Tools maintain their own state (anchor points, fill mode) and provide preview/result tiles
- WorldEditorPage/PhaserEditorScene delegates to currentShapeTool for all shape operations with unified setShapeMode()
- common/HexHighlightLayer (ShapeHighlightLayer) renders preview independently of tool type
- common/World methods use common/hexUtils (hexDistance, hexToRowCol, rowColToHex) for consistency
- Oval is axis-aligned in row/col space (simpler math, easier to reason about)
- Line uses Bresenham-style interpolation in row/col space

**Bug Fixes**:
- Fixed: First click in shape mode no longer sets tile on underlying layer (handleTap override)
- Fixed: Escape cancels current shape but stays in shape mode (cancelCurrentShape vs exitShapeMode)

### Recent Achievements (Session 2025-11-25)

#### WorldEditorPresenter Architecture Refactoring (Complete)
- **Presenter Pattern Implementation**: Created WorldEditorPage/WorldEditorPresenter to manage all UI state and component orchestration
  - Single source of truth for tool state (selectedTerrain, selectedUnit, selectedPlayer, brushMode, brushSize, placementMode)
  - Single source of truth for visual state (showGrid, showCoordinates, showHealth)
  - Direct presenter method calls replace EventBus for UI state changes
- **PageState Elimination**: Deleted WorldEditorPage/PageState.ts, merged all functionality into WorldEditorPresenter
- **Event Cleanup**: Removed obsolete event types from common/events.ts
  - Removed: TOOL_STATE_CHANGED, VISUAL_STATE_CHANGED, WORKFLOW_STATE_CHANGED, PAGE_STATE_CHANGED
  - Removed: GRID_SET_VISIBILITY, COORDINATES_SET_VISIBILITY, HEALTH_SET_VISIBILITY
  - Removed: GridSetVisibilityPayload, CoordinatesSetVisibilityPayload, HealthSetVisibilityPayload, PageStateChangedPayload
- **Console Panel Removal**: Removed unused console panel code from WorldEditorPage
  - Removed editorOutput property and initialization
  - Removed console panel sizing from setPanelSizes()
- **Reference Image Event Migration**: Replaced all REFERENCE_* EventBus events with presenter calls
  - ReferenceImagePanel calls presenter.setReferenceMode/Alpha/Position/Scale() directly
  - PhaserEditorComponent notifies presenter.onReferenceScaleUpdatedFromScene/onReferencePositionUpdatedFromScene()
  - Presenter calls ReferenceImagePanel.updateScaleDisplay/updatePositionDisplay() directly
  - Removed all 13 REFERENCE_* event types and 11 payload interfaces from events.ts
- **Simplified Data Flow**:
  - Tool selection: ToolsPanel.onClick() → presenter.selectTerrain() → updates state + calls phaserEditor directly
  - Visual state: index.ts.setShowGrid() → presenter.setShowGrid() → updates state + calls phaserEditor directly
  - Tile clicks: PhaserEditorComponent.handleTileClick() → presenter.handleTileClick() → modifies World data
  - Reference image controls: ReferenceImagePanel → presenter → PhaserEditorComponent.editorScene
  - Reference image drag/scroll: PhaserEditorScene → PhaserEditorComponent → presenter → ReferenceImagePanel

**Architecture Benefits**:
- Eliminated EventBus intermediary for UI state, reducing indirection
- Traceable data flow (search for `presenter.` to find all actions)
- World events still use EventBus (TILES_CHANGED, UNITS_CHANGED) - unchanged for cross-page compatibility
- Components receive presenter via setPresenter() for dependency injection
- Simpler debugging with direct method calls instead of pub/sub events

### Recent Achievements (Session 2025-01-10)

#### gRPC Error Handling for Web Pages (Complete)
- **HandleGRPCError Helper**: Created centralized error handler in `web/server/utils.go`
  - `codes.Unauthenticated` → Redirects to `/login?callbackURL=<current-url>`
  - `codes.NotFound` → Renders fun 404 "Lost in the Fog of War" page
  - `codes.PermissionDenied` → Renders 403 "Enemy Territory" page
- **Error Page Templates**: Created themed error pages
  - `web/templates/NotFoundPage.html`: War-themed 404 with navigation
  - `web/templates/ForbiddenPage.html`: Access denied page with login prompt
- **Updated Views**: All views calling gRPC services now use HandleGRPCError
  - WorldListView, WorldViewerPage, WorldEditorPage
  - GameListView, GameViewerPage, GameDetailPage
- **API Error Handling**: API requests use grpc-gateway default behavior
  - `codes.Unauthenticated` → HTTP 401 with JSON error body
  - `codes.NotFound` → HTTP 404 with JSON error body
  - `codes.PermissionDenied` → HTTP 403 with JSON error body

## Status
**Current Version**: 8.15 (gRPC Error Handling)
**Status**: Production-ready - Web pages properly handle authentication and error states
**Build Status**: Clean compilation with all TypeScript errors resolved
**Testing**: Jest (unit) + Playwright (e2e) with command interface and persistent test worlds
**Architecture**: Flexible AssetProvider system with presenter-driven animation framework, layer-based interaction system, and presenter-based UI state management
