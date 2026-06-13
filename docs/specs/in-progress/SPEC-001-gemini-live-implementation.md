# SPEC — Gemini Live API Implementation (Admin-initiated voice session)

**Status:** Draft, awaiting Sreyash's approval
**Date:** 2026-06-13
**Branch:** backend-setup
**Author:** huddle (Claude)

---

## 1. Goal

Add a working end-to-end voice session feature to the Live API platform:

- A new admin page at `/console/live` with an "Initiate Session" button.
- Click → backend opens a Gemini Live session → user talks via mic → Gemini responds with audio → both sides' transcripts are visible in real time.
- Token and model name read from environment.

## 2. Scope

**In scope:**
- New Go module `apps/api` with a WebSocket server.
- New SDK dependency: `github.com/googleapis/go-genai` (Go SDK for Gemini).
- A bidirectional WebSocket bridge `Browser ↔ Go ↔ Gemini Live`.
- PCM audio I/O in both directions.
- Transcript relay (input + output) from Gemini Live to the browser as JSON messages.
- New admin page `/console/live` in the UI with mic capture, audio playback, transcript panel, end-session button.
- A link from `/console` dashboard to `/console/live`.

**Out of scope (explicit):**
- No auth on the route (open).
- No RAG, no `retrieve_knowledge` tool, no MCP.
- No conversation persistence, no Postgres writes from this feature.
- No agent handoff, no queue.
- No multi-session support beyond what the server can hold in memory.
- No `/b/*` or `/p/*` route wiring.

## 3. Decisions (from huddle)

| Question | Decision |
|---|---|
| Model | `gemini-3.1-flash-live-preview`, read from env `GEMINI_MODEL` at runtime. |
| Route | New page `/console/live`. |
| Auth | Open. No middleware, no token. |
| Scope | Plain conversation only. Just a , no tools. |

## 4. Environment

`apps/api/.env.local` (exists, update values):

```bash
PORT=8080
GEMINI_API_KEY=__set_by_user__
GEMINI_MODEL=gemini-3.1-flash-live-preview
ALLOWED_ORIGIN=http://localhost:5173
SESSION_MAX_SECONDS=600
```

- Server fails fast at boot if `GEMINI_API_KEY` is empty.
- `GEMINI_API_KEY` is **never** logged. Logger redacts any field named `apiKey`/`api_key`/`Authorization`.

## 5. Architecture

```
Browser (Vite UI)
 │ WebSocket ws://localhost:8080/v1/live
 │ binary frames: PCM Int16 LE mono, 16 kHz (mic in) / 24 kHz (model out)
 │ text frames: JSON control {type, ...}
 ▼
Go API (apps/api)
 internal/live.Handler
 │ owns: Gemini session, WS conn, two goroutines (up/down), ctx
 ▼
google.golang.org/genai → Gemini Live (model from env)
```

Two goroutines per session:
- **`pumpIn`**: WS binary → Gemini `SendAudio`.
- **`pumpOut`**: Gemini events → WS (transcripts as JSON, audio as binary).

Session ends when: browser sends `{type:"end"}`, browser disconnects, or `SESSION_MAX_SECONDS` elapses.

## 6. WebSocket protocol

**Endpoint:** `ws://localhost:8080/v1/live`
**Subprotocol:** none (default)
**Content:** every frame is a **text/JSON** message. No binary frames on this socket. Audio is carried inside JSON as base64-encoded PCM (see §6.6 for the rationale).

### 6.0 Envelope (applies to all frames)

Every frame — in both directions — has this shape:

```ts
type Envelope<T extends MessageType, P> = {
  v: 1;            // protocol version. Bump on any breaking change.
  type: T;         // discriminator
  id: string;      // UUIDv4, unique per frame. Server-generated; client echoes on replies.
  ts: number;      // Unix epoch milliseconds, sender's clock.
  payload: P;      // type-specific body. Always present, never null.
};
```

**Rules:**
- The discriminator is `type`. Both peers must reject (with an `error` frame) any message whose `type` is unknown.
- `payload` is always an object, never `null`, even when empty — use `{}`.
- Unknown fields in `payload` are ignored (forward-compat). Missing required fields cause an `error` frame with `code: "bad_message"`.

### 6.1 Message types

