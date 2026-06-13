# SPEC-005: RAG Ingestion UI + Dedicated Live Session Page

**Date:** 2026-06-12
**Status:** DRAFT (awaiting Sreyash approval)
**Branch:** rag-poc
**Depends on:** SPEC-002 (RAG pipeline), SPEC-003 (live bridge), SPEC-004 (live session decisions), SPEC queue-and-processor spec

---

## 1. Goal

Close the loop on the RAG PoC:

1. A business user can upload documents (MD, TXT, HTML, PDF, DOCX) from a new `/b/knowledge` page and watch them move `PENDING → INGESTING → READY` in real time.
2. The user can see queue health at a glance (queue depth, processing, ready/failed today, last activity).
3. The live session panel gets a dedicated `/b/live` route, real voice-activity feedback, and a working mute control.
4. The `LiveSessionPanel` reads its WebSocket URL from an env var so it works in dev, staging, and prod.

UI is the source of truth: every API endpoint in this spec exists because a screen needs it. The spec defines the contracts in the order the screens consume them.

---

## 2. UX Flow (drives the API shape)

### 2.1 `/b/knowledge` — empty state

- Single card, centered, with two affordances: drag-and-drop zone + "Browse files" button.
- Below the card: supported formats (`MD, TXT, HTML, PDF, DOCX`) and the 10MB cap notice.
- To the right: queue-status strip with five KPIs (queue depth, processing, ready today, failed today, last activity).

### 2.2 Upload moment

- User drops a file. The card validates client-side (extension + size). If invalid, the card shows an inline error and rejects the file before the network call.
- On accept, the row appears in the list immediately with `status=PENDING` and a thin top progress bar tracking the XHR upload.
- When the upload finishes, the progress bar is replaced by a status pill (`PENDING → INGESTING → READY`).
- The list polls `GET /api/v1/documents` every 1.5s while the page is focused. On blur, polling stops.
- The list polls `GET /api/v1/queue/status` every 1.5s in parallel.

### 2.3 Failure

- A `READY` row shows `chunk_count` (e.g., "12 chunks").
- A `FAILED` row shows the error text (from `documents.error`) and a "Retry" button that calls `POST /api/v1/documents/:id/requeue`.
- A `PENDING` row older than 30s gets a "Stuck?" badge (the worker sweep on startup is the primary recovery, but this surfaces the rare case in the UI).

### 2.4 `/b/live`

- Full-bleed layout, no sidebar. Header: "Live session" + small status dot (idle / connecting / listening / speaking / error).
- Center: a 200px orb that pulses with actual voice activity (analyser-driven), not connection state.
- Bottom: large mic button (Mute / Resume, single button two states), Disconnect button, transcript panel pinned to the bottom.
- On page load: detect WS upgrade failure with a quick fetch to `/ws`. If the server returns 503, render a specific card: "Server missing GEMINI_API_KEY. Set it in `apps/api/.env.local` and restart."
- On mic permission denied: an explanation card with a link to browser help, *and* a text input so the user can still type into the conversation.

---

## 3. API Contracts (new)

All routes are mounted under `/api/v1/`.

### 3.1 `POST /api/v1/documents`

**Request:** `multipart/form-data` with a single `file` field.

**Server behavior:**
1. Reject if `Content-Length > 10485760` (10MB) → `413`.
2. Reject if extension is not in `[md, txt, html, pdf, docx]` → `415`.
3. Stream the file to a temp path under `DocumentsDir` (e.g., `{DocumentsDir}/{doc_id}.{ext}`).
4. In a single DuckDB transaction:
 - `INSERT INTO documents (id, source_name, source_type, status, last_queued_at) VALUES (?, ?, ?, 'PENDING', NULL)`.
 - `INSERT INTO jobs (id, type, payload) VALUES (?, 'INGEST_DOCUMENT', ?)`.
5. `XADD jobs.rag * job_id {job_id} type INGEST_DOCUMENT`.
6. `UPDATE documents SET last_queued_at = now() WHERE id = ?`.
7. Best-effort: if step 5 fails, return `500` with the row already created; the worker startup sweep will recover it.

**Response 201:**
```json
{
 "id": "uuid",
 "source_name": "report.pdf",
 "source_type": "pdf",
 "status": "PENDING",
 "chunk_count": 0,
 "error": null,
 "created_at": "2026-06-12T10:00:00Z",
 "updated_at": "2026-06-12T10:00:00Z"
}
```

