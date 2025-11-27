/**
 * Crossing Layer for Roads and Bridges
 *
 * This layer renders terrain improvements (roads on land, bridges on water)
 * as visual overlays on the hex grid. Roads and bridges are stored separately
 * from tiles to allow independent terrain modification while preserving crossings.
 *
 * Explicit connectivity rendering:
 * - Each crossing stores which of its 6 hex neighbors it connects to via connectsTo array
 * - If a crossing has no connections (all false), draws a horizontal line (left to right edge)
 * - Otherwise, draws lines from center toward each connected direction
 *
 * Depth: 5 (between tiles at 0 and units at 10)
 */

import * as Phaser from 'phaser';
import { BaseLayer, LayerConfig, ClickContext, LayerHitResult } from './LayerSystem';
import { hexToPixel, getNeighborCoord } from './hexUtils';
import { CrossingType, Crossing } from './World';

// =============================================================================
// Crossing Layer
// =============================================================================

/**
 * Layer for rendering roads and bridges with explicit connection-based graphics
 */
export class CrossingLayer extends BaseLayer {
    private crossingGraphics = new Map<string, Phaser.GameObjects.Graphics>();
    private crossingData = new Map<string, Crossing>();
    private tileWidth: number;
    private tileHeight: number;

    constructor(scene: Phaser.Scene, tileWidth: number) {
        super(scene, {
            name: 'crossings',
            coordinateSpace: 'hex',
            interactive: false, // Crossings are visual only, don't consume clicks
            depth: 5, // Between tiles (0) and units (10)
        });
        this.tileWidth = tileWidth;
        this.tileHeight = tileWidth; // Assuming square-ish hexes
    }

    public hitTest(context: ClickContext): LayerHitResult | null {
        // Crossings are visual only, never intercept clicks
        return LayerHitResult.TRANSPARENT;
    }

    /**
     * Get hex key from coordinates
     */
    private getKey(q: number, r: number): string {
        return `${q},${r}`;
    }

    /**
     * Get the direction indices where this crossing has connections
     * Reads directly from connectsTo array
     */
    private getConnectionDirections(q: number, r: number): number[] {
        const crossing = this.crossingData.get(this.getKey(q, r));
        if (!crossing) return [];

        const directions: number[] = [];
        for (let i = 0; i < 6; i++) {
            if (crossing.connectsTo[i]) {
                directions.push(i);
            }
        }
        return directions;
    }

    /**
     * Add or update a crossing at a hex coordinate
     */
    public setCrossing(q: number, r: number, crossing: Crossing): void {
        const key = this.getKey(q, r);

        if (crossing.type === CrossingType.CROSSING_TYPE_UNSPECIFIED) {
            this.removeCrossing(q, r);
            return;
        }

        // Store the crossing data
        this.crossingData.set(key, crossing);

        // Redraw this tile
        this.redrawTile(q, r);
    }

    /**
     * Remove crossing at a hex coordinate
     */
    public removeCrossing(q: number, r: number): void {
        const key = this.getKey(q, r);

        // Remove graphics
        const graphics = this.crossingGraphics.get(key);
        if (graphics) {
            graphics.destroy();
            this.crossingGraphics.delete(key);
        }

        // Remove data
        this.crossingData.delete(key);
    }

    /**
     * Redraw the crossing graphic for a single tile based on its explicit connections
     */
    private redrawTile(q: number, r: number): void {
        const key = this.getKey(q, r);
        const crossing = this.crossingData.get(key);

        if (!crossing) return;

        // Remove existing graphic
        const existing = this.crossingGraphics.get(key);
        if (existing) {
            existing.destroy();
        }

        // Create new graphics
        const graphics = this.scene.add.graphics();
        this.container.add(graphics);

        // Get world position for this tile's center
        const position = hexToPixel(q, r);
        graphics.setPosition(position.x, position.y);

        // Get explicit connection directions
        const connectionDirections = this.getConnectionDirections(q, r);

        if (connectionDirections.length === 0) {
            // No connections - draw default horizontal crossing
            this.drawDefaultCrossing(graphics, crossing.type);
        } else {
            // Draw connections in each specified direction
            for (const direction of connectionDirections) {
                this.drawConnectionInDirection(graphics, q, r, direction, crossing.type);
            }

            // Draw center cap to cover join points (only if multiple connections)
            if (connectionDirections.length > 1) {
                this.drawCenterCap(graphics, crossing.type);
            }
        }

        this.crossingGraphics.set(key, graphics);
    }

