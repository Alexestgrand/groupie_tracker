package handlers

import (
	"net/http"
)

// MapHandler gère la page de la carte interactive
func MapHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Carte Interactive",
	}

	renderTemplate(w, "map.html", data)
}
