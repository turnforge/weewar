import { ThemeManager } from './ThemeManager';
import { Modal } from './Modal';
import { ToastManager } from './ToastManager';
import { EventBus } from './EventBus';
import { BaseComponent } from './Component';
import { LCMComponent } from './LCMComponent';
import { SplashScreen } from '../lib/SplashScreen';
import { LifecycleController } from './LifecycleController';

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
    public override performLocalInit(): Promise<LCMComponent[]> | LCMComponent[] {
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

        // Initialize responsive header actions dropdown
        this.initializeHeaderActionsDropdown();
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
     * Initialize responsive header actions menu (drawer for mobile, dropdown for desktop)
     */
    protected initializeHeaderActionsDropdown(): void {
        const menuBtn = document.getElementById('header-actions-menu-btn');
        const dropdown = document.getElementById('header-actions-dropdown');
        const drawer = document.getElementById('header-actions-drawer');
        const sourceContainer = document.getElementById('header-buttons-source');

        if (!menuBtn || !sourceContainer) {
            return; // Elements don't exist on this page
        }

        // Tailwind md: breakpoint is 768px
        const MOBILE_BREAKPOINT = 768;
        const isMobile = () => window.innerWidth < MOBILE_BREAKPOINT;

        // Setup dropdown (desktop)
        if (dropdown) {
            const dropdownContent = dropdown.querySelector('#header-actions-dropdown-content') as HTMLElement;
            if (dropdownContent) {
                // Clone buttons from source into dropdown with dropdown styling
                const buttons = sourceContainer.querySelectorAll('.header-action-btn');
                buttons.forEach((button) => {
                    const clone = button.cloneNode(true) as HTMLElement;

                    // Convert to dropdown item styling
                    clone.classList.remove('px-4', 'py-2', 'rounded-md', 'shadow-sm', 'border', 'border-transparent', 'border-gray-300', 'dark:border-gray-600');
                    clone.classList.add('w-full', 'text-left', 'px-4', 'py-2', 'text-sm', 'hover:bg-gray-100', 'dark:hover:bg-gray-700', 'flex', 'items-center');

                    // Remove bg colors and use hover instead
                    clone.classList.remove('bg-blue-600', 'hover:bg-blue-700', 'bg-green-600', 'hover:bg-green-700', 'bg-white', 'dark:bg-gray-700', 'hover:bg-gray-50', 'dark:hover:bg-gray-600');
                    clone.classList.add('text-gray-700', 'dark:text-gray-200');

                    dropdownContent.appendChild(clone);
                });
            }
        }

        // Setup drawer (mobile)
        if (drawer) {
            const drawerContent = drawer.querySelector('#header-actions-drawer-content') as HTMLElement;
            const drawerContainer = drawer.querySelector('.header-drawer-container') as HTMLElement;
            const drawerBackdrop = drawer.querySelector('.header-drawer-backdrop') as HTMLElement;

            if (drawerContent) {
                // Clone buttons from source into drawer with drawer styling
                const buttons = sourceContainer.querySelectorAll('.header-action-btn');
                buttons.forEach((button) => {
                    const clone = button.cloneNode(true) as HTMLElement;

                    // Keep button styling but make full width
                    clone.classList.add('w-full', 'justify-center');

                    drawerContent.appendChild(clone);
                });
            }

            // Drawer open/close with animation
            const openDrawer = () => {
                drawer.classList.remove('hidden');
                // Trigger reflow to ensure animation plays
                void drawer.offsetHeight;
                requestAnimationFrame(() => {
                    drawerBackdrop.classList.remove('opacity-0');
                    drawerBackdrop.classList.add('bg-opacity-50');
                    drawerContainer.classList.remove('-translate-y-full');
                    drawerContainer.classList.add('translate-y-0');
                });
            };

            const closeDrawer = () => {
                drawerBackdrop.classList.remove('bg-opacity-50');
                drawerBackdrop.classList.add('opacity-0');
                drawerContainer.classList.remove('translate-y-0');
                drawerContainer.classList.add('-translate-y-full');

                // Wait for animation to complete before hiding
                setTimeout(() => {
                    drawer.classList.add('hidden');
                }, 300); // Match duration-300 in CSS
            };

            if (drawerBackdrop) {
                drawerBackdrop.addEventListener('click', closeDrawer);
            }

            // Store functions for use in toggle handler
            (drawer as any)._openDrawer = openDrawer;
            (drawer as any)._closeDrawer = closeDrawer;
        }

        // Toggle menu button handler
        menuBtn.addEventListener('click', (e) => {
            e.stopPropagation();

            if (isMobile()) {
                // Use drawer on mobile
                if (drawer) {
                    const isOpen = !drawer.classList.contains('hidden');
                    if (isOpen) {
                        (drawer as any)._closeDrawer();
                    } else {
                        (drawer as any)._openDrawer();
                    }
                }
            } else {
                // Use dropdown on desktop
                if (dropdown) {
                    dropdown.classList.toggle('hidden');
                }
            }
        });

        // Close handlers
        document.addEventListener('click', (e) => {
            // Close dropdown if open and clicked outside
            if (dropdown && !dropdown.classList.contains('hidden') && !dropdown.contains(e.target as Node)) {
                dropdown.classList.add('hidden');
            }
        });

        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                if (dropdown && !dropdown.classList.contains('hidden')) {
                    dropdown.classList.add('hidden');
                }
                if (drawer && !drawer.classList.contains('hidden')) {
                    (drawer as any)._closeDrawer();
                }
            }
        });
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

    protected dismissSplashScreen() {
        SplashScreen.dismiss();
    }

    static loadAfterPageLoaded<T>(pageName: string, PageClass: any, PageClassName: string) {
        // Initialize page when DOM is ready using LifecycleController
        document.addEventListener('DOMContentLoaded', async () => {
            // Create page instance (just basic setup)
            const page = new PageClass(PageClassName);
            
            // Make GameViewerPage available for e2e testing via command interface
            (window as any)[pageName] = page;
            
            // Create lifecycle controller with debug logging
            const lifecycleController = new LifecycleController(page.eventBus, LifecycleController.DefaultConfig);
            
            // Start breadth-first initialization
            await lifecycleController.initializeFromRoot(page);
        });
    }
}

