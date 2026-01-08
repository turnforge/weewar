package lib

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	"google.golang.org/protobuf/encoding/protojson"
)

const DefaultStartingCoins = 300
const DefaultPerTurnIncome = 300
const DefaultLandbaseIncome = 100
const DefaultNavalbaseIncome = 150
const DefaultAirportbaseIncome = 200
const DefaultMissilesiloIncome = 300
const DefaultMinesIncome = 500

// Tile type constants
const (
	TileTypeLandBase    = 1
	TileTypeNavalBase   = 2
	TileTypeAirport     = 3
	TileTypeDesert      = 4  // Desert terrain (cost 1.75 for infantry)
	TileTypeGrass       = 5  // Basic traversable terrain (cost 1.0 for infantry)
	TileTypeMissileSilo = 16
	TileTypeMines       = 20
)

// Unit type constants
const (
	UnitTypeSoldier         = 1  // Infantry/Trooper - can capture bases
	UnitTypeMedic           = 27 // Medic - can fix friendly units
	UnitTypeStratotanker    = 28 // Stratotanker - can fix air units
	UnitTypeEngineer        = 29 // Engineer - can capture and fix
	UnitTypeTugboat         = 31 // Tugboat - can fix naval units
	UnitTypeAircraftCarrier = 39 // Aircraft Carrier - can fix air units
)

// Default fix values for units that can repair other units
// These are used if fix_value is not specified in the rules JSON
// Formula: p = 0.05 * fix_value, giving probability of each roll succeeding
var DefaultFixValues = map[int32]int32{
	UnitTypeMedic:           10, // p=0.5: Expected ~5 health per fix action
	UnitTypeStratotanker:    10, // p=0.5: Expected ~5 health per fix action
	UnitTypeEngineer:        10, // p=0.5: Expected ~5 health per fix action
	UnitTypeTugboat:         10, // p=0.5: Expected ~5 health per fix action
	UnitTypeAircraftCarrier: 10, // p=0.5: Expected ~5 health per fix action
}

// DefaultFixTargets maps fixer unit ID -> list of target unit_terrain types it can fix
// Based on game rules:
// - Medic (27): Land unit, fixes Land units
// - Stratotanker (28): Air unit, fixes Air units
// - Engineer (29): Land unit, fixes Land units
// - Tugboat (31): Water unit, fixes Water units
// - Aircraft Carrier (39): Water unit, fixes Air units (special case)
var DefaultFixTargets = map[int32][]string{
	UnitTypeMedic:           {"Land"},
	UnitTypeStratotanker:    {"Air"},
	UnitTypeEngineer:        {"Land"},
	UnitTypeTugboat:         {"Water"},
	UnitTypeAircraftCarrier: {"Air"},
}

// Default Income available from various tile types if this is not already in our rules data json
// All other tiles do not generate income
var DefaultIncomeMap = map[int32]int32{
	TileTypeLandBase:    DefaultLandbaseIncome,
	TileTypeNavalBase:   DefaultNavalbaseIncome,
	TileTypeAirport:     DefaultAirportbaseIncome,
	TileTypeMissileSilo: DefaultMissilesiloIncome,
	TileTypeMines:       DefaultMinesIncome,
}

// GetTileIncomeFromConfig returns the income for a tile type using the provided IncomeConfig.
// Falls back to DefaultIncomeMap if IncomeConfig is nil or doesn't have a value for this tile type.
func GetTileIncomeFromConfig(tileType int32, incomeConfig *v1.IncomeConfig) int32 {
	if incomeConfig != nil {
		switch tileType {
		case TileTypeLandBase:
			if incomeConfig.LandbaseIncome > 0 {
				return incomeConfig.LandbaseIncome
			}
		case TileTypeNavalBase:
			if incomeConfig.NavalbaseIncome > 0 {
				return incomeConfig.NavalbaseIncome
			}
		case TileTypeAirport:
			if incomeConfig.AirportbaseIncome > 0 {
				return incomeConfig.AirportbaseIncome
			}
		case TileTypeMissileSilo:
			if incomeConfig.MissilesiloIncome > 0 {
				return incomeConfig.MissilesiloIncome
			}
		case TileTypeMines:
			if incomeConfig.MinesIncome > 0 {
				return incomeConfig.MinesIncome
			}
		}
	}

	// Fall back to default income map
	if income, ok := DefaultIncomeMap[tileType]; ok {
		return income
	}

	return 0
}

// CalculatePlayerBaseIncome calculates the total base income for a player based on their owned tiles.
// This is used during game creation to give players their initial base income.
func CalculatePlayerBaseIncome(playerId int32, worldData *v1.WorldData, incomeConfig *v1.IncomeConfig) int32 {
	totalIncome := int32(0)

	// Calculate income from owned tiles
	for _, tile := range worldData.TilesMap {
		if tile.Player == playerId {
			tileIncome := GetTileIncomeFromConfig(tile.TileType, incomeConfig)
			totalIncome += tileIncome
		}
	}

	// Add base game income if configured
	if incomeConfig != nil && incomeConfig.GameIncome > 0 {
		totalIncome += incomeConfig.GameIncome
	}

	return totalIncome
}

