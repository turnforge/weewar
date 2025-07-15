package weewar

// =============================================================================
// Helper methods to convert row/col to and from Q/R
// Note all game/map/world methods should be PURELY USING Q/R coords.
// These helpers are only when showing debug info or info to UI to players
// =============================================================================

// NumRows returns the number of rows in the map (calculated from bounds)
func (m *Map) NumRows() int {
	if m.minR > m.maxR {
		return 0 // Empty map
	}
	return m.maxR - m.minR + 1
}

// NumCols returns the number of columns in the map (calculated from bounds)
func (m *Map) NumCols() int {
	if m.minQ > m.maxQ {
		return 0 // Empty map
	}
	return m.maxQ - m.minQ + 1
}

// HexToRowCol converts cube coordinates to display coordinates (row, col)
// Uses a standard hex-to-array conversion (odd-row offset style)
func (m *Map) HexToRowCol(coord CubeCoord) (row, col int) {
	row = coord.R
	col = coord.Q + (coord.R+(coord.R&1))/2
	return row, col
}

// RowColToHex converts display coordinates (row, col) to cube coordinates
// Uses a standard array-to-hex conversion (odd-row offset style)
func (m *Map) RowColToHex(row, col int) CubeCoord {
	q := col - (row+(row&1))/2
	return NewCubeCoord(q, row)
}
