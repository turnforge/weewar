/**
 * Type-safe event system for component communication
 * Provides error isolation and prevents circular event propagation
 */

export interface EventPayload<T = any> {
    // The target/subject entity that this event relates to.
    subject: any

    timestamp: number;

    data: T;

    // The entity that emitted the event.
    emitter: any;
}

export type EventHandler<T = any> = (payload: EventPayload<T>) => void;

interface EventSubscription {
    name: string; // A way to dedup subscription to avoid getting the same event twice
    handler: EventHandler;
    subject: any;
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
     * @param subject - Filter by subject entity for subscribing events from (after checking for eventType)
     */
    public subscribe<T = any>(
        name: string,
        eventTypeOrTypes: string | string[], 
        subject: any,
        handler: EventHandler<T>, 
    ): () => void {
        if (typeof(eventTypeOrTypes) === "string") {
          eventTypeOrTypes = [eventTypeOrTypes]
        }
        for (const eventType of eventTypeOrTypes as string[]) {
          if (!this.subscribers.has(eventType)) {
              this.subscribers.set(eventType, []);
          }
          
          const subscription: EventSubscription = { name, handler, subject };
          this.subscribers.get(eventType)!.push(subscription);
          
          if (this.debugMode) {
              console.log(`[EventBus] subscribed to '${eventType}'`);
          }
        }
        
        // Return unsubscribe function
        return () => {
          this.unsubscribe(eventTypeOrTypes as string[], subject , handler);
        }
    }
    
    /**
     * Unsubscribe from an event type
     */
    public unsubscribe<T = any>(
        eventTypeOrTypes: (string|string[]), 
        subject: any,
        handler: EventHandler<T>
    ): void {
        if (typeof(eventTypeOrTypes) === "string") {
          eventTypeOrTypes = [eventTypeOrTypes]
        }
        for (const eventType of eventTypeOrTypes) {
            const subscriptions = this.subscribers.get(eventType);
            if (!subscriptions) return;
            
            const index = subscriptions.findIndex(
                sub => sub.subject === subject && sub.handler === handler
            );
            
            if (index !== -1) {
                subscriptions.splice(index, 1);
                if (this.debugMode) {
                    console.log(`[EventBus] '${subject}' unsubscribed from '${eventType}'`);
                }
            }
            
            // Clean up empty event types
            if (subscriptions.length === 0) {
                this.subscribers.delete(eventType);
            }
        }
    }
    
    /**
     * Emit an event to all subscribers (except the source)
     * @param eventType - The event type to emit
     * @param data - The event data payload
     * @param subject - The subject entity that this event relates to.
     * @param emitter - The entity that emitted the event.
     */
    public emit<T = any>(eventType: string, data: T, subject: any, emitter: any): void {
        const subscriptions = this.subscribers.get(eventType);
        if (!subscriptions || subscriptions.length === 0) {
            if (this.debugMode) {
                console.log(`[EventBus] No subscribers for event '${eventType}'`);
            }
            return;
        }
        
        const payload: EventPayload<T> = {
            emitter: emitter,
            subject: subject,
            timestamp: Date.now(),
            data
        };
        
        if (this.debugMode) {
            console.log(`[EventBus] Emitting '${eventType}' from '${emitter}' to ${subscriptions.length} subscribers`);
        }
        
        let successCount = 0;
        let errorCount = 0;
        
        // Call each handler with error isolation
        subscriptions.forEach(subscription => {
            // Source exclusion - don't send event back to the source or the emitter
            if (subscription.subject != null && (subscription.subject === emitter || subscription.subject === subject)) {
                if (this.debugMode) {
                    console.log(`[EventBus] Skipping subject '${subject}'`);
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
                    `in subject '${subscription.subject}':`, 
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
