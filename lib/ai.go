package weewar

// =============================================================================
// WeeWar AI Interface Definitions
// =============================================================================
// This file defines AI-specific interfaces and types for the WeeWar game system.
// These interfaces focus on AI decision-making, strategic analysis, and automated
// gameplay capabilities.

// =============================================================================
// AI Data Types
// =============================================================================

// AIAction represents an action an AI player can take
type AIAction struct {
	Type     string   `json:"type"`     // "move", "attack", "create_unit", "end_turn"
	UnitID   int      `json:"unitId"`   // ID of unit performing action
	From     Position `json:"from"`     // Starting position
	To       Position `json:"to"`       // Target position
	UnitType int      `json:"unitType"` // For create_unit actions
	Priority float64  `json:"priority"` // Action priority (0-1)
	Reason   string   `json:"reason"`   // Human-readable reason for action
}

// Threat represents a danger to a player
type Threat struct {
	Position     Position `json:"position"`     // Position of threat
	ThreatLevel  float64  `json:"threatLevel"`  // Severity (0-1)
	ThreatType   string   `json:"threatType"`   // Type of threat
	TargetUnit   *Unit    `json:"targetUnit"`   // Unit being threatened
	ThreatUnit   *Unit    `json:"threatUnit"`   // Unit posing threat
	Description  string   `json:"description"`  // Human-readable description
}

// Opportunity represents an advantage a player can take
type Opportunity struct {
	Position        Position `json:"position"`        // Position of opportunity
	OpportunityType string   `json:"opportunityType"` // Type of opportunity
	Value           float64  `json:"value"`           // Strategic value (0-1)
	RequiredUnit    *Unit    `json:"requiredUnit"`    // Unit that can exploit opportunity
	TargetUnit      *Unit    `json:"targetUnit"`      // Unit that can be targeted
	Description     string   `json:"description"`     // Human-readable description
}

// AIDifficulty represents AI difficulty levels
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
// AI Interface
// =============================================================================

// AIInterface provides AI decision-making capabilities
type AIInterface interface {
	// AI Decision Making
	// GetAIMove returns AI player's chosen action
	// Called by: AI turn processing, Auto-play systems, AI vs AI games
	// Returns: AI action (move, attack, end turn)
	GetAIMove(playerID int, difficulty AIDifficulty) (*AIAction, error)
	
	// EvaluatePosition returns strategic value of current position
	// Called by: AI difficulty adjustment, Game balance analysis, Statistics
	// Returns: Position evaluation score (higher = better for player)
	EvaluatePosition(playerID int) float64
	
	// GetBestMove returns optimal move for player
	// Called by: AI hint system, Tutorial mode, Move suggestion
	// Returns: Best available action for player
	GetBestMove(playerID int) (*AIAction, error)
	
	// AI Analysis
	// GetThreats returns immediate threats to player
	// Called by: AI defensive planning, Warning systems, Tutorial hints
	// Returns: Array of threat objects
	GetThreats(playerID int) []Threat
	
	// GetOpportunities returns attack/advancement opportunities
	// Called by: AI offensive planning, Hint systems, Strategy analysis
	// Returns: Array of opportunity objects
	GetOpportunities(playerID int) []Opportunity
	
	// GetStrategicValue returns value of specific position
	// Called by: AI movement planning, Territory evaluation, Base placement
	// Returns: Strategic value score
	GetStrategicValue(position Position) float64
	
	// AI Configuration
	// SetAIPersonality changes AI playing style
	// Called by: Game setup, Player preferences, Dynamic difficulty
	SetAIPersonality(playerID int, personality AIPersonality)
	
	// GetAIPersonality returns current AI playing style
	// Called by: UI display, Game statistics, AI debugging
	GetAIPersonality(playerID int) AIPersonality
	
	// SetAIThinkingTime sets maximum time AI can spend on decisions
	// Called by: Game settings, Performance tuning, Real-time constraints
	SetAIThinkingTime(playerID int, maxSeconds float64)
	
	// GetAIThinkingTime returns current AI thinking time limit
	// Called by: Performance monitoring, UI display
	GetAIThinkingTime(playerID int) float64
}

// =============================================================================
// AI Player Interface
// =============================================================================

// AIPlayer represents an AI player instance
type AIPlayer interface {
	// Player Information
	GetPlayerID() int
	GetDifficulty() AIDifficulty
	GetPersonality() AIPersonality
	
	// AI Decision Making
	ChooseAction(game GameInterface) (*AIAction, error)
	EvaluateGameState(game GameInterface) float64
	
	// AI Learning (for future implementation)
	UpdateFromGameResult(won bool, finalScore float64)
	SaveLearningData() ([]byte, error)
	LoadLearningData(data []byte) error
}

// =============================================================================
// AI Utility Functions
// =============================================================================

// AIActionType constants for action types
const (
	AIActionMove      = "move"
	AIActionAttack    = "attack"
	AIActionCreateUnit = "create_unit"
	AIActionEndTurn   = "end_turn"
	AIActionCapture   = "capture"
	AIActionRepair    = "repair"
)

// ThreatType constants for threat types
const (
	ThreatDirectAttack = "direct_attack"
	ThreatFlanking     = "flanking"
	ThreatBaseCapture  = "base_capture"
	ThreatEncirclement = "encirclement"
)

// OpportunityType constants for opportunity types
const (
	OpportunityWeakUnit     = "weak_unit"
	OpportunityUndefendedBase = "undefended_base"
	OpportunityFlanking     = "flanking"
	OpportunityTerritoryGain = "territory_gain"
)