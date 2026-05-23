package geocode

import (
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
)

// CNAdminResult is the admin hierarchy resolved from a point inside the
// Chinese administrative divisions dataset. For mainland queries this is
// usually fully populated (province + city + district). For HK / MO the
// city level is empty (the SAR is both province and city). For TW only
// province is available.
type CNAdminResult struct {
	ProvinceAD int
	Province   string // e.g. 浙江省
	CityAD     int
	City       string // e.g. 杭州市 (empty for HK/MO)
	DistrictAD int
	District   string // e.g. 西湖区
}

// CNAdminIndex is a spatial index of Chinese admin polygons. Point lookups
// walk four passes — districts first (most granular), then cities, then
// provinces, then patches. The patch tier carries hand-curated fallback
// polygons for disputed-territory gaps the official source leaves (e.g.
// eastern 藏南 between 墨脱县 and 察隅县), so it's checked last: real admin
// data wins everywhere it covers, the patch only fills the remaining holes.
type CNAdminIndex struct {
	// districts is a lightly-indexed slice: every entry carries its
	// bounding box so we can skip polygon tests for most features with
	// four float comparisons.
	districts []cnAdminFeature
	cities    []cnAdminFeature
	provinces []cnAdminFeature
	patches   []cnAdminFeature
	byADCode  map[int]cnAdminFeature
}

type cnAdminFeature struct {
	ADCode   int
	Name     string
	Level    string
	ParentAD int
	Bounds   orb.Bound
	Geometry orb.Geometry
}

// BuildCNAdminIndex ingests a CN-admin GeoJSON FeatureCollection (one feature
// per administrative polygon with adcode / name / level / parent properties)
// and returns a searchable index.
func BuildCNAdminIndex(fc *geojson.FeatureCollection) *CNAdminIndex {
	idx := &CNAdminIndex{
		byADCode: make(map[int]cnAdminFeature, len(fc.Features)),
	}
	for _, f := range fc.Features {
		ad := intProp(f.Properties, "adcode")
		if ad == 0 {
			continue
		}
		feat := cnAdminFeature{
			ADCode:   ad,
			Name:     stringProp(f.Properties, "name"),
			Level:    stringProp(f.Properties, "level"),
			ParentAD: intProp(f.Properties, "parent"),
			Bounds:   f.Geometry.Bound(),
			Geometry: f.Geometry,
		}
		idx.byADCode[ad] = feat
		switch feat.Level {
		case "district":
			idx.districts = append(idx.districts, feat)
		case "city":
			idx.cities = append(idx.cities, feat)
		case "province":
			idx.provinces = append(idx.provinces, feat)
		case "patch":
			idx.patches = append(idx.patches, feat)
		}
	}
	return idx
}

// cnAdminCountryCode maps a CN-admin province name to its ISO alpha-2
// country code. HK / MO / TW have their own codes; every other province is
// mainland China. The index by construction only covers Chinese territory,
// so the default of "CN" is safe even for province names not yet seen here.
func cnAdminCountryCode(province string) string {
	switch province {
	case "香港特别行政区":
		return "HK"
	case "澳门特别行政区":
		return "MO"
	case "台湾省":
		return "TW"
	default:
		return "CN"
	}
}

// Lookup finds the admin hierarchy for (lat, lon). Returns ok=false when
// the point is outside every known polygon (non-China or coastal gap).
func (idx *CNAdminIndex) Lookup(lat, lon float64) (CNAdminResult, bool) {
	if idx == nil {
		return CNAdminResult{}, false
	}
	pt := orb.Point{lon, lat}

	if f, ok := idx.pointInFeature(pt, idx.districts); ok {
		return idx.hierarchyFor(f), true
	}
	if f, ok := idx.pointInFeature(pt, idx.cities); ok {
		return idx.hierarchyFor(f), true
	}
	if f, ok := idx.pointInFeature(pt, idx.provinces); ok {
		return idx.hierarchyFor(f), true
	}
	if f, ok := idx.pointInFeature(pt, idx.patches); ok {
		return idx.hierarchyFor(f), true
	}
	return CNAdminResult{}, false
}

func (idx *CNAdminIndex) pointInFeature(pt orb.Point, feats []cnAdminFeature) (cnAdminFeature, bool) {
	for _, f := range feats {
		if !f.Bounds.Contains(pt) {
			continue
		}
		if polygonContainsPoint(f.Geometry, pt) {
			return f, true
		}
	}
	return cnAdminFeature{}, false
}

func polygonContainsPoint(g orb.Geometry, pt orb.Point) bool {
	switch geom := g.(type) {
	case orb.Polygon:
		return planar.PolygonContains(geom, pt)
	case orb.MultiPolygon:
		return planar.MultiPolygonContains(geom, pt)
	}
	return false
}

// hierarchyFor walks the ancestor chain from a matched feature (district,
// city, or province) up to the province level, filling in City/Province
// when they exist in the index.
func (idx *CNAdminIndex) hierarchyFor(f cnAdminFeature) CNAdminResult {
	var r CNAdminResult
	switch f.Level {
	case "district":
		r.DistrictAD = f.ADCode
		r.District = f.Name
		if parent, ok := idx.byADCode[f.ParentAD]; ok {
			switch parent.Level {
			case "city":
				r.CityAD = parent.ADCode
				r.City = parent.Name
				if grand, ok := idx.byADCode[parent.ParentAD]; ok && grand.Level == "province" {
					r.ProvinceAD = grand.ADCode
					r.Province = grand.Name
				}
			case "province":
				// HK/MO: district's direct parent is the province (SAR).
				r.ProvinceAD = parent.ADCode
				r.Province = parent.Name
			}
		}
	case "city":
		r.CityAD = f.ADCode
		r.City = f.Name
		if parent, ok := idx.byADCode[f.ParentAD]; ok && parent.Level == "province" {
			r.ProvinceAD = parent.ADCode
			r.Province = parent.Name
		}
	case "province":
		r.ProvinceAD = f.ADCode
		r.Province = f.Name
	case "patch":
		// Patch polygons fill gaps the official source leaves in disputed
		// territory. Surface the patch name as BOTH admin2 (district) and
		// city — the GeoNames nearest-city would otherwise be a settlement
		// in the de-facto-administering country (e.g. Pāsighāt in India for
		// 藏南), which jars next to a Chinese admin hierarchy. admin1
		// inherits from the parent province.
		r.DistrictAD = f.ADCode
		r.District = f.Name
		r.CityAD = f.ADCode
		r.City = f.Name
		if parent, ok := idx.byADCode[f.ParentAD]; ok && parent.Level == "province" {
			r.ProvinceAD = parent.ADCode
			r.Province = parent.Name
		}
	}
	return r
}

func intProp(p map[string]any, key string) int {
	v, ok := p[key]
	if !ok {
		return 0
	}
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	}
	return 0
}
