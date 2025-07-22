import { BasePage } from './BasePage';
import { EventBus, EventTypes } from './EventBus';
import { WorldViewer } from './WorldViewer';
import { WorldStatsPanel } from './WorldStatsPanel';
import { World } from './World';

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
class WorldDetailsPage extends BasePage {
    private currentWorldId: string | null;
    private isLoadingWorld: boolean = false;
    private world: World | null = null;
    
    // Component instances
    private worldViewer: WorldViewer | null = null;
    private worldStatsPanel: WorldStatsPanel | null = null;

    constructor() {
        super();
        this.loadInitialState();
        this.initializeSpecificComponents();
        this.bindSpecificEvents();
    }

    protected initializeSpecificComponents(): void {
        // Initialize components immediately
        this.initializeComponents();
    }
    
    /**
     * Initialize page components using the new simplified component architecture
     */
    private initializeComponents(): void {
        try {
            console.log('Initializing WorldDetailsPage components');
            
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
            
            // Create WorldViewer component
            const worldViewerRoot = this.ensureElement('[data-component="world-viewer"]', 'world-viewer-root');
            console.log('WorldDetailsPage: Creating WorldViewer with eventBus:', this.eventBus);
            this.worldViewer = new WorldViewer(worldViewerRoot, this.eventBus, true);
            
            // Create WorldStatsPanel component  
            const worldStatsRoot = this.ensureElement('[data-component="world-stats-panel"]', 'world-stats-root');
            this.worldStatsPanel = new WorldStatsPanel(worldStatsRoot, this.eventBus, true);
            
            console.log('WorldDetailsPage components initialized');
            
        } catch (error) {
            console.error('Failed to initialize components:', error);
            this.showToast('Error', 'Failed to initialize page components', 'error');
        }
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
        return element;
    }
    

    protected bindSpecificEvents(): void {
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
     * Load world data from the hidden JSON element in the page
     */
    private loadWorldDataFromElement(): any {
        try {
            const worldDataElement = document.getElementById('world-data-json');
            console.log(`World data element found: ${worldDataElement ? 'YES' : 'NO'}`);
            
            if (worldDataElement && worldDataElement.textContent) {
                console.log(`Raw world data content: ${worldDataElement.textContent.substring(0, 200)}...`);
                const worldData = JSON.parse(worldDataElement.textContent);
                
                if (worldData && worldData !== null) {
                    console.log('World data found in page element');
                    console.log(`World data keys: ${Object.keys(worldData).join(', ')}`);
                    if (worldData.tiles) {
                        console.log(`Tiles data keys: ${Object.keys(worldData.tiles).join(', ')}`);
                    }
                    if (worldData.world_units) {
                        console.log(`Units data length: ${worldData.world_units.length}`);
                    }
                    return worldData;
                }
            }
            console.log('No world data found in page element');
            return null;
        } catch (error) {
            console.error('Error parsing world data from page element:', error);
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
}

document.addEventListener('DOMContentLoaded', () => {
    const lc = new WorldDetailsPage();
});
