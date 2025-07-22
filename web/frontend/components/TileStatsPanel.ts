import { World, WorldObserver, WorldEvent, WorldEventType, TilesChangedEventData, UnitsChangedEventData, WorldLoadedEventData } from './World';
import { BaseComponent } from './Component';
import { EventBus } from './EventBus';
import { ComponentLifecycle } from './ComponentLifecycle';

/**
 * TileStatsPanel displays statistics about tiles and units on the world
 * 
 * This component demonstrates the new lifecycle architecture:
 * 1. initializeDOM() - Set up UI structure without dependencies
 * 2. injectDependencies() - Receive World instance when available
 * 3. activate() - Subscribe to World events and enable functionality
 * 
 * Architecture:
 * - Reads directly from World (single source of truth) via explicit setter
 * - Observes World events for automatic updates
 * - Creates its own DOM structure for statistics display
 * - Handles refresh button and automatic data updates
 */
export class TileStatsPanel extends BaseComponent implements WorldObserver {
    // Dependencies (injected via explicit setters)
    private world: World | null = null;
    
    // Internal state tracking
    private isUIBound = false;
    private isActivated = false;
    private pendingOperations: Array<() => void> = [];
    
    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('tile-stats-panel', rootElement, eventBus, debugMode);
    }
    
    // ComponentLifecycle Phase 1: Initialize DOM and discover children (no dependencies needed)
    public initializeDOM(): ComponentLifecycle[] {
        if (this.isUIBound) {
            this.log('Already bound to DOM, skipping');
            return [];
        }
        
        try {
            this.log('Binding TileStatsPanel to DOM');
            
            // Create the stats display structure
            this.createStatsDisplay();
            this.setupRefreshButton();
            
            this.isUIBound = true;
            this.log('TileStatsPanel bound to DOM successfully');
            
            // This is a leaf component - no children
            return [];
            
        } catch (error) {
            this.handleError('Failed to bind TileStatsPanel to DOM', error);
            throw error;
        }
    }
    
    // Phase 2: Inject dependencies - simplified to use explicit setters
    public injectDependencies(deps: Record<string, any>): void {
        this.log('TileStatsPanel: Dependencies injection phase - using explicit setters');
        
        // Dependencies should be set directly by parent using setters
        // This phase just validates that required dependencies are available
        if (!this.world) {
            throw new Error('TileStatsPanel requires world - use setWorld()');
        }
        
        this.log('Dependencies validation complete');
    }
    
    // Explicit dependency setters
    public setWorld(world: World): void {
        this.world = world;
        this.log('World set via explicit setter');
        
        // Subscribe to world events immediately when world is set
        if (this.world) {
            this.world.subscribe(this);
        }
    }
    
    // Explicit dependency getters
    public getWorld(): World | null {
        return this.world;
    }
    
    // Phase 3: Activate component
    public activate(): void {
        if (this.isActivated) {
            this.log('Already activated, skipping');
            return;
        }
        
        this.log('Activating TileStatsPanel');
        
        // Process any operations that were queued during UI binding
        this.processPendingOperations();
        
        // Initial stats refresh
        this.refreshStats();
        
        this.isActivated = true;
        this.log('TileStatsPanel activated successfully');
    }
    
    // Phase 4: Deactivate component
    public deactivate(): void {
        this.log('Deactivating TileStatsPanel');
        
        // Unsubscribe from World events
        if (this.world) {
            this.world.unsubscribe(this);
        }
        
        // Clear any pending operations
        this.pendingOperations = [];
        
        // Reset state
        this.isActivated = false;
        this.world = null;
        
        this.log('TileStatsPanel deactivated');
    }
    
    /**
     * Create the HTML structure for displaying stats
     */
    private createStatsDisplay(): void {
        this.rootElement.innerHTML = `
            <div class="tile-stats-panel h-full bg-white dark:bg-gray-800 p-4 overflow-y-auto">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">📊 World Statistics</h3>
                
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
    
    // Deferred Execution System
    
    /**
     * Execute operation when component is ready, or queue it for later
     */
    private executeWhenReady(operation: () => void): void {
        if (this.isActivated && this.world) {
            // Component is ready - execute immediately
            try {
                operation();
            } catch (error) {
                this.handleError('Operation failed', error);
            }
        } else {
            // Component not ready - queue for later
            this.pendingOperations.push(operation);
            this.log('Component not ready - operation queued');
        }
    }
    
    /**
     * Process all pending operations when component becomes ready
     */
    private processPendingOperations(): void {
        if (this.pendingOperations.length > 0) {
            this.log(`Processing ${this.pendingOperations.length} pending operations`);
            
            const operations = [...this.pendingOperations];
            this.pendingOperations = [];
            
            operations.forEach(operation => {
                try {
                    operation();
                } catch (error) {
                    this.handleError('Pending operation failed', error);
                }
            });
        }
    }
    
    /**
     * Set up refresh button event handler
     */
    private setupRefreshButton(): void {
        const refreshBtn = this.rootElement.querySelector('#refresh-stats-btn') as HTMLButtonElement;
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                this.executeWhenReady(() => this.refreshStats());
            });
            this.log('Refresh button bound');
        } else {
            this.log('Refresh button not found');
        }
    }
    
    // WorldObserver implementation
    public onWorldEvent(event: WorldEvent): void {
        switch (event.type) {
            case WorldEventType.WORLD_LOADED:
            case WorldEventType.TILES_CHANGED:
            case WorldEventType.UNITS_CHANGED:
            case WorldEventType.WORLD_CLEARED:
                // Auto-refresh stats when world data changes
                this.refreshStats();
                break;
        }
    }
    
    /**
     * Refresh stats by reading current data from World
     */
    public refreshStats(): void {
        if (!this.isActivated || !this.world) {
            this.log('Component not ready for stats refresh');
            return;
        }
        
        console.log('[TileStatsPanel] Refreshing stats from World data');
        
        // Get data directly from World
        const tilesData = this.world.getAllTiles();
        const unitsData = this.world.getAllUnits();
        
        // Transform to compatible format and update display
        this.updateTerrainStatsFromWorld(tilesData);
        this.updateUnitStatsFromWorld(unitsData);
        this.updatePlayerStatsFromWorld(unitsData);
    }
    
    /**
     * Update the stats display with current world data (legacy method for backward compatibility)
     */
    public updateStats(tilesData: Array<{ q: number; r: number; terrain: number; color: number }>, unitsData: { [key: string]: { unitType: number, playerId: number } }): void {
        this.executeWhenReady(() => {
            this.updateTerrainStats(tilesData);
            this.updateUnitStats(unitsData);
            this.updatePlayerStats(unitsData);
        });
    }
    
    /**
     * Update terrain statistics
     */
    private updateTerrainStats(tilesData: Array<{ q: number; r: number; terrain: number; color: number }>): void {
        const terrainContainer = this.findElement('#terrain-stats');
        const totalTilesElement = this.findElement('#total-tiles');
        
        if (!terrainContainer || !totalTilesElement) return;
        
        // Count terrain types
        const terrainCounts: { [key: number]: number } = {};
        tilesData.forEach(tile => {
            terrainCounts[tile.terrain] = (terrainCounts[tile.terrain] || 0) + 1;
        });
        
        // Terrain type names worldping
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
        const unitContainer = this.findElement('#unit-stats');
        const totalUnitsElement = this.findElement('#total-units');
        
        if (!unitContainer || !totalUnitsElement) return;
        
        // Count unit types
        const unitCounts: { [key: number]: number } = {};
        Object.values(unitsData).forEach(unit => {
            unitCounts[unit.unitType] = (unitCounts[unit.unitType] || 0) + 1;
        });
        
        // Unit type names worldping (basic set)
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
        const playerContainer = this.findElement('#player-stats');
        
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
     * Set up event listeners for the refresh button (legacy method for backward compatibility)
     */
    public onRefresh(callback: () => void): void {
        const refreshButton = this.findElement('#refresh-stats-btn');
        if (refreshButton) {
            refreshButton.addEventListener('click', callback);
        }
    }
    
    /**
     * Get activation status (legacy method for backward compatibility)
     */
    public getIsInitialized(): boolean {
        return this.isActivated;
    }
    
    /**
     * Update terrain statistics from World format data
     */
    private updateTerrainStatsFromWorld(tilesData: Array<{ q: number; r: number; tileType: number; playerId?: number }>): void {
        const terrainContainer = this.findElement('#terrain-stats');
        const totalTilesElement = this.findElement('#total-tiles');
        
        if (!terrainContainer || !totalTilesElement) return;
        
        // Count terrain types
        const terrainCounts: { [key: number]: number } = {};
        tilesData.forEach(tile => {
            terrainCounts[tile.tileType] = (terrainCounts[tile.tileType] || 0) + 1;
        });
        
        // Use same terrain names worldping
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
     * Update unit statistics from World format data
     */
    private updateUnitStatsFromWorld(unitsData: Array<{ q: number; r: number; unitType: number; playerId: number }>): void {
        const unitContainer = this.findElement('#unit-stats');
        const totalUnitsElement = this.findElement('#total-units');
        
        if (!unitContainer || !totalUnitsElement) return;
        
        // Count unit types
        const unitCounts: { [key: number]: number } = {};
        unitsData.forEach(unit => {
            unitCounts[unit.unitType] = (unitCounts[unit.unitType] || 0) + 1;
        });
        
        // Unit type names (simplified worldping)
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
     * Update player statistics from World format data
     */
    private updatePlayerStatsFromWorld(unitsData: Array<{ q: number; r: number; unitType: number; playerId: number }>): void {
        const playerContainer = this.findElement('#player-stats');
        if (!playerContainer) return;
        
        // Count units by player
        const playerCounts: { [key: number]: number } = {};
        unitsData.forEach(unit => {
            playerCounts[unit.playerId] = (playerCounts[unit.playerId] || 0) + 1;
        });
        
        // Player color worldping
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
     * Cleanup (legacy method for backward compatibility)
     */
    public destroy(): void {
        this.deactivate();
    }
    
    // BaseComponent lifecycle compatibility
    protected initializeComponent(): void {
        // This is handled by the new lifecycle system
        // Keep empty for backward compatibility
    }
    
    protected bindToDOM(): void {
        // This is handled by the new lifecycle system
        // Keep empty for backward compatibility
    }
    
    protected destroyComponent(): void {
        this.deactivate();
    }
}
