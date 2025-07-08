package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
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

func main() {
	fmt.Println("WeeWar Map Data Extractor")
	fmt.Println("========================")

	// Get the Maps directory
	mapsDir := "games/weewar/data/Maps"
	if _, err := os.Stat(mapsDir); os.IsNotExist(err) {
		fmt.Printf("Maps directory not found: %s\n", mapsDir)
		os.Exit(1)
	}

	// Get all HTML files in the Maps directory
	files, err := filepath.Glob(filepath.Join(mapsDir, "*.html"))
	if err != nil {
		fmt.Printf("Error reading Maps directory: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No HTML files found in Maps directory")
		os.Exit(1)
	}

	fmt.Printf("Found %d map files\n", len(files))

	var allMaps []MapData
	for _, file := range files {
		fmt.Printf("Processing %s...\n", filepath.Base(file))
		
		mapData, err := parseMapHTML(file)
		if err != nil {
			fmt.Printf("Error parsing %s: %v\n", file, err)
			continue
		}
		
		allMaps = append(allMaps, mapData)
		fmt.Printf("  Extracted: %s (%d players, %s)\n", mapData.Name, mapData.Players, mapData.Size)
	}

	// Create the output structure
	output := WeeWarMapsData{
		Maps: allMaps,
		Metadata: struct {
			Version     string `json:"version"`
			ExtractedAt string `json:"extractedAt"`
			TotalMaps   int    `json:"totalMaps"`
		}{
			Version:     "1.0",
			ExtractedAt: "2025-01-08", // Current date
			TotalMaps:   len(allMaps),
		},
	}

	// Write to JSON file
	outputPath := "games/weewar/weewar-maps.json"
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	if err := ioutil.WriteFile(outputPath, jsonData, 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nExtraction complete! Generated %s with %d maps\n", outputPath, len(allMaps))
}

func parseMapHTML(filePath string) (MapData, error) {
	var mapData MapData
	
	// Extract map ID from filename
	filename := filepath.Base(filePath)
	idStr := strings.TrimSuffix(filename, ".html")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return mapData, fmt.Errorf("invalid map ID in filename: %s", filename)
	}
	mapData.ID = id
	
	// Read the HTML file
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return mapData, fmt.Errorf("error reading file: %v", err)
	}

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(string(content)))
	if err != nil {
		return mapData, fmt.Errorf("error parsing HTML: %v", err)
	}

	// Extract data from HTML
	if err := extractMapData(doc, &mapData); err != nil {
		return mapData, fmt.Errorf("error extracting map data: %v", err)
	}

	return mapData, nil
}

func extractMapData(doc *html.Node, mapData *MapData) error {
	// Extract map name from h1 tag
	mapData.Name = extractMapName(doc)
	
	// Extract map image URL
	mapData.ImageURL = extractMapImage(doc, mapData.ID)
	
	// Extract metadata from the sidebar
	extractMapMetadata(doc, mapData)
	
	// Extract tile data
	mapData.Tiles = extractTileData(doc)
	
	// Extract initial units
	mapData.InitialUnits = extractInitialUnits(doc)
	
	// Calculate total tile count
	totalTiles := 0
	for _, count := range mapData.Tiles {
		totalTiles += count
	}
	mapData.TileCount = totalTiles
	
	return nil
}

func extractMapName(doc *html.Node) string {
	var name string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h1" {
			// Check if this h1 has class "mb-3" (map title)
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "mb-3") {
					// Get the text content of the h1
					name = getTextContent(n)
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	
	// Clean up the name (remove extra whitespace and "Pro" indicators)
	name = strings.TrimSpace(name)
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")
	
	return name
}

func extractMapImage(doc *html.Node, mapID int) string {
	var imageURL string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, attr := range n.Attr {
				if attr.Key == "src" && strings.Contains(attr.Val, "map-og.png") {
					imageURL = attr.Val
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	
	// If not found, construct expected path
	if imageURL == "" {
		imageURL = fmt.Sprintf("./%d_files/map-og.png", mapID)
	}
	
	return imageURL
}

func extractMapMetadata(doc *html.Node, mapData *MapData) {
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "p" {
			text := getTextContent(n)
			
			// Extract players
			if strings.Contains(text, "Players") {
				if matches := regexp.MustCompile(`Players\s+(\d+)`).FindStringSubmatch(text); len(matches) > 1 {
					if players, err := strconv.Atoi(matches[1]); err == nil {
						mapData.Players = players
					}
				}
			}
			
			// Extract size
			if strings.Contains(text, "Size") {
				if matches := regexp.MustCompile(`Size\s+([^(]+\(\d+\))`).FindStringSubmatch(text); len(matches) > 1 {
					// Clean up the size string (remove newlines and extra whitespace)
					size := strings.TrimSpace(matches[1])
					size = regexp.MustCompile(`\s+`).ReplaceAllString(size, " ")
					mapData.Size = size
				}
			}
			
			// Extract games played
			if strings.Contains(text, "Games played") {
				if matches := regexp.MustCompile(`Games played\s+(\d+)`).FindStringSubmatch(text); len(matches) > 1 {
					if games, err := strconv.Atoi(matches[1]); err == nil {
						mapData.GamesPlayed = games
					}
				}
			}
			
			// Extract coin settings
			if strings.Contains(text, "Coins") {
				extractCoinSettings(text, mapData)
			}
		}
		
		// Extract creator information
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "card bg-white") {
					// This is likely the creator card
					creatorText := getTextContent(n)
					if creatorText != "" {
						// Extract creator name (usually the first line of text)
						lines := strings.Split(creatorText, "\n")
						for _, line := range lines {
							line = strings.TrimSpace(line)
							if line != "" && !strings.Contains(line, "Created by") {
								mapData.Creator = line
								break
							}
						}
					}
				}
			}
		}
		
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
}

