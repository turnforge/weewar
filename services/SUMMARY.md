# Services Package Summary

This package contains the core service implementations for the LilBattle backend, including game logic, world management, and storage abstractions.

## Architecture

### Service Hierarchy

The service layer uses an embedding pattern for code reuse:

**Worlds Services:**
```
BaseWorldsService (base interface methods)
    ↓ embeds
BackendWorldsService (screenshot indexing, optimistic locking)
    ↓ embeds
GORMWorldsService / FSWorldsService (storage-specific implementations)
```

**Games Services:**
```
BaseGamesService (base interface methods)
    ↓ embeds
BackendGamesService (screenshot indexing, optimistic locking)
    ↓ embeds
GORMGamesService / FSGamesService (storage-specific implementations)
```

### Key Components

**Base Services**
- `BaseWorldsService`: Core world service interface
- `BackendWorldsService`: Mixin providing screenshot indexing and completion handling for worlds
  - Uses `WorldDataUpdater` interface for storage-agnostic operations
  - Handles screenshot completion callbacks with optimistic locking
  - Shared between GORM and filesystem implementations
- `BaseGamesService`: Core game service interface
- `BackendGamesService`: Mixin providing screenshot indexing and completion handling for games
  - Uses `GameStateUpdater` interface for storage-agnostic operations
  - Updates GameState.WorldData.ScreenshotIndexInfo (GameWorldDataGORM embedded without table)
  - GameState table includes flattened WorldData fields with screenshot_index_ prefix
  - Shared between GORM and filesystem implementations

**Storage Implementations**
- `gormbe/worlds_service.go`: Database-backed world storage using GORM (PostgreSQL)
- `fsbe/worlds_service.go`: Filesystem-backed world storage using JSON files
- `gaebe/worlds_service.go`: Google Cloud Datastore-backed world storage for App Engine
- All implement `WorldDataUpdater` interface:
  - `GetWorldData(ctx, id)`: Get current version for optimistic locking
  - `UpdateWorldDataIndexInfo(ctx, id, oldVersion, lastIndexedAt, needsIndexing)`: Update index info with version check
- `gormbe/games_service.go`: Database-backed game storage using GORM (PostgreSQL)
- `fsbe/games_service.go`: Filesystem-backed game storage using JSON files
- `gaebe/games_service.go`: Google Cloud Datastore-backed game storage for App Engine
- All implement `GameStateUpdater` interface:
  - `GetGameStateVersion(ctx, id)`: Get current version for optimistic locking
  - `UpdateGameStateScreenshotIndexInfo(ctx, id, oldVersion, lastIndexedAt, needsIndexing)`: Update GameState's WorldData screenshot index info with version check

**FileStore Services**
- `fsbe/filestore.go`: Local filesystem storage with path security
  - Prevents directory traversal attacks (rejects `..`, absolute paths)
  - Resolves all paths relative to BasePath
- `r2/filestore.go`: Cloudflare R2 object storage backend
  - Generates presigned URLs with multiple expiries (15m, 1h, 24h)
  - Uses S3-compatible API

**Screenshot Pipeline**
- `screenshots.go`: Batch processing screenshot indexer
  - Groups updates within 30-second windows (using gocurrent.Reducer)
  - Renders multiple themes per world/game (default, modern, fantasy)
  - Uploads to filestore at: `screenshots/{kind}/{id}/{theme}.{ext}`
  - Tracks errors per theme in `ThemeErrors` map
  - Calls completion callback after batch finishes

**Game Logic**
- `game.go`: Core game state management
- `world.go`: Hex grid coordinate system and tile/unit operations
- `moves.go`: Move processing and validation
- `rules_engine.go`: Game rules (movement costs, combat, unit data)
- `panels.go`: UI state management for game views
- `singleton_gameview_presenter.go`: Orchestrates UI updates

