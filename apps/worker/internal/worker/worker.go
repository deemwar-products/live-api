package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/deemwar/live-api/apps/worker/internal/config"
	"github.com/deemwar/live-api/apps/worker/internal/queue"
	"github.com/deemwar/live-api/apps/worker/internal/worker/processors"
)

// Worker runs two isolated goroutines — one per Redis stream.
type Worker struct {
	cfg       config.Config
	consumer  *queue.Consumer
	db        *sql.DB
	ragProc   processors.Processor
	writeProc processors.Processor
}

// New returns a Worker wired with its processors.
func New(
	cfg config.Config,
	consumer *queue.Consumer,
	db *sql.DB,
	ragProc processors.Processor,
	writeProc processors.Processor,
) *Worker {
	return &Worker{
		cfg:       cfg,
		consumer:  consumer,
		db:        db,
		ragProc:   ragProc,
		writeProc: writeProc,
	}
}

// Run starts both goroutines and blocks until ctx is cancelled.
// Both goroutines must exit before Run returns.
func (w *Worker) Run(ctx context.Context) error {
	if err := w.consumer.CreateGroup(ctx, w.cfg.RagStreamName, w.cfg.RagGroup); err != nil {
		return fmt.Errorf("create rag group: %w", err)
	}
	if err := w.consumer.CreateGroup(ctx, w.cfg.WriteStreamName, w.cfg.WriteGroup); err != nil {
		return fmt.Errorf("create write group: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		w.runLoop(ctx, w.cfg.RagStreamName, w.cfg.RagGroup, w.ragProc)
	}()

	go func() {
		defer wg.Done()
		w.runLoop(ctx, w.cfg.WriteStreamName, w.cfg.WriteGroup, w.writeProc)
	}()

	wg.Wait()
	return nil
}

// runLoop is the per-stream consumer loop. It exits when ctx is cancelled.
func (w *Worker) runLoop(ctx context.Context, stream, group string, proc processors.Processor) {
	blockDur := time.Duration(w.cfg.IdleBlockMS) * time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msgs, err := w.consumer.Read(ctx, stream, group, w.cfg.ConsumerName, blockDur)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("[worker] read error on %s: %v", stream, err)
			continue
		}

		for _, msg := range msgs {
			w.handle(ctx, stream, group, msg, proc)
		}
	}
}

// handle runs a single job through its full lifecycle.
func (w *Worker) handle(ctx context.Context, stream, group string, msg queue.Message, proc processors.Processor) {
	if err := w.setJobStatus(ctx, msg.JobID, "PROCESSING", ""); err != nil {
		log.Printf("[worker] set PROCESSING failed for %s: %v", msg.JobID, err)
	}

	if err := proc.Process(ctx, msg); err != nil {
		log.Printf("[worker] job %s (%s) failed: %v", msg.JobID, msg.Type, err)
		attempts := w.incrementAttempts(ctx, msg.JobID)
		if attempts >= w.cfg.MaxAttempts {
			_ = w.setJobStatus(ctx, msg.JobID, "FAILED", err.Error())
			if ackErr := w.consumer.Ack(ctx, stream, group, msg.StreamID); ackErr != nil {
				log.Printf("[worker] ack failed for %s: %v", msg.JobID, ackErr)
			}
		}
		// else: leave in PEL — Redis will redeliver on next XAUTOCLAIM
		return
	}

	_ = w.setJobStatus(ctx, msg.JobID, "COMPLETED", "")
	if err := w.consumer.Ack(ctx, stream, group, msg.StreamID); err != nil {
		log.Printf("[worker] ack failed for %s: %v", msg.JobID, err)
	}
}

func (w *Worker) setJobStatus(ctx context.Context, jobID, status, errMsg string) error {
	_, err := w.db.ExecContext(ctx,
		`UPDATE jobs SET status = ?, last_error = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, errMsg, jobID,
	)
	return err
}

func (w *Worker) incrementAttempts(ctx context.Context, jobID string) int {
	_, _ = w.db.ExecContext(ctx,
		`UPDATE jobs SET attempts = attempts + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		jobID,
	)
	var attempts int
	_ = w.db.QueryRowContext(ctx, `SELECT attempts FROM jobs WHERE id = ?`, jobID).Scan(&attempts)
	return attempts
}
