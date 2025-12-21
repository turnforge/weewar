/**
 * WorldEditorPresenter - Central orchestrator for WorldEditorPage
 *
 * Merges PageState functionality with orchestration responsibilities.
 * Components call presenter methods directly, presenter updates components directly.
 *
 * Responsibilities:
 * 1. Owns all UI state (tool, visual, workflow) - merged from PageState
 * 2. Orchestrates component interactions via direct method calls
 * 3. Handles World data operations
 * 4. Subscribes to World events (TILES_CHANGED, etc.) for coordination
 */

import { EventBus, EventSubscriber } from '@panyam/tsappkit';
import { World, TilesChangedEventData, UnitsChangedEventData, WorldLoadedEventData, CrossingType } from '../common/World';
import { WorldEventTypes } from '../common/events';
import { PhaserEditorComponent } from './PhaserEditorComponent';
import { EditorToolsPanel } from './ToolsPanel';
import { WorldStatsPanel } from '../common/WorldStatsPanel';
import { ReferenceImagePanel } from './ReferenceImagePanel';

// =========================================================================
// State Interfaces
// =========================================================================

export interface ToolState {
    selectedTerrain: number;
    selectedUnit: number;
    selectedPlayer: number;
    selectedCrossing: 'road' | 'bridge' | null;
    /** Preset connectsTo pattern for crossing placement (6 booleans for hex directions) */
    crossingConnectsTo: boolean[];
    placementMode: 'terrain' | 'unit' | 'crossing' | 'clear';
    brushMode: string;
    brushSize: number;
}

export interface VisualState {
    showGrid: boolean;
    showCoordinates: boolean;
    showHealth: boolean;
}

export interface WorkflowState {
    hasPendingWorldDataLoad: boolean;
    pendingGridState: boolean | null;
    lastAction: string;
}

export interface SavedUIState {
    terrain: number;
    unit: number;
    playerId: number;
    crossing: 'road' | 'bridge' | null;
    brushMode: string;
    brushSize: number;
    placementMode: 'terrain' | 'unit' | 'crossing' | 'clear';
    crossingConnectsTo: boolean[];
}

// =========================================================================
// Presenter Interface - Components depend on this, not the implementation
// =========================================================================

export interface IWorldEditorPresenter {
    // Tool State Actions
    selectTerrain(terrainType: number): void;
    selectUnit(unitType: number): void;
    selectPlayer(playerId: number): void;
    selectCrossing(crossingType: 'road' | 'bridge'): void;
    setCrossingConnectsTo(connectsTo: boolean[]): void;
    setBrushSize(mode: string, size: number): void;
    setPlacementMode(mode: 'terrain' | 'unit' | 'crossing' | 'clear'): void;

    // Visual State Actions
    setShowGrid(show: boolean): void;
    setShowCoordinates(show: boolean): void;
    setShowHealth(show: boolean): void;

    // Shape Tool Actions
    setShapeMode(shape: 'rectangle' | 'circle' | 'oval' | 'line' | null): void;
    setShapeFillMode(filled: boolean): void;
    getShapeMode(): 'rectangle' | 'circle' | 'oval' | 'line' | null;
    getShapeFillMode(): boolean;

    // Tile/Unit Click Handling
    handleTileClick(q: number, r: number): void;

    // Reference Image Actions
    setReferenceMode(mode: number): void;
    setReferenceAlpha(alpha: number): void;
    setReferencePosition(x: number, y: number): void;
    setReferenceScale(scaleX: number, scaleY: number): void;
    onReferencePositionUpdatedFromScene(x: number, y: number): void;
    onReferenceScaleUpdatedFromScene(scaleX: number, scaleY: number): void;

    // World Operations
    saveWorld(): Promise<{ success: boolean; worldId?: string; error?: string }>;
    clearWorld(): void;
    fillAllGrass(): void;
    shiftWorld(dQ: number, dR: number): void;

