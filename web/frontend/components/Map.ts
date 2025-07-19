import { HexCoord } from './phaser/hexUtils';

export interface TileData {
    tileType: number;
    playerId?: number; // Player ID for city terrains that support ownership
}

export interface UnitData {
    unitType: number;
    playerId: number;
}

export interface MapMetadata {
    name: string;
    width: number;
    height: number;
}

/**
 * Map class handles all map data management including tiles, units, and metadata.
 * This centralizes map state to reduce coupling with UI components.
 */
export class Map {
    private metadata: MapMetadata;
    private tiles: { [key: string]: TileData } = {};
    private units: { [key: string]: UnitData } = {};
    
    constructor(name: string = 'New Map', width: number = 40, height: number = 40) {
        this.metadata = { name, width, height };
    }
    
    // Map metadata methods
    public getName(): string {
        return this.metadata.name;
    }
    
    public setName(name: string): void {
        this.metadata.name = name;
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
    
    public getMetadata(): MapMetadata {
        return { ...this.metadata };
    }
    
    // Tile management methods
    public tileExistsAt(q: number, r: number): boolean {
        const key = `${q},${r}`;
        return key in this.tiles;
    }
    
    public getTileAt(q: number, r: number): TileData | null {
        const key = `${q},${r}`;
        return this.tiles[key] || null;
    }
    
    public setTileAt(q: number, r: number, tileType: number, playerId?: number): void {
        const key = `${q},${r}`;
        this.tiles[key] = { tileType, playerId };
    }
    
    public removeTileAt(q: number, r: number): boolean {
        const key = `${q},${r}`;
        if (key in this.tiles) {
            delete this.tiles[key];
            return true;
        }
        return false;
    }
    
    public getAllTiles(): Array<{ q: number; r: number; tileType: number; playerId?: number }> {
        const result: Array<{ q: number; r: number; tileType: number; playerId?: number }> = [];
        
        Object.entries(this.tiles).forEach(([key, tileData]) => {
            const [q, r] = key.split(',').map(Number);
            result.push({
                q,
                r,
                tileType: tileData.tileType,
                playerId: tileData.playerId
            });
        });
        
        return result;
    }
    
    public clearAllTiles(): void {
        this.tiles = {};
    }
    
    // Unit management methods
    public unitExistsAt(q: number, r: number): boolean {
        const key = `${q},${r}`;
        return key in this.units;
    }
    
    public getUnitAt(q: number, r: number): UnitData | null {
        const key = `${q},${r}`;
        return this.units[key] || null;
    }
    
    public setUnitAt(q: number, r: number, unitType: number, playerId: number): void {
        const key = `${q},${r}`;
        this.units[key] = { unitType, playerId };
    }
    
    public removeUnitAt(q: number, r: number): boolean {
        const key = `${q},${r}`;
        if (key in this.units) {
            delete this.units[key];
            return true;
        }
        return false;
    }
    
    public getAllUnits(): Array<{ q: number; r: number; unitType: number; playerId: number }> {
        const result: Array<{ q: number; r: number; unitType: number; playerId: number }> = [];
        
        Object.entries(this.units).forEach(([key, unitData]) => {
            const [q, r] = key.split(',').map(Number);
            result.push({
                q,
                r,
                unitType: unitData.unitType,
                playerId: unitData.playerId
            });
        });
        
        return result;
    }
    
    public clearAllUnits(): void {
        this.units = {};
    }
    
    // Utility methods
    public clearAll(): void {
        this.clearAllTiles();
        this.clearAllUnits();
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
    
    // Serialization methods
    public serialize(): {
        name: string;
        width: number;
        height: number;
        tiles: { [key: string]: { tileType: number; playerId?: number } };
        units: { [key: string]: { unitType: number; playerId: number } };
    } {
        return {
            name: this.metadata.name,
            width: this.metadata.width,
            height: this.metadata.height,
            tiles: { ...this.tiles },
            units: { ...this.units }
        };
    }
    
    public static deserialize(data: any): Map {
        const map = new Map(data.name || 'Untitled Map', data.width || 40, data.height || 40);
        
        // Handle tiles - support both tileType and tile_type formats
        if (data.tiles) {
            Object.entries(data.tiles).forEach(([key, tileData]: [string, any]) => {
                let tileType: number;
                let playerId: number | undefined;
                
                if (tileData.tileType !== undefined) {
                    tileType = tileData.tileType;
                    playerId = tileData.playerId;
                } else if (tileData.tile_type !== undefined) {
                    tileType = tileData.tile_type;
                    playerId = tileData.player || 0;
                } else {
                    console.warn(`Map.deserialize: No tileType found for tile at ${key}:`, tileData);
                    return; // Skip this tile
                }
                
                map.tiles[key] = { tileType, playerId };
            });
        }
        
        // Handle units - support both unitType/playerId and unit_type/player formats
        if (data.units) {
            Object.entries(data.units).forEach(([key, unitData]: [string, any]) => {
                let unitType: number, playerId: number;
                
                if (unitData.unitType !== undefined && unitData.playerId !== undefined) {
                    unitType = unitData.unitType;
                    playerId = unitData.playerId;
                } else if (unitData.unit_type !== undefined && unitData.player !== undefined) {
                    unitType = unitData.unit_type;
                    playerId = unitData.player;
                } else {
                    console.warn(`Map.deserialize: No unitType/playerId found for unit at ${key}:`, unitData);
                    return; // Skip this unit
                }
                
                map.units[key] = { unitType, playerId };
            });
        }
        
        // Handle map_units array format (from server)
        if (data.map_units && Array.isArray(data.map_units)) {
            data.map_units.forEach((unit: any) => {
                if (unit.q !== undefined && unit.r !== undefined && unit.unit_type !== undefined) {
                    const key = `${unit.q},${unit.r}`;
                    map.units[key] = {
                        unitType: unit.unit_type,
                        playerId: unit.player || 1
                    };
                }
            });
        }
        
        return map;
    }
    
    // Map validation
    public validate(): { isValid: boolean; errors: string[] } {
        const errors: string[] = [];
        
        // Check metadata
        if (!this.metadata.name || this.metadata.name.trim() === '') {
            errors.push('Map name cannot be empty');
        }
        
        if (this.metadata.width <= 0 || this.metadata.height <= 0) {
            errors.push('Map dimensions must be positive');
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
            if (unitData.playerId < 0 || unitData.playerId > 12) {
                errors.push(`Invalid player ID at ${key}: ${unitData.playerId}`);
            }
        });
        
        return {
            isValid: errors.length === 0,
            errors
        };
    }
    
    // Clone method for safe copying
    public clone(): Map {
        return Map.deserialize(this.serialize());
    }
}