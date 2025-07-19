import { BaseComponent } from './Component';
import { EventBus, EditorEventTypes, TerrainSelectedPayload, UnitSelectedPayload, BrushSizeChangedPayload, PlacementModeChangedPayload, PlayerChangedPayload, TileClickedPayload, PhaserReadyPayload, TilePaintedPayload, UnitPlacedPayload, TileClearedPayload, UnitRemovedPayload } from './EventBus';
import { PhaserMapEditor } from './phaser/PhaserMapEditor';

/**
 * PhaserEditorComponent - Manages the Phaser.js-based map editor interface using BaseComponent architecture
 * 
 * Responsibilities:
 * - Initialize and manage Phaser.js map editor lifecycle
 * - Handle editor-specific DOM container setup
 * - Emit tile click events to EventBus
 * - Listen for tool changes (terrain, unit, brush size) from EditorToolsPanel
 * - Manage map rendering, camera controls, and visual settings
 * - Handle map data loading and saving operations
 * - Manage reference image features for overlay/background
 * 
 * Does NOT handle:
 * - Tool selection UI (handled by EditorToolsPanel)
 * - Layout management (handled by parent dockview)
 * - Save/load UI (will be handled by SaveLoadComponent)
 * - Direct DOM manipulation outside of phaser-container
 */
export class PhaserEditorComponent extends BaseComponent {
    private phaserEditor: PhaserMapEditor | null = null;
    private isInitialized: boolean = false;
    
