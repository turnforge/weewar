# Next Steps - WeeWar Development

## Phase 2: Multiplayer Coordination Integration

### Coordinator Testing and Integration
**Priority**: High
**Status**: 90% Complete - Testing Needed

**Completed**:
- Core coordination protocol in TurnEngine
- File-based storage with atomic operations
- Callback architecture (OnProposalStarted/Accepted/Failed)
- CoordinatorGamesService wrapping FSGamesService

**Remaining Tasks**:
- [ ] Unit tests for coordinator consensus logic
- [ ] Manual test CLI for local multiplayer simulation
- [ ] WASM client updates to compute ExpectedResponse locally
- [ ] WASM client integration with coordinator service
- [ ] UI indicators for proposal status (pending/validating/accepted/rejected)
- [ ] Validator service implementation for independent validation
- [ ] End-to-end multiplayer gameplay testing
- [ ] Documentation for multiplayer setup and testing

---

## UI Polish & User Experience

### CLI Tool Enhancements
**Priority**: Medium
**Status**: ✅ COMPLETE

**Completed**:
- [x] Tile prefix parsing (t:A1, t:0,-3) for disambiguating tiles from units
- [x] Enhanced status command showing player coins, unit counts, tile counts
- [x] Position parser supports all coordinate formats with tile prefix
- [x] JSON output for all CLI commands with rich player information

**Design**:
- `t:` prefix forces tile lookup instead of unit (e.g., `ww options t:A1`)
- Status command now provides comprehensive game overview
- ParseTarget struct tracks ForceTile flag and provides GetTile() method

### Dashboard Homepage
**Priority**: Medium
**Status**: ✅ COMPLETE

**Completed**:
- [x] Hero section with welcome message and tagline
- [x] Stats cards showing total games and worlds with gradient backgrounds
- [x] Quick action buttons for "Start New Game" and "Create World"
- [x] Two-column layout with recent games and worlds (up to 6 each)
- [x] Screenshot thumbnails in activity cards
- [x] Empty states with helpful CTAs when no content exists
- [x] Full responsive design and dark mode support

**Design**:
- Visual dashboard replacing plain text links on homepage
- Leverages screenshot system for visual preview cards
- Immediate activity overview on landing

### World Selection & Listing System
**Priority**: Medium
**Status**: ✅ COMPLETE

**Completed**:
- [x] WorldFilterPanel reusable component (search, sort, filter)
- [x] WorldGrid component with responsive card layout
- [x] WorldList unified component (table and grid views)
- [x] SelectWorldPage for streamlined game creation workflow
- [x] View mode toggle (table/grid) with query param persistence
- [x] Action modes (manage vs select) for different contexts
- [x] Pagination support in both table and grid views
- [x] Auto-redirect from /games/new to /worlds/select
- [x] Responsive design with mobile bottom sheets

**Design**:
- Configurable listing component reused across WorldListingPage and SelectWorldPage
- "Manage" mode: Edit/Delete/Start Game actions via dropdown menu
- "Select" mode: Large Play buttons for quick game creation
- Smart defaults: Grid for selection, table for management
- Seamless user flow: /games/new → /worlds/select → click Play → game config

### World Editor Reference Image System
**Priority**: Medium
**Status**: ✅ COMPLETE

**Completed**:
- [x] ReferenceImageLayer integration into PhaserEditorScene
- [x] Layer-based input routing for independent drag/scroll handling
- [x] Overlay mode for aligning reference image over tiles without moving map
- [x] Background mode for reference image moving with world coordinates
- [x] Event emission for UI synchronization during interactive transformations
- [x] Circular reference prevention with value comparison guards
- [x] LayerManager extensions (processClick, processDrag, processScroll)

**Design**:
- Overlay mode (depth 1000): Interactive layer blocks camera events, allows independent drag/scale
- Background mode (depth -1): Non-interactive layer follows world coordinates
- Reference image stays in world coordinates when switching modes (no position jumping)
- Scene events bridge layer state changes to EventBus for ReferenceImagePanel updates
- Layer system provides clean separation of input handling vs rendering

### Screenshot and Preview System
**Priority**: Medium
**Status**: ✅ COMPLETE

