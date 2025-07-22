import * as Phaser from 'phaser';
import { PhaserWorldScene } from './PhaserWorldScene';
import { hexToPixel, pixelToHex, HexCoord, PixelCoord } from './hexUtils';

export class EditablePhaserWorldScene extends PhaserWorldScene {
    // Reference image system (editor-only)
    private referenceImage: Phaser.GameObjects.Sprite | null = null;
    private referenceMode: number = 0; // 0=hidden, 1=background, 2=overlay
    private referenceAlpha: number = 0.5; // Default alpha for reference image
    private referencePosition: { x: number; y: number } = { x: 0, y: 0 };
    private referenceScale: { x: number; y: number } = { x: 1, y: 1 };
    private referenceTextureKey: string | null = null;
    
    // Camera override system for different modes
    private originalCameraControls: boolean = true; // Track if camera controls world or reference
    
    constructor(config?: string | Phaser.Types.Scenes.SettingsConfig) {
        super(config || { key: 'EditablePhaserWorldScene' });
    }
    
    create() {
        // Call parent create method
        super.create();
        
        // Initialize reference image system
        this.setupReferenceImageSystem();
        
        // Override input handling for reference image controls
        this.setupReferenceInputHandling();
    }
    
    private setupReferenceImageSystem(): void {
        // Reference image will be added when user switches to reference mode
        console.log('[EditablePhaserWorldScene] Reference image system initialized');
    }
    
    private setupReferenceInputHandling(): void {
        // Override wheel event for reference zoom control in mode 2
        this.input.off('wheel'); // Remove parent wheel handler
        this.input.on('wheel', (pointer: Phaser.Input.Pointer, gameObjects: Phaser.GameObjects.GameObject[], deltaX: number, deltaY: number) => {
            this.handleWheelEvent(pointer, deltaX, deltaY);
        });
        
        // Override pointer events for reference dragging in mode 2
        this.input.off('pointermove'); // Remove parent pointermove handler
        this.input.on('pointermove', (pointer: Phaser.Input.Pointer) => {
            this.handlePointerMove(pointer);
        });
    }
    
    
    /**
     * Load reference image from clipboard
     */
    public async loadReferenceFromClipboard(): Promise<boolean> {
        try {
            // Check if clipboard API is available
            if (!navigator.clipboard) {
                console.warn('[EditablePhaserWorldScene] Clipboard API not available');
                return false;
            }
            
            console.log('[EditablePhaserWorldScene] Attempting to read clipboard...');
            
            // Try clipboard.read() first
            if (navigator.clipboard.read) {
                try {
                    const clipboardItems = await navigator.clipboard.read();
                    console.log(`[EditablePhaserWorldScene] Found ${clipboardItems.length} clipboard items`);
                    
                    for (const clipboardItem of clipboardItems) {
                        console.log(`[EditablePhaserWorldScene] Clipboard item types:`, clipboardItem.types);
                        
                        for (const type of clipboardItem.types) {
                            console.log(`[EditablePhaserWorldScene] Checking type: ${type}`);
                            if (type.startsWith('image/')) {
                                console.log(`[EditablePhaserWorldScene] Found image type: ${type}`);
                                const blob = await clipboardItem.getType(type);
                                console.log(`[EditablePhaserWorldScene] Blob size: ${blob.size} bytes`);
                                return this.replaceReferenceImage(blob);
                            }
                        }
                    }
                } catch (readError) {
                    console.warn('[EditablePhaserWorldScene] clipboard.read() failed:', readError);
                }
            }
            
            // Fallback: Try readText for data URLs
            if (navigator.clipboard.readText) {
                try {
                    console.log('[EditablePhaserWorldScene] Trying readText fallback...');
                    const text = await navigator.clipboard.readText();
                    if (text.startsWith('data:image/')) {
                        console.log('[EditablePhaserWorldScene] Found data URL in text');
                        return this.replaceReferenceImageFromDataURL(text);
                    }
                } catch (textError) {
                    console.warn('[EditablePhaserWorldScene] clipboard.readText() failed:', textError);
                }
            }
            
            console.warn('[EditablePhaserWorldScene] No image found in clipboard');
            return false;
            
        } catch (error) {
            console.error('[EditablePhaserWorldScene] Failed to read clipboard:', error);
            return false;
        }
    }
    
