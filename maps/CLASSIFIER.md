# Tile Classifier Implementation Plan

## 🎯 **Objective**
Create a robust tile classifier that compares extracted hex tiles against reference tiles and returns confidence scores for each match, with support for both grayscale and color modes.

## 📋 **Problem Analysis**

### **Input:**
- Extracted hex tiles from `hex_splitter` (hexagonal with transparent backgrounds)
- N reference tiles labeled as `TileId.png` 
- Need both grayscale and color matching modes

### **Output:**
- Confidence scores (0-1) for each reference tile
- Ranked list of matches with confidence levels

### **Challenges:**
1. **Transparency handling**: Hexagonal tiles with transparent backgrounds
2. **Size variations**: Tiles might have slightly different dimensions
3. **Lighting variations**: Different brightness/contrast between images
4. **Similar tiles**: Some terrain types might look very similar
5. **Noise**: Image compression artifacts and anti-aliasing
6. **Performance**: Need to classify many tiles efficiently

## 🔬 **Similarity Metrics Analysis**

### **Template Matching Approaches:**
1. **Normalized Cross-Correlation (NCC)**
   - ✅ Good for exact pattern matching
   - ✅ Robust to brightness changes
   - ❌ Sensitive to rotation/scale
   - **Use case**: Identical tile detection

2. **Structural Similarity Index (SSIM)**
   - ✅ Perceptually meaningful similarity
   - ✅ Handles luminance/contrast variations
   - ✅ Good for texture comparison
   - **Use case**: Perceptual tile similarity

3. **Mean Squared Error (MSE)**
   - ✅ Simple and fast
   - ❌ Sensitive to pixel-level noise
   - **Use case**: Baseline comparison

### **Distribution-Based Approaches:**
4. **Histogram Comparison**
   - ✅ Robust to spatial variations
   - ✅ Good for color/texture analysis
   - ❌ Loses spatial information
   - **Use case**: Color/intensity distribution matching

5. **Color Moments**
   - ✅ Compact representation
   - ✅ Good for color characteristics
   - **Use case**: Color-based classification

### **Feature-Based Approaches:**
6. **ORB/SIFT Keypoint Matching**
   - ✅ Robust to rotation/scale
   - ✅ Good for distinctive features
   - ❌ May fail on uniform textures
   - **Use case**: Tiles with distinctive patterns

## 🏗️ **Implementation Strategy**

### **1. Core Architecture**
```python
class TileClassifier:
    def __init__(self, reference_tiles_dir, grayscale_mode=False, debug_mode=False)
    def classify_tile(self, tile_image_path) -> List[Tuple[str, float]]
    def classify_batch(self, tile_dir) -> Dict[str, List[Tuple[str, float]]]
```

### **2. Multi-Metric Similarity Engine**

#### **Preprocessing Pipeline:**
- **Transparency handling**: Crop to hex bounding box or use alpha-weighted calculations
- **Size normalization**: Resize all tiles to consistent dimensions
- **Color space conversion**: RGB ↔ Grayscale based on mode
- **Intensity normalization**: Handle brightness/contrast variations

#### **Similarity Metrics (Grayscale Mode):**
1. **Normalized Cross-Correlation (NCC)**: Exact pattern matching
2. **Structural Similarity Index (SSIM)**: Perceptual similarity
3. **Histogram Correlation**: Intensity distribution matching
4. **Mean Squared Error (MSE)**: Pixel-wise difference (normalized)

#### **Similarity Metrics (Color Mode):**
1. **All grayscale metrics** applied per-channel (R,G,B)
2. **Color Histogram Comparison**: HSV/RGB histogram correlation
3. **Color Moments**: Mean, variance, skewness of color channels
4. **Color Coherence Vector**: Spatial color distribution analysis

#### **Feature-Based Matching (Optional Enhancement):**
- **ORB Keypoint Matching**: For tiles with distinctive features
- **Local Binary Patterns (LBP)**: Texture analysis
- **Hu Moments**: Shape-based similarity

### **3. Score Fusion & Ranking**
```python
def _fuse_similarity_scores(self, metrics_dict) -> float:
    # Weighted combination of multiple metrics
    # Configurable weights based on tile type characteristics
    
def _calibrate_confidence(self, raw_score) -> float:
    # Convert raw similarity to meaningful 0-1 confidence
    # Handle score normalization and outlier rejection
```

### **4. File Structure**
```
tile_classifier.py           # Main classifier implementation
reference_tiles/            # Directory with TileId.png files
├── 1.png                  # Grass tile
├── 2.png                  # Water tile  
├── 3.png                  # Mountain tile
└── ...
extracted_tiles/           # Input tiles from hex_splitter
├── 0_0.png               # Row 0, Col 0
├── 1_2.png               # Row 1, Col 2
└── ...
results/                   # Classification results
├── classification_results.json
└── debug_images/         # Visual comparison results
```