    /**
     * Draw the default crossing (horizontal line) when no connections are specified
     */
    private drawDefaultCrossing(graphics: Phaser.GameObjects.Graphics, crossingType: CrossingType): void {
        const halfWidth = this.tileWidth / 2;
        const startX = -halfWidth * 0.7;
        const endX = halfWidth * 0.7;

        if (crossingType === CrossingType.CROSSING_TYPE_ROAD) {
            this.drawRoadSegment(graphics, startX, 0, endX, 0);
        } else {
            this.drawBridgeSegment(graphics, startX, 0, endX, 0);
        }
    }

    // Standard width for crossings
    private static readonly CROSSING_WIDTH = 20;

    /**
     * Draw a center cap to cover join points where multiple segments meet
     */
    private drawCenterCap(graphics: Phaser.GameObjects.Graphics, crossingType: CrossingType): void {
        const width = CrossingLayer.CROSSING_WIDTH;
        const hw = width / 2;

        // Light border circle
        graphics.fillStyle(0x888888, 1.0);
        graphics.fillCircle(0, 0, hw + 1);

        // Dark surface circle
        graphics.fillStyle(0x3a3a3a, 1.0);
        graphics.fillCircle(0, 0, hw);
    }

    /**
     * Draw a filled polygon given an array of [x,y,x,y,...] coordinates
     */
    private fillPolygon(graphics: Phaser.GameObjects.Graphics, points: number[], color: number, alpha: number = 1.0): void {
        graphics.fillStyle(color, alpha);
        graphics.beginPath();
        graphics.moveTo(points[0], points[1]);
        for (let i = 2; i < points.length; i += 2) {
            graphics.lineTo(points[i], points[i + 1]);
        }
        graphics.closePath();
        graphics.fillPath();
    }

    /**
     * Draw a road segment - dark surface with light borders and yellow dashed center
     */
    private drawRoadSegment(graphics: Phaser.GameObjects.Graphics, x1: number, y1: number, x2: number, y2: number): void {
        const width = CrossingLayer.CROSSING_WIDTH;
        const dx = x2 - x1;
        const dy = y2 - y1;
        const length = Math.sqrt(dx * dx + dy * dy);
        if (length === 0) return;

        // Perpendicular vector
        const perpX = (-dy / length);
        const perpY = (dx / length);
        const hw = width / 2;

        // Light border/edge
        this.fillPolygon(graphics, [
            x1 + perpX * (hw + 1), y1 + perpY * (hw + 1),
            x2 + perpX * (hw + 1), y2 + perpY * (hw + 1),
            x2 - perpX * (hw + 1), y2 - perpY * (hw + 1),
            x1 - perpX * (hw + 1), y1 - perpY * (hw + 1),
        ], 0x888888);

        // Dark road surface
        this.fillPolygon(graphics, [
            x1 + perpX * hw, y1 + perpY * hw,
            x2 + perpX * hw, y2 + perpY * hw,
            x2 - perpX * hw, y2 - perpY * hw,
            x1 - perpX * hw, y1 - perpY * hw,
        ], 0x3a3a3a);

        // Yellow dashed center line
        graphics.lineStyle(2, 0xe0b000, 1.0);
        const dashLength = 5;
        const gapLength = 4;
        let currentPos = 2;

        while (currentPos < length - 2) {
            const dashEnd = Math.min(currentPos + dashLength, length - 2);
            const t1 = currentPos / length;
            const t2 = dashEnd / length;
            graphics.lineBetween(
                x1 + dx * t1, y1 + dy * t1,
                x1 + dx * t2, y1 + dy * t2
            );
            currentPos += dashLength + gapLength;
        }
    }

