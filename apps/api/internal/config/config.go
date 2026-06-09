package config

import "os"

type Config struct {
	Port      string
	RedisHost string
	RedisPort string
	DBPath    string
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6379"),
		DBPath:    getEnv("DB_PATH", "data/rag.db"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