// LoadRulesEngineFromFile loads a RulesEngine from separate rules and damage JSON files
// damageFilename can be empty string if damage distributions are not needed
func LoadRulesEngineFromFile(rulesFilename string, damageFilename string) (*RulesEngine, error) {
	rulesData, err := os.ReadFile(rulesFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules file %s: %w", rulesFilename, err)
	}

	var damageData []byte
	if damageFilename != "" {
		damageData, err = os.ReadFile(damageFilename)
		if err != nil {
			return nil, fmt.Errorf("failed to read damage file %s: %w", damageFilename, err)
		}
	}

	return LoadRulesEngineFromJSON(rulesData, damageData)
}

// LoadRulesEngineFromJSON loads a RulesEngine from separate rules and damage JSON bytes
// rulesJSON contains units, terrains, and terrainUnitProperties
// damageJSON contains unitUnitProperties (combat damage distributions)
// If damageJSON is nil, damage distributions won't be loaded (useful for minimal setups)
func LoadRulesEngineFromJSON(rulesJSON []byte, damageJSON []byte) (*RulesEngine, error) {
	// Parse the rules JSON structure first
	var rawData map[string]any
	if err := json.Unmarshal(rulesJSON, &rawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rules JSON: %w", err)
	}

	rulesEngine := &RulesEngine{
		RulesEngine: &v1.RulesEngine{
			Units:                 make(map[int32]*v1.UnitDefinition),
			Terrains:              make(map[int32]*v1.TerrainDefinition),
			TerrainUnitProperties: make(map[string]*v1.TerrainUnitProperties),
			UnitUnitProperties:    make(map[string]*v1.UnitUnitProperties),
			TerrainTypes:          make(map[int32]v1.TerrainType),
		},
	}

	// Load terrain types (city, nature, bridge, water, road)
	if terrainTypesData, ok := rawData["terrainTypes"].(map[string]any); ok {
		for idStr, typeStr := range terrainTypesData {
			id, err := strconv.ParseInt(idStr, 10, 32)
			if err != nil {
				continue
			}
			if typeName, ok := typeStr.(string); ok {
				rulesEngine.TerrainTypes[int32(id)] = parseTerrainType(typeName)
			}
		}
	}

	// Load terrains using protojson for proper field handling
	if terrainData, ok := rawData["terrains"].(map[string]any); ok {
		for idStr, terrainJson := range terrainData {
			id, err := strconv.ParseInt(idStr, 10, 32)
			if err != nil {
				continue // Skip invalid IDs
			}

			// Marshal back to JSON bytes for protojson.Unmarshal
			terrainBytes, err := json.Marshal(terrainJson)
			if err != nil {
				continue
			}

			terrain := &v1.TerrainDefinition{}
			unmarshaler := protojson.UnmarshalOptions{
				DiscardUnknown: true,
			}
			if err := unmarshaler.Unmarshal(terrainBytes, terrain); err != nil {
				return nil, fmt.Errorf("failed to unmarshal terrain %d: %w", id, err)
			}

			rulesEngine.Terrains[int32(id)] = terrain
		}
	}

	// Load units using protojson for proper field handling
	if unitData, ok := rawData["units"].(map[string]any); ok {
		for idStr, unitJson := range unitData {
			id, err := strconv.ParseInt(idStr, 10, 32)
			if err != nil {
				continue // Skip invalid IDs
			}

			// Marshal back to JSON bytes for protojson.Unmarshal
			unitBytes, err := json.Marshal(unitJson)
			if err != nil {
				continue
			}

			unit := &v1.UnitDefinition{}
			unmarshaler := protojson.UnmarshalOptions{
				DiscardUnknown: true,
			}
			if err := unmarshaler.Unmarshal(unitBytes, unit); err != nil {
				return nil, fmt.Errorf("failed to unmarshal unit %d: %w", id, err)
			}

			rulesEngine.Units[int32(id)] = unit
		}
	}

	// Load TerrainUnitProperties (centralized movement costs and terrain interactions)
	if terrainUnitPropsData, ok := rawData["terrainUnitProperties"].(map[string]any); ok {
		for key, propRaw := range terrainUnitPropsData {
			propBytes, err := json.Marshal(propRaw)
			if err != nil {
				continue
			}

			props := &v1.TerrainUnitProperties{}
			unmarshaler := protojson.UnmarshalOptions{
				DiscardUnknown: true,
			}
			if err := unmarshaler.Unmarshal(propBytes, props); err == nil {
				rulesEngine.TerrainUnitProperties[key] = props
			}
		}
	}

	// Load UnitUnitProperties (centralized combat interactions) from separate damage JSON
	if len(damageJSON) > 0 {
		var damageData map[string]any
		if err := json.Unmarshal(damageJSON, &damageData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal damage JSON: %w", err)
		}

		if unitUnitPropsData, ok := damageData["unitUnitProperties"].(map[string]any); ok {
			for key, propRaw := range unitUnitPropsData {
				propBytes, err := json.Marshal(propRaw)
				if err != nil {
					continue
				}

				props := &v1.UnitUnitProperties{}
				unmarshaler := protojson.UnmarshalOptions{
					DiscardUnknown: true,
				}
				if err := unmarshaler.Unmarshal(propBytes, props); err == nil {
					// Deduplicate damage ranges (source data may have duplicates)
					deduplicateDamageRanges(props.Damage)
					// Calculate expected damage from distribution
					calculateExpectedDamage(props.Damage)
					rulesEngine.UnitUnitProperties[key] = props
				}
			}
		}
	}

	// Set default income values for terrains
	SetDefaultIncomeValues(rulesEngine)

	// Set default fix values for repair units
	SetDefaultFixValues(rulesEngine)

	// Populate reference maps from centralized properties for fast lookup
	rulesEngine.PopulateReferenceMaps()

	// Validate the loaded data
	if err := rulesEngine.ValidateRules(); err != nil {
		return nil, fmt.Errorf("invalid rules data: %w", err)
	}

	return rulesEngine, nil
}

