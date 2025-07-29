import { BasePage } from '../lib/BasePage';
import { EventBus } from '../lib/EventBus';
import { WorldViewer } from './WorldViewer';
import { WorldStatsPanel } from './WorldStatsPanel';
import { World } from './World';
import { LCMComponent } from '../lib/LCMComponent';
import { LifecycleController } from '../lib/LifecycleController';

/**
 * World Details Page - Orchestrator for world viewing functionality
 * Responsible for:
 * - Data loading and coordination
 * - Component initialization and management
 * - Page-level event coordination
 * - Navigation and user actions
 * 
 * Does NOT handle:
 * - Direct DOM manipulation (delegated to components)
 * - Phaser management (delegated to WorldViewer)
 * - Statistics display (delegated to WorldStatsPanel)
 */
class WorldDetailsPage extends BasePage implements LCMComponent {
    private currentWorldId: string | null;
    private isLoadingWorld: boolean = false;
    private world: World | null = null;
    
    // Component instances
    private worldViewer: WorldViewer | null = null;
    private worldStatsPanel: WorldStatsPanel | null = null;

    /**
     * Load game configuration from hidden inputs (required by BasePage)
     * This method is called by BasePage constructor, but we're using external LifecycleController
     * so we make this a no-op and handle initialization through LCMComponent interface
     */
    protected initializeSpecificComponents(): void {
        console.log('WorldDetailsPage: initializeSpecificComponents() called by BasePage - doing minimal setup');
        this.loadInitialState(); // Load initial state here since constructor calls this
        console.log('WorldDetailsPage: Actual component initialization will be handled by LifecycleController');
    }
    
    /**
     * Subscribe to WorldViewer events before component creation
     */
    private subscribeToWorldViewerEvents(): void {
        // Subscribe to WorldViewer ready event BEFORE creating the component
        console.log('WorldDetailsPage: Subscribing to world-viewer-ready event');
        this.eventBus.subscribe('world-viewer-ready', () => {
            console.log('WorldDetailsPage: WorldViewer is ready, loading world data...');
            if (this.currentWorldId) {
              // Give Phaser time to fully initialize webgl context and scene
              setTimeout(async () => {
                await this.loadWorldData()
              }, 10)
            }
        }, 'world-details-page');
    }

    /**
     * Create WorldViewer and WorldStatsPanel component instances
     */
    private createComponents(): void {
        // Create WorldViewer component
        const worldViewerRoot = this.ensureElement('[data-component="world-viewer"]', 'world-viewer-root');
        console.log('WorldDetailsPage: Creating WorldViewer with eventBus:', this.eventBus);
        this.worldViewer = new WorldViewer(worldViewerRoot, this.eventBus, true);
        
        // Create WorldStatsPanel component  
        const worldStatsRoot = this.ensureElement('[data-component="world-stats-panel"]', 'world-stats-root');
        this.worldStatsPanel = new WorldStatsPanel(worldStatsRoot, this.eventBus, true);
        
        console.log('WorldDetailsPage: Components created');
    }
    
    /**
     * Ensure an element exists, create if missing
     * This is acceptable for page-level orchestration to find component root elements
     */
    private ensureElement(selector: string, fallbackId: string): HTMLElement {
        let element = document.querySelector(selector) as HTMLElement;
        if (!element) {
            console.warn(`Element not found: ${selector}, creating fallback`);
            element = document.createElement('div');
            element.id = fallbackId;
            element.className = 'w-full h-full';
            // Fallback should be more specific than just body
            const mainContainer = document.querySelector('main') || document.body;
            mainContainer.appendChild(element);
        }
        
        // Ensure element has an ID for Phaser container
        if (!element.id) {
            element.id = fallbackId;
        }
        
        return element;
    }
    

    /**
     * Bind page-specific events (required by BasePage)
     * This method is called by BasePage constructor, but we're using external LifecycleController
     * so we make this a no-op and handle event binding in LCMComponent.activate()
     */
    protected bindSpecificEvents(): void {
        console.log('WorldDetailsPage: bindSpecificEvents() called by BasePage - deferred to activate() phase');
    }

