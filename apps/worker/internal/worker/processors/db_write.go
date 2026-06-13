package processors

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/deemwar/live-api/apps/worker/internal/queue"
)

// DBWriteProcessor handles fast DB operations: job and document status updates.
type DBWriteProcessor struct {
	db *sql.DB
}

// NewDBWriteProcessor returns a processor backed by the given DuckDB connection.
func NewDBWriteProcessor(db *sql.DB) *DBWriteProcessor {
	return &DBWriteProcessor{db: db}
}

func (p *DBWriteProcessor) Process(ctx context.Context, msg queue.Message) error {
	var payload string
	err := p.db.QueryRowContext(ctx, `SELECT payload FROM jobs WHERE id = ?`, msg.JobID).Scan(&payload)
	if err != nil {
		return fmt.Errorf("fetch job payload %s: %w", msg.JobID, err)
	}

	switch msg.Type {
	case queue.TypeUpdateJobStatus:
		return p.updateJobStatus(ctx, payload)
	case queue.TypeUpdateDocumentStatus:
		return p.updateDocumentStatus(ctx, payload)
	default:
		return fmt.Errorf("unknown write job type: %s", msg.Type)
	}
}

type updateJobStatusPayload struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func (p *DBWriteProcessor) updateJobStatus(ctx context.Context, raw string) error {
	var pl updateJobStatusPayload
	if err := json.Unmarshal([]byte(raw), &pl); err != nil {
		return fmt.Errorf("parse UPDATE_JOB_STATUS payload: %w", err)
	}
	_, err := p.db.ExecContext(ctx,
		`UPDATE jobs SET status = ?, last_error = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		pl.Status, pl.Error, pl.JobID,
	)
	return err
}

type updateDocumentStatusPayload struct {
	DocumentID string `json:"document_id"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
}

func (p *DBWriteProcessor) updateDocumentStatus(ctx context.Context, raw string) error {
	var pl updateDocumentStatusPayload
	if err := json.Unmarshal([]byte(raw), &pl); err != nil {
		return fmt.Errorf("parse UPDATE_DOCUMENT_STATUS payload: %w", err)
	}
	_, err := p.db.ExecContext(ctx,
		`UPDATE documents SET status = ?, error = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		pl.Status, pl.Error, pl.DocumentID,
	)
	return err
}
