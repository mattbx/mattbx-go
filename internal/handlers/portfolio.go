package handlers

import (
	"net/http"

	"github.com/mattbx/mattbx-go/internal/ui"
)

func (s *Server) handleWorkProject(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	project, err := s.projects.GetBySlug(r.Context(), slug, s.isAdmin(r))
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}
	projects, err := s.projects.List(r.Context(), s.isAdmin(r))
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	p := s.page(r, project.Title, project.Summary, "work")
	s.render(w, r, http.StatusOK, ui.WorkDetail(p, project, projects))
}

func (s *Server) redirectPortfolioIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func (s *Server) redirectPortfolioProject(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/work/"+r.PathValue("slug"), http.StatusMovedPermanently)
}
