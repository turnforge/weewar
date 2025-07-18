/**
 * Generic keyboard shortcut manager for handling multi-key commands
 * across all application pages
 * 
 * Current Behavior (default):
 * - User types 'c3' → visual indicator → Enter/timeout → execute
 * - Requires explicit confirmation for execution
 * 
 * Future Immediate Mode (when enabled):
 * - User types 'c3' → immediate preview → Escape to cancel
 * - Provides instant feedback with option to cancel
 */

export interface ShortcutConfig {
    key: string;
    handler: (args?: string) => void;
    description: string;
    category?: string;
    requiresArgs?: boolean;
    argType?: 'number' | 'string';
    contextFilter?: (event: KeyboardEvent) => boolean;
    
    // Future: Preview handlers for immediate execution mode
    previewHandler?: (args?: string) => void;  // Called immediately as user types
    cancelHandler?: () => void;               // Called when user presses Escape
    executeImmediately?: boolean;             // Override global immediateExecution setting
}

export interface ShortcutManagerConfig {
    shortcuts: ShortcutConfig[];
    helpContainer?: string;
    timeout?: number; // ms to return to normal state (current: 3000ms, immediate mode: 300ms)
    immediateExecution?: boolean; // Future: enable immediate execution with preview
}

export enum KeyboardState {
    NORMAL = 'normal',
    AWAITING_ARGS = 'awaiting_args'
}

export class KeyboardShortcutManager {
    private shortcuts: Map<string, ShortcutConfig> = new Map();
    private state: KeyboardState = KeyboardState.NORMAL;
    private currentCommand: string = '';
    private currentArgs: string = '';
    private helpContainer: string | null = null;
    private timeout: number = 3000; // Default 3 second timeout
    private timeoutId: number | null = null;
    private helpOverlay: HTMLElement | null = null;
    private immediateExecution: boolean = false; // Future: enable immediate execution mode

    constructor(config: ShortcutManagerConfig) {
        this.helpContainer = config.helpContainer || null;
        this.timeout = config.timeout || 3000;
        this.immediateExecution = config.immediateExecution || false;
        
        // Register shortcuts
        config.shortcuts.forEach(shortcut => {
            this.shortcuts.set(shortcut.key, shortcut);
        });
        
        this.initialize();
    }

    private initialize(): void {
        // Global keydown listener
        document.addEventListener('keydown', this.handleKeydown.bind(this));
        
        // Show initial state indicator
        this.updateStateIndicator();
    }

    private handleKeydown(event: KeyboardEvent): void {
        const target = event.target as HTMLElement;
        
        // Skip if in input field, textarea, or contenteditable
        if (this.isInInputContext(target)) {
            return;
        }
        
        // Handle help key
        if (event.key === '?' && this.state === KeyboardState.NORMAL) {
            event.preventDefault();
            this.showHelp();
            return;
        }
        
        // Handle escape key
        if (event.key === 'Escape') {
            event.preventDefault();
            this.resetState();
            return;
        }
        
        // Handle state machine
        if (this.state === KeyboardState.NORMAL) {
            this.handleNormalState(event);
        } else if (this.state === KeyboardState.AWAITING_ARGS) {
            this.handleAwaitingArgsState(event);
        }
    }

    private handleNormalState(event: KeyboardEvent): void {
        const key = event.key.toLowerCase();
        const shortcut = this.shortcuts.get(key);
        
        if (shortcut) {
            event.preventDefault();
            
            // Check context filter
            if (shortcut.contextFilter && !shortcut.contextFilter(event)) {
                return;
            }
            
            if (shortcut.requiresArgs) {
                // Enter args waiting state
                this.state = KeyboardState.AWAITING_ARGS;
                this.currentCommand = key;
                this.currentArgs = '';
                this.updateStateIndicator();
                this.startTimeout();
            } else {
                // Execute immediately
                this.executeShortcut(shortcut);
            }
        }
    }

    private handleAwaitingArgsState(event: KeyboardEvent): void {
        const key = event.key;
        
        if (key >= '0' && key <= '9') {
            // Add digit to args
            event.preventDefault();
            this.currentArgs += key;
            this.updateStateIndicator();
            this.resetTimeout();
        } else if (key === 'Enter' || key === ' ') {
            // Execute command with args
            event.preventDefault();
            this.executeCurrentCommand();
        } else if (key === 'Backspace') {
            // Remove last digit
            event.preventDefault();
            this.currentArgs = this.currentArgs.slice(0, -1);
            this.updateStateIndicator();
            this.resetTimeout();
        }
    }

    private executeCurrentCommand(): void {
        const shortcut = this.shortcuts.get(this.currentCommand);
        if (shortcut && this.currentArgs) {
            this.executeShortcut(shortcut, this.currentArgs);
        }
        this.resetState();
    }

    private executeShortcut(shortcut: ShortcutConfig, args?: string): void {
        try {
            shortcut.handler(args);
        } catch (error) {
            console.error('Error executing shortcut:', error);
        }
    }

    private isInInputContext(element: HTMLElement): boolean {
        const tagName = element.tagName.toLowerCase();
        return (
            tagName === 'input' ||
            tagName === 'textarea' ||
            tagName === 'select' ||
            element.contentEditable === 'true' ||
            element.closest('.modal') !== null ||
            element.closest('[contenteditable="true"]') !== null
        );
    }

