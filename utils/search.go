package utils

import (
	"groupie-tracker-ng/models"
	"strings"
)

// SearchArtists recherche des artistes selon une requête (nom d'artiste ou membre)
func SearchArtists(artists []models.Artist, query string) []models.Artist {
	queryLower := strings.ToLower(query)
	filteredArtists := []models.Artist{}
	
	for _, artist := range artists {
		// Rechercher dans le nom de l'artiste
		if strings.Contains(strings.ToLower(artist.Name), queryLower) {
			filteredArtists = append(filteredArtists, artist)
			continue
		}
		
		// Rechercher dans les membres
		for _, member := range artist.Members {
			if strings.Contains(strings.ToLower(member), queryLower) {
				filteredArtists = append(filteredArtists, artist)
				break
			}
		}
	}
	
	return filteredArtists
}

// GetSuggestions génère des suggestions pour la barre de recherche
func GetSuggestions(artists []models.Artist, query string) []models.Artist {
	queryLower := strings.ToLower(query)
	suggestions := []models.Artist{}
	seen := make(map[int]bool)
	
	for _, artist := range artists {
		if seen[artist.ID] {
			continue
		}
		
		// Rechercher dans le nom de l'artiste
		if strings.Contains(strings.ToLower(artist.Name), queryLower) {
			suggestions = append(suggestions, artist)
			seen[artist.ID] = true
			continue
		}
		
		// Rechercher dans les membres
		for _, member := range artist.Members {
			if strings.Contains(strings.ToLower(member), queryLower) {
				if !seen[artist.ID] {
					suggestions = append(suggestions, artist)
					seen[artist.ID] = true
				}
				break
			}
		}
	}
	
	// Limiter à 10 suggestions
	if len(suggestions) > 10 {
		suggestions = suggestions[:10]
	}
	
	return suggestions
}
