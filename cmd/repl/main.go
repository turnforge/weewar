package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/chzyer/readline"
)

// Version information
const (
	Version = "2.0.0"
	Build   = "headless"
)

func main() {
	// Command line flags
	var (
		help    = flag.Bool("help", false, "Show help information")
		version = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	// Show version
	if *version {
		fmt.Printf("LilBattle CLI v%s (build %s) - Headless Game State Manipulator\n", Version, Build)
		return
	}

	// Show help
	if *help || len(flag.Args()) == 0 {
		showHelp()
		return
	}

	// Get game ID from arguments
	gameID := flag.Args()[0]

	// Create CLI instance
	cli, err := NewCLI(gameID)
	if err != nil {
		log.Fatalf("Failed to initialize CLI: %v", err)
	}
	defer cli.Close()

	fmt.Printf("LilBattle CLI - Game %s loaded\n", gameID)
	fmt.Println("Type 'help' for available commands, 'quit' to exit")
	fmt.Println("Use ↑/↓ arrow keys to navigate command history")

	// Execute any remaining command line arguments as commands
	if len(flag.Args()) > 1 {
		for _, cmd := range flag.Args()[1:] {
			fmt.Printf("> %s\n", cmd)
			result := cli.ExecuteCommand(cmd)
			if result == "quit" {
				return
			}
			fmt.Println(result)
		}
	}

	// Start interactive REPL
	startREPL(cli)
}

// showHelp displays help information
func showHelp() {
	fmt.Printf("LilBattle CLI v%s - Headless Game State Manipulator\n\n", Version)

	fmt.Println("USAGE:")
	fmt.Println("  lilbattle-cli <gameid> [commands...]")
	fmt.Println()

	fmt.Println("ARGUMENTS:")
	fmt.Println("  gameid               Game ID to load from storage/games/<gameid>/")
	fmt.Println("  commands             Optional commands to execute before entering REPL")
	fmt.Println()

	fmt.Println("OPTIONS:")
	fmt.Println("  -help                Show this help")
	fmt.Println("  -version             Show version information")
	fmt.Println()

	fmt.Println("GAME COMMANDS:")
	fmt.Println("  move <from> <to>     Move unit (e.g. \"move A1 3,4\")")
	fmt.Println("  attack <att> <tgt>   Attack target (e.g. \"attack A1 B2\")")
	fmt.Println("  select <unit>        Select unit and show options")
	fmt.Println("  end                  End current player's turn")
	fmt.Println("  status               Show game status")
	fmt.Println("  units                Show all units")
	fmt.Println("  player [ID]          Show player information")
	fmt.Println("  help                 Show command help")
	fmt.Println("  quit                 Exit the CLI")
	fmt.Println()

	fmt.Println("EXAMPLES:")
	fmt.Println("  # Load game and enter interactive mode")
	fmt.Println("  lilbattle-cli game123")
	fmt.Println()
	fmt.Println("  # Load game and execute commands")
	fmt.Println("  lilbattle-cli game123 \"move A1 3,4\" \"end\"")
	fmt.Println()
	fmt.Println("  # Show game status")
	fmt.Println("  lilbattle-cli game123 status")
	fmt.Println()

	fmt.Println("NOTES:")
	fmt.Println("  - Game state is automatically saved after each command")
	fmt.Println("  - Changes are immediately visible to the browser on refresh")
	fmt.Println("  - Game files are stored in storage/games/<gameid>/")
}

// startREPL starts the interactive REPL with readline support
func startREPL(cli *CLI) {
	for {
		// Read input with history support
		line, err := cli.readline.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				continue
			} else if err == io.EOF {
				fmt.Println("\nGoodbye!")
				break
			}
			log.Printf("Error reading input: %v", err)
			break
		}

		command := strings.TrimSpace(line)
		if command == "" {
			continue
		}

		// Execute command
		result := cli.ExecuteCommand(command)

		// Check for quit
		if result == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		// Show result
		fmt.Println(result)
		fmt.Println()
	}
}
