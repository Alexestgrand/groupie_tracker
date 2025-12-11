
## Architecture du projet

```
groupie-tracker-ng/
├── cmd/
│   └── main.go                 # Point d'entrée de l'application
├── api/
│   └── client.go               # Client HTTP pour consommer l'API externe
├── models/
│   └── artist.go               # Structures de données (Artist, Location, Date, Relation)
├── handlers/
│   ├── common.go               # Variables partagées (apiClient, cacheInstance)
│   ├── home.go                 # Handler page d'accueil
│   ├── artists.go              # Handler liste et détails des artistes
│   ├── search.go               # Handler recherche et suggestions
│   ├── location.go             # Handler concerts par lieu
│   ├── map.go                  # Handler carte interactive
│   ├── gims.go                 # Handler route spéciale /gims
│   └── template.go             # Fonction de rendu des templates
├── utils/
│   ├── errors.go               # Gestion centralisée des erreurs
│   ├── search.go               # Fonctions de recherche
│   ├── filter.go               # Fonctions de filtrage
│   └── years.go                # Utilitaires pour les années
├── cache/
│   └── cache.go                # Système de cache en mémoire
├── templates/
│   ├── layout.html             # Template de base avec navigation
│   ├── home.html               # Page d'accueil
│   ├── artists.html            # Liste des artistes
│   ├── artists_details.html     # Détails d'un artiste
│   └── error.html              # Pages d'erreur
├── static/
│   ├── css/
│   │   └── style.css           # Styles CSS
│   └── js/
│       └── script.js           # Scripts JavaScript
├── go.mod                      # Dépendances Go
└── README.md                   # Documentation du projet
```

---

## Structure des packages

### 1. `cmd/main.go` - Point d'entrée

**Responsabilité** : Configuration du serveur HTTP et des routes

**À implémenter** :
- Création du routeur HTTP (`http.NewServeMux()`)
- Configuration des routes statiques (`/static/`)
- Enregistrement des handlers pour chaque route
- Démarrage du serveur sur le port 8080

**Routes à configurer** :
```go
/                    → handlers.HomeHandler
/artists             → handlers.ArtistsHandler
/artist/{id}         → handlers.ArtistDetailHandler
/search               → handlers.SearchHandler
/suggestions          → handlers.SuggestionsHandler
/map                  → handlers.MapHandler
/location/{location}  → handlers.LocationHandler
/gims                 → handlers.GimsHandler
```

### 2. `api/client.go` - Client API

**Responsabilité** : Communication avec l'API externe Groupie Trackers

**À implémenter** :
- Structure `Client` avec `http.Client` et timeout
- Fonction `NewClient()` pour créer une instance
- Méthodes pour chaque endpoint :
  - `FetchArtists()` → `/api/artists`
  - `FetchLocations()` → `/api/locations`
  - `FetchDates()` → `/api/dates`
  - `FetchRelations()` → `/api/relation`
  - `FetchArtistDetail(id)` → combine toutes les données pour un artiste

**Points importants** :
- Gestion des erreurs HTTP (codes de statut)
- Parsing JSON avec `encoding/json`
- Timeout de 10 secondes pour éviter les blocages

### 3. `models/artist.go` - Modèles de données

**Responsabilité** : Définir les structures de données

**Structures à créer** :
```go
type Artist struct {
    ID           int
    Image        string
    Name         string
    Members      []string
    CreationDate int
    FirstAlbum   string
    Locations    string  // URL de l'endpoint
    ConcertDates string  // URL de l'endpoint
    Relations    string  // URL de l'endpoint
}

type Location struct {
    ID        int
    Locations []string
    Dates     string
}

type Date struct {
    ID    int
    Dates []string
}

type Relation struct {
    ID             int
    DatesLocations map[string][]string
}

type ArtistDetail struct {
    Artist
    ConcertDates []string
    Locations    []string
    Relations    map[string][]string
    BirthDates   map[string]string  // Bonus
    DeathDates   map[string]string  // Bonus
}

type FilterOptions struct {
    MinYear       int
    MaxYear       int
    MemberCount   []int
    Locations     []string
    FirstAlbumMin string
    FirstAlbumMax string
}
```

