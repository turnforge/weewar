import { BasePage } from './BasePage';
import { EventBus, EventTypes } from './EventBus';
import { WorldViewer } from './WorldViewer';
import { World } from './World';

/**
 * Start Game Page - Orchestrator for game configuration functionality
 * Responsible for:
 * - World data loading and preview coordination
 * - Game configuration management
 * - Player configuration handling
 * - Game creation workflow
 * 
 * Does NOT handle:
 * - Direct DOM manipulation (delegated to components)
 * - Phaser management (delegated to WorldViewer)
 * - Game logic (delegated to game engine)
 */
class StartGamePage extends BasePage {
    private currentWorldId: string | null;
    private isLoadingWorld: boolean = false;
    private world: World | null = null;
    private playerCount: number = 2; // Default to 2, will be updated from world data
    private gameConfig: GameConfiguration = {
        players: [],
        allowedUnits: [], // Will be populated with unit IDs
        turnTimeLimit: 0,
        teamMode: 'ffa'
    };
    
    // Component instances
    private worldViewer: WorldViewer | null = null;

    constructor() {
        super();
        this.loadInitialState();
        this.initializeSpecificComponents();
        this.bindSpecificEvents();
    }

    protected initializeSpecificComponents(): void {
        // Initialize components immediately
        this.initializeComponents();
    }
    
    /**
     * Initialize page components using the established component architecture
     */
    private initializeComponents(): void {
        try {
            console.log('Initializing StartGamePage components');
            
            // Subscribe to WorldViewer ready event BEFORE creating the component
            console.log('StartGamePage: Subscribing to world-viewer-ready event');
            this.eventBus.subscribe('world-viewer-ready', () => {
                console.log('StartGamePage: WorldViewer is ready, loading world data...');
                if (this.currentWorldId) {
                  // Give Phaser time to fully initialize webgl context and scene
                  setTimeout(async () => {
                    await this.loadWorldData()
                  }, 10)
                }
            }, 'start-game-page');
            
            // Create WorldViewer component for preview
            const worldViewerRoot = this.ensureElement('[data-component="world-viewer"]', 'world-viewer-root');
            console.log('StartGamePage: Creating WorldViewer with eventBus:', this.eventBus);
            this.worldViewer = new WorldViewer(worldViewerRoot, this.eventBus, true);
            
            console.log('StartGamePage components initialized');
            
        } catch (error) {
            console.error('Failed to initialize components:', error);
            this.showToast('Error', 'Failed to initialize page components', 'error');
        }
    }
    
    /**
     * Ensure an element exists, create if missing
     */
    private ensureElement(selector: string, fallbackId: string): HTMLElement {
        let element = document.querySelector(selector) as HTMLElement;
        if (!element) {
            console.warn(`Element not found: ${selector}, creating fallback`);
            element = document.createElement('div');
            element.id = fallbackId;
            element.className = 'w-full h-full';
            const mainContainer = document.querySelector('main') || document.body;
            mainContainer.appendChild(element);
        }
        return element;
    }

    protected bindSpecificEvents(): void {
        // Bind start game button
        const startGameButton = document.querySelector('[data-action="start-game"]');
        if (startGameButton) {
            startGameButton.addEventListener('click', this.startGame.bind(this));
        }

        // Bind turn limit selector
        const turnLimitSelect = document.querySelector('[data-config="turn-limit"]');
        if (turnLimitSelect) {
            turnLimitSelect.addEventListener('change', this.handleTurnLimitChange.bind(this));
        }
    }

    /** Load document data and set initial UI states */
    private loadInitialState(): void {
        const worldIdInput = document.getElementById("worldIdInput") as HTMLInputElement | null;
        const worldId = worldIdInput?.value.trim() || null;

        if (worldId) {
            this.currentWorldId = worldId;
            console.log(`Found World ID: ${this.currentWorldId}. Will load data after Phaser initialization.`);
        } else {
            console.error("World ID input element not found or has no value. Cannot load world.");
            this.showToast("Error", "Could not load world: World ID missing.", "error");
        }
    }

