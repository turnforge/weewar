**Purpose:**

This folder contains the page-centric TypeScript source code for the webapp, organized by feature pages with shared common code. Each page directory contains all components specific to that page, while shared code lives in the common/ directory.

**Directory Structure:**

## Page Organization

*   **Root-level Pages** (single-file pages): `HomePage.ts`, `LoginPage.ts`, `StartGamePage.ts`, `GameListingPage.ts`, `WorldListingPage.ts`, `WorldViewerPage.ts`, `UserDetailsPage.ts`, `AttackSimulatorPage.ts`
*   **GameViewerPage/** - Complex game viewing/playing page with multiple layout variants
*   **WorldEditorPage/** - World/map editor with tools and panels
*   **common/** - Shared code, utilities, and animations used across all pages

## Page-Centric Architecture

### GameViewerPage/
Complex interactive game interface with multiple layout variants:
*   **GameViewerPageBase.ts** - Abstract base class with all core game logic (WASM, presenter, panels, RPC methods)
*   **GameViewerPageDockView.ts** - Flexible dockable layout variant
*   **GameViewerPageGrid.ts** - Static CSS grid layout variant
*   **GameViewerPageMobile.ts** - Touch-optimized mobile variant
*   **GameState.ts** - WASM state management component
*   **PhaserGameScene.ts** - Phaser rendering for game board
*   **BuildOptionsModal.ts** - Modal for unit construction
*   **CompactSummaryCard.ts** - Mobile header summary
*   **TurnOptionsPanel.ts**, **UnitStatsPanel.ts**, **TerrainStatsPanel.ts** - UI panels
*   **DamageDistributionPanel.ts**, **GameLogPanel.ts** - Information displays

### WorldEditorPage/
World/map editor with tools and visualization:
*   **index.ts** - Main editor page orchestrator
*   **PageState.ts** - Editor state management
*   **PhaserEditorComponent.ts**, **PhaserEditorScene.ts** - Phaser integration for editing
*   **ToolsPanel.ts**, **ReferenceImagePanel.ts** - UI panels
*   **ReferenceImageDB.ts**, **ReferenceImageLayer.ts** - Reference image system
*   **tools/** - Shape drawing tools (ShapeTool, CircleTool, LineTool, OvalTool, RectangleTool)

### common/
Shared code across all pages:
*   **Core** - `World.ts`, `PhaserWorldScene.ts`, `LayerSystem.ts`, `BaseMapLayer.ts`, `HexHighlightLayer.ts`
*   **Utils** - `hexUtils.ts`, `ColorsAndNames.ts`, `ThemeUtils.ts`, `AssetThemePreference.ts`, `RulesTable.ts`
*   **Events** - `events.ts` (GameEventTypes, WorldEventTypes, EditorEventTypes)
*   **Panels** - `WorldStatsPanel.ts` (unified world statistics panel with tile/unit breakdowns and player distribution)
*   **animations/** - Animation system (`AnimationConfig.ts`, `effects/`)

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

## Recent Session Work (2025-11-26)

### WorldStatsPanel Unification
Merged TileStatsPanel (WorldEditorPage) and WorldStatsPanel (common) into a single unified component:

**Features:**
- Grid-based tile/unit breakdown with theme-specific icons and names
- Responsive flex-wrap layout for totals (adjusts columns based on width)
- Alphabetical sorting by terrain/unit name using theme's naming methods
- Player distribution tables with centered icon+name in first column
- Listens to TILES_CHANGED, UNITS_CHANGED, WORLD_LOADED events for auto-refresh
- Takes a World instance via `setWorld()` method

**Usage Pattern:**
```typescript
// Create panel and inject World dependency
this.worldStatsPanel = new WorldStatsPanel(container, this.eventBus, debugMode);
this.worldStatsPanel.setWorld(this.world);

// Panel auto-updates on world events, or manually refresh:
this.worldStatsPanel.refreshStats();
```

**Files Changed:**
- Removed: `WorldEditorPage/TileStatsPanel.ts`
- Updated: `common/WorldStatsPanel.ts` (complete rewrite)
- Updated: `WorldEditorPage/index.ts`, `WorldEditorPage/WorldEditorPresenter.ts`
- Updated: `WorldViewerPage.ts`

## Recent Session Work (2025-01-05)

### Phaser Scene Sizing Fix ✅
**Problem Solved**: Circular sizing issue where Phaser canvas would cause recursive growth with parent containers

**Root Causes:**
- Flexbox `min-height: auto` default prevented containers from shrinking below content size
- Canvas `object-fit: contain` maintained aspect ratio, causing height changes when width changed
- Missing `min-height: 0` constraints on flex children broke one-way sizing flow

**Solutions Implemented:**
- **common/PhaserWorldScene.ts**: Removed `object-fit: contain` from canvas styling
- **PhaserSceneView Template**: Created reusable BorderLayout component with built-in sizing constraints
- **FlexMode Parameter**: Automatic application of `flex: 1 1 0%; min-height: 0; min-width: 0;` to wrapper
- **Go Template Safety**: Fixed ZgotmplZ issue by using inline conditionals instead of style variables

**Key Pattern - min-height: 0**:
```typescript
// Critical for preventing circular sizing in flexbox:
// Container with min-height: 0 can shrink below content size
// This breaks: Canvas grows → Container grows → Parent grows → Loop
```

**Benefits:**
- One-way sizing flow: parent → canvas (never canvas → parent)
- Width changes don't affect height (no aspect ratio constraint)
- Works with all scene types (PhaserWorldScene, PhaserEditorScene, PhaserGameScene)
- Eliminates wrapper div boilerplate in pages

**Migrated Pages:**
- WorldViewerPage ✅ (scene only)
- WorldEditorPage ✅ (toolbar + scene, FlexMode="fixed")

**TypeScript Integration:**
- WorldEditorPage/PhaserEditorComponent.ts simplified - container ID no longer renamed
- All scene types (common/PhaserWorldScene, WorldEditorPage/PhaserEditorScene, GameViewerPage/PhaserGameScene) work unchanged
- Container ID from template SceneId parameter used directly

**Documentation:**
- `/web/templates/components/PhaserSceneView_README.md` - Component usage guide with block inheritance pattern
- `/web/templates/components/PhaserSceneView_INTEGRATION.md` - Migration guide with template conventions
- `/web/PHASER_SIZING_FIX_SUMMARY.md` - Complete technical summary
- `/web/WORLDEDITORPAGE_MIGRATION.md` - WorldEditorPage migration record

## Recent Session Work (2025-01-24)

### Layer System Architecture Complete ✅
*   **Generic WorldViewer**: `WorldViewer<TScene>` with template parameter for proper typing
*   **GameViewer Specialization**: `GameViewer extends WorldViewer<PhaserGameScene>` with game-specific layer access
*   **Layer-Based Interaction**: Direct layer manipulation (`getSelectionHighlightLayer()`, `getMovementHighlightLayer()`, etc.)
*   **Editor Integration**: WorldEditorPage/PhaserEditorComponent.ts uses layer callbacks for painting logic
*   **Callback Architecture**: Click handling through common/BaseMapLayer.ts callbacks with validation in components
*   **Brush Size Support**: Multi-tile painting with hex distance calculations in component layer

### Architecture Improvements ✅
*   **Scene Separation**: common/PhaserWorldScene.ts for rendering, page components for business logic
*   **Single Source of Truth**: common/World.ts model updates trigger observer pattern for visual updates
*   **Type Safety**: Proper TypeScript generics eliminate casting and improve developer experience
*   **Clean Separation**: UI logic in page components, rendering logic in common scenes, interaction through layers

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
*   **Layout-Specific Child Classes**: GameViewerPageDockView (flexible dockable), GameViewerPageGrid (static CSS grid), GameViewerPageMobile (touch-optimized)
*   **Abstract Method Pattern**: Child classes implement layout-specific concerns (initializeLayout, createPanels, getGameSceneContainer)
*   **Zero Logic Duplication**: All game logic lives in base class, children only handle structural differences
*   **Flexible Timing Control**: Child classes control when game scene is created (early for Grid/Mobile, late for DockView)
*   **Server-Side Template Rendering**: Each variant has its own HTML template with appropriate layout structure

**Mobile Variant Implementation ✅:**
- **GameViewerPage/GameViewerPageMobile.ts**: Mobile page with context-aware bottom action bar
- **lib/MobileBottomDrawer.ts**: Reusable drawer component (60-70% height, auto-close on backdrop tap)
- **GameViewerPage/CompactSummaryCard.ts**: Top banner showing terrain+unit selection info
- **Context-Aware Button Ordering**: Dynamic reordering inferred from allowed panels (unit/tile/nothing context)
- **setCompactSummaryCard RPC**: Presenter-rendered HTML sent via RPC (CompactSummaryCard.templar.html)
- **Bottom Drawers**: 5 drawers (unit stats, terrain stats, damage, actions, log), one open at a time
- **Layout Structure**: Header (70px) → Compact Card (56px absolute) → Game Scene → Bottom Bar (64px fixed)
- **Presenter-Driven**: All UI updates via RPC calls, no event bus subscriptions needed

**Architecture Benefits:**
- Clean separation of game logic from layout concerns
- Easy to add new layout variants without code duplication
- Can serve different variants based on user agent/screen size
- Each variant optimized for its use case (performance, flexibility, touch)
- Reusable components in web/lib for cross-page usage

**Template-Based Panel Rendering Pattern (Session 2025-01-04):**

Refactored inline HTML generation to use server-side Go templates for cleaner architecture:

**Before (Inline HTML in Go):**
```go
// services/gameview_presenter.go
func (s *GameViewPresenter) renderCompactSummaryCard(tile, unit) string {
    html := `<div class="flex items-center">`
    html += fmt.Sprintf(`<img data-tile-id="%d" />`, tile.TileType)
    // ... 50+ lines of HTML string concatenation
    return html
}
```

**After (Template-Based):**
```go
// services/gameview_presenter.go - Clean interface call
s.CompactSummaryCardPanel.SetCurrentData(ctx, tile, unit)

// cmd/lilbattle-wasm/browser.go - Template rendering
content := renderPanelTemplate(ctx, "CompactSummaryCard.templar.html", map[string]any{
    "Tile":  tile,
    "Unit":  unit,
    "Theme": theme,
})
```

**Panel Interface Architecture:**
1. **Interface Definition** (services/gameview_presenter.go):
   - `CompactSummaryCardPanel` interface with `SetCurrentData(tile, unit)`
   - Added to `BaseGameViewPresenter` struct

2. **Base Implementation** (services/panels.go):
   - `BaseCompactSummaryCardPanel` for CLI/non-browser (stores data only)

3. **Browser Implementation** (cmd/lilbattle-wasm/browser.go):
   - `BrowserCompactSummaryCardPanel` renders template and calls RPC

4. **Template File** (web/templates/CompactSummaryCard.templar.html):
   - Go template with conditionals and theme integration
   - Clean separation of structure from logic

**Benefits:**
- No HTML strings in Go code
- Template syntax highlighting and validation
- Consistent with other panels (TurnOptions, UnitStats, etc.)
- Easier to maintain and modify UI
- Clear separation: Go handles data, templates handle presentation

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

**WorldListingPage.ts** - Worlds listing page (root-level)
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

**GameListingPage.ts** - Games listing page (root-level)
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

## Recent Session Work (2025-11-27)

### Crossing System Redesign with Explicit Connectivity

Complete redesign of the crossing system to use explicit connectivity instead of automatic neighbor detection:

**New Crossing Data Model:**
- Changed from `CrossingType` enum to `Crossing` struct with `type` and `connectsTo[6]` boolean array
- Direction indices: 0=LEFT, 1=TOP_LEFT, 2=TOP_RIGHT, 3=RIGHT, 4=BOTTOM_RIGHT, 5=BOTTOM_LEFT
- Proto schema updated: `crossings` map now stores `Crossing` objects instead of `CrossingType` values

**Crossing Direction Preset UI:**
- Interactive hex SVG with 6 clickable direction lines
- Dropdown presets: Horizontal (L-R), Diagonal (TL-BR), T-Junction, Crossroad, Full
- Individual direction toggles via clicking on hex spokes
- Visual feedback: active directions show amber (#d97706), inactive show gray (#d1d5db)
- Larger interactive hex (w-32 h-32) for easier selection

**Editor Behavior Changes:**
- `toggleCrossing` now uses `deleteCrossing()` instead of `removeCrossing()` to only affect clicked tile
- Removing a crossing no longer affects neighbor connections
- Placing a crossing uses the preset `connectsTo` configuration from ToolsPanel

**Files Changed:**
- `protos/lilbattle/v1/game.proto`: Changed `Crossing` from enum to message with connectsTo array
- `common/hexUtils.ts`: Added `getOppositeDirection()` helper
- `common/World.ts`: Added `deleteCrossing()` method, updated crossing APIs
- `common/CrossingLayer.ts`: Render connections based on explicit `connectsTo` array
- `WorldEditorPage/WorldEditorPresenter.ts`: Added `setCrossingConnectsTo()`, updated toggle logic
- `WorldEditorPage/ToolsPanel.ts`: Crossing direction preset UI, SVG interaction
- `web/templates/panels/ToolsPanel.html`: Interactive hex SVG markup

### DefaultTheme Refactoring

Refactored `default.ts` to extend `BaseTheme` for consistency with other themes:

**Changes:**
- `DefaultTheme` now extends `BaseTheme` instead of implementing `ITheme` directly
- Added `mapping.json` for default theme with unit/terrain mappings and `natureTerrains` array
- Consolidated assets from `web/static/assets/v1/` to `web/static/assets/themes/default/`
- Added `getCrossingDisplayTileType()` method to `ITheme` interface and `BaseTheme`
- `WorldStatsPanel` now uses theme method instead of hardcoded tile type constants

**Benefits:**
- Consistent theme architecture across all themes (default, fantasy, modern)
- Shared `BaseTheme` methods reduce code duplication
- Single location for v1 assets
- Theme can provide crossing display tile types for WorldStatsPanel

**Files Changed:**
- `web/assets/themes/default.ts`: Refactored to extend BaseTheme
- `web/assets/themes/BaseTheme.ts`: Added `getCrossingDisplayTileType()` method
- `web/assets/themes/default/mapping.json`: New mapping file
- `web/pages/common/WorldStatsPanel.ts`: Uses theme method for crossing display
