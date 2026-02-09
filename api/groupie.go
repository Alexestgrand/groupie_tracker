package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"groupie-tracker-ng/models"
)

const groupieBase = "https://groupietrackers.herokuapp.com/api"

var groupieClient = &http.Client{Timeout: 15 * time.Second}

// FetchGroupieArtists récupère la liste des artistes depuis l'API Groupie (pour la carte / lieux).
func FetchGroupieArtists() ([]models.Artist, error) {
	resp, err := groupieClient.Get(groupieBase + "/artists")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("groupie artists: %d %s", resp.StatusCode, string(body))
	}
	var out []models.Artist
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

// FetchGroupieRelations récupère les relations (lieux ↔ dates) depuis l'API Groupie.
func FetchGroupieRelations() ([]models.Relation, error) {
	resp, err := groupieClient.Get(groupieBase + "/relation")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("groupie relation: %d %s", resp.StatusCode, string(body))
	}
	var out []models.Relation
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}
