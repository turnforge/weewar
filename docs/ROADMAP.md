# LilBattle Development Roadmap

## Overview
LilBattle is evolving from a comprehensive CLI-based turn-based strategy game into a full-featured web application template. This roadmap outlines the progression from game engine to web platform.

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
- [x] File-based storage system with `<LILBATTLE_DATA_ROOT>/storage/maps/<mapId>/` structure
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

### Visual Map Editor ✅
- [x] Canvas-based map editor with Phaser.js integration
- [x] Hexagonal grid rendering with proper hex math
- [x] Interactive tile and unit placement
- [x] Real-time visual feedback and layer management
- [x] Complete asset integration (tiles, units, sprites)

## ✅ Phase 4: Interactive Gameplay System (Completed August 2025)
**Status**: MAJOR MILESTONE COMPLETED  
**Timeline**: August 2025

### Core Gameplay Loop ✅
- [x] **Unit Movement Pipeline**: Complete click-to-move functionality with server validation
- [x] **Visual Game Interface**: Phaser.js game viewer with hex grid and sprites
- [x] **Move Execution**: ProcessMoves API integration with real-time visual updates
- [x] **State Management**: GameState controller with event-driven architecture
- [x] **Server Persistence**: SingletonGame maintains unit positions across moves

### WASM Client Architecture Revolution ✅
- [x] **90% Code Reduction**: Migrated from 374-line manual WASM to ~20-line generated client
- [x] **Type Safety**: Generated TypeScript interfaces with auto-conversion capabilities
- [x] **Direct Property Access**: Eliminated oneof complexity (`change.unitMoved` vs `change.changeType.case`)
- [x] **Action Object Pattern**: Server provides ready-to-use action objects, zero client reconstruction
- [x] **Service Reuse**: Same implementations work across HTTP, gRPC, and WASM transports
- [x] **Auto-Generation**: API changes automatically propagate to Go and TypeScript

### Technical Achievements ✅
- [x] **Event System**: GameState → GameViewer → Phaser update pipeline with granular events
- [x] **Highlight System**: Visual feedback for movement options and unit selection  
- [x] **Error Resolution**: Fixed BigInt serialization, oneof field handling, nil pointer issues
- [x] **Build Pipeline**: Successful TypeScript compilation with zero errors
- [x] **Protobuf Integration**: Seamless protobuf-es to Go protobuf serialization compatibility  
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

## ⚡ Phase 7.0: Comprehensive UI Architecture & Game Foundation (Completed)
**Status**: Completed  
**Timeline**: January 2025

### EventBus Architecture Completion ✅
- [x] **Lifecycle-Based Component System**: Template-scoped event binding for dynamic UI components in dockview containers
- [x] **EventBus Communication**: Type-safe, loosely-coupled component interaction with source filtering and error isolation
- [x] **Defensive Programming**: Robust state management with graceful error handling and automatic recovery mechanisms
- [x] **Observer Pattern Integration**: Unified Map and PageState architecture for reactive updates across all components
- [x] **Template-Scoped Event Binding**: Dynamic UI components work properly in layout systems without global conflicts

### Map Editor Polish ✅
- [x] **Unit Toggle Behavior**: Same unit+player removes unit, different unit/player replaces with intelligent tile placement
- [x] **City Tile Player Ownership**: Fixed city terrain rendering with proper player colors and ownership controls
- [x] **Reference Image System**: Complete scale and position controls with horizontal switch UI and mode visibility
- [x] **Per-Tab Number Overlays**: N/C/U keys toggle overlays per tab with persistent state management
- [x] **Auto-Tile Placement**: Units automatically place grass tiles when no terrain exists for seamless editing

### Backend Integration ✅
- [x] **Maps Delete Endpoint**: Complete DELETE /maps/{mapId} with proper HTTP method routing and redirects
- [x] **Web Route Architecture**: Clean HTTP method handling with proper REST semantics and comprehensive error handling
- [x] **Service Layer Integration**: Full integration with existing MapsService and file storage backend
- [x] **Frontend Error Resolution**: Fixed HTMX delete button integration with backend endpoints and proper form handling

