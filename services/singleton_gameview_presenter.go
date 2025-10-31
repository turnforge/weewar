package services

import (
	"context"
	"fmt"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"github.com/panyam/turnengine/games/weewar/web/assets/themes"
)

type BasePanel interface {
	SetTheme(t themes.Theme)
	SetRulesEngine(t *v1.RulesEngine)
}

type GameState interface {
	SetGameState(context.Context, *v1.SetGameStateRequest) (*v1.SetGameStateResponse, error)
	RemoveUnitAt(context.Context, *v1.RemoveUnitAtRequest) (*v1.RemoveUnitAtResponse, error)
	SetUnitAt(context.Context, *v1.SetUnitAtRequest) (*v1.SetUnitAtResponse, error)
	UpdateGameStatus(context.Context, *v1.UpdateGameStatusRequest) (*v1.UpdateGameStatusResponse, error)
}

type TurnOptionsPanel interface {
	BasePanel
	CurrentOptions() *v1.GetOptionsAtResponse
	CurrentUnit() *v1.Unit
	SetCurrentUnit(context.Context, *v1.Unit, *v1.GetOptionsAtResponse)
}

type UnitStatsPanel interface {
	BasePanel
	CurrentUnit() *v1.Unit
	SetCurrentUnit(context.Context, *v1.Unit)
}

type DamageDistributionPanel interface {
	BasePanel
	CurrentUnit() *v1.Unit
	SetCurrentUnit(context.Context, *v1.Unit)
}

type TerrainStatsPanel interface {
	BasePanel
	CurrentTile() *v1.Tile
	SetCurrentTile(context.Context, *v1.Tile)
}

type BuildOptionsModal interface {
	BasePanel
	Show(context.Context, *v1.Tile, []*v1.BuildUnitAction, int32)
	Hide(context.Context)
}

type GameScene interface {
	BasePanel
	ClearPaths(context.Context)
	ClearHighlights(context.Context)
	ShowPath(context.Context, *v1.ShowPathRequest)
	ShowHighlights(context.Context, *v1.ShowHighlightsRequest)
	// Animation methods
	MoveUnit(context.Context, *v1.MoveUnitAnimationRequest) (*v1.MoveUnitAnimationResponse, error)
	ShowAttackEffect(context.Context, *v1.ShowAttackEffectRequest) (*v1.ShowAttackEffectResponse, error)
	ShowHealEffect(context.Context, *v1.ShowHealEffectRequest) (*v1.ShowHealEffectResponse, error)
	ShowCaptureEffect(context.Context, *v1.ShowCaptureEffectRequest) (*v1.ShowCaptureEffectResponse, error)
	SetUnitAt(context.Context, *v1.SetUnitAtAnimationRequest) (*v1.SetUnitAtAnimationResponse, error)
	RemoveUnitAt(context.Context, *v1.RemoveUnitAtAnimationRequest) (*v1.RemoveUnitAtAnimationResponse, error)
}

type SingletonGameViewPresenterImpl struct {
	BaseGameViewPresenterImpl
	GamesService *SingletonGamesServiceImpl
	RulesEngine  *v1.RulesEngine
	Theme        themes.Theme

	// All the "UI Elements" we will change state of
	GameState               GameState
	TurnOptionsPanel        TurnOptionsPanel
	UnitStatsPanel          UnitStatsPanel
	DamageDistributionPanel DamageDistributionPanel
	TerrainStatsPanel       TerrainStatsPanel
	BuildOptionsModal       BuildOptionsModal
	GameScene               GameScene

	// State tracking for current selection
	selectedQ     *int32 // nil = no selection
	selectedR     *int32 // nil = no selection
	hasHighlights bool   // Track if highlights are currently shown
}

