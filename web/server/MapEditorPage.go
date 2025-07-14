package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	protos "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

// Toolbar buttons on the editor page
type TBButton struct {
	ButtonId  string
	IconImage string
	Label     string
	Command   string
}

type MapEditorPage struct {
	Header        Header
	IsOwner       bool
	MapId         string
	Map           *protos.Map
	Errors        map[string]string
	TBButtons     []*TBButton
	AllowCustomId bool
}

func (g *MapEditorPage) Copy() View { return &MapEditorPage{} }

func (v *MapEditorPage) SetupDefaults() {
	v.Header.Width = "w-full"
	v.Header.PageData = v
	v.Header.FixedHeader = true
	v.Header.ShowHomeButton = true
	v.Header.ShowLogoutButton = false
	v.Header.ShowComposeButton = false
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

	slog.Info("Loading composer for notation with ID: ", "nid", v.MapId)

	if v.MapId == "" {
		if loggedInUserId == "" {
			// For now enforce login even on new
			qs := r.URL.RawQuery
			if len(qs) > 0 {
				qs = "?" + qs
			}
			http.Redirect(w, r, fmt.Sprintf("/login?callbackURL=%s", fmt.Sprintf("/notations/new%s", qs)), http.StatusSeeOther)
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
			log.Println("Error getting notation: ", err)
			return err, false
		}

		v.IsOwner = loggedInUserId == resp.Map.CreatorId
		log.Println("LoggedUser: ", loggedInUserId, resp.Map.CreatorId)

		if !v.IsOwner {
			log.Println("Composer is NOT the owner.  Redirecting to view page...")
			if loggedInUserId == "" {
				http.Redirect(w, r, fmt.Sprintf("/login?callbackURL=%s", fmt.Sprintf("/notations/%s/compose", v.MapId)), http.StatusSeeOther)
			} else {
				http.Redirect(w, r, fmt.Sprintf("/notations/%s/view", v.MapId), http.StatusSeeOther)
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
