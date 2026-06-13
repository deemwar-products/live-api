// Package store provides the relational and vector storage layer for RAG.
// Both interfaces are implemented against a shared DuckDB connection.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/deemwar/live-api/apps/worker/internal/models"

	_ "github.com/marcboeker/go-duckdb"
)

// RelationalDBClient persists documents and parent chunks.
type RelationalDBClient interface {
	// SaveDocument inserts or updates a document.
	SaveDocument(ctx context.Context, d models.Document) error
	// GetDocument fetches a document by ID.
	GetDocument(ctx context.Context, id string) (models.Document, error)
	// UpdateDocumentStatus updates a document's status and optionally sets error.
	UpdateDocumentStatus(ctx context.Context, id, status, err string) error
	// IncrementChunkCount adds to the chunk_count for a document.
	IncrementChunkCount(ctx context.Context, docID string, n int) error

	// SaveParent inserts a parent chunk.
	SaveParent(ctx context.Context, p models.ParentChunk) error
	// SaveParents inserts multiple parent chunks in one batch.
	SaveParents(ctx context.Context, ps []models.ParentChunk) error
	// GetParentsByIDs fetches multiple parent chunks by ID.
	GetParentsByIDs(ctx context.Context, ids []string) ([]models.ParentChunk, error)

	// Close closes the underlying connection.
	Close() error
}

// VectorDBClient persists child chunks with embeddings and supports similarity search.
type VectorDBClient interface {
	// SaveChild inserts a child chunk.
	SaveChild(ctx context.Context, c models.ChildChunk) error
	// SaveChildren inserts multiple child chunks in one batch.
	SaveChildren(ctx context.Context, cs []models.ChildChunk) error

	// SearchTopK finds the k nearest child chunks to the query embedding.
	SearchTopK(ctx context.Context, queryEmbedding []float32, k int) ([]models.ChildChunk, error)

	// Close closes the underlying connection.
	Close() error
}

// duckDBStore implements both interfaces against a single DuckDB connection.
// This is the only implementation, and it's designed for the worker (the sole writer).
type duckDBStore struct {
	db *sql.DB
}

// NewDuckDBStore opens a DuckDB connection at the given path.
// The path can be a local file path (e.g., "data/rag.db") or an S3 URL.
// Schema migrations are managed externally by the API via goose.
func NewDuckDBStore(path string) (*duckDBStore, error) {
	db, err := openDB(path)
	if err != nil {
		return nil, err
	}
	return &duckDBStore{db: db}, nil
}

// NewDuckDBStoreWithDB wraps an existing *sql.DB connection without running
// migrations. Use this when migrations are managed externally (e.g., by the API
// via goose) and the caller already holds the single write connection.
func NewDuckDBStoreWithDB(db *sql.DB) (RelationalDBClient, VectorDBClient) {
	s := &duckDBStore{db: db}
	return s, s
}

func openDB(path string) (*sql.DB, error) {
	// DuckDB Go driver handles both local paths and S3 URLs transparently.
	return sql.Open("duckdb", path)
}

// SaveDocument implements RelationalDBClient.
func (s *duckDBStore) SaveDocument(ctx context.Context, d models.Document) error {
	query := `
		INSERT INTO documents (id, source_name, source_type, status, chunk_count, error, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, DEFAULT, DEFAULT)
		ON CONFLICT (id) DO UPDATE SET
			source_name = EXCLUDED.source_name,
			source_type = EXCLUDED.source_type,
			chunk_count = EXCLUDED.chunk_count,
			error = EXCLUDED.error,
			status = EXCLUDED.status,
			updated_at = NOW()
	`
	_, err := s.db.ExecContext(ctx, query, d.ID, d.SourceName, d.SourceType, d.Status, d.ChunkCount, d.Error)
	return err
}

// GetDocument implements RelationalDBClient.
func (s *duckDBStore) GetDocument(ctx context.Context, id string) (models.Document, error) {
	query := `
		SELECT id, source_name, source_type, status, chunk_count, error
		FROM documents WHERE id = $1
	`
	var d models.Document
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &d.SourceName, &d.SourceType, &d.Status, &d.ChunkCount, &d.Error,
	)
	if err == sql.ErrNoRows {
		return d, fmt.Errorf("document not found: %s", id)
	}
	return d, err
}

// UpdateDocumentStatus implements RelationalDBClient.
func (s *duckDBStore) UpdateDocumentStatus(ctx context.Context, id, status, errMsg string) error {
	query := `
		UPDATE documents SET status = $2, error = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $1
	`
	result, execErr := s.db.ExecContext(ctx, query, id, status, errMsg)
	if execErr != nil {
		return execErr
	}
	rows, raErr := result.RowsAffected()
	if raErr != nil {
		return raErr
	}
	if rows == 0 {
		return fmt.Errorf("document not found: %s", id)
	}
	return nil
}

// IncrementChunkCount implements RelationalDBClient.
func (s *duckDBStore) IncrementChunkCount(ctx context.Context, docID string, n int) error {
	query := `
		UPDATE documents SET chunk_count = chunk_count + $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1
	`
	result, err := s.db.ExecContext(ctx, query, docID, n)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("document not found: %s", docID)
	}
	return nil
}

// SaveParent implements RelationalDBClient.
func (s *duckDBStore) SaveParent(ctx context.Context, p models.ParentChunk) error {
	query := `
		INSERT INTO parent_chunks (id, document_id, content, token_count, position, created_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE SET
			content = EXCLUDED.content,
			token_count = EXCLUDED.token_count,
			position = EXCLUDED.position
	`
	_, err := s.db.ExecContext(ctx, query, p.ID, p.DocumentID, p.Content, p.TokenCount, p.Position)
	return err
}

