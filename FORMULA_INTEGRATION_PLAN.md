# Formula-Based Combat Integration Plan

## Current Architecture

### Combat Flow
```
CLI/UI Request
    ↓
Presenter.SceneClicked()
    ↓
SingletonGameViewPresenterImpl.ProcessSceneClicked()
    ↓
MoveProcessor.ProcessAttackUnit()
    ↓
RulesEngine.CalculateCombatDamage() [TABLE-BASED]
    ↓
Game State Updated
```

### Key Files
- **services/moves.go**: `ProcessAttackUnit()` - executes combat, applies damage
- **services/combat.go**: `CalculateCombatDamage()` - table-based lookup
- **services/combat_formula.go**: NEW - formula-based calculation
- **cmd/cli/cmd/attack.go**: `generateCombatDiagnostics()` - shows combat preview
- **services/singleton_gameview_presenter.go**: Orchestrates UI state updates

## Integration Strategy

### Phase 1: Parallel Systems (Current)
**Status**: ✅ Complete
- Formula-based system exists alongside table-based
- Both can coexist for validation
- No breaking changes to existing code

### Phase 2: Add Formula to Combat Diagnostics
**Goal**: Show formula-based predictions in CLI without changing game logic

**Changes**:
1. **cmd/cli/cmd/attack.go**:
   ```go
   func generateCombatDiagnostics(...) {
       // Existing table-based diagnostics

       // NEW: Add formula-based section
       sb.WriteString("\n[FORMULA-BASED PREDICTION]\n")

       // Calculate wound bonus
       woundBonus := rulesEngine.CalculateWoundBonus(defender, attackerCoord)

       // Create combat context
       ctx := &services.CombatContext{
           Attacker: attacker,
           AttackerTile: attackerTile,
           AttackerHealth: attackerHealth,
           Defender: defender,
           DefenderTile: defenderTile,
           DefenderHealth: defenderHealth,
           WoundBonus: woundBonus,
       }

       // Calculate hit probability
       p, _ := rulesEngine.CalculateHitProbability(ctx)
       sb.WriteString(fmt.Sprintf("Hit Probability (p): %.2f\n", p))

       // Show formula breakdown
       sb.WriteString(fmt.Sprintf("  A (base attack): %d\n", baseAttack))
       sb.WriteString(fmt.Sprintf("  Ta (terrain attack): %d\n", attackBonus))
       sb.WriteString(fmt.Sprintf("  D (base defense): %d\n", defense))
       sb.WriteString(fmt.Sprintf("  Td (terrain defense): %d\n", defenseBonus))
       sb.WriteString(fmt.Sprintf("  B (wound bonus): %d\n", woundBonus))

       // Generate damage distribution
       dist, _ := rulesEngine.GenerateDamageDistribution(ctx, 10000)
       sb.WriteString(fmt.Sprintf("Expected Damage: %.1f HP\n", dist.ExpectedDamage))
       sb.WriteString(fmt.Sprintf("Damage Range: %.0f-%.0f HP\n", dist.MinDamage, dist.MaxDamage))

       // Compare with table-based
       tableDist, _ := rulesEngine.GetCombatPrediction(attacker.UnitType, defender.UnitType)
       if tableDist != nil {
           diff := dist.ExpectedDamage - tableDist.ExpectedDamage
           sb.WriteString(fmt.Sprintf("\nFormula vs Table Difference: %.1f HP\n", diff))
       }
   }
   ```

**Benefits**:
- See formula calculations without changing game behavior
- Validate formula against existing table data
- Useful debugging information

### Phase 3: Integrate Formula into Combat Execution
**Goal**: Use formula for actual damage calculation in `ProcessAttackUnit()`

