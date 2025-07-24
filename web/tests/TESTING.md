# Testing Framework Documentation

## Overview

This document captures our testing decisions, tradeoffs, styles, progress, and learnings for the WeeWar frontend testing infrastructure.

## Architecture Decisions

### Real WASM vs Mocks
**Decision: Use Real WASM Binary**
- ✅ **Chosen Approach**: Load actual `weewar-cli.wasm` in Node.js test environment
- ❌ **Rejected**: Mock WASM functions with JavaScript implementations

**Rationale:**
- Tests the actual Go game logic, not mock behavior
- Catches real WASM compilation and integration issues  
- Ensures JS-WASM communication works correctly
- Provides confidence in rules engine correctness across different maps

**Tradeoffs:**
- Slightly slower test execution (~3-4 seconds vs <1 second for mocks)
- Requires WASM binary to be built before tests
- More complex Node.js environment setup

### Testing Framework Choice
**Decision: Jest + TypeScript + JSDOM**
- Jest for test runner and assertions
- TypeScript for type safety in tests
- JSDOM for DOM environment simulation
- No browser required - runs in Node.js

### Test Structure
**Decision: Component-based test organization**
```
tests/
├── gameState.test.ts      # GameState component tests
├── rulesEngine.test.ts    # Rules validation across maps
├── wasmLoading.test.ts    # WASM integration verification
├── helpers/
│   └── wasmTestUtils.ts   # Reusable test utilities
└── fixtures/
    └── testMaps.ts        # Standardized test map data
```

## Implementation Approach

### WASM Loading Strategy
**Pre-loading Pattern:**
1. Load `wasm_exec.js` to get Go runtime
2. Load WASM binary from filesystem 
3. Instantiate and run WASM module
4. Copy WASM functions from `global` to `window` for GameState compatibility
5. GameState detects pre-loaded WASM and skips its own loading

### Test Map Standardization
**Fixture-based Testing:**
- `SMALL_TEST_MAP`: 3x3 basic functionality testing
- `MEDIUM_TEST_MAP`: 5x5 varied terrain for pathfinding
- `COMBAT_TEST_MAP`: Adjacent units for combat testing

**Data Format Alignment:**
- Frontend uses `{q, r, tileType, player}` format
- WASM expects `{q, r, tile_type, player}` format  
- Conversion handled in `testMapToJSON()` function

### GameState Integration
**Modified Initialization Order:**
- Fixed constructor to initialize `gameData` before calling `super()`
- Added WASM pre-loading detection to skip DOM-based loading
- Separated WASM loading (`wasmLoaded`) from game initialization (`gameInitialized`)

## Current Test Coverage

### ✅ Working Tests
- **WASM Loading**: Verifies 11 WASM functions are available
- **Game Creation**: Creates games from standardized test maps
- **Movement Validation**: Tests movement option consistency across maps
- **Basic Error Handling**: Tests invalid coordinates and edge cases

### 🚧 In Progress Tests  
- **Attack Validation**: Some tests commented out pending fixes
- **Turn Management**: Partial implementation
- **Complex Scenarios**: Multi-turn game state consistency

### ⏳ Future Tests
- **Performance Testing**: WASM performance benchmarks
- **AI vs AI**: Automated gameplay testing
- **Regression Testing**: Comprehensive rule change validation

## Key Learnings

### WASM Integration Challenges
1. **Response Object Compatibility**: Node.js `fetch` mock needed proper Response-like object for `WebAssembly.instantiateStreaming`
2. **Global vs Window**: WASM functions appear on `global` but GameState expects them on `window`
3. **Initialization Timing**: GameState constructor order matters for test environment setup

### Data Format Challenges  
1. **JSON Structure**: Frontend and WASM use slightly different field naming conventions
2. **Type Compatibility**: Need consistent TypeScript interfaces between test fixtures and components
3. **Coordinate Systems**: Hex coordinate validation across different map representations

### Test Design Principles
1. **Real Data**: Use actual game logic instead of mocks where possible
2. **Isolation**: Each test gets fresh GameState instance to prevent cross-contamination
3. **Validation**: Comprehensive validation helpers for consistent assertions
4. **Documentation**: Self-documenting test names and clear error messages

## Performance Metrics

### Test Execution Times
- Full test suite: ~15-20 seconds
- WASM loading: ~1-2 seconds per instance
- Game creation: ~100-200ms per map
- Movement validation: ~50-100ms per test

### WASM Binary Stats
- Size: 11.7MB (`weewar-cli.wasm`)
- Loading time: ~1 second in Node.js
- Memory usage: Acceptable for test environment

## Style Guide

### Test Naming
```typescript
describe('Component Name', () => {
  describe('Feature Category', () => {
    test('should perform specific action under specific conditions', () => {
      // Test implementation
    });
  });
});
```

### Validation Pattern
```typescript
const response = gameState.getMovementOptions(q, r);
const validation = validateMovementOptions(response);
expect(validation.isValid).toBe(true);
expect(validation.coordinateCount).toBeGreaterThanOrEqual(0);
```

### Cleanup Pattern
```typescript
let gameState: GameState;
let cleanup: () => void;

beforeEach(async () => {
  const testSetup = await createTestGameState();
  gameState = testSetup.gameState;
  cleanup = testSetup.cleanup;
});

afterEach(() => {
  if (cleanup) cleanup();
});
```

## Future Improvements

### Short Term
1. Fix remaining commented-out tests
2. Add more edge case coverage
3. Improve error message quality
4. Add performance benchmarks

### Medium Term  
1. Integrate with CI/CD pipeline
2. Add visual regression testing for highlights
3. Create comprehensive map generator for testing
4. Add mutation testing for rule validation

### Long Term
1. Property-based testing for game rules
2. Automated gameplay simulation
3. Performance regression detection
4. Cross-browser WASM compatibility testing

## Commands

### Run All Tests
```bash
npm test
```

### Run Specific Test File
```bash
npm test -- --testPathPattern=gameState.test.ts
```

### Run Tests with Verbose Output
```bash
npm test -- --verbose
```

### Run Tests Matching Pattern
```bash
npm test -- --testNamePattern="should load WASM"
```

## Dependencies

### Core Testing
- `jest`: Test runner and assertions
- `ts-jest`: TypeScript integration
- `jest-environment-jsdom`: DOM environment simulation

### WASM Support
- Node.js built-in `WebAssembly` API
- File system access for WASM binary loading
- `TextEncoder`/`TextDecoder` polyfills for Node.js

### Type Safety
- `@types/jest`: Jest type definitions
- TypeScript interfaces for game data validation
- Comprehensive validation helper functions

---

*Last updated: January 2025*
*Testing framework status: Core functionality complete, expanding coverage*