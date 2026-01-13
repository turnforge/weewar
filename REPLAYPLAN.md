# Game Move Testing & Verification System

## Overview

A system for recording, replaying, and verifying game moves with expected outcomes. Reuses existing proto messages (GameMove, WorldChange) for expectations. Test cases include a starting state snapshot and expected changes for each move.

## Key Design Decisions

1. **Reuse existing protos**: Expectations use `GameMove` with `changes` field (list of `WorldChange`) - no custom expectation structs
2. **Full snapshot for starting state**: Test includes a `GameState` snapshot to reset to before replay
3. **List-based expectations**: Moves stored as a list, sorted by `(group_number, move_number)`
4. **ListMoves API**: Add to GamesService for fetching moves by group range
5. **Seed-based determinism**: Record and replay with same RNG seed
6. **CLI-driven stepping**: Enter/Space to advance through moves in browser

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     TestCase (Proto Message)                     │
│  - metadata (name, game_id, seed)                                │
│  - starting_state: GameState snapshot                            │
│  - expected_moves: []GameMove with populated changes             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      ww test CLI                                 │
│  - verify: replay from snapshot, compare actual vs expected      │
│  - record: capture starting state + expected changes             │
│  - step: replay with browser visualization                       │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      GamesService                                │
│  - ListMoves(game_id, from_group, to_group) → []GameMove        │
│  - ProcessMoves for replay                                       │
│  - RNG seed controls combat determinism                          │
└─────────────────────────────────────────────────────────────────┘
```

## Proto Definitions

### New Messages in `protos/lilbattle/v1/services/games.proto`

```protobuf
// Test case that can be stored/loaded
message TestCase {
  string id = 1;
  string name = 2;
  string description = 3;
  string game_id = 4;
  int64 seed = 5;  // RNG seed for deterministic replay

  // Snapshot of game state to start from
  GameState starting_state = 6;

  // Expected moves with their expected WorldChanges
  // Sorted by (group_number, move_number)
  repeated GameMove expected_moves = 7;
}

// ListMoves request
message ListMovesRequest {
  string game_id = 1;
  int64 from_group = 2;  // Inclusive, 0 = start
  int64 to_group = 3;    // Inclusive, 0 = end
}

message ListMovesResponse {
  repeated GameMove moves = 1;
}

// Verify test case
message VerifyTestRequest {
  TestCase test_case = 1;
  bool step_mode = 2;  // If true, pause after each move
}

message VerifyTestResponse {
  bool passed = 1;
  repeated MoveVerifyResult results = 2;
}

message MoveVerifyResult {
  int64 group_number = 1;
  int64 move_number = 2;
  bool passed = 3;
  repeated ChangeDifference differences = 4;
}

message ChangeDifference {
  string path = 1;  // e.g., "changes[0].unit_damaged.updated_unit.health"
  string expected = 2;
  string actual = 3;
}
```

### Add to GamesService

```protobuf
service GamesService {
  // ... existing methods ...

  // List moves for a game, optionally filtered by group range
  rpc ListMoves(ListMovesRequest) returns (ListMovesResponse);

  // Verify a test case against actual game behavior
  rpc VerifyTest(VerifyTestRequest) returns (VerifyTestResponse);

  // Record a test case from current game history
  rpc RecordTest(RecordTestRequest) returns (RecordTestResponse);
}
```

## Test File Format

**Location**: `~/dev-app-data/lilbattle/tests/{game_id}/{test_name}.json`

Using JSON (proto-compatible) instead of YAML for direct proto serialization:

```json
{
  "id": "combat-test-1",
  "name": "Combat validation - infantry vs tank",
  "description": "Verify damage calculation and counter-attack",
  "game_id": "c5380903",
  "seed": 12345,
  "starting_state": {
    "current_player": 1,
    "turn_counter": 1,
    "world_data": {
      "units_map": {
        "0,-2": {"shortcut": "A1", "unit_type": 1, "player": 1, "health": 10, "q": 0, "r": -2},
        "1,-2": {"shortcut": "B2", "unit_type": 5, "player": 2, "health": 10, "q": 1, "r": -2}
      },
      "tiles_map": { ... }
    }
  },
  "expected_moves": [
    {
      "group_number": 1,
      "move_number": 1,
      "player": 1,
      "description": "Move infantry A1 toward tank",
      "move_unit": {
        "from_q": 0, "from_r": -2,
        "to_q": 0, "to_r": -3
      },
      "changes": [
        {
          "unit_moved": {
            "previous_unit": {"q": 0, "r": -2, "shortcut": "A1", "distance_left": 3},
            "updated_unit": {"q": 0, "r": -3, "shortcut": "A1", "distance_left": 0}
          }
        }
      ]
    },
    {
      "group_number": 1,
      "move_number": 2,
      "player": 1,
      "description": "Infantry attacks tank",
      "attack_unit": {
        "attacker_q": 0, "attacker_r": -3,
        "defender_q": 1, "defender_r": -2
      },
      "changes": [
        {
          "unit_damaged": {
            "previous_unit": {"q": 1, "r": -2, "shortcut": "B2", "health": 10},
            "updated_unit": {"q": 1, "r": -2, "shortcut": "B2", "health": 7}
          }
        },
        {
          "unit_damaged": {
            "previous_unit": {"q": 0, "r": -3, "shortcut": "A1", "health": 10},
            "updated_unit": {"q": 0, "r": -3, "shortcut": "A1", "health": 5}
          }
        }
      ]
    },
    {
      "group_number": 1,
      "move_number": 3,
      "player": 1,
      "description": "End turn",
      "end_turn": {},
      "changes": [
        {
          "player_changed": {
            "previous_player": 1,
            "new_player": 2,
            "previous_turn": 1,
            "new_turn": 2
          }
        }
      ]
    }
  ]
}
```

## CLI Commands

```bash
# Verify test case against game behavior
ww test verify tests/combat-test.json
ww test verify tests/combat-test.json --verbose

