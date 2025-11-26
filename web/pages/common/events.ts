

// Define common event types for type safety
export const WorldEventTypes = {
    WORLD_DATA_LOADED: 'world-data-loaded',
    WORLD_STATS_UPDATED: 'world-stats-updated',
    WORLD_VIEWER_READY: 'world-viewer-ready',
    TILES_CHANGED: 'tiles-changed',      // Batch tile operations
    UNITS_CHANGED: 'units-changed',      // Batch unit operations
    WORLD_LOADED: 'world-loaded',
    WORLD_SAVED: 'world-saved',
    WORLD_CLEARED: 'world-cleared',
    WORLD_METADATA_CHANGED: 'world-metadata-changed'
} as const;

// Editor-specific event types for WorldEditorPage components
export const EditorEventTypes = {
    // Editor state events
    TERRAIN_SELECTED: 'terrain-selected',
    UNIT_SELECTED: 'unit-selected',
    BRUSH_SIZE_CHANGED: 'brush-size-changed',
    PLACEMENT_MODE_CHANGED: 'placement-mode-changed',
    PLAYER_CHANGED: 'player-changed',
    
    // World events
    WORLD_LOADED: 'world-loaded',
    WORLD_SAVED: 'world-saved',
    WORLD_CHANGED: 'world-changed',
    
    // UI events
    STATS_REFRESH_REQUESTED: 'stats-refresh-requested',
    KEYBOARD_SHORTCUT_TRIGGERED: 'keyboard-shortcut-triggered',
    
    // Phaser events
    PHASER_READY: 'phaser-ready',
    PHASER_ERROR: 'phaser-error',
    
    // Tools events
    TOOLS_UI_UPDATED: 'tools-ui-updated',
    GRID_TOGGLE: 'grid-toggle',
    COORDINATES_TOGGLE: 'coordinates-toggle'
} as const;

// Event payload type definitions for type safety
export interface WorldDataLoadedPayload {
    worldId: string;
    totalTiles: number;
    totalUnits: number;
    bounds: {
        minQ: number;
        maxQ: number;
        minR: number;
        maxR: number;
    };
    terrainCounts: { [terrainType: number]: number };
}

export interface WorldStatsUpdatedPayload {
    totalTiles: number;
    totalUnits: number;
    dimensions: {
        width: number;
        height: number;
    };
    terrainDistribution: {
        [terrainType: number]: {
            count: number;
            percentage: number;
            name: string;
        };
    };
}

export interface ComponentErrorPayload {
    componentId: string;
    error: string;
    details?: any;
}

// Editor-specific event payload interfaces
export interface TerrainSelectedPayload {
    terrainType: number;
    terrainName: string;
}

export interface UnitSelectedPayload {
    unitType: number;
    unitName: string;
    playerId: number;
}

export interface BrushSizeChangedPayload {
    brushSize: number;
    sizeName: string;
}

export interface PlacementModeChangedPayload {
    mode: 'terrain' | 'unit' | 'clear';
}

export interface PlayerChangedPayload {
    playerId: number;
}

export interface TileClickedPayload {
    q: number;
    r: number;
}

export interface PhaserReadyPayload {
    // Empty payload - just signals that Phaser is ready
}

export interface KeyboardShortcutPayload {
    command: string;
    args?: any;
}

export interface GridTogglePayload {
    showGrid: boolean;
}

export interface CoordinatesTogglePayload {
    showCoordinates: boolean;
}

// World modification event payloads
export interface TilePaintedPayload {
    q: number;
    r: number;
    terrainType: number;
    playerColor: number;
    brushSize: number;
}

export interface UnitPlacedPayload {
    q: number;
    r: number;
    unitType: number;
    playerId: number;
}

export interface TileClearedPayload {
    q: number;
    r: number;
}

export interface UnitRemovedPayload {
    q: number;
    r: number;
}

export type WorldEventType = typeof WorldEventTypes[keyof typeof WorldEventTypes];
export type EditorEventType = typeof EditorEventTypes[keyof typeof EditorEventTypes];
