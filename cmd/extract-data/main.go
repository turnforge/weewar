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

	"golang.org/x/net/html"
)

type UnitData struct {
	ID              int                           `json:"id"`
	Name            string                        `json:"name"`
	TerrainMovement map[string]float64            `json:"terrainMovement"`
	AttackMatrix    map[string]DamageDistribution `json:"attackMatrix"`
	BaseStats       UnitStats                     `json:"baseStats"`
}

type DamageDistribution struct {
	MinDamage     int             `json:"minDamage"`
	MaxDamage     int             `json:"maxDamage"`
	Probabilities map[int]float64 `json:"probabilities"`
}

type UnitStats struct {
	Cost       int  `json:"cost"`
	Health     int  `json:"health"`
	Movement   int  `json:"movement"`
	Attack     int  `json:"attack"`
	Defense    int  `json:"defense"`
	SightRange int  `json:"sightRange"`
	CanCapture bool `json:"canCapture"`
}

type TerrainData struct {
	ID           int                `json:"id"`
	Name         string             `json:"name"`
	MovementCost map[string]float64 `json:"movementCost"`
	DefenseBonus int                `json:"defenseBonus"`
	Properties   []string           `json:"properties"`
}

type WeeWarData struct {
	Units    []UnitData    `json:"units"`
	Terrains []TerrainData `json:"terrains"`
	Metadata struct {
		Version     string `json:"version"`
		ExtractedAt string `json:"extractedAt"`
		TotalUnits  int    `json:"totalUnits"`
		TotalTiles  int    `json:"totalTiles"`
	} `json:"metadata"`
}

func main() {
	log.Println("Starting WeeWar data extraction...")

	data := WeeWarData{
		Units:    make([]UnitData, 0),
		Terrains: make([]TerrainData, 0),
	}

	// Extract unit data
	log.Println("Extracting unit data...")
	for i := 1; i <= 44; i++ {
		unit, err := extractUnitData(i)
		if err != nil {
			log.Printf("Error extracting unit %d: %v", i, err)
			continue
		}
		data.Units = append(data.Units, unit)
	}

	// Extract terrain data
	log.Println("Extracting terrain data...")
	for i := 1; i <= 26; i++ {
		terrain, err := extractTerrainData(i)
		if err != nil {
			log.Printf("Error extracting terrain %d: %v", i, err)
			continue
		}
		data.Terrains = append(data.Terrains, terrain)
	}

	// Set metadata
	data.Metadata.Version = "1.0"
	data.Metadata.ExtractedAt = "2025-01-08"
	data.Metadata.TotalUnits = len(data.Units)
	data.Metadata.TotalTiles = len(data.Terrains)

	// Save to JSON
	outputPath := "weewar-data.json"
	if err := saveToJSON(data, outputPath); err != nil {
		log.Fatalf("Error saving data: %v", err)
	}

	log.Printf("Successfully extracted data for %d units and %d terrains", len(data.Units), len(data.Terrains))
	log.Printf("Data saved to %s", outputPath)
}

func extractUnitData(unitID int) (UnitData, error) {
	filePath := fmt.Sprintf("data/Units/%d.html", unitID)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return UnitData{}, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	doc, err := html.Parse(strings.NewReader(string(content)))
	if err != nil {
		return UnitData{}, fmt.Errorf("failed to parse HTML: %w", err)
	}

	unit := UnitData{
		ID:              unitID,
		TerrainMovement: make(map[string]float64),
		AttackMatrix:    make(map[string]DamageDistribution),
		BaseStats:       UnitStats{Health: 100}, // Default health
	}

	// Extract unit name from title
	unit.Name = extractUnitName(doc)

	// Extract terrain movement data
	extractTerrainMovement(doc, &unit)

	// Extract attack matrix data
	extractAttackMatrix(doc, &unit)

	return unit, nil
}