### 4. `handlers/` - Gestionnaires HTTP

**Responsabilité** : Traiter les requêtes HTTP et rendre les templates

#### `handlers/common.go`
```go
var (
    apiClient     = api.NewClient()
    cacheInstance = cache.GetInstance()
)
```

#### `handlers/home.go`
- Afficher la page d'accueil
- Utiliser `renderTemplate(w, "home.html", data)`

#### `handlers/artists.go`
- `ArtistsHandler` : Liste des artistes avec filtres et recherche
- `ArtistDetailHandler` : Détails d'un artiste par ID
- Récupérer les données depuis le cache ou l'API
- Appliquer les filtres et la recherche
- Rendre le template approprié

#### `handlers/search.go`
- `SearchHandler` : Résultats de recherche
- `SuggestionsHandler` : Retourne JSON pour les suggestions en temps réel

#### `handlers/location.go`
- `LocationHandler` : Liste des concerts à un lieu spécifique
- Extraire le lieu de l'URL
- Filtrer les relations par lieu

#### `handlers/map.go`
- `MapHandler` : Préparer les données pour la carte interactive
- Agréger les lieux et dates de concerts

#### `handlers/gims.go`
- `GimsHandler` : Route spéciale `/gims`
- Rechercher l'artiste GIMS et rediriger vers sa page

#### `handlers/template.go`
```go
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
    templates := template.Must(template.ParseFiles(
        "templates/layout.html",
        "templates/"+tmpl,
    ))
    templates.ExecuteTemplate(w, "layout", data)
}
```

### 5. `utils/` - Utilitaires

#### `utils/errors.go`
- `RenderError(w, statusCode, message)` : Afficher une page d'erreur
- `HandleError(w, err, statusCode)` : Gérer les erreurs de manière centralisée

#### `utils/search.go`
- `SearchArtists(artists, query)` : Rechercher dans les artistes
- `GetSuggestions(artists, query)` : Générer des suggestions pour la recherche

#### `utils/filter.go`
- `FilterArtists(artists, options)` : Filtrer les artistes selon les critères
- `ParseFilterOptions(queryParams)` : Parser les paramètres de requête

#### `utils/years.go`
- `GetAllYears()` : Générer une liste d'années pour les filtres

### 6. `cache/cache.go` - Système de cache

**Responsabilité** : Mettre en cache les données API pour améliorer les performances

**À implémenter** :
- Structure `Cache` avec mutex pour thread-safety
- Expiration automatique (5 minutes)
- Méthodes :
  - `GetArtists()` / `SetArtists()`
  - `GetLocations()` / `SetLocations()`
  - `GetDates()` / `SetDates()`
  - `GetRelations()` / `SetRelations()`
  - `GetArtistDetail(id)` / `SetArtistDetail(id, detail)`

### 7. `templates/` - Templates HTML

#### `templates/layout.html`
- Structure HTML de base
- Navigation
- Inclusion CSS et JS
- Bloc `{{block "content"}}` pour les pages enfants

#### `templates/home.html`
- Page d'accueil avec présentation
- Liens vers les différentes sections

#### `templates/artists.html`
- Barre de recherche
- Filtres (année, membres, album)
- Grille de cartes d'artistes

#### `templates/artists_details.html`
- Image et informations de l'artiste
- Liste des membres
- Liste des concerts avec dates et lieux

#### `templates/error.html`
- Page d'erreur personnalisée (404, 400, 500)

### 8. `static/` - Fichiers statiques

#### `static/css/style.css`
- Styles pour toute l'application
- Design responsive
- Variables CSS pour les couleurs

#### `static/js/script.js`
- Suggestions de recherche en temps réel
- Interactions utilisateur (optionnel)

---

## Étapes de réalisation

### Phase 1 : Configuration de base (Priorité 1)

1. **`cmd/main.go`**
   - Créer le serveur HTTP
   - Configurer les routes statiques
   - Enregistrer les routes principales (même si les handlers sont vides)

