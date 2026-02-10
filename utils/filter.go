package utils

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"groupie-tracker-ng/models"
)

// ParseFilterOptions parse les paramètres de requête en FilterOptions
func ParseFilterOptions(queryParams url.Values) models.FilterOptions {
	options := models.FilterOptions{}

	// Année de création minimum
	if minYearStr := queryParams.Get("minYear"); minYearStr != "" {
		if minYear, err := strconv.Atoi(minYearStr); err == nil {
			options.MinYear = minYear
		}
	}

	// Année de création maximum
	if maxYearStr := queryParams.Get("maxYear"); maxYearStr != "" {
		if maxYear, err := strconv.Atoi(maxYearStr); err == nil {
			options.MaxYear = maxYear
		}
	}

	// Nombre de membres (sélection multiple)
	if memberCounts := queryParams["memberCount"]; len(memberCounts) > 0 {
		options.MemberCount = make([]int, 0)
		for _, mcStr := range memberCounts {
			if mc, err := strconv.Atoi(mcStr); err == nil {
				options.MemberCount = append(options.MemberCount, mc)
			}
		}
	}

	// Lieux de concerts (sélection multiple)
	if locations := queryParams["location"]; len(locations) > 0 {
		options.Locations = locations
	}

	// Premier album date minimum
	if firstAlbumMin := queryParams.Get("firstAlbumMin"); firstAlbumMin != "" {
		options.FirstAlbumMin = firstAlbumMin
	}

	// Premier album date maximum
	if firstAlbumMax := queryParams.Get("firstAlbumMax"); firstAlbumMax != "" {
		options.FirstAlbumMax = firstAlbumMax
	}

	return options
}

