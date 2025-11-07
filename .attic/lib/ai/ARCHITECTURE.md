# WeeWar AI Toolkit Architecture

## Design Philosophy

The WeeWar AI system is designed as a **stateless helper library** that analyzes game states and proposes moves without modifying the core game engine. This approach provides maximum flexibility for different UI implementations while maintaining clean separation of concerns.

### Core Principles

1. **Stateless Design**: AI helpers are pure functions that take game state and return suggestions
2. **Non-Invasive Integration**: No modifications to existing Game, World, or Unit structures  
3. **Multiple AI Coexistence**: Different AI implementations can analyze the same game state
4. **Human Enhancement**: AI suggestions can assist human players without forcing automation
5. **Performance Optimized**: Caching and efficient algorithms for real-time gameplay

## Architecture Overview

```
┌─────────────────────┐    ┌──────────────────────┐    ┌─────────────────────┐
│                     │    │                      │    │                     │
│   UI Layer          │    │   AI Toolkit         │    │   Game Engine       │
│   (CLI/Web)         │    │                      │    │                     │
│                     │    │                      │    │                     │
├─────────────────────┤    ├──────────────────────┤    ├─────────────────────┤
│                     │    │                      │    │                     │
│ • Player Input      │◄──►│ • AIAdvisor          │◄──►│ • Game State        │
│ • Move Execution    │    │ • PositionEvaluator  │    │ • Move Validation    │
│ • Game Display      │    │ • Decision Strategies│    │ • Rules Engine      │
│                     │    │ • Move Generation    │    │ • Combat System     │
│                     │    │                      │    │                     │
└─────────────────────┘    └──────────────────────┘    └─────────────────────┘
```

## Core Components

### 1. AIAdvisor Interface

The central interface for all AI functionality:

```go
type AIAdvisor interface {
    SuggestMoves(game *Game, playerID int, options *AIOptions) (*MoveSuggestions, error)
    EvaluatePosition(game *Game, playerID int) *PositionEvaluation
    GetThreats(game *Game, playerID int) []Threat
    GetOpportunities(game *Game, playerID int) []Opportunity
}
```

**Design Rationale**:
- Single interface for all AI capabilities
- Game state passed as parameter (stateless)
- Options struct allows for runtime configuration
- Return structured suggestions rather than direct moves

### 2. Position Evaluation System

Multi-component evaluation framework:

```go
type PositionEvaluator struct {
    weights *EvaluationWeights
}

type EvaluationWeights struct {
    UnitValue        float64  // 0.25 - Raw unit strength
    UnitHealth       float64  // 0.15 - Unit condition  
    UnitPositioning  float64  // 0.15 - Tactical positions
    BaseControl      float64  // 0.20 - Production capacity
    IncomeControl    float64  // 0.15 - Economic advantage
    TerritoryControl float64  // 0.10 - Map control
    // ... additional factors
}
```

**Evaluation Components**:

1. **Material Evaluation** (40%)
   - Unit cost/value assessment using Advance Wars model
   - Health percentage weighting by unit value
   - Unit positioning on defensive terrain

2. **Economic Evaluation** (35%)
   - Base control for unit production capability
   - City control for income generation
   - Resource management efficiency

3. **Strategic Evaluation** (20%)
   - Threat analysis and defensive positioning
   - Attack opportunity identification
   - Territorial control assessment

4. **Positional Evaluation** (10%)
   - Unit mobility and movement options
   - Support network coordination
   - Strategic location control

### 3. Decision Strategy Hierarchy

Difficulty-based decision algorithms:

#### Easy AI (Random + Avoidance)
- **Algorithm**: Random selection from valid moves
- **Enhancement**: Filter moves that lose valuable units
- **Complexity**: O(n) where n = valid moves
- **Characteristics**: Unpredictable, makes obvious mistakes

#### Medium AI (Greedy + Prediction)
- **Algorithm**: Greedy selection based on immediate value
- **Enhancement**: Combat prediction for attack decisions
- **Complexity**: O(n * log n) for move sorting
- **Characteristics**: Tactical awareness, short-term planning

