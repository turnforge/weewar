package ai

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
)

// =============================================================================
// Decision Strategy Interface
// =============================================================================

// DecisionStrategy defines how AI makes decisions at different difficulty levels
type DecisionStrategy interface {
	SuggestMoves(game *weewar.Game, playerID int, options *AIOptions) (*MoveSuggestions, error)
	GetStrategyName() string
	GetComplexity() int // Relative computational complexity (1-10)
}

// =============================================================================
// Easy Strategy - Random + Basic Threat Avoidance
// =============================================================================

// EasyStrategy implements random decision making with basic threat avoidance
type EasyStrategy struct {
	rng       *rand.Rand
	evaluator *PositionEvaluator
}

// NewEasyStrategy creates a new easy AI strategy
func NewEasyStrategy(rng *rand.Rand, evaluator *PositionEvaluator) *EasyStrategy {
	return &EasyStrategy{
		rng:       rng,
		evaluator: evaluator,
	}
}

func (es *EasyStrategy) GetStrategyName() string {
	return "Random with Basic Avoidance"
}

func (es *EasyStrategy) GetComplexity() int {
	return 1
}

func (es *EasyStrategy) SuggestMoves(game *weewar.Game, playerID int, options *AIOptions) (*MoveSuggestions, error) {
	// Generate all valid moves
	allMoves, err := es.generateAllMoves(game, playerID)
	if err != nil {
		return nil, err
	}

	// Filter out obviously dangerous moves
	safeMoves := es.filterDangerousMoves(game, allMoves, playerID)

	// Randomly select from safe moves
	var selectedMove *MoveProposal
	if len(safeMoves) > 0 {
		selectedMove = safeMoves[es.rng.Intn(len(safeMoves))]
		selectedMove.Priority = 0.3 + es.rng.Float64()*0.4 // 0.3-0.7 range
	} else if len(allMoves) > 0 {
		// Fallback to any valid move if no safe ones
		selectedMove = allMoves[es.rng.Intn(len(allMoves))]
		selectedMove.Priority = 0.1 + es.rng.Float64()*0.3 // 0.1-0.4 range
	} else {
		// End turn if no moves available
		selectedMove = &MoveProposal{
			Action:   ActionEndTurn,
			Priority: 1.0,
			Reason:   "No other moves available",
			Category: CategoryPositional,
		}
	}

	return &MoveSuggestions{
		PrimaryMove:      selectedMove,
		AlternativeMoves: es.selectAlternatives(safeMoves, selectedMove, 2),
		Reasoning:        "Random selection with basic threat avoidance",
		Confidence:       0.3 + es.rng.Float64()*0.2, // Low confidence for random play
	}, nil
}

func (es *EasyStrategy) generateAllMoves(game *weewar.Game, playerID int) ([]*MoveProposal, error) {
	moves := make([]*MoveProposal, 0)

	playerUnits := game.GetUnitsForPlayer(int(playerID))
	for _, unit := range playerUnits {
		// Generate basic movement and attack moves
		unitMoves := es.generateUnitMoves(game, unit)
		moves = append(moves, unitMoves...)
	}

	// Add end turn option
	moves = append(moves, &MoveProposal{
		Action:   ActionEndTurn,
		Priority: 0.2,
		Reason:   "End turn",
		Category: CategoryPositional,
	})

	return moves, nil
}

func (es *EasyStrategy) generateUnitMoves(game *weewar.Game, unit *v1.Unit) []*MoveProposal {
	moves := make([]*MoveProposal, 0)

	// TODO: Integrate with actual game movement/attack systems
	// For now, generate placeholder moves

	return moves
}

func (es *EasyStrategy) filterDangerousMoves(game *weewar.Game, moves []*MoveProposal, playerID int) []*MoveProposal {
	safe := make([]*MoveProposal, 0)

	for _, move := range moves {
		if move.Action == ActionEndTurn {
			safe = append(safe, move)
			continue
		}

		// Simple safety check: don't move valuable units into obvious danger
		if !es.isMoveDangerous(game, move, playerID) {
			safe = append(safe, move)
		}
	}

	return safe
}

