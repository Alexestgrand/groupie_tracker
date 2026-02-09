package handlers

import (
	"net/http"

	"groupie-tracker-ng/api"
	"groupie-tracker-ng/utils"
)

// GimsHandler gère la route spéciale "gims"
func GimsHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier que la méthode HTTP est GET
	if r.Method != http.MethodGet {
		utils.RenderError(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	// Rechercher l'artiste "GIMS" sur Spotify
	spotifyArtist, err := spotifyClient.SearchArtist("GIMS")
	if err != nil {
		// Si GIMS n'est pas trouvé, essayer des variantes
		variants := []string{"Gims", "Maître Gims", "Maitre Gims"}
		var foundArtist *api.SpotifyArtist

		for _, variant := range variants {
			foundArtist, err = spotifyClient.SearchArtist(variant)
			if err == nil {
				break
			}
		}

		if foundArtist == nil {
			// Si toujours pas trouvé, afficher une erreur 404
			utils.RenderError(w, http.StatusNotFound, "Artiste GIMS non trouvé sur Spotify")
			return
		}

		spotifyArtist = foundArtist
	}

	// Valider que l'artiste a un ID valide
	if spotifyArtist.ID == "" {
		utils.RenderError(w, http.StatusInternalServerError, "Erreur lors de la récupération de l'artiste")
		return
	}

	// Rediriger vers la page de détails de GIMS
	http.Redirect(w, r, "/artist/"+spotifyArtist.ID, http.StatusSeeOther)
}
