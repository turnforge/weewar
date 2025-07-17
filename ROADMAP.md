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
**Timeline**: Completed 2025-01-14

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

## 📋 Phase 5: Games Management System (Planned)
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

## 🎯 Phase 5: Gameplay Integration (Planned)
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

## 🔮 Phase 6: Platform Features (Future)
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

**Last Updated**: 2025-01-17  
**Current Focus**: Phaser.js Map Editor (Completed)  
**Next Milestone**: WASM integration with Phaser editor for data persistence