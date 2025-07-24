import { BasePage } from './BasePage';
import { EventBus, EventTypes } from './EventBus';
import { GameViewer } from './GameViewer';
import { Unit, Tile, World } from './World';
import { GameState, GameCreateData, UnitSelectionData } from './GameState';
import { ComponentLifecycle } from './ComponentLifecycle';
import { LifecycleController } from './LifecycleController';
import { PLAYER_BG_COLORS } from './ColorsAndNames';
import { TerrainStatsPanel } from './TerrainStatsPanel';

/**
 * Game Viewer Page - Interactive game play interface
 * Responsible for:
 * - Loading world as a game instance
 * - Coordinating WASM game engine
 * - Managing game state and turn flow
 * - Handling player interactions (unit selection, movement, attacks)
 * - Providing game controls and UI feedback
 */
class GameViewerPage extends BasePage implements ComponentLifecycle {
    private currentWorldId: string | null;
    private world: World | null = null;
    private worldViewer: GameViewer | null = null;
    private gameState: GameState | null = null;
    private terrainStatsPanel: TerrainStatsPanel | null = null;
    
    // Game configuration from URL parameters
    private playerCount: number = 2;
    private maxTurns: number = 0;
    private gameConfig: GameConfiguration;
    
    // UI state
    private selectedUnit: any = null;
    private gameLog: string[] = [];

    constructor() {
        console.log('GameViewerPage: Constructor starting...'); 
        super(); // BasePage will call initializeSpecificComponents() and bindSpecificEvents()
        console.log('GameViewerPage: Constructor completed - lifecycle will be managed externally');
    }

    /**
     * Load game configuration from URL parameters and hidden inputs
     */
    private loadGameConfiguration(): void {
        // Get worldId from hidden input
        
        // Initialize gameConfig before calling super() to ensure it's available in initializeSpecificComponents()
        this.gameConfig = this.gameConfig || {
            playerCount: 2,
            maxTurns: 0,
            unitRestrictions: {},
            playerTypes: {},
            playerTeams: {}
        };
        const worldIdInput = document.getElementById("worldIdInput") as HTMLInputElement | null;
        this.currentWorldId = worldIdInput?.value.trim() || null;

        // Get basic config from hidden inputs
        const playerCountInput = document.getElementById("playerCount") as HTMLInputElement | null;
        const maxTurnsInput = document.getElementById("maxTurns") as HTMLInputElement | null;

        this.gameConfig.playerCount = parseInt(playerCountInput?.value || '2');
        this.gameConfig.maxTurns = parseInt(maxTurnsInput?.value || '0');

        // Parse URL parameters for detailed configuration
        const urlParams = new URLSearchParams(window.location.search);
        
        // Parse unit restrictions
        for (const [key, value] of urlParams.entries()) {
            if (key.startsWith('unit_') && value === 'allowed') {
                const unitId = key.substring(5);
                this.gameConfig.unitRestrictions[unitId] = 'allowed';
            } else if (key.startsWith('player_') && key.includes('_type')) {
                const playerId = key.split('_')[1];
                this.gameConfig.playerTypes[playerId] = value;
            } else if (key.startsWith('player_') && key.includes('_team')) {
                const playerId = key.split('_')[1];
                this.gameConfig.playerTeams[playerId] = parseInt(value);
            }
        }

        console.log('Game configuration loaded:', this.gameConfig);
    }

    /**
     * Initialize page-specific components (required by BasePage)
     * This method is called by BasePage constructor, but we're using external LifecycleController
     * so we make this a no-op and handle initialization through ComponentLifecycle interface
     */
    protected initializeSpecificComponents(): void {
        console.log('GameViewerPage: initializeSpecificComponents() called by BasePage - doing minimal setup');
        this.loadGameConfiguration(); // Load game config here since constructor calls this
        console.log('GameViewerPage: Actual component initialization will be handled by LifecycleController');
    }

