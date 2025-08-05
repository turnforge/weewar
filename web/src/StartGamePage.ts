import { BasePage } from '../lib/BasePage';
import { EventBus } from '../lib/EventBus';
import { PhaserWorldScene } from './phaser/PhaserWorldScene';
import { World } from './World';
import { LCMComponent } from '../lib/LCMComponent';
import { LifecycleController } from '../lib/LifecycleController';
import { WorldEventTypes } from './events';

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
class StartGamePage extends BasePage implements LCMComponent {
    private currentWorldId: string | null;
    private isLoadingWorld: boolean = false;
    private world: World;
    private playerCount: number;
    private gameConfig: GameConfiguration = {
        players: [],
        allowedUnits: [], // Will be populated with unit IDs
        turnTimeLimit: 0,
        teamMode: 'ffa'
    };
    
    // Component instances
    private worldScene: PhaserWorldScene

    // =============================================================================
    // LCMComponent Interface Implementation
    // =============================================================================

    /**
     * Phase 1: Initialize DOM and discover child components
     */
    performLocalInit(): Promise<LCMComponent[]> | LCMComponent[] {
        this.loadInitialState(); // Load initial state here since constructor calls this

        this.loadWorldData()
        
        // Subscribe to events BEFORE creating components - None here
        
        // Create child components
        this.createComponents();
        
        // Return child components for lifecycle management
        return [this.worldScene];
    }

    /**
     * Phase 3: Activate component when all dependencies are ready
     */
    async activate(): Promise<void> {
        // Bind events now that all components are ready
        this.bindPageSpecificEvents();
        this.worldScene.loadWorld(this.world);
        this.showToast('Success', 'World loaded successfully', 'success');
    }

    /**
     * Cleanup phase (called by lifecycle controller if needed)
     */
    deactivate(): void {
        // Remove event subscriptions
        this.removeSubscription(WorldEventTypes.WORLD_VIEWER_READY, null);
        
        this.destroy();
    }

    /**
     * Create PhaserWorldScene component instance
     */
    private createComponents(): void {
        // Create PhaserWorldScene component for preview
        const phaserContainer = this.ensureElement('#phaser-viewer-container', 'phaser-viewer-container');
        this.worldScene = new PhaserWorldScene(phaserContainer, this.eventBus, true);
    }

    /**
     * Internal method to bind page-specific events (called from activate() phase)
     */
    private bindPageSpecificEvents(): void {
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
        } else {
            console.error("World ID input element not found or has no value. Cannot load world.");
            this.showToast("Error", "Could not load world: World ID missing.", "error");
        }
    }

    /**
     * Load world data and coordinate between components
     */
    private loadWorldData(): void {
        // Load world data from the hidden JSON element
        const worldMetadataElement = document.getElementById('world-data-json');
        const worldTilesElement = document.getElementById('world-tiles-data-json');
        this.world = new World(this.eventBus).loadFromElement(worldMetadataElement!, worldTilesElement!);
        
        // Calculate player count from world units
        this.playerCount = this.world.playerCount

        // Initialize game configuration based on world
        this.initializeGameConfiguration();
        
        // Bind unit restriction events (units are now server-rendered)
        this.bindUnitRestrictionEvents();
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
    
    private handlePlayerConfigChange(event: Event, configType: 'type' | 'team'): void {
        const select = event.target as HTMLSelectElement;
        const playerId = parseInt(select.dataset.player || '0');
        const value = configType === 'team' ? parseInt(select.value) : select.value;
        
        const player = this.gameConfig.players.find(p => p.id === playerId);
        if (player) {
            if (configType === 'type') {
                player.type = value as PlayerType;
            } else if (configType === 'team') {
                player.team = value as number;
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
        
        this.validateGameConfiguration();
    }

    private handleTurnLimitChange(event: Event): void {
        const select = event.target as HTMLSelectElement;
        this.gameConfig.turnTimeLimit = parseInt(select.value);
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

        this.showToast('Success', 'Creating game...', 'success');
        
        // Call the CreateGame API
        const result = await this.callCreateGameAPI();
        
        // Redirect to the newly created game
        const gameViewerUrl = `/games/${result.gameId}/view`;
        this.showToast('Success', 'Game created! Redirecting...', 'success');
        
        // Small delay to show the toast
        setTimeout(() => {
            window.location.href = gameViewerUrl;
        }, 500);
    }

    // Call CreateGame API via gRPC gateway
    private async callCreateGameAPI(): Promise<{ gameId: string }> {
        // Prepare the request payload matching the updated proto structure
        const activePlayers = this.gameConfig.players.filter(p => p.type !== 'none');
        
        // Get game name from input field
        const gameNameInput = document.getElementById('game-name-title') as HTMLInputElement;
        const gameName = gameNameInput?.value?.trim() || 'New Game';

        const gameRequest = {
            game: {
                world_id: this.currentWorldId,
                name: gameName,
                description: 'Game created from StartGamePage',
                creator_id: 'default-user', // TODO: Get from authentication
                tags: [],
                config: {
                    players: activePlayers.map(p => ({
                        player_id: p.id,
                        player_type: p.type,
                        color: p.color,
                        team_id: p.team
                    })),
                    settings: {
                        allowed_units: this.gameConfig.allowedUnits.map(id => parseInt(id)),
                        turn_time_limit: this.gameConfig.turnTimeLimit,
                        team_mode: this.gameConfig.teamMode,
                        max_turns: 0 // Unlimited for now
                    }
                }
            }
        };

        // Make the gRPC gateway call
        const response = await fetch('/api/v1/games', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(gameRequest)
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`API call failed: ${response.status} - ${errorText}`);
        }

        const result = await response.json();
        
        if (!result.game || !result.game.id) {
            throw new Error('No game ID returned from server');
        }
        
        return { gameId: result.game.id };
    }

    public destroy(): void {
        // Clean up components
        if (this.worldScene) {
            this.worldScene.destroy();
            this.worldScene = null as any;
        }
        
        // Clean up world data
        this.world = null as any;
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

// Initialize page when DOM is ready using LifecycleController
document.addEventListener('DOMContentLoaded', async () => {
    // Create page instance (just basic setup)
    const page = new StartGamePage("StartGamePage");
    
    // Create lifecycle controller with debug logging
    const lifecycleController = new LifecycleController(page.eventBus, LifecycleController.DefaultConfig)
    
    // Start breadth-first initialization
    await lifecycleController.initializeFromRoot(page);
});
