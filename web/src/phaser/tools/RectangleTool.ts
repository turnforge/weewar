import { ShapeTool, HexCoord } from './ShapeTool';
import { World } from '../../World';

/**
 * Rectangle drawing tool.
 *
 * Workflow:
 * 1. First click: Set first corner
 * 2. Mouse move: Show rectangle preview from corner 1 to cursor
 * 3. Second click: Complete rectangle
 * 4. Escape: Cancel
 */
export class RectangleTool implements ShapeTool {
  public readonly name = 'Rectangle';

  private firstCorner: HexCoord | null = null;
  private secondCorner: HexCoord | null = null;
  private filled: boolean = true;
  private world: World;

  constructor(world: World, filled: boolean = true) {
    this.world = world;
    this.filled = filled;
  }

  addPoint(q: number, r: number): boolean {
    if (this.firstCorner === null) {
      // First click: Store first corner
      this.firstCorner = { q, r };
      return true; // More points needed
    } else {
      // Second click: Store second corner and complete
      this.secondCorner = { q, r };
      return false; // Shape complete
    }
  }

  getPreviewTiles(currentQ: number, currentR: number): HexCoord[] {
    if (this.firstCorner === null) {
      return []; // No preview until first point is set
    }

    // Show preview from first corner to current position
    // Always show outline for preview (filled = false)
    const tiles = this.world.rectFrom(
      this.firstCorner.q,
      this.firstCorner.r,
      currentQ,
      currentR,
      false // Always outline for preview
    );

    return tiles.map(([q, r]) => ({ q, r }));
  }

  getResultTiles(): HexCoord[] {
    if (this.firstCorner === null || this.secondCorner === null) {
      return []; // No result if incomplete
    }

    // Generate final rectangle with current fill setting
    const tiles = this.world.rectFrom(
      this.firstCorner.q,
      this.firstCorner.r,
      this.secondCorner.q,
      this.secondCorner.r,
      this.filled
    );

    return tiles.map(([q, r]) => ({ q, r }));
  }

  getAnchorPoints(): HexCoord[] {
    const points: HexCoord[] = [];
    if (this.firstCorner !== null) {
      points.push(this.firstCorner);
    }
    if (this.secondCorner !== null) {
      points.push(this.secondCorner);
    }
    return points;
  }

  reset(): void {
    this.firstCorner = null;
    this.secondCorner = null;
  }

  canComplete(): boolean {
    return this.firstCorner !== null && this.secondCorner !== null;
  }

  requiresKeyboardConfirm(): boolean {
    return false; // Rectangle auto-completes after 2 clicks
  }

  getStatusText(): string {
    if (this.firstCorner === null) {
      return 'Click first corner of rectangle';
    } else if (this.secondCorner === null) {
      return 'Click second corner (or press Escape to cancel)';
    } else {
      return 'Rectangle complete';
    }
  }

  isFilled(): boolean {
    return this.filled;
  }

  setFilled(filled: boolean): void {
    this.filled = filled;
  }
}
