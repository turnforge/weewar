/**
 * GameSyncManager handles real-time synchronization of game state for multiplayer.
 *
 * Architecture:
 * - Subscribes to GameSyncService via HTTP/Connect streaming (direct to server)
 * - When MovesPublished updates arrive from other players, calls WASM presenter's ApplyRemoteChanges
 * - Handles reconnection with sequence tracking
 *
 * Usage:
 * 1. Create manager with presenter client reference
 * 2. Call connect() to start subscription
 * 3. Manager automatically forwards remote changes to presenter
 * 4. Call disconnect() when leaving game
 */

import { GameViewPresenterClient } from '../../gen/wasmjs/lilbattle/v1/services/gameViewPresenterClient';
import { GameUpdate } from '../../gen/wasmjs/lilbattle/v1/models/interfaces';

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
    /** Base URL for the sync service (default: current origin) */
    baseUrl?: string;
}

export class GameSyncManager {
    private presenterClient: GameViewPresenterClient;
    private gameId: string;
    private lastSequence: number = 0;
    private state: SyncState = 'disconnected';
    private options: Required<GameSyncManagerOptions>;
    private reconnectTimeoutId: ReturnType<typeof setTimeout> | null = null;
    private abortController: AbortController | null = null;

    constructor(
        presenterClient: GameViewPresenterClient,
        gameId: string,
        options: GameSyncManagerOptions = {}
    ) {
        this.presenterClient = presenterClient;
        this.gameId = gameId;
        this.options = {
            onStateChange: options.onStateChange || (() => {}),
            onRemoteUpdate: options.onRemoteUpdate || (() => {}),
            autoReconnect: options.autoReconnect ?? true,
            reconnectDelayMs: options.reconnectDelayMs ?? 2000,
            baseUrl: options.baseUrl || (window.location.origin + "/api"),
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
     * Connect and start receiving game updates via HTTP streaming
     */
    connect(): void {
        if (this.state === 'connected' || this.state === 'connecting') {
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
        if (this.abortController) {
            this.abortController.abort();
            this.abortController = null;
        }
        this.setState('disconnected');
    }

    /**
     * Subscribe to game updates via gRPC-gateway REST endpoint
     * Uses GET with query params, returns newline-delimited JSON stream
     */
    private async subscribe(): Promise<void> {
        // gRPC-gateway REST endpoint: /v1/games/{game_id}/sync/subscribe
        const params = new URLSearchParams({
            from_sequence: this.lastSequence.toString(),
        });
        const url = `${this.options.baseUrl}/v1/sync/games/${this.gameId}/subscribe?${params}`;

        console.log(`[GameSyncManager] Subscribing to ${url}`);

        // Create abort controller for this subscription
        this.abortController = new AbortController();

        try {
            const response = await fetch(url, {
                method: 'GET',
                headers: {
                    'Accept': 'application/json',
                },
                signal: this.abortController.signal,
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            if (!response.body) {
                throw new Error('Response body is null');
            }

            this.setState('connected');
            await this.readStream(response.body);

        } catch (error: any) {
            if (error.name === 'AbortError') {
                console.log('[GameSyncManager] Subscription aborted');
                return;
            }

            console.error('[GameSyncManager] Subscription error:', error);
            this.setState('error', error.message);
            this.scheduleReconnect();
        }
    }

    /**
     * Read and process the streaming response
     * Connect streaming uses newline-delimited JSON messages
     */
    private async readStream(body: ReadableStream<Uint8Array>): Promise<void> {
        const reader = body.getReader();
        const decoder = new TextDecoder();
        let buffer = '';

        try {
            while (true) {
                const { done, value } = await reader.read();

                if (done) {
                    console.log('[GameSyncManager] Stream ended');
                    break;
                }

                buffer += decoder.decode(value, { stream: true });

                // Process complete messages (Connect uses newline-delimited JSON or envelope format)
                const lines = buffer.split('\n');
                buffer = lines.pop() || ''; // Keep incomplete line in buffer

                for (const line of lines) {
                    if (line.trim()) {
                        await this.processMessage(line);
                    }
                }
            }
        } catch (error: any) {
            if (error.name !== 'AbortError') {
                console.error('[GameSyncManager] Stream read error:', error);
                this.setState('error', error.message);
            }
        } finally {
            reader.releaseLock();
        }

        // Stream ended - schedule reconnect if not intentionally disconnected
        if (this.state !== 'disconnected') {
            this.scheduleReconnect();
        }
    }

    /**
     * Process a single message from the stream
     */
    private async processMessage(data: string): Promise<void> {
        try {
            // Connect streaming wraps messages in an envelope
            const envelope = JSON.parse(data);

            // Handle Connect envelope format: { "result": { ... } } or { "error": { ... } }
            let update: GameUpdate;
            if (envelope.result) {
                update = envelope.result;
            } else if (envelope.error) {
                console.error('[GameSyncManager] Server error:', envelope.error);
                return;
            } else {
                // Direct message format
                update = envelope;
            }

            await this.handleUpdate(update);
        } catch (error) {
            console.error('[GameSyncManager] Failed to parse message:', data, error);
        }
    }

    private async handleUpdate(update: GameUpdate): Promise<void> {
        // Track sequence for reconnection
        if (update.sequence > this.lastSequence) {
            this.lastSequence = update.sequence;
        }

        console.log(`[GameSyncManager] Received update seq=${update.sequence}`, update);

        // Notify callback
        this.options.onRemoteUpdate(update);

        // Handle MovesPublished - apply to local WASM presenter
        // The presenter will decide whether to apply based on group number
        if (update.movesPublished) {
            const movesPublished = update.movesPublished;

            console.log(`[GameSyncManager] Received moves from player ${movesPublished.player}, group ${movesPublished.groupNumber}`);

            const response = await this.presenterClient.applyRemoteChanges({
                gameId: this.gameId,
                moves: movesPublished.moves,
            })

            if (!response.success) {
                console.error('[GameSyncManager] Failed to apply remote changes:', response.error);
                if (response.requiresReload) {
                    console.warn('[GameSyncManager] State desync detected - reload required');
                    this.setState('error', 'State desync - reload required');
                }
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
