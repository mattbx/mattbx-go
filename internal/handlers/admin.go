package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/mattbx/mattbx-go/internal/db"
	"github.com/mattbx/mattbx-go/internal/markdown"
	"github.com/mattbx/mattbx-go/internal/slug"
	"github.com/mattbx/mattbx-go/internal/ui"
)

func (s *Server) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	posts, err := s.posts.List(r.Context(), true, 0)
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	projects, err := s.projects.List(r.Context(), true)
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	p := s.page(r, "Admin", "", "")
	s.render(w, r, http.StatusOK, ui.AdminDashboard(p, posts, projects))
}

// --- Posts ----------------------------------------------------------------

func (s *Server) handleNewPostForm(w http.ResponseWriter, r *http.Request) {
	s.renderPostForm(w, r, http.StatusOK, ui.PostFormView{
		Post:   &db.Post{Published: false},
		IsNew:  true,
		Action: "/admin/posts",
	})
}

func (s *Server) handleEditPostForm(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r)
	if !ok {
		return
	}
	post, err := s.posts.GetByID(r.Context(), id)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}
	s.renderPostForm(w, r, http.StatusOK, ui.PostFormView{
		Post:   post,
		Action: "/admin/posts/" + strconv.FormatInt(id, 10),
	})
}

func (s *Server) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	post := &db.Post{}
	if !s.bindPost(w, r, post, true, "/admin/posts") {
		return
	}
	if err := s.posts.Create(r.Context(), post); err != nil {
		s.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, post.PermalinkPath(), http.StatusSeeOther)
}

func (s *Server) handleUpdatePost(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r)
	if !ok {
		return
	}
	post, err := s.posts.GetByID(r.Context(), id)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}
	if !s.bindPost(w, r, post, false, "/admin/posts/"+strconv.FormatInt(id, 10)) {
		return
	}
	if err := s.posts.Update(r.Context(), post); err != nil {
		s.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, post.PermalinkPath(), http.StatusSeeOther)
}

// bindPost fills post from the submitted form, validates it, and re-renders
// the form with an explanation on failure. It reports whether binding
// succeeded. The half-filled post is passed straight back to the template, so
// a rejected submission never loses what was typed.
func (s *Server) bindPost(w http.ResponseWriter, r *http.Request, post *db.Post, isNew bool, action string) bool {
	fail := func(msg string) bool {
		s.renderPostForm(w, r, http.StatusUnprocessableEntity, ui.PostFormView{
			Post: post, IsNew: isNew, Action: action, Error: msg,
		})
		return false
	}

	if err := r.ParseForm(); err != nil {
		return fail("That form didn't come through. Try again.")
	}

	post.Title = strings.TrimSpace(r.PostFormValue("title"))
	post.Summary = strings.TrimSpace(r.PostFormValue("summary"))
	post.BodyMD = r.PostFormValue("body_md")
	post.Published = r.PostFormValue("published") == "1"
	post.Slug = resolveSlug(r.PostFormValue("slug"), post.Title)

	if post.Title == "" {
		return fail("Give the post a title.")
	}
	if post.Slug == "" {
		return fail("That title doesn't produce a usable URL. Set a slug by hand.")
	}

	taken, err := s.posts.SlugTaken(r.Context(), post.Slug, post.ID)
	if err != nil {
		s.serverError(w, r, err)
		return false
	}
	if taken {
		return fail("Another post already uses the slug “" + post.Slug + "”. Pick a different one.")
	}

	// Render and sanitize once, here, so nothing untrusted is ever stored.
	html, err := markdown.Render(post.BodyMD)
	if err != nil {
		s.serverError(w, r, err)
		return false
	}
	post.BodyHTML = html
	return true
}

func (s *Server) renderPostForm(w http.ResponseWriter, r *http.Request, status int, v ui.PostFormView) {
	p := s.page(r, v.Post.Title, "", "")
	s.render(w, r, status, ui.PostForm(p, v))
}

func (s *Server) handleConfirmDeletePost(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r)
	if !ok {
		return
	}
	post, err := s.posts.GetByID(r.Context(), id)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}
	idStr := strconv.FormatInt(id, 10)
	p := s.page(r, "Delete post", "", "")
	s.render(w, r, http.StatusOK, ui.ConfirmDelete(p, post.Title,
		"/admin/posts/"+idStr+"/delete", "/admin/posts/"+idStr+"/edit"))
}

