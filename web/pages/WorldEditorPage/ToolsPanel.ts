import { BaseComponent, EventBus, LCMComponent } from '@panyam/tsappkit';
import { EditorEventTypes } from '../common/events';
import { BRUSH_SIZE_NAMES } from '../common/ColorsAndNames'
import { IWorldEditorPresenter } from './WorldEditorPresenter';

/**
 * EditorToolsPanel Component - State generator and DOM owner for editor tools with tabbed interface
 *
 * This component demonstrates the new lifecycle architecture with explicit dependency injection:
 * 1. performLocalInit() - Set up UI controls and event handlers without dependencies
 * 2. injectDependencies() - Receive presenter when available
 * 3. activate() - Enable full functionality once all dependencies are ready
 *
 * Responsibilities:
 * - Own and manage terrain/unit button DOM elements and styling
 * - Generate state changes when users interact with controls
 * - Handle brush size dropdown and player selection dropdowns
 * - Maintain visual selection state (button highlights, etc.)
 * - Update state via direct presenter method calls
 * - Manage tab switching between Nature/City/Unit sections
 *
 * Architecture:
 * - Receives presenter instance via explicit setter
 * - Updates state directly when user interacts with controls
 * - Manages its own DOM elements without external interference
 * - Does NOT observe state changes (it's the generator, not observer)
 * - Provides tab switching API for keyboard shortcuts
 */
export class EditorToolsPanel extends BaseComponent {
    // Dependencies (injected via explicit setters)
    private presenter: IWorldEditorPresenter | null = null;

    // Internal state tracking
    private isUIBound = false;
    private isActivated = false;
    private pendingOperations: Array<() => void> = [];
    
