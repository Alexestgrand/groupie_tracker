package handlers

import (
	"net/http"
)

// ArtistsHandler gère la liste des artistes
func ArtistsHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Liste des Artistes",
	}

	renderTemplate(w, "artists.html", data)
}

// ArtistDetailHandler gère la page de détails d'un artiste
func ArtistDetailHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Détails de l'artiste",
	}

	renderTemplate(w, "artists_details.html", data)
}
