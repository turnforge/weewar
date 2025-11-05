**Purpose:**

This folder contains the core client-side TypeScript logic for the library components used in the WeeWar UI (in the ../src folder). These
components help with lifecycled loading, eventbuses, structured pages and DOM manipulation using a modern component-based architecture with strict separation of concerns and event-driven communication.

**Core Architecture Components:**

## Modern Component System (New)

*   **`EventBus.ts`**: Type-safe, synchronous event system with error isolation and source exclusion for inter-component communication
*   **`Component.ts`**: Base interface and abstract class defining standard component lifecycle with simplified constructor pattern
*   **`DESIGN_PRINCIPLES.md`**: Comprehensive documentation of architecture decisions, timing patterns, and critical lessons learned

## Component Features

*   **Strict DOM Scoping**: Components only access DOM within their root elements using `this.findElement()`
*   **Event-Driven Communication**: All inter-component communication through EventBus, no direct method calls
*   **Layout vs Behavior Separation**: Parents control layout/sizing, components handle internal behavior only
*   **HTMX Integration Ready**: Components support both initialization and hydration patterns
*   **Error Isolation**: Component failures don't cascade to other components
*   **Simplified Constructor Pattern**: `new Component(rootElement, eventBus)` - parent ensures root element exists

## Key Architecture Principles

*   **Separation of Concerns**: Clear boundaries between layout, behavior, and communication responsibilities
*   **Event-Driven**: Components communicate through EventBus events, never direct method calls  
*   **DOM Isolation**: Components only access DOM within their assigned root elements
*   **Error Resilience**: Component failures are isolated and don't affect other components
*   **Timing Awareness**: Proper handling of initialization order, race conditions, and async operations
*   **WebGL Integration**: Specialized patterns for graphics libraries like Phaser with timing considerations

## Critical Timing Patterns Learned

*   **TypeScript Field Initializers**: Avoid explicit `= null` for constructor-set fields
*   **Event Subscription Order**: Subscribe to events BEFORE creating components that emit them
*   **WebGL Context Readiness**: Use small setTimeout for graphics library initialization completion
*   **State → Subscribe → Create**: Strict three-phase initialization order
*   **Async in Handlers**: EventBus stays synchronous, handlers use `.then()/.catch()` for async operations

## Integration Capabilities

*   **HTMX**: Component hydration support for server-driven UI updates
*   **Toast/Modal Systems**: User feedback and interaction patterns
*   **Theme Management**: Coordinated theming across component boundaries

## Responsive UI Patterns (BasePage.ts)

### Responsive Header Menu System

**Purpose**: Provides adaptive menu behavior for header action buttons across all pages

**Implementation** (`initializeHeaderActionsDropdown()`):
- Detects screen width using 768px breakpoint (matches Tailwind `md:`)
- Clones buttons with `header-action-btn` class from source container
- Mobile (<768px): Opens animated drawer below header
- Desktop (≥768px): Opens traditional dropdown menu

**Mobile Drawer Features**:
- Full-screen overlay with backdrop (z-index 60)
- Slide-down animation (300ms, ease-out)
- Positioned at `top: 70px` (below header)
- Backdrop fade: opacity 0 → 0.5
- Maintains button styling (full-width layout)
- Closes on backdrop click or Escape key

**Desktop Dropdown Features**:
- Compact list-style menu
- Positioned below trigger button
- Hover effects on items
- Outside-click dismissal

**Animation Technique**:
- Uses `requestAnimationFrame` for smooth transitions
- Transform: `-translate-y-full` → `translate-y-0`
- Delayed cleanup (300ms) after close animation completes
- Prevents jarring visibility changes

**Why This Approach**:
- **Overflow Safety**: Drawer escapes `overflow:hidden` constraints common on mobile
- **Platform Consistency**: Drawer pattern standard on mobile, dropdown on desktop
- **Universal**: Works across all pages without page-specific modifications
- **Accessible**: Keyboard navigation (Escape), backdrop dismissal

**Usage**: Any page using Header.html automatically gets responsive menu behavior