func extractCoinSettings(text string, mapData *MapData) {
	// Extract Start of game
	if matches := regexp.MustCompile(`Start of game:\s*(\d+)`).FindStringSubmatch(text); len(matches) > 1 {
		if start, err := strconv.Atoi(matches[1]); err == nil {
			mapData.Coins.StartOfGame = start
		}
	}
	
	// Extract Per turn
	if matches := regexp.MustCompile(`Per turn:\s*(\d+)`).FindStringSubmatch(text); len(matches) > 1 {
		if perTurn, err := strconv.Atoi(matches[1]); err == nil {
			mapData.Coins.PerTurn = perTurn
		}
	}
	
	// Extract Per base
	if matches := regexp.MustCompile(`Per base:\s*(\d+)`).FindStringSubmatch(text); len(matches) > 1 {
		if perBase, err := strconv.Atoi(matches[1]); err == nil {
			mapData.Coins.PerBase = perBase
		}
	}
}

func extractTileData(doc *html.Node) map[string]int {
	tiles := make(map[string]int)
	
	var traverse func(*html.Node)
	inTilesSection := false
	
	traverse = func(n *html.Node) {
		// Check if we're entering the tiles section
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "name" && attr.Val == "tiles" {
					inTilesSection = true
					break
				}
			}
		}
		
		// Check if we're leaving the tiles section (hit the units section)
		if n.Type == html.ElementNode && n.Data == "h2" && inTilesSection {
			text := getTextContent(n)
			if strings.Contains(text, "Units") {
				inTilesSection = false
			}
		}
		
		// Extract tile data if we're in the tiles section
		if inTilesSection && n.Type == html.ElementNode && n.Data == "a" {
			// Check if this is a tile link
			for _, attr := range n.Attr {
				if attr.Key == "href" && strings.Contains(attr.Val, "tile/view.html") {
					// Extract tile name from data-bs-original-title
					tileName := ""
					for _, attr2 := range n.Attr {
						if attr2.Key == "data-bs-original-title" {
							tileName = attr2.Val
							break
						}
					}
					
					if tileName != "" {
						// Extract count from the <strong> tag
						count := extractCountFromNode(n)
						if count > 0 {
							tiles[tileName] = count
						}
					}
					break
				}
			}
		}
		
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	
	return tiles
}

func extractInitialUnits(doc *html.Node) map[string]int {
	units := make(map[string]int)
	
	var traverse func(*html.Node)
	inUnitsSection := false
	
	traverse = func(n *html.Node) {
		// Check if we're entering the units section
		if n.Type == html.ElementNode && n.Data == "h2" {
			text := getTextContent(n)
			if strings.Contains(text, "Units (Initial)") {
				inUnitsSection = true
			}
		}
		
		// Extract unit data if we're in the units section
		if inUnitsSection && n.Type == html.ElementNode && n.Data == "a" {
			// Check if this is a unit link
			for _, attr := range n.Attr {
				if attr.Key == "href" && strings.Contains(attr.Val, "unit/view.html") {
					// Extract unit name from data-bs-original-title
					unitName := ""
					for _, attr2 := range n.Attr {
						if attr2.Key == "data-bs-original-title" {
							unitName = attr2.Val
							break
						}
					}
					
					if unitName != "" {
						// Extract count from the <strong> tag
						count := extractCountFromNode(n)
						if count > 0 {
							units[unitName] = count
						}
					}
					break
				}
			}
		}
		
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	
	return units
}

func extractCountFromNode(n *html.Node) int {
	var traverse func(*html.Node) int
	traverse = func(n *html.Node) int {
		if n.Type == html.ElementNode && n.Data == "strong" {
			text := getTextContent(n)
			if count, err := strconv.Atoi(text); err == nil {
				return count
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if count := traverse(c); count > 0 {
				return count
			}
		}
		return 0
	}
	return traverse(n)
}

func getTextContent(n *html.Node) string {
	var text string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.TextNode {
			text += n.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(n)
	return strings.TrimSpace(text)
}