#!/usr/bin/env python3
"""
WeeWar Map Extractor - Modular Version

Reverse engineers WeeWar maps using modular components.
"""

import cv2
import numpy as np
import json
import argparse
from typing import Dict, List, Optional
from pathlib import Path

from grid_analyzer import HexGridAnalyzer, GridParams
from hex_generator import HexCellGenerator, HexCell
from tile_classifier import TileClassifier
from hex_grid_renderer import HexGridRenderer


class ModularMapExtractor:
    """Main class for extracting map data using modular components"""
    
    def __init__(self, data_dir: str = "../data", debug_mode: bool = False):
        self.data_dir = Path(data_dir)
        self.debug_mode = debug_mode
        self.maps_data: Dict[int, dict] = {}
        
        # Initialize modular components
        self.grid_analyzer = HexGridAnalyzer(debug_mode=debug_mode)
        self.hex_generator = HexCellGenerator(debug_mode=debug_mode)
        self.tile_classifier = TileClassifier(str(data_dir), debug_mode=debug_mode)
        self.renderer = HexGridRenderer()
        
        self._load_maps_data()
    
    def _load_maps_data(self):
        """Load map metadata"""
        maps_file = self.data_dir / "weewar-maps.json"
        if maps_file.exists():
            with open(maps_file, 'r') as f:
                data = json.load(f)
                self.maps_data = {m['id']: m for m in data['maps']}
        print(f"Loaded {len(self.maps_data)} maps")
    
    def extract_map(self, map_id: int) -> Optional[List[List[int]]]:
        """Extract map grid from preview image using modular approach"""
        if map_id not in self.maps_data:
            print(f"Map {map_id} not found in data")
            return None
        
        map_data = self.maps_data[map_id]
        image_path = self.data_dir / "Maps" / map_data['imageURL'].replace('./', '')
        
        if not image_path.exists():
            print(f"Image not found: {image_path}")
            return None
        
        # Load image
        image = cv2.imread(str(image_path))
        if image is None:
            print(f"Could not load image: {image_path}")
            return None
        
        print(f"Processing map {map_id}: {map_data['name']}")
        print(f"Expected tiles: {map_data['tileCount']}")
        print(f"Image size: {image.shape}")
        
        # Step 1: Analyze grid structure
        print("\\n=== Step 1: Analyzing grid structure ===")
        grid_params = self.grid_analyzer.analyze_grid_structure(image)
        
        if grid_params is None:
            print("Failed to analyze grid structure")
            return None
        
        print(f"Grid parameters: {grid_params}")
        
        # Step 2: Generate hex cells
        print("\\n=== Step 2: Generating hex cells ===")
        hex_cells = self.hex_generator.generate_hex_cells(image, grid_params)
        
        if not hex_cells:
            print("Failed to generate hex cells")
            return None
        
        # Step 3: Classify tiles
        print("\\n=== Step 3: Classifying tiles ===")
        classified_cells = self.tile_classifier.classify_hex_cells(image, hex_cells, grid_params)
        
        # Step 4: Convert to grid format
        print("\\n=== Step 4: Converting to grid format ===")
        grid = self._cells_to_grid(classified_cells)
        
        # Step 5: Validate results
        print("\\n=== Step 5: Validation ===")
        self._validate_extraction(grid, map_data)
        
        return grid
    
    def _cells_to_grid(self, hex_cells: List[HexCell]) -> List[List[int]]:
        """Convert hex cells to 2D grid format"""
        if not hex_cells:
            return []
        
        # Find grid dimensions
        max_row = max(cell.row for cell in hex_cells)
        max_col = max(cell.col for cell in hex_cells)
        
        # Initialize grid
        grid = [[0 for _ in range(max_col + 1)] for _ in range(max_row + 1)]
        
        # Fill grid
        for cell in hex_cells:
            grid[cell.row][cell.col] = cell.tile_id
        
        return grid
    
    def _validate_extraction(self, grid: List[List[int]], map_data: dict):
        """Validate extracted grid against expected map data"""
        # Count tiles
        tile_counts = {}
        total_tiles = 0
        
        for row in grid:
            for tile_id in row:
                if tile_id != 0:  # Skip empty tiles
                    tile_counts[tile_id] = tile_counts.get(tile_id, 0) + 1
                    total_tiles += 1
        
        print(f"Extracted {total_tiles} tiles (expected: {map_data['tileCount']})")
        print(f"Tile distribution: {tile_counts}")
        
        if 'tiles' in map_data:
            print(f"Expected distribution: {map_data['tiles']}")
        
        accuracy = (total_tiles / map_data['tileCount']) * 100 if map_data['tileCount'] > 0 else 0
        print(f"Tile count accuracy: {accuracy:.1f}%")
    
    def render_map(self, grid: List[List[int]]) -> np.ndarray:
        """Render extracted map for visualization"""
        tile_references = self.renderer.load_tile_references(self.data_dir / "Tiles")
        return self.renderer.render_hex_grid(grid, tile_references, highlight_errors=True)
    
    def generate_validation_report(self, map_id: int, output_dir: str = "outputs"):
        """Generate comprehensive validation report"""
        output_path = Path(output_dir)
        output_path.mkdir(parents=True, exist_ok=True)
        
        if map_id not in self.maps_data:
            print(f"Map {map_id} not found")
            return
        
        map_data = self.maps_data[map_id]
        
        # Extract map
        grid = self.extract_map(map_id)
        if not grid:
            print("Failed to extract map")
            return
        
        # Get original image path
        image_path = self.data_dir / "Maps" / map_data['imageURL'].replace('./', '')
        
        # Load tile references
        tile_references = self.renderer.load_tile_references(self.data_dir / "Tiles")
        
        # Save validation images
        self.renderer.save_validation_images(
            output_path, map_id, image_path, grid, tile_references
        )
        
        print(f"Validation report generated in {output_path}")


def main():
    """Main function with command line interface"""
    parser = argparse.ArgumentParser(description='Modular WeeWar Map Extractor')
    parser.add_argument('--map-id', type=int, help='Extract specific map ID')
    parser.add_argument('--debug', action='store_true', help='Enable debug mode')
    parser.add_argument('--validate', action='store_true', help='Generate validation report')
    parser.add_argument('--output-dir', default='outputs', help='Output directory')
    
    args = parser.parse_args()
    
    extractor = ModularMapExtractor(debug_mode=args.debug)
    
    if args.map_id:
        if args.validate:
            extractor.generate_validation_report(args.map_id, args.output_dir)
        else:
            grid = extractor.extract_map(args.map_id)
            if grid:
                print(f"\\nExtracted grid ({len(grid)}x{len(grid[0]) if grid else 0}):")
                for row in grid:
                    print(row)
                
                # Render the map
                rendered = extractor.render_map(grid)
                output_file = f"map_{args.map_id}_rendered.png"
                cv2.imwrite(output_file, rendered)
                print(f"Rendered map saved as {output_file}")
    else:
        # Test with first available map
        if extractor.maps_data:
            first_map_id = next(iter(extractor.maps_data.keys()))
            print(f"Testing with Map {first_map_id}")
            extractor.generate_validation_report(first_map_id, args.output_dir)


if __name__ == "__main__":
    main()