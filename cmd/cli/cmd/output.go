package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	"github.com/turnforge/weewar/lib"
)

// OutputFormatter handles formatting output in text or JSON
type OutputFormatter struct {
	JSON   bool
	Dryrun bool
}

// NewOutputFormatter creates a new formatter based on global flags
func NewOutputFormatter() *OutputFormatter {
	return &OutputFormatter{
		JSON:   isJSONOutput(),
		Dryrun: isDryrun(),
	}
}

// prefix adds [DRYRUN] prefix if in dryrun mode
func (f *OutputFormatter) prefix(text string) string {
	if f.Dryrun {
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			if line != "" {
				lines[i] = "[DRYRUN] " + line
			}
		}
		return strings.Join(lines, "\n")
	}
	return text
}

// Print outputs text or JSON based on format setting
func (f *OutputFormatter) Print(data any) error {
	if f.JSON {
		return f.PrintJSON(data)
	}
	return f.PrintText(data)
}

// PrintJSON outputs data as JSON
func (f *OutputFormatter) PrintJSON(data any) error {
	output := map[string]any{
		"data":   data,
		"dryrun": f.Dryrun,
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonBytes))
	return nil
}

// PrintText outputs data as human-readable text
func (f *OutputFormatter) PrintText(data any) error {
	var text string

	switch v := data.(type) {
	case string:
		text = v
	case fmt.Stringer:
		text = v.String()
	default:
		text = fmt.Sprintf("%v", v)
	}

	fmt.Println(f.prefix(text))
	return nil
}

// FormatOptions formats available options as text
func FormatOptions(pc *PresenterContext, position string) string {
	var sb strings.Builder

	// Get unit info
	if pc.TurnOptions.Unit != nil {
		unit := pc.TurnOptions.Unit
		coord := lib.CoordFromInt32(unit.Q, unit.R)
		sb.WriteString(fmt.Sprintf("Unit %s at %s:\n", position, coord.String()))
		sb.WriteString(fmt.Sprintf("  Type: %d, HP: %d, Moves: %f\n\n",
			unit.UnitType, unit.AvailableHealth, unit.DistanceLeft))
	} else {
		// No unit - show position for tile options
		sb.WriteString(fmt.Sprintf("Tile %s:\n\n", position))
	}

	// Get options
	if pc.TurnOptions.Options == nil || len(pc.TurnOptions.Options.Options) == 0 {
		sb.WriteString("No options available at this position\n")
		return sb.String()
	}

	sb.WriteString("Available options:\n")

	for i, option := range pc.TurnOptions.Options.Options {
		switch opt := option.OptionType.(type) {
		case *v1.GameOption_Move:
			moveOpt := opt.Move
			targetCoord := lib.CoordFromInt32(moveOpt.ToQ, moveOpt.ToR)
			sb.WriteString(fmt.Sprintf("%d. move to %s (cost: %f)\n",
				i+1, targetCoord.String(), moveOpt.MovementCost))

			// Add path if available
			if moveOpt.ReconstructedPath != nil {
				pathStr := lib.FormatPathCompact(moveOpt.ReconstructedPath)
				sb.WriteString(fmt.Sprintf("   Path: %s\n", pathStr))
			}

		case *v1.GameOption_Attack:
			attackOpt := opt.Attack
			targetCoord := lib.CoordFromInt32(attackOpt.DefenderQ, attackOpt.DefenderR)
			sb.WriteString(fmt.Sprintf("%d. attack %s (damage est: %d)\n",
				i+1, targetCoord.String(), attackOpt.DamageEstimate))

		case *v1.GameOption_Build:
			buildOpt := opt.Build
			unitName := fmt.Sprintf("type %d", buildOpt.UnitType) // fallback

			// Try to get the actual unit details from RulesEngine
			if pc.Presenter != nil && pc.Presenter.RulesEngine != nil {
				rulesEngine := &lib.RulesEngine{RulesEngine: pc.Presenter.RulesEngine}
				if unitDef, err := rulesEngine.GetUnitData(buildOpt.UnitType); err == nil {
					unitName = unitDef.Name

					// Build detailed info line
					var details []string

					// Classification (e.g., "Heavy Land")
					if unitDef.UnitClass != "" && unitDef.UnitTerrain != "" {
						details = append(details, fmt.Sprintf("%s %s", unitDef.UnitClass, unitDef.UnitTerrain))
					}

					// Movement points
					if unitDef.MovementPoints > 0 {
						details = append(details, fmt.Sprintf("⚡%.0f", unitDef.MovementPoints))
					}

					// Attack range
					if unitDef.AttackRange > 0 {
						details = append(details, fmt.Sprintf("🎯%d", unitDef.AttackRange))
					}

					// Defense
					if unitDef.Defense > 0 {
						details = append(details, fmt.Sprintf("🛡️%d", unitDef.Defense))
					}

					sb.WriteString(fmt.Sprintf("%d. build %s (cost: %d, type: %d)\n",
						i+1, unitName, buildOpt.Cost, buildOpt.UnitType))

					// Add details on next line
					if len(details) > 0 {
						sb.WriteString(fmt.Sprintf("   %s\n", strings.Join(details, " • ")))
					}

					// Add properties if available
					if len(unitDef.Properties) > 0 {
						for _, prop := range unitDef.Properties {
							sb.WriteString(fmt.Sprintf("   • %s\n", prop))
						}
					}
					continue // Skip the default format line
				}
			}

			// Fallback if we couldn't get unit data
			sb.WriteString(fmt.Sprintf("%d. build %s (cost: %d, type: %d)\n",
				i+1, unitName, buildOpt.Cost, buildOpt.UnitType))

		case *v1.GameOption_EndTurn:
			sb.WriteString(fmt.Sprintf("%d. end turn\n", i+1))
		}
	}

	return sb.String()
}

