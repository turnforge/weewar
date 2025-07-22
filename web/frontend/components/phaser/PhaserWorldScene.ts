import * as Phaser from 'phaser';
import { hexToPixel, pixelToHex, HexCoord, PixelCoord } from './hexUtils';

export class PhaserWorldScene extends Phaser.Scene {
    private tileWidth: number = 64;
    private tileHeight: number = 64;
    private yIncrement: number = 48; // 3/4 * tileHeight for pointy-topped hexes
    
    private tiles: Map<string, Phaser.GameObjects.Sprite> = new Map();
    private units: Map<string, Phaser.GameObjects.Sprite> = new Map();
    private gridGraphics: Phaser.GameObjects.Graphics | null = null;
    private coordinateTexts: Map<string, Phaser.GameObjects.Text> = new Map();
    
    private showGrid: boolean = false;
    private showCoordinates: boolean = false;
    
    // Theme management - initialize with current theme state
    private isDarkTheme: boolean = document.documentElement.classList.contains('dark');
    
    // Camera controls
    private cursors: Phaser.Types.Input.Keyboard.CursorKeys | null = null;
    private wasdKeys: any = null;
    private zoomSpeed: number = 0.01;
    private panSpeed: number = 100;
    
    // Mouse interaction
    private isMouseDown: boolean = false;
    private lastPointerPosition: { x: number; y: number } | null = null;
    private isPaintMode: boolean = false;
    private hasDragged: boolean = false;
    private paintModeStartPosition: { x: number; y: number } | null = null;
    
    // Asset loading
    private terrainsLoaded: boolean = false;
    private unitsLoaded: boolean = false;
    private sceneReadyCallback: (() => void) | null = null;
    private assetsReadyPromise: Promise<void> | null = null;
    private assetsReadyResolver: (() => void) | null = null;
    
    constructor(config?: string | Phaser.Types.Scenes.SettingsConfig) {
        super(config || { key: 'PhaserWorldScene' });
    }
    
    preload() {
        // Set up assets ready promise
        this.assetsReadyPromise = new Promise<void>((resolve) => {
            this.assetsReadyResolver = resolve;
        });
        
        // Track when all assets are loaded
        this.load.on('complete', () => {
            console.log('[PhaserWorldScene] All assets loaded successfully');
            this.terrainsLoaded = true;
            this.unitsLoaded = true;
            if (this.assetsReadyResolver) {
                this.assetsReadyResolver();
            }
        });
        
        // Load terrain assets
        this.loadTerrainAssets();
        
        // Load unit assets (for future use)
        this.loadUnitAssets();
    }
    
    create() {
        // Initialize graphics for grid
        this.gridGraphics = this.add.graphics();
        
        // Set up camera controls
        this.setupCameraControls();
        
        // Set up input handling
        this.setupInputHandling();
        
        // Set camera bounds to allow infinite scrolling
        this.cameras.main.setBounds(-10000, -10000, 20000, 20000);
        
        // Initialize grid and coordinates display
        this.updateGridDisplay();
        this.setShowCoordinates(this.showCoordinates);
        
        // Set initial theme
        this.updateTheme();
        
        console.log('[PhaserWorldScene] Scene created successfully');
        
        // Trigger scene ready callback if set
        if (this.sceneReadyCallback) {
            this.sceneReadyCallback();
            this.sceneReadyCallback = null;
        }
    }
    
    private loadTerrainAssets() {
        // City/Player terrains (have color variations 0-12 for different players)
        const cityTerrains = [1, 2, 3, 16, 20]; // Land Base, Naval Base, Airport Base, Missile Silo, Mines
        
        // Nature terrains (only have default texture, no player ownership)
        const natureTerrains = [4, 5, 6, 7, 8, 9, 10, 12, 14, 15, 17, 18, 19, 21, 22, 23, 25, 26];
        
        // Load city terrains with all color variations
        cityTerrains.forEach(type => {
            for (let color = 0; color <= 12; color++) {
                const assetPath = `/static/assets/v1/Tiles/${type}/${color}.png`;
                this.load.image(`terrain_${type}_${color}`, assetPath);
                if (color === 0) {
                    this.load.image(`terrain_${type}`, assetPath); // Default alias
                }
            }
        });
        
        // Load nature terrains with only default texture
        natureTerrains.forEach(type => {
            const assetPath = `/static/assets/v1/Tiles/${type}/0.png`;
            this.load.image(`terrain_${type}`, assetPath);
            this.load.image(`terrain_${type}_0`, assetPath); // Color 0 alias for consistency
        });
        
        console.log(`[PhaserWorldScene] Loading city terrain assets: ${cityTerrains.join(', ')} (with colors)`);
        console.log(`[PhaserWorldScene] Loading nature terrain assets: ${natureTerrains.join(', ')} (default only)`);
    }
    
