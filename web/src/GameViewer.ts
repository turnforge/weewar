import { EventBus } from '../lib/EventBus';
import { WorldViewer } from './WorldViewer';
import { PhaserGameScene } from './phaser/PhaserGameScene';
import { SelectionHighlightLayer, MovementHighlightLayer, AttackHighlightLayer } from './phaser/layers/HexHighlightLayer';

/**
 * GameViewer Component - Specialized WorldViewer for interactive gameplay
 * 
 * Extends WorldViewer<PhaserGameScene> to provide:
 * - Direct access to game-specific highlight layers
 * - Type-safe interaction with PhaserGameScene
 * - Game-specific helper methods
 */
export class GameViewer extends WorldViewer<PhaserGameScene> {

    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super(rootElement, eventBus, debugMode);
    }

    /**
     * Factory method override to create PhaserGameScene
     */
    protected createScene(): PhaserGameScene {
        return new PhaserGameScene();
    }

    /**
     * Override to handle GameViewer-specific initialization after Phaser scene is ready
     */
    protected async initializePhaserScene(): Promise<void> {
        // Call parent implementation first
        await super.initializePhaserScene();
        
        // GameViewer-specific setup can go here if needed
        console.log('GameViewer: Phaser scene initialized and ready for game interactions');
        
        // Emit a GameViewer-specific ready event for the parent page
        this.emit('game-viewer-ready', {
            componentId: this.componentId,
            success: true
        }, this, this);
    }

    /**
     * Get game-specific highlight layers with proper typing
     */
    public getSelectionHighlightLayer(): SelectionHighlightLayer | null {
        return this.scene?.getSelectionHighlightLayer() || null;
    }

    public getMovementHighlightLayer(): MovementHighlightLayer | null {
        return this.scene?.getMovementHighlightLayer() || null;
    }

    public getAttackHighlightLayer(): AttackHighlightLayer | null {
        return this.scene?.getAttackHighlightLayer() || null;
    }

    /**
     * Game-specific helper methods
     */
    public selectUnit(q: number, r: number, movableCoords: Array<{ q: number; r: number }>, attackableCoords: Array<{ q: number; r: number }>): void {
        if (this.scene) {
            this.scene.selectUnit(q, r, movableCoords, attackableCoords);
        }
    }

    public clearSelection(): void {
        if (this.scene) {
            this.scene.clearSelection();
        }
    }

    public setGameMode(mode: 'select' | 'move' | 'attack'): void {
        if (this.scene) {
            this.scene.setGameMode(mode);
        }
    }

    public getSelectedUnit(): { q: number; r: number; unitData: any } | null {
        return this.scene?.getSelectedUnit() || null;
    }
}
