import { EventBus } from '@panyam/tsappkit';
import { WorldEventTypes, WorldEventType } from './events';
import { rowColToHex, hexToRowCol, axialNeighbors, hexDistance, getDirectionIndex, getOppositeDirection, getNeighborCoord } from "./hexUtils";
import { Weewar_v1Deserializer as WD } from '../../gen/wasmjs/weewar/v1/factory';
import {
    World as ProtoWorld,
    WorldData as ProtoWorldData,
    Tile,
    Unit,
    UpdateWorldRequest,
    CreateWorldRequest,
    CrossingType,
    Crossing,
} from '../../gen/wasmjs/weewar/v1/models/interfaces'
import * as models from '../../gen/wasmjs/weewar/v1/models/models'
import { create, toJson } from '@bufbuild/protobuf';

export interface WorldEvent {
    type: WorldEventType;
    data: any;
}

// Using proto-generated Tile, Unit, CrossingType, and Crossing types directly
export { Tile, Unit, CrossingType, Crossing };

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

export interface CrossingChange {
    q: number;
    r: number;
    crossing: Crossing | null;  // null means crossing was removed
}

export interface TilesChangedEventData {
    changes: TileChange[];
}

export interface UnitsChangedEventData {
    changes: UnitChange[];
}

export interface CrossingsChangedEventData {
    changes: CrossingChange[];
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
    defaultGameConfig?: any; // GameConfiguration proto object
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
    public tiles: { [key: string]: Tile} = {};
    public units: { [key: string]: Unit} = {};
    public crossings: { [key: string]: Crossing } = {};
    
    // Persistence state
    private worldId: string | null = null;
    private isNewWorld: boolean = true;
    private hasUnsavedChanges: boolean = false;
    private version: number = 0;  // Version for optimistic locking (from WorldData.version)
    
    // EventBus for decentralized communication
    private eventBus: EventBus;

