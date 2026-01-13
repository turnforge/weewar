import * as Phaser from 'phaser';
import { PhaserWorldScene } from '../common/PhaserWorldScene';
import { hexToPixel } from '../common/hexUtils';
import { World } from '../common/World';
import { SelectionHighlightLayer, MovementHighlightLayer, AttackHighlightLayer, CaptureHighlightLayer } from '../common/HexHighlightLayer';
import { EventBus } from '@panyam/tsappkit';
import { GameViewPresenterClient as  GameViewPresenterClient } from '../../gen/wasmjs/lilbattle/v1/services/gameViewPresenterClient';
import { MoveUnitAction, AttackUnitAction, HighlightSpec } from '../../gen/wasmjs/lilbattle/v1/models/interfaces';

/**
 * PhaserGameScene extends PhaserWorldScene with game-specific interactive features.
 * 
 * This scene adds:
 * - Unit selection and visual highlighting
 * - Movement range display with pathfinding
 * - Attack range visualization
 * - Click-to-move and click-to-attack interactions
 * - Visual feedback for player actions
 * - Turn-based interaction states
 * 
 * Inherits from PhaserWorldScene:
 * - World as single source of truth for game data
 * - Tile and unit rendering using World data
 * - Camera controls and theme management
 * - Asset loading and coordinate conversion
 * - Self-contained Phaser.Game instance
 * - Callback system for external communication
 */
export class PhaserGameScene extends PhaserWorldScene {
    // Game-specific state
    public gameViewPresenterClient: GameViewPresenterClient;
    private selectedUnit: { q: number; r: number; unitData: any } | null = null;
    private gameMode: 'select' | 'move' | 'attack' = 'select';
    
    // Layer-based highlight system
    private _selectionHighlightLayer: SelectionHighlightLayer | null = null;
    private _movementHighlightLayer: MovementHighlightLayer | null = null;
    private _attackHighlightLayer: AttackHighlightLayer | null = null;
    private _captureHighlightLayer: CaptureHighlightLayer | null = null;
    
    // Path preview graphics for movement/attack visualization
    private pathPreview: Phaser.GameObjects.Graphics | null = null;
    
    // Interaction data from game engine
    private movableCoords: Array<{ q: number; r: number }> = [];
    private attackableCoords: Array<{ q: number; r: number }> = [];
    
    constructor(containerElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super(containerElement, eventBus, debugMode);
        // Override the scene key for this specific scene type
        // this.scene.settings.key = 'PhaserGameScene';
    }

    /**
     * Override create to add game-specific initialization
     */
    create() {
        // Call parent create first
        super.create();
        
        // Set up game-specific layers
        this.setupGameLayers();
    }

    public get selectionHighlightLayer(): SelectionHighlightLayer { return this._selectionHighlightLayer! }
    public get movementHighlightLayer(): MovementHighlightLayer { return this._movementHighlightLayer! }
    public get attackHighlightLayer(): AttackHighlightLayer { return this._attackHighlightLayer! }
    
    /**
     * Set up game-specific highlight layers
     */
    private setupGameLayers(): void {
        const layerManager = this.getLayerManager();
        if (!layerManager) {
            console.error('[PhaserGameScene] No LayerManager available');
            return;
        }
        
        // Create selection highlight layer
        this._selectionHighlightLayer = new SelectionHighlightLayer(this, this.tileWidth);
        layerManager.addLayer(this._selectionHighlightLayer);
        
        // Create movement highlight layer
        this._movementHighlightLayer = new MovementHighlightLayer(this, this.tileWidth);
        layerManager.addLayer(this._movementHighlightLayer);
        
        // Create attack highlight layer
        this._attackHighlightLayer = new AttackHighlightLayer(this, this.tileWidth);
        layerManager.addLayer(this._attackHighlightLayer);

        // Create capture highlight layer
        this._captureHighlightLayer = new CaptureHighlightLayer(this, this.tileWidth);
        layerManager.addLayer(this._captureHighlightLayer);
    }

    /**
     * Set the current game mode
     */
    public setGameMode(mode: 'select' | 'move' | 'attack'): void {
        this.gameMode = mode;
        
        // Update visual indicators based on mode
        this.updateModeVisuals();
    }