    /**
     * Internal method to bind page-specific events (called from activate() phase)
     */
    private bindPageSpecificEvents(): void {
        const mobileMenuButton = document.getElementById('mobile-menu-button');
        if (mobileMenuButton) {
            mobileMenuButton.addEventListener('click', () => {
              // Do things like sidebar drawers etc
            });
        }

        // Bind copy world button if it exists
        const copyButton = document.querySelector('[data-action="copy-world"]');
        if (copyButton) {
            copyButton.addEventListener('click', this.copyWorld.bind(this));
        }
    }

    /** Load document data and set initial UI states */
    private loadInitialState(): void {
        // Theme button state is handled by BasePage

        const worldIdInput = document.getElementById("worldIdInput") as HTMLInputElement | null;
        const worldId = worldIdInput?.value.trim() || null;

        if (worldId) {
            this.currentWorldId = worldId;
            console.log(`Found World ID: ${this.currentWorldId}. Will load data after Phaser initialization.`);
        } else {
            console.error("World ID input element not found or has no value. Cannot load document.");
            this.showToast("Error", "Could not load document: World ID missing.", "error");
        }
    }

    /**
     * Load world data and coordinate between components
     */
    private async loadWorldData(): Promise<void> {
        try {
            console.log(`WorldDetailsPage: Loading world data...`);
            
            // Load world data from the hidden JSON element
            const worldData = this.loadWorldDataFromElement();
            
            if (worldData) {
                this.world = World.deserialize(worldData);
                console.log('World data loaded successfully');
                
                // Use WorldViewer component to load the world
                if (this.worldViewer) {
                    await this.worldViewer.loadWorld(worldData);
                    this.showToast('Success', 'World loaded successfully', 'success');
                } else {
                    console.warn('WorldViewer component not available');
                }
                
            } else {
                console.error('No world data found');
                this.showToast('Error', 'No world data found', 'error');
            }
            
        } catch (error) {
            console.error('Failed to load world data:', error);
            this.showToast('Error', 'Failed to load world data', 'error');
        }
    }
    
    /**
     * Load world data from the hidden JSON elements in the page
     * Now loads from both world metadata and world tiles/units data
     */
    private loadWorldDataFromElement(): any {
        try {
            // Load world metadata
            const worldMetadataElement = document.getElementById('world-data-json');
            const worldTilesElement = document.getElementById('world-tiles-data-json');
            
            console.log(`World metadata element found: ${worldMetadataElement ? 'YES' : 'NO'}`);
            console.log(`World tiles element found: ${worldTilesElement ? 'YES' : 'NO'}`);
            
            if (!worldMetadataElement || !worldTilesElement) {
                console.log('Missing required world data elements');
                return null;
            }
            
            // Parse world metadata
            let worldMetadata = null;
            if (worldMetadataElement.textContent) {
                console.log(`Raw world metadata: ${worldMetadataElement.textContent.substring(0, 200)}...`);
                worldMetadata = JSON.parse(worldMetadataElement.textContent);
            }
            
            // Parse world tiles/units data
            let worldTilesData = null;
            if (worldTilesElement.textContent) {
                console.log(`Raw world tiles data: ${worldTilesElement.textContent.substring(0, 200)}...`);
                worldTilesData = JSON.parse(worldTilesElement.textContent);
            }
            
            if (worldMetadata && worldTilesData) {
                // Combine into format expected by World.loadFromData()
                const combinedData = {
                    // World metadata
                    name: worldMetadata.name || 'Untitled World',
                    Name: worldMetadata.name || 'Untitled World', // Both for compatibility
                    id: worldMetadata.id,
                    
                    // Calculate dimensions from tiles if present
                    width: 40,  // Default
                    height: 40, // Default
                    
                    // World tiles and units
                    tiles: worldTilesData.tiles || [],
                    units: worldTilesData.units || []
                };
                
                // Calculate actual dimensions from tile bounds
                if (combinedData.tiles && combinedData.tiles.length > 0) {
                    let maxQ = 0, maxR = 0, minQ = 0, minR = 0;
                    combinedData.tiles.forEach((tile: any) => {
                        if (tile.q > maxQ) maxQ = tile.q;
                        if (tile.q < minQ) minQ = tile.q;
                        if (tile.r > maxR) maxR = tile.r;
                        if (tile.r < minR) minR = tile.r;
                    });
                    combinedData.width = maxQ - minQ + 1;
                    combinedData.height = maxR - minR + 1;
                }
                
                console.log('Combined world data created');
                console.log(`World: ${combinedData.name}, Tiles: ${combinedData.tiles.length}, Units: ${combinedData.units.length}`);
                console.log(`Dimensions: ${combinedData.width}x${combinedData.height}`);
                
                return combinedData;
            }
            
            console.log('No valid world data found in page elements');
            return null;
        } catch (error) {
            console.error('Error parsing world data from page elements:', error);
            return null;
        }
    }
    