**Changes**:
1. **services/moves.go** - Update `ProcessAttackUnit()`:
   ```go
   func (m *MoveProcessor) ProcessAttackUnit(g *Game, ...) {
       // ... existing validation ...

       // NEW: Calculate wound bonus from attack history
       attackerCoord := services.CoordFromInt32(action.AttackerQ, action.AttackerR)
       woundBonus := g.rulesEngine.CalculateWoundBonus(defender, attackerCoord)

       // NEW: Create combat context for attacker -> defender
       attackerCtx := &services.CombatContext{
           Attacker: attacker,
           AttackerTile: g.World.TileAt(attackerCoord),
           AttackerHealth: attacker.AvailableHealth,
           Defender: defender,
           DefenderTile: g.World.TileAt(defenderCoord),
           DefenderHealth: defender.AvailableHealth,
           WoundBonus: woundBonus,
       }

       // NEW: Use formula-based damage calculation
       defenderDamage, err := g.rulesEngine.SimulateCombatDamage(attackerCtx, g.rng)
       if err != nil {
           return nil, fmt.Errorf("failed to calculate combat damage: %w", err)
       }

       // Counter-attack (no wound bonus for counter)
       if canCounter {
           counterCtx := &services.CombatContext{
               Attacker: defender,
               AttackerTile: g.World.TileAt(defenderCoord),
               AttackerHealth: defender.AvailableHealth,
               Defender: attacker,
               DefenderTile: g.World.TileAt(attackerCoord),
               DefenderHealth: attacker.AvailableHealth,
               WoundBonus: 0, // No wound bonus for counter-attacks
           }
           attackerDamage, err = g.rulesEngine.SimulateCombatDamage(counterCtx, g.rng)
       }

       // ... apply damage ...

       // NEW: Record attack in defender's history for wound bonus
       defender.AttackHistory = append(defender.AttackHistory, &v1.AttackRecord{
           Q: action.AttackerQ,
           R: action.AttackerR,
           IsRanged: services.CubeDistance(attackerCoord, defenderCoord) >= 2,
           TurnNumber: g.CurrentTurnCounter,
       })

       // ... rest of damage application and world changes ...
   }
   ```

2. **services/game.go** - Update `TopUpUnitIfNeeded()`:
   ```go
   func (g *Game) TopUpUnitIfNeeded(unit *v1.Unit) error {
       if unit.LastToppedupTurn < g.CurrentTurnCounter {
           // Reset movement and health
           // ...

           // NEW: Clear attack history for new turn
           unit.AttackHistory = nil
           unit.AttacksReceivedThisTurn = 0

           unit.LastToppedupTurn = g.CurrentTurnCounter
       }
       return nil
   }
   ```

**Benefits**:
- Uses actual attack formula with terrain bonuses
- Implements wound bonus accumulation
- More accurate combat than pre-calculated tables

### Phase 4: Implement Splash Damage
**Goal**: Add splash damage to adjacent units

**Changes**:
1. **services/combat_formula.go** - Add `ApplySplashDamage()`:
   ```go
   func (re *RulesEngine) ApplySplashDamage(
       world *World,
       attacker *v1.Unit,
       attackerTile *v1.Tile,
       defender *v1.Unit,
       rng *rand.Rand) ([]*SplashDamageResult, error) {

       attackerDef, _ := re.GetUnitData(attacker.UnitType)
       if attackerDef.SplashDamage == 0 {
           return nil, nil // No splash damage
       }

       defenderCoord := UnitGetCoord(defender)
       adjacentCoords := defenderCoord.Neighbors()

       results := []*SplashDamageResult{}

       for _, coord := range adjacentCoords {
           target := world.UnitAt(coord)
           if target == nil {
               continue
           }

           // Air units are unaffected by splash
           targetDef, _ := re.GetUnitData(target.UnitType)
           if targetDef.UnitTerrain == "Air" {
               continue
           }

           // Calculate splash (no wound bonus)
           ctx := &CombatContext{
               Attacker: attacker,
               AttackerTile: attackerTile,
               AttackerHealth: attacker.AvailableHealth,
               Defender: target,
               DefenderTile: world.TileAt(coord),
               DefenderHealth: target.AvailableHealth,
               WoundBonus: 0, // No wound bonus for splash
           }

           // Only deal splash if attack value > 4
           p, _ := re.CalculateHitProbability(ctx)
           attackValue := ... // Calculate from formula components
           if attackValue > 4 {
               damage, _ := re.SimulateCombatDamage(ctx, rng)
               results = append(results, &SplashDamageResult{
                   Target: target,
                   Damage: damage,
               })
           }
       }

       return results, nil
   }
   ```

2. **services/moves.go** - Call in `ProcessAttackUnit()`:
   ```go
   // After main attack damage

   // NEW: Apply splash damage
   splashResults, _ := g.rulesEngine.ApplySplashDamage(
       g.World, attacker, attackerTile, defender, g.rng)

   for _, splash := range splashResults {
       splash.Target.AvailableHealth -= splash.Damage
       // Record splash damage in world changes
       // ...
   }
   ```

**Benefits**:
- Complete attack formula implementation
- Handles friendly fire from splash
- Respects air unit immunity

## Presenter Integration

### Current Presenter Architecture
```
CLI → Presenter.SceneClicked()
        ↓
     ProcessSceneClicked() - orchestrates turn/move logic
        ↓
     GetCurrentState() - returns UI state with options
        ↓
     GetOptionsAt() - calculates available moves/attacks
```

### How Presenter Accesses Combat Info

1. **For Attack Options** (`GetOptionsAt()`):
   - Already uses `RulesEngine.GetAttackOptions()`
   - No changes needed - formula integration happens in `ProcessAttackUnit()`

