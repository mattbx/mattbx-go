// Package db owns the SQLite connection, schema migrations, and data access.
package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // pure-Go driver; keeps CGO_ENABLED=0 builds working
)

// Open connects to the SQLite database at path, creating the parent directory
// if needed (a freshly provisioned Disco volume starts empty) and applying any
// outstanding migrations.
func Open(path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create database directory %s: %w", dir, err)
		}
	}

	// WAL keeps reads from blocking on the writer; busy_timeout absorbs the
	// brief overlap when Disco runs the old and new containers together.
	dsn := "file:" + url.PathEscape(path) + "?" + url.Values{
		"_pragma": {
			"journal_mode(WAL)",
			"busy_timeout(5000)",
			"foreign_keys(1)",
			"synchronous(NORMAL)",
		},
	}.Encode()

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// A single connection serialises writes in Go rather than letting SQLite
	// return SQLITE_BUSY. At personal-site traffic this costs nothing.
	sqlDB.SetMaxOpenConns(1)

	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := Migrate(sqlDB); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return sqlDB, nil
}
