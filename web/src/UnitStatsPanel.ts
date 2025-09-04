import { BaseComponent } from '../lib/Component';
import { EventBus } from '../lib/EventBus';
import { LCMComponent } from '../lib/LCMComponent';
import { UNIT_NAMES } from './ColorsAndNames';
import { RulesTable } from './RulesTable';

interface UnitData {
    id?: number;
    unitType: number;
    health?: number;
    player: number;
    movementPoints?: number;
    attackRange?: number;
    hasActed?: boolean;
}

/**
 * UnitStatsPanel displays detailed information about a selected unit
 * 
 * This component shows:
 * - Unit type and visual representation from rules engine
 * - Basic unit stats (health, movement, range, status)
 * - Unit properties and abilities
 * - Unit-terrain movement costs for different terrain types
 * - Unit-unit combat damage distributions
 * 
 * The panel remains hidden until a unit is selected, then displays relevant info.
 * Uses the unit-stats-panel-template from HTML templates
 * Gets unit data from rules engine JSON embedded in page by Go backend
 */
export class UnitStatsPanel extends BaseComponent implements LCMComponent {
    private isUIBound = false;
    private isActivated = false;
    private currentUnit: UnitData | null = null;
    public rulesTable: RulesTable;

    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('unit-stats-panel', rootElement, eventBus, debugMode);
        this.rulesTable = new RulesTable();
    }

    // LCMComponent Phase 1: Initialize DOM structure
    public performLocalInit(): LCMComponent[] {
        if (this.isUIBound) {
            this.log('Already bound to DOM, skipping');
            return [];
        }

        this.log('Binding UnitStatsPanel to DOM using template');
        this.isUIBound = true;
        this.log('UnitStatsPanel bound to DOM successfully');
        
        // This is a leaf component - no children
        return [];
    }

    // Phase 2: No external dependencies needed
    public setupDependencies(): void {
        this.log('UnitStatsPanel: No dependencies required');
    }

    // Phase 3: Activate component
    public activate(): void {
        if (this.isActivated) {
            this.log('Already activated, skipping');
            return;
        }

        this.log('Activating UnitStatsPanel');
        this.isActivated = true;
        this.log('UnitStatsPanel activated successfully');
    }

    // Phase 4: Deactivate component
    public deactivate(): void {
        this.log('Deactivating UnitStatsPanel');
        this.currentUnit = null;
        this.isActivated = false;
        this.log('UnitStatsPanel deactivated');
    }

    /**
     * Update the panel with information about a selected unit
     */
    public updateUnitInfo(unit: UnitData): void {
        if (!this.isActivated) {
            throw new Error('Component not activated, cannot update unit info');
        }

        this.currentUnit = unit;
        this.log('Updating unit info for unit:', unit);

        // Hide no-selection state and show unit details
        const noSelectionDiv = this.findElement('#no-unit-selected');
        const unitDetailsDiv = this.findElement('#unit-details');
        
        if (noSelectionDiv) noSelectionDiv.classList.add('hidden');
        if (unitDetailsDiv) unitDetailsDiv.classList.remove('hidden');
        
        // Update unit information sections
        this.updateUnitHeader(unit);
        this.updateUnitStats(unit);
        this.updateUnitProperties(unit);
        
        // Show unit-terrain properties table
        this.generateUnitTerrainPropertiesTable(unit.unitType);
        
        // Show unit-unit combat table
        this.generateUnitCombatTable(unit.unitType);
    }

    /**
     * Clear unit selection and show empty state
     */
    public clearUnitInfo(): void {
        if (!this.isActivated) {
            return;
        }

        this.currentUnit = null;
        this.log('Clearing unit info');

        // Show no-selection state and hide unit details
        const noSelectionDiv = this.findElement('#no-unit-selected');
        const unitDetailsDiv = this.findElement('#unit-details');
        
        if (noSelectionDiv) noSelectionDiv.classList.remove('hidden');
        if (unitDetailsDiv) unitDetailsDiv.classList.add('hidden');
    }

    /**
     * Update unit header (icon, name, player, description)
     */
    private updateUnitHeader(unit: UnitData): void {
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
    private updateUnitStats(unit: UnitData): void {
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
            } else if (unit.health && unit.health < 50) {
                status = 'Damaged';
            }
            statusElement.textContent = status;
        }
    }

    /**
     * Update unit properties list
     */
    private updateUnitProperties(unit: UnitData): void {
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

    /**
     * Generate unit-terrain movement cost table
     */
    private generateUnitTerrainPropertiesTable(unitId: number): void {
        const container = this.findElement('#unit-terrain-properties');
        if (!container) return;

        // Get the table template
        const tableTemplate = document.getElementById('unit-terrain-properties-table-template') as HTMLTemplateElement;
        const rowTemplate = document.getElementById('terrain-row-template') as HTMLTemplateElement;
        
        if (!tableTemplate || !rowTemplate) {
            console.warn('Unit-terrain properties table templates not found');
            return;
        }
        
        // Clear existing content
        container.innerHTML = '';
        
        // Clone the table template
        const tableElement = tableTemplate.content.cloneNode(true) as DocumentFragment;
        const tbody = tableElement.querySelector('tbody');
        
        if (!tbody) {
            console.warn('Table body not found in template');
            return;
        }
        
        // Get all available terrains
        const commonTerrainIds = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26];
        let hasAnyTerrain = false;
        
        commonTerrainIds.forEach(terrainId => {
            const terrainDef = this.rulesTable.getTerrainDefinition(terrainId);
            const properties = this.rulesTable.getTerrainUnitProperties(terrainId, unitId);
            if (properties && terrainDef && terrainDef.name) {
                // Clone the row template
                const rowElement = rowTemplate.content.cloneNode(true) as DocumentFragment;
                const row = rowElement.querySelector('tr');
                
                if (row) {
                    // Get terrain-unit properties
                    const movementCost = this.rulesTable.getMovementCost(terrainId, unitId);
                    
                    // Fill in the row data
                    const terrainNameCell = row.querySelector('[data-terrain-name]');
                    const movementCostCell = row.querySelector('[data-movement-cost]');
                    const attackCell = row.querySelector('[data-attack]');
                    const defenseCell = row.querySelector('[data-defense]');
                    const healingCell = row.querySelector('[data-healing]');
                    const captureCell = row.querySelector('[data-capture]');
                    const buildCell = row.querySelector('[data-build]');
                    
                    if (terrainNameCell) terrainNameCell.textContent = terrainDef.name;
                    if (movementCostCell) {
                        if (movementCost >= 999) {
                            movementCostCell.textContent = 'Impassable';
                            movementCostCell.classList.add('text-red-600', 'dark:text-red-400');
                        } else {
                            movementCostCell.textContent = movementCost.toFixed(2);
                        }
                    }
                    if (attackCell) attackCell.textContent = properties?.attackBonus && properties.attackBonus !== 0 ? `${properties.attackBonus > 0 ? '+' : ''}${properties.attackBonus}` : '-';
                    if (defenseCell) defenseCell.textContent = properties?.defenseBonus && properties.defenseBonus !== 0 ? `${properties.defenseBonus > 0 ? '+' : ''}${properties.defenseBonus}` : '-';
                    if (healingCell) healingCell.textContent = properties?.healingBonus && properties.healingBonus > 0 ? `+${properties.healingBonus}` : '-';
                    if (captureCell) captureCell.textContent = properties?.canCapture ? '✓' : '-';
                    if (buildCell) buildCell.textContent = properties?.canBuild ? '✓' : '-';
                    
                    // Add alternating row colors
                    if (tbody.children.length % 2 === 1) {
                        row.classList.add('bg-gray-50', 'dark:bg-gray-700');
                    }
                    
                    tbody.appendChild(rowElement);
                    hasAnyTerrain = true;
                }
            }
        });
        
        // Only append the table if we have terrain to show
        if (hasAnyTerrain) {
            container.appendChild(tableElement);
        }
    }

    /**
     * Generate SVG histogram for damage distribution
     */
    private createDamageHistogram(damageDistribution: any): string {
        if (!damageDistribution || !damageDistribution.ranges || damageDistribution.ranges.length === 0) {
            return '<div class="text-gray-400 text-xs">No data</div>';
        }

        const ranges = damageDistribution.ranges;
        const width = 220;
        const height = 50;
        const topPadding = 5;
        const bottomPadding = 15; // Space for x-axis labels
        const chartHeight = height - topPadding - bottomPadding;
        const barSpacing = 1;
        const barWidth = (width / ranges.length) - barSpacing;
        
        // Find max probability for scaling
        const maxProbability = Math.max(...ranges.map((r: any) => r.probability || 0));
        
        // Create SVG with x-axis labels
        let svg = `<svg width="${width}" height="${height}" class="inline-block">`;
        
        ranges.forEach((range: any, index: number) => {
            const probability = range.probability || 0;
            const barHeight = (probability / maxProbability) * chartHeight;
            const x = index * (barWidth + barSpacing);
            const y = topPadding + chartHeight - barHeight;
            
            // Use the actual damage value (assuming minValue = maxValue for each bucket)
            const damageValue = Math.round(range.minValue);
            
            // Determine color based on damage value
            let fillColor = 'rgb(156, 163, 175)'; // gray-400
            if (damageValue >= 80) {
                fillColor = 'rgb(239, 68, 68)'; // red-500
            } else if (damageValue >= 60) {
                fillColor = 'rgb(251, 146, 60)'; // orange-400
            } else if (damageValue >= 40) {
                fillColor = 'rgb(250, 204, 21)'; // yellow-400
            } else if (damageValue >= 20) {
                fillColor = 'rgb(134, 239, 172)'; // green-300
            } else if (damageValue > 0) {
                fillColor = 'rgb(147, 197, 253)'; // blue-300
            }
            
            // Add bar with improved tooltip
            const percentageStr = (probability * 100).toFixed(1);
            const tooltipText = `${percentageStr}% of the time ${damageValue} damage dealt`;
            
            svg += `<rect x="${x}" y="${y}" width="${barWidth}" height="${barHeight}" 
                         fill="${fillColor}" opacity="0.8" 
                         class="hover:opacity-100 transition-opacity cursor-help"
                         data-tooltip="${tooltipText}">
                      <title>${tooltipText}</title>
                    </rect>`;
            
            // Add x-axis label for this bar (show every bar or every other bar depending on space)
            if (index % Math.ceil(ranges.length / 11) === 0 || index === ranges.length - 1) {
                svg += `<text x="${x + barWidth/2}" y="${height - 2}" 
                             text-anchor="middle" 
                             fill="currentColor" 
                             opacity="0.6" 
                             font-size="9">${damageValue}</text>`;
            }
        });
        
        // Add baseline
        svg += `<line x1="0" y1="${topPadding + chartHeight}" x2="${width}" y2="${topPadding + chartHeight}" 
                     stroke="currentColor" stroke-opacity="0.3" stroke-width="1"/>`;
        
        // Add expected damage marker if it exists
        if (damageDistribution.expectedDamage !== undefined) {
            const expectedX = (damageDistribution.expectedDamage / 100) * width;
            svg += `<line x1="${expectedX}" y1="${topPadding}" x2="${expectedX}" y2="${topPadding + chartHeight}" 
                         stroke="rgb(239, 68, 68)" stroke-width="2" stroke-dasharray="2,2" opacity="0.6">
                      <title>Expected: ${damageDistribution.expectedDamage.toFixed(1)}</title>
                    </line>`;
        }
        
        svg += '</svg>';
        
        // Return just the SVG without the duplicate summary text
        return svg;
    }

    /**
     * Generate unit-unit combat damage distribution table
     */
    private generateUnitCombatTable(unitId: number): void {
        const container = this.findElement('#unit-combat-properties');
        if (!container) return;

        // Get the table template
        const tableTemplate = document.getElementById('unit-combat-table-template') as HTMLTemplateElement;
        const rowTemplate = document.getElementById('combat-row-template') as HTMLTemplateElement;
        
        if (!tableTemplate || !rowTemplate) {
            console.warn('Unit combat table templates not found');
            return;
        }
        
        // Clear existing content
        container.innerHTML = '';
        
        // Clone the table template
        const tableElement = tableTemplate.content.cloneNode(true) as DocumentFragment;
        const tbody = tableElement.querySelector('tbody');
        
        if (!tbody) {
            console.warn('Table body not found in template');
            return;
        }
        
        // Get all available units for combat comparison
        const commonUnitIds = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15];
        let hasAnyCombat = false;
        
        commonUnitIds.forEach(targetUnitId => {
            const targetUnitDef = this.rulesTable.getUnitDefinition(targetUnitId);
            if (targetUnitDef && targetUnitDef.name) {
                // Get unit-unit combat properties
                const combatProps = this.rulesTable.getUnitUnitProperties(unitId, targetUnitId);
                
                if (combatProps && combatProps.damage) {
                    // Clone the row template
                    const rowElement = rowTemplate.content.cloneNode(true) as DocumentFragment;
                    const row = rowElement.querySelector('tr');
                    
                    if (row) {
                        // Fill in the row data
                        const targetNameCell = row.querySelector('[data-target-name]');
                        const damageHistogramCell = row.querySelector('[data-damage-histogram]');
                        
                        if (targetNameCell) targetNameCell.textContent = targetUnitDef.name;
                        
                        // Create histogram visualization
                        if (damageHistogramCell) {
                            const damageDistribution = combatProps.damage;
                            const histogram = this.createDamageHistogram(damageDistribution);
                            const summaryText = `<div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                Range: ${Math.round(damageDistribution.minDamage)}-${Math.round(damageDistribution.maxDamage)}, 
                                Avg: ${damageDistribution.expectedDamage.toFixed(0)}
                            </div>`;
                            damageHistogramCell.innerHTML = `<div>${histogram}${summaryText}</div>`;
                        }
                        
                        // Add alternating row colors
                        if (tbody.children.length % 2 === 1) {
                            row.classList.add('bg-gray-50', 'dark:bg-gray-700');
                        }
                        
                        tbody.appendChild(rowElement);
                        hasAnyCombat = true;
                    }
                }
            }
        });
        
        // Only append the table if we have combat data to show
        if (hasAnyCombat) {
            container.appendChild(tableElement);
        } else {
            // Show a message if no combat data available
            container.innerHTML = '<div class="text-sm text-gray-500 dark:text-gray-400 italic p-4">No combat data available for this unit</div>';
        }
    }

    /**
     * Get current unit info (for external access)
     */
    public getCurrentUnit(): UnitData | null {
        return this.currentUnit;
    }

    /**
     * Check if unit is currently selected
     */
    public hasUnitSelected(): boolean {
        return this.currentUnit !== null;
    }

    protected destroyComponent(): void {
        this.deactivate();
    }
}