2. **For Combat Preview** (UI tooltips):
   ```go
   // In GetOptionsAt() or new GetCombatPreview() method:

   func (p *SingletonGameViewPresenterImpl) GetCombatPreview(
       attackerCoord, defenderCoord AxialCoord) *CombatPreview {

       attacker := p.World.UnitAt(attackerCoord)
       defender := p.World.UnitAt(defenderCoord)

       // Calculate wound bonus
       woundBonus := p.RulesEngine.CalculateWoundBonus(defender, attackerCoord)

       // Create context
       ctx := &services.CombatContext{
           Attacker: attacker,
           AttackerTile: p.World.TileAt(attackerCoord),
           AttackerHealth: attacker.AvailableHealth,
           Defender: defender,
           DefenderTile: p.World.TileAt(defenderCoord),
           DefenderHealth: defender.AvailableHealth,
           WoundBonus: woundBonus,
       }

       // Generate distribution
       dist, _ := p.RulesEngine.GenerateDamageDistribution(ctx, 10000)

       return &CombatPreview{
           ExpectedDamage: dist.ExpectedDamage,
           MinDamage: dist.MinDamage,
           MaxDamage: dist.MaxDamage,
           HitProbability: p, // From CalculateHitProbability
           WoundBonus: woundBonus,
       }
   }
   ```

3. **CLI Access**:
   - CLI directly calls `generateCombatDiagnostics()` before executing attack
   - Already has access to `RulesEngine` via presenter
   - Just add formula calculations alongside existing diagnostics

## Migration Path

### Step 1: Add Formula Diagnostics to CLI ⏳
- Show formula calculations in `ww attack --verbose`
- Compare with table-based predictions
- No game logic changes

### Step 2: Add Integration Tests ⏳
- Test formula vs table for common scenarios
- Validate wound bonus calculations
- Test splash damage mechanics

### Step 3: Switch ProcessAttackUnit to Formula ⏳
- Replace `CalculateCombatDamage()` with `SimulateCombatDamage()`
- Add wound bonus tracking
- Keep old function for fallback

### Step 4: Add Splash Damage ⏳
- Implement splash damage calculation
- Add to ProcessAttackUnit flow
- Test with artillery units

### Step 5: Deprecate Table-Based System ⏳
- Remove pre-calculated damage distributions from JSON
- Remove `combat.go` (table-based system)
- Update documentation

## Testing Strategy

1. **Unit Tests**:
   - ✅ `TestCalculateHitProbability()` - formula correctness
   - ✅ `TestSimulateCombatDamage()` - dice rolling
   - ✅ `TestCalculateWoundBonus()` - wound bonus logic
   - ⏳ Test splash damage calculation
   - ⏳ Test attack history tracking

2. **Integration Tests**:
   - ⏳ Compare formula vs table for all unit matchups
   - ⏳ Test full combat flow with wound bonus
   - ⏳ Test turn transitions clear attack history

3. **CLI Manual Tests**:
   - ⏳ Attack same unit multiple times, verify wound bonus
   - ⏳ Attack from opposite sides, verify +3 bonus
   - ⏳ Artillery splash damage
   - ⏳ Air units ignore splash

## Open Questions

1. **Caching**: Should we cache damage distributions for common scenarios?
   - Pro: Faster UI tooltips
   - Con: More memory, complexity
   - **Decision**: Start without caching, add if needed

2. **Validation**: How to validate formula against table data?
   - Run comparison script: for each unit pair, compare formula vs table
   - Log differences, investigate outliers
   - **Action**: Create validation script in `cmd/validate-formula/`

3. **Backward Compatibility**: Should we keep table data?
   - Yes for Phase 2-3 (validation)
   - Remove in Phase 5 once formula is proven

4. **Splash Damage Edge Cases**:
   - Does splash damage count toward wound bonus? **No** (per ATTACK.md)
   - Can you kill your own units with splash? **Yes** (per ATTACK.md)
   - How to show splash targets in UI? **TBD** (future work)

## Success Criteria

✅ Formula calculations match expected values from ATTACK.md
✅ Unit tests pass for all formula components
⏳ CLI shows formula-based diagnostics
⏳ Formula damage matches table damage for no-wound-bonus scenarios
⏳ Wound bonus accumulates correctly across multiple attacks
⏳ Splash damage works for artillery units
⏳ Integration tests pass
⏳ No regressions in existing game behavior

## Timeline Estimate

- Phase 2 (CLI Diagnostics): 2-3 hours
- Phase 3 (Combat Integration): 4-6 hours
- Phase 4 (Splash Damage): 3-4 hours
- Phase 5 (Deprecation): 1-2 hours
- Testing throughout: 4-6 hours

**Total**: ~15-20 hours of development
