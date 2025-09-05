import { BaseAssetProvider } from './AssetProvider';
import { AllowedUnitIDs } from '../ColorsAndNames';

/**
 * Player color definitions for SVG templating
 */
interface PlayerColors {
    primary: string;
    secondary: string;
    accent?: string;
}

/**
 * Asset provider that uses SVG templates with color replacement
 * Loads base SVG templates and generates color variations for each player
 */
export class TemplateSVGAssetProvider extends BaseAssetProvider {
    private rasterSize: number;
    private fallbackToPNG: boolean;
    private processedAssets: Set<string> = new Set();
    
    // Player color schemes
    private playerColors: PlayerColors[] = [
        { primary: '#808080', secondary: '#606060', accent: '#404040' }, // Player 0 - Neutral gray
        { primary: '#ff4444', secondary: '#cc0000', accent: '#ffaaaa' }, // Player 1 - Red
        { primary: '#4444ff', secondary: '#0000cc', accent: '#aaaaff' }, // Player 2 - Blue
        { primary: '#44ff44', secondary: '#00cc00', accent: '#aaffaa' }, // Player 3 - Green
        { primary: '#ffff44', secondary: '#cccc00', accent: '#ffffaa' }, // Player 4 - Yellow
        { primary: '#ff44ff', secondary: '#cc00cc', accent: '#ffaaff' }, // Player 5 - Magenta
        { primary: '#44ffff', secondary: '#00cccc', accent: '#aaffff' }, // Player 6 - Cyan
        { primary: '#ff8844', secondary: '#cc6600', accent: '#ffccaa' }, // Player 7 - Orange
        { primary: '#8844ff', secondary: '#6600cc', accent: '#ccaaff' }, // Player 8 - Purple
        { primary: '#88ff44', secondary: '#66cc00', accent: '#ccffaa' }, // Player 9 - Lime
        { primary: '#ff4488', secondary: '#cc0066', accent: '#ffaacc' }, // Player 10 - Pink
        { primary: '#44ff88', secondary: '#00cc66', accent: '#aaffcc' }, // Player 11 - Teal
        { primary: '#8888ff', secondary: '#6666cc', accent: '#ccccff' }, // Player 12 - Light Blue
    ];
    
    constructor(rasterSize: number = 160, fallbackToPNG: boolean = true) {
        super();
        this.rasterSize = rasterSize;
        this.fallbackToPNG = fallbackToPNG;
        this.assetSize = { width: rasterSize, height: rasterSize };
    }
    
    preloadAssets(): void {
        if (!this.loader) {
            console.error('[TemplateSVGAssetProvider] Loader not configured');
            return;
        }
        
        // Load base SVG templates as text
        this.loadTerrainTemplates();
        this.loadUnitTemplates();
        
        // Set up error handling
        this.loader.on('loaderror', (file: any) => {
            console.warn(`[TemplateSVGAssetProvider] Failed to load template: ${file.key}`);
            
            if (this.fallbackToPNG && file.key.includes('_template')) {
                // Try to load PNG fallback
                this.loadPNGFallback(file.key);
            }
        });
    }
    
    private loadTerrainTemplates(): void {
        // Load city terrain templates
        this.cityTerrains.forEach(type => {
            const templatePath = `/static/assets/v1/Tiles/${type}/template.svg`;
            this.loader.text(`terrain_${type}_template`, templatePath);
        });
        
        // Load nature terrain templates
        this.natureTerrains.forEach(type => {
            const templatePath = `/static/assets/v1/Tiles/${type}/template.svg`;
            this.loader.text(`terrain_${type}_template`, templatePath);
        });
    }
    
    private loadUnitTemplates(): void {
        // Load unit templates
        AllowedUnitIDs.forEach(type => {
            const templatePath = `/static/assets/v1/Units/${type}/template.svg`;
            this.loader.text(`unit_${type}_template`, templatePath);
        });
    }
    
    private loadPNGFallback(templateKey: string): void {
        // Extract type from template key
        const match = templateKey.match(/(terrain|unit)_(\d+)_template/);
        if (!match) return;
        
        const assetType = match[1];
        const typeId = match[2];
        
        // Load PNG versions for all player colors
        for (let color = 0; color <= this.maxPlayers; color++) {
            const pngPath = assetType === 'terrain' 
                ? `/static/assets/v1/Tiles/${typeId}/${color}.png`
                : `/static/assets/v1/Units/${typeId}/${color}.png`;
            
            const textureKey = `${assetType}_${typeId}_${color}`;
            this.loader.image(textureKey, pngPath);
        }
    }
    
    protected onLoadComplete(): void {
        // Don't mark as ready yet - we need to post-process
        // The scene will call postProcessAssets() after load completes
    }
    
