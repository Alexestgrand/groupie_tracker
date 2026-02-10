# Groupie Tracker – Nouvelle Génération (25/26)

Application web permettant de visualiser, filtrer et explorer des artistes et leurs données.  
**Source de données : API Spotify** (aucune base de données ni fichier JSON local ; pas d’API Groupie).

---

## Contexte et objectif du projet

Le projet consiste à développer une application web complète pour explorer des données d’une API réelle centrée sur des artistes. L’application doit être robuste, claire et agréable à utiliser.

- **Logique côté serveur** : recherche, filtres, navigation et erreurs sont gérés en **Go**.
- **Frontend** : HTML et CSS (+ JS minimal pour les suggestions de recherche et le thème).
- **Objectifs pédagogiques** : manipulation de `net/http`, récupération et parsing JSON, organisation en packages (handlers, modèles, API), génération de pages avec `html/template`, gestion des erreurs, interface simple et lisible.

---

## API utilisée

L’application consomme **uniquement l’API Spotify** depuis le serveur Go :

- **Artistes** : liste et détails (nom, image, genres, popularité, followers, lien Spotify).
- **Limitation** : Spotify ne fournit pas les lieux/dates de concerts, ni année de création, membres ou premier album pour tous les artistes. Ces champs sont affichés lorsqu’ils sont disponibles, sinon « — ». La carte et les pages par lieu restent en place (événement interactif, structure) avec un message explicatif.

*Aucune base de données ni fichier JSON local n’est utilisé comme source principale.*

---

## Comment lancer le projet

### Prérequis

- **Go** 1.21+
- Un **compte Spotify for Developers** (gratuit) : [developer.spotify.com/dashboard](https://developer.spotify.com/dashboard)

### 1. Cloner le dépôt

```bash
git clone https://github.com/Alexestgrand/groupie_tracker.git
cd groupie_tracker
```

### 2. Variables d’environnement

Créez une application Spotify, récupérez **Client ID** et **Client Secret**, puis :

```bash
export SPOTIFY_CLIENT_ID="votre_client_id"
export SPOTIFY_CLIENT_SECRET="votre_client_secret"
```

Sous Windows (PowerShell) :

```powershell
$env:SPOTIFY_CLIENT_ID="votre_client_id"
$env:SPOTIFY_CLIENT_SECRET="votre_client_secret"
```

### 3. Lancer le serveur

À la **racine du projet** (dossier contenant `cmd/`, `templates/`, `static/`) :

```bash
go mod download
go run ./cmd/main.go
```

Le serveur écoute sur **http://localhost:8000**.

---

## Routes principales et leurs fonctions

| Route | Méthode | Rôle |
|-------|--------|------|
| `/` | GET | **Page d’accueil** : présentation de l’application, navigation vers la liste des artistes et la carte |
| `/artists` | GET | **Liste des artistes** : affichage en grille (image, nom, année de création, nombre de membres, lien détail) ; paramètres `?q=`, `?minYear=`, `?maxYear=`, `?memberCount=` pour recherche et filtres |
| `/artist/` | GET | **Détail d’un artiste** : image, nom, année de création, premier album, membres, genres, popularité, followers, lien Spotify, section concerts (message si non disponible) |
| `/search` | GET | **Recherche** : requête HTTP via `?q=...` ; recherche par nom d’artiste ; résultats avec mêmes filtres que la liste |
| `/suggestions` | GET | **Suggestions JSON** pour la barre de recherche (utilisé en JS pour l’autocomplétion) |
| `/map` | GET | **Carte interactive** : carte Leaflet pour visualiser les lieux de concerts ; message si données non fournies par l’API |
| `/location/` | GET | **Concerts à un lieu** : clic sur un lieu → requête serveur → page listant les concerts à ce lieu ; message si données non disponibles |
| **`/gims`** | GET | **Route nommée « gims »** : redirection vers la fiche de l’artiste GIMS (recherche par nom) |

---

## Fonctionnalités implémentées

### Obligatoires (sujet 25/26)

1. **Page d’accueil** — Présentation de l’application, navigation claire vers la liste des artistes et la carte.
2. **Liste des artistes** — Tous les artistes en cartes ; au minimum : image, nom, année de création, nombre de membres ; lien vers la page détaillée.
3. **Page de détails d’un artiste** — Image, nom, année de création, premier album, membres ; liste des concerts (dates + lieux) avec message si non fournie par l’API ; navigation (retour, autres pages).
4. **Barre de recherche** — Champ de recherche basé sur une requête HTTP ; recherche par nom d’artiste (et membre si disponible) ; **système de suggestion en JS** pour la barre de recherche.
5. **Filtres** — Filtre par intervalle (année de création min/max) ; filtre par sélection multiple (nombre de membres) ; combinaison possible des filtres. *(Filtre par lieux de concert : structure en place ; liste vide car Spotify ne fournit pas ces données.)*
6. **Carte interactive** — Page dédiée pour voir les lieux et dates de concert (carte Leaflet) ; message lorsque les données ne sont pas fournies par l’API.
7. **Événement interactif** — Clic sur un lieu (ou lien) déclenche une requête vers le serveur ; ex. clic sur un lieu → page listant les concerts à ce lieu (avec message si données non disponibles).
8. **Gestion d’erreurs** — Pages d’erreur personnalisées (404, erreurs de paramètres, etc.) ; pas de crash serveur ; erreurs gérées proprement en Go.

### Bonus

- **Thème sombre** : bouton de bascule, préférence stockée (`localStorage`), interface soignée.
- **UI/UX** : design cohérent, responsive, cartes et boutons avec retours visuels.

---

## Gestion du projet

- **Architecture** : packages `handlers`, `api`, `models`, `utils` ; entrée dans `cmd/main.go` ; templates dans `templates/`, static dans `static/`.
- **Tests réguliers** : lancement du serveur après modifications, vérification des routes et des pages dans le navigateur.
- **Interface** : conception soignée (palette, typographies, états au survol), navigation claire.

Détails dans **`ARCHITECTURE.md`**.

---

## Lien GitHub

Dépôt du projet :  
**https://github.com/Alexestgrand/groupie_tracker**

*(Projet versionné sur GitHub ; Gitea non utilisé.)*
