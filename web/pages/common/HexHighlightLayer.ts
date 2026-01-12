/**
 * Hex-based highlight layers for game interactions
 * 
 * These layers work in hex coordinate space and provide visual feedback
 * for movement, attack, and selection in the game.
 */

import * as Phaser from 'phaser';
import { BaseLayer, LayerConfig, ClickContext, LayerHitResult } from './LayerSystem';
import { hexToPixel } from './hexUtils';
import { MoveUnitAction, AttackUnitAction } from '../../gen/wasmjs/lilbattle/v1/models/interfaces';
import { AnimationConfig } from './animations/AnimationConfig';
import { DEFAULT_PLAYER_COLORS } from '../../assets/themes/BaseTheme';

// =============================================================================
// Hex Highlight Base Class
// =============================================================================

/**
 * Base class for all hex-based highlight layers
 */
export abstract class HexHighlightLayer extends BaseLayer {
    protected highlights = new Map<string, Phaser.GameObjects.Graphics>();
    protected paths: Phaser.GameObjects.Graphics[] = [];
    protected tileWidth: number;
    
    constructor(scene: Phaser.Scene, config: LayerConfig & { tileWidth: number }) {
        super(scene, { ...config, coordinateSpace: 'hex' });
        this.tileWidth = config.tileWidth;
    }
    