### Technical Architecture ✅
- [x] **Pure Observer Pattern**: All map changes go through Map class with Phaser updates via EventBus notifications
- [x] **Event Delegation Pattern**: Robust button handling that works within dockview and complex layout systems
- [x] **Error Recovery Systems**: Comprehensive error handling with user feedback and graceful degradation
- [x] **Component Encapsulation**: Each component owns its DOM elements with proper lifecycle and state management
- [x] **State Management**: Proper toggle state tracking with visual feedback and EventBus communication patterns

### Benefits Achieved ✅
- **Production-Ready UI Framework**: Complete component architecture ready for game mechanics implementation
- **Robust Error Handling**: Comprehensive error recovery and user feedback systems throughout the application
- **Clean Component Boundaries**: Proper encapsulation with clear ownership of DOM elements and state
- **Scalable Architecture**: EventBus and Observer patterns provide foundation for complex multiplayer features
- **Professional UX**: Polished editor with intuitive controls and seamless interaction patterns

## ✅ Phase 8: Game Mechanics Implementation (Completed)
**Status**: Rules Engine Integration Complete, CLI Production Ready  
**Timeline**: January 2025

### Foundation Discovery ✅
**Major Finding**: Comprehensive game engine already exists in lib/game.go and cmd/lilbattle-cli/
- **Complete Game Class**: Turn management, movement, combat, save/load, events
- **Professional CLI**: Full command interface with 15+ game commands
- **Coordinate System**: Proper AxialCoord (cube coordinates) throughout
- **Multiplayer Ready**: Player validation, turn cycling, victory conditions

### Rules Engine Integration ✅ COMPLETED
- [x] ~~Create RulesEngine struct to load lilbattle-data.json~~ ✅ COMPLETED
- [x] ~~Replace hardcoded movement costs with terrain-specific calculations~~ ✅ COMPLETED
- [x] ~~Replace simple damage with probability-based combat from attack matrices~~ ✅ COMPLETED
- [x] ~~Update CLI commands to work with data-driven rules~~ ✅ COMPLETED
- [x] ~~Add rule validation and unit stats commands~~ ✅ COMPLETED

### CLI Architecture Revolution ✅ COMPLETED  
- [x] ~~Complete CLI transformation from bloated 1785-line implementation to focused 500-line SimpleCLI~~ ✅ COMPLETED
- [x] ~~Position/Unit Parser System supporting unit IDs (A1, B12), Q/R coordinates, row/col coordinates~~ ✅ COMPLETED
- [x] ~~Essential game commands: move, attack, select, end, status, units, player, help, quit~~ ✅ COMPLETED
- [x] ~~Move recording system with serializable MoveList and JSON export~~ ✅ COMPLETED
- [x] ~~REPL Interactive Mode for persistent gameplay without reloading~~ ✅ COMPLETED
- [x] ~~World loading integration from <LILBATTLE_DATA_ROOT>/storage/maps/ with JSON parsing~~ ✅ COMPLETED
- [x] ~~Complete USER_GUIDE.md with examples, troubleshooting, and command reference~~ ✅ COMPLETED

### Production Quality CLI Features ✅ COMPLETED
- [x] ~~Real world integration: Successfully loads and plays with <LILBATTLE_DATA_ROOT>/storage/maps/small-world~~ ✅ COMPLETED
- [x] ~~Rules engine integration: Proper initialization with rules-data.json~~ ✅ COMPLETED
- [x] ~~Game state persistence: Complete game state maintained across commands~~ ✅ COMPLETED
- [x] ~~Position parser flexibility: Handles A1-Z99, Q/R coordinates, row/col formats~~ ✅ COMPLETED
- [x] ~~Command recording: Full session recording with timestamps and player tracking~~ ✅ COMPLETED
- [x] ~~Unix-friendly architecture: Pipe-to-REPL pattern for batch operations~~ ✅ COMPLETED
- [x] ~~Comprehensive testing: Successfully compiled and tested with actual world data~~ ✅ COMPLETED

