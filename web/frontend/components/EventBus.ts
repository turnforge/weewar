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
    COMPONENT_ERROR: 'component-error'
} as const;

export type EventType = typeof EventTypes[keyof typeof EventTypes];

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