func (es *EasyStrategy) isMoveDangerous(game *weewar.Game, move *MoveProposal, playerID int) bool {
	// Simplified danger assessment
	// TODO: Implement proper threat analysis
	return false
}

func (es *EasyStrategy) selectAlternatives(moves []*MoveProposal, selected *MoveProposal, count int) []*MoveProposal {
	alternatives := make([]*MoveProposal, 0, count)

	for _, move := range moves {
		if len(alternatives) >= count {
			break
		}
		if move != selected {
			alternatives = append(alternatives, move)
		}
	}

	return alternatives
}

// =============================================================================
// Medium Strategy - Greedy + Combat Prediction
// =============================================================================

// MediumStrategy implements tactical decision making with combat predictions
type MediumStrategy struct {
	evaluator   *PositionEvaluator
	rulesEngine *weewar.RulesEngine
}

// NewMediumStrategy creates a new medium AI strategy
func NewMediumStrategy(evaluator *PositionEvaluator, rulesEngine *weewar.RulesEngine) *MediumStrategy {
	return &MediumStrategy{
		evaluator:   evaluator,
		rulesEngine: rulesEngine,
	}
}

func (ms *MediumStrategy) GetStrategyName() string {
	return "Greedy with Combat Prediction"
}

func (ms *MediumStrategy) GetComplexity() int {
	return 3
}

func (ms *MediumStrategy) SuggestMoves(game *weewar.Game, playerID int, options *AIOptions) (*MoveSuggestions, error) {
	// Analyze current position
	threats := ms.analyzeThreats(game, playerID)
	opportunities := ms.analyzeOpportunities(game, playerID)

	// Priority-based decision making
	var selectedMove *MoveProposal
	var reasoning string

	// Priority 1: Handle immediate high threats
	if highThreat := ms.findHighThreat(threats); highThreat != nil {
		selectedMove = ms.generateThreatResponse(game, highThreat, playerID)
		reasoning = "Responding to immediate threat"
	}

	// Priority 2: Exploit high-value opportunities
	if selectedMove == nil {
		if opportunity := ms.findBestOpportunity(opportunities); opportunity != nil {
			selectedMove = ms.generateOpportunityMove(game, opportunity, playerID)
			reasoning = "Exploiting tactical opportunity"
		}
	}

	// Priority 3: Improve position tactically
	if selectedMove == nil {
		selectedMove = ms.generatePositionalMove(game, playerID)
		reasoning = "Improving tactical position"
	}

	// Fallback to end turn
	if selectedMove == nil {
		selectedMove = &MoveProposal{
			Action:   ActionEndTurn,
			Priority: 1.0,
			Reason:   "No beneficial moves found",
			Category: CategoryPositional,
		}
		reasoning = "Consolidating position"
	}

	// Generate alternatives
	alternatives := ms.generateAlternatives(game, playerID, selectedMove)

	return &MoveSuggestions{
		PrimaryMove:      selectedMove,
		AlternativeMoves: alternatives,
		Reasoning:        reasoning,
		Confidence:       0.6 + math.Min(selectedMove.Priority*0.3, 0.3),
	}, nil
}

func (ms *MediumStrategy) analyzeThreats(game *weewar.Game, playerID int) []Threat {
	// TODO: Implement threat analysis
	return make([]Threat, 0)
}

func (ms *MediumStrategy) analyzeOpportunities(game *weewar.Game, playerID int) []Opportunity {
	// TODO: Implement opportunity analysis
	return make([]Opportunity, 0)
}

func (ms *MediumStrategy) findHighThreat(threats []Threat) *Threat {
	for _, threat := range threats {
		if threat.ThreatLevel > 0.7 && threat.Urgency <= 1 {
			return &threat
		}
	}
	return nil
}

func (ms *MediumStrategy) findBestOpportunity(opportunities []Opportunity) *Opportunity {
	var best *Opportunity
	for _, opp := range opportunities {
		if opp.Value > 0.6 && (best == nil || opp.Value > best.Value) {
			oppCopy := opp
			best = &oppCopy
		}
	}
	return best
}