// NOTE - ONly API really needed here are "getters" and "move processors" so no Creations, Deletions, Listing or even
// GetGame needed - GetGame data is set when we create this
func NewSingletonGameViewPresenterImpl() *SingletonGameViewPresenterImpl {
	w := &SingletonGameViewPresenterImpl{
		BaseGameViewPresenterImpl: BaseGameViewPresenterImpl{
			// WorldsService: SingletonWorldsService
		},
		RulesEngine: DefaultRulesEngine().RulesEngine,
		Theme:       themes.NewDefaultTheme(), // Start with default theme
	}
	return w
}

// Our initial game loader
func (s *SingletonGameViewPresenterImpl) InitializeGame(ctx context.Context, req *v1.InitializeGameRequest) (resp *v1.InitializeGameResponse, err error) {
	s.GamesService.Load([]byte(req.GameData), []byte(req.GameState), []byte(req.MoveHistory))
	game := s.GamesService.SingletonGame
	gameState := s.GamesService.SingletonGameState
	// moveHistory := s.GamesService.SingletonGameMoveHistory

	// Now update the game state based on this
	// Fire all the browser changes here - we dont really care about waiting for them
	// And more importantly we cannot block for them on the thread that called us
	s.TurnOptionsPanel.SetCurrentUnit(ctx, nil, nil)
	fmt.Println("setTurnOpt Resp, Err: ", resp, err)

	s.GameState.SetGameState(ctx, &v1.SetGameStateRequest{
		Game:  game,
		State: gameState,
	})
	s.TerrainStatsPanel.SetCurrentTile(ctx, nil)
	s.UnitStatsPanel.SetCurrentUnit(ctx, nil)
	s.DamageDistributionPanel.SetCurrentUnit(ctx, nil)

	// Response state
	resp = &v1.InitializeGameResponse{
		Success:       true,
		CurrentPlayer: gameState.CurrentPlayer,
		TurnCounter:   gameState.TurnCounter,
		GameName:      game.Name,
	}
	return
}

func (s *SingletonGameViewPresenterImpl) SceneClicked(ctx context.Context, req *v1.SceneClickedRequest) (resp *v1.SceneClickedResponse, err error) {
	resp = &v1.SceneClickedResponse{}
	game := s.GamesService.SingletonGame
	gameState := s.GamesService.SingletonGameState
	rg, err := s.GamesService.GetRuntimeGame(game, gameState)
	q, r := req.Q, req.R
	coord := CoordFromInt32(q, r)

	// Get tile and unit data from World using coordinates
	switch req.Layer {
	case "movement-highlight":
		// User clicked on a movement highlight - execute the move
		s.executeMovementAction(ctx, q, r)
	case "base-map":
		wd := rg.World
		if err != nil {
			panic(err)
		}
		unit := wd.UnitAt(coord)
		tile := wd.TileAt(coord)

		// Always show terrain and unit info (methods handle nil)
		s.TerrainStatsPanel.SetCurrentTile(ctx, tile)
		s.UnitStatsPanel.SetCurrentUnit(ctx, unit)
		s.DamageDistributionPanel.SetCurrentUnit(ctx, unit)

		// Top up unit if present
		if unit != nil {
			rg.TopUpUnitIfNeeded(unit)
		}

		// Get options at this position (handles both unit and tile actions)
		optionsResp, err := s.GamesService.GetOptionsAt(ctx, &v1.GetOptionsAtRequest{
			Q: q,
			R: r,
		})
		if err == nil && optionsResp != nil && len(optionsResp.Options) > 0 {
			// Check if there are ONLY build options (no movement/attack options)
			buildOptions := extractBuildOptions(optionsResp)
			hasOnlyBuildOptions := len(buildOptions) > 0 && len(buildOptions) == len(optionsResp.Options)
			fmt.Printf("[SceneClicked] Options count: %d, Build options count: %d, hasOnlyBuildOptions: %v\n",
				len(optionsResp.Options), len(buildOptions), hasOnlyBuildOptions)

			if hasOnlyBuildOptions {
				// Show build modal instead of turn options panel
				playerCoins := getPlayerCoins(game, gameState.CurrentPlayer)
				fmt.Printf("[SceneClicked] Showing build modal with %d options, playerCoins=%d\n",
					len(buildOptions), playerCoins)
				s.BuildOptionsModal.Show(ctx, tile, buildOptions, playerCoins)
				// Still show selection highlight
				s.GameScene.ShowHighlights(ctx, &v1.ShowHighlightsRequest{
					Highlights: []*v1.HighlightSpec{{Q: q, R: r, Type: "selection"}},
				})
				s.hasHighlights = true
				s.selectedQ = &q
				s.selectedR = &r
			} else {
				// Show turn options panel as usual
				s.TurnOptionsPanel.SetCurrentUnit(ctx, unit, optionsResp)

				// Send visualization commands to show highlights
				highlights := buildHighlightSpecs(optionsResp, q, r)
				if len(highlights) > 0 {
					s.GameScene.ShowHighlights(ctx, &v1.ShowHighlightsRequest{
						Highlights: highlights,
					})
					s.hasHighlights = true
					s.selectedQ = &q
					s.selectedR = &r
				}
			}
		} else {
			// No options available - clear options and highlights
			s.TurnOptionsPanel.SetCurrentUnit(ctx, nil, nil)
			s.clearHighlightsAndSelection(ctx)
		}
	default:
		fmt.Println("[GameViewerPage] Unhandled layer click: ", req.Layer)
	}
	return
}