### **5. Key Features to Implement**

#### **Robustness Features:**
- **Multi-scale comparison**: Test at different sizes
- **Rotation tolerance**: Limited rotation invariance  
- **Noise filtering**: Gaussian blur preprocessing option
- **Confidence thresholding**: "No match" detection for unknown tiles

#### **Performance Features:**
- **Batch processing**: Classify multiple tiles efficiently
- **Caching**: Pre-compute reference tile features
- **Parallel processing**: Multi-threaded similarity calculations
- **Progress tracking**: Real-time classification progress

#### **Debug Features:**
- **Visual comparisons**: Side-by-side tile comparison images
- **Metric breakdown**: Individual similarity scores per metric
- **Confidence calibration**: Score distribution analysis
- **Confusion matrix**: Classification accuracy analysis

### **6. CLI Interface**
```bash
# Single tile classification
python tile_classifier.py --tile extracted_tiles/0_0.png --references reference_tiles/ --mode color

# Batch classification  
python tile_classifier.py --batch extracted_tiles/ --references reference_tiles/ --mode grayscale --debug

# Results analysis
python tile_classifier.py --analyze results/classification_results.json --visualize
```

### **7. Implementation Steps**
1. **Create TileClassifier class** with reference tile loading
2. **Implement preprocessing pipeline** (transparency, normalization)
3. **Add individual similarity metrics** (NCC, SSIM, histograms, etc.)
4. **Build score fusion system** with configurable weights
5. **Add batch processing capability** for multiple tiles
6. **Create debug visualization system** for result analysis
7. **Implement CLI interface** with comprehensive options
8. **Add performance optimizations** (caching, parallelization)

### **8. Success Metrics**
- **Accuracy**: >95% correct classification on known tile types
- **Confidence calibration**: Meaningful 0-1 scores
- **Performance**: <100ms per tile classification
- **Robustness**: Handle minor variations in lighting/size
- **Usability**: Simple CLI for batch processing

## 🔧 **Technical Implementation Details**

### **Transparency Handling Strategies:**
1. **Alpha masking**: Only compare pixels where both tiles have alpha > threshold
2. **Bounding box cropping**: Extract minimal rectangle containing non-transparent content
3. **Alpha-weighted metrics**: Weight similarity calculations by alpha channel values

### **Normalization Techniques:**
1. **Size normalization**: Resize to common dimensions (e.g., 64x64)
2. **Intensity normalization**: Histogram equalization or z-score normalization
3. **Contrast enhancement**: CLAHE (Contrast Limited Adaptive Histogram Equalization)

### **Score Fusion Approaches:**
1. **Simple averaging**: Equal weight to all metrics
2. **Weighted combination**: Learned or heuristic weights per metric
3. **Rank fusion**: Combine rankings rather than raw scores
4. **Machine learning**: Train classifier on metric combinations

### **Performance Optimizations:**
1. **Feature caching**: Pre-compute and cache reference tile features
2. **Early termination**: Skip expensive metrics if fast metrics show poor match
3. **Parallel processing**: Multi-threaded comparison across reference tiles
4. **Memory optimization**: Process tiles in batches to control memory usage

## 📊 **Evaluation Strategy**

### **Test Data Requirements:**
- Ground truth tile classifications for accuracy measurement
- Variations in lighting, rotation, scale for robustness testing
- Similar tile pairs for discrimination testing
- Novel tiles for "no match" detection testing

### **Metrics for Evaluation:**
- **Classification accuracy**: Percentage of correct top-1 predictions
- **Top-k accuracy**: Percentage where correct tile is in top-k predictions
- **Confidence calibration**: Relationship between confidence scores and accuracy
- **Processing speed**: Average time per tile classification
- **Memory usage**: Peak memory consumption during batch processing

## 🎯 **Future Enhancements**

### **Machine Learning Integration:**
- **Deep learning features**: Use pre-trained CNN features (ResNet, VGG)
- **Siamese networks**: Train specialized similarity networks
- **Metric learning**: Learn optimal distance functions for tile comparison
- **Active learning**: Iteratively improve with user feedback

### **Advanced Features:**
- **Tile synthesis**: Generate new tile variations for training
- **Hierarchical classification**: Group tiles by categories (terrain, building, etc.)
- **Spatial context**: Consider neighboring tiles for classification
- **Multi-resolution analysis**: Compare at multiple scales simultaneously

This comprehensive plan provides a roadmap for building a robust, extensible tile classification system that can handle the complexity of hex tile matching while providing meaningful confidence scores.