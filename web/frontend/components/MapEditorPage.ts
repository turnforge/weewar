import { ThemeManager } from './ThemeManager';
import { Modal } from './Modal';
import { ToastManager } from './ToastManager';
import { DockviewApi, DockviewComponent } from 'dockview-core';
import { PhaserPanel } from './PhaserPanel';

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
class MapEditorPage {
    private themeManager: typeof ThemeManager | null = null;
    private modal: Modal | null = null;
    private toastManager: ToastManager | null = null;

    private themeToggleButton: HTMLButtonElement | null = null;
    private themeToggleIcon: HTMLElement | null = null;

    private currentMapId: string | null = null;
    private isNewMap: boolean = false;
    private mapBounds: MapBounds

    private mapData: {
        name: string;
        width: number;
        height: number;
        tiles: { [key: string]: { tileType: number } };
        map_units: any[];
        // Cube coordinate bounds for proper coordinate validation
        // Map bounds data from GetMapBounds for rendering optimization
    } | null = null;
    
    // Editor state
    private currentTerrain: number = 1; // Default to grass
    private brushSize: number = 0; // Default to single hex
    private editorOutput: HTMLElement | null = null;

    // Dockview interface
    private dockview: DockviewApi | null = null;
    
    // Phaser panel for map editing
    private phaserPanel: PhaserPanel | null = null;

    // Change tracking for unsaved changes
    private hasUnsavedChanges: boolean = false;
    private originalMapData: string = '';

    constructor() {
        this.initializeComponents();
        this.initializeDockview();
        this.bindEvents();
        this.loadInitialState();
        this.setupUnsavedChangesWarning();
    }
    

