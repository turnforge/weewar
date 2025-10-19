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
    // destroy(): void;
}

/**
 * Abstract base class implementing common component functionality
 * Provides standard lifecycle management and event bus integration
 * 
 * All components auto-initialize in constructor AND implement LCMComponent
 * for coordination with other components when needed.
 */
export abstract class BaseComponent implements Component, LCMComponent, EventSubscriber {
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
    
    /**
     * Subscribe to an event using the new EventSubscriber pattern
     */
    protected addSubscription(eventType: string, target: any = null): void {
        this.eventBus.addSubscription(eventType, target, this);
    }
    
    /**
     * Unsubscribe from an event using the new EventSubscriber pattern  
     */
    protected removeSubscription(eventType: string, target: any = null): void {
        this.eventBus.removeSubscription(eventType, target, this);
    }
    
    /**
     * Emit an event from this component
     */
    protected emit<T = any>(eventType: string, data: T, target: any, emitter: any = null): void {
        this.eventBus.emit(eventType, data, target, emitter || this);
    }
    
    /**
     * Default implementation of EventSubscriber interface
     * Components can override this to handle events
     */
    public handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
        // Default: no event handling
        // Subclasses should override this method to handle specific events
        if (this.debugMode) {
            console.log(`[${this.componentId}] Received unhandled event: ${eventType}`);
        }
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
        if (this.debugMode) {
            console.log(`[${this.componentId}] ${message}`, data);
        }
    }
    
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
        // Remove component marker from DOM
        this.rootElement?.removeAttribute('data-component');
        
        this.log('Component deactivated successfully');
    }

    set innerHTML(innerHTML: string) {
      const [newHTML, allow] = this.shouldUpdateHtml(innerHTML)
      if (allow) {
        this.rootElement.innerHTML = newHTML
        this.htmlUpdated(newHTML)
      }
    }

    // Called BEFORE the HTML is about to be updated.  This is an opportunity
    // for subclasses to fix/tweak html being updated and also ignore an update
    // just in case.  Default behavior is to return the html as is
    shouldUpdateHtml(html: string): [string, boolean] {
      return [html, true]
    }

    // Called after the HTML for the component has been updated.
    // This is an opportunity to rebind any event listeners etc
    htmlUpdated(html: string) {
      // Do nothing
    }
    
    // LCMComponent implementation with default empty methods
    // Components can override these when they need coordination with other components
}
