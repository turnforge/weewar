import { PhaserWorldScene } from './PhaserWorldScene';
import { Unit, Tile, World } from '../World';

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

    constructor(config?: string | Phaser.Types.Scenes.SettingsConfig) {
        super(config || { key: 'PhaserEditorScene' });
    }

    /**
     * Override create to set up editor-specific layer callbacks
     */
    create() {
        // Call parent create first to set up layer system
        super.create();
        
        // Note: BaseMapLayer callbacks will be set by PhaserEditorComponent
        // The component will handle the painting logic and world updates
        
        console.log('[PhaserEditorScene] Editor scene created');
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
            console.log(`[PhaserEditorScene] Reference image loaded: ${imageUrl}`);
        });
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
                    console.log('[PhaserEditorScene] Reference image loaded from clipboard');
                    return true;
                }
            }
            
            console.warn('[PhaserEditorScene] No image found in clipboard');
            return false;
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
            console.log(`[PhaserEditorScene] Reference image loaded from file: ${file.name}`);
            return true;
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
        console.log(`[PhaserEditorScene] Reference mode set to: ${mode}`);
    }

    /**
     * Toggle reference image visibility
     */
    public toggleReferenceMode(): void {
        this.referenceMode = !this.referenceMode;
        if (this.referenceImage) {
            this.referenceImage.setVisible(this.referenceMode);
        }
        console.log(`[PhaserEditorScene] Reference mode: ${this.referenceMode}`);
    }

    /**
     * Set reference image alpha
     */
    public setReferenceAlpha(alpha: number): void {
        this.referenceAlpha = alpha;
        if (this.referenceImage) {
            this.referenceImage.setAlpha(alpha);
        }
        console.log(`[PhaserEditorScene] Reference alpha set to: ${alpha}`);
    }

    /**
     * Set reference image position
     */
    public setReferencePosition(x: number, y: number): void {
        this.referencePosition = { x, y };
        if (this.referenceImage) {
            this.referenceImage.setPosition(x, y);
        }
        console.log(`[PhaserEditorScene] Reference position set to: (${x}, ${y})`);
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
        console.log(`[PhaserEditorScene] Reference scale set to: (${x}, ${y})`);
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
        console.log(`[PhaserEditorScene] Reference scale from top-left set to: (${x}, ${y})`);
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
        console.log(`[PhaserEditorScene] Reference image cleared`);
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
        console.log('[PhaserEditorScene] Creating test pattern...');
        
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
        
        console.log('[PhaserEditorScene] Test pattern created');
    }
}
