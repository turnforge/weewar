import { PhaserWorldScene } from './PhaserWorldScene';
import { Unit, Tile, World } from '../World';
import * as models from '../../gen/wasmjs/weewar/v1/models'
import { EventBus } from '../../lib/EventBus';
import { TILE_WIDTH, hexToPixel, pixelToHex } from './hexUtils';
import { ReferenceImageLayer } from './layers/ReferenceImageLayer';
import { ShapeHighlightLayer } from './layers/HexHighlightLayer';

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
    // Reference image layer for proper overlay/background handling
    private referenceImageLayer: ReferenceImageLayer | null = null;

    // Shape preview layer for showing drag outline (rectangle, circle, ellipse, etc.)
    private shapePreviewLayer: ShapeHighlightLayer | null = null;

    // Editor-specific state
    private currentTerrain: number = 1; // Default grass (terrain type 1)
    private currentUnit: number = 1; // Default unit
    private currentPlayer: number = 0; // Default player
    private brushSize: number = 0; // Single tile
    private editorMode: 'terrain' | 'unit' | 'clear' = 'terrain';

    // Rectangle tool state
    private rectangleStartCoord: { q: number; r: number } | null = null;
    private isRectangleMode: boolean = false;

    constructor(containerElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super(containerElement, eventBus, debugMode);
        // Override the scene key for this specific scene type
        // this.scene.settings.key = 'PhaserEditorScene';
    }

    /**
     * Override create to add rectangle mode input handling after input system is ready
     */
    create() {
        super.create();
        // Setup rectangle mode input handling after parent's input setup
        this.setupRectangleInputHandling();
    }

    /**
     * Override setupLayerSystem to add ReferenceImageLayer and RectanglePreviewLayer for editor
     */
    protected setupLayerSystem(): void {
        // Call parent to set up base layers (baseMapLayer, exhaustedUnitsLayer)
        super.setupLayerSystem();

        // Create and add reference image layer for overlay/background support
        this.referenceImageLayer = new ReferenceImageLayer(this);
        this.layerManager.addLayer(this.referenceImageLayer);

        // Create and add shape preview layer for shape tools (rectangle, circle, ellipse, etc.)
        this.shapePreviewLayer = new ShapeHighlightLayer(this, TILE_WIDTH);
        this.layerManager.addLayer(this.shapePreviewLayer);
    }

    /**
     * Setup rectangle mode input handling
     */
    private setupRectangleInputHandling(): void {
        // Override the drag zone behavior to handle rectangle mode
        if (this.dragZone) {
            // Store the original drag handler
            const originalDragHandler = this.dragZone.listeners('drag')[0];

            // Remove the original handler
            this.dragZone.off('drag');

            // Add new handler that checks rectangle mode first
            this.dragZone.on('drag', (pointer: Phaser.Input.Pointer, dragX: number, dragY: number) => {
                // If in rectangle mode and we have a start coord, handle rectangle preview instead of camera drag
                if (this.isRectangleMode && this.rectangleStartCoord) {
                    // Convert pointer to hex coordinates
                    const worldPoint = this.cameras.main.getWorldPoint(pointer.x, pointer.y);
                    const hexCoord = pixelToHex(worldPoint.x, worldPoint.y);

                    // Show preview
                    this.showRectanglePreview(
                        this.rectangleStartCoord.q,
                        this.rectangleStartCoord.r,
                        hexCoord.q,
                        hexCoord.r
                    );
                    return; // Don't pan camera
                }

                // Otherwise, call original drag handler
                if (originalDragHandler) {
                    originalDragHandler.call(this, pointer, dragX, dragY);
                }
            });
        }

        // Handle pointer down for rectangle mode
        this.input.on('pointerdown', (pointer: Phaser.Input.Pointer) => {
            if (!this.isRectangleMode || pointer.button !== 0) return;

            // Convert pointer to hex coordinates
            const worldPoint = this.cameras.main.getWorldPoint(pointer.x, pointer.y);
            const hexCoord = pixelToHex(worldPoint.x, worldPoint.y);

            // Store start coordinate
            this.rectangleStartCoord = { q: hexCoord.q, r: hexCoord.r };
        });

        // Handle pointer up to apply rectangle
        this.input.on('pointerup', (pointer: Phaser.Input.Pointer) => {
            if (!this.isRectangleMode || !this.rectangleStartCoord || pointer.button !== 0) return;

            // Convert pointer to hex coordinates
            const worldPoint = this.cameras.main.getWorldPoint(pointer.x, pointer.y);
            const hexCoord = pixelToHex(worldPoint.x, worldPoint.y);

            // Apply rectangle
            this.applyRectangle(
                this.rectangleStartCoord.q,
                this.rectangleStartCoord.r,
                hexCoord.q,
                hexCoord.r
            );

            // Clear preview and reset
            this.clearRectanglePreview();
            this.rectangleStartCoord = null;
        });
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
     * Set editor mode (terrain, unit, clear)
     */
    public setEditorMode(mode: 'terrain' | 'unit' | 'clear'): void {
        this.editorMode = mode;
    }

    /**
     * Reference image functionality - delegated to ReferenceImageLayer
     */
    public setReferenceImage(imageUrl: string): void {
        if (this.referenceImageLayer) {
            this.referenceImageLayer.setReferenceImage(imageUrl);
        }
    }

    /**
     * Set reference image mode (0 = hidden, 1 = background, 2 = overlay)
     */
    public setReferenceMode(mode: number): void {
        if (this.referenceImageLayer) {
            this.referenceImageLayer.setReferenceMode(mode);
        }
    }

    /**
     * Toggle reference image visibility
     */
    public toggleReferenceMode(): void {
        if (this.referenceImageLayer) {
            const currentState = this.referenceImageLayer.getReferenceState();
            const newMode = currentState.mode === 0 ? 1 : 0;
            this.referenceImageLayer.setReferenceMode(newMode);
        }
    }

    /**
     * Set reference image alpha
     */
    public setReferenceAlpha(alpha: number): void {
        if (this.referenceImageLayer) {
            this.referenceImageLayer.setReferenceAlpha(alpha);
        }
    }

    /**
     * Set reference image position
     */
    public setReferencePosition(x: number, y: number): void {
        if (this.referenceImageLayer) {
            this.referenceImageLayer.setReferencePosition(x, y);
        }
    }

    /**
     * Set reference image scale
     */
    public setReferenceScale(x: number, y: number): void {
        if (this.referenceImageLayer) {
            this.referenceImageLayer.setReferenceScale(x, y);
            // Emit scale change event for WorldEditor
            this.events.emit('referenceScaleChanged', { x, y });
        }
    }

    /**
     * Set reference image scale with top-left corner as pivot
     */
    public setReferenceScaleFromTopLeft(x: number, y: number): void {
        if (this.referenceImageLayer) {
            this.referenceImageLayer.setReferenceScaleFromTopLeft(x, y);
            // Emit scale change event for WorldEditor
            const state = this.referenceImageLayer.getReferenceState();
            this.events.emit('referenceScaleChanged', { x: state.scale.x, y: state.scale.y });
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
        if (this.referenceImageLayer) {
            return this.referenceImageLayer.getReferenceState();
        }
        return null;
    }

    /**
     * Clear reference image
     */
    public clearReferenceImage(): void {
        if (this.referenceImageLayer) {
            this.referenceImageLayer.clearReferenceImage();
        }
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
        const refState = this.referenceImageLayer?.getReferenceState();
        return {
            terrain: this.currentTerrain,
            unit: this.currentUnit,
            player: this.currentPlayer,
            brushSize: this.brushSize,
            mode: this.editorMode,
            referenceMode: refState ? refState.mode > 0 : false,
            referenceAlpha: refState ? refState.alpha : 0.5
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
            { q: 0, r: 0, tileType: 5, player: 0, shortcut: "", lastActedTurn: 0, lastToppedupTurn: 0 },   // Grass
            { q: 1, r: 0, tileType: 6, player: 0, shortcut: "", lastActedTurn: 0, lastToppedupTurn: 0 },   // Desert
            { q: -1, r: 0, tileType: 7, player: 0, shortcut: "", lastActedTurn: 0, lastToppedupTurn: 0 },  // Water
            { q: 0, r: 1, tileType: 1, player: 1, shortcut: "", lastActedTurn: 0, lastToppedupTurn: 0 },   // Base
            { q: 0, r: -1, tileType: 2, player: 2, shortcut: "", lastActedTurn: 0, lastToppedupTurn: 0 },  // Factory
        ];

        testTiles.forEach(tile => this.setTile(tile));
        
        // Create test units
        const testUnits = [
            { q: 1, r: 1, unitType: 1, player: 1 },   // Infantry
            { q: -1, r: -1, unitType: 2, player: 2 }, // Tank
        ];

        testUnits.forEach(unit => this.setUnit(models.Unit.from(unit)));
    }

    // =============================================================================
    // Higher-level API methods (moved from PhaserWorldEditor)
    // =============================================================================

    // Event callbacks
    private onTileClickCallback: ((q: number, r: number) => void) | null = null;
    private onWorldChangeCallback: (() => void) | null = null;
    private onReferenceScaleChangeCallback: ((x: number, y: number) => void) | null = null;
    private onReferencePositionChangeCallback: ((x: number, y: number) => void) | null = null;

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

        // Subscribe to scene events and forward to callback
        this.events.on('referenceScaleChanged', (data: { x: number; y: number }) => {
            if (this.onReferenceScaleChangeCallback) {
                this.onReferenceScaleChangeCallback(data.x, data.y);
            }
        });
    }

    /**
     * Set callback for reference position change events
     */
    public onReferencePositionChange(callback: (x: number, y: number) => void): void {
        this.onReferencePositionChangeCallback = callback;

        // Subscribe to scene events and forward to callback
        this.events.on('referencePositionChanged', (data: { x: number; y: number }) => {
            if (this.onReferencePositionChangeCallback) {
                this.onReferencePositionChangeCallback(data.x, data.y);
            }
        });
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
                this.setTile({ q, r, tileType: terrain, player: color, shortcut: "", lastActedTurn: 0, lastToppedupTurn: 0 });
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
                this.setTile({ q, r, tileType: randomTerrain, player: 0, shortcut: "", lastActedTurn: 0, lastToppedupTurn: 0 });
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
                    this.setTile({ q, r, tileType: terrainType, player: 0, shortcut: "", lastActedTurn: 0, lastToppedupTurn: 0 });
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
        if (this.referenceImageLayer) {
            return await this.referenceImageLayer.loadReferenceFromFile(file);
        }
        return false;
    }

    /**
     * Load reference image from clipboard
     */
    public async loadReferenceFromClipboard(): Promise<boolean> {
        if (this.referenceImageLayer) {
            return await this.referenceImageLayer.loadReferenceFromClipboard();
        }
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

    /**
     * Enable rectangle mode
     */
    public setRectangleMode(enabled: boolean): void {
        this.isRectangleMode = enabled;
        if (!enabled) {
            this.clearRectanglePreview();
            this.rectangleStartCoord = null;
        }
    }

    /**
     * Check if rectangle mode is enabled
     */
    public isInRectangleMode(): boolean {
        return this.isRectangleMode;
    }

    /**
     * Show rectangle preview outline
     */
    public showRectanglePreview(startQ: number, startR: number, endQ: number, endR: number): void {
        if (!this.shapePreviewLayer || !this.world) return;

        // Use World's rectFrom method to get outline tiles (filled=false)
        const outlineTiles = this.world.rectFrom(startQ, startR, endQ, endR, false)
            .map(([q, r]) => ({ q, r }));

        this.shapePreviewLayer.showShapeOutline(outlineTiles);
    }

    /**
     * Clear rectangle preview
     */
    public clearRectanglePreview(): void {
        if (this.shapePreviewLayer) {
            this.shapePreviewLayer.clearPreview();
        }
    }

    /**
     * Apply terrain to all tiles in rectangle
     */
    public applyRectangle(startQ: number, startR: number, endQ: number, endR: number): void {
        if (!this.world) return;

        // Use World's rectFrom method to get all tiles (filled=true)
        const tiles = this.world.rectFrom(startQ, startR, endQ, endR, true);

        // Apply terrain to all tiles in the rectangle
        for (const [q, r] of tiles) {
            if (this.editorMode === 'terrain') {
                this.world.setTileAt(q, r, this.currentTerrain, 0);
            } else if (this.editorMode === 'unit') {
                this.world.setUnitAt(q, r, this.currentUnit, this.currentPlayer);
            } else if (this.editorMode === 'clear') {
                this.world.removeUnitAt(q, r);
            }
        }
    }
}
