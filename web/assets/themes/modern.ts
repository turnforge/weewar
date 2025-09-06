/**
 * Modern Military Theme Provider
 * Extends BaseTheme with Modern Military specific configuration
 */

import { BaseTheme } from './BaseTheme';
import mappingData from '../../static/assets/themes/modern/mapping.json';

/**
 * Modern Military Theme Implementation
 */
export class ModernTheme extends BaseTheme {
  protected basePath = '/static/assets/themes/modern';
  protected themeName = 'Modern Military';
  protected themeVersion = '1.0.0';

  constructor() {
    super(mappingData);
  }
}

// Export a singleton instance for convenience
export const modernTheme = new ModernTheme();

// Export for use in theme registry
export default ModernTheme;
