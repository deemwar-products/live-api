-- +goose Up
-- Child chunks hold the smaller search vectors with embeddings.
-- Embeddings are stored as DuckDB DOUBLE[] (which the Go driver
-- supports natively via []float64; the store layer converts from
-- []float32 — the Gemini SDK's output type — at the boundary).

CREATE TABLE IF NOT EXISTS child_chunks (
 id TEXT PRIMARY KEY,
 parent_id TEXT NOT NULL,
 content TEXT NOT NULL,
 token_count INTEGER NOT NULL,
 position INTEGER NOT NULL,
 embedding DOUBLE[] NOT NULL,
 created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

 FOREIGN KEY (parent_id) REFERENCES parent_chunks(id)
);

CREATE INDEX IF NOT EXISTS idx_child_chunks_parent_id ON child_chunks(parent_id);

-- Similarity search is done via array_cosine_distance on the embedding column.
-- Example query:
-- SELECT id, parent_id, content, token_count, position,
-- array_cosine_distance(embedding, $query_embedding) AS distance
-- FROM child_chunks
-- ORDER BY distance ASC
-- LIMIT $k;

-- +goose Down
DROP INDEX IF EXISTS idx_child_chunks_parent_id;
DROP TABLE IF EXISTS child_chunks;