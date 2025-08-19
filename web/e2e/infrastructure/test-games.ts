/**
 * Test Game Creation Utilities
 * 
 * Creates known game scenarios using the real game creation APIs.
 * These games can be used across multiple test runs and provide
 * consistent starting states for integration tests.
 */

export interface TestGameScenario {
  gameId: string;
  name: string;
  description: string;
  players: number;
  tiles: Array<{q: number, r: number, player: number, tileType: number}>;
  units: Array<{q: number, r: number, player: number, unitType: number, availableHealth?: number, distanceLeft?: number}>;
  turnCounter?: number;
  currentPlayer?: number;
}

export interface GameCreationResult {
  success: boolean;
  gameId?: string;
  error?: string;
  url?: string;
}

/**
 * Predefined test game scenarios
 */
export const TEST_SCENARIOS: Record<string, TestGameScenario> = {
  BASIC_MOVEMENT: {
    gameId: 'test-basic-movement',
    name: 'Basic Movement Test',
    description: '3x3 map with units for testing basic movement mechanics',
    players: 2,
    tiles: [
      {q: 0, r: 0, player: 0, tileType: 1}, {q: 1, r: 0, player: 0, tileType: 1}, {q: 2, r: 0, player: 0, tileType: 1},
      {q: 0, r: 1, player: 0, tileType: 1}, {q: 1, r: 1, player: 0, tileType: 1}, {q: 2, r: 1, player: 0, tileType: 1},
      {q: 0, r: 2, player: 0, tileType: 1}, {q: 1, r: 2, player: 0, tileType: 1}, {q: 2, r: 2, player: 0, tileType: 1}
    ],
    units: [
      {q: 0, r: 0, player: 1, unitType: 1, availableHealth: 100, distanceLeft: 3, turnCounter: 1},
      {q: 2, r: 2, player: 2, unitType: 1, availableHealth: 100, distanceLeft: 3, turnCounter: 1}
    ],
    turnCounter: 1,
    currentPlayer: 1
  },

  COMBAT_BASIC: {
    gameId: 'test-combat-basic',
    name: 'Basic Combat Test',
    description: 'Adjacent units ready for combat testing',
    players: 2,
    tiles: [
      {q: 0, r: 0, player: 0, tileType: 1}, {q: 1, r: 0, player: 0, tileType: 1}, {q: 2, r: 0, player: 0, tileType: 1},
      {q: 0, r: 1, player: 0, tileType: 1}, {q: 1, r: 1, player: 0, tileType: 1}, {q: 2, r: 1, player: 0, tileType: 1}
    ],
    units: [
      {q: 0, r: 0, player: 1, unitType: 1, availableHealth: 100, distanceLeft: 3},
      {q: 1, r: 0, player: 2, unitType: 1, availableHealth: 100, distanceLeft: 3},
      {q: 0, r: 1, player: 1, unitType: 1, availableHealth: 100, distanceLeft: 3}
    ],
    turnCounter: 1,
    currentPlayer: 1
  },

  TURN_FLOW: {
    gameId: 'test-turn-flow',
    name: 'Turn Management Test',
    description: 'Multi-unit scenario for testing turn mechanics',
    players: 2,
    tiles: Array.from({length: 25}, (_, i) => ({
      q: i % 5,
      r: Math.floor(i / 5),
      player: 0,
      tileType: 1
    })),
    units: [
      {q: 0, r: 0, player: 1, unitType: 1, availableHealth: 100, distanceLeft: 3},
      {q: 1, r: 0, player: 1, unitType: 1, availableHealth: 100, distanceLeft: 3},
      {q: 3, r: 4, player: 2, unitType: 1, availableHealth: 100, distanceLeft: 3},
      {q: 4, r: 4, player: 2, unitType: 1, availableHealth: 100, distanceLeft: 3}
    ],
    turnCounter: 1,
    currentPlayer: 1
  },

  ERROR_HANDLING: {
    gameId: 'test-error-handling',
    name: 'Error Handling Test',
    description: 'Constrained scenario for testing invalid moves and error conditions',
    players: 2,
    tiles: [
      {q: 0, r: 0, player: 0, tileType: 1}, {q: 1, r: 0, player: 0, tileType: 2}, // Mountain blocks path
      {q: 0, r: 1, player: 0, tileType: 1}, {q: 1, r: 1, player: 0, tileType: 1}
    ],
    units: [
      {q: 0, r: 0, player: 1, unitType: 1, availableHealth: 100, distanceLeft: 1}, // Limited movement
      {q: 1, r: 1, player: 2, unitType: 1, availableHealth: 100, distanceLeft: 3}
    ],
    turnCounter: 1,
    currentPlayer: 1
  }
};

