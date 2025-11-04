import { BaseComponent } from './Component';
import { LCMComponent } from './LCMComponent';
import { EventBus } from './EventBus';

/**
 * MobileBottomDrawer - Reusable bottom drawer component for mobile layouts
 *
 * Features:
 * - Slides up from bottom covering 60-70% of viewport
 * - Backdrop overlay that dims the content behind
 * - Auto-closes when backdrop is tapped
 * - Swipe down to close gesture
 * - Smooth slide-up/down animations
 * - Holds any panel content
 */
export class MobileBottomDrawer extends BaseComponent implements LCMComponent {
    private backdropElement: HTMLElement;
    private drawerElement: HTMLElement;
    private contentElement: HTMLElement;
    private closeButton: HTMLElement | null;
    private isOpen: boolean = false;
    private onCloseCallback?: () => void;

    // Swipe gesture tracking
    private touchStartY: number = 0;
    private touchCurrentY: number = 0;
    private isDragging: boolean = false;

    /**
     * Create a MobileBottomDrawer
     * @param rootElement - The root container element for the drawer
     * @param eventBus - Event bus for component communication
     * @param debugMode - Enable debug logging
     */
    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('mobile-bottom-drawer', rootElement, eventBus, debugMode);
    }

    /**
     * Phase 1: Initialize DOM and discover child components
     */
    async performLocalInit(): Promise<LCMComponent[]> {
        // Find drawer elements
        this.backdropElement = this.rootElement.querySelector('.drawer-backdrop') as HTMLElement;
        this.drawerElement = this.rootElement.querySelector('.drawer-container') as HTMLElement;
        this.contentElement = this.rootElement.querySelector('.drawer-content') as HTMLElement;
        this.closeButton = this.rootElement.querySelector('.drawer-close-btn') as HTMLElement;

        if (!this.backdropElement || !this.drawerElement || !this.contentElement) {
            throw new Error('MobileBottomDrawer: Required drawer elements not found');
        }

        // Bind event listeners
        this.bindEvents();

        return [];
    }

    /**
     * Bind event listeners for drawer interactions
     */
    private bindEvents(): void {
        // Close drawer on backdrop click
        this.backdropElement.addEventListener('click', (e) => {
            if (e.target === this.backdropElement) {
                this.close();
            }
        });

        // Close drawer on close button click
        if (this.closeButton) {
            this.closeButton.addEventListener('click', () => {
                this.close();
            });
        }

        // Close drawer on Escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.isOpen) {
                this.close();
            }
        });

        // Swipe-to-close gesture on drawer element
        this.drawerElement.addEventListener('touchstart', (e) => this.handleTouchStart(e), { passive: true });
        this.drawerElement.addEventListener('touchmove', (e) => this.handleTouchMove(e), { passive: false });
        this.drawerElement.addEventListener('touchend', (e) => this.handleTouchEnd(e), { passive: true });
    }

    /**
     * Handle touch start for swipe gesture
     */
    private handleTouchStart(e: TouchEvent): void {
        // Only track if touching the drawer itself (not scrollable content)
        const contentScrollTop = this.contentElement.scrollTop;

        // If content is scrolled down, let normal scrolling happen
        if (contentScrollTop > 0) {
            return;
        }

        this.touchStartY = e.touches[0].clientY;
        this.touchCurrentY = this.touchStartY;
        this.isDragging = true;
    }

    /**
     * Handle touch move for swipe gesture
     */
    private handleTouchMove(e: TouchEvent): void {
        if (!this.isDragging) return;

        this.touchCurrentY = e.touches[0].clientY;
        const deltaY = this.touchCurrentY - this.touchStartY;

        // Only allow downward swipes
        if (deltaY > 0) {
            // Prevent default scrolling when swiping down
            e.preventDefault();

            // Apply drag transform with resistance
            const dragAmount = Math.min(deltaY, 300); // Cap at 300px
            this.drawerElement.style.transform = `translateY(${dragAmount}px)`;

            // Reduce backdrop opacity based on drag amount
            const opacity = Math.max(0, 0.5 - (dragAmount / 600));
            this.backdropElement.style.background = `rgba(0, 0, 0, ${opacity})`;
        }
    }

    /**
     * Handle touch end for swipe gesture
     */
    private handleTouchEnd(e: TouchEvent): void {
        if (!this.isDragging) return;

        const deltaY = this.touchCurrentY - this.touchStartY;
        const velocity = deltaY / 300; // Rough velocity calculation

        // Close if swiped down more than 100px or with high velocity
        if (deltaY > 100 || velocity > 0.3) {
            this.close();
        } else {
            // Snap back to open position
            this.drawerElement.style.transform = '';
            this.backdropElement.style.background = '';
        }

        this.isDragging = false;
    }

    /**
     * Open the drawer with slide-up animation
     */
    public open(): void {
        if (this.isOpen) return;

        this.isOpen = true;

        // Reset any drag transform
        this.drawerElement.style.transform = '';
        this.backdropElement.style.background = '';

        // Add open class - CSS handles all animations
        this.rootElement.classList.add('open');

        // Emit event
        this.eventBus.emit('drawer-opened', { drawerId: this.componentId }, null, this);
    }

    /**
     * Close the drawer with slide-down animation
     */
    public close(): void {
        if (!this.isOpen) return;

        this.isOpen = false;

        // Reset any drag transform
        this.drawerElement.style.transform = '';
        this.backdropElement.style.background = '';

        // Remove open class - CSS handles animation
        this.rootElement.classList.remove('open');

        // Emit event
        this.eventBus.emit('drawer-closed', { drawerId: this.componentId }, null, this);

        // Call close callback if provided
        if (this.onCloseCallback) {
            this.onCloseCallback();
        }
    }

    /**
     * Toggle drawer open/closed
     */
    public toggle(): void {
        if (this.isOpen) {
            this.close();
        } else {
            this.open();
        }
    }

    /**
     * Check if drawer is currently open
     */
    public getIsOpen(): boolean {
        return this.isOpen;
    }

    /**
     * Set the content element for the drawer
     * @param element - The element to insert into the drawer content area
     */
    public setContent(element: HTMLElement): void {
        this.contentElement.innerHTML = '';
        this.contentElement.appendChild(element);
    }

    /**
     * Set a callback to be called when drawer closes
     * @param callback - Function to call on close
     */
    public setOnClose(callback: () => void): void {
        this.onCloseCallback = callback;
    }

    /**
     * Get the content container element
     */
    public getContentContainer(): HTMLElement {
        return this.contentElement;
    }
}
