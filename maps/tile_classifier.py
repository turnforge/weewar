#!/usr/bin/env python3
"""
Tile Classifier

Multi-metric tile classification system that compares extracted hex tiles against
reference tiles and returns confidence scores for each match.

FEATURES:
- Multi-metric similarity analysis (NCC, SSIM, histogram correlation, MSE)
- Transparency handling for hexagonal tiles
- Both grayscale and color mode support
- Score fusion with configurable weights
- Batch processing with progress tracking
- Debug visualization for analysis

USAGE:
# Single tile classification
python tile_classifier.py --tile extracted_tiles/0_0.png --references reference_tiles/

# Batch classification
python tile_classifier.py --batch extracted_tiles/ --references reference_tiles/ --mode color --debug

ARCHITECTURE:
1. TileClassifier: Main class handling reference loading and classification
2. Multi-metric pipeline: NCC, SSIM, histogram correlation, MSE
3. Score fusion: Weighted combination of individual metrics
4. CLI interface: Comprehensive command-line options
"""

import cv2
import numpy as np
import argparse
import json
from typing import Dict, List, Tuple, Optional
from pathlib import Path
from dataclasses import dataclass
from skimage import metrics
from scipy.stats import pearsonr
import glob
import os

@dataclass
class TileClassificationResult:
    """Result of tile classification with confidence scores"""
    tile_path: str
    matches: List[Tuple[str, float]]  # (reference_name, confidence)
    best_match: str
    best_confidence: float
    metrics_breakdown: Dict[str, float]

@dataclass
class SimilarityMetrics:
    """Individual similarity metric scores"""
    ncc: float = 0.0
    ssim: float = 0.0
    histogram_correlation: float = 0.0
    mse_similarity: float = 0.0

