package lib

import (
	"fmt"
	"strings"
	"time"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// copyUnit creates a deep copy of a unit with all fields
// This is used when recording unit states in WorldChange objects
func copyUnit(unit *v1.Unit) *v1.Unit {
	if unit == nil {
		return nil
	}

	// Deep copy attack history
	var attackHistory []*v1.AttackRecord
	if unit.AttackHistory != nil {
		attackHistory = make([]*v1.AttackRecord, len(unit.AttackHistory))
		for i, record := range unit.AttackHistory {
			attackHistory[i] = &v1.AttackRecord{
				Q:          record.Q,
				R:          record.R,
				IsRanged:   record.IsRanged,
				TurnNumber: record.TurnNumber,
			}
		}
	}

	return &v1.Unit{
		Q:                       unit.Q,
		R:                       unit.R,
		Player:                  unit.Player,
		UnitType:                unit.UnitType,
		Shortcut:                unit.Shortcut,
		AvailableHealth:         unit.AvailableHealth,
		DistanceLeft:            unit.DistanceLeft,
		LastActedTurn:           unit.LastActedTurn,
		LastToppedupTurn:        unit.LastToppedupTurn,
		AttacksReceivedThisTurn: unit.AttacksReceivedThisTurn,
		AttackHistory:           attackHistory,
		ProgressionStep:         unit.ProgressionStep,
		ChosenAlternative:       unit.ChosenAlternative,
		CaptureStartedTurn:      unit.CaptureStartedTurn,
	}
}

// ProcessMoves processes a set of moves in a transaction and returns a "log entry" of the changes as a result
func (g *Game) ProcessMoves(moves []*v1.GameMove) (err error) {
	for _, move := range moves {
		err := g.ProcessMove(move)
		if err != nil {
			return err
		}
	}
	return
}

// ProcessMove is the dispatcher for a move.
// The moves work is we submit a move to the game, it calls the correct move handler.
// Moves in a game are "known" so we can have a simple static dispatcher here.
// The move handler/processor update the Game state and also updates the action object
// indicating changes that were incurred as part of running the move. Note that
// since we are looking at "transactionality" in games we want to make sure all moves
// are first valid and ATOMIC and only then finally commit the changes for all the moves.
// For example we may have 3 moves where first two units are moved to a common location
// and then they attack another unit. Here if we treat it as a single unit attacking it
// will have different outcomes than a "combined" attack.
func (g *Game) ProcessMove(move *v1.GameMove) (err error) {
	if move.MoveType == nil {
		return fmt.Errorf("move type is nil")
	}
	move.IsPermanent = false
	move.SequenceNum = 0 // TODO: Set proper sequence number
	move.Changes = []*v1.WorldChange{}

	switch a := move.MoveType.(type) {
	case *v1.GameMove_MoveUnit:
		return g.ProcessMoveUnit(move, a.MoveUnit, false)
	case *v1.GameMove_AttackUnit:
		return g.ProcessAttackUnit(move, a.AttackUnit)
	case *v1.GameMove_BuildUnit:
		return g.ProcessBuildUnit(move, a.BuildUnit)
	case *v1.GameMove_CaptureBuilding:
		return g.ProcessCaptureBuilding(move, a.CaptureBuilding)
	case *v1.GameMove_HealUnit:
		return g.ProcessHealUnit(move, a.HealUnit)
	case *v1.GameMove_EndTurn:
		return g.ProcessEndTurn(move, a.EndTurn)
	default:
		return fmt.Errorf("unknown move type: %T", move.MoveType)
	}
}

// ProcessBuildUnit creates a new unit at the specified tile
func (g *Game) ProcessBuildUnit(move *v1.GameMove, action *v1.BuildUnitAction) (err error) {
	// Initialize the result object
	move.IsPermanent = true // Builds are permanent

	coord, err := g.FromPos(action.Pos)
	if err != nil {
		return fmt.Errorf("invalid build position: %w", err)
	}
	tile := g.World.TileAt(coord)
	if tile == nil {
		return fmt.Errorf("no tile at position %v", coord)
	}

	// Check if tile belongs to the current player
	if tile.Player != g.CurrentPlayer {
		return fmt.Errorf("tile at %v does not belong to player %d", coord, g.CurrentPlayer)
	}

	// Check if tile can build (must be a building that can produce units)
	terrainData, err := g.RulesEngine.GetTerrainData(tile.TileType)
	if err != nil || terrainData == nil {
		return fmt.Errorf("terrain data not found for tile type %d", tile.TileType)
	}

	if len(terrainData.BuildableUnitIds) == 0 {
		return fmt.Errorf("tile at %v cannot build units", coord)
	}

	// Check if the requested unit type can be built at this tile
	canBuild := false
	for _, buildableID := range terrainData.BuildableUnitIds {
		if buildableID == action.UnitType {
			canBuild = true
			break
		}
	}
	if !canBuild {
		return fmt.Errorf("tile at %v cannot build unit type %d", coord, action.UnitType)
	}

	// Check if unit type is allowed by game settings
	allowedUnits := g.Config.Settings.GetAllowedUnits()
	if allowedUnits != nil {
		// If allowedUnits is set (even if empty), enforce the restriction
		isAllowed := false
		for _, allowedID := range allowedUnits {
			if allowedID == action.UnitType {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			return fmt.Errorf("unit type %d is not allowed in this game", action.UnitType)
		}
	}

	// Check if tile has already built this turn (one build per turn per tile)
	if tile.LastActedTurn == g.TurnCounter {
		return fmt.Errorf("tile at %v has already built a unit this turn", coord)
	}

	// Check if there's already a unit at this position
	existingUnit := g.World.UnitAt(coord)
	if existingUnit != nil {
		return fmt.Errorf("cannot build unit at %v: position already occupied by unit %s", coord, existingUnit.Shortcut)
	}

	// Get unit definition for cost validation
	unitData, err := g.RulesEngine.GetUnitData(action.UnitType)
	if err != nil || unitData == nil {
		return fmt.Errorf("unit data not found for unit type %d", action.UnitType)
	}

	// Check if player has enough coins (from GameState.PlayerStates)
	playerState := g.GameState.PlayerStates[g.CurrentPlayer]
	if playerState == nil {
		return fmt.Errorf("player state not found for player %d", g.CurrentPlayer)
	}
	playerCoins := playerState.Coins

	if playerCoins < unitData.Coins {
		return fmt.Errorf("insufficient coins: need %d, have %d", unitData.Coins, playerCoins)
	}

	// Deduct coins from player (in GameState.PlayerStates)
	playerState.Coins -= unitData.Coins

	// Generate a shortcut for the new unit
	newShortcut := g.World.GenerateUnitShortcut(g.CurrentPlayer)

	// Create the new unit
	newUnit := &v1.Unit{
		Q:                int32(coord.Q),
		R:                int32(coord.R),
		Player:           g.CurrentPlayer,
		UnitType:         action.UnitType,
		Shortcut:         newShortcut,
		AvailableHealth:  unitData.Health,
		DistanceLeft:     0, // Newly built units cannot move this turn
		LastActedTurn:    g.TurnCounter,
		LastToppedupTurn: g.TurnCounter,
		ProgressionStep:  1, // Start at step 1 (already used build action)
	}

	// Add unit to the world
	g.World.AddUnit(newUnit)

	// Mark tile as having acted this turn
	tile.LastActedTurn = g.TurnCounter

	// Update timestamp
	g.GameState.UpdatedAt = tspb.New(time.Now())

	// Record the build in world changes
	buildChange := &v1.WorldChange{
		ChangeType: &v1.WorldChange_UnitBuilt{
			UnitBuilt: &v1.UnitBuiltChange{
				Unit:        copyUnit(newUnit),
				TileQ:       tile.Q,
				TileR:       tile.R,
				CoinsCost:   unitData.Coins,
				PlayerCoins: playerCoins - unitData.Coins,
			},
		},
	}
	move.Changes = append(move.Changes, buildChange)

	// Record the coin deduction
	coinsChange := &v1.WorldChange{
		ChangeType: &v1.WorldChange_CoinsChanged{
			CoinsChanged: &v1.CoinsChangedChange{
				PlayerId:      g.CurrentPlayer,
				PreviousCoins: playerCoins,
				NewCoins:      playerCoins - unitData.Coins,
				Reason:        "build",
			},
		},
	}
	move.Changes = append(move.Changes, coinsChange)

	return nil
}

// ProcessCaptureBuilding starts capturing a building with a unit.
// The capture completes at the start of the capturing player's next turn
// if the unit survives until then.
func (g *Game) ProcessCaptureBuilding(move *v1.GameMove, action *v1.CaptureBuildingAction) (err error) {
	coord, err := g.FromPos(action.Pos)
	if err != nil {
		return fmt.Errorf("invalid capture position: %w", err)
	}

	// Get the unit at the position
	unit := g.World.UnitAt(coord)
	if unit == nil {
		return fmt.Errorf("no unit at position %v", coord)
	}

	// Check if it's the correct player's turn
	if unit.Player != g.CurrentPlayer {
		return fmt.Errorf("unit does not belong to current player %d", g.CurrentPlayer)
	}

	// Apply lazy top-up pattern
	if err := g.TopUpUnitIfNeeded(unit); err != nil {
		return fmt.Errorf("failed to top-up unit: %w", err)
	}

	// Get the tile at the position
	tile := g.World.TileAt(coord)
	if tile == nil {
		return fmt.Errorf("no tile at position %v", coord)
	}

	// Check if tile is already owned by the capturing player
	if tile.Player == g.CurrentPlayer {
		return fmt.Errorf("tile at %v is already owned by player %d", coord, g.CurrentPlayer)
	}

	// Check if this unit type can capture
	terrainProps := g.RulesEngine.GetTerrainUnitPropertiesForUnit(tile.TileType, unit.UnitType)
	if terrainProps == nil || !terrainProps.CanCapture {
		return fmt.Errorf("unit type %d cannot capture tile type %d", unit.UnitType, tile.TileType)
	}

	// Check if unit is already capturing
	if unit.CaptureStartedTurn > 0 {
		return fmt.Errorf("unit is already capturing a building")
	}

	// Capture previous unit state
	previousUnit := copyUnit(unit)

	// Start the capture
	unit.CaptureStartedTurn = g.TurnCounter

	// Update progression: record chosen alternative and advance step
	unitDef, err := g.RulesEngine.GetUnitData(unit.UnitType)
	if err == nil && unitDef != nil {
		actionOrder := unitDef.ActionOrder
		if len(actionOrder) == 0 {
			actionOrder = []string{"move", "attack|capture"}
		}

		// If current step has pipe-separated alternatives, record the choice
		if int(unit.ProgressionStep) < len(actionOrder) {
			stepActions := actionOrder[unit.ProgressionStep]
			if strings.Contains(stepActions, "|") {
				unit.ChosenAlternative = "capture"
			}
		}

		// Advance to next step (capture action consumed)
		unit.ProgressionStep++
		unit.ChosenAlternative = "" // Clear for next step
	}

	// Update timestamp
	g.GameState.UpdatedAt = tspb.New(time.Now())

	// Capture updated unit state
	updatedUnit := copyUnit(unit)

	// Record the capture start in world changes
	captureChange := &v1.WorldChange{
		ChangeType: &v1.WorldChange_CaptureStarted{
			CaptureStarted: &v1.CaptureStartedChange{
				CapturingUnit: updatedUnit,
				TileQ:         tile.Q,
				TileR:         tile.R,
				TileType:      tile.TileType,
				CurrentOwner:  tile.Player,
			},
		},
	}
	move.Changes = append(move.Changes, captureChange)

	// Also record unit state change for UI updates
	unitChange := &v1.WorldChange{
		ChangeType: &v1.WorldChange_UnitMoved{
			UnitMoved: &v1.UnitMovedChange{
				PreviousUnit: previousUnit,
				UpdatedUnit:  updatedUnit,
			},
		},
	}
	move.Changes = append(move.Changes, unitChange)

	return nil
}

// ProcessHealUnit executes manual healing for a unit
func (g *Game) ProcessHealUnit(move *v1.GameMove, action *v1.HealUnitAction) (err error) {
	// Parse position
	coord, err := g.FromPos(action.Pos)
	if err != nil {
		return fmt.Errorf("invalid position: %w", err)
	}

	// Get unit at position
	unit := g.World.UnitAt(coord)
	if unit == nil {
		return fmt.Errorf("no unit at position %v", coord)
	}

	// Verify unit belongs to current player
	if unit.Player != g.CurrentPlayer {
		return fmt.Errorf("unit belongs to player %d, not current player %d", unit.Player, g.CurrentPlayer)
	}

	// Get unit definition
	unitData, err := g.RulesEngine.GetUnitData(unit.UnitType)
	if err != nil {
		return fmt.Errorf("failed to get unit data: %w", err)
	}

	// Check if unit is already at max health
	if unit.AvailableHealth >= unitData.Health {
		return fmt.Errorf("unit already at max health")
	}

	// Calculate heal amount (use the value from action, or calculate if not provided)
	healAmount := action.HealAmount
	if healAmount <= 0 {
		healAmount = g.calculateHealAmount(unit, unitData)
		if healAmount <= 0 {
			return fmt.Errorf("unit cannot heal on this terrain")
		}
	}

	// Capture previous state
	previousUnit := copyUnit(unit)

	// Apply healing
	newHealth := unit.AvailableHealth + healAmount
	if newHealth > unitData.Health {
		newHealth = unitData.Health
	}
	unit.AvailableHealth = newHealth

	// Mark unit as having acted this turn (so it doesn't auto-heal next turn)
	unit.LastActedTurn = g.TurnCounter

	// Advance progression (healing counts as an action)
	unit.ProgressionStep++
	unit.ChosenAlternative = ""

	// Capture updated state
	updatedUnit := copyUnit(unit)

	// Record the change
	healChange := &v1.WorldChange{
		ChangeType: &v1.WorldChange_UnitHealed{
			UnitHealed: &v1.UnitHealedChange{
				PreviousUnit: previousUnit,
				UpdatedUnit:  updatedUnit,
				HealAmount:   healAmount,
			},
		},
	}
	move.Changes = append(move.Changes, healChange)

	return nil
}

// ProcessEndTurn advances to next player's turn.
// For now a player can just end turn but in other games there may be some mandatory
// moves left.
func (g *Game) ProcessEndTurn(move *v1.GameMove, action *v1.EndTurnAction) (err error) {
	// Store previous state for GameLog
	// TODO - use a pushed world at ProcessMoves level instead of g.World each time
	previousPlayer := g.CurrentPlayer
	previousTurn := g.TurnCounter

	// Calculate income for ending player based on bases owned and their types
	// Use IncomeConfig from game configuration if available
	var incomeConfig *v1.IncomeConfig
	if g.Config != nil {
		incomeConfig = g.Config.IncomeConfigs
	}

	totalIncome := int32(0)
	for _, tile := range g.World.TilesByCoord() {
		if tile.Player == previousPlayer {
			// Get income for this tile type from IncomeConfig (falls back to defaults)
			tileIncome := GetTileIncomeFromConfig(tile.TileType, incomeConfig)
			totalIncome += tileIncome
		}
	}

	// Add base game income (income just for being in the game)
	if incomeConfig != nil && incomeConfig.GameIncome > 0 {
		totalIncome += incomeConfig.GameIncome
	}

	// Get player's current coins (from GameState.PlayerStates)
	playerState := g.GameState.PlayerStates[previousPlayer]
	if playerState == nil {
		return fmt.Errorf("player state not found for player %d", previousPlayer)
	}
	playerCoins := playerState.Coins

	// Calculate and add income
	income := totalIncome
	newCoins := playerCoins + income

	// Update player's coins (in GameState.PlayerStates)
	playerState.Coins = newCoins

	// Record the income change
	if income > 0 {
		coinsChange := &v1.WorldChange{
			ChangeType: &v1.WorldChange_CoinsChanged{
				CoinsChanged: &v1.CoinsChangedChange{
					PlayerId:      previousPlayer,
					PreviousCoins: playerCoins,
					NewCoins:      newCoins,
					Reason:        "income",
				},
			},
		}
		move.Changes = append(move.Changes, coinsChange)
	}

	// Advance to next player (1-based player system: Player 1, Player 2, etc.)
	// Player 0 is reserved for neutral, so we cycle between 1, 2, ..., PlayerCount
	// Use configured player count from game config, not from World (which counts units)
	numPlayers := g.NumPlayers()

	if g.CurrentPlayer == numPlayers {
		// Last player completes their turn, go back to player 1 and increment turn counter
		g.CurrentPlayer = 1
		g.TurnCounter++
	} else {
		// Move to next player
		g.CurrentPlayer++
	}

	// Top-up the INCOMING player's units and capture them as ResetUnits
	// This ensures remote clients receive the refreshed values
	incomingPlayerUnits := g.World.GetPlayerUnits(int(g.CurrentPlayer))
	resetUnits := make([]*v1.Unit, 0, len(incomingPlayerUnits))

	for _, unit := range incomingPlayerUnits {
		// Top-up the unit (restores movement, applies healing, resets progression)
		if err := g.TopUpUnitIfNeeded(unit); err != nil {
			fmt.Printf("ProcessEndTurn: Warning - failed to top-up unit at (%d,%d): %v\n",
				unit.Q, unit.R, err)
		}
		fmt.Printf("ProcessEndTurn: Adding resetUnit at (%d, %d) player=%d, distanceLeft=%f\n",
			unit.Q, unit.R, unit.Player, unit.DistanceLeft)
		resetUnit := copyUnit(unit)
		resetUnits = append(resetUnits, resetUnit)
	}

	// Check for victory conditions
	if winner, hasWinner := g.checkVictoryConditions(); hasWinner {
		g.GameState.WinningPlayer = winner
		g.GameState.Finished = true
		g.GameState.Status = v1.GameStatus_GAME_STATUS_ENDED

		// Update GameLog status when game ends
		// TODO - g.SetGameLogStatus("completed")
	}

	// Update timestamp
	g.GameState.UpdatedAt = tspb.New(time.Now())
	change := &v1.WorldChange{
		ChangeType: &v1.WorldChange_PlayerChanged{
			PlayerChanged: &v1.PlayerChangedChange{
				PreviousPlayer: int32(previousPlayer),
				NewPlayer:      int32(g.CurrentPlayer),
				PreviousTurn:   int32(previousTurn),
				NewTurn:        int32(g.TurnCounter),
				ResetUnits:     resetUnits,
			},
		},
	}

	move.Changes = append(move.Changes, change)

	return
}

// ProcessMoveUnit executes unit movement using cube coordinates
func (g *Game) ProcessMoveUnit(move *v1.GameMove, action *v1.MoveUnitAction, preventPassThrough bool) (err error) {
	// Initialize the result object

	from, err := g.FromPos(action.From)
	if err != nil {
		return fmt.Errorf("invalid from position: %w", err)
	}
	// Parse 'to' relative to 'from' to support directions like "L", "TR"
	to, err := g.FromPosWithBase(action.To, &from)
	if err != nil {
		return fmt.Errorf("invalid to position: %w", err)
	}
	unit := g.World.UnitAt(from)
	if unit == nil {
		return fmt.Errorf("unit is nil")
	}

	// Apply lazy top-up pattern - ensure unit has current turn's movement points
	if err := g.TopUpUnitIfNeeded(unit); err != nil {
		return fmt.Errorf("failed to top-up unit: %w", err)
	}

	// Check if it's the correct player's turn
	if unit.Player != g.CurrentPlayer {
		return fmt.Errorf("not player %d's turn", unit.Player)
	}

	// Find path to destination (validates move and returns path for animation)
	path, cost, err := g.RulesEngine.FindPathTo(unit, to, g.World, preventPassThrough)
	if err != nil {
		unitCoord := UnitGetCoord(unit)
		return fmt.Errorf("invalid move from %v to %v: %w", unitCoord, to, err)
	}

	// Store the reconstructed path in the action for animation purposes
	action.ReconstructedPath = path

	// Capture unit state before move
	previousUnit := copyUnit(unit)

	// Move unit using World unit management
	err = g.World.MoveUnit(unit, to)
	if err != nil {
		return fmt.Errorf("failed to move unit: %w", err)
	}

	// Get the moved unit from the world (handles copy-on-write correctly)
	movedUnit := g.World.UnitAt(to)
	if movedUnit == nil {
		return fmt.Errorf("moved unit not found at destination %v", to)
	}

	// Update unit stats on the moved unit
	movedUnit.DistanceLeft -= cost

	// Update progression: if distance_left reaches 0, advance to next step
	if movedUnit.DistanceLeft <= 0 {
		movedUnit.ProgressionStep++
		movedUnit.ChosenAlternative = "" // Clear for next step
	}

	// Capture unit state after move (using the moved unit, not the original)
	updatedUnit := copyUnit(movedUnit)
	updatedUnit.LastActedTurn = unit.LastActedTurn
	updatedUnit.LastToppedupTurn = unit.LastToppedupTurn

	// Update timestamp
	g.GameState.UpdatedAt = tspb.New(time.Now())

	// Record action in GameLog
	change := &v1.WorldChange{
		ChangeType: &v1.WorldChange_UnitMoved{
			UnitMoved: &v1.UnitMovedChange{
				PreviousUnit: previousUnit,
				UpdatedUnit:  updatedUnit,
			},
		},
	}

	move.Changes = append(move.Changes, change)
	return nil
}

// ProcessAttackUnit executes combat between units
func (g *Game) ProcessAttackUnit(move *v1.GameMove, action *v1.AttackUnitAction) (err error) {
	// Initialize the result object
	move.IsPermanent = true // Attacks are permanent (cannot be undone)

	attackerCoord, err := g.FromPos(action.Attacker)
	if err != nil {
		return fmt.Errorf("invalid attacker position: %w", err)
	}
	defenderCoord, err := g.FromPos(action.Defender)
	if err != nil {
		return fmt.Errorf("invalid defender position: %w", err)
	}
	attacker := g.World.UnitAt(attackerCoord)
	defender := g.World.UnitAt(defenderCoord)
	if attacker == nil || defender == nil {
		return fmt.Errorf("attacker or defender is nil")
	}

	// Apply lazy top-up pattern for both units
	if err := g.TopUpUnitIfNeeded(attacker); err != nil {
		return fmt.Errorf("failed to top-up attacker: %w", err)
	}
	if err := g.TopUpUnitIfNeeded(defender); err != nil {
		return fmt.Errorf("failed to top-up defender: %w", err)
	}

	// Check if it's the correct player's turn
	if attacker.Player != g.CurrentPlayer {
		return fmt.Errorf("not player %d's turn", attacker.Player)
	}

	// Check if units can attack each other
	if !g.CanAttackUnit(attacker, defender) {
		return fmt.Errorf("attacker cannot attack defender")
	}

	// Store original health for world changes
	attackerOriginalHealth := attacker.AvailableHealth
	defenderOriginalHealth := defender.AvailableHealth

	// Calculate wound bonus from defender's attack history
	woundBonus := g.RulesEngine.CalculateWoundBonus(defender, attackerCoord)

	// Create combat context for attacker -> defender
	attackerCtx := &CombatContext{
		Attacker:       attacker,
		AttackerTile:   g.World.TileAt(attackerCoord),
		AttackerHealth: attacker.AvailableHealth,
		Defender:       defender,
		DefenderTile:   g.World.TileAt(defenderCoord),
		DefenderHealth: defender.AvailableHealth,
		WoundBonus:     woundBonus,
	}

	// Calculate damage using formula-based system
	defenderDamage, err := g.RulesEngine.SimulateCombatDamage(attackerCtx, g.rng)
	if err != nil {
		return fmt.Errorf("failed to calculate combat damage: %w", err)
	}

	// Check if defender can counter-attack
	attackerDamage := int32(0)
	if canCounter, err := g.RulesEngine.CanUnitAttackTarget(defender, attacker); err == nil && canCounter {
		// Create combat context for counter-attack (no wound bonus)
		counterCtx := &CombatContext{
			Attacker:       defender,
			AttackerTile:   g.World.TileAt(defenderCoord),
			AttackerHealth: defender.AvailableHealth,
			Defender:       attacker,
			DefenderTile:   g.World.TileAt(attackerCoord),
			DefenderHealth: attacker.AvailableHealth,
			WoundBonus:     0, // No wound bonus for counter-attacks
		}

		attackerDamage, err = g.RulesEngine.SimulateCombatDamage(counterCtx, g.rng)
		if err != nil {
			// If counter-attack calculation fails, no counter damage
			attackerDamage = 0
		}
	}

	// Update progression: record chosen alternative and check if step is complete
	unitDef, err := g.RulesEngine.GetUnitData(attacker.UnitType)
	if err == nil && unitDef != nil {
		// Get action_order for this unit
		actionOrder := unitDef.ActionOrder
		if len(actionOrder) == 0 {
			actionOrder = []string{"move", "attack|capture"}
		}

		// If current step has pipe-separated alternatives, record the choice
		if int(attacker.ProgressionStep) < len(actionOrder) {
			stepActions := actionOrder[attacker.ProgressionStep]
			if strings.Contains(stepActions, "|") {
				attacker.ChosenAlternative = "attack"
			}
		}

		// Check if we've reached the action limit for attacks
		// For now, assume 1 attack per step (can enhance later with action_limits)
		// Since attack was performed, advance to next step
		attacker.ProgressionStep++
		attacker.ChosenAlternative = "" // Clear for next step

		// If next step is "retreat", set DistanceLeft to retreat_points
		if int(attacker.ProgressionStep) < len(actionOrder) {
			nextStepAction := actionOrder[attacker.ProgressionStep]
			if nextStepAction == "retreat" {
				attacker.DistanceLeft = unitDef.RetreatPoints
			}
		}
	}

	// Record attack in defender's history for future wound bonus calculations
	distance := CubeDistance(attackerCoord, defenderCoord)
	defender.AttackHistory = append(defender.AttackHistory, &v1.AttackRecord{
		Q:          int32(attackerCoord.Q),
		R:          int32(attackerCoord.R),
		IsRanged:   distance >= 2,
		TurnNumber: g.TurnCounter,
	})
	defender.AttacksReceivedThisTurn++

	// Apply damage
	defender.AvailableHealth -= int32(defenderDamage)
	if defender.AvailableHealth < 0 {
		defender.AvailableHealth = 0
	}

	attacker.AvailableHealth -= int32(attackerDamage)
	if attacker.AvailableHealth < 0 {
		attacker.AvailableHealth = 0
	}

	// Check if units were killed
	defenderKilled := defender.AvailableHealth <= 0
	attackerKilled := attacker.AvailableHealth <= 0

	// Add damage changes to world changes
	if defenderDamage > 0 {
		// Capture defender state before damage
		defenderPreviousUnit := copyUnit(defender)
		defenderPreviousUnit.AvailableHealth = defenderOriginalHealth

		// Capture defender state after damage
		defenderUpdatedUnit := copyUnit(defender)

		change := &v1.WorldChange{
			ChangeType: &v1.WorldChange_UnitDamaged{
				UnitDamaged: &v1.UnitDamagedChange{
					PreviousUnit: defenderPreviousUnit,
					UpdatedUnit:  defenderUpdatedUnit,
				},
			},
		}
		move.Changes = append(move.Changes, change)
	}

	if attackerDamage > 0 {
		// Capture attacker state before damage
		attackerPreviousUnit := copyUnit(attacker)
		attackerPreviousUnit.AvailableHealth = attackerOriginalHealth

		// Capture attacker state after damage
		attackerUpdatedUnit := copyUnit(attacker)

		change := &v1.WorldChange{
			ChangeType: &v1.WorldChange_UnitDamaged{
				UnitDamaged: &v1.UnitDamagedChange{
					PreviousUnit: attackerPreviousUnit,
					UpdatedUnit:  attackerUpdatedUnit,
				},
			},
		}
		move.Changes = append(move.Changes, change)
	}

	// Add kill changes if units were killed
	if defenderKilled {
		// Capture defender state before being killed (use original health before damage)
		defenderPreviousUnit := copyUnit(defender)
		defenderPreviousUnit.AvailableHealth = defenderOriginalHealth

		change := &v1.WorldChange{
			ChangeType: &v1.WorldChange_UnitKilled{
				UnitKilled: &v1.UnitKilledChange{
					PreviousUnit: defenderPreviousUnit,
				},
			},
		}
		move.Changes = append(move.Changes, change)
		g.World.RemoveUnit(defender)
	}

	if attackerKilled {
		// Capture attacker state before being killed (use original health before damage)
		attackerPreviousUnit := copyUnit(attacker)
		attackerPreviousUnit.AvailableHealth = attackerOriginalHealth

		change := &v1.WorldChange{
			ChangeType: &v1.WorldChange_UnitKilled{
				UnitKilled: &v1.UnitKilledChange{
					PreviousUnit: attackerPreviousUnit,
				},
			},
		}
		move.Changes = append(move.Changes, change)
		g.World.RemoveUnit(attacker)
	}

	// Apply splash damage to adjacent units (if attacker has splash damage capability)
	// Only if attacker is still alive (not killed by counter-attack)
	if !attackerKilled {
		// Get all 6 adjacent hexes around the defender
		var adjacentCoords [6]AxialCoord
		defenderCoord.Neighbors(&adjacentCoords)

		adjacentUnits := make([]*v1.Unit, 0, 6) // Pre-allocate for max 6 neighbors
		for _, coord := range adjacentCoords {
			if unit := g.World.UnitAt(coord); unit != nil {
				// Include all units (friendly and enemy), air units will be filtered by CalculateSplashDamage
				adjacentUnits = append(adjacentUnits, unit)
			}
		}

		if len(adjacentUnits) > 0 {
			splashTargets, err := g.RulesEngine.CalculateSplashDamage(
				attacker,
				g.World.TileAt(attackerCoord),
				defenderCoord,
				adjacentUnits,
				g.World,
				g.rng,
			)
			if err == nil && len(splashTargets) > 0 {
				// Apply splash damage to each target
				for _, target := range splashTargets {
					// Store original health before splash damage
					targetOriginalHealth := target.Unit.AvailableHealth

					// Apply splash damage
					target.Unit.AvailableHealth -= target.Damage
					if target.Unit.AvailableHealth < 0 {
						target.Unit.AvailableHealth = 0
					}

					// Check if unit was killed by splash
					targetKilled := target.Unit.AvailableHealth <= 0

					// Add damage change
					targetPreviousUnit := copyUnit(target.Unit)
					targetPreviousUnit.AvailableHealth = targetOriginalHealth

					targetUpdatedUnit := copyUnit(target.Unit)

					change := &v1.WorldChange{
						ChangeType: &v1.WorldChange_UnitDamaged{
							UnitDamaged: &v1.UnitDamagedChange{
								PreviousUnit: targetPreviousUnit,
								UpdatedUnit:  targetUpdatedUnit,
							},
						},
					}
					move.Changes = append(move.Changes, change)

					// Add kill change if unit was killed by splash
					if targetKilled {
						killChange := &v1.WorldChange{
							ChangeType: &v1.WorldChange_UnitKilled{
								UnitKilled: &v1.UnitKilledChange{
									PreviousUnit: targetPreviousUnit,
								},
							},
						}
						move.Changes = append(move.Changes, killChange)
						g.World.RemoveUnit(target.Unit)
					}
				}
			}
		}
	}

	// Update timestamp
	g.GameState.UpdatedAt = tspb.New(time.Now())

	return nil
}

// CanMoveUnit validates potential movement using Dijkstra-based pathfinding
// This checks if the target is reachable given terrain costs and available movement points
func (g *Game) CanMoveUnit(unit *v1.Unit, to AxialCoord, preventPassThrough bool) bool {
	if unit == nil {
		return false
	}

	// Check if it's the correct player's turn
	if unit.Player != g.CurrentPlayer {
		return false
	}

	// Check if destination is occupied by another unit
	destUnit := g.World.UnitAt(to)
	if destUnit != nil {
		return false
	}

	// Use Dijkstra to compute all reachable tiles based on terrain and movement points
	allPaths, err := g.RulesEngine.GetMovementOptions(g.World, unit, int(unit.DistanceLeft), preventPassThrough)
	if err != nil {
		return false
	}

	// Check if target coordinate is in the reachable tiles
	key := fmt.Sprintf("%d,%d", to.Q, to.R)
	_, reachable := allPaths.Edges[key]
	return reachable
}

// CanAttackUnit validates potential attack
func (g *Game) CanAttackUnit(attacker, defender *v1.Unit) bool {
	if attacker == nil || defender == nil {
		return false
	}

	// Check if it's the correct player's turn
	if attacker.Player != g.CurrentPlayer {
		return false
	}

	// Check if units are enemies
	if attacker.Player == defender.Player {
		return false
	}

	// Use rules engine for attack validation
	canAttack, err := g.RulesEngine.CanUnitAttackTarget(attacker, defender)
	if err != nil {
		return false
	}
	return canAttack
}

// CanAttack validates potential attack using position coordinates
func (g *Game) CanAttack(from, to AxialCoord) (bool, error) {
	attacker := g.World.UnitAt(from)
	if attacker == nil {
		return false, fmt.Errorf("no unit at attacker position (%d, %d)", from.Q, from.R)
	}

	defender := g.World.UnitAt(to)
	if defender == nil {
		return false, fmt.Errorf("no unit at target position (%d, %d)", to.Q, to.R)
	}

	return g.CanAttackUnit(attacker, defender), nil
}

// GetMovementOptions returns movement options for unit at given coordinates with full validation
func (g *Game) GetMovementOptions(q, r int32, preventPassThrough bool) (*v1.AllPaths, error) {
	unit := g.World.UnitAt(AxialCoord{Q: int(q), R: int(r)})
	if unit == nil {
		return nil, fmt.Errorf("no unit found at position (%d, %d)", q, r)
	}
	if unit.Player != g.CurrentPlayer {
		return nil, fmt.Errorf("unit belongs to player %d, but it's player %d's turn", unit.Player, g.CurrentPlayer)
	}
	if unit.AvailableHealth <= 0 {
		return nil, fmt.Errorf("unit has no health remaining")
	}
	if unit.DistanceLeft <= 0 {
		return nil, fmt.Errorf("unit has no movement points remaining")
	}
	return g.RulesEngine.GetMovementOptions(g.World, unit, int(unit.DistanceLeft), preventPassThrough)
}

// GetAttackOptions returns attack options for unit at given coordinates with full validation
func (g *Game) GetAttackOptions(q, r int32) ([]AxialCoord, error) {
	unit := g.World.UnitAt(AxialCoord{Q: int(q), R: int(r)})
	if unit == nil {
		return nil, fmt.Errorf("no unit found at position (%d, %d)", q, r)
	}
	if unit.Player != g.CurrentPlayer {
		return nil, fmt.Errorf("unit belongs to player %d, but it's player %d's turn", unit.Player, g.CurrentPlayer)
	}
	if unit.AvailableHealth <= 0 {
		return nil, fmt.Errorf("unit has no health remaining")
	}
	return g.RulesEngine.GetAttackOptions(g.World, unit)
}

// CanSelectUnit validates if unit at given coordinates can be selected by current player
func (g *Game) CanSelectUnit(q, r int32) (bool, string) {
	unit := g.World.UnitAt(AxialCoord{Q: int(q), R: int(r)})
	if unit == nil {
		return false, fmt.Sprintf("no unit found at position (%d, %d)", q, r)
	}
	if unit.Player != g.CurrentPlayer {
		return false, fmt.Sprintf("unit belongs to player %d, but it's player %d's turn", unit.Player, g.CurrentPlayer)
	}
	if unit.AvailableHealth <= 0 {
		return false, "unit has no health remaining"
	}
	return true, ""
}

// CanMove validates potential movement using position coordinates
func (g *Game) CanMove(from, to Position, preventPassThrough bool) (bool, error) {
	unit := g.World.UnitAt(from)
	return g.CanMoveUnit(unit, to, preventPassThrough), nil
}

// GetUnitAttackOptions returns all positions a unit can attack using rules engine
func (g *Game) GetUnitAttackOptionsFrom(q, r int) ([]AxialCoord, error) {
	return g.GetUnitAttackOptions(g.World.UnitAt(AxialCoord{q, r}))
}
func (g *Game) GetUnitAttackOptions(unit *v1.Unit) ([]AxialCoord, error) {
	return g.RulesEngine.GetAttackOptions(g.World, unit)
}
