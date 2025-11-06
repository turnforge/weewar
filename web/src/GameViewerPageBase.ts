import { BasePage } from '../lib/BasePage';
import WeewarBundle from '../gen/wasmjs';
import { GamesServiceClient } from '../gen/wasmjs/weewar/v1/gamesServiceClient';
import { GameViewerPageMethods, GameViewerPageClient as GameViewerPageClient } from '../gen/wasmjs/weewar/v1/gameViewerPageClient';
import { GameViewPresenterClient as GameViewPresenterClient } from '../gen/wasmjs/weewar/v1/gameViewPresenterClient';
import { SingletonInitializerServiceClient as SingletonInitializerClient } from '../gen/wasmjs/weewar/v1/singletonInitializerServiceClient';
import { EventBus } from '../lib/EventBus';
import { AssetThemePreference } from './AssetThemePreference';
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
    SetAllowedPanelsRequest, SetAllowedPanelsResponse,
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

/**
 * Panel type identifiers
 */
export type PanelId = 'terrain-stats' | 'unit-stats' | 'damage-distribution' | 'turn-options' | 'game-log' | 'build-options';

/**
 * Abstract base class for Game Viewer Page implementations.
 *
 * This class contains all the core game logic (WASM, presenter, panels, events)
 * but delegates layout-specific concerns to child classes.
 *
 * Child classes must implement:
 * - Layout initialization (DockView/Grid/Mobile)
 * - Panel container location and management
 * - Game scene creation timing and placement
 */
export abstract class GameViewerPageBase extends BasePage implements LCMComponent, GameViewerPageMethods {
    // =========================================================================
    // Protected Fields - Available to child classes
    // =========================================================================
    protected wasmBundle: WeewarBundle;
    protected gamesClient: GamesServiceClient;
    protected gameViewPresenterClient: GameViewPresenterClient;
    protected singletonInitializerClient: SingletonInitializerClient;
    protected currentGameId: string | null;

    // Core game components
    protected gameScene: PhaserGameScene;
    protected world: World;
    protected rulesTable: RulesTable = new RulesTable();

    // UI Panels
    protected terrainStatsPanel: TerrainStatsPanel;
    protected unitStatsPanel: UnitStatsPanel;
    protected damageDistributionPanel: DamageDistributionPanel;
    protected gameLogPanel: GameLogPanel;
    protected turnOptionsPanel: TurnOptionsPanel;
    protected buildOptionsModal: BuildOptionsModal;

    // =========================================================================
    // Abstract Methods - Must be implemented by child classes
    // =========================================================================

    /**
     * Initialize the layout system (DockView, CSS Grid, Mobile drawer, etc.)
     * Called during performLocalInit before components are created
     */
    protected abstract initializeLayout(): Promise<void>;

    /**
     * Create all panel instances and attach them to the DOM.
     * This is called after initializeLayout() and may differ between implementations:
     * - DockView: Creates panels lazily when added to dock
     * - Grid: Creates panels immediately and attaches to pre-existing containers
     * - Mobile: Creates panels and attaches to drawer system
     *
     * @returns Array of LCMComponent panels for lifecycle management
     */
    protected abstract createPanels(): LCMComponent[];

    /**
     * Get the container element where the Phaser game scene should render.
     * This may be a pre-existing element (Grid) or created dynamically (DockView).
     */
    protected abstract getGameSceneContainer(): HTMLElement;

    /**
     * Controls when the game scene is created:
     * - true: Create during performLocalInit (Grid, Mobile)
     * - false: Create later during layout initialization (DockView)
     */
    protected abstract shouldCreateGameSceneEarly(): boolean;

    /**
     * Called after the game scene is created.
     * Child classes can use this to perform post-creation setup.
     */
    protected abstract onGameSceneCreated(): void;

    /**
     * Show/focus a specific panel (used for mobile drawer or DockView focus)
     */
    protected abstract showPanel(panelId: PanelId): void;

    /**
     * Get the DOM element for a specific panel
     */
    protected abstract getPanelElement(panelId: PanelId): HTMLElement | null;