**Completed**:
- [x] Screenshot capture functionality in PhaserWorldScene with tight bounds clipping
- [x] Screenshot API endpoints for games and worlds
- [x] Screenshot buttons in GameViewerPage and WorldEditorPage
- [x] Screenshot thumbnails in GameList and WorldList pages
- [x] screenshot_url field in Game and World protos for flexible URL management
- [x] Default URL population with CDN-ready override capability
- [x] Fixed infinite canvas height growth by removing automatic ResizeObserver

**Design**:
- Screenshots stored in `~/dev-app-data/weewar/storage/{games|worlds}/{id}/screenshots/{screenshotName}`
- Automatic tight bounds calculation to avoid empty space
- Generic screenshot handler for reusability across resource types
- Proto field allows easy migration to CDN or external hosting
- Canvas sizing now controlled by CSS/parent page to avoid circular dependencies

### Visual Updates and Animations
**Priority**: Medium
**Status**: ✅ COMPLETE

**Completed**:
- [x] Animation framework with presenter-driven architecture
- [x] Promise-based animation API for sequencing and chaining
- [x] Smart batching for simultaneous effects (splash damage)
- [x] Configurable timing with instant mode support
- [x] Effect classes: ProjectileEffect, ExplosionEffect, HealBubblesEffect, CaptureEffect
- [x] Scene API: moveUnit(), showAttackEffect(), showHealEffect(), showCaptureEffect()
- [x] Unit lifecycle animations: setUnit (flash/appear), removeUnit (fade-out)
- [x] Particle system with runtime-generated textures
- [x] Path animation for smooth unit movement along hex paths
- [x] Attack sequences: attacker flash, projectile arc, impact explosions
- [x] Presenter integration with applyIncrementalChanges for WorldChange events
- [x] BrowserGameScene overrides forwarding animations to browser Phaser scene
- [x] Unit appear animation: scale bounce (0.5 → 1.5 → 1) instead of fade-in
- [x] Exhausted units/tiles gray highlight system (depth 13, above all other elements)
- [x] Selective highlight clearing to preserve exhausted state during interactions
- [x] Path-following movement animation with pauses at each tile (2025-01-05)
- [x] Full pathfinding route extraction from MoveUnitAction.reconstructed_path (2025-01-05)
- [x] Segment-by-segment animation with configurable pause timing (2025-01-05)

**Remaining Tasks**:
- [ ] Add loading states during move processing
- [ ] Prevent concurrent move submissions
- [ ] Add sound effects (moves, attacks, selections)
- [ ] Add retreat animations
- [ ] Add capture animations with building color changes

---

## Gameplay Features

### Combat System Completion
**Priority**: High
**Status**: Formula-based combat complete, UI polish needed

**Completed**:
- [x] Formula-based combat system with dice rolling simulation
- [x] Wound bonus tracking and accumulation within turns
- [x] Splash damage implementation with air unit immunity
- [x] Attack history management (cleared on turn change)
- [x] CLI attack command with detailed diagnostics
- [x] Comprehensive test coverage for combat formulas
- [x] Attack simulator tool at /rules/attacksim (Monte Carlo simulation with damage distributions)
- [x] Action progression system with index-based state machine

**Remaining Tasks**:
- [ ] Attack UI interactions in browser (clickable attack targets)
- [ ] Combat animations and visual effects
- [ ] Victory conditions (last player with units wins)
- [ ] Game over screen and restart flow

### Action Progression System
**Priority**: High
**Status**: ✅ COMPLETE

**Completed**:
- [x] Refactored from history-based to index-based state machine
- [x] Unit progression tracking via progression_step (index into action_order)
- [x] Pipe-separated alternatives support (e.g., "attack|capture")
- [x] Natural movement limiting via distance_left
- [x] Chosen alternative tracking for mutually exclusive actions
- [x] Integration with ProcessMoveUnit, ProcessAttackUnit, GetOptionsAt
- [x] Turn reset via TopUpUnitIfNeeded
- [x] Comprehensive test coverage
- [x] HTML extraction for action_order from WeeWar unit pages
- [x] Rules data split into core rules (92KB) and damage distributions (1.2MB)

**Design**:
- Units track progression_step (0-based index into UnitDefinition.action_order)
- Default action_order: ["move", "attack|capture"]
- Movement advances step when distance_left reaches 0
- Pipe-separated alternatives are mutually exclusive (choosing one locks out others)
- action_limits supported for repeated actions (e.g., {"attack": 2} for double attacks)
- State resets to step 0 on turn change

