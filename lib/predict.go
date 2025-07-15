package weewar

import (
	"fmt"
)

// =============================================================================
// Combat Prediction and Analysis System
// =============================================================================

// DamagePrediction represents combat damage prediction
type DamagePrediction struct {
	MinDamage      int             `json:"minDamage"`
	MaxDamage      int             `json:"maxDamage"`
	ExpectedDamage float64         `json:"expectedDamage"`
	Probabilities  map[int]float64 `json:"probabilities"`
}

// MovePrediction represents movement prediction and analysis
type MovePrediction struct {
	CanMove        bool                 `json:"canMove"`
	MovementCost   int                  `json:"movementCost"`
	RemainingMoves int                  `json:"remainingMoves"`
	Path           []PredictionPosition `json:"path"`
	TerrainEffects []string             `json:"terrainEffects"`
}

// PredictionPosition represents a map position for predictions
type PredictionPosition = CubeCoord

// CombatPredictor provides combat analysis capabilities
type CombatPredictor struct {
	assetManager *AssetManager
}

// NewCombatPredictor creates a new combat predictor
func NewCombatPredictor(assetManager *AssetManager) *CombatPredictor {
	return &CombatPredictor{
		assetManager: assetManager,
	}
}

// PredictDamage calculates damage prediction for an attack
func (cp *CombatPredictor) PredictDamage(game *Game, from, to CubeCoord) (*DamagePrediction, error) {
	// Get attacker and target units
	attacker := game.GetUnitAt(from)
	target := game.GetUnitAt(to)

	if attacker == nil {
		return nil, fmt.Errorf("no attacker unit at position")
	}
	if target == nil {
		return nil, fmt.Errorf("no target unit at position")
	}

	// Get unit data for both units
	attackerData, err := cp.assetManager.GetUnitData(attacker.UnitType)
	if err != nil {
		return nil, fmt.Errorf("failed to get attacker data: %w", err)
	}

	targetData, err := cp.assetManager.GetUnitData(target.UnitType)
	if err != nil {
		return nil, fmt.Errorf("failed to get target data: %w", err)
	}

	// Get damage distribution from attack matrix
	damageDistribution, exists := attackerData.AttackMatrix[targetData.Name]
	if !exists {
		return nil, fmt.Errorf("no attack data available for %s vs %s", attackerData.Name, targetData.Name)
	}

	// Calculate expected damage
	expectedDamage := 0.0
	probabilities := make(map[int]float64)

	for damageStr, probability := range damageDistribution.Probabilities {
		damage := 0
		fmt.Sscanf(damageStr, "%d", &damage)
		expectedDamage += float64(damage) * probability
		probabilities[damage] = probability
	}

	return &DamagePrediction{
		MinDamage:      damageDistribution.MinDamage,
		MaxDamage:      damageDistribution.MaxDamage,
		ExpectedDamage: expectedDamage,
		Probabilities:  probabilities,
	}, nil
}

// CanAttackPosition checks if an attack is valid and possible
func (cp *CombatPredictor) CanAttackPosition(game *Game, from, to CubeCoord) (bool, error) {
	return game.CanAttack(from, to)
}

// GetAttackOptions returns all positions a unit can attack from its current position
func (cp *CombatPredictor) GetAttackOptions(game *Game, unitPos Position) ([]PredictionPosition, error) {
	unit := game.GetUnitAt(unitPos)
	if unit == nil {
		return nil, fmt.Errorf("no unit at position %s", unitPos)
	}

	var attackPositions []PredictionPosition

	// Check all possible positions within attack range
	// For now, assume attack range is 1 (adjacent tiles)
	// TODO: Get actual attack range from unit data
	for dQ := -1; dQ <= 1; dQ++ {
		for dR := -1; dR <= 1; dR++ {
			if dQ == 0 && dR == 0 {
				continue // Skip self
			}

			target := unitPos.Plus(dQ, dR)

			// Check if attack is valid
			if canAttack, err := game.CanAttack(unitPos, target); err == nil && canAttack {
				attackPositions = append(attackPositions, target)
			}
		}
	}

	return attackPositions, nil
}

// MovementPredictor provides movement analysis capabilities
type MovementPredictor struct {
	assetManager *AssetManager
}

// NewMovementPredictor creates a new movement predictor
func NewMovementPredictor(assetManager *AssetManager) *MovementPredictor {
	return &MovementPredictor{
		assetManager: assetManager,
	}
}

// PredictMovement analyzes a potential movement
func (mp *MovementPredictor) PredictMovement(game *Game, from, to Position) (*MovePrediction, error) {
	unit := game.GetUnitAt(from)
	if unit == nil {
		return nil, fmt.Errorf("no unit at position %s", from)
	}

	// Check if movement is valid
	canMove, err := game.CanMove(from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to check movement: %w", err)
	}

	prediction := &MovePrediction{
		CanMove:        canMove,
		MovementCost:   0,
		RemainingMoves: unit.DistanceLeft,
		Path:           []PredictionPosition{},
		TerrainEffects: []string{},
	}

	if !canMove {
		return prediction, nil
	}

	// Calculate movement cost and path
	// TODO: Implement pathfinding to get actual path and cost
	// For now, use simple distance calculation
	prediction.MovementCost = 1 // Simplified
	prediction.RemainingMoves = unit.DistanceLeft - prediction.MovementCost

	// Add start and end positions to path
	prediction.Path = append(prediction.Path, from)
	prediction.Path = append(prediction.Path, to)

	// Analyze terrain effects
	targetTile := game.World.Map.TileAt(to)
	if targetTile != nil {
		terrainData, err := mp.assetManager.GetTerrainDataAsset(targetTile.TileType)
		if err == nil {
			prediction.TerrainEffects = append(prediction.TerrainEffects,
				fmt.Sprintf("Terrain: %s", terrainData.Name))
		}
	}

	return prediction, nil
}

// GetMovementOptions returns all positions a unit can move to
func (mp *MovementPredictor) GetMovementOptions(game *Game, unitPos Position) ([]PredictionPosition, error) {
	unit := game.GetUnitAt(unitPos)
	if unit == nil {
		return nil, fmt.Errorf("no unit at position %s", unitPos)
	}

	var movePositions []PredictionPosition

	// Simple implementation: check positions within movement range
	// TODO: Use proper pathfinding with movement costs
	maxRange := unit.DistanceLeft

	for dQ := -maxRange; dQ <= maxRange; dQ++ {
		for dR := -maxRange; dR <= maxRange; dR++ {
			if dQ == 0 && dR == 0 {
				continue // Skip self
			}

			target := unitPos.Plus(dQ, dR)

			// Check if movement is valid
			if canMove, err := game.CanMove(unitPos, target); err == nil && canMove {
				movePositions = append(movePositions, target)
			}
		}
	}

	return movePositions, nil
}

// GamePredictor combines combat and movement prediction
type GamePredictor struct {
	combat   *CombatPredictor
	movement *MovementPredictor
}

// NewGamePredictor creates a comprehensive game predictor
func NewGamePredictor(assetManager *AssetManager) *GamePredictor {
	return &GamePredictor{
		combat:   NewCombatPredictor(assetManager),
		movement: NewMovementPredictor(assetManager),
	}
}

// GetCombatPredictor returns the combat predictor
func (gp *GamePredictor) GetCombatPredictor() *CombatPredictor {
	return gp.combat
}

// GetMovementPredictor returns the movement predictor
func (gp *GamePredictor) GetMovementPredictor() *MovementPredictor {
	return gp.movement
}
