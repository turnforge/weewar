import { EventBus } from '../lib/EventBus';
import { HexCoord } from './phaser/hexUtils';
import { 
    World as ProtoWorld, 
    WorldData as ProtoWorldData, 
    Tile as ProtoTile, 
    Unit as ProtoUnit,
    TileSchema,
    UnitSchema,
    WorldDataSchema,
    WorldSchema,
    TerrainDefinition, UnitDefinition, 
} from '../gen/weewar/v1/models_pb';
import { 
    UpdateWorldRequest,
    CreateWorldRequest,
    UpdateWorldRequestSchema,
    CreateWorldRequestSchema
} from '../gen/weewar/v1/worlds_pb';
import { create, toJson } from '@bufbuild/protobuf';

export interface WorldEvent {
    type: WorldEventType;
    data: any;
}

export enum WorldEventType {
    TILES_CHANGED = 'tiles-changed',      // Batch tile operations
    UNITS_CHANGED = 'units-changed',      // Batch unit operations
    WORLD_LOADED = 'world-loaded',
    WORLD_SAVED = 'world-saved',
    WORLD_CLEARED = 'world-cleared',
    WORLD_METADATA_CHANGED = 'world-metadata-changed'
}

export class Tile {
  q: number;
  r: number;
  tileType: number;
  player: number;
}

export class Unit {
  q: number;
  r: number;
  unitType: number;
  player: number;
}

// Batch event data types
export interface TileChange {
    q: number;
    r: number;
    tile: Tile | null;  // null means tile was removed
}

export interface UnitChange {
    q: number;
    r: number;
    unit: Unit| null;  // null means unit was removed
}

export interface TilesChangedEventData {
    changes: TileChange[];
}

export interface UnitsChangedEventData {
    changes: UnitChange[];
}

export interface WorldLoadedEventData {
    worldId: string | null;
    isNewWorld: boolean;
    tileCount: number;
    unitCount: number;
}

export interface SaveResult {
    success: boolean;
    worldId?: string;
    error?: string;
}

class WorldBounds {
  MinQ: number;
  MaxQ: number;
  MinR: number;
  MaxR: number;
  StartingCoord: { q: number; r: number };
  StartingX: number;
  MinX: number;
  MinY: number;
  MaxX: number;
  MaxY: number;
  MinXCoord: {Q: number, R: number};
  MinYCoord: {Q: number, R: number};
  MaxXCoord: {Q: number, R: number};
  MaxYCoord: {Q: number, R: number};
}

export interface WorldMetadata {
    name: string;
    width: number;
    height: number;
}

/**
 * World class handles all world data management including tiles, units, and metadata.
 *
 * Features:
 * - EventBus-based communication for decentralized component architecture
 */
export class World {
    // Core data
    private metadata: WorldMetadata;
    private tiles: { [key: string]: Tile} = {};
    private units: { [key: string]: Unit} = {};
    
    // Persistence state
    private worldId: string | null = null;
    private isNewWorld: boolean = true;
    private hasUnsavedChanges: boolean = false;
    
    // EventBus for decentralized communication
    private eventBus: EventBus;

    // Client-side batching control
    private isBatching: boolean = false;
    private pendingTileChanges: TileChange[] = [];
    private pendingUnitChanges: UnitChange[] = [];
    
    constructor(eventBus: EventBus, public name: string = 'New World', width: number = 40, height: number = 40) {
        this.eventBus = eventBus;
        this.metadata = { name, width, height };
    }
    
    // EventBus communication - emit state changes as events
    private emitStateChange(eventType: WorldEventType, data: any, emitter: any = null): void {
        this.eventBus.emit(eventType, data, emitter || this, this);
    }
    
    // Client-controlled batching methods
    public startBatch(): void {
        this.isBatching = true;
    }

    public get id(): string | null {
      return this.worldId
    }
    
