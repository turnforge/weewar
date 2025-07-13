package weewar

import (
	"fmt"
	"math"

	"github.com/panyam/turnengine/internal/turnengine"
)

type HexPosition struct {
	Q, R int `json:"q,r"`
}

func NewHexPosition(q, r int) *HexPosition {
	return &HexPosition{Q: q, R: r}
}

func (hp *HexPosition) String() string {
	return fmt.Sprintf("(%d,%d)", hp.Q, hp.R)
}

func (hp *HexPosition) Equals(other turnengine.Position) bool {
	if otherHex, ok := other.(*HexPosition); ok {
		return hp.Q == otherHex.Q && hp.R == otherHex.R
	}
	return false
}

func (hp *HexPosition) Hash() string {
	return fmt.Sprintf("%d,%d", hp.Q, hp.R)
}

func (hp *HexPosition) S() int {
	return -hp.Q - hp.R
}

func (hp *HexPosition) ToCube() (int, int, int) {
	return hp.Q, hp.R, hp.S()
}

func (hp *HexPosition) Add(other *HexPosition) *HexPosition {
	return &HexPosition{Q: hp.Q + other.Q, R: hp.R + other.R}
}

func (hp *HexPosition) Subtract(other *HexPosition) *HexPosition {
	return &HexPosition{Q: hp.Q - other.Q, R: hp.R - other.R}
}

func (hp *HexPosition) Scale(factor int) *HexPosition {
	return &HexPosition{Q: hp.Q * factor, R: hp.R * factor}
}

