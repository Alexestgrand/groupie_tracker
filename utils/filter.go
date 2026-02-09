package utils

import (
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
		// Filtrer par année de création
		if options.MinYear > 0 && artist.CreationDate < options.MinYear {
			continue
		}
		if options.MaxYear > 0 && artist.CreationDate > options.MaxYear {
			continue
		}

		// Filtrer par nombre de membres
		if len(options.MemberCount) > 0 {
			memberCount := len(artist.Members)
			found := false
			for _, mc := range options.MemberCount {
				if memberCount == mc {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filtrer par premier album
		if options.FirstAlbumMin != "" || options.FirstAlbumMax != "" {
			albumDate, err := parseDate(artist.FirstAlbum)
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

		// Note: Le filtrage par lieu nécessite les relations, donc on le fait dans le handler
		filtered = append(filtered, artist)
	}

	return filtered
}

// parseDate parse une date au format DD-MM-YYYY
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("02-01-2006", dateStr)
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
