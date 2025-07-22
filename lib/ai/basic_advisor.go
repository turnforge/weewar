package ai

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	weewar "github.com/panyam/turnengine/games/weewar/lib"
)

// =============================================================================
// Basic AI Advisor Implementation
// =============================================================================

// BasicAIAdvisor provides AI decision-making across all difficulty levels
type BasicAIAdvisor struct {
	// Core components
	evaluator  *PositionEvaluator
	strategies map[AIDifficulty]DecisionStrategy

	// Random number generator for deterministic AI (if needed)
	rng *rand.Rand

	// Performance optimization
	moveCache    map[string][]*MoveProposal
	cacheTimeout time.Duration
}

// NewBasicAIAdvisor creates a new AI advisor with all difficulty levels
func NewBasicAIAdvisor(rulesEngine *weewar.RulesEngine) *BasicAIAdvisor {
	evaluator := NewPositionEvaluator(rulesEngine)

	advisor := &BasicAIAdvisor{
		evaluator:    evaluator,
		strategies:   make(map[AIDifficulty]DecisionStrategy),
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
		moveCache:    make(map[string][]*MoveProposal),
		cacheTimeout: time.Minute * 5,
	}

	// Initialize strategies for each difficulty level
	advisor.strategies[AIEasy] = NewEasyStrategy(advisor.rng, evaluator)
	advisor.strategies[AIMedium] = NewMediumStrategy(evaluator, rulesEngine)
	advisor.strategies[AIHard] = NewHardStrategy(evaluator, rulesEngine)
	advisor.strategies[AIExpert] = NewExpertStrategy(evaluator, rulesEngine)

	return advisor
}

// =============================================================================
// AIAdvisor Interface Implementation
// =============================================================================

// SuggestMoves returns AI move recommendations based on difficulty and personality
func (ba *BasicAIAdvisor) SuggestMoves(game *weewar.Game, playerID int, options *AIOptions) (*MoveSuggestions, error) {
	startTime := time.Now()

	// Validate inputs
	if game == nil {
		return nil, fmt.Errorf("game cannot be nil")
	}

	if options == nil {
		options = NewAIOptions()
	}

	// Configure evaluator for personality
	ba.configurePersonality(options.Personality)

	// Get the appropriate strategy
	strategy, exists := ba.strategies[options.Difficulty]
	if !exists {
		return nil, fmt.Errorf("unsupported difficulty level: %v", options.Difficulty)
	}

	// Generate move suggestions using the strategy
	suggestions, err := strategy.SuggestMoves(game, playerID, options)
	if err != nil {
		return nil, fmt.Errorf("strategy failed: %w", err)
	}

	// Add timing information
	suggestions.ThinkingTime = time.Since(startTime)

	// Add overall strategic reasoning if requested
	if options.ShowReasoning {
		suggestions.Reasoning = ba.generateStrategicReasoning(game, playerID, suggestions)
	}

	return suggestions, nil
}

// EvaluatePosition returns comprehensive position analysis
func (ba *BasicAIAdvisor) EvaluatePosition(game *weewar.Game, playerID int) (*PositionEvaluation, error) {
	if game == nil {
		return nil, fmt.Errorf("game cannot be nil")
	}

	return ba.evaluator.EvaluatePosition(game, playerID), nil
}

// GetThreats identifies immediate threats to the player
func (ba *BasicAIAdvisor) GetThreats(game *weewar.Game, playerID int) ([]Threat, error) {
	if game == nil {
		return nil, fmt.Errorf("game cannot be nil")
	}

	return ba.identifyThreats(game, playerID)
}

// GetOpportunities identifies attack and advancement opportunities
func (ba *BasicAIAdvisor) GetOpportunities(game *weewar.Game, playerID int) ([]Opportunity, error) {
	if game == nil {
		return nil, fmt.Errorf("game cannot be nil")
	}

	return ba.identifyOpportunities(game, playerID)
}

// GetStrategicValue returns the strategic value of a specific position
func (ba *BasicAIAdvisor) GetStrategicValue(game *weewar.Game, position weewar.AxialCoord) float64 {
	if game == nil || game.World == nil {
		return 0.0
	}

	return ba.evaluator.getStrategicLocationValue(position, game)
}

// =============================================================================
// Configuration and Personality
// =============================================================================

// configurePersonality adjusts the evaluator weights based on AI personality
func (ba *BasicAIAdvisor) configurePersonality(personality AIPersonality) {
	switch personality {
	case AIAggressive:
		ba.evaluator.SetWeights(NewAggressiveWeights())
	case AIDefensive:
		ba.evaluator.SetWeights(NewDefensiveWeights())
	case AIExpansionist:
		ba.evaluator.SetWeights(NewEconomicWeights())
	case AIBalanced:
		fallthrough
	default:
		ba.evaluator.SetWeights(NewBalancedWeights())
	}
}

// =============================================================================
// Move Generation and Analysis
// =============================================================================

