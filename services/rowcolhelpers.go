package services

// =============================================================================
// Helper methods to convert row/col to and from Q/R
// Note all game/map/world methods should be PURELY USING Q/R coords.
// These helpers are only when showing debug info or info to UI to players
// =============================================================================

// NumRows returns the number of rows in the map (calculated from bounds)
func (m *World) NumRows() int {
	if m.minR > m.maxR {
		return 0 // Empty map
	}
	return m.maxR - m.minR + 1
}

// NumCols returns the number of columns in the map (calculated from bounds)
func (m *World) NumCols() int {
	if m.minQ > m.maxQ {
		return 0 // Empty map
	}
	return m.maxQ - m.minQ + 1
}
