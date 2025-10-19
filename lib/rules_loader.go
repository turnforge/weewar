package weewar

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// LoadRulesEngineFromFile loads a RulesEngine from a canonical rules JSON file
func LoadRulesEngineFromFile(filename string) (*RulesEngine, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules file %s: %w", filename, err)
	}

	return LoadRulesEngineFromJSON(data)
}

// LoadRulesEngineFromJSON loads a RulesEngine from JSON bytes with proper proto field handling
func LoadRulesEngineFromJSON(jsonData []byte) (*RulesEngine, error) {
	// Parse the raw JSON structure first
	var rawData map[string]interface{}
	if err := json.Unmarshal(jsonData, &rawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw JSON: %w", err)
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

	// Load UnitUnitProperties (centralized combat interactions)
	if unitUnitPropsData, ok := rawData["unitUnitProperties"].(map[string]interface{}); ok {
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
				rulesEngine.UnitUnitProperties[key] = props
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