#### Hard AI (Multi-turn + Coordination)
- **Algorithm**: 2-3 move lookahead with unit coordination
- **Enhancement**: Strategic objective planning
- **Complexity**: O(b^d) where b=branching factor, d=depth
- **Characteristics**: Coordinated attacks, defensive formations

#### Expert AI (Minimax + Optimization)
- **Algorithm**: Minimax with alpha-beta pruning
- **Enhancement**: Transposition table, move ordering
- **Complexity**: O(b^d) optimized with pruning
- **Characteristics**: Near-optimal play, deep calculation

## Performance Considerations

### Caching Strategy

1. **Position Evaluation Cache**
   ```go
   type EvaluationCache struct {
       cache map[string]*PositionEvaluation
       ttl   time.Duration
   }
   ```

2. **Threat/Opportunity Cache**
   - Cache expensive spatial queries
   - Invalidate on game state changes
   - TTL-based expiration for memory management

3. **Move Generation Cache**
   - Cache valid moves for units that haven't moved
   - Invalidate on position changes

### Algorithm Optimizations

1. **Alpha-Beta Pruning**
   - Move ordering based on previous iteration results
   - Principal variation tracking
   - Iterative deepening for time management

2. **Transposition Tables**
   - Game state hashing for position repetition detection
   - Cached evaluation scores
   - Memory-bounded with replacement strategy

3. **Lazy Evaluation**
   - Calculate expensive metrics only when needed
   - Progressive refinement of move evaluations

## Integration with Existing Systems

### Game Engine Integration

The AI toolkit leverages existing game methods without modification:

```go
// Move generation uses existing queries
moveOptions, _ := game.GetUnitMovementOptions(unit)
attackOptions, _ := game.GetUnitAttackOptions(unit)

// Combat prediction uses existing rules engine
prediction, _ := game.GetRulesEngine().GetCombatPrediction(attackerID, defenderID)

// Move execution uses existing validation
err := game.MoveUnit(unit, targetPosition)
```

### Rules Engine Integration

- **Movement Costs**: Leverage existing terrain/unit movement matrices
- **Combat Calculations**: Use existing damage distribution system
- **Unit Data**: Access unit stats through existing data structures
- **Spatial Queries**: Utilize existing pathfinding and range calculations

## AI Personality System

Configurable AI behavior through weight adjustment:

### Aggressive Personality
```go
weights.AttackOptions = 0.25     // High offensive priority
weights.ThreatLevel = 0.05       // Low defensive concern
weights.UnitValue = 0.20         // Accept unit trades
```

### Defensive Personality  
```go
weights.ThreatLevel = 0.25       // High defensive priority
weights.AttackOptions = 0.05     // Conservative attacks
weights.BaseControl = 0.25       // Territory holding focus
```

### Economic Personality
```go
weights.IncomeControl = 0.25     // Economic expansion
weights.BaseControl = 0.25       // Production focus
weights.UnitValue = 0.15         // Resource conservation
```

## Extension Points

### Adding New AI Implementations

1. Implement `AIAdvisor` interface
2. Create custom `DecisionStrategy` 
3. Define personality-specific weights
4. Register with factory pattern

### Custom Evaluation Metrics

1. Extend `EvaluationWeights` struct
2. Implement evaluation method
3. Integrate into `PositionEvaluator`
4. Add to weight normalization

### Learning AI Framework (Future)

```go
type LearningAI interface {
    AIAdvisor
    Learn(gameHistory []Game, outcomes []GameResult) error
    SaveModel() ([]byte, error)
    LoadModel(data []byte) error
}
```

## Usage Examples

### AI vs AI Game
```go
aiAdvisor := NewBasicAIAdvisor()

for !game.IsGameOver() {
    playerID := game.GetCurrentPlayer()
    
    suggestions, _ := aiAdvisor.SuggestMoves(game, playerID, &AIOptions{
        Difficulty: AIHard,
        Personality: AIAggressive,
        MaxMoves: 1,
    })
    
    // Auto-execute for AI players
    move := suggestions.PrimaryMove
    game.ExecuteMove(move)
}
```

