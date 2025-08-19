/**
 * Basic Gameplay Integration Tests
 * 
 * Tests the real GameViewerPage via production endpoints with minimal mocking.
 * This is our starting point - we'll expand as we discover what we need.
 */

import { test, expect } from '@playwright/test';
import { TEST_SCENARIOS, createTestGame, deleteTestGame } from '../infrastructure/test-games';
import { installBasicApiMocking } from '../infrastructure/api-mocking';

test.describe('GameViewerPage - Basic Gameplay', () => {
  const SERVER_URL = 'http://localhost:8080';
  let testGameUrl: string;
  let testGameId: string;

  test.beforeAll(async () => {
    console.log('🔧 Setting up test game...');
    
    // First check if server is running
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

    // Create our test game using the real API
    const result = await createTestGame(TEST_SCENARIOS.BASIC_MOVEMENT, SERVER_URL);
    if (!result.success) {
      throw new Error(`Failed to create test game: ${result.error}\n\nMake sure the server supports the game creation API.`);
    }
    testGameUrl = result.url!;
    testGameId = result.gameId!;
    console.log(`✅ Test game created: ${testGameUrl}`);
  });

  test.afterAll(async () => {
    // Clean up the test game
    if (testGameId) {
      console.log(`🧹 Cleaning up test game: ${testGameId}`);
      const deleted = await deleteTestGame(testGameId, SERVER_URL);
      if (deleted) {
        console.log(`✅ Test game ${testGameId} deleted successfully`);
      } else {
        console.warn(`⚠️ Failed to delete test game ${testGameId} - manual cleanup may be needed`);
      }
    }
  });

  test.beforeEach(async ({ page }) => {
    // Install minimal API mocking (add more as needed)
    await installBasicApiMocking(page);
  });

  test('should load real GameViewerPage with test game data', async ({ page }) => {
    // Load the actual production page
    await page.goto(testGameUrl);
    
    // Wait for page to load completely
    await page.waitForLoadState('networkidle');
    
    // Verify the production page structure is present
    await expect(page.locator('#phaser-viewer-container')).toBeVisible();
    await expect(page.locator('#game-status')).toBeVisible();
    await expect(page.locator('#end-turn-btn')).toBeVisible();
    
    // Wait for game to finish loading
    await page.waitForFunction(() => {
      const status = document.getElementById('game-status')?.textContent;
      return status && !status.includes('Loading');
    }, { timeout: 10000 });
    
    // Verify game loaded successfully
    const gameStatus = await page.locator('#game-status').textContent();
    expect(gameStatus).not.toContain('Error');
    
    console.log('✅ Production GameViewerPage loaded successfully');
  });

  test('should display initial game state correctly', async ({ page }) => {
    await page.goto(testGameUrl);
    await page.waitForLoadState('networkidle');
    
    // Wait for game initialization
    await page.waitForFunction(() => {
      const status = document.getElementById('game-status')?.textContent;
      return status && !status.includes('Loading');
    }, { timeout: 10000 });
    
    // Verify initial turn state
    const turnCounter = await page.locator('#turn-counter').textContent();
    expect(turnCounter).toContain('Turn 1');
    
    // Verify Phaser canvas was created (game scene loaded)
    await expect(page.locator('#phaser-viewer-container canvas')).toBeVisible();
    
    // Verify no units are initially selected
    const selectedUnitInfo = page.locator('#selected-unit-info');
    const isHidden = await selectedUnitInfo.evaluate(el => 
      el.classList.contains('hidden') || el.style.display === 'none'
    );
    expect(isHidden).toBe(true);
    
    console.log('✅ Initial game state verified');
  });

  test('should handle basic user interaction', async ({ page }) => {
    await page.goto(testGameUrl);
    await page.waitForLoadState('networkidle');
    
    // Wait for game ready
    await page.waitForFunction(() => {
      const status = document.getElementById('game-status')?.textContent;
      return status && !status.includes('Loading');
    }, { timeout: 10000 });
    
    // Try clicking on the game area (basic interaction test)
    await page.locator('#phaser-viewer-container').click({ position: { x: 100, y: 100 } });
    
    // Wait a moment for any interaction processing
    await page.waitForTimeout(1000);
    
    // Verify the page is still responsive (no crashes)
    const gameStatus = await page.locator('#game-status').textContent();
    expect(gameStatus).toBeDefined();
    
    // Try clicking end turn button
    await page.locator('#end-turn-btn').click();
    await page.waitForTimeout(1000);
    
    // Verify page is still responsive after end turn
    const statusAfterEndTurn = await page.locator('#game-status').textContent();
    expect(statusAfterEndTurn).toBeDefined();
    
    console.log('✅ Basic interactions work without crashes');
  });

  // TODO: Add more specific tests as we build out the command interface
  // - Unit selection via high-level actions
  // - Movement commands
  // - Combat interactions
  // - Error handling scenarios
  
  test.skip('placeholder for command interface tests', async ({ page }) => {
    // These will be implemented as we add the command interface to GameViewerPage
    // Example structure:
    // await gameActions.selectUnit({q: 0, r: 0});
    // await gameActions.moveSelectedUnit({q: 1, r: 0});
    // await gameActions.endTurn();
  });
});