import { BaseComponent } from './Component';
import { EventBus, EditorEventTypes } from './EventBus';
import { WorldEditorPageState } from './WorldEditorPageState';
import { ComponentLifecycle } from './ComponentLifecycle';
import { BRUSH_SIZE_NAMES, TERRAIN_NAMES, UNIT_NAMES } from './ColorsAndNames'

/**
 * EditorToolsPanel Component - State generator and DOM owner for editor tools with tabbed interface
 * 
 * This component demonstrates the new lifecycle architecture with explicit dependency injection:
 * 1. initializeDOM() - Set up UI controls and event handlers without dependencies
 * 2. injectDependencies() - Receive WorldEditorPageState when available
 * 3. activate() - Enable full functionality once all dependencies are ready
 * 
 * Responsibilities:
 * - Own and manage terrain/unit button DOM elements and styling
 * - Generate page state changes when users interact with controls
 * - Handle brush size dropdown and player selection dropdowns
 * - Maintain visual selection state (button highlights, etc.)
 * - Update page state via direct method calls (not events)
 * - Manage tab switching between Nature/City/Unit sections
 * 
 * Architecture:
 * - Receives WorldEditorPageState instance via explicit setter (not dependency injection)
 * - Updates state directly when user interacts with controls
 * - Manages its own DOM elements without external interference
 * - Does NOT observe state changes (it's the generator, not observer)
 * - Provides tab switching API for keyboard shortcuts
 */
export class EditorToolsPanel extends BaseComponent {
    // Dependencies (injected via explicit setters)
    private pageState: WorldEditorPageState | null = null;
    
    // Internal state tracking
    private isUIBound = false;
    private isActivated = false;
    private pendingOperations: Array<() => void> = [];
    
