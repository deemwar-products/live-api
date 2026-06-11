package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/deemwar/live-api/apps/api/internal/logger"

	_ "github.com/marcboeker/go-duckdb"
)

var log = logger.New("db")

func OpenReadOnly(path string) (*sql.DB, error) {
	destination := fmt.Sprintf("file:%s?access_mode=read_only", path)
	log.Info("Opening read-only connection to %s", destination)
	return sql.Open("duckdb", destination)
}

func OpenReadWrite(path string) (*sql.DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db directory %s: %w", dir, err)
	}
	log.Info("Opening read-write connection to %s", path)
	return sql.Open("duckdb", path)
}