// generateAllValidMoves creates all possible moves for the current player
func (ba *BasicAIAdvisor) generateAllValidMoves(game *weewar.Game, playerID int) ([]*MoveProposal, error) {
	moves := make([]*MoveProposal, 0)

	// Get all units for the player
	playerUnits := game.GetUnitsForPlayer(playerID)

	for _, unit := range playerUnits {
		// Generate movement moves
		if moveMoves, err := ba.generateMovementMoves(game, unit); err == nil {
			moves = append(moves, moveMoves...)
		}

		// Generate attack moves
		if attackMoves, err := ba.generateAttackMoves(game, unit); err == nil {
			moves = append(moves, attackMoves...)
		}

		// TODO: Generate other move types (create unit, capture, repair)
	}

	// Always include end turn as an option
	endTurnMove := &MoveProposal{
		Action:   ActionEndTurn,
		UnitID:   -1,
		From:     weewar.AxialCoord{},
		To:       weewar.AxialCoord{},
		Priority: 0.1, // Low priority by default
		Risk:     0.0,
		Value:    0.0,
		Reason:   "End turn",
		Category: CategoryPositional,
	}
	moves = append(moves, endTurnMove)

	return moves, nil
}

// generateMovementMoves creates movement proposals for a specific unit
func (ba *BasicAIAdvisor) generateMovementMoves(game *weewar.Game, unit *weewar.Unit) ([]*MoveProposal, error) {
	moves := make([]*MoveProposal, 0)

	// Get valid movement positions (this would use the existing game methods)
	// For now, we'll generate some basic moves
	// TODO: Implement proper integration with game.GetUnitMovementOptions(unit)

	// Placeholder: create a few sample movement moves
	// In real implementation, this would use the rules engine

	return moves, nil
}

// generateAttackMoves creates attack proposals for a specific unit
func (ba *BasicAIAdvisor) generateAttackMoves(game *weewar.Game, unit *weewar.Unit) ([]*MoveProposal, error) {
	moves := make([]*MoveProposal, 0)

	// Get valid attack targets (this would use the existing game methods)
	// For now, we'll generate some basic attacks
	// TODO: Implement proper integration with game.GetUnitAttackOptions(unit)

	return moves, nil
}

// =============================================================================
// Threat and Opportunity Analysis
// =============================================================================

// identifyThreats analyzes the game state for threats to the player
func (ba *BasicAIAdvisor) identifyThreats(game *weewar.Game, playerID int) ([]Threat, error) {
	threats := make([]Threat, 0)

	playerUnits := game.GetUnitsForPlayer(playerID)

	// Check each player unit for potential threats
	for _, unit := range playerUnits {
		// Find enemy units that can attack this unit
		enemyThreats := ba.findEnemyThreats(game, unit, playerID)
		threats = append(threats, enemyThreats...)
	}

	return threats, nil
}

// findEnemyThreats finds enemy units that can threaten the given unit
func (ba *BasicAIAdvisor) findEnemyThreats(game *weewar.Game, targetUnit *weewar.Unit, playerID int) []Threat {
	threats := make([]Threat, 0)

	// Check all enemy players
	for pid := 0; pid < game.PlayerCount(); pid++ {
		if pid == playerID {
			continue // Skip own units
		}

		enemyUnits := game.GetUnitsForPlayer(pid)
		for _, enemyUnit := range enemyUnits {
			// Check if this enemy unit can attack the target unit
			// TODO: Use game.CanAttackUnit(enemyUnit, targetUnit)
			if ba.canUnitAttackTarget(game, enemyUnit, targetUnit) {
				threat := Threat{
					Position:    enemyUnit.Coord,
					ThreatLevel: ba.calculateThreatLevel(enemyUnit, targetUnit),
					ThreatType:  ThreatDirectAttack,
					TargetUnit:  targetUnit,
					ThreatUnit:  enemyUnit,
					Description: fmt.Sprintf("%s threatens %s", ba.getUnitName(enemyUnit), ba.getUnitName(targetUnit)),
					Urgency:     1, // Can attack this turn
					Solutions:   ba.generateThreatSolutions(game, targetUnit, enemyUnit),
				}
				threats = append(threats, threat)
			}
		}
	}

	return threats
}

// identifyOpportunities analyzes the game state for opportunities to exploit
func (ba *BasicAIAdvisor) identifyOpportunities(game *weewar.Game, playerID int) ([]Opportunity, error) {
	opportunities := make([]Opportunity, 0)

	playerUnits := game.GetUnitsForPlayer(playerID)

	// Check each player unit for opportunities
	for _, unit := range playerUnits {
		// Find weak enemy units this unit can attack
		attackOpportunities := ba.findAttackOpportunities(game, unit, playerID)
		opportunities = append(opportunities, attackOpportunities...)

		// Find undefended bases/cities this unit can capture
		captureOpportunities := ba.findCaptureOpportunities(game, unit, playerID)
		opportunities = append(opportunities, captureOpportunities...)
	}

	return opportunities, nil
}

