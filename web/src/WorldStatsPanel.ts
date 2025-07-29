import { BaseComponent } from '../lib/Component';
import { EventBus } from '../lib/EventBus';
import { WorldEventTypes, WorldDataLoadedPayload, WorldStatsUpdatedPayload } from './events';

/**
 * WorldStatsPanel Component - Manages world statistics display
 * Responsible for:
 * - Displaying world statistics (tiles, units, terrain distribution)
 * - Updating display when world data changes
 * - Managing basic info and calculated metrics
 * 
 * Layout and styling are handled by parent container and CSS classes.
 */
export class WorldStatsPanel extends BaseComponent {
    private statsData: WorldStatsUpdatedPayload | null = null;
    
    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('world-stats-panel', rootElement, eventBus, debugMode);
    }
    
    protected initializeComponent(): void {
        this.log('Initializing WorldStatsPanel component');
        
        // Subscribe to world data events
        this.subscribe<WorldDataLoadedPayload>(WorldEventTypes.WORLD_DATA_LOADED, (payload) => {
            this.handleWorldDataLoaded(payload.data);
        });
        
        this.subscribe<WorldStatsUpdatedPayload>(WorldEventTypes.WORLD_STATS_UPDATED, (payload) => {
            this.handleStatsUpdated(payload.data);
        });
        
        this.log('WorldStatsPanel component initialized');
    }
    
    protected bindToDOM(): void {
        try {
            this.log('Binding WorldStatsPanel to DOM');
            
            // Find or create basic stats section
            let basicStats = this.findElement('[data-stat-section="basic"]');
            if (!basicStats) {
                this.createBasicStatsSection();
            }
            
            // Find or create terrain stats section  
            let terrainStats = this.findElement('[data-stat-section="terrain"]');
            if (!terrainStats) {
                this.createTerrainStatsSection();
            }
            
            this.log('WorldStatsPanel bound to DOM');
            
        // Removed try/catch - let errors propagate for easier debugging
    }
    
    protected destroyComponent(): void {
        this.log('Destroying WorldStatsPanel component');
        this.statsData = null;
    }
    
    /**
     * Handle world data loaded event
     */
    private handleWorldDataLoaded(worldData: WorldDataLoadedPayload): void {
        this.log(`Received world data for world: ${worldData.worldId}`);
        
        // Calculate and update statistics
        this.calculateAndUpdateStats(worldData);
    }
    
    /**
     * Handle stats updated event
     */
    private handleStatsUpdated(statsData: WorldStatsUpdatedPayload): void {
        this.log('Received updated statistics');
        this.statsData = statsData;
        this.updateDisplay();
    }
    
    /**
     * Calculate statistics from world data and update display
     */
    private calculateAndUpdateStats(worldData: WorldDataLoadedPayload): void {
        try {
            // Calculate dimensions from bounds
            const width = worldData.bounds ? worldData.bounds.maxQ - worldData.bounds.minQ + 1 : 0;
            const height = worldData.bounds ? worldData.bounds.maxR - worldData.bounds.minR + 1 : 0;
            
            // Calculate terrain distribution
            const terrainDistribution: { [terrainType: number]: { count: number; percentage: number; name: string } } = {};
            
            Object.entries(worldData.terrainCounts).forEach(([terrainType, count]) => {
                const percentage = worldData.totalTiles > 0 ? Math.round((count / worldData.totalTiles) * 100) : 0;
                const terrainNum = parseInt(terrainType);
                
                terrainDistribution[terrainNum] = {
                    count: count,
                    percentage: percentage,
                    name: this.getTerrainName(terrainNum)
                };
            });
            
            // Create stats payload
            this.statsData = {
                totalTiles: worldData.totalTiles,
                totalUnits: worldData.totalUnits,
                dimensions: { width, height },
                terrainDistribution
            };
            
            // Update display
            this.updateDisplay();
            
            // Emit stats updated event for other components
            this.emit(WorldEventTypes.WORLD_STATS_UPDATED, this.statsData);
            
        // Removed try/catch - let errors propagate for easier debugging
    }
    
    /**
     * Update the display with current stats data
     */
    private updateDisplay(): void {
        if (!this.statsData) return;
        
        try {
            this.updateBasicStats();
            this.updateTerrainStats();
            
        // Removed try/catch - let errors propagate for easier debugging
    }
    
    /**
     * Update basic statistics display
     */
    private updateBasicStats(): void {
        if (!this.statsData) return;
        
        // Update total tiles
        const totalTilesElement = this.findElement('[data-stat="total-tiles"]');
        if (totalTilesElement) {
            totalTilesElement.textContent = this.statsData.totalTiles.toString();
        }
        
        // Update dimensions
        const dimensionsElement = this.findElement('[data-stat="dimensions"]');
        if (dimensionsElement) {
            dimensionsElement.textContent = `${this.statsData.dimensions.width} × ${this.statsData.dimensions.height}`;
        }
        
        // Update total units if element exists
        const totalUnitsElement = this.findElement('[data-stat="total-units"]');
        if (totalUnitsElement) {
            totalUnitsElement.textContent = this.statsData.totalUnits.toString();
        }
    }
    
    /**
     * Update terrain statistics display
     */
    private updateTerrainStats(): void {
        if (!this.statsData) return;
        
        // Update terrain distribution
        Object.entries(this.statsData.terrainDistribution).forEach(([terrainType, info]) => {
            const terrainElement = this.findElement(`[data-terrain="${terrainType}"]`);
            if (terrainElement) {
                terrainElement.textContent = `${info.count} (${info.percentage}%)`;
            }
        });
    }
    
    /**
     * Create basic stats section if missing
     */
    private createBasicStatsSection(): void {
        const section = document.createElement('div');
        section.setAttribute('data-stat-section', 'basic');
        section.className = 'mb-6';
        section.innerHTML = `
            <h3 class="text-sm font-medium text-gray-900 dark:text-white mb-2">Basic Info</h3>
            <div class="space-y-2 text-sm">
                <div class="flex justify-between">
                    <span class="text-gray-600 dark:text-gray-400">Dimensions:</span>
                    <span class="text-gray-900 dark:text-white" data-stat="dimensions">0 × 0</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-gray-600 dark:text-gray-400">Total Tiles:</span>
                    <span class="text-gray-900 dark:text-white" data-stat="total-tiles">0</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-gray-600 dark:text-gray-400">Total Units:</span>
                    <span class="text-gray-900 dark:text-white" data-stat="total-units">0</span>
                </div>
            </div>
        `;
        this.rootElement.appendChild(section);
    }
    
    /**
     * Create terrain stats section if missing
     */
    private createTerrainStatsSection(): void {
        const section = document.createElement('div');
        section.setAttribute('data-stat-section', 'terrain');
        section.className = 'mb-6';
        section.innerHTML = `
            <h3 class="text-sm font-medium text-gray-900 dark:text-white mb-2">Terrain Distribution</h3>
            <div class="space-y-2 text-sm">
                <div class="flex justify-between">
                    <span class="text-gray-600 dark:text-gray-400">🌱 Grass:</span>
                    <span class="text-gray-900 dark:text-white" data-terrain="1">0 (0%)</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-gray-600 dark:text-gray-400">🏜️ Desert:</span>
                    <span class="text-gray-900 dark:text-white" data-terrain="2">0 (0%)</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-gray-600 dark:text-gray-400">🌊 Water:</span>
                    <span class="text-gray-900 dark:text-white" data-terrain="3">0 (0%)</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-gray-600 dark:text-gray-400">⛰️ Mountain:</span>
                    <span class="text-gray-900 dark:text-white" data-terrain="16">0 (0%)</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-gray-600 dark:text-gray-400">🗿 Rock:</span>
                    <span class="text-gray-900 dark:text-white" data-terrain="20">0 (0%)</span>
                </div>
            </div>
        `;
        this.rootElement.appendChild(section);
    }
    
    /**
     * Get terrain name and icon by type
     */
    private getTerrainName(terrainType: number): string {
        const terrainNames: { [key: number]: string } = {
            1: '🌱 Grass',
            2: '🏜️ Desert', 
            3: '🌊 Water',
            16: '⛰️ Mountain',
            20: '🗿 Rock'
        };
        
        return terrainNames[terrainType] || `Terrain ${terrainType}`;
    }
    
    /**
     * Public API to manually update stats
     */
    public updateStats(worldData: WorldDataLoadedPayload): void {
        this.calculateAndUpdateStats(worldData);
    }
    
    /**
     * Get current stats data
     */
    public getStatsData(): WorldStatsUpdatedPayload | null {
        return this.statsData;
    }
}
