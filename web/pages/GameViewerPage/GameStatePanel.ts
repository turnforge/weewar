import { BaseComponent, EventBus, LCMComponent } from '@panyam/tsappkit';
import { ITheme } from '../../assets/themes/BaseTheme';
import { ThemeUtils } from '../common/ThemeUtils';

/**
 * Callback for when a player clicks "Join" on an open player slot
 */
export type JoinGameCallback = (playerId: number) => void;

/**
 * GameStatePanel displays game state information including:
 * - Players list with bases, units, and coins
 * - Current player indicator
 * - Round/turn information
 * - Current player's income breakdown
 *
 * The panel content is rendered by the Go presenter using Templar templates
 * and injected via the setGameStatePanelContent RPC call.
 */
export class GameStatePanel extends BaseComponent implements LCMComponent {
    private isUIBound = false;
    private isActivated = false;
    private theme: ITheme | null = null;
    private onJoinGame: JoinGameCallback | null = null;

    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('game-state-panel', rootElement, eventBus, debugMode);
    }

    /**
     * Set the callback to be called when user clicks a Join button
     */
    public setJoinGameCallback(callback: JoinGameCallback): void {
        this.onJoinGame = callback;
    }

    // LCMComponent Phase 1: Initialize DOM structure
    public performLocalInit(): LCMComponent[] {
        if (this.isUIBound) {
            this.log('Already bound to DOM, skipping');
            return [];
        }

        this.log('Binding GameStatePanel to DOM');
        this.isUIBound = true;
        this.log('GameStatePanel bound to DOM successfully');

        // This is a leaf component - no children
        return [];
    }

    // Phase 2: No external dependencies needed
    public setupDependencies(): void {
        this.log('GameStatePanel: No dependencies required');
    }

    // Phase 3: Activate component
    public activate(): void {
        if (this.isActivated) {
            this.log('Already activated, skipping');
            return;
        }

        this.log('Activating GameStatePanel');
        this.isActivated = true;
        this.log('GameStatePanel activated successfully');
    }

    // Phase 4: Deactivate component
    public deactivate(): void {
        this.log('Deactivating GameStatePanel');
        this.isActivated = false;
        this.log('GameStatePanel deactivated');
    }

    /**
     * Set the theme for styling
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
    }

    protected destroyComponent(): void {
        this.deactivate();
    }

    htmlUpdated(html: string) {
        this.hydrateThemeImages();
        this.bindJoinButtonEvents();
    }

    /**
     * Bind click events to Join buttons
     */
    private bindJoinButtonEvents(): void {
        const joinButtons = this.rootElement.querySelectorAll('.join-game-btn');
        joinButtons.forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.preventDefault();
                const playerId = parseInt((btn as HTMLElement).dataset.playerId || '0', 10);
                if (playerId > 0 && this.onJoinGame) {
                    this.onJoinGame(playerId);
                }
            });
        });
    }
}