// clearHighlightsAndSelection clears highlights if any are currently shown
func (s *SingletonGameViewPresenterImpl) clearHighlightsAndSelection(ctx context.Context) {
	s.GameScene.ClearPaths(ctx)
	if s.hasHighlights {
		s.GameScene.ClearHighlights(ctx)
		s.hasHighlights = false
		s.selectedQ = nil
		s.selectedR = nil
	}
}

// extractBuildOptions extracts all build options from the response
func extractBuildOptions(optionsResp *v1.GetOptionsAtResponse) []*v1.BuildUnitAction {
	if optionsResp == nil {
		return nil
	}

	buildOptions := []*v1.BuildUnitAction{}
	for _, option := range optionsResp.Options {
		if buildOpt := option.GetBuild(); buildOpt != nil {
			buildOptions = append(buildOptions, buildOpt)
		}
	}
	return buildOptions
}

// getPlayerCoins returns the current player's coin count from game configuration
func getPlayerCoins(game *v1.Game, playerID int32) int32 {
	if game == nil || game.Config == nil || game.Config.Players == nil {
		return 0
	}
	for _, player := range game.Config.Players {
		if player.PlayerId == playerID {
			return player.Coins
		}
	}
	return 0
}

// buildHighlightSpecs creates HighlightSpec array from GetOptionsAt response
// Extracts selection, movement, and attack highlights from the options
func buildHighlightSpecs(optionsResp *v1.GetOptionsAtResponse, selectedQ, selectedR int32) []*v1.HighlightSpec {
	if optionsResp == nil || len(optionsResp.Options) == 0 {
		return nil
	}

	highlights := []*v1.HighlightSpec{}

	// Add selection highlight for the clicked position
	highlights = append(highlights, &v1.HighlightSpec{
		Q:    selectedQ,
		R:    selectedR,
		Type: "selection",
	})

	// Extract highlights from options
	for _, option := range optionsResp.Options {
		if moveOpt := option.GetMove(); moveOpt != nil {
			// Add movement highlight
			highlights = append(highlights, &v1.HighlightSpec{
				Type:   "movement",
				Q:      moveOpt.ToQ,
				R:      moveOpt.ToR,
				Action: &v1.HighlightSpec_Move{Move: moveOpt},
			})
		} else if attackOpt := option.GetAttack(); attackOpt != nil {
			// Add attack highlight
			highlights = append(highlights, &v1.HighlightSpec{
				Q:      attackOpt.DefenderQ,
				R:      attackOpt.DefenderR,
				Type:   "attack",
				Action: &v1.HighlightSpec_Attack{Attack: attackOpt},
			})
		} else if buildOpt := option.GetBuild(); buildOpt != nil {
			// Add build highlight
			highlights = append(highlights, &v1.HighlightSpec{
				Q:      buildOpt.Q,
				R:      buildOpt.R,
				Type:   "build",
				Action: &v1.HighlightSpec_Build{Build: buildOpt},
			})
		}
	}

	return highlights
}

