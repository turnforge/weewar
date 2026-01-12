import { GameViewerPageBase, PanelId } from './GameViewerPageBase';
import { LCMComponent, MobileBottomDrawer } from '@panyam/tsappkit';
import { CompactSummaryCard } from './CompactSummaryCard';
import { TerrainStatsPanel } from './TerrainStatsPanel';
import { UnitStatsPanel } from './UnitStatsPanel';
import { DamageDistributionPanel } from './DamageDistributionPanel';
import { GameLogPanel } from './GameLogPanel';
import { TurnOptionsPanel } from './TurnOptionsPanel';
import { GameStatePanel } from './GameStatePanel';
import { PhaserGameScene } from './PhaserGameScene';
import { SetContentRequest, SetContentResponse, SetAllowedPanelsRequest, SetAllowedPanelsResponse } from '../../gen/wasmjs/lilbattle/v1/models/interfaces';

/**
 * Context-aware button ordering configuration
 */
interface ButtonOrderingConfig {
    unitSelected: PanelId[];
    tileSelected: PanelId[];
    nothingSelected: PanelId[];
}

/**
 * Default button ordering for different selection contexts
 */
const DEFAULT_BUTTON_ORDERING: ButtonOrderingConfig = {
    // Unit selected: prioritize unit info, actions, combat
    unitSelected: ['unit-stats', 'turn-options', 'damage-distribution', 'terrain-stats', 'game-state', 'game-log'],

    // Tile selected: prioritize terrain info, then unit/actions
    tileSelected: ['terrain-stats', 'turn-options', 'unit-stats', 'damage-distribution', 'game-state', 'game-log'],

    // Nothing selected: prioritize game log and general info
    nothingSelected: ['game-state', 'game-log', 'turn-options', 'terrain-stats', 'unit-stats', 'damage-distribution']
};

/**
 * Button metadata for rendering
 */
interface ButtonMetadata {
    id: PanelId;
    icon: string;
    label: string;
}

const BUTTON_METADATA: Record<PanelId, ButtonMetadata> = {
    'unit-stats': { id: 'unit-stats', icon: '🪖', label: 'Unit' },
    'terrain-stats': { id: 'terrain-stats', icon: '🗺️', label: 'Terrain' },
    'damage-distribution': { id: 'damage-distribution', icon: '⚔️', label: 'Damage' },
    'turn-options': { id: 'turn-options', icon: '🎯', label: 'Actions' },
    'game-log': { id: 'game-log', icon: '📜', label: 'Log' },
    'game-state': { id: 'game-state', icon: '👥', label: 'Players' },
    'build-options': { id: 'build-options', icon: '🏗️', label: 'Build' }
};

/**
 * GameViewerPage implementation for mobile devices with bottom drawers
 *
 * Features:
 * - Bottom action bar with context-aware button ordering
 * - Bottom drawers (60-70% height) for each panel
 * - Compact summary card at top showing terrain+unit info
 * - Only one drawer open at a time
 * - Auto-close on backdrop tap
 * - Touch-optimized interactions
 */
export class GameViewerPageMobile extends GameViewerPageBase {
    // Drawers for each panel
    private drawers: Map<PanelId, MobileBottomDrawer> = new Map();
    private currentOpenDrawer: PanelId | null = null;

    // Compact summary card
    private compactSummaryCard: CompactSummaryCard | null = null;

    // Bottom action bar
    private bottomBarElement: HTMLElement | null = null;
    private buttonOrdering: ButtonOrderingConfig = DEFAULT_BUTTON_ORDERING;
    private currentContext: 'unitSelected' | 'tileSelected' | 'nothingSelected' = 'nothingSelected';
    private allowedPanels: PanelId[] = ['game-log']; // Start with just game log, presenter will set the full list

    /**
     * Initialize mobile layout with drawers and bottom bar
     */
    protected async initializeLayout(): Promise<void> {
        // Verify required containers exist
        const requiredIds = [
            'mobile-compact-summary-card',
            'mobile-game-scene-container',
            'mobile-bottom-bar',
            'mobile-drawer-unit-stats',
            'mobile-drawer-terrain-stats',
            'mobile-drawer-damage-distribution',
            'mobile-drawer-turn-options',
            'mobile-drawer-game-log',
            'mobile-drawer-game-state'
        ];

        for (const id of requiredIds) {
            const element = document.getElementById(id);
            if (!element) {
                throw new Error(`GameViewerPageMobile: Required element '${id}' not found in template`);
            }
        }

        // Get bottom bar element
        this.bottomBarElement = document.getElementById('mobile-bottom-bar');

        // Initialize compact summary card
        const cardElement = document.getElementById('mobile-compact-summary-card');
        if (cardElement) {
            this.compactSummaryCard = new CompactSummaryCard(cardElement, this.eventBus, true);
        }

        // Initialize drawers
        await this.initializeDrawers();

        // Render initial button ordering
        this.renderBottomBar();
    }

