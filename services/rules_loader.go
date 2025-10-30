package services

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

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
	var rawData map[string]interface{}
	if err := json.Unmarshal(rulesJSON, &rawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rules JSON: %w", err)
	}

	rulesEngine := &RulesEngine{
		RulesEngine: &v1.RulesEngine{
			Units:                 make(map[int32]*v1.UnitDefinition),
			Terrains:              make(map[int32]*v1.TerrainDefinition),
			TerrainUnitProperties: make(map[string]*v1.TerrainUnitProperties),
			UnitUnitProperties:    make(map[string]*v1.UnitUnitProperties),
		},
	}

	// Load terrains using protojson for proper field handling
	if terrainData, ok := rawData["terrains"].(map[string]interface{}); ok {
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
	if unitData, ok := rawData["units"].(map[string]interface{}); ok {
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
	if terrainUnitPropsData, ok := rawData["terrainUnitProperties"].(map[string]interface{}); ok {
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
	if damageJSON != nil && len(damageJSON) > 0 {
		var damageData map[string]interface{}
		if err := json.Unmarshal(damageJSON, &damageData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal damage JSON: %w", err)
		}

		if unitUnitPropsData, ok := damageData["unitUnitProperties"].(map[string]interface{}); ok {
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

	// Populate reference maps from centralized properties for fast lookup
	rulesEngine.PopulateReferenceMaps()

	// Validate the loaded data
	if err := rulesEngine.ValidateRules(); err != nil {
		return nil, fmt.Errorf("invalid rules data: %w", err)
	}

	return rulesEngine, nil
}

// LoadRulesEngineFromLegacy loads a RulesEngine by converting from legacy weewar-data.json format
func LoadRulesEngineFromLegacy(filename string) (*RulesEngine, error) {
	// This would use the conversion logic from the CLI tool
	// For now, return an error suggesting to use the converter first
	return nil, fmt.Errorf("legacy format loading not implemented - use weewar-convert CLI tool first to convert %s to canonical format", filename)
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
