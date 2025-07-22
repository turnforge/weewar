# Design Principles

## Component Communication Architecture

### When to Use EventBus vs Direct Communication

The choice between EventBus and direct method calls depends on the component relationship:

#### ✅ Use Direct Method Calls For:

**Parent → Child Communication**
- Parent components can call methods directly on their child components
- Example: `MapEditorPage` calling `this.phaserEditorComponent.setShowGrid(true)`
- Rationale: Parent has direct reference to child, no need for event indirection

**Immediate Parent → Child with Event Handlers**
- Parent can attach event handlers directly to immediate children
- Example: Button click handlers in parent calling child methods
- Rationale: Simple, direct relationship with clear ownership

#### ✅ Use EventBus For:

**Sibling ↔ Sibling Communication**
- Components at the same level that need to communicate
- Example: `EditorToolsPanel` (sibling) ↔ `PhaserEditorComponent` (sibling)
- Rationale: Neither component should know about the other directly

**Grandparent → Grandchild Communication**
- When you need to communicate across multiple component layers
- Example: Deep nested components that need to react to top-level state changes
- Rationale: Avoids passing event handlers through multiple intermediate components

**Cross-Module Communication**
- When components in different modules/sections need to communicate
- Example: Reference image panel communicating with Phaser editor
- Rationale: Maintains loose coupling between different functional areas

**Lifecycle-Managed Component Communication**
- When components are created dynamically by layout systems (dockview, etc.)
- When component initialization order is not guaranteed
- Example: MapEditorPage → PhaserEditorComponent grid visibility
- Rationale: Component references may be null due to lifecycle timing

**Broadcast Communication**
- When one component needs to notify multiple other components
- Example: Map data changes that need to update stats, save indicators, etc.
- Rationale: One-to-many communication pattern

### Component Architecture Patterns

#### State→Subscribe→Create→Bind Pattern
All page-level components should follow this initialization order:
1. **State**: Set up initial state and data
2. **Subscribe**: Subscribe to EventBus events before components are created
3. **Create**: Initialize child components and UI elements
4. **Bind**: Bind template-specific events after components are created and templates are cloned

**Updated for Lifecycle Architecture:**
- Global event binding happens in `bindSpecificEvents()` during activation phase
- Template-specific event binding happens in component creation callbacks
- EventBus subscriptions happen early to catch all component communications

#### Separation of Concerns
- **Layout Components**: Handle only UI layout and direct user interactions
- **Business Logic Components**: Handle data processing and application logic
- **Communication**: Use appropriate pattern (direct calls vs EventBus) based on relationship

#### Error Isolation
- Each component should handle its own errors
- EventBus provides automatic error isolation between components
- Failed event handlers don't affect other subscribers

### Examples

#### ✅ Good: Parent→Child Direct Call (Legacy Pattern)
```typescript
// OLD PATTERN: Direct parent-child calls (pre-lifecycle architecture)
public setShowGrid(showGrid: boolean): void {
    if (this.phaserEditorComponent && this.phaserEditorComponent.getIsInitialized()) {
        this.phaserEditorComponent.setShowGrid(showGrid);
    }
}
```

#### ✅ Better: Lifecycle-Compatible EventBus Pattern
```typescript
// NEW PATTERN: EventBus for lifecycle-compatible communication
public setShowGrid(showGrid: boolean): void {
    // Update page state for persistence
    this.pageState?.setShowGrid(showGrid);
    
    // Emit event for loose coupling
    this.eventBus.emit<GridSetVisibilityPayload>(
        EditorEventTypes.GRID_SET_VISIBILITY,
        { show: showGrid },
        'map-editor-page'
    );
}
```

#### ✅ Good: Sibling EventBus Communication
```typescript
// EditorToolsPanel emitting tool selection for PhaserEditorComponent
this.emit<TerrainSelectedPayload>(EditorEventTypes.TERRAIN_SELECTED, {
    terrainType: terrainValue,
    terrainName: terrainName
});

// PhaserEditorComponent subscribing to tool changes
this.subscribe<TerrainSelectedPayload>(EditorEventTypes.TERRAIN_SELECTED, (payload) => {
    this.handleTerrainSelected(payload.data);
});
```

