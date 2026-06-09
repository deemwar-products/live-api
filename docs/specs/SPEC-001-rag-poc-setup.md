# SPEC-001: RAG POC Setup

**Date:** 2026-06-06  
**Status:** APPROVED  
**Branch:** rag-poc

---

## 1. Project Structure

```
apps/
├── api/
│   ├── cmd/server/main.go
│   ├── internal/config/
│   ├── internal/db/
│   ├── internal/handler/
│   ├── internal/service/
│   ├── internal/middleware/
│   └── pkg/rag/
└── worker/
    ├── cmd/worker/main.go
    └── internal/queue/
        └── processor/
shared/pkg/
Taskfile.yaml
docker-compose.yml
Dockerfile
```

---

## 2. Two Containers

| Container | Port | Responsibility |
|-----------|------|----------------|
| api | 8080 | HTTP + WebSocket |
| worker | - | Queue consumer |

---

## 3. Docker Compose

```yaml
services:
  api:
    build: .
    command: ./api
    ports: ["8080:8080"]
    environment:
      - MODE=api
      - REDIS_HOST=redis
    depends_on: [redis]

  worker:
    build: .
    command: ./worker
    environment:
      - MODE=worker
      - REDIS_HOST=redis
    depends_on: [redis]

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
```

---

## 4. Taskfile.yaml (Root)

```yaml
tasks:
  build: go build ./...
  docker:up: docker compose up --build
  docker:down: docker compose down
  test: go test -v -race ./...
  test:coverage: go test -coverprofile=coverage.out ./...
```

---

## 5. DuckDB Database

```sql
CREATE TABLE organizations (
    id UUID PRIMARY KEY,
    name VARCHAR(255),
    slug VARCHAR(100) UNIQUE,
    settings JSON,
    created_at TIMESTAMP
);

CREATE TABLE documents (
    id UUID PRIMARY KEY,
    org_id UUID,
    name VARCHAR(255),
    content TEXT,
    status VARCHAR(50),
    chunk_count INTEGER,
    created_at TIMESTAMP
);

CREATE TABLE document_chunks (
    id UUID PRIMARY KEY,
    document_id UUID,
    org_id UUID,
    chunk_text TEXT,
    embedding JSON,
    created_at TIMESTAMP
);
```

---

## 6. Redis Streams (Queue)

### Job Structure

```json
{
  "id": "job_abc123",
  "type": "process_document",
  "org_id": "org_xxx",
  "status": "queued",
  "created_at": "timestamp"
}
```

### Worker (No Polling!)

```go
// Blocks until job arrives - zero CPU when idle
streams, _ := redis.XReadGroup(
    "workers", "worker1",
    streams.StreamArgs{
        Streams: []string{"job_stream", "$"},
        Block:   0,
    },
)
```

### Status Events

```go
// On start
redis.XAdd("job_events", "*", map[string]interface{}{
    "job_id": "doc_123", "status": "processing",
})

// On complete
redis.XAdd("job_events", "*", map[string]interface{}{
    "job_id": "doc_123", "status": "completed",
})
```

---

## 7. HTTP API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /health | Health check |
| GET | /api/v1/queue/jobs | List jobs |
| GET | /api/v1/queue/jobs/:id | Job status |
| GET | /api/v1/queue/stats | Job counts |
| POST | /api/v1/queue/jobs | Create job |
| DELETE | /api/v1/queue/jobs/:id | Cancel job |

---

## 8. Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.22 |
| Web Framework | Gin |
| WebSocket | gorilla/websocket |
| Database | DuckDB |
| Queue | Redis Streams |
| Container | Docker |

---

## 9. Test Requirements

- **100% code coverage** for all packages
- Unit tests for handlers, services, queue
- Integration tests for Docker Compose
- Test commands in Taskfile.yaml

---

## Decision Log

| Date | Decision |
|------|----------|
| 2026-06-06 | Two-container architecture |
| 2026-06-06 | Redis Streams (no polling) |
| 2026-06-06 | DuckDB with JSON embeddings |
| 2026-06-06 | HTTP APIs for queue management |

---

## Deferred

- Chunking strategy
- Embedding model
- RAG pipeline
- MCP integration