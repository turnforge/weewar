import * as Phaser from 'phaser';
import { EditablePhaserMapScene } from './EditablePhaserMapScene';
import { hexToPixel, pixelToHex, HexCoord, PixelCoord } from './hexUtils';

export class PhaserMapEditor {
    private game: Phaser.Game | null = null;
    private scene: EditablePhaserMapScene | null = null;
    private containerElement: HTMLElement | null = null;
    private sceneReadyPromise: Promise<EditablePhaserMapScene> | null = null;
    private sceneReadyResolver: ((scene: EditablePhaserMapScene) => void) | null = null;
    
    private currentTerrain: number = 1;
    private currentColor: number = 0;
    private brushSize: number = 0;
    
    // Event callbacks
    private onTileClickCallback: ((q: number, r: number) => void) | null = null;
    private onMapChangeCallback: (() => void) | null = null;
    private onReferenceScaleChangeCallback: ((x: number, y: number) => void) | null = null;
    
    constructor(containerId: string) {
        this.containerElement = document.getElementById(containerId);
        
        if (!this.containerElement) {
            throw new Error(`Container element with ID '${containerId}' not found`);
        }
        
        this.initialize();
    }
    
    private initialize() {
        // Create the scene ready promise immediately
        this.sceneReadyPromise = new Promise<EditablePhaserMapScene>((resolve) => {
            this.sceneReadyResolver = resolve;
        });
        
        const config: Phaser.Types.Core.GameConfig = {
            type: Phaser.AUTO,
            parent: this.containerElement!,
            width: '100%',
            height: '100%',
            backgroundColor: '#2c3e50',
            scene: EditablePhaserMapScene,
            scale: {
                mode: Phaser.Scale.RESIZE,
                width: '100%',
                height: '100%'
            },
            physics: {
                default: 'arcade',
                arcade: {
                    debug: false
                }
            },
            input: {
                keyboard: true,
                mouse: true
            },
            render: {
                pixelArt: true,
                antialias: false
            }
        };
        
        this.game = new Phaser.Game(config);
        
        // Get reference to the scene once it's created
        this.game.events.once('ready', () => {
            this.scene = this.game!.scene.getScene('EditablePhaserMapScene') as EditablePhaserMapScene;
            
            // Set up event listeners
            this.setupEventListeners();
            
            // Resolve the scene ready promise
            if (this.sceneReadyResolver) {
                console.log('[PhaserMapEditor] Scene is ready, resolving promise');
                this.sceneReadyResolver(this.scene);
            }
            
            console.log('[PhaserMapEditor] Phaser game initialized successfully');
        });
    }
    
    private setupEventListeners() {
        if (!this.scene) return;
        
        // Listen for tile click events from the scene
        this.scene.events.on('tileClicked', (data: { q: number; r: number }) => {
            this.handleTileClick(data.q, data.r);
        });
        
        // Listen for reference scale change events
        this.scene.events.on('referenceScaleChanged', (data: { x: number; y: number }) => {
            if (this.onReferenceScaleChangeCallback) {
                this.onReferenceScaleChangeCallback(data.x, data.y);
            }
        });
    }
    
    private handleTileClick(q: number, r: number) {
        // Let the external callback handle the logic (MapEditorPage will decide what to do)
        if (this.onTileClickCallback) {
            this.onTileClickCallback(q, r);
        }
        
        // Note: onMapChangeCallback will be called by the specific paint methods when needed
    }
    
    /**
     * Wait for scene to be ready - this should be used by all methods that need the scene
     */
    public async waitForSceneReady(): Promise<EditablePhaserMapScene> {
        if (this.scene) {
            return this.scene;
        }
        
        if (!this.sceneReadyPromise) {
            throw new Error('[PhaserMapEditor] Scene ready promise not initialized');
        }
        
        console.log('[PhaserMapEditor] Waiting for scene to be ready...');
        return this.sceneReadyPromise;
    }

