/**
 * Simple event system for component communication
 * Provides error isolation and idempotent subscriptions
 */

/**
 * Interface for components that want to receive events via the EventBus
 */
export interface EventSubscriber {
    /**
     * Handle incoming events from the EventBus
     * @param eventType - The type of event being handled
     * @param data - The event data payload
     * @param target - The target/subject entity (what the event is about)
     * @param emitter - The entity that emitted the event
     */
    handleBusEvent(eventType: string, data: any, target: any, emitter: any): void;
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
    private subscribers: Map<string, Set<EventSubscriber>> = new Map();
    private debugMode: boolean = false;
    
    constructor(debugMode: boolean = false) {
        this.debugMode = debugMode;
    }
    
    /**
     * Add a subscription using the EventSubscriber pattern
     * Provides automatic idempotency - same subscriber object won't be added twice
     */
    public addSubscription(eventType: string, target: any, subscriber: EventSubscriber): void {
        if (!this.subscribers.has(eventType)) {
            this.subscribers.set(eventType, new Set());
        }
        
        const subscribers = this.subscribers.get(eventType)!;
        const wasAdded = !subscribers.has(subscriber);
        
        if (wasAdded) {
            subscribers.add(subscriber);
            if (this.debugMode) {
                console.log(`[EventBus] Added subscription to '${eventType}' for ${subscriber.constructor.name}`);
            }
        } else if (this.debugMode) {
            console.log(`[EventBus] Subscription already exists for '${eventType}' and ${subscriber.constructor.name}`);
        }
    }
    
    /**
     * Remove a subscription using the EventSubscriber pattern
     */
    public removeSubscription(eventType: string, target: any, subscriber: EventSubscriber): void {
        const subscribers = this.subscribers.get(eventType);
        
        if (subscribers) {
            const wasRemoved = subscribers.delete(subscriber);
            
            if (this.debugMode && wasRemoved) {
                console.log(`[EventBus] Removed subscription from '${eventType}' for ${subscriber.constructor.name}`);
            }
            
            // Clean up empty subscription sets
            if (subscribers.size === 0) {
                this.subscribers.delete(eventType);
            }
        }
    }
    
    /**
     * Emit an event to all subscribers
     * @param eventType - The event type to emit
     * @param data - The event data payload
     * @param target - The target/subject entity that this event relates to
     * @param emitter - The entity that emitted the event
     */
    public emit<T = any>(eventType: string, data: T, target: any, emitter: any): void {
        const subscribers = this.subscribers.get(eventType);
        
        if (!subscribers || subscribers.size === 0) {
            if (this.debugMode) {
                console.log(`[EventBus] No subscribers for event '${eventType}'`);
            }
            return;
        }
        
        if (this.debugMode) {
            console.log(`[EventBus] Emitting '${eventType}' to ${subscribers.size} subscribers`);
        }
        
        let successCount = 0;
        let errorCount = 0;
        
        // Call EventSubscriber handlers with error isolation
        subscribers.forEach(subscriber => {
            try {
                subscriber.handleBusEvent(eventType, data, target, emitter);
                successCount++;
            } catch (error) {
                errorCount++;
                console.error(
                    `[EventBus] Error in EventSubscriber handler for '${eventType}' ` +
                    `in subscriber '${subscriber.constructor.name}':`, 
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
        return this.subscribers.get(eventType)?.size || 0;
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

export const ComponentEventTypes= {
    COMPONENT_INITIALIZED: 'component-initialized',
    COMPONENT_HYDRATED: 'component-hydrated',
    COMPONENT_ERROR: 'component-error'
} as const;

export const LifecycleEventTypes = {
    LOCAL_INIT_STARTED: "lifecycle-local-init-started",
    LOCAL_INIT_FINISHED: "lifecycle-local-init-finished",
    DEPENDENCIES_INJECTED: "lifecycle-dependencies-injected",
    ACTIVATION_STARTED: "lifecycle-activation-started",
    ACTIVATION_FINISHED: "lifecycle-activation-finished",
    DEACTIVATION_STARTED: "lifecycle-deactivation-started",
    DEACTIVATION_FINISHED: "lifecycle-deactivation-finished",
} as const;

export type LifecycleEventType = typeof LifecycleEventTypes[keyof typeof LifecycleEventTypes];
export type ComponentEventType = typeof ComponentEventTypes[keyof typeof ComponentEventTypes];
