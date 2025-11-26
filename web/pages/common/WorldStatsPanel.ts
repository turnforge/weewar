import { BaseComponent } from '../../lib/Component';
import { EventBus } from '../../lib/EventBus';
import { LCMComponent } from '../../lib/LCMComponent';
import { WorldEventTypes } from './events';
import { Unit, Tile, World } from './World';
import { ITheme } from '../../assets/themes/BaseTheme';
import DefaultTheme from '../../assets/themes/default';
import ModernTheme from '../../assets/themes/modern';
import FantasyTheme from '../../assets/themes/fantasy';
import { AssetThemePreference } from './AssetThemePreference';

/**
 * Theme registry for creating theme instances
 */
const THEME_REGISTRY: Record<string, new () => ITheme> = {
    default: DefaultTheme,
    modern: ModernTheme,
    fantasy: FantasyTheme,
};

/**
 * WorldStatsPanel Component - Displays world statistics with tile and unit breakdowns
 *
 * Features:
 * - Grid-based tile breakdown with icons and counts
 * - Grid-based unit breakdown with icons and counts
 * - Player distribution tables for tiles and units
 * - Listens to TILES_CHANGED and UNITS_CHANGED events for automatic updates
 *
 * Usage:
 * - Pass a World instance via setWorld()
 * - Component creates its own DOM structure
 * - Automatically updates when world data changes
 */
type SortField = 'name' | 'count';
type SortDirection = 'asc' | 'desc';

export class WorldStatsPanel extends BaseComponent implements LCMComponent {
    // Dependencies
    private world: World | null = null;
    private theme: ITheme;

    // Internal state
    private isUIBound = false;
    private isActivated = false;

    // Sorting state for tiles and units grids
    private tilesSortField: SortField = 'count';
    private tilesSortDirection: SortDirection = 'desc';
    private unitsSortField: SortField = 'count';
    private unitsSortDirection: SortDirection = 'desc';

    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('world-stats-panel', rootElement, eventBus, debugMode);

