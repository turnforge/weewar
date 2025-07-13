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
