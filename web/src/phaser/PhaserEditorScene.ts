import { PhaserWorldScene } from './PhaserWorldScene';
import { Unit, Tile, World } from '../World';
import { EventBus } from '../../lib/EventBus';
import { hexToPixel, pixelToHex } from './hexUtils';

/**
 * PhaserEditorScene extends PhaserWorldScene with editor-specific functionality.
 * 
 * This scene adds:
 * - Reference image support for map creation
 * - Terrain and unit painting tools
 * - Brush size and mode selection
 * - Editor-specific UI controls and shortcuts
 * - World modification capabilities
 * 
 * Inherits from PhaserWorldScene:
 * - World as single source of truth for game data
 * - Tile and unit rendering using World data
 * - Camera controls and theme management
 * - Asset loading and coordinate conversion
 * - Self-contained Phaser.Game instance
 */
export class PhaserEditorScene extends PhaserWorldScene {
    // Reference image functionality
    private referenceImage: Phaser.GameObjects.Image | null = null;
    private referenceMode: boolean = false;
    private referenceAlpha: number = 0.5;
    private referencePosition: { x: number; y: number } = { x: 0, y: 0 };

    // Editor-specific state
    private currentTerrain: number = 1; // Default grass (terrain type 1)
    private currentUnit: number = 1; // Default unit
    private currentPlayer: number = 0; // Default player
    private brushSize: number = 0; // Single tile
    private editorMode: 'terrain' | 'unit' | 'erase' = 'terrain';

    constructor(containerElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super(containerElement, eventBus, debugMode);
        // Override the scene key for this specific scene type
        // this.scene.settings.key = 'PhaserEditorScene';
    }

    /**
     * Set current terrain type for painting
     */
    public setCurrentTerrain(terrainType: number): void {
        this.currentTerrain = terrainType;
    }

    /**
     * Set current unit type for painting
     */
    public setCurrentUnit(unitType: number): void {
        this.currentUnit = unitType;
    }

    /**
     * Set current player for painting
     */
    public setCurrentPlayer(player: number): void {
        this.currentPlayer = player;
    }

    /**
     * Set brush size for terrain painting
     */
    public setBrushSize(size: number): void {
        this.brushSize = size;
    }

    /**
     * Set editor mode (terrain, unit, erase)
     */
    public setEditorMode(mode: 'terrain' | 'unit' | 'erase'): void {
        this.editorMode = mode;
    }

    /**
     * Reference image functionality
     */
    public setReferenceImage(imageUrl: string): void {
        if (this.referenceImage) {
            this.referenceImage.destroy();
        }

        // Load and display reference image
        this.load.image('reference', imageUrl);
        this.load.start();
        
        this.load.once('complete', () => {
            this.referenceImage = this.add.image(
                this.referencePosition.x, 
                this.referencePosition.y, 
                'reference'
            );
            this.referenceImage.setAlpha(this.referenceAlpha);
            this.referenceImage.setDepth(-1); // Behind tiles
        });
    }

    /**
     * Set reference image mode (0 = hidden, 1 = background, 2 = overlay)
     */
    public setReferenceMode(mode: number): void {
        this.referenceMode = mode > 0;
        if (this.referenceImage) {
            this.referenceImage.setVisible(this.referenceMode);
            if (mode === 1) {
                this.referenceImage.setDepth(-1); // Background
            } else if (mode === 2) {
                this.referenceImage.setDepth(1000); // Overlay
            }
        }
    }

    /**
     * Toggle reference image visibility
     */
    public toggleReferenceMode(): void {
        this.referenceMode = !this.referenceMode;
        if (this.referenceImage) {
            this.referenceImage.setVisible(this.referenceMode);
        }
    }

    /**
     * Set reference image alpha
     */
    public setReferenceAlpha(alpha: number): void {
        this.referenceAlpha = alpha;
        if (this.referenceImage) {
            this.referenceImage.setAlpha(alpha);
        }
    }

    /**
     * Set reference image position
     */
    public setReferencePosition(x: number, y: number): void {
        this.referencePosition = { x, y };
        if (this.referenceImage) {
            this.referenceImage.setPosition(x, y);
        }
    }

    /**
     * Set reference image scale
     */
    public setReferenceScale(x: number, y: number): void {
        if (this.referenceImage) {
            this.referenceImage.setScale(x, y);
            // Emit scale change event for WorldEditor
            this.events.emit('referenceScaleChanged', { x, y });
        }
    }

