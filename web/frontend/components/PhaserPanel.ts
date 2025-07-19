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
    private onReferenceScaleChangeCallback: ((x: number, y: number) => void) | null = null;
    
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
        
        // Handle reference scale changes
        this.phaserEditor.onReferenceScaleChange((x: number, y: number) => {
            if (this.onReferenceScaleChangeCallback) {
                this.onReferenceScaleChangeCallback(x, y);
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
     * Clear all units from the map
     */
    public clearAllUnits(): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot clear units');
            return;
        }
        
        this.phaserEditor.clearAllUnits();
        this.log('All units cleared');
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
    
    public getUnitsData(): Array<{ q: number; r: number; unitType: number; playerId: number }> {
        if (!this.isInitialized || !this.phaserEditor) {
            return [];
        }
        
        return this.phaserEditor.getUnitsData();
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
    public async setTilesData(tiles: Array<{ q: number; r: number; terrain: number; color: number }>): Promise<void> {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot set tiles data');
            return;
        }
        
        try {
            await this.phaserEditor.setTilesData(tiles);
            this.log(`Loaded ${tiles.length} tiles`);
        } catch (error) {
            this.log(`Failed to load tiles: ${error}`);
        }
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
    
    public onReferenceScaleChange(callback: (x: number, y: number) => void): void {
        this.onReferenceScaleChangeCallback = callback;
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
    
    // Reference image methods (editor-only)
    
    /**
     * Load reference image from clipboard
     */
    public async loadReferenceFromClipboard(): Promise<boolean> {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot load reference image');
            return false;
        }
        
        try {
            const result = await this.phaserEditor.loadReferenceFromClipboard();
            this.log(result ? 'Reference image loaded from clipboard' : 'No image found in clipboard');
            return result;
        } catch (error) {
            this.log(`Failed to load reference image: ${error}`);
            return false;
        }
    }
    
    /**
     * Load reference image from file
     */
    public async loadReferenceFromFile(file: File): Promise<boolean> {
        this.log(`loadReferenceFromFile called with file: ${file.name} (${file.size} bytes)`);
        
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot load reference image');
            return false;
        }
        
        this.log('Phaser panel initialized, calling phaserEditor.loadReferenceFromFile');
        
        try {
            const result = await this.phaserEditor.loadReferenceFromFile(file);
            this.log(result ? `Reference image loaded from file: ${file.name}` : 'Failed to load file');
            return result;
        } catch (error) {
            this.log(`Failed to load reference image from file: ${error}`);
            return false;
        }
    }
    
    /**
     * Set reference image mode (0=hidden, 1=background, 2=overlay)
     */
    public setReferenceMode(mode: number): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot set reference mode');
            return;
        }
        
        this.phaserEditor.setReferenceMode(mode);
        const modeNames = ['hidden', 'background', 'overlay'];
        this.log(`Reference mode set to: ${modeNames[mode] || mode}`);
    }
    
    /**
     * Set reference image alpha transparency
     */
    public setReferenceAlpha(alpha: number): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot set reference alpha');
            return;
        }
        
        this.phaserEditor.setReferenceAlpha(alpha);
        this.log(`Reference alpha set to: ${alpha}`);
    }
    
    /**
     * Set reference image position (for mode 2)
     */
    public setReferencePosition(x: number, y: number): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot set reference position');
            return;
        }
        
        this.phaserEditor.setReferencePosition(x, y);
        this.log(`Reference position set to: (${x}, ${y})`);
    }
    
    /**
     * Set reference image scale (for mode 2)
     */
    public setReferenceScale(x: number, y: number): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot set reference scale');
            return;
        }
        
        this.phaserEditor.setReferenceScale(x, y);
        this.log(`Reference scale set to: (${x}, ${y})`);
    }
    
    /**
     * Set reference image scale with top-left corner as pivot
     */
    public setReferenceScaleFromTopLeft(x: number, y: number): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot set reference scale from top-left');
            return;
        }
        
        this.phaserEditor.setReferenceScaleFromTopLeft(x, y);
        this.log(`Reference scale set from top-left to: (${x}, ${y})`);
    }
    
    /**
     * Get reference image state
     */
    public getReferenceState(): {
        mode: number;
        alpha: number;
        position: { x: number; y: number };
        scale: { x: number; y: number };
        hasImage: boolean;
    } | null {
        if (!this.isInitialized || !this.phaserEditor) {
            return null;
        }
        
        return this.phaserEditor.getReferenceState();
    }
    
    /**
     * Clear reference image
     */
    public clearReferenceImage(): void {
        if (!this.isInitialized || !this.phaserEditor) {
            this.log('Phaser panel not initialized - cannot clear reference image');
            return;
        }
        
        this.phaserEditor.clearReferenceImage();
        this.log('Reference image cleared');
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