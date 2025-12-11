package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"groupie-tracker-ng/models"
	"groupie-tracker-ng/utils"
)

// LocationHandler gère la page listant les concerts à un lieu spécifique
func LocationHandler(w http.ResponseWriter, r *http.Request) {
	// Extraire le lieu de l'URL (format: /location/ville)
	locationPath := r.URL.Path[len("/location/"):]
	location, err := url.PathUnescape(locationPath)
	if err != nil {
		utils.RenderError(w, http.StatusBadRequest, "Lieu invalide")
		return
	}

	// Récupérer tous les artistes populaires
	spotifyArtists, err := spotifyClient.FetchPopularArtists()
	if err != nil {
		utils.HandleError(w, err, http.StatusInternalServerError)
		return
	}

	// Filtrer les artistes qui pourraient être liés à ce lieu
	// (basé sur le nom du lieu dans les genres ou recherche par nom)
	var relatedArtists []models.Artist
	locationLower := strings.ToLower(location)

	for _, sa := range spotifyArtists {
		// Vérifier si le lieu apparaît dans les genres ou le nom
		artistNameLower := strings.ToLower(sa.Name)
		matches := false

		// Vérifier dans les genres
		for _, genre := range sa.Genres {
			if strings.Contains(strings.ToLower(genre), locationLower) {
				matches = true
				break
			}
		}

		// Vérifier dans le nom de l'artiste
		if strings.Contains(artistNameLower, locationLower) {
			matches = true
		}

		if matches {
			imageURL := ""
			if len(sa.Images) > 0 {
				imageURL = sa.Images[0].URL
			}
			relatedArtists = append(relatedArtists, models.Artist{
				ID:         sa.ID,
				Name:       sa.Name,
				Image:      imageURL,
				Genres:     sa.Genres,
				SpotifyURL: fmt.Sprintf("https://open.spotify.com/artist/%s", sa.ID),
			})
		}
	}

	data := map[string]interface{}{
		"Title":    "Artistes liés à " + location,
		"Location": location,
		"Artists":  relatedArtists,
	}

	renderTemplate(w, "location.html", data)
}
