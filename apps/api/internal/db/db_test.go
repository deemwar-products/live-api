package db

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenRunsMigrations(t *testing.T) {
	database, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer database.Close()

	tables := []string{"organizations", "documents", "document_chunks"}
	for _, table := range tables {
		var count int
		if err := database.QueryRow(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ?`, table).Scan(&count); err != nil {
			t.Fatalf("query table %s: %v", table, err)
		}
		if count != 1 {
			t.Fatalf("expected table %s to exist, got count %d", table, count)
		}
	}
}

func TestOpenCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "rag.db")
	database, err := Open(path)
	if err != nil {
		t.Fatalf("Open with nested dir returned error: %v", err)
	}
	defer database.Close()
	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		t.Fatalf("expected dir to be created: %v", err)
	}
}

func TestOpenMkdirError(t *testing.T) {
	// /dev/null is a file, so creating a subdir under it fails MkdirAll.
	_, err := Open("/dev/null/bad/db.duckdb")
	if err == nil {
		t.Fatal("expected error for bad dir path")
	}
}

func TestMigrateOnClosedDB(t *testing.T) {
	database, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	database.Close()
	if err := database.migrate(); err == nil {
		t.Fatal("expected migrate to fail on closed db")
	}
}

func TestOpenParseError(t *testing.T) {
	if _, err := Open("://bad-dsn"); err == nil {
		t.Fatal("expected invalid DSN error")
	}
}

func TestOpenMigrationError(t *testing.T) {
	origOpenDuckDB := openDuckDB
	origRunMigrate := runMigrate
	defer func() {
		openDuckDB = origOpenDuckDB
		runMigrate = origRunMigrate
	}()

	openDuckDB = func(string) (*sql.DB, error) {
		return sql.Open("duckdb", "memory")
	}
	runMigrate = func(*DB) error {
		return errors.New("boom")
	}

	if _, err := Open("memory"); err == nil {
		t.Fatal("expected migrate failure")
	}
}
