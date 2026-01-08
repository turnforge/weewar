# Performance Issues Backlog

This document contains detailed issue descriptions for remaining performance optimizations.
Use these to create GitHub issues when ready.

---

## Critical Priority

### Issue: Combat simulation has exponential iterations with splash damage

**Labels:** `performance`, `backend`, `critical`

**Location:** `lib/combat_formula.go:99-106, 319-325`

**Current Behavior:**
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

**Impact:**
- SplashDamage=3, Health=10, 3 adjacent units = 540 iterations per attack
- Noticeable lag on attacks with high-health units and splash damage

**Suggested Fixes:**
1. **Mathematical probability** - Replace dice simulation with binomial distribution calculation
2. **Single simulation with multiplier** - Run once, apply variance for splash
3. **Pre-computed tables** - Cache damage distributions at startup

---

## High Priority

### Issue: Object allocation in game loop - hex vertices recreated each frame

**Labels:** `performance`, `frontend`, `high`

**Location:** `web/pages/common/PhaserWorldScene.ts:1566-1607`

**Current Behavior:**
```typescript
protected drawHexagon(q: number, r: number) {
    const vertices = [
        { x: position.x, y: position.y - halfHeight },
        // 6 object allocations per hex per frame
    ];
}
```

**Impact:**
- 6 objects created per hex per frame
- For 100 visible hexes at 60fps = 36,000 allocations/second
- Causes GC pressure and frame drops

**Suggested Fixes:**
1. **Static vertex arrays** - Pre-compute vertex offsets once, reuse
2. **Object pooling** - Reuse vertex objects from pool
3. **Typed arrays** - Use Float32Array instead of object literals

---

### Issue: Event listener leaks in Phaser scenes

**Labels:** `performance`, `frontend`, `memory-leak`, `high`

**Location:** `web/pages/WorldEditorPage/PhaserEditorScene.ts:86-127, 479, 493`

**Current Behavior:**
```typescript
this.input.on('pointermove', ...);
this.input.on('pointerout', ...);
// No corresponding off() in destroy()

// Re-registration on each callback set
this.events.on('referenceScaleChanged', ...);  // Duplicates each call
```

**Impact:**
- Memory grows over time as listeners accumulate
- Callbacks fire multiple times after scene recreation
- Eventually causes performance degradation

**Suggested Fixes:**
1. **Add cleanup in destroy()** - Call `this.input.off()` for all registered listeners
2. **Use `once()` where appropriate** - For one-time events
3. **Track registrations** - Prevent duplicate subscriptions with a Set

---

### Issue: Unit array linear removal is O(n)

**Labels:** `performance`, `backend`, `high`

**Location:** `lib/world.go:564-570, 626-632`

**Current Behavior:**
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

**Impact:**
- O(n) scan for every unit removal
- O(n) slice copy for removal
- With 50 units per player, up to 100 operations per removal

**Suggested Fixes:**
1. **Map-based index** - `unitIndexByPlayer map[playerID]map[unitPtr]int` for O(1) lookup
2. **Swap-and-pop** - Swap with last element, truncate slice (O(1) but changes order)
3. **Linked list** - If order matters and frequent removal needed

---

### Issue: Attack range generates O(r²) coordinates

**Labels:** `performance`, `backend`, `high`

**Location:** `lib/combat.go:94-97`, `lib/hex_coords.go:113-125`

**Current Behavior:**
```go
coordsInRange := unitCoord.Range(int(attackRange))
// For range 5: generates 61 coordinates even if only 3 have units
```

**Impact:**
- Range 5 = 61 coordinates allocated and checked
- Most coordinates are empty (no units)
- Wasted iteration over empty hexes

**Suggested Fixes:**
1. **Iterate units directly** - Loop through enemy units, check if in range
2. **Spatial hash** - Bucket units by grid region for faster lookups
3. **Early termination** - Stop if enough targets found

---

### Issue: Multiple presigned URL API calls per file

**Labels:** `performance`, `backend`, `r2`, `high`

**Location:** `services/r2/filestore.go:50-66, 196-214`

**Current Behavior:**
```go
// 3 API calls per file
s.Client.GetPresignedURL(ctx, file.Path, 15*time.Minute)
s.Client.GetPresignedURL(ctx, file.Path, time.Hour)
s.Client.GetPresignedURL(ctx, file.Path, 24*time.Hour)

// Listing 100 files = 300 API calls!
```

**Impact:**
- 3x API calls needed
- Slow file listing operations
- Potential rate limiting issues

**Suggested Fixes:**
1. **Lazy generation** - Only generate URLs when accessed
2. **Cache URLs** - Store with TTL slightly less than expiry
3. **Single long-lived URL** - Use one duration, refresh proactively

---

## Medium Priority

### Issue: String key formatting in hot paths

**Labels:** `performance`, `backend`, `medium`

