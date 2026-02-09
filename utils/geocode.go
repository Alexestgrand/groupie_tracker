package utils

import (
	"strings"
)

// Coords représente des coordonnées géographiques.
type Coords struct {
	Lat float64
	Lng float64
}

// Lieux courants de l'API Groupie (format "ville-pays") -> coordonnées approximatives.
var cityCoords = map[string]Coords{
	"london-uk":           {51.5074, -0.1278},
	"manchester-uk":       {53.4808, -2.2426},
	"birmingham-uk":       {52.4862, -1.8904},
	"glasgow-uk":          {55.8642, -4.2518},
	"liverpool-uk":        {53.4084, -2.9916},
	"paris-france":        {48.8566, 2.3522},
	"lyon-france":         {45.7640, 4.8357},
	"marseille-france":    {43.2965, 5.3698},
	"berlin-germany":      {52.5200, 13.4050},
	"munich-germany":      {48.1351, 11.5820},
	"hamburg-germany":     {53.5511, 9.9937},
	"amsterdam-netherlands": {52.3676, 4.9041},
	"rotterdam-netherlands": {51.9225, 4.4792},
	"madrid-spain":        {40.4168, -3.7038},
	"barcelona-spain":     {41.3851, 2.1734},
	"rome-italy":         {41.9028, 12.4964},
	"milan-italy":        {45.4642, 9.1900},
	"dublin-ireland":      {53.3498, -6.2603},
	"brussels-belgium":   {50.8503, 4.3517},
	"vienna-austria":      {48.2082, 16.3738},
	"zurich-switzerland":  {47.3769, 8.5417},
	"warsaw-poland":       {52.2297, 21.0122},
	"prague-czech_republic": {50.0755, 14.4378},
	"copenhagen-denmark":  {55.6761, 12.5683},
	"stockholm-sweden":    {59.3293, 18.0686},
	"oslo-norway":         {59.9139, 10.7522},
	"helsinki-finland":    {60.1699, 24.9384},
	"moscow-russia":       {55.7558, 37.6173},
	"st_petersburg-russia": {59.9343, 30.3351},
	"new-york-usa":        {40.7128, -74.0060},
	"los-angeles-usa":     {34.0522, -118.2437},
	"chicago-usa":         {41.8781, -87.6298},
	"san-francisco-usa":   {37.7749, -122.4194},
	"seattle-usa":         {47.6062, -122.3321},
	"boston-usa":          {42.3601, -71.0589},
	"austin-usa":         {30.2672, -97.7431},
	"denver-usa":          {39.7392, -104.9903},
	"toronto-canada":      {43.6532, -79.3832},
	"montreal-canada":     {45.5017, -73.5673},
	"vancouver-canada":    {49.2827, -123.1207},
	"mexico_city-mexico":  {19.4326, -99.1332},
	"sao_paulo-brazil":   {-23.5505, -46.6333},
	"buenos_aires-argentina": {-34.6037, -58.3816},
	"tokyo-japan":         {35.6762, 139.6503},
	"osaka-japan":         {34.6937, 135.5023},
	"seoul-south_korea":   {37.5665, 126.9780},
	"singapore-singapore": {1.3521, 103.8198},
	"hong_kong-hong_kong": {22.3193, 114.1694},
	"sydney-australia":   {-33.8688, 151.2093},
	"melbourne-australia": {-37.8136, 144.9631},
	"auckland-new_zealand": {-36.8509, 174.7645},
	"tel_aviv-israel":     {32.0853, 34.7818},
	"istanbul-turkey":     {41.0082, 28.9784},
	"athens-greece":       {37.9838, 23.7275},
	"lisbon-portugal":     {38.7223, -9.1393},
	"budapest-hungary":    {47.4979, 19.0402},
	"bucharest-romania":   {44.4268, 26.1025},
	"belgrade-serbia":     {44.7866, 20.4489},
	"philadelphia-usa":    {39.9526, -75.1652},
	"detroit-usa":         {42.3314, -83.0458},
	"nashville-usa":       {36.1627, -86.7816},
	"atlanta-usa":         {33.7490, -84.3880},
	"miami-usa":           {25.7617, -80.1918},
	"portland-usa":        {45.5152, -122.6784},
	"san-diego-usa":       {32.7157, -117.1611},
	"dallas-usa":          {32.7767, -96.7970},
	"houston-usa":         {29.7604, -95.3698},
	"phoenix-usa":         {33.4484, -112.0740},
	"minneapolis-usa":     {44.9778, -93.2650},
	"cleveland-usa":       {41.4993, -81.6944},
	"pittsburgh-usa":     {40.4406, -79.9959},
	"washington-usa":      {38.9072, -77.0369},
	"new_orleans-usa":     {29.9511, -90.0715},
	"calgary-canada":      {51.0447, -114.0719},
	"quito-ecuador":       {-0.1807, -78.4678},
	"bogota-colombia":     {4.7110, -74.0721},
	"lima-peru":           {-12.0464, -77.0428},
	"santiago-chile":      {-33.4489, -70.6693},
	"bogotá-colombia":     {4.7110, -74.0721},
}

// GetCoords retourne les coordonnées pour un lieu (ex: "london-uk"). Insensible à la casse.
func GetCoords(location string) (lat, lng float64, ok bool) {
	key := strings.ToLower(strings.ReplaceAll(location, " ", "_"))
	if c, ok := cityCoords[key]; ok {
		return c.Lat, c.Lng, true
	}
	// Inconnu : position par défaut pour afficher quand même le lieu
	return 20, 0, true
}
