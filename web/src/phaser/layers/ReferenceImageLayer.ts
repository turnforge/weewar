/**
 * Reference Image Layer for Map Editor
 * 
 * This layer works in world coordinate space and provides image overlay functionality
 * for tracing and reference during map editing. Supports positioning, scaling, and
 * transparency controls.
 */

import * as Phaser from 'phaser';
import { BaseLayer, LayerConfig, ClickContext, LayerHitResult } from '../LayerSystem';

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
    
    constructor(scene: Phaser.Scene) {
        super(scene, {
            name: 'reference-image',
            coordinateSpace: 'world',
            interactive: true, // Can be dragged when in overlay mode
            depth: -1, // Background by default, can be moved to 1000 for overlay
        });
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
    public clearReferenceImage(): void {
        if (this.referenceImage) {
            this.referenceImage.destroy();
            this.referenceImage = null;
        }
        
        this.referenceMode = 0;
        this.referenceTextureKey = null;
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
