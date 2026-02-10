package handlers

import (
	"net/http"

	"groupie-tracker-ng/utils"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier que la méthode HTTP est GET
	if r.Method != http.MethodGet {
		utils.RenderError(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	// Vérifier que l'URL est exactement "/"
	if r.URL.Path != "/" {
		utils.RenderError(w, http.StatusNotFound, "Page non trouvée")
		return
	}

	data := map[string]interface{}{
		"Title": "Accueil",
	}

	renderTemplate(w, "home.html", data)
}
