import { BaseComponent } from '../lib/Component';
import { EventBus } from '../lib/EventBus';
import { LCMComponent } from '../lib/LCMComponent';
import { GameState } from './GameState';
import { World } from './World';
import { 
    Unit, 
    GameOption,
    MoveOption,
    AttackOption,
    EndTurnOption,
    BuildUnitOption,
    CaptureBuildingOption,
    GetOptionsAtResponse
} from '../gen/wasm-clients/weewar/v1/models';

/**
 * TurnOptionsPanel displays available turn options at a selected position
 * 
 * This component shows:
 * - Available movement options with paths and costs
 * - Attack options with damage estimates
 * - End turn option when available
 * - Build/capture options when applicable
 * 
 * Similar to the CLI's "options" command, this provides a clear view
 * of all available actions at the current position.
 */
export class TurnOptionsPanel extends BaseComponent implements LCMComponent {
    private isUIBound = false;
    private isActivated = false;
    private gameState: GameState | null = null;
    private world: World | null = null;
    private currentOptions: GameOption[] = [];
    private selectedPosition: { q: number; r: number } | null = null;
    private selectedUnit: Unit | null = null;

    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('turn-options-panel', rootElement, eventBus, debugMode);
    }

    // LCMComponent Phase 1: Initialize DOM structure
    public performLocalInit(): LCMComponent[] {
        if (this.isUIBound) {
            this.log('Already bound to DOM, skipping');
            return [];
        }

        this.log('Binding TurnOptionsPanel to DOM using template');
        this.isUIBound = true;
        
        this.log('TurnOptionsPanel bound to DOM successfully');
        
        // This is a leaf component - no children
        return [];
    }

    // Phase 2: Setup dependencies
    public setupDependencies(): void {
        this.log('Setting up TurnOptionsPanel dependencies');
        
        // Dependencies will be set by parent component
        // GameState and World are managed by GameViewerPage
    }
    
    /**
     * Set the GameState dependency
     */
    public setGameState(gameState: GameState): void {
        this.gameState = gameState;
        this.log('GameState dependency set');
    }
    
    /**
     * Set the World dependency
     */
    public setWorld(world: World): void {
        this.world = world;
        this.log('World dependency set');
    }

    // Phase 3: Activate component
    public activate(): void {
        if (this.isActivated) {
            this.log('Already activated, skipping');
            return;
        }

        this.log('Activating TurnOptionsPanel');
        this.isActivated = true;
        
        // Show initial empty state
        this.showEmptyState();
        
        this.log('TurnOptionsPanel activated successfully');
    }

    // Phase 4: Deactivate component
    public deactivate(): void {
        this.log('Deactivating TurnOptionsPanel');
        this.currentOptions = [];
        this.selectedPosition = null;
        this.selectedUnit = null;
        this.isActivated = false;
        this.log('TurnOptionsPanel deactivated');
    }

    /**
     * Handle unit selection with pre-fetched options (avoids duplicate RPC)
     */
    public handleUnitSelectionWithOptions(q: number, r: number, unit: Unit, response: GetOptionsAtResponse): void {
        this.log(`Unit selected at (${q}, ${r}) with pre-fetched options`);
        this.selectedPosition = { q, r };
        this.selectedUnit = unit;
        
        if (!response || !response.options) {
            this.showEmptyOptions();
            return;
        }
        
        // Process and display options
        this.processOptions(response);
    }
    
    /**
     * Handle unit selection - fetch and display options (legacy, makes own RPC)
     */
    public async handleUnitSelection(q: number, r: number, unit: Unit): Promise<void> {
        this.log(`Unit selected at (${q}, ${r})`);
        this.selectedPosition = { q, r };
        this.selectedUnit = unit;
        
        // Show loading state
        this.showLoadingState();
        
        try {
            // Get options from GameState
            if (!this.gameState) {
                this.showError('GameState not available');
                return;
            }

            const response = await this.gameState.getOptionsAt(q, r);
            
            if (!response || !response.options) {
                this.showEmptyOptions();
                return;
            }

            // Process and display options
            this.processOptions(response);
        } catch (error) {
            this.log('Error fetching options:', error);
            this.showError('Failed to fetch options');
        }
    }

    /**
     * Handle tile selection - fetch and display options if there's a unit
     */
    public async handleTileSelection(q: number, r: number): Promise<void> {
        this.log(`Tile selected at (${q}, ${r})`);
        
        // Check if there's a unit at this position
        const unit = this.world?.getUnitAt(q, r);
        if (unit) {
            await this.handleUnitSelection(q, r, unit);
        } else {
            this.clearOptions();
        }
    }

    /**
     * Process options from the server response
     */
    private processOptions(response: GetOptionsAtResponse): void {
        // Options are already sorted by the server
        this.currentOptions = response.options || [];
        
        // Display the processed options
        this.displayOptions();
    }

    /**
     * Get the type of an option
     */
    private getOptionType(option: any): string {
        if (option.move) return 'move';
        if (option.attack) return 'attack';
        if (option.endTurn) return 'endTurn';
        if (option.build) return 'build';
        if (option.capture) return 'capture';
        return 'unknown';
    }

    /**
     * Extract path coordinates from a MoveOption
     */
    private extractPathCoords(moveOption: MoveOption): number[] | undefined {
        if (!moveOption.reconstructedPath || !moveOption.reconstructedPath.edges) {
            return undefined;
        }
        
        const coords: number[] = [];
        const edges = moveOption.reconstructedPath.edges;
        
        // Add the starting position (from the first edge)
        if (edges.length > 0) {
            coords.push(edges[0].fromQ, edges[0].fromR);
            
            // Add all the destination positions
            for (const edge of edges) {
                coords.push(edge.toQ, edge.toR);
            }
        }
        
        return coords.length >= 4 ? coords : undefined;
    }

    /**
     * Display the current options
     */
    private displayOptions(): void {
        const container = this.findElement('#options-list');
        if (!container) return;

        if (this.currentOptions.length === 0) {
            this.showEmptyOptions();
            return;
        }

        // Hide empty state, show options
        const emptyState = this.findElement('#no-options-selected');
        const optionsContainer = this.findElement('#options-container');
        if (emptyState) emptyState.classList.add('hidden');
        if (optionsContainer) optionsContainer.classList.remove('hidden');

        // Update header
        const headerElement = this.findElement('#options-header');
        if (headerElement && this.selectedPosition) {
            const unitText = this.selectedUnit ? ` (Unit ${this.selectedUnit.unitType})` : '';
            headerElement.textContent = `Options at (${this.selectedPosition.q}, ${this.selectedPosition.r})${unitText}`;
        }

        // Build options HTML
        let optionsHTML = '';
        this.currentOptions.forEach((option, index) => {
            const optionType = this.getOptionType(option);
            const iconClass = this.getOptionIcon(optionType);
            const colorClass = this.getOptionColor(optionType);
            
            let description = '';
            let details = '';
            
            if (option.move) {
                description = `Move to (${option.move.q || 0}, ${option.move.r || 0})`;
                if (option.move.movementCost !== undefined) {
                    details += `<span class="text-xs text-gray-500 dark:text-gray-400">Cost: ${option.move.movementCost}</span>`;
                }
            } else if (option.attack) {
                description = `Attack unit at (${option.attack.q || 0}, ${option.attack.r || 0})`;
                if (option.attack.damageEstimate !== undefined) {
                    details += `<span class="text-xs text-red-500 dark:text-red-400">Damage: ~${option.attack.damageEstimate}</span>`;
                }
            } else if (option.endTurn) {
                description = 'End Turn';
            } else if (option.build) {
                description = `Build unit (type ${option.build.unitType})`;
                if (option.build.cost !== undefined) {
                    details += `<span class="text-xs text-gray-500 dark:text-gray-400">Cost: ${option.build.cost}</span>`;
                }
            } else if (option.capture) {
                description = 'Capture';
            }

            optionsHTML += `
                <div class="option-item p-3 mb-2 rounded-lg bg-gray-50 dark:bg-gray-700 hover:bg-gray-100 dark:hover:bg-gray-600 cursor-pointer transition-colors"
                     data-option-index="${index}">
                    <div class="flex items-start">
                        <span class="${iconClass} ${colorClass} mr-3 text-lg">${this.getOptionEmoji(optionType)}</span>
                        <div class="flex-1">
                            <div class="font-medium text-sm text-gray-900 dark:text-white">
                                ${description}
                            </div>
                            ${details ? `<div class="mt-1">${details}</div>` : ''}
                        </div>
                    </div>
                </div>
            `;
        });

        container.innerHTML = optionsHTML;

        // Add click handlers
        container.querySelectorAll('.option-item').forEach(item => {
            item.addEventListener('click', (e) => {
                const index = parseInt((e.currentTarget as HTMLElement).dataset.optionIndex || '0');
                this.handleOptionClick(index);
            });
        });
    }

    /**
     * Get icon for option type
     */
    private getOptionIcon(type: string): string {
        switch (type) {
            case 'move': return 'text-blue-500';
            case 'attack': return 'text-red-500';
            case 'endTurn': return 'text-green-500';
            case 'build': return 'text-yellow-500';
            case 'capture': return 'text-purple-500';
            default: return 'text-gray-500';
        }
    }

    /**
     * Get color class for option type
     */
    private getOptionColor(type: string): string {
        return this.getOptionIcon(type);
    }

    /**
     * Get emoji for option type
     */
    private getOptionEmoji(type: string): string {
        switch (type) {
            case 'move': return '➡️';
            case 'attack': return '⚔️';
            case 'endTurn': return '✅';
            case 'build': return '🏗️';
            case 'capture': return '🏳️';
            default: return '❓';
        }
    }

    /**
     * Handle option click
     */
    private handleOptionClick(index: number): void {
        const option = this.currentOptions[index];
        if (!option) return;

        const optionType = this.getOptionType(option);
        this.log(`Option clicked: ${optionType}`);
        
        // Clear any existing paths
        this.eventBus.emit('clear-path-visualization', {}, this, null);
        
        // If this is a move option with a reconstructed path, visualize it
        if (option.move) {
            const pathCoords = this.extractPathCoords(option.move);
            if (pathCoords && pathCoords.length >= 4) {
                this.eventBus.emit('show-path-visualization', {
                    coords: pathCoords,
                    color: 0x00ff00, // Green for movement
                    thickness: 4
                }, this, null);
            }
        }
        
        // Emit event for the game to handle the action
        this.eventBus.emit('turn-option-selected', {
            option: option,
            position: this.selectedPosition,
            unit: this.selectedUnit
        }, this, null);
    }

    /**
     * Clear options display
     */
    public clearOptions(): void {
        this.currentOptions = [];
        this.selectedPosition = null;
        this.selectedUnit = null;
        this.showEmptyState();
        
        // Clear any path visualization
        this.eventBus.emit('clear-path-visualization', {}, this, null);
    }

    /**
     * Show empty state
     */
    private showEmptyState(): void {
        const emptyState = this.findElement('#no-options-selected');
        const optionsContainer = this.findElement('#options-container');
        if (emptyState) emptyState.classList.remove('hidden');
        if (optionsContainer) optionsContainer.classList.add('hidden');
    }

    /**
     * Show loading state
     */
    private showLoadingState(): void {
        const container = this.findElement('#options-list');
        if (container) {
            container.innerHTML = `
                <div class="text-center py-4">
                    <div class="text-gray-500 dark:text-gray-400">Loading options...</div>
                </div>
            `;
        }
    }

    /**
     * Show empty options message
     */
    private showEmptyOptions(): void {
        const container = this.findElement('#options-list');
        if (container) {
            container.innerHTML = `
                <div class="text-center py-4">
                    <div class="text-gray-500 dark:text-gray-400">No options available</div>
                </div>
            `;
        }
    }

    /**
     * Show error message
     */
    private showError(message: string): void {
        const container = this.findElement('#options-list');
        if (container) {
            container.innerHTML = `
                <div class="text-center py-4">
                    <div class="text-red-500 dark:text-red-400">${message}</div>
                </div>
            `;
        }
    }

    protected destroyComponent(): void {
        this.deactivate();
    }
}
