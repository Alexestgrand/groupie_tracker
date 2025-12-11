package models

// Artist représente un artiste (adapté pour Spotify)
type Artist struct {
	ID         string   `json:"id"`          // ID Spotify
	Image      string   `json:"image"`       // URL de l'image
	Name       string   `json:"name"`        // Nom de l'artiste
	Genres     []string `json:"genres"`      // Genres musicaux
	Popularity int      `json:"popularity"`  // Popularité (0-100)
	SpotifyURL string   `json:"spotify_url"` // Lien Spotify
}

type Location struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
	Dates     string   `json:"dates"`
}

type Date struct {
	ID    int      `json:"id"`
	Dates []string `json:"dates"`
}

type Relation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

// ArtistDetail représente un artiste avec toutes ses données complètes
type ArtistDetail struct {
	Artist
	Followers   int                    `json:"followers"` // Nombre de followers
	SpotifyInfo map[string]interface{} `json:"-"`         // Informations supplémentaires
}

type FilterOptions struct {
	MinYear       int
	MaxYear       int
	MemberCount   []int
	Locations     []string
	FirstAlbumMin string
	FirstAlbumMax string
}
