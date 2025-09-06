/**
 * Base Theme Class
 * Contains common functionality for all themes
 */

// Import from the game's color definitions
// These IDs are consistent across all themes
export const CITY_TERRAIN_IDS = [1, 2, 3, 6, 16, 20, 21, 25]; // Base, Hospital, Silo, Mines, City, Tower, etc.
export const NATURE_TERRAIN_IDS = [4, 5, 7, 8, 9, 10, 12, 14, 15, 23, 26]; // Desert, Grass, Mountains, etc.
export const BRIDGE_TERRAIN_IDS = [17, 18, 19]; // Regular, Shallow, Deep bridges
export const ROAD_TERRAIN_ID = 22;

// All valid unit IDs from ColorsAndNames.ts
export const UNIT_IDS = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 37, 38, 39, 40, 41, 44];

// Player color definitions - matches ColorsAndNames.ts and original Weewar colors
// Based on the PLAYER_COLORS text classes in dark mode (which is what's typically shown)
// These also need to match the PNG asset player colors (1.png = red, 2.png = blue, etc.)
export const PLAYER_COLORS = {
  0: { primary: '#888888', secondary: '#666666' }, // Neutral/unowned - gray
  1: { primary: '#f87171', secondary: '#dc2626' }, // Player 1: RED (text-red-400/600)
  2: { primary: '#60a5fa', secondary: '#2563eb' }, // Player 2: BLUE (text-blue-400/600)
  3: { primary: '#4ade80', secondary: '#16a34a' }, // Player 3: GREEN (text-green-400/600)
  4: { primary: '#facc15', secondary: '#ca8a04' }, // Player 4: YELLOW (text-yellow-400/600)
  5: { primary: '#fb923c', secondary: '#ea580c' }, // Player 5: ORANGE (text-orange-400/600)
  6: { primary: '#c084fc', secondary: '#9333ea' }, // Player 6: PURPLE (text-purple-400/600)
  7: { primary: '#f472b6', secondary: '#db2777' }, // Player 7: PINK (text-pink-400/600)
  8: { primary: '#22d3ee', secondary: '#0891b2' }, // Player 8: CYAN (text-cyan-400/600)
  9: { primary: '#22d3ee', secondary: '#0891b2' }, // Player 9: CYAN (same as 8 in ColorsAndNames)
  10: { primary: '#22d3ee', secondary: '#0891b2' }, // Player 10: CYAN (same as 8 in ColorsAndNames)
  11: { primary: '#22d3ee', secondary: '#0891b2' }, // Player 11: CYAN (same as 8 in ColorsAndNames)
  12: { primary: '#22d3ee', secondary: '#0891b2' }, // Player 12: CYAN (same as 8 in ColorsAndNames)
} as any;

/**
 * Theme interface that all themes must implement
 */
export interface ITheme {
  loadUnit(unitId: number, playerId: number): Promise<string>;
  loadTile(terrainId: number, playerId?: number): Promise<string>;
  getUnitPath(unitId: number): string | undefined;
  getTilePath(terrainId: number): string | undefined;
  getUnitAssetPath?(unitId: number, playerId: number): string | undefined;
  getTileAssetPath?(terrainId: number, playerId: number): string | undefined;
  getAvailableUnits(): number[]
  getAvailableTerrains(): number[]
  getThemeInfo(): ThemeInfo;
  getUnitName(unitId: number): string | undefined;
  getTerrainName(terrainId: number): string | undefined;
  applyPlayerColors?(svgContent: string, playerId: number): string;
}

export interface ThemeInfo {
  name: string;
  version: string;
  basePath: string;
  supportsTinting: boolean;
  needsPostProcessing: boolean;
  assetType: 'svg' | 'png' | 'mixed';
  playerColors: typeof PLAYER_COLORS;
}

export interface ThemeMapping {
  units: {
    [key: string]: {
      old: string;
      name: string;
      image: string;
    };
  };
  terrains: {
    [key: string]: {
      old: string;
      name: string;
      image: string;
    };
  };
}

/**
 * Base Theme Class with common functionality
 */
export abstract class BaseTheme implements ITheme {
  protected abstract basePath: string;
  protected abstract themeName: string;
  protected abstract themeVersion: string;
  protected unitMapping: ThemeMapping['units'];
  protected terrainMapping: ThemeMapping['terrains'];

  constructor(mapping: ThemeMapping) {
    this.unitMapping = mapping.units;
    this.terrainMapping = mapping.terrains;
  }

  /**
   * Gets the file path for a unit by ID
   */
  getUnitPath(unitId: number): string | undefined {
    const unit = this.unitMapping[unitId.toString()];
    return unit ? `${this.basePath}/${unit.image}` : undefined;
  }

  /**
   * Gets the file path for a terrain tile by ID
   */
  getTilePath(terrainId: number): string | undefined {
    const terrain = this.terrainMapping[terrainId.toString()];
    return terrain ? `${this.basePath}/${terrain.image}` : undefined;
  }