    /**
     * Load reference image from data URL
     */
    private async loadReferenceFromDataURL(dataURL: string): Promise<boolean> {
        try {
            // Generate unique texture key
            this.referenceTextureKey = `reference_${Date.now()}`;
            
            // Create image element to load the data URL
            const img = new Image();
            
            return new Promise((resolve) => {
                img.onload = () => {
                    // Create canvas to get image data
                    const canvas = document.createElement('canvas');
                    const ctx = canvas.getContext('2d');
                    
                    canvas.width = img.width;
                    canvas.height = img.height;
                    ctx!.drawImage(img, 0, 0);
                    
                    // Add texture to Phaser
                    this.textures.addCanvas(this.referenceTextureKey!, canvas);
                    
                    console.log(`[EditablePhaserWorldScene] Reference image loaded from data URL: ${img.width}x${img.height}`);
                    resolve(true);
                };
                
                img.onerror = () => {
                    console.error('[EditablePhaserWorldScene] Failed to load reference image from data URL');
                    resolve(false);
                };
                
                img.src = dataURL;
            });
            
        } catch (error) {
            console.error('[EditablePhaserWorldScene] Failed to load reference from data URL:', error);
            return false;
        }
    }
    
    /**
     * Load reference image from file
     */
    public async loadReferenceFromFile(file: File): Promise<boolean> {
        try {
            console.log(`[EditablePhaserWorldScene] Loading reference image from file: ${file.name} (${file.size} bytes, type: ${file.type})`);
            
            // File is already a Blob, so we can use it directly
            return this.replaceReferenceImage(file);
            
        } catch (error) {
            console.error('[EditablePhaserWorldScene] Failed to load reference from file:', error);
            return false;
        }
    }
    
    /**
     * Replace current reference image with new blob (preserves settings)
     */
    private async replaceReferenceImage(blob: Blob): Promise<boolean> {
        // Store current settings before clearing
        const currentMode = this.referenceMode;
        const currentAlpha = this.referenceAlpha;
        const currentPosition = { ...this.referencePosition };
        const currentScale = { ...this.referenceScale };
        
        // Clear existing image but preserve settings
        this.clearExistingReferenceTexture();
        
        // Load new image
        const result = await this.loadReferenceFromBlob(blob);
        if (result) {
            // Restore previous settings after successful load
            this.referenceAlpha = currentAlpha;
            this.referencePosition = currentPosition;
            this.referenceScale = currentScale;
            this.setReferenceMode(currentMode);
        }
        
        return result;
    }
    
    /**
     * Replace current reference image with data URL (preserves settings)
     */
    private async replaceReferenceImageFromDataURL(dataURL: string): Promise<boolean> {
        // Store current settings before clearing
        const currentMode = this.referenceMode;
        const currentAlpha = this.referenceAlpha;
        const currentPosition = { ...this.referencePosition };
        const currentScale = { ...this.referenceScale };
        
        // Clear existing image but preserve settings
        this.clearExistingReferenceTexture();
        
        // Load new image
        const result = await this.loadReferenceFromDataURL(dataURL);
        if (result) {
            // Restore previous settings after successful load
            this.referenceAlpha = currentAlpha;
            this.referencePosition = currentPosition;
            this.referenceScale = currentScale;
            this.setReferenceMode(currentMode);
        }
        
        return result;
    }
    
    /**
     * Clear existing reference texture without resetting settings
     */
    private clearExistingReferenceTexture(): void {
        this.hideReferenceImage();
        if (this.referenceTextureKey) {
            this.textures.remove(this.referenceTextureKey);
            this.referenceTextureKey = null;
        }
    }
    
