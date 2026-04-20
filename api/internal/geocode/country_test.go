package geocode

import (
	"testing"

	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCountries(t *testing.T) *geojson.FeatureCollection {
	t.Helper()
	raw := []byte(`{
        "type":"FeatureCollection",
        "features":[
            {"type":"Feature","properties":{"iso_a2":"XX","name_en":"Testland","name_zh":"测试国"},
             "geometry":{"type":"Polygon","coordinates":[[[0,0],[10,0],[10,10],[0,10],[0,0]]]}},
            {"type":"Feature","properties":{"iso_a2":"YY","name_en":"Overseas","name_zh":"海外"},
             "geometry":{"type":"MultiPolygon","coordinates":[
                [[[20,20],[30,20],[30,30],[20,30],[20,20]]],
                [[[40,40],[50,40],[50,50],[40,50],[40,40]]]
             ]}}
        ]}`)
	fc, err := geojson.UnmarshalFeatureCollection(raw)
	require.NoError(t, err)
	return fc
}

func TestLookupCountryInsidePolygon(t *testing.T) {
	fc := testCountries(t)
	c, ok := LookupCountry(fc, 5, 5)
	require.True(t, ok)
	assert.Equal(t, "XX", c.Code)
	assert.Equal(t, "Testland", c.Name)
	assert.Equal(t, "测试国", c.NameZh)
}

func TestLookupCountryMultiPolygonSecondPart(t *testing.T) {
	fc := testCountries(t)
	c, ok := LookupCountry(fc, 45, 45)
	require.True(t, ok)
	assert.Equal(t, "YY", c.Code)
	assert.Equal(t, "海外", c.NameZh)
}

func TestLookupCountryNotFound(t *testing.T) {
	fc := testCountries(t)
	_, ok := LookupCountry(fc, -50, -50)
	assert.False(t, ok)
}
