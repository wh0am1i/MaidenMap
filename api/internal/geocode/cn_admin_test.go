package geocode

import (
	"testing"

	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Province + city + adjacent districts. Points exercise the district-first
// match, the city fallback when no district covers, the province fallback
// for inland gaps, and an out-of-bounds miss.
func TestCNAdminIndexLookup(t *testing.T) {
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

	idx := BuildCNAdminIndex(fc)

	// Point inside 西湖区 — exercises district match with full hierarchy.
	got, ok := idx.Lookup(30.2, 120.15)
	require.True(t, ok)
	assert.Equal(t, 330106, got.DistrictAD)
	assert.Equal(t, "西湖区", got.District)
	assert.Equal(t, "杭州市", got.City)
	assert.Equal(t, "浙江省", got.Province)

	// Inside Hangzhou's city polygon but outside 西湖区 — district misses,
	// city catches: city level populated, no district.
	got, ok = idx.Lookup(30.5, 120.5)
	require.True(t, ok)
	assert.Equal(t, "浙江省", got.Province)
	assert.Equal(t, "杭州市", got.City)
	assert.Empty(t, got.District)

	// Inside 浙江省 but outside 杭州市 — both district and city miss,
	// province catches.
	got, ok = idx.Lookup(29, 122)
	require.True(t, ok)
	assert.Equal(t, "浙江省", got.Province)
	assert.Empty(t, got.City)
	assert.Empty(t, got.District)

	// Inside 观塘区 (HK) — district's parent is directly the SAR province.
	got, ok = idx.Lookup(22.31, 114.21)
	require.True(t, ok)
	assert.Equal(t, "观塘区", got.District)
	assert.Equal(t, "香港特别行政区", got.Province)
	assert.Empty(t, got.City) // HK has no separate city level

	// Point in the sea outside every polygon.
	_, ok = idx.Lookup(0, 0)
	assert.False(t, ok)
}

// Patch tier: when a point falls outside every district / city / province
// polygon but inside a hand-curated patch polygon (covering disputed gaps
// like eastern 藏南), the result surfaces the patch name in both admin2
// (district slot) and city, inheriting the province from the patch's
// parent. Real admin polygons still win wherever they cover.
func TestCNAdminIndexPatchTier(t *testing.T) {
	// Province polygon deliberately stops at lat 29 — mimics how Tianditu's
	// 西藏自治区 polygon ends short of the eastern 藏南 claim line. The patch
	// covers the strip south of that (lat 27–29).
	raw := []byte(`{
        "type":"FeatureCollection","features":[
            {"type":"Feature",
             "properties":{"adcode":540000,"name":"西藏自治区","level":"province","parent":100000},
             "geometry":{"type":"Polygon","coordinates":[[[78,29],[96,29],[96,36],[78,36],[78,29]]]}},
            {"type":"Feature",
             "properties":{"adcode":540423,"name":"墨脱县","level":"district","parent":540000},
             "geometry":{"type":"Polygon","coordinates":[[[93,30],[95,30],[95,32],[93,32],[93,30]]]}},
            {"type":"Feature",
             "properties":{"adcode":549999,"name":"藏南地区","level":"patch","parent":540000},
             "geometry":{"type":"Polygon","coordinates":[[[91,27],[97,27],[97,29.5],[91,29.5],[91,27]]]}}
        ]}`)
	fc, err := geojson.UnmarshalFeatureCollection(raw)
	require.NoError(t, err)
	idx := BuildCNAdminIndex(fc)

	// Inside 墨脱县 — district wins, patch ignored.
	got, ok := idx.Lookup(31, 94)
	require.True(t, ok)
	assert.Equal(t, "墨脱县", got.District)
	assert.Equal(t, "西藏自治区", got.Province)

	// Inside province polygon but outside any district — province catches,
	// not patch (patch only fires when everything above misses).
	got, ok = idx.Lookup(32, 80)
	require.True(t, ok)
	assert.Equal(t, "西藏自治区", got.Province)
	assert.Empty(t, got.District)

	// Outside province polygon (south of lat 29) but inside patch — patch
	// catches and sets both admin2 and city to the patch name.
	got, ok = idx.Lookup(27.5, 95.5)
	require.True(t, ok)
	assert.Equal(t, "藏南地区", got.District)
	assert.Equal(t, "藏南地区", got.City, "patch must populate city too so the GeoNames nearest doesn't leak through")
	assert.Equal(t, "西藏自治区", got.Province)

	// Far outside any polygon — miss.
	_, ok = idx.Lookup(0, 0)
	assert.False(t, ok)
}

func TestCNAdminIndexNilSafe(t *testing.T) {
	var idx *CNAdminIndex
	_, ok := idx.Lookup(30, 120)
	assert.False(t, ok)
}
