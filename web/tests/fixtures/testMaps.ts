/**
 * Test Map Fixtures
 * Contains standardized test maps for rules engine validation
 */

import { Tile, Unit } from "../../frontend/components/World"

export interface TestMap {
  name: string;
  description: string;
  tiles: Tile[];
  units: Unit[];
  players: number[];
  expectedDimensions: { rows: number; cols: number };
}

/**
 * Small 3x3 test map for basic functionality testing
 */
export const SMALL_TEST_MAP: TestMap = {
  name: "Small Test Map",
  description: "3x3 map with two units for basic movement/combat testing",
  tiles: [
    { q: 0, r: 0, player: 0, tileType: 1 },   // Plains
    { q: 1, r: 0, player: 0, tileType: 1 },   // Plains  
    { q: 2, r: 0, player: 0, tileType: 1 },   // Plains
    { q: 0, r: 1, player: 0, tileType: 1 },   // Plains
    { q: 1, r: 1, player: 0, tileType: 2 },   // Mountain
    { q: 2, r: 1, player: 0, tileType: 1 },   // Plains
    { q: 0, r: 2, player: 0, tileType: 1 },   // Plains
    { q: 1, r: 2, player: 0, tileType: 1 },   // Plains
    { q: 2, r: 2, player: 0, tileType: 1 },   // Plains
  ],
  units: [
    { unitType: 1, q: 0, r: 0, player: 1,  }, // Player 1 soldier
    { unitType: 1, q: 2, r: 2, player: 2,  }, // Player 2 soldier
  ],
  players: [1, 2],
  expectedDimensions: { rows: 3, cols: 3 }
};

/**
 * Medium test map with varied terrain for pathfinding testing
 */
export const MEDIUM_TEST_MAP: TestMap = {
  name: "Medium Test Map",
  description: "5x5 map with varied terrain and multiple units",
  tiles: [
    {q: 0, r: 0,  player: 0, tileType: 1 }, {q: 1, r: 0,  player: 0, tileType: 1 }, {q: 2, r: 0,  player: 0, tileType: 3 }, {q: 3, r: 0,  player: 0, tileType: 1 }, {q: 4, r: 0,  player: 0, tileType: 1 },
    {q: 0, r: 1,  player: 0, tileType: 1 }, {q: 1, r: 1,  player: 0, tileType: 2 }, {q: 2, r: 1,  player: 0, tileType: 3 }, {q: 3, r: 1,  player: 0, tileType: 2 }, {q: 4, r: 1,  player: 0, tileType: 1 },
    {q: 0, r: 2,  player: 0, tileType: 3 }, {q: 1, r: 2,  player: 0, tileType: 3 }, {q: 2, r: 2,  player: 0, tileType: 1 }, {q: 3, r: 2,  player: 0, tileType: 3 }, {q: 4, r: 2,  player: 0, tileType: 3 },
    {q: 0, r: 3,  player: 0, tileType: 1 }, {q: 1, r: 3,  player: 0, tileType: 2 }, {q: 2, r: 3,  player: 0, tileType: 3 }, {q: 3, r: 3,  player: 0, tileType: 2 }, {q: 4, r: 3,  player: 0, tileType: 1 },
    {q: 0, r: 4,  player: 0, tileType: 1 }, {q: 1, r: 4,  player: 0, tileType: 1 }, {q: 2, r: 4,  player: 0, tileType: 3 }, {q: 3, r: 4,  player: 0, tileType: 1 }, {q: 4, r: 4,  player: 0, tileType: 1 },
  ],
  units: [
    { unitType: 1, q: 0, r: 0, player: 1,  },
    { unitType: 1, q: 1, r: 0, player: 1,  },
    { unitType: 1, q: 3, r: 4, player: 2,  },
    { unitType: 1, q: 4, r: 4, player: 2,  },
  ],
  players: [1, 2],
  expectedDimensions: { rows: 5, cols: 5 }
};

/**
 * Combat test map with units in attack range
 */
export const COMBAT_TEST_MAP: TestMap = {
  name: "Combat Test Map",
  description: "Map designed for testing combat mechanics",
  tiles: [
    {q: 0, r: 0,  player: 0, tileType: 1 }, {q: 1, r: 0,  player: 0, tileType: 1 }, {q: 2, r: 0,  player: 0, tileType: 1 },
    {q: 0, r: 1,  player: 0, tileType: 1 }, {q: 1, r: 1,  player: 0, tileType: 1 }, {q: 2, r: 1,  player: 0, tileType: 1 },
    {q: 0, r: 2,  player: 0, tileType: 1 }, {q: 1, r: 2,  player: 0, tileType: 1 }, {q: 2, r: 2,  player: 0, tileType: 1 },
  ],
  units: [
    { unitType: 1, q: 0, r: 0, player: 1,  }, // Player 1 soldier
    { unitType: 1, q: 1, r: 0, player: 2,  }, // Player 2 soldier (adjacent)
    { unitType: 1, q: 2, r: 2, player: 2, },  // Damaged enemy unit
  ],
  players: [1, 2],
  expectedDimensions: { rows: 3, cols: 3 }
};

export const TEST_MAPS = [SMALL_TEST_MAP, MEDIUM_TEST_MAP, COMBAT_TEST_MAP]

/**
 * Convert test map to JSON format expected by WASM
 */
export function testMapToJSON(testMap: TestMap): string {
  const worldData = {
    name: testMap.name,
    description: testMap.description,
    tiles: testMap.tiles.map(tile => {return {
      q: tile.q, r: tile.r, tile_type: tile.tileType, player: tile.player
    }}),
    units: testMap.units.map(unit => {return {
      q: unit.q, r: unit.r, unit_type: unit.unitType, player: unit.player
    }}),
    players: testMap.players.map(id => ({ id, name: `Player ${id}` }))
  };
  
  return JSON.stringify(worldData);
}
