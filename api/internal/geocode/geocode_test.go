package geocode

import (
	"testing"

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
