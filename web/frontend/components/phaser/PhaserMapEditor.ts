import * as Phaser from 'phaser';
import { PhaserMapScene } from './PhaserMapScene';

export class PhaserMapEditor {
    private game: Phaser.Game | null = null;
    private scene: PhaserMapScene | null = null;
    private containerElement: HTMLElement | null = null;
    
    private currentTerrain: number = 1;
    private currentColor: number = 0;
    private brushSize: number = 0;
    
    // Event callbacks
    private onTileClickCallback: ((q: number, r: number) => void) | null = null;
    private onMapChangeCallback: (() => void) | null = null;
    
    constructor(containerId: string) {
        this.containerElement = document.getElementById(containerId);
        
        if (!this.containerElement) {
            throw new Error(`Container element with ID '${containerId}' not found`);
        }
        
        this.initialize();
    }
    
    private initialize() {
        const config: Phaser.Types.Core.GameConfig = {
            type: Phaser.AUTO,
            parent: this.containerElement!,
            width: '100%',
            height: '100%',
            backgroundColor: '#2c3e50',
            scene: PhaserMapScene,
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
            this.scene = this.game!.scene.getScene('PhaserMapScene') as PhaserMapScene;
            
            // Set up event listeners
            this.setupEventListeners();
            
            console.log('[PhaserMapEditor] Phaser game initialized successfully');
        });
    }
    
    private setupEventListeners() {
        if (!this.scene) return;
        
        // Listen for tile click events from the scene
        this.scene.events.on('tileClicked', (data: { q: number; r: number }) => {
            this.handleTileClick(data.q, data.r);
        });
    }
    
    private handleTileClick(q: number, r: number) {
        // Paint tile with current terrain and brush settings
        this.paintTile(q, r, this.currentTerrain, this.currentColor, this.brushSize);
        
        // Call external callback if set
        if (this.onTileClickCallback) {
            this.onTileClickCallback(q, r);
        }
        
        // Notify of map change
        if (this.onMapChangeCallback) {
            this.onMapChangeCallback();
        }
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
    
    public setTilesData(tiles: Array<{ q: number; r: number; terrain: number; color: number }>) {
        if (!this.scene) return;
        
        this.scene.clearAllTiles();
        
        tiles.forEach(tile => {
            this.scene!.setTile(tile.q, tile.r, tile.terrain, tile.color);
        });
    }
    
    // Event callbacks
    public onTileClick(callback: (q: number, r: number) => void) {
        this.onTileClickCallback = callback;
    }
    
    public onMapChange(callback: () => void) {
        this.onMapChangeCallback = callback;
    }
    
    // Camera controls
    public centerCamera(q: number = 0, r: number = 0) {
        if (!this.scene) return;
        
        // Convert hex coordinates to pixel coordinates
        const position = this.hexToPixel(q, r);
        
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
    
    // Helper method for hex to pixel conversion (matches scene implementation)
    private hexToPixel(q: number, r: number): { x: number; y: number } {
        const tileWidth = 64;
        const yIncrement = 48;
        
        // Match the Go implementation from map.go CenterXYForTile
        const x_coord = q;
        const z_coord = r;
        const col = x_coord + Math.floor((z_coord - (z_coord & 1)) / 2);
        const row = z_coord;
        
        let y = yIncrement * row;
        let x = tileWidth * col;
        
        if (row % 2 === 1) {
            x += tileWidth / 2;
        }
        
        return { x, y };
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
    
    // Cleanup
    public destroy() {
        if (this.game) {
            this.game.destroy(true);
            this.game = null;
        }
        
        this.scene = null;
        this.containerElement = null;
    }
}