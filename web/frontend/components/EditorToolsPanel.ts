import { BaseComponent } from './Component';
import { EventBus, EditorEventTypes, TerrainSelectedPayload, UnitSelectedPayload, BrushSizeChangedPayload, PlacementModeChangedPayload, PlayerChangedPayload } from './EventBus';

const BRUSH_SIZE_NAMES = ['Single (1 hex)', 'Small (3 hexes)', 'Medium (5 hexes)', 'Large (9 hexes)', 'X-Large (15 hexes)', 'XX-Large (25 hexes)'];

/**
 * EditorToolsPanel Component - Manages terrain/unit selection, brush size, and player controls
 * 
 * Responsibilities:
 * - Handle terrain button clicks and emit terrain selection events
 * - Handle unit button clicks and emit unit selection events  
 * - Handle brush size changes and emit brush size events
 * - Handle player selection changes and emit player events
 * - Manage visual selection state for tools
 * - Maintain radio button behavior for terrain/unit selection
 * 
 * Does NOT handle:
 * - Layout control (managed by parent/CSS)
 * - Direct state management (emits events instead)
 * - Cross-component DOM access (scoped to root element only)
 */
export class EditorToolsPanel extends BaseComponent {
    // Current selection state for UI updates
    private currentTerrain: number = 1;
    private currentUnit: number = 0;
    private currentBrushSize: number = 0;
    private currentPlayerId: number = 1;
    private currentPlacementMode: 'terrain' | 'unit' | 'clear' = 'terrain';
    
    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('editor-tools-panel', rootElement, eventBus, debugMode);
    }
    
    protected initializeComponent(): void {
        this.log('Initializing EditorToolsPanel component');
        
        // Subscribe to external state changes to keep UI in sync
        this.subscribe<TerrainSelectedPayload>(EditorEventTypes.TERRAIN_SELECTED, (payload) => {
            this.updateTerrainSelection(payload.data.terrainType);
        });
        
        this.subscribe<UnitSelectedPayload>(EditorEventTypes.UNIT_SELECTED, (payload) => {
            this.updateUnitSelection(payload.data.unitType);
        });
        
        this.subscribe<BrushSizeChangedPayload>(EditorEventTypes.BRUSH_SIZE_CHANGED, (payload) => {
            this.updateBrushSizeSelection(payload.data.brushSize);
        });
        
        this.subscribe<PlayerChangedPayload>(EditorEventTypes.PLAYER_CHANGED, (payload) => {
            this.updatePlayerSelection(payload.data.playerId);
        });
        
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
                        this.currentPlacementMode = 'clear';
                        this.updateButtonSelection(clickedButton);
                        
                        this.emit<PlacementModeChangedPayload>(EditorEventTypes.PLACEMENT_MODE_CHANGED, {
                            mode: 'clear'
                        });
                        
                        this.log('Selected clear mode');
                    } else {
                        // Terrain selection
                        this.currentTerrain = terrainValue;
                        this.currentPlacementMode = 'terrain';
                        this.updateButtonSelection(clickedButton);
                        
                        this.emit<TerrainSelectedPayload>(EditorEventTypes.TERRAIN_SELECTED, {
                            terrainType: terrainValue,
                            terrainName: terrainName
                        });
                        
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
                    
                    this.currentUnit = unitValue;
                    this.currentPlacementMode = 'unit';
                    this.updateButtonSelection(clickedButton);
                    
                    // Get current player selection
                    const playerSelect = this.findElement('#unit-player-color') as HTMLSelectElement;
                    if (playerSelect) {
                        this.currentPlayerId = parseInt(playerSelect.value);
                    }
                    
                    this.emit<UnitSelectedPayload>(EditorEventTypes.UNIT_SELECTED, {
                        unitType: unitValue,
                        unitName: unitName,
                        playerId: this.currentPlayerId
                    });
                    
                    this.log(`Selected unit: ${unitValue} (${unitName}) for player ${this.currentPlayerId}`);
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
                
                this.currentBrushSize = brushSize;
                
                this.emit<BrushSizeChangedPayload>(EditorEventTypes.BRUSH_SIZE_CHANGED, {
                    brushSize: brushSize,
                    sizeName: sizeName
                });
                
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
                
                this.currentPlayerId = playerId;
                
                this.emit<PlayerChangedPayload>(EditorEventTypes.PLAYER_CHANGED, {
                    playerId: playerId
                });
                
                this.log(`Player changed to: ${playerId}`);
            });
            
            this.log('Bound player control');
        } else {
            this.log('Player control not found');
        }
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
     * Update terrain selection state (called from external events)
     */
    private updateTerrainSelection(terrainType: number): void {
        const terrainButton = this.findElement(`[data-terrain="${terrainType}"]`);
        if (terrainButton) {
            this.updateButtonSelection(terrainButton as HTMLElement);
            this.currentTerrain = terrainType;
            this.currentPlacementMode = terrainType === 0 ? 'clear' : 'terrain';
        }
    }
    
    /**
     * Update unit selection state (called from external events)
     */
    private updateUnitSelection(unitType: number): void {
        const unitButton = this.findElement(`[data-unit="${unitType}"]`);
        if (unitButton) {
            this.updateButtonSelection(unitButton as HTMLElement);
            this.currentUnit = unitType;
            this.currentPlacementMode = 'unit';
        }
    }
    
    /**
     * Update brush size selection state (called from external events)
     */
    private updateBrushSizeSelection(brushSize: number): void {
        const brushSizeSelect = this.findElement('#brush-size') as HTMLSelectElement;
        if (brushSizeSelect) {
            brushSizeSelect.value = brushSize.toString();
            this.currentBrushSize = brushSize;
        }
    }
    
    /**
     * Update player selection state (called from external events)
     */
    private updatePlayerSelection(playerId: number): void {
        const playerSelect = this.findElement('#unit-player-color') as HTMLSelectElement;
        if (playerSelect) {
            playerSelect.value = playerId.toString();
            this.currentPlayerId = playerId;
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