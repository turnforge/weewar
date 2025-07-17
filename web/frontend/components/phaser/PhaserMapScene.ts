import * as Phaser from 'phaser';

export class PhaserMapScene extends Phaser.Scene {
    private tileSize: number = 32;
    private hexWidth: number = 56; // sqrt(3) * tileSize
    private hexHeight: number = 48; // 3/4 * tileSize * 2
    private hexRadius: number = 32;
    
    private tiles: Map<string, Phaser.GameObjects.Sprite> = new Map();
    private gridGraphics: Phaser.GameObjects.Graphics | null = null;
    private coordinateTexts: Map<string, Phaser.GameObjects.Text> = new Map();
    
    private showGrid: boolean = true;
    private showCoordinates: boolean = false;
    
    // Camera controls
    private cursors: Phaser.Types.Input.Keyboard.CursorKeys | null = null;
    private wasdKeys: any = null;
    private zoomSpeed: number = 0.1;
    private panSpeed: number = 200;
    
    // Mouse interaction
    private isMouseDown: boolean = false;
    private lastPointerPosition: { x: number; y: number } | null = null;
    
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
        // Mouse/touch drag for panning
        this.input.on('pointerdown', (pointer: Phaser.Input.Pointer) => {
            this.isMouseDown = true;
            this.lastPointerPosition = { x: pointer.x, y: pointer.y };
        });
        
        this.input.on('pointerup', (pointer: Phaser.Input.Pointer) => {
            this.isMouseDown = false;
            this.lastPointerPosition = null;
        });
        
        this.input.on('pointermove', (pointer: Phaser.Input.Pointer) => {
            if (this.isMouseDown && this.lastPointerPosition) {
                const deltaX = pointer.x - this.lastPointerPosition.x;
                const deltaY = pointer.y - this.lastPointerPosition.y;
                
                this.cameras.main.scrollX -= deltaX / this.cameras.main.zoom;
                this.cameras.main.scrollY -= deltaY / this.cameras.main.zoom;
                
                this.lastPointerPosition = { x: pointer.x, y: pointer.y };
            }
        });
        
        // Handle tile clicks
        this.input.on('pointerdown', (pointer: Phaser.Input.Pointer) => {
            if (pointer.button === 0) { // Left click
                const worldPoint = this.cameras.main.getWorldPoint(pointer.x, pointer.y);
                const hexCoords = this.pixelToHex(worldPoint.x, worldPoint.y);
                this.onTileClick(hexCoords.q, hexCoords.r);
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
    }
    
    // Convert pixel coordinates to hex coordinates (Q, R)
    private pixelToHex(x: number, y: number): { q: number; r: number } {
        const SQRT3 = Math.sqrt(3);
        
        // Convert to fractional hex coordinates
        const q = (SQRT3 * x - y) / (3 * this.hexRadius);
        const r = (2 * y) / (3 * this.hexRadius);
        
        return this.roundHex(q, r);
    }
    
    // Convert hex coordinates to pixel coordinates
    private hexToPixel(q: number, r: number): { x: number; y: number } {
        const SQRT3 = Math.sqrt(3);
        
        const x = this.hexRadius * (SQRT3 * q + SQRT3 * r / 2);
        const y = this.hexRadius * (3 * r / 2);
        
        return { x, y };
    }
    
    // Round fractional hex coordinates to integer coordinates
    private roundHex(q: number, r: number): { q: number; r: number } {
        const s = -q - r;
        
        let roundedQ = Math.round(q);
        let roundedR = Math.round(r);
        let roundedS = Math.round(s);
        
        const deltaQ = Math.abs(roundedQ - q);
        const deltaR = Math.abs(roundedR - r);
        const deltaS = Math.abs(roundedS - s);
        
        if (deltaQ > deltaR && deltaQ > deltaS) {
            roundedQ = -roundedR - roundedS;
        } else if (deltaR > deltaS) {
            roundedR = -roundedQ - roundedS;
        }
        
        return { q: roundedQ, r: roundedR };
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
        
        if (show) {
            // Show coordinates for all existing tiles
            this.tiles.forEach((tile, key) => {
                const [q, r] = key.split(',').map(Number);
                this.updateCoordinateText(q, r);
            });
        } else {
            // Hide all coordinate texts
            this.coordinateTexts.forEach(text => text.destroy());
            this.coordinateTexts.clear();
        }
    }
    
    private updateGridDisplay() {
        if (!this.gridGraphics) return;
        
        this.gridGraphics.clear();
        
        if (!this.showGrid) return;
        
        // Draw grid lines for visible tiles
        this.gridGraphics.lineStyle(1, 0x888888, 0.5);
        
        this.tiles.forEach((tile, key) => {
            const [q, r] = key.split(',').map(Number);
            this.drawHexagon(q, r);
        });
    }
    
    private drawHexagon(q: number, r: number) {
        if (!this.gridGraphics) return;
        
        const position = this.hexToPixel(q, r);
        const radius = this.hexRadius;
        
        // Draw hexagon outline
        this.gridGraphics.beginPath();
        
        for (let i = 0; i < 6; i++) {
            const angle = (Math.PI / 3) * i;
            const x = position.x + radius * Math.cos(angle);
            const y = position.y + radius * Math.sin(angle);
            
            if (i === 0) {
                this.gridGraphics.moveTo(x, y);
            } else {
                this.gridGraphics.lineTo(x, y);
            }
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
        
        // Create new coordinate text
        const text = this.add.text(position.x, position.y, `${q},${r}`, {
            fontSize: '12px',
            color: '#ffffff',
            stroke: '#000000',
            strokeThickness: 2
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