package handler

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wh0am1i/maidenmap/api/internal/data"
	"github.com/wh0am1i/maidenmap/api/internal/geocode"
)

func newRouter(t *testing.T) *gin.Engine {
	t.Helper()
	raw := []byte(`{"type":"FeatureCollection","features":[
        {"type":"Feature","properties":{"iso_a2":"DE","name_en":"Germany"},
         "geometry":{"type":"Polygon","coordinates":[[[5,47],[16,47],[16,55],[5,55],[5,47]]]}}
    ]}`)
	fc, err := geojson.UnmarshalFeatureCollection(raw)
	require.NoError(t, err)

	cities := []geocode.City{
		{Name: "Cottbus", Lat: 51.76, Lon: 14.33, CountryCode: "DE", Admin1Code: "BR", Admin2Code: "12071"},
	}
	ds := &data.Dataset{
		Countries: fc,
		Cities:    cities,
		KDTree:    geocode.BuildKDTree(cities),
		Admin1:    map[string]string{"DE.BR": "Brandenburg"},
		Admin2:    map[string]string{"DE.BR.12071": "Spree-Neiße"},
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/grid/:code", GridSingle(ds))
	r.GET("/api/grid", GridBatch(ds))
	return r
}

func TestGridSingle(t *testing.T) {
	r := newRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/grid/JO61", nil) // JO61 center (13, 51.5), inside DE polygon
	r.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "JO61", body["grid"])
	center := body["center"].(map[string]any)
	assert.InDelta(t, 13.0, center["lon"].(float64), 1e-3)
	assert.InDelta(t, 51.5, center["lat"].(float64), 1e-3)
	country := body["country"].(map[string]any)
	assert.Equal(t, "DE", country["code"])
	assert.Equal(t, "Germany", country["name"])
	assert.Equal(t, "Brandenburg", body["admin1"])
	assert.Equal(t, "Spree-Neiße", body["admin2"])
	assert.Equal(t, "Cottbus", body["city"])
}

func TestGridSingleInvalid(t *testing.T) {
	r := newRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/grid/ZZ00", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "invalid_grid", body["error"])
}

func TestGridBatch(t *testing.T) {
	r := newRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/grid?codes=JO61,ZZ00", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	var body struct {
		Results []map[string]any `json:"results"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Results, 2)
	assert.Equal(t, "JO61", body.Results[0]["grid"])
	assert.Equal(t, "ZZ00", body.Results[1]["grid"])
	assert.Equal(t, "invalid_grid", body.Results[1]["error"])
}

func TestGridBatchMissing(t *testing.T) {
	r := newRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/grid", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestGridBatchTooMany(t *testing.T) {
	r := newRouter(t)
	codes := ""
	for i := 0; i < 101; i++ {
		if i > 0 {
			codes += ","
		}
		codes += "JO65"
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/grid?codes="+codes, nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}
