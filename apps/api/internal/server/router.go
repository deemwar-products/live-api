// Package server wires the Gin engine, CORS, request logging, and
// mounts the live WS handler.
package server

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"api/internal/config"
	"api/internal/live"
)

// New builds and returns a configured *gin.Engine.
func New(cfg config.Config, log *slog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger(log))
	r.Use(cors(cfg.AllowedOrigin))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	handler := live.NewHandler(cfg, log)
	r.GET("/v1/live", handler.Handle)

	return r
}

// requestLogger is a tiny structured access log.
func requestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		log.Info("http",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
		)
	}
}

// cors is a minimal CORS middleware tailored for the live WS endpoint.
// Browsers open WS upgrades with a normal HTTP preflight, so the same
// Access-Control-* headers cover both cases.
func cors(allowedOrigin string) gin.HandlerFunc {
	allowed := strings.TrimSpace(allowedOrigin)
	return func(c *gin.Context) {
		if allowed != "" {
			c.Header("Access-Control-Allow-Origin", allowed)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Header("Access-Control-Max-Age", "86400")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		// Reject WS upgrades from a non-allowed origin. The WS library
		// runs CheckOrigin itself, but doing it here gives a clean 403
		// for bad origins and a log line.
		if c.Request.URL.Path == "/v1/live" && allowed != "" {
			origin := c.GetHeader("Origin")
			if origin != "" && origin != allowed {
				logFor(c).Warn("rejected ws upgrade from disallowed origin", "origin", origin)
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}
		c.Next()
	}
}
