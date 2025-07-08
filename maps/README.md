# WeeWar Map Extraction System

## Overview

This system reverse engineers WeeWar maps from preview images by detecting hexagonal grid structures and classifying tiles through computer vision techniques. The goal is to extract the tile layout from map preview images and convert them into structured data that can be used by the game engine.

## Key Features

- **Scale-independent detection** - Works with any image scale without hardcoded tile sizes
- **Multi-method approach** - Uses multiple detection strategies for robustness
- **Visual validation** - Renders extracted maps for easy comparison with originals
- **Comprehensive reporting** - Generates detailed validation reports with statistics

## Technical Strategy

### 1. Hexagonal Grid Detection

The system uses a multi-layered approach to detect the hexagonal grid structure:

#### Primary Method: Edge Detection + Contour Analysis
- Converts image to grayscale and applies Gaussian blur
- Uses Canny edge detection to find tile boundaries
- Finds contours and filters for hexagonal shapes (5-8 sides)
- Calculates hex centers from contour moments

#### Secondary Method: Template Matching
- Uses reference tile images as templates
- Performs multi-scale template matching (0.5x to 1.5x scale)
- Applies non-maximum suppression to remove duplicates
- Organizes detected centers into grid structure

#### Fallback Method: Color/Feature Analysis
- Analyzes dominant colors and texture patterns
- Uses K-means clustering to identify tile regions
- Matches clusters to known tile types

### 2. Tile Classification

Once the grid is detected, each hex region is classified:

#### Template Matching Classification
- Extracts hex regions based on detected grid
- Resizes reference tiles to match region size
- Uses normalized cross-correlation for matching
- Assigns confidence scores to each classification

#### Dominant Color Analysis
- Calculates dominant colors using K-means clustering
- Compares with pre-computed tile color signatures
- Provides additional confidence metric

### 3. Validation and Verification

The system validates results against known map data:

#### Tile Count Validation
- Compares extracted tile count with expected count from JSON
- Identifies significant discrepancies

#### Distribution Analysis
- Compares tile type distribution with expected values
- Calculates percentage differences for each tile type

#### Visual Validation
- Renders extracted map using reference tiles
- Generates side-by-side comparison with original
- Highlights mismatched or low-confidence tiles

## Implementation Details

### Core Classes

#### `MapExtractor`
Main class that orchestrates the extraction process:
- Loads tile references and map data
- Implements hex grid detection algorithms
- Performs tile classification
- Validates results against expected data

#### `TileInfo`
Represents a tile type with its reference image and properties:
- `id`: Unique tile identifier
- `name`: Human-readable tile name
- `image_path`: Path to reference tile image
- `reference_image`: OpenCV image array
- `dominant_color`: Pre-computed dominant color

#### `HexCell`
Represents a single hex cell in the grid:
- `row`, `col`: Grid position
- `center_x`, `center_y`: Pixel coordinates
- `tile_id`: Classified tile type
- `confidence`: Classification confidence score

### Key Algorithms

#### Non-Maximum Suppression
Removes duplicate detections that are too close together:
```python
def _non_max_suppression(self, matches, min_distance=40):
    # Sort by confidence and keep only distant matches
```

#### Grid Organization
Converts scattered hex centers into structured grid:
```python
def _organize_hex_centers(self, centers):
    # Group by y-coordinate (rows) then sort by x-coordinate (columns)
```

## Limitations

### Current Limitations

1. **Partial Occlusion**: Tiles partially covered by units or UI elements may be misclassified
2. **Similar Tiles**: Tiles with very similar visual appearance may be confused
3. **Image Quality**: Low-resolution or heavily compressed images may reduce accuracy
4. **Lighting Variations**: Different lighting conditions in source images may affect color-based classification
5. **Irregular Grids**: Maps with non-standard hex arrangements may not be detected properly

### Known Issues

1. **Edge Tiles**: Tiles at map edges may be harder to detect due to partial visibility
2. **Template Scaling**: Extreme scale differences may cause template matching to fail
3. **Color Accuracy**: Monitor calibration and image format may affect color matching
4. **Memory Usage**: Large maps or many reference tiles may consume significant memory

### Future Improvements

1. **Machine Learning**: Train a neural network for more robust tile classification
2. **Advanced Preprocessing**: Implement better image enhancement techniques
3. **Adaptive Thresholding**: Automatically adjust detection parameters per map
4. **Multi-Resolution Analysis**: Analyze maps at multiple resolutions for better accuracy

## Usage Instructions

### Basic Usage

1. **Install Dependencies**:
   ```bash
   pip install -r requirements.txt
   ```