/**
 * Create a test game using the real game creation API
 */
export async function createTestGame(
  scenario: TestGameScenario, 
  serverUrl: string = 'http://localhost:8080'
): Promise<GameCreationResult> {
  try {
    // Load the world IDs from the setup script
    const fs = require('fs');
    const path = require('path');
    const configPath = path.join(__dirname, 'test-world-ids.json');
    
    let worldIds: Record<string, string> = {};
    try {
      const configData = fs.readFileSync(configPath, 'utf8');
      worldIds = JSON.parse(configData);
    } catch (error) {
      throw new Error(`Test worlds not found. Please run: npm run setup-test-worlds\nError: ${error}`);
    }

    // Map scenario IDs to world IDs
    const worldIdMap: Record<string, string> = {
      'test-basic-movement': 'basic-movement',
      'test-combat-basic': 'combat-basic', 
      'test-turn-flow': 'turn-flow',
      'test-error-handling': 'error-handling'
    };

    const worldId = worldIdMap[scenario.gameId];
    if (!worldId || !worldIds[worldId]) {
      throw new Error(`No world found for scenario ${scenario.gameId}. Available worlds: ${Object.keys(worldIds).join(', ')}`);
    }

    const gamePayload = {
      game: {
        world_id: worldId, // Use the persistent test world
        name: scenario.name,
        description: scenario.description,
        creator_id: 'test-user',
        max_players: scenario.players,
        current_players: [
          { id: 'player1', name: 'Player 1', type: 'human' },
          { id: 'player2', name: 'Player 2', type: 'human' }
        ]
      },
      game_state: {
        turn_counter: scenario.turnCounter || 1,
        current_player: scenario.currentPlayer || 1
      }
    };

    console.log(`🔧 Creating game with world: ${worldId}`);
    
    // Create the game using the real API
    const createResponse = await fetch(`${serverUrl}/api/v1/games`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(gamePayload)
    });

    if (!createResponse.ok) {
      const error = await createResponse.text();
      return {
        success: false,
        error: `Failed to create game: ${createResponse.status} ${error}`
      };
    }

    const gameData = await createResponse.json();
    console.log('🔍 Game creation response:', JSON.stringify(gameData, null, 2));
    
    const actualGameId = gameData.game?.id;
    
    if (!actualGameId) {
      return {
        success: false,
        error: `API returned success but no game ID: ${JSON.stringify(gameData)}`
      };
    }
    
    return {
      success: true,
      gameId: actualGameId,
      url: `${serverUrl}/games/${actualGameId}/view`
    };

  } catch (error) {
    return {
      success: false,
      error: `Network error creating game: ${error}`
    };
  }
}

/**
 * Delete a test game (cleanup)
 */
export async function deleteTestGame(
  gameId: string, 
  serverUrl: string = 'http://localhost:8080'
): Promise<boolean> {
  try {
    const response = await fetch(`${serverUrl}/api/v1/games/${gameId}`, {
      method: 'DELETE'
    });
    return response.ok;
  } catch (error) {
    console.warn(`Failed to delete test game ${gameId}:`, error);
    return false;
  }
}

/**
 * Ensure all test games exist on the server
 */
export async function ensureTestGamesExist(serverUrl: string = 'http://localhost:8080'): Promise<void> {
  const results = await Promise.all(
    Object.values(TEST_SCENARIOS).map(scenario => 
      createTestGame(scenario, serverUrl)
    )
  );

  const failures = results.filter(r => !r.success);
  if (failures.length > 0) {
    console.warn('Some test games failed to create:', failures);
  }

  console.log(`Test games ready: ${results.filter(r => r.success).length}/${results.length}`);
}

/**
 * Get the URL for a test game
 */
export function getTestGameUrl(
  scenarioKey: keyof typeof TEST_SCENARIOS, 
  serverUrl: string = 'http://localhost:8080'
): string {
  const scenario = TEST_SCENARIOS[scenarioKey];
  return `${serverUrl}/games/${scenario.gameId}/view`;
}