#### ❌ Anti-Pattern: Direct References in Lifecycle Architecture
```typescript
// DON'T DO THIS - component references may be null due to initialization order
if (this.phaserEditorComponent) {
    this.phaserEditorComponent.setShowGrid(true); // May fail with lifecycle architecture
}
```

#### ❌ Anti-Pattern: Sibling Direct References
```typescript
// DON'T DO THIS - siblings shouldn't know about each other directly
editorToolsPanel.phaserComponent.setTerrain(terrainType);
```

## Component Design Guidelines

### Single Responsibility
Each component should have one clear purpose and responsibility.

### DOM Scoping
Components should only manipulate DOM elements within their designated scope.

### Event-Driven Architecture
Use events for loose coupling, but choose the appropriate communication method based on component relationships.

### Type Safety
All EventBus events should have strongly-typed payload interfaces.

### Error Handling
Components should gracefully handle missing dependencies and initialization failures.

## Lifecycle Architecture Patterns

### DOM Event Binding Timing Pattern

In lifecycle-based architectures, UI event binding must happen at the correct time in the component creation cycle.

#### ✅ Template-Scoped Event Binding
```typescript
// GOOD: Bind events after template is cloned and added to DOM
private createPhaserComponent() {
    const template = document.getElementById('canvas-panel-template');
    const container = template.cloneNode(true) as HTMLElement;
    
    return {
        element: container,
        init: () => {
            // Initialize component first
            this.phaserEditorComponent = new PhaserEditorComponent(container, ...);
            
            // Bind events to cloned template (correct scope)
            this.bindPhaserPanelEvents(container);
        }
    };
}

private bindPhaserPanelEvents(container: HTMLElement): void {
    // Search within the cloned container, not global document
    const checkbox = container.querySelector('#show-grid') as HTMLInputElement;
    if (checkbox) {
        checkbox.addEventListener('change', (e) => {
            this.setShowGrid((e.target as HTMLInputElement).checked);
        });
    }
}
```

#### ❌ Anti-Pattern: Global Event Binding
```typescript
// DON'T DO THIS - binding events before templates are created
protected bindSpecificEvents(): void {
    // This runs before dockview creates panels - elements don't exist yet!
    const checkbox = document.getElementById('show-grid'); // null!
    if (checkbox) {
        checkbox.addEventListener('change', ...);
    }
}
```

**Key Principles:**
1. **Timing**: Bind events AFTER templates are cloned and added to DOM
2. **Scope**: Use `container.querySelector()` instead of `document.getElementById()`
3. **Location**: Bind template-specific events in component initialization callbacks
4. **Separation**: Keep global events in `bindSpecificEvents()`, template events in component-specific methods

### EventBus + Page State Pattern

For state that needs both persistence and loose coupling:

```typescript
// GOOD: Update state AND emit events for best of both worlds
public setShowGrid(showGrid: boolean): void {
    // Update page state for persistence and consistency
    if (this.pageState) {
        this.pageState.setShowGrid(showGrid);
    }
    
    // Emit event for loose coupling with lifecycle-managed components
    this.eventBus.emit<GridSetVisibilityPayload>(
        EditorEventTypes.GRID_SET_VISIBILITY,
        { show: showGrid },
        'map-editor-page'
    );
}
```

## Advanced Patterns Discovered

### Container-Based Component Integration

When integrating complex components (like Phaser) into layout systems (like dockview):

#### ✅ Direct Element Passing Pattern
```typescript
// GOOD: Pass container element directly to avoid DOM scope issues
const container = this.findElement('#my-container');
this.complexComponent = new ComplexComponent(container);
```

