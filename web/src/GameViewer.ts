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
     * Phase 3: Activate component - Subscribe to World events
     */
    async activate(): Promise<void> {
        // Call parent activation first
        await super.activate();
        
        // ✅ Subscribe to World's coordinated update events
        this.addSubscription('world-updated', null);
    }

    /**
     * Handle events from the EventBus (World-coordinated events)
     */
    public handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
        switch(eventType) {
            case 'world-updated':
                console.log('GameViewer: Received world-updated event', data);
                this.handleWorldUpdated(data);
                break;
            
            default:
                // Call parent implementation for unhandled events
                super.handleBusEvent(eventType, data, target, emitter);
        }
    }

    /**
     * Handle world-updated events from World component
     * World has already updated its data, now update visual sprites
     */
    private handleWorldUpdated(data: { changes: any[], world: any }): void {
        const { changes } = data;
        
        // Process each change to update visual elements
        for (const change of changes) {
            if (change.unitMoved) {
                this.handleUnitMoved({
                    from: { q: change.unitMoved.previousUnit?.q || 0, r: change.unitMoved.previousUnit?.r || 0 },
                    to: { q: change.unitMoved.updatedUnit?.q || 0, r: change.unitMoved.updatedUnit?.r || 0 }
                });
            }
            
            if (change.unitDamaged) {
                this.handleUnitDamaged({
                    position: { q: change.unitDamaged.updatedUnit?.q || 0, r: change.unitDamaged.updatedUnit?.r || 0 },
                    previousHealth: change.unitDamaged.previousUnit?.availableHealth || 0,
                    newHealth: change.unitDamaged.updatedUnit?.availableHealth || 0
                });
            }
            
            if (change.unitKilled) {
                this.handleUnitKilled({
                    position: { q: change.unitKilled.previousUnit?.q || 0, r: change.unitKilled.previousUnit?.r || 0 },
                    player: change.unitKilled.previousUnit?.player || 0,
                    unitType: change.unitKilled.previousUnit?.unitType || 0
                });
            }
        }
    }

    /**
     * Factory method override to create PhaserGameScene
     */
    protected createScene(): PhaserGameScene {
        return new PhaserGameScene();
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
     * Set movement callback for game-specific move execution
     */
    public setMovementCallback(movementCallback?: (q: number, r: number, moveOption: any) => void): void {
        console.log('[GameViewer] setMovementCallback called');
        
        if (this.scene && 'setCallbacks' in this.scene) {
            // Update the PhaserGameScene callbacks to include movement callback
            const currentCallbacks = (this.scene as any).callbacks || {};
            (this.scene as any).setCallbacks({
                ...currentCallbacks,
                onMovementClicked: movementCallback
            });
            console.log('[GameViewer] Movement callback set on PhaserGameScene');
        } else {
            console.error('[GameViewer] No scene available or scene does not support game callbacks');
        }
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

    /**
     * Get the Phaser scene for advanced operations
     */
    public getScene(): PhaserGameScene | null {
        return this.scene;
    }

    /**
     * Handle world-changed events from GameState
     */
    private handleWorldChanged(data: { changes: any[], world: any }): void {
        console.log('GameViewer: Handling world-changed event with', data.changes.length, 'changes');
        
        // Reload the world data into the Phaser scene
        if (this.scene && data.world) {
            // This will refresh the entire scene with the updated world data
            this.scene.loadWorldData(data.world);
        }
    }

    /**
     * Handle unit-moved events from GameState
     */
    private handleUnitMoved(data: { from: { q: number, r: number }, to: { q: number, r: number } }): void {
        console.log('GameViewer: Handling unit-moved event from', data.from, 'to', data.to);
        
        // Update the unit position in the Phaser scene by removing and re-adding
        if (this.scene) {
            // Remove unit from old position
            this.scene.removeUnit(data.from.q, data.from.r);
            
            // Get the unit from the updated world data and add it at new position
            const world = this.scene.world;
            if (world) {
                const unit = world.getUnitAt(data.to.q, data.to.r);
                if (unit) {
                    this.scene.setUnit(unit);
                }
            }
        }
    }

    /**
     * Handle unit-damaged events from GameState
     */
    private handleUnitDamaged(data: { position: { q: number, r: number }, previousHealth: number, newHealth: number }): void {
        console.log('GameViewer: Handling unit-damaged event at', data.position);
        
        // Update unit health display or trigger damage animation
        if (this.scene) {
            // This would need to be implemented in PhaserGameScene
            // this.scene.updateUnitHealth(data.position.q, data.position.r, data.newHealth);
        }
    }

    /**
     * Handle unit-killed events from GameState
     */
    private handleUnitKilled(data: { position: { q: number, r: number }, player: number, unitType: string }): void {
        console.log('GameViewer: Handling unit-killed event at', data.position);
        
        // Remove unit from Phaser scene
        if (this.scene) {
            this.scene.removeUnit(data.position.q, data.position.r);
        }
    }
}
