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

			// Always show terrain info (even when unit is present)
			s.SetTerrainStats(ctx, tile)
			s.SetUnitStats(ctx, unit)
			s.SetUnitDamageDistribution(ctx, unit)
		}()

		// If there's a unit, also handle unit logic and show unit info in unit panel
		/*
			if unit != nil {
				s.handleUnitClick(q, r)
				// Update unit stats panel with unit info
			} else {
				// Empty tile clicked - clear selection
				s.clearSelection()
			}
		*/
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

// Handle tile clicks - show terrain info in TerrainStatsPanel
/**
  private handleTileClick(q: number, r: number, tile: any): void {
      if (!this.terrainStatsPanel) {
          console.warn('[GameViewerPage] TerrainStatsPanel not available');
          return;
      }

      // Show terrain info using shared World
      if (tile) {
          const terrainStats = this.rulesTable.getTerrainStatsAt(tile.tileType, tile.player);
          if (terrainStats) {
              // Update with actual coordinates
              const terrainStatsWithCoords = new TerrainStats(
                  terrainStats.terrainDefinition,
                  q,
                  r,
                  tile.player
              );
              this.terrainStatsPanel.updateTerrainStats(terrainStatsWithCoords);
          }
      }
  }
*/

// Handle unit clicks - select unit or show unit info
/**
  private handleUnitClick(q: number, r: number): void {
      // Handle async unit interaction using unified getOptionsAt
      this.gameState.getOptionsAt(q, r).then(async (response: GetOptionsAtResponse) => {
          // ✅ Use shared World for fast unit query
          const unit = this.world?.getUnitAt(q, r);

          // Debug logging
          console.log(`[GameViewerPage] Unit click at (${q}, ${r}):`, {
              unit: unit,
              response: response,
              currentPlayer: this.gameState.getCurrentPlayer(),
              turnCounter: this.gameState.getTurnCounter()
          });

          const options = response.options || [];

          const hasMovementOptions = options.some(opt => opt.move !== undefined);
          const hasAttackOptions = options.some(opt => opt.attack !== undefined);
          const hasOnlyEndTurn = options.length === 1 && options[0].endTurn !== undefined;

          if (hasMovementOptions || hasAttackOptions) {
              // This unit has actionable options - process it directly (no duplicate RPC)
              this.processUnitSelection(q, r, options, response);
          } else if (hasOnlyEndTurn) {
              // This position only has endTurn option - could be empty tile, enemy unit, or friendly unit with no actions

              // ✅ Use shared World for fast queries
              const tileUnit = this.world?.getUnitAt(q, r);

              if (tileUnit) {
                  // Get current player to check ownership
                  this.gameState.getCurrentGameState().then(gameState => {
                      const currentPlayer = gameState.currentPlayer;

                      console.log(`[GameViewerPage] Unit details:`, {
                          unitPlayer: tileUnit.player,
                          currentPlayer: currentPlayer,
                          distanceLeft: tileUnit.distanceLeft,
                          availableHealth: tileUnit.availableHealth,
                          turnCounter: tileUnit.turnCounter,
                          gameTurnCounter: gameState.turnCounter
                      });

                      if (tileUnit.player === currentPlayer) {
                          // This is our unit but it has no available actions
                          this.showToast('Info', `No actions available for unit at (${q}, ${r})`, 'info');
                      } else {
                          // This is an enemy unit
                          this.showToast('Info', `Enemy unit at (${q}, ${r})`, 'info');
                      }
                  }).catch(error => {
                      console.error('Failed to get current game state:', error);
                  });
              } else {
                  this.showToast('Info', `Empty tile at (${q}, ${r})`, 'info');
              }
          }
      }).catch(error => {
          console.error('[GameViewerPage] Failed to get options at position:', error);
      });
  }
*/

//  Handle movement clicks - execute actual unit moves
/**
  private handleMovementClick(q: number, r: number, moveOption: any): void {
      if (this.isProcessingMove) {
          console.warn('[GameViewerPage] Already processing a move, ignoring click');
          this.showToast('Warning', 'Move in progress...', 'warning');
          return;
      }

      if (!this.selectedUnitCoord) {
          console.warn('[GameViewerPage] No unit selected for movement');
          return;
      }

      // Check if clicking on the same position as the selected unit (deselection)
      if (this.selectedUnitCoord.q === q && this.selectedUnitCoord.r === r) {
          console.log('[GameViewerPage] Clicked on selected unit position - deselecting');
          this.clearSelection();
          this.clearAllHighlights();
          return;
      }

      // Execute the move
      this.executeMove(this.selectedUnitCoord, { q, r }, moveOption);
  }
*/

