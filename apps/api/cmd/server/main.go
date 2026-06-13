// Command server starts the Live API backend.
//
// Configuration is read from environment variables (see internal/config).
// The server fails fast at boot if GEMINI_API_KEY is missing or the
// resolved configuration is invalid.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api/internal/config"
	apilog "api/internal/logger"
	"api/internal/server"
)

func main() {
	log := apilog.New()

	cfg, err := config.Load()
	if err != nil {
		log.Error("config load failed", "error", err.Error())
		os.Exit(1)
	}

	log.Info("starting live-api-poc backend",
		"port", cfg.Port,
		"model", cfg.GeminiModel,
		"allowed_origin", cfg.AllowedOrigin,
		"session_max_seconds", int(cfg.SessionMaxSeconds.Seconds()),
		"gemini_api_key", apilog.RedactValue("GEMINI_API_KEY", cfg.GeminiAPIKey),
	)

	engine := server.New(cfg, log)

	// Graceful shutdown: catch SIGINT/SIGTERM and give in-flight WS
	// sessions a moment to close cleanly.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Info("shutdown signal received")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = ctx
	}()

	addr := fmt.Sprintf(":%s", cfg.Port)
	if err := engine.Run(addr); err != nil {
		log.Error("server exited", "error", err.Error())
		os.Exit(1)
	}
}
