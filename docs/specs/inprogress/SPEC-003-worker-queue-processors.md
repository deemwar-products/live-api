# Spec: Worker Queue & Processor Architecture

## Status
Approved — Sreyash Reddy, 2026-06-12

## Context

The worker is the sole DuckDB writer. It handles two distinct workloads:
- **RAG ingestion** — slow, heavy: parse → chunk → embed → write
- **DB writes** — fast, lightweight: status updates, inserts

If both share one queue, a long RAG job starves DB writes. Two isolated streams solve this.

---

## Decision

**Option A: Two Redis Streams, two goroutines.**

```
API
 ├─ XADD jobs.rag   ──► goroutine:rag   → parse, chunk, embed, write
 └─ XADD jobs.writes ──► goroutine:write → fast DB operations only
```

Each goroutine blocks on its own stream. Neither can starve the other.

---

## Stream Design

| Stream | Consumer Group | Goroutine | Job Types |
|---|---|---|---|
| `jobs.rag` | `workers.rag` | rag | `INGEST_DOCUMENT` |
| `jobs.writes` | `workers.writes` | write | `UPDATE_JOB_STATUS`, `UPDATE_DOCUMENT_STATUS` |

### Message format (Redis Stream fields)

Every message on both streams has two fields:

```
job_id  → UUID (matches the jobs table)
type    → job type string
```

The worker fetches the full payload from the `jobs` table using `job_id`. Redis only carries the routing key — not the full payload.

---

## Job Types & Payloads (stored in jobs.payload as JSON)

### `INGEST_DOCUMENT`
```json
{
  "document_id": "uuid",
  "source_name": "report.pdf",
  "source_type": "pdf"
}
```

### `UPDATE_JOB_STATUS`
```json
{
  "job_id": "uuid",
  "status": "COMPLETED|FAILED|PROCESSING",
  "error": "optional error string"
}
```

### `UPDATE_DOCUMENT_STATUS`
```json
{
  "document_id": "uuid",
  "status": "INGESTING|READY|FAILED",
  "error": "optional error string"
}
```

---

## File Structure

```
apps/worker/
├── cmd/worker/
│   └── main.go                      ← entry point, wires everything
└── internal/
    ├── config/
    │   └── config.go                ← add RagStreamName, WriteStreamName, group names
    ├── queue/
    │   ├── names.go                 ← stream/group name constants
    │   ├── producer.go              ← XADD helper (used by API and worker internally)
    │   └── consumer.go              ← XREADGROUP + XACK helpers
    └── worker/
        ├── worker.go                ← starts goroutines, handles shutdown
        └── processors/
            ├── processor.go         ← Processor interface + Message type
            ├── rag.go               ← handles INGEST_DOCUMENT
            └── db_write.go          ← handles UPDATE_JOB_STATUS, UPDATE_DOCUMENT_STATUS
```

---

## Processor Interface

```go
type Message struct {
    StreamID string            // Redis message ID (for XACK)
    JobID    string
    Type     string
}

type Processor interface {
    Process(ctx context.Context, msg Message) error
}
```

---

## Worker Loop (per goroutine)

```
loop:
  XREADGROUP stream group consumer BLOCK 5000ms COUNT 10
  for each message:
    fetch job from DuckDB by job_id
    update job status → PROCESSING
    call processor.Process(ctx, msg)
    on success: update job status → COMPLETED, XACK
    on error:   increment attempts
                if attempts >= MaxAttempts: update status → FAILED, XACK
                else: leave in PEL (Redis retries on next XAUTOCLAIM)
```

---

## Config Changes

Add to `config.go`:

```go
RagStreamName   string  // default: "jobs.rag"
WriteStreamName string  // default: "jobs.writes"
RagGroup        string  // default: "workers.rag"
WriteGroup      string  // default: "workers.writes"
```

Existing `StreamName` and `ConsumerGroup` fields are removed.

---

## Graceful Shutdown

Worker main listens on `context.Done()`. Each goroutine exits its loop when context is cancelled. In-flight jobs finish before exit (no forced kill mid-processing).

---

## Out of Scope (PoC)

- Dead letter queue
- XAUTOCLAIM (pending message redelivery on crash)
- Metrics endpoint (discussed, deferred)
- Multiple consumer instances
