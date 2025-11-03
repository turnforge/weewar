
import { ThemeManager } from '../lib/ThemeManager'; // For theme consistency if needed later
import { Modal } from '../lib/Modal';
import { ToastManager } from '../lib/ToastManager';
import { SplashScreen } from '../lib/SplashScreen';

/**
 * Manages the listing page logic
 */
class HomePage {
    private modal: Modal;
    private toastManager: ToastManager;
    private createNewButton: HTMLButtonElement | null; // Add reference

    constructor() {
        ThemeManager.init(); // Initialize theme handling
        this.modal = Modal.getInstance();
        this.toastManager = ToastManager.getInstance();
        this.createNewButton = document.getElementById('create-new-design-btn') as HTMLButtonElement | null; // Get the button
        this.bindEvents(); // Bind events
    }

    /**
     * Binds event listeners for the page.
     */
    private bindEvents(): void {
        // Bind click on the "Create New" button
        if (this.createNewButton) {
            this.createNewButton.addEventListener('click', this.handleCreateNewClick.bind(this));
        }

        // Use event delegation for clicks within the modal
        // Listen on the modal container or a higher-level element
        document.body.addEventListener('click', this.handleModalClick.bind(this));
        SplashScreen.dismiss()
    }

    /**
     * Handles the click on the main "Create New Design" button.
     */
    private handleCreateNewClick(): void {
        this.modal.show('create-design-modal'); // Show the new modal
    }

    /**
     * Handles clicks *inside* the modal using event delegation.
     */
    private handleModalClick(event: MouseEvent): void {
        // Ensure the click originated from within our specific modal content
        const modalContent = this.modal.getContentElement();
        if (!modalContent || !modalContent.contains(event.target as Node)) {
            return; // Click was outside the modal content area
        }

        // Check if a template option card was clicked
        const cardButton = (event.target as HTMLElement).closest('.template-option-card');
        if (cardButton instanceof HTMLButtonElement) {
            event.preventDefault(); // Prevent default button behavior
            const action = cardButton.dataset.action;
            const templateId = cardButton.dataset.templateId; // Might be undefined for blank

            this.modal.hide(); // Hide modal before redirecting

            let redirectUrl = '/designs/new';
            if (action === 'create-from-template' && templateId) {
                // Add templateId as a query parameter
                redirectUrl += `?templateId=${encodeURIComponent(templateId)}`;
                console.log(`Redirecting to create from template: ${redirectUrl}`);
            } else {
                // Default to blank (no query parameter)
                console.log(`Redirecting to create blank: ${redirectUrl}`);
            }

            // Perform the redirect
            window.location.href = redirectUrl;
            return; // Handled
        }

        // Check if the modal's cancel button was clicked
        const cancelButton = (event.target as HTMLElement).closest('#create-design-cancel');
        if (cancelButton) {
            event.preventDefault();
            this.modal.hide();
            return; // Handled
        }
    }
}

// Initialize the HomePage when the DOM is fully loaded
document.addEventListener('DOMContentLoaded', () => {
    (window as any).Page = new HomePage();
});
