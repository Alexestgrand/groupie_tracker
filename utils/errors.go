package utils

import (
	"html/template"
	"log"
	"net/http"
)

// ErrorData représente les données pour une page d'erreur
type ErrorData struct {
	StatusCode int
	Message    string
	Title      string
}

// RenderError affiche une page d'erreur personnalisée
func RenderError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)

	data := ErrorData{
		StatusCode: statusCode,
		Message:    message,
		Title:      getErrorTitle(statusCode),
	}

	tmpl, err := template.ParseFiles("templates/error.html")
	if err != nil {
		log.Printf("Erreur lors du parsing du template d'erreur: %v", err)
		http.Error(w, message, statusCode)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Erreur lors de l'exécution du template d'erreur: %v", err)
		http.Error(w, message, statusCode)
	}
}

// getErrorTitle retourne le titre approprié selon le code d'erreur
func getErrorTitle(statusCode int) string {
	switch statusCode {
	case http.StatusNotFound:
		return "Page non trouvée"
	case http.StatusBadRequest:
		return "Requête invalide"
	case http.StatusInternalServerError:
		return "Erreur serveur"
	default:
		return "Erreur"
	}
}

// HandleError gère les erreurs de manière centralisée
func HandleError(w http.ResponseWriter, err error, statusCode int) {
	log.Printf("Erreur: %v", err)
	RenderError(w, statusCode, err.Error())
}
