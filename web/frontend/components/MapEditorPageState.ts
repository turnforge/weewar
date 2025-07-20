/**
 * MapEditorPageState - Centralized state management for MapEditorPage
 * 
 * Provides single source of truth for all editor UI state that needs to be
 * shared between components within the MapEditorPage. Uses Observer pattern
 * for efficient component synchronization.
 */

// Observer pattern interfaces
export interface PageStateObserver {
    onPageStateEvent(event: PageStateEvent): void;
}

export interface PageStateEvent {
    type: PageStateEventType;
    data: any;
}

export enum PageStateEventType {
    TOOL_STATE_CHANGED = 'tool-state-changed',
    VISUAL_STATE_CHANGED = 'visual-state-changed', 
    WORKFLOW_STATE_CHANGED = 'workflow-state-changed'
}

// Page state data interfaces
export interface ToolState {
    selectedTerrain: number;
    selectedUnit: number;
    selectedPlayer: number;
    placementMode: 'terrain' | 'unit' | 'clear';
    brushSize: number;
}

export interface VisualState {
    showGrid: boolean;
    showCoordinates: boolean;
    // Note: theme is application-level, not page-level
}

export interface WorkflowState {
    hasPendingMapDataLoad: boolean;
    pendingGridState: boolean | null;
    lastAction: string;
}

// Event data interfaces for type safety
export interface ToolStateChangedEventData {
    previousState: ToolState;
    newState: ToolState;
    changedFields: (keyof ToolState)[];
}

export interface VisualStateChangedEventData {
    previousState: VisualState;
    newState: VisualState;
    changedFields: (keyof VisualState)[];
}

export interface WorkflowStateChangedEventData {
    previousState: WorkflowState;
    newState: WorkflowState;
    changedFields: (keyof WorkflowState)[];
}

// State persistence interface for save/restore operations
export interface SavedUIState {
    terrain: number;
    unit: number;
    playerId: number;
    brushSize: number;
    placementMode: 'terrain' | 'unit' | 'clear';
}

/**
 * MapEditorPageState - Centralized state management for map editor page
 * 
 * Features:
 * - Observer pattern for real-time component synchronization
 * - Type-safe state updates with granular change tracking
 * - State persistence for save/restore operations (keyboard shortcuts)
 * - Batched state updates for performance
 * - Comprehensive state validation
 */
export class MapEditorPageState {
    // Core state data
    private toolState: ToolState;
    private visualState: VisualState;
    private workflowState: WorkflowState;
    
    // Observer pattern
    private observers: Set<PageStateObserver> = new Set();
    
    // State persistence for undo/restore
    private savedUIState: SavedUIState | null = null;
    
    constructor() {
        // Initialize with sensible defaults
        this.toolState = {
            selectedTerrain: 1, // Default to grass
            selectedUnit: 0,    // Default to no unit
            selectedPlayer: 1,  // Default to player 1
            placementMode: 'terrain',
            brushSize: 0        // Default to single hex
        };
        
        this.visualState = {
            showGrid: false,
            showCoordinates: false
        };
        
        this.workflowState = {
            hasPendingMapDataLoad: false,
            pendingGridState: null,
            lastAction: 'initialized'
        };
    }
    
    // Observer pattern methods
    public subscribe(observer: PageStateObserver): void {
        this.observers.add(observer);
    }
    
    public unsubscribe(observer: PageStateObserver): void {
        this.observers.delete(observer);
    }
    
    private emit(event: PageStateEvent): void {
        this.observers.forEach(observer => {
            try {
                observer.onPageStateEvent(event);
            } catch (error) {
                console.error('Error in page state observer:', error);
            }
        });
    }
    
    // Tool state management
    public getToolState(): ToolState {
        return { ...this.toolState };
    }
    
    public updateToolState(updates: Partial<ToolState>): void {
        const previousState = { ...this.toolState };
        const changedFields: (keyof ToolState)[] = [];
        
        // Apply updates and track changes
        for (const [key, value] of Object.entries(updates) as [keyof ToolState, any][]) {
            if (this.toolState[key] !== value) {
                changedFields.push(key);
                (this.toolState as any)[key] = value;
            }
        }
        
        // Only emit if there were actual changes
        if (changedFields.length > 0) {
            this.workflowState.lastAction = `tool-update-${changedFields.join(',')}`;
            
            this.emit({
                type: PageStateEventType.TOOL_STATE_CHANGED,
                data: {
                    previousState,
                    newState: { ...this.toolState },
                    changedFields
                } as ToolStateChangedEventData
            });
        }
    }
    
    // Individual tool state setters for convenience
    public setSelectedTerrain(terrain: number): void {
        this.updateToolState({ 
            selectedTerrain: terrain,
            placementMode: terrain === 0 ? 'clear' : 'terrain'
        });
    }
    
    public setSelectedUnit(unit: number): void {
        this.updateToolState({ 
            selectedUnit: unit,
            placementMode: 'unit'
        });
    }
    
    public setSelectedPlayer(player: number): void {
        this.updateToolState({ selectedPlayer: player });
    }
    
    public setPlacementMode(mode: 'terrain' | 'unit' | 'clear'): void {
        this.updateToolState({ placementMode: mode });
    }
    
    public setBrushSize(size: number): void {
        this.updateToolState({ brushSize: size });
    }
    
    // Visual state management
    public getVisualState(): VisualState {
        return { ...this.visualState };
    }
    