// TurnOptionClicked handles when user clicks on a turn option in the TurnOptionsPanel
func (s *SingletonGameViewPresenterImpl) TurnOptionClicked(ctx context.Context, req *v1.TurnOptionClickedRequest) (resp *v1.TurnOptionClickedResponse, err error) {
	resp = &v1.TurnOptionClickedResponse{}

	// For now, just show path visualization for move options
	// In the future, this could execute the actual move/attack
	// Always clear previous paths first
	s.GameScene.ClearPaths(ctx)

	if req.OptionType == "move" && s.selectedQ != nil && s.selectedR != nil {
		// Get the options again to extract the path for this specific move
		optionsResp, err := s.GamesService.GetOptionsAt(ctx, &v1.GetOptionsAtRequest{
			Q: *s.selectedQ,
			R: *s.selectedR,
		})

		if err == nil && optionsResp != nil && int(req.OptionIndex) < len(optionsResp.Options) {
			option := optionsResp.Options[req.OptionIndex]
			if moveOpt := option.GetMove(); moveOpt != nil && moveOpt.ReconstructedPath != nil {
				// Extract path coordinates from the reconstructed path
				coords := extractPathCoords(moveOpt.ReconstructedPath)
				if len(coords) >= 4 {
					// Show green path for movement
					s.GameScene.ShowPath(ctx, &v1.ShowPathRequest{
						Coords:    coords,
						Color:     0x00ff00, // Green for movement
						Thickness: 4,
					})
				}
			}
		}
	}

	return
}

// BuildOptionClicked handles when user clicks a build option in the BuildOptionsModal
func (s *SingletonGameViewPresenterImpl) BuildOptionClicked(ctx context.Context, req *v1.BuildOptionClickedRequest) (resp *v1.BuildOptionClickedResponse, err error) {
	resp = &v1.BuildOptionClickedResponse{}

	// Get current game state
	gameState := s.GamesService.SingletonGameState
	if gameState == nil {
		fmt.Println("[Presenter] Game state is nil")
		return
	}

	// Create the build move
	gameMove := &v1.GameMove{
		Player: gameState.CurrentPlayer,
		MoveType: &v1.GameMove_BuildUnit{
			BuildUnit: &v1.BuildUnitAction{
				Q:        req.Q,
				R:        req.R,
				UnitType: req.UnitType,
			},
		},
	}

	fmt.Printf("[Presenter] Executing build of unit type %d at (%d,%d) for player %d\n",
		req.UnitType, req.Q, req.R, gameState.CurrentPlayer)

	// Execute the build move
	s.executeBuildAction(ctx, gameMove)

	return
}

// EndTurnButtonClicked handles when user clicks the end turn button
func (s *SingletonGameViewPresenterImpl) EndTurnButtonClicked(ctx context.Context, req *v1.EndTurnButtonClickedRequest) (resp *v1.EndTurnButtonClickedResponse, err error) {
	resp = &v1.EndTurnButtonClickedResponse{}
	s.executeEndTurnAction(ctx)
	return
}

// extractPathCoords converts a Path to a flat coordinate array
func extractPathCoords(path *v1.Path) []int32 {
	if path == nil || len(path.Edges) == 0 {
		return nil
	}

	coords := make([]int32, 0, len(path.Edges)*2+2)

	// Add starting position from first edge
	if len(path.Edges) > 0 {
		coords = append(coords, path.Edges[0].FromQ, path.Edges[0].FromR)
	}

	// Add all destination positions
	for _, edge := range path.Edges {
		coords = append(coords, edge.ToQ, edge.ToR)
	}

	return coords
}

