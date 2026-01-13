import type { GameConfiguration, GamePlayer, IncomeConfig } from '../gen/wasmjs/lilbattle/v1/models/interfaces';
import { BasePage, EventBus, LCMComponent, LifecycleController } from '@panyam/tsappkit';
import { PhaserWorldScene } from './common/PhaserWorldScene';
import { World } from './common/World';
import { WorldEventTypes } from './common/events';

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
        teams: [],
        incomeConfigs: {
            startingCoins: 100,
            gameIncome: 100,
            landbaseIncome: 100,
            navalbaseIncome: 100,
            airportbaseIncome: 100,
            missilesiloIncome: 100,
            minesIncome: 100
        },
        settings: {
            allowedUnits: [],
            turnTimeLimit: 0,
            teamMode: 'ffa',
            maxTurns: 0
        }
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
        this.loadInitialState();

        // Only load world data and create components if a world is selected
        if (this.currentWorldId) {
            this.loadWorldData();

            // Subscribe to events BEFORE creating components - None here

            // Create child components
            this.createComponents();

            // Return child components for lifecycle management
            return [this.worldScene];
        }

        // No world selected - return empty array (no child components)
        return [];
    }

    /**
     * Phase 3: Activate component when all dependencies are ready
     */
    async activate(): Promise<void> {
        // Bind events now that all components are ready
        this.bindPageSpecificEvents();

        // Only load world if one was selected
        if (this.currentWorldId && this.worldScene) {
            this.worldScene.loadWorld(this.world);
            this.showToast('Success', 'World loaded successfully', 'success');
        }

        // Dismiss splash screen
        super.dismissSplashScreen();
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
        // Create PhaserWorldScene component for preview - uses PhaserSceneView template with SceneId: "start-game-scene"
        const phaserContainer = this.ensureElement('#start-game-scene', 'start-game-scene');
        this.worldScene = new PhaserWorldScene(phaserContainer, this.eventBus, true);
    }

    /**
     * Internal method to bind page-specific events (called from activate() phase)
     */
    private bindPageSpecificEvents(): void {
        // Bind all start game buttons (desktop and mobile)
        const startGameButtons = document.querySelectorAll('[data-action="start-game"]');
        startGameButtons.forEach(button => {
            button.addEventListener('click', this.startGame.bind(this));
        });

        // Bind turn limit selector
        const turnLimitSelect = document.querySelector('[data-config="turn-limit"]');
        if (turnLimitSelect) {
            turnLimitSelect.addEventListener('change', this.handleTurnLimitChange.bind(this));
        }

        // Bind income input fields
        const incomeFields = ['starting-coins', 'game-income', 'landbase-income', 'navalbase-income', 'airportbase-income', 'missilesilo-income', 'mines-income'];
        incomeFields.forEach(field => {
            const input = document.querySelector(`[data-config="${field}"]`);
            if (input) {
                input.addEventListener('change', this.handleIncomeChange.bind(this));
            }
        });

        // Initialize bottom sheet for mobile
        this.initializeConfigBottomSheet();
    }

    /**
     * Initialize bottom sheet for mobile config panel
     */
    private initializeConfigBottomSheet(): void {
        const fab = document.getElementById('config-fab');
        const overlay = document.getElementById('config-overlay');
        const panel = document.getElementById('config-panel');
        const backdrop = document.getElementById('config-backdrop');
        const closeButton = document.getElementById('config-close');

        if (!fab || !overlay || !panel || !backdrop || !closeButton) {
            return; // Elements don't exist (probably on desktop)
        }

        // Open bottom sheet
        const openSheet = () => {
            overlay.classList.remove('hidden');
            // Force reflow to enable transition
            overlay.offsetHeight;
            panel.classList.remove('translate-y-full');
        };

        // Close bottom sheet
        const closeSheet = () => {
            panel.classList.add('translate-y-full');
            // Wait for animation to complete before hiding
            setTimeout(() => {
                overlay.classList.add('hidden');
            }, 300);
        };

        // Event listeners
        fab.addEventListener('click', openSheet);
        closeButton.addEventListener('click', closeSheet);
        backdrop.addEventListener('click', closeSheet);

        // ESC key to close
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && !overlay.classList.contains('hidden')) {
                closeSheet();
            }
        });
    }

    /** Load document data and set initial UI states */
    private loadInitialState(): void {
        const worldIdInput = document.getElementById("worldIdInput") as HTMLInputElement | null;
        const worldId = worldIdInput?.value.trim() || null;

        if (worldId) {
            this.currentWorldId = worldId;
        } else {
            // No world selected - this is a valid state, user needs to select a world
            this.currentWorldId = null;
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
     * Initialize game configuration from server-rendered HTML
     */
    private initializeGameConfiguration(): void {
        // Load player configuration from server-rendered elements
        // Use a Map to deduplicate by playerId (same template may be rendered in both desktop and mobile views)
        const playerMap = new Map<number, GamePlayer>();
        const playerElements = document.querySelectorAll('[data-config-section="players"] [data-player]');
        playerElements.forEach(el => {
            const playerId = parseInt((el as HTMLElement).dataset.player || '0');
            // Skip if we already have this player (deduplication)
            if (playerMap.has(playerId)) return;

            const typeSelect = el.querySelector('[data-config="type"]') as HTMLSelectElement;
            const teamSelect = el.querySelector('[data-config="team"]') as HTMLSelectElement;
            const coinsInput = el.querySelector('[data-config="coins"]') as HTMLInputElement;

            if (typeSelect && teamSelect && coinsInput) {
                playerMap.set(playerId, {
                    playerId: playerId,
                    userId: '', // Will be assigned by server when game starts
                    playerType: typeSelect.value,
                    color: '', // Color is handled by server rendering
                    teamId: parseInt(teamSelect.value),
                    name: `Player ${playerId}`,
                    isActive: true,
                    startingCoins: parseInt(coinsInput.value)
                });
            }
        });
        this.gameConfig.players = Array.from(playerMap.values());

        // Load allowed units from server-rendered checkboxes
        const unitCheckboxes = document.querySelectorAll('#unit-restriction-grid input[type="checkbox"]');
        this.gameConfig.settings!.allowedUnits = Array.from(unitCheckboxes)
            .filter(cb => (cb as HTMLInputElement).checked)
            .map(cb => parseInt((cb as HTMLInputElement).dataset.unit || '0'));

        // Load income configuration from server-rendered inputs
        this.initializeIncomeConfiguration();

        // Bind event listeners to server-rendered elements
        this.bindPlayerConfigurationEvents();
    }

    /**
     * Initialize income configuration from server-rendered HTML inputs
     */
    private initializeIncomeConfiguration(): void {
        const incomeFieldMap: Record<string, keyof IncomeConfig> = {
            'starting-coins': 'startingCoins',
            'game-income': 'gameIncome',
            'landbase-income': 'landbaseIncome',
            'navalbase-income': 'navalbaseIncome',
            'airportbase-income': 'airportbaseIncome',
            'missilesilo-income': 'missilesiloIncome',
            'mines-income': 'minesIncome'
        };

        for (const [dataConfig, configKey] of Object.entries(incomeFieldMap)) {
            const input = document.querySelector(`[data-config="${dataConfig}"]`) as HTMLInputElement;
            if (input) {
                const value = parseInt(input.value) || 0;
                this.gameConfig.incomeConfigs![configKey] = value;
            }
        }
    }

    /**
     * Bind event listeners to server-rendered player configuration elements
     */
    private bindPlayerConfigurationEvents(): void {
        const playersSection = document.querySelector('[data-config-section="players"]');
        if (!playersSection) return;

        // Bind events to all player config selects and inputs
        const typeSelects = playersSection.querySelectorAll('select[data-config="type"]');
        typeSelects.forEach(select => {
            select.addEventListener('change', (e) => this.handlePlayerConfigChange(e, 'type'));
        });

        const teamSelects = playersSection.querySelectorAll('select[data-config="team"]');
        teamSelects.forEach(select => {
            select.addEventListener('change', (e) => this.handlePlayerConfigChange(e, 'team'));
        });

        const coinsInputs = playersSection.querySelectorAll('input[data-config="coins"]');
        coinsInputs.forEach(input => {
            input.addEventListener('change', (e) => this.handlePlayerConfigChange(e, 'coins'));
        });
    }

    /**
     * Bind events to server-rendered unit restriction buttons
     */
    private bindUnitRestrictionEvents(): void {
        const unitButtons = document.querySelectorAll('.unit-restriction-button');
        unitButtons.forEach(button => {
            const checkbox = button.querySelector('input[type="checkbox"]') as HTMLInputElement;
            const mask = button.querySelector('.unit-mask') as HTMLElement;

            // Set initial mask state based on checkbox
            if (mask) {
                mask.style.opacity = checkbox?.checked ? '0' : '0.6';
            }

            // Handle button click (for clicks outside the checkbox)
            button.addEventListener('click', (e) => {
                // If user clicked directly on checkbox, let native behavior handle it
                // The checkbox's change event will handle the update
                if (e.target === checkbox) {
                    return;
                }

                e.preventDefault();
                if (checkbox) {
                    checkbox.checked = !checkbox.checked;
                    // Dispatch change event to trigger our handler
                    checkbox.dispatchEvent(new Event('change', { bubbles: true }));
                }
            });

            // Handle checkbox change (works for both direct clicks and button-triggered changes)
            if (checkbox) {
                checkbox.addEventListener('change', () => {
                    if (mask) {
                        mask.style.opacity = checkbox.checked ? '0' : '0.6';
                    }
                    this.handleUnitRestrictionChange({ target: checkbox } as any);
                });
            }
        });
    }
    
    private handlePlayerConfigChange(event: Event, configType: 'type' | 'team' | 'coins'): void {
        const target = event.target as HTMLSelectElement | HTMLInputElement;
        const playerId = parseInt(target.dataset.player || '0');

        const player = this.gameConfig.players?.find(p => p.playerId === playerId);
        if (player) {
            if (configType === 'type') {
                player.playerType = (target as HTMLSelectElement).value;
            } else if (configType === 'team') {
                player.teamId = parseInt((target as HTMLSelectElement).value);
            } else if (configType === 'coins') {
                const coins = parseInt((target as HTMLInputElement).value) || 300;
                player.startingCoins = coins;
            }
        }

        this.validateGameConfiguration();
    }

    private handleUnitRestrictionChange(event: Event): void {
        const checkbox = event.target as HTMLInputElement;
        const unitId = parseInt(checkbox.dataset.unit || '0');

        if (!this.gameConfig.settings) {
            this.gameConfig.settings = { allowedUnits: [], turnTimeLimit: 0, teamMode: 'ffa', maxTurns: 0 };
        }

        if (checkbox.checked) {
            if (!this.gameConfig.settings.allowedUnits.includes(unitId)) {
                this.gameConfig.settings.allowedUnits.push(unitId);
            }
        } else {
            this.gameConfig.settings.allowedUnits = this.gameConfig.settings.allowedUnits.filter(unit => unit !== unitId);
        }
        
        this.validateGameConfiguration();
    }

    private handleTurnLimitChange(event: Event): void {
        const select = event.target as HTMLSelectElement;
        if (this.gameConfig.settings) {
            this.gameConfig.settings.turnTimeLimit = parseInt(select.value);
        }
        this.validateGameConfiguration();
    }

    private handleIncomeChange(event: Event): void {
        const input = event.target as HTMLInputElement;
        const configType = input.dataset.config;
        const value = parseInt(input.value) || 0;

        if (!this.gameConfig.incomeConfigs) {
            this.gameConfig.incomeConfigs = {
                startingCoins: 100,
                gameIncome: 100,
                landbaseIncome: 100,
                navalbaseIncome: 100,
                airportbaseIncome: 100,
                missilesiloIncome: 100,
                minesIncome: 100
            };
        }

        switch (configType) {
            case 'starting-coins':
                this.gameConfig.incomeConfigs.startingCoins = value;
                break;
            case 'game-income':
                this.gameConfig.incomeConfigs.gameIncome = value;
                break;
            case 'landbase-income':
                this.gameConfig.incomeConfigs.landbaseIncome = value;
                break;
            case 'navalbase-income':
                this.gameConfig.incomeConfigs.navalbaseIncome = value;
                break;
            case 'airportbase-income':
                this.gameConfig.incomeConfigs.airportbaseIncome = value;
                break;
            case 'missilesilo-income':
                this.gameConfig.incomeConfigs.missilesiloIncome = value;
                break;
            case 'mines-income':
                this.gameConfig.incomeConfigs.minesIncome = value;
                break;
        }

        this.validateGameConfiguration();
    }

    private validateGameConfiguration(): boolean {
        const startButtons = document.querySelectorAll('[data-action="start-game"]') as NodeListOf<HTMLButtonElement>;
        let isValid = true;
        let errors: string[] = [];

        // Check if at least one unit type is allowed
        if (!this.gameConfig.settings?.allowedUnits || this.gameConfig.settings.allowedUnits.length === 0) {
            isValid = false;
            errors.push('At least one unit type must be allowed');
        }

        // Check if we have at least 2 active players
        const activePlayers = this.gameConfig.players?.filter(p => p.playerType !== 'none') || [];
        if (activePlayers.length < 2) {
            isValid = false;
            errors.push('At least 2 players are required');
        }

        // Update all start game buttons (desktop and mobile)
        startButtons.forEach(button => {
            button.disabled = !isValid;
            button.title = errors.length > 0 ? errors.join('; ') : '';
        });

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

        this.showToast('Info', 'Creating game...', 'info');

        try {
            // Call the CreateGame API
            const result = await this.callCreateGameAPI();

            // Redirect to the newly created game
            const gameViewerUrl = `/games/${result.gameId}/view`;
            this.showToast('Success', 'Game created! Redirecting...', 'success');

            // Small delay to show the toast
            setTimeout(() => {
                window.location.href = gameViewerUrl;
            }, 500);
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : 'Failed to create game';
            this.showToast('Error', errorMessage, 'error');
        }
    }

    // Call CreateGame API via gRPC gateway
    private async callCreateGameAPI(): Promise<{ gameId: string }> {
        // Prepare the request payload matching the updated proto structure
        const activePlayers = this.gameConfig.players?.filter(p => p.playerType !== 'none') || [];

        // Get game name from input field
        const gameNameInput = document.getElementById('game-name-title') as HTMLInputElement;
        const gameName = gameNameInput?.value?.trim() || 'New Game';

        // Get optional custom game ID
        const customGameIdInput = document.getElementById('custom-game-id') as HTMLInputElement;
        const customGameId = customGameIdInput?.value?.trim() || '';

        const gameRequest: Record<string, any> = {
            game: {
                world_id: this.currentWorldId,
                name: gameName,
                description: 'Game created from StartGamePage',
                creator_id: 'default-user', // TODO: Get from authentication
                tags: [],
                config: {
                    players: activePlayers.map(p => ({
                        player_id: p.playerId,
                        player_type: p.playerType,
                        color: p.color,
                        team_id: p.teamId,
                        name: p.name,
                        is_active: p.isActive,
                        starting_coins: p.startingCoins
                    })),
                    income_configs: this.gameConfig.incomeConfigs,
                    settings: {
                        allowed_units: this.gameConfig.settings?.allowedUnits || [],
                        turn_time_limit: this.gameConfig.settings?.turnTimeLimit || 0,
                        team_mode: this.gameConfig.settings?.teamMode || 'ffa',
                        max_turns: 0 // Unlimited for now
                    }
                }
            }
        };

        // Add custom game ID if provided
        if (customGameId) {
            gameRequest.game.id = customGameId;
        }

        // Make the gRPC gateway call
        const response = await fetch('/api/v1/games', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(gameRequest)
        });

        if (!response.ok) {
            let errorMessage = `Server error (${response.status})`;
            try {
                const errorData = await response.json();
                // gRPC gateway returns errors in { message: "...", code: N } format
                if (errorData.message) {
                    errorMessage = errorData.message;
                }
            } catch {
                // If not JSON, try plain text
                const errorText = await response.text();
                if (errorText) {
                    errorMessage = errorText;
                }
            }
            throw new Error(errorMessage);
        }

        const result = await response.json();

        // Check for field_errors (ID conflict)
        if (result.field_errors && Object.keys(result.field_errors).length > 0) {
            const suggestedId = result.field_errors['id'] || '';
            const gameNameInput = document.getElementById('game-name-title') as HTMLInputElement;
            const gameName = gameNameInput?.value?.trim() || 'New Game';

            // Redirect back to StartGamePage with error and suggested ID
            const redirectUrl = `/games/new?worldId=${encodeURIComponent(this.currentWorldId || '')}` +
                `&gameId=${encodeURIComponent(suggestedId)}` +
                `&gameName=${encodeURIComponent(gameName)}` +
                `&error=${encodeURIComponent('ID already exists. Suggested: ' + suggestedId)}`;
            window.location.href = redirectUrl;
            throw new Error('ID conflict - redirecting');
        }

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

// Type definitions for local game configuration state (will be mapped to proto types)
interface GameConfigurationLocal {
    players: PlayerLocal[];
    allowedUnits: string[];
    turnTimeLimit: number; // seconds, 0 = no limit
    teamMode: 'ffa' | 'teams';
    incomeConfig: {
        landbaseIncome: number;
        navalbaseIncome: number;
        airportbaseIncome: number;
        missilesiloIncome: number;
        minesIncome: number;
    };
}

interface PlayerLocal {
    id: number;
    color: string;
    type: PlayerType;
    team: number;
    startingCoins: number;
}

type PlayerType = 'human' | 'ai' | 'open' | 'none';

StartGamePage.loadAfterPageLoaded("StartGamePage", StartGamePage, "StartGamePage")
