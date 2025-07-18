import { PhaserMapEditor } from './phaser/PhaserMapEditor';

/**
 * PhaserPanel handles the Phaser.js-based map editor interface
 * This component manages the Phaser editor lifecycle and provides
 * integration with the main MapEditorPage
 */
export class PhaserPanel {
    private phaserEditor: PhaserMapEditor | null = null;
    private containerElement: HTMLElement | null = null;
    private isInitialized: boolean = false;
    
    // Event callbacks
    private onTileClickCallback: ((q: number, r: number) => void) | null = null;
    private onMapChangeCallback: (() => void) | null = null;
    private onLogCallback: ((message: string) => void) | null = null;
    
    constructor() {
        // Constructor kept minimal - initialize() must be called separately
    }
    
    /**
     * Initialize the Phaser panel with a container element
     */
    public initialize(containerId: string): boolean {
        try {
            this.containerElement = document.getElementById(containerId);
            if (!this.containerElement) {
                throw new Error(`Container element with ID '${containerId}' not found`);
            }
            
            // Create Phaser container div
            let phaserContainer = document.getElementById('phaser-container');
            if (!phaserContainer) {
                phaserContainer = document.createElement('div');
                phaserContainer.id = 'phaser-container';
                phaserContainer.style.width = '100%';
                phaserContainer.style.height = '100%';
                phaserContainer.style.minWidth = '800px';
                phaserContainer.style.minHeight = '600px';
                this.containerElement.appendChild(phaserContainer);
            }
            
            // Initialize Phaser editor
            this.phaserEditor = new PhaserMapEditor('phaser-container');
            
            // Set up internal event handlers
            this.setupEventHandlers();
            
            this.isInitialized = true;
            this.log('Phaser panel initialized successfully');
            
            return true;
            
        } catch (error) {
            this.log(`Failed to initialize Phaser panel: ${error}`);
            return false;
        }
    }
    
    /**
     * Set up event handlers for Phaser editor
     */
    private setupEventHandlers(): void {
        if (!this.phaserEditor) return;
        
        // Handle tile clicks
        this.phaserEditor.onTileClick((q, r) => {
            this.log(`Tile clicked: Q=${q}, R=${r}`);
            if (this.onTileClickCallback) {
                this.onTileClickCallback(q, r);
            }
        });
        
        // Handle map changes
        this.phaserEditor.onMapChange(() => {
            this.log('Map changed');
            if (this.onMapChangeCallback) {
                this.onMapChangeCallback();
            }
        });
    }
    
    /**
     * Paint a tile at the specified coordinates
     */
    public paintTile(q: number, r: number, terrain: number, color: number = 0, brushSize: number = 0): boolean {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot paint tile');
            return false;
        }
        
