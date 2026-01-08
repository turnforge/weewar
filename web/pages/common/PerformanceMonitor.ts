import * as Phaser from 'phaser';

/**
 * Performance monitoring utility for measuring frame times, FPS, and function execution.
 * Use this to establish baselines before optimization and verify improvements after.
 */
export class PerformanceMonitor {
    private enabled: boolean = false;
    private frameCount: number = 0;
    private lastFpsTime: number = 0;
    private fps: number = 0;
    private frameTimes: number[] = [];
    private lastFrameTime: number = 0;

    // Function timing
    private functionTimings: Map<string, { total: number; count: number; max: number }> = new Map();

    // Object tracking
    private objectCounts: Map<string, { created: number; destroyed: number }> = new Map();

    // Display element
    private displayElement: HTMLDivElement | null = null;

    // Phaser scene reference for object counting
    private phaserScene: Phaser.Scene | null = null;

    constructor() {
        this.lastFpsTime = performance.now();
        this.lastFrameTime = performance.now();
    }

    /**
     * Register a Phaser scene to monitor its GameObjects
     */
    setScene(scene: Phaser.Scene): void {
        this.phaserScene = scene;
    }

    enable(): void {
        this.enabled = true;
        this.createDisplayElement();
        console.log('[PerfMon] Performance monitoring enabled');
    }

    disable(): void {
        this.enabled = false;
        this.removeDisplayElement();
        console.log('[PerfMon] Performance monitoring disabled');
    }

    toggle(): void {
        if (this.enabled) {
            this.disable();
        } else {
            this.enable();
        }
    }

    isEnabled(): boolean {
        return this.enabled;
    }

    /**
     * Call at the start of each frame (in update())
     */
    frameStart(): void {
        if (!this.enabled) return;
        this.lastFrameTime = performance.now();
    }

    /**
     * Call at the end of each frame (in update())
     */
    frameEnd(): void {
        if (!this.enabled) return;

        const now = performance.now();
        const frameTime = now - this.lastFrameTime;

        // Track frame times (keep last 60)
        this.frameTimes.push(frameTime);
        if (this.frameTimes.length > 60) {
            this.frameTimes.shift();
        }

        // Calculate FPS every second
        this.frameCount++;
        if (now - this.lastFpsTime >= 1000) {
            this.fps = this.frameCount;
            this.frameCount = 0;
            this.lastFpsTime = now;
            this.updateDisplay();
        }
    }

    /**
     * Time a specific function. Returns the result of the function.
     */
    time<T>(name: string, fn: () => T): T {
        if (!this.enabled) return fn();

        const start = performance.now();
        const result = fn();
        const elapsed = performance.now() - start;

        let timing = this.functionTimings.get(name);
        if (!timing) {
            timing = { total: 0, count: 0, max: 0 };
            this.functionTimings.set(name, timing);
        }
        timing.total += elapsed;
        timing.count++;
        timing.max = Math.max(timing.max, elapsed);

        return result;
    }

    /**
     * Track object creation
     */
    trackCreate(type: string, count: number = 1): void {
        if (!this.enabled) return;

        let counts = this.objectCounts.get(type);
        if (!counts) {
            counts = { created: 0, destroyed: 0 };
            this.objectCounts.set(type, counts);
        }
        counts.created += count;
    }

    /**
     * Track object destruction
     */
    trackDestroy(type: string, count: number = 1): void {
        if (!this.enabled) return;

        let counts = this.objectCounts.get(type);
        if (!counts) {
            counts = { created: 0, destroyed: 0 };
            this.objectCounts.set(type, counts);
        }
        counts.destroyed += count;
    }

    /**
     * Reset all metrics (call this to start a fresh measurement)
     */
    reset(): void {
        this.functionTimings.clear();
        this.objectCounts.clear();
        this.frameTimes = [];
        console.log('[PerfMon] Metrics reset');
    }

    /**
     * Get a summary report
     */
    getReport(): string {
        const avgFrameTime = this.frameTimes.length > 0
            ? this.frameTimes.reduce((a, b) => a + b, 0) / this.frameTimes.length
            : 0;
        const maxFrameTime = this.frameTimes.length > 0
            ? Math.max(...this.frameTimes)
            : 0;

        let report = `=== Performance Report ===\n`;
        report += `FPS: ${this.fps}\n`;
        report += `Avg Frame Time: ${avgFrameTime.toFixed(2)}ms\n`;
        report += `Max Frame Time: ${maxFrameTime.toFixed(2)}ms\n`;
        report += `\n--- Function Timings (per second) ---\n`;

        for (const [name, timing] of this.functionTimings) {
            const avg = timing.count > 0 ? timing.total / timing.count : 0;
            report += `${name}: avg=${avg.toFixed(3)}ms, max=${timing.max.toFixed(3)}ms, calls=${timing.count}\n`;
        }

        report += `\n--- Object Counts (since reset) ---\n`;
        for (const [type, counts] of this.objectCounts) {
            report += `${type}: created=${counts.created}, destroyed=${counts.destroyed}\n`;
        }

        return report;
    }

