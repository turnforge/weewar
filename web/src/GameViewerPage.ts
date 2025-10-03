 import { BasePage } from '../lib/BasePage';
import { EventBus } from '../lib/EventBus';
import { PhaserGameScene } from './phaser/PhaserGameScene';
import { Unit, Tile, World } from './World';
import { GameState } from './GameState';
import { 
    GameState as ProtoGameState, 
    Game as ProtoGame, 
    GameConfiguration as ProtoGameConfiguration, 
    MoveOption, 
    AttackOption, 
    GameMove,
    GetOptionsAtResponse,
    GameOption,
    WorldData
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
import { GameActionsPanel, GameActionsCallbacks } from './GameActionsPanel';
import { TurnOptionsPanel } from './TurnOptionsPanel';
import { GameEventTypes, WorldEventTypes } from './events';
import { RulesTable, TerrainStats } from './RulesTable';
import { DockviewApi, DockviewComponent } from 'dockview-core';

/**
 * Result of a game action command - used for testing and accessibility
 */
export interface ActionResult {
    success: boolean;
    message: string;
    data?: any;
    error?: string;
}

/**
 * Game state information for testing and debugging
 */
export interface GameStateInfo {
    gameId: string;
    currentPlayer: number;
    turnCounter: number;
    selectedUnit?: { q: number, r: number };
    unitsCount: number;
    tilesCount: number;
}

/**
 * Command interface for testing and accessibility
 * These methods provide high-level game actions that can be easily tested
 */
export interface GameViewerCommands {
    selectUnitAt(q: number, r: number): Promise<ActionResult>;
    moveSelectedUnitTo(q: number, r: number): Promise<ActionResult>;
    attackWithSelectedUnit(q: number, r: number): Promise<ActionResult>;
    endCurrentPlayerTurn(): Promise<ActionResult>;
    getGameState(): Promise<GameStateInfo>;
    clearSelection(): ActionResult;
}

/**
 * Game Viewer Page - Interactive game play interface
 * Responsible for:
 * - Loading world as a game instance
 * - Coordinating WASM game engine
 * - Managing game state and turn flow
 * - Handling player interactions (unit selection, movement, attacks)
 * - Providing game controls and UI feedback
 */
export class GameViewerPage extends BasePage implements LCMComponent, GameViewerCommands {
    private currentGameId: string | null;
    private gameScene: PhaserGameScene
    private gameState: GameState
    private world: World  // ✅ Shared World component
    private terrainStatsPanel: TerrainStatsPanel
    private unitStatsPanel: UnitStatsPanel
    private damageDistributionPanel: DamageDistributionPanel
    private gameLogPanel: GameLogPanel
    private gameActionsPanel: GameActionsPanel
    private turnOptionsPanel: TurnOptionsPanel
    private rulesTable: RulesTable = new RulesTable();
    
    // Dockview interface
    private dockview: DockviewApi;
    private themeObserver: MutationObserver | null = null;
    
    // Game configuration accessed directly from WASM-cached Game proto
    
    // UI state - gameLog is now handled by GameLogPanel
    
    // Move execution state
    private selectedUnitCoord: { q: number, r: number } | null = null;
    private availableMovementOptions: MoveOption[] = [];
    private isProcessingMove: boolean = false;

    // =============================================================================
    // LCMComponent Interface Implementation
    // =============================================================================

    /**
     * Phase 1: Initialize DOM and discover child components
     */
    performLocalInit(): LCMComponent[] {
        // Load game config first
        this.currentGameId = (document.getElementById("gameIdInput") as HTMLInputElement).value.trim()
        if (!this.currentGameId) {
          throw new Error("Game Id Not Found")
        }
        
        // Subscribe to events BEFORE creating components
        this.subscribeToGameStateEvents();
        
        // Create child components
        this.createComponents();
        
        this.updateGameStatus('Game Loading...');
        
        // Start WASM and game data loading early (parallel to WorldViewer initialization)
        this.initializeGameWithWASM().then(() => {
            // Emit event to indicate game data is ready
            this.eventBus.emit(GameEventTypes.GAME_DATA_LOADED, { gameId: this.currentGameId }, this, this);
        }).catch(error => {
            console.error('GameViewerPage: WASM initialization failed:', error);
            this.updateGameStatus('WASM initialization failed');
        });

        console.assert(this.gameScene != null, "Game scene could not be created")
        console.assert(this.gameState != null, "gameState could not be created")
        console.assert(this.world != null, "World could not be created")
        console.assert(this.terrainStatsPanel != null, "terrainStatsPanel could not be created")
        console.assert(this.unitStatsPanel != null, "unitStatsPanel could not be created")
        console.assert(this.damageDistributionPanel != null, "damageDistributionPanel could not be created")
        console.assert(this.gameLogPanel != null, "gameLogPanel could not be created")
        console.assert(this.gameActionsPanel != null, "gameActionsPanel could not be created")
        console.assert(this.rulesTable != null, "rulesTable could not be created")
        
        // Return child components for lifecycle management
        // Note: World and GameState don't extend BaseComponent, so not included in lifecycle
        return [
            this.gameScene,
            this.terrainStatsPanel,
            this.unitStatsPanel,
            this.damageDistributionPanel,
            this.gameLogPanel,
            this.gameActionsPanel,
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
        }

        // Set up scene click callback now that gameScene is initialized
        this.gameScene.sceneClickedCallback = (context: any, layer: string, extra?: any): void => {
            const { hexQ: q, hexR: r } = context;
            
            // Get tile and unit data from World using coordinates
            switch (layer) {
                case 'movement-highlight':
                    // Get moveOption from the layer itself
                    const movementLayer = this.gameScene.movementHighlightLayer;
                    const moveOption = movementLayer?.getMoveOptionAt(q, r);
                    this.handleMovementClick(q, r, moveOption);
                    break;
                    
                case 'base-map':
                    const unit = this.world?.getUnitAt(q, r);
                    const tile = this.world?.getTileAt(q, r);
                    
                    // Always show terrain info (even when unit is present)
                    this.handleTileClick(q, r, tile);
                    
                    // If there's a unit, also handle unit logic and show unit info in unit panel
                    if (unit) {
                        this.handleUnitClick(q, r);
                        // Update unit stats panel with unit info
                        if (this.unitStatsPanel) {
                            this.unitStatsPanel.updateUnitInfo(unit);
                        }
                        // Update damage distribution panel with unit info
                        if (this.damageDistributionPanel) {
                            this.damageDistributionPanel.updateUnitInfo(unit);
                        }
                    } else {
                        // Empty tile clicked - clear selection
                        this.clearSelection();
                        // Clear unit info from unit stats panel when no unit
                        if (this.unitStatsPanel) {
                            this.unitStatsPanel.clearUnitInfo();
                        }
                        // Clear unit info from damage distribution panel when no unit
                        if (this.damageDistributionPanel) {
                            this.damageDistributionPanel.clearUnitInfo();
                        }
                    }
                    break;
                    
                default:
                    console.log(`[GameViewerPage] Unhandled layer click: ${layer}`);
            }
        };
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
        
        this.checkAndLoadWorldIntoViewer();
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

    /*
    public destroy(): void {
        // Save layout before destroying
        this.saveDockviewLayout();
        
        // Dispose dockview
        if (this.dockview) {
            this.dockview.dispose();
        }
        
        // Clean up theme observer
        if (this.themeObserver) {
            this.themeObserver.disconnect();
            this.themeObserver = null;
        }
        
        if (this.gameScene) {
            this.gameScene.destroy();
            this.gameScene = null as any;
        }
        
        if (this.gameState) {
            // GameState no longer has destroy method (not a BaseComponent)
            this.gameState = null as any;
        }

        if (this.terrainStatsPanel) {
            this.terrainStatsPanel.destroy();
            this.terrainStatsPanel = null as any;
        }

        if (this.gameLogPanel) {
            this.gameLogPanel.destroy();
            this.gameLogPanel = null as any;
        }

        if (this.gameActionsPanel) {
            this.gameActionsPanel.destroy();
            this.gameActionsPanel = null as any;
        }

        if (this.rulesTable) {
            this.rulesTable = null as any;
        }
        
        this.currentGameId = null;
    }
   */

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

    // State tracking for initialization
    private gameDataReady = false;

    /**
     * Handle events from the EventBus
     */
    public handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
        switch(eventType) {
            case GameEventTypes.GAME_DATA_LOADED:
                this.gameDataReady = true;
                
                // Check if both viewer and game data are ready
                this.checkAndLoadWorldIntoViewer();
                break;
            
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

        // ✅ Create GameState with direct EventBus connection (no DOM element needed)
        this.gameState = new GameState(this.eventBus);

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
                    case 'game-actions':
                        return this.createGameActionsComponent();
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

        // Add game actions panel (left side)
        this.dockview.addPanel({
            id: 'game-actions-panel',
            component: 'game-actions', 
            title: 'Actions',
            position: { 
                direction: 'left',
                referencePanel: 'main-game-panel'
            }
        });

        // Add game log panel (bottom of actions panel)
        this.dockview.addPanel({
            id: 'game-log-panel',
            component: 'game-log',
            title: 'Game Log', 
            position: { 
                direction: 'below',
                referencePanel: 'game-actions-panel'
            }
        });

        // Set panel sizes for optimal viewing
        setTimeout(() => {
            this.dockview.getPanel('terrain-stats-panel')?.api.setSize({ width: 320 });
            this.dockview.getPanel('game-actions-panel')?.api.setSize({ width: 280 });
            this.dockview.getPanel('game-log-panel')?.api.setSize({ height: 200 });
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
                // Set dependencies
                if (this.gameState) {
                    this.turnOptionsPanel.setGameState(this.gameState);
                }
                if (this.world) {
                    this.turnOptionsPanel.setWorld(this.world);
                }
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
     * Create game actions component
     */
    private createGameActionsComponent() {
        const template = document.getElementById('game-actions-panel-template');
        if (!template) {
            throw new Error('game-actions-panel-template not found');
        }

        const element = template.cloneNode(true) as HTMLElement;
        element.style.display = 'block';
        // Keep the template ID, don't create wrapper instance

        // Create callbacks object for GameActionsPanel
        const callbacks: GameActionsCallbacks = {
            onEndTurn: () => this.endCurrentPlayerTurn(),
            onShowAllUnits: () => this.showAllPlayerUnits(),
            onCenterOnAction: () => this.centerOnAction(),
            onMoveUnit: () => this.selectMoveMode(),
            onAttackUnit: () => this.selectAttackMode(),
            onUndo: () => this.undoMove()
        };

        return {
            element,
            init: () => {
                // Create GameActionsPanel with the cloned element and callbacks
                this.gameActionsPanel = new GameActionsPanel(element, this.eventBus, callbacks);
            },
            dispose: () => {
                // GameActionsPanel cleanup will be handled by LCM lifecycle
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


    /**
     * Check if both WorldViewer and game data are ready, then load world into viewer
     */
    private async checkAndLoadWorldIntoViewer(): Promise<void> {
        if (!this.gameDataReady) {
            console.warn('GameViewerPage: Waiting for both viewer and game data to be ready', {
                gameDataReady: this.gameDataReady
            });
            return;
        }
        
        try {
            // ✅ Get world data from WASM and load into shared World component
            const worldData = await this.gameState.getWorldData();
            const game = await this.gameState.getCurrentGame();
            
            // Load data into shared World component
            this.world.loadTilesAndUnits(worldData.tiles || [], worldData.units || []);
            this.world.setName(game.name || 'Untitled Game');
            
            // Load world into viewer using shared World
            if (this.gameScene && this.world) {
                await this.gameScene.loadWorld(this.world);
                this.showToast('Success', `Game loaded: ${game.name || this.world.getName() || 'Untitled'}`, 'success');
                
                // Hide the loading overlay now that the game is loaded
                this.hideLoadingOverlay();
                
                // Ensure the game canvas is properly sized after loading
                this.resizeGameCanvas();
                
                // Update UI with loaded game state
                const gameState = await this.gameState.getCurrentGameState();
                this.updateGameUIFromState(gameState);
                this.gameLogPanel.logGameEvent(`Game loaded: ${gameState.gameId}`, 'system');
            } else {
                throw new Error('GameScene or World not available');
            }
        } catch (error) {
            console.error('GameViewerPage: Failed to load world into viewer:', error);
            this.updateGameStatus('Failed to load world');
            this.showToast('Error', 'Failed to load world', 'error');
        }
    }

    /**
     * Initialize game using WASM game engine
     * This now handles both WASM loading and World creation in GameState
     */
    private async initializeGameWithWASM(): Promise<void> {
        if (!this.gameState) {
            throw new Error('GameState component not initialized');
        }

        // TODO - Come back to this Wait for WASM to be ready (only async part)
        // await this.gameState.waitUntilReady();
        
        // Load game data into WASM singletons and create World object in GameState
        await this.gameState.loadGameDataToWasm();
        
        // Refresh unit labels in Phaser scene with the loaded World data
        if (this.world && this.gameScene) {
            this.gameScene.refreshUnitLabels(this.world);
        }
    }

    /**
     * Bind page-specific events (required by BasePage)
     * This method is called by BasePage constructor, but we're using external LifecycleController
     * so we make this a no-op and handle event binding in LCMComponent.activate()
     */
    protected bindSpecificEvents(): void {}

    /**
     * Internal method to bind game-specific events (called from activate() phase)
     */
    private bindGameSpecificEvents(): void {
        // End Turn button
        const endTurnBtn = document.getElementById('end-turn-btn');
        if (endTurnBtn) {
            endTurnBtn.addEventListener('click', () => {
                this.endCurrentPlayerTurn(); // Use unified method
            });
        }

        // Undo Move button
        const undoBtn = document.getElementById('undo-move-btn');
        if (undoBtn) {
            undoBtn.addEventListener('click', this.undoMove.bind(this));
        }

        // Unit selection buttons
        const moveUnitBtn = document.getElementById('move-unit-btn');
        if (moveUnitBtn) {
            moveUnitBtn.addEventListener('click', this.selectMoveMode.bind(this));
        }

        const attackUnitBtn = document.getElementById('attack-unit-btn');
        if (attackUnitBtn) {
            attackUnitBtn.addEventListener('click', this.selectAttackMode.bind(this));
        }

        // Utility buttons
        const showAllUnitsBtn = document.getElementById('select-all-units-btn');
        if (showAllUnitsBtn) {
            showAllUnitsBtn.addEventListener('click', this.showAllPlayerUnits.bind(this));
        }

        const centerActionBtn = document.getElementById('center-on-action-btn');
        if (centerActionBtn) {
            centerActionBtn.addEventListener('click', this.centerOnAction.bind(this));
        }
    }

    // =============================================================================
    // GameViewerCommands Interface Implementation (for testing and accessibility)
    // =============================================================================

    /**
     * Select a unit at the given coordinates
     * This is the unified method used by both UI clicks and command interface
     */
    async selectUnitAt(q: number, r: number): Promise<ActionResult> {
        try {
            // Check if there's a unit at this position
            const unit = this.world?.getUnitAt(q, r);
            if (!unit) {
                const error = {
                    success: false,
                    message: `No unit found at position (${q}, ${r})`,
                    error: 'No unit at coordinates'
                } as ActionResult;
                return error;
            }

            // Check if this unit is already selected - if so, deselect it
            if (this.selectedUnitCoord && this.selectedUnitCoord.q === q && this.selectedUnitCoord.r === r) {
                this.clearSelection();
                return {
                    success: true,
                    message: `Unit deselected at (${q}, ${r})`,
                    data: { action: 'deselected' }
                } as ActionResult;
            }

            // Get options for this position to see if unit is selectable
            const response = await this.gameState.getOptionsAt(q, r);
            const options = response.options || [];
            
            const hasMovementOptions = options.some((opt: any) => opt.move !== undefined);
            const hasAttackOptions = options.some((opt: any) => opt.attack !== undefined);
            
            if (!hasMovementOptions && !hasAttackOptions) {
                const error = {
                    success: false,
                    message: `Unit at (${q}, ${r}) has no available actions`,
                    error: 'Unit not actionable',
                    data: { unit: unit, optionsCount: options.length }
                } as ActionResult;
                return error;
            }

            // Use existing unit selection logic
            this.processUnitSelection(q, r, options);

            // Success result
            const result = {
                success: true,
                message: `Unit selected at (${q}, ${r})`,
                data: { 
                    unit: unit, 
                    movementOptions: hasMovementOptions ? options.filter((opt: any) => opt.move).length : 0,
                    attackOptions: hasAttackOptions ? options.filter((opt: any) => opt.attack).length : 0
                }
            } as ActionResult;

            return result;

        } catch (error) {
            return {
                success: false,
                message: `Failed to select unit at (${q}, ${r})`,
                error: error instanceof Error ? error.message : String(error)
            } as ActionResult;
        }
    }

    /**
     * Move the currently selected unit to target coordinates
     */
    async moveSelectedUnitTo(q: number, r: number): Promise<ActionResult> {
        try {
            if (!this.selectedUnitCoord) {
                return {
                    success: false,
                    message: 'No unit selected',
                    error: 'Must select a unit before moving'
                };
            }

            if (this.isProcessingMove) {
                return {
                    success: false,
                    message: 'Move already in progress',
                    error: 'Another move is being processed'
                };
            }

            // Find the move option for this target
            const moveOption = this.availableMovementOptions.find(opt => opt.q === q && opt.r === r);
            if (!moveOption) {
                return {
                    success: false,
                    message: `Cannot move to (${q}, ${r}) - not a valid move target`,
                    error: 'Invalid move target',
                    data: { 
                        selectedUnit: this.selectedUnitCoord,
                        availableMoves: this.availableMovementOptions.map(opt => ({q: opt.q, r: opt.r}))
                    }
                };
            }

            // Execute the move using existing logic
            const fromCoord = this.selectedUnitCoord;
            await this.executeMove(fromCoord, { q, r }, moveOption, false); // Skip validation for command interface

            return {
                success: true,
                message: `Unit moved from (${fromCoord.q}, ${fromCoord.r}) to (${q}, ${r})`,
                data: { from: fromCoord, to: { q, r } }
            };

        } catch (error) {
            return {
                success: false,
                message: `Failed to move unit to (${q}, ${r})`,
                error: error instanceof Error ? error.message : String(error)
            };
        }
    }

    /**
     * Attack with the currently selected unit
     */
    async attackWithSelectedUnit(q: number, r: number): Promise<ActionResult> {
        // TODO: Implement attack functionality
        // For now, return not implemented
        return {
            success: false,
            message: 'Attack functionality not yet implemented',
            error: 'Feature not implemented',
            data: { targetPosition: { q, r } }
        };
    }

    /**
     * End the current player's turn 
     * This is the unified method used by both UI clicks and command interface
     */
    async endCurrentPlayerTurn(): Promise<ActionResult> {
        try {
            const currentPlayer = this.gameState.getCurrentPlayer();
            const currentTurn = this.gameState.getTurnCounter();

            // Execute the turn end logic
            await this.gameState.endTurn(currentPlayer);
            
            // Update UI state
            const newPlayer = this.gameState.getCurrentPlayer();
            const newTurn = this.gameState.getTurnCounter();
            
            this.updateGameStatus(`Ready - Player ${newPlayer}'s Turn`, newPlayer);
            this.updateTurnCounter(newTurn);
            this.clearUnitSelection();
            
            if (this.gameLogPanel) {
                this.gameLogPanel.logGameEvent(`Player ${newPlayer}'s turn begins`, 'system');
            }
            this.showToast('Info', `Player ${newPlayer}'s turn`, 'info');

            return {
                success: true,
                message: `Turn ended. Now Player ${newPlayer}'s turn`,
                data: { 
                    previousPlayer: currentPlayer,
                    currentPlayer: newPlayer,
                    previousTurn: currentTurn,
                    currentTurn: newTurn
                }
            } as ActionResult;

        } catch (error) {
            const errorMsg = error instanceof Error ? error.message : String(error);
            this.showToast('Error', errorMsg, 'error');
            return {
                success: false,
                message: 'Failed to end turn',
                error: errorMsg
            } as ActionResult;
        }
    }

    /**
     * Get current game state information
     */
    async getGameState(): Promise<GameStateInfo> {
        try {
            const gameState = await this.gameState.getCurrentGameState();
            
            return {
                gameId: gameState.gameId || this.currentGameId || 'unknown',
                currentPlayer: gameState.currentPlayer,
                turnCounter: gameState.turnCounter,
                selectedUnit: this.selectedUnitCoord || undefined,
                unitsCount: Object.keys(this.world.units).length,
                tilesCount: Object.keys(this.world.tiles).length
            };

        } catch (error) {
            console.error('Failed to get game state:', error);
            return {
                gameId: this.currentGameId || 'unknown',
                currentPlayer: -1,
                turnCounter: -1,
                selectedUnit: undefined,
                unitsCount: 0,
                tilesCount: 0
            };
        }
    }

    /**
     * Clear current unit selection
     */
    clearSelection(): ActionResult {
        try {
            this.clearUnitSelection();
            this.clearAllHighlights();
            
            return {
                success: true,
                message: 'Selection cleared'
            };
        } catch (error) {
            return {
                success: false,
                message: 'Failed to clear selection',
                error: error instanceof Error ? error.message : String(error)
            };
        }
    }

    /**
     * Game action handlers - all synchronous for immediate UI feedback
     */

    private undoMove(): void {
        this.showToast('Info', 'Undo not yet implemented', 'info');
    }

    private selectMoveMode(): void {
        this.showToast('Info', 'Click on a highlighted tile to move', 'info');
    }

    private selectAttackMode(): void {
        this.showToast('Info', 'Click on a highlighted enemy to attack', 'info');
    }

    private showAllPlayerUnits(): void {
        // ✅ Use GameState metadata
        const currentPlayer = this.gameState.getCurrentPlayer();
        
        // TODO: Highlight all player units and center camera
        this.showToast('Info', `Showing all Player ${currentPlayer} units`, 'info');
    }

    private centerOnAction(): void {
        // TODO: Center camera on the most recent action or selected unit
        this.showToast('Info', 'Centering view', 'info');
    }

    /*
    private async moveUnit(fromQ: number, fromR: number, toQ: number, toR: number): Promise<void> {
        if (!this.gameState?.isReady()) {
            this.showToast('Error', 'Game not ready', 'error');
            return;
        }

        // Async WASM call
        await this.gameState.moveUnit(fromQ, fromR, toQ, toR);
        
        // Immediate UI feedback
        this.logGameEvent(`Unit moved from (${fromQ},${fromR}) to (${toQ},${toR})`);
        this.showToast('Success', 'Unit moved successfully', 'success');
        
        // Clear selection after successful move
        this.clearUnitSelection();
    }

    private async attackUnit(attackerQ: number, attackerR: number, defenderQ: number, defenderR: number): Promise<void> {
        if (!this.gameState?.isReady()) {
            this.showToast('Error', 'Game not ready', 'error');
            return;
        }

        // Async WASM call
        await this.gameState.attackUnit(attackerQ, attackerR, defenderQ, defenderR);
        
        // Immediate UI feedback
        this.logGameEvent(`Attack: (${attackerQ},${attackerR}) → (${defenderQ},${defenderR})`);
        this.showToast('Success', 'Attack completed', 'success');
        
        // Clear selection after attack
        this.clearUnitSelection();
    }
   */

    private clearUnitSelection(): void {
        this.selectedUnitCoord = null;
        this.availableMovementOptions = [];
        
        // Hide unit info via GameActionsPanel
        this.gameActionsPanel?.hideSelectedUnit();
    }

    /**
     * Clear all highlight layers
     */
    private clearAllHighlights(): void {
        if (this.gameScene) {
            const selectionLayer = this.gameScene.selectionHighlightLayer
            const movementLayer = this.gameScene.movementHighlightLayer
            const attackLayer = this.gameScene.attackHighlightLayer
            
            if (selectionLayer) {
                selectionLayer.clearSelection();
            }
            if (movementLayer) {
                movementLayer.clearMovementOptions();
            }
            if (attackLayer) {
                attackLayer.clearAttackOptions();
            }
        }
    }

    /**
     * Handle unit clicks - select unit or show unit info
     */
    private handleUnitClick(q: number, r: number): void {
        // Handle async unit interaction using unified getOptionsAt
        this.gameState.getOptionsAt(q, r).then(async (response: GetOptionsAtResponse) => {
            // ✅ Use shared World for fast unit query
            const unit = this.world?.getUnitAt(q, r);
            
            // Debug logging
            console.log(`[GameViewerPage] Unit click at (${q}, ${r}):`, {
                unit: unit,
                response: response,
                currentPlayer: this.gameState.getCurrentPlayer(),
                turnCounter: this.gameState.getTurnCounter()
            });
            
            const options = response.options || [];
            
            const hasMovementOptions = options.some(opt => opt.move !== undefined);
            const hasAttackOptions = options.some(opt => opt.attack !== undefined);
            const hasOnlyEndTurn = options.length === 1 && options[0].endTurn !== undefined;
            
            if (hasMovementOptions || hasAttackOptions) {
                // This unit has actionable options - process it directly (no duplicate RPC)
                this.processUnitSelection(q, r, options, response);
            } else if (hasOnlyEndTurn) {
                // This position only has endTurn option - could be empty tile, enemy unit, or friendly unit with no actions
                
                // ✅ Use shared World for fast queries
                const tileUnit = this.world?.getUnitAt(q, r);
                
                if (tileUnit) {
                    // Get current player to check ownership
                    this.gameState.getCurrentGameState().then(gameState => {
                        const currentPlayer = gameState.currentPlayer;
                        
                        console.log(`[GameViewerPage] Unit details:`, {
                            unitPlayer: tileUnit.player,
                            currentPlayer: currentPlayer,
                            distanceLeft: tileUnit.distanceLeft,
                            availableHealth: tileUnit.availableHealth,
                            turnCounter: tileUnit.turnCounter,
                            gameTurnCounter: gameState.turnCounter
                        });
                        
                        if (tileUnit.player === currentPlayer) {
                            // This is our unit but it has no available actions
                            this.showToast('Info', `No actions available for unit at (${q}, ${r})`, 'info');
                        } else {
                            // This is an enemy unit
                            this.showToast('Info', `Enemy unit at (${q}, ${r})`, 'info');
                        }
                    }).catch(error => {
                        console.error('Failed to get current game state:', error);
                    });
                } else {
                    this.showToast('Info', `Empty tile at (${q}, ${r})`, 'info');
                }
            }
        }).catch(error => {
            console.error('[GameViewerPage] Failed to get options at position:', error);
        });
    }

    /**
     * Handle movement clicks - execute actual unit moves
     */
    private handleMovementClick(q: number, r: number, moveOption: any): void {
        if (this.isProcessingMove) {
            console.warn('[GameViewerPage] Already processing a move, ignoring click');
            this.showToast('Warning', 'Move in progress...', 'warning');
            return;
        }

        if (!this.selectedUnitCoord) {
            console.warn('[GameViewerPage] No unit selected for movement');
            return;
        }

        // Check if clicking on the same position as the selected unit (deselection)
        if (this.selectedUnitCoord.q === q && this.selectedUnitCoord.r === r) {
            console.log('[GameViewerPage] Clicked on selected unit position - deselecting');
            this.clearSelection();
            this.clearAllHighlights();
            return;
        }

        // Execute the move
        this.executeMove(this.selectedUnitCoord, { q, r }, moveOption);
    }

    /**
     * Handle tile clicks - show terrain info in TerrainStatsPanel
     */
    private handleTileClick(q: number, r: number, tile: any): void {
        if (!this.terrainStatsPanel) {
            console.warn('[GameViewerPage] TerrainStatsPanel not available');
            return;
        }

        // Show terrain info using shared World
        if (tile) {
            const terrainStats = this.rulesTable.getTerrainStatsAt(tile.tileType, tile.player);
            if (terrainStats) {
                // Update with actual coordinates
                const terrainStatsWithCoords = new TerrainStats(
                    terrainStats.terrainDefinition, 
                    q, 
                    r, 
                    tile.player
                );
                this.terrainStatsPanel.updateTerrainStats(terrainStatsWithCoords);
            }
        }
    }


    /**
     * Process unit selection with unified options format
     */
    private processUnitSelection(q: number, r: number, options: GameOption[], fullResponse?: GetOptionsAtResponse): void {
        // Extract movement and attack options from the unified options
        // Note: protobuf oneof fields become direct properties (e.g., option.move, option.attack)
        const movementOptions = options.filter(opt => opt.move !== undefined);
        const attackOptions = options.filter(opt => opt.attack !== undefined);
        
        // Log to verify we're getting all options
        console.log(`[GameViewerPage] Processing selection: ${options.length} total, ${movementOptions.length} moves, ${attackOptions.length} attacks`);
        
        // Extract MoveOption and AttackOption objects from the unified options
        const moveOptionObjects = movementOptions.map(option => option.move!);
        const attackOptionObjects = attackOptions.map(option => option.attack!);
        
        // Store selected unit info and available options for move execution
        this.selectedUnitCoord = { q, r };
        this.availableMovementOptions = moveOptionObjects;
        
        // Show selected unit in GameActionsPanel
        const selectedUnit = this.world?.getUnitAt(q, r);
        if (selectedUnit) {
            this.gameActionsPanel?.showSelectedUnit(selectedUnit);
            // Pass the full response to TurnOptionsPanel to avoid duplicate RPC
            if (fullResponse) {
                this.turnOptionsPanel?.handleUnitSelectionWithOptions(q, r, selectedUnit, fullResponse);
            } else {
                // Fallback if no response provided (shouldn't happen)
                this.turnOptionsPanel?.handleUnitSelection(q, r, selectedUnit);
            }
        }
        
        // Update GameViewer to show highlights using layer-based approach  
        if (this.gameScene) {
            // Clear previous selection
            const selectionLayer = this.gameScene.selectionHighlightLayer;
            const movementLayer = this.gameScene.movementHighlightLayer;
            const attackLayer = this.gameScene.attackHighlightLayer;
            
            if (selectionLayer && movementLayer && attackLayer) {
                // Select the unit
                selectionLayer.selectHex(q, r);
                
                // Show movement options using protobuf MoveOption objects
                movementLayer.showMovementOptions(moveOptionObjects);
                
                // Show attack options (convert to coordinates for now)
                const attackCoords = attackOptionObjects.map(attackOpt => ({ q: attackOpt.q, r: attackOpt.r }));
                attackLayer.showAttackOptions(attackCoords);
            } else {
                console.warn('[GameViewerPage] Some highlight layers not available');
            }
        }

        this.showToast('Success', `Unit selected at (${q}, ${r}) - ${movementOptions.length} moves, ${attackOptions.length} attacks available`, 'success');
    }

    /**
     * Execute a unit move using ProcessMoves API
     */
    private async executeMove(fromCoord: { q: number, r: number }, toCoord: { q: number, r: number }, moveOption: MoveOption, validateStates=true): Promise<void> {
        // Set processing state to prevent concurrent moves
        this.isProcessingMove = true;
        this.showToast('Info', 'Processing move...', 'info');

        try {
            // Use the ready-to-use action from the moveOption
            if (!moveOption.action) {
                throw new Error('Move option does not contain action object');
            }

            // ✅ Get current player from move option or query WASM
            const currentGameState = await this.gameState!.getCurrentGameState();
            
            const gameMove= models.GameMove.from({
                player: currentGameState.currentPlayer,
                moveUnit: moveOption.action,
            })!;

            if (validateStates) { // as debug check our state with the server is in sync before making moves
              await this.ensureInSyncWithServer()
            }

            // ✅ Call ProcessMoves API - this will trigger World updates via EventBus
            const worldChanges = await this.gameState!.processMoves([gameMove]);
            if (validateStates) { // as debug check our state with the server is in sync before making moves
              // Add a small delay to ensure all async operations complete
              await new Promise(resolve => setTimeout(resolve, 100));
              await this.ensureInSyncWithServer()
            }

            // Clear selection and highlights after successful move
            this.clearUnitSelection();
            this.clearAllHighlights();

            // Show success feedback
            this.showToast('Success', `Unit moved to (${toCoord.q}, ${toCoord.r})`, 'success');

            // ✅ Update game UI with fresh state from WASM
            const updatedGameState = await this.gameState!.getCurrentGameState();
            this.updateGameUIFromState(updatedGameState);

        } catch (error) {
            console.error('[GameViewerPage] Move execution failed:', error);
            
            // Show error feedback
            const errorMessage = error instanceof Error ? error.message : 'Move failed';
            this.showToast('Error', `Move failed: ${errorMessage}`, 'error');
            
        } finally {
            // Always clear processing state
            this.isProcessingMove = false;
        }
    }

    protected async ensureInSyncWithServer() {
      const backendState = await this.gameState.getCurrentGameState();
      
      if (backendState.currentPlayer != this.gameState.currentPlayer) {
        throw new Error(`Backend State (${backendState.currentPlayer}) != this.currentPlayer (${this.gameState.currentPlayer})`)
      }
      if (backendState.turnCounter != this.gameState.turnCounter) {
        throw new Error(`Backend State (${backendState.turnCounter}) != this.turnCounter (${this.gameState.turnCounter})`)
      }

      // Enhanced world data validation
      const backendWorldData = backendState.worldData as WorldData;
      const backendTiles = backendWorldData.tiles || []
      const backendUnits = backendWorldData.units || []
      
      if (backendTiles.length != Object.keys(this.world.tiles).length) {
        throw new Error(`Backend Tile Count (${backendTiles.length}) != this.tileCount(${Object.keys(this.world.tiles).length})`)
      }
      if (backendUnits.length != Object.keys(this.world.units).length) {
        throw new Error(`Backend Unit Count (${backendUnits.length}) != this.unitCount(${Object.keys(this.world.units).length})`)
      }
      
      // Create coordinate map for backend units
      const backendUnitMap = new Map<string, Unit>();
      for (const unit of backendUnits) {
        const key = `${unit.q},${unit.r}`;
        backendUnitMap.set(key, unit);
      }
      
      // Validate each frontend unit by coordinate lookup
      for (const [coord, frontendUnit] of Object.entries(this.world.units)) {
        const backendUnit = backendUnitMap.get(coord);
        
        if (!backendUnit) {
          throw new Error(`Frontend unit at ${coord} not found in backend state`);
        }
        
        // Check all unit fields for mismatches
        const mismatches = [];
        if (frontendUnit.q != backendUnit.q) mismatches.push(`q: ${frontendUnit.q} != ${backendUnit.q}`);
        if (frontendUnit.r != backendUnit.r) mismatches.push(`r: ${frontendUnit.r} != ${backendUnit.r}`);
        if (frontendUnit.unitType != backendUnit.unitType) mismatches.push(`unitType: ${frontendUnit.unitType} != ${backendUnit.unitType}`);
        if (frontendUnit.player != backendUnit.player) mismatches.push(`player: ${frontendUnit.player} != ${backendUnit.player}`);
        if (frontendUnit.availableHealth != backendUnit.availableHealth) mismatches.push(`availableHealth: ${frontendUnit.availableHealth} != ${backendUnit.availableHealth}`);
        if (frontendUnit.distanceLeft != backendUnit.distanceLeft) mismatches.push(`distanceLeft: ${frontendUnit.distanceLeft} != ${backendUnit.distanceLeft}`);
        if (frontendUnit.turnCounter != backendUnit.turnCounter) mismatches.push(`turnCounter: ${frontendUnit.turnCounter} != ${backendUnit.turnCounter}`);
        
        if (mismatches.length > 0) {
          console.error(`[ensureInSyncWithServer] Unit at ${coord} mismatches:`, mismatches);
          console.error(`[ensureInSyncWithServer] Frontend unit:`, frontendUnit);
          console.error(`[ensureInSyncWithServer] Backend unit:`, backendUnit);
          throw new Error(`Backend unit at ${coord} does not match Frontend unit: ${mismatches.join(', ')}`);
        }
      }
      
      // Also check for backend units not in frontend
      for (const [coord, backendUnit] of backendUnitMap.entries()) {
        if (!this.world.units[coord]) {
          throw new Error(`Backend unit at ${coord} not found in frontend state`);
        }
      }
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
    private updateGameStatus(status: string, currentPlayer?: number): void {
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
        this.updateGameStatus(`Ready - Player ${gameState.currentPlayer}'s Turn`, gameState.currentPlayer);
        
        // Update turn counter
        this.updateTurnCounter(gameState.turnCounter);
        
        // Update game actions panel with current game state
        if (this.gameActionsPanel) {
            this.gameActionsPanel.updateGameStatus(gameState.currentPlayer, gameState.turnCounter);
        }
    }
    
    private updateTurnCounter(turnCounter: number): void {
        const turnElement = document.getElementById('turn-counter');
        if (turnElement) {
            turnElement.textContent = `Turn ${turnCounter}`;
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
