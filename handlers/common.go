package handlers

import (
	"groupie-tracker-ng/api"
)

var (
	// Client Spotify uniquement (plus d'API Groupie Trackers)
	spotifyClient = api.NewSpotifyClient(
		"3d1c929d80ea451898e44f68db5d7b4d",
		"95b58a507fba44aab798f1d6ed73a9b4",
	)
)