// executeMovementAction executes a movement when user clicks on a movement highlight
func (s *SingletonGameViewPresenterImpl) executeMovementAction(ctx context.Context, targetQ, targetR int32) {
	// Get current options from TurnOptionsPanel
	currentOptions := s.TurnOptionsPanel.CurrentOptions()
	if currentOptions == nil || len(currentOptions.Options) == 0 {
		fmt.Println("[Presenter] No options available for movement")
		return
	}

	// Get current game state
	gameState := s.GamesService.SingletonGameState
	if gameState == nil {
		fmt.Println("[Presenter] Game state is nil")
		return
	}

	// Find the move option that matches the clicked coordinates
	var gameMove *v1.GameMove
	for _, option := range currentOptions.Options {
		if opt := option.GetMove(); opt != nil {
			if opt.ToQ == targetQ && opt.ToR == targetR {
				gameMove = &v1.GameMove{
					Player:   gameState.CurrentPlayer,
					MoveType: &v1.GameMove_MoveUnit{MoveUnit: opt},
				}
				fmt.Printf("[Presenter] Executing move from (%d,%d) to (%d,%d) for player %d\n",
					opt.FromQ, opt.FromR,
					opt.ToQ, opt.ToR,
					gameState.CurrentPlayer)
				break
			}
		} else if opt := option.GetAttack(); opt != nil {
			if opt.DefenderQ == targetQ && opt.DefenderR == targetR {
				gameMove = &v1.GameMove{
					Player:   gameState.CurrentPlayer,
					MoveType: &v1.GameMove_AttackUnit{AttackUnit: opt},
				}
				fmt.Printf("[Presenter] Executing attack from (%d,%d) to (%d,%d) for player %d\n",
					opt.AttackerQ, opt.AttackerR,
					opt.DefenderQ, opt.DefenderR,
					gameState.CurrentPlayer)
				break
			}
		}
	}

	if gameMove == nil {
		fmt.Println("[Presenter] Move/Attack option does not contain action object")
		return
	}

	// Call ProcessMoves to execute the move
	resp, err := s.GamesService.ProcessMoves(ctx, &v1.ProcessMovesRequest{
		Moves: []*v1.GameMove{gameMove},
	})

	if err != nil {
		fmt.Printf("[Presenter] Move execution failed: %v\n", err)
		return
	}

	fmt.Println("[Presenter] Move executed successfully")

	// Clear selection and highlights after successful move
	s.clearHighlightsAndSelection(ctx)
	s.TurnOptionsPanel.SetCurrentUnit(ctx, nil, nil)

	// Apply incremental updates from the move results
	s.applyIncrementalChanges(ctx, resp.MoveResults)
}

// executeEndTurnAction executes the end turn action
// executeBuildAction processes a build unit action
func (s *SingletonGameViewPresenterImpl) executeBuildAction(ctx context.Context, gameMove *v1.GameMove) {
	// Call ProcessMoves to execute the build
	_, err := s.GamesService.ProcessMoves(ctx, &v1.ProcessMovesRequest{
		Moves: []*v1.GameMove{gameMove},
	})

	if err != nil {
		fmt.Printf("[Presenter] Build action failed: %v\n", err)
		return
	}

	fmt.Printf("[Presenter] Build action completed successfully\n")

	// Hide the build modal
	s.BuildOptionsModal.Hide(ctx)

	// Clear selection and highlights
	s.clearHighlightsAndSelection(ctx)

	// Clear turn options panel
	s.TurnOptionsPanel.SetCurrentUnit(ctx, nil, nil)
}

