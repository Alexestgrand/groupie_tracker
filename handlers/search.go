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
	// Vérifier que la méthode HTTP est GET
	if r.Method != http.MethodGet {
		utils.RenderError(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	query := r.URL.Query().Get("q")

	// Valider la requête
	if query == "" {
		http.Redirect(w, r, "/artists", http.StatusSeeOther)
		return
	}

	// Valider la longueur de la requête
	if len(query) < 1 || len(query) > 100 {
		utils.RenderError(w, http.StatusBadRequest, "Requête de recherche invalide (1-100 caractères)")
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
	// Vérifier que la méthode HTTP est GET
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Méthode non autorisée"})
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{})
		return
	}

	// Valider la longueur de la requête
	if len(query) > 100 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Requête trop longue"})
		return
	}

	// Récupérer tous les artistes populaires
	spotifyArtists, err := spotifyClient.FetchPopularArtists()
	if err != nil {
		// En cas d'erreur, retourner un tableau vide plutôt qu'une erreur
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
