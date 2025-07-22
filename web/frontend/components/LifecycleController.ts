import { 
    ComponentLifecycle, 
    ComponentPhase, 
    ComponentLifecycleConfig, 
    ComponentLifecycleError, 
    ComponentDependencyError,
    ComponentDependencyDeclaration,
    ComponentLifecycleEvent,
    ComponentLifecycleEventCallback
} from './ComponentLifecycle';

/**
 * LifecycleController - Orchestrates breadth-first component initialization
 * 
 * This controller implements a breadth-first traversal of the component tree
 * followed by phase-wise initialization with synchronization barriers.
 * 
 * Process:
 * 1. Discovery Phase: Traverse component tree to find all components
 * 2. DOM Binding Phase: All components bind to DOM simultaneously  
 * 3. Dependency Injection Phase: All components receive dependencies
 * 4. Activation Phase: All components complete initialization
 * 
 * Each phase acts as a synchronization barrier - no component can proceed
 * to the next phase until ALL components have completed the current phase.
 */
export class LifecycleController {
    private allComponents: Set<ComponentLifecycle> = new Set();
    private componentPhases: Map<ComponentLifecycle, ComponentPhase> = new Map();
    private componentNames: Map<ComponentLifecycle, string> = new Map();
    private config: Required<ComponentLifecycleConfig>;
    private eventCallbacks: ComponentLifecycleEventCallback[] = [];
    
    constructor(config: ComponentLifecycleConfig = {}) {
        this.config = {
            phaseTimeoutMs: config.phaseTimeoutMs ?? 10000,
            continueOnError: config.continueOnError ?? false,
            validateDependencies: config.validateDependencies ?? true,
            enableDebugLogging: config.enableDebugLogging ?? false
        };
    }
    
    /**
     * Add a callback to receive lifecycle events
     */
    public onLifecycleEvent(callback: ComponentLifecycleEventCallback): void {
        this.eventCallbacks.push(callback);
    }
    
    /**
     * Initialize component tree starting from root component
     * 
     * @param rootComponent The root component to start initialization from
     * @param rootName Optional name for the root component (for debugging)
     * @returns Promise that resolves when all components are fully initialized
     */
    public async initializeFromRoot(
        rootComponent: ComponentLifecycle, 
        rootName: string = 'RootComponent'
    ): Promise<void> {
        try {
            this.log('Starting component tree initialization');
            
            // Phase 0: Discovery - Find all components in the tree
            await this.discoverComponents(rootComponent, rootName);
            this.log(`Discovered ${this.allComponents.size} components`);
            
            // Phase 1: DOM Initialization - All components initialize DOM
            await this.executePhase('initializeDOM', 'dom-bound');
            
            // Phase 2: Dependency Injection - All components receive dependencies
            await this.executeDependencyInjectionPhase();
            
            // Phase 3: Activation - All components complete initialization
            await this.executePhase('activate', 'activated');
            
            this.log('Component tree initialization complete');
            this.emitEvent('component-ready', 'activated', 'All Components');
            
        } catch (error) {
            this.logError('Component tree initialization failed', error);
            throw error;
        }
    }
    
    /**
     * Deactivate all components in reverse order
     */
    public async deactivateAll(): Promise<void> {
        this.log('Deactivating all components');
        
        // Deactivate in reverse order (children first, then parents)
        const componentsArray = Array.from(this.allComponents).reverse();
        
        await this.executePhaseForComponents(
            componentsArray, 
            'deactivate', 
            'deactivated'
        );
        
        this.log('All components deactivated');
    }
    