func (hp *HexPosition) Distance(other *HexPosition) int {
	return (abs(hp.Q-other.Q) + abs(hp.Q+hp.R-other.Q-other.R) + abs(hp.R-other.R)) / 2
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

type HexBoard struct {
	width     int
	height    int
	terrain   map[string]string
	entities  map[string]string
	pathfinder *HexPathfinder
}

func NewHexBoard(width, height int) *HexBoard {
	return &HexBoard{
		width:     width,
		height:    height,
		terrain:   make(map[string]string),
		entities:  make(map[string]string),
		pathfinder: NewHexPathfinder(),
	}
}

func (hb *HexBoard) GetWidth() int {
	return hb.width
}

func (hb *HexBoard) GetHeight() int {
	return hb.height
}

func (hb *HexBoard) IsValidPosition(pos turnengine.Position) bool {
	if hexPos, ok := pos.(*HexPosition); ok {
		q, r := hexPos.Q, hexPos.R
		
		if q < -(hb.height/2) || q >= hb.width-(hb.height/2) {
			return false
		}
		
		minR := max(0, -q)
		maxR := min(hb.height-1, hb.width-1-q)
		
		return r >= minR && r <= maxR
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (hb *HexBoard) GetNeighbors(pos turnengine.Position) []turnengine.Position {
	hexPos, ok := pos.(*HexPosition)
	if !ok {
		return nil
	}
	
	directions := []*HexPosition{
		{Q: 1, R: 0}, {Q: 1, R: -1}, {Q: 0, R: -1},
		{Q: -1, R: 0}, {Q: -1, R: 1}, {Q: 0, R: 1},
	}
	
	var neighbors []turnengine.Position
	for _, dir := range directions {
		neighbor := hexPos.Add(dir)
		if hb.IsValidPosition(neighbor) {
			neighbors = append(neighbors, neighbor)
		}
	}
	
	return neighbors
}

func (hb *HexBoard) GetDistance(from, to turnengine.Position) int {
	fromHex, ok1 := from.(*HexPosition)
	toHex, ok2 := to.(*HexPosition)
	if !ok1 || !ok2 {
		return -1
	}
	
	return fromHex.Distance(toHex)
}

func (hb *HexBoard) GetTerrain(pos turnengine.Position) (string, bool) {
	terrain, exists := hb.terrain[pos.Hash()]
	return terrain, exists
}

func (hb *HexBoard) SetTerrain(pos turnengine.Position, terrainType string) error {
	if !hb.IsValidPosition(pos) {
		return fmt.Errorf("invalid position: %s", pos.String())
	}
	
	hb.terrain[pos.Hash()] = terrainType
	return nil
}

func (hb *HexBoard) GetAllPositions() []turnengine.Position {
	var positions []turnengine.Position
	
	for q := -(hb.height / 2); q < hb.width-(hb.height/2); q++ {
		minR := max(0, -q)
		maxR := min(hb.height-1, hb.width-1-q)
		
		for r := minR; r <= maxR; r++ {
			positions = append(positions, &HexPosition{Q: q, R: r})
		}
	}
	
	return positions
}

func (hb *HexBoard) GetEntityAt(pos turnengine.Position) (string, bool) {
	entityID, exists := hb.entities[pos.Hash()]
	return entityID, exists
}

func (hb *HexBoard) SetEntityAt(pos turnengine.Position, entityID string) error {
	if !hb.IsValidPosition(pos) {
		return fmt.Errorf("invalid position: %s", pos.String())
	}
	
	if entityID == "" {
		delete(hb.entities, pos.Hash())
	} else {
		hb.entities[pos.Hash()] = entityID
	}
	
	return nil
}

func (hb *HexBoard) IsPositionOccupied(pos turnengine.Position) bool {
	_, exists := hb.entities[pos.Hash()]
	return exists
}

func (hb *HexBoard) GetMovementCost(pos turnengine.Position) int {
	terrain, exists := hb.GetTerrain(pos)
	if !exists {
		return 1
	}
	
	switch terrain {
	case "grass", "plain":
		return 1
	case "forest":
		return 2
	case "mountain":
		return 3
	case "swamp":
		return 3
	case "road":
		return 1
	case "water":
		return -1
	default:
		return 1
	}
}

func (hb *HexBoard) CanMoveThrough(pos turnengine.Position) bool {
	return hb.GetMovementCost(pos) > 0
}

type HexPathfinder struct{}

func NewHexPathfinder() *HexPathfinder {
	return &HexPathfinder{}
}

func (hpf *HexPathfinder) FindPath(board turnengine.Board, from, to turnengine.Position, movementCost func(turnengine.Position) int) ([]turnengine.Position, error) {
	hexBoard, ok := board.(*HexBoard)
	if !ok {
		return nil, fmt.Errorf("board is not a HexBoard")
	}
	
	_, ok1 := from.(*HexPosition)
	_, ok2 := to.(*HexPosition)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("positions are not HexPositions")
	}
	
	if !hexBoard.IsValidPosition(from) || !hexBoard.IsValidPosition(to) {
		return nil, fmt.Errorf("invalid start or end position")
	}
	
	if from.Equals(to) {
		return []turnengine.Position{from}, nil
	}
	
	openSet := []*turnengine.PathfindingNode{}
	closedSet := make(map[string]*turnengine.PathfindingNode)
	
	startNode := &turnengine.PathfindingNode{
		Position: from,
		GCost:    0,
		HCost:    hpf.CalculateDistance(from, to),
	}
	startNode.CalculateFCost()
	
	openSet = append(openSet, startNode)
	
	for len(openSet) > 0 {
		currentNode := openSet[0]
		currentIndex := 0
		
		for i, node := range openSet {
			if node.FCost < currentNode.FCost || (node.FCost == currentNode.FCost && node.HCost < currentNode.HCost) {
				currentNode = node
				currentIndex = i
			}
		}
		
		openSet = append(openSet[:currentIndex], openSet[currentIndex+1:]...)
		closedSet[currentNode.Position.Hash()] = currentNode
		
		if currentNode.Position.Equals(to) {
			return hpf.reconstructPath(currentNode), nil
		}
		
		neighbors := hexBoard.GetNeighbors(currentNode.Position)
		for _, neighbor := range neighbors {
			if !hexBoard.CanMoveThrough(neighbor) {
				continue
			}
			
			hash := neighbor.Hash()
			if _, inClosed := closedSet[hash]; inClosed {
				continue
			}
			
			newGCost := currentNode.GCost + movementCost(neighbor)
			
			var neighborNode *turnengine.PathfindingNode
			inOpen := false
			for _, node := range openSet {
				if node.Position.Equals(neighbor) {
					neighborNode = node
					inOpen = true
					break
				}
			}
			
			if !inOpen || newGCost < neighborNode.GCost {
				if neighborNode == nil {
					neighborNode = &turnengine.PathfindingNode{Position: neighbor}
					openSet = append(openSet, neighborNode)
				}
				
				neighborNode.GCost = newGCost
				neighborNode.HCost = hpf.CalculateDistance(neighbor, to)
				neighborNode.Parent = currentNode
				neighborNode.CalculateFCost()
			}
		}
	}
	
	return nil, fmt.Errorf("no path found")
}

func (hpf *HexPathfinder) CalculateDistance(from, to turnengine.Position) int {
	fromHex, ok1 := from.(*HexPosition)
	toHex, ok2 := to.(*HexPosition)
	if !ok1 || !ok2 {
		return 0
	}
	
	return fromHex.Distance(toHex)
}

func (hpf *HexPathfinder) reconstructPath(node *turnengine.PathfindingNode) []turnengine.Position {
	var path []turnengine.Position
	current := node
	
	for current != nil {
		path = append([]turnengine.Position{current.Position}, path...)
		current = current.Parent
	}
	
	return path
}

func (hb *HexBoard) LineDraw(from, to *HexPosition) []*HexPosition {
	distance := from.Distance(to)
	if distance == 0 {
		return []*HexPosition{from}
	}
	
	var results []*HexPosition
	for i := 0; i <= distance; i++ {
		t := float64(i) / float64(distance)
		lerped := hb.hexLerp(from, to, t)
		results = append(results, lerped)
	}
	
	return results
}

func (hb *HexBoard) hexLerp(a, b *HexPosition, t float64) *HexPosition {
	q := a.Q + int(math.Round(float64(b.Q-a.Q)*t))
	r := a.R + int(math.Round(float64(b.R-a.R)*t))
	return &HexPosition{Q: q, R: r}
}

func (hb *HexBoard) GetVisiblePositions(from turnengine.Position, sightRange int) []turnengine.Position {
	fromHex, ok := from.(*HexPosition)
	if !ok {
		return nil
	}
	
	var visible []turnengine.Position
	
	for q := fromHex.Q - sightRange; q <= fromHex.Q+sightRange; q++ {
		r1 := max(fromHex.R-sightRange, -q-sightRange)
		r2 := min(fromHex.R+sightRange, -q+sightRange)
		
		for r := r1; r <= r2; r++ {
			pos := &HexPosition{Q: q, R: r}
			if hb.IsValidPosition(pos) && fromHex.Distance(pos) <= sightRange {
				if hb.hasLineOfSight(fromHex, pos) {
					visible = append(visible, pos)
				}
			}
		}
	}
	
	return visible
}

func (hb *HexBoard) hasLineOfSight(from, to *HexPosition) bool {
	line := hb.LineDraw(from, to)
	
	for _, pos := range line {
		if pos.Equals(from) || pos.Equals(to) {
			continue
		}
		
		terrain, exists := hb.GetTerrain(pos)
		if exists {
			switch terrain {
			case "mountain", "forest":
				return false
			}
		}
	}
	
	return true
}