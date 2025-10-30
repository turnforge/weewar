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

	"github.com/antchfx/htmlquery"
	weewarv1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"golang.org/x/net/html"
)

// RulesData represents the core rules data structure (without damage distributions)
type RulesData struct {
	Units                 map[string]*weewarv1.UnitDefinition        `json:"units"`
	Terrains              map[string]*weewarv1.TerrainDefinition     `json:"terrains"`
	TerrainUnitProperties map[string]*weewarv1.TerrainUnitProperties `json:"terrainUnitProperties"`
}

// DamageData represents combat damage distributions (kept separate due to size)
type DamageData struct {
	UnitUnitProperties map[string]*weewarv1.UnitUnitProperties `json:"unitUnitProperties"`
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

	// Initialize rules data (core rules and damage distributions separately)
	rulesData := RulesData{
		Units:                 make(map[string]*weewarv1.UnitDefinition),
		Terrains:              make(map[string]*weewarv1.TerrainDefinition),
		TerrainUnitProperties: make(map[string]*weewarv1.TerrainUnitProperties),
	}
	damageData := DamageData{
		UnitUnitProperties: make(map[string]*weewarv1.UnitUnitProperties),
	}

	// Extract terrain data
	fmt.Println("Extracting terrain data...")
	if err := extractTerrainsData(tilesDir, &rulesData); err != nil {
		fmt.Printf("Error extracting terrain data: %v\n", err)
		os.Exit(1)
	}

	// Extract unit data
	fmt.Println("Extracting unit data...")
	if err := extractUnitsData(unitsDir, &rulesData, &damageData); err != nil {
		fmt.Printf("Error extracting unit data: %v\n", err)
		os.Exit(1)
	}

	// Write core rules output
	rulesPath := "weewar-rules.json"
	rulesJSON, err := json.MarshalIndent(rulesData, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling rules JSON: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(rulesPath, rulesJSON, 0644); err != nil {
		fmt.Printf("Error writing rules file: %v\n", err)
		os.Exit(1)
	}

	// Write damage distributions output
	damagePath := "weewar-damage.json"
	damageJSON, err := json.MarshalIndent(damageData, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling damage JSON: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(damagePath, damageJSON, 0644); err != nil {
		fmt.Printf("Error writing damage file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nExtraction complete! Generated 2 files:\n")
	fmt.Printf("- %s (core rules)\n", rulesPath)
	fmt.Printf("  - %d units\n", len(rulesData.Units))
	fmt.Printf("  - %d terrains\n", len(rulesData.Terrains))
	fmt.Printf("  - %d terrain-unit properties\n", len(rulesData.TerrainUnitProperties))
	fmt.Printf("- %s (damage distributions)\n", damagePath)
	fmt.Printf("  - %d unit-unit combat properties\n", len(damageData.UnitUnitProperties))
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

func extractUnitsData(unitsDir string, rulesData *RulesData, damageData *DamageData) error {
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

		// Extract unit-unit combat properties (damage distributions)
		if err := extractUnitCombatProperties(doc, int32(unitID), damageData); err != nil {
			return fmt.Errorf("error extracting combat properties for %s: %v", file, err)
		}
	}

	return nil
}

func extractTerrainName(doc *html.Node) string {
	// Use XPath to find h1 with class containing "mb-3"
	node := htmlquery.FindOne(doc, "//h1[contains(@class, 'mb-3')]")
	if node == nil {
		return ""
	}

	name := getTextContent(node)
	// Clean up name
	name = strings.TrimSpace(name)
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")
	return name
}

func extractTerrainUnitInteractions(doc *html.Node, terrainID int32, rulesData *RulesData) error {
	// Use XPath to find column headers
	var columnHeaders []string
	headerCells := htmlquery.Find(doc, "//thead/tr/th")
	for _, cell := range headerCells {
		headerText := getTextContent(cell)
		columnHeaders = append(columnHeaders, strings.ToLower(strings.TrimSpace(headerText)))
	}

	// Use XPath to find all tbody rows
	rows := htmlquery.Find(doc, "//tbody/tr")
	for _, row := range rows {
		extractUnitRowData(row, terrainID, columnHeaders, rulesData)
	}

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
		Health: 10, // Always 10 in WeeWar
	}

	// Extract name from page title
	if titleNode := htmlquery.FindOne(doc, "//title"); titleNode != nil {
		title := getTextContent(titleNode)
		if strings.Contains(title, " - Units - ") {
			parts := strings.Split(title, " - Units - ")
			if len(parts) > 0 {
				unitDef.Name = strings.TrimSpace(parts[0])
			}
		}
	}

	// Extract Movement points
	if node := htmlquery.FindOne(doc, "//p[strong[contains(text(), 'Movement')] and not(contains(., 'Build Percentage'))]"); node != nil {
		text := getTextContent(node)
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			if strings.Contains(line, "Movement") && i+1 < len(lines) {
				nextLine := strings.TrimSpace(strings.Join(lines[i+1:], "\n"))
				if matches := regexp.MustCompile(`(\d+(?:\.\d+)?)`).FindStringSubmatch(nextLine); len(matches) > 1 {
					unitDef.MovementPoints = parseMovementCost(matches[1])
				}
				break
			}
		}
	}

	// Extract Retreat points
	if node := htmlquery.FindOne(doc, "//p[strong[contains(text(), 'Retreat')]]"); node != nil {
		text := getTextContent(node)
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			if strings.Contains(line, "Retreat") && i+1 < len(lines) {
				nextLine := strings.TrimSpace(strings.Join(lines[i+1:], "\n"))
				if matches := regexp.MustCompile(`(\d+(?:\.\d+)?)`).FindStringSubmatch(nextLine); len(matches) > 1 {
					unitDef.RetreatPoints = parseMovementCost(matches[1])
				}
				break
			}
		}
	}

	// Extract Splash Damage
	if node := htmlquery.FindOne(doc, "//p[strong[contains(text(), 'Splash Damage')]]"); node != nil {
		text := getTextContent(node)
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			if strings.Contains(line, "Splash Damage") && i+1 < len(lines) {
				nextLine := strings.TrimSpace(strings.Join(lines[i+1:], "\n"))
				if matches := regexp.MustCompile(`^\d+`).FindStringSubmatch(nextLine); len(matches) > 0 {
					if val, err := strconv.Atoi(matches[0]); err == nil {
						unitDef.SplashDamage = int32(val)
					}
				}
				break
			}
		}
	}

	// Extract Defense
	if node := htmlquery.FindOne(doc, "//p[strong[contains(text(), 'Defense')]]"); node != nil {
		text := getTextContent(node)
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			if strings.Contains(line, "Defense") && i+1 < len(lines) {
				nextLine := strings.TrimSpace(strings.Join(lines[i+1:], "\n"))
				if matches := regexp.MustCompile(`^\d+`).FindStringSubmatch(nextLine); len(matches) > 0 {
					if val, err := strconv.Atoi(matches[0]); err == nil {
						unitDef.Defense = int32(val)
					}
				}
				break
			}
		}
	}

	// Extract Attack Range
	if node := htmlquery.FindOne(doc, "//p[strong[contains(text(), 'Attack Range')]]"); node != nil {
		text := getTextContent(node)
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			if strings.Contains(line, "Attack Range") && i+1 < len(lines) {
				nextLine := strings.TrimSpace(strings.Join(lines[i+1:], "\n"))
				// Check for range format: "2 - 3"
				if matches := regexp.MustCompile(`^(\d+)\s*-\s*(\d+)`).FindStringSubmatch(nextLine); len(matches) > 2 {
					if minRng, err := strconv.Atoi(matches[1]); err == nil {
						unitDef.MinAttackRange = int32(minRng)
					}
					if maxRng, err := strconv.Atoi(matches[2]); err == nil {
						unitDef.AttackRange = int32(maxRng)
					}
				} else if matches := regexp.MustCompile(`^(\d+)`).FindStringSubmatch(nextLine); len(matches) > 1 {
					// Single value format: "1 (adjacent...)"
					if rng, err := strconv.Atoi(matches[1]); err == nil {
						unitDef.AttackRange = int32(rng)
						unitDef.MinAttackRange = int32(rng) // Min and max are the same
					}
				}
				break
			}
		}
	}

	// Extract Coins
	if node := htmlquery.FindOne(doc, "//p[strong[contains(text(), 'Coins')] and not(contains(., 'Coins Spent'))]"); node != nil {
		text := getTextContent(node)
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

	// Extract unit classification (Type section)
	extractUnitClassification(doc, unitDef)

	// Extract attack table
	extractAttackTable(doc, unitDef)

	// Extract action order (Progression section)
	extractActionOrder(doc, unitDef)

	return unitDef, nil
}

