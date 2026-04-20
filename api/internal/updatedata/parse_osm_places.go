package updatedata

import (
	"context"
	"io"
	"math"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
)

// OSMPlace is a single place node extracted from an OpenStreetMap extract,
// filtered to those with a Chinese name tag. We keep only what we need for
// spatial matching against GeoNames.
type OSMPlace struct {
	Lat    float32
	Lon    float32
	NameZh string
}

// placeKindsForZhEnrichment are the OSM place tag values we consider for
// matching. Broader than strict city — many Chinese counties and townships
// are tagged town / suburb / hamlet.
var placeKindsForZhEnrichment = map[string]bool{
	"city":          true,
	"town":          true,
	"village":       true,
	"suburb":        true,
	"neighbourhood": true,
	"hamlet":        true,
}

// ParseOSMPlaces streams an OpenStreetMap .pbf file and returns nodes tagged
// as an inhabited place with a Chinese name. Ways and relations are skipped
// for simplicity — Chinese places in OSM reliably carry a representative
// node at the city center.
func ParseOSMPlaces(r io.Reader) ([]OSMPlace, error) {
	scanner := osmpbf.New(context.Background(), r, 4)
	defer scanner.Close()

	scanner.SkipWays = true
	scanner.SkipRelations = true

	var places []OSMPlace
	for scanner.Scan() {
		node, ok := scanner.Object().(*osm.Node)
		if !ok {
			continue
		}
		if !placeKindsForZhEnrichment[node.Tags.Find("place")] {
			continue
		}
		zh := node.Tags.Find("name:zh")
		if zh == "" {
			zh = node.Tags.Find("name:zh-Hans")
		}
		if zh == "" {
			continue
		}
		places = append(places, OSMPlace{
			Lat:    float32(node.Lat),
			Lon:    float32(node.Lon),
			NameZh: zh,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return places, nil
}

// NearestOSMPlaceWithin returns the closest place to (lat, lon) within
// maxKm kilometers, or ("", false) if none qualify. O(n) scan — the China
// place subset is small enough (~100k nodes) that this is fine for the
// ~thousands of GeoNames cities we'd enrich.
func NearestOSMPlaceWithin(places []OSMPlace, lat, lon float32, maxKm float64) (string, bool) {
	var bestName string
	bestDist := math.Inf(1)
	for _, p := range places {
		d := haversineKm(float64(lat), float64(lon), float64(p.Lat), float64(p.Lon))
		if d < bestDist {
			bestDist = d
			bestName = p.NameZh
		}
	}
	if bestDist <= maxKm {
		return bestName, true
	}
	return "", false
}

func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const r = 6371.0 // mean Earth radius in km
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	dφ := (lat2 - lat1) * math.Pi / 180
	dλ := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dφ/2)*math.Sin(dφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*math.Sin(dλ/2)*math.Sin(dλ/2)
	return 2 * r * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
