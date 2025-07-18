/**
 * TileStatsPanel displays statistics about tiles and units on the map
 */
export class TileStatsPanel {
    private containerElement: HTMLElement | null = null;
    private isInitialized: boolean = false;
    
    constructor() {
        // Constructor kept minimal - initialize() must be called separately
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
    }
    
    /**
     * Update the stats display with current map data
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
     * Cleanup
     */
    public destroy(): void {
        if (this.containerElement) {
            this.containerElement.innerHTML = '';
        }
        this.containerElement = null;
        this.isInitialized = false;
    }
}