/**
 * Move the currently selected unit to target coordinates
 */
/*
   async moveSelectedUnitTo(q: number, r: number): Promise<ActionResult> {
       try {
           if (!this.selectedUnitCoord) {
               return {
                   success: false,
                   message: 'No unit selected',
                   error: 'Must select a unit before moving'
               };
           }

           if (this.isProcessingMove) {
               return {
                   success: false,
                   message: 'Move already in progress',
                   error: 'Another move is being processed'
               };
           }

           // Find the move option for this target
           const moveOption = this.availableMovementOptions.find(opt => opt.q === q && opt.r === r);
           if (!moveOption) {
               return {
                   success: false,
                   message: `Cannot move to (${q}, ${r}) - not a valid move target`,
                   error: 'Invalid move target',
                   data: {
                       selectedUnit: this.selectedUnitCoord,
                       availableMoves: this.availableMovementOptions.map(opt => ({q: opt.q, r: opt.r}))
                   }
               };
           }

           // Execute the move using existing logic
           const fromCoord = this.selectedUnitCoord;
           await this.executeMove(fromCoord, { q, r }, moveOption, false); // Skip validation for command interface

           return {
               success: true,
               message: `Unit moved from (${fromCoord.q}, ${fromCoord.r}) to (${q}, ${r})`,
               data: { from: fromCoord, to: { q, r } }
           };

       } catch (error) {
           return {
               success: false,
               message: `Failed to move unit to (${q}, ${r})`,
               error: error instanceof Error ? error.message : String(error)
           };
       }
   }
*/

/**
 * End the current player's turn
 * This is the unified method used by both UI clicks and command interface
 */
/*
   async endCurrentPlayerTurn(): Promise<ActionResult> {
       try {
           const currentPlayer = this.gameState.getCurrentPlayer();
           const currentTurn = this.gameState.getTurnCounter();

           // Execute the turn end logic
           await this.gameState.endTurn(currentPlayer);

           // Update UI state
           const newPlayer = this.gameState.getCurrentPlayer();
           const newTurn = this.gameState.getTurnCounter();

           this.updateGameStatus(`Ready - Player ${newPlayer}'s Turn`, newPlayer);
           this.updateTurnCounter(newTurn);
           this.clearUnitSelection();

           if (this.gameLogPanel) {
               this.gameLogPanel.logGameEvent(`Player ${newPlayer}'s turn begins`, 'system');
           }
           this.showToast('Info', `Player ${newPlayer}'s turn`, 'info');

           return {
               success: true,
               message: `Turn ended. Now Player ${newPlayer}'s turn`,
               data: {
                   previousPlayer: currentPlayer,
                   currentPlayer: newPlayer,
                   previousTurn: currentTurn,
                   currentTurn: newTurn
               }
           } as ActionResult;

       } catch (error) {
           const errorMsg = error instanceof Error ? error.message : String(error);
           this.showToast('Error', errorMsg, 'error');
           return {
               success: false,
               message: 'Failed to end turn',
               error: errorMsg
           } as ActionResult;
       }
   }
*/

