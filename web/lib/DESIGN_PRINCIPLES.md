
# Component Design Principles

This document outlines the core principles for building components in this application, developed through practical experience to prevent DOM corruption, ensure separation of concerns, and maintain clean architecture.

## Evolution Summary

This architecture evolved from solving a critical DOM corruption bug where broad CSS selectors (`document.querySelectorAll('.text-gray-900, .text-white')`) accidentally matched the body element, replacing the entire page content with "34". This led us to develop strict component isolation principles and a simplified constructor pattern that eliminates initialization complexity while ensuring robust separation of concerns.

## Core Principles

### 1. Strict Separation of Concerns

**Components must have clear, non-overlapping responsibilities:**

- **Parent Components**: Data loading, orchestration, component coordination
- **Child Components**: Specific functionality within their assigned domain
- **Layout/Styling**: Handled by CSS classes and parent containers
- **Communication**: Through EventBus only, never direct method calls

### 2. DOM Scoping and Isolation

**Components must only access DOM within their root element:**

- Use `this.findElement()` and `this.findElements()` for DOM queries
- Never use `document.querySelector()` or global DOM access
- Each component owns and manages only its root element and children
- Root elements must be clearly defined and scoped

**Example:**
```typescript
// ✅ CORRECT - Scoped to component
const element = this.findElement('.my-button');

// ❌ WRONG - Global DOM access
const element = document.querySelector('.my-button');
```

### 3. Layout and Styling Separation

**Components should not control their own layout or external styling:**

- **Parent/CSS Controls**: Container size, positioning, layout, external spacing
- **Component Controls**: Internal behavior, content management, internal styling only
- Use CSS libraries (Tailwind) for styling, not JavaScript
- Components may style their internal elements but not their container

**Example:**
```typescript
// ❌ WRONG - Component controlling its own layout
this.rootElement.style.width = '100%';
this.rootElement.style.height = '500px';

// ✅ CORRECT - Let parent CSS handle layout
// Use CSS classes: w-full h-96 min-h-[500px]
```

### 4. LCMComponent Lifecycle Management

**All components must follow the 3-phase LCMComponent lifecycle:**

1. **performLocalInit()**: EventBus subscriptions, DOM setup, and child component discovery (breadth-first)
2. **setupDependencies()**: Dependency injection and validation 
3. **activate()**: Final initialization and coordination with other components
4. **deactivate()**: Clean up resources, unsubscribe from events

**Critical EventBus Subscription Timing**: EventBus subscriptions must happen **first** in performLocalInit(), before creating child components. This ensures that when children emit events immediately during their construction, parent components are already subscribed and ready to receive them.

```typescript
performLocalInit(): LCMComponent[] {
    // 1. FIRST: Subscribe to events before creating children
    this.subscribe('child-ready', this, this.handleChildReady);
    
    // 2. THEN: Create child components (they can emit events immediately)
    const child = new ChildComponent(rootElement, eventBus);
    
    // 3. FINALLY: Return children for lifecycle management
    return [child];
}
```

This lifecycle ensures proper initialization order and eliminates race conditions through synchronization barriers.

### 5. Event-Driven Communication

**Components communicate only through EventBus:**

- No direct method calls between components
- Use type-safe event definitions
- Events include source identification for debugging
- Error isolation - one component failure doesn't cascade
- Events are synchronous and allow multiple entities to react
- Events should not be sent back to the source component

**Example:**
```typescript
// ✅ CORRECT - EventBus communication
this.emit(EventTypes.WORLD_DATA_LOADED, worldData);

// ❌ WRONG - Direct method calls
otherComponent.updateData(worldData);
```

### 6. Error Isolation and Handling

**Component errors must not affect other components:**

- Each component handles its own errors
- Use `this.handleError()` for consistent error handling
- Emit error events for parent components to handle
- Continue operation even if one component fails

### 7. Resource Management

**Components must clean up after themselves:**

- Unsubscribe from all events in `destroy()`
- Release DOM references
- Clean up third-party library instances (Phaser, etc.)
- Reset component state

### 8. Hydration Pattern Support

**Components must support both initialization modes:**

