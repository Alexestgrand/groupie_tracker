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
	// Vérifier si le token est encore valide (avec une marge de 1 minute)
	if s.accessToken != "" && time.Now().Add(1*time.Minute).Before(s.tokenExpiry) {
		return nil
	}

	// Vérifier que les credentials sont configurés
	if s.clientID == "" || s.clientID == "your_client_id_here" {
		return fmt.Errorf("SPOTIFY_CLIENT_ID non configuré - définissez la variable d'environnement")
	}
	if s.clientSecret == "" || s.clientSecret == "your_client_secret_here" {
		return fmt.Errorf("SPOTIFY_CLIENT_SECRET non configuré - définissez la variable d'environnement")
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
		return fmt.Errorf("erreur réseau lors de l'authentification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("credentials Spotify invalides - vérifiez SPOTIFY_CLIENT_ID et SPOTIFY_CLIENT_SECRET")
		}
		return fmt.Errorf("erreur d'authentification Spotify (code %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp SpotifyTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("erreur lors du parsing de la réponse d'authentification: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return fmt.Errorf("token d'accès vide reçu de Spotify")
	}

	s.accessToken = tokenResp.AccessToken
	// Expiry avec une marge de sécurité
	expirySeconds := tokenResp.ExpiresIn
	if expirySeconds == 0 {
		expirySeconds = 3600 // Par défaut 1 heure
	}
	s.tokenExpiry = time.Now().Add(time.Duration(expirySeconds) * time.Second)

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

	// Récupérer les artistes depuis Spotify
	spotifyArtists, err := s.FetchPopularArtists()
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des artistes Spotify: %w", err)
	}

	if len(spotifyArtists) == 0 {
		return nil, fmt.Errorf("aucun artiste récupéré depuis Spotify")
	}

	// Convertir les artistes Spotify en modèles Artist avec année de création
	artists := make([]models.Artist, 0, len(spotifyArtists))
	for i, sa := range spotifyArtists {
		artist := models.Artist{
			ID:            i + 1,
			Name:          sa.Name,
			Image:         "",
			Members:       []string{},
			CreationDate:  0,
			FirstAlbum:    "",
			FirstAlbumDate: "",
			Locations:     "",
			ConcertDates:  "",
			Relations:     "",
			Genres:        []string{},
		}
		
		// Récupérer l'image la plus grande disponible
		if len(sa.Images) > 0 {
			artist.Image = sa.Images[0].URL
		}
		
		// Copier les genres
		if len(sa.Genres) > 0 {
			artist.Genres = make([]string, len(sa.Genres))
			copy(artist.Genres, sa.Genres)
		}
		
		// Récupérer le premier album pour obtenir l'année de création
		firstAlbum, firstAlbumDate, creationYear := s.getFirstAlbumAndYear(sa.ID)
		if firstAlbum != "" {
			artist.FirstAlbum = firstAlbum
		}
		if firstAlbumDate != "" {
			artist.FirstAlbumDate = firstAlbumDate
		}
		if creationYear > 0 {
			artist.CreationDate = creationYear
		}
		
		artists = append(artists, artist)
	}

	// Mettre en cache
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

			// Mettre à jour l'année de création et le premier album si pas déjà défini
			if detail.Artist.CreationDate == 0 || detail.Artist.FirstAlbum == "" {
				firstAlbum, firstAlbumDate, creationYear := s.getFirstAlbumAndYear(full.ID)
				if creationYear > 0 {
					detail.Artist.CreationDate = creationYear
				}
				if firstAlbum != "" {
					detail.Artist.FirstAlbum = firstAlbum
				}
				if firstAlbumDate != "" {
					detail.Artist.FirstAlbumDate = firstAlbumDate
				}
			}
			
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
		return nil, fmt.Errorf("authentification requise: %w", err)
	}

	// Limiter le nombre de résultats à 50 (limite API Spotify)
	if limit > 50 {
		limit = 50
	}
	if limit < 1 {
		limit = 1
	}

	searchURL := fmt.Sprintf("%s/search?q=%s&type=artist&limit=%d&market=FR", 
		SpotifyAPIURL, url.QueryEscape(query), limit)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de la requête: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erreur réseau lors de la recherche: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		// Token expiré, réessayer une fois
		s.accessToken = "" // Forcer le renouvellement
		if err := s.authenticate(); err != nil {
			return nil, fmt.Errorf("erreur de ré-authentification: %w", err)
		}
		// Réessayer la requête
		req.Header.Set("Authorization", "Bearer "+s.accessToken)
		resp, err = s.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("erreur réseau lors de la recherche (retry): %w", err)
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erreur API Spotify (code %d): %s", resp.StatusCode, string(body))
	}

	var searchResp SpotifySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing de la réponse: %w", err)
	}

	// Filtrer les artistes sans nom
	var validArtists []SpotifyArtist
	for _, artist := range searchResp.Artists.Items {
		if artist.Name != "" && artist.ID != "" {
			validArtists = append(validArtists, artist)
		}
	}

	return validArtists, nil
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