    public commitBatch(): void {
        if (!this.isBatching) {
            return;
        }
        
        // Emit batched changes
        if (this.pendingTileChanges.length > 0) {
            this.emitStateChange(WorldEventType.TILES_CHANGED, {
                changes: [...this.pendingTileChanges]
            } as TilesChangedEventData);
            this.pendingTileChanges = [];
        }
        
        if (this.pendingUnitChanges.length > 0) {
            this.emitStateChange(WorldEventType.UNITS_CHANGED, {
                changes: [...this.pendingUnitChanges]
            } as UnitsChangedEventData);
            this.pendingUnitChanges = [];
        }
        
        this.isBatching = false;
    }
    
    public cancelBatch(): void {
        this.isBatching = false;
        this.pendingTileChanges = [];
        this.pendingUnitChanges = [];
    }
    
    private addTileChange(q: number, r: number, tile: Tile | null): void {
        this.hasUnsavedChanges = true;
        
        if (this.isBatching) {
            // Add to batch
            this.pendingTileChanges.push({ q, r, tile });
        } else {
            // Emit immediately
            this.emitStateChange(WorldEventType.TILES_CHANGED, {
                changes: [{ q, r, tile }]
            } as TilesChangedEventData);
        }
    }
    
    private addUnitChange(q: number, r: number, unit: Unit | null): void {
        this.hasUnsavedChanges = true;
        
        if (this.isBatching) {
            // Add to batch
            this.pendingUnitChanges.push({ q, r, unit });
        } else {
            // Emit immediately
            this.emitStateChange(WorldEventType.UNITS_CHANGED, {
                changes: [{ q, r, unit }]
            } as UnitsChangedEventData);
        }
    }
    
    // Persistence methods
    public getWorldId(): string | null {
        return this.worldId;
    }
    
    public setWorldId(worldId: string | null): void {
        worldId = (worldId || "").trim();
        this.worldId = worldId;
        this.isNewWorld = worldId === "";
    }
    
    public getIsNewWorld(): boolean {
        return this.isNewWorld;
    }
    
    public getHasUnsavedChanges(): boolean {
        return this.hasUnsavedChanges;
    }
    
    public markAsSaved(): void {
        this.hasUnsavedChanges = false;
    }
    
    // World metadata methods
    public getName(): string {
        return this.metadata.name;
    }
    
    public setName(name: string): void {
        if (this.metadata.name !== name) {
            this.metadata.name = name;
            this.hasUnsavedChanges = true;
            this.emitStateChange(WorldEventType.WORLD_METADATA_CHANGED, {
                name, width: this.metadata.width, height: this.metadata.height
            });
        }
    }
    
    public getWidth(): number {
        return this.metadata.width;
    }
    
    public setWidth(width: number): void {
        this.metadata.width = width;
    }
    
    public getHeight(): number {
        return this.metadata.height;
    }
    
    public setHeight(height: number): void {
        this.metadata.height = height;
    }
    
    public getMetadata(): WorldMetadata {
        return { ...this.metadata };
    }
    
    // Tile management methods
    public tileExistsAt(q: number, r: number): boolean {
        const key = `${q},${r}`;
        return key in this.tiles;
    }
    
    public getTileAt(q: number, r: number): Tile | null {
        const key = `${q},${r}`;
        return this.tiles[key] || null;
    }
    
    public setTileAt(q: number, r: number, tileType: number, player: number): void {
        const key = `${q},${r}`;
        const tile: Tile = { q, r, tileType, player };
        this.tiles[key] = tile;
        this.addTileChange(q, r, tile);
    }
    
    public removeTileAt(q: number, r: number): boolean {
        const key = `${q},${r}`;
        if (key in this.tiles) {
            delete this.tiles[key];
            this.addTileChange(q, r, null);
            return true;
        }
        return false;
    }
    