// FormatGameStatus formats game status as text
func FormatGameStatus(game *v1.Game, state *v1.GameState) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Game: %s\n", game.Name))
	if game.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", game.Description))
	}
	sb.WriteString(fmt.Sprintf("\nTurn: %d\n", state.TurnCounter))
	sb.WriteString(fmt.Sprintf("Current Player: %d\n", state.CurrentPlayer))
	sb.WriteString(fmt.Sprintf("Game Status: %s\n", state.Status))

	if state.WinningPlayer != 0 {
		sb.WriteString(fmt.Sprintf("\nGame Over! Winner: Player %d\n", state.WinningPlayer))
	}

	// Count units per player
	unitCounts := make(map[int32]int)
	if state.WorldData != nil {
		for _, unit := range state.WorldData.UnitsMap {
			if unit != nil {
				unitCounts[unit.Player]++
			}
		}
	}

	// Count tiles per player
	tileCounts := make(map[int32]int)
	if state.WorldData != nil {
		for _, tile := range state.WorldData.TilesMap {
			if tile != nil && tile.Player > 0 {
				tileCounts[tile.Player]++
			}
		}
	}

	// Show player info
	if game.Config != nil && len(game.Config.Players) > 0 {
		sb.WriteString("\nPlayers:\n")
		for _, player := range game.Config.Players {
			indicator := ""
			if player.PlayerId == state.CurrentPlayer {
				indicator = " *"
			}
			sb.WriteString(fmt.Sprintf("  Player %d%s:\n", player.PlayerId, indicator))
			sb.WriteString(fmt.Sprintf("    Type: %s\n", player.PlayerType))
			if player.Name != "" {
				sb.WriteString(fmt.Sprintf("    Name: %s\n", player.Name))
			}
			sb.WriteString(fmt.Sprintf("    Coins: %d\n", player.Coins))
			sb.WriteString(fmt.Sprintf("    Units: %d\n", unitCounts[player.PlayerId]))
			if tileCounts[player.PlayerId] > 0 {
				sb.WriteString(fmt.Sprintf("    Tiles: %d\n", tileCounts[player.PlayerId]))
			}
			if player.TeamId > 0 {
				sb.WriteString(fmt.Sprintf("    Team: %d\n", player.TeamId))
			}
		}
	}

	return sb.String()
}