# Step through with browser visualization
ww test step tests/combat-test.json
# Press Enter to advance, 'q' to quit

# Record test case from current history
ww test record --output tests/new-test.json
ww test record --from-group 3 --to-group 5 --output tests/partial.json

# List moves for current game
ww test list-moves
ww test list-moves --from-group 1 --to-group 3

# List available test files
ww test list

# Run all tests for current game
ww test run-all
```

## Implementation Phases

### Phase 1: Proto & API Changes

1. **Add `description` to GameMove** (already done)

2. **Add ListMoves RPC to GamesService**
   - `protos/lilbattle/v1/services/games.proto`
   - Implement in `services/fsgames_service.go`
   - Implement in `services/singleton_games_service.go`

3. **Add TestCase message**
   - `protos/lilbattle/v1/services/games.proto` (or new `testing.proto`)

### Phase 2: Verification Logic

**New file**: `cmd/cli/testing/verifier.go`

```go
package testing

type Verifier struct {
    GamesService services.GamesService
}

// Compare expected WorldChanges against actual
func (v *Verifier) CompareChanges(expected, actual []*v1.WorldChange) []*ChangeDifference {
    // For each expected change, find matching actual change
    // Compare relevant fields (ignore timestamps, etc.)
    // Return differences
}

// Verify a single move
func (v *Verifier) VerifyMove(expected, actual *v1.GameMove) *MoveVerifyResult {
    result := &MoveVerifyResult{
        GroupNumber: expected.GroupNumber,
        MoveNumber:  expected.MoveNumber,
        Passed:      true,
    }

    diffs := v.CompareChanges(expected.Changes, actual.Changes)
    if len(diffs) > 0 {
        result.Passed = false
        result.Differences = diffs
    }

    return result
}

// Verify entire test case
func (v *Verifier) VerifyTestCase(tc *v1.TestCase) (*v1.VerifyTestResponse, error) {
    // 1. Reset game to starting_state with seed
    // 2. For each expected move:
    //    a. Re-execute the move via ProcessMoves
    //    b. Compare actual changes to expected changes
    //    c. Record differences
    // 3. Return results
}
```

### Phase 3: Recording

**New file**: `cmd/cli/testing/recorder.go`

```go
package testing

type Recorder struct {
    GamesService services.GamesService
}

// Record test case from current game history
func (r *Recorder) RecordTestCase(gameID string, fromGroup, toGroup int64) (*v1.TestCase, error) {
    // 1. Get current game state (for starting_state snapshot)
    // 2. Use ListMoves to get moves in range
    // 3. Build TestCase with moves as expected_moves
    // 4. Return TestCase
}
```

### Phase 4: CLI Integration

**New file**: `cmd/cli/cmd/test.go`

```go
var testCmd = &cobra.Command{
    Use:   "test",
    Short: "Test and verify game moves",
}

var testVerifyCmd = &cobra.Command{
    Use:   "verify <test-file>",
    Short: "Verify test case against game behavior",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 1. Load TestCase from JSON file
        // 2. Create verifier
        // 3. Run verification
        // 4. Print results
    },
}

