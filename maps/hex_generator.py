#!/usr/bin/env python3
"""
Hex Generator

Generates systematic hex cell positions based on grid parameters.
"""

import cv2
import numpy as np
from typing import List
from pathlib import Path
from dataclasses import dataclass
from grid_analyzer import GridParams


@dataclass
class HexCell:
    """Represents a hexagonal cell in the grid"""
    row: int
    col: int
    center_x: float
    center_y: float
    tile_id: int = 0
    confidence: float = 0.0


class HexCellGenerator:
    """Generates hex cells based on grid structure"""
    
    def __init__(self, debug_mode: bool = False):
        self.debug_mode = debug_mode
        self.debug_dir = Path("debug_images") if debug_mode else None
        
        if self.debug_mode:
            self.debug_dir.mkdir(exist_ok=True)
    
    def generate_hex_cells(self, image: np.ndarray, params: GridParams) -> List[HexCell]:
        """Generate hex cells based on analyzed grid structure"""
        hex_cells = []
        
        print(f"Generating {params.rows} x {params.cols} = {params.rows * params.cols} hex cells")
        
        for row in range(params.rows):
            for col in range(params.cols):
                # Calculate hex center position using hex grid geometry
                x = params.start_x + col * params.spacing_x
                
                # Apply row offset for hex pattern
                if row % 2 == 1:  # Odd rows are offset
                    x += params.row_offset
                
                y = params.start_y + row * params.spacing_y
                
                # Check if position is within image bounds
                if 0 <= x < image.shape[1] and 0 <= y < image.shape[0]:
                    hex_cell = HexCell(
                        row=row,
                        col=col,
                        center_x=x,
                        center_y=y,
                        tile_id=0,  # Will be classified later
                        confidence=0.0
                    )
                    hex_cells.append(hex_cell)
        
        if self.debug_mode:
            self._save_debug_hex_cells(image, hex_cells, self.debug_dir / "generated_cells.png")
        
        print(f"Generated {len(hex_cells)} valid hex cells")
        return hex_cells
    
    def _save_debug_hex_cells(self, image: np.ndarray, hex_cells: List[HexCell], path: Path):
        """Save debug image with hex cells marked"""
        debug_img = image.copy()
        
        for cell in hex_cells:
            # Draw center point
            cv2.circle(debug_img, (int(cell.center_x), int(cell.center_y)), 3, (0, 255, 0), -1)
            # Draw hex boundary circle
            cv2.circle(debug_img, (int(cell.center_x), int(cell.center_y)), 15, (255, 0, 0), 2)
            # Draw row/col text
            cv2.putText(debug_img, f"{cell.row},{cell.col}", 
                       (int(cell.center_x)-15, int(cell.center_y)-20), 
                       cv2.FONT_HERSHEY_SIMPLEX, 0.3, (0, 0, 255), 1)
        
        cv2.imwrite(str(path), debug_img)
        print(f"Debug image saved: {path}")


def main():
    """Test the hex generator"""
    from grid_analyzer import HexGridAnalyzer
    
    # Load test image
    image_path = "../data/Maps/1_files/map-og.png"
    image = cv2.imread(image_path)
    
    if image is None:
        print(f"Could not load image: {image_path}")
        return
    
    # First analyze grid structure
    analyzer = HexGridAnalyzer(debug_mode=True)
    params = analyzer.analyze_grid_structure(image)
    
    if not params:
        print("Failed to analyze grid structure")
        return
    
    # Generate hex cells
    generator = HexCellGenerator(debug_mode=True)
    hex_cells = generator.generate_hex_cells(image, params)
    
    print(f"Generated {len(hex_cells)} hex cells")
    for i, cell in enumerate(hex_cells[:10]):  # Show first 10
        print(f"  Cell {i}: row={cell.row}, col={cell.col}, pos=({cell.center_x:.1f}, {cell.center_y:.1f})")


if __name__ == "__main__":
    main()