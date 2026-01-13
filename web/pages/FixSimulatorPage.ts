import { BasePage, EventBus, LCMComponent, LifecycleController } from '@panyam/tsappkit';
import { ITheme } from '../assets/themes/BaseTheme';
import DefaultTheme from '../assets/themes/default';
import LilbattleBundle from '../gen/wasmjs';
import { GamesServiceClient } from '../gen/wasmjs/lilbattle/v1/services/gamesServiceClient';
import { SimulateFixRequest, SimulateFixResponse } from '../gen/wasmjs/lilbattle/v1/models/interfaces';

/**
 * Fix Simulator Page - Interactive repair simulator
 * Allows users to simulate fix (repair) outcomes between different units
 */
class FixSimulatorPage extends BasePage {
    private wasmBundle: LilbattleBundle | null = null;
    private gamesClient: GamesServiceClient | null = null;
    private theme: ITheme;

    // Canvas/container elements
    private fixingHexContainer: HTMLElement;
    private injuredHexContainer: HTMLElement;
    private healingChartCanvas: HTMLCanvasElement;

    // Form elements
    private fixingUnitSelect: HTMLSelectElement;
    private fixingUnitHealthInput: HTMLInputElement;
    private injuredUnitSelect: HTMLSelectElement;
    private numSimulationsInput: HTMLInputElement;
    private simulateButton: HTMLButtonElement;

    // Stat elements
    private meanHealingEl: HTMLElement;
    private fixValueEl: HTMLElement;

    // Description elements
    private fixingUnitNameEl: HTMLElement;
    private injuredUnitNameEl: HTMLElement;

    constructor() {
        super('fix-simulator-page', new EventBus(), false);
        this.theme = new DefaultTheme();
    }

    // Override lifecycle methods from BasePage
    protected override initializeSpecificComponents(): LCMComponent[] {
        // Get container elements
        this.fixingHexContainer = document.getElementById('fixing-hex-container')!;
        this.injuredHexContainer = document.getElementById('injured-hex-container')!;
        this.healingChartCanvas = document.getElementById('healing-chart') as HTMLCanvasElement;

        // Get form elements
        this.fixingUnitSelect = document.getElementById('fixing-unit') as HTMLSelectElement;
        this.fixingUnitHealthInput = document.getElementById('fixing-unit-health') as HTMLInputElement;
        this.injuredUnitSelect = document.getElementById('injured-unit') as HTMLSelectElement;
        this.numSimulationsInput = document.getElementById('num-simulations') as HTMLInputElement;
        this.simulateButton = document.getElementById('simulate-btn') as HTMLButtonElement;

        // Get stat elements
        this.meanHealingEl = document.getElementById('mean-healing')!;
        this.fixValueEl = document.getElementById('fix-value')!;

        // Get description elements
        this.fixingUnitNameEl = document.getElementById('fixing-unit-name')!;
        this.injuredUnitNameEl = document.getElementById('injured-unit-name')!;

        // Initialize async components
        this.initAsync();

        // No child components
        return [];
    }

    protected override bindSpecificEvents(): void {
        // Simulate button
        this.simulateButton.addEventListener('click', () => this.runSimulation());

        // Auto-simulate on form changes
        const autoSimulate = () => this.runSimulation();
        this.fixingUnitSelect.addEventListener('change', autoSimulate);
        this.fixingUnitHealthInput.addEventListener('input', autoSimulate);
        this.injuredUnitSelect.addEventListener('change', autoSimulate);
        this.numSimulationsInput.addEventListener('input', autoSimulate);
    }

    private async initAsync(): Promise<void> {
        // Load WASM
        await this.loadWASM();

        // Run initial simulation (dropdowns are already populated server-side)
        await this.runSimulation();
    }

    private async loadWASM(): Promise<void> {
        try {
            console.log('[FixSimulator] Loading WASM bundle...');
            this.wasmBundle = new LilbattleBundle();
            this.gamesClient = new GamesServiceClient(this.wasmBundle);
            await this.wasmBundle.loadWasm('/static/wasm/lilbattle-cli.wasm');
            await this.wasmBundle.waitUntilReady();
            console.log('[FixSimulator] WASM loaded successfully');
        } catch (error) {
            console.error('[FixSimulator] Failed to load WASM:', error);
            alert('Failed to load game engine');
        }
    }

    private async runSimulation(): Promise<void> {
        if (!this.gamesClient) {
            console.error('[FixSimulator] WASM not loaded yet');
            return;
        }

        // Build request
        const request: SimulateFixRequest = {
            fixingUnitType: parseInt(this.fixingUnitSelect.value),
            fixingUnitHealth: parseInt(this.fixingUnitHealthInput.value),
            injuredUnitType: parseInt(this.injuredUnitSelect.value),
            numSimulations: parseInt(this.numSimulationsInput.value),
        };

        console.log('[FixSimulator] Running simulation with:', request);

        try {
            // Call WASM RPC
            const response: SimulateFixResponse = await this.gamesClient.simulateFix(request);
            console.log('[FixSimulator] Simulation result:', response);

            // Update visualizations
            await this.renderHexes(request);
            this.renderChart(response);
            this.updateStats(response);
            this.updateDescription();
        } catch (error) {
            console.error('[FixSimulator] Simulation failed:', error);
            alert('Simulation failed: ' + error);
        }
    }