- **Initialize**: Create new DOM elements and functionality
- **Hydrate**: Bind to existing server-rendered DOM
- Validate DOM structure before hydration
- Create missing elements when validation fails

## Implementation Guidelines

### LCMComponent Lifecycle Pattern

All components implement the 3-phase LCMComponent lifecycle for proper initialization order:

```typescript
export class MyComponent extends BaseComponent implements LCMComponent {
    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('my-component', rootElement, eventBus, debugMode);
        // Note: Lifecycle methods are called by LifecycleController, not constructor
    }
    
    // Phase 1: Subscribe to events FIRST, then create children
    performLocalInit(): LCMComponent[] {
        // 1. CRITICAL: Subscribe to events before creating children
        this.subscribe(EventTypes.CHILD_READY, this, this.handleChildReady);
        this.subscribe(EventTypes.DATA_LOADED, this, this.handleDataLoaded);
        
        // 2. Set up DOM elements
        this.myButton = this.findElement('.my-button') || this.createButton();
        
        // 3. Create child components (they can emit events immediately)
        const child = new ChildComponent(childRoot, this.eventBus);
        
        // 4. Return children for lifecycle management
        return [child];
    }
    
    // Phase 2: Inject dependencies from other components
    setupDependencies(): void {
        // Set up dependencies from other components if needed
        // Example: this.dataService = serviceRegistry.get('data-service');
    }
    
    // Phase 3: Final activation after all components are ready
    activate(): void {
        // Set up DOM event handlers
        this.myButton?.addEventListener('click', () => this.handleClick());
        
        // Any final coordination with other components
        this.emit(EventTypes.COMPONENT_READY, { componentId: this.componentId }, this, this);
    }
    
    // Cleanup
    deactivate(): void {
        this.destroy();
    }
    
    // Component-specific cleanup
    protected destroyComponent(): void {
        this.mySpecificCleanup();
    }
}
```

### Parent Orchestration Pattern with LifecycleController

```typescript
export class ParentPage extends BasePage implements LCMComponent {
    private myComponent: MyComponent | null = null;
    
    // Phase 1: Subscribe to events BEFORE creating children
    performLocalInit(): LCMComponent[] {
        // 1. CRITICAL: Subscribe to child events first
        this.subscribe('child-ready', this, this.handleChildReady);
        
        // 2. Create child components (they can emit events immediately)
        const componentRoot = this.ensureElement('[data-component="my-component"]', 'fallback-id');
        this.myComponent = new MyComponent(componentRoot, this.eventBus, true);
        
        // 3. Return children for lifecycle management
        return this.myComponent ? [this.myComponent] : [];
    }
    
    // Phase 3: Final page activation
    activate(): void {
        super.activate(); // Bind base page events
        // Any additional page-specific coordination
    }
    
    private ensureElement(selector: string, fallbackId: string): HTMLElement {
        let element = document.querySelector(selector) as HTMLElement;
        if (!element) {
            element = document.createElement('div');
            element.id = fallbackId;
            element.className = 'w-full h-full';
            document.body.appendChild(element);
        }
        return element;
    }
}

// Initialize with LifecycleController for proper coordination
document.addEventListener('DOMContentLoaded', async () => {
    const page = new ParentPage('parent-page');
    const lifecycleController = new LifecycleController(page.eventBus);
    await lifecycleController.initializeFromRoot(page);
});
```

### Event Communication

```typescript
// Define event types
export const EventTypes = {
    DATA_LOADED: 'data-loaded',
    ERROR_OCCURRED: 'error-occurred'
} as const;

// Subscribe to events
this.subscribe<DataPayload>(EventTypes.DATA_LOADED, (payload) => {
    this.handleDataLoaded(payload.data);
});

// Emit events
this.emit(EventTypes.DATA_LOADED, { data: myData });
```

### DOM Scoping

```typescript
// ✅ CORRECT - Scoped DOM access
const button = this.findElement<HTMLButtonElement>('.action-button');
const inputs = this.findElements<HTMLInputElement>('input[type="text"]');

// ✅ CORRECT - Event handling within scope
button?.addEventListener('click', () => {
    this.handleButtonClick();
});
```

