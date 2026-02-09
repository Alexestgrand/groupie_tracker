package handlers

import (
	"net/http"
	"strings"

	"groupie-tracker-ng/api"
	"groupie-tracker-ng/utils"
)

// MapLocationData pour la carte : lieu + coordonnées + artistes et dates
type MapLocationData struct {
	Location string
	Dates    []string
	Artists  []string
	Lat      float64
	Lng      float64
}

// MapHandler gère la page de la carte interactive (données Groupie : lieux de concerts)
func MapHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RenderError(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	locationsList := []*MapLocationData{}

	groupieArtists, errArtists := api.FetchGroupieArtists()
	relations, errRel := api.FetchGroupieRelations()
	if errArtists == nil && errRel == nil {
		// Construire une map ID -> nom d'artiste
		idToName := make(map[int]string)
		for _, a := range groupieArtists {
			idToName[a.ID] = a.Name
		}
		// Agréger par lieu : lieu -> { dates, artistes }
		byLocation := make(map[string]*MapLocationData)
		for _, rel := range relations {
			artistName := idToName[rel.ID]
			for loc, dates := range rel.DatesLocations {
				locNorm := strings.TrimSpace(loc)
				if locNorm == "" {
					continue
				}
				if byLocation[locNorm] == nil {
					lat, lng, _ := utils.GetCoords(locNorm)
					byLocation[locNorm] = &MapLocationData{
						Location: locNorm,
						Dates:    []string{},
						Artists:  []string{},
						Lat:      lat,
						Lng:      lng,
					}
				}
				byLocation[locNorm].Dates = append(byLocation[locNorm].Dates, dates...)
				// Éviter les doublons d'artistes
				hasArtist := false
				for _, a := range byLocation[locNorm].Artists {
					if a == artistName {
						hasArtist = true
						break
					}
				}
				if !hasArtist {
					byLocation[locNorm].Artists = append(byLocation[locNorm].Artists, artistName)
				}
			}
		}
		for _, v := range byLocation {
			locationsList = append(locationsList, v)
		}
	}

	data := map[string]interface{}{
		"Title":     "Carte interactive",
		"Locations": locationsList,
	}

	renderTemplate(w, "map.html", data)
}
