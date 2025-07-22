import { BasePage } from './BasePage';
import { DockviewApi, DockviewComponent } from 'dockview-core';
import { PhaserEditorComponent } from './PhaserEditorComponent';
import { TileStatsPanel } from './TileStatsPanel';
import { KeyboardShortcutManager, ShortcutConfig, KeyboardState } from './KeyboardShortcutManager';
import { shouldIgnoreShortcut } from './DOMUtils';
import { World, WorldObserver, WorldEvent, WorldEventType, TilesChangedEventData, UnitsChangedEventData, WorldLoadedEventData } from './World';
import { WorldEditorPageState, PageStateObserver, PageStateEvent, PageStateEventType, ToolStateChangedEventData, VisualStateChangedEventData, WorkflowStateChangedEventData, ToolState } from './WorldEditorPageState';
import { EventBus, EditorEventTypes, TerrainSelectedPayload, UnitSelectedPayload, BrushSizeChangedPayload, PlacementModeChangedPayload, PlayerChangedPayload, TileClickedPayload, PhaserReadyPayload, GridSetVisibilityPayload, CoordinatesSetVisibilityPayload } from './EventBus';
import { EditorToolsPanel } from './EditorToolsPanel';
import { ReferenceImagePanel } from './ReferenceImagePanel';
import { ComponentLifecycle } from './ComponentLifecycle';
import { LifecycleController } from './LifecycleController';

const BRUSH_SIZE_NAMES = ['Single (1 hex)', 'Small (3 hexes)', 'Medium (5 hexes)', 'Large (9 hexes)', 'X-Large (15 hexes)', 'XX-Large (25 hexes)'];

/**
 * World Editor page with unified World architecture and centralized page state
 * Now implements ComponentLifecycle for breadth-first initialization
 */
class WorldEditorPage extends BasePage implements WorldObserver, PageStateObserver {
    private world: World | null = null;
    private pageState: WorldEditorPageState;
    private editorOutput: HTMLElement | null = null;

    // Dockview interface
    private dockview: DockviewApi | null = null;
    
    // Phaser editor component for world editing
    private phaserEditorComponent: PhaserEditorComponent | null = null;
    
    // TileStats panel for displaying statistics
    private tileStatsPanel: TileStatsPanel | null = null;
    
    // Editor tools panel for terrain/unit selection
    private editorToolsPanel: EditorToolsPanel | null = null;
    
    // Reference image panel for reference image controls
    private referenceImagePanel: ReferenceImagePanel | null = null;

    // Keyboard shortcut manager
    private keyboardShortcutManager: KeyboardShortcutManager | null = null;
    
    // Lifecycle controller for managing component initialization
    private lifecycleController: LifecycleController | null = null;

    // State management for undo/restore operations
    // Simplified state backup for preview/cancel functionality
    private savedToolState: ToolState | null = null;

    // UI state  
    private hasPendingWorldDataLoad: boolean = false;

    constructor() {
        super();
        // Basic setup only - detailed initialization moved to lifecycle phases
        this.pageState = new WorldEditorPageState();
        this.pageState.subscribe(this); // Subscribe to page state changes
        this.loadInitialState();
        this.subscribeToEditorEvents();
        
        // Initialize the lifecycle controller and start component initialization
        this.initializeWithLifecycleController();
    }
    
    /**
     * Initialize the page using the new lifecycle controller
     */
    private async initializeWithLifecycleController(): Promise<void> {
        try {
            // Create lifecycle controller with debug logging
            this.lifecycleController = new LifecycleController({
                enableDebugLogging: true,
                phaseTimeoutMs: 15000, // Increased timeout for complex initialization
                continueOnError: false // Fail fast for debugging
            });
            
            // Set up lifecycle event logging
            this.lifecycleController.onLifecycleEvent((event) => {
                console.log(`[Lifecycle] ${event.type}: ${event.componentName} - ${event.phase}`, event.error || '');
            });
            
            // Dependencies are set directly using explicit setters in initializeDOM phase
            
            // Start breadth-first initialization
            await this.lifecycleController.initializeFromRoot(this, 'WorldEditorPage');
            
            console.log('WorldEditorPage initialization complete via LifecycleController');
            
        } catch (error) {
            console.error('WorldEditorPage lifecycle initialization failed:', error);
            // Fallback to old initialization method if needed
            this.fallbackInitialization();
        }
    }
    
