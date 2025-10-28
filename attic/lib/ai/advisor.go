package ai

import (
	"time"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
)

// =============================================================================
// Core AI Advisor Interface
// =============================================================================

// AIAdvisor provides stateless AI decision-making capabilities for any game state
type AIAdvisor interface {
	// Core AI functionality
	SuggestMoves(game *weewar.Game, playerID int, options *AIOptions) (*MoveSuggestions, error)
	EvaluatePosition(game *weewar.Game, playerID int) (*PositionEvaluation, error)

	// Analysis functions
	GetThreats(game *weewar.Game, playerID int) ([]Threat, error)
	GetOpportunities(game *weewar.Game, playerID int) ([]Opportunity, error)
	GetStrategicValue(game *weewar.Game, position weewar.Position) float64
}

// =============================================================================
// Configuration Types
// =============================================================================

// AIOptions configures AI behavior for a specific suggestion request
type AIOptions struct {
	Difficulty    AIDifficulty  `json:"difficulty"`    // AI skill level
	Personality   AIPersonality `json:"personality"`   // AI playing style
	MaxMoves      int           `json:"maxMoves"`      // Max moves to suggest (default: 1)
	ThinkingTime  time.Duration `json:"thinkingTime"`  // Max time to spend (default: 1s)
	ShowReasoning bool          `json:"showReasoning"` // Include detailed reasoning
}

// AIDifficulty represents AI skill levels
type AIDifficulty int

const (
	AIEasy AIDifficulty = iota
	AIMedium
	AIHard
	AIExpert
)

func (d AIDifficulty) String() string {
	switch d {
	case AIEasy:
		return "easy"
	case AIMedium:
		return "medium"
	case AIHard:
		return "hard"
	case AIExpert:
		return "expert"
	default:
		return "unknown"
	}
}

// AIPersonality represents different AI playing styles
type AIPersonality int

const (
	AIAggressive AIPersonality = iota
	AIDefensive
	AIBalanced
	AIExpansionist
)

func (p AIPersonality) String() string {
	switch p {
	case AIAggressive:
		return "aggressive"
	case AIDefensive:
		return "defensive"
	case AIBalanced:
		return "balanced"
	case AIExpansionist:
		return "expansionist"
	default:
		return "unknown"
	}
}

// =============================================================================
// Move Suggestion Types
// =============================================================================

// MoveSuggestions contains AI move recommendations
type MoveSuggestions struct {
	PrimaryMove      *MoveProposal   `json:"primaryMove"`      // Best recommended move
	AlternativeMoves []*MoveProposal `json:"alternativeMoves"` // Other good options
	Reasoning        string          `json:"reasoning"`        // Overall strategic reasoning
	Confidence       float64         `json:"confidence"`       // AI confidence (0-1)
	ThinkingTime     time.Duration   `json:"thinkingTime"`     // Time spent calculating
}

// MoveProposal represents a specific move recommendation
type MoveProposal struct {
	// Move identification
	Action ActionType      `json:"action"` // Type of action
	UnitID int             `json:"unitID"` // Unit to act with
	From   weewar.Position `json:"from"`   // Current position
	To     weewar.Position `json:"to"`     // Target position

	// Move evaluation
	Priority float64 `json:"priority"` // Move quality score (0-1)
	Risk     float64 `json:"risk"`     // Risk assessment (0-1)
	Value    float64 `json:"value"`    // Expected value gain

	// Reasoning
	Reason   string       `json:"reason"`   // Human-readable explanation
	Category MoveCategory `json:"category"` // Strategic category
}

// ActionType represents the type of action being proposed
type ActionType int

const (
	ActionMove ActionType = iota
	ActionAttack
	ActionCreateUnit
	ActionCapture
	ActionRepair
	ActionEndTurn
)

func (a ActionType) String() string {
	switch a {
	case ActionMove:
		return "move"
	case ActionAttack:
		return "attack"
	case ActionCreateUnit:
		return "create_unit"
	case ActionCapture:
		return "capture"
	case ActionRepair:
		return "repair"
	case ActionEndTurn:
		return "end_turn"
	default:
		return "unknown"
	}
}

// MoveCategory represents the strategic purpose of a move
type MoveCategory int

const (
	CategoryOffensive MoveCategory = iota
	CategoryDefensive
	CategoryEconomic
	CategoryPositional
	CategoryTactical
)

func (c MoveCategory) String() string {
	switch c {
	case CategoryOffensive:
		return "offensive"
	case CategoryDefensive:
		return "defensive"
	case CategoryEconomic:
		return "economic"
	case CategoryPositional:
		return "positional"
	case CategoryTactical:
		return "tactical"
	default:
		return "unknown"
	}
}

// =============================================================================
// Position Evaluation Types
// =============================================================================