    /**
     * Load reference image from blob data
     */
    private async loadReferenceFromBlob(blob: Blob): Promise<boolean> {
        try {
            console.log(`[EditablePhaserWorldScene] Processing blob: size=${blob.size}, type=${blob.type}`);
            
            // Generate unique texture key
            this.referenceTextureKey = `reference_${Date.now()}`;
            console.log(`[EditablePhaserWorldScene] Generated texture key: ${this.referenceTextureKey}`);
            
            // Create image element to load the blob
            const img = new Image();
            const url = URL.createObjectURL(blob);
            console.log(`[EditablePhaserWorldScene] Created object URL: ${url}`);
            
            return new Promise((resolve) => {
                img.onload = () => {
                    console.log(`[EditablePhaserWorldScene] Image loaded successfully: ${img.width}x${img.height}`);
                    
                    try {
                        // Create canvas to get image data
                        const canvas = document.createElement('canvas');
                        const ctx = canvas.getContext('2d');
                        
                        if (!ctx) {
                            console.error('[EditablePhaserWorldScene] Failed to get canvas context');
                            URL.revokeObjectURL(url);
                            resolve(false);
                            return;
                        }
                        
                        canvas.width = img.width;
                        canvas.height = img.height;
                        ctx.drawImage(img, 0, 0);
                        console.log(`[EditablePhaserWorldScene] Drew image to canvas: ${canvas.width}x${canvas.height}`);
                        
                        // Add texture to Phaser
                        this.textures.addCanvas(this.referenceTextureKey!, canvas);
                        console.log(`[EditablePhaserWorldScene] Added texture to Phaser: ${this.referenceTextureKey}`);
                        
                        // Clean up
                        URL.revokeObjectURL(url);
                        
                        console.log(`[EditablePhaserWorldScene] Reference image loaded successfully: ${img.width}x${img.height}`);
                        resolve(true);
                    } catch (canvasError) {
                        console.error('[EditablePhaserWorldScene] Canvas processing error:', canvasError);
                        URL.revokeObjectURL(url);
                        resolve(false);
                    }
                };
                
                img.onerror = (error) => {
                    URL.revokeObjectURL(url);
                    console.error('[EditablePhaserWorldScene] Image load error:', error);
                    resolve(false);
                };
                
                console.log(`[EditablePhaserWorldScene] Setting image src to: ${url}`);
                img.src = url;
            });
            
        } catch (error) {
            console.error('[EditablePhaserWorldScene] Failed to load reference from blob:', error);
            return false;
        }
    }
    
    /**
     * Set reference image mode
     */
    public setReferenceMode(mode: number): void {
        this.referenceMode = mode;
        
        if (mode === 0) {
            // Hidden mode - remove reference image
            this.hideReferenceImage();
        } else {
            // Background or overlay mode - show reference image
            this.showReferenceImage();
        }
        
        // Update camera controls based on mode
        this.updateCameraControls();
    }
    
    /**
     * Show reference image
     */
    private showReferenceImage(): void {
        if (!this.referenceTextureKey || !this.textures.exists(this.referenceTextureKey)) {
            console.warn('[EditablePhaserWorldScene] No reference image loaded');
            return;
        }
        
        // Remove existing reference image
        if (this.referenceImage) {
            this.referenceImage.destroy();
            this.referenceImage = null;
        }
        
        // Create new reference image sprite
        this.referenceImage = this.add.sprite(
            this.referencePosition.x,
            this.referencePosition.y,
            this.referenceTextureKey
        );
        
        this.referenceImage.setOrigin(0.5, 0.5);
        this.referenceImage.setScale(this.referenceScale.x, this.referenceScale.y);
        this.referenceImage.setAlpha(this.referenceAlpha);
        
        // Set depth based on mode
        if (this.referenceMode === 1) {
            // Background mode - below tiles
            this.referenceImage.setDepth(-1);
        } else if (this.referenceMode === 2) {
            // Overlay mode - above tiles but below units
            this.referenceImage.setDepth(15);
        }
        
        console.log(`[EditablePhaserWorldScene] Reference image shown in mode ${this.referenceMode}`);
    }
    
    /**
     * Hide reference image
     */
    private hideReferenceImage(): void {
        if (this.referenceImage) {
            this.referenceImage.destroy();
            this.referenceImage = null;
        }
    }
    
    /**
     * Set reference image alpha
     */
    public setReferenceAlpha(alpha: number): void {
        this.referenceAlpha = Math.max(0, Math.min(1, alpha));
        
        if (this.referenceImage) {
            this.referenceImage.setAlpha(this.referenceAlpha);
        }
    }
    
