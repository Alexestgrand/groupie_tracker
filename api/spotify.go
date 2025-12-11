package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
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

// NewSpotifyClient crée un nouveau client Spotify
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

func (s *SpotifyClient) SearchArtist(artistName string) (*SpotifyArtist, error) {
	// S'authentifier si nécessaire
	if err := s.authenticate(); err != nil {
		return nil, err
	}

	// Construire l'URL de recherche
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
	// Rechercher des artistes populaires avec différentes requêtes
	queries := []string{"rock", "pop", "rap", "jazz", "electronic", "classical", "country", "reggae"}
	var allArtists []SpotifyArtist
	seen := make(map[string]bool)

	for _, query := range queries {
		artists, err := s.SearchArtists(query, 10)
		if err != nil {
			continue // Ignorer les erreurs et continuer
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

func (s *SpotifyClient) GetArtistInfo(artistName string) (map[string]interface{}, error) {
	spotifyArtist, err := s.SearchArtist(artistName)
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