    // State Getters
    getToolState(): ToolState;
    getVisualState(): VisualState;
    getWorkflowState(): WorkflowState;
    getCurrentTerrain(): number;
    getCurrentUnit(): number;
    getCurrentPlayer(): number;
    getCurrentPlacementMode(): 'terrain' | 'unit' | 'crossing' | 'clear';
    getCurrentCrossing(): 'road' | 'bridge' | null;
    getCurrentBrushMode(): string;
    getCurrentBrushSize(): number;
    getShowGrid(): boolean;
    getShowCoordinates(): boolean;
    getShowHealth(): boolean;
    getWorld(): World | null;

    // Tab State
    getActiveTab(): 'tiles' | 'unit';
    setActiveTab(tab: 'tiles' | 'unit'): void;

    // State Persistence
    saveUIState(): void;
    restoreUIState(): boolean;
    hasSavedUIState(): boolean;
    resetToDefaults(): void;

    // Workflow State
    setHasPendingWorldDataLoad(pending: boolean): void;
    getHasPendingWorldDataLoad(): boolean;
    setPendingGridState(state: boolean | null): void;
    getPendingGridState(): boolean | null;
    getLastAction(): string;
}

// =========================================================================
// Presenter Implementation
// =========================================================================

export class WorldEditorPresenter implements IWorldEditorPresenter, EventSubscriber {
    // Dependencies
    private eventBus: EventBus;
    private world: World | null = null;

    // Component References
    private phaserEditor: PhaserEditorComponent | null = null;
    private toolsPanel: EditorToolsPanel | null = null;
    private worldStatsPanel: WorldStatsPanel | null = null;
    private referenceImagePanel: ReferenceImagePanel | null = null;

    // Tool State
    private toolState: ToolState = {
        selectedTerrain: 1,
        selectedUnit: 0,
        selectedPlayer: 1,
        selectedCrossing: null,
        // Default preset: horizontal crossing (LEFT + RIGHT)
        crossingConnectsTo: [true, false, false, true, false, false],
        placementMode: 'terrain',
        brushMode: 'brush',
        brushSize: 0
    };

    // Visual State
    private visualState: VisualState = {
        showGrid: false,
        showCoordinates: false,
        showHealth: false
    };

    // Workflow State
    private workflowState: WorkflowState = {
        hasPendingWorldDataLoad: false,
        pendingGridState: null,
        lastAction: 'initialized'
    };

    // State persistence
    private savedUIState: SavedUIState | null = null;

    // Shape tool state
    private currentShapeMode: 'rectangle' | 'circle' | 'oval' | 'line' | null = null;
    private shapeFillMode: boolean = true;

    // Tab state
    private activeTab: 'tiles' | 'unit' = 'tiles';

    // Callbacks
    private onStatusChange?: (status: string) => void;
    private onToast?: (title: string, message: string, type: 'success' | 'error' | 'info') => void;
    private onSaveButtonStateChange?: (hasChanges: boolean) => void;

    constructor(eventBus: EventBus) {
        this.eventBus = eventBus;
    }

    // =========================================================================
    // Lifecycle
    // =========================================================================

    public initialize(world: World): void {
        this.world = world;
        this.subscribeToWorldEvents();
    }

    private subscribeToWorldEvents(): void {
        this.eventBus.addSubscription(WorldEventTypes.TILES_CHANGED, null, this);
        this.eventBus.addSubscription(WorldEventTypes.UNITS_CHANGED, null, this);
        this.eventBus.addSubscription(WorldEventTypes.WORLD_LOADED, null, this);
        this.eventBus.addSubscription(WorldEventTypes.WORLD_SAVED, null, this);
        this.eventBus.addSubscription(WorldEventTypes.WORLD_CLEARED, null, this);
    }

