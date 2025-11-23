package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"log"
	"log/slog"
	"net/http"
	"sort"

	protos "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	weewar "github.com/turnforge/weewar/services"
)

// Toolbar buttons on the editor page
type TBButton struct {
	ButtonId  string
	IconImage string
	Label     string
	Command   string
}

// TerrainType represents whether terrain is nature or player-controllable
type TerrainType2 int32

const (
	TerrainNature TerrainType2 = iota // Natural terrain (grass, mountains, water, etc.)
	TerrainPlayer                     // Player-controllable structures (bases, cities, etc.)
)

// TerrainData represents terrain type information
type TerrainData struct {
	ID           int32        // `json:"id"`
	Name         string       // `json:"name"`
	BaseMoveCost float64      // `json:"baseMoveCost"` // Base movement cost for this terrain
	DefenseBonus float64      // `json:"defenseBonus"`
	Type         TerrainType2 // `json:"type"` // Nature or Player terrain
	Properties   []string     // `json:"properties,omitempty"`
	// Note: Unit-specific movement costs in RulesEngine can override base cost
}

type TerrainType struct {
	TerrainData
	IconDataURL     string `json:"iconDataURL"`
	HasPlayerColors bool   `json:"hasPlayerColors"`
}

type UnitType struct {
	*protos.UnitDefinition
	IconDataURL string `json:"iconDataURL"`
}

type WorldEditorPage struct {
	BasePage
	Header         Header
	IsOwner        bool
	WorldId        string
	World          *protos.World
	WorldData      *protos.WorldData
	Errors         map[string]string
	TBButtons      []*TBButton
	AllowCustomId  bool
	NatureTerrains []TerrainType
	CityTerrains   []TerrainType
	UnitTypes      []UnitType
	PlayerCount    int
	Theme          string // Theme name from query parameter (default, fantasy, modern)
}

func (g *WorldEditorPage) Copy() View { return &WorldEditorPage{} }

// imageToDataURL converts an image to a data URL
func imageToDataURL(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}

	// Encode as base64
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return fmt.Sprintf("data:image/png;base64,%s", encoded), nil
}

func (v *WorldEditorPage) SetupDefaults() {
	v.Header.Width = "w-full"
	v.Header.PageData = v
	v.Header.FixedHeader = true
	v.Header.ShowHomeButton = true
	v.Header.ShowLogoutButton = false
	v.Header.ShowComposeButton = false

	// Initialize terrain types with actual asset images
	v.NatureTerrains = []TerrainType{}
	v.CityTerrains = []TerrainType{}
	v.PlayerCount = 4 // Default player count for world editor

	// Determine whether to use theme-based assets or PNG assets
	// Set useTheme = true to use theme assets, false for PNG assets
	useTheme := true     // Set to true to use theme-based assets
	themeName := v.Theme // Use theme from query parameter (set in Load method)

	// Get the theme manager
	tm := GetThemeManager()

	rulesEngine := weewar.DefaultRulesEngine()
	for i := int32(0); i <= 30; i++ {
		terrainData, err := rulesEngine.GetTerrainData(i)
		if err == nil && terrainData != nil {
			// Get the appropriate icon URL from theme manager
			iconDataURL := tm.GetTerrainIconURL(i, useTheme, themeName)

			// Get the themed name or use default
			terrainName := tm.GetTerrainName(i, terrainData.Name, useTheme, themeName)

			// Calculate base movement cost from terrain-unit properties (use average or default)
			baseMoveCost := 1.0 // Default
			// TODO: Could calculate average movement cost across all units for this terrain

			terrain := TerrainType{
				TerrainData: TerrainData{
					ID:           terrainData.Id,
					Name:         terrainName,
					BaseMoveCost: baseMoveCost,
					DefenseBonus: 0.0, // Defense bonus is now calculated per unit-terrain combination
				},
				IconDataURL:     iconDataURL,
				HasPlayerColors: false, // TODO: Add terrain type info to proto or use heuristic
			}

			// Use heuristic to determine terrain type based on ID
			// TODO: Add terrain type field to proto definition
			isPlayerTerrain := terrainData.Id == 1 || terrainData.Id == 2 || terrainData.Id == 3 ||
				terrainData.Id == 6 || terrainData.Id == 16 || terrainData.Id == 20 ||
				terrainData.Id == 21 || terrainData.Id == 25 // Base, Hospital, Silo, Mines, City, Tower

			if isPlayerTerrain {
				terrain.HasPlayerColors = true
				v.CityTerrains = append(v.CityTerrains, terrain)
				// log.Println("Appending City Terrains: ", terrain)
			} else if terrainData.Id != 0 { // Skip Clear (ID 0) since we have a dedicated button
				v.NatureTerrains = append(v.NatureTerrains, terrain)
				// log.Println("Appending Nature Terrains: ", terrain)
			}
		}
	}

	// Sort terrain lists by name for easier visual grouping
	// Clear should always be first in Nature Terrains
	sort.Slice(v.CityTerrains, func(i, j int) bool {
		//return v.CityTerrains[i].Name < v.CityTerrains[j].Name
		return v.CityTerrains[i].ID < v.CityTerrains[j].ID
	})
	sort.Slice(v.NatureTerrains, func(i, j int) bool {
		// Clear (ID 0) should always be first
		if v.NatureTerrains[i].ID == 0 {
			return true
		}
		if v.NatureTerrains[j].ID == 0 {
			return false
		}
		// return v.NatureTerrains[i].Name < v.NatureTerrains[j].Name
		return v.NatureTerrains[i].ID < v.NatureTerrains[j].ID
	})

	// Load unit types with icons
	v.UnitTypes = []UnitType{}

	for _, unitID := range AllowedUnitIDs {
		unitData, err := rulesEngine.GetUnitData(unitID)
		if unitData != nil && err == nil {
			// Get the appropriate icon URL from theme manager
			iconDataURL := tm.GetUnitIconURL(unitID, useTheme, themeName)

			// Create a copy of unitData with themed name
			themedUnitData := *unitData
			themedUnitData.Name = tm.GetUnitName(unitID, unitData.Name, useTheme, themeName)

			v.UnitTypes = append(v.UnitTypes, UnitType{
				UnitDefinition: &themedUnitData,
				IconDataURL:    iconDataURL,
			})
		}
	}

	v.Header.Styles = map[string]any{
		"FixedHeightHeader":          true,
		"HeaderHeightIfFixed":        "70px",
		"MinWidthForFullWidthMenu":   "24em",
		"MinWidthForHamburgerMenu":   "48em",
		"MinWidthForCompressingLogo": "24em",

		"EditorHeaderHeight":  "50px",
		"ToolbarButtonHeight": "30px",
		"HeaderHeight":        "90px",
		"StatusBarHeight":     "30px",
		"HeaderBarHeight":     "90px",
	}
	v.TBButtons = []*TBButton{
		/*
			{
				ButtonId:  "TB_Save",
				IconImage: "/static/icons/save.png",
				Label:     "Save (Cmd-s)",
				Command:   "saveDocument",
			},
		*/
		{
			ButtonId:  "TB_Refresh",
			IconImage: "/static/icons/Refresh.png",
			Label:     "Refresh (Cmd-enter)",
			Command:   "updatePreview",
		},
	}
}

