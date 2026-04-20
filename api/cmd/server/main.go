// Command maidenmap-server serves the MaidenMap JSON API.
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wh0am1i/maidenmap/api/internal/data"
	"github.com/wh0am1i/maidenmap/api/internal/handler"
	"github.com/wh0am1i/maidenmap/api/internal/middleware"
)

func main() {
	dataDir := flag.String("data-dir", envDefault("DATA_DIR", "./data"), "directory containing countries.geojson, cities.bin, admin_codes.json")
	addr := flag.String("addr", envDefault("LISTEN_ADDR", ":8080"), "HTTP listen address")
	rpm := flag.Int("rate-limit", envIntDefault("RATE_LIMIT_PER_MIN", 60), "per-IP requests per minute")
	trustedProxiesRaw := flag.String("trusted-proxies", envDefault("TRUSTED_PROXIES", "127.0.0.1"), "comma-separated trusted proxy CIDRs")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	slog.Info("loading data", "dir", *dataDir)
	ds, err := data.Load(*dataDir)
	if err != nil {
		slog.Error("data load failed", "err", err)
		os.Exit(1)
	}
	slog.Info("data loaded", "cities", len(ds.Cities), "countries", len(ds.Countries.Features))

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	if err := r.SetTrustedProxies(strings.Split(*trustedProxiesRaw, ",")); err != nil {
		slog.Error("SetTrustedProxies failed", "err", err)
		os.Exit(1)
	}

	api := r.Group("/api")
	api.Use(middleware.RateLimit(*rpm))
	api.GET("/health", handler.Health(ds))
	api.GET("/grid/:code", handler.GridSingle(ds))
	api.GET("/grid", handler.GridBatch(ds))

	// Explicit timeouts protect against slowloris and slow-client DoS.
	// Defaults are zero (infinite) which is unsafe on the public internet.
	srv := &http.Server{
		Addr:              *addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MiB
	}

	// Graceful shutdown on SIGINT / SIGTERM so in-flight requests complete.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("listening", "addr", *addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	<-shutdown
	slog.Info("shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}
	slog.Info("bye")
}

func envDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envIntDefault(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
