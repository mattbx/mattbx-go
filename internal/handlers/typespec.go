package handlers

import (
	"database/sql"
	"net/http"
	"sync"
	"time"

	"github.com/mattbx/mattbx-go/internal/db"
	"github.com/mattbx/mattbx-go/internal/markdown"
	"github.com/mattbx/mattbx-go/internal/ui"
)

// typeSpecimenPost is the sample post on /admin/type. Its body is a constant,
// rendered once like a saved post, so nothing is parsed per request.
var typeSpecimenPost = sync.OnceValues(func() (*db.Post, error) {
	html, err := markdown.Render(ui.TypeSpecimenMarkdown)
	if err != nil {
		return nil, err
	}
	return &db.Post{
		Slug:        "type-specimen",
		Title:       "A post title in context",
		Summary:     "The summary line under the title, the way every post has one.",
		BodyMD:      ui.TypeSpecimenMarkdown,
		BodyHTML:    html,
		Published:   true,
		PublishedAt: sql.NullTime{Time: time.Date(2026, 11, 8, 0, 0, 0, 0, time.UTC), Valid: true},
		Tags:        "type, specimen",
	}, nil
})

// handleTypeSpecimen shows every type role with live computed values.
func (s *Server) handleTypeSpecimen(w http.ResponseWriter, r *http.Request) {
	post, err := typeSpecimenPost()
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	p := s.page(r, "Type specimen", "", "")
	p.Scripts = []string{"typespec.js"}
	s.render(w, r, http.StatusOK, ui.TypeSpecimen(p, post))
}
