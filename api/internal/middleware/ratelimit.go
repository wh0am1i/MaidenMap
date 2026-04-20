// Package middleware contains HTTP middleware for rate limiting and logging.
package middleware

import (
	"net/http"
	"strconv"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
	"github.com/gin-gonic/gin"
)

// RateLimit returns a Gin middleware enforcing per-IP rate limiting.
// rpm is requests-per-minute allowed per client IP.
func RateLimit(rpm int) gin.HandlerFunc {
	lmt := tollbooth.NewLimiter(float64(rpm)/60.0, &limiter.ExpirableOptions{
		DefaultExpirationTTL: 3600, // seconds
	})
	lmt.SetIPLookups([]string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"})
	lmt.SetMessage(`{"error":"rate_limited","message":"too many requests"}`)
	lmt.SetMessageContentType("application/json")

	return func(c *gin.Context) {
		if httpErr := tollbooth.LimitByRequest(lmt, c.Writer, c.Request); httpErr != nil {
			c.Writer.Header().Set("Retry-After", strconv.Itoa(60))
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
	}
}
