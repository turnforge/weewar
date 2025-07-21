package weewar

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadRulesEngineFromFile loads a RulesEngine from a canonical rules JSON file
func LoadRulesEngineFromFile(filename string) (*RulesEngine, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules file %s: %w", filename, err)
	}

	return LoadRulesEngineFromJSON(data)
}

// LoadRulesEngineFromJSON loads a RulesEngine from JSON bytes
func LoadRulesEngineFromJSON(jsonData []byte) (*RulesEngine, error) {
	var rulesEngine RulesEngine
	if err := json.Unmarshal(jsonData, &rulesEngine); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rules data: %w", err)
	}

	// Validate the loaded data
	if err := rulesEngine.ValidateRules(); err != nil {
		return nil, fmt.Errorf("invalid rules data: %w", err)
	}

	return &rulesEngine, nil
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

// CreateGameWithRules creates a new game instance with a loaded RulesEngine
func CreateGameWithRules(world *World, rulesFile string, seed int64) (*Game, error) {
	// Load rules engine
	rulesEngine, err := LoadRulesEngineFromFile(rulesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load rules: %w", err)
	}

	// Create game with rules engine
	game, err := NewGame(world, rulesEngine, seed)
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	return game, nil
}