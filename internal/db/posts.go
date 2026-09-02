package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrNotFound is returned when a lookup matches no row. Handlers translate it
// into a 404 rather than a 500.
var ErrNotFound = errors.New("not found")

type Post struct {
	ID          int64
	Slug        string
	Title       string
	Summary     string
	BodyMD      string
	BodyHTML    string
	Published   bool
	PublishedAt sql.NullTime
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Date is the human-facing publication date, falling back to the creation date
// for drafts that have never been published.
func (p *Post) Date() time.Time {
	if p.PublishedAt.Valid {
		return p.PublishedAt.Time
	}
	return p.CreatedAt
}

type PostStore struct{ db *sql.DB }

func NewPostStore(sqlDB *sql.DB) *PostStore { return &PostStore{db: sqlDB} }

const postColumns = `id, slug, title, summary, body_md, body_html, published, published_at, created_at, updated_at`

func scanPost(row interface{ Scan(...any) error }) (*Post, error) {
	var p Post
	err := row.Scan(&p.ID, &p.Slug, &p.Title, &p.Summary, &p.BodyMD, &p.BodyHTML,
		&p.Published, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// List returns posts newest first. When includeDrafts is false only published
// posts are returned — that flag is the single gate protecting drafts.
func (s *PostStore) List(ctx context.Context, includeDrafts bool, limit int) ([]*Post, error) {
	q := `SELECT ` + postColumns + ` FROM posts`
	if !includeDrafts {
		q += ` WHERE published = 1`
	}
	q += ` ORDER BY COALESCE(published_at, created_at) DESC, id DESC`
	args := []any{}
	if limit > 0 {
		q += ` LIMIT ?`
		args = append(args, limit)
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	var out []*Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *PostStore) GetBySlug(ctx context.Context, slug string, includeDrafts bool) (*Post, error) {
	q := `SELECT ` + postColumns + ` FROM posts WHERE slug = ?`
	if !includeDrafts {
		q += ` AND published = 1`
	}
	return scanPost(s.db.QueryRowContext(ctx, q, slug))
}

func (s *PostStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	return scanPost(s.db.QueryRowContext(ctx, `SELECT `+postColumns+` FROM posts WHERE id = ?`, id))
}

func (s *PostStore) Create(ctx context.Context, p *Post) error {
	now := time.Now().UTC()
	p.CreatedAt, p.UpdatedAt = now, now
	p.PublishedAt = publishStamp(p.Published, p.PublishedAt, now)

	res, err := s.db.ExecContext(ctx,
		`INSERT INTO posts (slug, title, summary, body_md, body_html, published, published_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Slug, p.Title, p.Summary, p.BodyMD, p.BodyHTML, p.Published, p.PublishedAt, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create post: %w", err)
	}
	p.ID, err = res.LastInsertId()
	return err
}

func (s *PostStore) Update(ctx context.Context, p *Post) error {
	now := time.Now().UTC()
	p.UpdatedAt = now
	p.PublishedAt = publishStamp(p.Published, p.PublishedAt, now)

	res, err := s.db.ExecContext(ctx,
		`UPDATE posts SET slug = ?, title = ?, summary = ?, body_md = ?, body_html = ?,
		 published = ?, published_at = ?, updated_at = ? WHERE id = ?`,
		p.Slug, p.Title, p.Summary, p.BodyMD, p.BodyHTML, p.Published, p.PublishedAt, p.UpdatedAt, p.ID)
	if err != nil {
		return fmt.Errorf("update post: %w", err)
	}
	return checkAffected(res)
}

func (s *PostStore) Delete(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM posts WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	return checkAffected(res)
}

// SlugTaken reports whether slug belongs to a different row, so the admin form
// can show a friendly message instead of surfacing a UNIQUE constraint error.
func (s *PostStore) SlugTaken(ctx context.Context, slug string, excludeID int64) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM posts WHERE slug = ? AND id <> ?`, slug, excludeID).Scan(&n)
	return n > 0, err
}

// publishStamp sets published_at the first time something is published and
// clears it when it returns to draft, so the date always reflects reality.
func publishStamp(published bool, existing sql.NullTime, now time.Time) sql.NullTime {
	switch {
	case !published:
		return sql.NullTime{}
	case existing.Valid:
		return existing
	default:
		return sql.NullTime{Time: now, Valid: true}
	}
}

func checkAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