```ts
// ---------- server → client ----------
type ServerMsg =
  | Envelope<"ready",        ReadyPayload>
  | Envelope<"audio_out",    AudioOutPayload>
  | Envelope<"transcript",   TranscriptPayload>
  | Envelope<"status",       StatusPayload>
  | Envelope<"error",        ErrorPayload>;

interface ReadyPayload {
  sessionId: string;            // server-assigned session id (UUIDv4)
  model: string;                // resolved model name, e.g. "gemini-3.1-flash-live-preview"
  sampleRateOut: 24000;         // fixed: Gemini output rate
  sampleRateIn: 16000;          // fixed: required input rate
  inputEncoding: "pcm_s16le";   // fixed
  outputEncoding: "pcm_s16le";  // fixed
  maxSessionSeconds: number;    // server's hard cap, client may display
}

interface AudioOutPayload {
  // 24 kHz mono PCM Int16 little-endian, base64-encoded.
  // Server coalesces Gemini audio chunks into ~20–40 ms frames before sending.
  pcm: string;
  sampleRate: 24000;
  encoding: "pcm_s16le";
  channels: 1;
  durationMs: number;           // pcm length / (sampleRate * 2) — pre-computed server-side
  final: boolean;               // true when this is the last frame of the current model turn
}

interface TranscriptPayload {
  role: "user" | "model";
  text: string;                 // may be partial; replace prior partial for the same turn
  turnComplete: boolean;        // true → this turn is final, no more text for it
  turnId: string;               // groups partials. New turnId when turnComplete=true.
}

interface StatusPayload {
  state: "connecting" | "live" | "interrupted" | "ended";
  reason?: "user_ended" | "session_timeout" | "gemini_disconnected" | "error";
  elapsedMs?: number;           // present when state==="ended" — total session duration
}

interface ErrorPayload {
  code: ErrorCode;              // see §6.4
  message: string;              // human-readable, safe to surface in UI
  fatal: boolean;               // true → session is over, do not auto-retry
  cause?: string;               // debug-only, server may omit in production
}

// ---------- client → server ----------
type ClientMsg =
  | Envelope<"start",          StartPayload>
  | Envelope<"audio_in",       AudioInPayload>
  | Envelope<"interrupt",      InterruptPayload>     // user spoke over model
  | Envelope<"end",            EndPayload>
  | Envelope<"ping",           PingPayload>
  | Envelope<"pong",           PongPayload>;

interface StartPayload {
  // Reserved for future per-session config (system prompt override, locale, etc).
  // For now, server builds config from env + defaults. Client may send {}.
}

interface AudioInPayload {
  // 16 kHz mono PCM Int16 little-endian, base64-encoded.
  // Browser sends ~20 ms frames (~640 samples = 1280 bytes → ~1706 bytes b64).
  pcm: string;
  sampleRate: 16000;
  encoding: "pcm_s16le";
  channels: 1;
  durationMs: number;
}

interface InterruptPayload {
  // Tells server to flush pending model output. Corresponds to Gemini's
  // "context window compression" or a barge-in event.
  turnId: string;               // turn the client is interrupting
}

interface EndPayload {
  reason: "user_ended";         // only "user_ended" for now; reserved for future values
}

interface PingPayload {
  // Empty; presence of frame is the heartbeat.
}

interface PongPayload {
  // Response to a server ping. Empty; presence is enough.
}
```

### 6.2 Lifecycle (ordered message flow)

```
client                          server                          Gemini
  │ ─── start ──────────────────► │
  │                              │ ── Live.Connect ────────────► │
  │                              │ ◄── session ready ─────────── │
  │ ◄── ready ──────────────────  │
  │                              │
  │ ─── audio_in (N frames) ────► │ ── SendAudio ───────────────► │
  │                              │ ◄── inputTranscription ─────── │
  │ ◄── transcript (user) ──────  │                              │
  │                              │ ◄── modelTurn.audio ───────── │
  │ ◄── audio_out (M frames) ───  │                              │
  │ ◄── transcript (model) ─────  │                              │
  │                              │
  │ ─── end ────────────────────► │ ── Live.Close ──────────────► │
  │ ◄── status(ended) ──────────  │
  │ [close]                      │ [close]
```

- The server **must** send exactly one `ready` frame before any other server→client message, or one `error(fatal=true)` frame and close.
- The server **must** send `status.state="ended"` as the last server→client frame before closing.
- A `transcript` with `turnComplete=true` is the final text for that `turnId`. Clients should freeze it.

### 6.3 Heartbeats (ping/pong)

