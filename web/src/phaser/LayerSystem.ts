/**
 * Layered Overlay System for Phaser Scenes
 * 
 * This system provides a clean, reusable way to manage overlays and interactive elements
 * across different coordinate spaces (screen, world, hex) with proper event handling.
 */

import * as Phaser from 'phaser';
import { Tile, Unit } from '../World';

// =============================================================================
// Core Types and Interfaces
// =============================================================================

/**
 * Coordinate space that a layer operates in
 */
export type CoordinateSpace = 'screen' | 'world' | 'hex';

/**
 * Result of layer hit testing - determines event flow
 */
export enum LayerHitResult {
    TRANSPARENT = 'transparent',  // Pass through to layers below
    BLOCK = 'block',             // Block event from layers below but don't handle
    CONSUME = 'consume'          // Handle event and stop propagation
}

/**
 * Complete context for click events with all coordinate spaces pre-computed
 */
export interface ClickContext {
    // Raw input coordinates
    screenX: number;
    screenY: number;
    
    // World coordinates (accounting for camera transform)
    worldX: number;
    worldY: number;
    
    // Hex coordinates (for game logic)
    hexQ: number;
    hexR: number;
    
    // Game context (computed once for efficiency)
    tile?: Tile | null;
    unit?: Unit | null;
    
    // Layer that wants to handle this click
    layer?: string;
    
    // Metadata
    timestamp: number;
    button: number; // 0=left, 1=middle, 2=right
}

/**
 * Base layer interface - all layers must implement this
 */
export interface Layer {
    readonly name: string;
    readonly coordinateSpace: CoordinateSpace;
    readonly interactive: boolean;
    
    // Visibility and positioning
    visible: boolean;
    depth: number;
    
    // Core lifecycle methods
    hitTest(context: ClickContext): LayerHitResult | null;
    handleDrag?(context: ClickContext, deltaX: number, deltaY: number): boolean;
    
    // Display control
    show(): void;
    hide(): void;
    setPosition(x: number, y: number): void;
    setScale(x: number, y: number): void;
    setAlpha(alpha: number): void;
    setDepth(depth: number): void;
    
    // Lifecycle
    destroy(): void;
}

/**
 * Configuration for creating layers
 */
export interface LayerConfig {
    name: string;
    coordinateSpace: CoordinateSpace;
    interactive?: boolean;
    visible?: boolean;
    depth?: number;
    alpha?: number;
}

// =============================================================================
// Layer Manager
// =============================================================================

/**
 * Manages multiple layers with proper z-ordering and event dispatch
 */
export class LayerManager {
    private layers = new Map<string, Layer>();
    private scene: Phaser.Scene;
    private pixelToHexFn: (x: number, y: number) => { q: number; r: number };
    private getTileFn: (q: number, r: number) => Tile | null;
    private getUnitFn: (q: number, r: number) => Unit | null;
    
    constructor(
        scene: Phaser.Scene,
        pixelToHexFn: (x: number, y: number) => { q: number; r: number },
        getTileFn: (q: number, r: number) => Tile | null,
        getUnitFn: (q: number, r: number) => Unit | null
    ) {
        this.scene = scene;
        this.pixelToHexFn = pixelToHexFn;
        this.getTileFn = getTileFn;
        this.getUnitFn = getUnitFn;
    }
    
    /**
     * Add a layer to the manager
     */
    public addLayer(layer: Layer): void {
        if (this.layers.has(layer.name)) {
            console.warn(`[LayerManager] Layer '${layer.name}' already exists, replacing`);
            this.removeLayer(layer.name);
        }
        this.layers.set(layer.name, layer);
    }
    
    /**
     * Remove a layer from the manager
     */
    public removeLayer(name: string): boolean {
        const layer = this.layers.get(name);
        if (!layer) {
            console.warn(`[LayerManager] Layer '${name}' not found for removal`);
            return false;
        }
        
        layer.destroy();
        this.layers.delete(name);
        return true;
    }
    
    /**
     * Get a layer by name
     */
    public getLayer(name: string): Layer | null {
        return this.layers.get(name) || null;
    }
    
    /**
     * Get all layers sorted by depth (highest first)
     */
    public getSortedLayers(): Layer[] {
        return Array.from(this.layers.values()).sort((a, b) => b.depth - a.depth);
    }
    
