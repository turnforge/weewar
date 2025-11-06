import { ThemeManager } from '../lib/ThemeManager';
import { SplashScreen } from '../lib/SplashScreen';

/**
 * Manages the world listing page logic
 */
class WorldListingPage {
    constructor() {
        ThemeManager.init();
        this.init();
    }

    /**
     * Initialize page
     */
    private init(): void {
        // Page-specific initialization
    }
}

// Dismiss splash screen immediately when script loads (defer ensures DOM is ready)
SplashScreen.dismiss();

// Initialize the WorldListingPage when the DOM is fully loaded
document.addEventListener('DOMContentLoaded', () => {
    (window as any).Page = new WorldListingPage();
});
