package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"groupie-tracker-ng/models"
)

const (
	SpotifyAuthURL = "https://accounts.spotify.com/api/token"
	SpotifyAPIURL  = "https://api.spotify.com/v1"
)

type SpotifyClient struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
	accessToken  string
	tokenExpiry  time.Time
}

type SpotifyTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type SpotifyArtist struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Images []struct {
		URL string `json:"url"`
	} `json:"images"`
	Genres []string `json:"genres"`
}

type SpotifySearchResponse struct {
	Artists struct {
		Items []SpotifyArtist `json:"items"`
		Total int             `json:"total"`
	} `json:"artists"`
}

type SpotifyArtistFull struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Images []struct {
		URL    string `json:"url"`
		Height int    `json:"height"`
		Width  int    `json:"width"`
	} `json:"images"`
	Genres     []string `json:"genres"`
	Popularity int      `json:"popularity"`
	Followers  struct {
		Total int `json:"total"`
	} `json:"followers"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
}

// NewClient crée un nouveau client Spotify (remplace l'ancien NewClient de Groupie)
func NewClient() *SpotifyClient {
	// Récupérer les credentials depuis les variables d'environnement
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	// Si non définies, utiliser des valeurs par défaut (à configurer)
	if clientID == "" {
		clientID = "your_client_id_here"
	}
	if clientSecret == "" {
		clientSecret = "your_client_secret_here"
	}

	return &SpotifyClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewSpotifyClient crée un nouveau client Spotify avec credentials explicites
func NewSpotifyClient(clientID, clientSecret string) *SpotifyClient {
	return &SpotifyClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// authenticate obtient un token d'accès Spotify
func (s *SpotifyClient) authenticate() error {
	// Vérifier si le token est encore valide
	if s.accessToken != "" && time.Now().Before(s.tokenExpiry) {
		return nil
	}

	// Préparer les données pour la requête
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", SpotifyAuthURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("erreur lors de la création de la requête: %w", err)
	}

	// Encoder les credentials en base64
	credentials := base64.StdEncoding.EncodeToString([]byte(s.clientID + ":" + s.clientSecret))
	req.Header.Set("Authorization", "Basic "+credentials)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erreur lors de la requête d'authentification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("erreur d'authentification Spotify (code %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp SpotifyTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("erreur lors du parsing de la réponse: %w", err)
	}

	s.accessToken = tokenResp.AccessToken
	s.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// ============================================
// MÉTHODES COMPATIBLES AVEC L'ANCIENNE API
// ============================================

// FetchArtists récupère la liste d'artistes populaires (remplace l'ancien FetchArtists)
func (s *SpotifyClient) FetchArtists() ([]models.Artist, error) {
	spotifyArtists, err := s.FetchPopularArtists()
	if err != nil {
		return nil, err
	}

	// Convertir les artistes Spotify en modèles Artist
	artists := make([]models.Artist, 0, len(spotifyArtists))
	for i, sa := range spotifyArtists {
		artist := models.Artist{
			ID:           i + 1, // Générer un ID numérique
			Name:         sa.Name,
			Image:        "",
			Members:      []string{}, // Spotify ne fournit pas les membres
			CreationDate: 0,          // Spotify ne fournit pas l'année de création
			FirstAlbum:   "",         // Spotify ne fournit pas le premier album
			Locations:    "",
			ConcertDates: "",
			Relations:    "",
		}

		// Récupérer l'image si disponible
		if len(sa.Images) > 0 {
			artist.Image = sa.Images[0].URL
		}

		artists = append(artists, artist)
	}

	return artists, nil
}

// FetchArtistDetail récupère les détails complets d'un artiste (remplace l'ancien FetchArtistDetail)
func (s *SpotifyClient) FetchArtistDetail(artistID int) (*models.ArtistDetail, error) {
	// Récupérer tous les artistes pour trouver celui avec l'ID
	artists, err := s.FetchArtists()
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des artistes: %w", err)
	}

	// Trouver l'artiste
	var artist *models.Artist
	for i := range artists {
		if artists[i].ID == artistID {
			artist = &artists[i]
			break
		}
	}

	if artist == nil {
		return nil, fmt.Errorf("artiste avec ID %d non trouvé", artistID)
	}

	// Récupérer les détails complets depuis Spotify
	fullArtist, err := s.searchArtistByName(artist.Name)
	if err != nil {
		// Si erreur, utiliser les données de base
		detail := &models.ArtistDetail{
			Artist:       *artist,
			ConcertDates: []string{},
			Locations:    []string{},
			Relations:    make(map[string][]string),
			BirthDates:   make(map[string]string),
			DeathDates:   make(map[string]string),
		}
		return detail, nil
	}

	// Récupérer les informations complètes
	spotifyFull, err := s.GetArtistByID(fullArtist.ID)
	if err != nil {
		spotifyFull = nil
	}

	// Construire l'ArtistDetail avec les données Spotify
	detail := &models.ArtistDetail{
		Artist:       *artist,
		ConcertDates: []string{}, // Spotify ne fournit pas les dates de concerts
		Locations:    []string{}, // Spotify ne fournit pas les lieux
		Relations:    make(map[string][]string), // Spotify ne fournit pas les relations
		BirthDates:   make(map[string]string),
		DeathDates:   make(map[string]string),
	}

	// Mettre à jour avec les données Spotify si disponibles
	if spotifyFull != nil {
		if len(spotifyFull.Images) > 0 {
			detail.Image = spotifyFull.Images[0].URL
		}
	}

	return detail, nil
}

// FetchRelations retourne une liste vide (Spotify ne fournit pas ces données)
func (s *SpotifyClient) FetchRelations() ([]models.Relation, error) {
	return []models.Relation{}, nil
}

// FindArtistByName recherche un artiste par son nom (remplace l'ancien FindArtistByName)
func (s *SpotifyClient) FindArtistByName(name string) (*models.Artist, error) {
	spotifyArtist, err := s.searchArtistByName(name)
	if err != nil {
		return nil, err
	}

	// Convertir en modèle Artist
	artist := &models.Artist{
		ID:           1, // ID temporaire
		Name:         spotifyArtist.Name,
		Image:        "",
		Members:      []string{},
		CreationDate: 0,
		FirstAlbum:   "",
		Locations:    "",
		ConcertDates: "",
		Relations:    "",
	}

	if len(spotifyArtist.Images) > 0 {
		artist.Image = spotifyArtist.Images[0].URL
	}

	return artist, nil
}

// ============================================
// MÉTHODES SPOTIFY ORIGINALES
// ============================================

// searchArtistByName recherche un artiste par son nom
func (s *SpotifyClient) searchArtistByName(artistName string) (*SpotifyArtist, error) {
	if err := s.authenticate(); err != nil {
		return nil, err
	}

	searchURL := fmt.Sprintf("%s/search?q=%s&type=artist&limit=1", SpotifyAPIURL, url.QueryEscape(artistName))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de la requête: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la requête: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erreur API Spotify (code %d): %s", resp.StatusCode, string(body))
	}

	var searchResp SpotifySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing: %w", err)
	}

	if len(searchResp.Artists.Items) == 0 {
		return nil, fmt.Errorf("artiste non trouvé sur Spotify")
	}

	return &searchResp.Artists.Items[0], nil
}

// SearchArtists recherche plusieurs artistes sur Spotify
func (s *SpotifyClient) SearchArtists(query string, limit int) ([]SpotifyArtist, error) {
	if err := s.authenticate(); err != nil {
		return nil, err
	}

	searchURL := fmt.Sprintf("%s/search?q=%s&type=artist&limit=%d", SpotifyAPIURL, url.QueryEscape(query), limit)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de la requête: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la requête: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erreur API Spotify (code %d): %s", resp.StatusCode, string(body))
	}

	var searchResp SpotifySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing: %w", err)
	}

	return searchResp.Artists.Items, nil
}

// GetArtistByID récupère un artiste complet par son ID Spotify
func (s *SpotifyClient) GetArtistByID(artistID string) (*SpotifyArtistFull, error) {
	if err := s.authenticate(); err != nil {
		return nil, err
	}

	artistURL := fmt.Sprintf("%s/artists/%s", SpotifyAPIURL, artistID)

	req, err := http.NewRequest("GET", artistURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de la requête: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la requête: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erreur API Spotify (code %d): %s", resp.StatusCode, string(body))
	}

	var artist SpotifyArtistFull
	if err := json.NewDecoder(resp.Body).Decode(&artist); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing: %w", err)
	}

	return &artist, nil
}

// FetchPopularArtists récupère une liste d'artistes populaires
func (s *SpotifyClient) FetchPopularArtists() ([]SpotifyArtist, error) {
	// Vérifier l'authentification une seule fois
	if err := s.authenticate(); err != nil {
		return nil, fmt.Errorf("Spotify: %w", err)
	}

	queries := []string{"rock", "pop", "rap", "jazz", "electronic", "indie", "metal"}
	var allArtists []SpotifyArtist
	seen := make(map[string]bool)

	for _, query := range queries {
		artists, err := s.SearchArtists(query, 15)
		if err != nil {
			continue
		}
		for _, artist := range artists {
			if !seen[artist.ID] {
				allArtists = append(allArtists, artist)
				seen[artist.ID] = true
			}
		}
	}

	// Si aucune requête n'a rien renvoyé, une dernière tentative large
	if len(allArtists) == 0 {
		artists, err := s.SearchArtists("artist", 50)
		if err != nil {
			return nil, fmt.Errorf("aucun artiste récupéré: %w", err)
		}
		for _, artist := range artists {
			if !seen[artist.ID] {
				allArtists = append(allArtists, artist)
				seen[artist.ID] = true
			}
		}
	}

	return allArtists, nil
}

// GetArtistInfo récupère les informations complètes d'un artiste
func (s *SpotifyClient) GetArtistInfo(artistName string) (map[string]interface{}, error) {
	spotifyArtist, err := s.searchArtistByName(artistName)
	if err != nil {
		return nil, err
	}

	// Récupérer les informations complètes
	fullArtist, err := s.GetArtistByID(spotifyArtist.ID)
	if err != nil {
		// Si erreur, utiliser les données de base
		info := map[string]interface{}{
			"spotify_id": spotifyArtist.ID,
			"genres":     spotifyArtist.Genres,
		}
		if len(spotifyArtist.Images) > 0 {
			info["image_url"] = spotifyArtist.Images[0].URL
		}
		return info, nil
	}

	info := map[string]interface{}{
		"spotify_id":  fullArtist.ID,
		"genres":      fullArtist.Genres,
		"popularity":  fullArtist.Popularity,
		"followers":   fullArtist.Followers.Total,
		"spotify_url": fullArtist.ExternalURLs.Spotify,
	}

	if len(fullArtist.Images) > 0 {
		info["image_url"] = fullArtist.Images[0].URL
	}

	return info, nil
}

// ConvertSpotifyArtistToModel convertit un SpotifyArtist en models.Artist
func (s *SpotifyClient) ConvertSpotifyArtistToModel(sa SpotifyArtist, id int) models.Artist {
	artist := models.Artist{
		ID:           id,
		Name:         sa.Name,
		Image:        "",
		Members:      []string{},
		CreationDate: 0,
		FirstAlbum:   "",
		Locations:    "",
		ConcertDates: "",
		Relations:    "",
	}

	if len(sa.Images) > 0 {
		artist.Image = sa.Images[0].URL
	}

	return artist
}

// GetSpotifyIDFromArtistID récupère l'ID Spotify à partir de l'ID interne
func (s *SpotifyClient) GetSpotifyIDFromArtistID(artistID int) (string, error) {
	artists, err := s.FetchArtists()
	if err != nil {
		return "", err
	}

	// L'ID interne est juste un index, on doit chercher par nom
	// Pour simplifier, on va utiliser une recherche
	if artistID > 0 && artistID <= len(artists) {
		artist := artists[artistID-1]
		spotifyArtist, err := s.searchArtistByName(artist.Name)
		if err == nil {
			return spotifyArtist.ID, nil
		}
	}

	return "", fmt.Errorf("ID Spotify non trouvé pour l'artiste %d", artistID)
}
