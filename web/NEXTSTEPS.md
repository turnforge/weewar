# Next Steps for Web Module (v8.0)

## Recent Accomplishments ✅
- **E2E Testing Infrastructure**: Complete Playwright-based integration testing for GameViewerPage
  - Real production endpoint testing via `/games/{gameId}/view`
  - Minimal API mocking strategy with surgical fetch patching
  - Proper game lifecycle management (create → test → cleanup)
  - Test separation: Jest (unit) vs Playwright (e2e) with proper configuration
  - Foundation ready for command interface implementation

## Immediate Priorities

### 1. Command Interface Implementation 
- **GameViewerPage Command Interface**: Add high-level testing commands
  - `selectUnit(q, r)` - Select unit at hex coordinates
  - `moveSelectedUnit(q, r)` - Move selected unit to target
  - `attackWithSelectedUnit(q, r)` - Attack target with selected unit
  - `endTurn()` - End current player turn
  - `getGameState()` - Retrieve current game state for assertions
- **LLM-Friendly Failure Reporting**: Structured test result reporting with step tracking
- **Accessibility Benefits**: Command interface improves accessibility for screen readers

### 2. Container Management Fixes  
- **StartGamePage**: Review and fix if similar container issues exist
- **Other Phaser Integration Points**: Audit all pages using PhaserWorldScene

### 3. E2E Test Expansion
- **Additional Test Scenarios**: Implement COMBAT_BASIC, TURN_FLOW, ERROR_HANDLING scenarios
- **Cross-deployment Testing**: Verify tests work against test/production deployments
- **Head-full Debugging**: Enhanced debugging capabilities for test failures

## Short-term Development

### 4. Architecture Consistency
- **Remaining Components**: Ensure all components follow unified patterns established in v8.0
- **Container Targeting**: Standardize container selection patterns across all pages
- **Error Handling**: Consistent error handling for container not found scenarios
- **Debug Logging**: Improve Phaser initialization debug output

### 5. Code Quality
- **Remove Dead Code**: Clean up any remaining references to eliminated wrapper classes
- **Import Cleanup**: Remove unused imports from wrapper elimination
- **Type Safety**: Ensure all method signatures are properly typed after refactoring
- **Documentation**: Update component documentation to reflect new architecture

## Medium-term Goals

### 6. Performance Optimization
- **Phaser Initialization**: Optimize scene startup time and asset loading
- **Container Management**: Minimize DOM queries during initialization
- **Memory Usage**: Profile memory usage after wrapper elimination
- **Rendering Performance**: Benchmark FIT vs RESIZE scale mode performance

### 7. Enhanced Debugging
- **Container Debugging**: Add better error messages for container mismatches
- **Initialization Tracing**: Enhanced logging for Phaser setup process
- **Canvas Placement Validation**: Runtime checks for proper canvas placement
- **Development Tools**: Better debugging tools for Phaser integration issues

### 8. Testing & Validation
- **Integration Tests**: Create tests for Phaser container management
- **Regression Tests**: Ensure wrapper elimination doesn't break existing functionality
- **Canvas Placement Tests**: Automated tests for proper canvas positioning
- **Cross-browser Testing**: Verify fixes work across different browsers

## Long-term Architecture

### 9. Phaser Integration Improvements
- **Scene Management**: Better lifecycle coordination between LCMComponent and Phaser
- **Asset Management**: Centralized asset loading and caching strategies
- **Event Coordination**: Improved integration between Phaser events and EventBus
- **Error Recovery**: Better error handling and recovery for Phaser initialization failures

### 10. Development Experience
- **Hot Reload**: Ensure wrapper elimination doesn't break development hot reload
- **Debugging Tools**: Enhanced tools for debugging Phaser integration issues
- **Documentation**: Comprehensive guides for Phaser integration patterns
- **Template Components**: Reusable patterns for new Phaser-based components

## Technical Debt

### 11. Legacy Cleanup
- **Template Cleanup**: Remove any template references to eliminated wrappers
- **CSS Cleanup**: Remove styles for eliminated wrapper elements
- **Configuration Cleanup**: Remove any configuration for eliminated components
- **Build Optimization**: Remove eliminated files from build process

### 12. Architecture Refinement
- **Pattern Standardization**: Ensure all similar components follow same patterns
- **Interface Consistency**: Standardize interfaces across all Phaser components
- **Error Handling**: Consistent error handling patterns throughout
- **Resource Management**: Proper cleanup and resource management patterns

## Success Metrics

### Functionality Validation
- [ ] All pages render Phaser canvas inside correct container
- [ ] No infinite canvas growth in any scenario
- [ ] World editor functionality fully operational
- [ ] Reference image features working properly
- [ ] No TypeScript compilation errors

### Performance Validation
- [ ] Phaser initialization time maintained or improved
- [ ] Memory usage stable or reduced after wrapper elimination
- [ ] Canvas rendering performance maintained
- [ ] Responsive behavior improved with FIT scale mode

### Code Quality Validation
- [ ] All dead code removed
- [ ] Import statements cleaned up
- [ ] Documentation updated
- [ ] Test coverage maintained
- [ ] Architecture consistency achieved

## Development Guidelines Going Forward

### Container Management
- Always target specific Phaser containers (`#phaser-viewer-container`) not outer wrappers
- Validate container existence before passing to PhaserWorldScene
- Use consistent error handling for missing containers
- Document container hierarchy expectations

### Phaser Integration
- Use PhaserWorldScene as base class for all Phaser functionality
- Extend with specific scene classes (PhaserEditorScene) for specialized features
- Maintain LCMComponent lifecycle integration
- Use FIT scale mode for stable canvas behavior

### Component Development
- Eliminate unnecessary wrapper layers
- Prefer direct method calls over wrapper forwarding
- Maintain clear separation between component and scene responsibilities
- Follow established patterns for consistency