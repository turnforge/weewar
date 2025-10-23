package services

// TerrainType represents whether terrain is nature or player-controllable
type TerrainType int32

const (
	TerrainNature TerrainType = iota // Natural terrain (grass, mountains, water, etc.)
	TerrainPlayer                    // Player-controllable structures (bases, cities, etc.)
)

// TerrainData represents terrain type information
type TerrainData struct {
	ID           int32       // `json:"id"`
	Name         string      // `json:"name"`
	BaseMoveCost float64     // `json:"baseMoveCost"` // Base movement cost for this terrain
	DefenseBonus float64     // `json:"defenseBonus"`
	Type         TerrainType // `json:"type"` // Nature or Player terrain
	Properties   []string    // `json:"properties,omitempty"`
	// Note: Unit-specific movement costs in RulesEngine can override base cost
}
