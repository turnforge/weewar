# Web Module Summary

## Purpose
The web module provides a modern web interface for the WeeWar turn-based strategy game, featuring a professional map editor, readonly map viewer, and comprehensive game management system.

## Current Architecture (v5.0)

### Unified Map Architecture with Observer Pattern
- **Single Source of Truth**: Enhanced Map class serves as central data store for all map operations
- **Event-Driven Communication**: MapObserver interface with type-safe MapEvent system
- **Batched Performance**: TileChange and UnitChange arrays with setTimeout-based scheduling
- **Self-contained Persistence**: Map class handles save/load operations directly
- **Automatic Change Tracking**: Eliminates manual change marking throughout components

### Core Components

#### Frontend Components (`web/frontend/components/`)
- **Map.ts**: Enhanced with Observer pattern, batched events, and self-contained persistence
- **MapEditorPage.ts**: Refactored to use Map as single source of truth, implements MapObserver
- **PhaserEditorComponent.ts**: Phaser.js-based map editor with WebGL rendering
- **EventBus.ts**: Type-safe event system for inter-component communication
- **Component.ts**: Base class with lifecycle management and DOM scoping

#### Backend Services (`pkg/services/`)
- **MapsService**: gRPC service for map CRUD operations
- **File-based Storage**: Maps stored in `./storage/maps/<mapId>/` structure
- **Connect Bindings**: HTTP API integration with frontend

### Key Features

#### Map Editor
- **Phaser.js Integration**: WebGL-accelerated rendering with professional UX
- **Coordinate Accuracy**: Pixel-perfect matching with Go backend implementation
- **Observer Pattern**: Real-time component synchronization via Map events
- **Keyboard Shortcuts**: Comprehensive shortcut system for rapid map building
- **Professional Tools**: Terrain painting, unit placement, brush sizes, player management

#### Component Architecture
- **Event-Driven Design**: Components communicate via EventBus and Map Observer pattern
- **Clean Separation**: UI components focus on presentation, Map handles data operations
- **Type Safety**: Comprehensive TypeScript interfaces prevent runtime errors
- **Lifecycle Management**: Proper initialization, cleanup, and error handling

### Recent Achievements (Session 2025-01-20)

#### Unified Map Architecture Implementation
- Enhanced Map class with comprehensive Observer pattern support
- Implemented MapObserver interface with type-safe event handling
- Added batched event system for performance optimization
- Created self-contained persistence methods in Map class
- Refactored MapEditorPage to use Map as single source of truth
- Removed redundant state properties and manual change tracking
- Fixed all compilation errors and achieved clean build

#### Component State Management Architecture 
- Created MapEditorPageState class for centralized page-level state management
- Established proper component encapsulation with DOM ownership principles
- Refactored EditorToolsPanel to be state generator and exclusive DOM owner
- Eliminated cross-component DOM manipulation violations
- Implemented clean state flow: User clicks → Component updates state → State emits events → Components observe
- Separated state levels: Page-level (tools), Application-level (theme), Component-level (local UI)

#### Architecture Benefits
- **Code Reduction**: MapEditorPage simplified from 2700+ lines through centralization
- **Data Consistency**: Single source of truth eliminates scattered data copies
- **Performance**: Batched events reduce UI update frequency
- **Component Boundaries**: Proper encapsulation with each component owning its DOM elements
- **State Management**: Clean separation of state generators vs state observers
- **Maintainability**: Centralized logic easier to debug and extend
- **Type Safety**: Comprehensive interfaces prevent runtime errors

### Technical Specifications

#### Coordinate System
- **Backend**: Cube coordinates (Q/R) with proper hex mathematics
- **Frontend**: Matches backend exactly with tileWidth=64, tileHeight=64, yIncrement=48
- **Conversion**: Row/col intermediate step using odd-row offset layout
- **Accuracy**: Pixel-perfect coordinate mapping between frontend and backend

#### Rendering Pipeline
- **Phaser.js**: WebGL-accelerated rendering engine
- **Dynamic Grid**: Infinite grid system rendering only visible hexes
- **Professional Interaction**: Paint-on-release, drag-to-pan, modifier key painting
- **Asset Integration**: Direct static URLs for tile/unit graphics

#### Data Persistence
- **Map Class**: Handles own save/load operations with server API calls
- **Server Format**: Compatible with protobuf definitions for CreateMap/UpdateMap
- **Client Loading**: Supports both server data and HTML element parsing
- **Change Tracking**: Automatic via Observer pattern, eliminates manual marking

### Development Guidelines

#### Observer Pattern Usage
- Implement MapObserver interface in components needing map updates
- Subscribe to Map events in component initialization
- Handle specific event types (TILES_CHANGED, UNITS_CHANGED, MAP_LOADED, etc.)
- Use batched events for performance optimization

#### Component Development
- Extend BaseComponent for lifecycle management
- Use EventBus for inter-component communication
- Implement proper cleanup in destroyComponent()
- Scope DOM queries to component containers

#### Map Operations
- Use Map class methods for all tile/unit operations
- Let Map handle persistence and change tracking automatically
- Subscribe to Map events for UI updates
- Avoid direct manipulation of map data outside Map class

### Next Development Priorities

#### Component Integration Completion
- Update PhaserEditorComponent to subscribe to Map events
- Update TileStatsPanel to read from Map instead of Phaser
- Remove redundant getTilesData/setTilesData methods
- Test complete component synchronization via Map events

#### Performance Optimization
- Performance testing with large maps
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
**Current Version**: 5.1 (Component State Management with Encapsulation)  
**Status**: Production-ready with proper component boundaries and state management  
**Build Status**: Clean compilation with all TypeScript errors resolved  
**Next Milestone**: Complete component integration with unified state management and Games Management System