
## UI Design Principles

* UI Components (eg Pages, Views, Panels) must be scoped in their responsibilities.
* A component should be initialized with a "Root" element (eg a Div).
* A component must only access the html elements within its root Element and that too only to create custom components
  classes from them.
* A component must only get references to its child components as it owns them.
* A component can only modify parameters of other components it refers to changes its behavior and properties.  It must
  NEVER search for HTML elements in other components as it violates the seperation of concerns principle.
* A component must use a page level event bus to send events so other components can react to notifications.  Event
  buses should be synchronous and allow multiple entities to act on events.  This is in addition to event handlers.
  An event should not be sent to the source of the event. 
  