    // Client-side batching control
    private isBatching: boolean = false;
    private pendingTileChanges: TileChange[] = [];
    private pendingUnitChanges: UnitChange[] = [];
    private pendingCrossingChanges: CrossingChange[] = [];
    
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
            this.emitStateChange(WorldEventTypes.TILES_CHANGED, {
                changes: [...this.pendingTileChanges]
            } as TilesChangedEventData);
            this.pendingTileChanges = [];
        }

        if (this.pendingUnitChanges.length > 0) {
            this.emitStateChange(WorldEventTypes.UNITS_CHANGED, {
                changes: [...this.pendingUnitChanges]
            } as UnitsChangedEventData);
            this.pendingUnitChanges = [];
        }

        if (this.pendingCrossingChanges.length > 0) {
            this.emitStateChange(WorldEventTypes.CROSSINGS_CHANGED, {
                changes: [...this.pendingCrossingChanges]
            } as CrossingsChangedEventData);
            this.pendingCrossingChanges = [];
        }

        this.isBatching = false;
    }

    public cancelBatch(): void {
        this.isBatching = false;
        this.pendingTileChanges = [];
        this.pendingUnitChanges = [];
        this.pendingCrossingChanges = [];
    }
    
    private addTileChange(q: number, r: number, tile: Tile | null): void {
        this.hasUnsavedChanges = true;
        
        if (this.isBatching) {
            // Add to batch
            this.pendingTileChanges.push({ q, r, tile });
        } else {
            // Emit immediately
            this.emitStateChange(WorldEventTypes.TILES_CHANGED, {
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
            this.emitStateChange(WorldEventTypes.UNITS_CHANGED, {
                changes: [{ q, r, unit }]
            } as UnitsChangedEventData);
        }
    }

    private addCrossingChange(q: number, r: number, crossing: Crossing | null): void {
        this.hasUnsavedChanges = true;

        if (this.isBatching) {
            // Add to batch
            this.pendingCrossingChanges.push({ q, r, crossing });
        } else {
            // Emit immediately
            this.emitStateChange(WorldEventTypes.CROSSINGS_CHANGED, {
                changes: [{ q, r, crossing }]
            } as CrossingsChangedEventData);
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
            this.emitStateChange(WorldEventTypes.WORLD_METADATA_CHANGED, {
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

    public getDefaultGameConfig(): any {
        return this.metadata.defaultGameConfig;
    }

    public setDefaultGameConfig(config: any): void {
        this.metadata.defaultGameConfig = config;
        this.hasUnsavedChanges = true;
        this.emitStateChange(WorldEventTypes.WORLD_METADATA_CHANGED, {
            name: this.metadata.name, width: this.metadata.width, height: this.metadata.height
        });
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
        const tile = { q, r, tileType, player, number: 0, shortcut: "", lastActedTurn: 0, lastToppedupTurn: 0} as Tile;
        this.tiles[key] = tile;
        this.addTileChange(q, r, tile);
    }

    /**
     * Set a tile directly with full Tile object (preserves all runtime state)
     */
    public setTileDirect(tile: Tile): void {
        const key = `${tile.q},${tile.r}`;
        this.tiles[key] = tile;
        this.addTileChange(tile.q, tile.r, tile);
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
        return Object.values(this.tiles);
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
    
    /**
     * Set a unit directly with full Unit object (preserves all runtime state)
     */
    public setUnitDirect(unit: Unit): void {
        const key = `${unit.q},${unit.r}`;
        const existingUnit = this.units[key];
        this.units[key] = unit;
        this.addUnitChange(unit.q, unit.r, unit);
    }
    
    public setUnitAt(q: number, r: number, unitType: number, player: number): void {
        // Ensure there's a tile at this location - auto-place grass if none exists
        if (!this.tileExistsAt(q, r)) {
            this.setTileAt(q, r, 1, 0); // Terrain type 1 = Grass, no player ownership
        }

        // Place or replace the unit at this location
        const key = `${q},${r}`;
        const unit = WD.from(models.Unit,{ q, r, unitType, player });
        this.units[key] = unit;
        this.addUnitChange(q, r, unit);
    }
    
    public removeUnitAt(q: number, r: number): boolean {
        const key = `${q},${r}`;
        const existingUnit = this.units[key];
        
        if (key in this.units) {
            delete this.units[key];
            
            this.addUnitChange(q, r, null);
            return true;
        }
        
        return false;
    }
    
    public getAllUnits(): Array<Unit> {
        const result: Array<Unit> = [];

        Object.values(this.units).forEach((unitData: Unit) => {
            result.push(WD.from(models.Unit,{
                q: unitData.q,
                r: unitData.r,
                unitType: unitData.unitType,
                player: unitData.player,
                availableHealth: unitData.availableHealth,
                distanceLeft: unitData.distanceLeft,
                lastActedTurn: unitData.lastActedTurn,
                lastToppedupTurn: unitData.lastToppedupTurn,
            }));
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

    // Crossing management methods

    /**
     * Get the crossing at a hex coordinate (returns null if none)
     */
    public getCrossingAt(q: number, r: number): Crossing | null {
        const key = `${q},${r}`;
        return this.crossings[key] || null;
    }

    /**
     * Get the crossing type at a hex coordinate
     */
    public getCrossingTypeAt(q: number, r: number): CrossingType {
        const crossing = this.getCrossingAt(q, r);
        return crossing?.type || CrossingType.CROSSING_TYPE_UNSPECIFIED;
    }

    public hasCrossing(q: number, r: number): boolean {
        const crossing = this.getCrossingAt(q, r);
        return crossing !== null && crossing.type !== CrossingType.CROSSING_TYPE_UNSPECIFIED;
    }

    public hasRoad(q: number, r: number): boolean {
        return this.getCrossingTypeAt(q, r) === CrossingType.CROSSING_TYPE_ROAD;
    }

    public hasBridge(q: number, r: number): boolean {
        return this.getCrossingTypeAt(q, r) === CrossingType.CROSSING_TYPE_BRIDGE;
    }

    /**
     * Create an empty crossing with no connections
     */
    private createEmptyCrossing(type: CrossingType): Crossing {
        return {
            type,
            connectsTo: [false, false, false, false, false, false]
        };
    }

    /**
     * Add a bidirectional crossing connection between two neighboring tiles.
     * Creates crossings if they don't exist. Extends existing crossings.
     *
     * @param fromQ Source tile Q coordinate
     * @param fromR Source tile R coordinate
     * @param toQ Target tile Q coordinate
     * @param toR Target tile R coordinate
     * @param type Crossing type (road or bridge)
     * @returns true if connection was added, false if tiles are not neighbors
     */
    public addCrossingConnection(fromQ: number, fromR: number, toQ: number, toR: number, type: CrossingType): boolean {
        // Get direction from fromTile to toTile
        const directionToTarget = getDirectionIndex(fromQ, fromR, toQ, toR);
        if (directionToTarget === null) {
            // Tiles are not neighbors
            return false;
        }

        const directionToSource = getOppositeDirection(directionToTarget);
        const fromKey = `${fromQ},${fromR}`;
        const toKey = `${toQ},${toR}`;

        // Get or create crossing at source tile
        let fromCrossing = this.crossings[fromKey];
        if (!fromCrossing) {
            fromCrossing = this.createEmptyCrossing(type);
            this.crossings[fromKey] = fromCrossing;
        }
        fromCrossing.connectsTo[directionToTarget] = true;

        // Get or create crossing at target tile
        let toCrossing = this.crossings[toKey];
        if (!toCrossing) {
            toCrossing = this.createEmptyCrossing(type);
            this.crossings[toKey] = toCrossing;
        }
        toCrossing.connectsTo[directionToSource] = true;

        // Emit changes for both tiles
        this.addCrossingChange(fromQ, fromR, fromCrossing);
        this.addCrossingChange(toQ, toR, toCrossing);

        return true;
    }

    /**
     * Remove a crossing at a hex coordinate and update all connected neighbors.
     * Connected neighbors have their reciprocal connections cleared.
     * Neighbors with no remaining connections are also removed.
     *
     * @param q Q coordinate
     * @param r R coordinate
     * @returns true if crossing was removed
     */
    public removeCrossing(q: number, r: number): boolean {
        const key = `${q},${r}`;
        const crossing = this.crossings[key];

        if (!crossing) {
            return false;
        }

        // Update all connected neighbors
        for (let dir = 0; dir < 6; dir++) {
            if (crossing.connectsTo[dir]) {
                const [nq, nr] = getNeighborCoord(q, r, dir);
                const neighborKey = `${nq},${nr}`;
                const neighborCrossing = this.crossings[neighborKey];

                if (neighborCrossing) {
                    const oppositeDir = getOppositeDirection(dir);
                    neighborCrossing.connectsTo[oppositeDir] = false;

                    // Check if neighbor has any remaining connections
                    const hasConnections = neighborCrossing.connectsTo.some(c => c);
                    if (!hasConnections) {
                        // Remove neighbor crossing entirely
                        delete this.crossings[neighborKey];
                        this.addCrossingChange(nq, nr, null);
                    } else {
                        // Update neighbor crossing
                        this.addCrossingChange(nq, nr, neighborCrossing);
                    }
                }
            }
        }

        // Remove the crossing itself
        delete this.crossings[key];
        this.addCrossingChange(q, r, null);

        return true;
    }

    /**
     * Delete crossing at specified location without affecting neighbors
     * Use this for simple toggle behavior in the editor
     */
    public deleteCrossing(q: number, r: number): boolean {
        const key = `${q},${r}`;
        if (!this.crossings[key]) {
            return false;
        }
        delete this.crossings[key];
        this.addCrossingChange(q, r, null);
        return true;
    }

    /**
     * Legacy method for simple crossing placement (creates isolated crossing)
     * @deprecated Use addCrossingConnection for explicit connectivity
     */
    public setCrossing(q: number, r: number, crossingType: CrossingType): void {
        const key = `${q},${r}`;
        if (crossingType === CrossingType.CROSSING_TYPE_UNSPECIFIED) {
            this.removeCrossing(q, r);
        } else {
            // Create crossing with no connections (isolated)
            const crossing = this.createEmptyCrossing(crossingType);
            this.crossings[key] = crossing;
            this.addCrossingChange(q, r, crossing);
        }
    }

    public getAllCrossings(): Array<{ q: number; r: number; crossing: Crossing }> {
        return Object.entries(this.crossings).map(([key, crossing]) => {
            const [q, r] = key.split(',').map(Number);
            return { q, r, crossing };
        });
    }

    public clearAllCrossings(): void {
        this.crossings = {};
        this.hasUnsavedChanges = true;
    }

    // Utility methods
    public clearAll(): void {
        this.clearAllTiles();
        this.clearAllUnits();
        this.clearAllCrossings();

        this.emitStateChange(WorldEventTypes.WORLD_CLEARED, {});
    }

    /**
     * Shift all world data (tiles, units, crossings) by the given delta.
     * This re-keys everything without changing the visual appearance.
     * Useful for normalizing world coordinates after placing tiles at arbitrary positions.
     *
     * @param dQ Delta to add to all Q coordinates
     * @param dR Delta to add to all R coordinates
     */
    public shiftWorld(dQ: number, dR: number): void {
        if (dQ === 0 && dR === 0) return;

        // Shift tiles
        const oldTiles = { ...this.tiles };
        this.tiles = {};
        for (const [key, tile] of Object.entries(oldTiles)) {
            const newQ = tile.q + dQ;
            const newR = tile.r + dR;
            const newKey = `${newQ},${newR}`;
            this.tiles[newKey] = { ...tile, q: newQ, r: newR };
        }

        // Shift units
        const oldUnits = { ...this.units };
        this.units = {};
        for (const [key, unit] of Object.entries(oldUnits)) {
            const newQ = unit.q + dQ;
            const newR = unit.r + dR;
            const newKey = `${newQ},${newR}`;
            this.units[newKey] = { ...unit, q: newQ, r: newR };
        }

        // Shift crossings
        const oldCrossings = { ...this.crossings };
        this.crossings = {};
        for (const [key, crossing] of Object.entries(oldCrossings)) {
            const [q, r] = key.split(',').map(Number);
            const newQ = q + dQ;
            const newR = r + dR;
            const newKey = `${newQ},${newR}`;
            this.crossings[newKey] = crossing;
        }

        this.hasUnsavedChanges = true;

        // Emit WORLD_LOADED to trigger a full redraw instead of incremental updates
        this.emitStateChange(WorldEventTypes.WORLD_LOADED, {
            worldId: this.worldId,
            isNewWorld: this.isNewWorld,
            tileCount: Object.keys(this.tiles).length,
            unitCount: Object.keys(this.units).length
        } as WorldLoadedEventData);
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
        // Build World metadata (separate from WorldData)
        const worldMetadata: ProtoWorld = WD.from(models.World,{
            id: this.worldId || undefined,
            name: this.metadata.name || 'Untitled World',
            description: '',
            tags: [],
            difficulty: 'medium',
            creatorId: 'editor-user',
            defaultGameConfig: this.metadata.defaultGameConfig || undefined
        });

        // Build WorldData using new map-based storage
        // Convert crossings to proto format (Crossing objects with type as number)
        const crossingsForProto: Record<string, any> = {};
        for (const [key, crossing] of Object.entries(this.crossings)) {
            crossingsForProto[key] = {
                type: crossing.type as number,
                connectsTo: crossing.connectsTo
            };
        }

        const worldData: ProtoWorldData = WD.from(models.WorldData,{
            tilesMap: this.tiles,
            unitsMap: this.units,
            crossings: crossingsForProto,
            version: this.version
        });

        // Build request payload based on whether it's a new world or update
        let request: CreateWorldRequest | UpdateWorldRequest;
        let url: string;
        let method: string;

        if (this.isNewWorld) {
            // CreateWorldRequest
            request = WD.from(models.CreateWorldRequest,{
                world: worldMetadata,
                worldData: worldData
            });
            url = '/api/v1/worlds';
            method = 'POST';
        } else {
            // UpdateWorldRequest  
            const world = WD.from(models.World,{ ...worldMetadata, id: this.worldId!  })
            request = WD.from(models.UpdateWorldRequest,{
                world: world,
                worldData: worldData,
                clearWorld: false // Don't clear existing data
                // Note: update_mask is optional, omitting for now
            });
            url = `/api/v1/worlds/${this.worldId}`;
            method = 'PATCH';
        }

        // Convert protobuf request to JSON for HTTP call
        const requestJson = request
        // const requestJson = this.isNewWorld ? models.CreateWorldRequest.from(request): models.UpdateWorldRequest.from(request);

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

            // Update version from response for optimistic locking
            if (result.worldData?.version !== undefined) {
                this.version = result.worldData.version;
            }

            this.markAsSaved();

            this.emitStateChange(WorldEventTypes.WORLD_SAVED, {
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
        // Support both old array format and new map format
        const tilesSource = worldTilesData.tilesMap || worldTilesData.tiles_map || worldTilesData.tiles || {};
        const unitsSource = worldTilesData.unitsMap || worldTilesData.units_map || worldTilesData.units || {};

        const combinedData = {
            // World metadata
            name: worldMetadata.name || 'Untitled World',
            Name: worldMetadata.name || 'Untitled World', // Both for compatibility
            id: worldMetadata.id,
            defaultGameConfig: worldMetadata.defaultGameConfig || worldMetadata.default_game_config,

            // Calculate dimensions from tiles if present
            width: 40,  // Default
            height: 40, // Default

            // World tiles, units, and crossings - pass through for loadFromData to handle
            tilesMap: tilesSource,
            unitsMap: unitsSource,
            crossings: worldTilesData.crossings || {},

            // Version for optimistic locking
            version: worldTilesData.version || 0
        };

        // Calculate actual dimensions from tile bounds
        const tileValues = Array.isArray(tilesSource) ? tilesSource : Object.values(tilesSource);
        if (tileValues.length > 0) {
            let maxQ = 0, maxR = 0, minQ = 0, minR = 0;
            tileValues.forEach((tile: any) => {
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
        this.crossings = {};

        // Load metadata - handle both old and new formats
        if (data.name) this.metadata.name = data.name;
        if (data.Name) this.metadata.name = data.Name; // Backend format
        if (data.width) this.metadata.width = data.width;
        if (data.height) this.metadata.height = data.height;
        if (data.defaultGameConfig) this.metadata.defaultGameConfig = data.defaultGameConfig;

        // Load version for optimistic locking
        if (data.version !== undefined) this.version = data.version;

        // Determine tiles and units source - prefer new map format, fall back to array format
        let tiles = data.tilesMap || data.tiles_map || data.tiles || [];
        let units = data.unitsMap || data.units_map || data.units || [];

        // Load crossings if present
        const crossings = data.crossings || {};
        for (const [key, value] of Object.entries(crossings)) {
            const crossingData = value as any;
            this.crossings[key] = {
                type: crossingData.type ?? CrossingType.CROSSING_TYPE_UNSPECIFIED,
                connectsTo: crossingData.connectsTo || crossingData.connects_to || [false, false, false, false, false, false]
            };
        }

        return this.loadTilesAndUnits(tiles, units);
    }

    loadTilesAndUnits(tiles: any, units: any): World {
        // Batch load tiles - handle both array and map formats
        const tileChanges: TileChange[] = [];
        tiles = tiles || {};

        // Process tiles - check if it's array (old format) or map (new format)
        const tileEntries = Array.isArray(tiles)
            ? tiles.map((t: any) => [`${t.q || 0},${t.r || 0}`, t])
            : Object.entries(tiles);

        for (const [key, tileData] of tileEntries) {
            // Parse coordinates from key if needed (map format uses "q,r" keys)
            const [keyQ, keyR] = key.split(',').map(Number);
            const q = (tileData as any).q ?? keyQ ?? 0;
            const r = (tileData as any).r ?? keyR ?? 0;
            const coordKey = `${q},${r}`;
            let tileType: number;
            let player = 0;

            if ((tileData as any).tileType !== undefined) {
                tileType = (tileData as any).tileType;
                player = (tileData as any).player || 0;
            } else if ((tileData as any).tile_type !== undefined) {
                tileType = (tileData as any).tile_type;
                player = (tileData as any).player || 0;
            } else {
                continue; // Skip invalid tile
            }

            const tile: Tile = { q, r, tileType, player, shortcut: (tileData as any).shortcut || "", lastActedTurn: 0, lastToppedupTurn: 0 };
            this.tiles[coordKey] = tile;
            tileChanges.push({ q, r, tile });
        }

        // Batch load units - handle both array and map formats
        const unitChanges: UnitChange[] = [];
        units = units || {};

        // Process units - check if it's array (old format) or map (new format)
        const unitEntries = Array.isArray(units)
            ? units.map((u: any) => [`${u.q || 0},${u.r || 0}`, u])
            : Object.entries(units);

        for (const [key, unitData] of unitEntries) {
            // Parse coordinates from key if needed (map format uses "q,r" keys)
            const [keyQ, keyR] = key.split(',').map(Number);
            const q = (unitData as any).q ?? keyQ ?? 0;
            const r = (unitData as any).r ?? keyR ?? 0;
            const coordKey = `${q},${r}`;
            let unitType: number, player: number;

            if ((unitData as any).unitType !== undefined && (unitData as any).player !== undefined) {
                unitType = (unitData as any).unitType;
                player = (unitData as any).player;
            } else if ((unitData as any).unit_type !== undefined && (unitData as any).player !== undefined) {
                unitType = (unitData as any).unit_type;
                player = (unitData as any).player;
            } else {
                continue; // Skip invalid unit
            }

            // Pass all unit data, including runtime state (distanceLeft, availableHealth, etc.)
            const unit: Unit = WD.from(models.Unit,{
                q,
                r,
                unitType,
                player,
                shortcut: (unitData as any).shortcut || "",
                // Include runtime state
                availableHealth: (unitData as any).availableHealth || (unitData as any).available_health || 10,
                distanceLeft: ((unitData as any).distanceLeft && (unitData as any).distanceLeft > 0) ? (unitData as any).distanceLeft :
                             ((unitData as any).distance_left && (unitData as any).distance_left > 0) ? (unitData as any).distance_left : 3,
                turnCounter: (unitData as any).turnCounter || (unitData as any).turn_counter || 1
            });
            this.units[coordKey] = unit;
            unitChanges.push({ q, r, unit });
        }
        
        // Emit batched changes immediately
        if (tileChanges.length > 0) {
            this.emitStateChange(WorldEventTypes.TILES_CHANGED, {
                changes: tileChanges
            } as TilesChangedEventData);
        }
        
        if (unitChanges.length > 0) {
            this.emitStateChange(WorldEventTypes.UNITS_CHANGED, {
                changes: unitChanges
            } as UnitsChangedEventData);
        }
        
        this.hasUnsavedChanges = false;
        
        this.emitStateChange(WorldEventTypes.WORLD_LOADED, {
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

    public radialNeighbours(q: number, r: number, radius: number): [number, number][] {
        const minq = q - radius;
        const maxq = q + radius;
        const minr = r - radius;
        const maxr = r + radius;
        const out = [] as [number, number][]
        for (let bq = minq; bq <= maxq; bq++) {
            for (let br = minr; br <= maxr; br++) {
                // Use cube distance to determine if tile is within brush radius
                const distance = Math.abs(bq - q) + Math.abs(br - r) + Math.abs(-bq - br - (-q - r));
                if (distance <= radius * 2) { // Hex distance formula
                  out.push([bq, br])
                }
            }
        }
        return out
    }

    /**
     * Get tiles in a rectangular region based on row/col coordinates
     * @param startQ Starting Q coordinate
     * @param startR Starting R coordinate
     * @param endQ Ending Q coordinate
     * @param endR Ending R coordinate
     * @param filled If true, returns all tiles in the rectangle. If false, returns only outline tiles.
     * @returns Array of [q, r] coordinate tuples
     */
    public rectFrom(startQ: number, startR: number, endQ: number, endR: number, filled: boolean = true): [number, number][] {
        // Convert to row/col coordinates for proper rectangular selection
        const startRowCol = hexToRowCol(startQ, startR);
        const endRowCol = hexToRowCol(endQ, endR);

        const minRow = Math.min(startRowCol.row, endRowCol.row);
        const maxRow = Math.max(startRowCol.row, endRowCol.row);
        const minCol = Math.min(startRowCol.col, endRowCol.col);
        const maxCol = Math.max(startRowCol.col, endRowCol.col);

        const out: [number, number][] = [];

        for (let row = minRow; row <= maxRow; row++) {
            for (let col = minCol; col <= maxCol; col++) {
                // If not filled, only include outline tiles
                if (!filled) {
                    if (row !== minRow && row !== maxRow && col !== minCol && col !== maxCol) {
                        continue; // Skip interior tiles
                    }
                }

                // Convert back to hex coordinates
                const hex = rowColToHex(row, col);
                out.push([hex.q, hex.r]);
            }
        }

        return out;
    }

    /**
     * Generate tiles for a circle in hex coordinates
     * @param centerQ Center Q coordinate
     * @param centerR Center R coordinate
     * @param radius Radius in hex tiles
     * @param filled If true, fill the circle; if false, only return outline
     * @returns Array of [q, r] coordinate pairs
     */
    public circleFrom(centerQ: number, centerR: number, radius: number, filled: boolean = true): [number, number][] {
        const out: [number, number][] = [];

        // Scan bounding box
        for (let q = centerQ - radius; q <= centerQ + radius; q++) {
            for (let r = centerR - radius; r <= centerR + radius; r++) {
                const dist = hexDistance(centerQ, centerR, q, r);

                if (filled) {
                    // Include all tiles within radius
                    if (dist <= radius) {
                        out.push([q, r]);
                    }
                } else {
                    // Only include outline tiles (distance exactly equals radius)
                    if (dist === radius) {
                        out.push([q, r]);
                    }
                }
            }
        }

        return out;
    }

    /**
     * Generate tiles for an axis-aligned oval/ellipse in hex coordinates
     * @param centerQ Center Q coordinate
     * @param centerR Center R coordinate
     * @param radiusX Horizontal radius in row/col space
     * @param radiusY Vertical radius in row/col space
     * @param filled If true, fill the oval; if false, only return outline
     * @returns Array of [q, r] coordinate pairs
     */
    public ovalFrom(centerQ: number, centerR: number, radiusX: number, radiusY: number, filled: boolean = true): [number, number][] {
        const out: [number, number][] = [];

        // Convert center to row/col
        const centerRowCol = hexToRowCol(centerQ, centerR);

        // Scan bounding box in row/col space
        const minRow = Math.floor(centerRowCol.row - radiusY);
        const maxRow = Math.ceil(centerRowCol.row + radiusY);
        const minCol = Math.floor(centerRowCol.col - radiusX);
        const maxCol = Math.ceil(centerRowCol.col + radiusX);

        for (let row = minRow; row <= maxRow; row++) {
            for (let col = minCol; col <= maxCol; col++) {
                // Ellipse formula: (dx/radiusX)^2 + (dy/radiusY)^2 <= 1
                const dx = col - centerRowCol.col;
                const dy = row - centerRowCol.row;
                const normalizedDist = (dx * dx) / (radiusX * radiusX) + (dy * dy) / (radiusY * radiusY);

                if (filled) {
                    // Include all tiles within ellipse
                    if (normalizedDist <= 1) {
                        const hex = rowColToHex(row, col);
                        out.push([hex.q, hex.r]);
                    }
                } else {
                    // Outline: tiles on the edge (within a small threshold of boundary)
                    // Use threshold to account for discretization
                    if (normalizedDist >= 0.85 && normalizedDist <= 1.15) {
                        const hex = rowColToHex(row, col);
                        out.push([hex.q, hex.r]);
                    }
                }
            }
        }

        return out;
    }

    /**
     * Generate tiles for a line/path through multiple points using Bresenham algorithm
     * @param points Array of {q, r} coordinates defining the path
     * @returns Array of [q, r] coordinate pairs for all tiles along the line
     */
    public lineFrom(points: { q: number; r: number }[]): [number, number][] {
        if (points.length < 2) {
            return points.map(p => [p.q, p.r] as [number, number]);
        }

        const out: [number, number][] = [];
        const visited = new Set<string>();

        // Draw line segment between each consecutive pair of points
        for (let i = 0; i < points.length - 1; i++) {
            const start = points[i];
            const end = points[i + 1];

            // Bresenham line algorithm in hex coordinates
            // Convert to row/col for linear interpolation
            const startRowCol = hexToRowCol(start.q, start.r);
            const endRowCol = hexToRowCol(end.q, end.r);

            const dx = endRowCol.col - startRowCol.col;
            const dy = endRowCol.row - startRowCol.row;
            const steps = Math.max(Math.abs(dx), Math.abs(dy));

            for (let step = 0; step <= steps; step++) {
                const t = steps === 0 ? 0 : step / steps;
                const col = Math.round(startRowCol.col + t * dx);
                const row = Math.round(startRowCol.row + t * dy);

                const hex = rowColToHex(row, col);
                const key = `${hex.q},${hex.r}`;

                if (!visited.has(key)) {
                    visited.add(key);
                    out.push([hex.q, hex.r]);
                }
            }
        }

        return out;
    }

    public floodNeighbors(q: number, r: number, radius: number): [number, number][] {
        let queue = [[q,r]] as [number, number][]
        const minq = q - radius;
        const maxq = q + radius;
        const minr = r - radius;
        const maxr = r + radius;
        const startingTile = this.getTileAt(q, r)
        const visited = {} as any
        visited[q + ":" + r] = true
        for (var i = 0;i < queue.length;i++) {
            const [nextq, nextr] = queue[i];
            const key = nextq + ":" + nextr
            console.log("Visiting: ", key)

            // go through all children
            const neighbors = axialNeighbors(nextq, nextr)
            for (var j = 0;j < 6;j++) {
               const [cq, cr] = neighbors[j]
               if (cq >= minq && cq <= maxq && cr >= minr && cr <= maxr) {
                  const ckey = cq + ":" + cr
                  if (!visited[ckey]) {
                      const childTile = this.getTileAt(cq, cr)
                      if (childTile == startingTile || (
                              childTile != null && startingTile != null &&
                              childTile.tileType == startingTile.tileType &&
                              childTile.player == startingTile.player)) {
                          queue.push([cq, cr])
                          visited[cq + ":" + cr] = true
                      }
                  }
               }
            }
        }
        return queue
    }
}
