import * as models from '../../gen/wasmjs/lilbattle/v1/models/models'
import { Lilbattle_v1Deserializer as WD } from '../../gen/wasmjs/lilbattle/v1/factory'
import { EventBus } from '@panyam/tsappkit';
import { PhaserWorldScene } from '../common/PhaserWorldScene';
import { Unit, Tile, World } from '../common/World';
import { ShapeHighlightLayer } from '../common/HexHighlightLayer';
import { TILE_WIDTH, hexToPixel, pixelToHex, hexToRowCol } from '../common/hexUtils';
import { ReferenceImageLayer } from './ReferenceImageLayer';
import { ShapeTool } from './tools/ShapeTool';
import { RectangleTool } from './tools/RectangleTool';
import { CircleTool } from './tools/CircleTool';
import { OvalTool } from './tools/OvalTool';
import { LineTool } from './tools/LineTool';

/**
 * Information about the tile/unit under the mouse cursor
 */
export interface HoverInfo {
    q: number;
    r: number;
    row: number;
    col: number;
    tileType?: number;
    tilePlayer?: number;
    unitType?: number;
    unitPlayer?: number;
}

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

    // Shape tool state (multi-click system)
    private currentShapeTool: ShapeTool | null = null;
    private isInShapeDrawingMode: boolean = false;
    private shapeFillMode: boolean = true; // Fill vs outline mode

    constructor(containerElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super(containerElement, eventBus, debugMode);
        // Override the scene key for this specific scene type
        // this.scene.settings.key = 'PhaserEditorScene';
    }

    /**
     * Override create to add shape tool input handling after input system is ready
     */
    create() {
        super.create();
        // Setup shape tool input handling after parent's input setup
        this.setupShapeToolInputHandling();
        // Setup hover tracking for status bar
        this.setupHoverTracking();
    }

    /**
     * Setup hover tracking for status bar display
     */
    private setupHoverTracking(): void {
        this.input.on('pointermove', (pointer: Phaser.Input.Pointer) => {
            if (!this.onHoverCallback) return;

            // Convert pointer to world coordinates then to hex
            const worldPoint = this.cameras.main.getWorldPoint(pointer.x, pointer.y);
            const hexCoord = pixelToHex(worldPoint.x, worldPoint.y);
            const { row, col } = hexToRowCol(hexCoord.q, hexCoord.r);

            // Look up tile and unit at this position
            const tile = this.world?.getTileAt(hexCoord.q, hexCoord.r);
            const unit = this.world?.getUnitAt(hexCoord.q, hexCoord.r);

            const info: HoverInfo = {
                q: hexCoord.q,
                r: hexCoord.r,
                row,
                col,
            };

            if (tile) {
                info.tileType = tile.tileType;
                if (tile.player !== undefined && tile.player !== 0) {
                    info.tilePlayer = tile.player;
                }
            }

            if (unit) {
                info.unitType = unit.unitType;
                if (unit.player !== undefined) {
                    info.unitPlayer = unit.player;
                }
            }

            this.onHoverCallback(info);
        });

        // Clear hover info when pointer leaves the canvas
        this.input.on('pointerout', () => {
            if (this.onHoverCallback) {
                this.onHoverCallback(null);
            }
        });
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
     * Setup shape tool input handling (multi-click system)
     */
    private setupShapeToolInputHandling(): void {
        // Handle clicks to collect shape points
        this.input.on('pointerdown', (pointer: Phaser.Input.Pointer) => {
            if (!this.isInShapeDrawingMode || !this.currentShapeTool || pointer.button !== 0) return;

            // Convert pointer to hex coordinates
            const worldPoint = this.cameras.main.getWorldPoint(pointer.x, pointer.y);
            const hexCoord = pixelToHex(worldPoint.x, worldPoint.y);

            // Add point to the shape tool
            const needsMorePoints = this.currentShapeTool.addPoint(hexCoord.q, hexCoord.r);

            if (!needsMorePoints) {
                // Shape is complete - apply it
                this.applyCurrentShape();

                // Clear preview and reset tool for next shape
                this.clearShapePreview();
                this.currentShapeTool.reset();
            }
        });

        // Handle mouse movement to show live preview
        this.input.on('pointermove', (pointer: Phaser.Input.Pointer) => {
            if (!this.isInShapeDrawingMode || !this.currentShapeTool) return;

            // Convert pointer to hex coordinates
            const worldPoint = this.cameras.main.getWorldPoint(pointer.x, pointer.y);
            const hexCoord = pixelToHex(worldPoint.x, worldPoint.y);

            // Get preview tiles from the tool
            const previewTiles = this.currentShapeTool.getPreviewTiles(hexCoord.q, hexCoord.r);

            // Show preview
            if (this.shapePreviewLayer && previewTiles.length > 0) {
                this.shapePreviewLayer.showShapeOutline(previewTiles);
            }
        });

        // Handle keyboard input
        this.input.keyboard?.on('keydown', (event: KeyboardEvent) => {
            if (!this.isInShapeDrawingMode || !this.currentShapeTool) return;

            if (event.key === 'Escape') {
                // Cancel current shape but stay in shape mode
                this.cancelCurrentShape();
            } else if (event.key === 'Enter' && this.currentShapeTool.requiresKeyboardConfirm()) {
                // Complete shape if it requires keyboard confirmation (polygon, path, etc.)
                if (this.currentShapeTool.canComplete()) {
                    this.applyCurrentShape();
                    this.clearShapePreview();
                    this.currentShapeTool.reset();
                }
            }
        });

        // Disable camera panning when in shape drawing mode
        // We'll handle this in the drag handler by checking isInShapeDrawingMode
        if (this.dragZone) {
            // Store the original drag handler
            const originalDragHandler = this.dragZone.listeners('drag')[0];

            // Remove the original handler
            this.dragZone.off('drag');

            // Add new handler that disables pan during shape mode
            this.dragZone.on('drag', (pointer: Phaser.Input.Pointer, dragX: number, dragY: number) => {
                // Disable camera pan when in shape drawing mode
                if (this.isInShapeDrawingMode) {
                    return; // Don't pan camera
                }

                // Otherwise, call original drag handler
                if (originalDragHandler) {
                    originalDragHandler.call(this, pointer, dragX, dragY);
                }
            });
        }
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
     * Get reference image layer (for dependency injection)
     */
    public getReferenceImageLayer(): ReferenceImageLayer | null {
        return this.referenceImageLayer;
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

        testUnits.forEach(unit => this.setUnit(WD.from(models.Unit, unit)));
    }

    // =============================================================================
    // Higher-level API methods (moved from PhaserWorldEditor)
    // =============================================================================

    // Event callbacks
    private onTileClickCallback: ((q: number, r: number) => void) | null = null;
    private onWorldChangeCallback: (() => void) | null = null;
    private onReferenceScaleChangeCallback: ((x: number, y: number) => void) | null = null;
    private onReferencePositionChangeCallback: ((x: number, y: number) => void) | null = null;
    private onHoverCallback: ((info: HoverInfo | null) => void) | null = null;

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
     * Set callback for mouse hover events over the map
     */
    public onHover(callback: (info: HoverInfo | null) => void): void {
        this.onHoverCallback = callback;
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

    // Note: Reference image loading methods removed - now handled by ReferenceImagePanel

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
     * Set shape mode - creates and activates the appropriate shape tool
     * @param shapeType Type of shape: 'rectangle', 'circle', 'oval', 'line', or null to disable
     */
    public setShapeMode(shapeType: 'rectangle' | 'circle' | 'oval' | 'line' | null): void {
        if (shapeType === null) {
            this.exitShapeMode();
            return;
        }

        if (!this.world) return;

        // Create the appropriate shape tool
        switch (shapeType) {
            case 'rectangle':
                this.currentShapeTool = new RectangleTool(this.world, this.shapeFillMode);
                break;
            case 'circle':
                this.currentShapeTool = new CircleTool(this.world, this.shapeFillMode);
                break;
            case 'oval':
                this.currentShapeTool = new OvalTool(this.world, this.shapeFillMode);
                break;
            case 'line':
                this.currentShapeTool = new LineTool(this.world);
                break;
            default:
                console.warn(`Unknown shape type: ${shapeType}`);
                return;
        }

        this.isInShapeDrawingMode = true;
    }

    /**
     * Check if rectangle mode is enabled (for backward compatibility)
     */
    public isInRectangleMode(): boolean {
        return this.isInShapeDrawingMode && this.currentShapeTool instanceof RectangleTool;
    }

    /**
     * Set fill mode for shapes (filled vs outline)
     */
    public setShapeFillMode(filled: boolean): void {
        this.shapeFillMode = filled;
        if (this.currentShapeTool) {
            this.currentShapeTool.setFilled(filled);
        }
    }

    /**
     * Get current shape fill mode
     */
    public getShapeFillMode(): boolean {
        return this.shapeFillMode;
    }

    /**
     * Exit shape drawing mode (cancel current shape)
     * Escape behavior: Cancel the current shape being drawn but stay in rectangle mode
     */
    private cancelCurrentShape(): void {
        // Reset the current shape but keep the tool active
        if (this.currentShapeTool) {
            this.currentShapeTool.reset();
        }
        this.clearShapePreview();
    }

    /**
     * Exit shape drawing mode completely (disables tool)
     */
    private exitShapeMode(): void {
        this.isInShapeDrawingMode = false;
        if (this.currentShapeTool) {
            this.currentShapeTool.reset();
            this.currentShapeTool = null;
        }
        this.clearShapePreview();
    }

    /**
     * Clear shape preview
     */
    private clearShapePreview(): void {
        if (this.shapePreviewLayer) {
            this.shapePreviewLayer.clearPreview();
        }
    }

    /**
     * Apply current shape to the world
     */
    private applyCurrentShape(): void {
        if (!this.currentShapeTool || !this.world) return;

        // Get result tiles from the shape tool
        const tiles = this.currentShapeTool.getResultTiles();

        // Apply terrain/units to all tiles in the shape
        for (const { q, r } of tiles) {
            if (this.editorMode === 'terrain') {
                this.world.setTileAt(q, r, this.currentTerrain, 0);
            } else if (this.editorMode === 'unit') {
                this.world.setUnitAt(q, r, this.currentUnit, this.currentPlayer);
            } else if (this.editorMode === 'clear') {
                this.world.removeUnitAt(q, r);
            }
        }
    }

    /**
     * Override handleTap to prevent tile painting during shape mode
     */
    protected override handleTap(pointer: Phaser.Input.Pointer): void {
        // Block tile painting when in shape mode
        if (this.isInShapeDrawingMode) {
            return; // Don't call parent handleTap
        }

        // Otherwise, call parent to handle normal tile painting
        super.handleTap(pointer);
    }
}