- The server sends `{type:"ping"}` every **15 s** of silence (no audio, no transcript in either direction).
- Client must respond with `{type:"pong"}` within **5 s**. Two missed pongs → server treats as dead, sends `status(ended, reason="gemini_disconnected")` and closes.
- Client may also send `{type:"ping"}` to check liveness. Server responds with `{type:"pong"}` within 1 s.

### 6.4 Error codes

| Code | fatal | Meaning |
|---|---|---|
| `bad_message` | true | Malformed JSON, unknown `type`, or missing required fields. |
| `bad_audio` | false | One frame had invalid PCM (wrong length, wrong sample rate). Dropped, others continue. |
| `gemini_unavailable` | true | Gemini session could not be created. |
| `gemini_disconnected` | true | Upstream Gemini session closed unexpectedly. |
| `session_timeout` | true | `SESSION_MAX_SECONDS` exceeded. |
| `rate_limited` | false | Client sent frames too fast. Slow down, no session teardown. |
| `internal` | true | Unhandled server error. |

### 6.5 Example exchange (raw)

```json
// client → server
{"v":1,"type":"start","id":"7c4…","ts":1718280000123,"payload":{}}

// server → client
{"v":1,"type":"ready","id":"a1b…","ts":1718280000456,"payload":{
  "sessionId":"sess_01HXY…",
  "model":"gemini-3.1-flash-live-preview",
  "sampleRateOut":24000,"sampleRateIn":16000,
  "inputEncoding":"pcm_s16le","outputEncoding":"pcm_s16le",
  "maxSessionSeconds":600
}}

// client → server (audio in, ~20 ms of 16 kHz PCM)
{"v":1,"type":"audio_in","id":"b2c…","ts":1718280001234,"payload":{
  "pcm":"<base64 of 1280 bytes>","sampleRate":16000,
  "encoding":"pcm_s16le","channels":1,"durationMs":20
}}

// server → client (transcript of what the user said, partial)
{"v":1,"type":"transcript","id":"c3d…","ts":1718280001500,"payload":{
  "role":"user","text":"How do I","turnComplete":false,"turnId":"t_42"
}}

// server → client (model audio, ~20 ms of 24 kHz PCM)
{"v":1,"type":"audio_out","id":"d4e…","ts":1718280002100,"payload":{
  "pcm":"<base64>","sampleRate":24000,
  "encoding":"pcm_s16le","channels":1,"durationMs":20,"final":false
}}

// client → server
{"v":1,"type":"end","id":"e5f…","ts":1718280099999,"payload":{"reason":"user_ended"}}

// server → client
{"v":1,"type":"status","id":"f6g…","ts":1718280100050,"payload":{
  "state":"ended","reason":"user_ended","elapsedMs":99876
}}
```

### 6.6 Why JSON-only (no binary frames) — design note

- **Pro**: one parser path on both sides, easy to log/replay/debug, schema is in the spec.
- **Pro**: works through every WS-aware proxy and CDN without subprotocol tricks.
- **Pro**: cleanly versioned; old clients can talk to new servers as long as `v` matches.
- **Con**: ~33% base64 overhead on audio. At 16 kHz mono × 16 bit × 20 ms = 1280 bytes audio → ~1706 bytes b64. Negligible at 50 KB/s audio.
- **Con**: no zero-copy streaming. Mitigation: server coalesces to ~20 ms frames anyway (matches browser `AudioWorklet` cadence), so per-frame overhead is bounded.
- If profiling later shows base64 is the bottleneck, switch to a `Sec-WebSocket-Protocol: audio.pcm` subprotocol and split audio to binary frames. The discriminated `type` field makes the migration additive.

## 7. Backend file layout

```
apps/api/
├── go.mod # module api, go 1.22+
├── go.sum
├── README.md # how to set GEMINI_API_KEY, run, test
├── .env # (existing, base)
├── .env.local # (existing, updated with GEMINI_API_KEY + GEMINI_MODEL)
├── cmd/
│ └── server/
│ └── main.go # wires config, logger, router; calls ListenAndServe
└── internal/
 ├── config/
 │ └── config.go # Load(): reads env, returns Config or error
 ├── logger/
 │ └── logger.go # slog wrapper, redaction helper
 ├── live/
 │ ├── session.go # LiveSession: owns Gemini conn + WS + 2 goroutines
 │ ├── protocol.go # ClientMsg / ServerMsg structs + JSON helpers
 │ ├── audio.go # PCM Int16 ↔ base64 helpers (used by tests + pump goroutines)
 │ └── handler.go # Gin handler that upgrades to WS
 └── server/
 ├── router.go # Gin engine, mounts /v1/live + CORS
 └── middleware.go  # CORS, request logging (no auth)
```