### Development Workflow Benefits ✅ COMPLETED
- [x] ~~Headless testing platform for CI/CD integration~~ ✅ COMPLETED
- [x] ~~Game state debugging and interactive rule validation~~ ✅ COMPLETED
- [x] ~~Map testing platform: Load any stored map for immediate testing~~ ✅ COMPLETED
- [x] ~~Move validation: Real-time feedback on valid/invalid moves~~ ✅ COMPLETED
- [x] ~~Session recording: Capture scenarios for documentation and bug reproduction~~ ✅ COMPLETED
- [x] ~~Complete documentation: USER_GUIDE.md with workflow patterns~~ ✅ COMPLETED

## 🤖 Phase 9: AI Player System (Completed)
**Status**: AI Toolkit Complete  
**Timeline**: January 2025

### AI Architecture Foundation ✅
- [x] **Comprehensive AI Toolkit**: Complete stateless AI helper library in `lib/ai/` 
- [x] **AIAdvisor Interface**: Core interface for move suggestions, position evaluation, threats, and opportunities
- [x] **Multiple Difficulty Levels**: Easy (Random + Avoidance), Medium (Greedy + Combat Prediction), Hard (Multi-turn Planning), Expert (Minimax + Alpha-Beta)
- [x] **AI Personality System**: Configurable weights for Aggressive, Defensive, Balanced, and Expansionist play styles
- [x] **Position Evaluator**: Comprehensive position analysis with material, economic, tactical, and strategic components

### Technical Implementation ✅
- [x] **BasicAIAdvisor**: Complete implementation supporting all difficulty levels with strategy pattern
- [x] **Position Evaluation System**: Multi-component evaluation with configurable weights and personality-specific tuning
- [x] **Decision Strategies**: Four distinct algorithms from random selection to minimax search with optimization
- [x] **Threat and Opportunity Analysis**: Advanced game state analysis for defensive and offensive decision making
- [x] **Performance Optimization**: Caching systems, transposition tables, and alpha-beta pruning for Expert level

### AI Integration Design ✅
- [x] **Stateless Architecture**: AI helpers analyze any game state without maintaining internal state
- [x] **Flexible Integration**: Works with CLI, web interface, or any UI layer through simple API calls
- [x] **Human Enhancement**: AI suggestions can assist human players or provide move hints
- [x] **Multiple AI Coexistence**: Different AI personalities can analyze the same game state simultaneously
- [x] **Game Engine Integration**: Leverages existing Game methods, RulesEngine, and combat prediction systems

### Documentation and Architecture ✅
- [x] **Comprehensive ARCHITECTURE.md**: Complete design documentation with implementation details and usage examples
- [x] **Performance Analysis**: Complexity analysis and optimization strategies for each difficulty level
- [x] **Extension Framework**: Clear guidelines for adding new AI personalities and evaluation metrics
- [x] **Integration Examples**: AI vs AI games, human assistance, and multiple AI analysis patterns

### AI Capabilities ✅
- [x] **Move Suggestions**: Primary move recommendations with alternatives and detailed reasoning
- [x] **Position Evaluation**: Comprehensive analysis with material, economic, tactical, and strategic scores  
- [x] **Threat Detection**: Identification of immediate dangers with severity assessment and solution suggestions
- [x] **Opportunity Recognition**: Discovery of tactical advantages with value assessment and execution requirements
- [x] **Strategic Analysis**: Long-term position assessment with strengths, weaknesses, and key factors

## 🎯 Phase 11: WASM Architecture Modernization (In Progress)
**Status**: Generated WASM Architecture Migration
**Timeline**: January 2025