    /**
     * Draw a bridge segment - like road but with support pillars on each side
     */
    private drawBridgeSegment(graphics: Phaser.GameObjects.Graphics, x1: number, y1: number, x2: number, y2: number): void {
        const width = CrossingLayer.CROSSING_WIDTH;
        const dx = x2 - x1;
        const dy = y2 - y1;
        const length = Math.sqrt(dx * dx + dy * dy);
        if (length === 0) return;

        // Perpendicular vector
        const perpX = (-dy / length);
        const perpY = (dx / length);
        const hw = width / 2;

        // Light border/edge (same as road)
        this.fillPolygon(graphics, [
            x1 + perpX * (hw + 1), y1 + perpY * (hw + 1),
            x2 + perpX * (hw + 1), y2 + perpY * (hw + 1),
            x2 - perpX * (hw + 1), y2 - perpY * (hw + 1),
            x1 - perpX * (hw + 1), y1 - perpY * (hw + 1),
        ], 0x888888);

        // Dark bridge surface
        this.fillPolygon(graphics, [
            x1 + perpX * hw, y1 + perpY * hw,
            x2 + perpX * hw, y2 + perpY * hw,
            x2 - perpX * hw, y2 - perpY * hw,
            x1 - perpX * hw, y1 - perpY * hw,
        ], 0x3a3a3a);

        // Draw 3 support pillars on each side: start, middle, end
        const pillarRadius = 3;
        const pillarOffset = hw + pillarRadius + 1;

        graphics.fillStyle(0x505050, 1.0);
        for (const t of [0.1, 0.5, 0.9]) {
            const px = x1 + dx * t;
            const py = y1 + dy * t;

            // Pillar on each side
            graphics.fillCircle(px + perpX * pillarOffset, py + perpY * pillarOffset, pillarRadius);
            graphics.fillCircle(px - perpX * pillarOffset, py - perpY * pillarOffset, pillarRadius);
        }

        // Yellow dashed center line (same as road)
        graphics.lineStyle(2, 0xe0b000, 1.0);
        const dashLength = 5;
        const gapLength = 4;
        let currentPos = 2;

        while (currentPos < length - 2) {
            const dashEnd = Math.min(currentPos + dashLength, length - 2);
            const t1 = currentPos / length;
            const t2 = dashEnd / length;
            graphics.lineBetween(
                x1 + dx * t1, y1 + dy * t1,
                x1 + dx * t2, y1 + dy * t2
            );
            currentPos += dashLength + gapLength;
        }
    }

    /**
     * Draw a connection line from current tile center toward a neighbor in the given direction
     * We draw from center to the edge (halfway to neighbor center)
     */
    private drawConnectionInDirection(
        graphics: Phaser.GameObjects.Graphics,
        fromQ: number, fromR: number,
        direction: number,
        crossingType: CrossingType
    ): void {
        // Get neighbor coordinate in this direction
        const [toQ, toR] = getNeighborCoord(fromQ, fromR, direction);

        // Calculate relative position of neighbor center from our center
        const fromPos = hexToPixel(fromQ, fromR);
        const toPos = hexToPixel(toQ, toR);

        // Direction vector from current tile to neighbor (relative to our position at 0,0)
        const dx = toPos.x - fromPos.x;
        const dy = toPos.y - fromPos.y;

        // Draw from center (0,0) to halfway point (edge of our hex)
        const endX = dx / 2;
        const endY = dy / 2;

        if (crossingType === CrossingType.CROSSING_TYPE_ROAD) {
            this.drawRoadSegment(graphics, 0, 0, endX, endY);
        } else {
            this.drawBridgeSegment(graphics, 0, 0, endX, endY);
        }
    }

    /**
     * Clear all crossings
     */
    public clearAllCrossings(): void {
        for (const graphics of this.crossingGraphics.values()) {
            graphics.destroy();
        }
        this.crossingGraphics.clear();
        this.crossingData.clear();
    }

    /**
     * Load crossings from a map of coordinate keys to Crossing objects
     */
    public loadCrossings(crossings: { [key: string]: Crossing }): void {
        // Clear existing crossings
        this.clearAllCrossings();

        // First, store all crossing data
        for (const [key, crossing] of Object.entries(crossings)) {
            const [q, r] = key.split(',').map(Number);
            if (!isNaN(q) && !isNaN(r) && crossing.type !== CrossingType.CROSSING_TYPE_UNSPECIFIED) {
                this.crossingData.set(key, crossing);
            }
        }

        // Then, draw all tiles
        for (const [key] of this.crossingData) {
            const [q, r] = key.split(',').map(Number);
            this.redrawTile(q, r);
        }
    }

    /**
     * Check if there's a crossing at the given hex coordinate
     */
    public hasCrossing(q: number, r: number): boolean {
        return this.crossingData.has(this.getKey(q, r));
    }

    public destroy(): void {
        this.clearAllCrossings();
        super.destroy();
    }
}
