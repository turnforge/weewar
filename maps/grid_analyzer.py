#!/usr/bin/env python3
"""
Grid Analyzer

Analyzes hex grid structure from edge-detected images.
"""

import cv2
import numpy as np
import math
from typing import Optional, Dict, List
from pathlib import Path
from dataclasses import dataclass


@dataclass
class GridParams:
    """Parameters defining the hex grid structure"""
    hex_width: int          # Width of hex tile in pixels
    hex_height: int         # Height of hex tile in pixels  
    rows: int              # Number of rows
    cols: int              # Number of columns
    row_offset: float      # X offset for odd rows (0 or hex_width/2)
    start_x: int           # X coordinate of first hex center
    start_y: int           # Y coordinate of first hex center
    spacing_x: float       # Horizontal spacing between centers
    spacing_y: float       # Vertical spacing between centers


class HexGridAnalyzer:
    """Analyzes hex grid structure from edge detection"""
    
    def __init__(self, debug_mode: bool = False):
        self.debug_mode = debug_mode
        self.debug_dir = Path("debug_images") if debug_mode else None
        
        if self.debug_mode:
            self.debug_dir.mkdir(exist_ok=True)
    
    def analyze_grid_structure(self, image: np.ndarray, expected_tiles: int = 34) -> Optional[GridParams]:
        """Analyze hex grid structure from map boundary"""
        # Get edge image
        edges = self._get_edge_image(image)
        
        if self.debug_mode:
            cv2.imwrite(str(self.debug_dir / "structure_edges.png"), edges)
        
        # Find map boundaries
        boundaries = self._find_map_boundaries(edges)
        if not boundaries:
            print("Failed to find map boundaries")
            return None
        
        if self.debug_mode:
            print(f"Map boundaries: {boundaries}")
        
        # Calculate hex grid parameters from boundaries and expected tile count
        params = self._calculate_grid_from_boundaries(image, boundaries, expected_tiles)
        
        if self.debug_mode:
            print(f"Calculated grid params: {params}")
        
        return params
    
    def _get_edge_image(self, image: np.ndarray) -> np.ndarray:
        """Get edge-detected image"""
        gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        clahe = cv2.createCLAHE(clipLimit=2.0, tileGridSize=(8,8))
        enhanced = clahe.apply(gray)
        edges = cv2.Canny(enhanced, 30, 90)
        return edges
    
    def _find_map_boundaries(self, edges: np.ndarray) -> Optional[Dict]:
        """Find the overall boundaries of the hex map"""
        height, width = edges.shape
        
        # Find boundaries using 4-direction projections
        boundaries = {}
        
        # Horizontal projection (top/bottom boundaries)
        horizontal_projection = np.sum(edges, axis=1)
        non_zero_rows = np.where(horizontal_projection > np.max(horizontal_projection) * 0.15)[0]
        
        if len(non_zero_rows) < 2:
            return None
        
        boundaries['top'] = non_zero_rows[0]
        boundaries['bottom'] = non_zero_rows[-1]
        boundaries['height'] = boundaries['bottom'] - boundaries['top']
        
        # Vertical projection (left/right boundaries)
        vertical_projection = np.sum(edges, axis=0)
        non_zero_cols = np.where(vertical_projection > np.max(vertical_projection) * 0.15)[0]
        
        if len(non_zero_cols) < 2:
            return None
        
        boundaries['left'] = non_zero_cols[0]
        boundaries['right'] = non_zero_cols[-1]
        boundaries['width'] = boundaries['right'] - boundaries['left']
        
        if self.debug_mode:
            self._save_boundary_debug(edges, boundaries)
        
        return boundaries
    
    def _calculate_grid_from_boundaries(self, image: np.ndarray, boundaries: Dict, expected_tiles: int) -> GridParams:
        """Calculate hex grid parameters from map boundaries and expected tile count"""
        
        map_width = boundaries['width']
        map_height = boundaries['height']
        
        # Estimate grid dimensions based on expected tile count
        # Try different row/col combinations that multiply to approximately expected_tiles
        best_params = None
        best_score = float('inf')
        
        for rows in range(4, 10):  # Reasonable range for hex maps
            cols = expected_tiles // rows
            if cols < 4 or cols > 10:  # Keep reasonable
                continue
            
            if abs(rows * cols - expected_tiles) <= 6:  # Allow some tolerance
                # Calculate hex dimensions for this configuration
                hex_height = map_height // rows
                hex_width = map_width // cols
                
                # Score based on how close we are to expected tiles and reasonable hex proportions
                tile_count_error = abs(rows * cols - expected_tiles)
                aspect_ratio = hex_width / hex_height if hex_height > 0 else 1
                aspect_error = abs(aspect_ratio - 1.0)  # Prefer roughly square hexes
                
                score = tile_count_error + aspect_error * 5
                
                if score < best_score:
                    best_score = score
                    best_params = {
                        'rows': rows,
                        'cols': cols,
                        'hex_width': hex_width,
                        'hex_height': hex_height
                    }
        
        if not best_params:
            # Fallback: use square root approximation
            approx_side = int(np.sqrt(expected_tiles))
            best_params = {
                'rows': approx_side,
                'cols': approx_side + 1,
                'hex_width': map_width // (approx_side + 1),
                'hex_height': map_height // approx_side
            }
        
        # Calculate spacing and starting positions
        spacing_x = map_width / best_params['cols']
        spacing_y = map_height / best_params['rows']
        
        start_x = boundaries['left'] + spacing_x // 2
        start_y = boundaries['top'] + spacing_y // 2
        
        # For hex grids, odd rows are typically offset by half hex width
        row_offset = spacing_x // 2
        
        return GridParams(
            hex_width=best_params['hex_width'],
            hex_height=best_params['hex_height'],
            rows=best_params['rows'],
            cols=best_params['cols'],
            row_offset=row_offset,
            start_x=int(start_x),
            start_y=int(start_y),
            spacing_x=spacing_x,
            spacing_y=spacing_y
        )
    
    def _save_boundary_debug(self, edges: np.ndarray, boundaries: Dict):
        """Save debug image showing detected boundaries"""
        height, width = edges.shape
        
        # Create RGB image for better visualization
        debug_img = cv2.cvtColor(edges, cv2.COLOR_GRAY2BGR)
        
        # Draw boundary lines
        cv2.line(debug_img, (0, boundaries['top']), (width, boundaries['top']), (0, 255, 0), 2)  # Top - green
        cv2.line(debug_img, (0, boundaries['bottom']), (width, boundaries['bottom']), (0, 255, 0), 2)  # Bottom - green
        cv2.line(debug_img, (boundaries['left'], 0), (boundaries['left'], height), (255, 0, 0), 2)  # Left - blue
        cv2.line(debug_img, (boundaries['right'], 0), (boundaries['right'], height), (255, 0, 0), 2)  # Right - blue
        
        # Draw bounding box
        cv2.rectangle(debug_img, 
                     (boundaries['left'], boundaries['top']), 
                     (boundaries['right'], boundaries['bottom']), 
                     (0, 0, 255), 2)  # Red rectangle
        
        # Add text with dimensions
        cv2.putText(debug_img, f"W: {boundaries['width']}, H: {boundaries['height']}", 
                   (10, 30), cv2.FONT_HERSHEY_SIMPLEX, 0.7, (255, 255, 255), 2)
        
        cv2.imwrite(str(self.debug_dir / "map_boundaries.png"), debug_img)
    
    def _save_projection_debug(self, projection: np.ndarray, direction: str, height: int, width: int, transpose: bool = False):
        """Save debug visualization of projection"""
        if direction == "horizontal":
            proj_img = np.zeros((height, width), dtype=np.uint8)
            for y, value in enumerate(projection):
                line_width = int((value / np.max(projection)) * width) if np.max(projection) > 0 else 0
                proj_img[y, :line_width] = 255
        else:  # vertical
            proj_img = np.zeros((height, width), dtype=np.uint8)
            for x, value in enumerate(projection):
                line_height = int((value / np.max(projection)) * height) if np.max(projection) > 0 else 0
                proj_img[-line_height:, x] = 255
        
        cv2.imwrite(str(self.debug_dir / f"{direction}_projection.png"), proj_img)


def main():
    """Test the grid analyzer"""
    # Load test image
    image_path = "../data/Maps/1_files/map-og.png"
    image = cv2.imread(image_path)
    
    if image is None:
        print(f"Could not load image: {image_path}")
        return
    
    # Analyze grid with expected tile count
    analyzer = HexGridAnalyzer(debug_mode=True)
    params = analyzer.analyze_grid_structure(image, expected_tiles=34)
    
    if params:
        print(f"Successfully analyzed grid structure:")
        print(f"  Dimensions: {params.hex_width}x{params.hex_height}")
        print(f"  Grid size: {params.rows} rows x {params.cols} cols = {params.rows * params.cols} total")
        print(f"  Spacing: {params.spacing_x:.1f}x{params.spacing_y:.1f}")
        print(f"  Row offset: {params.row_offset:.1f}")
        print(f"  Start position: ({params.start_x}, {params.start_y})")
    else:
        print("Failed to analyze grid structure")


if __name__ == "__main__":
    main()