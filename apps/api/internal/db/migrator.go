package db

import (
	"context"
	"database/sql"
	"io/fs"
	"os"

	"github.com/pressly/goose/v3"
	gdb "github.com/pressly/goose/v3/database"
)

// Migrate runs all pending migrations from the given directory path.
func Migrate(conn *sql.DB, migrationsPath string) error {
	return MigrateWithFS(conn, os.DirFS(migrationsPath))
}

// MigrateWithFS runs all pending migrations from the given FS.
// Accepts any fs.FS — use for tests with fstest.MapFS.
func MigrateWithFS(conn *sql.DB, fsys fs.FS) error {
	p, err := goose.NewProvider(gdb.DialectCustom, conn, fsys, goose.WithStore(duckdbStore{}))
	if err != nil {
		return err
	}
	_, err = p.Up(context.Background())
	return err
}
