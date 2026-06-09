package config

import "testing"

func TestLoadUsesFallbacks(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("REDIS_HOST", "")
	t.Setenv("REDIS_PORT", "")
	t.Setenv("DB_PATH", "")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Fatalf("expected default port, got %q", cfg.Port)
	}
	if cfg.RedisHost != "localhost" {
		t.Fatalf("expected default redis host, got %q", cfg.RedisHost)
	}
	if cfg.RedisPort != "6379" {
		t.Fatalf("expected default redis port, got %q", cfg.RedisPort)
	}
	if cfg.DBPath != "data/rag.db" {
		t.Fatalf("expected default db path, got %q", cfg.DBPath)
	}
}

func TestLoadUsesEnvVars(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("REDIS_HOST", "redis")
	t.Setenv("REDIS_PORT", "6380")
	t.Setenv("DB_PATH", "custom.db")

	cfg := Load()

	if cfg.Port != "9090" || cfg.RedisHost != "redis" || cfg.RedisPort != "6380" || cfg.DBPath != "custom.db" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}
