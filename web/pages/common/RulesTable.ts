import { 
    TerrainDefinition, UnitDefinition, TerrainUnitProperties, UnitUnitProperties
} from '../../gen/wasmjs/lilbattle/v1/models/interfaces'

/**
 * Terrain stats class combining TerrainDefinition with tile coordinate data
 * Extends the generated TerrainDefinition with position information
 */
export class TerrainStats {
    public readonly terrainDefinition: TerrainDefinition;
    public readonly q: number;
    public readonly r: number;
    public readonly player: number;

    constructor(terrainDefinition: TerrainDefinition, q: number, r: number, player: number = 0) {
        this.terrainDefinition = terrainDefinition;
        this.q = q;
        this.r = r;
        this.player = player;
    }

    // Convenience getters that delegate to TerrainDefinition
    get id(): number { return this.terrainDefinition.id; }
    get name(): string { return this.terrainDefinition.name; }
    get description(): string { return this.terrainDefinition.description; }
}

export class RulesTable {
    // Cached rules engine data (loaded from page JSON)
    private terrainDefinitions: { [id: number]: TerrainDefinition } = {};
    private unitDefinitions: { [id: number]: UnitDefinition } = {};
    private terrainUnitProperties: { [key: string]: TerrainUnitProperties } = {};
    private unitUnitProperties: { [key: string]: UnitUnitProperties } = {};

    constructor() {
        // Load rules engine data from page
        this.loadRulesEngineData();
    }

    /**
     * Get unit definition by ID
     */
    public getUnitDefinition(unitId: number): UnitDefinition | null {
        return this.unitDefinitions[unitId] || null;
    }

    /**
     * Get terrain definition by ID
     */
    public getTerrainDefinition(terrainId: number): TerrainDefinition | null {
        return this.terrainDefinitions[terrainId] || null;
    }

    /**
     * Get movement cost for a unit on specific terrain
     */
    public getMovementCost(terrainId: number, unitId: number): number {
        const key = `${terrainId}:${unitId}`;
        const properties = this.terrainUnitProperties[key];
        return properties?.movementCost ?? 1.0; // Default movement cost
    }

    /**
     * Get terrain-unit properties
     */
    public getTerrainUnitProperties(terrainId: number, unitId: number): TerrainUnitProperties | null {
        const key = `${terrainId}:${unitId}`;
        return this.terrainUnitProperties[key] || null;
    }

    /**
     * Get unit-vs-unit combat properties
     */
    public getUnitUnitProperties(attackerId: number, defenderId: number): UnitUnitProperties | null {
        const key = `${attackerId}:${defenderId}`;
        return this.unitUnitProperties[key] || null;
    }

    /**
     * Get terrain stats for a tile at the specified coordinates
     * Combines World tile data with TerrainDefinition from rules engine
     */
    public getTerrainStatsAt(tileId: number, player: number): TerrainStats | null {
        // Look up the TerrainDefinition using the tile's tileType
        const terrainDefinition = this.terrainDefinitions[tileId];
        if (!terrainDefinition) {
            return null;
        }

        // Create and return TerrainStats instance
        const terrainStats = new TerrainStats(terrainDefinition, 0, 0, player);

        return terrainStats;
    }

    /**
     * Load rules engine data from embedded JSON in page
     */
    private loadRulesEngineData(): void {
        // Load terrain definitions
        const terrainElement = document.getElementById('terrain-data-json');
        if (terrainElement && terrainElement.textContent) {
            const terrainData = JSON.parse(terrainElement.textContent);
            this.terrainDefinitions = terrainData;
        }

        // Load unit definitions  
        const unitElement = document.getElementById('unit-data-json');
        if (unitElement && unitElement.textContent) {
            const unitData = JSON.parse(unitElement.textContent);
            this.unitDefinitions = unitData;
        }

        // Load terrain-unit properties (centralized)
        const terrainUnitElement = document.getElementById('terrain-unit-properties-json');
        if (terrainUnitElement && terrainUnitElement.textContent) {
            this.terrainUnitProperties = JSON.parse(terrainUnitElement.textContent);
        }

        // Load unit-unit combat properties (centralized)
        const unitUnitElement = document.getElementById('unit-unit-properties-json');
        if (unitUnitElement && unitUnitElement.textContent) {
            this.unitUnitProperties = JSON.parse(unitUnitElement.textContent);
        }
    }
}
