import { LCMComponent } from '../lib/LCMComponent';
import { EventBus } from '../lib/EventBus';

export interface GameActionsCallbacks {
    onEndTurn: () => void;
    onShowAllUnits: () => void;
    onCenterOnAction: () => void;
    onMoveUnit: () => void;
    onAttackUnit: () => void;
    onUndo: () => void;
}

/**
 * GameActionsPanel - Manages game action buttons and selected unit info
 * 
 * Features:
 * - Selected unit information display
 * - Move and attack action buttons
 * - Turn management (end turn, undo)
 * - Unit selection utilities (show all units, center on action)
 * - Game status display
 */
export class GameActionsPanel implements LCMComponent {
    private element: HTMLElement;
    private eventBus: EventBus;
    private callbacks: GameActionsCallbacks;

    constructor(element: HTMLElement, eventBus: EventBus, callbacks: GameActionsCallbacks) {
        this.element = element;
        this.eventBus = eventBus;
        this.callbacks = callbacks;
    }

    performLocalInit(): LCMComponent[] {
        this.initializeEventHandlers();
        return [];
    }

    setupDependencies(): void {
        // No dependencies needed
    }

    async activate(): Promise<void> {
        // Component is ready to use
    }

    handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
        // Handle any events that affect game actions
    }

    deactivate(): void {
        // Clean up if needed
    }

    /**
     * Show selected unit information
     */
    public showSelectedUnit(unit: any): void {
        const unitInfoPanel = this.element.querySelector('#selected-unit-info');
        const unitDetails = this.element.querySelector('#unit-details');
        
        if (unitInfoPanel && unitDetails) {
            unitDetails.innerHTML = `
                <div><strong>Type:</strong> ${unit.unitType || 'Unknown'}</div>
                <div><strong>Position:</strong> (${unit.q}, ${unit.r})</div>
                <div><strong>Player:</strong> ${unit.player || 'Unknown'}</div>
                <div><strong>Health:</strong> ${unit.availableHealth || 'Unknown'}</div>
                <div><strong>Distance Left:</strong> ${unit.distanceLeft || 'Unknown'}</div>
            `;
            unitInfoPanel.classList.remove('hidden');
        }
    }

    /**
     * Hide selected unit information
     */
    public hideSelectedUnit(): void {
        const unitInfoPanel = this.element.querySelector('#selected-unit-info');
        if (unitInfoPanel) {
            unitInfoPanel.classList.add('hidden');
        }
    }

    /**
     * Update game status information
     */
    public updateGameStatus(currentPlayer: number, turnCounter: number): void {
        const currentPlayerEl = this.element.querySelector('#current-player-display');
        const turnDisplayEl = this.element.querySelector('#turn-display');
        
        if (currentPlayerEl) {
            currentPlayerEl.textContent = `Player ${currentPlayer}`;
        }
        
        if (turnDisplayEl) {
            turnDisplayEl.textContent = turnCounter.toString();
        }
    }

    /**
     * Enable or disable the undo button
     */
    public setUndoEnabled(enabled: boolean): void {
        const undoBtn = this.element.querySelector('#undo-action-btn') as HTMLButtonElement;
        if (undoBtn) {
            undoBtn.disabled = !enabled;
            if (enabled) {
                undoBtn.className = 'w-full inline-flex items-center justify-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-yellow-600 hover:bg-yellow-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-yellow-500';
            } else {
                undoBtn.className = 'w-full inline-flex items-center justify-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-gray-400 bg-gray-200 dark:bg-gray-700 dark:text-gray-500 cursor-not-allowed';
            }
        }
    }

    private initializeEventHandlers(): void {
        // Move unit button
        const moveUnitBtn = this.element.querySelector('#move-unit-btn');
        if (moveUnitBtn) {
            moveUnitBtn.addEventListener('click', () => this.callbacks.onMoveUnit());
        }

        // Attack unit button  
        const attackUnitBtn = this.element.querySelector('#attack-unit-btn');
        if (attackUnitBtn) {
            attackUnitBtn.addEventListener('click', () => this.callbacks.onAttackUnit());
        }

        // Show all units button
        const selectAllUnitsBtn = this.element.querySelector('#select-all-units-btn');
        if (selectAllUnitsBtn) {
            selectAllUnitsBtn.addEventListener('click', () => this.callbacks.onShowAllUnits());
        }

        // Center on action button
        const centerOnActionBtn = this.element.querySelector('#center-on-action-btn');
        if (centerOnActionBtn) {
            centerOnActionBtn.addEventListener('click', () => this.callbacks.onCenterOnAction());
        }

        // End turn button
        const endTurnBtn = this.element.querySelector('#end-turn-action-btn');
        if (endTurnBtn) {
            endTurnBtn.addEventListener('click', () => this.callbacks.onEndTurn());
        }
        
        // Undo button
        const undoBtn = this.element.querySelector('#undo-action-btn');
        if (undoBtn) {
            undoBtn.addEventListener('click', () => this.callbacks.onUndo());
        }
    }
}
