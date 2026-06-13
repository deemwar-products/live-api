package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// QueueStatusHandler returns five KPIs about the RAG pipeline.
// All fields are independent: if one query fails, that field becomes null
// and the others are still returned. The UI renders "—" for nulls.
type QueueStatusHandler struct {
	DB *sql.DB
	Queue QueueProducer // used only to satisfy the interface for mockability
	StreamName string
}

func NewQueueStatusHandler(db *sql.DB, queue QueueProducer, streamName string) *QueueStatusHandler {
	return &QueueStatusHandler{DB: db, Queue: queue, StreamName: streamName}
}

type QueueStatus struct {
	QueueDepth *int64 `json:"queue_depth"`
	Processing *int `json:"processing"`
	ReadyToday *int `json:"ready_today"`
	FailedToday *int `json:"failed_today"`
	LastCompletedAt *time.Time `json:"last_completed_at"`
}

// Status handles GET /api/v1/queue/status
func (h *QueueStatusHandler) Status(c *gin.Context) {
	ctx := c.Request.Context()
	out := QueueStatus{}

	if h.Queue != nil {
		if n, err := h.queueDepth(ctx); err == nil {
			out.QueueDepth = &n
		} else {
			log.Warn("queue depth query failed: %v", err)
		}
	}

	if n, err := h.queryCount(ctx, `SELECT count(*) FROM jobs WHERE status = 'PROCESSING'`); err == nil {
		out.Processing = &n
	} else {
		log.Warn("processing count failed: %v", err)
	}

	if n, err := h.queryCount(ctx,
		`SELECT count(*) FROM documents WHERE status = 'READY' AND created_at >= CURRENT_DATE`); err == nil {
		out.ReadyToday = &n
	} else {
		log.Warn("ready_today count failed: %v", err)
	}

	if n, err := h.queryCount(ctx,
		`SELECT count(*) FROM jobs WHERE status = 'FAILED' AND updated_at >= CURRENT_DATE`); err == nil {
		out.FailedToday = &n
	} else {
		log.Warn("failed_today count failed: %v", err)
	}

	var ts sql.NullTime
	if err := h.DB.QueryRowContext(ctx,
		`SELECT max(updated_at) FROM jobs WHERE status IN ('COMPLETED','FAILED')`,
	).Scan(&ts); err == nil && ts.Valid {
		t := ts.Time
		out.LastCompletedAt = &t
	}

	c.JSON(http.StatusOK, out)
}

func (h *QueueStatusHandler) queryCount(ctx context.Context, q string) (int, error) {
	var n int
	if err := h.DB.QueryRowContext(ctx, q).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// queueDepth returns XLEN for the stream. If the queue producer doesn't
// expose this, callers will see queue_depth = null.
func (h *QueueStatusHandler) queueDepth(ctx context.Context) (int64, error) {
	if x, ok := h.Queue.(interface {
		XLen(ctx context.Context, stream string) (int64, error)
	}); ok {
		return x.XLen(ctx, h.StreamName)
	}
	return 0, nil
}
