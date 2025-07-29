
import { 
    LCMComponent, 
    LCMComponentConfig, 
    LCMComponentEvent,
} from './LCMComponent';
import { EventBus, LifecycleEventTypes, LifecycleEventType } from './EventBus';
import { ComponentLifecycleEvent } from './events';

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
    private allComponents: LCMComponent[];
    private componentsByLevel: LCMComponent[][];
    private config: Required<LCMComponentConfig>;
    
    constructor(protected eventBus: EventBus, config: LCMComponentConfig = {}) {
        this.config = {
            phaseTimeoutMs: config.phaseTimeoutMs ?? 10000,
            continueOnError: config.continueOnError ?? false,
            validateDependencies: false,
            enableDebugLogging: config.enableDebugLogging ?? false
        };
        this.allComponents = [];
        this.componentsByLevel = [];
    }
    
    /**
     * Initialize component tree starting from root component
     * 
     * @param rootComponent The root component to start initialization from
     * @param rootName Optional name for the root component (for debugging)
     * @returns Promise that resolves when all components are fully initialized
     */
    public async initializeFromRoot(rootComponent: LCMComponent): Promise<void> {
        this.log('Starting component tree initialization');
        
        // Phase 0: Discovery - Find all components in the tree
        await this.performLocalInit(rootComponent);
        
        // Phase 1 - Performance post-load setup where each parent will have the change
        // to set dependencies on its children (either to its siblings or to itself).
        await this.performLocalInit(rootComponent);

        // Phase 2: Dependency Injection - All components receive dependencies
        await this.injectDependencies();

        await this.activate();
        
        this.log('Component tree initialization complete');
        // this.emitEvent('component-ready', 'activated', 'All Components');
    }
    
    /**
     * Phase 0: Discover all components via breadth-first traversal
     */
    private async performLocalInit(rootComponent: LCMComponent): Promise<void> {
        const visited = new Set<LCMComponent>();
        let queue = [rootComponent] as LCMComponent[]
        
        for (let level = 0; queue.length > 0; level ++) {
            const newqueue = [] as LCMComponent[];
            const promises = new Array<Promise<LCMComponent[]>>()
            for (let i = 0;i < queue.length;i++) {
                const component = queue[i]

                // Enforce tree structure: each component should have exactly one parent
                if (visited.has(component)) {
                    throw new Error(
                        `Component hierarchy integrity violation: Component '${component.constructor.name}' ` +
                        `discovered as child of multiple parents at level ${level}. ` +
                        `Component hierarchies must form a tree structure where each component has exactly one parent.`
                    );
                }
                
                if (newqueue.length == 0) {
                  this.componentsByLevel.push(newqueue)
                }

                this.emitEvent(LifecycleEventTypes.LOCAL_INIT_STARTED, component, level)
                const result = component.performLocalInit()
                promises.push(Promise.resolve(result).then((children: LCMComponent[]) => {
                  for (const ch of children) {
                    this.allComponents.push(ch)
                    newqueue.push(ch)
                  }
                  return children
                }))

                // now wait for all the promises to finish
                await Promise.all(promises)

                // TODO - Handle errors
                this.emitEvent(LifecycleEventTypes.LOCAL_INIT_FINISHED, component, level)
            }
            queue = newqueue
        }
    }

    /**
     * Phase 1: Dependency injection.
     * Here all our components are already discovered so we can call directly
     */
    protected injectDependencies(): void {
      // This is guaranteed to be in order of levels
      for (const comp of this.allComponents) {
          comp.setupDependencies()
      }
    }

    /**
     * Phase 2: Activate all components
     *
     * Here all our components are already discovered and their dependencies setup
     * so we can call directly
     */
    protected async activate(): Promise<void> {
        // This is guaranteed to be in order of levels
        for (let level = 0;level < this.componentsByLevel.length; level++) {
            const promises = new Array<Promise<void>>()
            const levelComps = this.componentsByLevel[level]
            console.log("Clearing components in level: ", level)
            for (const comp of levelComps) {
                this.emitEvent(LifecycleEventTypes.ACTIVATION_STARTED , comp, 0)
                const result = comp.activate()
                promises.push(Promise.resolve(result).then(() => {
                    this.emitEvent(LifecycleEventTypes.ACTIVATION_FINISHED, comp, 0)
                }))
            }
            await Promise.all(promises)
        }
    }
    
    /**
     * Deactivate all components in reverse order.  This is usually called when a page quits or a component is
     * deactivated
     */
    public async deactivateAll(): Promise<void> {
        this.log('Deactivating all components');
        // destruction should leaf to the root so we are removing things from bottom to top
        for (let level = this.componentsByLevel.length - 1;level >= 0; level--) {
            const promises = new Array<Promise<void>>()
            const levelComps = this.componentsByLevel[level]
            console.log("Clearing components in level: ", level)
            for (const comp of levelComps) {
                this.emitEvent(LifecycleEventTypes.DEACTIVATION_STARTED , comp, level)
                const result = comp.deactivate()
                promises.push(Promise.resolve(result).then(() => {
                    this.emitEvent(LifecycleEventTypes.DEACTIVATION_FINISHED, comp, level)
                }))
            }
            await Promise.all(promises)
        }
    }
    
    /**
     * Emit lifecycle event to registered callbacks
     */
    private emitEvent(
        event: LifecycleEventType,
        component: LCMComponent,
        level: number,
        error: Error | null = null,
    ): void {
        this.eventBus.emit<ComponentLifecycleEvent>(
            event,
            { error: error, success: error == null},
            this, 
            component
        );
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
    /*
    public getComponentStatus(): Array<{ name: string; phase: ComponentPhase }> {
        return Array.from(this.allComponents).map(component => ({
            name: this.getComponentName(component),
            phase: this.componentPhases.get(component) || 'created'
        }));
    }
    */
}
