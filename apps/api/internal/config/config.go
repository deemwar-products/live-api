// Package config loads runtime configuration from environment variables.
//
// All values are read once at startup. The server fails fast if a required
// value (notably GEMINI_API_KEY) is missing — see Load.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config is the resolved runtime configuration for the API server.
type Config struct {
	// Port the HTTP/WS server binds to. Defaults to "8080".
	Port string

	// GeminiAPIKey authenticates with the Gemini Live API. Required.
	GeminiAPIKey string

	// GeminiModel is the live model identifier passed to Live.Connect.
	// Defaults to "gemini-2.0-flash-live-001" when unset.
	GeminiModel string

	// AllowedOrigin is the browser origin permitted to open the WS.
	// Defaults to "http://localhost:5173" (Vite dev server).
	AllowedOrigin string

	// SessionMaxSeconds caps each WS session duration. Default 600s (10 min).
	SessionMaxSeconds time.Duration

	// PingInterval is how often the server pings the client on idle. Default 15s.
	PingInterval time.Duration

	// PongTimeout is how long the server waits for a pong reply. Default 5s.
	PongTimeout time.Duration
}

// Load reads environment variables and returns a fully resolved Config.
// It returns an error if GEMINI_API_KEY is empty — the server cannot
// function without it and we'd rather fail at boot than on first request.
func Load() (Config, error) {
	cfg := Config{
		Port: getenv("PORT", "8080"),
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		GeminiModel: getenv("GEMINI_MODEL", "gemini-2.0-flash-live-001"),
		AllowedOrigin: getenv("ALLOWED_ORIGIN", "http://localhost:5173"),
		SessionMaxSeconds: time.Duration(getenvInt("SESSION_MAX_SECONDS", 600)) * time.Second,
		PingInterval: time.Duration(getenvInt("WS_PING_INTERVAL_SECONDS", 15)) * time.Second,
		PongTimeout: time.Duration(getenvInt("WS_PONG_TIMEOUT_SECONDS", 5)) * time.Second,
	}

	if cfg.GeminiAPIKey == "" {
		return Config{}, errors.New("GEMINI_API_KEY is required; set it in apps/api/.env.local")
	}
	if cfg.SessionMaxSeconds <= 0 {
		return Config{}, fmt.Errorf("SESSION_MAX_SECONDS must be > 0, got %d", int(cfg.SessionMaxSeconds.Seconds()))
	}
	if cfg.PingInterval <= 0 || cfg.PongTimeout <= 0 {
		return Config{}, errors.New("WS_PING_INTERVAL_SECONDS and WS_PONG_TIMEOUT_SECONDS must be > 0")
	}
	if cfg.PongTimeout >= cfg.PingInterval {
		return Config{}, fmt.Errorf("WS_PONG_TIMEOUT_SECONDS (%s) must be < WS_PING_INTERVAL_SECONDS (%s)", cfg.PongTimeout, cfg.PingInterval)
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
