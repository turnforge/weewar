# WeeWar WASM Web Interface

This directory contains the web interface for WeeWar's WASM modules, providing a complete browser-based experience for playing and editing WeeWar games.

## Files Overview

### HTML Pages
- **`index.html`** - Main landing page with navigation and combined demo
- **`cli.html`** - Dedicated Game CLI interface with full debugging capabilities  
- **`editor.html`** - Dedicated Map Editor interface with advanced tools

### WASM Modules (built by `scripts/build-wasm.sh`)
- **`../wasm/weewar-cli.wasm`** - Main game CLI (14MB)
- **`../wasm/editor.wasm`** - Map editor (14MB)  
- **`../wasm/wasm_exec.js`** - Go WASM runtime (20KB)

## Features

### Game CLI (`cli.html`)
- ✅ **Complete Game Management**: Create, save, load games
- ✅ **Command Execution**: Full CLI command set (move, attack, status, etc.)
- ✅ **Real-time Rendering**: PNG generation with multiple sizes
- ✅ **Save/Load System**: Download/upload game files
- ✅ **Debug Console**: Detailed operation logging and status
- ✅ **Responsive Design**: Works on desktop and mobile

### Map Editor (`editor.html`) 
- ✅ **Terrain Painting**: 5 terrain types with emoji palette
- ✅ **Advanced Brushes**: Variable size (1-91 hexes), flood fill
- ✅ **Undo/Redo**: 50-step history with visual feedback
- ✅ **Map Validation**: Real-time issue detection
- ✅ **Export Options**: Generate playable games (2-4 players)
- ✅ **Advanced Tools**: Island generator, mountain ridges, randomization
- ✅ **Click-to-Paint**: Click rendered map to set coordinates

## JavaScript API

### CLI Functions
```javascript
// Game Management
weewarCreateGame(playerCount)
weewarLoadGame(jsonData)
weewarSaveGame()

// Command Execution  
weewarExecuteCommand(command)
weewarGetGameState()

// Rendering & Settings
weewarRenderGame(width, height)
weewarSetVerbose(enabled)
weewarSetDisplayMode(mode)
```

### Editor Functions
```javascript
// Map Management
editorCreate()
editorNewMap(rows, cols)
editorGetMapInfo()
editorValidateMap()

// Terrain Editing
editorSetBrushTerrain(type)
editorSetBrushSize(size)  
editorPaintTerrain(row, col)
editorFloodFill(row, col)
editorRemoveTerrain(row, col)

// History
editorUndo() / editorRedo()
editorCanUndo() / editorCanRedo()

// Export & Rendering
editorRenderMap(width, height)
editorExportToGame(playerCount)
```

## Running Locally

### Option 1: Simple HTTP Server
```bash
# From the web directory
python3 -m http.server 8000
# Visit http://localhost:8000
```

### Option 2: Node.js Server
```bash
npx serve .
# Visit http://localhost:3000
```

### Option 3: Live Server (VS Code)
Install the "Live Server" extension and right-click any HTML file → "Open with Live Server"

## Development

### Building WASM Modules
```bash
# From the main weewar directory
./scripts/build-wasm.sh
```

### Debugging
- **Browser Console**: Check for JavaScript errors and WASM load issues
- **Network Tab**: Verify WASM files are loading (14MB each)
- **Debug Panels**: Each page includes debug information panels
- **Console Output**: Real-time operation logging with timestamps

### Performance Notes
- **Initial Load**: ~28MB total (both WASM modules)
- **Runtime**: Pure client-side execution, no server required
- **Memory**: Go garbage collector manages WASM memory automatically
- **Rendering**: Base64 PNG generation for cross-browser compatibility

## Browser Compatibility

- ✅ **Chrome/Edge**: Full support (recommended)
- ✅ **Firefox**: Full support
- ✅ **Safari**: Full support (iOS/macOS)
- ❌ **IE**: Not supported (no WASM support)

## Deployment

These files can be deployed to any static web hosting service:
- **GitHub Pages**
- **Netlify** 
- **Vercel**
- **AWS S3**
- **Any HTTP server**

No server-side processing required - everything runs in the browser!