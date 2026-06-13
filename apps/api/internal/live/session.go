package live

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"google.golang.org/genai"

	"api/internal/config"
)

// Session is a single end-to-end WebSocket ↔ Gemini Live bridge.
//
// Lifecycle:
// - Construct via NewSession.
// - Run(ctx) returns when either side closes, the timeout fires, or
// an unrecoverable error occurs.
// - Cancel the context to force a clean teardown.
type Session struct {
	id string
	cfg config.Config
	log *slog.Logger

	// liveAt is the time the session actually went Live (used for elapsedMs).
	liveAt time.Time

	ws *websocket.Conn
	gemini *genai.Session
	model string

	// writeMu serialises writes to the browser WS. Gorilla requires
	// exactly one writer at a time.
	writeMu sync.Mutex

	// audioInFrames counts audio_in frames received from the browser.
	// Used to throttle debug logging of the (otherwise ~50/s) audio stream,
	// and watched by audioInWatchdog to confirm the browser is sending
	// audio at all.
	audioInFrames atomic.Uint64
}

// NewSession opens the upstream Gemini Live connection and returns a
// Session ready to be Run against a browser WS. The provided context
// is only used to establish the upstream — use Session.Run with a
// fresh context for the actual session lifetime.
func NewSession(ctx context.Context, cfg config.Config, log *slog.Logger, ws *websocket.Conn, model string) (*Session, error) {
	log.Debug("creating genai client", "model", model)
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: cfg.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
		HTTPOptions: genai.HTTPOptions{APIVersion: "v1alpha"},
	})
	if err != nil {
		log.Error("create genai client failed", "error", err.Error())
		return nil, fmt.Errorf("create genai client: %w", err)
	}

	connectCfg := &genai.LiveConnectConfig{
		ResponseModalities: []genai.Modality{genai.ModalityAudio},
		InputAudioTranscription: &genai.AudioTranscriptionConfig{},
		OutputAudioTranscription: &genai.AudioTranscriptionConfig{},
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: defaultSystemPrompt}},
		},
	}

	log.Debug("connecting to gemini live", "model", model)
	gSession, err := client.Live.Connect(ctx, model, connectCfg)
	if err != nil {
		log.Error("connect gemini live failed", "model", model, "error", err.Error())
		return nil, fmt.Errorf("connect gemini live: %w", err)
	}
	log.Debug("gemini live connected", "model", model)

	id := "sess_" + uuid.NewString()
	return &Session{
		id: id,
		cfg: cfg,
		log: log.With("session_id", id),
		ws: ws,
		gemini: gSession,
		model: model,
	}, nil
}

// SessionID returns the id assigned at construction.
func (s *Session) SessionID() string { return s.id }

// Run drives the session: it sends the initial Ready frame, then spawns
// pumpIn (browser→Gemini) and pumpOut (Gemini→browser). It returns
// when either side closes or an unrecoverable error occurs. The session
// is capped at cfg.SessionMaxSeconds via the derived context.
func (s *Session) Run(parent context.Context) error {
	ctx, cancel := context.WithTimeout(parent, s.cfg.SessionMaxSeconds)
	defer cancel()

	// 1. Send Ready.
	if err := s.sendReady(); err != nil {
		s.sendFatal(CodeInternal, "failed to send ready", err)
		return err
	}

	// 2. Heartbeats: server-initiated pings with pong timeout enforcement.
	pingTicker := time.NewTicker(s.cfg.PingInterval)
	defer pingTicker.Stop()

	// 3. Goroutines: in, out, heartbeat watcher, audio-in watchdog.
	errCh := make(chan error, 3)
	go func() { errCh <- s.pumpIn(ctx) }()
	go func() { errCh <- s.pumpOut(ctx) }()
	go func() { errCh <- s.heartbeatLoop(ctx, pingTicker) }()
	go s.audioInWatchdog(ctx)

	// 4. Wait for first error (or timeout) and tear down.
	err := <-errCh
	cancel() // stop the other goroutines

	s.log.Info("session ending", "error", errString(err))
	_ = s.sendStatus(StatusPayload{
		State: StateEnded,
		Reason: classifyEndReason(err),
		ElapsedMs: s.elapsedMs(),
	})
	_ = s.ws.Close()
	return err
}

// ---------- pump goroutines ----------