**Data Extraction**:
- cmd/extract-rules-data now extracts action_order from Progression HTML sections
- Outputs two files: weewar-rules.json (core) and weewar-damage.json (damage distributions)
- RulesEngine loader updated to load both files independently
- Supports lazy loading - damage data only loaded when needed for combat

---

### Advanced Game Mechanics
**Priority**: Medium
**Status**: Economy System Complete with Configurable Income

**Completed**:
- [x] Unit production from buildings (build system)
- [x] Resource management system (coin deduction and validation)
- [x] Build validation (ownership, terrain compatibility, one-build-per-turn)
- [x] CLI build command with confirmation prompts
- [x] Web UI build options modal with unit stats and costs
- [x] ProcessBuildUnit with comprehensive validations
- [x] UnitBuiltChange tracking in world history
- [x] Economy system with coin deduction on build
- [x] Income generation on end turn based on terrain type
- [x] CoinsChangedChange WorldChange event tracking
- [x] FSGamesService fix to persist GameConfig changes
- [x] Move command error propagation (no false success messages)
- [x] Terrain-based income system (income_per_turn field in TerrainDefinition)
- [x] DefaultIncomeMap with varied rates (Land Base: 100, Naval: 150, Airport: 200, Silo: 300, Mines: 500)
- [x] StartGamePage server-rendered configuration with GameConfiguration proto
- [x] Income configuration UI (currently using defaults from terrain definitions)
- [x] Starting coins configuration per player (default: 300)
- [x] Comprehensive unit tests for income system (5 tests covering all scenarios)

**Remaining Tasks**:
- [ ] Building capture mechanics
- [ ] Fog of war implementation
- [ ] Turn time limits and timers
- [ ] Team play mechanics (already in proto, needs implementation)

---

### AI Opponents
**Priority**: Medium
**Status**: Foundation exists (services/ai/), needs integration

**Tasks**:
- [ ] Integrate existing advisor system into AI player
- [ ] Implement basic AI decision making
- [ ] Add AI difficulty levels
- [ ] AI turn processing and move execution
- [ ] AI vs AI testing

---

## Phase 3: Production Readiness

### Database Migration
**Priority**: High
**Status**: Not Started

**Tasks**:
- [ ] Design PostgreSQL schema for games, states, history, proposals
- [ ] Implement DatabaseGamesService
- [ ] Migration utilities from file storage to database
- [ ] Connection pooling and performance optimization
- [ ] Backup and restore procedures

---

### Real-Time Updates
**Priority**: High
**Status**: Not Started

**Tasks**:
- [ ] WebSocket server implementation
- [ ] Client WebSocket connection management
- [ ] Real-time game state push notifications
- [ ] Proposal status updates via WebSocket
- [ ] Player presence and online status
- [ ] Connection recovery and reconnection logic

---

### Performance Optimization
**Priority**: Medium
**Status**: Not Started

**Tasks**:
- [ ] Load testing with concurrent games
- [ ] Profiling and optimization of hot paths
- [ ] Caching strategy review and optimization
- [ ] Sub-100ms move processing target
- [ ] 60fps rendering target for Phaser scenes
- [ ] Memory leak detection and fixing

---

## Testing & Quality

### Comprehensive Test Coverage
**Priority**: High
**Status**: Some coverage exists, needs expansion

**Tasks**:
- [ ] Unit tests for all move processor functions
- [ ] Integration tests for multiplayer coordination
- [ ] End-to-end browser tests with Playwright
- [ ] Error scenario testing (network failures, invalid moves, concurrent access)
- [ ] Performance regression tests
- [ ] Load testing for server capacity

---

### Code Quality
**Priority**: Medium
**Status**: Ongoing

**Tasks**:
- [ ] Remove unused helper methods (legacy GameState methods)
- [ ] Optimize event system to reduce redundant notifications
- [ ] Improve user-facing error messages
- [ ] Add comprehensive inline documentation
- [ ] Code review and refactoring pass

---

## Documentation

### User Documentation
**Priority**: Medium
**Status**: Minimal

**Tasks**:
- [ ] Complete user guide for CLI tool
- [ ] Browser UI tutorial and walkthrough
- [ ] Map editor guide
- [ ] Game rules and mechanics documentation
- [ ] Multiplayer setup guide
- [ ] Troubleshooting guide

---

### Developer Documentation
**Priority**: Medium
**Status**: Good foundation in ARCHITECTURE.md

