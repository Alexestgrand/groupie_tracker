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
	"sync"
	"time"

	"groupie-tracker-ng/models"
)

const (
	SpotifyAuthURL = "https://accounts.spotify.com/api/token"
	SpotifyAPIURL  = "https://api.spotify.com/v1"
)

const artistsCacheTTL = 5 * time.Minute

type SpotifyClient struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
	accessToken  string
	tokenExpiry  time.Time
	// Cache liste artistes pour que l'ID reste stable (détail par ID)
	mu           sync.Mutex
	cachedArtists []models.Artist
	cacheTime     time.Time
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

// Réponses API pour top tracks, albums, related artists
type spotifyTopTracksResp struct {
	Tracks []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		DurationMs  int    `json:"duration_ms"`
		PreviewURL  string `json:"preview_url"`
		ExternalURLs struct { Spotify string `json:"spotify"` } `json:"external_urls"`
		Album struct {
			Name   string `json:"name"`
			Images []struct { URL string `json:"url"` } `json:"images"`
		} `json:"album"`
	} `json:"tracks"`
}

type spotifyArtistAlbumsResp struct {
	Items []struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		ReleaseDate  string `json:"release_date"`
		TotalTracks  int    `json:"total_tracks"`
		ExternalURLs struct { Spotify string `json:"spotify"` } `json:"external_urls"`
		Images       []struct { URL string `json:"url"` } `json:"images"`
	} `json:"items"`
}

type spotifyRelatedArtistsResp struct {
	Artists []struct {
		ID            string   `json:"id"`
		Name          string   `json:"name"`
		Genres        []string `json:"genres"`
		ExternalURLs  struct { Spotify string `json:"spotify"` } `json:"external_urls"`
		Images        []struct { URL string `json:"url"` } `json:"images"`
	} `json:"artists"`
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

// FetchArtists récupère la liste d'artistes populaires et la met en cache (IDs stables)
func (s *SpotifyClient) FetchArtists() ([]models.Artist, error) {
	s.mu.Lock()
	if len(s.cachedArtists) > 0 && time.Since(s.cacheTime) < artistsCacheTTL {
		out := make([]models.Artist, len(s.cachedArtists))
		copy(out, s.cachedArtists)
		s.mu.Unlock()
		return out, nil
	}
	s.mu.Unlock()

	spotifyArtists, err := s.FetchPopularArtists()
	if err != nil {
		return nil, err
	}

	artists := make([]models.Artist, 0, len(spotifyArtists))
	for i, sa := range spotifyArtists {
		artist := models.Artist{
			ID:           i + 1,
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
		if len(sa.Genres) > 0 {
			artist.Genres = sa.Genres
		}
		artists = append(artists, artist)
	}

	s.mu.Lock()
	s.cachedArtists = artists
	s.cacheTime = time.Now()
	s.mu.Unlock()
	return artists, nil
}

// FetchArtistDetail récupère les détails d'un artiste par ID (utilise le cache pour cohérence)
func (s *SpotifyClient) FetchArtistDetail(artistID int) (*models.ArtistDetail, error) {
	// Utiliser le cache pour retrouver le même artiste que sur la liste
	artists, err := s.FetchArtists()
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des artistes: %w", err)
	}

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

	// Copie pour ne pas modifier le cache
	artistCopy := *artist

	detail := &models.ArtistDetail{
		Artist:         artistCopy,
		ConcertDates:   []string{},
		Locations:      []string{},
		Relations:      make(map[string][]string),
		BirthDates:     make(map[string]string),
		DeathDates:     make(map[string]string),
		TopTracks:      []models.TrackInfo{},
		Albums:         []models.AlbumInfo{},
		RelatedArtists: []models.RelatedArtistInfo{},
	}

	// Enrichir avec l'API Spotify
	spotifyFull, _ := s.searchArtistByName(artist.Name)
	if spotifyFull != nil {
		full, _ := s.GetArtistByID(spotifyFull.ID)
		if full != nil {
			if len(full.Images) > 0 {
				artistCopy.Image = full.Images[0].URL
			}
			artistCopy.SpotifyURL = full.ExternalURLs.Spotify
			artistCopy.Genres = full.Genres
			artistCopy.Popularity = full.Popularity
			artistCopy.Followers = full.Followers.Total
			detail.Artist = artistCopy

			// Top titres, albums, artistes similaires (ignorer erreurs pour ne pas casser la page)
			if tracks, err := s.getArtistTopTracks(full.ID); err == nil && len(tracks) > 0 {
				detail.TopTracks = tracks
			}
			if albums, err := s.getArtistAlbums(full.ID); err == nil && len(albums) > 0 {
				detail.Albums = albums
			}
			if related, err := s.getRelatedArtists(full.ID); err == nil && len(related) > 0 {
				detail.RelatedArtists = related
			}
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

// getArtistTopTracks récupère les titres les plus populaires d'un artiste (market FR)
func (s *SpotifyClient) getArtistTopTracks(spotifyArtistID string) ([]models.TrackInfo, error) {
	if err := s.authenticate(); err != nil {
		return nil, err
	}
	u := fmt.Sprintf("%s/artists/%s/top-tracks?market=FR", SpotifyAPIURL, spotifyArtistID)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("top tracks: %d", resp.StatusCode)
	}
	var data spotifyTopTracksResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	out := make([]models.TrackInfo, 0, len(data.Tracks))
	for _, t := range data.Tracks {
		info := models.TrackInfo{
			Name:       t.Name,
			SpotifyURL: t.ExternalURLs.Spotify,
			AlbumName:  t.Album.Name,
			DurationMs: t.DurationMs,
			PreviewURL: t.PreviewURL,
		}
		out = append(out, info)
	}
	return out, nil
}

// getArtistAlbums récupère les albums d'un artiste (max 20)
func (s *SpotifyClient) getArtistAlbums(spotifyArtistID string) ([]models.AlbumInfo, error) {
	if err := s.authenticate(); err != nil {
		return nil, err
	}
	u := fmt.Sprintf("%s/artists/%s/albums?limit=20&market=FR", SpotifyAPIURL, spotifyArtistID)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("albums: %d", resp.StatusCode)
	}
	var data spotifyArtistAlbumsResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	out := make([]models.AlbumInfo, 0, len(data.Items))
	for _, a := range data.Items {
		img := ""
		if len(a.Images) > 0 {
			img = a.Images[0].URL
		}
		out = append(out, models.AlbumInfo{
			Name:        a.Name,
			SpotifyURL:  a.ExternalURLs.Spotify,
			ReleaseDate: a.ReleaseDate,
			ImageURL:    img,
			TotalTracks: a.TotalTracks,
		})
	}
	return out, nil
}

// getRelatedArtists récupère les artistes similaires
func (s *SpotifyClient) getRelatedArtists(spotifyArtistID string) ([]models.RelatedArtistInfo, error) {
	if err := s.authenticate(); err != nil {
		return nil, err
	}
	u := fmt.Sprintf("%s/artists/%s/related-artists", SpotifyAPIURL, spotifyArtistID)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("related: %d", resp.StatusCode)
	}
	var data spotifyRelatedArtistsResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	out := make([]models.RelatedArtistInfo, 0, len(data.Artists))
	for _, a := range data.Artists {
		img := ""
		if len(a.Images) > 0 {
			img = a.Images[0].URL
		}
		out = append(out, models.RelatedArtistInfo{
			Name:       a.Name,
			ImageURL:   img,
			SpotifyURL: a.ExternalURLs.Spotify,
			Genres:     a.Genres,
		})
	}
	return out, nil
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