### Generated WASM Architecture Discovery ✅
- [x] **Buf Plugin Analysis**: Complete understanding of protoc-gen-go-wasmjs generated files
- [x] **Generated Go WASM**: `gen/wasm/lilbattle_v1_services.wasm.go` with proper service injection pattern
- [x] **Generated TypeScript Client**: `web/frontend/gen/wasm-clients/lilbattle_v1_servicesClient.client.ts` with type-safe APIs
- [x] **Migration Path Identification**: Clear path from 374-line manual WASM to ~20-line dependency injection
- [x] **Service Integration Strategy**: Existing services can be directly injected into generated exports

### Current Manual WASM Issues ✅ IDENTIFIED
- [x] **Massive Boilerplate**: 374 lines of repetitive validation and response formatting
- [x] **Type Unsafety**: Manual `js.Value` conversions and `any` types throughout
- [x] **Code Duplication**: Service logic reimplemented in WASM instead of reusing existing services
- [x] **Testing Difficulties**: Global game state prevents proper unit testing
- [x] **Maintenance Burden**: API changes require manual updates in multiple places

### New Architecture Benefits ✅ ANALYZED
- [x] **90% Code Reduction**: Manual bindings → dependency injection wrapper
- [x] **Type Safety**: Protobuf types throughout, no `any`/`js.Value` conversions
- [x] **Service Reuse**: Same implementations work for HTTP, gRPC, and WASM transports
- [x] **Auto-Generation**: API changes automatically propagate to Go and TypeScript
- [x] **Standard Patterns**: Follows established gRPC/Connect conventions
- [x] **Testability**: Service mocks enable proper unit testing

### Implementation Plan (Next Phase)
- [ ] **Service Wiring**: Update `cmd/lilbattle-wasm/main.go` to use generated exports with service injection
- [ ] **Build Integration**: Update Makefile to build generated WASM instead of manual approach
- [ ] **Frontend Migration**: Replace GameState.ts manual WASM calls with generated TypeScript client
- [ ] **Legacy Cleanup**: Remove obsolete manual WASM binding code

### Generated Architecture Pattern
```go
// New main.go (~20 lines)
func main() {
    exports := &lilbattle_v1_services.LilBattle_v1_servicesServicesExports{
        GamesService:  services.NewGamesService(store),
        UsersService:  services.NewUsersService(store), 
        WorldsService: services.NewWorldsService(store),
    }
    exports.RegisterAPI()
    select {} // Keep running
}
```

```typescript
// New frontend (~50 lines)
const client = new LilBattle_v1_servicesClient();
await client.loadWasm();
const response = await client.gamesService.createGame(request);
```

## ✅ Phase 10: Interactive Web Gameplay (Mostly Complete)
**Status**: WASM Integration Complete, Unit Interaction In Progress  
**Timeline**: January 2025

### GameViewerPage Foundation ✅ COMPLETED
- [x] ~~Complete GameViewerPage architecture with lifecycle controller and WASM bridge~~ ✅ COMPLETED
- [x] ~~External orchestration pattern with breadth-first component initialization~~ ✅ COMPLETED
- [x] ~~ComponentLifecycle interface with multi-phase initialization~~ ✅ COMPLETED
- [x] ~~GameState component with async WASM loading and synchronous operations~~ ✅ COMPLETED
- [x] ~~Game control UI with turn management, unit selection, and game log~~ ✅ COMPLETED
- [x] ~~StartGamePage integration with URL-based configuration~~ ✅ COMPLETED

### WASM Integration Completion ✅ COMPLETED
- [x] ~~**Debug WASM Loading Issues**: Fix WASM path resolution and module loading in GameState component~~ ✅ COMPLETED
- [x] ~~**Resolve World Data Loading**: Fix null world data in GameViewerPage and template integration~~ ✅ COMPLETED
- [x] ~~**Fix Initialization Sequence**: Prevent multiple initialization calls and race conditions~~ ✅ COMPLETED
- [x] ~~**Test WASM API Integration**: Verify all WASM functions (createGameFromMap, selectUnit, moveUnit, attackUnit, endTurn) work~~ ✅ COMPLETED
- [x] ~~**Complete Data Format Alignment**: Ensure WASM responses match TypeScript interface expectations~~ ✅ COMPLETED