**Dependencies:**
- `github.com/googleapis/go-genai` — Gemini SDK
- `github.com/gin-gonic/gin` — HTTP framework
- `github.com/gorilla/websocket` — WS upgrade (called from inside the Gin handler)
- Standard library `log/slog` for logging (no Logrus/Zap needed)

## 8. Backend behaviors

### 8.1 Boot
1. `config.Load()` — reads env, returns `Config{Port, GeminiAPIKey, GeminiModel, AllowedOrigin, SessionMaxSeconds}`.
2. If `GeminiAPIKey == ""` → log fatal, exit 1.
3. Build `genai.Client` with API key.
4. Build Gin engine via `router.New(cfg, genaiClient)` returning `*gin.Engine`.
5. `engine.Run(":" + cfg.Port)`.

### 8.2 Session open (`/v1/live` upgrade)
1. Check `Origin` header against `AllowedOrigin`; reject if mismatch.
2. Upgrade to WebSocket using `gorilla/websocket.Upgrader` (called from inside the Gin handler).
3. Send `genai.Live.Connect(model, liveConfig)` where `liveConfig` enables:
 - `ResponseModalities: []string{"AUDIO"}`
 - `InputAudioTranscription: &AudioTranscriptionConfig{}` (transcribe mic)
 - `OutputAudioTranscription: &AudioTranscriptionConfig{}` (transcribe model audio)
 - `SystemInstruction: &Content{Parts: []Part{{Text: <default prompt>}}}`
4. Send `{type:"ready", sessionId, sampleRateOut:24000}`.
5. Start `pumpIn` and `pumpOut` goroutines.
6. Defer: close Gemini session, close WS, wait for goroutines.

### 8.3 `pumpIn` (browser → Gemini)
- Read WS messages in a loop.
- Binary → `client.Live.SendAudio(ctx, pcmBytes)`.
- Text → JSON-decode; if `{type:"end"}` → cancel ctx, return.
- On read/write error → log, cancel ctx, return.

### 8.4 `pumpOut` (Gemini → browser)
- Read Gemini events in a loop.
- `ServerContent{ModelTurn.Parts}` with `InlineData{MIME:"audio/pcm"}` → forward bytes as binary WS frame.
- `ServerContent{ModelTurn.Parts}` with `Text` → wrap as `{type:"transcript", role:"model", turnComplete, ts}`.
- `ServerContent{InputTranscription}` → wrap as `{type:"transcript", role:"user", turnComplete, ts}`.
- `ServerContent{TurnComplete}` → flush last `transcript` with `turnComplete:true`.
- `GoAway` or `Closed` → send `{type:"status", state:"ended"}`, cancel ctx.
- On read/write error → send `{type:"error", code:"internal", message}`, cancel ctx.

### 8.5 Timeouts
- `SESSION_MAX_SECONDS` (default 600): `context.WithTimeout` on session root. On expiry → send `{type:"status", state:"ended", reason:"session_timeout"}`, close.

## 9. UI file layout

```
apps/ui/src/
├── pages/
│ └── live-session/
│ ├── index.tsx # the /console/live page
│ ├── use-live-session.ts  # Zustand store + WS lifecycle hook
│ ├── audio.ts # mic capture + playback using AudioContext
│ ├── labels.ts # page-specific labels
│ └── parts/
│ ├── initiate-card.tsx # shown when state=idle
│ ├── session-view.tsx # shown when state=connecting|live
│ ├── transcript-list.tsx # append-only transcript list
│ ├── status-pill.tsx # connection state
│ └── controls-bar.tsx # mic toggle + end button
├── lib/
│ └── live-ws.ts # tiny WS client with auto-reconnect OFF
└── labels/
 └── business.ts # add LIVE_SESSION labels
```

