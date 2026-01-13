# Multiplayer Sync Architecture

This document describes the real-time synchronization system for multiplayer games.

**Issue**: https://github.com/turnforge/lilbattle/issues/41
**Branch**: `feature/multiplayer-sync-41`

## Overview

The sync system enables multiplayer gameplay using a **local-first** architecture:

1. **Originator** (active player): Processes moves locally for immediate feedback, calls GamesService.ProcessMoves
2. **Server**: GamesService validates, persists, then calls SyncService.Broadcast
3. **Viewer** (other players): Receives WorldChanges via Subscribe stream, applies via presenter

```
Frontend A (originator)                Server                         Frontend B (viewer)
    │                                      │                               │
    │ WASM Presenter                       │                               │ WASM Presenter
    │      ↓                               │                               │
    │ ProcessMoves (local)                 │                               │
    │      ↓                               │                               │
    │ Immediate UI update                  │                               │
    │      ↓                               │                               │
    │──── GamesService.ProcessMoves() ────→│ Validate + Persist            │
    │                                      │      ↓                        │
    │←─── Response ────────────────────────│ SyncService.Broadcast()       │
    │                                      │      ↓                        │
    │                                      │════ gRPC Stream ═════════════→│
    │                                      │   (GameUpdate)                │
    │                                      │                               │ ApplyRemoteChanges
```

## Key Design Decisions

### 1. Service Separation

