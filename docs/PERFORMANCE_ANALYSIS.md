# Performance Analysis Report

This document catalogs performance anti-patterns, N+1 queries, unnecessary re-renders, and inefficient algorithms found in the codebase.

## Executive Summary

| Category | Critical | High | Medium | Low |
|----------|----------|------|--------|-----|
| Backend (Go) | 2 | 3 | 6 | 3 |
| Frontend (TypeScript/Phaser) | 3 | 3 | 3 | 1 |
| Database/Storage | 4 | 2 | 2 | 0 |
| **Total** | **9** | **8** | **11** | **4** |

### Fixed Issues

- [x] **#1 Dijkstra Priority Queue** - PR #51: Replaced O(n²) array with heap-based O(n log n)
- [x] **#2 Grid/Coordinate Display** - PR #57: Camera tracking, diff-based updates, object pooling
- [x] **#3 Disabled File Cache** - PR #56: Refactored caching to BackendGamesService level
- [x] **#4 N+1 Database Inserts** - PR #54: Batch insert for game moves
- [x] **#6 Missing Database Indexes** - Added GORM tags for indexes on game_moves and index_states
- [x] **#12 Multiple Filter Passes** - PR #53: Single-pass categorization
- [x] **#24 Unbounded Slice Growth** - PR #52: Pre-allocate slices in hot paths

---

## Critical Issues

### 1. ~~O(n^2) Dijkstra Priority Queue~~ [FIXED - PR #51]
**File:** `lib/rules_engine.go:566-578`

The pathfinding algorithm uses a linear array for the priority queue, resulting in O(n) extraction per iteration instead of O(log n).

```go
popMinCoord := func() queueItem {
    minIdx := 0
    for i := 1; i < len(queue); i++ {
        if queue[i].cost < queue[minIdx].cost {
            minIdx = i
        }
    }
    current := queue[minIdx]
    queue = append(queue[:minIdx], queue[minIdx+1:]...)  // O(n) removal
    return current
}
```

**Impact:** Every movement option query and pathfinding operation has O(n^2) complexity instead of O(n log n).

**Fix:** Use `container/heap` package for min-heap implementation.

---

### 2. ~~Grid/Coordinate Display Rebuilt Every Frame~~ [FIXED - PR #57]
**File:** `web/pages/common/PhaserWorldScene.ts:1529-1564`

`updateCoordinatesDisplay()` is called every frame and destroys/recreates ALL coordinate texts:

```typescript
private updateCoordinatesDisplay() {
    this.coordinateTexts.forEach(text => text.destroy());  // Destroys all
    this.coordinateTexts.clear();

    for (let q = minQ; q <= maxQ; q++) {
        for (let r = minR; r <= maxR; r++) {
            this.updateCoordinateText(q, r);  // Recreates all
        }
    }
}
```

**Impact:** For a 40x40 visible area, 1600+ text objects created and destroyed per frame causing severe GC pressure.

**Fix:** Only update texts that changed position or came into/left view.

---

### 3. ~~Disabled File Cache~~ [FIXED - PR #56]
**File:** `services/fsbe/games_service.go:134-147`

The cache is explicitly disabled with `if false`:

```go
if false {  // CACHE DISABLED!
    if game, ok := s.gameCache[req.Id]; ok {
        // Never executes
    }
}
```

**Impact:** Every `GetGame()` call performs 3 separate file reads (metadata, state, history) even for the same game.

**Fix:** Re-enable cache with proper invalidation strategy.

---

### 4. ~~N+1 Individual Database Inserts~~ [FIXED - PR #54]
**File:** `services/gormbe/genid.go:66-80` and `services/gormbe/games_service.go:403-427`

ID generation and move saving use individual inserts in loops:

```go
// Individual ID inserts
for i := range numids {
    err := storage.Create(gid).Error  // One insert per ID
}

// Individual move inserts
for i, move := range group.Moves {
    if err := s.storage.Create(moveGorm).Error {  // One insert per move
```

**Impact:** Creating 100 IDs = 100 database round trips. Saving 10 moves = 10 inserts.

**Fix:** Use batch inserts with `CreateInBatches()`.

---

### 5. Combat Simulation Exponential Iterations
**File:** `lib/combat_formula.go:99-106, 319-325`

Combat runs dice simulation loops, and splash damage multiplies this:

```go
for range ctx.AttackerHealth {     // Up to 30 iterations
    for range 6 {                   // 6 dice rolls each
        roll := rng.Float64()
    }
}

// Splash damage runs FULL simulation per count
for i := int32(0); i < attackerDef.SplashDamage; i++ {
    damage, err := re.SimulateCombatDamage(ctx, rng)  // Full simulation!
}
```

**Impact:** SplashDamage=3, Health=10, 3 adjacent = 540 iterations per attack.

**Fix:** Use mathematical probability calculation or accumulate from single simulation.

---

