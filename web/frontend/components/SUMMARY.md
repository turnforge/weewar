# ./web/frontend/components/ Summary

**Purpose:**

This folder contains the core client-side TypeScript logic for LeetCoach interactivity, managing UI state, user events, API interactions, and DOM manipulation using a Composition pattern for section rendering and avoiding major frameworks (except for Excalidraw's React and Toast UI Editor dependencies).

**Key Files/Components & Recent Refactoring:**

*   **Composition Refactor:** Sections are managed by `BaseSection.ts` (as `SectionContainer`), which instantiates mode-specific `[Type]View.ts` and `[Type]Edit.ts` components. These mode components now handle their own data fetching (`ContentApi`) and saving logic.
    *   **Refactored:** `Text`, `Drawing`, `SystemDescription`.
    *   **Pending Refactor:** `Plot`.

*   **`BaseSection.ts` (Section Container):** Manages section frame, universal controls, mode state, orchestrates `View`/`Edit` component lifecycle, passes `designId`/`sectionId` and `onSaveSuccess`/`onCancel` callbacks.
*   **View Mode Components:**
    *   `TextSectionView.ts`: Fetches Markdown, renders it as HTML using Toast UI Editor Viewer. Handles theme changes.
    *   `DrawingSectionView.ts`: Fetches SVG previews, renders the correct one. Handles theme changes.
    *   `SystemDescriptionView.ts`: Fetches DSL, displays in `<pre><code>`, has Validate/Generate Diagram buttons, calls respective APIs, displays results/placeholders.
*   **Edit Mode Components:**
    *   `TextSectionEdit.ts`: Initializes **Toast UI Editor** with fetched Markdown. Handles content changes, Save (gets Markdown, calls `ContentApi`), Cancel. Signals container via callbacks. Handles theme changes.
    *   `DrawingSectionEdit.ts`: Initializes Excalidraw (via `ExcalidrawWrapper.tsx`). Handles Save (generates JSON/SVGs, calls `ContentApi` multiple times). Signals container. Handles theme changes.
    *   `SystemDescriptionEdit.ts`: Uses `<textarea>` for DSL. Handles Save (calls `ContentApi`), Cancel. Validate button calls `SystemModelApi`.
*   **Managers & Handlers:** (`SectionManager.ts`, `LlmInteractionHandler.ts`, `ThemeManager.ts`, `Modal.ts`, `ToastManager.ts`, `TableOfContents.ts`, `DocumentTitle.ts`, `FullscreenHandler.ts`) - Core logic largely the same, but `SectionManager` now manages `BaseSection` (container) instances.
*   **Page Entry Points:** (`DesignEditorPage.ts`, `HomePage.ts`, `LoginPage.ts`, `MapDetailsPage.ts`) - `DesignEditorPage` simplified as containers manage their own content loading. `MapDetailsPage.ts` provides foundation for maps functionality.
*   **Utilities:** (`Api.ts`, `TemplateLoader.ts`, `types.ts`, `converters.ts`, `ExcalidrawWrapper.tsx`).

**Key Concepts/Responsibilities:**

*   Client-Side Interactivity, Component Orchestration, API Interaction, DOM Manipulation.
*   **Composition Pattern:** Encapsulated view/edit logic for sections.
*   Markdown Editing/Viewing for Text sections via Toast UI Editor.
