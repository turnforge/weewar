/**
 * Modern Military Theme Provider
 * Extends BaseTheme with Modern Military specific configuration
 */

import { BaseTheme, parseThemeManifest } from './BaseTheme';
import mappingData from '../../static/assets/themes/modern/mapping.json';

const manifest = parseThemeManifest(mappingData);

/**
 * Modern Military Theme Implementation
 */
export class ModernTheme extends BaseTheme {
  constructor() {
    super(manifest);
  }
}

// Export a singleton instance for convenience
export const modernTheme = new ModernTheme();

// Export for use in theme registry
export default ModernTheme;
