# Animation Framework

## Purpose

Provides a pluggable animation system for game actions in GameViewerPage/PhaserGameScene. Animations are presenter-driven, keeping the scene as a dumb renderer that executes visual commands.

## Architecture

### Design Principles

1. **Presenter-driven**: Presenter interprets game logic and calls scene animation methods
2. **Scene is dumb**: Scene has no game logic, just executes rendering instructions
3. **Promise-based**: All animations return Promises for easy sequencing
4. **Configurable timing**: Single config file controls all animation speeds
5. **Smart batching**: Effects can play simultaneously (e.g., splash damage explosions)

### API Design

Scene provides simple methods that presenter calls:

```typescript
// Unit lifecycle
setUnit(unit, options?: {flash?, appear?}): Promise<void>
moveUnit(unit, path: {q,r}[]): Promise<void>
removeUnit(q, r, options?: {animate?}): Promise<void>

// Visual effects
showAttackEffect(from, to, damage, splashTargets?): Promise<void>
showHealEffect(q, r, amount): Promise<void>
showCaptureEffect(q, r): Promise<void>
showExplosion(q, r, intensity): Promise<void>
```

## Configuration

### AnimationConfig.ts

Central configuration for all animation timings:

- `MOVE_DURATION_PER_HEX`: Time to move one hex (200ms)
- `MOVE_PAUSE_PER_HEX`: Pause at each tile during movement (100ms)
- `ATTACK_FLASH_DURATION`: Attacker flash duration (150ms)
- `PROJECTILE_DURATION`: Projectile flight time (300ms)
- `EXPLOSION_DURATION`: Explosion effect duration (300ms)
- `HEAL_DURATION`: Healing bubbles duration (400ms)
- `CAPTURE_DURATION`: Capture effect duration (500ms)
- `FADE_OUT_DURATION`: Unit death fade (250ms)
- `FLASH_DURATION`: Damage flash (200ms)
- `APPEAR_DURATION`: Unit spawn fade-in (200ms)

**Instant mode**: Set any duration to 0 to disable that animation.

### AnimationVisualConfig

Visual parameters for effects:

- Projectile arc height
- Particle counts and scaling
- Color ranges for explosions and healing
- Flash colors

## Effect Classes

Located in `effects/` directory. Each effect is self-contained and returns a Promise.

### ProjectileEffect.ts

Ballistic projectile that arcs from source to target.

**Features**:
- Parabolic arc trajectory
- Configurable arc height
- Smooth animation using Phaser tweens

**Usage**:
```typescript
const projectile = new ProjectileEffect(scene, fromX, fromY, toX, toY);
await projectile.play();
```

### ExplosionEffect.ts

Particle burst effect for impacts and damage.

**Features**:
- Scales particle count by damage intensity
- Color gradient (red/orange/yellow)
- Supports simultaneous multiple explosions (splash damage)

**Usage**:
```typescript
// Single explosion
const explosion = new ExplosionEffect(scene, x, y, intensity);
await explosion.play();

// Multiple simultaneous (splash damage)
await ExplosionEffect.playMultiple(scene, [
  {x: pos1.x, y: pos1.y, intensity: 10},
  {x: pos2.x, y: pos2.y, intensity: 5}
]);
```

### HealBubblesEffect.ts

Rising bubble particles for healing effects.

**Features**:
- Green/cyan colored particles
- Float upward with gravity
- Configurable particle count

**Usage**:
```typescript
const heal = new HealBubblesEffect(scene, x, y, healAmount);
await heal.play();
```

### CaptureEffect.ts

Pulse effect for tile/building capture.

**Features**:
- Expanding circle with fade
- Orange color theme
- Simple and clear visual feedback

**Usage**:
```typescript
const capture = new CaptureEffect(scene, x, y);
await capture.play();
```

## Integration Points

### common/PhaserWorldScene Enhancements

**Modified methods**:
- `setUnit()`: Now async with flash/appear options
- `removeUnit()`: Now async with animate option
- `create()`: Creates particle texture for effects

**New methods**:
- `moveUnit()`: Animates unit along hex path with pauses at each tile
- `showAttackEffect()`: Full attack sequence (flash → projectile → explosion)
- `showHealEffect()`: Healing bubble animation
- `showCaptureEffect()`: Capture pulse animation
- `showExplosion()`: Standalone explosion utility
- `createParticleTexture()`: Generates white circle texture for particles

### Movement Animation Details

The `moveUnit()` method implements pathfinding-aware movement:

