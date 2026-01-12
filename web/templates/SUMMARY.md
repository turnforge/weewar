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

The template system uses a custom `templar` engine which extends Go's standard `html/template` with composition features:

**Key Directives:**
- **Namespace:** `{{# namespace "lilbattle" #}}` - Defines template namespace for isolation
- **Include:** `{{# include "path/to/template.html" #}}` - Includes another template inline
- **Extend:** `{{# extend "goapplib/BasePage.html" #}}` - Inherits from a parent template
- **Block Definitions:** `{{ define "BlockName" }}...{{ end }}` - Define/override blocks

**goapplib Integration:**
Templates extend shared components from the `goapplib` package via templar vendoring:
- `@goapplib/BasePage.html` - Base page layout with header, body, scripts blocks
- `@goapplib/components/EntityListing.html` - Reusable entity listing with table/grid views
- Pages override blocks like `Header`, `Body`, `ExtraHeaderButtons`
- Shared components reduce duplication across projects

**Templar Vendoring (2025-01-12):**
Dependencies are managed via `templar.yaml` and fetched to `templar_modules/`:

```yaml
# templar.yaml
sources:
  goapplib:
    url: github.com/panyam/goapplib
    path: templates          # Only fetch templates/ subdirectory
    ref: main
    # include: ["**/*.html"] # Optional: glob patterns to include
    # exclude: ["*_test.*"]  # Optional: glob patterns to exclude

vendor_dir: ./templar_modules
search_paths:
  - .
  - ./templar_modules
```

**Vendored Directory Structure (flat):**
```
web/templates/
├── templar.yaml           # Configuration
├── templar.lock           # Lock file (auto-generated, gitignored)
├── templar_modules/       # Vendored dependencies
│   └── goapplib/          # Flat: sourcename/files...
│       ├── BasePage.html
│       ├── Header.html
│       └── components/
│           ├── EntityListing.html
│           └── ...
├── BasePage.html          # Local override extending @goapplib/BasePage.html
└── ...
```

**@source Syntax:**
Use `@sourcename/path` to reference vendored templates:
```html
{{# namespace "GoalBase" "@goapplib/BasePage.html" #}}
{{# namespace "EL" "@goapplib/components/EntityListing.html" #}}
```

**Commands:**
- `templar get` - Fetch/update all dependencies
- `TEMPLAR_DEBUG=1 templar get` - Debug mode showing extracted files

**Template Hierarchy Example:**
```html
{{# namespace "lilbattle" #}}
{{# include "goapplib/BasePage.html" #}}
{{# extend "goapplib/BasePage.html" #}}

{{ define "Header" }}
  {{# include "Header.html" #}}
{{ end }}

{{ define "Body" }}
  <!-- Page-specific content -->
{{ end }}
```

**Path Resolution:** Automatic template discovery from configured root directories

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

### Responsive Layout Patterns

#### Bottom Sheet Pattern (Mobile Overlays)

**Purpose:** Provides mobile-optimized UI by showing secondary panels as slide-up overlays instead of fixed sidebars.

**Implementation Pattern:**

1. **Desktop Layout:** Two-column with fixed sidebar
```html
<div class="flex h-screen">
  <!-- Main content (full-width on mobile) -->
  <div class="w-full lg:w-[calc(100%-300px)]">
    <div id="preview-container"></div>
  </div>

  <!-- Sidebar (hidden on mobile) -->
  <div class="hidden lg:block lg:w-[300px]">
    <!-- Panel content -->
  </div>
</div>
```

2. **Mobile FAB Button:**
```html
<button id="stats-fab" class="fixed bottom-6 right-6 z-40 lg:hidden ...">
  <svg><!-- Icon --></svg>
  <span>Stats</span>
</button>
```

3. **Bottom Sheet Overlay:**
```html
<div id="stats-overlay" class="fixed inset-0 z-50 hidden lg:hidden">
  <!-- Semi-transparent backdrop -->
  <div id="stats-backdrop" class="absolute inset-0 bg-black/50"></div>

  <!-- Slide-up panel -->
  <div id="stats-panel" class="absolute bottom-0 left-0 right-0 bg-white rounded-t-2xl shadow-xl transform translate-y-full transition-transform duration-300 max-h-[85vh] flex flex-col">
    <!-- Handle bar -->
    <div class="w-12 h-1 bg-gray-300 rounded-full mx-auto mt-3 mb-2"></div>

    <!-- Header with close button -->
    <div class="flex items-center justify-between p-4 border-b">
      <h2>Panel Title</h2>
      <button id="stats-close">×</button>
    </div>

    <!-- Scrollable content -->
    <div class="flex-1 overflow-y-auto p-4">
      <!-- Panel content here -->
    </div>
  </div>
</div>
```