    public getAllTiles(): Array<Tile> {
        const result: Array<Tile> = [];
        
        Object.entries(this.tiles).forEach(([key, tileData]) => {
            const [q, r] = key.split(',').map(Number);
            result.push({
                q,
                r,
                tileType: tileData.tileType,
                player: tileData.player
            });
        });
        
        return result;
    }
    
    public clearAllTiles(): void {
        // Batch all tile removals
        const allTiles = this.getAllTiles();
        this.tiles = {};
        
        // Add all removals to pending changes
        allTiles.forEach(tile => {
            this.addTileChange(tile.q, tile.r, null);
        });
    }
    
    // Unit management methods
    public unitExistsAt(q: number, r: number): boolean {
        const key = `${q},${r}`;
        return key in this.units;
    }
    
    public getUnitAt(q: number, r: number): Unit | null {
        const key = `${q},${r}`;
        return this.units[key] || null;
    }
    
    public setUnitAt(q: number, r: number, unitType: number, player: number): void {
        // Ensure there's a tile at this location - auto-place grass if none exists
        if (!this.tileExistsAt(q, r)) {
            this.setTileAt(q, r, 1, 0); // Terrain type 1 = Grass, no player ownership
        }
        
        // Check if same unit type and player already exists - if so, remove it (toggle behavior)
        const existingUnit = this.getUnitAt(q, r);
        if (existingUnit && existingUnit.unitType === unitType && existingUnit.player === player) {
            // Same unit type and player - remove the unit (toggle off)
            this.removeUnitAt(q, r);
            return;
        }
        
        // Different unit type/player or no existing unit - place/replace the unit
        const key = `${q},${r}`;
        const unit: Unit = { q, r, unitType, player };
        this.units[key] = unit;
        this.addUnitChange(q, r, unit);
    }
    
    public removeUnitAt(q: number, r: number): boolean {
        const key = `${q},${r}`;
        if (key in this.units) {
            delete this.units[key];
            this.addUnitChange(q, r, null);
            return true;
        }
        return false;
    }
    
    public getAllUnits(): Array<Unit> {
        const result: Array<Unit> = [];
        
        Object.entries(this.units).forEach(([key, unitData]) => {
            const [q, r] = key.split(',').map(Number);
            result.push({
                q,
                r,
                unitType: unitData.unitType,
                player: unitData.player
            });
        });
        
        return result;
    }
    
    public clearAllUnits(): void {
        // Batch all unit removals
        const allUnits = this.getAllUnits();
        this.units = {};
        
        // Add all removals to pending changes
        allUnits.forEach(unit => {
            this.addUnitChange(unit.q, unit.r, null);
        });
    }
    
    // Utility methods
    public clearAll(): void {
        this.clearAllTiles();
        this.clearAllUnits();
        
        this.emitStateChange(WorldEventType.WORLD_CLEARED, {});
    }
    
    public fillAllTerrain(tileType: number, player: number, viewport?: { minQ: number, maxQ: number, minR: number, maxR: number }): void {
        // If viewport is provided, only fill visible area, otherwise fill entire world bounds
        if (viewport) {
            for (let q = viewport.minQ; q <= viewport.maxQ; q++) {
                for (let r = viewport.minR; r <= viewport.maxR; r++) {
                    this.setTileAt(q, r, tileType, player);
                }
            }
        } else {
            // Fill based on current world bounds or a reasonable default area
            const bounds = this.getBounds();
            const minQ = bounds ? bounds.minQ : -10;
            const maxQ = bounds ? bounds.maxQ : 10;
            const minR = bounds ? bounds.minR : -10;
            const maxR = bounds ? bounds.maxR : 10;
            
            for (let q = minQ; q <= maxQ; q++) {
                for (let r = minR; r <= maxR; r++) {
                    this.setTileAt(q, r, tileType, player);
                }
            }
        }
    }
    
    
    public getTileCount(): number {
        return Object.keys(this.tiles).length;
    }
    
