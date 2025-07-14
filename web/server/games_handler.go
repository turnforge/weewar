package server

import (
	"fmt"
	"log"
	"net/http"
)

func (r *RootViewsHandler) setupGamesMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", r.ViewRenderer(Copier(&GameListingPage{}), ""))
	mux.HandleFunc("/new", r.ViewRenderer(Copier(&GameCreatorPage{}), ""))
	mux.HandleFunc("/{gameId}/view", r.ViewRenderer(Copier(&GameDetailPage{}), ""))
	mux.HandleFunc("/{gameId}/copy", func(w http.ResponseWriter, r *http.Request) {
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
