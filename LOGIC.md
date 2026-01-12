# Game Logic Refactoring Plan

This document tracks the migration of game logic from the services/presenter layer into `lib.Game`.

## Issue Reference
- GitHub Issue: #40 - Add controller methods to lib.Game with unified position parsing

## Goals

1. **Unified Position Handling**: Use `Position` proto with `Label`, `Q`, `R` fields throughout
2. **Controller Methods on lib.Game**: Add high-level methods like `Move()`, `Attack()`, `Build()`, `Capture()`, `EndTurn()`
3. **Clean Presenter**: Presenter only handles UI events and calls lib.Game controller methods
4. **Better Testing**: Enable pure Go tests instead of shell script e2e tests

## Current Status

### Phase 1: Proto Changes (COMPLETED)
- Added `Position` message type with `Label`, `Q`, `R` fields
- Updated action messages to use `Position`:
  - `MoveUnitAction`: `From`, `To` as Position
  - `AttackUnitAction`: `Attacker`, `Defender` as Position
  - `BuildUnitAction`: `Pos` as Position
  - `CaptureBuildingAction`: `Pos` as Position
  - `GetOptionsAtRequest`: `Pos` as Position

### Phase 2: Position Parsing Methods (COMPLETED)
Added to `lib/game.go`:
```go
// FromPos converts a Position proto to an AxialCoord
// Sets Q/R on the Position when parsing from label
func (g *Game) FromPos(pos *v1.Position) (AxialCoord, error)

// FromPosWithBase converts Position with optional base for relative directions
func (g *Game) FromPosWithBase(pos *v1.Position, base *AxialCoord) (AxialCoord, error)

// Pos parses a position string and returns a ParseTarget
func (g *Game) Pos(input string, from ...string) (*ParseTarget, error)
```

### Phase 3: Fix Broken Code (COMPLETED)

Files fixed:
- [x] `lib/moves.go` - Updated ProcessMoveUnit, ProcessAttackUnit, ProcessBuildUnit, ProcessCaptureBuilding
- [x] `lib/sort_utils.go` - Updated MoveUnitActionLess, AttackUnitActionLess
- [x] `services/games_service.go` - Updated GetOptionsAt, GetUnitOptions, GetTileOptions
- [x] `services/gameview_presenter.go` - Updated buildHighlightSpecs, executeMovementAction, BuildOptionClicked, SceneClicked
- [x] `services/options_formatter.go` - Updated FormatMoveUnitAction, FormatAttackUnitAction
- [x] `cmd/cli/cmd/options.go` - Updated to use Position in SceneClickedRequest and option field references
- [x] `cmd/cli/cmd/assert.go` - Updated to use Position and new field names
- [x] `cmd/cli/cmd/attack.go` - Updated to use Position in SceneClickedRequest
- [x] `cmd/cli/cmd/build.go` - Updated to use Position in BuildOptionClickedRequest
- [x] `cmd/cli/cmd/capture.go` - Updated to use Position in SceneClickedRequest
- [x] `cmd/cli/cmd/move.go` - Updated to use Position in SceneClickedRequest
- [x] `cmd/cli/cmd/output.go` - Updated option field references (To.Q, Defender.Q, Pos.Q)

### Phase 4: Add Controller Methods to lib.Game (COMPLETED)

Added these controller methods to `lib/game.go`:

- `Move(unit, target string) ([]*v1.WorldChange, error)` - Move unit to target position
- `Attack(attacker, defender string) ([]*v1.WorldChange, error)` - Attack target from attacker
- `Build(tile string, unitType int32) ([]*v1.WorldChange, error)` - Build unit at tile
- `Capture(unit string) ([]*v1.WorldChange, error)` - Start capturing building
- `EndTurn() ([]*v1.WorldChange, error)` - Advance to next player

All methods use `ParseTarget.Position()` to convert parsed targets to Position protos.
The target parameter in Move/Attack supports relative directions like "R", "TL,TR".

### Phase 5: Move GetOptionsAt Logic to lib.Game (COMPLETED)

Added option generation methods to `lib/game.go`:
- `GetOptionsAt(position string) (*v1.GetOptionsAtResponse, error)` - Main entry point
- `GetUnitOptions(unit *v1.Unit) ([]*v1.GameOption, *v1.AllPaths, error)` - Move/attack/capture options
- `GetTileOptions(tile *v1.Tile) ([]*v1.GameOption, error)` - Build options
- `FilterBuildOptionsByAllowedUnits(buildableUnits, allowedUnits []int32) []int32` - Helper

Updated `services/games_service.go` to delegate to `rtGame.GetOptionsAt()`.

