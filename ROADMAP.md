# WeeWar Development Roadmap

## Overview
WeeWar is evolving from a comprehensive CLI-based turn-based strategy game into a full-featured web application template. This roadmap outlines the progression from game engine to web platform.

## ✅ Phase 1: Game Engine Foundation (Completed)
**Status**: Production-ready  
**Timeline**: Completed 2024-2025

### Core Engine ✅
- [x] Unified Game Architecture with interface-driven design
- [x] Hex Board System with sophisticated grid and pathfinding
- [x] Combat System with probabilistic damage and authentic mechanics
- [x] Movement System with terrain-specific costs and A* pathfinding
- [x] Complete unit database (44 unit types, 26 terrain types)
- [x] Authentic game data integration from tinyattack.com

### Professional CLI Interface ✅
- [x] REPL with chess notation (A1, B2, etc.)
- [x] PNG rendering with hex grid visualization
- [x] Session recording and replay capabilities
- [x] Comprehensive testing suite (100+ tests)
- [x] Save/load functionality with JSON persistence

## ✅ Phase 2: Web Foundation (Completed January 2025)
**Status**: Production-ready  
**Timeline**: Completed 2025-01-14

### Backend Infrastructure ✅
- [x] Complete gRPC service architecture (MapsService, GamesService, UsersService)
- [x] File-based storage system with `./storage/maps/<mapId>/` structure
- [x] Enhanced protobuf models with hex coordinates (MapTile, MapUnit)
- [x] Full CRUD operations for maps with metadata and game data separation
- [x] Connect bindings for web API integration

### Frontend Architecture ✅
- [x] Professional view system (MapListingPage, MapEditorPage, MapDetailPage)
- [x] Template system with Tailwind CSS styling and responsive design
- [x] Route handling via setupMapsMux() with clean URL structure
- [x] Navigation flow: List → Create/Edit → View workflow

### Current Web Capabilities ✅
- [x] `/maps` - Professional maps listing with grid layout, search, and sort
- [x] `/maps/new` - Route ready for map editor implementation
- [x] `/maps/{id}/edit` - Route ready for map editor implementation  
- [x] `/maps/{id}/view` - Map details and metadata display
- [x] File persistence with JSON storage for all map data

## ✅ Phase 3: Map Editor Implementation (Completed January 2025)
**Status**: Completed  
**Timeline**: Completed 2025-01-17

### WASM-Based Editor ✅
- [x] Professional 3-panel editor layout ported from `oldweb/editor.html`
- [x] Complete terrain painting interface with 5 terrain types (Grass, Desert, Water, Mountain, Rock)
- [x] Brush system with 6 sizes from single hex to XX-Large (91 hexes)
- [x] Paint, flood fill, and terrain removal tools with coordinate targeting
- [x] Undo/redo history system with availability indicators
- [x] Map rendering with multiple output sizes and PNG export
- [x] Game export functionality for 2/3/4 player games with JSON download
- [x] Advanced tools: pattern generation, island creation, mountain ridges, terrain stats

### Editor Integration ✅
- [x] Complete TypeScript integration with proper event delegation
- [x] WASM module ready with Go backend providing all editor functions
- [x] Clean architecture following established XYZPage.ts → gen/XYZPage.html pattern
- [x] Professional UI with Tailwind CSS and dark mode support
- [x] Real-time console output and status tracking

### TypeScript Component ✅
- [x] MapEditorPage.ts component with full WASM integration structure
- [x] Data-attribute based event handling (no global namespace pollution)
- [x] Theme management integration with existing ThemeManager
- [x] Responsive design with mobile-friendly layout
- [x] Toast notifications and modal dialog support ready

### Current Status ✅
- Interactive canvas-based editor with real-time hex grid visualization
- Canvas terrain painting with click-to-paint functionality and coordinate tracking
- Map resizing controls with Add/Remove buttons on all 4 sides of canvas
- Grid-based terrain palette showing all 6 terrain types with visual icons
- Streamlined 2-panel layout (removed rendering/export panels, kept Advanced Tools)
- Clean event delegation using data attributes with proper TypeScript types
- Consolidated editorGetMapBounds WASM function for efficient data retrieval
- Default map size set to 5x5 on startup for better user experience
- Enhanced client-side coordinate conversion with proper XYToQR implementation
- Ready for WASM build and backend API integration

## ✅ Phase 4: Phaser.js Map Editor (Completed January 2025)
**Status**: Completed  
**Timeline**: January 2025

### Phaser.js Integration ✅
- [x] Complete migration from canvas-based to Phaser.js WebGL rendering
- [x] Professional map editor with modern game engine foundation
- [x] Dynamic hex grid system covering entire visible camera area
- [x] Accurate coordinate conversion matching Go backend (`lib/map.go`) exactly
- [x] Interactive controls: mouse wheel zoom, drag pan, keyboard navigation

