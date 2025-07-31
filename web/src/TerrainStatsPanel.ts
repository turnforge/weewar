import { BaseComponent } from '../lib/Component';
import { EventBus } from '../lib/EventBus';
import { LCMComponent } from '../lib/LCMComponent';
import { TERRAIN_NAMES } from './ColorsAndNames';
import { TerrainStats , RulesTable } from './RulesTable';

interface UnitData {
    ID: number;
    Name: string;
    MovementPoints: number;
    AttackRange: number;
    Health: number;
    Properties: string[];
}

/**
 * TerrainStatsPanel displays detailed information about a selected terrain tile
 * 
 * This component shows:
 * - Terrain type and visual representation from rules engine
 * - Movement costs for different unit types from movement matrix
 * - Defense bonuses from terrain data
 * - Coordinate information
 * - Player ownership (if applicable)
 * 
 * The panel remains hidden until terrain is selected, then displays relevant info.
 * Uses the terrain-stats-panel-template from TerrainStatsPanel.html
 * Gets terrain data from rules engine JSON embedded in page by Go backend
 */
export class TerrainStatsPanel extends BaseComponent implements LCMComponent {
    private isUIBound = false;
    private isActivated = false;
    private currentTerrain: TerrainStats | null = null;
    public rulesTable: RulesTable

    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('terrain-stats-panel', rootElement, eventBus, debugMode);
        this.rulesTable = new RulesTable()
    }

    // LCMComponent Phase 1: Initialize DOM structure
    public performLocalInit(): LCMComponent[] {
        if (this.isUIBound) {
            this.log('Already bound to DOM, skipping');
            return [];
        }

        this.log('Binding TerrainStatsPanel to DOM using template');
        this.isUIBound = true;
        this.log('TerrainStatsPanel bound to DOM successfully');
        
        // This is a leaf component - no children
        return [];
    }

    // Phase 2: No external dependencies needed
    public setupDependencies(): void {
        this.log('TerrainStatsPanel: No dependencies required');
    }

    // Phase 3: Activate component
    public activate(): void {
        if (this.isActivated) {
            this.log('Already activated, skipping');
            return;
        }

        this.log('Activating TerrainStatsPanel');
        this.isActivated = true;
        this.log('TerrainStatsPanel activated successfully');
    }

    // Phase 4: Deactivate component
    public deactivate(): void {
        this.log('Deactivating TerrainStatsPanel');
        this.currentTerrain = null;
        this.isActivated = false;
        this.log('TerrainStatsPanel deactivated');
    }

    /**
     * Update the panel with information about a selected terrain tile
     */
    public updateTerrainStats(terrainStats: TerrainStats): void {
        if (!this.isActivated) {
            throw new Error('Component not activated, cannot update terrain info');
        }

        this.currentTerrain = terrainStats;
        this.log('Updating terrain info for tile:', terrainStats);

        // Hide no-selection state and show terrain details
        const noSelectionDiv = this.findElement('#no-terrain-selected');
        const terrainDetailsDiv = this.findElement('#terrain-details');
        
        if (noSelectionDiv) noSelectionDiv.classList.add('hidden');
        if (terrainDetailsDiv) terrainDetailsDiv.classList.remove('hidden');
        
        // Update terrain header information
        this.updateTerrainHeader(terrainStats);
        
        // Update movement cost (use rules engine data if available)
        const movementCost = terrainStats.baseMoveCost;
        this.updateMovementCost(movementCost);
        
        // Update defense bonus (use rules engine data if available)
        const defenseBonus = terrainStats.defenseBonus;
        this.updateDefenseBonus(defenseBonus);
        
        // Update player ownership if applicable
        this.updatePlayerOwnership(terrainStats.player);
        
        // Update terrain properties using rules engine data
        this.updateTerrainProperties(terrainStats);
    }

    /**
     * Clear terrain selection and show empty state
     */
    public clearTerrainStats(): void {
        if (!this.isActivated) {
            return;
        }

        this.currentTerrain = null;
        this.log('Clearing terrain info');

        // Show no-selection state and hide terrain details
        const noSelectionDiv = this.findElement('#no-terrain-selected');
        const terrainDetailsDiv = this.findElement('#terrain-details');
        
        if (noSelectionDiv) noSelectionDiv.classList.remove('hidden');
        if (terrainDetailsDiv) terrainDetailsDiv.classList.add('hidden');
    }

    /**
     * Update the terrain header (icon, name, coordinates, description)
     */
    private updateTerrainHeader(terrainStats: TerrainStats): void {
        const iconElement = this.findElement('#terrain-icon');
        const nameElement = this.findElement('#terrain-name');
        const coordsElement = this.findElement('#terrain-coordinates');
        const descElement = this.findElement('#terrain-description');

        if (iconElement) {
            const terrainData = TERRAIN_NAMES[terrainStats.id] || { icon: '🎨' };
            iconElement.textContent = terrainData.icon;
        }

        if (nameElement) {
            // Use rules engine name if available, fallback to terrainStats name
            const displayName = terrainStats.name;
            nameElement.textContent = displayName;
        }

        if (coordsElement) {
            coordsElement.textContent = `(${terrainStats.q}, ${terrainStats.r})`;
        }

        if (descElement) {
            // Use rules engine description if available, fallback to terrainStats description
            const description = terrainStats.description;
            descElement.textContent = description;
        }
    }

    /**
     * Update the movement cost display
     */
    private updateMovementCost(cost: number): void {
        const costElement = this.findElement('#movement-cost');
        if (costElement) {
            costElement.textContent = cost.toFixed(1);
        }
    }

    /**
     * Update the defense bonus display
     */
    private updateDefenseBonus(bonus: number): void {
        const bonusElement = this.findElement('#defense-bonus');
        if (bonusElement) {
            const sign = bonus >= 0 ? '+' : '';
            bonusElement.textContent = `${sign}${(bonus * 100).toFixed(0)}%`;
        }
    }

    /**
     * Update player ownership display
     */
    private updatePlayerOwnership(player?: number): void {
        const ownershipDiv = this.findElement('#player-ownership');
        const playerElement = this.findElement('#owner-player');

        if (player !== undefined && player > 0) {
            if (ownershipDiv) ownershipDiv.classList.remove('hidden');
            if (playerElement) playerElement.textContent = `Player ${player}`;
        } else {
            if (ownershipDiv) ownershipDiv.classList.add('hidden');
        }
    }

    /**
     * Update terrain properties list using rules engine data
     */
    private updateTerrainProperties(terrainStats: TerrainStats): void {
        const propertiesList = this.findElement('#properties-list');
        if (!propertiesList) return;

        const properties: Array<{name: string, value: string}> = [];

        // Add basic properties
        properties.push({
            name: 'Type ID',
            value: terrainStats.id.toString()
        });

        properties.push({
            name: 'Hex Coordinate',
            value: `Q:${terrainStats.q}, R:${terrainStats.r}`
        });

        // Add rules engine data if available
        if (terrainStats) {
            properties.push({
                name: 'Base Move Cost',
                value: terrainStats.baseMoveCost.toFixed(1)
            });

            if (terrainStats.defenseBonus !== 0) {
                const sign = terrainStats.defenseBonus >= 0 ? '+' : '';
                properties.push({
                    name: 'Defense Bonus',
                    value: `${sign}${(terrainStats.defenseBonus * 100).toFixed(0)}%`
                });
            }

            properties.push({
                name: 'Terrain Type',
                value: terrainStats.type.toString()
            });
        }

        // Add movement costs for different unit types if available
        const mm = this.rulesTable.movementMatrix
        if (mm && mm.costs) {
            const unitMovements: string[] = [];
            
            // Show movement costs for first few unit types as examples
            Object.entries(mm.costs).slice(0, 3).forEach(([unitId, terrainCosts]) => {
                const cost = (terrainCosts as any)[terrainStats.id];
                if (cost !== undefined) {
                    const unitDef = this.rulesTable.getUnitDefinition(parseInt(unitId));
                    const unitName = unitDef?.name || `Unit ${unitId}`;
                    unitMovements.push(`${unitName}: ${cost.toFixed(1)}`);
                }
            });
            
            if (unitMovements.length > 0) {
                properties.push({
                    name: 'Unit Movement Costs',
                    value: unitMovements.join(', ')
                });
            }
        }

        // Generate HTML
        let propertiesHTML = '';
        properties.forEach(property => {
            propertiesHTML += `
                <div class="text-sm text-gray-600 dark:text-gray-300">
                    <span class="font-medium">${property.name}:</span> ${property.value}
                </div>
            `;
        });

        propertiesList.innerHTML = propertiesHTML || 
            '<div class="text-sm text-gray-500 dark:text-gray-400 italic">No properties available</div>';
    }

    /**
     * Get current terrain info (for external access)
     */
    public getCurrentTerrain(): TerrainStats | null {
        return this.currentTerrain;
    }

    /**
     * Check if terrain is currently selected
     */
    public hasTerrainSelected(): boolean {
        return this.currentTerrain !== null;
    }

    /**
     * Get terrain data from rules engine (for external access)
     */
    public getTerrainData(tileType: number): TerrainStats | null {
        return this.rulesTable.getTerrainStatsAt(tileType, 0);
    }

    protected destroyComponent(): void {
        this.deactivate();
    }
}