  /**
   * Loads a unit SVG with the specified player's colors
   * @param unitId The unit type ID
   * @param playerId The player ID (0-12, where 0 is neutral)
   * @returns Promise<string> The SVG content with player colors applied
   */
  async loadUnit(unitId: number, playerId: number): Promise<string> {
    const path = this.getUnitPath(unitId);
    if (!path) {
      throw new Error(`Unit ID ${unitId} not found in ${this.themeName} theme mapping`);
    }

    try {
      const response = await fetch(path);
      if (!response.ok) {
        throw new Error(`Failed to fetch unit: ${response.statusText}`);
      }
      const svgText = await response.text();
      
      // Apply player colors (including neutral/0)
      return this.applyPlayerColors(svgText, playerId);
    } catch (error) {
      console.error(`Failed to load unit ${unitId} for player ${playerId}:`, error);
      throw error;
    }
  }

  /**
   * Loads a terrain tile SVG with optional player colors (for city tiles)
   * @param terrainId The terrain type ID
   * @param playerId Optional player ID (0-12) for owned city tiles, 0 or undefined for neutral
   * @returns Promise<string> The SVG content with player colors applied if applicable
   */
  async loadTile(terrainId: number, playerId?: number): Promise<string> {
    const path = this.getTilePath(terrainId);
    if (!path) {
      throw new Error(`Terrain ID ${terrainId} not found in ${this.themeName} theme mapping`);
    }

    try {
      const response = await fetch(path);
      if (!response.ok) {
        throw new Error(`Failed to fetch tile: ${response.statusText}`);
      }
      const svgText = await response.text();
      
      // Apply player colors if it's a city tile
      // Use neutral gray (player 0) if no player specified but it's a city tile
      if (this.isCityTile(terrainId)) {
        const effectivePlayerId = playerId ?? 0; // Default to neutral
        return this.applyPlayerColors(svgText, effectivePlayerId);
      }
      
      return svgText;
    } catch (error) {
      console.error(`Failed to load tile ${terrainId} for player ${playerId}:`, error);
      throw error;
    }
  }

  /**
   * Applies player colors to an SVG by modifying the playerColor gradient
   * This is a common implementation that works for all themes using the playerColor gradient system
   */
  applyPlayerColors(svgContent: string, playerId: number): string {
    const parser = new DOMParser();
    const svgDoc = parser.parseFromString(svgContent, 'image/svg+xml');
    
    // Find the playerColor gradient
    const gradient = svgDoc.querySelector('linearGradient#playerColor');
    if (gradient) {
      const colors = PLAYER_COLORS[playerId] || PLAYER_COLORS[0]; // Fall back to neutral
      const stops = gradient.querySelectorAll('stop');
      if (stops.length >= 2) {
        stops[0].setAttribute('stop-color', colors.secondary);
        stops[1].setAttribute('stop-color', colors.primary);
      }
    }
    
    const serializer = new XMLSerializer();
    return serializer.serializeToString(svgDoc);
  }

  /**
   * Checks if a terrain ID is a city/building tile that should be tinted
   */
  protected isCityTile(terrainId: number): boolean {
    return CITY_TERRAIN_IDS.includes(terrainId);
  }

  /**
   * Checks if a terrain ID is a nature tile
   */
  protected isNatureTile(terrainId: number): boolean {
    return NATURE_TERRAIN_IDS.includes(terrainId);
  }

  /**
   * Checks if a terrain ID is a bridge
   */
  protected isBridgeTile(terrainId: number): boolean {
    return BRIDGE_TERRAIN_IDS.includes(terrainId);
  }

  /**
   * Gets theme metadata
   */
  getThemeInfo(): ThemeInfo {
    return {
      name: this.themeName,
      version: this.themeVersion,
      basePath: this.basePath,
      supportsTinting: true,
      needsPostProcessing: true,
      assetType: 'svg',
      playerColors: PLAYER_COLORS
    };
  }

  /**
   * Helper method to get unit name by ID
   */
  getUnitName(unitId: number): string | undefined {
    const unit = this.unitMapping[unitId.toString()];
    return unit?.name;
  }

  /**
   * Helper method to get terrain name by ID
   */
  getTerrainName(terrainId: number): string | undefined {
    const terrain = this.terrainMapping[terrainId.toString()];
    return terrain?.name;
  }

  /**
   * Validates if a unit ID exists in this theme
   */
  hasUnit(unitId: number): boolean {
    return unitId.toString() in this.unitMapping;
  }

  /**
   * Validates if a terrain ID exists in this theme
   */
  hasTerrain(terrainId: number): boolean {
    return terrainId.toString() in this.terrainMapping;
  }

  /**
   * Gets all available unit IDs in this theme
   */
  getAvailableUnits(): number[] {
    return Object.keys(this.unitMapping).map(id => parseInt(id)).sort((a, b) => a - b);
  }

  /**
   * Gets all available terrain IDs in this theme
   */
  getAvailableTerrains(): number[] {
    return Object.keys(this.terrainMapping).map(id => parseInt(id)).sort((a, b) => a - b);
  }
}
