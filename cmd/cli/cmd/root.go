package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	gameID    string
	serverURL string
	jsonOut   bool
	verbose   bool
	dryrun    bool
	confirm   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "ww",
	Short:        "LilBattle CLI - Command-line interface for LilBattle games",
	SilenceUsage: true,
	Long: `LilBattle CLI provides a command-line interface for playing LilBattle games.

Examples:
  ww options A1                    Show available options for unit A1
  ww move A1 R                     Move unit A1 to the right
  ww attack A1 TR                  Attack top-right from unit A1
  ww endturn                       End current player's turn
  ww status                        Show game status
  ww units                         List all units

Global Flags:
  --game-id string       Game ID to operate on (or set LILBATTLE_GAME_ID env var)
  --server string        Server URL to connect to (or set LILBATTLE_SERVER env var)
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.lilbattle.yaml)")
	rootCmd.PersistentFlags().StringVar(&gameID, "game-id", "", "game ID to operate on (env: LILBATTLE_GAME_ID)")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "", "server URL to connect to (env: LILBATTLE_SERVER)")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "show detailed debug information")
	rootCmd.PersistentFlags().BoolVar(&dryrun, "dryrun", false, "preview changes without saving to disk")
	rootCmd.PersistentFlags().BoolVar(&confirm, "confirm", true, "prompt for confirmation on destructive actions")

	// Bind flags to viper
	viper.BindPFlag("game-id", rootCmd.PersistentFlags().Lookup("game-id"))
	viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
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

		// Search config in home directory with name ".lilbattle" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".lilbattle")
	}

	// Read environment variables with LILBATTLE_ prefix
	viper.SetEnvPrefix("LILBATTLE")
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
		return "", fmt.Errorf("game ID is required (set --game-id flag or LILBATTLE_GAME_ID env var)")
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

// getServerURL returns the server URL if configured, empty string for local mode
func getServerURL() string {
	if rootCmd.PersistentFlags().Changed("server") {
		return serverURL
	}
	return viper.GetString("server")
}