    /**
     * Load world data and coordinate between components
     */
    private async loadWorldData(): Promise<void> {
        try {
            console.log(`StartGamePage: Loading world data...`);
            
            // Load world data from the hidden JSON element
            const worldData = this.loadWorldDataFromElement();
            
            if (worldData) {
                this.world = World.deserialize(worldData);
                console.log('World data loaded successfully');
                
                // Calculate player count from world units
                this.playerCount = this.calculatePlayerCountFromWorld(worldData);
                console.log('Detected player count:', this.playerCount);
                
                // Initialize game configuration based on world
                this.initializeGameConfiguration();
                
                // Bind unit restriction events (units are now server-rendered)
                this.bindUnitRestrictionEvents();
                
                // Use WorldViewer component to load the world
                if (this.worldViewer) {
                    await this.worldViewer.loadWorld(worldData);
                    this.showToast('Success', 'World loaded successfully', 'success');
                } else {
                    console.warn('WorldViewer component not available');
                }
                
            } else {
                console.error('No world data found');
                this.showToast('Error', 'No world data found', 'error');
            }
            
        } catch (error) {
            console.error('Failed to load world data:', error);
            this.showToast('Error', 'Failed to load world data', 'error');
        }
    }
    
    /**
     * Calculate player count from world units
     */
    private calculatePlayerCountFromWorld(worldData: any): number {
        if (!worldData || !worldData.world_units) {
            return 2; // Default fallback
        }
        
        // Find the highest player ID in world units
        let maxPlayer = 0;
        for (const unit of worldData.world_units) {
            if (unit.player && unit.player > maxPlayer) {
                maxPlayer = unit.player;
            }
        }
        
        // Player IDs are 1-based, so player count is maxPlayer
        // Ensure minimum of 2 players
        return Math.max(2, maxPlayer);
    }
    
    /**
     * Initialize game configuration based on detected player count
     */
    private initializeGameConfiguration(): void {
        const playerColors = ['red', 'blue', 'green', 'yellow', 'purple', 'orange'];
        
        this.gameConfig.players = [];
        for (let i = 0; i < this.playerCount; i++) {
            this.gameConfig.players.push({
                id: i + 1,
                color: playerColors[i % playerColors.length],
                type: i === 0 ? 'human' : 'ai', // Player 1 is human, others are AI
                team: i + 1 // Each player starts on their own team
            });
        }
        
        // Initialize all units as allowed by default (get from server-rendered checkboxes)
        const unitCheckboxes = document.querySelectorAll('#unit-restriction-grid input[type="checkbox"]');
        this.gameConfig.allowedUnits = Array.from(unitCheckboxes).map(cb => (cb as HTMLInputElement).dataset.unit || '');
        
        // Update the player configuration UI
        this.updatePlayerConfigurationUI();
    }
    
    /**
     * Update the player configuration UI elements
     */
    private updatePlayerConfigurationUI(): void {
        const playersSection = document.querySelector('[data-config-section="players"]');
        if (!playersSection) return;
        
        // Find the players container
        const playersContainer = playersSection.querySelector('.space-y-3');
        if (!playersContainer) return;
        
        // Clear existing player elements
        playersContainer.innerHTML = '';
        
        // Create player configuration elements
        for (let i = 0; i < this.playerCount; i++) {
            const player = this.gameConfig.players[i];
            const playerElement = this.createPlayerConfigElement(player, i);
            playersContainer.appendChild(playerElement);
        }
    }
    
    /**
     * Create a player configuration element
     */
    private createPlayerConfigElement(player: Player, index: number): HTMLElement {
        const div = document.createElement('div');
        div.className = 'flex items-center justify-between p-3 border border-gray-200 dark:border-gray-600 rounded-lg';
        
        const colorClass = this.getPlayerColorClass(player.color);
        
        div.innerHTML = `
            <div class="flex items-center space-x-2">
                <div class="w-4 h-4 ${colorClass} rounded-full border border-gray-300"></div>
                <span class="text-sm font-medium text-gray-900 dark:text-white">Player ${player.id}</span>
            </div>
            <div class="flex items-center space-x-2">
                <select class="text-xs border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white" 
                        data-player="${player.id}" 
                        data-config="type"
                        ${index === 0 ? 'disabled' : ''}>
                    <option value="human" ${player.type === 'human' ? 'selected' : ''}>Human</option>
                    <option value="ai" ${player.type === 'ai' ? 'selected' : ''}>AI</option>
                    <option value="open" disabled>Open Invite</option>
                </select>
                <select class="text-xs border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white" 
                        data-player="${player.id}" 
                        data-config="team">
                    <option value="0">None</option>
                    ${this.generateTeamOptions(player.team)}
                </select>
            </div>
        `;
        
        // Bind event listeners
        const typeSelect = div.querySelector('[data-config="type"]') as HTMLSelectElement;
        const teamSelect = div.querySelector('[data-config="team"]') as HTMLSelectElement;
        
        if (typeSelect) {
            typeSelect.addEventListener('change', (e) => this.handlePlayerConfigChange(e, 'type'));
        }
        if (teamSelect) {
            teamSelect.addEventListener('change', (e) => this.handlePlayerConfigChange(e, 'team'));
        }
        
        return div;
    }
    
