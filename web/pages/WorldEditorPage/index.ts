import { BasePage } from '../../lib/BasePage';
import { DockviewApi, DockviewComponent } from 'dockview-core';
import { PhaserEditorComponent } from './PhaserEditorComponent';
import { TileStatsPanel } from './TileStatsPanel';
import { AssetThemePreference } from '../common/AssetThemePreference';
import { KeyboardShortcutManager, ShortcutConfig, KeyboardState } from '../../lib/KeyboardShortcutManager';
import { shouldIgnoreShortcut } from '../../lib/DOMUtils';
import { Unit, Tile, World, TilesChangedEventData, UnitsChangedEventData, WorldLoadedEventData } from '../common/World';
import { ToolState } from './WorldEditorPresenter';
import { EventBus } from '../../lib/EventBus';
import { WorldEventType, WorldEventTypes, EditorEventTypes, TerrainSelectedPayload, UnitSelectedPayload, BrushSizeChangedPayload, PlacementModeChangedPayload, PlayerChangedPayload, TileClickedPayload, PhaserReadyPayload } from '../common/events';
import { EditorToolsPanel } from './ToolsPanel';
import { ReferenceImagePanel } from './ReferenceImagePanel';
import { LCMComponent } from '../../lib/LCMComponent';
import { LifecycleController } from '../../lib/LifecycleController';
import { WorldEditorPresenter } from './WorldEditorPresenter';

/**
 * World Editor page with unified World architecture and centralized page state
 * Now implements LCMComponent for breadth-first initialization
 */
class WorldEditorPage extends BasePage {
    private world: World;
    private presenter: WorldEditorPresenter;

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
        // Create World instance early so child components can use it
        this.createWorldInstance();

        // Create presenter and initialize with world
        this.presenter = new WorldEditorPresenter(this.eventBus);
        this.presenter.initialize(this.world);

        // Register callbacks for presenter to update UI
        this.presenter.setStatusChangeCallback((status) => this.updateEditorStatus(status));
        this.presenter.setToastCallback((title, message, type) => this.showToast(title, message, type));
        this.presenter.setSaveButtonStateCallback((hasChanges) => this.updateSaveButtonStateFromPresenter(hasChanges));

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
            this.referenceImagePanel.setPresenter(this.presenter);