func (ms *MediumStrategy) generateThreatResponse(game *weewar.Game, threat *Threat, playerID int) *MoveProposal {
	// TODO: Generate appropriate response to threat
	return &MoveProposal{
		Action:   ActionMove,
		Priority: 0.8,
		Reason:   "Responding to threat",
		Category: CategoryDefensive,
	}
}

func (ms *MediumStrategy) generateOpportunityMove(game *weewar.Game, opportunity *Opportunity, playerID int) *MoveProposal {
	// TODO: Generate move to exploit opportunity
	return &MoveProposal{
		Action:   ActionAttack,
		Priority: 0.7,
		Reason:   "Exploiting opportunity",
		Category: CategoryOffensive,
	}
}

func (ms *MediumStrategy) generatePositionalMove(game *weewar.Game, playerID int) *MoveProposal {
	// TODO: Generate positional improvement move
	return &MoveProposal{
		Action:   ActionMove,
		Priority: 0.5,
		Reason:   "Improving position",
		Category: CategoryPositional,
	}
}

func (ms *MediumStrategy) generateAlternatives(game *weewar.Game, playerID int, selected *MoveProposal) []*MoveProposal {
	// TODO: Generate alternative moves
	return make([]*MoveProposal, 0)
}

// =============================================================================
// Hard Strategy - Multi-turn Planning + Coordination
// =============================================================================

// HardStrategy implements multi-turn planning with unit coordination
type HardStrategy struct {
	evaluator   *PositionEvaluator
	rulesEngine *weewar.RulesEngine
	planDepth   int
}

// NewHardStrategy creates a new hard AI strategy
func NewHardStrategy(evaluator *PositionEvaluator, rulesEngine *weewar.RulesEngine) *HardStrategy {
	return &HardStrategy{
		evaluator:   evaluator,
		rulesEngine: rulesEngine,
		planDepth:   2, // Look ahead 2 turns
	}
}

func (hs *HardStrategy) GetStrategyName() string {
	return "Multi-turn Planning with Coordination"
}

func (hs *HardStrategy) GetComplexity() int {
	return 6
}

func (hs *HardStrategy) SuggestMoves(game *weewar.Game, playerID int, options *AIOptions) (*MoveSuggestions, error) {
	// Multi-turn analysis
	// currentEval := hs.evaluator.EvaluatePosition(game, playerID)

	// Generate candidate moves with lookahead
	candidateMoves := hs.generateCandidateMoves(game, playerID)

	// Evaluate each move with planning horizon
	bestMove := hs.evaluateMovesWithLookahead(game, playerID, candidateMoves)

	if bestMove == nil {
		bestMove = &MoveProposal{
			Action:   ActionEndTurn,
			Priority: 1.0,
			Reason:   "No beneficial moves in planning horizon",
			Category: CategoryPositional,
		}
	}

	alternatives := hs.selectBestAlternatives(candidateMoves, bestMove, 3)

	return &MoveSuggestions{
		PrimaryMove:      bestMove,
		AlternativeMoves: alternatives,
		Reasoning:        fmt.Sprintf("Multi-turn analysis (depth %d) with coordination", hs.planDepth),
		Confidence:       0.7 + math.Min(bestMove.Priority*0.2, 0.2),
	}, nil
}

func (hs *HardStrategy) generateCandidateMoves(game *weewar.Game, playerID int) []*MoveProposal {
	// TODO: Generate candidate moves with strategic filtering
	return make([]*MoveProposal, 0)
}

func (hs *HardStrategy) evaluateMovesWithLookahead(game *weewar.Game, playerID int, moves []*MoveProposal) *MoveProposal {
	var bestMove *MoveProposal
	bestScore := -math.Inf(1)

	for _, move := range moves {
		// Simulate move and evaluate resulting position
		score := hs.evaluateMoveWithLookahead(game, playerID, move)
		if score > bestScore {
			bestScore = score
			bestMove = move
		}
	}

	return bestMove
}

func (hs *HardStrategy) evaluateMoveWithLookahead(game *weewar.Game, playerID int, move *MoveProposal) float64 {
	// TODO: Implement lookahead evaluation
	// This would simulate the move and recursively evaluate the resulting position
	return 0.0
}