// pumpIn reads messages from the browser WS and forwards them to Gemini.
func (s *Session) pumpIn(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// No fixed read deadline; the heartbeat loop sets a short one
		// right after sending a ping, then resets it once the pong
		// arrives (in handleClientMessage → TypePing/Pong branch).
		_ = s.ws.SetReadDeadline(time.Time{})
		msgType, raw, err := s.ws.ReadMessage()
		if err != nil {
			return err
		}
		// We only accept text/JSON from the browser; binary is not used.
		if msgType != websocket.TextMessage {
			s.sendNonFatal(CodeBadMessage, "binary frames are not accepted", nil)
			continue
		}
		var env Envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			s.sendNonFatal(CodeBadMessage, "invalid json envelope", err)
			continue
		}
		if env.V != ProtocolVersion {
			s.sendNonFatal(CodeBadMessage, fmt.Sprintf("unsupported protocol version %d", env.V), nil)
			continue
		}
		if err := s.handleClientMessage(ctx, env); err != nil {
			return err
		}
	}
}

// handleClientMessage dispatches based on the envelope's Type field.
func (s *Session) handleClientMessage(ctx context.Context, env Envelope) error {
	switch env.Type {
	case TypeStart:
		// Session is already live after Run begins; start is a no-op for now.
		s.log.Debug("client: start")
		return nil
	case TypeAudioIn:
		return s.handleAudioIn(ctx, env.Payload)
	case TypeInterrupt:
		// For now, treat interrupt as activity-end so Gemini stops generating.
		// Future: drive an explicit barge-in via SendClientContent.
		s.log.Debug("client: interrupt — sending AudioStreamEnd to gemini")
		if err := s.gemini.SendRealtimeInput(genai.LiveRealtimeInput{
			AudioStreamEnd: true,
		}); err != nil {
			s.log.Error("send AudioStreamEnd to gemini failed", "error", err.Error())
			return err
		}
		return nil
	case TypeEnd:
		// Graceful shutdown: close the session root context.
		s.log.Debug("client: end")
		return io.EOF
	case TypePing:
		s.log.Debug("client: ping")
		return s.writeJSON(Envelope{
			V: ProtocolVersion,
			Type: TypePong,
			ID: uuid.NewString(),
			TS: time.Now().UnixMilli(),
			Payload: json.RawMessage(`{}`),
		})
	case TypePong:
		// Pong arrived within the deadline — clear the read deadline.
		s.log.Debug("client: pong")
		_ = s.ws.SetReadDeadline(time.Time{})
		return nil
	default:
		s.log.Warn("client: unknown message type", "type", env.Type)
		s.sendNonFatal(CodeBadMessage, "unknown message type: "+env.Type, nil)
		return nil
	}
}

// handleAudioIn validates the payload, decodes the PCM, and forwards it
// to Gemini as a realtime audio input chunk.
func (s *Session) handleAudioIn(ctx context.Context, raw json.RawMessage) error {
	var p AudioInPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		s.log.Warn("audio_in: invalid payload", "error", err.Error())
		s.sendNonFatal(CodeBadMessage, "invalid audio_in payload", err)
		return nil
	}
	if p.SampleRate != SampleRateIn || p.Encoding != Encoding || p.Channels != Channels {
		s.log.Warn("audio_in: format mismatch",
			"expected_sample_rate", SampleRateIn, "got_sample_rate", p.SampleRate,
			"expected_encoding", Encoding, "got_encoding", p.Encoding,
			"expected_channels", Channels, "got_channels", p.Channels)
		s.sendNonFatal(CodeBadAudio, fmt.Sprintf("expected %dHz %s %dch, got %dHz %s %dch",
			SampleRateIn, Encoding, Channels, p.SampleRate, p.Encoding, p.Channels), nil)
		return nil
	}
	pcm, err := DecodePCMBase64(p.PCM)
	if err != nil {
		s.log.Warn("audio_in: decode pcm failed", "error", err.Error())
		s.sendNonFatal(CodeBadAudio, "decode pcm failed", err)
		return nil
	}

	frameNo := s.audioInFrames.Add(1)
	if frameNo == 1 {
		s.log.Info("audio_in: first frame received from browser",
			"bytes", len(pcm), "duration_ms", p.DurationMs)
	}
	if frameNo == 1 || frameNo%50 == 0 {
		s.log.Debug("audio_in: forwarding to gemini",
			"frame_no", frameNo, "bytes", len(pcm), "duration_ms", p.DurationMs)
	}

	if err := s.gemini.SendRealtimeInput(genai.LiveRealtimeInput{
		Audio: &genai.Blob{
			MIMEType: fmt.Sprintf("audio/pcm;rate=%d", SampleRateIn),
			Data: pcm,
		},
	}); err != nil {
		s.log.Error("send audio to gemini failed", "frame_no", s.audioInFrames, "error", err.Error())
		return err
	}
	return nil
}

// pumpOut reads messages from Gemini and forwards them to the browser.
func (s *Session) pumpOut(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		msg, err := s.gemini.Receive()
		if err != nil {
			s.log.Debug("gemini receive ended", "error", errString(err))
			return err
		}
		if msg == nil {
			continue
		}
		if err := s.handleServerMessage(ctx, msg); err != nil {
			return err
		}
	}
}

