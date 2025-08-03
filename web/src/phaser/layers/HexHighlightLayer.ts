/**
 * Hex-based highlight layers for game interactions
 * 
 * These layers work in hex coordinate space and provide visual feedback
 * for movement, attack, and selection in the game.
 */

import * as Phaser from 'phaser';
import { BaseLayer, LayerConfig, ClickContext, LayerHitResult } from '../LayerSystem';
import { hexToPixel } from '../hexUtils';
import { MoveOption, AttackOption } from '../../../gen/weewar/v1/games_pb';

// =============================================================================
// Hex Highlight Base Class
// =============================================================================

/**
 * Base class for all hex-based highlight layers
 */
export abstract class HexHighlightLayer extends BaseLayer {
    protected highlights = new Map<string, Phaser.GameObjects.Graphics>();
    protected tileWidth: number;
    
    constructor(scene: Phaser.Scene, config: LayerConfig & { tileWidth: number }) {
        super(scene, { ...config, coordinateSpace: 'hex' });
        this.tileWidth = config.tileWidth;
    }
    
    /**
     * Add highlight at hex coordinate
     */
    protected addHighlight(q: number, r: number, color: number, alpha: number = 0.3, strokeColor?: number, strokeWidth?: number): void {
        const key = `${q},${r}`;
        
        // Remove existing highlight if present
        this.removeHighlight(q, r);
        
        // Create new highlight positioned directly in world coordinates (like tiles/units)
        const highlight = this.scene.add.graphics();
        
        // Set fill style
        highlight.fillStyle(color, alpha);
        if (strokeColor !== undefined && strokeWidth !== undefined) {
            highlight.lineStyle(strokeWidth, strokeColor, 1.0);
        }
        
        // Get world position and set highlight position directly
        const position = hexToPixel(q, r);
        highlight.setPosition(position.x, position.y);
        highlight.setDepth(this.depth);

        const points: Phaser.Geom.Point[] = [];
        const halfWidth = 32
        const halfHeight = 32
        const halfWidth2 = halfWidth;
        
        // Pointy-topped hexagon vertices (starting from top point, going clockwise)
        points.push(new Phaser.Geom.Point(0, - halfHeight))                    // Top
        points.push(new Phaser.Geom.Point(halfWidth2, - halfHeight * 0.5 )) // Top-right
        points.push(new Phaser.Geom.Point(halfWidth2, halfHeight * 0.5 )) // Bottom-right
        points.push(new Phaser.Geom.Point(0, + halfHeight ))                    // Bottom
        points.push(new Phaser.Geom.Point(- halfWidth2, + halfHeight * 0.5 )) // Bottom-left
        points.push(new Phaser.Geom.Point(- halfWidth2, - halfHeight * 0.5 ))  // Top-left
        
        // Create and draw polygon
        const polygon = new Phaser.Geom.Polygon(points);
        highlight.fillPoints(polygon.points, true);
        if (strokeColor !== undefined && strokeWidth !== undefined) {
            highlight.strokePoints(polygon.points, true);
        }
        
        // Store reference (no container needed)
        this.highlights.set(key, highlight);
    }
    
    /**
     * Remove highlight at hex coordinate
     */
    protected removeHighlight(q: number, r: number): void {
        const key = `${q},${r}`;
        const highlight = this.highlights.get(key);
        
        if (highlight) {
            highlight.destroy();
            this.highlights.delete(key);
        }
    }
    
    /**
     * Check if there's a highlight at the given hex coordinate
     */
    protected hasHighlight(q: number, r: number): boolean {
        const key = `${q},${r}`;
        return this.highlights.has(key);
    }
    
    /**
     * Clear all highlights
     */
    protected clearHighlights(): void {
        for (const highlight of this.highlights.values()) {
            highlight.destroy();
        }
        this.highlights.clear();
    }
    
    
    public destroy(): void {
        this.clearHighlights();
        super.destroy();
    }
}

// =============================================================================
// Selection Highlight Layer
// =============================================================================

/**
 * Shows yellow highlight for currently selected unit
 */
export class SelectionHighlightLayer extends HexHighlightLayer {
    private selectedCoord: { q: number; r: number } | null = null;
    
    constructor(scene: Phaser.Scene, tileWidth: number) {
        super(scene, {
            name: 'selection-highlight',
            coordinateSpace: 'hex',
            interactive: false, // Selection highlights don't consume clicks
            depth: 10, // High priority visual
            tileWidth
        });
    }
    
    public hitTest(context: ClickContext): LayerHitResult | null {
        // Selection highlights are visual only, never intercept clicks
        return LayerHitResult.TRANSPARENT;
    }
    
    /**
     * Show selection highlight at hex coordinate
     */
    public selectHex(q: number, r: number): void {
        console.log(`[SelectionHighlightLayer] Selecting hex (${q}, ${r})`);
        
        // Clear previous selection
        this.clearSelection();
        
        // Add new selection highlight (yellow with border)
        this.addHighlight(q, r, 0xFFFF00, 0.3, 0xFFFF00, 4);
        this.selectedCoord = { q, r };
    }
    
    /**
     * Clear current selection
     */
    public clearSelection(): void {
        if (this.selectedCoord) {
            console.log(`[SelectionHighlightLayer] Clearing selection`);
            this.removeHighlight(this.selectedCoord.q, this.selectedCoord.r);
            this.selectedCoord = null;
        }
    }
    
