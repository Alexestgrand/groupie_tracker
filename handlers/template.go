package handlers

import (
	"html/template"
	"log"
	"net/http"
)

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {

	templates := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/"+tmpl,
	))

	// Exécuter le template "layout" (nom de base du fichier layout.html)
	if err := templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("Erreur lors du rendu du template %s: %v", tmpl, err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}
