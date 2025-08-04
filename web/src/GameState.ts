import { EventBus } from '../lib/EventBus';
import Weewar_v1_servicesClient from '../gen/wasm-clients/weewar_v1_servicesClient.client';
import { ProcessMovesRequest, ProcessMovesResponse, GetGameRequest, GetGameStateRequest, GetOptionsAtRequest, GameMove, WorldChange, MoveUnitAction, AttackUnitAction, EndTurnAction, GameState as ProtoGameState, Game as ProtoGame, WorldData } from '../gen/wasm-clients/weewar/v1/models'
import { create } from '@bufbuild/protobuf';

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
 * GameState - Lightweight WASM interface and game metadata manager
 * 
 * Core responsibilities:
 * 1. WASM client wrapper (getOptionsAt, processMoves)
 * 2. Game metadata cache (gameId, currentPlayer, turnCounter)
 * 3. Subscribe to changes to keep metadata in sync  
 * 4. Direct EventBus integration
 */
export class GameState {
    private client: Weewar_v1_servicesClient;
    private eventBus: EventBus;
    private wasmLoadPromise: Promise<void> | null;
    private wasmLoaded: boolean = false;
    
    // ✅ Lightweight game metadata only
    private gameId: string = '';
    private currentPlayer: number = 1;
    private turnCounter: number = 1;
    private gameName: string = '';

    constructor(eventBus: EventBus) {
        this.eventBus = eventBus;
        
        // Initialize WASM client with Go compatibility enabled
        this.client = new Weewar_v1_servicesClient();
        this.wasmLoadPromise = this.loadWASMModule();
        
        // ✅ Subscribe to server-changes to keep metadata in sync
        this.eventBus.addSubscription('server-changes', null, this);
    }

