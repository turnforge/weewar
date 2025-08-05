 import { BasePage } from '../lib/BasePage';
import { EventBus } from '../lib/EventBus';
import { PhaserGameScene } from './phaser/PhaserGameScene';
import { Unit, Tile, World } from './World';
import { GameState, UnitSelectionData } from './GameState';
import { GameState as ProtoGameState, Game as ProtoGame, GameConfiguration as ProtoGameConfiguration, MoveOption, AttackOption, GameMove } from '../gen/wasm-clients/weewar/v1/models';
import { WeewarV1Deserializer } from '../gen/wasm-clients/weewar/v1/deserializer';
import { create } from '@bufbuild/protobuf';
import { LCMComponent } from '../lib/LCMComponent';
import { LifecycleController } from '../lib/LifecycleController';
import { PLAYER_BG_COLORS } from './ColorsAndNames';
import { TerrainStatsPanel } from './TerrainStatsPanel';
import { GameEventTypes, WorldEventTypes } from './events';
import { RulesTable, TerrainStats } from './RulesTable';

/**
 * Game Viewer Page - Interactive game play interface
 * Responsible for:
 * - Loading world as a game instance
 * - Coordinating WASM game engine
 * - Managing game state and turn flow
 * - Handling player interactions (unit selection, movement, attacks)
 * - Providing game controls and UI feedback
 */
class GameViewerPage extends BasePage implements LCMComponent {
    private currentGameId: string | null;
    private gameScene: PhaserGameScene
    private gameState: GameState
    private world: World  // ✅ Shared World component
    private terrainStatsPanel: TerrainStatsPanel
    private rulesTable: RulesTable = new RulesTable();
    
    // Game configuration accessed directly from WASM-cached Game proto
    
    // UI state
    private selectedUnit: any = null;
    private gameLog: string[] = [];
    
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
        
        // Subscribe to events BEFORE creating components
        this.subscribeToGameStateEvents();
        
        // Create child components
        this.createComponents();
        
        // Initialize basic UI state
        this.updateGameStatus('Game Loading...');
        this.initializeGameLog();
        
        // Start WASM and game data loading early (parallel to WorldViewer initialization)
        if (!this.currentGameId) {
          throw new Error("Game Id Not Found")
        }
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
        console.assert(this.rulesTable != null, "rulesTable could not be created")
        
