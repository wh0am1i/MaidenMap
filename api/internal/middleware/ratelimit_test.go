package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitAllowsBurstThenBlocks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimit(2)) // 2 req/min
	r.GET("/x", func(c *gin.Context) { c.Status(200) })

	hit := func() int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/x", nil)
		req.RemoteAddr = "1.2.3.4:1111"
		r.ServeHTTP(w, req)
		return w.Code
	}

	assert.Equal(t, http.StatusOK, hit())
	assert.Equal(t, http.StatusOK, hit())
	assert.Equal(t, http.StatusTooManyRequests, hit())
}

func TestRateLimitPerIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimit(1))
	r.GET("/x", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	req.RemoteAddr = "1.2.3.4:1111"
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/x", nil)
	req.RemoteAddr = "5.6.7.8:2222" // different IP
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
