package config

import (
	"os"
	"sync"

	"github.com/deemwar/live-api/apps/api/internal/logger"
)

var (
	config *Config
	once  sync.Once
	log = logger.New("config")
)

type Config struct {
	Port      string
	RedisHost string
	RedisPort string
	DBPath    string
}

func Load() *Config {
	once.Do(func() {
		config = &Config{
			Port:      getEnv("PORT", "8080"),
			RedisHost: getEnv("REDIS_HOST", "localhost"),
			RedisPort: getEnv("REDIS_PORT", "6379"),
			DBPath:    getEnv("DB_PATH", "data/rag.db"),
		}
	})
	return config
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)

	if !exists {
		log.Warn("Environment variable %s not set, using default value: %s", key, fallback)
		return fallback
	}

	if value == "" {
		log.Warn("Environment variable %s is empty, using default value: %s", key, fallback)
		return fallback
	}

	return value

}

func (c *Config) LogConfiguration() {
	log.Info("EnvironmentConfiguration:\n \tPort: %s\n \tRedisHost: %s\n \tRedisPort: %s\n \tDBPath: %s", c.Port, c.RedisHost, c.RedisPort, c.DBPath)
}