    /**
     * Get only interactive layers sorted by depth
     */
    public getInteractiveLayers(): Layer[] {
        return this.getSortedLayers().filter(layer => layer.interactive && layer.visible);
    }
    
    /**
     * Show a layer
     */
    public showLayer(name: string): boolean {
        const layer = this.getLayer(name);
        if (layer) {
            layer.show();
            return true;
        }
        console.warn(`[LayerManager] Cannot show layer '${name}' - not found`);
        return false;
    }
    
    /**
     * Hide a layer
     */
    public hideLayer(name: string): boolean {
        const layer = this.getLayer(name);
        if (layer) {
            layer.hide();
            return true;
        }
        console.warn(`[LayerManager] Cannot hide layer '${name}' - not found`);
        return false;
    }
    
    /**
     * Set layer depth (z-order)
     */
    public setLayerDepth(name: string, depth: number): boolean {
        const layer = this.getLayer(name);
        if (layer) {
            layer.setDepth(depth);
            return true;
        }
        console.warn(`[LayerManager] Cannot set depth for layer '${name}' - not found`);
        return false;
    }
    
    /**
     * Perform hit testing and return ClickContext with layer information
     * Returns null if no layer wants to handle the click
     */
    public getClickContext(pointer: Phaser.Input.Pointer): ClickContext | null {
        // Create complete context with all coordinate spaces
        const hexCoords = this.pixelToHexFn(pointer.worldX, pointer.worldY);
        const context: ClickContext = {
            screenX: pointer.x,
            screenY: pointer.y,
            worldX: pointer.worldX,
            worldY: pointer.worldY,
            hexQ: hexCoords.q,
            hexR: hexCoords.r,
            tile: this.getTileFn(hexCoords.q, hexCoords.r),
            unit: this.getUnitFn(hexCoords.q, hexCoords.r),
            timestamp: Date.now(),
            button: pointer.button
        };
        
        // Process interactive layers by depth (highest first)
        const interactiveLayers = this.getInteractiveLayers();
        
        for (const layer of interactiveLayers) {
            const hitResult = layer.hitTest(context);
            
            if (hitResult === LayerHitResult.CONSUME || hitResult === LayerHitResult.BLOCK) {
                // Set the layer that wants to handle it and return context
                context.layer = layer.name;
                return context;
            }
            // TRANSPARENT continues to next layer
        }
        
        return null; // No layer wants to handle the click
    }
    
    /**
     * Clean up all layers
     */
    public destroy(): void {
        for (const layer of this.layers.values()) {
            layer.destroy();
        }
        this.layers.clear();
    }
}

// =============================================================================
// Base Layer Implementation
// =============================================================================

/**
 * Abstract base class for layers with common functionality
 */
export abstract class BaseLayer implements Layer {
    public readonly name: string;
    public readonly coordinateSpace: CoordinateSpace;
    public readonly interactive: boolean;
    
    public visible: boolean = true;
    public depth: number = 0;
    
    protected scene: Phaser.Scene;
    protected container: Phaser.GameObjects.Container;
    
    constructor(scene: Phaser.Scene, config: LayerConfig) {
        this.scene = scene;
        this.name = config.name;
        this.coordinateSpace = config.coordinateSpace;
        this.interactive = config.interactive ?? false;
        this.visible = config.visible ?? true;
        this.depth = config.depth ?? 0;
        
        // Create container for all layer objects
        this.container = scene.add.container(0, 0);
        this.container.setDepth(this.depth);
        this.container.setAlpha(config.alpha ?? 1.0);
        this.container.setVisible(this.visible);
    }
    
    public abstract hitTest(context: ClickContext): LayerHitResult | null;
    
    public show(): void {
        this.visible = true;
        this.container.setVisible(true);
    }
    
    public hide(): void {
        this.visible = false;
        this.container.setVisible(false);
    }
    
    public setPosition(x: number, y: number): void {
        this.container.setPosition(x, y);
    }
    
    public setScale(x: number, y: number): void {
        this.container.setScale(x, y);
    }
    
    public setAlpha(alpha: number): void {
        this.container.setAlpha(alpha);
    }
    
    public setDepth(depth: number): void {
        this.depth = depth;
        this.container.setDepth(depth);
    }
    
    public destroy(): void {
        this.container.destroy();
    }
}
