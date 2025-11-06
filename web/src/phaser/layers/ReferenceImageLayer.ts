/**
 * Reference Image Layer for Map Editor
 * 
 * This layer works in world coordinate space and provides image overlay functionality
 * for tracing and reference during map editing. Supports positioning, scaling, and
 * transparency controls.
 */

import * as Phaser from 'phaser';
import { BaseLayer, LayerConfig, ClickContext, LayerHitResult } from '../LayerSystem';
import { ReferenceImageDB } from '../../lib/ReferenceImageDB';

export interface ReferenceImageState {
    mode: number; // 0=hidden, 1=background, 2=overlay
    alpha: number;
    position: { x: number; y: number };
    scale: { x: number; y: number };
    hasImage: boolean;
}

/**
 * Reference image layer for editor functionality
 */
export class ReferenceImageLayer extends BaseLayer {
    private referenceImage: Phaser.GameObjects.Image | null = null;
    private referenceMode: number = 0; // 0=hidden, 1=background, 2=overlay
    private referenceAlpha: number = 0.5;
    private referencePosition: { x: number; y: number } = { x: 0, y: 0 };
    private referenceScale: { x: number; y: number } = { x: 1, y: 1 };
    private referenceTextureKey: string | null = null;

    // Drag state
    private isDragging: boolean = false;
    private dragStartX: number = 0;
    private dragStartY: number = 0;
    private dragStartImageX: number = 0;
    private dragStartImageY: number = 0;

    // IndexedDB storage for persistent reference images
    private imageDB: ReferenceImageDB = new ReferenceImageDB();
    private worldId: string | null = null;
    
    constructor(scene: Phaser.Scene, worldId?: string) {
        super(scene, {
            name: 'reference-image',
            coordinateSpace: 'world',
            interactive: true, // Can be dragged when in overlay mode
            depth: -1, // Background by default, can be moved to 1000 for overlay
        });

        // Set worldId if provided
        if (worldId) {
            this.worldId = worldId;
        }
    }

    /**
     * Initialize lifecycle - setup IndexedDB and load from storage
     */
    public async init(): Promise<void> {
        // Initialize IndexedDB
        await this.imageDB.init();

        // Load from storage if worldId is set
        if (this.worldId) {
            await this.loadFromStorage();
        }
    }
    
    public hitTest(context: ClickContext): LayerHitResult | null {
        // Only interactive when image exists and is in overlay mode
        if (!this.referenceImage || this.referenceMode !== 2 || !this.visible) {
            return LayerHitResult.TRANSPARENT;
        }
        
        // Check if click is within image bounds
        if (this.isPointInImageBounds(context.worldX, context.worldY)) {
            return LayerHitResult.BLOCK; // Block map interaction, allow image dragging
        }
        
        return LayerHitResult.TRANSPARENT;
    }
    
    public handleClick(context: ClickContext): boolean {
        if (!this.referenceImage || this.referenceMode !== 2) {
            return false;
        }
        
        // Start drag operation
        this.startDrag(context.worldX, context.worldY);
        return true;
    }
    
    public handleDrag(context: ClickContext, deltaX: number, deltaY: number): boolean {
        if (!this.isDragging || !this.referenceImage) {
            return false;
        }

        // Update image position based on drag
        const newX = this.dragStartImageX + (context.worldX - this.dragStartX);
        const newY = this.dragStartImageY + (context.worldY - this.dragStartY);

        this.setReferencePosition(newX, newY);

        // Emit position change event for UI sync
        this.scene.events.emit('referencePositionChanged', { x: newX, y: newY });

        return true;
    }

