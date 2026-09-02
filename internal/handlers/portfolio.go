package handlers

import (
	"net/http"

	"github.com/mattbx/mattbx-go/internal/ui"
)

// Both handlers are registered behind requirePortfolio in Routes, so reaching
// them at all means the visitor holds a valid portfolio session.

func (s *Server) handlePortfolioIndex(w http.ResponseWriter, r *http.Request) {
	projects, err := s.projects.List(r.Context(), s.isAdmin(r))
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	p := s.page(r, "Selected work", "", "portfolio")
	s.render(w, r, http.StatusOK, ui.PortfolioIndex(p, projects))
}

func (s *Server) handlePortfolioProject(w http.ResponseWriter, r *http.Request) {
	project, err := s.projects.GetBySlug(r.Context(), r.PathValue("slug"), s.isAdmin(r))
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}
	p := s.page(r, project.Title, project.Summary, "portfolio")
	s.render(w, r, http.StatusOK, ui.PortfolioProject(p, project))
}
