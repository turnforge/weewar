import * as Phaser from 'phaser';
import { hexToRowCol, hexToPixel, pixelToHex, HexCoord, PixelCoord } from './hexUtils';
import { Unit, Tile, World } from '../World';
import { LayerManager } from './LayerSystem';
import { BaseMapLayer, MapLayerCallbacks } from './layers/BaseMapLayer';

export class PhaserWorldScene extends Phaser.Scene {
    // Phaser game instance (self-contained) - renamed to avoid conflict with Phaser's game property
    private phaserGame: Phaser.Game | null = null;
    private containerElement: HTMLElement | null = null;
    private isInitialized: boolean = false;
    private initializePromise: Promise<void> | null = null;
    private initializeResolver: (() => void) | null = null;

    protected tileWidth: number = 64;
    protected tileHeight: number = 64;
    protected yIncrement: number = 48; // 3/4 * tileHeight for pointy-topped hexes
    
    // World as single source of truth for game data
    public world: World | null = null;
    
    // Visual sprite maps (for rendering only, not game data)
    protected tileSprites: Map<string, Phaser.GameObjects.Sprite> = new Map();
    protected unitSprites: Map<string, Phaser.GameObjects.Sprite> = new Map();
    protected gridGraphics: Phaser.GameObjects.Graphics | null = null;
    protected coordinateTexts: Map<string, Phaser.GameObjects.Text> = new Map();
    
    protected showGrid: boolean = false;
    protected showCoordinates: boolean = false;
    
    // Theme management - initialize with current theme state
    protected isDarkTheme: boolean = document.documentElement.classList.contains('dark');
    
    // Camera controls
    protected cursors: Phaser.Types.Input.Keyboard.CursorKeys | null = null;
    protected wasdKeys: any = null;
    protected zoomSpeed: number = 0.01;
    protected panSpeed: number = 100;

    // Layer system for managing overlays and interactions
    protected layerManager: LayerManager | null = null;
    protected baseMapLayer: BaseMapLayer | null = null;
    
    // Game interaction callbacks (optional, only used by GameViewerPage)
    protected tileClickCallback?: ((q: number, r: number) => boolean)
    protected unitClickCallback?: ((q: number, r: number) => boolean)
    
    // Mouse interaction
    protected isMouseDown: boolean = false;
    protected lastPointerPosition: { x: number; y: number } | null = null;
    protected hasDragged: boolean = false;
    
    // Asset loading
    private terrainsLoaded: boolean = false;
    private unitsLoaded: boolean = false;
    private sceneReadyCallback: (() => void) | null = null;
    private assetsReadyPromise: Promise<void> | null = null;
    private assetsReadyResolver: (() => void) | null = null;
    
    constructor(config?: string | Phaser.Types.Scenes.SettingsConfig) {
        super(config || { key: 'PhaserWorldScene' });
    }

    /**
     * Initialize the scene with its own Phaser.Game instance
     * @param containerId - ID of the HTML element to render into
     */
    public async initialize(containerId: string): Promise<void> {
        if (this.isInitialized) {
            console.log('[PhaserWorldScene] Already initialized');
            return;
        }

        if (this.initializePromise) {
            return this.initializePromise;
        }

        this.initializePromise = new Promise<void>((resolve) => {
            this.initializeResolver = resolve;
        });

        this.containerElement = document.getElementById(containerId);
        if (!this.containerElement) {
            throw new Error(`Container element with ID '${containerId}' not found`);
        }

        // Ensure container has proper styling for Phaser
        this.containerElement.style.width = '100%';
        this.containerElement.style.height = '100%';
        this.containerElement.style.minWidth = '600px';
        this.containerElement.style.minHeight = '400px';

        // Get container dimensions
        const containerWidth = this.containerElement.clientWidth || 800;
        const containerHeight = this.containerElement.clientHeight || 600;
        const width = Math.max(containerWidth, 400);
        const height = Math.max(containerHeight, 300);

        const config: Phaser.Types.Core.GameConfig = {
            type: Phaser.AUTO,
            parent: this.containerElement,
            width: width,
            height: height,
            backgroundColor: '#2c3e50',
            scene: this, // Use this scene instance directly
            scale: {
                mode: Phaser.Scale.RESIZE,
                width: width,
                height: height
            },
            physics: {
                default: 'arcade',
                arcade: {
                    debug: false
                }
            },
            input: {
                keyboard: true,
                mouse: true
            },
            render: {
                pixelArt: true,
                antialias: false
            }
        };

        this.phaserGame = new Phaser.Game(config);
            
        return this.initializePromise;
    }

    /**
     * Check if the scene is initialized and ready
     */
    public getIsInitialized(): boolean {
        return this.isInitialized;
    }

