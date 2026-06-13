-- +goose Up
-- Track when each PENDING document was last pushed to jobs.rag.
-- The worker startup sweep uses this to detect orphaned documents
-- (DB row written, but Redis XADD lost or never happened).
-- DuckDB does not support partial indexes, so we use a composite index
-- over (status, last_queued_at). The sweep predicate filters at query time.

ALTER TABLE documents ADD COLUMN last_queued_at TIMESTAMP NULL;

CREATE INDEX IF NOT EXISTS idx_documents_status_queued_at
 ON documents(status, last_queued_at);

-- +goose Down
DROP INDEX IF EXISTS idx_documents_status_queued_at;
ALTER TABLE documents DROP COLUMN IF EXISTS last_queued_at;
