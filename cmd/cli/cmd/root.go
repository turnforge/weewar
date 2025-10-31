package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	gameID   string
	jsonOut  bool
	verbose  bool
	dryrun   bool
	confirm  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ww",
	Short: "WeeWar CLI - Command-line interface for WeeWar games",
	Long: `WeeWar CLI provides a command-line interface for playing WeeWar games.

Examples:
  ww options A1                    Show available options for unit A1
  ww move A1 R                     Move unit A1 to the right
  ww attack A1 TR                  Attack top-right from unit A1
  ww endturn                       End current player's turn
  ww status                        Show game status
  ww units                         List all units

Global Flags:
  --game-id string       Game ID to operate on (or set GAME_ID env var)
  --json                 Output in JSON format
  --verbose              Show detailed debug information
  --dryrun               Preview changes without saving to disk`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.weewar.yaml)")
	rootCmd.PersistentFlags().StringVar(&gameID, "game-id", "", "game ID to operate on (env: GAME_ID)")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "show detailed debug information")
	rootCmd.PersistentFlags().BoolVar(&dryrun, "dryrun", false, "preview changes without saving to disk")
	rootCmd.PersistentFlags().BoolVar(&confirm, "confirm", true, "prompt for confirmation on destructive actions")

	// Bind flags to viper
	viper.BindPFlag("game-id", rootCmd.PersistentFlags().Lookup("game-id"))
	viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("dryrun", rootCmd.PersistentFlags().Lookup("dryrun"))
	viper.BindPFlag("confirm", rootCmd.PersistentFlags().Lookup("confirm"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Search config in home directory with name ".weewar" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".weewar")
	}

	// Read environment variables with WEEWAR_ prefix
	viper.SetEnvPrefix("WEEWAR")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// getGameID retrieves the game ID from env var or flag (flag overrides)
func getGameID() (string, error) {
	var id string

	// Check if --game-id flag was explicitly provided
	if rootCmd.PersistentFlags().Changed("game-id") {
		// Flag was provided, use it (overrides env var)
		id = gameID
	} else {
		// No flag provided, get from env var via viper
		id = viper.GetString("game-id")
	}

	if id == "" {
		return "", fmt.Errorf("game ID is required (set --game-id flag or WEEWAR_GAME_ID env var)")
	}
	return id, nil
}

// isJSONOutput returns whether JSON output is requested
func isJSONOutput() bool {
	return viper.GetBool("json")
}

// isVerbose returns whether verbose output is requested
func isVerbose() bool {
	return viper.GetBool("verbose")
}

// isDryrun returns whether dryrun mode is active
func isDryrun() bool {
	return viper.GetBool("dryrun")
}

// shouldConfirm returns whether confirmation prompts should be shown
func shouldConfirm() bool {
	return viper.GetBool("confirm")
}