// FormatUnits formats all units as text
func FormatUnits(pc *PresenterContext, state *v1.GameState) string {
	if state.WorldData == nil || len(state.WorldData.UnitsMap) == 0 {
		return "No units found\n"
	}

	pc, _, _, _, rtGame, err := GetGame()
	if err != nil {
		panic(err)
	}

	// Group units by player
	unitsByPlayer := make(map[int32][]*v1.Unit)
	numPlayers := int32(0)

	for _, unit := range state.WorldData.UnitsMap {
		if unit != nil {
			if err := rtGame.TopUpUnitIfNeeded(unit); err != nil {
				panic(err)
			}
			unitsByPlayer[unit.Player] = append(unitsByPlayer[unit.Player], unit)
			if unit.Player > numPlayers {
				numPlayers = unit.Player
			}
		}
	}

	var sb strings.Builder
	// Iterate starting from current player and wrap around
	for i := int32(0); i < numPlayers; i++ {
		playerID := (i + state.CurrentPlayer) % numPlayers
		if playerID == 0 {
			playerID = numPlayers
		}
		units := unitsByPlayer[playerID]
		if len(units) == 0 {
			continue
		}
		turnIndicator := ""
		if playerID == state.CurrentPlayer {
			turnIndicator = " *"
		}
		sb.WriteString(fmt.Sprintf("Player %d units%s:\n", playerID, turnIndicator))

		for _, unit := range units {
			coord := lib.CoordFromInt32(unit.Q, unit.R)
			unitID := unit.Shortcut
			if unitID == "" {
				// Player 1 -> 'A', Player 2 -> 'B', etc.
				playerLetter := string(rune('A' + playerID - 1))
				unitID = fmt.Sprintf("%s?", playerLetter)
			}
			sb.WriteString(fmt.Sprintf("  %s: Type %d at %s (HP: %d, Moves: %f)\n",
				unitID, unit.UnitType, coord.String(),
				unit.AvailableHealth, unit.DistanceLeft))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatTiles formats all tiles as text
func FormatTiles(pc *PresenterContext, state *v1.GameState) string {
	if state.WorldData == nil || len(state.WorldData.TilesMap) == 0 {
		return "No tiles found\n"
	}

	pc, _, _, _, rtGame, err := GetGame()
	if err != nil {
		panic(err)
	}

	// Group tiles by player
	tilesByPlayer := make(map[int32][]*v1.Tile)
	numPlayers := int32(0)

	for _, tile := range state.WorldData.TilesMap {
		if tile != nil && tile.Player > 0 {
			if err := rtGame.TopUpTileIfNeeded(tile); err != nil {
				panic(err)
			}
			tilesByPlayer[tile.Player] = append(tilesByPlayer[tile.Player], tile)
			if tile.Player > numPlayers {
				numPlayers = tile.Player
			}
		}
	}

	var sb strings.Builder
	// Iterate starting from current player and wrap around
	for i := int32(0); i < numPlayers; i++ {
		playerID := (i + state.CurrentPlayer) % numPlayers
		if playerID == 0 {
			playerID = numPlayers
		}
		tiles := tilesByPlayer[playerID]
		if len(tiles) == 0 {
			continue
		}
		turnIndicator := ""
		if playerID == state.CurrentPlayer {
			turnIndicator = " *"
		}
		sb.WriteString(fmt.Sprintf("Player %d tiles%s:\n", playerID, turnIndicator))

		for _, tile := range tiles {
			coord := lib.CoordFromInt32(tile.Q, tile.R)
			tileID := tile.Shortcut
			if tileID == "" {
				// Player 1 -> 'A', Player 2 -> 'B', etc.
				playerLetter := string(rune('A' + playerID - 1))
				tileID = fmt.Sprintf("%s?", playerLetter)
			}
			sb.WriteString(fmt.Sprintf("  %s: Type %d at %s\n",
				tileID, tile.TileType, coord.String()))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