func (hs *HardStrategy) selectBestAlternatives(moves []*MoveProposal, selected *MoveProposal, count int) []*MoveProposal {
	alternatives := make([]*MoveProposal, 0)

	// Sort moves by priority
	sort.Slice(moves, func(i, j int) bool {
		return moves[i].Priority > moves[j].Priority
	})

	for _, move := range moves {
		if len(alternatives) >= count {
			break
		}
		if move != selected {
			alternatives = append(alternatives, move)
		}
	}

	return alternatives
}

// =============================================================================
// Expert Strategy - Minimax + Advanced Optimization
// =============================================================================

// ExpertStrategy implements minimax algorithm with advanced optimizations
type ExpertStrategy struct {
	evaluator          *PositionEvaluator
	rulesEngine        *weewar.RulesEngine
	maxDepth           int
	transpositionTable map[string]float64
	moveOrdering       *MoveOrderer
}

// MoveOrderer helps optimize alpha-beta pruning through move ordering
type MoveOrderer struct {
	history map[string]int // Move history for ordering
}

// NewExpertStrategy creates a new expert AI strategy
func NewExpertStrategy(evaluator *PositionEvaluator, rulesEngine *weewar.RulesEngine) *ExpertStrategy {
	return &ExpertStrategy{
		evaluator:          evaluator,
		rulesEngine:        rulesEngine,
		maxDepth:           3, // 3-move lookahead
		transpositionTable: make(map[string]float64),
		moveOrdering:       &MoveOrderer{history: make(map[string]int)},
	}
}

func (es *ExpertStrategy) GetStrategyName() string {
	return "Minimax with Alpha-Beta Pruning"
}

func (es *ExpertStrategy) GetComplexity() int {
	return 10
}

func (es *ExpertStrategy) SuggestMoves(game *weewar.Game, playerID int, options *AIOptions) (*MoveSuggestions, error) {
	startTime := time.Now()

	// Adjust search depth based on thinking time
	searchDepth := es.calculateSearchDepth(options.ThinkingTime)

	// Clear transposition table periodically to prevent memory issues
	if len(es.transpositionTable) > 10000 {
		es.transpositionTable = make(map[string]float64)
	}

	// Generate root moves
	rootMoves := es.generateRootMoves(game, playerID)

	// Order moves for better alpha-beta efficiency
	es.moveOrdering.OrderMoves(rootMoves)

	// Minimax search with iterative deepening
	bestMove, bestScore := es.iterativeDeepening(game, playerID, rootMoves, searchDepth, options.ThinkingTime)

	if bestMove == nil {
		bestMove = &MoveProposal{
			Action:   ActionEndTurn,
			Priority: 1.0,
			Reason:   "No moves found in search",
			Category: CategoryPositional,
		}
	} else {
		bestMove.Priority = es.normalizeScore(bestScore)
	}

	// Select alternatives from search results
	alternatives := es.selectSearchAlternatives(rootMoves, bestMove, 3)

	searchTime := time.Since(startTime)
	reasoning := fmt.Sprintf("Minimax search (depth %d, %d nodes, %.2fs)",
		searchDepth, len(rootMoves), searchTime.Seconds())

	return &MoveSuggestions{
		PrimaryMove:      bestMove,
		AlternativeMoves: alternatives,
		Reasoning:        reasoning,
		Confidence:       0.8 + math.Min(bestMove.Priority*0.15, 0.15),
	}, nil
}

func (es *ExpertStrategy) calculateSearchDepth(thinkingTime time.Duration) int {
	// Adjust depth based on available thinking time
	seconds := thinkingTime.Seconds()
	if seconds < 1.0 {
		return 2
	} else if seconds < 5.0 {
		return 3
	} else if seconds < 15.0 {
		return 4
	} else {
		return 5
	}
}

func (es *ExpertStrategy) generateRootMoves(game *weewar.Game, playerID int) []*MoveProposal {
	// TODO: Generate all legal moves for minimax search
	return make([]*MoveProposal, 0)
}

