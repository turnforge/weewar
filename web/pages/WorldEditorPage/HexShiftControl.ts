/**
 * HexShiftControl - A floating hex-shaped control for shifting world coordinates
 *
 * Features:
 * - 6 clickable wedges for each hex direction
 * - Center displays step size N, click to cycle (1, 5, 10, 25)
 * - Compact SVG-based design
 */

import { AxialNeighborDeltas } from '../common/hexUtils';

export interface HexShiftControlOptions {
    /** Root element to render into */
    rootElement: HTMLElement;
    /** Callback when a direction is clicked. Receives (dQ * step, dR * step) */
    onShift: (dQ: number, dR: number) => void;
    /** Initial step size (default: 1) */
    initialStep?: number;
    /** Size of the control in pixels (default: 80) */
    size?: number;
}

const DEFAULT_STEP = 1;
const MIN_STEP = 1;
const MAX_STEP = 50;

// Wedge to direction mapping for pointy-top hex
// Wedge 0: top-right edge -> TOP_RIGHT (index 2)
// Wedge 1: right edge -> RIGHT (index 3)
// Wedge 2: bottom-right edge -> BOTTOM_RIGHT (index 4)
// Wedge 3: bottom-left edge -> BOTTOM_LEFT (index 5)
// Wedge 4: left edge -> LEFT (index 0)
// Wedge 5: top-left edge -> TOP_LEFT (index 1)
const WEDGE_TO_DIRECTION = [2, 3, 4, 5, 0, 1];

// Direction names for tooltips
const DIRECTION_NAMES = ['Left', 'Top-Left', 'Top-Right', 'Right', 'Bottom-Right', 'Bottom-Left'];

/**
 * Generate tooltip text for a wedge
 */
function getWedgeTooltip(wedgeIndex: number, step: number): string {
    const directionIndex = WEDGE_TO_DIRECTION[wedgeIndex];
    const delta = AxialNeighborDeltas[directionIndex];
    const dQ = delta.q * step;
    const dR = delta.r * step;
    const dirName = DIRECTION_NAMES[directionIndex];
    return `${dirName}: dQ=${dQ}, dR=${dR}`;
}

/**
 * Generate the SVG template for the hex shift control
 */
function generateTemplate(size: number, step: number): string {
    const cx = size / 2;
    const cy = size / 2;
    const outerRadius = size / 2 - 2;
    const innerRadius = outerRadius * 0.4;

    // Calculate hex vertices (pointy-top, starting from top going clockwise)
    const hexPoints: [number, number][] = [];
    for (let i = 0; i < 6; i++) {
        const angle = (Math.PI / 3) * i - Math.PI / 2;
        hexPoints.push([
            cx + outerRadius * Math.cos(angle),
            cy + outerRadius * Math.sin(angle)
        ]);
    }

    // Generate wedge paths with tooltips
    const wedgePaths = [];
    for (let i = 0; i < 6; i++) {
        const p1 = hexPoints[i];
        const p2 = hexPoints[(i + 1) % 6];
        const angle1 = (Math.PI / 3) * i - Math.PI / 2;
        const angle2 = (Math.PI / 3) * ((i + 1) % 6) - Math.PI / 2;
        const inner1 = [cx + innerRadius * Math.cos(angle1), cy + innerRadius * Math.sin(angle1)];
        const inner2 = [cx + innerRadius * Math.cos(angle2), cy + innerRadius * Math.sin(angle2)];

        const d = `M ${inner1[0]} ${inner1[1]} L ${p1[0]} ${p1[1]} L ${p2[0]} ${p2[1]} L ${inner2[0]} ${inner2[1]} Z`;
        const tooltip = getWedgeTooltip(i, step);
        wedgePaths.push(`<path data-wedge="${i}" d="${d}" fill="#4a5568" stroke="#2d3748" stroke-width="1" style="transition: fill 0.15s; cursor: pointer;"><title data-wedge-title="${i}">${tooltip}</title></path>`);
    }

    const inputSize = (innerRadius - 2) * 2;
    const inputOffset = cx - (inputSize / 2);

    return `
        <svg width="${size}" height="${size}" viewBox="0 0 ${size} ${size}">
            ${wedgePaths.join('\n            ')}
            <circle data-center="true" cx="${cx}" cy="${cy}" r="${innerRadius - 2}" fill="#2d3748" stroke="#4a5568" stroke-width="2" style="transition: fill 0.15s;"><title>Step size</title></circle>
            <foreignObject x="${inputOffset}" y="${inputOffset}" width="${inputSize}" height="${inputSize}">
                <input data-step-input="true" type="number" value="${step}" min="${MIN_STEP}" max="${MAX_STEP}"
                    style="width: 100%; height: 100%; border: none; background: transparent; color: #e2e8f0;
                           font-size: ${size * 0.18}px; font-weight: bold; font-family: system-ui, sans-serif;
                           text-align: center; outline: none; -moz-appearance: textfield; appearance: textfield;"
                    title="Step size - type a number or use arrows" />
            </foreignObject>
        </svg>
        <style>
            input[data-step-input]::-webkit-outer-spin-button,
            input[data-step-input]::-webkit-inner-spin-button {
                -webkit-appearance: none;
                margin: 0;
            }
        </style>
    `;
}

