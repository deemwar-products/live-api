// Package logger wraps log/slog with a small redaction helper.
//
// Redact replaces the value of any key matching a sensitive name with
// "[redacted]" (preserving first/last 4 chars when long enough) so we
// can safely log structs that contain API keys, tokens, or auth
// headers.
package logger

import (
	"log/slog"
	"os"
	"strings"
)

const redactedPlaceholder = "[redacted]"

// sensitiveKeySubstrings — case-insensitive substring matches against
// field names. If a key contains any of these, its value is redacted.
var sensitiveKeySubstrings = []string{
	"apikey",
	"api_key",
	"authorization",
	"access_token",
	"refreshtoken",
	"refresh_token",
	"password",
	"secret",
}

// New returns a JSON slog.Logger writing to stderr. Level is read from
// the LOG_LEVEL env var (one of: debug, info, warn, error). Default
// is info. Suitable for containerized/local dev.
func New() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: parseLogLevel(os.Getenv("LOG_LEVEL")),
	}))
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// RedactValue returns a safe string to log for a value attached to a
// potentially-sensitive key. The key is matched case-insensitively
// against the sensitiveKeySubstrings list.
func RedactValue(key string, value any) any {
	if !isSensitive(key) {
		return value
	}
	s, ok := value.(string)
	if !ok {
		return redactedPlaceholder
	}
	if len(s) <= 8 {
		return redactedPlaceholder
	}
	return s[:4] + "…" + s[len(s)-4:]
}

// isSensitive reports whether the given key name (e.g. "GEMINI_API_KEY"
// or "Authorization") should be treated as sensitive.
func isSensitive(key string) bool {
	k := strings.ToLower(strings.ReplaceAll(key, "-", "_"))
	for _, s := range sensitiveKeySubstrings {
		if strings.Contains(k, s) {
			return true
		}
	}
	return false
}
