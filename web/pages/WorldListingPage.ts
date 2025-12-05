import { ThemeManager, SplashScreen, BasePage } from '@panyam/tsappkit';

/**
 * Manages the game listing page logic
 */
class WorldListingPage extends BasePage {}
WorldListingPage.loadAfterPageLoaded("gameListingPage", WorldListingPage, "WorldListingPage")