            childComponents.push(this.referenceImagePanel);
        }
        
        // Create EditorToolsPanel as a lifecycle-managed component using template
        const toolsTemplate = document.getElementById('tools-panel-template');
        if (toolsTemplate) {
            // Use the template element directly - it already has proper structure and styling
            this.editorToolsPanel = new EditorToolsPanel(toolsTemplate, this.eventBus, true);

            // Set dependencies directly using explicit setters
            this.editorToolsPanel.setPresenter(this.presenter);
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
        this.phaserEditorComponent.setPresenter(this.presenter);
        this.phaserEditorComponent.setWorld(this.world);

        childComponents.push(this.phaserEditorComponent);

        // Register components with presenter
        this.presenter.registerPhaserEditor(this.phaserEditorComponent);
        this.presenter.registerToolsPanel(this.editorToolsPanel);
        this.presenter.registerTileStatsPanel(this.tileStatsPanel);
        this.presenter.registerReferenceImagePanel(this.referenceImagePanel);

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
            console.log(`Failed to load world: ${error}`);
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
    
    // Dependencies are set directly using explicit setters - no ComponentDependencyDeclaration needed
    
    // World event handlers via EventBus
    private handleWorldLoaded(data: WorldLoadedEventData): void {
        this.updateEditorStatus('Loaded');
        this.updateSaveButtonState();
        // Load game configuration from world into the form
        this.loadGameConfig();
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
        this.addSubscription(WorldEventTypes.WORLD_LOADED, this);
        this.addSubscription(WorldEventTypes.WORLD_SAVED, this);
        this.addSubscription(WorldEventTypes.TILES_CHANGED, this);
        this.addSubscription(WorldEventTypes.UNITS_CHANGED, this);
        this.addSubscription(WorldEventTypes.WORLD_CLEARED, this);
        this.addSubscription(WorldEventTypes.WORLD_METADATA_CHANGED, this);
        
        // Note: Tool state changes now handled via presenter
        // EditorToolsPanel directly calls presenter methods, which updates components
        
        // Subscribe to Phaser ready event
        this.addSubscription(EditorEventTypes.PHASER_READY, this);
        
        // World changes are automatically tracked by World class via Observer pattern
    }

    /**
     * Handle events from the EventBus
     */
    public handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
        switch(eventType) {
            case WorldEventTypes.WORLD_LOADED:
                this.handleWorldLoaded(data);
                break;
            
            case WorldEventTypes.WORLD_SAVED:
                this.handleWorldSaved(data);
                break;
            
            case WorldEventTypes.TILES_CHANGED:
            case WorldEventTypes.UNITS_CHANGED:
            case WorldEventTypes.WORLD_CLEARED:
            case WorldEventTypes.WORLD_METADATA_CHANGED:
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
                    case 'gameConfig':
                        return this.createGameConfigComponent();
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

        const screenshotButton = document.getElementById('capture-screenshot-btn');
        if (screenshotButton) {
            screenshotButton.addEventListener('click', this.handleScreenshotClick.bind(this));
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
                    this.fillAllGrass();
                    break;
                case 'create-island-world':
                    this.createIslandWorld();
                    break;
                case 'create-mountain-ridge':
                    this.createMountainRidge();
                    break;
                case 'show-terrain-stats':
                    this.showTerrainStats();
                    break;
                case 'randomize-terrain':
                    this.randomizeTerrain();
                    break;
                case 'clear-world':
                    this.clearWorld();
                    break;
                case 'download-image':
                    this.downloadImage();
                    break;
                case 'download-game-data':
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
                console.log(`Grid checkbox changed to: ${checked}`);
            });
            console.log('Grid checkbox event handler bound');
        } else {
            console.log('Grid checkbox not found in Phaser panel');
        }
        
        const showCoordinatesCheckbox = container.querySelector('#show-coordinates') as HTMLInputElement;
        if (showCoordinatesCheckbox) {
            showCoordinatesCheckbox.addEventListener('change', (e) => {
                const checked = (e.target as HTMLInputElement).checked;
                this.setShowCoordinates(checked);
                console.log(`Coordinates checkbox changed to: ${checked}`);
            });
            console.log('Coordinates checkbox event handler bound');
        } else {
            console.log('Coordinates checkbox not found in Phaser panel');
        }
        
        const showHealthCheckbox = container.querySelector('#show-health') as HTMLInputElement;
        if (showHealthCheckbox) {
            showHealthCheckbox.addEventListener('change', (e) => {
                const checked = (e.target as HTMLInputElement).checked;
                this.setShowHealth(checked);
                console.log(`Health checkbox changed to: ${checked}`);
            });
            console.log('Health checkbox event handler bound');
        } else {
            console.log('Health checkbox not found in Phaser panel');
        }

        // Brush/Fill/Rectangle tool selector
        const brushSizeSelect = document.getElementById('brush-size') as HTMLSelectElement;
        const shapeFillToggle = document.getElementById('shape-fill-toggle') as HTMLLabelElement;
        const shapeFillModeCheckbox = document.getElementById('shape-fill-mode') as HTMLInputElement;

        if (brushSizeSelect) {
            brushSizeSelect.addEventListener('change', (e) => {
                const value = (e.target as HTMLSelectElement).value;

                // Shape modes map
                const shapeMode: { [key: string]: 'rectangle' | 'circle' | 'oval' | 'line' | null } = {
                    'rect': 'rectangle',
                    'circle': 'circle',
                    'oval': 'oval',
                    'line': 'line'
                };

                if (shapeMode[value]) {
                    // Shape mode (rectangle, circle, oval, line)
                    const shape = shapeMode[value]!;
                    this.setBrushSize(value, 0);
                    if (this.phaserEditorComponent && this.phaserEditorComponent.editorScene) {
                        this.phaserEditorComponent.editorScene.setShapeMode(shape);
                    }
                    // Show fill/outline toggle (except for line)
                    if (shapeFillToggle) {
                        if (shape === 'line') {
                            shapeFillToggle.classList.add('hidden');
                        } else {
                            shapeFillToggle.classList.remove('hidden');
                        }
                    }
                    console.log(`${shape.charAt(0).toUpperCase() + shape.slice(1)} mode activated (multi-click)`);
                } else if (value.startsWith('fill:')) {
                    // Fill mode - extract radius
                    const radius = parseInt(value.substring(5));
                    this.setBrushSize("fill", radius);
                    if (this.phaserEditorComponent && this.phaserEditorComponent.editorScene) {
                        this.phaserEditorComponent.editorScene.setShapeMode(null);
                    }
                    // Hide fill/outline toggle
                    if (shapeFillToggle) {
                        shapeFillToggle.classList.add('hidden');
                    }
                    console.log(`Fill mode activated with radius: ${radius}`);
                } else {
                    // Brush mode - extract size
                    const size = parseInt(value);
                    this.setBrushSize("brush", size);
                    if (this.phaserEditorComponent && this.phaserEditorComponent.editorScene) {
                        this.phaserEditorComponent.editorScene.setShapeMode(null);
                    }
                    // Hide fill/outline toggle
                    if (shapeFillToggle) {
                        shapeFillToggle.classList.add('hidden');
                    }
                    console.log(`Brush size changed to: ${size}`);
                }
            });
            console.log('Brush/Fill/Rectangle tool selector event handler bound');
        } else {
            console.log('Brush/Fill/Rectangle tool selector not found');
        }

        // Shape fill/outline toggle
        if (shapeFillModeCheckbox) {
            shapeFillModeCheckbox.addEventListener('change', (e) => {
                const isFilled = (e.target as HTMLInputElement).checked;
                if (this.phaserEditorComponent && this.phaserEditorComponent.editorScene) {
                    this.phaserEditorComponent.editorScene.setShapeFillMode(isFilled);
                }
                console.log(`Shape fill mode: ${isFilled ? 'filled' : 'outline'}`);
            });
            console.log('Shape fill/outline toggle event handler bound');
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
            
            // Brush/Fill/Shape tool selector (b + letter)
            {
                key: 'b',
                handler: (args?: string) => this.selectBrushGroup(args),
                previewHandler: (args?: string) => this.previewBrushGroup(args),
                cancelHandler: () => this.cancelBrushSelection(),
                description: 'Select brush/fill/shape tool (b=brush, f=fill, s=shapes)',
                category: 'Tools',
                requiresArgs: true,
                argType: 'string',
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
        this.presenter.selectTerrain(terrain);
        this.updateBrushInfo();
    }

    public setBrushSize(mode: string, size: number): void {
        this.presenter.setBrushSize(mode, size);
        this.updateBrushInfo();
    }

    /**
     * Setup event handlers for game config panel inputs
     */
    private setupGameConfigHandlers(): void {
        // Save button handler - saves world to disk
        const saveButton = document.getElementById('save-game-config');
        if (saveButton) {
            saveButton.addEventListener('click', () => {
                this.saveGameConfig();
                this.saveWorld();
            });
        }

        // Add input event listeners to detect changes and update game config
        const inputIds = [
            'config-num-players',
            'config-starting-coins',
            'config-game-income',
            'config-landbase-income',
            'config-navalbase-income',
            'config-airportbase-income',
            'config-missilesilo-income',
            'config-mines-income'
        ];

        inputIds.forEach(id => {
            const input = document.getElementById(id) as HTMLInputElement;
            if (input) {
                input.addEventListener('input', () => this.saveGameConfig());
            }
        });
    }

    /**
     * Load game configuration into the form inputs
     */
    private loadGameConfig(): void {
        const gameConfig = this.world?.getDefaultGameConfig();

        if (!gameConfig) {
            console.log('No game configuration found in world, using defaults');
            return;
        }

        // Set number of players
        const numPlayersInput = document.getElementById('config-num-players') as HTMLInputElement;
        if (numPlayersInput && gameConfig.players) {
            numPlayersInput.value = gameConfig.players.length.toString();
        }

        // Set income config values
        if (gameConfig.income_configs) {
            const incomeConfig = gameConfig.income_configs;

            (document.getElementById('config-starting-coins') as HTMLInputElement)!.value = incomeConfig.starting_coins?.toString() || '300';
            (document.getElementById('config-game-income') as HTMLInputElement)!.value = incomeConfig.game_income?.toString() || '0';
            (document.getElementById('config-landbase-income') as HTMLInputElement)!.value = incomeConfig.landbase_income?.toString() || '100';
            (document.getElementById('config-navalbase-income') as HTMLInputElement)!.value = incomeConfig.navalbase_income?.toString() || '150';
            (document.getElementById('config-airportbase-income') as HTMLInputElement)!.value = incomeConfig.airportbase_income?.toString() || '200';
            (document.getElementById('config-missilesilo-income') as HTMLInputElement)!.value = incomeConfig.missilesilo_income?.toString() || '300';
            (document.getElementById('config-mines-income') as HTMLInputElement)!.value = incomeConfig.mines_income?.toString() || '500';
        }

        console.log('Game configuration loaded into form');
    }

    /**
     * Save game configuration to the world
     */
    private async saveGameConfig(): Promise<void> {
        // Collect values from inputs
        const numPlayers = parseInt((document.getElementById('config-num-players') as HTMLInputElement)?.value || '2');
        const startingCoins = parseInt((document.getElementById('config-starting-coins') as HTMLInputElement)?.value || '300');
        const gameIncome = parseInt((document.getElementById('config-game-income') as HTMLInputElement)?.value || '0');
        const landbaseIncome = parseInt((document.getElementById('config-landbase-income') as HTMLInputElement)?.value || '100');
        const navalbaseIncome = parseInt((document.getElementById('config-navalbase-income') as HTMLInputElement)?.value || '150');
        const airportbaseIncome = parseInt((document.getElementById('config-airportbase-income') as HTMLInputElement)?.value || '200');
        const missilesiloIncome = parseInt((document.getElementById('config-missilesilo-income') as HTMLInputElement)?.value || '300');
        const minesIncome = parseInt((document.getElementById('config-mines-income') as HTMLInputElement)?.value || '500');

        // Create GameConfiguration object
        const gameConfig = {
            players: Array.from({ length: numPlayers }, (_, i) => ({
                player_id: i + 1,
                player_type: 'human',
                color: `player${i + 1}`,
                team_id: 0,
                name: `Player ${i + 1}`,
                is_active: true,
                starting_coins: startingCoins,
                coins: startingCoins
            })),
            teams: [],
            income_configs: {
                starting_coins: startingCoins,
                game_income: gameIncome,
                landbase_income: landbaseIncome,
                navalbase_income: navalbaseIncome,
                airportbase_income: airportbaseIncome,
                missilesilo_income: missilesiloIncome,
                mines_income: minesIncome
            },
            settings: {}
        };

        // Update world object with new config
        if (this.world) {
            this.world.setDefaultGameConfig(gameConfig);
        }

        console.log('Game configuration updated:', gameConfig);
    }
    
    public setShowGrid(showGrid: boolean): void {
        this.presenter.setShowGrid(showGrid);
        console.log(`Grid visibility set to: ${showGrid}`);
    }

    public setShowCoordinates(showCoordinates: boolean): void {
        this.presenter.setShowCoordinates(showCoordinates);
        console.log(`Coordinates visibility set to: ${showCoordinates}`);
    }

    public setShowHealth(showHealth: boolean): void {
        this.presenter.setShowHealth(showHealth);
        console.log(`Health visibility set to: ${showHealth}`);
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
            console.log('Filled world with grass via World Observer pattern');
        } else {
            console.log('World not available, cannot fill grass');
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
            console.log('World or Phaser panel not available, cannot create mountain ridge');
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
            console.log(`Total tiles: ${tiles.length}`);
        } else {
        }
       */
    }

    public randomizeTerrain(): void {
        console.log('Randomizing terrain...');
        
        if (this.phaserEditorComponent && this.phaserEditorComponent.getIsInitialized()) {
            this.phaserEditorComponent.randomizeTerrain();
            console.log('Terrain randomized using Phaser');
        } else {
            console.log('Phaser panel not available, cannot randomize terrain');
        }
    }

    public clearWorld(): void {
        console.log('Clearing entire world...');
        
        // Clear world data - this will trigger observer notifications to update Phaser
        if (this.world) {
            this.world.clearAll();
            console.log('World data cleared - Phaser will update via observer pattern');
        } else {
            console.log('World instance not available');
        }
        
        // Show success message
        this.showToast('World Cleared', 'All tiles and units have been removed', 'info');
        console.log('Clear world operation completed');
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
            console.log(`Save failed: ${error}`);
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

    private async handleScreenshotClick(): Promise<void> {
        const worldId = this.world?.getWorldId();
        if (!worldId) {
            console.error('No world ID available');
            this.showToast('Error', 'No world ID available', 'error');
            return;
        }

        if (!this.phaserEditorComponent || !this.phaserEditorComponent.editorScene) {
            this.showToast('Error', 'Editor scene not initialized', 'error');
            return;
        }

        try {
            // Capture screenshot from Phaser scene
            const blob = await this.phaserEditorComponent.editorScene.captureScreenshotAsync('image/png', 0.92);

            if (!blob) {
                this.showToast('Error', 'Failed to capture screenshot', 'error');
                return;
            }

            // Upload to server
            const formData = new FormData();
            formData.append('screenshot', blob, 'screenshot.png');

            const themeName = AssetThemePreference.get()
            const previewUrl = `/worlds/${worldId}/screenshots/${themeName}`;
            const response = await fetch(previewUrl, {
                method: 'POST',
                body: formData
            });

            if (response.ok) {
                this.showToast('Success', 'Screenshot saved successfully', 'success');
                this.addWorldPreviewUrl(previewUrl)
            } else {
                this.showToast('Error', 'Failed to save screenshot', 'error');
            }
        } catch (error) {
            console.error('Screenshot error:', error);
            this.showToast('Error', 'Failed to capture or save screenshot', 'error');
        }
    }

    private async addWorldPreviewUrl(previewUrl: string): Promise<void> {
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
            console.log(`Updating world title to: ${newTitle}`);
            
            // Save the world (this will include the title update)
            await this.saveWorld();
            
            this.showToast('Success', 'World title updated', 'success');
            
        } catch (error) {
            console.error('Failed to save world title:', error);
            console.log(`Failed to save world title: ${error}`);
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

    /**
     * Get brush size values from the select dropdown
     */
    private getBrushSizeValues(): number[] {
        const brushSizeSelect = document.getElementById('brush-size') as HTMLSelectElement;
        if (!brushSizeSelect) return [0];

        const values: number[] = [];
        for (let i = 0; i < brushSizeSelect.options.length; i++) {
            values.push(parseInt(brushSizeSelect.options[i].value));
        }
        return values;
    }

    /**
     * Get brush size display name from the select dropdown
     */
    private getBrushSizeName(brushSize: number): string {
        const brushSizeSelect = document.getElementById('brush-size') as HTMLSelectElement;
        if (!brushSizeSelect) return 'Unknown';

        for (let i = 0; i < brushSizeSelect.options.length; i++) {
            if (parseInt(brushSizeSelect.options[i].value) === brushSize) {
                return brushSizeSelect.options[i].text;
            }
        }
        return 'Unknown';
    }

    private updateBrushInfo(): void {
        const brushInfo = document.getElementById('brush-info');
        if (brushInfo) {
            const toolState = this.presenter.getToolState();
            const currentTerrain = toolState.selectedTerrain;
            const currentBrushSize = toolState.brushSize;
            const brushSizeName = this.getBrushSizeName(currentBrushSize);
            brushInfo.textContent = `Current: Terrain ${currentTerrain}, ${brushSizeName}`;
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
                console.log('PhaserEditorComponent initialized using template directly');
                
                // Bind grid and coordinates checkboxes
                this.bindPhaserPanelEvents(template);
            },
            dispose: () => {
                if (this.phaserEditorComponent) {
                    this.phaserEditorComponent.deactivate();
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

    private createGameConfigComponent() {
        const template = document.getElementById('game-config-panel-template');
        if (!template) {
            console.error('Game config panel template not found');
            return { element: document.createElement('div'), init: () => {}, dispose: () => {} };
        }

        // Use the template element directly - no cloning needed
        template.style.display = 'block';

        return {
            element: template,
            init: () => {
                // Setup event handlers for game config inputs
                this.setupGameConfigHandlers();
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

        // Add game config panel to the right of Phaser (260px width)
        this.dockview.addPanel({
            id: 'gameConfig',
            component: 'gameConfig',
            title: '⚙️ Game Configuration',
            position: { direction: 'right', referencePanel: 'phaser' }
        });

        // Add TileStats panel below the Game Config panel
        this.dockview.addPanel({
            id: 'tilestats',
            component: 'tilestats',
            title: '📊 World Statistics',
            position: { direction: 'below', referencePanel: 'gameConfig' }
        });
        
        // Add Reference Image panel below the TileStats panel
        this.dockview.addPanel({
            id: 'referenceImage',
            component: 'referenceImage',
            title: '🖼️ Reference Image',
            position: { direction: 'below', referencePanel: 'tilestats' }
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

            // Set right panel (Game Config) to 260px width
            const gameConfigPanel = this.dockview.getPanel('gameConfig');
            if (gameConfigPanel) {
                gameConfigPanel.api.setSize({ width: 260 });
            }

            // Set reference image panel to 300px height to accommodate controls
            const referenceImagePanel = this.dockview.getPanel('referenceImage');
            if (referenceImagePanel) {
                referenceImagePanel.api.setSize({ height: 300 });
            }

            console.log('Panel sizes set: Tools=270px, Advanced=260px, ReferenceImage=300px, World Editor=remaining');
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

    // Callback from presenter for save button state changes
    private updateSaveButtonStateFromPresenter(hasChanges: boolean): void {
        const saveButton = document.getElementById('save-world-btn');
        if (saveButton) {
            if (hasChanges) {
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
    
    // Phaser panel methods
    // OLD METHOD REMOVED: initializePhaserPanel - now handled by PhaserEditorComponent
    
    /**
     * Center the camera on the loaded world by calculating bounds and focusing on center
     */
    private centerCameraOnWorld(): void {
        if (!this.phaserEditorComponent || !this.phaserEditorComponent.getIsInitialized() || !this.world) {
            console.log('Cannot center camera - components not ready');
            return;
        }
        
        const allTiles = this.world.getAllTiles();
        if (allTiles.length === 0) {
            console.log('No tiles to center camera on');
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
        
        console.log(`Centering camera on Q=${centerQ}, R=${centerR} (bounds: Q=${minQ}-${maxQ}, R=${minR}-${maxR})`);
        
        // Center the camera using the PhaserEditorComponent's method
        this.phaserEditorComponent.centerCamera(centerQ, centerR);
    }
    
    // Simplified state management for preview/cancel functionality
    private saveUIState(): void {
        this.savedToolState = { ...this.presenter.getToolState() };
    }

    private restoreUIState(): void {
        if (!this.savedToolState) return;

        this.presenter.selectTerrain(this.savedToolState.selectedTerrain);
        this.presenter.selectUnit(this.savedToolState.selectedUnit);
        this.presenter.selectPlayer(this.savedToolState.selectedPlayer);
        this.presenter.setBrushSize(this.savedToolState.brushMode, this.savedToolState.brushSize);

        // Clear saved state
        this.savedToolState = null as any;
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
            this.presenter.selectTerrain(0);
            this.showPreviewIndicator('Preview: Clear mode');
            return;
        }

        // Use visual index worldping
        const terrainId = this.getTerrainIdByNatureIndex(index);
        const terrainName = this.getTerrainNameByNatureIndex(index);

        if (terrainId !== null) {
            this.saveUIState();
            this.setBrushTerrain(terrainId);
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
            this.presenter.selectUnit(unitId);
            const currentPlayer = this.presenter.getToolState().selectedPlayer;
            this.showPreviewIndicator(`Preview: ${unitName} for player ${currentPlayer}`);
        }
    }

    private previewPlayer(args?: string): void {
        const playerId = parseInt(args || '1');

        if (playerId >= 1 && playerId <= 4) {
            this.saveUIState();
            this.presenter.selectPlayer(playerId);

            const unitPlayerSelect = document.getElementById('unit-player-color') as HTMLSelectElement;
            if (unitPlayerSelect) {
                unitPlayerSelect.value = playerId.toString();
            }

            this.showPreviewIndicator(`Preview: Player ${playerId} selected`);
        }
    }
    
    private previewBrushSize(args?: string): void {
        const index = parseInt(args || '1'); // 1-based index

        const brushSizeSelect = document.getElementById('brush-size') as HTMLSelectElement;
        if (!brushSizeSelect || index < 1 || index > brushSizeSelect.options.length) {
            return;
        }

        this.saveUIState();
        const option = brushSizeSelect.options[index - 1];
        const value = option.value;

        if (value.startsWith('fill:')) {
            const radius = parseInt(value.substring(5));
            this.setBrushSize("fill", radius);
            brushSizeSelect.value = value;
            this.showPreviewIndicator(`Preview: ${option.text}`);
        } else {
            const size = parseInt(value);
            this.setBrushSize("brush", size);
            brushSizeSelect.value = value;
            this.showPreviewIndicator(`Preview: ${option.text}`);
        }
    }
    
    private cancelSelection(): void {
        this.restoreUIState();
        this.hidePreviewIndicator();
        console.log('Selection cancelled - state restored');
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
            console.log('Switched to Nature terrain tab');
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
            console.log('Switched to City terrain tab');
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
            console.log('Switched to Unit tab');
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
        
        console.log(`${tab.charAt(0).toUpperCase() + tab.slice(1)} tab overlay ${visible ? 'shown' : 'hidden'}`);
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
        
        console.log(`Number input: ${this.numberInputBuffer}`);
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
                
                console.log(`Number input: ${this.numberInputBuffer}`);
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
        console.log(`Selected item ${index} in ${activeTab} tab`);
        
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
        console.log(`Selected item ${index} in ${activeTab} tab`);
    }
    
    private previewByNumberInActiveTab(args?: string): void {
        if (!args || !this.editorToolsPanel) return;
        
        const index = parseInt(args);
        if (isNaN(index)) return;
        
        this.editorToolsPanel.showNumberOverlays();
        const activeTab = this.editorToolsPanel.getActiveTab();
        console.log(`Preview: item ${index} in ${activeTab} tab`);
    }
    
    private cancelNumberSelection(): void {
        if (this.editorToolsPanel) {
            this.editorToolsPanel.hideNumberOverlays();
        }
        console.log('Cancelled number selection');
    }
    
    private selectPlayer(args?: string): void {
        const playerId = parseInt(args || '1');
        
        this.hidePreviewIndicator(); // Hide preview indicator when committing
        
        if (playerId >= 1 && playerId <= 4) {
            this.presenter.selectPlayer(playerId);

            // Update player selector in UI
            const unitPlayerSelect = document.getElementById('unit-player-color') as HTMLSelectElement;
            if (unitPlayerSelect) {
                unitPlayerSelect.value = playerId.toString();
            }

            this.showToast('Player Selected', `Player ${playerId} selected`, 'success');
        } else {
            this.showToast('Invalid Selection', `Player ${playerId} not available`, 'error');
        }
    }
    
    private selectBrushGroup(args?: string): void {
        if (!args) return;

        const brushSizeSelect = document.getElementById('brush-size') as HTMLSelectElement;
        if (!brushSizeSelect) return;

        // Map keys to optgroup labels
        const groupMap: { [key: string]: string } = {
            'b': 'Brush',
            'f': 'Fill',
            's': 'Shapes'
        };

        const targetGroup = groupMap[args.toLowerCase()];
        if (!targetGroup) {
            this.showToast('Invalid Key', `Press b (brush), f (fill), or s (shapes)`, 'error');
            return;
        }

        // Find first option in the target optgroup
        for (let i = 0; i < brushSizeSelect.options.length; i++) {
            const option = brushSizeSelect.options[i];
            const optgroup = option.parentElement;

            if (optgroup && optgroup.tagName === 'OPTGROUP') {
                const groupLabel = optgroup.getAttribute('label');
                if (groupLabel === targetGroup) {
                    // Select this option
                    brushSizeSelect.value = option.value;
                    brushSizeSelect.dispatchEvent(new Event('change'));
                    this.showToast('Tool Selected', `${groupLabel}: ${option.text}`, 'success');
                    console.log(`Selected ${groupLabel}: ${option.text}`);
                    return;
                }
            }
        }

        this.showToast('Not Found', `${targetGroup} optgroup not found`, 'error');
    }

    private previewBrushGroup(args?: string): void {
        if (!args) return;

        const groupMap: { [key: string]: string } = {
            'b': 'Brush',
            'f': 'Fill',
            's': 'Shapes'
        };

        const targetGroup = groupMap[args.toLowerCase()];
        if (targetGroup) {
            console.log(`Preview: Will select first item in ${targetGroup} group`);
            // Could show a visual preview here if desired
        }
    }

    private cancelBrushSelection(): void {
        console.log('Cancelled brush selection');
        // Clean up any preview state if needed
    }

    private openBrushDropdown(): void {
        const brushSizeSelect = document.getElementById('brush-size') as HTMLSelectElement;
        if (brushSizeSelect) {
            // Focus the dropdown and show options
            brushSizeSelect.focus();
            brushSizeSelect.click();
            console.log('Opened brush/fill tool selector');
        }
    }

    private selectBrushSize(args?: string): void {
        const index = parseInt(args || '1'); // 1-based index

        this.hidePreviewIndicator(); // Hide preview indicator when committing

        const brushSizeSelect = document.getElementById('brush-size') as HTMLSelectElement;
        if (!brushSizeSelect || index < 1 || index > brushSizeSelect.options.length) {
            this.showToast('Invalid Selection', `Brush size ${index} not available`, 'error');
            return;
        }

        const option = brushSizeSelect.options[index - 1];
        const value = option.value;

        if (value.startsWith('fill:')) {
            const radius = parseInt(value.substring(5));
            this.setBrushSize("fill", radius);
        } else {
            const size = parseInt(value);
            this.setBrushSize("brush", size);
        }

        // Update brush size selector in UI and trigger onchange
        brushSizeSelect.value = value;
        brushSizeSelect.dispatchEvent(new Event('change'));

        this.showToast('Brush Tool Selected', `${option.text} selected`, 'success');
    }
    
    private activateClearMode(): void {
        // Clear any pending number input
        this.clearNumberInput();
        this.presenter.setPlacementMode('clear');
        this.showToast('Clear Mode', 'Clear mode activated - press R to reset', 'info');
    }

    private resetToDefaults(): void {
        if (this.editorToolsPanel) {
            this.editorToolsPanel.hideNumberOverlays();
        }

        // Reset to default terrain (grass)
        this.presenter.selectTerrain(1);
        this.presenter.setBrushSize("brush", 0);
        this.presenter.selectPlayer(1);

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
        
        console.log('Reset all tools to defaults');
        
        // Show toast notification
        this.showToast('Reset Complete', 'All tools reset to defaults', 'info');
    }
    
    // Note: Unit button selection is now handled by EditorToolsPanel internally

    // Reference image methods moved to ReferenceImagePanel - no longer needed here

    // Public methods for Phaser panel (for backward compatibility with UI)
    public initializePhaser(): void {
        console.log('Phaser initialization now handled by PhaserEditorComponent in dockview');
    }
    
    // Old EventBus handlers removed - components now use presenter directly
    
    private async handlePhaserReady() {
        console.log('EventBus: Phaser editor is ready');

        // Set ReferenceImagePanel dependencies now that Phaser is ready
        if (this.phaserEditorComponent && this.referenceImagePanel && this.world) {
            const referenceImageLayer = this.phaserEditorComponent.editorScene.getReferenceImageLayer();
            if (referenceImageLayer) {
                this.referenceImagePanel.setReferenceImageLayer(referenceImageLayer);
            }

            const worldId = this.world.getWorldId();
            if (worldId) {
                this.referenceImagePanel.setWorldId(worldId);
            }
        }

        // Load world data if available
        if (this.world && this.phaserEditorComponent) {
            // Give Phaser time to fully initialize webgl context and scene
            await this.phaserEditorComponent.editorScene.loadWorld(this.world);
            this.hasPendingWorldDataLoad = false;
            this.tileStatsPanel.refreshStats();
        }

        // Dismiss splash screen once everything is loaded and ready
        super.dismissSplashScreen();
    }
}

// Initialize the editor when DOM is ready
document.addEventListener('DOMContentLoaded', async () => {
    // Create page instance (just basic setup)
    const page = new WorldEditorPage("WorldEditorPage");
    
    // Create lifecycle controller with debug logging
    const lifecycleController = new LifecycleController(page.eventBus, LifecycleController.DefaultConfig)
    
    // Start breadth-first initialization
    await lifecycleController.initializeFromRoot(page);
});
