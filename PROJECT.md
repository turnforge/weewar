# LilBattle Project Documentation

## Overview

LilBattle is a turn-based strategy game built with Go backend, TypeScript frontend, and WebAssembly for high-performance game logic. The project implements a local-first multiplayer architecture where game validation happens in WASM clients with server-side coordination for consensus.

### Core Technologies
- **Backend**: Go with protobuf for game logic and coordination
- **Frontend**: TypeScript with Phaser for 2D hex-based rendering
- **Templates**: Templar engine with goapplib for shared components (namespace/include/extend)
- **Communication**: WebAssembly bridge for client-server interaction
- **Coordination**: TurnEngine framework for distributed validation
- **Build System**: Continuous builds with devloop for hot reloading

---

## Architecture

### Key Components

**Services Layer (`services/`)**
- **Core Game Logic**: World, Game, RulesEngine - Pure game state and runtime logic
- **Move Processing**: DefaultMoveProcessor validates and processes game moves with transaction support
- **Service Implementations**: BaseGamesServiceImpl, FSGamesService, SingletonGamesService
- **WASM Integration**: WasmGamesService for client-side game logic execution
- **Multiplayer Coordination**: CoordinatorGamesService with K-of-N validation

**Frontend (`web/`)**
- **GameState**: Lightweight controller managing WASM interactions
- **GameViewer**: Phaser-based view rendering hex maps and units
- **Event System**: Clean separation between game logic and UI updates
- **Splash Screen System**: Pre-load overlays with progress tracking

### Technical Highlights

**Distributed Validation**: Local-first architecture where each player's WASM validates moves with server coordinating K-of-N consensus without running game logic.

**Transaction Safety**: Parent-child transaction model with copy-on-write semantics for safe rollback and ordered change application.

**Service Reusability**: Same service implementations work across HTTP, gRPC, and WASM transports through interface abstraction.

**Centralized Rules Engine**: Proto-based single source of truth with O(1) lookup while eliminating data duplication.

---

## Current Status

**Core Gameplay**: ✅ **PRODUCTION READY**
- Complete unit movement, attack, build, capture systems
- Transaction-safe state management with copy-on-write
- Server-side persistence with atomic operations
- Formula-based combat with wound bonuses and splash damage
- Economy system with configurable terrain income

**CLI Tools**: ✅ **PRODUCTION READY**
- **ww**: Modern CLI with subcommands (status, units, options, move, attack, build, endturn)
- **lilbattle-cli**: Headless REPL for game state manipulation
- Direction shortcuts, unit shortcuts, JSON output mode
- Comprehensive diagnostics and debugging capabilities

**Multiplayer Coordination**: 🔄 **90% COMPLETE**
- Core coordination protocol implemented in TurnEngine
- File-based storage with atomic operations ready
- Testing and client integration pending

**UI & Polish**: 🔄 **ONGOING**
- World viewer/editor with reference images, shape tools
- Game viewer with multiple layout variants (dock, grid, mobile)
- Responsive design with mobile bottom sheets
- Feature flags for selective UI visibility
- Animation system with effects (projectile, explosion, heal, capture)

---

## Recent Achievements

### Templar Template Migration (2025-12-09)
- Migrated to templar template engine with namespace/include/extend directives
- Integrated goapplib for shared page components (BasePage, Header, etc.)
- Template inheritance: pages extend `goapplib/BasePage.html` with custom blocks
- Component templates use `.templar.html` extension for presenter-rendered panels
- Cleaner template organization with reduced duplication

### Feature Flags and Navigation System (2025-11-07)
- Environment-based flags (LILBATTLE_HIDE_GAMES, LILBATTLE_HIDE_WORLDS) for UI visibility control
- Unified navigation tabs in floating drawer (Games | Worlds | Profile)
- Consistent header pattern site-wide matching GameViewerPage architecture
- Active tab highlighting with BasePage.ActiveTab field
- Smart homepage redirect based on visible tabs
- Mobile responsive drawer with slide-down animation