1. **Full path support**: Accepts array of hex coordinates representing the complete pathfinding route
2. **Segment-by-segment animation**: Moves through each waypoint sequentially
3. **Pause at tiles**: Brief pause at each tile for board game feel (configurable via `MOVE_PAUSE_PER_HEX`)
4. **Smooth transitions**: Uses cubic easing between tiles
5. **Fallback**: If path has only 2 points, animates direct movement

**Path extraction**: Presenter extracts the full path from `MoveUnitAction.reconstructed_path` which contains all waypoints calculated by the pathfinding algorithm.

### Presenter Integration (TODO)

Presenter should interpret `WorldChange[]` and call appropriate scene methods:

```typescript
// Movement
if (change.unitMoved) {
  await scene.moveUnit(updatedUnit, derivedPath);
}

// Attack with damage
if (change.unitDamaged) {
  await scene.showAttackEffect(attackerPos, defenderPos, damage);
  await scene.setUnit(updatedUnit, { flash: true });
}

// Death
if (change.unitKilled) {
  await scene.removeUnit(unit.q, unit.r, { animate: true });
}

// Healing
if (change.unitHealed) {
  await scene.showHealEffect(unit.q, unit.r, healAmount);
  await scene.setUnit(updatedUnit);
}
```

## Technical Implementation

### Phaser Techniques Used

1. **Tweens**: Position/alpha/tint interpolation for smooth transitions
2. **Particle Emitters**: Burst effects for explosions and healing
3. **Graphics**: Simple shapes for projectiles and particles
4. **Timeline chaining**: Sequential segment animation for movement paths
5. **Promises**: Async/await for sequencing and batching

### Particle System

Uses Phaser's built-in particle emitter with a simple white circle texture (`particle`). Texture is generated at runtime in `create()`:

```typescript
graphics.fillStyle(0xffffff);
graphics.fillCircle(8, 8, 8);
graphics.generateTexture('particle', 16, 16);
```

Particles are tinted via emitter config to create colored effects.

### Coordinate Conversion

All effects use `hexToPixel(q, r)` from common/hexUtils.ts to convert hex coordinates to world pixel positions for rendering.

### Depth Layering

- Projectiles: Depth 15 (above units)
- Particles: Depth 20 (above everything)
- Effects layer between units (10) and UI (15+)

## Future Enhancements

### Potential Improvements

1. **Sprite sheet animations**: Replace simple tweens with frame animations for unit movement
2. **Sound integration**: Add sound effect hooks to each animation
3. **Camera tracking**: Smooth camera follow during long movements
4. **Animation speed settings**: User preference for animation speeds
5. **Skip button**: Allow player to skip/fast-forward animations
6. **Custom particles**: Create textured particles (smoke, sparks, debris) instead of circles
7. **Trail effects**: Add movement trails for projectiles
8. **Screen shake**: Add camera shake for explosions
9. **Status effect animations**: Animations for buffs/debuffs/status conditions
10. **Terrain interaction**: Dust clouds, water splashes based on terrain type

### Extension Pattern

To add new animations:

1. Create effect class in `effects/` extending pattern:
   ```typescript
   export class NewEffect {
     constructor(scene, x, y, params) { ... }
     public play(): Promise<void> { ... }
   }
   ```

2. Add timing constants to `common/animations/AnimationConfig.ts`

3. Add scene method in `common/PhaserWorldScene.ts`:
   ```typescript
   public showNewEffect(q, r, params): Promise<void> {
     const pos = hexToPixel(q, r);
     const effect = new NewEffect(this, pos.x, pos.y, params);
     return effect.play();
   }
   ```

4. Call from presenter when appropriate `WorldChange` occurs

## File Structure

```
web/pages/common/animations/
├── SUMMARY.md                      # This file
├── AnimationConfig.ts              # Timing and visual configuration
└── effects/
    ├── ProjectileEffect.ts         # Ballistic projectile
    ├── ExplosionEffect.ts          # Particle burst
    ├── HealBubblesEffect.ts        # Healing bubbles
    └── CaptureEffect.ts            # Capture pulse
```

## Performance Considerations

- Particle emitters are destroyed after animation completes
- Projectile graphics are cleaned up on impact
- Tweens automatically clean up when complete
- All effects respect instant mode (duration = 0) for zero overhead
- Simultaneous effects share render batches where possible

## Testing

To test animations without presenter integration:

```typescript
// In browser console with scene reference
await scene.moveUnit(unit, [{q:0,r:0}, {q:1,r:0}, {q:2,r:0}]);
await scene.showAttackEffect({q:0,r:0}, {q:1,r:0}, 10);
await scene.showHealEffect(0, 0, 5);
await scene.showCaptureEffect(0, 0);
```
