import { BasePage, EventBus, LCMComponent, LifecycleController } from '@panyam/tsappkit';
import { ITheme } from '../assets/themes/BaseTheme';
import DefaultTheme from '../assets/themes/default';
import LilbattleBundle from '../gen/wasmjs';
import { GamesServiceClient } from '../gen/wasmjs/lilbattle/v1/services/gamesServiceClient';
import { SimulateAttackRequest, SimulateAttackResponse } from '../gen/wasmjs/lilbattle/v1/models/interfaces';

/**
 * Attack Simulator Page - Interactive combat simulator
 * Allows users to simulate combat between different units on different terrains
 */
class AttackSimulatorPage extends BasePage {
    private wasmBundle: LilbattleBundle | null = null;
    private gamesClient: GamesServiceClient | null = null;
    private theme: ITheme;

    // Canvas/container elements
    private attackerHexContainer: HTMLElement;
    private defenderHexContainer: HTMLElement;
    private attackerChartCanvas: HTMLCanvasElement;
    private defenderChartCanvas: HTMLCanvasElement;

    // Form elements
    private attackerUnitSelect: HTMLSelectElement;
    private attackerTerrainSelect: HTMLSelectElement;
    private attackerHealthInput: HTMLInputElement;
    private defenderUnitSelect: HTMLSelectElement;
    private defenderTerrainSelect: HTMLSelectElement;
    private defenderHealthInput: HTMLInputElement;
    private woundBonusInput: HTMLInputElement;
    private numSimulationsInput: HTMLInputElement;
    private simulateButton: HTMLButtonElement;

    // Stat elements
    private attackerMeanDamageEl: HTMLElement;
    private attackerKillProbEl: HTMLElement;
    private defenderMeanDamageEl: HTMLElement;
    private defenderKillProbEl: HTMLElement;

    constructor() {
        super('attack-simulator-page', new EventBus(), false);
        this.theme = new DefaultTheme();
    }