    /**
     * Initialize all drawer components
     */
    private async initializeDrawers(): Promise<void> {
        const drawerIds: PanelId[] = [
            'unit-stats',
            'terrain-stats',
            'damage-distribution',
            'turn-options',
            'game-log',
            'game-state'
        ];

        for (const panelId of drawerIds) {
            const drawerElement = document.getElementById(`mobile-drawer-${panelId}`);
            if (drawerElement) {
                const drawer = new MobileBottomDrawer(drawerElement, this.eventBus, true);

                // Set close callback to track open drawer and update button highlights
                drawer.setOnClose(() => {
                    if (this.currentOpenDrawer === panelId) {
                        this.currentOpenDrawer = null;
                        this.updateButtonHighlights();
                    }
                });

                this.drawers.set(panelId, drawer);
            }
        }
    }

    /**
     * Create all panel instances and attach to drawer containers
     */
    protected createPanels(): LCMComponent[] {
        const panels: LCMComponent[] = [];

        // Compact summary card
        if (this.compactSummaryCard) {
            panels.push(this.compactSummaryCard);
        }

        // Unit Stats Panel
        const unitStatsContainer = document.querySelector('#mobile-drawer-unit-stats .drawer-content') as HTMLElement;
        if (unitStatsContainer) {
            this.unitStatsPanel = new UnitStatsPanel(unitStatsContainer, this.eventBus, true);
            panels.push(this.unitStatsPanel);
        }

        // Terrain Stats Panel
        const terrainStatsContainer = document.querySelector('#mobile-drawer-terrain-stats .drawer-content') as HTMLElement;
        if (terrainStatsContainer) {
            this.terrainStatsPanel = new TerrainStatsPanel(terrainStatsContainer, this.eventBus, true);
            panels.push(this.terrainStatsPanel);
        }

        // Damage Distribution Panel
        const damageContainer = document.querySelector('#mobile-drawer-damage-distribution .drawer-content') as HTMLElement;
        if (damageContainer) {
            this.damageDistributionPanel = new DamageDistributionPanel(damageContainer, this.eventBus, true);
            panels.push(this.damageDistributionPanel);
        }

        // Turn Options Panel
        const turnOptionsContainer = document.querySelector('#mobile-drawer-turn-options .drawer-content') as HTMLElement;
        if (turnOptionsContainer) {
            this.turnOptionsPanel = new TurnOptionsPanel(turnOptionsContainer, this.eventBus, true);
            panels.push(this.turnOptionsPanel);
        }

        // Game Log Panel
        const gameLogContainer = document.querySelector('#mobile-drawer-game-log .drawer-content') as HTMLElement;
        if (gameLogContainer) {
            this.gameLogPanel = new GameLogPanel(gameLogContainer, this.eventBus);
            panels.push(this.gameLogPanel);
        }

        // Game State Panel
        const gameStateContainer = document.querySelector('#mobile-drawer-game-state .drawer-content') as HTMLElement;
        if (gameStateContainer) {
            this.gameStatePanel = new GameStatePanel(gameStateContainer, this.eventBus, true);
            panels.push(this.gameStatePanel);
        }

        // Add drawers to lifecycle management
        for (const drawer of this.drawers.values()) {
            panels.push(drawer);
        }

        return panels;
    }

    /**
     * Get game scene container (pre-exists in template)
     */
    protected getGameSceneContainer(): HTMLElement {
        const container = document.getElementById('mobile-game-scene-container');
        if (!container) {
            throw new Error('GameViewerPageMobile: mobile-game-scene-container not found');
        }
        return container;
    }

    /**
     * Game scene should be created early (map is primary on mobile)
     */
    protected shouldCreateGameSceneEarly(): boolean {
        return true;
    }

    /**
     * Called after game scene is created
     */
    protected onGameSceneCreated(): void {
        // No event subscriptions needed - presenter drives all updates via RPC calls
    }

    /**
     * Override to set theme on compact summary card
     */
    setupDependencies(): void {
        super.setupDependencies();

        // Set theme on compact summary card
        const assetProvider = this.gameScene?.getAssetProvider();
        if (assetProvider && this.compactSummaryCard) {
            const theme = assetProvider.getTheme();
            this.compactSummaryCard.setTheme(theme);
        }
    }

