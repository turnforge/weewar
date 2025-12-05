import { GameViewerPageBase, PanelId } from './GameViewerPageBase';
import { LCMComponent } from '@panyam/tsappkit';
import { DockviewApi, DockviewComponent } from 'dockview-core';
import { TerrainStatsPanel } from './TerrainStatsPanel';
import { UnitStatsPanel } from './UnitStatsPanel';
import { DamageDistributionPanel } from './DamageDistributionPanel';
import { GameLogPanel } from './GameLogPanel';
import { TurnOptionsPanel } from './TurnOptionsPanel';
import { GameStatePanel } from './GameStatePanel';
import { PhaserGameScene } from './PhaserGameScene';

/**
 * GameViewerPage implementation using DockView for flexible panel layout.
 *
 * Features:
 * - Drag-and-drop panel arrangement
 * - Resizable panels
 * - Layout persistence (localStorage)
 * - Theme-aware styling
 * - Lazy panel creation (created when first added to dock)
 */
export class GameViewerPageDockView extends GameViewerPageBase {
    private dockview: DockviewApi;
    private themeObserver: MutationObserver | null = null;

    // Panel instances are created by DockView factories
    // We keep references here for lifecycle management

    /**
     * Initialize DockView layout system
     */
    protected async initializeLayout(): Promise<void> {
        const container = document.getElementById('dockview-container');
        if (!container) {
            throw new Error('GameViewerPageDockView: dockview-container not found');
        }

        // Apply theme class based on current theme
        const isDarkMode = document.documentElement.classList.contains('dark');
        container.className = isDarkMode ? 'dockview-theme-dark flex-1' : 'dockview-theme-light flex-1';

        // Listen for theme changes
        this.themeObserver = new MutationObserver((mutations) => {
            mutations.forEach((mutation) => {
                if (mutation.type === 'attributes' && mutation.attributeName === 'class') {
                    const isDarkMode = document.documentElement.classList.contains('dark');
                    container.className = isDarkMode ? 'dockview-theme-dark flex-1' : 'dockview-theme-light flex-1';
                }
            });
        });

        this.themeObserver.observe(document.documentElement, {
            attributes: true,
            attributeFilter: ['class']
        });

        const dockviewComponent = new DockviewComponent(container, {
            createComponent: (options: any) => {
                switch (options.name) {
                    case 'main-game':
                        return this.createMainGameComponent();
                    case 'terrain-stats':
                        return this.createTerrainStatsComponent();
                    case 'unit-stats':
                        return this.createUnitStatsComponent();
                    case 'damage-distribution':
                        return this.createDamageDistributionComponent();
                    case 'turn-options':
                        return this.createTurnOptionsComponent();
                    case 'game-log':
                        return this.createGameLogComponent();
                    case 'game-state':
                        return this.createGameStateComponent();
                    default:
                        return {
                            element: document.createElement('div'),
                            init: () => {},
                            dispose: () => {}
                        };
                }
            }
        });

        this.dockview = dockviewComponent.api;

        // Load saved layout or create default
        const savedLayout = this.loadDockviewLayout();
        if (savedLayout) {
            try {
                this.dockview.fromJSON(savedLayout);
            } catch (e) {
                console.warn('Failed to restore game viewer dockview layout, using default', e);
                this.configureDefaultGameLayout();
            }
        } else {
            this.configureDefaultGameLayout();
        }

        // Save layout on changes
        this.dockview.onDidLayoutChange(() => {
            this.saveDockviewLayout();
        });
    }

    /**
     * Create panels - for DockView, this returns empty since panels are created lazily
     */
    protected createPanels(): LCMComponent[] {
        // Panels are created lazily by DockView when added to the layout
        // They're stored in the base class fields (terrainStatsPanel, etc.)
        // and will be populated when createXComponent() methods are called

        // Return panels that have been created so far
        return [
            this.terrainStatsPanel,
            this.unitStatsPanel,
            this.damageDistributionPanel,
            this.gameLogPanel,
            this.turnOptionsPanel,
            this.gameStatePanel,
        ].filter(p => p != null);
    }

    /**
     * Get game scene container - in DockView, this is created dynamically
     * Uses PhaserSceneView with SceneId: "game-viewer-scene"
     */
    protected getGameSceneContainer(): HTMLElement {
        // Find the container within the DockView panel instance
        const container = document.querySelector('#main-game-panel-instance #game-viewer-scene') as HTMLElement;
        if (!container) {
            throw new Error('GameViewerPageDockView: game-viewer-scene not found in DockView panel');
        }
        return container;
    }

