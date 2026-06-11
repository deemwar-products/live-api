package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaults(t *testing.T) {
	c := Defaults()
	if c.DBPath != "data/rag.db" {
		t.Errorf("default DBPath: got %q want %q", c.DBPath, "data/rag.db")
	}
	if c.RedisAddr != "localhost:6379" {
		t.Errorf("default RedisAddr: got %q", c.RedisAddr)
	}
	if c.ConsumerGroup != "workers" {
		t.Errorf("default ConsumerGroup: got %q", c.ConsumerGroup)
	}
	if c.StreamName != "jobs.stream" {
		t.Errorf("default StreamName: got %q", c.StreamName)
	}
	if c.MaxAttempts != 3 {
		t.Errorf("default MaxAttempts: got %d want 3", c.MaxAttempts)
	}
	if c.EmbeddingModel != "text-embedding-004" {
		t.Errorf("default EmbeddingModel: got %q", c.EmbeddingModel)
	}
	if c.EmbeddingDim != 768 {
		t.Errorf("default EmbeddingDim: got %d", c.EmbeddingDim)
	}
	if c.TargetChildTokens != 150 {
		t.Errorf("default TargetChildTokens: got %d", c.TargetChildTokens)
	}
	if c.OverlapTokens != 30 {
		t.Errorf("default OverlapTokens: got %d", c.OverlapTokens)
	}
	if c.IdleBlockMS != 5000 {
		t.Errorf("default IdleBlockMS: got %d", c.IdleBlockMS)
	}
}

func TestLoad_DefaultsWhenUnset(t *testing.T) {
	clearEnv(t)
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.RedisAddr != "localhost:6379" {
		t.Errorf("RedisAddr: got %q", c.RedisAddr)
	}
	if c.ConsumerName == "" {
		t.Error("ConsumerName should default to non-empty")
	}
}

func TestLoad_RedisHostDefaultPort(t *testing.T) {
	t.Setenv("REDIS_HOST", "redis.example")
	// REDIS_PORT unset
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.RedisAddr != "redis.example:6379" {
		t.Errorf("RedisAddr: got %q want redis.example:6379", c.RedisAddr)
	}
}

func TestLoad_AllEnvs(t *testing.T) {
	t.Setenv("DB_PATH", "/var/lib/rag.db")
	t.Setenv("REDIS_HOST", "redis.example")
	t.Setenv("REDIS_PORT", "6380")
	t.Setenv("REDIS_PASSWORD", "secret")
	t.Setenv("REDIS_DB", "2")
	t.Setenv("CONSUMER_GROUP", "g1")
	t.Setenv("STREAM_NAME", "s1")
	t.Setenv("CONSUMER_NAME", "c1")
	t.Setenv("MAX_ATTEMPTS", "5")
	t.Setenv("GEMINI_API_KEY", "gk")
	t.Setenv("EMBEDDING_MODEL", "text-embedding-005")
	t.Setenv("EMBEDDING_DIM", "1536")
	t.Setenv("TARGET_CHILD_TOKENS", "200")
	t.Setenv("OVERLAP_TOKENS", "40")
	t.Setenv("IDLE_BLOCK_MS", "1000")

	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.DBPath != "/var/lib/rag.db" {
		t.Errorf("DBPath: got %q", c.DBPath)
	}
	if c.RedisAddr != "redis.example:6380" {
		t.Errorf("RedisAddr: got %q", c.RedisAddr)
	}
	if c.RedisPassword != "secret" {
		t.Errorf("RedisPassword: got %q", c.RedisPassword)
	}
	if c.RedisDB != 2 {
		t.Errorf("RedisDB: got %d", c.RedisDB)
	}
	if c.ConsumerGroup != "g1" {
		t.Errorf("ConsumerGroup: got %q", c.ConsumerGroup)
	}
	if c.StreamName != "s1" {
		t.Errorf("StreamName: got %q", c.StreamName)
	}
	if c.ConsumerName != "c1" {
		t.Errorf("ConsumerName: got %q", c.ConsumerName)
	}
	if c.MaxAttempts != 5 {
		t.Errorf("MaxAttempts: got %d", c.MaxAttempts)
	}
	if c.GeminiAPIKey != "gk" {
		t.Errorf("GeminiAPIKey: got %q", c.GeminiAPIKey)
	}
	if c.EmbeddingModel != "text-embedding-005" {
		t.Errorf("EmbeddingModel: got %q", c.EmbeddingModel)
	}
	if c.EmbeddingDim != 1536 {
		t.Errorf("EmbeddingDim: got %d", c.EmbeddingDim)
	}
	if c.TargetChildTokens != 200 {
		t.Errorf("TargetChildTokens: got %d", c.TargetChildTokens)
	}
	if c.OverlapTokens != 40 {
		t.Errorf("OverlapTokens: got %d", c.OverlapTokens)
	}
	if c.IdleBlockMS != 1000 {
		t.Errorf("IdleBlockMS: got %d", c.IdleBlockMS)
	}
}

func TestLoad_GeminiAlias(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "via-google")
	// GEMINI_API_KEY is unset
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.GeminiAPIKey != "via-google" {
		t.Errorf("GeminiAPIKey: got %q want %q", c.GeminiAPIKey, "via-google")
	}
}