### Interactive Gameplay Features (Current Priority - In Progress)
- [x] ~~**Terrain Click System Working**: Fixed rules engine integration and WASM exports for terrain stats~~ ✅ COMPLETED
- [x] ~~**PhaserGameScene Implementation**: Extended PhaserWorldScene with unit selection, movement, and attack highlighting~~ ✅ COMPLETED
- [x] ~~**WASM Export Resolution**: Added missing getTerrainStatsAt, canSelectUnit, getTileInfo functions~~ ✅ COMPLETED
- [x] ~~**Rules Engine Integration Fix**: Fixed UI methods to use game's rules engine instead of global empty instance~~ ✅ COMPLETED
- [x] ~~**Callback System Working**: Fixed timing issues with callback setup after scene initialization~~ ✅ COMPLETED
- [ ] **Unit Movement Integration**: Connect WASM GetMovementOptions to PhaserGameScene green highlighting system
- [ ] **Unit Attack Integration**: Connect WASM GetAttackOptions to PhaserGameScene red highlighting system
- [ ] **Click-to-Move Implementation**: Enable clicking highlighted movement tiles to execute unit moves
- [ ] **Click-to-Attack Implementation**: Enable clicking highlighted attack targets to execute combat
- [ ] **Complete Turn Management**: Test full turn cycle with proper UI updates and state synchronization

### GameLog System Implementation ✅ COMPLETED
- [x] ~~**GameLog Architecture Design**: Comprehensive design with Go-centric architecture and SaveHandler interface pattern~~ ✅ COMPLETED
- [x] ~~**Core GameLog System**: GameAction, WorldChange, GameLogEntry, GameSession data structures~~ ✅ COMPLETED
- [x] ~~**Simplified SaveHandler Interface**: Only Save() method needed - UI handles loading directly~~ ✅ COMPLETED
- [x] ~~**Game Integration**: GameLog integrated into Game struct with automatic recording~~ ✅ COMPLETED
- [x] ~~**Action Recording**: MoveUnit, AttackUnit, NextTurn all automatically logged with detailed change tracking~~ ✅ COMPLETED
- [x] ~~**WASM Integration**: saveGame() and loadGame() functions exposed to frontend~~ ✅ COMPLETED
- [x] ~~**Frontend Integration**: GameState.ts save/load methods with JavaScript bridge functions~~ ✅ COMPLETED
- [x] ~~**BrowserSaveHandler**: Moved to WASM package, directly instantiated by game creators~~ ✅ COMPLETED
- [x] ~~**Architecture Simplification**: Removed Load/List/Delete methods, factory pattern, complex session management~~ ✅ COMPLETED

### AI Integration and Advanced Features (Week 3-4 - Future)
- [ ] **AI Integration with Web Interface**: Add WASM bindings for AI toolkit (`lib/ai/` package)
- [ ] **AI Web UI**: Create web interface for AI difficulty and personality selection
- [ ] **AI Move Suggestions**: Implement AI move hints and analysis in game interface
- [ ] **AI vs AI Games**: Add AI vs AI mode with visualization and speed controls
- [ ] **Advanced Game Controls**: Undo/redo, center camera, show all units, game history

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

## Phase 12: Asset Provider Architecture (Completed)
**Status**: Production-ready  
**Timeline**: Completed January 2025

### AssetProvider System Implementation ✅
- [x] **Interface-based Asset Loading**: Clean AssetProvider interface for swappable asset packs
- [x] **PNGAssetProvider**: Default provider for existing PNG assets with proper player color handling
- [x] **TemplateSVGAssetProvider**: Advanced provider with SVG template processing and color replacement
- [x] **Dynamic Asset Resolution**: Support for different rasterization sizes (160px default, configurable)
- [x] **Fallback System**: Automatic PNG fallback when SVG assets are unavailable