    // =========================================================================
    // LCMComponent Interface Implementation
    // =========================================================================

    /**
     * Phase 1: Initialize DOM and discover child components
     */
    async performLocalInit(): Promise<LCMComponent[]> {
        // Load game config first
        this.currentGameId = (document.getElementById("gameIdInput") as HTMLInputElement).value.trim();
        if (!this.currentGameId) {
            throw new Error("Game Id Not Found");
        }

        // Subscribe to events BEFORE creating components
        this.subscribeToGameStateEvents();

        // Initialize layout system (DockView/Grid/Mobile)
        await this.initializeLayout();

        // Create shared World component first
        this.world = new World(this.eventBus, 'Game World');

        // Create build options modal (always separate from layout)
        this.createBuildOptionsModal();

        // Create game scene early if required by layout
        if (this.shouldCreateGameSceneEarly()) {
            this.createGameScene();
        }

        // Create panels (implementation-specific)
        const panels = this.createPanels();

        this.updateGameStatusBanner('Game Loading...');

        // Kick off WASM loading
        await this.loadWASM();

        // Return child components for lifecycle management
        return [
            this.gameScene,
            ...panels,
            this.buildOptionsModal,
        ].filter(c => c != null);
    }

    /**
     * Phase 2: Inject dependencies
     */
    setupDependencies(): void {
        // Pass the theme to the stats panels
        const assetProvider = this.gameScene?.getAssetProvider();
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
        if (this.gameScene) {
            this.gameScene.gameViewPresenterClient = this.gameViewPresenterClient;
        }
        if (this.turnOptionsPanel) {
            this.turnOptionsPanel.gameViewPresenterClient = this.gameViewPresenterClient;
        }
        if (this.buildOptionsModal) {
            this.buildOptionsModal.gameViewPresenterClient = this.gameViewPresenterClient;
        }
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

        // Register browser service
        this.wasmBundle.registerBrowserService('GameViewerPage', this);

        // Initialize the presenter by setting it game data now that all UI components are ready
        await this.initializePresenter();

        // Expose gameScene to console for testing animations
        (window as any).gameScene = this.gameScene;
        console.log("🎮 gameScene exposed to window for animation testing");
        console.log("Try: gameScene.moveUnit(unit, path) or gameScene.showAttackEffect({q:0,r:0}, {q:1,r:0}, 10)");
    }

    // =========================================================================
    // Protected Helper Methods - Available to child classes
    // =========================================================================

    /**
     * Create the BuildOptionsModal (always separate from layout system)
     */
    protected createBuildOptionsModal(): void {
        const modalElement = document.getElementById('build-options-modal');
        if (!modalElement) {
            throw new Error('GameViewerPageBase: build-options-modal element not found');
        }
        this.buildOptionsModal = new BuildOptionsModal(modalElement, this.eventBus, true);
    }

    /**
     * Create the Phaser game scene
     */
    protected createGameScene(): void {
        const container = this.getGameSceneContainer();
        if (!container) {
            throw new Error('GameViewerPageBase: Cannot create game scene - container not found');
        }

        this.gameScene = new PhaserGameScene(container, this.eventBus, true);
        this.onGameSceneCreated();
    }

    /**
     * Load WASM bundle and initialize clients
     */
    protected async loadWASM(): Promise<void> {
        this.wasmBundle = new WeewarBundle();
        this.gamesClient = new GamesServiceClient(this.wasmBundle);
        this.gameViewPresenterClient = new GameViewPresenterClient(this.wasmBundle);
        this.singletonInitializerClient = new SingletonInitializerClient(this.wasmBundle);
        await this.wasmBundle.loadWasm((document.getElementById("wasmBundlePathField") as HTMLInputElement).value);
        await this.wasmBundle.waitUntilReady();
    }

