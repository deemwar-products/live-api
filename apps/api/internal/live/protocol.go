// Package live implements the WebSocket ↔ Gemini Live API bridge.
//
// protocol.go defines the wire format that travels over the WS — every
// frame is a JSON envelope (see SPEC §6). Server-side mirrors of the
// client/server message types live here; encoding/decoding helpers
// keep `omitempty` discipline consistent.
package live

import "encoding/json"

// ProtocolVersion is bumped on any breaking change to the wire format.
const ProtocolVersion = 1

// Envelope is the universal wrapper for every frame in both directions.
// The discriminated `Type` field is required; `Payload` is always an
// object (use `{}` if there are no fields, never null).
type Envelope struct {
	V int `json:"v"`
	Type string `json:"type"`
	ID string `json:"id"`
	TS int64 `json:"ts"`
	Payload json.RawMessage `json:"payload"`
}

// ---------- Client → Server payloads ----------

// StartPayload is the body of {type:"start"}. Currently empty; reserved
// for future per-session overrides (, locale, etc).
type StartPayload struct{}

// AudioInPayload is the body of {type:"audio_in"}. PCM is 16 kHz mono
// Int16 little-endian, base64-encoded.
type AudioInPayload struct {
	PCM string `json:"pcm"`
	SampleRate int `json:"sampleRate"`
	Encoding string `json:"encoding"`
	Channels int `json:"channels"`
	DurationMs int `json:"durationMs"`
}

// InterruptPayload is the body of {type:"interrupt"}. Barge-in signal.
type InterruptPayload struct {
	TurnID string `json:"turnId"`
}

// EndPayload is the body of {type:"end"}.
type EndPayload struct {
	Reason string `json:"reason"`
}

// PingPayload is the body of {type:"ping"}.
type PingPayload struct{}

// PongPayload is the body of {type:"pong"}.
type PongPayload struct{}

// ---------- Server → Client payloads ----------

// ReadyPayload is sent once after the Gemini session is established.
type ReadyPayload struct {
	SessionID string `json:"sessionId"`
	Model string `json:"model"`
	SampleRateOut int `json:"sampleRateOut"`
	SampleRateIn int `json:"sampleRateIn"`
	InputEncoding string `json:"inputEncoding"`
	OutputEncoding string `json:"outputEncoding"`
	MaxSessionSeconds int `json:"maxSessionSeconds"`
}

// AudioOutPayload is the body of {type:"audio_out"}. PCM is 24 kHz
// mono Int16 little-endian, base64-encoded. `Final=true` marks the
// last frame of the current model turn.
type AudioOutPayload struct {
	PCM string `json:"pcm"`
	SampleRate int `json:"sampleRate"`
	Encoding string `json:"encoding"`
	Channels int `json:"channels"`
	DurationMs int `json:"durationMs"`
	Final bool `json:"final"`
}

// TranscriptPayload is the body of {type:"transcript"}. May be
// partial; the client freezes a turn when TurnComplete=true.
type TranscriptPayload struct {
	Role string `json:"role"`
	Text string `json:"text"`
	TurnComplete bool `json:"turnComplete"`
	TurnID string `json:"turnId"`
}

// StatusPayload is the body of {type:"status"}. Reason is set on
// "ended" frames; ElapsedMs is set on terminal status.
type StatusPayload struct {
	State string `json:"state"`
	Reason string `json:"reason,omitempty"`
	ElapsedMs int64 `json:"elapsedMs,omitempty"`
}

// ErrorPayload is the body of {type:"error"}. Fatal=true means the
// session is over and the client should not auto-retry.
type ErrorPayload struct {
	Code string `json:"code"`
	Message string `json:"message"`
	Fatal bool `json:"fatal"`
	Cause string `json:"cause,omitempty"`
}

// Message type constants — keep in sync with the spec's `type` field.
const (
	TypeReady = "ready"
	TypeAudioOut = "audio_out"
	TypeTranscript = "transcript"
	TypeStatus = "status"
	TypeError = "error"

	TypeStart = "start"
	TypeAudioIn = "audio_in"
	TypeInterrupt = "interrupt"
	TypeEnd = "end"
	TypePing = "ping"
	TypePong = "pong"
)

// Status state values.
const (
	StateConnecting = "connecting"
	StateLive = "live"
	StateInterrupted = "interrupted"
	StateEnded = "ended"
)

// Transcript role values.
const (
	RoleUser = "user"
	RoleModel = "model"
)

// Reason values for StatusPayload.Reason and EndPayload.Reason.
const (
	ReasonUserEnded = "user_ended"
	ReasonSessionTimeout = "session_timeout"
	ReasonGeminiDisconnected = "gemini_disconnected"
	ReasonError = "error"
)

// Error codes — see SPEC §6.4.
const (
	CodeBadMessage = "bad_message"
	CodeBadAudio = "bad_audio"
	CodeGeminiUnavailable = "gemini_unavailable"
	CodeGeminiDisconnected = "gemini_disconnected"
	CodeSessionTimeout = "session_timeout"
	CodeRateLimited = "rate_limited"
	CodeInternal = "internal"
)
