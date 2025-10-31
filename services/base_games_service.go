package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GamesServiceImpl interface {
	v1.GamesServiceServer
	GetRuntimeGame(game *v1.Game, gameState *v1.GameState) (*Game, error)
}

type BaseGamesServiceImpl struct {
	v1.UnimplementedGamesServiceServer
	Self GamesServiceImpl // The actual implementation
}

type WorldsServiceImpl interface {
	v1.WorldsServiceServer
}

type BaseWorldsServiceImpl struct {
	v1.UnimplementedWorldsServiceServer
	Self WorldsServiceImpl // The actual implementation
}

// ProcessMoves processes moves for an existing game on the wasm side.
// Unlike the service side games service - it wont persist any changes - it only will return the diffs.
func (s *BaseGamesServiceImpl) ProcessMoves(ctx context.Context, req *v1.ProcessMovesRequest) (resp *v1.ProcessMovesResponse, err error) {
	if len(req.Moves) == 0 {
		return nil, fmt.Errorf("at least one move is required")
	}

	gameresp, err := s.Self.GetGame(ctx, &v1.GetGameRequest{Id: req.GameId})
	if err != nil || gameresp.Game == nil {
		return nil, err
	}
	if gameresp.State == nil {
		panic("Game state cannot be nil")
	}
	if gameresp.History == nil {
		panic("Game history cannot cannot be nil")
	}

	// Get the runtime game corresponding to this game Id, we can create it on the fly
	// or we can cache it somewhere, or in the case of wasm just have a singleton
	rtGame, err := s.Self.GetRuntimeGame(gameresp.Game, gameresp.State)
	if err != nil {
		return nil, err
	}

	// TRANSACTIONAL FIX: Create transaction snapshot for move processing
	// ProcessMoves will operate on the snapshot, ApplyChangeResults will apply to original
	originalWorld := rtGame.World
	rtGame.World = originalWorld.Push() // Create transaction layer
	// Get the moves validted by the move processor, it is upto the move processor
	// to decide how "transactional" it wants to be - ie fail after  N moves,
	// success only if all moves succeeds etc.  Note that at this point the game
	// state has not changed and neither has the Runtime Game object.  Both the
	// GameState and the Runtime Game are checkpointed at before the moves started
	var dmp MoveProcessor
	results, err := dmp.ProcessMoves(rtGame, req.Moves)
	if err != nil {
		return nil, err
	}
	resp = &v1.ProcessMovesResponse{
		MoveResults: results,
	}

	// Create a new move group to track this batch of processed moves
	startTime := time.Now()
	moveGroup := &v1.GameMoveGroup{
		StartedAt:   timestamppb.New(startTime),
		EndedAt:     timestamppb.New(startTime), // TODO: Set proper end time after processing
		Moves:       req.Moves,
		MoveResults: results,
	}

	// Add the move group to history
	gameresp.History.Groups = append(gameresp.History.Groups, moveGroup)

	// Now that we have the results, we want to update our gamestate by applying the
	// results - this would also set the next "checkoint" to after the reuslts.
	// It is upto the storage to see how the runtime game is also updated.  For example
	// a storage that persists the gameState may just not do anythign and let it be
	// reconstructed on the next load
	s.ApplyChangeResults(results, rtGame, gameresp.Game, gameresp.State, gameresp.History)

	// Update the end time after processing is complete
	moveGroup.EndedAt = timestamppb.New(time.Now())

	// And then save it
	_, err = s.Self.UpdateGame(ctx, &v1.UpdateGameRequest{
		GameId:     req.GameId,
		NewGame:    gameresp.Game,
		NewState:   gameresp.State,
		NewHistory: gameresp.History,
	})

	return resp, err
}

