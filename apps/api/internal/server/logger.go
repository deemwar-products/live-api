package server

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// logFor returns a per-request logger if one has been attached, else
// a no-op so callers can log unconditionally.
func logFor(c *gin.Context) *slog.Logger {
	if v, ok := c.Get("log"); ok {
		if l, ok := v.(*slog.Logger); ok {
			return l
		}
	}
	return slog.Default()
}
