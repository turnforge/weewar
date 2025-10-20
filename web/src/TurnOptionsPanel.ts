import { BaseComponent } from '../lib/Component';
import { EventBus } from '../lib/EventBus';
import { LCMComponent } from '../lib/LCMComponent';
import { GameState } from './GameState';
import { World } from './World';
import { GameViewPresenterServiceClient as  GameViewPresenterClient } from '../gen/wasmjs/weewar/v1/gameViewPresenterClient';
import { ITheme } from '../assets/themes/BaseTheme';
import { ThemeUtils } from './ThemeUtils';
import {
    Unit,
    GameOption,
    MoveOption,
    AttackOption,
    EndTurnOption,
    BuildUnitOption,
    CaptureBuildingOption,
    GetOptionsAtResponse
} from '../gen/wasmjs/weewar/v1/interfaces'

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
    public gameViewPresenterClient: GameViewPresenterClient;
    private world: World | null = null;
    private theme: ITheme | null = null;
    private currentOptions: GameOption[] = [];
    private selectedPosition: { q: number; r: number } | null = null;
    private selectedUnit: Unit | null = null;

    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('turn-options-panel', rootElement, eventBus, debugMode);
    }
    
    /**
     * Set the World dependency
     */
    public setWorld(world: World): void {
        this.world = world;
        this.log('World dependency set');
    }

    /**
     * Set the theme for getting unit names and images
     */
    public setTheme(theme: ITheme): void {
        this.theme = theme;
    }

    /**
     * Hydrate theme images after Go template renders HTML
     * Call this after the HTML content is injected by the Go backend
     */
    public async hydrateThemeImages(): Promise<void> {
        await ThemeUtils.hydrateThemeImages(this.rootElement, this.theme, this.debugMode);
        this.setupOptionClickHandlers();
    }

    /**
     * Setup click handlers for option buttons
     */
    private setupOptionClickHandlers(): void {
        const buttons = this.rootElement.querySelectorAll('.turn-option-button');
        buttons.forEach(button => {
            button.addEventListener('click', (e) => {
                const target = e.currentTarget as HTMLElement;
                const optionIndex = parseInt(target.getAttribute('data-option-index') || '-1');
                const optionType = target.getAttribute('data-option-type');
                const q = parseInt(target.getAttribute('data-q') || '0');
                const r = parseInt(target.getAttribute('data-r') || '0');

                this.log(`Option clicked: type=${optionType}, index=${optionIndex}, position=(${q},${r})`);

                // Call presenter directly
                this.gameViewPresenterClient.turnOptionClicked({
                    gameId: "",
                    optionIndex: optionIndex,
                    optionType: optionType || "",
                    q: q,
                    r: r,
                });
            });
        });
    }

    // Phase 3: Activate component
    public activate(): void {
        // Show initial empty state
        this.showEmptyState();
    }

    // Phase 4: Deactivate component
    public deactivate(): void {
        this.log('Deactivating TurnOptionsPanel');
        this.currentOptions = [];
        this.selectedPosition = null;
        this.selectedUnit = null;
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
     * Process options from the server response
     */
    private processOptions(response: GetOptionsAtResponse): void {
        // Options are already sorted by the server
        this.currentOptions = response.options || [];
        
        // Display the processed options
        // this.displayOptions();
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

    // protected destroyComponent(): void { this.deactivate(); }
}
