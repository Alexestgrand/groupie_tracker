package handlers

import (
	"net/http"

	"groupie-tracker-ng/utils"
)

// MapHandler gère la page de la carte interactive
func MapHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RenderError(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	// Récupérer toutes les données nécessaires
	artists, err := apiClient.FetchArtists()
	if err != nil {
		utils.HandleError(w, err, http.StatusInternalServerError)
		return
	}

	// Note: Spotify ne fournit pas de lieux de concerts
	// On affiche un message indiquant que cette fonctionnalité n'est pas disponible
	type LocationData struct {
		Location string
		Dates    []string
		Artists  []string
	}

	locationsList := []*LocationData{}

	data := map[string]interface{}{
		"Title":     "Carte Interactive",
		"Artists":   artists,
		"Locations": locationsList,
	}

	renderTemplate(w, "map.html", data)
}
