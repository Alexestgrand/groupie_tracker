package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"groupie-tracker-ng/models"
	"groupie-tracker-ng/utils"
)

// SearchHandler gère la recherche d'artistes (même formulaire de filtres que /artists)
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RenderError(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Redirect(w, r, "/artists", http.StatusSeeOther)
		return
	}
	if len(query) < 1 || len(query) > 100 {
		utils.RenderError(w, http.StatusBadRequest, "Requête de recherche invalide (1-100 caractères)")
		return
	}

	artists, err := apiClient.FetchArtists()
	apiError := ""
	if err != nil {
		artists = []models.Artist{}
		apiError = err.Error()
	}

	filteredArtists := utils.SearchArtists(artists, query)
	filterOptions := utils.ParseFilterOptions(r.URL.Query())
	filteredArtists = utils.FilterArtists(filteredArtists, filterOptions)

	locationsList := []string{}
	minYear := ""
	maxYear := ""
	if filterOptions.MinYear > 0 {
		minYear = strconv.Itoa(filterOptions.MinYear)
	}
	if filterOptions.MaxYear > 0 {
		maxYear = strconv.Itoa(filterOptions.MaxYear)
	}
	memberSelected := make(map[int]bool)
	for _, mc := range filterOptions.MemberCount {
		memberSelected[mc] = true
	}
	locationSelected := make(map[string]bool)
	for _, loc := range filterOptions.Locations {
		locationSelected[loc] = true
	}

	data := map[string]interface{}{
		"Title":            "Résultats pour « " + query + " »",
		"Artists":          filteredArtists,
		"Query":            query,
		"Locations":        locationsList,
		"MinYear":          minYear,
		"MaxYear":          maxYear,
		"Member1":          memberSelected[1],
		"Member2":          memberSelected[2],
		"Member3":          memberSelected[3],
		"Member4":          memberSelected[4],
		"Member5":          memberSelected[5],
		"LocationSelected": locationSelected,
		"APIError":         apiError,
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
