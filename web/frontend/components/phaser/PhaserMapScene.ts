import * as Phaser from 'phaser';

export class PhaserMapScene extends Phaser.Scene {
    private tileWidth: number = 64;
    private tileHeight: number = 64;
    private yIncrement: number = 48; // 3/4 * tileHeight for pointy-topped hexes
    
    private tiles: Map<string, Phaser.GameObjects.Sprite> = new Map();
    private gridGraphics: Phaser.GameObjects.Graphics | null = null;
    private coordinateTexts: Map<string, Phaser.GameObjects.Text> = new Map();
    
    private showGrid: boolean = false;
    private showCoordinates: boolean = false;
    
    // Camera controls
    private cursors: Phaser.Types.Input.Keyboard.CursorKeys | null = null;
    private wasdKeys: any = null;
    private zoomSpeed: number = 0.1;
    private panSpeed: number = 200;
    
    // Mouse interaction
    private isMouseDown: boolean = false;
    private lastPointerPosition: { x: number; y: number } | null = null;
    private isPaintMode: boolean = false;
    private hasDragged: boolean = false;
    private paintModeStartPosition: { x: number; y: number } | null = null;
    
    // Asset loading
    private terrainsLoaded: boolean = false;
    private unitsLoaded: boolean = false;
    
    constructor() {
        super({ key: 'PhaserMapScene' });
    }
    
    preload() {
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
        
        console.log('[PhaserMapScene] Scene created successfully');
    }
    
    private loadTerrainAssets() {
        // Load terrain assets from the copied assets directory
        const terrainTypes = [1, 2, 3, 16, 20]; // Grass, Desert, Water, Mountain, Rock
        
        terrainTypes.forEach(type => {
            const assetPath = `/static/assets/v1/Tiles/${type}/0.png`;
            this.load.image(`terrain_${type}`, assetPath);
        });
        
        // Load additional terrain variations for colors
        [1, 2, 3, 16, 20].forEach(type => {
            for (let color = 0; color <= 12; color++) {
                const assetPath = `/static/assets/v1/Tiles/${type}/${color}.png`;
                this.load.image(`terrain_${type}_${color}`, assetPath);
            }
        });
    }
    
    private loadUnitAssets() {
        // Load some basic unit assets for testing
        const unitTypes = [1, 2, 3, 4, 5]; // First 5 unit types
        
        unitTypes.forEach(type => {
            for (let color = 0; color <= 12; color++) {
                const assetPath = `/static/assets/v1/Units/${type}/${color}.png`;
                this.load.image(`unit_${type}_${color}`, assetPath);
            }
        });
    }
    
