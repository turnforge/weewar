import { ThemeManager } from './ThemeManager';
import { Modal } from './Modal';
import { ToastManager } from './ToastManager';
import { DockviewApi, DockviewComponent } from 'dockview-core';
import { PhaserMapEditor } from './phaser/PhaserMapEditor';

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
    private editorCanvas: HTMLElement | null = null;
    private mapCanvas: HTMLCanvasElement | null = null;
    private canvasContext: CanvasRenderingContext2D | null = null;
    private editorOutput: HTMLElement | null = null;
    
    // Tile dimensions for client-side calculations
    private tileDimensions: {
        tileWidth: number;
        tileHeight: number;
        yIncrement: number;
    } | null = null;
    
    // Scroll and padding management
    private scrollOffset: { x: number; y: number } = { x: 150, y: 150 }; // Padding around map for extension
    private canvasSize : { width: number; height: number } = { width: 400, height: 400 }; // Padding around map for extension

    // WASM interface
    private wasmModule: any = null;
    private wasmInitialized: boolean = false;
    private wasmWorld: any = null;
    private wasmViewState: any = null;
    private wasmCanvasRenderer: any = null;

    // Dockview interface
    private dockview: DockviewApi | null = null;
    
    // Phaser editor for testing
    private phaserEditor: PhaserMapEditor | null = null;

    constructor() {
        this.initializeComponents();
        this.initializeDockview();
        this.bindEvents();
        this.loadInitialState();
        this.initializeWasm();
    }
    
    // Calculate canvas size based on actual map bounds from WASM
    private calculateCanvasSize(mapWidth: number, mapHeight: number): { width: number; height: number } {
        // Update map bounds from WASM first
        this.updateMapBoundsFromWasm();
        
        // Use the updated bounds from mapData if available
        if (this.mapBounds && this.mapBounds.MinX !== undefined && this.mapBounds.MaxX !== undefined && 
            this.mapBounds.MinY !== undefined && this.mapBounds.MaxY !== undefined) {
            
            const mapPixelWidth = this.mapBounds.MaxX - this.mapBounds.MinX;
            const mapPixelHeight = this.mapBounds.MaxY - this.mapBounds.MinY;
            
            const canvasSize = {
                width: Math.max(mapPixelWidth + this.scrollOffset.x * 2, 400),
                height: Math.max(mapPixelHeight + this.scrollOffset.y * 2, 300)
            };
            
            this.logToConsole(`Canvas size calculated from map bounds: ${canvasSize.width}x${canvasSize.height} (map: ${mapPixelWidth}x${mapPixelHeight}, padding: ${this.scrollOffset.x}x${this.scrollOffset.y})`);
            return canvasSize;
        }
        
        // Fallback to tile-based calculation if bounds aren't available
        if (!this.tileDimensions) {
            return {
                width: Math.max(mapWidth * 40 + this.scrollOffset.x * 2, 400),
                height: Math.max(mapHeight * 35 + this.scrollOffset.y * 2, 300)
            };
        }
        
        const mapPixelWidth = mapWidth * this.tileDimensions.tileWidth;
        const mapPixelHeight = mapHeight * this.tileDimensions.yIncrement;
        
        return {
            width: Math.max(mapPixelWidth + this.scrollOffset.x * 2, 400),
            height: Math.max(mapPixelHeight + this.scrollOffset.y * 2, 300)
        };
    }
    
    // Client-side coordinate conversion using cached tile dimensions
    private clientPixelToCoords(x: number, y: number): { q: number; r: number; withinBounds: boolean } {
        if (!this.tileDimensions) {
            // Fallback - use WASM function
            const result = (window as any).pixelToCoords(x, y);
            return {
                q: result.data.cubeQ,
                r: result.data.cubeR,
                withinBounds: result.data.withinBounds
            };
        }
        
        // Adjust for scroll offset
        const adjustedX = x - this.scrollOffset.x;
        const adjustedY = y - this.scrollOffset.y;
        
        // Use proper XYToQR conversion
        const cubeCoord = this.xyToQR(adjustedX, adjustedY);
        
        // Check if within current map bounds (this is for UI purposes, not coordinate conversion)
        const withinBounds = this.mapData ? 
            (cubeCoord.q >= 0 && cubeCoord.q < this.mapData.width && cubeCoord.r >= 0 && cubeCoord.r < this.mapData.height) : false;
        
        return { q: cubeCoord.q, r: cubeCoord.r, withinBounds };
    }

    // XYToQR converts screen coordinates to cube coordinates
    // Based on the Go implementation in lib/map.go
    private xyToQR(x: number, y: number): { q: number; r: number } {
        if (!this.tileDimensions) {
            throw new Error("Tile dimensions not available");
        }

        const SQRT3 = 1.732050808; // sqrt(3)
        
        // Convert normalized origin to pixel coordinates (currently 0,0)
        const originPixelX = 0.0;
        const originPixelY = 0.0;
        
        // Translate screen coordinates to hex coordinate space
        const hexX = x - originPixelX;
        const hexY = y - originPixelY;
        
        // For pointy-topped hexagons, convert pixel coordinates to fractional hex coordinates
        // Using inverse of the hex-to-pixel conversion formulas:
        // x = size * sqrt(3) * (q + r/2)  =>  q = (sqrt(3) * x) / (y * 3)
        // y = size * 3/2 * r             =>  r = (y * 2.0 / 3.0)
        
        // Calculate fractional q coordinate
        const fractionalQ = (hexX * SQRT3) / (this.tileDimensions.tileWidth * 3.0);
        
        // Calculate fractional r coordinate
        const fractionalR = (hexY * 2.0) / (this.tileDimensions.tileHeight * 3.0);
        
        // Round to nearest integer coordinates using cube coordinate rounding
        return this.roundCubeCoord(fractionalQ, fractionalR);
    }

    // CenterXYForTile returns the center pixel coordinates for a given cube coordinate
    // Based on the Go implementation in lib/map.go
    private centerXYForTile(q: number, r: number): { x: number; y: number } {
        if (!this.tileDimensions) {
            throw new Error("Tile dimensions not available");
        }

        // Convert normalized origin to pixel coordinates (currently 0,0)
        const originPixelX = 0.0;
        const originPixelY = 0.0;
        
        // For pointy-topped hexagons with odd-r layout:
        // x = size * sqrt(3) * (q + r/2)
        // y = size * 3/2 * r
        
        const x = originPixelX + this.tileDimensions.tileWidth * 1.732050808 * (q + r / 2.0);
        const y = originPixelY + this.tileDimensions.tileHeight * 3.0 / 2.0 * r;
        
        return { x, y };
    }

    // Round cube coordinates to nearest integer while maintaining constraint s = -q - r
    // Based on the Go implementation in lib/map.go
    private roundCubeCoord(fractionalQ: number, fractionalR: number): { q: number; r: number } {
        // Calculate s from the cube coordinate constraint: s = -q - r
        const fractionalS = -fractionalQ - fractionalR;
        
        // Round each coordinate to nearest integer
        let roundedQ = Math.round(fractionalQ);
        let roundedR = Math.round(fractionalR);
        let roundedS = Math.round(fractionalS);
        
        // Calculate rounding deltas
        const deltaQ = Math.abs(roundedQ - fractionalQ);
        const deltaR = Math.abs(roundedR - fractionalR);
        const deltaS = Math.abs(roundedS - fractionalS);
        
        // Fix the coordinate with the largest rounding error to maintain constraint
        if (deltaQ > deltaR && deltaQ > deltaS) {
            roundedQ = -roundedR - roundedS;
        } else if (deltaR > deltaS) {
            roundedR = -roundedQ - roundedS;
        } else {
            roundedS = -roundedQ - roundedR;
        }
        
        // Return the rounded cube coordinate (s is implicit)
        return { q: roundedQ, r: roundedR };
    }
    
    // Set scroll offset (for panning/zooming)
    public setScrollOffset(x: number, y: number): void {
        this.scrollOffset = { x, y };
        // Request editor refresh after scroll change
        this.editorRefresh();
    }
    
    // Get current scroll offset
    public getScrollOffset(): { x: number; y: number } {
        return { ...this.scrollOffset };
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
        this.editorCanvas = document.getElementById('editor-canvas-container');
        this.mapCanvas = document.getElementById('map-canvas') as HTMLCanvasElement;
        this.editorOutput = document.getElementById('editor-output');
        
        // Initialize canvas context
        if (this.mapCanvas) {
            this.canvasContext = this.mapCanvas.getContext('2d');
            this.initializeCanvas();
        }

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
                    case 'canvas':
                        return this.createCanvasComponent();
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

        const exportButton = document.getElementById('export-map-btn');
        if (exportButton) {
            exportButton.addEventListener('click', this.exportMap.bind(this));
        }


        const clearConsoleButton = document.getElementById('clear-console-btn');
        if (clearConsoleButton) {
            clearConsoleButton.addEventListener('click', this.clearConsole.bind(this));
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


        // Rendering buttons
        document.querySelectorAll('[data-action="render-map"]').forEach(button => {
            button.addEventListener('click', (e) => {
                const target = e.target as HTMLElement;
                const width = parseInt(target.dataset.width || '600');
                const height = parseInt(target.dataset.height || '450');
                this.renderEditor(width, height);
            });
        });

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
        document.querySelector('[data-action="create-test-pattern"]')?.addEventListener('click', () => {
            this.createTestPattern();
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
        document.querySelector('[data-action="test-phaser-pattern"]')?.addEventListener('click', () => {
            this.testPhaserPattern();
        });
        document.querySelector('[data-action="destroy-phaser"]')?.addEventListener('click', () => {
            this.destroyPhaser();
        });

        // Canvas interactions
        if (this.mapCanvas) {
            this.mapCanvas.addEventListener('click', this.handleCanvasClick.bind(this));
            this.mapCanvas.addEventListener('mousemove', this.handleCanvasMouseMove.bind(this));
        }
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
    }

    private async initializeWasm(): Promise<void> {
        try {
            this.logToConsole('Loading WASM module...');
            
            // Check if WASM functions are available
            if (typeof (window as any).editorSetCanvas === 'undefined') {
                this.logToConsole('WASM functions not available - loading WASM module...');
                await this.loadWasmModule();
                // Check again after loading
                if (typeof (window as any).editorSetCanvas === 'undefined') {
                    throw new Error('WASM module loaded but functions not available');
                }
            }

            this.wasmInitialized = true;
            
            // Get initial map info from the global editor
            this.logToConsole('Getting initial map info...');
            const mapInfoResult = (window as any).editorGetMapInfo();
            if (!mapInfoResult.success) {
                throw new Error(mapInfoResult.error);
            }
            
            // Update local map data with initial info
            this.mapData = {
                name: mapInfoResult.data.filename || 'Default Map',
                width: mapInfoResult.data.width,
                height: mapInfoResult.data.height,
                tiles: {},
                map_units: []
            };
            
            // Get map bounds and tile dimensions from WASM
            this.updateMapBoundsFromWasm();
            
            // Calculate canvas size based on actual map bounds
            const canvasSize = this.calculateCanvasSize(this.mapData.width, this.mapData.height);
            
            // Bind WASM editor to canvas
            this.logToConsole('Binding WASM editor to canvas...');
            const canvasResult = (window as any).editorSetCanvas('map-canvas', canvasSize.width, canvasSize.height);
            if (!canvasResult.success) {
                throw new Error(canvasResult.error);
            }
            this.logToConsole(`Editor bound to canvas: ${canvasResult.data.canvasID}`);
            
            // Apply canvas size
            this.resizeCanvas(canvasSize.width, canvasSize.height);
            
            // ScrollOffset is already initialized with padding values
            
            this.updateEditorStatus('Ready');
            this.logToConsole('WASM editor initialized successfully');
            
            // Populate terrain palette with all available terrain types
            // await this.populateTerrainPalette();
            
            // Request initial editor refresh
            this.editorRefresh();
            
        } catch (error) {
            console.error('Failed to initialize WASM:', error);
            this.logToConsole(`WASM initialization failed: ${error}`);
            this.updateEditorStatus('WASM Error');
        }
    }

    private async loadWasmModule(): Promise<void> {
        return new Promise((resolve, reject) => {
            this.logToConsole('Loading editor.wasm...');
            
            // Create a new WebAssembly instance
            const go = new (window as any).Go();
            
            WebAssembly.instantiateStreaming(fetch('/static/wasm/editor.wasm'), go.importObject)
                .then((result) => {
                    this.logToConsole('WASM module instantiated, starting...');
                    go.run(result.instance);
                    
                    // Wait a bit for the module to register its functions
                    setTimeout(() => {
                        this.logToConsole('WASM module should be ready');
                        resolve();
                    }, 100);
                })
                .catch((error) => {
                    this.logToConsole(`WASM loading failed: ${error}`);
                    reject(error);
                });
        });
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
        
        // Update WASM editor brush terrain
        if (this.wasmInitialized) {
            try {
                const result = (window as any).editorSetBrushTerrain(terrain);
                if (!result.success) {
                    this.logToConsole(`WASM setBrushTerrain failed: ${result.error}`);
                }
            } catch (error) {
                this.logToConsole(`WASM setBrushTerrain error: ${error}`);
            }
        }
        
        const terrainNames = ['Unknown', 'Grass', 'Desert', 'Water', 'Mountain', 'Rock'];
        this.logToConsole(`Brush terrain set to: ${terrainNames[terrain]}`);
        this.updateBrushInfo();
        this.updateTerrainButtonSelection(terrain);
    }

    public setBrushSize(size: number): void {
        this.brushSize = size;
        
        // Update WASM editor brush size
        if (this.wasmInitialized) {
            try {
                const result = (window as any).editorSetBrushSize(size);
                if (!result.success) {
                    this.logToConsole(`WASM setBrushSize failed: ${result.error}`);
                }
            } catch (error) {
                this.logToConsole(`WASM setBrushSize error: ${error}`);
            }
        }
        
        const sizeNames = ['Single (1 hex)', 'Small (7 hexes)', 'Medium (19 hexes)', 'Large (37 hexes)', 'X-Large (61 hexes)', 'XX-Large (91 hexes)'];
        this.logToConsole(`Brush size set to: ${sizeNames[size]}`);
        this.updateBrushInfo();
    }
    
    public setShowGrid(showGrid: boolean): void {
        if (this.wasmInitialized) {
            try {
                const result = (window as any).editorSetShowGrid(showGrid);
                if (!result.success) {
                    this.logToConsole(`WASM setShowGrid failed: ${result.error}`);
                } else {
                    this.logToConsole(`Grid visibility set to: ${showGrid}`);
                    this.editorRefresh();
                }
            } catch (error) {
                this.logToConsole(`WASM setShowGrid error: ${error}`);
            }
        }
    }
    
    public setShowCoordinates(showCoordinates: boolean): void {
        if (this.wasmInitialized) {
            try {
                const result = (window as any).editorSetShowCoordinates(showCoordinates);
                if (!result.success) {
                    this.logToConsole(`WASM setShowCoordinates failed: ${result.error}`);
                } else {
                    this.logToConsole(`Coordinate visibility set to: ${showCoordinates}`);
                    this.editorRefresh();
                }
            } catch (error) {
                this.logToConsole(`WASM setShowCoordinates error: ${error}`);
            }
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


    public renderEditor(width: number, height: number): void {
        this.logToConsole(`Rendering map at ${width}×${height}...`);
        // TODO: Implement WASM rendering
        
        // WASM now handles rendering automatically when map changes
        this.logToConsole('Map rendering handled by WASM');
    }
    
    // Request a refresh from the WASM editor (non-blocking)
    private editorRefresh(): void {
        if (!this.wasmInitialized) {
            return;
        }
        
        try {
            const result = (window as any).editorRender();
            if (!result.success) {
                this.logToConsole(`Editor refresh failed: ${result.error}`);
            }
        } catch (error) {
            this.logToConsole(`Editor refresh error: ${error}`);
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
        // TODO: Implement with WASM
    }

    public createTestPattern(): void {
        this.logToConsole('Creating test pattern...');
        // TODO: Implement test pattern generation
    }

    public createIslandMap(): void {
        this.logToConsole('Creating island map...');
        // TODO: Implement island generation
    }

    public createMountainRidge(): void {
        this.logToConsole('Creating mountain ridge...');
        // TODO: Implement mountain ridge generation
    }

    public showTerrainStats(): void {
        this.logToConsole('Terrain statistics:');
        this.logToConsole('- Grass: 0 tiles');
        this.logToConsole('- Desert: 0 tiles');
        this.logToConsole('- Water: 0 tiles');
        this.logToConsole('- Mountain: 0 tiles');
        this.logToConsole('- Rock: 0 tiles');
        // TODO: Calculate actual stats from map data
    }

    public randomizeTerrain(): void {
        this.logToConsole('Randomizing terrain...');
        // TODO: Implement terrain randomization
    }

    // Update map bounds and tile dimensions from WASM
    private updateMapBoundsFromWasm(): void {
        if (!this.wasmInitialized || !this.mapData) {
            return;
        }
        
        try {
            // Get map bounds and tile dimensions from WASM
            const boundsResult = (window as any).editorGetMapBounds();
            if (boundsResult && boundsResult.success && boundsResult.data) {
                const bounds = boundsResult.data;
                
                // Update tile dimensions if available
                const tileDims = bounds.tileDimensions
                if (tileDims.tileWidth && tileDims.tileHeight && tileDims.yIncrement) {
                    this.tileDimensions = tileDims
                    this.logToConsole(`Tile dimensions updated: ${JSON.stringify(this.tileDimensions)}`);
                }
                
                this.setMapBounds(bounds.bounds)
                
                this.logToConsole(`Map bounds updated: Q(${bounds.minQ}-${bounds.maxQ}) R(${bounds.minR}-${bounds.maxR}) Starting(${bounds.startingCoord.q},${bounds.startingCoord.r})`);
            } else {
                this.logToConsole(`Failed to get map bounds from WASM: ${boundsResult?.error || 'Unknown error'}`);
            }
        } catch (error) {
            this.logToConsole(`Error getting map bounds from WASM: ${error}`);
        }
    }

    private setMapBounds(bounds: MapBounds)  {
        // Update mapData with bounds information
        this.mapBounds = bounds
    }

    // Canvas management methods
    private initializeCanvas(): void {
        if (!this.mapCanvas) return;
        
        // Canvas is now bound to WASM editor and will be rendered automatically
        this.logToConsole('Canvas initialized');
    }

    // resizeCanvas updates the canvas size in the DOM to match WASM-calculated dimensions
    private resizeCanvas(width: number, height: number): void {
        if (!this.mapCanvas) return;
        
        this.logToConsole(`Time: ${performance.now()} Resizing canvas DOM to ${width}x${height}`);
        
        // Update canvas DOM element size (this clears the canvas content)
        this.mapCanvas.width = width;
        this.mapCanvas.height = height;
        this.canvasSize = {width, height};
        
        // Update canvas CSS size for proper display
        this.mapCanvas.style.width = `${width}px`;
        this.mapCanvas.style.height = `${height}px`;
        
        // Tell WASM to update its viewport to match DOM canvas size
        if (this.wasmInitialized) {
            try {
                const result = (window as any).editorSetViewPort(this.scrollOffset.x, this.scrollOffset.y, width, height);
                if (result.success) {
                    this.logToConsole(`Time: ${performance.now()} WASM viewport updated to ${width}x${height}`);
                    
                    // Force a re-render with proper game rendering (not simplified hexagons)
                    const renderResult = (window as any).editorRender();
                    if (renderResult.success) {
                        this.logToConsole(`Time: ${performance.now()} Map re-rendered with proper terrain images`);
                    } else {
                        this.logToConsole(`Re-render failed: ${renderResult.error}`);
                    }
                } else {
                    this.logToConsole(`Failed to update WASM viewport: ${result.error}`);
                }
            } catch (error) {
                this.logToConsole(`Error updating WASM viewport: ${error}`);
            }
        }
        
        this.logToConsole(`Time: ${performance.now()} Canvas resized to ${width}x${height}`);
    }

    // renderMapCanvas() method removed - WASM now pushes updates directly to canvas

    private handleCanvasClick(event: MouseEvent): void {
        if (!this.mapCanvas || !this.mapData) return;

        const rect = this.mapCanvas.getBoundingClientRect();
        const x = event.clientX - rect.left;
        const y = event.clientY - rect.top;
        
        const coords = this.pixelToHex(x, y);
        if (coords) {
            // Paint at clicked coordinates (works both within and outside current map bounds)
            this.paintHexAtCoords(coords);
        }
    }

    private handleCanvasMouseMove(event: MouseEvent): void {
        if (!this.mapCanvas) return;

        const rect = this.mapCanvas.getBoundingClientRect();
        const x = event.clientX - rect.left;
        const y = event.clientY - rect.top;

        // console.log(`Event XY: (${event.clientX}, ${event.clientY}), Rect XY: (${rect.left}, ${rect.top}), localXY: (${x}, ${y})`)
        
        // Use client-side coordinate conversion for performance
        const coords = this.clientPixelToCoords(x, y);
        
        // Update coordinate inputs with cube coordinates
        const rowInput = document.getElementById('paint-row') as HTMLInputElement;
        const colInput = document.getElementById('paint-col') as HTMLInputElement;
        
        if (rowInput) rowInput.value = coords.r.toString();
        if (colInput) colInput.value = coords.q.toString();
        
        // Show if cursor is outside map bounds (for potential extension)
        if (!coords.withinBounds) {
            // Visual feedback for extension area
            this.mapCanvas.style.cursor = 'pointer';
        } else {
            this.mapCanvas.style.cursor = 'default';
        }
    }

    private hexToPixel(q: number, r: number): {x: number, y: number} | null {
        if (!this.mapCanvas || !this.mapData) return null;

          const result = (window as any).editorHexToPixel(q, r);
          if (result && result.success && result.data) {
              const { x, y } = result.data;
              this.logToConsole(`HexCoord (${q}, ${r}) -> XY (${x}, ${y}))`);
              return result.data
          }
          return null
    }

    private pixelToHex(mx: number, my: number): {row: number, col: number, cubeQ: number, cubeR: number} | null {
        if (!this.mapCanvas || !this.mapData) return null;

        const x = mx - this.scrollOffset.x;
        const y = my - this.scrollOffset.y;

        console.log("Doing pixelToHex, x, y: ", x, y, this.mapData);
        console.log("Mouse X/Y: ", mx, my);
        
        // Use WASM coordinate conversion for maximum accuracy
        if (this.wasmInitialized && (window as any).editorPixelToCoords) {
            try {
                const result = (window as any).editorPixelToCoords(x, y);
                if (result && result.success && result.data) {
                    const { row, col, cubeQ, cubeR, withinBounds } = result.data;
                    const rev = this.hexToPixel(cubeQ, cubeR)!
                    this.logToConsole(`Pixel (${x}, ${y}) -> RowCal (${row}, ${col}) = QR(${cubeQ}, ${cubeR}) [WASM conversion, within bounds], Reverse: ${rev.x}, ${rev.y}`);
                    return result.data
                }
            } catch (error) {
                console.warn('WASM coordinate conversion failed, falling back to browser calculation:', error);
            }
        }
        
        // Fallback to browser-side calculation using proper coordinate conversion
        if (!this.tileDimensions) {
            this.logToConsole('Tile dimensions not available for fallback coordinate conversion');
            return null;
        }
        
        try {
            // Use our proper coordinate conversion methods
            const cubeCoord = this.xyToQR(x, y);
            
            // Convert cube coordinates to row/col for backward compatibility
            // Note: This assumes a simple mapping where row=r and col=q
            const row = cubeCoord.r;
            const col = cubeCoord.q;
            
            // Check if within map bounds for UI purposes
            const withinBounds = this.mapData ? 
                (row >= 0 && row < this.mapData.height && col >= 0 && col < this.mapData.width) : false;
            
            if (withinBounds) {
                this.logToConsole(`Pixel (${x}, ${y}) -> Hex (${row}, ${col}) Q=${cubeCoord.q} R=${cubeCoord.r} [Browser fallback with proper hex math]`);
                return { row, col, cubeQ: cubeCoord.q, cubeR: cubeCoord.r };
            }
            
            this.logToConsole(`Pixel (${x}, ${y}) -> Out of bounds (${row}, ${col}) Q=${cubeCoord.q} R=${cubeCoord.r} [Browser fallback with proper hex math]`);
            return null;
        } catch (error) {
            this.logToConsole(`Coordinate conversion failed: ${error}`);
            return null;
        }
    }

    private paintHexAtCoords(coord: {row: number, col: number, cubeQ: number, cubeR: number}): void {
        try {
            // Get current terrain type from selected terrain button
            const selectedTerrainButton = document.querySelector('.terrain-button.bg-blue-100') as HTMLElement;
            const terrainType = selectedTerrainButton ? 
                parseInt(selectedTerrainButton.getAttribute('data-terrain') || '1') : 1;
            
            // Get current brush size from dropdown
            const brushSizeSelect = document.getElementById('brush-size') as HTMLSelectElement;
            const brushSize = brushSizeSelect ? parseInt(brushSizeSelect.value) : 0;
            
            // Use stateless WASM editor function
            const result = (window as any).editorSetTilesAt(coord.cubeQ, coord.cubeR, terrainType, brushSize);
            
            if (result.success) {
                // Update local map data to stay in sync (only the center hex for simplicity)
                // result.data is a list of coords on which the given terrain type was set
                const mapData = this.mapData
                if (mapData) {
                    result.data.coords.forEach((coord: {q: number, r: number}) => {
                      const key = `${coord.q},${coord.r}`;
                      if (terrainType <= 0) {
                        delete mapData.tiles[key]
                      } else {
                        mapData.tiles[key] = { tileType: terrainType };
                      }
                    })
                }

                this.setMapBounds(result.data.newBounds)

                // Now see if bounds have changed
                const newWidth = Math.max(this.scrollOffset.x * 2 + (result.data.newBounds.MaxX - result.data.newBounds.MinX), 400)
                const newHeight = Math.max(this.scrollOffset.y * 2 + (result.data.newBounds.MaxY - result.data.newBounds.MinY), 300)

                if (this.canvasSize.width != newWidth || this.canvasSize.height != newWidth) {
                  this.resizeCanvas(newWidth, newHeight)
                }
                
                // Log the paint action with brush info
                const sizeNames = ['Single', 'Small', 'Medium', 'Large', 'X-Large', 'XX-Large'];
                const sizeName = sizeNames[brushSize] || 'Unknown';
                // this.logToConsole(`Set terrain ${terrainType} at (Q=${col}, R=${row}) with ${sizeName} brush (radius=${brushSize})`);
            } else {
                this.logToConsole(`Set terrain failed: ${result.error}`);
            }
        } catch (error) {
            this.logToConsole(`Paint error: ${error}`);
        }
    }

    private async saveMap(): Promise<void> {
        if (!this.mapData) {
            this.toastManager?.showToast('Error', 'No map data to save', 'error');
            return;
        }

        try {
            this.logToConsole('Saving map...');
            this.updateEditorStatus('Saving...');

            const url = this.isNewMap ? '/api/maps' : `/api/maps/${this.currentMapId}`;
            const method = this.isNewMap ? 'POST' : 'PUT';

            const response = await fetch(url, {
                method,
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(this.mapData),
            });

            if (response.ok) {
                const result = await response.json();
                this.logToConsole('Map saved successfully');
                this.updateEditorStatus('Saved');
                this.toastManager?.showToast('Success', 'Map saved successfully', 'success');
                
                // If this was a new map, update the current map ID
                if (this.isNewMap && result.id) {
                    this.currentMapId = result.id;
                    this.isNewMap = false;
                    // Update URL without reload
                    history.replaceState(null, '', `/maps/${result.id}/edit`);
                }
            } else {
                throw new Error(`Save failed: ${response.statusText}`);
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

    private createCanvasComponent() {
        const template = document.getElementById('canvas-panel-template');
        if (!template) {
            console.error('Canvas panel template not found');
            return { element: document.createElement('div'), init: () => {}, dispose: () => {} };
        }
        
        const container = template.cloneNode(true) as HTMLElement;
        container.style.display = 'block';
        container.style.width = '100%';
        container.style.height = '100%';
        
        return {
            element: container,
            init: () => {
                // Find the canvas element within this cloned template and update references
                const canvasElement = container.querySelector('#map-canvas') as HTMLCanvasElement;
                if (canvasElement) {
                    this.mapCanvas = canvasElement;
                    this.canvasContext = canvasElement.getContext('2d');
                }
                
                // Update the canvas container reference
                const canvasContainer = container.querySelector('#editor-canvas-container');
                if (canvasContainer) {
                    this.editorCanvas = canvasContainer as HTMLElement;
                }
                
                // Rebind canvas events
                if (this.mapCanvas) {
                    this.mapCanvas.addEventListener('click', this.handleCanvasClick.bind(this));
                    this.mapCanvas.addEventListener('mousemove', this.handleCanvasMouseMove.bind(this));
                }
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

        // Add canvas panel first (center)
        this.dockview.addPanel({
            id: 'canvas',
            component: 'canvas',
            title: '🗺️ Map Canvas'
        });

        // Add tools panel to the left of canvas
        this.dockview.addPanel({
            id: 'tools',
            component: 'tools',
            title: '🎨 Tools & Terrain',
            position: { direction: 'left', referencePanel: 'canvas' }
        });

        // Add advanced tools panel to the right of canvas
        this.dockview.addPanel({
            id: 'advancedTools',
            component: 'advancedTools',
            title: '🔧 Advanced Tools',
            position: { direction: 'right', referencePanel: 'canvas' }
        });

        // Add console panel below canvas
        this.dockview.addPanel({
            id: 'console',
            component: 'console',
            title: '💻 Console',
            position: { direction: 'below', referencePanel: 'canvas' }
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

    public destroy(): void {
        // Save layout before destroying
        this.saveDockviewLayout();
        
        // Dispose dockview
        if (this.dockview) {
            this.dockview.dispose();
        }
        
        // Destroy Phaser editor if it exists
        if (this.phaserEditor) {
            this.phaserEditor.destroy();
        }
    }
    
    // Phaser test methods
    public initializePhaser(): void {
        try {
            this.logToConsole('Initializing Phaser editor...');
            
            // Hide the existing canvas
            if (this.mapCanvas) {
                this.mapCanvas.style.display = 'none';
            }
            
            // Create container for Phaser
            const canvasContainer = document.getElementById('editor-canvas-container');
            if (!canvasContainer) {
                throw new Error('Canvas container not found');
            }
            
            // Create Phaser container div
            let phaserContainer = document.getElementById('phaser-container');
            if (!phaserContainer) {
                phaserContainer = document.createElement('div');
                phaserContainer.id = 'phaser-container';
                phaserContainer.style.width = '100%';
                phaserContainer.style.height = '100%';
                phaserContainer.style.minWidth = '800px';
                phaserContainer.style.minHeight = '600px';
                canvasContainer.appendChild(phaserContainer);
            }
            
            // Initialize Phaser editor
            this.phaserEditor = new PhaserMapEditor('phaser-container');
            
            // Set up event handlers
            this.phaserEditor.onTileClick((q, r) => {
                this.logToConsole(`Phaser tile clicked: Q=${q}, R=${r}`);
            });
            
            this.phaserEditor.onMapChange(() => {
                this.logToConsole('Phaser map changed');
            });
            
            this.logToConsole('Phaser editor initialized successfully!');
            this.toastManager?.showToast('Phaser', 'Phaser editor initialized', 'success');
            
        } catch (error) {
            this.logToConsole(`Failed to initialize Phaser: ${error}`);
            this.toastManager?.showToast('Error', 'Failed to initialize Phaser', 'error');
        }
    }
    
    public testPhaserPattern(): void {
        if (!this.phaserEditor) {
            this.logToConsole('Phaser editor not initialized');
            this.toastManager?.showToast('Error', 'Initialize Phaser first', 'error');
            return;
        }
        
        try {
            this.logToConsole('Creating test pattern with Phaser...');
            
            // Test negative coordinate support
            this.phaserEditor.createTestPattern();
            
            // Test individual tile painting
            this.phaserEditor.paintTile(-5, -5, 1, 0); // Grass at negative coordinates
            this.phaserEditor.paintTile(5, 5, 2, 1);   // Desert at positive coordinates
            this.phaserEditor.paintTile(-3, 3, 3, 2);  // Water at mixed coordinates
            
            // Test brush painting
            this.phaserEditor.setBrushSize(1);
            this.phaserEditor.paintTile(0, 0, 16, 0, 1); // Mountain with small brush
            
            this.logToConsole('Test pattern created successfully!');
            this.toastManager?.showToast('Success', 'Test pattern created', 'success');
            
        } catch (error) {
            this.logToConsole(`Failed to create test pattern: ${error}`);
            this.toastManager?.showToast('Error', 'Failed to create test pattern', 'error');
        }
    }
    
    public destroyPhaser(): void {
        if (this.phaserEditor) {
            this.logToConsole('Destroying Phaser editor...');
            
            this.phaserEditor.destroy();
            this.phaserEditor = null;
            
            // Remove Phaser container
            const phaserContainer = document.getElementById('phaser-container');
            if (phaserContainer) {
                phaserContainer.remove();
            }
            
            // Show the original canvas
            if (this.mapCanvas) {
                this.mapCanvas.style.display = 'block';
            }
            
            this.logToConsole('Phaser editor destroyed');
            this.toastManager?.showToast('Info', 'Phaser editor destroyed', 'info');
        } else {
            this.logToConsole('No Phaser editor to destroy');
        }
    }
}

// Initialize the editor when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    new MapEditorPage();
});
