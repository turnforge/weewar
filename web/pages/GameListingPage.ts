import { ThemeManager, SplashScreen, BasePage } from '@panyam/tsappkit';

/**
 * Manages the game listing page logic
 */
class GameListingPage extends BasePage {}
GameListingPage.loadAfterPageLoaded("gameListingPage", GameListingPage, "GameListingPage")