    public getUnitCount(): number {
        return Object.keys(this.units).length;
    }
    
    public getBounds(): { minQ: number; maxQ: number; minR: number; maxR: number } | null {
        const allTiles = this.getAllTiles();
        const allUnits = this.getAllUnits();
        
        if (allTiles.length === 0 && allUnits.length === 0) {
            return null;
        }
        
        const allCoords = [
            ...allTiles.map(t => ({ q: t.q, r: t.r })),
            ...allUnits.map(u => ({ q: u.q, r: u.r }))
        ];
        
        const qs = allCoords.map(c => c.q);
        const rs = allCoords.map(c => c.r);
        
        return {
            minQ: Math.min(...qs),
            maxQ: Math.max(...qs),
            minR: Math.min(...rs),
            maxR: Math.max(...rs)
        };
    }
    
    // Self-contained persistence methods
    public async save(): Promise<SaveResult> {
        // Transform TypeScript World data into protobuf-compatible format
        const tiles: Array<ProtoTile> = Object.values(this.tiles).map(tile => 
            create(TileSchema, {
                q: tile.q,
                r: tile.r,
                tileType: tile.tileType,
                player: tile.player
            })
        );
        
        const units: Array<ProtoUnit> = Object.values(this.units).map(unit => 
            create(UnitSchema, {
                q: unit.q,
                r: unit.r,
                player: unit.player,
                unitType: unit.unitType,
                availableHealth: (unit as any).available_health || 100,
                distanceLeft: (unit as any).distance_left || 0,
                turnCounter: (unit as any).turn_counter || 0
            })
        );

        // Build World metadata (separate from WorldData)
        const worldMetadata: ProtoWorld = create(WorldSchema, {
            id: this.worldId || undefined,
            name: this.metadata.name || 'Untitled World',
            description: '',
            tags: [],
            difficulty: 'medium',
            creatorId: 'editor-user'
        });

        // Build WorldData (tiles and units)
        const worldData: ProtoWorldData = create(WorldDataSchema, {
            tiles: tiles,
            units: units
        });

        // Build request payload based on whether it's a new world or update
        let request: CreateWorldRequest | UpdateWorldRequest;
        let url: string;
        let method: string;

        if (this.isNewWorld) {
            // CreateWorldRequest
            request = create(CreateWorldRequestSchema, {
                world: worldMetadata,
                worldData: worldData
            });
            url = '/api/v1/worlds';
            method = 'POST';
        } else {
            // UpdateWorldRequest  
            request = create(UpdateWorldRequestSchema, {
                world: create(WorldSchema, {
                    ...worldMetadata,
                    id: this.worldId!
                }),
                worldData: worldData,
                clearWorld: false // Don't clear existing data
                // Note: update_mask is optional, omitting for now
            });
            url = `/api/v1/worlds/${this.worldId}`;
            method = 'PATCH';
        }

        // Convert protobuf request to JSON for HTTP call
        const requestJson = toJson(
            this.isNewWorld ? CreateWorldRequestSchema : UpdateWorldRequestSchema, 
            request
        );

        const response = await fetch(url, {
            method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(requestJson),
        });

        if (response.ok) {
            const result = await response.json();
            const newWorldId = result.world?.id || result.id;
            
            if (this.isNewWorld && newWorldId) {
                this.worldId = newWorldId;
                this.isNewWorld = false;
            }
            
            this.markAsSaved();
            
            this.emitStateChange(WorldEventType.WORLD_SAVED, {
                worldId: this.worldId, success: true
            });
            
            return { success: true, worldId: newWorldId };
        } else {
            const errorText = await response.text();
            throw new Error(`Save failed: ${response.status} ${response.statusText} - ${errorText}`);
        }
    }