### Coordinate System Accuracy ✅
- [x] Fixed coordinate conversion to match backend implementation exactly
- [x] `tileWidth=64, tileHeight=64, yIncrement=48` matching `lib/map.go`
- [x] Row/col conversion using odd-row offset layout from `lib/hex_coords.go`
- [x] Pixel-perfect click-to-hex coordinate mapping
- [x] Eliminated coordinate drift between frontend and backend

### Professional Mouse Interaction ✅
- [x] Paint on mouse up (not down) to prevent accidental painting during camera movement
- [x] Drag detection with threshold to distinguish between painting and panning
- [x] Camera pan on drag without modifier keys for smooth navigation
- [x] Paint mode with Alt/Cmd/Ctrl + drag for continuous painting
- [x] Immediate paint on modifier key down for responsive feedback

### UI Architecture Improvements ✅
- [x] PhaserPanel class for clean editor logic separation
- [x] Grid and coordinate toggles moved from ToolsPanel to PhaserPanel
- [x] Removed "Switch to Canvas" button (legacy canvas system eliminated)
- [x] Event callback system for tile clicks and map changes
- [x] Clean initialization and cleanup methods

### Dynamic Grid System ✅
- [x] Camera viewport bounds calculation for efficient grid rendering
- [x] Dynamic hex coordinate range based on visible area (not fixed radius)
- [x] Efficient rendering of only visible grid hexes for performance
- [x] Automatic grid updates when camera moves or zooms
- [x] Performance optimization for large coordinate ranges

### Benefits Achieved ✅
- **Modern Architecture**: WebGL-accelerated rendering with professional game engine
- **Coordinate Accuracy**: Pixel-perfect frontend/backend coordinate matching
- **Professional UX**: Intuitive controls preventing accidental tile painting
- **Performance**: Dynamic rendering covering only visible area
- **Maintainability**: Clean component separation with event-driven architecture
- **Extensibility**: Phaser.js foundation enables advanced features (animations, effects)

## 🎯 Phase 5: Readonly Map Viewer (Completed)
**Status**: Production-ready  
**Timeline**: Completed January 2025

### Complete Readonly Map Display System ✅
- [x] PhaserViewer component for readonly map display without editing capabilities
- [x] MapDetailsPage integration with full map loading and statistics
- [x] Proper Phaser.js initialization timing and WebGL context management
- [x] Real-time map statistics calculation and display
- [x] Copy map functionality for creating new maps from existing ones
- [x] Template integration with proper JavaScript bundle loading

### Critical Bug Fixes and Learnings ✅
- [x] **DOM Corruption Prevention**: Fixed dangerous CSS selectors that could replace entire page content
- [x] **Phaser Timing Issues**: Resolved WebGL framebuffer errors with proper initialization sequencing  
- [x] **Container Sizing**: Implemented proper container dimension handling for Phaser canvas
- [x] **Asset Loading**: Ensured proper asset loading sequence before map data visualization
- [x] **Error Handling**: Added comprehensive error handling for WebGL and initialization failures

### Architecture Insights ✅
- **Critical Learning**: Broad CSS selectors like `.text-gray-900, .text-white` can match major page elements (including `<body>`)
- **DOM Safety**: Always scope DOM queries to specific container elements to prevent accidental page-wide changes
- **Phaser Timing**: WebGL contexts require proper container sizing before initialization to avoid framebuffer errors
- **Race Conditions**: Map data loading must be sequenced after Phaser initialization to prevent DOM corruption
- **Template Integration**: JavaScript bundle inclusion requires proper template structure and build system coordination

## ⌨️ Phase 6: Keyboard Shortcut System (Completed)
**Status**: Production-ready  
**Timeline**: Completed January 2025

### Comprehensive Keyboard Shortcuts ✅
- [x] Generic KeyboardShortcutManager class for reusable architecture
- [x] Multi-key command system: `n12` (nature terrain), `c5` (city terrain), `u3` (unit type)
- [x] Smart number handling with backspace editing and timeout management
- [x] Context-aware shortcuts (disable in input fields, modals)
- [x] Help system with `?` key overlay showing all available shortcuts
- [x] Toast notifications and visual feedback for all shortcut actions

### Map Editor Shortcuts ✅
- [x] `n<index>` - Select nature terrain by index (1-5: Grass, Desert, Water, Mountain, Rock)
- [x] `c<index>` - Select city terrain by index (1-4: city variants)
- [x] `u<index>` - Select unit type for current player (1-20: all unit types)
- [x] `p<number>` - Set current player (1-4)
- [x] `b<size>` - Set brush size (0-5: Single to XX-Large)
- [x] `esc` - Reset all tools to defaults
- [x] `?` - Show comprehensive help overlay with categorized shortcuts

