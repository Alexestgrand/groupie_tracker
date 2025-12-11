package handlers

import (
	"net/http"

	"groupie-tracker-ng/api"
)

// GimsHandler gère la route spéciale "gims"
func GimsHandler(w http.ResponseWriter, r *http.Request) {
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
			// Si toujours pas trouvé, rediriger vers la liste des artistes
			http.Redirect(w, r, "/artists", http.StatusSeeOther)
			return
		}

		spotifyArtist = foundArtist
	}

	// Rediriger vers la page de détails de GIMS
	http.Redirect(w, r, "/artist/"+spotifyArtist.ID, http.StatusSeeOther)
}