Plus:
- `apps/ui/src/main.tsx` — add `<Route path="/console/live" element={<LiveSessionPage />} />`.
- `apps/ui/src/lib/paths.ts` — add `CONSOLE.live = "/console/live"`.
- `apps/ui/src/pages/console/index.tsx` — add a single CTA card linking to `/console/live`. (Stays small, doesn't dominate the dashboard.)
- `apps/ui/.env` or `vite.config.ts` — `VITE_API_WS_URL=ws://localhost:8080/v1/live`.

### 9.1 UI states

| State | Visible |
|---|---|
| `idle` | Initiate card: title, short copy, "Start live session" button. Mic permission not yet requested. |
| `connecting` | Initiate card shows spinner + "Connecting…" — no mic permission yet. |
| `live` | Session view: status pill (green, "Live · 0:42"), controls bar (Mute, End), transcript list scrolling. Mic permission requested here. |
| `ended` | Session view shows "Session ended" with a "Start a new session" button to return to `idle`. |

### 9.2 Audio pipeline (browser)

- **Capture**: `navigator.mediaDevices.getUserMedia({audio:{sampleRate:16000, channelCount:1, echoCancellation:true, noiseSuppression:true}})` → `MediaStreamSource` → `AudioWorkletNode` (PCM emitter) → post `Float32Array` to main thread → encode to `Int16Array` → send as binary WS frame every ~20 ms.
- **Playback**: incoming binary → `Int16Array` → `Float32Array` → `AudioBuffer` (24 kHz mono) → `AudioBufferSourceNode` → `AudioContext.destination`. Use a small jitter buffer (queue next 2–3 frames before starting).
- **No new dependencies.** `AudioContext`, `AudioWorklet`, `getUserMedia` are browser-native.

### 9.3 Transcript rendering

- Plain list, scroll-anchored to bottom on new entries.
- Each entry: speaker label (`You` / `Gemini`), monospace-friendly body, faint timestamp on hover.
- Streaming-friendly: a `transcript` message with `turnComplete=false` updates the current turn in place; `turnComplete=true` freezes it.
- Tailwind tokens only. New tokens (if needed) go in `src/styles/index.css` `@theme` block.

## 10. Verification (manual checklist)

The "perfectly working" bar Sreyash asked for is verified by this checklist, run end-to-end before marking done:

1. **Boot**: `cd apps/api && go run ./cmd/server` starts; missing `GEMINI_API_KEY` exits with a clear error.
2. **UI**: `cd apps/ui && npm run dev` serves on `http://localhost:5173`.
3. **Navigate**: `/console` (after onboarding stub) renders; "Start live session" link goes to `/console/live`.
4. **Initiate**: Click "Start live session" → status pill shows "Connecting" → within ~2s shows "Live".
5. **Speak**: Speak into the mic. Transcript shows "You" entries in near real time.
6. **Listen**: Gemini's audio plays through the speakers. Model transcript appears in the panel.
7. **End**: Click "End" → status changes to "Ended". Browser stops capturing audio.
8. **Timeout**: Leave a session open for 10 min → server sends `status=ended` with `reason=session_timeout`.
9. **Disconnect**: Kill the API process mid-session → browser sees WS close, status changes to "Ended", no JS errors in console.
10. **CORS**: Try opening from a different origin (e.g. `http://localhost:5174`) → upgrade rejected.
11. **Typecheck**: `cd apps/ui && npm run typecheck` clean. `cd apps/api && go build ./...` clean.

No automated e2e — manual checklist is the contract.

## 11. Risks

- **Gemini SDK API drift**: the `googleapis/go-genai` package is recent. The `Live.Connect` signature and event types may shift. Pin to a known-good version in `go.mod`. If types are unstable, the build will fail and we adjust.
- **Mic permission UX**: first-time users will be blocked by the browser permission prompt. The "Live" view must not request mic until the user clicks "Start", otherwise it auto-prompts on page load.
- **Audio glitches on first playback**: expect the first ~200ms of model audio to be choppy because of cold-start. Document this in the manual checklist, not a bug.
- **Cost**: a 10-minute live session consumes audio tokens continuously. `SESSION_MAX_SECONDS=600` is the only brake. Remind in README.

## 12. Out of scope, deferred to follow-ups

- Real auth + role gating (token from `gemini-live-architecture.md`).
- RAG via `retrieve_knowledge` tool.
- Conversation persistence to Postgres.
- Agent handoff / queue.
- Multi-tenant session isolation.
- Session recording + S3 upload.
- WebRTC path for lower-latency audio if WS relay proves insufficient.
- LLM-as-judge feedback scoring (worker job).
- `/b/*` and `/p/*` routing trees.
