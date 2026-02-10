package handlers

import (
	"net/http"
	"net/url"

	"groupie-tracker-ng/models"
	"groupie-tracker-ng/utils"
)

// LocationHandler gère la page listant les concerts à un lieu.
// L'API Spotify ne fournit pas de lieux ni dates de concerts ; la page affiche un message explicatif.
func LocationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RenderError(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	locationPath := r.URL.Path[len("/location/"):]
	if locationPath == "" {
		utils.RenderError(w, http.StatusBadRequest, "Lieu non spécifié")
		return
	}

	location, err := url.PathUnescape(locationPath)
	if err != nil || len(location) < 1 || len(location) > 200 {
		utils.RenderError(w, http.StatusBadRequest, "Lieu invalide")
		return
	}

	// Source unique : Spotify — pas de données concerts par lieu
	var relatedArtists []models.Artist
	concertsByArtist := make(map[int][]string)

	data := map[string]interface{}{
		"Title":            "Concerts à " + location,
		"Location":         location,
		"Artists":          relatedArtists,
		"ConcertsByArtist": concertsByArtist,
	}

	renderTemplate(w, "location.html", data)
}