    private initializeComponents(): void {
        const mapIdInput = document.getElementById("mapIdInput") as HTMLInputElement | null;
        const isNewMapInput = document.getElementById("isNewMap") as HTMLInputElement | null;
        
        this.currentMapId = mapIdInput?.value.trim() || null;
        this.isNewMap = isNewMapInput?.value === "true";

        ThemeManager.init();
        this.modal = Modal.init();
        this.toastManager = ToastManager.init();

        this.themeToggleButton = document.getElementById('theme-toggle-button') as HTMLButtonElement;
        this.themeToggleIcon = document.getElementById('theme-toggle-icon');
        this.editorOutput = document.getElementById('editor-output');

        if (!this.themeToggleButton || !this.themeToggleIcon) {
            console.warn("Theme toggle button or icon element not found in Header.");
        }

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
                    case 'console':
                        return this.createConsoleComponent();
                    case 'advancedTools':
                        return this.createAdvancedToolsComponent();
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

    private bindEvents(): void {
        // Theme toggle
        if (this.themeToggleButton) {
            this.themeToggleButton.addEventListener('click', this.handleThemeToggleClick.bind(this));
        }

        // Header buttons
        const saveButton = document.getElementById('save-map-btn');
        if (saveButton) {
            saveButton.addEventListener('click', this.saveMap.bind(this));
        }

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
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
                    // Remove selection from all buttons
                    document.querySelectorAll('.terrain-button').forEach(btn => {
                        btn.classList.remove('bg-blue-100', 'dark:bg-blue-900', 'border-blue-500');
                    });
                    
                    // Add selection to clicked button
                    clickedButton.classList.add('bg-blue-100', 'dark:bg-blue-900', 'border-blue-500');
                    
                    // Update current terrain (no longer needed, but keeping for compatibility)
                    this.currentTerrain = parseInt(terrain);
                    this.logToConsole(`Selected terrain: ${terrain}`);
                }
            });
        });

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
    }

    private loadInitialState(): void {
        this.updateThemeButtonState();
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
        this.mapData = {
            name: "New Map",
            width: 8,
            height: 8,
            tiles: {},
            map_units: []
        };
        this.updateEditorStatus('New Map');
        this.logToConsole('New map initialized');
    }

    private async loadExistingMap(mapId: string): Promise<void> {
        try {
            // TODO: Load map data from API
            this.logToConsole(`Loading map data for ${mapId}...`);
            this.updateEditorStatus('Loading...');
            
            // Placeholder - will be replaced with actual API call
            setTimeout(() => {
                this.mapData = {
                    name: `Map ${mapId}`,
                    width: 8,
                    height: 8,
                    tiles: {},
                    map_units: []
                };
                this.updateEditorStatus('Loaded');
                this.logToConsole('Map data loaded');
            }, 1000);
            
        } catch (error) {
            console.error('Failed to load map:', error);
            this.logToConsole(`Failed to load map: ${error}`);
            this.updateEditorStatus('Load Error');
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
        
        const sizeNames = ['Single (1 hex)', 'Small (7 hexes)', 'Medium (19 hexes)', 'Large (37 hexes)', 'X-Large (61 hexes)', 'XX-Large (91 hexes)'];
        this.logToConsole(`Brush size set to: ${sizeNames[size]}`);
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
        this.toastManager?.showToast('Download', 'Image download not yet implemented', 'info');
    }

    public exportToGame(players: number): void {
        this.logToConsole(`Exporting as ${players}-player game...`);
        // TODO: Implement game export
        this.toastManager?.showToast('Export', `${players}-player game export not yet implemented`, 'info');
    }

    public downloadGameData(): void {
        this.logToConsole('Downloading game data...');
        // TODO: Implement game data download
        this.toastManager?.showToast('Download', 'Game data download not yet implemented', 'info');
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

    private setMapBounds(bounds: MapBounds)  {
        // Update mapData with bounds information
        this.mapBounds = bounds
    }

    // Canvas management methods removed - now handled by Phaser panel

    private async saveMap(): Promise<void> {
        if (!this.mapData) {
            this.toastManager?.showToast('Error', 'No map data to save', 'error');
            return;
        }

        try {
            this.logToConsole('Saving map...');
            this.updateEditorStatus('Saving...');

            // Collect current map data including tiles from Phaser panel
            const mapToSave = {
                ...this.mapData
            };

            // Get terrain data from Phaser panel if available
            if (this.phaserPanel && this.phaserPanel.getIsInitialized()) {
                const tilesData = this.phaserPanel.getTilesData();
                mapToSave.tiles = {};
                
                tilesData.forEach(tile => {
                    const key = `${tile.q},${tile.r}`;
                    mapToSave.tiles[key] = {
                        tileType: tile.terrain,
                        playerId: tile.playerId || 0
                    };
                });
                
                this.logToConsole(`Saving ${tilesData.length} tiles`);
            }

            const url = this.isNewMap ? '/api/maps' : `/api/maps/${this.currentMapId}`;
            const method = this.isNewMap ? 'POST' : 'PUT';

            const response = await fetch(url, {
                method,
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(mapToSave),
            });

            if (response.ok) {
                const result = await response.json();
                this.logToConsole('Map saved successfully');
                this.updateEditorStatus('Saved');
                this.toastManager?.showToast('Success', 'Map saved successfully', 'success');
                
                // Mark as saved (clears unsaved changes flag)
                this.markAsSaved();
                
                // If this was a new map, update the current map ID
                if (this.isNewMap && result.id) {
                    this.currentMapId = result.id;
                    this.isNewMap = false;
                    // Update URL without reload
                    history.replaceState(null, '', `/maps/${result.id}/edit`);
                    this.logToConsole(`Map ID updated to: ${result.id}`);
                }
            } else {
                const errorText = await response.text();
                throw new Error(`Save failed: ${response.status} ${response.statusText} - ${errorText}`);
            }
        } catch (error) {
            console.error('Save failed:', error);
            this.logToConsole(`Save failed: ${error}`);
            this.updateEditorStatus('Save Error');
            this.toastManager?.showToast('Error', 'Failed to save map', 'error');
        }
    }

    private exportMap(): void {
        this.logToConsole('Exporting map...');
        // TODO: Implement map export functionality
        this.toastManager?.showToast('Export', 'Export functionality not yet implemented', 'info');
    }

    private async saveMapTitle(newTitle: string): Promise<void> {
        if (!newTitle.trim()) {
            this.toastManager?.showToast('Error', 'Map title cannot be empty', 'error');
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
            this.toastManager?.showToast('Success', 'Map title updated', 'success');
            
        } catch (error) {
            console.error('Failed to save map title:', error);
            this.logToConsole(`Failed to save map title: ${error}`);
            this.toastManager?.showToast('Error', 'Failed to update map title', 'error');
            
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
            const sizeNames = ['Single (1 hex)', 'Small (7 hexes)', 'Medium (19 hexes)', 'Large (37 hexes)', 'X-Large (61 hexes)', 'XX-Large (91 hexes)'];
            brushInfo.textContent = `Current: ${terrainNames[this.currentTerrain]}, ${sizeNames[this.brushSize]}`;
        }
    }

    private updateTerrainButtonSelection(terrain: number): void {
        document.querySelectorAll('.terrain-button').forEach(button => {
            const buttonTerrain = button.getAttribute('data-terrain');
            if (buttonTerrain === terrain.toString()) {
                button.classList.add('bg-blue-100', 'dark:bg-blue-900', 'border-blue-500');
            } else {
                button.classList.remove('bg-blue-100', 'dark:bg-blue-900', 'border-blue-500');
            }
        });
    }

    // Theme management
    private handleThemeToggleClick(): void {
        const currentSetting = ThemeManager.getCurrentThemeSetting();
        const nextSetting = ThemeManager.getNextTheme(currentSetting);
        ThemeManager.setTheme(nextSetting);
        this.updateThemeButtonState(nextSetting);
    }

    private updateThemeButtonState(currentTheme?: string): void {
        if (!this.themeToggleButton || !this.themeToggleIcon) return;

        const themeToDisplay = currentTheme || ThemeManager.getCurrentThemeSetting();
        const iconSVG = ThemeManager.getIconSVG(themeToDisplay);
        const label = `Toggle theme (currently: ${ThemeManager.getThemeLabel(themeToDisplay)})`;

        this.themeToggleIcon.innerHTML = iconSVG;
        this.themeToggleButton.setAttribute('aria-label', label);
        this.themeToggleButton.setAttribute('title', label);
    }

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

    private createDefaultDockviewLayout(): void {
        if (!this.dockview) return;

        // Add Phaser panel first (center)
        this.dockview.addPanel({
            id: 'phaser',
            component: 'phaser',
            title: '🎮 Phaser Editor'
        });

        // Add tools panel to the left of Phaser
        this.dockview.addPanel({
            id: 'tools',
            component: 'tools',
            title: '🎨 Tools & Terrain',
            position: { direction: 'left', referencePanel: 'phaser' }
        });

        // Add advanced tools panel to the right of Phaser
        this.dockview.addPanel({
            id: 'advancedTools',
            component: 'advancedTools',
            title: '🔧 Advanced & View',
            position: { direction: 'right', referencePanel: 'phaser' }
        });

        // Add console panel below Phaser
        this.dockview.addPanel({
            id: 'console',
            component: 'console',
            title: '💻 Console',
            position: { direction: 'below', referencePanel: 'phaser' }
        });
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
            // Get current terrain type from selected terrain button
            const selectedTerrainButton = document.querySelector('.terrain-button.bg-blue-100') as HTMLElement;
            const terrainType = selectedTerrainButton ? 
                parseInt(selectedTerrainButton.getAttribute('data-terrain') || '1') : 1;
            
            // Get current brush size from dropdown
            const brushSizeSelect = document.getElementById('brush-size') as HTMLSelectElement;
            const brushSize = brushSizeSelect ? parseInt(brushSizeSelect.value) : 0;
            
            // Paint with current settings
            this.phaserPanel?.paintTile(q, r, terrainType, 0, brushSize);
            
            // Update coordinate inputs
            const rowInput = document.getElementById('paint-row') as HTMLInputElement;
            const colInput = document.getElementById('paint-col') as HTMLInputElement;
            
            if (rowInput) rowInput.value = r.toString();
            if (colInput) colInput.value = q.toString();
            
        } catch (error) {
            this.logToConsole(`Phaser paint error: ${error}`);
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
