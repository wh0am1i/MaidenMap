// Package middleware contains HTTP middleware for rate limiting and logging.
package middleware

import (
	"net/http"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
	"github.com/gin-gonic/gin"
)

// RateLimit returns a Gin middleware enforcing per-IP rate limiting.
// rpm is requests-per-minute allowed per client IP.
//
// The limiter keys on gin.Context.ClientIP(), which honors the router's
// SetTrustedProxies configuration. This matters: tollbooth's built-in
// LimitByRequest trusts X-Forwarded-For unconditionally, which lets any
// client rotate a fake header per request and bypass the limit. Keying on
// ClientIP() means X-Forwarded-For is only trusted when the direct peer
// (RemoteAddr) is in the configured trusted-proxies CIDR list.
func RateLimit(rpm int) gin.HandlerFunc {
	lmt := tollbooth.NewLimiter(float64(rpm)/60.0, &limiter.ExpirableOptions{
		DefaultExpirationTTL: 3600, // seconds
	})

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if ip == "" {
			ip = c.Request.RemoteAddr
		}
		if httpErr := tollbooth.LimitByKeys(lmt, []string{ip}); httpErr != nil {
			c.Writer.Header().Set("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limited",
				"message": "too many requests",
			})
			return
		}
		c.Next()
	}
}
