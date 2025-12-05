import { BasePage, EventBus, LCMComponent, LifecycleController } from '@panyam/tsappkit';
import { WorldEventTypes } from './common/events';
import { PhaserWorldScene } from './common/PhaserWorldScene';
import { WorldStatsPanel } from './common/WorldStatsPanel';
import { World } from './common/World';

/**
 * World Details Page - Orchestrator for world viewing functionality
 * Responsible for:
 * - Data loading and coordination
 * - Component initialization and management
 * - Page-level event coordination
 * - Navigation and user actions
 * 
 * Does NOT handle:
 * - Direct DOM manipulation (delegated to components)
 * - Phaser management (delegated to WorldViewer)
 * - Statistics display (delegated to WorldStatsPanel)
 */
class WorldViewerPage extends BasePage implements LCMComponent {
    private currentWorldId: string | null;
    private isLoadingWorld: boolean = false;
    private world: World;
    
    // Component instances
    private worldScene: PhaserWorldScene;
    private worldStatsPanel: WorldStatsPanel;
    private worldStatsPanelMobile: WorldStatsPanel | null = null;

    // =============================================================================
    // LCMComponent Interface Implementation
    // =============================================================================

    /**
     * Phase 1: Initialize DOM and discover child components
     */
    performLocalInit(): LCMComponent[] {
        // 1. FIRST: Load World from DOM elements (canonical source of truth)
        const worldMetadataElement = document.getElementById('world-data-json');
        const worldTilesElement = document.getElementById('world-tiles-data-json');
        this.world = new World(this.eventBus).loadFromElement(worldMetadataElement!, worldTilesElement!);
        
        // 2. THEN: Subscribe to events BEFORE creating components - None yet
        
        // 3. FINALLY: Create child components
        this.createComponents();
        
        // Return child components for lifecycle management
        const childComponents: LCMComponent[] = [];
        childComponents.push(this.worldScene!); // Should exist - fail if not
        childComponents.push(this.worldStatsPanel as any); // Should exist - fail if not
        if (this.worldStatsPanelMobile) {
            childComponents.push(this.worldStatsPanelMobile as any);
        }
        return childComponents;
    }

    /**
     * Phase 2: Inject dependencies
     */
    setupDependencies(): void {
        // Set up scene click callback now that worldScene is initialized
        this.worldScene.sceneClickedCallback = () => {
          console.log("here...")
        }
    }

    /**
     * Phase 3: Activate component when all dependencies are ready
     */
    async activate(): Promise<void> {
        // Bind events now that all components are ready
        this.bindPageSpecificEvents();
        this.worldScene.loadWorld(this.world);
        this.showToast('Success', 'World loaded successfully', 'success');

        // Dismiss splash screen once world is loaded
        super.dismissSplashScreen();
    }

    /**
     * Cleanup phase (called by lifecycle controller if needed)
     */
    deactivate(): void {
        // Remove event subscriptions
        this.removeSubscription(WorldEventTypes.WORLD_VIEWER_READY, null);
        
        // Clean up components
        if (this.worldScene) {
            this.worldScene.deactivate();
            // Force set as null.  We are pushing for fail fast on values
            // that should NOT be null
            this.worldScene = null as any;
        }
        
        if (this.worldStatsPanel) {
            this.worldStatsPanel.deactivate();
            this.worldStatsPanel = null as any;
        }

        if (this.worldStatsPanelMobile) {
            this.worldStatsPanelMobile.deactivate();
            this.worldStatsPanelMobile = null;
        }

        // Clean up world data
        this.world = null as any
        this.currentWorldId = null;
    }

    /**
     * Create PhaserWorldScene and WorldStatsPanel component instances
     */
    private createComponents(): void {
        // Create PhaserWorldScene component - uses PhaserSceneView template with SceneId: "world-viewer-scene"
        const phaserContainer = this.ensureElement('#world-viewer-scene', 'world-viewer-scene');
        this.worldScene = new PhaserWorldScene(phaserContainer, this.eventBus, true);
        this.worldScene.setWorld(this.world);

        // Create WorldStatsPanel component for desktop sidebar
        const worldStatsContainer = this.ensureElement('[data-component="world-stats-panel"]', 'world-stats-root');
        this.worldStatsPanel = new WorldStatsPanel(worldStatsContainer, this.eventBus, true);
        this.worldStatsPanel.setWorld(this.world);

        // Create WorldStatsPanel component for mobile bottom sheet (if container exists)
        const mobileStatsContainer = document.querySelector('[data-component="world-stats-panel-mobile"]') as HTMLElement;
        if (mobileStatsContainer) {
            this.worldStatsPanelMobile = new WorldStatsPanel(mobileStatsContainer, this.eventBus, true);
            this.worldStatsPanelMobile.setWorld(this.world);
        }
    }

    /**
     * Internal method to bind page-specific events (called from activate() phase)
     */
    private bindPageSpecificEvents(): void {
        const mobileMenuButton = document.getElementById('mobile-menu-button');
        if (mobileMenuButton) {
            mobileMenuButton.addEventListener('click', () => {
              // Do things like sidebar drawers etc
            });
        }

        // Bind copy world button if it exists
        const copyButton = document.querySelector('[data-action="copy-world"]');
        if (copyButton) {
            copyButton.addEventListener('click', this.copyWorld.bind(this));
        }

        // Bind bottom sheet controls (mobile only)
        this.initializeBottomSheet();
    }

    /**
     * Initialize bottom sheet for mobile stats panel
     */
    private initializeBottomSheet(): void {
        const fab = document.getElementById('stats-fab');
        const overlay = document.getElementById('stats-overlay');
        const panel = document.getElementById('stats-panel');
        const backdrop = document.getElementById('stats-backdrop');
        const closeButton = document.getElementById('stats-close');

        if (!fab || !overlay || !panel || !backdrop || !closeButton) {
            return; // Elements don't exist (probably on desktop)
        }

        // Open bottom sheet
        const openSheet = () => {
            overlay.classList.remove('hidden');
            // Force reflow to enable transition
            overlay.offsetHeight;
            panel.classList.remove('translate-y-full');
        };

        // Close bottom sheet
        const closeSheet = () => {
            panel.classList.add('translate-y-full');
            // Wait for animation to complete before hiding
            setTimeout(() => {
                overlay.classList.add('hidden');
            }, 300);
        };

        // Event listeners
        fab.addEventListener('click', openSheet);
        closeButton.addEventListener('click', closeSheet);
        backdrop.addEventListener('click', closeSheet);

        // Close on escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && !overlay.classList.contains('hidden')) {
                closeSheet();
            }
        });
    }


    /** Copy world functionality */
    private copyWorld(): void {
        if (!this.currentWorldId) {
            this.showToast('Error', 'No world ID available for copying', 'error');
            return;
        }
        
        // Navigate to editor page with copy mode
        const copyUrl = `/worlds/new?copy=${this.currentWorldId}`;
        window.location.href = copyUrl;
    }
}

// Initialize page when DOM is ready using LifecycleController
document.addEventListener('DOMContentLoaded', async () => {
    // Create page instance (just basic setup)
    const page = new WorldViewerPage("WorldViewerPage");
    
    // Create lifecycle controller with debug logging
    const lifecycleController = new LifecycleController(page.eventBus, LifecycleController.DefaultConfig)
    
    // Start breadth-first initialization
    await lifecycleController.initializeFromRoot(page);
});