        // Initialize theme based on user preference
        const themeName = AssetThemePreference.get() || AssetThemePreference.DEFAULT_THEME;
        const ThemeClass = THEME_REGISTRY[themeName] || FantasyTheme;
        this.theme = new ThemeClass();
    }

    // =========================================================================
    // LCMComponent Lifecycle
    // =========================================================================

    public performLocalInit(): LCMComponent[] {
        if (this.isUIBound) {
            this.log('Already bound to DOM, skipping');
            return [];
        }

        this.log('Binding WorldStatsPanel to DOM');
        this.createStatsDisplay();
        this.isUIBound = true;
        this.log('WorldStatsPanel bound to DOM successfully');

        return [];
    }

    public setupDependencies(): void {
        this.log('WorldStatsPanel: Dependencies set via setWorld()');
    }

    public activate(): void {
        if (this.isActivated) {
            this.log('Already activated, skipping');
            return;
        }

        this.log('Activating WorldStatsPanel');
        this.isActivated = true;

        // Initial refresh if world is set
        if (this.world) {
            this.refreshStats();
        }

        this.log('WorldStatsPanel activated successfully');
    }

    public deactivate(): void {
        this.log('Deactivating WorldStatsPanel');

        // Unsubscribe from World events
        if (this.world) {
            this.removeSubscription(WorldEventTypes.TILES_CHANGED, this.world);
            this.removeSubscription(WorldEventTypes.UNITS_CHANGED, this.world);
            this.removeSubscription(WorldEventTypes.WORLD_LOADED, this.world);
        }

        this.isActivated = false;
        this.world = null;
        this.log('WorldStatsPanel deactivated');
    }

    // =========================================================================
    // Dependency Injection
    // =========================================================================

    public setWorld(world: World): void {
        // Unsubscribe from previous world
        if (this.world) {
            this.removeSubscription(WorldEventTypes.TILES_CHANGED, this.world);
            this.removeSubscription(WorldEventTypes.UNITS_CHANGED, this.world);
            this.removeSubscription(WorldEventTypes.WORLD_LOADED, this.world);
        }

        this.world = world;
        this.log('World set via setter');

        // Subscribe to world events
        if (world) {
            this.addSubscription(WorldEventTypes.TILES_CHANGED, world);
            this.addSubscription(WorldEventTypes.UNITS_CHANGED, world);
            this.addSubscription(WorldEventTypes.WORLD_LOADED, world);
        }

        // Refresh if already activated
        if (this.isActivated) {
            this.refreshStats();
        }
    }

    public getWorld(): World | null {
        return this.world;
    }

    // =========================================================================
    // Event Handling
    // =========================================================================

    public handleBusEvent(eventType: string, data: any, subject: any, emitter: any): void {
        switch (eventType) {
            case WorldEventTypes.TILES_CHANGED:
            case WorldEventTypes.UNITS_CHANGED:
            case WorldEventTypes.WORLD_LOADED:
                this.refreshStats();
                break;
            default:
                super.handleBusEvent(eventType, data, subject, emitter);
        }
    }

    // =========================================================================
    // DOM Creation
    // =========================================================================

    private createStatsDisplay(): void {
        this.rootElement.innerHTML = `
            <div class="world-stats-panel h-full bg-white dark:bg-gray-800 overflow-y-auto p-2">
                <!-- Tiles Section -->
                <div class="mb-4">
                    <div class="flex items-center justify-between mb-2">
                        <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300">Tiles</h4>
                        <div id="tiles-sort-controls" class="hidden flex items-center gap-1">
                            ${this.createSortControls('tiles')}
                        </div>
                    </div>
                    <div id="tiles-grid" class="flex flex-wrap gap-2">
                        <!-- Tile stats will be populated here -->
                    </div>
                </div>

                <!-- Units Section -->
                <div class="mb-4">
                    <div class="flex items-center justify-between mb-2">
                        <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300">Units (Initial)</h4>
                        <div id="units-sort-controls" class="hidden flex items-center gap-1">
                            ${this.createSortControls('units')}
                        </div>
                    </div>
                    <div id="units-grid" class="flex flex-wrap gap-2">
                        <!-- Unit stats will be populated here -->
                    </div>
                </div>

                <!-- Player Tile Distribution -->
                <div class="mb-4">
                    <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Tiles by Player</h4>
                    <div id="player-tiles-table" class="text-xs">
                        <!-- Player tile distribution will be populated here -->
                    </div>
                </div>

                <!-- Player Unit Distribution -->
                <div class="mb-4">
                    <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Units by Player</h4>
                    <div id="player-units-table" class="text-xs">
                        <!-- Player unit distribution will be populated here -->
                    </div>
                </div>
            </div>
        `;

        // Bind sort control event handlers
        this.bindSortControls();
    }

    private createSortControls(prefix: string): string {
        const sortField = prefix === 'tiles' ? this.tilesSortField : this.unitsSortField;
        const sortDirection = prefix === 'tiles' ? this.tilesSortDirection : this.unitsSortDirection;

        const nameSelected = sortField === 'name' ? 'selected' : '';
        const countSelected = sortField === 'count' ? 'selected' : '';

        const directionIcon = sortDirection === 'asc'
            ? '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4" />'
            : '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 4v12m0 0l-4-4m4 4l4-4m6 0V4m0 0l4 4m-4-4l-4 4" />';
        const directionTitle = sortDirection === 'asc' ? 'Ascending (click to change)' : 'Descending (click to change)';

        return `
            <select id="${prefix}-sort-field" class="text-xs px-1 py-0.5 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300">
                <option value="name" ${nameSelected}>Name</option>
                <option value="count" ${countSelected}>Count</option>
            </select>
            <button id="${prefix}-sort-direction" class="p-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-600 text-gray-600 dark:text-gray-400" title="${directionTitle}">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    ${directionIcon}
                </svg>
            </button>
        `;
    }

    private bindSortControls(): void {
        // Tiles sort controls
        const tilesSortField = this.findElement('#tiles-sort-field') as HTMLSelectElement;
        const tilesSortDirection = this.findElement('#tiles-sort-direction') as HTMLButtonElement;

        if (tilesSortField) {
            tilesSortField.addEventListener('change', () => {
                this.tilesSortField = tilesSortField.value as SortField;
                this.refreshTilesGrid();
            });
        }

        if (tilesSortDirection) {
            tilesSortDirection.addEventListener('click', () => {
                this.tilesSortDirection = this.tilesSortDirection === 'asc' ? 'desc' : 'asc';
                this.updateSortDirectionIcon('tiles', this.tilesSortDirection);
                this.refreshTilesGrid();
            });
        }

        // Units sort controls
        const unitsSortField = this.findElement('#units-sort-field') as HTMLSelectElement;
        const unitsSortDirection = this.findElement('#units-sort-direction') as HTMLButtonElement;

        if (unitsSortField) {
            unitsSortField.addEventListener('change', () => {
                this.unitsSortField = unitsSortField.value as SortField;
                this.refreshUnitsGrid();
            });
        }

        if (unitsSortDirection) {
            unitsSortDirection.addEventListener('click', () => {
                this.unitsSortDirection = this.unitsSortDirection === 'asc' ? 'desc' : 'asc';
                this.updateSortDirectionIcon('units', this.unitsSortDirection);
                this.refreshUnitsGrid();
            });
        }
    }

    private updateSortDirectionIcon(prefix: string, direction: SortDirection): void {
        const button = this.findElement(`#${prefix}-sort-direction`) as HTMLButtonElement;
        if (!button) return;

        // Up arrow for ascending, down arrow for descending
        const svg = direction === 'asc'
            ? '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4" /></svg>'
            : '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 4v12m0 0l-4-4m4 4l4-4m6 0V4m0 0l4 4m-4-4l-4 4" /></svg>';
        button.innerHTML = svg;
        button.title = direction === 'asc' ? 'Ascending (click to change)' : 'Descending (click to change)';
    }

    private refreshTilesGrid(): void {
        if (!this.world) return;
        this.updateTilesGrid(this.world.getAllTiles());
    }

    private refreshUnitsGrid(): void {
        if (!this.world) return;
        this.updateUnitsGrid(this.world.getAllUnits());
    }

    // =========================================================================
    // Stats Refresh
    // =========================================================================

    public refreshStats(): void {
        if (!this.isActivated || !this.world) {
            this.log('Component not ready for stats refresh');
            return;
        }

        const tiles = this.world.getAllTiles();
        const units = this.world.getAllUnits();

        this.updateTilesGrid(tiles);
        this.updateUnitsGrid(units);

        // Collect all unique players from both tiles and units for consistent table columns
        const allPlayers = this.collectAllPlayers(tiles, units);
        this.updatePlayerTilesTable(tiles, allPlayers);
        this.updatePlayerUnitsTable(units, allPlayers);
    }

    /**
     * Collect all unique player IDs from tiles and units
     */
    private collectAllPlayers(tiles: Tile[], units: Unit[]): number[] {
        const playerSet = new Set<number>();

        // Collect from tiles (only player-owned tiles)
        tiles.forEach(tile => {
            if (tile.player > 0) {
                playerSet.add(tile.player);
            }
        });

        // Collect from units
        units.forEach(unit => {
            if (unit.player > 0) {
                playerSet.add(unit.player);
            }
        });

        return Array.from(playerSet).sort((a, b) => a - b);
    }

    // =========================================================================
    // Tiles Grid
    // =========================================================================

    private updateTilesGrid(tiles: Tile[]): void {
        const container = this.findElement('#tiles-grid');
        const sortControls = this.findElement('#tiles-sort-controls');
        if (!container) return;

        // Count tiles by type
        const tileCounts = new Map<number, number>();
        tiles.forEach(tile => {
            const count = tileCounts.get(tile.tileType) || 0;
            tileCounts.set(tile.tileType, count + 1);
        });

        // Build array with names for sorting
        const tileData = Array.from(tileCounts.entries()).map(([tileType, count]) => ({
            tileType,
            count,
            name: this.theme.getTerrainName(tileType) || `Type ${tileType}`
        }));

        // Show/hide sort controls based on item count
        if (sortControls) {
            if (tileData.length > 1) {
                sortControls.classList.remove('hidden');
            } else {
                sortControls.classList.add('hidden');
            }
        }

        // Sort based on current sort settings
        tileData.sort((a, b) => {
            let comparison: number;
            if (this.tilesSortField === 'name') {
                comparison = a.name.localeCompare(b.name);
            } else {
                comparison = a.count - b.count;
            }
            return this.tilesSortDirection === 'asc' ? comparison : -comparison;
        });

        // Generate grid items with icon, name, and count
        let html = '';
        for (const { tileType, count, name } of tileData) {
            html += `
                <div class="flex items-center gap-1 px-2 py-1 bg-gray-50 dark:bg-gray-700 rounded border border-gray-200 dark:border-gray-600" title="${name}">
                    <div class="w-6 h-6 flex-shrink-0 flex items-center justify-center" data-tile-icon="${tileType}"></div>
                    <span class="text-xs text-gray-700 dark:text-gray-300 whitespace-nowrap">${name}</span>
                    <span class="text-xs text-gray-500 dark:text-gray-400 font-medium">×${count}</span>
                </div>
            `;
        }

        container.innerHTML = html || '<div class="text-sm text-gray-500 dark:text-gray-400 italic">No tiles</div>';

        // Load tile images
        this.loadTileIcons(tileData.map(d => d.tileType));
    }

    private async loadTileIcons(tileTypes: number[]): Promise<void> {
        for (const tileType of tileTypes) {
            const iconContainer = this.findElement(`[data-tile-icon="${tileType}"]`);
            if (iconContainer) {
                await this.theme.setTileImage(tileType, 0, iconContainer as HTMLElement);
            }
        }
    }

    // =========================================================================
    // Units Grid
    // =========================================================================

    private updateUnitsGrid(units: Unit[]): void {
        const container = this.findElement('#units-grid');
        const sortControls = this.findElement('#units-sort-controls');
        if (!container) return;

        // Count units by type
        const unitCounts = new Map<number, number>();
        units.forEach(unit => {
            const count = unitCounts.get(unit.unitType) || 0;
            unitCounts.set(unit.unitType, count + 1);
        });

        // Build array with names for sorting
        const unitData = Array.from(unitCounts.entries()).map(([unitType, count]) => ({
            unitType,
            count,
            name: this.theme.getUnitName(unitType) || `Type ${unitType}`
        }));

        // Show/hide sort controls based on item count
        if (sortControls) {
            if (unitData.length > 1) {
                sortControls.classList.remove('hidden');
            } else {
                sortControls.classList.add('hidden');
            }
        }

        // Sort based on current sort settings
        unitData.sort((a, b) => {
            let comparison: number;
            if (this.unitsSortField === 'name') {
                comparison = a.name.localeCompare(b.name);
            } else {
                comparison = a.count - b.count;
            }
            return this.unitsSortDirection === 'asc' ? comparison : -comparison;
        });

        // Generate grid items with icon, name, and count
        let html = '';
        for (const { unitType, count, name } of unitData) {
            html += `
                <div class="flex items-center gap-1 px-2 py-1 bg-gray-50 dark:bg-gray-700 rounded border border-gray-200 dark:border-gray-600" title="${name}">
                    <div class="w-6 h-6 flex-shrink-0 flex items-center justify-center" data-unit-icon="${unitType}"></div>
                    <span class="text-xs text-gray-700 dark:text-gray-300 whitespace-nowrap">${name}</span>
                    <span class="text-xs text-gray-500 dark:text-gray-400 font-medium">×${count}</span>
                </div>
            `;
        }

        container.innerHTML = html || '<div class="text-sm text-gray-500 dark:text-gray-400 italic">No units</div>';

        // Load unit images
        this.loadUnitIcons(unitData.map(d => d.unitType));
    }

    private async loadUnitIcons(unitTypes: number[]): Promise<void> {
        for (const unitType of unitTypes) {
            const iconContainer = this.findElement(`[data-unit-icon="${unitType}"]`);
            if (iconContainer) {
                // Use player 0 (neutral) for the overview
                await this.theme.setUnitImage(unitType, 0, iconContainer as HTMLElement);
            }
        }
    }

    // =========================================================================
    // Player Distribution Tables
    // =========================================================================

    private updatePlayerTilesTable(tiles: Tile[], allPlayers: number[]): void {
        const container = this.findElement('#player-tiles-table');
        if (!container) return;

        // Get tile counts per player
        const playerTileCounts = new Map<number, Map<number, number>>();
        const allTileTypes = new Set<number>();

        tiles.forEach(tile => {
            // Only count tiles with player ownership (player > 0)
            if (tile.player > 0) {
                if (!playerTileCounts.has(tile.player)) {
                    playerTileCounts.set(tile.player, new Map());
                }
                const playerMap = playerTileCounts.get(tile.player)!;
                playerMap.set(tile.tileType, (playerMap.get(tile.tileType) || 0) + 1);
                allTileTypes.add(tile.tileType);
            }
        });

        if (allTileTypes.size === 0) {
            container.innerHTML = '<div class="text-gray-500 dark:text-gray-400 italic">No player-owned tiles</div>';
            return;
        }

        // Sort tile types by name
        const tileTypesWithNames = Array.from(allTileTypes).map(t => ({
            type: t,
            name: this.theme.getTerrainName(t) || `Type ${t}`
        }));
        tileTypesWithNames.sort((a, b) => a.name.localeCompare(b.name));

        // Build table using allPlayers for consistent columns
        let html = '<table class="w-full border-collapse">';

        // Header row
        html += '<thead><tr><th class="text-center p-1 border-b border-gray-200 dark:border-gray-600">Tile</th>';
        for (const player of allPlayers) {
            html += `<th class="text-center p-1 border-b border-gray-200 dark:border-gray-600">P${player}</th>`;
        }
        html += '</tr></thead><tbody>';

        // Data rows with icon + name centered in first cell
        for (const { type: tileType, name } of tileTypesWithNames) {
            html += `<tr>
                <td class="p-1">
                    <div class="flex flex-col items-center">
                        <div class="w-6 h-6 flex items-center justify-center" data-player-tile-icon="${tileType}"></div>
                        <span class="text-xs text-gray-700 dark:text-gray-300">${name}</span>
                    </div>
                </td>`;
            for (const player of allPlayers) {
                const count = playerTileCounts.get(player)?.get(tileType) || 0;
                html += `<td class="text-center p-1 text-gray-600 dark:text-gray-400 align-middle">${count || '-'}</td>`;
            }
            html += '</tr>';
        }

        html += '</tbody></table>';
        container.innerHTML = html;

        // Load tile icons
        this.loadPlayerTileIcons(tileTypesWithNames.map(t => t.type));
    }

    private async loadPlayerTileIcons(tileTypes: number[]): Promise<void> {
        for (const tileType of tileTypes) {
            const iconContainer = this.findElement(`[data-player-tile-icon="${tileType}"]`);
            if (iconContainer) {
                await this.theme.setTileImage(tileType, 0, iconContainer as HTMLElement);
            }
        }
    }

    private updatePlayerUnitsTable(units: Unit[], allPlayers: number[]): void {
        const container = this.findElement('#player-units-table');
        if (!container) return;

        // Get unit counts per player
        const playerUnitCounts = new Map<number, Map<number, number>>();
        const allUnitTypes = new Set<number>();

        units.forEach(unit => {
            if (unit.player > 0) {
                if (!playerUnitCounts.has(unit.player)) {
                    playerUnitCounts.set(unit.player, new Map());
                }
                const playerMap = playerUnitCounts.get(unit.player)!;
                playerMap.set(unit.unitType, (playerMap.get(unit.unitType) || 0) + 1);
                allUnitTypes.add(unit.unitType);
            }
        });

        if (allUnitTypes.size === 0) {
            container.innerHTML = '<div class="text-gray-500 dark:text-gray-400 italic">No units</div>';
            return;
        }

        // Sort unit types by name
        const unitTypesWithNames = Array.from(allUnitTypes).map(t => ({
            type: t,
            name: this.theme.getUnitName(t) || `Type ${t}`
        }));
        unitTypesWithNames.sort((a, b) => a.name.localeCompare(b.name));

        // Build table using allPlayers for consistent columns
        let html = '<table class="w-full border-collapse">';

        // Header row
        html += '<thead><tr><th class="text-center p-1 border-b border-gray-200 dark:border-gray-600">Unit</th>';
        for (const player of allPlayers) {
            html += `<th class="text-center p-1 border-b border-gray-200 dark:border-gray-600">P${player}</th>`;
        }
        html += '</tr></thead><tbody>';

        // Data rows with icon + name centered in first cell
        for (const { type: unitType, name } of unitTypesWithNames) {
            html += `<tr>
                <td class="p-1">
                    <div class="flex flex-col items-center">
                        <div class="w-6 h-6 flex items-center justify-center" data-player-unit-icon="${unitType}"></div>
                        <span class="text-xs text-gray-700 dark:text-gray-300">${name}</span>
                    </div>
                </td>`;
            for (const player of allPlayers) {
                const count = playerUnitCounts.get(player)?.get(unitType) || 0;
                html += `<td class="text-center p-1 text-gray-600 dark:text-gray-400 align-middle">${count || '-'}</td>`;
            }
            html += '</tr>';
        }

        html += '</tbody></table>';
        container.innerHTML = html;

        // Load unit icons
        this.loadPlayerUnitIcons(unitTypesWithNames.map(t => t.type));
    }

    private async loadPlayerUnitIcons(unitTypes: number[]): Promise<void> {
        for (const unitType of unitTypes) {
            const iconContainer = this.findElement(`[data-player-unit-icon="${unitType}"]`);
            if (iconContainer) {
                await this.theme.setUnitImage(unitType, 0, iconContainer as HTMLElement);
            }
        }
    }

    // =========================================================================
    // Legacy/Compatibility
    // =========================================================================

    public destroy(): void {
        this.deactivate();
    }

    protected destroyComponent(): void {
        this.deactivate();
    }
}