    public handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
        switch (eventType) {
            case WorldEventTypes.TILES_CHANGED:
                this.onTilesChanged(data as TilesChangedEventData);
                break;
            case WorldEventTypes.UNITS_CHANGED:
                this.onUnitsChanged(data as UnitsChangedEventData);
                break;
            case WorldEventTypes.WORLD_LOADED:
                this.onWorldLoaded(data as WorldLoadedEventData);
                break;
            case WorldEventTypes.WORLD_SAVED:
                this.onWorldSaved(data);
                break;
            case WorldEventTypes.WORLD_CLEARED:
                this.onWorldCleared();
                break;
        }
    }

    private onTilesChanged(data: TilesChangedEventData): void {
        this.onSaveButtonStateChange?.(this.world?.getHasUnsavedChanges() ?? false);
    }

    private onUnitsChanged(data: UnitsChangedEventData): void {
        this.onSaveButtonStateChange?.(this.world?.getHasUnsavedChanges() ?? false);
    }

    private onWorldLoaded(data: WorldLoadedEventData): void {
        this.onStatusChange?.('Loaded');
    }

    private onWorldSaved(data: any): void {
        this.onStatusChange?.('Saved');
        this.onSaveButtonStateChange?.(false);
        if (data.success) {
            this.onToast?.('Success', 'World saved successfully', 'success');
        }
    }

    private onWorldCleared(): void {
        this.onStatusChange?.('Cleared');
        this.onSaveButtonStateChange?.(this.world?.getHasUnsavedChanges() ?? false);
    }

    // =========================================================================
    // Component Registration
    // =========================================================================

    public registerPhaserEditor(editor: PhaserEditorComponent): void {
        this.phaserEditor = editor;
        this.syncToolStateToPhaser();
        this.syncVisualStateToPhaser();
    }

    public registerToolsPanel(panel: EditorToolsPanel): void {
        this.toolsPanel = panel;
    }

    public registerWorldStatsPanel(panel: WorldStatsPanel): void {
        this.worldStatsPanel = panel;
    }

    public registerReferenceImagePanel(panel: ReferenceImagePanel): void {
        this.referenceImagePanel = panel;
    }

    // =========================================================================
    // Callback Registration
    // =========================================================================

    public setStatusChangeCallback(callback: (status: string) => void): void {
        this.onStatusChange = callback;
    }

    public setToastCallback(callback: (title: string, message: string, type: 'success' | 'error' | 'info') => void): void {
        this.onToast = callback;
    }

    public setSaveButtonStateCallback(callback: (hasChanges: boolean) => void): void {
        this.onSaveButtonStateChange = callback;
    }

    // =========================================================================
    // Tool State Actions
    // =========================================================================

    public selectTerrain(terrainType: number): void {
        this.toolState.selectedTerrain = terrainType;
        this.toolState.placementMode = terrainType === 0 ? 'clear' : 'terrain';
        this.workflowState.lastAction = 'select-terrain';
        this.syncToolStateToPhaser();
    }

    public selectUnit(unitType: number): void {
        this.toolState.selectedUnit = unitType;
        this.toolState.placementMode = 'unit';
        this.workflowState.lastAction = 'select-unit';
        this.syncToolStateToPhaser();
    }

    public selectPlayer(playerId: number): void {
        this.toolState.selectedPlayer = playerId;
        this.workflowState.lastAction = 'select-player';
        this.syncToolStateToPhaser();
    }

    public selectCrossing(crossingType: 'road' | 'bridge'): void {
        this.toolState.selectedCrossing = crossingType;
        this.toolState.placementMode = 'crossing';
        this.workflowState.lastAction = 'select-crossing';
        this.syncToolStateToPhaser();
    }

    /**
     * Set the preset connectsTo pattern for crossing placement
     * @param connectsTo Array of 6 booleans for hex directions
     */
    public setCrossingConnectsTo(connectsTo: boolean[]): void {
        this.toolState.crossingConnectsTo = [...connectsTo];
        this.workflowState.lastAction = 'set-crossing-connects-to';
    }

    /**
     * Get the current preset connectsTo pattern
     */
    public getCrossingConnectsTo(): boolean[] {
        return [...this.toolState.crossingConnectsTo];
    }

    public setBrushSize(mode: string, size: number): void {
        this.toolState.brushMode = mode;
        this.toolState.brushSize = size;
        this.workflowState.lastAction = 'set-brush-size';
        this.syncToolStateToPhaser();
    }

    public setPlacementMode(mode: 'terrain' | 'unit' | 'crossing' | 'clear'): void {
        this.toolState.placementMode = mode;
        this.workflowState.lastAction = 'set-placement-mode';
        this.syncToolStateToPhaser();
    }

    private syncToolStateToPhaser(): void {
        if (!this.phaserEditor?.editorScene) return;

        const scene = this.phaserEditor.editorScene;
        scene.setCurrentTerrain?.(this.toolState.selectedTerrain);
        scene.setCurrentUnit?.(this.toolState.selectedUnit);
        scene.setCurrentPlayer?.(this.toolState.selectedPlayer);
        scene.setBrushSize?.(this.toolState.brushSize);
        // Note: brushMode is handled by setShapeMode or via setEditorMode, not a separate method
    }

    // =========================================================================
    // Visual State Actions
    // =========================================================================

    public setShowGrid(show: boolean): void {
        this.visualState.showGrid = show;
        this.workflowState.lastAction = 'set-show-grid';
        this.phaserEditor?.editorScene?.setShowGrid?.(show);
    }

    public setShowCoordinates(show: boolean): void {
        this.visualState.showCoordinates = show;
        this.workflowState.lastAction = 'set-show-coordinates';
        this.phaserEditor?.editorScene?.setShowCoordinates?.(show);
    }

    public setShowHealth(show: boolean): void {
        this.visualState.showHealth = show;
        this.workflowState.lastAction = 'set-show-health';
        this.phaserEditor?.editorScene?.setShowUnitHealth?.(show);
    }

    private syncVisualStateToPhaser(): void {
        if (!this.phaserEditor?.editorScene) return;

        const scene = this.phaserEditor.editorScene;
        scene.setShowGrid?.(this.visualState.showGrid);
        scene.setShowCoordinates?.(this.visualState.showCoordinates);
        scene.setShowUnitHealth?.(this.visualState.showHealth);
    }

    // =========================================================================
    // Shape Tool Actions
    // =========================================================================

    public setShapeMode(shape: 'rectangle' | 'circle' | 'oval' | 'line' | null): void {
        this.currentShapeMode = shape;
        this.workflowState.lastAction = `set-shape-mode-${shape}`;
        this.phaserEditor?.editorScene?.setShapeMode?.(shape);
    }

    public setShapeFillMode(filled: boolean): void {
        this.shapeFillMode = filled;
        this.workflowState.lastAction = `set-shape-fill-${filled}`;
        this.phaserEditor?.editorScene?.setShapeFillMode?.(filled);
    }

    public getShapeMode(): 'rectangle' | 'circle' | 'oval' | 'line' | null {
        return this.currentShapeMode;
    }

    public getShapeFillMode(): boolean {
        return this.shapeFillMode;
    }

    // =========================================================================
    // Tile/Unit Click Handling
    // =========================================================================

    public handleTileClick(q: number, r: number): void {
        if (!this.world) return;

        const playerId = this.getPlayerIdForTerrain(this.toolState.selectedTerrain);

        switch (this.toolState.placementMode) {
            case 'terrain':
                this.paintTerrain(q, r, playerId);
                break;
            case 'unit':
                this.placeUnit(q, r);
                break;
            case 'crossing':
                this.toggleCrossing(q, r);
                break;
            case 'clear':
                this.clearTile(q, r);
                break;
        }
    }

    private paintTerrain(q: number, r: number, playerId: number): void {
        if (!this.world) return;

        if (this.toolState.brushSize === 0) {
            // Toggle behavior for single tile: if same terrain type and player, remove it
            const existingTile = this.world.getTileAt(q, r);
            if (existingTile &&
                existingTile.tileType === this.toolState.selectedTerrain &&
                existingTile.player === playerId) {
                this.world.removeTileAt(q, r);
                this.world.removeUnitAt(q, r);
                return;
            }
            this.world.setTileAt(q, r, this.toolState.selectedTerrain, playerId);
        } else {
            const tiles = this.getTilesForBrush(q, r);
            tiles.forEach(([tq, tr]) => {
                this.world!.setTileAt(tq, tr, this.toolState.selectedTerrain, playerId);
            });
        }
    }

    private placeUnit(q: number, r: number): void {
        if (!this.world) return;

        if (this.toolState.brushSize === 0) {
            // Toggle behavior for single unit: if same unit type and player, remove it
            const existingUnit = this.world.getUnitAt(q, r);
            if (existingUnit &&
                existingUnit.unitType === this.toolState.selectedUnit &&
                existingUnit.player === this.toolState.selectedPlayer) {
                this.world.removeUnitAt(q, r);
                return;
            }
            this.world.setUnitAt(q, r, this.toolState.selectedUnit, this.toolState.selectedPlayer);
        } else {
            const tiles = this.getTilesForBrush(q, r);
            tiles.forEach(([tq, tr]) => {
                this.world!.setUnitAt(tq, tr, this.toolState.selectedUnit, this.toolState.selectedPlayer);
            });
        }
    }

    /**
     * Ensure appropriate terrain exists for a crossing
     * - Roads require land (non-water) - if water or empty, set to Plains
     * - Bridges require water - if not water or empty, set to regular water
     */
    private ensureTerrainForCrossing(q: number, r: number, crossingType: CrossingType): void {
        if (!this.world) return;

        const tile = this.world.getTileAt(q, r);
        const isRoad = crossingType === CrossingType.CROSSING_TYPE_ROAD;
        const theme = this.phaserEditor?.editorScene?.getAssetProvider()?.getTheme()!;

        if (!tile || theme.canPlaceCrossing(tile.tileType, crossingType)) {
            this.world.setTileAt(q, r, theme.defaultCrossingTerrain(crossingType), 0);
        }
    }

    /**
     * Handle clicking on a tile while in crossing mode.
     * Toggle behavior: click to place with preset connectsTo pattern, click again to remove.
     */
    private toggleCrossing(q: number, r: number): void {
        if (!this.world || !this.toolState.selectedCrossing) return;

        const crossingType = this.toolState.selectedCrossing === 'road'
            ? CrossingType.CROSSING_TYPE_ROAD
            : CrossingType.CROSSING_TYPE_BRIDGE;

        const placeCrossing = (tq: number, tr: number) => {
            // Toggle: if crossing exists at this location, delete it (without affecting neighbors)
            if (this.world!.hasCrossing(tq, tr)) {
                this.world!.deleteCrossing(tq, tr);
            } else {
                // Ensure terrain is appropriate
                this.ensureTerrainForCrossing(tq, tr, crossingType);

                // Place crossing with preset connectsTo pattern
                const key = `${tq},${tr}`;
                this.world!.crossings[key] = {
                    type: crossingType,
                    connectsTo: [...this.toolState.crossingConnectsTo]
                };

                // Emit change event
                this.world!['addCrossingChange'](tq, tr, this.world!.crossings[key]);
            }
        };

        if (this.toolState.brushSize === 0) {
            placeCrossing(q, r);
        } else {
            const tiles = this.getTilesForBrush(q, r);
            tiles.forEach(([tq, tr]) => {
                placeCrossing(tq, tr);
            });
        }
    }

    private clearTile(q: number, r: number): void {
        if (!this.world) return;

        if (this.toolState.brushSize === 0) {
            this.world.removeTileAt(q, r);
            this.world.removeUnitAt(q, r);
            this.world.removeCrossing(q, r);
        } else {
            const tiles = this.getTilesForBrush(q, r);
            tiles.forEach(([tq, tr]) => {
                this.world!.removeTileAt(tq, tr);
                this.world!.removeUnitAt(tq, tr);
                this.world!.removeCrossing(tq, tr);
            });
        }
    }

    private getTilesForBrush(q: number, r: number): [number, number][] {
        if (!this.world) return [[q, r]];

        if (this.toolState.brushMode === 'brush') {
            return this.world.radialNeighbours(q, r, this.toolState.brushSize);
        } else if (this.toolState.brushMode === 'fill') {
            return this.world.floodNeighbors(q, r, this.toolState.brushSize);
        }
        return [[q, r]];
    }

    private getPlayerIdForTerrain(terrainType: number): number {
        // Use theme to determine if terrain is a city tile that supports player ownership
        const theme = this.phaserEditor?.editorScene?.getAssetProvider()?.getTheme();
        const isCityTile = theme?.isCityTile(terrainType) ?? false;
        return isCityTile ? this.toolState.selectedPlayer : 0;
    }

    // =========================================================================
    // Reference Image Actions
    // =========================================================================

    public setReferenceMode(mode: number): void {
        this.phaserEditor?.editorScene?.setReferenceMode?.(mode);
    }

    public setReferenceAlpha(alpha: number): void {
        this.phaserEditor?.editorScene?.setReferenceAlpha?.(alpha);
    }

    public setReferencePosition(x: number, y: number): void {
        this.phaserEditor?.editorScene?.setReferencePosition?.(x, y);
    }

    public setReferenceScale(scaleX: number, scaleY: number): void {
        this.phaserEditor?.editorScene?.setReferenceScale?.(scaleX, scaleY);
    }

    public onReferencePositionUpdatedFromScene(x: number, y: number): void {
        this.referenceImagePanel?.updatePositionDisplay(x, y);
    }

    public onReferenceScaleUpdatedFromScene(scaleX: number, scaleY: number): void {
        this.referenceImagePanel?.updateScaleDisplay(scaleX, scaleY);
    }

    // =========================================================================
    // World Operations
    // =========================================================================

    public async saveWorld(): Promise<{ success: boolean; worldId?: string; error?: string }> {
        if (!this.world) {
            return { success: false, error: 'No world loaded' };
        }

        this.onStatusChange?.('Saving...');

        try {
            const result = await this.world.save();
            return result;
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : 'Unknown error';
            this.onStatusChange?.('Save Error');
            this.onToast?.('Error', `Failed to save: ${errorMessage}`, 'error');
            return { success: false, error: errorMessage };
        }
    }

    public clearWorld(): void {
        if (!this.world) return;
        this.world.clearAll();
        this.onToast?.('World Cleared', 'All tiles and units have been removed', 'info');
    }

    public fillAllGrass(): void {
        if (!this.world) return;
        this.world.fillAllTerrain(1, 0);
        this.onToast?.('Fill Complete', 'All visible tiles filled with grass', 'info');
    }

    public shiftWorld(dQ: number, dR: number): void {
        if (!this.world) return;
        if (dQ === 0 && dR === 0) return;
        this.world.shiftWorld(dQ, dR);
        this.workflowState.lastAction = `shift-world-${dQ}-${dR}`;
    }

    // =========================================================================
    // State Getters
    // =========================================================================

    public getToolState(): ToolState {
        return { ...this.toolState };
    }

    public getVisualState(): VisualState {
        return { ...this.visualState };
    }

    public getWorkflowState(): WorkflowState {
        return { ...this.workflowState };
    }

    public getCurrentTerrain(): number {
        return this.toolState.selectedTerrain;
    }

    public getCurrentUnit(): number {
        return this.toolState.selectedUnit;
    }

    public getCurrentPlayer(): number {
        return this.toolState.selectedPlayer;
    }

    public getCurrentPlacementMode(): 'terrain' | 'unit' | 'crossing' | 'clear' {
        return this.toolState.placementMode;
    }

    public getCurrentCrossing(): 'road' | 'bridge' | null {
        return this.toolState.selectedCrossing;
    }

    public getCurrentBrushMode(): string {
        return this.toolState.brushMode;
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

    public getShowHealth(): boolean {
        return this.visualState.showHealth;
    }

    public getWorld(): World | null {
        return this.world;
    }

    // =========================================================================
    // Tab State
    // =========================================================================

    public getActiveTab(): 'tiles' | 'unit' {
        return this.activeTab;
    }

    public setActiveTab(tab: 'tiles' | 'unit'): void {
        this.activeTab = tab;
        this.toolsPanel?.switchToTab?.(tab);
    }

    // =========================================================================
    // State Persistence
    // =========================================================================

    public saveUIState(): void {
        this.savedUIState = {
            terrain: this.toolState.selectedTerrain,
            unit: this.toolState.selectedUnit,
            playerId: this.toolState.selectedPlayer,
            crossing: this.toolState.selectedCrossing,
            brushMode: this.toolState.brushMode,
            brushSize: this.toolState.brushSize,
            placementMode: this.toolState.placementMode,
            crossingConnectsTo: [...this.toolState.crossingConnectsTo]
        };
        this.workflowState.lastAction = 'ui-state-saved';
    }

    public restoreUIState(): boolean {
        if (!this.savedUIState) {
            return false;
        }

        this.toolState.selectedTerrain = this.savedUIState.terrain;
        this.toolState.selectedUnit = this.savedUIState.unit;
        this.toolState.selectedPlayer = this.savedUIState.playerId;
        this.toolState.selectedCrossing = this.savedUIState.crossing;
        this.toolState.brushMode = this.savedUIState.brushMode;
        this.toolState.brushSize = this.savedUIState.brushSize;
        this.toolState.placementMode = this.savedUIState.placementMode;
        this.toolState.crossingConnectsTo = [...this.savedUIState.crossingConnectsTo];

        this.syncToolStateToPhaser();
        this.workflowState.lastAction = 'ui-state-restored';
        return true;
    }

    public hasSavedUIState(): boolean {
        return this.savedUIState !== null;
    }

    public resetToDefaults(): void {
        this.toolState = {
            selectedTerrain: 1,
            selectedUnit: 0,
            selectedPlayer: 1,
            selectedCrossing: null,
            placementMode: 'terrain',
            brushMode: 'brush',
            brushSize: 0,
            crossingConnectsTo: [true, false, false, true, false, false]  // Default: horizontal
        };

        this.syncToolStateToPhaser();
        this.workflowState.lastAction = 'reset-to-defaults';
    }

    // =========================================================================
    // Workflow State
    // =========================================================================

    public setHasPendingWorldDataLoad(pending: boolean): void {
        this.workflowState.hasPendingWorldDataLoad = pending;
    }

    public getHasPendingWorldDataLoad(): boolean {
        return this.workflowState.hasPendingWorldDataLoad;
    }

    public setPendingGridState(state: boolean | null): void {
        this.workflowState.pendingGridState = state;
    }

    public getPendingGridState(): boolean | null {
        return this.workflowState.pendingGridState;
    }

    public getLastAction(): string {
        return this.workflowState.lastAction;
    }

    // =========================================================================
    // Validation & Debug
    // =========================================================================

    public validate(): { isValid: boolean; errors: string[] } {
        const errors: string[] = [];

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

        const validModes = ['terrain', 'unit', 'crossing', 'clear'];
        if (!validModes.includes(this.toolState.placementMode)) {
            errors.push(`Invalid placement mode: ${this.toolState.placementMode}`);
        }

        return { isValid: errors.length === 0, errors };
    }

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
}
