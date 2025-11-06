/**
 * Hex coordinate utility functions
 * These match the Go implementation from lib/map.go
 */

export interface HexCoord {
    q: number;
    r: number;
}

export interface PixelCoord {
    x: number;
    y: number;
}

// Tile constants matching the Go implementation
export const TILE_WIDTH = 64;
export const TILE_HEIGHT = 64;
export const Y_INCREMENT = 48;

/**
 * Convert hex coordinates to pixel coordinates
 * Matches lib/map.go CenterXYForTile
 */
export function hexToPixel(q: number, r: number, tileWidth=TILE_WIDTH, tileHeight=TILE_HEIGHT, yIncrement=Y_INCREMENT): PixelCoord {
  // Match the Go implementation from map.go CenterXYForTile
  const { row, col } = hexToRowCol(q, r);

  let y = yIncrement * row;
  let x = tileWidth * col;

  if ((row & 1) === 1) {
    x += tileWidth / 2;
  }
  return { x, y };
}

/**
 * Convert pixel coordinates to hex coordinates
 * Matches lib/map.go XYToQR
 */
export function pixelToHex(x: number, y: number, tileWidth=TILE_WIDTH, tileHeight=TILE_HEIGHT, yIncrement=Y_INCREMENT): HexCoord {
    // Match the Go implementation from map.go XYToQR
  const row = Math.floor((y + tileHeight / 2) / yIncrement);
  let halfDists = Math.floor(1 + Math.abs(x * 2 / tileWidth));
  if ((row & 1) !== 0) {
    halfDists = Math.floor(1 + Math.abs((x - tileWidth / 2) * 2 / tileWidth));
  }

  let col = Math.floor(halfDists / 2);
  if (x < 0) {
    col = -col;
  }

  return rowColToHex(row, col);
}

/**
 * Convert row/col coordinates to hex coordinates
 * RowColToHex: oddr_to_cube conversion
 */
export function rowColToHex(row: number, col: number): HexCoord {
    const x = col - Math.floor((row - (row & 1)) / 2);
    const z = row;
    const q = x;
    const r = z;
    return { q, r };
}

/**
 * Convert hex coordinates to row/col coordinates
 * HexToRowCol: cube_to_oddr conversion
 */
export function hexToRowCol(q: number, r: number): { row: number; col: number } {
    const row = r;
    const col = q + Math.floor((r - (r & 1)) / 2);
    return { row, col };
}


export const AxialNeighborDeltas = [
	{q: -1, r: 0}, // LEFT
	{q: 0, r: -1}, // TOP_LEFT
	{q: 1, r: -1}, // TOP_RIGHT
	{q: 1, r: 0},  // RIGHT
	{q: 0, r: 1},  // BOTTOM_RIGHT
	{q: -1, r: 1}, // BOTTOM_LEFT
]

export function axialNeighbors(q: number, r: number): [number, number][] {
  let out = [] as any;
	for (var i = 0;i < 6;i++) {
    out.push([q + AxialNeighborDeltas[i].q, r + AxialNeighborDeltas[i].r])
	}
  return out
}