    /**
     * Destroy the scene and its Phaser.Game instance
     */
    public destroy(): void {
        console.log('[PhaserWorldScene] Destroying scene');
        
        // Clean up layer system
        if (this.layerManager) {
            this.layerManager.destroy();
            this.layerManager = null;
        }
        this.baseMapLayer = null;
        
        if (this.phaserGame) {
            this.phaserGame.destroy(true);
            this.phaserGame = null;
        }
        
        this.containerElement = null;
        this.isInitialized = false;
        this.initializePromise = null;
        this.initializeResolver = null;
        this.world = null;
    }

    /**
     * Set the World instance as the single source of truth for game data
     */
    public setWorld(world: World): void {
        this.world = world;
        console.log('[PhaserWorldScene] World set as source of truth');
    }

    /**
     * Set interaction callbacks for game-specific functionality
     * @param tileCallback - Called when tile is clicked, return true to emit event, false to suppress
     * @param unitCallback - Called when unit is clicked, return true to emit event, false to suppress
     */
    public setInteractionCallbacks(
        tileCallback?: (q: number, r: number) => boolean,
        unitCallback?: (q: number, r: number) => boolean
    ): void {
        console.log('[PhaserWorldScene] setInteractionCallbacks called');
        console.log('[PhaserWorldScene] Received tileCallback:', !!tileCallback);
        console.log('[PhaserWorldScene] Received unitCallback:', !!unitCallback);
        
        this.tileClickCallback = tileCallback
        this.unitClickCallback = unitCallback
        
        // Update base map layer callbacks if available
        if (this.baseMapLayer) {
            this.baseMapLayer.setCallbacks({
                onTileClicked: tileCallback,
                onUnitClicked: unitCallback,
                onEmptySpaceClicked: tileCallback
            });
        }
        
        console.log('[PhaserWorldScene] Stored tileClickCallback:', !!this.tileClickCallback);
        console.log('[PhaserWorldScene] Stored unitClickCallback:', !!this.unitClickCallback);
        console.log('[PhaserWorldScene] Interaction callbacks set successfully');
    }

    
    preload() {
        // Set up assets ready promise
        this.assetsReadyPromise = new Promise<void>((resolve) => {
            this.assetsReadyResolver = resolve;
        });
        
        // Track when all assets are loaded
        this.load.on('complete', () => {
            console.log('[PhaserWorldScene] All assets loaded successfully');
            this.terrainsLoaded = true;
            this.unitsLoaded = true;
            if (this.assetsReadyResolver) {
                this.assetsReadyResolver();
            }
        });
        
        // Load terrain assets
        this.loadTerrainAssets();
        
        // Load unit assets (for future use)
        this.loadUnitAssets();
    }
    
    create() {
        // Initialize graphics for grid
        this.gridGraphics = this.add.graphics();
        
        // Set up camera controls
        this.setupCameraControls();
        
        // Set up layer system
        this.setupLayerSystem();
        
        // Set up input handling
        this.setupInputHandling();
        
        // Set camera bounds to allow infinite scrolling
        this.cameras.main.setBounds(-10000, -10000, 20000, 20000);
        
        // Initialize grid and coordinates display
        this.updateGridDisplay();
        this.setShowCoordinates(this.showCoordinates);
        
        // Set initial theme
        this.updateTheme();
        
        console.log('[PhaserWorldScene] Scene created successfully');
        
        // Mark as initialized and resolve promise
        this.isInitialized = true;
        console.log('[PhaserWorldScene] isInitialized set to true');
        if (this.initializeResolver) {
            this.initializeResolver();
            this.initializeResolver = null;
        }
        
        // Trigger scene ready callback if set
        if (this.sceneReadyCallback) {
            this.sceneReadyCallback();
            this.sceneReadyCallback = null;
        }
    }
    
    private loadTerrainAssets() {
        // City/Player terrains (have color variations 0-12 for different players)
        const cityTerrains = [1, 2, 3, 16, 20]; // Land Base, Naval Base, Airport Base, Missile Silo, Mines
        
        // Nature terrains (only have default texture, no player ownership)
        const natureTerrains = [4, 5, 6, 7, 8, 9, 10, 12, 14, 15, 17, 18, 19, 21, 22, 23, 25, 26];
        
        // Load city terrains with all color variations
        cityTerrains.forEach(type => {
            for (let color = 0; color <= 12; color++) {
                const assetPath = `/static/assets/v1/Tiles/${type}/${color}.png`;
                this.load.image(`terrain_${type}_${color}`, assetPath);
                if (color === 0) {
                    this.load.image(`terrain_${type}`, assetPath); // Default alias
                }
            }
        });
        
        // Load nature terrains with only default texture
        natureTerrains.forEach(type => {
            const assetPath = `/static/assets/v1/Tiles/${type}/0.png`;
            this.load.image(`terrain_${type}`, assetPath);
            this.load.image(`terrain_${type}_0`, assetPath); // Color 0 alias for consistency
        });
        
        console.log(`[PhaserWorldScene] Loading city terrain assets: ${cityTerrains.join(', ')} (with colors)`);
        console.log(`[PhaserWorldScene] Loading nature terrain assets: ${natureTerrains.join(', ')} (default only)`);
    }
    