    private loadUnitAssets() {
        // Load all available unit assets with player colors
        const unitTypes = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 24, 25, 26, 27, 28, 29];
        
        unitTypes.forEach(type => {
            for (let color = 0; color <= 12; color++) {
                const assetPath = `/static/assets/v1/Units/${type}/${color}.png`;
                this.load.image(`unit_${type}_${color}`, assetPath);
                if (color === 0) {
                    this.load.image(`unit_${type}`, assetPath); // Default alias
                }
            }
        });
        
        console.log(`[PhaserWorldScene] Loading unit assets: ${unitTypes.join(', ')} (with colors)`);
    }
    
    private setupCameraControls() {
        // Create cursor keys for camera movement
        // this.cursors = this.input.keyboard!.createCursorKeys();
        
        // Add WASD keys  
        // this.wasdKeys = this.input.keyboard!.addKeys('W,S,A,D');
        
        // Set up document-level keydown listener to handle input context properly
        /*
        document.addEventListener('keydown', (event: KeyboardEvent) => {
            // If user is focused on an input field, don't let Phaser prevent default for arrow keys
            if (this.isInInputContext()) {
                const arrowKeys = ['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown'];
                if (arrowKeys.includes(event.key)) {
                    // Stop the event from reaching Phaser's keyboard handlers
                    event.stopImmediatePropagation();
                }
            }
        }, true); // Use capture phase to intercept before Phaser
       */
        
        // Mouse wheel zoom - zoom around cursor position
        this.input.on('wheel', (pointer: Phaser.Input.Pointer, gameObjects: Phaser.GameObjects.GameObject[], deltaX: number, deltaY: number) => {
            const camera = this.cameras.main;
            const oldZoom = camera.zoom;
            const zoomFactor = deltaY > 0 ? 1 - this.zoomSpeed : 1 + this.zoomSpeed;
            const newZoom = Phaser.Math.Clamp(oldZoom * zoomFactor, 0.1, 3);
            
            // Calculate world coordinates under mouse cursor before zoom
            const worldX = camera.scrollX + (pointer.x - camera.centerX) / oldZoom;
            const worldY = camera.scrollY + (pointer.y - camera.centerY) / oldZoom;
            
            // Apply the zoom
            camera.setZoom(newZoom);
            
            // Calculate new camera position to keep world point under cursor
            const newScrollX = worldX - (pointer.x - camera.centerX) / newZoom;
            const newScrollY = worldY - (pointer.y - camera.centerY) / newZoom;
            
            camera.scrollX = newScrollX;
            camera.scrollY = newScrollY;
        });
    }
    
    private setupInputHandling() {
        // Mouse/touch interaction handling
        this.input.on('pointerdown', (pointer: Phaser.Input.Pointer) => {
            if (pointer.button === 0) { // Left click only
                this.isMouseDown = true;
                this.hasDragged = false;
                this.lastPointerPosition = { x: pointer.x, y: pointer.y };
                this.paintModeStartPosition = { x: pointer.x, y: pointer.y };
                
                // Check for paint mode (Alt or Cmd key)
                this.isPaintMode = pointer.event.altKey || pointer.event.metaKey || pointer.event.ctrlKey;
                
                // If in paint mode, immediately paint at start position
                if (this.isPaintMode) {
                    const worldPoint = this.cameras.main.getWorldPoint(pointer.x, pointer.y);
                    const hexCoords = pixelToHex(worldPoint.x, worldPoint.y);
                    this.onTileClick(hexCoords.q, hexCoords.r);
                }
            }
        });
        
        this.input.on('pointerup', (pointer: Phaser.Input.Pointer) => {
            if (pointer.button === 0) { // Left click only
                // Only paint on mouse up if we didn't drag and weren't in paint mode
                if (!this.hasDragged && !this.isPaintMode) {
                    const worldPoint = this.cameras.main.getWorldPoint(pointer.x, pointer.y);
                    const hexCoords = pixelToHex(worldPoint.x, worldPoint.y);
                    this.onTileClick(hexCoords.q, hexCoords.r);
                }
                
                // Reset state
                this.isMouseDown = false;
                this.lastPointerPosition = null;
                this.isPaintMode = false;
                this.hasDragged = false;
                this.paintModeStartPosition = null;
            }
        });
        
        this.input.on('pointermove', (pointer: Phaser.Input.Pointer) => {
            if (this.isMouseDown && this.lastPointerPosition) {
                const deltaX = pointer.x - this.lastPointerPosition.x;
                const deltaY = pointer.y - this.lastPointerPosition.y;
                
                // Check if we've moved enough to consider it a drag
                const dragThreshold = 5; // pixels
                if (Math.abs(deltaX) > dragThreshold || Math.abs(deltaY) > dragThreshold) {
                    this.hasDragged = true;
                }
                
                if (this.isPaintMode) {
                    // Paint mode: paint at current position
                    const worldPoint = this.cameras.main.getWorldPoint(pointer.x, pointer.y);
                    const hexCoords = pixelToHex(worldPoint.x, worldPoint.y);
                    this.onTileClick(hexCoords.q, hexCoords.r);
                } else {
                    // Pan mode: move camera
                    this.cameras.main.scrollX -= deltaX / this.cameras.main.zoom;
                    this.cameras.main.scrollY -= deltaY / this.cameras.main.zoom;
                }
                
                this.lastPointerPosition = { x: pointer.x, y: pointer.y };
            }
        });
    }
    
    update() {
        // Handle camera movement with keyboard only if not in input context
        if (!this.isInInputContext()) {
            const camera = this.cameras.main;
            
            if (this.cursors?.left.isDown || this.wasdKeys?.A.isDown) {
                camera.scrollX -= this.panSpeed * (1 / camera.zoom);
            }
            if (this.cursors?.right.isDown || this.wasdKeys?.D.isDown) {
                camera.scrollX += this.panSpeed * (1 / camera.zoom);
            }
            if (this.cursors?.up.isDown || this.wasdKeys?.W.isDown) {
                camera.scrollY -= this.panSpeed * (1 / camera.zoom);
            }
            if (this.cursors?.down.isDown || this.wasdKeys?.S.isDown) {
                camera.scrollY += this.panSpeed * (1 / camera.zoom);
            }
        }
        
        // Update grid and coordinates when camera moves or zooms
        this.updateGridDisplay();
        this.updateCoordinatesDisplay();
    }
    
    /**
     * Check if user is currently focused on an input field
     */
    private isInInputContext(): boolean {
        const activeElement = document.activeElement as HTMLElement;
        if (!activeElement) return false;
        
        const tagName = activeElement.tagName.toLowerCase();
        return (
            tagName === 'input' ||
            tagName === 'textarea' ||
            tagName === 'select' ||
            activeElement.contentEditable === 'true' ||
            activeElement.closest('.modal') !== null ||
            activeElement.closest('[contenteditable="true"]') !== null
        );
    }

    // Helper functions for coordinate conversion (matching lib/hex_coords.go)
    private hexToRowCol(q: number, r: number): { row: number; col: number } {
        // HexToRowCol: cube_to_oddr conversion
        const x = q;
        const z = r;
        const col = x + Math.floor((z - (z & 1)) / 2);
        const row = z;
        return { row, col };
    }
    
    private rowColToHex(row: number, col: number): { q: number; r: number } {
        // RowColToHex: oddr_to_cube conversion
        const x = col - Math.floor((row - (row & 1)) / 2);
        const z = row;
        const q = x;
        const r = z;
        return { q, r };
    }
    
    
    // Public methods for world manipulation
    public setTile(q: number, r: number, terrainType: number, color: number = 0) {
        const key = `${q},${r}`;
        const position = hexToPixel(q, r);
        
        // Remove existing tile if it exists
        if (this.tiles.has(key)) {
            this.tiles.get(key)?.destroy();
        }
        
        // Create new tile sprite
        const textureKey = `terrain_${terrainType}_${color}`;
        
        if (this.textures.exists(textureKey)) {
            const tileSprite = this.add.sprite(position.x, position.y, textureKey);
            tileSprite.setOrigin(0.5, 0.5);
            this.tiles.set(key, tileSprite);
        } else {
            console.warn(`[PhaserWorldScene] Texture not found: ${textureKey}`);
            console.warn(`[PhaserWorldScene] Available textures:`, this.textures.list);
            
            // Try fallback to basic terrain texture without color
            const fallbackKey = `terrain_${terrainType}`;
            if (this.textures.exists(fallbackKey)) {
                console.log(`[PhaserWorldScene] Using fallback texture: ${fallbackKey}`);
                const tileSprite = this.add.sprite(position.x, position.y, fallbackKey);
                tileSprite.setOrigin(0.5, 0.5);
                this.tiles.set(key, tileSprite);
            } else {
                console.error(`[PhaserWorldScene] Fallback texture also not found: ${fallbackKey}`);
            }
        }
        
        // Update coordinate text if enabled
        if (this.showCoordinates) {
            this.updateCoordinateText(q, r);
        }
    }
    
    public removeTile(q: number, r: number) {
        const key = `${q},${r}`;
        
        if (this.tiles.has(key)) {
            this.tiles.get(key)?.destroy();
            this.tiles.delete(key);
        }
        
        // Remove coordinate text
        if (this.coordinateTexts.has(key)) {
            this.coordinateTexts.get(key)?.destroy();
            this.coordinateTexts.delete(key);
        }
    }
    
    public clearAllTiles() {
        this.tiles.forEach(tile => tile.destroy());
        this.tiles.clear();
        
        this.coordinateTexts.forEach(text => text.destroy());
        this.coordinateTexts.clear();
    }
    
    // Unit management methods
    public setUnit(q: number, r: number, unitType: number, color: number = 1) {
        const key = `${q},${r}`;
        const position = hexToPixel(q, r);
        
        // Remove existing unit if it exists
        if (this.units.has(key)) {
            this.units.get(key)?.destroy();
        }
        
        // Create new unit sprite
        const textureKey = `unit_${unitType}_${color}`;
        
        if (this.textures.exists(textureKey)) {
            const unitSprite = this.add.sprite(position.x, position.y, textureKey);
            unitSprite.setOrigin(0.5, 0.5);
            unitSprite.setDepth(10); // Units render above tiles
            this.units.set(key, unitSprite);
        } else {
            console.warn(`[PhaserWorldScene] Unit texture not found: ${textureKey}`);
            
            // Try fallback to basic unit texture without color
            const fallbackKey = `unit_${unitType}`;
            if (this.textures.exists(fallbackKey)) {
                console.log(`[PhaserWorldScene] Using fallback unit texture: ${fallbackKey}`);
                const unitSprite = this.add.sprite(position.x, position.y, fallbackKey);
                unitSprite.setOrigin(0.5, 0.5);
                unitSprite.setDepth(10);
                this.units.set(key, unitSprite);
            } else {
                console.error(`[PhaserWorldScene] Fallback unit texture also not found: ${fallbackKey}`);
            }
        }
    }
    
    public removeUnit(q: number, r: number) {
        const key = `${q},${r}`;
        
        if (this.units.has(key)) {
            this.units.get(key)?.destroy();
            this.units.delete(key);
        }
    }
    
    public clearAllUnits() {
        this.units.forEach(unit => unit.destroy());
        this.units.clear();
    }
    
    public setShowGrid(show: boolean) {
        this.showGrid = show;
        this.updateGridDisplay();
    }
    
    public setShowCoordinates(show: boolean) {
        this.showCoordinates = show;
        this.updateCoordinatesDisplay();
    }
    
    public setTheme(isDark: boolean) {
        this.isDarkTheme = isDark;
        this.updateTheme();
    }
    
    private updateTheme() {
        // Update camera background color based on theme
        const backgroundColor = this.isDarkTheme ? 0x1f2937 : 0xf3f4f6; // gray-800 : gray-100
        this.cameras.main.setBackgroundColor(backgroundColor);
        
        // Update grid and coordinates
        this.updateGridDisplay();
        this.updateCoordinatesDisplay();
    }
    
    private updateCoordinatesDisplay() {
        // Clear existing coordinate texts
        this.coordinateTexts.forEach(text => text.destroy());
        this.coordinateTexts.clear();
        
        if (!this.showCoordinates) return;
        
        // Get camera bounds (same logic as grid)
        const camera = this.cameras.main;
        const worldView = camera.worldView;
        
        const padding = Math.max(this.tileWidth, this.tileHeight) * 2;
        const minX = worldView.x - padding;
        const maxX = worldView.x + worldView.width + padding;
        const minY = worldView.y - padding;
        const maxY = worldView.y + worldView.height + padding;
        
        // Convert bounds to hex coordinates
        const topLeft = pixelToHex(minX, minY);
        const topRight = pixelToHex(maxX, minY);
        const bottomLeft = pixelToHex(minX, maxY);
        const bottomRight = pixelToHex(maxX, maxY);
        
        // Find the bounding box in hex coordinates
        const minQ = Math.min(topLeft.q, topRight.q, bottomLeft.q, bottomRight.q) - 1;
        const maxQ = Math.max(topLeft.q, topRight.q, bottomLeft.q, bottomRight.q) + 1;
        const minR = Math.min(topLeft.r, topRight.r, bottomLeft.r, bottomRight.r) - 1;
        const maxR = Math.max(topLeft.r, topRight.r, bottomLeft.r, bottomRight.r) + 1;
        
        // Show coordinates for all visible hexes
        for (let q = minQ; q <= maxQ; q++) {
            for (let r = minR; r <= maxR; r++) {
                this.updateCoordinateText(q, r);
            }
        }
    }
    
    private updateGridDisplay() {
        if (!this.gridGraphics) return;
        
        this.gridGraphics.clear();
        
        if (!this.showGrid) return;
        
        // Get camera bounds
        const camera = this.cameras.main;
        const worldView = camera.worldView;
        
        // Add some padding to ensure grid covers entire visible area
        const padding = Math.max(this.tileWidth, this.tileHeight) * 2;
        const minX = worldView.x - padding;
        const maxX = worldView.x + worldView.width + padding;
        const minY = worldView.y - padding;
        const maxY = worldView.y + worldView.height + padding;
        
        // Convert bounds to hex coordinates to find range
        const topLeft = pixelToHex(minX, minY);
        const topRight = pixelToHex(maxX, minY);
        const bottomLeft = pixelToHex(minX, maxY);
        const bottomRight = pixelToHex(maxX, maxY);
        
        // Find the bounding box in hex coordinates
        const minQ = Math.min(topLeft.q, topRight.q, bottomLeft.q, bottomRight.q) - 2;
        const maxQ = Math.max(topLeft.q, topRight.q, bottomLeft.q, bottomRight.q) + 2;
        const minR = Math.min(topLeft.r, topRight.r, bottomLeft.r, bottomRight.r) - 2;
        const maxR = Math.max(topLeft.r, topRight.r, bottomLeft.r, bottomRight.r) + 2;
        
        // Draw grid lines with theme-appropriate colors
        const gridColor = this.isDarkTheme ? 0xcccccc : 0x444444;
        const gridAlpha = this.isDarkTheme ? 0.4 : 0.3;
        this.gridGraphics.lineStyle(1, gridColor, gridAlpha);
        
        // Draw grid for all visible hexes
        for (let q = minQ; q <= maxQ; q++) {
            for (let r = minR; r <= maxR; r++) {
                this.drawHexagon(q, r);
            }
        }
    }
    
    private drawHexagon(q: number, r: number) {
        if (!this.gridGraphics) return;
        
        const position = hexToPixel(q, r);
        const halfWidth = this.tileWidth / 2;
        const halfHeight = this.tileHeight / 2;
        
        // Draw pointy-topped hexagon outline based on tile dimensions
        this.gridGraphics.beginPath();
        
        // Pointy-topped hexagon vertices (starting from top point, going clockwise)
        const vertices = [
            { x: position.x, y: position.y - halfHeight },                    // Top
            { x: position.x + halfWidth * 0.866, y: position.y - halfHeight * 0.5 }, // Top-right
            { x: position.x + halfWidth * 0.866, y: position.y + halfHeight * 0.5 }, // Bottom-right
            { x: position.x, y: position.y + halfHeight },                    // Bottom
            { x: position.x - halfWidth * 0.866, y: position.y + halfHeight * 0.5 }, // Bottom-left
            { x: position.x - halfWidth * 0.866, y: position.y - halfHeight * 0.5 }  // Top-left
        ];
        
        this.gridGraphics.moveTo(vertices[0].x, vertices[0].y);
        for (let i = 1; i < vertices.length; i++) {
            this.gridGraphics.lineTo(vertices[i].x, vertices[i].y);
        }
        this.gridGraphics.closePath();
        this.gridGraphics.strokePath();
    }
    
    private updateCoordinateText(q: number, r: number) {
        if (!this.showCoordinates) return;
        
        const key = `${q},${r}`;
        const position = hexToPixel(q, r);
        
        // Remove existing text
        if (this.coordinateTexts.has(key)) {
            this.coordinateTexts.get(key)?.destroy();
        }
        
        // Create coordinate text with both Q/R and row/col using proper conversion
        const { row, col } = this.hexToRowCol(q, r);
        
        const coordText = `QR:${q}, ${r}\nRC:${row}, ${col}`;
        
        const textColor = this.isDarkTheme ? '#ffffff' : '#000000';
        const strokeColor = this.isDarkTheme ? '#000000' : '#ffffff';
        
        const text = this.add.text(position.x, position.y, coordText, {
            fontSize: '10px',
            color: textColor,
            stroke: strokeColor,
            strokeThickness: 1,
            align: 'center'
        });
        
        text.setOrigin(0.5, 0.5);
        this.coordinateTexts.set(key, text);
    }
    
    // Test method to create a simple pattern
    public createTestPattern() {
        // Clear existing tiles
        this.clearAllTiles();
        
        // Create a simple pattern with different terrain types
        const patterns = [
            { q: 0, r: 0, terrain: 1, color: 0 },   // Grass
            { q: 1, r: 0, terrain: 2, color: 0 },   // Desert
            { q: -1, r: 0, terrain: 3, color: 0 },  // Water
            { q: 0, r: 1, terrain: 16, color: 0 },  // Mountain
            { q: 0, r: -1, terrain: 20, color: 0 }, // Rock
            { q: 1, r: -1, terrain: 1, color: 1 },  // Grass (different color)
            { q: -1, r: 1, terrain: 2, color: 2 },  // Desert (different color)
        ];
        
        patterns.forEach(pattern => {
            this.setTile(pattern.q, pattern.r, pattern.terrain, pattern.color);
        });
        
        // Update grid display
        this.updateGridDisplay();
        
        console.log('[PhaserWorldScene] Test pattern created with negative coordinates support');
    }
    
    // Callback for tile clicks (to be overridden by parent)
    private onTileClick(q: number, r: number) {
        console.log(`[PhaserWorldScene] Tile clicked: Q=${q}, R=${r}`);
        
        // DISABLED: Old terrain painting logic - now handled by WorldEditorPage
        // The WorldEditorPage will handle all placement logic based on the current mode
        
        // Just emit the tile click event for the WorldEditorPage to handle
        this.events.emit('tileClicked', { q, r });
    }
    
    private getCurrentTerrainSelection(): { terrain: number; color: number } {
        // Find selected terrain button
        const selectedButton = document.querySelector('.terrain-button.bg-blue-100, .terrain-button.bg-blue-900') as HTMLElement;
        if (selectedButton) {
            const terrain = parseInt(selectedButton.getAttribute('data-terrain') || '5');
            const hasColors = selectedButton.getAttribute('data-has-colors') === 'true';
            
            // If it's a city terrain, get the selected player color
            if (hasColors) {
                const playerColorSelect = document.getElementById('player-color') as HTMLSelectElement;
                const color = playerColorSelect ? parseInt(playerColorSelect.value) : 0;
                return { terrain, color };
            } else {
                // Nature terrain always uses color 0
                return { terrain, color: 0 };
            }
        }
        return { terrain: 5, color: 0 }; // Default to grass
    }
    
    private getCurrentBrushSize(): number {
        const brushSelect = document.getElementById('brush-size') as HTMLSelectElement;
        return brushSelect ? parseInt(brushSelect.value) : 0;
    }
    
    private paintTileArea(centerQ: number, centerR: number, terrain: number, color: number, brushSize: number) {
        if (brushSize === 0) {
            // Single tile
            this.setTile(centerQ, centerR, terrain, color);
        } else {
            // Multiple tiles in radius
            const radius = this.getBrushRadius(brushSize);
            for (let q = centerQ - radius; q <= centerQ + radius; q++) {
                for (let r = centerR - radius; r <= centerR + radius; r++) {
                    // Use cube distance to determine if tile is within brush radius
                    const distance = Math.abs(q - centerQ) + Math.abs(r - centerR) + Math.abs(-q - r - (-centerQ - centerR));
                    if (distance <= radius * 2) { // Hex distance formula
                        this.setTile(q, r, terrain, color);
                    }
                }
            }
        }
    }
    
    private clearTileArea(centerQ: number, centerR: number, brushSize: number) {
        if (brushSize === 0) {
            // Single tile
            this.removeTile(centerQ, centerR);
        } else {
            // Multiple tiles in radius
            const radius = this.getBrushRadius(brushSize);
            for (let q = centerQ - radius; q <= centerQ + radius; q++) {
                for (let r = centerR - radius; r <= centerR + radius; r++) {
                    // Use cube distance to determine if tile is within brush radius
                    const distance = Math.abs(q - centerQ) + Math.abs(r - centerR) + Math.abs(-q - r - (-centerQ - centerR));
                    if (distance <= radius * 2) { // Hex distance formula
                        this.removeTile(q, r);
                    }
                }
            }
        }
    }
    
    private getBrushRadius(brushSize: number): number {
        switch (brushSize) {
            case 1: return 1;   // Small (3 hexes)
            case 3: return 2;   // Medium (5 hexes) 
            case 5: return 3;   // Large (9 hexes)
            case 10: return 4;  // X-Large (15 hexes)
            case 15: return 5;  // XX-Large (21 hexes)
            default: return 0;  // Single hex
        }
    }
    
    // Scene ready callback
    public onSceneReady(callback: () => void): void {
        if (this.sceneReadyCallback) {
            // Scene is already ready, call immediately
            callback();
        } else {
            this.sceneReadyCallback = callback;
        }
    }
    
    // Wait for assets to be loaded
    public async waitForAssetsReady(): Promise<void> {
        if (this.terrainsLoaded && this.unitsLoaded) {
            return Promise.resolve();
        }
        
        if (!this.assetsReadyPromise) {
            throw new Error('[PhaserWorldScene] Assets ready promise not initialized');
        }
        
        console.log('[PhaserWorldScene] Waiting for assets to be ready...');
        return this.assetsReadyPromise;
    }
    
    // Get all tiles data (for integration with WASM)
    public getTilesData(): Array<{ q: number; r: number; terrain: number; color: number }> {
        const tilesData: Array<{ q: number; r: number; terrain: number; color: number }> = [];
        
        this.tiles.forEach((tile, key) => {
            const [q, r] = key.split(',').map(Number);
            // Extract terrain and color from texture key
            const textureKey = tile.texture.key;
            const match = textureKey.match(/terrain_(\d+)_(\d+)/);
            
            if (match) {
                tilesData.push({
                    q,
                    r,
                    terrain: parseInt(match[1]),
                    color: parseInt(match[2])
                });
            }
        });
        
        return tilesData;
    }
    
    // Get all units data (for integration with WASM)
    public getUnitsData(): Array<{ q: number; r: number; unitType: number; playerId: number }> {
        const unitsData: Array<{ q: number; r: number; unitType: number; playerId: number }> = [];
        
        this.units.forEach((unit, key) => {
            const [q, r] = key.split(',').map(Number);
            // Extract unitType and playerId from texture key
            const textureKey = unit.texture.key;
            const match = textureKey.match(/unit_(\d+)_(\d+)/);
            
            if (match) {
                unitsData.push({
                    q,
                    r,
                    unitType: parseInt(match[1]),
                    playerId: parseInt(match[2])
                });
            }
        });
        
        return unitsData;
    }
}