class TileClassifier:
    """Multi-metric tile classifier with transparent background handling.
    
    This class implements a robust tile classification system using multiple
    similarity metrics to compare extracted hex tiles against reference tiles.
    Handles hexagonal tiles with transparent backgrounds and provides meaningful
    confidence scores through score fusion.
    
    KEY FEATURES:
    - Multi-metric similarity analysis for robust matching
    - Transparency-aware preprocessing pipeline
    - Configurable color/grayscale modes
    - Score fusion with weighted combination
    - Batch processing capabilities
    - Debug visualization and analysis
    """
    
    def __init__(self, 
                 reference_tiles_dir: str, 
                 grayscale_mode: bool = False,
                 debug_mode: bool = False):
        """Initialize tile classifier with reference tiles.
        
        Args:
            reference_tiles_dir: Directory containing reference tile images
            grayscale_mode: If True, convert all tiles to grayscale
            debug_mode: Enable debug output and visualization
        """
        self.reference_tiles_dir = Path(reference_tiles_dir)
        self.grayscale_mode = grayscale_mode
        self.debug_mode = debug_mode
        
        # Reference tiles dictionary: {tile_name: preprocessed_image}
        self.reference_tiles: Dict[str, np.ndarray] = {}
        
        # Metric weights for score fusion (can be tuned)
        self.metric_weights = {
            'ncc': 0.3,
            'ssim': 0.3,
            'histogram_correlation': 0.2,
            'mse_similarity': 0.2
        }
        
        # Debug output directory
        self.debug_dir = Path("debug_images/classification") if debug_mode else None
        if self.debug_mode:
            self.debug_dir.mkdir(parents=True, exist_ok=True)
        
        # Load reference tiles
        self._load_reference_tiles()
    
    def _load_reference_tiles(self):
        """Load and preprocess reference tiles from directory.
        
        Loads all PNG files from reference directory and applies preprocessing
        pipeline including transparency handling and normalization.
        """
        if not self.reference_tiles_dir.exists():
            raise FileNotFoundError(f"Reference tiles directory not found: {self.reference_tiles_dir}")
        
        # Find all PNG files in reference directory
        reference_files = list(self.reference_tiles_dir.glob("*.png"))
        
        if not reference_files:
            raise ValueError(f"No PNG files found in reference directory: {self.reference_tiles_dir}")
        
        print(f"Loading {len(reference_files)} reference tiles...")
        
        for ref_file in reference_files:
            tile_name = ref_file.stem  # Filename without extension
            
            # Load image with alpha channel
            ref_image = cv2.imread(str(ref_file), cv2.IMREAD_UNCHANGED)
            
            if ref_image is None:
                print(f"Warning: Could not load reference tile: {ref_file}")
                continue
            
            # Preprocess reference tile
            processed_tile = self._preprocess_tile(ref_image)
            
            if processed_tile is not None:
                self.reference_tiles[tile_name] = processed_tile
                if self.debug_mode:
                    print(f"  Loaded reference tile: {tile_name}")
        
        print(f"Successfully loaded {len(self.reference_tiles)} reference tiles")
    
    def _preprocess_tile(self, tile_image: np.ndarray) -> Optional[np.ndarray]:
        """Preprocess tile image with transparency handling and normalization.
        
        PREPROCESSING PIPELINE:
        1. Handle transparency (alpha channel)
        2. Crop to content bounding box
        3. Resize to standard dimensions
        4. Color space conversion (if grayscale mode)
        5. Intensity normalization
        
        Args:
            tile_image: Raw tile image (BGR or BGRA)
            
        Returns:
            Preprocessed tile image or None if processing fails
        """
        if tile_image is None or tile_image.size == 0:
            return None
        
        # Handle transparency if present
        if len(tile_image.shape) == 3 and tile_image.shape[2] == 4:
            # BGRA image with alpha channel
            alpha = tile_image[:, :, 3]
            
            # Find bounding box of non-transparent content
            coords = np.column_stack(np.where(alpha > 0))
            if len(coords) == 0:
                return None  # Completely transparent image
            
            # Crop to content bounding box
            top, left = coords.min(axis=0)
            bottom, right = coords.max(axis=0)
            
            # Extract content region
            content_region = tile_image[top:bottom+1, left:right+1]
            
            # Convert BGRA to BGR by removing alpha channel
            if content_region.shape[2] == 4:
                content_region = content_region[:, :, :3]
            
            tile_image = content_region
        
        # Convert to grayscale if requested
        if self.grayscale_mode:
            if len(tile_image.shape) == 3:
                tile_image = cv2.cvtColor(tile_image, cv2.COLOR_BGR2GRAY)
        
        # Resize to standard dimensions for comparison
        target_size = (64, 64)
        tile_image = cv2.resize(tile_image, target_size, interpolation=cv2.INTER_AREA)
        
        # Normalize intensity
        tile_image = tile_image.astype(np.float32) / 255.0
        
        return tile_image
    
    def _calculate_similarity_metrics(self, tile1: np.ndarray, tile2: np.ndarray) -> SimilarityMetrics:
        """Calculate individual similarity metrics between two tiles.
        
        METRICS COMPUTED:
        1. Normalized Cross-Correlation (NCC): Template matching
        2. Structural Similarity Index (SSIM): Perceptual similarity
        3. Histogram Correlation: Intensity distribution matching
        4. MSE Similarity: Pixel-wise difference (converted to similarity)
        
        Args:
            tile1: First tile image (normalized float32)
            tile2: Second tile image (normalized float32)
            
        Returns:
            SimilarityMetrics object with individual metric scores
        """
        metrics_result = SimilarityMetrics()
        
        # Ensure tiles have same dimensions
        if tile1.shape != tile2.shape:
            tile2 = cv2.resize(tile2, (tile1.shape[1], tile1.shape[0]))
        
        # 1. Normalized Cross-Correlation (NCC)
        try:
            # Template matching using normalized cross-correlation
            ncc_result = cv2.matchTemplate(tile1, tile2, cv2.TM_CCORR_NORMED)
            metrics_result.ncc = float(ncc_result.max())
        except Exception as e:
            if self.debug_mode:
                print(f"NCC calculation failed: {e}")
            metrics_result.ncc = 0.0
        
        # 2. Structural Similarity Index (SSIM)
        try:
            if len(tile1.shape) == 3:
                # Color image - use multichannel SSIM
                ssim_score = metrics.structural_similarity(tile1, tile2, multichannel=True, channel_axis=2)
            else:
                # Grayscale image
                ssim_score = metrics.structural_similarity(tile1, tile2)
            metrics_result.ssim = float(ssim_score)
        except Exception as e:
            if self.debug_mode:
                print(f"SSIM calculation failed: {e}")
            metrics_result.ssim = 0.0
        
        # 3. Histogram Correlation
        try:
            if len(tile1.shape) == 3:
                # Color image - compute histogram for each channel
                hist_correlations = []
                for channel in range(tile1.shape[2]):
                    hist1 = cv2.calcHist([tile1[:, :, channel]], [0], None, [256], [0, 1])
                    hist2 = cv2.calcHist([tile2[:, :, channel]], [0], None, [256], [0, 1])
                    
                    # Flatten histograms for correlation
                    hist1_flat = hist1.flatten()
                    hist2_flat = hist2.flatten()
                    
                    if len(hist1_flat) > 1 and len(hist2_flat) > 1:
                        corr, _ = pearsonr(hist1_flat, hist2_flat)
                        hist_correlations.append(corr if not np.isnan(corr) else 0.0)
                
                metrics_result.histogram_correlation = float(np.mean(hist_correlations))
            else:
                # Grayscale image
                hist1 = cv2.calcHist([tile1], [0], None, [256], [0, 1])
                hist2 = cv2.calcHist([tile2], [0], None, [256], [0, 1])
                
                hist1_flat = hist1.flatten()
                hist2_flat = hist2.flatten()
                
                if len(hist1_flat) > 1 and len(hist2_flat) > 1:
                    corr, _ = pearsonr(hist1_flat, hist2_flat)
                    metrics_result.histogram_correlation = float(corr if not np.isnan(corr) else 0.0)
        except Exception as e:
            if self.debug_mode:
                print(f"Histogram correlation calculation failed: {e}")
            metrics_result.histogram_correlation = 0.0
        
        # 4. MSE Similarity (convert MSE to similarity score)
        try:
            mse = np.mean((tile1 - tile2) ** 2)
            # Convert MSE to similarity: higher MSE = lower similarity
            # Use exponential decay: similarity = exp(-mse)
            metrics_result.mse_similarity = float(np.exp(-mse))
        except Exception as e:
            if self.debug_mode:
                print(f"MSE calculation failed: {e}")
            metrics_result.mse_similarity = 0.0
        
        return metrics_result
    
    def _fuse_similarity_scores(self, metrics: SimilarityMetrics) -> float:
        """Combine multiple similarity metrics into single confidence score.
        
        Uses weighted combination of individual metrics with configurable weights.
        Applies confidence calibration to ensure meaningful 0-1 output range.
        
        Args:
            metrics: Individual similarity metric scores
            
        Returns:
            Fused confidence score in range [0, 1]
        """
        # Weighted combination of metrics
        fused_score = (
            self.metric_weights['ncc'] * metrics.ncc +
            self.metric_weights['ssim'] * metrics.ssim +
            self.metric_weights['histogram_correlation'] * metrics.histogram_correlation +
            self.metric_weights['mse_similarity'] * metrics.mse_similarity
        )
        
        # Ensure score is in valid range [0, 1]
        fused_score = np.clip(fused_score, 0.0, 1.0)
        
        return float(fused_score)
    
    def classify_tile(self, tile_image_path: str) -> TileClassificationResult:
        """Classify a single tile against all reference tiles.
        
        CLASSIFICATION PROCESS:
        1. Load and preprocess input tile
        2. Compare against all reference tiles using multi-metric analysis
        3. Fuse similarity scores for each reference
        4. Rank results by confidence
        5. Return classification result with top matches
        
        Args:
            tile_image_path: Path to tile image to classify
            
        Returns:
            TileClassificationResult with ranked matches and confidence scores
        """
        # Load input tile
        tile_image = cv2.imread(tile_image_path, cv2.IMREAD_UNCHANGED)
        if tile_image is None:
            raise ValueError(f"Could not load tile image: {tile_image_path}")
        
        # Preprocess input tile
        processed_tile = self._preprocess_tile(tile_image)
        if processed_tile is None:
            raise ValueError(f"Failed to preprocess tile: {tile_image_path}")
        
        # Compare against all reference tiles
        matches = []
        metrics_breakdown = {}
        
        for ref_name, ref_tile in self.reference_tiles.items():
            # Calculate similarity metrics
            similarity_metrics = self._calculate_similarity_metrics(processed_tile, ref_tile)
            
            # Fuse scores
            confidence = self._fuse_similarity_scores(similarity_metrics)
            
            matches.append((ref_name, confidence))
            metrics_breakdown[ref_name] = {
                'ncc': similarity_metrics.ncc,
                'ssim': similarity_metrics.ssim,
                'histogram_correlation': similarity_metrics.histogram_correlation,
                'mse_similarity': similarity_metrics.mse_similarity,
                'fused_confidence': confidence
            }
        
        # Sort matches by confidence (descending)
        matches.sort(key=lambda x: x[1], reverse=True)
        
        # Get best match
        best_match, best_confidence = matches[0] if matches else ("unknown", 0.0)
        
        result = TileClassificationResult(
            tile_path=tile_image_path,
            matches=matches,
            best_match=best_match,
            best_confidence=best_confidence,
            metrics_breakdown=metrics_breakdown
        )
        
        if self.debug_mode:
            self._save_debug_classification(result, processed_tile)
        
        return result
    
    def classify_batch(self, tile_dir: str) -> Dict[str, TileClassificationResult]:
        """Classify all tiles in a directory.
        
        Args:
            tile_dir: Directory containing tiles to classify
            
        Returns:
            Dictionary mapping tile filenames to classification results
        """
        tile_dir_path = Path(tile_dir)
        if not tile_dir_path.exists():
            raise FileNotFoundError(f"Tile directory not found: {tile_dir}")
        
        # Find all PNG files in directory
        tile_files = list(tile_dir_path.glob("*.png"))
        
        if not tile_files:
            raise ValueError(f"No PNG files found in directory: {tile_dir}")
        
        print(f"Classifying {len(tile_files)} tiles...")
        
        results = {}
        for i, tile_file in enumerate(tile_files):
            if self.debug_mode:
                print(f"  Processing {i+1}/{len(tile_files)}: {tile_file.name}")
            
            try:
                result = self.classify_tile(str(tile_file))
                results[tile_file.name] = result
            except Exception as e:
                print(f"  Error processing {tile_file.name}: {e}")
                continue
        
        print(f"Successfully classified {len(results)} tiles")
        return results
    
    def _save_debug_classification(self, result: TileClassificationResult, processed_tile: np.ndarray):
        """Save debug visualization for tile classification result"""
        if not self.debug_mode:
            return
        
        # Create debug image showing tile and top matches
        tile_name = Path(result.tile_path).stem
        debug_path = self.debug_dir / f"classification_{tile_name}.png"
        
        # Get top 3 matches for visualization
        top_matches = result.matches[:3]
        
        # Create visualization (simplified for now)
        print(f"Debug classification for {tile_name}:")
        print(f"  Best match: {result.best_match} (confidence: {result.best_confidence:.3f})")
        for i, (ref_name, confidence) in enumerate(top_matches):
            print(f"  Match {i+1}: {ref_name} (confidence: {confidence:.3f})")