4. **Media Query Controls:**
```html
<style>
  /* Ensure FAB only shows on mobile */
  @media (max-width: 1023px) {
    #stats-fab { display: flex !important; }
  }
  @media (min-width: 1024px) {
    #stats-fab, #stats-overlay { display: none !important; }
  }
</style>
```

**TypeScript Integration:** See `web/src/SUMMARY.md` for handler implementation.

**Pages Using Pattern:**
- `WorldViewerPage.html` - Stats panel with FAB button
- `StartGamePage.html` - Config panel with FAB button

#### Responsive Header Menu System

**Purpose:** Desktop buttons collapse into responsive menu on mobile - drawer for narrow screens with overflow constraints, dropdown for wide screens.

**Implementation in Header.html:**
- Desktop (≥768px): Buttons shown inline via `md:flex` visibility
- Mobile (<768px): Three-dot menu button triggers drawer or dropdown based on constraints
- **Header Actions Drawer**: Full-overlay drawer that slides down from header (z-index 60)
- **Header Actions Dropdown**: Positioned dropdown for unconstrained layouts (z-index 50)
- TypeScript: `initializeHeaderActionsDropdown()` in BasePage.ts handles responsive switching

**Architecture:**
```
┌─ All Pages ─────────────────────────────────┐
│ Header.html (shared across all pages)       │
│ ├─ Desktop: Buttons inline (#header-buttons-desktop) │
│ ├─ Mobile: "..." menu button                │
│ ├─ Dropdown: Fixed positioning (#header-actions-dropdown) │
│ └─ Drawer: Full overlay (#header-actions-drawer) │
└──────────────────────────────────────────────┘

┌─ BasePage.ts ───────────────────────────────┐
│ initializeHeaderActionsDropdown()           │
│ ├─ Clones buttons from source container     │
│ ├─ Detects screen width (768px breakpoint)  │
│ ├─ Mobile: Opens drawer with slide animation│
│ └─ Desktop: Opens dropdown below button     │
└──────────────────────────────────────────────┘
```

**Drawer Features (Mobile):**
- Full-screen overlay with semi-transparent backdrop
- Positioned directly below header (top: 70px)
- Slide-down/up animation (300ms duration, ease-out)
- Backdrop fade animation (opacity 0 → 0.5)
- Auto-close on backdrop click or Escape key
- Buttons maintain full styling with full-width layout

**Dropdown Features (Desktop):**
- Compact menu positioned below button
- List-style items with hover effects
- Traditional dropdown styling

**Usage Pattern:**
```html
{{ block "ExtraHeaderButtons" . }}
  <!-- Add 'header-action-btn' class to include in mobile menu -->
  <button class="header-action-btn px-4 py-2 ...">Edit</button>
  <button class="header-action-btn px-4 py-2 ...">Create Game</button>
{{ end }}
```

**Architecture (Updated 2025-01-05):**
- Single global drawer (Header.html) positioned via CSS - no DOM manipulation
- Desktop (≥768px): Drawer positioned absolutely inline with header, always visible
- Mobile (<768px): Drawer slides down from top when "..." button clicked
- Buttons remain in drawer permanently - event listeners never lost
- CSS handles all responsive positioning - JavaScript only toggles visibility

**Technical Details:**
- Breakpoint: 768px (matches Tailwind `md:` breakpoint)
- Drawer z-index: 60 (above all content)
- Desktop: `position: absolute; right: 60px; top: 0;` (inline with header)
- Mobile: `position: fixed; inset: 0;` with `top: 70px` container
- Drawer animation: `transform: translateY(-100%)` → `translateY(0)` (300ms)
- Backdrop animation: `opacity: 0` → `opacity: 1; background: rgba(0,0,0,0.5)` (300ms)
- Event handling: `requestAnimationFrame` for smooth open, 300ms delay for close
- Button stacking: Horizontal on desktop, vertical on mobile (CSS flex-direction)

