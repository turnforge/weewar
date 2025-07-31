import { LCMComponent } from '../lib/LCMComponent';
import { BaseComponent } from '../lib/Component';
import { EventBus } from '../lib/EventBus';
import { EditorEventTypes, ReferenceLoadFromFilePayload, ReferenceSetModePayload, ReferenceSetAlphaPayload, ReferenceSetPositionPayload, ReferenceSetScalePayload, ReferenceScaleChangedPayload, ReferenceStateChangedPayload, ReferenceImageLoadedPayload } from './events';

/**
 * ReferenceImagePanel - Demonstrates new lifecycle architecture
 * 
 * This component showcases the breadth-first lifecycle pattern:
 * 1. performLocalInit() - Set up UI controls without dependencies
 * 2. setupDependencies() - Receive PhaserEditorComponent when available
 * 3. activate() - Enable functionality once all dependencies ready
 * 
 * Benefits:
 * - No initialization order dependencies
 * - Graceful handling of missing dependencies
 * - Clear separation of concerns across lifecycle phases
 */
export class ReferenceImagePanel extends BaseComponent {
    // Dependencies (injected in phase 2)
    private toastCallback?: (title: string, message: string, type: 'success' | 'error' | 'info') => void;
    
    // Internal state
    private isUIBound = false;
    private isActivated = false;
    private pendingOperations: Array<() => void> = [];
    