    /**
     * Fallback initialization if lifecycle controller fails
     */
    private fallbackInitialization(): void {
        console.warn('Using fallback initialization method');
        this.initializeSpecificComponents();
        this.initializeDockview();
        this.bindSpecificEvents();
        this.initializeKeyboardShortcuts();
        this.setupUnsavedChangesWarning();
    }
    
    // ComponentLifecycle implementation
    
    /**
     * Phase 1: Initialize DOM and discover child components
     */
    public initializeDOM(): ComponentLifecycle[] {
        try {
            console.log('WorldEditorPage: Starting DOM initialization phase');
            
            // Initialize basic components first
            this.initializeSpecificComponents();
            
            // Create child components that implement ComponentLifecycle
            const childComponents: ComponentLifecycle[] = [];
            
            // Create ReferenceImagePanel as a lifecycle-managed component
            const referenceTemplate = document.getElementById('reference-image-panel-template');
            if (referenceTemplate) {
                const referenceContainer = referenceTemplate.cloneNode(true) as HTMLElement;
                referenceContainer.style.display = 'block';
                this.referenceImagePanel = new ReferenceImagePanel(referenceContainer, this.eventBus, true);
                
                // Set dependencies directly using explicit setters
                this.referenceImagePanel.setToastCallback((title: string, message: string, type: 'success' | 'error' | 'info') => {
                    this.showToast(title, message, type);
                });
                
                // PhaserEditorComponent communication via EventBus - no direct dependency needed
                
                childComponents.push(this.referenceImagePanel);
                console.log('WorldEditorPage: Created ReferenceImagePanel child component with dependencies');
            }
            
            console.log(`WorldEditorPage: DOM initialization complete, discovered ${childComponents.length} child components`);
            return childComponents;
            
        } catch (error) {
            console.error('WorldEditorPage: DOM initialization failed:', error);
            throw error;
        }
    }
    
    /**
     * Phase 2: Inject dependencies from lifecycle controller
     */
    public injectDependencies(deps: Record<string, any>): void {
        console.log('WorldEditorPage: Injecting dependencies:', Object.keys(deps));
        
        // WorldEditorPage doesn't need any specific dependencies from other components
        // It provides dependencies to child components instead
        
        // Store a reference to the lifecycle controller if provided
        if (deps.lifecycleController) {
            console.log('WorldEditorPage: Lifecycle controller reference injected');
        }
        
        console.log('WorldEditorPage: Dependencies injection complete');
    }
    
    /**
     * Phase 3: Activate the component when all dependencies are ready
     */
    public activate(): void {
        try {
            console.log('WorldEditorPage: Starting activation phase');
            
            // Bind events now that all components are ready
            this.bindSpecificEvents();
            this.initializeKeyboardShortcuts();
            this.setupUnsavedChangesWarning();
            
            // Set cross-component dependencies now that all components are created
            this.setupCrossComponentDependencies();
            
            // Initialize dockview now that all child components are ready
            this.initializeDockview();
            
            // Update UI state
            this.updateEditorStatus('Ready');
            
            console.log('WorldEditorPage: Activation complete');
            
        } catch (error) {
            console.error('WorldEditorPage: Activation failed:', error);
            throw error;
        }
    }
    
    /**
     * Set up dependencies between components that require each other
     * 
     * Note: Using EventBus communication for loose coupling instead of direct dependencies
     * Components communicate via events rather than direct method calls
     */
    private setupCrossComponentDependencies(): void {
        // ReferenceImagePanel and PhaserEditorComponent communicate via EventBus
        // No direct dependencies needed - they remain decoupled
        
        console.log('WorldEditorPage: Components use EventBus communication - no direct dependencies needed');
    }
    
    /**
     * Phase 4: Deactivate and cleanup
     */
    public deactivate(): void {
        console.log('WorldEditorPage: Starting deactivation');
        
        // Use existing destroy method for cleanup
        this.destroy();
        
        console.log('WorldEditorPage: Deactivation complete');
    }
    
    // Dependencies are set directly using explicit setters - no ComponentDependencyDeclaration needed
    
