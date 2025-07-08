package weewar

import (
	"fmt"

	"github.com/panyam/turnengine/internal/turnengine"
)

type WeeWarGame struct {
	gameState     *turnengine.GameState
	gameEngine    *turnengine.GameEngine
	combatSystem  *WeeWarCombatSystem
	movementSystem *WeeWarMovementSystem
	board         *HexBoard
	unitData      map[string]UnitData
	terrainData   map[string]TerrainData
}

type WeeWarConfig struct {
	BoardWidth    int                 `json:"boardWidth"`
	BoardHeight   int                 `json:"boardHeight"`
	Players       []WeeWarPlayer      `json:"players"`
	StartingUnits map[string][]string `json:"startingUnits"`
	TerrainMap    [][]string          `json:"terrainMap"`
	MapData       *MapData            `json:"mapData,omitempty"`
}

type WeeWarPlayer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Team  int    `json:"team"`
}

func NewWeeWarGame(config WeeWarConfig) (*WeeWarGame, error) {
	// Load WeeWar data
	data, err := loadWeeWarData()
	if err != nil {
		return nil, fmt.Errorf("failed to load WeeWar data: %w", err)
	}

	// Create game systems
	combatSystem, err := NewWeeWarCombatSystem()
	if err != nil {
		return nil, fmt.Errorf("failed to create combat system: %w", err)
	}

	movementSystem, err := NewWeeWarMovementSystem()
	if err != nil {
		return nil, fmt.Errorf("failed to create movement system: %w", err)
	}

	// Create hex board
	board := NewHexBoard(config.BoardWidth, config.BoardHeight)

	// Create game engine
	gameEngine := turnengine.NewGameEngine()

	// Register command handlers
	registerCommandHandlers(gameEngine, combatSystem, movementSystem)

	// Create players
	players := make([]turnengine.Player, len(config.Players))
	for i, p := range config.Players {
		players[i] = turnengine.Player{
			ID:        p.ID,
			Name:      p.Name,
			Team:      p.Team,
			Status:    "active",
			Resources: map[string]interface{}{
				"funds": 10000,
			},
			Metadata: make(map[string]interface{}),
		}
	}

	// Create game state
	gameState := turnengine.NewGameState("weewar", players, data)

	// Initialize world with components and systems
	world := gameState.World
	registerWeeWarComponents(world)
	world.RegisterSystem(combatSystem)
	world.RegisterSystem(movementSystem)

	game := &WeeWarGame{
		gameState:      gameState,
		gameEngine:     gameEngine,
		combatSystem:   combatSystem,
		movementSystem: movementSystem,
		board:          board,
		unitData:       make(map[string]UnitData),
		terrainData:    make(map[string]TerrainData),
	}

	// Index unit and terrain data
	for _, unit := range data.Units {
		game.unitData[unit.Name] = unit
	}
	for _, terrain := range data.Terrains {
		game.terrainData[terrain.Name] = terrain
	}

	// Initialize board terrain
	if err := game.initializeTerrain(config.TerrainMap); err != nil {
		return nil, fmt.Errorf("failed to initialize terrain: %w", err)
	}

	// Initialize starting units
	if err := game.initializeStartingUnits(config.StartingUnits); err != nil {
		return nil, fmt.Errorf("failed to initialize starting units: %w", err)
	}

	return game, nil
}

func (wg *WeeWarGame) GetGameState() *turnengine.GameState {
	return wg.gameState
}

func (wg *WeeWarGame) GetBoard() *HexBoard {
	return wg.board
}

func (wg *WeeWarGame) ProcessCommand(command *turnengine.Command) error {
	return wg.gameState.ProcessCommand(wg.gameEngine, command)
}

func (wg *WeeWarGame) GetUnitAt(pos *HexPosition) *turnengine.Entity {
	entityID, exists := wg.board.GetEntityAt(pos)
	if !exists {
		return nil
	}

	entity, exists := wg.gameState.World.GetEntity(entityID)
	if !exists {
		return nil
	}

	return entity
}

func (wg *WeeWarGame) GetVisibleUnits(playerID string) []*turnengine.Entity {
	var visibleUnits []*turnengine.Entity

	// Get all entities belonging to this player
	playerEntities := wg.gameState.World.QueryEntities("team", "position")
	
	for _, entity := range playerEntities {
		team, exists := entity.GetComponent("team")
		if !exists {
			continue
		}

		teamID, ok := team["teamId"].(float64)
		if !ok {
			continue
		}

		// Find player by team ID
		player, exists := wg.gameState.GetPlayer(playerID)
		if !exists || player.Team != int(teamID) {
			continue
		}

		// Get unit position
		pos, err := wg.getEntityPosition(entity)
		if err != nil {
			continue
		}

		// Calculate visible positions for this unit
		sightRange := wg.getUnitSightRange(entity)
		visiblePositions := wg.board.GetVisiblePositions(pos, sightRange)

		// Check for enemy units in visible positions
		for _, visiblePos := range visiblePositions {
			if enemyUnit := wg.GetUnitAt(visiblePos.(*HexPosition)); enemyUnit != nil {
				enemyTeam, exists := enemyUnit.GetComponent("team")
				if !exists {
					continue
				}

				enemyTeamID, ok := enemyTeam["teamId"].(float64)
				if !ok || int(enemyTeamID) == player.Team {
					continue
				}

				// Add enemy unit to visible units
				visibleUnits = append(visibleUnits, enemyUnit)
			}
		}
	}

	return visibleUnits
}

func (wg *WeeWarGame) initializeTerrain(terrainMap [][]string) error {
	for y, row := range terrainMap {
		for x, terrainType := range row {
			pos := &HexPosition{Q: x, R: y}
			if err := wg.board.SetTerrain(pos, terrainType); err != nil {
				return fmt.Errorf("failed to set terrain at (%d,%d): %w", x, y, err)
			}
		}
	}
	return nil
}

