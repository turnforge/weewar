/**
 * Rules Engine Validation Tests
 * Tests game rules consistency across different maps and scenarios
 */

import { createTestGameState, validateMovementOptions, validateAttackOptions } from './helpers/wasmTestUtils';
import { SMALL_TEST_MAP, MEDIUM_TEST_MAP, COMBAT_TEST_MAP, TEST_MAPS, testMapToJSON } from './fixtures/testMaps';
import { GameState } from '../frontend/components/GameState';

describe('Rules Engine Validation', () => {
  
  describe('Movement Rules Consistency', () => {
    
    test('should provide consistent movement rules across different map sizes', async () => {
      const results: Array<{ mapName: string; unitMovement: any; valid: boolean }> = [];
      
      // Test movement rules on different map sizes
      for (const [mapName, testMap] of Object.entries(TEST_MAPS)) {
        const { gameState, cleanup } = await createTestGameState();
        
        try {
          const mapData = testMapToJSON(testMap);
          await gameState.createGameFromMap(mapData);
          
          // Test movement for first unit
          const unit = testMap.units[0];
          const movementResponse = gameState.getMovementOptions(unit.q, unit.r);
          const validation = validateMovementOptions(movementResponse);
          
          results.push({
            mapName,
            unitMovement: movementResponse,
            valid: validation.isValid
          });
          
        } finally {
          cleanup();
        }
      }
      
      // All maps should return valid movement data structures
      results.forEach(result => {
        expect(result.valid).toBe(true);
      });
      
      // Log results for manual verification
      console.log('Movement rules validation results:', 
        results.map(r => ({ 
          map: r.mapName, 
          valid: r.valid,
          moveCount: r.unitMovement.success ? r.unitMovement.data.length : 0
        }))
      );
    });

    test('should respect movement range limitations', async () => {
      const { gameState, cleanup } = await createTestGameState();
      
      try {
        const mapData = testMapToJSON(MEDIUM_TEST_MAP);
        await gameState.createGameFromMap(mapData);
        
        // Test movement options for a unit
        const unit = MEDIUM_TEST_MAP.units[0];
        const movementResponse = gameState.getMovementOptions(unit.q, unit.r);
        
        if (movementResponse.success) {
          // Movement should be within reasonable range (not the entire map)
          const moveCount = movementResponse.data.length;
          const totalTiles = Object.keys(MEDIUM_TEST_MAP.tiles).length;
          
          // Movement options should be less than total tiles (reasonable constraint)
          expect(moveCount).toBeLessThan(totalTiles);
          
          console.log("mrp: ", movementResponse.data)
          // All movement coordinates should be valid tile positions
          movementResponse.data.forEach((move: any) => {
            // All movement coordinates should be valid tile positions
            const foundIndex = MEDIUM_TEST_MAP.tiles.findIndex((tile) => tile.q == move.coord.q && tile.r == move.coord.r)
            expect(foundIndex >= 0).toBe(true)
          });
        }
        
      } finally {
        cleanup();
      }
    });

    test('should prevent movement to occupied tiles', async () => {
      const { gameState, cleanup } = await createTestGameState();
      
      try {
        const mapData = testMapToJSON(COMBAT_TEST_MAP);
        await gameState.createGameFromMap(mapData);
        
        // Get movement options for first unit
        const unit = COMBAT_TEST_MAP.units[0];
        const movementResponse = gameState.getMovementOptions(unit.q, unit.r);
        
        if (movementResponse.success) {
          // Check that movement options don't include tiles occupied by other units
          const occupiedPositions = COMBAT_TEST_MAP.units
            .filter(u => u.q !== unit.q || u.r !== unit.r) // Exclude the moving unit itself
            .map(u => `${u.q},${u.r}`);
          
          const movementPositions = movementResponse.data.map((move: any) => 
            `${move.coord.q},${move.coord.r}`
          );
          
          occupiedPositions.forEach(occupied => {
            expect(movementPositions).not.toContain(occupied);
          });
        }
        
      } finally {
        cleanup();
      }
    });
  });

  describe('Combat Rules Validation', () => {
    
    /*
    test('should provide consistent attack rules across maps', async () => {
      const results: Array<{ mapName: string; attackTargets: number; valid: boolean }> = [];
      
      for (const [mapName, testMap] of Object.entries(TEST_MAPS)) {
        const { gameState, cleanup } = await createTestGameState();
        
        try {
          const mapData = testMapToJSON(testMap);
          await gameState.createGameFromMap(mapData);
          
          // Test attack options for first unit
          const unit = testMap.units[0];
          const attackResponse = gameState.getAttackOptions(unit.q, unit.r);
          const validation = validateAttackOptions(attackResponse);
          
          results.push({
            mapName,
            attackTargets: attackResponse.success ? attackResponse.data?.length : 0,
            valid: validation.isValid
          });
          
        } finally {
          cleanup();
        }
      }
      
      // All maps should return valid attack data structures
      results.forEach(result => {
        expect(result.valid).toBe(true);
      });
      
      console.log('Attack rules validation results:', results);
    });

    test('should only allow attacks on enemy units within range', async () => {
      const { gameState, cleanup } = await createTestGameState();
      
      try {
        const mapData = testMapToJSON(COMBAT_TEST_MAP);
        await gameState.createGameFromMap(mapData);
        
        const attacker = COMBAT_TEST_MAP.units[0]; // Player 1 unit
        const attackResponse = gameState.getAttackOptions(attacker.q, attacker.r);
        
        if (attackResponse.success) {
          // Should only be able to attack enemy units
          const enemyUnits = COMBAT_TEST_MAP.units.filter(u => u.player !== attacker.player);
          const attackPositions = attackResponse.data.map((attack: any) => 
            `${attack.coord.q},${attack.coord.r}`
          );
          
          // All attack targets should correspond to enemy unit positions
          attackPositions.forEach((attackPos: any) => {
            const hasEnemyAtPosition = enemyUnits.some(enemy => 
              `${enemy.q},${enemy.r}` === attackPos
            );
            expect(hasEnemyAtPosition).toBe(true);
          });
        }
        
      } finally {
        cleanup();
      }
    });
   */
  });

  /*
  describe('Turn Management Rules', () => {
    test('should enforce turn order across different player counts', async () => {
      const { gameState, cleanup } = await createTestGameState();
      
      try {
        const mapData = testMapToJSON(SMALL_TEST_MAP);
        await gameState.createGameFromMap(mapData);
        
        const initialState = gameState.getGameState();
        const initialPlayer = initialState.currentPlayer;
        const playerCount = initialState.players.length;
        
        // End turn and verify player changes
        const newState = gameState.endTurn();
        
        // Player should change or turn counter should increment
        const playerChanged = newState.currentPlayer !== initialPlayer;
        const turnIncremented = newState.turnCounter > initialState.turnCounter;
        
        expect(playerChanged || turnIncremented).toBe(true);
        
        // Current player should be valid
        expect(newState.currentPlayer).toBeGreaterThan(0);
        expect(newState.currentPlayer).toBeLessThanOrEqual(playerCount);
        
      } finally {
        cleanup();
      }
    });

    test('should maintain game state consistency after multiple turns', async () => {
      const { gameState, cleanup } = await createTestGameState();
      
      try {
        const mapData = testMapToJSON(SMALL_TEST_MAP);
        await gameState.createGameFromMap(mapData);
        
        const initialState = gameState.getGameState();
        const initialUnitCount = initialState.allUnits.length;
        
        // End several turns
        let currentState = initialState;
        for (let i = 0; i < 4; i++) {
          currentState = gameState.endTurn();
        }
        
        // Game should still be valid
        expect(currentState.allUnits.length).toBe(initialUnitCount); // Units shouldn't disappear
        expect(currentState.currentPlayer).toBeGreaterThan(0);
        expect(currentState.turnCounter).toBeGreaterThan(initialState.turnCounter);
        
      } finally {
        cleanup();
      }
    });
  });

  describe('Game State Integrity', () => {
    
    test('should maintain unit consistency across different maps', async () => {
      const results: Array<{ mapName: string; expectedUnits: number; actualUnits: number }> = [];
      
      for (const [mapName, testMap] of Object.entries(TEST_MAPS)) {
        const { gameState, cleanup } = await createTestGameState();
        
        try {
          const mapData = testMapToJSON(testMap);
          const gameData = await gameState.createGameFromMap(mapData);
          
          results.push({
            mapName,
            expectedUnits: testMap.units.length,
            actualUnits: gameData.allUnits.length
          });
          
        } finally {
          cleanup();
        }
      }
      
      // All maps should have matching unit counts
      results.forEach(result => {
        expect(result.actualUnits).toBe(result.expectedUnits);
      });
    });

    test('should preserve tile information correctly', async () => {
      const { gameState, cleanup } = await createTestGameState();
      
      try {
        const mapData = testMapToJSON(MEDIUM_TEST_MAP);
        await gameState.createGameFromMap(mapData);
        
        // Test several tile positions
        const testTiles = [
          { q: 0, r: 0, expectedType: 1 },
          { q: 1, r: 1, expectedType: 2 }, // Mountain
          { q: 2, r: 0, expectedType: 3 }  // Different terrain
        ];
        
        testTiles.forEach(testTile => {
          const terrainStats = gameState.getTerrainStatsAt(testTile.q, testTile.r);
          if (terrainStats) {
            expect(terrainStats.tileType).toBe(testTile.expectedType);
          }
        });
        
      } finally {
        cleanup();
      }
    });
  });

  describe('Error Handling and Edge Cases', () => {
    
    test('should handle empty maps gracefully', async () => {
      const { gameState, cleanup } = await createTestGameState();
      
      try {
        const emptyMap = {
          name: "Empty Map",
          tiles: [ {q: 0, r: 0,  player: 0, tileType: 1 } ],
          units: [],
          players: [1, 2]
        };
        
        const mapData = JSON.stringify(emptyMap);
        const gameData = await gameState.createGameFromMap(mapData);
        
        expect(gameData.allUnits.length).toBe(0);
        expect(gameData.players.length).toBeGreaterThan(0);
        
      } finally {
        cleanup();
      }
    });

    test('should handle single unit maps', async () => {
      const { gameState, cleanup } = await createTestGameState();
      
      try {
        const singleUnitMap = {
          name: "Single Unit Map",
          tiles: [ {q: 0, r: 0,  player: 0, tileType: 1 }, {q: 1, r: 0,  player: 0, tileType: 1 } ],
          units: [{ unitType: 1, q: 0, r: 0, player: 1 }],
          players: [1]
        };
        
        const mapData = JSON.stringify(singleUnitMap);
        const gameData = await gameState.createGameFromMap(mapData);
        
        expect(gameData.allUnits.length).toBe(1);
        
        // Should still be able to get movement options
        const movementResponse = gameState.getMovementOptions(0, 0);
        expect(validateMovementOptions(movementResponse).isValid).toBe(true);
        
      } finally {
        cleanup();
      }
    });
  });
 */
});