// handleServerMessage translates one LiveServerMessage into zero or more
// browser frames.
func (s *Session) handleServerMessage(ctx context.Context, msg *genai.LiveServerMessage) error {
	if msg.SetupComplete != nil {
		s.log.Debug("gemini: setup complete")
	}
	if msg.GoAway != nil {
		s.log.Info("gemini: sent goaway")
		return io.EOF
	}
	if msg.ServerContent != nil {
		return s.handleServerContent(ctx, msg.ServerContent)
	}
	if msg.ToolCall != nil {
		s.log.Debug("gemini: tool call (unhandled)")
	}
	return nil
}

// handleServerContent unpacks the per-turn events: model audio, transcripts,
// turnComplete, interrupted, generationComplete.
func (s *Session) handleServerContent(ctx context.Context, c *genai.LiveServerContent) error {
	if c.InputTranscription != nil {
		s.log.Debug("gemini: input transcription",
			"text", c.InputTranscription.Text, "finished", c.InputTranscription.Finished)
		if err := s.sendTranscript(TranscriptPayload{
			Role: RoleUser,
			Text: c.InputTranscription.Text,
			TurnComplete: c.InputTranscription.Finished,
			TurnID: "t_user",
		}); err != nil {
			return err
		}
	}
	if c.OutputTranscription != nil {
		s.log.Debug("gemini: output transcription",
			"text", c.OutputTranscription.Text, "finished", c.OutputTranscription.Finished)
		if err := s.sendTranscript(TranscriptPayload{
			Role: RoleModel,
			Text: c.OutputTranscription.Text,
			TurnComplete: c.OutputTranscription.Finished,
			TurnID: "t_model",
		}); err != nil {
			return err
		}
	}
	if c.ModelTurn != nil {
		s.log.Debug("gemini: model turn", "parts", len(c.ModelTurn.Parts))
		for _, part := range c.ModelTurn.Parts {
			if part.InlineData == nil {
				s.log.Debug("gemini: model turn part has no inline data")
				continue
			}
			s.log.Debug("gemini: model turn inline data",
				"mime_type", part.InlineData.MIMEType, "bytes", len(part.InlineData.Data))
			if !isAudioMime(part.InlineData.MIMEType) {
				s.log.Debug("gemini: skipping non-audio inline data", "mime_type", part.InlineData.MIMEType)
				continue
			}
			if err := s.sendAudioOut(AudioOutPayload{
				PCM: EncodePCMBase64(part.InlineData.Data),
				SampleRate: SampleRateOut,
				Encoding: Encoding,
				Channels: Channels,
				DurationMs: PCMDurationMs(len(part.InlineData.Data), SampleRateOut),
				Final: c.TurnComplete,
			}); err != nil {
				return err
			}
		}
	}
	if c.Interrupted {
		s.log.Debug("gemini: interrupted")
		if err := s.sendStatus(StatusPayload{State: StateInterrupted}); err != nil {
			return err
		}
	}
	if c.TurnComplete {
		s.log.Debug("gemini: model turn complete")
		// TurnComplete arrives as its own ServerContent message, separate
		// from the transcription/audio chunks. Forward it as an empty
		// turnComplete marker for both roles so the client freezes the
		// current turns and starts fresh entries for the next exchange.
		if err := s.sendTranscript(TranscriptPayload{
			Role: RoleModel,
			Text: "",
			TurnComplete: true,
			TurnID: "t_model",
		}); err != nil {
			return err
		}
		if err := s.sendTranscript(TranscriptPayload{
			Role: RoleUser,
			Text: "",
			TurnComplete: true,
			TurnID: "t_user",
		}); err != nil {
			return err
		}
	}
	return nil
}

// heartbeatLoop sends periodic pings and tears down the session if the
// client doesn't pong back in time.
func (s *Session) heartbeatLoop(ctx context.Context, ticker *time.Ticker) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Send a ping and wait up to PongTimeout for a pong.
			pingEnv := Envelope{
				V: ProtocolVersion,
				Type: TypePing,
				ID: uuid.NewString(),
				TS: time.Now().UnixMilli(),
				Payload: json.RawMessage(`{}`),
			}
			if err := s.writeJSON(pingEnv); err != nil {
				return err
			}
			if err := s.ws.SetReadDeadline(time.Now().Add(s.cfg.PongTimeout)); err != nil {
				return err
			}
			// The actual pong handling is in pumpIn; the read deadline
			// above applies to ALL reads including the next pumpIn
			// ReadMessage. We reset the deadline inside pumpIn after
			// each successful read.
		}
	}
}

