package main

import (
	"groupie-tracker-ng/handlers"
	"log"
	"net/http"
)

func main() {

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/artists", handlers.ArtistsHandler)
	http.HandleFunc("/artist/", handlers.ArtistDetailHandler)
	http.HandleFunc("/search", handlers.SearchHandler)
	http.HandleFunc("/suggestions", handlers.SuggestionsHandler)
	http.HandleFunc("/map", handlers.MapHandler)
	http.HandleFunc("/location/", handlers.LocationHandler)

	port := ":8000"
	log.Printf("🚀 Serveur démarré sur http://localhost%s", port)
	log.Printf("🚀 Serveur également accessible sur http://[::1]%s", port)

	//suppression des messages inutiles

	log.Fatal(http.ListenAndServe(port, nil))
}
