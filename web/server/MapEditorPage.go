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

	protos "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
)

// Toolbar buttons on the editor page
type TBButton struct {
	ButtonId  string
	IconImage string
	Label     string
	Command   string
}

type TerrainType struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	MoveCost     int    `json:"moveCost"`
	DefenseBonus int    `json:"defenseBonus"`
	IconDataURL  string `json:"iconDataURL"`
	HasPlayerColors bool `json:"hasPlayerColors"`
}

type UnitType struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	IconDataURL string `json:"iconDataURL"`
}

type MapEditorPage struct {
	BasePage
	Header           Header
	IsOwner          bool
	MapId            string
	Map              *protos.Map
	Errors           map[string]string
	TBButtons        []*TBButton
	AllowCustomId    bool
	NatureTerrains   []TerrainType
	CityTerrains     []TerrainType
	UnitTypes        []UnitType
	PlayerCount      int
}

func (g *MapEditorPage) Copy() View { return &MapEditorPage{} }

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

func (v *MapEditorPage) SetupDefaults() {
	v.Header.Width = "w-full"
	v.Header.PageData = v
	v.Header.FixedHeader = true
	v.Header.ShowHomeButton = true
	v.Header.ShowLogoutButton = false
	v.Header.ShowComposeButton = false
	
	// Initialize terrain types with actual asset images
	v.NatureTerrains = []TerrainType{}
	v.CityTerrains = []TerrainType{}
	v.PlayerCount = 4 // Default player count for map editor
	
	// No longer need hardcoded map - terrain type is now in TerrainData struct
	
	// Create asset manager to load terrain tile images
	assetManager := weewar.NewAssetManager("./assets/v1") // Adjust path as needed
	
	for i := 0; i <= 26; i++ {
		terrainData := weewar.GetTerrainData(i)
		if terrainData != nil {
			// Try to load the actual tile image
			var iconDataURL string
			if assetManager.HasTileAsset(i) {
				if img, err := assetManager.GetTileImage(i); err == nil {
					if dataURL, err := imageToDataURL(img); err == nil {
						iconDataURL = dataURL
					}
				}
			}
			
			// If we couldn't load the image, use a fallback or empty string
			if iconDataURL == "" {
				// Could add emoji fallbacks here if needed
				iconDataURL = "" // Template can handle missing icons
			}
			
			terrain := TerrainType{
				ID:              terrainData.ID,
				Name:            terrainData.Name,
				MoveCost:        terrainData.MoveCost,
				DefenseBonus:    terrainData.DefenseBonus,
				IconDataURL:     iconDataURL,
				HasPlayerColors: terrainData.Type == weewar.TerrainPlayer,
			}
			
			if terrainData.Type == weewar.TerrainPlayer {
				v.CityTerrains = append(v.CityTerrains, terrain)
			} else {
				v.NatureTerrains = append(v.NatureTerrains, terrain)
			}
		}
	}
	
	// Sort terrain lists by ID
	sort.Slice(v.CityTerrains, func(i, j int) bool {
		return v.CityTerrains[i].ID < v.CityTerrains[j].ID
	})
	sort.Slice(v.NatureTerrains, func(i, j int) bool {
		return v.NatureTerrains[i].ID < v.NatureTerrains[j].ID
	})
	
	// Load unit types with icons
	v.UnitTypes = []UnitType{}
	unitIDs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 24, 25, 26, 27, 28, 29}
	
	for _, unitID := range unitIDs {
		unitData, err := assetManager.GetUnitData(unitID)
		if err != nil {
			continue
		}
		if unitData != nil {
			// Try to load the unit icon (default color 0)
			var iconDataURL string
			if assetManager.HasUnitAsset(unitID, 0) {
				if img, err := assetManager.GetUnitImage(unitID, 0); err == nil {
					if dataURL, err := imageToDataURL(img); err == nil {
						iconDataURL = dataURL
					}
				}
			}
			
			v.UnitTypes = append(v.UnitTypes, UnitType{
				ID:          unitData.ID,
				Name:        unitData.Name,
				IconDataURL: iconDataURL,
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

func (v *MapEditorPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	v.Header.Load(r, w, vc)
	v.SetupDefaults()
	queryParams := r.URL.Query()
	v.MapId = r.PathValue("mapId")
	templateName := queryParams.Get("template")
	loggedInUserId := vc.AuthMiddleware.GetLoggedInUserId(r)

	slog.Info("Loading composer for map with ID: ", "nid", v.MapId)

	if v.MapId == "" {
		if false && loggedInUserId == "" {
			// For now enforce login even on new
			qs := r.URL.RawQuery
			if len(qs) > 0 {
				qs = "?" + qs
			}
			http.Redirect(w, r, fmt.Sprintf("/login?callbackURL=%s", fmt.Sprintf("/maps/new%s", qs)), http.StatusSeeOther)
			return nil, true
		}
		v.IsOwner = true
		v.Map = &protos.Map{}
		if v.Map.Name == "" {
			v.Map.Name = "Untitled Map"
		}
		log.Println("Using template: ", templateName)
	} else {
		client, _ := vc.ClientMgr.GetMapsSvcClient()
		resp, err := client.GetMap(context.Background(), &protos.GetMapRequest{
			Id: v.MapId,
		})
		if err != nil {
			log.Println("Error getting map: ", err)
			return err, false
		}

		v.IsOwner = loggedInUserId == resp.Map.CreatorId
		log.Println("LoggedUser: ", loggedInUserId, resp.Map.CreatorId)

		if false && !v.IsOwner {
			log.Println("Composer is NOT the owner.  Redirecting to view page...")
			if loggedInUserId == "" {
				http.Redirect(w, r, fmt.Sprintf("/login?callbackURL=%s", fmt.Sprintf("/maps/%s/compose", v.MapId)), http.StatusSeeOther)
			} else {
				http.Redirect(w, r, fmt.Sprintf("/maps/%s/view", v.MapId), http.StatusSeeOther)
			}
			return nil, true
		}

		v.Map = resp.Map
		v.Header.RightMenuItems = []HeaderMenuItem{
			{Title: "Save", Id: "saveRightButton", Link: "javascript:void(0)"},
			{Title: "Delete", Id: "deleteRightButton", Link: "javascript:void(0)"},
		}
	}
	return
}