/**
     * TODO * Use this to update turn options html
		 * Display the current options
*/
/*
   private displayOptions(): void {
       const container = this.findElement('#options-list');
       if (!container) return;

       if (this.currentOptions.length === 0) {
           this.showEmptyOptions();
           return;
       }

       // Hide empty state, show options
       const emptyState = this.findElement('#no-options-selected');
       const optionsContainer = this.findElement('#options-container');
       if (emptyState) emptyState.classList.add('hidden');
       if (optionsContainer) optionsContainer.classList.remove('hidden');

       // Update header
       const headerElement = this.findElement('#options-header');
       if (headerElement && this.selectedPosition) {
           const unitText = this.selectedUnit ? ` (Unit ${this.selectedUnit.unitType})` : '';
           headerElement.textContent = `Options at (${this.selectedPosition.q}, ${this.selectedPosition.r})${unitText}`;
       }

       // Build options HTML
       let optionsHTML = '';
       this.currentOptions.forEach((option, index) => {
           const optionType = this.getOptionType(option);
           const iconClass = this.getOptionIcon(optionType);
           const colorClass = this.getOptionColor(optionType);

           let description = '';
           let details = '';

           if (option.move) {
               description = `Move to (${option.move.q || 0}, ${option.move.r || 0})`;
               if (option.move.movementCost !== undefined) {
                   details += `<span class="text-xs text-gray-500 dark:text-gray-400">Cost: ${option.move.movementCost}</span>`;
               }
           } else if (option.attack) {
               description = `Attack unit at (${option.attack.q || 0}, ${option.attack.r || 0})`;
               if (option.attack.damageEstimate !== undefined) {
                   details += `<span class="text-xs text-red-500 dark:text-red-400">Damage: ~${option.attack.damageEstimate}</span>`;
               }
           } else if (option.endTurn) {
               description = 'End Turn';
           } else if (option.build) {
               description = `Build unit (type ${option.build.unitType})`;
               if (option.build.cost !== undefined) {
                   details += `<span class="text-xs text-gray-500 dark:text-gray-400">Cost: ${option.build.cost}</span>`;
               }
           } else if (option.capture) {
               description = 'Capture';
           }

           optionsHTML += `
               <div class="option-item p-3 mb-2 rounded-lg bg-gray-50 dark:bg-gray-700 hover:bg-gray-100 dark:hover:bg-gray-600 cursor-pointer transition-colors"
                    data-option-index="${index}">
                   <div class="flex items-start">
                       <span class="${iconClass} ${colorClass} mr-3 text-lg">${this.getOptionEmoji(optionType)}</span>
                       <div class="flex-1">
                           <div class="font-medium text-sm text-gray-900 dark:text-white">
                               ${description}
                           </div>
                           ${details ? `<div class="mt-1">${details}</div>` : ''}
                       </div>
                   </div>
               </div>
           `;
       });

       container.innerHTML = optionsHTML;

       // Add click handlers
       container.querySelectorAll('.option-item').forEach(item => {
           item.addEventListener('click', (e) => {
               const index = parseInt((e.currentTarget as HTMLElement).dataset.optionIndex || '0');
               this.handleOptionClick(index);
           });
       });
   }
*/

/**
 * Get icon for option type
 */
/*
   private getOptionIcon(type: string): string {
       switch (type) {
           case 'move': return 'text-blue-500';
           case 'attack': return 'text-red-500';
           case 'endTurn': return 'text-green-500';
           case 'build': return 'text-yellow-500';
           case 'capture': return 'text-purple-500';
           default: return 'text-gray-500';
       }
   }
*/

/**
 * Get color class for option type
 */
/*
   private getOptionColor(type: string): string {
       return this.getOptionIcon(type);
   }
*/

/**
 * Get emoji for option type
 */
/*
   private getOptionEmoji(type: string): string {
       switch (type) {
           case 'move': return '➡️';
           case 'attack': return '⚔️';
           case 'endTurn': return '✅';
           case 'build': return '🏗️';
           case 'capture': return '🏳️';
           default: return '❓';
       }
   }
*/

/**
 * Handle unit selection - fetch and display options (legacy, makes own RPC)
 */
/* Use this when TurnOptionsPanel calls us with "unit clicked"

   public async handleUnitSelection(q: number, r: number, unit: Unit): Promise<void> {
       this.log(`Unit selected at (${q}, ${r})`);
       this.selectedPosition = { q, r };
       this.selectedUnit = unit;

       // Show loading state
       this.showLoadingState();

       try {
           const response = await this.gameState.getOptionsAt(q, r);

           if (!response || !response.options) {
               this.showEmptyOptions();
               return;
           }

           // Process and display options
           this.processOptions(response);
       } catch (error) {
           this.log('Error fetching options:', error);
           this.showError('Failed to fetch options');
       }
   }
*/

/**
 * Handle tile selection - fetch and display options if there's a unit
 */
/*
    * Use this when TurnOptionsPanel calls us with TileClicked
    *
   public async handleTileSelection(q: number, r: number): Promise<void> {
       this.log(`Tile selected at (${q}, ${r})`);

       // Check if there's a unit at this position
       const unit = this.world?.getUnitAt(q, r);
       if (unit) {
           await this.handleUnitSelection(q, r, unit);
       } else {
           this.clearOptions();
       }
   }
*/
