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

func TestApplyChinaSARTransformHongKong(t *testing.T) {
	resp := gridResponse{
		Country: &countryResp{
			Code: "HK",
			Name: biName{En: "Hong Kong", Zh: "香港"},
		},
		Admin1: biName{En: "Kwun Tong District", Zh: "观塘"},
		Admin2: biName{En: "", Zh: ""},
		City:   biName{En: "Kwun Tong", Zh: "观塘"},
	}
	applyChinaSARTransform(&resp)

	// Country rewritten to CN
	require.NotNil(t, resp.Country)
	assert.Equal(t, "CN", resp.Country.Code)
	assert.Equal(t, "People's Republic of China", resp.Country.Name.En)
	assert.Equal(t, "中华人民共和国", resp.Country.Name.Zh)

	// SAR name takes admin1; original admin1 (district) drops to admin2
	assert.Equal(t, "Hong Kong Special Administrative Region", resp.Admin1.En)
	assert.Equal(t, "香港特别行政区", resp.Admin1.Zh)
	assert.Equal(t, "Kwun Tong District", resp.Admin2.En)
	assert.Equal(t, "观塘", resp.Admin2.Zh)

	// City unchanged
	assert.Equal(t, "Kwun Tong", resp.City.En)
	assert.Equal(t, "观塘", resp.City.Zh)
}

func TestApplyChinaSARTransformHongKongUsingDataV(t *testing.T) {
	// When DataV populated admin fields, admin1 already carries the SAR name
	// and admin2 the district — the handler should only rewrite country and
	// leave the hierarchy alone (no swap).
	resp := gridResponse{
		Country: &countryResp{
			Code: "HK",
			Name: biName{En: "Hong Kong", Zh: "香港"},
		},
		Admin1:    biName{En: "Hong Kong Special Administrative Region", Zh: "香港特别行政区"},
		Admin2:    biName{En: "", Zh: "观塘区"},
		City:      biName{En: "Kwun Tong", Zh: "观塘"},
		usedDataV: true,
	}
	applyChinaSARTransform(&resp)

	assert.Equal(t, "CN", resp.Country.Code)
	assert.Equal(t, "香港特别行政区", resp.Admin1.Zh)
	assert.Equal(t, "观塘区", resp.Admin2.Zh)
}

func TestApplyChinaSARTransformTaiwan(t *testing.T) {
	// Taiwan already has a province/county admin hierarchy — just rewrite country.
	resp := gridResponse{
		Country: &countryResp{
			Code: "TW",
			Name: biName{En: "Taiwan", Zh: "台湾"},
		},
		Admin1: biName{En: "Taiwan", Zh: "台湾省"},
		Admin2: biName{En: "Miaoli", Zh: "苗栗县"},
		City:   biName{En: "Miaoli", Zh: "苗栗"},
	}
	applyChinaSARTransform(&resp)

	assert.Equal(t, "CN", resp.Country.Code)
	assert.Equal(t, "中华人民共和国", resp.Country.Name.Zh)
	// Admin hierarchy preserved
	assert.Equal(t, "台湾省", resp.Admin1.Zh)
	assert.Equal(t, "苗栗县", resp.Admin2.Zh)
}

func TestApplyChinaSARTransformNoOpForOtherCountries(t *testing.T) {
	resp := gridResponse{
		Country: &countryResp{
			Code: "DE",
			Name: biName{En: "Germany", Zh: "德国"},
		},
		Admin1: biName{En: "Berlin", Zh: "柏林"},
	}
	applyChinaSARTransform(&resp)

	assert.Equal(t, "DE", resp.Country.Code)
	assert.Equal(t, "德国", resp.Country.Name.Zh)
	assert.Equal(t, "Berlin", resp.Admin1.En)
}

func TestApplyChinaSARTransformHandlesNilCountry(t *testing.T) {
	// Ocean point — no country. Should not panic.
	resp := gridResponse{Country: nil}
	applyChinaSARTransform(&resp)
	assert.Nil(t, resp.Country)
}
