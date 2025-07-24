import { BaseComponent, ComponentState } from './Component';
import { EventBus } from './EventBus';

/**
 * Extended state interface for GameState component
 */
export interface GameStateData extends ComponentState {
    wasmLoaded: boolean;
    gameInitialized: boolean;
    currentPlayer: number;
    turnCounter: number;
    status: string;
    mapSize: { rows: number; cols: number };
    winner: number;
    hasWinner: boolean;
}

/**
 * Type definitions for WASM API responses
 */
export interface WASMResponse {
    success: boolean;
    message: string;
    data: any;
}

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

export interface UnitSelectionData {
    unit: any;
    movableCoords: Array<{ coord: { Q: number; R: number }; cost: number }>;
    attackableCoords: Array<{ Q: number; R: number }>;
}

/**
 * GameState component manages WASM module loading and game state
 * Provides a clean interface between UI components and the Go WASM game engine
 * 
 * This component follows the thin wrapper principle - it only manages loading
 * and state synchronization, delegating all game logic to the WASM module.
 */
export class GameState extends BaseComponent {
    private wasm: any;
    private gameData: GameStateData;
    private wasmLoadPromise: Promise<void> | null;

    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        // Initialize gameData before calling super to ensure it's available in initializeComponent
        super('game-state', rootElement, eventBus, debugMode);
        
        // Initialize gameData after super() since we need this.state from BaseComponent
        this.gameData = {
            ...this.state,
            wasmLoaded: false,
            gameInitialized: false,
            currentPlayer: 0,
            turnCounter: 1,
            status: 'loading',
            mapSize: { rows: 0, cols: 0 },
            winner: -1,
            hasWinner: false
        };
        
