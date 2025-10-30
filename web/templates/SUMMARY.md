# ./web/templates/ Summary

**Purpose:**

This folder contains all the Go HTML template files (`*.html`) used for server-side rendering (SSR) and defining client-side template fragments. The template system uses Go's `html/template` package with custom extensions via the `templar` engine, which provides advanced composition features. All templates heavily use Tailwind CSS utility classes for styling.

**Key Files:**

*   `BasePage.html`: The main site layout template. Defines the `<html>`, `<head>`, and `<body>` structure. Includes placeholders (`{{ block }}`) for header, main body content, and footer/post-body scripts. Also includes containers for modals (`ModalContainer.html`) and toasts (`ToastContainer.html`). Sets up basic dark mode script.
*   `Header.html`: Template for the top navigation bar, including logo, site name, theme toggle button, mobile menu button, and a block for extra page-specific buttons.
*   `ModalContainer.html`, `ToastContainer.html`: Define the basic structure and IDs for modal dialogs and toast notifications managed by corresponding TypeScript singletons.
*   `TemplateRegistry.html`: **Crucial File.** Contains definitions for various UI fragments loaded dynamically by client-side TypeScript using the `TemplateLoader`. These are *not* standard Go templates but rather HTML snippets wrapped in `<div data-template-id="...">`. Examples include:
    *   Modals: `llm-dialog`, `llm-results`, `section-type-selector`, `create-design-modal`.
    *   Section-specific view/edit structures: `text-section-view`, `text-section-edit`, `drawing-section-view`, `drawing-section-edit`, `plot-section-view`, `plot-section-edit`, `system_description-section-view`, `system_description-section-edit`.
    *   Other UI elements: `suggested-section-card`.
*   `Section.html`: Template for the outer frame of a single section, including its header (number, title, icon), universal controls (move, delete, add, settings, LLM, fullscreen), section-specific action bars (for System Description view/edit), and the main content container (`.section-content`) where the specific view/edit component renders.
*   `HomePage.html`, `DesignEditorPage.html`, `LoginPage.html`: Top-level page templates that include `BasePage.html` and compose other smaller templates (like `DesignList.html`, `TableOfContents.html`, `DocumentTitle.html`).
*   `WorldListingPage.html`, `WorldViewerPage.html`, `WorldEditorPage.html`: Complete worlds management templates for listing, viewing, and editing worlds.
*   `WorldList.html`: Reusable component for displaying worlds in a table layout with search, sort, action controls, and screenshot thumbnails.
*   `GameListingPage.html`, `GameViewerPage.html`: Complete games management templates for listing and viewing games.
*   `GameList.html`: Reusable component for displaying games in a table layout with screenshot thumbnails and metadata.
*   `WorldEditorPage.html`: **Interactive canvas-based world editor** with streamlined 2-panel layout:
    *   Left sidebar: World management, grid-based terrain palette (6 types), brush settings, painting tools, history controls
    *   Center: Real-time updating HTML5 canvas with hex grid visualization and world resize controls
    *   Right sidebar: Advanced tools only (rendering/export panels removed)
    *   Interactive canvas with click-to-paint terrain and Add/Remove buttons for world resizing
    *   WASM integration ready with clean data-attribute event handling
*   `DesignList.html`, `TableOfContents.html`, `DocumentTitle.html`, `SectionsList.html`: Templates for major reusable UI components rendered server-side initially (and potentially updated client-side later).
*   `gen/`: Subfolder containing bundled JavaScript output referenced by page templates (e.g., `{{# include "gen/DesignEditorPage.html" #}}`).

**Key Concepts/Responsibilities:**

*   **HTML Structure:** Defines the semantic HTML for all pages and components.
*   **SSR Rendering:** Provides the templates processed by the Go backend (`web/frontend`) to generate the initial HTML sent to the browser.
*   **Client-Side Templates:** Acts as a registry for HTML fragments used by TypeScript components to dynamically update the UI without full page reloads.
*   **Styling:** Relies heavily on Tailwind CSS utility classes defined within the HTML.

---

## Template System Architecture

### Templar Engine

The template system uses a custom `templar` engine (imported from `github.com/panyam/goutils`) which extends Go's standard `html/template` with additional features:

**Key Features:**
- **Template Composition:** `{{# include "path/to/template.html" #}}` syntax for including other templates
- **Block Definitions:** `{{ block "name" . }}...{{ end }}` for defining overridable sections
- **Nested Includes:** Templates can include other templates recursively
- **Path Resolution:** Automatic template discovery from configured root directories

**Template Processing Flow:**
1. Backend loads templates from `web/templates/` directory
2. Templar parses templates and resolves all includes
3. Go processes `{{ block }}`, `{{ range }}`, `{{ if }}` directives
4. Final HTML sent to browser with embedded data

### Server-Side Rendering (SSR)

Templates are rendered on the backend in `web/frontend/server.go`:

```go
// Example: Rendering a page with data
tmpl := templar.Get("WorldEditorPage.html")
err := tmpl.Execute(w, data)
```