    private loadUnitAssets() {
        // Load all available unit assets with player colors
        const unitTypes = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 24, 25, 26, 27, 28, 29];
        
        unitTypes.forEach(type => {
            for (let color = 0; color <= 12; color++) {
                const assetPath = `/static/assets/v1/Units/${type}/${color}.png`;
                this.load.image(`unit_${type}_${color}`, assetPath);
                if (color === 0) {
                    this.load.image(`unit_${type}`, assetPath); // Default alias
                }
            }
        });
        
        console.log(`[PhaserWorldScene] Loading unit assets: ${unitTypes.join(', ')} (with colors 0-12)`);
        
        // Add completion handlers to track loading progress
        this.load.on('filecomplete-image', (key: string, type: string, data: any) => {
            if (key.startsWith('unit_')) {
                console.log(`[PhaserWorldScene] Unit texture loaded: ${key}`);
            }
        });
        
        this.load.on('complete', () => {
            console.log('[PhaserWorldScene] All asset loading complete');
            // List first few unit textures to verify they're loaded
            const unitKeys = this.textures.getTextureKeys().filter(key => key.startsWith('unit_')).slice(0, 10);
            console.log('[PhaserWorldScene] Sample loaded unit textures:', unitKeys);
        });
    }
    
    private setupCameraControls() {
        // Create cursor keys for camera movement
        // this.cursors = this.input.keyboard!.createCursorKeys();
        
        // Add WASD keys  
        // this.wasdKeys = this.input.keyboard!.addKeys('W,S,A,D');
        
        // Set up document-level keydown listener to handle input context properly
        /*
        document.addEventListener('keydown', (event: KeyboardEvent) => {
            // If user is focused on an input field, don't let Phaser prevent default for arrow keys
            if (this.isInInputContext()) {
                const arrowKeys = ['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown'];
                if (arrowKeys.includes(event.key)) {
                    // Stop the event from reaching Phaser's keyboard handlers
                    event.stopImmediatePropagation();
                }
            }
        }, true); // Use capture phase to intercept before Phaser
       */
        
        // Mouse wheel zoom - zoom around cursor position
        this.input.on('wheel', (pointer: Phaser.Input.Pointer, gameObjects: Phaser.GameObjects.GameObject[], deltaX: number, deltaY: number) => {
            const camera = this.cameras.main;
            const oldZoom = camera.zoom;
            const zoomFactor = deltaY > 0 ? 1 - this.zoomSpeed : 1 + this.zoomSpeed;
            const newZoom = Phaser.Math.Clamp(oldZoom * zoomFactor, 0.1, 3);
            
            // Calculate world coordinates under mouse cursor before zoom
            const worldX = camera.scrollX + (pointer.x - camera.centerX) / oldZoom;
            const worldY = camera.scrollY + (pointer.y - camera.centerY) / oldZoom;
            
            // Apply the zoom
            camera.setZoom(newZoom);
            
            // Calculate new camera position to keep world point under cursor
            const newScrollX = worldX - (pointer.x - camera.centerX) / newZoom;
            const newScrollY = worldY - (pointer.y - camera.centerY) / newZoom;
            
            camera.scrollX = newScrollX;
            camera.scrollY = newScrollY;
        });
    }
    
    /**
     * Set up the layer system for managing overlays and interactions
     */
    private setupLayerSystem(): void {
        console.log('[PhaserWorldScene] Setting up layer system');
        
        // Create layer manager with coordinate conversion functions
        this.layerManager = new LayerManager(
            this,
            (x: number, y: number) => pixelToHex(x, y),
            (q: number, r: number) => this.world?.getTileAt(q, r) || null,
            (q: number, r: number) => this.world?.getUnitAt(q, r) || null
        );
        
        // Create base map layer for default interactions
        this.baseMapLayer = new BaseMapLayer(this, {
            onTileClicked: this.tileClickCallback,
            onUnitClicked: this.unitClickCallback,
            onEmptySpaceClicked: this.tileClickCallback
        });
        
        // Add base map layer to manager
        this.layerManager.addLayer(this.baseMapLayer);
        
        console.log('[PhaserWorldScene] Layer system initialized');
    }
    
    private setupInputHandling() {
        // Mouse/touch interaction handling
        this.input.on('pointerdown', (pointer: Phaser.Input.Pointer) => {
            if (pointer.button === 0) { // Left click only
                this.isMouseDown = true;
                this.hasDragged = false;
                this.lastPointerPosition = { x: pointer.x, y: pointer.y };
            }
        });
        
        this.input.on('pointerup', (pointer: Phaser.Input.Pointer) => {
            if (pointer.button === 0) { // Left click only
                // Only handle click if we didn't drag
                if (!this.hasDragged) {
                    // Use layer system for click handling if available, fallback to direct handling
                    if (this.layerManager) {
                        const handled = this.layerManager.handleClick(pointer);
                        if (!handled) {
                            console.log('[PhaserWorldScene] No layer handled click, using fallback');
                            const worldPoint = this.cameras.main.getWorldPoint(pointer.x, pointer.y);
                            const hexCoords = pixelToHex(worldPoint.x, worldPoint.y);
                            this.onTileClick(hexCoords.q, hexCoords.r);
                        }
                    } else {
                        // Fallback to direct handling
                        const worldPoint = this.cameras.main.getWorldPoint(pointer.x, pointer.y);
                        const hexCoords = pixelToHex(worldPoint.x, worldPoint.y);
                        this.onTileClick(hexCoords.q, hexCoords.r);
                    }
                }
                
                // Reset state
                this.isMouseDown = false;
                this.lastPointerPosition = null;
                this.hasDragged = false;
            }
        });
        
        this.input.on('pointermove', (pointer: Phaser.Input.Pointer) => {
            if (this.isMouseDown && this.lastPointerPosition) {
                const deltaX = pointer.x - this.lastPointerPosition.x;
                const deltaY = pointer.y - this.lastPointerPosition.y;
                
                // Check if we've moved enough to consider it a drag
                const dragThreshold = 5; // pixels
                if (Math.abs(deltaX) > dragThreshold || Math.abs(deltaY) > dragThreshold) {
                    this.hasDragged = true;
                }
                
                // Pan camera
                this.cameras.main.scrollX -= deltaX / this.cameras.main.zoom;
                this.cameras.main.scrollY -= deltaY / this.cameras.main.zoom;
                
                this.lastPointerPosition = { x: pointer.x, y: pointer.y };
            }
        });
    }
    
    update() {
        // Handle camera movement with keyboard only if not in input context
        if (!this.isInInputContext()) {
            const camera = this.cameras.main;
            
            if (this.cursors?.left.isDown || this.wasdKeys?.A.isDown) {
                camera.scrollX -= this.panSpeed * (1 / camera.zoom);
            }
            if (this.cursors?.right.isDown || this.wasdKeys?.D.isDown) {
                camera.scrollX += this.panSpeed * (1 / camera.zoom);
            }
            if (this.cursors?.up.isDown || this.wasdKeys?.W.isDown) {
                camera.scrollY -= this.panSpeed * (1 / camera.zoom);
            }
            if (this.cursors?.down.isDown || this.wasdKeys?.S.isDown) {
                camera.scrollY += this.panSpeed * (1 / camera.zoom);
            }
        }
        
        // Update grid and coordinates when camera moves or zooms
        this.updateGridDisplay();
        this.updateCoordinatesDisplay();
    }
    
    /**
     * Check if user is currently focused on an input field
     */
    private isInInputContext(): boolean {
        const activeElement = document.activeElement as HTMLElement;
        if (!activeElement) return false;
        
        const tagName = activeElement.tagName.toLowerCase();
        return (
            tagName === 'input' ||
            tagName === 'textarea' ||
            tagName === 'select' ||
            activeElement.contentEditable === 'true' ||
            activeElement.closest('.modal') !== null ||
            activeElement.closest('[contenteditable="true"]') !== null
        );
    }

    // Public methods for world manipulation
    public setTile(tile: Tile) {
        const q = tile.q;
        const r = tile.r;
        const key = `${q},${r}`;
        const terrainType = tile.tileType;
        const color = tile.player;
        const position = hexToPixel(q, r);
        
        // Remove existing tile if it exists
        if (this.tileSprites.has(key)) {
            this.tileSprites.get(key)?.destroy();
        }
        
        // Create new tile sprite
        const textureKey = `terrain_${terrainType}_${color}`;
        
        if (this.textures.exists(textureKey)) {
            const tileSprite = this.add.sprite(position.x, position.y, textureKey);
            tileSprite.setOrigin(0.5, 0.5);
            this.tileSprites.set(key, tileSprite);
        } else {
            console.warn(`[PhaserWorldScene] Texture not found: ${textureKey}`);
            console.warn(`[PhaserWorldScene] Available textures:`, this.textures.list);
            
            // Try fallback to basic terrain texture without color
            const fallbackKey = `terrain_${terrainType}`;
            if (this.textures.exists(fallbackKey)) {
                console.log(`[PhaserWorldScene] Using fallback texture: ${fallbackKey}`);
                const tileSprite = this.add.sprite(position.x, position.y, fallbackKey);
                tileSprite.setOrigin(0.5, 0.5);
                this.tileSprites.set(key, tileSprite);
            } else {
                console.error(`[PhaserWorldScene] Fallback texture also not found: ${fallbackKey}`);
            }
        }
        
        // Update coordinate text if enabled
        if (this.showCoordinates) {
            this.updateCoordinateText(q, r);
        }
    }
    
    public removeTile(q: number, r: number) {
        const key = `${q},${r}`;
        
        if (this.tileSprites.has(key)) {
            this.tileSprites.get(key)?.destroy();
            this.tileSprites.delete(key);
        }
        
        // Remove coordinate text
        if (this.coordinateTexts.has(key)) {
            this.coordinateTexts.get(key)?.destroy();
            this.coordinateTexts.delete(key);
        }
    }
    
    public clearAllTiles() {
        this.tileSprites.forEach(tile => tile.destroy());
        this.tileSprites.clear();
        
        this.coordinateTexts.forEach(text => text.destroy());
        this.coordinateTexts.clear();
    }
    
    // Unit management methods
    public setUnit(unit: Unit) {
        const q = unit.q;
        const r = unit.r;
        const unitType = unit.unitType;
        const color = unit.player;
        const key = `${q},${r}`;
        const position = hexToPixel(q, r);
        
        console.log(`[PhaserWorldScene] setUnit called: q=${q}, r=${r}, unitType=${unitType}, color=${color}`);
        
        // Remove existing unit if it exists
        if (this.unitSprites.has(key)) {
            this.unitSprites.get(key)?.destroy();
        }
        
        // Create new unit sprite
        const textureKey = `unit_${unitType}_${color}`;
        console.log(`[PhaserWorldScene] Looking for texture: ${textureKey}`);
        console.log(`[PhaserWorldScene] Texture exists: ${this.textures.exists(textureKey)}`);
        
        if (this.textures.exists(textureKey)) {
            const unitSprite = this.add.sprite(position.x, position.y, textureKey);
            unitSprite.setOrigin(0.5, 0.5);
            unitSprite.setDepth(10); // Units render above tiles
            this.unitSprites.set(key, unitSprite);
        } else {
            console.warn(`[PhaserWorldScene] Unit texture not found: ${textureKey}`);
            
            // Try fallback to basic unit texture without color
            const fallbackKey = `unit_${unitType}`;
            if (this.textures.exists(fallbackKey)) {
                console.log(`[PhaserWorldScene] Using fallback unit texture: ${fallbackKey}`);
                const unitSprite = this.add.sprite(position.x, position.y, fallbackKey);
                unitSprite.setOrigin(0.5, 0.5);
                unitSprite.setDepth(10);
                this.unitSprites.set(key, unitSprite);
            } else {
                console.error(`[PhaserWorldScene] Fallback unit texture also not found: ${fallbackKey}`);
            }
        }
    }
    
    public removeUnit(q: number, r: number) {
        const key = `${q},${r}`;
        
        if (this.unitSprites.has(key)) {
            this.unitSprites.get(key)?.destroy();
            this.unitSprites.delete(key);
        }
    }
    
    public clearAllUnits() {
        this.unitSprites.forEach(unit => unit.destroy());
        this.unitSprites.clear();
    }
    
    public setShowGrid(show: boolean) {
        this.showGrid = show;
        this.updateGridDisplay();
    }
    
    public setShowCoordinates(show: boolean) {
        this.showCoordinates = show;
        this.updateCoordinatesDisplay();
    }
    
    public setTheme(isDark: boolean) {
        this.isDarkTheme = isDark;
        this.updateTheme();
    }
    
    private updateTheme() {
        // Update camera background color based on theme
        const backgroundColor = this.isDarkTheme ? 0x1f2937 : 0xf3f4f6; // gray-800 : gray-100
        this.cameras.main.setBackgroundColor(backgroundColor);
        
        // Update grid and coordinates
        this.updateGridDisplay();
        this.updateCoordinatesDisplay();
    }
    
    private updateCoordinatesDisplay() {
        // Clear existing coordinate texts
        this.coordinateTexts.forEach(text => text.destroy());
        this.coordinateTexts.clear();
        
        if (!this.showCoordinates) return;
        
        // Get camera bounds (same logic as grid)
        const camera = this.cameras.main;
        const worldView = camera.worldView;
        
        const padding = Math.max(this.tileWidth, this.tileHeight) * 2;
        const minX = worldView.x - padding;
        const maxX = worldView.x + worldView.width + padding;
        const minY = worldView.y - padding;
        const maxY = worldView.y + worldView.height + padding;
        
        // Convert bounds to hex coordinates
        const topLeft = pixelToHex(minX, minY);
        const topRight = pixelToHex(maxX, minY);
        const bottomLeft = pixelToHex(minX, maxY);
        const bottomRight = pixelToHex(maxX, maxY);
        
        // Find the bounding box in hex coordinates
        const minQ = Math.min(topLeft.q, topRight.q, bottomLeft.q, bottomRight.q) - 1;
        const maxQ = Math.max(topLeft.q, topRight.q, bottomLeft.q, bottomRight.q) + 1;
        const minR = Math.min(topLeft.r, topRight.r, bottomLeft.r, bottomRight.r) - 1;
        const maxR = Math.max(topLeft.r, topRight.r, bottomLeft.r, bottomRight.r) + 1;
        
        // Show coordinates for all visible hexes
        for (let q = minQ; q <= maxQ; q++) {
            for (let r = minR; r <= maxR; r++) {
                this.updateCoordinateText(q, r);
            }
        }
    }
    
    private updateGridDisplay() {
        if (!this.gridGraphics) return;
        
        this.gridGraphics.clear();
        
        if (!this.showGrid) return;
        
        // Get camera bounds
        const camera = this.cameras.main;
        const worldView = camera.worldView;
        
        // Add some padding to ensure grid covers entire visible area
        const padding = Math.max(this.tileWidth, this.tileHeight) * 2;
        const minX = worldView.x - padding;
        const maxX = worldView.x + worldView.width + padding;
        const minY = worldView.y - padding;
        const maxY = worldView.y + worldView.height + padding;
        
        // Convert bounds to hex coordinates to find range
        const topLeft = pixelToHex(minX, minY);
        const topRight = pixelToHex(maxX, minY);
        const bottomLeft = pixelToHex(minX, maxY);
        const bottomRight = pixelToHex(maxX, maxY);
        
        // Find the bounding box in hex coordinates
        const minQ = Math.min(topLeft.q, topRight.q, bottomLeft.q, bottomRight.q) - 2;
        const maxQ = Math.max(topLeft.q, topRight.q, bottomLeft.q, bottomRight.q) + 2;
        const minR = Math.min(topLeft.r, topRight.r, bottomLeft.r, bottomRight.r) - 2;
        const maxR = Math.max(topLeft.r, topRight.r, bottomLeft.r, bottomRight.r) + 2;
        
        // Draw grid lines with theme-appropriate colors
        const gridColor = this.isDarkTheme ? 0xcccccc : 0x444444;
        const gridAlpha = this.isDarkTheme ? 0.4 : 0.3;
        this.gridGraphics.lineStyle(1, gridColor, gridAlpha);
        
        // Draw grid for all visible hexes
        for (let q = minQ; q <= maxQ; q++) {
            for (let r = minR; r <= maxR; r++) {
                this.drawHexagon(q, r);
            }
        }
    }
    
    protected drawHexagon(q: number, r: number) {
        if (!this.gridGraphics) return;
        
        const position = hexToPixel(q, r);
        const halfWidth = this.tileWidth / 2;
        const halfHeight = this.tileHeight / 2;
        
        // Draw pointy-topped hexagon outline based on tile dimensions
        this.gridGraphics.beginPath();
        
        // Pointy-topped hexagon vertices (starting from top point, going clockwise)
        const vertices = [
            { x: position.x, y: position.y - halfHeight },                    // Top
            { x: position.x + halfWidth * 0.866, y: position.y - halfHeight * 0.5 }, // Top-right
            { x: position.x + halfWidth * 0.866, y: position.y + halfHeight * 0.5 }, // Bottom-right
            { x: position.x, y: position.y + halfHeight },                    // Bottom
            { x: position.x - halfWidth * 0.866, y: position.y + halfHeight * 0.5 }, // Bottom-left
            { x: position.x - halfWidth * 0.866, y: position.y - halfHeight * 0.5 }  // Top-left
        ];
        
        this.gridGraphics.moveTo(vertices[0].x, vertices[0].y);
        for (let i = 1; i < vertices.length; i++) {
            this.gridGraphics.lineTo(vertices[i].x, vertices[i].y);
        }
        this.gridGraphics.closePath();
        this.gridGraphics.strokePath();
    }

    /**
     * Draw hexagon shape at given position
     */
    protected drawHexagonShape(graphics: Phaser.GameObjects.Graphics, x: number, y: number, size: number): void {
        const points: number[] = [];
        
        for (let i = 0; i < 6; i++) {
            const angle = (Math.PI / 3) * i;
            const hexX = x + size * Math.cos(angle);
            const hexY = y + size * Math.sin(angle);
            points.push(hexX, hexY);
        }
        
        graphics.fillPoints(points, true);
        graphics.strokePoints(points, true);
    }
    
    private updateCoordinateText(q: number, r: number) {
        if (!this.showCoordinates) return;
        
        const key = `${q},${r}`;
        const position = hexToPixel(q, r);
        
        // Remove existing text
        if (this.coordinateTexts.has(key)) {
            this.coordinateTexts.get(key)?.destroy();
        }
        
        // Create coordinate text with both Q/R and row/col using proper conversion
        const { row, col } = hexToRowCol(q, r);
        
        const coordText = `QR:${q}, ${r}\nRC:${row}, ${col}`;
        
        const textColor = this.isDarkTheme ? '#ffffff' : '#000000';
        const strokeColor = this.isDarkTheme ? '#000000' : '#ffffff';
        
        const text = this.add.text(position.x, position.y, coordText, {
            fontSize: '10px',
            color: textColor,
            stroke: strokeColor,
            strokeThickness: 1,
            align: 'center'
        });
        
        text.setOrigin(0.5, 0.5);
        this.coordinateTexts.set(key, text);
    }
    
    // Test method to create a simple pattern
    public createTestPattern() {
        // Clear existing tiles
        this.clearAllTiles();
        
        // Create a simple pattern with different terrain types
        const patterns = [
            { q: 0, r: 0, tileType: 1, player: 0 },   // Grass
            { q: 1, r: 0, tileType: 2, player: 0 },   // Desert
            { q: -1, r: 0, tileType: 3, player: 0 },  // Water
            { q: 0, r: 1, tileType: 16, player: 0 },  // Mountain
            { q: 0, r: -1, tileType: 20, player: 0 }, // Rock
            { q: 1, r: -1, tileType: 1, player: 1 },  // Grass (different player)
            { q: -1, r: 1, tileType: 2, player: 2 },  // Desert (different player)
        ];
        
        patterns.forEach(pattern => { this.setTile(pattern); });
        
        // Update grid display
        this.updateGridDisplay();
        
        console.log('[PhaserWorldScene] Test pattern created with negative coordinates support');
    }
    
    // Callback for tile clicks (handles both game interaction and editor events)
    protected onTileClick(q: number, r: number) {
        console.log(`[PhaserWorldScene] Tile clicked: Q=${q}, R=${r}`);
        console.log(`[PhaserWorldScene] tileClickCallback:`, !!this.tileClickCallback);
        console.log(`[PhaserWorldScene] unitClickCallback:`, !!this.unitClickCallback);
        
        // Check if there's a unit at this position first for game interaction
        if (this.world) {
            const unit = this.world.getUnitAt(q, r);
            console.log(`[PhaserWorldScene] Unit found at position:`, !!unit);
            if (unit && this.unitClickCallback) {
                console.log(`[PhaserWorldScene] Calling unitClickCallback for unit at Q=${q}, R=${r}`);
                const shouldEmit = this.unitClickCallback(q, r);
                console.log(`[PhaserWorldScene] unitClickCallback returned:`, shouldEmit);
                if (!shouldEmit) {
                    return; // Unit callback handled it and suppressed event
                }
            }
        }
        
        // Call tile callback if set
        if (this.tileClickCallback) {
            console.log(`[PhaserWorldScene] Calling tileClickCallback for Q=${q}, R=${r}`);
            const shouldEmit = this.tileClickCallback(q, r);
            console.log(`[PhaserWorldScene] tileClickCallback returned:`, shouldEmit);
            if (!shouldEmit) {
                return; // Tile callback handled it and suppressed event
            }
        }
        
        // Default behavior: emit the tile click event for WorldEditorPage or other listeners
        console.log(`[PhaserWorldScene] Emitting tileClicked event`);
        this.events.emit('tileClicked', { q, r });
    }
    
    
    /**
     * Load world data into the scene
     */
    public async loadWorldData(world: World): Promise<void> {
        if (!this.isInitialized) {
            throw new Error('[PhaserWorldScene] Scene not initialized. Call initialize() first.');
        }

        // Set world as source of truth
        this.setWorld(world);

        // Wait for assets to be ready before placing tiles/units
        await this.waitForAssetsReady();
        console.log('[PhaserWorldScene] Assets ready, loading world data');
        
        // Clear existing content
        this.clearAllTiles();
        this.clearAllUnits();
        
        // Load tiles from World
        const tiles = world.getAllTiles();
        if (tiles.length > 0) {
            tiles.forEach(tile => {
                this.setTile(tile);
            });
            console.log(`[PhaserWorldScene] Loaded ${tiles.length} tiles`);
        }
        
        // Load units from World
        const units = world.getAllUnits();
        if (units.length > 0) {
            units.forEach(unit => {
                this.setUnit(unit);
            });
            console.log(`[PhaserWorldScene] Loaded ${units.length} units`);
        }
        
        // Center camera on the loaded world
        this.centerCameraOnWorld();
    }

    /**
     * Center the camera on the loaded world by calculating bounds and focusing on center
     */
    public centerCameraOnWorld(): void {
        if (!this.world) {
            console.log('[PhaserWorldScene] Cannot center camera - no world loaded');
            return;
        }
        
        const allTiles = this.world.getAllTiles();
        if (allTiles.length === 0) {
            console.log('[PhaserWorldScene] No tiles to center camera on');
            return;
        }
        
        // Calculate bounds of all tiles
        let minQ = allTiles[0].q;
        let maxQ = allTiles[0].q;
        let minR = allTiles[0].r;
        let maxR = allTiles[0].r;
        
        allTiles.forEach(tile => {
            minQ = Math.min(minQ, tile.q);
            maxQ = Math.max(maxQ, tile.q);
            minR = Math.min(minR, tile.r);
            maxR = Math.max(maxR, tile.r);
        });
        
        // Calculate center point
        const centerQ = Math.floor((minQ + maxQ) / 2);
        const centerR = Math.floor((minR + maxR) / 2);
        
        console.log(`[PhaserWorldScene] Centering camera on Q=${centerQ}, R=${centerR} (bounds: Q=${minQ}-${maxQ}, R=${minR}-${maxR})`);
        
        // Convert hex coordinates to pixel coordinates
        const centerPixel = hexToPixel(centerQ, centerR);
        
        // Center camera on position
        if (this.cameras?.main) {
            this.cameras.main.centerOn(centerPixel.x, centerPixel.y);
        }
    }

    /**
     * Get current zoom level
     */
    public getZoom(): number {
        return this.cameras?.main?.zoom || 1;
    }

    /**
     * Set zoom level
     */
    public setZoom(zoom: number): void {
        if (this.cameras?.main) {
            this.cameras.main.setZoom(zoom);
        }
    }

    /**
     * Resize the scene
     */
    public resize(width?: number, height?: number): void {
        if (this.phaserGame && this.containerElement) {
            const w = width || this.containerElement.clientWidth;
            const h = height || this.containerElement.clientHeight;
            this.phaserGame.scale.resize(w, h);
        }
    }

    // Scene ready callback
    public onSceneReady(callback: () => void): void {
        if (this.isInitialized) {
            // Scene is already ready, call immediately
            callback();
        } else {
            this.sceneReadyCallback = callback;
        }
    }
    
    // Wait for assets to be loaded
    public async waitForAssetsReady(): Promise<void> {
        if (this.terrainsLoaded && this.unitsLoaded) {
            return Promise.resolve();
        }
        
        if (!this.assetsReadyPromise) {
            throw new Error('[PhaserWorldScene] Assets ready promise not initialized');
        }
        
        console.log('[PhaserWorldScene] Waiting for assets to be ready...');
        return this.assetsReadyPromise;
    }
    
    // Get all tiles data (for integration with WASM)
    public getTilesData(): Array<Tile> {
        const tilesData: Array<Tile> = [];
        
        this.tileSprites.forEach((tile, key) => {
            const [q, r] = key.split(',').map(Number);
            // Extract terrain and color from texture key
            const textureKey = tile.texture.key;
            const match = textureKey.match(/terrain_(\d+)_(\d+)/);
            
            if (match) {
                tilesData.push({
                    q,
                    r,
                    tileType: parseInt(match[1]),
                    player: parseInt(match[2])
                });
            }
        });
        
        return tilesData;
    }
    
    // Get all units data (for integration with WASM)
    public getUnitsData(): Array<Unit> {
        const unitsData: Array<Unit> = [];
        
        this.unitSprites.forEach((unit, key) => {
            const [q, r] = key.split(',').map(Number);
            // Extract unitType and playerId from texture key
            const textureKey = unit.texture.key;
            const match = textureKey.match(/unit_(\d+)_(\d+)/);
            
            if (match) {
                unitsData.push({
                    q,
                    r,
                    unitType: parseInt(match[1]),
                    player: parseInt(match[2])
                });
            }
        });
        
        return unitsData;
    }
    
    /**
     * Get access to the layer manager for external layer management
     */
    public getLayerManager(): LayerManager | null {
        return this.layerManager;
    }
}
