-- +goose Up
-- Tracks uploaded documents through the ingestion lifecycle.
-- Status: PENDING | INGESTING | READY | FAILED

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

CREATE INDEX IF NOT EXISTS idx_documents_created_at ON documents(created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_documents_created_at;
DROP TABLE IF EXISTS documents;