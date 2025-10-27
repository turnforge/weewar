package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	weewarv1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"golang.org/x/net/html"
)

// RulesData represents the complete rules data structure matching our proto schema
type RulesData struct {
	Units                 map[string]*weewarv1.UnitDefinition        `json:"units"`
	Terrains              map[string]*weewarv1.TerrainDefinition     `json:"terrains"`
	TerrainUnitProperties map[string]*weewarv1.TerrainUnitProperties `json:"terrainUnitProperties"`
	UnitUnitProperties    map[string]*weewarv1.UnitUnitProperties    `json:"unitUnitProperties"`
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
		Units:                 make(map[string]*weewarv1.UnitDefinition),
		Terrains:              make(map[string]*weewarv1.TerrainDefinition),
		TerrainUnitProperties: make(map[string]*weewarv1.TerrainUnitProperties),
		UnitUnitProperties:    make(map[string]*weewarv1.UnitUnitProperties),
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
	outputPath := "weewar-rules-new.json"
	jsonData, err := json.MarshalIndent(rulesData, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
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
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("error reading file %s: %v", file, err)
		}

		doc, err := html.Parse(strings.NewReader(string(content)))
		if err != nil {
			return fmt.Errorf("error parsing HTML %s: %v", file, err)
		}

		// Extract terrain definition
		terrainDef := &weewarv1.TerrainDefinition{
			Id:   int32(terrainID),
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
		log.Println("Processing file: ", filename)

		fmt.Printf("Processing unit %d...\n", unitID)

		// Read HTML
		content, err := os.ReadFile(file)
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
	// First, find the table headers to understand column layout
	var columnHeaders []string
	var foundTable bool

	var findHeaders func(*html.Node)
	findHeaders = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "thead" && !foundTable {
			// Extract column headers
			for row := n.FirstChild; row != nil; row = row.NextSibling {
				if row.Type == html.ElementNode && row.Data == "tr" {
					for cell := row.FirstChild; cell != nil; cell = cell.NextSibling {
						if cell.Type == html.ElementNode && cell.Data == "th" {
							headerText := getTextContent(cell)
							columnHeaders = append(columnHeaders, strings.ToLower(strings.TrimSpace(headerText)))
						}
					}
					foundTable = true
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if !foundTable {
				findHeaders(c)
			}
		}
	}
	findHeaders(doc)

	// Now find the tbody and process rows with column header knowledge
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tbody" {
			// Process each row in the table
			for row := n.FirstChild; row != nil; row = row.NextSibling {
				if row.Type == html.ElementNode && row.Data == "tr" {
					extractUnitRowData(row, terrainID, columnHeaders, rulesData)
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

func extractUnitRowData(row *html.Node, terrainID int32, columnHeaders []string, rulesData *RulesData) {
	var unitID int32 = -1
	properties := &weewarv1.TerrainUnitProperties{
		TerrainId:    terrainID,
		MovementCost: 1.0, // Default movement cost
	}

	cellIndex := 0
	for cell := row.FirstChild; cell != nil; cell = cell.NextSibling {
		if cell.Type == html.ElementNode && cell.Data == "td" {
			// Check if we have a valid column header for this index
			if cellIndex < len(columnHeaders) {
				columnName := columnHeaders[cellIndex]
				cellText := getTextContent(cell)

				switch columnName {
				case "unit":
					unitID = extractUnitIDFromCell(cell)
				case "attack":
					if attack := parseModifierValue(cellText); attack != 0 {
						properties.AttackBonus = attack
					}
				case "defense":
					if defense := parseModifierValue(cellText); defense != 0 {
						properties.DefenseBonus = defense
					}
				case "movement":
					properties.MovementCost = parseMovementCost(cellText)
				case "heal":
					if heal := parseModifierValue(cellText); heal > 0 {
						properties.HealingBonus = heal
					}
				case "captures":
					properties.CanCapture = containsCheckmark(cell)
				case "builds":
					properties.CanBuild = containsCheckmark(cell)
				}
			}
			cellIndex++
		}
	}

	// Create terrain-unit property key using our centralized format
	if unitID > 0 {
		properties.UnitId = unitID
		key := fmt.Sprintf("%d:%d", terrainID, unitID)
		rulesData.TerrainUnitProperties[key] = properties
	}
}

func extractUnitDefinition(doc *html.Node, unitID int32) (*weewarv1.UnitDefinition, error) {
	unitDef := &weewarv1.UnitDefinition{
		Id:     unitID,
		Health: 100, // Always 100 in WeeWar
		// MinAttackRange and AttackRange will be extracted from HTML
	}

	// Extract basic properties from the sidebar
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "p" {
			text := getTextContent(n)

			// Extract different properties based on <strong> labels
			if strings.Contains(text, "Movement") && !strings.Contains(text, "Build Percentage") {
				// Extract movement points - the number appears after <br>
				log.Println("Lines: ", text)
				lines := strings.Split(text, "\n")
				for i, line := range lines {
					if strings.Contains(line, "Movement") && i+1 < len(lines) {
						nextLine := strings.TrimSpace(strings.TrimSpace(strings.Join(lines[i+1:], "\n")))
						matches := regexp.MustCompile(`(\d+(?:\.\d+)?)`).FindStringSubmatch(nextLine)
						log.Println("Here???, ", lines[i], "Next: ", nextLine, "Matches: ", matches, "Matches[1]: ", matches[1], matches[0])
						if len(matches) > 1 {
							unitDef.MovementPoints = parseMovementCost(matches[1])
						}
						break
					}
				}
			} else if strings.Contains(text, "Attack Range") {
				// Extract attack range - the number appears after <br>
				// Two formats:
				//   1. Single value: "1 (adjacent enemy units)" -> AttackRange=1, MinAttackRange=1
				//   2. Range: "2 - 3" -> AttackRange=3 (max), MinAttackRange=2
				log.Println("Attack Range text: ", text)
				lines := strings.Split(text, "\n")
				for i, line := range lines {
					if strings.Contains(line, "Attack Range") && i+1 < len(lines) {
						// nextLine := strings.TrimSpace(lines[i+1])
						nextLine := strings.TrimSpace(strings.TrimSpace(strings.Join(lines[i+1:], "\n")))
						log.Println("Attack Range - Next line: ", nextLine)

						// Check for range format: "2 - 3"
						if matches := regexp.MustCompile(`^(\d+)\s*-\s*(\d+)`).FindStringSubmatch(nextLine); len(matches) > 2 {
							log.Println("Matched range format: ", matches)
							if minRng, err := strconv.Atoi(matches[1]); err == nil {
								unitDef.MinAttackRange = int32(minRng)
							}
							if maxRng, err := strconv.Atoi(matches[2]); err == nil {
								unitDef.AttackRange = int32(maxRng)
							}
						} else if matches := regexp.MustCompile(`^(\d+)`).FindStringSubmatch(nextLine); len(matches) > 1 {
							// Single value format: "1 (adjacent...)"
							log.Println("Matched single value format: ", matches)
							if rng, err := strconv.Atoi(matches[1]); err == nil {
								unitDef.AttackRange = int32(rng)
								unitDef.MinAttackRange = int32(rng) // Min and max are the same
							}
						} else {
							log.Println("NO MATCH for Attack Range!")
						}
						break
					}
				}
			} else if strings.Contains(text, "Coins") && !strings.Contains(text, "Coins Spent") {
				// Extract coin cost - the number appears after the coin image
				lines := strings.Split(text, "\n")
				for _, line := range lines {
					line = strings.Replace(strings.TrimSpace(line), ",", "", -1)
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
							rulesData.UnitUnitProperties[key] = &weewarv1.UnitUnitProperties{
								AttackerId: attackerID,
								DefenderId: defenderID,
								Damage:     damage,
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

func extractDamageDistribution(card *html.Node) *weewarv1.DamageDistribution {
	damage := &weewarv1.DamageDistribution{
		Ranges: []*weewarv1.DamageRange{},
	}

	// Use a map to deduplicate damage values (HTML may have duplicate tooltips)
	damageMap := make(map[int]*weewarv1.DamageRange)

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
								dmgFloat := float64(dmg)

								// Only add if we haven't seen this damage value before
								if _, exists := damageMap[dmg]; !exists {
									damageMap[dmg] = &weewarv1.DamageRange{
										MinValue:    dmgFloat,
										MaxValue:    dmgFloat,
										Probability: prob / 100.0, // Convert percentage to decimal
									}

									if damage.MinDamage == 0 || dmgFloat < damage.MinDamage {
										damage.MinDamage = dmgFloat
									}
									if dmgFloat > damage.MaxDamage {
										damage.MaxDamage = dmgFloat
									}
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

	if len(damageMap) == 0 {
		return nil // No damage data found
	}

	// Convert map to sorted slice by damage value
	for dmg := int(damage.MinDamage); dmg <= int(damage.MaxDamage); dmg++ {
		if damageRange, exists := damageMap[dmg]; exists {
			damage.Ranges = append(damage.Ranges, damageRange)
		}
	}

	return damage
}