    /**
     * Log the report to console
     */
    logReport(): void {
        console.log(this.getReport());
    }

    private createDisplayElement(): void {
        if (this.displayElement) return;

        this.displayElement = document.createElement('div');
        this.displayElement.id = 'perf-monitor';
        this.displayElement.style.cssText = `
            position: fixed;
            top: 10px;
            right: 10px;
            background: rgba(0, 0, 0, 0.8);
            color: #0f0;
            font-family: monospace;
            font-size: 12px;
            padding: 10px;
            border-radius: 5px;
            z-index: 10000;
            min-width: 200px;
            white-space: pre;
        `;
        document.body.appendChild(this.displayElement);
    }

    private removeDisplayElement(): void {
        if (this.displayElement) {
            this.displayElement.remove();
            this.displayElement = null;
        }
    }

    /**
     * Get counts of various Phaser GameObject types
     */
    private getPhaserObjectCounts(): { sprites: number; texts: number; graphics: number; total: number; poolSize: number } | null {
        if (!this.phaserScene) return null;

        const children = this.phaserScene.children.list;
        let sprites = 0;
        let texts = 0;
        let graphics = 0;

        for (const child of children) {
            if (child instanceof Phaser.GameObjects.Sprite) {
                sprites++;
            } else if (child instanceof Phaser.GameObjects.Text) {
                texts++;
            } else if (child instanceof Phaser.GameObjects.Graphics) {
                graphics++;
            }
        }

        // Try to get pool size from scene if available
        let poolSize = 0;
        const scene = this.phaserScene as any;
        if (scene.coordinateTextPool) {
            poolSize = scene.coordinateTextPool.length;
        }

        return {
            sprites,
            texts,
            graphics,
            total: children.length,
            poolSize
        };
    }

    private updateDisplay(): void {
        if (!this.displayElement) return;

        const avgFrameTime = this.frameTimes.length > 0
            ? this.frameTimes.reduce((a, b) => a + b, 0) / this.frameTimes.length
            : 0;

        let display = `FPS: ${this.fps}\n`;
        display += `Frame: ${avgFrameTime.toFixed(2)}ms\n`;

        // Show top 5 slowest functions
        const sortedTimings = Array.from(this.functionTimings.entries())
            .map(([name, t]) => ({ name, avg: t.count > 0 ? t.total / t.count : 0 }))
            .sort((a, b) => b.avg - a.avg)
            .slice(0, 5);

        if (sortedTimings.length > 0) {
            display += `\n-- Slow Functions --\n`;
            for (const t of sortedTimings) {
                display += `${t.name}: ${t.avg.toFixed(2)}ms\n`;
            }
        }

        // Show object churn
        let totalCreated = 0;
        let totalDestroyed = 0;
        for (const counts of this.objectCounts.values()) {
            totalCreated += counts.created;
            totalDestroyed += counts.destroyed;
        }
        if (totalCreated > 0 || totalDestroyed > 0) {
            display += `\n-- Objects/sec --\n`;
            display += `Created: ${totalCreated}\n`;
            display += `Destroyed: ${totalDestroyed}\n`;
        }

        // Show Phaser GameObject counts
        if (this.phaserScene) {
            const phaserCounts = this.getPhaserObjectCounts();
            if (phaserCounts) {
                display += `\n-- Phaser Objects --\n`;
                display += `Sprites: ${phaserCounts.sprites}\n`;
                display += `Texts: ${phaserCounts.texts}\n`;
                display += `Graphics: ${phaserCounts.graphics}\n`;
                display += `Total: ${phaserCounts.total}\n`;
                if (phaserCounts.poolSize > 0) {
                    display += `Pool: ${phaserCounts.poolSize}\n`;
                }
            }
        }

        this.displayElement.textContent = display;

        // Reset per-second counters
        this.functionTimings.clear();
        this.objectCounts.clear();
    }
}

// Global instance for easy access
export const perfMon = new PerformanceMonitor();

// Expose to window for console access
if (typeof window !== 'undefined') {
    (window as any).perfMon = perfMon;
}
