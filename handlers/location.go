package handlers

import (
	"net/http"
	"net/url"

	"groupie-tracker-ng/models"
	"groupie-tracker-ng/utils"
)

// LocationHandler gère la page listant les concerts à un lieu spécifique
func LocationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RenderError(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	// Extraire le lieu de l'URL (format: /location/ville)
	locationPath := r.URL.Path[len("/location/"):]

	// Valider que le chemin n'est pas vide
	if locationPath == "" {
		utils.RenderError(w, http.StatusBadRequest, "Lieu non spécifié")
		return
	}

	location, err := url.PathUnescape(locationPath)
	if err != nil {
		utils.RenderError(w, http.StatusBadRequest, "Lieu invalide dans l'URL")
		return
	}

	// Valider la longueur du lieu
	if len(location) < 1 || len(location) > 100 {
		utils.RenderError(w, http.StatusBadRequest, "Nom de lieu invalide")
		return
	}

	// Note: Spotify ne fournit pas de données de concerts/lieux
	// On retourne une liste vide avec un message
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