    public loadFromElement(worldMetadataElement: HTMLElement, worldTilesElement: HTMLElement): World {
        // Now handles dual element loading: metadata + tiles/units data
        // Parse world metadata
        let worldMetadata = null;
        if (worldMetadataElement.textContent) {
            worldMetadata = JSON.parse(worldMetadataElement.textContent);
        }
        
        // Parse world tiles/units data
        let worldTilesData = null;
        if (worldTilesElement.textContent) {
            worldTilesData = JSON.parse(worldTilesElement.textContent);
        }
        
        if (!worldMetadata || !worldTilesData) {
            throw new Error('Failed to parse world metadata or tiles data');
        }
        // Combine into format expected by loadFromData()
        const combinedData = {
            // World metadata
            name: worldMetadata.name || 'Untitled World',
            Name: worldMetadata.name || 'Untitled World', // Both for compatibility
            id: worldMetadata.id,
            
            // Calculate dimensions from tiles if present
            width: 40,  // Default
            height: 40, // Default
            
            // World tiles and units
            tiles: worldTilesData.tiles || [],
            units: worldTilesData.units || []
        };
        
        // Calculate actual dimensions from tile bounds
        if (combinedData.tiles && combinedData.tiles.length > 0) {
            let maxQ = 0, maxR = 0, minQ = 0, minR = 0;
            combinedData.tiles.forEach((tile: any) => {
                if (tile.q > maxQ) maxQ = tile.q;
                if (tile.q < minQ) minQ = tile.q;
                if (tile.r > maxR) maxR = tile.r;
                if (tile.r < minR) minR = tile.r;
            });
            combinedData.width = maxQ - minQ + 1;
            combinedData.height = maxR - minR + 1;
        }
        
        return this.loadFromData(combinedData);
    }
    
    public loadFromData(data: any): World {
        if (!data) {
            throw new Error('No world data provided');
        }
        
        // Clear existing data without emitting events
        this.tiles = {};
        this.units = {};
        
        // Load metadata - handle both old and new formats
        if (data.name) this.metadata.name = data.name;
        if (data.Name) this.metadata.name = data.Name; // Backend format
        if (data.width) this.metadata.width = data.width;
        if (data.height) this.metadata.height = data.height;

        return this.loadTilesAndUnits(data.tiles, data.units)
    }

    loadTilesAndUnits(tiles: any, units: any): World {
        // Batch load tiles - handle both array and map formats
        const tileChanges: TileChange[] = [];
        tiles = tiles || []
        // Old map format for backward compatibility
        tiles.forEach((tileData: any) => {
            const q = tileData.q || 0
            const r = tileData.r || 0
            const key = q + "," + r;
            let tileType: number;
            let player = 0
            
            if (tileData.tileType !== undefined) {
                tileType = tileData.tileType;
                player = tileData.player;
            } else if (tileData.tile_type !== undefined) {
                tileType = tileData.tile_type;
                player = tileData.player || 0;
            } else {
                return; // Skip invalid tile
            }
            
            const tile: Tile = { q, r, tileType, player };
            this.tiles[key] = tile;
            tileChanges.push({ q, r, tile });
        });
        
        // Batch load units - handle both array and map formats
        const unitChanges: UnitChange[] = [];
        units = units || []
        // Old map format for backward compatibility
        units.forEach((unitData: any) => {
            const q = unitData.q || 0
            const r = unitData.r || 0
            const key = q + "," + r;
            let unitType: number, player: number;
            
            if (unitData.unitType !== undefined && unitData.player !== undefined) {
                unitType = unitData.unitType;
                player = unitData.player;
            } else if (unitData.unit_type !== undefined && unitData.player !== undefined) {
                unitType = unitData.unit_type;
                player = unitData.player;
            } else {
                return; // Skip invalid unit
            }
            
            const unit: Unit = { q, r, unitType, player };
            this.units[key] = unit;
            unitChanges.push({ q, r, unit });
        });
        
        // Emit batched changes immediately
        if (tileChanges.length > 0) {
            this.emitStateChange(WorldEventType.TILES_CHANGED, {
                changes: tileChanges
            } as TilesChangedEventData);
        }
        
        if (unitChanges.length > 0) {
            this.emitStateChange(WorldEventType.UNITS_CHANGED, {
                changes: unitChanges
            } as UnitsChangedEventData);
        }
        
        this.hasUnsavedChanges = false;
        
        this.emitStateChange(WorldEventType.WORLD_LOADED, {
            worldId: this.worldId,
            isNewWorld: this.isNewWorld,
            tileCount: this.getTileCount(),
            unitCount: this.getUnitCount()
        } as WorldLoadedEventData);
        return this
    }
    