    public handleScroll(context: ClickContext, deltaY: number): boolean {
        if (!this.referenceImage || this.referenceMode !== 2) {
            return false;
        }

        // Scale the reference image around the mouse cursor position
        const scaleFactor = deltaY > 0 ? 0.99 : 1.01; // Zoom in/out by 1%
        const newScaleX = this.referenceScale.x * scaleFactor;
        const newScaleY = this.referenceScale.y * scaleFactor;

        // Get mouse position in world coordinates
        const mouseWorldX = context.worldX;
        const mouseWorldY = context.worldY;

        // Calculate the offset from the image position to the mouse
        const offsetX = mouseWorldX - this.referencePosition.x;
        const offsetY = mouseWorldY - this.referencePosition.y;

        // When scaling, adjust position so the point under cursor stays in place
        const newOffsetX = offsetX * scaleFactor;
        const newOffsetY = offsetY * scaleFactor;

        const newPosX = mouseWorldX - newOffsetX;
        const newPosY = mouseWorldY - newOffsetY;

        // Apply new scale and position
        this.setReferenceScale(newScaleX, newScaleY);
        this.setReferencePosition(newPosX, newPosY);

        // Emit events for UI sync
        this.scene.events.emit('referenceScaleChanged', { x: newScaleX, y: newScaleY });
        this.scene.events.emit('referencePositionChanged', { x: newPosX, y: newPosY });

        return true; // We handled the scroll
    }
    
    /**
     * Set reference image from URL
     */
    public setReferenceImage(imageUrl: string): void {
        if (this.referenceImage) {
            this.referenceImage.destroy();
            this.referenceImage = null;
        }
        
        // Generate unique texture key
        this.referenceTextureKey = `reference_${Date.now()}`;
        
        // Load and display reference image
        this.scene.load.image(this.referenceTextureKey, imageUrl);
        this.scene.load.start();
        
        this.scene.load.once('complete', () => {
            this.createReferenceImageSprite();
        });
    }
    
    /**
     * Create the reference image sprite after texture is loaded
     */
    private createReferenceImageSprite(): void {
        if (!this.referenceTextureKey) return;
        
        // Remove existing image
        if (this.referenceImage) {
            this.referenceImage.destroy();
        }
        
        // Create new reference image sprite
        this.referenceImage = this.scene.add.image(
            this.referencePosition.x,
            this.referencePosition.y,
            this.referenceTextureKey
        );
        
        // Apply current settings
        this.referenceImage.setAlpha(this.referenceAlpha);
        this.referenceImage.setScale(this.referenceScale.x, this.referenceScale.y);
        this.referenceImage.setVisible(this.referenceMode > 0);
        
        // Set depth based on mode
        if (this.referenceMode === 1) {
            this.referenceImage.setDepth(-1); // Background
        } else if (this.referenceMode === 2) {
            this.referenceImage.setDepth(1000); // Overlay
        }
        
        // Add to container
        this.container.add(this.referenceImage);
    }
    
    /**
     * Set reference mode (0=hidden, 1=background, 2=overlay)
     */
    public setReferenceMode(mode: number): void {
        this.referenceMode = mode;
        
        if (this.referenceImage) {
            this.referenceImage.setVisible(mode > 0);
            
            if (mode === 1) {
                this.referenceImage.setDepth(-1); // Background
                this.setDepth(-1);
            } else if (mode === 2) {
                this.referenceImage.setDepth(1000); // Overlay
                this.setDepth(1000);
            }
        }
    }
    
    /**
     * Set reference image alpha
     */
    public setReferenceAlpha(alpha: number): void {
        this.referenceAlpha = alpha;
        
        if (this.referenceImage) {
            this.referenceImage.setAlpha(alpha);
        }
    }
    
    /**
     * Set reference image position
     */
    public setReferencePosition(x: number, y: number): void {
        this.referencePosition = { x, y };
        
        if (this.referenceImage) {
            this.referenceImage.setPosition(x, y);
        }
    }
    
    /**
     * Set reference image scale
     */
    public setReferenceScale(x: number, y: number): void {
        this.referenceScale = { x, y };
        
        if (this.referenceImage) {
            this.referenceImage.setScale(x, y);
        }
    }
    
    /**
     * Set reference scale from top-left origin
     */
    public setReferenceScaleFromTopLeft(x: number, y: number): void {
        if (this.referenceImage) {
            // Set origin to top-left for scaling
            this.referenceImage.setOrigin(0, 0);
            this.referenceImage.setScale(x, y);
            this.referenceScale = { x, y };
        }
    }
    
