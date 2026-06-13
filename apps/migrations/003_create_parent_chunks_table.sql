-- +goose Up
-- Parent chunks hold the full semantic context for LLM consumption.

CREATE TABLE IF NOT EXISTS parent_chunks (
 id TEXT PRIMARY KEY,
 document_id TEXT NOT NULL,
 content TEXT NOT NULL,
 token_count INTEGER NOT NULL,
 position INTEGER NOT NULL,
 created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

 FOREIGN KEY (document_id) REFERENCES documents(id)
);

CREATE INDEX IF NOT EXISTS idx_parent_chunks_document_id ON parent_chunks(document_id);

-- +goose Down
DROP INDEX IF EXISTS idx_parent_chunks_document_id;
DROP TABLE IF EXISTS parent_chunks;