// SaveRulesEngineToFile saves a RulesEngine to a JSON file
func SaveRulesEngineToFile(rulesEngine *RulesEngine, filename string) error {
	data, err := json.MarshalIndent(rulesEngine, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal rules engine: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write rules file %s: %w", filename, err)
	}

	return nil
}

// deduplicateDamageRanges removes duplicate damage values from DamageDistribution
// This handles legacy data files that may have duplicate ranges
func deduplicateDamageRanges(damage *v1.DamageDistribution) {
	if damage == nil || len(damage.Ranges) == 0 {
		return
	}

	seenDamageValues := make(map[float64]bool)
	uniqueRanges := make([]*v1.DamageRange, 0, len(damage.Ranges))

	for _, damageRange := range damage.Ranges {
		// Skip empty ranges or ranges without valid damage values
		if damageRange == nil {
			continue
		}

		damageValue := damageRange.MinValue

		// Only add if we haven't seen this damage value before
		if !seenDamageValues[damageValue] {
			seenDamageValues[damageValue] = true
			uniqueRanges = append(uniqueRanges, damageRange)
		}
	}

	damage.Ranges = uniqueRanges
}

// calculateExpectedDamage calculates and sets the expected damage from the distribution ranges
func calculateExpectedDamage(damage *v1.DamageDistribution) {
	if damage == nil || len(damage.Ranges) == 0 {
		return
	}

	expectedDamage := 0.0
	totalProbability := 0.0

	// Calculate expected value: sum of (damage * probability) for each range
	for _, damageRange := range damage.Ranges {
		if damageRange == nil {
			continue
		}

		// For single-value ranges (min == max), use that value
		// For ranges, use the midpoint
		avgDamage := (damageRange.MinValue + damageRange.MaxValue) / 2.0
		probability := damageRange.Probability

		expectedDamage += avgDamage * probability
		totalProbability += probability
	}

	// Normalize if probabilities don't sum to 1.0
	if totalProbability > 0 && totalProbability != 1.0 {
		expectedDamage = expectedDamage / totalProbability
	}

	damage.ExpectedDamage = expectedDamage
}

// SetDefaultIncomeValues sets default income_per_turn values for terrain types using DefaultIncomeMap
func SetDefaultIncomeValues(re *RulesEngine) {
	for tileID, terrain := range re.Terrains {
		// Check if this tile ID has a default income value
		if income, ok := DefaultIncomeMap[tileID]; ok {
			terrain.IncomePerTurn = income
		}
	}
}

// SetDefaultFixValues sets default fix_value for units that can repair other units
// Only sets the value if it's not already specified in the loaded data
func SetDefaultFixValues(re *RulesEngine) {
	for unitID, unit := range re.Units {
		// Only set if fix_value is not already defined (0 means unset)
		if unit.FixValue == 0 {
			if fixValue, ok := DefaultFixValues[unitID]; ok {
				unit.FixValue = fixValue
			}
		}
	}
}

// parseTerrainType converts a string terrain type name to the proto enum
func parseTerrainType(typeName string) v1.TerrainType {
	switch typeName {
	case "city":
		return v1.TerrainType_TERRAIN_TYPE_CITY
	case "nature":
		return v1.TerrainType_TERRAIN_TYPE_NATURE
	case "bridge":
		return v1.TerrainType_TERRAIN_TYPE_BRIDGE
	case "water":
		return v1.TerrainType_TERRAIN_TYPE_WATER
	case "road":
		return v1.TerrainType_TERRAIN_TYPE_ROAD
	default:
		return v1.TerrainType_TERRAIN_TYPE_UNSPECIFIED
	}
}
