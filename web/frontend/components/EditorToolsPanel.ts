import { BaseComponent } from './Component';
import { EventBus, EditorEventTypes } from './EventBus';
import { MapEditorPageState } from './MapEditorPageState';

const BRUSH_SIZE_NAMES = ['Single (1 hex)', 'Small (3 hexes)', 'Medium (5 hexes)', 'Large (9 hexes)', 'X-Large (15 hexes)', 'XX-Large (25 hexes)'];

/**
 * EditorToolsPanel Component - State generator and DOM owner for editor tools
 * 
 * Responsibilities:
 * - Own and manage terrain/unit button DOM elements and styling
 * - Generate page state changes when users interact with controls
 * - Handle brush size dropdown and player selection dropdowns
 * - Maintain visual selection state (button highlights, etc.)
 * - Update page state via direct method calls (not events)
 * 
 * Architecture:
 * - Receives MapEditorPageState instance from parent
 * - Updates state directly when user interacts with controls
 * - Manages its own DOM elements without external interference
 * - Does NOT observe state changes (it's the generator, not observer)
 */
export class EditorToolsPanel extends BaseComponent {
    private pageState: MapEditorPageState | null = null;
    
    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('editor-tools-panel', rootElement, eventBus, debugMode);
    }
    
    // Method to inject page state from parent
    public setPageState(pageState: MapEditorPageState): void {
        this.pageState = pageState;
        // Initialize UI to match current state
        this.syncUIWithPageState();
    }
    
    protected initializeComponent(): void {
        this.log('Initializing EditorToolsPanel component as state generator');
        // No event subscriptions - this component generates state, doesn't observe it
        this.log('EditorToolsPanel component initialized');
    }
    
    protected bindToDOM(): void {
        try {
            this.log('Binding EditorToolsPanel to DOM');
            
            // Bind terrain button events
            this.bindTerrainButtons();
            
            // Bind unit button events  
            this.bindUnitButtons();
            
            // Bind brush size selection
            this.bindBrushSizeControl();
            
            // Bind player selection
            this.bindPlayerControl();
            
            this.log('EditorToolsPanel bound to DOM');
            
        } catch (error) {
            this.handleError('Failed to bind EditorToolsPanel to DOM', error);
        }
    }
    
    protected destroyComponent(): void {
        this.log('Destroying EditorToolsPanel component');
        // Event listeners will be automatically removed when DOM elements are destroyed
        // EventBus subscriptions are automatically cleaned up by BaseComponent
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
                    const terrainName = this.getTerrainName(terrainValue);
                    
                    if (terrainValue === 0) {
                        // Clear mode
                        this.updateButtonSelection(clickedButton);
                        
                        // Update page state directly
                        if (this.pageState) {
                            this.pageState.setPlacementMode('clear');
                        }
                        
                        this.log('Selected clear mode');
                    } else {
                        // Terrain selection
                        this.updateButtonSelection(clickedButton);
                        
                        // Update page state directly
                        if (this.pageState) {
                            this.pageState.setSelectedTerrain(terrainValue);
                        }
                        
                        this.log(`Selected terrain: ${terrainValue} (${terrainName})`);
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
                    const unitName = this.getUnitName(unitValue);
                    
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
                const brushSize = parseInt(target.value);
                const sizeName = BRUSH_SIZE_NAMES[brushSize] || `Size ${brushSize}`;
                
                // Update page state directly
                if (this.pageState) {
                    this.pageState.setBrushSize(brushSize);
                }
                
                this.log(`Brush size changed to: ${sizeName}`);
            });
            
            this.log('Bound brush size control');
        } else {
            this.log('Brush size control not found');
        }
    }
    
    /**
     * Bind player selection control events
     */
    private bindPlayerControl(): void {
        const playerSelect = this.findElement('#unit-player-color') as HTMLSelectElement;
        
        if (playerSelect) {
            playerSelect.addEventListener('change', (e) => {
                const target = e.target as HTMLSelectElement;
                const playerId = parseInt(target.value);
                
                // Update page state directly
                if (this.pageState) {
                    this.pageState.setSelectedPlayer(playerId);
                }
                
                this.log(`Player changed to: ${playerId}`);
            });
            
            this.log('Bound player control');
        } else {
            this.log('Player control not found');
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
        
        // Update player dropdown
        this.updatePlayerDropdown(toolState.selectedPlayer);
        
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
     * Update player dropdown (internal method for UI sync)
     */
    private updatePlayerDropdown(playerId: number): void {
        const playerSelect = this.findElement('#unit-player-color') as HTMLSelectElement;
        if (playerSelect) {
            playerSelect.value = playerId.toString();
        }
    }
    
    
    /**
     * Get human-readable terrain name
     */
    private getTerrainName(terrainType: number): string {
        const terrainNames: { [key: number]: string } = {
            0: 'Clear',
            1: 'Grass',
            2: 'Desert', 
            3: 'Water',
            4: 'Mountain',
            5: 'Rock',
            // Add more terrain types as needed
        };
        
        return terrainNames[terrainType] || `Terrain ${terrainType}`;
    }
    
    /**
     * Get human-readable unit name
     */
    private getUnitName(unitType: number): string {
        const unitNames: { [key: number]: string } = {
            1: 'Infantry',
            2: 'Mech',
            3: 'Recon',
            4: 'Tank',
            5: 'Medium Tank',
            6: 'Neo Tank',
            7: 'APC',
            8: 'Artillery',
            9: 'Rocket',
            10: 'Anti-Air',
            11: 'Missile',
            12: 'Fighter',
            13: 'Bomber',
            14: 'B-Copter',
            15: 'T-Copter',
            16: 'Battleship',
            17: 'Cruiser',
            18: 'Lander',
            19: 'Sub',
            // Add more unit types as needed
        };
        
        return unitNames[unitType] || `Unit ${unitType}`;
    }
    
    /**
     * Get current tool state for external queries
     */
    public getCurrentState() {
        return {
            terrain: this.currentTerrain,
            unit: this.currentUnit,
            brushSize: this.currentBrushSize,
            playerId: this.currentPlayerId,
            placementMode: this.currentPlacementMode
        };
    }
}