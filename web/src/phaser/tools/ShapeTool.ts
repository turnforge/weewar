/**
 * ShapeTool Interface
 *
 * Base interface for shape drawing tools in the world editor.
 * Supports multi-click shape creation with live preview.
 */

export interface HexCoord {
  q: number;
  r: number;
}

/**
 * Base interface for shape drawing tools.
 *
 * Shape tools support a multi-click workflow:
 * 1. User clicks to add anchor points
 * 2. Preview updates as mouse moves
 * 3. Shape completes after required points collected
 * 4. User can cancel with Escape key
 */
export interface ShapeTool {
  /**
   * Display name for this tool (e.g., "Rectangle", "Circle")
   */
  readonly name: string;

  /**
   * Add a point to the shape.
   * @param q Hex Q coordinate
   * @param r Hex R coordinate
   * @returns true if more points are needed, false if shape is complete
   */
  addPoint(q: number, r: number): boolean;

  /**
   * Get preview tiles for the shape based on current mouse position.
   * Called during pointermove to show live preview.
   * @param currentQ Current mouse Q coordinate
   * @param currentR Current mouse R coordinate
   * @returns Array of hex coordinates for preview highlights
   */
  getPreviewTiles(currentQ: number, currentR: number): HexCoord[];

  /**
   * Get final tiles for the completed shape.
   * Called when shape is finished to apply terrain/units.
   * @returns Array of hex coordinates for the final shape
   */
  getResultTiles(): HexCoord[];

  /**
   * Get anchor points collected so far.
   * Used for visual feedback (showing markers at clicked positions).
   * @returns Array of collected anchor points
   */
  getAnchorPoints(): HexCoord[];

  /**
   * Reset the tool state, clearing all collected points.
   */
  reset(): void;

  /**
   * Check if the shape can be completed.
   * For shapes with fixed point counts (rectangle, circle), this is automatic.
   * For variable point shapes (polygon, path), user must press Enter.
   * @returns true if shape has enough points to complete
   */
  canComplete(): boolean;

  /**
   * Check if this tool requires keyboard confirmation (Enter key).
   * @returns true if Enter is needed to complete, false if auto-completes
   */
  requiresKeyboardConfirm(): boolean;

  /**
   * Get current status text for user feedback.
   * @returns Instruction text (e.g., "Click first corner", "Click second corner")
   */
  getStatusText(): string;

  /**
   * Whether to render preview as filled or outline.
   */
  isFilled(): boolean;

  /**
   * Set fill mode for the shape.
   * @param filled true for filled, false for outline only
   */
  setFilled(filled: boolean): void;
}
