package services

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GamesServiceImpl interface {
	v1.GamesServiceServer
	GetRuntimeGame(game *v1.Game, gameState *v1.GameState) (*weewar.Game, error)
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

	// Get the moves validted by the move processor, it is upto the move processor
	// to decide how "transactional" it wants to be - ie fail after  N moves,
	// success only if all moves succeeds etc.  Note that at this point the game
	// state has not changed and neither has the Runtime Game object.  Both the
	// GameState and the Runtime Game are checkpointed at before the moves started
	var dmp weewar.DefaultMoveProcessor
	for _, move := range req.Moves {
		fmt.Print("Found Move: ", move, move.MoveType)
	}
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

	// Check what's at this position
	unit := rtGame.World.UnitAt(weewar.AxialCoord{Q: int(req.Q), R: int(req.R)})

	if unit == nil {
		// Empty tile - check for building/capture options, then end turn
		// TODO: Add building construction options if this is a city/factory tile
		// For now, just add end turn
		options = append(options, &v1.GameOption{
			OptionType: &v1.GameOption_EndTurn{
				EndTurn: &v1.EndTurnOption{},
			},
		})
	} else if unit.Player != rtGame.CurrentPlayer {
		// Enemy unit - only end turn available (can't interact with enemy units directly)
		options = append(options, &v1.GameOption{
			OptionType: &v1.GameOption_EndTurn{
				EndTurn: &v1.EndTurnOption{},
			},
		})
	} else {
		// Our unit - get all available options
		var dmp weewar.DefaultMoveProcessor

		// Get movement options if unit has movement left
		if unit.AvailableHealth > 0 && unit.DistanceLeft > 0 {
			tileOptions, err := dmp.GetMovementOptions(rtGame, req.Q, req.R)
			if err == nil {
				for _, tileOption := range tileOptions {
					// Create ready-to-use MoveUnitAction
					moveAction := &v1.MoveUnitAction{
						FromQ: req.Q,
						FromR: req.R,
						ToQ:   int32(tileOption.Coord.Q),
						ToR:   int32(tileOption.Coord.R),
					}

					options = append(options, &v1.GameOption{
						OptionType: &v1.GameOption_Move{
							Move: &v1.MoveOption{
								Q:            int32(tileOption.Coord.Q),
								R:            int32(tileOption.Coord.R),
								MovementCost: int32(tileOption.Cost),
								Action:       moveAction,
							},
						},
					})
				}
			}
		}

		// Get attack options if unit can attack
		if unit.AvailableHealth > 0 {
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
									Q:                int32(coord.Q),
									R:                int32(coord.R),
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

		// TODO: Add capture building options if unit can capture buildings at this location
		// TODO: Add build unit options if this is a production facility

		// Always add end turn option
		options = append(options, &v1.GameOption{
			OptionType: &v1.GameOption_EndTurn{
				EndTurn: &v1.EndTurnOption{},
			},
		})
	}

	return &v1.GetOptionsAtResponse{
		Options:         options,
		CurrentPlayer:   rtGame.CurrentPlayer,
		GameInitialized: rtGame != nil && rtGame.World != nil,
	}, nil
}

func (b *BaseGamesServiceImpl) ApplyChangeResults(changes []*v1.GameMoveResult, rtGame *weewar.Game, game *v1.Game, state *v1.GameState, history *v1.GameMoveHistory) error {
	// Apply each change to both runtime game and protobuf data structures
	for _, moveResult := range changes {
		for _, change := range moveResult.Changes {
			err := b.applyWorldChange(change, rtGame, state)
			if err != nil {
				return fmt.Errorf("failed to apply world change: %w", err)
			}
		}
	}

	// Update protobuf GameState with final runtime world state
	state.WorldData = b.convertRuntimeWorldToProto(rtGame.World)
	state.UpdatedAt = timestamppb.New(time.Now())

	return nil
}

// applyWorldChange applies a single WorldChange to both runtime game and protobuf state
func (b *BaseGamesServiceImpl) applyWorldChange(change *v1.WorldChange, rtGame *weewar.Game, state *v1.GameState) error {
	switch changeType := change.ChangeType.(type) {
	case *v1.WorldChange_UnitMoved:
		return b.applyUnitMoved(changeType.UnitMoved, rtGame)
	case *v1.WorldChange_UnitDamaged:
		return b.applyUnitDamaged(changeType.UnitDamaged, rtGame)
	case *v1.WorldChange_UnitKilled:
		return b.applyUnitKilled(changeType.UnitKilled, rtGame)
	case *v1.WorldChange_PlayerChanged:
		return b.applyPlayerChanged(changeType.PlayerChanged, rtGame)
	default:
		return fmt.Errorf("unknown world change type")
	}
}

// applyUnitMoved moves a unit in the runtime game
func (b *BaseGamesServiceImpl) applyUnitMoved(change *v1.UnitMovedChange, rtGame *weewar.Game) error {
	fromCoord := weewar.AxialCoord{Q: int(change.FromQ), R: int(change.FromR)}
	toCoord := weewar.AxialCoord{Q: int(change.ToQ), R: int(change.ToR)}

	// Move unit in runtime game
	unit := rtGame.World.UnitAt(fromCoord)
	if unit == nil {
		return fmt.Errorf("unit not found at %v", fromCoord)
	}

	// Remove from old position and add to new position
	return rtGame.World.MoveUnit(unit, toCoord)
}

// applyUnitDamaged updates unit health in the runtime game
func (b *BaseGamesServiceImpl) applyUnitDamaged(change *v1.UnitDamagedChange, rtGame *weewar.Game) error {
	coord := weewar.AxialCoord{Q: int(change.Q), R: int(change.R)}

	unit := rtGame.World.UnitAt(coord)
	if unit == nil {
		return fmt.Errorf("unit not found at %v", coord)
	}

	unit.AvailableHealth = change.NewHealth
	return nil
}

// applyUnitKilled removes a unit from the runtime game
func (b *BaseGamesServiceImpl) applyUnitKilled(change *v1.UnitKilledChange, rtGame *weewar.Game) error {
	coord := weewar.AxialCoord{Q: int(change.Q), R: int(change.R)}
	unit := rtGame.World.UnitAt(coord)

	err := rtGame.World.RemoveUnit(unit)
	if err != nil {
		return fmt.Errorf("unit not found at %v", coord)
	}
	return nil
}

// applyPlayerChanged updates game state for turn/player changes
func (b *BaseGamesServiceImpl) applyPlayerChanged(change *v1.PlayerChangedChange, rtGame *weewar.Game) error {
	rtGame.CurrentPlayer = change.NewPlayer
	rtGame.TurnCounter = change.NewTurn
	return nil
}

// convertRuntimeWorldToProto converts runtime world state to protobuf WorldData
func (b *BaseGamesServiceImpl) convertRuntimeWorldToProto(world *weewar.World) *v1.WorldData {
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
			Q:               int32(coord.Q),
			R:               int32(coord.R),
			Player:          int32(unit.Player),
			UnitType:        int32(unit.UnitType),
			AvailableHealth: int32(unit.AvailableHealth),
			DistanceLeft:    int32(unit.DistanceLeft),
			TurnCounter:     int32(unit.TurnCounter),
		}
		worldData.Units = append(worldData.Units, protoUnit)
	}

	return worldData
}