func (v *WorldEditorPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	v.Header.Load(r, w, vc)

	// Read query parameters first (before SetupDefaults)
	queryParams := r.URL.Query()

	v.Theme = ThemeFromRequest(r)
	v.SetupDefaults()
	v.WorldId = r.PathValue("worldId")
	templateName := queryParams.Get("template")
	loggedInUserId := vc.AuthMiddleware.GetLoggedInUserId(r)

	slog.Info("Loading composer for world with ID: ", "nid", v.WorldId)

	if v.WorldId == "" {
		if false && loggedInUserId == "" {
			// For now enforce login even on new
			qs := r.URL.RawQuery
			if len(qs) > 0 {
				qs = "?" + qs
			}
			http.Redirect(w, r, fmt.Sprintf("/login?callbackURL=%s", fmt.Sprintf("/worlds/new%s", qs)), http.StatusSeeOther)
			return nil, true
		}
		v.IsOwner = true
		v.World = &protos.World{}
		v.WorldData = &protos.WorldData{}
		if v.World.Name == "" {
			v.World.Name = "Untitled World"
		}
		log.Println("Using template: ", templateName)
	} else {
		client, _ := vc.ClientMgr.GetWorldsSvcClient()
		resp, err := client.GetWorld(context.Background(), &protos.GetWorldRequest{
			Id: v.WorldId,
		})
		if err != nil {
			log.Println("Error getting world: ", err)
			return err, false
		}

		v.IsOwner = loggedInUserId == resp.World.CreatorId
		log.Println("LoggedUser: ", loggedInUserId, resp.World.CreatorId)

		if false && !v.IsOwner {
			log.Println("Composer is NOT the owner.  Redirecting to view page...")
			if loggedInUserId == "" {
				http.Redirect(w, r, fmt.Sprintf("/login?callbackURL=%s", fmt.Sprintf("/worlds/%s/compose", v.WorldId)), http.StatusSeeOther)
			} else {
				http.Redirect(w, r, fmt.Sprintf("/worlds/%s/view", v.WorldId), http.StatusSeeOther)
			}
			return nil, true
		}

		v.World = resp.World
		v.WorldData = resp.WorldData
		v.Header.RightMenuItems = []HeaderMenuItem{
			{Title: "Save", Id: "saveRightButton", Link: "javascript:void(0)"},
			{Title: "Delete", Id: "deleteRightButton", Link: "javascript:void(0)"},
		}
	}
	return
}
