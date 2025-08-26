import { BaseComponent } from '../lib/Component';
import { EventBus } from '../lib/EventBus';
import { LCMComponent } from '../lib/LCMComponent';
import { TERRAIN_NAMES, UNIT_NAMES } from './ColorsAndNames';
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
        
        // Update movement cost - now calculated from terrain-unit properties
        // For display purposes, show average or use a default unit (unit ID 1 - Soldier)
        const defaultMovementCost = this.rulesTable.getMovementCost(terrainStats.id, 1);
        this.updateMovementCost(defaultMovementCost);
        
        // Defense bonus is now per terrain-unit combination, skip general display
        // this.updateDefenseBonus(0); // Could calculate average if needed
        
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
        
        // Also clear unit info
        this.clearUnitInfo();
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
                value: 'Varies by unit (see table below)'
            });
        }

        // Generate HTML for properties
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
        
        // Add terrain-unit properties table if terrain is selected
        if (terrainStats) {
            this.generateTerrainUnitPropertiesTable(terrainStats.id, propertiesList);
        }
    }

    /**
     * Generate terrain-unit properties table using HTML templates
     */
    private generateTerrainUnitPropertiesTable(terrainId: number, container: HTMLElement): void {
        // Get the table template
        const tableTemplate = document.getElementById('terrain-unit-properties-table-template') as HTMLTemplateElement;
        const rowTemplate = document.getElementById('unit-row-template') as HTMLTemplateElement;
        
        if (!tableTemplate || !rowTemplate) {
            console.warn('Terrain-unit properties table templates not found');
            return;
        }
        
        // Clone the table template
        const tableElement = tableTemplate.content.cloneNode(true) as DocumentFragment;
        const tbody = tableElement.querySelector('tbody');
        
        if (!tbody) {
            console.warn('Table body not found in template');
            return;
        }
        
        // Get all available units (common unit IDs)
        const commonUnitIds = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15];
        let hasAnyUnits = false;
        
        commonUnitIds.forEach(unitId => {
            const unitDef = this.rulesTable.getUnitDefinition(unitId);
            if (unitDef && unitDef.name) {
                // Clone the row template
                const rowElement = rowTemplate.content.cloneNode(true) as DocumentFragment;
                const row = rowElement.querySelector('tr');
                
                if (row) {
                    // Get terrain-unit properties
                    const properties = this.rulesTable.getTerrainUnitProperties(terrainId, unitId);
                    const movementCost = this.rulesTable.getMovementCost(terrainId, unitId);
                    
                    // Fill in the row data
                    const unitNameCell = row.querySelector('[data-unit-name]');
                    const movementCostCell = row.querySelector('[data-movement-cost]');
                    const healingCell = row.querySelector('[data-healing]');
                    const captureCell = row.querySelector('[data-capture]');
                    const buildCell = row.querySelector('[data-build]');
                    
                    if (unitNameCell) unitNameCell.textContent = unitDef.name;
                    if (movementCostCell) movementCostCell.textContent = movementCost.toFixed(1);
                    if (healingCell) healingCell.textContent = properties?.healingBonus && properties.healingBonus > 0 ? `+${properties.healingBonus}` : '-';
                    if (captureCell) captureCell.textContent = properties?.canCapture ? '✓' : '-';
                    if (buildCell) buildCell.textContent = properties?.canBuild ? '✓' : '-';
                    
                    // Add alternating row colors
                    if (tbody.children.length % 2 === 1) {
                        row.classList.add('bg-gray-50', 'dark:bg-gray-700');
                    }
                    
                    tbody.appendChild(rowElement);
                    hasAnyUnits = true;
                }
            }
        });
        
        // Only append the table if we have units to show
        if (hasAnyUnits) {
            container.appendChild(tableElement);
        }
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

    /**
     * Update unit information display when a unit is present on the tile
     */
    public updateUnitInfo(unit: any): void {
        if (!this.isActivated) {
            return;
        }

        this.log('Updating unit info:', unit);

        // Show unit details section
        const unitDetailsDiv = this.findElement('#unit-details');
        if (unitDetailsDiv) unitDetailsDiv.classList.remove('hidden');

        // Update unit header
        this.updateUnitHeader(unit);
        
        // Update unit stats
        this.updateUnitStats(unit);
        
        // Update unit properties
        this.updateUnitProperties(unit);
    }

    /**
     * Clear unit information display
     */
    public clearUnitInfo(): void {
        if (!this.isActivated) {
            return;
        }

        this.log('Clearing unit info');

        // Hide unit details section
        const unitDetailsDiv = this.findElement('#unit-details');
        if (unitDetailsDiv) unitDetailsDiv.classList.add('hidden');
    }

    /**
     * Update unit header (icon, name, player, description)
     */
    private updateUnitHeader(unit: any): void {
        const iconElement = this.findElement('#unit-icon');
        const nameElement = this.findElement('#unit-name');
        const playerElement = this.findElement('#unit-player');
        const descElement = this.findElement('#unit-description');

        if (iconElement) {
            // Use the same image path pattern as Phaser
            const unitType = unit.unitType;
            const color = unit.player || 0;
            const imagePath = `/static/assets/v1/Units/${unitType}/${color}.png`;
            
            // Create an img element instead of using text content
            iconElement.innerHTML = `<img src="${imagePath}" alt="Unit ${unitType}" class="w-8 h-8 object-contain" style="image-rendering: pixelated;" onerror="this.style.display='none'; this.nextSibling.style.display='inline';">
                                     <span style="display:none;">⚔️</span>`;
        }

        if (nameElement) {
            const unitDef = this.rulesTable.getUnitDefinition(unit.unitType);
            const unitName = unitDef?.name || UNIT_NAMES[unit.unitType]?.name || `Unit ${unit.unitType}`;
            nameElement.textContent = unitName;
        }

        if (playerElement) {
            playerElement.textContent = `Player ${unit.player}`;
        }

        if (descElement) {
            const unitDef = this.rulesTable.getUnitDefinition(unit.unitType);
            descElement.textContent = unitDef?.description || 'Military unit';
        }
    }

    /**
     * Update unit stats (health, movement, range, status)
     */
    private updateUnitStats(unit: any): void {
        const healthElement = this.findElement('#unit-health');
        const movementElement = this.findElement('#unit-movement');
        const rangeElement = this.findElement('#unit-range');
        const statusElement = this.findElement('#unit-status');

        if (healthElement) {
            healthElement.textContent = unit.health?.toString() || '100';
        }

        if (movementElement) {
            const unitDef = this.rulesTable.getUnitDefinition(unit.unitType);
            movementElement.textContent = unitDef?.movementPoints?.toString() || unit.movementPoints?.toString() || '3';
        }

        if (rangeElement) {
            const unitDef = this.rulesTable.getUnitDefinition(unit.unitType);
            rangeElement.textContent = unitDef?.attackRange?.toString() || unit.attackRange?.toString() || '1';
        }

        if (statusElement) {
            // Determine status based on unit state
            let status = 'Ready';
            if (unit.hasActed) {
                status = 'Used';
            } else if (unit.health < 50) {
                status = 'Damaged';
            }
            statusElement.textContent = status;
        }
    }

    /**
     * Update unit properties list
     */
    private updateUnitProperties(unit: any): void {
        const propertiesList = this.findElement('#unit-properties-list');
        if (!propertiesList) return;

        const properties: Array<{name: string, value: string}> = [];

        // Add basic unit properties
        properties.push({
            name: 'Unit ID',
            value: unit.id?.toString() || 'N/A'
        });

        properties.push({
            name: 'Unit Type',
            value: unit.unitType?.toString() || 'N/A'
        });

        properties.push({
            name: 'Owner',
            value: `Player ${unit.player}`
        });

        if (unit.health !== undefined) {
            properties.push({
                name: 'Health',
                value: `${unit.health}/100`
            });
        }

        // Add unit definition properties if available
        const unitDef = this.rulesTable.getUnitDefinition(unit.unitType);
        if (unitDef) {
            if (unitDef.movementPoints !== undefined) {
                properties.push({
                    name: 'Max Movement',
                    value: unitDef.movementPoints.toString()
                });
            }

            if (unitDef.attackRange !== undefined) {
                properties.push({
                    name: 'Attack Range',
                    value: unitDef.attackRange.toString()
                });
            }

            if (unitDef.properties && unitDef.properties.length > 0) {
                properties.push({
                    name: 'Special Abilities',
                    value: unitDef.properties.join(', ')
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
            '<div class="text-sm text-gray-500 dark:text-gray-400 italic">No unit properties available</div>';
    }

    protected destroyComponent(): void {
        this.deactivate();
    }
}
