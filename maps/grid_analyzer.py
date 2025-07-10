#!/usr/bin/env python3
"""
Grid Analyzer

Analyzes hex grid structure from edge-detected images.
"""

from ipdb import set_trace
import cv2
import numpy as np
import math
from typing import Optional, Dict, List
from pathlib import Path
from dataclasses import dataclass

MIN_LINE_LENGTH = 10 # Only keep lines with minimum length when detecting vertical lines
MAX_ROWS = 20
DEFAULT_NUM_STARTING_COLS = 7
DEFAULT_NUM_STARTING_ROWS = 7
HEIGHT_FACTOR = 1
# HEIGHT_FACTOR = 1.15
# HEIGHT_FACTOR = math.sqrt(5) * 0.75

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
    
    def analyze_grid_structure(self, image: np.ndarray, expected_tiles: int) -> Optional[GridParams]:
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
        """Find map boundaries using 4-directional edge images and OR combination"""
        height, width = edges.shape
        
        # Get 4-directional boundary edge images
        projections = self._get_4_directional_projections(edges)
        
        if self.debug_mode:
            self._save_projection_debug_4dir(projections, height, width)
        
        # Combine all 4 edge images using OR operation to get outer boundary
        combined_boundary = np.zeros((height, width), dtype=np.uint8)
        for direction, edge_img in projections.items():
            combined_boundary = cv2.bitwise_or(combined_boundary, edge_img)
        
        if self.debug_mode:
            cv2.imwrite(str(self.debug_dir / "combined_boundary.png"), combined_boundary)
        
        # Find boundaries from the combined edge image
        boundaries = {}
        
        # Find the actual extent of the combined boundary
        coords = np.where(combined_boundary > 0)
        if len(coords[0]) == 0:
            print("No boundary pixels found")
            return None
        
        boundaries['top'] = np.min(coords[0])
        boundaries['bottom'] = np.max(coords[0])
        boundaries['left'] = np.min(coords[1])
        boundaries['right'] = np.max(coords[1])
        
        boundaries['height'] = boundaries['bottom'] - boundaries['top']
        boundaries['width'] = boundaries['right'] - boundaries['left']
        
        # Analyze hex grid using geometric constraints instead of pattern spacing
        hex_info = self._analyze_hex_geometry(combined_boundary, boundaries, projections)
        boundaries.update(hex_info)
        
        if self.debug_mode:
            self._save_boundary_debug(edges, boundaries)
            print(f"Geometric analysis results: {hex_info}")
        
        return boundaries
    
    def _get_4_directional_projections(self, edges: np.ndarray) -> Dict[str, np.ndarray]:
        """Get boundary edge images from 4 directions, filtering for vertical edges in left/right projections"""
        height, width = edges.shape
        
        projections = {}
        edge_thickness = 3  # Thicker edges to handle jaggedness and improve segment detection
        
        # Create a mask for vertical edges using gradient analysis
        vertical_edge_mask = self._create_vertical_edge_mask(edges)
        
        # Create 4 separate edge images (same size as original)
        view_from_top = np.zeros((height, width), dtype=np.uint8)
        view_from_bottom = np.zeros((height, width), dtype=np.uint8)
        view_from_left = np.zeros((height, width), dtype=np.uint8)
        view_from_right = np.zeros((height, width), dtype=np.uint8)
        
        # View from top: for each column, mark the first edge pixel from top
        for col in range(width):
            column_data = edges[:, col]
            if np.any(column_data > 0):
                first_edge = np.argmax(column_data > 0)
                # Mark the edge pixel with some thickness
                for t in range(edge_thickness):
                    if first_edge + t < height:
                        view_from_top[first_edge + t, col] = 255
        
        # View from bottom: for each column, mark the first edge pixel from bottom
        for col in range(width):
            column_data = edges[:, col]
            if np.any(column_data > 0):
                last_edge = height - 1 - np.argmax(column_data[::-1] > 0)
                # Mark the edge pixel with some thickness
                for t in range(edge_thickness):
                    if last_edge - t >= 0:
                        view_from_bottom[last_edge - t, col] = 255
        
        # View from left: for each row, mark the first edge pixel from left
        for row in range(height):
            row_data = edges[row, :]
            if np.any(row_data > 0):
                first_edge = np.argmax(row_data > 0)
                # Mark the edge pixel with some thickness
                for t in range(edge_thickness):
                    if first_edge + t < width:
                        view_from_left[row, first_edge + t] = 255
        
        # View from right: for each row, mark the first edge pixel from right
        for row in range(height):
            row_data = edges[row, :]
            if np.any(row_data > 0):
                last_edge = width - 1 - np.argmax(row_data[::-1] > 0)
                # Mark the edge pixel with some thickness
                for t in range(edge_thickness):
                    if last_edge - t >= 0:
                        view_from_right[row, last_edge - t] = 255
        
        # Create vertical-only versions for better column detection
        view_from_left_vertical = np.zeros((height, width), dtype=np.uint8)
        view_from_right_vertical = np.zeros((height, width), dtype=np.uint8)
        
        # View from left (vertical edges only): for each row, mark the first VERTICAL edge pixel from left
        for row in range(height):
            row_data = edges[row, :]
            vertical_row_data = vertical_edge_mask[row, :]
            if np.any(row_data > 0) and np.any(vertical_row_data > 0):
                # Find first edge pixel that is also vertical
                edge_positions = np.where(row_data > 0)[0]
                vertical_positions = np.where(vertical_row_data > 0)[0]
                
                # Find intersection of edge and vertical positions
                vertical_edge_positions = np.intersect1d(edge_positions, vertical_positions)
                
                if len(vertical_edge_positions) > 0:
                    first_edge = vertical_edge_positions[0]
                    # Mark the edge pixel with some thickness
                    for t in range(edge_thickness):
                        if first_edge + t < width:
                            view_from_left_vertical[row, first_edge + t] = 255
        
        # View from right (vertical edges only): for each row, mark the first VERTICAL edge pixel from right
        for row in range(height):
            row_data = edges[row, :]
            vertical_row_data = vertical_edge_mask[row, :]
            if np.any(row_data > 0) and np.any(vertical_row_data > 0):
                # Find last edge pixel that is also vertical
                edge_positions = np.where(row_data > 0)[0]
                vertical_positions = np.where(vertical_row_data > 0)[0]
                
                # Find intersection of edge and vertical positions
                vertical_edge_positions = np.intersect1d(edge_positions, vertical_positions)
                
                if len(vertical_edge_positions) > 0:
                    last_edge = vertical_edge_positions[-1]
                    # Mark the edge pixel with some thickness
                    for t in range(edge_thickness):
                        if last_edge - t >= 0:
                            view_from_right_vertical[row, last_edge - t] = 255
        
        projections['view_from_top'] = view_from_top
        projections['view_from_bottom'] = view_from_bottom
        projections['view_from_left'] = view_from_left
        projections['view_from_right'] = view_from_right
        projections['view_from_left_vertical'] = view_from_left_vertical
        projections['view_from_right_vertical'] = view_from_right_vertical
        projections['vertical_edge_mask'] = vertical_edge_mask
        
        return projections
    
    def _create_vertical_edge_mask(self, edges: np.ndarray) -> np.ndarray:
        """Create a mask that identifies vertical line structures using morphological operations"""
        height, width = edges.shape
        
        # Use morphological operations to detect vertical line structures
        # Create a vertical kernel (tall and narrow)
        vertical_kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (1, 20))
        
        # Apply morphological opening to detect vertical lines
        vertical_lines = cv2.morphologyEx(edges, cv2.MORPH_OPEN, vertical_kernel)
        
        # Dilate slightly to connect nearby vertical segments
        dilate_kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (3, 5))
        vertical_mask = cv2.dilate(vertical_lines, dilate_kernel, iterations=1)
        
        if self.debug_mode:
            cv2.imwrite(str(self.debug_dir / "vertical_edge_mask.png"), vertical_mask)
            print(f"Created vertical edge mask with {np.sum(vertical_mask > 0)} vertical edge pixels")
        
        return vertical_mask
    
    def _analyze_hex_geometry(self, combined_boundary: np.ndarray, boundaries: Dict, projections: Dict[str, np.ndarray]) -> Dict:
        """Analyze hex grid using geometric constraints from boundary measurements"""
        hex_info = {}
        
        # Detect vertical lines to find true column boundaries
        vertical_lines = self._detect_vertical_lines(combined_boundary, projections, method=getattr(self, 'detection_method', 'column_based'))
        hex_info.update(vertical_lines)
        
        # Extract vertical line segments for better span analysis
        line_segments = self._extract_vertical_line_segments(projections)
        hex_info.update(line_segments)
        
        # Extract all vertical segments from unified projection
        all_segments = self._extract_all_vertical_segments(projections)
        
        # Group segments into rows
        row_groups = self._get_matching_segments_in_rows(all_segments)
        
        if self.debug_mode and row_groups:
            # Use the same vertical edge mask that was used for segment extraction
            unified_vertical = projections.get('vertical_edge_mask')
            if unified_vertical is None:
                # Fallback to old method if vertical_edge_mask not available
                height, width = projections['view_from_left_vertical'].shape
                unified_vertical = np.zeros((height, width), dtype=np.uint8)
                unified_vertical = cv2.bitwise_or(unified_vertical, projections['view_from_left_vertical'])
                unified_vertical = cv2.bitwise_or(unified_vertical, projections['view_from_right_vertical'])
            self._save_row_groups_debug(unified_vertical, row_groups)
        
        hex_info['all_segments'] = all_segments
        hex_info['row_groups'] = row_groups
        
        # Calculate spans from matched pairs
        spans = self._calculate_spans_from_matched_pairs(
            line_segments.get('left_lines', []), 
            line_segments.get('right_lines', []), 
            line_segments.get('matched_pairs', [])
        )
        hex_info['line_segment_spans'] = spans
        
        # Use advanced constraint solver with matched pairs
        if line_segments.get('matched_pairs'):
            constraint_solution = self._solve_constraints_from_matched_pairs(
                matched_pairs=line_segments.get('matched_pairs', []),
                left_lines=line_segments.get('left_lines', []),
                right_lines=line_segments.get('right_lines', []),
                image_width=combined_boundary.shape[1],  # Image width from boundary shape
                error_margin=5
            )
            hex_info['constraint_solution'] = constraint_solution
        
        # Measure actual span distances from the boundary
        span_measurements = self._measure_boundary_spans(combined_boundary, vertical_lines)
        hex_info.update(span_measurements)
        
        # Use geometric constraint solver to find best grid parameters
        grid_solution = self._solve_hex_constraints(
            span_measurements['max_horizontal_span'],
            boundaries['width'],
            boundaries['height'],
            projections,
            span_measurements,
            combined_boundary,
            vertical_lines
        )
        hex_info.update(grid_solution)
        
        return hex_info
    
    def _detect_vertical_lines(self, combined_boundary: np.ndarray, projections: Dict = None, method: str = 'column_based') -> Dict:
        """Detect purely vertical lines using vertical-only projections when available
        
        Args:
            method: 'column_based' or 'hough_lines'
        """
        if combined_boundary.size == 0:
            return {'vertical_line_positions': [], 'leftmost_vertical': None, 'rightmost_vertical': None}
        
        if method == 'column_based':
            return self._detect_vertical_lines_column_based(combined_boundary, projections)
        else:
            return self._detect_vertical_lines_hough(combined_boundary, projections)
    
    def _detect_vertical_lines_column_based(self, combined_boundary: np.ndarray, projections: Dict = None) -> Dict:
        """Detect vertical lines by analyzing column-wise pixel presence"""
        if not projections or 'view_from_left_vertical' not in projections or 'view_from_right_vertical' not in projections:
            return {'vertical_line_positions': [], 'leftmost_vertical': None, 'rightmost_vertical': None}
        
        height, width = combined_boundary.shape
        left_vertical = projections['view_from_left_vertical']
        right_vertical = projections['view_from_right_vertical']
        
        # For each column, check if there are sufficient vertical pixels
        vertical_x_positions = []
        min_threshold = height * 0.1  # At least 10% of height should have vertical pixels
        
        for col in range(width):
            left_pixels = np.sum(left_vertical[:, col] > 0)
            right_pixels = np.sum(right_vertical[:, col] > 0)
            
            # If either left or right projection has a strong vertical presence
            if left_pixels > min_threshold or right_pixels > min_threshold:
                vertical_x_positions.append(col)
        
        # Clean up nearby positions (merge columns that are very close)
        cleaned_positions = []
        for pos in vertical_x_positions:
            if not cleaned_positions or abs(pos - cleaned_positions[-1]) > 5:
                cleaned_positions.append(pos)
        
        # Create helper functions for Y->X coordinate queries
        def get_x_at_y_left(y: int) -> int:
            """Get X coordinate of vertical line at given Y coordinate on left edge"""
            if y < 0 or y >= height:
                return None
            row_pixels = np.where(left_vertical[y, :] > 0)[0]
            return row_pixels[0] if len(row_pixels) > 0 else None
        
        def get_x_at_y_right(y: int) -> int:
            """Get X coordinate of vertical line at given Y coordinate on right edge"""
            if y < 0 or y >= height:
                return None
            row_pixels = np.where(right_vertical[y, :] > 0)[0]
            return row_pixels[-1] if len(row_pixels) > 0 else None
        
        result = {
            'vertical_line_positions': cleaned_positions,
            'leftmost_vertical': cleaned_positions[0] if cleaned_positions else None,
            'rightmost_vertical': cleaned_positions[-1] if cleaned_positions else None,
            'num_vertical_lines': len(cleaned_positions),
            'detection_method': 'column_based',
            'get_x_at_y_left': get_x_at_y_left,
            'get_x_at_y_right': get_x_at_y_right
        }
        
        if self.debug_mode:
            # Create a combined image for visualization
            combined_vertical = cv2.bitwise_or(left_vertical, right_vertical)
            self._save_vertical_lines_debug(combined_vertical, cleaned_positions)
            print(f"Column-based detection: {len(cleaned_positions)} vertical lines at positions: {cleaned_positions}")
        
        return result
    
    def _extract_vertical_line_segments(self, projections: Dict, x_tolerance: int = 3, gap_tolerance: int = 5) -> Dict:
        """Extract vertical line segments from left and right projections
        
        Args:
            projections: Dictionary containing vertical projections
            x_tolerance: Pixel tolerance for considering X positions as "same line"
            gap_tolerance: Number of consecutive missing pixels to tolerate as gaps
            
        Returns:
            Dictionary with left_lines and right_lines arrays
        """
        if not projections or 'view_from_left_vertical' not in projections or 'view_from_right_vertical' not in projections:
            return {'left_lines': [], 'right_lines': []}
        
        height, width = projections['view_from_left_vertical'].shape
        left_vertical = projections['view_from_left_vertical']
        right_vertical = projections['view_from_right_vertical']
        
        # Extract left vertical lines
        left_lines = self._extract_lines_from_side(left_vertical, 'left', x_tolerance, gap_tolerance)
        
        # Extract right vertical lines  
        right_lines = self._extract_lines_from_side(right_vertical, 'right', x_tolerance, gap_tolerance)
        
        # Find unique matching pairs
        matched_pairs = self._find_matching_left_right_segments(left_lines, right_lines)
        
        if self.debug_mode:
            print(f"Extracted {len(left_lines)} left vertical line segments")
            print(f"Extracted {len(right_lines)} right vertical line segments")
            self._save_line_segments_debug(left_vertical, right_vertical, left_lines, right_lines)
            self._save_matched_pairs_debug(left_vertical, right_vertical, left_lines, right_lines, matched_pairs)
        
        return {
            'left_lines': left_lines,
            'right_lines': right_lines,
            'matched_pairs': matched_pairs
        }
    
    def _extract_lines_from_side(self, projection: np.ndarray, side: str, x_tolerance: int, gap_tolerance: int) -> List[Dict]:
        """Extract vertical line segments from one side projection
        
        Returns:
            List of line dictionaries with keys: start_x, start_y, end_y, length
        """
        height, width = projection.shape
        lines = []
        
        current_line = None
        gap_count = 0
        
        for y in range(height):
            # Get X coordinate for this Y position
            row_pixels = np.where(projection[y, :] > 0)[0]
            
            if len(row_pixels) > 0:
                # Choose leftmost or rightmost pixel based on side
                if side == 'left':
                    x_coord = row_pixels[0]  # Leftmost pixel
                else:  # right
                    x_coord = row_pixels[-1]  # Rightmost pixel
                
                if current_line is None:
                    # Start new line
                    current_line = {
                        'start_x': x_coord,
                        'start_y': y,
                        'end_y': y,
                        'length': 1
                    }
                    gap_count = 0
                else:
                    # Check if we're continuing the same line
                    x_diff = abs(x_coord - current_line['start_x'])
                    
                    if x_diff <= x_tolerance:
                        # Continue current line
                        current_line['end_y'] = y
                        current_line['length'] = current_line['end_y'] - current_line['start_y'] + 1
                        gap_count = 0
                    else:
                        # X position changed significantly - finish current line and start new one
                        if current_line['length'] >= MIN_LINE_LENGTH:  
                            lines.append(current_line.copy())
                        
                        current_line = {
                            'start_x': x_coord,
                            'start_y': y,
                            'end_y': y,
                            'length': 1
                        }
                        gap_count = 0
            else:
                # No pixel found at this Y position
                if current_line is not None:
                    gap_count += 1
                    
                    if gap_count <= gap_tolerance:
                        # Tolerate small gaps - continue current line
                        current_line['end_y'] = y
                        current_line['length'] = current_line['end_y'] - current_line['start_y'] + 1
                    else:
                        # Gap too large - finish current line
                        if current_line['length'] >= MIN_LINE_LENGTH:  # Only keep lines with minimum length
                            lines.append(current_line.copy())
                        current_line = None
                        gap_count = 0
        
        # Don't forget the last line if it exists
        if current_line is not None and current_line['length'] >= MIN_LINE_LENGTH:
            lines.append(current_line)
        
        return lines
    
    def _calculate_spans_from_line_segments(self, left_lines: List[Dict], right_lines: List[Dict]) -> List[Dict]:
        """Calculate spans by finding overlapping left and right line segments
        
        Returns:
            List of span dictionaries with keys: left_x, right_x, span, start_y, end_y, overlap_length
        """
        spans = []
        
        # Sort both arrays by start_y
        left_lines_sorted = sorted(left_lines, key=lambda x: x['start_y'])
        right_lines_sorted = sorted(right_lines, key=lambda x: x['start_y'])
        
        # Find overlapping segments
        for left_line in left_lines_sorted:
            for right_line in right_lines_sorted:
                # Check if there's vertical overlap
                overlap_start = max(left_line['start_y'], right_line['start_y'])
                overlap_end = min(left_line['end_y'], right_line['end_y'])
                
                if overlap_start <= overlap_end:
                    # There's overlap - calculate span
                    overlap_length = overlap_end - overlap_start + 1
                    span = right_line['start_x'] - left_line['start_x']
                    
                    # Only consider significant overlaps
                    if overlap_length >= 20:  # Minimum overlap length
                        spans.append({
                            'left_x': left_line['start_x'],
                            'right_x': right_line['start_x'],
                            'span': span,
                            'start_y': overlap_start,
                            'end_y': overlap_end,
                            'overlap_length': overlap_length,
                            'left_line': left_line,
                            'right_line': right_line
                        })
        
        # Sort spans by overlap length (longest first) to prioritize the most significant spans
        spans.sort(key=lambda x: x['overlap_length'], reverse=True)
        
        if self.debug_mode:
            print(f"Found {len(spans)} overlapping line segment pairs")
            for i, span in enumerate(spans[:5]):  # Show top 5 spans
                print(f"  Span {i+1}: {span['span']}px (Y:{span['start_y']}-{span['end_y']}, overlap:{span['overlap_length']}px)")
        
        return spans
    
    def _save_line_segments_debug(self, left_vertical: np.ndarray, right_vertical: np.ndarray, left_lines: List[Dict], right_lines: List[Dict]):
        """Save debug visualization of extracted line segments"""
        if not self.debug_mode:
            return
        
        height, width = left_vertical.shape
        
        # Create RGB debug image
        debug_img = np.zeros((height, width, 3), dtype=np.uint8)
        
        # Show original projections in grayscale
        debug_img[:, :, 0] = left_vertical  # Red channel for left
        debug_img[:, :, 2] = right_vertical  # Blue channel for right
        
        # Draw extracted line segments
        for i, line in enumerate(left_lines):
            color = (0, 255, 0)  # Green for left lines
            cv2.line(debug_img, (line['start_x'], line['start_y']), (line['start_x'], line['end_y']), color, 2)
            # Add text label
            cv2.putText(debug_img, f"L{i}", (line['start_x'] + 5, line['start_y'] + 15), 
                       cv2.FONT_HERSHEY_SIMPLEX, 0.4, color, 1)
        
        for i, line in enumerate(right_lines):
            color = (0, 255, 255)  # Yellow for right lines  
            cv2.line(debug_img, (line['start_x'], line['start_y']), (line['start_x'], line['end_y']), color, 2)
            # Add text label
            cv2.putText(debug_img, f"R{i}", (line['start_x'] - 25, line['start_y'] + 15), 
                       cv2.FONT_HERSHEY_SIMPLEX, 0.4, color, 1)
        
        cv2.imwrite(str(self.debug_dir / "vertical_line_segments.png"), debug_img)
    
    def _find_matching_left_right_segments(self, left_lines: List[Dict], right_lines: List[Dict]) -> List[tuple]:
        """Find unique 1:1 matching between left and right line segments
        
        Returns:
            List of tuples (left_index, right_index) representing matched pairs
        """
        # First, find all potential matches (overlapping segments)
        potential_matches = []
        
        for left_idx, left_line in enumerate(left_lines):
            for right_idx, right_line in enumerate(right_lines):
                # Check if there's vertical overlap
                overlap_start = max(left_line['start_y'], right_line['start_y'])
                overlap_end = min(left_line['end_y'], right_line['end_y'])
                
                if overlap_start <= overlap_end:
                    overlap_length = overlap_end - overlap_start + 1
                    
                    # Only consider significant overlaps
                    if overlap_length >= 20:  # Minimum overlap length
                        potential_matches.append({
                            'left_idx': left_idx,
                            'right_idx': right_idx,
                            'overlap_length': overlap_length,
                            'left_length': left_line['length'],
                            'right_length': right_line['length']
                        })
        
        # Sort by overlap length (descending) to prioritize better matches
        potential_matches.sort(key=lambda x: x['overlap_length'], reverse=True)
        
        # Perform greedy matching - assign best matches first
        used_left = set()
        used_right = set()
        matched_pairs = []
        
        for match in potential_matches:
            left_idx = match['left_idx']
            right_idx = match['right_idx']
            
            # If both segments are still available, match them
            if left_idx not in used_left and right_idx not in used_right:
                matched_pairs.append((left_idx, right_idx))
                used_left.add(left_idx)
                used_right.add(right_idx)
                
                if self.debug_mode:
                    print(f"Matched L{left_idx} with R{right_idx} (overlap: {match['overlap_length']}px)")
        
        if self.debug_mode:
            print(f"Found {len(matched_pairs)} unique left-right segment pairs")
            print(f"Unmatched left segments: {set(range(len(left_lines))) - used_left}")
            print(f"Unmatched right segments: {set(range(len(right_lines))) - used_right}")
        
        return matched_pairs
    
    def _extract_all_vertical_segments(self, projections: Dict, x_tolerance: int = 3, gap_tolerance: int = 5) -> List[Dict]:
        """Extract all vertical line segments from vertical edge mask containing ALL vertical lines
        
        Args:
            projections: Dictionary containing vertical projections and vertical_edge_mask
            x_tolerance: Pixel tolerance for considering X positions as "same line"
            gap_tolerance: Number of consecutive missing pixels to tolerate as gaps
            
        Returns:
            List of all vertical line segment dictionaries
        """
        if not projections or 'vertical_edge_mask' not in projections:
            print("Warning: vertical_edge_mask not found, falling back to left/right boundary projections")
            if 'view_from_left_vertical' not in projections or 'view_from_right_vertical' not in projections:
                return []
            # Fallback to old method
            height, width = projections['view_from_left_vertical'].shape
            unified_vertical = np.zeros((height, width), dtype=np.uint8)
            unified_vertical = cv2.bitwise_or(unified_vertical, projections['view_from_left_vertical'])
            unified_vertical = cv2.bitwise_or(unified_vertical, projections['view_from_right_vertical'])
        else:
            # Use the full vertical edge mask that contains ALL vertical lines, not just boundaries
            unified_vertical = projections['vertical_edge_mask']
            height, width = unified_vertical.shape
        
        # Extract all vertical segments from the unified projection
        all_segments = []
        
        for col in range(width):
            # Get vertical segments in this column
            column_segments = self._extract_vertical_segments_from_column(unified_vertical[:, col], col, gap_tolerance)
            all_segments.extend(column_segments)
        
        # Filter segments by minimum length and merge nearby segments
        filtered_segments = []
        for segment in all_segments:
            if segment['length'] >= MIN_LINE_LENGTH:
                # Check if this segment should be merged with an existing one
                merged = False
                for existing in filtered_segments:
                    # If segments are at similar X positions and close in Y, merge them
                    if (abs(segment['start_x'] - existing['start_x']) <= x_tolerance and
                        abs(segment['start_y'] - existing['end_y']) <= gap_tolerance):
                        # Merge segments
                        existing['end_y'] = max(existing['end_y'], segment['end_y'])
                        existing['length'] = existing['end_y'] - existing['start_y'] + 1
                        merged = True
                        break
                
                if not merged:
                    filtered_segments.append(segment)
        
        # Now group segments that are part of the same thick vertical line
        # Group segments with X positions within a small tolerance (5 pixels)
        logical_segments = []
        used_segments = set()
        
        for i, segment in enumerate(filtered_segments):
            if i in used_segments:
                continue
                
            # Find all segments that belong to the same thick line
            line_group = [segment]
            used_segments.add(i)
            
            for j, other_segment in enumerate(filtered_segments):
                if j in used_segments or j == i:
                    continue
                    
                # Check if segments are part of the same thick line (close X positions and overlapping Y)
                x_diff = abs(segment['start_x'] - other_segment['start_x'])
                y_overlap = max(0, min(segment['end_y'], other_segment['end_y']) - 
                              max(segment['start_y'], other_segment['start_y']) + 1)
                
                if x_diff <= 5 and y_overlap > 10:  # Within 5px horizontally and overlap vertically
                    line_group.append(other_segment)
                    used_segments.add(j)
            
            # Create a single logical segment representing this thick line
            min_x = min(seg['start_x'] for seg in line_group)
            max_x = max(seg['start_x'] for seg in line_group)
            min_y = min(seg['start_y'] for seg in line_group) 
            max_y = max(seg['end_y'] for seg in line_group)
            
            logical_segment = {
                'start_x': (min_x + max_x) // 2,  # Center X of the thick line
                'start_y': min_y,
                'end_y': max_y,
                'length': max_y - min_y + 1,
                'thickness': max_x - min_x + 1,
                'component_count': len(line_group)  # How many thin segments make up this thick line
            }
            logical_segments.append(logical_segment)
        
        filtered_segments = logical_segments
        
        if self.debug_mode:
            print(f"Extracted {len(filtered_segments)} vertical segments from unified projection")
            self._save_all_vertical_segments_debug(unified_vertical, filtered_segments)
        
        return filtered_segments
    
    def _extract_vertical_segments_from_column(self, column_data: np.ndarray, col_x: int, gap_tolerance: int) -> List[Dict]:
        """Extract vertical line segments from a single column"""
        segments = []
        current_segment = None
        gap_count = 0
        
        for y, pixel_value in enumerate(column_data):
            if pixel_value > 0:
                if current_segment is None:
                    # Start new segment
                    current_segment = {
                        'start_x': col_x,
                        'start_y': y,
                        'end_y': y,
                        'length': 1
                    }
                    gap_count = 0
                else:
                    # Continue current segment
                    current_segment['end_y'] = y
                    current_segment['length'] = current_segment['end_y'] - current_segment['start_y'] + 1
                    gap_count = 0
            else:
                # No pixel found
                if current_segment is not None:
                    gap_count += 1
                    
                    if gap_count <= gap_tolerance:
                        # Tolerate small gaps - continue current segment
                        current_segment['end_y'] = y
                        current_segment['length'] = current_segment['end_y'] - current_segment['start_y'] + 1
                    else:
                        # Gap too large - finish current segment
                        if current_segment['length'] >= MIN_LINE_LENGTH:
                            segments.append(current_segment.copy())
                        current_segment = None
                        gap_count = 0
        
        # Don't forget the last segment if it exists
        if current_segment is not None and current_segment['length'] >= MIN_LINE_LENGTH:
            segments.append(current_segment)
        
        return segments
    
    def _get_matching_segments_in_rows(self, all_segments: List[Dict], min_overlap: int = 20) -> List[Dict]:
        """Group vertical line segments that align horizontally in the same row
        
        Args:
            all_segments: List of all vertical line segment dictionaries
            min_overlap: Minimum overlap length to consider segments as being in same row
            
        Returns:
            List of row dictionaries, each containing all vertical segments in that row
        """
        # Sort segments by Y position to process row by row
        sorted_segments = sorted(all_segments, key=lambda x: x['start_y'])
        
        # Group segments that have overlapping Y ranges
        row_groups = []
        used_segments = set()
        
        for i, base_segment in enumerate(sorted_segments):
            if i in used_segments:
                continue
                
            # Start a new row group with this segment
            row_segments = [base_segment]
            used_segments.add(i)
            
            # Find all other segments that overlap with this Y range
            base_start_y = base_segment['start_y']
            base_end_y = base_segment['end_y']
            
            for j, other_segment in enumerate(sorted_segments):
                if j in used_segments or j == i:
                    continue
                
                # Check for Y overlap
                overlap_start = max(base_start_y, other_segment['start_y'])
                overlap_end = min(base_end_y, other_segment['end_y'])
                
                if overlap_start <= overlap_end:
                    overlap_length = overlap_end - overlap_start + 1
                    
                    if overlap_length >= min_overlap:
                        row_segments.append(other_segment)
                        used_segments.add(j)
                        
                        # Expand the Y range to include this segment
                        base_start_y = min(base_start_y, other_segment['start_y'])
                        base_end_y = max(base_end_y, other_segment['end_y'])
            
            # Sort segments in this row by X position (left to right)
            row_segments.sort(key=lambda x: x['start_x'])
            
            # Only keep rows with multiple segments (at least 3 for meaningful analysis)
            if len(row_segments) >= 3:
                row_groups.append({
                    'y_start': base_start_y,
                    'y_end': base_end_y,
                    'y_range': base_end_y - base_start_y + 1,
                    'segments': row_segments,
                    'segment_count': len(row_segments),
                    'leftmost_x': row_segments[0]['start_x'],
                    'rightmost_x': row_segments[-1]['start_x'],
                    'span': row_segments[-1]['start_x'] - row_segments[0]['start_x'],
                    'x_positions': [s['start_x'] for s in row_segments]
                })
        
        # Sort row groups by segment count (most segments first) and then by Y range
        row_groups.sort(key=lambda x: (x['segment_count'], x['y_range']), reverse=True)
        
        if self.debug_mode:
            print(f"Found {len(row_groups)} row groups with 3+ segments")
            for i, row in enumerate(row_groups[:5]):  # Show top 5
                print(f"  Row {i+1}: {row['segment_count']} segments, Y:{row['y_start']}-{row['y_end']}, "
                      f"span:{row['span']}px")
                print(f"    X-positions: {row['x_positions']}")
        
        return row_groups
    
    def _save_all_vertical_segments_debug(self, unified_vertical: np.ndarray, all_segments: List[Dict]):
        """Save debug visualization of all extracted vertical segments"""
        if not self.debug_mode:
            return
        
        height, width = unified_vertical.shape
        
        # Create RGB debug image
        debug_img = np.zeros((height, width, 3), dtype=np.uint8)
        
        # Show original unified projection in grayscale
        debug_img[:, :, 0] = unified_vertical // 2  # Dimmed red channel
        debug_img[:, :, 1] = unified_vertical // 2  # Dimmed green channel
        debug_img[:, :, 2] = unified_vertical // 2  # Dimmed blue channel
        
        # Draw all vertical segments with different colors
        colors = [
            (0, 255, 0),    # Green
            (255, 255, 0),  # Yellow  
            (255, 0, 255),  # Magenta
            (0, 255, 255),  # Cyan
            (255, 128, 0),  # Orange
            (128, 255, 0),  # Lime
            (255, 0, 128),  # Pink
            (0, 128, 255),  # Sky blue
            (128, 0, 255),  # Violet
            (255, 255, 128), # Light yellow
        ]
        
        for i, segment in enumerate(all_segments):
            color = colors[i % len(colors)]
            
            # Draw vertical line segment
            cv2.line(debug_img, 
                    (segment['start_x'], segment['start_y']), 
                    (segment['start_x'], segment['end_y']), 
                    color, 2)
            
            # Add segment index label
            cv2.putText(debug_img, f"S{i}", 
                       (segment['start_x'] + 3, segment['start_y'] + 15), 
                       cv2.FONT_HERSHEY_SIMPLEX, 0.3, color, 1)
        
        cv2.imwrite(str(self.debug_dir / "all_vertical_segments.png"), debug_img)
    
    def _save_row_groups_debug(self, unified_vertical: np.ndarray, row_groups: List[Dict]):
        """Save debug visualization showing row groups with all segments"""
        if not self.debug_mode:
            return
        
        height, width = unified_vertical.shape
        
        # Create RGB debug image
        debug_img = np.zeros((height, width, 3), dtype=np.uint8)
        
        # Show original unified projection in very dim grayscale
        debug_img[:, :, 0] = unified_vertical // 4
        debug_img[:, :, 1] = unified_vertical // 4 
        debug_img[:, :, 2] = unified_vertical // 4
        
        # Use different colors for each row group
        row_colors = [
            (0, 255, 0),    # Green
            (255, 255, 0),  # Yellow  
            (255, 0, 255),  # Magenta
            (0, 255, 255),  # Cyan
            (255, 128, 0),  # Orange
            (128, 255, 0),  # Lime
            (255, 0, 128),  # Pink
            (0, 128, 255),  # Sky blue
            (128, 0, 255),  # Violet
            (255, 255, 128), # Light yellow
        ]
        
        for row_idx, row in enumerate(row_groups):
            color = row_colors[row_idx % len(row_colors)]
            
            # Draw all segments in this row with the same color
            for seg_idx, segment in enumerate(row['segments']):
                # Draw vertical line segment
                cv2.line(debug_img, 
                        (segment['start_x'], segment['start_y']), 
                        (segment['start_x'], segment['end_y']), 
                        color, 3)
                
                # Add segment index within row
                cv2.putText(debug_img, f"R{row_idx}S{seg_idx}", 
                           (segment['start_x'] + 3, segment['start_y'] + 15), 
                           cv2.FONT_HERSHEY_SIMPLEX, 0.3, color, 1)
            
            # Draw connecting line between leftmost and rightmost
            mid_y = (row['y_start'] + row['y_end']) // 2
            cv2.line(debug_img, 
                    (row['leftmost_x'], mid_y), 
                    (row['rightmost_x'], mid_y), 
                    color, 1)
            
            # Add row summary
            cv2.putText(debug_img, f"Row{row_idx}: {row['segment_count']} segs, {row['span']}px", 
                       (10, 20 + row_idx * 20), 
                       cv2.FONT_HERSHEY_SIMPLEX, 0.5, color, 1)
        
        cv2.imwrite(str(self.debug_dir / "row_groups_all_segments.png"), debug_img)
    
    def _save_matched_pairs_debug(self, left_vertical: np.ndarray, right_vertical: np.ndarray, 
                                 left_lines: List[Dict], right_lines: List[Dict], 
                                 matched_pairs: List[tuple]):
        """Save debug visualization showing only the matched left-right pairs"""
        if not self.debug_mode:
            return
        
        height, width = left_vertical.shape
        
        # Create RGB debug image
        debug_img = np.zeros((height, width, 3), dtype=np.uint8)
        
        # Show original projections in grayscale (dimmed)
        debug_img[:, :, 0] = left_vertical // 3  # Dimmed red channel for left
        debug_img[:, :, 2] = right_vertical // 3  # Dimmed blue channel for right
        
        # Draw only the matched pairs with connecting lines
        colors = [
            (0, 255, 0),    # Green
            (255, 255, 0),  # Yellow  
            (255, 0, 255),  # Magenta
            (0, 255, 255),  # Cyan
            (255, 128, 0),  # Orange
            (128, 255, 0),  # Lime
            (255, 0, 128),  # Pink
            (0, 128, 255),  # Sky blue
            (128, 0, 255),  # Violet
            (255, 255, 128), # Light yellow
        ]
        
        for i, (left_idx, right_idx) in enumerate(matched_pairs):
            left_line = left_lines[left_idx]
            right_line = right_lines[right_idx]
            
            # Use cycling colors for different pairs
            color = colors[i % len(colors)]
            
            # Draw left line segment
            cv2.line(debug_img, 
                    (left_line['start_x'], left_line['start_y']), 
                    (left_line['start_x'], left_line['end_y']), 
                    color, 3)
            
            # Draw right line segment
            cv2.line(debug_img, 
                    (right_line['start_x'], right_line['start_y']), 
                    (right_line['start_x'], right_line['end_y']), 
                    color, 3)
            
            # Draw connecting line between the midpoints
            left_mid_y = (left_line['start_y'] + left_line['end_y']) // 2
            right_mid_y = (right_line['start_y'] + right_line['end_y']) // 2
            cv2.line(debug_img, 
                    (left_line['start_x'], left_mid_y), 
                    (right_line['start_x'], right_mid_y), 
                    color, 1)
            
            # Add labels
            cv2.putText(debug_img, f"L{left_idx}", 
                       (left_line['start_x'] + 5, left_line['start_y'] + 15), 
                       cv2.FONT_HERSHEY_SIMPLEX, 0.4, color, 1)
            cv2.putText(debug_img, f"R{right_idx}", 
                       (right_line['start_x'] - 25, right_line['start_y'] + 15), 
                       cv2.FONT_HERSHEY_SIMPLEX, 0.4, color, 1)
            
            # Add span measurement
            span = right_line['start_x'] - left_line['start_x']
            mid_x = (left_line['start_x'] + right_line['start_x']) // 2
            cv2.putText(debug_img, f"{span}px", 
                       (mid_x - 20, left_mid_y - 10), 
                       cv2.FONT_HERSHEY_SIMPLEX, 0.3, color, 1)
        
        cv2.imwrite(str(self.debug_dir / "matched_pairs_only.png"), debug_img)
    
    def _calculate_spans_from_matched_pairs(self, left_lines: List[Dict], right_lines: List[Dict], 
                                          matched_pairs: List[tuple]) -> List[Dict]:
        """Calculate spans from unique matched left-right pairs
        
        Returns:
            List of span dictionaries with keys: left_x, right_x, span, start_y, end_y, overlap_length
        """
        spans = []
        
        for left_idx, right_idx in matched_pairs:
            left_line = left_lines[left_idx]
            right_line = right_lines[right_idx]
            
            # Calculate overlap
            overlap_start = max(left_line['start_y'], right_line['start_y'])
            overlap_end = min(left_line['end_y'], right_line['end_y'])
            overlap_length = overlap_end - overlap_start + 1
            
            # Calculate span
            span = right_line['start_x'] - left_line['start_x']
            
            spans.append({
                'left_x': left_line['start_x'],
                'right_x': right_line['start_x'],
                'span': span,
                'start_y': overlap_start,
                'end_y': overlap_end,
                'overlap_length': overlap_length,
                'left_line_idx': left_idx,
                'right_line_idx': right_idx,
                'left_line': left_line,
                'right_line': right_line
            })
        
        # Sort spans by overlap length (longest first) to prioritize the most significant spans
        spans.sort(key=lambda x: x['overlap_length'], reverse=True)
        
        if self.debug_mode:
            print(f"Calculated {len(spans)} spans from matched pairs")
            for i, span in enumerate(spans):
                print(f"  Span {i+1}: {span['span']}px (L{span['left_line_idx']}-R{span['right_line_idx']}, Y:{span['start_y']}-{span['end_y']}, overlap:{span['overlap_length']}px)")
        
        return spans
    
    def _solve_constraints_from_matched_pairs(self, matched_pairs: List[tuple], left_lines: List[Dict], 
                                            right_lines: List[Dict], image_width: int, 
                                            error_margin: int = 3) -> Dict:
        """Advanced constraint solver using matched pairs of line segments
        
        Args:
            matched_pairs: List of (left_idx, right_idx) tuples
            left_lines: List of left line segment dictionaries
            right_lines: List of right line segment dictionaries  
            image_width: Total width of the image
            error_margin: Pixel error margin for coordinate measurements
            
        Returns:
            Dictionary with candidate column counts and supporting evidence
        """
        from collections import defaultdict
        
        cols_matched = defaultdict(set)  # which matched spans support which ncols
        needs_extra_half_tile_with_ncols = {}
        
        if self.debug_mode:
            print(f"Testing constraint solver with image_width={image_width}, error_margin={error_margin}")
        
        for numcols in range(5, 100):  # Test various column counts
            col_matched = False
            for left_idx, right_idx in matched_pairs:
                left_line = left_lines[left_idx]
                right_line = right_lines[right_idx]
                left_startx = left_line['start_x'] - error_margin
                left_endx = left_line['start_x'] + error_margin
                right_startx = right_line['start_x'] - error_margin
                right_endx = right_line['start_x'] + error_margin
                
                # Test with error margins
                for lx in range(left_startx, left_endx + 1):
                    if col_matched: break
                    for rx in range(right_startx, right_endx + 1):
                        matched, needs_half_col, hex_width = constraint_matched_for_pair(numcols, lx, rx, image_width)
                        if matched:
                            cols_matched[numcols].add((left_idx, right_idx, hex_width))
                            needs_extra_half_tile_with_ncols[numcols] = needs_half_col
                            col_matched = True
                            break
        
        # set_trace(context=21)
        # Process results to find best candidates
        column_candidates = []
        
        for numcols in sorted(cols_matched.keys()):
            supporting_pairs = []
            hex_widths = []
            for (leftidx, rightidx, hex_width) in cols_matched[numcols]:
                supporting_pairs.append((leftidx, rightidx))
                hex_widths.append(hex_width)
            
            # Calculate statistics for this column count
            avg_hex_width = sum(hex_widths) / len(hex_widths) if hex_widths else 0
            has_half_tile = needs_extra_half_tile_with_ncols.get(numcols, False)
            
            candidate = {
                'cols': numcols,
                'supporting_pairs': list(supporting_pairs),
                'num_supporting_pairs': len(supporting_pairs),
                'hex_width_samples': hex_widths,
                'avg_hex_width': avg_hex_width,
                'has_half_tile': has_half_tile,
                'confidence': len(supporting_pairs)  # More supporting pairs = higher confidence
            }
            
            column_candidates.append(candidate)
        
        # Sort by confidence (number of supporting pairs) and then by column count
        column_candidates.sort(key=lambda x: (x['confidence'], -x['cols']), reverse=True)
        
        if self.debug_mode:
            print(f"Found {len(column_candidates)} column candidates")
            for i, candidate in enumerate(column_candidates[:5]):  # Show top 5
                print(f"  Candidate {i+1}: {candidate['cols']} cols, {candidate['num_supporting_pairs']} supporting pairs, "
                      f"avg hex_width={candidate['avg_hex_width']:.1f}, half_tile={candidate['has_half_tile']}")
        
        return {
            'column_candidates': column_candidates,
            'top_candidate': column_candidates[0] if column_candidates else None,
            'total_candidates': len(column_candidates)
        }
    
    def _detect_vertical_lines_hough(self, combined_boundary: np.ndarray, projections: Dict = None) -> Dict:
        """Detect vertical lines using Hough Line Transform (original method)"""
        if combined_boundary.size == 0:
            return {'vertical_line_positions': [], 'leftmost_vertical': None, 'rightmost_vertical': None}
        
        # If we have vertical-only projections, use them for better detection
        if projections and 'view_from_left_vertical' in projections and 'view_from_right_vertical' in projections:
            # Create a combined vertical-only boundary for line detection
            vertical_boundary = np.zeros_like(combined_boundary)
            vertical_boundary = cv2.bitwise_or(vertical_boundary, projections['view_from_left_vertical'])
            vertical_boundary = cv2.bitwise_or(vertical_boundary, projections['view_from_right_vertical'])
            
            if self.debug_mode:
                cv2.imwrite(str(self.debug_dir / "vertical_boundary_for_detection.png"), vertical_boundary)
                print(f"Using vertical-only boundary for Hough line detection")
            
            detection_image = vertical_boundary
        else:
            # Fallback to original combined boundary
            detection_image = combined_boundary
        
        # Apply Hough Line Transform to detect lines
        lines = cv2.HoughLinesP(detection_image, 1, np.pi/180, threshold=30, minLineLength=50, maxLineGap=15)
        
        vertical_x_positions = []
        
        if lines is not None:
            for line in lines:
                x1, y1, x2, y2 = line[0]
                
                # Calculate angle of the line
                if x2 - x1 != 0:
                    angle = np.arctan2(y2 - y1, x2 - x1) * 180 / np.pi
                else:
                    angle = 90  # Perfectly vertical
                
                # Filter for vertical lines (±10 degrees from vertical for some flexibility)
                if abs(angle - 90) <= 10 or abs(angle + 90) <= 10:
                    # Use average X position of the line
                    avg_x = (x1 + x2) / 2
                    vertical_x_positions.append(avg_x)
        
        # Remove duplicates and sort
        vertical_x_positions = sorted(list(set([int(x) for x in vertical_x_positions])))
        
        # Clean up nearby positions (merge lines that are very close)
        cleaned_positions = []
        for pos in vertical_x_positions:
            if not cleaned_positions or abs(pos - cleaned_positions[-1]) > 15:
                cleaned_positions.append(pos)
        
        result = {
            'vertical_line_positions': cleaned_positions,
            'leftmost_vertical': cleaned_positions[0] if cleaned_positions else None,
            'rightmost_vertical': cleaned_positions[-1] if cleaned_positions else None,
            'num_vertical_lines': len(cleaned_positions),
            'detection_method': 'hough_lines'
        }
        
        if self.debug_mode:
            self._save_vertical_lines_debug(detection_image, cleaned_positions)
            print(f"Hough-based detection: {len(cleaned_positions)} vertical lines at positions: {cleaned_positions}")
        
        return result
    
    def _measure_boundary_spans(self, boundary_img: np.ndarray, vertical_lines: Dict = None) -> Dict:
        """Measure actual spans from boundary image, optionally using detected vertical lines"""
        height, width = boundary_img.shape
        measurements = {}
        
        # Find boundary pixels
        boundary_coords = np.where(boundary_img > 0)
        if len(boundary_coords[0]) == 0:
            return {'max_horizontal_span': 0, 'max_vertical_span': 0}
        
        # If we have vertical line information, use it for more accurate measurements
        if vertical_lines and vertical_lines.get('leftmost_vertical') is not None and vertical_lines.get('rightmost_vertical') is not None:
            # Use true vertical line boundaries
            leftmost = vertical_lines['leftmost_vertical']
            rightmost = vertical_lines['rightmost_vertical']
            max_horizontal_span = rightmost - leftmost
            
            if self.debug_mode:
                print(f"Using vertical line boundaries: {leftmost} to {rightmost} = {max_horizontal_span}px span")
            
            measurements['max_horizontal_span'] = max_horizontal_span
            measurements['leftmost_boundary'] = leftmost
            measurements['rightmost_boundary'] = rightmost
            measurements['boundary_source'] = 'vertical_lines'
        else:
            # Fallback to original row-by-row measurement
            horizontal_spans = []
            for row in range(height):
                row_pixels = np.where(boundary_img[row, :] > 0)[0]
                if len(row_pixels) >= 2:
                    span = np.max(row_pixels) - np.min(row_pixels)
                    horizontal_spans.append(span)
            
            measurements['max_horizontal_span'] = max(horizontal_spans) if horizontal_spans else 0
            measurements['boundary_source'] = 'pixel_spans'
        
        # Measure vertical spans across different columns  
        vertical_spans = []
        for col in range(width):
            col_pixels = np.where(boundary_img[:, col] > 0)[0]
            if len(col_pixels) >= 2:
                span = np.max(col_pixels) - np.min(col_pixels)
                vertical_spans.append(span)
        
        measurements['max_vertical_span'] = max(vertical_spans) if vertical_spans else 0
        measurements['avg_vertical_span'] = np.mean(vertical_spans) if vertical_spans else 0
        measurements['all_vertical_spans'] = vertical_spans[:10]     # Sample for debugging
        
        return measurements
    
    def _solve_hex_constraints(self, measured_span: int, total_width: int, total_height: int, projections: Dict[str, np.ndarray], span_measurements: Dict, combined_boundary: np.ndarray, vertical_lines: Dict) -> Dict:
        """Solve geometric constraints using detected vertical lines"""
        # Use detected vertical lines to calculate grid parameters directly
        vertical_positions = vertical_lines.get('vertical_line_positions', [])
        leftmost_vertical = vertical_lines.get('leftmost_vertical')
        rightmost_vertical = vertical_lines.get('rightmost_vertical')
        
        if len(vertical_positions) >= 2 and leftmost_vertical is not None and rightmost_vertical is not None:
            # Calculate hex_width from spacing between consecutive vertical lines
            spacings = []
            for i in range(1, len(vertical_positions)):
                spacing = vertical_positions[i] - vertical_positions[i-1]
                spacings.append(spacing)
            
            # Use average spacing as hex_width
            hex_width = int(np.mean(spacings)) if spacings else 60
            
            # Calculate columns from number of vertical lines
            cols = len(vertical_positions)
            
            # Calculate actual span from leftmost to rightmost vertical lines
            actual_span = rightmost_vertical - leftmost_vertical
            
            # Center spacing is same as hex_width for hexagonal grids
            center_spacing = hex_width
            
            if self.debug_mode:
                print(f"Direct calculation from vertical lines:")
                print(f"  Detected vertical lines: {len(vertical_positions)}")
                print(f"  Leftmost: {leftmost_vertical}, Rightmost: {rightmost_vertical}")
                print(f"  Spacings: {spacings}")
                print(f"  Calculated hex_width: {hex_width}")
                print(f"  Calculated cols: {cols}")
                print(f"  Actual span: {actual_span}")
            
            best_solution = {
                'cols': cols,
                'hex_width': hex_width,
                'center_spacing': center_spacing,
                'expected_span': actual_span,
                'expected_total_width': actual_span,
                'span_error': 0,  # No error since we're using actual measurements
                'width_error': 0,
                'total_error': 0,
                'vertical_line_spacings': spacings,
                'leftmost_boundary': leftmost_vertical,
                'rightmost_boundary': rightmost_vertical
            }
        else:
            # Fallback to original brute-force method if vertical line detection fails
            if self.debug_mode:
                print(f"Vertical line detection failed, falling back to brute-force method")
                print(f"  Detected lines: {len(vertical_positions)}")
            
            best_solution = self._brute_force_constraint_solving(measured_span, total_width, total_height)
        
        # Calculate rows by counting actual vertical segments from boundary data
        if best_solution:
            hex_height = int(best_solution['hex_width'] * HEIGHT_FACTOR)  # Square tiles
            
            # Count vertical segments directly from the boundary measurements
            rows = self._count_vertical_segments(combined_boundary, total_height)
            
            # Calculate vertical spacing based on detected row count
            if rows > 1:
                vertical_spacing = (total_height - hex_height) / (rows - 1)
            else:
                vertical_spacing = hex_height * 0.75  # Fallback
            
            if self.debug_mode:
                print(f"Detected {rows} vertical segments")
                print(f"Calculated vertical spacing: {vertical_spacing:.1f} pixels")
            
            best_solution['rows'] = rows
            best_solution['hex_height'] = hex_height
            best_solution['hex_side_length'] = best_solution['hex_width']  # For compatibility
            best_solution['vertical_spacing'] = vertical_spacing
        
        return best_solution or {'hex_side_length': 60, 'cols': DEFAULT_NUM_STARTING_COLS, 'rows': DEFAULT_NUM_STARTING_ROWS}
    
    def _brute_force_constraint_solving(self, measured_span: int, total_width: int, total_height: int) -> Dict:
        """Fallback brute-force constraint solving method"""
        best_solution = None
        best_error = float('inf')
        
        # Try different numbers of columns and hex sizes
        for cols in range(5, 100):  # Reasonable range for WeeWar maps
            for hex_width in range(30, 100):  # Reasonable hex size range
                
                # For hexagonal grids, horizontal center spacing is hex_width
                center_spacing = hex_width # hex_width *is* the center spacing
                
                # Calculate expected span for this configuration
                # Span from leftmost to rightmost hex centers would be (cols-1) * center_spacing
                expected_span = (cols - 1) * center_spacing
                
                # Calculate expected total width
                # Could be cols * center_spacing or cols * center_spacing + hex_width/2 (for offset)
                expected_total_1 = cols * center_spacing
                expected_total_2 = cols * center_spacing + hex_width/2
                
                # Check how well this matches our measurements
                span_error = abs(measured_span - expected_span)
                width_error_1 = abs(total_width - expected_total_1)
                width_error_2 = abs(total_width - expected_total_2)
                width_error = min(width_error_1, width_error_2)
                
                # Combined error metric
                total_error = span_error + width_error
                
                if total_error < best_error:
                    best_error = total_error
                    best_solution = {
                        'cols': cols,
                        'hex_width': hex_width,
                        'center_spacing': center_spacing,
                        'expected_span': expected_span,
                        'expected_total_width': expected_total_1 if width_error_1 < width_error_2 else expected_total_2,
                        'span_error': span_error,
                        'width_error': width_error,
                        'total_error': total_error
                    }
        
        return best_solution or {'hex_side_length': 60, 'cols': DEFAULT_NUM_STARTING_COLS, 'rows': DEFAULT_NUM_STARTING_ROWS}
    
    def _count_vertical_segments(self, combined_boundary: np.ndarray, total_height: int) -> int:
        """Count vertical line segments (like "|" pipes) in the boundary data"""
        
        if combined_boundary.size == 0:
            return 7  # Fallback
        
        height, width = combined_boundary.shape
        
        # Look for vertical line segments by analyzing columns
        # A vertical segment would show as a continuous vertical line in a column
        vertical_segments = []
        
        for col in range(width):
            column_data = combined_boundary[:, col]
            if np.any(column_data > 0):
                # Find continuous vertical segments in this column
                segments = self._find_continuous_segments(column_data)
                if segments:
                    vertical_segments.extend(segments)
        
        # Count distinct vertical segments
        # Group segments that are at similar X positions (same hex boundary)
        if len(vertical_segments) == 0:
            return 7  # Fallback
        
        # Filter out very short segments (noise)
        long_segments = [seg for seg in vertical_segments if seg['length'] > 20]
        
        # Group segments by their vertical position ranges
        unique_vertical_regions = self._group_vertical_segments(long_segments, height)
        
        row_count = len(unique_vertical_regions)
        
        if self.debug_mode:
            print(f"Vertical segment analysis:")
            print(f"  Total segments found: {len(vertical_segments)}")
            print(f"  Long segments (>20px): {len(long_segments)}")
            print(f"  Unique vertical regions: {row_count}")
        
        # Return reasonable row count
        if row_count > MAX_ROWS:
            raise f"Found too many rows: {row_count}"
        return max(5, row_count)
    
    def _find_continuous_segments(self, column_data: np.ndarray) -> List[Dict]:
        """Find continuous non-zero segments in a column"""
        segments = []
        start = None
        
        for i, val in enumerate(column_data):
            if val > 0 and start is None:
                start = i
            elif val == 0 and start is not None:
                segments.append({
                    'start': start,
                    'end': i - 1,
                    'length': i - start
                })
                start = None
        
        # Handle segment that extends to the end
        if start is not None:
            segments.append({
                'start': start,
                'end': len(column_data) - 1,
                'length': len(column_data) - start
            })
        
        return segments
    
    def _group_vertical_segments(self, segments: List[Dict], total_height: int) -> List[Dict]:
        """Group segments that represent distinct vertical bands/levels"""
        if not segments:
            return []
        
        # Instead of grouping overlapping segments, look for distinct Y-level bands
        # Each hex row should create segments at roughly the same Y-levels
        
        # Extract the center Y-position of each segment
        segment_centers = [(s['start'] + s['end']) / 2 for s in segments]
        
        if not segment_centers:
            return []
        
        # Sort centers and look for gaps that indicate different row levels
        sorted_centers = sorted(segment_centers)
        
        # Find significant gaps between segment centers
        gaps = []
        for i in range(1, len(sorted_centers)):
            gap = sorted_centers[i] - sorted_centers[i-1]
            gaps.append(gap)
        
        # A significant gap indicates a new row level
        # Use adaptive threshold based on total height
        min_row_spacing = total_height / 15  # Expect at least this much space between rows
        significant_gaps = [i for i, gap in enumerate(gaps) if gap > min_row_spacing]
        
        # Number of distinct levels = number of significant gaps + 1
        num_levels = len(significant_gaps) + 1
        
        if self.debug_mode:
            print(f"  Segment centers: {sorted_centers[:10]}...")  # Show first 10
            print(f"  Significant gaps (>{min_row_spacing:.1f}): {len(significant_gaps)}")
            print(f"  Calculated levels: {num_levels}")
        
        # Return mock groups (we just need the count)
        return [{'level': i} for i in range(num_levels)]
    
    def _count_hex_rows_from_edges(self, left_edge: np.ndarray, right_edge: np.ndarray, total_height: int) -> int:
        """Count the number of hex rows by analyzing vertical features in edge images"""
        
        # Create vertical profiles by summing horizontally across each edge image
        left_profile = np.sum(left_edge, axis=1) if left_edge.size > 0 else np.array([])
        right_profile = np.sum(right_edge, axis=1) if right_edge.size > 0 else np.array([])
        
        # Count peaks/features in both profiles
        left_rows = self._count_vertical_features(left_profile)
        right_rows = self._count_vertical_features(right_profile)
        
        # Use the profile that gives a more reasonable count
        detected_rows = max(left_rows, right_rows) if left_rows > 0 and right_rows > 0 else max(left_rows, right_rows, 5)
        
        if self.debug_mode:
            print(f"Row counting: left_profile detected {left_rows} rows, right_profile detected {right_rows} rows")
            print(f"Using {detected_rows} rows")
        
        return max(5, min(detected_rows, 12))  # Reasonable bounds
    
    def _count_vertical_features(self, profile: np.ndarray) -> int:
        """Count vertical features (step patterns) in a 1D profile representing hex row positions"""
        if len(profile) < 10 or np.max(profile) == 0:
            return 0
        
        from scipy.signal import find_peaks
        from scipy.ndimage import gaussian_filter1d
        
        # For hex step patterns, look for transitions rather than peaks
        # Light smoothing to preserve step structure
        smoothed = gaussian_filter1d(profile.astype(float), sigma=1.5)
        
        # Method 1: Look for significant steps/transitions in the profile
        diff_profile = np.diff(smoothed)
        
        # Find positions where the profile changes significantly (steps)
        threshold = np.std(diff_profile) * 1.5  # Adaptive threshold
        step_positions = np.where(np.abs(diff_profile) > threshold)[0]
        
        # Group nearby step positions (within ~30 pixels) as belonging to the same hex row
        if len(step_positions) > 0:
            grouped_steps = []
            current_group = [step_positions[0]]
            
            for pos in step_positions[1:]:
                if pos - current_group[-1] < 30:  # Same hex row
                    current_group.append(pos)
                else:  # New hex row
                    grouped_steps.append(current_group)
                    current_group = [pos]
            
            grouped_steps.append(current_group)
            num_step_groups = len(grouped_steps)
        else:
            num_step_groups = 0
        
        # Method 2: Try traditional peak detection with relaxed parameters
        max_val = np.max(smoothed)
        peaks, _ = find_peaks(smoothed, 
                             height=max_val * 0.1,  # Very low threshold
                             distance=15)           # Closer spacing allowed
        
        num_peaks = len(peaks)
        
        # Use the method that gives a more reasonable result
        detected_features = max(num_step_groups, num_peaks)
        
        if self.debug_mode:
            print(f"  Step groups: {num_step_groups}, Peaks: {num_peaks}, Using: {detected_features}")
        
        return detected_features
    
    def _find_pattern_spacing(self, projection: np.ndarray) -> int:
        """Find the repeating pattern spacing in a projection"""
        if len(projection) < 10:
            return 0
        
        # Skip if projection is all zeros
        if np.max(projection) == 0:
            return 0
        
        # Look for peaks and valleys in the projection to find pattern spacing
        from scipy.signal import find_peaks
        
        # For sparse edge data, don't smooth too much - preserve the edge positions
        from scipy.ndimage import gaussian_filter1d
        smoothed = gaussian_filter1d(projection.astype(float), sigma=1)
        
        # Lower the threshold and distance for sparse edge data
        max_val = np.max(smoothed)
        if max_val == 0:
            return 0
            
        # Find peaks with lower threshold for sparse data
        peaks, _ = find_peaks(smoothed, height=max_val * 0.1, distance=10)
        
        if len(peaks) < 2:
            # Try finding any non-zero positions as potential peaks
            nonzero_positions = np.where(projection > 0)[0]
            if len(nonzero_positions) >= 2:
                # Use the spacing between non-zero regions
                spacings = np.diff(nonzero_positions)
                # Filter out very small spacings (likely same feature)
                valid_spacings = spacings[spacings > 5]
                if len(valid_spacings) > 0:
                    return int(np.median(valid_spacings))
            return 0
        
        # Calculate spacing between peaks
        peak_spacings = np.diff(peaks)
        
        if len(peak_spacings) > 0:
            # Return median spacing (most common hex spacing)
            return int(np.median(peak_spacings))
        
        return 0
    
    def _find_longest_continuous_line(self, projection: np.ndarray) -> int:
        """Find the length of the longest continuous non-zero segment"""
        if len(projection) == 0:
            return 0
        
        max_length = 0
        current_length = 0
        
        for value in projection:
            if value > 0:
                current_length += 1
                max_length = max(max_length, current_length)
            else:
                current_length = 0
        
        return max_length
    
    def _calculate_grid_from_boundaries(self, image: np.ndarray, boundaries: Dict, expected_tiles: int) -> GridParams:
        """Calculate hex grid parameters using geometric constraint solution"""
        
        map_width = boundaries['width']
        map_height = boundaries['height']
        
        # Use the geometric solution from constraint solver
        solution = boundaries
        cols = solution.get('cols', DEFAULT_NUM_STARTING_COLS)
        rows = solution.get('rows', DEFAULT_NUM_STARTING_ROWS)
        hex_width = solution.get('hex_width', 60)
        hex_height = solution.get('hex_height', int(hex_width * HEIGHT_FACTOR))
        center_spacing = solution.get('center_spacing', hex_width)
        
        # Calculate spacing based on the geometric solution
        spacing_x = center_spacing
        spacing_y = solution.get('vertical_spacing', hex_height * 0.75)  # Use calculated spacing
        
        # Calculate starting positions (center of first hex)
        start_x = boundaries['left'] + hex_width // 2
        start_y = boundaries['top'] + hex_height // 2
        
        # Row offset for hex pattern (odd rows offset by half spacing)
        row_offset = spacing_x // 2
        
        print(f"Geometric solution: {cols} cols x {rows} rows = {cols * rows} positions")
        print(f"Hex dimensions: {hex_width}x{hex_height}")
        print(f"Center spacing: {spacing_x:.1f}x{spacing_y:.1f}")
        print(f"Solution error: {solution.get('total_error', 'unknown')}")
        
        return GridParams(
            hex_width=hex_width,
            hex_height=hex_height,
            rows=rows,
            cols=cols,
            row_offset=row_offset,
            start_x=start_x,
            start_y=start_y,
            spacing_x=spacing_x,
            spacing_y=spacing_y
        )
    
    def _fallback_grid_calculation(self, boundaries: Dict, expected_tiles: int) -> GridParams:
        """Fallback calculation when hex side length detection fails"""
        map_width = boundaries['width']
        map_height = boundaries['height']
        
        # Use square root approximation
        approx_side = int(np.sqrt(expected_tiles * 1.4))  # Slightly larger for hex shape
        
        rows = approx_side
        cols = approx_side
        
        spacing_x = map_width / cols
        spacing_y = map_height / rows
        
        hex_width = int(spacing_x)
        hex_height = int(spacing_y)
        
        start_x = boundaries['left'] + hex_width // 2
        start_y = boundaries['top'] + hex_height // 2
        row_offset = spacing_x // 2
        
        return GridParams(
            hex_width=hex_width,
            hex_height=hex_height,
            rows=rows,
            cols=cols,
            row_offset=row_offset,
            start_x=start_x,
            start_y=start_y,
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
    
    def _save_projection_debug_4dir(self, projections: Dict[str, np.ndarray], height: int, width: int):
        """Save debug visualization of 4-directional edge images"""
        # Save each individual edge image for clear visualization
        for direction, edge_img in projections.items():
            filename = f"edge_{direction}.png"
            cv2.imwrite(str(self.debug_dir / filename), edge_img)
        
        # Create a combined RGB visualization where each direction gets a color channel
        combined_img = np.zeros((height, width, 3), dtype=np.uint8)
        
        # Assign colors to each direction for the combined view
        if 'view_from_top' in projections:
            combined_img[:, :, 1] = projections['view_from_top']  # Green channel
        if 'view_from_bottom' in projections:
            combined_img[:, :, 0] = projections['view_from_bottom']  # Blue channel  
        if 'view_from_left' in projections:
            combined_img[:, :, 2] = projections['view_from_left']  # Red channel
        if 'view_from_right' in projections:
            # Combine with green channel (will appear cyan where overlapping)
            combined_img[:, :, 1] = cv2.bitwise_or(combined_img[:, :, 1], projections['view_from_right'])
        
        cv2.imwrite(str(self.debug_dir / "4dir_edges_combined.png"), combined_img)
        
        # Also create a simple grayscale combined view (OR of all edges)
        combined_gray = np.zeros((height, width), dtype=np.uint8)
        for direction, edge_img in projections.items():
            if not direction.endswith('_vertical'):  # Only use original projections for combined view
                combined_gray = cv2.bitwise_or(combined_gray, edge_img)
        
        cv2.imwrite(str(self.debug_dir / "4dir_edges_gray.png"), combined_gray)
        
        # Create separate combined view for vertical-only projections
        if 'view_from_left_vertical' in projections and 'view_from_right_vertical' in projections:
            combined_vertical = np.zeros((height, width), dtype=np.uint8)
            combined_vertical = cv2.bitwise_or(combined_vertical, projections['view_from_left_vertical'])
            combined_vertical = cv2.bitwise_or(combined_vertical, projections['view_from_right_vertical'])
            cv2.imwrite(str(self.debug_dir / "vertical_only_edges.png"), combined_vertical)
    
    def _save_vertical_lines_debug(self, combined_boundary: np.ndarray, vertical_positions: List[int]):
        """Save debug visualization of detected vertical lines"""
        if not self.debug_mode or not vertical_positions:
            return
        
        # Create RGB image for better visualization
        height, width = combined_boundary.shape
        debug_img = cv2.cvtColor(combined_boundary, cv2.COLOR_GRAY2BGR)
        
        # Draw detected vertical lines
        for i, x_pos in enumerate(vertical_positions):
            # Use different colors for leftmost (red), rightmost (blue), and middle (green)
            if i == 0:  # Leftmost
                color = (0, 0, 255)  # Red
                thickness = 2
            elif i == len(vertical_positions) - 1:  # Rightmost
                color = (255, 0, 0)  # Blue
                thickness = 2
            else:  # Middle lines
                color = (0, 255, 0)  # Green
                thickness = 1
            
            # Draw vertical line
            cv2.line(debug_img, (x_pos, 0), (x_pos, height), color, thickness)
            
            # Add text label with X-coordinate
            cv2.putText(debug_img, f"{x_pos}", (x_pos - 20, 30 + i * 20), cv2.FONT_HERSHEY_SIMPLEX, 0.5, color, 1)
        
        # Add summary text
        if len(vertical_positions) >= 2:
            span = vertical_positions[-1] - vertical_positions[0]
            spacings = [vertical_positions[i] - vertical_positions[i-1] for i in range(1, len(vertical_positions))]
            avg_spacing = int(np.mean(spacings)) if spacings else 0
            
            summary_text = [
                f"Lines: {len(vertical_positions)}",
                f"Span: {span}px",
                f"Avg spacing: {avg_spacing}px"
            ]
            
            for i, text in enumerate(summary_text):
                cv2.putText(debug_img, text, (10, height - 60 + i * 20), 
                           cv2.FONT_HERSHEY_SIMPLEX, 0.6, (255, 255, 255), 2)
        
        # Save debug image
        cv2.imwrite(str(self.debug_dir / "vertical_lines_detected.png"), debug_img)
    
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


def constraint_matched_for_pair(cols, lx, rx, image_width, min_width=40, max_width=80):
    span_width = rx - lx
    
    # Check if span is divisible by column count
    if span_width <= 0:
        return False, False, 0

    # Check if image width can be filled with full hexes
    hex_width = -1
    midhw = image_width // cols

    def close_enough(value, another, delta = 0.01):
        return abs(value - another) <= delta

    err = 3
    for hw in range (midhw - err, midhw + err + 1):
        if hw < min_width or hw > max_width: continue
        ncols_in_span = span_width / hw
        ncols_in_image = image_width / hw
        ncols_in_image2 = (2 * image_width) / hw

        # if cols == 14: print(locals()) set_trace(context=14)

        if not close_enough(ncols_in_span, int(ncols_in_span + 0.5)):
            continue
        if close_enough(ncols_in_image, int(ncols_in_image + 0.5)):
            # we have a candidate
            return True, True, hw
        if close_enough(2 * ncols_in_image, int(ncols_in_image2 + 0.5)):
            # we have a candidate with half a tile
            hex_width = hw
            return True, False, hw

    return False, False, -1

def main():
    """Analyze hex grid structure from command line or test with default image"""
    import argparse
    
    parser = argparse.ArgumentParser(description='Analyze hex grid structure in WeeWar map images')
    parser.add_argument('--image', type=str, help='Path to the map image to analyze')
    parser.add_argument('--expected-tiles', type=int, default=34, help='Expected number of tiles in the map')
    parser.add_argument('--debug', action='store_true', help='Enable debug mode with visualization')
    parser.add_argument('--detection-method', type=str, choices=['column_based', 'hough_lines'], default='column_based', help='Method for vertical line detection')
    
    args = parser.parse_args()
    
    # Use provided image path or default test image
    if args.image:
        image_path = args.image
    else:
        image_path = "../data/Maps/1_files/map-og.png"
        print(f"No image specified, using default: {image_path}")
    
    # Load image
    image = cv2.imread(image_path)
    
    if image is None:
        print(f"Could not load image: {image_path}")
        return
    
    print(f"Analyzing grid structure for: {image_path}")
    
    # Analyze grid with expected tile count
    analyzer = HexGridAnalyzer(debug_mode=args.debug)
    analyzer.detection_method = args.detection_method
    params = analyzer.analyze_grid_structure(image, expected_tiles=args.expected_tiles)
    
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
