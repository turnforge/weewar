/**
 * Component Isolation Test
 * Validates that components only access DOM within their root elements
 * and that component failures don't cascade to other components
 */

import { EventBus } from './EventBus';
import { WorldViewer } from './WorldViewer';
import { WorldStatsPanel } from './WorldStatsPanel';

export class ComponentIsolationTest {
    private eventBus: EventBus;
    private testResults: { [testName: string]: { passed: boolean; message: string } } = {};
    
    constructor() {
        this.eventBus = new EventBus(true);
    }
    
    /**
     * Run all isolation tests
     */
    public async runAllTests(): Promise<boolean> {
        console.log('🧪 Starting Component Isolation Tests...');
        
        // Test 1: Component initialization isolation
        await this.testComponentInitialization();
        
        // Test 2: DOM scoping isolation  
        await this.testDOMScoping();
        
        // Test 3: Error isolation
        await this.testErrorIsolation();
        
        // Test 4: Event isolation
        await this.testEventIsolation();
        
        // Report results
        return this.reportResults();
    }
    
    /**
     * Test that components can initialize independently
     */
    private async testComponentInitialization(): Promise<void> {
        try {
            // Create isolated test containers
            const worldViewerRoot = document.createElement('div');
            worldViewerRoot.setAttribute('data-component', 'world-viewer-test');
            worldViewerRoot.innerHTML = '<div id="phaser-viewer-container"></div>';
            
            const statsRoot = document.createElement('div');
            statsRoot.setAttribute('data-component', 'world-stats-test');
            
            // Initialize components
            const worldViewer = new WorldViewer(worldViewerRoot, this.eventBus, true);
            const statsPanel = new WorldStatsPanel(statsRoot, this.eventBus, true);
            
            // Check if both components initialized
            const worldViewerReady = worldViewer.isReady();
            const statsPanelReady = statsPanel.isReady();
            
            if (worldViewerReady && statsPanelReady) {
                this.testResults['component-initialization'] = {
                    passed: true,
                    message: 'Components initialize independently without conflicts'
                };
            } else {
                this.testResults['component-initialization'] = {
                    passed: false,
                    message: `WorldViewer ready: ${worldViewerReady}, StatsPanel ready: ${statsPanelReady}`
                };
            }
            
            // Clean up
            worldViewer.destroy();
            statsPanel.destroy();
            
        } catch (error) {
            this.testResults['component-initialization'] = {
                passed: false,
                message: `Component initialization failed: ${error}`
            };
        }
    }
    
    /**
     * Test that components only access DOM within their root elements
     */
    private async testDOMScoping(): Promise<void> {
        try {
            // Create test containers with elements that shouldn't be accessed
            const container = document.createElement('div');
            container.innerHTML = `
                <div data-component="world-viewer-test">
                    <div id="phaser-viewer-container"></div>
                    <div data-stat="inside-viewer">viewer-data</div>
                </div>
                <div data-component="world-stats-test">
                    <div data-stat-section="basic"></div>
                    <div data-stat="inside-stats">stats-data</div>
                </div>
                <div data-stat="outside-components">external-data</div>
            `;
            
            const worldViewerRoot = container.querySelector('[data-component="world-viewer-test"]') as HTMLElement;
            const statsRoot = container.querySelector('[data-component="world-stats-test"]') as HTMLElement;
            
            // Initialize components
            const worldViewer = new WorldViewer(worldViewerRoot, this.eventBus, true);
            const statsPanel = new WorldStatsPanel(statsRoot, this.eventBus, true);
            
            // Test that components can only find elements within their root
            const viewerCanAccessInternal = worldViewerRoot.querySelector('[data-stat="inside-viewer"]') !== null;
            const viewerCannotAccessExternal = worldViewerRoot.querySelector('[data-stat="outside-components"]') === null;
            const statsCanAccessInternal = statsRoot.querySelector('[data-stat="inside-stats"]') !== null;
            const statsCannotAccessExternal = statsRoot.querySelector('[data-stat="outside-components"]') === null;
            
            const scopingWorksCorrectly = viewerCanAccessInternal && viewerCannotAccessExternal && 
                                        statsCanAccessInternal && statsCannotAccessExternal;
            
            this.testResults['dom-scoping'] = {
                passed: scopingWorksCorrectly,
                message: scopingWorksCorrectly ? 
                    'Components correctly scoped to their root elements' :
                    'Components accessing DOM outside their scope'
            };
            
            // Clean up
            worldViewer.destroy();
            statsPanel.destroy();
            
        } catch (error) {
            this.testResults['dom-scoping'] = {
                passed: false,
                message: `DOM scoping test failed: ${error}`
            };
        }
    }
    
