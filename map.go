package weewar

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"strings"
)

type MapData struct {
	ID          int                `json:"id"`
	Name        string             `json:"name"`
	ImageURL    string             `json:"imageURL"`
	Creator     string             `json:"creator"`
	Players     int                `json:"players"`
	Size        string             `json:"size"`
	TileCount   int                `json:"tileCount"`
	GamesPlayed int                `json:"gamesPlayed"`
	Coins       CoinSettings       `json:"coins"`
	Tiles       map[string]int     `json:"tiles"`
	InitialUnits map[string]int    `json:"initialUnits"`
	CreatedOn   string             `json:"createdOn,omitempty"`
	LastUpdated string             `json:"lastUpdated,omitempty"`
	Favorited   int                `json:"favorited,omitempty"`
	WinStats    map[string]float64 `json:"winStats,omitempty"`
}

type CoinSettings struct {
	StartOfGame int `json:"startOfGame"`
	PerTurn     int `json:"perTurn"`
	PerBase     int `json:"perBase"`
}

type WeeWarMapsData struct {
	Maps []MapData `json:"maps"`
	Metadata struct {
		Version     string `json:"version"`
		ExtractedAt string `json:"extractedAt"`
		TotalMaps   int    `json:"totalMaps"`
	} `json:"metadata"`
}

type WeeWarMapSystem struct {
	mapsData WeeWarMapsData
}

func NewWeeWarMapSystem() (*WeeWarMapSystem, error) {
	mapsData, err := loadWeeWarMaps()
	if err != nil {
		return nil, fmt.Errorf("failed to load WeeWar maps: %w", err)
	}

	return &WeeWarMapSystem{
		mapsData: mapsData,
	}, nil
}

func (wms *WeeWarMapSystem) GetMapByID(mapID int) (*MapData, error) {
	for _, mapData := range wms.mapsData.Maps {
		if mapData.ID == mapID {
			return &mapData, nil
		}
	}
	return nil, fmt.Errorf("map with ID %d not found", mapID)
}

func (wms *WeeWarMapSystem) GetMapByName(name string) (*MapData, error) {
	for _, mapData := range wms.mapsData.Maps {
		if strings.EqualFold(mapData.Name, name) {
			return &mapData, nil
		}
	}
	return nil, fmt.Errorf("map with name %s not found", name)
}

func (wms *WeeWarMapSystem) GetAllMaps() []MapData {
	return wms.mapsData.Maps
}

func (wms *WeeWarMapSystem) GetMapsByPlayerCount(playerCount int) []MapData {
	var maps []MapData
	for _, mapData := range wms.mapsData.Maps {
		if mapData.Players == playerCount {
			maps = append(maps, mapData)
		}
	}
	return maps
}

func (wms *WeeWarMapSystem) CreateGameConfigFromMap(mapData *MapData) (*WeeWarConfig, error) {
	// Calculate board dimensions from tile count
	// We need to ensure the board size can accommodate all tiles
	boardSize := int(math.Ceil(math.Sqrt(float64(mapData.TileCount))))
	
	// Ensure minimum board size
	if boardSize < 3 {
		boardSize = 3
	}
	
	// Create players
	players := make([]WeeWarPlayer, mapData.Players)
	for i := 0; i < mapData.Players; i++ {
		players[i] = WeeWarPlayer{
			ID:   fmt.Sprintf("player_%d", i+1),
			Name: fmt.Sprintf("Player %d", i+1),
			Team: i,
		}
	}

	// Generate terrain map from tile data
	terrainMap := generateTerrainMap(mapData.Tiles, boardSize, boardSize)

	// Create starting units configuration
	startingUnits := make(map[string][]string)
	for i := 0; i < mapData.Players; i++ {
		playerID := fmt.Sprintf("player_%d", i+1)
		startingUnits[playerID] = distributeUnitsToPlayer(mapData.InitialUnits, i, mapData.Players)
	}

	config := &WeeWarConfig{
		BoardWidth:    boardSize,
		BoardHeight:   boardSize,
		Players:       players,
		StartingUnits: startingUnits,
		TerrainMap:    terrainMap,
		MapData:       mapData, // Add reference to original map data
	}

	return config, nil
}

func generateTerrainMap(tiles map[string]int, width, height int) [][]string {
	terrainMap := make([][]string, height)
	for i := range terrainMap {
		terrainMap[i] = make([]string, width)
	}

	// Create a list of all tiles to place
	var tileList []string
	for terrainType, count := range tiles {
		for i := 0; i < count; i++ {
			tileList = append(tileList, terrainType)
		}
	}

	// Place tiles on the map
	// This is a simple distribution - in a real implementation, you'd want
	// to use the actual map layout from the image
	index := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if index < len(tileList) {
				terrainMap[y][x] = tileList[index]
				index++
			} else {
				// Fill remaining spaces with grass
				terrainMap[y][x] = "Grass"
			}
		}
	}

	return terrainMap
}

