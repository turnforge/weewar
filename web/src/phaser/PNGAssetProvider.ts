import { BaseAssetProvider } from './AssetProvider';
import { AllowedUnitIDs } from '../ColorsAndNames';

/**
 * Asset provider for standard PNG assets
 * Loads individual PNG files for each terrain/unit and player color combination
 */
export class PNGAssetProvider extends BaseAssetProvider {
    constructor() {
        super();
        this.assetSize = { width: 64, height: 64 };
    }
    
    /**
     * Override to handle nature terrains properly
     */
    getTerrainTexture(tileType: number, player: number): string {
        // Nature terrains always use the neutral texture
        if (this.natureTerrains.includes(tileType)) {
            return `terrain_${tileType}_0`;
        }
        return `terrain_${tileType}_${player}`;
    }
    
    preloadAssets(): void {
        if (!this.loader) {
            console.error('[PNGAssetProvider] Loader not configured');
            return;
        }
        
        // Load city terrains with all color variations
        this.cityTerrains.forEach(type => {
            for (let color = 0; color <= this.maxPlayers; color++) {
                const assetPath = `/static/assets/v1/Tiles/${type}/${color}.png`;
                const textureKey = `terrain_${type}_${color}`;
                this.loader.image(textureKey, assetPath);
                
                if (color === 0) {
                    // Create default alias
                    this.loader.image(`terrain_${type}`, assetPath);
                }
            }
        });
        
        // Load nature terrains with only default texture
        this.natureTerrains.forEach(type => {
            const assetPath = `/static/assets/v1/Tiles/${type}/0.png`;
            this.loader.image(`terrain_${type}`, assetPath);
            this.loader.image(`terrain_${type}_0`, assetPath);
            
            // Nature terrains don't have player colors, but we'll handle aliases differently
            // Since we can't add duplicate textures, we'll just use the base texture key
        });
        
        // Load unit assets with player colors
        AllowedUnitIDs.forEach(type => {
            for (let color = 0; color <= this.maxPlayers; color++) {
                const assetPath = `/static/assets/v1/Units/${type}/${color}.png`;
                const textureKey = `unit_${type}_${color}`;
                this.loader.image(textureKey, assetPath);
                
                if (color === 0) {
                    // Create default alias
                    this.loader.image(`unit_${type}`, assetPath);
                }
            }
        });
        
        // Set up error handling for missing assets
        this.loader.on('loaderror', (file: any) => {
            console.warn(`[PNGAssetProvider] Failed to load asset: ${file.key} from ${file.url}`);
        });
    }
}