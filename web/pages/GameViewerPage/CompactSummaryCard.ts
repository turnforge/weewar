import { LCMComponent, BaseComponent, EventBus } from '@panyam/tsappkit';
import { ITheme } from '../../assets/themes/BaseTheme';
import { ThemeUtils } from '../common/ThemeUtils';

/**
 * CompactSummaryCard - Shows terrain and unit info in a compact horizontal bar
 *
 * Displays at the top of mobile layout, below header:
 * - Terrain info: type, defense bonus, movement cost
 * - Unit info (if present): type, health, movement, attack range
 *
 * Content is rendered server-side by Go presenter and injected via setCompactSummaryCard()
 */
export class CompactSummaryCard extends BaseComponent implements LCMComponent {
    private cardElement: HTMLElement;
    private theme: ITheme | null = null;

    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('compact-summary-card', rootElement, eventBus, debugMode);
    }

    async performLocalInit(): Promise<LCMComponent[]> {
        this.cardElement = this.rootElement;

        // Initially hidden
        this.hide();

        return [];
    }

    /**
     * Set the HTML content for the card (called by presenter)
     */
    public set innerHTML(html: string) {
        this.cardElement.innerHTML = html;

        // Auto-show if content is set
        if (html && html.trim() !== '') {
            this.show();
            // Hydrate theme images asynchronously
            this.hydrateThemeImages().catch(err =>
                console.error('[CompactSummaryCard] Error hydrating theme images:', err)
            );
        } else {
            this.hide();
        }
    }

    /**
     * Hydrate theme images in the card content
     */
    public async hydrateThemeImages(): Promise<void> {
        if (this.theme) {
            await ThemeUtils.hydrateThemeImages(this.cardElement, this.theme, this.debugMode);
        }
    }

    /**
     * Set the theme for image hydration
     */
    public setTheme(theme: ITheme): void {
        this.theme = theme;
    }

    /**
     * Hide the card
     */
    public hide(): void {
        this.cardElement.classList.add('hidden');
    }

    /**
     * Show the card
     */
    private show(): void {
        this.cardElement.classList.remove('hidden');
    }

    /**
     * Clear the card content and hide it
     */
    public clear(): void {
        this.cardElement.innerHTML = '';
        this.hide();
    }
}