// audioInWatchdog logs a periodic summary of audio_in frames received from
// the browser. This is the fastest way to tell, from the server side
// alone, whether the browser's mic capture is sending anything at all —
// if "received" stays at 0 after the first tick, the problem is in the
// browser (mic permission, AudioContext/worklet, or the WS never reaching
// "open"), not in the Gemini bridge.
func (s *Session) audioInWatchdog(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	var last uint64
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			total := s.audioInFrames.Load()
			received := total - last
			if received == 0 {
				s.log.Warn("audio_in: no frames received from browser in last 2s", "total_frames", total)
			} else {
				s.log.Info("audio_in: heartbeat", "frames_last_2s", received, "total_frames", total)
			}
			last = total
		}
	}
}

// ---------- outbound helpers ----------

func (s *Session) sendReady() error {
	payload, err := json.Marshal(ReadyPayload{
		SessionID: s.id,
		Model: s.model,
		SampleRateOut: SampleRateOut,
		SampleRateIn: SampleRateIn,
		InputEncoding: Encoding,
		OutputEncoding: Encoding,
		MaxSessionSeconds: int(s.cfg.SessionMaxSeconds.Seconds()),
	})
	if err != nil {
		return err
	}
	if err := s.writeJSON(Envelope{
		V: ProtocolVersion,
		Type: TypeReady,
		ID: uuid.NewString(),
		TS: time.Now().UnixMilli(),
		Payload: payload,
	}); err != nil {
		return err
	}
	s.liveAt = time.Now()
	s.log.Info("session live", "model", s.model)
	return nil
}

func (s *Session) sendStatus(p StatusPayload) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return s.writeJSON(Envelope{
		V: ProtocolVersion,
		Type: TypeStatus,
		ID: uuid.NewString(),
		TS: time.Now().UnixMilli(),
		Payload: payload,
	})
}

func (s *Session) sendTranscript(p TranscriptPayload) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return s.writeJSON(Envelope{
		V: ProtocolVersion,
		Type: TypeTranscript,
		ID: uuid.NewString(),
		TS: time.Now().UnixMilli(),
		Payload: payload,
	})
}

func (s *Session) sendAudioOut(p AudioOutPayload) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return s.writeJSON(Envelope{
		V: ProtocolVersion,
		Type: TypeAudioOut,
		ID: uuid.NewString(),
		TS: time.Now().UnixMilli(),
		Payload: payload,
	})
}

func (s *Session) sendFatal(code, msg string, cause error) {
	payload, _ := json.Marshal(ErrorPayload{
		Code: code,
		Message: msg,
		Fatal: true,
		Cause: errString(cause),
	})
	_ = s.writeJSON(Envelope{
		V: ProtocolVersion,
		Type: TypeError,
		ID: uuid.NewString(),
		TS: time.Now().UnixMilli(),
		Payload: payload,
	})
}

func (s *Session) sendNonFatal(code, msg string, cause error) {
	payload, _ := json.Marshal(ErrorPayload{
		Code: code,
		Message: msg,
		Fatal: false,
		Cause: errString(cause),
	})
	_ = s.writeJSON(Envelope{
		V: ProtocolVersion,
		Type: TypeError,
		ID: uuid.NewString(),
		TS: time.Now().UnixMilli(),
		Payload: payload,
	})
}

// writeJSON serialises the envelope and sends it as a single text frame.
func (s *Session) writeJSON(e Envelope) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_ = s.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return s.ws.WriteJSON(e)
}

// ---------- misc helpers ----------

func (s *Session) elapsedMs() int64 {
	if s.liveAt.IsZero() {
		return 0
	}
	return time.Since(s.liveAt).Milliseconds()
}

func classifyEndReason(err error) string {
	if err == nil {
		return ReasonUserEnded
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ReasonSessionTimeout
	}
	if errors.Is(err, io.EOF) {
		return ReasonGeminiDisconnected
	}
	var wsCloseErr *websocket.CloseError
	if errors.As(err, &wsCloseErr) {
		return ReasonGeminiDisconnected
	}
	return ReasonError
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// isAudioMime reports whether mime identifies PCM audio. Gemini sends
// the sample rate as a parameter, e.g. "audio/pcm;rate=24000", so this
// matches on the base type rather than an exact string.
func isAudioMime(mime string) bool {
	base := mime
	if i := strings.IndexByte(mime, ';'); i >= 0 {
		base = mime[:i]
	}
	switch strings.ToLower(strings.TrimSpace(base)) {
	case "audio/pcm", "audio/l16", "audio/raw":
		return true
	default:
		return false
	}
}

// defaultSystemPrompt is the sent to Gemini. Intentionally
// minimal — POC scope is plain conversation.
const defaultSystemPrompt = "You are a concise, friendly voice assistant for an internal demo. " +
	"Keep replies short (1-3 sentences) and natural for spoken conversation. " +
	" give at most 3 items unless they ask for more."
