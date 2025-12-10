package handlers

import (
	"net/http"
)

// LocationHandler gère la page listant les concerts à un lieu spécifique
func LocationHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}