    /**
     * Show/focus a specific panel (opens its drawer)
     */
    protected showPanel(panelId: PanelId): void {
        // If same drawer is already open, close it (toggle behavior)
        if (this.currentOpenDrawer === panelId) {
            const drawer = this.drawers.get(panelId);
            if (drawer) {
                drawer.close();
                this.currentOpenDrawer = null;
                this.updateButtonHighlights();
            }
            return;
        }

        // Close current drawer if different
        if (this.currentOpenDrawer && this.currentOpenDrawer !== panelId) {
            const currentDrawer = this.drawers.get(this.currentOpenDrawer);
            if (currentDrawer) {
                currentDrawer.close();
            }
        }

        // Open requested drawer
        const drawer = this.drawers.get(panelId);
        if (drawer) {
            drawer.open();
            this.currentOpenDrawer = panelId;
            this.updateButtonHighlights();
        }
    }

    /**
     * Update button highlights based on which drawer is open
     */
    private updateButtonHighlights(): void {
        if (!this.bottomBarElement) return;

        // Remove active class from all buttons
        this.bottomBarElement.querySelectorAll('.bottom-bar-button').forEach(button => {
            button.classList.remove('active');
        });

        // Add active class to button for currently open drawer
        if (this.currentOpenDrawer) {
            const activeButton = this.bottomBarElement.querySelector(
                `[data-panel-id="${this.currentOpenDrawer}"]`
            );
            if (activeButton) {
                activeButton.classList.add('active');
            }
        }
    }

    /**
     * Get panel DOM element
     */
    protected getPanelElement(panelId: PanelId): HTMLElement | null {
        const drawer = this.drawers.get(panelId);
        return drawer ? drawer.getContentContainer() : null;
    }

    /**
     * Render the bottom action bar with current button ordering
     */
    private renderBottomBar(): void {
        if (!this.bottomBarElement) return;

        // Get button order for current context, filtered by allowed panels
        const orderedPanelIds = this.buttonOrdering[this.currentContext].filter(
            panelId => this.allowedPanels.includes(panelId)
        );

        console.log('[GameViewerPageMobile] renderBottomBar:', {
            currentContext: this.currentContext,
            allowedPanels: this.allowedPanels,
            orderedPanelIds: orderedPanelIds
        });

        // Build HTML for buttons
        const buttonsHtml = orderedPanelIds.map(panelId => {
            const metadata = BUTTON_METADATA[panelId];
            if (!metadata) {
                console.error(`[GameViewerPageMobile] No metadata for panel: ${panelId}`);
                return '';
            }
            return `
                <button
                    data-panel-id="${metadata.id}"
                    class="bottom-bar-button flex flex-col items-center justify-center px-4 py-2 text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 transition-colors">
                    <span class="text-2xl">${metadata.icon}</span>
                    <span class="text-xs mt-1">${metadata.label}</span>
                </button>
            `;
        }).join('');

        this.bottomBarElement.innerHTML = buttonsHtml;

        // Bind click handlers
        this.bottomBarElement.querySelectorAll('.bottom-bar-button').forEach(button => {
            button.addEventListener('click', () => {
                const panelId = button.getAttribute('data-panel-id') as PanelId;
                this.showPanel(panelId);
            });
        });
    }


    /**
     * Override setCompactSummaryCard to show content in compact card
     */
    async setCompactSummaryCard(request: SetContentRequest): Promise<SetContentResponse> {
        console.log("setCompactSummaryCard called on the browser:", request);

        if (this.compactSummaryCard) {
            this.compactSummaryCard.innerHTML = request.innerHtml;
        }

        return {};
    }

    /**
     * Override setAllowedPanels to update bottom bar buttons
     */
    async setAllowedPanels(request: SetAllowedPanelsRequest): Promise<SetAllowedPanelsResponse> {
        console.log("setAllowedPanels called on mobile:", request.panelIds);

        // Update allowed panels list
        this.allowedPanels = request.panelIds as PanelId[];

        // Infer context from allowed panels (since events aren't emitted by GameScene)
        if (this.allowedPanels.includes('unit-stats') || this.allowedPanels.includes('damage-distribution')) {
            this.currentContext = 'unitSelected';
        } else if (this.allowedPanels.includes('terrain-stats')) {
            this.currentContext = 'tileSelected';
        } else {
            this.currentContext = 'nothingSelected';
        }

        // Re-render bottom bar with new panel set and context
        this.renderBottomBar();
        this.updateButtonHighlights();

        return {};
    }
}

// Register mobile page variant
GameViewerPageMobile.loadAfterPageLoaded("gameViewerpage", GameViewerPageMobile, "GameViewerPageMobile");