func (wg *WeeWarGame) initializeStartingUnits(startingUnits map[string][]string) error {
	for playerID, units := range startingUnits {
		// Find player
		player, exists := wg.gameState.GetPlayer(playerID)
		if !exists {
			return fmt.Errorf("player not found: %s", playerID)
		}

		// Create starting units for this player
		for i, unitType := range units {
			entity := wg.gameState.World.CreateEntity("")
			
			// Add components
			if err := wg.addUnitComponents(entity, unitType, player.Team); err != nil {
				return fmt.Errorf("failed to add components for unit %s: %w", unitType, err)
			}

			// Place unit on board (simple placement for now)
			startPos := &HexPosition{Q: i, R: player.Team}
			if err := wg.placeUnitOnBoard(entity, startPos); err != nil {
				return fmt.Errorf("failed to place unit on board: %w", err)
			}
		}
	}
	return nil
}

func (wg *WeeWarGame) addUnitComponents(entity *turnengine.Entity, unitType string, teamID int) error {
	unitData, exists := wg.unitData[unitType]
	if !exists {
		return fmt.Errorf("unknown unit type: %s", unitType)
	}

	// Add position component (will be set when placed on board)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0, Z: 0})

	// Add health component
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	// Add movement component
	entity.AddComponent(&MovementComponent{
		Range:     unitData.BaseStats.Movement,
		MovesLeft: unitData.BaseStats.Movement,
	})

	// Add combat component
	entity.AddComponent(&CombatComponent{
		Attack:  unitData.BaseStats.Attack,
		Defense: unitData.BaseStats.Defense,
	})

	// Add unit type component
	entity.AddComponent(&UnitTypeComponent{
		UnitType: unitType,
		Cost:     unitData.BaseStats.Cost,
	})

	// Add team component
	entity.AddComponent(&TeamComponent{TeamID: teamID})

	return nil
}

func (wg *WeeWarGame) placeUnitOnBoard(entity *turnengine.Entity, pos *HexPosition) error {
	// Update entity position
	position, exists := entity.GetComponent("position")
	if !exists {
		return fmt.Errorf("entity has no position component")
	}

	position["x"] = float64(pos.Q)
	position["y"] = float64(pos.R)
	position["z"] = 0.0
	entity.Components["position"] = position

	// Update board tracking
	return wg.board.SetEntityAt(pos, entity.ID)
}

func (wg *WeeWarGame) getEntityPosition(entity *turnengine.Entity) (*HexPosition, error) {
	position, exists := entity.GetComponent("position")
	if !exists {
		return nil, fmt.Errorf("entity has no position component")
	}

	x, xOk := position["x"].(float64)
	y, yOk := position["y"].(float64)
	if !xOk || !yOk {
		return nil, fmt.Errorf("invalid position component")
	}

	return &HexPosition{Q: int(x), R: int(y)}, nil
}

func (wg *WeeWarGame) getUnitSightRange(entity *turnengine.Entity) int {
	// Get unit type
	unitType, exists := entity.GetComponent("unitType")
	if !exists {
		return 2 // Default sight range
	}

	unitTypeName, ok := unitType["unitType"].(string)
	if !ok {
		return 2
	}

	// Get sight range from unit data
	unitData, exists := wg.unitData[unitTypeName]
	if !exists {
		return 2
	}

	if unitData.BaseStats.SightRange > 0 {
		return unitData.BaseStats.SightRange
	}

	return 2 // Default sight range
}

func registerWeeWarComponents(world *turnengine.World) {
	RegisterWeeWarComponents(world.ComponentRegistry)
}

func registerCommandHandlers(engine *turnengine.GameEngine, combatSystem *WeeWarCombatSystem, movementSystem *WeeWarMovementSystem) {
	// Register move command
	moveValidator := &MoveCommandValidator{movementSystem: movementSystem}
	moveProcessor := &MoveCommandProcessor{movementSystem: movementSystem}
	engine.RegisterCommandHandler("move", moveValidator, moveProcessor)

	// Register attack command
	attackValidator := &AttackCommandValidator{combatSystem: combatSystem}
	attackProcessor := &AttackCommandProcessor{combatSystem: combatSystem}
	engine.RegisterCommandHandler("attack", attackValidator, attackProcessor)
}

// Command validators and processors
type MoveCommandValidator struct {
	movementSystem *WeeWarMovementSystem
}

func (mcv *MoveCommandValidator) ValidateCommand(gameState *turnengine.GameState, command *turnengine.Command) error {
	// Validate move command
	unitID, ok := command.Data["unitId"].(string)
	if !ok {
		return fmt.Errorf("missing unitId in move command")
	}

	_, exists := gameState.World.GetEntity(unitID)
	if !exists {
		return fmt.Errorf("unit not found: %s", unitID)
	}

	// More validation logic here...
	return nil
}

type MoveCommandProcessor struct {
	movementSystem *WeeWarMovementSystem
}

func (mcp *MoveCommandProcessor) ProcessCommand(gameState *turnengine.GameState, command *turnengine.Command) error {
	// Process move command
	// Implementation details...
	return nil
}

type AttackCommandValidator struct {
	combatSystem *WeeWarCombatSystem
}

func (acv *AttackCommandValidator) ValidateCommand(gameState *turnengine.GameState, command *turnengine.Command) error {
	// Validate attack command
	// Implementation details...
	return nil
}

type AttackCommandProcessor struct {
	combatSystem *WeeWarCombatSystem
}

func (acp *AttackCommandProcessor) ProcessCommand(gameState *turnengine.GameState, command *turnengine.Command) error {
	// Process attack command
	// Implementation details...
	return nil
}