import { BasePage } from './BasePage';
import { DockviewApi, DockviewComponent } from 'dockview-core';
import { PhaserPanel } from './PhaserPanel';
import { TileStatsPanel } from './TileStatsPanel';
import { KeyboardShortcutManager, ShortcutConfig, KeyboardState } from './KeyboardShortcutManager';

const BRUSH_SIZE_NAMES = ['Single (1 hex)', 'Small (3 hexes)', 'Medium (5 hexes)', 'Large (9 hexes)', 'X-Large (15 hexes)', 'XX-Large (25 hexes)'];

class MapBounds {
  MinQ: number;
  MaxQ: number;
  MinR: number;
  MaxR: number;
  StartingCoord: { q: number; r: number };
  StartingX: number;
  MinX: number;
  MinY: number;
  MaxX: number;
  MaxY: number;
  MinXCoord: {Q: number, R: number};
  MinYCoord: {Q: number, R: number};
  MaxXCoord: {Q: number, R: number};
  MaxYCoord: {Q: number, R: number};
}

/**
 * Map Editor page with WASM integration for hex-based map editing
 */
class MapEditorPage extends BasePage {
    private currentMapId: string | null = null;
    private isNewMap: boolean = false;
    private mapBounds: MapBounds

    private mapData: {
        name: string;
        width: number;
        height: number;
        tiles: { [key: string]: { tileType: number } };
        units: { [key: string]: { unitType: number, playerId: number } };
        // Cube coordinate bounds for proper coordinate validation
        // Map bounds data from GetMapBounds for rendering optimization
    } | null = null;
    
    // Editor state
    private currentTerrain: number = 1; // Default to grass
    private currentUnit: number = 0; // Default to no unit
    private currentPlayerId: number = 1; // Default to player 1
    private placementMode: 'terrain' | 'unit' | 'clear' = 'terrain'; // Track what we're placing
    private brushSize: number = 0; // Default to single hex
    private editorOutput: HTMLElement | null = null;

    // Dockview interface
    private dockview: DockviewApi | null = null;
    
    // Phaser panel for map editing
    private phaserPanel: PhaserPanel | null = null;
    
    // TileStats panel for displaying statistics
    private tileStatsPanel: TileStatsPanel | null = null;

    // Keyboard shortcut manager
    private keyboardShortcutManager: KeyboardShortcutManager | null = null;

    // State management for undo/restore operations
    private savedUIState: {
        terrain: number;
        unit: number;
        playerId: number;
        brushSize: number;
        placementMode: 'terrain' | 'unit' | 'clear';
    } | null = null;

    // Change tracking for unsaved changes
    private hasUnsavedChanges: boolean = false;
    private originalMapData: string = '';
    private hasPendingMapDataLoad: boolean = false;

    constructor() {
        super();
        this.initializeSpecificComponents();
        this.initializeDockview();
        this.bindSpecificEvents();
        this.initializeKeyboardShortcuts();
        this.loadInitialState();
        this.setupUnsavedChangesWarning();
    }
    