### World Editor Enhancements
- Multi-click shape tools (Rectangle, Circle, Oval, Line) with preview system
- Reference image persistence with IndexedDB storage
- Automatic restore with correct scale/position on page reload
- Fill/Outline toggle for shape tools
- Clean separation: ReferenceImagePanel (loading) vs ReferenceImageLayer (display)

### Responsive Layout System
- Mobile-first design with bottom sheets and FABs
- Responsive header button system with automatic dropdown on mobile
- Touch-friendly interfaces across all major pages
- Context-aware button ordering in GameViewerPageMobile

### Combat and Economy Systems
- Formula-based combat: p = 0.05 * (((A + Ta) - (D + Td)) + B) + 0.5
- Wound bonus tracking with attack history
- Splash damage with air unit immunity and friendly fire
- Terrain-based income generation (configurable per terrain type)
- Starting coins configuration per player

### Action Progression System
- Index-based state machine replacing history-based tracking
- Pipe-separated alternatives (e.g., "attack|capture")
- Natural movement limiting via distance_left
- Action limits for repeated actions

### Visual Polish
- Animation framework: projectile arcs, explosions, heal bubbles, capture effects
- Path-following movement with segment-by-segment animation
- Unit lifecycle animations (appear, fade-out, flash)
- Exhausted units/tiles gray highlight system
- Smart batching for simultaneous effects

---

## Critical Bug Fixes

### Unit Shortcuts Being Lost (2025-10-25)
**Problem**: Shortcuts corrupted after moves/turn changes
**Solution**: Created copyUnit() helper to preserve all fields including Shortcut
**Files**: services/moves.go

### FSGamesService Cache Causing Stale Data (2025-10-25)
**Problem**: Browser showed stale state after CLI updates
**Solution**: Disabled cache in GetGame() - always read fresh from disk
**Files**: services/fsgames_service.go

### Unit Duplication Bug (2025-10-22)
**Problem**: Units appearing at both old and new positions after moves
**Solution**: Implemented copy-on-write in World.MoveUnit()
**Files**: services/world.go, services/moves.go

### Lazy Top-Up Bug (2025-10-24)
**Problem**: Units not moving despite having movement options
**Solution**: Added topUpUnitIfNeeded() calls before validation
**Files**: services/moves.go

---

## Known Issues

### CLI Options Command Async Updates
**Status**: 🚧 KNOWN ISSUE
**Severity**: Medium
**Problem**: Options command returns before panels populated (async goroutines)
**Proposed**: Create Cmd panel versions with channel-based callbacks

### Visual Updates Use Full Reload
**Status**: 🚧 MINOR POLISH
**Severity**: Low
**Problem**: Entire scene reloaded instead of targeted updates
**Impact**: Slightly slower rendering, functionally correct

---

## Next Steps

### Phase 2: Multiplayer Integration (High Priority)

**Coordinator Testing**:
- [ ] Unit tests for coordinator consensus logic
- [ ] Manual test CLI for local multiplayer simulation
- [ ] WASM client updates to compute ExpectedResponse locally
- [ ] UI indicators for proposal status (pending/validating/accepted/rejected)
- [ ] End-to-end multiplayer gameplay testing

### UI Polish & Features (Medium Priority)

**World Editor**:
- [ ] Implement flood fill algorithm with radial limit
- [ ] Add visual feedback for shape tools (cursor, anchors, status text)
- [ ] Context-aware Clear button (brush vs flood clear)
- [ ] Keyboard shortcuts for tools

**Gameplay**:
- [ ] Attack UI interactions in browser (clickable targets)
- [ ] Combat animations and visual effects
- [ ] Victory conditions and game over screen
- [ ] Building capture mechanics
- [ ] Fog of war implementation

**Mobile Support**:
- [ ] Touch controls for unit selection and movement
- [ ] Touch gesture support (pinch to zoom, pan)
- [ ] Mobile performance optimization

### Phase 3: Production Readiness (High Priority)

**Database Migration**:
- [ ] Design PostgreSQL schema for games, states, history, proposals
- [ ] Implement DatabaseGamesService
- [ ] Migration utilities from file storage to database
- [ ] Connection pooling and performance optimization

