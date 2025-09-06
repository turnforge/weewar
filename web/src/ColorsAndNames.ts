
export const BRUSH_SIZE_NAMES = ['Single (1 hex)', 'Small (3 hexes)', 'Medium (5 hexes)', 'Large (9 hexes)', 'X-Large (15 hexes)', 'XX-Large (25 hexes)'];

export const AllowedUnitIDs = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 37, 38, 39, 40, 41, 44];
export const CityTerrainIds = [1, 2, 3, 6, 16, 20, 21, 25]; // Base, Hospital, Silo, Mines, City, Tower

// These IDs are used throughout the system and should remain here
// The actual names come from the active theme

// Player colors - text colors
export const PLAYER_COLORS: { [key: number]: string } = {
    1: 'text-blue-600 dark:text-red-400',
    2: 'text-blue-600 dark:text-blue-400',
    3: 'text-green-600 dark:text-green-400',
    4: 'text-yellow-600 dark:text-yellow-400',
    5: 'text-orange-600 dark:text-orange-400',
    6: 'text-purple-600 dark:text-purple-400',
    7: 'text-pink-600 dark:text-pink-400',
    8: 'text-cyan-600 dark:text-cyan-400',
    9: 'text-cyan-600 dark:text-cyan-400',
    10: 'text-cyan-600 dark:text-cyan-400',
    11: 'text-cyan-600 dark:text-cyan-400',
    12: 'text-cyan-600 dark:text-cyan-400',
};

// Player background colors for status displays  
export const PLAYER_BG_COLORS: { [key: number]: string } = {
    1: 'bg-sky-900 text-red-800 dark:bg-sky-900 dark:text-red-200',
    2: 'bg-red-100 text-blue-800 dark:bg-red-900 dark:text-blue-200',
    3: 'bg-yellow-100 text-blue-800 dark:bg-yellow-900 dark:text-gray-200',
    4: 'bg-gray-100 text-blue-800 dark:bg-gray-900 dark:text-yellow-200',
    5: 'bg-pink-100 text-blue-800 dark:bg-pink-900 dark:text-pink-200',
    6: 'bg-orange-100 text-blue-800 dark:bg-orange-900 dark:text-orange-200',
    7: 'bg-teal-100 text-blue-800 dark:bg-teal-900 dark:text-teal-200',
    8: 'bg-green-100 text-blue-800 dark:bg-green-900 dark:text-green-200',
    9: 'bg-indigo-100 text-blue-800 dark:bg-indigo-900 dark:text-indigo-200',
    10: 'bg-brown-100 text-brown-800 dark:bg-brown-900 dark:text-brown-200',
    11: 'bg-cyan-100 text-blue-800 dark:bg-cyan-900 dark:text-purple-200',
    12: 'bg-purple-100 text-blue-800 dark:bg-purple-900 dark:text-cyan-200',
};