    /**
     * Generate team options HTML based on player count
     */
    private generateTeamOptions(selectedTeam: number): string {
        let options = '';
        for (let i = 1; i <= this.playerCount; i++) {
            options += `<option value="${i}" ${selectedTeam === i ? 'selected' : ''}>Team ${i}</option>`;
        }
        return options;
    }
    
    /**
     * Get CSS class for player color
     */
    private getPlayerColorClass(color: string): string {
        const colorWorld: { [key: string]: string } = {
            'red': 'bg-red-500',
            'blue': 'bg-blue-500',
            'green': 'bg-green-500',
            'yellow': 'bg-yellow-500',
            'purple': 'bg-purple-500',
            'orange': 'bg-orange-500'
        };
        return colorWorld[color] || 'bg-gray-500';
    }
    
    /**
     * Bind events to server-rendered unit restriction buttons
     */
    private bindUnitRestrictionEvents(): void {
        const unitButtons = document.querySelectorAll('.unit-restriction-button');
        unitButtons.forEach(button => {
            button.addEventListener('click', (e) => {
                e.preventDefault();
                const checkbox = button.querySelector('input[type="checkbox"]') as HTMLInputElement;
                const mask = button.querySelector('.unit-mask') as HTMLElement;
                
                checkbox.checked = !checkbox.checked;
                mask.style.opacity = checkbox.checked ? '0' : '0.6';
                
                // Trigger the restriction change handler
                const syntheticEvent = new Event('change');
                Object.defineProperty(syntheticEvent, 'target', {
                    writable: false,
                    value: checkbox
                });
                this.handleUnitRestrictionChange(syntheticEvent);
            });
        });
    }
    
    
    /**
     * Load world data from the hidden JSON element in the page
     */
    private loadWorldDataFromElement(): any {
        try {
            const worldDataElement = document.getElementById('world-data-json');
            console.log(`World data element found: ${worldDataElement ? 'YES' : 'NO'}`);
            
            if (worldDataElement && worldDataElement.textContent) {
                console.log(`Raw world data content: ${worldDataElement.textContent.substring(0, 200)}...`);
                const worldData = JSON.parse(worldDataElement.textContent);
                
                if (worldData && worldData !== null) {
                    console.log('World data found in page element');
                    return worldData;
                }
            }
            console.log('No world data found in page element');
            return null;
        } catch (error) {
            console.error('Error parsing world data from page element:', error);
            return null;
        }
    }

    private handlePlayerConfigChange(event: Event, configType: 'type' | 'team'): void {
        const select = event.target as HTMLSelectElement;
        const playerId = parseInt(select.dataset.player || '0');
        const value = configType === 'team' ? parseInt(select.value) : select.value;
        
        const player = this.gameConfig.players.find(p => p.id === playerId);
        if (player) {
            if (configType === 'type') {
                player.type = value as PlayerType;
                console.log(`Player ${playerId} type changed to: ${value}`);
            } else if (configType === 'team') {
                player.team = value as number;
                console.log(`Player ${playerId} team changed to: ${value}`);
            }
        }
        
        this.validateGameConfiguration();
    }

    private handleUnitRestrictionChange(event: Event): void {
        const checkbox = event.target as HTMLInputElement;
        const unitId = checkbox.dataset.unit || '';
        
        if (checkbox.checked) {
            if (!this.gameConfig.allowedUnits.includes(unitId)) {
                this.gameConfig.allowedUnits.push(unitId);
            }
        } else {
            this.gameConfig.allowedUnits = this.gameConfig.allowedUnits.filter(unit => unit !== unitId);
        }
        
        console.log('Allowed units updated:', this.gameConfig.allowedUnits);
        this.validateGameConfiguration();
    }

