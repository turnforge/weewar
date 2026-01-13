import { CrossingType } from "../../gen/lilbattle/v1/models/models_pb"
import {
  ThemeManifest,
  ThemeManifestSchema,
  PlayerColor,
  UnitMapping,
  TerrainMapping
} from "../../gen/lilbattle/v1/models/themes_pb"
import { fromJson, type JsonValue } from "@bufbuild/protobuf"

/**
 * Base Theme Class
 * Contains common functionality for all themes
 */

// Re-export proto types for convenience
export type { ThemeManifest, PlayerColor, UnitMapping, TerrainMapping }

// Terrain classification constants (RulesEngine domain, shared across all themes)
export const CITY_TERRAIN_IDS = [1, 2, 3, 6, 16, 20, 21, 25]; // Base, Hospital, Silo, Mines, City, Tower, etc.
export const NATURE_TERRAIN_IDS = [4, 5, 7, 8, 9, 10, 12, 14, 15, 23, 26]; // Desert, Grass, Mountains, etc.
export const BRIDGE_TERRAIN_IDS = [17, 18, 19]; // Regular, Shallow, Deep bridges
export const ROAD_TERRAIN_ID = 22;
export const WATER_TERRAIN_IDS = [10, 14, 15, 23]


// Tile type constants - crossings map to these for display purposes
export const TILE_TYPE_PLAINS = 5;
export const TILE_TYPE_ROAD = 22;
export const TILE_TYPE_BRIDGE_SHALLOW = 18;
export const TILE_TYPE_BRIDGE_REGULAR = 17;
export const TILE_TYPE_BRIDGE_ROCKY = 23;
export const TILE_TYPE_BRIDGE_DEEP = 19;

// Underlying water tile types for determining bridge depth
export const TILE_TYPE_WATER_SHALLOW = 14;
export const TILE_TYPE_WATER_REGULAR = 10;
export const TILE_TYPE_WATER_ROCKY = 23;
export const TILE_TYPE_WATER_DEEP = 15;

// Fallback player colors (used if mapping.json doesn't have playerColors)
export const DEFAULT_PLAYER_COLORS: { [key: number]: PlayerColor } = {
  0:  { primary: '#888888', secondary: '#666666', name: 'Neutral' },
  1:  { primary: '#60a5fa', secondary: '#2563eb', name: 'Blue' },
  2:  { primary: '#f87171', secondary: '#dc2626', name: 'Red' },
  3:  { primary: '#facc15', secondary: '#ca8a04', name: 'Yellow' },
  4:  { primary: '#f0f0f0', secondary: '#888888', name: 'White' },
  5:  { primary: '#f472b6', secondary: '#db2777', name: 'Pink' },
  6:  { primary: '#fb923c', secondary: '#ea580c', name: 'Orange' },
  7:  { primary: '#1f2937', secondary: '#111827', name: 'Black' },
  8:  { primary: '#2dd4bf', secondary: '#14b8a6', name: 'Teal' },
  9:  { primary: '#1e3a8a', secondary: '#1e40af', name: 'Navy Blue' },
  10: { primary: '#a16207', secondary: '#854d0e', name: 'Brown' },
  11: { primary: '#22d3ee', secondary: '#0891b2', name: 'Cyan' },
  12: { primary: '#c084fc', secondary: '#9333ea', name: 'Purple' },
} as any;

/**
 * Extended ThemeInfo for runtime use (adds fields not in proto)
 */
export interface ThemeInfoRuntime {
  name: string;
  version: string;
  basePath: string;
  assetType: 'svg' | 'png' | 'mixed';
  needsPostProcessing: boolean;
  supportsTinting: boolean;
  playerColors: { [key: number]: PlayerColor };
}

/**
 * Theme interface that all themes must implement
 */
export interface ITheme {
  loadUnit(unitId: number, playerId: number): Promise<string>;
  loadTile(terrainId: number, playerId?: number): Promise<string>;
  isCityTile(tileId: number): boolean
  isWaterTile(tileId: number): boolean
  getUnitPath(unitId: number): string | undefined;
  getTilePath(terrainId: number): string | undefined;
  getUnitAssetPath?(unitId: number, playerId: number): string | undefined;
  getTileAssetPath?(terrainId: number, playerId: number): string | undefined;
  getAvailableUnits(): number[]
  getAvailableTerrains(): number[]
  getThemeInfo(): ThemeInfoRuntime;
  getUnitName(unitId: number): string | undefined;
  getTerrainName(terrainId: number): string | undefined;
  getUnitDescription?(unitId: number): string | undefined;
  getTerrainDescription?(terrainId: number): string | undefined;
  setUnitImage(unitId: number, playerId: number, targetElement: HTMLElement): Promise<void>;
  setTileImage(tileId: number, playerId: number, targetElement: HTMLElement): Promise<void>;
  applyPlayerColors?(svgContent: string, playerId: number): string;
  getPlayerColor?(playerId: number): PlayerColor | undefined;
  canPlaceCrossing(tileType: number, crossingType: number): boolean;
  defaultCrossingTerrain(crossingType: number): number;
  getCrossingDisplayTileType(crossingType: number, underlyingTileType: number): number;
}

