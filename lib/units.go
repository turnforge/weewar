package weewar

type UnitStats struct {
	Cost       int  `json:"cost"`
	Health     int  `json:"health"`
	Movement   int  `json:"movement"`
	Attack     int  `json:"attack"`
	Defense    int  `json:"defense"`
	SightRange int  `json:"sightRange"`
	CanCapture bool `json:"canCapture"`
}

type UnitData struct {
	BaseStats       UnitStats                     `json:"baseStats"`
	ID              int                           `json:"id"`
	Name            string                        `json:"name"`
	TerrainMovement map[string]float64            `json:"terrainMovement"`
	AttackMatrix    map[string]DamageDistribution `json:"attackMatrix"`
}

type DamageDistribution struct {
	MinDamage     int                `json:"minDamage"`
	MaxDamage     int                `json:"maxDamage"`
	Probabilities map[string]float64 `json:"probabilities"`
}

// Unit represents a runtime unit instance in the game
type Unit struct {
	UnitType int // Reference to UnitData by ID

	// Runtime state
	DistanceLeft    int // Movement points remaining this turn
	AvailableHealth int // Current health points
	TurnCounter     int // Which turn this unit was created/last acted

	// Position on the map
	Coord CubeCoord `json:"coord"` // Cube coordinate position

	// Player ownership
	PlayerID int
}

// GetPosition returns the unit's cube coordinate position
func (u *Unit) GetPosition() CubeCoord {
	return u.Coord
}

// SetPosition sets the unit's position using cube coordinates
func (u *Unit) SetPosition(coord CubeCoord) {
	u.Coord = coord
}

// NewUnit creates a new unit instance
func NewUnit(unitType, playerID int) *Unit {
	return &Unit{
		UnitType:        unitType,
		PlayerID:        playerID,
		DistanceLeft:    0, // Will be set based on UnitData
		AvailableHealth: 0, // Will be set based on UnitData
		TurnCounter:     0,
	}
}

func (u *Unit) Clone() *Unit {
	return &Unit{
		UnitType:        u.UnitType,
		DistanceLeft:    u.DistanceLeft,
		AvailableHealth: u.AvailableHealth,
		TurnCounter:     u.TurnCounter,
		Coord:           u.Coord,
		PlayerID:        u.PlayerID,
	}
}
