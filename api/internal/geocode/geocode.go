package geocode

import "github.com/paulmach/orb/geojson"

// Geocoder resolves a (lat, lon) point into a full location record.
type Geocoder struct {
	Countries       *geojson.FeatureCollection
	CountriesByCode map[string]Country // optional; used as fallback when polygon lookup misses
	CNAdmin         *CNAdminIndex      // optional; polygon-accurate admin hierarchy for CN family
	KDTree          *KDTree
	Admin1          map[string]AdminEntry
	Admin2          map[string]AdminEntry
}

// Result is the output of a reverse-geocode lookup.
type Result struct {
	Country      *Country // nil if point is not in any polygon AND no city-derived fallback
	Admin1       AdminEntry
	Admin2       AdminEntry
	CityName     string
	CityNameZh   string
	UsedCNAdmin  bool // true when admin1/admin2 came from the CN-admin polygon index
}

// Lookup performs country + admin + nearest-city lookup.
//
// Country resolution order:
//  1. Point-in-polygon against Natural Earth country data.
//  2. If that misses but the nearest city has a CountryCode present in
//     CountriesByCode, use that. Handles coastal / offshore grid-cell
//     centers that fall into ocean just outside a country's coastline.
//  3. If the CN-admin index hits, it overrides — that index is authoritative
//     for the China family (mainland + HK / MO / TW) and includes territory
//     Natural Earth assigns to neighbors (藏南 / 阿克赛钦 land in IN polygons
//     there). The province name decides the ISO code via cnAdminCountryCode.
//
// Admin / city for the CN family also come from the CN-admin index when
// available — it's authoritative for Chinese admin boundaries, unlike the
// nearest-city approach which mis-assigns grid cells that straddle district
// borders. Everywhere else uses GeoNames nearest-city admin codes.
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

	if g.CNAdmin != nil {
		if dv, ok := g.CNAdmin.Lookup(lat, lon); ok {
			// CN-admin hit ⇒ this point is in CN family per the authoritative
			// Chinese source. Override Natural Earth's country judgment,
			// which disagrees over disputed territory.
			if g.CountriesByCode != nil {
				if c, ok := g.CountriesByCode[cnAdminCountryCode(dv.Province)]; ok {
					res.Country = &c
				}
			}
			// Admin1 English prefers the curated provinceEnByZh table
			// (keeps "Inner Mongolia" / "Tibet" / SAR full names). Falls
			// back to pinyin for anything not in the table.
			a1En := ProvinceEn(dv.Province)
			if a1En == "" {
				a1En = chineseToEn(dv.Province)
			}
			res.Admin1 = AdminEntry{Zh: dv.Province, En: a1En}
			switch {
			case dv.District != "":
				res.Admin2 = AdminEntry{Zh: dv.District, En: chineseToEn(dv.District)}
			case dv.City != "":
				res.Admin2 = AdminEntry{Zh: dv.City, En: chineseToEn(dv.City)}
			default:
				res.Admin2 = AdminEntry{}
			}
			// "最近城市" on mainland CN follows the CN-admin city — GeoNames'
			// nearest point can sit in a different administrative city
			// (PM00ad's nearest GeoNames city is Fuyang, but the grid center
			// is in 杭州市 / 西湖区). HK / MO / TW have no separate city level
			// in CN-admin; keep the GeoNames nearest city there (愉景湾,
			// 大堂区, 台北市) rather than duplicating admin2 / admin1.
			if dv.City != "" {
				res.CityName = chineseToEn(dv.City)
				res.CityNameZh = dv.City
			}
			res.UsedCNAdmin = true
		}
	}
	return res
}