### Phase 6: Simplify Presenter (DEFERRED)

The presenter needs to go through `ProcessMoves` for persistence (saves to storage, records history).
Full refactoring would require changing the ProcessMoves architecture.

Current state:
- Presenter uses `GamesService.GetOptionsAt()` which delegates to `lib.Game.GetOptionsAt()`
- `executeMovementAction` finds matching options and creates GameMove objects
- This pattern works and maintains persistence guarantees

The controller methods (`g.Move()`, `g.Attack()`, etc.) are primarily useful for:
- Testing (Phase 7) - clean API without needing full service layer
- CLI usage - direct game manipulation
- Future WASM/local-only scenarios where persistence isn't needed

### Phase 7: Write Go Tests (COMPLETED)

Created comprehensive controller tests in `tests/controller_test.go`.

### Phase 8: Remove MoveProcessor Struct (COMPLETED)

The `MoveProcessor` struct was removed since it was only used as a method namespace
with no state. All `Process*` methods are now directly on the `Game` struct.

**Before:**
```go
type MoveProcessor struct {}

func (m *MoveProcessor) ProcessMoves(game *Game, moves []*v1.GameMove) error
func (m *MoveProcessor) ProcessMoveUnit(g *Game, move *v1.GameMove, action *v1.MoveUnitAction, preventPassThrough bool) error
// etc.
```

**After:**
```go
func (g *Game) ProcessMoves(moves []*v1.GameMove) error
func (g *Game) ProcessMoveUnit(move *v1.GameMove, action *v1.MoveUnitAction, preventPassThrough bool) error
// etc.
```

This simplifies the API - callers now use `game.ProcessMoves(moves)` instead of
`(&MoveProcessor{}).ProcessMoves(game, moves)`.

**Files updated:**
- `lib/moves.go` - Removed MoveProcessor struct, changed method receivers
- `lib/game.go` - Updated controller methods to call `g.ProcessX()` directly
- `services/games_service.go` - Updated to call `rtGame.ProcessMoves()`
- `tests/*.go` - All test files updated to use Game methods

### Test Infrastructure (COMPLETED)

Test files in `tests/` created:
- `TestMoveController` - Move with relative direction ("R")
- `TestMoveControllerWithCoordinates` - Move using Q,R coordinates ("1,0")
- `TestMoveControllerMultipleDirections` - Move with chained directions ("R,R")
- `TestCaptureController` - Start capture, verify CaptureStartedTurn
- `TestCaptureCompletesAfterTurn` - Full capture lifecycle across turns
- `TestEndTurnController` - Player cycling and turn counter
- `TestAttackController` - Combat between units
- `TestBuildController` - Unit construction at bases
- `TestGetOptionsAtController` - Unit options (move/attack)
- `TestGetOptionsAtTileController` - Tile options (build)

Test infrastructure includes:
- `TestGameSetup` helper struct
- `AddTile`, `AddPlayerTile`, `AddUnit`, `AddUnitWithShortcut` helpers
- `AddGrassTiles` for creating movement terrain

Key fixes during testing:
- Fixed `FromPosWithBase` to use pre-resolved Q/R coordinates
- Fixed tile shortcut indexing (Player/Shortcut must be set before AddTile)
- Fixed `TopUpUnitIfNeeded` interaction with test setup
- Fixed `ProcessEndTurn` to use `g.NumPlayers()` instead of `g.World.PlayerCount()`

## Files Modified

### Proto Files
- `protos/lilbattle/v1/models/models.proto` - Added Position message, updated action messages
- `protos/lilbattle/v1/models/games_service.proto` - Updated GetOptionsAtRequest
- `protos/lilbattle/v1/services/games.proto` - Updated HTTP bindings

### Go Files
- `lib/game.go` - Added FromPos, FromPosWithBase, Pos, NumPlayers, controller methods (Move, Attack, Build, Capture, EndTurn, GetOptionsAt)
- `lib/moves.go` - Removed MoveProcessor struct, methods now on Game; uses Position fields
- `lib/sort_utils.go` - Updated to use Position fields
- `lib/rules_loader.go` - Added TileTypeGrass and UnitTypeSoldier constants
- `services/games_service.go` - Updated to use Position, calls `rtGame.ProcessMoves()` directly
- `services/gameview_presenter.go` - Updated to use Position
- `services/options_formatter.go` - Updated to use Position
- `tests/controller_test.go` - New comprehensive controller tests
- `tests/imports.go` - Removed MoveProcessor export, added test constants
- `tests/*.go` - All test files updated to use Game methods directly

### Remaining Files to Fix
All files have been updated for the Position migration. The `cmd/repl` package is deprecated and not maintained.