### 6. ~~Missing Database Indexes~~ [FIXED - In Progress]
**File:** `protos/weewar/v1/gorm/models.proto` and `protos/weewar/v1/gorm/indexer.proto`

Multiple queries on unindexed columns:

```go
// IndexState queries
query.Where("entity_type = ?", ...).Where("indexed_at lte ?", ...)

// GameMove queries
s.storage.Where("game_id = ?", gameId).Order("group_number asc")
```

**Missing indexes:**
- `(entity_type, entity_id)` composite
- `indexed_at` for range queries
- `(game_id)` on GameMove
- `(game_id, group_number)` composite

**Impact:** Full table scans on every query.

---

## High Priority Issues

### 7. Object Allocation in Game Loop
**File:** `web/pages/common/PhaserWorldScene.ts:1566-1607`

Hexagon vertex arrays created for every hex every frame:

```typescript
protected drawHexagon(q: number, r: number) {
    const vertices = [
        { x: position.x, y: position.y - halfHeight },
        // 6 object allocations per hex per frame
    ];
}
```

**Fix:** Cache vertex positions or use object pooling.

---

### 8. Event Listener Leaks
**File:** `web/pages/WorldEditorPage/PhaserEditorScene.ts:86-127, 479, 493`

Event listeners registered without cleanup:

```typescript
this.input.on('pointermove', ...);
this.input.on('pointerout', ...);
// No corresponding off() in destroy()

// Re-registration on each callback set
this.events.on('referenceScaleChanged', ...);  // Duplicates each call
```

**Fix:** Add cleanup in `destroy()` method, prevent duplicate subscriptions.

---

### 9. Unit Array Linear Removal
**File:** `lib/world.go:564-570, 626-632`

Unit removal from player array is O(n):

```go
for i, u := range w.unitsByPlayer[oldPlayerID] {
    if u == oldunit {
        w.unitsByPlayer[oldPlayerID] = append(
            w.unitsByPlayer[oldPlayerID][:i],
            w.unitsByPlayer[oldPlayerID][i+1:]...)
        break
    }
}
```

**Fix:** Use map-based index for O(1) lookups/removal.

---

### 10. Attack Range O(r^2) Allocation
**File:** `lib/combat.go:94-97`, `lib/hex_coords.go:113-125`

Range() generates ALL coordinates within radius:

```go
coordsInRange := unitCoord.Range(int(attackRange))
// For range 5: generates 61 coordinates even if only 3 have units
```

**Fix:** Iterate neighbors directly or break early when found.

---

### 11. Multiple Presigned URL API Calls
**File:** `services/r2/filestore.go:50-66, 196-214`

Each file generates 3 separate R2 API calls for presigned URLs:

```go
// 3 API calls per file
s.Client.GetPresignedURL(ctx, file.Path, 15*time.Minute)
s.Client.GetPresignedURL(ctx, file.Path, time.Hour)
s.Client.GetPresignedURL(ctx, file.Path, 24*time.Hour)

// Listing 100 files = 300 API calls!
```

**Fix:** Cache presigned URLs or generate lazily.

---

### 12. ~~Multiple Filter/Map Array Passes~~ [FIXED - PR #53]
**File:** `web/pages/GameViewerPage/PhaserGameScene.ts:233-238`

Six separate filter operations on same array:

```typescript
const selections = highlights.filter(h => h.type === 'selection');
const movements = highlights.filter(h => h.type === 'movement').map(...);
const attacks = highlights.filter(h => h.type === 'attack').map(...);
const captures = highlights.filter(h => h.type === 'capture');
const exhausted = highlights.filter(h => h.type === 'exhausted');
const capturing = highlights.filter(h => h.type === 'capturing');
```

**Fix:** Single reduce() to categorize all at once.

---

## Medium Priority Issues

### 13. String Key Formatting in Hot Paths
**Files:** `lib/rules_engine.go:636`, `lib/hex_coords.go:290-291`, `lib/world.go:380`

```go
key := fmt.Sprintf("%d,%d", coord.Q, coord.R)  // In Dijkstra inner loop
```

**Fix:** Use integer-based keying (e.g., `Q*10000 + R`) or struct keys.

---

### 14. Full File Reads for Partial Updates
**File:** `services/fsbe/games_service.go:62-97`

```go
// Reads entire GameState to get single Version field
gameState, err := storage.LoadFSArtifact[*v1.GameState](...)
return gameState.Version, nil
```

**Fix:** Store version separately or use partial read.

---

### 15. Income Calculation Scans All Tiles
**File:** `lib/moves.go:441-447`

```go
for _, tile := range g.World.TilesByCoord() {  // ALL tiles
    if tile.Player == previousPlayer {
        totalIncome += GetTileIncomeFromConfig(...)
    }
}
```

**Fix:** Add `GetPlayerTiles()` method.

---

### 16. Excessive Unit Copying
**File:** `lib/moves.go:606-608, 788-789`

Full deep copy then immediate field overwrites:

