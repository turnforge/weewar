import { 
    TerrainDefinition, UnitDefinition, MovementMatrix
} from '../gen/weewar/v1/models_pb';

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
    get id(): number { return this.terrainDefinition.ID; }
    get name(): string { return this.terrainDefinition.Name; }
    get baseMoveCost(): number { return this.terrainDefinition.BaseMoveCost; }
    get defenseBonus(): number { return this.terrainDefinition.DefenseBonus; }
    get type(): number { return this.terrainDefinition.type; }
    get description(): string { return this.terrainDefinition.description; }
}

export class RulesTable {
    // Cached rules engine data (loaded from page JSON)
    private terrainDefinitions: { [id: number]: TerrainDefinition } = {};
    private unitDefinitions: { [id: number]: UnitDefinition } = {};
    public movementMatrix: MovementMatrix | null = null;

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
     * Get terrain stats for a tile at the specified coordinates
     * Combines World tile data with TerrainDefinition from rules engine
     */
    public getTerrainStatsAt(tileId: number, player: number): TerrainStats | null {
        // Look up the TerrainDefinition using the tile's tileType
        const terrainDefinition = this.terrainDefinitions[tileId];
        if (!terrainDefinition) {
            console.log(`No terrain definition found for tile type ${tileId}`);
            return null;
        }

        // Create and return TerrainStats instance
        const terrainStats = new TerrainStats(terrainDefinition, 0, 0, player);
        console.log(`Created terrain stats for tile type: ${tileId}, player: ${player}`, {
            name: terrainStats.name,
            type: terrainStats.type,
            baseMoveCost: terrainStats.baseMoveCost,
            defenseBonus: terrainStats.defenseBonus
        });

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
            console.log('Loaded terrain definitions:', { count: Object.keys(this.terrainDefinitions).length });
        }

        // Load unit definitions  
        const unitElement = document.getElementById('unit-data-json');
        if (unitElement && unitElement.textContent) {
            const unitData = JSON.parse(unitElement.textContent);
            this.unitDefinitions = unitData;
            console.log('Loaded unit definitions:', { count: Object.keys(this.unitDefinitions).length });
        }

        // Load movement matrix
        const movementElement = document.getElementById('movement-matrix-json');
        if (movementElement && movementElement.textContent) {
            this.movementMatrix = JSON.parse(movementElement.textContent);
            console.log('Loaded movement matrix with', { unitTypes: Object.keys(this.movementMatrix?.costs || {}).length });
        }
    }
}
