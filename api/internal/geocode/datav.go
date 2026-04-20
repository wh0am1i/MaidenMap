package geocode

import (
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
)

// DataVResult is the admin hierarchy resolved from a point inside the
// Chinese Administrative Divisions (DataV) dataset. For mainland queries
// this is usually fully populated (province + city + district). For HK/MO
// the city level is empty since DataV skips it. For TW only province is
// available.
type DataVResult struct {
	ProvinceAD int
	Province   string // e.g. 浙江省
	CityAD     int
	City       string // e.g. 杭州市 (empty for HK/MO)
	DistrictAD int
	District   string // e.g. 西湖区
}

// DataVIndex is a spatial index of Chinese admin polygons. Point lookups go
// through two passes — first districts (most granular), then provinces as
// a safety net for spots districts don't cover (rare, e.g. disputed areas
// at province boundaries).
type DataVIndex struct {
	// districts is a lightly-indexed slice: every entry carries its
	// bounding box so we can skip polygon tests for most features with
	// four float comparisons.
	districts []dataVFeature
	provinces []dataVFeature
	byADCode  map[int]dataVFeature
}

type dataVFeature struct {
	ADCode   int
	Name     string
	Level    string
	ParentAD int
	Bounds   orb.Bound
	Geometry orb.Geometry
}

// BuildDataVIndex ingests a DataV-style GeoJSON FeatureCollection (produced
// by updatedata.EncodeDataVNodes) and returns a searchable index.
func BuildDataVIndex(fc *geojson.FeatureCollection) *DataVIndex {
	idx := &DataVIndex{
		byADCode: make(map[int]dataVFeature, len(fc.Features)),
	}
	for _, f := range fc.Features {
		ad := intProp(f.Properties, "adcode")
		if ad == 0 {
			continue
		}
		feat := dataVFeature{
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
		case "province":
			idx.provinces = append(idx.provinces, feat)
		}
	}
	return idx
}

// Lookup finds the admin hierarchy for (lat, lon). Returns ok=false when
// the point is outside every known polygon (non-China or coastal gap).
func (idx *DataVIndex) Lookup(lat, lon float64) (DataVResult, bool) {
	if idx == nil {
		return DataVResult{}, false
	}
	pt := orb.Point{lon, lat}

	if f, ok := idx.pointInFeature(pt, idx.districts); ok {
		return idx.hierarchyFor(f), true
	}
	if f, ok := idx.pointInFeature(pt, idx.provinces); ok {
		return idx.hierarchyFor(f), true
	}
	return DataVResult{}, false
}

func (idx *DataVIndex) pointInFeature(pt orb.Point, feats []dataVFeature) (dataVFeature, bool) {
	for _, f := range feats {
		if !f.Bounds.Contains(pt) {
			continue
		}
		if polygonContainsPoint(f.Geometry, pt) {
			return f, true
		}
	}
	return dataVFeature{}, false
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

// hierarchyFor walks the ancestor chain from a matched feature (district or
// province) up to the province level, filling in City/Province when they
// exist in the index.
func (idx *DataVIndex) hierarchyFor(f dataVFeature) DataVResult {
	var r DataVResult
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
	case "province":
		r.ProvinceAD = f.ADCode
		r.Province = f.Name
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
