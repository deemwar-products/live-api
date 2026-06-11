# SPEC-002: RAG Worker + Ingestion Pipeline

**Date:** 2026-06-11
**Status:** DRAFT
**Branch:** rag-poc
**Supersedes:** worker scaffold portion of SPEC-001

---

## 1. Goal

Build the `apps/worker` service that:
- Consumes jobs from Redis Stream `jobs.stream`
- Executes a real RAG ingestion pipeline: format detection → markdown conversion → semantic chunking → parent-child decomposition → embedding generation → DuckDB storage
- Updates `jobs` row status throughout (`QUEUED` → `PROCESSING` → `COMPLETED` / `FAILED`)
- Records `attempts` and `last_error`
- Acks the message on completion (or after 3 failed attempts)
- Handles SIGTERM gracefully (finishes in-flight job, exits)

The API and worker **share the same DuckDB file** (local in dev, S3 in dev/prod environments). DuckDB's single-writer constraint is solved by making the worker the **sole writer** to the chunk tables. The API only reads.

---

## 2. Worker Service Structure

```
apps/worker/
├── cmd/worker/main.go # Entry point: load config, open DB, start consumer
├── Dockerfile
├── Taskfile.yaml
├── .env # Base config (committed)
├── .env.local # Local overrides (gitignored)
├── .env.dev.example # Dev server template (committed)
├── .env.prod.example # Production template (committed)
├── go.mod
└── internal/
 ├── config/
 │ └── config.go # PORT unused; REDIS_HOST/PORT, DB_PATH, GEMINI_API_KEY, CONSUMER_GROUP
 ├── worker/
 │ ├── worker.go # Consumer loop, signal handling, lifecycle
 │ ├── processor.go # Processor interface + registry
 │ └── processors/
 │ └── index_document.go # Orchestrates the full ingestion pipeline
 ├── rag/
 │ ├── chunking/
 │ │ ├── chunking.go # ChunkingEngine, semantic split, sliding window
 │ │ └── chunking_test.go
 │ ├── embedding/
 │ │ ├── client.go # EmbeddingClient interface
 │ │ ├── gemini.go # Gemini impl using google.golang.org/genai
 │ │ └── embedding_test.go
 │ ├── store/
 │ │ ├── relational.go # RelationalDBClient interface + DuckDB impl
 │ │ ├── vector.go # VectorDBClient interface + DuckDB impl
 │ │ └── store_test.go
 │ ├── fetcher/
 │ │ ├── fetcher.go # FetchHighAccuracyContext
 │ │ └── fetcher_test.go
 │ └── document/
 │ ├── detect.go # Format detection (magic bytes + extension)
 │ ├── markdown.go # passthrough (already markdown)
 │ ├── text.go # passthrough (plain text → wrap in markdown)
 │ ├── html.go # HTML → MD (using gomarkdown or md-to-html lib)
 │ ├── pdf.go # PDF → MD (using github.com/ledongthuc/pdf)
 │ ├── docx.go # DOCX → MD (using github.com/nguyenthenguyen/docx)
 │ └── document_test.go
 └── queue/
 ├── consumer.go # Redis Streams XReadGroup wrapper
 ├── consumer_test.go
 ├── jobs.go # Job struct + serialization
 └── jobs_test.go
```

---

## 3. RAG Engine Design

### 3.1 ChunkingEngine

Pure functions, no I/O. Wraps the `tiktoken-go` `cl100k_base` encoder.

```go
type ChunkingEngine struct {
 encoder *tiktoken.Tiktoken
 targetChildTokens int // e.g., 150
 overlapTokens int // e.g., 30
}

func (e *ChunkingEngine) SplitDocumentSemantically(ctx context.Context, rawMarkdown string) ([]string, error)
func (e *ChunkingEngine) IngestDocument(ctx context.Context, docID, rawMarkdown string) ([]ParentChunk, []ChildChunk, error)
```

**Semantic split algorithm:**
1. Split markdown into sentences (regex-based, sentence boundary detection).
2. Call `EmbeddingClient.EmbedBatch` on all sentences in one batched Gemini call.
3. Compute cosine distance between adjacent sentence pairs.
4. Threshold = 75th percentile of distances (or fixed `0.5` for POC simplicity).
5. Wherever distance > threshold, cut a new semantic block.
6. Return ordered list of semantic blocks.

