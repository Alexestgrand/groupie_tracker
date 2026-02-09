package models

// Artist représente un artiste (API Groupie ou Spotify)
type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Locations    string   `json:"locations"`
	ConcertDates string   `json:"concertDates"`
	Relations    string   `json:"relations"`
	// Champs optionnels (ex. API Spotify)
	SpotifyURL string   `json:"-"`
	Genres     []string `json:"-"`
	Popularity int      `json:"-"`
	Followers  int      `json:"-"`
}

// Location représente les lieux de concerts d'un artiste
type Location struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
	Dates     string   `json:"dates"`
}

// Date représente les dates de concerts d'un artiste
type Date struct {
	ID    int      `json:"id"`
	Dates []string `json:"dates"`
}

// Relation représente les relations dates-lieux d'un artiste
type Relation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

// ArtistDetail représente un artiste avec toutes ses données complètes
type ArtistDetail struct {
	Artist
	ConcertDates []string            `json:"concertDates"` // Dates de concerts
	Locations    []string            `json:"locations"`    // Lieux de concerts
	Relations    map[string][]string `json:"relations"`    // Relations dates-lieux
	BirthDates   map[string]string   `json:"birthDates"`   // Dates de naissance des membres (bonus)
	DeathDates   map[string]string   `json:"deathDates"`   // Dates de mort des membres (bonus)
}

// FilterOptions représente les options de filtrage
type FilterOptions struct {
	MinYear       int      // Année de création minimum
	MaxYear       int      // Année de création maximum
	MemberCount   []int    // Nombre de membres (sélection multiple)
	Locations     []string // Lieux de concerts (sélection multiple)
	FirstAlbumMin string   // Premier album date minimum (format: DD-MM-YYYY)
	FirstAlbumMax string   // Premier album date maximum (format: DD-MM-YYYY)
}
