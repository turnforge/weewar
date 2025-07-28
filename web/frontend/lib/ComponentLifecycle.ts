/**
 * ComponentLifecycle Interface - Defines multi-phase component initialization
 * 
 * This interface implements a breadth-first lifecycle pattern that eliminates
 * initialization order dependencies and race conditions through synchronization barriers.
 * 
 * Lifecycle Phases:
 * 1. Construction - Components are created but not initialized
 * 2. bindToDOM() - Basic DOM setup, discover child components  
 * 3. injectDependencies() - Receive references to other components
 * 4. activate() - Final setup when all dependencies are ready
 * 
 * Key Benefits:
 * - Order Independence: Components can be created in any sequence
 * - Async Safety: Each phase waits for all components before proceeding
 * - Clear Dependencies: Explicit injection points prevent race conditions
 * - Error Isolation: Component failures don't cascade to others
 */

export interface ComponentLifecycle {
    /**
     * Phase 1: Initialize DOM and discover child components
     * 
     * This phase should:
     * - Set up basic DOM elements and event listeners
     * - Create child components (but don't initialize them)
     * - Return array of child components for lifecycle controller discovery
     * 
     * This phase must be synchronous and should not:
     * - Access other components or external dependencies
     * - Perform async operations
     * - Emit events or notifications
     * 
     * @returns Array of child components to be managed by lifecycle controller
     */
    initializeDOM(): ComponentLifecycle[];
    
    /**
     * Phase 2: Inject dependencies from parent/siblings
     * 
     * This phase should:
     * - Receive and store references to required dependencies
     * - Validate that required dependencies are provided
     * - Set up internal state based on dependencies
     * 
     * This phase can be async and may:
     * - Load external data or resources
     * - Perform validation or setup operations
     * - Initialize internal components that depend on injected references
     * 
     * @param deps Record of dependency name to dependency instance
     * @returns Promise<void> or void - can be async
     */
    injectDependencies(deps: Record<string, any>): Promise<void> | void;
    
    /**
     * Phase 3: Activate component when all dependencies are ready
     * 
     * This phase should:
     * - Complete final initialization
     * - Enable component functionality
     * - Start listening for external events
     * - Begin normal operation
     * 
     * This phase can be async and may:
     * - Connect to external services
     * - Load initial data
     * - Emit ready notifications
     * 
     * @returns Promise<void> or void - can be async
     */
    activate(): Promise<void> | void;
    
    /**
     * Cleanup phase: Deactivate component and clean up resources
     * 
     * This should:
     * - Stop all ongoing operations
     * - Remove event listeners
     * - Clean up external connections
     * - Dispose of child components
     * 
     * @returns Promise<void> or void - can be async
     */
    deactivate(): Promise<void> | void;
}

/**
 * Type representing the current phase of a component's lifecycle
 */
export type ComponentPhase = 'created' | 'dom-bound' | 'dependencies-injected' | 'activated' | 'deactivated';

/**
 * Interface for components that need to declare their dependencies
 * This is optional but helps with debugging and validation
 */
export interface ComponentDependencyDeclaration {
    /**
     * Declare required dependencies that must be provided in injectDependencies
     * Component will not activate if these are missing
     */
    getRequiredDependencies(): string[];
    
    /**
     * Declare optional dependencies that may be provided
     * Component should gracefully handle if these are missing
     */
    getOptionalDependencies(): string[];
}

/**
 * Configuration for component lifecycle behavior
 */
export interface ComponentLifecycleConfig {
    /**
     * Maximum time to wait for a lifecycle phase to complete (ms)
     * Default: 10000 (10 seconds)
     */
    phaseTimeoutMs?: number;
    
    /**
     * Whether to continue if individual components fail during a phase
     * Default: false (fail fast)
     */
    continueOnError?: boolean;
    
    /**
     * Whether to validate dependencies against declared requirements
     * Default: true
     */
    validateDependencies?: boolean;
    
    /**
     * Enable debug logging for lifecycle phases
     * Default: false
     */
    enableDebugLogging?: boolean;
}

/**
 * Error thrown when a component lifecycle operation fails
 */
export class ComponentLifecycleError extends Error {
    constructor(
        public readonly componentName: string,
        public readonly phase: ComponentPhase,
        message: string,
        public readonly cause?: Error
    ) {
        super(`${componentName}.${phase}: ${message}`);
        this.name = 'ComponentLifecycleError';
    }
}

/**
 * Error thrown when component dependencies are not satisfied
 */
export class ComponentDependencyError extends ComponentLifecycleError {
    constructor(
        componentName: string,
        public readonly missingDependencies: string[],
        public readonly providedDependencies: string[]
    ) {
        super(
            componentName,
            'dependencies-injected',
            `Missing required dependencies: ${missingDependencies.join(', ')}. Provided: ${providedDependencies.join(', ')}`
        );
        this.name = 'ComponentDependencyError';
    }
}

/**
 * Event emitted during component lifecycle transitions
 */
export interface ComponentLifecycleEvent {
    type: 'phase-start' | 'phase-complete' | 'phase-error' | 'component-ready';
    phase: ComponentPhase;
    componentName: string;
    timestamp: number;
    error?: Error;
    metadata?: Record<string, any>;
}

/**
 * Callback for lifecycle events
 */
export type ComponentLifecycleEventCallback = (event: ComponentLifecycleEvent) => void;