// SaveParents implements RelationalDBClient.
func (s *duckDBStore) SaveParents(ctx context.Context, ps []models.ParentChunk) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO parent_chunks (id, document_id, content, token_count, position, created_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE SET
			content = EXCLUDED.content,
			token_count = EXCLUDED.token_count,
			position = EXCLUDED.position
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range ps {
		if _, err := stmt.ExecContext(ctx, p.ID, p.DocumentID, p.Content, p.TokenCount, p.Position); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetParentsByIDs implements RelationalDBClient.
func (s *duckDBStore) GetParentsByIDs(ctx context.Context, ids []string) ([]models.ParentChunk, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	// Build "id IN ($1, $2, ...)" with one placeholder per id since the
	// driver doesn't support []string as VARCHAR[].
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}
	query := `
		SELECT id, document_id, content, token_count, position
		FROM parent_chunks WHERE id IN (` + strings.Join(placeholders, ",") + `)
	`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.ParentChunk
	for rows.Next() {
		var p models.ParentChunk
		if err := rows.Scan(&p.ID, &p.DocumentID, &p.Content, &p.TokenCount, &p.Position); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

// SaveChild implements VectorDBClient.
func (s *duckDBStore) SaveChild(ctx context.Context, c models.ChildChunk) error {
	// Use JSON array string and cast to DOUBLE[] — the driver supports string.
	query := `
		INSERT INTO child_chunks (id, parent_id, content, token_count, position, embedding, created_at)
		VALUES ($1, $2, $3, $4, $5, CAST($6 AS DOUBLE[]), DEFAULT)
		ON CONFLICT (id) DO UPDATE SET
			content = EXCLUDED.content,
			token_count = EXCLUDED.token_count,
			embedding = EXCLUDED.embedding
	`
	_, err := s.db.ExecContext(ctx, query, c.ID, c.ParentID, c.Content, c.TokenCount, c.Position, toJSONArray(c.Embedding))
	return err
}

// SaveChildren implements VectorDBClient.
func (s *duckDBStore) SaveChildren(ctx context.Context, cs []models.ChildChunk) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO child_chunks (id, parent_id, content, token_count, position, embedding, created_at)
		VALUES ($1, $2, $3, $4, $5, CAST($6 AS DOUBLE[]), DEFAULT)
		ON CONFLICT (id) DO UPDATE SET
			content = EXCLUDED.content,
			token_count = EXCLUDED.token_count,
			embedding = EXCLUDED.embedding
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, c := range cs {
		if _, err := stmt.ExecContext(ctx, c.ID, c.ParentID, c.Content, c.TokenCount, c.Position, toJSONArray(c.Embedding)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// SearchTopK implements VectorDBClient.
// Uses DuckDB's list_dot_product to compute cosine similarity manually,
// since array_cosine_distance has type resolution bugs with DOUBLE[] columns in DuckDB.
func (s *duckDBStore) SearchTopK(ctx context.Context, queryEmbedding []float32, k int) ([]models.ChildChunk, error) {
	// Build a subquery with the query vector and use list_dot_product
	// to compute cosine distance: 1 - cosine_similarity.
	// cosine_similarity = dot(a,b) / (||a|| * ||b||)
	literal := embeddingToDuckDBArray(queryEmbedding)
	query := `
		SELECT id, parent_id, content, token_count, position,
			1.0 - (
				list_dot_product(embedding, ` + literal + `) /
				(sqrt(list_dot_product(embedding, embedding)) * sqrt(list_dot_product(` + literal + `, ` + literal + `)))
			) AS distance
		FROM child_chunks
		ORDER BY distance ASC
		LIMIT ` + strconv.Itoa(k) + `
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.ChildChunk
	for rows.Next() {
		var c models.ChildChunk
		var dist float64
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Content, &c.TokenCount, &c.Position, &dist); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

// Close implements both RelationalDBClient and VectorDBClient.
func (s *duckDBStore) Close() error {
	return s.db.Close()
}

// toJSONArray formats a []float32 as a DuckDB-compatible JSON array
// string like "[1.0,2.0,3.0]". The driver has no native Go binding for
// primitive arrays, so we serialize at the storage boundary and the
// SQL casts the string back to DOUBLE[].
func toJSONArray(v []float32) string {
	if v == nil {
		return "[]"
	}
	var b strings.Builder
	b.Grow(len(v) * 8)
	b.WriteByte('[')
	for i, x := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(x), 'g', -1, 32))
	}
	b.WriteByte(']')
	return b.String()
}

// embeddingToDuckDBArray formats a []float32 as a DuckDB DOUBLE[] literal
// like "CAST([1.0, 2.0] AS DOUBLE[])" suitable for direct interpolation
// into SQL. The driver cannot bind primitive slices, so for read paths
// we inline the literal at the cost of one extra allocation per search.
func embeddingToDuckDBArray(v []float32) string {
	if v == nil {
		return "CAST(NULL AS DOUBLE[])"
	}
	var b strings.Builder
	b.Grow(len(v) * 10)
	b.WriteString("CAST([")
	for i, x := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		// FormatFloat with -1 precision strips trailing zeros, so for
		// whole numbers we manually append ".0" so DuckDB parses them
		// as DOUBLEs rather than integers.
		s := strconv.FormatFloat(float64(x), 'f', -1, 64)
		b.WriteString(s)
		if !strings.ContainsAny(s, ".eE") {
			b.WriteString(".0")
		}
	}
	b.WriteString("] AS DOUBLE[])")
	return b.String()
}