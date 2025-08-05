# Web Module Summary

## Purpose
The web module provides a modern web interface for the WeeWar turn-based strategy game, featuring a professional world editor, readonly world viewer, and comprehensive game management system.

## Current Architecture (v8.0)

### LCMComponent Lifecycle Architecture with EventSubscriber Pattern
- **4-Phase Lifecycle Management**: performLocalInit() → setupDependencies() → activate() → deactivate()
- **Breadth-First Initialization**: LifecycleController orchestrates component coordination with synchronization barriers
- **EventSubscriber Interface**: Type-safe event handling via handleBusEvent() method, replacing callback-based subscriptions
- **World-Centric Data Management**: Enhanced World class serves as single source of truth with observer pattern
- **Race Condition Elimination**: Synchronization barriers prevent timing issues and component coordination problems

### Core Components

#### Frontend Components (`web/src/`)
- **World.ts**: Enhanced with Observer pattern, batched events, and self-contained persistence
- **WorldEditorPage.ts**: Refactored to use World as single source of truth, implements WorldObserver
- **PhaserEditorComponent.ts**: Phaser.js-based world editor with WebGL rendering

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

### Recent Achievements (Session 2025-08-05)

#### Phaser Architecture Unification (v8.0)
- **Wrapper Elimination**: Removed PhaserWorldEditor and PhaserPanel unnecessary wrapper classes
- **Unified Scene Architecture**: PhaserWorldScene as base class, PhaserEditorScene as editor extension
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
- Enhanced World class with comprehensive Observer pattern support
- Implemented WorldObserver interface with type-safe event handling
- Added batched event system for performance optimization
- Created self-contained persistence methods in World class
- Refactored WorldEditorPage to use World as single source of truth
- Removed redundant state properties and manual change tracking
- Fixed all compilation errors and achieved clean build

#### Previous Session: Component State Management Architecture 
- Created WorldEditorPageState class for centralized page-level state management
- Established proper component encapsulation with DOM ownership principles
- Refactored EditorToolsPanel to be state generator and exclusive DOM owner
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
- Use World class methods for all tile/unit operations (single source of truth)
- Pass World objects (not raw data) to Phaser components via loadWorld() method
- Subscribe to World events for automatic UI synchronization
- Let World handle persistence and change tracking automatically

### Next Development Priorities

#### Component Integration Completion
- Update PhaserEditorComponent to subscribe to World events
- Update TileStatsPanel to read from World instead of Phaser
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

## Status
**Current Version**: 8.0 (Phaser Architecture Unification)  
**Status**: Production-ready with unified Phaser scene architecture and container management fixes  
**Build Status**: Clean compilation with all TypeScript errors resolved  
**Architecture**: Simplified wrapper-free Phaser integration with proper container targeting  
**Next Milestone**: Additional container fixes for other pages and performance optimization
