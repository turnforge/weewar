import { BaseComponent, DOMValidation } from './Component';
import { EventBus, EventPayload, EventTypes, WorldDataLoadedPayload } from './EventBus';
import { PhaserViewer } from './PhaserViewer';
import { World } from './World';
import { ComponentLifecycle } from './ComponentLifecycle';

/**
 * WorldViewer Component - Manages Phaser-based world visualization
 * Responsible for:
 * - Phaser initialization and lifecycle management
 * - World data rendering (tiles and units)
 * - Camera controls and viewport management
 * - Theme and display options
 * 
 * Layout and styling are handled by parent container and CSS classes.
 */
export class WorldViewer extends BaseComponent implements ComponentLifecycle {
    private phaserViewer: PhaserViewer | null;
    private loadedWorldData: WorldDataLoadedPayload | null;
    private viewerContainer: HTMLElement | null;
    
    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        console.log('WorldViewer constructor: received eventBus:', eventBus);
        super('world-viewer', rootElement, eventBus, debugMode);
    }
    
    protected initializeComponent(): void {
        this.log('Initializing WorldViewer component');
        
        // Subscribe to world data events
        this.subscribe<WorldDataLoadedPayload>(EventTypes.WORLD_DATA_LOADED, (payload) => {
            this.handleWorldDataLoaded(payload);
        });
        
        this.log('WorldViewer component initialized');
    }
    
    protected bindToDOM(): void {
        this.log('Binding WorldViewer to DOM');
        
        // Find the phaser-viewer-container within the root element
        let phaserContainer = this.rootElement.querySelector('#phaser-viewer-container') as HTMLElement;
        
        if (!phaserContainer) {
            // If not found as child, check if root element IS the phaser container
            if (this.rootElement.id === 'phaser-viewer-container') {
                phaserContainer = this.rootElement;
            } else {
                // Create the phaser container as a child
                console.warn('phaser-viewer-container not found, creating one');
                phaserContainer = document.createElement('div');
                phaserContainer.id = 'phaser-viewer-container';
                phaserContainer.className = 'w-full h-full min-h-96';
                this.rootElement.appendChild(phaserContainer);
            }
        }
        
        this.viewerContainer = phaserContainer;
        
        // Ensure the container has the right classes
        if (!this.viewerContainer.classList.contains('w-full')) {
            this.viewerContainer.className = 'w-full h-full min-h-96';
        }
        
        this.log('WorldViewer bound to DOM, container:', this.viewerContainer);
        
        // Phaser initialization will happen in activate() phase, not here
        console.log('WorldViewer: DOM binding complete, waiting for activate() phase');
    }
    
    protected destroyComponent(): void {
        this.log('Destroying WorldViewer component');
        
        // Clean up Phaser viewer
        if (this.phaserViewer) {
            this.phaserViewer.destroy();
            this.phaserViewer = null;
        }
        
        this.loadedWorldData = null;
        this.viewerContainer = null;
    }
    
    
    public validateDOM(rootElement: HTMLElement): DOMValidation {
        const validation: DOMValidation = {
            isValid: true,
            missingElements: [],
            invalidElements: [],
            warnings: []
        };
        
        // Check for Phaser container
        const phaserContainer = rootElement.querySelector('#phaser-viewer-container');
        if (!phaserContainer) {
            validation.isValid = false;
            validation.missingElements.push('phaser-viewer-container');
        }
        
        return validation;
    }
    
    /**
     * Initialize the Phaser viewer
     */
    private initializePhaserViewer(): void {
        console.log('WorldViewer: initializePhaserViewer() called');
        
        // Guard against multiple initialization
        if (this.phaserViewer) {
            console.log('WorldViewer: PhaserViewer already initialized, skipping');
            return;
        }
        
        if (!this.viewerContainer) {
            throw new Error('Viewer container not available');
        }
        
        this.log('Initializing Phaser viewer');
        console.log('WorldViewer: viewerContainer is:', this.viewerContainer);
        
        // Create new PhaserViewer instance
        this.phaserViewer = new PhaserViewer();
        
        // Set up logging
        this.phaserViewer.onLog((message: string) => {
            this.log(`PhaserViewer: ${message}`);
        });
        
        // Initialize with container ID - Phaser will adapt to whatever size the parent provides
        this.phaserViewer.initialize(this.viewerContainer.id);
        
        // Emit ready event
        console.log('WorldViewer: Emitting WORLD_VIEWER_READY event');
        this.emit(EventTypes.WORLD_VIEWER_READY, {
            componentId: this.componentId,
            success: true
        });
        console.log('WorldViewer: WORLD_VIEWER_READY event emitted');
        
        // Load world data if we have it
        if (this.loadedWorldData) {
            this.loadWorldIntoViewer(this.loadedWorldData);
        }
    }
    
    /**
     * Handle world data loaded event
     */
    private handleWorldDataLoaded(payload: EventPayload<WorldDataLoadedPayload>): void {
        this.log(`Received world data for world: ${payload.data.worldId}`);
        this.loadedWorldData = payload.data;
        
        // Load into Phaser if viewer is ready
        if (this.phaserViewer && this.phaserViewer.getIsInitialized()) {
            this.loadWorldIntoViewer(payload.data);
        }
    }
    
    /**
     * Load world data into the Phaser viewer
     */
    private async loadWorldIntoViewer(worldData: WorldDataLoadedPayload): Promise<void> {
        if (!this.phaserViewer || !this.phaserViewer.getIsInitialized()) {
            this.log('Phaser viewer not ready, deferring world load');
            return;
        }
        this.log('Loading world data into Phaser viewer');
        
        // Convert world data to Phaser format
        const tilesArray: Array<{ q: number; r: number; terrain: number; color: number }> = [];
        const unitsArray: Array<{ q: number; r: number; unitType: number; playerId: number }> = [];
        
        // Process tiles from bounds
        if (worldData.bounds) {
            for (let q = worldData.bounds.minQ; q <= worldData.bounds.maxQ; q++) {
                for (let r = worldData.bounds.minR; r <= worldData.bounds.maxR; r++) {
                    // This would need to be coordinated with the world data structure
                    // For now, create placeholder logic
                    tilesArray.push({
                        q: q,
                        r: r,
                        terrain: 1, // Default grass
                        color: 0
                    });
                }
            }
        }
        
        // Load into Phaser viewer
        await this.phaserViewer.loadWorldData(tilesArray, unitsArray);
    }
    
    /**
     * Public API for loading world data
     */
    public async loadWorld(worldData: any): Promise<void> {
        if (!worldData) {
            throw new Error('No world data provided');
        }
        
        this.log('Loading world data');
        
        // Process world data
        const world = World.deserialize(worldData);
        const allTiles = world.getAllTiles();
        const allUnits = world.getAllUnits();
        
        // Convert to arrays
        const tilesArray: Array<{ q: number; r: number; terrain: number; color: number }> = [];
        const unitsArray: Array<{ q: number; r: number; unitType: number; playerId: number }> = [];
        
        allTiles.forEach(tile => {
            tilesArray.push({
                q: tile.q,
                r: tile.r,
                terrain: tile.tileType,
                color: tile.playerId || 0
            });
        });
        
        allUnits.forEach(unit => {
            unitsArray.push({
                q: unit.q,
                r: unit.r,
                unitType: unit.unitType,
                playerId: unit.playerId
            });
        });
        
        // Calculate bounds and stats
        const bounds = world.getBounds();
        
        // Store world data
        this.loadedWorldData = {
            worldId: worldData.id || 'unknown',
            totalTiles: allTiles.length,
            totalUnits: allUnits.length,
            bounds: bounds ? {
                minQ: bounds.minQ,
                maxQ: bounds.maxQ,
                minR: bounds.minR,
                maxR: bounds.maxR
            } : { minQ: 0, maxQ: 0, minR: 0, maxR: 0 },
            terrainCounts: this.calculateTerrainCounts(allTiles)
        };
        
        // Load into Phaser if ready
        if (this.phaserViewer && this.phaserViewer.getIsInitialized()) {
            await this.phaserViewer.loadWorldData(tilesArray, unitsArray);
        }
        
        // Emit data loaded event for other components
        this.emit(EventTypes.WORLD_DATA_LOADED, this.loadedWorldData);
    }
    
    /**
     * Calculate terrain counts for statistics
     */
    private calculateTerrainCounts(tiles: any[]): { [terrainType: number]: number } {
        const counts: { [terrainType: number]: number } = {};
        
        tiles.forEach(tile => {
            counts[tile.tileType] = (counts[tile.tileType] || 0) + 1;
        });
        
        return counts;
    }
    
    /**
     * Set display options
     */
    public setShowGrid(show: boolean): void {
        if (this.phaserViewer) {
            this.phaserViewer.setShowGrid(show);
        }
    }
    
    public setShowCoordinates(show: boolean): void {
        if (this.phaserViewer) {
            this.phaserViewer.setShowCoordinates(show);
        }
    }
    
    public setTheme(isDark: boolean): void {
        if (this.phaserViewer) {
            this.phaserViewer.setTheme(isDark);
        }
    }
    
    /**
     * Camera controls
     */
    public getZoom(): number {
        return this.phaserViewer?.getZoom() || 1;
    }
    
    public setZoom(zoom: number): void {
        if (this.phaserViewer) {
            this.phaserViewer.setZoom(zoom);
        }
    }
    
    /**
     * Resize the viewer
     */
    public resize(width?: number, height?: number): void {
        if (this.phaserViewer && this.viewerContainer) {
            const w = width || this.viewerContainer.clientWidth;
            const h = height || this.viewerContainer.clientHeight;
            this.phaserViewer.resize(w, h);
        }
    }
    
    /**
     * Check if viewer is ready
     */
    public isPhaserReady(): boolean {
        return this.phaserViewer?.getIsInitialized() || false;
    }

    // =============================================================================
    // ComponentLifecycle Interface Implementation
    // =============================================================================

    /**
     * Phase 1: Initialize DOM and discover child components
     */
    initializeDOM(): ComponentLifecycle[] {
        console.log('WorldViewer: initializeDOM() - Phase 1');
        
        // DOM setup is already done in bindToDOM(), just return no child components
        return [];
    }

    /**
     * Phase 2: Inject dependencies (none needed for WorldViewer)
     */
    injectDependencies(deps: Record<string, any>): void {
        console.log('WorldViewer: injectDependencies() - Phase 2', Object.keys(deps));
        // WorldViewer doesn't need external dependencies
    }

    /**
     * Phase 3: Activate component - Initialize Phaser here
     */
    async activate(): Promise<void> {
        console.log('WorldViewer: activate() - Phase 3 - Initializing Phaser');
        
        // Now initialize Phaser in the proper lifecycle phase
        this.initializePhaserViewer();
        
        console.log('WorldViewer: activation complete');
    }

    /**
     * Cleanup phase (called by lifecycle controller if needed)
     */
    deactivate(): void {
        console.log('WorldViewer: deactivate() - cleanup');
        this.destroyComponent();
    }
}
