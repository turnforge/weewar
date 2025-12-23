/**
 * GameSyncManager handles real-time synchronization of game state for multiplayer.
 *
 * Architecture:
 * - Subscribes to GameSyncService for game updates via streaming
 * - When MovesPublished updates arrive from other players, calls presenter's ApplyRemoteChanges
 * - Handles reconnection with sequence tracking
 *
 * Usage:
 * 1. Create manager with presenter client reference
 * 2. Call connect() to start subscription
 * 3. Manager automatically forwards remote changes to presenter
 * 4. Call disconnect() when leaving game
 */

import WeewarBundle from '../../gen/wasmjs';
import { GameSyncServiceClient } from '../../gen/wasmjs/weewar/v1/services/gameSyncServiceClient';
import { GameViewPresenterClient } from '../../gen/wasmjs/weewar/v1/services/gameViewPresenterClient';
import { GameUpdate, SubscribeRequest } from '../../gen/wasmjs/weewar/v1/models/interfaces';

export type SyncState = 'disconnected' | 'connecting' | 'connected' | 'reconnecting' | 'error';

export interface GameSyncManagerOptions {
    /** Callback when connection state changes */
    onStateChange?: (state: SyncState, error?: string) => void;
    /** Callback when a remote update is received */
    onRemoteUpdate?: (update: GameUpdate) => void;
    /** Auto-reconnect on disconnect (default: true) */
    autoReconnect?: boolean;
    /** Reconnection delay in ms (default: 2000) */
    reconnectDelayMs?: number;
}

export class GameSyncManager {
    private syncClient: GameSyncServiceClient;
    private presenterClient: GameViewPresenterClient;
    private gameId: string;
    private playerId: string;
    private lastSequence: number = 0;
    private state: SyncState = 'disconnected';
    private options: Required<GameSyncManagerOptions>;
    private reconnectTimeoutId: ReturnType<typeof setTimeout> | null = null;
    private isSubscribing: boolean = false;

    constructor(
        wasmBundle: WeewarBundle,
        presenterClient: GameViewPresenterClient,
        gameId: string,
        playerId: string,
        options: GameSyncManagerOptions = {}
    ) {
        this.syncClient = new GameSyncServiceClient(wasmBundle);
        this.presenterClient = presenterClient;
        this.gameId = gameId;
        this.playerId = playerId;
        this.options = {
            onStateChange: options.onStateChange || (() => {}),
            onRemoteUpdate: options.onRemoteUpdate || (() => {}),
            autoReconnect: options.autoReconnect ?? true,
            reconnectDelayMs: options.reconnectDelayMs ?? 2000,
        };
    }

    /**
     * Get current connection state
     */
    getState(): SyncState {
        return this.state;
    }

    /**
     * Get last received sequence number (for debugging/stats)
     */
    getLastSequence(): number {
        return this.lastSequence;
    }

    /**
     * Connect and start receiving game updates
     */
    connect(): void {
        if (this.state === 'connected' || this.isSubscribing) {
            console.log('[GameSyncManager] Already connected or connecting');
            return;
        }

        this.setState('connecting');
        this.subscribe();
    }

    /**
     * Disconnect from game updates
     */
    disconnect(): void {
        console.log('[GameSyncManager] Disconnecting');
        this.clearReconnectTimeout();
        this.isSubscribing = false;
        this.setState('disconnected');
    }

    private subscribe(): void {
        if (this.isSubscribing) {
            return;
        }

        this.isSubscribing = true;
        console.log(`[GameSyncManager] Subscribing to game ${this.gameId} from sequence ${this.lastSequence}`);

        const request: SubscribeRequest = {
            gameId: this.gameId,
            playerId: this.playerId,
            fromSequence: this.lastSequence,
        };

        this.syncClient.subscribe(request, (update, error, done) => {
            return this.handleStreamMessage(update, error, done);
        });
    }

    /**
     * Handle streaming message callback
     * @returns false to stop the stream, true to continue
     */
    private handleStreamMessage(
        update: GameUpdate | null,
        error: string | null,
        done: boolean
    ): boolean {
        // Handle errors
        if (error) {
            console.error('[GameSyncManager] Stream error:', error);
            this.isSubscribing = false;
            this.setState('error', error);
            this.scheduleReconnect();
            return false; // Stop stream
        }

        // Handle stream completion
        if (done) {
            console.log('[GameSyncManager] Stream completed');
            this.isSubscribing = false;
            if (this.state !== 'disconnected') {
                this.scheduleReconnect();
            }
            return false; // Stream is done
        }

        // Handle update
        if (update) {
            // Update to connected state on first message
            if (this.state !== 'connected') {
                this.setState('connected');
            }

            this.handleUpdate(update);
        }

        return true; // Continue receiving
    }

    private async handleUpdate(update: GameUpdate): Promise<void> {
        // Track sequence for reconnection
        if (update.sequence > this.lastSequence) {
            this.lastSequence = update.sequence;
        }

        console.log(`[GameSyncManager] Received update seq=${update.sequence}`, update);

        // Notify callback
        this.options.onRemoteUpdate(update);

        // Handle MovesPublished - apply to local presenter
        if (update.movesPublished) {
            const movesPublished = update.movesPublished;

            // Skip our own moves (we already applied them locally)
            if (movesPublished.player.toString() === this.playerId) {
                console.log('[GameSyncManager] Skipping own moves');
                return;
            }

            console.log(`[GameSyncManager] Applying remote moves from player ${movesPublished.player}`);

            try {
                const response = await this.presenterClient.applyRemoteChanges({
                    gameId: this.gameId,
                    moves: movesPublished.moves,
                });

                if (!response.success) {
                    console.error('[GameSyncManager] Failed to apply remote changes:', response.error);
                    if (response.requiresReload) {
                        console.warn('[GameSyncManager] State desync detected - reload required');
                        // Notify the page that a reload is needed
                        this.setState('error', 'State desync - reload required');
                    }
                }
            } catch (err) {
                console.error('[GameSyncManager] Error applying remote changes:', err);
            }
        }

        // Handle PlayerJoined
        if (update.playerJoined) {
            console.log(`[GameSyncManager] Player ${update.playerJoined.playerId} joined`);
        }

        // Handle PlayerLeft
        if (update.playerLeft) {
            console.log(`[GameSyncManager] Player ${update.playerLeft.playerId} left`);
        }

        // Handle GameEnded
        if (update.gameEnded) {
            console.log(`[GameSyncManager] Game ended: winner=${update.gameEnded.winner}, reason=${update.gameEnded.reason}`);
        }
    }

    private setState(state: SyncState, error?: string): void {
        if (this.state !== state) {
            console.log(`[GameSyncManager] State: ${this.state} -> ${state}`);
            this.state = state;
            this.options.onStateChange(state, error);
        }
    }

    private scheduleReconnect(): void {
        if (!this.options.autoReconnect || this.state === 'disconnected') {
            return;
        }

        this.clearReconnectTimeout();
        this.setState('reconnecting');

        console.log(`[GameSyncManager] Reconnecting in ${this.options.reconnectDelayMs}ms`);
        this.reconnectTimeoutId = setTimeout(() => {
            this.reconnectTimeoutId = null;
            this.subscribe();
        }, this.options.reconnectDelayMs);
    }

    private clearReconnectTimeout(): void {
        if (this.reconnectTimeoutId !== null) {
            clearTimeout(this.reconnectTimeoutId);
            this.reconnectTimeoutId = null;
        }
    }
}