    /**
     * Game scene should be created late (when DockView panel is initialized)
     */
    protected shouldCreateGameSceneEarly(): boolean {
        return false; // DockView creates it when panel is added
    }

    /**
     * Called after game scene is created
     */
    protected onGameSceneCreated(): void {
        // No additional setup needed for DockView
    }

    /**
     * Show/focus a specific panel
     */
    protected showPanel(panelId: PanelId): void {
        // Map panelId to DockView panel ID
        const dockPanelId = this.mapPanelIdToDockId(panelId);
        const panel = this.dockview.getPanel(dockPanelId);
        if (panel) {
            panel.api.setActive();
        }
    }

    /**
     * Get panel DOM element
     */
    protected getPanelElement(panelId: PanelId): HTMLElement | null {
        const dockPanelId = this.mapPanelIdToDockId(panelId);
        const panel = this.dockview.getPanel(dockPanelId);
        return panel?.view.content?.element || null;
    }

    /**
     * Map generic panel ID to DockView-specific panel ID
     */
    private mapPanelIdToDockId(panelId: PanelId): string {
        const mapping: Record<PanelId, string> = {
            'terrain-stats': 'terrain-stats-panel',
            'unit-stats': 'unit-stats-panel',
            'damage-distribution': 'damage-distribution-panel',
            'turn-options': 'turn-options-panel',
            'game-log': 'game-log-panel',
            'game-state': 'game-state-panel',
            'build-options': '' // Not in DockView
        };
        return mapping[panelId] || '';
    }

    // =========================================================================
    // DockView Layout Configuration
    // =========================================================================

    /**
     * Configure the default DockView layout
     *
     * Layout:
     * - Left Column: Top = Players + Game Log tabs (Players selected), Bottom = Turn Options
     * - Center Column: Game scene
     * - Right Column: Top = Terrain Stats, Bottom = Unit Stats + Damage Distribution tabs (Unit Stats selected)
     */
    private configureDefaultGameLayout(): void {
        // === CENTER COLUMN: Main game panel ===
        this.dockview.addPanel({
            id: 'main-game-panel',
            component: 'main-game',
            title: 'Game',
            position: { direction: 'right' }
        });

        // === LEFT COLUMN ===
        // Top row: Players panel (will have Game Log as tab)
        this.dockview.addPanel({
            id: 'game-state-panel',
            component: 'game-state',
            title: 'Players',
            position: {
                direction: 'left',
                referencePanel: 'main-game-panel'
            }
        });

        // Add Game Log as tab within Players panel group (Players stays selected by default since it was added first)
        this.dockview.addPanel({
            id: 'game-log-panel',
            component: 'game-log',
            title: 'Game Log',
            position: {
                direction: 'within',
                referencePanel: 'game-state-panel'
            }
        });

        // Bottom row: Turn Options panel
        this.dockview.addPanel({
            id: 'turn-options-panel',
            component: 'turn-options',
            title: 'Turn Options',
            position: {
                direction: 'below',
                referencePanel: 'game-state-panel'
            }
        });

        // === RIGHT COLUMN ===
        // Top row: Terrain Stats panel
        this.dockview.addPanel({
            id: 'terrain-stats-panel',
            component: 'terrain-stats',
            title: 'Terrain Info',
            position: {
                direction: 'right',
                referencePanel: 'main-game-panel'
            }
        });

        // Bottom row: Unit Stats panel (will have Damage Distribution as tab)
        this.dockview.addPanel({
            id: 'unit-stats-panel',
            component: 'unit-stats',
            title: 'Unit Info',
            position: {
                direction: 'below',
                referencePanel: 'terrain-stats-panel'
            }
        });

        // Add Damage Distribution as tab within Unit Stats panel group (Unit Stats stays selected by default)
        this.dockview.addPanel({
            id: 'damage-distribution-panel',
            component: 'damage-distribution',
            title: 'Damage Distribution',
            position: {
                direction: 'within',
                referencePanel: 'unit-stats-panel'
            }
        });

        // Set panel sizes and ensure correct tabs are selected
        setTimeout(() => {
            this.dockview.getPanel('terrain-stats-panel')?.api.setSize({ width: 320 });
            this.dockview.getPanel('game-state-panel')?.api.setSize({ width: 280 });

            // Ensure Players tab is selected (not Game Log)
            this.dockview.getPanel('game-state-panel')?.api.setActive();
            // Ensure Unit Info tab is selected (not Damage Distribution)
            this.dockview.getPanel('unit-stats-panel')?.api.setActive();
        }, 100);
    }

    /**
     * Save DockView layout to localStorage
     */
    private saveDockviewLayout(): void {
        if (!this.dockview) return;

        const layout = this.dockview.toJSON();
        localStorage.setItem('game-viewer-dockview-layout', JSON.stringify(layout));
    }

