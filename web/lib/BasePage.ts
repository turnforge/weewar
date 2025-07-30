import { ThemeManager } from './ThemeManager';
import { Modal } from './Modal';
import { ToastManager } from './ToastManager';
import { EventBus } from './EventBus';
import { BaseComponent } from './Component';
import { LCMComponent } from './LCMComponent';

/**
 * Base class for all pages that provides common UI components and functionality
 * Implements proper LCMComponent lifecycle management for pages
 */
export abstract class BasePage extends BaseComponent {
    protected themeManager: typeof ThemeManager;
    protected modal: Modal;
    protected toastManager: ToastManager

    protected themeToggleButton: HTMLButtonElement;
    protected themeToggleIcon: HTMLElement;

    // Constructor now just uses document as the rootElement
    constructor(public readonly componentId: string, eventBus: EventBus | null = null, public readonly debugMode: boolean = false) {
        // Mark as component in DOM for debugging
        super(componentId, document.body, eventBus, debugMode)

        this.initializeBaseComponents();
        
        // Bind base events first
        this.bindBaseEvents();
    }

    // LCMComponent Phase 1: Initialize page structure and discover child components
    public override performLocalInit(): LCMComponent[] {
        this.log('BasePage: Starting local initialization');
        
        // Then initialize page-specific components and discover children
        return this.initializeSpecificComponents();
    }
    
    // LCMComponent Phase 3: Activate the page (bind events after all components are ready)
    public override activate(): void {
        this.log('BasePage: Activating page');
        
        // Then bind page-specific events
        this.bindSpecificEvents();
        
        this.log('BasePage: Page activation complete');
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
        this.themeToggleIcon = document.getElementById('theme-toggle-icon')!;

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
    protected showModal(templateId: string, data?: any): void {
        this.modal?.show(templateId, data);
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
     * Should return any child components that need lifecycle management
     */
    protected initializeSpecificComponents(): LCMComponent[] { return []; }

    /**
     * Abstract method that subclasses must implement to bind their specific events
     */
    protected bindSpecificEvents(): void {}

    /**
     * Component-specific cleanup logic (required by BaseComponent)
     */
    protected destroyComponent(): void {
        this.log('BasePage: Cleaning up base page components');
        
        // Clean up base components
        this.modal = null as any;
        this.toastManager = null as any;
        this.themeToggleButton = null as any;
        this.themeToggleIcon = null as any;
    }
    
    /**
     * Ensure an element exists, create if missing
     * This is acceptable for page-level orchestration to find component root elements
     */
    protected ensureElement(selector: string, fallbackId: string): HTMLElement {
        let element = document.querySelector(selector) as HTMLElement;
        if (!element) {
            console.warn(`Element not found: ${selector}, creating fallback`);
            element = document.createElement('div');
            element.id = fallbackId;
            element.className = 'w-full h-full';
            // Fallback should be more specific than just body
            const mainContainer = document.querySelector('main') || document.body;
            mainContainer.appendChild(element);
        }
        
        // Ensure element has an ID for Phaser container
        if (!element.id) {
            element.id = fallbackId;
        }
        
        return element;
    }
}