    /**
     * Subscribe to WorldViewer and GameState events before component creation
     */
    private subscribeToWorldViewerEvents(): void {
        // Subscribe BEFORE creating WorldViewer to catch initialization events
        this.eventBus.subscribe('world-viewer-ready', (payload) => {
            console.log('GameViewerPage: WorldViewer ready event received', payload);
            
            // Now that WorldViewer scene is ready, set up the interaction callbacks
            if (this.worldViewer) {
                console.log('GameViewerPage: Setting interaction callbacks after scene ready');
                this.worldViewer.setInteractionCallbacks(
                    this.onTileClicked,
                    this.onUnitClicked
                );
                console.log('GameViewerPage: Interaction callbacks set after scene ready');
            }
            
            if (this.currentWorldId) {
                console.log('GameViewerPage: WorldId found, proceeding to load world:', this.currentWorldId);
                // WebGL context timing - wait for next event loop tick
                setTimeout(async () => {
                    console.log('GameViewerPage: Starting loadWorldAndInitializeGame...');
                    await this.loadWorldAndInitializeGame();
                }, 10);
            } else {
                console.warn('GameViewerPage: No currentWorldId found!');
            }
        }, 'game-viewer-page');

        // GameState notification events (for system coordination, not user interaction responses)
        this.eventBus.subscribe('wasm-loaded', (payload) => {
            console.log('GameViewerPage: WASM loaded successfully');
        }, 'game-viewer-page');

        this.eventBus.subscribe('game-created', (payload) => {
            const gameData: GameCreateData = payload.data;
            console.log('GameViewerPage: Game created notification', gameData);
            // Game UI already updated synchronously, this is just for logging/coordination
        }, 'game-viewer-page');

        this.eventBus.subscribe('unit-moved', (payload) => {
            console.log('GameViewerPage: Unit moved notification', payload.data);
            // Could trigger animations, sound effects, etc.
        }, 'game-viewer-page');

        this.eventBus.subscribe('unit-attacked', (payload) => {
            console.log('GameViewerPage: Unit attacked notification', payload.data);
            // Could trigger combat animations, sound effects, etc.
        }, 'game-viewer-page');

        this.eventBus.subscribe('turn-ended', (payload) => {
            const gameData: GameCreateData = payload.data;
            console.log('GameViewerPage: Turn ended notification', gameData);
            // Could trigger end-of-turn animations, notifications, etc.
        }, 'game-viewer-page');
    }

    /**
     * Create WorldViewer and GameState component instances
     */
    private createWorldViewerComponent(): void {
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
        const terrainStatsContainer = document.getElementById('terrain-stats-panel');
        if (!terrainStatsContainer) {
            throw new Error('GameViewerPage: terrain-stats-panel container not found');
        }
        this.terrainStatsPanel = new TerrainStatsPanel(terrainStatsContainer, this.eventBus, true);
        console.log('GameViewerPage: TerrainStatsPanel created:', this.terrainStatsPanel);
    }

    /**
     * Load world data and initialize game
     */
    private async loadWorldAndInitializeGame(): Promise<void> {
        try {
            console.log('Loading world and initializing game...');

            // Load world data from hidden element
            const worldData = this.loadWorldDataFromElement();
            if (!worldData) {
                throw new Error('No world data found');
            }

            // Deserialize world
            this.world = World.deserialize(worldData);
            
            // Load world into viewer
            if (this.worldViewer) {
                await this.worldViewer.loadWorld(worldData);
                this.showToast('Success', `Game loaded: ${this.world.getName() || 'Untitled'}`, 'success');
            }

            // Initialize game using WASM
            console.log('About to initialize game with WASM...');
            try {
                await this.initializeGameWithWASM();
                console.log('WASM initialization completed successfully');
            } catch (error) {
                console.error('WASM initialization failed, but continuing with world display:', error);
                // Continue without WASM for now - still show the map
                this.updateGameStatus('Map loaded - WASM initialization failed');
            }
            
            // Update UI will be handled by GameState events

        } catch (error) {
            console.error('Failed to load world and initialize game:', error);
            this.showToast('Error', 'Failed to load game', 'error');
        }
    }

