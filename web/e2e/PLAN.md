# GameViewerPage Integration Testing Plan

## 🎯 Goals
1. Test the **real production GameViewerPage** via actual `/games/{gameId}/view` endpoints
2. Create **high-level game actions** abstracted from UI implementation details  
3. Support both **headless and head-full** modes for debugging
4. Enable **LLM-friendly failure reporting** with structured step tracking
5. Test against **real game scenarios** using actual game data

## 📋 Implementation Phases

### Phase 1: Infrastructure Setup ✅ COMPLETE
- [x] **Test Game Creation API**: Helper to create known test game scenarios
- [x] **Server Management**: Manual server management with clear error messages  
- [x] **API Mocking**: Surgical fetch patching framework with blueprint for expansion
- [x] **Basic Playwright Config**: Head/headless modes, test isolation, proper cleanup
- [x] **Game Lifecycle Management**: Create → test → cleanup with proper storage management
- [x] **Persistent Test Worlds**: Setup/cleanup scripts for reusable test world management

### Phase 2: Command Interface ✅ COMPLETE  
- [x] **GameViewerPage Command Interface**: High-level game action methods
  - [x] `selectUnitAt(q, r)` with validation and error reporting
  - [x] `moveSelectedUnitTo(q, r)` with path validation
  - [x] `endCurrentPlayerTurn()` with state management
  - [x] `getGameState()` comprehensive queries
  - [x] `clearSelection()` selection management
- [x] **Unified Architecture**: Single implementation for UI and command interface
- [x] **E2E Testing Utilities**: GameActions and GameTestUtils classes
- [x] **Structured Response Format**: ActionResult interface for consistent error reporting

### Phase 3: Test Scenarios Expansion 🚧 IN PROGRESS
- [ ] **Complete Test World Setup**: Fix world data storage and configuration  
- [ ] **Action Testing**: Comprehensive testing of all command interface methods
- [ ] **Multi-scenario Coverage**: Basic movement, combat, turn flow, error handling
- [ ] **Cross-deployment Testing**: Verify tests work against test/production deployments

### Phase 4: Debugging & Observability
- [ ] **Screenshot Capture**: Automatic screenshots on failures
- [ ] **Step Logging**: Detailed execution traces
- [ ] **LLM Integration**: Structured failure reports for MCP
- [ ] **Video Recording**: Full session recording for complex failures

## 🏗️ Architecture

```
Test Server (Go)  ←→  Production GameViewerPage  ←→  Integration Tests
     ↓                          ↓                        ↓
Test Game APIs         Command Interface           Game Actions API
Real WASM Engine       Real EventBus              Structured Results
Test Database          Real Phaser Scene          Failure Tracking
```

## 📁 File Structure
```
tests/integration/
├── PLAN.md                    # This plan
├── infrastructure/
│   ├── test-server.ts         # Server management
│   ├── test-games.ts          # Game creation utilities  
│   └── api-mocking.ts         # Fetch patching
├── actions/
│   ├── game-actions.ts        # High-level game actions
│   └── command-interface.ts   # GameViewerPage command methods
├── scenarios/
│   ├── basic-gameplay.spec.ts # Core game flow tests
│   ├── combat.spec.ts         # Attack/damage scenarios
│   └── error-handling.spec.ts # Invalid action tests
└── utils/
    ├── debugging.ts           # Screenshots, logging
    └── llm-reporting.ts       # Structured failure output
```

## 🎮 Test Game Scenarios

### Scenario 1: Basic Movement
- **Game ID**: `test-basic-movement`
- **Setup**: 3x3 map, Player 1 unit at (0,0), Player 2 unit at (2,2)
- **Tests**: Select unit, show movement options, execute valid move

### Scenario 2: Combat Engagement  
- **Game ID**: `test-combat-basic`
- **Setup**: Adjacent enemy units ready for combat
- **Tests**: Attack mechanics, health reduction, unit elimination

### Scenario 3: Turn Management
- **Game ID**: `test-turn-flow`
- **Setup**: Multi-unit scenario requiring turn coordination
- **Tests**: End turn, player switching, movement point reset

### Scenario 4: Error Conditions
- **Game ID**: `test-error-handling`  
- **Setup**: Limited movement, blocked paths
- **Tests**: Invalid moves, out of range, occupied spaces

## 🔧 Command Interface Design

Add to `GameViewerPage.ts`:
```typescript
// Command interface for testing (also great for accessibility)
export interface GameViewerCommands {
  selectUnitAt(q: number, r: number): Promise<ActionResult>;
  moveSelectedUnitTo(q: number, r: number): Promise<ActionResult>;
  attackWithSelectedUnit(q: number, r: number): Promise<ActionResult>;  
  endCurrentPlayerTurn(): Promise<ActionResult>;
  getGameState(): GameStateInfo;
  getAvailableActions(): Action[];
}
```

## 📊 LLM-Friendly Failure Format

```json
{
  "testName": "should move unit successfully",
  "scenario": "test-basic-movement", 
  "failed": true,
  "step": "moveUnit.move",
  "action": {
    "type": "moveSelectedUnit",
    "params": {"q": 1, "r": 0},
    "expected": "unit moves to target position",
    "actual": "unit remains at original position"
  },
  "gameState": {
    "currentPlayer": 1,
    "selectedUnit": {"q": 0, "r": 0},
    "turnCounter": 1
  },
  "screenshot": "failure-move-unit-1234.png",
  "trace": ["waitForGameReady", "selectUnit", "moveSelectedUnit"],
  "suggestions": [
    "Check if movement path is blocked",
    "Verify unit has movement points remaining",
    "Ensure target position is valid"
  ]
}
```

## 🚀 Next Steps

**Phase 1** implementation:
1. Create test game creation utilities
2. Set up basic server management  
3. Configure Playwright for production page testing
4. Implement basic API mocking pattern

This approach will give us **maximum confidence** in the real GameViewerPage while providing **excellent debugging capabilities** for complex failure scenarios.