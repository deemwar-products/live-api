package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	gdb "github.com/pressly/goose/v3/database"
)

// duckdbStore implements database.Store for DuckDB.
// DuckDB does not support GENERATED ALWAYS AS IDENTITY or AUTOINCREMENT,
// so we create a SEQUENCE manually and use DEFAULT nextval(...).
type duckdbStore struct{}

const gooseTable = "goose_db_version"
const gooseSeq = "goose_db_version_id_seq"

func (duckdbStore) Tablename() string { return gooseTable }

func (duckdbStore) CreateVersionTable(ctx context.Context, db gdb.DBTxConn) error {
	if _, err := db.ExecContext(ctx, fmt.Sprintf(
		`CREATE SEQUENCE IF NOT EXISTS %s`, gooseSeq,
	)); err != nil {
		return fmt.Errorf("create goose sequence: %w", err)
	}
	_, err := db.ExecContext(ctx, fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id         BIGINT    NOT NULL DEFAULT nextval('%s') PRIMARY KEY,
			version_id BIGINT    NOT NULL,
			is_applied BOOLEAN   NOT NULL,
			tstamp     TIMESTAMP DEFAULT now()
		)`, gooseTable, gooseSeq,
	))
	return err
}

func (duckdbStore) Insert(ctx context.Context, db gdb.DBTxConn, req gdb.InsertRequest) error {
	_, err := db.ExecContext(ctx,
		fmt.Sprintf(`INSERT INTO %s (version_id, is_applied) VALUES (?, ?)`, gooseTable),
		req.Version, true,
	)
	return err
}

func (duckdbStore) Delete(ctx context.Context, db gdb.DBTxConn, version int64) error {
	_, err := db.ExecContext(ctx,
		fmt.Sprintf(`DELETE FROM %s WHERE version_id = ?`, gooseTable),
		version,
	)
	return err
}

func (duckdbStore) GetMigration(ctx context.Context, db gdb.DBTxConn, version int64) (*gdb.GetMigrationResult, error) {
	var ts time.Time
	var isApplied bool
	err := db.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT tstamp, is_applied FROM %s WHERE version_id = ? ORDER BY tstamp DESC LIMIT 1`, gooseTable),
		version,
	).Scan(&ts, &isApplied)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gdb.ErrVersionNotFound
		}
		return nil, err
	}
	return &gdb.GetMigrationResult{Timestamp: ts, IsApplied: isApplied}, nil
}

func (duckdbStore) GetLatestVersion(ctx context.Context, db gdb.DBTxConn) (int64, error) {
	var version sql.NullInt64
	err := db.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT MAX(version_id) FROM %s`, gooseTable),
	).Scan(&version)
	if err != nil {
		return 0, err
	}
	if !version.Valid {
		return 0, gdb.ErrVersionNotFound
	}
	return version.Int64, nil
}

func (duckdbStore) ListMigrations(ctx context.Context, db gdb.DBTxConn) ([]*gdb.ListMigrationsResult, error) {
	rows, err := db.QueryContext(ctx,
		fmt.Sprintf(`SELECT version_id, is_applied FROM %s ORDER BY id DESC`, gooseTable),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []*gdb.ListMigrationsResult
	for rows.Next() {
		var r gdb.ListMigrationsResult
		if err := rows.Scan(&r.Version, &r.IsApplied); err != nil {
			return nil, err
		}
		results = append(results, &r)
	}
	return results, rows.Err()
}
