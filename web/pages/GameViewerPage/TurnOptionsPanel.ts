import { GameViewPresenterClient as  GameViewPresenterClient } from '../../gen/wasmjs/lilbattle/v1/services/gameViewPresenterClient';
import { BaseComponent, EventBus, LCMComponent } from '@panyam/tsappkit';
import { World } from '../common/World';
import { ITheme } from '../../assets/themes/BaseTheme';
import { ThemeUtils } from '../common/ThemeUtils';

/**
 * TurnOptionsPanel displays available turn options at a selected position
 * 
 * This component shows:
 * - Available movement options with paths and costs
 * - Attack options with damage estimates
 * - End turn option when available
 * - Build/capture options when applicable
 * 
 * Similar to the CLI's "options" command, this provides a clear view
 * of all available actions at the current position.
 */
export class TurnOptionsPanel extends BaseComponent implements LCMComponent {
    public gameViewPresenterClient: GameViewPresenterClient;
    private world: World | null = null;
    private theme: ITheme | null = null;

    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('turn-options-panel', rootElement, eventBus, debugMode);
    }
    
    /**
     * Set the World dependency
     */
    public setWorld(world: World): void {
        this.world = world;
        this.log('World dependency set');
    }

    /**
     * Set the theme for getting unit names and images
     */
    public setTheme(theme: ITheme): void {
        this.theme = theme;
    }

    /**
     * Hydrate theme images after Go template renders HTML
     * Call this after the HTML content is injected by the Go backend
     */
    public async hydrateThemeImages(): Promise<void> {
        await ThemeUtils.hydrateThemeImages(this.rootElement, this.theme, this.debugMode);
        this.setupOptionClickHandlers();
    }

    /**
     * Setup click handlers for option buttons
     */
    private setupOptionClickHandlers(): void {
        const buttons = this.rootElement.querySelectorAll('.turn-option-button');
        buttons.forEach(button => {
            button.addEventListener('click', (e) => {
                const target = e.currentTarget as HTMLElement;
                const optionIndex = parseInt(target.getAttribute('data-option-index') || '-1');
                const optionType = target.getAttribute('data-option-type');
                const q = parseInt(target.getAttribute('data-q') || '0');
                const r = parseInt(target.getAttribute('data-r') || '0');

                this.log(`Option clicked: type=${optionType}, index=${optionIndex}, position=(${q},${r})`);

                // Call presenter directly
                this.gameViewPresenterClient.turnOptionClicked({
                    gameId: "",
                    optionIndex: optionIndex,
                    optionType: optionType || "",
                    pos: { label: "", q: q, r: r, }
                });
            });
        });
    }

    htmlUpdated(html: string) {
        this.hydrateThemeImages()
    }
}
