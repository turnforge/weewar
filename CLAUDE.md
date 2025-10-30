
## Understand the Project First
- This is a project for building a template for webapps.  Go through the README and the various .md files to understand the project and the motivation and where we are.

## Coding Style and Conservativeness

- Be conservative on how many comments are you are adding or modifying unless it is absolutely necessary (for example a comment could be contradicting what is going on - in which case it is prudent to modify it).  
- When modifying files just focus on areas where the change is required instead of diving into a full fledged refactor.
- Make sure you ignore 'gen' and 'node_modules' as it has a lot of files you wont need for most things and are either auto generated or just package dependencies
- When updating .md files and in commit messages use emojis and flowerly languages sparingly.  We dont want to be too grandios or overpromising.
- Make sure the playwright tool is setup so you can inspect the browser when we are implementing and testing the Dashboard features.
- Do not refer to claude or anthropic or gemini in your commit messages
- Do not rebuild the server - it will be continuosly be rebuilt and run by the devloop.
- Find the root cause of an issue before figuring out a solution.  Fix problems.
- Do not create workarounds for issues without asking.  Always find the root cause of an issue and fix it.
- The web module automatically builds when files are changed - DO NOT run npm build or npm run build commands.
- Proto files are automatically regenerated when changed - DO NOT run buf generate commands.
- In general DONT be defensive by catching errors or null checking objects that when null would make the whole page fail anyway.    Dont just try/catch to log errors - let exceptions happen naturally so errors are NOT covered up and error locations are easier to identify.  We are still in experimenting/revising phase so we should harden as far as possible and identify failure modes rather than covering them up with try/catches (or even null checks when somethigns are really mandatory for the game to function).   Let us use preconditions more when possible.

## Continuous Builds

Builds for frontend, wasm, backend are all running continuously and can be queried using the `devloop` cli tool.   devloop is a command for watching and live reloading your projects.  It is like Air + Make on steroids.   You have the following devloop commands:
- `devloop config` - Get configuration from running devloop server
- `devloop paths` - List all file patterns being watched
- `devloop trigger <rulename>` - Trigger execution of a specific rule
- `devloop logs <rulename>`  - Stream logs from running devloop server
- `devloop status <rulename>` - Get status of rules from running devloop server

## Rules Data Extraction

The `cmd/extract-rules-data` tool scrapes game rules from saved WeeWar HTML pages:

**Data Sources:**
- `~/dev-app-data/weewar/data/Tiles/*.html` - Terrain pages with unit interaction tables
- `~/dev-app-data/weewar/data/Units/*.html` - Unit pages with stats and combat damage charts

**Extraction Architecture:**
- Uses `ExtractHtmlTable()` utility for consistent table parsing across all scrapers
- All functions use htmlquery/XPath instead of manual tree traversal
- Generates two files: `weewar-rules.json` (93KB core rules) and `weewar-damage.json` (1.2MB combat data)

**Key Extraction Functions:**
- `extractTerrainUnitInteractions`: Parses terrain-unit interaction tables using htmlquery with column indexing
- `extractUnitDefinition`: Extracts unit stats, classification, attack ranges, costs
- `extractAttackTable`: Parses attack matrices using ExtractHtmlTable for base damage values
- `extractUnitCombatProperties`: Extracts damage distributions from card tooltips
- `extractActionOrder`: Parses Progression badges for action sequences

**Running the Extractor:**
```bash
cd cmd/extract-rules-data
go run .
# Outputs: weewar-rules.json and weewar-damage.json
cp *.json ../../assets/
# Then run buf generate to update proto-generated code
```

**Proto Field Naming:** Proto uses snake_case in JSON but camelCase in Go (e.g., `buildable_unit_ids` → `BuildableUnitIds`)

## Summary instructions

- When you are using compact, please focus on test output and code changes

- For the ROADMAP.md always use the top-level ./ROADMAP.md so we have a global view of the roadmap instead of being fragemented in various folders.

## Session Workflow Memories
- When you checkpoint update all relevant .md files with our latest understanding, statuses and progress in the current session and then commit.