    // WorldObserver implementation
    public onWorldEvent(event: WorldEvent): void {
        switch (event.type) {
            case WorldEventType.WORLD_LOADED:
                const loadedData = event.data as WorldLoadedEventData;
                this.updateEditorStatus('Loaded');
                this.updateSaveButtonState();
                break;
                
            case WorldEventType.WORLD_SAVED:
                this.updateEditorStatus('Saved');
                this.updateSaveButtonState();
                if (event.data.success && event.data.worldId) {
                    // Update URL if this was a new world
                    if (this.world?.getIsNewWorld()) {
                        history.replaceState(null, '', `/worlds/${event.data.worldId}/edit`);
                    }
                }
                break;
                
            case WorldEventType.TILES_CHANGED:
            case WorldEventType.UNITS_CHANGED:
                // World data changed, update UI state
                this.updateSaveButtonState();
                break;
                
            case WorldEventType.WORLD_CLEARED:
                this.updateSaveButtonState();
                break;
                
            case WorldEventType.WORLD_METADATA_CHANGED:
                this.updateSaveButtonState();
                break;
        }
    }
    
    // PageStateObserver implementation
    public onPageStateEvent(event: PageStateEvent): void {
        switch (event.type) {
            case PageStateEventType.TOOL_STATE_CHANGED:
                // Tool state changes are handled by components that need them
                // WorldEditorPage mainly coordinates but doesn't need to react to tool changes
                this.logToConsole(`Tool state changed: ${JSON.stringify(event.data)}`);
                break;
                
            case PageStateEventType.VISUAL_STATE_CHANGED:
                // Visual state changes might affect display
                this.logToConsole(`Visual state changed: ${JSON.stringify(event.data)}`);
                break;
                
            case PageStateEventType.WORKFLOW_STATE_CHANGED:
                // Workflow state changes affect the overall page flow
                this.logToConsole(`Workflow state changed: ${JSON.stringify(event.data)}`);
                break;
        }
    }
    
    /**
     * Subscribe to editor-specific events before components are created
     * This prevents race conditions where components emit events before subscribers are ready
     */
    private subscribeToEditorEvents(): void {
        console.log('WorldEditorPage: Subscribing to editor events');
        
        // Note: Tool state changes now handled via PageState Observer pattern
        // EditorToolsPanel directly updates pageState, which notifies observers
        
        // Subscribe to tile clicks from Phaser
        this.eventBus.subscribe<TileClickedPayload>(EditorEventTypes.TILE_CLICKED, (payload) => {
            this.handlePhaserTileClick(payload.data.q, payload.data.r);
        }, 'world-editor-page');
        
        // Subscribe to Phaser ready event
        this.eventBus.subscribe(EditorEventTypes.PHASER_READY, () => {
            this.handlePhaserReady();
        }, 'world-editor-page');
        
        // World changes are automatically tracked by World class via Observer pattern
        
        console.log('WorldEditorPage: Editor event subscriptions complete');
    }

