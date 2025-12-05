import { ThemeManager, Modal, ToastManager } from '@panyam/tsappkit';

/**
 * Main application initialization
 */
class UserDetailsPage {
    private themeManager: typeof ThemeManager | null = null;
    private modal: Modal | null = null;
    private toastManager: ToastManager | null = null;

    private themeToggleButton: HTMLButtonElement | null = null;
    private themeToggleIcon: HTMLElement | null = null;

    private currentUserId: string | null = null;
    private isLoadingUser: boolean = false; // Loading state

    constructor() {
        this.initializeComponents();
        this.bindEvents();
        this.loadInitialState();
    }

    private initializeComponents(): void {
        const designIdInput = document.getElementById("designIdInput") as HTMLInputElement | null;
        const designId = designIdInput?.value.trim() || null; // Allow null if input not found/empty

        ThemeManager.init();
        this.modal = Modal.init();
        this.toastManager = ToastManager.init();

        this.themeToggleButton = document.getElementById('theme-toggle-button') as HTMLButtonElement;
        this.themeToggleIcon = document.getElementById('theme-toggle-icon');

        if (!this.themeToggleButton || !this.themeToggleIcon) {
            console.warn("Theme toggle button or icon element not found in Header.");
        }
    }

    private bindEvents(): void {
        if (this.themeToggleButton) {
            this.themeToggleButton.addEventListener('click', this.handleThemeToggleClick.bind(this));
        }

        const mobileMenuButton = document.getElementById('mobile-menu-button');
        if (mobileMenuButton) {
            mobileMenuButton.addEventListener('click', () => {
              // Do things like sidebar drawers etc
            });
        }

        const saveButton = document.querySelector('header button.bg-blue-600');
        if (saveButton) {
            saveButton.addEventListener('click', this.saveDocument.bind(this));
        }

        const exportButton = document.querySelector('header button.bg-gray-200');
        if (exportButton) {
            exportButton.addEventListener('click', this.exportDocument.bind(this));
        }
    }

    /** Load document data and set initial UI states */
    private loadInitialState(): void {
        this.updateThemeButtonState();

        const designIdInput = document.getElementById("designIdInput") as HTMLInputElement | null;
        const designId = designIdInput?.value.trim() || null;

        if (designId) {
            this.currentUserId = designId;
            this.loadUserData(this.currentUserId);
        } else {
            console.error("User ID input element not found or has no value. Cannot load document.");
            this.toastManager?.showToast("Error", "Could not load document: User ID missing.", "error");
        }
    }

    /**
     * Fetches design metadata, initializes section shells, and triggers content loading for each section.
     */
    private async loadUserData(designId: string): Promise<void> {
        // TODO: Show global loading indicator
        // here is where we would do "reload" via ajax - this coul dbe via ajax or via htmx
    }

    /** Handles click on the new theme toggle button */
    private handleThemeToggleClick(): void {
        const currentSetting = ThemeManager.getCurrentThemeSetting();
        const nextSetting = ThemeManager.getNextTheme(currentSetting);
        ThemeManager.setTheme(nextSetting);
        this.updateThemeButtonState(nextSetting); // Update icon to reflect the *new* state

        // Optional: Show toast feedback
        // this.toastManager?.showToast('Theme Changed', `Switched to ${ThemeManager.getThemeLabel(nextSetting)}`, 'info', 2000);

        // Notify all other child components when themes have changed - or we could do this via event bus
    }
 
    /** Updates the theme toggle button's icon and aria-label */
    private updateThemeButtonState(currentTheme?: string): void {
        if (!this.themeToggleButton || !this.themeToggleIcon) return;

        const themeToDisplay = currentTheme || ThemeManager.getCurrentThemeSetting();
        const iconSVG = ThemeManager.getIconSVG(themeToDisplay);
        const label = `Toggle theme (currently: ${ThemeManager.getThemeLabel(themeToDisplay)})`;

        this.themeToggleIcon.innerHTML = iconSVG;
        this.themeToggleButton.setAttribute('aria-label', label);
        this.themeToggleButton.setAttribute('title', label); // Add tooltip
    }

    /** Save document (Placeholder - needs full implementation later) */
    private saveDocument(): void {
        // This full save logic will be replaced by incremental saves triggered by component callbacks
        this.toastManager?.showToast('Save Action', 'Incremental saves handle updates. Full save TBD.', 'info');
    }

    /** Export document (Placeholder) */
    private exportDocument(): void {
        if (this.toastManager) {
            this.toastManager.showToast('Export started', 'Your document is being prepared for export.', 'info');
            setTimeout(() => {
                this.toastManager?.showToast('Export complete', 'Document export simulation finished.', 'success');
            }, 1500);
        }
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const lc = new UserDetailsPage();
});
