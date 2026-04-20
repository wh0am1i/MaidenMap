// Package handler wires HTTP endpoints to the geocoding service.
package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wh0am1i/maidenmap/api/internal/data"
)

// Health returns a handler reporting service status and dataset metadata.
func Health(ds *data.Dataset) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":          "ok",
			"data_updated_at": ds.UpdatedAt.UTC().Format(time.RFC3339),
			"cities_count":    len(ds.Cities),
			"countries_count": len(ds.Countries.Features),
		})
	}
}
