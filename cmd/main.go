package main

import (
	"log"
	"net/http"

	"groupie-tracker-ng/handlers"
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

	port := ":8080"
	log.Printf("🚀 Serveur démarré sur http://localhost%s", port)
	log.Printf("📁 Fichiers statiques servis depuis ./static")
	log.Printf("📄 Templates servis depuis ./templates")

	log.Fatal(http.ListenAndServe(port, nil))
}
