# ./web/templates/ Summary

**Purpose:**

This folder contains all the Go HTML template files (`*.html`) used for server-side rendering (SSR) and defining client-side template fragments. It utilizes the `html/template` syntax extended by the `templar` engine and heavily uses Tailwind CSS classes for styling.

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
*   `MapListingPage.html`, `MapDetailPage.html`: Maps management templates for listing and viewing individual maps.
*   `MapList.html`: Reusable component for displaying maps in a grid layout with search, sort, and action controls.
*   `DesignList.html`, `TableOfContents.html`, `DocumentTitle.html`, `SectionsList.html`: Templates for major reusable UI components rendered server-side initially (and potentially updated client-side later).
*   `gen/`: Subfolder containing bundled JavaScript output referenced by page templates (e.g., `{{# include "gen/DesignEditorPage.html" #}}`).

**Key Concepts/Responsibilities:**

*   **HTML Structure:** Defines the semantic HTML for all pages and components.
*   **SSR Rendering:** Provides the templates processed by the Go backend (`web/frontend`) to generate the initial HTML sent to the browser.
*   **Client-Side Templates:** Acts as a registry for HTML fragments used by TypeScript components to dynamically update the UI without full page reloads.
*   **Styling:** Relies heavily on Tailwind CSS utility classes defined within the HTML.