**Tasks**:
- [ ] API documentation for services
- [ ] WASM integration guide
- [ ] Contributing guidelines
- [ ] Development environment setup guide
- [ ] Testing strategy and guidelines
- [ ] Deployment guide for production

---

## Mobile & Accessibility

### Mobile Support
**Priority**: Low
**Status**: 🔄 IN PROGRESS - Page Variant Architecture Ready for Mobile

**Completed**:
- [x] Responsive layouts for mobile screens (WorldViewerPage, StartGamePage)
- [x] Mobile-optimized UI components (bottom sheets, FABs, responsive headers)
- [x] Responsive world listing with grid/table views
- [x] Mobile navigation improvements (dropdown menus for actions)
- [x] Responsive game configuration panel with bottom sheet
- [x] Splash screen system across all pages
- [x] GameViewerPageBase abstract class architecture (2025-11-04)
- [x] GameViewerPageDockView implementation (flexible layout)
- [x] GameViewerPageGrid implementation (static grid layout)
- [x] Page variant pattern ready for mobile implementation
- [x] Responsive header menu system with drawer/dropdown switching (2025-01-05)
- [x] Animated mobile drawer below header with backdrop fade (2025-01-05)
- [x] Toast positioning adjusted for mobile bottom bars (2025-01-05)
- [x] Single global drawer with CSS positioning (no DOM manipulation) (2025-01-05)
- [x] Event listeners preserved on all header buttons (End Turn, Undo, Screenshot) (2025-01-05)
- [x] Consistent button content across all GameViewer page variants (2025-01-05)

**Remaining Tasks**:
- [x] GameViewerPageMobile with bottom drawer system (2025-11-04)
- [x] Context-aware button ordering (unit/tile/nothing selected) (2025-11-04)
- [x] MobileBottomDrawer reusable component (2025-11-04)
- [x] CompactSummaryCard for terrain+unit info (2025-11-04)
- [x] CompactSummaryCard template refactoring (inline HTML → templar.html) (2025-01-04)
- [x] Presenter-driven mobile UI updates via RPC (2025-01-04)
- [x] Panel interface architecture for CompactCard (2025-01-04)
- [x] TurnOptionsPanel template structure fix (2025-01-05)
- [ ] Touch controls for unit selection and movement
- [ ] Touch gesture support (pinch to zoom, pan)
- [ ] Mobile performance optimization
- [ ] Touch-friendly hex selection and highlighting

**Architecture Implementation (2025-11-04 / 2025-01-04)**:
- **TypeScript Components**:
  - GameViewerPageMobile.ts: Mobile page variant with context-aware bottom bar
  - MobileBottomDrawer.ts: Reusable drawer component (60-70% height, auto-close)
  - CompactSummaryCard.ts: Top banner showing terrain+unit selection info
  - GameViewerPageMobile.html: Mobile template with 5 drawers and bottom action bar
- **Button Ordering**: Dynamic reordering inferred from allowed panels (no event bus)
- **Presenter-Driven**: All UI updates via RPC calls (SetAllowedPanels, SetCompactSummaryCard)
- **Template Refactoring (2025-01-04)**:
  - CompactSummaryCard.templar.html: Server-side Go template with theme integration
  - Panel interface: CompactSummaryCardPanel with Base/Browser implementations
  - Clean separation: Go handles data, templates handle presentation
  - Consistent with TurnOptions, UnitStats, TerrainStats panel patterns

---

### Accessibility
**Priority**: Low
**Status**: Not Started

**Tasks**:
- [ ] Screen reader support
- [ ] Keyboard navigation for all UI elements
- [ ] High contrast mode
- [ ] Colorblind-friendly themes
- [ ] ARIA labels and semantic HTML

---

## Future Considerations

### Advanced Features (Phase 4+)
- [ ] Replay system with move history playback
- [ ] Spectator mode for ongoing games
- [ ] Tournament system and matchmaking
- [ ] Player rankings and statistics
- [ ] Custom game modes and modding support
- [ ] Cross-platform mobile apps (React Native / Flutter)
- [ ] Social features (friends list, chat, lobbies)

---

## Status Summary

**Phase 1 (Core Gameplay)**: ✅ COMPLETE
**Phase 2 (Multiplayer)**: 🔄 90% COMPLETE - Testing needed
**Phase 3 (Production)**: 🚧 NOT STARTED
**Polish & UX**: 🚧 ONGOING

**Next Focus**: CLI synchronous panels → Multiplayer testing → Production readiness
