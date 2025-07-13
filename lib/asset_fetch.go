//go:build js && wasm
// +build js,wasm

package weewar

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"sync"
	"syscall/js"
)

// =============================================================================
// WASM Fetch-Based Asset System - HTTP Lazy Loading for Browser
// =============================================================================

// FetchAssetManager is a WASM-specific AssetManager that uses HTTP fetch for lazy loading
type FetchAssetManager struct {
	baseURL    string
	tileCache  map[int]image.Image
	unitCache  map[string]image.Image // key: "unitId_playerColor"
	cacheMutex sync.RWMutex
	loaded     bool
}

// NewFetchAssetManager creates a new fetch-based asset manager for WASM
func NewFetchAssetManager(baseURL string) *FetchAssetManager {
	// Default to current origin if no baseURL provided
	if baseURL == "" {
		baseURL = "."
	}
	
	// Remove trailing slash if present
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	
	return &FetchAssetManager{
		baseURL:    baseURL,
		tileCache:  make(map[int]image.Image),
		unitCache:  make(map[string]image.Image),
		cacheMutex: sync.RWMutex{},
	}
}

// GetTileImage returns the tile image for a given tile type using HTTP fetch
func (fam *FetchAssetManager) GetTileImage(tileType int) (image.Image, error) {
	fam.cacheMutex.RLock()
	if img, exists := fam.tileCache[tileType]; exists {
		fam.cacheMutex.RUnlock()
		return img, nil
	}
	fam.cacheMutex.RUnlock()
	
	// Construct URL for tile asset
	tileURL := fmt.Sprintf("%s/data/Tiles/%d_files/0.png", fam.baseURL, tileType)
	fmt.Printf("🔍 Fetching tile asset: %s\n", tileURL)
	
	img, err := fam.fetchImage(tileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tile image for type %d from %s: %w", tileType, tileURL, err)
	}
	
	// Cache the image
	fam.cacheMutex.Lock()
	fam.tileCache[tileType] = img
	fam.cacheMutex.Unlock()
	
	fmt.Printf("✅ Successfully loaded tile %d\n", tileType)
	return img, nil
}

// GetUnitImage returns the unit image for a given unit type and player color using HTTP fetch
func (fam *FetchAssetManager) GetUnitImage(unitType int, playerColor int) (image.Image, error) {
	key := fmt.Sprintf("%d_%d", unitType, playerColor)
	
	fam.cacheMutex.RLock()
	if img, exists := fam.unitCache[key]; exists {
		fam.cacheMutex.RUnlock()
		return img, nil
	}
	fam.cacheMutex.RUnlock()
	
	// Construct URL for unit asset
	unitURL := fmt.Sprintf("%s/data/Units/%d_files/%d.png", fam.baseURL, unitType, playerColor)
	fmt.Printf("🔍 Fetching unit asset: %s\n", unitURL)
	
	img, err := fam.fetchImage(unitURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch unit image for type %d, color %d from %s: %w", unitType, playerColor, unitURL, err)
	}
	
	// Cache the image
	fam.cacheMutex.Lock()
	fam.unitCache[key] = img
	fam.cacheMutex.Unlock()
	
	fmt.Printf("✅ Successfully loaded unit %d, player %d\n", unitType, playerColor)
	return img, nil
}

// fetchImage fetches an image from a URL using JavaScript fetch API
func (fam *FetchAssetManager) fetchImage(url string) (image.Image, error) {
	// Create a channel to wait for the async fetch operation
	done := make(chan struct {
		img image.Image
		err error
	}, 1)
	
	// Call JavaScript fetch API
	js.Global().Call("fetch", url).Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		response := args[0]
		
		// Check if response is ok
		if !response.Get("ok").Bool() {
			statusText := response.Get("statusText").String()
			done <- struct {
				img image.Image
				err error
			}{nil, fmt.Errorf("HTTP %d: %s", response.Get("status").Int(), statusText)}
			return nil
		}
		
		// Get array buffer from response
		response.Call("arrayBuffer").Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			arrayBuffer := args[0]
			
			// Convert ArrayBuffer to Go byte slice
			uint8Array := js.Global().Get("Uint8Array").New(arrayBuffer)
			length := uint8Array.Get("length").Int()
			data := make([]byte, length)
			js.CopyBytesToGo(data, uint8Array)
			
			// Decode PNG using bytes.NewReader to avoid corrupting binary data
			img, err := png.Decode(bytes.NewReader(data))
			done <- struct {
				img image.Image
				err error
			}{img, err}
			
			return nil
		})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			jsErr := args[0]
			done <- struct {
				img image.Image
				err error
			}{nil, fmt.Errorf("arrayBuffer error: %s", jsErr.Get("message").String())}
			return nil
		}))
		
		return nil
	})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsErr := args[0]
		done <- struct {
			img image.Image
			err error
		}{nil, fmt.Errorf("fetch error: %s", jsErr.Get("message").String())}
		return nil
	}))
	
	// Wait for the async operation to complete
	result := <-done
	return result.img, result.err
}

// HasTileAsset checks if a tile asset exists (always returns true for fetch-based)
func (fam *FetchAssetManager) HasTileAsset(tileType int) bool {
	// For fetch-based loading, we assume assets exist and handle 404s during fetch
	return true
}

// HasUnitAsset checks if a unit asset exists (always returns true for fetch-based)
func (fam *FetchAssetManager) HasUnitAsset(unitType int, playerColor int) bool {
	// For fetch-based loading, we assume assets exist and handle 404s during fetch
	return true
}

// PreloadCommonAssets preloads commonly used assets for better performance
func (fam *FetchAssetManager) PreloadCommonAssets() error {
	fmt.Println("🚀 Starting preload of common assets...")
	
	// Preload common tile types (1-26)
	for i := 1; i <= 26; i++ {
		_, err := fam.GetTileImage(i)
		if err != nil {
			fmt.Printf("⚠️ Warning: Could not preload tile %d: %v\n", i, err)
		}
	}
	
	// Preload common unit types with basic player colors (0-5)
	for unitType := 1; unitType <= 44; unitType++ {
		for playerColor := 0; playerColor <= 5; playerColor++ {
			_, err := fam.GetUnitImage(unitType, playerColor)
			if err != nil {
				// Continue on error - not all combinations exist
				continue
			}
		}
	}
	
	fam.loaded = true
	fmt.Println("✅ Asset preloading complete!")
	return nil
}

// GetCacheStats returns statistics about cached assets
func (fam *FetchAssetManager) GetCacheStats() (int, int) {
	fam.cacheMutex.RLock()
	defer fam.cacheMutex.RUnlock()
	
	return len(fam.tileCache), len(fam.unitCache)
}

// ClearCache clears all cached assets
func (fam *FetchAssetManager) ClearCache() {
	fam.cacheMutex.Lock()
	defer fam.cacheMutex.Unlock()
	
	fam.tileCache = make(map[int]image.Image)
	fam.unitCache = make(map[string]image.Image)
	fmt.Println("🗑️ Asset cache cleared")
}

// IsLoaded returns whether assets have been preloaded
func (fam *FetchAssetManager) IsLoaded() bool {
	return fam.loaded
}