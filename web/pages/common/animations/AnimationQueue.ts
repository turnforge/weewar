/**
 * AnimationQueue ensures animations run sequentially using simple promise chaining.
 *
 * Usage:
 *   const queue = new AnimationQueue();
 *   queue.enqueue(() => gameScene.moveUnit(unit, path));
 *   queue.enqueue(() => gameScene.showExplosion(q, r), (result) => {
 *     console.log('Explosion done');
 *     return true; // continue with next animation
 *   });
 */
export class AnimationQueue {
    private currentPromise: Promise<boolean> = Promise.resolve(true);
    private debug: boolean = true; // Enable debug logging

    /**
     * Enqueue an animation to run after all currently queued animations complete.
     *
     * @param animator Async function that performs the animation
     * @param callback Optional callback called when animator completes.
     *                 Return false to stop processing remaining animations.
     */
    enqueue<T>(
        animator: () => Promise<T>,
        callback?: (result: T) => boolean | void
    ): void {
        if (this.debug) {
            console.log('[AnimationQueue] Enqueuing animation');
        }

        this.currentPromise = this.currentPromise.then(async (shouldContinue) => {
            if (!shouldContinue) {
                if (this.debug) {
                    console.log('[AnimationQueue] Skipping animation (stopped)');
                }
                return false;
            }

            try {
                if (this.debug) {
                    console.log('[AnimationQueue] Running animation');
                }
                const result = await animator();
                if (this.debug) {
                    console.log('[AnimationQueue] Animation completed');
                }
                if (callback) {
                    const continueProcessing = callback(result);
                    return continueProcessing !== false;
                }
                return true;
            } catch (error) {
                console.error('[AnimationQueue] Animation failed:', error);
                return true; // Continue despite errors
            }
        });
    }

    /**
     * Wait for all queued animations to complete.
     * Note: This waits for animations enqueued before AND during this call.
     * If new animations are enqueued while waiting, it will wait for those too.
     */
    async waitForCompletion(): Promise<void> {
        // Keep waiting until currentPromise settles AND no new promises were added
        let lastPromise: Promise<boolean>;
        do {
            lastPromise = this.currentPromise;
            await lastPromise;
        } while (lastPromise !== this.currentPromise);
    }

    /**
     * Clear the queue by resetting the promise chain.
     */
    clear(): void {
        this.currentPromise = Promise.resolve(true);
    }
}

// Global animation queue instance
let globalAnimationQueue: AnimationQueue | null = null;

export function getAnimationQueue(): AnimationQueue {
    if (!globalAnimationQueue) {
        globalAnimationQueue = new AnimationQueue();
    }
    return globalAnimationQueue;
}

export function resetAnimationQueue(): void {
    if (globalAnimationQueue) {
        globalAnimationQueue.clear();
    }
    globalAnimationQueue = new AnimationQueue();
}
