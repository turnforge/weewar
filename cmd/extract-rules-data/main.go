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

// RulesData represents the complete rules data structure matching our proto schema
type RulesData struct {
	Units                  map[string]UnitDefinition         `json:"units"`
	Terrains              map[string]TerrainDefinition      `json:"terrains"`
	TerrainUnitProperties map[string]TerrainUnitProperties  `json:"terrainUnitProperties"`
	UnitUnitProperties    map[string]UnitUnitProperties     `json:"unitUnitProperties"`
}

// UnitDefinition matches our proto UnitDefinition
type UnitDefinition struct {
	ID              int32    `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Health          int32    `json:"health"`          // Maximum health points
	Coins           int32    `json:"coins"`           // Cost to build
	MovementPoints  int32    `json:"movementPoints"`  // Movement per turn
	AttackRange     int32    `json:"attackRange"`     // Max attack range
	MinAttackRange  int32    `json:"minAttackRange"`  // Min attack range
	Properties      []string `json:"properties"`      // Special properties/abilities
}

// TerrainDefinition matches our proto TerrainDefinition
type TerrainDefinition struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Type        int32  `json:"type"`
	Description string `json:"description"`
}

// TerrainUnitProperties matches our proto (centralized version)
type TerrainUnitProperties struct {
	MovementCost   float64 `json:"movementCost"`
	HealingBonus   int32   `json:"healingBonus,omitempty"`
	CanBuild       bool    `json:"canBuild,omitempty"`
	CanCapture     bool    `json:"canCapture,omitempty"`
}

// UnitUnitProperties represents unit-vs-unit combat data
type UnitUnitProperties struct {
	Damage *DamageDistribution `json:"damage,omitempty"`
}

// DamageDistribution represents damage probability distribution
type DamageDistribution struct {
	Min     int32                `json:"min"`
	Max     int32                `json:"max"`
	Buckets map[string]float64   `json:"buckets"` // damage_value -> probability
}

func main() {
	fmt.Println("WeeWar Rules Data Extractor")
	fmt.Println("===========================")

	// Check for required directories
	tilesDir := os.Getenv("HOME") + "/dev-app-data/weewar/data/Tiles"
	unitsDir := os.Getenv("HOME") + "/dev-app-data/weewar/data/Units"
	
	if _, err := os.Stat(tilesDir); os.IsNotExist(err) {
		fmt.Printf("Tiles directory not found: %s\n", tilesDir)
		os.Exit(1)
	}
	
	if _, err := os.Stat(unitsDir); os.IsNotExist(err) {
		fmt.Printf("Units directory not found: %s\n", unitsDir)
		os.Exit(1)
	}

	// Initialize rules data
	rulesData := RulesData{
		Units:                  make(map[string]UnitDefinition),
		Terrains:              make(map[string]TerrainDefinition),
		TerrainUnitProperties: make(map[string]TerrainUnitProperties),
		UnitUnitProperties:    make(map[string]UnitUnitProperties),
	}

	// Extract terrain data
	fmt.Println("Extracting terrain data...")
	if err := extractTerrainsData(tilesDir, &rulesData); err != nil {
		fmt.Printf("Error extracting terrain data: %v\n", err)
		os.Exit(1)
	}

	// Extract unit data  
	fmt.Println("Extracting unit data...")
	if err := extractUnitsData(unitsDir, &rulesData); err != nil {
		fmt.Printf("Error extracting unit data: %v\n", err)
		os.Exit(1)
	}

	// Write output
	outputPath := "games/weewar/weewar-rules.json"
	jsonData, err := json.MarshalIndent(rulesData, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	if err := ioutil.WriteFile(outputPath, jsonData, 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nExtraction complete! Generated %s\n", outputPath)
	fmt.Printf("- %d units\n", len(rulesData.Units))
	fmt.Printf("- %d terrains\n", len(rulesData.Terrains))
	fmt.Printf("- %d terrain-unit properties\n", len(rulesData.TerrainUnitProperties))
	fmt.Printf("- %d unit-unit combat properties\n", len(rulesData.UnitUnitProperties))
}

func extractTerrainsData(tilesDir string, rulesData *RulesData) error {
	files, err := filepath.Glob(filepath.Join(tilesDir, "*.html"))
	if err != nil {
		return fmt.Errorf("error reading tiles directory: %v", err)
	}

	for _, file := range files {
		filename := filepath.Base(file)
		terrainIDStr := strings.TrimSuffix(filename, ".html")
		terrainID, err := strconv.Atoi(terrainIDStr)
		if err != nil {
			fmt.Printf("Skipping invalid terrain file: %s\n", filename)
			continue
		}

		fmt.Printf("Processing terrain %d...\n", terrainID)
		
		// Read HTML
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("error reading file %s: %v", file, err)
		}

		doc, err := html.Parse(strings.NewReader(string(content)))
		if err != nil {
			return fmt.Errorf("error parsing HTML %s: %v", file, err)
		}

		// Extract terrain definition
		terrainDef := TerrainDefinition{
			ID:   int32(terrainID),
			Name: extractTerrainName(doc),
			Type: 1, // Default terrain type - could be enhanced to parse from HTML
		}
		rulesData.Terrains[terrainIDStr] = terrainDef

		// Extract terrain-unit properties from the units interaction table
		if err := extractTerrainUnitInteractions(doc, int32(terrainID), rulesData); err != nil {
			return fmt.Errorf("error extracting terrain-unit interactions for %s: %v", file, err)
		}
	}

	return nil
}

func extractUnitsData(unitsDir string, rulesData *RulesData) error {
	files, err := filepath.Glob(filepath.Join(unitsDir, "*.html"))
	if err != nil {
		return fmt.Errorf("error reading units directory: %v", err)
	}

	for _, file := range files {
		filename := filepath.Base(file)
		unitIDStr := strings.TrimSuffix(filename, ".html")
		unitID, err := strconv.Atoi(unitIDStr)
		if err != nil {
			fmt.Printf("Skipping invalid unit file: %s\n", filename)
			continue
		}

		fmt.Printf("Processing unit %d...\n", unitID)
		
		// Read HTML
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("error reading file %s: %v", file, err)
		}

		doc, err := html.Parse(strings.NewReader(string(content)))
		if err != nil {
			return fmt.Errorf("error parsing HTML %s: %v", file, err)
		}

		// Extract unit definition
		unitDef, err := extractUnitDefinition(doc, int32(unitID))
		if err != nil {
			return fmt.Errorf("error extracting unit definition for %s: %v", file, err)
		}
		rulesData.Units[unitIDStr] = unitDef

		// Extract unit-unit combat properties
		if err := extractUnitCombatProperties(doc, int32(unitID), rulesData); err != nil {
			return fmt.Errorf("error extracting combat properties for %s: %v", file, err)
		}
	}

	return nil
}

func extractTerrainName(doc *html.Node) string {
	var name string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h1" {
			// Look for the main terrain name
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "mb-3") {
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
	
	// Clean up name
	name = strings.TrimSpace(name)
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")
	return name
}

func extractTerrainUnitInteractions(doc *html.Node, terrainID int32, rulesData *RulesData) error {
	// Find the units interaction table
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tbody" {
			// Process each row in the table
			for row := n.FirstChild; row != nil; row = row.NextSibling {
				if row.Type == html.ElementNode && row.Data == "tr" {
					extractUnitRowData(row, terrainID, rulesData)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	
	return nil
}

func extractUnitRowData(row *html.Node, terrainID int32, rulesData *RulesData) {
	var unitID int32 = -1
	properties := TerrainUnitProperties{
		MovementCost: 1.0, // Default movement cost
	}
	
	cellIndex := 0
	for cell := row.FirstChild; cell != nil; cell = cell.NextSibling {
		if cell.Type == html.ElementNode && cell.Data == "td" {
			cellText := getTextContent(cell)
			
			switch cellIndex {
			case 0: // Unit name and ID extraction
				unitID = extractUnitIDFromCell(cell)
			case 3: // Movement cost column
				properties.MovementCost = parseMovementCost(cellText)
			case 6: // Heal column
				if heal := parseModifierValue(cellText); heal > 0 {
					properties.HealingBonus = heal
				}
			case 7: // Captures column
				properties.CanCapture = containsCheckmark(cell)
			case 8: // Builds column
				properties.CanBuild = containsCheckmark(cell)
			}
			cellIndex++
		}
	}
	
	// Create terrain-unit property key using our centralized format
	if unitID > 0 {
		key := fmt.Sprintf("%d:%d", terrainID, unitID)
		rulesData.TerrainUnitProperties[key] = properties
	}
}

func extractUnitDefinition(doc *html.Node, unitID int32) (UnitDefinition, error) {
	unitDef := UnitDefinition{
		ID:             unitID,
		Health:         100, // Always 100 in WeeWar
		MinAttackRange: 1,   // Default minimum attack range
	}
	
	// Extract basic properties from the sidebar
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "p" {
			text := getTextContent(n)
			
			// Extract different properties based on <strong> labels
			if strings.Contains(text, "Movement") && !strings.Contains(text, "Build Percentage") {
				// Extract movement points - the number appears after <br>
				lines := strings.Split(text, "\n")
				for i, line := range lines {
					if strings.Contains(line, "Movement") && i+1 < len(lines) {
						nextLine := strings.TrimSpace(lines[i+1])
						if matches := regexp.MustCompile(`^(\d+)`).FindStringSubmatch(nextLine); len(matches) > 1 {
							if mov, err := strconv.Atoi(matches[1]); err == nil {
								unitDef.MovementPoints = int32(mov)
							}
						}
						break
					}
				}
			} else if strings.Contains(text, "Attack Range") {
				// Extract attack range - the number appears after <br>
				lines := strings.Split(text, "\n")
				for i, line := range lines {
					if strings.Contains(line, "Attack Range") && i+1 < len(lines) {
						nextLine := strings.TrimSpace(lines[i+1])
						if matches := regexp.MustCompile(`^(\d+)`).FindStringSubmatch(nextLine); len(matches) > 1 {
							if rng, err := strconv.Atoi(matches[1]); err == nil {
								unitDef.AttackRange = int32(rng)
							}
						}
						break
					}
				}
			} else if strings.Contains(text, "Coins") && !strings.Contains(text, "Coins Spent") {
				// Extract coin cost - the number appears after the coin image
				lines := strings.Split(text, "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if matches := regexp.MustCompile(`^\d+$`).FindStringSubmatch(line); len(matches) > 0 {
						if cost, err := strconv.Atoi(matches[0]); err == nil {
							unitDef.Coins = int32(cost)
							break
						}
					}
				}
			}
		}
		
		// Extract name from page title
		if n.Type == html.ElementNode && n.Data == "title" {
			title := getTextContent(n)
			if strings.Contains(title, " - Units - ") {
				parts := strings.Split(title, " - Units - ")
				if len(parts) > 0 {
					unitDef.Name = strings.TrimSpace(parts[0])
				}
			}
		}
		
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	
	return unitDef, nil
}

func extractUnitCombatProperties(doc *html.Node, attackerID int32, rulesData *RulesData) error {
	// Extract combat damage distributions from the damage charts
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			// Look for unit damage cards
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "card") {
					if defenderID := extractDefenderIDFromCard(n); defenderID > 0 {
						if damage := extractDamageDistribution(n); damage != nil {
							key := fmt.Sprintf("%d:%d", attackerID, defenderID)
							rulesData.UnitUnitProperties[key] = UnitUnitProperties{
								Damage: damage,
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
	
	return nil
}

// Utility functions
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

func extractUnitIDFromCell(cell *html.Node) int32 {
	// Extract unit ID from href like "unit/view.html?unitId=1"
	var traverse func(*html.Node) int32
	traverse = func(n *html.Node) int32 {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && strings.Contains(attr.Val, "unitId=") {
					if matches := regexp.MustCompile(`unitId=(\d+)`).FindStringSubmatch(attr.Val); len(matches) > 1 {
						if id, err := strconv.Atoi(matches[1]); err == nil {
							return int32(id)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if result := traverse(c); result > 0 {
				return result
			}
		}
		return 0
	}
	return traverse(cell)
}

func parseModifierValue(text string) int32 {
	text = strings.TrimSpace(text)
	if text == "" || strings.Contains(text, "muted") || text == "0" {
		return 0
	}
	
	// Parse values like "+2", "-1", "6"
	if matches := regexp.MustCompile(`([+-]?\d+)`).FindStringSubmatch(text); len(matches) > 1 {
		if val, err := strconv.Atoi(matches[1]); err == nil {
			return int32(val)
		}
	}
	
	return 0
}

func parseMovementCost(text string) float64 {
	text = strings.TrimSpace(text)
	if text == "" {
		return 1.0 // Default movement cost
	}
	
	// Parse values like "1", "1.25", "2"
	if matches := regexp.MustCompile(`(\d+(?:\.\d+)?)`).FindStringSubmatch(text); len(matches) > 1 {
		if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return val
		}
	}
	
	return 1.0
}

func containsCheckmark(cell *html.Node) bool {
	// Look for fa-check class indicating a checkmark
	var traverse func(*html.Node) bool
	traverse = func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "span" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "fa-check") {
					return true
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if traverse(c) {
				return true
			}
		}
		return false
	}
	return traverse(cell)
}

func extractDefenderIDFromCard(card *html.Node) int32 {
	// Find the defender unit ID from the card header link
	var traverse func(*html.Node) int32
	traverse = func(n *html.Node) int32 {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && strings.Contains(attr.Val, "unitId=") {
					if matches := regexp.MustCompile(`unitId=(\d+)`).FindStringSubmatch(attr.Val); len(matches) > 1 {
						if id, err := strconv.Atoi(matches[1]); err == nil {
							return int32(id)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if result := traverse(c); result > 0 {
				return result
			}
		}
		return 0
	}
	return traverse(card)
}

func extractDamageDistribution(card *html.Node) *DamageDistribution {
	damage := &DamageDistribution{
		Buckets: make(map[string]float64),
	}
	
	// Extract damage probabilities from tooltip data
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, attr := range n.Attr {
				if attr.Key == "data-bs-original-title" && strings.Contains(attr.Val, "% of the time") {
					// Parse tooltip like "72.2% of the time Soldier deals 1 damage"
					if matches := regexp.MustCompile(`(\d+(?:\.\d+)?)% of the time.*?deals (\d+) damage`).FindStringSubmatch(attr.Val); len(matches) > 2 {
						if prob, err := strconv.ParseFloat(matches[1], 64); err == nil {
							if dmg, err := strconv.Atoi(matches[2]); err == nil {
								damage.Buckets[matches[2]] = prob / 100.0 // Convert percentage to decimal
								if damage.Min == 0 || int32(dmg) < damage.Min {
									damage.Min = int32(dmg)
								}
								if int32(dmg) > damage.Max {
									damage.Max = int32(dmg)
								}
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
	traverse(card)
	
	if len(damage.Buckets) == 0 {
		return nil // No damage data found
	}
	
	return damage
}