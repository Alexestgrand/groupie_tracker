package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"groupie-tracker-ng/models"
	"groupie-tracker-ng/utils"
)

// SearchHandler gère la recherche d'artistes
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Redirect(w, r, "/artists", http.StatusSeeOther)
		return
	}

	// Rechercher les artistes sur Spotify
	spotifyArtists, err := spotifyClient.SearchArtists(query, 20)
	if err != nil {
		utils.HandleError(w, err, http.StatusInternalServerError)
		return
	}

	// Convertir en format Artist
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
		"Title":   "Résultats de recherche pour: " + query,
		"Artists": artists,
		"Query":   query,
	}

	renderTemplate(w, "artists.html", data)
}

// SuggestionsHandler retourne des suggestions JSON pour la barre de recherche
func SuggestionsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{})
		return
	}

	// Récupérer tous les artistes populaires
	spotifyArtists, err := spotifyClient.FetchPopularArtists()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{})
		return
	}

	// Convertir en format Artist pour utiliser la fonction getSuggestions
	artists := make([]models.Artist, len(spotifyArtists))
	for i, sa := range spotifyArtists {
		artists[i] = models.Artist{
			ID:   sa.ID,
			Name: sa.Name,
		}
	}

	// Générer les suggestions
	suggestions := utils.GetSuggestions(artists, query)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}
