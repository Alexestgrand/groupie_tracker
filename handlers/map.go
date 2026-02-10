package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"

	"groupie-tracker-ng/utils"
)

// MapLocationData pour la carte : lieu + coordonnées + artistes et dates (non utilisé avec Spotify seul)
type MapLocationData struct {
	Location string
	Dates    []string
	Artists  []string
	Lat      float64
	Lng      float64
}

// mapLocJSON pour le script JS (champs en minuscules pour JSON)
type mapLocJSON struct {
	Location string `json:"location"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Artists  string `json:"artists"`
	Dates    string `json:"dates"`
}

// MapHandler gère la page de la carte interactive.
// L'API Spotify ne fournit pas de lieux ni dates de concerts ; la carte est affichée à titre indicatif.
func MapHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RenderError(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	// Source unique : Spotify — pas de données lieux/concerts
	locationsList := []*MapLocationData{}
	locationsJSON := "[]"
	if len(locationsList) > 0 {
		js := make([]mapLocJSON, 0, len(locationsList))
		for _, loc := range locationsList {
			if loc == nil {
				continue
			}
			artists := ""
			dates := ""
			for i, a := range loc.Artists {
				if i > 0 { artists += ", " }
				artists += a
			}
			for i, d := range loc.Dates {
				if i > 0 { dates += ", " }
				dates += d
			}
			js = append(js, mapLocJSON{Location: loc.Location, Lat: loc.Lat, Lng: loc.Lng, Artists: artists, Dates: dates})
		}
		if b, err := json.Marshal(js); err == nil {
			locationsJSON = string(b)
		}
	}

	data := map[string]interface{}{
		"Title":         "Carte interactive",
		"Locations":     locationsList,
		"LocationsJSON": template.JS(locationsJSON),
	}

	renderTemplate(w, "map.html", data)
}
