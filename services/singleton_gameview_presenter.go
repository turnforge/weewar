package services

import (
	"bytes"
	"context"
	"fmt"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	lib "github.com/panyam/turnengine/games/weewar/lib"
	"github.com/panyam/turnengine/games/weewar/web/assets/themes"
	tmpls "github.com/panyam/turnengine/games/weewar/web/templates"
)

type SingletonGameViewPresenterImpl struct {
	BaseGameViewPresenterImpl
	GameViewerPage v1.GameViewerPageClient
	GamesService   *SingletonGamesServiceImpl
	RulesEngine    *v1.RulesEngine
	Theme          themes.Theme

	// State tracking for current selection
	selectedQ    *int32  // nil = no selection
	selectedR    *int32  // nil = no selection
	hasHighlights bool   // Track if highlights are currently shown
}

// NOTE - ONly API really needed here are "getters" and "move processors" so no Creations, Deletions, Listing or even
// GetGame needed - GetGame data is set when we create this
func NewSingletonGameViewPresenterImpl() *SingletonGameViewPresenterImpl {
	w := &SingletonGameViewPresenterImpl{
		BaseGameViewPresenterImpl: BaseGameViewPresenterImpl{
			// WorldsService: SingletonWorldsService
		},
		RulesEngine: lib.DefaultRulesEngine().RulesEngine,
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
	go func() {
		resp, err := s.GameViewerPage.SetTurnOptionsContent(ctx, &v1.SetContentRequest{
			InnerHtml: "<div class='text-center text-gray-500'>Select a unit to see options</div>",
		})
		fmt.Println("setTurnOpt Resp, Err: ", resp, err)

		s.GameViewerPage.SetGameState(ctx, &v1.SetGameStateRequest{
			Game:  game,
			State: gameState,
		})
		s.SetTerrainStats(ctx, nil)
		s.SetUnitStats(ctx, nil)
		s.SetUnitDamageDistribution(ctx, nil)
	}()

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
	q, r := req.Q, req.R
	coord := lib.CoordFromInt32(q, r)

	// Get tile and unit data from World using coordinates
	switch req.Layer {
	case "movement-highlight":
		// Get moveOption from the layer itself
		/*
		   const movementLayer = this.gameScene.movementHighlightLayer;
		   const moveOption = movementLayer?.getMoveOptionAt(q, r);
		   this.handleMovementClick(q, r, moveOption);
		*/
		break
	case "base-map":
		go func() {
			rg, err := s.GamesService.GetRuntimeGame(game, gameState)
			wd := rg.World
			if err != nil {
				panic(err)
			}
			unit := wd.UnitAt(coord)
			tile := wd.TileAt(coord)

			// Always show terrain and unit info (methods handle nil)
			s.SetTerrainStats(ctx, tile)
			s.SetUnitStats(ctx, unit)
			s.SetUnitDamageDistribution(ctx, unit)

			// Only proceed with options and highlights if there's a unit
			if unit != nil {
				// Get options at this position and update TurnOptionsPanel
				optionsResp, err := s.GamesService.GetOptionsAt(ctx, &v1.GetOptionsAtRequest{
					Q: q,
					R: r,
				})
				if err == nil && optionsResp != nil && len(optionsResp.Options) > 0 {
					s.SetTurnOptions(ctx, optionsResp, unit)

					// Send visualization commands to show highlights
					highlights := buildHighlightSpecs(optionsResp, q, r)
					if len(highlights) > 0 {
						s.GameViewerPage.ShowHighlights(ctx, &v1.ShowHighlightsRequest{
							Highlights: highlights,
						})
						s.hasHighlights = true
						s.selectedQ = &q
						s.selectedR = &r
					}
				} else {
					// Unit exists but no options available
					s.SetTurnOptions(ctx, &v1.GetOptionsAtResponse{Options: nil}, nil)
					s.clearHighlightsAndSelection(ctx)
				}
			} else {
				// No unit at clicked position - clear options and highlights
				s.SetTurnOptions(ctx, &v1.GetOptionsAtResponse{Options: nil}, nil)
				s.clearHighlightsAndSelection(ctx)
			}
		}()
	default:
		fmt.Println("[GameViewerPage] Unhandled layer click: ", req.Layer)
	}
	return
}

func (s *SingletonGameViewPresenterImpl) renderPanelTemplate(_ context.Context, templatefile string, data any) (content string) {
	tmpl, err := tmpls.Templates.Loader.Load(templatefile, "")
	if err == nil {
		buf := bytes.NewBufferString("")
		err = tmpls.Templates.RenderHtmlTemplate(buf, tmpl[0], "", data, nil)
		if err == nil {
			content = buf.String()
		}
	}
	if err != nil {
		panic(err)
	}
	return
}

func (s *SingletonGameViewPresenterImpl) SetUnitStats(ctx context.Context, unit *v1.Unit) {
	content := s.renderPanelTemplate(ctx, "UnitStatsPanel.templar.html", map[string]any{
		"Unit":       unit,
		"RulesTable": s.RulesEngine,
		"Theme":      s.Theme, // Pass theme to template
	})
	s.GameViewerPage.SetUnitStatsContent(ctx, &v1.SetContentRequest{
		InnerHtml: content,
	})
}

func (s *SingletonGameViewPresenterImpl) SetUnitDamageDistribution(ctx context.Context, unit *v1.Unit) {
	content := s.renderPanelTemplate(ctx, "DamageDistributionPanel.templar.html", map[string]any{
		"Unit":       unit,
		"RulesTable": s.RulesEngine,
		"Theme":      s.Theme, // Pass theme to template
	})
	s.GameViewerPage.SetDamageDistributionContent(ctx, &v1.SetContentRequest{
		InnerHtml: content,
	})
}

func (s *SingletonGameViewPresenterImpl) SetTerrainStats(ctx context.Context, tile *v1.Tile) {
	content := s.renderPanelTemplate(ctx, "TerrainStatsPanel.templar.html", map[string]any{
		"Tile":       tile,
		"RulesTable": s.RulesEngine,
		"Theme":      s.Theme, // Pass theme to template
	})
	s.GameViewerPage.SetTerrainStatsContent(ctx, &v1.SetContentRequest{
		InnerHtml: content,
	})
}

func (s *SingletonGameViewPresenterImpl) SetTurnOptions(ctx context.Context, response *v1.GetOptionsAtResponse, unit *v1.Unit) {
	content := s.renderPanelTemplate(ctx, "TurnOptionsPanel.templar.html", map[string]any{
		"Options": response.GetOptions(),
		"Unit":    unit,
		"Theme":   s.Theme,
	})
	s.GameViewerPage.SetTurnOptionsContent(ctx, &v1.SetContentRequest{
		InnerHtml: content,
	})
}

// clearHighlightsAndSelection clears highlights if any are currently shown
func (s *SingletonGameViewPresenterImpl) clearHighlightsAndSelection(ctx context.Context) {
	if s.hasHighlights {
		s.GameViewerPage.ClearHighlights(ctx, &v1.ClearHighlightsRequest{
			Types: []string{}, // Empty = clear all
		})
		s.hasHighlights = false
		s.selectedQ = nil
		s.selectedR = nil
	}
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
				Q:    moveOpt.Q,
				R:    moveOpt.R,
				Type: "movement",
			})
		} else if attackOpt := option.GetAttack(); attackOpt != nil {
			// Add attack highlight
			highlights = append(highlights, &v1.HighlightSpec{
				Q:    attackOpt.Q,
				R:    attackOpt.R,
				Type: "attack",
			})
		}
	}

	return highlights
}
