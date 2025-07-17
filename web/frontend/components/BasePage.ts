import { ThemeManager } from './ThemeManager';
import { Modal } from './Modal';
import { ToastManager } from './ToastManager';

/**
 * Base class for all pages that provides common UI components and functionality
 */
export abstract class BasePage {
    protected themeManager: typeof ThemeManager | null = null;
    protected modal: Modal | null = null;
    protected toastManager: ToastManager | null = null;

    protected themeToggleButton: HTMLButtonElement | null = null;
    protected themeToggleIcon: HTMLElement | null = null;

    constructor() {
        this.initializeBaseComponents();
        this.bindBaseEvents();
    }

    /**
     * Initialize common UI components that all pages need
     */
    protected initializeBaseComponents(): void {
        // Initialize core UI managers
        ThemeManager.init();
        this.modal = Modal.init();
        this.toastManager = ToastManager.init();

        // Get theme toggle elements
        this.themeToggleButton = document.getElementById('theme-toggle-button') as HTMLButtonElement;
        this.themeToggleIcon = document.getElementById('theme-toggle-icon');

        if (!this.themeToggleButton || !this.themeToggleIcon) {
            console.warn("Theme toggle button or icon element not found in Header.");
        }
    }

    /**
     * Bind common event handlers that all pages need
     */
    protected bindBaseEvents(): void {
        // Theme toggle
        if (this.themeToggleButton) {
            this.themeToggleButton.addEventListener('click', this.handleThemeToggleClick.bind(this));
        }

        // Initialize theme button state
        this.updateThemeButtonState();
    }

    /**
     * Handle theme toggle button clicks
     */
    protected handleThemeToggleClick(): void {
        const currentSetting = ThemeManager.getCurrentThemeSetting();
        const nextSetting = ThemeManager.getNextTheme(currentSetting);
        ThemeManager.setTheme(nextSetting);
        this.updateThemeButtonState(nextSetting);
    }

    /**
     * Update the theme toggle button state and appearance
     */
    protected updateThemeButtonState(currentTheme?: string): void {
        if (!this.themeToggleButton || !this.themeToggleIcon) return;

        const themeToDisplay = currentTheme || ThemeManager.getCurrentThemeSetting();
        const iconSVG = ThemeManager.getIconSVG(themeToDisplay);
        const label = `Toggle theme (currently: ${ThemeManager.getThemeLabel(themeToDisplay)})`;

        this.themeToggleIcon.innerHTML = iconSVG;
        this.themeToggleButton.setAttribute('aria-label', label);
        this.themeToggleButton.setAttribute('title', label);
    }

    /**
     * Show a toast notification
     */
    protected showToast(title: string, message: string, type: 'success' | 'error' | 'info' | 'warning' = 'info', duration?: number): void {
        this.toastManager?.showToast(title, message, type, duration);
    }

    /**
     * Show a modal dialog
     */
    protected showModal(title: string, content: string, options?: any): void {
        this.modal?.show(title, content, options);
    }

    /**
     * Hide the modal dialog
     */
    protected hideModal(): void {
        this.modal?.hide();
    }

    /**
     * Get the current theme setting
     */
    protected getCurrentTheme(): string {
        return ThemeManager.getCurrentThemeSetting();
    }

    /**
     * Check if the current theme is dark mode
     */
    protected isDarkMode(): boolean {
        return document.documentElement.classList.contains('dark');
    }

    /**
     * Abstract method that subclasses must implement to initialize their specific components
     */
    protected abstract initializeSpecificComponents(): void;

    /**
     * Abstract method that subclasses must implement to bind their specific events
     */
    protected abstract bindSpecificEvents(): void;

    /**
     * Abstract method that subclasses can implement to handle cleanup
     */
    public abstract destroy?(): void;
}