### Benefits Achieved ✅
- **Rapid Workflow**: Significantly faster map building with keyboard-first approach
- **One-handed Operation**: Optimized for mouse + keyboard workflow
- **Reusable Architecture**: Framework can be used across all application pages
- **Professional UX**: Industry-standard keyboard shortcut conventions
- **Context Intelligence**: Smart activation based on current page and input state
- **Clean Architecture**: Separation of concerns between input handling and UI updates

### Technical Implementation ✅
- **State Machine**: NORMAL ↔ AWAITING_ARGS with visual indicators
- **Input Validation**: Proper bounds checking with error feedback
- **UI Synchronization**: Updates terrain/unit buttons, dropdowns, and visual state
- **Help Generation**: Auto-generated help content from shortcut configuration
- **Error Handling**: Clear validation messages for invalid inputs

## ⚡ Phase 5.1: Unified Map Architecture (Completed)
**Status**: Completed  
**Timeline**: January 2025

### Observer Pattern Implementation ✅
- [x] Enhanced Map class with comprehensive Observer pattern support
- [x] MapObserver interface with type-safe event handling
- [x] Batched event emissions for performance optimization
- [x] Self-contained persistence (Map can save/load itself)
- [x] Single source of truth architecture eliminating data duplication
- [x] MapEditorPage refactored to use Map as central data store

### Technical Achievements ✅
- [x] **Map Class Enhancement**: Added Observer interfaces, batched events, and persistence methods
- [x] **Event System**: MapEvent with types (TILES_CHANGED, UNITS_CHANGED, MAP_LOADED, MAP_SAVED, MAP_CLEARED, MAP_METADATA_CHANGED)
- [x] **Batched Changes**: TileChange and UnitChange arrays with setTimeout-based batch scheduling
- [x] **Data Consolidation**: Removed redundant properties from MapEditorPage (currentMapId, isNewMap, hasUnsavedChanges)
- [x] **Self-contained Operations**: Map handles its own loading from server data and HTML elements
- [x] **Automatic Change Tracking**: Map changes automatically tracked without manual markAsChanged calls

### Architecture Benefits ✅
- **Single Source of Truth**: All map data flows through enhanced Map class
- **Event-Driven Updates**: Components react to Map changes via Observer pattern
- **Performance Optimization**: Batched events prevent excessive UI updates
- **Clean Separation**: MapEditorPage focuses on UI orchestration, Map handles data
- **Type Safety**: Comprehensive TypeScript interfaces for all event data
- **Maintainability**: Centralized map logic easier to debug and extend

## ⚡ Phase 5.2: Component State Management (Completed)
**Status**: Completed  
**Timeline**: January 2025

### Page-Level State Architecture ✅
- [x] MapEditorPageState class with comprehensive state management for page-level UI
- [x] Proper separation of Page-level vs Application-level vs Component-level state
- [x] Component encapsulation with DOM ownership principles enforced
- [x] State generator pattern with EditorToolsPanel owning its DOM interactions
- [x] Elimination of cross-component DOM manipulation violations

### Technical Achievements ✅
- [x] **MapEditorPageState**: Centralized state for tool selection, visual settings, and workflow state
- [x] **Component Boundaries**: EditorToolsPanel owns terrain/unit buttons, generates state changes
- [x] **Encapsulation Enforcement**: Removed MapEditorPage's direct manipulation of component DOM elements
- [x] **State Flow**: User clicks → Component updates state → State emits events → Other components observe
- [x] **DOM Ownership**: Each component manages only its own DOM elements and CSS classes
- [x] **Type Safety**: Comprehensive state interfaces with granular change tracking

### Design Principles Established ✅
- **Component DOM Ownership**: Components own their DOM elements, no external manipulation
- **State Generator Pattern**: UI components generate state changes, don't just observe them
- **Proper Encapsulation**: MapEditorPage coordinates but never touches component internals
- **Clean State Flow**: Unidirectional data flow from user action to state to UI updates
- **Separation of Concerns**: Page-level (tools), Application-level (theme), Component-level (local UI)
- **Event-Driven Architecture**: State changes drive component updates via Observer pattern

## ⚡ Phase 5.3: Component Architecture Cleanup (Completed)
**Status**: Completed  
**Timeline**: January 2025

### Technical Debt Reduction ✅
- [x] Comprehensive cleanup of MapEditorPage with dead code elimination
- [x] Component reference streamlining and initialization pattern improvements
- [x] Panel integration optimization between EditorToolsPanel, TileStatsPanel, and PhaserEditor
- [x] Import cleanup and removal of unnecessary dependencies
- [x] Method consolidation and code organization improvements
- [x] State management simplification and complexity reduction

