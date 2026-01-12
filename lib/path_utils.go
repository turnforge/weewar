package lib

import (
	"fmt"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// extractPathCoords converts a Path to a flat coordinate array
func ExtractPathCoords(path *v1.Path) (coords []int32) {
	if path != nil && len(path.Edges) != 0 {
		coords = make([]int32, 0, len(path.Edges)*2+2)

		// Add starting position from first edge
		if len(path.Edges) > 0 {
			coords = append(coords, path.Edges[0].FromQ, path.Edges[0].FromR)
		}

		// Add all destination positions
		for _, edge := range path.Edges {
			coords = append(coords, edge.ToQ, edge.ToR)
		}
	}
	return
}

// ReconstructPath reconstructs a complete path from source to destination using AllPaths
// Returns the path as a sequence of edges from source to destination
func ReconstructPath(allPaths *v1.AllPaths, destQ, destR int32) (*v1.Path, error) {
	if allPaths == nil {
		return nil, fmt.Errorf("allPaths is nil")
	}

	// Check if destination is reachable
	destKey := fmt.Sprintf("%d,%d", destQ, destR)
	_, exists := allPaths.Edges[destKey]
	if !exists {
		return nil, fmt.Errorf("destination (%d,%d) not reachable from source (%d,%d)",
			destQ, destR, allPaths.SourceQ, allPaths.SourceR)
	}

	// Build path by walking backwards from destination to source
	var pathEdges []*v1.PathEdge
	currentQ, currentR := destQ, destR
	totalCost := 0.0

	for {
		// Get edge leading to current position
		key := fmt.Sprintf("%d,%d", currentQ, currentR)
		edge, exists := allPaths.Edges[key]
		if !exists {
			// We've reached the source (no edge leads to source)
			break
		}

		// Add edge to path (we'll reverse later)
		pathEdges = append(pathEdges, edge)
		totalCost = edge.TotalCost // Total cost is stored in the final edge

		// Move to parent
		currentQ = edge.FromQ
		currentR = edge.FromR

		// Check if we've reached the source
		if currentQ == allPaths.SourceQ && currentR == allPaths.SourceR {
			break
		}
	}

	// Reverse the path to get source->destination order
	for i := 0; i < len(pathEdges)/2; i++ {
		j := len(pathEdges) - 1 - i
		pathEdges[i], pathEdges[j] = pathEdges[j], pathEdges[i]
	}

	return &v1.Path{
		Edges:     pathEdges,
		TotalCost: totalCost,
	}, nil
}

// GetReachableDestinations extracts all reachable destinations from AllPaths
// Returns a map of destination coordinates to their total movement costs
func GetReachableDestinations(allPaths *v1.AllPaths) map[string]float64 {
	if allPaths == nil || allPaths.Edges == nil {
		return make(map[string]float64)
	}

	destinations := make(map[string]float64)
	for key, edge := range allPaths.Edges {
		destinations[key] = edge.TotalCost
	}
	return destinations
}

// GetMovementCostTo returns the total movement cost to reach a specific destination
// Returns -1 if the destination is not reachable
func GetMovementCostTo(allPaths *v1.AllPaths, destQ, destR int32) float64 {
	if allPaths == nil || allPaths.Edges == nil {
		return -1
	}

	key := fmt.Sprintf("%d,%d", destQ, destR)
	if edge, exists := allPaths.Edges[key]; exists {
		return edge.TotalCost
	}
	return -1
}