```go
updatedUnit := copyUnit(movedUnit)  // Deep copies AttackHistory
updatedUnit.LastActedTurn = unit.LastActedTurn  // Overwrites
```

**Fix:** Create `copyUnitWithFields()` or use struct literals.

---

### 17. Multiple Passes in buildIndexes
**File:** `lib/world.go:115-176`

Four separate passes over tiles and units:

```go
// Pass 1: Track tile shortcuts
// Pass 2: Generate missing tile shortcuts
// Pass 3: Track unit shortcuts
// Pass 4: Generate unit shortcuts + build index
```

**Fix:** Combine into single pass per entity type.

---

### 18. Sequential Database Queries
**File:** `services/gormbe/games_service.go:378-384`

```go
game, err = s.GameDAL.Get(...)
state, err = s.GameStateDAL.Get(...)      // Waits for game
moves, err = s.GameMoveDAL.List(...)      // Waits for state
```

**Fix:** Use goroutines or database joins.

---

### 19. Camera Events Every Frame
**File:** `web/pages/common/PhaserWorldScene.ts:832-879`

```typescript
update() {
    if (positionChanged) {
        this.events.emit('camera-moved', {...});  // Triggers downstream updates
    }
}
```

**Fix:** Debounce camera events or batch updates.

---

### 20. Direction Recalculation in Sort
**File:** `lib/sort_utils.go:56-71`

Direction calculated fresh for every comparison during sort:

```go
dirA := GetDirection(fromA, toA)  // Recalculated O(n log n) times
```

**Fix:** Pre-compute directions before sorting.

---

### 21. Wound Bonus Full History Scan
**File:** `lib/combat_formula.go:191-239`

```go
for _, attack := range defender.AttackHistory {  // Full scan every combat
    // ... calculations
}
```

**Fix:** Cache wounds per turn, reset each turn.

---

### 22. Excessive Console.log Statements
**Files:** Multiple frontend files (100+ statements)

```typescript
if (this.debug) {
    console.log('[AnimationQueue] Enqueuing animation');
}
```

**Fix:** Use conditional compilation or log levels.

---

### 23. Hexagon Points Allocation
**File:** `web/pages/common/HexHighlightLayer.ts:57-68`

```typescript
const points: Phaser.Geom.Point[] = [];
points.push(new Phaser.Geom.Point(...));  // 6 allocations per highlight
```

**Fix:** Reuse static point arrays.

---

## Low Priority Issues

### 24. ~~Unbounded Slice Growth~~ [FIXED - PR #52]
**File:** `lib/moves.go:841-851`

```go
var adjacentUnits []*v1.Unit  // Starts at capacity 0
for _, coord := range adjacentCoords {
    adjacentUnits = append(adjacentUnits, unit)  // Reallocations
}
```

**Fix:** Pre-allocate with `make([]*v1.Unit, 0, 6)`.

---

### 25. Path Reversal After Construction
**File:** `lib/rules_engine.go:464-490`

```go
// Builds path in reverse, then reverses entire slice
for i := 0; i < len(pathEdges)/2; i++ {
    pathEdges[i], pathEdges[j] = pathEdges[j], pathEdges[i]
}
```

**Fix:** Build path forwards or use deque.

---

### 26. GORM Updates() All Columns
**File:** `services/gormbe/genid.go:93, 108`

```go
gid.VerifiedAt = time.Now()
return storage.Updates(gid).Error  // Updates ALL fields
```

**Fix:** Use `Updates(map[string]interface{}{...})`.

---

### 27. Disabled Cache Still Allocates
**File:** `services/fsbe/games_service.go:46-49`

```go
gameCache:    make(map[string]*v1.Game),     // Allocated but unused
stateCache:   make(map[string]*v1.GameState),
historyCache: make(map[string]*v1.GameMoveHistory),
```

**Fix:** Remove or re-enable.

---

## Recommended Priority Actions

### Immediate (High Impact)
1. **Replace Dijkstra queue with heap** - Biggest pathfinding gain
2. **Enable FSBE cache** - Eliminate redundant file reads
3. **Add database indexes** - Fix full table scans
4. **Batch database inserts** - Eliminate N+1 patterns

### Short Term (Architecture)
5. **Fix grid display** - Only update changed coordinates
6. **Add event listener cleanup** - Prevent memory leaks
7. **Use map-based unit index** - O(1) unit removal
8. **Cache combat calculations** - Reduce simulation iterations

### Medium Term (Optimization)
9. **Pre-allocate slices** - Reduce GC pressure
10. **Use integer coordinate keys** - Faster lookups
11. **Batch presigned URL generation** - Reduce API calls
12. **Single-pass array categorization** - Reduce iterations

---

## Testing Recommendations

1. **Benchmark pathfinding** with maps of varying sizes (10x10, 50x50, 100x100)
2. **Profile combat** with high-health units and splash damage
3. **Monitor frame rate** during camera pan on large maps
4. **Load test** game listing with 1000+ games
5. **Memory profile** long-running game sessions for leaks