    /**
     * Handle EventBus events - specifically server-changes to keep metadata in sync
     */
    public handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
        if (eventType === 'server-changes') {
            this.updateMetadataFromChanges(data.changes);
        }
    }
    
    /**
     * Update game metadata from world changes
     */
    private updateMetadataFromChanges(changes: WorldChange[]): void {
        for (const change of changes) {
            if (change.playerChanged) {
                this.currentPlayer = change.playerChanged.newPlayer;
                this.turnCounter = change.playerChanged.newTurn;
            }
        }
    }

    /**
     * Load the WASM module using generated client
     */
    private async loadWASMModule(): Promise<void> {
        console.log('[GameState] Loading WASM module with generated client...');
    
        await this.client.loadWasm('/static/wasm/weewar-cli.wasm');
        
        // Wait for Go-exported functions to be available on window.weewar
        await this.waitForGoFunctions();
        
        this.wasmLoaded = true;

        console.log('[GameState] WASM module loaded successfully via generated client');
        this.eventBus.emit('wasm-loaded', { success: true }, this, this);
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
                console.log('[GameState] Go functions are now available on window.weewar');
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
        return this.gameId;
    }

    /**
     * Set the game ID for this session
     */
    public setGameId(gameId: string): void {
        this.gameId = gameId;
    }

    /**
     * ✅ Essential metadata getters (no WASM calls needed)
     */
    public getCurrentPlayer(): number {
        return this.currentPlayer;
    }

    public getTurnCounter(): number {
        return this.turnCounter;
    }

    public getGameName(): string {
        return this.gameName;
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

        if (!this.gameId) {
            throw new Error('Game ID not set. Call setGameId() first.');
        }

        console.log('[GameState] Processing moves:', moves);

        // Create request for ProcessMoves service
        const request = ProcessMovesRequest.from({
            gameId: this.gameId,
            moves: moves
        });

        // Call the ProcessMoves service  
        const response: ProcessMovesResponse = await client.gamesService.processMoves(request);

        // Extract world changes from move results (each move result contains its own changes)
        const worldChanges: WorldChange[] = [];
        for (const moveResult of response.moveResults || []) {
            worldChanges.push(...(moveResult.changes || []));
        }
        
        console.log('[GameState] Received ProcessMoves response:', {
            moveResultsCount: response.moveResults?.length || 0,
            totalWorldChanges: worldChanges.length,
            worldChanges: worldChanges
        });

        // ✅ Direct EventBus emit for World to coordinate
        this.eventBus.emit('server-changes', { changes: worldChanges }, this, this);
        
        // Return changes for any components that still need them
        return worldChanges;
    }

    /**
     * Helper function to create GameMove for unit movement
     */
    public static createMoveUnitAction(fromQ: number, fromR: number, toQ: number, toR: number, playerId: number): GameMove {
        const moveAction = MoveUnitAction.from({
            fromQ: fromQ,
            fromR: fromR,
            toQ: toQ,
            toR: toR
        });

        return GameMove.from({ player: playerId, moveUnit: moveAction, });
    }

    /**
     * Helper function to create GameMove for unit attack
     */
    public static createAttackUnitAction(attackerQ: number, attackerR: number, defenderQ: number, defenderR: number, playerId: number): GameMove {
        const attackAction = AttackUnitAction.from({
            attackerQ: attackerQ,
            attackerR: attackerR,
            defenderQ: defenderQ,
            defenderR: defenderR
        });

        return GameMove.from({ player: playerId, attackUnit: attackAction });
    }

    /**
     * Helper function to create GameMove for end turn
     */
    public static createEndTurnAction(playerId: number): GameMove {
        const endTurnAction = EndTurnAction.from({});
        return GameMove.from({ player: playerId, endTurn: endTurnAction });
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
        console.log(`[GameState] Raw game data from page:`, gameElement.textContent?.substring(0, 100) + '...');
        console.log(`[GameState] Raw game state from page:`, (gameStateElement?.textContent || 'null').substring(0, 100) + '...');
        console.log(`[GameState] Raw history from page:`, (historyElement?.textContent || 'null').substring(0, 100) + '...');
        
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
        
        console.log(`[GameState] Calling WASM loadGameData with game data bytes`);
        const wasmResult = weewar.loadGameData(gameBytes, gameStateBytes, historyBytes);
        
        if (!wasmResult.success) {
            throw new Error(`WASM load failed: ${wasmResult.error}`);
        }
        
        console.log('[GameState] Game data loaded into WASM singletons:', wasmResult.message);
        
        // Extract game ID and initial metadata from loaded data
        if (gameElement?.textContent) {
            try {
                const gameData = JSON.parse(gameElement.textContent);
                this.gameId = gameData.id || 'test';
                this.gameName = gameData.name || 'Untitled Game';
                console.log('[GameState] Extracted game ID:', this.gameId);
            } catch (error) {
                console.log('[GameState] Could not parse game ID from JSON, using default');
                this.gameId = 'test';
            }
        }
        
        // ✅ Extract initial game state metadata (currentPlayer, turnCounter)
        if (gameStateElement?.textContent && gameStateElement.textContent.trim() !== 'null') {
            try {
                const gameStateData = JSON.parse(gameStateElement.textContent);
                this.currentPlayer = gameStateData.currentPlayer || 1;
                this.turnCounter = gameStateData.turnCounter || 1;
                console.log('[GameState] Extracted initial game state:', {
                    currentPlayer: this.currentPlayer,
                    turnCounter: this.turnCounter
                });
            } catch (error) {
                console.log('[GameState] Could not parse game state from JSON, using defaults');
            }
        }
        
        // Emit event to indicate WASM data is loaded and ready for queries
        this.eventBus.emit('wasm-data-loaded', { gameId: this.gameId }, this, this);
    }

    /**
     * Query current game state from WASM (no caching)
     */
    public async getCurrentGameState(): Promise<ProtoGameState> {
        const client = await this.ensureWASMLoaded();
        const request = GetGameStateRequest.from({ gameId: this.gameId });
        const response = await client.gamesService.getGameState(request);
        return response.state || ProtoGameState.from({});
    }

    /**
     * Query current game data from WASM (no caching)
     */
    public async getCurrentGame(): Promise<ProtoGame> {
        const client = await this.ensureWASMLoaded();
        const request = GetGameRequest.from({ id: this.gameId });
        const response = await client.gamesService.getGame(request);
        return response.game || ProtoGame.from({});
    }

    /**
     * Query current world data from WASM (no caching)
     */
    public async getWorldData(): Promise<WorldData> {
        const gameState = await this.getCurrentGameState();
        return gameState.worldData || WorldData.from({ tiles: [], units: [] });
    }

    /**
     * ✅ Simple endTurn method (still used by GameViewerPage)
     */
    public async endTurn(playerId: number): Promise<void> {
        const endTurnMove = GameState.createEndTurnAction(playerId);
        await this.processMoves([endTurnMove]);
    }


    /**
     * ✅ Get all options at a position (core WASM method)
     */
    public async getOptionsAt(q: number, r: number): Promise<any> {
        console.log(`[GameState] getOptionsAt called with q=${q}, r=${r}`);
        const client = await this.ensureWASMLoaded();
        
        try {
            if (!this.gameId) {
                console.log('[GameState] No game ID available for getOptionsAt');
                return { options: [], currentPlayer: 0, gameInitialized: false };
            }

            const request = GetOptionsAtRequest.from({
                gameId: this.gameId,
                q: q,
                r: r
            });
            console.log('[GameState] getOptionsAt request:', request);

            const response = await client.gamesService.getOptionsAt(request);
            console.log('[GameState] getOptionsAt response:', response);
            
            console.log(`[GameState] getOptionsAt(${q}, ${r}): ${response.options?.length || 0} options, currentPlayer: ${response.currentPlayer}`);
            return response;
        } catch (error) {
            console.log(`[GameState] Error in getOptionsAt: ${error}`);
            return { options: [], currentPlayer: 0, gameInitialized: false };
        }
    }

    /**
     * Initialize game save/load bridge functions for WASM BrowserSaveHandler
     * These functions are called by the Go BrowserSaveHandler implementation
     */
    public static initializeSaveBridge2(): void {
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