**Why Drawer for Mobile:**
- Avoids overflow:hidden clipping issues on mobile layouts
- Consistent with mobile drawer patterns (drawers > dropdowns on small screens)
- Works across all pages regardless of page-specific overflow constraints
- Full-screen backdrop prevents interaction with underlying content

**Pages Using Pattern:**
- All pages with Header.html (universal implementation)
- GameViewerPageMobile (End Turn button in menu)
- WorldViewerPage (Edit/Create Game buttons in menu)
- StartGamePage (action buttons in menu)

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

6. **Responsive Design:**
   - Use Tailwind breakpoints (md: 768px, lg: 1024px)
   - Add explicit media queries with `!important` for critical show/hide behavior
   - Test both mobile (< 768px) and desktop (> 1024px) layouts
   - Ensure FAB buttons are properly positioned and accessible on mobile

### Reusable Components (Session 2025-11-03)

#### WorldFilterPanel.html
Reusable search and filter controls for world listings.

**Parameters:**
- `SearchInputId`: ID for search input element
- `SortSelectId`: ID for sort select element
- `HideCreateButton`: Boolean to hide "Create New" button
- `HtmxEnabled`: Boolean to enable HTMX attributes for live filtering
- `Query`: Current search query
- `Sort`: Current sort order

**Features:**
- Search input with icon
- Sort dropdown (modified desc/asc, title A-Z/Z-A)
- Optional "Create New World" button
- HTMX support for live search/filter

**Usage:**
```go
{{ template "WorldFilterPanel" (dict "SearchInputId" "world-search" "SortSelectId" "world-sort" "HideCreateButton" true "HtmxEnabled" true "Query" .Query "Sort" .Sort) }}
```

#### WorldGrid.html
Grid view for displaying worlds as cards.

**Parameters:**
- `Worlds`: Array of World objects
- `ActionMode`: "manage" or "select" - controls action buttons

**Features:**
- Responsive grid (1-4 columns based on screen size)
- World preview images with aspect-video ratio
- Action mode support:
  - **manage**: Three-dot menu with Edit/Delete/Start Game
  - **select**: Large green "Play" button
- Empty states customized per mode
- Pagination-ready

**Usage:**
```go
{{ template "WorldGrid" (dict "Worlds" .Worlds "ActionMode" .ActionMode) }}
```

#### WorldList.html
Unified world listing component supporting both table and grid views.

**Parameters (from WorldListView):**
- `Worlds`: Array of World objects
- `ViewMode`: "table" or "grid"
- `ActionMode`: "manage" or "select"
- `Paginator`: Pagination state
- `Query`: Search query
- `Sort`: Sort order

**Features:**
- View mode toggle (table/grid) in manage mode
- Table view: Traditional list with dropdown actions
- Grid view: Card-based gallery using WorldGrid component
- Shared pagination controls for both views
- Integrates WorldFilterPanel
- Empty states

**Pages Using:**
- WorldListingPage (manage mode, default table view)
- SelectWorldPage (select mode, default grid view)

#### SelectWorldPage.html
Dedicated page for world selection during game creation.

**Features:**
- Preset to ActionMode="select"
- Grid view by default
- Large "Play" buttons on world cards
- Header buttons: "Manage Worlds", "Create World"
- Simplified workflow for game creation

**Route:** `/worlds/select`

**User Flow:**
```
/games/new (no worldId)
  → redirect to /worlds/select
  → click Play on world
  → /games/new?worldId=X
```

### Mobile GameViewer Templates (Session 2025-01-04)

#### CompactSummaryCard.templar.html
Template for mobile compact card showing terrain and unit selection info.

**Features:**
- Renders server-side by Go presenter via template engine
- Displays terrain type with icon and name
- Shows unit type with icon, name, and health (HP: X/10)
- Conditional rendering (tile-only, unit-only, or both)
- Theme-hydrated images via data attributes
- Responsive spacing and typography

**Template Architecture:**
- Uses Go template syntax (`{{ if }}`, `{{ $theme.GetTerrainName }}`)
- Receives data: `{ Tile, Unit, Theme }`
- HTML rendered in Go, sent to browser via RPC
- Browser hydrates theme images asynchronously

