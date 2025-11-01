package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"github.com/panyam/turnengine/games/weewar/services"
)

// attackCmd represents the attack command
var attackCmd = &cobra.Command{
	Use:   "attack <attacker> <target>",
	Short: "Attack a unit",
	Long: `Attack a target unit with your unit.
Positions can be unit IDs (like A1) or coordinates (like 3,4).
The <target> position can also be a direction or sequence of directions: L, R, TL, TR, BL, BR.
Multiple directions can be chained with commas to target distant units: TL,TL,TR.

Examples:
  ww attack A1 B2              Attack unit B2 with unit A1
  ww attack A1 TR              Attack top-right neighbor with A1
  ww attack A1 TL,TL           Attack unit 2 steps top-left from A1
  ww attack 3,4 5,6            Attack position 5,6 with unit at 3,4
  ww attack A1 B2 --dryrun     Preview attack outcome without saving`,
	Args: cobra.ExactArgs(2),
	RunE: runAttack,
}

func init() {
	rootCmd.AddCommand(attackCmd)
}

func runAttack(cmd *cobra.Command, args []string) error {
	attackerPos := args[0]
	targetPos := args[1]

	ctx := context.Background()
	pc, _, _, _, rtGame, err := GetGame()
	if err != nil {
		return err
	}

	// Parse attacker position
	attackerTarget, err := services.ParsePositionOrUnit(rtGame, attackerPos)
	if err != nil {
		return fmt.Errorf("invalid attacker position: %w", err)
	}
	attackerCoord := attackerTarget.GetCoordinate()

	// Parse target position with context (supports directions)
	targetTarget, err := services.ParsePositionOrUnitWithContext(rtGame, targetPos, &attackerCoord)
	if err != nil {
		return fmt.Errorf("invalid target position: %w", err)
	}
	targetCoord := targetTarget.GetCoordinate()

	if isVerbose() {
		fmt.Printf("[VERBOSE] Attacking from %s to %s\n", attackerCoord.String(), targetCoord.String())
	}

	// Two-click pattern: Click attacker to select, then click target to attack
	// Click 1: Select attacker on base-map layer
	_, err = pc.Presenter.SceneClicked(ctx, &v1.SceneClickedRequest{
		GameId: gameID,
		Q:      int32(attackerCoord.Q),
		R:      int32(attackerCoord.R),
		Layer:  "base-map",
	})
	if err != nil {
		return fmt.Errorf("failed to select attacker: %w", err)
	}

	if isVerbose() {
		fmt.Printf("[VERBOSE] Attacker selected at %s\n", attackerCoord.String())
	}

	// Get combat diagnostics before executing attack
	var combatDiagnostics string
	if isVerbose() || isDryrun() {
		attacker := rtGame.World.UnitAt(attackerCoord)
		attackerTile := rtGame.World.TileAt(attackerCoord)
		defender := rtGame.World.UnitAt(targetCoord)
		defenderTile := rtGame.World.TileAt(targetCoord)

		if attacker != nil && defender != nil {
			combatDiagnostics = generateCombatDiagnostics(
				rtGame.GetRulesEngine(),
				attacker, attackerTile, attacker.AvailableHealth,
				defender, defenderTile, defender.AvailableHealth)
			fmt.Print(combatDiagnostics)
		}
	}

	// Click 2: Click target on movement-highlight layer to execute attack
	_, err = pc.Presenter.SceneClicked(ctx, &v1.SceneClickedRequest{
		GameId: gameID,
		Q:      int32(targetCoord.Q),
		R:      int32(targetCoord.R),
		Layer:  "movement-highlight",
	})
	if err != nil {
		return fmt.Errorf("failed to execute attack: %w", err)
	}

	if isVerbose() {
		fmt.Printf("[VERBOSE] Attack executed on %s\n", targetCoord.String())
	}

	// Save state unless in dryrun mode
	if err := savePresenterState(pc, isDryrun()); err != nil {
		return err
	}

	// Format output
	formatter := NewOutputFormatter()

	if formatter.JSON {
		data := map[string]any{
			"game_id": gameID,
			"action":  "attack",
			"attacker": map[string]int{
				"q": attackerCoord.Q,
				"r": attackerCoord.R,
			},
			"target": map[string]int{
				"q": targetCoord.Q,
				"r": targetCoord.R,
			},
			"success": true,
		}
		return formatter.PrintJSON(data)
	}

	// Text output
	var sb strings.Builder
	sb.WriteString("Attack: Success\n")
	sb.WriteString(fmt.Sprintf("  Attacked from %s to %s\n", attackerCoord.String(), targetCoord.String()))
	sb.WriteString(fmt.Sprintf("\nCurrent player: %d, Turn: %d\n",
		pc.GameState.State.CurrentPlayer, pc.GameState.State.TurnCounter))

	return formatter.PrintText(sb.String())
}

