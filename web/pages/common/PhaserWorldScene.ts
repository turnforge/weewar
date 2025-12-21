import * as Phaser from 'phaser';
import { TILE_HEIGHT, TILE_WIDTH, Y_INCREMENT, hexToRowCol, hexToPixel, pixelToHex, HexCoord, PixelCoord } from './hexUtils';
import { TilesChangedEventData, UnitsChangedEventData, CrossingsChangedEventData, WorldLoadedEventData, Unit, Tile, World } from './World';
import { LayerManager } from './LayerSystem';
import { BaseMapLayer } from './BaseMapLayer';
import { CrossingLayer } from './CrossingLayer';
import { ClickContext } from './LayerSystem';
import { WorldEventType, WorldEventTypes } from './events';
import { LCMComponent, EventBus } from '@panyam/tsappkit';
import { AssetProvider } from '../../assets/providers/AssetProvider';
import { AssetThemePreference } from './AssetThemePreference';
import { AnimationConfig } from './animations/AnimationConfig';
import { ProjectileEffect } from './animations/effects/ProjectileEffect';
import { ExplosionEffect } from './animations/effects/ExplosionEffect';
import { HealBubblesEffect } from './animations/effects/HealBubblesEffect';
import { CaptureEffect } from './animations/effects/CaptureEffect';
import { ExhaustedUnitsHighlightLayer, CapturingFlagLayer } from './HexHighlightLayer';

const UNIT_TILE_RATIO = 0.9

export class PhaserWorldScene extends Phaser.Scene implements LCMComponent {
    // Container element and lifecycle management
    private containerElement: HTMLElement;
    private eventBus: EventBus;
    private debugMode: boolean;
    
    // Phaser game instance (self-contained) - renamed to avoid conflict with Phaser's game property
    private phaserGame: Phaser.Game | null = null;
    private isInitialized: boolean = false;

    protected tileWidth: number = TILE_WIDTH;
    protected tileHeight: number = TILE_HEIGHT;
    // protected yIncrement: number = Y_INCREMENT; // 3/4 * tileHeight for pointy-topped hexes
    
    // World as single source of truth for game data
    public world: World | null = null;
    
    // Visual sprite maps (for rendering only, not game data)
    protected tileSprites: Map<string, Phaser.GameObjects.Sprite> = new Map();
    protected unitSprites: Map<string, Phaser.GameObjects.Sprite> = new Map();
    protected unitLabels: Map<string, {
        healthText: Phaser.GameObjects.Text,
        healthBg?: Phaser.GameObjects.Graphics,
        distanceText?: Phaser.GameObjects.Text
    }> = new Map();
    protected gridGraphics: Phaser.GameObjects.Graphics | null = null;
    protected coordinateTexts: Map<string, Phaser.GameObjects.Text> = new Map();
    
    protected showGrid: boolean = false;
    protected showCoordinates: boolean = false;
    
    // Unit label settings
    protected showUnitHealth: boolean = true;
    protected showUnitDistance: boolean = false;
    protected selectedUnitCoord: { q: number, r: number } | null = null;
    
    // Theme management - initialize with current theme state
    protected isDarkTheme: boolean = document.documentElement.classList.contains('dark');
    
    // Camera controls
    protected cursors: Phaser.Types.Input.Keyboard.CursorKeys | null = null;
    protected wasdKeys: any = null;
    protected zoomSpeed: number = 0.01;
    protected panSpeed: number = 100;

    // Layer system for managing overlays and interactions
    protected layerManager: LayerManager;
    protected baseMapLayer: BaseMapLayer | null = null;
    protected crossingLayer: CrossingLayer | null = null;
    protected exhaustedUnitsLayer: ExhaustedUnitsHighlightLayer | null = null;
    protected capturingFlagLayer: CapturingFlagLayer | null = null;

    // Game interaction callback (unified, only used by GameViewerPage)
    public sceneClickedCallback: (context: ClickContext, layer: string, extra?: any) => void;
    
    // Mouse interaction - removed manual tracking, using Phaser's built-in events
    
    // Asset loading
    private terrainsLoaded: boolean = false;
    private unitsLoaded: boolean = false;
    private sceneReadyCallback: (() => void) | null = null;
    private assetsReadyPromise: Promise<void>;
    private assetsReadyResolver: (() => void) | null = null;
    
    // Camera drag zone
    protected dragZone: Phaser.GameObjects.Zone | null = null;
    
    // Asset provider for loading and managing textures
    private assetProvider: AssetProvider;

    constructor(containerElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false, assetProvider?: AssetProvider) {
        super('phaser-world-scene');
        this.containerElement = containerElement;
        this.eventBus = eventBus;
        this.debugMode = debugMode;
        
        // Initialize asset provider (default to PNG if not provided)
        this.assetProvider = assetProvider || this.createDefaultAssetProvider();
    }

    getContainerElement(): HTMLElement {
      return this.containerElement
    }
    
    /**
     * Create the default asset provider based on configuration
     */
    private createDefaultAssetProvider(): AssetProvider {
        // Sync localStorage to cookie to ensure backend has latest preference
        AssetThemePreference.sync();

        // Check URL parameters for asset configuration
        const urlParams = new URLSearchParams(window.location.search);

        // Theme fallback priority: URL param > localStorage > default
        let themeName = urlParams.get('theme');
        if (!themeName) {
            themeName = localStorage.getItem('assetTheme') || AssetThemePreference.DEFAULT_THEME;
        }

        const svgSize = urlParams.get('svgSize');

        if (this.debugMode) {
            console.log(`[PhaserWorldScene] Using theme: ${themeName}`);
        }

        const rasterSize = svgSize ? parseInt(svgSize) : 160;
        return new AssetProvider(themeName, rasterSize, this.debugMode);
    }
    
    /**
     * Set a new asset provider (requires scene reload)
     */
    public setAssetProvider(provider: AssetProvider): void {
        if (this.assetProvider) {
            this.assetProvider.dispose?.();
        }
        this.assetProvider = provider;
        
        // Would need to reload the scene to apply new assets
        if (this.debugMode) {
            console.log('[PhaserWorldScene] Asset provider changed. Scene reload required.');
        }
    }

    /**
     * Get the current asset provider
     */
    public getAssetProvider(): AssetProvider {
        return this.assetProvider;
    }

    // =========================================================
    // EventBus Integration for World Synchronization
    // =========================================================

