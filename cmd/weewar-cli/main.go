package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	weewar "github.com/panyam/turnengine/games/weewar/lib"
)

// Version information
const (
	Version = "1.0.0"
	Build   = "dev"
)

func main() {
	// Command line flags
	var (
		interactive = flag.Bool("interactive", false, "Start in interactive mode")
		newGame     = flag.Bool("new", false, "Create a new game")
		players     = flag.Int("players", 2, "Number of players for new game (2-6)")
		loadFile    = flag.String("load", "", "Load game from file")
		saveFile    = flag.String("save", "", "Save game to file after commands")
		renderFile  = flag.String("render", "", "Render game to PNG file")
		width       = flag.Int("width", 800, "Render width in pixels")
		height      = flag.Int("height", 600, "Render height in pixels")
		batch       = flag.String("batch", "", "Execute commands from batch file")
		record      = flag.String("record", "", "Record session to file")
		verbose     = flag.Bool("verbose", false, "Enable verbose output")
		compact     = flag.Bool("compact", false, "Use compact display mode")
		autoRender  = flag.Bool("autorender", false, "Auto-render game state after each command")
		maxRenders  = flag.Int("maxrenders", 10, "Maximum number of auto-rendered files to keep (0 disables)")
		renderDir   = flag.String("renderdir", "/tmp/turnengine/autorenders", "Directory for auto-rendered files")
		version     = flag.Bool("version", false, "Show version information")
		help        = flag.Bool("help", false, "Show help information")
	)
	flag.Parse()

	// Show version
	if *version {
		fmt.Printf("WeeWar CLI v%s (build %s)\n", Version, Build)
		return
	}

	// Show help
	if *help {
		showHelp()
		return
	}

	// Validate flags
	if *players < 2 || *players > 6 {
		log.Fatalf("Invalid number of players: %d (must be 2-6)", *players)
	}

	// Create CLI instance
	var cli *weewar.WeeWarCLI
	var game *weewar.Game

	// Initialize game
	if *loadFile != "" {
		// Load existing game
		fmt.Printf("Loading game from %s...\n", *loadFile)
		if err := loadGameFromFile(*loadFile, &game); err != nil {
			log.Fatalf("Failed to load game: %v", err)
		}
		fmt.Println("Game loaded successfully")
	} else if *newGame || *interactive {
		// Create new game
		fmt.Printf("Creating new game with %d players...\n", *players)
		var err error
		game, err = createNewGame(*players)
		if err != nil {
			log.Fatalf("Failed to create game: %v", err)
		}
		fmt.Println("Game created successfully")
	}

	// Create CLI
	cli = weewar.NewWeeWarCLI(game)

	// Set options
	if *verbose {
		cli.SetVerbose(true)
	}
	if *compact {
		cli.SetDisplayMode(weewar.DisplayCompact)
	}
	// Always set maxRenders from command line flag
	cli.SetMaxRenders(*maxRenders)
	cli.SetRenderDir(*renderDir)
	if *autoRender {
		cli.SetAutoRender(true)
	}

	// Start recording if requested
	if *record != "" {
		if err := cli.RecordSession(*record); err != nil {
			log.Fatalf("Failed to start recording: %v", err)
		}
		fmt.Printf("Recording session to %s\n", *record)
		defer cli.StopRecording()
	}

	// Execute batch commands if provided
	if *batch != "" {
		fmt.Printf("Executing batch commands from %s...\n", *batch)
		if err := cli.ExecuteBatchCommands(*batch); err != nil {
			log.Fatalf("Batch execution failed: %v", err)
		}
		fmt.Println("Batch commands completed successfully")
	}

	// Execute remaining command line arguments as commands
	if len(flag.Args()) > 0 {
		for _, cmd := range flag.Args() {
			fmt.Printf("Executing: %s\n", cmd)
			response := cli.ExecuteCommand(cmd)
			fmt.Printf("Result: %s\n", response.Message)
			if !response.Success {
				log.Printf("Command failed: %s", response.Error)
			}
		}
	}

	// Save game if requested
	if *saveFile != "" {
		fmt.Printf("Saving game to %s...\n", *saveFile)
		if err := cli.SaveGameToFile(*saveFile); err != nil {
			log.Fatalf("Failed to save game: %v", err)
		}
		fmt.Println("Game saved successfully")
	}

	// Render game if requested
	if *renderFile != "" {
		fmt.Printf("Rendering game to %s (%dx%d)...\n", *renderFile, *width, *height)
		if err := cli.RenderToFile(*renderFile, *width, *height); err != nil {
			log.Fatalf("Failed to render game: %v", err)
		}
		fmt.Println("Game rendered successfully")
	}

	// Start interactive mode if requested
	if *interactive {
		cli.StartInteractiveMode()
	}
}

