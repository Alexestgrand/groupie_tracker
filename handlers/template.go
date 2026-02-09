package handlers

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"strings"
)

var templateFuncs = template.FuncMap{
	"join": strings.Join,
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	// 1. Parser les templates (avec join pour les listes)
	templates, err := template.New("").Funcs(templateFuncs).ParseFiles(
		"templates/layout.html",
		"templates/"+tmpl,
	)
	if err != nil {
		log.Printf("Erreur lors du parsing du template %s: %v", tmpl, err)
		http.Error(w, "Erreur interne du serveur - Template non trouvé", http.StatusInternalServerError)
		return
	}

	// 2. Utiliser un buffer pour préparer le rendu en mémoire
	buf := new(bytes.Buffer)
	if err := templates.ExecuteTemplate(buf, "layout", data); err != nil {
		log.Printf("Erreur lors du rendu du template %s: %v", tmpl, err)
		// Ici, rien n'a été envoyé à 'w', donc on peut envoyer une erreur propre
		http.Error(w, "Erreur interne du serveur - Erreur de rendu", http.StatusInternalServerError)
		return
	}

	// 3. Si tout est ok, envoyer le contenu du buffer à la réponse
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}