    /**
     * Initialize game using WASM game engine
     */
    private async initializeGameWithWASM(): Promise<void> {
        if (!this.gameState) {
            throw new Error('GameState component not initialized');
        }

        // Wait for WASM to be ready (only async part)
        await this.gameState.waitUntilReady();
        
        // Create game from world data via WASM (synchronous once WASM loaded)
        const worldDataStr = JSON.stringify(this.loadWorldDataFromElement());
        const gameData = await this.gameState.createGameFromMap(worldDataStr);
        
        // Debug: Check what player the game started with vs what we expect
        console.log('[GameViewerPage] Game creation debug:', {
            gameDataCurrentPlayer: gameData.currentPlayer,
            worldData: this.loadWorldDataFromElement(),
            // Show units in world data to see their player ownership
            units: this.loadWorldDataFromElement()?.units || []
        });
        
        
        // Update UI synchronously
        this.updateGameUIFromState(gameData);
        this.logGameEvent(`Game started with ${this.gameConfig.playerCount} players`);
        console.log('Game initialized with WASM engine');
    }

    /**
     * Bind page-specific events (required by BasePage)
     * This method is called by BasePage constructor, but we're using external LifecycleController
     * so we make this a no-op and handle event binding in ComponentLifecycle.activate()
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
     * Load world data from hidden element
     */
    private loadWorldDataFromElement(): any {
        try {
            const worldDataElement = document.getElementById('world-data-json');
            console.log('GameViewerPage: Found world-data-json element:', !!worldDataElement);
            
            if (worldDataElement) {
                const rawContent = worldDataElement.textContent;
                console.log('GameViewerPage: Raw world data content:', JSON.parse(rawContent || "{}")) //  ? rawContent.substring(0, 200) + '...' : 'null/empty');
                
                if (rawContent && rawContent.trim() !== '' && rawContent.trim() !== 'null') {
                    const worldData = JSON.parse(rawContent);
                    console.log('GameViewerPage: Parsed world data successfully:', !!worldData);
                    return worldData && worldData !== null ? worldData : null;
                } else {
                    console.warn('GameViewerPage: World data element is empty or contains null');
                }
            } else {
                console.error('GameViewerPage: world-data-json element not found in DOM');
            }
            return null;
        } catch (error) {
            console.error('GameViewerPage: Error parsing world data:', error);
            return null;
        }
    }

    /**
     * Game action handlers - all synchronous for immediate UI feedback
     */
    private endTurn(): void {
        if (!this.gameState?.isReady()) {
            this.showToast('Error', 'Game not ready', 'error');
            return;
        }

        try {
            console.log('Ending current player\'s turn...');
            
            // Synchronous WASM call
            const gameData = this.gameState.endTurn();
            
            // Immediate UI update
            this.updateGameUIFromState(gameData);
            this.clearUnitSelection();
            
            this.logGameEvent(`Player ${gameData.currentPlayer}'s turn begins`);
            this.showToast('Info', `Player ${gameData.currentPlayer}'s turn`, 'info');
            
        } catch (error) {
            console.error('Failed to end turn:', error);
            this.showToast('Error', 'Failed to end turn', 'error');
        }
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

        try {
            // Synchronous WASM call
            const gameData = this.gameState.getGameState();
            
            console.log(`Showing all units for Player ${gameData.currentPlayer}`);
            // TODO: Highlight all player units and center camera
            this.showToast('Info', `Showing all Player ${gameData.currentPlayer} units`, 'info');
            
        } catch (error) {
            console.error('Failed to get game state:', error);
        }
    }

    private centerOnAction(): void {
        console.log('Centering on action');
        // TODO: Center camera on the most recent action or selected unit
        this.showToast('Info', 'Centering view', 'info');
    }

