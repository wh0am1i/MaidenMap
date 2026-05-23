package geocode

import (
	"testing"

	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeocodeLandPoint(t *testing.T) {
	fc := testCountries(t)
	cities := []City{
		{Name: "TestCapital", NameZh: "测试首府", Lat: 5, Lon: 5, CountryCode: "XX", Admin1Code: "AA", Admin2Code: "001"},
	}
	tree := BuildKDTree(cities)
	a1 := map[string]AdminEntry{"XX.AA": {En: "Test Admin1", Zh: "测试一级"}}
	a2 := map[string]AdminEntry{"XX.AA.001": {En: "Test Admin2", Zh: "测试二级"}}

	g := &Geocoder{Countries: fc, KDTree: tree, Admin1: a1, Admin2: a2}
	res := g.Lookup(5, 5)

	require.NotNil(t, res.Country)
	assert.Equal(t, "XX", res.Country.Code)
	assert.Equal(t, "Testland", res.Country.Name)
	assert.Equal(t, "测试国", res.Country.NameZh)
	assert.Equal(t, "Test Admin1", res.Admin1.En)
	assert.Equal(t, "测试一级", res.Admin1.Zh)
	assert.Equal(t, "Test Admin2", res.Admin2.En)
	assert.Equal(t, "测试二级", res.Admin2.Zh)
	assert.Equal(t, "TestCapital", res.CityName)
	assert.Equal(t, "测试首府", res.CityNameZh)
}

func TestGeocodeOcean(t *testing.T) {
	fc := testCountries(t)
	g := &Geocoder{Countries: fc, KDTree: BuildKDTree(nil)}
	res := g.Lookup(-50, -50)
	assert.Nil(t, res.Country)
	assert.Equal(t, AdminEntry{}, res.Admin1)
	assert.Equal(t, AdminEntry{}, res.Admin2)
	assert.Empty(t, res.CityName)
	assert.Empty(t, res.CityNameZh)
}

func TestGeocodeNoCities(t *testing.T) {
	fc := testCountries(t)
	g := &Geocoder{Countries: fc, KDTree: BuildKDTree(nil)}
	res := g.Lookup(5, 5)
	require.NotNil(t, res.Country)
	assert.Equal(t, "XX", res.Country.Code)
	assert.Empty(t, res.CityName)
}

// For CN-family queries, the CN-admin index should override GeoNames for admin1, admin2
// AND city — otherwise the "最近城市" field leaks wrong data (e.g. the grid
// center sits in 西湖区/杭州市 but the nearest GeoNames city15000 is Fuyang).
func TestGeocodeCNFamilyUsesCNAdminForCity(t *testing.T) {
	cnCountries := []byte(`{
        "type":"FeatureCollection",
        "features":[
            {"type":"Feature","properties":{"iso_a2":"CN","name_en":"China","name_zh":"中国"},
             "geometry":{"type":"Polygon","coordinates":[[[100,20],[140,20],[140,50],[100,50],[100,20]]]}}
        ]}`)
	fc, err := geojson.UnmarshalFeatureCollection(cnCountries)
	require.NoError(t, err)

	// GeoNames nearest city to (30.15, 120.04) is Fuyang — which is what we
	// must NOT end up emitting for CN.
	cities := []City{{Name: "Fuyang", NameZh: "", Lat: 30.05, Lon: 119.96, CountryCode: "CN", Admin1Code: "02", Admin2Code: "0203"}}

	// CNAdmin: three-level hierarchy — 浙江省 → 杭州市 → 西湖区 — covering the query point.
	dvRaw := []byte(`{
        "type":"FeatureCollection",
        "features":[
            {"type":"Feature",
             "properties":{"adcode":330000,"name":"浙江省","level":"province","parent":100000},
             "geometry":{"type":"Polygon","coordinates":[[[118,28],[123,28],[123,32],[118,32],[118,28]]]}},
            {"type":"Feature",
             "properties":{"adcode":330100,"name":"杭州市","level":"city","parent":330000},
             "geometry":{"type":"Polygon","coordinates":[[[119.5,29.5],[120.8,29.5],[120.8,30.6],[119.5,30.6],[119.5,29.5]]]}},
            {"type":"Feature",
             "properties":{"adcode":330106,"name":"西湖区","level":"district","parent":330100},
             "geometry":{"type":"Polygon","coordinates":[[[119.9,30.0],[120.2,30.0],[120.2,30.3],[119.9,30.3],[119.9,30.0]]]}}
        ]}`)
	dvFC, err := geojson.UnmarshalFeatureCollection(dvRaw)
	require.NoError(t, err)

	g := &Geocoder{
		Countries:       fc,
		CountriesByCode: map[string]Country{"CN": {Code: "CN", Name: "China", NameZh: "中国"}},
		CNAdmin:           BuildCNAdminIndex(dvFC),
		KDTree:          BuildKDTree(cities),
		Admin1:          map[string]AdminEntry{"CN.02": {En: "Zhejiang", Zh: ""}},
		Admin2:          map[string]AdminEntry{"CN.02.0203": {En: "Fuyang", Zh: ""}},
	}

	res := g.Lookup(30.15, 120.04)
	require.NotNil(t, res.Country)
	assert.Equal(t, "CN", res.Country.Code)
	assert.True(t, res.UsedCNAdmin, "CN-admin should have hit for this CN point")
	assert.Equal(t, "浙江省", res.Admin1.Zh)
	assert.Equal(t, "Zhejiang", res.Admin1.En)
	assert.Equal(t, "西湖区", res.Admin2.Zh)
	// City must follow CN-admin — not GeoNames' Fuyang. English is the pinyin
	// romanization of the Chinese city (the standard international form).
	assert.Equal(t, "杭州市", res.CityNameZh)
	assert.Equal(t, "Hangzhou", res.CityName)
	// Same for admin2 — pinyin-derived English with "District" suffix.
	assert.Equal(t, "Xihu District", res.Admin2.En)
}

// disputedFixture sets up a Natural-Earth fixture that asserts India over the
// box [70,25]–[100,40] (covers 藏南 and 阿克赛钦), plus a CN-admin fixture that
// asserts China for the same area (西藏/错那 around NL68kd, 新疆/和田县 around
// 阿克赛钦). Used by the regression tests below to verify CN-admin overrides the
// Natural Earth country judgment for disputed territory.
func disputedFixture(t *testing.T) (*geojson.FeatureCollection, *CNAdminIndex) {
	t.Helper()
	ne := []byte(`{
        "type":"FeatureCollection",
        "features":[
            {"type":"Feature","properties":{"iso_a2":"IN","name_en":"India","name_zh":"印度"},
             "geometry":{"type":"Polygon","coordinates":[[[70,25],[100,25],[100,40],[70,40],[70,25]]]}}
        ]}`)
	fc, err := geojson.UnmarshalFeatureCollection(ne)
	require.NoError(t, err)

	dv := []byte(`{
        "type":"FeatureCollection",
        "features":[
            {"type":"Feature",
             "properties":{"adcode":540000,"name":"西藏自治区","level":"province","parent":100000},
             "geometry":{"type":"Polygon","coordinates":[[[78,27],[96,27],[96,36],[78,36],[78,27]]]}},
            {"type":"Feature",
             "properties":{"adcode":542300,"name":"山南市","level":"city","parent":540000},
             "geometry":{"type":"Polygon","coordinates":[[[91,27.5],[94,27.5],[94,29],[91,29],[91,27.5]]]}},
            {"type":"Feature",
             "properties":{"adcode":542329,"name":"错那县","level":"district","parent":542300},
             "geometry":{"type":"Polygon","coordinates":[[[91.5,27.7],[93.5,27.7],[93.5,28.5],[91.5,28.5],[91.5,27.7]]]}},
            {"type":"Feature",
             "properties":{"adcode":650000,"name":"新疆维吾尔自治区","level":"province","parent":100000},
             "geometry":{"type":"Polygon","coordinates":[[[73,34],[90,34],[90,40],[73,40],[73,34]]]}},
            {"type":"Feature",
             "properties":{"adcode":652100,"name":"和田地区","level":"city","parent":650000},
             "geometry":{"type":"Polygon","coordinates":[[[77,34.5],[82,34.5],[82,36],[77,36],[77,34.5]]]}},
            {"type":"Feature",
             "properties":{"adcode":652101,"name":"和田县","level":"district","parent":652100},
             "geometry":{"type":"Polygon","coordinates":[[[78,34.7],[81,34.7],[81,35.5],[78,35.5],[78,34.7]]]}}
        ]}`)
	dvFC, err := geojson.UnmarshalFeatureCollection(dv)
	require.NoError(t, err)
	return fc, BuildCNAdminIndex(dvFC)
}

// Natural Earth places NL68kd in India; CN-admin places it in 西藏/错那县.
// The output must follow CN-admin — for a HAM tool the Chinese admin hierarchy
// is the authoritative answer for grids that CN-admin covers.
func TestGeocodeDisputedRegionCNAdminOverridesIN(t *testing.T) {
	fc, dv := disputedFixture(t)
	g := &Geocoder{
		Countries: fc,
		CountriesByCode: map[string]Country{
			"IN": {Code: "IN", Name: "India", NameZh: "印度"},
			"CN": {Code: "CN", Name: "China", NameZh: "中国"},
		},
		CNAdmin:  dv,
		KDTree: BuildKDTree(nil),
	}
	// NL68kd center = (28.145833, 92.875)
	res := g.Lookup(28.145833, 92.875)
	require.NotNil(t, res.Country)
	assert.Equal(t, "CN", res.Country.Code)
	assert.Equal(t, "China", res.Country.Name)
	assert.Equal(t, "中国", res.Country.NameZh)
	assert.True(t, res.UsedCNAdmin)
	assert.Equal(t, "西藏自治区", res.Admin1.Zh)
	assert.Equal(t, "Tibet", res.Admin1.En)
	assert.Equal(t, "错那县", res.Admin2.Zh)
}

// Same bug shape as 藏南, different polygon: 阿克赛钦 is in 新疆/和田县.
func TestGeocodeAksaiChinCNAdminOverridesIN(t *testing.T) {
	fc, dv := disputedFixture(t)
	g := &Geocoder{
		Countries: fc,
		CountriesByCode: map[string]Country{
			"IN": {Code: "IN", Name: "India", NameZh: "印度"},
			"CN": {Code: "CN", Name: "China", NameZh: "中国"},
		},
		CNAdmin:  dv,
		KDTree: BuildKDTree(nil),
	}
	res := g.Lookup(35.0, 79.5)
	require.NotNil(t, res.Country)
	assert.Equal(t, "CN", res.Country.Code)
	assert.True(t, res.UsedCNAdmin)
	assert.Equal(t, "新疆维吾尔自治区", res.Admin1.Zh)
	assert.Equal(t, "Xinjiang", res.Admin1.En)
	assert.Equal(t, "和田县", res.Admin2.Zh)
}

// CN-admin-derived SAR mapping: 香港特别行政区 must resolve to country=HK even
// when Natural Earth (hypothetically or before SAR split-out) classifies the
// area as something else. Verifies the province→ISO mapping for SARs.
func TestGeocodeHKSARMappingFromCNAdmin(t *testing.T) {
	ne := []byte(`{
        "type":"FeatureCollection",
        "features":[
            {"type":"Feature","properties":{"iso_a2":"CN","name_en":"China","name_zh":"中国"},
             "geometry":{"type":"Polygon","coordinates":[[[113,21],[115,21],[115,23],[113,23],[113,21]]]}}
        ]}`)
	fc, err := geojson.UnmarshalFeatureCollection(ne)
	require.NoError(t, err)
	dv := []byte(`{
        "type":"FeatureCollection",
        "features":[
            {"type":"Feature",
             "properties":{"adcode":810000,"name":"香港特别行政区","level":"province","parent":100000},
             "geometry":{"type":"Polygon","coordinates":[[[114,22],[114.5,22],[114.5,22.5],[114,22.5],[114,22]]]}},
            {"type":"Feature",
             "properties":{"adcode":810001,"name":"中西区","level":"district","parent":810000},
             "geometry":{"type":"Polygon","coordinates":[[[114.1,22.2],[114.2,22.2],[114.2,22.3],[114.1,22.3],[114.1,22.2]]]}}
        ]}`)
	dvFC, err := geojson.UnmarshalFeatureCollection(dv)
	require.NoError(t, err)

	g := &Geocoder{
		Countries: fc,
		CountriesByCode: map[string]Country{
			"CN": {Code: "CN", Name: "China", NameZh: "中国"},
			"HK": {Code: "HK", Name: "Hong Kong", NameZh: "香港"},
		},
		CNAdmin:  BuildCNAdminIndex(dvFC),
		KDTree: BuildKDTree(nil),
	}
	res := g.Lookup(22.25, 114.15)
	require.NotNil(t, res.Country)
	assert.Equal(t, "HK", res.Country.Code)
	assert.Equal(t, "Hong Kong", res.Country.Name)
	assert.True(t, res.UsedCNAdmin)
	assert.Equal(t, "香港特别行政区", res.Admin1.Zh)
}

// Regression: a point inside India but outside any CN-admin polygon must NOT
// be flipped to CN. New Delhi sits well clear of disputed boundaries.
func TestGeocodeIndiaNotOverriddenWhenOutsideCNAdmin(t *testing.T) {
	fc, dv := disputedFixture(t)
	g := &Geocoder{
		Countries: fc,
		CountriesByCode: map[string]Country{
			"IN": {Code: "IN", Name: "India", NameZh: "印度"},
			"CN": {Code: "CN", Name: "China", NameZh: "中国"},
		},
		CNAdmin:  dv,
		KDTree: BuildKDTree(nil),
	}
	// New Delhi (28.6139, 77.2090) is inside the IN fixture, outside all CN-admin polygons.
	res := g.Lookup(28.6139, 77.2090)
	require.NotNil(t, res.Country)
	assert.Equal(t, "IN", res.Country.Code)
	assert.False(t, res.UsedCNAdmin)
}

// Polygon lookup misses (offshore) but nearest city has a country code —
// CountriesByCode fallback recovers it.
func TestGeocodeFallsBackToCityCountryWhenPolygonMisses(t *testing.T) {
	fc := testCountries(t)
	cities := []City{
		{Name: "CoastTown", NameZh: "海滨", Lat: 5, Lon: 5, CountryCode: "XX", Admin1Code: "AA", Admin2Code: "001"},
	}
	byCode := map[string]Country{
		"XX": {Code: "XX", Name: "Testland", NameZh: "测试国"},
	}
	g := &Geocoder{
		Countries:       fc,
		CountriesByCode: byCode,
		KDTree:          BuildKDTree(cities),
		Admin1:          map[string]AdminEntry{"XX.AA": {En: "Prov", Zh: "省"}},
		Admin2:          map[string]AdminEntry{"XX.AA.001": {En: "City", Zh: "市"}},
	}

	// Query at (15, 15) — outside XX polygon (0,0)-(10,10); polygon misses.
	// Nearest city is still CoastTown in XX, so country gets filled by code.
	res := g.Lookup(15, 15)
	require.NotNil(t, res.Country)
	assert.Equal(t, "XX", res.Country.Code)
	assert.Equal(t, "Testland", res.Country.Name)
	assert.Equal(t, "测试国", res.Country.NameZh)
	assert.Equal(t, "CoastTown", res.CityName)
	assert.Equal(t, "省", res.Admin1.Zh)
}