        // Initialize WASM loading after gameData is set up
        this.wasmLoadPromise = this.loadWASMModule();
    }

    protected initializeComponent(): void {
        this.log('Initializing GameState component...');
        
        // WASM loading is now handled in constructor after gameData is initialized
        // No additional initialization needed here
    }

    protected bindToDOM(): void {
        // GameState is a data component with no DOM interactions
        // It communicates via EventBus only
    }

    protected destroyComponent(): void {
        this.wasm = null;
        this.wasmLoadPromise = null;
    }

    public getGameData(): GameStateData {
        return { ...this.gameData };
    }

    /**
     * Load the WASM module asynchronously
     */
    private async loadWASMModule(): Promise<void> {
        this.log('Loading WASM module...');
        
        // Check if WASM is already loaded (for testing environments)
        if ((window as any).weewarCreateGameFromMap) {
            this.log('WASM module already loaded (pre-loaded in test environment)');
            
            this.wasm = {
                createGameFromMap: (window as any).weewarCreateGameFromMap,
                getGameState: (window as any).weewarGetGameState,
                selectUnit: (window as any).weewarSelectUnit,
                moveUnit: (window as any).weewarMoveUnit,
                attackUnit: (window as any).weewarAttackUnit,
                endTurn: (window as any).weewarEndTurn,
                getTerrainStatsAt: (window as any).weewarGetTerrainStatsAt,
                canSelectUnit: (window as any).weewarCanSelectUnit,
                getTileInfo: (window as any).weewarGetTileInfo,
                getMovementOptions: (window as any).weewarGetMovementOptions,
                getAttackOptions: (window as any).weewarGetAttackOptions
            };

            this.gameData.wasmLoaded = true;
            this.gameData.lastUpdated = Date.now();

            this.log('WASM module ready (pre-loaded)');
            this.emit('wasm-loaded', { success: true });
            return;
        }
        
        // Load Go's WASM support
        if (!(window as any).Go) {
            const script = document.createElement('script');
            script.src = '/static/wasm/wasm_exec.js';
            document.head.appendChild(script);
            
            await new Promise<void>((resolve, reject) => {
                script.onload = () => resolve();
                script.onerror = () => reject(new Error('Failed to load wasm_exec.js'));
            });
        }

        // Initialize Go WASM runtime
        const go = new (window as any).Go();
        const wasmModule = await WebAssembly.instantiateStreaming(
            fetch('/static/wasm/weewar-cli.wasm'),
            go.importObject
        );

        // Run the WASM module
        go.run(wasmModule.instance);

        // Verify WASM APIs are available
        if (!(window as any).weewarCreateGameFromMap) {
            throw new Error('WASM APIs not found - module may not have loaded correctly');
        }

        this.wasm = {
            createGameFromMap: (window as any).weewarCreateGameFromMap,
            getGameState: (window as any).weewarGetGameState,
            selectUnit: (window as any).weewarSelectUnit,
            moveUnit: (window as any).weewarMoveUnit,
            attackUnit: (window as any).weewarAttackUnit,
            endTurn: (window as any).weewarEndTurn,
            getTerrainStatsAt: (window as any).weewarGetTerrainStatsAt,
            canSelectUnit: (window as any).weewarCanSelectUnit,
            getTileInfo: (window as any).weewarGetTileInfo,
            getMovementOptions: (window as any).weewarGetMovementOptions,
            getAttackOptions: (window as any).weewarGetAttackOptions
        };

        this.gameData.wasmLoaded = true;
        this.gameData.lastUpdated = Date.now();

        this.log('WASM module loaded successfully');
        this.emit('wasm-loaded', { success: true });

        /** Disabling try/catch for now
        try {
        } catch (error) {
            this.handleError('Failed to load WASM module', error);
            this.emit('wasm-load-error', { error });
        }
       */
    }

    /**
     * Create a new game from map data
     */
    public async createGameFromMap(mapData: string): Promise<GameCreateData> {
        // Ensure WASM is loaded (only async part)
        await this.ensureWASMLoaded();

        this.log(`Creating game with players`);
        
        const response: WASMResponse = this.wasm.createGameFromMap(mapData);
        
        if (!response.success) {
            throw new Error(`Game creation failed: ${response.message}`);
        }

        // Update game state from WASM response
        const gameData: GameCreateData = response.data;
        this.updateGameState(gameData);

        this.gameData.gameInitialized = true;
        this.log('Game created successfully', gameData);
        
        // Emit notification event (not for response handling)
        this.emit('game-created', gameData);
        return gameData;
    }

    /**
     * Get current game state from WASM (synchronous once WASM loaded)
     */
    public getGameState(): GameCreateData {
        this.ensureWASMLoadedSync();

        const response: WASMResponse = this.wasm.getGameState();
        
        if (!response.success) {
            throw new Error(`Get game state failed: ${response.message}`);
        }

        const gameData: GameCreateData = response.data;
        this.updateGameState(gameData);
        
        // Emit notification event
        this.emit('game-state-updated', gameData);
        return gameData;
    }

    /**
     * Select a unit and get movement/attack options (synchronous)
     */
    public selectUnit(q: number, r: number): UnitSelectionData {
        this.ensureWASMLoadedSync();

        const response: WASMResponse = this.wasm.selectUnit(q, r);
        
        if (!response.success) {
            throw new Error(`Unit selection failed: ${response.message}`);
        }

        const selectionData: UnitSelectionData = response.data;
        this.log('Unit selected', selectionData);
        
        // Emit notification event
        this.emit('unit-selected', selectionData);
        return selectionData;
    }

    /**
     * Move a unit from one position to another (synchronous)
     */
    public moveUnit(fromQ: number, fromR: number, toQ: number, toR: number): any {
        this.ensureWASMLoadedSync();

        const response: WASMResponse = this.wasm.moveUnit(fromQ, fromR, toQ, toR);
        
        if (!response.success) {
            throw new Error(`Unit move failed: ${response.message}`);
        }

        this.log('Unit moved successfully', response.data);
        
        // Emit notification event
        this.emit('unit-moved', { 
            from: { q: fromQ, r: fromR }, 
            to: { q: toQ, r: toR }, 
            result: response.data 
        });
        
        return response.data;
    }

    /**
     * Attack with one unit against another (synchronous)
     */
    public attackUnit(attackerQ: number, attackerR: number, defenderQ: number, defenderR: number): any {
        this.ensureWASMLoadedSync();

        const response: WASMResponse = this.wasm.attackUnit(attackerQ, attackerR, defenderQ, defenderR);
        
        if (!response.success) {
            throw new Error(`Attack failed: ${response.message}`);
        }

        this.log('Attack completed', response.data);
        
        // Emit notification event
        this.emit('unit-attacked', { 
            attacker: { q: attackerQ, r: attackerR },
            defender: { q: defenderQ, r: defenderR },
            result: response.data 
        });
        
        return response.data;
    }

    /**
     * End current player's turn (synchronous)
     */
    public endTurn(): GameCreateData {
        this.ensureWASMLoadedSync();

        const response: WASMResponse = this.wasm.endTurn();
        
        if (!response.success) {
            throw new Error(`End turn failed: ${response.message}`);
        }

        const gameData: GameCreateData = response.data;
        this.updateGameState(gameData);

        this.log('Turn ended successfully', gameData);
        
        // Emit notification event
        this.emit('turn-ended', gameData);
        
        return gameData;
    }

    /**
     * Get detailed terrain stats for a specific tile (synchronous)
     */
    public getTerrainStatsAt(q: number, r: number): any {
        const response = this.ensureWASMLoadedSync().getTerrainStatsAt(q, r);
        if (!response.success) {
          return null
        }
        return response.data;
    }

    /**
     * Check if a unit at the given position can be selected by current player (synchronous)
     */
    public canSelectUnit(q: number, r: number): boolean {
        this.ensureWASMLoadedSync();

        const response: WASMResponse = this.wasm.canSelectUnit(q, r);
        
        if (!response.success) {
            // For boolean checks, return false on error rather than throwing
            this.log('CanSelectUnit check failed:', response.message);
            return false;
        }

        return response.data;
    }

    /**
     * Get basic tile information (synchronous)
     */
    public getTileInfo(q: number, r: number): any {
        this.ensureWASMLoadedSync();

        const response: WASMResponse = this.wasm.getTileInfo(q, r);
        
        if (!response.success) {
            throw new Error(`Get tile info failed: ${response.message}`);
        }

        this.log('Retrieved tile info for', { q, r, data: response.data });
        return response.data;
    }

    /**
     * Get movement options for unit at position (synchronous)
     */
    public getMovementOptions(q: number, r: number): WASMResponse {
        return this.ensureWASMLoadedSync().getMovementOptions(q, r);
    }

    /**
     * Get attack options for unit at position (synchronous)
     */
    public getAttackOptions(q: number, r: number): WASMResponse {
        return this.ensureWASMLoadedSync().getAttackOptions(q, r);
    }

    /**
     * Ensure WASM module is loaded before API calls
     */
    private async ensureWASMLoaded(): Promise<any> {
        if (this.gameData.wasmLoaded && this.wasm) {
            return this.wasm;
        }

        if (!this.wasmLoadPromise) {
            throw new Error('WASM loading not started');
        }

        await this.wasmLoadPromise;

        if (!this.gameData.wasmLoaded || !this.wasm) {
            throw new Error('WASM module failed to load');
        }
        return this.wasm
    }

    /**
     * Ensure WASM module is loaded (synchronous version for game actions)
     */
    private ensureWASMLoadedSync(): any {
        if (!this.gameData.wasmLoaded || !this.wasm) {
            throw new Error('WASM module not loaded. Call waitUntilReady() first during initialization.');
        }
        return this.wasm
    }

    /**
     * Check if WASM is ready for synchronous operations
     */
    public isReady(): boolean {
        return this.gameData.wasmLoaded && this.gameData.gameInitialized;
    }

    /**
     * Wait for WASM to be ready (only use during initialization)
     */
    public async waitUntilReady(): Promise<void> {
        await this.ensureWASMLoaded();
    }

    /**
     * Update internal game state from WASM response
     */
    private updateGameState(gameData: GameCreateData): void {
        this.gameData.currentPlayer = gameData.currentPlayer;
        this.gameData.turnCounter = gameData.turnCounter;
        this.gameData.status = gameData.status;
        this.gameData.mapSize = gameData.mapSize;
        this.gameData.winner = gameData.winner;
        this.gameData.hasWinner = gameData.hasWinner;
        this.gameData.lastUpdated = Date.now();
    }
}
