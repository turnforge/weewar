package cmd

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/turnforge/lilbattle/lib"
	"github.com/turnforge/lilbattle/web/assets/themes"
)

// mapCmd represents the map command
var mapCmd = &cobra.Command{
	Use:   "map",
	Short: "Display the game map in the terminal",
	Long: `Render and display the current game map as an inline image in the terminal.
Requires iTerm2, Kitty, or another terminal with inline image support.

Examples:
  ww map
  ww map --labels         # Show unit labels (Shortcut:MP/Health)
  ww map --no-labels      # Hide unit labels
  ww map --tile-labels    # Show tile labels (Shortcut)
  ww map -o map.png       # Save to file instead of displaying`,
	RunE: runMap,
}

var (
	showLabels     bool
	showTileLabels bool
	outputFile     string
)

func init() {
	rootCmd.AddCommand(mapCmd)
	mapCmd.Flags().BoolVar(&showLabels, "labels", true, "Show unit labels (Shortcut:MP/Health)")
	mapCmd.Flags().BoolVar(&showTileLabels, "tile-labels", true, "Show tile labels (Shortcut)")

	// Default to environment variable if set
	defaultOutput := os.Getenv("LILBATTLE_MAP_OUTPUT")
	mapCmd.Flags().StringVarP(&outputFile, "output", "o", defaultOutput, "Save image to file instead of displaying (env: LILBATTLE_MAP_OUTPUT)")
}

func runMap(cmd *cobra.Command, args []string) error {
	gc, err := GetGameContext()
	if err != nil {
		return err
	}

	if gc.State == nil {
		return fmt.Errorf("game state not initialized")
	}

	state := gc.State
	if state.WorldData == nil {
		return fmt.Errorf("world data not available")
	}

	// Create theme for rendering using cityTerrains from default rules engine
	theme := themes.NewDefaultTheme(lib.DefaultRulesEngine().GetCityTerrains())
	renderer, err := themes.NewPNGWorldRenderer(theme)
	if err != nil {
		return fmt.Errorf("failed to create renderer: %w", err)
	}

	// Set up render options
	options := lib.DefaultRenderOptions()
	options.ShowUnitLabels = showLabels
	options.ShowTileLabels = showTileLabels

	// Render the map
	pngData, _, err := renderer.Render(state.WorldData.TilesMap, state.WorldData.UnitsMap, options)
	if err != nil {
		return fmt.Errorf("failed to render map: %w", err)
	}

	// If output file specified, save to file
	if outputFile != "" {
		if err := os.WriteFile(outputFile, pngData, 0644); err != nil {
			return fmt.Errorf("failed to write image to %s: %w", outputFile, err)
		}
		fmt.Printf("Map saved to %s\n", outputFile)
		return nil
	}

	// Display inline image using iTerm2 escape sequence
	// Format: ESC ] 1337 ; File = [args] : base64_data BEL
	// Using inline=1 and preserveAspectRatio=1 to maintain correct dimensions
	encoded := base64.StdEncoding.EncodeToString(pngData)
	fmt.Printf("\033]1337;File=inline=1;preserveAspectRatio=1:%s\a", encoded)
	fmt.Println() // newline after image

	return nil
}
