import { BaseComponent } from '../lib/Component';
import { EventBus } from '../lib/EventBus';
import Weewar_v1_servicesClient from '../gen/wasm-clients/weewar_v1_servicesClient.client';
import { ProcessMovesRequest, ProcessMovesResponse, ProcessMovesRequestSchema } from '../gen/weewar/v1/games_pb';
import { World, GameMove, WorldChange, GameMoveSchema, MoveUnitAction, MoveUnitActionSchema, AttackUnitAction, AttackUnitActionSchema, EndTurnAction, EndTurnActionSchema } from '../gen/weewar/v1/models_pb';
import { create } from '@bufbuild/protobuf';

/**
 * Minimal game state interface focused on core game data
 */
export interface GameStateData {
    wasmLoaded: boolean;
    gameId: string;
    currentPlayer: number;
    turnCounter: number;
    status: string;
    world: World | null; // Shared world object that all UI components reference
}

/**
 * Legacy interface for backward compatibility with GameViewerPage
 * TODO: Remove once GameViewerPage is updated to use new architecture
 */
export interface GameCreateData {
    currentPlayer: number;
    turnCounter: number;
    status: string;
    allUnits: any[];
    players: any[];
    teams: any[];
    mapSize: { rows: number; cols: number };
    winner: number;
    hasWinner: boolean;
}

/**
 * Legacy interface for backward compatibility with GameViewerPage  
 * TODO: Remove once GameViewerPage is updated to use new architecture
 */
export interface UnitSelectionData {
    unit: any;
    movableCoords: Array<{ coord: { Q: number; R: number }; cost: number }>;
    attackableCoords: Array<{ Q: number; R: number }>;
}

/**
 * GameState component - Minimal controller for ProcessMoves and world state management
 * 
 * Core responsibilities:
 * 1. Process game moves via ProcessMoves service
 * 2. Apply world changes to internal state
 * 3. Notify observers (UI components) of state changes
 * 
 * This replaces the previous 13+ manual WASM methods with a clean service-based approach.
 */
export class GameState extends BaseComponent {
    private client: Weewar_v1_servicesClient;
    private gameData: GameStateData;
    private wasmLoadPromise: Promise<void> | null;

    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('game-state', rootElement, eventBus, debugMode);
        
        // Initialize minimal game data
        this.gameData = {
            ...this.state,
            wasmLoaded: false,
            gameId: '', // Will be set when game is loaded/created
            currentPlayer: 0,
            turnCounter: 1,
            status: 'loading',
            world: null // Will be populated when game/world is loaded
        };
        
