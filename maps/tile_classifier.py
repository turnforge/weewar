#!/usr/bin/env python3
"""
Tile Classifier

Classifies hex tiles by comparing with reference images.
"""

import cv2
import numpy as np
import json
from typing import Dict, List, Tuple, Optional
from pathlib import Path
from dataclasses import dataclass
from hex_generator import HexCell
from grid_analyzer import GridParams


@dataclass
class TileInfo:
    """Information about a tile type"""
    id: int
    name: str
    image_path: str
    reference_image: Optional[np.ndarray] = None
    dominant_color: Optional[Tuple[int, int, int]] = None


class TileClassifier:
    """Classifies hex tiles using reference images"""
    
    def __init__(self, data_dir: str = "../data", debug_mode: bool = False):
        self.data_dir = Path(data_dir)
        self.debug_mode = debug_mode
        self.debug_dir = Path("debug_images") if debug_mode else None
        self.tile_references: Dict[int, TileInfo] = {}
        
        if self.debug_mode:
            self.debug_dir.mkdir(exist_ok=True)
        
        self._load_tile_references()
    
    def _load_tile_references(self):
        """Load tile reference images"""
        tiles_dir = self.data_dir / "Tiles"
        if not tiles_dir.exists():
            print(f"Tiles directory not found: {tiles_dir}")
            return
        
        for tile_dir in tiles_dir.iterdir():
            if tile_dir.is_dir():
                try:
                    tile_id = int(tile_dir.name)
                    tile_image_path = tile_dir / "0.png"
                    
                    if tile_image_path.exists():
                        tile_info = TileInfo(
                            id=tile_id,
                            name=f"Tile_{tile_id}",
                            image_path=str(tile_image_path)
                        )
                        
                        # Load reference image
                        img = cv2.imread(str(tile_image_path))
                        if img is not None:
                            tile_info.reference_image = img
                            tile_info.dominant_color = self._get_dominant_color(img)
                            self.tile_references[tile_id] = tile_info
                            
                except ValueError:
                    continue
        
        print(f"Loaded {len(self.tile_references)} tile references")
    
    def _get_dominant_color(self, image: np.ndarray) -> Tuple[int, int, int]:
        """Calculate dominant color of an image"""
        # Reshape image to be a list of pixels
        pixels = image.reshape(-1, 3)
        
        # Use k-means clustering to find dominant color
        from sklearn.cluster import KMeans
        kmeans = KMeans(n_clusters=3, random_state=42, n_init=10)
        kmeans.fit(pixels)
        
        # Get the most frequent cluster center
        labels = kmeans.labels_
        counts = np.bincount(labels)
        dominant_cluster = np.argmax(counts)
        dominant_color = kmeans.cluster_centers_[dominant_cluster]
        
        return tuple(map(int, dominant_color))
    
    def classify_hex_cells(self, image: np.ndarray, hex_cells: List[HexCell], params: GridParams) -> List[HexCell]:
        """Classify all hex cells"""
        print(f"Classifying {len(hex_cells)} hex cells...")
        
        classified_cells = []
        
        for i, cell in enumerate(hex_cells):
            tile_id, confidence = self._classify_hex_at_position(image, cell, params)
            
            # Update cell with classification
            cell.tile_id = tile_id
            cell.confidence = confidence
            classified_cells.append(cell)
            
            if self.debug_mode and i < 5:  # Debug first 5 cells
                self._save_hex_region_debug(image, cell, params, i)
        
        # Print classification summary
        tile_counts = {}
        total_confidence = 0
        for cell in classified_cells:
            if cell.tile_id != 0:
                tile_counts[cell.tile_id] = tile_counts.get(cell.tile_id, 0) + 1
                total_confidence += cell.confidence
        
        avg_confidence = total_confidence / len(classified_cells) if classified_cells else 0
        print(f"Classification summary:")
        print(f"  Tile counts: {tile_counts}")
        print(f"  Average confidence: {avg_confidence:.3f}")
        
        return classified_cells
    
    def _classify_hex_at_position(self, image: np.ndarray, cell: HexCell, params: GridParams) -> Tuple[int, float]:
        """Classify hex tile at specific position"""
        # Extract hex region around center
        half_width = params.hex_width // 2
        half_height = params.hex_height // 2
        
        center_x, center_y = int(cell.center_x), int(cell.center_y)
        
        x1 = max(0, center_x - half_width)
        y1 = max(0, center_y - half_height)
        x2 = min(image.shape[1], center_x + half_width)
        y2 = min(image.shape[0], center_y + half_height)
        
        if x2 <= x1 or y2 <= y1:
            return 0, 0.0
        
        hex_region = image[y1:y2, x1:x2]
        
        if hex_region.size == 0:
            return 0, 0.0
        
        return self._classify_hex_region(hex_region)
    
    def _classify_hex_region(self, region: np.ndarray) -> Tuple[int, float]:
        """Classify a hex region using template matching"""
        if not self.tile_references:
            return 0, 0.0
        
        best_match_id = 0
        best_confidence = 0.0
        
        # Try template matching with each reference tile
        for tile_id, tile_info in self.tile_references.items():
            if tile_info.reference_image is None:
                continue
            
            template = tile_info.reference_image
            
            # Resize template to match region size
            h, w = region.shape[:2]
            if h == 0 or w == 0:
                continue
                
            template_resized = cv2.resize(template, (w, h))
            
            # Template matching
            try:
                result = cv2.matchTemplate(region, template_resized, cv2.TM_CCOEFF_NORMED)
                confidence = np.max(result)
                
                if confidence > best_confidence:
                    best_confidence = confidence
                    best_match_id = tile_id
            except cv2.error:
                continue
        
        return best_match_id, best_confidence
    
    def _save_hex_region_debug(self, image: np.ndarray, cell: HexCell, params: GridParams, index: int):
        """Save debug image of extracted hex region"""
        half_width = params.hex_width // 2
        half_height = params.hex_height // 2
        
        center_x, center_y = int(cell.center_x), int(cell.center_y)
        
        x1 = max(0, center_x - half_width)
        y1 = max(0, center_y - half_height)
        x2 = min(image.shape[1], center_x + half_width)
        y2 = min(image.shape[0], center_y + half_height)
        
        if x2 > x1 and y2 > y1:
            hex_region = image[y1:y2, x1:x2]
            if hex_region.size > 0:
                cv2.imwrite(str(self.debug_dir / f"hex_region_{index}_tile_{cell.tile_id}.png"), hex_region)


def main():
    """Test the tile classifier"""
    from grid_analyzer import HexGridAnalyzer
    from hex_generator import HexCellGenerator
    
    # Load test image
    image_path = "../data/Maps/1_files/map-og.png"
    image = cv2.imread(image_path)
    
    if image is None:
        print(f"Could not load image: {image_path}")
        return
    
    # Analyze grid structure
    analyzer = HexGridAnalyzer(debug_mode=True)
    params = analyzer.analyze_grid_structure(image)
    
    if not params:
        print("Failed to analyze grid structure")
        return
    
    # Generate hex cells
    generator = HexCellGenerator(debug_mode=True)
    hex_cells = generator.generate_hex_cells(image, params)
    
    # Classify tiles
    classifier = TileClassifier(debug_mode=True)
    classified_cells = classifier.classify_hex_cells(image, hex_cells, params)
    
    print(f"\nClassified {len(classified_cells)} cells:")
    for cell in classified_cells[:10]:  # Show first 10
        print(f"  Cell ({cell.row},{cell.col}): tile_id={cell.tile_id}, confidence={cell.confidence:.3f}")


if __name__ == "__main__":
    main()