/**
 * WASM Testing Utilities
 * Helper functions for testing with real WASM binary
 */

import { GameState } from '../../frontend/components/GameState';
import { EventBus } from '../../frontend/components/EventBus';

/**
 * Create a minimal DOM environment for testing GameState
 */
export function createMinimalDOM(): HTMLElement {
  // Create a minimal container for GameState
  const container = document.createElement('div');
  container.id = 'test-game-state-container';
  container.style.display = 'none';
  document.body.appendChild(container);
  return container;
}

/**
 * Clean up DOM after testing
 */
export function cleanupDOM(container: HTMLElement): void {
  if (container.parentNode) {
    container.parentNode.removeChild(container);
  }
}

/**
 * Pre-load WASM for testing using the same approach as wasmLoading.test.ts
 */
async function preloadWASM(): Promise<void> {
  // Skip if already loaded
  if ((window as any).weewarCreateGameFromMap) {
    return;
  }

  const fs = require('fs');
  const path = require('path');
  
  const wasmPath = path.join(__dirname, '../../static/wasm/weewar-cli.wasm');
  const wasmExecPath = path.join(__dirname, '../../static/wasm/wasm_exec.js');
  
  // Load Go runtime (same as wasmLoading.test.ts)
  const wasmExecCode = fs.readFileSync(wasmExecPath, 'utf8');
  eval(wasmExecCode);
  
  // Make Go available on both global and window
  (window as any).Go = (global as any).Go;
  
  // Load WASM binary (same as wasmLoading.test.ts)
  const wasmBuffer = fs.readFileSync(wasmPath);
  const arrayBuffer = wasmBuffer.buffer.slice(
    wasmBuffer.byteOffset, 
    wasmBuffer.byteOffset + wasmBuffer.byteLength
  );
  
  const go = new (global as any).Go();
  
  const wasmModule = await WebAssembly.instantiate(arrayBuffer, go.importObject);
  
  // Run WASM (same as wasmLoading.test.ts)  
  go.run(wasmModule.instance);
  
  // Wait for WASM functions to be available on global
  await new Promise(resolve => setTimeout(resolve, 1000));
  
  // Copy all WASM functions from global to window for GameState compatibility
  const globalKeys = Object.keys(global).filter(k => k.startsWith('weewar'));
  globalKeys.forEach(funcName => {
    (window as any)[funcName] = (global as any)[funcName];
  });
  
  console.log('WASM pre-loaded for GameState testing. Available functions:', globalKeys);
}

/**
 * Create GameState instance for testing
 */
export async function createTestGameState(): Promise<{ gameState: GameState; eventBus: EventBus; cleanup: () => void }> {
  // Pre-load WASM before creating GameState
  await preloadWASM();
  
  const eventBus = new EventBus(false); // Disable debug mode for cleaner test output
  const container = createMinimalDOM();
  
  const gameState = new GameState(container, eventBus, false);
  
  // Since WASM is pre-loaded, GameState should recognize it immediately
  // But we still call waitUntilReady() to ensure the component is properly initialized
    await gameState.waitUntilReady();
  try {
  } catch (error) {
    console.error('GameState WASM loading failed, but WASM should be pre-loaded. Error:', error);
    console.log('Available WASM functions on window:', Object.keys(window).filter(k => k.startsWith('weewar')));
    throw error;
  }
  
  const cleanup = () => {
    gameState.destroy();
    cleanupDOM(container);
  };
  
  return { gameState, eventBus, cleanup };
}

/**
 * Wait for event with timeout
 */
export function waitForEvent(eventBus: EventBus, eventName: string, timeoutMs: number = 5000): Promise<any> {
  return new Promise((resolve, reject) => {
    const handler = (data: any) => {
      clearTimeout(timeout);
      resolve(data);
    };
    
    const timeout = setTimeout(() => {
      eventBus.unsubscribe(eventName, 'test-waiter', handler);
      reject(new Error(`Timeout waiting for event: ${eventName}`));
    }, timeoutMs);
    
    eventBus.subscribe(eventName, handler, 'test-waiter');
  });
}

/**
 * Verify game state structure
 */
export interface GameStateValidation {
  hasCurrentPlayer: boolean;
  hasTurnCounter: boolean;
  hasMapSize: boolean;
  hasPlayers: boolean;
  hasUnits: boolean;
  playerCount: number;
  unitCount: number;
}

export function validateGameState(gameData: any): GameStateValidation {
  return {
    hasCurrentPlayer: typeof gameData.currentPlayer === 'number',
    hasTurnCounter: typeof gameData.turnCounter === 'number',
    hasMapSize: gameData.mapSize && typeof gameData.mapSize.rows === 'number',
    hasPlayers: Array.isArray(gameData.players),
    hasUnits: gameData.unitCount > 0,
    playerCount: gameData.players?.length || 0,
    unitCount: gameData.unitCount,
  };
}

/**
 * Test coordinate validation helper
 */
export function isValidCoordinate(q: number, r: number): boolean {
  return Number.isInteger(q) && Number.isInteger(r) && q >= 0 && r >= 0;
}

/**
 * Verify movement options structure
 */
export function validateMovementOptions(response: any): {
  isValid: boolean;
  hasData: boolean;
  coordinateCount: number;
  errors: string[];
} {
  const errors: string[] = [];
  
  if (!response) {
    errors.push('Response is null or undefined');
    return { isValid: false, hasData: false, coordinateCount: 0, errors };
  }
  
  if (typeof response.success !== 'boolean') {
    errors.push('Response missing boolean success field');
  }
  
  if (!response.success) {
    return { isValid: true, hasData: false, coordinateCount: 0, errors };
  }
  
  if (!Array.isArray(response.data)) {
    errors.push('Response data is not an array');
    return { isValid: false, hasData: false, coordinateCount: 0, errors };
  }
  
  // Validate coordinate structure
  for (let i = 0; i < response.data.length; i++) {
    const coord = response.data[i];
    if (!coord.coord || typeof coord.coord.q !== 'number' || typeof coord.coord.r !== 'number') {
      errors.push(`Invalid coordinate structure at index ${i}`);
    }
  }
  
  return {
    isValid: errors.length === 0,
    hasData: response.data.length > 0,
    coordinateCount: response.data.length,
    errors
  };
}

/**
 * Verify attack options structure
 */
export function validateAttackOptions(response: any): {
  isValid: boolean;
  hasTargets: boolean;
  targetCount: number;
  errors: string[];
} {
  const errors: string[] = [];
  
  if (!response) {
    errors.push('Response is null or undefined');
    return { isValid: false, hasTargets: false, targetCount: 0, errors };
  }
  
  if (typeof response.success !== 'boolean') {
    errors.push('Response missing boolean success field');
  }
  
  if (!response.success) {
    return { isValid: true, hasTargets: false, targetCount: 0, errors };
  }
  
  if (!Array.isArray(response.data)) {
    errors.push('Response data is not an array');
    return { isValid: false, hasTargets: false, targetCount: 0, errors };
  }
  
  // Validate coordinate structure
  for (let i = 0; i < response.data.length; i++) {
    const coord = response.data[i];
    if (!coord.coord || typeof coord.coord.q !== 'number' || typeof coord.coord.r !== 'number') {
      errors.push(`Invalid coordinate structure at index ${i}`);
    }
  }
  
  return {
    isValid: errors.length === 0,
    hasTargets: response.data.length > 0,
    targetCount: response.data.length,
    errors
  };
}
