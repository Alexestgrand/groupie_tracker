package handlers

import (
	"net/http"
	"strconv"

	"groupie-tracker-ng/models"
	"groupie-tracker-ng/utils"
)

// GimsHandler gère la route spéciale "gims"
func GimsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RenderError(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	// Rechercher l'artiste "GIMS" dans l'API
	artist, err := apiClient.FindArtistByName("GIMS")
	if err != nil {
		// Essayer des variantes
		variants := []string{"Gims", "Maître Gims", "Maitre Gims"}
		var foundArtist *models.Artist

		for _, variant := range variants {
			foundArtist, err = apiClient.FindArtistByName(variant)
			if err == nil {
				break
			}
		}

		if foundArtist == nil {
			utils.RenderError(w, http.StatusNotFound, "Artiste GIMS non trouvé")
			return
		}

		artist = foundArtist
	}

	// Rediriger vers la page de détails de GIMS
	http.Redirect(w, r, "/artist/"+strconv.Itoa(artist.ID), http.StatusSeeOther)
}