func (s *Server) handleDeletePost(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r)
	if !ok {
		return
	}
	if err := s.posts.Delete(r.Context(), id); err != nil {
		s.handleStoreError(w, r, err)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// --- Projects -------------------------------------------------------------

func (s *Server) handleNewProjectForm(w http.ResponseWriter, r *http.Request) {
	s.renderProjectForm(w, r, http.StatusOK, ui.ProjectFormView{
		Project: &db.Project{},
		IsNew:   true,
		Action:  "/admin/projects",
	})
}

func (s *Server) handleEditProjectForm(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r)
	if !ok {
		return
	}
	project, err := s.projects.GetByID(r.Context(), id)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}
	s.renderProjectForm(w, r, http.StatusOK, ui.ProjectFormView{
		Project: project,
		Action:  "/admin/projects/" + strconv.FormatInt(id, 10),
	})
}

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	project := &db.Project{}
	if !s.bindProject(w, r, project, true, "/admin/projects") {
		return
	}
	if err := s.projects.Create(r.Context(), project); err != nil {
		s.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/portfolio/"+project.Slug, http.StatusSeeOther)
}

func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r)
	if !ok {
		return
	}
	project, err := s.projects.GetByID(r.Context(), id)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}
	if !s.bindProject(w, r, project, false, "/admin/projects/"+strconv.FormatInt(id, 10)) {
		return
	}
	if err := s.projects.Update(r.Context(), project); err != nil {
		s.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/portfolio/"+project.Slug, http.StatusSeeOther)
}

func (s *Server) bindProject(w http.ResponseWriter, r *http.Request, project *db.Project, isNew bool, action string) bool {
	fail := func(msg string) bool {
		s.renderProjectForm(w, r, http.StatusUnprocessableEntity, ui.ProjectFormView{
			Project: project, IsNew: isNew, Action: action, Error: msg,
		})
		return false
	}

	if err := r.ParseForm(); err != nil {
		return fail("That form didn't come through. Try again.")
	}

	project.Title = strings.TrimSpace(r.PostFormValue("title"))
	project.Summary = strings.TrimSpace(r.PostFormValue("summary"))
	project.BodyMD = r.PostFormValue("body_md")
	project.Role = strings.TrimSpace(r.PostFormValue("role"))
	project.Tech = strings.TrimSpace(r.PostFormValue("tech"))
	project.LinkURL = strings.TrimSpace(r.PostFormValue("link_url"))
	project.RepoURL = strings.TrimSpace(r.PostFormValue("repo_url"))
	project.Published = r.PostFormValue("published") == "1"
	project.Slug = resolveSlug(r.PostFormValue("slug"), project.Title)

	if raw := strings.TrimSpace(r.PostFormValue("sort_order")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return fail("Order must be a whole number.")
		}
		project.SortOrder = n
	}

	if project.Title == "" {
		return fail("Give the project a title.")
	}
	if project.Slug == "" {
		return fail("That title doesn't produce a usable URL. Set a slug by hand.")
	}
	for _, u := range []string{project.LinkURL, project.RepoURL} {
		if u != "" && !isHTTPURL(u) {
			return fail("Links must start with http:// or https://.")
		}
	}

	taken, err := s.projects.SlugTaken(r.Context(), project.Slug, project.ID)
	if err != nil {
		s.serverError(w, r, err)
		return false
	}
	if taken {
		return fail("Another project already uses the slug “" + project.Slug + "”. Pick a different one.")
	}

	html, err := markdown.Render(project.BodyMD)
	if err != nil {
		s.serverError(w, r, err)
		return false
	}
	project.BodyHTML = html
	return true
}

func (s *Server) renderProjectForm(w http.ResponseWriter, r *http.Request, status int, v ui.ProjectFormView) {
	p := s.page(r, v.Project.Title, "", "")
	s.render(w, r, status, ui.ProjectForm(p, v))
}

func (s *Server) handleConfirmDeleteProject(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r)
	if !ok {
		return
	}
	project, err := s.projects.GetByID(r.Context(), id)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}
	idStr := strconv.FormatInt(id, 10)
	p := s.page(r, "Delete project", "", "")
	s.render(w, r, http.StatusOK, ui.ConfirmDelete(p, project.Title,
		"/admin/projects/"+idStr+"/delete", "/admin/projects/"+idStr+"/edit"))
}

func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r)
	if !ok {
		return
	}
	if err := s.projects.Delete(r.Context(), id); err != nil {
		s.handleStoreError(w, r, err)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// --- Shared ---------------------------------------------------------------

// pathID parses the {id} wildcard, rendering a 404 for anything unparseable.
func (s *Server) pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		s.notFound(w, r)
		return 0, false
	}
	return id, true
}

// resolveSlug prefers what the author typed and falls back to the title, so
// editing a title never silently breaks an existing URL.
func resolveSlug(submitted, title string) string {
	if s := slug.Make(submitted); s != "" {
		return s
	}
	return slug.Make(title)
}

// isHTTPURL keeps javascript: and data: URLs out of project links, which are
// rendered as href attributes outside the Markdown sanitizer's reach.
func isHTTPURL(raw string) bool {
	return strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://")
}
