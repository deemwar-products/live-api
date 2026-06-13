package config

import (
 "os"
 "testing"
)

func TestConfig_WhenLoadCalled_ThenReturnsConfigStruct(t *testing.T) {
 Reset()
 cfg := Load()
 if cfg == nil {
 t.Fatal("Load returned nil")
 }
 if cfg.Port == "" {
 t.Error("Port not set")
 }
 if cfg.RedisHost == "" {
 t.Error("RedisHost not set")
 }
 if cfg.RedisPort == "" {
 t.Error("RedisPort not set")
 }
 if cfg.DBPath == "" {
 t.Error("DBPath not set")
 }
}

func TestConfig_WhenLoadCalledMultipleTimes_ThenReturnsSameInstance(t *testing.T) {
 Reset()
 cfg1 := Load()
 cfg2 := Load()
 if cfg1 != cfg2 {
 t.Error("Load should return same instance")
 }
}

func TestGetEnv_WhenEnvVarNotSet_ThenReturnsDefault(t *testing.T) {
 os.Unsetenv("TEST_VAR_UNSET")
 result := getEnv("TEST_VAR_UNSET", "default")
 if result != "default" {
 t.Errorf("Expected default, got %s", result)
 }
}

func TestGetEnv_WhenEnvVarEmpty_ThenReturnsDefault(t *testing.T) {
 os.Setenv("TEST_VAR_EMPTY", "")
 defer os.Unsetenv("TEST_VAR_EMPTY")
 result := getEnv("TEST_VAR_EMPTY", "default")
 if result != "default" {
 t.Errorf("Expected default, got %s", result)
 }
}

func TestGetEnv_WhenEnvVarSet_ThenReturnsValue(t *testing.T) {
 os.Setenv("TEST_VAR_SET", "value123")
 defer os.Unsetenv("TEST_VAR_SET")
 result := getEnv("TEST_VAR_SET", "default")
 if result != "value123" {
 t.Errorf("Expected value123, got %s", result)
 }
}

func TestConfig_WhenLogConfigurationCalled_ThenLogsConfig(t *testing.T) {
 Reset()
 cfg := Load()
 cfg.LogConfiguration()
}

func TestConfig_WhenLogConfigurationCalled_ThenLogsCustomConfig(t *testing.T) {
 Reset()
 cfg := Load()
 cfg.Port = "9999"
 cfg.LogConfiguration()
}

func TestConfig_WhenLoadWithEnvVarsSet_ThenUsesThem(t *testing.T) {
 t.Setenv("PORT", "1234")
 t.Setenv("REDIS_HOST", "redis.example.com")
 t.Setenv("REDIS_PORT", "6380")
 t.Setenv("DB_PATH", "/tmp/custom.db")
 Reset()

 cfg := Load()
 if cfg.Port != "1234" {
 t.Errorf("expected Port=1234, got %s", cfg.Port)
 }
 if cfg.RedisHost != "redis.example.com" {
 t.Errorf("expected RedisHost=redis.example.com, got %s", cfg.RedisHost)
 }
 if cfg.RedisPort != "6380" {
 t.Errorf("expected RedisPort=6380, got %s", cfg.RedisPort)
 }
 if cfg.DBPath != "/tmp/custom.db" {
 t.Errorf("expected DBPath=/tmp/custom.db, got %s", cfg.DBPath)
 }
}
