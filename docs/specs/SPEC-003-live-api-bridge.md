# SPEC-003: Live API Bridge with RAG Tool

**Date:** 2026-06-11
**Status:** DRAFT
**Branch:** rag-poc
**Depends on:** SPEC-002

---

## 1. Goal

Expose RAG as a callable tool to Gemini Live API so the model can retrieve context from uploaded documents during a conversation.

The client (web / mobile) opens a **WebSocket** to our API. Our API maintains its own session with Gemini Live API, relays audio/video/text bidirectionally, and registers the RAG tool so Gemini can call it as needed.

---

## 2. Architecture

```
User (voice/text)
 ↓ WebSocket
API WebSocket endpoint (SPEC-001: gorilla/websocket)
 ↔︎ Gemini Live API (google.golang.org/genai, LiveClient)
 ↔︎ RAG tool (FetchHighAccuracyContext from SPEC-002)
 ↔︎ DuckDB (chunk tables)
```

Key types from `genai`:
- `Client` created with `BackendGeminiAPI + APIKey`
- `LiveClient` via `client.Live.StartChat(ctx, model, config)`
- `Tool` with `FunctionDeclaration` for `search_documents`

---

## 3. Function Declaration (RAG Tool)

```go
searchDocumentsTool := &genai.Tool{
 FunctionDeclarations: []*genai.FunctionDeclaration{{
 Name: "search_documents",
 Description: "Search uploaded documents for relevant context. Use this when the user's question asks about something from a document.",
 Parameters: &genai.Schema{
 Type: genai.TypeObject,
 Properties: map[string]*genai.Schema{
 "query": {Type: genai.TypeString, Description: "The user's question or search phrase"},
 },
 Required: []string{"query"},
 },
 }},
}}
```

When Gemini decides it needs context, the response contains a `FunctionCall` part. We handle it by:

1. Extract `query` from `funcCall.Args["query"]`
2. Call `Fetcher.FetchHighAccuracyContext(ctx, query)` (from SPEC-002)
3. Return the context string as a `FunctionResponse` part in a follow-up `GenerateContent` call with the same history.

---

## 4. Session Lifecycle

1. **Client connects:** `ws://api/ws?session_id=xxx` (or auto-generate session ID).
2. **API starts GEMINI session:** `liveClient.StartChat(ctx, "gemini-2.0-live", config)`.
 - config includes our `searchDocumentsTool`.
3. **Bidirectional relay:** API reads from client's WebSocket → sends to `liveClient.SendContent()`; API reads from `liveClient.Receive()` → writes to client's WebSocket.
4. **If Gemini calls tool:**
 - The response `Part` is a `FunctionCall`.
 - Execute tool locally.
 - Send `FunctionResponse` back with the tool's output.
5. **On disconnect:** `liveClient.End()` → close client WebSocket → cleanup session.

---

## 5. API Changes (extend SPEC-001's server)

New endpoint:

| Method | Endpoint | Description |
|--------|----------|-------------|
| `WS` | `/ws` | WebSocket upgrade. Query param: `session_id` (optional). |

The WebSocket handler gets the underlying `*liveClient` for the session. Sessions stored in `map[string]*LiveSession` with automatic cleanup on disconnect.

---

## 6. Configuration

- `GEMINI_MODEL = gemini-2.0-live` (or whichever is the current Live model)
- `LIVE_API_KEY` (same as `GEMINI_API_KEY` from SPEC-002)
- `SESSION_TIMEOUT = 30m` (or similar — active session expiry)
- `MAX_CONCURRENT_SESSIONS` (default: 10) — rate limit to protect your Gemini quota

---

## 7. Session State Management

Sessions are ephemeral (in-memory). After a client disconnects, it's gone — no persistence needed for the POC.

| Session Field | Type |
|---------------|------|
| session_id | string |
| client_ws | *websocket.Conn |
| live_client | *genai.LiveClient |
| started_at | time.Time |

On disconnect: remove from map, `live_client.End()`, close client WS.

---

## 8. Error Handling

- If Gemini API errors: send a text message to the client explaining the error, don't kill the session.
- If client disconnects unexpectedly: `live_client.End()` + cleanup.
- If RAG tool errors: return the error as the function response (so Gemini can tell the user "I couldn't find that information").

---

## 9. Tests

- **Unit:** tool declaration parsing, session map logic, error transformation
- **Integration:** start a real WebSocket, connect to a mock Gemini Live (or recorded responses), verify tool call → response round-trip
- **E2E:** real WebSocket → real Gemini Live (requires API key), push audio → verify response

---

## 10. Out of Scope

- Session persistence to database (future: rejoin a session)
- Multi-tenant session isolation (future: session scoping by org)
- Voice activity detection tuning
- Real-time transcription customization (Gemini handles this)

---

## Decision Log

| Date | Decision |
|------|----------|
| 2026-06-11 | API bridges Gemini Live (client → API → Gemini Live) |
| 2026-06-11 | Tool declared inline via genai.Tool |
| 2026-06-11 | Session lives in-memory, not persisted |
| 2026-06-11 | RAG tool = FetchHighAccuracyContext from SPEC-002 |