func TestLoad_GeminiKeyTakesPrecedence(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "primary")
	t.Setenv("GOOGLE_API_KEY", "alias")
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.GeminiAPIKey != "primary" {
		t.Errorf("GeminiAPIKey: got %q want primary", c.GeminiAPIKey)
	}
}

func TestLoad_REDIS_ADDROverridesHostPort(t *testing.T) {
	t.Setenv("REDIS_HOST", "host1")
	t.Setenv("REDIS_PORT", "1111")
	t.Setenv("REDIS_ADDR", "host2:2222")
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.RedisAddr != "host2:2222" {
		t.Errorf("RedisAddr: got %q want host2:2222", c.RedisAddr)
	}
}

func TestLoad_InvalidInts(t *testing.T) {
	cases := []string{
		"REDIS_DB", "MAX_ATTEMPTS", "EMBEDDING_DIM",
		"TARGET_CHILD_TOKENS", "OVERLAP_TOKENS", "IDLE_BLOCK_MS",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			t.Setenv(name, "not-a-number")
			_, err := Load()
			if err == nil {
				t.Errorf("expected error for %s", name)
			}
		})
	}
}

func TestLoad_RangeChecks(t *testing.T) {
	cases := []struct{ name, val string }{
		{"MAX_ATTEMPTS", "0"},
		{"EMBEDDING_DIM", "-1"},
		{"TARGET_CHILD_TOKENS", "0"},
		{"OVERLAP_TOKENS", "-1"},
		{"IDLE_BLOCK_MS", "-1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(tc.name, tc.val)
			_, err := Load()
			if err == nil {
				t.Errorf("expected error for %s=%s", tc.name, tc.val)
			}
		})
	}
}

func TestLoad_EnvError(t *testing.T) {
	t.Run("invalid REDIS_DB", func(t *testing.T) {
		t.Setenv("REDIS_DB", "abc")
		_, err := Load()
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "REDIS_DB") {
			t.Errorf("wrong error: %v", err)
		}
	})
}

func TestLoad_ConsumerNameDefaults(t *testing.T) {
	clearEnv(t)
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !strings.HasPrefix(c.ConsumerName, "worker-") {
		t.Errorf("ConsumerName: got %q want worker- prefix", c.ConsumerName)
	}
}

func TestDefaultConsumerName(t *testing.T) {
	goodHost := func() (string, error) { return "box1", nil }
	badHost := func() (string, error) { return "", errors.New("nope") }
	pid := func() int { return 4242 }

	if got := DefaultConsumerName(goodHost, pid); got != "worker-box1-4242" {
		t.Errorf("good: got %q", got)
	}
	if got := DefaultConsumerName(badHost, pid); got != "worker-unknown-4242" {
		t.Errorf("bad: got %q", got)
	}
}

func TestEnsureDBDir_Local(t *testing.T) {
	dir := t.TempDir()
	c := Config{DBPath: filepath.Join(dir, "sub", "rag.db")}
	if err := c.EnsureDBDir(); err != nil {
		t.Fatalf("EnsureDBDir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "sub")); err != nil {
		t.Errorf("expected dir created: %v", err)
	}
}

func TestEnsureDBDir_S3Skips(t *testing.T) {
	c := Config{DBPath: "s3://bucket/rag.db"}
	if err := c.EnsureDBDir(); err != nil {
		t.Errorf("S3 EnsureDBDir should be no-op, got %v", err)
	}
}

func TestEnsureDBDir_EmptyDir(t *testing.T) {
	c := Config{DBPath: "rag.db"} // no dir component
	if err := c.EnsureDBDir(); err != nil {
		t.Errorf("empty dir EnsureDBDir: %v", err)
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name string
		mutate func(*Config)
		wantErr bool
	}{
		{"valid", func(c *Config) { c.DBPath = "x"; c.RedisAddr = "h:p"; c.StreamName = "s"; c.ConsumerGroup = "g"; c.ConsumerName = "cn" }, false},
		{"missing DBPath", func(c *Config) { c.DBPath = "" }, true},
		{"missing RedisAddr", func(c *Config) { c.RedisAddr = "" }, true},
		{"missing StreamName", func(c *Config) { c.StreamName = "" }, true},
		{"missing ConsumerGroup", func(c *Config) { c.ConsumerGroup = "" }, true},
		{"missing ConsumerName", func(c *Config) { c.ConsumerName = "" }, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := Defaults()
			tc.mutate(&c)
			err := c.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"DB_PATH", "REDIS_HOST", "REDIS_PORT", "REDIS_ADDR", "REDIS_PASSWORD",
		"REDIS_DB", "CONSUMER_GROUP", "STREAM_NAME", "CONSUMER_NAME",
		"MAX_ATTEMPTS", "GEMINI_API_KEY", "GOOGLE_API_KEY", "EMBEDDING_MODEL",
		"EMBEDDING_DIM", "TARGET_CHILD_TOKENS", "OVERLAP_TOKENS", "IDLE_BLOCK_MS",
	} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}
}