func extractUnitCombatProperties(doc *html.Node, attackerID int32, damageData *DamageData) error {
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
							damageData.UnitUnitProperties[key] = &weewarv1.UnitUnitProperties{
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

// extractUnitClassification extracts the unit type (Light/Heavy/Stealth) and terrain (Air/Land/Water)
func extractUnitClassification(doc *html.Node, unitDef *weewarv1.UnitDefinition) {
	// Use XPath to find the <p> element containing <strong>Type</strong>
	// Format: <p><strong>Type</strong>...<br>Light\t\t\tLand\t\t</p>
	node := htmlquery.FindOne(doc, "//p[strong[text()='Type'] and span[contains(@aria-label, 'influences')]]")
	if node == nil {
		return
	}

	text := getTextContent(node)

	// Try to find "Light", "Heavy", or "Stealth" and "Air", "Land", or "Water"
	classTypes := []string{"Light", "Heavy", "Stealth"}
	terrainTypes := []string{"Air", "Land", "Water"}

	var foundClass, foundTerrain string
	for _, ct := range classTypes {
		if strings.Contains(text, ct) {
			foundClass = ct
			break
		}
	}
	for _, tt := range terrainTypes {
		if strings.Contains(text, tt) {
			foundTerrain = tt
			break
		}
	}

	if foundClass != "" && foundTerrain != "" {
		unitDef.UnitClass = foundClass
		unitDef.UnitTerrain = foundTerrain
		log.Printf("  Unit Classification: %s %s\n", unitDef.UnitClass, unitDef.UnitTerrain)
	}
}

// extractAttackTable extracts the attack table showing base attack values against different unit classes
func extractAttackTable(doc *html.Node, unitDef *weewarv1.UnitDefinition) {
	unitDef.AttackVsClass = make(map[string]int32)

	// Use XPath to find the table following the "Attack" h3 heading
	// XPath: find h3 containing "Attack", then get the following table
	tableNode := htmlquery.FindOne(doc, "//h3[contains(., 'Attack')]/following-sibling::table[1]")
	if tableNode == nil {
		log.Printf("  No attack table found\n")
		return
	}

	// Extract table using our utility
	table := ExtractHtmlTable(tableNode)

	// Parse the attack table
	if !table.HasHeader || len(table.Rows) < 2 {
		log.Printf("  No valid attack table found\n")
		return
	}

	// Get headers from first row
	headers := table.Rows[0]
	log.Printf("  Attack Table Headers: %v\n", headers)

	// Process data rows (skip first row which is headers)
	for i := 1; i < len(table.Rows); i++ {
		row := table.Rows[i]
		if len(row) == 0 {
			continue
		}

		// First cell is the row class (Light/Heavy/Stealth)
		rowClass := row[0]

		// Remaining cells correspond to column headers
		for j := 1; j < len(row) && j < len(headers); j++ {
			cellValue := row[j]
			columnName := headers[j]

			// Skip "n/a" or empty values
			if strings.Contains(cellValue, "n/a") || cellValue == "" {
				continue
			}

			// Parse attack value
			if attackValue, err := strconv.Atoi(cellValue); err == nil {
				key := fmt.Sprintf("%s:%s", rowClass, columnName)
				unitDef.AttackVsClass[key] = int32(attackValue)
				log.Printf("    Attack[%s] = %d\n", key, attackValue)
			}
		}
	}
}

// extractActionOrder extracts the action_order from the Progression section
// The HTML structure is:
//   <p><strong>Progression</strong>...
//     <span class="badge bg-success fw-normal">
//       <span class="tip">move 3</span>
//     </span>
//     <span class="badge bg-success fw-normal">
//       <span class="tip">attack 1</span>
//       OR
//       <span class="tip">capture</span>
//     </span>
//   </p>
// We extract action names and join alternatives with "|" (e.g., "attack|capture")
// Numeric values like "move 3" or "attack 1-2" are ignored (extracted elsewhere)
func extractActionOrder(doc *html.Node, unitDef *weewarv1.UnitDefinition) {
	// Find the <p> element containing <strong>Progression</strong>
	progNode := htmlquery.FindOne(doc, "//p[strong[text()='Progression']]")
	if progNode == nil {
		log.Printf("  No Progression section found\n")
		return
	}

	// Find all badge spans within this paragraph
	badges := htmlquery.Find(progNode, ".//span[contains(@class, 'badge bg-success')]")

	actionOrder := []string{}

	for _, badge := range badges {
		// Get text content of the entire badge
		text := getTextContent(badge)
		text = strings.TrimSpace(text)

		// Normalize whitespace
		text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

		// Check if this badge contains " OR "
		if strings.Contains(text, " OR ") {
			// Split by OR and extract action names
			parts := strings.Split(text, " OR ")
			actionNames := []string{}
			for _, part := range parts {
				actionName := extractActionName(strings.TrimSpace(part))
				if actionName != "" {
					actionNames = append(actionNames, actionName)
				}
			}
			// Join with "|"
			if len(actionNames) > 0 {
				actionOrder = append(actionOrder, strings.Join(actionNames, "|"))
			}
		} else {
			// Single action
			actionName := extractActionName(text)
			if actionName != "" {
				actionOrder = append(actionOrder, actionName)
			}
		}
	}

	unitDef.ActionOrder = actionOrder
	if len(actionOrder) > 0 {
		log.Printf("  Action Order: %v\n", actionOrder)
	}
}

// extractActionName extracts just the action name from text like "move 3" or "attack 1-2"
// Ignores numeric values which are extracted elsewhere (movement points, attack range, etc.)
// Returns the action name in lowercase
func extractActionName(text string) string {
	// Common action names - check in order of specificity
	actions := []string{"capture", "attack", "move", "fix", "build", "repair"}

	text = strings.ToLower(text)

	// Find the first matching action name
	for _, action := range actions {
		if strings.Contains(text, action) {
			return action
		}
	}

	return ""
}
