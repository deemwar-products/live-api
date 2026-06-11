package db

import (
	"crypto/sha256"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"

	"github.com/deemwar/live-api/apps/api/internal/logger"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

var migLog = logger.New("migrator")

var migrationFileRegex = regexp.MustCompile(`^(\d+)_(.+)\.sql$`)

type migration struct {
	version  string
	name     string
	content  string
	checksum string
}

func loadMigrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var migrations []migration
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := migrationFileRegex.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}

		content, err := migrationFS.ReadFile(path.Join("migrations", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}

		sum := sha256.Sum256(content)
		migrations = append(migrations, migration{
			version:  matches[1],
			name:     matches[2],
			content:  string(content),
			checksum: fmt.Sprintf("%x", sum),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}

func appliedVersions(conn *sql.DB) (map[string]string, error) {
	var exists bool
	err := conn.QueryRow("SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_name = 'schema_migrations'").Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("check schema_migrations exists: %w", err)
	}
	if !exists {
		return make(map[string]string), nil
	}

	rows, err := conn.Query("SELECT version, checksum FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("query schema_migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]string)
	for rows.Next() {
		var version, checksum string
		if err := rows.Scan(&version, &checksum); err != nil {
			return nil, err
		}
		applied[version] = checksum
	}

	return applied, nil
}

func Migrate(conn *sql.DB) error {
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	applied, err := appliedVersions(conn)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		stored, exists := applied[m.version]
		if exists {
			if stored != m.checksum {
				return fmt.Errorf("migration %s has been modified after being applied (stored=%s, current=%s)", m.version, stored, m.checksum)
			}
			migLog.Info("Skipping migration %s_%s (already applied)", m.version, m.name)
			continue
		}

		migLog.Info("Applying migration %s_%s", m.version, m.name)
		if _, err := conn.Exec(m.content); err != nil {
			return fmt.Errorf("apply %s: %w", m.version, err)
		}

		_, err = conn.Exec(
			"INSERT INTO schema_migrations (version, name, checksum) VALUES (?, ?, ?)",
			m.version, m.name, m.checksum,
		)
		if err != nil {
			return fmt.Errorf("record %s: %w", m.version, err)
		}
	}

	return nil
}
