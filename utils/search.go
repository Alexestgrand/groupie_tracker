package utils

import (
	"groupie-tracker-ng/models"
	"strings"
)

// SearchArtists recherche des artistes selon une requête
func SearchArtists(artists []models.Artist, query string) []models.Artist {
	filteredArtists := []models.Artist{}
	for _, artist := range artists {
		if strings.Contains(artist.Name, query) {
			filteredArtists = append(filteredArtists, artist)
		}
	}
	return filteredArtists
}

// GetSuggestions génère des suggestions pour la barre de recherche
func GetSuggestions(artists []models.Artist, query string) []string {
	suggestions := []string{}
	for _, artist := range artists {
		if strings.Contains(artist.Name, query) {
			suggestions = append(suggestions, artist.Name)
		}
	}
	return suggestions
}
