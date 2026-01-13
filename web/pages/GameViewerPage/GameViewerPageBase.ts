import { BasePage, EventBus, LCMComponent, LifecycleController } from '@panyam/tsappkit';
import LilbattleBundle from '../../gen/wasmjs';
import { GamesServiceClient } from '../../gen/wasmjs/lilbattle/v1/services/gamesServiceClient';
import { GameViewerPageMethods, GameViewerPageClient as GameViewerPageClient } from '../../gen/wasmjs/lilbattle/v1/services/gameViewerPageClient';
import { GameViewPresenterClient as GameViewPresenterClient } from '../../gen/wasmjs/lilbattle/v1/services/gameViewPresenterClient';
import { SingletonInitializerServiceClient as SingletonInitializerClient } from '../../gen/wasmjs/lilbattle/v1/services/singletonInitializerServiceClient';
import { AssetThemePreference } from '../common/AssetThemePreference';
import { PhaserGameScene } from './PhaserGameScene';
import { Unit, Tile, World } from '../common/World';
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
} from '../../gen/wasmjs/lilbattle/v1/models/interfaces';
import * as models from '../../gen/wasmjs/lilbattle/v1/models/models';
import { create } from '@bufbuild/protobuf';
import { PLAYER_BG_COLORS } from '../common/ColorsAndNames';
import { TerrainStatsPanel } from './TerrainStatsPanel';
import { UnitStatsPanel } from './UnitStatsPanel';
import { DamageDistributionPanel } from './DamageDistributionPanel';
import { GameLogPanel } from './GameLogPanel';
import { TurnOptionsPanel } from './TurnOptionsPanel';
import { BuildOptionsModal } from './BuildOptionsModal';
import { GameStatePanel } from './GameStatePanel';
import { RulesTable, TerrainStats } from '../common/RulesTable';
import { GameSyncManager, SyncState } from './GameSyncManager';
import { AnimationQueue } from '../common/animations/AnimationQueue';

/**
 * Panel type identifiers
 */