    // Tab state management
    private activeTab: 'nature' | 'city' | 'unit' = 'nature';
    
    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('editor-tools-panel', rootElement, eventBus, debugMode);
    }
    
    // ComponentLifecycle Phase 1: Initialize DOM and discover children (no dependencies needed)
    public initializeDOM(): ComponentLifecycle[] {
        if (this.isUIBound) {
            this.log('Already bound to DOM, skipping');
            return [];
        }
        
        try {
            this.log('Binding EditorToolsPanel to DOM');
            
            // Set up UI elements and event handlers (independent of dependencies)
            this.bindTabButtons();
            this.bindTerrainButtons();
            this.bindUnitButtons();
            this.bindBrushSizeControl();
            this.bindPlayerControl();
            this.bindCityPlayerControl();
            
            this.isUIBound = true;
            this.log('EditorToolsPanel bound to DOM successfully');
            
            // This is a leaf component - no children
            return [];
            
        } catch (error) {
            this.handleError('Failed to bind EditorToolsPanel to DOM', error);
            throw error;
        }
    }
    
    // Phase 2: Inject dependencies - simplified to use explicit setters
    public injectDependencies(deps: Record<string, any>): void {
        this.log('EditorToolsPanel: Dependencies injection phase - using explicit setters');
        
        // Dependencies should be set directly by parent using setters
        // This phase just validates that required dependencies are available
        if (!this.pageState) {
            throw new Error('EditorToolsPanel requires page state - use setPageState()');
        }
        
        this.log('Dependencies validation complete');
    }
    
    // Explicit dependency setters
    public setPageState(pageState: WorldEditorPageState): void {
        this.pageState = pageState;
        this.log('Page state set via explicit setter');
        
        // If already activated, sync UI immediately
        if (this.isActivated) {
            this.syncUIWithPageState();
        }
    }
    
    // Explicit dependency getters
    public getPageState(): WorldEditorPageState | null {
        return this.pageState;
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
        
        // Sync UI to match current page state
        this.syncUIWithPageState();
        
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
        this.pageState = null;
        this.activeTab = 'nature';
        
        // Clean up overlays
        this.hideNumberOverlays();
        
        this.log('EditorToolsPanel deactivated');
    }
    
    // Deferred Execution System
    
    /**
     * Execute operation when component is ready, or queue it for later
     */
    private executeWhenReady(operation: () => void): void {
        if (this.isActivated && this.pageState) {
            // Component is ready - execute immediately
            try {
                operation();
            } catch (error) {
                this.handleError('Operation failed', error);
            }
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
                try {
                    operation();
                } catch (error) {
                    this.handleError('Pending operation failed', error);
                }
            });
        }
    }
    
    protected initializeComponent(): void {
        // This is handled by the new lifecycle system
        // Keep empty for backward compatibility
    }
    
    protected bindToDOM(): void {
        // This is handled by the new lifecycle system
        // Keep empty for backward compatibility
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
                const tabName = clickedButton.getAttribute('data-tab') as 'nature' | 'city' | 'unit';
                
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
    public switchToTab(tabName: 'nature' | 'city' | 'unit'): void {
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
    public getActiveTab(): 'nature' | 'city' | 'unit' {
        return this.activeTab;
    }
    
    /**
     * Select item by index within the active tab
     */
    public selectByIndex(index: number): void {
        this.executeWhenReady(() => {
            switch (this.activeTab) {
                case 'nature':
                    this.selectNatureTerrainByIndex(index);
                    break;
                case 'city':
                    this.selectCityTerrainByIndex(index);
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
            case 'nature':
                selector = '[data-nature-index]';
                break;
            case 'city':
                selector = '[data-city-index]';
                break;
            case 'unit':
                selector = '[data-unit-index]';
                break;
        }
        
        const buttons = this.findElements(selector);
        buttons.forEach(button => {
            const index = button.getAttribute(`data-${this.activeTab}-index`);
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
     * Select nature terrain by index
     */
    private selectNatureTerrainByIndex(index: number): void {
        const button = this.findElement(`[data-nature-index="${index}"]`);
        if (button) {
            (button as HTMLElement).click();
        }
    }
    
    /**
     * Select city terrain by index
     */
    private selectCityTerrainByIndex(index: number): void {
        const button = this.findElement(`[data-city-index="${index}"]`);
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
                    const terrainName = TERRAIN_NAMES[terrainValue].name
                    
                    if (terrainValue === 0) {
                        // Clear mode
                        this.executeWhenReady(() => {
                            this.updateButtonSelection(clickedButton);
                            
                            // Update page state directly
                            if (this.pageState) {
                                this.pageState.setPlacementMode('clear');
                            }
                            
                            this.log('Selected clear mode');
                        });
                    } else {
                        // Terrain selection
                        this.executeWhenReady(() => {
                            this.updateButtonSelection(clickedButton);
                            
                            // Update page state directly
                            if (this.pageState) {
                                this.pageState.setSelectedTerrain(terrainValue);
                            }
                            
                            this.log(`Selected terrain: ${terrainValue} (${terrainName})`);
                        });
                    }
                }
            });
        });
        
        this.log(`Bound ${terrainButtons.length} terrain buttons`);
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
                    const unitName = UNIT_NAMES[unitValue].name;
                    
                    this.executeWhenReady(() => {
                        this.updateButtonSelection(clickedButton);
                        
                        // Get current player selection
                        const playerSelect = this.findElement('#unit-player-color') as HTMLSelectElement;
                        const currentPlayer = playerSelect ? parseInt(playerSelect.value) : 1;
                        
                        // Update page state directly
                        if (this.pageState) {
                            this.pageState.setSelectedUnit(unitValue);
                            this.pageState.setSelectedPlayer(currentPlayer);
                        }
                        
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
                    const brushSize = parseInt(target.value);
                    const sizeName = BRUSH_SIZE_NAMES[brushSize] || `Size ${brushSize}`;
                    
                    // Update page state directly
                    if (this.pageState) {
                        this.pageState.setBrushSize(brushSize);
                    }
                    
                    this.log(`Brush size changed to: ${sizeName}`);
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
                    
                    // Update page state directly
                    if (this.pageState) {
                        this.pageState.setSelectedPlayer(playerId);
                    }
                    
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
                    
                    // Update page state directly
                    if (this.pageState) {
                        this.pageState.setSelectedPlayer(playerId);
                    }
                    
                    this.log(`City player changed to: ${playerId}`);
                });
            });
            
            this.log('Bound city player control');
        } else {
            this.log('City player control not found');
        }
    }
    
    /**
     * Sync UI controls with current page state
     */
    private syncUIWithPageState(): void {
        if (!this.pageState) return;
        
        const toolState = this.pageState.getToolState();
        
        // Update terrain button selection
        this.updateTerrainButtonHighlight(toolState.selectedTerrain);
        
        // Update unit button selection 
        this.updateUnitButtonHighlight(toolState.selectedUnit);
        
        // Update brush size dropdown
        this.updateBrushSizeDropdown(toolState.brushSize);
        
        // Update player dropdowns
        this.updatePlayerDropdowns(toolState.selectedPlayer);
        
        this.log('UI synced with page state');
    }
    
    /**
     * Update visual selection state for terrain/unit buttons
     */
    private updateButtonSelection(selectedButton: HTMLElement): void {
        // Remove selection from all terrain and unit buttons within this component
        const allButtons = this.findElements('.terrain-button, .unit-button');
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
    private updateBrushSizeDropdown(brushSize: number): void {
        const brushSizeSelect = this.findElement('#brush-size') as HTMLSelectElement;
        if (brushSizeSelect) {
            brushSizeSelect.value = brushSize.toString();
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
        if (this.pageState) {
            const toolState = this.pageState.getToolState();
            return {
                terrain: toolState.selectedTerrain,
                unit: toolState.selectedUnit,
                brushSize: toolState.brushSize,
                playerId: toolState.selectedPlayer,
                placementMode: toolState.placementMode
            };
        }
        return {
            terrain: 1,
            unit: 0,
            brushSize: 0,
            playerId: 1,
            placementMode: 'terrain' as const
        };
    }
}