**Response error shape (all 4xx/5xx):**
```json
{ "error": "string", "code": "string" }
```

### 3.2 `GET /api/v1/documents`

**Query params:** `limit` (default 50, max 200), `offset` (default 0).

**Response 200:**
```json
{
 "items": [
 {
 "id": "uuid",
 "source_name": "report.pdf",
 "source_type": "pdf",
 "status": "PENDING|INGESTING|READY|FAILED",
 "chunk_count": 12,
 "error": null,
 "created_at": "...",
 "updated_at": "..."
 }
 ],
 "total": 123
}
```

### 3.3 `GET /api/v1/documents/:id`

Same single-object shape as a list item. `404` if not found.

### 3.4 `POST /api/v1/documents/:id/requeue`

**Server behavior:**
1. Load document. `404` if missing. `409` if `status == 'READY'`.
2. Reset `status = 'PENDING'`, `last_queued_at = NULL`, `error = NULL`.
3. `XADD jobs.rag * job_id {new_job_id} type INGEST_DOCUMENT` (new job row, new job_id).
4. `UPDATE documents SET last_queued_at = now() WHERE id = ?`.

**Response 202:** `{ "id": "uuid", "status": "PENDING" }`.

### 3.5 `GET /api/v1/queue/status`

Five numbers, one query each against DuckDB (cheap, indexed) plus one `XLEN` for queue depth:

```json
{
 "queue_depth": 3, // XLEN jobs.rag
 "processing": 1, // SELECT count(*) FROM jobs WHERE status='PROCESSING'
 "ready_today": 7, // SELECT count(*) FROM documents WHERE status='READY' AND created_at >= today()
 "failed_today": 0, // SELECT count(*) FROM jobs WHERE status='FAILED' AND updated_at >= today()
 "last_completed_at": "2026-06-12T10:14:32Z" // SELECT max(updated_at) FROM jobs WHERE status IN ('COMPLETED','FAILED')
}
```

If DuckDB query fails for any field, return `null` for that field (not a 500). The UI renders "—" for nulls.

---

## 4. Schema Change

### `005_add_last_queued_at.sql`

```sql
ALTER TABLE documents ADD COLUMN last_queued_at TIMESTAMP NULL;
CREATE INDEX IF NOT EXISTS idx_documents_pending_queue
 ON documents(status, last_queued_at)
 WHERE status = 'PENDING';
```

Worker startup sweep uses the partial index to find orphaned PENDING rows fast.

---

## 5. Worker Changes

### 5.1 Startup sweep

In `cmd/worker/main.go`, after the DB opens and before consumer goroutines start:

```go
sweepOrphanedDocuments(ctx, db, redisClient)
// sweepOrphanedDocuments:
// 1. SELECT id FROM documents
// WHERE status = 'PENDING'
// AND (last_queued_at IS NULL OR last_queued_at < now() - interval '5 minutes')
// 2. For each: insert a new jobs row, XADD jobs.rag, UPDATE documents.last_queued_at = now()
```

A 5-minute threshold avoids re-pushing documents the API is about to enqueue.

### 5.2 New payload shape for `INGEST_DOCUMENT`

Current SPEC-002 payload has `raw_bytes` (base64). With file-on-disk, the payload becomes:

```json
{
 "document_id": "uuid",
 "source_name": "report.pdf",
 "source_type": "pdf",
 "file_path": "/var/lib/live-api/documents/{doc_id}.pdf"
}
```

The API and worker must share a `DocumentsDir` mount. In Docker Compose, both services mount the same volume.

---

## 6. Config Changes

### `apps/api/internal/config/config.go` — add:
```go
DocumentsDir string // default: "../data/documents" (local), "/data/documents" (docker)
UploadMaxBytes int // default: 10485760
```

### `apps/worker/internal/config/config.go` — add:
```go
DocumentsDir string // default: same as API
```

### `docker-compose.yml` — share the volume:
```yaml
volumes:
 - api_documents:/data/documents
```
on both `api` and `worker` services.

---

## 7. UI Implementation

### 7.1 Routes

Add to the existing React Router setup (currently console-only at `/b`):

```tsx
<Route path="/b/knowledge" element={<KnowledgePage />} />
<Route path="/b/live" element={<LivePage />} />
```

Both wrap in the existing `<ProtectedRoute>` (per `apps/ui/AGENTS.md`).

### 7.2 New files