        try {
            this.phaserEditor.paintTile(q, r, terrain, color, brushSize);
            this.log(`Painted terrain ${terrain} at Q=${q}, R=${r} with brush size ${brushSize}`);
            return true;
        } catch (error) {
            this.log(`Failed to paint tile: ${error}`);
            return false;
        }
    }
    
    /**
     * Set terrain type for painting
     */
    public setTerrain(terrain: number): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot set terrain');
            return;
        }
        
        this.phaserEditor.setTerrain(terrain);
        this.log(`Terrain set to: ${terrain}`);
    }
    
    /**
     * Set brush size for painting
     */
    public setBrushSize(size: number): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot set brush size');
            return;
        }
        
        this.phaserEditor.setBrushSize(size);
        this.log(`Brush size set to: ${size}`);
    }
    
    /**
     * Toggle grid visibility
     */
    public setShowGrid(show: boolean): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot toggle grid');
            return;
        }
        
        this.phaserEditor.setShowGrid(show);
        this.log(`Grid visibility set to: ${show}`);
    }
    
    /**
     * Toggle coordinate visibility
     */
    public setShowCoordinates(show: boolean): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot toggle coordinates');
            return;
        }
        
        this.phaserEditor.setShowCoordinates(show);
        this.log(`Coordinate visibility set to: ${show}`);
    }
    
    /**
     * Set theme for editor (light/dark)
     */
    public setTheme(isDark: boolean): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot set theme');
            return;
        }
        
        this.phaserEditor.setTheme(isDark);
        this.log(`Theme set to: ${isDark ? 'dark' : 'light'}`);
    }
    
    /**
     * Remove a tile at the specified coordinates
     */
    public removeTile(q: number, r: number): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot remove tile');
            return;
        }
        
        this.phaserEditor.removeTile(q, r);
        this.log(`Removed tile at Q=${q}, R=${r}`);
    }
    
    /**
     * Paint a unit at the specified coordinates
     */
    public paintUnit(q: number, r: number, unitType: number, playerId: number): boolean {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot paint unit');
            return false;
        }
        
        try {
            this.phaserEditor.paintUnit(q, r, unitType, playerId);
            this.log(`Painted unit ${unitType} (player ${playerId}) at Q=${q}, R=${r}`);
            return true;
        } catch (error) {
            this.log(`Failed to paint unit: ${error}`);
            return false;
        }
    }
    
    /**
     * Remove a unit at the specified coordinates
     */
    public removeUnit(q: number, r: number): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot remove unit');
            return;
        }
        
        try {
            this.phaserEditor.removeUnit(q, r);
            this.log(`Removed unit at Q=${q}, R=${r}`);
        } catch (error) {
            this.log(`Failed to remove unit: ${error}`);
        }
    }
    
    /**
     * Clear all tiles from the map
     */
    public clearAllTiles(): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot clear tiles');
            return;
        }
        
        this.phaserEditor.clearAllTiles();
        this.log('All tiles cleared');
    }
    
    /**
     * Create a test pattern for debugging
     */
    public createTestPattern(): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot create test pattern');
            return;
        }
        
        this.phaserEditor.createTestPattern();
        this.log('Test pattern created');
    }
    
    /**
     * Center camera on specific coordinates
     */
    public centerCamera(q: number = 0, r: number = 0): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot center camera');
            return;
        }
        
        this.phaserEditor.centerCamera(q, r);
        this.log(`Camera centered on Q=${q}, R=${r}`);
    }
    
    /**
     * Set camera zoom level
     */
    public setZoom(zoom: number): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot set zoom');
            return;
        }
        
        this.phaserEditor.setZoom(zoom);
        this.log(`Zoom set to: ${zoom}`);
    }
    
    /**
     * Get current camera zoom level
     */
    public getZoom(): number {
        if (!this.isInitialized || !this.phaserEditor) {
            return 1;
        }
        
        return this.phaserEditor.getZoom();
    }
    
    /**
     * Get current tiles data
     */
    public getTilesData(): Array<{ q: number; r: number; terrain: number; color: number }> {
        if (!this.isInitialized || !this.phaserEditor) {
            return [];
        }
        
        return this.phaserEditor.getTilesData();
    }
    
    /**
     * Get the current viewport center in hex coordinates
     */
    public getViewportCenter(): { q: number; r: number } {
        if (!this.isInitialized || !this.phaserEditor) {
            return { q: 0, r: 0 };
        }
        
        return this.phaserEditor.getViewportCenter();
    }
    
    /**
     * Set tiles data (load a map)
     */
    public setTilesData(tiles: Array<{ q: number; r: number; terrain: number; color: number }>): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot set tiles data');
            return;
        }
        
        this.phaserEditor.setTilesData(tiles);
        this.log(`Loaded ${tiles.length} tiles`);
    }
    
    /**
     * Advanced map generation methods
     */
    public fillAllTerrain(terrain: number, color: number = 0): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot fill terrain');
            return;
        }
        
        this.phaserEditor.fillAllTerrain(terrain, color);
        this.log(`Filled all terrain with type ${terrain}`);
    }
    
    public randomizeTerrain(): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot randomize terrain');
            return;
        }
        
        this.phaserEditor.randomizeTerrain();
        this.log('Terrain randomized');
    }
    
    public createIslandPattern(centerQ: number = 0, centerR: number = 0, radius: number = 5): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot create island pattern');
            return;
        }
        
        this.phaserEditor.createIslandPattern(centerQ, centerR, radius);
        this.log(`Created island pattern at Q=${centerQ}, R=${centerR} with radius ${radius}`);
    }
    
    /**
     * Event callback setters
     */
    public onTileClick(callback: (q: number, r: number) => void): void {
        this.onTileClickCallback = callback;
    }
    
    public onMapChange(callback: () => void): void {
        this.onMapChangeCallback = callback;
    }
    
    public onLog(callback: (message: string) => void): void {
        this.onLogCallback = callback;
    }
    
    /**
     * Show/hide the panel
     */
    public show(): void {
        if (this.containerElement) {
            this.containerElement.style.display = 'block';
        }
    }
    
    public hide(): void {
        if (this.containerElement) {
            this.containerElement.style.display = 'none';
        }
    }
    
    /**
     * Check if panel is initialized
     */
    /**
     * Set callback for when scene is ready
     */
    public onSceneReady(callback: () => void): void {
        if (!this.isInitialized || !this.phaserEditor) {
            console.warn('[PhaserPanel] Cannot set scene ready callback - not initialized');
            return;
        }
        
        this.phaserEditor.onSceneReady(callback);
    }
    
    public getIsInitialized(): boolean {
        return this.isInitialized;
    }
    
    /**
     * Resize the panel
     */
    public resize(width: number, height: number): void {
        if (!this.isInitialized || !this.phaserEditor) {
            return;
        }
        
        this.phaserEditor.resize(width, height);
    }
    
    /**
     * Destroy the panel and clean up resources
     */
    public destroy(): void {
        if (this.phaserEditor) {
            this.phaserEditor.destroy();
            this.phaserEditor = null;
        }
        
        // Remove Phaser container
        const phaserContainer = document.getElementById('phaser-container');
        if (phaserContainer) {
            phaserContainer.remove();
        }
        
        this.isInitialized = false;
        this.containerElement = null;
        this.onTileClickCallback = null;
        this.onMapChangeCallback = null;
        this.onLogCallback = null;
        
        this.log('Phaser panel destroyed');
    }
    
    /**
     * Internal logging method
     */
    private log(message: string): void {
        if (this.onLogCallback) {
            this.onLogCallback(`[PhaserPanel] ${message}`);
        } else {
            console.log(`[PhaserPanel] ${message}`);
        }
    }
}