    public updateVisualState(updates: Partial<VisualState>): void {
        const previousState = { ...this.visualState };
        const changedFields: (keyof VisualState)[] = [];
        
        // Apply updates and track changes
        for (const [key, value] of Object.entries(updates) as [keyof VisualState, any][]) {
            if (this.visualState[key] !== value) {
                changedFields.push(key);
                (this.visualState as any)[key] = value;
            }
        }
        
        // Only emit if there were actual changes
        if (changedFields.length > 0) {
            this.workflowState.lastAction = `visual-update-${changedFields.join(',')}`;
            
            this.emit({
                type: PageStateEventType.VISUAL_STATE_CHANGED,
                data: {
                    previousState,
                    newState: { ...this.visualState },
                    changedFields
                } as VisualStateChangedEventData
            });
        }
    }
    
    public setShowGrid(show: boolean): void {
        this.updateVisualState({ showGrid: show });
    }
    
    public setShowCoordinates(show: boolean): void {
        this.updateVisualState({ showCoordinates: show });
    }
    
    // Workflow state management
    public getWorkflowState(): WorkflowState {
        return { ...this.workflowState };
    }
    
    public updateWorkflowState(updates: Partial<WorkflowState>): void {
        const previousState = { ...this.workflowState };
        const changedFields: (keyof WorkflowState)[] = [];
        
        // Apply updates and track changes
        for (const [key, value] of Object.entries(updates) as [keyof WorkflowState, any][]) {
            if (this.workflowState[key] !== value) {
                changedFields.push(key);
                (this.workflowState as any)[key] = value;
            }
        }
        
        // Only emit if there were actual changes
        if (changedFields.length > 0) {
            this.emit({
                type: PageStateEventType.WORKFLOW_STATE_CHANGED,
                data: {
                    previousState,
                    newState: { ...this.workflowState },
                    changedFields
                } as WorkflowStateChangedEventData
            });
        }
    }
    
    public setHasPendingMapDataLoad(pending: boolean): void {
        this.updateWorkflowState({ hasPendingMapDataLoad: pending });
    }
    
    public setPendingGridState(state: boolean | null): void {
        this.updateWorkflowState({ pendingGridState: state });
    }
    
    // State persistence for keyboard shortcuts and undo operations
    public saveUIState(): void {
        this.savedUIState = {
            terrain: this.toolState.selectedTerrain,
            unit: this.toolState.selectedUnit,
            playerId: this.toolState.selectedPlayer,
            brushSize: this.toolState.brushSize,
            placementMode: this.toolState.placementMode
        };
        this.updateWorkflowState({ lastAction: 'ui-state-saved' });
    }
    
    public restoreUIState(): boolean {
        if (!this.savedUIState) {
            return false;
        }
        
        this.updateToolState({
            selectedTerrain: this.savedUIState.terrain,
            selectedUnit: this.savedUIState.unit,
            selectedPlayer: this.savedUIState.playerId,
            brushSize: this.savedUIState.brushSize,
            placementMode: this.savedUIState.placementMode
        });
        
        this.updateWorkflowState({ lastAction: 'ui-state-restored' });
        return true;
    }
    
    public hasSavedUIState(): boolean {
        return this.savedUIState !== null;
    }
    
    // Reset methods for keyboard shortcuts
    public resetToDefaults(): void {
        this.updateToolState({
            selectedTerrain: 1,
            selectedUnit: 0,
            selectedPlayer: 1,
            placementMode: 'terrain',
            brushSize: 0
        });
        this.updateWorkflowState({ lastAction: 'reset-to-defaults' });
    }
    
    // Utility methods
    public getCurrentTerrain(): number {
        return this.toolState.selectedTerrain;
    }
    
    public getCurrentUnit(): number {
        return this.toolState.selectedUnit;
    }
    
    public getCurrentPlayer(): number {
        return this.toolState.selectedPlayer;
    }
    
    public getCurrentPlacementMode(): 'terrain' | 'unit' | 'clear' {
        return this.toolState.placementMode;
    }
    
    public getCurrentBrushSize(): number {
        return this.toolState.brushSize;
    }
    
    public getShowGrid(): boolean {
        return this.visualState.showGrid;
    }
    
    public getShowCoordinates(): boolean {
        return this.visualState.showCoordinates;
    }
    
    // Serialization for debugging and persistence
    public serialize(): {
        toolState: ToolState;
        visualState: VisualState;
        workflowState: WorkflowState;
        hasSavedState: boolean;
    } {
        return {
            toolState: { ...this.toolState },
            visualState: { ...this.visualState },
            workflowState: { ...this.workflowState },
            hasSavedState: this.savedUIState !== null
        };
    }
    
    // State validation
    public validate(): { isValid: boolean; errors: string[] } {
        const errors: string[] = [];
        
        // Validate tool state
        if (this.toolState.selectedTerrain < 0 || this.toolState.selectedTerrain > 26) {
            errors.push(`Invalid terrain: ${this.toolState.selectedTerrain}`);
        }
        
        if (this.toolState.selectedUnit < 0 || this.toolState.selectedUnit > 20) {
            errors.push(`Invalid unit: ${this.toolState.selectedUnit}`);
        }
        
        if (this.toolState.selectedPlayer < 1 || this.toolState.selectedPlayer > 4) {
            errors.push(`Invalid player: ${this.toolState.selectedPlayer}`);
        }
        
        if (this.toolState.brushSize < 0 || this.toolState.brushSize > 15) {
            errors.push(`Invalid brush size: ${this.toolState.brushSize}`);
        }
        
        const validModes = ['terrain', 'unit', 'clear'];
        if (!validModes.includes(this.toolState.placementMode)) {
            errors.push(`Invalid placement mode: ${this.toolState.placementMode}`);
        }
        
        return {
            isValid: errors.length === 0,
            errors
        };
    }
    
    // Debug and monitoring
    public getLastAction(): string {
        return this.workflowState.lastAction;
    }
    
    public getObserverCount(): number {
        return this.observers.size;
    }
}