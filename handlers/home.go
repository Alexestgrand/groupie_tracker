package handlers

import (
	"net/http"
)

// HomeHandler gère la page d'accueil
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"Title": "Accueil",
	}

	renderTemplate(w, "home.html", data)
}