    /**
     * Load saved DockView layout from localStorage
     */
    private loadDockviewLayout(): any {
        const saved = localStorage.getItem('game-viewer-dockview-layout');
        return saved ? JSON.parse(saved) : null;
    }

    // =========================================================================
    // DockView Component Factories
    // =========================================================================

    /**
     * Create main game (Phaser) component
     */
    private createMainGameComponent() {
        const template = document.getElementById('main-game-panel-template');
        if (!template) {
            throw new Error('main-game-panel-template not found');
        }

        const element = template.cloneNode(true) as HTMLElement;
        element.style.display = 'block';
        element.id = 'main-game-panel-instance';

        return {
            element,
            init: () => {
                // Find the Phaser container within the cloned template (PhaserSceneView with SceneId: "game-viewer-scene")
                const phaserContainer = element.querySelector('#game-viewer-scene') as HTMLElement;
                if (phaserContainer) {
                    // Create PhaserGameScene with the container
                    this.gameScene = new PhaserGameScene(phaserContainer, this.eventBus, true);
                    this.onGameSceneCreated();
                }
            },
            dispose: () => {
                // Cleanup handled by LCM lifecycle
            },
            onDidResize: () => {
                if (this.gameScene) {
                    const phaserContainer = element.querySelector('#game-viewer-scene') as HTMLElement;
                    if (phaserContainer) {
                        const width = phaserContainer.clientWidth;
                        const height = phaserContainer.clientHeight;
                        this.gameScene.resize(width, height);
                    }
                }
            }
        };
    }

    /**
     * Create terrain stats component
     */
    private createTerrainStatsComponent() {
        const template = document.getElementById('terrain-stats-panel-template');
        if (!template) {
            throw new Error('terrain-stats-panel-template not found');
        }

        const element = template.cloneNode(true) as HTMLElement;
        element.style.display = 'block';

        return {
            element,
            init: () => {
                this.terrainStatsPanel = new TerrainStatsPanel(element, this.eventBus, true);
            },
            dispose: () => {}
        };
    }

    /**
     * Create unit stats component
     */
    private createUnitStatsComponent() {
        const template = document.getElementById('unit-stats-panel-template');
        if (!template) {
            throw new Error('unit-stats-panel-template not found');
        }

        const element = template.cloneNode(true) as HTMLElement;
        element.style.display = 'block';

        return {
            element,
            init: () => {
                this.unitStatsPanel = new UnitStatsPanel(element, this.eventBus, true);
            },
            dispose: () => {}
        };
    }

    /**
     * Create turn options component
     */
    private createTurnOptionsComponent() {
        const template = document.getElementById('turn-options-panel-template');
        if (!template) {
            throw new Error('turn-options-panel-template not found');
        }

        const element = template.cloneNode(true) as HTMLElement;
        element.style.display = 'block';

        return {
            element,
            init: () => {
                this.turnOptionsPanel = new TurnOptionsPanel(element, this.eventBus, true);
            },
            dispose: () => {}
        };
    }

    /**
     * Create damage distribution component
     */
    private createDamageDistributionComponent() {
        const template = document.getElementById('damage-distribution-panel-template');
        if (!template) {
            throw new Error('damage-distribution-panel-template not found');
        }

        const element = template.cloneNode(true) as HTMLElement;
        element.style.display = 'block';

        return {
            element,
            init: () => {
                this.damageDistributionPanel = new DamageDistributionPanel(element, this.eventBus, true);
            },
            dispose: () => {}
        };
    }

    /**
     * Create game log component
     */
    private createGameLogComponent() {
        const template = document.getElementById('game-log-panel-template');
        if (!template) {
            throw new Error('game-log-panel-template not found');
        }

        const element = template.cloneNode(true) as HTMLElement;
        element.style.display = 'block';

        return {
            element,
            init: () => {
                this.gameLogPanel = new GameLogPanel(element, this.eventBus);
            },
            dispose: () => {}
        };
    }

    /**
     * Create game state panel component
     */
    private createGameStateComponent() {
        const template = document.getElementById('game-state-panel-template');
        if (!template) {
            throw new Error('game-state-panel-template not found');
        }

        const element = template.cloneNode(true) as HTMLElement;
        element.style.display = 'block';

        return {
            element,
            init: () => {
                this.gameStatePanel = new GameStatePanel(element, this.eventBus, true);
            },
            dispose: () => {}
        };
    }
}

GameViewerPageDockView.loadAfterPageLoaded("gameViewerpage", GameViewerPageDockView, "GameViewerPageDockView")
