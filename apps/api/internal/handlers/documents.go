package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/deemwar/live-api/apps/api/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var log = logger.New("handlers")

// RagStreamName is the Redis stream the RAG worker consumes from.
const RagStreamName = "jobs.rag"

// AllowedExtensions is the set of source_type values the API will accept.
var AllowedExtensions = map[string]string{
	"md": "md", "markdown": "md",
	"txt": "txt",
	"html": "html", "htm": "html",
	"pdf": "pdf",
	"docx": "docx",
}

// QueueProducer is the minimal Redis interface needed for XADD.
// Defined here so tests can mock it without standing up Redis.
type QueueProducer interface {
	XAdd(ctx context.Context, stream string, values ...any) (string, error)
}

type RedisClient struct {
	C *redis.Client
}

func (r *RedisClient) XAdd(ctx context.Context, stream string, values ...any) (string, error) {
	return r.C.XAdd(ctx, &redis.XAddArgs{Stream: stream, Values: values}).Result()
}

func (r *RedisClient) XLen(ctx context.Context, stream string) (int64, error) {
	return r.C.XLen(ctx, stream).Result()
}

// DocumentsHandler holds the dependencies for document upload/list/requeue.
type DocumentsHandler struct {
	DB *sql.DB
	Queue QueueProducer
	DocumentsDir string
	UploadMaxBytes int
}

func NewDocumentsHandler(db *sql.DB, queue QueueProducer, documentsDir string, uploadMaxBytes int) *DocumentsHandler {
	return &DocumentsHandler{
		DB: db,
		Queue: queue,
		DocumentsDir: documentsDir,
		UploadMaxBytes: uploadMaxBytes,
	}
}

