package handlers

import (
	"html/template"
	"log"
	"net/http"
)

// renderTemplate rend un template HTML avec les données fournies
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	// Parser les templates avec le layout de base
	templates := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/"+tmpl,
	))

	// Exécuter le template layout qui contient le bloc "content"
	// Le bloc "content" sera rempli par le template enfant (home.html, artists.html, etc.)
	if err := templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("Erreur lors du rendu du template %s: %v", tmpl, err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}
