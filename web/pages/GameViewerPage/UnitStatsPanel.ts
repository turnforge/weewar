import { BaseComponent } from '../../lib/Component';
import { EventBus } from '../../lib/EventBus';
import { LCMComponent } from '../../lib/LCMComponent';
import { RulesTable } from '../common/RulesTable';
import { ITheme } from '../../assets/themes/BaseTheme';
import { ThemeUtils } from '../common/ThemeUtils';

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
    private isActivated = false;
    public rulesTable: RulesTable;
    private theme: ITheme | null = null;

    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('unit-stats-panel', rootElement, eventBus, debugMode);
        this.rulesTable = new RulesTable();
    }

    // Phase 4: Deactivate component
    public deactivate(): void {
        this.log('Deactivating UnitStatsPanel');
        this.isActivated = false;
        this.log('UnitStatsPanel deactivated');
    }

    /**
     * Set the theme for getting unit names
     */
    public setTheme(theme: ITheme): void {
        this.theme = theme;
    }

    /**
     * Hydrate theme images after Go template renders HTML
     * Call this after the HTML content is injected by the Go backend
     */
    public async hydrateThemeImages(): Promise<void> {
        await ThemeUtils.hydrateThemeImages(this.rootElement, this.theme, this.debugMode);
    }

    /**
     * Update the panel with information about a selected unit
     * NOTE: This method is being phased out in favor of Go template rendering
     */
    public updateUnitInfo(unit: UnitData): void {

        // Update unit information sections
        this.updateUnitHeader(unit);
        this.updateUnitStats(unit);
        this.updateUnitProperties(unit);
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
            const unitType = unit.unitType;
            const playerId = unit.player || 0;
            
            if (this.theme) {
                // Use the theme's setUnitImage method to handle all the complexity
                this.theme.setUnitImage(unitType, playerId, iconElement);
            } else {
                // Fallback to default PNG assets
                const imagePath = `/static/assets/themes/default/Units/${unitType}/${playerId}.png`;
                iconElement.innerHTML = `<img src="${imagePath}" alt="Unit ${unitType}" class="w-8 h-8 object-contain" style="image-rendering: pixelated;" onerror="this.style.display='none'; this.nextSibling.style.display='inline';">
                                         <span style="display:none;">⚔️</span>`;
            }
        }

        if (nameElement) {
            // Use theme-specific name if available, otherwise fallback to rules engine name
            const unitDef = this.rulesTable.getUnitDefinition(unit.unitType);
            const unitName = this.theme?.getUnitName(unit.unitType) || unitDef?.name || `Unit ${unit.unitType}`;
            nameElement.textContent = unitName;
        }

        if (playerElement) {
            playerElement.textContent = `Player ${unit.player}`;
        }

        if (descElement) {
            // Use theme-specific description if available, otherwise fallback to rules engine description
            const unitDef = this.rulesTable.getUnitDefinition(unit.unitType);
            const description = this.theme?.getUnitDescription?.(unit.unitType) || unitDef?.description || 'Military unit';
            descElement.textContent = description;
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
}