    /**
     * Get currently selected coordinate
     */
    public getSelection(): { q: number; r: number } | null {
        return this.selectedCoord;
    }
}

// =============================================================================
// Movement Highlight Layer
// =============================================================================

/**
 * Shows green highlights for valid movement positions
 */
export class MovementHighlightLayer extends HexHighlightLayer {
    private onMoveCallback?: (q: number, r: number, moveOption: MoveOption) => void;
    private movementOptions: Map<string, MoveOption> = new Map();
    
    constructor(scene: Phaser.Scene, tileWidth: number, onMoveCallback?: (q: number, r: number, moveOption: MoveOption) => void) {
        super(scene, {
            name: 'movement-highlight',
            coordinateSpace: 'hex',
            interactive: true, // Movement highlights consume clicks
            depth: 5, // Below selection, above base map
            tileWidth
        });
        
        this.onMoveCallback = onMoveCallback;
    }
    
    public hitTest(context: ClickContext): LayerHitResult | null {
        if (!this.visible) return null;
        
        // Only consume clicks if there's a movement highlight at this position
        if (this.hasHighlight(context.hexQ, context.hexR)) {
            return LayerHitResult.CONSUME;
        }
        
        return LayerHitResult.TRANSPARENT;
    }
    
    public handleClick(context: ClickContext): boolean {
        console.log(`[MovementHighlightLayer] Movement click at (${context.hexQ}, ${context.hexR})`);
        
        // Get the move option for this coordinate
        const coordKey = `${context.hexQ},${context.hexR}`;
        const moveOption = this.movementOptions.get(coordKey);
        
        if (this.onMoveCallback && moveOption) {
            this.onMoveCallback(context.hexQ, context.hexR, moveOption);
        } else {
            console.warn(`[MovementHighlightLayer] No move option found for (${context.hexQ}, ${context.hexR})`);
        }
        
        return true; // Event handled
    }
    
    /**
     * Show movement options using protobuf MoveOption objects
     */
    public showMovementOptions(moveOptions: MoveOption[]): void {
        console.log(`[MovementHighlightLayer] Showing ${moveOptions.length} movement options`);
        
        // Clear existing highlights and stored options
        this.clearHighlights();
        this.movementOptions.clear();
        
        // Add highlights for each valid movement position and store the MoveOption data
        moveOptions.forEach(moveOption => {
            // Green highlight with subtle border
            this.addHighlight(moveOption.q, moveOption.r, 0x00FF00, 0.2, 0x00FF00, 2);
            
            // Store the move option for click handling
            const coordKey = `${moveOption.q},${moveOption.r}`;
            this.movementOptions.set(coordKey, moveOption);
        });
    }
    
    /**
     * Clear all movement highlights
     */
    public clearMovementOptions(): void {
        console.log(`[MovementHighlightLayer] Clearing movement options`);
        this.clearHighlights();
        this.movementOptions.clear();
    }
    
    /**
     * Get the move option for a specific coordinate (if any)
     */
    public getMoveOptionAt(q: number, r: number): MoveOption | undefined {
        const coordKey = `${q},${r}`;
        return this.movementOptions.get(coordKey);
    }
}

// =============================================================================
// Attack Highlight Layer
// =============================================================================

/**
 * Shows red highlights for valid attack targets
 */
export class AttackHighlightLayer extends HexHighlightLayer {
    private onAttackCallback?: (q: number, r: number) => void;
    
    constructor(scene: Phaser.Scene, tileWidth: number, onAttackCallback?: (q: number, r: number) => void) {
        super(scene, {
            name: 'attack-highlight',
            coordinateSpace: 'hex',
            interactive: true, // Attack highlights consume clicks
            depth: 6, // Same level as movement, both are action highlights
            tileWidth
        });
        
        this.onAttackCallback = onAttackCallback;
    }
    
    public hitTest(context: ClickContext): LayerHitResult | null {
        if (!this.visible) return null;
        
        // Only consume clicks if there's an attack highlight at this position
        if (this.hasHighlight(context.hexQ, context.hexR)) {
            return LayerHitResult.CONSUME;
        }
        
        return LayerHitResult.TRANSPARENT;
    }
    
    public handleClick(context: ClickContext): boolean {
        console.log(`[AttackHighlightLayer] Attack click at (${context.hexQ}, ${context.hexR})`);
        
        if (this.onAttackCallback) {
            this.onAttackCallback(context.hexQ, context.hexR);
        }
        
        return true; // Event handled
    }
    
    /**
     * Show attack options
     */
    public showAttackOptions(coords: Array<{ q: number; r: number }>): void {
        console.log(`[AttackHighlightLayer] Showing ${coords.length} attack options`);
        
        // Clear existing highlights
        this.clearHighlights();
        
        // Add highlights for each valid attack target
        coords.forEach(coord => {
            // Red highlight with border
            this.addHighlight(coord.q, coord.r, 0xFF0000, 0.2, 0xFF0000, 2);
        });
    }
    
    /**
     * Clear all attack highlights
     */
    public clearAttackOptions(): void {
        console.log(`[AttackHighlightLayer] Clearing attack options`);
        this.clearHighlights();
    }
}