#### ❌ Anti-Pattern: String ID Lookup
```typescript
// AVOID: Global DOM lookup can fail in isolated component scopes
this.complexComponent = new ComplexComponent('my-container-id');
```

**Rationale**: Layout systems may create isolated DOM scopes where global `document.getElementById()` fails.

### Timing and Initialization Patterns

#### Component Reference Timing
```typescript
// GOOD: Async event emission allows parent assignments to complete
setTimeout(() => {
    this.emit(COMPONENT_READY, {});
}, 0);
```

#### Pending State Pattern
```typescript
// GOOD: Store user actions when components aren't ready yet
public setFeature(enabled: boolean): void {
    if (this.component && this.component.isReady()) {
        this.component.setFeature(enabled);
    } else {
        this.pendingFeatureState = enabled; // Apply later
    }
}
```

#### Visibility-Based Initialization
```typescript
// GOOD: Wait for proper container dimensions before initializing
private waitForContainerVisible(element: HTMLElement): void {
    const checkVisibility = () => {
        const rect = element.getBoundingClientRect();
        if (rect.width > 0 && rect.height > 0) {
            this.initializeComponent(element);
        } else {
            setTimeout(checkVisibility, 50);
        }
    };
    setTimeout(checkVisibility, 50);
}
```

### Constructor Flexibility Pattern

When creating reusable components that work in different contexts:

```typescript
// GOOD: Accept both string IDs and direct elements
constructor(containerIdOrElement: string | HTMLElement) {
    if (typeof containerIdOrElement === 'string') {
        this.element = document.getElementById(containerIdOrElement);
    } else {
        this.element = containerIdOrElement;
    }
}
```

### Graceful Degradation Pattern

```typescript
// GOOD: Default to sensible behavior when configuration is missing
if (this.isNewMap) {
    this.initializeNewMap();
} else if (this.currentMapId) {
    this.loadExistingMap(this.currentMapId);
} else {
    // Graceful fallback instead of error
    this.isNewMap = true;
    this.initializeNewMap();
}
```

### Critical Timing Considerations

#### WebGL Initialization Timing
- WebGL contexts require visible containers with proper dimensions
- Always check `getBoundingClientRect()` before initializing graphics components
- Use polling pattern for layout-dependent initialization

#### EventBus Subscription Timing
- Subscribe to events BEFORE creating components that emit them
- Use the State→Subscribe→Create pattern consistently

#### Component Assignment Timing
- Event emission should be asynchronous when component references are being assigned
- Use `setTimeout(() => emit(), 0)` to allow assignment completion

## Testing and Debugging Patterns

### Progressive Debug Logging
When troubleshooting complex initialization sequences:

1. **Add comprehensive logging** to understand timing and state
2. **Identify root causes** through systematic debugging
3. **Remove debug logging** once issues are resolved
4. **Document the learnings** in design principles

### Container Dimension Debugging
```typescript
// Useful debug pattern for layout issues
const rect = element.getBoundingClientRect();
console.log(`Dimensions: ${rect.width}x${rect.height}, Visible: ${rect.width > 0 && rect.height > 0}`);
```

### Event Binding Debugging
```typescript
// Useful pattern for debugging missing event handlers
private bindPhaserPanelEvents(container: HTMLElement): void {
    const checkbox = container.querySelector('#show-grid') as HTMLInputElement;
    if (checkbox) {
        checkbox.addEventListener('change', (e) => {
            this.logToConsole(`Grid checkbox changed to: ${checked}`);
            this.setShowGrid(checked);
        });
        this.logToConsole('Grid checkbox event handler bound');
    } else {
        this.logToConsole('Grid checkbox not found in Phaser panel');
    }
}\n```

### Lifecycle Architecture Debugging
When components don't respond to UI interactions:
1. **Check timing**: Are events bound before or after template creation?
2. **Check scope**: Are you using `document.getElementById()` or `container.querySelector()`?
3. **Check references**: Are component references null due to initialization order?
4. **Check EventBus**: Are events being emitted and received correctly?