    async postProcessAssets(): Promise<void> {
        console.log('[TemplateSVGAssetProvider] Starting post-processing of SVG templates');
        
        const promises: Promise<void>[] = [];
        
        // Process terrain templates
        for (const type of [...this.cityTerrains, ...this.natureTerrains]) {
            const templateKey = `terrain_${type}_template`;
            const svgTemplate = this.scene.cache.text.get(templateKey);
            
            if (svgTemplate) {
                // For nature terrains, only create neutral variant
                const maxColor = this.natureTerrains.includes(type) ? 0 : this.maxPlayers;
                
                for (let player = 0; player <= maxColor; player++) {
                    const textureKey = `terrain_${type}_${player}`;
                    promises.push(this.createColorVariant(svgTemplate, textureKey, player));
                    
                    // For nature terrains, create aliases for all players
                    if (this.natureTerrains.includes(type) && player === 0) {
                        for (let p = 1; p <= this.maxPlayers; p++) {
                            this.processedAssets.add(`terrain_${type}_${p}`);
                        }
                    }
                }
            }
        }
        
        // Process unit templates
        for (const unitType of AllowedUnitIDs) {
            const templateKey = `unit_${unitType}_template`;
            const svgTemplate = this.scene.cache.text.get(templateKey);
            
            if (svgTemplate) {
                for (let player = 0; player <= this.maxPlayers; player++) {
                    const textureKey = `unit_${unitType}_${player}`;
                    promises.push(this.createColorVariant(svgTemplate, textureKey, player));
                }
            }
        }
        
        // Wait for all processing to complete
        await Promise.all(promises);
        
        // Create aliases for nature terrains
        this.createNatureTerrainAliases();
        
        console.log(`[TemplateSVGAssetProvider] Post-processing complete. Processed ${this.processedAssets.size} assets`);
        
        // Now mark as ready
        this.ready = true;
        if (this.onComplete) {
            this.onComplete();
        }
    }
    
    private async createColorVariant(
        svgTemplate: string,
        textureKey: string,
        player: number
    ): Promise<void> {
        const colors = this.playerColors[player] || this.playerColors[0];
        
        // Replace color placeholders in SVG
        let processedSVG = svgTemplate
            .replace(/\{\{PRIMARY_COLOR\}\}/g, colors.primary)
            .replace(/\{\{SECONDARY_COLOR\}\}/g, colors.secondary)
            .replace(/\{\{ACCENT_COLOR\}\}/g, colors.accent || colors.primary)
            .replace(/\{\{PLAYER_ID\}\}/g, player.toString())
            .replace(/\{\{PLAYER_NUMBER\}\}/g, (player + 1).toString());
        
        // Add gradient definitions if not present
        if (processedSVG.includes('{{PLAYER_GRADIENT}}')) {
            const gradientDef = `
                <defs>
                    <linearGradient id="playerGradient${player}" x1="0%" y1="0%" x2="100%" y2="100%">
                        <stop offset="0%" style="stop-color:${colors.primary};stop-opacity:1" />
                        <stop offset="100%" style="stop-color:${colors.secondary};stop-opacity:1" />
                    </linearGradient>
                    <radialGradient id="playerRadialGradient${player}">
                        <stop offset="0%" style="stop-color:${colors.primary};stop-opacity:1" />
                        <stop offset="100%" style="stop-color:${colors.secondary};stop-opacity:0.8" />
                    </radialGradient>
                </defs>
            `;
            
            processedSVG = processedSVG
                .replace('</svg>', `${gradientDef}</svg>`)
                .replace(/\{\{PLAYER_GRADIENT\}\}/g, `url(#playerGradient${player})`)
                .replace(/\{\{PLAYER_RADIAL_GRADIENT\}\}/g, `url(#playerRadialGradient${player})`);
        }
        
        // Convert processed SVG to texture
        await this.svgToTexture(processedSVG, textureKey);
        this.processedAssets.add(textureKey);
    }
    
    private async svgToTexture(svgString: string, textureKey: string): Promise<void> {
        return new Promise((resolve, reject) => {
            try {
                // Create blob from SVG string
                const blob = new Blob([svgString], { type: 'image/svg+xml;charset=utf-8' });
                const url = URL.createObjectURL(blob);
                
                // Create image element
                const img = new Image();
                
                img.onload = () => {
                    try {
                        // Create canvas and draw SVG at desired resolution
                        const canvas = document.createElement('canvas');
                        canvas.width = this.rasterSize;
                        canvas.height = this.rasterSize;
                        const ctx = canvas.getContext('2d');
                        
                        if (!ctx) {
                            throw new Error('Failed to get canvas context');
                        }
                        
                        // Enable image smoothing for better quality
                        ctx.imageSmoothingEnabled = true;
                        ctx.imageSmoothingQuality = 'high';
                        
                        // Draw the SVG to canvas at the target size
                        ctx.drawImage(img, 0, 0, this.rasterSize, this.rasterSize);
                        
                        // Add to Phaser's texture manager
                        this.scene.textures.addCanvas(textureKey, canvas);
                        
                        // Clean up
                        URL.revokeObjectURL(url);
                        resolve();
                    } catch (error) {
                        console.error(`[TemplateSVGAssetProvider] Error processing ${textureKey}:`, error);
                        URL.revokeObjectURL(url);
                        reject(error);
                    }
                };
                
                img.onerror = () => {
                    console.error(`[TemplateSVGAssetProvider] Failed to load SVG for ${textureKey}`);
                    URL.revokeObjectURL(url);
                    reject(new Error(`Failed to load SVG for ${textureKey}`));
                };
                
                img.src = url;
            } catch (error) {
                console.error(`[TemplateSVGAssetProvider] Error creating texture ${textureKey}:`, error);
                reject(error);
            }
        });
    }
    
    private createNatureTerrainAliases(): void {
        // For nature terrains, we'll handle the aliasing in the getTerrainTexture method
        // Since Phaser doesn't support addTexture, we can't create true aliases
        // Instead, getTerrainTexture will return the base texture for all players
    }
    
    getTerrainTexture(tileType: number, player: number): string {
        const textureKey = `terrain_${tileType}_${player}`;
        
        // For nature terrains, always use the neutral texture
        if (this.natureTerrains.includes(tileType)) {
            return `terrain_${tileType}_0`;
        }
        
        return textureKey;
    }
    
    dispose(): void {
        // Clean up processed textures
        this.processedAssets.forEach(key => {
            if (this.scene.textures.exists(key)) {
                this.scene.textures.remove(key);
            }
        });
        
        this.processedAssets.clear();
        super.dispose();
    }
}