    // Serialization methods - now matches backend array format
    public serialize(): {
        Name: string;
        PlayerCount: number;
        Tiles: Array<{ q: number; r: number ; tileType: number; player: number }>;
        Units: Array<{ q: number; r: number ; UnitType: number; Player: number }>;
    } {
        // Convert map format to array format matching backend
        const tilesArray = Object.entries(this.tiles).map(([coordKey, tileData]) => {
            const [q, r] = coordKey.split(',').map(Number);
            return {
                q: q,
                r: r,
                tileType: tileData.tileType,
                player: tileData.player || 0
            };
        });

        const unitsArray = Object.entries(this.units).map(([coordKey, unitData]) => {
            const [q, r] = coordKey.split(',').map(Number);
            return {
                q: q,
                r: r,
                UnitType: unitData.unitType,
                Player: unitData.player
            };
        });

        return {
            Name: this.metadata.name,
            PlayerCount: 2, // TODO: Track actual player count
            Tiles: tilesArray,
            Units: unitsArray
        };
    }
    
    public static deserialize(eventBus: EventBus, data: any): World {
        const world = new World(eventBus, data.name || 'Untitled World', data.width || 40, data.height || 40);
        world.loadFromData(data);
        return world;
    }
    
    // World validation
    public validate(): { isValid: boolean; errors: string[] } {
        const errors: string[] = [];
        
        // Check metadata
        if (!this.metadata.name || this.metadata.name.trim() === '') {
            errors.push('World name cannot be empty');
        }
        
        if (this.metadata.width <= 0 || this.metadata.height <= 0) {
            errors.push('World dimensions must be positive');
        }
        
        // Check tiles
        Object.entries(this.tiles).forEach(([key, tileData]) => {
            const [q, r] = key.split(',').map(Number);
            if (isNaN(q) || isNaN(r)) {
                errors.push(`Invalid tile coordinate: ${key}`);
            }
            if (tileData.tileType < 0) {
                errors.push(`Invalid tile type at ${key}: ${tileData.tileType}`);
            }
        });
        
        // Check units
        Object.entries(this.units).forEach(([key, unitData]) => {
            const [q, r] = key.split(',').map(Number);
            if (isNaN(q) || isNaN(r)) {
                errors.push(`Invalid unit coordinate: ${key}`);
            }
            if (unitData.unitType < 0) {
                errors.push(`Invalid unit type at ${key}: ${unitData.unitType}`);
            }
            if (unitData.player < 0 || unitData.player > 12) {
                errors.push(`Invalid player ID at ${key}: ${unitData.player}`);
            }
        });
        
        return {
            isValid: errors.length === 0,
            errors
        };
    }
    
    // Clone method for safe copying
    public clone(): World {
        return World.deserialize(this.eventBus, this.serialize());
    }
    
    /**
     * Calculate player count from world units
     */
    public get playerCount(): number {
        // Find the highest player ID in world units
        let maxPlayer = 0;
        for (const key in this.units) {
            const unit = this.units[key]
            if (unit.player && unit.player > maxPlayer) {
                maxPlayer = unit.player;
            }
        }
        
        // Player IDs are 1-based, so player count is maxPlayer
        // Ensure minimum of 2 players
        return Math.max(2, maxPlayer);
    }
}
