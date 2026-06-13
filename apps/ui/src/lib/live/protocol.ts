/**
 * Wire protocol for /v1/live. Mirror of the Go types in
 * apps/api/internal/live/protocol.go and SPEC §6.
 *
 * All frames share the Envelope shape: {v, type, id, ts, payload}.
 * The discriminated `type` field is the source of truth; payloads
 * are unknown fields-tolerant (forward-compat) but missing required
 * fields produce a no-op on the client (we don't try to recover).
 */

export const PROTOCOL_VERSION = 1 as const;

export type MessageType =
 | "ready"
 | "audio_out"
 | "transcript"
 | "status"
 | "error"
 | "start"
 | "audio_in"
 | "interrupt"
 | "end"
 | "ping"
 | "pong";

export interface Envelope<T extends MessageType, P> {
 v: typeof PROTOCOL_VERSION;
 type: T;
 id: string;
 ts: number;
 payload: P;
}

// ---------- Server → Client payloads ----------

export interface ReadyPayload {
 sessionId: string;
 model: string;
 sampleRateOut: 24000;
 sampleRateIn: 16000;
 inputEncoding: "pcm_s16le";
 outputEncoding: "pcm_s16le";
 maxSessionSeconds: number;
}

export interface AudioOutPayload {
 pcm: string; // base64 of 24 kHz mono Int16 LE PCM
 sampleRate: 24000;
 encoding: "pcm_s16le";
 channels: 1;
 durationMs: number;
 final: boolean;
}

export type TranscriptRole = "user" | "model";

export interface TranscriptPayload {
 role: TranscriptRole;
 text: string;
 turnComplete: boolean;
 turnId: string;
}

export type StatusState = "connecting" | "live" | "interrupted" | "ended";
export type StatusReason =
 | "user_ended"
 | "session_timeout"
 | "gemini_disconnected"
 | "error";

export interface StatusPayload {
 state: StatusState;
 reason?: StatusReason;
 elapsedMs?: number;
}

export type ErrorCode =
 | "bad_message"
 | "bad_audio"
 | "gemini_unavailable"
 | "gemini_disconnected"
 | "session_timeout"
 | "rate_limited"
 | "internal"
 | "mic_denied"
 | "mic_error";

export interface ErrorPayload {
 code: ErrorCode;
 message: string;
 fatal: boolean;
 cause?: string;
}

// ---------- Client → Server payloads ----------

export interface StartPayload {
 // Reserved for future overrides; send {} for now.
}

export interface AudioInPayload {
 pcm: string; // base64 of 16 kHz mono Int16 LE PCM
 sampleRate: 16000;
 encoding: "pcm_s16le";
 channels: 1;
 durationMs: number;
}

export interface InterruptPayload {
 // Optional: which model turn we're interrupting. The server uses
 // Gemini's stream-level interrupt; UI doesn't need to set this.
 turnId?: string;
 reason?: "user_barge_in" | "user_stop";
}

export interface EndPayload {
 reason: "user_ended";
}

export interface PingPayload {
 // Empty — presence is the heartbeat.
}

export interface PongPayload {
 // Empty — response to a server ping.
}

// ---------- Discriminated unions ----------

export type ServerMsg =
 | Envelope<"ready", ReadyPayload>
 | Envelope<"audio_out", AudioOutPayload>
 | Envelope<"transcript", TranscriptPayload>
 | Envelope<"status", StatusPayload>
 | Envelope<"error", ErrorPayload>
 | Envelope<"ping", PingPayload>;

export type ClientMsg =
 | Envelope<"start", StartPayload>
 | Envelope<"audio_in", AudioInPayload>
 | Envelope<"interrupt", InterruptPayload>
 | Envelope<"end", EndPayload>
 | Envelope<"ping", PingPayload>
 | Envelope<"pong", PongPayload>;

// PCM constants — must stay in sync with apps/api/internal/live/audio.go.
export const SAMPLE_RATE_IN = 16000;
export const SAMPLE_RATE_OUT = 24000;
export const CHANNELS = 1;
export const ENCODING = "pcm_s16le" as const;
