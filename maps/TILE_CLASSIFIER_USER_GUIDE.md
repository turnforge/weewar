# Tile Classifier User Guide

A comprehensive guide for using the multi-metric tile classification system to identify hex tiles from WeeWar maps.

## 📋 **Table of Contents**

1. [Quick Start](#quick-start)
2. [Installation & Setup](#installation--setup)
3. [Directory Structure](#directory-structure)
4. [Usage Examples](#usage-examples)
5. [Command-Line Options](#command-line-options)
6. [Understanding Results](#understanding-results)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)
9. [Advanced Usage](#advanced-usage)

---

## 🚀 **Quick Start**

### **Step 1: Prepare Your Data**
```bash
# Create directory structure
mkdir reference_tiles extracted_tiles results

# Place your reference tiles in reference_tiles/
# - 1.png (grass tile)
# - 2.png (water tile)
# - 3.png (mountain tile)
# etc.
```

### **Step 2: Extract Tiles from Map**
```bash
# First extract individual tiles from your map
python hex_splitter.py --image map.png --output-dir extracted_tiles --rows 7 --cols 7
```

### **Step 3: Classify Tiles**
```bash
# Classify all extracted tiles
python tile_classifier.py --batch extracted_tiles --references reference_tiles --output results.json
```

### **Step 4: View Results**
```bash
# Results are saved in results.json with confidence scores and matches
cat results.json
```

---

## 🛠️ **Installation & Setup**

### **Dependencies**
```bash
pip install opencv-python numpy scikit-image scipy pathlib
```

### **Required Files**
- `tile_classifier.py` - Main classifier script
- `hex_splitter.py` - For extracting tiles from maps (optional)
- Reference tile images (PNG format)
- Extracted tile images to classify

---

## 📁 **Directory Structure**

```
project/
├── tile_classifier.py          # Main classifier script
├── hex_splitter.py            # Tile extraction tool
├── reference_tiles/           # Reference tile images
│   ├── 1.png                 # Grass tile
│   ├── 2.png                 # Water tile
│   ├── 3.png                 # Mountain tile
│   ├── 4.png                 # Forest tile
│   └── ...
├── extracted_tiles/           # Tiles to classify
│   ├── 00_00.png             # Row 0, Col 0
│   ├── 00_01.png             # Row 0, Col 1
│   ├── 01_00.png             # Row 1, Col 0
│   └── ...
├── results/                   # Classification results
│   ├── classification_results.json
│   └── debug_images/         # Debug visualizations (if enabled)
└── debug_images/             # Debug output directory
    └── classification/
```

---

## 💡 **Usage Examples**

### **Example 1: Single Tile Classification**

Classify one specific tile to see detailed analysis:

```bash
python tile_classifier.py \
  --tile extracted_tiles/00_00.png \
  --references reference_tiles \
  --mode color \
  --debug
```

**Output:**
```
Loading 5 reference tiles...
Successfully loaded 5 reference tiles
Classifying single tile: extracted_tiles/00_00.png

Classification Results:
Best match: 1 (confidence: 0.876)
Top 5 matches:
  1. 1: 0.876
  2. 3: 0.234
  3. 2: 0.123
  4. 4: 0.098
  5. 5: 0.067

Results saved to: classification_results.json
```

### **Example 2: Batch Classification**

Classify all tiles in a directory:

```bash
python tile_classifier.py \
  --batch extracted_tiles \
  --references reference_tiles \
  --mode color \
  --output batch_results.json
```

**Output:**
```
Loading 5 reference tiles...
Successfully loaded 5 reference tiles
Classifying 49 tiles...
Successfully classified 49 tiles

Batch Classification Summary:
Total tiles classified: 49
High confidence matches (>0.8): 42
Classification accuracy: 85.7%

Batch results saved to: batch_results.json
```

### **Example 3: Grayscale Mode (Faster)**

Use grayscale mode for faster processing:

```bash
python tile_classifier.py \
  --batch extracted_tiles \
  --references reference_tiles \
  --mode grayscale \
  --output grayscale_results.json
```

### **Example 4: Debug Mode Analysis**

Enable debug mode to see detailed metric breakdown:

```bash
python tile_classifier.py \
  --tile extracted_tiles/00_00.png \
  --references reference_tiles \
  --debug
```

**Debug Output:**
```
Loading 5 reference tiles...
  Loaded reference tile: 1
  Loaded reference tile: 2
  Loaded reference tile: 3
  Loaded reference tile: 4
  Loaded reference tile: 5
Successfully loaded 5 reference tiles

Debug classification for 00_00:
  Best match: 1 (confidence: 0.876)
  Match 1: 1 (confidence: 0.876)
  Match 2: 3 (confidence: 0.234)
  Match 3: 2 (confidence: 0.123)
```

---

## ⚙️ **Command-Line Options**

### **Required Arguments**
| Option | Description | Example |
|--------|-------------|---------|
| `--references` | Directory containing reference tile images | `--references reference_tiles` |

### **Input Options (Choose One)**
| Option | Description | Example |
|--------|-------------|---------|
| `--tile` | Single tile image to classify | `--tile extracted_tiles/00_00.png` |
| `--batch` | Directory containing tiles to classify | `--batch extracted_tiles` |

### **Processing Options**
| Option | Values | Default | Description |
|--------|--------|---------|-------------|
| `--mode` | `color`, `grayscale` | `color` | Classification mode |
| `--output` | filename | `classification_results.json` | Output file path |
| `--debug` | flag | disabled | Enable debug mode |

### **Complete Command Reference**
```bash
python tile_classifier.py \
  --references <reference_dir> \
  [--tile <single_tile_path> | --batch <tiles_directory>] \
  [--mode {color,grayscale}] \
  [--output <output_file>] \
  [--debug]
```

---

## 📊 **Understanding Results**

### **Single Tile Results**
```json
{
  "tile_path": "extracted_tiles/00_00.png",
  "best_match": "1",
  "best_confidence": 0.876,
  "matches": [
    ["1", 0.876],
    ["3", 0.234],
    ["2", 0.123]
  ],
  "metrics_breakdown": {
    "1": {
      "ncc": 0.892,
      "ssim": 0.834,
      "histogram_correlation": 0.901,
      "mse_similarity": 0.887,
      "fused_confidence": 0.876
    }
  }
}
```

### **Batch Results**
```json
{
  "00_00.png": {
    "best_match": "1",
    "best_confidence": 0.876,
    "matches": [["1", 0.876], ["3", 0.234]],
    "metrics_breakdown": {...}
  },
  "00_01.png": {
    "best_match": "2",
    "best_confidence": 0.923,
    "matches": [["2", 0.923], ["1", 0.156]],
    "metrics_breakdown": {...}
  }
}
```

### **Confidence Score Interpretation**
| Score Range | Interpretation | Action |
|-------------|----------------|---------|
| 0.9 - 1.0 | Excellent match | High confidence |
| 0.8 - 0.9 | Good match | Reliable |
| 0.6 - 0.8 | Moderate match | Review manually |
| 0.4 - 0.6 | Poor match | Likely incorrect |
| 0.0 - 0.4 | Very poor match | Definitely incorrect |

### **Individual Metrics Explained**
- **NCC (Normalized Cross-Correlation)**: Template matching similarity (0-1)
- **SSIM (Structural Similarity Index)**: Perceptual similarity (0-1)
- **Histogram Correlation**: Color/intensity distribution similarity (-1 to 1)
- **MSE Similarity**: Pixel-wise difference converted to similarity (0-1)

---

## 🎯 **Best Practices**

### **1. Reference Tile Quality**
- Use high-quality, clean reference tiles
- Ensure reference tiles are representative of actual map tiles
- Include all tile types you expect to encounter
- Use consistent naming (1.png, 2.png, etc.)

### **2. Tile Extraction**
- Extract tiles at consistent size and quality
- Use hex_splitter with proper parameters
- Ensure tiles are properly centered on hex boundaries

### **3. Classification Settings**
- Use **color mode** for better accuracy when tiles have distinct colors
- Use **grayscale mode** for faster processing when color isn't critical
- Enable **debug mode** when fine-tuning or troubleshooting

### **4. Result Analysis**
- Review tiles with confidence < 0.8 manually
- Check for consistent patterns in misclassification
- Update reference tiles if needed

### **5. Performance Tips**
- Process tiles in batches rather than individually
- Use grayscale mode for large datasets
- Pre-process reference tiles once and reuse

---

## 🔧 **Troubleshooting**

### **Common Issues**

#### **"Reference tiles directory not found"**
```bash
# Check if directory exists
ls -la reference_tiles/

# Create directory if missing
mkdir reference_tiles

# Verify PNG files exist
ls reference_tiles/*.png
```

#### **"No PNG files found in reference directory"**
```bash
# Check file extensions
ls reference_tiles/

# Ensure files are .png (not .PNG, .jpg, etc.)
# Convert if needed:
for f in reference_tiles/*.jpg; do
    convert "$f" "${f%.jpg}.png"
done
```

#### **"Could not load tile image"**
```bash
# Check file permissions
ls -la extracted_tiles/00_00.png

# Verify file is not corrupted
file extracted_tiles/00_00.png
```

#### **Low Classification Accuracy**
1. **Check reference tile quality**
   - Ensure they're clean and representative
   - Try different reference tiles

2. **Verify tile extraction quality**
   - Check hex_splitter parameters
   - Ensure proper tile centering

3. **Adjust classification mode**
   - Try grayscale if color is causing issues
   - Enable debug mode to see metric breakdown

#### **Slow Performance**
1. **Use grayscale mode**
   ```bash
   --mode grayscale
   ```

2. **Process in smaller batches**
   ```bash
   # Split large directories
   mkdir batch1 batch2
   mv extracted_tiles/0*.png batch1/
   mv extracted_tiles/1*.png batch2/
   ```

---

## 🔬 **Advanced Usage**

### **Custom Metric Weights**
The classifier uses configurable weights for combining metrics. To modify:

```python
# In tile_classifier.py, modify the metric_weights dictionary
self.metric_weights = {
    'ncc': 0.4,                    # Increase NCC weight
    'ssim': 0.3,                   # Standard SSIM weight
    'histogram_correlation': 0.2,   # Standard histogram weight
    'mse_similarity': 0.1          # Decrease MSE weight
}
```

### **Batch Processing with Custom Scripts**
```python
#!/usr/bin/env python3
"""Custom batch processing script"""

from tile_classifier import TileClassifier
import json

# Initialize classifier
classifier = TileClassifier(
    reference_tiles_dir="reference_tiles",
    grayscale_mode=False,
    debug_mode=True
)

# Process multiple directories
directories = ["map1_tiles", "map2_tiles", "map3_tiles"]
all_results = {}

for directory in directories:
    print(f"Processing {directory}...")
    results = classifier.classify_batch(directory)
    all_results[directory] = results

# Save combined results
with open("combined_results.json", "w") as f:
    json.dump(all_results, f, indent=2)
```

### **Integration with Map Processing Pipeline**
```bash
#!/bin/bash
# Complete map processing pipeline

# 1. Extract tiles from map
python hex_splitter.py --image map.png --output-dir extracted_tiles --rows 7 --cols 7

# 2. Classify tiles
python tile_classifier.py --batch extracted_tiles --references reference_tiles --output results.json

# 3. Generate map visualization
python map_visualizer.py --tiles extracted_tiles --classification results.json --output map_visualization.png
```

---

## 📈 **Performance Benchmarks**

### **Typical Processing Times**
| Operation | Color Mode | Grayscale Mode |
|-----------|------------|----------------|
| Single tile | ~200ms | ~100ms |
| 49 tiles (7x7) | ~10s | ~5s |
| 100 tiles | ~20s | ~10s |

### **Memory Usage**
- **Reference tiles**: ~5MB per tile (64x64 normalized)
- **Processing**: ~50MB for typical batch
- **Results**: ~1KB per tile classification

---

## 🎯 **Use Cases**

### **Game Development**
- Reverse engineer map layouts from screenshots
- Validate procedurally generated maps
- Analyze competitor game maps

### **Data Analysis**
- Classify terrain types in satellite imagery
- Analyze hex-based strategy games
- Process game state recognition

### **Quality Assurance**
- Verify map generation algorithms
- Test tile rendering consistency
- Validate game asset integrity

---

## 🔗 **Related Tools**

- **hex_splitter.py**: Extract tiles from hex grid images
- **grid_analyzer.py**: Analyze hex grid structure
- **hex_generator.py**: Generate hex grid coordinates
- **map_visualizer.py**: Visualize classification results

---

## 📝 **Tips for Success**

1. **Start small**: Test with a few tiles before processing large batches
2. **Use debug mode**: Enable debugging when learning the system
3. **Validate results**: Manually check high-confidence matches
4. **Iterate**: Refine reference tiles based on results
5. **Document**: Keep notes on what works for your specific use case

---

This user guide provides comprehensive coverage of the tile classifier system. For additional questions or advanced customization, refer to the source code comments and CLASSIFIER.md implementation plan.