package weewar

import (
	"fmt"
)

// =============================================================================
// Combat Prediction and Analysis System
// =============================================================================

// DamagePrediction represents combat damage prediction
type DamagePrediction struct {
	MinDamage      int                `json:"minDamage"`
	MaxDamage      int                `json:"maxDamage"`
	ExpectedDamage float64            `json:"expectedDamage"`
	Probabilities  map[int]float64    `json:"probabilities"`
}

// MovePrediction represents movement prediction and analysis
type MovePrediction struct {
	CanMove        bool                    `json:"canMove"`
	MovementCost   int                     `json:"movementCost"`
	RemainingMoves int                     `json:"remainingMoves"`
	Path           []PredictionPosition    `json:"path"`
	TerrainEffects []string                `json:"terrainEffects"`
}

// PredictionPosition represents a map position for predictions
type PredictionPosition struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

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
func (cp *CombatPredictor) PredictDamage(game *Game, fromRow, fromCol, toRow, toCol int) (*DamagePrediction, error) {
	// Get attacker and target units
	attacker := game.GetUnitAt(fromRow, fromCol)
	target := game.GetUnitAt(toRow, toCol)
	
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
func (cp *CombatPredictor) CanAttackPosition(game *Game, fromRow, fromCol, toRow, toCol int) (bool, error) {
	return game.CanAttack(fromRow, fromCol, toRow, toCol)
}

// GetAttackOptions returns all positions a unit can attack from its current position
func (cp *CombatPredictor) GetAttackOptions(game *Game, unitRow, unitCol int) ([]PredictionPosition, error) {
	unit := game.GetUnitAt(unitRow, unitCol)
	if unit == nil {
		return nil, fmt.Errorf("no unit at position (%d, %d)", unitRow, unitCol)
	}
	
	var attackPositions []PredictionPosition
	
	// Check all possible positions within attack range
	// For now, assume attack range is 1 (adjacent tiles)
	// TODO: Get actual attack range from unit data
	for dRow := -1; dRow <= 1; dRow++ {
		for dCol := -1; dCol <= 1; dCol++ {
			if dRow == 0 && dCol == 0 {
				continue // Skip self
			}
			
			targetRow := unitRow + dRow
			targetCol := unitCol + dCol
			
			// Check if attack is valid
			if canAttack, err := game.CanAttack(unitRow, unitCol, targetRow, targetCol); err == nil && canAttack {
				attackPositions = append(attackPositions, PredictionPosition{Row: targetRow, Col: targetCol})
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
func (mp *MovementPredictor) PredictMovement(game *Game, fromRow, fromCol, toRow, toCol int) (*MovePrediction, error) {
	unit := game.GetUnitAt(fromRow, fromCol)
	if unit == nil {
		return nil, fmt.Errorf("no unit at position (%d, %d)", fromRow, fromCol)
	}
	
	// Check if movement is valid
	canMove, err := game.CanMove(fromRow, fromCol, toRow, toCol)
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
	prediction.Path = append(prediction.Path, PredictionPosition{Row: fromRow, Col: fromCol})
	prediction.Path = append(prediction.Path, PredictionPosition{Row: toRow, Col: toCol})
	
	// Analyze terrain effects
	targetTile := game.GetTileAt(toRow, toCol)
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
func (mp *MovementPredictor) GetMovementOptions(game *Game, unitRow, unitCol int) ([]PredictionPosition, error) {
	unit := game.GetUnitAt(unitRow, unitCol)
	if unit == nil {
		return nil, fmt.Errorf("no unit at position (%d, %d)", unitRow, unitCol)
	}
	
	var movePositions []PredictionPosition
	
	// Simple implementation: check positions within movement range
	// TODO: Use proper pathfinding with movement costs
	maxRange := unit.DistanceLeft
	
	for dRow := -maxRange; dRow <= maxRange; dRow++ {
		for dCol := -maxRange; dCol <= maxRange; dCol++ {
			if dRow == 0 && dCol == 0 {
				continue // Skip self
			}
			
			targetRow := unitRow + dRow
			targetCol := unitCol + dCol
			
			// Check if movement is valid
			if canMove, err := game.CanMove(unitRow, unitCol, targetRow, targetCol); err == nil && canMove {
				movePositions = append(movePositions, PredictionPosition{Row: targetRow, Col: targetCol})
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