package updatedata

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mini fixture covering the gb-to-adcode rules at each level plus the
// boundary-line filter. Exercises the same shape Tianditu ships.
func TestLoadTiandituSyntheticThreeLevel(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		tianditProvinceFile: `{"type":"FeatureCollection","features":[
            {"type":"Feature","properties":{"gb":"156540000","name":"西藏自治区"},
             "geometry":{"type":"Polygon","coordinates":[[[78,27],[96,27],[96,36],[78,36],[78,27]]]}},
            {"type":"Feature","properties":{"name":"境界线"},
             "geometry":{"type":"MultiLineString","coordinates":[[[0,0],[1,1]]]}}
        ]}`,
		tianditCityFile: `{"type":"FeatureCollection","features":[
            {"type":"Feature","properties":{"gb":"156540400","name":"林芝市"},
             "geometry":{"type":"Polygon","coordinates":[[[93,28],[97,28],[97,30],[93,30],[93,28]]]}},
            {"type":"Feature","properties":{"gb":"156659001","name":"石河子市"},
             "geometry":{"type":"Polygon","coordinates":[[[85,44],[86,44],[86,45],[85,45],[85,44]]]}}
        ]}`,
		tianditCountyFile: `{"type":"FeatureCollection","features":[
            {"type":"Feature","properties":{"gb":"156540423","name":"墨脱县"},
             "geometry":{"type":"Polygon","coordinates":[[[94,28],[96,28],[96,30],[94,30],[94,28]]]}}
        ]}`,
	}
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0644))
	}

	nodes, err := LoadTianditu(dir, "")
	require.NoError(t, err)
	require.Len(t, nodes, 4) // 1 province (境界线 filtered) + 2 cities + 1 district

	byAD := map[int]CNAdminNode{}
	for _, n := range nodes {
		byAD[n.ADCode] = n
	}

	// Province: parent collapses to country root.
	assert.Equal(t, "province", byAD[540000].Level)
	assert.Equal(t, "西藏自治区", byAD[540000].Name)
	assert.Equal(t, countryAD, byAD[540000].ParentAD)

	// Normal prefecture-level city: parent is the province.
	assert.Equal(t, "city", byAD[540400].Level)
	assert.Equal(t, 540000, byAD[540400].ParentAD)

	// 县级市 directly under province (no prefecture): derived parent
	// 650000 still happens to be the province — the city-rule (last-4-zero)
	// does the right thing here even when the prefecture slot is "90".
	assert.Equal(t, "city", byAD[659001].Level)
	assert.Equal(t, 650000, byAD[659001].ParentAD)

	// District: parent is the city.
	assert.Equal(t, "district", byAD[540423].Level)
	assert.Equal(t, 540400, byAD[540423].ParentAD)
}

func TestLoadTiandituRejectsBadGB(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{tianditProvinceFile, tianditCityFile, tianditCountyFile} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, n),
			[]byte(`{"type":"FeatureCollection","features":[
                {"type":"Feature","properties":{"gb":"840012345","name":"NotChina"},"geometry":null},
                {"type":"Feature","properties":{"gb":"abc","name":"BadFormat"},"geometry":null},
                {"type":"Feature","properties":{"gb":"156000000","name":"ZeroAD"},"geometry":null}
            ]}`), 0644))
	}
	nodes, err := LoadTianditu(dir, "")
	require.NoError(t, err)
	assert.Empty(t, nodes, "non-156-prefix, malformed, and zero adcodes must all be dropped")
}

// Patch features carry their full CNAdminNode shape in properties since they
// don't follow GB/T 2260 — verify the loader passes them through verbatim.
func TestLoadTiandituWithPatch(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{tianditProvinceFile, tianditCityFile, tianditCountyFile} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, n),
			[]byte(`{"type":"FeatureCollection","features":[]}`), 0644))
	}
	patchPath := filepath.Join(dir, "patch.geojson")
	require.NoError(t, os.WriteFile(patchPath, []byte(`{
        "type":"FeatureCollection","features":[
            {"type":"Feature",
             "properties":{"adcode":549999,"name":"藏南地区","level":"patch","parent":540000},
             "geometry":{"type":"Polygon","coordinates":[[[95,27.3],[96,27.3],[96,28.3],[95,28.3],[95,27.3]]]}}
        ]}`), 0644))

	nodes, err := LoadTianditu(dir, patchPath)
	require.NoError(t, err)
	require.Len(t, nodes, 1)
	assert.Equal(t, 549999, nodes[0].ADCode)
	assert.Equal(t, "patch", nodes[0].Level)
	assert.Equal(t, 540000, nodes[0].ParentAD)
	assert.Equal(t, "藏南地区", nodes[0].Name)
}

