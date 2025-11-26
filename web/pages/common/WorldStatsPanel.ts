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
export class WorldStatsPanel extends BaseComponent implements LCMComponent {
    // Dependencies
    private world: World | null = null;
    private theme: ITheme;

    // Internal state
    private isUIBound = false;
    private isActivated = false;

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
                    <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Tiles</h4>
                    <div id="tiles-grid" class="flex flex-wrap gap-2">
                        <!-- Tile stats will be populated here -->
                    </div>
                </div>

                <!-- Units Section -->
                <div class="mb-4">
                    <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Units (Initial)</h4>
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
        this.updatePlayerTilesTable(tiles);
        this.updatePlayerUnitsTable(units);
    }

    // =========================================================================
    // Tiles Grid
    // =========================================================================

    private updateTilesGrid(tiles: Tile[]): void {
        const container = this.findElement('#tiles-grid');
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

        // Sort alphabetically by name
        tileData.sort((a, b) => a.name.localeCompare(b.name));

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

        // Sort alphabetically by name
        unitData.sort((a, b) => a.name.localeCompare(b.name));

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

    private updatePlayerTilesTable(tiles: Tile[]): void {
        const container = this.findElement('#player-tiles-table');
        if (!container) return;

        // Get unique players from tiles (only city tiles have player ownership)
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

        if (playerTileCounts.size === 0) {
            container.innerHTML = '<div class="text-gray-500 dark:text-gray-400 italic">No player-owned tiles</div>';
            return;
        }

        // Sort players; sort tile types by name
        const players = Array.from(playerTileCounts.keys()).sort((a, b) => a - b);
        const tileTypesWithNames = Array.from(allTileTypes).map(t => ({
            type: t,
            name: this.theme.getTerrainName(t) || `Type ${t}`
        }));
        tileTypesWithNames.sort((a, b) => a.name.localeCompare(b.name));

        // Build table
        let html = '<table class="w-full border-collapse">';

        // Header row
        html += '<thead><tr><th class="text-center p-1 border-b border-gray-200 dark:border-gray-600">Tile</th>';
        for (const player of players) {
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
            for (const player of players) {
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

    private updatePlayerUnitsTable(units: Unit[]): void {
        const container = this.findElement('#player-units-table');
        if (!container) return;

        // Get unique players and unit types
        const playerUnitCounts = new Map<number, Map<number, number>>();
        const allUnitTypes = new Set<number>();

        units.forEach(unit => {
            if (!playerUnitCounts.has(unit.player)) {
                playerUnitCounts.set(unit.player, new Map());
            }
            const playerMap = playerUnitCounts.get(unit.player)!;
            playerMap.set(unit.unitType, (playerMap.get(unit.unitType) || 0) + 1);
            allUnitTypes.add(unit.unitType);
        });

        if (playerUnitCounts.size === 0) {
            container.innerHTML = '<div class="text-gray-500 dark:text-gray-400 italic">No units</div>';
            return;
        }

        // Sort players; sort unit types by name
        const players = Array.from(playerUnitCounts.keys()).sort((a, b) => a - b);
        const unitTypesWithNames = Array.from(allUnitTypes).map(t => ({
            type: t,
            name: this.theme.getUnitName(t) || `Type ${t}`
        }));
        unitTypesWithNames.sort((a, b) => a.name.localeCompare(b.name));

        // Build table
        let html = '<table class="w-full border-collapse">';

        // Header row
        html += '<thead><tr><th class="text-center p-1 border-b border-gray-200 dark:border-gray-600">Unit</th>';
        for (const player of players) {
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
            for (const player of players) {
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