**Integration Pattern:**
```go
// services/gameview_presenter.go
s.CompactSummaryCardPanel.SetCurrentData(ctx, tile, unit)

// cmd/lilbattle-wasm/browser.go
content := renderPanelTemplate(ctx, "CompactSummaryCard.templar.html", map[string]any{
    "Tile":  tile,
    "Unit":  unit,
    "Theme": theme,
})
go b.GameViewerPage.SetCompactSummaryCard(ctx, &v1.SetContentRequest{
    InnerHtml: content,
})
```

**Design:**
- 56px fixed height, absolute positioned below header
- Larger icons (32x32) for touch targets
- Font-semibold labels for visibility
- HP badge with background for emphasis
- Horizontal layout with divider between terrain/unit

### PhaserSceneView Component (Session 2025-01-05)

#### components/PhaserSceneView.html
Reusable BorderLayout template for Phaser scenes solving circular sizing problems.

**Purpose:**
Provides a consistent, maintainable pattern for integrating Phaser scenes into pages with proper sizing constraints to prevent canvas from influencing parent container size.

**Architecture - BorderLayout with 5 Regions:**
- **North**: Optional toolbar/header (fixed size)
- **South**: Optional footer/status bar (fixed size)
- **East**: Optional right sidebar (fixed width)
- **West**: Optional left sidebar (fixed width)
- **Center**: Phaser scene container (takes remaining space, never grows parent)

**Key Parameters:**
- `SceneId`: ID for scene container (default: "phaser-scene-container")
- `FlexMode`: Controls wrapper behavior - "fill" (flex-1 + min-height:0), "fixed" (100% size), "auto" (natural)
- `NorthContent`, `SouthContent`, `EastContent`, `WestContent`: Optional HTML content for regions
- `CenterClass`: Additional CSS classes for center region
- `WrapperClass`: Additional CSS classes for wrapper

**Critical Sizing Constraints:**
- Uses `min-height: 0` and `min-width: 0` on flex children to prevent circular sizing
- FlexMode="fill" applies `flex: 1 1 0%; min-height: 0; min-width: 0;` to wrapper automatically
- Eliminates need for manual wrapper divs in pages
- Ensures one-way sizing flow: parent → canvas (never canvas → parent)

**Standard IDs:**
- `phaser-scene-view-wrapper`: Main container
- `phaser-scene-view-north/south/east/west`: Region containers
- `phaser-scene-view-center`: Center region containing scene
- `[SceneId]`: Actual scene container (customizable)

**Usage Pattern:**
```html
{{/* Simple scene only */}}
{{ template "PhaserSceneView" ( dict
  "SceneId" "my-scene"
  "FlexMode" "fill"
) }}

{{/* With toolbar */}}
{{ define "MyToolbar" }}
  <div class="p-2">Toolbar content</div>
{{ end }}

{{ template "PhaserSceneView" ( dict
  "NorthContent" (template "MyToolbar" .)
  "SceneId" "my-scene"
  "FlexMode" "fill"
) }}
```

**TypeScript Integration:**
Works with all Phaser scene types (PhaserWorldScene, PhaserEditorScene, PhaserGameScene):
```typescript
const container = document.getElementById('my-scene');
this.scene = new PhaserWorldScene(container, this.eventBus);
await this.scene.performLocalInit();
await this.scene.activate();
```

**Template Convention (2025-01-05):**
All templates use `{{ define }}` blocks instead of root-level HTML to avoid leaking globals:

```html
{{# include "components/PhaserSceneView.html" #}}
{{# include "panels/MyToolbar.html" #}}

{{/* Override blocks for regions */}}
{{ define "PhaserSceneView_North" }}
  <div id="phaser-scene-view-north" class="flex-shrink-0" style="flex-shrink: 0">
    {{ template "MyToolbar" }}
  </div>
{{ end }}

{{/* Wrap in named template for reuse */}}
{{ define "MyPanel" }}
  {{ template "PhaserSceneView" (dict "SceneId" "my-scene" "FlexMode" "fill") }}
{{ end }}
```

**Pages Migrated:**
- WorldViewerPage ✅ (scene only, FlexMode="fill")
- WorldEditorPage ✅ (toolbar + scene, FlexMode="fixed")

**Pages To Migrate:**
- GameViewerPageDockView (optional controls + scene)
- GameViewerPageMobile (header + scene + action bar)
- StartGamePage (preview scene)