```
src/
├── pages/business/knowledge/
│ ├── index.tsx # KnowledgePage
│ ├── parts/
│ │ ├── upload-card.tsx # drag-and-drop + browse
│ │ ├── document-list.tsx # list with row states
│ │ ├── document-row.tsx
│ │ ├── queue-status-strip.tsx
│ │ └── file-validator.ts # pure function: validate file
│ ├── api.ts # fetch wrappers, polling hook
│ ├── labels.ts
│ └── mocks.ts # typed factories for tests
├── pages/business/live/
│ ├── index.tsx # LivePage (full-bleed)
│ └── labels.ts
├── components/ui/
│ ├── progress-bar.tsx
│ └── voice-orb.tsx # extracted from LiveSessionPanel
└── features/live/
 └── voice-activity.ts # pure function: compute RMS from analyser buffer
```

### 7.3 LiveSessionPanel fixes

- `WS_URL` reads from `import.meta.env.VITE_WS_URL` with fallback to `ws://${window.location.hostname}:8080/ws`.
- Add a mute state. When muted, the `MediaRecorder.ondataavailable` handler drops chunks but the WS stays open. Send `{type:"control", action:"mute"|"unmute"}` so the server can pause STT.
- Replace Orb pulse logic with real voice activity: an `AnalyserNode` tapped from the mic stream, polled via `requestAnimationFrame`, RMS computed in a pure function `computeRMS(uint8Array) → number` in `features/live/voice-activity.ts`. The orb scales its pulse by RMS, not by connection state.
- Detect WS upgrade failure: on `connect`, fire `fetch('/ws', { method: 'HEAD' })` first. If non-2xx, render the "missing GEMINI_API_KEY" card and don't open the WS.
- New `LivePage` (full-bleed) hosts the panel; remove it from the console.

### 7.4 Polling hook

```ts
// api.ts
function usePolledQuery<T>(key: string, fetcher: () => Promise<T>, intervalMs = 1500) {
 // pauses on document.hidden, resumes on visible
}
```

Used by both document list and queue status strip.

---

## 8. Error Handling

| Layer | Failure | UX |
|---|---|---|
| Client | Wrong file type or > 10MB | Inline error in upload card, no network call |
| API | File validation fails | 415 / 413 with error JSON, list re-renders without the new row |
| API | DB write succeeds, XADD fails | 500, but worker sweep recovers on next start |
| Worker | Job processing fails | `status=FAILED`, `error` populated, list shows retry button |
| Worker | Worker not running | Queue status shows queue_depth > 0 + processing = 0 |
| Live | Server missing GEMINI_API_KEY | Specific card on /b/live, link to fix |
| Live | WS disconnect mid-session | Auto-reconnect once, then show error |

---

## 9. Testing

### 9.1 API (Go)

- `internal/handlers/documents_test.go`:
 - Upload happy path → 201, row created, XADD called (mock Redis).
 - 10MB+1 file → 413.
 - `.exe` extension → 415.
 - DB succeeds, XADD throws → 500, row still in DB.
- `internal/handlers/queue_status_test.go`:
 - Five fields populated correctly; one failing field returns `null`, not 500.
- `internal/handlers/requeue_test.go`:
 - Re-queue READY → 409.
 - Re-queue FAILED → 202, new job_id.

### 9.2 Worker (Go)

- `internal/worker/sweep_test.go`:
 - Mock DB with 2 PENDING rows (1 with `last_queued_at` > 5min ago, 1 fresh) → only the old one re-queued.
- Integration: upload a real MD file via the API, run worker one iteration, assert document is `READY` with chunk_count > 0. Uses testcontainers for Redis + a real DuckDB file.

### 9.3 UI (TS)

- Pure function tests (Vitest):
 - `file-validator.test.ts`: accepts MD/TXT/HTML/PDF/DOCX under 10MB; rejects everything else.
 - `voice-activity.test.ts`: `computeRMS` returns expected values for known buffers.
- Component tests (Testing Library, per `apps/ui/AGENTS.md` — components stay thin):
 - `upload-card.test.tsx`: drop a valid file → fires onUpload. Drop an invalid one → shows error, no onUpload.
 - `document-row.test.tsx`: render with `status=READY, chunk_count=12` → shows "12 chunks".
 - `queue-status-strip.test.tsx`: render with all five fields null → renders five "—" placeholders.
- One Playwright E2E (golden path): dev server up, navigate to `/b/knowledge`, drop a small MD file, poll until `READY`, screenshot the final list.