**Real-Time Updates**:
- [ ] WebSocket server implementation
- [ ] Client WebSocket connection management
- [ ] Real-time game state push notifications
- [ ] Player presence and online status

**Performance**:
- [ ] Load testing with concurrent games
- [ ] Profiling and optimization of hot paths
- [ ] Sub-100ms move processing target
- [ ] 60fps rendering target for Phaser scenes

### Testing & Quality (High Priority)

- [ ] Unit tests for all move processor functions
- [ ] Integration tests for multiplayer coordination
- [ ] End-to-end browser tests with Playwright
- [ ] Error scenario testing (network failures, invalid moves)
- [ ] Performance regression tests

### Documentation (Medium Priority)

**User Docs**:
- [ ] Complete user guide for CLI tool
- [ ] Browser UI tutorial and walkthrough
- [ ] Map editor guide
- [ ] Game rules and mechanics documentation

**Developer Docs**:
- [ ] API documentation for services
- [ ] WASM integration guide
- [ ] Contributing guidelines
- [ ] Development environment setup guide

---

## Future Considerations (Phase 4+)

- Replay system with move history playback
- Spectator mode for ongoing games
- Tournament system and matchmaking
- Player rankings and statistics
- Custom game modes and modding support
- AI opponents (foundation exists in services/ai/)
- Accessibility improvements (screen reader, keyboard nav, high contrast)
- Cross-platform mobile apps

---

## Debugging Quick Reference

### Check Unit Shortcuts
```bash
jq '.world_data.units[] | {shortcut, q, r}' ~/dev-app-data/lilbattle/storage/games/{gameId}/state.json
```

### Check Unit Movement Points
```bash
ww --verbose options B1 | grep "DistanceLeft"
jq '.world_data.units[] | {shortcut, distance_left, last_topped_up_turn}' state.json
```

### View Recent Moves
```bash
jq '.groups[-1]' ~/dev-app-data/lilbattle/storage/games/{gameId}/history.json
```

### Check Game Status
```bash
jq '{current_player, turn_counter, status}' ~/dev-app-data/lilbattle/storage/games/{gameId}/state.json
```

### List Units for Player
```bash
jq '.world_data.units[] | select(.player == 1) | {shortcut, q, r, moves: .distance_left}' state.json
```

---

## Game Storage Structure

**Games**: `~/dev-app-data/lilbattle/storage/games/{gameId}/`
- metadata.json: Game configuration (players, teams, settings, world_id)
- state.json: Current game state (tiles, units, current_player, turn_counter)
- history.json: Move history (groups of moves with results)

**Worlds**: `~/dev-app-data/lilbattle/storage/worlds/{worldId}/`
- metadata.json: World metadata (name, description, creator)
- world.json: Map data (tiles, starting units)

**Proto Field Naming**: Snake_case in JSON but camelCase in Go (e.g., available_health → AvailableHealth)

---

## CLI Command Reference

```bash
export LILBATTLE_GAME_ID=<gameId>  # Or use --game-id flag

ww status                    # Show game state (players, coins, units, tiles)
ww units                     # List all units
ww options B1                # Show available moves for unit B1
ww options t:A1              # Show build options for tile A1
ww move B1 0,-3             # Move unit by coordinates
ww move B1 R                # Move unit by direction (L/R/TL/TR/BL/BR)
ww attack A1 B2             # Attack unit
ww build t:A1 trooper       # Build a unit at tile A1
ww build t:A1 5             # Build unit type 5 at tile A1
ww endturn                  # End current player's turn

# Flags
ww --verbose units          # Show debug output
ww --dryrun move B1 R      # Preview move without saving
ww --confirm=false build t:A1 tank  # Build without confirmation
ww --json status            # Output as JSON
```

---

## Configuration

### Feature Flags (.env)
```bash
# Hide games UI and navigation (APIs remain accessible)
LILBATTLE_HIDE_GAMES=true

# Hide worlds UI and navigation (APIs remain accessible)
LILBATTLE_HIDE_WORLDS=false
```

### Development (.env.dev)
```bash
LILBATTLE_BASE_URL=http://localhost:8080
ONEAUTH_JWT_SECRET_KEY=DevSecretKey123456789012345678901234567890
```
