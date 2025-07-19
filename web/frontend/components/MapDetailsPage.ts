import { BasePage } from './BasePage';
import { PhaserViewer } from './PhaserViewer';
import { Map } from './Map';

/**
 * Map Details page with readonly map viewer
 */
class MapDetailsPage extends BasePage {
    private currentMapId: string | null = null;
    private isLoadingMap: boolean = false;
    private phaserViewer: PhaserViewer | null = null;
    private map: Map | null = null;

    constructor() {
        super();
        this.initializeSpecificComponents();
        this.bindSpecificEvents();
        this.loadInitialState();
    }

    protected initializeSpecificComponents(): void {
        // Initialize Phaser viewer with delay to ensure container is properly sized
        setTimeout(() => {
            this.initializePhaserViewer();
        }, 1000);
    }
    
    private initializePhaserViewer(): void {
        // Initialize Phaser viewer
        this.phaserViewer = new PhaserViewer();
        
        // Set up logging
        this.phaserViewer.onLog((message: string) => {
            console.log(message);
        });
        
        // Initialize the viewer with the container
        const success = this.phaserViewer.initialize('phaser-viewer-container');
        if (!success) {
            console.error('Failed to initialize Phaser viewer');
            this.showToast('Error', 'Failed to initialize map viewer', 'error');
            return;
        }
        
        console.log('MapDetailsPage application initialized with Phaser viewer');
        
        // Now that Phaser is initialized, load the map data
        if (this.currentMapId) {
            this.loadMapData();
        }
    }

    protected bindSpecificEvents(): void {
        const mobileMenuButton = document.getElementById('mobile-menu-button');
        if (mobileMenuButton) {
            mobileMenuButton.addEventListener('click', () => {
              // Do things like sidebar drawers etc
            });
        }

        // Bind copy map button if it exists
        const copyButton = document.querySelector('button:has(svg[d*="M8 16H6"])'); // Copy icon SVG path
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
     * Load map data from the hidden JSON element and display it
     */
    private async loadMapData(): Promise<void> {
        try {
            console.log(`MapDetailsPage: Loading map data...`);
            
            // Load map data from the hidden JSON element (similar to MapEditorPage)
            const mapData = this.loadMapDataFromElement();
            
            if (mapData) {
                this.map = Map.deserialize(mapData);
                console.log('Map data loaded successfully');
                
                // Load the map into the Phaser viewer
                await this.loadMapIntoViewer();
                
                // Update statistics
                this.updateMapStatistics();
                
                this.showToast('Success', 'Map loaded successfully', 'success');
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
    
    /**
     * Load the map data into the Phaser viewer
     */
    private async loadMapIntoViewer(): Promise<void> {
        if (!this.phaserViewer || !this.phaserViewer.getIsInitialized() || !this.map) {
            console.log('Skipping Phaser viewer load - preconditions not met');
            return;
        }
        
        try {
            // Get tiles data
            const allTiles = this.map.getAllTiles();
            const tilesArray: Array<{ q: number; r: number; terrain: number; color: number }> = [];
            
            allTiles.forEach(tile => {
                tilesArray.push({
                    q: tile.q,
                    r: tile.r,
                    terrain: tile.tileType,
                    color: tile.playerId || 0
                });
            });
            
            // Get units data
            const allUnits = this.map.getAllUnits();
            const unitsArray: Array<{ q: number; r: number; unitType: number; playerId: number }> = [];
            
            allUnits.forEach(unit => {
                unitsArray.push({
                    q: unit.q,
                    r: unit.r,
                    unitType: unit.unitType,
                    playerId: unit.playerId
                });
            });
            
            // Load into viewer
            await this.phaserViewer.loadMapData(tilesArray, unitsArray);
            
            console.log(`Loaded ${tilesArray.length} tiles and ${unitsArray.length} units into viewer`);
            
        } catch (error) {
            console.error('Failed to load map into viewer:', error);
            throw error;
        }
    }
    
    /**
     * Update the map statistics sidebar with real data
     */
    private updateMapStatistics(): void {
        if (!this.map) return;
        
        const allTiles = this.map.getAllTiles();
        const allUnits = this.map.getAllUnits();
        
        // Count terrain types
        const terrainCounts: { [key: number]: number } = {};
        allTiles.forEach(tile => {
            terrainCounts[tile.tileType] = (terrainCounts[tile.tileType] || 0) + 1;
        });
        
        // Update basic stats
        this.updateBasicStats(allTiles.length, allUnits.length);
        this.updateTerrainDistribution(terrainCounts, allTiles.length);
    }
    
    /**
     * Update basic statistics in the sidebar
     */
    private updateBasicStats(totalTiles: number, totalUnits: number): void {
        // Update total tiles - use specific selector within the statistics section only
        const statsSection = document.querySelector('.w-80'); // The sidebar with statistics
        if (statsSection) {
            const elements = statsSection.querySelectorAll('.text-gray-900, .text-white');
            elements.forEach(el => {
                if (el.parentElement?.textContent?.includes('Total Tiles:')) {
                    el.textContent = totalTiles.toString();
                }
            });
        }
        
        // Could also update map dimensions if available from bounds
        const bounds = this.map?.getBounds();
        if (bounds && statsSection) {
            const width = bounds.maxQ - bounds.minQ + 1;
            const height = bounds.maxR - bounds.minR + 1;
            
            const dimensionsElements = statsSection.querySelectorAll('.text-gray-900, .text-white');
            dimensionsElements.forEach(el => {
                if (el.parentElement?.textContent?.includes('Dimensions:')) {
                    el.textContent = `${width} × ${height}`;
                }
            });
        }
    }
    
    /**
     * Update terrain distribution in the sidebar
     */
    private updateTerrainDistribution(terrainCounts: { [key: number]: number }, totalTiles: number): void {
        // Terrain type mapping
        const terrainNames: { [key: number]: string } = {
            1: '🌱 Grass',
            2: '🏜️ Desert', 
            3: '🌊 Water',
            16: '⛰️ Mountain',
            20: '🗿 Rock'
        };
        
        // Update terrain counts - scope to statistics sidebar only
        const statsSection = document.querySelector('.w-80'); // The sidebar with statistics
        if (statsSection) {
            Object.entries(terrainNames).forEach(([terrainType, name]) => {
                const count = terrainCounts[parseInt(terrainType)] || 0;
                const percentage = totalTiles > 0 ? Math.round((count / totalTiles) * 100) : 0;
                
                const elements = statsSection.querySelectorAll('.text-gray-900, .text-white');
                elements.forEach(el => {
                    if (el.parentElement?.textContent?.includes(name)) {
                        el.textContent = `${count} (${percentage}%)`;
                    }
                });
            });
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
        // Clean up Phaser viewer
        if (this.phaserViewer) {
            this.phaserViewer.destroy();
            this.phaserViewer = null;
        }
        
        // Clean up map data
        this.map = null;
        this.currentMapId = null;
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const lc = new MapDetailsPage();
});
