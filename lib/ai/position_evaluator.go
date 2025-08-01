package ai

import (
	"math"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
)

// =============================================================================
// Position Evaluator
// =============================================================================

// PositionEvaluator provides comprehensive game position analysis
type PositionEvaluator struct {
	// Configuration
	weights *EvaluationWeights

	// Cached calculations for performance
	threatCache      map[string][]Threat
	opportunityCache map[string][]Opportunity

	// Game data references
	rulesEngine *weewar.RulesEngine
}

// EvaluationWeights configures the importance of different position factors
type EvaluationWeights struct {
	// Material evaluation weights (40% total)
	UnitValue       float64 `json:"unitValue"`       // 0.25 - Raw unit strength
	UnitHealth      float64 `json:"unitHealth"`      // 0.10 - Unit condition
	UnitPositioning float64 `json:"unitPositioning"` // 0.05 - Tactical positions

	// Economic evaluation weights (35% total)
	BaseControl   float64 `json:"baseControl"`   // 0.20 - Production capacity
	IncomeControl float64 `json:"incomeControl"` // 0.15 - Economic advantage

	// Strategic evaluation weights (15% total)
	TerritoryControl float64 `json:"territoryControl"` // 0.05 - privateMap control
	ThreatLevel      float64 `json:"threatLevel"`      // 0.05 - Defensive concerns
	AttackOptions    float64 `json:"attackOptions"`    // 0.05 - Offensive potential

	// Positional evaluation weights (10% total)
	MobilityFactor float64 `json:"mobilityFactor"` // 0.05 - Movement flexibility
	SupportNetwork float64 `json:"supportNetwork"` // 0.05 - Unit coordination
}

// =============================================================================
// Constructor and Configuration
// =============================================================================

// NewPositionEvaluator creates a new position evaluator with balanced weights
func NewPositionEvaluator(rulesEngine *weewar.RulesEngine) *PositionEvaluator {
	pe := &PositionEvaluator{
		weights:          NewBalancedWeights(),
		threatCache:      make(map[string][]Threat),
		opportunityCache: make(map[string][]Opportunity),
		rulesEngine:      rulesEngine,
	}

	return pe
}

// NewBalancedWeights returns default balanced evaluation weights
func NewBalancedWeights() *EvaluationWeights {
	return &EvaluationWeights{
		UnitValue:        0.25,
		UnitHealth:       0.10,
		UnitPositioning:  0.05,
		BaseControl:      0.20,
		IncomeControl:    0.15,
		TerritoryControl: 0.05,
		ThreatLevel:      0.05,
		AttackOptions:    0.05,
		MobilityFactor:   0.05,
		SupportNetwork:   0.05,
	}
}

// NewAggressiveWeights returns weights optimized for aggressive play
func NewAggressiveWeights() *EvaluationWeights {
	weights := NewBalancedWeights()
	weights.AttackOptions = 0.15 // Increased focus on offense
	weights.ThreatLevel = 0.02   // Less defensive concern
	weights.UnitValue = 0.20     // Reduced material conservation
	weights.UnitHealth = 0.08    // Less health concern
	return weights
}

// NewDefensiveWeights returns weights optimized for defensive play
func NewDefensiveWeights() *EvaluationWeights {
	weights := NewBalancedWeights()
	weights.ThreatLevel = 0.15    // High defensive priority
	weights.AttackOptions = 0.02  // Reduced aggression
	weights.BaseControl = 0.25    // Focus on holding territory
	weights.SupportNetwork = 0.08 // Better unit coordination
	return weights
}

// NewEconomicWeights returns weights optimized for economic play
func NewEconomicWeights() *EvaluationWeights {
	weights := NewBalancedWeights()
	weights.IncomeControl = 0.25 // Economic expansion priority
	weights.BaseControl = 0.25   // Production capacity focus
	weights.UnitValue = 0.18     // Less unit-focused
	weights.AttackOptions = 0.02 // Conservative attacks
	return weights
}

// SetWeights configures the evaluator with custom weights
func (pe *PositionEvaluator) SetWeights(weights *EvaluationWeights) {
	pe.weights = weights
}

// =============================================================================
// Core Evaluation Methods
// =============================================================================