// FilterArtists filtre les artistes selon les critères
func FilterArtists(artists []models.Artist, options models.FilterOptions) []models.Artist {
	filtered := make([]models.Artist, 0)

	for _, artist := range artists {
		// Filtrer par année de création (ignorer si non fournie, ex. API Spotify)
		if options.MinYear > 0 && artist.CreationDate != 0 && artist.CreationDate < options.MinYear {
			continue
		}
		if options.MaxYear > 0 && artist.CreationDate != 0 && artist.CreationDate > options.MaxYear {
			continue
		}

		// Filtrer par nombre de membres
		if len(options.MemberCount) > 0 {
			memberCount := len(artist.Members)
			
			// Si pas de membres dans les données, estimer basé sur le nom
			if memberCount == 0 {
				nameLower := strings.ToLower(artist.Name)
				// Détecter les groupes (mots-clés communs)
				groupKeywords := []string{" & ", " and ", " feat", " ft.", " feat.", " featuring", " vs ", " x ", " + "}
				isGroup := false
				for _, keyword := range groupKeywords {
					if strings.Contains(nameLower, keyword) {
						isGroup = true
						break
					}
				}
				
				// Détecter les groupes avec "The" au début (souvent des groupes)
				if strings.HasPrefix(nameLower, "the ") && len(nameLower) > 4 {
					isGroup = true
				}
				
				// Détecter les groupes avec des mots comme "band", "group", "collective", etc.
				groupWords := []string{" band", " group", " collective", " ensemble", " orchestra", " quartet", " trio"}
				for _, word := range groupWords {
					if strings.Contains(nameLower, word) {
						isGroup = true
						break
					}
				}
				
				if isGroup {
					memberCount = 2 // Groupe (au moins 2 membres)
				} else {
					memberCount = 1 // Probablement solo
				}
			}
			
			found := false
			for _, mc := range options.MemberCount {
				if mc == 5 && memberCount >= 5 {
					found = true
					break
				} else if memberCount == mc {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filtrer par premier album (utiliser FirstAlbumDate si disponible, sinon essayer de parser FirstAlbum)
		if options.FirstAlbumMin != "" || options.FirstAlbumMax != "" {
			var albumDate time.Time
			var err error
			
			// Essayer d'abord avec FirstAlbumDate (format Spotify: YYYY, YYYY-MM, ou YYYY-MM-DD)
			if artist.FirstAlbumDate != "" {
				albumDate, err = parseSpotifyDate(artist.FirstAlbumDate)
			} else if artist.FirstAlbum != "" {
				// Fallback: essayer de parser depuis le nom ou autre source
				albumDate, err = parseDate(artist.FirstAlbum)
			}
			
			if err != nil {
				continue // Ignorer si la date est invalide
			}

			if options.FirstAlbumMin != "" {
				minDate, err := parseDate(options.FirstAlbumMin)
				if err == nil && albumDate.Before(minDate) {
					continue
				}
			}

			if options.FirstAlbumMax != "" {
				maxDate, err := parseDate(options.FirstAlbumMax)
				if err == nil && albumDate.After(maxDate) {
					continue
				}
			}
		}

		// Filtrer par lieu (villes/pays populaires)
		// Note: Comme Spotify ne fournit pas directement les lieux, on utilise une correspondance approximative
		// basée sur le nom de l'artiste et les genres pour déterminer une origine probable
		if len(options.Locations) > 0 {
			hasLocation := false
			nameLower := strings.ToLower(artist.Name)
			
			// Correspondances approximatives basées sur le nom et les genres
			locationKeywords := map[string][]string{
				"paris":     {"french", "france", "français", "paris"},
				"lyon":      {"french", "france", "français"},
				"marseille": {"french", "france", "français"},
				"london":    {"british", "uk", "england", "english"},
				"manchester": {"british", "uk", "england"},
				"new york":   {"american", "usa", "us", "hip-hop", "rap"},
				"los angeles": {"american", "usa", "us", "california"},
				"berlin":     {"german", "germany", "deutschland", "electronic"},
				"madrid":     {"spanish", "spain", "español"},
				"barcelona":  {"spanish", "spain", "catalan"},
				"rome":       {"italian", "italy", "italia"},
				"milan":      {"italian", "italy"},
				"amsterdam":  {"dutch", "netherlands", "holland"},
				"tokyo":      {"japanese", "japan", "j-pop"},
				"seoul":      {"korean", "korea", "k-pop"},
			}
			
			for _, selectedLoc := range options.Locations {
				locLower := strings.ToLower(selectedLoc)
				
				// Vérifier si le nom de l'artiste contient le lieu
				if strings.Contains(nameLower, locLower) {
					hasLocation = true
					break
				}
				
				// Vérifier les mots-clés associés au lieu
				if keywords, ok := locationKeywords[locLower]; ok {
					for _, keyword := range keywords {
						// Vérifier dans le nom
						if strings.Contains(nameLower, keyword) {
							hasLocation = true
							break
						}
						// Vérifier dans les genres
						for _, genre := range artist.Genres {
							if strings.Contains(strings.ToLower(genre), keyword) {
								hasLocation = true
								break
							}
						}
						if hasLocation {
							break
						}
					}
				}
				
				if hasLocation {
					break
				}
			}
			
			if !hasLocation {
				continue
			}
		}

		filtered = append(filtered, artist)
	}

	return filtered
}

// parseDate parse une date au format DD-MM-YYYY ou YYYY-MM-DD
func parseDate(dateStr string) (time.Time, error) {
	// Essayer d'abord le format DD-MM-YYYY
	if t, err := time.Parse("02-01-2006", dateStr); err == nil {
		return t, nil
	}
	// Essayer le format YYYY-MM-DD
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		return t, nil
	}
	// Essayer juste l'année YYYY
	if t, err := time.Parse("2006", dateStr); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("format de date non reconnu: %s", dateStr)
}

// parseSpotifyDate parse une date Spotify (YYYY, YYYY-MM, ou YYYY-MM-DD)
func parseSpotifyDate(dateStr string) (time.Time, error) {
	if len(dateStr) == 4 {
		// Format YYYY
		return time.Parse("2006", dateStr)
	} else if len(dateStr) == 7 {
		// Format YYYY-MM
		return time.Parse("2006-01", dateStr)
	} else if len(dateStr) >= 10 {
		// Format YYYY-MM-DD (prendre les 10 premiers caractères)
		return time.Parse("2006-01-02", dateStr[:10])
	}
	return time.Time{}, fmt.Errorf("format de date Spotify non reconnu: %s", dateStr)
}

// FilterArtistsByLocation filtre les artistes qui ont des concerts dans les lieux spécifiés
func FilterArtistsByLocation(artists []models.Artist, relations []models.Relation, locations []string) []models.Artist {
	if len(locations) == 0 {
		return artists
	}

	// Créer un map pour vérifier rapidement si un lieu est dans la liste
	locationMap := make(map[string]bool)
	for _, loc := range locations {
		locationMap[strings.ToLower(loc)] = true
	}

	filtered := make([]models.Artist, 0)

	for _, artist := range artists {
		// Trouver les relations de cet artiste
		var artistRelation *models.Relation
		for i := range relations {
			if relations[i].ID == artist.ID {
				artistRelation = &relations[i]
				break
			}
		}

		if artistRelation == nil {
			continue
		}

		// Vérifier si l'artiste a des concerts dans les lieux demandés
		hasLocation := false
		for location := range artistRelation.DatesLocations {
			locationLower := strings.ToLower(location)
			if locationMap[locationLower] {
				hasLocation = true
				break
			}
		}

		if hasLocation {
			filtered = append(filtered, artist)
		}
	}

	return filtered
}

// FilterArtistsByGenres filtre les artistes par genres (alternative aux lieux pour Spotify)
func FilterArtistsByGenres(artists []models.Artist, genres []string) []models.Artist {
	if len(genres) == 0 {
		return artists
	}

	// Créer un map pour vérifier rapidement si un genre est dans la liste
	genreMap := make(map[string]bool)
	for _, genre := range genres {
		genreMap[strings.ToLower(genre)] = true
	}

	filtered := make([]models.Artist, 0)

	for _, artist := range artists {
		// Vérifier si l'artiste a au moins un des genres demandés
		hasGenre := false
		for _, genre := range artist.Genres {
			genreLower := strings.ToLower(genre)
			if genreMap[genreLower] {
				hasGenre = true
				break
			}
		}

		if hasGenre {
			filtered = append(filtered, artist)
		}
	}

	return filtered
}

// GetUniqueGenres récupère la liste des genres uniques de tous les artistes
func GetUniqueGenres(artists []models.Artist) []string {
	genreMap := make(map[string]bool)
	for _, artist := range artists {
		for _, genre := range artist.Genres {
			genreMap[genre] = true
		}
	}

	genres := make([]string, 0, len(genreMap))
	for genre := range genreMap {
		genres = append(genres, genre)
	}
	return genres
}

// GetPopularLocations retourne une liste de lieux populaires pour le filtre
func GetPopularLocations() []string {
	return []string{
		"Paris", "Lyon", "Marseille", "Toulouse", "Nice", "Nantes", "Strasbourg", "Montpellier",
		"Bordeaux", "Lille", "Rennes", "Reims", "Le Havre", "Saint-Étienne", "Toulon",
		"London", "Manchester", "Liverpool", "Birmingham", "Glasgow", "Edinburgh",
		"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia",
		"Berlin", "Munich", "Hamburg", "Cologne", "Frankfurt", "Stuttgart",
		"Madrid", "Barcelona", "Valencia", "Seville", "Bilbao",
		"Rome", "Milan", "Naples", "Turin", "Palermo",
		"Amsterdam", "Rotterdam", "Utrecht", "The Hague",
		"Brussels", "Antwerp", "Ghent",
		"Vienna", "Zurich", "Geneva", "Basel",
		"Stockholm", "Oslo", "Copenhagen", "Helsinki",
		"Warsaw", "Prague", "Budapest", "Bucharest",
		"Moscow", "Saint Petersburg",
		"Tokyo", "Seoul", "Beijing", "Shanghai", "Hong Kong",
		"Sydney", "Melbourne", "Auckland",
		"Toronto", "Montreal", "Vancouver",
		"Mexico City", "São Paulo", "Rio de Janeiro", "Buenos Aires",
	}
}