// createNewGame creates a new game with specified number of players
func createNewGame(playerCount int) (*weewar.Game, error) {
	// Create test map
	testMap := weewar.NewMap(8, 12, false)

	// Add varied terrain
	for row := 0; row < 8; row++ {
		for col := 0; col < 12; col++ {
			tileType := 1 // Default to grass
			if (row+col)%4 == 0 {
				tileType = 2 // Some desert
			} else if (row+col)%7 == 0 {
				tileType = 3 // Some water
			} else if (row+col)%11 == 0 {
				tileType = 4 // Some mountains
			}

			tile := weewar.NewTile(row, col, tileType)
			testMap.AddTile(tile)
		}
	}

	// Note: Neighbor connections calculated on-demand

	// Create game
	seed := time.Now().UnixNano()
	return weewar.NewGame(playerCount, testMap, seed)
}

// loadGameFromFile loads a game from a file
func loadGameFromFile(filename string, game **weewar.Game) error {
	saveData, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read save file: %w", err)
	}

	loadedGame, err := weewar.LoadGame(saveData)
	if err != nil {
		return fmt.Errorf("failed to load game: %w", err)
	}

	*game = loadedGame
	return nil
}

// showHelp displays help information
func showHelp() {
	fmt.Printf("WeeWar CLI v%s - Command Line Interface for WeeWar Games\n\n", Version)

	fmt.Println("USAGE:")
	fmt.Println("  weewar-cli [options] [commands...]")
	fmt.Println()

	fmt.Println("OPTIONS:")
	fmt.Println("  -interactive         Start in interactive mode")
	fmt.Println("  -new                 Create a new game")
	fmt.Println("  -players N           Number of players for new game (2-6, default: 2)")
	fmt.Println("  -load FILE           Load game from file")
	fmt.Println("  -save FILE           Save game to file after commands")
	fmt.Println("  -render FILE         Render game to PNG file")
	fmt.Println("  -width N             Render width in pixels (default: 800)")
	fmt.Println("  -height N            Render height in pixels (default: 600)")
	fmt.Println("  -batch FILE          Execute commands from batch file")
	fmt.Println("  -record FILE         Record session to file")
	fmt.Println("  -verbose             Enable verbose output")
	fmt.Println("  -compact             Use compact display mode")
	fmt.Println("  -autorender          Auto-render game state after each command")
	fmt.Println("  -maxrenders N        Maximum number of auto-rendered files to keep (default: 10, 0 disables)")
	fmt.Println("  -renderdir DIR       Directory for auto-rendered files (default: /tmp/turnengine/autorenders)")
	fmt.Println("  -version             Show version information")
	fmt.Println("  -help                Show this help")
	fmt.Println()

	fmt.Println("GAME COMMANDS:")
	fmt.Println("  move A1 B2           Move unit from A1 to B2")
	fmt.Println("  attack A1 B2         Attack unit at B2 with unit at A1")
	fmt.Println("  status               Show current game status")
	fmt.Println("  map                  Display the game map")
	fmt.Println("  units                Show all units")
	fmt.Println("  player [N]           Show player information")
	fmt.Println("  end                  End current player's turn")
	fmt.Println("  help [command]       Show help for specific command")
	fmt.Println("  quit                 Exit the game")
	fmt.Println()

	fmt.Println("EXAMPLES:")
	fmt.Println("  # Start new interactive game")
	fmt.Println("  weewar-cli -new -interactive")
	fmt.Println()
	fmt.Println("  # Load game and show status")
	fmt.Println("  weewar-cli -load mygame.json status")
	fmt.Println()
	fmt.Println("  # Create game, make moves, and save")
	fmt.Println("  weewar-cli -new 'move A1 B2' 'end' -save mygame.json")
	fmt.Println()
	fmt.Println("  # Render game to PNG")
	fmt.Println("  weewar-cli -load mygame.json -render game.png")
	fmt.Println()
	fmt.Println("  # Execute batch commands")
	fmt.Println("  weewar-cli -new -batch commands.txt")
	fmt.Println()

	fmt.Println("BATCH FILE FORMAT:")
	fmt.Println("  # Comments start with #")
	fmt.Println("  move A1 B2")
	fmt.Println("  attack B2 C3")
	fmt.Println("  end")
	fmt.Println("  # Another comment")
	fmt.Println("  status")
	fmt.Println()

	fmt.Println("POSITION FORMAT:")
	fmt.Println("  Positions use chess notation: A1, B2, C3, etc.")
	fmt.Println("  Columns are A-Z, rows are 1-99")
	fmt.Println()

	fmt.Println("For more information, visit: https://github.com/panyam/turnengine")
}
