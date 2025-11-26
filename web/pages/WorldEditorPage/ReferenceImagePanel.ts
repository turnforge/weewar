import { LCMComponent } from '../../lib/LCMComponent';
import { BaseComponent } from '../../lib/Component';
import { EventBus } from '../../lib/EventBus';
import { ReferenceImageDB } from './ReferenceImageDB';
import { ReferenceImageLayer } from './ReferenceImageLayer';
import { IWorldEditorPresenter } from './WorldEditorPresenter';

/**
 * ReferenceImagePanel - Manages reference image loading and controls
 *
 * This component handles ALL reference image loading operations:
 * - Loading from IndexedDB storage (auto-restore on page load)
 * - Loading from file uploads
 * - Loading from clipboard
 *
 * It manages:
 * - IndexedDB persistence (per-world storage)
 * - localStorage for position/scale offsets
 * - Direct control of ReferenceImageLayer for display
 *
 * The ReferenceImageLayer only handles:
 * - Display and rendering
 * - Mouse/touch interaction (drag, scroll)
 *
 * Lifecycle:
 * 1. performLocalInit() - Set up UI controls without dependencies
 * 2. setupDependencies() - Receive dependencies (layer, worldId, toast)
 * 3. activate() - Enable functionality and auto-restore from storage
 */
export class ReferenceImagePanel extends BaseComponent {
    // Dependencies (injected in phase 2)
    private toastCallback?: (title: string, message: string, type: 'success' | 'error' | 'info') => void;
    private referenceImageLayer?: ReferenceImageLayer;
    private worldId?: string;
    private presenter: IWorldEditorPresenter | null = null;

    // Internal state
    private isUIBound = false;
    private isActivated = false;
    private pendingOperations: Array<() => void> = [];

    // IndexedDB storage for reference images
    private imageDB: ReferenceImageDB = new ReferenceImageDB();

    // Reference image state cache (updated via EventBus)
    private referenceState = {
        scale: { x: 1.0, y: 1.0 },
        position: { x: 0, y: 0 },
        alpha: 0.5,
        mode: 0,
        isLoaded: false
    };

    // localStorage keys
    private static readonly STORAGE_KEY_POSITION = 'referenceImagePosition';
    private static readonly STORAGE_KEY_SCALE = 'referenceImageScale';

    // Debounced save
    private savePositionDebounceTimer: number | null = null;
    private saveScaleDebounceTimer: number | null = null;
    private static readonly SAVE_DEBOUNCE_MS = 300; // 300ms debounce

    constructor(rootElement: HTMLElement, eventBus: EventBus, debugMode: boolean = false) {
        super('reference-image-panel', rootElement, eventBus, debugMode);
    }
    
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

    // Explicit dependency setters
    public setToastCallback(callback: (title: string, message: string, type: 'success' | 'error' | 'info') => void): void {
        this.toastCallback = callback;
        this.log('Toast callback set via explicit setter');
    }

    public setReferenceImageLayer(layer: ReferenceImageLayer): void {
        this.referenceImageLayer = layer;
        this.log('ReferenceImageLayer set via explicit setter');
    }

    public setWorldId(worldId: string): void {
        this.worldId = worldId;

        // Restore from storage when worldId is set
        this.loadReferenceFromStorage();
        this.log(`WorldId set via explicit setter: ${worldId}`);
    }

    public setPresenter(presenter: IWorldEditorPresenter): void {
        this.presenter = presenter;
        this.log('Presenter set via explicit setter');
    }

    // Explicit dependency getters
    public getToastCallback(): ((title: string, message: string, type: 'success' | 'error' | 'info') => void) | undefined {
        return this.toastCallback;
    }

    public getReferenceImageLayer(): ReferenceImageLayer | undefined {
        return this.referenceImageLayer;
    }

    public getWorldId(): string | undefined {
        return this.worldId;
    }

