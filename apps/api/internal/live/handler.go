package live

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"api/internal/config"
)

// Handler is the Gin handler that upgrades an HTTP request to a
// WebSocket and runs a Session. It is intentionally thin — the heavy
// lifting lives in Session.
type Handler struct {
	Cfg config.Config
	Log *slog.Logger

	// Upgrader is the gorilla/websocket upgrader. CheckOrigin is
	// enforced at the router via middleware; the upgrader itself is
	// permissive because we never reach this point for a bad origin.
	Upgrader websocket.Upgrader
}

// NewHandler builds a Handler with safe defaults.
func NewHandler(cfg config.Config, log *slog.Logger) *Handler {
	return &Handler{
		Cfg: cfg,
		Log: log,
		Upgrader: websocket.Upgrader{
			ReadBufferSize: 16 << 10,
			WriteBufferSize: 16 << 10,
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// Handle is the Gin entry point. It upgrades, builds a Session, and
// runs it to completion. We deliberately use a fresh background
// context (not Gin's request context) so the session isn't cancelled
// when the handler returns.
func (h *Handler) Handle(c *gin.Context) {
	ws, err := h.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.Log.Warn("ws upgrade failed", "error", err.Error())
		return
	}

	sess, err := NewSession(context.Background(), h.Cfg, h.Log, ws, h.Cfg.GeminiModel)
	if err != nil {
		h.Log.Error("session open failed", "error", err.Error())
		_ = ws.WriteJSON(Envelope{
			V: ProtocolVersion,
			Type: TypeError,
			ID: "boot-" + time.Now().Format(time.RFC3339Nano),
			TS: time.Now().UnixMilli(),
			Payload: mustJSON(ErrorPayload{
				Code: CodeGeminiUnavailable,
				Message: "failed to open Gemini session",
				Fatal: true,
				Cause: err.Error(),
			}),
		})
		_ = ws.Close()
		return
	}

	_ = sess.Run(context.Background())
}

func mustJSON(v any) []byte {
	b, _ := jsonMarshal(v)
	return b
}