**Location:** `lib/rules_engine.go:636`, `lib/hex_coords.go:290-291`, `lib/world.go:380`

**Current Behavior:**
```go
key := fmt.Sprintf("%d,%d", coord.Q, coord.R)  // In Dijkstra inner loop
```

**Impact:**
- String allocation on every coordinate lookup
- fmt.Sprintf is relatively slow
- Called thousands of times per pathfinding

**Suggested Fixes:**
1. **Integer key** - `key := int64(Q)*10000 + int64(R)` (assumes R < 10000)
2. **Struct key** - Use `HexCoord` directly as map key (needs hash)
3. **Pre-compute** - Store string keys in coordinate struct

---

### Issue: Full file reads for partial updates

**Labels:** `performance`, `backend`, `medium`

**Location:** `services/fsbe/games_service.go:62-97`

**Current Behavior:**
```go
// Reads entire GameState to get single Version field
gameState, err := storage.LoadFSArtifact[*v1.GameState](...)
return gameState.Version, nil
```

**Impact:**
- Reads full JSON file just for one field
- GameState can be large with many units/tiles

**Suggested Fixes:**
1. **Separate version file** - Store version in `.version` sidecar file
2. **JSON streaming** - Use streaming parser to extract single field
3. **Header extraction** - Put version at start of file for partial read

---

### Issue: Income calculation scans all tiles

**Labels:** `performance`, `backend`, `medium`

**Location:** `lib/moves.go:441-447`

**Current Behavior:**
```go
for _, tile := range g.World.TilesByCoord() {  // ALL tiles
    if tile.Player == previousPlayer {
        totalIncome += GetTileIncomeFromConfig(...)
    }
}
```

**Impact:**
- Scans every tile on map each turn
- Most tiles don't belong to player
- 100x100 map = 10,000 iterations

**Suggested Fixes:**
1. **GetPlayerTiles()** - Add method returning only player's tiles
2. **Maintain count** - Track income-generating tiles per player
3. **Incremental update** - Update income when tiles change ownership

---

### Issue: Excessive unit copying with immediate overwrites

**Labels:** `performance`, `backend`, `medium`

**Location:** `lib/moves.go:606-608, 788-789`

**Current Behavior:**
```go
updatedUnit := copyUnit(movedUnit)  // Deep copies AttackHistory
updatedUnit.LastActedTurn = unit.LastActedTurn  // Overwrites
```

**Impact:**
- Deep copy of AttackHistory slice (could be large)
- Immediately overwrites fields that were just copied

**Suggested Fixes:**
1. **copyUnitWithFields()** - Copy with field overrides in one step
2. **Shallow copy + selective deep** - Only deep copy what's needed
3. **Builder pattern** - `NewUnitFrom(old).WithLastActedTurn(t).Build()`

---

### Issue: Multiple passes in buildIndexes

**Labels:** `performance`, `backend`, `medium`

**Location:** `lib/world.go:115-176`

**Current Behavior:**
```go
// Pass 1: Track tile shortcuts
// Pass 2: Generate missing tile shortcuts
// Pass 3: Track unit shortcuts
// Pass 4: Generate unit shortcuts + build index
```

**Impact:**
- 4 separate iterations over data
- Could be combined into fewer passes

**Suggested Fixes:**
1. **Two-pass** - Collect existing, then generate missing in second pass
2. **Single pass with deferred** - Track what needs generation, do at end
3. **Lazy shortcuts** - Generate shortcuts on-demand

---

### Issue: Sequential database queries could be parallel

**Labels:** `performance`, `backend`, `database`, `medium`

**Location:** `services/gormbe/games_service.go:378-384`

**Current Behavior:**
```go
game, err = s.GameDAL.Get(...)
state, err = s.GameStateDAL.Get(...)      // Waits for game
moves, err = s.GameMoveDAL.List(...)      // Waits for state
```

**Impact:**
- 3 sequential round trips to database
- Total latency = sum of all queries

**Suggested Fixes:**
1. **Goroutines** - Run queries in parallel with errgroup
2. **Database join** - Single query with joins
3. **Batch API** - Create GetGameWithState() method

---

### Issue: Camera events fire every frame during movement

**Labels:** `performance`, `frontend`, `medium`

**Location:** `web/pages/common/PhaserWorldScene.ts:832-879`

**Current Behavior:**
```typescript
update() {
    if (positionChanged) {
        this.events.emit('camera-moved', {...});  // Triggers downstream updates
    }
}
```

**Impact:**
- Event emitted every frame during pan/zoom
- Triggers cascading updates downstream
- 60 events per second during movement

**Suggested Fixes:**
1. **Debounce** - Emit at most once per 100ms
2. **Batch updates** - Collect changes, emit once at end of frame
3. **Dirty flag** - Set flag, check in requestAnimationFrame

---

### Issue: Direction recalculated in every sort comparison

**Labels:** `performance`, `backend`, `medium`

**Location:** `lib/sort_utils.go:56-71`

