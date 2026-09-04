package db

import (
	"context"
	"path/filepath"
	"testing"
)

// testDB opens a throwaway database with all migrations applied.
func testDB(t *testing.T) *PostStore {
	t.Helper()
	sqlDB, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	return NewPostStore(sqlDB)
}

func TestMigrateAppliesAndIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	sqlDB, err := Open(path)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}

	var version, count int
	if err := sqlDB.QueryRow(`SELECT MAX(version), COUNT(*) FROM schema_migrations`).Scan(&version, &count); err != nil {
		t.Fatalf("read version: %v", err)
	}
	// Not a fixed number: this asserts every embedded migration file was
	// applied exactly once, not "there are exactly N migrations" — which
	// would need editing every time a new migration is added.
	files, err := loadMigrations()
	if err != nil {
		t.Fatal(err)
	}
	if count != len(files) {
		t.Fatalf("applied %d migrations, want %d (one per embedded file)", count, len(files))
	}
	if version != files[len(files)-1].version {
		t.Fatalf("schema version = %d, want %d (the newest migration)", version, files[len(files)-1].version)
	}

	for _, table := range []string{"posts", "projects"} {
		var name string
		if err := sqlDB.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name); err != nil {
			t.Fatalf("table %s missing after migrate: %v", table, err)
		}
	}
	sqlDB.Close()

	// Reopening must not re-apply migration 1 (which would fail on CREATE TABLE).
	again, err := Open(path)
	if err != nil {
		t.Fatalf("second open re-applied migrations: %v", err)
	}
	defer again.Close()

	var applied int
	if err := again.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&applied); err != nil {
		t.Fatal(err)
	}
	if applied != count {
		t.Fatalf("schema_migrations has %d rows after two opens, want %d (unchanged from the first open)", applied, count)
	}
}

func TestPostLifecycle(t *testing.T) {
	ctx := context.Background()
	store := testDB(t)

	p := &Post{Slug: "hello", Title: "Hello", BodyMD: "# hi", BodyHTML: "<h1>hi</h1>"}
	if err := store.Create(ctx, p); err != nil {
		t.Fatalf("create: %v", err)
	}
	if p.ID == 0 {
		t.Fatal("create did not populate ID")
	}
	if p.PublishedAt.Valid {
		t.Fatal("an unpublished post must not have published_at set")
	}

	// A draft is invisible to the public read path but visible to admin.
	if _, err := store.GetBySlug(ctx, "hello", false); err != ErrNotFound {
		t.Fatalf("draft leaked to public lookup: err = %v", err)
	}
	if _, err := store.GetBySlug(ctx, "hello", true); err != nil {
		t.Fatalf("admin lookup of draft failed: %v", err)
	}

	// Publishing stamps the date once and keeps it stable across edits.
	p.Published = true
	if err := store.Update(ctx, p); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if !p.PublishedAt.Valid {
		t.Fatal("publishing did not set published_at")
	}
	first := p.PublishedAt.Time

	p.Title = "Hello again"
	if err := store.Update(ctx, p); err != nil {
		t.Fatalf("edit: %v", err)
	}
	if !p.PublishedAt.Time.Equal(first) {
		t.Fatalf("editing moved published_at: %v -> %v", first, p.PublishedAt.Time)
	}

	published, err := store.List(ctx, false, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(published) != 1 {
		t.Fatalf("published list has %d posts, want 1", len(published))
	}

	// Unpublishing clears the stamp so a re-publish reflects the real date.
	p.Published = false
	if err := store.Update(ctx, p); err != nil {
		t.Fatal(err)
	}
	if p.PublishedAt.Valid {
		t.Fatal("unpublishing left published_at set")
	}

	taken, err := store.SlugTaken(ctx, "hello", 0)
	if err != nil {
		t.Fatal(err)
	}
	if !taken {
		t.Fatal("SlugTaken should report a conflict for a different row")
	}
	if taken, _ = store.SlugTaken(ctx, "hello", p.ID); taken {
		t.Fatal("SlugTaken must ignore the row being edited")
	}

	if err := store.Delete(ctx, p.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := store.Delete(ctx, p.ID); err != ErrNotFound {
		t.Fatalf("deleting a missing row = %v, want ErrNotFound", err)
	}
}
