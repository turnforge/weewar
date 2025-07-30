import { BasePage } from '../lib/BasePage';
import { EventBus } from '../lib/EventBus';
import { WorldViewer } from './WorldViewer';
import { WorldStatsPanel } from './WorldStatsPanel';
import { World } from './World';
import { LCMComponent } from '../lib/LCMComponent';
import { LifecycleController } from '../lib/LifecycleController';
import { WorldEventTypes } from './events';

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
class WorldDetailsPage extends BasePage implements LCMComponent {
    private currentWorldId: string | null;
    private isLoadingWorld: boolean = false;
    private world: World;
    
    // Component instances
    private worldViewer: WorldViewer;
    private worldStatsPanel: WorldStatsPanel;

    // =============================================================================
    // LCMComponent Interface Implementation
    // =============================================================================

    /**
     * Phase 1: Initialize DOM and discover child components
     */
    performLocalInit(): LCMComponent[] {
        console.log('WorldDetailsPage: performLocalInit() - Phase 1');
        
        // 1. FIRST: Load World from DOM elements (canonical source of truth)
        const worldMetadataElement = document.getElementById('world-data-json');
        const worldTilesElement = document.getElementById('world-tiles-data-json');
        this.world = new World(this.eventBus).loadFromElement(worldMetadataElement!, worldTilesElement!);
        console.log('WorldDetailsPage: World object loaded from DOM elements');
        
        // 2. THEN: Subscribe to events BEFORE creating components
        this.subscribeToWorldViewerEvents();
        
        // 3. FINALLY: Create child components
        this.createComponents();
        
        console.log('WorldDetailsPage: DOM initialized, returning child components');
        
        // Return child components for lifecycle management
        const childComponents: LCMComponent[] = [];
        childComponents.push(this.worldViewer!); // Should exist - fail if not
        childComponents.push(this.worldStatsPanel as any); // Should exist - fail if not
        return childComponents;
    }

    /**
     * Phase 2: Inject dependencies (none needed for WorldDetailsPage)
     */
    setupDependencies(): void {
        console.log('WorldDetailsPage: setupDependencies() - Phase 2');
        // WorldDetailsPage doesn't need external dependencies
    }

    /**
     * Phase 3: Activate component when all dependencies are ready
     */
    async activate(): Promise<void> {
        console.log('WorldDetailsPage: activate() - Phase 3');
        
        // Bind events now that all components are ready
        this.bindPageSpecificEvents();
        
        console.log('WorldDetailsPage: activation complete');
    }

    /**
     * Cleanup phase (called by lifecycle controller if needed)
     */
    deactivate(): void {
        console.log('WorldDetailsPage: deactivate() - cleanup');
        this.destroy();
    }
    
    /**
     * Subscribe to WorldViewer events before component creation
     */
    private subscribeToWorldViewerEvents(): void {
        // Subscribe to WorldViewer ready event BEFORE creating the component
        console.log('WorldDetailsPage: Subscribing to WORLD_VIEWER_READY event');
        this.eventBus.subscribe(WorldEventTypes.WORLD_VIEWER_READY, null, () => {
            console.log('WorldDetailsPage: WorldViewer is ready, passing World object...');
            // Give Phaser time to fully initialize webgl context and scene
                // Pass the canonical World object directly
                this.worldViewer!.loadWorld(this.world!);
                this.showToast('Success', 'World loaded successfully', 'success');
            // setTimeout(async () => { }, 10);
        });
    }

    /**
     * Create WorldViewer and WorldStatsPanel component instances
     */
    private createComponents(): void {
        // Create WorldViewer component
        const worldViewerRoot = this.ensureElement('[data-component="world-viewer"]', 'world-viewer-root');
        console.log('WorldDetailsPage: Creating WorldViewer with eventBus:', this.eventBus);
        this.worldViewer = new WorldViewer(worldViewerRoot, this.eventBus, true);
        
        // Create WorldStatsPanel component - pass the content div, not the container with header
        const worldStatsContainer = this.ensureElement('[data-component="world-stats-panel"]', 'world-stats-root');
        const worldStatsContent = worldStatsContainer.querySelector('.p-4.space-y-4') as HTMLElement;
        if (!worldStatsContent) {
            throw new Error('WorldDetailsPage: WorldStatsPanel content div not found');
        }
        this.worldStatsPanel = new WorldStatsPanel(worldStatsContent, this.eventBus, true);
        
        console.log('WorldDetailsPage: Components created');
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

    public destroy(): void {
        // Clean up components
        if (this.worldViewer) {
            this.worldViewer.destroy();
            // Force set as null.  We are pushing for fail fast on values
            // that should NOT be null
            this.worldViewer = null as any;
        }
        
        if (this.worldStatsPanel) {
            this.worldStatsPanel.destroy();
            this.worldStatsPanel = null as any;
        }
        
        // Clean up world data
        this.world = null as any
        this.currentWorldId = null;
    }
}

// Initialize page when DOM is ready using LifecycleController
document.addEventListener('DOMContentLoaded', async () => {
    console.log('DOM loaded, starting WorldDetailsPage initialization...');

    // Create page instance (just basic setup)
    const page = new WorldDetailsPage("WorldDetailsPage");
    
    // Create lifecycle controller with debug logging
    const lifecycleController = new LifecycleController(page.eventBus, LifecycleController.DefaultConfig)
    
    // Start breadth-first initialization
    await lifecycleController.initializeFromRoot(page);
    
    console.log('WorldDetailsPage fully initialized via LifecycleController');
});
