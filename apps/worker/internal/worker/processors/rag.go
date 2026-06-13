package processors

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/deemwar/live-api/apps/worker/internal/queue"
	"github.com/deemwar/live-api/apps/worker/internal/rag/chunking"
	"github.com/deemwar/live-api/apps/worker/internal/rag/fetcher"
	"github.com/deemwar/live-api/apps/worker/internal/rag/store"
)

// RAGProcessor handles INGEST_DOCUMENT jobs.
// It runs the full ingestion pipeline: fetch → chunk → embed → persist.
type RAGProcessor struct {
	db      *sql.DB
	relDB   store.RelationalDBClient
	vecDB   store.VectorDBClient
	chunker *chunking.ChunkingEngine
	fetcher fetcher.Fetcher
}

// NewRAGProcessor wires the processor with its dependencies.
func NewRAGProcessor(
	db *sql.DB,
	relDB store.RelationalDBClient,
	vecDB store.VectorDBClient,
	chunker *chunking.ChunkingEngine,
	f fetcher.Fetcher,
) *RAGProcessor {
	return &RAGProcessor{
		db:      db,
		relDB:   relDB,
		vecDB:   vecDB,
		chunker: chunker,
		fetcher: f,
	}
}

type ingestDocumentPayload struct {
	DocumentID string `json:"document_id"`
	SourceName string `json:"source_name"`
	SourceType string `json:"source_type"`
}

func (p *RAGProcessor) Process(ctx context.Context, msg queue.Message) error {
	var raw string
	if err := p.db.QueryRowContext(ctx, `SELECT payload FROM jobs WHERE id = ?`, msg.JobID).Scan(&raw); err != nil {
		return fmt.Errorf("fetch job payload %s: %w", msg.JobID, err)
	}

	var pl ingestDocumentPayload
	if err := json.Unmarshal([]byte(raw), &pl); err != nil {
		return fmt.Errorf("parse INGEST_DOCUMENT payload: %w", err)
	}

	if err := p.relDB.UpdateDocumentStatus(ctx, pl.DocumentID, "INGESTING", ""); err != nil {
		return fmt.Errorf("mark INGESTING %s: %w", pl.DocumentID, err)
	}

	content, err := p.fetcher.Fetch(ctx, pl.SourceName, pl.SourceType)
	if err != nil {
		_ = p.relDB.UpdateDocumentStatus(ctx, pl.DocumentID, "FAILED", err.Error())
		return fmt.Errorf("fetch content %s: %w", pl.SourceName, err)
	}

	parents, children, err := p.chunker.IngestDocument(ctx, pl.DocumentID, content)
	if err != nil {
		_ = p.relDB.UpdateDocumentStatus(ctx, pl.DocumentID, "FAILED", err.Error())
		return fmt.Errorf("chunk %s: %w", pl.DocumentID, err)
	}

	if err := p.relDB.SaveParents(ctx, parents); err != nil {
		_ = p.relDB.UpdateDocumentStatus(ctx, pl.DocumentID, "FAILED", err.Error())
		return fmt.Errorf("save parents %s: %w", pl.DocumentID, err)
	}

	if err := p.vecDB.SaveChildren(ctx, children); err != nil {
		_ = p.relDB.UpdateDocumentStatus(ctx, pl.DocumentID, "FAILED", err.Error())
		return fmt.Errorf("save children %s: %w", pl.DocumentID, err)
	}

	if err := p.relDB.IncrementChunkCount(ctx, pl.DocumentID, len(parents)); err != nil {
		return fmt.Errorf("increment chunk count %s: %w", pl.DocumentID, err)
	}

	return p.relDB.UpdateDocumentStatus(ctx, pl.DocumentID, "READY", "")
}
