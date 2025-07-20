import { Map, MapObserver, MapEvent, MapEventType, TilesChangedEventData, UnitsChangedEventData, MapLoadedEventData } from './Map';

/**
 * TileStatsPanel displays statistics about tiles and units on the map
 * Now reads directly from Map (single source of truth) and observes Map events
 */
export class TileStatsPanel implements MapObserver {
    private containerElement: HTMLElement | null = null;
    private isInitialized: boolean = false;
    private map: Map | null = null;
    
    constructor(map?: Map | null) {
        if (map) {
            this.map = map;
            this.map.subscribe(this);
        }
    }
    
    /**
     * Initialize the TileStats panel with a container element
     */
    public initialize(containerId: string): boolean {
        try {
            this.containerElement = document.getElementById(containerId);
            if (!this.containerElement) {
                throw new Error(`Container element with ID '${containerId}' not found`);
            }
            
            // Create the stats display
            this.createStatsDisplay();
            
            this.isInitialized = true;
            console.log('[TileStatsPanel] Panel initialized successfully');
            
            return true;
            
        } catch (error) {
            console.error(`[TileStatsPanel] Failed to initialize: ${error}`);
            return false;
        }
    }

    /**
     * Initialize the TileStats panel with a container element directly
     */
    public initializeWithElement(containerElement: HTMLElement): boolean {
        try {
            this.containerElement = containerElement;
            if (!this.containerElement) {
                throw new Error('Container element is null or undefined');
            }
            
            // Create the stats display
            this.createStatsDisplay();
            
            this.isInitialized = true;
            console.log('[TileStatsPanel] Panel initialized successfully with element');
            
            return true;
            
        } catch (error) {
            console.error(`[TileStatsPanel] Failed to initialize with element: ${error}`);
            return false;
        }
    }
    
    /**
     * Create the HTML structure for displaying stats
     */
    private createStatsDisplay(): void {
        if (!this.containerElement) return;
        
        this.containerElement.innerHTML = `
            <div class="tile-stats-panel h-full bg-white dark:bg-gray-800 p-4 overflow-y-auto">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">📊 Map Statistics</h3>
                
                <!-- Terrain Stats -->
                <div class="mb-6">
                    <h4 class="text-md font-medium text-gray-700 dark:text-gray-300 mb-3">🌍 Terrain Types</h4>
                    <div id="terrain-stats" class="space-y-2">
                        <!-- Terrain stats will be populated here -->
                    </div>
                    <div class="mt-2 pt-2 border-t border-gray-200 dark:border-gray-600">
                        <div class="text-sm font-medium text-gray-600 dark:text-gray-400">
                            Total Tiles: <span id="total-tiles" class="text-blue-600 dark:text-blue-400">0</span>
                        </div>
                    </div>
                </div>
                
                <!-- Unit Stats -->
                <div class="mb-6">
                    <h4 class="text-md font-medium text-gray-700 dark:text-gray-300 mb-3">🪖 Units</h4>
                    <div id="unit-stats" class="space-y-2">
                        <!-- Unit stats will be populated here -->
                    </div>
                    <div class="mt-2 pt-2 border-t border-gray-200 dark:border-gray-600">
                        <div class="text-sm font-medium text-gray-600 dark:text-gray-400">
                            Total Units: <span id="total-units" class="text-purple-600 dark:text-purple-400">0</span>
                        </div>
                    </div>
                </div>
                
                <!-- Player Stats -->
                <div class="mb-6">
                    <h4 class="text-md font-medium text-gray-700 dark:text-gray-300 mb-3">👥 Player Distribution</h4>
                    <div id="player-stats" class="space-y-2">
                        <!-- Player stats will be populated here -->
                    </div>
                </div>
                
                <!-- Refresh Button -->
                <div class="mt-6">
                    <button 
                        id="refresh-stats-btn" 
                        class="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors text-sm font-medium"
                    >
                        🔄 Refresh Stats
                    </button>
                </div>
            </div>
        `;
        
        // Set up refresh button handler
        this.setupRefreshButton();
    }
    