    private async renderHexes(request: SimulateFixRequest): Promise<void> {
        console.log('[FixSimulator] renderHexes - request:', request);

        // Clear containers
        this.fixingHexContainer.innerHTML = '';
        this.injuredHexContainer.innerHTML = '';

        // Create container divs for units (use grass terrain as background)
        const fixingTerrainDiv = document.createElement('div');
        fixingTerrainDiv.className = 'relative w-32 h-32 mx-auto';
        this.fixingHexContainer.appendChild(fixingTerrainDiv);

        const injuredTerrainDiv = document.createElement('div');
        injuredTerrainDiv.className = 'relative w-32 h-32 mx-auto';
        this.injuredHexContainer.appendChild(injuredTerrainDiv);

        // Render grass terrain (ID 1) as background for both
        console.log('[FixSimulator] Loading terrain backgrounds');
        await this.theme.setTileImage(1, 0, fixingTerrainDiv);
        await this.theme.setTileImage(1, 0, injuredTerrainDiv);

        // Create overlay divs for units on top of terrain
        const fixingUnitDiv = document.createElement('div');
        fixingUnitDiv.className = 'absolute inset-0 w-full h-full';
        fixingTerrainDiv.appendChild(fixingUnitDiv);

        const injuredUnitDiv = document.createElement('div');
        injuredUnitDiv.className = 'absolute inset-0 w-full h-full';
        injuredTerrainDiv.appendChild(injuredUnitDiv);

        // Render unit images (both use player 1 = blue since they're friendly)
        console.log('[FixSimulator] Loading fixing unit:', request.fixingUnitType);
        await this.theme.setUnitImage(request.fixingUnitType, 1, fixingUnitDiv);
        console.log('[FixSimulator] Loading injured unit:', request.injuredUnitType);
        await this.theme.setUnitImage(request.injuredUnitType, 1, injuredUnitDiv);
        console.log('[FixSimulator] renderHexes complete');

        // Add health label below the fixing unit hex
        const fixingHealthLabel = document.createElement('div');
        fixingHealthLabel.className = 'text-center mt-2 font-semibold text-gray-900 dark:text-white';
        fixingHealthLabel.textContent = `HP: ${request.fixingUnitHealth}`;
        this.fixingHexContainer.appendChild(fixingHealthLabel);

        // Injured unit doesn't need health label (it's about restoration, not current health)
    }

    private renderChart(response: SimulateFixResponse): void {
        console.log('[FixSimulator] renderChart - healing dist:', response.healingDistribution);

        // Render healing distribution
        this.drawBarChart(
            this.healingChartCanvas,
            response.healingDistribution,
            'Health Restored',
            '#10b981' // Green color for healing
        );
    }

    private drawBarChart(
        canvas: HTMLCanvasElement,
        distribution: { [key: number]: number },
        title: string,
        color: string
    ): void {
        console.log(`[FixSimulator] drawBarChart - ${title}:`, distribution);

        const ctx = canvas.getContext('2d')!;
        const width = canvas.width;
        const height = canvas.height;

        // Clear canvas
        ctx.clearRect(0, 0, width, height);

        // Convert distribution to sorted array
        const data = Object.entries(distribution)
            .map(([healing, count]) => ({ healing: parseInt(healing), count }))
            .sort((a, b) => a.healing - b.healing);

        console.log(`[FixSimulator] drawBarChart - ${title} data array:`, data);

        if (data.length === 0) {
            console.warn(`[FixSimulator] drawBarChart - ${title} has no data!`);
            // Draw a message indicating no data
            ctx.fillStyle = this.isDarkMode() ? '#fff' : '#000';
            ctx.font = '14px sans-serif';
            ctx.textAlign = 'center';
            ctx.fillText('This unit cannot perform repairs (Fix Value = 0)', width / 2, height / 2);
            return;
        }

        // Chart dimensions
        const padding = 40;
        const chartWidth = width - 2 * padding;
        const chartHeight = height - 2 * padding;

        // Find max count for scaling
        const maxCount = Math.max(...data.map(d => d.count));

        // Bar width
        const barWidth = chartWidth / data.length;

        // Draw bars
        data.forEach((d, i) => {
            const barHeight = (d.count / maxCount) * chartHeight;
            const x = padding + i * barWidth;
            const y = height - padding - barHeight;

            ctx.fillStyle = color;
            ctx.fillRect(x, y, barWidth * 0.8, barHeight);

            // Draw healing value below bar
            ctx.fillStyle = this.isDarkMode() ? '#fff' : '#000';
            ctx.font = '10px sans-serif';
            ctx.textAlign = 'center';
            ctx.fillText(d.healing.toString(), x + barWidth * 0.4, height - padding + 15);
        });

        // Draw axes
        ctx.strokeStyle = this.isDarkMode() ? '#666' : '#ccc';
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.moveTo(padding, height - padding);
        ctx.lineTo(width - padding, height - padding);
        ctx.moveTo(padding, padding);
        ctx.lineTo(padding, height - padding);
        ctx.stroke();
    }

    private updateStats(response: SimulateFixResponse): void {
        this.meanHealingEl.textContent = response.meanHealing.toFixed(2);
        this.fixValueEl.textContent = response.fixValue.toString();
    }

    private updateDescription(): void {
        // Update the unit names in the description
        const fixingUnitOption = this.fixingUnitSelect.selectedOptions[0];
        const injuredUnitOption = this.injuredUnitSelect.selectedOptions[0];

        // Extract just the unit name (before any parentheses)
        const fixingUnitName = fixingUnitOption?.textContent?.split('(')[0]?.trim() || 'Unknown';
        const injuredUnitName = injuredUnitOption?.textContent || 'Unknown';

        this.fixingUnitNameEl.textContent = fixingUnitName;
        this.injuredUnitNameEl.textContent = injuredUnitName;
    }
}

FixSimulatorPage.loadAfterPageLoaded("FixSimulatorPage", FixSimulatorPage, "FixSimulatorPage")
