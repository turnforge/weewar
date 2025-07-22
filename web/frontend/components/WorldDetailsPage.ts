import { BasePage } from './BasePage';
import { EventBus, EventTypes } from './EventBus';
import { MapViewer } from './MapViewer';
import { MapStatsPanel } from './MapStatsPanel';
import { Map } from './Map';

/**
 * Map Details Page - Orchestrator for map viewing functionality
 * Responsible for:
 * - Data loading and coordination
 * - Component initialization and management
 * - Page-level event coordination
 * - Navigation and user actions
 * 
 * Does NOT handle:
 * - Direct DOM manipulation (delegated to components)
 * - Phaser management (delegated to MapViewer)
 * - Statistics display (delegated to MapStatsPanel)
 */
class MapDetailsPage extends BasePage {
    private currentMapId: string | null;
    private isLoadingMap: boolean = false;
    private map: Map | null = null;
    
    // Component instances
    private mapViewer: MapViewer | null = null;
    private mapStatsPanel: MapStatsPanel | null = null;

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
            console.log('Initializing MapDetailsPage components');
            
            // Subscribe to MapViewer ready event BEFORE creating the component
            console.log('MapDetailsPage: Subscribing to map-viewer-ready event');
            this.eventBus.subscribe('map-viewer-ready', () => {
                console.log('MapDetailsPage: MapViewer is ready, loading map data...');
                if (this.currentMapId) {
                  // Give Phaser time to fully initialize webgl context and scene
                  setTimeout(async () => {
                    await this.loadMapData()
                  }, 10)
                }
            }, 'map-details-page');
            
            // Create MapViewer component
            const mapViewerRoot = this.ensureElement('[data-component="map-viewer"]', 'map-viewer-root');
            console.log('MapDetailsPage: Creating MapViewer with eventBus:', this.eventBus);
            this.mapViewer = new MapViewer(mapViewerRoot, this.eventBus, true);
            
            // Create MapStatsPanel component  
            const mapStatsRoot = this.ensureElement('[data-component="map-stats-panel"]', 'map-stats-root');
            this.mapStatsPanel = new MapStatsPanel(mapStatsRoot, this.eventBus, true);
            
            console.log('MapDetailsPage components initialized');
            
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

        // Bind copy map button if it exists
        const copyButton = document.querySelector('[data-action="copy-map"]');
        if (copyButton) {
            copyButton.addEventListener('click', this.copyMap.bind(this));
        }
    }

    /** Load document data and set initial UI states */
    private loadInitialState(): void {
        // Theme button state is handled by BasePage

        const mapIdInput = document.getElementById("mapIdInput") as HTMLInputElement | null;
        const mapId = mapIdInput?.value.trim() || null;

        if (mapId) {
            this.currentMapId = mapId;
            console.log(`Found Map ID: ${this.currentMapId}. Will load data after Phaser initialization.`);
        } else {
            console.error("Map ID input element not found or has no value. Cannot load document.");
            this.showToast("Error", "Could not load document: Map ID missing.", "error");
        }
    }

    /**
     * Load map data and coordinate between components
     */
    private async loadMapData(): Promise<void> {
        try {
            console.log(`MapDetailsPage: Loading map data...`);
            
            // Load map data from the hidden JSON element
            const mapData = this.loadMapDataFromElement();
            
            if (mapData) {
                this.map = Map.deserialize(mapData);
                console.log('Map data loaded successfully');
                
                // Use MapViewer component to load the map
                if (this.mapViewer) {
                    await this.mapViewer.loadMap(mapData);
                    this.showToast('Success', 'Map loaded successfully', 'success');
                } else {
                    console.warn('MapViewer component not available');
                }
                
            } else {
                console.error('No map data found');
                this.showToast('Error', 'No map data found', 'error');
            }
            
        } catch (error) {
            console.error('Failed to load map data:', error);
            this.showToast('Error', 'Failed to load map data', 'error');
        }
    }
    
    /**
     * Load map data from the hidden JSON element in the page
     */
    private loadMapDataFromElement(): any {
        try {
            const mapDataElement = document.getElementById('map-data-json');
            console.log(`Map data element found: ${mapDataElement ? 'YES' : 'NO'}`);
            
            if (mapDataElement && mapDataElement.textContent) {
                console.log(`Raw map data content: ${mapDataElement.textContent.substring(0, 200)}...`);
                const mapData = JSON.parse(mapDataElement.textContent);
                
                if (mapData && mapData !== null) {
                    console.log('Map data found in page element');
                    console.log(`Map data keys: ${Object.keys(mapData).join(', ')}`);
                    if (mapData.tiles) {
                        console.log(`Tiles data keys: ${Object.keys(mapData.tiles).join(', ')}`);
                    }
                    if (mapData.map_units) {
                        console.log(`Units data length: ${mapData.map_units.length}`);
                    }
                    return mapData;
                }
            }
            console.log('No map data found in page element');
            return null;
        } catch (error) {
            console.error('Error parsing map data from page element:', error);
            return null;
        }
    }
    

    // Theme management is handled by BasePage

    /** Copy map functionality */
    private copyMap(): void {
        if (!this.currentMapId) {
            this.showToast('Error', 'No map ID available for copying', 'error');
            return;
        }
        
        // Navigate to editor page with copy mode
        const copyUrl = `/maps/new?copy=${this.currentMapId}`;
        window.location.href = copyUrl;
    }

    public destroy(): void {
        // Clean up components
        if (this.mapViewer) {
            this.mapViewer.destroy();
            this.mapViewer = null;
        }
        
        if (this.mapStatsPanel) {
            this.mapStatsPanel.destroy();
            this.mapStatsPanel = null;
        }
        
        // Clean up map data
        this.map = null;
        this.currentMapId = null;
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const lc = new MapDetailsPage();
});
