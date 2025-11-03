import { EventBus } from '../lib/EventBus';
import WeewarBundle from '../gen/wasmjs';
import { GamesServiceClient } from '../gen/wasmjs/weewar/v1/gamesServiceClient';
import { 
    ProcessMovesRequest, 
    ProcessMovesResponse, 
    GetGameRequest, 
    GetGameStateRequest, 
    GetOptionsAtRequest, 
    GetOptionsAtResponse,
    GameMove, 
    WorldChange, 
    MoveUnitAction, 
    AttackUnitAction, 
    EndTurnAction, 
    GameState as ProtoGameState, 
    Game as ProtoGame, 
    WorldData 
} from '../gen/wasmjs/weewar/v1/interfaces'

import * as models from '../gen/wasmjs/weewar/v1/models'

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
    private wasmBundle: WeewarBundle;
    private gamesClient: GamesServiceClient;
    private eventBus: EventBus;
    
    // Lightweight game metadata only
    private gameId: string = '';
    public currentPlayer: number = 1;
    public turnCounter: number = 1;
    private gameName: string = '';

    constructor(eventBus: EventBus) {
        this.eventBus = eventBus;
        
        // Create base bundle with module configuration
        this.wasmBundle  = new WeewarBundle();

        // Create service clients using composition
        this.gamesClient = new GamesServiceClient(this.wasmBundle);
        // Register browser API implementation when ready
        // const browserAPI = new BrowserAPIClient(wasmBundle);
        // this.wasmBundle.registerBrowserService('BrowserAPI', new BrowserAPIImpl());

        this.loadWASMModule();
        
        // ✅ Subscribe to server-changes to keep metadata in sync
        this.eventBus.addSubscription('server-changes', null, this);
    }

    /**
     * Handle EventBus events - specifically server-changes to keep metadata in sync
     */
    public handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
        if (eventType === 'server-changes') {
            console.log("Came here 11111")
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
        await this.wasmBundle.loadWasm('/static/wasm/weewar-cli.wasm');
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
        if (!this.gameId) {
            throw new Error('Game ID not set. Call setGameId() first.');
        }

        // Get the current WASM state before processing moves (for comparison)
        const preState = await this.getCurrentGameState();
        const preWorldData = preState.worldData;

        // Create request for ProcessMoves service
        const request = {
            gameId: this.gameId,
            moves: moves
        };

        // Call the ProcessMoves service  
        const response: ProcessMovesResponse = await this.gamesClient.processMoves(request);

        // Extract world changes from move results (each move result contains its own changes)
        const worldChanges: WorldChange[] = [];
        for (const moveResult of response.moveResults || []) {
            worldChanges.push(...(moveResult.changes || []));
        }

        // Get the actual WASM state after processing moves to ensure synchronization
        const postState = await this.getCurrentGameState();
        const postWorldData = postState.worldData;
        this.eventBus.emit('server-changes', { changes: worldChanges }, this, this);
        
        // Return changes for any components that still need them
        return worldChanges;
    }

    /**
     * Helper function to create GameMove for unit movement
     */
    public static createMoveUnitAction(fromQ: number, fromR: number, toQ: number, toR: number, playerId: number): GameMove {
        const moveAction = models.MoveUnitAction.from({
            fromQ: fromQ,
            fromR: fromR,
            toQ: toQ,
            toR: toR
        });

        return models.GameMove.from({ player: playerId, moveUnit: moveAction, });
    }

    /**
     * Helper function to create GameMove for unit attack
     */
    public static createAttackUnitAction(attackerQ: number, attackerR: number, defenderQ: number, defenderR: number, playerId: number): GameMove {
        const attackAction = models.AttackUnitAction.from({
            attackerQ: attackerQ,
            attackerR: attackerR,
            defenderQ: defenderQ,
            defenderR: defenderR
        });

        return models.GameMove.from({ player: playerId, attackUnit: attackAction });
    }

    /**
     * Helper function to create GameMove for end turn
     */
    public static createEndTurnAction(playerId: number): GameMove {
        const endTurnAction = models.EndTurnAction.from({});
        return models.GameMove.from({ player: playerId, endTurn: endTurnAction });
    }

    /**
     * Load game data into WASM singletons from page elements
     * This populates the WASM singleton objects that serve as the source of truth
     */
    public async loadGameDataToWasm(): Promise<void> {
        await this.wasmBundle.ensureReady();
        
        // Get raw JSON data from page elements
        const gameElement = document.getElementById('game.data-json');
        const gameStateElement = document.getElementById('game-state-data-json');
        const historyElement = document.getElementById('game-history-data-json');
        
        if (!gameElement?.textContent || gameElement.textContent.trim() === 'null') {
            throw new Error('No game data found in page elements');
        }
        
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
        
        // Extract game ID and initial metadata from loaded data
        if (gameElement?.textContent) {
            this.gameId = 'test';
            const gameData = JSON.parse(gameElement.textContent);
            this.gameId = gameData.id || 'test';
            this.gameName = gameData.name || 'Untitled Game';
        }
        
        // ✅ Extract initial game state metadata (currentPlayer, turnCounter)
        if (gameStateElement?.textContent && gameStateElement.textContent.trim() !== 'null') {
            const gameStateData = JSON.parse(gameStateElement.textContent);
            this.currentPlayer = gameStateData.currentPlayer || 1;
            this.turnCounter = gameStateData.turnCounter || 1;
        }
        
        // Emit event to indicate WASM data is loaded and ready for queries
        this.eventBus.emit('wasm-data-loaded', { gameId: this.gameId }, this, this);
    }

    /**
     * Query current game state from WASM (no caching)
     */
    public async getCurrentGameState(): Promise<ProtoGameState> {
        await this.wasmBundle.ensureReady();
        const request = models.GetGameStateRequest.from({ gameId: this.gameId });
        const response = await this.gamesClient.getGameState(request);
        return response.state || models.GameState.from({});
    }

    /**
     * Query current game data from WASM (no caching)
     */
    public async getCurrentGame(): Promise<ProtoGame> {
        const client = await this.wasmBundle.ensureReady();
        const request = models.GetGameRequest.from({ id: this.gameId });
        const response = await this.gamesClient.getGame(request);
        return response.game || models.Game.from({});
    }

    /**
     * Query current world data from WASM (no caching)
     */
    public async getWorldData(): Promise<WorldData> {
        const gameState = await this.getCurrentGameState();
        return gameState.worldData || models.WorldData.from({ tiles: [], units: [] });
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
    public async getOptionsAt(q: number, r: number): Promise<GetOptionsAtResponse> {
        if (!this.gameId) {
            return models.GetOptionsAtResponse.from({ 
                options: [], 
                currentPlayer: 0, 
                gameInitialized: false 
            });
        }

        const request = models.GetOptionsAtRequest.from({
            gameId: this.gameId,
            q: q,
            r: r
        });

        const response = await this.gamesClient.getOptionsAt(request);
        return response;
    }
}
