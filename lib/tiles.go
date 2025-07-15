package weewar

// TerrainData represents terrain type information
type TerrainData struct {
	ID           int
	Name         string
	MoveCost     int
	DefenseBonus int
}

// Tile represents a single hex tile on the map
type Tile struct {
	Coord CubeCoord `json:"coord"`

	TileType int `json:"tileType"` // Reference to TerrainData by ID

	// Optional: Unit occupying this tile
	Unit *Unit `json:"unit"`
}

// NewTile creates a new tile at the specified position
func NewTile(coord CubeCoord, tileType int) *Tile {
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
