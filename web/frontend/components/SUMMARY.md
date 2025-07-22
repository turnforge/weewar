# ./web/frontend/components/ Summary

**Purpose:**

This folder contains the core client-side TypeScript logic for the webapp, managing UI state, user events, API interactions, and DOM manipulation using a modern component-based architecture with strict separation of concerns and event-driven communication.

**Core Architecture Components:**

## Modern Component System (New)

*   **`EventBus.ts`**: Type-safe, synchronous event system with error isolation and source exclusion for inter-component communication
*   **`Component.ts`**: Base interface and abstract class defining standard component lifecycle with simplified constructor pattern
*   **`MapViewer.ts`**: Phaser-based map visualization component with proper DOM scoping and event-driven initialization  
*   **`MapStatsPanel.ts`**: Statistics display component with safe DOM selectors and event-driven updates
*   **`MapDetailsPage.ts`**: Orchestrator page following new architecture - handles data loading and component coordination only
*   **`UI_DESIGN_PRINCIPLES.md`**: Comprehensive documentation of architecture decisions, timing patterns, and critical lessons learned

## Component Features

*   **Strict DOM Scoping**: Components only access DOM within their root elements using `this.findElement()`
*   **Event-Driven Communication**: All inter-component communication through EventBus, no direct method calls
*   **Layout vs Behavior Separation**: Parents control layout/sizing, components handle internal behavior only
*   **HTMX Integration Ready**: Components support both initialization and hydration patterns
*   **Error Isolation**: Component failures don't cascade to other components
*   **Simplified Constructor Pattern**: `new Component(rootElement, eventBus)` - parent ensures root element exists

## Legacy Components (Being Migrated)

*   **Section-Based System**: `BaseSection.ts`, `TextSectionView/Edit.ts`, `DrawingSectionView/Edit.ts` - older composition pattern
*   **Managers & Handlers**: `ThemeManager.ts`, `Modal.ts`, `ToastManager.ts`, `TableOfContents.ts` - utility components  
*   **Map Editor**: `MapEditorPage.ts` - interactive canvas-based hex grid map editor (needs migration to new architecture)
*   **Other Pages**: `DesignEditorPage.ts`, `HomePage.ts`, `LoginPage.ts` - various page implementations

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

*   **Phaser.js**: WebGL-based map rendering with proper timing handling
*   **HTMX**: Component hydration support for server-driven UI updates  
*   **Canvas/WebGL**: Specialized initialization patterns for graphics contexts
*   **Toast/Modal Systems**: User feedback and interaction patterns
*   **Theme Management**: Coordinated theming across component boundaries

## Recent Session Work (2025-01-20)

### Component Architecture Cleanup ✅
*   **MapEditorPage Streamlining**: Removed dead code and consolidated component management patterns
*   **Panel Integration Optimization**: Improved coordination between EditorToolsPanel, TileStatsPanel, and PhaserEditor
*   **Reference Management**: Cleaner component initialization and lifecycle patterns
*   **State Management Consolidation**: Reduced complexity in page-level state handling

### Dead Code Elimination ✅
*   **Unused Method Removal**: Eliminated obsolete functionality and redundant state management
*   **Import Cleanup**: Removed unnecessary dependencies and unused imports
*   **Code Organization**: Improved maintainability through method consolidation
*   **Technical Debt Reduction**: Streamlined component boundaries and event handling
