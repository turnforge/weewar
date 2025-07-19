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

**Broadcast Communication**
- When one component needs to notify multiple other components
- Example: Map data changes that need to update stats, save indicators, etc.
- Rationale: One-to-many communication pattern

### Component Architecture Patterns

#### State→Subscribe→Create Pattern
All page-level components should follow this initialization order:
1. **State**: Set up initial state and data
2. **Subscribe**: Subscribe to EventBus events before components are created
3. **Create**: Initialize child components and UI elements

#### Separation of Concerns
- **Layout Components**: Handle only UI layout and direct user interactions
- **Business Logic Components**: Handle data processing and application logic
- **Communication**: Use appropriate pattern (direct calls vs EventBus) based on relationship

#### Error Isolation
- Each component should handle its own errors
- EventBus provides automatic error isolation between components
- Failed event handlers don't affect other subscribers

### Examples

#### ✅ Good: Parent→Child Direct Call
```typescript
// MapEditorPage toggling grid on its child component
public setShowGrid(showGrid: boolean): void {
    if (this.phaserEditorComponent && this.phaserEditorComponent.getIsInitialized()) {
        this.phaserEditorComponent.setShowGrid(showGrid);
    }
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

#### ❌ Anti-Pattern: Unnecessary EventBus for Parent→Child
```typescript
// DON'T DO THIS - parent already has direct access to child
this.eventBus.emit('GRID_TOGGLE', { showGrid: true }, 'map-editor-page');
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