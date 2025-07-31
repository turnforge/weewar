

// Define common event types for type safety
export const WorldEventTypes = {
    WORLD_DATA_LOADED: 'world-data-loaded',
    WORLD_DATA_ERROR: 'world-data-error', 
    WORLD_STATS_UPDATED: 'world-stats-updated',
    WORLD_VIEWER_READY: 'world-viewer-ready',
    WORLD_VIEWER_ERROR: 'world-viewer-error',
} as const;

export const GameEventTypes = {
    GAME_DATA_LOADED: 'game-data-loaded',
    GAME_DATA_ERROR: 'game-data-error', 
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
    TILE_CLICKED: 'tile-clicked',
    
    // UI events
    STATS_REFRESH_REQUESTED: 'stats-refresh-requested',
    KEYBOARD_SHORTCUT_TRIGGERED: 'keyboard-shortcut-triggered',
    
    // Phaser events
    PHASER_READY: 'phaser-ready',
    PHASER_ERROR: 'phaser-error',
    
    // Reference image events (ReferenceImagePanel → PhaserEditorComponent)
    REFERENCE_LOAD_FROM_CLIPBOARD: 'reference-load-from-clipboard',
    REFERENCE_LOAD_FROM_FILE: 'reference-load-from-file',
    REFERENCE_SET_MODE: 'reference-set-mode',
    REFERENCE_SET_ALPHA: 'reference-set-alpha',
    REFERENCE_SET_POSITION: 'reference-set-position',
    REFERENCE_SET_SCALE: 'reference-set-scale',
    REFERENCE_CLEAR: 'reference-clear',
    
    // Reference image events (PhaserEditorComponent → ReferenceImagePanel)
    REFERENCE_SCALE_CHANGED: 'reference-scale-changed',
    REFERENCE_STATE_CHANGED: 'reference-state-changed',
    REFERENCE_ALPHA_CHANGED: 'reference-alpha-changed',
    REFERENCE_MODE_CHANGED: 'reference-mode-changed',
    REFERENCE_IMAGE_LOADED: 'reference-image-loaded',
    
    // World modification events
    TILE_PAINTED: 'tile-painted',
    UNIT_PLACED: 'unit-placed',
    TILE_CLEARED: 'tile-cleared',
    UNIT_REMOVED: 'unit-removed',
    
    // Tools events
    TOOLS_UI_UPDATED: 'tools-ui-updated',
    GRID_TOGGLE: 'grid-toggle',
    COORDINATES_TOGGLE: 'coordinates-toggle',
    GRID_SET_VISIBILITY: 'grid-set-visibility',
    COORDINATES_SET_VISIBILITY: 'coordinates-set-visibility',
    
    // Page state events (centralized state management)
    PAGE_STATE_CHANGED: 'page-state-changed',
    TOOL_STATE_CHANGED: 'tool-state-changed',
    VISUAL_STATE_CHANGED: 'visual-state-changed',
    WORKFLOW_STATE_CHANGED: 'workflow-state-changed'
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

export interface GridSetVisibilityPayload {
    show: boolean;
}

export interface CoordinatesSetVisibilityPayload {
    show: boolean;
}

// Reference image event payloads

// ReferenceImagePanel → PhaserEditorComponent events
export interface ReferenceLoadFromFilePayload {
    file: File;
}

export interface ReferenceSetModePayload {
    mode: number; // 0=hidden, 1=background, 2=overlay
}

export interface ReferenceSetAlphaPayload {
    alpha: number; // 0.0 to 1.0
}

export interface ReferenceSetPositionPayload {
    x: number;
    y: number;
}

export interface ReferenceSetScalePayload {
    scaleX: number;
    scaleY: number;
}

// PhaserEditorComponent → ReferenceImagePanel events
export interface ReferenceScaleChangedPayload {
    scaleX: number;
    scaleY: number;
}

export interface ReferenceStateChangedPayload {
    scale: { x: number; y: number };
    position: { x: number; y: number };
    alpha: number;
    mode: number;
    isLoaded: boolean;
}

export interface ReferencePositionChangedPayload {
    x: number;
    y: number;
}

export interface ReferenceAlphaChangedPayload {
    alpha: number;
}

export interface ReferenceModeChangedPayload {
    mode: number; // 0=hidden, 1=background, 2=overlay
}

export interface ReferenceImageLoadedPayload {
    source: string; // 'clipboard' | 'file' | filename
    width: number;
    height: number;
    url: string; // Object URL for the image
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

// Page state event payloads (import from WorldEditorPageState.ts)
export interface PageStateChangedPayload {
    eventType: 'tool-state-changed' | 'visual-state-changed' | 'workflow-state-changed';
    data: any; // Will be ToolStateChangedEventData | VisualStateChangedEventData | WorkflowStateChangedEventData
}

export type WorldEventType = typeof WorldEventTypes[keyof typeof WorldEventTypes];
export type EditorEventType = typeof EditorEventTypes[keyof typeof EditorEventTypes];