    private setupCameraControls() {
        // Create cursor keys for camera movement
        this.cursors = this.input.keyboard!.createCursorKeys();
        
        // Add WASD keys
        this.wasdKeys = this.input.keyboard!.addKeys('W,S,A,D');
        
        // Mouse wheel zoom
        this.input.on('wheel', (pointer: Phaser.Input.Pointer, gameObjects: Phaser.GameObjects.GameObject[], deltaX: number, deltaY: number) => {
            const camera = this.cameras.main;
            const zoomFactor = deltaY > 0 ? 1 - this.zoomSpeed : 1 + this.zoomSpeed;
            const newZoom = Phaser.Math.Clamp(camera.zoom * zoomFactor, 0.1, 3);
            camera.setZoom(newZoom);
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
                    const hexCoords = this.pixelToHex(worldPoint.x, worldPoint.y);
                    this.onTileClick(hexCoords.q, hexCoords.r);
                }
            }
        });
        
        this.input.on('pointerup', (pointer: Phaser.Input.Pointer) => {
            if (pointer.button === 0) { // Left click only
                // Only paint on mouse up if we didn't drag and weren't in paint mode
                if (!this.hasDragged && !this.isPaintMode) {
                    const worldPoint = this.cameras.main.getWorldPoint(pointer.x, pointer.y);
                    const hexCoords = this.pixelToHex(worldPoint.x, worldPoint.y);
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
                    const hexCoords = this.pixelToHex(worldPoint.x, worldPoint.y);
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
        // Handle camera movement with keyboard
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
        
        // Update grid and coordinates when camera moves or zooms
        this.updateGridDisplay();
        this.updateCoordinatesDisplay();
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
    
    // Convert pixel coordinates to hex coordinates (matching lib/map.go XYToQR)
    private pixelToHex(x: number, y: number): { q: number; r: number } {
        // Match the Go implementation from map.go XYToQR
        const row = Math.floor((y + this.tileHeight / 2) / this.yIncrement);
        
        let halfDists = Math.floor(1 + Math.abs(x * 2 / this.tileWidth));
        if ((row & 1) !== 0) {
            halfDists = Math.floor(1 + Math.abs((x - this.tileWidth / 2) * 2 / this.tileWidth));
        }
        
        let col = Math.floor(halfDists / 2);
        if (x < 0) {
            col = -col;
        }
        
        return this.rowColToHex(row, col);
    }
    
    // Convert hex coordinates to pixel coordinates (matching lib/map.go CenterXYForTile)
    private hexToPixel(q: number, r: number): { x: number; y: number } {
        // Match the Go implementation from map.go CenterXYForTile
        const { row, col } = this.hexToRowCol(q, r);
        
        let y = this.yIncrement * row;
        let x = this.tileWidth * col;
        
        if ((row & 1) === 1) {
            x += this.tileWidth / 2;
        }
        
        return { x, y };
    }
    
    // Public methods for map manipulation
    public setTile(q: number, r: number, terrainType: number, color: number = 0) {
        const key = `${q},${r}`;
        const position = this.hexToPixel(q, r);
        
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
            console.warn(`[PhaserMapScene] Texture not found: ${textureKey}`);
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
    
    public setShowGrid(show: boolean) {
        this.showGrid = show;
        this.updateGridDisplay();
    }
    
    public setShowCoordinates(show: boolean) {
        this.showCoordinates = show;
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
        const topLeft = this.pixelToHex(minX, minY);
        const topRight = this.pixelToHex(maxX, minY);
        const bottomLeft = this.pixelToHex(minX, maxY);
        const bottomRight = this.pixelToHex(maxX, maxY);
        
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
        const topLeft = this.pixelToHex(minX, minY);
        const topRight = this.pixelToHex(maxX, minY);
        const bottomLeft = this.pixelToHex(minX, maxY);
        const bottomRight = this.pixelToHex(maxX, maxY);
        
        // Find the bounding box in hex coordinates
        const minQ = Math.min(topLeft.q, topRight.q, bottomLeft.q, bottomRight.q) - 2;
        const maxQ = Math.max(topLeft.q, topRight.q, bottomLeft.q, bottomRight.q) + 2;
        const minR = Math.min(topLeft.r, topRight.r, bottomLeft.r, bottomRight.r) - 2;
        const maxR = Math.max(topLeft.r, topRight.r, bottomLeft.r, bottomRight.r) + 2;
        
        // Draw grid lines
        this.gridGraphics.lineStyle(1, 0x888888, 0.3);
        
        // Draw grid for all visible hexes
        for (let q = minQ; q <= maxQ; q++) {
            for (let r = minR; r <= maxR; r++) {
                this.drawHexagon(q, r);
            }
        }
    }
    
    private drawHexagon(q: number, r: number) {
        if (!this.gridGraphics) return;
        
        const position = this.hexToPixel(q, r);
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
        const position = this.hexToPixel(q, r);
        
        // Remove existing text
        if (this.coordinateTexts.has(key)) {
            this.coordinateTexts.get(key)?.destroy();
        }
        
        // Create coordinate text with both Q/R and row/col using proper conversion
        const { row, col } = this.hexToRowCol(q, r);
        
        const coordText = `QR:${q}, ${r}\nRC:${row}, ${col}`;
        
        const text = this.add.text(position.x, position.y, coordText, {
            fontSize: '10px',
            color: '#ffffff',
            stroke: '#000000',
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
        
        console.log('[PhaserMapScene] Test pattern created with negative coordinates support');
    }
    
    // Callback for tile clicks (to be overridden by parent)
    private onTileClick(q: number, r: number) {
        console.log(`[PhaserMapScene] Tile clicked: Q=${q}, R=${r}`);
        
        // Emit event that can be caught by the parent component
        this.events.emit('tileClicked', { q, r });
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
}