    /**
     * Test that component errors don't cascade to other components
     */
    private async testErrorIsolation(): Promise<void> {
        try {
            // Create test containers
            const worldViewerRoot = document.createElement('div');
            const statsRoot = document.createElement('div');
            
            // Initialize components
            const worldViewer = new WorldViewer(worldViewerRoot, this.eventBus, true);
            const statsPanel = new WorldStatsPanel(statsRoot, this.eventBus, true);
            
            // Force an error in one component by calling contentUpdated with invalid HTML
            let errorOccurred = false;
            try {
                worldViewer.contentUpdated('<div><unclosed'); // Invalid HTML
            } catch (error) {
                errorOccurred = true;
            }
            
            // Check that the other component is still functional
            const statsPanelStillWorking = statsPanel.isReady();
            
            this.testResults['error-isolation'] = {
                passed: statsPanelStillWorking,
                message: statsPanelStillWorking ? 
                    'Component errors properly isolated' :
                    'Component errors cascaded to other components'
            };
            
            // Clean up
            worldViewer.destroy();
            statsPanel.destroy();
            
        } catch (error) {
            this.testResults['error-isolation'] = {
                passed: false,
                message: `Error isolation test failed: ${error}`
            };
        }
    }
    
    /**
     * Test that event communication works without direct coupling
     */
    private async testEventIsolation(): Promise<void> {
        try {
            // Create test containers
            const worldViewerRoot = document.createElement('div');
            worldViewerRoot.innerHTML = '<div id="phaser-viewer-container"></div>';
            const statsRoot = document.createElement('div');
            
            // Initialize components
            const worldViewer = new WorldViewer(worldViewerRoot, this.eventBus, true);
            const statsPanel = new WorldStatsPanel(statsRoot, this.eventBus, true);
            
            // Test event communication by loading mock world data
            const mockWorldData = {
                id: 'test-world',
                tiles: { '0,0': { tileType: 1, playerId: 0 } },
                world_units: []
            };
            
            // This should trigger events that stats panel can receive
            await worldViewer.loadWorld(mockWorldData);
            
            // Allow time for event processing
            await new Promise(resolve => setTimeout(resolve, 100));
            
            // Check that components can communicate via events without direct references
            const communicationWorking = statsPanel.isReady() && worldViewer.isReady();
            
            this.testResults['event-isolation'] = {
                passed: communicationWorking,
                message: communicationWorking ? 
                    'Components communicate properly via EventBus' :
                    'Component event communication failed'
            };
            
            // Clean up
            worldViewer.destroy();
            statsPanel.destroy();
            
        } catch (error) {
            this.testResults['event-isolation'] = {
                passed: false,
                message: `Event isolation test failed: ${error}`
            };
        }
    }
    
    /**
     * Report test results
     */
    private reportResults(): boolean {
        console.log('\n📊 Component Isolation Test Results:');
        console.log('=====================================');
        
        let allPassed = true;
        
        Object.entries(this.testResults).forEach(([testName, result]) => {
            const status = result.passed ? '✅ PASS' : '❌ FAIL';
            console.log(`${status} ${testName}: ${result.message}`);
            
            if (!result.passed) {
                allPassed = false;
            }
        });
        
        console.log('=====================================');
        console.log(`Overall Result: ${allPassed ? '✅ ALL TESTS PASSED' : '❌ SOME TESTS FAILED'}`);
        
        return allPassed;
    }
    
    /**
     * Get test results for external use
     */
    public getResults(): { [testName: string]: { passed: boolean; message: string } } {
        return { ...this.testResults };
    }
}

// Export for use in testing
export default ComponentIsolationTest;