    protected initializeSpecificComponents(): void {
        const mapIdInput = document.getElementById("mapIdInput") as HTMLInputElement | null;
        const isNewMapInput = document.getElementById("isNewMap") as HTMLInputElement | null;
        
        this.currentMapId = mapIdInput?.value.trim() || null;
        this.isNewMap = isNewMapInput?.value === "true";

        this.editorOutput = document.getElementById('editor-output');

        this.logToConsole('Map Editor initialized');
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
                    if (this.phaserPanel) {
                        this.phaserPanel.setTheme(isDarkMode);
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

        this.logToConsole('Dockview initialized');
    }

    protected bindSpecificEvents(): void {
        // Header buttons
        const saveButton = document.getElementById('save-map-btn');
        if (saveButton) {
            saveButton.addEventListener('click', this.saveMap.bind(this));
        }

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            // Don't interfere with input fields
            const target = e.target as HTMLElement;
            if (target && (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA')) {
                // Only handle our specific shortcuts in input fields, let other keys pass through
                if ((e.ctrlKey || e.metaKey) && e.key === 's') {
                    e.preventDefault();
                    if (this.hasUnsavedChanges) {
                        this.saveMap();
                    }
                }
                return;
            }
            
            // Ctrl+S or Cmd+S to save
            if ((e.ctrlKey || e.metaKey) && e.key === 's') {
                e.preventDefault();
                if (this.hasUnsavedChanges) {
                    this.saveMap();
                }
            }
        });

        const exportButton = document.getElementById('export-map-btn');
        if (exportButton) {
            exportButton.addEventListener('click', this.exportMap.bind(this));
        }


        const clearConsoleButton = document.getElementById('clear-console-btn');
        if (clearConsoleButton) {
            clearConsoleButton.addEventListener('click', this.clearConsole.bind(this));
        }

        // Map title editing
        const mapTitleInput = document.getElementById('map-title-input') as HTMLInputElement;
        const saveTitleButton = document.getElementById('save-title-btn') as HTMLButtonElement;
        const cancelTitleButton = document.getElementById('cancel-title-btn') as HTMLButtonElement;
        
        if (mapTitleInput && saveTitleButton && cancelTitleButton) {
            let originalTitle = mapTitleInput.value;
            let isEditing = false;
            
            const updateEditingState = (editing: boolean) => {
                isEditing = editing;
                if (editing) {
                    mapTitleInput.classList.add('editing');
                    saveTitleButton.classList.remove('hidden');
                    cancelTitleButton.classList.remove('hidden');
                } else {
                    mapTitleInput.classList.remove('editing');
                    saveTitleButton.classList.add('hidden');
                    cancelTitleButton.classList.add('hidden');
                }
            };
            
            const cancelEditing = () => {
                mapTitleInput.value = originalTitle;
                mapTitleInput.blur();
                updateEditingState(false);
                resizeInput();
            };
            
            const saveTitle = () => {
                const newTitle = mapTitleInput.value.trim();
                if (newTitle && newTitle !== originalTitle) {
                    this.saveMapTitle(newTitle);
                    originalTitle = newTitle; // Update original after successful save
                }
                mapTitleInput.blur();
                updateEditingState(false);
            };
            
            // Focus events for editing state
            mapTitleInput.addEventListener('focus', () => {
                updateEditingState(true);
            });
            
            mapTitleInput.addEventListener('blur', (e) => {
                // Don't blur if clicking on save/cancel buttons
                const relatedTarget = e.relatedTarget as HTMLElement;
                if (relatedTarget && (relatedTarget.id === 'save-title-btn' || relatedTarget.id === 'cancel-title-btn')) {
                    return;
                }
                
                // Auto-save if there are changes
                const newTitle = mapTitleInput.value.trim();
                if (newTitle && newTitle !== originalTitle) {
                    this.saveMapTitle(newTitle);
                    originalTitle = newTitle;
                } else if (!newTitle) {
                    mapTitleInput.value = originalTitle;
                }
                updateEditingState(false);
            });
            
            // Input changes
            mapTitleInput.addEventListener('input', () => {
                resizeInput();
                const hasChanges = mapTitleInput.value.trim() !== originalTitle;
                // Update button states based on changes
                if (hasChanges && mapTitleInput.value.trim()) {
                    saveTitleButton.classList.remove('opacity-50');
                    saveTitleButton.disabled = false;
                } else {
                    saveTitleButton.classList.add('opacity-50');
                    saveTitleButton.disabled = true;
                }
            });
            
            // Keyboard shortcuts
            mapTitleInput.addEventListener('keydown', (e) => {
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
                mapTitleInput.style.width = 'auto';
                mapTitleInput.style.width = Math.max(120, mapTitleInput.scrollWidth + 20) + 'px';
            };
            mapTitleInput.addEventListener('input', resizeInput);
            resizeInput(); // Initial resize
        }


        // Terrain palette buttons - radio button behavior
        document.querySelectorAll('.terrain-button').forEach(button => {
            button.addEventListener('click', (e) => {
                const clickedButton = e.currentTarget as HTMLElement;
                const terrain = clickedButton.getAttribute('data-terrain');
                if (terrain) {
                    // Remove selection from all terrain and unit buttons
                    document.querySelectorAll('.terrain-button, .unit-button').forEach(btn => {
                        btn.classList.remove('bg-blue-100', 'dark:bg-blue-900', 'border-blue-500');
                    });
                    
                    // Add selection to clicked button
                    clickedButton.classList.add('bg-blue-100', 'dark:bg-blue-900', 'border-blue-500');
                    
                    // Update editor state
                    const terrainValue = parseInt(terrain);
                    if (terrainValue === 0) {
                        this.placementMode = 'clear';
                        this.logToConsole('Selected clear mode');
                    } else {
                        this.currentTerrain = terrainValue;
                        this.placementMode = 'terrain';
                        this.logToConsole(`Selected terrain: ${terrain}`);
                    }
                }
            });
        });

        // Unit palette buttons - radio button behavior
        document.querySelectorAll('.unit-button').forEach(button => {
            button.addEventListener('click', (e) => {
                const clickedButton = e.currentTarget as HTMLElement;
                const unit = clickedButton.getAttribute('data-unit');
                if (unit) {
                    // Remove selection from all terrain and unit buttons
                    document.querySelectorAll('.terrain-button, .unit-button').forEach(btn => {
                        btn.classList.remove('bg-blue-100', 'dark:bg-blue-900', 'border-blue-500');
                    });
                    
                    // Add selection to clicked button
                    clickedButton.classList.add('bg-blue-100', 'dark:bg-blue-900', 'border-blue-500');
                    
                    // Update editor state
                    this.currentUnit = parseInt(unit);
                    this.placementMode = 'unit';
                    
                    // Get current player selection
                    const unitPlayerSelect = document.getElementById('unit-player-color') as HTMLSelectElement;
                    if (unitPlayerSelect) {
                        this.currentPlayerId = parseInt(unitPlayerSelect.value);
                    }
                    
                    this.logToConsole(`Selected unit: ${unit} for player ${this.currentPlayerId}`);
                }
            });
        });

        // Unit player color selector
        const unitPlayerSelect = document.getElementById('unit-player-color') as HTMLSelectElement;
        if (unitPlayerSelect) {
            unitPlayerSelect.addEventListener('change', (e) => {
                this.currentPlayerId = parseInt((e.target as HTMLSelectElement).value);
                if (this.placementMode === 'unit') {
                    this.logToConsole(`Unit player changed to: ${this.currentPlayerId}`);
                }
            });
        }

        // Brush size selector
        const brushSizeSelect = document.getElementById('brush-size') as HTMLSelectElement;
        if (brushSizeSelect) {
            brushSizeSelect.addEventListener('change', (e) => {
                this.setBrushSize(parseInt((e.target as HTMLSelectElement).value));
            });
        }

        // Painting action buttons
        document.querySelector('[data-action="paint-terrain"]')?.addEventListener('click', () => {
            this.paintTerrain();
        });
        document.querySelector('[data-action="flood-fill"]')?.addEventListener('click', () => {
            this.floodFill();
        });
        document.querySelector('[data-action="remove-terrain"]')?.addEventListener('click', () => {
            this.removeTerrain();
        });

        // Visual options
        const showGridCheckbox = document.getElementById('show-grid') as HTMLInputElement;
        if (showGridCheckbox) {
            showGridCheckbox.addEventListener('change', (e) => {
                const checked = (e.target as HTMLInputElement).checked;
                this.setShowGrid(checked);
            });
        }
        
        const showCoordinatesCheckbox = document.getElementById('show-coordinates') as HTMLInputElement;
        if (showCoordinatesCheckbox) {
            showCoordinatesCheckbox.addEventListener('change', (e) => {
                const checked = (e.target as HTMLInputElement).checked;
                this.setShowCoordinates(checked);
            });
        }



        // Export buttons
        document.querySelectorAll('[data-action="export-game"]').forEach(button => {
            button.addEventListener('click', (e) => {
                const target = e.target as HTMLElement;
                const players = parseInt(target.dataset.players || '2');
                this.exportToGame(players);
            });
        });

        // Advanced tool buttons
        document.querySelector('[data-action="fill-all-grass"]')?.addEventListener('click', () => {
            this.fillAllGrass();
        });
        document.querySelector('[data-action="create-island-map"]')?.addEventListener('click', () => {
            this.createIslandMap();
        });
        document.querySelector('[data-action="create-mountain-ridge"]')?.addEventListener('click', () => {
            this.createMountainRidge();
        });
        document.querySelector('[data-action="show-terrain-stats"]')?.addEventListener('click', () => {
            this.showTerrainStats();
        });
        document.querySelector('[data-action="randomize-terrain"]')?.addEventListener('click', () => {
            this.randomizeTerrain();
        });
        document.querySelector('[data-action="clear-map"]')?.addEventListener('click', () => {
            this.clearMap();
        });
        document.querySelector('[data-action="download-image"]')?.addEventListener('click', () => {
            this.downloadImage();
        });
        document.querySelector('[data-action="download-game-data"]')?.addEventListener('click', () => {
            this.downloadGameData();
        });
        
        // Phaser test buttons
        document.querySelector('[data-action="init-phaser"]')?.addEventListener('click', () => {
            this.initializePhaser();
        });
        
        // Reference image controls
        const loadReferenceBtn = document.getElementById('load-reference-btn');
        if (loadReferenceBtn) {
            loadReferenceBtn.addEventListener('click', () => {
                this.loadReferenceFromClipboard();
            });
        }
        
        const referenceModeSelect = document.getElementById('reference-mode') as HTMLSelectElement;
        if (referenceModeSelect) {
            referenceModeSelect.addEventListener('change', (e) => {
                const mode = parseInt((e.target as HTMLSelectElement).value);
                this.setReferenceMode(mode);
            });
        }
        
        const referenceAlphaSlider = document.getElementById('reference-alpha') as HTMLInputElement;
        const referenceAlphaValue = document.getElementById('reference-alpha-value');
        if (referenceAlphaSlider && referenceAlphaValue) {
            referenceAlphaSlider.addEventListener('input', (e) => {
                const alpha = parseInt((e.target as HTMLInputElement).value) / 100;
                referenceAlphaValue.textContent = `${Math.round(alpha * 100)}%`;
                this.setReferenceAlpha(alpha);
            });
        }
        
        const resetPositionBtn = document.getElementById('reference-reset-position');
        if (resetPositionBtn) {
            resetPositionBtn.addEventListener('click', () => {
                this.resetReferencePosition();
            });
        }
        
        const resetScaleBtn = document.getElementById('reference-reset-scale');
        if (resetScaleBtn) {
            resetScaleBtn.addEventListener('click', () => {
                this.resetReferenceScale();
            });
        }
        
        // X/Y Scale controls
        const scaleXMinusBtn = document.getElementById('reference-scale-x-minus');
        const scaleXPlusBtn = document.getElementById('reference-scale-x-plus');
        const scaleYMinusBtn = document.getElementById('reference-scale-y-minus');
        const scaleYPlusBtn = document.getElementById('reference-scale-y-plus');
        const scaleXInput = document.getElementById('reference-scale-x-value') as HTMLInputElement;
        const scaleYInput = document.getElementById('reference-scale-y-value') as HTMLInputElement;
        
        if (scaleXMinusBtn) {
            scaleXMinusBtn.addEventListener('click', () => {
                this.adjustReferenceScaleX(-0.01);
            });
        }
        
        if (scaleXPlusBtn) {
            scaleXPlusBtn.addEventListener('click', () => {
                this.adjustReferenceScaleX(0.01);
            });
        }
        
        if (scaleYMinusBtn) {
            scaleYMinusBtn.addEventListener('click', () => {
                this.adjustReferenceScaleY(-0.01);
            });
        }
        
        if (scaleYPlusBtn) {
            scaleYPlusBtn.addEventListener('click', () => {
                this.adjustReferenceScaleY(0.01);
            });
        }
        
        // Input field change handlers
        if (scaleXInput) {
            scaleXInput.addEventListener('change', () => {
                const value = parseFloat(scaleXInput.value);
                if (!isNaN(value)) {
                    this.setReferenceScaleX(value);
                }
            });
        }
        
        if (scaleYInput) {
            scaleYInput.addEventListener('change', () => {
                const value = parseFloat(scaleYInput.value);
                if (!isNaN(value)) {
                    this.setReferenceScaleY(value);
                }
            });
        }
        
        const clearReferenceBtn = document.getElementById('clear-reference-btn');
        if (clearReferenceBtn) {
            clearReferenceBtn.addEventListener('click', () => {
                this.clearReferenceImage();
            });
        }
        
        // File input and load from file button
        const fileInput = document.getElementById('reference-file-input') as HTMLInputElement;
        const loadFileBtn = document.getElementById('load-reference-file-btn');
        
        if (loadFileBtn && fileInput) {
            loadFileBtn.addEventListener('click', () => {
                fileInput.click();
            });
            
            fileInput.addEventListener('change', (e) => {
                const file = (e.target as HTMLInputElement).files?.[0];
                if (file) {
                    this.loadReferenceFromFile(file);
                }
            });
        }
    }

    private initializeKeyboardShortcuts(): void {
        const shortcuts: ShortcutConfig[] = [
            // Nature terrain shortcuts (n + number)
            {
                key: 'n',
                handler: (args?: string) => this.selectNatureTerrain(args),
                previewHandler: (args?: string) => this.previewNatureTerrain(args),
                cancelHandler: () => this.cancelSelection(),
                description: 'Select nature terrain by index',
                category: 'Terrain',
                requiresArgs: true,
                argType: 'number'
            },
            
            // City terrain shortcuts (c + number)  
            {
                key: 'c',
                handler: (args?: string) => this.selectCityTerrain(args),
                previewHandler: (args?: string) => this.previewCityTerrain(args),
                cancelHandler: () => this.cancelSelection(),
                description: 'Select city terrain by index',
                category: 'Terrain',
                requiresArgs: true,
                argType: 'number'
            },
            
            // Unit selection shortcuts (u + number)
            {
                key: 'u',
                handler: (args?: string) => this.selectUnit(args),
                previewHandler: (args?: string) => this.previewUnit(args),
                cancelHandler: () => this.cancelSelection(),
                description: 'Select unit type by index',
                category: 'Units',
                requiresArgs: true,
                argType: 'number'
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
                argType: 'number'
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
                argType: 'number'
            },
            
            
            // Reset shortcuts (esc)
            {
                key: 'Escape',
                handler: () => this.resetToDefaults(),
                description: 'Reset all tools to defaults',
                category: 'General'
            }
        ];

        this.keyboardShortcutManager = new KeyboardShortcutManager({
            shortcuts,
            timeout: 3000,
            immediateExecution: true, // Enable immediate execution mode
            previewDelay: 300, // 300ms preview delay
            onStateChange: (state, command) => this.handleKeyboardStateChange(state, command)
        });
        
        this.logToConsole('Keyboard shortcuts initialized with immediate execution mode');
    }

    private handleKeyboardStateChange(state: KeyboardState, command?: string): void {
        if (state === KeyboardState.AWAITING_ARGS && command) {
            // Show overlays immediately when entering AWAITING_ARGS state
            let overlayType: 'nature' | 'city' | 'unit' | null = null;
            
            switch (command) {
                case 'n':
                    overlayType = 'nature';
                    break;
                case 'c':
                    overlayType = 'city';
                    break;
                case 'u':
                    overlayType = 'unit';
                    break;
            }
            
            if (overlayType) {
                this.showNumberOverlays(overlayType);
                this.logToConsole(`Showing ${overlayType} shortcut overlays`);
            }
        } else if (state === KeyboardState.NORMAL) {
            // Hide overlays when returning to NORMAL state
            this.hideNumberOverlays();
        }
    }

    private loadInitialState(): void {
        // Theme button state is handled by BasePage
        this.updateEditorStatus('Initializing...');

        if (this.isNewMap) {
            this.logToConsole('Time: ${performance.now()} Creating new map...');
            this.initializeNewMap();
        } else if (this.currentMapId) {
            this.logToConsole(`Time: ${performance.now()} Loading existing map: ${this.currentMapId}`);
            this.loadExistingMap(this.currentMapId);
        } else {
            this.logToConsole('Error: No map ID provided');
            this.updateEditorStatus('Error');
        }
        
        // Initialize Phaser panel as the default editor
        setTimeout(() => {
            this.initializePhaserPanel();
        }, 1000);
    }


    private initializeNewMap(): void {
        // Try to load template map data from hidden element first
        const templateMapData = this.loadMapDataFromElement();
        
        if (templateMapData) {
            this.mapData = templateMapData;
            this.logToConsole('New map initialized with template data');
        } else {
            this.mapData = {
                name: "New Map",
                width: 8,
                height: 8,
                tiles: {},
                units: {},
            };
            this.logToConsole('New map initialized with default data');
        }
        
        this.updateEditorStatus('New Map');
    }

    private async loadExistingMap(mapId: string): Promise<void> {
        try {
            this.logToConsole(`Loading map data for ${mapId}...`);
            this.updateEditorStatus('Loading...');
            
            // Load map data from hidden element (passed from backend)
            const mapData = this.loadMapDataFromElement();
            
            if (mapData) {
                this.mapData = mapData;
                this.updateEditorStatus('Loaded');
                this.logToConsole('Map data loaded from server');
                
                // Mark that we have map data to load into Phaser
                this.hasPendingMapDataLoad = true;
            } else {
                throw new Error('No map data found in page');
            }
            
        } catch (error) {
            console.error('Failed to load map:', error);
            this.logToConsole(`Failed to load map: ${error}`);
            this.updateEditorStatus('Load Error');
        }
    }
    
    /**
     * Load map data from hidden element in the HTML
     */
    private loadMapDataFromElement(): any {
        try {
            const mapDataElement = document.getElementById('map-data-json');
            this.logToConsole(`Map data element found: ${mapDataElement ? 'YES' : 'NO'}`);
            
            if (mapDataElement && mapDataElement.textContent) {
                this.logToConsole(`Raw map data content: ${mapDataElement.textContent.substring(0, 200)}...`);
                const mapData = JSON.parse(mapDataElement.textContent);
                
                if (mapData && mapData !== null) {
                    this.logToConsole('Map data found in page element');
                    this.logToConsole(`Map data keys: ${Object.keys(mapData).join(', ')}`);
                    if (mapData.tiles) {
                        this.logToConsole(`Tiles data keys: ${Object.keys(mapData.tiles).join(', ')}`);
                    }
                    if (mapData.map_units) {
                        this.logToConsole(`Units data length: ${mapData.map_units.length}`);
                    }
                    return mapData;
                }
            }
            this.logToConsole('No map data found in page element');
            return null;
        } catch (error) {
            console.error('Error parsing map data from element:', error);
            this.logToConsole(`Error parsing map data: ${error}`);
            return null;
        }
    }
    
    /**
     * Load map data (tiles and units) into the Phaser scene
     */
    private async loadMapDataIntoPhaser(): Promise<void> {
        this.logToConsole('loadMapDataIntoPhaser called');
        this.logToConsole(`Phaser panel initialized: ${this.phaserPanel ? this.phaserPanel.getIsInitialized() : 'panel is null'}`);
        this.logToConsole(`Map data exists: ${this.mapData ? 'YES' : 'NO'}`);
        
        if (!this.phaserPanel || !this.phaserPanel.getIsInitialized() || !this.mapData) {
            this.logToConsole('Skipping Phaser data load - preconditions not met');
            return;
        }
        
        try {
            // Load tiles first using setTilesData for better performance
            if (this.mapData.tiles) {
                const tilesArray: Array<{ q: number; r: number; terrain: number; color: number }> = [];
                Object.entries(this.mapData.tiles).forEach(([key, tileData]: [string, any]) => {
                    const [q, r] = key.split(',').map(Number);
                    if (tileData.tile_type !== undefined) {
                        tilesArray.push({
                            q,
                            r,
                            terrain: tileData.tile_type,
                            color: tileData.player || 0
                        });
                    }
                });
                
                if (tilesArray.length > 0) {
                    this.logToConsole(`Attempting to load ${tilesArray.length} tiles`);
                    await this.phaserPanel.setTilesData(tilesArray);
                    this.logToConsole(`Loaded ${tilesArray.length} tiles into Phaser`);
                }
            }
            
            // Load units AFTER tiles are loaded - ensure proper rendering order
            const serverData = this.mapData as any;
            if (serverData.map_units && Array.isArray(serverData.map_units)) {
                let unitsLoaded = 0;
                
                // Add delay to ensure tiles are rendered first and textures are loaded
                setTimeout(() => {
                    serverData.map_units.forEach((unit: any) => {
                        if (unit.q !== undefined && unit.r !== undefined && unit.unit_type !== undefined) {
                            const playerId = unit.player || 1;
                            
                            this.logToConsole(`Loading unit ${unit.unit_type} (player ${playerId}) at Q=${unit.q}, R=${unit.r}`);
                            
                            // Paint unit in Phaser (units render above tiles due to depth=10)
                            const success = this.phaserPanel!.paintUnit(unit.q, unit.r, unit.unit_type, playerId);
                            if (success) {
                                unitsLoaded++;
                            } else {
                                this.logToConsole(`Failed to paint unit ${unit.unit_type} at Q=${unit.q}, R=${unit.r}`);
                            }
                            
                            // Also store in mapData.units for consistency
                            const key = `${unit.q},${unit.r}`;
                            if (this.mapData && !this.mapData.units) {
                                this.mapData.units = {};
                            }
                            if (this.mapData && this.mapData.units) {
                                this.mapData.units[key] = {
                                    unitType: unit.unit_type,
                                    playerId: playerId
                                };
                            }
                        }
                    });
                    this.logToConsole(`Successfully loaded ${unitsLoaded} units into Phaser`);
                    
                    // Refresh tile stats after all loading is complete
                    this.refreshTileStats();
                }, 300); // Increased delay to ensure tiles are rendered first
            } else {
                // No units to load, refresh stats immediately
                this.refreshTileStats();
            }
            
        } catch (error) {
            console.error('Error loading map data into Phaser:', error);
            this.logToConsole(`Error loading into Phaser: ${error}`);
        }
    }

    /**
     * Show loading indicator on map
     */
    private showMapLoadingIndicator(): void {
        const mapContainer = document.getElementById('phaser-container');
        if (mapContainer) {
            const loadingDiv = document.createElement('div');
            loadingDiv.id = 'map-loading-overlay';
            loadingDiv.className = 'absolute inset-0 bg-gray-900/50 flex items-center justify-center z-50';
            loadingDiv.innerHTML = `
                <div class="bg-white dark:bg-gray-800 px-4 py-3 rounded-lg shadow-lg border border-gray-200 dark:border-gray-600">
                    <div class="flex items-center space-x-3">
                        <div class="animate-spin h-5 w-5 border-2 border-blue-500 border-t-transparent rounded-full"></div>
                        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Loading map...</span>
                    </div>
                </div>
            `;
            mapContainer.appendChild(loadingDiv);
        }
    }

    /**
     * Hide loading indicator
     */
    private hideMapLoadingIndicator(): void {
        const loadingDiv = document.getElementById('map-loading-overlay');
        if (loadingDiv) {
            loadingDiv.remove();
        }
    }

    // Editor functions called by the template

    public setBrushTerrain(terrain: number): void {
        this.currentTerrain = terrain;
        
        const terrainNames = ['Unknown', 'Grass', 'Desert', 'Water', 'Mountain', 'Rock'];
        this.logToConsole(`Brush terrain set to: ${terrainNames[terrain]}`);
        this.updateBrushInfo();
        this.updateTerrainButtonSelection(terrain);
    }

    public setBrushSize(size: number): void {
        this.brushSize = size;
        
        this.logToConsole(`Brush size set to: ${BRUSH_SIZE_NAMES[size]}`);
        this.updateBrushInfo();
    }
    
    public setShowGrid(showGrid: boolean): void {
        if (this.phaserPanel && this.phaserPanel.getIsInitialized()) {
            this.phaserPanel.setShowGrid(showGrid);
            this.logToConsole(`Grid visibility set to: ${showGrid}`);
        } else {
            this.logToConsole('Phaser panel not available for grid toggle');
        }
    }
    
    public setShowCoordinates(showCoordinates: boolean): void {
        if (this.phaserPanel && this.phaserPanel.getIsInitialized()) {
            this.phaserPanel.setShowCoordinates(showCoordinates);
            this.logToConsole(`Coordinate visibility set to: ${showCoordinates}`);
        } else {
            this.logToConsole('Phaser panel not available for coordinates toggle');
        }
    }

    public paintTerrain(): void {
        const rowInput = document.getElementById('paint-row') as HTMLInputElement;
        const colInput = document.getElementById('paint-col') as HTMLInputElement;
        
        if (rowInput && colInput) {
            const row = parseInt(rowInput.value);
            const col = parseInt(colInput.value);
            this.logToConsole(`Painting terrain ${this.currentTerrain} at (${row}, ${col})`);
            // TODO: Implement actual painting logic with WASM
        }
    }

    public floodFill(): void {
        const rowInput = document.getElementById('paint-row') as HTMLInputElement;
        const colInput = document.getElementById('paint-col') as HTMLInputElement;
        
        if (rowInput && colInput) {
            const row = parseInt(rowInput.value);
            const col = parseInt(colInput.value);
            this.logToConsole(`Flood filling with terrain ${this.currentTerrain} from (${row}, ${col})`);
            // TODO: Implement flood fill logic with WASM
        }
    }

    public removeTerrain(): void {
        const rowInput = document.getElementById('paint-row') as HTMLInputElement;
        const colInput = document.getElementById('paint-col') as HTMLInputElement;
        
        if (rowInput && colInput) {
            const row = parseInt(rowInput.value);
            const col = parseInt(colInput.value);
            this.logToConsole(`Removing terrain at (${row}, ${col})`);
            // TODO: Implement terrain removal logic with WASM
        }
    }

    public downloadImage(): void {
        this.logToConsole('Downloading map image...');
        // TODO: Implement image download
        this.showToast('Download', 'Image download not yet implemented', 'info');
    }

    public exportToGame(players: number): void {
        this.logToConsole(`Exporting as ${players}-player game...`);
        // TODO: Implement game export
        this.showToast('Export', `${players}-player game export not yet implemented`, 'info');
    }

    public downloadGameData(): void {
        this.logToConsole('Downloading game data...');
        // TODO: Implement game data download
        this.showToast('Download', 'Game data download not yet implemented', 'info');
    }

    // Advanced tool functions
    public fillAllGrass(): void {
        this.logToConsole('Filling all tiles with grass...');
        
        if (this.phaserPanel && this.phaserPanel.getIsInitialized()) {
            this.phaserPanel.fillAllTerrain(1, 0); // Terrain type 1 = Grass
            this.logToConsole('All tiles filled with grass using Phaser');
        } else {
            this.logToConsole('Phaser panel not available, cannot fill grass');
        }
    }

    public createIslandMap(): void {
        this.logToConsole('Creating island map...');
        
        if (this.phaserPanel && this.phaserPanel.getIsInitialized()) {
            // Get current viewport center
            const center = this.phaserPanel.getViewportCenter();
            this.logToConsole(`Creating island at viewport center: Q=${center.q}, R=${center.r}`);
            
            // Create island pattern at viewport center with radius 5
            this.phaserPanel.createIslandPattern(center.q, center.r, 5);
            this.logToConsole('Island map created using Phaser');
        } else {
            this.logToConsole('Phaser panel not available, cannot create island map');
        }
    }

    public createMountainRidge(): void {
        this.logToConsole('Creating mountain ridge...');
        
        if (this.phaserPanel && this.phaserPanel.getIsInitialized()) {
            // Get current viewport center
            const center = this.phaserPanel.getViewportCenter();
            this.logToConsole(`Creating mountain ridge at viewport center: Q=${center.q}, R=${center.r}`);
            
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
                        this.phaserPanel.paintTile(q, r, 4, 0); // Mountain
                    } else {
                        this.phaserPanel.paintTile(q, r, 5, 0); // Rock
                    }
                }
            }
            this.logToConsole('Mountain ridge created using Phaser');
        } else {
            this.logToConsole('Phaser panel not available, cannot create mountain ridge');
        }
    }

    public showTerrainStats(): void {
        this.logToConsole('Calculating terrain statistics...');
        
        if (this.phaserPanel && this.phaserPanel.getIsInitialized()) {
            const tiles = this.phaserPanel.getTilesData();
            const stats = {
                grass: 0,
                desert: 0,
                water: 0,
                mountain: 0,
                rock: 0,
                other: 0
            };
            
            tiles.forEach(tile => {
                switch (tile.terrain) {
                    case 1: stats.grass++; break;
                    case 2: stats.desert++; break;
                    case 3: stats.water++; break;
                    case 4: stats.mountain++; break;
                    case 5: stats.rock++; break;
                    default: stats.other++; break;
                }
            });
            
            this.logToConsole('Terrain statistics:');
            this.logToConsole(`- Grass: ${stats.grass} tiles`);
            this.logToConsole(`- Desert: ${stats.desert} tiles`);
            this.logToConsole(`- Water: ${stats.water} tiles`);
            this.logToConsole(`- Mountain: ${stats.mountain} tiles`);
            this.logToConsole(`- Rock: ${stats.rock} tiles`);
            if (stats.other > 0) {
                this.logToConsole(`- Other: ${stats.other} tiles`);
            }
            this.logToConsole(`Total tiles: ${tiles.length}`);
        } else {
            this.logToConsole('Phaser panel not available, cannot calculate stats');
        }
    }

    public randomizeTerrain(): void {
        this.logToConsole('Randomizing terrain...');
        
        if (this.phaserPanel && this.phaserPanel.getIsInitialized()) {
            this.phaserPanel.randomizeTerrain();
            this.logToConsole('Terrain randomized using Phaser');
        } else {
            this.logToConsole('Phaser panel not available, cannot randomize terrain');
        }
    }

    public clearMap(): void {
        this.logToConsole('Clearing entire map...');
        
        if (this.phaserPanel && this.phaserPanel.getIsInitialized()) {
            // Clear all tiles and units from Phaser
            this.phaserPanel.clearAllTiles();
            this.phaserPanel.clearAllUnits();
            
            // Clear mapData as well
            if (this.mapData) {
                this.mapData.tiles = {};
                this.mapData.units = {};
            }
            
            this.markAsChanged();
            this.logToConsole('Map cleared successfully');
            this.showToast('Map Cleared', 'All tiles and units have been removed', 'info');
        } else {
            this.logToConsole('Phaser panel not available, cannot clear map');
        }
    }

    private setMapBounds(bounds: MapBounds)  {
        // Update mapData with bounds information
        this.mapBounds = bounds
    }

    // Canvas management methods removed - now handled by Phaser panel

    private async saveMap(): Promise<void> {
        if (!this.mapData) {
            this.showToast('Error', 'No map data to save', 'error');
            return;
        }

        try {
            this.logToConsole('Saving map...');
            this.updateEditorStatus('Saving...');

            // Build tiles data in the correct format for CreateMap API
            const tiles: { [key: string]: any } = {};
            if (this.phaserPanel && this.phaserPanel.getIsInitialized()) {
                const tilesData = this.phaserPanel.getTilesData();
                
                tilesData.forEach(tile => {
                    const key = `${tile.q},${tile.r}`;
                    tiles[key] = {
                        q: tile.q,
                        r: tile.r,
                        tile_type: tile.terrain,
                        player: tile.color
                    };
                });
                
                this.logToConsole(`Saving ${tilesData.length} tiles`);
            }

            // Build units data in the correct format for CreateMap API
            const mapUnits: any[] = [];
            if (this.mapData.units) {
                Object.entries(this.mapData.units).forEach(([key, unit]) => {
                    const [q, r] = key.split(',').map(Number);
                    mapUnits.push({
                        q: q,
                        r: r,
                        player: unit.playerId,
                        unit_type: unit.unitType
                    });
                });
                this.logToConsole(`Saving ${mapUnits.length} units`);
            }

            // Build the CreateMapRequest structure
            const createMapRequest = {
                map: {
                    id: this.currentMapId || 'new-map',
                    name: this.mapData.name || 'Untitled Map',
                    description: '',
                    tags: [],
                    difficulty: 'medium',
                    creator_id: 'editor-user', // TODO: Get actual user ID
                    tiles: tiles,
                    map_units: mapUnits
                }
            };

            const url = this.isNewMap ? '/api/v1/maps' : `/api/v1/maps/${this.currentMapId}`;
            const method = this.isNewMap ? 'POST' : 'PATCH';

            const response = await fetch(url, {
                method,
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(createMapRequest),
            });

            if (response.ok) {
                const result = await response.json();
                this.logToConsole('Map saved successfully');
                this.logToConsole(`Response: ${JSON.stringify(result)}`);
                this.updateEditorStatus('Saved');
                this.showToast('Success', 'Map saved successfully', 'success');
                
                // Mark as saved (clears unsaved changes flag)
                this.markAsSaved();
                
                // If this was a new map, update the current map ID
                const mapId = result.map?.id || result.id;
                if (this.isNewMap && mapId) {
                    this.currentMapId = mapId;
                    this.isNewMap = false;
                    // Update URL without reload
                    history.replaceState(null, '', `/maps/${mapId}/edit`);
                    this.logToConsole(`Map ID updated to: ${mapId}`);
                    this.logToConsole(`URL updated to: /maps/${mapId}/edit`);
                }
            } else {
                const errorText = await response.text();
                throw new Error(`Save failed: ${response.status} ${response.statusText} - ${errorText}`);
            }
        } catch (error) {
            console.error('Save failed:', error);
            this.logToConsole(`Save failed: ${error}`);
            this.updateEditorStatus('Save Error');
            this.showToast('Error', 'Failed to save map', 'error');
        }
    }

    private exportMap(): void {
        this.logToConsole('Exporting map...');
        // TODO: Implement map export functionality
        this.showToast('Export', 'Export functionality not yet implemented', 'info');
    }

    private async saveMapTitle(newTitle: string): Promise<void> {
        if (!newTitle.trim()) {
            this.showToast('Error', 'Map title cannot be empty', 'error');
            return;
        }

        const oldTitle = this.mapData?.name || 'Untitled Map';

        // Update the local map data
        if (this.mapData) {
            this.mapData.name = newTitle;
        }

        try {
            this.logToConsole(`Updating map title to: ${newTitle}`);
            
            // Save the map (this will include the title update)
            await this.saveMap();
            
            this.logToConsole('Map title updated successfully');
            this.showToast('Success', 'Map title updated', 'success');
            
        } catch (error) {
            console.error('Failed to save map title:', error);
            this.logToConsole(`Failed to save map title: ${error}`);
            this.showToast('Error', 'Failed to update map title', 'error');
            
            // Revert the title on error
            if (this.mapData) {
                this.mapData.name = oldTitle;
            }
            const mapTitleInput = document.getElementById('map-title-input') as HTMLInputElement;
            if (mapTitleInput) {
                mapTitleInput.value = oldTitle;
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
        console.log(`[MapEditor] ${message}`);
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
            brushInfo.textContent = `Current: ${terrainNames[this.currentTerrain]}, ${BRUSH_SIZE_NAMES[this.brushSize]}`;
        }
    }

    private updateTerrainButtonSelection(terrain: number): void {
        // Remove selection from all terrain and unit buttons
        document.querySelectorAll('.terrain-button, .unit-button').forEach(button => {
            button.classList.remove('bg-blue-100', 'dark:bg-blue-900', 'border-blue-500');
        });
        
        // Add selection to the correct terrain button
        document.querySelectorAll('.terrain-button').forEach(button => {
            const buttonTerrain = button.getAttribute('data-terrain');
            if (buttonTerrain === terrain.toString()) {
                button.classList.add('bg-blue-100', 'dark:bg-blue-900', 'border-blue-500');
            }
        });
    }

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
            init: () => {
                // Tools panel is already initialized through global event binding
            },
            dispose: () => {}
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
                // Phaser will handle its own initialization
                this.logToConsole('Phaser panel ready for initialization');
            },
            dispose: () => {}
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
            init: () => {
                // Initialize TileStats panel
                this.initializeTileStatsPanel();
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
                // Reference image panel is already initialized through global event binding
            },
            dispose: () => {}
        };
    }

    private createDefaultDockviewLayout(): void {
        if (!this.dockview) return;

        // Add main Phaser Map editor panel first (center) - will take remaining width
        this.dockview.addPanel({
            id: 'phaser',
            component: 'phaser',
            title: '🗺️ Map Editor'
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
            title: '📊 Map Statistics',
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

            this.logToConsole('Panel sizes set: Tools=270px, Advanced=260px, ReferenceImage=300px, Map Editor=remaining');
        } catch (error) {
            this.logToConsole(`Failed to set panel sizes: ${error}`);
        }
    }

    private saveDockviewLayout(): void {
        if (!this.dockview) return;
        
        const layout = this.dockview.toJSON();
        localStorage.setItem('map-editor-dockview-layout', JSON.stringify(layout));
    }
    
    private loadDockviewLayout(): any {
        const saved = localStorage.getItem('map-editor-dockview-layout');
        return saved ? JSON.parse(saved) : null;
    }

    // Unsaved changes tracking
    private setupUnsavedChangesWarning(): void {
        // Store initial map state
        this.originalMapData = JSON.stringify(this.mapData);
        
        // Browser beforeunload warning
        window.addEventListener('beforeunload', (e) => {
            if (this.hasUnsavedChanges) {
                e.preventDefault();
                e.returnValue = 'You have unsaved changes. Are you sure you want to leave?';
                return 'You have unsaved changes. Are you sure you want to leave?';
            }
        });
        
        // Track changes in map data
        this.trackMapChanges();
        
        // Initialize save button state
        setTimeout(() => {
            this.updateSaveButtonState();
        }, 100);
    }
    
    private trackMapChanges(): void {
        // Track changes in map title
        const mapTitleInput = document.getElementById('map-title-input') as HTMLInputElement;
        if (mapTitleInput) {
            mapTitleInput.addEventListener('input', () => {
                this.markAsChanged();
            });
        }
        
        // Track changes in Phaser panel (terrain painting, etc.)
        if (this.phaserPanel) {
            this.phaserPanel.onMapChange(() => {
                this.markAsChanged();
            });
        }
    }
    
    private markAsChanged(): void {
        if (!this.hasUnsavedChanges) {
            this.hasUnsavedChanges = true;
            this.updateSaveButtonState();
            this.logToConsole('Map has unsaved changes');
        }
        
        // Auto-refresh TileStats when map changes
        this.refreshTileStats();
    }
    
    private markAsSaved(): void {
        this.hasUnsavedChanges = false;
        this.originalMapData = JSON.stringify(this.mapData);
        this.updateSaveButtonState();
        this.logToConsole('Map changes saved');
    }
    
    private updateSaveButtonState(): void {
        const saveButton = document.getElementById('save-map-btn');
        if (saveButton) {
            if (this.hasUnsavedChanges) {
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
        
        // Destroy Phaser panel if it exists
        if (this.phaserPanel) {
            this.phaserPanel.destroy();
        }
        
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
    private initializePhaserPanel(): void {
        try {
            this.logToConsole('Initializing Phaser panel as default editor...');
            
            // Initialize Phaser panel
            this.phaserPanel = new PhaserPanel();
            
            // Set up logging callback
            this.phaserPanel.onLog((message) => {
                this.logToConsole(message);
            });
            
            // Set up event handlers
            this.phaserPanel.onTileClick((q, r) => {
                this.handlePhaserTileClick(q, r);
            });
            
            this.phaserPanel.onMapChange(() => {
                this.logToConsole('Phaser map changed');
                this.markAsChanged();
            });
            
            this.phaserPanel.onReferenceScaleChange((x: number, y: number) => {
                this.updateReferenceScaleDisplay();
            });
            
            // Initialize the panel
            const success = this.phaserPanel.initialize('editor-canvas-container');
            
            if (success) {
                // Apply current UI settings to Phaser
                const showGridCheckbox = document.getElementById('show-grid') as HTMLInputElement;
                const showCoordinatesCheckbox = document.getElementById('show-coordinates') as HTMLInputElement;
                
                if (showGridCheckbox) {
                    this.phaserPanel.setShowGrid(showGridCheckbox.checked);
                }
                if (showCoordinatesCheckbox) {
                    this.phaserPanel.setShowCoordinates(showCoordinatesCheckbox.checked);
                }
                
                // Set initial theme
                const isDarkMode = document.documentElement.classList.contains('dark');
                this.phaserPanel.setTheme(isDarkMode);
                
                this.updateEditorStatus('Ready');
                this.logToConsole('Phaser panel initialized successfully as default!');
                
                // Check if we have pending map data to load
                if (this.hasPendingMapDataLoad) {
                    this.logToConsole('Loading pending map data into Phaser...');
                    this.showMapLoadingIndicator();
                    // Use setTimeout with loading indicator
                    setTimeout(async () => {
                        await this.loadMapDataIntoPhaser();
                        this.hasPendingMapDataLoad = false;
                        this.hideMapLoadingIndicator();
                    }, 100);
                }
            } else {
                throw new Error('Failed to initialize Phaser panel');
            }
            
        } catch (error) {
            this.logToConsole(`Failed to initialize Phaser panel: ${error}`);
            this.updateEditorStatus('Phaser Error');
        }
    }
    
    private handlePhaserTileClick(q: number, r: number): void {
        try {
            // Update coordinate inputs
            const rowInput = document.getElementById('paint-row') as HTMLInputElement;
            const colInput = document.getElementById('paint-col') as HTMLInputElement;
            
            if (rowInput) rowInput.value = r.toString();
            if (colInput) colInput.value = q.toString();
            
            // Handle different placement modes
            this.logToConsole(`Click at Q=${q}, R=${r} in ${this.placementMode} mode`);
            
            if (this.placementMode === 'clear') {
                this.handleClearClick(q, r);
            } else if (this.placementMode === 'unit') {
                this.handleUnitPlacement(q, r);
            } else if (this.placementMode === 'terrain') {
                this.handleTerrainPlacement(q, r);
            }
            
        } catch (error) {
            this.logToConsole(`Phaser click error: ${error}`);
        }
    }
    
    private handleClearClick(q: number, r: number): void {
        // Get current brush size from dropdown
        const brushSizeSelect = document.getElementById('brush-size') as HTMLSelectElement;
        const brushSize = brushSizeSelect ? parseInt(brushSizeSelect.value) : 0;
        
        if (brushSize === 0) {
            // Single tile clear
            this.clearSingleTile(q, r);
        } else {
            // Multi-tile clear with brush
            this.clearTileArea(q, r, brushSize);
        }
        
        this.markAsChanged();
    }
    
    private clearSingleTile(q: number, r: number): void {
        // First priority: remove unit if exists
        if (this.unitExistsAt(q, r)) {
            this.removeUnitAt(q, r);
            this.logToConsole(`Removed unit at Q=${q}, R=${r}`);
            return;
        }
        
        // Second priority: remove tile if exists
        if (this.tileExistsAt(q, r)) {
            // Remove from mapData if it exists there
            if (this.mapData) {
                const tileKey = `${q},${r}`;
                delete this.mapData.tiles[tileKey];
            }
            this.phaserPanel?.removeTile(q, r);
            this.logToConsole(`Removed tile at Q=${q}, R=${r}`);
        } else {
            this.logToConsole(`Nothing to clear at Q=${q}, R=${r}`);
        }
    }
    
    private clearTileArea(centerQ: number, centerR: number, brushSize: number): void {
        const radius = this.getBrushRadius(brushSize);
        let clearedCount = 0;
        
        for (let dq = -radius; dq <= radius; dq++) {
            for (let dr = -radius; dr <= radius; dr++) {
                // Use cube distance to determine if tile is within brush radius
                const distance = Math.abs(dq) + Math.abs(dr) + Math.abs(-dq - dr);
                if (distance <= radius * 2) {
                    const q = centerQ + dq;
                    const r = centerR + dr;
                    
                    // Clear unit first if it exists
                    if (this.unitExistsAt(q, r)) {
                        this.removeUnitAt(q, r);
                        clearedCount++;
                    }
                    
                    // Then clear tile if it exists
                    if (this.tileExistsAt(q, r)) {
                        if (this.mapData) {
                            const tileKey = `${q},${r}`;
                            delete this.mapData.tiles[tileKey];
                        }
                        this.phaserPanel?.removeTile(q, r);
                        clearedCount++;
                    }
                }
            }
        }
        
        this.logToConsole(`Cleared ${clearedCount} tiles/units with brush size ${brushSize} at Q=${centerQ}, R=${centerR}`);
    }
    
    private getBrushRadius(brushSize: number): number {
        switch (brushSize) {
            case 0: return 0;  // Single
            case 1: return 1;  // Small (7 hexes)
            case 2: return 2;  // Medium (19 hexes)
            case 3: return 3;  // Large (37 hexes)
            case 4: return 4;  // X-Large (61 hexes)
            case 5: return 5;  // XX-Large (91 hexes)
            default: return 0;
        }
    }
    
    private handleUnitPlacement(q: number, r: number): void {
        if (!this.mapData) return;
        
        const tileKey = `${q},${r}`;
        const unitKey = `${q},${r}`;
        
        // Check if tile exists at this location by asking Phaser panel
        if (!this.tileExistsAt(q, r)) {
            this.logToConsole(`Cannot place unit at Q=${q}, R=${r} - no tile exists`);
            return;
        }
        
        // Check if there's already a unit at this location
        const existingUnit = this.getUnitAt(q, r);
        
        if (existingUnit && existingUnit.unitType === this.currentUnit) {
            // Same unit type exists - toggle it off (remove it)
            this.removeUnitAt(q, r);
            this.logToConsole(`Removed unit ${this.currentUnit} at Q=${q}, R=${r} (toggle)`);
        } else {
            // Different unit or no unit - place/replace the unit
            this.setUnitAt(q, r, this.currentUnit, this.currentPlayerId);
            this.logToConsole(`Placed unit ${this.currentUnit} (player ${this.currentPlayerId}) at Q=${q}, R=${r}`);
        }
        
        this.markAsChanged();
    }
    
    private handleTerrainPlacement(q: number, r: number): void {
        // Get current brush size from dropdown
        const brushSizeSelect = document.getElementById('brush-size') as HTMLSelectElement;
        const brushSize = brushSizeSelect ? parseInt(brushSizeSelect.value) : 0;
        
        // Determine color based on terrain type
        const playerColor = this.getPlayerColorForTerrain(this.currentTerrain);
        
        // Paint terrain with current settings
        this.phaserPanel?.paintTile(q, r, this.currentTerrain, playerColor, brushSize);
        
        // Update mapData tiles
        if (this.mapData) {
            // For now, just update the clicked tile in mapData
            // The full brush area will be handled by the Phaser panel
            const tileKey = `${q},${r}`;
            this.mapData.tiles[tileKey] = {
                tileType: this.currentTerrain
            };
        }
        
        this.logToConsole(`Painted terrain ${this.currentTerrain} (color ${playerColor}) at Q=${q}, R=${r} with brush size ${brushSize}`);
        this.markAsChanged();
    }
    
    /**
     * Get the appropriate player color for a terrain type
     */
    private getPlayerColorForTerrain(terrainType: number): number {
        // City terrains that support player colors
        const cityTerrains = [1, 2, 3, 16, 20]; // Land Base, Naval Base, Airport Base, Missile Silo, Mines
        
        if (cityTerrains.includes(terrainType)) {
            // For city terrains, get the selected player color
            const playerColorSelect = document.getElementById('player-color') as HTMLSelectElement;
            return playerColorSelect ? parseInt(playerColorSelect.value) : 0;
        } else {
            // For nature terrains, always use color 0 (neutral)
            return 0;
        }
    }

    /**
     * Check if a tile exists at the given coordinates
     */
    private tileExistsAt(q: number, r: number): boolean {
        if (!this.phaserPanel || !this.phaserPanel.getIsInitialized()) {
            return false;
        }
        
        const tilesData = this.phaserPanel.getTilesData();
        return tilesData.some(tile => tile.q === q && tile.r === r);
    }
    
    /**
     * Check if a unit exists at the given coordinates
     */
    private unitExistsAt(q: number, r: number): boolean {
        if (!this.phaserPanel || !this.phaserPanel.getIsInitialized()) {
            return false;
        }
        
        const unitsData = this.phaserPanel.getUnitsData();
        return unitsData.some(unit => unit.q === q && unit.r === r);
    }
    
    /**
     * Get unit data at the given coordinates (returns null if no unit exists)
     */
    private getUnitAt(q: number, r: number): { unitType: number; playerId: number } | null {
        if (!this.phaserPanel || !this.phaserPanel.getIsInitialized()) {
            return null;
        }
        
        const unitsData = this.phaserPanel.getUnitsData();
        const unit = unitsData.find(unit => unit.q === q && unit.r === r);
        return unit ? { unitType: unit.unitType, playerId: unit.playerId } : null;
    }
    
    /**
     * Set unit at the given coordinates
     */
    private setUnitAt(q: number, r: number, unitType: number, playerId: number): void {
        // Update mapData
        if (this.mapData) {
            this.mapData.units = this.mapData.units || {};
            const unitKey = `${q},${r}`;
            this.mapData.units[unitKey] = { unitType, playerId };
        }
        
        // Update Phaser scene
        this.phaserPanel?.paintUnit(q, r, unitType, playerId);
    }
    
    /**
     * Remove unit at the given coordinates
     */
    private removeUnitAt(q: number, r: number): void {
        // Remove from mapData
        if (this.mapData && this.mapData.units) {
            const unitKey = `${q},${r}`;
            delete this.mapData.units[unitKey];
        }
        
        // Remove from Phaser scene
        this.phaserPanel?.removeUnit(q, r);
    }
    
    // TileStats panel methods
    private initializeTileStatsPanel(): void {
        try {
            this.logToConsole('Initializing TileStats panel...');
            
            // Initialize TileStats panel
            this.tileStatsPanel = new TileStatsPanel();
            
            // Initialize the panel
            const success = this.tileStatsPanel.initialize('tilestats-container');
            
            if (success) {
                // Set up refresh button handler
                this.tileStatsPanel.onRefresh(() => {
                    this.refreshTileStats();
                });
                
                // Initial stats update
                this.refreshTileStats();
                
                this.logToConsole('TileStats panel initialized successfully!');
            } else {
                throw new Error('Failed to initialize TileStats panel');
            }
            
        } catch (error) {
            this.logToConsole(`Failed to initialize TileStats panel: ${error}`);
        }
    }
    
    private refreshTileStats(): void {
        if (!this.tileStatsPanel || !this.tileStatsPanel.getIsInitialized()) {
            return;
        }
        
        // Get tiles data from Phaser panel
        const tilesData = this.phaserPanel?.getTilesData() || [];
        
        // Get units data from mapData
        const unitsData = this.mapData?.units || {};
        
        // Update the stats panel
        this.tileStatsPanel.updateStats(tilesData, unitsData);
        
        this.logToConsole(`Stats refreshed: ${tilesData.length} tiles, ${Object.keys(unitsData).length} units`);
    }
    
    // State management for undo/restore operations
    private saveUIState(): void {
        this.savedUIState = {
            terrain: this.currentTerrain,
            unit: this.currentUnit,
            playerId: this.currentPlayerId,
            brushSize: this.brushSize,
            placementMode: this.placementMode
        };
    }
    
    private restoreUIState(): void {
        if (this.savedUIState) {
            this.currentTerrain = this.savedUIState.terrain;
            this.currentUnit = this.savedUIState.unit;
            this.currentPlayerId = this.savedUIState.playerId;
            this.brushSize = this.savedUIState.brushSize;
            this.placementMode = this.savedUIState.placementMode;
            
            // Update UI elements
            this.updateTerrainButtonSelection(this.currentTerrain);
            this.updateUnitButtonSelection(this.currentUnit);
            
            const brushSizeSelect = document.getElementById('brush-size') as HTMLSelectElement;
            if (brushSizeSelect) {
                brushSizeSelect.value = this.brushSize.toString();
            }
            
            const unitPlayerSelect = document.getElementById('unit-player-color') as HTMLSelectElement;
            if (unitPlayerSelect) {
                unitPlayerSelect.value = this.currentPlayerId.toString();
            }
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
    
    // Visual index mapping functions
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
    
    // Dynamic number overlay system
    private showNumberOverlays(type: 'nature' | 'city' | 'unit'): void {
        this.hideNumberOverlays(); // Remove any existing overlays
        
        let selector = '';
        let maxIndex = 0;
        
        switch (type) {
            case 'nature':
                selector = '[data-nature-index]';
                maxIndex = document.querySelectorAll(selector).length;
                break;
            case 'city':
                selector = '[data-city-index]';
                maxIndex = document.querySelectorAll(selector).length;
                break;
            case 'unit':
                selector = '[data-unit-index]';
                maxIndex = document.querySelectorAll(selector).length;
                break;
        }
        
        const buttons = document.querySelectorAll(selector);
        buttons.forEach((button) => {
            const element = button as HTMLElement;
            let index = '';
            
            switch (type) {
                case 'nature':
                    index = element.getAttribute('data-nature-index') || '';
                    break;
                case 'city':
                    index = element.getAttribute('data-city-index') || '';
                    break;
                case 'unit':
                    index = element.getAttribute('data-unit-index') || '';
                    break;
            }
            
            if (index) {
                const overlay = document.createElement('div');
                overlay.className = 'shortcut-number-overlay absolute top-0 right-0 bg-blue-500 text-white text-xs font-bold rounded-full w-5 h-5 flex items-center justify-center z-10 -mt-1 -mr-1';
                overlay.textContent = index;
                
                // Make button container relative if not already
                if (getComputedStyle(element).position === 'static') {
                    element.style.position = 'relative';
                }
                
                element.appendChild(overlay);
            }
        });
        
        this.logToConsole(`Showing ${type} shortcut numbers (1-${maxIndex - 1})`);
    }
    
    private hideNumberOverlays(): void {
        const overlays = document.querySelectorAll('.shortcut-number-overlay');
        overlays.forEach(overlay => overlay.remove());
    }

    // Preview handlers for immediate execution mode
    private previewNatureTerrain(args?: string): void {
        const index = parseInt(args || '1');
        
        if (index === 0) {
            // N+0 for clear mode
            this.saveUIState();
            this.placementMode = 'clear';
            this.updateTerrainButtonSelection(0);
            this.showPreviewIndicator('Preview: Clear mode');
            return;
        }
        
        // Use visual index mapping
        const terrainId = this.getTerrainIdByNatureIndex(index);
        const terrainName = this.getTerrainNameByNatureIndex(index);
        
        if (terrainId !== null) {
            this.saveUIState();
            this.setBrushTerrain(terrainId);
            this.placementMode = 'terrain';
            this.updateTerrainButtonSelection(terrainId);
            this.showPreviewIndicator(`Preview: ${terrainName} terrain`);
        }
    }
    
    private previewCityTerrain(args?: string): void {
        const index = parseInt(args || '1');
        
        // Use visual index mapping
        const terrainId = this.getTerrainIdByCityIndex(index);
        const terrainName = this.getTerrainNameByCityIndex(index);
        
        if (terrainId !== null) {
            this.saveUIState();
            this.setBrushTerrain(terrainId);
            this.placementMode = 'terrain';
            this.updateTerrainButtonSelection(terrainId);
            this.showPreviewIndicator(`Preview: ${terrainName}`);
        }
    }
    
    private previewUnit(args?: string): void {
        const index = parseInt(args || '1');
        
        // Use visual index mapping
        const unitId = this.getUnitIdByIndex(index);
        const unitName = this.getUnitNameByIndex(index);
        
        if (unitId !== null) {
            this.saveUIState();
            this.currentUnit = unitId;
            this.placementMode = 'unit';
            this.updateUnitButtonSelection(unitId);
            this.showPreviewIndicator(`Preview: ${unitName} for player ${this.currentPlayerId}`);
        }
    }
    
    private previewPlayer(args?: string): void {
        const playerId = parseInt(args || '1');
        
        if (playerId >= 1 && playerId <= 4) {
            this.saveUIState();
            this.currentPlayerId = playerId;
            
            const unitPlayerSelect = document.getElementById('unit-player-color') as HTMLSelectElement;
            if (unitPlayerSelect) {
                unitPlayerSelect.value = playerId.toString();
            }
            
            this.showPreviewIndicator(`Preview: Player ${playerId} selected`);
        }
    }
    
    private previewBrushSize(args?: string): void {
        const index = parseInt(args || '1'); // 1-based index
        
        // Map 1-based index to actual brush size values
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

    // Keyboard shortcut handlers
    private selectNatureTerrain(args?: string): void {
        const index = parseInt(args || '1');
        
        this.hidePreviewIndicator(); // Hide preview indicator when committing
        this.hideNumberOverlays(); // Hide number overlays when committing
        
        if (index === 0) {
            // N+0 for clear mode
            this.placementMode = 'clear';
            this.updateTerrainButtonSelection(0);
            this.showToast('Clear Mode', 'Clear mode selected', 'success');
            return;
        }
        
        // Use visual index mapping
        const terrainId = this.getTerrainIdByNatureIndex(index);
        const terrainName = this.getTerrainNameByNatureIndex(index);
        
        if (terrainId !== null) {
            this.setBrushTerrain(terrainId);
            this.placementMode = 'terrain';
            this.updateTerrainButtonSelection(terrainId);
            this.showToast('Terrain Selected', `${terrainName} selected`, 'success');
        } else {
            this.showToast('Invalid Selection', `Nature terrain ${index} not available`, 'error');
        }
    }
    
    private selectCityTerrain(args?: string): void {
        const index = parseInt(args || '1');
        
        this.hidePreviewIndicator(); // Hide preview indicator when committing
        this.hideNumberOverlays(); // Hide number overlays when committing
        
        // Use visual index mapping
        const terrainId = this.getTerrainIdByCityIndex(index);
        const terrainName = this.getTerrainNameByCityIndex(index);
        
        if (terrainId !== null) {
            this.setBrushTerrain(terrainId);
            this.placementMode = 'terrain';
            this.updateTerrainButtonSelection(terrainId);
            this.showToast('City Terrain Selected', `${terrainName} selected`, 'success');
        } else {
            this.showToast('Invalid Selection', `City terrain ${index} not available`, 'error');
        }
    }
    
    private selectUnit(args?: string): void {
        const index = parseInt(args || '1');
        
        this.hidePreviewIndicator(); // Hide preview indicator when committing
        this.hideNumberOverlays(); // Hide number overlays when committing
        
        // Use visual index mapping
        const unitId = this.getUnitIdByIndex(index);
        const unitName = this.getUnitNameByIndex(index);
        
        if (unitId !== null) {
            this.currentUnit = unitId;
            this.placementMode = 'unit';
            this.updateUnitButtonSelection(unitId);
            this.showToast('Unit Selected', `${unitName} selected for player ${this.currentPlayerId}`, 'success');
        } else {
            this.showToast('Invalid Selection', `Unit ${index} not available`, 'error');
        }
    }
    
    private selectPlayer(args?: string): void {
        const playerId = parseInt(args || '1');
        
        this.hidePreviewIndicator(); // Hide preview indicator when committing
        
        if (playerId >= 1 && playerId <= 4) {
            this.currentPlayerId = playerId;
            
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
        
        // Map 1-based index to actual brush size values
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
    
    private resetToDefaults(): void {
        this.hidePreviewIndicator(); // Hide preview indicator when committing
        
        // Reset to default terrain (grass)
        this.setBrushTerrain(1);
        this.placementMode = 'terrain';
        this.setBrushSize(0);
        this.currentPlayerId = 1;
        
        // Update UI elements
        this.updateTerrainButtonSelection(1);
        
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
    
    private updateUnitButtonSelection(unitType: number): void {
        // Remove selection from all terrain and unit buttons
        document.querySelectorAll('.terrain-button, .unit-button').forEach(btn => {
            btn.classList.remove('bg-blue-100', 'dark:bg-blue-900', 'border-blue-500');
        });
        
        // Add selection to clicked unit button
        const unitButton = document.querySelector(`[data-unit="${unitType}"]`);
        if (unitButton) {
            unitButton.classList.add('bg-blue-100', 'dark:bg-blue-900', 'border-blue-500');
        }
    }

    // Reference image methods
    private async loadReferenceFromClipboard(): Promise<void> {
        if (!this.phaserPanel || !this.phaserPanel.getIsInitialized()) {
            this.showToast('Error', 'Phaser panel not ready', 'error');
            return;
        }
        
        try {
            const success = await this.phaserPanel.loadReferenceFromClipboard();
            if (success) {
                this.showToast('Success', 'Reference image loaded from clipboard', 'success');
                this.updateReferenceStatus('Image loaded');
                
                // Enable mode selector
                const modeSelect = document.getElementById('reference-mode') as HTMLSelectElement;
                if (modeSelect && modeSelect.value === '0') {
                    modeSelect.value = '1'; // Default to background mode
                    this.setReferenceMode(1);
                }
            } else {
                this.showToast('Error', 'No image found in clipboard', 'error');
            }
        } catch (error) {
            this.logToConsole(`Failed to load reference image: ${error}`);
            this.showToast('Error', 'Failed to load reference image', 'error');
        }
    }
    
    private async loadReferenceFromFile(file: File): Promise<void> {
        if (!this.phaserPanel || !this.phaserPanel.getIsInitialized()) {
            this.showToast('Error', 'Phaser panel not ready', 'error');
            return;
        }
        
        try {
            this.logToConsole(`Loading reference image from file: ${file.name} (${file.size} bytes)`);
            const success = await this.phaserPanel.loadReferenceFromFile(file);
            if (success) {
                this.showToast('Success', `Reference image loaded: ${file.name}`, 'success');
                this.updateReferenceStatus(`File loaded: ${file.name}`);
                
                // Enable mode selector
                const modeSelect = document.getElementById('reference-mode') as HTMLSelectElement;
                if (modeSelect && modeSelect.value === '0') {
                    modeSelect.value = '1'; // Default to background mode
                    this.setReferenceMode(1);
                }
            } else {
                this.showToast('Error', 'Failed to load image file', 'error');
            }
        } catch (error) {
            this.logToConsole(`Failed to load reference image from file: ${error}`);
            this.showToast('Error', 'Failed to load reference image', 'error');
        }
    }
    
    private setReferenceMode(mode: number): void {
        if (!this.phaserPanel || !this.phaserPanel.getIsInitialized()) {
            return;
        }
        
        this.phaserPanel.setReferenceMode(mode);
        
        // Update UI dropdown to reflect current mode
        const modeSelect = document.getElementById('reference-mode') as HTMLSelectElement;
        if (modeSelect && modeSelect.value !== mode.toString()) {
            modeSelect.value = mode.toString();
        }
        
        // Show/hide position controls based on mode
        const positionControls = document.getElementById('reference-position-controls');
        if (positionControls) {
            positionControls.style.display = mode === 2 ? 'block' : 'none';
        }
        
        // Update scale display when switching to overlay mode
        if (mode === 2) {
            this.updateReferenceScaleDisplay();
        }
        
        const modeNames = ['Hidden', 'Background', 'Overlay'];
        this.logToConsole(`Reference mode set to: ${modeNames[mode]}`);
        this.updateReferenceStatus(mode === 0 ? 'Hidden' : `${modeNames[mode]} mode`);
    }
    
    private setReferenceAlpha(alpha: number): void {
        if (!this.phaserPanel || !this.phaserPanel.getIsInitialized()) {
            return;
        }
        
        this.phaserPanel.setReferenceAlpha(alpha);
        this.logToConsole(`Reference alpha set to: ${Math.round(alpha * 100)}%`);
    }
    
    private resetReferencePosition(): void {
        if (!this.phaserPanel || !this.phaserPanel.getIsInitialized()) {
            return;
        }
        
        this.phaserPanel.setReferencePosition(0, 0);
        this.logToConsole('Reference position reset to center');
        this.showToast('Position Reset', 'Reference image centered', 'success');
    }
    
    private resetReferenceScale(): void {
        if (!this.phaserPanel || !this.phaserPanel.getIsInitialized()) {
            return;
        }
        
        this.phaserPanel.setReferenceScaleFromTopLeft(1, 1);
        this.logToConsole('Reference scale reset to 100%');
        this.showToast('Scale Reset', 'Reference image scale reset', 'success');
        this.updateReferenceScaleDisplay();
    }
    
    private adjustReferenceScaleX(delta: number): void {
        if (!this.phaserPanel || !this.phaserPanel.getIsInitialized()) {
            return;
        }
        
        const state = this.phaserPanel.getReferenceState();
        if (!state) return;
        
        const newScaleX = Math.max(0.1, Math.min(5.0, state.scale.x + delta));
        this.phaserPanel.setReferenceScaleFromTopLeft(newScaleX, state.scale.y);
        this.updateReferenceScaleDisplay();
        this.logToConsole(`Reference X scale: ${newScaleX.toFixed(2)}`);
    }
    
    private adjustReferenceScaleY(delta: number): void {
        if (!this.phaserPanel || !this.phaserPanel.getIsInitialized()) {
            return;
        }
        
        const state = this.phaserPanel.getReferenceState();
        if (!state) return;
        
        const newScaleY = Math.max(0.1, Math.min(5.0, state.scale.y + delta));
        this.phaserPanel.setReferenceScaleFromTopLeft(state.scale.x, newScaleY);
        this.updateReferenceScaleDisplay();
        this.logToConsole(`Reference Y scale: ${newScaleY.toFixed(2)}`);
    }
    
    private setReferenceScaleX(scaleX: number): void {
        if (!this.phaserPanel || !this.phaserPanel.getIsInitialized()) {
            return;
        }
        
        const state = this.phaserPanel.getReferenceState();
        if (!state) return;
        
        const clampedScale = Math.max(0.1, Math.min(5.0, scaleX));
        this.phaserPanel.setReferenceScaleFromTopLeft(clampedScale, state.scale.y);
        this.updateReferenceScaleDisplay();
        this.logToConsole(`Reference X scale: ${clampedScale.toFixed(2)}`);
    }
    
    private setReferenceScaleY(scaleY: number): void {
        if (!this.phaserPanel || !this.phaserPanel.getIsInitialized()) {
            return;
        }
        
        const state = this.phaserPanel.getReferenceState();
        if (!state) return;
        
        const clampedScale = Math.max(0.1, Math.min(5.0, scaleY));
        this.phaserPanel.setReferenceScaleFromTopLeft(state.scale.x, clampedScale);
        this.updateReferenceScaleDisplay();
        this.logToConsole(`Reference Y scale: ${clampedScale.toFixed(2)}`);
    }
    
    private updateReferenceScaleDisplay(): void {
        if (!this.phaserPanel || !this.phaserPanel.getIsInitialized()) {
            return;
        }
        
        const state = this.phaserPanel.getReferenceState();
        if (!state) return;
        
        const scaleXInput = document.getElementById('reference-scale-x-value') as HTMLInputElement;
        const scaleYInput = document.getElementById('reference-scale-y-value') as HTMLInputElement;
        
        if (scaleXInput) {
            scaleXInput.value = state.scale.x.toFixed(2);
        }
        
        if (scaleYInput) {
            scaleYInput.value = state.scale.y.toFixed(2);
        }
    }
    
    private clearReferenceImage(): void {
        if (!this.phaserPanel || !this.phaserPanel.getIsInitialized()) {
            return;
        }
        
        this.phaserPanel.clearReferenceImage();
        
        // Reset UI controls
        const modeSelect = document.getElementById('reference-mode') as HTMLSelectElement;
        if (modeSelect) {
            modeSelect.value = '0';
        }
        
        const alphaSlider = document.getElementById('reference-alpha') as HTMLInputElement;
        const alphaValue = document.getElementById('reference-alpha-value');
        if (alphaSlider && alphaValue) {
            alphaSlider.value = '50';
            alphaValue.textContent = '50%';
        }
        
        // Hide position controls
        const positionControls = document.getElementById('reference-position-controls');
        if (positionControls) {
            positionControls.style.display = 'none';
        }
        
        this.updateReferenceStatus('No reference image loaded');
        this.logToConsole('Reference image cleared');
        this.showToast('Cleared', 'Reference image removed', 'success');
    }
    
    private updateReferenceStatus(status: string): void {
        const statusElement = document.getElementById('reference-status');
        if (statusElement) {
            statusElement.textContent = status;
        }
    }

    // Public methods for Phaser panel (for backward compatibility with UI)
    public initializePhaser(): void {
        this.initializePhaserPanel();
    }
}

// Initialize the editor when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    new MapEditorPage();
});