export class HexShiftControl {
    private rootElement: HTMLElement;
    private step: number;
    private size: number;
    private onShift: (dQ: number, dR: number) => void;

    constructor(options: HexShiftControlOptions) {
        this.rootElement = options.rootElement;
        this.onShift = options.onShift;
        this.step = Math.max(MIN_STEP, Math.min(MAX_STEP, options.initialStep ?? DEFAULT_STEP));
        this.size = options.size ?? 80;

        this.render();
        this.bindEvents();
    }

    public setStep(step: number): void {
        this.step = Math.max(MIN_STEP, Math.min(MAX_STEP, step));
        const input = this.rootElement.querySelector('[data-step-input]') as HTMLInputElement;
        if (input) {
            input.value = String(this.step);
        }
        this.updateWedgeTooltips();
    }

    public getStep(): number {
        return this.step;
    }

    private updateWedgeTooltips(): void {
        for (let i = 0; i < 6; i++) {
            const titleEl = this.rootElement.querySelector(`[data-wedge-title="${i}"]`);
            if (titleEl) {
                titleEl.textContent = getWedgeTooltip(i, this.step);
            }
        }
    }

    private render(): void {
        this.rootElement.innerHTML = generateTemplate(this.size, this.step);
    }

    private bindEvents(): void {
        // Bind wedge events
        const wedges = this.rootElement.querySelectorAll('[data-wedge]');
        wedges.forEach((wedge) => {
            const wedgeIndex = parseInt(wedge.getAttribute('data-wedge') || '0', 10);
            const directionIndex = WEDGE_TO_DIRECTION[wedgeIndex];

            wedge.addEventListener('mouseenter', () => {
                wedge.setAttribute('fill', '#667eea');
            });
            wedge.addEventListener('mouseleave', () => {
                wedge.setAttribute('fill', '#4a5568');
            });
            wedge.addEventListener('click', (e) => {
                e.stopPropagation();
                const delta = AxialNeighborDeltas[directionIndex];
                this.onShift(delta.q * this.step, delta.r * this.step);
            });
        });

        // Bind step input events
        const stepInput = this.rootElement.querySelector('[data-step-input]') as HTMLInputElement;
        if (stepInput) {
            stepInput.addEventListener('change', () => {
                const newStep = parseInt(stepInput.value, 10);
                if (!isNaN(newStep)) {
                    this.step = Math.max(MIN_STEP, Math.min(MAX_STEP, newStep));
                    stepInput.value = String(this.step);
                    this.updateWedgeTooltips();
                }
            });
            // Also update on blur in case user tabs away
            stepInput.addEventListener('blur', () => {
                const newStep = parseInt(stepInput.value, 10);
                if (isNaN(newStep) || newStep < MIN_STEP) {
                    this.step = MIN_STEP;
                } else if (newStep > MAX_STEP) {
                    this.step = MAX_STEP;
                } else {
                    this.step = newStep;
                }
                stepInput.value = String(this.step);
                this.updateWedgeTooltips();
            });
        }
    }

    public destroy(): void {
        this.rootElement.innerHTML = '';
    }
}
