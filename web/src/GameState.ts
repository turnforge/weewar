import { BaseComponent } from '../lib/Component';
import { EventBus } from '../lib/EventBus';
import Weewar_v1_servicesClient from '../gen/wasm-clients/weewar_v1_servicesClient.client';
import { ProcessMovesRequest, ProcessMovesResponse, ProcessMovesRequestSchema, GetGameRequest, GetGameRequestSchema, GetGameStateRequest, GetGameStateRequestSchema, GetOptionsAtRequest, GetOptionsAtRequestSchema } from '../gen/weewar/v1/games_pb';
import { GameMove, WorldChange, GameMoveSchema, MoveUnitAction, MoveUnitActionSchema, AttackUnitAction, AttackUnitActionSchema, EndTurnAction, EndTurnActionSchema, GameState as ProtoGameState, GameStateSchema as ProtoGameStateSchema, Game as ProtoGame, GameSchema as ProtoGameSchema } from '../gen/weewar/v1/models_pb';
import { create } from '@bufbuild/protobuf';
import { World } from './World';

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
    private wasmLoadPromise: Promise<void> | null;
    private wasmLoaded: boolean = false;
    private world: World;
    status: string
    
    // Local cache of Game and GameState protos for query optimization (avoid WASM calls)
    // Source of truth is WASM, this is just a performance cache
    private cachedGame: ProtoGame;
    private cachedGameState: ProtoGameState;

    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('game-state', rootElement, eventBus, debugMode);
        
        // Initialize WASM client with Go compatibility enabled
        this.client = new Weewar_v1_servicesClient({
            oneofToJson: 'auto',    // Enable automatic oneof field conversion for Go compatibility (requests)
            oneofFromJson: 'auto',  // Enable automatic oneof field conversion from Go format (responses)
            emitDefaults: false,    // Match Go protobuf behavior - don't emit zero values
            bigIntHandler: (value: bigint) => value.toString() // Convert BigInt to string for JSON
        });
        this.wasmLoadPromise = this.loadWASMModule();
        
        // Initialize with empty objects (will be populated by loadGameDataToWasm)
        this.world = new World(eventBus, 'Loading...');
        this.status = 'loading'
        this.cachedGame = create(ProtoGameSchema, {
            id: '',
            name: 'Loading...',
            creatorId: '',
            worldId: ''
        });
        this.cachedGameState = create(ProtoGameStateSchema, {
            gameId: '',
            currentPlayer: 0,
            turnCounter: 1,
        });
    }

    protected initializeComponent(): void {
        this.log('Initializing WASM-centric GameState controller...');
        
        // WASM initialization happens automatically in constructor
        // Game data loading is now handled by GameViewerPage calling loadGameDataToWasm()
    }

    protected destroyComponent(): void {
        // this.client = null;
        this.wasmLoadPromise = null;
    }

    /**
     * Load the WASM module using generated client
     */
    private async loadWASMModule(): Promise<void> {
        this.log('Loading WASM module with generated client...');
    
        await this.client.loadWasm('/static/wasm/weewar-cli.wasm');
        
        // Wait for Go-exported functions to be available on window.weewar
        await this.waitForGoFunctions();
        
        this.wasmLoaded = true;

        this.log('WASM module loaded successfully via generated client');
        this.emit('wasm-loaded', { success: true }, this);
    }
    
    /**
     * Wait for Go-exported functions to be available on window.weewar
     */
    private async waitForGoFunctions(): Promise<void> {
        const maxWaitTime = 10000; // 10 seconds
        const checkInterval = 100; // 100ms
        let elapsed = 0;
        
        while (elapsed < maxWaitTime) {
            const weewar = (window as any).weewar;
            if (weewar && weewar.loadGameData) {
                this.log('Go functions are now available on window.weewar');
                return;
            }
            
            await new Promise(resolve => setTimeout(resolve, checkInterval));
            elapsed += checkInterval;
        }
        
        throw new Error('Timeout waiting for Go-exported functions to be available');
    }

    /**
     * Ensure WASM module is loaded before API calls
     */
    private async ensureWASMLoaded(): Promise<Weewar_v1_servicesClient> {
        if (this.wasmLoaded && this.client.isReady()) {
            return this.client;
        }

        if (!this.wasmLoadPromise) {
            throw new Error('WASM loading not started');
        }

        await this.wasmLoadPromise;

        if (!this.wasmLoaded || !this.client.isReady()) {
            throw new Error('WASM module failed to load');
        }
        return this.client;
    }

    /**
     * Check if WASM is ready for operations
     */
    public isReady(): boolean {
        return this.wasmLoaded;
    }

    /**
     * Wait for WASM to be ready (use during initialization)
     */
    public async waitUntilReady(): Promise<void> {
        await this.ensureWASMLoaded();
    }
    
    /**
     * Get the current game ID
     */
    public getGameId(): string {
        return this.cachedGameState.gameId;
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

        const gameId = this.cachedGame.id
        if (!gameId) {
            throw new Error('Game ID not set. Call setGameId() first.');
        }

        this.log('Processing moves:', moves);

        // Create request for ProcessMoves service
        const request = create(ProcessMovesRequestSchema, {
            gameId: gameId,
            moves: moves
        });

        // Call the ProcessMoves service  
        const response: ProcessMovesResponse = await client.gamesService.processMoves(request);

        // Extract world changes from move results (each move result contains its own changes)
        const worldChanges: WorldChange[] = [];
        for (const moveResult of response.moveResults || []) {
            worldChanges.push(...(moveResult.changes || []));
        }
        
        this.log('Received ProcessMoves response:', {
            moveResultsCount: response.moveResults?.length || 0,
            totalWorldChanges: worldChanges.length,
            worldChanges: worldChanges
        });

        // Apply changes to internal state and notify observers
        this.applyWorldChanges(worldChanges);
        
        return worldChanges;
    }

    /**
     * Apply world changes to World object and cached GameState for UI rendering
     * Note: Authoritative game state is maintained in WASM, this only updates UI layer
     */
    private applyWorldChanges(changes: WorldChange[]): void {
        let worldUpdated = false;
        let gameStateUpdated = false;

        // Process each world change and update both World object and cached GameState
        for (const change of changes) {
            if (change.changeType.case === 'unitMoved') {
                this.applyUnitMovedToWorld(change.changeType.value);
                worldUpdated = true;
                this.log('Applied unit move to World object:', change.changeType.value);
            }

            if (change.changeType.case === 'unitDamaged') {
                this.applyUnitDamagedToWorld(change.changeType.value);
                worldUpdated = true;
                this.log('Applied unit damage to World object:', change.changeType.value);
            }

            if (change.changeType.case === 'unitKilled') {
                this.applyUnitKilledToWorld(change.changeType.value);
                worldUpdated = true;
                this.log('Applied unit death to World object:', change.changeType.value);
            }
            
            // Update cached GameState for player changes
            if (change.changeType.case === 'playerChanged') {
                this.cachedGameState = create(ProtoGameStateSchema, {
                    ...this.cachedGameState,
                    currentPlayer: change.changeType.value.newPlayer,
                    turnCounter: change.changeType.value.newTurn
                });
                gameStateUpdated = true;
                this.log('Updated cached GameState for player change:', change.changeType.value);
            }
        }

        if (worldUpdated || gameStateUpdated) {
            this.notifyObservers(changes);
        }
    }

    /**
     * Apply unit movement to the shared World object
     */
    private applyUnitMovedToWorld(unitMoved: any): void {
        // Get the unit at the source position
        const unit = this.world.getUnitAt(unitMoved.fromQ, unitMoved.fromR);
        if (!unit) {
            this.log(`No unit found at (${unitMoved.fromQ}, ${unitMoved.fromR}) to move`);
            return;
        }

        // Remove unit from source position
        this.world.removeUnitAt(unitMoved.fromQ, unitMoved.fromR);
        
        // Place unit at destination position
        this.world.setUnitAt(unitMoved.toQ, unitMoved.toR, unit.unitType, unit.player);
        
        this.log(`Moved unit from (${unitMoved.fromQ}, ${unitMoved.fromR}) to (${unitMoved.toQ}, ${unitMoved.toR})`);
    }

    /**
     * Apply unit damage to the shared World object
     */
    private applyUnitDamagedToWorld(unitDamaged: any): void {
        // Get the unit at the specified position
        const unit = this.world.getUnitAt(unitDamaged.q, unitDamaged.r);
        if (!unit) {
            this.log(`No unit found at (${unitDamaged.q}, ${unitDamaged.r}) to damage`);
            return;
        }

        // Note: The World class doesn't currently track health, so we just log this change
        // The actual health tracking would happen in a more detailed unit model
        this.log(`Unit at (${unitDamaged.q}, ${unitDamaged.r}) damaged: ${unitDamaged.previousHealth} -> ${unitDamaged.newHealth}`);
    }

    /**
     * Apply unit death to the shared World object
     */
    private applyUnitKilledToWorld(unitKilled: any): void {
        // Remove the unit from the world
        const removed = this.world.removeUnitAt(unitKilled.q, unitKilled.r);
        if (removed) {
            this.log(`Removed killed unit at (${unitKilled.q}, ${unitKilled.r}): player ${unitKilled.player} unit type ${unitKilled.unitType}`);
        } else {
            this.log(`No unit found at (${unitKilled.q}, ${unitKilled.r}) to remove`);
        }
    }

    /**
     * Notify all observers (UI components) of world state changes
     */
    private notifyObservers(changes: WorldChange[]): void {
        // Emit specific events for different types of changes
        this.emit('world-changed', { 
            changes: changes,
            world: this.world
        }, this);

        // Emit granular events for specific UI components
        for (const change of changes) {
            if (change.changeType.case === 'playerChanged') {
                this.emit('turn-ended', {
                    previousPlayer: change.changeType.value.previousPlayer,
                    currentPlayer: change.changeType.value.newPlayer,
                    turnCounter: change.changeType.value.newTurn
                }, this);
            }

            if (change.changeType.case === 'unitMoved') {
                this.emit('unit-moved', {
                    from: { q: change.changeType.value.fromQ, r: change.changeType.value.fromR },
                    to: { q: change.changeType.value.toQ, r: change.changeType.value.toR }
                }, this);
            }

            if (change.changeType.case === 'unitDamaged') {
                this.emit('unit-damaged', {
                    position: { q: change.changeType.value.q, r: change.changeType.value.r },
                    previousHealth: change.changeType.value.previousHealth,
                    newHealth: change.changeType.value.newHealth
                }, this);
            }

            if (change.changeType.case === 'unitKilled') {
                this.emit('unit-killed', {
                    position: { q: change.changeType.value.q, r: change.changeType.value.r },
                    player: change.changeType.value.player,
                    unitType: change.changeType.value.unitType
                }, this);
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
     * Load game data into WASM singletons from page elements
     * This populates the WASM singleton objects that serve as the source of truth
     */
    public async loadGameDataToWasm(): Promise<void> {
        await this.ensureWASMLoaded();
        
        // Get raw JSON data from page elements
        const gameElement = document.getElementById('game.data-json');
        const gameStateElement = document.getElementById('game-state-data-json');
        const historyElement = document.getElementById('game-history-data-json');
        
        if (!gameElement?.textContent || gameElement.textContent.trim() === 'null') {
            throw new Error('No game data found in page elements');
        }
        
        // Debug: Log the actual content to understand what we're getting
        this.log('Raw game data from page:', gameElement.textContent?.substring(0, 100) + '...');
        this.log('Raw game state from page:', (gameStateElement?.textContent || 'null').substring(0, 100) + '...');
        this.log('Raw history from page:', (historyElement?.textContent || 'null').substring(0, 100) + '...');
        
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
        
        this.log('Calling WASM loadGameData with game data bytes');
        const wasmResult = weewar.loadGameData(gameBytes, gameStateBytes, historyBytes);
        
        if (!wasmResult.success) {
            throw new Error(`WASM load failed: ${wasmResult.error}`);
        }
        
        this.log('Game data loaded into WASM singletons:', wasmResult.message);
        
        // Now get the loaded game data from WASM to initialize our World object
        await this.initializeWorldFromWasm();
    }

    /**
     * Initialize local World and cached GameState from WASM data
     */
    private async initializeWorldFromWasm(): Promise<void> {
        const client = await this.ensureWASMLoaded();
        
        // Get game data from WASM to extract game ID and world data
        const req = create(GetGameRequestSchema, { id: 'test' })
        const gameResponse = await client.gamesService.getGame(req);
        
        if (!gameResponse.game || !gameResponse.state) {
            throw new Error('No game data returned from WASM');
        }
        
        // Update cached Game and GameState from WASM response
        this.cachedGame = gameResponse.game;
        this.cachedGameState = gameResponse.state;
        
        // Update World object from game data for UI rendering
        if (gameResponse.state.worldData) {
            this.world.setName(gameResponse.game.name || 'Untitled Game');
            this.world.loadTilesAndUnits(
                gameResponse.state.worldData.tiles || [],
                gameResponse.state.worldData.units || []
            );
        }
        
        this.log('World and cached GameState initialized from WASM data');
        
        // Notify observers that world has been loaded/updated
        this.emit('world-loaded', { world: this.world }, this);
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
     * Returns current game state data from local cache (avoids WASM calls)
     */
    public getGameState(): ProtoGameState {
        return this.cachedGameState;
    }
    
    /**
     * Get cached Game proto object for instant access (avoids WASM calls)
     * Source of truth is WASM, this is just a performance cache
     */
    public getGame(): ProtoGame {
        return this.cachedGame;
    }

    /**
     * Get the World object for UI rendering
     * GameState owns the World, other components should access it via this getter
     */
    public getWorld(): World {
        return this.world;
    }

    /**
     * Legacy method for compatibility with GameViewerPage
     * Creates a MoveUnit action and processes it
     */
    public async moveUnit(fromQ: number, fromR: number, toQ: number, toR: number): Promise<void> {
        // Get current player from cached GameState
        const currentPlayer = this.cachedGameState.currentPlayer || 1;
        
        const moveAction = GameState.createMoveUnitAction(fromQ, fromR, toQ, toR, currentPlayer);
        await this.processMoves([moveAction]);
    }

    /**
     * Legacy method for compatibility with GameViewerPage
     * Creates an AttackUnit action and processes it
     */
    public async attackUnit(attackerQ: number, attackerR: number, defenderQ: number, defenderR: number): Promise<void> {
        // Get current player from cached GameState
        const currentPlayer = this.cachedGameState.currentPlayer || 1;
        
        const attackAction = GameState.createAttackUnitAction(attackerQ, attackerR, defenderQ, defenderR, currentPlayer);
        await this.processMoves([attackAction]);
    }


    /**
     * Legacy method for compatibility with GameViewerPage
     * Returns unit info at the specified position if there is one
     */
    public async getTileInfo(q: number, r: number): Promise<any> {
        try {
            // Get unit info from the world data
            const unit = this.world.getUnitAt(q, r);
            if (unit) {
                this.log(`getTileInfo(${q}, ${r}): Unit player=${unit.player}, type=${unit.unitType}`);
                return {
                    hasUnit: true,
                    player: unit.player,
                    unitType: unit.unitType,
                    // Add other unit properties as needed
                };
            } else {
                this.log(`getTileInfo(${q}, ${r}): No unit found`);
                return {
                    hasUnit: false
                };
            }
        } catch (error) {
            this.log(`Error in getTileInfo: ${error}`);
            return null;
        }
    }


    /**
     * New unified method to get all options at a position
     * Replaces canSelectUnit, getMovementOptions, getAttackOptions
     */
    public async getOptionsAt(q: number, r: number): Promise<any> {
        const client = await this.ensureWASMLoaded();
        
        try {
            const gameId = this.cachedGame?.id;
            if (!gameId) {
                this.log('No game ID available for getOptionsAt');
                return { options: [], currentPlayer: 0, gameInitialized: false };
            }

            const request = create(GetOptionsAtRequestSchema, {
                gameId: gameId,
                q: q,
                r: r
            });

            const response = await client.gamesService.getOptionsAt(request);
            
            this.log(`getOptionsAt(${q}, ${r}): ${response.options?.length || 0} options, currentPlayer: ${response.currentPlayer}`);
            return response;
        } catch (error) {
            this.log(`Error in getOptionsAt: ${error}`);
            return { options: [], currentPlayer: 0, gameInitialized: false };
        }
    }

    /**
     * Initialize game save/load bridge functions for WASM BrowserSaveHandler
     * These functions are called by the Go BrowserSaveHandler implementation
     */
    public static initializeSaveBridge(): void {
        // Set up bridge functions that WASM BrowserSaveHandler expects
        (window as any).gameSaveHandler = async (sessionData: string): Promise<string> => {
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
        };
        
        console.log('Game save/load bridge functions initialized');
    }
}