    // Theme management is handled by BasePage

    /** Copy world functionality */
    private copyWorld(): void {
        if (!this.currentWorldId) {
            this.showToast('Error', 'No world ID available for copying', 'error');
            return;
        }
        
        // Navigate to editor page with copy mode
        const copyUrl = `/worlds/new?copy=${this.currentWorldId}`;
        window.location.href = copyUrl;
    }

    public destroy(): void {
        // Clean up components
        if (this.worldViewer) {
            this.worldViewer.destroy();
            this.worldViewer = null;
        }
        
        if (this.worldStatsPanel) {
            this.worldStatsPanel.destroy();
            this.worldStatsPanel = null;
        }
        
        // Clean up world data
        this.world = null;
        this.currentWorldId = null;
    }

    // =============================================================================
    // LCMComponent Interface Implementation
    // =============================================================================

    /**
     * Phase 1: Initialize DOM and discover child components
     */
    performLocalInit(): LCMComponent[] {
        console.log('WorldDetailsPage: performLocalInit() - Phase 1');
        
        // Subscribe to events BEFORE creating components
        this.subscribeToWorldViewerEvents();
        
        // Create child components
        this.createComponents();
        
        console.log('WorldDetailsPage: DOM initialized, returning child components');
        
        // Return child components for lifecycle management
        const childComponents: LCMComponent[] = [];
        if (this.worldViewer) {
            childComponents.push(this.worldViewer);
        }
        if (this.worldStatsPanel && (this.worldStatsPanel as any).performLocalInit) {
            childComponents.push(this.worldStatsPanel as any);
        }
        return childComponents;
    }

    /**
     * Phase 2: Inject dependencies (none needed for WorldDetailsPage)
     */
    setupDependencies(): void {
        console.log('WorldDetailsPage: setupDependencies() - Phase 2');
        // WorldDetailsPage doesn't need external dependencies
    }

    /**
     * Phase 3: Activate component when all dependencies are ready
     */
    async activate(): Promise<void> {
        console.log('WorldDetailsPage: activate() - Phase 3');
        
        // Bind events now that all components are ready
        this.bindPageSpecificEvents();
        
        console.log('WorldDetailsPage: activation complete');
    }

    /**
     * Cleanup phase (called by lifecycle controller if needed)
     */
    deactivate(): void {
        console.log('WorldDetailsPage: deactivate() - cleanup');
        this.destroy();
    }
}

// Initialize page when DOM is ready using LifecycleController
document.addEventListener('DOMContentLoaded', async () => {
    console.log('DOM loaded, starting WorldDetailsPage initialization...');

    // Create page-level event bus
    const eventBus = new EventBus(true); // Enable debug mode
    
    // Create page instance (just basic setup)
    const worldDetailsPage = new WorldDetailsPage(eventBus);
    
    // Create lifecycle controller with debug logging
    const lifecycleController = new LifecycleController(eventBus, {
        enableDebugLogging: true,
        phaseTimeoutMs: 15000, // Increased timeout for component loading
        continueOnError: false // Fail fast for debugging
    });
    
    // Start breadth-first initialization
    await lifecycleController.initializeFromRoot(worldDetailsPage);
    
    console.log('WorldDetailsPage fully initialized via LifecycleController');
});
