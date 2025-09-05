import * as Phaser from 'phaser';

/**
 * Interface for providing game assets to the WorldScene
 * Allows swapping between different asset packs and formats (PNG, SVG, etc.)
 */
export interface AssetProvider {
    /**
     * Configure the provider with Phaser's loader and scene
     */
    configure(loader: Phaser.Loader.LoaderPlugin, scene: Phaser.Scene): void;
    
    /**
     * Queue all assets for loading (called during Phaser preload)
     */
    preloadAssets(): void;
    
    /**
     * Post-process assets after loading (e.g., apply color templates)
     * Returns a promise that resolves when processing is complete
     */
    postProcessAssets?(): Promise<void>;
    
    /**
     * Get the texture key for a terrain tile
     */
    getTerrainTexture(tileType: number, player: number): string;
    
    /**
     * Get the texture key for a unit
     */
    getUnitTexture(unitType: number, player: number): string;
    
    /**
     * Get asset dimensions
     */
    getAssetSize(): { width: number, height: number };
    
    /**
     * Check if all assets are loaded and ready
     */
    isReady(): boolean;
    
    /**
     * Progress callback
     */
    onProgress?(progress: number): void;
    
    /**
     * Completion callback
     */
    onComplete?(): void;
    
    /**
     * Clean up loaded assets
     */
    dispose?(): void;
}

/**
 * Base class with common functionality for asset providers
 */
export abstract class BaseAssetProvider implements AssetProvider {
    protected loader: Phaser.Loader.LoaderPlugin;
    protected scene: Phaser.Scene;
    protected ready: boolean = false;
    protected assetSize: { width: number, height: number } = { width: 64, height: 64 };
    
    // Terrain and unit type definitions
    protected readonly cityTerrains = [1, 2, 3, 16, 20]; // Land Base, Naval Base, Airport Base, Missile Silo, Mines
    protected readonly natureTerrains = [4, 5, 6, 7, 8, 9, 10, 12, 14, 15, 17, 18, 19, 21, 22, 23, 25, 26];
    protected readonly maxPlayers = 12;
    
    onProgress?(progress: number): void;
    onComplete?(): void;
    
    configure(loader: Phaser.Loader.LoaderPlugin, scene: Phaser.Scene): void {
        this.loader = loader;
        this.scene = scene;
        
        // Set up progress tracking
        this.loader.on('progress', (value: number) => {
            if (this.onProgress) {
                this.onProgress(value);
            }
        });
        
        this.loader.on('complete', () => {
            // Mark as ready after base loading
            // Subclasses may override if they need post-processing
            this.onLoadComplete();
        });
    }
    
    protected onLoadComplete(): void {
        this.ready = true;
        if (this.onComplete) {
            this.onComplete();
        }
    }
    
    abstract preloadAssets(): void;
    
    getAssetSize(): { width: number, height: number } {
        return this.assetSize;
    }
    
    isReady(): boolean {
        return this.ready;
    }
    
    getTerrainTexture(tileType: number, player: number): string {
        return `terrain_${tileType}_${player}`;
    }
    
    getUnitTexture(unitType: number, player: number): string {
        return `unit_${unitType}_${player}`;
    }
    
    dispose(): void {
        // Subclasses can override to clean up textures
        this.ready = false;
    }
}