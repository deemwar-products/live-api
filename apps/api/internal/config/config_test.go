package config

import (
 "os"
 "sync"
 "testing"
)

func reset(t *testing.T) {
 t.Helper()
 config = nil
 once = sync.Once{}
}

func TestConfig_WhenLoadCalled_ThenReturnsConfigStruct(t *testing.T) {
 reset(t)
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
 reset(t)
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
 reset(t)
 cfg := Load()
 cfg.LogConfiguration()
}

func TestConfig_WhenLogConfigurationCalled_ThenLogsCustomConfig(t *testing.T) {
 reset(t)
 cfg := Load()
 cfg.Port = "9999"
 cfg.LogConfiguration()
}