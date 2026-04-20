package geocode

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeocodeLandPoint(t *testing.T) {
	fc := testCountries(t)
	cities := []City{
		{Name: "TestCapital", Lat: 5, Lon: 5, CountryCode: "XX", Admin1Code: "AA", Admin2Code: "001"},
	}
	tree := BuildKDTree(cities)
	a1 := map[string]string{"XX.AA": "Test Admin1"}
	a2 := map[string]string{"XX.AA.001": "Test Admin2"}

	g := &Geocoder{Countries: fc, KDTree: tree, Admin1: a1, Admin2: a2}
	res := g.Lookup(5, 5)

	require.NotNil(t, res.Country)
	assert.Equal(t, "XX", res.Country.Code)
	assert.Equal(t, "Testland", res.Country.Name)
	assert.Equal(t, "Test Admin1", res.Admin1)
	assert.Equal(t, "Test Admin2", res.Admin2)
	assert.Equal(t, "TestCapital", res.City)
}

func TestGeocodeOcean(t *testing.T) {
	fc := testCountries(t)
	g := &Geocoder{Countries: fc, KDTree: BuildKDTree(nil), Admin1: nil, Admin2: nil}
	res := g.Lookup(-50, -50)
	assert.Nil(t, res.Country)
	assert.Empty(t, res.Admin1)
	assert.Empty(t, res.Admin2)
	assert.Empty(t, res.City)
}

func TestGeocodeNoCities(t *testing.T) {
	fc := testCountries(t)
	g := &Geocoder{Countries: fc, KDTree: BuildKDTree(nil), Admin1: nil, Admin2: nil}
	res := g.Lookup(5, 5)
	require.NotNil(t, res.Country)
	assert.Equal(t, "XX", res.Country.Code)
	assert.Empty(t, res.City)
}
