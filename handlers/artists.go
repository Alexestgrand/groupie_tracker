package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"groupie-tracker-ng/models"
	"groupie-tracker-ng/utils"
)

// ArtistsHandler gère la liste des artistes avec filtres et recherche
func ArtistsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RenderError(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	// Récupérer les artistes depuis l'API Spotify
	artists, err := apiClient.FetchArtists()
	apiError := ""
	if err != nil {
		// Ne pas faire planter la page : afficher une liste vide et un message
		artists = []models.Artist{}
		apiError = err.Error()
	}

	// Appliquer la recherche si présente
	query := r.URL.Query().Get("q")
	if query != "" && len(query) >= 1 {
		artists = utils.SearchArtists(artists, query)
	}

	// Appliquer les filtres (adaptés pour Spotify)
	filterOptions := utils.ParseFilterOptions(r.URL.Query())
	
	// Note: Spotify ne fournit pas de lieux de concerts, donc on ignore ce filtre
	// Appliquer les autres filtres (année, membres, etc.)
	artists = utils.FilterArtists(artists, filterOptions)

	// Spotify ne fournit pas de lieux, donc liste vide
	locationsList := []string{}

	// Préparer les données pour les filtres
	minYear := ""
	maxYear := ""
	if filterOptions.MinYear > 0 {
		minYear = strconv.Itoa(filterOptions.MinYear)
	}
	if filterOptions.MaxYear > 0 {
		maxYear = strconv.Itoa(filterOptions.MaxYear)
	}

	// Vérifier quels nombres de membres sont sélectionnés
	memberSelected := make(map[int]bool)
	for _, mc := range filterOptions.MemberCount {
		memberSelected[mc] = true
	}

	// Vérifier quels lieux sont sélectionnés
	locationSelected := make(map[string]bool)
	for _, loc := range filterOptions.Locations {
		locationSelected[loc] = true
	}

	data := map[string]interface{}{
		"Title":           "Liste des Artistes",
		"Artists":         artists,
		"Query":           query,
		"Locations":       locationsList,
		"MinYear":         minYear,
		"MaxYear":         maxYear,
		"Member1":         memberSelected[1],
		"Member2":         memberSelected[2],
		"Member3":         memberSelected[3],
		"Member4":         memberSelected[4],
		"Member5":         memberSelected[5],
		"LocationSelected": locationSelected,
		"APIError":        apiError,
	}

	renderTemplate(w, "artists.html", data)
}

// ArtistDetailHandler gère la page de détails d'un artiste
func ArtistDetailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RenderError(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	// Extraire l'ID de l'URL (format: /artist/1)
	artistIDStr := strings.TrimPrefix(r.URL.Path, "/artist/")
	
	// Valider l'ID
	artistID, err := strconv.Atoi(artistIDStr)
	if err != nil || artistID <= 0 {
		utils.RenderError(w, http.StatusBadRequest, "ID d'artiste invalide")
		return
	}

	// Récupérer les détails complets de l'artiste
	detail, err := apiClient.FetchArtistDetail(artistID)
	if err != nil {
		utils.RenderError(w, http.StatusNotFound, "Artiste non trouvé")
		return
	}

	data := map[string]interface{}{
		"Title":  "Détails de " + detail.Name,
		"Artist": detail,
	}

	renderTemplate(w, "artists_details.html", data)
}