    /**
     * Get current reference image state
     */
    public getReferenceState(): ReferenceImageState {
        return {
            mode: this.referenceMode,
            alpha: this.referenceAlpha,
            position: { ...this.referencePosition },
            scale: { ...this.referenceScale },
            hasImage: this.referenceImage !== null
        };
    }
    
    /**
     * Clear reference image
     */
    public async clearReferenceImage(): Promise<void> {
        if (this.referenceImage) {
            this.referenceImage.destroy();
            this.referenceImage = null;
        }

        this.referenceMode = 0;
        this.referenceTextureKey = null;

        // Also clear from IndexedDB storage
        await this.clearFromStorage();
    }
    
    /**
     * Load reference image from IndexedDB storage
     * Called automatically during init() if worldId is set
     */
    private async loadFromStorage(): Promise<boolean> {
        if (!this.worldId) {
            console.warn('[ReferenceImageLayer] Cannot load from storage: worldId not set');
            return false;
        }

        const record = await this.imageDB.getImageRecord(this.worldId);
        if (record) {
            const imageUrl = URL.createObjectURL(record.imageBlob);
            this.setReferenceImage(imageUrl);

            // Always set to background mode when restoring
            // (Hidden is inconvenient, Overlay risks accidental scale/position changes)
            this.setReferenceMode(1); // 1 = background
            console.log('[ReferenceImageLayer] Restored reference image in background mode');

            return true;
        }

        return false;
    }

    /**
     * Clear reference image from IndexedDB storage
     */
    public async clearFromStorage(): Promise<void> {
        if (!this.worldId) {
            console.warn('[ReferenceImageLayer] Cannot clear from storage: worldId not set');
            return;
        }

        await this.imageDB.deleteImage(this.worldId);
        console.log('[ReferenceImageLayer] Cleared reference image from storage');
    }

    /**
     * Load reference image from clipboard
     */
    public async loadReferenceFromClipboard(): Promise<boolean> {
            const items = await navigator.clipboard.read();

            for (const item of items) {
                if (item.types.includes('image/png') || item.types.includes('image/jpeg')) {
                    const imageBlob = await item.getType(item.types.find(type => type.startsWith('image/')) || '');
                    const imageUrl = URL.createObjectURL(imageBlob);

                    this.setReferenceImage(imageUrl);

                    // Save to IndexedDB if worldId is set
                    if (this.worldId) {
                        await this.imageDB.saveImage(this.worldId, imageBlob);
                    }

                    return true;
                }
            }

            console.warn('[ReferenceImageLayer] No image found in clipboard');
            return false;
    }

    /**
     * Load reference image from file
     */
    public async loadReferenceFromFile(file: File): Promise<boolean> {
            if (!file.type.startsWith('image/')) {
                console.error('[ReferenceImageLayer] File is not an image:', file.type);
                return false;
            }

            const imageUrl = URL.createObjectURL(file);
            this.setReferenceImage(imageUrl);

            // Save to IndexedDB if worldId is set
            if (this.worldId) {
                await this.imageDB.saveImage(this.worldId, file, file.name);
            }

            return true;
    }
    
    /**
     * Check if a point is within the image bounds
     */
    private isPointInImageBounds(worldX: number, worldY: number): boolean {
        if (!this.referenceImage) return false;
        
        const bounds = this.referenceImage.getBounds();
        return bounds.contains(worldX, worldY);
    }
    
    /**
     * Start drag operation
     */
    private startDrag(worldX: number, worldY: number): void {
        if (!this.referenceImage) return;
        
        this.isDragging = true;
        this.dragStartX = worldX;
        this.dragStartY = worldY;
        this.dragStartImageX = this.referenceImage.x;
        this.dragStartImageY = this.referenceImage.y;
    }
    
    /**
     * Stop drag operation
     */
    public stopDrag(): void {
        if (this.isDragging) {
            this.isDragging = false;
        }
    }
    
    public destroy(): void {
        this.clearReferenceImage();
        super.destroy();
    }
}
