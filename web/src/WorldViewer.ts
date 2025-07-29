import { BaseComponent } from '../lib/Component';
import { EventBus, EventPayload, } from '../lib/EventBus';
import { WorldEventTypes, WorldDataLoadedPayload } from './events';
import { PhaserWorldScene } from './phaser/PhaserWorldScene';
import { PhaserGameScene } from './phaser/PhaserGameScene';
import { Unit, Tile, World } from './World';
import { LCMComponent } from '../lib/LCMComponent';

/**
 * WorldViewer Component - Manages Phaser-based world visualization
 * Responsible for:
 * - Phaser initialization and lifecycle management
 * - World data rendering (tiles and units)
 * - Camera controls and viewport management
 * - Theme and display options
 * 
 * Layout and styling are handled by parent container and CSS classes.
 * 
 * @template TScene - The type of Phaser scene to use (defaults to PhaserWorldScene)
 */
export class WorldViewer<TScene extends PhaserWorldScene = PhaserWorldScene> extends BaseComponent implements LCMComponent {
    protected scene: TScene | null = null;
    private loadedWorldData: WorldDataLoadedPayload | null;
    private viewerContainer: HTMLElement | null;
    
    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        console.log('WorldViewer constructor: received eventBus:', eventBus);
        super('world-viewer', rootElement, eventBus, debugMode);
    }

    /**
     * Factory method for creating the scene - can be overridden by subclasses
     */
    protected createScene(): TScene {
        return new PhaserWorldScene() as TScene;
    }
    
    protected initializeComponent(): void {
        this.log('Initializing WorldViewer component');
        
        // Subscribe to world data events
        this.subscribe<WorldDataLoadedPayload>(WorldEventTypes.WORLD_DATA_LOADED, this, (payload) => {
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
        
        // Clean up Phaser scene (it manages its own game instance)
        if (this.scene) {
            this.scene.destroy();
            this.scene = null;
        }
        
        this.loadedWorldData = null;
        this.viewerContainer = null;
    }
    
    
    // validateDOM method removed - not needed in pure LCMComponent approach
    
    /**
     * Initialize the appropriate Phaser scene (PhaserWorldScene or PhaserGameScene)
     */
    private async initializePhaserScene(): Promise<void> {
        console.log(`WorldViewer: initializePhaserScene() called`);
        
        // Guard against multiple initialization
        if (this.scene) {
            console.log('WorldViewer: Phaser scene already initialized, skipping');
            return;
        }
        
        if (!this.viewerContainer) {
            throw new Error('Viewer container not available');
        }
        
        // Create scene using factory method
        this.log('Creating Phaser scene using factory method');
        this.scene = this.createScene();
        
        // Initialize it with the container
        await this.scene.initialize(this.viewerContainer.id);
        
        this.log(`Phaser scene initialized successfully`);
        
        // Emit ready event
        console.log('WorldViewer: Emitting WORLD_VIEWER_READY event');
        this.emit(WorldEventTypes.WORLD_VIEWER_READY, {
            componentId: this.componentId,
            success: true
        }, this);
        console.log('WorldViewer: WORLD_VIEWER_READY event emitted');
        
        // Load world data if we have it
        if (this.loadedWorldData) {
            await this.loadWorldIntoScene();
        }
    }
    
    /**
     * Handle world data loaded event
     */
    private handleWorldDataLoaded(payload: EventPayload<WorldDataLoadedPayload>): void {
        this.log(`Received world data for world: ${payload.data.worldId}`);
        this.loadedWorldData = payload.data;
        
        // Load into Phaser if scene is ready
        if (this.scene && this.scene.getIsInitialized()) {
            this.loadWorldIntoScene();
        }
    }
    
    /**
     * Load world data into the PhaserWorldScene
     */
    private async loadWorldIntoScene(): Promise<void> {
        if (!this.scene || !this.scene.getIsInitialized()) {
            this.log('Phaser scene not ready, deferring world load');
            return;
        }
        
        if (!this.loadedWorldData) {
            this.log('No world data available to load');
            return;
        }
        
        this.log('Loading world data into Phaser scene');
        
        // This method will be called after loadWorld() sets up the world data properly
        // For now, we need to reconstruct the World from loadedWorldData
        // TODO: This is a bit awkward - we should refactor to avoid this conversion
        console.log('WorldViewer: loadWorldIntoScene called but needs world instance');
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
        if (this.scene && this.scene.getIsInitialized()) {
            await this.scene.loadWorldData(world);
        }
        
        // Emit data loaded event for other components
        this.emit(WorldEventTypes.WORLD_DATA_LOADED, this.loadedWorldData, this);
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
        if (this.scene) {
            this.scene.setShowGrid(show);
        }
    }
    
    public setShowCoordinates(show: boolean): void {
        if (this.scene) {
            this.scene.setShowCoordinates(show);
        }
    }
    
    public setTheme(isDark: boolean): void {
        if (this.scene) {
            this.scene.setTheme(isDark);
        }
    }
    
    /**
     * Camera controls
     */
    public getZoom(): number {
        return this.scene?.getZoom() || 1;
    }
    
    public setZoom(zoom: number): void {
        if (this.scene) {
            this.scene.setZoom(zoom);
        }
    }
    
    /**
     * Resize the viewer
     */
    public resize(width?: number, height?: number): void {
        if (this.scene && this.viewerContainer) {
            const w = width || this.viewerContainer.clientWidth;
            const h = height || this.viewerContainer.clientHeight;
            this.scene.resize(w, h);
        }
    }
    
    /**
     * Check if viewer is ready
     */
    public isPhaserReady(): boolean {
        return this.scene?.getIsInitialized() || false;
    }

    /**
     * Set interaction callbacks for game-specific functionality
     */
    public setInteractionCallbacks(
        tileCallback?: (q: number, r: number) => boolean,
        unitCallback?: (q: number, r: number) => boolean
    ): void {
        console.log('[WorldViewer] setInteractionCallbacks called');
        console.log('[WorldViewer] tileCallback:', !!tileCallback);
        console.log('[WorldViewer] unitCallback:', !!unitCallback);
        console.log('[WorldViewer] this.scene exists:', !!this.scene);
        
        if (this.scene) {
            console.log('[WorldViewer] Calling scene.setInteractionCallbacks');
            this.scene.setInteractionCallbacks(tileCallback, unitCallback);
            console.log('[WorldViewer] scene.setInteractionCallbacks completed');
        } else {
            console.error('[WorldViewer] No scene available to set callbacks on');
        }
    }

    // =============================================================================
    // LCMComponent Interface Implementation
    // =============================================================================

    /**
     * Phase 1: Initialize DOM and discover child components
     */
    performLocalInit(): LCMComponent[] {
        console.log('WorldViewer: performLocalInit() - Phase 1');
        
        // DOM setup is already done in bindToDOM(), just return no child components
        return [];
    }

    /**
     * Phase 2: Inject dependencies (none needed for WorldViewer)
     */
    setupDependencies(): void {
        console.log('WorldViewer: setupDependencies() - Phase 2')
        // WorldViewer doesn't need external dependencies
    }

    /**
     * Phase 3: Activate component - Initialize Phaser here
     */
    async activate(): Promise<void> {
        console.log('WorldViewer: activate() - Phase 3 - Initializing Phaser');
        
        // Now initialize PhaserWorldScene in the proper lifecycle phase
        await this.initializePhaserScene();
        
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
