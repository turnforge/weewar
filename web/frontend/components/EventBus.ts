/**
 * Type-safe event system for component communication
 * Provides error isolation and prevents circular event propagation
 */

export interface EventPayload<T = any> {
    source: string;
    timestamp: number;
    data: T;
}

export type EventHandler<T = any> = (payload: EventPayload<T>) => void;

interface EventSubscription {
    handler: EventHandler;
    componentId: string;
}

/**
 * Synchronous event bus for component communication
 * Features:
 * - Type-safe event names and payloads
 * - Error isolation (one handler failure doesn't stop others)
 * - Source exclusion (events not sent back to source)
 * - Debug logging for troubleshooting
 */
export class EventBus {
    private subscribers: Map<string, EventSubscription[]> = new Map();
    private debugMode: boolean = false;
    
    constructor(debugMode: boolean = false) {
        this.debugMode = debugMode;
    }
    
    /**
     * Subscribe to an event type
     * @param eventType - The event type to listen for
     * @param handler - Function to call when event is fired
     * @param componentId - ID of the subscribing component (for source exclusion)
     */
    public subscribe<T = any>(
        eventType: string, 
        handler: EventHandler<T>, 
        componentId: string
    ): () => void {
        if (!this.subscribers.has(eventType)) {
            this.subscribers.set(eventType, []);
        }
        
        const subscription: EventSubscription = { handler, componentId };
        this.subscribers.get(eventType)!.push(subscription);
        
        if (this.debugMode) {
            console.log(`[EventBus] Component '${componentId}' subscribed to '${eventType}'`);
        }
        
        // Return unsubscribe function
        return () => this.unsubscribe(eventType, componentId, handler);
    }
    
    /**
     * Unsubscribe from an event type
     */
    public unsubscribe<T = any>(
        eventType: string, 
        componentId: string, 
        handler: EventHandler<T>
    ): void {
        const subscriptions = this.subscribers.get(eventType);
        if (!subscriptions) return;
        
        const index = subscriptions.findIndex(
            sub => sub.componentId === componentId && sub.handler === handler
        );
        
        if (index !== -1) {
            subscriptions.splice(index, 1);
            if (this.debugMode) {
                console.log(`[EventBus] Component '${componentId}' unsubscribed from '${eventType}'`);
            }
        }
        
        // Clean up empty event types
        if (subscriptions.length === 0) {
            this.subscribers.delete(eventType);
        }
    }
    
    /**
     * Emit an event to all subscribers (except the source)
     * @param eventType - The event type to emit
     * @param data - The event data payload
     * @param sourceComponentId - ID of the component emitting the event
     */
    public emit<T = any>(eventType: string, data: T, sourceComponentId: string): void {
        const subscriptions = this.subscribers.get(eventType);
        if (!subscriptions || subscriptions.length === 0) {
            if (this.debugMode) {
                console.log(`[EventBus] No subscribers for event '${eventType}'`);
            }
            return;
        }
        
        const payload: EventPayload<T> = {
            source: sourceComponentId,
            timestamp: Date.now(),
            data
        };
        
        if (this.debugMode) {
            console.log(`[EventBus] Emitting '${eventType}' from '${sourceComponentId}' to ${subscriptions.length} subscribers`);
        }
        
        let successCount = 0;
        let errorCount = 0;
        
        // Call each handler with error isolation
        subscriptions.forEach(subscription => {
            // Source exclusion - don't send event back to the source
            if (subscription.componentId === sourceComponentId) {
                if (this.debugMode) {
                    console.log(`[EventBus] Skipping source component '${sourceComponentId}'`);
                }
                return;
            }
            
            try {
                subscription.handler(payload);
                successCount++;
            } catch (error) {
                errorCount++;
                console.error(
                    `[EventBus] Error in event handler for '${eventType}' ` +
                    `in component '${subscription.componentId}':`, 
                    error
                );
                // Continue with other handlers - error isolation
            }
        });
        
        if (this.debugMode) {
            console.log(
                `[EventBus] Event '${eventType}' completed: ` +
                `${successCount} success, ${errorCount} errors`
            );
        }
    }
    
    /**
     * Get all event types that have subscribers
     */
    public getEventTypes(): string[] {
        return Array.from(this.subscribers.keys());
    }
    
    /**
     * Get subscriber count for an event type
     */
    public getSubscriberCount(eventType: string): number {
        return this.subscribers.get(eventType)?.length || 0;
    }
    
    /**
     * Clear all subscriptions (useful for cleanup)
     */
    public clear(): void {
        this.subscribers.clear();
        if (this.debugMode) {
            console.log('[EventBus] All subscriptions cleared');
        }
    }
    
    /**
     * Enable or disable debug logging
     */
    public setDebugMode(enabled: boolean): void {
        this.debugMode = enabled;
    }
}

// Define common event types for type safety
export const EventTypes = {
    MAP_DATA_LOADED: 'map-data-loaded',
    MAP_DATA_ERROR: 'map-data-error', 
    MAP_STATS_UPDATED: 'map-stats-updated',
    MAP_VIEWER_READY: 'map-viewer-ready',
    MAP_VIEWER_ERROR: 'map-viewer-error',
    COMPONENT_INITIALIZED: 'component-initialized',
    COMPONENT_HYDRATED: 'component-hydrated',
    COMPONENT_ERROR: 'component-error'
} as const;

// Editor-specific event types for MapEditorPage components
export const EditorEventTypes = {
    // Editor state events
    TERRAIN_SELECTED: 'terrain-selected',
    UNIT_SELECTED: 'unit-selected',
    BRUSH_SIZE_CHANGED: 'brush-size-changed',
    PLACEMENT_MODE_CHANGED: 'placement-mode-changed',
    PLAYER_CHANGED: 'player-changed',
    
    // Map events
    MAP_LOADED: 'map-loaded',
    MAP_SAVED: 'map-saved',
    MAP_CHANGED: 'map-changed',
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
    
    // Map modification events
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

export type EventType = typeof EventTypes[keyof typeof EventTypes];
export type EditorEventType = typeof EditorEventTypes[keyof typeof EditorEventTypes];

// Event payload type definitions for type safety
export interface MapDataLoadedPayload {
    mapId: string;
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

export interface MapStatsUpdatedPayload {
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

// Map modification event payloads
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

// Page state event payloads (import from MapEditorPageState.ts)
export interface PageStateChangedPayload {
    eventType: 'tool-state-changed' | 'visual-state-changed' | 'workflow-state-changed';
    data: any; // Will be ToolStateChangedEventData | VisualStateChangedEventData | WorkflowStateChangedEventData
}