### Technical Architecture ✅
- [x] **Clean Separation**: WorldScene focuses on rendering, AssetProvider handles asset management
- [x] **Format Agnostic**: Support for PNG, SVG, or procedurally generated assets
- [x] **Post-Processing Pipeline**: Template variable replacement for player colors and gradients
- [x] **Memory Optimization**: Configurable raster sizes for balancing quality vs memory (0.5x-4x zoom)
- [x] **URL Parameter Configuration**: Easy testing with ?useSVG=true&svgSize=256

### Benefits Achieved ✅
- **Asset Pack Swapping**: Easy switching between different art styles or resolutions
- **Reduced File Count**: SVG templates eliminate need for 13 color variants per asset
- **Dynamic Theming**: Player colors can be changed at runtime
- **Better Scaling**: SVG rasterization at higher resolutions improves zoom quality
- **Testable Architecture**: Clean interface allows mocking for tests

## ✅ Phase 13: Incremental UI Updates & Turn Management (Completed)
**Status**: Production-ready
**Timeline**: Completed January 2025

### Incremental Update Architecture ✅
- [x] **SetTileAt/SetUnitAt Methods**: Direct incremental updates to World without full refresh
- [x] **RemoveTileAt/RemoveUnitAt Methods**: Explicit entity removal with clear intent
- [x] **UpdateGameStatus Method**: Separate UI status updates for player turn and turn counter
- [x] **World Transaction Layer Fix**: GetPlayerUnits properly falls back to parent layer for transactional operations
- [x] **Presenter Delta Processing**: applyIncrementalChanges processes WorldChange events from ProcessMoves

### End Turn Functionality ✅
- [x] **EndTurnButtonClicked RPC**: Wire up End Turn button to presenter
- [x] **End Turn Button State Management**: Enable/disable based on current player turn
- [x] **Turn Transition Updates**: Automatic UI updates when player changes
- [x] **Unit Reset Handling**: PlayerChanged events properly reset all units for new turn

### UI Simplification ✅
- [x] **GameActionsPanel Removal**: Eliminated unnecessary panel for cleaner UI
- [x] **Direct Button Wiring**: End Turn button directly wired to presenter
- [x] **Streamlined Layout**: Simplified dockview layout with fewer panels

### CLI Enhancement ✅
- [x] **Direction Shortcuts**: Support for L, R, TL, TR, BL, BR in move/attack commands
- [x] **ParseDirection Function**: Parse direction strings to NeighborDirection enum
- [x] **Context-Aware Parsing**: ParsePositionOrUnitWithContext for relative position resolution
- [x] **Multiple Naming Conventions**: Support TL/UL/LU variations for user convenience
- [x] **Updated Help Text**: Comprehensive documentation of direction shortcuts

### Technical Achievements ✅
- **Efficient Updates**: No more full SetGameState calls after moves - only deltas
- **EventBus Integration**: World changes trigger EventBus events for automatic UI updates
- **Transaction Safety**: World layering properly supports transactional game operations
- **User Experience**: Faster, more intuitive CLI with hex neighbor shortcuts
- **Clean Architecture**: Presenter drives all UI updates through well-defined methods

**Last Updated**: 2026-01-12
**Current Focus**: Monetization strategy and ad integration
**Recent Milestone**: Monetization strategy document complete

**Recent Achievements**:
- Google App Engine deployment configuration (app.yaml, .gcloudignore)
- Environment-based backend selection (CLI flag → env var → default)
- GCP service account setup with proper IAM roles
- OAuth2 configuration for Google authentication
- Privacy Policy and Terms of Service pages with proper templates
- Reusable Footer component for listing pages
- Template system modernization with templar and goapplib integration

## ✅ Phase 14: Google App Engine Deployment (Completed)
**Status**: Production-ready
**Timeline**: Completed January 2025

