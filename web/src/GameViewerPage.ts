 import { BasePage } from '../lib/BasePage';
import { EventBus } from '../lib/EventBus';
import { GameViewer } from './GameViewer';
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
    private worldViewer: GameViewer
    private gameState: GameState
    private terrainStatsPanel: TerrainStatsPanel
    private rulesTable: RulesTable
    
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
        console.log('GameViewerPage: performLocalInit() - Phase 1');

        // Load game config first
        this.loadGameConfiguration();
        
        // Create shared RulesTable instance
        this.rulesTable = new RulesTable();
        
        // Subscribe to events BEFORE creating components
        this.subscribeToGameStateEvents();
        
        // Create child components
        this.createComponents();
        
        // Initialize basic UI state
        this.updateGameStatus('Game Loading...');
        this.initializeGameLog();
        
        // Start WASM and game data loading early (parallel to WorldViewer initialization)
        if (this.currentGameId) {
            console.log('GameViewerPage: Starting early WASM initialization for gameId:', this.currentGameId);
            this.initializeGameWithWASM().then(() => {
                console.log('GameViewerPage: WASM initialization completed');
                // Emit event to indicate game data is ready
                this.eventBus.emit(GameEventTypes.GAME_DATA_LOADED, { gameId: this.currentGameId }, this, this);
            }).catch(error => {
                console.error('GameViewerPage: WASM initialization failed:', error);
                this.updateGameStatus('WASM initialization failed');
            });
        }
        
        console.log('GameViewerPage: DOM initialized, returning child components');

        console.assert(this.worldViewer != null, "World viewer could not be created")
        console.assert(this.gameState != null, "gameState could not be created")
        console.assert(this.terrainStatsPanel != null, "terrainStatsPanel could not be created")
        console.assert(this.rulesTable != null, "rulesTable could not be created")
        
        // Return child components for lifecycle management
        return [
            this.worldViewer,
            this.gameState,
            this.terrainStatsPanel,
        ]
    }

    /**
     * Phase 3: Activate component when all dependencies are ready
     */
    async activate(): Promise<void> {
        console.log('GameViewerPage: activate() - Phase 3');
        
        // Bind events now that all components are ready
        this.bindGameSpecificEvents();
        
        console.log('GameViewerPage: activation complete');
    }

    public destroy(): void {
        if (this.worldViewer) {
            this.worldViewer.destroy();
            this.worldViewer = null as any;
        }
        
        if (this.gameState) {
            this.gameState.destroy();
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
     * Load basic game initialization data (just gameId from DOM)
     * Game configuration is now accessed directly from WASM-cached Game proto
     */
    private loadGameConfiguration(): void {
        // Get gameId from hidden input
        const gameIdInput = document.getElementById("gameIdInput") as HTMLInputElement | null;
        this.currentGameId = gameIdInput?.value.trim() || null;

        console.log('Game ID loaded:', this.currentGameId);
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
    private worldViewerReady = false;
    private gameDataReady = false;

    /**
     * Handle events from the EventBus
     */
    public handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
        switch(eventType) {
            case WorldEventTypes.WORLD_VIEWER_READY:
                console.log('GameViewerPage: GameViewer ready event received', data);
                
                this.worldViewerReady = true;
                
                // Set up interaction callbacks when viewer is ready
                if (this.worldViewer) {
                    console.log('GameViewerPage: Setting interaction callbacks after scene ready');
                    this.worldViewer.setInteractionCallbacks(
                        this.onTileClicked,
                        this.onUnitClicked
                    );
                    
                    // Set up movement callback for game-specific move execution
                    this.worldViewer.setMovementCallback(this.onMovementClicked);
                    
                    console.log('GameViewerPage: Interaction callbacks set after scene ready');
                }
                
                // Check if both viewer and game data are ready
                this.checkAndLoadWorldIntoViewer();
                break;
            
            case GameEventTypes.GAME_DATA_LOADED:
                console.log('GameViewerPage: Game data ready event received', data);
                
                this.gameDataReady = true;
                
                // Check if both viewer and game data are ready
                this.checkAndLoadWorldIntoViewer();
                break;
            
            case 'unit-moved':
                console.log('GameViewerPage: Unit moved notification', data);
                // Could trigger animations, sound effects, etc.
                break;
            
            case 'unit-attacked':
                console.log('GameViewerPage: Unit attacked notification', data);
                // Could trigger combat animations, sound effects, etc.
                break;
            
            case 'turn-ended':
                console.log('GameViewerPage: Turn ended notification', data);
                // Could trigger end-of-turn animations, notifications, etc.
                break;
            
            default:
                // Call parent implementation for unhandled events
                super.handleBusEvent(eventType, data, target, emitter);
        }
    }

    /**
     * Create WorldViewer and GameState component instances
     */
    private createComponents(): void {
        const worldViewerContainer = document.getElementById('phaser-viewer-container');
        if (!worldViewerContainer) {
            throw new Error('GameViewerPage: phaser-viewer-container not found');
        }
        // Pass element directly (not string ID) as per UI_DESIGN_PRINCIPLES.md
        this.worldViewer = new GameViewer(worldViewerContainer, this.eventBus, true);

        // Create GameState component (no specific container needed)
        const gameStateContainer = document.createElement('div');
        gameStateContainer.style.display = 'none'; // Hidden data component
        document.body.appendChild(gameStateContainer);
        this.gameState = new GameState(gameStateContainer, this.eventBus, true);
        console.log('GameViewerPage: GameState created:', this.gameState);

        // Create TerrainStatsPanel component
        const terrainStatsContainer = document.getElementById('terrain-stats-container');
        if (!terrainStatsContainer) {
            throw new Error('GameViewerPage: terrain-stats-container not found');
        }
        this.terrainStatsPanel = new TerrainStatsPanel(terrainStatsContainer, this.eventBus, true);
        console.log('GameViewerPage: TerrainStatsPanel created:', this.terrainStatsPanel);
    }

    /**
     * Check if both WorldViewer and game data are ready, then load world into viewer
     */
    private async checkAndLoadWorldIntoViewer(): Promise<void> {
        if (!this.worldViewerReady || !this.gameDataReady) {
            console.log('GameViewerPage: Waiting for both viewer and game data to be ready', {
                worldViewerReady: this.worldViewerReady,
                gameDataReady: this.gameDataReady
            });
            return;
        }

        console.log('GameViewerPage: Both WorldViewer and game data are ready, loading world into viewer');
        
        try {
            // Get the World object from GameState (GameState owns it now)
            const world = this.gameState.getWorld();
            const game = this.gameState.getGame();
            
            // Load world into viewer
            if (this.worldViewer && world) {
                await this.worldViewer.loadWorld(world);
                this.showToast('Success', `Game loaded: ${game.name || world.getName() || 'Untitled'}`, 'success');
                
                // Update UI with loaded game state
                const gameState = this.gameState.getGameState();
                this.updateGameUIFromState(gameState);
                this.logGameEvent(`Game loaded: ${gameState.gameId}`);
                
                console.log('GameViewerPage: World successfully loaded into viewer');
            } else {
                throw new Error('WorldViewer or World not available');
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
        
        console.log('Game initialized with WASM engine - data loaded into WASM singletons and World created');
    }

    /**
     * Bind page-specific events (required by BasePage)
     * This method is called by BasePage constructor, but we're using external LifecycleController
     * so we make this a no-op and handle event binding in LCMComponent.activate()
     */
    protected bindSpecificEvents(): void {
        console.log('GameViewerPage: bindSpecificEvents() called by BasePage - deferred to activate() phase');
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

        console.log('Ending current player\'s turn...');
        
        // Async WASM call
        await this.gameState.endTurn(this.gameState.getGameState().currentPlayer);
        
        // Get updated game state and update UI
        const gameState = this.gameState.getGameState();
        this.updateGameUIFromState(gameState);
        this.clearUnitSelection();
        
        this.logGameEvent(`Player ${gameState.currentPlayer}'s turn begins`);
        this.showToast('Info', `Player ${gameState.currentPlayer}'s turn`, 'info');
    }

    private undoMove(): void {
        console.log('Undo move requested');
        // TODO: Implement undo functionality with WASM
        this.showToast('Info', 'Undo not yet implemented', 'info');
    }

    private selectMoveMode(): void {
        if (!this.selectedUnit) {
            this.showToast('Warning', 'Select a unit first', 'warning');
            return;
        }
        console.log('Move mode selected for unit:', this.selectedUnit);
        // Movement options are already loaded from unit selection
        // TODO: Integrate with Phaser to highlight valid move tiles
        this.showToast('Info', 'Click on a highlighted tile to move', 'info');
    }

    private selectAttackMode(): void {
        if (!this.selectedUnit) {
            this.showToast('Warning', 'Select a unit first', 'warning');
            return;
        }
        console.log('Attack mode selected for unit:', this.selectedUnit);
        // Attack options are already loaded from unit selection
        // TODO: Integrate with Phaser to highlight valid attack targets
        this.showToast('Info', 'Click on a highlighted enemy to attack', 'info');
    }

    private showAllPlayerUnits(): void {
        if (!this.gameState?.isReady()) {
            return;
        }

        // Synchronous WASM call
        const gameData = this.gameState.getGameState();
        
        console.log(`Showing all units for Player ${gameData.currentPlayer}`);
        // TODO: Highlight all player units and center camera
        this.showToast('Info', `Showing all Player ${gameData.currentPlayer} units`, 'info');
    }

    private centerOnAction(): void {
        console.log('Centering on action');
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
        if (this.worldViewer) {
            const selectionLayer = this.worldViewer.getSelectionHighlightLayer();
            const movementLayer = this.worldViewer.getMovementHighlightLayer();
            const attackLayer = this.worldViewer.getAttackHighlightLayer();
            
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
     * Handle tile click from PhaserWorldScene - show terrain info in TerrainStatsPanel
     * @returns false to suppress event emission (we handle it completely)
     */
    private onTileClicked = (q: number, r: number): boolean => {
        console.log(`[GameViewerPage] Tile clicked callback: Q=${q}, R=${r}`);

        if (!this.gameState?.isReady()) {
            console.warn('[GameViewerPage] Game not ready for tile clicks');
            return false; // Suppress event emission
        }

        if (!this.terrainStatsPanel) {
            console.warn('[GameViewerPage] TerrainStatsPanel not available');
            return false; // Suppress event emission
        }

        // Get terrain info using RulesTable
        const world = this.gameState?.getWorld();
        if (world) {
            const tile = world.getTileAt(q, r);
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
        
        return false; // We handled it completely, suppress event emission
    };

    /**
     * Handle unit click from PhaserWorldScene - select unit or show unit info
     * @returns false to suppress event emission (we handle it completely)
     */
    private onUnitClicked = (q: number, r: number): boolean => {
        console.log(`[GameViewerPage] Unit clicked callback: Q=${q}, R=${r}`);

        if (!this.gameState?.isReady()) {
            console.warn('[GameViewerPage] Game not ready for unit clicks');
            return false; // Suppress event emission
        }

        // Handle async unit interaction using unified getOptionsAt
        this.gameState.getOptionsAt(q, r).then(response => {
            console.log(`[GameViewerPage] Options at (${q}, ${r}):`, response);
            
            const options = response.options || [];
            const hasMovementOptions = options.some((opt: any) => opt.move !== undefined);
            const hasAttackOptions = options.some((opt: any) => opt.attack !== undefined);
            const hasOnlyEndTurn = options.length === 1 && options[0].endTurn !== undefined;
            
            if (hasMovementOptions || hasAttackOptions) {
                // This unit has actionable options - select it
                this.selectUnitAt(q, r, options);
            } else if (hasOnlyEndTurn) {
                // This is either an empty tile or enemy unit - just show info
                console.log(`[GameViewerPage] Non-actionable position at Q=${q}, R=${r}`);
                
                // Get basic tile info to show details (async)
                this.gameState?.getTileInfo(q, r).then(tileInfo => {
                    if (tileInfo?.hasUnit) {
                        this.showToast('Info', `Enemy unit at (${q}, ${r})`, 'info');
                    } else {
                        this.showToast('Info', `Empty tile at (${q}, ${r})`, 'info');
                    }
                }).catch(error => {
                    console.error('Failed to get tile info:', error);
                });
            }
        }).catch(error => {
            console.error('[GameViewerPage] Failed to get options at position:', error);
        });
        
        return false; // Suppress event emission
    };

    /**
     * Handle movement clicks - execute actual unit moves
     * @returns false to suppress event emission (we handle it completely)
     */
    private onMovementClicked = (q: number, r: number, moveOption: MoveOption): void => {
        console.log(`[GameViewerPage] Movement clicked at (${q}, ${r}) with option:`, moveOption);

        if (!this.gameState?.isReady()) {
            console.warn('[GameViewerPage] Game not ready for movement');
            return;
        }

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
    };

    /**
     * Select unit and show movement/attack highlights
     */
    private selectUnitAt(q: number, r: number, options: any[]): void {
        console.log(`[GameViewerPage] Selecting unit at Q=${q}, R=${r} with ${options.length} options`);
        
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
        if (this.worldViewer) {
            // Clear previous selection
            const selectionLayer = this.worldViewer.getSelectionHighlightLayer();
            const movementLayer = this.worldViewer.getMovementHighlightLayer();
            const attackLayer = this.worldViewer.getAttackHighlightLayer();
            
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

            const gameMove= GameMove.from({
                player: this.gameState!.getGameState().currentPlayer,
                moveUnit: moveOption.action,
            })!;

            // Call ProcessMoves API
            const worldChanges = await this.gameState!.processMoves([gameMove]);
            
            console.log('[GameViewerPage] Move executed successfully, world changes:', worldChanges);

            // Clear selection and highlights after successful move
            this.clearUnitSelection();
            this.clearAllHighlights();

            // Show success feedback
            this.showToast('Success', `Unit moved to (${toCoord.q}, ${toCoord.r})`, 'success');

            // Update game UI with new state
            const gameState = this.gameState!.getGameState();
            this.updateGameUIFromState(gameState);

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
        const turnElement = document.getElementById('turn-counter');
        if (turnElement) {
            turnElement.textContent = `Turn ${gameState.turnCounter}`;
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