    /**
     * Initialize game using WASM game engine
     */
    protected async initializePresenter(): Promise<void> {
        // Get raw JSON data from page elements
        const gameElement = document.getElementById('game.data-json')!;
        const gameStateElement = document.getElementById('game-state-data-json')!;
        const historyElement = document.getElementById('game-history-data-json')!;

        if (!gameElement?.textContent || gameElement.textContent.trim() === 'null') {
            throw new Error('No game data found in page elements');
        }

        // Call presenter to initialize
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

    /**
     * Subscribe to GameState events
     */
    protected subscribeToGameStateEvents(): void {
        this.addSubscription(WorldEventTypes.WORLD_VIEWER_READY, this);
        this.addSubscription(GameEventTypes.GAME_DATA_LOADED, this);
        this.addSubscription('unit-moved', this);
        this.addSubscription('unit-attacked', this);
        this.addSubscription('turn-ended', this);
    }

    /**
     * Bind game-specific DOM events
     */
    protected bindGameSpecificEvents(): void {
        // End Turn button
        const endTurnBtn = document.getElementById('end-turn-btn');
        if (endTurnBtn) {
            endTurnBtn.addEventListener('click', () => {
                this.gameViewPresenterClient.endTurnButtonClicked({
                    gameId: this.currentGameId || ""
                });
            });
        }

        // Screenshot button
        const screenshotBtn = document.getElementById('capture-screenshot-btn');
        if (screenshotBtn) {
            screenshotBtn.addEventListener('click', () => this.handleScreenshotClick());
        }
    }

    /**
     * Handle Screenshot button click
     */
    protected async handleScreenshotClick(): Promise<void> {
        if (!this.currentGameId) {
            console.error('No game ID available');
            this.showToast('Error', 'No game ID available', 'error');
            return;
        }

        try {
            const blob = await this.gameScene.captureScreenshotAsync('image/png', 0.92);

            if (!blob) {
                this.showToast('Error', 'Failed to capture screenshot', 'error');
                return;
            }

            const formData = new FormData();
            formData.append('screenshot', blob, 'screenshot.png');

            const themeName = AssetThemePreference.get()
            const previewUrl = `/games/${this.currentGameId}/screenshots/${themeName}`;
            const response = await fetch(previewUrl, {
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

    private async addGamePreviewUrl(previewUrl: string): Promise<void> {
    }

    /**
     * Update game status banner
     */
    protected updateGameStatusBanner(status: string, currentPlayer?: number): void {
        const statusElement = document.getElementById('game-status');
        if (statusElement) {
            statusElement.textContent = status;

            const playerColorClass = currentPlayer ? PLAYER_BG_COLORS[currentPlayer] : 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
            statusElement.className = `inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${playerColorClass}`;
        }
    }

    /**
     * Update turn counter display
     */
    protected updateTurnCounter(turnCounter: number): void {
        const turnElement = document.getElementById('turn-counter');
        if (turnElement) {
            turnElement.textContent = `Turn ${turnCounter}`;
        }
    }

    /**
     * Update game UI from game state
     */
    protected updateGameUIFromState(gameState: ProtoGameState): void {
        this.updateGameStatusBanner(`Ready - Player ${gameState.currentPlayer}'s Turn`, gameState.currentPlayer);
        this.updateTurnCounter(gameState.turnCounter);
    }

    /**
     * Update End Turn button enabled/disabled state
     */
    protected updateEndTurnButtonState(currentPlayer: number): void {
        const endTurnBtn = document.getElementById('end-turn-btn') as HTMLButtonElement;
        if (endTurnBtn) {
            // TODO: Get the actual player ID from the game/user context
            const isOurTurn = currentPlayer === 1;

            endTurnBtn.disabled = !isOurTurn;

            if (isOurTurn) {
                endTurnBtn.classList.remove('opacity-50', 'cursor-not-allowed');
                endTurnBtn.classList.add('hover:bg-green-700');
            } else {
                endTurnBtn.classList.add('opacity-50', 'cursor-not-allowed');
                endTurnBtn.classList.remove('hover:bg-green-700');
            }
        }
    }

    /**
     * Hide the loading overlay
     */
    protected hideLoadingOverlay(): void {
        const gameLoadingOverlay = document.getElementById('game-loading') as HTMLElement;
        if (gameLoadingOverlay) {
            gameLoadingOverlay.style.display = 'none';
        }

        super.dismissSplashScreen();
    }

    /**
     * Resize the game canvas
     */
    protected resizeGameCanvas(): void {
        if (this.gameScene) {
            const container = this.getGameSceneContainer();
            if (container) {
                const width = container.clientWidth;
                const height = container.clientHeight;

                setTimeout(() => {
                    if (this.gameScene) {
                        this.gameScene.resize(width, height);
                        this.gameScene.centerCameraOnWorld();
                    }
                }, 100);
            }
        }
    }

    // =========================================================================
    // Event Bus Handling
    // =========================================================================

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
            case 'unit-attacked':
            case 'turn-ended':
                // Could trigger animations, sound effects, etc.
                break;

            default:
                super.handleBusEvent(eventType, data, target, emitter);
        }
    }

    /**
     * Show path visualization on the game scene
     */
    protected showPathVisualization(coords: number[], color: number, thickness: number): void {
        if (!this.gameScene) return;

        const movementLayer = this.gameScene.movementHighlightLayer;
        if (movementLayer) {
            movementLayer.clearAllPaths();
            movementLayer.addPath(coords, color, thickness);
        }
    }

    /**
     * Clear path visualization from the game scene
     */
    protected clearPathVisualization(): void {
        if (!this.gameScene) return;

        const movementLayer = this.gameScene.movementHighlightLayer;
        if (movementLayer) {
            movementLayer.clearAllPaths();
        }
    }

    // =========================================================================
    // GameViewerPageMethods Interface - Browser RPC Methods
    // =========================================================================

    async setTurnOptionsContent(request: SetContentRequest) {
        console.log("setTurnOptionsContent called on the browser: ", request);
        this.turnOptionsPanel.innerHTML = request.innerHtml;
        await this.turnOptionsPanel.hydrateThemeImages();
        return {};
    }

    async showBuildOptions(request: ShowBuildOptionsRequest): Promise<ShowBuildOptionsResponse> {
        console.log("showBuildOptions called on the browser:", request);

        if (request.hide) {
            this.buildOptionsModal.hide();
        } else {
            await this.buildOptionsModal.show(request.innerHtml, request.q, request.r);
        }

        return {};
    }

    async setUnitStatsContent(request: SetContentRequest) {
        console.log("setUnitStatsContent called on the browser");
        this.unitStatsPanel.innerHTML = request.innerHtml;
        await this.unitStatsPanel.hydrateThemeImages();
        return {};
    }

    async setDamageDistributionContent(request: SetContentRequest) {
        console.log("setDamageDistributionContent called on the browser");
        this.damageDistributionPanel.innerHTML = request.innerHtml;
        await this.damageDistributionPanel.hydrateThemeImages();
        return {};
    }

    async setTerrainStatsContent(request: SetContentRequest) {
        console.log("setTerrainStatsContent called on the browser");
        this.terrainStatsPanel.innerHTML = request.innerHtml;
        await this.terrainStatsPanel.hydrateThemeImages();
        return {};
    }

    /**
     * Set compact summary card content (mobile-specific, no-op for desktop/grid)
     */
    async setCompactSummaryCard(request: SetContentRequest): Promise<SetContentResponse> {
        // Default implementation: no-op for desktop and grid layouts
        // Mobile layout overrides this to show compact card
        return {};
    }

    /**
     * Set allowed panels and their order (mobile-specific, no-op for desktop/grid)
     */
    async setAllowedPanels(request: SetAllowedPanelsRequest): Promise<SetAllowedPanelsResponse> {
        console.log("setAllowedPanels called on the browser:", request.panelIds);
        // Default implementation: no-op for desktop and grid layouts
        // Mobile layout overrides this to update bottom bar buttons
        return {};
    }

    async showHighlights(request: ShowHighlightsRequest) {
        console.log("showHighlights called:", request);
        if (request.highlights) {
            this.gameScene.showHighlights(request.highlights);
        }
        return {};
    }

    async clearHighlights(request: ClearHighlightsRequest) {
        console.log("clearHighlights called:", request);
        this.gameScene.clearHighlights(request.types || []);
        return {};
    }

    async showPath(request: ShowPathRequest) {
        console.log("showPath called:", request);
        if (request.coords) {
            this.gameScene.showPath(request.coords, request.color, request.thickness);
        }
        return {};
    }

    async clearPaths(request: ClearPathsRequest) {
        console.log("clearPaths called:", request);
        this.gameScene.clearPaths();
        return {};
    }

    async logMessage(request: LogMessageRequest) {
        console.log("logMessage called on the browser");
        return {};
    }

    async setGameState(req: SetGameStateRequest) {
        console.log("setGameState called on the browser");
        const worldData = req.state!.worldData!;
        const game = req.game!;

        this.world.loadTilesAndUnits(worldData.tiles || [], worldData.units || []);
        this.world.setName(game.name || 'Untitled Game');

        await this.gameScene.loadWorld(this.world);
        this.showToast('Success', `Game loaded: ${game.name || this.world.getName() || 'Untitled'}`, 'success');

        this.hideLoadingOverlay();
        this.resizeGameCanvas();

        this.updateGameUIFromState(req.state!);
        this.gameLogPanel.logGameEvent(`Game loaded: ${req.state!.gameId}`, 'system');
        return {};
    }

    async setTileAt(request: { q: number, r: number, tile: Tile }) {
        console.log("setTileAt called on the browser:", request);
        this.world.setTileDirect(request.tile);
        return {};
    }

    async setUnitAt(request: SetUnitAtRequest): Promise<SetUnitAtResponse> {
        console.log("setUnitAt called on the browser:", request);
        if (request.unit) {
            await this.gameScene.setUnit(request.unit, { flash: request.flash, appear: request.appear });
            this.world.setUnitDirect(request.unit);
        }
        return {};
    }

    async removeTileAt(request: { q: number, r: number }) {
        console.log("removeTileAt called on the browser:", request);
        this.world.removeTileAt(request.q, request.r);
        return {};
    }

    async removeUnitAt(request: RemoveUnitAtRequest): Promise<RemoveUnitAtResponse> {
        console.log("removeUnitAt called on the browser:", request);
        await this.gameScene.removeUnit(request.q, request.r, { animate: request.animate });
        this.world.removeUnitAt(request.q, request.r);
        return {};
    }

    async moveUnit(request: MoveUnitRequest): Promise<MoveUnitResponse> {
        console.log("moveUnit called on the browser:", request);
        if (request.unit && request.path) {
            await this.gameScene.moveUnit(request.unit, request.path);
            this.world.setUnitDirect(request.unit);
        }
        return {};
    }

    async showAttackEffect(request: ShowAttackEffectRequest): Promise<ShowAttackEffectResponse> {
        console.log("showAttackEffect called on the browser:", request);
        await this.gameScene.showAttackEffect(
            { q: request.fromQ, r: request.fromR },
            { q: request.toQ, r: request.toR },
            request.damage,
            request.splashTargets
        );
        return {};
    }

    async showHealEffect(request: ShowHealEffectRequest): Promise<ShowHealEffectResponse> {
        console.log("showHealEffect called on the browser:", request);
        await this.gameScene.showHealEffect(request.q, request.r, request.amount);
        return {};
    }

    async showCaptureEffect(request: ShowCaptureEffectRequest): Promise<ShowCaptureEffectResponse> {
        console.log("showCaptureEffect called on the browser:", request);
        await this.gameScene.showCaptureEffect(request.q, request.r);
        return {};
    }

    async updateGameStatus(request: { currentPlayer: number, turnCounter: number }) {
        console.log("updateGameStatus called on the browser:", request);
        this.updateGameStatusBanner(`Ready - Player ${request.currentPlayer}'s Turn`, request.currentPlayer);
        this.updateTurnCounter(request.turnCounter);
        this.updateEndTurnButtonState(request.currentPlayer);
        return {};
    }
}