/**
 * Parse mapping.json into ThemeManifest proto
 */
export function parseThemeManifest(json: JsonValue): ThemeManifest {
  return fromJson(ThemeManifestSchema, json);
}

/**
 * Base Theme Class with common functionality
 * Uses ThemeManifest proto for data storage
 */
export abstract class BaseTheme implements ITheme {
  protected manifest: ThemeManifest;

  constructor(manifest: ThemeManifest) {
    this.manifest = manifest;
    // Populate default player colors if not specified in manifest
    if (!this.manifest.playerColors || Object.keys(this.manifest.playerColors).length === 0) {
      this.manifest.playerColors = { ...DEFAULT_PLAYER_COLORS };
    }
  }

  /**
   * Gets the file path for a unit by ID
   */
  getUnitPath(unitId: number): string | undefined {
    const unit = this.manifest.units[unitId];
    if (!unit) return undefined;
    return `${this.manifest.themeInfo?.basePath}/${unit.image}`;
  }

  /**
   * Gets the file path for a terrain tile by ID
   */
  getTilePath(terrainId: number): string | undefined {
    const terrain = this.manifest.terrains[terrainId];
    if (!terrain) return undefined;
    return `${this.manifest.themeInfo?.basePath}/${terrain.image}`;
  }

  /**
   * Loads a unit SVG with the specified player's colors
   */
  async loadUnit(unitId: number, playerId: number): Promise<string> {
    const path = this.getUnitPath(unitId);
    if (!path) {
      throw new Error(`Unit ID ${unitId} not found in ${this.manifest.themeInfo?.name} theme mapping`);
    }

    const response = await fetch(path);
    if (!response.ok) {
      throw new Error(`Failed to fetch unit: ${response.statusText}`);
    }
    const svgText = await response.text();
    return this.applyPlayerColors(svgText, playerId);
  }

  /**
   * Loads a terrain tile SVG with optional player colors (for city tiles)
   */
  async loadTile(terrainId: number, playerId?: number): Promise<string> {
    const path = this.getTilePath(terrainId);
    if (!path) {
      throw new Error(`Terrain ID ${terrainId} not found in ${this.manifest.themeInfo?.name} theme mapping`);
    }

    const response = await fetch(path);
    if (!response.ok) {
      throw new Error(`Failed to fetch tile: ${response.statusText}`);
    }
    const svgText = await response.text();

    if (this.isCityTile(terrainId)) {
      const effectivePlayerId = playerId ?? 0;
      return this.applyPlayerColors(svgText, effectivePlayerId);
    }

    return svgText;
  }

  /**
   * Applies player colors to an SVG by modifying the playerColor gradient
   */
  applyPlayerColors(svgContent: string, playerId: number): string {
    const parser = new DOMParser();
    const svgDoc = parser.parseFromString(svgContent, 'image/svg+xml');

    const gradient = svgDoc.querySelector('linearGradient#playerColor');
    if (gradient) {
      const colors = this.getPlayerColor(playerId);
      if (colors) {
        const stops = gradient.querySelectorAll('stop');
        if (stops.length >= 2) {
          stops[0].setAttribute('stop-color', colors.secondary);
          stops[1].setAttribute('stop-color', colors.primary);
        }
      }
    }

    const serializer = new XMLSerializer();
    return serializer.serializeToString(svgDoc);
  }

  isCityTile(terrainId: number): boolean {
    return CITY_TERRAIN_IDS.includes(terrainId);
  }

  isWaterTile(tileType: number): boolean {
    return WATER_TERRAIN_IDS.includes(tileType)
  }

  isNatureTile(terrainId: number): boolean {
    return NATURE_TERRAIN_IDS.includes(terrainId);
  }

  isBridgeTile(terrainId: number): boolean {
    return BRIDGE_TERRAIN_IDS.includes(terrainId);
  }

  getThemeInfo(): ThemeInfoRuntime {
    const info = this.manifest.themeInfo;
    return {
      name: info?.name ?? 'Unknown',
      version: info?.version ?? '1.0.0',
      basePath: info?.basePath ?? '',
      assetType: (info?.assetType as 'svg' | 'png' | 'mixed') ?? 'svg',
      needsPostProcessing: info?.needsPostProcessing ?? true,
      supportsTinting: true,
      playerColors: this.manifest.playerColors as { [key: number]: PlayerColor }
    };
  }