// EvaluatePosition returns comprehensive position evaluation
func (pe *PositionEvaluator) EvaluatePosition(game *weewar.Game, playerID int32) *PositionEvaluation {
	eval := &PositionEvaluation{
		ComponentScores: make(map[string]float64),
		Strengths:       make([]string, 0),
		Weaknesses:      make([]string, 0),
		KeyFactors:      make([]string, 0),
	}

	// Calculate component scores
	materialScore := pe.evaluateMaterial(game, playerID)
	economicScore := pe.evaluateEconomic(game, playerID)
	tacticalScore := pe.evaluateTactical(game, playerID)
	strategicScore := pe.evaluateStrategic(game, playerID)

	// Weight and combine scores
	eval.MaterialScore = materialScore
	eval.EconomicScore = economicScore
	eval.TacticalScore = tacticalScore
	eval.StrategicScore = strategicScore

	eval.OverallScore =
		materialScore*0.40 +
			economicScore*0.35 +
			tacticalScore*0.15 +
			strategicScore*0.10

	// Store detailed component scores
	eval.ComponentScores["unit_value"] = pe.evaluateUnitValue(game, playerID)
	eval.ComponentScores["unit_health"] = pe.evaluateUnitHealth(game, playerID)
	eval.ComponentScores["unit_positioning"] = pe.evaluateUnitPositioning(game, playerID)
	eval.ComponentScores["base_control"] = pe.evaluateBaseControl(game, playerID)
	eval.ComponentScores["income_control"] = pe.evaluateIncomeControl(game, playerID)
	eval.ComponentScores["territory_control"] = pe.evaluateTerritoryControl(game, playerID)
	eval.ComponentScores["threat_level"] = pe.evaluateThreatLevel(game, playerID)
	eval.ComponentScores["attack_options"] = pe.evaluateAttackOptions(game, playerID)
	eval.ComponentScores["mobility"] = pe.evaluateMobility(game, playerID)
	eval.ComponentScores["support_network"] = pe.evaluateSupportNetwork(game, playerID)

	// Analyze strengths and weaknesses
	pe.analyzeStrengthsWeaknesses(eval)

	// Set confidence based on evaluation clarity
	eval.Confidence = pe.calculateConfidence(eval)

	return eval
}

// =============================================================================
// Material Evaluation (40% weight)
// =============================================================================

func (pe *PositionEvaluator) evaluateMaterial(game *weewar.Game, playerID int32) float64 {
	unitValue := pe.evaluateUnitValue(game, playerID) * pe.weights.UnitValue
	unitHealth := pe.evaluateUnitHealth(game, playerID) * pe.weights.UnitHealth
	unitPositioning := pe.evaluateUnitPositioning(game, playerID) * pe.weights.UnitPositioning

	return (unitValue + unitHealth + unitPositioning) /
		(pe.weights.UnitValue + pe.weights.UnitHealth + pe.weights.UnitPositioning)
}

func (pe *PositionEvaluator) evaluateUnitValue(game *weewar.Game, playerID int32) float64 {
	// playerUnits := game.GetUnitsForPlayer(int(playerID))
	playerValue := 0.0
	totalValue := 0.0

	// Calculate total unit values for all players
	for pid := range game.World.PlayerCount() {
		units := game.GetUnitsForPlayer(int(pid))
		for _, unit := range units {
			unitCost := pe.getUnitCost(unit.UnitType)
			if pid == playerID {
				playerValue += unitCost
			}
			totalValue += unitCost
		}
	}

	if totalValue == 0 {
		return 0.5 // Neutral if no units
	}

	return playerValue / totalValue
}

func (pe *PositionEvaluator) evaluateUnitHealth(game *weewar.Game, playerID int32) float64 {
	playerUnits := game.GetUnitsForPlayer(int(playerID))
	totalHealthScore := 0.0
	totalValue := 0.0

	for _, unit := range playerUnits {
		healthPercent := float64(unit.AvailableHealth) / 100.0
		unitValue := pe.getUnitCost(unit.UnitType)

		totalHealthScore += healthPercent * unitValue
		totalValue += unitValue
	}

	if totalValue == 0 {
		return 0.5 // Neutral if no units
	}

	return totalHealthScore / totalValue
}