    // Public API methods
    public setTerrain(terrain: number) {
        this.currentTerrain = terrain;
        console.log(`[PhaserMapEditor] Current terrain set to: ${terrain}`);
    }
    
    public setColor(color: number) {
        this.currentColor = color;
        console.log(`[PhaserMapEditor] Current color set to: ${color}`);
    }
    
    public setBrushSize(size: number) {
        this.brushSize = size;
        console.log(`[PhaserMapEditor] Brush size set to: ${size}`);
    }
    
    public setShowGrid(show: boolean) {
        this.scene?.setShowGrid(show);
    }
    
    public setShowCoordinates(show: boolean) {
        this.scene?.setShowCoordinates(show);
    }
    
    public setTheme(isDark: boolean) {
        this.scene?.setTheme(isDark);
    }
    
    public paintTile(q: number, r: number, terrain: number, color: number = 0, brushSize: number = 0) {
        if (!this.scene) return;
        
        if (brushSize === 0) {
            // Single tile
            this.scene.setTile(q, r, terrain, color);
        } else {
            // Multi-tile brush (simplified implementation)
            const radius = this.getBrushRadius(brushSize);
            
            for (let dq = -radius; dq <= radius; dq++) {
                for (let dr = -radius; dr <= radius; dr++) {
                    if (Math.abs(dq) + Math.abs(dr) + Math.abs(-dq - dr) <= radius * 2) {
                        this.scene.setTile(q + dq, r + dr, terrain, color);
                    }
                }
            }
        }
    }
    
    private getBrushRadius(brushSize: number): number {
        switch (brushSize) {
            case 0: return 0;  // Single
            case 1: return 1;  // Small (7 hexes)
            case 2: return 2;  // Medium (19 hexes)
            case 3: return 3;  // Large (37 hexes)
            case 4: return 4;  // X-Large (61 hexes)
            case 5: return 5;  // XX-Large (91 hexes)
            default: return 0;
        }
    }
    
    public removeTile(q: number, r: number) {
        this.scene?.removeTile(q, r);
    }
    
    public clearAllTiles() {
        this.scene?.clearAllTiles();
    }
    
    public clearAllUnits() {
        this.scene?.clearAllUnits();
    }
    
    public createTestPattern() {
        this.scene?.createTestPattern();
        
        if (this.onMapChangeCallback) {
            this.onMapChangeCallback();
        }
    }
    
    public resize(width: number, height: number) {
        if (this.game) {
            this.game.scale.resize(width, height);
        }
    }
    
    public getTilesData(): Array<{ q: number; r: number; terrain: number; color: number }> {
        return this.scene?.getTilesData() || [];
    }
    
    public getUnitsData(): Array<{ q: number; r: number; unitType: number; playerId: number }> {
        return this.scene?.getUnitsData() || [];
    }
    
    public async setTilesData(tiles: Array<{ q: number; r: number; terrain: number; color: number }>) {
        try {
            const scene = await this.waitForSceneReady();
            console.log(`[PhaserMapEditor] Setting tiles data: ${tiles.length} tiles`);
            
            // Wait for assets to be ready before placing tiles
            await scene.waitForAssetsReady();
            console.log(`[PhaserMapEditor] Assets ready, placing tiles`);
            
            scene.clearAllTiles();
            
            tiles.forEach(tile => {
                scene.setTile(tile.q, tile.r, tile.terrain, tile.color);
            });
            
            console.log(`[PhaserMapEditor] Successfully loaded ${tiles.length} tiles`);
        } catch (error) {
            console.error('[PhaserMapEditor] Failed to set tiles data:', error);
        }
    }
    
    // Event callbacks
    public onTileClick(callback: (q: number, r: number) => void) {
        this.onTileClickCallback = callback;
    }
    
    public onSceneReady(callback: () => void) {
        if (this.scene) {
            this.scene.onSceneReady(callback);
        }
    }
    