2. **`templates/layout.html`**
   - Créer la structure HTML de base
   - Ajouter la navigation
   - Tester que le serveur démarre

### Phase 2 : Modèles et API (Priorité 1)

3. **`models/artist.go`**
   - Définir toutes les structures de données
   - Ajouter les tags JSON appropriés

4. **`api/client.go`**
   - Implémenter `NewClient()`
   - Implémenter `FetchArtists()`
   - Tester avec un appel API simple
   - Implémenter les autres méthodes (`FetchLocations`, `FetchDates`, `FetchRelations`)
   - Implémenter `FetchArtistDetail()` qui combine toutes les données

### Phase 3 : Handlers de base (Priorité 1)

5. **`handlers/common.go`**
   - Créer les variables partagées

6. **`handlers/template.go`**
   - Implémenter `renderTemplate()`

7. **`handlers/home.go`**
   - Implémenter `HomeHandler`
   - Créer `templates/home.html` avec du contenu simple

8. **`handlers/artists.go`**
   - Implémenter `ArtistsHandler` (sans filtres pour l'instant)
   - Récupérer les artistes depuis l'API
   - Créer `templates/artists.html` avec une liste simple
   - Implémenter `ArtistDetailHandler`
   - Créer `templates/artists_details.html`

### Phase 4 : Recherche et filtres (Priorité 1)

9. **`utils/search.go`**
   - Implémenter `SearchArtists()`
   - Implémenter `GetSuggestions()`

10. **`handlers/search.go`**
    - Implémenter `SearchHandler`
    - Implémenter `SuggestionsHandler` (retourne JSON)

11. **`utils/filter.go`**
    - Implémenter `ParseFilterOptions()`
    - Implémenter `FilterArtists()`

12. **`handlers/artists.go`** (modification)
    - Ajouter la gestion des filtres dans `ArtistsHandler`
    - Ajouter la gestion de la recherche

            // Baptise ----------------------------------------------------------------------------
13. **`templates/artists.html`** (modification) 
    - Ajouter le formulaire de recherche
    - Ajouter les formulaires de filtres
    - Intégrer les suggestions JavaScript

14. **`static/js/script.js`**
    - Implémenter les suggestions en temps réel

### Phase 5 : Fonctionnalités avancées (Priorité 1)

15. **`handlers/location.go`**
    - Implémenter `LocationHandler`
    - Créer `templates/location.html` (ou réutiliser un template existant)

16. **`handlers/map.go`**
    - Implémenter `MapHandler`
    - Préparer les données pour la carte
    - Créer `templates/map.html`

17. **`handlers/gims.go`**
    - Implémenter `GimsHandler`

### Phase 6 : Gestion d'erreurs (Priorité 1)

18. **`utils/errors.go`**
    - Implémenter `RenderError()`
    - Implémenter `HandleError()`

19. **`templates/error.html`**
    - Créer une page d'erreur personnalisée

20. **Modifier tous les handlers**
    - Ajouter la gestion d'erreurs appropriée
    - Gérer les cas 404, 400, 500

### Phase 7 : Cache (Bonus - Priorité 2)

21. **`cache/cache.go`**
    - Implémenter la structure `Cache`
    - Implémenter toutes les méthodes Get/Set
    - Ajouter l'expiration automatique

22. **Modifier les handlers**
    - Utiliser le cache avant d'appeler l'API
    - Mettre à jour le cache après les appels API

### Phase 8 : Styling et UX (Priorité 1)

23. **`static/css/style.css`**
    - Créer un design moderne et responsive
    - Styliser tous les composants
    - Ajouter des animations (optionnel)

24. **Améliorer les templates**
    - Ajouter des classes CSS appropriées
    - Améliorer l'UX générale

### Phase 9 : Tests et finitions (Priorité 1)

25. **Tester toutes les fonctionnalités**
    - Tester chaque route
    - Tester les filtres et la recherche
    - Tester la gestion d'erreurs
    - Tester sur différents navigateurs

26. **Optimisations**
    - Vérifier les performances
    - Optimiser les requêtes API
    - Améliorer le cache si nécessaire

27. **Documentation**
    - Compléter le README.md
    - Ajouter des commentaires dans le code

---

## Flux de données

### Requête utilisateur → Réponse

```
1. Utilisateur fait une requête HTTP
   ↓
2. main.go route la requête vers le handler approprié
   ↓
3. Handler vérifie le cache
   ├─ Si données en cache → utilise le cache
   └─ Sinon → appelle api/client.go
   ↓
4. api/client.go fait un appel HTTP à l'API externe
   ↓
5. Les données JSON sont parsées en structures Go
   ↓
6. Les données sont mises en cache
   ↓
7. Handler applique filtres/recherche (si nécessaire)
   ↓
8. Handler prépare les données pour le template
   ↓
9. handler/template.go rend le template HTML
   ↓
10. Réponse HTML envoyée à l'utilisateur
```

### Exemple : Affichage de la liste des artistes

```
GET /artists
  ↓
handlers.ArtistsHandler
  ↓
cache.GetArtists() → vide
  ↓
api.FetchArtists()
  ↓
HTTP GET https://groupietrackers.herokuapp.com/api/artists
  ↓
Parse JSON → []models.Artist
  ↓
cache.SetArtists(artists)
  ↓
utils.FilterArtists(artists, options)
  ↓
utils.SearchArtists(filteredArtists, query)
  ↓
renderTemplate("artists.html", data)
  ↓
HTML avec liste des artistes
```

---

## Bonnes pratiques

### 1. Gestion d'erreurs

- **Toujours** vérifier les erreurs retournées par les fonctions
- Utiliser `utils.HandleError()` pour une gestion centralisée
- Ne jamais faire de `panic()` en production
- Logger les erreurs avec `log.Printf()`

### 2. Code Go

- Suivre les conventions Go (noms de fonctions, variables)
- Commenter les fonctions publiques
- Séparer les responsabilités (un handler = une responsabilité)
- Éviter les imports inutilisés

### 3. Templates HTML

- Utiliser le template de base (`layout.html`) pour éviter la duplication
- Utiliser les blocs Go templates (`{{block}}`, `{{define}}`)
- Échapper les données utilisateur avec `{{.}}` (automatique en Go)

### 4. Performance

- Utiliser le cache pour réduire les appels API
- Limiter le timeout des requêtes HTTP (10 secondes)
- Éviter les boucles imbriquées inutiles

### 5. Sécurité

- Valider et nettoyer les entrées utilisateur
- Échapper les données dans les templates (automatique)
- Gérer les erreurs sans exposer d'informations sensibles

### 6. Tests

- Tester chaque handler individuellement
- Tester les cas d'erreur (404, 500, etc.)
- Tester les filtres et la recherche avec différents paramètres

---

## Points d'attention

### ⚠️ Erreurs courantes à éviter

1. **Oublier de gérer les erreurs** : Toujours vérifier `err != nil`
2. **Appels API sans timeout** : Risque de blocage indéfini
3. **Parsing JSON incorrect** : Vérifier que les tags JSON correspondent à l'API
4. **Routes mal configurées** : L'ordre des routes dans `main.go` est important
5. **Cache non thread-safe** : Utiliser des mutex pour protéger le cache
6. **Templates non trouvés** : Vérifier les chemins des fichiers templates

### 💡 Conseils

- **Commencer simple** : Implémenter d'abord les fonctionnalités de base, puis ajouter les filtres et la recherche
- **Tester régulièrement** : Tester après chaque fonctionnalité ajoutée
- **Utiliser le cache** : C'est un bonus mais ça améliore grandement les performances
- **Documenter au fur et à mesure** : Ajouter des commentaires pendant le développement

---

## Ressources utiles

- [Documentation Go net/http](https://pkg.go.dev/net/http)
- [Documentation Go html/template](https://pkg.go.dev/html/template)
- [API Groupie Trackers](https://groupietrackers.herokuapp.com/api)
- [Go by Example](https://gobyexample.com/)

---

**Bon développement ! 🚀**