**Data Binding:**
- Templates receive data via the `.` (dot) context
- Access fields: `{{ .WorldId }}`, `{{ .GameId }}`, etc.
- Range over collections: `{{ range .Worlds }}...{{ end }}`
- Conditional rendering: `{{ if .IsLoggedIn }}...{{ end }}`

**Common SSR Patterns:**
1. **Page Templates:** Top-level templates that extend BasePage.html
2. **Component Templates:** Reusable fragments included by multiple pages
3. **Generated Templates:** `gen/` folder contains build-time generated templates with bundled JS

### Client-Side Template Loading

The `TemplateLoader` singleton (in `web/src/lib/TemplateLoader.ts`) manages dynamic template loading:

**How It Works:**
1. Client-side TypeScript requests a template by ID
2. TemplateLoader fetches HTML from `TemplateRegistry.html` via selector
3. Template fragment is cloned and inserted into DOM
4. Event handlers and data binding applied by TypeScript components

**Template Registry Format:**
```html
<!-- In TemplateRegistry.html -->
<div data-template-id="my-template">
  <div class="...">
    <!-- Template content -->
  </div>
</div>
```

**Usage in TypeScript:**
```typescript
const html = TemplateLoader.getInstance().loadTemplate("my-template");
container.innerHTML = html;
```

### Generated Templates (`gen/` folder)

Build process generates templates that bundle TypeScript/JavaScript:

**Generation Process:**
1. `web/src/` TypeScript compiled by esbuild
2. Output JavaScript written to `web/gen/`
3. Corresponding `.html` templates in `web/templates/gen/` reference the JS
4. Page templates include generated templates: `{{# include "gen/GameViewerPage.html" #}}`

**Key Generated Pages:**
- `GameViewerPage.html` - Main game UI with Phaser canvas
- `WorldEditorPage.html` - Interactive world editor
- `StartGamePage.html` - Game initialization and setup

**Benefits:**
- Code splitting per page
- Bundle optimization
- Separate TypeScript compilation per page module
- Clean separation between SSR data and client JS

### WASM Integration

Templates interact with WASM modules for game logic:

**Pattern:**
1. Page template loads WASM module via script tag
2. TypeScript initializes WASM and creates client instances
3. Event handlers call WASM functions via typed interfaces
4. WASM responses update DOM through TypeScript components

**Example Flow (Game Viewer):**
```
User clicks tile →
TypeScript handler →
WASM GetOptionsAt() →
TypeScript updates UI with options →
User selects move →
TypeScript ProcessMove() →
WASM validates/executes →
TypeScript refreshes scene
```

### Dark Mode & Theming

Implemented via inline script in `BasePage.html`:

**Mechanism:**
- Checks `localStorage.theme` or system preference
- Adds/removes `dark` class on `<html>` element
- Tailwind CSS classes respond: `bg-white dark:bg-gray-900`
- Theme toggle button updates localStorage and class

**Theme Classes:**
- Light: Default Tailwind colors
- Dark: `dark:` prefix variants (bg-gray-900, text-white, etc.)

### Modal & Toast Systems

Managed by TypeScript singletons with template-defined structure:

**Modals (`ModalContainer.html`):**
- Fixed overlay with centered content area
- TypeScript `ModalManager` controls visibility
- Dynamic content loaded from TemplateRegistry
- Backdrop click to dismiss

**Toasts (`ToastContainer.html`):**
- Fixed position notifications (top-right)
- TypeScript `ToastManager` shows/hides messages
- Auto-dismiss after timeout
- Success/error/info styling variants

### Component Update Patterns

**Full Page Reload:**
```go
// Backend redirects after action
http.Redirect(w, r, "/games/"+gameId, http.StatusSeeOther)
```

**Partial Update (AJAX):**
```typescript
// Client fetches HTML fragment
const response = await fetch("/api/games/"+gameId+"/options");
const html = await response.text();
container.innerHTML = html;
```

**WASM-Driven Update:**
```typescript
// Direct DOM manipulation from WASM response
const state = wasmClient.GetGameState();
updateScene(state); // TypeScript updates Phaser or DOM
```

### Best Practices

1. **Separation of Concerns:**
   - Templates define structure (HTML)
   - Tailwind defines styling (CSS classes)
   - TypeScript defines behavior (JS)
   - WASM defines game logic

2. **Template Naming:**
   - Pages: `*Page.html` (e.g., GameViewerPage.html)
   - Components: Descriptive names (e.g., WorldList.html)
   - Modals: `*Modal.html` or registered in TemplateRegistry
   - Generated: `gen/*.html`

3. **Data Flow:**
   - SSR: Backend → Template → HTML
   - Client: User Action → TypeScript → WASM → DOM Update
   - Avoid mixing SSR data with client state

4. **Performance:**
   - Use TemplateLoader for fragments (avoid full rerenders)
   - Cache WASM responses when appropriate
   - Lazy-load heavy components (Phaser scenes)
   - Debounce rapid updates (e.g., canvas painting)

5. **Accessibility:**
   - Use semantic HTML elements
   - Include ARIA labels where needed
   - Ensure keyboard navigation works
   - Test with screen readers
