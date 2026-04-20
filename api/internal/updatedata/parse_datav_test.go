package updatedata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchDataVChinaDrillsToDistricts(t *testing.T) {
	// Stubs a miniature DataV tree:
	//   country 100000 → 1 province 330000
	//   province 330000 → 1 city 330100
	//   city 330100 → 2 districts 330106, 330110
	// Also provides HK 810000 that skips the city level.
	handler := func(w http.ResponseWriter, r *http.Request) {
		// URL like /100000_full.json — pull the adcode out.
		name := strings.TrimPrefix(r.URL.Path, "/")
		name = strings.TrimSuffix(name, "_full.json")
		ad, err := strconv.Atoi(name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		switch ad {
		case 100000:
			writeFC(w, []stubFeature{
				{adcode: 330000, name: "浙江省", level: "province", children: 1, parent: 100000},
				{adcode: 810000, name: "香港特别行政区", level: "province", children: 18, parent: 100000},
			})
		case 330000:
			writeFC(w, []stubFeature{
				{adcode: 330100, name: "杭州市", level: "city", children: 2, parent: 330000},
			})
		case 330100:
			writeFC(w, []stubFeature{
				{adcode: 330106, name: "西湖区", level: "district", parent: 330100},
				{adcode: 330110, name: "余杭区", level: "district", parent: 330100},
			})
		case 810000:
			writeFC(w, []stubFeature{
				{adcode: 810017, name: "观塘区", level: "district", parent: 810000},
			})
		default:
			http.NotFound(w, r)
		}
	}
	srv := httptest.NewServer(http.HandlerFunc(handler))
	defer srv.Close()

	nodes, err := FetchDataVChina(srv.URL, 4)
	require.NoError(t, err)

	// Expected:
	//   2 provinces (from level 1)
	//   1 city + 1 district (HK direct) from level 2
	//   2 districts from level 3
	// Total: 6 nodes.
	assert.Len(t, nodes, 6)

	byAD := map[int]DataVNode{}
	for _, n := range nodes {
		byAD[n.ADCode] = n
	}
	assert.Equal(t, "浙江省", byAD[330000].Name)
	assert.Equal(t, "province", byAD[330000].Level)
	assert.Equal(t, 100000, byAD[330000].ParentAD)

	assert.Equal(t, "杭州市", byAD[330100].Name)
	assert.Equal(t, "city", byAD[330100].Level)
	assert.Equal(t, 330000, byAD[330100].ParentAD)

	assert.Equal(t, "西湖区", byAD[330106].Name)
	assert.Equal(t, "district", byAD[330106].Level)
	assert.Equal(t, 330100, byAD[330106].ParentAD)

	// HK drills straight to district (no city level).
	assert.Equal(t, "观塘区", byAD[810017].Name)
	assert.Equal(t, 810000, byAD[810017].ParentAD)
}

// Real DataV level-1 mixes int adcodes with one string marker feature
// ("100000_JD" for the Nine-Dash Line). The parser must drop the string
// feature and keep going rather than failing the whole decode.
func TestFetchDataVChinaHandlesStringAdcodeMarker(t *testing.T) {
	body := `{"type":"FeatureCollection","features":[
        {"type":"Feature","properties":{"adcode":110000,"name":"北京市","level":"province","childrenNum":1,"parent":{"adcode":100000}},
         "geometry":{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,1],[0,0]]]}},
        {"type":"Feature","properties":{"adcode":"100000_JD","name":"九段线","level":"nation"},
         "geometry":{"type":"MultiLineString","coordinates":[]}}
    ]}`
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/100000_full.json" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(body))
			return
		}
		// Empty children for 110000 keeps the test single-level.
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"type":"FeatureCollection","features":[]}`))
	}
	srv := httptest.NewServer(http.HandlerFunc(handler))
	defer srv.Close()

	nodes, err := FetchDataVChina(srv.URL, 1)
	require.NoError(t, err)
	// String-adcode marker is filtered out; only 北京市 survives.
	require.Len(t, nodes, 1)
	assert.Equal(t, 110000, nodes[0].ADCode)
	assert.Equal(t, "北京市", nodes[0].Name)
}

func TestFetchDataVChinaContinuesOnChildError(t *testing.T) {
	// Simulate Taiwan: country returns it, but the drill _full returns 404.
	// We must not fail the whole run; we just log and keep the province node.
	handler := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/100000_full.json" {
			writeFC(w, []stubFeature{
				{adcode: 710000, name: "台湾省", level: "province", children: 1, parent: 100000},
			})
			return
		}
		http.NotFound(w, r)
	}
	srv := httptest.NewServer(http.HandlerFunc(handler))
	defer srv.Close()

	nodes, err := FetchDataVChina(srv.URL, 2)
	require.NoError(t, err)
	assert.Len(t, nodes, 1)
	assert.Equal(t, "台湾省", nodes[0].Name)
}

// --- helpers ---

type stubFeature struct {
	adcode   int
	name     string
	level    string
	children int
	parent   int
}

func writeFC(w http.ResponseWriter, feats []stubFeature) {
	type geom struct {
		Type        string        `json:"type"`
		Coordinates [][][]float64 `json:"coordinates"`
	}
	type feat struct {
		Type       string          `json:"type"`
		Properties DataVProperties `json:"properties"`
		Geometry   geom            `json:"geometry"`
	}
	type fc struct {
		Type     string `json:"type"`
		Features []feat `json:"features"`
	}

	out := fc{Type: "FeatureCollection"}
	for _, f := range feats {
		props := DataVProperties{
			ADCode:      flexInt(f.adcode),
			Name:        f.name,
			Level:       f.level,
			ChildrenNum: f.children,
		}
		props.Parent.ADCode = flexInt(f.parent)
		out.Features = append(out.Features, feat{
			Type:       "Feature",
			Properties: props,
			Geometry: geom{
				Type:        "Polygon",
				Coordinates: [][][]float64{{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}},
			},
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
	_ = fmt.Sprintf // silence unused
}
