import { ITheme, ThemeInfo, PLAYER_COLORS, CITY_TERRAIN_IDS } from './BaseTheme';

/**
 * Default theme using the original PNG assets from v1
 * These assets are pre-colored for each player and don't need post-processing
 */
export default class DefaultTheme implements ITheme {
    protected basePath = '/static/assets/v1';
    protected themeName = 'Default (PNG)';
    protected themeVersion = '1.0.0';
    
    // Unit names from the original game
    private unitNames: { [key: number]: string } = {
        1: 'Infantry',
        2: 'Mech',
        3: 'Recon',
        4: 'Tank',
        5: 'Medium Tank',
        6: 'Neo Tank',
        7: 'APC',
        8: 'Artillery',
        9: 'Rocket',
        10: 'Anti-Air',
        11: 'Missile',
        12: 'Fighter',
        13: 'Bomber',
        14: 'B-Copter',
        15: 'T-Copter',
        16: 'Battleship',
        17: 'Cruiser',
        18: 'Lander',
        19: 'Sub',
        20: 'Mech',
        21: 'Missile (Std)',
        22: 'Missile (Nuke)',
        24: 'Sailboat',
        25: 'Artillery (Mega)',
        26: 'Artillery (Quick)',
        27: 'Medic',
        28: 'Stratotanker',
        29: 'Engineer',
        30: 'Goliath RC',
        31: 'Tugboat',
        32: 'Sea Mine',
        33: 'Drone',
        37: 'Cruiser',
        38: 'Missile (Anti Air)',
        39: 'Aircraft Carrier',
        40: 'Miner',
        41: 'Paratrooper',
        44: 'Anti Aircraft (Advanced)',
    };
    
    // Terrain names from the original game
    private terrainNames: { [key: number]: string } = {
        0: 'Clear',
        1: 'Land Base',
        2: 'Naval Base',
        3: 'Airport Base',
        4: 'Dessert',
        5: 'Grass',
        6: 'Hospital',
        7: 'Mountains',
        8: 'Swamp',
        9: 'Forest',
        10: 'Water (Regular)',
        12: 'Lava',
        14: 'Water (Shallow)',
        15: 'Water (Deep)',
        16: 'Missile Silo',
        17: 'Bridge (Regular)',
        18: 'Bridge (Shallow)',
        19: 'Bridge (Deep)',
        20: 'Mines',
        21: 'City',
        22: 'Road',
        23: 'Water (Rocky)',
        25: 'Guard Tower',
        26: 'Snow',
    };
    
    // Nature terrains that only have neutral colors
    private natureTerrains = [4, 5, 7, 8, 9, 10, 12, 14, 15, 17, 18, 19, 22, 23, 26];
    
    getThemeInfo(): ThemeInfo {
        return {
            name: this.themeName,
            version: this.themeVersion,
            basePath: this.basePath,
            supportsTinting: false,
            needsPostProcessing: false,
            assetType: 'png',
            playerColors: PLAYER_COLORS,
        };
    }
    
    getUnitName(unitId: number): string | undefined {
        return this.unitNames[unitId];
    }
    
    getTerrainName(terrainId: number): string | undefined {
        return this.terrainNames[terrainId];
    }
    
    getUnitPath(unitId: number): string | undefined {
        // PNG assets don't use templates, return undefined
        return undefined;
    }
    
    getTilePath(terrainId: number): string | undefined {
        // PNG assets don't use templates, return undefined
        return undefined;
    }
    
    /**
     * Returns the direct path to a pre-colored unit asset
     */
    getUnitAssetPath(unitId: number, playerId: number): string | undefined {
        if (!this.unitNames[unitId]) {
            return undefined;
        }
        return `${this.basePath}/Units/${unitId}/${playerId}.png`;
    }
    
    /**
     * Returns the direct path to a pre-colored terrain asset
     */
    getTileAssetPath(terrainId: number, playerId: number): string | undefined {
        if (!this.terrainNames[terrainId]) {
            return undefined;
        }
        
        // Nature terrains always use player 0 (neutral)
        const effectivePlayer = this.natureTerrains.includes(terrainId) ? 0 : playerId;
        return `${this.basePath}/Tiles/${terrainId}/${effectivePlayer}.png`;
    }
    
    getAvailableUnits(): number[] {
        return Object.keys(this.unitNames).map(id => parseInt(id));
    }
    
    getAvailableTerrains(): number[] {
        return Object.keys(this.terrainNames).map(id => parseInt(id));
    }
    
    async loadUnit(unitId: number, playerId: number): Promise<string> {
        // PNG assets are loaded directly, not as SVG templates
        throw new Error('Default theme uses pre-colored PNG assets. Use getUnitAssetPath() instead.');
    }
    
    async loadTile(terrainId: number, playerId?: number): Promise<string> {
        // PNG assets are loaded directly, not as SVG templates
        throw new Error('Default theme uses pre-colored PNG assets. Use getTileAssetPath() instead.');
    }
}