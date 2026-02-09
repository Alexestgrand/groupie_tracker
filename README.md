# Groupie Tracker - Version Spotify

Application web permettant de visualiser, filtrer et explorer les données d'artistes depuis l'API Spotify.

## Configuration

### Credentials Spotify

Pour utiliser l'API Spotify, vous devez configurer vos credentials :

1. Créez une application sur [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)
2. Récupérez votre `Client ID` et `Client Secret`
3. Configurez les variables d'environnement :

```bash
export SPOTIFY_CLIENT_ID="votre_client_id"
export SPOTIFY_CLIENT_SECRET="votre_client_secret"
```

Ou modifiez directement le fichier `api/spotify.go` ligne 88-91 pour mettre vos credentials.

## Installation

```bash
go mod download
```

## Lancement

```bash
go run cmd/main.go
```

Le serveur sera accessible sur `http://localhost:8000`

## Fonctionnalités

- ✅ Liste des artistes populaires depuis Spotify
- ✅ Recherche d'artistes
- ✅ Suggestions de recherche en temps réel
- ✅ Page de détails d'artiste avec genres, popularité, followers
- ✅ Filtres par genres (via recherche)
- ✅ Design moderne avec palette marron/beige

## Limitations

L'API Spotify ne fournit pas :
- ❌ Lieux de concerts
- ❌ Dates de concerts
- ❌ Membres des groupes
- ❌ Année de création
- ❌ Premier album

Ces fonctionnalités sont désactivées ou affichent un message d'information.

## Structure du projet

```
groupie-tracker-ng/
├── api/
│   └── spotify.go          # Client API Spotify
├── handlers/               # Gestionnaires HTTP
├── models/                 # Modèles de données
├── templates/              # Templates HTML
├── static/                 # Fichiers statiques (CSS, JS)
├── utils/                  # Utilitaires
└── cmd/
    └── main.go            # Point d'entrée
```

## Technologies

- Go (net/http, html/template)
- HTML/CSS/JavaScript (minimal, uniquement pour suggestions)
- API Spotify