    /**
     * Add highlight at hex coordinate
     */
    public addHighlight(q: number, r: number, color: number, alpha: number = 0.3, strokeColor?: number, strokeWidth?: number): void {
        const key = `${q},${r}`;
        
        // Remove existing highlight if present
        this.removeHighlight(q, r);
        
        // Create new highlight positioned directly in world coordinates (like tiles/units)
        const highlight = this.scene.add.graphics();
        
        // Add to container so it respects layer depth
        this.container.add(highlight);
        
        // Set fill style
        highlight.fillStyle(color, alpha);
        if (strokeColor !== undefined && strokeWidth !== undefined) {
            highlight.lineStyle(strokeWidth, strokeColor, 1.0);
        }
        
        // Get world position and set highlight position directly
        const position = hexToPixel(q, r);
        highlight.setPosition(position.x, position.y);

        const points: Phaser.Geom.Point[] = [];
        const halfWidth = this.tileWidth / 2;
        const halfHeight = this.tileWidth / 2;  // Use tileWidth for both to maintain aspect ratio
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
    
    /**
     * Add a path through hex coordinates
     * @param coords Array of coordinates as [q1, r1, q2, r2, ...]
     * @param color Hex color for the path
     * @param thickness Line thickness
     * @returns Index of the added path
     */
    public addPath(coords: number[], color: number = 0x00ff00, thickness: number = 3): number {
        if (coords.length < 4) {
            console.warn('Path needs at least 2 points (4 coordinates)');
            return -1;
        }
        
        // Create graphics object for the path
        const pathGraphics = this.scene.add.graphics();
        this.container.add(pathGraphics);
        
        // Set line style
        pathGraphics.lineStyle(thickness, color, 1.0);
        
        // Start the path
        const startPos = hexToPixel(coords[0], coords[1]);
        pathGraphics.moveTo(startPos.x, startPos.y);
        
        // Draw lines through each coordinate pair
        for (let i = 2; i < coords.length; i += 2) {
            const pos = hexToPixel(coords[i], coords[i + 1]);
            pathGraphics.lineTo(pos.x, pos.y);
        }
        
        // Stroke the path
        pathGraphics.strokePath();
        
        // Add small circles at each waypoint for clarity
        pathGraphics.fillStyle(color, 0.8);
        for (let i = 0; i < coords.length; i += 2) {
            const pos = hexToPixel(coords[i], coords[i + 1]);
            pathGraphics.fillCircle(pos.x, pos.y, thickness * 1.5);
        }
        
        // Store and return index
        this.paths.push(pathGraphics);
        return this.paths.length - 1;
    }
    
    /**
     * Remove a path by index
     * @param index Index of the path to remove
     */
    public removePath(index: number): void {
        if (index >= 0 && index < this.paths.length) {
            const path = this.paths[index];
            if (path) {
                path.destroy();
                this.paths.splice(index, 1);
            }
        }
    }
    
    /**
     * Clear all paths
     */
    public clearAllPaths(): void {
        for (const path of this.paths) {
            if (path) {
                path.destroy();
            }
        }
        this.paths = [];
    }
    
    public destroy(): void {
        this.clearHighlights();
        this.clearAllPaths();
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
            depth: 12, // Above units (depth 10) and below labels (depth 15)
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
    private movementOptions: Map<string, MoveUnitAction> = new Map();
    private coordinateTexts: Map<string, Phaser.GameObjects.Text> = new Map();
    private showDebugCoordinates: boolean = true;
    
    constructor(scene: Phaser.Scene, tileWidth: number) {
        super(scene, {
            name: 'movement-highlight',
            coordinateSpace: 'hex',
            interactive: true, // Movement highlights consume clicks
            depth: 11, // Above units (depth 10), below selection (depth 12)
            tileWidth
        });
    }
    
    public hitTest(context: ClickContext): LayerHitResult | null {
        if (!this.visible) return null;
        
        // Only consume clicks if there's a movement highlight at this position
        if (this.hasHighlight(context.hexQ, context.hexR)) {
            return LayerHitResult.CONSUME;
        }
        
        return LayerHitResult.TRANSPARENT;
    }
    
    
    /**
     * Show movement options using protobuf MoveUnitAction objects
     */
    public showMovementOptions(moveOptions: MoveUnitAction[]): void {
        // Clear existing highlights and stored options
        this.clearHighlights();
        this.clearCoordinateTexts();
        this.movementOptions.clear();
        
        // Add highlights for each valid movement position and store the MoveUnitAction data
        moveOptions.forEach(moveOption => {
            // Green highlight with subtle border
            this.addHighlight(moveOption.to!.q, moveOption.to!.r, 0x00FF00, 0.2, 0x00FF00, 2);
            
            // Store the move option for click handling
            const coordKey = `${moveOption.to!.q},${moveOption.to!.r}`;
            this.movementOptions.set(coordKey, moveOption);
            
            // Add debug coordinate text if enabled
            if (this.showDebugCoordinates) {
                this.addCoordinateText(moveOption.to!.q, moveOption.to!.r, moveOption);
            }
        });
    }
    
    /**
     * Clear all movement highlights
     */
    public clearMovementOptions(): void {
        this.clearHighlights();
        this.clearCoordinateTexts();
        this.movementOptions.clear();
    }
    
    /**
     * Toggle debug coordinate display on movement options
     */
    public setShowDebugCoordinates(show: boolean): void {
        this.showDebugCoordinates = show;
        
        // If currently showing movement options, refresh them to show/hide coordinates
        if (this.movementOptions.size > 0) {
            const currentOptions = Array.from(this.movementOptions.values());
            this.showMovementOptions(currentOptions);
        }
    }
    
    /**
     * Add coordinate text overlay at hex position
     */
    private addCoordinateText(q: number, r: number, moveOption: MoveUnitAction): void {
        const key = `${q},${r}`;
        const position = hexToPixel(q, r);
        
        // Create text showing Q/R coordinates and movement cost
        const coordText = `Q:${q} R:${r}`;
        const costText = moveOption.movementCost ? `Cost:${moveOption.movementCost}` : '';
        const displayText = costText ? `${coordText}\n${costText}` : coordText;
        
        const text = this.scene.add.text(position.x, position.y, displayText, {
            fontSize: '10px',
            color: '#ffffff',
            stroke: '#000000',
            strokeThickness: 2,
            align: 'center',
            fontFamily: 'Arial'
        });
        
        // Add to container so it respects layer depth
        this.container.add(text);
        
        text.setOrigin(0.5, 0.5);
        
        this.coordinateTexts.set(key, text);
    }
    
    /**
     * Clear all coordinate text overlays
     */
    private clearCoordinateTexts(): void {
        for (const text of this.coordinateTexts.values()) {
            text.destroy();
        }
        this.coordinateTexts.clear();
    }
    
    /**
     * Get the move option for a specific coordinate (if any)
     */
    public getMoveUnitActionAt(q: number, r: number): MoveUnitAction | undefined {
        const coordKey = `${q},${r}`;
        return this.movementOptions.get(coordKey);
    }
    
    /**
     * Override destroy to clean up coordinate texts
     */
    public destroy(): void {
        this.clearHighlights();
        this.clearCoordinateTexts();
        super.destroy();
    }
}

// =============================================================================
// Attack Highlight Layer
// =============================================================================

/**
 * Shows red highlights for valid attack targets
 */
export class AttackHighlightLayer extends HexHighlightLayer {
    constructor(scene: Phaser.Scene, tileWidth: number) {
        super(scene, {
            name: 'attack-highlight',
            coordinateSpace: 'hex',
            interactive: true, // Attack highlights consume clicks
            depth: 11, // Above units (depth 10), same as movement highlights
            tileWidth
        });
    }
    
    public hitTest(context: ClickContext): LayerHitResult | null {
        if (!this.visible) return null;
        
        // Only consume clicks if there's an attack highlight at this position
        if (this.hasHighlight(context.hexQ, context.hexR)) {
            return LayerHitResult.CONSUME;
        }
        
        return LayerHitResult.TRANSPARENT;
    }
    
    
    /**
     * Show attack options
     */
    public showAttackOptions(coords: AttackUnitAction[]): void {
        // Clear existing highlights
        this.clearHighlights();
        
        // Add highlights for each valid attack target
        coords.forEach(coord => {
            // Red highlight with border
            this.addHighlight(coord.defender!.q, coord.defender!.r, 0xFF0000, 0.2, 0xFF0000, 2);
        });
    }
    
    /**
     * Clear all attack highlights
     */
    public clearAttackOptions(): void {
        this.clearHighlights();
    }
}

// =============================================================================
// Capture Highlight Layer
// =============================================================================

/**
 * Shows purple highlights for valid capture targets (unit's own tile)
 * This is the interactive highlight for clicking to execute capture
 */
export class CaptureHighlightLayer extends HexHighlightLayer {
    constructor(scene: Phaser.Scene, tileWidth: number) {
        super(scene, {
            name: 'capture-highlight',
            coordinateSpace: 'hex',
            interactive: true, // Capture highlights consume clicks
            depth: 11, // Same as movement/attack highlights
            tileWidth
        });
    }

    public hitTest(context: ClickContext): LayerHitResult | null {
        if (!this.visible) return null;

        // Only consume clicks if there's a capture highlight at this position
        if (this.hasHighlight(context.hexQ, context.hexR)) {
            return LayerHitResult.CONSUME;
        }

        return LayerHitResult.TRANSPARENT;
    }

    /**
     * Show capture option at hex coordinate
     */
    public showCaptureOption(q: number, r: number): void {
        // Purple highlight with border to indicate capture option
        this.addHighlight(q, r, 0x9932CC, 0.3, 0x9932CC, 3);
    }

    /**
     * Clear all capture highlights
     */
    public clearCaptureOptions(): void {
        this.clearHighlights();
    }
}

// =============================================================================
// Exhausted Units Highlight Layer
// =============================================================================

/**
 * Shows gray highlights for units with no movement points left
 */
export class ExhaustedUnitsHighlightLayer extends HexHighlightLayer {
    constructor(scene: Phaser.Scene, tileWidth: number) {
        super(scene, {
            name: 'exhausted-units-highlight',
            coordinateSpace: 'hex',
            interactive: false, // Exhausted highlights don't consume clicks
            depth: 13, // Above units (depth 10) and selection (depth 12) for visibility
            tileWidth
        });
    }

    public hitTest(context: ClickContext): LayerHitResult | null {
        // Exhausted highlights are visual only, never intercept clicks
        return LayerHitResult.TRANSPARENT;
    }

    /**
     * Mark a unit as exhausted (no movement points)
     */
    public markExhausted(q: number, r: number): void {
        // Add gray highlight (semi-transparent gray)
        this.addHighlight(q, r, 0x404040, 0.4);
    }

    /**
     * Remove exhausted status from a unit
     */
    public unmarkExhausted(q: number, r: number): void {
        this.removeHighlight(q, r);
    }

    /**
     * Clear all exhausted highlights (e.g., at turn end)
     */
    public clearAllExhausted(): void {
        this.clearHighlights();
    }

    /**
     * Check if a unit is marked as exhausted
     */
    public isMarkedExhausted(q: number, r: number): boolean {
        return this.hasHighlight(q, r);
    }
}

// =============================================================================
// Shape Highlight Layer
// =============================================================================

/**
 * Shows blue outline preview for shape tools (rectangle, circle, ellipse, etc.) during drag operation
 */
export class ShapeHighlightLayer extends HexHighlightLayer {
    constructor(scene: Phaser.Scene, tileWidth: number) {
        super(scene, {
            name: 'shape-preview',
            coordinateSpace: 'hex',
            interactive: false, // Preview is visual only, doesn't consume clicks
            depth: 14, // Above exhausted layer (13) for clear visibility
            tileWidth
        });
    }

    public hitTest(context: ClickContext): LayerHitResult | null {
        // Shape preview is visual only, never intercept clicks
        return LayerHitResult.TRANSPARENT;
    }

    /**
     * Show shape outline preview for tiles that will be affected
     * @param outlineTiles Array of {q, r} coordinates representing the outline
     */
    public showShapeOutline(outlineTiles: Array<{q: number, r: number}>): void {
        // Clear existing highlights
        this.clearHighlights();

        // Add blue outline highlights for each tile
        outlineTiles.forEach(coord => {
            // Blue highlight with bright blue stroke for visibility
            this.addHighlight(coord.q, coord.r, 0x0080FF, 0.2, 0x00BFFF, 3);
        });
    }

    /**
     * Clear the shape preview
     */
    public clearPreview(): void {
        this.clearHighlights();
    }
}

// =============================================================================
// Capturing Flag Layer
// =============================================================================

/**
 * Shows animated flag indicators on tiles being captured.
 * Uses a flag sprite with wave animation to indicate capture in progress.
 */
export class CapturingFlagLayer extends BaseLayer {
    private flags = new Map<string, {
        sprite: Phaser.GameObjects.Graphics,
        tween: Phaser.Tweens.Tween
    }>();
    private tileWidth: number;

    constructor(scene: Phaser.Scene, tileWidth: number) {
        super(scene, {
            name: 'capturing-flag',
            coordinateSpace: 'hex',
            interactive: false, // Flags don't consume clicks
            depth: 14 // Above exhausted layer (13) for visibility
        });
        this.tileWidth = tileWidth;
    }

    public hitTest(context: ClickContext): LayerHitResult | null {
        // Flags are visual only, never intercept clicks
        return LayerHitResult.TRANSPARENT;
    }

    /**
     * Show a capturing flag at hex coordinate with player color
     * @param q Hex Q coordinate
     * @param r Hex R coordinate
     * @param player Player ID for flag color (defaults to 1)
     */
    public showFlag(q: number, r: number, player: number = 1): void {
        const key = `${q},${r}`;

        // Remove existing flag if present
        this.hideFlag(q, r);

        // Get world position
        const position = hexToPixel(q, r);

        // Get player color
        const playerColor = DEFAULT_PLAYER_COLORS[player] || DEFAULT_PLAYER_COLORS[1];
        const primaryColor = parseInt(playerColor.primary.replace('#', ''), 16);

        // Create flag using graphics (simple triangular pennant)
        const flag = this.scene.add.graphics();
        this.container.add(flag);

        // Position flag at top-right of the tile
        const flagX = position.x + this.tileWidth * 0.25;
        const flagY = position.y - this.tileWidth * 0.25;
        flag.setPosition(flagX, flagY);

        // Draw flag pole
        flag.lineStyle(2, 0x8B4513); // Brown pole
        flag.moveTo(0, -15);
        flag.lineTo(0, 15);
        flag.strokePath();

        // Draw flag pennant (triangle) with player color fill and black stroke
        flag.fillStyle(primaryColor, 1);
        flag.lineStyle(2, 0x000000); // Black stroke for prominence
        flag.beginPath();
        flag.moveTo(0, -15);
        flag.lineTo(20, -10);
        flag.lineTo(0, -5);
        flag.closePath();
        flag.fillPath();
        flag.strokePath();

        // Create wave animation by rotating the flag slightly
        const tween = this.scene.tweens.add({
            targets: flag,
            angle: { from: -5, to: 5 },
            duration: AnimationConfig.FLAG_WAVE_DURATION,
            yoyo: true,
            repeat: -1,
            ease: 'Sine.easeInOut'
        });

        // Store reference
        this.flags.set(key, { sprite: flag, tween });
    }

    /**
     * Hide a capturing flag at hex coordinate
     */
    public hideFlag(q: number, r: number): void {
        const key = `${q},${r}`;
        const flagData = this.flags.get(key);

        if (flagData) {
            flagData.tween.stop();
            flagData.sprite.destroy();
            this.flags.delete(key);
        }
    }

    /**
     * Clear all capturing flags
     */
    public clearAllFlags(): void {
        for (const [key, flagData] of this.flags) {
            flagData.tween.stop();
            flagData.sprite.destroy();
        }
        this.flags.clear();
    }

    /**
     * Check if there's a flag at the given hex coordinate
     */
    public hasFlag(q: number, r: number): boolean {
        const key = `${q},${r}`;
        return this.flags.has(key);
    }

    public destroy(): void {
        this.clearAllFlags();
        super.destroy();
    }
}