    /**
     * Handle events from the EventBus
     */
    public handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
        switch(eventType) {
            case WorldEventTypes.WORLD_LOADED:
                this.handleWorldLoaded(data);
                break;
            
            case WorldEventTypes.TILES_CHANGED:
                this.handleTilesChanged(data);
                break;
            
            case WorldEventTypes.UNITS_CHANGED:
                this.handleUnitsChanged(data);
                break;

            case WorldEventTypes.CROSSINGS_CHANGED:
                this.handleCrossingsChanged(data);
                break;

            case WorldEventTypes.WORLD_CLEARED:
                this.handleWorldCleared();
                break;
            
            default:
                if (this.debugMode) {
                    console.log(`[PhaserWorldScene] Unhandled event: ${eventType}`);
                }
        }
    }

    /**
     * Subscribe to World change events for automatic synchronization
     */
    private subscribeToWorldEvents(): void {
        if (this.debugMode) {
            console.log('[PhaserWorldScene] Subscribing to World events');
        }
        
        // Subscribe to World events for automatic synchronization
        this.eventBus.addSubscription(WorldEventTypes.WORLD_LOADED, null, this);
        this.eventBus.addSubscription(WorldEventTypes.TILES_CHANGED, null, this);
        this.eventBus.addSubscription(WorldEventTypes.UNITS_CHANGED, null, this);
        this.eventBus.addSubscription(WorldEventTypes.CROSSINGS_CHANGED, null, this);
        this.eventBus.addSubscription(WorldEventTypes.WORLD_CLEARED, null, this);
    }

    /**
     * Unsubscribe from all EventBus events
     */
    private unsubscribeFromWorldEvents(): void {
        this.eventBus.removeSubscription(WorldEventTypes.WORLD_LOADED, null, this);
        this.eventBus.removeSubscription(WorldEventTypes.TILES_CHANGED, null, this);
        this.eventBus.removeSubscription(WorldEventTypes.UNITS_CHANGED, null, this);
        this.eventBus.removeSubscription(WorldEventTypes.CROSSINGS_CHANGED, null, this);
        this.eventBus.removeSubscription(WorldEventTypes.WORLD_CLEARED, null, this);
    }

    /**
     * Handle world loaded event - sync all data
     */
    private async handleWorldLoaded(data: WorldLoadedEventData): Promise<void> {
        if (this.debugMode) {
            console.log('[PhaserWorldScene] World loaded, waiting for assets before updating scene');
        }

        // Wait for assets to be ready before trying to create sprites
        await this.waitForAssetsReady();

        if (this.debugMode) {
            console.log('[PhaserWorldScene] Assets ready, updating scene display');
        }

        // Clear existing sprites first to ensure a clean redraw
        this.clearAllTiles();
        this.clearAllUnits();
        if (this.crossingLayer) {
            this.crossingLayer.clearAllCrossings();
        }

        // Load tile data from World into scene
        if (this.world) {
            const tiles = this.world.getAllTiles();
            if (tiles.length > 0) {
                tiles.forEach(tile => this.setTile(tile));
            }

            // Create all units fresh
            const units = this.world.getAllUnits();
            if (units.length > 0) {
                units.forEach(unit => this.setUnit(unit));
            }

            // Load crossings from World (using raw crossings map)
            if (this.crossingLayer) {
                this.crossingLayer.loadCrossings(this.world.crossings);
            }
        }
    }
    
    /**
     * Handle tiles changed event - sync tile updates
     */
    private async handleTilesChanged(data: TilesChangedEventData): Promise<void> {
        // Ensure assets are ready
        await this.waitForAssetsReady();
        
        if (this.debugMode) {
            console.log(`[PhaserWorldScene] Updating ${data.changes.length} tile changes in scene`);
        }
        
        // Update individual tiles in scene based on World changes
        for (const change of data.changes) {
            if (change.tile) {
                this.setTile(change.tile);
            } else {
                // Tile was removed
                this.removeTile(change.q, change.r);
            }
        }
    }
    
    /**
     * Handle units changed event - sync unit updates
     */
    private async handleUnitsChanged(data: UnitsChangedEventData): Promise<void> {
        // Ensure assets are ready
        await this.waitForAssetsReady();
        
        if (this.debugMode) {
            console.log(`[PhaserWorldScene] Updating ${data.changes.length} unit changes in scene`);
        }
        
        // Update individual units in scene based on World changes
        for (const change of data.changes) {
            if (change.unit) {
                this.setUnit(change.unit);
            } else {
                // Unit was removed
                this.removeUnit(change.q, change.r);
            }
        }
    }
    
    /**
     * Handle crossings changed event - sync crossing updates
     */
    private handleCrossingsChanged(data: CrossingsChangedEventData): void {
        if (!this.crossingLayer) return;

        if (this.debugMode) {
            console.log(`[PhaserWorldScene] Updating ${data.changes.length} crossing changes in scene`);
        }

        // Update individual crossings in scene based on World changes
        for (const change of data.changes) {
            if (change.crossing !== null) {
                this.crossingLayer.setCrossing(change.q, change.r, change.crossing);
            } else {
                // Crossing was removed
                this.crossingLayer.removeCrossing(change.q, change.r);
            }
        }
    }

    /**
     * Handle world cleared event - clear all display
     */
    private handleWorldCleared(): void {
        if (this.debugMode) {
            console.log('[PhaserWorldScene] World cleared, clearing scene display');
        }

        this.clearAllTiles();
        this.clearAllUnits();
        if (this.crossingLayer) {
            this.crossingLayer.clearAllCrossings();
        }
    }

    // =========================================================
    // LCMComponent Interface Implementation
    // =========================================================

    // Phase 1: Initialize DOM and discover child components
    async performLocalInit(): Promise<LCMComponent[]> {
        // Validate container element is ready
        if (!this.containerElement) {
            throw new Error('PhaserWorldScene: Container element is required');
        }
        
        if (this.debugMode) {
            console.log('[PhaserWorldScene] DOM validation complete');
        }
        
        // Subscribe to world events for automatic synchronization
        this.subscribeToWorldEvents();
        
        await this.initializePhaser();
        return []; // Leaf component - no children
    }
    
    // Phase 2: Setup dependencies (not needed for this component)
    setupDependencies(): void {}

    // Phase 3: Activate component - Initialize Phaser here
    async activate(): Promise<void> {}

    // Cleanup phase
    deactivate(): void {
        this.destroy();
    }
    
    // Public destroy method for LCMComponent compatibility
    destroy(): void {
        if (this.debugMode) {
            console.log('[PhaserWorldScene] Destroying component');
        }
        this.destroyPhaser();
    }

    /**
     * Initialize Phaser with the container element
     */
    private async initializePhaser(): Promise<void> {
        if (this.isInitialized) {
            console.warn('[PhaserWorldScene] Already initialized');
            return;
        }

        // Get container dimensions from template-styled element
        const containerWidth = this.containerElement.clientWidth || 800;
        const containerHeight = this.containerElement.clientHeight || 600;
        const width = Math.max(containerWidth, 400);
        const height = Math.max(containerHeight, 300);

        const config: Phaser.Types.Core.GameConfig = {
            type: Phaser.AUTO,
            parent: this.containerElement.id || this.containerElement,
            width: width,
            height: height,
            backgroundColor: '#2c3e50',
            scene: this, // Use this scene instance directly
            scale: {
                mode: Phaser.Scale.RESIZE,
                width: width,
                height: height,
                // Prevent canvas from growing parent
                autoCenter: Phaser.Scale.NO_CENTER
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
        this.isInitialized = true;

        // Ensure canvas doesn't cause parent to grow
        // Phaser inserts the canvas into the parent container
        if (this.phaserGame.canvas) {
            const canvas = this.phaserGame.canvas;
            canvas.style.display = 'block';
            canvas.style.width = '100%';
            canvas.style.height = '100%';
            // Note: Do NOT use object-fit:contain as it maintains aspect ratio
            // and causes height to change when width changes. We want the canvas
            // to fill the container completely - Phaser's Scale.RESIZE mode
            // handles the internal rendering correctly.
        }

        // Set up assets ready promise
        this.assetsReadyPromise = new Promise<void>((resolve) => {
            this.assetsReadyResolver = resolve;
        });
    }

    /**
     * Check if the scene is initialized and ready
     */
    public getIsInitialized(): boolean {
        return this.isInitialized;
    }

    /**
     * Clear all cached assets (useful for debugging/reloading themes)
     */
    public clearCache(): void {
        if (!this.textures || !this.cache) {
            console.warn('[PhaserWorldScene] Scene not initialized');
            return;
        }
        
        console.log('[PhaserWorldScene] Clearing all caches...');
        
        // Clear texture cache (except system textures)
        const textureKeys = this.textures.getTextureKeys();
        textureKeys.forEach(key => {
            // Don't remove system textures
            if (key !== '__DEFAULT' && key !== '__MISSING' && key !== '__WHITE') {
                this.textures.remove(key);
            }
        });
        
        // Clear other caches
        this.cache.json.destroy();
        this.cache.text.destroy();
        this.cache.binary.destroy();
        
        console.log('[PhaserWorldScene] Cache cleared');
    }
    
    /**
     * Destroy Phaser game instance and clean up
     */
    private destroyPhaser(): void {
        // Unsubscribe from world events
        this.unsubscribeFromWorldEvents();

        // Clean up layer system
        if (this.layerManager) {
            this.layerManager.destroy();
        }
        this.baseMapLayer = null;
        this.crossingLayer = null;

        if (this.phaserGame) {
            this.phaserGame.destroy(true);
            this.phaserGame = null;
        }
        
        this.isInitialized = false;
        this.world = null;
    }


    /**
     * Set the World instance as the single source of truth for game data
     * Also loads crossings if the crossing layer is ready
     */
    public setWorld(world: World): void {
        this.world = world;
        // Load crossings if the layer is ready
        if (this.crossingLayer && world.crossings) {
            this.crossingLayer.loadCrossings(world.crossings);
        }
    }
    
    async preload() {
        // Configure asset provider
        this.assetProvider.configure(this.load, this);
        
        // Set up progress tracking
        this.assetProvider.onProgress = (progress: number) => {
            if (this.debugMode) {
                // console.log(`[PhaserWorldScene] Loading assets: ${Math.round(progress * 100)}%`);
            }
        };
        
        // Track when all assets are loaded
        this.load.on('complete', async () => {
            // Perform post-processing if needed (e.g., for SVG templates)
            if (this.assetProvider.postProcessAssets) {
                if (this.debugMode) {
                    console.log('[PhaserWorldScene] Post-processing assets...');
                }
                await this.assetProvider.postProcessAssets();
            }
            
            this.terrainsLoaded = true;
            this.unitsLoaded = true;
            if (this.assetsReadyResolver) {
                this.assetsReadyResolver();
            }
        });
        
        // Load assets through provider - await to ensure all assets are queued
        await this.assetProvider.preloadAssets();
        
        if (this.debugMode) {
            console.log('[PhaserWorldScene] All assets queued for loading');
        }
    }
    
    create() {
        // Initialize graphics for grid
        this.gridGraphics = this.add.graphics();

        // Create particle texture for effects
        this.createParticleTexture();

        // Set up camera controls
        this.setupCameraControls();
        
        // Set up layer system
        this.setupLayerSystem();

        // Load crossings if world was set before create() was called
        if (this.world && this.crossingLayer && this.world.crossings) {
            this.crossingLayer.loadCrossings(this.world.crossings);
        }

        // Set up input handling
        this.setupInputHandling();
        
        // Set camera bounds to allow infinite scrolling
        this.cameras.main.setBounds(-10000, -10000, 20000, 20000);
        
        // Initialize grid and coordinates display
        this.updateGridDisplay();
        this.setShowCoordinates(this.showCoordinates);
        
        // Set initial theme
        this.updateTheme();
        
        // Mark as initialized and resolve promise
        this.isInitialized = true;
        
        // Trigger scene ready callback if set
        if (this.sceneReadyCallback) {
            this.sceneReadyCallback();
            this.sceneReadyCallback = null;
        }
    }
    
    // Asset loading is now handled by AssetProvider
    // These methods are no longer needed as standalone functions
    
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
            // Check if any layer wants to handle the scroll (e.g., reference image in overlay mode)
            if (this.layerManager) {
                const layerHandledScroll = this.layerManager.processScroll(pointer, deltaY);
                if (layerHandledScroll) {
                    // Layer handled the scroll, don't zoom camera
                    return;
                }
            }

            const camera = this.cameras.main;
            const oldZoom = camera.zoom;
            const oldScrollX = camera.scrollX;
            const oldScrollY = camera.scrollY;

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

            // Emit camera events for zoom and position changes
            if (oldZoom !== camera.zoom) {
                this.events.emit('camera-zoomed', {
                    zoom: camera.zoom,
                    oldZoom: oldZoom,
                    deltaZoom: camera.zoom - oldZoom
                });
            }

            if (oldScrollX !== camera.scrollX || oldScrollY !== camera.scrollY) {
                this.events.emit('camera-moved', {
                    scrollX: camera.scrollX,
                    scrollY: camera.scrollY,
                    deltaX: camera.scrollX - oldScrollX,
                    deltaY: camera.scrollY - oldScrollY
                });
            }
        });
    }
    
    /**
     * Set up the layer system for managing overlays and interactions
     */
    protected setupLayerSystem(): void {
        // Create layer manager with coordinate conversion functions
        this.layerManager = new LayerManager(
            this,
            (x: number, y: number) => pixelToHex(x, y),
            (q: number, r: number) => this.world?.getTileAt(q, r) || null,
            (q: number, r: number) => this.world?.getUnitAt(q, r) || null
        );

        // Create base map layer for default interactions
        this.baseMapLayer = new BaseMapLayer(this);

        // Add base map layer to manager
        this.layerManager.addLayer(this.baseMapLayer);

        // Create crossing layer for roads and bridges (depth 5, between tiles and units)
        this.crossingLayer = new CrossingLayer(this, this.tileWidth);
        this.layerManager.addLayer(this.crossingLayer);

        // Create exhausted units highlight layer
        this.exhaustedUnitsLayer = new ExhaustedUnitsHighlightLayer(this, this.tileWidth);
        this.layerManager.addLayer(this.exhaustedUnitsLayer);

        // Create capturing flag layer for units actively capturing buildings
        this.capturingFlagLayer = new CapturingFlagLayer(this, this.tileWidth);
        this.layerManager.addLayer(this.capturingFlagLayer);
    }
    
    private setupInputHandling() {
        // Separate tap detection from drag detection using Phaser's built-in approach
        
        // Track pointer state for tap detection
        let pointerDownPosition: { x: number, y: number } | null = null;
        let pointerDownTime: number = 0;
        const TAP_TIME_THRESHOLD = 300; // ms - max time for a tap
        const TAP_DISTANCE_THRESHOLD = 10; // pixels - max movement for a tap
        
        // Handle pointer down - start tracking for tap vs drag
        this.input.on('pointerdown', (pointer: Phaser.Input.Pointer) => {
            if (pointer.button === 0) { // Left click only
                pointerDownPosition = { x: pointer.x, y: pointer.y };
                pointerDownTime = Date.now();

                // Let layers handle the click (e.g., start drag for reference image in overlay mode)
                if (this.layerManager) {
                    this.layerManager.processClick(pointer);
                }
            }
        });
        
        // Handle pointer up - determine if this was a tap
        this.input.on('pointerup', (pointer: Phaser.Input.Pointer) => {
            if (pointer.button === 0 && pointerDownPosition) { // Left click only
                const timeDelta = Date.now() - pointerDownTime;
                const distance = Math.sqrt(
                    Math.pow(pointer.x - pointerDownPosition.x, 2) +
                    Math.pow(pointer.y - pointerDownPosition.y, 2)
                );

                // This is a tap if it's quick and doesn't move much
                if (timeDelta < TAP_TIME_THRESHOLD && distance < TAP_DISTANCE_THRESHOLD) {
                    this.handleTap(pointer);
                }

                // Notify layers to stop any drag operations
                if (this.layerManager) {
                    this.layerManager.stopDrag();
                }

                // Reset tracking
                pointerDownPosition = null;
                pointerDownTime = 0;
            }
        });
        
        // Set up camera drag using Phaser's built-in drag detection
        this.setupDragZone();
        
        // Handle camera panning via drag zone
        if (this.dragZone) {
            this.dragZone.on('drag', (pointer: Phaser.Input.Pointer, dragX: number, dragY: number) => {
            // Calculate drag delta
            const deltaX = pointer.x - pointer.prevPosition.x;
            const deltaY = pointer.y - pointer.prevPosition.y;

            // Check if any layer wants to handle the drag (e.g., reference image in overlay mode)
            if (this.layerManager) {
                const layerHandledDrag = this.layerManager.processDrag(pointer, deltaX, deltaY);
                if (layerHandledDrag) {
                    // Layer handled the drag, don't pan camera
                    return;
                }
            }

            // Pan camera opposite to drag direction
            const camera = this.cameras.main;
            const oldScrollX = camera.scrollX;
            const oldScrollY = camera.scrollY;

            camera.scrollX -= deltaX / camera.zoom;
            camera.scrollY -= deltaY / camera.zoom;

            // Emit camera moved event if position changed
            if (oldScrollX !== camera.scrollX || oldScrollY !== camera.scrollY) {
                this.events.emit('camera-moved', {
                    scrollX: camera.scrollX,
                    scrollY: camera.scrollY,
                    deltaX: camera.scrollX - oldScrollX,
                    deltaY: camera.scrollY - oldScrollY
                });
            }

            // Keep drag zone centered on camera
            if (this.dragZone) {
                this.dragZone.setPosition(camera.centerX, camera.centerY);
            }
        });
        }
    }
    
    /**
     * Set up drag zone for camera panning
     */
    private setupDragZone(): void {
        // Create an invisible full-screen zone for drag detection
        this.dragZone = this.add.zone(0, 0, this.cameras.main.width * 2, this.cameras.main.height * 2);
        this.dragZone.setOrigin(0.5, 0.5);
        this.dragZone.setInteractive({ draggable: true });
        
        // Position drag zone at camera center and make it follow camera
        this.cameras.main.on('cameramove', () => {
            if (this.dragZone) {
                this.dragZone.setPosition(this.cameras.main.centerX, this.cameras.main.centerY);
            }
        });
    }
    
    /**
     * Update drag zone size when scene is resized
     */
    private updateDragZoneSize(): void {
        if (this.dragZone) {
            this.dragZone.setSize(this.cameras.main.width * 2, this.cameras.main.height * 2);
            this.dragZone.setPosition(this.cameras.main.centerX, this.cameras.main.centerY);
        }
    }
    
    /**
     * Handle tap events (clicks without drag)
     */
    protected handleTap(pointer: Phaser.Input.Pointer): void {
        // Use layer system for hit testing, then send to callback
        if (this.layerManager && this.sceneClickedCallback) {
            const clickContext = this.layerManager.getClickContext(pointer);
            if (clickContext) {
                this.sceneClickedCallback(clickContext, clickContext.layer || 'unknown');
            }
        }
    }
    
    update() {
        // Capture camera state before potential changes
        const camera = this.cameras.main;
        const oldScrollX = camera.scrollX;
        const oldScrollY = camera.scrollY;
        const oldZoom = camera.zoom;
        
        // Handle camera movement with keyboard only if not in input context
        if (!this.isInInputContext()) {
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
        
        // Check if camera position or zoom changed and emit events
        const positionChanged = (oldScrollX !== camera.scrollX || oldScrollY !== camera.scrollY);
        const zoomChanged = (oldZoom !== camera.zoom);
        
        if (positionChanged) {
            this.events.emit('camera-moved', { 
                scrollX: camera.scrollX, 
                scrollY: camera.scrollY,
                deltaX: camera.scrollX - oldScrollX,
                deltaY: camera.scrollY - oldScrollY
            });
        }
        
        if (zoomChanged) {
            this.events.emit('camera-zoomed', { 
                zoom: camera.zoom, 
                oldZoom: oldZoom,
                deltaZoom: camera.zoom - oldZoom
            });
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
        
        // Get texture key from asset provider
        const textureKey = this.assetProvider.getTerrainTexture(terrainType, color);
        
        if (this.textures.exists(textureKey)) {
            const tileSprite = this.add.sprite(position.x, position.y, textureKey);
            tileSprite.setOrigin(0.5, 0.5);
            // Scale sprite to match hex tile size - use provider's display size
            const displaySize = this.assetProvider.getDisplaySize();
            tileSprite.setDisplaySize(displaySize.width, displaySize.height);
            this.tileSprites.set(key, tileSprite); // Use coordinate key, not texture key
        } else {
            // Try fallback without player color
            const fallbackKey = this.assetProvider.getTerrainTexture(terrainType, 0);
            if (this.textures.exists(fallbackKey)) {
                const tileSprite = this.add.sprite(position.x, position.y, fallbackKey);
                tileSprite.setOrigin(0.5, 0.5);
                const displaySize = this.assetProvider.getDisplaySize();
                tileSprite.setDisplaySize(displaySize.width, displaySize.height);
                this.tileSprites.set(key, tileSprite);
            } else {
                console.error(`[PhaserWorldScene] Texture not found: ${textureKey} or ${fallbackKey}`);
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
    public setUnit(unit: Unit, options?: { flash?: boolean, appear?: boolean }): Promise<void> {
        return new Promise((resolve) => {
            const q = unit.q;
            const r = unit.r;
            const unitType = unit.unitType;
            const color = unit.player;
            const key = `${q},${r}`;
            const position = hexToPixel(q, r);

            // Preserve shortcut from World if incoming unit doesn't have one
            if (!unit.shortcut && this.world) {
                const existingUnit = this.world.getUnitAt(q, r);
                if (existingUnit?.shortcut) {
                    (unit as any).shortcut = existingUnit.shortcut;
                }
            }

            // Remove existing unit and labels if they exist
            if (this.unitSprites.has(key)) {
                this.unitSprites.get(key)?.destroy();
            }
            this.removeUnitLabels(key);

            // Get texture key from asset provider
            const textureKey = this.assetProvider.getUnitTexture(unitType, color);

            let unitSprite: Phaser.GameObjects.Sprite | undefined;

            if (this.textures.exists(textureKey)) {
                unitSprite = this.add.sprite(position.x, position.y, textureKey);
                unitSprite.setOrigin(0.5, 0.5);
                unitSprite.setDepth(10); // Units render above tiles
                // Scale sprite to match hex tile size - use provider's display size
                const displaySize = this.assetProvider.getDisplaySize();
                unitSprite.setDisplaySize(displaySize.width * UNIT_TILE_RATIO, displaySize.height * UNIT_TILE_RATIO);
                this.unitSprites.set(key, unitSprite);
            } else {
                // Try fallback without player color
                const fallbackKey = this.assetProvider.getUnitTexture(unitType, 0);
                if (this.textures.exists(fallbackKey)) {
                    unitSprite = this.add.sprite(position.x, position.y, fallbackKey);
                    unitSprite.setOrigin(0.5, 0.5);
                    unitSprite.setDepth(10);
                    const displaySize = this.assetProvider.getDisplaySize();
                    unitSprite.setDisplaySize(displaySize.width * UNIT_TILE_RATIO, displaySize.height * UNIT_TILE_RATIO);
                    this.unitSprites.set(key, unitSprite);
                } else {
                    console.error(`[PhaserWorldScene] Unit texture not found: ${textureKey} or ${fallbackKey}`);
                }
            }

            // Create unit labels
            this.createUnitLabels(unit, position);

            // Apply animations if requested
            if (unitSprite) {
                if (options?.appear && AnimationConfig.APPEAR_DURATION > 0) {
                    // Scale bounce animation: 0.5 → 2 → 1
                    unitSprite.setScale(0.5);
                    this.tweens.add({
                        targets: unitSprite,
                        scale: 1.5,
                        duration: AnimationConfig.APPEAR_DURATION / 2,
                        ease: 'Quad.easeOut',
                        onComplete: () => {
                            this.tweens.add({
                                targets: unitSprite,
                                scale: 1,
                                duration: AnimationConfig.APPEAR_DURATION / 2,
                                ease: 'Quad.easeIn',
                                onComplete: () => resolve()
                            });
                        }
                    });
                } else if (options?.flash && AnimationConfig.FLASH_DURATION > 0) {
                    // Flash animation (tint and back)
                    this.tweens.add({
                        targets: unitSprite,
                        tint: 0xff0000,
                        duration: AnimationConfig.FLASH_DURATION / 2,
                        yoyo: true,
                        ease: 'Cubic.easeInOut',
                        onComplete: () => {
                            unitSprite!.clearTint();
                            resolve();
                        }
                    });
                } else {
                    resolve();
                }
            } else {
                resolve();
            }
        });
    }
    
    public removeUnit(q: number, r: number, options?: { animate?: boolean }): Promise<void> {
        return new Promise((resolve) => {
            const key = `${q},${r}`;
            const sprite = this.unitSprites.get(key);

            if (!sprite) {
                this.removeUnitLabels(key);
                resolve();
                return;
            }

            if (options?.animate && AnimationConfig.FADE_OUT_DURATION > 0) {
                // Fade out animation
                this.tweens.add({
                    targets: sprite,
                    alpha: 0,
                    duration: AnimationConfig.FADE_OUT_DURATION,
                    ease: 'Cubic.easeIn',
                    onComplete: () => {
                        sprite.destroy();
                        this.unitSprites.delete(key);
                        this.removeUnitLabels(key);
                        resolve();
                    }
                });
            } else {
                // Instant removal
                sprite.destroy();
                this.unitSprites.delete(key);
                this.removeUnitLabels(key);
                resolve();
            }
        });
    }

    /**
     * Move a unit along a path with animation.
     * @param unit The unit to move
     * @param path Array of hex coordinates representing the movement path
     */
    public moveUnit(unit: Unit, path: { q: number, r: number }[]): Promise<void> {
        return new Promise((resolve) => {
            if (path.length < 2) {
                // No movement
                resolve();
                return;
            }

            const startKey = `${path[0].q},${path[0].r}`;
            const sprite = this.unitSprites.get(startKey);

            if (!sprite) {
                console.warn(`[PhaserWorldScene] No sprite found for unit at ${path[0].q},${path[0].r}`);
                resolve();
                return;
            }

            // Calculate pixel positions for the path
            const pixelPath = path.map(coord => hexToPixel(coord.q, coord.r));
            const totalDuration = (path.length - 1) * AnimationConfig.MOVE_DURATION_PER_HEX;

            if (totalDuration === 0) {
                // Instant mode - just update position
                const endPos = pixelPath[pixelPath.length - 1];
                sprite.setPosition(endPos.x, endPos.y);

                // Update sprite map key
                this.unitSprites.delete(startKey);
                const endKey = `${path[path.length - 1].q},${path[path.length - 1].r}`;
                this.unitSprites.set(endKey, sprite);

                // Update labels
                this.removeUnitLabels(startKey);
                this.createUnitLabels(unit, endPos);

                resolve();
                return;
            }

            // Chain tweens for smooth movement along path with pauses at each tile
            let currentIndex = 1;

            const animateNextSegment = () => {
                if (currentIndex >= pixelPath.length) {
                    // Animation complete - update sprite map and labels
                    this.unitSprites.delete(startKey);
                    const endKey = `${path[path.length - 1].q},${path[path.length - 1].r}`;
                    this.unitSprites.set(endKey, sprite);

                    this.removeUnitLabels(startKey);
                    const endPos = pixelPath[pixelPath.length - 1];
                    this.createUnitLabels(unit, endPos);

                    resolve();
                    return;
                }

                // Animate movement to next tile
                this.tweens.add({
                    targets: sprite,
                    x: pixelPath[currentIndex].x,
                    y: pixelPath[currentIndex].y,
                    duration: AnimationConfig.MOVE_DURATION_PER_HEX,
                    ease: 'Cubic.easeInOut',
                    onComplete: () => {
                        // Add pause at this tile before moving to next segment
                        if (AnimationConfig.MOVE_PAUSE_PER_HEX > 0) {
                            this.time.delayedCall(AnimationConfig.MOVE_PAUSE_PER_HEX, () => {
                                currentIndex++;
                                animateNextSegment();
                            });
                        } else {
                            // No pause - continue immediately
                            currentIndex++;
                            animateNextSegment();
                        }
                    }
                });
            };

            animateNextSegment();
        });
    }

    /**
     * Show attack effect with projectile and explosion.
     * Handles splash damage by creating multiple simultaneous explosions.
     * @param from Attacker hex coordinates
     * @param to Defender hex coordinates
     * @param damage Damage amount (scales explosion intensity)
     * @param splashTargets Optional array of additional splash damage targets
     */
    public showAttackEffect(
        from: { q: number, r: number },
        to: { q: number, r: number },
        damage: number,
        splashTargets?: { q: number, r: number, damage: number }[]
    ): Promise<void> {
        return new Promise(async (resolve) => {
            const fromPos = hexToPixel(from.q, from.r);
            const toPos = hexToPixel(to.q, to.r);

            // 1. Flash attacker
            const attackerSprite = this.unitSprites.get(`${from.q},${from.r}`);
            if (attackerSprite && AnimationConfig.ATTACK_FLASH_DURATION > 0) {
                await new Promise<void>((flashResolve) => {
                    this.tweens.add({
                        targets: attackerSprite,
                        tint: 0xff6600,
                        duration: AnimationConfig.ATTACK_FLASH_DURATION / 2,
                        yoyo: true,
                        ease: 'Cubic.easeInOut',
                        onComplete: () => {
                            attackerSprite.clearTint();
                            flashResolve();
                        }
                    });
                });
            }

            // 2. Fire projectile
            const projectile = new ProjectileEffect(this, fromPos.x, fromPos.y, toPos.x, toPos.y);
            await projectile.play();

            // 3. Create explosions (main target + splash targets simultaneously)
            const explosionTargets = [
                { x: toPos.x, y: toPos.y, intensity: damage }
            ];

            if (splashTargets) {
                for (const target of splashTargets) {
                    const targetPos = hexToPixel(target.q, target.r);
                    explosionTargets.push({ x: targetPos.x, y: targetPos.y, intensity: target.damage });
                }
            }

            // Play all explosions simultaneously
            await ExplosionEffect.playMultiple(this, explosionTargets);

            resolve();
        });
    }

    /**
     * Show healing effect with rising bubbles.
     * @param q Hex Q coordinate
     * @param r Hex R coordinate
     * @param amount Heal amount (currently not used, but available for scaling)
     */
    public showHealEffect(q: number, r: number, amount: number = 1): Promise<void> {
        const pos = hexToPixel(q, r);
        const healEffect = new HealBubblesEffect(this, pos.x, pos.y, amount);
        return healEffect.play();
    }

    /**
     * Show capture/occupation effect.
     * @param q Hex Q coordinate
     * @param r Hex R coordinate
     */
    public showCaptureEffect(q: number, r: number): Promise<void> {
        const pos = hexToPixel(q, r);
        const captureEffect = new CaptureEffect(this, pos.x, pos.y);
        return captureEffect.play();
    }

    /**
     * Show standalone explosion effect.
     * Utility method for creating explosions without attack context.
     * @param q Hex Q coordinate
     * @param r Hex R coordinate
     * @param intensity Explosion intensity (scales particle count)
     */
    public showExplosion(q: number, r: number, intensity: number = 1): Promise<void> {
        const pos = hexToPixel(q, r);
        const explosion = new ExplosionEffect(this, pos.x, pos.y, intensity);
        return explosion.play();
    }

    /**
     * Create a simple circular particle texture for particle effects.
     * This texture is used by explosion and heal effects.
     */
    private createParticleTexture(): void {
        const graphics = this.add.graphics();
        graphics.fillStyle(0xffffff);
        graphics.fillCircle(8, 8, 8);
        graphics.generateTexture('particle', 16, 16);
        graphics.destroy();
    }

    /**
     * Create health and distance labels for a unit
     */
    private createUnitLabels(unit: Unit, position: { x: number, y: number }): void {
        const key = `${unit.q},${unit.r}`;
        const labels: {
            healthText: Phaser.GameObjects.Text,
            healthBg?: Phaser.GameObjects.Graphics,
            distanceText?: Phaser.GameObjects.Text
        } = {
            healthText: null as any
        };

        const displaySize = this.assetProvider.getDisplaySize();
        const bgColor = 0x3d2817; // Dark brown
        const bgAlpha = 0.7;
        const padding = 3;

        // Create combined label if enabled
        if (this.showUnitHealth) {
            const health = unit.availableHealth || 10;
            const movementPoints = unit.distanceLeft || 0;

            // Format: "Shortcut: MP/Health" (e.g., "B1: 3/10") or "MP/Health" if no shortcut
            const labelText = unit.shortcut
                ? `${unit.shortcut}:${movementPoints}/${health}`
                : `${movementPoints}/${health}`;

            // Position label below the unit, aligned with bottom of tile
            const labelY = position.y + (displaySize.height / 2) - 8; // Just inside bottom edge of tile

            const healthText = this.add.text(position.x, labelY, labelText, {
                fontSize: '10px',
                color: '#ffffff',
                fontFamily: 'Arial',
                fontStyle: 'bold'
            });
            healthText.setOrigin(0.5, 0.5);
            healthText.setDepth(16); // Above background

            // Create background for health label
            const healthBg = this.add.graphics();
            healthBg.fillStyle(bgColor, bgAlpha);
            const healthBounds = healthText.getBounds();
            healthBg.fillRoundedRect(
                healthBounds.x - padding,
                healthBounds.y - padding,
                healthBounds.width + padding * 2,
                healthBounds.height + padding * 2,
                4
            );
            healthBg.setDepth(15); // Below text

            labels.healthText = healthText;
            labels.healthBg = healthBg;
        }

        // Create distance label if enabled and there's a selected unit
        if (this.showUnitDistance && this.selectedUnitCoord) {
            const distance = this.calculateHexDistance(
                unit.q, unit.r,
                this.selectedUnitCoord.q, this.selectedUnitCoord.r
            );

            // Position distance label below the unit, above the health label
            const labelY = position.y + (displaySize.height / 2) - 20; // Above health label

            const distanceText = this.add.text(position.x, labelY, distance.toString(), {
                fontSize: '9px',
                color: '#00aaff',
                fontFamily: 'Arial',
                fontStyle: 'bold'
            });
            distanceText.setOrigin(0.5, 0.5);
            distanceText.setDepth(16); // Above units
            labels.distanceText = distanceText;
        }

        this.unitLabels.set(key, labels);
    }
    
    /**
     * Remove unit labels
     */
    private removeUnitLabels(key: string): void {
        const labels = this.unitLabels.get(key);
        if (labels) {
            if (labels.healthText) {
                labels.healthText.destroy();
            }
            if (labels.healthBg) {
                labels.healthBg.destroy();
            }
            if (labels.distanceText) {
                labels.distanceText.destroy();
            }
            this.unitLabels.delete(key);
        }
    }
    
    /**
     * Update existing unit labels with fresh unit data
     */
    private updateUnitLabels(unit: Unit): void {
        const key = `${unit.q},${unit.r}`;
        const labels = this.unitLabels.get(key);

        if (labels && labels.healthText && this.showUnitHealth) {
            const health = unit.availableHealth || 10;
            const movementPoints = unit.distanceLeft || 0;

            // Format: "Shortcut: MP/Health" (e.g., "B1: 3/10") or "MP/Health" if no shortcut
            const labelText = unit.shortcut
                ? `${unit.shortcut}: ${movementPoints}/${health}`
                : `${movementPoints}/${health}`;

            labels.healthText.setText(labelText);

            // Update background size to match new text
            if (labels.healthBg) {
                const bgColor = 0x3d2817;
                const bgAlpha = 0.7;
                const padding = 3;

                labels.healthBg.clear();
                labels.healthBg.fillStyle(bgColor, bgAlpha);
                const healthBounds = labels.healthText.getBounds();
                labels.healthBg.fillRoundedRect(
                    healthBounds.x - padding,
                    healthBounds.y - padding,
                    healthBounds.width + padding * 2,
                    healthBounds.height + padding * 2,
                    4
                );
            }
        }
    }
    
    /**
     * Public method to refresh all unit labels with current World data
     */
    public refreshUnitLabels(world: any): void {
        const allUnits = world.getAllUnits();
        
        for (const unit of allUnits) {
            this.updateUnitLabels(unit);
        }
    }
    
    /**
     * Calculate hex distance between two coordinates
     */
    private calculateHexDistance(q1: number, r1: number, q2: number, r2: number): number {
        // Convert axial coordinates to cube coordinates for distance calculation
        const s1 = -q1 - r1;
        const s2 = -q2 - r2;
        
        return Math.max(Math.abs(q1 - q2), Math.abs(r1 - r2), Math.abs(s1 - s2));
    }
    
    public clearAllUnits() {
        this.unitSprites.forEach(unit => unit.destroy());
        this.unitSprites.clear();

        // Clear all unit labels
        this.unitLabels.forEach(labels => {
            if (labels.healthText) labels.healthText.destroy();
            if (labels.healthBg) labels.healthBg.destroy();
            if (labels.distanceText) labels.distanceText.destroy();
        });
        this.unitLabels.clear();
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
    
    /**
     * Set unit health label visibility
     */
    public setShowUnitHealth(show: boolean) {
        this.showUnitHealth = show;
        this.refreshAllUnitLabels();
    }
    
    /**
     * Set unit distance label visibility and selected unit
     */
    public setShowUnitDistance(show: boolean, selectedUnitCoord?: { q: number, r: number }) {
        this.showUnitDistance = show;
        this.selectedUnitCoord = selectedUnitCoord || null;
        this.refreshAllUnitLabels();
    }
    
    /**
     * Refresh all unit labels (useful when settings change)
     */
    private refreshAllUnitLabels(): void {
        if (!this.world) return;
        
        // Recreate labels for all units
        for (const unit of this.world.getAllUnits()) {
            const key = `${unit.q},${unit.r}`;
            const position = hexToPixel(unit.q, unit.r);
            
            // Remove existing labels
            this.removeUnitLabels(key);
            
            // Create new labels with current settings
            this.createUnitLabels(unit, position);
        }
    }
    
    private updateTheme() {
        if (!this.cameras?.main) return ;
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
            { x: position.x - halfWidth * 0.866, y: position.y - halfHeight * 0.5 }    // Top-left
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

    /**
     * Load world data into the scene
     */
    public async loadWorld(world: World): Promise<void> {
        if (!this.isInitialized) {
            throw new Error('[PhaserWorldScene] Scene not initialized. Call initialize() first.');
        }

        // Set world as source of truth
        this.setWorld(world);

        // Wait for assets to be ready before placing tiles/units
        await this.waitForAssetsReady();
        
        // Clear existing content
        this.clearAllTiles();
        this.clearAllUnits();
        
        // Load tiles from World
        const tiles = world.getAllTiles();
        if (tiles.length > 0) {
            tiles.forEach(tile => {
                this.setTile(tile);
            });
        }
        
        // Load units from World
        const units = world.getAllUnits();
        if (units.length > 0) {
            units.forEach(unit => {
                this.setUnit(unit);
            });
        }
        
        // Center camera on the loaded world
        this.centerCameraOnWorld();
    }

    /**
     * Center the camera on the loaded world by calculating bounds and focusing on center
     */
    public centerCameraOnWorld(): void {
        if (!this.world) {
            return;
        }
        
        const allTiles = this.world.getAllTiles();
        if (allTiles.length === 0) {
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
     * Resize the scene - should be called explicitly by parent page when needed
     * No longer uses automatic ResizeObserver to avoid circular dependencies
     */
    public resize(width?: number, height?: number): void {
        if (this.phaserGame && this.containerElement) {
            const w = width || this.containerElement.clientWidth;
            const h = height || this.containerElement.clientHeight;
            
            this.phaserGame.scale.resize(w, h);
            
            // Update drag zone size to match new camera dimensions
            this.updateDragZoneSize();
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
        
        return this.assetsReadyPromise;
    }
    
    /**
     * Get access to the layer manager for external layer management
     */
    public getLayerManager(): LayerManager | null {
        return this.layerManager;
    }

    /**
     * Get the exhausted units highlight layer for marking units/tiles with no moves
     */
    public getExhaustedUnitsLayer(): ExhaustedUnitsHighlightLayer | null {
        return this.exhaustedUnitsLayer;
    }

    /**
     * Get the capturing flag layer for showing active capture indicators
     */
    public getCapturingFlagLayer(): CapturingFlagLayer | null {
        return this.capturingFlagLayer;
    }

    /**
     * Calculate the tight bounding box of all tiles in the world
     * @param padding - Padding in pixels to add around the bounds
     * @returns Bounding box {x, y, width, height} or null if no tiles
     */
    private calculateMapBounds(padding: number = 20): { x: number; y: number; width: number; height: number } | null {
        if (!this.world) {
            return null;
        }

        const allTiles = this.world.getAllTiles();
        if (allTiles.length === 0) {
            return null;
        }

        // Calculate hex coordinate bounds
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

        // Convert to pixel coordinates
        const topLeft = hexToPixel(minQ, minR);
        const bottomRight = hexToPixel(maxQ, maxR);

        // Calculate bounds with tile dimensions
        const halfWidth = this.tileWidth / 2;
        const halfHeight = this.tileHeight / 2;

        const x = topLeft.x - halfWidth - padding;
        const y = topLeft.y - halfHeight - padding;
        const width = (bottomRight.x - topLeft.x) + this.tileWidth + (padding * 2);
        const height = (bottomRight.y - topLeft.y) + this.tileHeight + (padding * 2);

        return { x, y, width, height };
    }

    /**
     * Capture a screenshot of the current scene (clipped to map bounds)
     * @param callback - Function to receive the screenshot as a Blob
     * @param type - Image type (image/png or image/jpeg)
     * @param quality - Quality for JPEG (0-1)
     */
    public captureScreenshot(callback: (blob: Blob | null) => void, type: string = 'image/png', quality: number = 0.92): void {
        if (!this.phaserGame || !this.isInitialized) {
            console.error('[PhaserWorldScene] Cannot capture screenshot: scene not initialized');
            callback(null);
            return;
        }

        // Calculate map bounds for clipping
        const bounds = this.calculateMapBounds(20);

        if (!bounds) {
            console.error('[PhaserWorldScene] No tiles to capture');
            callback(null);
            return;
        }

        // Convert world coordinates to screen coordinates
        const camera = this.cameras.main;
        const screenX = (bounds.x - camera.scrollX) * camera.zoom;
        const screenY = (bounds.y - camera.scrollY) * camera.zoom;
        const screenWidth = bounds.width * camera.zoom;
        const screenHeight = bounds.height * camera.zoom;

        // Clamp to canvas bounds
        const canvas = this.game.canvas;
        const clampedX = Math.max(0, Math.min(screenX, canvas.width));
        const clampedY = Math.max(0, Math.min(screenY, canvas.height));
        const clampedWidth = Math.min(screenWidth, canvas.width - clampedX);
        const clampedHeight = Math.min(screenHeight, canvas.height - clampedY);

        // Use Phaser's snapshotArea for clipped capture
        this.game.renderer.snapshotArea(
            clampedX,
            clampedY,
            clampedWidth,
            clampedHeight,
            (image: HTMLImageElement | Phaser.Display.Color) => {
                if (image instanceof HTMLImageElement) {
                    // Convert image to blob
                    const canvas = document.createElement('canvas');
                    canvas.width = image.width;
                    canvas.height = image.height;
                    const ctx = canvas.getContext('2d');

                    if (ctx) {
                        ctx.drawImage(image, 0, 0);
                        canvas.toBlob((blob) => {
                            callback(blob);
                        }, type, quality);
                    } else {
                        console.error('[PhaserWorldScene] Failed to get canvas context');
                        callback(null);
                    }
                } else {
                    console.error('[PhaserWorldScene] Unexpected snapshot result');
                    callback(null);
                }
            }
        );
    }

    /**
     * Capture a screenshot and return it as a Promise
     * @param type - Image type (image/png or image/jpeg)
     * @param quality - Quality for JPEG (0-1)
     * @returns Promise resolving to screenshot Blob
     */
    public async captureScreenshotAsync(type: string = 'image/png', quality: number = 0.92): Promise<Blob | null> {
        return new Promise((resolve) => {
            this.captureScreenshot(resolve, type, quality);
        });
    }
}