        // Return child components for lifecycle management
        // Note: World and GameState don't extend BaseComponent, so not included in lifecycle
        return [
            this.gameScene,
            this.terrainStatsPanel,
        ]
    }

    /**
     * Phase 2: Inject dependencies
     */
    setupDependencies(): void {
        // Set up scene click callback now that gameScene is initialized
        this.gameScene.sceneClickedCallback = this.onSceneClicked;
    }

    /**
     * Phase 3: Activate component when all dependencies are ready
     */
    async activate(): Promise<void> {
        // Bind events now that all components are ready
        this.bindGameSpecificEvents();
    }

    public destroy(): void {
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

        if (this.rulesTable) {
            this.rulesTable = null as any;
        }
        
        this.currentGameId = null;
        this.selectedUnit = null;
        this.gameLog = [];
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

    // State tracking for initialization
    private gameSceneReady = false;
    private gameDataReady = false;

    /**
     * Handle events from the EventBus
     */
    public handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
        switch(eventType) {
            case WorldEventTypes.WORLD_VIEWER_READY:
                this.gameSceneReady = true;
                
                // Set up unified map callback when viewer is ready
                this.gameScene.sceneClickedCallback = this.onSceneClicked
                
                // Check if both viewer and game data are ready
                this.checkAndLoadWorldIntoViewer();
                break;
            
            case GameEventTypes.GAME_DATA_LOADED:
                this.gameDataReady = true;
                
                // Check if both viewer and game data are ready
                this.checkAndLoadWorldIntoViewer();
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

        const gameViewerContainer = document.getElementById('phaser-viewer-container');
        if (!gameViewerContainer) {
            throw new Error('GameViewerPage: phaser-viewer-container not found');
        }
        // Create PhaserGameScene as LCMComponent with container element
        this.gameScene = new PhaserGameScene(gameViewerContainer, this.eventBus, true);

        // ✅ Create GameState with direct EventBus connection (no DOM element needed)
        this.gameState = new GameState(this.eventBus);

        // Create TerrainStatsPanel component
        const terrainStatsContainer = document.getElementById('terrain-stats-container');
        if (!terrainStatsContainer) {
            throw new Error('GameViewerPage: terrain-stats-container not found');
        }
        this.terrainStatsPanel = new TerrainStatsPanel(terrainStatsContainer, this.eventBus, true);
    }

    /**
     * Check if both WorldViewer and game data are ready, then load world into viewer
     */
    private async checkAndLoadWorldIntoViewer(): Promise<void> {
        if (!this.gameSceneReady || !this.gameDataReady) {
            console.warn('GameViewerPage: Waiting for both viewer and game data to be ready', {
                gameSceneReady: this.gameSceneReady,
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
                
                // Update UI with loaded game state
                const gameState = await this.gameState.getCurrentGameState();
                this.updateGameUIFromState(gameState);
                this.logGameEvent(`Game loaded: ${gameState.gameId}`);
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

        // Wait for WASM to be ready (only async part)
        await this.gameState.waitUntilReady();
        
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
    protected bindSpecificEvents(): void {
    }

    /**
     * Internal method to bind game-specific events (called from activate() phase)
     */
    private bindGameSpecificEvents(): void {
        // End Turn button
        const endTurnBtn = document.getElementById('end-turn-btn');
        if (endTurnBtn) {
            endTurnBtn.addEventListener('click', this.endTurn.bind(this));
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



    /**
     * Game action handlers - all synchronous for immediate UI feedback
     */
    private async endTurn(): Promise<void> {
        if (!this.gameState?.isReady()) {
            this.showToast('Error', 'Game not ready', 'error');
            return;
        }
        
        // ✅ Use GameState metadata
        const currentPlayer = this.gameState.getCurrentPlayer();
        await this.gameState.endTurn(currentPlayer);
        
        // ✅ Use GameState metadata for UI updates
        const newPlayer = this.gameState.getCurrentPlayer();
        const turnCounter = this.gameState.getTurnCounter();
        
        this.updateGameStatus(`Ready - Player ${newPlayer}'s Turn`, newPlayer);
        this.updateTurnCounter(turnCounter);
        this.clearUnitSelection();
        
        this.logGameEvent(`Player ${newPlayer}'s turn begins`);
        this.showToast('Info', `Player ${newPlayer}'s turn`, 'info');
    }

    private undoMove(): void {
        this.showToast('Info', 'Undo not yet implemented', 'info');
    }

    private selectMoveMode(): void {
        if (!this.selectedUnit) {
            this.showToast('Warning', 'Select a unit first', 'warning');
            return;
        }
        // TODO: Integrate with Phaser to highlight valid move tiles
        this.showToast('Info', 'Click on a highlighted tile to move', 'info');
    }

    private selectAttackMode(): void {
        if (!this.selectedUnit) {
            this.showToast('Warning', 'Select a unit first', 'warning');
            return;
        }
        // Attack options are already loaded from unit selection
        // TODO: Integrate with Phaser to highlight valid attack targets
        this.showToast('Info', 'Click on a highlighted enemy to attack', 'info');
    }

    private showAllPlayerUnits(): void {
        if (!this.gameState?.isReady()) {
            return;
        }

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
        this.selectedUnit = null;
        this.selectedUnitCoord = null;
        this.availableMovementOptions = [];
        
        // Hide unit info panel
        const unitInfoPanel = document.getElementById('selected-unit-info');
        if (unitInfoPanel) {
            unitInfoPanel.classList.add('hidden');
        }
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
     * Callback methods for Phaser scene interactions
     */

    /**
     * Unified map click handler - handles all clicks with context about what was clicked
     */
    private onSceneClicked = (context: any, layer: string, extra?: any): void => {
        if (!this.gameState?.isReady()) {
            console.warn('[GameViewerPage] Game not ready for map clicks');
            return;
        }

        const { hexQ: q, hexR: r } = context;
        
        // Get tile and unit data from World using coordinates
        const tile = this.world?.getTileAt(q, r);
        const unit = this.world?.getUnitAt(q, r);

        console.log(`[GameViewerPage] Map clicked at (${q}, ${r}) on layer '${layer}'`, { tile, unit, extra });

        switch (layer) {
            case 'movement-highlight':
                // Get moveOption from the layer itself
                let moveOption = null;
                if (this.gameScene && 'getMovementHighlightLayer' in this.gameScene) {
                    const movementLayer = (this.gameScene as any).getMovementHighlightLayer();
                    moveOption = movementLayer?.getMoveOptionAt(q, r);
                }
                this.handleMovementClick(q, r, moveOption);
                break;
                
            case 'base-map':
                if (unit) {
                    this.handleUnitClick(q, r);
                } else {
                    this.handleTileClick(q, r, tile);
                }
                break;
                
            default:
                console.log(`[GameViewerPage] Unhandled layer click: ${layer}`);
        }
    };

    /**
     * Handle unit clicks - select unit or show unit info
     */
    private handleUnitClick(q: number, r: number): void {
        // Handle async unit interaction using unified getOptionsAt
        this.gameState.getOptionsAt(q, r).then(async response => {
            // ✅ Use shared World for fast unit query
            const unit = this.world?.getUnitAt(q, r);
            
            const options = response.options || [];
            
            const hasMovementOptions = options.some((opt: any) => opt.move !== undefined);
            const hasAttackOptions = options.some((opt: any) => opt.attack !== undefined);
            const hasOnlyEndTurn = options.length === 1 && options[0].endTurn !== undefined;
            
            if (hasMovementOptions || hasAttackOptions) {
                // This unit has actionable options - select it
                this.selectUnitAt(q, r, options);
            } else if (hasOnlyEndTurn) {
                // This position only has endTurn option - could be empty tile, enemy unit, or friendly unit with no actions
                
                // ✅ Use shared World for fast queries
                const tileUnit = this.world?.getUnitAt(q, r);
                
                if (tileUnit) {
                    // Get current player to check ownership
                    this.gameState.getCurrentGameState().then(gameState => {
                        const currentPlayer = gameState.currentPlayer;
                        
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
     * Select unit and show movement/attack highlights
     */
    private selectUnitAt(q: number, r: number, options: any[]): void {
        // console.log(`[GameViewerPage] Selecting unit at Q=${q}, R=${r} with ${options.length} options`);
        
        if (!this.gameState?.isReady()) {
            console.warn('[GameViewerPage] Game not ready for unit selection');
            return;
        }

        // Process the provided options (from getOptionsAt)
        this.processUnitSelection(q, r, options);
    }

    /**
     * Process unit selection with unified options format
     */
    private processUnitSelection(q: number, r: number, options: any[]): void {
        // Extract movement and attack options from the unified options
        // Note: protobuf oneof fields become direct properties (e.g., option.move, option.attack)
        const movementOptions = options.filter(opt => opt.move !== undefined);
        const attackOptions = options.filter(opt => opt.attack !== undefined);
        
        console.log(`[GameViewerPage] Unit selected: ${movementOptions.length} moves, ${attackOptions.length} attacks available`);
        
        // Extract MoveOption and AttackOption objects from the unified options
        const moveOptionObjects = movementOptions.map((option: any) => option.move);
        const attackOptionObjects = attackOptions.map((option: any) => option.attack);
        
        // Store selected unit info and available options for move execution
        this.selectedUnitCoord = { q, r };
        this.availableMovementOptions = moveOptionObjects;
        
        console.log('[GameViewerPage] Extracted protobuf objects:', {
            moveOptionObjects,
            attackOptionObjects,
            sampleMoveOption: moveOptionObjects[0],
            sampleAttackOption: attackOptionObjects[0]
        });

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
                
                console.log('[GameViewerPage] Highlights sent to layers');
            } else {
                console.warn('[GameViewerPage] Some highlight layers not available');
            }
        }

        this.showToast('Success', `Unit selected at (${q}, ${r}) - ${movementOptions.length} moves, ${attackOptions.length} attacks available`, 'success');
    }

    /**
     * Execute a unit move using ProcessMoves API
     */
    private async executeMove(fromCoord: { q: number, r: number }, toCoord: { q: number, r: number }, moveOption: MoveOption): Promise<void> {
        console.log(`[GameViewerPage] Executing move from (${fromCoord.q}, ${fromCoord.r}) to (${toCoord.q}, ${toCoord.r})`);

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
            
            const gameMove= GameMove.from({
                player: currentGameState.currentPlayer,
                moveUnit: moveOption.action,
            })!;

            // ✅ Call ProcessMoves API - this will trigger World updates via EventBus
            const worldChanges = await this.gameState!.processMoves([gameMove]);
            
            console.log('[GameViewerPage] Move executed successfully, world changes:', worldChanges);

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
    }
    
    private updateTurnCounter(turnCounter: number): void {
        const turnElement = document.getElementById('turn-counter');
        if (turnElement) {
            turnElement.textContent = `Turn ${turnCounter}`;
        }
    }

    private updateSelectedUnitInfo(unit: any): void {
        const unitDetails = document.getElementById('unit-details');
        if (unitDetails) {
            unitDetails.innerHTML = `
                <div><strong>Type:</strong> ${unit.type || 'Unknown'}</div>
                <div><strong>Position:</strong> (${unit.q}, ${unit.r})</div>
                <div><strong>Player:</strong> ${unit.playerId || 'Unknown'}</div>
                <div><strong>Health:</strong> ${unit.health || 'Unknown'}</div>
            `;
        }
    }

    private initializeGameLog(): void {
        this.gameLog = [];
    }

    private logGameEvent(message: string): void {
        this.gameLog.push(message);
        
        // Update game log UI
        const gameLogElement = document.getElementById('game-log');
        if (gameLogElement) {
            const logEntry = document.createElement('div');
            logEntry.textContent = message;
            logEntry.className = 'text-xs text-gray-600 dark:text-gray-300';
            
            gameLogElement.appendChild(logEntry);
            
            // Keep only last 20 entries
            if (gameLogElement.children.length > 20) {
                gameLogElement.removeChild(gameLogElement.firstChild!);
            }
            
            // Scroll to bottom
            gameLogElement.scrollTop = gameLogElement.scrollHeight;
        }
    }
}

// Initialize page when DOM is ready using LifecycleController
document.addEventListener('DOMContentLoaded', async () => {
    // Create page instance (just basic setup)
    const gameViewerPage = new GameViewerPage("GameViewerPage");
    
    // Create lifecycle controller with debug logging
    const lifecycleController = new LifecycleController(gameViewerPage.eventBus, LifecycleController.DefaultConfig);
    
    // Start breadth-first initialization
    await lifecycleController.initializeFromRoot(gameViewerPage);
});