// GetOptionsAt returns all available options at a specific position
func (s *BaseGamesServiceImpl) GetOptionsAt(ctx context.Context, req *v1.GetOptionsAtRequest) (*v1.GetOptionsAtResponse, error) {
	// Load game data using the service implementation
	gameresp, err := s.Self.GetGame(ctx, &v1.GetGameRequest{Id: req.GameId})
	if err != nil || gameresp.Game == nil {
		return &v1.GetOptionsAtResponse{
			Options:         []*v1.GameOption{},
			CurrentPlayer:   0,
			GameInitialized: false,
		}, nil
	}
	if gameresp.State == nil {
		return &v1.GetOptionsAtResponse{
			Options:         []*v1.GameOption{},
			CurrentPlayer:   0,
			GameInitialized: false,
		}, nil
	}

	// Get the runtime game
	rtGame, err := s.Self.GetRuntimeGame(gameresp.Game, gameresp.State)
	if err != nil {
		return &v1.GetOptionsAtResponse{
			Options:         []*v1.GameOption{},
			CurrentPlayer:   gameresp.State.CurrentPlayer,
			GameInitialized: false,
		}, nil
	}

	var options []*v1.GameOption
	var allPaths *v1.AllPaths

	// Check what's at this position
	unit := rtGame.World.UnitAt(AxialCoord{Q: int(req.Q), R: int(req.R)})

	// Lazy top-up: If this is a unit, ensure it's refreshed for the current turn
	if unit != nil {
		if err := rtGame.TopUpUnitIfNeeded(unit); err != nil {
			return &v1.GetOptionsAtResponse{
				Options:         []*v1.GameOption{},
				CurrentPlayer:   gameresp.State.CurrentPlayer,
				GameInitialized: false,
			}, fmt.Errorf("failed to top-up unit: %w", err)
		}
	}

	// Check if there's a tile at this position and get its actions
	tile := rtGame.World.TileAt(AxialCoord{Q: int(req.Q), R: int(req.R)})
	if tile != nil {
		// Lazy top-up: Ensure tile is refreshed for the current turn
		if err := rtGame.TopUpTileIfNeeded(tile); err != nil {
			return &v1.GetOptionsAtResponse{
				Options:         []*v1.GameOption{},
				CurrentPlayer:   gameresp.State.CurrentPlayer,
				GameInitialized: false,
			}, fmt.Errorf("failed to top-up tile: %w", err)
		}

		// Only check tile actions if tile belongs to current player
		if tile.Player == rtGame.CurrentPlayer {
			// Get terrain definition for tile-specific actions
			terrainDef, err := rtGame.rulesEngine.GetTerrainData(tile.TileType)
			if err == nil {
				// Get current player's coins (default to 100 for now, will be from game config later)
				// TODO: Get actual player coins from rtGame.Game.Config.Players[currentPlayer].Coins
				playerCoins := int32(100)

				// Get allowed actions for this tile
				tileActions := rtGame.rulesEngine.GetAllowedActionsForTile(tile, terrainDef, playerCoins)

				// Generate options based on allowed tile actions
				for _, action := range tileActions {
					switch action {
					case "build":
						// Generate build unit options from terrainDef.BuildableUnitIds
						for _, unitTypeID := range terrainDef.BuildableUnitIds {
							// Get unit definition to retrieve cost
							unitDef, err := rtGame.rulesEngine.GetUnitData(unitTypeID)
							if err != nil {
								continue // Skip if we can't get unit definition
							}

							// Only show units the player can afford
							if unitDef.Coins <= playerCoins {
								options = append(options, &v1.GameOption{
									OptionType: &v1.GameOption_Build{
										Build: &v1.BuildUnitOption{
											Q:        req.Q,
											R:        req.R,
											UnitType: unitTypeID,
											Cost:     unitDef.Coins,
										},
									},
								})
							}
						}
					}
				}
			}
		}
	}

	if unit != nil {
		// Our unit - get available options based on action progression
		var dmp MoveProcessor

		// Get unit definition for progression rules
		unitDef, err := rtGame.rulesEngine.GetUnitData(unit.UnitType)
		if err != nil {
			// If we can't get unit def, default to all actions
			unitDef = &v1.UnitDefinition{
				ActionOrder: []string{"move", "attack|capture"},
			}
		}

		// Get allowed actions based on progression state
		allowedActions := rtGame.rulesEngine.GetAllowedActionsForUnit(unit, unitDef)

		// If no actions allowed (progression complete), only end turn available
		if len(allowedActions) == 0 {
			options = append(options, &v1.GameOption{
				OptionType: &v1.GameOption_EndTurn{
					EndTurn: &v1.EndTurnOption{},
				},
			})
		} else {
			// Check if "move" is allowed at current progression step
			moveAllowed := containsAction(allowedActions, "move")

			// Get movement options if unit has movement left and move is allowed
			if unit.AvailableHealth > 0 && unit.DistanceLeft > 0 && moveAllowed {
				pathsResult, err := dmp.GetMovementOptions(rtGame, req.Q, req.R)
				if err == nil {
					allPaths = pathsResult

					// Create move options from AllPaths
					for key, edge := range allPaths.Edges {
						// Create ready-to-use MoveUnitAction
						moveAction := &v1.MoveUnitAction{
							FromQ: req.Q,
							FromR: req.R,
							ToQ:   edge.ToQ,
							ToR:   edge.ToR,
						}

						path, err := ReconstructPath(allPaths, edge.ToQ, edge.ToR)
						if err != nil {
							panic(err)
						}
						options = append(options, &v1.GameOption{
							OptionType: &v1.GameOption_Move{
								Move: &v1.MoveOption{
									MovementCost:      edge.TotalCost,
									Action:            moveAction,
									ReconstructedPath: path,
								},
							},
						})
						_ = key // Using key just to avoid unused variable warning
					}
				}
			}

			// Check if "attack" is allowed at current progression step
			attackAllowed := containsAction(allowedActions, "attack")

			// Get attack options if unit can attack and attack is allowed
			if unit.AvailableHealth > 0 && attackAllowed {
				attackCoords, err := dmp.GetAttackOptions(rtGame, req.Q, req.R)
				if err == nil {
					for _, coord := range attackCoords {
						// Get target unit info for rich attack option data
						targetUnit := rtGame.World.UnitAt(coord)
						if targetUnit != nil {
							// Calculate estimated damage (simplified for now)
							damageEstimate := int32(50) // TODO: Use proper damage calculation from rules engine

							// Create ready-to-use AttackUnitAction
							attackAction := &v1.AttackUnitAction{
								AttackerQ: req.Q,
								AttackerR: req.R,
								DefenderQ: int32(coord.Q),
								DefenderR: int32(coord.R),
							}

							options = append(options, &v1.GameOption{
								OptionType: &v1.GameOption_Attack{
									Attack: &v1.AttackOption{
										TargetUnitType:   targetUnit.UnitType,
										TargetUnitHealth: targetUnit.AvailableHealth,
										CanAttack:        true,
										DamageEstimate:   damageEstimate,
										Action:           attackAction,
									},
								},
							})
						}
					}
				}
			}

			// TODO: Add capture building options if "capture" is allowed
			// TODO: Add build unit options if "build" is allowed
		}

		// Only add the endturn option if unit belongs to current player
		if unit.Player == rtGame.CurrentPlayer {
			options = append(options, &v1.GameOption{
				OptionType: &v1.GameOption_EndTurn{
					EndTurn: &v1.EndTurnOption{},
				},
			})
		}
	} else if tile != nil {
		// No unit present - show end turn only if:
		// - tile is not owned (player == 0), OR
		// - tile is owned by current player
		if tile.Player == 0 || tile.Player == rtGame.CurrentPlayer {
			options = append(options, &v1.GameOption{
				OptionType: &v1.GameOption_EndTurn{
					EndTurn: &v1.EndTurnOption{},
				},
			})
		}
	}

	// Sort it for convinience too
	sort.Slice(options, func(i, j int) bool {
		return GameOptionLess(options[i], options[j])
	})

	return &v1.GetOptionsAtResponse{
		Options:         options,
		CurrentPlayer:   rtGame.CurrentPlayer,
		GameInitialized: rtGame != nil && rtGame.World != nil,
		AllPaths:        allPaths,
	}, nil
}