**Current Behavior:**
```go
sort.Slice(items, func(i, j int) bool {
    dirA := GetDirection(fromA, toA)  // Recalculated O(n log n) times
    dirB := GetDirection(fromB, toB)
    return dirA < dirB
})
```

**Impact:**
- GetDirection called O(n log n) times during sort
- Same direction calculated multiple times for same item

**Suggested Fixes:**
1. **Pre-compute** - Calculate directions before sort, store in struct
2. **Schwartzian transform** - Map to (item, key), sort by key, map back
3. **Cached getter** - Memoize GetDirection results

---

### Issue: Wound bonus scans full attack history every combat

**Labels:** `performance`, `backend`, `medium`

**Location:** `lib/combat_formula.go:191-239`

**Current Behavior:**
```go
for _, attack := range defender.AttackHistory {  // Full scan every combat
    // ... calculations
}
```

**Impact:**
- History grows throughout game
- Late-game units may have 20+ attacks in history
- Scanned on every combat calculation

**Suggested Fixes:**
1. **Cache wounds per turn** - Store computed wound count, reset each turn
2. **Rolling window** - Only keep last N attacks in history
3. **Pre-computed field** - Add `WoundsThisTurn` field to Unit

---

### Issue: Excessive console.log statements in production

**Labels:** `performance`, `frontend`, `medium`

**Location:** Multiple frontend files (100+ statements)

**Current Behavior:**
```typescript
if (this.debug) {
    console.log('[AnimationQueue] Enqueuing animation');
}
```

**Impact:**
- String formatting even when debug=false
- Console calls have overhead
- Clutters browser console

**Suggested Fixes:**
1. **Build-time removal** - Use terser to strip in production
2. **Log levels** - Proper logger with level check before format
3. **Conditional compilation** - `if (process.env.NODE_ENV === 'development')`

---

### Issue: Hexagon points allocated fresh each highlight

**Labels:** `performance`, `frontend`, `medium`

**Location:** `web/pages/common/HexHighlightLayer.ts:57-68`

**Current Behavior:**
```typescript
const points: Phaser.Geom.Point[] = [];
points.push(new Phaser.Geom.Point(...));  // 6 allocations per highlight
```

**Impact:**
- 6 Point objects per highlight
- Highlights updated frequently during gameplay
- Adds to GC pressure

**Suggested Fixes:**
1. **Static point array** - Reuse same 6 points, update coordinates
2. **Number array** - Use `[x1,y1,x2,y2,...]` instead of Point objects
3. **Pre-computed offsets** - Store relative offsets, add position at draw time

---

## Low Priority

### Issue: Path reversal after construction

**Labels:** `performance`, `backend`, `low`

**Location:** `lib/rules_engine.go:464-490`

**Current Behavior:**
```go
// Builds path in reverse, then reverses entire slice
for i := 0; i < len(pathEdges)/2; i++ {
    pathEdges[i], pathEdges[j] = pathEdges[j], pathEdges[i]
}
```

**Impact:**
- Extra O(n) pass to reverse
- Path typically short (< 20 edges)

**Suggested Fixes:**
1. **Build forwards** - Traverse from start to end instead
2. **Deque** - Use double-ended queue, prepend elements
3. **Accept reverse** - If consumers can handle reversed order

---

### Issue: GORM Updates() sends all columns

**Labels:** `performance`, `backend`, `database`, `low`

**Location:** `services/gormbe/genid.go:93, 108`

**Current Behavior:**
```go
gid.VerifiedAt = time.Now()
return storage.Updates(gid).Error  // Updates ALL fields
```

**Impact:**
- Sends unchanged columns over wire
- Larger UPDATE statement
- Minor network overhead

**Suggested Fixes:**
1. **Selective update** - `Updates(map[string]interface{}{"verified_at": time.Now()})`
2. **Select clause** - `Select("verified_at").Updates(gid)`
3. **Raw SQL** - For simple single-column updates

---

### Issue: Disabled cache maps still allocated

**Labels:** `performance`, `backend`, `low`

**Location:** `services/fsbe/games_service.go:46-49`

**Current Behavior:**
```go
gameCache:    make(map[string]*v1.Game),     // Allocated but unused
stateCache:   make(map[string]*v1.GameState),
historyCache: make(map[string]*v1.GameMoveHistory),
```

**Impact:**
- Small memory waste (empty maps)
- Confusion about whether cache is enabled

**Suggested Fixes:**
1. **Remove if unused** - Delete cache fields entirely
2. **Re-enable cache** - If caching was disabled for debugging
3. **Lazy init** - Only allocate when cache is enabled

---

## Summary

| Priority | Count | Categories |
|----------|-------|------------|
| Critical | 1 | Combat simulation |
| High | 5 | Memory, allocations, API calls |
| Medium | 11 | Various optimizations |
| Low | 3 | Minor improvements |
| **Total** | **20** | |