    /**
     * Set reference image position (for mode 2)
     */
    public setReferencePosition(x: number, y: number): void {
        this.referencePosition.x = x;
        this.referencePosition.y = y;
        
        if (this.referenceImage) {
            this.referenceImage.setPosition(x, y);
        }
    }
    
    /**
     * Set reference image scale (for mode 2)
     */
    public setReferenceScale(x: number, y: number): void {
        this.referenceScale.x = x;
        this.referenceScale.y = y;
        
        if (this.referenceImage) {
            this.referenceImage.setScale(x, y);
        }
        
        // Emit event to notify UI of scale change
        this.events.emit('referenceScaleChanged', { x, y });
    }
    
    /**
     * Set reference image scale with top-left corner as pivot point
     */
    public setReferenceScaleFromTopLeft(newScaleX: number, newScaleY: number): void {
        if (!this.referenceImage) {
            this.setReferenceScale(newScaleX, newScaleY);
            return;
        }
        
        // Get current dimensions and scale
        const currentScaleX = this.referenceScale.x;
        const currentScaleY = this.referenceScale.y;
        const imageWidth = this.referenceImage.width;
        const imageHeight = this.referenceImage.height;
        
        // Calculate current top-left corner position
        // Reference image is centered (origin 0.5, 0.5), so top-left is center minus half the scaled dimensions
        const currentTopLeftX = this.referencePosition.x - (imageWidth * currentScaleX) / 2;
        const currentTopLeftY = this.referencePosition.y - (imageHeight * currentScaleY) / 2;
        
        // Calculate what the new center position should be to keep top-left corner fixed
        const newCenterX = currentTopLeftX + (imageWidth * newScaleX) / 2;
        const newCenterY = currentTopLeftY + (imageHeight * newScaleY) / 2;
        
        // Update position and scale
        this.setReferencePosition(newCenterX, newCenterY);
        this.setReferenceScale(newScaleX, newScaleY);
    }
    
    /**
     * Update camera controls based on reference mode
     */
    private updateCameraControls(): void {
        if (this.referenceMode === 2) {
            // Mode 2: Only reference image responds to camera
            this.originalCameraControls = false;
        } else {
            // Mode 0 or 1: Normal camera controls for world
            this.originalCameraControls = true;
        }
    }
    
    /**
     * Handle wheel event - delegates to handleZoom
     */
    private handleWheelEvent(pointer: Phaser.Input.Pointer, deltaX: number, deltaY: number): void {
        this.handleZoom(deltaY, pointer);
    }
    
    /**
     * Handle pointer move event - checks for dragging and delegates to handlePan or parent
     */
    private handlePointerMove(pointer: Phaser.Input.Pointer): void {
        // Check if we're in drag mode (from parent class state)
        const isMouseDown = (this as any).isMouseDown;
        const lastPointerPosition = (this as any).lastPointerPosition;
        const isPaintMode = (this as any).isPaintMode;
        
        if (isMouseDown && lastPointerPosition) {
            const deltaX = pointer.x - lastPointerPosition.x;
            const deltaY = pointer.y - lastPointerPosition.y;
            
            // Check if we've moved enough to consider it a drag
            const dragThreshold = 5; // pixels
            if (Math.abs(deltaX) > dragThreshold || Math.abs(deltaY) > dragThreshold) {
                (this as any).hasDragged = true;
            }
            
            if (isPaintMode) {
                // Paint mode: paint at current position (handled by parent)
                const worldPoint = this.cameras.main.getWorldPoint(pointer.x, pointer.y);
                const hexCoords = pixelToHex(worldPoint.x, worldPoint.y);
                (this as any).onTileClick(hexCoords.q, hexCoords.r);
            } else {
                // Pan mode: delegate to handlePan
                this.handlePan(deltaX, deltaY);
            }
            
            // Update last pointer position
            (this as any).lastPointerPosition = { x: pointer.x, y: pointer.y };
        }
    }
    