// generateCombatDiagnostics produces detailed combat calculation information
// This is a static, testable function that allows testing with arbitrary health values
func generateCombatDiagnostics(
	rulesEngine *services.RulesEngine,
	attacker *v1.Unit, attackerTile *v1.Tile, attackerHealth int32,
	defender *v1.Unit, defenderTile *v1.Tile, defenderHealth int32) string {

	var sb strings.Builder
	sb.WriteString("\n[COMBAT DIAGNOSTICS]\n")
	sb.WriteString(strings.Repeat("=", 60) + "\n")

	// Unit information
	attackerData, err := rulesEngine.GetUnitData(attacker.UnitType)
	defenderData, err := rulesEngine.GetUnitData(defender.UnitType)

	attackerCoord := services.CoordFromInt32(attacker.Q, attacker.R)
	defenderCoord := services.CoordFromInt32(defender.Q, defender.R)

	sb.WriteString(fmt.Sprintf("\nAttacker: %s (Type %d) at %s\n", attacker.Shortcut, attacker.UnitType, attackerCoord.String()))
	sb.WriteString(fmt.Sprintf("  Player: %d\n", attacker.Player))
	sb.WriteString(fmt.Sprintf("  Health: %d/%d\n", attackerHealth, attackerData.Health))
	sb.WriteString(fmt.Sprintf("  Unit Type: %s\n", attackerData.Name))

	// Get attacker's terrain
	if attackerTile != nil {
		attackerTerrainData, _ := rulesEngine.GetTerrainData(attackerTile.TileType)
		if attackerTerrainData != nil {
			sb.WriteString(fmt.Sprintf("  Terrain: %s (Type %d)\n", attackerTerrainData.Name, attackerTile.TileType))
		}
	}

	sb.WriteString(fmt.Sprintf("\nDefender: %s (Type %d) at %s\n", defender.Shortcut, defender.UnitType, defenderCoord.String()))
	sb.WriteString(fmt.Sprintf("  Player: %d\n", defender.Player))
	sb.WriteString(fmt.Sprintf("  Health: %d/%d\n", defenderHealth, defenderData.Health))
	sb.WriteString(fmt.Sprintf("  Unit Type: %s\n", defenderData.Name))

	// Get defender's terrain
	if defenderTile != nil {
		defenderTerrainData, _ := rulesEngine.GetTerrainData(defenderTile.TileType)
		if defenderTerrainData != nil {
			sb.WriteString(fmt.Sprintf("  Terrain: %s (Type %d)\n", defenderTerrainData.Name, defenderTile.TileType))
		}
	}

	// Combat calculations
	sb.WriteString("\n" + strings.Repeat("-", 60) + "\n")
	sb.WriteString("Combat Calculations:\n")
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	// Attacker -> Defender damage
	attackDist, canAttack := rulesEngine.GetCombatPrediction(attacker.UnitType, defender.UnitType)
	sb.WriteString(fmt.Sprintf("\nAttacker -> Defender:\n"))
	if canAttack && attackDist != nil {
		sb.WriteString(fmt.Sprintf("  Can Attack: YES\n"))
		sb.WriteString(fmt.Sprintf("  Expected Damage: %.1f\n", attackDist.ExpectedDamage))
		if len(attackDist.Ranges) > 0 {
			sb.WriteString(fmt.Sprintf("  Damage Distribution:\n"))
			for i, dr := range attackDist.Ranges {
				sb.WriteString(fmt.Sprintf("    Range %d: %.0f-%.0f damage (%.1f%% probability)\n",
					i+1, dr.MinValue, dr.MaxValue, dr.Probability*100))
			}
		}
	} else {
		sb.WriteString(fmt.Sprintf("  Can Attack: NO\n"))
	}

	// Defender -> Attacker counter-attack
	counterDist, canCounter := rulesEngine.GetCombatPrediction(defender.UnitType, attacker.UnitType)
	sb.WriteString(fmt.Sprintf("\nDefender -> Attacker (Counter-attack):\n"))
	if canCounter && counterDist != nil {
		// Check if defender can actually reach attacker
		canReach, _ := rulesEngine.CanUnitAttackTarget(defender, attacker)
		if canReach {
			sb.WriteString(fmt.Sprintf("  Can Counter: YES\n"))
			sb.WriteString(fmt.Sprintf("  Expected Damage: %.1f\n", counterDist.ExpectedDamage))
			if len(counterDist.Ranges) > 0 {
				sb.WriteString(fmt.Sprintf("  Damage Distribution:\n"))
				for i, dr := range counterDist.Ranges {
					sb.WriteString(fmt.Sprintf("    Range %d: %.0f-%.0f damage (%.1f%% probability)\n",
						i+1, dr.MinValue, dr.MaxValue, dr.Probability*100))
				}
			}
		} else {
			sb.WriteString(fmt.Sprintf("  Can Counter: NO (out of range)\n"))
		}
	} else {
		sb.WriteString(fmt.Sprintf("  Can Counter: NO (unit type cannot counter-attack)\n"))
	}

	// Formula-based calculations
	sb.WriteString("\n" + strings.Repeat("-", 60) + "\n")
	sb.WriteString("Formula-Based Calculations:\n")
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	// Calculate wound bonus
	woundBonus := rulesEngine.CalculateWoundBonus(defender, attackerCoord)

	// Create combat context for attacker -> defender
	attackerCtx := &services.CombatContext{
		Attacker:       attacker,
		AttackerTile:   attackerTile,
		AttackerHealth: attackerHealth,
		Defender:       defender,
		DefenderTile:   defenderTile,
		DefenderHealth: defenderHealth,
		WoundBonus:     woundBonus,
	}

	// Calculate hit probability and show formula breakdown
	p, err := rulesEngine.CalculateHitProbability(attackerCtx)
	if err == nil {
		sb.WriteString("\nAttacker -> Defender (Formula):\n")

		// Get attack value from attack_vs_class table
		attackKey := fmt.Sprintf("%s:%s", defenderData.UnitClass, defenderData.UnitTerrain)
		baseAttack, hasAttack := attackerData.AttackVsClass[attackKey]

		if hasAttack {
			// Get terrain bonuses
			attackerTerrainProps := rulesEngine.GetTerrainUnitPropertiesForUnit(attackerTile.TileType, attacker.UnitType)
			defenderTerrainProps := rulesEngine.GetTerrainUnitPropertiesForUnit(defenderTile.TileType, defender.UnitType)

			var attackBonus, defenseBonus int32
			if attackerTerrainProps != nil {
				attackBonus = attackerTerrainProps.AttackBonus
			}
			if defenderTerrainProps != nil {
				defenseBonus = defenderTerrainProps.DefenseBonus
			}

			// Show formula breakdown
			sb.WriteString("  Formula: p = 0.05 * (((A + Ta) - (D + Td)) + B) + 0.5\n")
			sb.WriteString(fmt.Sprintf("    A (base attack vs %s:%s): %d\n", defenderData.UnitClass, defenderData.UnitTerrain, baseAttack))
			sb.WriteString(fmt.Sprintf("    Ta (terrain attack bonus): %d\n", attackBonus))
			sb.WriteString(fmt.Sprintf("    D (base defense): %d\n", defenderData.Defense))
			sb.WriteString(fmt.Sprintf("    Td (terrain defense bonus): %d\n", defenseBonus))
			sb.WriteString(fmt.Sprintf("    B (wound bonus): %d", woundBonus))
			if woundBonus > 0 {
				sb.WriteString(fmt.Sprintf(" [%d previous attack(s)]\n", len(defender.AttackHistory)))
			} else {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("  Hit Probability (p): %.2f (%.0f%%)\n", p, p*100))

			// Generate damage distribution from simulations
			dist, err := rulesEngine.GenerateDamageDistribution(attackerCtx, 10000)
			if err == nil {
				sb.WriteString(fmt.Sprintf("  Expected Damage: %.1f HP (from 10,000 simulations)\n", dist.ExpectedDamage))
				sb.WriteString(fmt.Sprintf("  Damage Range: %.0f-%.0f HP\n", dist.MinDamage, dist.MaxDamage))

				// Show top damage probabilities
				if len(dist.Ranges) > 0 {
					sb.WriteString("  Most Likely Outcomes:\n")
					// Sort ranges by probability (descending) and show top 5
					type rangeProb struct {
						damage float64
						prob   float64
					}
					sorted := make([]rangeProb, len(dist.Ranges))
					for i, r := range dist.Ranges {
						sorted[i] = rangeProb{damage: r.MinValue, prob: r.Probability}
					}
					// Simple bubble sort (good enough for small arrays)
					for i := 0; i < len(sorted); i++ {
						for j := i + 1; j < len(sorted); j++ {
							if sorted[j].prob > sorted[i].prob {
								sorted[i], sorted[j] = sorted[j], sorted[i]
							}
						}
					}
					// Show top 5
					for i := 0; i < len(sorted) && i < 5; i++ {
						sb.WriteString(fmt.Sprintf("    %.0f HP: %.1f%%\n", sorted[i].damage, sorted[i].prob*100))
					}
				}

				// Compare with table-based prediction
				if canAttack && attackDist != nil {
					diff := dist.ExpectedDamage - attackDist.ExpectedDamage
					sb.WriteString(fmt.Sprintf("  Difference from table: %+.1f HP\n", diff))
				}
			} else {
				sb.WriteString(fmt.Sprintf("  Error generating distribution: %v\n", err))
			}
		} else {
			sb.WriteString(fmt.Sprintf("  Cannot attack %s:%s units\n", defenderData.UnitClass, defenderData.UnitTerrain))
		}
	} else {
		sb.WriteString(fmt.Sprintf("\n  Formula calculation error: %v\n", err))
	}

	// Counter-attack formula (no wound bonus)
	if canCounter && counterDist != nil {
		canReach, _ := rulesEngine.CanUnitAttackTarget(defender, attacker)
		if canReach {
			counterCtx := &services.CombatContext{
				Attacker:       defender,
				AttackerTile:   defenderTile,
				AttackerHealth: defenderHealth,
				Defender:       attacker,
				DefenderTile:   attackerTile,
				DefenderHealth: attackerHealth,
				WoundBonus:     0, // No wound bonus for counter-attacks
			}

			counterP, err := rulesEngine.CalculateHitProbability(counterCtx)
			if err == nil {
				sb.WriteString("\nDefender -> Attacker (Counter Formula):\n")
				sb.WriteString(fmt.Sprintf("  Hit Probability (p): %.2f (%.0f%%)\n", counterP, counterP*100))

				// Generate counter damage distribution
				counterDist, err := rulesEngine.GenerateDamageDistribution(counterCtx, 10000)
				if err == nil {
					sb.WriteString(fmt.Sprintf("  Expected Damage: %.1f HP (from 10,000 simulations)\n", counterDist.ExpectedDamage))
					sb.WriteString(fmt.Sprintf("  Damage Range: %.0f-%.0f HP\n", counterDist.MinDamage, counterDist.MaxDamage))

					// Compare with table
					if counterDist != nil {
						// Get original table distribution
						tableDist, _ := rulesEngine.GetCombatPrediction(defender.UnitType, attacker.UnitType)
						if tableDist != nil {
							diff := counterDist.ExpectedDamage - tableDist.ExpectedDamage
							sb.WriteString(fmt.Sprintf("  Difference from table: %+.1f HP\n", diff))
						}
					}
				}
			}
		}
	}

	sb.WriteString("\n" + strings.Repeat("=", 60) + "\n\n")

	return sb.String()
}
