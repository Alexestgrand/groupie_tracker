package handlers

import (
	"net/http"
)

// SearchHandler gère la recherche d'artistes
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/artists", http.StatusSeeOther)
}

// SuggestionsHandler retourne des suggestions JSON pour la barre de recherche
func SuggestionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}
