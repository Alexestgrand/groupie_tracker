package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"groupie-tracker-ng/models"
	"groupie-tracker-ng/utils"
)

// ArtistsHandler gère la liste des artistes
func ArtistsHandler(w http.ResponseWriter, r *http.Request) {
	// Récupérer les artistes depuis l'API Spotify
	spotifyArtists, err := spotifyClient.FetchPopularArtists()
	if err != nil {
		utils.HandleError(w, err, http.StatusInternalServerError)
		return
	}

	// Convertir les artistes Spotify en format Artist
	artists := make([]models.Artist, len(spotifyArtists))
	for i, sa := range spotifyArtists {
		imageURL := ""
		if len(sa.Images) > 0 {
			imageURL = sa.Images[0].URL
		}
		artists[i] = models.Artist{
			ID:         sa.ID,
			Name:       sa.Name,
			Image:      imageURL,
			Genres:     sa.Genres,
			SpotifyURL: fmt.Sprintf("https://open.spotify.com/artist/%s", sa.ID),
		}
	}

	data := map[string]interface{}{
		"Title":   "Liste des Artistes",
		"Artists": artists,
	}

	renderTemplate(w, "artists.html", data)
}

// ArtistDetailHandler gère la page de détails d'un artiste
func ArtistDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Extraire l'ID Spotify de l'URL (format: /artist/4uLU6hMCjMI75M1A2tKUQC)
	artistID := strings.TrimPrefix(r.URL.Path, "/artist/")

	if artistID == "" {
		utils.RenderError(w, http.StatusBadRequest, "ID d'artiste invalide")
		return
	}

	// Récupérer les détails complets de l'artiste depuis l'API Spotify
	fullArtist, err := spotifyClient.GetArtistByID(artistID)
	if err != nil {
		utils.RenderError(w, http.StatusNotFound, "Artiste non trouvé")
		return
	}

	// Construire l'objet ArtistDetail
	imageURL := ""
	if len(fullArtist.Images) > 0 {
		imageURL = fullArtist.Images[0].URL
	}

	detail := models.ArtistDetail{
		Artist: models.Artist{
			ID:         fullArtist.ID,
			Name:       fullArtist.Name,
			Image:      imageURL,
			Genres:     fullArtist.Genres,
			Popularity: fullArtist.Popularity,
			SpotifyURL: fullArtist.ExternalURLs.Spotify,
		},
		Followers: fullArtist.Followers.Total,
		SpotifyInfo: map[string]interface{}{
			"spotify_id":  fullArtist.ID,
			"genres":      fullArtist.Genres,
			"popularity":  fullArtist.Popularity,
			"followers":   fullArtist.Followers.Total,
			"spotify_url": fullArtist.ExternalURLs.Spotify,
		},
	}

	data := map[string]interface{}{
		"Title":  "Détails de " + detail.Name,
		"Artist": detail, // ArtistDetail avec Artist embed
	}

	renderTemplate(w, "artists_details.html", data)
}
