package services

import (
	"context"
	"fmt"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	lib "github.com/turnforge/weewar/lib"
	"github.com/turnforge/weewar/web/assets/themes"
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

type CompactSummaryCardPanel interface {
	BasePanel
	SetCurrentData(context.Context, *v1.Tile, *v1.Unit)
}

type GameStatePanel interface {
	BasePanel
	Update(context.Context, *v1.Game, *v1.GameState)
}

type GameScene interface {
	BasePanel
	ClearPaths(context.Context)
	ClearHighlights(context.Context, *v1.ClearHighlightsRequest)
	ShowPath(context.Context, *v1.ShowPathRequest)
	ShowHighlights(context.Context, *v1.ShowHighlightsRequest)
	// Animation methods
	MoveUnit(context.Context, *v1.MoveUnitRequest) (*v1.MoveUnitResponse, error)
	ShowAttackEffect(context.Context, *v1.ShowAttackEffectRequest) (*v1.ShowAttackEffectResponse, error)
	ShowHealEffect(context.Context, *v1.ShowHealEffectRequest) (*v1.ShowHealEffectResponse, error)
	ShowCaptureEffect(context.Context, *v1.ShowCaptureEffectRequest) (*v1.ShowCaptureEffectResponse, error)
	SetUnitAt(context.Context, *v1.SetUnitAtRequest) (*v1.SetUnitAtResponse, error)
	RemoveUnitAt(context.Context, *v1.RemoveUnitAtRequest) (*v1.RemoveUnitAtResponse, error)
}

// type GameViewPresenter interface { v1.GameViewPresenterServer }

type GameViewerPageClient interface {
	SetAllowedPanels(context.Context, *v1.SetAllowedPanelsRequest) (*v1.SetAllowedPanelsResponse, error)
	SetCompactSummaryCard(context.Context, *v1.SetContentRequest) (*v1.SetContentResponse, error)
}

type BaseGameViewPresenter struct {
	GamesService GamesService
	RulesEngine  *v1.RulesEngine
	Theme        themes.Theme

	// All the "UI Elements" we will change state of
	GameState               GameState
	GameStatePanel          GameStatePanel
	TurnOptionsPanel        TurnOptionsPanel
	UnitStatsPanel          UnitStatsPanel
	DamageDistributionPanel DamageDistributionPanel
	TerrainStatsPanel       TerrainStatsPanel
	BuildOptionsModal       BuildOptionsModal
	CompactSummaryCardPanel CompactSummaryCardPanel
	GameScene               GameScene
	GameViewerPage          GameViewerPageClient // For mobile-specific RPC calls

	// State tracking for current selection
	selectedQ     *int32 // nil = no selection
	selectedR     *int32 // nil = no selection
	hasHighlights bool   // Track if highlights are currently shown
}

type GameViewPresenter struct {
	BaseGameViewPresenter
}

// NOTE - ONly API really needed here are "getters" and "move processors" so no Creations, Deletions, Listing or even
// GetGame needed - GetGame data is set when we create this
func NewGameViewPresenter() *GameViewPresenter {
	re := lib.DefaultRulesEngine()
	w := &GameViewPresenter{
		BaseGameViewPresenter: BaseGameViewPresenter{
			// WorldsService: WorldsService
			RulesEngine: re.RulesEngine,
			Theme:       themes.NewDefaultTheme(re.GetCityTerrains()), // Start with default theme
		},
	}
	return w
}

// Our initial game loader
func (s *GameViewPresenter) InitializeGame(ctx context.Context, req *v1.InitializeGameRequest) (resp *v1.InitializeGameResponse, err error) {
	getGameResp, err := s.GamesService.GetGame(ctx, &v1.GetGameRequest{Id: req.GameId})
	if err != nil {
		// TODO - handle gracefully
		panic(err)
	}
	game := getGameResp.Game
	gameState := getGameResp.State
	// moveHistory := s.GamesService.GameMoveHistory

	// Now update the game state based on this
	// Fire all the browser changes here - we dont really care about waiting for them
	// And more importantly we cannot block for them on the thread that called us
	s.TurnOptionsPanel.SetCurrentUnit(ctx, nil, nil)
	fmt.Println("setTurnOpt Resp, Err: ", resp, err)

	s.GameState.SetGameState(ctx, &v1.SetGameStateRequest{
		Game:  game,
		State: gameState,
	})
	s.GameStatePanel.Update(ctx, game, gameState)
	s.TerrainStatsPanel.SetCurrentTile(ctx, nil)
	s.UnitStatsPanel.SetCurrentUnit(ctx, nil)
	s.DamageDistributionPanel.SetCurrentUnit(ctx, nil)

	// Initialize mobile panels (no-op for desktop/grid)
	if s.GameViewerPage != nil {
		go s.GameViewerPage.SetAllowedPanels(ctx, &v1.SetAllowedPanelsRequest{
			PanelIds: []string{"game-state", "game-log", "turn-options"},
		})
	}

	// Note: Visual updates like exhausted highlights are deferred to ClientReady
	// which is called after the browser scene is fully initialized

	// Response state
	resp = &v1.InitializeGameResponse{
		Success:       true,
		CurrentPlayer: gameState.CurrentPlayer,
		TurnCounter:   gameState.TurnCounter,
		GameName:      game.Name,
	}
	return
}

// ClientReady is called by the browser after the UI/scene is fully initialized.
// This is when we can safely send visual updates like highlights.
func (s *GameViewPresenter) ClientReady(ctx context.Context, req *v1.ClientReadyRequest) (resp *v1.ClientReadyResponse, err error) {
	getGameResp, err := s.GamesService.GetGame(ctx, &v1.GetGameRequest{Id: req.GameId})
	if err != nil {
		return &v1.ClientReadyResponse{Success: false}, err
	}
	game := getGameResp.Game
	gameState := getGameResp.State

	// Now that the scene is ready, apply visual state
	s.refreshExhaustedHighlights(ctx, game, gameState)
	s.refreshCapturingHighlights(ctx, game, gameState)

	return &v1.ClientReadyResponse{Success: true}, nil
}

func (s *GameViewPresenter) GetGame(ctx context.Context, gameId string) (resp *v1.GetGameResponse, err error) {
	getGameResp, err := s.GamesService.GetGame(ctx, &v1.GetGameRequest{Id: gameId})
	if err != nil {
		// TODO - handle gracefully
		panic(err)
	}
	return getGameResp, err
}

func (s *GameViewPresenter) SceneClicked(ctx context.Context, req *v1.SceneClickedRequest) (resp *v1.SceneClickedResponse, err error) {
	resp = &v1.SceneClickedResponse{}
	getGameResp, err := s.GetGame(ctx, req.GameId)
	if err != nil {
		return
	}
	game := getGameResp.Game
	gameState := getGameResp.State
	rg, err := s.GamesService.GetRuntimeGame(game, gameState)
	q, r := req.Q, req.R
	coord := CoordFromInt32(q, r)

	// Get tile and unit data from World using coordinates
	switch req.Layer {
	case "movement-highlight", "capture-highlight":
		// User clicked on a movement/capture highlight - execute the action
		if err := s.executeMovementAction(ctx, game, gameState, q, r); err != nil {
			return nil, err
		}
	case "base-map":
		// Toggle behavior: if clicking the same position that's already selected, deselect
		if s.selectedQ != nil && s.selectedR != nil && *s.selectedQ == q && *s.selectedR == r {
			s.clearHighlightsAndSelection(ctx)
			s.TurnOptionsPanel.SetCurrentUnit(ctx, nil, nil)
			return
		}

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

		// Update mobile-specific panels
		if s.GameViewerPage != nil {
			// Update allowed panels based on what's selected
			allowedPanels := []string{"game-state", "game-log", "turn-options"}
			if tile != nil {
				allowedPanels = append(allowedPanels, "terrain-stats")
			}
			if unit != nil {
				allowedPanels = append(allowedPanels, "unit-stats", "damage-distribution")
			}
			fmt.Printf("[Presenter] SceneClicked: Setting allowed panels: %v (tile=%v, unit=%v)\n",
				allowedPanels, tile != nil, unit != nil)
			go s.GameViewerPage.SetAllowedPanels(ctx, &v1.SetAllowedPanelsRequest{
				PanelIds: allowedPanels,
			})
		}

		// Update compact summary card with combined terrain+unit info
		if s.CompactSummaryCardPanel != nil {
			s.CompactSummaryCardPanel.SetCurrentData(ctx, tile, unit)
		}

		// Top up unit if present
		if unit != nil {
			rg.TopUpUnitIfNeeded(unit)
		}

		// Get options at this position (handles both unit and tile actions)
		optionsResp, err := s.GamesService.GetOptionsAt(ctx, &v1.GetOptionsAtRequest{
			GameId: req.GameId,
			Q:      q,
			R:      r,
		})
		if err == nil && optionsResp != nil && len(optionsResp.Options) > 0 {
			// Check if there are ONLY build options (no movement/attack options)
			buildOptions := extractBuildOptions(optionsResp)
			hasOnlyBuildOptions := len(buildOptions) > 0 && len(buildOptions) == len(optionsResp.Options)
			fmt.Printf("[SceneClicked] Options count: %d, Build options count: %d, hasOnlyBuildOptions: %v\n",
				len(optionsResp.Options), len(buildOptions), hasOnlyBuildOptions)

			// Always populate TurnOptionsPanel so CLI can access options
			s.TurnOptionsPanel.SetCurrentUnit(ctx, unit, optionsResp)

			if hasOnlyBuildOptions {
				// Show build modal for web UI
				playerCoins := getPlayerCoins(gameState, gameState.CurrentPlayer)
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

// clearHighlightsAndSelection clears interactive highlights (selection, movement, attack, capture) but preserves exhausted highlights
func (s *GameViewPresenter) clearHighlightsAndSelection(ctx context.Context) {
	s.GameScene.ClearPaths(ctx)
	if s.hasHighlights {
		// Clear only interactive highlights, not exhausted
		s.GameScene.ClearHighlights(ctx, &v1.ClearHighlightsRequest{
			Types: []string{"selection", "movement", "attack", "build", "capture"},
		})
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

// getPlayerCoins returns the current player's coin count from game state
func getPlayerCoins(state *v1.GameState, playerID int32) int32 {
	if state == nil || state.PlayerStates == nil {
		return 0
	}
	if playerState := state.PlayerStates[playerID]; playerState != nil {
		return playerState.Coins
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
		} else if captureOpt := option.GetCapture(); captureOpt != nil {
			// Add capture highlight (same tile as unit)
			highlights = append(highlights, &v1.HighlightSpec{
				Q:      captureOpt.Q,
				R:      captureOpt.R,
				Type:   "capture",
				Action: &v1.HighlightSpec_Capture{Capture: captureOpt},
			})
		}
	}

	return highlights
}

// TurnOptionClicked handles when user clicks on a turn option in the TurnOptionsPanel
func (s *GameViewPresenter) TurnOptionClicked(ctx context.Context, req *v1.TurnOptionClickedRequest) (resp *v1.TurnOptionClickedResponse, err error) {
	resp = &v1.TurnOptionClickedResponse{}

	// Always clear previous paths first
	s.GameScene.ClearPaths(ctx)

	switch req.OptionType {
	case "move":
		// Show path visualization for move options
		if s.selectedQ != nil && s.selectedR != nil {
			optionsResp, err := s.GamesService.GetOptionsAt(ctx, &v1.GetOptionsAtRequest{
				GameId: req.GameId,
				Q:      *s.selectedQ,
				R:      *s.selectedR,
			})

			if err == nil && optionsResp != nil && int(req.OptionIndex) < len(optionsResp.Options) {
				option := optionsResp.Options[req.OptionIndex]
				if moveOpt := option.GetMove(); moveOpt != nil && moveOpt.ReconstructedPath != nil {
					coords := ExtractPathCoords(moveOpt.ReconstructedPath)
					if len(coords) >= 4 {
						s.GameScene.ShowPath(ctx, &v1.ShowPathRequest{
							Coords:    coords,
							Color:     0x00ff00, // Green for movement
							Thickness: 4,
						})
					}
				}
			}
		}

	case "capture":
		// Execute capture action
		if err = s.executeCaptureFromOption(ctx, req); err != nil {
			fmt.Printf("[Presenter] Capture failed: %v\n", err)
		}
	}

	return
}

// executeCaptureFromOption executes a capture action from TurnOptionClicked
func (s *GameViewPresenter) executeCaptureFromOption(ctx context.Context, req *v1.TurnOptionClickedRequest) error {
	// Get current game state
	getGameResp, err := s.GetGame(ctx, req.GameId)
	if err != nil {
		return err
	}
	game := getGameResp.Game
	gameState := getGameResp.State

	// Get the capture option from the stored options
	if s.TurnOptionsPanel == nil || s.TurnOptionsPanel.CurrentOptions() == nil {
		return fmt.Errorf("no options available")
	}

	options := s.TurnOptionsPanel.CurrentOptions()
	if int(req.OptionIndex) >= len(options.Options) {
		return fmt.Errorf("invalid option index: %d", req.OptionIndex)
	}

	option := options.Options[req.OptionIndex]
	captureOpt := option.GetCapture()
	if captureOpt == nil {
		return fmt.Errorf("option at index %d is not a capture option", req.OptionIndex)
	}

	// Create the capture move
	gameMove := &v1.GameMove{
		Player: gameState.CurrentPlayer,
		MoveType: &v1.GameMove_CaptureBuilding{
			CaptureBuilding: captureOpt,
		},
	}

	// Process the move
	resp, err := s.GamesService.ProcessMoves(ctx, &v1.ProcessMovesRequest{
		GameId: game.Id,
		Moves:  []*v1.GameMove{gameMove},
	})
	if err != nil {
		return err
	}

	fmt.Printf("[Presenter] Capture executed at (%d,%d)\n", captureOpt.Q, captureOpt.R)

	// Apply incremental updates (uses original gameState, changes come from resp.Moves)
	s.applyIncrementalChanges(ctx, game, gameState, resp.Moves, gameMove)

	return nil
}

// BuildOptionClicked handles when user clicks a build option in the BuildOptionsModal
func (s *GameViewPresenter) BuildOptionClicked(ctx context.Context, req *v1.BuildOptionClickedRequest) (resp *v1.BuildOptionClickedResponse, err error) {
	resp = &v1.BuildOptionClickedResponse{}

	// Get current game state
	getGameResp, err := s.GetGame(ctx, req.GameId)
	if err != nil {
		return
	}

	// Create the build move
	gameMove := &v1.GameMove{
		Player: getGameResp.State.CurrentPlayer,
		MoveType: &v1.GameMove_BuildUnit{
			BuildUnit: &v1.BuildUnitAction{
				Q:        req.Q,
				R:        req.R,
				UnitType: req.UnitType,
			},
		},
	}

	fmt.Printf("[Presenter] Executing build of unit type %d at (%d,%d) for player %d\n",
		req.UnitType, req.Q, req.R, getGameResp.State.CurrentPlayer)

	// Execute the build move
	s.executeBuildAction(ctx, getGameResp.Game, getGameResp.State, gameMove)

	return
}

// EndTurnButtonClicked handles when user clicks the end turn button
func (s *GameViewPresenter) EndTurnButtonClicked(ctx context.Context, req *v1.EndTurnButtonClickedRequest) (resp *v1.EndTurnButtonClickedResponse, err error) {
	resp = &v1.EndTurnButtonClickedResponse{}

	// Get current game state
	getGameResp, err := s.GetGame(ctx, req.GameId)
	if err != nil {
		return
	}

	s.executeEndTurnAction(ctx, getGameResp.Game, getGameResp.State)
	return
}

// executeMovementAction executes a movement when user clicks on a movement highlight
func (s *GameViewPresenter) executeMovementAction(ctx context.Context, game *v1.Game, gameState *v1.GameState, targetQ, targetR int32) error {
	// Get current options from TurnOptionsPanel
	currentOptions := s.TurnOptionsPanel.CurrentOptions()
	if currentOptions == nil || len(currentOptions.Options) == 0 {
		return fmt.Errorf("no options available for movement")
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
		} else if opt := option.GetCapture(); opt != nil {
			if opt.Q == targetQ && opt.R == targetR {
				gameMove = &v1.GameMove{
					Player:   gameState.CurrentPlayer,
					MoveType: &v1.GameMove_CaptureBuilding{CaptureBuilding: opt},
				}
				fmt.Printf("[Presenter] Executing capture at (%d,%d) for player %d\n",
					opt.Q, opt.R,
					gameState.CurrentPlayer)
				break
			}
		}
	}

	if gameMove == nil {
		return fmt.Errorf("no valid move, attack, or capture option found for target position (%d,%d)", targetQ, targetR)
	}

	// Call ProcessMoves to execute the move
	resp, err := s.GamesService.ProcessMoves(ctx, &v1.ProcessMovesRequest{GameId: game.Id, Moves: []*v1.GameMove{gameMove}})
	if err != nil {
		return fmt.Errorf("move execution failed: %w", err)
	}

	fmt.Println("[Presenter] Move executed successfully")

	// Apply other incremental updates (skip UnitMoved animations since we handled above)
	s.applyIncrementalChanges(ctx, game, gameState, resp.Moves, gameMove)

	return nil
}

// executeEndTurnAction executes the end turn action
// executeBuildAction processes a build unit action
func (s *GameViewPresenter) executeBuildAction(ctx context.Context, game *v1.Game, gameState *v1.GameState, gameMove *v1.GameMove) {
	// Call ProcessMoves to execute the build
	resp, err := s.GamesService.ProcessMoves(ctx, &v1.ProcessMovesRequest{GameId: game.Id, Moves: []*v1.GameMove{gameMove}})
	if err != nil {
		fmt.Printf("[Presenter] Build action failed: %v\n", err)
		return
	}

	fmt.Printf("[Presenter] Build action completed successfully\n")

	// Hide the build modal
	s.BuildOptionsModal.Hide(ctx)

	// Apply incremental updates from the move results
	s.applyIncrementalChanges(ctx, game, gameState, resp.Moves, gameMove)
}

func (s *GameViewPresenter) executeEndTurnAction(ctx context.Context, game *v1.Game, gameState *v1.GameState) {
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
		GameId: game.Id,
		Moves:  []*v1.GameMove{gameMove},
	})

	if err != nil {
		fmt.Printf("[Presenter] End turn failed: %v\n", err)
		return
	}

	fmt.Printf("[Presenter] Turn ended, new current player: %d\n", gameState.CurrentPlayer)

	// Apply incremental updates from the move results
	s.applyIncrementalChanges(ctx, game, gameState, resp.Moves, gameMove)
}

// applyIncrementalChanges processes WorldChange objects and calls incremental browser update methods
func (s *GameViewPresenter) applyIncrementalChanges(ctx context.Context, game *v1.Game, gameState *v1.GameState, moveResults []*v1.GameMove, gameMove *v1.GameMove) {
	// Clear selection and highlights
	s.clearHighlightsAndSelection(ctx)
	s.TurnOptionsPanel.SetCurrentUnit(ctx, nil, nil)
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
					unitMoved := changeType.UnitMoved
					if gameMove != nil && gameMove.GetMoveUnit() != nil {
						moveAction := gameMove.GetMoveUnit()
						coords := ExtractPathCoords(moveAction.ReconstructedPath)
						path = []*v1.HexCoord{}
						for i := 0; i < len(coords); i += 2 {
							path = append(path, &v1.HexCoord{Q: coords[i], R: coords[i+1]})
						}
						/*
							// If this was a move action with a reconstructed path, animate it with the full path
							// Build full path from reconstructed_path if available
							if moveAction.ReconstructedPath != nil && len(moveAction.ReconstructedPath.Edges) > 0 {
								// Extract coordinates from path edges
								path = make([]*v1.HexCoord, 0, len(moveAction.ReconstructedPath.Edges)+1)
								// Add starting point
								path = append(path, &v1.HexCoord{
									Q: moveAction.ReconstructedPath.Edges[0].FromQ,
									R: moveAction.ReconstructedPath.Edges[0].FromR,
								})
								// Add each destination along the path
								for _, edge := range moveAction.ReconstructedPath.Edges {
									path = append(path, &v1.HexCoord{
										Q: edge.ToQ,
										R: edge.ToR,
									})
								}
							} else {
								// Fallback to simple 2-point path
								path = []*v1.HexCoord{
									{Q: moveAction.FromQ, R: moveAction.FromR},
									{Q: moveAction.ToQ, R: moveAction.ToR},
								}
							}
						*/
					}

					// Animate the movement with full path
					s.GameScene.MoveUnit(ctx, &v1.MoveUnitRequest{
						Unit: unitMoved.UpdatedUnit,
						Path: path,
					})

					/*
						// Animate unit movement
						s.GameScene.MoveUnit(ctx, &v1.MoveUnitRequest{
							Unit: updatedUnit,
							Path: path,
						})
					*/
				}

			case *v1.WorldChange_UnitDamaged:
				updatedUnit := changeType.UnitDamaged.UpdatedUnit
				if updatedUnit != nil {
					// Update unit with flash effect for now
					// TODO: Enhance with attack animation when we have move context
					s.GameScene.SetUnitAt(ctx, &v1.SetUnitAtRequest{
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
					s.GameScene.RemoveUnitAt(ctx, &v1.RemoveUnitAtRequest{
						Q:       previousUnit.Q,
						R:       previousUnit.R,
						Animate: true,
					})
				}

			case *v1.WorldChange_UnitBuilt:
				// Add newly built unit to the game state
				builtUnit := changeType.UnitBuilt.Unit
				if builtUnit != nil {
					// Add unit with appear animation
					s.GameScene.SetUnitAt(ctx, &v1.SetUnitAtRequest{
						Q:      builtUnit.Q,
						R:      builtUnit.R,
						Unit:   builtUnit,
						Appear: true,
					})
				}

			case *v1.WorldChange_PlayerChanged:
				// Clear exhausted and capturing highlights for new turn
				// (units reset, and some captures may have completed)
				s.GameScene.ClearHighlights(ctx, &v1.ClearHighlightsRequest{
					Types: []string{"exhausted", "capturing"},
				})

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
				s.GameState.UpdateGameStatus(ctx, &v1.UpdateGameStatusRequest{
					CurrentPlayer: gameState.CurrentPlayer,
					TurnCounter:   gameState.TurnCounter,
				})

			case *v1.WorldChange_CoinsChanged:
				// Refresh game state panel to show updated coin balances
				// The GameState panel automatically displays player coins from the game state
				s.GameState.UpdateGameStatus(ctx, &v1.UpdateGameStatusRequest{
					CurrentPlayer: gameState.CurrentPlayer,
					TurnCounter:   gameState.TurnCounter,
				})

			case *v1.WorldChange_CaptureStarted:
				// Show capture effect animation and capturing flag
				captureStarted := changeType.CaptureStarted
				if captureStarted != nil {
					// One-shot pulse animation for immediate feedback
					s.GameScene.ShowCaptureEffect(ctx, &v1.ShowCaptureEffectRequest{
						Q: captureStarted.TileQ,
						R: captureStarted.TileR,
					})
					// Persistent flag indicator for ongoing capture
					s.GameScene.ShowHighlights(ctx, &v1.ShowHighlightsRequest{
						Highlights: []*v1.HighlightSpec{{
							Q:    captureStarted.TileQ,
							R:    captureStarted.TileR,
							Type: "capturing",
						}},
					})
					fmt.Printf("[Presenter] Capture started at (%d,%d) by player %d\n",
						captureStarted.TileQ, captureStarted.TileR, captureStarted.CapturingUnit.Player)
				}

			case *v1.WorldChange_TileCaptured:
				// Tile ownership changed - update tile display
				tileCaptured := changeType.TileCaptured
				if tileCaptured != nil {
					fmt.Printf("[Presenter] Tile captured at (%d,%d): player %d -> player %d\n",
						tileCaptured.TileQ, tileCaptured.TileR,
						tileCaptured.PreviousOwner, tileCaptured.NewOwner)
					// The game state will be refreshed below
				}

			default:
				fmt.Printf("[Presenter] Unknown world change type: %T\n", changeType)
			}
		}
	}

	// After applying all changes, refresh exhausted and capturing highlights and update game state panel
	s.refreshExhaustedHighlights(ctx, game, gameState)
	s.refreshCapturingHighlights(ctx, game, gameState)
	s.GameStatePanel.Update(ctx, game, gameState)
}

// refreshExhaustedHighlights updates the exhausted highlights for all units with no movement points
func (s *GameViewPresenter) refreshExhaustedHighlights(ctx context.Context, _ *v1.Game, gameState *v1.GameState) {
	// Build list of exhausted units/tiles
	var exhaustedHighlights []*v1.HighlightSpec

	// Check all units for the current player (using map-based storage)
	for _, unit := range gameState.WorldData.UnitsMap {
		if unit.Player == gameState.CurrentPlayer {
			// Mark as exhausted if no movement points left
			if unit.DistanceLeft <= 0 {
				exhaustedHighlights = append(exhaustedHighlights, &v1.HighlightSpec{
					Q:    unit.Q,
					R:    unit.R,
					Type: "exhausted",
				})
			}
		}
	}

	// Send exhausted highlights to browser
	if len(exhaustedHighlights) > 0 {
		s.GameScene.ShowHighlights(ctx, &v1.ShowHighlightsRequest{
			Highlights: exhaustedHighlights,
		})
	}
}

// refreshCapturingHighlights updates the capturing flag indicators for all units currently capturing
func (s *GameViewPresenter) refreshCapturingHighlights(ctx context.Context, _ *v1.Game, gameState *v1.GameState) {
	// Build list of units currently capturing
	var capturingHighlights []*v1.HighlightSpec

	// Check all units for any that are currently capturing (using map-based storage)
	for _, unit := range gameState.WorldData.UnitsMap {
		// CaptureStartedTurn > 0 means capture is in progress
		// TODO: When capture duration becomes configurable (N turns instead of 1),
		// also check that CaptureStartedTurn <= CurrentTurn - N to ensure capture
		// is still within the valid window
		if unit.CaptureStartedTurn > 0 {
			capturingHighlights = append(capturingHighlights, &v1.HighlightSpec{
				Q:      unit.Q,
				R:      unit.R,
				Type:   "capturing",
				Player: unit.Player,
			})
		}
	}

	// Send capturing highlights to browser
	if len(capturingHighlights) > 0 {
		s.GameScene.ShowHighlights(ctx, &v1.ShowHighlightsRequest{
			Highlights: capturingHighlights,
		})
	}
}