    /**
     * Set reference image scale with top-left corner as pivot
     */
    public setReferenceScaleFromTopLeft(x: number, y: number): void {
        if (this.referenceImage) {
            // Set origin to top-left for scaling
            this.referenceImage.setOrigin(0, 0);
            this.referenceImage.setScale(x, y);
            // Emit scale change event for WorldEditor
            this.events.emit('referenceScaleChanged', { x, y });
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
        if (!this.referenceImage) {
            return {
                mode: 0,
                alpha: this.referenceAlpha,
                position: { x: this.referencePosition.x, y: this.referencePosition.y },
                scale: { x: 1, y: 1 },
                hasImage: false
            };
        }

        return {
            mode: this.referenceMode ? 1 : 0,
            alpha: this.referenceAlpha,
            position: { 
                x: this.referenceImage.x, 
                y: this.referenceImage.y 
            },
            scale: { 
                x: this.referenceImage.scaleX, 
                y: this.referenceImage.scaleY 
            },
            hasImage: true
        };
    }

    /**
     * Clear reference image
     */
    public clearReferenceImage(): void {
        if (this.referenceImage) {
            this.referenceImage.destroy();
            this.referenceImage = null;
        }
        this.referenceMode = false;
    }

    /**
     * Get current editor state for UI synchronization
     */
    public getEditorState(): {
        terrain: number;
        unit: number;
        player: number;
        brushSize: number;
        mode: string;
        referenceMode: boolean;
        referenceAlpha: number;
    } {
        return {
            terrain: this.currentTerrain,
            unit: this.currentUnit,
            player: this.currentPlayer,
            brushSize: this.brushSize,
            mode: this.editorMode,
            referenceMode: this.referenceMode,
            referenceAlpha: this.referenceAlpha
        };
    }

    /**
     * Create a test pattern for debugging
     */
    public createTestPattern(): void {
        
        // Clear existing
        this.clearAllTiles();
        this.clearAllUnits();
        
        // Create test tiles
        const testTiles = [
            { q: 0, r: 0, tileType: 5, player: 0 },   // Grass
            { q: 1, r: 0, tileType: 6, player: 0 },   // Desert
            { q: -1, r: 0, tileType: 7, player: 0 },  // Water
            { q: 0, r: 1, tileType: 1, player: 1 },   // Base
            { q: 0, r: -1, tileType: 2, player: 2 },  // Factory
        ];

        testTiles.forEach(tile => this.setTile(tile));
        
        // Create test units
        const testUnits = [
            { q: 1, r: 1, unitType: 1, player: 1 },   // Infantry
            { q: -1, r: -1, unitType: 2, player: 2 }, // Tank
        ];

        testUnits.forEach(unit => this.setUnit(Unit.from(unit)));
    }

    // =============================================================================
    // Higher-level API methods (moved from PhaserWorldEditor)
    // =============================================================================

    // Event callbacks
    private onTileClickCallback: ((q: number, r: number) => void) | null = null;
    private onWorldChangeCallback: (() => void) | null = null;
    private onReferenceScaleChangeCallback: ((x: number, y: number) => void) | null = null;

    /**
     * Set terrain type for painting (compatibility with PhaserEditorComponent)
     */
    public async setTerrain(terrain: number): Promise<void> {
        this.setCurrentTerrain(terrain);
    }

    /**
     * Set color/player for painting (compatibility with PhaserEditorComponent)
     */
    public async setColor(color: number): Promise<void> {
        this.setCurrentPlayer(color);
    }

    /**
     * Set tiles data (load world data) - compatibility method
     */
    public async setTilesData(tiles: Array<Tile>): Promise<void> {
        // Wait for assets to be ready before placing tiles
        await this.waitForAssetsReady();
        
        this.clearAllTiles();
        
        tiles.forEach(tile => {
            this.setTile(tile, 0); // Use 0 brush size for data loading
        });
    }

    /**
     * Get tiles data - compatibility method
     */
    public getTilesData(): Array<Tile> {
        const tilesData: Array<Tile> = [];
        
        this.tileSprites.forEach((tile, key) => {
            const [q, r] = key.split(',').map(Number);
            // Extract terrain and color from texture key
            const textureKey = tile.texture.key;
            const match = textureKey.match(/terrain_(\d+)_(\d+)/);
            
            if (match) {
                tilesData.push({
                    q,
                    r,
                    tileType: parseInt(match[1]),
                    player: parseInt(match[2])
                });
            }
        });
        
        return tilesData;
    }

    /**
     * Set callback for tile click events
     */
    public onTileClick(callback: (q: number, r: number) => void): void {
        this.onTileClickCallback = callback;
    }

    /**
     * Set callback for world change events
     */
    public onWorldChange(callback: () => void): void {
        this.onWorldChangeCallback = callback;
    }

    /**
     * Set callback for reference scale change events
     */
    public onReferenceScaleChange(callback: (x: number, y: number) => void): void {
        this.onReferenceScaleChangeCallback = callback;
    }

    /**
     * Get the current viewport center in hex coordinates
     */
    public getViewportCenter(): { q: number, r: number } {
        if (!this.cameras?.main) return { q: 0, r: 0 };
        
        const camera = this.cameras.main;
        
        // Get the center of the viewport in world coordinates
        const centerX = camera.scrollX + (camera.width / 2) / camera.zoom;
        const centerY = camera.scrollY + (camera.height / 2) / camera.zoom;
        
        // Convert pixel coordinates to hex coordinates using hexUtils
        return pixelToHex(centerX, centerY);
    }

    /**
     * Center camera on hex coordinates
     */
    public centerCamera(q: number = 0, r: number = 0): void {
        if (!this.cameras?.main) return;
        
        // Convert hex coordinates to pixel coordinates using hexUtils
        const position = hexToPixel(q, r);
        
        // Center camera on position
        this.cameras.main.centerOn(position.x, position.y);
    }

    /**
     * Fill all terrain with specified type and color
     */
    public fillAllTerrain(terrain: number, color: number = 0): void {
        // This would need to iterate through all tiles in a reasonable area
        // For now, implement a simple pattern
        for (let q = -20; q <= 20; q++) {
            for (let r = -20; r <= 20; r++) {
                this.setTile({ q, r, tileType: terrain, player: color });
            }
        }
        
        if (this.onWorldChangeCallback) {
            this.onWorldChangeCallback();
        }
    }

    /**
     * Randomize terrain across the map
     */
    public randomizeTerrain(): void {
        const terrainTypes = [5, 6, 7, 8, 9]; // Common terrain types
        
        for (let q = -20; q <= 20; q++) {
            for (let r = -20; r <= 20; r++) {
                const randomTerrain = terrainTypes[Math.floor(Math.random() * terrainTypes.length)];
                this.setTile({ q, r, tileType: randomTerrain, player: 0 });
            }
        }
        
        if (this.onWorldChangeCallback) {
            this.onWorldChangeCallback();
        }
    }

    /**
     * Create island pattern centered at given coordinates
     */
    public createIslandPattern(centerQ: number = 0, centerR: number = 0, radius: number = 5): void {
        // Clear existing
        this.clearAllTiles();
        
        // Create island with water around edges, land in center
        for (let q = centerQ - radius; q <= centerQ + radius; q++) {
            for (let r = centerR - radius; r <= centerR + radius; r++) {
                const distance = Math.max(Math.abs(q - centerQ), Math.abs(r - centerR));
                
                if (distance <= radius) {
                    const terrainType = distance <= radius - 2 ? 5 : 7; // Grass or water
                    this.setTile({ q, r, tileType: terrainType, player: 0 });
                }
            }
        }
        
        if (this.onWorldChangeCallback) {
            this.onWorldChangeCallback();
        }
    }

    /**
     * Load reference image from file
     */
    public async loadReferenceFromFile(file: File): Promise<boolean> {
        if (!file.type.startsWith('image/')) {
            console.error('[PhaserEditorScene] File is not an image:', file.type);
            return false;
        }
        const imageUrl = URL.createObjectURL(file);
        this.setReferenceImage(imageUrl);
        return true;
    }

    /**
     * Load reference image from clipboard
     */
    public async loadReferenceFromClipboard(): Promise<boolean> {
        // Read from clipboard
        const items = await navigator.clipboard.read();
        
        for (const item of items) {
            if (item.types.includes('image/png') || item.types.includes('image/jpeg')) {
                const imageBlob = await item.getType(item.types.find(type => type.startsWith('image/')) || '');
                const imageUrl = URL.createObjectURL(imageBlob);
                
                this.setReferenceImage(imageUrl);
                return true;
            }
        }
        
        console.warn('[PhaserEditorScene] No image found in clipboard');
        return false;
    }

    /**
     * Handle tile click events (called internally)
     */
    protected handleTileClickInternal(q: number, r: number): void {
        if (this.onTileClickCallback) {
            this.onTileClickCallback(q, r);
        }
        
        // Note: onWorldChangeCallback will be called by specific paint methods when needed
    }

    /**
     * Set tile with brush size support - override parent method to match expected signature
     */
    public setTile(tile: Tile, brushSize: number = 0): void {
        // Call parent setTile method which expects just the tile
        super.setTile(tile);
    }
}
