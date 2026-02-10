# Groupie Tracker

Application web moderne pour explorer et découvrir des artistes via l'API Spotify.

## 🚀 Démarrage rapide

### Prérequis
- Go 1.21+
- Compte Spotify for Developers ([créer une app](https://developer.spotify.com/dashboard))

### Installation

1. **Cloner le dépôt**
```bash
git clone https://github.com/Alexestgrand/groupie_tracker.git
cd groupie_tracker
```

2. **Configurer les credentials Spotify**
```bash
export SPOTIFY_CLIENT_ID="votre_client_id"
export SPOTIFY_CLIENT_SECRET="votre_client_secret"
```

3. **Lancer le serveur**
```bash
go run ./cmd/main.go
```

L'application est accessible sur **http://localhost:8000**

## 📋 Fonctionnalités

- **Liste d'artistes** : Grille de cartes avec images, noms, années de création
- **Recherche** : Recherche en temps réel avec suggestions automatiques
- **Filtres avancés** :
  - Date de création (min/max)
  - Date du premier album
  - Nombre de membres (solo, groupe)
  - Lieux (villes/pays populaires)
- **Page détail artiste** : 
  - Statistiques (popularité, followers, année de création)
  - Top titres avec aperçus
  - Albums avec pochette
  - Artistes similaires
- **Thème sombre** : Basculement automatique avec préférence sauvegardée

## 🛣️ Routes

| Route | Description |
|-------|-------------|
| `/` | Page d'accueil |
| `/artists` | Liste des artistes avec filtres |
| `/artist/{id}` | Détails d'un artiste |
| `/search?q=...` | Recherche d'artistes |
| `/suggestions?q=...` | API suggestions (JSON) |
| `/gims` | Redirection vers l'artiste GIMS |

## 🏗️ Structure

```
groupie_tracker/
├── cmd/main.go          # Point d'entrée, routes HTTP
├── api/spotify.go       # Client API Spotify
├── handlers/            # Gestionnaires HTTP
├── models/              # Structures de données
├── utils/               # Utilitaires (filtres, recherche)
├── templates/           # Templates HTML
└── static/              # CSS et JavaScript
```

## 🔧 Configuration

Les credentials Spotify sont requis via variables d'environnement :
- `SPOTIFY_CLIENT_ID`
- `SPOTIFY_CLIENT_SECRET`

Ou utilisez le script `start.sh` qui charge automatiquement un fichier `.env` s'il existe.

## 📝 Documentation

Pour une documentation complète du code, voir `CODE_DOCUMENTATION.md` (non versionné, généré localement).

## 📄 Licence

Projet académique - Groupie Tracker (25/26)
