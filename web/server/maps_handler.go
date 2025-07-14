package server

import (
	"fmt"
	"log"
	"net/http"
)

func (r *RootViewsHandler) setupMapsMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", r.ViewRenderer(Copier(&MapListingPage{}), ""))
	mux.HandleFunc("/new", r.ViewRenderer(Copier(&MapEditorPage{}), ""))
	mux.HandleFunc("/{mapId}/view", r.ViewRenderer(Copier(&MapDetailPage{}), ""))
	mux.HandleFunc("/{mapId}/edit", r.ViewRenderer(Copier(&MapEditorPage{}), ""))
	mux.HandleFunc("/{mapId}/copy", func(w http.ResponseWriter, r *http.Request) {
		notationId := r.PathValue("notationId")
		http.Redirect(w, r, fmt.Sprintf("/appitems/new?copyFrom=%s", notationId), http.StatusFound)
	})
	mux.HandleFunc("/{mapid}", func(w http.ResponseWriter, r *http.Request) {
		// Handle Delete here
		log.Println("=============")
		log.Println("Catch all - should not be coming here if not a delete call", r.Header)
		log.Println("=============")
		http.Redirect(w, r, "/", http.StatusFound)
	})
	return mux
}
