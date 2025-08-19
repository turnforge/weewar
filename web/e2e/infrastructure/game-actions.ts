/**
 * High-level game action utilities for e2e testing
 * 
 * These utilities provide easy access to GameViewerPage command interface
 * and handle common testing patterns for game interactions.
 */

import { Page } from '@playwright/test';
import { ActionResult, GameStateInfo } from '../../src/GameViewerPage';

/**
 * Game action wrapper for Playwright tests
 * Provides access to GameViewerPage command interface through browser execution
 */
export class GameActions {
    constructor(private page: Page) {}

    /**
     * Select a unit at the given hex coordinates
     */
    async selectUnit(q: number, r: number): Promise<ActionResult> {
        return await this.page.evaluate(async ({ q, r }) => {
            const gameViewerPage = (window as any).gameViewerPage;
            if (!gameViewerPage) {
                return {
                    success: false,
                    message: 'GameViewerPage not found on window object',
                    error: 'Page not properly initialized'
                } as ActionResult;
            }
            return await gameViewerPage.selectUnitAt(q, r);
        }, { q, r });
    }

    /**
     * Move the currently selected unit to target coordinates
     */
    async moveSelectedUnit(q: number, r: number): Promise<ActionResult> {
        return await this.page.evaluate(async ({ q, r }) => {
            const gameViewerPage = (window as any).gameViewerPage;
            if (!gameViewerPage) {
                return {
                    success: false,
                    message: 'GameViewerPage not found on window object',
                    error: 'Page not properly initialized'
                } as ActionResult;
            }
            return await gameViewerPage.moveSelectedUnitTo(q, r);
        }, { q, r });
    }

    /**
     * Attack with the currently selected unit
     */
    async attackWithSelectedUnit(q: number, r: number): Promise<ActionResult> {
        return await this.page.evaluate(async ({ q, r }) => {
            const gameViewerPage = (window as any).gameViewerPage;
            if (!gameViewerPage) {
                return {
                    success: false,
                    message: 'GameViewerPage not found on window object',
                    error: 'Page not properly initialized'
                } as ActionResult;
            }
            return await gameViewerPage.attackWithSelectedUnit(q, r);
        }, { q, r });
    }

    /**
     * End the current player's turn
     */
    async endTurn(): Promise<ActionResult> {
        return await this.page.evaluate(async () => {
            const gameViewerPage = (window as any).gameViewerPage;
            if (!gameViewerPage) {
                return {
                    success: false,
                    message: 'GameViewerPage not found on window object',
                    error: 'Page not properly initialized'
                } as ActionResult;
            }
            return await gameViewerPage.endCurrentPlayerTurn();
        });
    }

    /**
     * Get current game state information
     */
    async getGameState(): Promise<GameStateInfo> {
        return await this.page.evaluate(async () => {
            const gameViewerPage = (window as any).gameViewerPage;
            if (!gameViewerPage) {
                return {
                    gameId: 'unknown',
                    currentPlayer: -1,
                    turnCounter: -1,
                    selectedUnit: undefined,
                    unitsCount: 0,
                    tilesCount: 0
                } as GameStateInfo;
            }
            return await gameViewerPage.getGameState();
        });
    }

    /**
     * Clear current unit selection
     */
    async clearSelection(): Promise<ActionResult> {
        return await this.page.evaluate(async () => {
            const gameViewerPage = (window as any).gameViewerPage;
            if (!gameViewerPage) {
                return {
                    success: false,
                    message: 'GameViewerPage not found on window object',
                    error: 'Page not properly initialized'
                } as ActionResult;
            }
            return gameViewerPage.clearSelection();
        });
    }

    /**
     * Wait for game to be ready for interactions
     */
    async waitForGameReady(timeoutMs: number = 10000): Promise<void> {
        await this.page.waitForFunction(() => {
            const gameViewerPage = (window as any).gameViewerPage;
            return gameViewerPage && gameViewerPage.gameState && gameViewerPage.gameState.isReady();
        }, { timeout: timeoutMs });
    }

