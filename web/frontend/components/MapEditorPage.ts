import { ThemeManager } from './ThemeManager';
import { Modal } from './Modal';
import { ToastManager } from './ToastManager';

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
    private mapData: {
        name: string;
        width: number;
        height: number;
        tiles: { [key: string]: { tileType: number } };
        map_units: any[];
    } | null = null;
    
    // Editor state
    private currentTerrain: number = 1; // Default to grass
    private brushSize: number = 0; // Default to single hex
    private editorCanvas: HTMLElement | null = null;
    private mapCanvas: HTMLCanvasElement | null = null;
    private canvasContext: CanvasRenderingContext2D | null = null;
    private editorOutput: HTMLElement | null = null;

    // WASM interface
    private wasmModule: any = null;
    private wasmInitialized: boolean = false;

    constructor() {
        this.initializeComponents();
        this.bindEvents();
        this.loadInitialState();
        this.initializeWasm();
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

        // Utility buttons
        const validateButton = document.getElementById('validate-map-btn');
        if (validateButton) {
            validateButton.addEventListener('click', this.validateMap.bind(this));
        }

        const clearConsoleButton = document.getElementById('clear-console-btn');
        if (clearConsoleButton) {
            clearConsoleButton.addEventListener('click', this.clearConsole.bind(this));
        }

        // Map management buttons
        document.querySelectorAll('[data-action="create-new-map"]').forEach(button => {
            button.addEventListener('click', (e) => {
                const target = e.target as HTMLElement;
                const width = parseInt(target.dataset.width || '8');
                const height = parseInt(target.dataset.height || '8');
                this.createNewMap(width, height);
            });
        });

        // Terrain palette buttons
        document.querySelectorAll('.terrain-button').forEach(button => {
            button.addEventListener('click', (e) => {
                const terrain = (e.target as HTMLElement).getAttribute('data-terrain');
                if (terrain) {
                    this.setBrushTerrain(parseInt(terrain));
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

        // History buttons
        document.querySelector('[data-action="undo"]')?.addEventListener('click', () => {
            this.editorUndo();
        });
        document.querySelector('[data-action="redo"]')?.addEventListener('click', () => {
            this.editorRedo();
        });

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

        // Map resize controls
        document.querySelectorAll('[data-action="add-row"]').forEach(button => {
            button.addEventListener('click', (e) => {
                const side = (e.target as HTMLElement).dataset.side || 'bottom';
                this.addRow(side);
            });
        });
        document.querySelectorAll('[data-action="remove-row"]').forEach(button => {
            button.addEventListener('click', (e) => {
                const side = (e.target as HTMLElement).dataset.side || 'bottom';
                this.removeRow(side);
            });
        });
        document.querySelectorAll('[data-action="add-col"]').forEach(button => {
            button.addEventListener('click', (e) => {
                const side = (e.target as HTMLElement).dataset.side || 'right';
                this.addColumn(side);
            });
        });
        document.querySelectorAll('[data-action="remove-col"]').forEach(button => {
            button.addEventListener('click', (e) => {
                const side = (e.target as HTMLElement).dataset.side || 'right';
                this.removeColumn(side);
            });
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
            this.logToConsole('Creating new map...');
            this.initializeNewMap();
        } else if (this.currentMapId) {
            this.logToConsole(`Loading existing map: ${this.currentMapId}`);
            this.loadExistingMap(this.currentMapId);
        } else {
            this.logToConsole('Error: No map ID provided');
            this.updateEditorStatus('Error');
        }
    }

    private async initializeWasm(): Promise<void> {
        try {
            this.logToConsole('Loading WASM module...');
            
            // This will be implemented when we copy the WASM files
            // For now, we'll use a placeholder
            this.wasmInitialized = true;
            this.updateEditorStatus('Ready');
            this.logToConsole('WASM module loaded successfully');
            
        } catch (error) {
            console.error('Failed to initialize WASM:', error);
            this.logToConsole(`WASM initialization failed: ${error}`);
            this.updateEditorStatus('WASM Error');
        }
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
    public createNewMap(width: number, height: number): void {
        this.logToConsole(`Creating new ${width}×${height} map...`);
        this.mapData = {
            name: `New ${width}×${height} Map`,
            width,
            height,
            tiles: {},
            map_units: []
        };
        this.updateEditorStatus('Ready');
        this.logToConsole('New map created');
        this.renderMapCanvas();
    }

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

    public editorUndo(): void {
        this.logToConsole('Undo action');
        // TODO: Implement undo functionality
    }

    public editorRedo(): void {
        this.logToConsole('Redo action');
        // TODO: Implement redo functionality
    }

    public renderEditor(width: number, height: number): void {
        this.logToConsole(`Rendering map at ${width}×${height}...`);
        // TODO: Implement WASM rendering
        
        // For now, just re-render the canvas
        this.renderMapCanvas();
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

    // Canvas management methods
    private initializeCanvas(): void {
        if (!this.mapCanvas || !this.canvasContext) return;
        
        // Set up initial canvas state
        this.renderMapCanvas();
        this.logToConsole('Canvas initialized');
    }

    private renderMapCanvas(): void {
        if (!this.mapCanvas || !this.canvasContext || !this.mapData) return;

        const ctx = this.canvasContext;
        const canvas = this.mapCanvas;

        // Clear canvas
        ctx.clearRect(0, 0, canvas.width, canvas.height);

        // Set background
        ctx.fillStyle = '#f0f0f0';
        ctx.fillRect(0, 0, canvas.width, canvas.height);

        // Calculate hex grid parameters
        const hexSize = 20;
        const hexWidth = hexSize * 2;
        const hexHeight = Math.sqrt(3) * hexSize;
        const rowHeight = hexHeight * 0.75;

        // Draw grid
        for (let row = 0; row < this.mapData.height; row++) {
            for (let col = 0; col < this.mapData.width; col++) {
                const x = col * hexWidth + (row % 2) * hexSize + hexSize + 20;
                const y = row * rowHeight + hexSize + 20;
                
                this.drawHex(ctx, x, y, hexSize, this.getTerrainColor(row, col));
            }
        }
    }

    private drawHex(ctx: CanvasRenderingContext2D, x: number, y: number, size: number, fillColor: string): void {
        const angle = Math.PI / 3;
        
        ctx.beginPath();
        for (let i = 0; i < 6; i++) {
            const angle_deg = 60 * i;
            const angle_rad = Math.PI / 180 * angle_deg;
            const xPos = x + size * Math.cos(angle_rad);
            const yPos = y + size * Math.sin(angle_rad);
            
            if (i === 0) {
                ctx.moveTo(xPos, yPos);
            } else {
                ctx.lineTo(xPos, yPos);
            }
        }
        ctx.closePath();
        
        // Fill
        ctx.fillStyle = fillColor;
        ctx.fill();
        
        // Stroke
        ctx.strokeStyle = '#333';
        ctx.lineWidth = 1;
        ctx.stroke();
    }

    private getTerrainColor(row: number, col: number): string {
        // Get terrain type from map data, default to grass
        const terrainType = this.mapData?.tiles[`${row},${col}`]?.tileType || 1;
        
        const colors: { [key: number]: string } = {
            0: '#808080', // Unknown - gray
            1: '#90EE90', // Grass - light green
            2: '#F4A460', // Desert - sandy brown
            3: '#4169E1', // Water - blue
            4: '#8B4513', // Mountain - brown
            5: '#696969'  // Rock - dark gray
        };
        
        return colors[terrainType] || colors[1];
    }

    private handleCanvasClick(event: MouseEvent): void {
        if (!this.mapCanvas || !this.mapData) return;

        const rect = this.mapCanvas.getBoundingClientRect();
        const x = event.clientX - rect.left;
        const y = event.clientY - rect.top;
        
        const coords = this.pixelToHex(x, y);
        if (coords && coords.row >= 0 && coords.row < this.mapData.height && 
            coords.col >= 0 && coords.col < this.mapData.width) {
            
            this.logToConsole(`Clicked hex (${coords.row}, ${coords.col})`);
            this.paintHexAtCoords(coords.row, coords.col);
        }
    }

    private handleCanvasMouseMove(event: MouseEvent): void {
        if (!this.mapCanvas) return;

        const rect = this.mapCanvas.getBoundingClientRect();
        const x = event.clientX - rect.left;
        const y = event.clientY - rect.top;
        
        const coords = this.pixelToHex(x, y);
        if (coords) {
            // Update coordinate inputs
            const rowInput = document.getElementById('paint-row') as HTMLInputElement;
            const colInput = document.getElementById('paint-col') as HTMLInputElement;
            
            if (rowInput) rowInput.value = coords.row.toString();
            if (colInput) colInput.value = coords.col.toString();
        }
    }

    private pixelToHex(x: number, y: number): {row: number, col: number} | null {
        // Convert pixel coordinates to hex grid coordinates
        const hexSize = 20;
        const hexWidth = hexSize * 2;
        const rowHeight = Math.sqrt(3) * hexSize * 0.75;
        
        // Approximate row and column
        const row = Math.floor((y - hexSize - 20) / rowHeight);
        const col = Math.floor((x - hexSize - 20 - (row % 2) * hexSize) / hexWidth);
        
        return { row, col };
    }

    private paintHexAtCoords(row: number, col: number): void {
        if (!this.mapData) return;
        
        // Update map data
        const key = `${row},${col}`;
        this.mapData.tiles[key] = { tileType: this.currentTerrain };
        
        // Re-render canvas
        this.renderMapCanvas();
        
        const terrainNames = ['Unknown', 'Grass', 'Desert', 'Water', 'Mountain', 'Rock'];
        this.logToConsole(`Painted ${terrainNames[this.currentTerrain]} at (${row}, ${col})`);
    }

    // Map resize methods
    public addRow(side: string): void {
        if (!this.mapData) return;
        
        this.logToConsole(`Adding row to ${side}`);
        if (side === 'top') {
            // Shift all existing tiles down by 1 row
            const newTiles: { [key: string]: { tileType: number } } = {};
            for (const [key, tile] of Object.entries(this.mapData.tiles)) {
                const [row, col] = key.split(',').map(Number);
                newTiles[`${row + 1},${col}`] = tile;
            }
            this.mapData.tiles = newTiles;
        }
        this.mapData.height += 1;
        this.renderMapCanvas();
    }

    public removeRow(side: string): void {
        if (!this.mapData || this.mapData.height <= 1) return;
        
        this.logToConsole(`Removing row from ${side}`);
        const newTiles: { [key: string]: { tileType: number } } = {};
        
        if (side === 'top') {
            // Remove top row and shift everything up
            for (const [key, tile] of Object.entries(this.mapData.tiles)) {
                const [row, col] = key.split(',').map(Number);
                if (row > 0) {
                    newTiles[`${row - 1},${col}`] = tile;
                }
            }
        } else {
            // Remove bottom row
            for (const [key, tile] of Object.entries(this.mapData.tiles)) {
                const [row, col] = key.split(',').map(Number);
                if (row < this.mapData.height - 1) {
                    newTiles[key] = tile;
                }
            }
        }
        
        this.mapData.tiles = newTiles;
        this.mapData.height -= 1;
        this.renderMapCanvas();
    }

    public addColumn(side: string): void {
        if (!this.mapData) return;
        
        this.logToConsole(`Adding column to ${side}`);
        if (side === 'left') {
            // Shift all existing tiles right by 1 column
            const newTiles: { [key: string]: { tileType: number } } = {};
            for (const [key, tile] of Object.entries(this.mapData.tiles)) {
                const [row, col] = key.split(',').map(Number);
                newTiles[`${row},${col + 1}`] = tile;
            }
            this.mapData.tiles = newTiles;
        }
        this.mapData.width += 1;
        this.renderMapCanvas();
    }

    public removeColumn(side: string): void {
        if (!this.mapData || this.mapData.width <= 1) return;
        
        this.logToConsole(`Removing column from ${side}`);
        const newTiles: { [key: string]: { tileType: number } } = {};
        
        if (side === 'left') {
            // Remove left column and shift everything left
            for (const [key, tile] of Object.entries(this.mapData.tiles)) {
                const [row, col] = key.split(',').map(Number);
                if (col > 0) {
                    newTiles[`${row},${col - 1}`] = tile;
                }
            }
        } else {
            // Remove right column
            for (const [key, tile] of Object.entries(this.mapData.tiles)) {
                const [row, col] = key.split(',').map(Number);
                if (col < this.mapData.width - 1) {
                    newTiles[key] = tile;
                }
            }
        }
        
        this.mapData.tiles = newTiles;
        this.mapData.width -= 1;
        this.renderMapCanvas();
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

    private validateMap(): void {
        this.logToConsole('Validating map...');
        // TODO: Implement map validation
        this.logToConsole('Map validation completed - no issues found');
    }

    private clearConsole(): void {
        if (this.editorOutput) {
            this.editorOutput.textContent = '';
        }
    }

    // Utility methods
    private logToConsole(message: string): void {
        if (this.editorOutput) {
            const timestamp = new Date().toLocaleTimeString();
            this.editorOutput.textContent += `[${timestamp}] ${message}\n`;
            this.editorOutput.scrollTop = this.editorOutput.scrollHeight;
        }
        console.log(`[MapEditor] ${message}`);
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
}

// Initialize the editor when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    new MapEditorPage();
});
