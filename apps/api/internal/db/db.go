package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/marcboeker/go-duckdb"
)

type DB struct {
	*sql.DB
}

var (
	mkdirAll   = os.MkdirAll
	openDuckDB = func(path string) (*sql.DB, error) { return sql.Open("duckdb", path) }
	runMigrate = func(d *DB) error { return d.migrate() }
)

func Open(path string) (*DB, error) {
	if path == ":memory:" {
		path = "memory"
	}

	if path != ":memory:" {
		dir := filepath.Dir(path)
		if dir != "." {
			if err := mkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("mkdir db dir: %w", err)
			}
		}
	}

	conn, err := openDuckDB(path)
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
	}
	d := &DB{conn}
	if err := runMigrate(d); err != nil {
		conn.Close()
		return nil, err
	}
	return d, nil
}

func (d *DB) migrate() error {
	_, err := d.Exec(`
		CREATE TABLE IF NOT EXISTS organizations (
			id VARCHAR PRIMARY KEY,
			name VARCHAR(255),
			slug VARCHAR(100) UNIQUE,
			settings JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS documents (
			id VARCHAR PRIMARY KEY,
			org_id VARCHAR,
			name VARCHAR(255),
			content TEXT,
			status VARCHAR(50),
			chunk_count INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS document_chunks (
			id VARCHAR PRIMARY KEY,
			document_id VARCHAR,
			org_id VARCHAR,
			chunk_text TEXT,
			embedding JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("migrate duckdb: %w", err)
	}
	return nil
}
