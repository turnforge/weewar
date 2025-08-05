import { BasePage } from '../lib/BasePage';
import { DockviewApi, DockviewComponent } from 'dockview-core';
import { PhaserEditorComponent } from './PhaserEditorComponent';
import { TileStatsPanel } from './TileStatsPanel';
import { KeyboardShortcutManager, ShortcutConfig, KeyboardState } from '../lib/KeyboardShortcutManager';
import { shouldIgnoreShortcut } from '../lib/DOMUtils';
import { Unit, Tile, World, WorldEventType, TilesChangedEventData, UnitsChangedEventData, WorldLoadedEventData } from './World';
import { WorldEditorPageState, PageStateEventType, ToolStateChangedEventData, VisualStateChangedEventData, WorkflowStateChangedEventData, ToolState } from './WorldEditorPageState';
import { EventBus } from '../lib/EventBus';
import { EditorEventTypes, TerrainSelectedPayload, UnitSelectedPayload, BrushSizeChangedPayload, PlacementModeChangedPayload, PlayerChangedPayload, TileClickedPayload, PhaserReadyPayload, GridSetVisibilityPayload, CoordinatesSetVisibilityPayload } from './events';
import { EditorToolsPanel } from './EditorToolsPanel';
import { ReferenceImagePanel } from './ReferenceImagePanel';
import { LCMComponent } from '../lib/LCMComponent';
import { LifecycleController } from '../lib/LifecycleController';
import { BRUSH_SIZE_NAMES , TERRAIN_NAMES } from "./ColorsAndNames"

/**
 * World Editor page with unified World architecture and centralized page state
 * Now implements LCMComponent for breadth-first initialization
 */
class WorldEditorPage extends BasePage {
    private world: World;
    private pageState: WorldEditorPageState;
    private editorOutput: HTMLElement;

    // Dockview interface
    private dockview: DockviewApi;
    
    // Phaser editor component for world editing
    private phaserEditorComponent: PhaserEditorComponent;
    
    // TileStats panel for displaying statistics
    private tileStatsPanel: TileStatsPanel;
    
    // Editor tools panel for terrain/unit selection
    private editorToolsPanel: EditorToolsPanel;
    
    // Reference image panel for reference image controls
    private referenceImagePanel: ReferenceImagePanel;

    // Keyboard shortcut manager
    private keyboardShortcutManager: KeyboardShortcutManager;
    
    // Lifecycle controller for managing component initialization
    private lifecycleController: LifecycleController;

    // State management for undo/restore operations
    // Simplified state backup for preview/cancel functionality
    private savedToolState: ToolState;

    // UI state  
    private hasPendingWorldDataLoad: boolean = false;
    
    // LCMComponent implementation
    
    /**
     * Phase 1: Initialize DOM and discover child components
     */
    public performLocalInit(): LCMComponent[] {
        this.pageState = new WorldEditorPageState(this.eventBus);
        
        // Create World instance early so child components can use it
        this.createWorldInstance();

        this.subscribeToEditorEvents();

        this.initializeSpecificComponents();
        
        // Create child components that implement LCMComponent
        const childComponents: LCMComponent[] = [];
        
        // Create ReferenceImagePanel as a lifecycle-managed component using template
        const referenceTemplate = document.getElementById('reference-image-panel-template');
        if (referenceTemplate) {
            // Use the template element directly - it already has proper structure and styling
            this.referenceImagePanel = new ReferenceImagePanel(referenceTemplate, this.eventBus, true);
            
            // Set dependencies directly using explicit setters
            this.referenceImagePanel.setToastCallback((title: string, message: string, type: 'success' | 'error' | 'info') => {
                this.showToast(title, message, type);
            });
            
            childComponents.push(this.referenceImagePanel);
        }
        
        // Create EditorToolsPanel as a lifecycle-managed component using template
        const toolsTemplate = document.getElementById('tools-panel-template');
        if (toolsTemplate) {
            // Use the template element directly - it already has proper structure and styling
            this.editorToolsPanel = new EditorToolsPanel(toolsTemplate, this.eventBus, true);
            
            // Set dependencies directly using explicit setters  
            this.editorToolsPanel.setPageState(this.pageState);
            
            childComponents.push(this.editorToolsPanel);
        }
        
        // Create TileStatsPanel as a lifecycle-managed component using template
        const tileStatsTemplate = document.getElementById('tilestats-panel-template');
        if (!tileStatsTemplate) {
            throw new Error('tilestats-panel-template not found in DOM');
        }
        
        // Use the template element directly - it already has proper structure and styling
        this.tileStatsPanel = new TileStatsPanel(tileStatsTemplate, this.eventBus, true);
        
        // Set dependencies directly using explicit setters
        if (this.world) {
            this.tileStatsPanel.setWorld(this.world);
        }
        
        childComponents.push(this.tileStatsPanel);
        
        // Create PhaserEditorComponent as a lifecycle-managed component using template
        const canvasTemplate = document.getElementById('canvas-panel-template');
        if (!canvasTemplate) {
            throw new Error('canvas-panel-template not found in DOM');
        }
        
        // Use the template element directly - it already has proper structure and styling
        this.phaserEditorComponent = new PhaserEditorComponent("PhaseEditorComponent", canvasTemplate, this.eventBus, true);
        
        // Set dependencies directly using explicit setters
        this.phaserEditorComponent.setPageState(this.pageState);
        // this.phaserEditorComponent.setWorld(this.world);
        
        childComponents.push(this.phaserEditorComponent);
        // Initialize dockview now that all child components are ready
        this.initializeDockview();
        
        return childComponents;
    }
    
    /**
     * Create World instance and initialize it
     */
    private createWorldInstance(): void {
        // Read initial state from DOM
        const worldIdInput = document.getElementById("worldIdInput") as HTMLInputElement | null;
        const worldId = worldIdInput?.value.trim() || null;
        this.loadWorld(worldId);
    }

    private async loadWorld(worldId: string | null): Promise<void> {
        try {
            const worldMetadataElement = document.getElementById('world-data-json');
            const worldTilesElement = document.getElementById('world-tiles-data-json');
            this.world = new World(this.eventBus).loadFromElement(worldMetadataElement!, worldTilesElement!);
            this.world.setWorldId(worldId)
            // await this.world!.load(worldId);
            this.hasPendingWorldDataLoad = true;
        } catch (error) {
            console.error('Failed to load world:', error);
            this.logToConsole(`Failed to load world: ${error}`);
            this.updateEditorStatus('Load Error');
        }
    }
    
    /**
     * Phase 2: Inject dependencies from lifecycle controller
     */
    public setupDependencies(): void {
        // World is now created in performLocalInit, just update status
        this.updateEditorStatus('Initializing...');
    }
    
    /**
     * Phase 3: Activate the component when all dependencies are ready
     */
    public activate(): void {
        // Bind events now that all components are ready
        this.bindSpecificEvents();
        this.initializeKeyboardShortcuts();
        this.setupUnsavedChangesWarning();
        
        // Update UI state
        this.updateEditorStatus('Ready');
    }
    
    /**
     * Phase 4: Deactivate and cleanup
     */
    public deactivate(): void {
        // Use existing destroy method for cleanup
        this.destroy();
    }
    
    // Dependencies are set directly using explicit setters - no ComponentDependencyDeclaration needed
    
    // World event handlers via EventBus
    private handleWorldLoaded(data: WorldLoadedEventData): void {
        this.updateEditorStatus('Loaded');
        this.updateSaveButtonState();
    }
    
    private handleWorldSaved(data: any): void {
        this.updateEditorStatus('Saved');
        this.updateSaveButtonState();
        if (data.success && data.worldId) {
            // Update URL if this was a new world
            if (this.world?.getIsNewWorld()) {
                history.replaceState(null, '', `/worlds/${data.worldId}/edit`);
            }
        }
    }
    
