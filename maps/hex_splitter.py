#!/usr/bin/env python3
"""
Hex Splitter

Extracts individual hex tiles from WeeWar map images as separate PNG files
with transparent backgrounds to prevent neighbor tile bleeding.
"""

import cv2
import numpy as np
import argparse
from typing import List
from pathlib import Path
from dataclasses import dataclass
from grid_analyzer import HexGridAnalyzer, GridParams
from hex_generator import HexCellGenerator, HexCell


class HexSplitter:
    """Splits hex grid images into individual tile images"""
    
    def __init__(self, output_dir: str = "hex_tiles", debug_mode: bool = False):
        self.output_dir = Path(output_dir)
        self.debug_mode = debug_mode
        
        # Create output directory
        self.output_dir.mkdir(exist_ok=True)
        
        if self.debug_mode:
            print(f"Output directory: {self.output_dir}")
    
    def split_hex_tiles(self, image: np.ndarray, params: GridParams) -> List[str]:
        """Split the image into individual hex tiles"""
        
        # First generate hex cell positions
        generator = HexCellGenerator(debug_mode=False)
        hex_cells = generator.generate_hex_cells(image, params)
        
        if self.debug_mode:
            print(f"Splitting {len(hex_cells)} hex cells")
        
        # Extract each hex tile
        extracted_files = []
        for cell in hex_cells:
            tile_filename = self._extract_hex_tile(image, cell, params)
            if tile_filename:
                extracted_files.append(tile_filename)
        
        print(f"Extracted {len(extracted_files)} hex tiles to {self.output_dir}")
        return extracted_files
    
    def _extract_hex_tile(self, image: np.ndarray, cell: HexCell, params: GridParams) -> str:
        """Extract a single hex tile with transparent background"""
        
        # Calculate extraction region (slightly larger than hex to ensure full coverage)
        margin = 5
        tile_size = max(params.hex_width, params.hex_height) + 2 * margin
        
        # Calculate bounding box
        x_start = int(cell.center_x - tile_size // 2)
        y_start = int(cell.center_y - tile_size // 2)
        x_end = x_start + tile_size
        y_end = y_start + tile_size
        
        # Ensure bounds are within image
        height, width = image.shape[:2]
        x_start = max(0, x_start)
        y_start = max(0, y_start)
        x_end = min(width, x_end)
        y_end = min(height, y_end)
        
        if x_end <= x_start or y_end <= y_start:
            if self.debug_mode:
                print(f"Skipping cell {cell.row},{cell.col} - invalid bounds")
            return None
        
        # Extract the region
        tile_region = image[y_start:y_end, x_start:x_end]
        
        if tile_region.size == 0:
            return None
        
        # Create hexagonal mask
        # Use a larger radius to ensure we capture the full tile
        mask_radius = max(params.hex_width, params.hex_height) // 2 * 0.95  # Slightly smaller than extraction region
        mask = self._create_hex_mask(tile_region.shape, 
                                   cell.center_x - x_start, 
                                   cell.center_y - y_start,
                                   mask_radius)
        
        # Apply mask to create transparent background
        tile_with_alpha = self._apply_hex_mask(tile_region, mask)
        
        # Save the tile
        filename = f"{cell.row}_{cell.col}.png"
        output_path = self.output_dir / filename
        
        cv2.imwrite(str(output_path), tile_with_alpha)
        
        if self.debug_mode:
            print(f"Extracted tile {filename}: center=({cell.center_x:.1f}, {cell.center_y:.1f})")
        
        return str(output_path)
    
    def _create_hex_mask(self, shape: tuple, center_x: float, center_y: float, radius: float) -> np.ndarray:
        """Create a hexagonal mask for clean tile extraction"""
        height, width = shape[:2]
        mask = np.zeros((height, width), dtype=np.uint8)
        
        # Create actual hexagon vertices
        # Regular hexagon has 6 vertices at 60-degree intervals
        hex_points = []
        for i in range(6):
            angle = i * np.pi / 3  # 60 degrees in radians
            x = center_x + radius * np.cos(angle)
            y = center_y + radius * np.sin(angle)
            hex_points.append([int(x), int(y)])
        
        # Convert to numpy array for OpenCV
        hex_points = np.array(hex_points, dtype=np.int32)
        
        # Fill the hexagonal region
        cv2.fillPoly(mask, [hex_points], 255)
        
        return mask
    
    def _apply_hex_mask(self, tile_region: np.ndarray, mask: np.ndarray) -> np.ndarray:
        """Apply hexagonal mask to create transparent background"""
        height, width = tile_region.shape[:2]
        
        # Create RGBA image (BGR + Alpha channel)
        if len(tile_region.shape) == 3:
            # Color image
            tile_rgba = cv2.cvtColor(tile_region, cv2.COLOR_BGR2BGRA)
        else:
            # Grayscale image
            tile_rgba = cv2.cvtColor(tile_region, cv2.COLOR_GRAY2BGRA)
        
        # Apply mask to alpha channel (255 = opaque, 0 = transparent)
        tile_rgba[:, :, 3] = mask
        
        return tile_rgba


def main():
    """Extract individual hex tiles from command line"""
    parser = argparse.ArgumentParser(description='Extract individual hex tiles from WeeWar map images')
    parser.add_argument('--image', type=str, required=True, help='Path to the map image to split')
    parser.add_argument('--output-dir', type=str, default='hex_tiles', help='Output directory for extracted tiles')
    parser.add_argument('--rows', type=int, help='Override number of rows (overrides detection)')
    parser.add_argument('--cols', type=int, help='Override number of columns (overrides detection)')
    parser.add_argument('--vert-spacing', type=float, help='Override vertical spacing in pixels (overrides detection)')
    parser.add_argument('--expected-tiles', type=int, default=34, help='Expected number of tiles in the map')
    parser.add_argument('--debug', action='store_true', help='Enable debug mode with verbose output')
    
    args = parser.parse_args()
    
    # Load image
    image = cv2.imread(args.image)
    
    if image is None:
        print(f"Could not load image: {args.image}")
        return
    
    print(f"Splitting hex tiles from: {args.image}")
    
    # Analyze grid structure
    analyzer = HexGridAnalyzer(debug_mode=args.debug)
    params = analyzer.analyze_grid_structure(image, expected_tiles=args.expected_tiles)
    
    if not params:
        print("Failed to analyze grid structure")
        return
    
    # Apply command-line overrides if provided
    if args.rows is not None:
        print(f"Overriding rows: {params.rows} -> {args.rows}")
        params.rows = args.rows
    
    if args.cols is not None:
        print(f"Overriding cols: {params.cols} -> {args.cols}")
        params.cols = args.cols
    
    if args.vert_spacing is not None:
        print(f"Overriding vertical spacing: {params.spacing_y:.1f} -> {args.vert_spacing}")
        params.spacing_y = args.vert_spacing
    
    print(f"Using grid parameters:")
    print(f"  Grid size: {params.rows} rows x {params.cols} cols = {params.rows * params.cols} total")
    print(f"  Spacing: {params.spacing_x:.1f}x{params.spacing_y:.1f}")
    
    # Split into individual tiles
    splitter = HexSplitter(output_dir=args.output_dir, debug_mode=args.debug)
    extracted_files = splitter.split_hex_tiles(image, params)
    
    print(f"\nSuccessfully extracted {len(extracted_files)} hex tiles")
    print(f"Output directory: {args.output_dir}")


if __name__ == "__main__":
    main()