    private handleUnitSelection(selectionData: UnitSelectionData): void {
        this.selectedUnit = selectionData.unit;
        console.log('Unit selected:', selectionData);
        
        // Update selected unit info panel
        this.updateSelectedUnitInfo(selectionData.unit);
        
        // Show unit action buttons
        const unitInfoPanel = document.getElementById('selected-unit-info');
        if (unitInfoPanel) {
            unitInfoPanel.classList.remove('hidden');
        }

        // TODO: Highlight movement and attack options on the map
        console.log('Movement options:', selectionData.movableCoords);
        console.log('Attack options:', selectionData.attackableCoords);
    }

    private moveUnit(fromQ: number, fromR: number, toQ: number, toR: number): void {
        if (!this.gameState?.isReady()) {
            this.showToast('Error', 'Game not ready', 'error');
            return;
        }

        try {
            // Synchronous WASM call
            const result = this.gameState.moveUnit(fromQ, fromR, toQ, toR);
            
            // Immediate UI feedback
            this.logGameEvent(`Unit moved from (${fromQ},${fromR}) to (${toQ},${toR})`);
            this.showToast('Success', 'Unit moved successfully', 'success');
            
            // Clear selection after successful move
            this.clearUnitSelection();
            
        } catch (error) {
            console.error('Failed to move unit:', error);
            const errorMessage = error instanceof Error ? error.message : 'Failed to move unit';
            this.showToast('Error', errorMessage, 'error');
        }
    }

    private attackUnit(attackerQ: number, attackerR: number, defenderQ: number, defenderR: number): void {
        if (!this.gameState?.isReady()) {
            this.showToast('Error', 'Game not ready', 'error');
            return;
        }

        try {
            // Synchronous WASM call
            const result = this.gameState.attackUnit(attackerQ, attackerR, defenderQ, defenderR);
            
            // Immediate UI feedback
            this.logGameEvent(`Attack: (${attackerQ},${attackerR}) → (${defenderQ},${defenderR})`);
            this.showToast('Success', 'Attack completed', 'success');
            
            // Clear selection after attack
            this.clearUnitSelection();
            
        } catch (error) {
            console.error('Failed to attack:', error);
            const errorMessage = error instanceof Error ? error.message : 'Failed to attack';
            this.showToast('Error', errorMessage, 'error');
        }
    }

    private clearUnitSelection(): void {
        this.selectedUnit = null;
        
        // Hide unit info panel
        const unitInfoPanel = document.getElementById('selected-unit-info');
        if (unitInfoPanel) {
            unitInfoPanel.classList.add('hidden');
        }
        
        // TODO: Clear visual selection highlights
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

        try {
            // Get terrain info from WASM via ui.go
            const terrainStats = this.gameState.getTerrainStatsAt(q, r);
            
            console.log('[GameViewerPage] Retrieved terrain stats:', terrainStats);
            
            // Update terrain stats panel with the data
            this.terrainStatsPanel.updateTerrainInfo({
                name: terrainStats.name || 'Unknown Terrain',
                tileType: terrainStats.tileType || 0,
                movementCost: terrainStats.movementCost || 1.0,
                defenseBonus: terrainStats.defenseBonus || 0.0,
                description: terrainStats.description || 'No description available',
                q: q,
                r: r,
                player: terrainStats.player
            });

            return true; // We handled it completely, suppress event emission

        } catch (error) {
            console.error('[GameViewerPage] Failed to get terrain stats:', error);
            // Clear terrain panel on error
            this.terrainStatsPanel.clearTerrainInfo();
            return false; // Suppress event emission on error
        }
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

        try {
            // Check if this unit can be selected by current player
            const canSelect = this.gameState.canSelectUnit(q, r);
            
            // Debug info to understand why selection might be failing
            console.log(`[GameViewerPage] Unit selection debug:`, {
                position: `(${q}, ${r})`,
                canSelect: canSelect,
                currentPlayer: (this.gameState as any).gameData?.currentPlayer,
                gameInitialized: (this.gameState as any).gameData?.gameInitialized
            });
            
            if (canSelect) {
                // This is a selectable unit - use existing selection logic
                this.selectUnitAt(q, r);
            } else {
                // This is an enemy or non-selectable unit - just show info
                console.log(`[GameViewerPage] Non-selectable unit at Q=${q}, R=${r}`);
                
                // Get basic tile info to show enemy unit details
                const tileInfo = this.gameState.getTileInfo(q, r);
                console.log('[GameViewerPage] Enemy unit tile info:', tileInfo);
                
                // For now, just show a message - could extend to show unit details panel later
                this.showToast('Info', `Enemy unit at (${q}, ${r})`, 'info');
            }

            return true; // We handled it completely, suppress event emission

        } catch (error) {
            console.error('[GameViewerPage] Failed to handle unit click:', error);
            return false; // Suppress event emission on error
        }
    };