    /**
     * Phase 0: Discover all components via breadth-first traversal
     */
    private async discoverComponents(
        rootComponent: ComponentLifecycle, 
        rootName: string
    ): Promise<void> {
        const queue: Array<{ component: ComponentLifecycle, name: string }> = [
            { component: rootComponent, name: rootName }
        ];
        
        while (queue.length > 0) {
            const { component, name } = queue.shift()!;
            
            // Skip if already discovered
            if (this.allComponents.has(component)) {
                continue;
            }
            
            // Register component
            this.allComponents.add(component);
            this.componentPhases.set(component, 'created');
            this.componentNames.set(component, name);
            
            this.log(`Discovered component: ${name}`);
            
            try {
                // Discover children by calling initializeDOM
                const children = component.initializeDOM();
                this.componentPhases.set(component, 'dom-bound');
                
                // Add children to discovery queue
                children.forEach((child, index) => {
                    const childName = `${name}.Child${index}`;
                    queue.push({ component: child, name: childName });
                });
                
                this.log(`${name} bound to DOM, discovered ${children.length} children`);
                
            } catch (error) {
                this.handleComponentError(component, 'dom-bound', error as Error);
            }
        }
    }
    
    /**
     * Execute a lifecycle phase for all components with synchronization barrier
     */
    private async executePhase(
        methodName: keyof ComponentLifecycle,
        targetPhase: ComponentPhase
    ): Promise<void> {
        // Skip initializeDOM as it was already called during discovery
        if (methodName === 'initializeDOM') {
            return;
        }
        
        await this.executePhaseForComponents(
            Array.from(this.allComponents),
            methodName,
            targetPhase
        );
    }
    
    /**
     * Execute dependency injection phase with dependency resolution
     */
    private async executeDependencyInjectionPhase(): Promise<void> {
        this.log('=== Starting Dependency Injection Phase ===');
        
        const promises = Array.from(this.allComponents).map(async (component) => {
            const name = this.getComponentName(component);
            
            try {
                // Build dependencies for this component
                const dependencies = this.buildDependenciesFor(component);
                
                // Validate dependencies if enabled
                if (this.config.validateDependencies) {
                    this.validateDependencies(component, dependencies);
                }
                
                this.log(`Injecting dependencies into ${name}:`, Object.keys(dependencies));
                
                // Inject dependencies (may be async)
                const result = component.injectDependencies(dependencies);
                if (result instanceof Promise) {
                    await result;
                }
                
                this.componentPhases.set(component, 'dependencies-injected');
                this.emitEvent('phase-complete', 'dependencies-injected', name);
                this.log(`✓ ${name} dependencies injected`);
                
            } catch (error) {
                this.handleComponentError(component, 'dependencies-injected', error as Error);
            }
        });
        
        // Synchronization barrier: Wait for all dependency injections
        await this.waitForPromises(promises, 'dependencies-injected');
        
        this.log('=== Dependency Injection Phase Complete ===');
    }
    
    /**
     * Execute a lifecycle phase for specific components
     */
    private async executePhaseForComponents(
        components: ComponentLifecycle[],
        methodName: keyof ComponentLifecycle,
        targetPhase: ComponentPhase
    ): Promise<void> {
        this.log(`=== Starting ${targetPhase} Phase ===`);
        
        const promises = components.map(async (component) => {
            const name = this.getComponentName(component);
            
            try {
                this.emitEvent('phase-start', targetPhase, name);
                this.log(`Starting ${targetPhase} for ${name}`);
                
                // Execute the lifecycle method (may be async)
                const method = component[methodName] as Function;
                const result = method.call(component);
                if (result instanceof Promise) {
                    await result;
                }
                
                this.componentPhases.set(component, targetPhase);
                this.emitEvent('phase-complete', targetPhase, name);
                this.log(`✓ ${name}.${methodName}() completed`);
                
            } catch (error) {
                this.handleComponentError(component, targetPhase, error as Error);
            }
        });
        
        // Synchronization barrier: Wait for all components to complete phase
        await this.waitForPromises(promises, targetPhase);
        
        this.log(`=== ${targetPhase} Phase Complete ===`);
    }
    