    private resetState(): void {
        this.state = KeyboardState.NORMAL;
        this.currentCommand = '';
        this.currentArgs = '';
        this.clearTimeout();
        this.updateStateIndicator();
        this.hideHelp();
    }

    private startTimeout(): void {
        this.timeoutId = window.setTimeout(() => {
            this.resetState();
        }, this.timeout);
    }

    private resetTimeout(): void {
        this.clearTimeout();
        this.startTimeout();
    }

    private clearTimeout(): void {
        if (this.timeoutId) {
            window.clearTimeout(this.timeoutId);
            this.timeoutId = null;
        }
    }

    private updateStateIndicator(): void {
        // Remove existing indicator
        const existingIndicator = document.getElementById('keyboard-state-indicator');
        if (existingIndicator) {
            existingIndicator.remove();
        }

        // Only show indicator when not in normal state
        if (this.state === KeyboardState.NORMAL) {
            return;
        }

        // Create new indicator
        const indicator = document.createElement('div');
        indicator.id = 'keyboard-state-indicator';
        indicator.className = 'fixed top-4 right-4 bg-blue-600 text-white px-3 py-2 rounded-lg shadow-lg z-50 font-mono text-sm';
        
        if (this.state === KeyboardState.AWAITING_ARGS) {
            const shortcut = this.shortcuts.get(this.currentCommand);
            const description = shortcut ? shortcut.description : 'Unknown command';
            indicator.innerHTML = `
                <div class="flex items-center space-x-2">
                    <span>${this.currentCommand.toUpperCase()}</span>
                    <span class="text-blue-200">${this.currentArgs || '_'}</span>
                </div>
                <div class="text-xs text-blue-200 mt-1">${description}</div>
            `;
        }

        document.body.appendChild(indicator);
    }

    private showHelp(): void {
        if (this.helpOverlay) {
            this.hideHelp();
            return;
        }

        this.helpOverlay = document.createElement('div');
        this.helpOverlay.id = 'keyboard-help-overlay';
        this.helpOverlay.className = 'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50';
        
        const helpContent = document.createElement('div');
        helpContent.className = 'bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-2xl max-h-[80vh] overflow-y-auto p-6';
        
        helpContent.innerHTML = this.generateHelpContent();
        
        this.helpOverlay.appendChild(helpContent);
        document.body.appendChild(this.helpOverlay);
        
        // Close on click outside or escape
        this.helpOverlay.addEventListener('click', (e) => {
            if (e.target === this.helpOverlay) {
                this.hideHelp();
            }
        });
    }

    private hideHelp(): void {
        if (this.helpOverlay) {
            this.helpOverlay.remove();
            this.helpOverlay = null;
        }
    }

    private generateHelpContent(): string {
        const categories = new Map<string, ShortcutConfig[]>();
        
        // Group shortcuts by category
        this.shortcuts.forEach(shortcut => {
            const category = shortcut.category || 'General';
            if (!categories.has(category)) {
                categories.set(category, []);
            }
            categories.get(category)!.push(shortcut);
        });

        let html = `
            <div class="flex items-center justify-between mb-4">
                <h2 class="text-xl font-bold text-gray-900 dark:text-white">Keyboard Shortcuts</h2>
                <button class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200" onclick="this.closest('#keyboard-help-overlay').remove()">
                    <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                    </svg>
                </button>
            </div>
        `;

        categories.forEach((shortcuts, category) => {
            html += `
                <div class="mb-6">
                    <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-200 mb-3">${category}</h3>
                    <div class="space-y-2">
            `;

            shortcuts.forEach(shortcut => {
                const keyDisplay = shortcut.requiresArgs 
                    ? `${shortcut.key.toUpperCase()}<span class="text-blue-600 dark:text-blue-400">&lt;number&gt;</span>`
                    : shortcut.key.toUpperCase();
                
                html += `
                    <div class="flex items-center justify-between py-2 px-3 bg-gray-50 dark:bg-gray-700 rounded">
                        <span class="text-sm text-gray-700 dark:text-gray-300">${shortcut.description}</span>
                        <kbd class="px-2 py-1 text-xs font-mono bg-gray-200 dark:bg-gray-600 text-gray-800 dark:text-gray-200 rounded">${keyDisplay}</kbd>
                    </div>
                `;
            });

            html += `
                    </div>
                </div>
            `;
        });

        html += `
            <div class="mt-6 pt-4 border-t border-gray-200 dark:border-gray-600">
                <p class="text-sm text-gray-600 dark:text-gray-400 text-center">
                    Press <kbd class="px-2 py-1 text-xs font-mono bg-gray-200 dark:bg-gray-600 rounded">ESC</kbd> to cancel any command or 
                    <kbd class="px-2 py-1 text-xs font-mono bg-gray-200 dark:bg-gray-600 rounded">?</kbd> to close this help
                </p>
            </div>
        `;

        return html;
    }

    public destroy(): void {
        document.removeEventListener('keydown', this.handleKeydown.bind(this));
        this.resetState();
        this.hideHelp();
    }

    public getState(): KeyboardState {
        return this.state;
    }

    public getCurrentCommand(): string {
        return this.currentCommand;
    }

    public getCurrentArgs(): string {
        return this.currentArgs;
    }
}