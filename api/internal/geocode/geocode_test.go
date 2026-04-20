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

// For CN-family queries, DataV should override GeoNames for admin1, admin2
// AND city — otherwise the "最近城市" field leaks wrong data (e.g. the grid
// center sits in 西湖区/杭州市 but the nearest GeoNames city15000 is Fuyang).
func TestGeocodeCNFamilyUsesDataVForCity(t *testing.T) {
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

	// DataV: three-level hierarchy — 浙江省 → 杭州市 → 西湖区 — covering the query point.
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
		DataV:           BuildDataVIndex(dvFC),
		KDTree:          BuildKDTree(cities),
		Admin1:          map[string]AdminEntry{"CN.02": {En: "Zhejiang", Zh: ""}},
		Admin2:          map[string]AdminEntry{"CN.02.0203": {En: "Fuyang", Zh: ""}},
	}

	res := g.Lookup(30.15, 120.04)
	require.NotNil(t, res.Country)
	assert.Equal(t, "CN", res.Country.Code)
	assert.True(t, res.UsedDataV, "DataV should have hit for this CN point")
	assert.Equal(t, "浙江省", res.Admin1.Zh)
	assert.Equal(t, "Zhejiang", res.Admin1.En)
	assert.Equal(t, "西湖区", res.Admin2.Zh)
	// City must follow DataV — not GeoNames' Fuyang. English is the pinyin
	// romanization of the Chinese city (the standard international form).
	assert.Equal(t, "杭州市", res.CityNameZh)
	assert.Equal(t, "Hangzhou", res.CityName)
	// Same for admin2 — pinyin-derived English with "District" suffix.
	assert.Equal(t, "Xihu District", res.Admin2.En)
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
