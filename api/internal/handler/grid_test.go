package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wh0am1i/maidenmap/api/internal/data"
	"github.com/wh0am1i/maidenmap/api/internal/geocode"
)

func testDataset(t *testing.T) *data.Dataset {
	t.Helper()
	raw := []byte(`{
        "type":"FeatureCollection",
        "features":[
            {"type":"Feature","properties":{"iso_a2":"XX","name_en":"Testland","name_zh":"测试国"},
             "geometry":{"type":"Polygon","coordinates":[[[0,0],[20,0],[20,20],[0,20],[0,0]]]}}
        ]}`)
	fc, err := geojson.UnmarshalFeatureCollection(raw)
	require.NoError(t, err)

	cities := []geocode.City{
		{Name: "TestCity", NameZh: "测试市", Lat: 10, Lon: 10, CountryCode: "XX", Admin1Code: "AA", Admin2Code: "001"},
	}
	return &data.Dataset{
		Countries: fc,
		Cities:    cities,
		KDTree:    geocode.BuildKDTree(cities),
		Admin1:    map[string]geocode.AdminEntry{"XX.AA": {En: "Test Admin1", Zh: "测试一级"}},
		Admin2:    map[string]geocode.AdminEntry{"XX.AA.001": {En: "Test Admin2", Zh: "测试二级"}},
		UpdatedAt: time.Now(),
	}
}

func newRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ds := testDataset(t)
	r.GET("/api/grid/:code", GridSingle(ds))
	r.GET("/api/grid", GridBatch(ds))
	return r
}

func TestGridSingleSuccess(t *testing.T) {
	r := newRouter(t)
	w := httptest.NewRecorder()
	// JJ55: lon field J (0–20°) + digit 5 → center 11°E; lat field J (0–10°) + digit 5 → center 5.5°N.
	// Center (5.5, 11) is inside the XX polygon (0,0)–(20,20); nearest city is TestCity at (10,10).
	req := httptest.NewRequest(http.MethodGet, "/api/grid/JJ55", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	assert.Equal(t, "JJ55", body["grid"])
	center := body["center"].(map[string]any)
	assert.InDelta(t, 5.5, center["lat"], 1e-3)
	assert.InDelta(t, 11.0, center["lon"], 1e-3)

	country := body["country"].(map[string]any)
	assert.Equal(t, "XX", country["code"])
	name := country["name"].(map[string]any)
	assert.Equal(t, "Testland", name["en"])
	assert.Equal(t, "测试国", name["zh"])

	a1 := body["admin1"].(map[string]any)
	assert.Equal(t, "Test Admin1", a1["en"])
	assert.Equal(t, "测试一级", a1["zh"])

	a2 := body["admin2"].(map[string]any)
	assert.Equal(t, "Test Admin2", a2["en"])
	assert.Equal(t, "测试二级", a2["zh"])

	city := body["city"].(map[string]any)
	assert.Equal(t, "TestCity", city["en"])
	assert.Equal(t, "测试市", city["zh"])
}

func TestGridSingleInvalid(t *testing.T) {
	r := newRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/grid/BAD", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGridBatchMixed(t *testing.T) {
	r := newRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/grid?codes=JJ55,BAD,JJ55", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	results := body["results"].([]any)
	require.Len(t, results, 3)

	first := results[0].(map[string]any)
	assert.Equal(t, "JJ55", first["grid"])
	city := first["city"].(map[string]any)
	assert.Equal(t, "测试市", city["zh"])

	second := results[1].(map[string]any)
	assert.Equal(t, "invalid_grid", second["error"])
}

func TestGridBatchMissingCodes(t *testing.T) {
	r := newRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/grid", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