func (b *BaseGamesServiceImpl) ApplyChangeResults(changes []*v1.GameMoveResult, rtGame *Game, game *v1.Game, state *v1.GameState, history *v1.GameMoveHistory) error {

	// TRANSACTIONAL FIX: Temporary rollback to original world for ordered application
	if parent := rtGame.World.Pop(); parent != nil {
		rtGame.World = parent // Switch back to original world
	}

	// Apply each change to runtime game (now the original, not the transaction snapshot)
	for _, moveResult := range changes {
		for _, change := range moveResult.Changes {
			err := b.applyWorldChange(change, rtGame, state)
			if err != nil {
				return fmt.Errorf("failed to apply world change: %w", err)
			}
		}
	}

	state.WorldData = b.convertRuntimeWorldToProto(rtGame.World)
	state.UpdatedAt = timestamppb.New(time.Now())

	return nil
}

// applyWorldChange applies a single WorldChange to both runtime game and protobuf state
func (b *BaseGamesServiceImpl) applyWorldChange(change *v1.WorldChange, rtGame *Game, state *v1.GameState) error {
	switch changeType := change.ChangeType.(type) {
	case *v1.WorldChange_UnitMoved:
		return b.applyUnitMoved(changeType.UnitMoved, rtGame)
	case *v1.WorldChange_UnitDamaged:
		return b.applyUnitDamaged(changeType.UnitDamaged, rtGame)
	case *v1.WorldChange_UnitKilled:
		return b.applyUnitKilled(changeType.UnitKilled, rtGame)
	case *v1.WorldChange_PlayerChanged:
		return b.applyPlayerChanged(changeType.PlayerChanged, rtGame, state)
	default:
		return fmt.Errorf("unknown world change type")
	}
}

