package config

import (
	"os"
	"strconv"
	"sync"

	"github.com/deemwar/live-api/apps/api/internal/logger"
)

var (
	config *Config
	once sync.Once
	log  = logger.New("config")
)

type Config struct {
	Port string
	RedisHost string
	RedisPort string
	RedisPassword string
	RedisDB int
	DBPath string
	MigrationsPath string
	GeminiAPIKey string
	GeminiModel string
	MaxSessions int
	DocumentsDir string
	UploadMaxBytes int
}

func Load() *Config {
	once.Do(func() {
		config = &Config{
			Port: getEnv("PORT", "8080"),
			RedisHost: getEnv("REDIS_HOST", "localhost"),
			RedisPort: getEnv("REDIS_PORT", "6379"),
			RedisPassword: getEnv("REDIS_PASSWORD", ""),
			RedisDB: getEnvInt("REDIS_DB", 0),
			DBPath: getEnv("DB_PATH", "../data/data.db"),
			MigrationsPath: getEnv("MIGRATIONS_PATH", "../migrations"),
			GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
			GeminiModel: getEnv("GEMINI_MODEL", "gemini-2.0-flash-live-001"),
			MaxSessions: 10,
			DocumentsDir: getEnv("DOCUMENTS_DIR", "../data/documents"),
			UploadMaxBytes: 10 * 1024 * 1024,
		}
	})
	return config
}

// Reset clears the singleton state. Test-only.
func Reset() {
	config = nil
	once = sync.Once{}
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

func getEnvInt(key string, fallback int) int {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Warn("Environment variable %s is not a valid int (%q), using default %d", key, value, fallback)
		return fallback
	}
	return parsed
}

func (c *Config) LogConfiguration() {
	log.Info("EnvironmentConfiguration:\n \tPort: %s\n \tRedisHost: %s\n \tRedisPort: %s\n \tDBPath: %s\n \tMigrationsPath: %s\n \tGeminiModel: %s\n \tDocumentsDir: %s", c.Port, c.RedisHost, c.RedisPort, c.DBPath, c.MigrationsPath, c.GeminiModel, c.DocumentsDir)
}
