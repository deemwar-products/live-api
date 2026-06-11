package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Config holds all runtime configuration for the worker.
type Config struct {
	DBPath string
	RedisAddr string // host:port
	RedisPassword string
	RedisDB int
	ConsumerGroup string
	StreamName string
	ConsumerName string
	MaxAttempts int
	GeminiAPIKey string
	EmbeddingModel string
	EmbeddingDim int
	TargetChildTokens int
	OverlapTokens int
	// IdleBlockMS is how long XREADGROUP blocks when no messages are available.
	// Lower = more responsive shutdown; higher = less Redis round-trips when idle.
	IdleBlockMS int
}

// Defaults returns a Config populated with sensible defaults.
func Defaults() Config {
	return Config{
		DBPath: "data/rag.db",
		RedisAddr: "localhost:6379",
		RedisPassword: "",
		RedisDB: 0,
		ConsumerGroup: "workers",
		StreamName: "jobs.stream",
		ConsumerName: "", // set in Load() if empty
		MaxAttempts: 3,
		GeminiAPIKey: "",
		EmbeddingModel: "text-embedding-004",
		EmbeddingDim: 768,
		TargetChildTokens: 150,
		OverlapTokens: 30,
		IdleBlockMS: 5000,
	}
}

// Load reads configuration from environment variables.
// Returns an error if a required value is missing or invalid.
func Load() (Config, error) {
	c := Defaults()

	if v := os.Getenv("DB_PATH"); v != "" {
		c.DBPath = v
	}
	if v := os.Getenv("REDIS_HOST"); v != "" {
		port := os.Getenv("REDIS_PORT")
		if port == "" {
			port = "6379"
		}
		c.RedisAddr = v + ":" + port
	}

	if v := os.Getenv("REDIS_ADDR"); v != "" {
		c.RedisAddr = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		c.RedisPassword = v
	}
	if v := os.Getenv("REDIS_DB"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return c, fmt.Errorf("REDIS_DB: %w", err)
		}
		c.RedisDB = n
	}
	if v := os.Getenv("CONSUMER_GROUP"); v != "" {
		c.ConsumerGroup = v
	}
	if v := os.Getenv("STREAM_NAME"); v != "" {
		c.StreamName = v
	}
	if v := os.Getenv("CONSUMER_NAME"); v != "" {
		c.ConsumerName = v
	}
	if v := os.Getenv("MAX_ATTEMPTS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return c, fmt.Errorf("MAX_ATTEMPTS: %w", err)
		}
		if n < 1 {
			return c, fmt.Errorf("MAX_ATTEMPTS must be >= 1, got %d", n)
		}
		c.MaxAttempts = n
	}
	if v := os.Getenv("GEMINI_API_KEY"); v != "" {
		c.GeminiAPIKey = v
	} else if v := os.Getenv("GOOGLE_API_KEY"); v != "" {
		c.GeminiAPIKey = v
	}
	if v := os.Getenv("EMBEDDING_MODEL"); v != "" {
		c.EmbeddingModel = v
	}
	if v := os.Getenv("EMBEDDING_DIM"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return c, fmt.Errorf("EMBEDDING_DIM: %w", err)
		}
		if n < 1 {
			return c, fmt.Errorf("EMBEDDING_DIM must be >= 1, got %d", n)
		}
		c.EmbeddingDim = n
	}
	if v := os.Getenv("TARGET_CHILD_TOKENS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return c, fmt.Errorf("TARGET_CHILD_TOKENS: %w", err)
		}
		if n < 1 {
			return c, fmt.Errorf("TARGET_CHILD_TOKENS must be >= 1, got %d", n)
		}
		c.TargetChildTokens = n
	}
	if v := os.Getenv("OVERLAP_TOKENS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return c, fmt.Errorf("OVERLAP_TOKENS: %w", err)
		}
		if n < 0 {
			return c, fmt.Errorf("OVERLAP_TOKENS must be >= 0, got %d", n)
		}
		c.OverlapTokens = n
	}
	if v := os.Getenv("IDLE_BLOCK_MS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return c, fmt.Errorf("IDLE_BLOCK_MS: %w", err)
		}
		if n < 0 {
			return c, fmt.Errorf("IDLE_BLOCK_MS must be >= 0, got %d", n)
		}
		c.IdleBlockMS = n
	}

	if c.ConsumerName == "" {
		c.ConsumerName = DefaultConsumerName(os.Hostname, os.Getpid)
	}

	return c, nil
}

// HostnameFunc returns the OS hostname, or an error.
type HostnameFunc func() (string, error)

// PidFunc returns the OS process ID.
type PidFunc func() int

// DefaultConsumerName returns a consumer name like "worker-<hostname>-<pid>".
// If hostname lookup fails, "unknown" is used. Exported for testing.
func DefaultConsumerName(hostname HostnameFunc, pid PidFunc) string {
	host, err := hostname()
	if err != nil {
		host = "unknown"
	}
	return fmt.Sprintf("worker-%s-%d", host, pid())
}

// EnsureDBDir creates the parent directory for DBPath if it does not exist.
// No-op for S3 paths.
func (c Config) EnsureDBDir() error {
	if isS3Path(c.DBPath) {
		return nil
	}
	dir := filepath.Dir(c.DBPath)
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func isS3Path(p string) bool {
	return len(p) >= 5 && p[:5] == "s3://"
}

// Validate returns an error if the config is missing required values.
func (c Config) Validate() error {
	if c.DBPath == "" {
		return errors.New("DB_PATH is required")
	}
	if c.RedisAddr == "" {
		return errors.New("REDIS_HOST or REDIS_ADDR is required")
	}
	if c.StreamName == "" {
		return errors.New("STREAM_NAME is required")
	}
	if c.ConsumerGroup == "" {
		return errors.New("CONSUMER_GROUP is required")
	}
	if c.ConsumerName == "" {
		return errors.New("CONSUMER_NAME is required")
	}
	return nil
}