**Parent-child decomposition:**
- For each semantic block:
 - If `tokenCount(block) <= targetChildTokens` → **identity mapping**: 1 parent + 1 child with identical content.
 - Else → sliding window with `overlapTokens` overlap, producing N children. Parent holds the full block.

### 3.2 EmbeddingClient

```go
type EmbeddingClient interface {
 Embed(ctx context.Context, text string) ([]float32, error)
 EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}
```

**Gemini implementation** uses `google.golang.org/genai`:
```go
client, _ := genai.NewClient(ctx, &genai.ClientConfig{
 APIKey: apiKey,
 Backend: genai.BackendGeminiAPI,
})
// Client.Models.Embed or batched equivalent
```

Model: `text-embedding-004` (768-dim, default).

### 3.3 Store interfaces

```go
type RelationalDBClient interface {
 SaveParent(ctx context.Context, p ParentChunk) error
 SaveParents(ctx context.Context, ps []ParentChunk) error
 GetParentsByIDs(ctx context.Context, ids []string) ([]ParentChunk, error)
 SaveDocument(ctx context.Context, d Document) error
 GetDocument(ctx context.Context, id string) (Document, error)
 IncrementChunkCount(ctx context.Context, docID string, n int) error
}

type VectorDBClient interface {
 SaveChild(ctx context.Context, c ChildChunk) error
 SaveChildren(ctx context.Context, cs []ChildChunk) error
 SearchTopK(ctx context.Context, queryEmbedding []float32, k int) ([]ChildChunk, error)
}
```

