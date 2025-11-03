/**
 * Utility for managing the splash screen that loads before JavaScript
 */
export class SplashScreen {
  private static readonly SPLASH_ID = 'splash-screen';
  private static dismissed = false;

  /**
   * Dismiss the splash screen with a fade-out animation
   * Safe to call multiple times - only dismisses once
   */
  static dismiss(): void {
    if (this.dismissed) {
      return;
    }

    const splash = document.getElementById(this.SPLASH_ID);
    if (!splash) {
      return;
    }

    this.dismissed = true;

    // Fade out
    splash.style.opacity = '0';

    // Remove from DOM after animation completes
    setTimeout(() => {
      splash.remove();
    }, 300); // Match transition-opacity duration-300 from Tailwind
  }

  /**
   * Update the splash screen message (if it hasn't been dismissed yet)
   */
  static updateMessage(title?: string, message?: string): void {
    if (this.dismissed) {
      return;
    }

    const splash = document.getElementById(this.SPLASH_ID);
    if (!splash) {
      return;
    }

    if (title) {
      const titleEl = splash.querySelector('[data-splash-title]');
      if (titleEl) {
        titleEl.textContent = title;
      }
    }

    if (message) {
      const messageEl = splash.querySelector('[data-splash-message]');
      if (messageEl) {
        messageEl.textContent = message;
      }
    }
  }

  /**
   * Update the splash screen progress bar
   * @param percent - Progress percentage (0-100)
   */
  static updateProgress(percent: number): void {
    if (this.dismissed) {
      return;
    }

    const splash = document.getElementById(this.SPLASH_ID);
    if (!splash) {
      return;
    }

    // Clamp between 0-100
    const clampedPercent = Math.max(0, Math.min(100, percent));

    const progressBar = splash.querySelector('[data-splash-progress-bar]') as HTMLElement;
    if (progressBar) {
      progressBar.style.width = `${clampedPercent}%`;
    }

    const progressText = splash.querySelector('[data-splash-progress-text]');
    if (progressText) {
      progressText.textContent = `${Math.round(clampedPercent)}%`;
    }
  }

  /**
   * Update both message and progress at once
   */
  static update(options: {
    title?: string;
    message?: string;
    progress?: number;
  }): void {
    if (options.title !== undefined || options.message !== undefined) {
      this.updateMessage(options.title, options.message);
    }
    if (options.progress !== undefined) {
      this.updateProgress(options.progress);
    }
  }

  /**
   * Check if splash screen is still visible
   */
  static isVisible(): boolean {
    return !this.dismissed && !!document.getElementById(this.SPLASH_ID);
  }
}