export type PanelId = 'terrain-stats' | 'unit-stats' | 'damage-distribution' | 'turn-options' | 'game-log' | 'game-state' | 'build-options';

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
    protected wasmBundle: LilbattleBundle;
    protected gamesClient: GamesServiceClient;
    protected gameViewPresenterClient: GameViewPresenterClient;
    protected singletonInitializerClient: SingletonInitializerClient;
    protected currentGameId: string | null;
    private clientReadySent: boolean = false;

    // Multiplayer sync
    protected syncManager: GameSyncManager | null = null;

    // Core game components
    protected gameScene: PhaserGameScene;
    protected world: World;
    protected rulesTable: RulesTable = new RulesTable();
    protected animationQueue: AnimationQueue = new AnimationQueue();

    // UI Panels
    protected terrainStatsPanel: TerrainStatsPanel;
    protected unitStatsPanel: UnitStatsPanel;
    protected damageDistributionPanel: DamageDistributionPanel;
    protected gameLogPanel: GameLogPanel;
    protected turnOptionsPanel: TurnOptionsPanel;
    protected buildOptionsModal: BuildOptionsModal;
    protected gameStatePanel: GameStatePanel;

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

        // Register browser service
        this.wasmBundle.registerBrowserService('GameViewerPage', this);

        // Initialize the presenter by setting it game data now that all UI components are ready
        await this.initializePresenter();

        // Initialize multiplayer sync if enabled
        this.initializeMultiplayerSync();

        // Expose gameScene and animationQueue to console for testing
        (window as any).gameScene = this.gameScene;
        (window as any).animationQueue = this.animationQueue;
        console.log("🎮 gameScene and animationQueue exposed to window for testing");
        console.log("Try: animationQueue.enqueue(() => gameScene.moveUnit(unit, path))");
    }

    /**
     * Initialize multiplayer sync manager
     * Override in child classes or configure via game settings
     */
    protected initializeMultiplayerSync(): void {
        // Check if multiplayer sync is enabled (e.g., via page config or game data)
        const enableSync = this.isMultiplayerSyncEnabled();
        if (!enableSync || !this.currentGameId) {
            return;
        }

        console.log('[GameViewerPage] Initializing multiplayer sync');
        this.syncManager = new GameSyncManager(
            this.gameViewPresenterClient,
            this.currentGameId,
            {
                onStateChange: (state, error) => this.onSyncStateChange(state, error),
                onRemoteUpdate: (update) => this.onRemoteUpdate(update),
            }
        );
        this.syncManager.connect();
    }

    /**
     * Check if multiplayer sync should be enabled for this game.
     * Override in child classes to implement custom logic.
     * @returns true if sync should be enabled
     */
    protected isMultiplayerSyncEnabled(): boolean {
        // Check for sync=true query parameter (for testing)
        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.get('sync') === 'true') {
            return true;
        }
        // TODO: Check game config for multiplayer mode
        return false;
    }

    /**
     * Handle sync state changes
     */
    protected onSyncStateChange(state: SyncState, error?: string): void {
        console.log(`[GameViewerPage] Sync state: ${state}`, error || '');

        // Update UI to show connection state
        const syncIndicator = document.getElementById('sync-indicator');
        if (syncIndicator) {
            syncIndicator.className = `sync-state-${state}`;
            syncIndicator.title = error || state;
        }

        // Log to game log
        if (state === 'connected') {
            this.gameLogPanel?.logGameEvent('Multiplayer sync connected', 'system');
        } else if (state === 'error') {
            this.gameLogPanel?.logGameEvent(`Sync error: ${error}`, 'system');
        }
    }

    /**
     * Handle remote game updates
     */
    protected onRemoteUpdate(update: any): void {
        // Log significant updates to game log
        if (update.movesPublished) {
            this.gameLogPanel?.logGameEvent(
                `Player ${update.movesPublished.player} made ${update.movesPublished.moves?.length || 0} move(s)`,
                'moves'
            );
        }
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
        this.wasmBundle = new LilbattleBundle();
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

        // Get viewer user ID from page data attribute or hidden input
        const viewerUserIdElement = document.getElementById('viewer-user-id') || document.getElementById('viewerUserIdInput');
        const viewerUserId = viewerUserIdElement?.textContent?.trim() || (viewerUserIdElement as HTMLInputElement)?.value || '';

        // Call presenter to initialize
        const response = await this.singletonInitializerClient.initializeSingleton({
            gameId: this.currentGameId || "",
            gameData: gameElement!.textContent,
            gameState: gameStateElement?.textContent || '{}',
            moveHistory: historyElement?.textContent || '{"gameId":"","groups":[]}',
            viewerUserId: viewerUserId,
        });

        if (!response.response!.success) {
            throw new Error(`WASM load failed: ${response.response!.error}`);
        }
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

        // Set up JoinGame callback on GameStatePanel
        if (this.gameStatePanel) {
            this.gameStatePanel.setJoinGameCallback((playerId: number) => {
                this.handleJoinGame(playerId);
            });
        }
    }

    /**
     * Handle Join Game button click
     */
    protected async handleJoinGame(playerId: number): Promise<void> {
        if (!this.currentGameId) {
            this.showToast('Error', 'No game ID available', 'error');
            return;
        }

        try {
            const response = await this.gamesClient.joinGame({
                gameId: this.currentGameId,
                playerId: playerId,
            });

            if (response.game) {
                this.showToast('Success', `You joined as Player ${playerId}`, 'success');
                // Reload the page to refresh game state with new player assignment
                window.location.reload();
            }
        } catch (error) {
            console.error('JoinGame error:', error);
            const errorMsg = error instanceof Error ? error.message : 'Failed to join game';
            this.showToast('Error', errorMsg, 'error');
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
                }, 50);
            }
        }
    }

    // =========================================================================
    // GameViewerPageMethods Interface - Browser RPC Methods
    // =========================================================================

    setTurnOptionsContent(request: SetContentRequest) {
        this.turnOptionsPanel.innerHTML = request.innerHtml;
        return {};
    }

    async showBuildOptions(request: ShowBuildOptionsRequest): Promise<ShowBuildOptionsResponse> {

        if (request.hide) {
            this.buildOptionsModal.hide();
        } else {
            await this.buildOptionsModal.show(request.innerHtml, request.q, request.r);
        }

        return {};
    }

    setUnitStatsContent(request: SetContentRequest) {
        this.unitStatsPanel.innerHTML = request.innerHtml;
        return {};
    }

    setDamageDistributionContent(request: SetContentRequest) {
        this.damageDistributionPanel.innerHTML = request.innerHtml;
        return {};
    }

    setTerrainStatsContent(request: SetContentRequest) {
        this.terrainStatsPanel.innerHTML = request.innerHtml;
        return {};
    }

    setGameStatePanelContent(request: SetContentRequest): SetContentResponse {
        this.gameStatePanel.innerHTML = request.innerHtml;
        return {};
    }

    /**
     * Set compact summary card content (mobile-specific, no-op for desktop/grid)
     */
    setCompactSummaryCard(request: SetContentRequest): SetContentResponse {
        // Default implementation: no-op for desktop and grid layouts
        // Mobile layout overrides this to show compact card
        return {};
    }

    /**
     * Set allowed panels and their order (mobile-specific, no-op for desktop/grid)
     */
    setAllowedPanels(request: SetAllowedPanelsRequest): SetAllowedPanelsResponse {
        // Default implementation: no-op for desktop and grid layouts
        // Mobile layout overrides this to update bottom bar buttons
        return {};
    }

    showHighlights(request: ShowHighlightsRequest) {
        console.log("showHighlights called:", request);
        if (request.highlights) {
            this.gameScene.showHighlights(request.highlights);
        }
        return {};
    }

    clearHighlights(request: ClearHighlightsRequest) {
        console.log("clearHighlights called:", request);
        this.gameScene.clearHighlights(request.types || []);
        return {};
    }

    showPath(request: ShowPathRequest) {
        console.log("showPath called:", request);
        if (request.coords) {
            this.gameScene.showPath(request.coords, request.color, request.thickness);
        }
        return {};
    }

    clearPaths(request: ClearPathsRequest) {
        console.log("clearPaths called:", request);
        this.gameScene.clearPaths();
        return {};
    }

    logMessage(request: LogMessageRequest) {
        return {};
    }

    async setGameState(req: SetGameStateRequest) {
        const worldData = req.state!.worldData!;
        const game = req.game!;

        // Handle both map format (tilesMap/unitsMap) and array format (tiles/units)
        const tiles = (worldData as any).tilesMap || (worldData as any).tiles_map || [];
        const units = (worldData as any).unitsMap || (worldData as any).units_map || [];
        this.world.loadTilesAndUnits(tiles, units);
        this.world.setName(game.name || 'Untitled Game');

        await this.gameScene.loadWorld(this.world);
        this.showToast('Success', `Game loaded: ${game.name || this.world.getName() || 'Untitled'}`, 'success');

        this.hideLoadingOverlay();
        this.resizeGameCanvas();

        this.updateGameUIFromState(req.state!);
        this.gameLogPanel.logGameEvent(`Game loaded: ${req.state!.gameId}`, 'system');

        // Notify presenter that client is ready for visual updates (only once on initial load)
        if (!this.clientReadySent && this.currentGameId) {
            this.clientReadySent = true;
            // Fire and forget - don't block setGameState return
            this.gameViewPresenterClient.clientReady({ gameId: this.currentGameId }).catch(err => {
                console.error('[GameViewerPage] clientReady failed:', err);
            });
        }
        return {};
    }

    async setTileAt(request: { q: number, r: number, tile: Tile }) {
        this.world.setTileDirect(request.tile);
        return {};
    }

    async setUnitAt(request: SetUnitAtRequest): Promise<SetUnitAtResponse> {
        if (request.unit) {
            const unit = request.unit;
            const options = { flash: request.flash, appear: request.appear };

            // Queue animation, update world data AFTER animation completes
            this.animationQueue.enqueue(
                () => this.gameScene.setUnit(unit, options),
                () => {
                    this.world.setUnitDirect(unit);
                    return true;
                }
            );
        }
        return {};
    }

    async removeTileAt(request: { q: number, r: number }) {
        this.world.removeTileAt(request.q, request.r);
        return {};
    }

    async removeUnitAt(request: RemoveUnitAtRequest): Promise<RemoveUnitAtResponse> {
        const { q, r, animate } = request;

        // Queue animation, update world data AFTER animation completes
        this.animationQueue.enqueue(
            () => this.gameScene.removeUnit(q, r, { animate }),
            () => {
                this.world.removeUnitAt(q, r);
                return true;
            }
        );
        return {};
    }

    async moveUnit(request: MoveUnitRequest): Promise<MoveUnitResponse> {
        if (request.unit && request.path && request.path.length >= 2) {
            const unit = request.unit;
            const path = request.path;
            const oldPos = request.path[0];

            console.log(`[GameViewerPage] Queueing moveUnit: ${unit.shortcut} from (${oldPos.q},${oldPos.r}) to (${path[path.length-1].q},${path[path.length-1].r}), path length: ${path.length}`);

            // Queue animation, update world data AFTER animation completes
            this.animationQueue.enqueue(
                () => {
                    console.log(`[GameViewerPage] Executing moveUnit animation for ${unit.shortcut}`);
                    return this.gameScene.moveUnit(unit, path);
                },
                () => {
                    // Update World data model after animation
                    this.world.removeUnitAt(oldPos.q, oldPos.r);
                    this.world.setUnitDirect(unit);
                    return true; // Continue with next animation
                }
            );
        } else if (request.unit) {
            // Fallback: No valid path, just update unit at its current position
            console.warn('[GameViewerPage] moveUnit called without valid path, unit:', request.unit);
            this.world.setUnitDirect(request.unit);
        }
        return {};
    }

    async showAttackEffect(request: ShowAttackEffectRequest): Promise<ShowAttackEffectResponse> {
        const from = { q: request.fromQ, r: request.fromR };
        const to = { q: request.toQ, r: request.toR };
        const damage = request.damage;
        const splashTargets = request.splashTargets;
        this.animationQueue.enqueue(() => this.gameScene.showAttackEffect(from, to, damage, splashTargets));
        return {};
    }

    async showHealEffect(request: ShowHealEffectRequest): Promise<ShowHealEffectResponse> {
        const { q, r, amount } = request;
        this.animationQueue.enqueue(() => this.gameScene.showHealEffect(q, r, amount));
        return {};
    }

    async showCaptureEffect(request: ShowCaptureEffectRequest): Promise<ShowCaptureEffectResponse> {
        const { q, r } = request;
        this.animationQueue.enqueue(() => this.gameScene.showCaptureEffect(q, r));
        return {};
    }

    async updateGameStatus(request: { currentPlayer: number, turnCounter: number }) {
        this.updateTurnCounter(request.turnCounter);
        this.updateEndTurnButtonState(request.currentPlayer);
        return {};
    }
}
