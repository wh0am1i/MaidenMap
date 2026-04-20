package handler

import (
	"encoding/json"
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

func TestHealthOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ds := &data.Dataset{
		Countries: &geojson.FeatureCollection{Features: []*geojson.Feature{{}}},
		Cities:    []geocode.City{{Name: "A"}, {Name: "B"}},
		UpdatedAt: time.Date(2026, 4, 1, 3, 0, 0, 0, time.UTC),
	}
	r := gin.New()
	r.GET("/api/health", Health(ds))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "ok", body["status"])
	assert.EqualValues(t, 2, body["cities_count"])
	assert.EqualValues(t, 1, body["countries_count"])
	assert.Equal(t, "2026-04-01T03:00:00Z", body["data_updated_at"])
}