    /**
     * Clear unit selection and all highlights
     */
    public clearSelection(): void {
        this.selectedUnit = null;
        this.movableCoords = [];
        this.attackableCoords = [];
        
        // Clear all visual highlights using layer system
        if (this._selectionHighlightLayer) {
            this._selectionHighlightLayer.clearSelection();
        }
        if (this._movementHighlightLayer) {
            this._movementHighlightLayer.clearMovementOptions();
        }
        if (this._attackHighlightLayer) {
            this._attackHighlightLayer.clearAttackOptions();
        }
        
        // Disable distance labels
        this.setShowUnitDistance(false);
        
        // Return to select mode
        this.setGameMode('select');
    }

    /**
     * Update visuals based on current mode
     */
    private updateModeVisuals(): void {
        // Clear previous mode-specific highlights
        this.clearPathPreview();
        
        switch (this.gameMode) {
            case 'move':
                // Emphasize movement options
                this.emphasizeMovementHighlights();
                break;
            case 'attack':
                // Emphasize attack options
                this.emphasizeAttackHighlights();
                break;
            case 'select':
                // Show both movement and attack options equally
                this.normalizeHighlights();
                break;
        }
    }

    /**
     * Highlight the selected unit
     */
    private highlightSelectedUnit(q: number, r: number): void {
        if (this._selectionHighlightLayer) {
            this._selectionHighlightLayer.selectHex(q, r);
        } else {
            console.warn('[PhaserGameScene] Selection highlight layer not available');
        }
    }

    /**
     * Emphasize movement highlights for move mode
     */
    private emphasizeMovementHighlights(): void {
        if (this._movementHighlightLayer) {
            this._movementHighlightLayer.setAlpha(1.0); // Full opacity
        }
        if (this._attackHighlightLayer) {
            this._attackHighlightLayer.setAlpha(0.3); // Reduced opacity
        }
    }

    /**
     * Emphasize attack highlights for attack mode
     */
    private emphasizeAttackHighlights(): void {
        if (this._attackHighlightLayer) {
            this._attackHighlightLayer.setAlpha(1.0); // Full opacity
        }
        if (this._movementHighlightLayer) {
            this._movementHighlightLayer.setAlpha(0.3); // Reduced opacity
        }
    }

    /**
     * Normalize highlights for select mode
     */
    private normalizeHighlights(): void {
        if (this._movementHighlightLayer) {
            this._movementHighlightLayer.setAlpha(0.7); // Normal opacity
        }
        if (this._attackHighlightLayer) {
            this._attackHighlightLayer.setAlpha(0.7); // Normal opacity
        }
    }

    /**
     * Clear path preview
     */
    private clearPathPreview(): void {
        if (this.pathPreview) {
            this.pathPreview.destroy();
            this.pathPreview = null;
        }
    }

    /**
     * Get current game mode
     */
    public getGameMode(): 'select' | 'move' | 'attack' {
        return this.gameMode;
    }

    // =========================================================================
    // Visualization Command Methods (called by presenter via GameViewerPage)
    // =========================================================================

