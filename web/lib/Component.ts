import { EventBus, EventSubscriber, ComponentEventTypes } from './EventBus';
import { LCMComponent } from './LCMComponent';

/**
 * Base interface for all UI components
 * Enforces separation of concerns and standard lifecycle
 */
export interface Component {
    /**
     * Unique identifier for this component instance
     */
    readonly componentId: string;
    
    /**
     * Root DOM element that this component owns and manages
     */
    readonly rootElement: HTMLElement;
    
    /**
     * Handle dynamic content updates (e.g., from HTMX or server responses)
     * @param newHTML - New HTML content to replace current content
     */
    contentUpdated(newHTML: string): void;
    
    /**
     * Clean up the component and release resources
     * Should unsubscribe from events, clean up DOM, and release memory
     */
    destroy(): void;
}

/**
 * Abstract base class implementing common component functionality
 * Provides standard lifecycle management and event bus integration
 * 
 * All components auto-initialize in constructor AND implement LCMComponent
 * for coordination with other components when needed.
 */
export abstract class BaseComponent implements Component, LCMComponent {
    protected eventUnsubscribers: (() => void)[] = [];
    protected _eventBus: EventBus;
    
    constructor(public readonly componentId: string, public readonly rootElement: HTMLElement, eventBus: EventBus | null = null, public readonly debugMode: boolean = false) {
        // Mark as component in DOM for debugging
        this._eventBus = eventBus || new EventBus();
        this.rootElement.setAttribute('data-component', this.componentId);
    }

    public get eventBus(): EventBus {
      return this._eventBus
    }
    
    public contentUpdated(newHTML: string): void {
        this.log('Content updated, re-binding to DOM');
        
        // Update the DOM
        this.rootElement.innerHTML = newHTML;
        
        // Note: In pure LCMComponent approach, re-binding should be handled
        // by the component's LCMComponent lifecycle methods if needed
    }
    
    public destroy(): void {
        this.log('Destroying component...');

        // Unsubscribe from all events
        this.eventUnsubscribers.forEach(unsubscribe => unsubscribe());
        this.eventUnsubscribers = [];
        
        // Call component-specific cleanup
        this.destroyComponent();
        
        // Remove component marker from DOM
        this.rootElement?.removeAttribute('data-component');
        
        this.log('Component destroyed successfully');
    }
    
    /**
     * Subscribe to an event with automatic cleanup on destroy
     */
    protected subscribe<T = any>(eventType: string, target: any, handler: EventHandler<T>): void {
        const unsubscribe = this.eventBus.subscribe(eventType, target, handler);
        this.eventUnsubscribers.push(unsubscribe);
    }
    
    /**
     * Emit an event from this component
     */
    protected emit<T = any>(eventType: string, data: T, target: any, emitter: any = null): void {
        this.eventBus.emit(eventType, data, target, emitter || this)
    }
    
    /**
     * Find elements within this component's root element only
     * Enforces separation of concerns - no cross-component DOM access
     */
    protected findElement<T extends HTMLElement = HTMLElement>(selector: string): T | null {
        return this.rootElement.querySelector<T>(selector);
    }
    
    /**
     * Find multiple elements within this component's root element only
     */
    protected findElements<T extends HTMLElement = HTMLElement>(selector: string): T[] {
        return Array.from(this.rootElement.querySelectorAll<T>(selector));
    }
    
    /**
     * Log messages with component identification
     */
    protected log(message: string, data: any = null): void {
        if (this.debugMode) { console.log(`[${this.componentId}] ${message}`, data);
        }
    }
    
    // Abstract methods that components must implement
    
    /**
     * Component-specific cleanup logic
     * Called during destroy before base cleanup
     */
    protected abstract destroyComponent(): void;
    
    // LCMComponent implementation with default empty methods
    // Components can override these when they need coordination with other components
    
    /**
     * Default lifecycle method: discover and return child components
     * Override this if your component creates child components that need lifecycle management
     */
    public performLocalInit(): Promise<LCMComponent[]>  | LCMComponent[] {
        // Default: no child components
        return [];
    }
    
    /**
     * Default lifecycle method: inject dependencies  
     * Override this if your component needs dependencies from other components
     */
    public setupDependencies(): void | Promise<void> {
        // Default: no dependencies needed
    }
    
    /**
     * Default lifecycle method: activate component for coordination
     * Override this if your component needs to coordinate with other components after initialization
     */
    public activate(): void | Promise<void> {
        // Default: no coordination needed - component is already auto-initialized
    }
    
    /**
     * Default lifecycle method: deactivate component
     * Override this if your component needs cleanup during lifecycle management
     */
    public deactivate(): void | Promise<void> {
        // Default: use standard destroy method
        this.destroy();
    }
}
