/**
 * Base Map Layer for Default Interactions
 * 
 * This layer works in hex coordinate space and handles default tile/unit clicks
 * when no other interactive layer consumes the event. Acts as the fallback
 * for basic map interactions.
 */

import * as Phaser from 'phaser';
import { BaseLayer, LayerConfig, ClickContext, LayerHitResult } from '../LayerSystem';

/**
 * Base map layer that handles default tile and unit interactions
 */
export class BaseMapLayer extends BaseLayer {
    constructor(scene: Phaser.Scene) {
        super(scene, {
            name: 'base-map',
            coordinateSpace: 'hex',
            interactive: true,
            depth: 0, // Lowest priority - only handles events no other layer wants
        });
    }
    
    public hitTest(context: ClickContext): LayerHitResult | null {
        // Base map layer always consumes events that reach it
        // This ensures there's always a fallback handler
        return LayerHitResult.CONSUME;
    }
}