## Anti-Patterns to Avoid

### 1. Cross-Component DOM Access
```typescript
// ❌ WRONG - Accessing other component's DOM
const otherElement = document.querySelector('#other-component-element');
```

### 2. Global Event Listeners
```typescript
// ❌ WRONG - Global event listener
document.addEventListener('click', this.handleClick);

// ✅ CORRECT - Scoped event listener
this.findElement('.my-button')?.addEventListener('click', this.handleClick);
```

### 3. Layout Control in Components
```typescript
// ❌ WRONG - Component controlling its layout
this.rootElement.style.position = 'absolute';
this.rootElement.style.top = '50px';
```

### 4. Direct Component Communication
```typescript
// ❌ WRONG - Direct method calls
this.otherComponent.updateDisplay(data);

// ✅ CORRECT - Event-based communication
this.emit(EventTypes.DATA_UPDATED, data);
```

### 5. Unsafe DOM Queries
```typescript
// ❌ WRONG - Broad selectors that can match unintended elements
const elements = document.querySelectorAll('.text-gray-900, .text-white');

// ✅ CORRECT - Specific, scoped selectors
const elements = this.findElements('.stat-value');
```

## Benefits of This Approach

1. **Prevents DOM Corruption**: Scoped access prevents accidental modification of other components
2. **Improves Maintainability**: Clear responsibilities make code easier to understand and modify
3. **Enables Reusability**: Components can be used in different contexts without modification
4. **Supports Testing**: Isolated components can be tested independently
5. **Facilitates HTMX Integration**: Hydration pattern works with server-sent HTML fragments
6. **Error Resilience**: Component failures don't cascade to other parts of the application

## Migration Strategy

When converting existing code to this component architecture:

1. **Identify Component Boundaries**: Determine logical component divisions
2. **Extract DOM Logic**: Move DOM manipulation into scoped component methods
3. **Replace Global Queries**: Convert `document.querySelector` to `this.findElement`
4. **Add Event Communication**: Replace direct method calls with EventBus events
5. **Implement Lifecycle**: Add proper initialization and cleanup
6. **Test Isolation**: Verify components don't affect each other

This architecture ensures robust, maintainable components that work well together while maintaining clear separation of concerns.

## Key Design Decisions & Learnings

### 1. Why We Simplified the Constructor Pattern

**Problem**: Initial architecture had complex `initialize()` vs `hydrate()` methods with validation logic, creating bloated components and unclear initialization paths.

**Solution**: Single constructor pattern where parent ensures root element exists:
```typescript
// Before: Complex initialization
if (specificElement) {
    component.initialize({...config});
} else {
    component.hydrate(fallbackElement, eventBus);
}

// After: Simple constructor
parentElem = ensureElement(selector);
component = new Component(parentElem, eventBus);
```

**Benefits**: 
- Eliminates initialization complexity and edge cases
- Clear responsibility: parent handles layout, component handles behavior
- Works for both empty and pre-populated DOM scenarios

### 2. The "Find or Create" Pattern in bindToDOM()

**Key Insight**: Components should handle both cases automatically:
- **Case 1**: Empty root element → create missing DOM elements
- **Case 2**: Pre-populated root → bind to existing elements
- **Runtime**: `contentUpdated()` → re-bind after HTMX updates

```typescript
protected bindToDOM(): void {
    // Handles both empty and pre-populated root elements automatically
    this.myButton = this.findElement('.my-button') || this.createButton();
    this.setupEventHandlers();
}
```

### 3. Layout vs Behavior Separation

**Critical Principle**: Components should never control their own layout/positioning.

**Parent/CSS Controls**: Size, position, layout, external spacing
**Component Controls**: Internal behavior, content management, internal styling only

This prevents components from interfering with each other's layout and makes them truly reusable.

### 4. Why We Use Data Attributes for Component Boundaries

**Problem**: Class-based selectors like `.text-gray-900, .text-white` can accidentally match unintended elements.

**Solution**: Specific data attributes create clear boundaries:
```html
<!-- Template with clear component boundaries -->
<div data-component="world-viewer">
    <div id="phaser-viewer-container"></div>
</div>
<div data-component="world-stats-panel">
    <div data-stat-section="basic">
        <span data-stat="total-tiles">64</span>
    </div>
</div>
```