// getFirstAlbumAndYear récupère le premier album d'un artiste, sa date de sortie et l'année de création
func (s *SpotifyClient) getFirstAlbumAndYear(spotifyArtistID string) (string, string, int) {
	if err := s.authenticate(); err != nil {
		return "", "", 0
	}
	
	// Récupérer les albums triés par date (les plus anciens en premier)
	u := fmt.Sprintf("%s/artists/%s/albums?limit=50&market=FR&include_groups=album", SpotifyAPIURL, spotifyArtistID)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return "", "", 0
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", "", 0
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", "", 0
	}
	
	var data spotifyArtistAlbumsResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", 0
	}
	
	if len(data.Items) == 0 {
		return "", "", 0
	}
	
	// Trouver le premier album (le plus ancien)
	oldestAlbum := data.Items[0]
	oldestDate := ""
	oldestYear := 0
	
	for _, album := range data.Items {
		if album.ReleaseDate == "" {
			continue
		}
		// Parse la date (format peut être YYYY, YYYY-MM, ou YYYY-MM-DD)
		var year int
		if len(album.ReleaseDate) >= 4 {
			fmt.Sscanf(album.ReleaseDate[:4], "%d", &year)
		}
		if year > 0 && (oldestYear == 0 || year < oldestYear) {
			oldestYear = year
			oldestAlbum = album
			oldestDate = album.ReleaseDate
		}
	}
	
	if oldestYear > 0 {
		return oldestAlbum.Name, oldestDate, oldestYear
	}
	
	// Si pas d'année trouvée, retourner le premier album sans année
	if oldestAlbum.ReleaseDate != "" {
		return oldestAlbum.Name, oldestAlbum.ReleaseDate, 0
	}
	return oldestAlbum.Name, "", 0
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

// FetchPopularArtists récupère une liste d'artistes populaires depuis Spotify
func (s *SpotifyClient) FetchPopularArtists() ([]SpotifyArtist, error) {
	// Vérifier l'authentification une seule fois
	if err := s.authenticate(); err != nil {
		return nil, fmt.Errorf("erreur d'authentification Spotify: %w", err)
	}

	// Liste de requêtes variées pour obtenir une diversité d'artistes
	queries := []string{
		"year:2020-2025",           // Artistes récents
		"genre:rock",               // Rock
		"genre:pop",                // Pop
		"genre:hip-hop",            // Hip-hop/Rap
		"genre:jazz",               // Jazz
		"genre:electronic",          // Électronique
		"genre:indie",              // Indie
		"genre:metal",              // Metal
		"genre:country",            // Country
		"genre:reggae",             // Reggae
		"genre:blues",              // Blues
		"genre:classical",          // Classique
		"tag:new",                  // Nouveautés
		"tag:hipster",              // Artistes émergents
	}

	var allArtists []SpotifyArtist
	seen := make(map[string]bool)
	targetCount := 100 // Objectif : au moins 100 artistes

	// Première passe : récupérer des artistes avec différentes requêtes
	for _, query := range queries {
		if len(allArtists) >= targetCount {
			break
		}
		
		artists, err := s.SearchArtists(query, 20)
		if err != nil {
			// Continuer avec la requête suivante en cas d'erreur
			continue
		}
		
		for _, artist := range artists {
			if !seen[artist.ID] && artist.Name != "" {
				allArtists = append(allArtists, artist)
				seen[artist.ID] = true
			}
		}
	}

	// Si on n'a pas assez d'artistes, utiliser des recherches par nom d'artistes populaires
	if len(allArtists) < 50 {
		popularArtistNames := []string{
			"The Weeknd", "Taylor Swift", "Ed Sheeran", "Drake", "Ariana Grande",
			"Billie Eilish", "Post Malone", "Dua Lipa", "Bad Bunny", "The Beatles",
			"Queen", "Michael Jackson", "Elvis Presley", "Madonna", "Eminem",
			"Rihanna", "Beyoncé", "Adele", "Bruno Mars", "Justin Bieber",
			"Coldplay", "Imagine Dragons", "Maroon 5", "OneRepublic", "The Chainsmokers",
			"Calvin Harris", "David Guetta", "Martin Garrix", "Avicii", "Skrillex",
		}

		for _, name := range popularArtistNames {
			if len(allArtists) >= targetCount {
				break
			}
			
			artists, err := s.SearchArtists(name, 5)
			if err != nil {
				continue
			}
			
			for _, artist := range artists {
				if !seen[artist.ID] && artist.Name != "" {
					allArtists = append(allArtists, artist)
					seen[artist.ID] = true
				}
			}
		}
	}

	// Dernière tentative : recherche générique si toujours pas assez
	if len(allArtists) < 20 {
		artists, err := s.SearchArtists("artist", 50)
		if err != nil {
			return nil, fmt.Errorf("impossible de récupérer des artistes: %w", err)
		}
		
		for _, artist := range artists {
			if !seen[artist.ID] && artist.Name != "" {
				allArtists = append(allArtists, artist)
				seen[artist.ID] = true
			}
		}
	}

	if len(allArtists) == 0 {
		return nil, fmt.Errorf("aucun artiste récupéré depuis Spotify - vérifiez vos credentials")
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