        // Initialize WASM client and loading
        this.client = new Weewar_v1_servicesClient();
        this.wasmLoadPromise = this.loadWASMModule();
    }

    protected initializeComponent(): void {
        this.log('Initializing minimal GameState controller...');
        
        // Try to load game data from page elements after WASM is loaded
        this.wasmLoadPromise?.then(() => {
            this.loadGameFromPageData();
        }).catch((error) => {
            this.log('Could not load game data from page during initialization:', error);
        });
    }

    protected bindToDOM(): void {
        // GameState is a data controller with no DOM interactions
        // It communicates via EventBus only
    }

    protected destroyComponent(): void {
        // this.client = null;
        this.wasmLoadPromise = null;
    }

    public getGameData(): GameStateData {
        return { ...this.gameData };
    }

    /**
     * Load the WASM module using generated client
     */
    private async loadWASMModule(): Promise<void> {
        this.log('Loading WASM module with generated client...');
        
        try {
            await this.client.loadWasm('/static/wasm/weewar-cli.wasm');
            
            this.gameData.wasmLoaded = true;
            this.gameData.lastUpdated = Date.now();

            this.log('WASM module loaded successfully via generated client');
            this.emit('wasm-loaded', { success: true });
        } catch (error) {
            this.log('Failed to load WASM module:', error);
            this.emit('wasm-load-error', { error });
            throw error;
        }
    }

    /**
     * Ensure WASM module is loaded before API calls
     */
    private async ensureWASMLoaded(): Promise<Weewar_v1_servicesClient> {
        if (this.gameData.wasmLoaded && this.client.isReady()) {
            return this.client;
        }

        if (!this.wasmLoadPromise) {
            throw new Error('WASM loading not started');
        }

        await this.wasmLoadPromise;

        if (!this.gameData.wasmLoaded || !this.client.isReady()) {
            throw new Error('WASM module failed to load');
        }
        return this.client;
    }

    /**
     * Check if WASM is ready for operations
     */
    public isReady(): boolean {
        return this.gameData.wasmLoaded;
    }

    /**
     * Wait for WASM to be ready (use during initialization)
     */
    public async waitUntilReady(): Promise<void> {
        await this.ensureWASMLoaded();
    }

    /**
     * Set the current game ID for subsequent move processing
     */
    public setGameId(gameId: string): void {
        this.gameData.gameId = gameId;
        this.gameData.lastUpdated = Date.now();
        this.log('Game ID set to:', gameId);
    }

    /**
     * Set the shared world object that all UI components reference
     */
    public setWorld(world: World): void {
        this.gameData.world = world;
        this.gameData.lastUpdated = Date.now();
        this.log('World object set');
        
        // Notify observers that world has been loaded/updated
        this.emit('world-loaded', { world: world });
    }

    /**
     * Get the shared world object (used by all UI components)
     */
    public getWorld(): World | null {
        return this.gameData.world;
    }

    /**
     * CORE METHOD: Process game moves and apply world changes
     * 
     * This is the primary interface for all game actions:
     * - Unit movements
     * - Unit attacks  
     * - End turn actions
     * - Any other game state modifications
     */
    public async processMoves(moves: GameMove[]): Promise<WorldChange[]> {
        const client = await this.ensureWASMLoaded();

        if (!this.gameData.gameId) {
            throw new Error('Game ID not set. Call setGameId() first.');
        }

        try {
            this.log('Processing moves:', moves);

            // Create request for ProcessMoves service
            const request = create(ProcessMovesRequestSchema, {
                gameId: this.gameData.gameId,
                moves: moves
            });

            // Call the ProcessMoves service  
            const response: ProcessMovesResponse = await client.gamesService.processMoves(request);

            // Extract world changes from response
            const worldChanges = response.changes || [];
            
            this.log('Received world changes:', worldChanges);

            // Apply changes to internal state and notify observers
            this.applyWorldChanges(worldChanges);
            
            return worldChanges;

        } catch (error: any) {
            this.log('ProcessMoves failed:', error);
            throw new Error(`ProcessMoves failed: ${error.message}`);
        }
    }

    /**
     * Apply world changes to internal game state and shared World object
     */
    private applyWorldChanges(changes: WorldChange[]): void {
        let stateChanged = false;

        // Process each world change and update internal state + shared world
        for (const change of changes) {
            // Handle different types of world changes using union type
            if (change.changeType.case === 'playerChanged') {
                this.gameData.currentPlayer = change.changeType.value.newPlayer;
                this.gameData.turnCounter = change.changeType.value.newTurn;
                stateChanged = true;
                this.log('Player changed:', change.changeType.value);
            }
            
            if (change.changeType.case === 'unitMoved') {
                this.applyUnitMovedToWorld(change.changeType.value);
                stateChanged = true;
                this.log('Unit moved:', change.changeType.value);
            }

            if (change.changeType.case === 'unitDamaged') {
                this.applyUnitDamagedToWorld(change.changeType.value);
                stateChanged = true;
                this.log('Unit damaged:', change.changeType.value);
            }

            if (change.changeType.case === 'unitKilled') {
                this.applyUnitKilledToWorld(change.changeType.value);
                stateChanged = true;
                this.log('Unit killed:', change.changeType.value);
            }
        }

        if (stateChanged) {
            this.gameData.lastUpdated = Date.now();
            this.notifyObservers(changes);
        }
    }

    /**
     * Apply unit movement to the shared World object
     */
    private applyUnitMovedToWorld(unitMoved: any): void {
        if (!this.gameData.world || !this.gameData.world.worldData?.tiles) {
            this.log('Cannot apply unit move - world not loaded');
            return;
        }

        // TODO: Find and move the unit in the world tiles array
        // This will require accessing world.worldData.tiles and finding the unit at fromQ,fromR
        // then moving it to toQ,toR
        this.log('Applying unit move to world:', unitMoved);
    }

    /**
     * Apply unit damage to the shared World object
     */
    private applyUnitDamagedToWorld(unitDamaged: any): void {
        if (!this.gameData.world || !this.gameData.world.worldData?.tiles) {
            this.log('Cannot apply unit damage - world not loaded');
            return;
        }

        // TODO: Find the unit at q,r and update its health
        this.log('Applying unit damage to world:', unitDamaged);
    }

    /**
     * Apply unit death to the shared World object
     */
    private applyUnitKilledToWorld(unitKilled: any): void {
        if (!this.gameData.world || !this.gameData.world.worldData?.tiles) {
            this.log('Cannot apply unit death - world not loaded');
            return;
        }

        // TODO: Find and remove the unit at q,r from the world
        this.log('Applying unit death to world:', unitKilled);
    }

    /**
     * Notify all observers (UI components) of world state changes
     */
    private notifyObservers(changes: WorldChange[]): void {
        // Emit specific events for different types of changes
        this.emit('world-changed', { 
            changes: changes,
            gameState: this.getGameData()
        });

        // Emit granular events for specific UI components
        for (const change of changes) {
            if (change.changeType.case === 'playerChanged') {
                this.emit('turn-ended', {
                    previousPlayer: change.changeType.value.previousPlayer,
                    currentPlayer: change.changeType.value.newPlayer,
                    turnCounter: change.changeType.value.newTurn
                });
            }

            if (change.changeType.case === 'unitMoved') {
                this.emit('unit-moved', {
                    from: { q: change.changeType.value.fromQ, r: change.changeType.value.fromR },
                    to: { q: change.changeType.value.toQ, r: change.changeType.value.toR }
                });
            }

            if (change.changeType.case === 'unitDamaged') {
                this.emit('unit-damaged', {
                    position: { q: change.changeType.value.q, r: change.changeType.value.r },
                    previousHealth: change.changeType.value.previousHealth,
                    newHealth: change.changeType.value.newHealth
                });
            }

            if (change.changeType.case === 'unitKilled') {
                this.emit('unit-killed', {
                    position: { q: change.changeType.value.q, r: change.changeType.value.r },
                    player: change.changeType.value.player,
                    unitType: change.changeType.value.unitType
                });
            }
        }
    }

    /**
     * Helper function to create GameMove for unit movement
     */
    public static createMoveUnitAction(fromQ: number, fromR: number, toQ: number, toR: number, playerId: number): GameMove {
        const moveAction = create(MoveUnitActionSchema, {
            fromQ: fromQ,
            fromR: fromR,
            toQ: toQ,
            toR: toR
        });

        return create(GameMoveSchema, {
            player: playerId,
            moveType: {
                case: 'moveUnit',
                value: moveAction
            }
        });
    }

    /**
     * Helper function to create GameMove for unit attack
     */
    public static createAttackUnitAction(attackerQ: number, attackerR: number, defenderQ: number, defenderR: number, playerId: number): GameMove {
        const attackAction = create(AttackUnitActionSchema, {
            attackerQ: attackerQ,
            attackerR: attackerR,
            defenderQ: defenderQ,
            defenderR: defenderR
        });

        return create(GameMoveSchema, {
            player: playerId,
            moveType: {
                case: 'attackUnit',
                value: attackAction
            }
        });
    }

    /**
     * Helper function to create GameMove for end turn
     */
    public static createEndTurnAction(playerId: number): GameMove {
        const endTurnAction = create(EndTurnActionSchema, {});

        return create(GameMoveSchema, {
            player: playerId,
            moveType: {
                case: 'endTurn',
                value: endTurnAction
            }
        });
    }

    /**
     * Load game data from pre-populated page elements (v1.Game, v1.GameState, v1.GameMoveHistory)
     * This replaces the old createGameFromMap approach with direct data loading
     */
    public loadGameFromPageData(): void {
        try {
            // Load v1.Game data from the element IDs that match the template
            const gameElement = document.getElementById('game.data-json');
            if (gameElement?.textContent && gameElement.textContent.trim() !== 'null') {
                const gameData = JSON.parse(gameElement.textContent);
                this.setGameId(gameData.id);
                this.gameData.currentPlayer = gameData.currentPlayer || 1;
                this.gameData.turnCounter = gameData.turnCounter || 1;
                this.gameData.status = gameData.status || 'active';
                this.log('Loaded game data:', gameData);
                
                // If game has world data, set it
                if (gameData.world) {
                    this.setWorld(gameData.world);
                    this.log('Loaded world from game data');
                }
            }

            // Load v1.GameState data 
            const gameStateElement = document.getElementById('game-state-data-json');
            if (gameStateElement?.textContent && gameStateElement.textContent.trim() !== 'null') {
                const gameStateData = JSON.parse(gameStateElement.textContent);
                
                // Update game state fields
                if (gameStateData.currentPlayer !== undefined) {
                    this.gameData.currentPlayer = gameStateData.currentPlayer;
                }
                if (gameStateData.turnCounter !== undefined) {
                    this.gameData.turnCounter = gameStateData.turnCounter;
                }
                if (gameStateData.status !== undefined) {
                    this.gameData.status = gameStateData.status;
                }
                
                this.log('Loaded game state data:', gameStateData);
            }

            // Load v1.GameMoveHistory data (optional)
            const historyElement = document.getElementById('game-history-data-json');
            if (historyElement?.textContent && historyElement.textContent.trim() !== 'null') {
                const historyData = JSON.parse(historyElement.textContent);
                this.log('Loaded game move history:', historyData);
                // TODO: Store history if needed for replay functionality
            }

            // Update last updated timestamp
            this.gameData.lastUpdated = Date.now();
            
            // Emit game loaded event
            this.emit('game-loaded', { 
                gameId: this.gameData.gameId,
                currentPlayer: this.gameData.currentPlayer,
                turnCounter: this.gameData.turnCounter,
                status: this.gameData.status
            });

        } catch (error: any) {
            this.log('Failed to load game data from page:', error);
            throw new Error(`Failed to load game data: ${error.message}`);
        }
    }

    /**
     * Legacy method for compatibility with GameViewerPage
     * Creates an EndTurn action and processes it
     */
    public async endTurn(playerId: number): Promise<void> {
        const endTurnMove = GameState.createEndTurnAction(playerId);
        await this.processMoves([endTurnMove]);
    }

    /**
     * Legacy method for compatibility with GameViewerPage
     * Returns current game state data
     */
    public getGameState(): any {
        return this.getGameData();
    }

    /**
     * Legacy method for compatibility with GameViewerPage
     * Creates a MoveUnit action and processes it
     */
    public async moveUnit(fromQ: number, fromR: number, toQ: number, toR: number): Promise<void> {
        const moveAction = GameState.createMoveUnitAction(fromQ, fromR, toQ, toR, this.gameData.currentPlayer);
        await this.processMoves([moveAction]);
    }

    /**
     * Legacy method for compatibility with GameViewerPage
     * Creates an AttackUnit action and processes it
     */
    public async attackUnit(attackerQ: number, attackerR: number, defenderQ: number, defenderR: number): Promise<void> {
        const attackAction = GameState.createAttackUnitAction(attackerQ, attackerR, defenderQ, defenderR, this.gameData.currentPlayer);
        await this.processMoves([attackAction]);
    }

    /**
     * Legacy method for compatibility with GameViewerPage
     * TODO: Implement using services or local world access
     */
    public async getTerrainStatsAt(q: number, r: number): Promise<any> {
        // TODO: Access world data directly or use a service
        this.log('getTerrainStatsAt called - needs implementation');
        return null;
    }

    /**
     * Legacy method for compatibility with GameViewerPage
     * Uses canSelectUnit service
     */
    public async canSelectUnit(q: number, r: number, playerId: number): Promise<boolean> {
        const client = await this.ensureWASMLoaded();
        
        // TODO: Implement using gamesService.canSelectUnit
        this.log('canSelectUnit called - needs implementation');
        return false;
    }

    /**
     * Legacy method for compatibility with GameViewerPage
     * TODO: Implement using local world access
     */
    public async getTileInfo(q: number, r: number): Promise<any> {
        // TODO: Access world data directly
        this.log('getTileInfo called - needs implementation');
        return null;
    }

    /**
     * Legacy method for compatibility with GameViewerPage
     * Uses getMovementOptions service
     */
    public async getMovementOptions(q: number, r: number, playerId: number): Promise<any[]> {
        const client = await this.ensureWASMLoaded();
        
        // TODO: Implement using gamesService.getMovementOptions
        this.log('getMovementOptions called - needs implementation');
        return [];
    }

    /**
     * Legacy method for compatibility with GameViewerPage
     * Uses getAttackOptions service
     */
    public async getAttackOptions(q: number, r: number, playerId: number): Promise<any[]> {
        const client = await this.ensureWASMLoaded();
        
        // TODO: Implement using gamesService.getAttackOptions
        this.log('getAttackOptions called - needs implementation');
        return [];
    }

    /**
     * Initialize game save/load bridge functions for WASM BrowserSaveHandler
     * These functions are called by the Go BrowserSaveHandler implementation
     */
    public static initializeSaveBridge(): void {
        // Set up bridge functions that WASM BrowserSaveHandler expects
        (window as any).gameSaveHandler = async (sessionData: string): Promise<string> => {
            try {
                const response = await fetch('/api/v1/games/sessions', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: sessionData
                });
                
                if (!response.ok) {
                    throw new Error(`Save failed: ${response.statusText}`);
                }
                
                const result = await response.json();
                return JSON.stringify({ success: true, sessionId: result.sessionId });
            } catch (error: any) {
                return JSON.stringify({ success: false, error: error.message });
            }
        };
        
        console.log('Game save/load bridge functions initialized');
    }
}
