package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestOpenReadOnly_WhenPathValid_ThenReturnsConnection(t *testing.T) {
 dir := t.TempDir()
 path := filepath.Join(dir, "test.db")

 readWrite, err := OpenReadWrite(path)
 if err != nil {
 t.Fatalf("failed to create test db: %v", err)
 }
 if _, err := readWrite.Exec("CREATE TABLE foo (id INTEGER)"); err != nil {
 t.Fatalf("failed to create table: %v", err)
 }
 readWrite.Close()

 conn, err := OpenReadOnly(path)
 if err != nil {
 t.Fatalf("OpenReadOnly failed: %v", err)
 }
 defer conn.Close()

 var count int
 if err := conn.QueryRow("SELECT COUNT(*) FROM foo").Scan(&count); err != nil {
 t.Errorf("read query failed: %v", err)
 }
}

func TestOpenReadOnly_WhenPathMissing_ThenReturnsError(t *testing.T) {
 dir := t.TempDir()
 missing := filepath.Join(dir, "does-not-exist.db")

 conn, err := OpenReadOnly(missing)
 if err == nil {
 conn.Close()
 t.Fatal("expected error for missing path, got nil")
 }
}

func TestOpenReadWrite_WhenPathInMissingDir_ThenCreatesDir(t *testing.T) {
 dir := t.TempDir()
 nested := filepath.Join(dir, "a", "b", "c", "test.db")

 conn, err := OpenReadWrite(nested)
 if err != nil {
 t.Fatalf("OpenReadWrite failed: %v", err)
 }
 defer conn.Close()

 if _, err := os.Stat(filepath.Dir(nested)); err != nil {
 t.Errorf("expected directory to exist: %v", err)
 }
}

func TestOpenReadWrite_WhenMkdirFails_ThenReturnsError(t *testing.T) {
 dir := t.TempDir()
 blocker := filepath.Join(dir, "blocker")
 if err := os.WriteFile(blocker, []byte("not a dir"), 0644); err != nil {
 t.Fatalf("setup: %v", err)
 }

 badPath := filepath.Join(blocker, "test.db")
 conn, err := OpenReadWrite(badPath)
 if err == nil {
 conn.Close()
 t.Fatal("expected error when mkdir fails, got nil")
 }
}

func TestOpenReadWrite_WhenPathIsInvalid_ThenReturnsError(t *testing.T) {
 badPath := "/dev/null/cannot/be/created/test.db"
 conn, err := OpenReadWrite(badPath)
 if err == nil {
 conn.Close()
 t.Fatal("expected error for invalid path, got nil")
 }
}

func TestOpenReadWrite_WhenPathValid_ThenReturnsConnection(t *testing.T) {
 dir := t.TempDir()
 path := filepath.Join(dir, "rw.db")

 conn, err := OpenReadWrite(path)
 if err != nil {
 t.Fatalf("OpenReadWrite failed: %v", err)
 }
 defer conn.Close()

 if _, err := conn.Exec("CREATE TABLE bar (id INTEGER, name TEXT)"); err != nil {
 t.Errorf("write query failed: %v", err)
 }
}

func TestMigrate_WhenFreshDatabase_ThenAppliesAllMigrations(t *testing.T) {
	conn := newTestDB(t)

	if err := Migrate(conn, "../../../migrations"); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	assertTableExists(t, conn, "goose_db_version")
	assertTableExists(t, conn, "jobs")
	assertTableExists(t, conn, "documents")
	assertTableExists(t, conn, "parent_chunks")
	assertTableExists(t, conn, "child_chunks")
}

func TestMigrate_WhenRunTwice_ThenIsIdempotent(t *testing.T) {
	conn := newTestDB(t)

	if err := Migrate(conn, "../../../migrations"); err != nil {
		t.Fatalf("first Migrate failed: %v", err)
	}
	if err := Migrate(conn, "../../../migrations"); err != nil {
		t.Fatalf("second Migrate failed: %v", err)
	}

	assertTableExists(t, conn, "jobs")
}

func TestMigrate_WhenConnIsClosed_ThenReturnsError(t *testing.T) {
	conn := newTestDB(t)
	conn.Close()

	err := Migrate(conn, "../../../migrations")
	if err == nil {
		t.Fatal("expected error on closed connection, got nil")
	}
}

func TestMigrateWithFS_WhenInvalidSQL_ThenReturnsError(t *testing.T) {
	conn := newTestDB(t)

	badFS := fstest.MapFS{
		"001_bad.sql": &fstest.MapFile{Data: []byte("-- +goose Up\nTHIS IS NOT VALID SQL AT ALL\n-- +goose Down\n")},
	}

	err := MigrateWithFS(conn, badFS)
	if err == nil {
		t.Fatal("expected error from invalid SQL, got nil")
	}
}

func newTestDB(t *testing.T) *sql.DB {
 t.Helper()
 dir := t.TempDir()
 path := filepath.Join(dir, "test.db")
 conn, err := OpenReadWrite(path)
 if err != nil {
 t.Fatalf("failed to open test db: %v", err)
 }
 t.Cleanup(func() { conn.Close() })
 return conn
}

func assertTableExists(t *testing.T, conn *sql.DB, table string) {
 t.Helper()
 var exists bool
 if err := conn.QueryRow(
 "SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_name = ?",
 table,
 ).Scan(&exists); err != nil {
 t.Fatalf("check table %s: %v", table, err)
 }
 if !exists {
 t.Errorf("expected table %s to exist", table)
 }
}

