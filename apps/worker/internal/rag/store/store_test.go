package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/deemwar/live-api/apps/worker/internal/models"
)

// newTestStore creates a fresh DuckDB-backed store with the test schema applied.
// The store is closed automatically when the test ends.
func newTestStore(t *testing.T) *duckDBStore {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	s, err := NewDuckDBStore(dbPath)
	if err != nil {
		t.Fatalf("NewDuckDBStore: %v", err)
	}
	if err := createTestSchema(s); err != nil {
		t.Fatalf("createTestSchema: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// createTestSchema creates the tables required by the store without goose.
// Mirrors apps/migrations/ but kept inline for test isolation.
func createTestSchema(s *duckDBStore) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS documents (
			id TEXT PRIMARY KEY,
			source_name TEXT NOT NULL,
			source_type TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'PENDING',
			chunk_count INTEGER NOT NULL DEFAULT 0,
			error TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS parent_chunks (
			id TEXT PRIMARY KEY,
			document_id TEXT NOT NULL,
			content TEXT NOT NULL,
			token_count INTEGER NOT NULL,
			position INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (document_id) REFERENCES documents(id)
		)`,
		`CREATE TABLE IF NOT EXISTS child_chunks (
			id TEXT PRIMARY KEY,
			parent_id TEXT NOT NULL,
			content TEXT NOT NULL,
			token_count INTEGER NOT NULL,
			position INTEGER NOT NULL,
			embedding DOUBLE[] NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (parent_id) REFERENCES parent_chunks(id)
		)`,
	}
	for _, q := range stmts {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func TestNewDuckDBStore_OpensValidDB(t *testing.T) {
	dir := t.TempDir()
	s, err := NewDuckDBStore(filepath.Join(dir, "x.db"))
	if err != nil {
		t.Fatalf("NewDuckDBStore: %v", err)
	}
	defer s.Close()
}

func TestNewDuckDBStoreWithDB(t *testing.T) {
	s := newTestStore(t)
	relDB, vecDB := NewDuckDBStoreWithDB(s.db)
	if relDB == nil {
		t.Fatal("expected non-nil RelationalDBClient")
	}
	if vecDB == nil {
		t.Fatal("expected non-nil VectorDBClient")
	}
}

func TestSaveAndGetDocument(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	doc := models.Document{
		ID: "doc1",
		SourceName: "test.md",
		SourceType: "markdown",
		Status: models.DocStatusPending,
	}
	if err := s.SaveDocument(ctx, doc); err != nil {
		t.Fatalf("SaveDocument: %v", err)
	}
	got, err := s.GetDocument(ctx, "doc1")
	if err != nil {
		t.Fatalf("GetDocument: %v", err)
	}
	if got.ID != "doc1" {
		t.Errorf("ID: got %q", got.ID)
	}
	if got.SourceName != "test.md" {
		t.Errorf("SourceName: got %q", got.SourceName)
	}
	if got.Status != models.DocStatusPending {
		t.Errorf("Status: got %q", got.Status)
	}
	if got.ChunkCount != 0 {
		t.Errorf("ChunkCount: got %d", got.ChunkCount)
	}
}

func TestGetDocument_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.GetDocument(context.Background(), "missing")
	if err == nil {
		t.Error("expected error for missing doc")
	}
}

func TestUpdateDocumentStatus(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if err := s.SaveDocument(ctx, models.Document{ID: "d1", SourceName: "x", SourceType: "markdown", Status: models.DocStatusPending}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateDocumentStatus(ctx, "d1", models.DocStatusIngesting, ""); err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetDocument(ctx, "d1")
	if got.Status != models.DocStatusIngesting {
		t.Errorf("Status: got %q", got.Status)
	}
}

func TestUpdateDocumentStatus_WithError(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if err := s.SaveDocument(ctx, models.Document{ID: "d1", SourceName: "x", SourceType: "markdown", Status: models.DocStatusPending}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateDocumentStatus(ctx, "d1", models.DocStatusFailed, "bad things"); err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetDocument(ctx, "d1")
	if got.Status != models.DocStatusFailed {
		t.Errorf("Status: got %q", got.Status)
	}
	if got.Error != "bad things" {
		t.Errorf("Error: got %q", got.Error)
	}
}

func TestUpdateDocumentStatus_MissingDoc(t *testing.T) {
	s := newTestStore(t)
	err := s.UpdateDocumentStatus(context.Background(), "missing", models.DocStatusReady, "")
	if err == nil {
		t.Error("expected error for missing doc")
	}
}

func TestIncrementChunkCount(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if err := s.SaveDocument(ctx, models.Document{ID: "d1", SourceName: "x", SourceType: "markdown", Status: models.DocStatusPending}); err != nil {
		t.Fatal(err)
	}
	if err := s.IncrementChunkCount(ctx, "d1", 3); err != nil {
		t.Fatal(err)
	}
	if err := s.IncrementChunkCount(ctx, "d1", 2); err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetDocument(ctx, "d1")
	if got.ChunkCount != 5 {
		t.Errorf("ChunkCount: got %d want 5", got.ChunkCount)
	}
}

func TestIncrementChunkCount_MissingDoc(t *testing.T) {
	s := newTestStore(t)
	err := s.IncrementChunkCount(context.Background(), "missing", 1)
	if err == nil {
		t.Error("expected error for missing doc")
	}
}

func TestSaveParentAndGetByIDs(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if err := s.SaveDocument(ctx, models.Document{ID: "d1", SourceName: "x", SourceType: "markdown", Status: models.DocStatusPending}); err != nil {
		t.Fatal(err)
	}
	parents := []models.ParentChunk{
		{ID: "p1", DocumentID: "d1", Content: "hello", TokenCount: 5, Position: 0},
		{ID: "p2", DocumentID: "d1", Content: "world", TokenCount: 5, Position: 1},
	}
	if err := s.SaveParents(ctx, parents); err != nil {
		t.Fatalf("SaveParents: %v", err)
	}
	got, err := s.GetParentsByIDs(ctx, []string{"p1", "p2"})
	if err != nil {
		t.Fatalf("GetParentsByIDs: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d parents, want 2", len(got))
	}
}

func TestGetParentsByIDs_Empty(t *testing.T) {
	s := newTestStore(t)
	got, err := s.GetParentsByIDs(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestSaveParents_Empty(t *testing.T) {
	s := newTestStore(t)
	if err := s.SaveParents(context.Background(), nil); err != nil {
		t.Errorf("empty SaveParents: %v", err)
	}
}

func TestSaveChildAndSearchTopK(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if err := s.SaveDocument(ctx, models.Document{ID: "d1", SourceName: "x", SourceType: "markdown", Status: models.DocStatusPending}); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveParent(ctx, models.ParentChunk{ID: "p1", DocumentID: "d1", Content: "parent", TokenCount: 5, Position: 0}); err != nil {
		t.Fatal(err)
	}

	// Three children: one identical to query, one orthogonal, one opposite.
	queryVec := []float32{1, 0, 0}
	children := []models.ChildChunk{
		{ID: "c1", ParentID: "p1", Content: "match", TokenCount: 1, Position: 0, Embedding: queryVec},
		{ID: "c2", ParentID: "p1", Content: "orth", TokenCount: 1, Position: 1, Embedding: []float32{0, 1, 0}},
		{ID: "c3", ParentID: "p1", Content: "opp", TokenCount: 1, Position: 2, Embedding: []float32{-1, 0, 0}},
	}
	if err := s.SaveChildren(ctx, children); err != nil {
		t.Fatalf("SaveChildren: %v", err)
	}

	got, err := s.SearchTopK(ctx, queryVec, 2)
	if err != nil {
		t.Fatalf("SearchTopK: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 results, got %d", len(got))
	}
	if got[0].ID != "c1" {
		t.Errorf("expected c1 first (closest), got %s", got[0].ID)
	}
	if got[1].ID != "c2" {
		t.Errorf("expected c2 second (orthogonal), got %s", got[1].ID)
	}
}

func TestSaveChildren_Empty(t *testing.T) {
	s := newTestStore(t)
	if err := s.SaveChildren(context.Background(), nil); err != nil {
		t.Errorf("empty SaveChildren: %v", err)
	}
}

func TestSearchTopK_Empty(t *testing.T) {
	s := newTestStore(t)
	got, err := s.SearchTopK(context.Background(), []float32{1, 0}, 5)
	if err != nil {
		t.Fatalf("SearchTopK: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 results from empty table, got %d", len(got))
	}
}

func TestSaveChild_Single(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if err := s.SaveDocument(ctx, models.Document{ID: "d1", SourceName: "x", SourceType: "markdown", Status: models.DocStatusPending}); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveParent(ctx, models.ParentChunk{ID: "p1", DocumentID: "d1", Content: "p", TokenCount: 1, Position: 0}); err != nil {
		t.Fatal(err)
	}
	c := models.ChildChunk{ID: "c1", ParentID: "p1", Content: "x", TokenCount: 1, Position: 0, Embedding: []float32{1}}
	if err := s.SaveChild(ctx, c); err != nil {
		t.Fatalf("SaveChild: %v", err)
	}
}

func TestClose(t *testing.T) {
	dir := t.TempDir()
	s, err := NewDuckDBStore(filepath.Join(dir, "x.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}
