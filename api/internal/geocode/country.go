package geocode

import (
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
)

// Country is the result of a point-in-polygon lookup.
type Country struct {
	Code string
	Name string
}

// LookupCountry returns the country whose polygon contains (lat, lon).
// The feature collection must have features with `iso_a2` and `name_en` properties.
func LookupCountry(fc *geojson.FeatureCollection, lat, lon float64) (Country, bool) {
	pt := orb.Point{lon, lat} // orb uses [lon, lat]
	for _, f := range fc.Features {
		if !contains(f.Geometry, pt) {
			continue
		}
		return Country{
			Code: stringProp(f.Properties, "iso_a2"),
			Name: stringProp(f.Properties, "name_en"),
		}, true
	}
	return Country{}, false
}

func contains(g orb.Geometry, pt orb.Point) bool {
	switch geom := g.(type) {
	case orb.Polygon:
		return planar.PolygonContains(geom, pt)
	case orb.MultiPolygon:
		return planar.MultiPolygonContains(geom, pt)
	}
	return false
}

func stringProp(p map[string]any, key string) string {
	if v, ok := p[key].(string); ok {
		return v
	}
	return ""
}