// HK / MO districts have no prefecture-level tier above them — the default
// parent rule (drop last two digits) lands at a code that doesn't exist
// (810100 has no city or province feature). repairDistrictParents retargets
// these to the province (drop last four digits → 810000 = 香港特别行政区).
func TestLoadTiandituRepairsOrphanedHKDistrictParent(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, tianditProvinceFile),
		[]byte(`{"type":"FeatureCollection","features":[
            {"type":"Feature","properties":{"gb":"156810000","name":"香港特别行政区"},
             "geometry":{"type":"Polygon","coordinates":[[[114,22],[115,22],[115,23],[114,23],[114,22]]]}}
        ]}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, tianditCityFile),
		[]byte(`{"type":"FeatureCollection","features":[
            {"type":"Feature","properties":{"gb":"156810000","name":"香港特别行政区"},
             "geometry":{"type":"Polygon","coordinates":[[[114,22],[115,22],[115,23],[114,23],[114,22]]]}}
        ]}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, tianditCountyFile),
		[]byte(`{"type":"FeatureCollection","features":[
            {"type":"Feature","properties":{"gb":"156810101","name":"中西区"},
             "geometry":{"type":"Polygon","coordinates":[[[114.1,22.2],[114.2,22.2],[114.2,22.3],[114.1,22.3],[114.1,22.2]]]}}
        ]}`), 0644))

	nodes, err := LoadTianditu(dir, "")
	require.NoError(t, err)

	// SAR dedup: 香港特别行政区 in both 省 and 市 — we keep the province only.
	var hk, district *CNAdminNode
	for i := range nodes {
		switch nodes[i].ADCode {
		case 810000:
			hk = &nodes[i]
		case 810101:
			district = &nodes[i]
		}
	}
	require.NotNil(t, hk)
	assert.Equal(t, "province", hk.Level)

	require.NotNil(t, district)
	assert.Equal(t, "district", district.Level)
	assert.Equal(t, 810000, district.ParentAD,
		"中西区's parent must be retargeted to the SAR province (no 810100 city exists)")
}

// 澳门 ships in ALL three Tianditu files with the same gb 156820000. The
// 县 file entry would otherwise overwrite the province in byADCode and break
// SAR queries (admin1 ends up empty, country fails to flip to MO). The
// SAR-dedup rule drops level≠province entries with ad%10000==0.
func TestLoadTiandituSkipsSARInCountyAndCityFiles(t *testing.T) {
	dir := t.TempDir()
	for _, fname := range []string{tianditProvinceFile, tianditCityFile, tianditCountyFile} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, fname),
			[]byte(`{"type":"FeatureCollection","features":[
                {"type":"Feature","properties":{"gb":"156820000","name":"澳门特别行政区"},
                 "geometry":{"type":"Polygon","coordinates":[[[113.5,22.1],[113.6,22.1],[113.6,22.2],[113.5,22.2],[113.5,22.1]]]}}
            ]}`), 0644))
	}
	nodes, err := LoadTianditu(dir, "")
	require.NoError(t, err)
	require.Len(t, nodes, 1, "only the province entry survives; city and district duplicates dropped")
	assert.Equal(t, "province", nodes[0].Level)
	assert.Equal(t, 820000, nodes[0].ADCode)
}

// 县级市 directly under a province (中山市, 东莞市, 儋州市, 嘉峪关市)
// sit in the 县 file with the prefecture slot zeroed (XXYY00 with trailing
// 00). The naive district parent rule (drop last 2 digits) would give the
// city itself as parent, breaking the hierarchy walk. The XXYY00 case must
// skip the prefecture tier and parent directly to the province.
func TestLoadTiandituCountyLevelCityHasProvinceParent(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, tianditProvinceFile),
		[]byte(`{"type":"FeatureCollection","features":[
            {"type":"Feature","properties":{"gb":"156440000","name":"广东省"},
             "geometry":{"type":"Polygon","coordinates":[[[112,21],[117,21],[117,25],[112,25],[112,21]]]}}
        ]}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, tianditCityFile),
		[]byte(`{"type":"FeatureCollection","features":[]}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, tianditCountyFile),
		[]byte(`{"type":"FeatureCollection","features":[
            {"type":"Feature","properties":{"gb":"156442000","name":"中山市"},
             "geometry":{"type":"Polygon","coordinates":[[[113,22],[114,22],[114,23],[113,23],[113,22]]]}}
        ]}`), 0644))

	nodes, err := LoadTianditu(dir, "")
	require.NoError(t, err)

	var county *CNAdminNode
	for i := range nodes {
		if nodes[i].ADCode == 442000 {
			county = &nodes[i]
		}
	}
	require.NotNil(t, county)
	assert.Equal(t, "district", county.Level)
	assert.Equal(t, 440000, county.ParentAD, "中山市's parent must be 广东省, not itself")
}

func TestLoadTiandituMissingPatchIsOK(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{tianditProvinceFile, tianditCityFile, tianditCountyFile} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, n),
			[]byte(`{"type":"FeatureCollection","features":[]}`), 0644))
	}
	nodes, err := LoadTianditu(dir, filepath.Join(dir, "does_not_exist.geojson"))
	require.NoError(t, err)
	assert.Empty(t, nodes)
}
