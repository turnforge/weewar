/**
 * Common DOM utility functions for consistent behavior across components
 */

/**
 * Checks if the user is currently typing in an input field, textarea, or other editable element.
 * This is used to prevent keyboard shortcuts from interfering with user input.
 * 
 * @param element - The target element from a keyboard event
 * @returns true if the user is in an input context, false otherwise
 */
export function isInInputContext(element: HTMLElement | null): boolean {
    if (!element) return false;
    
    const tagName = element.tagName.toLowerCase();
    
    // Direct input elements
    if (tagName === 'input' || tagName === 'textarea' || tagName === 'select') {
        return true;
    }
    
    // Contenteditable elements
    if (element.contentEditable === 'true' || element.isContentEditable) {
        return true;
    }
    
    // Check if element is inside input-related containers
    if (element.closest('input') !== null ||
        element.closest('textarea') !== null ||
        element.closest('select') !== null ||
        element.closest('[contenteditable="true"]') !== null ||
        element.closest('[contenteditable=""]') !== null) {
        return true;
    }
    
    // Check if element is inside a modal (modals often contain forms)
    if (element.closest('.modal') !== null) {
        return true;
    }
    
    // Check for specific input field IDs that might be problematic
    const inputFieldIds = ['map-title-input'];
    if (inputFieldIds.includes(element.id) || 
        inputFieldIds.some(id => element.closest(`#${id}`) !== null)) {
        return true;
    }
    
    return false;
}

/**
 * Checks if modifier keys are pressed (Ctrl, Alt, Cmd, Shift).
 * This is commonly used to filter keyboard shortcuts.
 * 
 * @param event - The keyboard event
 * @returns true if any modifier keys are pressed, false otherwise
 */
export function hasModifierKeys(event: KeyboardEvent): boolean {
    return event.ctrlKey || event.altKey || event.metaKey || event.shiftKey;
}

/**
 * Combined check for whether keyboard shortcuts should be ignored.
 * This checks both modifier keys and input context.
 * 
 * @param event - The keyboard event
 * @returns true if shortcuts should be ignored, false if they can be processed
 */
export function shouldIgnoreShortcut(event: KeyboardEvent): boolean {
    const target = event.target as HTMLElement;
    return hasModifierKeys(event) || isInInputContext(target);
}