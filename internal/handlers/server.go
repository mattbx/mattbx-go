// Package handlers wires HTTP routes to the stores and templates.
package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/mattbx/mattbx-go/internal/auth"
	"github.com/mattbx/mattbx-go/internal/config"
	"github.com/mattbx/mattbx-go/internal/db"
	"github.com/mattbx/mattbx-go/internal/ui"
)

type Server struct {
	cfg      *config.Config
	db       *sql.DB
	posts    *db.PostStore
	projects *db.ProjectStore
	sessions *auth.Manager
	logins   *auth.Limiter
	log      *slog.Logger
}

func New(cfg *config.Config, sqlDB *sql.DB, log *slog.Logger) *Server {
	return &Server{
		cfg:      cfg,
		db:       sqlDB,
		posts:    db.NewPostStore(sqlDB),
		projects: db.NewProjectStore(sqlDB),
		sessions: auth.NewManager(cfg.SessionSecret, cfg.AdminPassword, cfg.PortfolioPassword, cfg.Development()),
		logins:   auth.NewLimiter(5, 15*time.Minute),
		log:      log,
	}
}

// page builds the shell context for a request. IsAdmin is presentation only —
// it reveals edit affordances and drafts, never grants access on its own.
func (s *Server) page(r *http.Request, title, description, nav string) ui.Page {
	return ui.Page{
		Title:       title,
		Description: description,
		Nav:         nav,
		IsAdmin:     s.isAdmin(r),
		BaseURL:     s.cfg.BaseURL,
		Path:        r.URL.Path,
	}
}

func (s *Server) isAdmin(r *http.Request) bool {
	return s.sessions.Authenticated(r, auth.ScopeAdmin)
}

// render buffers the component before writing so a mid-render failure produces
// a clean 500 instead of a half-written page under a 200.
func (s *Server) render(w http.ResponseWriter, r *http.Request, status int, c templ.Component) {
	var buf bytes.Buffer
	if err := c.Render(r.Context(), &buf); err != nil {
		s.log.Error("render failed", "path", r.URL.Path, "err", err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	p := s.page(r, "Not found", "", "")
	s.render(w, r, http.StatusNotFound, ui.ErrorPage(p, http.StatusNotFound,
		"That page isn't here",
		"The link may be out of date, or the page may never have been published."))
}

// serverError logs the cause and shows the visitor nothing about it.
func (s *Server) serverError(w http.ResponseWriter, r *http.Request, err error) {
	s.log.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
	p := s.page(r, "Something went wrong", "", "")
	s.render(w, r, http.StatusInternalServerError, ui.ErrorPage(p, http.StatusInternalServerError,
		"Something went wrong",
		"That's on me, not you. Try again in a moment."))
}

// handleStoreError maps a store error to the right response: a missing row is
// a 404, anything else is a 500.
func (s *Server) handleStoreError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, db.ErrNotFound) {
		s.notFound(w, r)
		return
	}
	s.serverError(w, r, err)
}

// timeoutContext derives a bounded context from the request.
func timeoutContext(r *http.Request, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), d)
}
