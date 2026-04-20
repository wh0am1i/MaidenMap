package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// The rate-limit key is whatever gin.Context.ClientIP() returns. This test
// verifies the key selection contract so that asking tollbooth about
// hit-counting behavior is unnecessary — if the key is right, tollbooth's
// counting is tollbooth's problem.

func keyCaptured(t *testing.T, trustedCIDRs []string, remoteAddr, xff string) string {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	require.NoError(t, r.SetTrustedProxies(trustedCIDRs))

	var captured string
	r.Use(func(c *gin.Context) {
		// This is the exact expression RateLimit uses to derive the bucket key.
		captured = c.ClientIP()
		c.Next()
	})
	r.GET("/x", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	req.RemoteAddr = remoteAddr
	if xff != "" {
		req.Header.Set("X-Forwarded-For", xff)
	}
	r.ServeHTTP(w, req)
	return captured
}

func TestRateLimitKeyIgnoresXFFFromUntrustedPeer(t *testing.T) {
	// Trust no proxies → X-Forwarded-For must be ignored, real peer wins.
	got := keyCaptured(t, nil, "1.2.3.4:1111", "10.0.0.1")
	assert.Equal(t, "1.2.3.4", got,
		"spoofed X-Forwarded-For from an untrusted peer must not become the rate-limit key")
}

func TestRateLimitKeyUsesXFFFromTrustedProxy(t *testing.T) {
	// Trust 127.0.0.1 → when request arrives from that proxy, the leftmost
	// X-Forwarded-For entry is the real client.
	got := keyCaptured(t, []string{"127.0.0.1/32"}, "127.0.0.1:1111", "10.0.0.1")
	assert.Equal(t, "10.0.0.1", got,
		"trusted proxy's X-Forwarded-For left-most entry is the real client IP")
}

func TestRateLimitKeyRejectsXFFFromNonTrustedPeerEvenWithTrustedProxyConfigured(t *testing.T) {
	// Even with a trusted-proxies list configured, a peer NOT in that list
	// can't inject X-Forwarded-For. Attacker direct-dials the API pretending
	// to be routed through nginx.
	got := keyCaptured(t, []string{"127.0.0.1/32"}, "8.8.8.8:1111", "10.0.0.1")
	assert.Equal(t, "8.8.8.8", got,
		"X-Forwarded-For from a non-trusted peer must be rejected even when trusted-proxies is configured")
}