| Service | Responsibilities |
|---------|-----------------|
| **GamesService** | Move validation, persistence, calls SyncService.Broadcast |
| **SyncService** | Pure pub/sub: Subscribe, Broadcast, sequence numbers |
| **lib/** | RNG/seed management (blinded seed for anti-cheat) |

Services communicate via gRPC clients, allowing them to run on different hosts.

### 2. Works Without Presenter

The sync operates at the `ProcessMoves` level, not the presenter level:
- `ProcessMoves` generates WorldChanges (the canonical diff)
- `SyncService.Broadcast` fans out to all subscribers
- Any client (presenter, CLI, API) can receive and apply them

### 3. Two Paths: Verify vs Apply

| Client Role | What Happens |
|-------------|--------------|
| **Originator** | Changes already applied locally → just verify server response matches |
| **Viewer** | Receive WorldChanges → apply to local state + trigger UI updates |

### 4. Sequence Numbers for Reconnection

Each `GameUpdate` has a monotonic sequence number. Clients track the last seen sequence and can resume from that point on reconnect.

### 5. gocurrent FanOut for Efficient Broadcasting

Uses `github.com/panyam/gocurrent` FanOut primitive:
- One FanOut per game
- Subscribers get their own output channel
- Non-blocking sends via goroutines
- Automatic cleanup on disconnect

## RNG and Seed Management (lib/)

The lib package handles all RNG operations for deterministic, reproducible gameplay.

### Current Implementation

**Location**: `lib/game.go`, `lib/combat.go`

```go
// Game struct holds the RNG state
type Game struct {
    Seed int64      `json:"seed"` // Random seed for deterministic gameplay
    rng  *rand.Rand              // RNG instance seeded from Seed
    // ...
}

// NewGame initializes RNG from seed
func NewGame(..., seed int64) *Game {
    return &Game{
        Seed: seed,
        rng:  rand.New(rand.NewSource(seed)),
        // ...
    }
}
```

**How RNG is used**:
1. Combat damage rolls: `RulesEngine.CalculateCombatDamage(attackerID, defenderID, game.rng)`
2. Damage distribution sampling: `rollDamageFromDistribution(dist, rng)` uses `rng.Float64()`

**Key files**:
| File | Purpose |
|------|---------|
| `lib/game.go:27-30` | Game.Seed and Game.rng fields |
| `lib/game.go:49` | RNG initialization from seed |
| `lib/combat.go:12-24` | CalculateCombatDamage uses RNG |
| `lib/combat.go:42-74` | rollDamageFromDistribution samples from probability ranges |

### Client Usage

**For local-first (WASM) clients**:
```go
// Create game with initial seed (from server or game creation)
game := lib.NewGame(gameProto, stateProto, world, rulesEngine, initialSeed)

// ProcessMoves uses game.rng internally for combat
game.ProcessMoves(moves)

// RNG state advances deterministically - same seed + same moves = same outcomes
```

**For multiplayer sync**:
1. Server provides initial seed when game is created
2. Both client and server use the same seed
3. Same sequence of moves produces identical RNG draws
4. If outcomes differ, server's result is authoritative (client reloads)

### Current Limitations

The current implementation uses a simple shared seed. This means:
- All players know the seed (stored in game state)
- A malicious client could pre-calculate outcomes to cherry-pick moves
- Acceptable for trusted environments (same device, friends playing together)

## RNG Security: Blinded Seed (Planned)

To prevent cheating in competitive multiplayer, we plan to implement a **blinded seed** approach.

**Note**: This is a future enhancement, not yet implemented.

### The Problem

If clients know the RNG seed before committing to an action, they can:
- Pre-calculate outcomes for all possible actions
- Choose the action with the most favorable result

### The Planned Solution

```
1. Server has a game secret (never revealed)
2. For each action: seed = SHA256(secret || game_id || group_number || move_index)
3. Client computes RNG locally for immediate feedback
4. Client reports RNG values consumed in ProcessMoves request
5. Server recomputes with blinded seed and verifies values match
```

**Why it works:**
- Client can't predict the seed because it depends on action metadata (group_number, move_index) that doesn't exist until they commit
- Server can verify the client used the correct RNG sequence
- Deterministic: same action always produces same outcome

### Example Flow

```
1. Client commits: "Attack unit A1 → B2" (this becomes move_index=0 in group_42)
                        ↓
2. Server computes: seed = SHA256(secret || "game123" || 42 || 0)
                        ↓
3. Server validates: RNG(seed) produces [0.73, 0.21] → matches client's reported values?
                        ↓
4. If match: Accept. If not: Reject (cheating or bug)
```

## Proto Definitions

### Service (`protos/lilbattle/v1/services/sync.proto`)

```protobuf
service GameSyncService {
  // Subscribe to game updates (streaming)
  rpc Subscribe(SubscribeRequest) returns (stream GameUpdate);

  // Broadcast update to all subscribers (called by GamesService)
  rpc Broadcast(BroadcastRequest) returns (BroadcastResponse);
}
```

### Messages (`protos/lilbattle/v1/models/sync.proto`)

Key messages:
- `SubscribeRequest`: game_id, player_id, from_sequence
- `GameUpdate`: sequence + oneof (MovesPublished, PlayerJoined, PlayerLeft, GameEnded, InitialState)
- `BroadcastRequest`: game_id, update (GameUpdate)
- `BroadcastResponse`: subscriber_count, sequence

## Implementation Status

### Completed
- [x] Phase 1: Define sync.proto (GameSync service)
- [x] Phase 2: Create sync_service.go (using gocurrent FanOut)
- [x] Phase 3: Add ApplyRemoteChanges RPC to presenter.proto
- [x] Phase 4: Implement ApplyRemoteChanges in gameview_presenter.go
- [x] Phase 5: Register GameSync service in grpcserver.go
- [x] Phase 6: Integrate GamesService → SyncService.Broadcast (via OnMovesSaved callback)
- [x] Phase 7a: Add GameSyncClient to ClientMgr
- [x] Phase 7b: Create Connect adapter for GameSyncService
- [x] Phase 7c: Register HTTP/Connect handlers for sync
- [x] Phase 7d: Create GameSyncManager.ts for frontend

### Pending
- [ ] Integration testing with multiple browser tabs
- [ ] Connection status UI indicator

## Future: Commit-Reveal Protocol

For fully trustless/peer-to-peer scenarios (serverless/local-first vision), we may implement commit-reveal:

### Why Commit-Reveal?

Blinded seed trusts the server not to leak the secret. Commit-reveal is cryptographically trustless - useful for:
- Peer-to-peer games without central server
- Serverless architecture where all clients validate
- High-stakes competitive play

### How It Works

```
Phase 1: COMMIT (both sides lock in, hidden)
┌─────────────────────────────────────────────────────────┐
│ Client: hash(action + client_nonce) → sends commitment  │
│ Server: hash(seed + server_nonce) → sends commitment    │
└─────────────────────────────────────────────────────────┘
                           ↓
Phase 2: REVEAL (both sides reveal, verify against commitment)
┌─────────────────────────────────────────────────────────┐
│ Client: reveals action + client_nonce                   │
│ Server: reveals seed + server_nonce                     │
│ Both verify: hash matches commitment? ✓                 │
└─────────────────────────────────────────────────────────┘
                           ↓
Phase 3: COMPUTE (deterministic outcome)
┌─────────────────────────────────────────────────────────┐
│ final_seed = hash(client_nonce || server_nonce)         │
│ outcome = RNG(final_seed).apply(action)                 │
└─────────────────────────────────────────────────────────┘
```

### Security Guarantees
- Client can't change action after seeing server's seed (already committed)
- Server can't change seed after seeing client's action (already committed)
- Final seed depends on BOTH nonces, so neither party controls it

### Trade-offs vs Blinded Seed

| | Blinded Seed | Commit-Reveal |
|---|---|---|
| **Round trips** | 1 | 2 |
| **Complexity** | Simple | More complex |
| **Trust model** | Trust server | Trustless (cryptographic) |
| **Latency** | Lower | Higher |
| **Use case** | Server-authoritative | Peer-to-peer / local-first |

## Files Reference

| File | Purpose |
|------|---------|
| `protos/lilbattle/v1/services/sync.proto` | Service definition |
| `protos/lilbattle/v1/models/sync.proto` | Message definitions |
| `services/sync_service.go` | Server implementation (pure pub/sub) |
| `services/games_service.go` | BaseGamesService with OnMovesSaved callback |
| `services/backend_games_service.go` | InitializeSyncBroadcast() sets up broadcast |
| `services/clientmgr.go` | GetGameSyncSvcClient() for gRPC client |
| `services/gameview_presenter.go` | ApplyRemoteChanges for viewer updates |
| `web/server/connect.go` | ConnectGameSyncServiceAdapter |
| `web/server/api.go` | GameSyncService HTTP handler registration |
| `cmd/backend/main.go` | GameSyncService gRPC registration |
| `lib/game.go` | Game struct with Seed field, RNG initialization |
| `lib/combat.go` | Combat damage using RNG for deterministic rolls |
| `lib/changes.go` | Existing ApplyChanges for WorldChange application |
| `web/pages/GameViewerPage/GameSyncManager.ts` | Frontend sync manager |
| `web/pages/GameViewerPage/GameViewerPageBase.ts` | Base page with sync integration |
| `web/gen/wasmjs/.../gameSyncServiceClient.ts` | Generated WASM service client |

## Integration Points

### GamesService Integration

After ProcessMoves succeeds, GamesService should call SyncService.Broadcast:

```go
// In GamesService.ProcessMoves, after SaveMoveGroup succeeds:
if s.syncClient != nil {
    s.syncClient.Broadcast(ctx, &v1.BroadcastRequest{
        GameId: req.GameId,
        Update: &v1.GameUpdate{
            UpdateType: &v1.GameUpdate_MovesPublished{
                MovesPublished: &v1.MovesPublished{
                    Player:      currentPlayer,
                    Moves:       resp.Moves,
                    GroupNumber: groupNumber,
                },
            },
        },
    })
}
```

### Frontend Integration

The `GameSyncManager` class handles sync for multiplayer games:

```typescript
// GameSyncManager is integrated into GameViewerPageBase
// Enable sync via URL parameter: ?sync=true

// Or programmatically:
protected isMultiplayerSyncEnabled(): boolean {
    return true; // For multiplayer games
}
```

**Flow:**
1. On game load: `GameSyncManager.connect()` subscribes to GameSyncService
2. On local move: ProcessMoves processes locally, server broadcasts to others
3. On remote update: GameSyncManager calls `presenter.applyRemoteChanges()`
4. On disconnect: Auto-reconnect with `from_sequence` for missed updates

**Key files:**
- `GameSyncManager.ts`: Handles subscription, reconnection, state tracking
- `GameViewerPageBase.ts`: Integrates sync manager, provides callbacks
