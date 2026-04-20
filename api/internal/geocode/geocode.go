package geocode

import "github.com/paulmach/orb/geojson"

// Geocoder resolves a (lat, lon) point into a full location record.
type Geocoder struct {
	Countries       *geojson.FeatureCollection
	CountriesByCode map[string]Country // optional; used as fallback when polygon lookup misses
	KDTree          *KDTree
	Admin1          map[string]AdminEntry
	Admin2          map[string]AdminEntry
}

// Result is the output of a reverse-geocode lookup.
type Result struct {
	Country    *Country // nil if point is not in any polygon AND no city-derived fallback
	Admin1     AdminEntry
	Admin2     AdminEntry
	CityName   string
	CityNameZh string
}

// Lookup performs country + admin + nearest-city lookup.
//
// Country resolution order:
//  1. Point-in-polygon against Natural Earth country data.
//  2. If that misses but the nearest city has a CountryCode present in
//     CountriesByCode, use that. Handles coastal / offshore grid-cell
//     centers that fall into ocean just outside a country's coastline.
func (g *Geocoder) Lookup(lat, lon float64) Result {
	var res Result
	if c, ok := LookupCountry(g.Countries, lat, lon); ok {
		res.Country = &c
	}
	if city, ok := g.KDTree.Nearest(lat, lon); ok {
		res.CityName = city.Name
		res.CityNameZh = city.NameZh
		// Only trust admin codes when the city is in the same country we detected.
		if res.Country == nil || res.Country.Code == city.CountryCode {
			res.Admin1, res.Admin2 = ResolveAdminNames(g.Admin1, g.Admin2, city.CountryCode, city.Admin1Code, city.Admin2Code)
		}
		// Fallback: polygon missed but we have a nearest city with a country.
		if res.Country == nil && city.CountryCode != "" && g.CountriesByCode != nil {
			if c, ok := g.CountriesByCode[city.CountryCode]; ok {
				res.Country = &c
			}
		}
	}
	return res
}