type Document struct {
	ID string `json:"id"`
	SourceName string `json:"source_name"`
	SourceType string `json:"source_type"`
	Status string `json:"status"`
	ChunkCount int `json:"chunk_count"`
	Error *string `json:"error"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Upload handles POST /api/v1/documents.
//
// 1. Reject if Content-Length > UploadMaxBytes -> 413.
// 2. Reject if extension is not in AllowedExtensions -> 415.
// 3. Stream the file to {DocumentsDir}/{doc_id}.{ext}.
// 4. INSERT document + job in one transaction.
// 5. XADD jobs.rag.
// 6. UPDATE documents.last_queued_at = now().
func (h *DocumentsHandler) Upload(c *gin.Context) {
	if c.Request.ContentLength > int64(h.UploadMaxBytes) {
		writeError(c, http.StatusRequestEntityTooLarge, "payload_too_large", fmt.Sprintf("file exceeds %d bytes", h.UploadMaxBytes))
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		writeError(c, http.StatusBadRequest, "missing_file", "form field 'file' is required")
		return
	}
	if fileHeader.Size > int64(h.UploadMaxBytes) {
		writeError(c, http.StatusRequestEntityTooLarge, "payload_too_large", fmt.Sprintf("file exceeds %d bytes", h.UploadMaxBytes))
		return
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(fileHeader.Filename), "."))
	sourceType, ok := AllowedExtensions[ext]
	if !ok {
		writeError(c, http.StatusUnsupportedMediaType, "unsupported_type", "allowed types: md, txt, html, pdf, docx")
		return
	}

	docID := uuid.NewString()
	jobID := uuid.NewString()
	storedPath := filepath.Join(h.DocumentsDir, docID+"."+ext)

	if err := os.MkdirAll(h.DocumentsDir, 0755); err != nil {
		writeError(c, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}

	// Stream the upload to disk (bounded by UploadMaxBytes).
	dst, err := os.Create(storedPath)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}
	src, err := fileHeader.Open()
	if err != nil {
		dst.Close()
		os.Remove(storedPath)
		writeError(c, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}
	if _, err := io.Copy(dst, io.LimitReader(src, int64(h.UploadMaxBytes)+1)); err != nil {
		dst.Close()
		src.Close()
		os.Remove(storedPath)
		writeError(c, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}
	dst.Close()
	src.Close()

	payload, _ := json.Marshal(map[string]string{
		"document_id": docID,
		"source_name": fileHeader.Filename,
		"source_type": sourceType,
		"file_path": storedPath,
	})

	// Pin to a single conn: DuckDB is single-writer and database/sql's pool
	// routes each call to a different conn, causing "Duplicate key" on UPDATE
	// because the new conn sees a stale snapshot.
	conn, err := h.DB.Conn(c.Request.Context())
	if err != nil {
		os.Remove(storedPath)
		writeError(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	defer conn.Close()

	tx, err := conn.BeginTx(c.Request.Context(), nil)
	if err != nil {
		os.Remove(storedPath)
		writeError(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	now := time.Now().UTC()
	if _, err := tx.ExecContext(c.Request.Context(),
		`INSERT INTO documents (id, source_name, source_type, status, chunk_count, created_at, updated_at) VALUES (?, ?, ?, 'PENDING', 0, ?, ?)`,
		docID, fileHeader.Filename, sourceType, now, now,
	); err != nil {
		tx.Rollback()
		os.Remove(storedPath)
		writeError(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	if _, err := tx.ExecContext(c.Request.Context(),
		`INSERT INTO jobs (id, type, payload, status, attempts, created_at, updated_at) VALUES (?, ?, ?, 'QUEUED', 0, ?, ?)`,
		jobID, "INGEST_DOCUMENT", string(payload), now, now,
	); err != nil {
		tx.Rollback()
		os.Remove(storedPath)
		writeError(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	if err := tx.Commit(); err != nil {
		os.Remove(storedPath)
		writeError(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	// Push to Redis. If this fails, the worker startup sweep recovers.
	streamID, xerr := h.Queue.XAdd(c.Request.Context(), RagStreamName, "job_id", jobID, "type", "INGEST_DOCUMENT")
	if xerr != nil {
		log.Warn("XADD failed for doc %s (sweep will recover): %v", docID, xerr)
		writeError(c, http.StatusInternalServerError, "queue_error", "queued locally; worker will pick up on next startup")
		return
	}
	log.Info("Queued doc %s as job %s (stream %s)", docID, jobID, streamID)

	// Same conn → no snapshot drift.
	if _, err := conn.ExecContext(c.Request.Context(),
		`UPDATE documents SET last_queued_at = CURRENT_TIMESTAMP WHERE id = ?`, docID,
	); err != nil {
		log.Warn("failed to set last_queued_at for %s: %v", docID, err)
	}

	var doc Document
	if err := conn.QueryRowContext(c.Request.Context(),
		`SELECT id, source_name, source_type, status, chunk_count, error, created_at, updated_at FROM documents WHERE id = ?`,
		docID,
	).Scan(&doc.ID, &doc.SourceName, &doc.SourceType, &doc.Status, &doc.ChunkCount, &doc.Error, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
		writeError(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	c.JSON(http.StatusCreated, doc)
}

// List handles GET /api/v1/documents?limit=&offset=
func (h *DocumentsHandler) List(c *gin.Context) {
	limit := atoiDefault(c.Query("limit"), 50)
	if limit > 200 {
		limit = 200
	}
	offset := atoiDefault(c.Query("offset"), 0)

	rows, err := h.DB.QueryContext(c.Request.Context(),
		`SELECT id, source_name, source_type, status, chunk_count, error, created_at, updated_at
 FROM documents ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	defer rows.Close()

	items := make([]Document, 0)
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.SourceName, &d.SourceType, &d.Status, &d.ChunkCount, &d.Error, &d.CreatedAt, &d.UpdatedAt); err != nil {
			writeError(c, http.StatusInternalServerError, "db_error", err.Error())
			return
		}
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		writeError(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	var total int
	if err := h.DB.QueryRowContext(c.Request.Context(), `SELECT COUNT(*) FROM documents`).Scan(&total); err != nil {
		writeError(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

// Get handles GET /api/v1/documents/:id
func (h *DocumentsHandler) Get(c *gin.Context) {
	id := c.Param("id")
	doc, err := loadDocument(c.Request.Context(), h.DB, id)
	if errors.Is(err, errNotFound) {
		writeError(c, http.StatusNotFound, "not_found", "document not found")
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	c.JSON(http.StatusOK, doc)
}

// Requeue handles POST /api/v1/documents/:id/requeue
func (h *DocumentsHandler) Requeue(c *gin.Context) {
	id := c.Param("id")
	doc, err := loadDocument(c.Request.Context(), h.DB, id)
	if errors.Is(err, errNotFound) {
		writeError(c, http.StatusNotFound, "not_found", "document not found")
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	if doc.Status == "READY" {
		writeError(c, http.StatusConflict, "already_ready", "document is already READY")
		return
	}

	conn, err := h.DB.Conn(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	defer conn.Close()

	// Re-derive source_type and file_path from the document row.
	var sourceType string
	if err := conn.QueryRowContext(c.Request.Context(), `SELECT source_type FROM documents WHERE id = ?`, id).Scan(&sourceType); err != nil {
		writeError(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	ext := sourceType
	if sourceType == "html" {
		ext = "html"
	}
	filePath := filepath.Join(h.DocumentsDir, id+"."+ext)

	jobID := uuid.NewString()
	payload, _ := json.Marshal(map[string]string{
		"document_id": id,
		"source_name": doc.SourceName,
		"source_type": sourceType,
		"file_path": filePath,
	})

	now := time.Now().UTC()
	tx, err := conn.BeginTx(c.Request.Context(), nil)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	if _, err := tx.ExecContext(c.Request.Context(),
		`UPDATE documents SET status = 'PENDING', last_queued_at = NULL, error = NULL, updated_at = ? WHERE id = ?`,
		now, id,
	); err != nil {
		tx.Rollback()
		writeError(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	if _, err := tx.ExecContext(c.Request.Context(),
		`INSERT INTO jobs (id, type, payload, status, attempts, created_at, updated_at) VALUES (?, ?, ?, 'QUEUED', 0, ?, ?)`,
		jobID, "INGEST_DOCUMENT", string(payload), now, now,
	); err != nil {
		tx.Rollback()
		writeError(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	streamID, xerr := h.Queue.XAdd(c.Request.Context(), RagStreamName, "job_id", jobID, "type", "INGEST_DOCUMENT")
	if xerr != nil {
		log.Warn("XADD failed on requeue for doc %s (sweep will recover): %v", id, xerr)
		writeError(c, http.StatusInternalServerError, "queue_error", "queued locally; worker will pick up on next startup")
		return
	}
	log.Info("Re-queued doc %s as job %s (stream %s)", id, jobID, streamID)

	if _, err := conn.ExecContext(c.Request.Context(),
		`UPDATE documents SET last_queued_at = CURRENT_TIMESTAMP WHERE id = ?`, id,
	); err != nil {
		log.Warn("failed to set last_queued_at on requeue for %s: %v", id, err)
	}

	c.JSON(http.StatusAccepted, gin.H{"id": id, "status": "PENDING"})
}

// ── helpers ────────────────────────────────────────────────────────────────

var errNotFound = errors.New("document not found")

func loadDocument(ctx context.Context, db *sql.DB, id string) (Document, error) {
	var d Document
	err := db.QueryRowContext(ctx,
		`SELECT id, source_name, source_type, status, chunk_count, error, created_at, updated_at FROM documents WHERE id = ?`,
		id,
	).Scan(&d.ID, &d.SourceName, &d.SourceType, &d.Status, &d.ChunkCount, &d.Error, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Document{}, errNotFound
	}
	return d, err
}

func writeError(c *gin.Context, status int, code, msg string) {
	c.JSON(status, gin.H{"error": msg, "code": code})
}

func atoiDefault(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return fallback
	}
	return n
}
