# Groupie Tracker

Application web pour visualiser, filtrer et explorer des artistes via l’**API Spotify**.  
(Utilisation de l’API Spotify autorisée pour ce projet.)

## Objectif

Proposer une interface claire pour parcourir des artistes (liste, recherche, filtres, détail) avec une expérience utilisateur soignée (thème clair/sombre, responsive).

## Lancer le projet

### 1. Variables d’environnement Spotify

Créez une application sur [Spotify for Developers](https://developer.spotify.com/dashboard), récupérez le **Client ID** et le **Client Secret**, puis :

```bash
export SPOTIFY_CLIENT_ID="votre_client_id"
export SPOTIFY_CLIENT_SECRET="votre_client_secret"
```

### 2. Installation et exécution

```bash
go mod download
go run ./cmd/main.go
```

Le serveur écoute sur **http://localhost:8000**.

## Routes principales

| Route | Rôle |
|-------|------|
| `GET /` | Page d’accueil, lien vers la liste des artistes |
| `GET /artists` | Liste des artistes (avec recherche `?q=`, filtres) |
| `GET /artist/:id` | Détail d’un artiste (image, nom, genres, lien Spotify) |
| `GET /search?q=...` | Recherche par nom (redirige vers liste filtrée) |
| `GET /suggestions?q=...` | Suggestions JSON pour l’autocomplétion (JS) |
| `GET /map` | Carte / lieux (données limitées avec Spotify) |
| `GET /location/:lieu` | Concerts à un lieu (données limitées avec Spotify) |
| `GET /gims` | Redirection vers la fiche de l’artiste GIMS |

## Fonctionnalités

- **Obligatoires**  
  - Page d’accueil, liste d’artistes (image, nom, lien détail), page détail artiste  
  - Barre de recherche (requête HTTP) + suggestions en JS  
  - Filtres (année min/max, nombre de membres)  
  - Carte et page par lieu (structure en place)  
  - Clic sur un lieu → requête serveur → page concerts à ce lieu  
  - Gestion d’erreurs (404, 400, 500), route `/gims`  

- **Bonus**  
  - Thème sombre (toggle + préférence stockée)

## Limitations (API Spotify)

L’API Spotify ne fournit pas : lieux/dates de concerts, membres de groupe, année de création, premier album. Ces champs sont vides ou désactivés ; la carte et la page par lieu restent en place pour une éventuelle autre source de données.

## Lien GitHub

(À compléter si le dépôt est sur GitHub.)
