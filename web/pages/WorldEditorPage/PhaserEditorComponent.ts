import { BaseComponent } from '../../lib/Component';
import { LCMComponent } from '../../lib/LCMComponent';
import { EventBus } from '../../lib/EventBus';
import { EditorEventTypes } from '../common/events';
import { PhaserEditorScene } from './PhaserEditorScene';
import { IWorldEditorPresenter } from './WorldEditorPresenter';
import { Unit, Tile, World } from '../common/World';

/**
 * PhaserEditorComponent - Manages the Phaser.js-based world editor interface using BaseComponent architecture
 * 
 * Responsibilities:
 * - Initialize and manage Phaser.js world editor lifecycle
 * - Handle editor-specific DOM container setup
 * - Emit tile click events to EventBus
 * - Listen for tool changes (terrain, unit, brush size) from EditorToolsPanel
 * - Manage world rendering, camera controls, and visual settings
 * - Handle world data loading and saving operations
 * - Manage reference image features for overlay/background
 * 
 * Does NOT handle:
 * - Tool selection UI (handled by EditorToolsPanel)
 * - Layout management (handled by parent dockview)
 * - Save/load UI (will be handled by SaveLoadComponent)
 * - Direct DOM manipulation outside of phaser-container
 */
export class PhaserEditorComponent extends BaseComponent implements LCMComponent {
    public editorScene: PhaserEditorScene;
    private isInitialized: boolean = false;

    // Dependencies (injected in phase 2)
    private presenter: IWorldEditorPresenter | null = null;
    private world: World;
    
    // =============================================================================
    // LCMComponent Interface Implementation
    // =============================================================================

    /**
     * Phase 1: Initialize DOM and discover child components
     */
    public performLocalInit(): LCMComponent[] {
        this.log('PhaserEditorComponent: performLocalInit() - Phase 1');
        
        // Subscribe to EventBus events now that dependencies are available
        this.subscribeToEvents();
        
        // Set up Phaser container within our root element
        this.setupPhaserContainer();
        
        // Bind toolbar event handlers
        this.bindToolbarEvents();
        
        this.log('PhaserEditorComponent: DOM setup complete');
        
        // This is a leaf component - no children
        return [];
    }

    /**
     * Phase 2: Inject dependencies
     */
    public setupDependencies(): void {
        this.log('PhaserEditorComponent: setupDependencies() - Phase 2');

        // Dependencies should be set by parent using setters
        // This phase validates that required dependencies are available
        if (!this.presenter) {
            throw new Error('PhaserEditorComponent requires presenter - use setPresenter()');
        }

        this.log('PhaserEditorComponent: Dependencies validation complete');
    }

    /**
     * Phase 3: Activate component when all dependencies are ready
     */
    public async activate(): Promise<void> {
        this.log('PhaserEditorComponent: activate() - Phase 3');
        
        // Initialize Phaser editor now that dependencies are ready
        await this.initializePhaserEditor();
        
        this.log('PhaserEditorComponent: activation complete');
    }

    // Explicit dependency setters
    public setPresenter(presenter: IWorldEditorPresenter): void {
        this.presenter = presenter;
        this.log('Presenter dependency set via explicit setter');
    }

    public setWorld(world: World): void {
        this.world = world;
        this.log('World dependency set via explicit setter');
    }

    // Explicit dependency getters
    public getPresenter(): IWorldEditorPresenter | null {
        return this.presenter;
    }

    public getWorld(): World | null {
        return this.world;
    }

    /**
     * Subscribe to all EventBus events (called in activate phase)
     */
    private subscribeToEvents(): void {
        // World events and visual state are handled by PhaserWorldScene directly
        // Reference image events now go through presenter
        // No EventBus subscriptions needed
        this.log('No EventBus subscriptions needed - presenter handles all events');
    }
    
    /**
     * Bind toolbar event handlers
     */
    private bindToolbarEvents(): void {
        // Bind clear tile button
        const clearTileBtn = this.findElement('#clear-tile-btn');
        if (clearTileBtn) {
            clearTileBtn.addEventListener('click', () => {
                this.activateClearMode();
            });
            this.log('Clear tile button bound');
        }
    }
    
    /**
     * Activate clear mode
     */
    private activateClearMode(): void {
        this.presenter!.setPlacementMode('clear');
        this.log('Clear mode activated via toolbar button');
    }
    
