import { BasePage } from './BasePage';
import { EventBus, EventTypes } from './EventBus';
import { WorldViewer } from './WorldViewer';
import { World } from './World';
import { ComponentLifecycle } from './ComponentLifecycle';
import { LifecycleController } from './LifecycleController';

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
class StartGamePage extends BasePage implements ComponentLifecycle {
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
        console.log('StartGamePage: Constructor starting...');
        super(); // BasePage will call initializeSpecificComponents() and bindSpecificEvents() 
        console.log('StartGamePage: Constructor completed - lifecycle will be managed externally');
    }

    /**
     * Load initial state (required by BasePage)
     * This method is called by BasePage constructor, but we're using external LifecycleController
     * so we make this a no-op and handle initialization through ComponentLifecycle interface
     */
    protected initializeSpecificComponents(): void {
        console.log('StartGamePage: initializeSpecificComponents() called by BasePage - doing minimal setup');
        this.loadInitialState(); // Load initial state here since constructor calls this
        console.log('StartGamePage: Actual component initialization will be handled by LifecycleController');
    }
    
    /**
     * Subscribe to WorldViewer events before component creation
     */
    private subscribeToWorldViewerEvents(): void {
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
    }

    /**
     * Create WorldViewer component instance
     */
    private createComponents(): void {
        // Create WorldViewer component for preview
        const worldViewerRoot = this.ensureElement('[data-component="world-viewer"]', 'world-viewer-root');
        console.log('StartGamePage: Creating WorldViewer with eventBus:', this.eventBus);
        this.worldViewer = new WorldViewer(worldViewerRoot, this.eventBus, true);
        
        console.log('StartGamePage: Components created');
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

    /**
     * Bind page-specific events (required by BasePage)
     * This method is called by BasePage constructor, but we're using external LifecycleController
     * so we make this a no-op and handle event binding in ComponentLifecycle.activate()
     */
    protected bindSpecificEvents(): void {
        console.log('StartGamePage: bindSpecificEvents() called by BasePage - deferred to activate() phase');
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
        if (!worldData || !worldData.units) {
            return 2; // Default fallback
        }
        
        // Find the highest player ID in world units
        let maxPlayer = 0;
        for (const unit of worldData.units) {
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
     * Load world data from the hidden JSON elements in the page
     * Now loads from both world metadata and world tiles/units data
     */
    private loadWorldDataFromElement(): any {
        try {
            // Load world metadata
            const worldMetadataElement = document.getElementById('world-data-json');
            const worldTilesElement = document.getElementById('world-tiles-data-json');
            
            console.log(`World metadata element found: ${worldMetadataElement ? 'YES' : 'NO'}`);
            console.log(`World tiles element found: ${worldTilesElement ? 'YES' : 'NO'}`);
            
            if (!worldMetadataElement || !worldTilesElement) {
                console.log('Missing required world data elements');
                return null;
            }
            
            // Parse world metadata
            let worldMetadata = null;
            if (worldMetadataElement.textContent) {
                console.log(`Raw world metadata: ${worldMetadataElement.textContent.substring(0, 200)}...`);
                worldMetadata = JSON.parse(worldMetadataElement.textContent);
            }
            
            // Parse world tiles/units data
            let worldTilesData = null;
            if (worldTilesElement.textContent) {
                console.log(`Raw world tiles data: ${worldTilesElement.textContent.substring(0, 200)}...`);
                worldTilesData = JSON.parse(worldTilesElement.textContent);
            }
            
            if (worldMetadata && worldTilesData) {
                // Combine into format expected by World.loadFromData()
                const combinedData = {
                    // World metadata
                    name: worldMetadata.name || 'Untitled World',
                    Name: worldMetadata.name || 'Untitled World', // Both for compatibility
                    id: worldMetadata.id,
                    
                    // Calculate dimensions from tiles if present
                    width: 40,  // Default
                    height: 40, // Default
                    
                    // World tiles and units
                    tiles: worldTilesData.tiles || [],
                    units: worldTilesData.units || []
                };
                
                // Calculate actual dimensions from tile bounds
                if (combinedData.tiles && combinedData.tiles.length > 0) {
                    let maxQ = 0, maxR = 0, minQ = 0, minR = 0;
                    combinedData.tiles.forEach((tile: any) => {
                        if (tile.q > maxQ) maxQ = tile.q;
                        if (tile.q < minQ) minQ = tile.q;
                        if (tile.r > maxR) maxR = tile.r;
                        if (tile.r < minR) minR = tile.r;
                    });
                    combinedData.width = maxQ - minQ + 1;
                    combinedData.height = maxR - minR + 1;
                }
                
                console.log('Combined world data created for StartGamePage');
                console.log(`World: ${combinedData.name}, Tiles: ${combinedData.tiles.length}, Units: ${combinedData.units.length}`);
                console.log(`Dimensions: ${combinedData.width}x${combinedData.height}`);
                
                return combinedData;
            }
            
            console.log('No valid world data found in page elements');
            return null;
        } catch (error) {
            console.error('Error parsing world data from page elements:', error);
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
            this.showToast('Success', 'Creating game...', 'success');
            
            // Call the CreateGame API
            const result = await this.callCreateGameAPI();
            
            // Redirect to the newly created game
            const gameViewerUrl = `/games/${result.gameId}/view`;
            console.log('Game created successfully, redirecting to:', gameViewerUrl);
            
            this.showToast('Success', 'Game created! Redirecting...', 'success');
            
            // Small delay to show the toast
            setTimeout(() => {
                window.location.href = gameViewerUrl;
            }, 500);
            
        } catch (error: any) {
            console.error('Failed to start game:', error);
            this.showToast('Error', `Failed to start game: ${error.message}`, 'error');
        }
    }

    // Call CreateGame API via gRPC gateway
    private async callCreateGameAPI(): Promise<{ gameId: string }> {
        // Prepare the request payload matching the updated proto structure
        const activePlayers = this.gameConfig.players.filter(p => p.type !== 'none');
        
        const gameRequest = {
            game: {
                world_id: this.currentWorldId,
                name: `Game from ${this.world?.name || 'World'}`,
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

        console.log('CreateGame API request:', gameRequest);
        
        // Make the gRPC gateway call
        const response = await fetch('/v1/games', {
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
        console.log('CreateGame API response:', result);
        
        if (!result.game || !result.game.id) {
            throw new Error('No game ID returned from server');
        }
        
        return { gameId: result.game.id };
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

    // =============================================================================
    // ComponentLifecycle Interface Implementation
    // =============================================================================

    /**
     * Phase 1: Initialize DOM and discover child components
     */
    initializeDOM(): ComponentLifecycle[] {
        console.log('StartGamePage: initializeDOM() - Phase 1');
        
        // Subscribe to events BEFORE creating components
        this.subscribeToWorldViewerEvents();
        
        // Create child components
        this.createComponents();
        
        console.log('StartGamePage: DOM initialized, returning child components');
        
        // Return child components for lifecycle management
        const childComponents: ComponentLifecycle[] = [];
        if (this.worldViewer) {
            childComponents.push(this.worldViewer);
        }
        return childComponents;
    }

    /**
     * Phase 2: Inject dependencies (none needed for StartGamePage)
     */
    injectDependencies(deps: Record<string, any>): void {
        console.log('StartGamePage: injectDependencies() - Phase 2', Object.keys(deps));
        // StartGamePage doesn't need external dependencies
    }

    /**
     * Phase 3: Activate component when all dependencies are ready
     */
    async activate(): Promise<void> {
        console.log('StartGamePage: activate() - Phase 3');
        
        // Bind events now that all components are ready
        this.bindPageSpecificEvents();
        
        console.log('StartGamePage: activation complete');
    }

    /**
     * Cleanup phase (called by lifecycle controller if needed)
     */
    deactivate(): void {
        console.log('StartGamePage: deactivate() - cleanup');
        this.destroy();
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
    console.log('DOM loaded, starting StartGamePage initialization...');
    
    // Create page instance (just basic setup)
    const startGamePage = new StartGamePage();
    
    // Create lifecycle controller with debug logging
    const lifecycleController = new LifecycleController({
        enableDebugLogging: true,
        phaseTimeoutMs: 15000, // Increased timeout for component loading
        continueOnError: false // Fail fast for debugging
    });
    
    // Set up lifecycle event logging
    lifecycleController.onLifecycleEvent((event) => {
        console.log(`[StartGame Lifecycle] ${event.type}: ${event.componentName} - ${event.phase}`, event.error || '');
    });
    
    // Start breadth-first initialization
    await lifecycleController.initializeFromRoot(startGamePage, 'StartGamePage');
    
    console.log('StartGamePage fully initialized via LifecycleController');
});