// PositionEvaluation contains comprehensive position analysis
type PositionEvaluation struct {
	// Overall scores
	OverallScore float64 `json:"overallScore"` // Total position score (-1 to 1)
	Confidence   float64 `json:"confidence"`   // Evaluation confidence (0-1)

	// Component scores
	MaterialScore  float64 `json:"materialScore"`  // Unit values and health
	EconomicScore  float64 `json:"economicScore"`  // Base/city control
	TacticalScore  float64 `json:"tacticalScore"`  // Unit positioning
	StrategicScore float64 `json:"strategicScore"` // Long-term advantages

	// Detailed breakdown
	ComponentScores map[string]float64 `json:"componentScores"` // Individual metric scores

	// Analysis
	Strengths  []string `json:"strengths"`  // Position strengths
	Weaknesses []string `json:"weaknesses"` // Position weaknesses
	KeyFactors []string `json:"keyFactors"` // Most important factors
}

// =============================================================================
// Threat and Opportunity Types
// =============================================================================

// Threat represents a danger to the player
type Threat struct {
	// Threat identification
	Position    weewar.Position `json:"position"`    // Position of threat
	ThreatLevel float64         `json:"threatLevel"` // Severity (0-1)
	ThreatType  ThreatType      `json:"threatType"`  // Type of threat

	// Units involved
	TargetUnit *v1.Unit `json:"targetUnit"` // Unit being threatened
	ThreatUnit *v1.Unit `json:"threatUnit"` // Unit posing threat

	// Analysis
	Description string   `json:"description"` // Human-readable description
	Urgency     int      `json:"urgency"`     // Turns until threat materializes
	Solutions   []string `json:"solutions"`   // Possible responses
}

// ThreatType categorizes different types of threats
type ThreatType int

const (
	ThreatDirectAttack ThreatType = iota
	ThreatFlanking
	ThreatBaseCapture
	ThreatEncirclement
	ThreatEconomic
)

func (t ThreatType) String() string {
	switch t {
	case ThreatDirectAttack:
		return "direct_attack"
	case ThreatFlanking:
		return "flanking"
	case ThreatBaseCapture:
		return "base_capture"
	case ThreatEncirclement:
		return "encirclement"
	case ThreatEconomic:
		return "economic"
	default:
		return "unknown"
	}
}

// Opportunity represents an advantage the player can exploit
type Opportunity struct {
	// Opportunity identification
	Position        weewar.Position `json:"position"`        // Position of opportunity
	OpportunityType OpportunityType `json:"opportunityType"` // Type of opportunity
	Value           float64         `json:"value"`           // Strategic value (0-1)

	// Units involved
	RequiredUnit *v1.Unit `json:"requiredUnit"` // Unit that can exploit opportunity
	TargetUnit   *v1.Unit `json:"targetUnit"`   // Unit that can be targeted (if applicable)

	// Analysis
	Description  string   `json:"description"`  // Human-readable description
	Difficulty   float64  `json:"difficulty"`   // Execution difficulty (0-1)
	TimeWindow   int      `json:"timeWindow"`   // Turns before opportunity expires
	Requirements []string `json:"requirements"` // What's needed to execute
}

// OpportunityType categorizes different types of opportunities
type OpportunityType int

const (
	OpportunityWeakUnit OpportunityType = iota
	OpportunityUndefendedBase
	OpportunityFlanking
	OpportunityTerritoryGain
	OpportunityEconomic
	OpportunityTactical
)

func (o OpportunityType) String() string {
	switch o {
	case OpportunityWeakUnit:
		return "weak_unit"
	case OpportunityUndefendedBase:
		return "undefended_base"
	case OpportunityFlanking:
		return "flanking"
	case OpportunityTerritoryGain:
		return "territory_gain"
	case OpportunityEconomic:
		return "economic"
	case OpportunityTactical:
		return "tactical"
	default:
		return "unknown"
	}
}

// =============================================================================
// Utility Functions
// =============================================================================

// NewAIOptions creates default AI options
func NewAIOptions() *AIOptions {
	return &AIOptions{
		Difficulty:    AIMedium,
		Personality:   AIBalanced,
		MaxMoves:      1,
		ThinkingTime:  time.Second,
		ShowReasoning: false,
	}
}

// WithDifficulty sets the AI difficulty level
func (opts *AIOptions) WithDifficulty(difficulty AIDifficulty) *AIOptions {
	opts.Difficulty = difficulty
	return opts
}

// WithPersonality sets the AI personality
func (opts *AIOptions) WithPersonality(personality AIPersonality) *AIOptions {
	opts.Personality = personality
	return opts
}

// WithMaxMoves sets the maximum number of moves to suggest
func (opts *AIOptions) WithMaxMoves(maxMoves int) *AIOptions {
	opts.MaxMoves = maxMoves
	return opts
}

// WithThinkingTime sets the maximum thinking time
func (opts *AIOptions) WithThinkingTime(duration time.Duration) *AIOptions {
	opts.ThinkingTime = duration
	return opts
}

// WithReasoning enables detailed reasoning output
func (opts *AIOptions) WithReasoning() *AIOptions {
	opts.ShowReasoning = true
	return opts
}