    private handleWorldDataChanged(): void {
        // World data changed, update UI state
        this.updateSaveButtonState();
    }
    
    
    /**
     * Subscribe to editor-specific events before components are created
     * This prevents race conditions where components emit events before subscribers are ready
     */
    private subscribeToEditorEvents(): void {
        // Subscribe to World events via EventBus
        this.addSubscription(WorldEventType.WORLD_LOADED, this);
        this.addSubscription(WorldEventType.WORLD_SAVED, this);
        this.addSubscription(WorldEventType.TILES_CHANGED, this);
        this.addSubscription(WorldEventType.UNITS_CHANGED, this);
        this.addSubscription(WorldEventType.WORLD_CLEARED, this);
        this.addSubscription(WorldEventType.WORLD_METADATA_CHANGED, this);
        
        // Note: Tool state changes now handled via PageState Observer pattern
        // EditorToolsPanel directly updates pageState, which notifies observers
        
        // Subscribe to Phaser ready event
        this.addSubscription(EditorEventTypes.PHASER_READY, this);
        
        // World changes are automatically tracked by World class via Observer pattern
    }

    /**
     * Handle events from the EventBus
     */
    public handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
        switch(eventType) {
            case WorldEventType.WORLD_LOADED:
                this.handleWorldLoaded(data);
                break;
            
            case WorldEventType.WORLD_SAVED:
                this.handleWorldSaved(data);
                break;
            
            case WorldEventType.TILES_CHANGED:
            case WorldEventType.UNITS_CHANGED:
            case WorldEventType.WORLD_CLEARED:
            case WorldEventType.WORLD_METADATA_CHANGED:
                this.handleWorldDataChanged();
                break;
            
            case EditorEventTypes.PHASER_READY:
                this.handlePhaserReady().then(() => {});
                break;
            
            default:
                // Call parent implementation for unhandled events
                super.handleBusEvent(eventType, data, target, emitter);
        }
    }

    protected initializeSpecificComponents(): LCMComponent[] {
        const worldIdInput = document.getElementById("worldIdInput") as HTMLInputElement | null;
        const isNewWorldInput = document.getElementById("isNewWorld") as HTMLInputElement | null;
        
        // World ID and new world state are now handled by the World instance

        this.editorOutput = document.getElementById('editor-output')!;
        return [];
    }

    private initializeDockview(): void {
        const container = document.getElementById('dockview-container');
        if (!container) {
            console.error('❌ DockView container not found');
            return;
        }

        // Apply theme class based on current theme
        const isDarkMode = document.documentElement.classList.contains('dark');
        container.className = isDarkMode ? 'dockview-theme-dark flex-1' : 'dockview-theme-light flex-1';
        
        // Listen for theme changes
        const observer = new MutationObserver((mutations) => {
            mutations.forEach((mutation) => {
                if (mutation.type === 'attributes' && mutation.attributeName === 'class') {
                    const isDarkMode = document.documentElement.classList.contains('dark');
                    container.className = isDarkMode ? 'dockview-theme-dark flex-1' : 'dockview-theme-light flex-1';
                    
                    // Update Phaser editor theme
                    if (this.phaserEditorComponent) {
                        this.phaserEditorComponent.setTheme(isDarkMode);
                    }
                }
            });
        });
        
        observer.observe(document.documentElement, {
            attributes: true,
            attributeFilter: ['class']
        });
        
        // Create DockView component
        const dockviewComponent = new DockviewComponent(container, {
            createComponent: (options: any) => {
                switch (options.name) {
                    case 'tools':
                        return this.createToolsComponent();
                    case 'phaser':
                        return this.createPhaserComponent();
                    case 'tilestats':
                        return this.createTileStatsComponent();
                    case 'console':
                        return this.createConsoleComponent();
                    case 'advancedTools':
                        return this.createAdvancedToolsComponent();
                    case 'referenceImage':
                        return this.createReferenceImageComponent();
                    default:
                        return {
                            element: document.createElement('div'),
                            init: () => {},
                            dispose: () => {}
                        };
                }
            }
        });

        this.dockview = dockviewComponent.api;
        
        // Load saved layout or create default
        const savedLayout = this.loadDockviewLayout();
        if (savedLayout) {
            try {
                this.dockview.fromJSON(savedLayout);
            } catch (e) {
                console.warn('Failed to restore dockview layout, using default', e);
                this.createDefaultDockviewLayout();
            }
        } else {
            this.createDefaultDockviewLayout();
        }
        
        // Save layout on changes
        this.dockview.onDidLayoutChange(() => {
            this.saveDockviewLayout();
        });

    }

    protected bindSpecificEvents(): void {
        // Header buttons
        const saveButton = document.getElementById('save-world-btn');
        if (saveButton) {
            saveButton.addEventListener('click', this.saveWorld.bind(this));
        }

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            // Don't interfere with input fields (except for specific global shortcuts)
            if (shouldIgnoreShortcut(e)) {
                // Only handle our specific shortcuts in input fields, let other keys pass through
                if ((e.ctrlKey || e.metaKey) && e.key === 's') {
                    e.preventDefault();
                    if (this.world?.getHasUnsavedChanges()) {
                        this.saveWorld();
                    }
                }
                return;
            }
            
            // Ctrl+S or Cmd+S to save
            if ((e.ctrlKey || e.metaKey) && e.key === 's') {
                e.preventDefault();
                if (this.world?.getHasUnsavedChanges()) {
                    this.saveWorld();
                }
            }
        });

        const exportButton = document.getElementById('export-world-btn');
        if (exportButton) {
            exportButton.addEventListener('click', this.exportWorld.bind(this));
        }


        const clearConsoleButton = document.getElementById('clear-console-btn');
        if (clearConsoleButton) {
            clearConsoleButton.addEventListener('click', this.clearConsole.bind(this));
        }

        // World title editing
        const worldTitleInput = document.getElementById('world-title-input') as HTMLInputElement;
        const saveTitleButton = document.getElementById('save-title-btn') as HTMLButtonElement;
        const cancelTitleButton = document.getElementById('cancel-title-btn') as HTMLButtonElement;
        
        if (worldTitleInput && saveTitleButton && cancelTitleButton) {
            let originalTitle = worldTitleInput.value;
            let isEditing = false;
            
            const updateEditingState = (editing: boolean) => {
                isEditing = editing;
                if (editing) {
                    worldTitleInput.classList.add('editing');
                    saveTitleButton.classList.remove('hidden');
                    cancelTitleButton.classList.remove('hidden');
                } else {
                    worldTitleInput.classList.remove('editing');
                    saveTitleButton.classList.add('hidden');
                    cancelTitleButton.classList.add('hidden');
                }
            };
            
            const cancelEditing = () => {
                worldTitleInput.value = originalTitle;
                worldTitleInput.blur();
                updateEditingState(false);
                resizeInput();
            };
            
            const saveTitle = () => {
                const newTitle = worldTitleInput.value.trim();
                if (newTitle && newTitle !== originalTitle) {
                    this.saveWorldTitle(newTitle);
                    originalTitle = newTitle; // Update original after successful save
                }
                worldTitleInput.blur();
                updateEditingState(false);
            };
            
            // Focus events for editing state
            worldTitleInput.addEventListener('focus', () => {
                updateEditingState(true);
            });
            
            worldTitleInput.addEventListener('blur', (e) => {
                // Don't blur if clicking on save/cancel buttons
                const relatedTarget = e.relatedTarget as HTMLElement;
                if (relatedTarget && (relatedTarget.id === 'save-title-btn' || relatedTarget.id === 'cancel-title-btn')) {
                    return;
                }
                
                // Auto-save if there are changes
                const newTitle = worldTitleInput.value.trim();
                if (newTitle && newTitle !== originalTitle) {
                    this.saveWorldTitle(newTitle);
                    originalTitle = newTitle;
                } else if (!newTitle) {
                    worldTitleInput.value = originalTitle;
                }
                updateEditingState(false);
            });
            
            // Input changes
            worldTitleInput.addEventListener('input', () => {
                resizeInput();
                const hasChanges = worldTitleInput.value.trim() !== originalTitle;
                // Update button states based on changes
                if (hasChanges && worldTitleInput.value.trim()) {
                    saveTitleButton.classList.remove('opacity-50');
                    saveTitleButton.disabled = false;
                } else {
                    saveTitleButton.classList.add('opacity-50');
                    saveTitleButton.disabled = true;
                }
            });
            
            // Keyboard shortcuts
            worldTitleInput.addEventListener('keydown', (e) => {
                if (e.key === 'Enter') {
                    e.preventDefault();
                    saveTitle();
                } else if (e.key === 'Escape') {
                    e.preventDefault();
                    cancelEditing();
                }
            });
            
            // Button events
            saveTitleButton.addEventListener('click', (e) => {
                e.preventDefault();
                saveTitle();
            });
            
            cancelTitleButton.addEventListener('click', (e) => {
                e.preventDefault();
                cancelEditing();
            });
            
            // Auto-resize input based on content
            const resizeInput = () => {
                worldTitleInput.style.width = 'auto';
                worldTitleInput.style.width = Math.max(120, worldTitleInput.scrollWidth + 20) + 'px';
            };
            worldTitleInput.addEventListener('input', resizeInput);
            resizeInput(); // Initial resize
        }

        // Export buttons
        document.querySelectorAll('[data-action="export-game"]').forEach(button => {
            button.addEventListener('click', (e) => {
                const target = e.target as HTMLElement;
                const players = parseInt(target.dataset.players || '2');
                this.exportToGame(players);
            });
        });

        // Advanced tool buttons - Use event delegation for dockview compatibility
        document.addEventListener('click', (e) => {
            const target = e.target as HTMLElement;
            const action = target.getAttribute('data-action');
            
            switch (action) {
                case 'fill-all-grass':
                    console.log('Fill All Grass clicked via delegation');
                    this.fillAllGrass();
                    break;
                case 'create-island-world':
                    console.log('Create Island World clicked via delegation');
                    this.createIslandWorld();
                    break;
                case 'create-mountain-ridge':
                    console.log('Create Mountain Ridge clicked via delegation');
                    this.createMountainRidge();
                    break;
                case 'show-terrain-stats':
                    console.log('Show Terrain Stats clicked via delegation');
                    this.showTerrainStats();
                    break;
                case 'randomize-terrain':
                    console.log('Randomize Terrain clicked via delegation');
                    this.randomizeTerrain();
                    break;
                case 'clear-world':
                    console.log('Clear World clicked via delegation');
                    this.clearWorld();
                    break;
                case 'download-image':
                    console.log('Download Image clicked via delegation');
                    this.downloadImage();
                    break;
                case 'download-game-data':
                    console.log('Download Game Data clicked via delegation');
                    this.downloadGameData();
                    break;
            }
        });
        
        // Phaser test buttons
        document.querySelector('[data-action="init-phaser"]')?.addEventListener('click', () => {
            this.initializePhaser();
        });
        
        // Reference image controls are now handled by ReferenceImagePanel directly
        // No event handlers needed here - ReferenceImagePanel binds its own DOM events
    }
    
    /**
     * Bind events specific to the Phaser panel (called when panel is created)
     */
    private bindPhaserPanelEvents(container: HTMLElement): void {
        // Visual options - grid and coordinates checkboxes
        const showGridCheckbox = container.querySelector('#show-grid') as HTMLInputElement;
        if (showGridCheckbox) {
            showGridCheckbox.addEventListener('change', (e) => {
                const checked = (e.target as HTMLInputElement).checked;
                this.setShowGrid(checked);
                this.logToConsole(`Grid checkbox changed to: ${checked}`);
            });
            this.logToConsole('Grid checkbox event handler bound');
        } else {
            this.logToConsole('Grid checkbox not found in Phaser panel');
        }
        
        const showCoordinatesCheckbox = container.querySelector('#show-coordinates') as HTMLInputElement;
        if (showCoordinatesCheckbox) {
            showCoordinatesCheckbox.addEventListener('change', (e) => {
                const checked = (e.target as HTMLInputElement).checked;
                this.setShowCoordinates(checked);
                this.logToConsole(`Coordinates checkbox changed to: ${checked}`);
            });
            this.logToConsole('Coordinates checkbox event handler bound');
        } else {
            this.logToConsole('Coordinates checkbox not found in Phaser panel');
        }
    }

    private initializeKeyboardShortcuts(): void {
        // Context filter to ignore shortcuts when modifier keys are pressed OR when in input fields
        const noModifiersFilter = (event: KeyboardEvent): boolean => {
            return !shouldIgnoreShortcut(event);
        };
        
        const shortcuts: ShortcutConfig[] = [
            // Tab switching shortcuts (single key press)
            {
                key: 'n',
                handler: () => this.switchToNatureTab(),
                description: 'Switch to Nature terrain tab',
                category: 'Navigation',
                requiresArgs: false,
                contextFilter: noModifiersFilter
            },
            {
                key: 'c',
                handler: () => this.switchToCityTab(),
                description: 'Switch to City terrain tab',
                category: 'Navigation',
                requiresArgs: false,
                contextFilter: noModifiersFilter
            },
            {
                key: 'u',
                handler: () => this.switchToUnitTab(),
                description: 'Switch to Unit tab',
                category: 'Navigation',
                requiresArgs: false,
                contextFilter: noModifiersFilter
            },
            
            // Multi-digit number selection within active tab (s + number)
            {
                key: 's',
                handler: (args?: string) => this.selectByNumberInActiveTab(args),
                previewHandler: (args?: string) => this.previewByNumberInActiveTab(args),
                cancelHandler: () => this.cancelNumberSelection(),
                description: 'Select item by number in active tab (s + 1-99)',
                category: 'Selection',
                requiresArgs: true,
                argType: 'number',
                contextFilter: noModifiersFilter
            },
            
            
            // Player selection shortcuts (p + number)
            {
                key: 'p',
                handler: (args?: string) => this.selectPlayer(args),
                previewHandler: (args?: string) => this.previewPlayer(args),
                cancelHandler: () => this.cancelSelection(),
                description: 'Set current player',
                category: 'Units',
                requiresArgs: true,
                argType: 'number',
                contextFilter: noModifiersFilter
            },
            
            // Brush size shortcuts (b + number)
            {
                key: 'b',
                handler: (args?: string) => this.selectBrushSize(args),
                previewHandler: (args?: string) => this.previewBrushSize(args),
                cancelHandler: () => this.cancelSelection(),
                description: 'Set brush size (1-6)',
                category: 'Tools',
                requiresArgs: true,
                argType: 'number',
                contextFilter: noModifiersFilter
            },
            
            
            // Clear mode and reset shortcuts
            {
                key: 'Escape',
                handler: () => this.activateClearMode(),
                description: 'Activate clear mode',
                category: 'Tools',
                contextFilter: noModifiersFilter
                /*
            },
            {
                key: 'r',
                handler: () => this.resetToDefaults(),
                description: 'Reset all tools to defaults',
                category: 'General',
                contextFilter: noModifiersFilter
               */
            }
        ];

        this.keyboardShortcutManager = new KeyboardShortcutManager({
            shortcuts,
            timeout: 2000,
            immediateExecution: true,
            previewDelay: 300,
            onStateChange: (state, command) => this.handleKeyboardStateChange(state, command)
        });
        
        // Add custom number input handler for pure digit sequences
        this.setupCustomNumberInput();
        
    }

    /**
     * Show loading indicator on world
     */


    // Editor functions called by the template

    public setBrushTerrain(terrain: number): void {
        if (this.pageState) {
            this.pageState.setSelectedTerrain(terrain);
        }
        
        this.updateBrushInfo();
        // Button selection now handled by EditorToolsPanel component
    }

    public setBrushSize(size: number): void {
        if (this.pageState) {
            this.pageState.setBrushSize(size);
        }
        
        this.updateBrushInfo();
    }
    
    public setShowGrid(showGrid: boolean): void {
        // Update page state - this will emit visual state changed event
        if (this.pageState) {
            this.pageState.setShowGrid(showGrid);
        }
        
        // Emit event to PhaserEditorComponent via EventBus
        this.eventBus.emit<GridSetVisibilityPayload>(
            EditorEventTypes.GRID_SET_VISIBILITY,
            { show: showGrid },
            this,
            this.pageState
        );
        this.logToConsole(`Grid visibility set to: ${showGrid}`);
    }
    
    public setShowCoordinates(showCoordinates: boolean): void {
        // Update page state - this will emit visual state changed event
        if (this.pageState) {
            this.pageState.setShowCoordinates(showCoordinates);
        }
        
        // Emit event to PhaserEditorComponent via EventBus
        this.eventBus.emit<CoordinatesSetVisibilityPayload>(
            EditorEventTypes.COORDINATES_SET_VISIBILITY,
            { show: showCoordinates },
            this,
            this.pageState
        );
        this.logToConsole(`Coordinates visibility set to: ${showCoordinates}`);
    }

    public downloadImage(): void {
        // TODO: Implement image download
        this.showToast('Download', 'Image download not yet implemented', 'info');
    }

    public exportToGame(players: number): void {
        // TODO: Implement game export
        this.showToast('Export', `${players}-player game export not yet implemented`, 'info');
    }

    public downloadGameData(): void {
        // TODO: Implement game data download
        this.showToast('Download', 'Game data download not yet implemented', 'info');
    }

    // Advanced tool functions
    public fillAllGrass(): void {
        
        if (this.world) {
            this.world.fillAllTerrain(1, 0); // Terrain type 1 = Grass
            this.logToConsole('Filled world with grass via World Observer pattern');
        } else {
            this.logToConsole('World not available, cannot fill grass');
        }
    }

    public createIslandWorld(): void {
        
        if (this.phaserEditorComponent && this.phaserEditorComponent.getIsInitialized()) {
            // Get current viewport center
            const center = this.phaserEditorComponent.getViewportCenter();
            
            // Create island pattern at viewport center with radius 5
            this.phaserEditorComponent.createIslandPattern(center.q, center.r, 5);
        } else {
        }
    }

    public createMountainRidge(): void {
        
        if (this.phaserEditorComponent && this.phaserEditorComponent.getIsInitialized() && this.world) {
            // Get current viewport center
            const center = this.phaserEditorComponent.getViewportCenter();
            
            // Create a horizontal mountain ridge centered around viewport center
            const ridgeWidth = 9; // from -4 to +4
            const ridgeHeight = 5; // from -2 to +2
            const startQ = center.q - Math.floor(ridgeWidth / 2);
            const startR = center.r - Math.floor(ridgeHeight / 2);
            
            for (let q = startQ; q < startQ + ridgeWidth; q++) {
                for (let r = startR; r < startR + ridgeHeight; r++) {
                    const relativeR = r - center.r;
                    // Create a ridge pattern - mountains in center, rocks on edges
                    if (Math.abs(relativeR) <= 1) {
                        this.world.setTileAt(q, r, 4, 0); // Mountain
                    } else {
                        this.world.setTileAt(q, r, 5, 0); // Rock
                    }
                }
            }
        } else {
            this.logToConsole('World or Phaser panel not available, cannot create mountain ridge');
        }
    }

    public showTerrainStats(): void {
        
      /*
        if (this.phaserEditorComponent && this.phaserEditorComponent.getIsInitialized()) {
            const tiles = this.world?.getAllTiles();
            const stats = {
                grass: 0,
                desert: 0,
                water: 0,
                mountain: 0,
                rock: 0,
                other: 0
            };
            
            tiles.forEach((tile: any) => {
                switch (tile.terrain) {
                    case 1: stats.grass++; break;
                    case 2: stats.desert++; break;
                    case 3: stats.water++; break;
                    case 4: stats.mountain++; break;
                    case 5: stats.rock++; break;
                    default: stats.other++; break;
                }
            });
            
            if (stats.other > 0) {
            }
            this.logToConsole(`Total tiles: ${tiles.length}`);
        } else {
        }
       */
    }

    public randomizeTerrain(): void {
        this.logToConsole('Randomizing terrain...');
        
        if (this.phaserEditorComponent && this.phaserEditorComponent.getIsInitialized()) {
            this.phaserEditorComponent.randomizeTerrain();
            this.logToConsole('Terrain randomized using Phaser');
        } else {
            this.logToConsole('Phaser panel not available, cannot randomize terrain');
        }
    }

    public clearWorld(): void {
        console.log('clearWorld() method called');
        this.logToConsole('Clearing entire world...');
        
        // Clear world data - this will trigger observer notifications to update Phaser
        if (this.world) {
            console.log('World instance exists, calling clearAll()');
            this.world.clearAll();
            this.logToConsole('World data cleared - Phaser will update via observer pattern');
        } else {
            console.log('World instance is null!');
            this.logToConsole('World instance not available');
        }
        
        // Show success message
        this.showToast('World Cleared', 'All tiles and units have been removed', 'info');
        this.logToConsole('Clear world operation completed');
    }

    // Canvas management methods removed - now handled by Phaser panel

    private async saveWorld(): Promise<void> {
        if (!this.world) {
            this.showToast('Error', 'No world data to save', 'error');
            return;
        }

        try {
            this.updateEditorStatus('Saving...');
            const result = await this.world.save();

            if (result.success) {
                this.showToast('Success', 'World saved successfully', 'success');
            } else {
                throw new Error(result.error || 'Unknown save error');
            }
        } catch (error) {
            console.error('Save failed:', error);
            this.logToConsole(`Save failed: ${error}`);
            this.updateEditorStatus('Save Error');
            this.showToast('Error', 'Failed to save world', 'error');
        }
    }

    private async exportWorld(): Promise<void> {
        if (!this.world || !this.phaserEditorComponent || !this.phaserEditorComponent.getIsInitialized()) {
            this.showToast('Error', 'No world data to export', 'error');
            return;
        }

        // World now handles its own export operations
        const result = await this.world.save();
        
        if (result.success) {
            this.showToast('Success', 'World exported successfully', 'success');
        } else {
            this.showToast('Error', result.error || 'Failed to export world', 'error');
        }
    }

    private async saveWorldTitle(newTitle: string): Promise<void> {
        if (!newTitle.trim()) {
            this.showToast('Error', 'World title cannot be empty', 'error');
            return;
        }

        const oldTitle = this.world?.getName() || 'Untitled World';

        // Update the local world data
        if (this.world) {
            this.world.setName(newTitle);
        }

        try {
            this.logToConsole(`Updating world title to: ${newTitle}`);
            
            // Save the world (this will include the title update)
            await this.saveWorld();
            
            this.showToast('Success', 'World title updated', 'success');
            
        } catch (error) {
            console.error('Failed to save world title:', error);
            this.logToConsole(`Failed to save world title: ${error}`);
            this.showToast('Error', 'Failed to update world title', 'error');
            
            // Revert the title on error
            if (this.world) {
                this.world.setName(oldTitle);
            }
            const worldTitleInput = document.getElementById('world-title-input') as HTMLInputElement;
            if (worldTitleInput) {
                worldTitleInput.value = oldTitle;
            }
        }
    }


    private clearConsole(): void {
        if (this.editorOutput) {
            this.editorOutput.innerHTML = '';
        }
    }

    // Utility methods
    private logToConsole(message: string): void {
        if (this.editorOutput) {
            const timestamp = new Date().toLocaleTimeString();
            const logEntry = `[${timestamp}] ${message}`;
            
            // Use innerHTML to properly handle line breaks
            const currentContent = this.editorOutput.innerHTML;
            this.editorOutput.innerHTML = currentContent + (currentContent ? '<br>' : '') + this.escapeHtml(logEntry);
            this.editorOutput.scrollTop = this.editorOutput.scrollHeight;
        }
        console.log(`[WorldEditor] ${message}`);
    }

    private escapeHtml(text: string): string {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    private updateEditorStatus(status: string): void {
        const statusElement = document.getElementById('editor-status');
        if (statusElement) {
            statusElement.textContent = status;
            
            // Update status color based on state
            statusElement.className = 'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium';
            if (status.includes('Error')) {
                statusElement.className += ' bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200';
            } else if (status === 'Ready' || status === 'Saved' || status === 'Loaded') {
                statusElement.className += ' bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
            } else {
                statusElement.className += ' bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200';
            }
        }
    }

    private updateBrushInfo(): void {
        const brushInfo = document.getElementById('brush-info');
        if (brushInfo) {
            const currentTerrain = this.pageState?.getToolState().selectedTerrain || 1;
            const currentBrushSize = this.pageState?.getToolState().brushSize || 0;
            brushInfo.textContent = `Current: ${TERRAIN_NAMES[currentTerrain]}, ${BRUSH_SIZE_NAMES[currentBrushSize]}`;
        }
    }

    // Note: Terrain button selection is now handled by EditorToolsPanel internally

    // Theme management is handled by BasePage

    // Dockview panel creation methods
    private createToolsComponent() {
        // Use the lifecycle-managed EditorToolsPanel if available
        if (this.editorToolsPanel) {
            const container = this.editorToolsPanel.rootElement;
            return {
                element: container,
                init: () => {
                    // Panel is already initialized through lifecycle controller
                    console.log('EditorToolsPanel dockview component initialized (lifecycle-managed)');
                },
                dispose: () => {
                    // Cleanup handled by lifecycle controller
                }
            };
        }
        
        // Fallback: use the template element directly from the DOM
        const template = document.getElementById('tools-panel-template');
        if (!template) {
            console.error('Tools panel template not found');
            return { element: document.createElement('div'), init: () => {}, dispose: () => {} };
        }
        
        // Use the template element directly - no cloning needed
        template.style.display = 'block';
        
        return {
            element: template,
            init: () => {
                console.log('EditorToolsPanel dockview component initialized (fallback mode using template)');
            },
            dispose: () => {}
        };
    }

    private createPhaserComponent() {
        // Use the lifecycle-managed PhaserEditorComponent if available
        if (this.phaserEditorComponent) {
            // The lifecycle-managed component already has the template structure
            const container = this.phaserEditorComponent.rootElement;
            
            return {
                element: container,
                init: () => {
                    console.log('PhaserEditorComponent dockview component initialized (lifecycle-managed)');
                    // Bind grid and coordinates checkboxes since the template is already in the component
                    this.bindPhaserPanelEvents(container);
                },
                dispose: () => {
                    // Cleanup handled by lifecycle controller
                }
            };
        }
        
        // Fallback: use the template element directly from the DOM
        const template = document.getElementById('canvas-panel-template');
        if (!template) {
            console.error('Phaser panel template not found');
            return { element: document.createElement('div'), init: () => {}, dispose: () => {} };
        }
        
        // Use the template element directly - no cloning needed
        template.style.display = 'block';
        
        return {
            element: template,
            init: () => {
                // Initialize PhaserEditorComponent using the template directly
                // PhaserEditorComponent will find the #editor-canvas-container within this template
                this.phaserEditorComponent = new PhaserEditorComponent("PhaserEditorComponent", template, this.eventBus, this.debugMode);
                this.logToConsole('PhaserEditorComponent initialized using template directly');
                
                // Bind grid and coordinates checkboxes
                this.bindPhaserPanelEvents(template);
            },
            dispose: () => {
                if (this.phaserEditorComponent) {
                    this.phaserEditorComponent.destroy();
                    this.phaserEditorComponent = null as any;
                }
            }
        };
    }

    private createTileStatsComponent() {
        // Use the lifecycle-managed TileStatsPanel if available
        if (this.tileStatsPanel) {
            const container = this.tileStatsPanel.rootElement;
            return {
                element: container,
                init: () => {
                    // Panel is already initialized through lifecycle controller
                    console.log('TileStatsPanel dockview component initialized (lifecycle-managed)');
                },
                dispose: () => {
                    // Cleanup handled by lifecycle controller
                }
            };
        }
        
        // Fallback: use the template element directly from the DOM
        const template = document.getElementById('tilestats-panel-template');
        if (!template) {
            console.error('TileStats panel template not found');
            return { element: document.createElement('div'), init: () => {}, dispose: () => {} };
        }
        
        // Use the template element directly - no cloning needed
        template.style.display = 'block';
        
        return {
            element: template,
            init: () => {
                console.log('TileStatsPanel dockview component initialized (fallback mode using template)');
            },
            dispose: () => {}
        };
    }

    private createConsoleComponent() {
        const template = document.getElementById('console-panel-template');
        if (!template) {
            console.error('Console panel template not found');
            return { element: document.createElement('div'), init: () => {}, dispose: () => {} };
        }
        
        // Use the template element directly - no cloning needed
        template.style.display = 'block';
        
        return {
            element: template,
            init: () => {
                // Find the editor output element within the template
                const outputElement = template.querySelector('#editor-output');
                if (outputElement) {
                    this.editorOutput = outputElement as HTMLElement;
                }
            },
            dispose: () => {}
        };
    }

    private createAdvancedToolsComponent() {
        const template = document.getElementById('advanced-tools-panel-template');
        if (!template) {
            console.error('Advanced tools panel template not found');
            return { element: document.createElement('div'), init: () => {}, dispose: () => {} };
        }
        
        // Use the template element directly - no cloning needed
        template.style.display = 'block';
        
        return {
            element: template,
            init: () => {
                // Advanced tools panel is already initialized through global event binding
            },
            dispose: () => {}
        };
    }
    
    private createReferenceImageComponent() {
        // Use the lifecycle-managed ReferenceImagePanel if available
        if (this.referenceImagePanel) {
            const container = this.referenceImagePanel.rootElement;
            return {
                element: container,
                init: () => {
                    // Panel is already initialized through lifecycle controller
                    console.log('ReferenceImagePanel dockview component initialized (lifecycle-managed)');
                },
                dispose: () => {
                    // Cleanup handled by lifecycle controller
                }
            };
        }
        
        // Fallback: use the template element directly from the DOM
        const template = document.getElementById('reference-image-panel-template');
        if (!template) {
            console.error('Reference image panel template not found');
            return { element: document.createElement('div'), init: () => {}, dispose: () => {} };
        }
        
        // Use the template element directly - no cloning needed
        template.style.display = 'block';
        
        return {
            element: template,
            init: () => {
                console.log('ReferenceImagePanel dockview component initialized (fallback mode using template)');
            },
            dispose: () => {}
        };
    }

    private createDefaultDockviewLayout(): void {
        if (!this.dockview) return;

        // Add main Phaser World editor panel first (center) - will take remaining width
        this.dockview.addPanel({
            id: 'phaser',
            component: 'phaser',
            title: '🗺️ World Editor'
        });

        // Add tools panel to the left of Phaser (270px width)
        this.dockview.addPanel({
            id: 'tools',
            component: 'tools',
            title: '🎨 Tools & Terrain',
            position: { direction: 'left', referencePanel: 'phaser' }
        });

        // Add advanced tools panel to the right of Phaser (260px width)
        this.dockview.addPanel({
            id: 'advancedTools',
            component: 'advancedTools',
            title: '🔧 Advanced & View',
            position: { direction: 'right', referencePanel: 'phaser' }
        });

        // Add TileStats panel below the Advanced Tools panel
        this.dockview.addPanel({
            id: 'tilestats',
            component: 'tilestats',
            title: '📊 World Statistics',
            position: { direction: 'below', referencePanel: 'advancedTools' }
        });
        
        // Add Reference Image panel below the TileStats panel
        this.dockview.addPanel({
            id: 'referenceImage',
            component: 'referenceImage',
            title: '🖼️ Reference Image',
            position: { direction: 'below', referencePanel: 'tilestats' }
        });

        // Add console panel below Phaser (250px height)
        this.dockview.addPanel({
            id: 'console',
            component: 'console',
            title: '💻 Console',
            position: { direction: 'below', referencePanel: 'phaser' }
        });

        // Set panel sizes after layout is created
        setTimeout(() => {
            this.setPanelSizes();
        }, 100);
    }

    private setPanelSizes(): void {
        if (!this.dockview) return;

            // Set left panel (Tools) to 270px width
            const toolsPanel = this.dockview.getPanel('tools');
            if (toolsPanel) {
                toolsPanel.api.setSize({ width: 270 });
            }

            // Set right panel (Advanced Tools) to 260px width
            const advancedToolsPanel = this.dockview.getPanel('advancedTools');
            if (advancedToolsPanel) {
                advancedToolsPanel.api.setSize({ width: 260 });
            }

            const consolePanel = this.dockview.getPanel('console');
            if (consolePanel) {
                consolePanel.api.setSize({ height: 250 });
            }
            
            // Set reference image panel to 300px height to accommodate controls
            const referenceImagePanel = this.dockview.getPanel('referenceImage');
            if (referenceImagePanel) {
                referenceImagePanel.api.setSize({ height: 300 });
            }

            this.logToConsole('Panel sizes set: Tools=270px, Advanced=260px, ReferenceImage=300px, World Editor=remaining');
    }

    private saveDockviewLayout(): void {
        if (!this.dockview) return;
        
        const layout = this.dockview.toJSON();
        localStorage.setItem('world-editor-dockview-layout', JSON.stringify(layout));
    }
    
    private loadDockviewLayout(): any {
        const saved = localStorage.getItem('world-editor-dockview-layout');
        return saved ? JSON.parse(saved) : null;
    }

    // Unsaved changes tracking
    private setupUnsavedChangesWarning(): void {
        // Browser beforeunload warning
        window.addEventListener('beforeunload', (e) => {
            if (this.world?.getHasUnsavedChanges()) {
                e.preventDefault();
                e.returnValue = 'You have unsaved changes. Are you sure you want to leave?';
                return 'You have unsaved changes. Are you sure you want to leave?';
            }
        });
        
        // Initialize save button state
        setTimeout(() => {
            this.updateSaveButtonState();
        }, 100);
    }
    
    // World changes are now automatically tracked via Observer pattern
    // No need for manual tracking
    
    private updateSaveButtonState(): void {
        const saveButton = document.getElementById('save-world-btn');
        if (saveButton && this.world) {
            if (this.world.getHasUnsavedChanges()) {
                saveButton.classList.remove('opacity-50');
                saveButton.classList.add('bg-blue-600', 'hover:bg-blue-700');
                saveButton.removeAttribute('disabled');
            } else {
                saveButton.classList.add('opacity-50');
                saveButton.classList.remove('bg-blue-600', 'hover:bg-blue-700');
                saveButton.setAttribute('disabled', 'true');
            }
        }
    }

    public destroy(): void {
        // Save layout before destroying
        this.saveDockviewLayout();
        
        // Dispose dockview
        if (this.dockview) {
            this.dockview.dispose();
        }
        
        // Destroy Phaser component if it exists (will be handled by dockview component disposal)
        // this.phaserEditorComponent cleanup is handled in createPhaserComponent dispose callback
        
        // Destroy TileStats panel if it exists
        if (this.tileStatsPanel) {
            this.tileStatsPanel.destroy();
        }
        
        // Destroy keyboard shortcut manager if it exists
        if (this.keyboardShortcutManager) {
            this.keyboardShortcutManager.destroy();
        }
    }
    
    // Phaser panel methods
    // OLD METHOD REMOVED: initializePhaserPanel - now handled by PhaserEditorComponent
    
        
    private refreshTileStats(): void {
        if (!this.tileStatsPanel || !this.tileStatsPanel.getIsInitialized()) {
            return;
        }
        
        // TileStatsPanel now reads directly from World
        this.tileStatsPanel.refreshStats();
    }
    
    /**
     * Center the camera on the loaded world by calculating bounds and focusing on center
     */
    private centerCameraOnWorld(): void {
        if (!this.phaserEditorComponent || !this.phaserEditorComponent.getIsInitialized() || !this.world) {
            this.logToConsole('Cannot center camera - components not ready');
            return;
        }
        
        const allTiles = this.world.getAllTiles();
        if (allTiles.length === 0) {
            this.logToConsole('No tiles to center camera on');
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
        
        this.logToConsole(`Centering camera on Q=${centerQ}, R=${centerR} (bounds: Q=${minQ}-${maxQ}, R=${minR}-${maxR})`);
        
        // Center the camera using the PhaserEditorComponent's method
        this.phaserEditorComponent.centerCamera(centerQ, centerR);
    }
    
    // Simplified state management for preview/cancel functionality
    private saveUIState(): void {
        if (this.pageState) {
            // Save current tool state as a snapshot
            this.savedToolState = { ...this.pageState.getToolState() };
        }
    }
    
    private restoreUIState(): void {
        if (this.savedToolState && this.pageState) {
            // Restore tool state via pageState methods
            this.pageState.setSelectedTerrain(this.savedToolState.selectedTerrain);
            this.pageState.setSelectedUnit(this.savedToolState.selectedUnit);
            this.pageState.setSelectedPlayer(this.savedToolState.selectedPlayer);
            this.pageState.setBrushSize(this.savedToolState.brushSize);
            // placementMode is automatically set by setSelectedTerrain/setSelectedUnit
            
            // UI element updates are handled by EditorToolsPanel via pageState observers
            
            // Clear saved state
            this.savedToolState = null as any;
        }
    }
    
    private showPreviewIndicator(message: string): void {
        // Add visual indicator for preview state
        const existingIndicator = document.getElementById('preview-indicator');
        if (existingIndicator) {
            existingIndicator.textContent = message;
            return;
        }
        
        const indicator = document.createElement('div');
        indicator.id = 'preview-indicator';
        indicator.className = 'fixed top-16 right-4 bg-orange-500 text-white px-3 py-2 rounded-lg shadow-lg z-40 font-medium text-sm';
        indicator.textContent = message;
        document.body.appendChild(indicator);
    }
    
    private hidePreviewIndicator(): void {
        const indicator = document.getElementById('preview-indicator');
        if (indicator) {
            indicator.remove();
        }
    }
    
    // Visual index worldping functions
    private getTerrainIdByNatureIndex(index: number): number | null {
        if (index === 0) return 0; // Clear button
        
        const button = document.querySelector(`[data-nature-index="${index}"]`) as HTMLElement;
        if (button) {
            return parseInt(button.getAttribute('data-terrain') || '0');
        }
        return null;
    }
    
    private getTerrainIdByCityIndex(index: number): number | null {
        const button = document.querySelector(`[data-city-index="${index}"]`) as HTMLElement;
        if (button) {
            return parseInt(button.getAttribute('data-terrain') || '0');
        }
        return null;
    }
    
    private getUnitIdByIndex(index: number): number | null {
        const button = document.querySelector(`[data-unit-index="${index}"]`) as HTMLElement;
        if (button) {
            return parseInt(button.getAttribute('data-unit') || '0');
        }
        return null;
    }
    
    // Helper functions to get terrain names from buttons
    private getTerrainNameByNatureIndex(index: number): string {
        if (index === 0) return 'Clear';
        
        const button = document.querySelector(`[data-nature-index="${index}"]`) as HTMLElement;
        if (button) {
            const title = button.getAttribute('title') || '';
            // Extract name from title (e.g., "Grass (Move: 1, Defense: 0)" -> "Grass")
            const name = title.split('(')[0].trim();
            return name || 'Unknown';
        }
        return 'Unknown';
    }
    
    private getTerrainNameByCityIndex(index: number): string {
        const button = document.querySelector(`[data-city-index="${index}"]`) as HTMLElement;
        if (button) {
            const title = button.getAttribute('title') || '';
            // Extract name from title (e.g., "Land Base (Move: 1, Defense: 0)" -> "Land Base")
            const name = title.split('(')[0].trim();
            return name || 'Unknown';
        }
        return 'Unknown';
    }
    
    private getUnitNameByIndex(index: number): string {
        const button = document.querySelector(`[data-unit-index="${index}"]`) as HTMLElement;
        if (button) {
            const title = button.getAttribute('title') || '';
            return title || 'Unknown';
        }
        return 'Unknown';
    }
    

    // Preview handlers for immediate execution mode
    private previewNatureTerrain(args?: string): void {
        const index = parseInt(args || '1');
        
        if (index === 0) {
            // N+0 for clear mode
            this.saveUIState();
            if (this.pageState) {
                this.pageState.setSelectedTerrain(0);
            }
            // Button selection handled by EditorToolsPanel
            this.showPreviewIndicator('Preview: Clear mode');
            return;
        }
        
        // Use visual index worldping
        const terrainId = this.getTerrainIdByNatureIndex(index);
        const terrainName = this.getTerrainNameByNatureIndex(index);
        
        if (terrainId !== null) {
            this.saveUIState();
            this.setBrushTerrain(terrainId);
            // Placement mode updated via pageState.setSelectedTerrain()
            // Button selection handled by EditorToolsPanel
            this.showPreviewIndicator(`Preview: ${terrainName} terrain`);
        }
    }
    
    private previewCityTerrain(args?: string): void {
        const index = parseInt(args || '1');
        
        // Use visual index worldping
        const terrainId = this.getTerrainIdByCityIndex(index);
        const terrainName = this.getTerrainNameByCityIndex(index);
        
        if (terrainId !== null) {
            this.saveUIState();
            this.setBrushTerrain(terrainId);
            // Placement mode updated via pageState.setSelectedTerrain()
            // Button selection handled by EditorToolsPanel
            this.showPreviewIndicator(`Preview: ${terrainName}`);
        }
    }
    
    private previewUnit(args?: string): void {
        const index = parseInt(args || '1');
        
        // Use visual index worldping
        const unitId = this.getUnitIdByIndex(index);
        const unitName = this.getUnitNameByIndex(index);
        
        if (unitId !== null) {
            this.saveUIState();
            if (this.pageState) {
                this.pageState.setSelectedUnit(unitId);
            }
            // Button selection handled by EditorToolsPanel
            const currentPlayer = this.pageState?.getToolState().selectedPlayer || 1;
            this.showPreviewIndicator(`Preview: ${unitName} for player ${currentPlayer}`);
        }
    }
    
    private previewPlayer(args?: string): void {
        const playerId = parseInt(args || '1');
        
        if (playerId >= 1 && playerId <= 4) {
            this.saveUIState();
            if (this.pageState) {
                this.pageState.setSelectedPlayer(playerId);
            }
            
            const unitPlayerSelect = document.getElementById('unit-player-color') as HTMLSelectElement;
            if (unitPlayerSelect) {
                unitPlayerSelect.value = playerId.toString();
            }
            
            this.showPreviewIndicator(`Preview: Player ${playerId} selected`);
        }
    }
    
    private previewBrushSize(args?: string): void {
        const index = parseInt(args || '1'); // 1-based index
        
        // World 1-based index to actual brush size values
        const brushSizeValues = [0, 1, 3, 5, 10, 15]; // Corresponds to the select options
        
        if (index >= 1 && index <= brushSizeValues.length) {
            this.saveUIState();
            const actualSize = brushSizeValues[index - 1];
            this.setBrushSize(actualSize);
            
            const brushSizeSelect = document.getElementById('brush-size') as HTMLSelectElement;
            if (brushSizeSelect) {
                brushSizeSelect.value = actualSize.toString();
            }
            
            this.showPreviewIndicator(`Preview: ${BRUSH_SIZE_NAMES[index - 1]} brush`);
        }
    }
    
    private cancelSelection(): void {
        this.restoreUIState();
        this.hidePreviewIndicator();
        this.logToConsole('Selection cancelled - state restored');
    }

    // Tab switching handlers
    private switchToNatureTab(): void {
        if (this.editorToolsPanel) {
            // If already on nature tab, toggle overlay for nature tab
            if (this.editorToolsPanel.getActiveTab() === 'nature') {
                this.toggleOverlayForTab('nature');
                return;
            }
            
            // Switch to nature tab and turn on overlay
            this.editorToolsPanel.switchToTab('nature');
            this.setOverlayForTab('nature', true);
            this.logToConsole('Switched to Nature terrain tab');
        }
    }
    
    private switchToCityTab(): void {
        if (this.editorToolsPanel) {
            // If already on city tab, toggle overlay for city tab
            if (this.editorToolsPanel.getActiveTab() === 'city') {
                this.toggleOverlayForTab('city');
                return;
            }
            
            // Switch to city tab and turn on overlay
            this.editorToolsPanel.switchToTab('city');
            this.setOverlayForTab('city', true);
            this.logToConsole('Switched to City terrain tab');
        }
    }
    
    private switchToUnitTab(): void {
        if (this.editorToolsPanel) {
            // If already on unit tab, toggle overlay for unit tab
            if (this.editorToolsPanel.getActiveTab() === 'unit') {
                this.toggleOverlayForTab('unit');
                return;
            }
            
            // Switch to unit tab and turn on overlay
            this.editorToolsPanel.switchToTab('unit');
            this.setOverlayForTab('unit', true);
            this.logToConsole('Switched to Unit tab');
        }
    }
    
    // Number overlay control methods per tab
    private setOverlayForTab(tab: 'nature' | 'city' | 'unit', visible: boolean): void {
        if (!this.editorToolsPanel) return;
        
        this.overlayState[tab] = visible;
        
        // If this is the active tab, update the visual overlay
        if (this.editorToolsPanel.getActiveTab() === tab) {
            if (visible) {
                this.editorToolsPanel.showNumberOverlays();
            } else {
                this.editorToolsPanel.hideNumberOverlays();
            }
        }
        
        this.logToConsole(`${tab.charAt(0).toUpperCase() + tab.slice(1)} tab overlay ${visible ? 'shown' : 'hidden'}`);
    }
    
    private toggleOverlayForTab(tab: 'nature' | 'city' | 'unit'): void {
        const currentState = this.overlayState[tab];
        this.setOverlayForTab(tab, !currentState);
    }
    
    private isAnyOverlayVisible(): boolean {
        return Object.values(this.overlayState).some(visible => visible);
    }
    
    // Called when tab is switched (not via keyboard shortcuts)
    private onTabSwitched(tabName: 'nature' | 'city' | 'unit'): void {
        // Update overlay visibility based on the new tab's state
        const shouldShow = this.overlayState[tabName];
        
        if (this.editorToolsPanel) {
            if (shouldShow) {
                this.editorToolsPanel.showNumberOverlays();
            } else {
                this.editorToolsPanel.hideNumberOverlays();
            }
        }
    }
    
    // Custom number input handling for pure digit sequences
    private numberInputBuffer: string = '';
    private numberInputTimeout: number | null = null;
    private isInKeyboardShortcutMode: boolean = false;
    
    // Number overlay toggle state per tab
    private overlayState = {
        nature: false,
        city: false,
        unit: false
    };
    
    private setupCustomNumberInput(): void {
        // Add our own keydown listener that runs before the KeyboardShortcutManager
        document.addEventListener('keydown', (event) => {
            // Only handle if we're not in keyboard shortcut mode and should not ignore shortcuts
            if (this.isInKeyboardShortcutMode || shouldIgnoreShortcut(event)) {
                return;
            }
            
            // Handle digit input
            if (event.key >= '0' && event.key <= '9') {
                event.preventDefault();
                this.handleDigitInput(event.key);
            } else if (event.key === 'Enter' && this.numberInputBuffer) {
                event.preventDefault();
                this.executeNumberSelection();
            } else if (event.key === 'Escape' && this.numberInputBuffer) {
                event.preventDefault();
                this.clearNumberInput();
            } else if (event.key === 'Backspace' && this.numberInputBuffer) {
                event.preventDefault();
                this.removeLastDigit();
            }
        }, true); // Use capture phase to run before other handlers
    }
    
    private handleKeyboardStateChange(state: KeyboardState, command?: string): void {
        this.isInKeyboardShortcutMode = state !== KeyboardState.NORMAL;
        
        if (state === KeyboardState.AWAITING_ARGS && command === 's') {
            // Show overlays for 's' command (temporary)
            if (this.editorToolsPanel) {
                this.editorToolsPanel.showNumberOverlays();
            }
        } else if (state === KeyboardState.NORMAL) {
            // When returning to normal state, restore overlay based on current tab's state
            if (this.editorToolsPanel) {
                const activeTab = this.editorToolsPanel.getActiveTab();
                const shouldShow = this.overlayState[activeTab];
                
                if (shouldShow) {
                    this.editorToolsPanel.showNumberOverlays();
                } else {
                    this.editorToolsPanel.hideNumberOverlays();
                }
            }
        }
    }
    
    private handleDigitInput(digit: string): void {
        // Clear any existing timeout
        if (this.numberInputTimeout) {
            clearTimeout(this.numberInputTimeout);
        }
        
        // Add digit to buffer
        this.numberInputBuffer += digit;
        
        // Show overlays when starting number input
        if (this.editorToolsPanel && this.numberInputBuffer.length === 1) {
            this.editorToolsPanel.showNumberOverlays();
        }
        
        // Set timeout to execute after 1.5 seconds of no input
        this.numberInputTimeout = window.setTimeout(() => {
            this.executeNumberSelection();
        }, 1500);
        
        this.logToConsole(`Number input: ${this.numberInputBuffer}`);
    }
    
    private removeLastDigit(): void {
        if (this.numberInputBuffer.length > 0) {
            this.numberInputBuffer = this.numberInputBuffer.slice(0, -1);
            
            if (this.numberInputBuffer.length === 0) {
                this.clearNumberInput();
            } else {
                // Reset timeout
                if (this.numberInputTimeout) {
                    clearTimeout(this.numberInputTimeout);
                }
                this.numberInputTimeout = window.setTimeout(() => {
                    this.executeNumberSelection();
                }, 1500);
                
                this.logToConsole(`Number input: ${this.numberInputBuffer}`);
            }
        }
    }
    
    private executeNumberSelection(): void {
        if (!this.editorToolsPanel || !this.numberInputBuffer) return;
        
        const index = parseInt(this.numberInputBuffer);
        if (isNaN(index)) {
            this.clearNumberInput();
            return;
        }
        
        const activeTab = this.editorToolsPanel.getActiveTab();
        this.editorToolsPanel.selectByIndex(index);
        this.logToConsole(`Selected item ${index} in ${activeTab} tab`);
        
        this.clearNumberInput();
    }
    
    private clearNumberInput(): void {
        this.numberInputBuffer = '';
        if (this.numberInputTimeout) {
            clearTimeout(this.numberInputTimeout);
            this.numberInputTimeout = null;
        }
        
        // Restore overlay state based on current tab's toggle state
        if (this.editorToolsPanel) {
            const activeTab = this.editorToolsPanel.getActiveTab();
            const shouldShow = this.overlayState[activeTab];
            
            if (shouldShow) {
                this.editorToolsPanel.showNumberOverlays();
            } else {
                this.editorToolsPanel.hideNumberOverlays();
            }
        }
    }
    
    // Multi-digit number selection handlers for s+number shortcut
    private selectByNumberInActiveTab(args?: string): void {
        if (!args || !this.editorToolsPanel) return;
        
        const index = parseInt(args);
        if (isNaN(index)) return;
        
        const activeTab = this.editorToolsPanel.getActiveTab();
        this.editorToolsPanel.selectByIndex(index);
        this.editorToolsPanel.hideNumberOverlays();
        this.logToConsole(`Selected item ${index} in ${activeTab} tab`);
    }
    
    private previewByNumberInActiveTab(args?: string): void {
        if (!args || !this.editorToolsPanel) return;
        
        const index = parseInt(args);
        if (isNaN(index)) return;
        
        this.editorToolsPanel.showNumberOverlays();
        const activeTab = this.editorToolsPanel.getActiveTab();
        this.logToConsole(`Preview: item ${index} in ${activeTab} tab`);
    }
    
    private cancelNumberSelection(): void {
        if (this.editorToolsPanel) {
            this.editorToolsPanel.hideNumberOverlays();
        }
        this.logToConsole('Cancelled number selection');
    }
    
    private selectPlayer(args?: string): void {
        const playerId = parseInt(args || '1');
        
        this.hidePreviewIndicator(); // Hide preview indicator when committing
        
        if (playerId >= 1 && playerId <= 4) {
            if (this.pageState) {
                this.pageState.setSelectedPlayer(playerId);
            }
            
            // Update player selector in UI
            const unitPlayerSelect = document.getElementById('unit-player-color') as HTMLSelectElement;
            if (unitPlayerSelect) {
                unitPlayerSelect.value = playerId.toString();
            }
            
            // Show toast notification
            this.showToast('Player Selected', `Player ${playerId} selected`, 'success');
        } else {
            this.showToast('Invalid Selection', `Player ${playerId} not available`, 'error');
        }
    }
    
    private selectBrushSize(args?: string): void {
        const index = parseInt(args || '1'); // 1-based index
        
        this.hidePreviewIndicator(); // Hide preview indicator when committing
        
        // World 1-based index to actual brush size values
        const brushSizeValues = [0, 1, 3, 5, 10, 15]; // Corresponds to the select options
        
        if (index >= 1 && index <= brushSizeValues.length) {
            const actualSize = brushSizeValues[index - 1];
            this.setBrushSize(actualSize);
            
            // Update brush size selector in UI and trigger onchange
            const brushSizeSelect = document.getElementById('brush-size') as HTMLSelectElement;
            if (brushSizeSelect) {
                brushSizeSelect.value = actualSize.toString();
                // Trigger the onchange event
                brushSizeSelect.dispatchEvent(new Event('change'));
            }
            
            // Show toast notification
            this.showToast('Brush Size Selected', `${BRUSH_SIZE_NAMES[index - 1]} brush selected`, 'success');
        } else {
            this.showToast('Invalid Selection', `Brush size ${index} not available`, 'error');
        }
    }
    
    private activateClearMode(): void {
        // Clear any pending number input
        this.clearNumberInput();
        
        if (this.pageState) {
            this.pageState.setPlacementMode('clear');
        }
        this.showToast('Clear Mode', 'Clear mode activated - press R to reset', 'info');
    }
    
    private resetToDefaults(): void {
        if (this.editorToolsPanel) {
            this.editorToolsPanel.hideNumberOverlays();
        }
        
        // Reset to default terrain (grass) via pageState
        if (this.pageState) {
            this.pageState.setSelectedTerrain(1);
            this.pageState.setBrushSize(0);
            this.pageState.setSelectedPlayer(1);
        }
        
        // Update UI elements
        this.setBrushTerrain(1);
        
        const brushSizeSelect = document.getElementById('brush-size') as HTMLSelectElement;
        if (brushSizeSelect) {
            brushSizeSelect.value = '0';
        }
        
        const unitPlayerSelect = document.getElementById('unit-player-color') as HTMLSelectElement;
        if (unitPlayerSelect) {
            unitPlayerSelect.value = '1';
        }
        
        // Remove selection from unit buttons
        document.querySelectorAll('.unit-button').forEach(btn => {
            btn.classList.remove('bg-blue-100', 'dark:bg-blue-900', 'border-blue-500');
        });
        
        this.logToConsole('Reset all tools to defaults');
        
        // Show toast notification
        this.showToast('Reset Complete', 'All tools reset to defaults', 'info');
    }
    
    // Note: Unit button selection is now handled by EditorToolsPanel internally

    // Reference image methods moved to ReferenceImagePanel - no longer needed here

    // Public methods for Phaser panel (for backward compatibility with UI)
    public initializePhaser(): void {
        this.logToConsole('Phaser initialization now handled by PhaserEditorComponent in dockview');
    }
    
    // Old EventBus handlers removed - components now use pageState directly
    
    private async handlePhaserReady() {
        console.log('WorldEditorPage: Phaser ready via EventBus');
        this.logToConsole('EventBus: Phaser editor is ready');
        
        // Load world data if available
        if (this.world && this.phaserEditorComponent) {
            console.log('WorldEditorPage: Loading world data into Phaser editor');
            // Give Phaser time to fully initialize webgl context and scene
            await this.phaserEditorComponent.loadWorld(this.world);
            this.hasPendingWorldDataLoad = false;
            this.refreshTileStats();
        }
    }
}

// Initialize the editor when DOM is ready
document.addEventListener('DOMContentLoaded', async () => {
    console.log('DOM loaded, starting WorldEditorPage initialization...');

    // Create page instance (just basic setup)
    const page = new WorldEditorPage("WorldEditorPage");
    
    // Create lifecycle controller with debug logging
    const lifecycleController = new LifecycleController(page.eventBus, LifecycleController.DefaultConfig)
    
    // Start breadth-first initialization
    await lifecycleController.initializeFromRoot(page);
    
    console.log('WorldEditorPage fully initialized via LifecycleController');
});