## Debugging Guide

### Game Storage Structure

Games are stored in `~/dev-app-data/weewar/storage/games/{gameId}/`:
- **metadata.json**: Game configuration (players, teams, settings, world_id)
- **state.json**: Current game state (tiles, units, current_player, turn_counter)
- **history.json**: Move history (groups of moves with results)

Worlds are stored in `~/dev-app-data/weewar/storage/worlds/{worldId}/`:
- **metadata.json**: World metadata (name, description, creator)
- **world.json**: Map data (tiles, starting units)

### Reading Game State with jq

**Check game status:**
```bash
jq '{current_player, turn_counter, status}' ~/dev-app-data/weewar/storage/games/{gameId}/state.json
```

**List all units:**
```bash
jq '.world_data.units[] | {q, r, player, unit_type, shortcut, health: .available_health, moves: .distance_left}' ~/dev-app-data/weewar/storage/games/{gameId}/state.json
```

**List units for specific player:**
```bash
jq '.world_data.units[] | select(.player == 1) | {shortcut, q, r, moves: .distance_left}' ~/dev-app-data/weewar/storage/games/{gameId}/state.json
```

**Check specific tile:**
```bash
jq '.world_data.tiles[] | select(.q == 0 and .r == -2)' ~/dev-app-data/weewar/storage/games/{gameId}/state.json
```

**View recent moves:**
```bash
jq '.groups[-3:] | .[] | {started_at, moves: .moves | length, results: .move_results | length}' ~/dev-app-data/weewar/storage/games/{gameId}/history.json
```

**Check player configuration:**
```bash
jq '.config.players[] | {player_id, player_type, color, team_id}' ~/dev-app-data/weewar/storage/games/{gameId}/metadata.json
```

### Proto Field Naming Convention

**Important**: Proto fields use snake_case in JSON but camelCase in Go structs:
- JSON: `available_health`, `distance_left`, `unit_type`, `current_player`, `turn_counter`
- Go: `AvailableHealth`, `DistanceLeft`, `UnitType`, `CurrentPlayer`, `TurnCounter`

When reading JSON files, always use snake_case. When writing Go code, use camelCase.

### CLI Debugging Commands

The `ww` CLI tool is installed in GOBIN and available globally (to rebuild this binary run `make cli` from the weewar
folder):

**Basic commands:**
```bash
export WEEWAR_GAME_ID=c5380903  # Or use --game-id flag

ww status                    # Show game state
ww units                     # List all units
ww options B1                # Show available moves for unit B1
ww move B1 0,-3             # Move unit by coordinates
ww move B1 R                # Move unit by direction (L/R/TL/TR/BL/BR)
ww attack A1 B2             # Attack unit
ww endturn                  # End current player's turn

# Flags
ww --verbose units          # Show debug output
ww --dryrun move B1 R      # Preview move without saving
ww --json status            # Output as JSON
```

**Direction shortcuts:** L (left), R (right), TL (top-left), TR (top-right), BL (bottom-left), BR (bottom-right)

### Key Service Files and Architecture

**Core Game Logic** (`services/`):
- **game.go**: Game struct, NewGame(), topUpUnitIfNeeded(), validation
- **world.go**: World struct, hex coordinate management, unit/tile operations
- **moves.go**: MoveProcessor, ProcessMoves(), ProcessMoveUnit(), ProcessAttackUnit(), copyUnit()
- **base_games_service.go**: BaseGamesServiceImpl, ProcessMoves RPC endpoint
- **fsgames_service.go**: FSGamesService file storage with caching
- **singleton_games_service.go**: SingletonGamesService for in-memory operations
- **singleton_gameview_presenter.go**: Presenter orchestrating UI updates
- **panels.go**: BaseGameState, BaseTurnOptionsPanel, BaseUnitPanel, etc.
- **rules_engine.go**: RulesEngine for movement costs, combat, unit data