func extractTerrainData(tileID int) (TerrainData, error) {
	filePath := fmt.Sprintf("data/Tiles/%d.html", tileID)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return TerrainData{}, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	doc, err := html.Parse(strings.NewReader(string(content)))
	if err != nil {
		return TerrainData{}, fmt.Errorf("failed to parse HTML: %w", err)
	}

	terrain := TerrainData{
		ID:           tileID,
		MovementCost: make(map[string]float64),
		Properties:   make([]string, 0),
	}

	// Extract terrain name from title
	terrain.Name = extractTerrainName(doc)

	// Extract movement cost data
	extractTerrainMovementCosts(doc, &terrain)

	return terrain, nil
}

func extractUnitName(doc *html.Node) string {
	// Find the title element
	title := findElement(doc, "title")
	if title != nil && title.FirstChild != nil {
		titleText := title.FirstChild.Data
		// Extract unit name from title like "Artillery (Mega) - Units - Tiny Attack"
		parts := strings.Split(titleText, " - ")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	return "Unknown Unit"
}

func extractTerrainName(doc *html.Node) string {
	// Find the title element
	title := findElement(doc, "title")
	if title != nil && title.FirstChild != nil {
		titleText := title.FirstChild.Data
		// Extract terrain name from title like "Forest - Tiles - Tiny Attack"
		parts := strings.Split(titleText, " - ")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	return "Unknown Terrain"
}

func extractTerrainMovement(doc *html.Node, unit *UnitData) {
	// Find tables containing terrain movement data
	tables := findAllElements(doc, "table")

	for _, table := range tables {
		rows := findAllElements(table, "tr")
		for _, row := range rows {
			cells := findAllElements(row, "td")
			if len(cells) >= 4 {
				// Look for terrain name in first cell
				terrainName := extractTerrainNameFromCell(cells[0])
				if terrainName != "" {
					// Extract movement cost from appropriate cell
					if movementCost := extractMovementCost(cells); movementCost != 0 {
						unit.TerrainMovement[terrainName] = movementCost
					}
				}
			}
		}
	}
}

func extractAttackMatrix(doc *html.Node, unit *UnitData) {
	// Find attack damage section
	cards := findAllElementsWithClass(doc, "div", "card")

	for _, card := range cards {
		// Extract target unit name from card header
		targetUnit := extractTargetUnitFromCard(card)
		if targetUnit != "" {
			// Extract damage distribution
			if dist := extractDamageDistribution(card); dist.MinDamage > 0 {
				unit.AttackMatrix[targetUnit] = dist
			}
		}
	}
}

func extractTerrainMovementCosts(doc *html.Node, terrain *TerrainData) {
	// Find tables containing movement cost data for different unit types
	tables := findAllElements(doc, "table")

	for _, table := range tables {
		rows := findAllElements(table, "tr")
		for _, row := range rows {
			cells := findAllElements(row, "td")
			if len(cells) >= 2 {
				// Extract unit type and movement cost
				unitType := extractUnitTypeFromCell(cells[0])
				if unitType != "" {
					if cost := extractMovementCostFromCell(cells[1]); cost != 0 {
						terrain.MovementCost[unitType] = cost
					}
				}
			}
		}
	}
}

func extractTerrainNameFromCell(cell *html.Node) string {
	// Look for strong tag with terrain name
	strong := findElement(cell, "strong")
	if strong != nil && strong.FirstChild != nil {
		link := findElement(strong, "a")
		if link != nil && link.FirstChild != nil {
			return strings.TrimSpace(link.FirstChild.Data)
		}
	}
	return ""
}

func extractMovementCost(cells []*html.Node) float64 {
	// Movement cost is typically in the 4th cell (index 3)
	if len(cells) > 3 {
		text := extractTextFromNode(cells[3])
		// Parse numbers like "1.25", "2", "-1" (impassable)
		if num, err := strconv.ParseFloat(strings.TrimSpace(text), 64); err == nil {
			return num
		}
	}
	return 0
}

func extractTargetUnitFromCard(card *html.Node) string {
	// Find card header with unit name
	header := findElementWithClass(card, "div", "card-header")
	if header != nil {
		strong := findElement(header, "strong")
		if strong != nil {
			link := findElement(strong, "a")
			if link != nil && link.FirstChild != nil {
				return strings.TrimSpace(link.FirstChild.Data)
			}
		}
	}
	return ""
}

func extractDamageDistribution(card *html.Node) DamageDistribution {
	dist := DamageDistribution{
		Probabilities: make(map[int]float64),
	}

	// Find damage probability data in tooltips
	body := findElementWithClass(card, "div", "card-body")
	if body != nil {
		tips := findAllElementsWithClass(body, "div", "tip")
		for _, tip := range tips {
			if damageInfo := extractDamageInfo(tip); damageInfo.damage > 0 {
				dist.Probabilities[damageInfo.damage] = damageInfo.probability

				if dist.MinDamage == 0 || damageInfo.damage < dist.MinDamage {
					dist.MinDamage = damageInfo.damage
				}
				if damageInfo.damage > dist.MaxDamage {
					dist.MaxDamage = damageInfo.damage
				}
			}
		}
	}

	return dist
}

type damageInfo struct {
	damage      int
	probability float64
}

func extractDamageInfo(tip *html.Node) damageInfo {
	// Extract from aria-label attribute like "39.9% of the time Artillery (Mega) deals 5 damage to Aircraft Carrier."
	for _, attr := range tip.Attr {
		if attr.Key == "aria-label" {
			return parseDamageInfo(attr.Val)
		}
	}
	return damageInfo{}
}

func parseDamageInfo(text string) damageInfo {
	// Parse text like "39.9% of the time Artillery (Mega) deals 5 damage to Aircraft Carrier."
	percentRegex := regexp.MustCompile(`(\d+\.?\d*)%`)
	damageRegex := regexp.MustCompile(`deals (\d+) damage`)

	var info damageInfo

	if matches := percentRegex.FindStringSubmatch(text); len(matches) > 1 {
		if percent, err := strconv.ParseFloat(matches[1], 64); err == nil {
			info.probability = percent / 100.0
		}
	}

	if matches := damageRegex.FindStringSubmatch(text); len(matches) > 1 {
		if damage, err := strconv.Atoi(matches[1]); err == nil {
			info.damage = damage
		}
	}

	return info
}

func extractUnitTypeFromCell(cell *html.Node) string {
	return extractTextFromNode(cell)
}

func extractMovementCostFromCell(cell *html.Node) float64 {
	text := extractTextFromNode(cell)
	if num, err := strconv.ParseFloat(strings.TrimSpace(text), 64); err == nil {
		return num
	}
	return 0
}

func extractTextFromNode(node *html.Node) string {
	if node == nil {
		return ""
	}

	var text strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			text.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(node)

	return strings.TrimSpace(text.String())
}

func findElement(node *html.Node, tag string) *html.Node {
	if node == nil {
		return nil
	}

	if node.Type == html.ElementNode && node.Data == tag {
		return node
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if found := findElement(c, tag); found != nil {
			return found
		}
	}

	return nil
}

func findAllElements(node *html.Node, tag string) []*html.Node {
	var elements []*html.Node

	if node == nil {
		return elements
	}

	if node.Type == html.ElementNode && node.Data == tag {
		elements = append(elements, node)
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		elements = append(elements, findAllElements(c, tag)...)
	}

	return elements
}

func findElementWithClass(node *html.Node, tag, class string) *html.Node {
	if node == nil {
		return nil
	}

	if node.Type == html.ElementNode && node.Data == tag {
		for _, attr := range node.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, class) {
				return node
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if found := findElementWithClass(c, tag, class); found != nil {
			return found
		}
	}

	return nil
}

func findAllElementsWithClass(node *html.Node, tag, class string) []*html.Node {
	var elements []*html.Node

	if node == nil {
		return elements
	}

	if node.Type == html.ElementNode && node.Data == tag {
		for _, attr := range node.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, class) {
				elements = append(elements, node)
				break
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		elements = append(elements, findAllElementsWithClass(c, tag, class)...)
	}

	return elements
}

func saveToJSON(data WeeWarData, filename string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Marshal to JSON with pretty formatting
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	return nil
}