    protected initializeSpecificComponents(): void {
        const worldIdInput = document.getElementById("worldIdInput") as HTMLInputElement | null;
        const isNewWorldInput = document.getElementById("isNewWorld") as HTMLInputElement | null;
        
        // World ID and new world state are now handled by the World instance

        this.editorOutput = document.getElementById('editor-output');

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


        // NOTE: Terrain/unit button bindings, player selection, and brush size controls 
        // are now handled by EditorToolsPanel component via EventBus

        // Visual options (grid/coordinates) now bound in bindPhaserPanelEvents



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


    private loadInitialState(): void {
        // Theme button state is handled by BasePage
        this.updateEditorStatus('Initializing...');

        // Read initial state from DOM
        const worldIdInput = document.getElementById("worldIdInput") as HTMLInputElement | null;
        const isNewWorldInput = document.getElementById("isNewWorld") as HTMLInputElement | null;
        
        const worldId = worldIdInput?.value.trim() || null;
        const isNewWorld = isNewWorldInput?.value === "true";

        // Create World instance and subscribe to events
        this.world = new World('New World', 8, 8);
        this.world.subscribe(this);
        
        if (!isNewWorld && worldId) {
            // Load existing world
            this.world.setWorldId(worldId);
            this.loadExistingWorld(worldId);
        } else {
            // Initialize new world
            this.initializeNewWorld();
        }
        
        // Phaser component initialization will be handled by dockview when the component is created
    }


    private initializeNewWorld(): void {
        // Try to load template world data from hidden element first
        try {
            this.world!.loadFromElement('world-data-json');
            this.hasPendingWorldDataLoad = true;
        } catch (error) {
            // No template data, world is already initialized as empty
            console.log('No template data found, using empty world');
            this.hasPendingWorldDataLoad = true;
        }
        
        this.updateEditorStatus('New World');
    }

    private async loadExistingWorld(worldId: string): Promise<void> {
        try {
            await this.world!.load(worldId);
            this.hasPendingWorldDataLoad = true;
        } catch (error) {
            console.error('Failed to load world:', error);
            this.logToConsole(`Failed to load world: ${error}`);
            this.updateEditorStatus('Load Error');
        }
    }
    
    /**
     * Load world data from hidden element in the HTML
     */
    
    /**
     * Load world data (tiles and units) into the Phaser scene
     */
    private async loadWorldDataIntoPhaser(): Promise<void> {
        
        if (!this.phaserEditorComponent || !this.phaserEditorComponent.getIsInitialized() || !this.world) {
            return;
        }
        
        try {
            // Load tiles first using setTilesData for better performance
            const allTiles = this.world.getAllTiles();
            if (allTiles.length > 0) {
                const tilesArray: Array<{ q: number; r: number; terrain: number; color: number }> = [];
                allTiles.forEach(tile => {
                    tilesArray.push({
                        q: tile.q,
                        r: tile.r,
                        terrain: tile.tileType,
                        color: tile.playerId || 0 // Use the player ID from the tile data
                    });
                });
                
                if (tilesArray.length > 0) {
                    await this.phaserEditorComponent.setTilesData(tilesArray);
                }
            }
            
            // Load units AFTER tiles are loaded - ensure proper rendering order
            const allUnits = this.world.getAllUnits();
            if (allUnits.length > 0) {
                let unitsLoaded = 0;
                
                // Add delay to ensure tiles are rendered first and textures are loaded
                setTimeout(() => {
                    allUnits.forEach((unit) => {
                        
                        // Paint unit in Phaser (units render above tiles due to depth=10)
                        const success = this.phaserEditorComponent!.paintUnit(unit.q, unit.r, unit.unitType, unit.playerId);
                        if (success) {
                            unitsLoaded++;
                        } else {
                        }
                    });
                    
                    // Refresh tile stats after all loading is complete
                    this.refreshTileStats();
                    
                    // Center camera on the loaded world
                    this.centerCameraOnWorld();
                }, 300); // Increased delay to ensure tiles are rendered first
            } else {
                // No units to load, refresh stats immediately
                this.refreshTileStats();
                
                // Center camera on the loaded world
                this.centerCameraOnWorld();
            }
            
        } catch (error) {
            console.error('Error loading world data into Phaser:', error);
            this.logToConsole(`Error loading into Phaser: ${error}`);
        }
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
            'world-editor-page'
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
            'world-editor-page'
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
        
        if (this.phaserEditorComponent && this.phaserEditorComponent.getIsInitialized()) {
            const tiles = this.phaserEditorComponent.getTilesData();
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

        try {
            // World now handles its own export operations
            const result = await this.world.save();
            
            if (result.success) {
                this.showToast('Success', 'World exported successfully', 'success');
            } else {
                this.showToast('Error', result.error || 'Failed to export world', 'error');
            }
        } catch (error) {
            console.error('Export failed:', error);
            this.showToast('Error', 'Export failed', 'error');
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
            const terrainNames = ['Unknown', 'Grass', 'Desert', 'Water', 'Mountain', 'Rock'];
            const currentTerrain = this.pageState?.getToolState().selectedTerrain || 1;
            const currentBrushSize = this.pageState?.getToolState().brushSize || 0;
            brushInfo.textContent = `Current: ${terrainNames[currentTerrain]}, ${BRUSH_SIZE_NAMES[currentBrushSize]}`;
        }
    }

    // Note: Terrain button selection is now handled by EditorToolsPanel internally

    // Theme management is handled by BasePage

    // Dockview panel creation methods
    private createToolsComponent() {
        const template = document.getElementById('tools-panel-template');
        if (!template) {
            console.error('Tools panel template not found');
            return { element: document.createElement('div'), init: () => {}, dispose: () => {} };
        }
        
        const container = template.cloneNode(true) as HTMLElement;
        container.style.display = 'block';
        container.style.width = '100%';
        container.style.height = '100%';
        
        return {
            element: container,
            init: async () => {
                // Initialize EditorToolsPanel component using new lifecycle
                await this.initializeEditorToolsPanelLifecycle(container);
            },
            dispose: () => {
                // Clean up EditorToolsPanel
                if (this.editorToolsPanel) {
                    this.editorToolsPanel.destroy();
                    this.editorToolsPanel = null;
                }
            }
        };
    }

    private createPhaserComponent() {
        const template = document.getElementById('canvas-panel-template');
        if (!template) {
            console.error('Phaser panel template not found');
            return { element: document.createElement('div'), init: () => {}, dispose: () => {} };
        }
        
        const container = template.cloneNode(true) as HTMLElement;
        container.style.display = 'block';
        container.style.width = '100%';
        container.style.height = '100%';
        
        return {
            element: container,
            init: () => {
                // Initialize PhaserEditorComponent
                this.phaserEditorComponent = new PhaserEditorComponent(container, this.eventBus, this.pageState, this.world, true);
                this.logToConsole('PhaserEditorComponent initialized');
                
                // Bind grid and coordinates checkboxes now that the template is in the DOM
                this.bindPhaserPanelEvents(container);
            },
            dispose: () => {
                if (this.phaserEditorComponent) {
                    this.phaserEditorComponent.destroy();
                    this.phaserEditorComponent = null;
                }
            }
        };
    }

    private createTileStatsComponent() {
        // Create a container for the TileStats panel
        const container = document.createElement('div');
        container.id = 'tilestats-container';
        container.style.width = '100%';
        container.style.height = '100%';
        
        return {
            element: container,
            init: async () => {
                // Initialize TileStatsPanel component using new lifecycle
                await this.initializeTileStatsPanelLifecycle(container);
            },
            dispose: () => {
                if (this.tileStatsPanel) {
                    this.tileStatsPanel.destroy();
                    this.tileStatsPanel = null;
                }
            }
        };
    }

    private createConsoleComponent() {
        const template = document.getElementById('console-panel-template');
        if (!template) {
            console.error('Console panel template not found');
            return { element: document.createElement('div'), init: () => {}, dispose: () => {} };
        }
        
        const container = template.cloneNode(true) as HTMLElement;
        container.style.display = 'block';
        container.style.width = '100%';
        container.style.height = '100%';
        
        return {
            element: container,
            init: () => {
                // Find the editor output element within this cloned template
                const outputElement = container.querySelector('#editor-output');
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
        
        const container = template.cloneNode(true) as HTMLElement;
        container.style.display = 'block';
        container.style.width = '100%';
        container.style.height = '100%';
        
        return {
            element: container,
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
        
        // Fallback to template-based creation if lifecycle panel not available
        const template = document.getElementById('reference-image-panel-template');
        if (!template) {
            console.error('Reference image panel template not found');
            return { element: document.createElement('div'), init: () => {}, dispose: () => {} };
        }
        
        const container = template.cloneNode(true) as HTMLElement;
        container.style.display = 'block';
        container.style.width = '100%';
        container.style.height = '100%';
        
        return {
            element: container,
            init: () => {
                console.log('ReferenceImagePanel dockview component initialized (fallback mode)');
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

        try {
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
        } catch (error) {
            this.logToConsole(`Failed to set panel sizes: ${error}`);
        }
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
    
    
    
    
    
    

    
    
    /**
     * Get unit data at the given coordinates (returns null if no unit exists)
     */
    private getUnitAt(q: number, r: number): { unitType: number; playerId: number } | null {
        if (!this.phaserEditorComponent || !this.phaserEditorComponent.getIsInitialized()) {
            return null;
        }
        
        const unitsData = this.phaserEditorComponent.getUnitsData();
        const unit = unitsData.find(unit => unit.q === q && unit.r === r);
        return unit ? { unitType: unit.unitType, playerId: unit.playerId } : null;
    }
    
    /**
     * Set unit at the given coordinates - Observer pattern handles Phaser updates
     */
    private setUnitAt(q: number, r: number, unitType: number, playerId: number): void {
        // Update world data - Observer pattern will handle Phaser updates
        if (this.world) {
            this.world.setUnitAt(q, r, unitType, playerId);
        }
    }
    
    
    // EditorToolsPanel methods
    private async initializeEditorToolsPanelLifecycle(container: HTMLElement): Promise<void> {
        try {
            this.logToConsole('Initializing EditorToolsPanel with lifecycle...');
            
            // Create EditorToolsPanel component
            this.editorToolsPanel = new EditorToolsPanel(container, this.eventBus, true);
            
            // Phase 1: Initialize DOM
            await this.editorToolsPanel.initializeDOM();
            
            // Phase 2: Set dependencies directly using explicit setters
            this.editorToolsPanel.setPageState(this.pageState);
            await this.editorToolsPanel.injectDependencies({});
            
            // Phase 3: Activate component
            await this.editorToolsPanel.activate();
            
            this.logToConsole('EditorToolsPanel initialized with lifecycle architecture');
            
        } catch (error) {
            this.logToConsole(`Failed to initialize EditorToolsPanel with lifecycle: ${error}`);
        }
    }
    
    // Legacy method for backward compatibility
    private initializeEditorToolsPanel(container: HTMLElement): void {
        try {
            this.logToConsole('Initializing EditorToolsPanel...');
            
            // Create EditorToolsPanel component
            this.editorToolsPanel = new EditorToolsPanel(container, this.eventBus, true);
            
            // Inject page state so EditorToolsPanel can generate state changes
            this.editorToolsPanel.setPageState(this.pageState);
            
            this.logToConsole('EditorToolsPanel initialized with page state');
            
        } catch (error) {
            this.logToConsole(`Failed to initialize EditorToolsPanel: ${error}`);
        }
    }
    
    // TileStats panel methods
    private async initializeTileStatsPanelLifecycle(container: HTMLElement): Promise<void> {
        try {
            this.logToConsole('Initializing TileStatsPanel with lifecycle...');
            
            // Create TileStatsPanel component
            this.tileStatsPanel = new TileStatsPanel(container, this.eventBus, true);
            
            // Phase 1: Initialize DOM
            await this.tileStatsPanel.initializeDOM();
            
            // Phase 2: Set dependencies directly using explicit setters
            if (this.world) {
                this.tileStatsPanel.setWorld(this.world);
            } else {
                throw new Error('World is not available for TileStatsPanel');
            }
            await this.tileStatsPanel.injectDependencies({});
            
            // Phase 3: Activate component
            await this.tileStatsPanel.activate();
            
            this.logToConsole('TileStatsPanel initialized with lifecycle architecture');
            
        } catch (error) {
            this.logToConsole(`Failed to initialize TileStatsPanel with lifecycle: ${error}`);
        }
    }
    
    // Legacy method for backward compatibility - now uses lifecycle internally
    private initializeTileStatsPanel(container: HTMLElement): void {
        // Just call the lifecycle method
        this.initializeTileStatsPanelLifecycle(container).catch(error => {
            this.logToConsole(`Failed to initialize TileStats panel: ${error}`);
        });
    }
    
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
            this.savedToolState = null;
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
    
    private handlePhaserReady(): void {
        console.log('WorldEditorPage: Phaser ready via EventBus');
        this.logToConsole('EventBus: Phaser editor is ready');
        
        // Load pending world data if available
        if (this.hasPendingWorldDataLoad) {
            console.log('WorldEditorPage: Loading pending world data');
            // Give Phaser time to fully initialize webgl context and scene
            setTimeout(() => {
                this.loadWorldDataIntoPhaser();
            }, 10);
        }
    }
    
    private handlePhaserTileClick(q: number, r: number): void {
        try {
            // Update coordinate inputs
            const rowInput = document.getElementById('paint-row') as HTMLInputElement;
            const colInput = document.getElementById('paint-col') as HTMLInputElement;
            
            if (rowInput) rowInput.value = r.toString();
            if (colInput) colInput.value = q.toString();
            
            // Log the click
            const currentMode = this.pageState?.getToolState().placementMode || 'terrain';
            this.logToConsole(`Tile clicked at Q=${q}, R=${r} in ${currentMode} mode`);
            
            // Note: Actual tile painting is now handled directly by PhaserEditorComponent
            // This method just updates UI elements that need to react to clicks
            
        } catch (error) {
            this.logToConsole(`Tile click handler error: ${error}`);
        }
    }
}

// Initialize the editor when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    new WorldEditorPage();
});