### Code Quality Improvements ✅
- [x] **Dead Code Elimination**: Removed unused methods, obsolete state properties, and redundant functionality
- [x] **Component Lifecycle**: Simplified initialization and cleanup patterns for better maintainability
- [x] **Reference Management**: Cleaner component reference patterns and predictable initialization order
- [x] **Event System Optimization**: Streamlined event delegation and component communication
- [x] **Architectural Clarity**: Better separation of concerns and component boundary clarification
- [x] **Maintainability**: Improved readability through consolidated methods and reduced cognitive complexity

### Benefits Achieved ✅
- **Cleaner Codebase**: Significant reduction in technical debt and improved code organization
- **Better Maintainability**: Simplified component architecture with clearer boundaries
- **Enhanced Readability**: Consolidated functionality and streamlined component interfaces
- **Reduced Complexity**: Simplified state management and component communication patterns
- **Improved Architecture**: Better separation of concerns and more predictable component behavior

## 📋 Phase 6: Games Management System (Planned)
**Status**: Planned  
**Timeline**: February 2025

### Games Infrastructure
- [ ] GamesService implementation with file-based storage
- [ ] Game state management with turn-based mechanics
- [ ] Player management and game session handling
- [ ] Game listing and creation workflows

### Web Interface
- [ ] Games listing page similar to maps listing
- [ ] Game creation wizard with map selection
- [ ] Game details page with current state display
- [ ] Player dashboard and game management

## 🎯 Phase 7: Gameplay Integration (Planned)
**Status**: Future  
**Timeline**: Q2 2025

### Web-Based Gameplay
- [ ] Integration of CLI game engine with web interface
- [ ] Real-time game state updates and turn management
- [ ] Player actions via web interface
- [ ] Game rendering and visualization in browser

### Advanced Features
- [ ] AI player support for single-player games
- [ ] Multiplayer session management
- [ ] Tournament mode with rankings and statistics
- [ ] Advanced analytics and game history

## 🔮 Phase 8: Platform Features (Future)
**Status**: Future vision  
**Timeline**: 2025-2026

### Community Features
- [ ] User profiles and authentication system
- [ ] Map sharing and community galleries
- [ ] Rating and review systems
- [ ] Social features and player interactions

### Advanced Capabilities
- [ ] Real-time multiplayer with WebSocket support
- [ ] Mobile-responsive design and PWA features
- [ ] Advanced AI using game theory and machine learning
- [ ] Integration with external gaming platforms

## Technical Architecture Goals

### Current Architecture Strengths
- **Clean separation**: Backend (gRPC), Frontend (Templates), Storage (Files)
- **Scalable design**: Interface-driven with clear contracts
- **Performance**: File-based storage with metadata/data separation
- **Maintainability**: Well-documented with comprehensive testing

### Future Architecture Evolution
- **Database migration**: Move from file storage to proper database
- **Caching layer**: Add Redis/memcached for performance
- **Microservices**: Split into focused service components
- **Container deployment**: Docker and Kubernetes support

## Success Metrics

### Phase 2 Achievements ✅
- Professional maps listing page with real data from file storage
- Complete backend API with full CRUD operations
- Clean routing and navigation flow
- Foundation ready for editor implementation

### Phase 3 Achievements ✅
- Professional map editor with complete terrain painting interface
- WASM integration architecture ready for Go backend connection
- Clean TypeScript component following project conventions
- Professional 3-panel layout with all editor tools and controls

### Phase 4 Goals 🎯
- WASM build integration and backend API connection
- Save/load functionality with file storage
- Complete map creation and editing workflow
- Games management system implementation

### Recent Session Progress (2025-01-17) ✅
- **Phaser.js Architecture Complete**: Fully implemented WebGL-based map editor
- **Coordinate System Fixed**: Pixel-perfect matching between frontend and backend
- **Professional UX Implemented**: Intuitive mouse interaction preventing accidental painting  
- **Component Architecture**: Clean PhaserPanel separation with event-driven communication
- **Legacy System Removal**: Complete elimination of old canvas system
- **Documentation Updated**: ARCHITECTURE.md v4.0 with comprehensive technical specifications

### Long-term Vision 🔮
- Full-featured web-based turn-based strategy platform
- Community-driven map and game creation
- Professional gaming experience with modern web technologies
- Template system usable for other turn-based games

---

**Last Updated**: 2025-01-20  
**Current Focus**: Component Architecture Cleanup and Technical Debt Reduction (Completed)  
**Next Milestone**: Component Integration Completion and Games Management System