    /**
     * Set up refresh button event handler
     */
    private setupRefreshButton(): void {
        const refreshBtn = document.getElementById('refresh-stats-btn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                this.refreshStats();
            });
        }
    }
    
    // MapObserver implementation
    public onMapEvent(event: MapEvent): void {
        switch (event.type) {
            case MapEventType.MAP_LOADED:
            case MapEventType.TILES_CHANGED:
            case MapEventType.UNITS_CHANGED:
            case MapEventType.MAP_CLEARED:
                // Auto-refresh stats when map data changes
                this.refreshStats();
                break;
        }
    }
    
    /**
     * Refresh stats by reading current data from Map
     */
    public refreshStats(): void {
        if (!this.isInitialized || !this.map) return;
        
        console.log('[TileStatsPanel] Refreshing stats from Map data');
        
        // Get data directly from Map
        const tilesData = this.map.getAllTiles();
        const unitsData = this.map.getAllUnits();
        
        // Transform to compatible format and update display
        this.updateTerrainStatsFromMap(tilesData);
        this.updateUnitStatsFromMap(unitsData);
        this.updatePlayerStatsFromMap(unitsData);
    }
    
    /**
     * Update the stats display with current map data (legacy method for backward compatibility)
     */
    public updateStats(tilesData: Array<{ q: number; r: number; terrain: number; color: number }>, unitsData: { [key: string]: { unitType: number, playerId: number } }): void {
        if (!this.isInitialized) return;
        
        this.updateTerrainStats(tilesData);
        this.updateUnitStats(unitsData);
        this.updatePlayerStats(unitsData);
    }
    
    /**
     * Update terrain statistics
     */
    private updateTerrainStats(tilesData: Array<{ q: number; r: number; terrain: number; color: number }>): void {
        const terrainContainer = document.getElementById('terrain-stats');
        const totalTilesElement = document.getElementById('total-tiles');
        
        if (!terrainContainer || !totalTilesElement) return;
        
        // Count terrain types
        const terrainCounts: { [key: number]: number } = {};
        tilesData.forEach(tile => {
            terrainCounts[tile.terrain] = (terrainCounts[tile.terrain] || 0) + 1;
        });
        
        // Terrain type names mapping
        const terrainNames: { [key: number]: { name: string, icon: string, color: string } } = {
            1: { name: 'Grass', icon: '🌱', color: 'text-green-600 dark:text-green-400' },
            2: { name: 'Desert', icon: '🏜️', color: 'text-yellow-600 dark:text-yellow-400' },
            3: { name: 'Water', icon: '🌊', color: 'text-blue-600 dark:text-blue-400' },
            4: { name: 'Mountain', icon: '⛰️', color: 'text-gray-600 dark:text-gray-400' },
            5: { name: 'Rock', icon: '🪨', color: 'text-gray-700 dark:text-gray-300' },
            16: { name: 'Missile Silo', icon: '🚀', color: 'text-red-600 dark:text-red-400' },
            20: { name: 'Mines', icon: '⛏️', color: 'text-orange-600 dark:text-orange-400' }
        };
        
        // Generate terrain stats HTML
        let terrainHTML = '';
        Object.entries(terrainCounts).forEach(([terrain, count]) => {
            const terrainNum = parseInt(terrain);
            const terrainInfo = terrainNames[terrainNum] || { name: `Terrain ${terrain}`, icon: '🎨', color: 'text-gray-600 dark:text-gray-400' };
            
            terrainHTML += `
                <div class="flex justify-between items-center py-1">
                    <span class="text-sm text-gray-700 dark:text-gray-300">
                        ${terrainInfo.icon} ${terrainInfo.name}
                    </span>
                    <span class="text-sm font-medium ${terrainInfo.color}">${count}</span>
                </div>
            `;
        });
        
        terrainContainer.innerHTML = terrainHTML || '<div class="text-sm text-gray-500 dark:text-gray-400 italic">No tiles placed</div>';
        totalTilesElement.textContent = tilesData.length.toString();
    }
    
    /**
     * Update unit statistics
     */
    private updateUnitStats(unitsData: { [key: string]: { unitType: number, playerId: number } }): void {
        const unitContainer = document.getElementById('unit-stats');
        const totalUnitsElement = document.getElementById('total-units');
        
        if (!unitContainer || !totalUnitsElement) return;
        
        // Count unit types
        const unitCounts: { [key: number]: number } = {};
        Object.values(unitsData).forEach(unit => {
            unitCounts[unit.unitType] = (unitCounts[unit.unitType] || 0) + 1;
        });
        
        // Unit type names mapping (basic set)
        const unitNames: { [key: number]: { name: string, icon: string } } = {
            1: { name: 'Infantry', icon: '🪖' },
            2: { name: 'Tank', icon: '🛡️' },
            3: { name: 'Artillery', icon: '💥' },
            4: { name: 'Scout', icon: '🔍' },
            5: { name: 'Anti-Air', icon: '🎯' },
            19: { name: 'Rocket Launcher', icon: '🚀' }
        };
        
        // Generate unit stats HTML
        let unitHTML = '';
        Object.entries(unitCounts).forEach(([unitType, count]) => {
            const unitNum = parseInt(unitType);
            const unitInfo = unitNames[unitNum] || { name: `Unit ${unitType}`, icon: '🪖' };
            
            unitHTML += `
                <div class="flex justify-between items-center py-1">
                    <span class="text-sm text-gray-700 dark:text-gray-300">
                        ${unitInfo.icon} ${unitInfo.name}
                    </span>
                    <span class="text-sm font-medium text-purple-600 dark:text-purple-400">${count}</span>
                </div>
            `;
        });
        
        unitContainer.innerHTML = unitHTML || '<div class="text-sm text-gray-500 dark:text-gray-400 italic">No units placed</div>';
        totalUnitsElement.textContent = Object.keys(unitsData).length.toString();
    }
    
    /**
     * Update player statistics
     */
    private updatePlayerStats(unitsData: { [key: string]: { unitType: number, playerId: number } }): void {
        const playerContainer = document.getElementById('player-stats');
        
        if (!playerContainer) return;
        
        // Count units per player
        const playerCounts: { [key: number]: number } = {};
        Object.values(unitsData).forEach(unit => {
            playerCounts[unit.playerId] = (playerCounts[unit.playerId] || 0) + 1;
        });
        
        // Player colors
        const playerColors: { [key: number]: string } = {
            1: 'text-red-600 dark:text-red-400',
            2: 'text-blue-600 dark:text-blue-400',
            3: 'text-green-600 dark:text-green-400',
            4: 'text-yellow-600 dark:text-yellow-400',
            5: 'text-orange-600 dark:text-orange-400',
            6: 'text-purple-600 dark:text-purple-400',
            7: 'text-pink-600 dark:text-pink-400',
            8: 'text-cyan-600 dark:text-cyan-400'
        };
        
        // Generate player stats HTML
        let playerHTML = '';
        Object.entries(playerCounts).forEach(([playerId, count]) => {
            const playerNum = parseInt(playerId);
            const colorClass = playerColors[playerNum] || 'text-gray-600 dark:text-gray-400';
            
            playerHTML += `
                <div class="flex justify-between items-center py-1">
                    <span class="text-sm text-gray-700 dark:text-gray-300">
                        👤 Player ${playerId}
                    </span>
                    <span class="text-sm font-medium ${colorClass}">${count} units</span>
                </div>
            `;
        });
        
        playerContainer.innerHTML = playerHTML || '<div class="text-sm text-gray-500 dark:text-gray-400 italic">No player units</div>';
    }
    
    /**
     * Set up event listeners for the refresh button
     */
    public onRefresh(callback: () => void): void {
        const refreshButton = document.getElementById('refresh-stats-btn');
        if (refreshButton) {
            refreshButton.addEventListener('click', callback);
        }
    }
    
    /**
     * Get initialization status
     */
    public getIsInitialized(): boolean {
        return this.isInitialized;
    }
    
    /**
     * Update terrain statistics from Map format data
     */
    private updateTerrainStatsFromMap(tilesData: Array<{ q: number; r: number; tileType: number; playerId?: number }>): void {
        const terrainContainer = document.getElementById('terrain-stats');
        const totalTilesElement = document.getElementById('total-tiles');
        
        if (!terrainContainer || !totalTilesElement) return;
        
        // Count terrain types
        const terrainCounts: { [key: number]: number } = {};
        tilesData.forEach(tile => {
            terrainCounts[tile.tileType] = (terrainCounts[tile.tileType] || 0) + 1;
        });
        
        // Use same terrain names mapping
        const terrainNames: { [key: number]: { name: string, icon: string, color: string } } = {
            1: { name: 'Grass', icon: '🌱', color: 'text-green-600 dark:text-green-400' },
            2: { name: 'Desert', icon: '🏜️', color: 'text-yellow-600 dark:text-yellow-400' },
            3: { name: 'Water', icon: '🌊', color: 'text-blue-600 dark:text-blue-400' },
            4: { name: 'Mountain', icon: '⛰️', color: 'text-gray-600 dark:text-gray-400' },
            5: { name: 'Rock', icon: '🪨', color: 'text-gray-700 dark:text-gray-300' },
        };
        
        // Generate HTML for terrain stats
        let terrainHTML = '';
        Object.entries(terrainCounts).forEach(([terrainType, count]) => {
            const terrain = terrainNames[parseInt(terrainType)];
            if (terrain) {
                terrainHTML += `
                    <div class="flex justify-between items-center py-1">
                        <span class="text-sm text-gray-700 dark:text-gray-300">
                            ${terrain.icon} ${terrain.name}
                        </span>
                        <span class="text-sm font-medium ${terrain.color}">${count} tiles</span>
                    </div>
                `;
            }
        });
        
        terrainContainer.innerHTML = terrainHTML || '<div class="text-sm text-gray-500 dark:text-gray-400 italic">No terrain data</div>';
        totalTilesElement.textContent = tilesData.length.toString();
    }
    
    /**
     * Update unit statistics from Map format data
     */
    private updateUnitStatsFromMap(unitsData: Array<{ q: number; r: number; unitType: number; playerId: number }>): void {
        const unitContainer = document.getElementById('unit-stats');
        const totalUnitsElement = document.getElementById('total-units');
        
        if (!unitContainer || !totalUnitsElement) return;
        
        // Count unit types
        const unitCounts: { [key: number]: number } = {};
        unitsData.forEach(unit => {
            unitCounts[unit.unitType] = (unitCounts[unit.unitType] || 0) + 1;
        });
        
        // Unit type names (simplified mapping)
        const unitNames: { [key: number]: { name: string, icon: string, color: string } } = {
            1: { name: 'Infantry', icon: '🚶', color: 'text-blue-600 dark:text-blue-400' },
            2: { name: 'Tank', icon: '🚗', color: 'text-red-600 dark:text-red-400' },
            3: { name: 'Artillery', icon: '💥', color: 'text-orange-600 dark:text-orange-400' },
            4: { name: 'Air Unit', icon: '✈️', color: 'text-sky-600 dark:text-sky-400' },
            5: { name: 'Naval Unit', icon: '🚢', color: 'text-cyan-600 dark:text-cyan-400' },
        };
        
        // Generate HTML for unit stats
        let unitHTML = '';
        Object.entries(unitCounts).forEach(([unitType, count]) => {
            const unit = unitNames[parseInt(unitType)] || { name: `Unit ${unitType}`, icon: '🪖', color: 'text-gray-600 dark:text-gray-400' };
            unitHTML += `
                <div class="flex justify-between items-center py-1">
                    <span class="text-sm text-gray-700 dark:text-gray-300">
                        ${unit.icon} ${unit.name}
                    </span>
                    <span class="text-sm font-medium ${unit.color}">${count} units</span>
                </div>
            `;
        });
        
        unitContainer.innerHTML = unitHTML || '<div class="text-sm text-gray-500 dark:text-gray-400 italic">No units</div>';
        totalUnitsElement.textContent = unitsData.length.toString();
    }
    
    /**
     * Update player statistics from Map format data
     */
    private updatePlayerStatsFromMap(unitsData: Array<{ q: number; r: number; unitType: number; playerId: number }>): void {
        const playerContainer = document.getElementById('player-stats');
        if (!playerContainer) return;
        
        // Count units by player
        const playerCounts: { [key: number]: number } = {};
        unitsData.forEach(unit => {
            playerCounts[unit.playerId] = (playerCounts[unit.playerId] || 0) + 1;
        });
        
        // Player color mapping
        const playerColors: { [key: number]: string } = {
            1: 'text-blue-600 dark:text-blue-400',
            2: 'text-red-600 dark:text-red-400',
            3: 'text-green-600 dark:text-green-400',
            4: 'text-yellow-600 dark:text-yellow-400',
        };
        
        // Generate HTML for player stats
        let playerHTML = '';
        Object.entries(playerCounts).forEach(([playerId, count]) => {
            const playerNum = parseInt(playerId);
            const colorClass = playerColors[playerNum] || 'text-gray-600 dark:text-gray-400';
            
            playerHTML += `
                <div class="flex justify-between items-center py-1">
                    <span class="text-sm text-gray-700 dark:text-gray-300">
                        👤 Player ${playerId}
                    </span>
                    <span class="text-sm font-medium ${colorClass}">${count} units</span>
                </div>
            `;
        });
        
        playerContainer.innerHTML = playerHTML || '<div class="text-sm text-gray-500 dark:text-gray-400 italic">No player units</div>';
    }
    
    /**
     * Cleanup
     */
    public destroy(): void {
        // Unsubscribe from Map events
        if (this.map) {
            this.map.unsubscribe(this);
        }
        
        if (this.containerElement) {
            this.containerElement.innerHTML = '';
        }
        this.containerElement = null;
        this.isInitialized = false;
    }
}