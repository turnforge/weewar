import { BasePage } from '../lib/BasePage';
import { EventBus } from '../lib/EventBus';
import { PhaserWorldScene } from './phaser/PhaserWorldScene';
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
class WorldViewerPage extends BasePage implements LCMComponent {
    private currentWorldId: string | null;
    private isLoadingWorld: boolean = false;
    private world: World;
    
    // Component instances
    private worldScene: PhaserWorldScene;
    private worldStatsPanel: WorldStatsPanel;

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
        
        // 2. THEN: Subscribe to events BEFORE creating components
        this.subscribeToWorldViewerEvents();
        
        // 3. FINALLY: Create child components
        this.createComponents();
        
        // Return child components for lifecycle management
        const childComponents: LCMComponent[] = [];
        childComponents.push(this.worldScene!); // Should exist - fail if not
        childComponents.push(this.worldStatsPanel as any); // Should exist - fail if not
        return childComponents;
    }

    /**
     * Phase 3: Activate component when all dependencies are ready
     */
    async activate(): Promise<void> {
        // Bind events now that all components are ready
        this.bindPageSpecificEvents();
    }

    /**
     * Cleanup phase (called by lifecycle controller if needed)
     */
    deactivate(): void {
        // Remove event subscriptions
        this.removeSubscription(WorldEventTypes.WORLD_VIEWER_READY, null);
        
        this.destroy();
    }
    
    /**
     * Subscribe to WorldViewer events before component creation
     */
    private subscribeToWorldViewerEvents(): void {
        // Subscribe to WorldViewer ready event BEFORE creating the component
        this.addSubscription(WorldEventTypes.WORLD_VIEWER_READY, null);
    }
    
    /**
     * Handle incoming events from the EventBus
     */
    public handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
        switch(eventType) {
            case WorldEventTypes.WORLD_VIEWER_READY:
                // Pass the canonical World object directly
                this.worldScene.loadWorld(this.world);
                this.showToast('Success', 'World loaded successfully', 'success');
                break;
                
            default:
                // Call parent implementation for unhandled events
                super.handleBusEvent(eventType, data, target, emitter);
        }
    }

    /**
     * Create PhaserWorldScene and WorldStatsPanel component instances
     */
    private createComponents(): void {
        // Create PhaserWorldScene component
        const worldViewerRoot = this.ensureElement('[data-component="world-viewer"]', 'world-viewer-root');
        this.worldScene = new PhaserWorldScene(worldViewerRoot, this.eventBus, true);
        
        // Create WorldStatsPanel component - pass the content div, not the container with header
        const worldStatsContainer = this.ensureElement('[data-component="world-stats-panel"]', 'world-stats-root');
        const worldStatsContent = worldStatsContainer.querySelector('.p-4.space-y-4') as HTMLElement;
        if (!worldStatsContent) {
            throw new Error('WorldViewerPage: WorldStatsPanel content div not found');
        }
        this.worldStatsPanel = new WorldStatsPanel(worldStatsContent, this.eventBus, true);
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
        if (this.worldScene) {
            this.worldScene.destroy();
            // Force set as null.  We are pushing for fail fast on values
            // that should NOT be null
            this.worldScene = null as any;
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
    // Create page instance (just basic setup)
    const page = new WorldViewerPage("WorldViewerPage");
    
    // Create lifecycle controller with debug logging
    const lifecycleController = new LifecycleController(page.eventBus, LifecycleController.DefaultConfig)
    
    // Start breadth-first initialization
    await lifecycleController.initializeFromRoot(page);
});
