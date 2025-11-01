import { BasePage } from '../lib/BasePage';
import WeewarBundle from '../gen/wasmjs';
import { GamesServiceServiceClient } from '../gen/wasmjs/weewar/v1/gamesServiceClient';
import { GameViewerPageMethods, GameViewerPageServiceClient as GameViewerPageClient } from '../gen/wasmjs/weewar/v1/gameViewerPageClient';
import { GameViewPresenterServiceClient as  GameViewPresenterClient } from '../gen/wasmjs/weewar/v1/gameViewPresenterClient';
import { SingletonInitializerServiceServiceClient as SingletonInitializerServiceClient } from '../gen/wasmjs/weewar/v1/singletonInitializerServiceClient';
import { EventBus } from '../lib/EventBus';
import { PhaserGameScene } from './phaser/PhaserGameScene';
import { Unit, Tile, World } from './World';
import {
    GameState as ProtoGameState,
    SetGameStateRequest, SetGameStateResponse,
    SetContentRequest, SetContentResponse,
	  LogMessageRequest, LogMessageResponse,
    ShowHighlightsRequest, ShowHighlightsResponse,
    ClearHighlightsRequest, ClearHighlightsResponse,
    ShowPathRequest, ShowPathResponse,
    ClearPathsRequest, ClearPathsResponse,
    ShowBuildOptionsRequest, ShowBuildOptionsResponse,
    HighlightSpec,
    MoveUnitRequest, MoveUnitResponse,
    ShowAttackEffectRequest, ShowAttackEffectResponse,
    ShowHealEffectRequest, ShowHealEffectResponse,
    ShowCaptureEffectRequest, ShowCaptureEffectResponse,
    SetUnitAtRequest, SetUnitAtResponse,
    RemoveUnitAtRequest, RemoveUnitAtResponse,
} from '../gen/wasmjs/weewar/v1/interfaces';
import * as models from '../gen/wasmjs/weewar/v1/models';
import { create } from '@bufbuild/protobuf';
import { LCMComponent } from '../lib/LCMComponent';
import { LifecycleController } from '../lib/LifecycleController';
import { PLAYER_BG_COLORS } from './ColorsAndNames';
import { TerrainStatsPanel } from './TerrainStatsPanel';
import { UnitStatsPanel } from './UnitStatsPanel';
import { DamageDistributionPanel } from './DamageDistributionPanel';
import { GameLogPanel } from './GameLogPanel';
import { TurnOptionsPanel } from './TurnOptionsPanel';
import { BuildOptionsModal } from './BuildOptionsModal';
import { GameEventTypes, WorldEventTypes } from './events';
import { RulesTable, TerrainStats } from './RulesTable';
import { DockviewApi, DockviewComponent } from 'dockview-core';

/**
 * Game Viewer Page - Interactive game play interface
 * Responsible for:
 * - Loading world as a game instance
 * - Coordinating WASM game engine
 * - Managing game state and turn flow
 * - Handling player interactions (unit selection, movement, attacks)
 * - Providing game controls and UI feedback
 */
export class GameViewerPage extends BasePage implements LCMComponent, GameViewerPageMethods {
    private wasmBundle: WeewarBundle;
    private gamesServiceClient: GamesServiceServiceClient;
    private gameViewPresenterClient: GameViewPresenterClient;
    private singletonInitializerClient : SingletonInitializerServiceClient;
    private currentGameId: string | null;
    private gameScene: PhaserGameScene
    private world: World  // ✅ Shared World component
    private terrainStatsPanel: TerrainStatsPanel
    private unitStatsPanel: UnitStatsPanel
    private damageDistributionPanel: DamageDistributionPanel
    private gameLogPanel: GameLogPanel
    private turnOptionsPanel: TurnOptionsPanel
    private buildOptionsModal: BuildOptionsModal
    private rulesTable: RulesTable = new RulesTable();
    
    // Dockview interface
    private dockview: DockviewApi;
    private themeObserver: MutationObserver | null = null;
    
    // =============================================================================
    // LCMComponent Interface Implementation
    // =============================================================================