func distributeUnitsToPlayer(initialUnits map[string]int, playerIndex, totalPlayers int) []string {
	var playerUnits []string
	
	// Distribute units evenly among players
	for unitType, totalCount := range initialUnits {
		// Calculate how many units this player should get
		baseCount := totalCount / totalPlayers
		remainder := totalCount % totalPlayers
		
		playerCount := baseCount
		if playerIndex < remainder {
			playerCount++
		}
		
		// Add the units to this player's list
		for i := 0; i < playerCount; i++ {
			playerUnits = append(playerUnits, unitType)
		}
	}
	
	return playerUnits
}

func (wms *WeeWarMapSystem) GetMapStatistics() map[string]interface{} {
	stats := map[string]interface{}{
		"totalMaps":          len(wms.mapsData.Maps),
		"mapsByPlayerCount":  make(map[int]int),
		"mapsBySize":         make(map[string]int),
		"mostPlayedMap":      nil,
		"mostPlayedGames":    0,
		"totalGamesPlayed":   0,
	}

	playerCounts := make(map[int]int)
	sizes := make(map[string]int)
	var mostPlayedMap *MapData
	mostPlayedGames := 0
	totalGames := 0

	for _, mapData := range wms.mapsData.Maps {
		// Count by player count
		playerCounts[mapData.Players]++
		
		// Count by size
		size := extractSizeCategory(mapData.Size)
		sizes[size]++
		
		// Track most played map
		if mapData.GamesPlayed > mostPlayedGames {
			mostPlayedGames = mapData.GamesPlayed
			mostPlayedMap = &mapData
		}
		
		totalGames += mapData.GamesPlayed
	}

	stats["mapsByPlayerCount"] = playerCounts
	stats["mapsBySize"] = sizes
	stats["totalGamesPlayed"] = totalGames
	if mostPlayedMap != nil {
		stats["mostPlayedMap"] = map[string]interface{}{
			"name":        mostPlayedMap.Name,
			"gamesPlayed": mostPlayedMap.GamesPlayed,
		}
	}

	return stats
}

func extractSizeCategory(sizeStr string) string {
	if strings.Contains(sizeStr, "Small") {
		return "Small"
	} else if strings.Contains(sizeStr, "Medium") {
		return "Medium"
	} else if strings.Contains(sizeStr, "Large") {
		return "Large"
	}
	return "Unknown"
}

func loadWeeWarMaps() (WeeWarMapsData, error) {
	var mapsData WeeWarMapsData
	
	content, err := ioutil.ReadFile("games/weewar/weewar-maps.json")
	if err != nil {
		return mapsData, fmt.Errorf("failed to read weewar-maps.json: %w", err)
	}

	if err := json.Unmarshal(content, &mapsData); err != nil {
		return mapsData, fmt.Errorf("failed to unmarshal WeeWar maps data: %w", err)
	}

	return mapsData, nil
}

// Helper function to create a game with a specific map
func CreateWeeWarGameWithMap(mapID int) (*WeeWarGame, error) {
	mapSystem, err := NewWeeWarMapSystem()
	if err != nil {
		return nil, fmt.Errorf("failed to create map system: %w", err)
	}

	mapData, err := mapSystem.GetMapByID(mapID)
	if err != nil {
		return nil, fmt.Errorf("failed to get map: %w", err)
	}

	config, err := mapSystem.CreateGameConfigFromMap(mapData)
	if err != nil {
		return nil, fmt.Errorf("failed to create game config: %w", err)
	}

	game, err := NewWeeWarGame(*config)
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	return game, nil
}

// Helper function to create a game with a specific map by name
func CreateWeeWarGameWithMapName(mapName string) (*WeeWarGame, error) {
	mapSystem, err := NewWeeWarMapSystem()
	if err != nil {
		return nil, fmt.Errorf("failed to create map system: %w", err)
	}

	mapData, err := mapSystem.GetMapByName(mapName)
	if err != nil {
		return nil, fmt.Errorf("failed to get map: %w", err)
	}

	config, err := mapSystem.CreateGameConfigFromMap(mapData)
	if err != nil {
		return nil, fmt.Errorf("failed to create game config: %w", err)
	}

	game, err := NewWeeWarGame(*config)
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	return game, nil
}