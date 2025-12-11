package handlers

import (
	"html/template"
	"log"
	"net/http"
)

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	// Parser les templates avec gestion d'erreur
	templates, err := template.ParseFiles(
		"templates/layout.html",
		"templates/"+tmpl,
	)
	if err != nil {
		log.Printf("Erreur lors du parsing du template %s: %v", tmpl, err)
		http.Error(w, "Erreur interne du serveur - Template non trouvé", http.StatusInternalServerError)
		return
	}

	// Exécuter le template avec gestion d'erreur
	if err := templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("Erreur lors du rendu du template %s: %v", tmpl, err)
		http.Error(w, "Erreur interne du serveur - Erreur de rendu", http.StatusInternalServerError)
		return
	}
}