var testStepCmd = &cobra.Command{
    Use:   "step <test-file>",
    Short: "Step through moves with browser",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 1. Load TestCase
        // 2. Reset game to starting_state
        // 3. For each move:
        //    a. Show move info in CLI
        //    b. Execute via presenter (animates in browser)
        //    c. Compare and show result
        //    d. Wait for Enter
    },
}

var testRecordCmd = &cobra.Command{
    Use:   "record",
    Short: "Record test case from history",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 1. Create recorder
        // 2. Record test case
        // 3. Write to JSON file
    },
}

var testListMovesCmd = &cobra.Command{
    Use:   "list-moves",
    Short: "List moves from history",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 1. Call ListMoves RPC
        // 2. Print moves
    },
}
```

### Phase 5: Browser Stepping

**New file**: `cmd/cli/testing/stepper.go`

```go
package testing

type Stepper struct {
    Presenter *services.GameViewPresenter
    Verifier  *Verifier
}

func (s *Stepper) StepThrough(tc *v1.TestCase) error {
    // Reset to starting state
    s.resetToState(tc.StartingState, tc.Seed)

    for _, expectedMove := range tc.ExpectedMoves {
        fmt.Printf("\nMove %d.%d: %s\n",
            expectedMove.GroupNumber,
            expectedMove.MoveNumber,
            expectedMove.Description)

        // Execute move via presenter (triggers browser animation)
        actualMove := s.executeMove(expectedMove)

        // Verify
        result := s.Verifier.VerifyMove(expectedMove, actualMove)
        s.printResult(result)

        // Wait for user
        fmt.Print("Press Enter to continue (q to quit)...")
        if s.readInput() == "q" {
            return nil
        }
    }

    return nil
}
```

## File Structure

```
protos/lilbattle/v1/services/
└── games.proto             # Add ListMoves, TestCase, VerifyTest

services/
├── fsgames_service.go      # Implement ListMoves
└── singleton_games_service.go

cmd/cli/
├── testing/
│   ├── verifier.go         # Change comparison logic
│   ├── recorder.go         # Test case recording
│   └── stepper.go          # Browser stepping
└── cmd/
    └── test.go             # CLI commands

~/dev-app-data/lilbattle/tests/
└── {game_id}/
    ├── combat-basic.json
    └── movement-costs.json
```

## Critical Files to Modify

1. **`protos/lilbattle/v1/services/games.proto`**
   - Add `ListMovesRequest`, `ListMovesResponse`
   - Add `TestCase` message
   - Add `ListMoves` RPC

2. **`services/fsgames_service.go`**
   - Implement `ListMoves`

3. **`services/singleton_games_service.go`**
   - Implement `ListMoves`

4. **`lib/game.go`**
   - Add `ResetRNG()` method
   - Add `LoadState(GameState)` method for snapshot restore

5. **`cmd/cli/cmd/root.go`**
   - Register test subcommand

## Verification Logic Details

When comparing `WorldChange` entries:

1. **Match by type**: Find corresponding change of same type
2. **Compare key fields**:
   - `UnitMovedChange`: Compare `updated_unit.q`, `updated_unit.r`, `updated_unit.distance_left`
   - `UnitDamagedChange`: Compare `updated_unit.health`
   - `UnitKilledChange`: Compare `previous_unit.shortcut`
   - `PlayerChangedChange`: Compare `new_player`, `new_turn`
   - `CoinsChangedChange`: Compare `new_coins`

3. **Ignore timestamps and sequence numbers** - these are non-deterministic

## Workflow Example

```bash
# 1. Play game normally
ww move A1 R
ww attack A1 B2
ww endturn

# 2. Record test case from history
ww test record --output tests/combat-test.json

# 3. View the recorded test
cat ~/dev-app-data/lilbattle/tests/c5380903/combat-test.json | jq .

# 4. Verify (should pass)
ww test verify tests/combat-test.json
# Output: All 3 moves passed

# 5. Make code changes to combat formula...

# 6. Re-verify
ww test verify tests/combat-test.json
# Output: Move 1.2 FAILED
#   changes[0].unit_damaged.updated_unit.health: expected 7, got 6

# 7. Debug with browser stepping
ww test step tests/combat-test.json
# Step through each move, see animations, identify issue
```

## Benefits

1. **Proto reuse**: No custom structs - uses existing GameMove, WorldChange
2. **Snapshot-based**: Start from known state, not beginning of game
3. **ListMoves API**: Useful for other features (history viewer, replays)
4. **GamesService agnostic**: Works with FS, gRPC, or in-memory
5. **Deterministic**: Seed-based RNG for exact reproduction
6. **Incremental stepping**: See each move in browser during debug
