package weewar

// TerrainType represents whether terrain is nature or player-controllable
type TerrainType int

const (
	TerrainNature TerrainType = iota // Natural terrain (grass, mountains, water, etc.)
	TerrainPlayer                    // Player-controllable structures (bases, cities, etc.)
)

// TerrainData represents terrain type information
type TerrainData struct {
	ID           int
	Name         string
	MoveCost     int
	DefenseBonus int
	Type         TerrainType // Nature or Player terrain
}

// Tile represents a single hex tile on the map
type Tile struct {
	Coord AxialCoord `json:"coord"`

	TileType int `json:"tileType"` // Reference to TerrainData by ID

	// Optional: Unit occupying this tile
	Unit *Unit `json:"unit"`
}

// NewTile creates a new tile at the specified position
func NewTile(coord AxialCoord, tileType int) *Tile {
	return &Tile{
		Coord:    coord,
		TileType: tileType,
	}
}

func (t *Tile) Clone() *Tile {
	return &Tile{
		Coord:    t.Coord,
		TileType: t.TileType,
		Unit:     nil, // Units are cloned separately
	}
}
