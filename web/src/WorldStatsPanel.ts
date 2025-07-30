import { BaseComponent } from '../lib/Component';
import { EventBus } from '../lib/EventBus';
import { LCMComponent } from '../lib/LCMComponent';
import { WorldEventTypes, WorldDataLoadedPayload, WorldStatsUpdatedPayload } from './events';
import { TERRAIN_NAMES } from './ColorsAndNames';

/**
 * WorldStatsPanel Component - Manages world statistics display
 * Responsible for:
 * - Displaying world statistics (tiles, units, terrain distribution)
 * - Updating display when world data changes
 * - Managing basic info and calculated metrics
 * 
 * Layout and styling are handled by parent container and CSS classes.
 * Uses the worldstats-panel-template from WorldEditorPage.html
 */
export class WorldStatsPanel extends BaseComponent implements LCMComponent {
    private statsData: WorldStatsUpdatedPayload | null = null;
    private isUIBound = false;
    private isActivated = false;
    
    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('world-stats-panel', rootElement, eventBus, debugMode);
    }
    
    // LCMComponent Phase 1: Initialize DOM structure
    public performLocalInit(): LCMComponent[] {
        if (this.isUIBound) {
            this.log('Already bound to DOM, skipping');
            return [];
        }

        this.log('Binding WorldStatsPanel to DOM');
        this.validateDOMStructure();
        this.isUIBound = true;
        this.log('WorldStatsPanel bound to DOM successfully');
        
        // This is a leaf component - no children
        return [];
    }

    // Phase 2: No external dependencies needed
    public setupDependencies(): void {
        this.log('WorldStatsPanel: No dependencies required');
    }

    // Phase 3: Activate component - Subscribe to events here
    public activate(): void {
        if (this.isActivated) {
            this.log('Already activated, skipping');
            return;
        }

        this.log('Activating WorldStatsPanel');
        
        // Subscribe to world data events
        this.subscribe<WorldDataLoadedPayload>(WorldEventTypes.WORLD_DATA_LOADED, this, (payload) => {
            this.handleWorldDataLoaded(payload.data);
        });
        
        this.subscribe<WorldStatsUpdatedPayload>(WorldEventTypes.WORLD_STATS_UPDATED, this, (payload) => {
            this.handleStatsUpdated(payload.data);
        });
        
        this.isActivated = true;
        this.log('WorldStatsPanel activated successfully');
    }

    // Phase 4: Deactivate component
    public deactivate(): void {
        this.log('Deactivating WorldStatsPanel');
        this.statsData = null;
        this.isActivated = false;
        this.log('WorldStatsPanel deactivated');
    }

    /**
     * Validate that the required DOM structure exists within rootElement
     */
    private validateDOMStructure(): void {
        // Check for required sections within rootElement only
        const basicSection = this.findElement('[data-stat-section="basic"]');
        const terrainSection = this.findElement('[data-stat-section="terrain"]');
        
        if (!basicSection) {
            throw new Error('WorldStatsPanel: [data-stat-section="basic"] element not found within rootElement');
        }
        
        if (!terrainSection) {
            throw new Error('WorldStatsPanel: [data-stat-section="terrain"] element not found within rootElement');
        }
        
        this.log('DOM structure validated successfully');
    }
    
    protected destroyComponent(): void {
        this.deactivate();
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
                name: TERRAIN_NAMES[terrainNum].name
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
        this.emit(WorldEventTypes.WORLD_STATS_UPDATED, this.statsData, this, this);
    }
    
    /**
     * Update the display with current stats data
     */
    private updateDisplay(): void {
        if (!this.statsData) return;
        
        this.updateBasicStats();
        this.updateTerrainStats();
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
