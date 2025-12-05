import { LCMComponent, EventBus } from '@panyam/tsappkit';

/**
 * GameLogPanel - Manages game event logging and filtering
 * 
 * Features:
 * - Event logging with timestamps
 * - Log filtering by type (all, moves, combat, system)
 * - Log clearing functionality
 * - Entry count and last update tracking
 * - Template-based UI with proper event handling
 */
export class GameLogPanel implements LCMComponent {
    private eventBus: EventBus;
    private gameLog: string[] = [];

    constructor(readonly element: HTMLElement, eventBus: EventBus) {
        this.eventBus = eventBus;
    }

    // =============================================================================
    // LCMComponent Interface Implementation
    // =============================================================================

    performLocalInit(): LCMComponent[] {
        this.initializeEventHandlers();
        this.initializeGameLog();
        return [];
    }

    setupDependencies(): void {
        // No dependencies needed
    }

    async activate(): Promise<void> {
        // Component is ready to use
    }

    handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
        // Handle any game events that should be logged
        // This can be extended to automatically log game events
    }

    deactivate(): void {
        // Clean up any listeners if needed
        this.gameLog = [];
    }

    // =============================================================================
    // Public API
    // =============================================================================

    /**
     * Add a new entry to the game log
     */
    public logGameEvent(message: string, type: 'system' | 'moves' | 'combat' = 'system'): void {
        this.gameLog.push(message);
        
        // Add the entry to the UI
        this.addLogEntryToUI(message, type);
        
        // Update statistics
        this.updateLogStats();
        
        // Keep only last 50 entries
        if (this.gameLog.length > 50) {
            this.gameLog.shift();
            this.removeOldestLogEntry();
        }
    }

    /**
     * Clear all log entries
     */
    public clearLog(): void {
        this.gameLog = [];
        const logContainer = this.element.querySelector('#game-log');
        if (logContainer) {
            logContainer.innerHTML = '';
        }
        this.updateLogStats();
    }

    /**
     * Get current log entries
     */
    public getLogEntries(): string[] {
        return [...this.gameLog];
    }

    // =============================================================================
    // Private Implementation
    // =============================================================================

    /**
     * Initialize event handlers for the panel
     */
    private initializeEventHandlers(): void {
        // Set up log filtering buttons
        const filterBtns = this.element.querySelectorAll('.log-filter-btn');
        filterBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const target = e.target as HTMLElement;
                const filter = target.dataset.filter;
                this.handleLogFilter(filter || 'all');
            });
        });

        // Set up clear log button
        const clearBtn = this.element.querySelector('#clear-log-btn');
        if (clearBtn) {
            clearBtn.addEventListener('click', () => this.handleClearLog());
        }
    }

    /**
     * Initialize the game log with default entries
     */
    private initializeGameLog(): void {
        this.gameLog = [];
        // The template already has some default entries, so we'll keep those
        this.updateLogStats();
    }

    /**
     * Add a log entry to the UI
     */
    private addLogEntryToUI(message: string, type: string): void {
        const logContainer = this.element.querySelector('#game-log');
        const template = document.getElementById('log-entry-template') as HTMLTemplateElement;
        
        if (logContainer && template) {
            // Clone the template
            const entryElement = template.content.cloneNode(true) as DocumentFragment;
            const logEntry = entryElement.querySelector('.log-entry') as HTMLElement;
            
            if (logEntry) {
                // Set up the entry
                const timestamp = new Date().toLocaleTimeString();
                logEntry.setAttribute('data-timestamp', timestamp);
                logEntry.setAttribute('data-type', type);
                logEntry.classList.add(`${type}-log`);
                
                // Set content
                const timestampSpan = logEntry.querySelector('.timestamp');
                const messageSpan = logEntry.querySelector('.message');
                
                if (timestampSpan) {
                    timestampSpan.textContent = `[${timestamp}]`;
                }
                
                if (messageSpan) {
                    messageSpan.textContent = message;
                }
                
                // Add to container
                logContainer.appendChild(entryElement);
                
                // Scroll to bottom
                logContainer.scrollTop = logContainer.scrollHeight;
            }
        }
    }

    /**
     * Remove the oldest log entry from UI
     */
    private removeOldestLogEntry(): void {
        const logContainer = this.element.querySelector('#game-log');
        if (logContainer && logContainer.children.length > 50) {
            logContainer.removeChild(logContainer.firstElementChild!);
        }
    }

    /**
     * Handle log filtering
     */
    private handleLogFilter(filter: string): void {
        // Update active filter button
        const filterBtns = this.element.querySelectorAll('.log-filter-btn');
        filterBtns.forEach(btn => {
            btn.classList.toggle('active', btn.getAttribute('data-filter') === filter);
        });

        // Show/hide log entries based on filter
        const logEntries = this.element.querySelectorAll('.log-entry');
        logEntries.forEach(entry => {
            const entryType = entry.getAttribute('data-type') || 'system';
            const shouldShow = filter === 'all' || entryType === filter;
            entry.classList.toggle('hidden', !shouldShow);
        });
    }

    /**
     * Handle clearing the game log
     */
    private handleClearLog(): void {
        this.clearLog();
    }

    /**
     * Update log statistics display
     */
    private updateLogStats(): void {
        const entryCountEl = this.element.querySelector('#log-entry-count');
        const lastUpdateEl = this.element.querySelector('#log-last-update');
        
        if (entryCountEl) {
            entryCountEl.textContent = `${this.gameLog.length} entries`;
        }
        
        if (lastUpdateEl) {
            lastUpdateEl.textContent = 'Just now';
        }
    }
}
