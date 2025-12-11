package utils

import (
	"net/url"

	"groupie-tracker-ng/models"
)

func parseFilterOptions(queryParams url.Values) models.FilterOptions {
	return models.FilterOptions{}
}

func filterArtists(artists []models.Artist, options models.FilterOptions) []models.Artist {
	return artists
}