**Benefits**:
- Prevents accidental cross-component DOM access
- Makes component boundaries visible in templates
- Enables safe component-scoped selectors

### 5. EventBus Over Direct Method Calls

**Decision**: All inter-component communication goes through EventBus, never direct method calls.

**Why**: 
- Prevents tight coupling between components
- Enables error isolation (one component failure doesn't cascade)
- Makes component relationships explicit and debuggable
- Supports one-to-many communication patterns naturally

### 6. HTMX Integration Through Content Updates

**Scenario**: Server sends new HTML fragment via HTMX, component needs to rebind.

**Solution**: `contentUpdated(newHTML)` method automatically updates DOM and rebinds:
```typescript
// HTMX sends new content
component.contentUpdated(newHTMLFromServer);
// Component automatically updates innerHTML and calls bindToDOM()
```

This makes components work seamlessly with server-driven UI updates.

### 7. Template Patterns for Component Safety

**Best Practice**: Use data attributes to create safe, specific selectors:

```html
<!-- ✅ GOOD: Specific component boundaries -->
<div data-component="login-form">
    <input data-field="username" />
    <input data-field="password" />
    <button data-action="submit">Login</button>
</div>

<!-- ❌ BAD: Generic classes that might match other elements -->
<div class="form">
    <input class="input" />
    <input class="input" />
    <button class="button">Login</button>
</div>
```

### 8. Component Testing and Validation

**Approach**: Test component isolation with automated validation:
- Components can initialize independently
- DOM scoping prevents cross-component access
- Error isolation prevents cascading failures
- Event communication works without direct coupling

See `ComponentIsolationTest.ts` for our validation suite.

## Architecture Benefits Realized

1. **DOM Corruption Prevention**: ✅ Eliminated the original bug through strict scoping
2. **Simplified Development**: ✅ Constructor pattern is much easier to use and understand
3. **Runtime Flexibility**: ✅ Components handle both static and dynamic content seamlessly
4. **Error Resilience**: ✅ Component failures don't affect other components
5. **HTMX Compatibility**: ✅ Ready for server-driven UI updates
6. **Maintainability**: ✅ Clear separation of concerns makes code easier to modify
7. **Reusability**: ✅ Components work in different contexts without modification

This architecture evolved through practical problem-solving and has proven robust in preventing the types of issues that led to our original DOM corruption bug.

## Critical Timing and Initialization Lessons

### 1. TypeScript Class Field Initializers vs Constructor Assignment

**Problem Discovered**: When class fields are explicitly initialized in the class body, TypeScript may reset them after constructor execution, overwriting values set in methods called by the constructor.

```typescript
// ❌ PROBLEMATIC: Explicit initializer can reset value
class Component {
    private viewerContainer: HTMLElement | null = null; // This resets after constructor!
    
    constructor() {
        this.bindToDOM(); // Sets viewerContainer
        // TypeScript may reset viewerContainer to null here
    }
}

// ✅ CORRECT: No explicit initializer 
class Component {
    private viewerContainer: HTMLElement | null; // No initializer
    
    constructor() {
        this.bindToDOM(); // Sets viewerContainer and it stays set
    }
}
```

**Key Learning**: Avoid explicit `= null` initializers for fields that get set during construction. Use TypeScript's type system instead: `field: Type | null`.

### 2. Event Subscription vs Component Creation Order

**Problem**: Race condition where components emit events before subscribers are registered.

```typescript
// ❌ WRONG ORDER: Component emits before subscription
const component = new Component(); // Emits 'ready' event immediately
eventBus.subscribe('ready', handler); // Too late!

// ✅ CORRECT ORDER: Subscribe before creating
eventBus.subscribe('ready', handler); // Ready to receive
const component = new Component(); // Emits 'ready' event → handler called
```

**LCMComponent Pattern**: In performLocalInit(), ALWAYS subscribe to events before creating child components:

```typescript
performLocalInit(): LCMComponent[] {
    // 1. FIRST: Subscribe to events (fire and forget mechanism)
    this.subscribe('child-ready', this, this.handleChildReady);
    
    // 2. THEN: Create children who can emit immediately
    const child = new ChildComponent(root, eventBus);
    
    return [child];
}
```

**Key Insight**: EventBus is a "fire and forget" mechanism, not a callback approach. Subscriptions must be in place before events are fired.

### 3. WebGL/Canvas Initialization Timing

**Problem**: Even when Phaser reports "initialized", the WebGL context and scene may still be setting up asynchronously.

```typescript
// ❌ PROBLEMATIC: Immediate usage after "initialized"
this.phaserViewer.initialize(containerId); // Returns true
this.phaserViewer.loadWorldData(data); // May fail - WebGL not fully ready

// ✅ WORKING: Small delay for WebGL context completion
this.phaserViewer.initialize(containerId);
setTimeout(() => {
    this.phaserViewer.loadWorldData(data); // WebGL context ready
}, 10); // Next event loop tick
```

**Key Learning**: Graphics libraries often need a tick for full initialization even when they report "ready".

### 4. Async Operations in Event Handlers

**Problem**: EventBus handlers may need to perform async operations without blocking the event system.

```typescript
// ✅ CORRECT: Async handler with proper error handling
eventBus.subscribe('ready', () => {
    // Non-blocking async operation
    this.loadData()
        .then(() => console.log('Success'))
        .catch(error => this.handleError(error));
});

// ❌ AVOID: Making EventBus itself async
// EventBus should stay synchronous for performance
```

**Pattern**: EventBus stays synchronous, individual handlers use `.then()/.catch()` or wrap async operations.

### 5. Constructor Execution Order in Inheritance

**Learning**: In TypeScript/JavaScript inheritance, parent constructor completes fully before child-specific initialization.

```typescript
class BasePage {
    constructor() {
        this.loadInitialState();    // Must happen first
        this.initializeComponents(); // Then components  
        this.bindEvents();          // Then event binding
    }
}
```

**Critical**: `loadInitialState()` must run before component creation so data is available when components emit ready events.

### 6. Component State Initialization Patterns

**Discovered Pattern**: Three-phase initialization works best:

```typescript
// Phase 1: State setup (synchronous)
this.loadInitialState(); // Set currentWorldId, etc.

// Phase 2: Event subscriptions (before component creation)  
this.eventBus.subscribe('component-ready', this.handleReady);

// Phase 3: Component creation (may emit events immediately)
this.component = new Component(root, eventBus);
```

### 7. Debug Logging for Timing Issues

**Essential Practice**: When debugging timing issues, log:
- Constructor entry/exit
- Event emissions
- Event receptions  
- Async operation start/completion

```typescript
// Timing debug pattern
console.log('About to emit ready event');
this.emit('ready', data);
console.log('Ready event emitted');

// In subscriber
console.log('Received ready event, starting async work');
```

### 8. WebGL Context and Container Sizing

**Timing Issue**: WebGL contexts need containers to have dimensions before initialization works properly.

```typescript
// Pattern for WebGL initialization
protected bindToDOM(): void {
    // 1. Ensure container exists and has dimensions
    this.viewerContainer = this.findElement('#container') || this.createContainer();
    
    // 2. Initialize immediately (container is ready)
    this.initializePhaserViewer();
    
    // 3. But delay actual usage for WebGL context readiness  
    setTimeout(() => this.loadWorldData(), 10);
}
```

## Updated Best Practices

1. **No explicit class field initializers** for constructor-set fields
2. **Subscribe before create** - in performLocalInit(), ALWAYS subscribe to events before creating child components
3. **EventBus is fire-and-forget** - subscriptions must be in place before events are emitted
4. **Separate "initialized" from "ready"** - graphics libraries need settling time
5. **Use setTimeout(fn, 0-10ms)** for WebGL context readiness, not longer delays
6. **performLocalInit order**: Subscribe → DOM setup → Create children → Return children
7. **Async in handlers, sync EventBus** - keep event system simple
8. **Debug timing issues** with comprehensive logging
9. **Test race conditions** by varying initialization order

These patterns prevent timing-related bugs and ensure components work reliably across different execution environments and load conditions.