func (pe *PositionEvaluator) evaluateUnitPositioning(game *weewar.Game, playerID int32) float64 {
	playerUnits := game.GetUnitsForPlayer(int(playerID))
	totalPositionalScore := 0.0

	for _, unit := range playerUnits {
		positionScore := pe.evaluateUnitPosition(unit, game)
		unitValue := pe.getUnitCost(unit.UnitType)

		totalPositionalScore += positionScore * unitValue
	}

	// Normalize by total unit value
	totalValue := 0.0
	for _, unit := range playerUnits {
		totalValue += pe.getUnitCost(unit.UnitType)
	}

	if totalValue == 0 {
		return 0.5
	}

	return math.Min(totalPositionalScore/totalValue, 1.0)
}

// =============================================================================
// Economic Evaluation (35% weight)
// =============================================================================

func (pe *PositionEvaluator) evaluateEconomic(game *weewar.Game, playerID int32) float64 {
	baseControl := pe.evaluateBaseControl(game, playerID) * pe.weights.BaseControl
	incomeControl := pe.evaluateIncomeControl(game, playerID) * pe.weights.IncomeControl

	return (baseControl + incomeControl) /
		(pe.weights.BaseControl + pe.weights.IncomeControl)
}

func (pe *PositionEvaluator) evaluateBaseControl(game *weewar.Game, playerID int32) float64 {
	controlledBases := 0.0
	totalBases := 0.0

	// Iterate through all terrain tiles to find bases
	if game.World != nil {
		for _, terrain := range game.World.TilesByCoord() {
			if pe.isProductionBase(terrain.TileType) {
				totalBases++
				if pe.isControlledByPlayer(weewar.TileGetCoord(terrain), playerID, game) {
					controlledBases++
				}
			}
		}
	}

	if totalBases == 0 {
		return 0.5 // Neutral if no bases
	}

	return controlledBases / totalBases
}

func (pe *PositionEvaluator) evaluateIncomeControl(game *weewar.Game, playerID int32) float64 {
	controlledCities := 0.0
	totalCities := 0.0

	if game.World != nil {
		for _, terrain := range game.World.TilesByCoord() {
			if pe.isIncomeBuilding(terrain.TileType) {
				totalCities++
				if pe.isControlledByPlayer(weewar.TileGetCoord(terrain), playerID, game) {
					controlledCities++
				}
			}
		}
	}

	if totalCities == 0 {
		return 0.5 // Neutral if no income buildings
	}

	return controlledCities / totalCities
}

// =============================================================================
// Tactical Evaluation (15% weight)
// =============================================================================

func (pe *PositionEvaluator) evaluateTactical(game *weewar.Game, playerID int32) float64 {
	territoryControl := pe.evaluateTerritoryControl(game, playerID) * pe.weights.TerritoryControl
	threatLevel := pe.evaluateThreatLevel(game, playerID) * pe.weights.ThreatLevel
	attackOptions := pe.evaluateAttackOptions(game, playerID) * pe.weights.AttackOptions

	return (territoryControl + threatLevel + attackOptions) /
		(pe.weights.TerritoryControl + pe.weights.ThreatLevel + pe.weights.AttackOptions)
}

func (pe *PositionEvaluator) evaluateTerritoryControl(game *weewar.Game, playerID int32) float64 {
	// Simple territory control based on unit positions
	// TODO: Implement proper territory influence calculation
	return 0.5
}

func (pe *PositionEvaluator) evaluateThreatLevel(game *weewar.Game, playerID int32) float64 {
	threats := pe.identifyThreats(game, playerID)
	totalThreat := 0.0

	for _, threat := range threats {
		targetValue := pe.getUnitCost(threat.TargetUnit.UnitType)
		threatValue := targetValue * threat.ThreatLevel
		totalThreat += threatValue
	}

	// Return inverted score (lower threats = better evaluation)
	maxThreat := pe.getTotalUnitValue(game, playerID)
	if maxThreat == 0 {
		return 1.0
	}

	normalizedThreat := math.Min(totalThreat/maxThreat, 1.0)
	return 1.0 - normalizedThreat
}

