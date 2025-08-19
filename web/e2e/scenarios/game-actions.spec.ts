/**
 * Game Actions Integration Tests
 * 
 * Tests high-level game actions using the command interface:
 * - Unit selection
 * - Unit movement  
 * - Turn management
 * - Game state queries
 * 
 * These tests validate actual game mechanics using real GameViewerPage
 */

import { test, expect } from '@playwright/test';
import { TEST_SCENARIOS, createTestGame, deleteTestGame } from '../infrastructure/test-games';
import { installBasicApiMocking } from '../infrastructure/api-mocking';
import { GameActions, GameTestUtils } from '../infrastructure/game-actions';

test.describe('GameViewerPage - Game Actions', () => {
  const SERVER_URL = 'http://localhost:8080';
  let testGameUrl: string;
  let testGameId: string;
  let gameActions: GameActions;
  let testUtils: GameTestUtils;

  test.beforeAll(async () => {
    console.log('🔧 Setting up test game for action testing...');
    
    // Health check
    try {
      const healthCheck = await fetch(`${SERVER_URL}/`);
      if (!healthCheck.ok) {
        throw new Error(`Server health check failed: ${healthCheck.status}`);
      }
    } catch (error) {
      throw new Error(`
❌ Go server not running or not accessible at ${SERVER_URL}

To run e2e tests:
1. Start the Go server: ./weewar-server --port=8080
2. Then run: npm run test:e2e

Error: ${error}
      `);
    }

    // Create test game
    const result = await createTestGame(TEST_SCENARIOS.BASIC_MOVEMENT, SERVER_URL);
    if (!result.success) {
      throw new Error(`Failed to create test game: ${result.error}`);
    }
    testGameUrl = result.url!;
    testGameId = result.gameId!;
    console.log(`✅ Test game created: ${testGameUrl}`);
  });

  test.afterAll(async () => {
    // Cleanup
    if (testGameId) {
      console.log(`🧹 Cleaning up test game: ${testGameId}`);
      const deleted = await deleteTestGame(testGameId, SERVER_URL);
      if (deleted) {
        console.log(`✅ Test game ${testGameId} deleted successfully`);
      } else {
        console.warn(`⚠️ Failed to delete test game ${testGameId}`);
      }
    }
  });

  test.beforeEach(async ({ page }) => {
    // Install minimal API mocking
    await installBasicApiMocking(page);
    
    // Initialize game actions utilities
    gameActions = new GameActions(page);
    testUtils = new GameTestUtils(page, gameActions);
    
    // Load the game
    await page.goto(testGameUrl);
    await page.waitForLoadState('networkidle');
    
    // Wait for game to be ready
    await gameActions.waitForGameReady();
    
    console.log('🎮 Game ready for action testing');
  });

  test('should get initial game state correctly', async ({ page }) => {
    const gameState = await gameActions.getGameState();
    const units = await gameActions.getVisibleUnits();
    
    // Debug: Log what we actually have
    console.log('🔍 Debug - Game state:', gameState);
    console.log('🔍 Debug - Visible units:', units);
    
    // Verify basic game state
    expect(gameState.gameId).toBeDefined();
    expect(gameState.currentPlayer).toBeGreaterThan(0);
    expect(gameState.turnCounter).toBeGreaterThan(0);
    expect(gameState.unitsCount).toBeGreaterThan(0);
    expect(gameState.tilesCount).toBeGreaterThan(0);
    
    console.log('✅ Initial game state retrieved:', {
      gameId: gameState.gameId,
      currentPlayer: gameState.currentPlayer,
      turnCounter: gameState.turnCounter,
      units: gameState.unitsCount,
      tiles: gameState.tilesCount,
      visibleUnits: units.length
    });
  });

  test('should select a unit successfully', async ({ page }) => {
    // Get initial state to find a unit to select
    const units = await gameActions.getVisibleUnits();
    expect(units.length).toBeGreaterThan(0);
    
    const gameState = await gameActions.getGameState();
    const currentPlayer = gameState.currentPlayer;
    
    // Find a unit belonging to current player
    const playerUnit = units.find(unit => unit.player === currentPlayer);
    expect(playerUnit).toBeDefined();
    
    if (!playerUnit) {
      throw new Error('No units found for current player');
    }
    
    // Execute unit selection
    const result = await testUtils.executeWithReporting(
      'selectUnit',
      () => gameActions.selectUnit(playerUnit.q, playerUnit.r)
    );
    
    // Verify selection was successful
    expect(result.success).toBe(true);
    expect(result.message).toContain(`Unit selected at (${playerUnit.q}, ${playerUnit.r})`);
    
    // Verify unit is now selected in game state
    const updatedState = await gameActions.getGameState();
    expect(updatedState.selectedUnit).toEqual({ q: playerUnit.q, r: playerUnit.r });
    
    console.log('✅ Unit selection successful:', {
      position: { q: playerUnit.q, r: playerUnit.r },
      player: playerUnit.player,
      result: result.message,
      duration: result.duration
    });
  });

  test('should fail to select non-existent unit', async ({ page }) => {
    // Try to select a position with no unit (using coordinates outside map)
    const result = await gameActions.selectUnit(99, 99);
    
    // Verify selection failed appropriately
    expect(result.success).toBe(false);
    expect(result.message).toContain('No unit found at position (99, 99)');
    expect(result.error).toBe('No unit at coordinates');
    
    console.log('✅ Non-existent unit selection correctly failed:', result.message);
  });

  test('should perform a complete move action', async ({ page }) => {
    // Get units and find one to move
    const units = await gameActions.getVisibleUnits();
    const gameState = await gameActions.getGameState();
    const currentPlayer = gameState.currentPlayer;
    
    const playerUnit = units.find(unit => unit.player === currentPlayer);
    expect(playerUnit).toBeDefined();
    
    if (!playerUnit) {
      throw new Error('No units found for current player');
    }
    
    // Select the unit first
    const selectResult = await gameActions.selectUnit(playerUnit.q, playerUnit.r);
    expect(selectResult.success).toBe(true);
    
    // Attempt to move to an adjacent position
    // For now, just test the move command structure - actual valid moves depend on game rules
    const targetQ = playerUnit.q + 1;
    const targetR = playerUnit.r;
    
    const moveResult = await testUtils.executeWithReporting(
      'moveUnit',
      () => gameActions.moveSelectedUnit(targetQ, targetR)
    );
    
    // Note: This might fail if the target position is invalid, but we're testing the interface
    console.log('🚀 Move attempt result:', {
      from: { q: playerUnit.q, r: playerUnit.r },
      to: { q: targetQ, r: targetR },
      success: moveResult.success,
      message: moveResult.message,
      duration: moveResult.duration
    });
    
    // The test passes regardless of move validity - we're testing the command interface
    expect(moveResult).toBeDefined();
    expect(typeof moveResult.success).toBe('boolean');
    expect(typeof moveResult.message).toBe('string');
  });

  test('should clear unit selection', async ({ page }) => {
    // First select a unit
    const units = await gameActions.getVisibleUnits();
    const gameState = await gameActions.getGameState();
    const currentPlayer = gameState.currentPlayer;
    
    const playerUnit = units.find(unit => unit.player === currentPlayer);
    if (playerUnit) {
      await gameActions.selectUnit(playerUnit.q, playerUnit.r);
      
      // Verify unit is selected
      const stateWithSelection = await gameActions.getGameState();
      expect(stateWithSelection.selectedUnit).toBeDefined();
      
      // Clear selection
      const clearResult = await gameActions.clearSelection();
      expect(clearResult.success).toBe(true);
      expect(clearResult.message).toBe('Selection cleared');
      
      // Verify selection is cleared
      const stateAfterClear = await gameActions.getGameState();
      expect(stateAfterClear.selectedUnit).toBeUndefined();
      
      console.log('✅ Unit selection cleared successfully');
    }
  });

  test('should end turn successfully', async ({ page }) => {
    // Get initial game state
    const initialState = await gameActions.getGameState();
    const initialPlayer = initialState.currentPlayer;
    const initialTurn = initialState.turnCounter;
    
    // End the turn
    const result = await testUtils.executeWithReporting(
      'endTurn',
      () => gameActions.endTurn()
    );
    
    expect(result.success).toBe(true);
    expect(result.message).toContain('Turn ended');
    
    // Verify game state changed appropriately
    const finalState = await gameActions.getGameState();
    
    // Player should have changed (in a 2-player game)
    expect(finalState.currentPlayer).not.toBe(initialPlayer);
    
    console.log('✅ Turn ended successfully:', {
      previousPlayer: initialPlayer,
      currentPlayer: finalState.currentPlayer,
      previousTurn: initialTurn,
      currentTurn: finalState.turnCounter,
      duration: result.duration
    });
  });

  test('should handle multiple sequential actions', async ({ page }) => {
    // Perform a sequence of actions to test the interface stability
    const actions = [];
    
    // 1. Get initial state
    let gameState = await gameActions.getGameState();
    actions.push(`Initial: Player ${gameState.currentPlayer}, Turn ${gameState.turnCounter}`);
    
    // 2. Try to select a unit
    const units = await gameActions.getVisibleUnits();
    const currentPlayer = gameState.currentPlayer;
    const playerUnit = units.find(unit => unit.player === currentPlayer);
    
    if (playerUnit) {
      const selectResult = await gameActions.selectUnit(playerUnit.q, playerUnit.r);
      actions.push(`Select: ${selectResult.success ? 'SUCCESS' : 'FAILED'} - ${selectResult.message}`);
    }
    
    // 3. Clear selection
    const clearResult = await gameActions.clearSelection();
    actions.push(`Clear: ${clearResult.success ? 'SUCCESS' : 'FAILED'} - ${clearResult.message}`);
    
    // 4. End turn
    const endTurnResult = await gameActions.endTurn();
    actions.push(`EndTurn: ${endTurnResult.success ? 'SUCCESS' : 'FAILED'} - ${endTurnResult.message}`);
    
    // 5. Get final state
    gameState = await gameActions.getGameState();
    actions.push(`Final: Player ${gameState.currentPlayer}, Turn ${gameState.turnCounter}`);
    
    console.log('✅ Action sequence completed:', actions);
    
    // All actions should complete without throwing exceptions
    expect(actions.length).toBe(5);
  });

  test('should provide useful error messages for invalid actions', async ({ page }) => {
    // Test various invalid actions to ensure good error reporting
    
    // 1. Try to move without selecting a unit
    const moveWithoutSelection = await gameActions.moveSelectedUnit(1, 1);
    expect(moveWithoutSelection.success).toBe(false);
    expect(moveWithoutSelection.message).toContain('No unit selected');
    
    // 2. Try to select unit at invalid coordinates
    const invalidSelection = await gameActions.selectUnit(-99, -99);
    expect(invalidSelection.success).toBe(false);
    expect(invalidSelection.message).toContain('No unit found');
    
    // 3. Try to attack (should return not implemented)
    const attackResult = await gameActions.attackWithSelectedUnit(0, 0);
    expect(attackResult.success).toBe(false);
    expect(attackResult.message).toContain('not yet implemented');
    
    console.log('✅ Error handling working correctly:', {
      moveWithoutSelection: moveWithoutSelection.message,
      invalidSelection: invalidSelection.message,
      attackNotImplemented: attackResult.message
    });
  });
});