### Human Player with AI Assistance
```go
// Get AI suggestions
suggestions, _ := aiAdvisor.SuggestMoves(game, humanPlayerID, &AIOptions{
    Difficulty: AIExpert,
    MaxMoves: 3,
})

// Present to human player
fmt.Printf("AI suggests: %s (Score: %.2f)\n", 
    suggestions.PrimaryMove.Reason, 
    suggestions.PrimaryMove.Priority)

// Human chooses or overrides
humanChoice := getUserInput()
game.ExecuteMove(humanChoice)
```

### Multiple AI Analysis
```go
// Compare different AI approaches
basicAI := NewBasicAIAdvisor()
advancedAI := NewAdvancedAIAdvisor()

basic, _ := basicAI.SuggestMoves(game, playerID, &AIOptions{Difficulty: AIMedium})
advanced, _ := advancedAI.SuggestMoves(game, playerID, &AIOptions{Difficulty: AIExpert})

// Analyze differences for AI development
compareMoves(basic.PrimaryMove, advanced.PrimaryMove)
```

## Testing Strategy

### Unit Testing
- Position evaluation with known game states
- Move generation completeness and validity
- Decision strategy consistency across difficulty levels

### Integration Testing  
- AI vs AI games with different configurations
- Human vs AI games with move verification
- Performance testing with complex game states

### Regression Testing
- AI behavior consistency across game versions
- Performance regression detection
- Game balance verification

## Future Enhancements

### Planned Features
1. **Machine Learning Integration**: Neural networks for position evaluation
2. **Opening Book**: Pre-computed strong opening sequences
3. **Endgame Tables**: Perfect play databases for simplified positions
4. **Dynamic Difficulty**: AI that adapts to player skill level

### Research Areas
1. **Monte Carlo Tree Search**: Alternative to minimax for complex positions  
2. **Reinforcement Learning**: Self-improving AI through game experience
3. **Multi-Agent Cooperation**: Team-based AI coordination
4. **Real-time Constraints**: Anytime algorithms for time-pressured games

## Web Interface Integration

### Current Status (v10.2)
The AI toolkit is ready for web interface integration through the completed GameViewerPage architecture. The WASM bridge provides clean APIs for AI integration:

```typescript
// AI integration through GameState component
const aiSuggestions = await this.gameState.getAISuggestions(difficulty, personality);
const aiMove = await this.gameState.executeAIMove(playerID, difficulty);
```

### Planned Integration Features
- **AI Difficulty Selection**: Web UI dropdown for Easy/Medium/Hard/Expert
- **AI Personality Configuration**: Aggressive/Defensive/Balanced/Economic personality settings
- **AI Move Hints**: Show AI suggestions to human players with reasoning
- **AI vs AI Games**: Automated games with speed controls and visualization
- **AI Analysis Panel**: Position evaluation, threats, and opportunities display

## Current Integration Status

### WASM Bridge Ready ✅
- AI toolkit architecturally complete and tested via CLI
- Stateless design perfect for WASM integration
- Clean APIs ready for JavaScript binding

### Web Architecture Ready ✅ 
- GameViewerPage foundation with lifecycle controller complete
- GameState component with WASM bridge architecture ready
- Component communication via EventBus for AI coordination

### Next Steps
1. **WASM AI Bindings**: Add AI functions to cmd/weewar-wasm/main.go
2. **Frontend AI Components**: Create AI selection and analysis UI components
3. **AI Game Mode**: Implement AI vs Human and AI vs AI game modes
4. **Performance Testing**: Validate AI performance in web environment

## Conclusion

The WeeWar AI toolkit provides a flexible, performant foundation for computer players while maintaining clean separation from the core game engine. The stateless design enables easy integration with different UI frameworks while the modular architecture supports future enhancements and research directions.

The system successfully balances sophistication with maintainability, providing engaging AI opponents across multiple difficulty levels without compromising the existing game architecture. With the completion of the interactive game viewer foundation, the AI system is ready for full web interface integration.