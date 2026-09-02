package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Project struct {
	ID        int64
	Slug      string
	Title     string
	Summary   string
	BodyMD    string
	BodyHTML  string
	Role      string
	Tech      string // comma-separated; see TechList
	LinkURL   string
	RepoURL   string
	SortOrder int
	Published bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TechList splits the comma-separated Tech field for rendering as tags.
func (p *Project) TechList() []string {
	var out []string
	for _, part := range strings.Split(p.Tech, ",") {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}

type ProjectStore struct{ db *sql.DB }

func NewProjectStore(sqlDB *sql.DB) *ProjectStore { return &ProjectStore{db: sqlDB} }

const projectColumns = `id, slug, title, summary, body_md, body_html, role, tech,
	link_url, repo_url, sort_order, published, created_at, updated_at`

func scanProject(row interface{ Scan(...any) error }) (*Project, error) {
	var p Project
	err := row.Scan(&p.ID, &p.Slug, &p.Title, &p.Summary, &p.BodyMD, &p.BodyHTML,
		&p.Role, &p.Tech, &p.LinkURL, &p.RepoURL, &p.SortOrder, &p.Published,
		&p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// List returns projects in curated order (sort_order ascending, newest first
// within a tie), optionally including unpublished drafts.
func (s *ProjectStore) List(ctx context.Context, includeDrafts bool) ([]*Project, error) {
	q := `SELECT ` + projectColumns + ` FROM projects`
	if !includeDrafts {
		q += ` WHERE published = 1`
	}
	q += ` ORDER BY sort_order ASC, created_at DESC, id DESC`

	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var out []*Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *ProjectStore) GetBySlug(ctx context.Context, slug string, includeDrafts bool) (*Project, error) {
	q := `SELECT ` + projectColumns + ` FROM projects WHERE slug = ?`
	if !includeDrafts {
		q += ` AND published = 1`
	}
	return scanProject(s.db.QueryRowContext(ctx, q, slug))
}

func (s *ProjectStore) GetByID(ctx context.Context, id int64) (*Project, error) {
	return scanProject(s.db.QueryRowContext(ctx, `SELECT `+projectColumns+` FROM projects WHERE id = ?`, id))
}

func (s *ProjectStore) Create(ctx context.Context, p *Project) error {
	now := time.Now().UTC()
	p.CreatedAt, p.UpdatedAt = now, now

	res, err := s.db.ExecContext(ctx,
		`INSERT INTO projects (slug, title, summary, body_md, body_html, role, tech,
		 link_url, repo_url, sort_order, published, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Slug, p.Title, p.Summary, p.BodyMD, p.BodyHTML, p.Role, p.Tech,
		p.LinkURL, p.RepoURL, p.SortOrder, p.Published, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create project: %w", err)
	}
	p.ID, err = res.LastInsertId()
	return err
}

func (s *ProjectStore) Update(ctx context.Context, p *Project) error {
	p.UpdatedAt = time.Now().UTC()

	res, err := s.db.ExecContext(ctx,
		`UPDATE projects SET slug = ?, title = ?, summary = ?, body_md = ?, body_html = ?,
		 role = ?, tech = ?, link_url = ?, repo_url = ?, sort_order = ?, published = ?,
		 updated_at = ? WHERE id = ?`,
		p.Slug, p.Title, p.Summary, p.BodyMD, p.BodyHTML, p.Role, p.Tech,
		p.LinkURL, p.RepoURL, p.SortOrder, p.Published, p.UpdatedAt, p.ID)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}
	return checkAffected(res)
}

func (s *ProjectStore) Delete(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	return checkAffected(res)
}

func (s *ProjectStore) SlugTaken(ctx context.Context, slug string, excludeID int64) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM projects WHERE slug = ? AND id <> ?`, slug, excludeID).Scan(&n)
	return n > 0, err
}