    private handleTurnLimitChange(event: Event): void {
        const select = event.target as HTMLSelectElement;
        this.gameConfig.turnTimeLimit = parseInt(select.value);
        console.log('Turn time limit changed to:', this.gameConfig.turnTimeLimit);
        this.validateGameConfiguration();
    }


    private validateGameConfiguration(): boolean {
        const startButton = document.querySelector('[data-action="start-game"]') as HTMLButtonElement;
        let isValid = true;
        let errors: string[] = [];

        // Check if at least one unit type is allowed
        if (this.gameConfig.allowedUnits.length === 0) {
            isValid = false;
            errors.push('At least one unit type must be allowed');
        }

        // Check if we have at least 2 active players
        const activePlayers = this.gameConfig.players.filter(p => p.type !== 'none');
        if (activePlayers.length < 2) {
            isValid = false;
            errors.push('At least 2 players are required');
        }

        if (startButton) {
            startButton.disabled = !isValid;
            startButton.title = errors.length > 0 ? errors.join('; ') : '';
        }

        return isValid;
    }

    private async startGame(): Promise<void> {
        if (!this.validateGameConfiguration()) {
            this.showToast('Error', 'Please fix configuration errors before starting the game', 'error');
            return;
        }

        if (!this.currentWorldId) {
            this.showToast('Error', 'No world selected', 'error');
            return;
        }

        try {
            console.log('Starting game with configuration:', this.gameConfig);
            
            // Build URL parameters for GameViewerPage
            const urlParams = new URLSearchParams();
            
            // Add player count
            const activePlayers = this.gameConfig.players.filter(p => p.type !== 'none');
            urlParams.set('playerCount', activePlayers.length.toString());
            
            // Add turn time limit (if set)
            if (this.gameConfig.turnTimeLimit > 0) {
                urlParams.set('maxTurns', '0'); // For now, we'll use maxTurns=0 (unlimited)
                urlParams.set('turnTimeLimit', this.gameConfig.turnTimeLimit.toString());
            }
            
            // Add allowed units as query parameters
            this.gameConfig.allowedUnits.forEach(unitId => {
                urlParams.set(`unit_${unitId}`, 'allowed');
            });
            
            // Add team configuration
            activePlayers.forEach(player => {
                urlParams.set(`player_${player.id}_type`, player.type);
                urlParams.set(`player_${player.id}_team`, player.team.toString());
            });
            
            // Redirect to GameViewerPage with world and configuration
            const gameViewerUrl = `/games/${this.currentWorldId}/view?${urlParams.toString()}`;
            
            console.log('Redirecting to GameViewer:', gameViewerUrl);
            this.showToast('Success', 'Starting game...', 'success');
            
            // Small delay to show the toast
            setTimeout(() => {
                window.location.href = gameViewerUrl;
            }, 500);
            
        } catch (error) {
            console.error('Failed to start game:', error);
            this.showToast('Error', 'Failed to start game', 'error');
        }
    }

    // Placeholder for future CreateGame API call
    private async callCreateGameAPI(): Promise<any> {
        // This will eventually call the CreateGame RPC endpoint
        const gameRequest = {
            worldId: this.currentWorldId,
            players: this.gameConfig.players.filter(p => p.type !== 'none').map(p => ({
                playerId: p.id,
                playerType: p.type,
                color: p.color,
                teamId: p.team
            })),
            gameSettings: {
                allowedUnits: this.gameConfig.allowedUnits,
                turnTimeLimit: this.gameConfig.turnTimeLimit,
                teamMode: this.gameConfig.teamMode
            }
        };

        console.log('CreateGame RPC request would be:', gameRequest);
        
        // TODO: Replace with actual gRPC call
        // return await grpcClient.createGame(gameRequest);
        
        return { gameId: 'placeholder-game-id' };
    }

    public destroy(): void {
        // Clean up components
        if (this.worldViewer) {
            this.worldViewer.destroy();
            this.worldViewer = null;
        }
        
        // Clean up world data
        this.world = null;
        this.currentWorldId = null;
    }
}

// Type definitions for game configuration
interface GameConfiguration {
    players: Player[];
    allowedUnits: string[];
    turnTimeLimit: number; // seconds, 0 = no limit
    teamMode: 'ffa' | 'teams';
}

interface Player {
    id: number;
    color: string;
    type: PlayerType;
    team: number;
}

type PlayerType = 'human' | 'ai' | 'open' | 'none';

document.addEventListener('DOMContentLoaded', () => {
    const startGamePage = new StartGamePage();
});