    /**
     * Set up the Phaser container element
     */
    private setupPhaserContainer(): void {
        // PhaserSceneView template creates the container with id="phaser-container"
        // All sizing constraints are handled by the template
        const phaserContainer = this.findElement('#phaser-container');

        if (!phaserContainer) {
            throw new Error('Phaser container #phaser-container not found - ensure PhaserSceneView template is included');
        }

        this.log('Phaser container found, sizing handled by PhaserSceneView template');
    }

    /**
     * Wait for container to become visible before initializing Phaser
     */
    private async waitForContainerVisible(containerElement: HTMLElement): Promise<void> {
        return new Promise<void>((resolve) => {
            const checkVisibility = () => {
                const rect = containerElement.getBoundingClientRect();
                
                if (true || (rect.width > 0 && rect.height > 0)) {
                    this.log('Container is visible, ready for Phaser initialization');
                    resolve();
                } else {
                    // Check again after a short delay
                    setTimeout(checkVisibility, 50);
                }
            };
            
            // Start checking
            setTimeout(checkVisibility, 50);
        });
    }
    
    /**
     * Initialize the Phaser editor (called in activate phase)
     */
    private async initializePhaserEditor(): Promise<void> {
        this.log('Initializing Phaser editor...');
        
        // Find the container element that we just set up
        const containerElement = this.findElement('#phaser-container');
        if (!containerElement) {
            throw new Error('Phaser container element not found after setup');
        }
        
        // Wait for container to have dimensions before initializing Phaser
        await this.waitForContainerVisible(containerElement);
        
        // Create Phaser editor scene instance directly with the element
        this.editorScene = new PhaserEditorScene(containerElement, this.eventBus, this.debugMode);
        
        // Initialize the scene using LCMComponent lifecycle
        await this.editorScene.performLocalInit();
        this.editorScene.setupDependencies();
        await this.editorScene.activate();
        
        // Wait for assets to be ready before considering the scene ready
        this.log('Waiting for assets to be ready...');
        await this.editorScene.waitForAssetsReady();
        this.log('Assets are ready');
        
        // Set up event handlers
        await this.setupPhaserEventHandlers();
        
        // Apply current theme
        const isDarkMode = document.documentElement.classList.contains('dark');
        this.editorScene.setTheme(isDarkMode);
        
        this.isInitialized = true;
        this.log('Phaser editor initialized successfully');
        
        // Emit ready event for other components - now assets are actually ready
        this.emit(EditorEventTypes.PHASER_READY, {}, this, this);
    }
    
    /**
     * Set up event handlers for Phaser editor
     */
    private async setupPhaserEventHandlers(): Promise<void> {
        if (!this.editorScene) return;
        
        // Scene should be ready after activate() call
        // Set up unified scene click callback for editor functionality
        this.editorScene.sceneClickedCallback = (context: any, layer: string, extra?: any) => {
            const { hexQ: q, hexR: r, tile, unit } = context;
            this.log(`Scene clicked at Q=${q}, R=${r} on layer '${layer}'`, { tile, unit });
            
            // Handle painting based on current mode (works for both tile and unit clicks)
            this.handleTileClick(q, r);
        };
        
        // Handle world changes
        this.editorScene.onWorldChange(() => {
            this.log('World changed in Phaser');
            this.emit(EditorEventTypes.WORLD_CHANGED, {}, this, this);
        });

        // Handle reference scale changes - notify presenter to update panel
        this.editorScene.onReferenceScaleChange((x: number, y: number) => {
            this.presenter?.onReferenceScaleUpdatedFromScene(x, y);
        });

        // Handle reference position changes - notify presenter to update panel
        this.editorScene.onReferencePositionChange((x: number, y: number) => {
            this.presenter?.onReferencePositionUpdatedFromScene(x, y);
        });

        this.log('Phaser event handlers setup complete');
    }
    
    
    /**
     * Handle tile clicks for painting - delegates to presenter
     */
    private handleTileClick(q: number, r: number): void {
        if (!this.editorScene || !this.isInitialized) {
            return;
        }
        this.presenter!.handleTileClick(q, r);
    }
    
    /**
     * Check if Phaser editor is initialized
     */
    public getIsInitialized(): boolean {
        return this.isInitialized;
    }
    