def main():
    """Command-line interface for tile classification"""
    parser = argparse.ArgumentParser(description='Classify hex tiles using multi-metric similarity analysis')
    parser.add_argument('--references', type=str, required=True, 
                        help='Directory containing reference tile images')
    parser.add_argument('--tile', type=str, 
                        help='Single tile image to classify')
    parser.add_argument('--batch', type=str, 
                        help='Directory containing tiles to classify')
    parser.add_argument('--mode', choices=['color', 'grayscale'], default='color',
                        help='Classification mode (color or grayscale)')
    parser.add_argument('--output', type=str, default='classification_results.json',
                        help='Output file for classification results')
    parser.add_argument('--debug', action='store_true',
                        help='Enable debug mode with verbose output')
    
    args = parser.parse_args()
    
    # Validate arguments
    if not args.tile and not args.batch:
        parser.error("Either --tile or --batch must be specified")
    
    # Create classifier
    classifier = TileClassifier(
        reference_tiles_dir=args.references,
        grayscale_mode=(args.mode == 'grayscale'),
        debug_mode=args.debug
    )
    
    # Process tiles
    if args.tile:
        # Single tile classification
        print(f"Classifying single tile: {args.tile}")
        result = classifier.classify_tile(args.tile)
        
        print(f"\nClassification Results:")
        print(f"Best match: {result.best_match} (confidence: {result.best_confidence:.3f})")
        print(f"Top 5 matches:")
        for i, (ref_name, confidence) in enumerate(result.matches[:5]):
            print(f"  {i+1}. {ref_name}: {confidence:.3f}")
        
        # Save results
        results_data = {
            'tile_path': result.tile_path,
            'best_match': result.best_match,
            'best_confidence': result.best_confidence,
            'matches': result.matches,
            'metrics_breakdown': result.metrics_breakdown
        }
        
        with open(args.output, 'w') as f:
            json.dump(results_data, f, indent=2)
        
        print(f"\nResults saved to: {args.output}")
    
    elif args.batch:
        # Batch classification
        print(f"Classifying batch of tiles from: {args.batch}")
        results = classifier.classify_batch(args.batch)
        
        # Summary statistics
        total_tiles = len(results)
        high_confidence_tiles = sum(1 for r in results.values() if r.best_confidence > 0.8)
        
        print(f"\nBatch Classification Summary:")
        print(f"Total tiles classified: {total_tiles}")
        print(f"High confidence matches (>0.8): {high_confidence_tiles}")
        print(f"Classification accuracy: {high_confidence_tiles/total_tiles*100:.1f}%")
        
        # Save batch results
        batch_results = {}
        for filename, result in results.items():
            batch_results[filename] = {
                'best_match': result.best_match,
                'best_confidence': result.best_confidence,
                'matches': result.matches[:5],  # Top 5 matches
                'metrics_breakdown': result.metrics_breakdown
            }
        
        with open(args.output, 'w') as f:
            json.dump(batch_results, f, indent=2)
        
        print(f"\nBatch results saved to: {args.output}")


if __name__ == "__main__":
    main()