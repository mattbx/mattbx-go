package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type migration struct {
	version int
	name    string
	sql     string
}

// Migrate applies every migration newer than the recorded schema version, each
// in its own transaction. It is safe to call on every boot.
//
// Keep migrations additive: Disco's zero-downtime deploy briefly runs the old
// and new containers against the same volume, so the previous version must
// still work against the new schema. For a destructive change, use Disco's
// "hook:deploy:start:before" service instead.
func Migrate(sqlDB *sql.DB) error {
	if _, err := sqlDB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		name       TEXT NOT NULL,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	var current int
	if err := sqlDB.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&current); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}

	all, err := loadMigrations()
	if err != nil {
		return err
	}

	for _, m := range all {
		if m.version <= current {
			continue
		}
		if err := applyMigration(sqlDB, m); err != nil {
			return fmt.Errorf("migration %04d_%s: %w", m.version, m.name, err)
		}
	}
	return nil
}

func applyMigration(sqlDB *sql.DB, m migration) error {
	tx, err := sqlDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(m.sql); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO schema_migrations (version, name) VALUES (?, ?)`, m.version, m.name); err != nil {
		return err
	}
	return tx.Commit()
}

// loadMigrations reads migrations/NNNN_name.sql and returns them in version order.
func loadMigrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return nil, err
	}

	var out []migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		base := strings.TrimSuffix(e.Name(), ".sql")
		numPart, namePart, found := strings.Cut(base, "_")
		if !found {
			return nil, fmt.Errorf("migration %q must be named NNNN_description.sql", e.Name())
		}
		version, err := strconv.Atoi(numPart)
		if err != nil {
			return nil, fmt.Errorf("migration %q has a non-numeric version prefix: %w", e.Name(), err)
		}
		body, err := fs.ReadFile(migrationFS, path.Join("migrations", e.Name()))
		if err != nil {
			return nil, err
		}
		out = append(out, migration{version: version, name: namePart, sql: string(body)})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].version < out[j].version })

	for i := 1; i < len(out); i++ {
		if out[i].version == out[i-1].version {
			return nil, fmt.Errorf("duplicate migration version %d", out[i].version)
		}
	}
	return out, nil
}
