/**
 * Medieval Fantasy Theme Provider
 * Extends BaseTheme with Medieval Fantasy specific configuration
 */

import { BaseTheme, parseThemeManifest } from './BaseTheme';
import mappingData from '../../static/assets/themes/fantasy/mapping.json';

const manifest = parseThemeManifest(mappingData);

/**
 * Medieval Fantasy Theme Implementation
 */
export class FantasyTheme extends BaseTheme {
  constructor() {
    super(manifest);
  }
}

// Export a singleton instance for convenience
export const fantasyTheme = new FantasyTheme();

// Export for use in theme registry
export default FantasyTheme;
