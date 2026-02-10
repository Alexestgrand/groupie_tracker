package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// formatDuration convertit des millisecondes en "m:ss"
func formatDuration(ms int) string {
	if ms <= 0 {
		return "0:00"
	}
	sec := ms / 1000
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}

// formatNumber formate un nombre avec des séparateurs de milliers
func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%.1fM", float64(n)/1000000)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	funcMap := template.FuncMap{
		"join":           strings.Join,
		"urlpath":        url.PathEscape,
		"formatDuration": formatDuration,
		"formatNumber":   formatNumber,
	}
	// 1. Parser les templates (join + urlpath pour les listes et liens)
	templates, err := template.New("").Funcs(funcMap).ParseFiles(
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
