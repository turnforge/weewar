import * as Phaser from 'phaser';
import { PhaserWorldScene } from './PhaserWorldScene';
import { hexToPixel } from './hexUtils';
import { World } from '../World';
import { SelectionHighlightLayer, MovementHighlightLayer, AttackHighlightLayer } from './layers/HexHighlightLayer';
import { EventBus } from '../../lib/EventBus';

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
    private selectedUnit: { q: number; r: number; unitData: any } | null = null;
    private gameMode: 'select' | 'move' | 'attack' = 'select';
    private currentPlayer: number = 1;
    private isPlayerTurn: boolean = true;
    
    // Layer-based highlight system
    private _selectionHighlightLayer: SelectionHighlightLayer | null = null;
    private _movementHighlightLayer: MovementHighlightLayer | null = null;
    private _attackHighlightLayer: AttackHighlightLayer | null = null;
    
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
    }
    
    /**
     * Get tile data at specific coordinates (for callback functions)
     */
    public getTileAt(q: number, r: number): any {
        if (!this.world) {
            return null;
        }
        return this.world.getTileAt(q, r);
    }

    /**
     * Get unit data at specific coordinates (for callback functions) 
     */
    public getUnitAt(q: number, r: number): any {
        if (!this.world) {
            return null;
        }
        return this.world.getUnitAt(q, r);
    }

    /**
     * Check if there's a unit at the specified coordinates
     */
    public hasUnitAt(q: number, r: number): boolean {
        if (!this.world) {
            return false;
        }
        return this.world.getUnitAt(q, r) !== null;
    }

    /**
     * Check if there's a tile at the specified coordinates
     */
    public hasTileAt(q: number, r: number): boolean {
        if (!this.world) {
            return false;
        }
        return this.world.getTileAt(q, r) !== null;
    }

    /**
     * Handle selection mode clicks
     */
    private handleSelectClick(q: number, r: number): void {
        // Check if there's a unit at this position
        const unitAtPosition = this.getUnitAt(q, r);
        
        if (unitAtPosition) {
            // Clear previous selection
            this.clearSelection();
            
            // Set new selection (visual will be updated by external callback)
            this.selectedUnit = {
                q: q,
                r: r,
                unitData: unitAtPosition
            };
            
            // Highlight selected unit
            this.highlightSelectedUnit(q, r);
        } else {
            // Empty tile click in select mode - clear selection
            this.clearSelection();
        }
    }

    /**
     * Handle movement mode clicks
     */
    private handleMoveClick(q: number, r: number): void {
        if (!this.selectedUnit) {
            console.warn('[PhaserGameScene] Move click but no unit selected');
            return;
        }

        // Check if this is a valid movement target
        const isValidMove = this.movableCoords.some(coord => coord.q === q && coord.r === r);
        
        if (isValidMove) {
            // External callback will handle the actual move through game engine
            // We just provide visual feedback here
            this.showMovePreview(this.selectedUnit.q, this.selectedUnit.r, q, r);
            
            // Return to select mode after move attempt
            this.setGameMode('select');
        } else {
            console.warn(`[PhaserGameScene] Invalid move target: Q=${q}, R=${r}`);
            // Could show visual feedback for invalid move
        }
    }

    /**
     * Handle attack mode clicks
     */
    private handleAttackClick(q: number, r: number): void {
        if (!this.selectedUnit) {
            console.warn('[PhaserGameScene] Attack click but no unit selected');
            return;
        }

        // Check if this is a valid attack target
        const isValidAttack = this.attackableCoords.some(coord => coord.q === q && coord.r === r);
        
        if (isValidAttack) {
            // External callback will handle the actual attack through game engine
            // We just provide visual feedback here
            this.showAttackPreview(this.selectedUnit.q, this.selectedUnit.r, q, r);
            
            // Return to select mode after attack attempt
            this.setGameMode('select');
        } else {
            console.warn(`[PhaserGameScene] Invalid attack target: Q=${q}, R=${r}`);
        }
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
     * Set the selected unit and show movement/attack options
     */
    public selectUnit(q: number, r: number, movableCoords: Array<{ q: number; r: number }>, attackableCoords: Array<{ q: number; r: number }>): void {
        const unitData = this.getUnitAt(q, r);
        if (!unitData) {
            console.warn(`[PhaserGameScene] No unit found at Q=${q}, R=${r} for selection`);
            return;
        }

        // Store selection state
        this.selectedUnit = { q, r, unitData };
        this.movableCoords = movableCoords;
        this.attackableCoords = attackableCoords;

        // Update visuals
        this.highlightSelectedUnit(q, r);
        // Convert coordinate objects to MoveOption-like objects for compatibility
        const moveOptions = movableCoords.map(coord => ({ q: coord.q, r: coord.r, movementCost: 1 }));
        this.showMovementOptions(moveOptions);
        this.showAttackOptions(attackableCoords);
        
        // Enable distance labels for the selected unit
        this.setShowUnitDistance(true, { q, r });
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
     * Show movement options as green highlights using MoveOption objects
     */
    private showMovementOptions(moveOptions: any[]): void {
        if (this._movementHighlightLayer) {
            this._movementHighlightLayer.showMovementOptions(moveOptions);
        } else {
            console.warn('[PhaserGameScene] Movement highlight layer not available');
        }
    }

    /**
     * Show attack options as red highlights
     */
    private showAttackOptions(attackableCoords: Array<{ q: number; r: number }>): void {
        if (this._attackHighlightLayer) {
            this._attackHighlightLayer.showAttackOptions(attackableCoords);
        } else {
            console.warn('[PhaserGameScene] Attack highlight layer not available');
        }
    }

    /**
     * Show move preview line
     */
    private showMovePreview(fromQ: number, fromR: number, toQ: number, toR: number): void {
        this.clearPathPreview();

        this.pathPreview = this.add.graphics();
        this.pathPreview.lineStyle(3, 0x00FF00, 0.8); // Green line

        const fromPos = hexToPixel(fromQ, fromR);
        const toPos = hexToPixel(toQ, toR);

        this.pathPreview.beginPath();
        this.pathPreview.moveTo(fromPos.x, fromPos.y);
        this.pathPreview.lineTo(toPos.x, toPos.y);
        this.pathPreview.strokePath();
        
        this.pathPreview.setDepth(15); // Above everything
    }

    /**
     * Show attack preview line
     */
    private showAttackPreview(fromQ: number, fromR: number, toQ: number, toR: number): void {
        this.clearPathPreview();

        this.pathPreview = this.add.graphics();
        this.pathPreview.lineStyle(3, 0xFF0000, 0.8); // Red line

        const fromPos = hexToPixel(fromQ, fromR);
        const toPos = hexToPixel(toQ, toR);

        this.pathPreview.beginPath();
        this.pathPreview.moveTo(fromPos.x, fromPos.y);
        this.pathPreview.lineTo(toPos.x, toPos.y);
        this.pathPreview.strokePath();
        
        this.pathPreview.setDepth(15); // Above everything
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
     * Set current player (affects interaction permissions)
     */
    public setCurrentPlayer(playerId: number): void {
        this.currentPlayer = playerId;
    }

    /**
     * Set whether it's the player's turn (affects interaction availability)
     */
    public setPlayerTurn(isPlayerTurn: boolean): void {
        this.isPlayerTurn = isPlayerTurn;
        
        if (!isPlayerTurn) {
            // Clear selection when it's not player's turn
            this.clearSelection();
        }
    }

    /**
     * Get current selection state for external queries
     */
    public getSelectedUnit(): { q: number; r: number; unitData: any } | null {
        return this.selectedUnit;
    }

    /**
     * Get current game mode
     */
    public getGameMode(): 'select' | 'move' | 'attack' {
        return this.gameMode;
    }

    /**
     * Get layer instances for direct manipulation
     */

    /**
     * Manual cleanup when scene is destroyed
     */
    public destroy(): void {
        this.clearPathPreview();
        super.destroy();
    }
}
