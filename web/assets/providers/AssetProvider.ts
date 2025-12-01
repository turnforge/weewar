import * as Phaser from 'phaser';
import { TILE_WIDTH, TILE_HEIGHT } from '../../pages/common/hexUtils';
import { ITheme, DEFAULT_PLAYER_COLORS } from '../themes/BaseTheme';
import DefaultTheme from '../themes/default';
import ModernTheme from '../themes/modern';
import FantasyTheme from '../themes/fantasy';

/**
 * Theme registry - maps theme names to their implementations
 */
const THEME_REGISTRY: Record<string, new () => ITheme> = {
    default: DefaultTheme,
    modern: ModernTheme,
    fantasy: FantasyTheme,
};

/**
 * Universal Asset Provider that works with any theme
 * The theme provides the configuration and paths, the provider handles all Phaser operations
 */
export class AssetProvider {
    private loader: Phaser.Loader.LoaderPlugin;
    private scene: Phaser.Scene;
    private theme: ITheme;
    private themeName: string;
    private ready: boolean = false;
    private rasterSize: number;
    private debugMode: boolean;
    private processedAssets: Set<string> = new Set();
    
    // These could also come from the theme if we want full flexibility
    private readonly maxPlayers = 12;
    private readonly natureTerrains = [4, 5, 7, 8, 9, 10, 12, 14, 15, 17, 18, 19, 22, 23, 26];
    
    onProgress?: (progress: number) => void;
    onComplete?: () => void;
    
    constructor(themeName: string = 'default', rasterSize: number = 160, debugMode: boolean = false) {
        this.themeName = themeName;
        this.rasterSize = rasterSize;
        this.debugMode = debugMode;
        
        // Initialize the theme
        const ThemeClass = THEME_REGISTRY[themeName];
        if (!ThemeClass) {
            console.warn(`[AssetProvider] Theme "${themeName}" not found, defaulting to 'default'`);
            this.theme = new DefaultTheme();
            this.themeName = 'default';
        } else {
            this.theme = new ThemeClass();
        }
        
        const themeInfo = this.theme.getThemeInfo();
        console.log(`[AssetProvider] Initialized with ${themeInfo.name} theme (${themeInfo.assetType}, needs processing: ${themeInfo.needsPostProcessing})`);
    }
    
