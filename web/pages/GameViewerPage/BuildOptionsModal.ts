import { BaseComponent, EventBus, LCMComponent } from '@panyam/tsappkit';
import { GameViewPresenterClient as GameViewPresenterClient } from '../../gen/wasmjs/lilbattle/v1/services/gameViewPresenterClient';
import { ITheme } from '../../assets/themes/BaseTheme';
import { ThemeUtils } from '../common/ThemeUtils';

/**
 * BuildOptionsModal displays a modal dialog with available unit build options
 *
 * This component shows:
 * - List of buildable units for the selected tile
 * - Unit icons, names, classifications, and stats
 * - Build costs and current player coins
 * - Disabled state for units that cannot be afforded
 *
 * The modal is shown/hidden via the ShowBuildOptions RPC from the presenter.
 * When a build option is clicked, it calls back to the presenter with the selection.
 */
export class BuildOptionsModal extends BaseComponent implements LCMComponent {
    public gameViewPresenterClient: GameViewPresenterClient;
    private theme: ITheme | null = null;
    private modalOverlay: HTMLElement | null = null;
    private modalContent: HTMLElement | null = null;
    private modalBody: HTMLElement | null = null;
    private isProcessing: boolean = false;
    private currentQ: number = 0;  // Current tile Q coordinate
    private currentR: number = 0;  // Current tile R coordinate

    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('build-options-modal', rootElement, eventBus, debugMode);
    }

    /**
     * Set the theme for getting unit names and images
     */
    public setTheme(theme: ITheme): void {
        this.theme = theme;
    }

    /**
     * Initialize modal structure if needed
     */
    async performLocalInit(): Promise<LCMComponent[]> {
        // The root element IS the modal overlay
        this.modalOverlay = this.rootElement;
        this.modalContent = this.rootElement.querySelector('.modal-content');
        this.modalBody = this.rootElement.querySelector('.modal-body');

        if (!this.modalOverlay || !this.modalContent || !this.modalBody) {
            throw new Error('BuildOptionsModal: Required modal elements not found');
        }

        // Setup close handlers
        this.setupCloseHandlers();

        return [];
    }

    /**
     * Setup click handlers for closing the modal
     */
    private setupCloseHandlers(): void {
        // Close on overlay click (but not on modal content click)
        this.modalOverlay?.addEventListener('click', (e) => {
            if (e.target === this.modalOverlay && !this.isProcessing) {
                this.hide();
            }
        });

        // Close on Cancel button click
        const cancelBtn = this.rootElement.querySelector('.cancel-button');
        cancelBtn?.addEventListener('click', () => {
            if (!this.isProcessing) {
                this.hide();
            }
        });

        // Close on Escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.isVisible() && !this.isProcessing) {
                this.hide();
            }
        });
    }

    /**
     * Show the modal with build options content
     * Called by GameViewerPage in response to ShowBuildOptions RPC
     */
    public async show(innerHtml: string, q: number, r: number): Promise<void> {
        if (!this.modalBody || !this.modalOverlay) return;

        // Store the tile coordinates
        this.currentQ = q;
        this.currentR = r;

        // Set the content
        this.modalBody.innerHTML = innerHtml;

        // Hydrate theme images
        await ThemeUtils.hydrateThemeImages(this.modalBody, this.theme, this.debugMode);

        // Setup click handlers for build options
        this.setupBuildOptionClickHandlers();

        // Show the modal
        this.modalOverlay.classList.remove('hidden');
        this.modalOverlay.classList.add('flex');

        // Focus trap - focus first enabled button
        setTimeout(() => {
            const firstButton = this.modalBody?.querySelector('.build-option-button:not([disabled])') as HTMLElement;
            firstButton?.focus();
        }, 100);
    }

    /**
     * Hide the modal
     */
    public hide(): void {
        if (!this.modalOverlay) return;
        this.setProcessingState(false)
        this.modalOverlay.classList.add('hidden');
        this.modalOverlay.classList.remove('flex');
        this.isProcessing = false;
    }

    /**
     * Check if modal is currently visible
     */
    public isVisible(): boolean {
        return this.modalOverlay?.classList.contains('flex') ?? false;
    }

    /**
     * Setup click handlers for build option buttons
     */
    private setupBuildOptionClickHandlers(): void {
        const buttons = this.modalBody?.querySelectorAll('.build-option-button');
        buttons?.forEach(button => {
            button.addEventListener('click', async (e) => {
                // Prevent clicking if already processing or button is disabled
                if (this.isProcessing || (button as HTMLButtonElement).disabled) {
                    return;
                }

                const target = e.currentTarget as HTMLElement;
                const unitType = parseInt(target.getAttribute('data-unit-type') || '0');
                const cost = parseInt(target.getAttribute('data-cost') || '0');

                this.log(`Build option clicked: q=${this.currentQ}, r=${this.currentR}, unitType=${unitType}, cost=${cost}`);

                // Mark as processing and gray out the modal
                this.isProcessing = true;
                this.setProcessingState(true);

                // Call presenter to handle the build action
                try {
                    await this.gameViewPresenterClient.buildOptionClicked({
                        gameId: "",
                        pos: { label: "", q: this.currentQ, r: this.currentR, },
                        unitType: unitType,
                    });
                } catch (error) {
                    console.error('Build option clicked error:', error);
                    this.isProcessing = false;
                    this.setProcessingState(false);
                }

                // Note: Modal will be hidden by the presenter after build completes
            });
        });
    }

    /**
     * Set the modal to a processing state (grayed out, no interaction)
     */
    private setProcessingState(processing: boolean): void {
        if (!this.modalContent) return;

        if (processing) {
            this.modalContent.classList.add('opacity-50', 'pointer-events-none');

            // Add a loading spinner overlay
            const spinner = document.createElement('div');
            spinner.className = 'absolute inset-0 flex items-center justify-center bg-black/20 rounded-lg';
            spinner.innerHTML = `
                <div class="animate-spin rounded-full h-12 w-12 border-4 border-blue-500 border-t-transparent"></div>
            `;
            spinner.id = 'build-modal-spinner';
            this.modalContent.appendChild(spinner);
        } else {
            this.modalContent.classList.remove('opacity-50', 'pointer-events-none');

            // Remove spinner
            const spinner = this.modalContent.querySelector('#build-modal-spinner');
            spinner?.remove();
        }
    }
}