// applyUnitMoved moves a unit in the runtime game
func (b *BaseGamesServiceImpl) applyUnitMoved(change *v1.UnitMovedChange, rtGame *Game) error {
	if change.PreviousUnit == nil || change.UpdatedUnit == nil {
		return fmt.Errorf("missing unit data in UnitMovedChange")
	}

	fromCoord := AxialCoord{Q: int(change.PreviousUnit.Q), R: int(change.PreviousUnit.R)}
	toCoord := AxialCoord{Q: int(change.UpdatedUnit.Q), R: int(change.UpdatedUnit.R)}

	// Move unit in runtime game
	unit := rtGame.World.UnitAt(fromCoord)
	if unit == nil {
		return fmt.Errorf("unit not found at %v", fromCoord)
	}

	// Update unit with complete state from the change
	unit.AvailableHealth = change.UpdatedUnit.AvailableHealth
	unit.DistanceLeft = change.UpdatedUnit.DistanceLeft
	unit.LastActedTurn = change.UpdatedUnit.LastActedTurn
	unit.LastToppedupTurn = change.UpdatedUnit.LastToppedupTurn

	// Remove from old position and add to new position
	return rtGame.World.MoveUnit(unit, toCoord)
}

// applyUnitDamaged updates unit health in the runtime game
func (b *BaseGamesServiceImpl) applyUnitDamaged(change *v1.UnitDamagedChange, rtGame *Game) error {
	if change.UpdatedUnit == nil {
		return fmt.Errorf("missing updated unit data in UnitDamagedChange")
	}

	coord := AxialCoord{Q: int(change.UpdatedUnit.Q), R: int(change.UpdatedUnit.R)}

	unit := rtGame.World.UnitAt(coord)
	if unit == nil {
		return fmt.Errorf("unit not found at %v", coord)
	}

	// Update unit with complete state from the change
	unit.AvailableHealth = change.UpdatedUnit.AvailableHealth
	unit.DistanceLeft = change.UpdatedUnit.DistanceLeft
	unit.LastActedTurn = change.UpdatedUnit.LastActedTurn
	unit.LastToppedupTurn = change.UpdatedUnit.LastToppedupTurn
	return nil
}

// applyUnitKilled removes a unit from the runtime game
func (b *BaseGamesServiceImpl) applyUnitKilled(change *v1.UnitKilledChange, rtGame *Game) error {
	if change.PreviousUnit == nil {
		return fmt.Errorf("missing previous unit data in UnitKilledChange")
	}

	coord := AxialCoord{Q: int(change.PreviousUnit.Q), R: int(change.PreviousUnit.R)}
	unit := rtGame.World.UnitAt(coord)

	err := rtGame.World.RemoveUnit(unit)
	if err != nil {
		return fmt.Errorf("unit not found at %v", coord)
	}
	return nil
}

// applyPlayerChanged updates game state for turn/player changes
func (b *BaseGamesServiceImpl) applyPlayerChanged(change *v1.PlayerChangedChange, rtGame *Game, state *v1.GameState) error {
	rtGame.CurrentPlayer = change.NewPlayer
	rtGame.TurnCounter = change.NewTurn

	// Also update the protobuf GameState
	state.CurrentPlayer = change.NewPlayer
	state.TurnCounter = change.NewTurn

	return nil
}

// convertRuntimeWorldToProto converts runtime world state to protobuf WorldData
func (b *BaseGamesServiceImpl) convertRuntimeWorldToProto(world *World) *v1.WorldData {
	worldData := &v1.WorldData{
		Tiles: []*v1.Tile{},
		Units: []*v1.Unit{},
	}

	// Convert runtime tiles to protobuf tiles
	for coord, tile := range world.TilesByCoord() {
		protoTile := &v1.Tile{
			Q:        int32(coord.Q),
			R:        int32(coord.R),
			TileType: int32(tile.TileType),
			Player:   int32(tile.Player),
		}
		worldData.Tiles = append(worldData.Tiles, protoTile)
	}

	// Convert runtime units to protobuf units
	for coord, unit := range world.UnitsByCoord() {
		protoUnit := &v1.Unit{
			Q:                int32(coord.Q),
			R:                int32(coord.R),
			Player:           int32(unit.Player),
			UnitType:         int32(unit.UnitType),
			Shortcut:         unit.Shortcut, // Preserve shortcut
			AvailableHealth:  int32(unit.AvailableHealth),
			DistanceLeft:     unit.DistanceLeft,
			LastActedTurn:    int32(unit.LastActedTurn),
			LastToppedupTurn: int32(unit.LastToppedupTurn),
		}
		worldData.Units = append(worldData.Units, protoUnit)
	}

	return worldData
}