func (pe *PositionEvaluator) evaluateAttackOptions(game *weewar.Game, playerID int32) float64 {
	opportunities := pe.identifyOpportunities(game, playerID)
	totalOpportunity := 0.0

	for _, opp := range opportunities {
		if opp.OpportunityType == OpportunityWeakUnit && opp.TargetUnit != nil {
			targetValue := pe.getUnitCost(opp.TargetUnit.UnitType)
			opportunityValue := targetValue * opp.Value
			totalOpportunity += opportunityValue
		}
	}

	// Normalize by enemy total unit value
	enemyValue := 0.0
	for pid := range game.World.PlayerCount() {
		if pid != playerID {
			enemyValue += pe.getTotalUnitValue(game, pid)
		}
	}

	if enemyValue == 0 {
		return 0.0
	}

	return math.Min(totalOpportunity/enemyValue, 1.0)
}

// =============================================================================
// Strategic Evaluation (10% weight)
// =============================================================================

func (pe *PositionEvaluator) evaluateStrategic(game *weewar.Game, playerID int32) float64 {
	mobility := pe.evaluateMobility(game, playerID) * pe.weights.MobilityFactor
	supportNetwork := pe.evaluateSupportNetwork(game, playerID) * pe.weights.SupportNetwork

	return (mobility + supportNetwork) /
		(pe.weights.MobilityFactor + pe.weights.SupportNetwork)
}

func (pe *PositionEvaluator) evaluateMobility(game *weewar.Game, playerID int32) float64 {
	playerUnits := game.GetUnitsForPlayer(int(playerID))
	totalMobility := 0.0

	for _, unit := range playerUnits {
		// Calculate movement options for each unit
		// For now, use simple movement points
		unitData, err := pe.rulesEngine.GetUnitData(unit.UnitType)
		if err != nil {
			panic(err) // we should know about all units and they should be valid
		}
		mobility := float64(unitData.MovementPoints) / 10.0 // Normalize by max expected movement
		totalMobility += mobility
	}

	if len(playerUnits) == 0 {
		return 0.0
	}

	return math.Min(totalMobility/float64(len(playerUnits)), 1.0)
}

func (pe *PositionEvaluator) evaluateSupportNetwork(game *weewar.Game, playerID int32) float64 {
	playerUnits := game.GetUnitsForPlayer(int(playerID))
	supportScore := 0.0

	for _, unit := range playerUnits {
		nearbyAllies := pe.countNearbyAllies(unit, game, playerID)
		supportScore += math.Min(float64(nearbyAllies)/3.0, 1.0) // Cap at 3 nearby allies
	}

	if len(playerUnits) == 0 {
		return 0.0
	}

	return supportScore / float64(len(playerUnits))
}

// =============================================================================
// Helper Methods
// =============================================================================

func (pe *PositionEvaluator) getUnitCost(unitTypeID int32) float64 {
	unitData, err := pe.rulesEngine.GetUnitData(unitTypeID)
	if err != nil {
		panic(err)
	}
	// Use health as a proxy for unit cost (higher health = more expensive units)
	// TODO: Add cost field to UnitDefinition proto or load from separate data source
	return float64(unitData.Health)
}

func (pe *PositionEvaluator) getTotalUnitValue(game *weewar.Game, playerID int32) float64 {
	units := game.GetUnitsForPlayer(int(playerID))
	total := 0.0

	for _, unit := range units {
		total += pe.getUnitCost(unit.UnitType)
	}

	return total
}

func (pe *PositionEvaluator) evaluateUnitPosition(unit *v1.Unit, game *weewar.Game) float64 {
	positionScore := 0.0

	// Terrain defensive bonus
	if game.World != nil {
		if unitAt := game.World.UnitAt(weewar.UnitGetCoord(unit)); unitAt != nil && pe.rulesEngine != nil {
			if terrainData, err := pe.rulesEngine.GetTerrainData(unitAt.UnitType); err == nil {
				positionScore += terrainData.DefenseBonus * 0.3
			}
		}
	}

	// Strategic location value (proximity to objectives)
	positionScore += pe.getStrategicLocationValue(weewar.UnitGetCoord(unit), game) * 0.7

	return math.Min(positionScore, 1.0)
}

