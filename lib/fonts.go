package weewar

import (
	"fmt"
	"sync"

	"github.com/panyam/turnengine/games/weewar/assets"
	"github.com/tdewolff/canvas"
)

var (
	defaultFontFamily *canvas.FontFamily
	fontInitOnce      sync.Once
)

func init() {
	defaultFontFamily = initializeEmbeddedFont()
}

// GetDefaultFontFamily returns a font family that works in WASM environments
func GetDefaultFontFamily() *canvas.FontFamily {
	return defaultFontFamily
}

// initializeEmbeddedFont initializes the embedded font
func initializeEmbeddedFont() *canvas.FontFamily {
	fontFamily := canvas.NewFontFamily("Roboto")

	// Load font directly from embedded byte data (works in WASM)
	if err := fontFamily.LoadFont(assets.RobotoRegularTTF, 0, canvas.FontRegular); err != nil {
		fmt.Println("Error loading embedded font: ", err)

		// For WASM, we can't use system fonts, so create a minimal fallback
		// that won't crash but will at least allow the program to run
		fontFamily = canvas.NewFontFamily("fallback")
		fmt.Println("Using fallback font family (no actual font loaded)")
		return fontFamily
	}

	fmt.Println("Successfully loaded embedded Roboto font")
	return fontFamily
}
