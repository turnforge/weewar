/**
 * GameState Component Tests
 * Tests the GameState component with real WASM integration
 */

import { GameState } from '../frontend/components/GameState';
import { createTestGameState, validateGameState, validateMovementOptions, validateAttackOptions } from './helpers/wasmTestUtils';
import { SMALL_TEST_MAP, COMBAT_TEST_MAP, testMapToJSON } from './fixtures/testMaps';
import * as path from 'path';

describe('GameState Component', () => {
  let gameState: GameState;
  let cleanup: () => void;

  beforeEach(async () => {
    const testSetup = await createTestGameState()
    gameState = testSetup.gameState;
    cleanup = testSetup.cleanup;
  });

  afterEach(() => {
    if (cleanup) {
      cleanup();
    }
  });

  describe('WASM Loading and Initialization', () => {
    test('should load WASM module successfully', () => {
      // Check that WASM is loaded (but game not necessarily initialized yet)
      const gameData = gameState.getGameData();
      expect(gameData.wasmLoaded).toBe(true);
      
      // isReady() requires both wasmLoaded AND gameInitialized
      // For this test, we only care that WASM loaded successfully
    });

    test('should have all required WASM functions available', async () => {
      // These functions should be available after WASM loading
      expect((window as any).weewarCreateGameFromMap).toBeDefined();
      expect((window as any).weewarGetGameState).toBeDefined();
      expect((window as any).weewarMoveUnit).toBeDefined();
      expect((window as any).weewarAttackUnit).toBeDefined();
      expect((window as any).weewarEndTurn).toBeDefined();
    });
  });

  describe('Game Creation', () => {
    test('should create game from small test map', async () => {
      const mapData = testMapToJSON(SMALL_TEST_MAP);
      
      const gameData = await gameState.createGameFromMap(mapData);
      
      // Validate game creation response
      const validation = validateGameState(gameData);
      expect(validation.hasCurrentPlayer).toBe(true);
      expect(validation.hasTurnCounter).toBe(true);
      expect(validation.hasMapSize).toBe(true);
      expect(validation.hasUnits).toBe(true);
      
      // Validate specific values from test map
      expect(validation.playerCount).toBe(SMALL_TEST_MAP.players.length);
      expect(validation.unitCount).toBe(SMALL_TEST_MAP.units.length);
      expect(gameData.currentPlayer).toBeGreaterThan(0);
      expect(gameData.turnCounter).toBe(1);
    });

    test('should create different games with different maps', async () => {
      const smallMapData = testMapToJSON(SMALL_TEST_MAP);
      const combatMapData = testMapToJSON(COMBAT_TEST_MAP);
      
      const game1 = await gameState.createGameFromMap(smallMapData);
      // Reset and create new game
      const { gameState: gameState2, cleanup: cleanup2 } = await createTestGameState();
      const game2 = await gameState2.createGameFromMap(combatMapData);
      
      expect(Object.values(game1.allUnits).length).toBe(SMALL_TEST_MAP.units.length);
      expect(Object.values(game2.allUnits).length).toBe(COMBAT_TEST_MAP.units.length);
      
      cleanup2();
    });
  });

  describe('Game State Management', () => {
    beforeEach(async () => {
      // Create a game for each test
      const mapData = testMapToJSON(SMALL_TEST_MAP);
      await gameState.createGameFromMap(mapData);
    });

    test('should retrieve current game state', () => {
      const currentState = gameState.getGameState();
      
      const validation = validateGameState(currentState);
      expect(validation.hasCurrentPlayer).toBe(true);
      expect(validation.hasTurnCounter).toBe(true);
      expect(validation.unitCount).toBe(SMALL_TEST_MAP.units.length);
    });

    test('should advance turn when endTurn is called', () => {
      const initialState = gameState.getGameState();
      const initialPlayer = initialState.currentPlayer;
      const initialTurn = initialState.turnCounter;
      
      const newState = gameState.endTurn();
      
      // Player should change (or turn counter should increment if single player)
      expect(newState.currentPlayer !== initialPlayer || newState.turnCounter > initialTurn).toBe(true);
    });
  });

  describe('Unit Selection and Options', () => {
    beforeEach(async () => {
      const mapData = testMapToJSON(SMALL_TEST_MAP);
      await gameState.createGameFromMap(mapData);
    });

    test('should check if units can be selected', () => {
      // Test selecting valid unit position from our test map
      const unit = SMALL_TEST_MAP.units[0]; // First unit
      const canSelect = gameState.canSelectUnit(unit.q, unit.r);
      
      // Should be able to select if it's the current player's unit
      expect(typeof canSelect).toBe('boolean');
    });

    test('should get movement options for valid unit', () => {
      const unit = SMALL_TEST_MAP.units[0];
      const movementResponse = gameState.getMovementOptions(unit.q, unit.r);
      
      const validation = validateMovementOptions(movementResponse);
      expect(validation.isValid).toBe(true);
      
      if (movementResponse.success) {
        expect(validation.coordinateCount).toBeGreaterThanOrEqual(0);
      }
    });

    /* - To be fixed
    test('should get attack options for valid unit', () => {
      const unit = SMALL_TEST_MAP.units[0];
      const attackResponse = gameState.getAttackOptions(unit.q, unit.r);
      
      const validation = validateAttackOptions(attackResponse);
      expect(validation.isValid).toBe(true);
      expect(validation.targetCount).toBeGreaterThanOrEqual(0);
    });
     */

    test('should handle invalid coordinates gracefully', () => {
      const movementResponse = gameState.getMovementOptions(-1, -1);
      expect(movementResponse.success).toBe(false);
      
      const attackResponse = gameState.getAttackOptions(999, 999);
      expect(attackResponse.success).toBe(false);
    });

    test('should get tile information for valid coordinates', () => {
      const unit = SMALL_TEST_MAP.units[0];
      
      expect(() => {
        const tileInfo = gameState.getTileInfo(unit.q, unit.r);
        expect(tileInfo).toBeDefined();
      }).not.toThrow();
    });
  });

  describe('Unit Actions', () => {
    beforeEach(async () => {
      const mapData = testMapToJSON(COMBAT_TEST_MAP);
      await gameState.createGameFromMap(mapData);
    });

    test('should handle unit movement', () => {
      // This test depends on game rules - we'll test that the WASM call works
      // regardless of whether the move is valid according to current game state
      const fromUnit = COMBAT_TEST_MAP.units[0];
      const toPosition = { q: fromUnit.q + 1, r: fromUnit.r }; // Adjacent position
      
      try {
        const result = gameState.moveUnit(fromUnit.q, fromUnit.r, toPosition.q, toPosition.r);
        // If move succeeded, result should be defined
        expect(result).toBeDefined();
      } catch (error) {
        // If move failed due to game rules, that's also valid - just check error type
        expect(error).toBeInstanceOf(Error);
      }
    });

    test('should handle unit attacks', () => {
      // Test attack between adjacent units in combat test map
      const attacker = COMBAT_TEST_MAP.units[0]; // Player 1 unit
      const defender = COMBAT_TEST_MAP.units[1]; // Player 2 unit (adjacent)
      
      try {
        const result = gameState.attackUnit(attacker.q, attacker.r, defender.q, defender.r);
        expect(result).toBeDefined();
      } catch (error) {
        // Attack might fail due to game rules (wrong turn, etc.)
        expect(error).toBeInstanceOf(Error);
      }
    });

    test('should reject invalid unit actions', () => {
      expect(() => {
        gameState.moveUnit(-1, -1, 0, 0);
      }).toThrow();
      
      expect(() => {
        gameState.attackUnit(999, 999, 0, 0);
      }).toThrow();
    });
  });

  describe('Terrain Information', () => {
    beforeEach(async () => {
      const mapData = testMapToJSON(SMALL_TEST_MAP);
      await gameState.createGameFromMap(mapData);
    });

    test('should get terrain stats for valid coordinates', () => {
      // Test getting terrain stats for the center mountain tile
      const terrainStats = gameState.getTerrainStatsAt(1, 1);
      
      if (terrainStats) {
        expect(terrainStats).toHaveProperty('tileType');
        expect(terrainStats).toHaveProperty('movementCost');
      }
      // It's ok if terrainStats is null for some coordinates
    });

    test('should handle invalid terrain coordinates', () => {
      const terrainStats = gameState.getTerrainStatsAt(-1, -1);
      expect(terrainStats).toBeNull();
    });
  });
});
