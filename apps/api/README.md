# Live API — backend

Go backend for the admin voice session. Proxies audio + transcripts
between the browser and Gemini Live over a single WebSocket
(`/v1/live`). See `docs/specs/in-progress/SPEC-001-gemini-live-implementation.md`
for the wire format.

## Requirements

- Go 1.24+ (the module's `go.mod` requires toolchain `1.25`, which
 `go` will auto-download on first build)
- A Gemini API key with Live API access

## Configuration

Copy `.env` to `.env.local` and fill in:

```bash
cp .env .env.local
# edit .env.local — set GEMINI_API_KEY
```

Variables (all have safe defaults except the API key):

| Variable | Default | Notes |
|---|---|---|
| `PORT` | `8080` | Gin binds `:PORT` |
| `GEMINI_API_KEY` | _required_ | Server exits at boot if empty |
| `GEMINI_MODEL` | `gemini-2.0-flash-live-001` | Passed to `Live.Connect` |
| `ALLOWED_ORIGIN` | `http://localhost:5173` | Browser origin allowed on the WS handshake |
| `SESSION_MAX_SECONDS` | `600` | Hard cap per session |
| `WS_PING_INTERVAL_SECONDS` | `15` | Server-side heartbeat |
| `WS_PONG_TIMEOUT_SECONDS` | `5` | Must be < ping interval |

The server reads from the process environment, not the `.env` files
directly. For local dev, load the file before running:

```bash
set -a; source .env.local; set +a
go run ./cmd/server
```

Or use any `.env` loader of your choice.

## Run

```bash
go run ./cmd/server
# {"time":"…","level":"INFO","msg":"starting live-api-poc backend", "port":"8080", …}
```

Health check:

```bash
curl http://localhost:8080/healthz
# {"status":"ok"}
```

## Endpoints

- `GET /healthz` — liveness probe.
- `GET /v1/live` — WebSocket upgrade. On success, the server sends a
 `ready` JSON envelope and begins proxying to Gemini Live. See the
 spec for the full protocol.

## Layout

```
cmd/server/ # main
internal/
 config/ # env loader
 logger/ # slog + redaction
 live/ # WS protocol, audio helpers, session bridge, handler
 server/ # Gin router, middleware
```

## Logging

Structured JSON to stderr. The API key and any field whose name
contains `apikey`, `authorization`, `secret`, `password`, or `token`
is scrubbed to `[redacted]` (or first-4 + last-4 when long enough).