// findAttackOpportunities finds weak enemy units that can be attacked
func (ba *BasicAIAdvisor) findAttackOpportunities(game *weewar.Game, attackerUnit *weewar.Unit, playerID int) []Opportunity {
	opportunities := make([]Opportunity, 0)

	// Check all enemy players
	for pid := 0; pid < game.PlayerCount(); pid++ {
		if pid == playerID {
			continue
		}

		enemyUnits := game.GetUnitsForPlayer(pid)
		for _, enemyUnit := range enemyUnits {
			// Check if we can attack this enemy unit
			if ba.canUnitAttackTarget(game, attackerUnit, enemyUnit) {
				opportunityValue := ba.calculateAttackOpportunityValue(attackerUnit, enemyUnit)

				if opportunityValue > 0.3 { // Only consider good opportunities
					opportunity := Opportunity{
						Position:        enemyUnit.Coord,
						OpportunityType: OpportunityWeakUnit,
						Value:           opportunityValue,
						RequiredUnit:    attackerUnit,
						TargetUnit:      enemyUnit,
						Description:     fmt.Sprintf("Attack weak %s with %s", ba.getUnitName(enemyUnit), ba.getUnitName(attackerUnit)),
						Difficulty:      0.2, // Direct attacks are usually easy
						TimeWindow:      1,   // Available this turn
						Requirements:    []string{"Unit in attack range"},
					}
					opportunities = append(opportunities, opportunity)
				}
			}
		}
	}

	return opportunities
}

// findCaptureOpportunities finds bases or cities that can be captured
func (ba *BasicAIAdvisor) findCaptureOpportunities(game *weewar.Game, unit *weewar.Unit, playerID int) []Opportunity {
	opportunities := make([]Opportunity, 0)

	// TODO: Implement base/city capture opportunity detection
	// This would check for undefended bases/cities within movement range

	return opportunities
}

// =============================================================================
// Helper Methods
// =============================================================================

// canUnitAttackTarget checks if one unit can attack another
func (ba *BasicAIAdvisor) canUnitAttackTarget(game *weewar.Game, attacker, target *weewar.Unit) bool {
	// TODO: Implement proper attack range checking using game methods
	// For now, use simple distance check
	distance := ba.calculateDistance(attacker.Coord, target.Coord)
	return distance <= 1.0 // Assume range of 1 for most units
}

// calculateDistance returns the distance between two positions
func (ba *BasicAIAdvisor) calculateDistance(pos1, pos2 weewar.AxialCoord) float64 {
	// Hexagonal distance calculation
	dx := float64(pos1.Q - pos2.Q)
	dy := float64(pos1.R - pos2.R)
	dz := float64(-pos1.Q - pos1.R + pos2.Q + pos2.R)
	return (math.Abs(dx) + math.Abs(dy) + math.Abs(dz)) / 2.0
}

// calculateThreatLevel assesses how dangerous a threat is
func (ba *BasicAIAdvisor) calculateThreatLevel(threatUnit, targetUnit *weewar.Unit) float64 {
	// Simple threat calculation based on unit health
	threatHealth := float64(threatUnit.AvailableHealth) / 100.0
	targetHealth := float64(targetUnit.AvailableHealth) / 100.0

	// Higher threat level if attacker is healthy and target is weak
	return threatHealth * (1.0 - targetHealth*0.5)
}

// calculateAttackOpportunityValue assesses how good an attack opportunity is
func (ba *BasicAIAdvisor) calculateAttackOpportunityValue(attacker, target *weewar.Unit) float64 {
	attackerHealth := float64(attacker.AvailableHealth) / 100.0
	targetHealth := float64(target.AvailableHealth) / 100.0

	// Good opportunity if attacker is healthy and target is weak
	return attackerHealth * (1.0 - targetHealth)
}

// getUnitName returns a human-readable unit name
func (ba *BasicAIAdvisor) getUnitName(unit *weewar.Unit) string {
	// TODO: Get actual unit name from unit data
	return fmt.Sprintf("Unit%d", unit.UnitType)
}

// generateThreatSolutions suggests ways to deal with a threat
func (ba *BasicAIAdvisor) generateThreatSolutions(game *weewar.Game, target, threat *weewar.Unit) []string {
	solutions := make([]string, 0)

	solutions = append(solutions, "Move unit to safety")
	solutions = append(solutions, "Counter-attack the threatening unit")
	solutions = append(solutions, "Reinforce with nearby units")

	return solutions
}

// generateStrategicReasoning provides overall strategic context
func (ba *BasicAIAdvisor) generateStrategicReasoning(game *weewar.Game, playerID int, suggestions *MoveSuggestions) string {
	if suggestions.PrimaryMove == nil {
		return "No clear strategic direction identified"
	}

	switch suggestions.PrimaryMove.Category {
	case CategoryOffensive:
		return "Focusing on offensive operations to gain material advantage"
	case CategoryDefensive:
		return "Prioritizing defensive positioning to protect valuable units"
	case CategoryEconomic:
		return "Emphasizing economic growth through base/city control"
	case CategoryPositional:
		return "Improving unit positioning for future tactical opportunities"
	case CategoryTactical:
		return "Executing tactical maneuvers to exploit immediate opportunities"
	default:
		return "Adapting strategy based on current position assessment"
	}
}
