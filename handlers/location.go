package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"groupie-tracker-ng/api"
	"groupie-tracker-ng/models"
	"groupie-tracker-ng/utils"
)

// LocationHandler gère la page listant les concerts à un lieu (données Groupie)
func LocationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RenderError(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	locationPath := r.URL.Path[len("/location/"):]
	if locationPath == "" {
		utils.RenderError(w, http.StatusBadRequest, "Lieu non spécifié")
		return
	}

	location, err := url.PathUnescape(locationPath)
	if err != nil || len(location) < 1 || len(location) > 200 {
		utils.RenderError(w, http.StatusBadRequest, "Lieu invalide")
		return
	}

	var relatedArtists []models.Artist
	concertsByArtist := make(map[int][]string)

	groupieArtists, errA := api.FetchGroupieArtists()
	relations, errR := api.FetchGroupieRelations()
	if errA == nil && errR == nil {
		locationLower := strings.ToLower(location)
		for _, artist := range groupieArtists {
			var rel *models.Relation
			for i := range relations {
				if relations[i].ID == artist.ID {
					rel = &relations[i]
					break
				}
			}
			if rel == nil {
				continue
			}
			var dates []string
			for loc, d := range rel.DatesLocations {
				if strings.ToLower(loc) == locationLower {
					dates = d
					break
				}
			}
			if len(dates) > 0 {
				relatedArtists = append(relatedArtists, artist)
				concertsByArtist[artist.ID] = dates
			}
		}
	}

	data := map[string]interface{}{
		"Title":            "Concerts à " + location,
		"Location":         location,
		"Artists":          relatedArtists,
		"ConcertsByArtist": concertsByArtist,
	}

	renderTemplate(w, "location.html", data)
}