    /**
     * Show highlights on the game board
     * @param highlights Array of HighlightSpec from presenter
     */
    public showHighlights(highlights: HighlightSpec[]): void {
        if (!highlights || highlights.length === 0) {
            return;
        }

        // Group highlights by type in single pass (instead of 6 separate filter operations)
        const selections: HighlightSpec[] = [];
        const movements: MoveUnitAction[] = [];
        const attacks: AttackUnitAction[] = [];
        const captures: HighlightSpec[] = [];
        const exhausted: HighlightSpec[] = [];
        const capturing: HighlightSpec[] = [];

        for (const h of highlights) {
            switch (h.type) {
                case 'selection': selections.push(h); break;
                case 'movement': if (h.move) movements.push(h.move); break;
                case 'attack': if (h.attack) attacks.push(h.attack); break;
                case 'capture': captures.push(h); break;
                case 'exhausted': exhausted.push(h); break;
                case 'capturing': capturing.push(h); break;
            }
        }

        // Apply selection highlights (typically just one)
        if (this._selectionHighlightLayer && selections.length > 0) {
            selections.forEach(h => {
                this._selectionHighlightLayer!.selectHex(h.q, h.r);
            });
        }

        // Apply movement highlights
        if (this._movementHighlightLayer && movements.length > 0) {
            // Convert to MoveOption-like objects for the layer
            this._movementHighlightLayer.showMovementOptions(movements);
        }

        // Apply attack highlights
        if (this._attackHighlightLayer && attacks.length > 0) {
            this._attackHighlightLayer.showAttackOptions(attacks);
        }

        // Apply capture highlights (interactive - for clicking to execute capture)
        if (this._captureHighlightLayer && captures.length > 0) {
            captures.forEach(h => {
                this._captureHighlightLayer!.showCaptureOption(h.q, h.r);
            });
        }

        // Apply exhausted highlights
        const exhaustedLayer = this.getExhaustedUnitsLayer();
        if (exhaustedLayer && exhausted.length > 0) {
            exhausted.forEach(h => {
                exhaustedLayer.markExhausted(h.q, h.r);
            });
        }

        // Apply capturing flag highlights
        const capturingFlagLayer = this.getCapturingFlagLayer();
        if (capturingFlagLayer && capturing.length > 0) {
            capturing.forEach(h => {
                capturingFlagLayer.showFlag(h.q, h.r, h.player);
            });
        }
    }

    /**
     * Clear highlights from the game board
     * @param types Array of highlight types to clear, empty = clear all
     */
    public clearHighlights(types: string[]): void {
        const clearAll = !types || types.length === 0;

        if ((clearAll || types.includes('selection')) && this._selectionHighlightLayer) {
            this._selectionHighlightLayer.clearSelection();
        }

        if ((clearAll || types.includes('movement')) && this._movementHighlightLayer) {
            this._movementHighlightLayer.clearMovementOptions();
        }

        if ((clearAll || types.includes('attack')) && this._attackHighlightLayer) {
            this._attackHighlightLayer.clearAttackOptions();
        }

        if ((clearAll || types.includes('capture')) && this._captureHighlightLayer) {
            this._captureHighlightLayer.clearCaptureOptions();
        }

        const exhaustedLayer = this.getExhaustedUnitsLayer();
        if ((clearAll || types.includes('exhausted')) && exhaustedLayer) {
            exhaustedLayer.clearAllExhausted();
        }

        const capturingFlagLayer = this.getCapturingFlagLayer();
        if ((clearAll || types.includes('capturing')) && capturingFlagLayer) {
            capturingFlagLayer.clearAllFlags();
        }
    }

    /**
     * Show a path on the game board
     * @param coords Flat array of coordinates [q1, r1, q2, r2, ...]
     * @param color Hex color (e.g., 0x00FF00)
     * @param thickness Line thickness
     */
    public showPath(coords: number[], color: number, thickness: number): void {
        if (!coords || coords.length < 4) {
            return;
        }

        // Use movement layer for path drawing
        if (this._movementHighlightLayer) {
            this._movementHighlightLayer.addPath(coords, color, thickness);
        }
    }

    /**
     * Clear all paths from the game board
     */
    public clearPaths(): void {
        if (this._movementHighlightLayer) {
            this._movementHighlightLayer.clearAllPaths();
        }
    }

    /**
     * Override handleTap to call presenter directly instead of callback
     */
    protected handleTap(pointer: Phaser.Input.Pointer): void {
        // Use layer system for hit testing
        if (this.layerManager && this.gameViewPresenterClient) {
            const clickContext = this.layerManager.getClickContext(pointer);
            if (clickContext) {
                // Call presenter directly
                this.gameViewPresenterClient.sceneClicked({
                    gameId: "", // Will be filled by presenter
                    pos: {q: clickContext.hexQ, r: clickContext.hexR, label: ""},
                    layer: clickContext.layer || 'unknown',
                });
            }
        }
    }

    /**
     * Manual cleanup when scene is destroyed
     */
    public destroy(): void {
        this.clearPathPreview();
        super.destroy();
    }
}
