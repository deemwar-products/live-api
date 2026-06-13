package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/deemwar/live-api/apps/api/internal/db"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockQueue records XAdd calls without touching Redis.
type mockQueue struct {
	mu sync.Mutex
	calls []mockXAddCall
	fail bool
}

type mockXAddCall struct {
	Stream string
	Fields map[string]any
}

func (m *mockQueue) XAdd(_ context.Context, stream string, values ...any) (string, error) {
	if m.fail {
		return "", errMockQueue
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	fields := make(map[string]any, len(values)/2)
	for i := 0; i+1 < len(values); i += 2 {
		k, _ := values[i].(string)
		fields[k] = values[i+1]
	}
	m.calls = append(m.calls, mockXAddCall{Stream: stream, Fields: fields})
	return "1-0", nil
}

var errMockQueue = &mockQueueErr{msg: "simulated XADD failure"}

type mockQueueErr struct{ msg string }

func (e *mockQueueErr) Error() string { return e.msg }

func newTestServer(t *testing.T, q QueueProducer) (*gin.Engine, *sql.DB, string) {
	t.Helper()
	dir := t.TempDir()
	conn, err := db.OpenReadWrite(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	if err := db.Migrate(conn, "../../../migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	docDir := filepath.Join(dir, "documents")
	if err := os.MkdirAll(docDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	h := NewDocumentsHandler(conn, q, docDir, 1024*1024) // 1MB cap
	engine := gin.New()
	engine.POST("/documents", h.Upload)
	engine.GET("/documents", h.List)
	engine.GET("/documents/:id", h.Get)
	engine.POST("/documents/:id/requeue", h.Requeue)
	return engine, conn, docDir
}

func uploadRequest(t *testing.T, filename, content string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := io.Copy(fw, strings.NewReader(content)); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	mw.Close()
	req := httptest.NewRequest("POST", "/documents", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func TestUpload_HappyPath_Returns201(t *testing.T) {
	q := &mockQueue{}
	engine, conn, dir := newTestServer(t, q)
	defer os.RemoveAll(dir)

	req := uploadRequest(t, "hello.md", "# hello world")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}

	var doc Document
	if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if doc.Status != "PENDING" {
		t.Errorf("expected status PENDING, got %q", doc.Status)
	}
	if doc.SourceType != "md" {
		t.Errorf("expected source_type md, got %q", doc.SourceType)
	}
	if len(q.calls) != 1 {
		t.Errorf("expected 1 XADD, got %d", len(q.calls))
	}
	if q.calls[0].Stream != "jobs.rag" {
		t.Errorf("expected stream jobs.rag, got %q", q.calls[0].Stream)
	}

	// File should be on disk.
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Errorf("expected 1 file on disk, got %d (err=%v)", len(entries), err)
	}

	// Document row should have last_queued_at set.
	var lastQueued sql.NullTime
	if err := conn.QueryRow(`SELECT last_queued_at FROM documents WHERE id = ?`, doc.ID).Scan(&lastQueued); err != nil {
		t.Fatalf("query last_queued_at: %v", err)
	}
	if !lastQueued.Valid {
		t.Errorf("expected last_queued_at to be set after successful upload (got NULL)")
	}
}

func TestUpload_FileTooLarge_Returns413(t *testing.T) {
	q := &mockQueue{}
	engine, _, dir := newTestServer(t, q)
	defer os.RemoveAll(dir)

	// Create a 2MB file (cap is 1MB).
	big := strings.Repeat("a", 2*1024*1024)
	req := uploadRequest(t, "big.md", big)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", w.Code)
	}
	if len(q.calls) != 0 {
		t.Errorf("expected no XADD, got %d", len(q.calls))
	}
}

func TestUpload_UnsupportedExtension_Returns415(t *testing.T) {
	q := &mockQueue{}
	engine, _, dir := newTestServer(t, q)
	defer os.RemoveAll(dir)

	req := uploadRequest(t, "evil.exe", "MZ")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", w.Code)
	}
}

func TestUpload_XADDFails_Returns500_AndRowPersists(t *testing.T) {
	q := &mockQueue{fail: true}
	engine, conn, _ := newTestServer(t, q)

	req := uploadRequest(t, "hello.md", "# hi")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%s", w.Code, w.Body.String())
	}
	// Row should still be there for the worker sweep to recover.
	var n int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM documents WHERE source_name = 'hello.md'`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 document row, got %d", n)
	}
}

func TestList_EmptyDB_ReturnsEmptyItems(t *testing.T) {
	q := &mockQueue{}
	engine, _, _ := newTestServer(t, q)

	req := httptest.NewRequest("GET", "/documents", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body struct {
		Items []Document `json:"items"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Total != 0 || len(body.Items) != 0 {
		t.Errorf("expected empty list, got total=%d items=%d", body.Total, len(body.Items))
	}
}

func TestGet_NonExistent_Returns404(t *testing.T) {
	q := &mockQueue{}
	engine, _, _ := newTestServer(t, q)

	req := httptest.NewRequest("GET", "/documents/missing", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestRequeue_ReadyDocument_Returns409(t *testing.T) {
	q := &mockQueue{}
	engine, conn, _ := newTestServer(t, q)

	if _, err := conn.Exec(
		`INSERT INTO documents (id, source_name, source_type, status) VALUES ('d1', 'a.md', 'md', 'READY')`,
	); err != nil {
		t.Fatalf("seed: %v", err)
	}

	req := httptest.NewRequest("POST", "/documents/d1/requeue", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestRequeue_FailedDocument_Returns202(t *testing.T) {
	q := &mockQueue{}
	engine, conn, _ := newTestServer(t, q)

	if _, err := conn.Exec(
		`INSERT INTO documents (id, source_name, source_type, status, error) VALUES ('d2', 'b.md', 'md', 'FAILED', 'bad pdf')`,
	); err != nil {
		t.Fatalf("seed: %v", err)
	}

	req := httptest.NewRequest("POST", "/documents/d2/requeue", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body=%s", w.Code, w.Body.String())
	}
	if len(q.calls) != 1 {
		t.Errorf("expected 1 XADD on requeue, got %d", len(q.calls))
	}
	var status string
	if err := conn.QueryRow(`SELECT status FROM documents WHERE id = 'd2'`).Scan(&status); err != nil {
		t.Fatalf("query: %v", err)
	}
	if status != "PENDING" {
		t.Errorf("expected PENDING after requeue, got %q", status)
	}
}