    /**
     * Set theme for the editor
     */
    public setTheme(isDark: boolean): void {
        if (this.editorScene && this.isInitialized) {
            this.editorScene.setTheme(isDark);
            this.log(`Theme set to: ${isDark ? 'dark' : 'light'}`);
        }
    }
    
    /**
     * Set grid visibility
     */
    public setShowGrid(show: boolean): void {
        if (this.editorScene && this.isInitialized) {
            this.editorScene.setShowGrid(show);
            this.log(`Grid visibility set to: ${show}`);
        }
    }
    
    /**
     * Set coordinate visibility
     */
    public setShowCoordinates(show: boolean): void {
        if (this.editorScene && this.isInitialized) {
            this.editorScene.setShowCoordinates(show);
            this.log(`Coordinate visibility set to: ${show}`);
        }
    }
    
    /**
     * Get viewport center for world generation
     */
    public getViewportCenter(): { q: number; r: number } {
        if (this.editorScene && this.isInitialized) {
            return this.editorScene.getViewportCenter();
        }
        return { q: 0, r: 0 };
    }
    
    /**
     * Center camera on specific coordinates
     */
    public centerCamera(q: number = 0, r: number = 0): void {
        if (this.editorScene && this.isInitialized) {
            this.editorScene.centerCamera(q, r);
            this.log(`Camera centered on Q=${q}, R=${r}`);
        }
    }
    
    /**
     * World generation methods
     */
    public fillAllTerrain(terrain: number, color: number = 0): void {
        if (this.editorScene && this.isInitialized) {
            this.editorScene.fillAllTerrain(terrain, color);
            this.log(`Filled all terrain with type ${terrain}`);
        }
    }
    
    public randomizeTerrain(): void {
        if (this.editorScene && this.isInitialized) {
            this.editorScene.randomizeTerrain();
            this.log('Terrain randomized');
        }
    }
    
    public createIslandPattern(centerQ: number, centerR: number, radius: number): void {
        if (this.editorScene && this.isInitialized) {
            this.editorScene.createIslandPattern(centerQ, centerR, radius);
            this.log(`Created island pattern at Q=${centerQ}, R=${centerR} with radius ${radius}`);
        }
    }
    
    /**
     * Reference image methods (display controls only, loading handled by ReferenceImagePanel)
     */
    public setReferenceMode(mode: number): void {
        if (this.editorScene && this.isInitialized) {
            this.editorScene.setReferenceMode(mode);
            const modeNames = ['hidden', 'background', 'overlay'];
            this.log(`Reference mode set to: ${modeNames[mode] || mode}`);
        }
    }
    
    public setReferenceAlpha(alpha: number): void {
        if (this.editorScene && this.isInitialized) {
            this.editorScene.setReferenceAlpha(alpha);
            this.log(`Reference alpha set to: ${alpha}`);
        }
    }
    
    public setReferencePosition(x: number, y: number): void {
        if (this.editorScene && this.isInitialized) {
            this.editorScene.setReferencePosition(x, y);
            this.log(`Reference position set to: (${x}, ${y})`);
        }
    }
    
    public setReferenceScale(x: number, y: number): void {
        if (this.editorScene && this.isInitialized) {
            this.editorScene.setReferenceScale(x, y);
            this.log(`Reference scale set to: (${x}, ${y})`);
        }
    }
    
    public setReferenceScaleFromTopLeft(x: number, y: number): void {
        if (this.editorScene && this.isInitialized) {
            this.editorScene.setReferenceScaleFromTopLeft(x, y);
            this.log(`Reference scale set from top-left to: (${x}, ${y})`);
        }
    }
    
    public getReferenceState(): {
        mode: number;
        alpha: number;
        position: { x: number; y: number };
        scale: { x: number; y: number };
        hasImage: boolean;
    } | null {
        if (this.editorScene && this.isInitialized) {
            return this.editorScene.getReferenceState();
        }
        return null;
    }
    
    public clearReferenceImage(): void {
        if (this.editorScene && this.isInitialized) {
            this.editorScene.clearReferenceImage();
            this.log('Reference image cleared');
        }
    }
    
    /**
     * Register reference scale change callback
     */
    public onReferenceScaleChange(callback: (x: number, y: number) => void): void {
        if (this.editorScene && this.isInitialized) {
            this.editorScene.onReferenceScaleChange(callback);
        }
    }
}