    // Override lifecycle methods from BasePage
    protected override initializeSpecificComponents(): LCMComponent[] {
        // Get container elements
        this.attackerHexContainer = document.getElementById('attacker-hex-container')!;
        this.defenderHexContainer = document.getElementById('defender-hex-container')!;
        this.attackerChartCanvas = document.getElementById('attacker-chart') as HTMLCanvasElement;
        this.defenderChartCanvas = document.getElementById('defender-chart') as HTMLCanvasElement;

        // Get form elements
        this.attackerUnitSelect = document.getElementById('attacker-unit') as HTMLSelectElement;
        this.attackerTerrainSelect = document.getElementById('attacker-terrain') as HTMLSelectElement;
        this.attackerHealthInput = document.getElementById('attacker-health') as HTMLInputElement;
        this.defenderUnitSelect = document.getElementById('defender-unit') as HTMLSelectElement;
        this.defenderTerrainSelect = document.getElementById('defender-terrain') as HTMLSelectElement;
        this.defenderHealthInput = document.getElementById('defender-health') as HTMLInputElement;
        this.woundBonusInput = document.getElementById('wound-bonus') as HTMLInputElement;
        this.numSimulationsInput = document.getElementById('num-simulations') as HTMLInputElement;
        this.simulateButton = document.getElementById('simulate-btn') as HTMLButtonElement;

        // Get stat elements
        this.attackerMeanDamageEl = document.getElementById('attacker-mean-damage')!;
        this.attackerKillProbEl = document.getElementById('attacker-kill-prob')!;
        this.defenderMeanDamageEl = document.getElementById('defender-mean-damage')!;
        this.defenderKillProbEl = document.getElementById('defender-kill-prob')!;

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
        this.attackerUnitSelect.addEventListener('change', autoSimulate);
        this.attackerTerrainSelect.addEventListener('change', autoSimulate);
        this.attackerHealthInput.addEventListener('input', autoSimulate);
        this.defenderUnitSelect.addEventListener('change', autoSimulate);
        this.defenderTerrainSelect.addEventListener('change', autoSimulate);
        this.defenderHealthInput.addEventListener('input', autoSimulate);
        this.woundBonusInput.addEventListener('input', autoSimulate);
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
            console.log('[AttackSimulator] Loading WASM bundle...');
            this.wasmBundle = new LilbattleBundle();
            this.gamesClient = new GamesServiceClient(this.wasmBundle);
            await this.wasmBundle.loadWasm('/static/wasm/lilbattle-cli.wasm');
            await this.wasmBundle.waitUntilReady();
            console.log('[AttackSimulator] WASM loaded successfully');
        } catch (error) {
            console.error('[AttackSimulator] Failed to load WASM:', error);
            alert('Failed to load game engine');
        }
    }

    private async runSimulation(): Promise<void> {
        if (!this.gamesClient) {
            console.error('[AttackSimulator] WASM not loaded yet');
            return;
        }

        // Build request
        const request: SimulateAttackRequest = {
            attackerUnitType: parseInt(this.attackerUnitSelect.value),
            attackerTerrain: parseInt(this.attackerTerrainSelect.value),
            attackerHealth: parseInt(this.attackerHealthInput.value),
            defenderUnitType: parseInt(this.defenderUnitSelect.value),
            defenderTerrain: parseInt(this.defenderTerrainSelect.value),
            defenderHealth: parseInt(this.defenderHealthInput.value),
            woundBonus: parseInt(this.woundBonusInput.value),
            numSimulations: parseInt(this.numSimulationsInput.value),
        };

        console.log('[AttackSimulator] Running simulation with:', request);

        try {
            // Call WASM RPC
            const response: SimulateAttackResponse = await this.gamesClient.simulateAttack(request);
            console.log('[AttackSimulator] Simulation result:', response);

            // Update visualizations
            await this.renderHexes(request);
            this.renderCharts(response);
            this.updateStats(response);
        } catch (error) {
            console.error('[AttackSimulator] Simulation failed:', error);
            alert('Simulation failed: ' + error);
        }
    }

    private async renderHexes(request: SimulateAttackRequest): Promise<void> {
        console.log('[AttackSimulator] renderHexes - request:', request);

        // Clear containers
        this.attackerHexContainer.innerHTML = '';
        this.defenderHexContainer.innerHTML = '';

        // Create container divs for terrain and unit
        const attackerTerrainDiv = document.createElement('div');
        attackerTerrainDiv.className = 'relative w-32 h-32 mx-auto';
        this.attackerHexContainer.appendChild(attackerTerrainDiv);

        const defenderTerrainDiv = document.createElement('div');
        defenderTerrainDiv.className = 'relative w-32 h-32 mx-auto';
        this.defenderHexContainer.appendChild(defenderTerrainDiv);

        // Render terrain tiles (player 0 = neutral for terrain)
        console.log('[AttackSimulator] Loading attacker terrain:', request.attackerTerrain);
        await this.theme.setTileImage(request.attackerTerrain, 0, attackerTerrainDiv);
        console.log('[AttackSimulator] Loading defender terrain:', request.defenderTerrain);
        await this.theme.setTileImage(request.defenderTerrain, 0, defenderTerrainDiv);

        // Create overlay divs for units on top of terrain
        const attackerUnitDiv = document.createElement('div');
        attackerUnitDiv.className = 'absolute inset-0 w-full h-full';
        attackerTerrainDiv.appendChild(attackerUnitDiv);

        const defenderUnitDiv = document.createElement('div');
        defenderUnitDiv.className = 'absolute inset-0 w-full h-full';
        defenderTerrainDiv.appendChild(defenderUnitDiv);

        // Render unit images (player 1 = blue for attacker, player 2 = red for defender)
        console.log('[AttackSimulator] Loading attacker unit:', request.attackerUnitType);
        await this.theme.setUnitImage(request.attackerUnitType, 1, attackerUnitDiv);
        console.log('[AttackSimulator] Loading defender unit:', request.defenderUnitType);
        await this.theme.setUnitImage(request.defenderUnitType, 2, defenderUnitDiv);
        console.log('[AttackSimulator] renderHexes complete');

        // Add health labels below the hex
        const attackerHealthLabel = document.createElement('div');
        attackerHealthLabel.className = 'text-center mt-2 font-semibold text-gray-900 dark:text-white';
        attackerHealthLabel.textContent = `HP: ${request.attackerHealth}`;
        this.attackerHexContainer.appendChild(attackerHealthLabel);

        const defenderHealthLabel = document.createElement('div');
        defenderHealthLabel.className = 'text-center mt-2 font-semibold text-gray-900 dark:text-white';
        defenderHealthLabel.textContent = `HP: ${request.defenderHealth}`;
        this.defenderHexContainer.appendChild(defenderHealthLabel);
    }

    private renderCharts(response: SimulateAttackResponse): void {
        console.log('[AttackSimulator] renderCharts - attacker dist:', response.attackerDamageDistribution);
        console.log('[AttackSimulator] renderCharts - defender dist:', response.defenderDamageDistribution);

        // Render attacker damage distribution
        this.drawBarChart(
            this.attackerChartCanvas,
            response.attackerDamageDistribution,
            'Damage to Defender',
            '#3b82f6'
        );

        // Render defender damage distribution
        this.drawBarChart(
            this.defenderChartCanvas,
            response.defenderDamageDistribution,
            'Damage to Attacker',
            '#ef4444'
        );
    }

    private drawBarChart(
        canvas: HTMLCanvasElement,
        distribution: { [key: number]: number },
        title: string,
        color: string
    ): void {
        console.log(`[AttackSimulator] drawBarChart - ${title}:`, distribution);

        const ctx = canvas.getContext('2d')!;
        const width = canvas.width;
        const height = canvas.height;

        // Clear canvas
        ctx.clearRect(0, 0, width, height);

        // Convert distribution to sorted array
        const data = Object.entries(distribution)
            .map(([damage, count]) => ({ damage: parseInt(damage), count }))
            .sort((a, b) => a.damage - b.damage);

        console.log(`[AttackSimulator] drawBarChart - ${title} data array:`, data);

        if (data.length === 0) {
            console.warn(`[AttackSimulator] drawBarChart - ${title} has no data!`);
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

            // Draw damage value below bar
            ctx.fillStyle = this.isDarkMode() ? '#fff' : '#000';
            ctx.font = '10px sans-serif';
            ctx.textAlign = 'center';
            ctx.fillText(d.damage.toString(), x + barWidth * 0.4, height - padding + 15);
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

    private updateStats(response: SimulateAttackResponse): void {
        this.attackerMeanDamageEl.textContent = response.attackerMeanDamage.toFixed(2);
        this.attackerKillProbEl.textContent = (response.attackerKillProbability * 100).toFixed(1) + '%';
        this.defenderMeanDamageEl.textContent = response.defenderMeanDamage.toFixed(2);
        this.defenderKillProbEl.textContent = (response.defenderKillProbability * 100).toFixed(1) + '%';
    }

    private getUnitName(unitId: number): string {
        const option = this.attackerUnitSelect.querySelector(`option[value="${unitId}"]`) ||
                       this.defenderUnitSelect.querySelector(`option[value="${unitId}"]`);
        return option?.textContent || 'Unknown';
    }

    private getTerrainName(terrainId: number): string {
        const option = this.attackerTerrainSelect.querySelector(`option[value="${terrainId}"]`) ||
                       this.defenderTerrainSelect.querySelector(`option[value="${terrainId}"]`);
        return option?.textContent || 'Unknown';
    }
}

AttackSimulatorPage.loadAfterPageLoaded("AttackSimulatorPage", AttackSimulatorPage, "AttackSimulatorPage")
