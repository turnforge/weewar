package weewar

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/panyam/turnengine/internal/turnengine"
)

type WeeWarCombatSystem struct {
	unitData     map[string]UnitData
	terrainData  map[string]TerrainData
	damageMatrix map[string]map[string]DamageDistribution
	rng          *rand.Rand
}

type UnitData struct {
	ID              int                           `json:"id"`
	Name            string                        `json:"name"`
	TerrainMovement map[string]float64            `json:"terrainMovement"`
	AttackMatrix    map[string]DamageDistribution `json:"attackMatrix"`
	BaseStats       UnitStats                     `json:"baseStats"`
}

type DamageDistribution struct {
	MinDamage     int               `json:"minDamage"`
	MaxDamage     int               `json:"maxDamage"`
	Probabilities map[string]float64 `json:"probabilities"`
}

type UnitStats struct {
	Cost        int     `json:"cost"`
	Health      int     `json:"health"`
	Movement    int     `json:"movement"`
	Attack      int     `json:"attack"`
	Defense     int     `json:"defense"`
	SightRange  int     `json:"sightRange"`
	CanCapture  bool    `json:"canCapture"`
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

func NewWeeWarCombatSystem() (*WeeWarCombatSystem, error) {
	// Load WeeWar data
	data, err := loadWeeWarData()
	if err != nil {
		return nil, fmt.Errorf("failed to load WeeWar data: %w", err)
	}

	system := &WeeWarCombatSystem{
		unitData:     make(map[string]UnitData),
		terrainData:  make(map[string]TerrainData),
		damageMatrix: make(map[string]map[string]DamageDistribution),
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// Index units by name
	for _, unit := range data.Units {
		system.unitData[unit.Name] = unit
		system.damageMatrix[unit.Name] = unit.AttackMatrix
	}

	// Index terrains by name
	for _, terrain := range data.Terrains {
		system.terrainData[terrain.Name] = terrain
	}

	return system, nil
}

func (wcs *WeeWarCombatSystem) Name() string {
	return "WeeWarCombatSystem"
}

func (wcs *WeeWarCombatSystem) Priority() int {
	return 100
}

func (wcs *WeeWarCombatSystem) Update(world *turnengine.World) error {
	// This system doesn't run on every update, only when combat is initiated
	return nil
}

func (wcs *WeeWarCombatSystem) ResolveCombat(attackerEntity, defenderEntity *turnengine.Entity, board turnengine.Board) (turnengine.CombatResult, error) {
	// Extract unit types
	attackerType, err := wcs.getUnitType(attackerEntity)
	if err != nil {
		return turnengine.CombatResult{}, fmt.Errorf("failed to get attacker unit type: %w", err)
	}

	defenderType, err := wcs.getUnitType(defenderEntity)
	if err != nil {
		return turnengine.CombatResult{}, fmt.Errorf("failed to get defender unit type: %w", err)
	}

	// Get attacker and defender positions
	attackerPos, err := wcs.getEntityPosition(attackerEntity)
	if err != nil {
		return turnengine.CombatResult{}, fmt.Errorf("failed to get attacker position: %w", err)
	}

	defenderPos, err := wcs.getEntityPosition(defenderEntity)
	if err != nil {
		return turnengine.CombatResult{}, fmt.Errorf("failed to get defender position: %w", err)
	}

	// Check if attacker can reach defender
	if !wcs.canAttack(attackerPos, defenderPos, board) {
		return turnengine.CombatResult{}, fmt.Errorf("target out of range")
	}

	// Get attacker's health for damage calculation
	attackerHealth, err := wcs.getEntityHealth(attackerEntity)
	if err != nil {
		return turnengine.CombatResult{}, fmt.Errorf("failed to get attacker health: %w", err)
	}

	defenderHealth, err := wcs.getEntityHealth(defenderEntity)
	if err != nil {
		return turnengine.CombatResult{}, fmt.Errorf("failed to get defender health: %w", err)
	}

	// Calculate damage based on WeeWar damage matrix
	damage := wcs.calculateDamage(attackerType, defenderType, attackerHealth)
	
	// Apply terrain defense bonus
	defenseBonus := wcs.getTerrainDefenseBonus(board, defenderPos)
	finalDamage := wcs.applyDefenseBonus(damage, defenseBonus)

	result := turnengine.CombatResult{
		Hit:            true,
		DefenderDamage: finalDamage,
		DefenderKilled: (defenderHealth - finalDamage) <= 0,
	}

	// Counter-attack if defender survives and can attack back
	if !result.DefenderKilled && wcs.canCounterAttack(defenderType, attackerType, defenderPos, attackerPos, board) {
		counterDamage := wcs.calculateDamage(defenderType, attackerType, defenderHealth-finalDamage)
		result.AttackerDamage = counterDamage
		result.AttackerKilled = (attackerHealth - counterDamage) <= 0
	}

	// Apply damage to entities
	if err := wcs.applyDamage(attackerEntity, result.AttackerDamage); err != nil {
		return result, fmt.Errorf("failed to apply attacker damage: %w", err)
	}

	if err := wcs.applyDamage(defenderEntity, result.DefenderDamage); err != nil {
		return result, fmt.Errorf("failed to apply defender damage: %w", err)
	}

	return result, nil
}

func (wcs *WeeWarCombatSystem) calculateDamage(attackerType, defenderType string, attackerHealth int) int {
	// Get damage distribution for this unit matchup
	damageMatrix, exists := wcs.damageMatrix[attackerType]
	if !exists {
		return 1 // Default minimal damage
	}

	distribution, exists := damageMatrix[defenderType]
	if !exists {
		return 1 // Default minimal damage
	}

	// Sample from probability distribution
	damage := wcs.sampleDamageDistribution(distribution)
	
	// Apply health scaling (damaged units do less damage)
	healthRatio := float64(attackerHealth) / 100.0
	scaledDamage := int(float64(damage) * healthRatio)
	
	if scaledDamage < 1 {
		scaledDamage = 1
	}

	return scaledDamage
}

func (wcs *WeeWarCombatSystem) sampleDamageDistribution(dist DamageDistribution) int {
	if len(dist.Probabilities) == 0 {
		return 1
	}

	// Generate random number
	roll := wcs.rng.Float64()
	
	// Sample from probability distribution
	cumulative := 0.0
	for damageStr, prob := range dist.Probabilities {
		cumulative += prob
		if roll <= cumulative {
			// Convert string key to int
			var damage int
			fmt.Sscanf(damageStr, "%d", &damage)
			return damage
		}
	}
	
	// Fallback to max damage
	return dist.MaxDamage
}

func (wcs *WeeWarCombatSystem) getUnitType(entity *turnengine.Entity) (string, error) {
	unitType, exists := entity.GetComponent("unitType")
	if !exists {
		return "", fmt.Errorf("entity has no unitType component")
	}

	unitTypeName, ok := unitType["unitType"].(string)
	if !ok {
		return "", fmt.Errorf("invalid unitType value")
	}

	return unitTypeName, nil
}

func (wcs *WeeWarCombatSystem) getEntityPosition(entity *turnengine.Entity) (turnengine.Position, error) {
	posComp, exists := entity.GetComponent("position")
	if !exists {
		return nil, fmt.Errorf("entity has no position component")
	}

	q, qOk := posComp["q"].(float64)
	r, rOk := posComp["r"].(float64)
	if !qOk || !rOk {
		return nil, fmt.Errorf("invalid position component")
	}

	return &HexPosition{Q: int(q), R: int(r)}, nil
}

func (wcs *WeeWarCombatSystem) getEntityHealth(entity *turnengine.Entity) (int, error) {
	health, exists := entity.GetComponent("health")
	if !exists {
		return 0, fmt.Errorf("entity has no health component")
	}

	current, ok := health["current"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid health value")
	}

	return int(current), nil
}

func (wcs *WeeWarCombatSystem) canAttack(attackerPos, defenderPos turnengine.Position, board turnengine.Board) bool {
	distance := board.GetDistance(attackerPos, defenderPos)
	return distance == 1 // Most units can only attack adjacent positions
}

func (wcs *WeeWarCombatSystem) canCounterAttack(defenderType, attackerType string, defenderPos, attackerPos turnengine.Position, board turnengine.Board) bool {
	// Check if defender can attack back (adjacent and has attack capability)
	distance := board.GetDistance(defenderPos, attackerPos)
	if distance != 1 {
		return false
	}

	// Check if defender unit has attack capability against attacker
	damageMatrix, exists := wcs.damageMatrix[defenderType]
	if !exists {
		return false
	}

	_, canAttack := damageMatrix[attackerType]
	return canAttack
}

func (wcs *WeeWarCombatSystem) getTerrainDefenseBonus(board turnengine.Board, pos turnengine.Position) int {
	terrain, exists := board.GetTerrain(pos)
	if !exists {
		return 0
	}

	terrainData, exists := wcs.terrainData[terrain]
	if !exists {
		return 0
	}

	return terrainData.DefenseBonus
}

func (wcs *WeeWarCombatSystem) applyDefenseBonus(damage, defenseBonus int) int {
	// Apply defense bonus as damage reduction
	reducedDamage := damage - (defenseBonus / 10) // Simple conversion
	if reducedDamage < 1 {
		reducedDamage = 1
	}
	return reducedDamage
}

func (wcs *WeeWarCombatSystem) applyDamage(entity *turnengine.Entity, damage int) error {
	if damage <= 0 {
		return nil
	}

	health, exists := entity.GetComponent("health")
	if !exists {
		return fmt.Errorf("entity has no health component")
	}

	current, ok := health["current"].(float64)
	if !ok {
		return fmt.Errorf("invalid health current value")
	}

	newHealth := int(current) - damage
	if newHealth < 0 {
		newHealth = 0
	}

	health["current"] = float64(newHealth)
	entity.Components["health"] = health

	return nil
}

func loadWeeWarData() (WeeWarData, error) {
	var data WeeWarData
	
	content, err := ioutil.ReadFile("games/weewar/weewar-data.json")
	if err != nil {
		return data, fmt.Errorf("failed to read weewar-data.json: %w", err)
	}

	if err := json.Unmarshal(content, &data); err != nil {
		return data, fmt.Errorf("failed to unmarshal WeeWar data: %w", err)
	}

	return data, nil
}