    // Tab state management
    private activeTab: 'tiles' | 'unit' = 'tiles';
    
    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('editor-tools-panel', rootElement, eventBus, debugMode);
    }
    
    // LCMComponent Phase 1: Initialize DOM and discover children (no dependencies needed)
    public performLocalInit(): LCMComponent[] {
        if (this.isUIBound) {
            this.log('Already bound to DOM, skipping');
            return [];
        }
        
        this.log('Binding EditorToolsPanel to DOM');
        
        // Set up UI elements and event handlers (independent of dependencies)
        this.bindTabButtons();
        this.bindTerrainButtons();
        this.bindCrossingButtons();
        this.bindCrossingDirectionPreset();
        this.bindUnitButtons();
        this.bindBrushSizeControl();
        this.bindPlayerControl();
        this.bindCityPlayerControl();
        
        this.isUIBound = true;
        this.log('EditorToolsPanel bound to DOM successfully');
        
        // This is a leaf component - no children
        return [];
    }
    
    // Phase 2: Inject dependencies - simplified to use explicit setters
    public injectDependencies(): void {
        this.log('EditorToolsPanel: Dependencies injection phase - using explicit setters');

        // Dependencies should be set directly by parent using setters
        // This phase just validates that required dependencies are available
        if (!this.presenter) {
            throw new Error('EditorToolsPanel requires presenter - use setPresenter()');
        }

        this.log('Dependencies validation complete');
    }

    // Explicit dependency setters
    public setPresenter(presenter: IWorldEditorPresenter): void {
        this.presenter = presenter;
        this.log('Presenter set via explicit setter');

        // If already activated, sync UI immediately
        if (this.isActivated) {
            this.syncUIWithPresenter();
        }
    }

    // Explicit dependency getters
    public getPresenter(): IWorldEditorPresenter | null {
        return this.presenter;
    }
    
    // Phase 3: Activate component
    public activate(): void {
        if (this.isActivated) {
            this.log('Already activated, skipping');
            return;
        }

        this.log('Activating EditorToolsPanel');

        // Process any operations that were queued during UI binding
        this.processPendingOperations();

        // Sync UI to match current state
        this.syncUIWithPresenter();

        // Ensure default tab is shown
        this.switchToTab(this.activeTab);

        this.isActivated = true;
        this.log('EditorToolsPanel activated successfully');
    }

    // Phase 4: Deactivate component
    public deactivate(): void {
        this.log('Deactivating EditorToolsPanel');

        // Clear any pending operations
        this.pendingOperations = [];

        // Reset state
        this.isActivated = false;
        this.presenter = null;
        this.activeTab = 'tiles';

        // Clean up overlays
        this.hideNumberOverlays();

        this.log('EditorToolsPanel deactivated');
    }
    
    // Deferred Execution System
    
    /**
     * Execute operation when component is ready, or queue it for later
     */
    private executeWhenReady(operation: () => void): void {
        if (this.isActivated && this.presenter) {
            // Component is ready - execute immediately
            operation();
        } else {
            // Component not ready - queue for later
            this.pendingOperations.push(operation);
            this.log('Component not ready - operation queued');
        }
    }
    
    /**
     * Process all pending operations when component becomes ready
     */
    private processPendingOperations(): void {
        if (this.pendingOperations.length > 0) {
            this.log(`Processing ${this.pendingOperations.length} pending operations`);
            
            const operations = [...this.pendingOperations];
            this.pendingOperations = [];
            
            operations.forEach(operation => {
                operation();
            });
        }
    }
    
    protected destroyComponent(): void {
        this.deactivate();
    }
    
    /**
     * Bind tab button click events
     */
    private bindTabButtons(): void {
        const tabButtons = this.findElements('.tab-button');

        tabButtons.forEach(button => {
            button.addEventListener('click', (e) => {
                const clickedButton = e.currentTarget as HTMLElement;
                const tabName = clickedButton.getAttribute('data-tab') as 'tiles' | 'unit';

                if (tabName) {
                    this.switchToTab(tabName);
                }
            });
        });

        this.log(`Bound ${tabButtons.length} tab buttons`);
    }

    /**
     * Switch to a specific tab
     */
    public switchToTab(tabName: 'tiles' | 'unit'): void {
        this.activeTab = tabName;

        // Update tab button active states
        const tabButtons = this.findElements('.tab-button');
        tabButtons.forEach(button => {
            const buttonTab = button.getAttribute('data-tab');
            if (buttonTab === tabName) {
                button.classList.add('active-tab');
            } else {
                button.classList.remove('active-tab');
            }
        });

        // Show/hide tab content
        const tabContents = this.findElements('.tab-content');
        tabContents.forEach(content => {
            const contentId = content.id;
            if (contentId === `${tabName}-tab-content`) {
                content.classList.remove('hidden');
            } else {
                content.classList.add('hidden');
            }
        });

        this.log(`Switched to ${tabName} tab`);
    }

    /**
     * Get current active tab
     */
    public getActiveTab(): 'tiles' | 'unit' {
        return this.activeTab;
    }
    
    /**
     * Select item by index within the active tab
     */
    public selectByIndex(index: number): void {
        this.executeWhenReady(() => {
            switch (this.activeTab) {
                case 'tiles':
                    this.selectTilesByIndex(index);
                    break;
                case 'unit':
                    this.selectUnitByIndex(index);
                    break;
            }
        });
    }

    /**
     * Show number overlays for the active tab
     */
    public showNumberOverlays(): void {
        let selector = '';

        switch (this.activeTab) {
            case 'tiles':
                selector = '[data-tiles-index]';
                break;
            case 'unit':
                selector = '[data-unit-index]';
                break;
        }

        const buttons = this.findElements(selector);
        buttons.forEach(button => {
            const indexAttr = this.activeTab === 'tiles' ? 'data-tiles-index' : 'data-unit-index';
            const index = button.getAttribute(indexAttr);
            if (index && parseInt(index) > 0) {
                this.addNumberOverlay(button as HTMLElement, index);
            }
        });

        this.log(`Showing number overlays for ${this.activeTab} tab`);
    }
    
    /**
     * Hide all number overlays
     */
    public hideNumberOverlays(): void {
        const overlays = this.findElements('.shortcut-number-overlay');
        overlays.forEach(overlay => overlay.remove());
    }
    
    /**
     * Add number overlay to a specific element
     */
    private addNumberOverlay(element: HTMLElement, number: string): void {
        // Remove existing overlay if present
        const existingOverlay = element.querySelector('.shortcut-number-overlay');
        if (existingOverlay) {
            existingOverlay.remove();
        }
        
        // Create new overlay
        const overlay = document.createElement('div');
        overlay.className = 'shortcut-number-overlay absolute top-0 right-0 bg-blue-500 text-white text-xs font-bold rounded-full w-5 h-5 flex items-center justify-center z-10 -mt-1 -mr-1';
        overlay.textContent = number;
        
        // Position relative parent
        element.style.position = 'relative';
        element.appendChild(overlay);
    }
    
    /**
     * Select tile/crossing by index in the tiles tab
     */
    private selectTilesByIndex(index: number): void {
        // Try terrain button first
        let button = this.findElement(`[data-tiles-index="${index}"]`);
        if (button) {
            (button as HTMLElement).click();
            return;
        }
        // Try crossing button
        button = this.findElement(`[data-crossing-index="${index}"]`);
        if (button) {
            (button as HTMLElement).click();
        }
    }

    /**
     * Select unit by index
     */
    private selectUnitByIndex(index: number): void {
        const button = this.findElement(`[data-unit-index="${index}"]`);
        if (button) {
            (button as HTMLElement).click();
        }
    }
    
    /**
     * Bind terrain button click events
     */
    private bindTerrainButtons(): void {
        const terrainButtons = this.findElements('.terrain-button');
        
        terrainButtons.forEach(button => {
            button.addEventListener('click', (e) => {
                const clickedButton = e.currentTarget as HTMLElement;
                const terrain = clickedButton.getAttribute('data-terrain');
                
                if (terrain) {
                    const terrainValue = parseInt(terrain);
                    // Get terrain name from button's title or text content
                    const terrainButton = clickedButton.querySelector('.text-xs.truncate');
                    const terrainName = terrainButton?.textContent || clickedButton.title.split('(')[0].trim() || `Terrain ${terrainValue}`;

                    if (terrainValue === 0) {
                        // Clear mode
                        this.executeWhenReady(() => {
                            this.updateButtonSelection(clickedButton);
                            this.presenter!.setPlacementMode('clear');
                            this.log('Selected clear mode');
                        });
                    } else {
                        // Terrain selection
                        this.executeWhenReady(() => {
                            this.updateButtonSelection(clickedButton);
                            this.presenter!.selectTerrain(terrainValue);
                            this.log(`Selected terrain: ${terrainValue} (${terrainName})`);
                        });
                    }
                }
            });
        });
        
        this.log(`Bound ${terrainButtons.length} terrain buttons`);
    }

    /**
     * Bind crossing button click events (Road/Bridge)
     */
    private bindCrossingButtons(): void {
        const crossingButtons = this.findElements('.crossing-button');

        crossingButtons.forEach(button => {
            button.addEventListener('click', (e) => {
                const clickedButton = e.currentTarget as HTMLElement;
                const crossing = clickedButton.getAttribute('data-crossing');

                if (crossing) {
                    // Get crossing name from button's text content
                    const crossingNameEl = clickedButton.querySelector('.text-xs.truncate');
                    const crossingName = crossingNameEl?.textContent || crossing;

                    this.executeWhenReady(() => {
                        this.updateButtonSelection(clickedButton);
                        this.presenter!.selectCrossing(crossing as 'road' | 'bridge');
                        this.log(`Selected crossing: ${crossing} (${crossingName})`);
                    });
                }
            });
        });

        this.log(`Bound ${crossingButtons.length} crossing buttons`);
    }

    // Crossing direction presets: [LEFT, TOP_LEFT, TOP_RIGHT, RIGHT, BOTTOM_RIGHT, BOTTOM_LEFT]
    private static readonly CROSSING_PRESETS: { [key: string]: boolean[] } = {
        'horizontal': [true, false, false, true, false, false],        // L + R
        'diagonal-tlbr': [false, true, false, false, true, false],     // TL + BR
        'diagonal-trbl': [false, false, true, false, false, true],     // TR + BL
        't-left': [true, true, false, false, true, false],             // L + TL + BR
        't-right': [false, false, true, true, false, true],            // TR + R + BL
        'cross': [true, false, false, true, true, true],               // L + R + BR + BL (4-way)
        'custom': [false, false, false, false, false, false],          // User-defined
    };

    /** Current crossing direction state */
    private crossingDirections: boolean[] = [true, false, false, true, false, false]; // Default: horizontal

    /**
     * Bind crossing direction preset controls
     */
    private bindCrossingDirectionPreset(): void {
        const presetSelect = this.findElement('#crossing-preset-select') as HTMLSelectElement;

        if (!presetSelect) {
            this.log('Crossing preset select not found');
            return;
        }

        // Bind preset dropdown change
        presetSelect.addEventListener('change', () => {
            const preset = presetSelect.value;
            if (preset !== 'custom' && EditorToolsPanel.CROSSING_PRESETS[preset]) {
                this.crossingDirections = [...EditorToolsPanel.CROSSING_PRESETS[preset]];
                this.updateCrossingDirectionVisuals();
                this.executeWhenReady(() => {
                    this.presenter!.setCrossingConnectsTo(this.crossingDirections);
                    this.log(`Set crossing preset: ${preset}`);
                });
            }
        });

        // Bind individual direction line clicks
        for (let i = 0; i < 6; i++) {
            const line = document.getElementById(`crossing-dir-${i}`);
            if (line) {
                line.addEventListener('click', () => {
                    this.crossingDirections[i] = !this.crossingDirections[i];
                    this.updateCrossingDirectionVisuals();

                    // Switch to "custom" in dropdown
                    presetSelect.value = 'custom';

                    this.executeWhenReady(() => {
                        this.presenter!.setCrossingConnectsTo(this.crossingDirections);
                        this.log(`Toggled crossing direction ${i}: ${this.crossingDirections[i]}`);
                    });
                });
            }
        }

        // Initialize visual state
        this.updateCrossingDirectionVisuals();
        this.log('Bound crossing direction preset controls');
    }

    /**
     * Update the visual state of crossing direction lines
     */
    private updateCrossingDirectionVisuals(): void {
        for (let i = 0; i < 6; i++) {
            const line = document.getElementById(`crossing-dir-${i}`);
            if (line) {
                if (this.crossingDirections[i]) {
                    // Active: show in brown/road color
                    line.setAttribute('stroke', '#d97706'); // amber-600
                } else {
                    // Inactive: dim gray
                    line.setAttribute('stroke', '#d1d5db'); // gray-300
                }
            }
        }
    }

    /**
     * Bind unit button click events
     */
    private bindUnitButtons(): void {
        const unitButtons = this.findElements('.unit-button');
        
        unitButtons.forEach(button => {
            button.addEventListener('click', (e) => {
                const clickedButton = e.currentTarget as HTMLElement;
                const unit = clickedButton.getAttribute('data-unit');
                
                if (unit) {
                    const unitValue = parseInt(unit);
                    // Get unit name from button's text content
                    const unitButton = clickedButton.querySelector('.text-xs.truncate');
                    const unitName = unitButton?.textContent || `Unit ${unitValue}`;

                    this.executeWhenReady(() => {
                        this.updateButtonSelection(clickedButton);

                        // Get current player selection
                        const playerSelect = this.findElement('#unit-player-color') as HTMLSelectElement;
                        const currentPlayer = playerSelect ? parseInt(playerSelect.value) : 1;

                        this.presenter!.selectUnit(unitValue);
                        this.presenter!.selectPlayer(currentPlayer);

                        this.log(`Selected unit: ${unitValue} (${unitName}) for player ${currentPlayer}`);
                    });
                }
            });
        });
        
        this.log(`Bound ${unitButtons.length} unit buttons`);
    }
    
    /**
     * Bind brush size control events
     */
    private bindBrushSizeControl(): void {
        const brushSizeSelect = this.findElement('#brush-size') as HTMLSelectElement;

        if (brushSizeSelect) {
            brushSizeSelect.addEventListener('change', (e) => {
                const target = e.target as HTMLSelectElement;
                this.executeWhenReady(() => {
                    const value = target.value;

                    let mode: string;
                    let size: number;
                    let displayName: string;

                    if (value.startsWith('fill:')) {
                        // Fill mode - extract radius
                        mode = "fill";
                        size = parseInt(value.substring(5));
                        const option = target.selectedOptions[0];
                        displayName = option?.text || `Fill ${size}`;
                    } else {
                        // Brush mode - extract size
                        mode = "brush";
                        size = parseInt(value);
                        const option = target.selectedOptions[0];
                        displayName = option?.text || `Brush ${size}`;
                    }

                    this.presenter!.setBrushSize(mode, size);
                    this.log(`Brush/Fill tool changed to: ${mode} with size ${size} (${displayName})`);
                });
            });

            this.log('Bound brush size control');
        } else {
            this.log('Brush size control not found');
        }
    }
    
    /**
     * Bind unit player selection control events
     */
    private bindPlayerControl(): void {
        const playerSelect = this.findElement('#unit-player-color') as HTMLSelectElement;

        if (playerSelect) {
            playerSelect.addEventListener('change', (e) => {
                this.executeWhenReady(() => {
                    const target = e.target as HTMLSelectElement;
                    const playerId = parseInt(target.value);
                    this.presenter!.selectPlayer(playerId);
                    this.log(`Unit player changed to: ${playerId}`);
                });
            });

            this.log('Bound unit player control');
        } else {
            this.log('Unit player control not found');
        }
    }

    /**
     * Bind city player selection control events
     */
    private bindCityPlayerControl(): void {
        const cityPlayerSelect = this.findElement('#player-color') as HTMLSelectElement;

        if (cityPlayerSelect) {
            cityPlayerSelect.addEventListener('change', (e) => {
                this.executeWhenReady(() => {
                    const target = e.target as HTMLSelectElement;
                    const playerId = parseInt(target.value);
                    this.presenter!.selectPlayer(playerId);
                    this.log(`City player changed to: ${playerId}`);
                });
            });

            this.log('Bound city player control');
        } else {
            this.log('City player control not found');
        }
    }
    
    /**
     * Sync UI controls with presenter state
     */
    private syncUIWithPresenter(): void {
        if (!this.presenter) return;

        const toolState = this.presenter.getToolState();

        // Update terrain button selection
        this.updateTerrainButtonHighlight(toolState.selectedTerrain);

        // Update unit button selection
        this.updateUnitButtonHighlight(toolState.selectedUnit);

        // Update brush size dropdown
        this.updateBrushSizeDropdown(toolState.brushMode, toolState.brushSize);

        // Update player dropdowns
        this.updatePlayerDropdowns(toolState.selectedPlayer);

        this.log('UI synced with presenter');
    }
    
    /**
     * Update visual selection state for terrain/unit/crossing buttons
     */
    private updateButtonSelection(selectedButton: HTMLElement): void {
        // Remove selection from all terrain, unit, and crossing buttons within this component
        const allButtons = this.findElements('.terrain-button, .unit-button, .crossing-button');
        allButtons.forEach(btn => {
            btn.classList.remove('bg-blue-100', 'dark:bg-blue-900', 'border-blue-500');
        });

        // Add selection to clicked button
        selectedButton.classList.add('bg-blue-100', 'dark:bg-blue-900', 'border-blue-500');
    }
    
    /**
     * Update terrain button highlight (internal method for UI sync)
     */
    private updateTerrainButtonHighlight(terrainType: number): void {
        const terrainButton = this.findElement(`[data-terrain="${terrainType}"]`);
        if (terrainButton) {
            this.updateButtonSelection(terrainButton as HTMLElement);
        }
    }
    
    /**
     * Update unit button highlight (internal method for UI sync) 
     */
    private updateUnitButtonHighlight(unitType: number): void {
        const unitButton = this.findElement(`[data-unit="${unitType}"]`);
        if (unitButton) {
            this.updateButtonSelection(unitButton as HTMLElement);
        }
    }
    
    /**
     * Update brush size dropdown (internal method for UI sync)
     */
    private updateBrushSizeDropdown(brushMode: string, brushSize: number): void {
        const brushSizeSelect = this.findElement('#brush-size') as HTMLSelectElement;
        if (brushSizeSelect) {
            // Reconstruct the value based on mode
            const value = brushMode === "fill" ? `fill:${brushSize}` : brushSize.toString();
            brushSizeSelect.value = value;
        }
    }
    
    /**
     * Update player dropdowns (internal method for UI sync)
     */
    private updatePlayerDropdowns(playerId: number): void {
        // Update unit player dropdown
        const unitPlayerSelect = this.findElement('#unit-player-color') as HTMLSelectElement;
        if (unitPlayerSelect) {
            unitPlayerSelect.value = playerId.toString();
        }
        
        // Update city player dropdown  
        const cityPlayerSelect = this.findElement('#player-color') as HTMLSelectElement;
        if (cityPlayerSelect) {
            cityPlayerSelect.value = playerId.toString();
        }
    }
    
    /**
     * Get current tool state for external queries
     */
    public getCurrentState() {
        if (!this.presenter) {
            // Default state if presenter not set
            return {
                terrain: 1,
                unit: 0,
                brushMode: "brush",
                brushSize: 0,
                playerId: 1,
                placementMode: 'terrain' as const
            };
        }
        const toolState = this.presenter.getToolState();
        return {
            terrain: toolState.selectedTerrain,
            unit: toolState.selectedUnit,
            brushMode: toolState.brushMode,
            brushSize: toolState.brushSize,
            playerId: toolState.selectedPlayer,
            placementMode: toolState.placementMode
        };
    }
}
