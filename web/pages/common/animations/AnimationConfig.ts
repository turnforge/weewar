/**
 * Animation timing configuration.
 *
 * All durations are in milliseconds.
 * Set any duration to 0 for instant (no animation).
 * Set all to 0 for instant mode gameplay.
 */
export const AnimationConfig = {
  /** Duration per hex tile when moving units (ms) */
  MOVE_DURATION_PER_HEX: 200,

  /** Pause duration at each tile during movement (ms) */
  MOVE_PAUSE_PER_HEX: 100,

  /** Duration of attacker flash when initiating attack (ms) */
  ATTACK_FLASH_DURATION: 150,

  /** Duration of projectile flight from attacker to target (ms) */
  PROJECTILE_DURATION: 300,

  /** Duration of explosion particle effect (ms) */
  EXPLOSION_DURATION: 300,

  /** Duration of healing bubble particle effect (ms) */
  HEAL_DURATION: 400,

  /** Duration of capture/occupation animation (ms) */
  CAPTURE_DURATION: 500,

  /** Duration of unit fade-out on death (ms) */
  FADE_OUT_DURATION: 250,

  /** Duration of flash effect (damage, events) (ms) */
  FLASH_DURATION: 200,

  /** Duration of appear/fade-in effect (ms) */
  APPEAR_DURATION: 200,

  /** Duration of one complete flag wave cycle (ms) */
  FLAG_WAVE_DURATION: 600,
};

/**
 * Animation visual configuration.
 */
export const AnimationVisualConfig = {
  /** Projectile arc height multiplier (0 = straight line, 1 = high arc) */
  PROJECTILE_ARC_HEIGHT: 0.5,

  /** Base particle count for explosions */
  EXPLOSION_PARTICLE_COUNT: 20,

  /** Particle count multiplier per damage point */
  EXPLOSION_PARTICLES_PER_DAMAGE: 2,

  /** Maximum explosion particle count */
  EXPLOSION_PARTICLE_MAX: 50,

  /** Heal bubble particle count */
  HEAL_PARTICLE_COUNT: 15,

  /** Flash tint color for damage */
  FLASH_DAMAGE_COLOR: 0xff0000,

  /** Flash tint color for attacks */
  FLASH_ATTACK_COLOR: 0xff6600,

  /** Explosion particle colors [min, max] for color range */
  EXPLOSION_COLOR_MIN: 0xff3300,
  EXPLOSION_COLOR_MAX: 0xffff00,

  /** Heal bubble colors [min, max] for color range */
  HEAL_COLOR_MIN: 0x00ff00,
  HEAL_COLOR_MAX: 0x00ffff,
} as const;