    /**
     * Select unit and show movement/attack highlights
     */
    private selectUnitAt(q: number, r: number): void {
        console.log(`[GameViewerPage] Selecting unit at Q=${q}, R=${r}`);
        
        if (!this.gameState?.isReady()) {
            console.warn('[GameViewerPage] Game not ready for unit selection');
            return;
        }

        try {
            // Get movement options from WASM
            const movementResult = this.gameState.getMovementOptions(q, r);
            console.log('[GameViewerPage] Movement options:', movementResult);
            
            // Get attack options from WASM  
            const attackResult = this.gameState.getAttackOptions(q, r);
            console.log('[GameViewerPage] Attack options:', attackResult);
            
            // Get unit info
            const unitInfo = this.gameState.getTileInfo(q, r);
            console.log('[GameViewerPage] Unit info:', unitInfo);
            
            // Convert WASM results to coordinate arrays
            const movableCoords = movementResult.success && movementResult.data ? 
                movementResult.data.map((pos: any) => ({ q: pos.q, r: pos.r })) : [];
            
            const attackableCoords = attackResult.success && attackResult.data ?
                attackResult.data.map((pos: any) => ({ q: pos.q, r: pos.r })) : [];
            
            console.log(`[GameViewerPage] Unit selection: ${movableCoords.length} movement options, ${attackableCoords.length} attack options`);
            
            // Update GameViewer to show highlights using layer-based approach
            if (this.worldViewer) {
                // Clear previous selection
                const selectionLayer = this.worldViewer.getSelectionHighlightLayer();
                const movementLayer = this.worldViewer.getMovementHighlightLayer();
                const attackLayer = this.worldViewer.getAttackHighlightLayer();
                
                if (selectionLayer && movementLayer && attackLayer) {
                    // Select the unit
                    selectionLayer.selectHex(q, r);
                    
                    // Show movement options
                    movementLayer.showMovementOptions(movableCoords);
                    
                    // Show attack options
                    attackLayer.showAttackOptions(attackableCoords);
                    
                    console.log('[GameViewerPage] Highlights sent to layers');
                } else {
                    console.warn('[GameViewerPage] Some highlight layers not available');
                }
            }
            
            // Update UI with unit info
            if (unitInfo.success && unitInfo.data) {
                this.updateSelectedUnitInfo(unitInfo.data);
            }
            
            // Add to game log
            console.log(`Unit selected at (${q}, ${r}) - ${movableCoords.length} moves, ${attackableCoords.length} attacks available`);
            
        } catch (error) {
            console.error('[GameViewerPage] Failed to select unit:', error);
            this.showToast('Error', 'Failed to select unit', 'error');
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

    private updateGameUIFromState(gameData: GameCreateData): void {
        // Update game status with player-specific color - use player ID directly
        this.updateGameStatus(`Ready - Player ${gameData.currentPlayer}'s Turn`, gameData.currentPlayer);
        
        // Update turn counter
        const turnElement = document.getElementById('turn-counter');
        if (turnElement) {
            turnElement.textContent = `Turn ${gameData.turnCounter}`;
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

    // =============================================================================
    // ComponentLifecycle Interface Implementation
    // =============================================================================

    /**
     * Phase 1: Initialize DOM and discover child components
     */
    initializeDOM(): ComponentLifecycle[] {
        console.log('GameViewerPage: initializeDOM() - Phase 1');
        
        // Subscribe to events BEFORE creating components
        this.subscribeToWorldViewerEvents();
        
        // Create child components
        this.createWorldViewerComponent();
        
        // Initialize basic UI state
        this.updateGameStatus('Game Loading...');
        this.initializeGameLog();
        
        console.log('GameViewerPage: DOM initialized, returning child components');
        
        // Return child components for lifecycle management
        const childComponents: ComponentLifecycle[] = [];
        if (this.worldViewer) {
            childComponents.push(this.worldViewer);
        }
        if (this.gameState) {
            childComponents.push(this.gameState);
        }
        if (this.terrainStatsPanel) {
            childComponents.push(this.terrainStatsPanel);
        }
        return childComponents;
    }

    /**
     * Phase 2: Inject dependencies (none needed for GameViewerPage)
     */
    injectDependencies(deps: Record<string, any>): void {
        console.log('GameViewerPage: injectDependencies() - Phase 2', Object.keys(deps));
        // GameViewerPage doesn't need external dependencies
    }

    /**
     * Phase 3: Activate component when all dependencies are ready
     */
    async activate(): Promise<void> {
        console.log('GameViewerPage: activate() - Phase 3');
        
        // Bind events now that all components are ready
        this.bindGameSpecificEvents();

        // Note: Interaction callbacks are set in world-viewer-ready event handler
        // after the PhaserGameScene is actually created and ready
        
        // Wait for world viewer to be ready, then load world and initialize game
        if (this.currentWorldId) {
            console.log('GameViewerPage: WorldId found, loading world and initializing game...');
            // Small delay to ensure WorldViewer is fully ready
            setTimeout(async () => {
                await this.loadWorldAndInitializeGame();
            }, 50);
        } else {
            console.warn('GameViewerPage: No currentWorldId found!');
        }
        
        console.log('GameViewerPage: activation complete');
    }

    /**
     * Cleanup phase (called by lifecycle controller if needed)
     */
    deactivate(): void {
        console.log('GameViewerPage: deactivate() - cleanup');
        this.destroy();
    }

    public destroy(): void {
        if (this.worldViewer) {
            this.worldViewer.destroy();
            this.worldViewer = null;
        }
        
        if (this.gameState) {
            this.gameState.destroy();
            this.gameState = null;
        }

        if (this.terrainStatsPanel) {
            this.terrainStatsPanel.destroy();
            this.terrainStatsPanel = null;
        }
        
        this.world = null;
        this.currentWorldId = null;
        this.selectedUnit = null;
        this.gameLog = [];
    }
}

// Type definitions - using type alias instead of interface for simple data structures
type GameConfiguration = {
    playerCount: number;
    maxTurns: number;
    unitRestrictions: { [unitId: string]: string };
    playerTypes: { [playerId: string]: string };
    playerTeams: { [playerId: string]: number };
};

// Initialize page when DOM is ready using LifecycleController
document.addEventListener('DOMContentLoaded', async () => {
    console.log('DOM loaded, starting GameViewerPage initialization...');
    
    // Create page instance (just basic setup)
    const gameViewerPage = new GameViewerPage();
    
    // Create lifecycle controller with debug logging
    const lifecycleController = new LifecycleController({
        enableDebugLogging: true,
        phaseTimeoutMs: 15000, // Increased timeout for WASM loading
        continueOnError: false // Fail fast for debugging
    });
    
    // Set up lifecycle event logging
    lifecycleController.onLifecycleEvent((event) => {
        console.log(`[GameViewer Lifecycle] ${event.type}: ${event.componentName} - ${event.phase}`, event.error || '');
    });
    
    // Start breadth-first initialization
    await lifecycleController.initializeFromRoot(gameViewerPage, 'GameViewerPage');
    
    console.log('GameViewerPage fully initialized via LifecycleController');
});
