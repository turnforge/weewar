/**
 * Default Theme Provider
 * Extends BaseTheme with PNG-specific asset loading
 * Uses pre-colored PNG assets that don't need post-processing
 */

import { BaseTheme, parseThemeManifest, ThemeInfoRuntime, CITY_TERRAIN_IDS } from './BaseTheme';
import mappingData from './default/mapping.json';

const manifest = parseThemeManifest(mappingData);

/**
 * Default Theme Implementation
 * Uses PNG assets with pre-colored player variants
 */
export default class DefaultTheme extends BaseTheme {
    constructor() {
        super(manifest);
    }

    /**
     * Override getThemeInfo for PNG-specific settings
     */
    getThemeInfo(): ThemeInfoRuntime {
        const info = this.manifest.themeInfo;
        return {
            name: info?.name ?? 'Default (PNG)',
            version: info?.version ?? '1.0.0',
            basePath: info?.basePath ?? '/static/assets/themes/default',
            supportsTinting: false,
            needsPostProcessing: false,
            assetType: 'png',
            playerColors: this.manifest.playerColors as any,
        };
    }

    /**
     * Override getUnitPath - PNG assets don't use template paths
     */
    getUnitPath(unitId: number): string | undefined {
        // PNG assets use getUnitAssetPath instead
        return undefined;
    }

    /**
     * Override getTilePath - PNG assets don't use template paths
     */
    getTilePath(terrainId: number): string | undefined {
        // PNG assets use getTileAssetPath instead
        return undefined;
    }

    /**
     * Returns the direct path to a pre-colored unit asset
     */
    getUnitAssetPath(unitId: number, playerId: number): string | undefined {
        const unit = this.manifest.units[unitId];
        if (!unit) {
            return undefined;
        }
        const basePath = this.manifest.themeInfo?.basePath ?? '/static/assets/themes/default';
        // PNG assets are stored as: {basePath}/{image}/{playerId}.png
        return `${basePath}/${unit.image}/${playerId}.png`;
    }

    /**
     * Returns the direct path to a pre-colored terrain asset
     */
    getTileAssetPath(terrainId: number, playerId: number): string | undefined {
        const terrain = this.manifest.terrains[terrainId];
        if (!terrain) {
            return undefined;
        }

        const basePath = this.manifest.themeInfo?.basePath ?? '/static/assets/themes/default';
        // Only city terrains use player colors; all others use player 0 (neutral)
        const effectivePlayer = CITY_TERRAIN_IDS.includes(terrainId) ? playerId : 0;
        return `${basePath}/${terrain.image}/${effectivePlayer}.png`;
    }

    /**
     * Override loadUnit - PNG theme uses pre-colored assets
     */
    async loadUnit(unitId: number, playerId: number): Promise<string> {
        throw new Error('Default theme uses pre-colored PNG assets. Use getUnitAssetPath() instead.');
    }

    /**
     * Override loadTile - PNG theme uses pre-colored assets
     */
    async loadTile(terrainId: number, playerId?: number): Promise<string> {
        throw new Error('Default theme uses pre-colored PNG assets. Use getTileAssetPath() instead.');
    }

    /**
     * Override setUnitImage for PNG assets
     */
    async setUnitImage(unitId: number, playerId: number, targetElement: HTMLElement): Promise<void> {
        const path = this.getUnitAssetPath(unitId, playerId);
        if (!path) {
            targetElement.innerHTML = '⚔️';
            return;
        }

        targetElement.innerHTML = '';
        const img = document.createElement('img');
        img.src = path;
        img.alt = this.getUnitName(unitId) || `Unit ${unitId}`;
        img.className = 'w-full h-full object-contain';
        img.style.imageRendering = 'pixelated';

        img.onerror = () => {
            targetElement.innerHTML = '⚔️';
        };

        targetElement.appendChild(img);
    }

    /**
     * Override setTileImage for PNG assets
     */
    async setTileImage(tileId: number, playerId: number, targetElement: HTMLElement): Promise<void> {
        const path = this.getTileAssetPath(tileId, playerId);
        if (!path) {
            targetElement.innerHTML = '🏞️';
            return;
        }

        targetElement.innerHTML = '';
        const img = document.createElement('img');
        img.src = path;
        img.alt = this.getTerrainName(tileId) || `Terrain ${tileId}`;
        img.className = 'w-full h-full object-contain';
        img.style.imageRendering = 'pixelated';

        img.onerror = () => {
            targetElement.innerHTML = '🏞️';
        };

        targetElement.appendChild(img);
    }
}
