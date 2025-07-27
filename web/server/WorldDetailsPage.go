package server

import (
	"context"
	"log"
	"net/http"

	protos "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

type WorldDetailsPage struct {
	BasePage
	Header  Header
	World   *protos.World // Use the same type as WorldEditorPage for consistency
	WorldId string
}

func (p *WorldDetailsPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	p.WorldId = r.PathValue("worldId")
	if p.WorldId == "" {
		http.Error(w, "World ID is required", http.StatusBadRequest)
		return nil, true
	}

	p.Title = "World Details"
	p.Header.Load(r, w, vc)

	// Fetch the World using the client manager
	client, err := vc.ClientMgr.GetWorldsSvcClient()
	if err != nil {
		log.Printf("Error getting Worlds client: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return nil, true
	}

	req := &protos.GetWorldRequest{
		Id: p.WorldId,
	}

	resp, err := client.GetWorld(context.Background(), req)
	if err != nil {
		log.Printf("Error fetching World %s: %v", p.WorldId, err)
		http.Error(w, "World not found", http.StatusNotFound)
		return nil, true
	}

	if resp.World != nil {
		// Use the World data for display
		p.World = resp.World
		p.Title = p.World.Name
	}

	return nil, false
}

// GetWorldDataForJS converts the protobuf World to the format expected by the TypeScript frontend
func (p *WorldDetailsPage) GetWorldDataForJS() map[string]interface{} {
	if p.World == nil {
		return nil
	}
	
	// Convert protobuf to the TypeScript-compatible format
	result := map[string]interface{}{
		"Name":   p.World.Name,
		"name":   p.World.Name, // Both for compatibility
		"width":  40,  // Default width - we'll need to calculate this from tiles
		"height": 40,  // Default height - we'll need to calculate this from tiles
	}
	
	// Add tiles from WorldData if present
	if p.World.WorldData != nil {
		// Convert tiles array
		tiles := make([]map[string]interface{}, 0, len(p.World.WorldData.Tiles))
		maxQ, maxR := int32(0), int32(0)
		minQ, minR := int32(0), int32(0)
		
		for _, tile := range p.World.WorldData.Tiles {
			tiles = append(tiles, map[string]interface{}{
				"q":        tile.Q,
				"r":        tile.R,
				"tileType": tile.TileType,
				"player":   tile.Player,
			})
			
			// Calculate bounds
			if tile.Q > maxQ { maxQ = tile.Q }
			if tile.Q < minQ { minQ = tile.Q }
			if tile.R > maxR { maxR = tile.R }
			if tile.R < minR { minR = tile.R }
		}
		
		// Calculate actual dimensions from tile bounds
		if len(tiles) > 0 {
			result["width"] = int(maxQ - minQ + 1)
			result["height"] = int(maxR - minR + 1)
		}
		
		result["tiles"] = tiles
		
		// Convert units array
		units := make([]map[string]interface{}, 0, len(p.World.WorldData.Units))
		for _, unit := range p.World.WorldData.Units {
			units = append(units, map[string]interface{}{
				"q":        unit.Q,
				"r":        unit.R,
				"player":   unit.Player,
				"unitType": unit.UnitType,
				"available_health": unit.AvailableHealth,
				"distance_left":    unit.DistanceLeft,
				"turn_counter":     unit.TurnCounter,
			})
		}
		result["units"] = units
	}
	
	return result
}

func (p *WorldDetailsPage) Copy() View {
	return &WorldDetailsPage{}
}
