package handler

import (
	"math"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wh0am1i/maidenmap/api/internal/data"
	"github.com/wh0am1i/maidenmap/api/internal/geocode"
	"github.com/wh0am1i/maidenmap/api/internal/maidenhead"
)

const maxBatchSize = 100

type biName struct {
	En string `json:"en"`
	Zh string `json:"zh"`
}

type centerResp struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type countryResp struct {
	Code string `json:"code"`
	Name biName `json:"name"`
}

type gridResponse struct {
	Grid    string       `json:"grid"`
	Center  centerResp   `json:"center"`
	Country *countryResp `json:"country"`
	Admin1  biName       `json:"admin1"`
	Admin2  biName       `json:"admin2"`
	City    biName       `json:"city"`

	// usedDataV is set when the admin fields were resolved via the DataV
	// polygon index (CN family). The SAR-transform branch uses it to skip
	// the admin1→admin2 swap that only makes sense for the GeoNames path.
	usedDataV bool `json:"-"`
}

type gridError struct {
	Grid    string `json:"grid"`
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// GridSingle handles GET /api/grid/:code.
func GridSingle(ds *data.Dataset) gin.HandlerFunc {
	g := &geocode.Geocoder{
		Countries: ds.Countries, CountriesByCode: ds.CountriesByCode,
		DataV:  ds.DataV,
		KDTree: ds.KDTree, Admin1: ds.Admin1, Admin2: ds.Admin2,
	}
	return func(c *gin.Context) {
		resp, err := resolve(c.Param("code"), g)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grid", "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

// GridBatch handles GET /api/grid?codes=A,B,C.
func GridBatch(ds *data.Dataset) gin.HandlerFunc {
	g := &geocode.Geocoder{
		Countries: ds.Countries, CountriesByCode: ds.CountriesByCode,
		DataV:  ds.DataV,
		KDTree: ds.KDTree, Admin1: ds.Admin1, Admin2: ds.Admin2,
	}
	return func(c *gin.Context) {
		raw := c.Query("codes")
		if strings.TrimSpace(raw) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing_codes", "message": "query param 'codes' is required"})
			return
		}
		codes := strings.Split(raw, ",")
		if len(codes) > maxBatchSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": "too_many_codes", "message": "at most 100 codes per request"})
			return
		}

		type item any
		results := make([]item, 0, len(codes))
		for _, code := range codes {
			code = strings.TrimSpace(code)
			resp, err := resolve(code, g)
			if err != nil {
				results = append(results, gridError{Grid: code, Error: "invalid_grid", Message: err.Error()})
				continue
			}
			results = append(results, resp)
		}
		c.JSON(http.StatusOK, gin.H{"results": results})
	}
}

func resolve(code string, g *geocode.Geocoder) (gridResponse, error) {
	loc, err := maidenhead.Parse(code)
	if err != nil {
		return gridResponse{}, err
	}
	r := g.Lookup(loc.Lat, loc.Lon)
	resp := gridResponse{
		Grid:      loc.Grid,
		Center:    centerResp{Lat: round4(loc.Lat), Lon: round4(loc.Lon)},
		Admin1:    biName{En: r.Admin1.En, Zh: r.Admin1.Zh},
		Admin2:    biName{En: r.Admin2.En, Zh: r.Admin2.Zh},
		City:      biName{En: r.CityName, Zh: r.CityNameZh},
		usedDataV: r.UsedDataV,
	}
	if r.Country != nil {
		resp.Country = &countryResp{Code: r.Country.Code, Name: biName{En: r.Country.Name, Zh: r.Country.NameZh}}
	}
	applyChinaSARTransform(&resp)
	return resp, nil
}

// China PRC's People's Republic identity used when we subsume HK/MO/TW.
var prcCountry = countryResp{
	Code: "CN",
	Name: biName{En: "People's Republic of China", Zh: "中华人民共和国"},
}

// sarAsAdmin1 maps ISO alpha-2 of a Chinese SAR to the admin1 entry we insert
// when demoting it under CN. HK and MO have no natural province-level admin
// in GeoNames, so shifting the original admin1 (the district) into admin2
// makes room for the SAR name at admin1. Uses the full "特别行政区" form
// per the project's China-map labelling convention.
var sarAsAdmin1 = map[string]biName{
	"HK": {En: "Hong Kong Special Administrative Region", Zh: "香港特别行政区"},
	"MO": {En: "Macao Special Administrative Region", Zh: "澳门特别行政区"},
}

// applyChinaSARTransform folds Hong Kong, Macao, and Taiwan query results
// into mainland China per the project's product decision:
//
//   - Country always becomes CN for HK / MO / TW.
//   - When DataV populated the admin fields, admin1 already holds the SAR /
//     province name (e.g. 香港特别行政区, 台湾省) and admin2 the district, so
//     no swap is needed.
//   - Otherwise (GeoNames-only path, e.g. DataV missing) HK / MO admin1 is a
//     district; shift it down to admin2 and insert the SAR name at admin1.
//
// No-op for any other country.
func applyChinaSARTransform(resp *gridResponse) {
	if resp.Country == nil {
		return
	}
	code := resp.Country.Code
	if code != "HK" && code != "MO" && code != "TW" {
		return
	}
	resp.Country = &prcCountry
	if resp.usedDataV {
		return
	}
	if sar, ok := sarAsAdmin1[code]; ok {
		resp.Admin2 = resp.Admin1
		resp.Admin1 = sar
	}
}

func round4(f float64) float64 {
	return math.Round(f*10000) / 10000
}
