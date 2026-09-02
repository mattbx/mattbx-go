package db

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestDriverTimeRoundTrip(t *testing.T) {
	d, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "p.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	if _, err := d.Exec(`CREATE TABLE t (a DATETIME NOT NULL, b DATETIME)`); err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 9, 2, 15, 4, 5, 0, time.UTC)
	if _, err := d.Exec(`INSERT INTO t (a, b) VALUES (?, ?)`, want, nil); err != nil {
		t.Fatal(err)
	}
	var got time.Time
	var nul sql.NullTime
	if err := d.QueryRow(`SELECT a, b FROM t`).Scan(&got, &nul); err != nil {
		t.Fatalf("scan into time.Time failed: %v", err)
	}
	if !got.UTC().Equal(want) {
		t.Fatalf("round trip mismatch: got %v want %v", got.UTC(), want)
	}
	if nul.Valid {
		t.Fatalf("NULL should scan as invalid")
	}
	var raw string
	d.QueryRow(`SELECT a FROM t`).Scan(&raw)
	t.Logf("OK: time.Time round-trips; stored on disk as %q", raw)
}