    /**
     * Phase 1: Initialize DOM and discover child components
     */
    async performLocalInit(): Promise<LCMComponent[]> {
        // Load game config first
        this.currentGameId = (document.getElementById("gameIdInput") as HTMLInputElement).value.trim()
        if (!this.currentGameId) {
          throw new Error("Game Id Not Found")
        }
        
        // Subscribe to events BEFORE creating components
        this.subscribeToGameStateEvents();
        
        // Create child components
        this.createComponents();

        this.updateGameStatusBanner('Game Loading...');

        await this.loadWASM() // kick off loading

        // Return child components for lifecycle management
        // Note: World and GameState don't extend BaseComponent, so not included in lifecycle
        return [
            this.gameScene,
            this.terrainStatsPanel,
            this.unitStatsPanel,
            this.damageDistributionPanel,
            this.gameLogPanel,
            this.buildOptionsModal,
        ]
    }

    /**
     * Phase 2: Inject dependencies
     */
    setupDependencies(): void {
        // Pass the theme to the stats panels
        const assetProvider = this.gameScene.getAssetProvider();
        if (assetProvider) {
            const theme = assetProvider.getTheme();
            if (this.terrainStatsPanel) {
                this.terrainStatsPanel.setTheme(theme);
            }
            if (this.unitStatsPanel) {
                this.unitStatsPanel.setTheme(theme);
            }
            if (this.damageDistributionPanel) {
                this.damageDistributionPanel.setTheme(theme);
            }
            if (this.turnOptionsPanel) {
                this.turnOptionsPanel.setTheme(theme);
            }
            if (this.buildOptionsModal) {
                this.buildOptionsModal.setTheme(theme);
            }
        }

        // Set presenter client on components so they can call presenter directly
        this.gameScene.gameViewPresenterClient = this.gameViewPresenterClient;
        this.turnOptionsPanel.gameViewPresenterClient = this.gameViewPresenterClient;
        this.buildOptionsModal.gameViewPresenterClient = this.gameViewPresenterClient;
    }

    /**
     * Phase 3: Activate component when all dependencies are ready
     */
    async activate(): Promise<void> {
        // Bind events now that all components are ready
        this.bindGameSpecificEvents();
        
        // Subscribe to path visualization events
        this.eventBus.addSubscription('show-path-visualization', null, this);
        this.eventBus.addSubscription('clear-path-visualization', null, this);
        
        // TODO _ this will be done by initialize WASM
        this.wasmBundle.registerBrowserService('GameViewerPage', this)

        // Initialize the presenter by setting it game data now that all UI components are ready
        await this.initializePresenter();

        // Expose gameScene to console for testing animations
        (window as any).gameScene = this.gameScene;
        console.log("🎮 gameScene exposed to window for animation testing");
        console.log("Try: gameScene.moveUnit(unit, path) or gameScene.showAttackEffect({q:0,r:0}, {q:1,r:0}, 10)");
    }
    
    
    /**
     * Show path visualization on the game scene
     */
    private showPathVisualization(coords: number[], color: number, thickness: number): void {
        if (!this.gameScene) return;
        
        // Get the movement highlight layer (or selection layer) to draw paths
        const movementLayer = this.gameScene.movementHighlightLayer;
        if (movementLayer) {
            // Clear any existing paths first
            movementLayer.clearAllPaths();
            // Add the new path
            movementLayer.addPath(coords, color, thickness);
        }
    }
    
    /**
     * Clear path visualization from the game scene
     */
    private clearPathVisualization(): void {
        if (!this.gameScene) return;
        
        // Clear paths from movement layer
        const movementLayer = this.gameScene.movementHighlightLayer;
        if (movementLayer) {
            movementLayer.clearAllPaths();
        }
    }

    /**
     * Subscribe to GameState events
     */
    private subscribeToGameStateEvents(): void {
        // GameViewer ready event - set up interaction callbacks and load world
        this.addSubscription(WorldEventTypes.WORLD_VIEWER_READY, this);
        
        // Game data ready event - WASM and game data loaded
        this.addSubscription(GameEventTypes.GAME_DATA_LOADED, this);
        
        // GameState notification events (for system coordination, not user interaction responses)
        this.addSubscription('unit-moved', this);
        this.addSubscription('unit-attacked', this);
        this.addSubscription('turn-ended', this);
    }