    // Reference image state cache (updated via EventBus)
    private referenceState = {
        scale: { x: 1.0, y: 1.0 },
        position: { x: 0, y: 0 },
        alpha: 0.5,
        mode: 0,
        isLoaded: false
    };
    
    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('reference-image-panel', rootElement, eventBus, debugMode);
    }
    
    // Dependencies are now set directly using explicit setters instead of ComponentDependencyDeclaration
    
    // LCMComponent Phase 1: Initialize DOM and discover children (no dependencies needed)
    public performLocalInit(): LCMComponent[] {
        if (this.isUIBound) {
            this.log('Already bound to DOM, skipping');
            return [];
        }
        
        this.log('Binding ReferenceImagePanel to DOM');
        
        // Set up UI elements and event handlers
        this.bindLoadingControls();
        this.bindDisplayModeControls();
        this.bindAlphaControls();
        this.bindPositionControls();
        this.bindScaleControls();
        this.bindPositionTranslationControls();
        this.bindClearControls();
        
        this.isUIBound = true;
        this.log('ReferenceImagePanel bound to DOM successfully');
        
        // This is a leaf component - no children
        return [];
    }
    
    // Phase 2: Inject dependencies - simplified to use explicit setters
    public setupDependencies(): void {
        this.log('ReferenceImagePanel: Dependencies injection phase - using explicit setters');
        
        // Dependencies should be set directly by parent using setters
        // This phase just validates that required dependencies are available
        if (!this.toastCallback) {
            throw new Error('ReferenceImagePanel requires toast callback - use setToastCallback()');
        }
        
        this.log('Dependencies validation complete');
    }
    
    // Explicit dependency setters
    public setToastCallback(callback: (title: string, message: string, type: 'success' | 'error' | 'info') => void): void {
        this.toastCallback = callback;
        this.log('Toast callback set via explicit setter');
    }
    
    // Explicit dependency getters
    public getToastCallback(): ((title: string, message: string, type: 'success' | 'error' | 'info') => void) | undefined {
        return this.toastCallback;
    }
    
    // Phase 3: Activate component
    public activate(): void {
        if (this.isActivated) {
            this.log('Already activated, skipping');
            return;
        }
        
        this.log('Activating ReferenceImagePanel');
        
        // Subscribe to EventBus events from PhaserEditorComponent
        this.subscribeToReferenceEvents();
        
        // Process any operations that were queued during UI binding
        this.processPendingOperations();
        
        // Update UI state - no longer dependent on PhaserEditorComponent availability
        this.updateUIState();
        
        this.isActivated = true;
        this.log('ReferenceImagePanel activated successfully');
    }
    
    /**
     * Subscribe to reference image events from PhaserEditorComponent
     */
    private subscribeToReferenceEvents(): void {
        // Subscribe to scale changes from direct Phaser interaction
        this.addSubscription(EditorEventTypes.REFERENCE_SCALE_CHANGED, this);
        
        // Subscribe to state changes from direct Phaser interaction
        this.addSubscription(EditorEventTypes.REFERENCE_STATE_CHANGED, this);
        
        this.log('Subscribed to reference image EventBus events');
    }

    /**
     * Handle events from the EventBus
     */
    public handleBusEvent(eventType: string, data: any, target: any, emitter: any): void {
        switch(eventType) {
            case EditorEventTypes.REFERENCE_SCALE_CHANGED:
                this.handleReferenceScaleChanged(data);
                break;
            
            case EditorEventTypes.REFERENCE_STATE_CHANGED:
                this.handleReferenceStateChanged(data);
                break;
            
            default:
                // Call parent implementation for unhandled events
                super.handleBusEvent(eventType, data, target, emitter);
        }
    }
    
    /**
     * Handle reference scale changed event from PhaserEditorComponent
     */
    private handleReferenceScaleChanged(data: ReferenceScaleChangedPayload): void {
        this.log(`Received reference scale changed: ${data.scaleX}, ${data.scaleY}`);
        
        // Update local state cache
        this.referenceState.scale.x = data.scaleX;
        this.referenceState.scale.y = data.scaleY;
        
        // Update UI display
        this.updateReferenceScaleDisplay();
    }
    
    /**
     * Handle reference state changed event from PhaserEditorComponent
     */
    private handleReferenceStateChanged(data: ReferenceStateChangedPayload): void {
        this.log(`Received reference state changed:`, data);
        
        // Update local state cache safely, preserving structure
        if (data.scale) {
            this.referenceState.scale = { ...data.scale };
        }
        if (data.position) {
            this.referenceState.position = { ...data.position };
        }
        if (typeof data.alpha === 'number') {
            this.referenceState.alpha = data.alpha;
        }
        if (typeof data.mode === 'number') {
            this.referenceState.mode = data.mode;
        }
        if (typeof data.isLoaded === 'boolean') {
            this.referenceState.isLoaded = data.isLoaded;
        }
        
        // Update UI based on new state
        this.updateReferenceScaleDisplay();
        this.updateReferencePositionDisplay();
        this.updateReferenceStatus(data.isLoaded ? 
            (data.mode === 0 ? 'Hidden' : ['Hidden', 'Background', 'Overlay'][data.mode] + ' mode') : 
            'No reference image loaded'
        );
    }
    
    // Phase 4: Deactivate component
    public deactivate(): void {
        this.log('Deactivating ReferenceImagePanel');
        
        // Clear any pending operations
        this.pendingOperations = [];
        
        // Reset state
        this.isActivated = false;
        this.toastCallback = undefined;
        
        this.log('ReferenceImagePanel deactivated');
    }
    
    // UI Binding Methods (Phase 1)
    
    private bindLoadingControls(): void {
        // Load from clipboard button
        const loadReferenceBtn = this.rootElement.querySelector('#load-reference-btn') as HTMLButtonElement;
        if (loadReferenceBtn) {
            loadReferenceBtn.addEventListener('click', () => {
                this.executeWhenReady(() => this.loadReferenceFromClipboard());
            });
            this.log('Load from clipboard button bound');
        }
        
        // File input and load from file button
        const fileInput = this.rootElement.querySelector('#reference-file-input') as HTMLInputElement;
        const loadFileBtn = this.rootElement.querySelector('#load-reference-file-btn') as HTMLButtonElement;
        
        if (loadFileBtn && fileInput) {
            loadFileBtn.addEventListener('click', () => {
                fileInput.click();
            });
            
            fileInput.addEventListener('change', (e) => {
                const target = e.target as HTMLInputElement;
                if (target.files && target.files.length > 0) {
                    this.executeWhenReady(() => this.loadReferenceFromFile(target.files![0]));
                }
            });
            
            this.log('File loading controls bound');
        }
    }
    
    private bindDisplayModeControls(): void {
        // Bind radio button controls for reference mode
        const modeRadios = this.rootElement.querySelectorAll('input[name="reference-mode"]') as NodeListOf<HTMLInputElement>;
        modeRadios.forEach(radio => {
            radio.addEventListener('change', (e) => {
                if ((e.target as HTMLInputElement).checked) {
                    const mode = parseInt((e.target as HTMLInputElement).value);
                    this.executeWhenReady(() => this.setReferenceMode(mode));
                }
            });
            
            // Also bind click events to the label/button div for better UX
            const label = radio.closest('label');
            if (label) {
                label.addEventListener('click', () => {
                    if (!radio.checked) {
                        radio.checked = true;
                        const mode = parseInt(radio.value);
                        this.executeWhenReady(() => this.setReferenceMode(mode));
                    }
                });
            }
        });
        this.log('Display mode radio buttons bound');
    }
    
    private bindAlphaControls(): void {
        const alphaSlider = this.rootElement.querySelector('#reference-alpha') as HTMLInputElement;
        const alphaValue = this.rootElement.querySelector('#reference-alpha-value') as HTMLElement;
        
        if (alphaSlider && alphaValue) {
            alphaSlider.addEventListener('input', (e) => {
                const alpha = parseInt((e.target as HTMLInputElement).value) / 100;
                alphaValue.textContent = `${Math.round(alpha * 100)}%`;
                this.executeWhenReady(() => this.setReferenceAlpha(alpha));
            });
            this.log('Alpha transparency slider bound');
        }
    }
    
    private bindPositionControls(): void {
        const resetPositionBtn = this.rootElement.querySelector('#reference-reset-position') as HTMLButtonElement;
        if (resetPositionBtn) {
            resetPositionBtn.addEventListener('click', () => {
                this.executeWhenReady(() => this.resetReferencePosition());
            });
        }
        
        const resetScaleBtn = this.rootElement.querySelector('#reference-reset-scale') as HTMLButtonElement;
        if (resetScaleBtn) {
            resetScaleBtn.addEventListener('click', () => {
                this.executeWhenReady(() => this.resetReferenceScale());
            });
        }
        
        this.log('Position controls bound');
    }
    
    private bindScaleControls(): void {
        // X Scale controls
        const scaleXMinusBtn = this.rootElement.querySelector('#reference-scale-x-minus') as HTMLButtonElement;
        const scaleXPlusBtn = this.rootElement.querySelector('#reference-scale-x-plus') as HTMLButtonElement;
        const scaleXInput = this.rootElement.querySelector('#reference-scale-x-value') as HTMLInputElement;
        
        if (scaleXMinusBtn) {
            scaleXMinusBtn.addEventListener('click', () => {
                this.executeWhenReady(() => this.adjustReferenceScaleX(-0.01));
            });
        }
        
        if (scaleXPlusBtn) {
            scaleXPlusBtn.addEventListener('click', () => {
                this.executeWhenReady(() => this.adjustReferenceScaleX(0.01));
            });
        }
        
        if (scaleXInput) {
            scaleXInput.addEventListener('change', () => {
                const value = parseFloat(scaleXInput.value);
                if (!isNaN(value)) {
                    this.executeWhenReady(() => this.setReferenceScaleX(value));
                }
            });
        }
        
        // Y Scale controls
        const scaleYMinusBtn = this.rootElement.querySelector('#reference-scale-y-minus') as HTMLButtonElement;
        const scaleYPlusBtn = this.rootElement.querySelector('#reference-scale-y-plus') as HTMLButtonElement;
        const scaleYInput = this.rootElement.querySelector('#reference-scale-y-value') as HTMLInputElement;
        
        if (scaleYMinusBtn) {
            scaleYMinusBtn.addEventListener('click', () => {
                this.executeWhenReady(() => this.adjustReferenceScaleY(-0.01));
            });
        }
        
        if (scaleYPlusBtn) {
            scaleYPlusBtn.addEventListener('click', () => {
                this.executeWhenReady(() => this.adjustReferenceScaleY(0.01));
            });
        }
        
        if (scaleYInput) {
            scaleYInput.addEventListener('change', () => {
                const value = parseFloat(scaleYInput.value);
                if (!isNaN(value)) {
                    this.executeWhenReady(() => this.setReferenceScaleY(value));
                }
            });
        }
        
        this.log('Scale controls bound');
    }
    
    private bindPositionTranslationControls(): void {
        // X Position controls
        const positionXMinusBtn = this.rootElement.querySelector('#reference-position-x-minus') as HTMLButtonElement;
        const positionXPlusBtn = this.rootElement.querySelector('#reference-position-x-plus') as HTMLButtonElement;
        const positionXInput = this.rootElement.querySelector('#reference-position-x-value') as HTMLInputElement;
        
        if (positionXMinusBtn) {
            positionXMinusBtn.addEventListener('click', () => {
                this.executeWhenReady(() => this.adjustReferencePositionX(-1)); // 1 pixel increments
            });
        }
        
        if (positionXPlusBtn) {
            positionXPlusBtn.addEventListener('click', () => {
                this.executeWhenReady(() => this.adjustReferencePositionX(1));
            });
        }
        
        if (positionXInput) {
            positionXInput.addEventListener('change', () => {
                const value = parseFloat(positionXInput.value);
                if (!isNaN(value)) {
                    this.executeWhenReady(() => this.setReferencePositionX(value));
                }
            });
        }
        
        // Y Position controls
        const positionYMinusBtn = this.rootElement.querySelector('#reference-position-y-minus') as HTMLButtonElement;
        const positionYPlusBtn = this.rootElement.querySelector('#reference-position-y-plus') as HTMLButtonElement;
        const positionYInput = this.rootElement.querySelector('#reference-position-y-value') as HTMLInputElement;
        
        if (positionYMinusBtn) {
            positionYMinusBtn.addEventListener('click', () => {
                this.executeWhenReady(() => this.adjustReferencePositionY(-1));
            });
        }
        
        if (positionYPlusBtn) {
            positionYPlusBtn.addEventListener('click', () => {
                this.executeWhenReady(() => this.adjustReferencePositionY(1));
            });
        }
        
        if (positionYInput) {
            positionYInput.addEventListener('change', () => {
                const value = parseFloat(positionYInput.value);
                if (!isNaN(value)) {
                    this.executeWhenReady(() => this.setReferencePositionY(value));
                }
            });
        }
        
        this.log('Position translation controls bound');
    }
    
    private bindClearControls(): void {
        const clearReferenceBtn = this.rootElement.querySelector('#clear-reference-btn') as HTMLButtonElement;
        if (clearReferenceBtn) {
            clearReferenceBtn.addEventListener('click', () => {
                this.executeWhenReady(() => this.clearReferenceImage());
            });
            this.log('Clear reference button bound');
        }
    }
    
    // Deferred Execution System
    
    /**
     * Execute operation when component is ready, or queue it for later
     */
    private executeWhenReady(operation: () => void): void {
        if (this.isActivated) {
            // Component is ready - execute immediately
            operation();
        } else {
            // Component not ready - queue for later
            this.pendingOperations.push(operation);
            this.showToast('Info', 'Component not ready - operation queued', 'info');
        }
    }
    
    /**
     * Process all pending operations when component becomes ready
     */
    private processPendingOperations(): void {
        if (this.pendingOperations.length > 0) {
            this.log(`Processing ${this.pendingOperations.length} pending operations`);
            
            const operations = [...this.pendingOperations];
            this.pendingOperations = [];
            
            operations.forEach(operation => {
                operation();
            });
        }
    }
    
    /**
     * Update UI state - no longer dependent on PhaserEditor availability
     */
    private updateUIState(): void {
        // Enable all controls - communication via EventBus means no direct dependency needed
        const controls = this.rootElement.querySelectorAll('button, input, select');
        
        controls.forEach(control => {
            const element = control as HTMLElement;
            element.removeAttribute('disabled');
            element.classList.remove('opacity-50', 'cursor-not-allowed');
        });
        
        // Update status message
        this.updateReferenceStatus('No reference image loaded');
    }
    
    // Reference Image Operations (Phase 3 - when dependencies are available)
    
    private async loadReferenceFromClipboard(): Promise<void> {
          this.log('Loading reference image directly from clipboard');
          this.showToast('Loading', 'Loading reference image from clipboard...', 'info');
          this.updateReferenceStatus('Loading from clipboard...');
          
          // Check if clipboard API is available
          if (!navigator.clipboard || !navigator.clipboard.read) {
              throw new Error('Clipboard API not supported in this browser');
          }
          
          // Read from clipboard
          const clipboardItems = await navigator.clipboard.read();
          let imageFound = false;
          
          for (const clipboardItem of clipboardItems) {
              for (const type of clipboardItem.types) {
                  if (type.startsWith('image/')) {
                      const blob = await clipboardItem.getType(type);
                      await this.loadReferenceFromBlob(blob, 'clipboard');
                      imageFound = true;
                      break;
                  }
              }
              if (imageFound) break;
          }
          
          if (!imageFound) {
              throw new Error('No image found in clipboard');
          }
    }
    
    private async loadReferenceFromFile(file: File): Promise<void> {
        this.log(`Loading reference image directly from file: ${file.name} (${file.size} bytes)`);
        this.showToast('Loading', `Loading reference image from ${file.name}...`, 'info');
        this.updateReferenceStatus(`Loading from ${file.name}...`);
        
        // Validate file type
        if (!file.type.startsWith('image/')) {
            throw new Error('Selected file is not an image');
        }
        
        // Load image directly
        await this.loadReferenceFromBlob(file, file.name);
    }
    
    /**
     * Common method to load reference image from blob/file
     */
    private async loadReferenceFromBlob(blob: Blob, source: string): Promise<void> {
        // Create object URL for the blob
        const imageUrl = URL.createObjectURL(blob);
        
        // Create image element to validate and get dimensions
        const img = new Image();
        
        await new Promise<void>((resolve, reject) => {
            img.onload = () => resolve();
            img.onerror = () => reject(new Error('Failed to load image'));
            img.src = imageUrl;
        });
        
        this.log(`Reference image loaded successfully: ${img.width}x${img.height} from ${source}`);
        
        // Update UI state first
        this.referenceState.isLoaded = true;
        this.updateReferenceStatus(`Loaded from ${source} (${img.width}x${img.height})`);
        this.showToast('Success', `Reference image loaded from ${source}`, 'success');
        this.updateUIState();
        
        // Set default mode to background if currently hidden - BEFORE emitting event
        this.setDefaultMode();
        
        // Notify EventBus that image was loaded (notification, not request)
        // This happens AFTER setDefaultMode so Phaser gets the image with correct mode
        this.eventBus.emit<ReferenceImageLoadedPayload>(
            EditorEventTypes.REFERENCE_IMAGE_LOADED, 
            {
                source: source,
                width: img.width,
                height: img.height,
                url: imageUrl
            }, 
            this,
            this
        );
        
        // Clean up the object URL after a delay to allow other components to use it
        setTimeout(() => {
            URL.revokeObjectURL(imageUrl);
        }, 1000);
    }
    
    private setDefaultMode(): void {
        // Enable mode selector and default to background mode
        const modeSelect = this.rootElement.querySelector('#reference-mode') as HTMLSelectElement;
        if (modeSelect && modeSelect.value === '0') {
            modeSelect.value = '1'; // Default to background mode
            this.setReferenceMode(1);
        }
    }
    
    private setReferenceMode(mode: number): void {
        // Emit event to PhaserEditorComponent via EventBus
        this.eventBus.emit<ReferenceSetModePayload>(
            EditorEventTypes.REFERENCE_SET_MODE, 
            { mode }, 
            this,
            this
        );
        
        // Update UI radio buttons to reflect current mode
        const modeRadios = this.rootElement.querySelectorAll('input[name="reference-mode"]') as NodeListOf<HTMLInputElement>;
        modeRadios.forEach(radio => {
            const isSelected = radio.value === mode.toString();
            radio.checked = isSelected;
            
            // Update the visual styling of the button
            const buttonDiv = radio.nextElementSibling as HTMLElement;
            if (buttonDiv) {
                if (isSelected) {
                    // Selected state - blue background with white text
                    buttonDiv.className = 'text-xs px-2 py-1 text-center rounded border border-gray-300 dark:border-gray-600 transition-colors duration-200 bg-blue-500 text-white font-medium';
                } else {
                    // Unselected state - gray background
                    buttonDiv.className = 'text-xs px-2 py-1 text-center rounded border border-gray-300 dark:border-gray-600 transition-colors duration-200 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600';
                }
            }
        });
        
        // Show/hide controls based on mode
        const positionControls = this.rootElement.querySelector('#reference-position-controls') as HTMLElement;
        const resetPositionBtn = this.rootElement.querySelector('#reference-reset-position') as HTMLElement;
        
        if (positionControls) {
            if (mode === 0) {
                // Hidden mode - hide all controls
                positionControls.style.display = 'none';
            } else if (mode === 1) {
                // Background mode - show scale controls but hide position controls
                positionControls.style.display = 'block';
                if (resetPositionBtn) {
                    resetPositionBtn.style.display = 'none';
                }
            } else if (mode === 2) {
                // Overlay mode - show all controls
                positionControls.style.display = 'block';
                if (resetPositionBtn) {
                    resetPositionBtn.style.display = 'block';
                }
            }
        }
        
        // Note: Scale display will be updated when PhaserEditorComponent responds via EventBus
        
        const modeNames = ['Hidden', 'Background', 'Overlay'];
        this.log(`Reference mode set to: ${modeNames[mode]}`);
        this.updateReferenceStatus(mode === 0 ? 'Hidden' : `${modeNames[mode]} mode`);
    }
    
    private setReferenceAlpha(alpha: number): void {
        // Emit event to PhaserEditorComponent via EventBus
        this.eventBus.emit<ReferenceSetAlphaPayload>(
            EditorEventTypes.REFERENCE_SET_ALPHA, 
            { alpha }, 
            this,
            this
        );
        this.log(`Reference alpha set to: ${Math.round(alpha * 100)}%`);
    }
    
    private resetReferencePosition(): void {
        // Emit event to PhaserEditorComponent via EventBus
        this.eventBus.emit<ReferenceSetPositionPayload>(
            EditorEventTypes.REFERENCE_SET_POSITION, 
            { x: 0, y: 0 }, 
            this,
            this
        );
        this.log('Reference position reset to center');
        this.showToast('Position Reset', 'Reference image centered', 'success');
    }
    
    private resetReferenceScale(): void {
        // Emit event to PhaserEditorComponent via EventBus
        this.eventBus.emit<ReferenceSetScalePayload>(
            EditorEventTypes.REFERENCE_SET_SCALE, 
            { scaleX: 1, scaleY: 1 }, 
            this,
            this
        );
        this.log('Reference scale reset to 100%');
        this.showToast('Scale Reset', 'Reference image scale reset', 'success');
        // Scale display will be updated when PhaserEditorComponent responds via EventBus
    }
    
    private adjustReferenceScaleX(delta: number): void {
        const currentScaleX = this.referenceState?.scale?.x ?? 1.0;
        const currentScaleY = this.referenceState?.scale?.y ?? 1.0;
        const newScaleX = Math.max(0.1, Math.min(5.0, currentScaleX + delta));
        
        // Emit event to PhaserEditorComponent via EventBus
        this.eventBus.emit<ReferenceSetScalePayload>(
            EditorEventTypes.REFERENCE_SET_SCALE, 
            { scaleX: newScaleX, scaleY: currentScaleY }, 
            this,
            this
        );
        
        // Update local state cache
        if (this.referenceState?.scale) {
            this.referenceState.scale.x = newScaleX;
        }
        this.updateReferenceScaleDisplay();
        this.log(`Reference X scale: ${newScaleX.toFixed(2)}`);
    }
    
    private adjustReferenceScaleY(delta: number): void {
        const currentScaleX = this.referenceState?.scale?.x ?? 1.0;
        const currentScaleY = this.referenceState?.scale?.y ?? 1.0;
        const newScaleY = Math.max(0.1, Math.min(5.0, currentScaleY + delta));
        
        // Emit event to PhaserEditorComponent via EventBus
        this.eventBus.emit<ReferenceSetScalePayload>(
            EditorEventTypes.REFERENCE_SET_SCALE, 
            { scaleX: currentScaleX, scaleY: newScaleY }, 
            this,
            this
        );
        
        // Update local state cache
        if (this.referenceState?.scale) {
            this.referenceState.scale.y = newScaleY;
        }
        this.updateReferenceScaleDisplay();
        this.log(`Reference Y scale: ${newScaleY.toFixed(2)}`);
    }
    
    private setReferenceScaleX(scaleX: number): void {
        const clampedScale = Math.max(0.1, Math.min(5.0, scaleX));
        const currentScaleY = this.referenceState?.scale?.y ?? 1.0;
        
        // Emit event to PhaserEditorComponent via EventBus
        this.eventBus.emit<ReferenceSetScalePayload>(
            EditorEventTypes.REFERENCE_SET_SCALE, 
            { scaleX: clampedScale, scaleY: currentScaleY }, 
            this,
            this
        );
        
        // Update local state cache
        if (this.referenceState?.scale) {
            this.referenceState.scale.x = clampedScale;
        }
        this.updateReferenceScaleDisplay();
        this.log(`Reference X scale: ${clampedScale.toFixed(2)}`);
    }
    
    private setReferenceScaleY(scaleY: number): void {
        const clampedScale = Math.max(0.1, Math.min(5.0, scaleY));
        const currentScaleX = this.referenceState?.scale?.x ?? 1.0;
        
        // Emit event to PhaserEditorComponent via EventBus
        this.eventBus.emit<ReferenceSetScalePayload>(
            EditorEventTypes.REFERENCE_SET_SCALE, 
            { scaleX: currentScaleX, scaleY: clampedScale }, 
            this,
            this
        );
        
        // Update local state cache
        if (this.referenceState?.scale) {
            this.referenceState.scale.y = clampedScale;
        }
        this.updateReferenceScaleDisplay();
        this.log(`Reference Y scale: ${clampedScale.toFixed(2)}`);
    }
    
    private updateReferenceScaleDisplay(): void {
        const scaleXInput = this.rootElement.querySelector('#reference-scale-x-value') as HTMLInputElement;
        const scaleYInput = this.rootElement.querySelector('#reference-scale-y-value') as HTMLInputElement;
        
        // Defensive programming: ensure scale values exist
        const scaleX = this.referenceState?.scale?.x ?? 1.0;
        const scaleY = this.referenceState?.scale?.y ?? 1.0;
        
        if (scaleXInput) {
            scaleXInput.value = scaleX.toFixed(2);
        }
        
        if (scaleYInput) {
            scaleYInput.value = scaleY.toFixed(2);
        }
    }
    
    private adjustReferencePositionX(delta: number): void {
        const currentPositionX = this.referenceState?.position?.x ?? 0;
        const currentPositionY = this.referenceState?.position?.y ?? 0;
        const newPositionX = currentPositionX + delta;
        
        // Emit event to PhaserEditorComponent via EventBus
        this.eventBus.emit<ReferenceSetPositionPayload>(
            EditorEventTypes.REFERENCE_SET_POSITION, 
            { x: newPositionX, y: currentPositionY }, 
            this, this
        );
        
        // Update local state cache
        if (this.referenceState?.position) {
            this.referenceState.position.x = newPositionX;
        }
        this.updateReferencePositionDisplay();
        this.log(`Reference X position: ${newPositionX}`);
    }
    
    private adjustReferencePositionY(delta: number): void {
        const currentPositionX = this.referenceState?.position?.x ?? 0;
        const currentPositionY = this.referenceState?.position?.y ?? 0;
        const newPositionY = currentPositionY + delta;
        
        // Emit event to PhaserEditorComponent via EventBus
        this.eventBus.emit<ReferenceSetPositionPayload>(
            EditorEventTypes.REFERENCE_SET_POSITION, 
            { x: currentPositionX, y: newPositionY }, 
            this, this
        );
        
        // Update local state cache
        if (this.referenceState?.position) {
            this.referenceState.position.y = newPositionY;
        }
        this.updateReferencePositionDisplay();
        this.log(`Reference Y position: ${newPositionY}`);
    }
    
    private setReferencePositionX(positionX: number): void {
        const currentPositionY = this.referenceState?.position?.y ?? 0;
        
        // Emit event to PhaserEditorComponent via EventBus
        this.eventBus.emit<ReferenceSetPositionPayload>(
            EditorEventTypes.REFERENCE_SET_POSITION, 
            { x: positionX, y: currentPositionY }, 
            this, this
        );
        
        // Update local state cache
        if (this.referenceState?.position) {
            this.referenceState.position.x = positionX;
        }
        this.updateReferencePositionDisplay();
        this.log(`Reference X position: ${positionX}`);
    }
    
    private setReferencePositionY(positionY: number): void {
        const currentPositionX = this.referenceState?.position?.x ?? 0;
        
        // Emit event to PhaserEditorComponent via EventBus
        this.eventBus.emit<ReferenceSetPositionPayload>(
            EditorEventTypes.REFERENCE_SET_POSITION, 
            { x: currentPositionX, y: positionY }, 
            this, this
        );
        
        // Update local state cache
        if (this.referenceState?.position) {
            this.referenceState.position.y = positionY;
        }
        this.updateReferencePositionDisplay();
        this.log(`Reference Y position: ${positionY}`);
    }
    
    private updateReferencePositionDisplay(): void {
        const positionXInput = this.rootElement.querySelector('#reference-position-x-value') as HTMLInputElement;
        const positionYInput = this.rootElement.querySelector('#reference-position-y-value') as HTMLInputElement;
        
        // Defensive programming: ensure position values exist
        const positionX = this.referenceState?.position?.x ?? 0;
        const positionY = this.referenceState?.position?.y ?? 0;
        
        if (positionXInput) {
            positionXInput.value = positionX.toString();
        }
        
        if (positionYInput) {
            positionYInput.value = positionY.toString();
        }
    }
    
    private clearReferenceImage(): void {
        // Emit event to PhaserEditorComponent via EventBus
        this.eventBus.emit(EditorEventTypes.REFERENCE_CLEAR, {}, this, this);
        
        // Reset local state cache
        this.referenceState = {
            scale: { x: 1.0, y: 1.0 },
            position: { x: 0, y: 0 },
            alpha: 0.5,
            mode: 0,
            isLoaded: false
        };
        
        // Reset UI controls - set mode to Hidden (0) and update radio button styling
        this.setReferenceMode(0);
        
        const alphaSlider = this.rootElement.querySelector('#reference-alpha') as HTMLInputElement;
        const alphaValue = this.rootElement.querySelector('#reference-alpha-value') as HTMLElement;
        if (alphaSlider && alphaValue) {
            alphaSlider.value = '50';
            alphaValue.textContent = '50%';
        }
        
        // Hide position controls and reset position button visibility
        const positionControls = this.rootElement.querySelector('#reference-position-controls') as HTMLElement;
        const resetPositionBtn = this.rootElement.querySelector('#reference-reset-position') as HTMLElement;
        if (positionControls) {
            positionControls.style.display = 'none';
        }
        if (resetPositionBtn) {
            resetPositionBtn.style.display = 'block'; // Reset to default state
        }
        
        // Reset position inputs
        const positionXInput = this.rootElement.querySelector('#reference-position-x-value') as HTMLInputElement;
        const positionYInput = this.rootElement.querySelector('#reference-position-y-value') as HTMLInputElement;
        if (positionXInput) {
            positionXInput.value = '0';
        }
        if (positionYInput) {
            positionYInput.value = '0';
        }
        
        this.updateReferenceStatus('No reference image loaded');
        this.log('Reference image cleared');
        this.showToast('Cleared', 'Reference image removed', 'success');
    }
    
    private updateReferenceStatus(status: string): void {
        const statusElement = this.rootElement.querySelector('#reference-status') as HTMLElement;
        if (statusElement) {
            statusElement.textContent = status;
        }
    }
    
    // Helper method to show toast notifications
    private showToast(title: string, message: string, type: 'success' | 'error' | 'info' = 'info'): void {
        if (this.toastCallback) {
            this.toastCallback(title, message, type);
        } else {
            this.log(`Toast: ${title} - ${message} (${type})`);
        }
    }
    
    protected destroyComponent(): void {
        this.deactivate();
    }
}