    public onMapChange(callback: () => void) {
        this.onMapChangeCallback = callback;
    }
    
    public onReferenceScaleChange(callback: (x: number, y: number) => void) {
        this.onReferenceScaleChangeCallback = callback;
    }
    
    // Camera controls
    public centerCamera(q: number = 0, r: number = 0) {
        if (!this.scene) return;
        
        // Convert hex coordinates to pixel coordinates
        const position = hexToPixel(q, r);
        
        // Center camera on position
        this.scene.cameras.main.centerOn(position.x, position.y);
    }
    
    public setZoom(zoom: number) {
        if (!this.scene) return;
        
        this.scene.cameras.main.setZoom(Phaser.Math.Clamp(zoom, 0.1, 3));
    }
    
    public getZoom(): number {
        return this.scene?.cameras.main.zoom || 1;
    }
    
    /**
     * Get the current viewport center in hex coordinates
     */
    public getViewportCenter(): HexCoord {
        if (!this.scene) return { q: 0, r: 0 };
        
        const camera = this.scene.cameras.main;
        
        // Get the center of the viewport in world coordinates
        const centerX = camera.scrollX + (camera.width / 2) / camera.zoom;
        const centerY = camera.scrollY + (camera.height / 2) / camera.zoom;
        
        // Convert pixel coordinates to hex coordinates
        return pixelToHex(centerX, centerY);
    }
    
    /**
     * Get the current camera position in pixel coordinates
     */
    public getCameraPosition(): PixelCoord {
        if (!this.scene) return { x: 0, y: 0 };
        
        const camera = this.scene.cameras.main;
        return { x: camera.scrollX, y: camera.scrollY };
    }
    
    // Advanced map generation methods
    public fillAllTerrain(terrain: number, color: number = 0) {
        if (!this.scene) return;
        
        const tiles = this.scene.getTilesData();
        tiles.forEach(tile => {
            this.scene!.setTile(tile.q, tile.r, terrain, color);
        });
        
        if (this.onMapChangeCallback) {
            this.onMapChangeCallback();
        }
    }
    
    public randomizeTerrain() {
        if (!this.scene) return;
        
        const terrains = [1, 2, 3, 16, 20]; // Grass, Desert, Water, Mountain, Rock
        const colors = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12];
        
        const tiles = this.scene.getTilesData();
        tiles.forEach(tile => {
            const randomTerrain = terrains[Math.floor(Math.random() * terrains.length)];
            const randomColor = colors[Math.floor(Math.random() * colors.length)];
            this.scene!.setTile(tile.q, tile.r, randomTerrain, randomColor);
        });
        