    /**
     * Override camera zoom to handle reference mode
     */
    public handleZoom(delta: number, pointer: Phaser.Input.Pointer): void {
        // Get zoom speed from parent class
        const zoomSpeed = (this as any).zoomSpeed || 0.01;
        
        if (this.referenceMode === 2 && this.referenceImage) {
            // Mode 2: Zoom only the reference image with zoom-to-cursor behavior
            const oldScaleX = this.referenceScale.x;
            const oldScaleY = this.referenceScale.y;
            
            const zoomFactor = delta > 0 ? 1 + zoomSpeed : 1 - zoomSpeed;
            const newScaleX = oldScaleX * zoomFactor;
            const newScaleY = oldScaleY * zoomFactor;
            
            // Clamp scale to reasonable bounds
            const minScale = 0.1;
            const maxScale = 5.0;
            
            if (newScaleX >= minScale && newScaleX <= maxScale) {
                // Get world coordinates under mouse cursor before zoom
                const camera = this.cameras.main;
                const worldMouseX = camera.scrollX + (pointer.x - camera.centerX) / camera.zoom;
                const worldMouseY = camera.scrollY + (pointer.y - camera.centerY) / camera.zoom;
                
                // Calculate offset from reference image center to mouse position
                const offsetX = worldMouseX - this.referencePosition.x;
                const offsetY = worldMouseY - this.referencePosition.y;
                
                // Calculate how much the offset changes due to scaling
                const scaleChangeX = newScaleX / oldScaleX;
                const scaleChangeY = newScaleY / oldScaleY;
                
                // Adjust reference image position to keep mouse point in same relative location
                const newPositionX = this.referencePosition.x + offsetX * (1 - scaleChangeX);
                const newPositionY = this.referencePosition.y + offsetY * (1 - scaleChangeY);
                
                // Apply new scale and position
                this.setReferenceScale(newScaleX, newScaleY);
                this.setReferencePosition(newPositionX, newPositionY);
            }
        } else {
            // Mode 0 or 1: Normal camera zoom
            const camera = this.cameras.main;
            const oldZoom = camera.zoom;
            const zoomFactor = delta > 0 ? 1 - zoomSpeed : 1 + zoomSpeed;
            const newZoom = Phaser.Math.Clamp(oldZoom * zoomFactor, 0.1, 3);
            
            // Calculate world coordinates under mouse cursor before zoom (same as parent)
            const worldX = camera.scrollX + (pointer.x - camera.centerX) / oldZoom;
            const worldY = camera.scrollY + (pointer.y - camera.centerY) / oldZoom;
            
            // Apply the zoom
            camera.setZoom(newZoom);
            
            // Calculate new camera position to keep world point under cursor
            const newScrollX = worldX - (pointer.x - camera.centerX) / newZoom;
            const newScrollY = worldY - (pointer.y - camera.centerY) / newZoom;
            
            camera.scrollX = newScrollX;
            camera.scrollY = newScrollY;
        }
    }
    
    /**
     * Override camera pan to handle reference mode
     */
    public handlePan(deltaX: number, deltaY: number): void {
        if (this.referenceMode === 2 && this.referenceImage) {
            // Mode 2: Pan only the reference image
            this.setReferencePosition(
                this.referencePosition.x + deltaX,
                this.referencePosition.y + deltaY
            );
        } else {
            // Mode 0 or 1: Normal camera pan
            const camera = this.cameras.main;
            camera.scrollX -= deltaX / camera.zoom;
            camera.scrollY -= deltaY / camera.zoom;
        }
    }
    
    /**
     * Get reference image state
     */
    public getReferenceState(): {
        mode: number;
        alpha: number;
        position: { x: number; y: number };
        scale: { x: number; y: number };
        hasImage: boolean;
    } {
        return {
            mode: this.referenceMode,
            alpha: this.referenceAlpha,
            position: { ...this.referencePosition },
            scale: { ...this.referenceScale },
            hasImage: this.referenceTextureKey !== null
        };
    }
    
    /**
     * Clear reference image
     */
    public clearReferenceImage(): void {
        this.hideReferenceImage();
        
        if (this.referenceTextureKey) {
            this.textures.remove(this.referenceTextureKey);
            this.referenceTextureKey = null;
        }
        
        // Reset reference state
        this.referenceMode = 0;
        this.referencePosition = { x: 0, y: 0 };
        this.referenceScale = { x: 1, y: 1 };
        this.referenceAlpha = 0.5;
        
        console.log('[EditablePhaserWorldScene] Reference image cleared');
    }
}