func (es *ExpertStrategy) iterativeDeepening(game *weewar.Game, playerID int, moves []*MoveProposal, maxDepth int, maxTime time.Duration) (*MoveProposal, float64) {
	var bestMove *MoveProposal
	bestScore := -math.Inf(1)
	startTime := time.Now()

	// Iterative deepening from depth 1 to maxDepth
	for depth := 1; depth <= maxDepth; depth++ {
		// Check if we're running out of time
		if time.Since(startTime) > maxTime*8/10 { // Use 80% of available time
			break
		}

		// Search at current depth
		move, score := es.minimaxRoot(game, playerID, moves, depth)

		if move != nil {
			bestMove = move
			bestScore = score
		}
	}

	return bestMove, bestScore
}

func (es *ExpertStrategy) minimaxRoot(game *weewar.Game, playerID int, moves []*MoveProposal, depth int) (*MoveProposal, float64) {
	var bestMove *MoveProposal
	bestScore := -math.Inf(1)
	alpha := -math.Inf(1)
	beta := math.Inf(1)

	for _, move := range moves {
		// Simulate move (would need game state copying)
		// score := es.minimax(simulatedGame, depth-1, alpha, beta, false, playerID)
		score := 0.0 // Placeholder

		if score > bestScore {
			bestScore = score
			bestMove = move
		}

		alpha = math.Max(alpha, score)
		if beta <= alpha {
			break // Alpha-beta cutoff
		}
	}

	return bestMove, bestScore
}

func (es *ExpertStrategy) minimax(game *weewar.Game, depth int, alpha, beta float64, maximizing bool, playerID int32) float64 {
	// Base case: depth reached or game over
	if depth == 0 || game.Status != weewar.GameStatusPlaying {
		evaluation := es.evaluator.EvaluatePosition(game, playerID)
		return evaluation.OverallScore
	}

	// Check transposition table
	gameState := es.hashGameState(game)
	if cached, exists := es.transpositionTable[gameState]; exists {
		return cached
	}

	var score float64

	if maximizing {
		score = -math.Inf(1)
		// Generate moves for current player
		// For each move, simulate and recurse
		// TODO: Implement move generation and simulation

	} else {
		score = math.Inf(1)
		// Generate moves for opponent
		// For each move, simulate and recurse
		// TODO: Implement opponent move generation
	}

	// Store in transposition table
	es.transpositionTable[gameState] = score

	return score
}

func (es *ExpertStrategy) hashGameState(game *weewar.Game) string {
	// TODO: Implement efficient game state hashing
	return fmt.Sprintf("turn_%d_player_%d", game.TurnCounter, game.CurrentPlayer)
}

func (es *ExpertStrategy) normalizeScore(score float64) float64 {
	// Normalize minimax score to 0-1 range for priority
	return math.Max(0.0, math.Min(1.0, (score+1.0)/2.0))
}

func (es *ExpertStrategy) selectSearchAlternatives(moves []*MoveProposal, selected *MoveProposal, count int) []*MoveProposal {
	alternatives := make([]*MoveProposal, 0)

	// Sort by priority and select top alternatives
	sort.Slice(moves, func(i, j int) bool {
		return moves[i].Priority > moves[j].Priority
	})

	for _, move := range moves {
		if len(alternatives) >= count {
			break
		}
		if move != selected {
			alternatives = append(alternatives, move)
		}
	}

	return alternatives
}

// OrderMoves sorts moves for better alpha-beta pruning efficiency
func (mo *MoveOrderer) OrderMoves(moves []*MoveProposal) {
	sort.Slice(moves, func(i, j int) bool {
		// Order by: attacks first, then by historical success
		scoreI := mo.getMoveOrderingScore(moves[i])
		scoreJ := mo.getMoveOrderingScore(moves[j])
		return scoreI > scoreJ
	})
}

func (mo *MoveOrderer) getMoveOrderingScore(move *MoveProposal) float64 {
	score := 0.0

	// Attacks have higher priority
	if move.Action == ActionAttack {
		score += 10.0
	}

	// Add historical success bonus
	moveKey := fmt.Sprintf("%s_%d_%d", move.Action.String(), move.From.Q, move.From.R)
	if history, exists := mo.history[moveKey]; exists {
		score += float64(history) * 0.1
	}

	return score
}
