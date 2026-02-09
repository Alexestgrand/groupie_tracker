package handlers

import (
	"encoding/json"
	"net/http"

	"groupie-tracker-ng/models"
	"groupie-tracker-ng/utils"
)

// SearchHandler gère la recherche d'artistes
func SearchHandler(w http.ResponseWriter, r *http.Request) {
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

	// Récupérer tous les artistes
	artists, err := apiClient.FetchArtists()
	if err != nil {
		utils.HandleError(w, err, http.StatusInternalServerError)
		return
	}

	// Rechercher les artistes
	filteredArtists := utils.SearchArtists(artists, query)

	data := map[string]interface{}{
		"Title":   "Résultats de recherche pour: " + query,
		"Artists": filteredArtists,
		"Query":   query,
	}

	renderTemplate(w, "artists.html", data)
}

// SuggestionsHandler retourne des suggestions JSON pour la barre de recherche
func SuggestionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Méthode non autorisée"})
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]models.Artist{})
		return
	}

	// Valider la longueur de la requête
	if len(query) > 100 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Requête trop longue"})
		return
	}

	// Récupérer tous les artistes
	artists, err := apiClient.FetchArtists()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]models.Artist{})
		return
	}

	// Générer les suggestions
	suggestions := utils.GetSuggestions(artists, query)

	// Convertir en format simple pour le frontend
	type Suggestion struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	}

	result := make([]Suggestion, len(suggestions))
	for i, artist := range suggestions {
		result[i] = Suggestion{
			Name: artist.Name,
			ID:   artist.ID,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