func (s *SingletonGameViewPresenterImpl) executeEndTurnAction(ctx context.Context) {
	gameState := s.GamesService.SingletonGameState
	if gameState == nil {
		fmt.Println("[Presenter] Game state is nil")
		return
	}

	fmt.Printf("[Presenter] Ending turn for player %d\n", gameState.CurrentPlayer)

	// Create end turn move
	gameMove := &v1.GameMove{
		Player: gameState.CurrentPlayer,
		MoveType: &v1.GameMove_EndTurn{
			EndTurn: &v1.EndTurnAction{},
		},
	}

	// Call ProcessMoves to execute end turn
	resp, err := s.GamesService.ProcessMoves(ctx, &v1.ProcessMovesRequest{
		Moves: []*v1.GameMove{gameMove},
	})

	if err != nil {
		fmt.Printf("[Presenter] End turn failed: %v\n", err)
		return
	}

	fmt.Printf("[Presenter] Turn ended, new current player: %d\n", s.GamesService.SingletonGameState.CurrentPlayer)

	// Clear selection and highlights
	s.clearHighlightsAndSelection(ctx)
	s.TurnOptionsPanel.SetCurrentUnit(ctx, nil, nil)

	// Apply incremental updates from the move results
	s.applyIncrementalChanges(ctx, resp.MoveResults)
}

// applyIncrementalChanges processes WorldChange objects and calls incremental browser update methods
func (s *SingletonGameViewPresenterImpl) applyIncrementalChanges(ctx context.Context, moveResults []*v1.GameMoveResult) {
	for _, result := range moveResults {
		for _, change := range result.Changes {
			switch changeType := change.ChangeType.(type) {
			case *v1.WorldChange_UnitMoved:
				prevUnit := changeType.UnitMoved.PreviousUnit
				updatedUnit := changeType.UnitMoved.UpdatedUnit
				if prevUnit != nil && updatedUnit != nil {
					// Build path for animation (simple: previous -> new)
					path := []*v1.HexCoord{
						{Q: prevUnit.Q, R: prevUnit.R},
						{Q: updatedUnit.Q, R: updatedUnit.R},
					}
					// Animate unit movement
					s.GameScene.MoveUnit(ctx, &v1.MoveUnitAnimationRequest{
						Unit: updatedUnit,
						Path: path,
					})
				}

			case *v1.WorldChange_UnitDamaged:
				updatedUnit := changeType.UnitDamaged.UpdatedUnit
				if updatedUnit != nil {
					// Update unit with flash effect for now
					// TODO: Enhance with attack animation when we have move context
					s.GameScene.SetUnitAt(ctx, &v1.SetUnitAtAnimationRequest{
						Q:     updatedUnit.Q,
						R:     updatedUnit.R,
						Unit:  updatedUnit,
						Flash: true,
					})
				}

			case *v1.WorldChange_UnitKilled:
				previousUnit := changeType.UnitKilled.PreviousUnit
				if previousUnit != nil {
					// Remove unit with death animation
					s.GameScene.RemoveUnitAt(ctx, &v1.RemoveUnitAtAnimationRequest{
						Q:       previousUnit.Q,
						R:       previousUnit.R,
						Animate: true,
					})
				}

			case *v1.WorldChange_PlayerChanged:
				// Reset all units for new turn (lazy top-up pattern)
				if changeType.PlayerChanged.ResetUnits != nil {
					for _, resetUnit := range changeType.PlayerChanged.ResetUnits {
						s.GameState.SetUnitAt(ctx, &v1.SetUnitAtRequest{
							Q:    resetUnit.Q,
							R:    resetUnit.R,
							Unit: resetUnit,
						})
					}
				}
				// Update game UI status with new current player and turn counter
				gameState := s.GamesService.SingletonGameState
				s.GameState.UpdateGameStatus(ctx, &v1.UpdateGameStatusRequest{
					CurrentPlayer: gameState.CurrentPlayer,
					TurnCounter:   gameState.TurnCounter,
				})

			default:
				fmt.Printf("[Presenter] Unknown world change type: %T\n", changeType)
			}
		}
	}
}