    /**
     * Get units visible on the game board
     */
    async getVisibleUnits(): Promise<Array<{q: number, r: number, player: number}>> {
        return await this.page.evaluate(async () => {
            const gameViewerPage = (window as any).gameViewerPage;
            if (!gameViewerPage || !gameViewerPage.world) {
                return [];
            }
            
            const units = [];
            for (const [coord, unit] of Object.entries(gameViewerPage.world.units)) {
                units.push({
                    q: (unit as any).q,
                    r: (unit as any).r,
                    player: (unit as any).player
                });
            }
            return units;
        });
    }

    /**
     * Perform a complete move action: select unit + move to target
     */
    async performMove(fromQ: number, fromR: number, toQ: number, toR: number): Promise<{
        selectResult: ActionResult;
        moveResult: ActionResult;
    }> {
        const selectResult = await this.selectUnit(fromQ, fromR);
        
        if (!selectResult.success) {
            return {
                selectResult,
                moveResult: {
                    success: false,
                    message: 'Unit selection failed, move not attempted',
                    error: 'Selection prerequisite failed'
                }
            };
        }

        const moveResult = await this.moveSelectedUnit(toQ, toR);
        
        return { selectResult, moveResult };
    }

    /**
     * Assert that an action result was successful
     */
    assertActionSuccess(result: ActionResult, actionName: string): void {
        if (!result.success) {
            throw new Error(`${actionName} failed: ${result.message}${result.error ? ` (${result.error})` : ''}`);
        }
    }

    /**
     * Wait for specific player's turn
     */
    async waitForPlayerTurn(expectedPlayer: number, timeoutMs: number = 5000): Promise<void> {
        await this.page.waitForFunction(
            (player) => {
                const gameViewerPage = (window as any).gameViewerPage;
                return gameViewerPage && 
                       gameViewerPage.gameState && 
                       gameViewerPage.gameState.getCurrentPlayer() === player;
            },
            expectedPlayer,
            { timeout: timeoutMs }
        );
    }
}

/**
 * Enhanced game action results for better test reporting
 */
export interface TestActionResult extends ActionResult {
    duration?: number;
    step?: string;
    gameState?: GameStateInfo;
}

/**
 * Test-specific utilities for debugging and reporting
 */
export class GameTestUtils {
    constructor(private page: Page, private gameActions: GameActions) {}

    /**
     * Execute an action with enhanced error reporting
     */
    async executeWithReporting<T extends ActionResult>(
        actionName: string,
        action: () => Promise<T>
    ): Promise<TestActionResult> {
        const startTime = Date.now();
        
        try {
            const result = await action();
            const duration = Date.now() - startTime;
            const gameState = await this.gameActions.getGameState();
            
            return {
                ...result,
                duration,
                step: actionName,
                gameState
            } as TestActionResult;
            
        } catch (error) {
            const duration = Date.now() - startTime;
            const gameState = await this.gameActions.getGameState();
            
            return {
                success: false,
                message: `${actionName} threw exception`,
                error: error instanceof Error ? error.message : String(error),
                duration,
                step: actionName,
                gameState
            } as TestActionResult;
        }
    }

    /**
     * Take a screenshot with descriptive filename
     */
    async captureGameState(description: string): Promise<string> {
        const timestamp = Date.now();
        const filename = `game-state-${description}-${timestamp}.png`;
        await this.page.screenshot({ path: `test-results/${filename}` });
        return filename;
    }

    /**
     * Log game state for debugging
     */
    async logGameState(context: string): Promise<void> {
        const gameState = await this.gameActions.getGameState();
        const units = await this.gameActions.getVisibleUnits();
        
        console.log(`[${context}] Game State:`, {
            gameId: gameState.gameId,
            currentPlayer: gameState.currentPlayer,
            turnCounter: gameState.turnCounter,
            selectedUnit: gameState.selectedUnit,
            unitsCount: gameState.unitsCount,
            visibleUnits: units.length
        });
    }
}