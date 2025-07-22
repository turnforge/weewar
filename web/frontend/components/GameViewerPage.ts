import { BasePage } from './BasePage';
import { EventBus, EventTypes } from './EventBus';
import { MapViewer } from './MapViewer';
import { Map } from './Map';

/**
 * Game Viewer Page - Interactive game play interface
 * Responsible for:
 * - Loading map as a game instance
 * - Coordinating WASM game engine
 * - Managing game state and turn flow
 * - Handling player interactions (unit selection, movement, attacks)
 * - Providing game controls and UI feedback
 */
class GameViewerPage extends BasePage {
    private currentMapId: string | null;
    private map: Map | null = null;
    private mapViewer: MapViewer | null = null;
    
    // Game configuration from URL parameters
    private playerCount: number = 2;
    private maxTurns: number = 0;
    private gameConfig: GameConfiguration = {
        playerCount: 2,
        maxTurns: 0,
        unitRestrictions: {},
        playerTypes: {},
        playerTeams: {}
    };
    
    // Game state
    private currentPlayer: number = 1;
    private currentTurn: number = 1;
    private selectedUnit: any = null;
    private gameLog: string[] = [];

    constructor() {
        super();
        this.loadGameConfiguration();
    }

    /**
     * Load game configuration from URL parameters and hidden inputs
     */
    private loadGameConfiguration(): void {
        // Get mapId from hidden input
        const mapIdInput = document.getElementById("mapIdInput") as HTMLInputElement | null;
        this.currentMapId = mapIdInput?.value.trim() || null;

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
     * Follows State→Subscribe→Create→Bind pattern from UI_DESIGN_PRINCIPLES.md
     */
    protected initializeSpecificComponents(): void {
        try {
            console.log('Initializing GameViewerPage components');

            // State: Load game configuration is already done in constructor
            
            // Subscribe: Subscribe to events BEFORE creating components
            this.subscribeToMapViewerEvents();
            
            // Create: Initialize child components  
            this.createMapViewerComponent();
            
            // Initialize game UI state
            this.updateGameStatus('Game Loading...');
            this.initializeGameLog();

        } catch (error) {
            console.error('Failed to initialize GameViewerPage components:', error);
            this.showToast('Error', 'Failed to initialize game interface', 'error');
        }
    }

    /**
     * Subscribe to MapViewer events before component creation
     */
    private subscribeToMapViewerEvents(): void {
        // Subscribe BEFORE creating MapViewer to catch initialization events
        this.eventBus.subscribe('map-viewer-ready', () => {
            console.log('GameViewerPage: MapViewer ready, loading map...');
            if (this.currentMapId) {
                // WebGL context timing - wait for next event loop tick
                setTimeout(async () => {
                    await this.loadMapAndInitializeGame();
                }, 10);
            }
        }, 'game-viewer-page');
    }

    /**
     * Create MapViewer component instance
     */
    private createMapViewerComponent(): void {
        const mapViewerContainer = document.getElementById('phaser-viewer-container');
        if (mapViewerContainer) {
            // Pass element directly (not string ID) as per UI_DESIGN_PRINCIPLES.md
            this.mapViewer = new MapViewer(mapViewerContainer, this.eventBus, true);
        } else {
            console.warn('GameViewerPage: phaser-viewer-container not found');
        }
    }

    /**
     * Load map data and initialize game
     */
    private async loadMapAndInitializeGame(): Promise<void> {
        try {
            console.log('Loading map and initializing game...');

            // Load map data from hidden element
            const mapData = this.loadMapDataFromElement();
            if (!mapData) {
                throw new Error('No map data found');
            }

            // Deserialize map
            this.map = Map.deserialize(mapData);
            
            // Load map into viewer
            if (this.mapViewer) {
                await this.mapViewer.loadMap(mapData);
                this.showToast('Success', `Game loaded: ${this.map.getName() || 'Untitled'}`, 'success');
            }

            // Initialize game state
            this.initializeGameState();
            
            // Update UI
            this.updateGameStatus('Ready - Player 1\'s Turn');
            this.logGameEvent(`Game started with ${this.gameConfig.playerCount} players`);
            this.logGameEvent('Player 1\'s turn begins');

        } catch (error) {
            console.error('Failed to load map and initialize game:', error);
            this.showToast('Error', 'Failed to load game', 'error');
        }
    }

    /**
     * Initialize game state
     */
    private initializeGameState(): void {
        this.currentPlayer = 1;
        this.currentTurn = 1;
        this.selectedUnit = null;
        
        // Update turn counter
        this.updateTurnCounter();
        
        // TODO: Initialize WASM game engine here
        console.log('Game state initialized');
    }

    /**
     * Bind page-specific events (required by BasePage)
     */
    protected bindSpecificEvents(): void {
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
     * Load map data from hidden element
     */
    private loadMapDataFromElement(): any {
        try {
            const mapDataElement = document.getElementById('map-data-json');
            if (mapDataElement && mapDataElement.textContent) {
                const mapData = JSON.parse(mapDataElement.textContent);
                return mapData && mapData !== null ? mapData : null;
            }
            return null;
        } catch (error) {
            console.error('Error parsing map data:', error);
            return null;
        }
    }

    /**
     * Game action handlers
     */
    private endTurn(): void {
        console.log(`Player ${this.currentPlayer} ends turn`);
        
        // Move to next player
        this.currentPlayer = (this.currentPlayer % this.gameConfig.playerCount) + 1;
        
        // If back to player 1, increment turn counter
        if (this.currentPlayer === 1) {
            this.currentTurn++;
        }
        
        // Update UI
        this.updateGameStatus(`Ready - Player ${this.currentPlayer}'s Turn`);
        this.updateTurnCounter();
        this.logGameEvent(`Player ${this.currentPlayer}'s turn begins`);
        
        // Clear selection
        this.clearUnitSelection();
        
        // TODO: Call WASM endTurn() function
        this.showToast('Info', `Player ${this.currentPlayer}'s turn`, 'info');
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
        // TODO: Highlight valid move tiles
        this.showToast('Info', 'Click on a tile to move', 'info');
    }

    private selectAttackMode(): void {
        if (!this.selectedUnit) {
            this.showToast('Warning', 'Select a unit first', 'warning');
            return;
        }
        console.log('Attack mode selected for unit:', this.selectedUnit);
        // TODO: Highlight valid attack targets
        this.showToast('Info', 'Click on an enemy unit to attack', 'info');
    }

    private showAllPlayerUnits(): void {
        console.log(`Showing all units for Player ${this.currentPlayer}`);
        // TODO: Highlight all player units and center camera
        this.showToast('Info', `Showing all Player ${this.currentPlayer} units`, 'info');
    }

    private centerOnAction(): void {
        console.log('Centering on action');
        // TODO: Center camera on the most recent action or selected unit
        this.showToast('Info', 'Centering view', 'info');
    }

    /**
     * Unit selection management
     */
    private selectUnit(unit: any): void {
        this.selectedUnit = unit;
        console.log('Unit selected:', unit);
        
        // Update selected unit info panel
        this.updateSelectedUnitInfo(unit);
        
        // Show unit action buttons
        const unitInfoPanel = document.getElementById('selected-unit-info');
        if (unitInfoPanel) {
            unitInfoPanel.classList.remove('hidden');
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
     * UI update functions
     */
    private updateGameStatus(status: string): void {
        const statusElement = document.getElementById('game-status');
        if (statusElement) {
            statusElement.textContent = status;
            statusElement.className = 'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
        }
    }

    private updateTurnCounter(): void {
        const turnElement = document.getElementById('turn-counter');
        if (turnElement) {
            turnElement.textContent = `Turn ${this.currentTurn}`;
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

    public destroy(): void {
        if (this.mapViewer) {
            this.mapViewer.destroy();
            this.mapViewer = null;
        }
        
        this.map = null;
        this.currentMapId = null;
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

// Initialize page when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    const gameViewerPage = new GameViewerPage();
});