    // Public methods called by presenter when scale/position change from scene drag/scroll
    public updateScaleDisplay(scaleX: number, scaleY: number): void {
        // Only update if values actually changed (prevent circular updates)
        const changed = this.referenceState.scale.x !== scaleX ||
                       this.referenceState.scale.y !== scaleY;
        if (!changed) return;

        this.referenceState.scale.x = scaleX;
        this.referenceState.scale.y = scaleY;
        this.updateReferenceScaleDisplay();
        this.saveScaleToLocalStorage();
        this.log(`Scale display updated: ${scaleX}, ${scaleY}`);
    }

    public updatePositionDisplay(x: number, y: number): void {
        // Only update if values actually changed (prevent circular updates)
        const changed = this.referenceState.position.x !== x ||
                       this.referenceState.position.y !== y;
        if (!changed) return;

        this.referenceState.position.x = x;
        this.referenceState.position.y = y;
        this.updateReferencePositionDisplay();
        this.savePositionToLocalStorage();
        this.log(`Position display updated: ${x}, ${y}`);
    }

    // Phase 3: Activate component
    public async activate(): Promise<void> {
        if (this.isActivated) {
            this.log('Already activated, skipping');
            return;
        }

        this.log('Activating ReferenceImagePanel');

        // Initialize IndexedDB
        await this.imageDB.init();

        // Load saved position and scale from localStorage
        this.loadSavedPositionAndScale();

        // Process any operations that were queued during UI binding
        this.processPendingOperations();

        // Update UI state - no longer dependent on PhaserEditorComponent availability
        this.updateUIState();

        this.isActivated = true;
        this.log('ReferenceImagePanel activated successfully');
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

    /**
     * Load reference image from IndexedDB storage
     * Called automatically during activate() if worldId is set
     */
    private async loadReferenceFromStorage(): Promise<boolean> {
        if (!this.worldId) {
            this.log('Cannot load from storage: worldId not set');
            return false;
        }

        this.log(`Loading reference image from storage for world ${this.worldId}`);

        try {
            const record = await this.imageDB.getImageRecord(this.worldId);
            if (record) {
                this.log(`Found reference image in storage: ${record.filename || 'unnamed'}`);
                await this.loadReferenceFromBlob(record.imageBlob, 'storage');
                return true;
            } else {
                this.log('No reference image found in storage');
                return false;
            }
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
            this.log(`Failed to load from storage: ${errorMessage}`);
            return false;
        }
    }

    private async loadReferenceFromClipboard(): Promise<void> {
          this.log('Loading reference image directly from clipboard');
          this.showToast('Loading', 'Loading reference image from clipboard...', 'info');
          this.updateReferenceStatus('Loading from clipboard...');

          try {
              // Check if clipboard API is available
              if (!navigator.clipboard || !navigator.clipboard.read) {
                  throw new Error('Clipboard API not supported in this browser. Try using the file upload option instead.');
              }

              // Request clipboard-read permission
              // Note: This may trigger a browser permission prompt
              const permission = await navigator.permissions.query({ name: 'clipboard-read' as PermissionName });

              if (permission.state === 'denied') {
                  throw new Error('Clipboard access denied. Please grant clipboard permissions in your browser settings.');
              }

              // Read from clipboard (this may also trigger permission prompt if not already granted)
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
                  throw new Error('No image found in clipboard. Copy an image and try again.');
              }
          } catch (error) {
              const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
              this.log(`Failed to load from clipboard: ${errorMessage}`);
              this.showToast('Error', errorMessage, 'error');
              this.updateReferenceStatus('Failed to load from clipboard');
              throw error;
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
     * Handles display via layer and persistence via IndexedDB
     */
    private async loadReferenceFromBlob(blob: Blob, source: string): Promise<void> {
        if (!this.referenceImageLayer) {
            throw new Error('ReferenceImageLayer not available');
        }

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

        // Set the image on the layer for display
        this.referenceImageLayer.setReferenceImage(imageUrl);

        // Set default mode to background (always use background when loading/restoring)
        this.referenceImageLayer.setReferenceMode(1); // 1 = background
        this.referenceState.mode = 1;
        this.setReferenceMode(1); // Update UI controls

        // Apply saved position and scale values from localStorage
        this.applySavedPositionAndScale();

        // Save to IndexedDB if not loading from storage
        if (source !== 'storage' && this.worldId) {
            const filename = source === 'clipboard' ? 'clipboard-image' : source;
            await this.imageDB.saveImage(this.worldId, blob, filename);
            this.log(`Saved reference image to IndexedDB for world ${this.worldId}`);
        }

        // Clean up the object URL after a delay to allow layer to use it
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
        // Notify presenter to update Phaser scene
        this.presenter?.setReferenceMode(mode);

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
        // Notify presenter to update Phaser scene
        this.presenter?.setReferenceAlpha(alpha);
        this.log(`Reference alpha set to: ${Math.round(alpha * 100)}%`);
    }
    
    private resetReferencePosition(): void {
        // Notify presenter to update Phaser scene
        this.presenter?.setReferencePosition(0, 0);
        this.referenceState.position = { x: 0, y: 0 };
        this.updateReferencePositionDisplay();
        this.log('Reference position reset to center');
        this.showToast('Position Reset', 'Reference image centered', 'success');
    }
    
    private resetReferenceScale(): void {
        // Notify presenter to update Phaser scene
        this.presenter?.setReferenceScale(1, 1);
        this.referenceState.scale = { x: 1, y: 1 };
        this.updateReferenceScaleDisplay();
        this.log('Reference scale reset to 100%');
        this.showToast('Scale Reset', 'Reference image scale reset', 'success');
    }
    
    private adjustReferenceScaleX(delta: number): void {
        const currentScaleX = this.referenceState?.scale?.x ?? 1.0;
        const currentScaleY = this.referenceState?.scale?.y ?? 1.0;
        const newScaleX = Math.max(0.1, Math.min(5.0, currentScaleX + delta));

        // Notify presenter to update Phaser scene
        this.presenter?.setReferenceScale(newScaleX, currentScaleY);

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

        // Notify presenter to update Phaser scene
        this.presenter?.setReferenceScale(currentScaleX, newScaleY);

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

        // Notify presenter to update Phaser scene
        this.presenter?.setReferenceScale(clampedScale, currentScaleY);

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

        // Notify presenter to update Phaser scene
        this.presenter?.setReferenceScale(currentScaleX, clampedScale);

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
            scaleXInput.value = scaleX.toString();
        }

        if (scaleYInput) {
            scaleYInput.value = scaleY.toString();
        }
    }
    
    private adjustReferencePositionX(delta: number): void {
        const currentPositionX = this.referenceState?.position?.x ?? 0;
        const currentPositionY = this.referenceState?.position?.y ?? 0;
        const newPositionX = currentPositionX + delta;

        // Notify presenter to update Phaser scene
        this.presenter?.setReferencePosition(newPositionX, currentPositionY);

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

        // Notify presenter to update Phaser scene
        this.presenter?.setReferencePosition(currentPositionX, newPositionY);

        // Update local state cache
        if (this.referenceState?.position) {
            this.referenceState.position.y = newPositionY;
        }
        this.updateReferencePositionDisplay();
        this.log(`Reference Y position: ${newPositionY}`);
    }

    private setReferencePositionX(positionX: number): void {
        const currentPositionY = this.referenceState?.position?.y ?? 0;

        // Notify presenter to update Phaser scene
        this.presenter?.setReferencePosition(positionX, currentPositionY);

        // Update local state cache
        if (this.referenceState?.position) {
            this.referenceState.position.x = positionX;
        }
        this.updateReferencePositionDisplay();
        this.log(`Reference X position: ${positionX}`);
    }

    private setReferencePositionY(positionY: number): void {
        const currentPositionX = this.referenceState?.position?.x ?? 0;

        // Notify presenter to update Phaser scene
        this.presenter?.setReferencePosition(currentPositionX, positionY);

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

    /**
     * Load saved position and scale from localStorage and populate input fields
     */
    private loadSavedPositionAndScale(): void {
        try {
            // Load position
            const savedPosition = localStorage.getItem(ReferenceImagePanel.STORAGE_KEY_POSITION);
            if (savedPosition) {
                const position = JSON.parse(savedPosition);
                this.referenceState.position.x = position.x ?? 0;
                this.referenceState.position.y = position.y ?? 0;
                this.updateReferencePositionDisplay();
                this.log(`Loaded saved position: ${position.x}, ${position.y}`);
            }

            // Load scale
            const savedScale = localStorage.getItem(ReferenceImagePanel.STORAGE_KEY_SCALE);
            if (savedScale) {
                const scale = JSON.parse(savedScale);
                this.referenceState.scale.x = scale.x ?? 1.0;
                this.referenceState.scale.y = scale.y ?? 1.0;
                this.updateReferenceScaleDisplay();
                this.log(`Loaded saved scale: ${scale.x}, ${scale.y}`);
            }
        } catch (error) {
            this.log(`Failed to load saved position/scale: ${error}`);
        }
    }

    /**
     * Apply saved position and scale to the reference image layer
     * Called when a new reference image is loaded
     */
    private applySavedPositionAndScale(): void {
        if (!this.referenceImageLayer) {
            this.log('ReferenceImageLayer not available, cannot apply position/scale');
            return;
        }

        // Apply position directly to layer
        this.referenceImageLayer.setReferencePosition(
            this.referenceState.position.x,
            this.referenceState.position.y
        );

        // Apply scale directly to layer
        this.referenceImageLayer.setReferenceScale(
            this.referenceState.scale.x,
            this.referenceState.scale.y
        );

        this.log(`Applied saved position (${this.referenceState.position.x}, ${this.referenceState.position.y}) and scale (${this.referenceState.scale.x}, ${this.referenceState.scale.y})`);
    }

    /**
     * Save position to localStorage with debouncing
     */
    private savePositionToLocalStorage(): void {
        // Clear existing timer
        if (this.savePositionDebounceTimer !== null) {
            window.clearTimeout(this.savePositionDebounceTimer);
        }

        // Set new timer
        this.savePositionDebounceTimer = window.setTimeout(() => {
            try {
                const position = {
                    x: this.referenceState.position.x,
                    y: this.referenceState.position.y
                };
                localStorage.setItem(ReferenceImagePanel.STORAGE_KEY_POSITION, JSON.stringify(position));
                this.log(`Saved position to localStorage: ${position.x}, ${position.y}`);
            } catch (error) {
                this.log(`Failed to save position: ${error}`);
            }
            this.savePositionDebounceTimer = null;
        }, ReferenceImagePanel.SAVE_DEBOUNCE_MS);
    }

    /**
     * Save scale to localStorage with debouncing
     */
    private saveScaleToLocalStorage(): void {
        // Clear existing timer
        if (this.saveScaleDebounceTimer !== null) {
            window.clearTimeout(this.saveScaleDebounceTimer);
        }

        // Set new timer
        this.saveScaleDebounceTimer = window.setTimeout(() => {
            try {
                const scale = {
                    x: this.referenceState.scale.x,
                    y: this.referenceState.scale.y
                };
                localStorage.setItem(ReferenceImagePanel.STORAGE_KEY_SCALE, JSON.stringify(scale));
                this.log(`Saved scale to localStorage: ${scale.x}, ${scale.y}`);
            } catch (error) {
                this.log(`Failed to save scale: ${error}`);
            }
            this.saveScaleDebounceTimer = null;
        }, ReferenceImagePanel.SAVE_DEBOUNCE_MS);
    }

    private async clearReferenceImage(): Promise<void> {
        if (!this.referenceImageLayer) {
            this.log('ReferenceImageLayer not available');
            return;
        }

        // Clear from layer (display)
        await this.referenceImageLayer.clearReferenceImage();

        // Clear from IndexedDB (storage)
        if (this.worldId) {
            await this.imageDB.deleteImage(this.worldId);
            this.log(`Cleared reference image from IndexedDB for world ${this.worldId}`);
        }

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