func (pe *PositionEvaluator) getStrategicLocationValue(pos weewar.AxialCoord, game *weewar.Game) float64 {
	// Simplified strategic value based on distance to map center
	// TODO: Implement proper strategic location evaluation
	return 0.5
}

func (pe *PositionEvaluator) countNearbyAllies(unit *v1.Unit, game *weewar.Game, playerID int32) int {
	count := 0
	playerUnits := game.GetUnitsForPlayer(int(playerID))

	for _, ally := range playerUnits {
		if weewar.UnitGetCoord(ally) != weewar.UnitGetCoord(unit) {
			distance := pe.calculateDistance(weewar.UnitGetCoord(unit), weewar.UnitGetCoord(ally))
			if distance <= 2.0 { // Within 2 hexes
				count++
			}
		}
	}

	return count
}

func (pe *PositionEvaluator) calculateDistance(pos1, pos2 weewar.AxialCoord) float64 {
	// Hexagonal distance calculation
	dx := float64(pos1.Q - pos2.Q)
	dy := float64(pos1.R - pos2.R)
	return math.Max(math.Abs(dx), math.Max(math.Abs(dy), math.Abs(dx+dy)))
}

func (pe *PositionEvaluator) isProductionBase(terrainTypeID int32) bool {
	panic("not implemented with real types")
	/*
		// Base terrain type IDs (from game data analysis)
		productionBases := map[int]bool{
			10: true, // Land Base
			11: true, // Naval Base
			12: true, // Airport Base
		}

		return productionBases[terrainTypeID]
	*/
}

func (pe *PositionEvaluator) isIncomeBuilding(terrainTypeID int32) bool {
	panic("not implemented")
	/*
		// Income building terrain type IDs
		incomeBuildings := map[int]bool{
			5: true, // City
			6: true, // Hospital (if it provides income)
		}

		return incomeBuildings[terrainTypeID]
	*/
}

func (pe *PositionEvaluator) isControlledByPlayer(pos weewar.AxialCoord, playerID int32, game *weewar.Game) bool {
	// Check if there's a friendly unit on this position
	if game.World != nil {
		if unit := game.World.UnitAt(pos); unit != nil {
			return unit.Player == playerID
		}
	}

	// TODO: Implement proper base/city control mechanics
	return false
}

func (pe *PositionEvaluator) identifyThreats(game *weewar.Game, playerID int32) []Threat {
	// TODO: Implement threat identification
	// This is a placeholder that would analyze enemy units that can attack player units
	return make([]Threat, 0)
}

func (pe *PositionEvaluator) identifyOpportunities(game *weewar.Game, playerID int32) []Opportunity {
	// TODO: Implement opportunity identification
	// This is a placeholder that would analyze weak enemy units, undefended bases, etc.
	return make([]Opportunity, 0)
}

func (pe *PositionEvaluator) analyzeStrengthsWeaknesses(eval *PositionEvaluation) {
	// Analyze component scores to identify strengths and weaknesses
	for component, score := range eval.ComponentScores {
		if score > 0.7 {
			eval.Strengths = append(eval.Strengths, component)
		} else if score < 0.3 {
			eval.Weaknesses = append(eval.Weaknesses, component)
		}

		if score > 0.6 || score < 0.4 {
			eval.KeyFactors = append(eval.KeyFactors, component)
		}
	}
}

func (pe *PositionEvaluator) calculateConfidence(eval *PositionEvaluation) float64 {
	// Calculate confidence based on score variance
	// Higher variance = lower confidence
	scores := make([]float64, 0, len(eval.ComponentScores))
	for _, score := range eval.ComponentScores {
		scores = append(scores, score)
	}

	if len(scores) == 0 {
		return 0.5
	}

	// Calculate variance
	mean := 0.0
	for _, score := range scores {
		mean += score
	}
	mean /= float64(len(scores))

	variance := 0.0
	for _, score := range scores {
		diff := score - mean
		variance += diff * diff
	}
	variance /= float64(len(scores))

	// Convert variance to confidence (lower variance = higher confidence)
	confidence := 1.0 - math.Min(variance*4.0, 1.0) // Scale variance to 0-1 range
	return math.Max(confidence, 0.1)                // Minimum confidence of 0.1
}