// SimulateAttack simulates combat between two units and returns damage distributions
func (s *BaseGamesServiceImpl) SimulateAttack(ctx context.Context, req *v1.SimulateAttackRequest) (resp *v1.SimulateAttackResponse, err error) {
	resp = &v1.SimulateAttackResponse{}

	// Set default number of simulations if not provided
	numSims := req.NumSimulations
	if numSims <= 0 {
		numSims = 1000
	}

	// Create mock units and tiles for simulation
	attackerUnit := &v1.Unit{
		Q:               0,
		R:               0,
		Player:          1,
		UnitType:        req.AttackerUnitType,
		AvailableHealth: req.AttackerHealth,
	}
	attackerTile := &v1.Tile{
		Q:        0,
		R:        0,
		TileType: req.AttackerTerrain,
	}

	defenderUnit := &v1.Unit{
		Q:               1,
		R:               0,
		Player:          2,
		UnitType:        req.DefenderUnitType,
		AvailableHealth: req.DefenderHealth,
	}
	defenderTile := &v1.Tile{
		Q:        1,
		R:        0,
		TileType: req.DefenderTerrain,
	}

	// Get rules engine
	rulesEngine := DefaultRulesEngine()

	// Simulate attacker -> defender
	attackerCtx := &CombatContext{
		Attacker:       attackerUnit,
		AttackerTile:   attackerTile,
		AttackerHealth: req.AttackerHealth,
		Defender:       defenderUnit,
		DefenderTile:   defenderTile,
		DefenderHealth: req.DefenderHealth,
		WoundBonus:     req.WoundBonus,
	}

	attackerDist, err := rulesEngine.GenerateDamageDistribution(attackerCtx, int(numSims))
	if err != nil {
		return nil, fmt.Errorf("failed to generate attacker damage distribution: %w", err)
	}

	// Convert attacker distribution to map
	attackerDamageMap := make(map[int32]int32)
	attackerMeanDamage := 0.0
	attackerKillCount := int32(0)
	for _, dmgRange := range attackerDist.Ranges {
		damage := int32(dmgRange.MinValue)
		count := int32(float64(numSims) * dmgRange.Probability)
		attackerDamageMap[damage] = count
		attackerMeanDamage += dmgRange.MinValue * dmgRange.Probability
		if damage >= req.DefenderHealth {
			attackerKillCount += count
		}
	}

	// Simulate defender -> attacker (counter-attack)
	defenderCtx := &CombatContext{
		Attacker:       defenderUnit,
		AttackerTile:   defenderTile,
		AttackerHealth: req.DefenderHealth,
		Defender:       attackerUnit,
		DefenderTile:   attackerTile,
		DefenderHealth: req.AttackerHealth,
		WoundBonus:     0, // No wound bonus for counter-attack
	}

	defenderDist, err := rulesEngine.GenerateDamageDistribution(defenderCtx, int(numSims))
	defenderDamageMap := make(map[int32]int32)
	defenderMeanDamage := 0.0
	defenderKillCount := int32(0)

	if err == nil && defenderDist != nil {
		for _, dmgRange := range defenderDist.Ranges {
			damage := int32(dmgRange.MinValue)
			count := int32(float64(numSims) * dmgRange.Probability)
			defenderDamageMap[damage] = count
			defenderMeanDamage += dmgRange.MinValue * dmgRange.Probability
			if damage >= req.AttackerHealth {
				defenderKillCount += count
			}
		}
	}

	resp.AttackerDamageDistribution = attackerDamageMap
	resp.DefenderDamageDistribution = defenderDamageMap
	resp.AttackerMeanDamage = attackerMeanDamage
	resp.DefenderMeanDamage = defenderMeanDamage
	resp.AttackerKillProbability = float64(attackerKillCount) / float64(numSims)
	resp.DefenderKillProbability = float64(defenderKillCount) / float64(numSims)

	return resp, nil
}