  getPlayerColor(playerId: number): PlayerColor | undefined {
    return this.manifest.playerColors[playerId] || this.manifest.playerColors[0] || DEFAULT_PLAYER_COLORS[playerId];
  }

  getUnitName(unitId: number): string | undefined {
    return this.manifest.units[unitId]?.name;
  }

  getTerrainName(terrainId: number): string | undefined {
    return this.manifest.terrains[terrainId]?.name;
  }

  getUnitDescription(unitId: number): string | undefined {
    return this.manifest.units[unitId]?.description;
  }

  getTerrainDescription(terrainId: number): string | undefined {
    return this.manifest.terrains[terrainId]?.description;
  }

  hasUnit(unitId: number): boolean {
    return unitId in this.manifest.units;
  }

  hasTerrain(terrainId: number): boolean {
    return terrainId in this.manifest.terrains;
  }

  getAvailableUnits(): number[] {
    return Object.keys(this.manifest.units).map(id => parseInt(id)).sort((a, b) => a - b);
  }

  getAvailableTerrains(): number[] {
    return Object.keys(this.manifest.terrains).map(id => parseInt(id)).sort((a, b) => a - b);
  }

  async setUnitImage(unitId: number, playerId: number, targetElement: HTMLElement): Promise<void> {
    try {
      const svgContent = await this.loadUnit(unitId, playerId);
      const blob = new Blob([svgContent], { type: 'image/svg+xml;charset=utf-8' });
      const url = URL.createObjectURL(blob);

      targetElement.innerHTML = '';
      const img = document.createElement('img');
      img.src = url;
      img.alt = this.getUnitName(unitId) || `Unit ${unitId}`;
      img.className = 'w-full h-full object-contain';

      img.onload = () => URL.revokeObjectURL(url);
      img.onerror = () => {
        URL.revokeObjectURL(url);
        targetElement.innerHTML = '⚔️';
      };

      targetElement.appendChild(img);
    } catch (error) {
      console.error(`Failed to set unit image for unit ${unitId}, player ${playerId}:`, error);
      targetElement.innerHTML = '⚔️';
    }
  }

  async setTileImage(tileId: number, playerId: number, targetElement: HTMLElement): Promise<void> {
    try {
      const svgContent = await this.loadTile(tileId, playerId);
      const blob = new Blob([svgContent], { type: 'image/svg+xml;charset=utf-8' });
      const url = URL.createObjectURL(blob);

      targetElement.innerHTML = '';
      const img = document.createElement('img');
      img.src = url;
      img.alt = this.getTerrainName(tileId) || `Terrain ${tileId}`;
      img.className = 'w-full h-full object-contain';

      img.onload = () => URL.revokeObjectURL(url);
      img.onerror = () => {
        URL.revokeObjectURL(url);
        targetElement.innerHTML = '🏞️';
      };

      targetElement.appendChild(img);
    } catch (error) {
      console.error(`Failed to set tile image for tile ${tileId}, player ${playerId}:`, error);
      targetElement.innerHTML = '🏞️';
    }
  }

  canPlaceCrossing(tileType: number, crossingType: number): boolean {
    if (crossingType == CrossingType.ROAD) {
      return !this.isWaterTile(tileType) && !this.isCityTile(tileType)
    } else if (crossingType == CrossingType.BRIDGE) {
      return this.isWaterTile(tileType)
    }
    return false
  }

  defaultCrossingTerrain(crossingType: number): number {
    if (crossingType == CrossingType.ROAD) {
      return TILE_TYPE_PLAINS;
    } else if (crossingType == CrossingType.BRIDGE) {
      return TILE_TYPE_WATER_REGULAR;
    }
    return TILE_TYPE_PLAINS;
  }

  getCrossingDisplayTileType(crossingType: number, underlyingTileType: number): number {
    if (crossingType === CrossingType.ROAD) {
      return TILE_TYPE_ROAD;
    } else if (crossingType === CrossingType.BRIDGE) {
      switch (underlyingTileType) {
        case TILE_TYPE_WATER_SHALLOW:
          return TILE_TYPE_BRIDGE_SHALLOW;
        case TILE_TYPE_WATER_ROCKY:
          return TILE_TYPE_BRIDGE_ROCKY;
        case TILE_TYPE_WATER_DEEP:
          return TILE_TYPE_BRIDGE_DEEP;
        default:
          return TILE_TYPE_BRIDGE_REGULAR;
      }
    }
    return TILE_TYPE_ROAD;
  }
}
