package live

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"google.golang.org/genai"

	"github.com/deemwar/live-api/apps/api/internal/config"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Handler handles the /ws endpoint.
// It upgrades the HTTP connection to WebSocket, starts a Gemini Live session,
// and relays audio bidirectionally.
type Handler struct {
	cfg      *config.Config
	sessions *SessionManager
	genai    *genai.Client
}

// NewHandler returns a wired Handler.
func NewHandler(cfg *config.Config, sessions *SessionManager, client *genai.Client) *Handler {
	return &Handler{cfg: cfg, sessions: sessions, genai: client}
}

// clientMsg is the shape the UI sends over the WebSocket.
// type = "audio" → base64-encoded PCM16 audio chunk (audio/pcm;rate=16000)
// type = "end"   → signals end of client turn
type clientMsg struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"` // base64 raw PCM bytes
}

// serverMsg is what we send back to the UI.
// type = "audio"      → base64-encoded PCM16 audio from Gemini
// type = "transcript" → output transcription text
// type = "error"      → error string
// type = "turn_end"   → Gemini turn complete
type serverMsg struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
}

// ServeHTTP is the Gin handler for GET /ws.
func (h *Handler) ServeHTTP(c *gin.Context) {
	if h.cfg.GeminiAPIKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Gemini API key not configured"})
		return
	}

	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[live] websocket upgrade failed: %v", err)
		return
	}

	sessionID := c.Query("session_id")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	geminiSess, err := h.genai.Live.Connect(ctx, h.cfg.GeminiModel, &genai.LiveConnectConfig{
		ResponseModalities: []genai.Modality{genai.ModalityAudio},
	})
	if err != nil {
		writeErr(wsConn, fmt.Sprintf("failed to connect to Gemini: %v", err))
		wsConn.Close()
		return
	}

	sess := &LiveSession{
		ID:         sessionID,
		ClientConn: wsConn,
		GeminiSess: geminiSess,
		StartedAt:  time.Now(),
	}

	if err := h.sessions.Add(sess); err != nil {
		writeErr(wsConn, err.Error())
		_ = geminiSess.Close()
		wsConn.Close()
		return
	}
	defer h.sessions.Remove(sessionID)

	// client → Gemini goroutine
	clientDone := make(chan struct{})
	go func() {
		defer close(clientDone)
		h.relayClientToGemini(geminiSess, wsConn)
	}()

	// Gemini → client goroutine
	geminiDone := make(chan struct{})
	go func() {
		defer close(geminiDone)
		h.relayGeminiToClient(geminiSess, wsConn, sessionID)
	}()

	// block until either side closes
	select {
	case <-clientDone:
	case <-geminiDone:
	case <-ctx.Done():
	}
}

// relayClientToGemini reads audio chunks from the client WebSocket
// and forwards them to Gemini as realtime audio input.
func (h *Handler) relayClientToGemini(sess *genai.Session, wsConn *websocket.Conn) {
	for {
		_, raw, err := wsConn.ReadMessage()
		if err != nil {
			return
		}
		var msg clientMsg
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		switch msg.Type {
		case "audio":
			if err := sess.SendRealtimeInput(genai.LiveRealtimeInput{
				Audio: &genai.Blob{
					MIMEType: "audio/pcm;rate=16000",
					Data:     []byte(msg.Data), // raw PCM bytes (not base64 — see note in UI)
				},
			}); err != nil {
				log.Printf("[live] send audio to gemini failed: %v", err)
				return
			}
		case "end":
			// client signals turn end — nothing to do, Gemini detects activity automatically
		}
	}
}

// relayGeminiToClient reads server messages from Gemini and forwards
// audio chunks and transcripts back to the client WebSocket.
func (h *Handler) relayGeminiToClient(sess *genai.Session, wsConn *websocket.Conn, sessionID string) {
	for {
		msg, err := sess.Receive()
		if err != nil {
			log.Printf("[live] receive from gemini failed (session %s): %v", sessionID, err)
			return
		}

		if msg.ServerContent != nil {
			sc := msg.ServerContent
			if sc.ModelTurn != nil {
				for _, part := range sc.ModelTurn.Parts {
					if part.InlineData != nil && part.InlineData.Data != nil {
						sendJSON(wsConn, serverMsg{Type: "audio", Data: string(part.InlineData.Data)})
					}
				}
			}
			if sc.OutputTranscription != nil && sc.OutputTranscription.Text != "" {
				sendJSON(wsConn, serverMsg{Type: "transcript", Data: sc.OutputTranscription.Text})
			}
			if sc.TurnComplete {
				sendJSON(wsConn, serverMsg{Type: "turn_end"})
			}
		}

		if msg.ToolCall != nil {
			// RAG tool calls — placeholder for SPEC-004 extension
			log.Printf("[live] tool call received (session %s): %+v", sessionID, msg.ToolCall)
		}
	}
}

func sendJSON(conn *websocket.Conn, v any) {
	b, _ := json.Marshal(v)
	_ = conn.WriteMessage(websocket.TextMessage, b)
}

func writeErr(conn *websocket.Conn, msg string) {
	sendJSON(conn, serverMsg{Type: "error", Data: msg})
}
