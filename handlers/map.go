package handlers

import (
	"fmt"
	"net/http"

	"groupie-tracker-ng/models"
	"groupie-tracker-ng/utils"
)

// MapHandler gère la page de la carte interactive
func MapHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier que la méthode HTTP est GET
	if r.Method != http.MethodGet {
		utils.RenderError(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	// Récupérer tous les artistes populaires depuis Spotify
	spotifyArtists, err := spotifyClient.FetchPopularArtists()
	if err != nil {
		utils.HandleError(w, err, http.StatusInternalServerError)
		return
	}

	// Vérifier qu'on a des artistes
	if len(spotifyArtists) == 0 {
		utils.RenderError(w, http.StatusNotFound, "Aucun artiste disponible pour la carte")
		return
	}

	// Préparer les données pour la carte
	// Organiser les artistes par genres (qui peuvent représenter des "régions" musicales)
	type GenreData struct {
		Genre   string
		Artists []models.Artist
	}

	genreMap := make(map[string][]models.Artist)

	// Grouper les artistes par genre
	for _, sa := range spotifyArtists {
		imageURL := ""
		if len(sa.Images) > 0 {
			imageURL = sa.Images[0].URL
		}

		artist := models.Artist{
			ID:         sa.ID,
			Name:       sa.Name,
			Image:      imageURL,
			Genres:     sa.Genres,
			SpotifyURL: fmt.Sprintf("https://open.spotify.com/artist/%s", sa.ID),
		}

		// Ajouter l'artiste à chaque genre auquel il appartient
		for _, genre := range sa.Genres {
			genreMap[genre] = append(genreMap[genre], artist)
		}

		// Si l'artiste n'a pas de genre, l'ajouter à "Autres"
		if len(sa.Genres) == 0 {
			genreMap["Autres"] = append(genreMap["Autres"], artist)
		}
	}

	// Convertir en slice pour le template
	genreDataList := make([]GenreData, 0, len(genreMap))
	for genre, artists := range genreMap {
		genreDataList = append(genreDataList, GenreData{
			Genre:   genre,
			Artists: artists,
		})
	}

	data := map[string]interface{}{
		"Title":     "Carte Interactive",
		"GenreData": genreDataList,
		"Artists":   spotifyArtists,
	}

	renderTemplate(w, "map.html", data)
}
