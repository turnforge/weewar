/**
 * Asset Theme Preference Manager
 *
 * Manages user's asset theme preference (fantasy, classic, etc.)
 * Stores preference in both localStorage (frontend) and cookie (backend sync)
 */
export class AssetThemePreference {
    private static readonly STORAGE_KEY = 'assetTheme';
    private static readonly COOKIE_NAME = 'assetTheme';
    private static readonly COOKIE_MAX_AGE = 365 * 24 * 60 * 60; // 1 year in seconds
    private static readonly DEFAULT_THEME = 'fantasy';

    /**
     * Get the current asset theme preference
     * Priority: URL query param > localStorage > cookie > default
     */
    public static get(): string {
        // Check URL query parameter first
        const urlParams = new URLSearchParams(window.location.search);
        const urlTheme = urlParams.get('theme');
        if (urlTheme) {
            return urlTheme;
        }

        // Check localStorage
        const localStorageTheme = localStorage.getItem(AssetThemePreference.STORAGE_KEY);
        if (localStorageTheme) {
            return localStorageTheme;
        }

        // Check cookie
        const cookieTheme = AssetThemePreference.getCookie(AssetThemePreference.COOKIE_NAME);
        if (cookieTheme) {
            return cookieTheme;
        }

        // Fall back to default
        return AssetThemePreference.DEFAULT_THEME;
    }

    /**
     * Save asset theme preference to both localStorage and cookie
     */
    public static set(theme: string): void {
        // Save to localStorage for frontend
        localStorage.setItem(AssetThemePreference.STORAGE_KEY, theme);

        // Save to cookie for backend
        AssetThemePreference.setCookie(theme);
    }

    /**
     * Get cookie value by name
     */
    private static getCookie(name: string): string | null {
        const nameEQ = name + "=";
        const ca = document.cookie.split(';');
        for (let i = 0; i < ca.length; i++) {
            let c = ca[i];
            while (c.charAt(0) === ' ') c = c.substring(1, c.length);
            if (c.indexOf(nameEQ) === 0) return c.substring(nameEQ.length, c.length);
        }
        return null;
    }

    /**
     * Set cookie value
     */
    private static setCookie(theme: string): void {
        const expires = new Date();
        expires.setTime(expires.getTime() + AssetThemePreference.COOKIE_MAX_AGE * 1000);
        document.cookie = `${AssetThemePreference.COOKIE_NAME}=${theme}; expires=${expires.toUTCString()}; path=/; SameSite=Strict`;
    }

    /**
     * Sync localStorage to cookie if they don't match
     * Call this on page load to ensure backend has the latest preference
     */
    public static sync(): void {
        const localStorageTheme = localStorage.getItem(AssetThemePreference.STORAGE_KEY);
        const cookieTheme = AssetThemePreference.getCookie(AssetThemePreference.COOKIE_NAME);

        // If localStorage has a value but cookie doesn't match, sync cookie
        if (localStorageTheme && localStorageTheme !== cookieTheme) {
            AssetThemePreference.setCookie(localStorageTheme);
        }
    }

    /**
     * Clear asset theme preference (will use default)
     */
    public static clear(): void {
        localStorage.removeItem(AssetThemePreference.STORAGE_KEY);
        document.cookie = `${AssetThemePreference.COOKIE_NAME}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/; SameSite=Strict`;
    }

    /**
     * Get the default theme
     */
    public static getDefault(): string {
        return AssetThemePreference.DEFAULT_THEME;
    }
}