**Both implemented against DuckDB** using the same `*sql.DB` connection the worker already has. (No separate vector DB — DuckDB's `array_cosine_distance` does what we need.)

### 3.4 Fetcher (used by API in SPEC-003)

```go
type Fetcher struct {
 vector VectorDBClient
 relational RelationalDBClient
 embedding EmbeddingClient
}

func (f *Fetcher) FetchHighAccuracyContext(ctx context.Context, userQuery string) (string, error)
```

**Pipeline:**
1. Embed user query.
2. `vector.SearchTopK(ctx, queryEmbedding, 12)` — retrieve 12 nearest children.
3. Dedupe `parent_id`s via `map[string]struct{}`.
4. `relational.GetParentsByIDs(ctx, uniqueIDs)` — batch fetch.
5. Assemble context string with `---` dividers between parent blocks.

---

## 4. DuckDB Schema (new migrations)

### 004_create_documents_table.sql

```sql
CREATE TABLE IF NOT EXISTS documents (
 id TEXT PRIMARY KEY,
 source_name TEXT NOT NULL,
 source_type TEXT NOT NULL,
 status TEXT NOT NULL DEFAULT 'PENDING'
 CHECK (status IN ('PENDING', 'INGESTING', 'READY', 'FAILED')),
 chunk_count INTEGER NOT NULL DEFAULT 0,
 error TEXT,
 created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
 updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 005_create_parent_chunks_table.sql

```sql
CREATE TABLE IF NOT EXISTS parent_chunks (
 id TEXT PRIMARY KEY,
 document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
 content TEXT NOT NULL,
 token_count INTEGER NOT NULL,
 position INTEGER NOT NULL,
 created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_parent_chunks_document_id ON parent_chunks(document_id);
```

### 006_create_child_chunks_table.sql

```sql
CREATE TABLE IF NOT EXISTS child_chunks (
 id TEXT PRIMARY KEY,
 parent_id TEXT NOT NULL REFERENCES parent_chunks(id) ON DELETE CASCADE,
 content TEXT NOT NULL,
 token_count INTEGER NOT NULL,
 position INTEGER NOT NULL,
 embedding FLOAT[] NOT NULL,
 created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_child_chunks_parent_id ON child_chunks(parent_id);
```

**Similarity search query** (used by `VectorDBClient.SearchTopK`):
```sql
SELECT id, parent_id, content, token_count, position,
 array_cosine_distance(embedding, $1::FLOAT[]) AS distance
FROM child_chunks
ORDER BY distance ASC
LIMIT $2;
```

---

## 5. Worker Consumer Loop

### 5.1 Lifecycle

1. On startup: `XGROUP CREATE jobs.stream workers $ MKSTREAM` (idempotent — ignore BUSYGROUP).
2. Trap SIGTERM / SIGINT.
3. Consumer name = `worker-{hostname}-{pid}` (Suren's call from huddle).
4. Loop:
 - `XREADGROUP GROUP workers {consumerName} BLOCK 5000 COUNT 1 STREAMS jobs.stream >`
 - For each message:
 - `processor.Process(ctx, job)` — runs the actual work.
 - `XACK jobs.stream workers {messageID}` on success or after final failure.
 - On error before `XACK`: message stays pending, can be reclaimed via `XAUTOCLAIM` on next startup.
5. On SIGTERM: finish in-flight `Process` call, exit cleanly.

### 5.2 Retry policy

- Max 3 attempts.
- On each failure: increment `jobs.attempts`, write `last_error`, set status back to `QUEUED` (so retry), and **ACK the message** to prevent re-delivery. The job is back in the queue because it's still `QUEUED` in DuckDB; the next worker pickup will see it. *(Alternative: don't ACK and let Redis re-deliver. Simpler reasoning: ACK + re-insert job into stream. But this complicates delivery guarantees. We go with ACK + leave as QUEUED; a small periodic re-enqueue can sweep stale QUEUED jobs.)*
- On 3rd failure: status = `FAILED`, ACK, do not requeue.

### 5.3 Processor interface

```go
type Processor interface {
 Type() string // matches jobs.type
 Process(ctx context.Context, job *Job) error
}

type Registry struct {
 procs map[string]Processor
}

func (r *Registry) Register(p Processor) { ... }
func (r *Registry) Get(typeName string) (Processor, bool) { ... }
```

For POC, one processor: `index_document`.

### 5.4 index_document processor

```go
type IndexDocumentProcessor struct {
 db *sql.DB
 embedding embedding.EmbeddingClient
 store *store.CombinedStore // wraps relational + vector on the same DuckDB
 chunking *chunking.ChunkingEngine
 document *document.Converter
}

func (p *IndexDocumentProcessor) Process(ctx context.Context, job *Job) error {
 // 1. Parse payload
 var payload IndexDocumentPayload
 if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil { return err }
 // payload = { document_id, source_name, source_type, raw_bytes (base64) }
 // 2. Update document status to INGESTING
 p.store.UpdateDocumentStatus(ctx, payload.DocumentID, "INGESTING")
 // 3. Detect & convert to markdown
 markdown, err := p.document.Convert(ctx, payload.RawBytes, payload.SourceType)
 if err != nil { return err }
 // 4. Chunk + embed
 parents, children, err := p.chunking.IngestDocument(ctx, payload.DocumentID, markdown)
 if err != nil { return err }
 // 5. Save to DuckDB
 if err := p.store.SaveParents(ctx, parents); err != nil { return err }
 if err := p.store.SaveChildren(ctx, children); err != nil { return err }
 if err := p.store.IncrementChunkCount(ctx, payload.DocumentID, len(parents)); err != nil { return err }
 // 6. Mark document READY
 return p.store.UpdateDocumentStatus(ctx, payload.DocumentID, "READY")
}
```

---

## 6. Configuration

Worker reads:
- `DB_PATH` (default: `data/rag.db`; in Docker: `/var/lib/live-api/rag.db`; on S3: `s3://bucket/rag.db`)
- `REDIS_HOST`, `REDIS_PORT`
- `GEMINI_API_KEY` (or `GOOGLE_API_KEY` — same key)
- `CONSUMER_GROUP` (default: `workers`)
- `STREAM_NAME` (default: `jobs.stream`)
- `MAX_ATTEMPTS` (default: 3)
- `TARGET_CHILD_TOKENS` (default: 150)
- `OVERLAP_TOKENS` (default: 30)
- `EMBEDDING_MODEL` (default: `text-embedding-004`)
- `EMBEDDING_DIM` (default: 768)

The API will need a `POST /documents` endpoint that creates the `documents` row, the `jobs` row, and pushes to Redis. That's the only write the API does to chunks-related tables — and it does it before the worker takes over.

---

## 7. API Endpoints (this SPEC)

Even though SPEC-003 covers the Live API bridge, the API needs basic REST endpoints to be useful:

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check (already exists) |
| `POST` | `/api/v1/documents` | Upload document. Body: multipart/form-data with file. Creates `documents` row + `jobs` row + Redis push. Returns `document_id`. |
| `GET` | `/api/v1/documents/{id}` | Fetch document + status + chunk count. |

The `POST /chat` endpoint is **NOT** in this spec — it's part of SPEC-003 (Live API bridge). For SPEC-002, retrieval via the Fetcher lives in the `internal/rag` package but isn't exposed as a public API yet.

---

## 8. Tech Stack

| Component | Library |
|-----------|---------|
| Gemini SDK | `google.golang.org/genai` |
| Tokenizer | `github.com/pkoukk/tiktoken-go` (`cl100k_base`) |
| PDF parser | `github.com/ledongthuc/pdf` |
| DOCX parser | `github.com/nguyenthenguyen/docx` |
| HTML→MD | `github.com/gomarkdown/markdown` (or just `html.UnescapeString` + light cleanup for POC) |
| Redis client | `github.com/redis/go-redis/v9` |
| DuckDB | `github.com/marcboeker/go-duckdb` (already in go.mod) |
| UUID | `github.com/google/uuid` (already in go.mod) |
| Web framework | `github.com/gin-gonic/gin` (matches API) |

---

## 9. Docker Compose Update

```yaml
services:
 api:
 build: context: apps/api
 command: ./api
 ports: ["8080:8080"]
 environment:
 REDIS_HOST: redis
 DB_PATH: /data/rag.db
 GEMINI_API_KEY: ${GEMINI_API_KEY}
 volumes:
 - rag_data:/data
 depends_on: [redis]

 worker:
 build: context: apps/worker
 command: ./worker
 environment:
 REDIS_HOST: redis
 DB_PATH: /data/rag.db
 GEMINI_API_KEY: ${GEMINI_API_KEY}
 volumes:
 - rag_data:/data
 depends_on: [redis, api] # api runs migrations first

 redis:
 image: redis:7-alpine
 ports: ["6379:6379"]

volumes:
 rag_data:
```

The api runs migrations on startup, so the worker can `depends_on: [api]` (with a `condition: service_completed_successfully` if we add healthchecks).

---

## 10. Test Requirements (100% coverage)

- `internal/worker/`: consumer loop, processor registry, signal handling
- `internal/rag/chunking/`: semantic split, sliding window, identity mapping shortcut, cosine math
- `internal/rag/embedding/`: Gemini client (with mock HTTP server using `httptest`)
- `internal/rag/store/`: relational + vector DuckDB ops
- `internal/rag/fetcher/`: dedupe logic, context assembly
- `internal/rag/document/`: format detection, each converter
- `internal/queue/`: consumer + jobs serialization
- Integration test: push a real markdown doc → run worker one iteration → assert document is READY and chunks exist. Uses testcontainers for Redis + local DuckDB file.

---

## 11. Out of Scope (deferred to SPEC-003 or later)

- WebSocket bridge to Gemini Live API
- RAG-as-a-tool exposed to Gemini
- Multi-tenant document scoping
- Document deletion / re-ingestion
- Real embedding model selection (we hardcode `text-embedding-004`)
- Rate limiting / Gemini API quota handling

---

## Decision Log

| Date | Decision |
|------|----------|
| 2026-06-11 | DuckDB stores embeddings (FLOAT[]), uses `array_cosine_distance` — no separate vector DB |
| 2026-06-11 | Worker is the sole writer to `documents`, `parent_chunks`, `child_chunks` tables |
| 2026-06-11 | API only reads from these tables; writes go through Redis → worker |
| 2026-06-11 | Multi-format input: MD, TXT, HTML, PDF, DOCX |
| 2026-06-11 | Format detection by magic bytes + filename extension |
| 2026-06-11 | Gemini `text-embedding-004` for embeddings (same API key as chat) |
| 2026-06-11 | Tokenizer: `tiktoken-go` with `cl100k_base` |
| 2026-06-11 | Retry policy: 3 attempts, then FAILED + ACK |
| 2026-06-11 | Consumer name: `worker-{hostname}-{pid}` |
| 2026-06-11 | Shared DuckDB file: local in dev, S3 (`s3://bucket/rag.db`) in dev-server / prod |
| 2026-06-11 | POST /chat endpoint deferred to SPEC-003 |