    /**
     * Handle events from the EventBus
     */
    public handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
        switch(eventType) {
            case 'show-path-visualization':
                this.showPathVisualization(data.coords, data.color, data.thickness);
                break;
                
            case 'clear-path-visualization':
                this.clearPathVisualization();
                break;
            
            case 'unit-moved':
                // Could trigger animations, sound effects, etc.
                break;
            
            case 'unit-attacked':
                // Could trigger combat animations, sound effects, etc.
                break;
            
            case 'turn-ended':
                // Could trigger end-of-turn animations, notifications, etc.
                break;
            
            default:
                // Call parent implementation for unhandled events
                super.handleBusEvent(eventType, data, target, emitter);
        }
    }

    /**
     * Create WorldViewer, World, and GameState component instances
     */
    private createComponents(): void {
        // ✅ Create shared World component first (subscribes first to server-changes)
        this.world = new World(this.eventBus, 'Game World');

        // Create BuildOptionsModal (separate from DockView)
        const modalElement = document.getElementById('build-options-modal');
        if (!modalElement) {
            throw new Error('GameViewerPage: build-options-modal element not found');
        }
        this.buildOptionsModal = new BuildOptionsModal(modalElement, this.eventBus, true);

        // Initialize DockView layout
        this.initializeDockView();
    }

    /**
     * Initialize DockView layout with game panels
     */
    private initializeDockView(): void {
        const container = document.getElementById('dockview-container');
        if (!container) {
            throw new Error('GameViewerPage: dockview-container not found');
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
     * Configure the default DockView layout for optimal game viewing
     */
    private configureDefaultGameLayout(): void {
        // Add main game panel (center)
        this.dockview.addPanel({
            id: 'main-game-panel',
            component: 'main-game',
            title: 'Game',
            position: { direction: 'right' }
        });

        // Add terrain stats panel (right side)
        this.dockview.addPanel({
            id: 'terrain-stats-panel', 
            component: 'terrain-stats',
            title: 'Terrain Info',
            position: { 
                direction: 'right',
                referencePanel: 'main-game-panel'
            }
        });

        // Add unit stats panel (below terrain stats panel)
        this.dockview.addPanel({
            id: 'unit-stats-panel',
            component: 'unit-stats',
            title: 'Unit Info',
            position: { 
                direction: 'below',
                referencePanel: 'terrain-stats-panel'
            }
        });

        // Add damage distribution panel (below unit stats panel)
        this.dockview.addPanel({
            id: 'damage-distribution-panel',
            component: 'damage-distribution',
            title: 'Damage Distribution',
            position: { 
                direction: 'below',
                referencePanel: 'unit-stats-panel'
            }
        });

        // Add turn options panel (below damage distribution panel)
        this.dockview.addPanel({
            id: 'turn-options-panel',
            component: 'turn-options',
            title: 'Turn Options',
            position: {
                direction: 'below',
                referencePanel: 'damage-distribution-panel'
            }
        });

        // Add game log panel (left side)
        this.dockview.addPanel({
            id: 'game-log-panel',
            component: 'game-log',
            title: 'Game Log',
            position: {
                direction: 'left',
                referencePanel: 'main-game-panel'
            }
        });

        // Set panel sizes for optimal viewing
        setTimeout(() => {
            this.dockview.getPanel('terrain-stats-panel')?.api.setSize({ width: 320 });
            this.dockview.getPanel('game-log-panel')?.api.setSize({ width: 280 });
        }, 100);
    }

    /**
     * Save the current DockView layout to localStorage
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
                // Find the Phaser container within the cloned template
                const phaserContainer = element.querySelector('#phaser-viewer-container') as HTMLElement;
                if (phaserContainer) {
                    // Create PhaserGameScene with the container
                    this.gameScene = new PhaserGameScene(phaserContainer, this.eventBus, true);
                }
            },
            dispose: () => {
                // PhaserGameScene cleanup will be handled by LCM lifecycle
                // Component disposal is managed by DockView
            },
            onDidResize: () => {
                // Handle panel resize events - resize the Phaser scene
                if (this.gameScene) {
                    // Get the current container size
                    const phaserContainer = element.querySelector('#phaser-viewer-container') as HTMLElement;
                    if (phaserContainer) {
                        const width = phaserContainer.clientWidth;
                        const height = phaserContainer.clientHeight;
                        
                        // Use the public resize method
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
        // Keep the template ID, don't create wrapper instance

        return {
            element,
            init: () => {
                // Create TerrainStatsPanel with the cloned element
                this.terrainStatsPanel = new TerrainStatsPanel(element, this.eventBus, true);
            },
            dispose: () => {
                // TerrainStatsPanel cleanup will be handled by LCM lifecycle
                // Component disposal is managed by DockView
            }
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
        // Keep the template ID, don't create wrapper instance

        return {
            element,
            init: () => {
                // Create UnitStatsPanel with the cloned element
                this.unitStatsPanel = new UnitStatsPanel(element, this.eventBus, true);
            },
            dispose: () => {
                // UnitStatsPanel cleanup will be handled by LCM lifecycle
                // Component disposal is managed by DockView
            }
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
                // Create TurnOptionsPanel with the cloned element
                this.turnOptionsPanel = new TurnOptionsPanel(element, this.eventBus, true);
            },
            dispose: () => {
                // TurnOptionsPanel cleanup will be handled by LCM lifecycle
                // Component disposal is managed by DockView
            }
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
                // Create DamageDistributionPanel with the cloned element
                this.damageDistributionPanel = new DamageDistributionPanel(element, this.eventBus, true);
            },
            dispose: () => {
                // DamageDistributionPanel cleanup will be handled by LCM lifecycle
                // Component disposal is managed by DockView
            }
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
        // Keep the template ID, don't create wrapper instance

        return {
            element,
            init: () => {
                // Create GameLogPanel with the cloned element
                this.gameLogPanel = new GameLogPanel(element, this.eventBus);
            },
            dispose: () => {
                // GameLogPanel cleanup will be handled by LCM lifecycle
                // Component disposal is managed by DockView
            }
        };
    }

    private async loadWASM(): Promise<void> {
        // Create base bundle with module configuration
        this.wasmBundle  = new WeewarBundle();
        this.gamesServiceClient = new GamesServiceServiceClient(this.wasmBundle);
        this.gameViewPresenterClient = new GameViewPresenterClient(this.wasmBundle);
        this.singletonInitializerClient = new SingletonInitializerServiceClient(this.wasmBundle);
        await this.wasmBundle.loadWasm('/static/wasm/weewar-cli.wasm');
        await this.wasmBundle.waitUntilReady()
    }

    /**
     * Initialize game using WASM game engine
     * This now handles both WASM loading and World creation in GameState
     */
    private async initializePresenter(): Promise<void> {
        // Get raw JSON data from page elements
        const gameElement = document.getElementById('game.data-json')!;
        const gameStateElement = document.getElementById('game-state-data-json')!;
        const historyElement = document.getElementById('game-history-data-json')!;
        
        if (!gameElement?.textContent || gameElement.textContent.trim() === 'null') {
            throw new Error('No game data found in page elements');
        }

        if (false) {
            // Convert JSON strings to Uint8Array for WASM
            const gameBytes = new TextEncoder().encode(gameElement.textContent);
            const gameStateBytes = new TextEncoder().encode(
                gameStateElement?.textContent && gameStateElement.textContent.trim() !== 'null'
                    ? gameStateElement.textContent
                    : '{}'
            );
            const historyBytes = new TextEncoder().encode(
                historyElement?.textContent && historyElement.textContent.trim() !== 'null'
                    ? historyElement.textContent
                    : '{"gameId":"","groups":[]}'
            );

            // Call WASM loadGameData function - check if it exists first
            const weewar = (window as any).weewar;
            if (!weewar || !weewar.loadGameData) {
                throw new Error('WASM loadGameData function not available. WASM module may not be fully loaded.');
            }

            const wasmResult = weewar.loadGameData(gameBytes, gameStateBytes, historyBytes);

            if (!wasmResult.success) {
                throw new Error(`WASM load failed: ${wasmResult.error}`);
            }

            // 3. Call presenter to initialize (ONE proto RPC call does everything!)
            const response = await this.gameViewPresenterClient.initializeGame({ gameId: this.currentGameId || "", });
            
            if (!response.success) {
                throw new Error(`WASM load failed: ${response.error}`);
            }
        } else {
            // 3. Call presenter to initialize (ONE proto RPC call does everything!)
            const response = await this.singletonInitializerClient.initializeSingleton({
                gameId: this.currentGameId || "",
                gameData: gameElement!.textContent,
                gameState: gameStateElement?.textContent || '{}',
                moveHistory: historyElement?.textContent || '{"gameId":"","groups":[]}',
            });
            
            if (!response.response!.success) {
                throw new Error(`WASM load failed: ${response.response!.error}`);
            }
        }
    }

    /**
     * Internal method to bind game-specific events (called from activate() phase)
     */
    private bindGameSpecificEvents(): void {
        // End Turn button
        const endTurnBtn = document.getElementById('end-turn-btn')!;
        endTurnBtn.addEventListener('click', () => {
          this.gameViewPresenterClient.endTurnButtonClicked({
              gameId: this.currentGameId || ""
          });
        });

        // Screenshot button
        const screenshotBtn = document.getElementById('capture-screenshot-btn');
        if (screenshotBtn) {
            screenshotBtn.addEventListener('click', () => this.handleScreenshotClick());
        }
    }

    /**
     * Handle Screenshot button click
     */
    private async handleScreenshotClick(): Promise<void> {
        if (!this.currentGameId) {
            console.error('No game ID available');
            this.showToast('Error', 'No game ID available', 'error');
            return;
        }

        try {
            // Capture screenshot from Phaser scene
            const blob = await this.gameScene.captureScreenshotAsync('image/png', 0.92);

            if (!blob) {
                this.showToast('Error', 'Failed to capture screenshot', 'error');
                return;
            }

            // Upload to server
            const formData = new FormData();
            formData.append('screenshot', blob, 'screenshot.png');

            const response = await fetch(`/games/${this.currentGameId}/screenshot`, {
                method: 'POST',
                body: formData
            });

            if (response.ok) {
                this.showToast('Success', 'Screenshot saved successfully', 'success');
            } else {
                this.showToast('Error', 'Failed to save screenshot', 'error');
            }
        } catch (error) {
            console.error('Screenshot error:', error);
            this.showToast('Error', 'Failed to capture or save screenshot', 'error');
        }
    }

    /**
     * Handle End Turn button click
     */
    private async handleEndTurnClick(): Promise<void> {
        if (!this.currentGameId) {
            console.error('No game ID available');
            return;
        }

        // Call presenter
        await this.gameViewPresenterClient.endTurnButtonClicked({
            gameId: this.currentGameId
        });
    }


    /**
     * Hide the loading overlay in the main game panel
     */
    private hideLoadingOverlay(): void {
        // Find the loading overlay in the main game panel instance
        const gameLoadingOverlay = document.querySelector('#main-game-panel-instance #game-loading') as HTMLElement;
        if (gameLoadingOverlay) {
            gameLoadingOverlay.style.display = 'none';
        }
    }

    /**
     * Resize the game canvas to fit the container properly
     */
    private resizeGameCanvas(): void {
        if (this.gameScene) {
            // Find the Phaser container in the main game panel
            const phaserContainer = document.querySelector('#main-game-panel-instance #phaser-viewer-container') as HTMLElement;
            if (phaserContainer) {
                // Force a resize to ensure the canvas fits the container
                const width = phaserContainer.clientWidth;
                const height = phaserContainer.clientHeight;
                
                // Add a small delay to ensure DOM has settled
                setTimeout(() => {
                    if (this.gameScene) {
                        this.gameScene.resize(width, height);

                        // Center camera on the world after loading
                        this.gameScene.centerCameraOnWorld();
                    }
                }, 100);
            }
        }
    }

    /**
     * UI update functions
     */
    private updateGameStatusBanner(status: string, currentPlayer?: number): void {
        const statusElement = document.getElementById('game-status');
        if (statusElement) {
            statusElement.textContent = status;

            // Use player-specific background color, fallback to green for general messages
            const playerColorClass = currentPlayer ? PLAYER_BG_COLORS[currentPlayer] : 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
            statusElement.className = `inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${playerColorClass}`;
        }
    }

    private updateGameUIFromState(gameState: ProtoGameState): void {
        // Update game status with player-specific color - use player ID directly
        this.updateGameStatusBanner(`Ready - Player ${gameState.currentPlayer}'s Turn`, gameState.currentPlayer);
        
        // Update turn counter
        this.updateTurnCounter(gameState.turnCounter);
    }
    
    private updateTurnCounter(turnCounter: number): void {
        const turnElement = document.getElementById('turn-counter');
        if (turnElement) {
            turnElement.textContent = `Turn ${turnCounter}`;
        }
    }

    // Presenter interface methods
  async setTurnOptionsContent(request: SetContentRequest) {
    console.log("setTurnOptionsContent called on the browser: ", request)
    this.turnOptionsPanel.innerHTML = request.innerHtml
    // Hydrate theme images and setup click handlers after Go template renders HTML
    await this.turnOptionsPanel.hydrateThemeImages()
    return {}
  }

  async showBuildOptions(request: ShowBuildOptionsRequest): Promise<ShowBuildOptionsResponse> {
    console.log("showBuildOptions called on the browser:", request);

    if (request.hide) {
      // Hide the modal
      this.buildOptionsModal.hide();
    } else {
      // Show the modal with the rendered content and tile coordinates
      await this.buildOptionsModal.show(request.innerHtml, request.q, request.r);
    }

    return {}
  }

	async setUnitStatsContent(request: SetContentRequest) {
    console.log("setUnitStatsContent called on the browser")
    this.unitStatsPanel.innerHTML = request.innerHtml
    // Hydrate theme images after Go template renders HTML
    await this.unitStatsPanel.hydrateThemeImages()
    return {}
  }

	async setDamageDistributionContent(request: SetContentRequest) {
    console.log("setDamageDistributionContent called on the browser")
    this.damageDistributionPanel.innerHTML = request.innerHtml
    await this.damageDistributionPanel.hydrateThemeImages()
    return {}
  }
	async setTerrainStatsContent(request: SetContentRequest) {
    console.log("setTerrainStatsContent called on the browser")
    this.terrainStatsPanel.innerHTML = request.innerHtml
    await this.terrainStatsPanel.hydrateThemeImages()
    return {}
  }
	// Visualization command methods - delegate to PhaserGameScene
  async showHighlights(request: ShowHighlightsRequest) {
    console.log("showHighlights called:", request);
    if (request.highlights) {
      this.gameScene.showHighlights(request.highlights);
    }
    return {}
  }

  async clearHighlights(request: ClearHighlightsRequest) {
    console.log("clearHighlights called:", request);
    this.gameScene.clearHighlights(request.types || []);
    return {}
  }

  async showPath(request: ShowPathRequest) {
    console.log("showPath called:", request);
    if (request.coords) {
      this.gameScene.showPath(request.coords, request.color, request.thickness);
    }
    return {}
  }

  async clearPaths(request: ClearPathsRequest) {
    console.log("clearPaths called:", request);
    this.gameScene.clearPaths();
    return {}
  }

  async logMessage(request: LogMessageRequest) {
    console.log("logMessage called on the browser")
    return {}
  }
	async setGameState(req: SetGameStateRequest) {
    console.log("setGameState called on the browser")
    const worldData = req.state!.worldData!
    const game = req.game!
    // Load data into shared World component
    this.world.loadTilesAndUnits(worldData.tiles || [], worldData.units || []);
    this.world.setName(game.name || 'Untitled Game');

    // Load world into viewer using shared World
    await this.gameScene.loadWorld(this.world);
    this.showToast('Success', `Game loaded: ${game.name || this.world.getName() || 'Untitled'}`, 'success');

    // Hide the loading overlay now that the game is loaded
    this.hideLoadingOverlay();

    // Ensure the game canvas is properly sized after loading
    this.resizeGameCanvas();

    // Update UI with loaded game state
    this.updateGameUIFromState(req.state!);
    this.gameLogPanel.logGameEvent(`Game loaded: ${req.state!.gameId}`, 'system');
    return {}
  }

  // Incremental update methods
  async setTileAt(request: { q: number, r: number, tile: Tile }) {
    console.log("setTileAt called on the browser:", request);
    this.world.setTileDirect(request.tile);
    return {}
  }

  async setUnitAt(request: SetUnitAtRequest): Promise<SetUnitAtResponse> {
    console.log("setUnitAt called on the browser:", request);
    if (request.unit) {
      await this.gameScene.setUnit(request.unit, { flash: request.flash, appear: request.appear });
      // Update world after animation completes
      this.world.setUnitDirect(request.unit);
    }
    return {}
  }

  async removeTileAt(request: { q: number, r: number }) {
    console.log("removeTileAt called on the browser:", request);
    this.world.removeTileAt(request.q, request.r);
    return {}
  }

  async removeUnitAt(request: RemoveUnitAtRequest): Promise<RemoveUnitAtResponse> {
    console.log("removeUnitAt called on the browser:", request);
    await this.gameScene.removeUnit(request.q, request.r, { animate: request.animate });
    // Update world after animation completes
    this.world.removeUnitAt(request.q, request.r);
    return {}
  }

  async moveUnit(request: MoveUnitRequest): Promise<MoveUnitResponse> {
    console.log("moveUnit called on the browser:", request);
    if (request.unit && request.path) {
      await this.gameScene.moveUnit(request.unit, request.path);
      // Update world after animation completes
      this.world.setUnitDirect(request.unit);
    }
    return {}
  }

  async showAttackEffect(request: ShowAttackEffectRequest): Promise<ShowAttackEffectResponse> {
    console.log("showAttackEffect called on the browser:", request);
    await this.gameScene.showAttackEffect(
      { q: request.fromQ, r: request.fromR },
      { q: request.toQ, r: request.toR },
      request.damage,
      request.splashTargets
    );
    return {}
  }

  async showHealEffect(request: ShowHealEffectRequest): Promise<ShowHealEffectResponse> {
    console.log("showHealEffect called on the browser:", request);
    await this.gameScene.showHealEffect(request.q, request.r, request.amount);
    return {}
  }

  async showCaptureEffect(request: ShowCaptureEffectRequest): Promise<ShowCaptureEffectResponse> {
    console.log("showCaptureEffect called on the browser:", request);
    await this.gameScene.showCaptureEffect(request.q, request.r);
    return {}
  }

  async updateGameStatus(request: { currentPlayer: number, turnCounter: number }) {
    console.log("updateGameStatus called on the browser:", request);
    // Update the game status banner
    this.updateGameStatusBanner(`Ready - Player ${request.currentPlayer}'s Turn`, request.currentPlayer);
    // Update turn counter
    this.updateTurnCounter(request.turnCounter);
    // Update End Turn button state (only enabled for Player 1 for now - TODO: make configurable)
    this.updateEndTurnButtonState(request.currentPlayer);
    return {}
  }

  /**
   * Update End Turn button enabled/disabled state based on current player
   */
  private updateEndTurnButtonState(currentPlayer: number): void {
    const endTurnBtn = document.getElementById('end-turn-btn') as HTMLButtonElement;
    if (endTurnBtn) {
      // TODO: Get the actual player ID from the game/user context
      // For now, assume we're playing as Player 1
      const isOurTurn = currentPlayer === 1;

      endTurnBtn.disabled = !isOurTurn;

      // Update visual state
      if (isOurTurn) {
        endTurnBtn.classList.remove('opacity-50', 'cursor-not-allowed');
        endTurnBtn.classList.add('hover:bg-green-700');
      } else {
        endTurnBtn.classList.add('opacity-50', 'cursor-not-allowed');
        endTurnBtn.classList.remove('hover:bg-green-700');
      }
    }
  }
}

// Initialize page when DOM is ready using LifecycleController
document.addEventListener('DOMContentLoaded', async () => {
    // Create page instance (just basic setup)
    const gameViewerPage = new GameViewerPage("GameViewerPage");
    
    // Make GameViewerPage available for e2e testing via command interface
    (window as any).gameViewerPage = gameViewerPage;
    
    // Create lifecycle controller with debug logging
    const lifecycleController = new LifecycleController(gameViewerPage.eventBus, LifecycleController.DefaultConfig);
    
    // Start breadth-first initialization
    await lifecycleController.initializeFromRoot(gameViewerPage);
});