    /**
     * Configure the provider with Phaser's loader and scene
     */
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
            this.onLoadComplete();
        });
    }
    
    /**
     * Queue all assets for loading based on theme configuration
     */
    async preloadAssets(): Promise<void> {
        if (!this.loader) {
            console.error('[AssetProvider] Loader not configured');
            return;
        }
        
        const themeInfo = this.theme.getThemeInfo();
        console.log(`[AssetProvider] Loading ${themeInfo.name} assets`);
        
        if (!themeInfo.needsPostProcessing) {
            // Load pre-colored assets directly (PNG or pre-colored SVG)
            await this.loadPreColoredAssets();
        } else {
            // Load templates that need color processing
            await this.loadTemplates();
        }
    }
    
    /**
     * Load pre-colored assets (like PNGs with player colors baked in)
     */
    private async loadPreColoredAssets(): Promise<void> {
        const themeInfo = this.theme.getThemeInfo();
        const units = this.theme.getAvailableUnits();
        const terrains = this.theme.getAvailableTerrains();
        
        console.log(`[AssetProvider] Loading ${units.length} units and ${terrains.length} terrains (pre-colored ${themeInfo.assetType})`);
        
        // Load terrain assets
        for (const terrainId of terrains) {
            // Nature terrains only need neutral variant
            const maxPlayer = this.natureTerrains.includes(terrainId) ? 0 : this.maxPlayers;
            
            for (let player = 0; player <= maxPlayer; player++) {
                const path = this.theme.getTileAssetPath?.(terrainId, player);
                if (path) {
                    const key = `terrain_${terrainId}_${player}`;
                    
                    if (themeInfo.assetType === 'png') {
                        this.loader.image(key, path);
                    } else {
                        // Pre-colored SVG
                        this.loader.svg(key, path, { width: this.rasterSize, height: this.rasterSize });
                    }
                    
                    this.processedAssets.add(key);
                }
            }
            
            // Create aliases for nature terrains
            if (this.natureTerrains.includes(terrainId)) {
                for (let p = 1; p <= this.maxPlayers; p++) {
                    // These will reference the neutral texture
                    this.processedAssets.add(`terrain_${terrainId}_${p}`);
                }
            }
        }
        
        // Load unit assets
        for (const unitId of units) {
            for (let player = 0; player <= this.maxPlayers; player++) {
                const path = this.theme.getUnitAssetPath?.(unitId, player);
                if (path) {
                    const key = `unit_${unitId}_${player}`;
                    
                    if (themeInfo.assetType === 'png') {
                        this.loader.image(key, path);
                    } else {
                        // Pre-colored SVG
                        this.loader.svg(key, path, { width: this.rasterSize, height: this.rasterSize });
                    }
                    
                    this.processedAssets.add(key);
                }
            }
        }
    }
    
    /**
     * Load SVG templates that need color processing
     */
    private async loadTemplates(): Promise<void> {
        const units = this.theme.getAvailableUnits();
        const terrains = this.theme.getAvailableTerrains();
        
        console.log(`[AssetProvider] Loading ${units.length} unit and ${terrains.length} terrain templates`);
        
        // Load terrain templates
        for (const terrainId of terrains) {
            const path = this.theme.getTilePath(terrainId);
            if (path) {
                const key = `terrain_${terrainId}_template`;
                this.loader.text(key, path);
            }
        }
        
        // Load unit templates
        for (const unitId of units) {
            const path = this.theme.getUnitPath(unitId);
            if (path) {
                const key = `unit_${unitId}_template`;
                this.loader.text(key, path);
            }
        }
    }
    
    /**
     * Called when base loading is complete
     */
    private onLoadComplete(): void {
        const themeInfo = this.theme.getThemeInfo();
        
        if (!themeInfo.needsPostProcessing) {
            // Pre-colored assets are ready immediately
            this.ready = true;
            if (this.onComplete) {
                this.onComplete();
            }
        } else {
            // Templates need post-processing
            console.log('[AssetProvider] Templates loaded, awaiting post-processing');
        }
    }
    
    /**
     * Post-process templates to create player-colored variants
     */
    async postProcessAssets(): Promise<void> {
        const themeInfo = this.theme.getThemeInfo();
        
        if (!themeInfo.needsPostProcessing) {
            console.log('[AssetProvider] No post-processing needed for this theme');
            return;
        }
        
        console.log('[AssetProvider] Creating player-colored variants from templates');
        
        const promises: Promise<void>[] = [];
        
        // Process terrain templates
        const terrains = this.theme.getAvailableTerrains();
        for (const terrainId of terrains) {
            const templateKey = `terrain_${terrainId}_template`;
            const template = this.scene.cache.text.get(templateKey);
            
            if (template) {
                // Nature terrains only need neutral
                const maxPlayer = this.natureTerrains.includes(terrainId) ? 0 : this.maxPlayers;
                
                for (let player = 0; player <= maxPlayer; player++) {
                    const textureKey = `terrain_${terrainId}_${player}`;
                    promises.push(this.createColorVariant(template, textureKey, player));
                }
                
                // Create aliases for nature terrains
                if (this.natureTerrains.includes(terrainId)) {
                    for (let p = 1; p <= this.maxPlayers; p++) {
                        this.processedAssets.add(`terrain_${terrainId}_${p}`);
                    }
                }
            }
        }
        
        // Process unit templates
        const units = this.theme.getAvailableUnits();
        for (const unitId of units) {
            const templateKey = `unit_${unitId}_template`;
            const template = this.scene.cache.text.get(templateKey);
            
            if (template) {
                for (let player = 0; player <= this.maxPlayers; player++) {
                    const textureKey = `unit_${unitId}_${player}`;
                    promises.push(this.createColorVariant(template, textureKey, player));
                }
            }
        }
        
        // Wait for all processing
        await Promise.all(promises);
        console.log(`[AssetProvider] Created ${promises.length} textures`);
        
        // Mark as ready
        this.ready = true;
        if (this.onComplete) {
            this.onComplete();
        }
    }
    
    /**
     * Create a player-colored variant from an SVG template
     */
    private async createColorVariant(svgTemplate: string, textureKey: string, player: number): Promise<void> {
        // Apply player colors
        const processedSVG = this.applyPlayerColors(svgTemplate, player);
        
        // Convert to texture
        await this.svgToTexture(processedSVG, textureKey);
        this.processedAssets.add(textureKey);
    }
    
    /**
     * Apply player colors to an SVG template
     */
    private applyPlayerColors(svgContent: string, playerId: number): string {
        // If theme has custom color application, use it
        if (this.theme.applyPlayerColors) {
            return this.theme.applyPlayerColors(svgContent, playerId);
        }
        
        // Default implementation using playerColor gradient
        const parser = new DOMParser();
        const svgDoc = parser.parseFromString(svgContent, 'image/svg+xml');
        
        const gradient = svgDoc.querySelector('linearGradient#playerColor');
        if (gradient) {
            const colors = DEFAULT_PLAYER_COLORS[playerId] || DEFAULT_PLAYER_COLORS[0];
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
     * Convert SVG string to Phaser texture
     */
    private async svgToTexture(svgString: string, textureKey: string): Promise<void> {
        return new Promise((resolve, reject) => {
            try {
                const blob = new Blob([svgString], { type: 'image/svg+xml;charset=utf-8' });
                const url = URL.createObjectURL(blob);
                const img = new Image();
                
                img.onload = () => {
                    try {
                        const canvas = document.createElement('canvas');
                        canvas.width = this.rasterSize;
                        canvas.height = this.rasterSize;
                        const ctx = canvas.getContext('2d');
                        
                        if (!ctx) {
                            throw new Error('Failed to get canvas context');
                        }
                        
                        ctx.imageSmoothingEnabled = true;
                        ctx.imageSmoothingQuality = 'high';
                        ctx.drawImage(img, 0, 0, this.rasterSize, this.rasterSize);
                        
                        this.scene.textures.addCanvas(textureKey, canvas);
                        
                        URL.revokeObjectURL(url);
                        resolve();
                    } catch (error) {
                        URL.revokeObjectURL(url);
                        reject(error);
                    }
                };
                
                img.onerror = () => {
                    URL.revokeObjectURL(url);
                    reject(new Error(`Failed to load SVG for ${textureKey}`));
                };
                
                img.src = url;
            } catch (error) {
                reject(error);
            }
        });
    }
    
    /**
     * Get the texture key for a terrain tile
     */
    getTerrainTexture(tileType: number, player: number): string {
        // Nature terrains always use neutral
        if (this.natureTerrains.includes(tileType)) {
            return `terrain_${tileType}_0`;
        }
        return `terrain_${tileType}_${player}`;
    }
    
    /**
     * Get the texture key for a unit
     */
    getUnitTexture(unitType: number, player: number): string {
        return `unit_${unitType}_${player}`;
    }
    
    /**
     * Get asset dimensions based on theme
     */
    getAssetSize(): { width: number, height: number } {
        const themeInfo = this.theme.getThemeInfo();
        
        // PNG themes typically use 64x64
        if (themeInfo.assetType === 'png') {
            return { width: 64, height: 64 };
        }
        
        // SVG themes use the raster size
        return { width: this.rasterSize, height: this.rasterSize };
    }
    
    /**
     * Get display dimensions for sprites
     */
    getDisplaySize(): { width: number, height: number } {
        const themeInfo = this.theme.getThemeInfo();
        
        if (themeInfo.assetType === 'png') {
            return { width: 64, height: 64 };
        }
        
        // SVG assets need adjusted height for hex overlap
        return { width: TILE_WIDTH, height: TILE_HEIGHT - 4 };
    }
    
    /**
     * Check if assets are ready
     */
    isReady(): boolean {
        return this.ready;
    }
    
    /**
     * Get the current theme
     */
    getTheme(): ITheme {
        return this.theme;
    }
    
    /**
     * Change the active theme
     */
    setTheme(themeName: string): void {
        if (this.themeName === themeName) return;
        
        const ThemeClass = THEME_REGISTRY[themeName];
        if (!ThemeClass) {
            throw new Error(`Theme "${themeName}" not found`);
        }
        
        this.dispose();
        this.themeName = themeName;
        this.theme = new ThemeClass();
        this.ready = false;
        
        const themeInfo = this.theme.getThemeInfo();
        console.log(`[AssetProvider] Switched to ${themeInfo.name} theme`);
    }
    
    /**
     * Clean up loaded assets
     */
    dispose(): void {
        this.processedAssets.forEach(key => {
            if (this.scene?.textures.exists(key)) {
                this.scene.textures.remove(key);
            }
        });
        
        this.processedAssets.clear();
        this.ready = false;
    }
}