2. **Run Extraction**:
   ```python
   from map_extractor import MapExtractor
   
   extractor = MapExtractor()
   grid = extractor.extract_map(map_id=1)
   ```

3. **Generate Validation Report**:
   ```python
   extractor.generate_validation_report(map_id=1)
   ```

### Command Line Usage

```bash
# Extract a specific map
python map_extractor.py --map-id 1

# Extract all maps
python map_extractor.py --all

# Generate validation report
python map_extractor.py --validate --map-id 1
```

## Adding New Maps

### Step 1: Add Map Data
1. Update `weewar-maps.json` with new map information:
   ```json
   {
     "id": 99,
     "name": "New Map",
     "imageURL": "./99_files/map-og.png",
     "tileCount": 50,
     "tiles": {
       "Grass": 20,
       "Water": 15,
       "Mountains": 10,
       "Forest": 5
     }
   }
   ```

### Step 2: Add Map Image
1. Create directory: `data/Maps/99_files/`
2. Place map preview image: `data/Maps/99_files/map-og.png`

### Step 3: Extract and Validate
1. Run extraction:
   ```python
   extractor = MapExtractor()
   grid = extractor.extract_map(99)
   ```

2. Review validation report and adjust if needed

### Step 4: Manual Verification
1. Compare rendered map with original image
2. Check tile count and distribution statistics
3. Verify high-confidence classifications
4. Manually correct any obvious errors

## Validation Process

### Automatic Validation

The system performs several automatic validation checks:

1. **Tile Count Check**: Compares extracted count with expected count
2. **Distribution Analysis**: Verifies tile type percentages
3. **Confidence Analysis**: Reports average confidence per tile type
4. **Grid Completeness**: Ensures no missing tiles in expected positions

### Manual Validation

For critical validation, perform these manual checks:

1. **Visual Comparison**: Compare rendered map with original image
2. **Spot Checking**: Manually verify a sample of tiles
3. **Edge Cases**: Pay special attention to edge tiles and similar-looking tiles
4. **Statistical Review**: Check for unreasonable tile distributions

### Validation Report

The system generates an HTML validation report containing:

- **Summary Statistics**: Tile counts, accuracy metrics, confidence scores
- **Side-by-side Images**: Original vs rendered comparison
- **Confidence Heatmap**: Visual representation of classification confidence
- **Error Analysis**: Details of misclassified or low-confidence tiles
- **Recommendations**: Suggested improvements or manual corrections

## File Structure

```
maps/
├── README.md                 # This documentation
├── map_extractor.py         # Main extraction system
├── hex_grid_renderer.py     # Visualization utilities
├── requirements.txt         # Python dependencies
├── outputs/                 # Generated files
│   ├── rendered_maps/       # Rendered map images
│   ├── comparisons/         # Side-by-side comparisons
│   └── reports/             # Validation reports
└── examples/                # Example usage scripts
```

## Troubleshooting

### Common Issues

#### No Hex Cells Detected
- **Cause**: Poor image quality or unusual hex arrangement
- **Solution**: Try adjusting edge detection parameters or use manual grid specification

#### Low Classification Confidence
- **Cause**: Similar-looking tiles or poor image quality
- **Solution**: Add more reference tiles or improve image preprocessing

#### Incorrect Tile Count
- **Cause**: Missed tiles or false detections
- **Solution**: Review hex detection parameters and validate grid structure

#### Memory Issues
- **Cause**: Large images or many reference tiles
- **Solution**: Resize images or process in batches

### Debug Mode

Enable debug mode for detailed logging:
```python
extractor = MapExtractor(debug=True)
```

This will output:
- Intermediate processing images
- Detection statistics
- Classification confidence scores
- Grid organization details

### Performance Optimization

For better performance:
1. **Resize Images**: Scale down large images before processing
2. **Limit Reference Tiles**: Use only necessary tile types
3. **Adjust Detection Parameters**: Tune for your specific use case
4. **Use Caching**: Cache processed reference tiles

## Contributing

### Adding New Detection Methods

1. Add method to `MapExtractor` class
2. Update `detect_hex_grid()` to use new method as fallback
3. Add unit tests for new method
4. Update documentation

### Improving Classification

1. Add new tile features (texture, edges, etc.)
2. Implement in `_classify_hex_region()` method
3. Add validation for new features
4. Update confidence calculation

### Extending Validation

1. Add new validation metrics to `_validate_extraction()`
2. Update validation report generation
3. Add new visualization options
4. Document new validation features

## License

This map extraction system is part of the TurnEngine project and follows the same license terms.