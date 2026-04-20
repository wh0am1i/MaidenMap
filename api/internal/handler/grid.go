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
}

type gridError struct {
	Grid    string `json:"grid"`
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// GridSingle handles GET /api/grid/:code.
func GridSingle(ds *data.Dataset) gin.HandlerFunc {
	g := &geocode.Geocoder{
		Countries: ds.Countries, KDTree: ds.KDTree, Admin1: ds.Admin1, Admin2: ds.Admin2,
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
		Countries: ds.Countries, KDTree: ds.KDTree, Admin1: ds.Admin1, Admin2: ds.Admin2,
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
		Grid:   loc.Grid,
		Center: centerResp{Lat: round4(loc.Lat), Lon: round4(loc.Lon)},
		Admin1: biName{En: r.Admin1.En, Zh: r.Admin1.Zh},
		Admin2: biName{En: r.Admin2.En, Zh: r.Admin2.Zh},
		City:   biName{En: r.CityName, Zh: r.CityNameZh},
	}
	if r.Country != nil {
		resp.Country = &countryResp{Code: r.Country.Code, Name: biName{En: r.Country.Name, Zh: r.Country.NameZh}}
	}
	return resp, nil
}

func round4(f float64) float64 {
	return math.Round(f*10000) / 10000
}