**Position Parsing**
- `position_parser.go`: Unified parser supporting multiple formats
  - Unit shortcuts: `A1`, `B2`
  - Q,R coordinates: `0,-3`, `5,2`
  - Row,Col coordinates: `r4,5`
  - Direction shortcuts: `L`, `R`, `TL`, `TR`, `BL`, `BR`
  - Tile prefix: `t:A1` (forces tile lookup instead of unit)

## Key Patterns

### Optimistic Locking
WorldData uses version-based optimistic locking to prevent concurrent update conflicts:
- Each update increments `WorldData.Version`
- Save operations include WHERE clause checking old version
- Screenshot indexer checks version before updating IndexInfo

### Screenshot Indexing Flow
1. `UpdateWorld` increments version, sets `NeedsIndexing=true`
2. Sends item to `ScreenShotIndexer` with new version
3. Indexer batches items (30s window), renders all themes
4. Completion callback checks version matches, updates IndexInfo
5. If version mismatch, skip update (world was modified again)

### Path Security (FileStore)
Multi-layered validation prevents directory traversal:
1. Reject absolute paths
2. Clean path (removes redundant separators)
3. Check for `..` prefix
4. Resolve to absolute path and verify within BasePath

### Lazy Top-Up Pattern
Units don't automatically reset movement at turn start. Instead, they're "topped up" on-demand:
- `LastToppedupTurn`: Last turn unit was refreshed
- `LastActedTurn`: Last turn unit performed action
- `topUpUnitIfNeeded()` called before movement/attack/options

### Presenter Architecture
CLI → FSGamesService.GetGame() → SingletonGamesService (in-memory) → SingletonGameViewPresenterImpl

### Authentication Flow

**gRPC Authentication** (`services/server/grpcserver.go`):
- Uses oneauth gRPC interceptors for authentication enforcement
- `PublicMethods []string` field configures methods that don't require authentication
- Environment controls:
  - `DISABLE_API_AUTH=true`: Skip all authentication (development mode)
  - `ENABLE_SWITCH_AUTH=true`: Allow X-Switch-User header for testing

**Public Methods** (no auth required):
- `WorldsService`: ListWorlds, GetWorld, GetWorlds
- `GamesService`: ListGames, GetGame, GetGames, SimulateAttack
- `GameSyncService`: Subscribe (for spectating)

**Private Methods** (auth required):
- All Create, Update, Delete operations
- ProcessMoves, GetOptionsAt, Broadcast

## File Organization

- `services/` - Core service implementations
  - `authz/` - Authorization utilities (user identity extraction, permission checks)
  - `gormbe/` - Database-backed services (GORM/PostgreSQL)
  - `fsbe/` - Filesystem-backed services (local JSON files)
  - `gaebe/` - Google Cloud Datastore-backed services (App Engine)
  - `singleton/` - In-memory single-game services (used by WASM and CLI)
  - `r2/` - Cloudflare R2 storage client
- Proto-generated code in `gen/go/lilbattle/v1/`
- Datastore entity code in `gen/datastore/` (generated by protoc-gen-dal)

## WASM Build Considerations

The `services` package is used by both server-side code and WASM (browser) builds. Some packages have build constraints:

**Server-only packages** (`//go:build !wasm`):
- `gormbe/` - Database operations not available in browser
- `fsbe/` - Filesystem operations not available in browser
- `gaebe/` - Datastore operations not available in browser
- `server/` - gRPC server infrastructure

**WASM-compatible packages**:
- `singleton/` - In-memory game state, used by browser WASM
- `authz/` - Has both server (`authz.go`) and WASM stub (`authz_wasm.go`)

**authz package architecture**:
- `authz.go` (`!wasm`): Extracts user ID from gRPC context via oneauth middleware
- `authz_wasm.go` (`wasm`): No-op stub since browser-side auth isn't needed (server enforces all authorization)

The server remains the source of truth for authorization. WASM clients already know the current user for UI purposes, but all permission checks happen server-side when processing requests.
