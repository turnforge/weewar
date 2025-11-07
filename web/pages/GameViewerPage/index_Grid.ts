import { GameViewerPageBase, PanelId } from './GameViewerPageBase';
import { LCMComponent } from '../../lib/LCMComponent';
import { TerrainStatsPanel } from './TerrainStatsPanel';
import { UnitStatsPanel } from './UnitStatsPanel';
import { DamageDistributionPanel } from './DamageDistributionPanel';
import { GameLogPanel } from './GameLogPanel';
import { TurnOptionsPanel } from './TurnOptionsPanel';
import { PhaserGameScene } from './PhaserGameScene';

/**
 * GameViewerPage implementation using CSS Grid for static panel layout.
 *
 * Features:
 * - Fixed panel positions (no drag-and-drop)
 * - Simpler DOM structure
 * - Faster initialization (no DockView overhead)
 * - All containers pre-exist in HTML template
 * - Game scene created early
 */
export class GameViewerPageGrid extends GameViewerPageBase {
    /**
     * Initialize CSS Grid layout - containers already exist in template
     */
    protected async initializeLayout(): Promise<void> {
        // Grid layout is defined in CSS and HTML template
        // All containers pre-exist, so no dynamic setup needed

        // Verify required containers exist
        const requiredIds = [
            'grid-game-scene-container',
            'grid-terrain-stats-container',
            'grid-unit-stats-container',
            'grid-damage-distribution-container',
            'grid-turn-options-container',
            'grid-game-log-container'
        ];

        for (const id of requiredIds) {
            const element = document.getElementById(id);
            if (!element) {
                throw new Error(`GameViewerPageGrid: Required container '${id}' not found in template`);
            }
        }
    }

    /**
     * Create all panel instances and attach to pre-existing containers
     */
    protected createPanels(): LCMComponent[] {
        const panels: LCMComponent[] = [];

        // Terrain Stats Panel
        const terrainElement = document.getElementById('grid-terrain-stats-container');
        if (terrainElement) {
            this.terrainStatsPanel = new TerrainStatsPanel(terrainElement, this.eventBus, true);
            panels.push(this.terrainStatsPanel);
        }

        // Unit Stats Panel
        const unitElement = document.getElementById('grid-unit-stats-container');
        if (unitElement) {
            this.unitStatsPanel = new UnitStatsPanel(unitElement, this.eventBus, true);
            panels.push(this.unitStatsPanel);
        }

        // Damage Distribution Panel
        const damageElement = document.getElementById('grid-damage-distribution-container');
        if (damageElement) {
            this.damageDistributionPanel = new DamageDistributionPanel(damageElement, this.eventBus, true);
            panels.push(this.damageDistributionPanel);
        }

        // Turn Options Panel
        const turnOptionsElement = document.getElementById('grid-turn-options-container');
        if (turnOptionsElement) {
            this.turnOptionsPanel = new TurnOptionsPanel(turnOptionsElement, this.eventBus, true);
            panels.push(this.turnOptionsPanel);
        }

        // Game Log Panel
        const gameLogElement = document.getElementById('grid-game-log-container');
        if (gameLogElement) {
            this.gameLogPanel = new GameLogPanel(gameLogElement, this.eventBus);
            panels.push(this.gameLogPanel);
        }

        return panels;
    }

    /**
     * Get game scene container (pre-exists in template)
     */
    protected getGameSceneContainer(): HTMLElement {
        const container = document.getElementById('grid-game-scene-container');
        if (!container) {
            throw new Error('GameViewerPageGrid: grid-game-scene-container not found');
        }
        return container;
    }

    /**
     * Game scene should be created early (not deferred)
     */
    protected shouldCreateGameSceneEarly(): boolean {
        return true; // Create immediately during performLocalInit
    }

    /**
     * Called after game scene is created
     */
    protected onGameSceneCreated(): void {
        // No additional setup needed for Grid
    }

    /**
     * Show/focus a specific panel (no-op for Grid - all panels always visible)
     */
    protected showPanel(panelId: PanelId): void {
        // In grid layout, all panels are always visible
        // Could scroll to panel if needed
        const element = this.getPanelElement(panelId);
        if (element) {
            element.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
        }
    }

    /**
     * Get panel DOM element
     */
    protected getPanelElement(panelId: PanelId): HTMLElement | null {
        const mapping: Record<PanelId, string> = {
            'terrain-stats': 'grid-terrain-stats-container',
            'unit-stats': 'grid-unit-stats-container',
            'damage-distribution': 'grid-damage-distribution-container',
            'turn-options': 'grid-turn-options-container',
            'game-log': 'grid-game-log-container',
            'build-options': '' // Modal, not in grid
        };

        const containerId = mapping[panelId];
        return containerId ? document.getElementById(containerId) : null;
    }
}

GameViewerPageGrid.loadAfterPageLoaded("gameViewerpage", GameViewerPageGrid, "GameViewerPageGrid")