**CLI Tool** (`cmd/cli/`):
- **cmd/presenter.go**: createPresenter(), savePresenterState()
- **cmd/status.go**: Status command implementation
- **cmd/units.go**: Units listing command
- **cmd/options.go**: Options display command
- **cmd/move.go**: Move execution command
- **cmd/endturn.go**: End turn command

**Utilities** (`services/`):
- **position_parser.go**: ParsePosition(), ParseUnitShortcut(), ParseDirection()
- **path_display.go**: FormatPath(), DisplayPath() for CLI output
- **options_formatter.go**: FormatOptions() for CLI display
- **utils.go**: NewWorld(), various conversion helpers

### Common Debugging Patterns

**1. Check if unit has movement points:**
```bash
ww --verbose options B1 | grep "DistanceLeft"
```

**2. Inspect move history to see what happened:**
```bash
jq '.groups[-1] | {moves: .moves, results: .move_results | length}' ~/dev-app-data/weewar/storage/games/{gameId}/history.json
```

**3. Verify unit shortcuts are preserved:**
```bash
jq '.world_data.units[] | {shortcut, q, r}' ~/dev-app-data/weewar/storage/games/{gameId}/state.json
```

**4. Check lazy top-up fields:**
```bash
jq '.world_data.units[] | {shortcut, last_topped_up_turn, distance_left, turn: .last_acted_turn}' ~/dev-app-data/weewar/storage/games/{gameId}/state.json
```

**5. Debug presenter state loading:**
```go
// In services/singleton_gameview_presenter.go
fmt.Printf("Game loaded: turn=%d, player=%d\n", state.TurnCounter, state.CurrentPlayer)
```

**6. Trace move processing:**
```go
// Add to services/moves.go ProcessMoveUnit
fmt.Printf("ProcessMoveUnit: unit at (%d,%d) moving to (%d,%d), cost=%d, distance_left=%d\n",
    unit.Q, unit.R, to.Q, to.R, cost, unit.DistanceLeft)
```

### FSGamesService Cache Gotcha

**Problem**: FSGamesService has in-memory caches (gameCache, stateCache, historyCache). When CLI modifies files on disk, the gRPC server's cache becomes stale.

**Current Solution**: Cache is disabled in GetGame() to always read fresh from disk.

**Files**: services/fsgames_service.go:~150-160

**Note**: If cache is re-enabled in the future, implement file watching or cache invalidation.

### Presenter Architecture

**Presenter Flow**: CLI → FSGamesService.GetGame() → SingletonGamesService (in-memory) → SingletonGameViewPresenterImpl

**Key Pattern**: CLI loads game from disk into FSGamesService, copies to SingletonGamesService, creates presenter, makes changes, then saves back via FSGamesService.UpdateGame()

**Files**:
- `cmd/cli/cmd/presenter.go`: createPresenter(), savePresenterState()
- `services/singleton_gameview_presenter.go`: ProcessSceneClicked(), ProcessMoves()
- `services/panels.go`: Base and Browser panel implementations

### Unit Copy Helper Function

**Important**: When creating unit copies for history recording, always use `copyUnit()` helper function in services/moves.go. This ensures all fields (including Shortcut) are preserved.

**Pattern**:
```go
// Correct - uses helper
previousUnit := copyUnit(unit)

// Incorrect - manual copy may miss fields
previousUnit := &v1.Unit{Q: unit.Q, R: unit.R, ...} // May forget Shortcut!
```

**Why**: When new proto fields are added to Unit, only one location needs updating.

### Lazy Top-Up Pattern

**Concept**: Units don't automatically reset movement points at turn start. Instead, they're "topped up" on-demand when accessed.

**Implementation**: `topUpUnitIfNeeded()` in services/game.go checks `unit.LastToppedupTurn < game.TurnCounter`

**Called by**:
- ProcessMoveUnit (before validating movement)
- ProcessAttackUnit (before validating attack)
- GetOptionsAt (before calculating options)

**Fields**:
- `LastToppedupTurn`: Last turn when unit was refreshed
- `LastActedTurn`: Last turn when unit performed an action
- `DistanceLeft`: Current remaining movement points

