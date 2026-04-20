package geocode

import (
	"testing"

	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Two adjacent districts + a province container. The points exercise the
// district-first match, the province fallback, and an out-of-bounds miss.
func TestDataVIndexLookup(t *testing.T) {
	raw := []byte(`{
        "type": "FeatureCollection",
        "features": [
            {"type":"Feature",
             "properties":{"adcode":330000,"name":"浙江省","level":"province","parent":100000},
             "geometry":{"type":"Polygon","coordinates":[[[118,28],[123,28],[123,32],[118,32],[118,28]]]}},
            {"type":"Feature",
             "properties":{"adcode":330100,"name":"杭州市","level":"city","parent":330000},
             "geometry":{"type":"Polygon","coordinates":[[[119,30],[121,30],[121,31],[119,31],[119,30]]]}},
            {"type":"Feature",
             "properties":{"adcode":330106,"name":"西湖区","level":"district","parent":330100},
             "geometry":{"type":"Polygon","coordinates":[[[120,30.1],[120.3,30.1],[120.3,30.4],[120,30.4],[120,30.1]]]}},
            {"type":"Feature",
             "properties":{"adcode":810000,"name":"香港特别行政区","level":"province","parent":100000},
             "geometry":{"type":"Polygon","coordinates":[[[114,22],[114.5,22],[114.5,22.6],[114,22.6],[114,22]]]}},
            {"type":"Feature",
             "properties":{"adcode":810017,"name":"观塘区","level":"district","parent":810000},
             "geometry":{"type":"Polygon","coordinates":[[[114.2,22.3],[114.23,22.3],[114.23,22.33],[114.2,22.33],[114.2,22.3]]]}}
        ]}`)

	fc, err := geojson.UnmarshalFeatureCollection(raw)
	require.NoError(t, err)

	idx := BuildDataVIndex(fc)

	// Point inside 西湖区 — exercises district match with full hierarchy.
	got, ok := idx.Lookup(30.2, 120.15)
	require.True(t, ok)
	assert.Equal(t, 330106, got.DistrictAD)
	assert.Equal(t, "西湖区", got.District)
	assert.Equal(t, "杭州市", got.City)
	assert.Equal(t, "浙江省", got.Province)

	// Inside Hangzhou's city polygon but outside 西湖区 — district lookup
	// misses, so the point-in-polygon test falls through to the province.
	got, ok = idx.Lookup(30.5, 120.5)
	require.True(t, ok)
	assert.Equal(t, "浙江省", got.Province)
	assert.Empty(t, got.District)
	assert.Empty(t, got.City) // we don't index city polygons for lookup, only districts/provinces

	// Inside 观塘区 (HK) — district's parent is directly the SAR province.
	got, ok = idx.Lookup(22.31, 114.21)
	require.True(t, ok)
	assert.Equal(t, "观塘区", got.District)
	assert.Equal(t, "香港特别行政区", got.Province)
	assert.Empty(t, got.City) // HK has no city level in DataV

	// Point in the sea outside every polygon.
	_, ok = idx.Lookup(0, 0)
	assert.False(t, ok)
}

func TestDataVIndexNilSafe(t *testing.T) {
	var idx *DataVIndex
	_, ok := idx.Lookup(30, 120)
	assert.False(t, ok)
}
