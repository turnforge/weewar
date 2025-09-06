/**
 * Medieval Fantasy Theme Provider
 * Extends BaseTheme with Medieval Fantasy specific configuration
 */

import { BaseTheme } from './BaseTheme';
import mappingData from '../../static/assets/themes/fantasy/mapping.json';

/**
 * Medieval Fantasy Theme Implementation
 */
export class FantasyTheme extends BaseTheme {
  protected basePath = '/static/assets/themes/fantasy';
  protected themeName = 'Medieval Fantasy';
  protected themeVersion = '1.0.0';

  constructor() {
    super(mappingData);
  }
}

// Export a singleton instance for convenience
export const fantasyTheme = new FantasyTheme();

// Export for use in theme registry
export default FantasyTheme;