    /**
     * Build dependency map for a specific component
     */
    private buildDependenciesFor(component: ComponentLifecycle): Record<string, any> {
        const dependencies: Record<string, any> = {};
        const componentName = this.getComponentName(component);
        
        this.log(`Built dependencies for ${componentName}:`, Object.keys(dependencies));
        return dependencies;
    }
    
    /**
     * Validate that required dependencies are provided
     */
    private validateDependencies(
        component: ComponentLifecycle, 
        providedDeps: Record<string, any>
    ): void {
        // Check if component implements dependency declaration interface
        const declarationComponent = component as ComponentLifecycle & ComponentDependencyDeclaration;
        
        if (typeof declarationComponent.getRequiredDependencies === 'function') {
            const required = declarationComponent.getRequiredDependencies();
            const provided = Object.keys(providedDeps);
            const missing = required.filter(dep => !provided.includes(dep));
            
            if (missing.length > 0) {
                throw new ComponentDependencyError(
                    this.getComponentName(component),
                    missing,
                    provided
                );
            }
        }
    }
    
    /**
     * Wait for all promises with timeout and error handling
     */
    private async waitForPromises(promises: Promise<any>[], phaseName: string): Promise<void> {
        const timeoutPromise = new Promise((_, reject) => {
            setTimeout(() => {
                reject(new Error(`Phase ${phaseName} timed out after ${this.config.phaseTimeoutMs}ms`));
            }, this.config.phaseTimeoutMs);
        });
        
        try {
            if (this.config.continueOnError) {
                // Use allSettled to continue even if some components fail
                await Promise.race([
                    Promise.allSettled(promises),
                    timeoutPromise
                ]);
            } else {
                // Use all for fail-fast behavior
                await Promise.race([
                    Promise.all(promises),
                    timeoutPromise
                ]);
            }
        } catch (error) {
            this.logError(`Phase ${phaseName} failed`, error);
            throw error;
        }
    }
    
    /**
     * Handle component error during lifecycle
     */
    private handleComponentError(
        component: ComponentLifecycle,
        phase: ComponentPhase,
        error: Error
    ): void {
        const name = this.getComponentName(component);
        const lifecycleError = new ComponentLifecycleError(name, phase, error.message, error);
        
        this.emitEvent('phase-error', phase, name, lifecycleError);
        this.logError(`✗ ${name}.${phase} failed`, lifecycleError);
        
        if (!this.config.continueOnError) {
            throw lifecycleError;
        }
    }
    
    /**
     * Get component name for logging/debugging
     */
    private getComponentName(component: ComponentLifecycle): string {
        return this.componentNames.get(component) || component.constructor.name || 'UnknownComponent';
    }
    
    /**
     * Emit lifecycle event to registered callbacks
     */
    private emitEvent(
        type: ComponentLifecycleEvent['type'],
        phase: ComponentPhase,
        componentName: string,
        error?: Error
    ): void {
        const event: ComponentLifecycleEvent = {
            type,
            phase,
            componentName,
            timestamp: Date.now(),
            error
        };
        
        this.eventCallbacks.forEach(callback => {
            try {
                callback(event);
            } catch (callbackError) {
                console.error('Lifecycle event callback failed:', callbackError);
            }
        });
    }
    
    /**
     * Log message if debug logging is enabled
     */
    private log(message: string, ...args: any[]): void {
        if (this.config.enableDebugLogging) {
            console.log(`[LifecycleController] ${message}`, ...args);
        }
    }
    
    /**
     * Log error message
     */
    private logError(message: string, error?: any): void {
        console.error(`[LifecycleController] ${message}`, error);
    }
    
    /**
     * Get current status of all components
     */
    public getComponentStatus(): Array<{ name: string; phase: ComponentPhase }> {
        return Array.from(this.allComponents).map(component => ({
            name: this.getComponentName(component),
            phase: this.componentPhases.get(component) || 'created'
        }));
    }
    
    /**
     * Get total number of discovered components
     */
    public getComponentCount(): number {
        return this.allComponents.size;
    }
}