### 9.4 Coverage target

- 100% on pure functions (file-validator, voice-activity, queue-status shape, payload serializer).
- ≥ 80% on handlers and components.
- E2E covers the user-visible happy path on the new pages.

---

## 10. File Layout (delta)

**API new files:**
- `apps/api/internal/handlers/documents.go` (POST + GET + requeue)
- `apps/api/internal/handlers/queue_status.go`
- `apps/api/internal/handlers/documents_test.go`
- `apps/api/internal/handlers/queue_status_test.go`
- `apps/migrations/005_add_last_queued_at.sql`

**API modified files:**
- `apps/api/internal/config/config.go` (add DocumentsDir, UploadMaxBytes)
- `apps/api/internal/server/routes.go` (mount the three new routes)
- `apps/api/internal/server/server.go` (wire dependencies)
- `docker-compose.yml` (shared documents volume)

**Worker new files:**
- `apps/worker/internal/worker/sweep.go`
- `apps/worker/internal/worker/sweep_test.go`

**Worker modified files:**
- `apps/worker/cmd/worker/main.go` (call sweep on startup)
- `apps/worker/internal/config/config.go` (add DocumentsDir)
- `apps/worker/internal/rag/document/parse.go` (read from file path, not base64)
- `apps/worker/internal/worker/processors/rag.go` (new payload shape)

**UI new files:** (see §7.2)

**UI modified files:**
- `apps/ui/src/components/ui/live-session-panel.tsx` (env var, mute, real orb, error card)
- `apps/ui/src/pages/console/index.tsx` (remove LiveSessionPanel)
- `apps/ui/src/app/router.tsx` (or equivalent) — add the two new routes
- `apps/ui/src/labels/common.ts` (add new strings)

---

## 11. Out of Scope (PoC)

- Document deletion UI (the table already supports `DELETE`; no UI for it yet)
- Multi-tenant document scoping
- Re-ingestion / version history
- SSE / WebSocket for live updates (polling is the contract)
- Pause/resume of the queue from the UI
- Audio recording in the upload card (voice notes) — separate spec
- Live transcript persistence (transcripts die with the session)
- Browser notifications when a doc is ready

---

## 12. Decision Log

| Date | Decision | Rationale |
|---|---|---|
| 2026-06-12 | UI-first spec order | User picked UI-then-API. API contracts designed to match screens, not the reverse. |
| 2026-06-12 | `last_queued_at` column on documents | Lets the worker sweep detect orphaned PENDING rows without a separate outbox table. |
| 2026-06-12 | Worker startup sweep, no retry button | Self-healing recovery beats user-driven recovery for a PoC. Retry button comes later if needed. |
| 2026-06-12 | 10MB upload cap | Enough for any demo document; bounded to keep job payloads small. |
| 2026-06-12 | 1.5s polling | Feels live without hammering the API. Pauses on tab blur. |
| 2026-06-12 | File-on-disk payload, not base64 in Redis | DuckDB and Gemini handle 10MB fine over the wire, but base64 in Redis bloats the stream and the job payload. Path-based is simpler. |
| 2026-06-12 | `GET /queue/status` returns nulls, not 500s | Partial data is more useful than no data. UI renders "—" for missing fields. |
| 2026-06-12 | Orb pulse driven by real RMS, not connection state | A "listening" pulse that fires during silence is dishonest UI. |
| 2026-06-12 | Mute sends a control message to server | Lets the server pause STT so we don't bill for known-muted audio. Client-side mute alone is a PoC shortcut. |
| 2026-06-12 | `VITE_WS_URL` env var, not a config file | One line in `.env.local` is enough; no need for a runtime config system at this stage. |
| 2026-06-12 | One spec, one shot, fully tested | User direction. All §9 tests must pass before merge. |

---

## 13. Open Questions (deferred, not blocking)

- When the upload finishes, do we want a brief "uploading" badge on the row for the gap between XHR end and worker pickup? Or is `PENDING → INGESTING` fast enough? (Probably fast enough; revisit if it feels laggy.)
- Should the queue-status strip be its own card or embedded in the upload card's header? Spec puts it as a separate strip; can merge if the empty state gets crowded.
- Should `/b/live` support text-only mode (no mic) for environments where the user can't grant mic permission? Yes — §2.4 says "denied-mic user can still type" — but the text-input UI is a follow-up, not in this spec.