    // Current tool state (synced from EditorToolsPanel)
    private currentTerrain: number = 1;
    private currentUnit: number = 0;
    private currentBrushSize: number = 0;
    private currentPlayerId: number = 1;
    private currentPlacementMode: 'terrain' | 'unit' | 'clear' = 'terrain';
    
    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('phaser-editor', rootElement, eventBus, debugMode);
    }
    
    protected initializeComponent(): void {
        this.log('Initializing PhaserEditorComponent');
        
        // Subscribe to tool changes from EditorToolsPanel
        this.subscribe<TerrainSelectedPayload>(EditorEventTypes.TERRAIN_SELECTED, (payload) => {
            this.handleTerrainSelected(payload.data);
        });
        
        this.subscribe<UnitSelectedPayload>(EditorEventTypes.UNIT_SELECTED, (payload) => {
            this.handleUnitSelected(payload.data);
        });
        
        this.subscribe<BrushSizeChangedPayload>(EditorEventTypes.BRUSH_SIZE_CHANGED, (payload) => {
            this.handleBrushSizeChanged(payload.data);
        });
        
        this.subscribe<PlacementModeChangedPayload>(EditorEventTypes.PLACEMENT_MODE_CHANGED, (payload) => {
            this.handlePlacementModeChanged(payload.data);
        });
        
        this.subscribe<PlayerChangedPayload>(EditorEventTypes.PLAYER_CHANGED, (payload) => {
            this.handlePlayerChanged(payload.data);
        });
        
        this.log('PhaserEditorComponent component initialized');
    }
    
    protected bindToDOM(): void {
        try {
            this.log('Binding PhaserEditorComponent to DOM');
            
            // Set up Phaser container within our root element
            this.setupPhaserContainer();
            
            this.log('Phaser container setup complete');
            
            // Now initialize Phaser editor with the properly set up container
            this.initializePhaserEditor();
            
            this.log('PhaserEditorComponent bound to DOM');
            
        } catch (error) {
            this.handleError('Failed to bind PhaserEditorComponent to DOM', error);
        }
    }
    
    protected destroyComponent(): void {
        this.log('Destroying PhaserEditorComponent');
        
        // Destroy Phaser editor
        if (this.phaserEditor) {
            this.phaserEditor.destroy();
            this.phaserEditor = null;
        }
        
        // Remove Phaser container
        const phaserContainer = document.getElementById('phaser-container');
        if (phaserContainer) {
            phaserContainer.remove();
        }
        
        this.isInitialized = false;
        this.log('PhaserEditorComponent destroyed');
    }
    
    /**
     * Set up the Phaser container element
     */
    private setupPhaserContainer(): void {
        // First try to find the existing editor-canvas-container from the template
        let phaserContainer = this.findElement('#editor-canvas-container');
        
        if (phaserContainer) {
            // Rename the existing container to phaser-container for PhaserMapEditor
            phaserContainer.id = 'phaser-container';
            this.log('Using existing editor-canvas-container as phaser-container');
        } else {
            // Fallback: create new container if template container not found
            phaserContainer = document.createElement('div');
            phaserContainer.id = 'phaser-container';
            phaserContainer.style.width = '100%';
            phaserContainer.style.height = '100%';
            phaserContainer.style.minWidth = '800px';
            phaserContainer.style.minHeight = '600px';
            this.rootElement.appendChild(phaserContainer);
            this.log('Created new phaser-container');
        }
        
        this.log('Phaser container setup complete');
    }

    /**
     * Wait for container to become visible before initializing Phaser
     */
    private waitForContainerVisible(containerElement: HTMLElement): void {
        const checkVisibility = () => {
            const rect = containerElement.getBoundingClientRect();
            
            if (rect.width > 0 && rect.height > 0) {
                // Continue with Phaser initialization
                this.phaserEditor = new PhaserMapEditor(containerElement);
                this.setupPhaserEventHandlers();
                
                const isDarkMode = document.documentElement.classList.contains('dark');
                this.phaserEditor.setTheme(isDarkMode);
                
                this.isInitialized = true;
                this.log('Phaser editor initialized successfully');
                
                // Emit ready event for other components (async to allow parent assignment to complete)
                setTimeout(() => {
                    this.emit(EditorEventTypes.PHASER_READY, {});
                }, 0);
            } else {
                // Check again after a short delay
                setTimeout(checkVisibility, 50);
            }
        };
        
        // Start checking
        setTimeout(checkVisibility, 50);
    }
    
    /**
     * Initialize the Phaser editor
     */
    private initializePhaserEditor(): void {
        try {
            this.log('Initializing Phaser editor...');
            
            // Find the container element that we just set up
            const containerElement = this.findElement('#phaser-container');
            if (!containerElement) {
                throw new Error('Phaser container element not found after setup');
            }
            
            // Wait for container to have dimensions before initializing Phaser
            const rect = containerElement.getBoundingClientRect();
            if (rect.width === 0 || rect.height === 0) {
                this.waitForContainerVisible(containerElement);
                return;
            }
            
            // Create Phaser editor instance with the element directly
            this.phaserEditor = new PhaserMapEditor(containerElement);
            
            // Set up event handlers
            this.setupPhaserEventHandlers();
            
            // Apply current theme
            const isDarkMode = document.documentElement.classList.contains('dark');
            this.phaserEditor.setTheme(isDarkMode);
            
            this.isInitialized = true;
            this.log('Phaser editor initialized successfully');
            
            // Emit ready event for other components (async to allow parent assignment to complete)
            setTimeout(() => {
                this.emit(EditorEventTypes.PHASER_READY, {});
            }, 0);
            
        } catch (error) {
            this.handleError('Failed to initialize Phaser editor', error);
        }
    }
    
    /**
     * Set up event handlers for Phaser editor
     */
    private setupPhaserEventHandlers(): void {
        if (!this.phaserEditor) return;
        
        // Handle tile clicks - emit to EventBus for other components
        this.phaserEditor.onTileClick((q, r) => {
            this.log(`Tile clicked: Q=${q}, R=${r}`);
            
            this.emit<TileClickedPayload>(EditorEventTypes.TILE_CLICKED, {
                q: q,
                r: r
            });
            
            // Handle painting based on current mode
            this.handleTileClick(q, r);
        });
        
        // Handle map changes
        this.phaserEditor.onMapChange(() => {
            this.log('Map changed in Phaser');
            this.emit(EditorEventTypes.MAP_CHANGED, {});
        });
        
        // Handle reference scale changes
        this.phaserEditor.onReferenceScaleChange((x: number, y: number) => {
            this.emit(EditorEventTypes.REFERENCE_SCALE_CHANGED, { x, y });
        });
        
        this.log('Phaser event handlers setup complete');
    }
    
    /**
     * Handle tile clicks for painting
     */
    private handleTileClick(q: number, r: number): void {
        if (!this.phaserEditor || !this.isInitialized) {
            return;
        }
        
        try {
            switch (this.currentPlacementMode) {
                case 'terrain':
                    this.phaserEditor.paintTile(q, r, this.currentTerrain, 0, this.currentBrushSize);
                    this.log(`Painted terrain ${this.currentTerrain} at Q=${q}, R=${r} with brush size ${this.currentBrushSize}`);
                    
                    // Emit tile painted event for other components (like Map data management)
                    this.emit<TilePaintedPayload>(EditorEventTypes.TILE_PAINTED, {
                        q: q,
                        r: r,
                        terrainType: this.currentTerrain,
                        playerColor: 0,
                        brushSize: this.currentBrushSize
                    });
                    break;
                    
                case 'unit':
                    this.phaserEditor.paintUnit(q, r, this.currentUnit, this.currentPlayerId);
                    this.log(`Painted unit ${this.currentUnit} (player ${this.currentPlayerId}) at Q=${q}, R=${r}`);
                    
                    // Emit unit placed event for other components
                    this.emit<UnitPlacedPayload>(EditorEventTypes.UNIT_PLACED, {
                        q: q,
                        r: r,
                        unitType: this.currentUnit,
                        playerId: this.currentPlayerId
                    });
                    break;
                    
                case 'clear':
                    this.phaserEditor.removeTile(q, r);
                    this.phaserEditor.removeUnit(q, r);
                    this.log(`Cleared tile and unit at Q=${q}, R=${r}`);
                    
                    // Emit separate events for tile and unit clearing
                    this.emit<TileClearedPayload>(EditorEventTypes.TILE_CLEARED, { q: q, r: r });
                    this.emit<UnitRemovedPayload>(EditorEventTypes.UNIT_REMOVED, { q: q, r: r });
                    break;
            }
        } catch (error) {
            this.handleError(`Failed to handle tile click at Q=${q}, R=${r}`, error);
        }
    }
    
    /**
     * Event handlers for tool changes from EditorToolsPanel
     */
    private handleTerrainSelected(data: { terrainType: number; terrainName: string }): void {
        this.currentTerrain = data.terrainType;
        this.currentPlacementMode = data.terrainType === 0 ? 'clear' : 'terrain';
        
        if (this.phaserEditor) {
            this.phaserEditor.setTerrain(data.terrainType);
        }
        
        this.log(`Terrain selection updated: ${data.terrainType} (${data.terrainName})`);
    }
    
    private handleUnitSelected(data: { unitType: number; unitName: string; playerId: number }): void {
        this.currentUnit = data.unitType;
        this.currentPlayerId = data.playerId;
        this.currentPlacementMode = 'unit';
        
        this.log(`Unit selection updated: ${data.unitType} (${data.unitName}) for player ${data.playerId}`);
    }
    
    private handleBrushSizeChanged(data: { brushSize: number; sizeName: string }): void {
        this.currentBrushSize = data.brushSize;
        
        if (this.phaserEditor) {
            this.phaserEditor.setBrushSize(data.brushSize);
        }
        
        this.log(`Brush size updated: ${data.sizeName}`);
    }
    
    private handlePlacementModeChanged(data: { mode: 'terrain' | 'unit' | 'clear' }): void {
        this.currentPlacementMode = data.mode;
        this.log(`Placement mode changed to: ${data.mode}`);
    }
    
    private handlePlayerChanged(data: { playerId: number }): void {
        this.currentPlayerId = data.playerId;
        this.log(`Player changed to: ${data.playerId}`);
    }
    
    // Public API methods (for external access)
    
    /**
     * Check if Phaser editor is initialized
     */
    public getIsInitialized(): boolean {
        return this.isInitialized;
    }
    
    /**
     * Set theme for the editor
     */
    public setTheme(isDark: boolean): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.setTheme(isDark);
            this.log(`Theme set to: ${isDark ? 'dark' : 'light'}`);
        }
    }
    
    /**
     * Set grid visibility
     */
    public setShowGrid(show: boolean): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.setShowGrid(show);
            this.log(`Grid visibility set to: ${show}`);
        }
    }
    
    /**
     * Set coordinate visibility
     */
    public setShowCoordinates(show: boolean): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.setShowCoordinates(show);
            this.log(`Coordinate visibility set to: ${show}`);
        }
    }
    
    /**
     * Load map tiles data
     */
    public async setTilesData(tiles: Array<{ q: number; r: number; terrain: number; color: number }>): Promise<void> {
        if (this.phaserEditor && this.isInitialized) {
            try {
                await this.phaserEditor.setTilesData(tiles);
                this.log(`Loaded ${tiles.length} tiles into Phaser`);
            } catch (error) {
                this.handleError('Failed to load tiles data', error);
            }
        }
    }
    
    /**
     * Paint a unit at specific coordinates
     */
    public paintUnit(q: number, r: number, unitType: number, playerId: number): boolean {
        if (this.phaserEditor && this.isInitialized) {
            try {
                this.phaserEditor.paintUnit(q, r, unitType, playerId);
                this.log(`Painted unit ${unitType} (player ${playerId}) at Q=${q}, R=${r}`);
                return true;
            } catch (error) {
                this.handleError(`Failed to paint unit at Q=${q}, R=${r}`, error);
                return false;
            }
        }
        return false;
    }
    
    /**
     * Get current tiles data
     */
    public getTilesData(): Array<{ q: number; r: number; terrain: number; color: number }> {
        if (this.phaserEditor && this.isInitialized) {
            return this.phaserEditor.getTilesData();
        }
        return [];
    }
    
    /**
     * Get current units data
     */
    public getUnitsData(): Array<{ q: number; r: number; unitType: number; playerId: number }> {
        if (this.phaserEditor && this.isInitialized) {
            return this.phaserEditor.getUnitsData();
        }
        return [];
    }
    
    /**
     * Get viewport center for map generation
     */
    public getViewportCenter(): { q: number; r: number } {
        if (this.phaserEditor && this.isInitialized) {
            return this.phaserEditor.getViewportCenter();
        }
        return { q: 0, r: 0 };
    }
    
    /**
     * Map generation methods
     */
    public fillAllTerrain(terrain: number, color: number = 0): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.fillAllTerrain(terrain, color);
            this.log(`Filled all terrain with type ${terrain}`);
        }
    }
    
    public randomizeTerrain(): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.randomizeTerrain();
            this.log('Terrain randomized');
        }
    }
    
    public createIslandPattern(centerQ: number, centerR: number, radius: number): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.createIslandPattern(centerQ, centerR, radius);
            this.log(`Created island pattern at Q=${centerQ}, R=${centerR} with radius ${radius}`);
        }
    }
    
    public clearAllTiles(): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.clearAllTiles();
            this.log('All tiles cleared');
        }
    }
    
    public clearAllUnits(): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.clearAllUnits();
            this.log('All units cleared');
        }
    }
    
    public paintTile(q: number, r: number, terrain: number, color: number, brushSize: number = 0): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.paintTile(q, r, terrain, color, brushSize);
        }
    }
    
    public removeTile(q: number, r: number): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.removeTile(q, r);
        }
    }
    
    public removeUnit(q: number, r: number): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.removeUnit(q, r);
        }
    }
    
    /**
     * Reference image methods
     */
    public async loadReferenceFromClipboard(): Promise<boolean> {
        if (this.phaserEditor && this.isInitialized) {
            try {
                const result = await this.phaserEditor.loadReferenceFromClipboard();
                this.log(result ? 'Reference image loaded from clipboard' : 'No image found in clipboard');
                return result;
            } catch (error) {
                this.handleError('Failed to load reference image from clipboard', error);
                return false;
            }
        }
        return false;
    }
    
    public async loadReferenceFromFile(file: File): Promise<boolean> {
        if (this.phaserEditor && this.isInitialized) {
            try {
                const result = await this.phaserEditor.loadReferenceFromFile(file);
                this.log(result ? `Reference image loaded from file: ${file.name}` : 'Failed to load file');
                return result;
            } catch (error) {
                this.handleError(`Failed to load reference image from file: ${file.name}`, error);
                return false;
            }
        }
        return false;
    }
    
    public setReferenceMode(mode: number): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.setReferenceMode(mode);
            const modeNames = ['hidden', 'background', 'overlay'];
            this.log(`Reference mode set to: ${modeNames[mode] || mode}`);
        }
    }
    
    public setReferenceAlpha(alpha: number): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.setReferenceAlpha(alpha);
            this.log(`Reference alpha set to: ${alpha}`);
        }
    }
    
    public setReferencePosition(x: number, y: number): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.setReferencePosition(x, y);
            this.log(`Reference position set to: (${x}, ${y})`);
        }
    }
    
    public setReferenceScale(x: number, y: number): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.setReferenceScale(x, y);
            this.log(`Reference scale set to: (${x}, ${y})`);
        }
    }
    
    public setReferenceScaleFromTopLeft(x: number, y: number): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.setReferenceScaleFromTopLeft(x, y);
            this.log(`Reference scale set from top-left to: (${x}, ${y})`);
        }
    }
    
    public getReferenceState(): {
        mode: number;
        alpha: number;
        position: { x: number; y: number };
        scale: { x: number; y: number };
        hasImage: boolean;
    } | null {
        if (this.phaserEditor && this.isInitialized) {
            return this.phaserEditor.getReferenceState();
        }
        return null;
    }
    
    public clearReferenceImage(): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.clearReferenceImage();
            this.log('Reference image cleared');
        }
    }
    
    /**
     * Set callback for when Phaser scene is ready
     */
    public onSceneReady(callback: () => void): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.onSceneReady(callback);
        } else {
            this.log('Cannot set scene ready callback - Phaser not initialized');
        }
    }
    
    /**
     * Register map change callback
     */
    public onMapChange(callback: () => void): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.onMapChange(callback);
        }
    }
    
    /**
     * Register reference scale change callback
     */
    public onReferenceScaleChange(callback: (x: number, y: number) => void): void {
        if (this.phaserEditor && this.isInitialized) {
            this.phaserEditor.onReferenceScaleChange(callback);
        }
    }
}