        if (this.onMapChangeCallback) {
            this.onMapChangeCallback();
        }
    }
    
    public createIslandPattern(centerQ: number = 0, centerR: number = 0, radius: number = 5) {
        if (!this.scene) return;
        
        // Clear existing tiles
        this.scene.clearAllTiles();
        
        // Create island pattern
        for (let q = -radius; q <= radius; q++) {
            for (let r = -radius; r <= radius; r++) {
                const distance = Math.abs(q) + Math.abs(r) + Math.abs(-q - r);
                
                if (distance <= radius * 2) {
                    const actualQ = centerQ + q;
                    const actualR = centerR + r;
                    
                    let terrain: number;
                    let color: number = 0;
                    
                    if (distance <= radius) {
                        terrain = 1; // Grass in center
                    } else if (distance <= radius * 1.5) {
                        terrain = 2; // Desert around grass
                    } else {
                        terrain = 3; // Water around edge
                    }
                    
                    this.scene.setTile(actualQ, actualR, terrain, color);
                }
            }
        }
        
        if (this.onMapChangeCallback) {
            this.onMapChangeCallback();
        }
    }
    
    /**
     * Paint a unit at the specified coordinates
     */
    public paintUnit(q: number, r: number, unitType: number, playerId: number) {
        if (!this.scene) return;
        
        // For now, we'll use the scene's setTile method to represent units
        // In the future, this should be replaced with actual unit sprites
        this.scene.setUnit(q, r, unitType, playerId);
        
        if (this.onMapChangeCallback) {
            this.onMapChangeCallback();
        }
        
        console.log(`[PhaserMapEditor] Painted unit ${unitType} (player ${playerId}) at Q=${q}, R=${r}`);
    }
    
    /**
     * Remove a unit at the specified coordinates
     */
    public removeUnit(q: number, r: number) {
        if (!this.scene) return;
        
        // Remove unit from the scene
        this.scene.removeUnit(q, r);
        
        if (this.onMapChangeCallback) {
            this.onMapChangeCallback();
        }
        
        console.log(`[PhaserMapEditor] Removed unit at Q=${q}, R=${r}`);
    }
    
    // Cleanup
    public destroy() {
        if (this.game) {
            this.game.destroy(true);
            this.game = null;
        }
        
        this.scene = null;
        this.containerElement = null;
    }
    
    // Reference image methods (editor-only)
    
    /**
     * Load reference image from clipboard
     */
    public async loadReferenceFromClipboard(): Promise<boolean> {
        try {
            const scene = await this.waitForSceneReady();
            return scene.loadReferenceFromClipboard();
        } catch (error) {
            console.error('[PhaserMapEditor] Failed to load reference from clipboard:', error);
            return false;
        }
    }
    
    /**
     * Load reference image from file
     */
    public async loadReferenceFromFile(file: File): Promise<boolean> {
        console.log(`[PhaserMapEditor] loadReferenceFromFile called with: ${file.name}`);
        
        try {
            const scene = await this.waitForSceneReady();
            console.log(`[PhaserMapEditor] Scene ready, type: ${scene.constructor.name}`);
            
            // Check if the method exists
            if (typeof scene.loadReferenceFromFile !== 'function') {
                console.error('[PhaserMapEditor] loadReferenceFromFile method not found on scene');
                return false;
            }
            
            console.log(`[PhaserMapEditor] Method exists, calling scene.loadReferenceFromFile`);
            const result = await scene.loadReferenceFromFile(file);
            console.log(`[PhaserMapEditor] Scene loadReferenceFromFile returned: ${result}`);
            return result;
        } catch (error) {
            console.error('[PhaserMapEditor] Error in loadReferenceFromFile:', error);
            return false;
        }
    }
    
    /**
     * Set reference image mode
     */
    public setReferenceMode(mode: number): void {
        if (this.scene) {
            this.scene.setReferenceMode(mode);
        }
    }
    
    /**
     * Set reference image alpha
     */
    public setReferenceAlpha(alpha: number): void {
        if (this.scene) {
            this.scene.setReferenceAlpha(alpha);
        }
    }
    
    /**
     * Set reference image position
     */
    public setReferencePosition(x: number, y: number): void {
        if (this.scene) {
            this.scene.setReferencePosition(x, y);
        }
    }
    
    /**
     * Set reference image scale
     */
    public setReferenceScale(x: number, y: number): void {
        if (this.scene) {
            this.scene.setReferenceScale(x, y);
        }
    }
    
    /**
     * Set reference image scale with top-left corner as pivot
     */
    public setReferenceScaleFromTopLeft(x: number, y: number): void {
        if (this.scene) {
            this.scene.setReferenceScaleFromTopLeft(x, y);
        }
    }
    
    /**
     * Get reference image state
     */
    public getReferenceState(): {
        mode: number;
        alpha: number;
        position: { x: number; y: number };
        scale: { x: number; y: number };
        hasImage: boolean;
    } | null {
        if (this.scene) {
            return this.scene.getReferenceState();
        }
        return null;
    }
    
    /**
     * Clear reference image
     */
    public clearReferenceImage(): void {
        if (this.scene) {
            this.scene.clearReferenceImage();
        }
    }
}