### Deployment Infrastructure ✅
- [x] **app.yaml Configuration**: Go 1.24 runtime with static file handlers
- [x] **.gcloudignore Setup**: Proper exclusions for dev files, node_modules, tests
- [x] **Environment Configuration**: LILBATTLE_ENV-based config loading (dev/production)
- [x] **Backend Selection**: Priority-based config resolution (CLI flag → env var → default)
- [x] **Secrets Management**: Symlinked configs folder to external secrets directory

### Authentication Setup ✅
- [x] **GCP Service Account**: lilbattle-cli-admin with App Engine and Datastore roles
- [x] **OAuth2 Configuration**: Google OAuth with proper callback URLs
- [x] **Multi-environment OAuth**: Support for localhost and production callback URLs
- [x] **X/Twitter OAuth 2.0**: Custom Twitter OAuth handler with PKCE support (required by Twitter API)

### Legal Pages ✅
- [x] **Privacy Policy Template**: Proper page structure with sections and styling
- [x] **Terms of Service Template**: Fixed template inheritance and section numbering
- [x] **Footer Component**: Reusable footer with Privacy/Terms links for listing pages

### Production Storage ✅
- [x] **GAE/Datastore Backend**: Complete Datastore backend implementation (gaebe package)
  - Proto definitions with datastore_tags annotations (`protos/lilbattle/v1/datastore/models.proto`)
  - Generated entity types and converters via protoc-gen-dal (`gen/datastore/`)
  - Composite indexes for needs_indexing queries (`index.yaml`)
  - Cross-entity transactions for game move atomicity
  - Backend selection via env vars: `WORLDS_SERVICE_BE=gae`, `GAMES_SERVICE_BE=gae`
- [x] **Local Backend Deployment**: File-based storage for development/testing

## ✅ Phase 15: Monetization (In Progress)
**Status**: Phase 1 Complete
**Timeline**: January 2026+

### Monetization Strategy ✅
- [x] **Competitive Analysis**: Analyzed BGA, Tabletopia, BGG, Yucata models
- [x] **Strategy Document**: Created MONETIZATION.md with phased approach
- [x] **Revenue Projections**: Estimated $420-$6,600/month based on DAU
- [x] **Technical Requirements**: Identified CSP, component, and integration needs

### Phase 1: Foundation Ads ✅
- [x] **Feature Flags**: LILBATTLE_ADS_ENABLED, LILBATTLE_ADS_FOOTER, LILBATTLE_ADS_HOME, LILBATTLE_ADS_LISTING, LILBATTLE_AD_NETWORK_ID
- [x] **AdSlot Component**: Reusable ad container with size variants (leaderboard, mrec, skyscraper, mobile-banner)
- [x] **AdScript Component**: AdSense script loader in BasePage head
- [x] **Footer Banner Ads**: 728x90 leaderboard (desktop), 320x50 banner (mobile)
- [x] **Homepage Mid-Section Ad**: 300x250 between quick actions and recent activity
- [x] **CSS Styles**: Responsive ad containers with dark mode and placeholder support
- [x] **CSP Updates**: Added Google AdSense domains to Content-Security-Policy
- [ ] **Listing Page Native Ads**: Styled like game/world cards (requires EntityListing changes)
- [ ] **Google AdSense Account Setup**: Account creation and publisher ID configuration

### Phase 2: Game Integration (Post-Launch)
- [ ] **Game End Screen**: Victory/defeat modal with interstitial ad
- [ ] **Turn Transition Ads**: Brief interstitial every 3-5 turns
- [ ] **Right Panel Ads**: Desktop sidebar placement

### Phase 3: Rewarded Ads (Future)
- [ ] **Bonus Coins**: Watch ad for +100 coins
- [ ] **Undo Move**: Watch ad to undo last action
- [ ] **Rewarded Video SDK**: IronSource or Unity Ads integration

### Phase 4: Premium Tier (Future)
- [ ] **Ad-Free Subscription**: $3-5/month to remove ads
- [ ] **Stripe Integration**: Payment processing
- [ ] **Premium Status**: User model and UI badges
