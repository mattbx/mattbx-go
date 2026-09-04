package handlers

import (
	"net/http"

	"github.com/mattbx/mattbx-go/internal/auth"
	"github.com/mattbx/mattbx-go/internal/ui/static"
)

// Routes builds the full route table.
//
// Access control lives here and only here: a route is either registered
// bare (public) or wrapped in requireAdmin/requirePortfolio. Handlers never
// re-check permissions, so there is one place to audit.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	requireAdmin := s.sessions.Require(auth.ScopeAdmin)
	requirePortfolio := s.sessions.Require(auth.ScopePortfolio)

	// --- Public -----------------------------------------------------------
	mux.HandleFunc("GET /{$}", s.handleHome)
	mux.HandleFunc("GET /blog", s.handleBlogIndex)
	mux.HandleFunc("GET /blog/{slug}", s.handleBlogPost)
	mux.HandleFunc("GET /feed.xml", s.handleFeed)
	mux.HandleFunc("GET /robots.txt", s.handleRobots)
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.Handle("GET /static/", static.Handler())

	// Micropub uses its own bearer-token middleware (see micropub.go) rather
	// than a cookie scope, wired the same way requireAdmin/requirePortfolio
	// are below — every route's access control still reads from this file.
	requireMicropubToken := s.requireMicropubToken
	mux.Handle("GET /micropub", requireMicropubToken(http.HandlerFunc(s.handleMicropubQuery)))
	mux.Handle("POST /micropub", requireMicropubToken(http.HandlerFunc(s.handleMicropubAction)))

	// --- Sign in / out ----------------------------------------------------
	mux.HandleFunc("GET /portfolio/login", s.handlePortfolioLoginForm)
	mux.HandleFunc("POST /portfolio/login", s.handlePortfolioLogin)
	mux.HandleFunc("GET /admin/login", s.handleAdminLoginForm)
	mux.HandleFunc("POST /admin/login", s.handleAdminLogin)
	mux.HandleFunc("POST /admin/logout", s.handleAdminLogout)
	mux.HandleFunc("POST /portfolio/logout", s.handlePortfolioLogout)

	// --- Portfolio (gated) ------------------------------------------------
	// Registered before the {slug} route so /portfolio/login stays reachable;
	// ServeMux prefers the more specific pattern regardless of order, but the
	// grouping keeps the precedence obvious to a reader.
	mux.Handle("GET /portfolio", requirePortfolio(http.HandlerFunc(s.handlePortfolioIndex)))
	mux.Handle("GET /portfolio/{slug}", requirePortfolio(http.HandlerFunc(s.handlePortfolioProject)))

	// --- Admin (gated) ----------------------------------------------------
	mux.Handle("GET /admin", requireAdmin(http.HandlerFunc(s.handleAdminDashboard)))

	mux.Handle("GET /admin/posts/new", requireAdmin(http.HandlerFunc(s.handleNewPostForm)))
	mux.Handle("POST /admin/posts", requireAdmin(http.HandlerFunc(s.handleCreatePost)))
	mux.Handle("GET /admin/posts/{id}/edit", requireAdmin(http.HandlerFunc(s.handleEditPostForm)))
	mux.Handle("POST /admin/posts/{id}", requireAdmin(http.HandlerFunc(s.handleUpdatePost)))
	mux.Handle("GET /admin/posts/{id}/delete", requireAdmin(http.HandlerFunc(s.handleConfirmDeletePost)))
	mux.Handle("POST /admin/posts/{id}/delete", requireAdmin(http.HandlerFunc(s.handleDeletePost)))

	mux.Handle("GET /admin/projects/new", requireAdmin(http.HandlerFunc(s.handleNewProjectForm)))
	mux.Handle("POST /admin/projects", requireAdmin(http.HandlerFunc(s.handleCreateProject)))
	mux.Handle("GET /admin/projects/{id}/edit", requireAdmin(http.HandlerFunc(s.handleEditProjectForm)))
	mux.Handle("POST /admin/projects/{id}", requireAdmin(http.HandlerFunc(s.handleUpdateProject)))
	mux.Handle("GET /admin/projects/{id}/delete", requireAdmin(http.HandlerFunc(s.handleConfirmDeleteProject)))
	mux.Handle("POST /admin/projects/{id}/delete", requireAdmin(http.HandlerFunc(s.handleDeleteProject)))

	// Anything unmatched gets the styled 404 rather than ServeMux's plain one.
	mux.HandleFunc("/", s.notFound)

	return s.securityHeaders(s.requestLog(mux))
}

// securityHeaders applies defence-in-depth headers to every response.
func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("X-Frame-Options", "DENY")

		// The site ships no JavaScript and loads nothing cross-origin, so the
		// policy can be strict. It is skipped in development because Air's
		// proxy injects an inline live-reload script that 'none' would block.
		if !s.cfg.Development() {
			h.Set("Content-Security-Policy",
				"default-src 'self'; script-src 'none'; style-src 'self'; "+
					"img-src 'self' data:; base-uri 'none'; form-action 'self'; "+
					"frame-ancestors 'none'")
		}

		// Gated areas must never end up in a search index or a shared cache.
		if isPrivatePath(r.URL.Path) {
			h.Set("X-Robots-Tag", "noindex, nofollow")
			h.Set("Cache-Control", "no-store")
		}

		next.ServeHTTP(w, r)
	})
}

func isPrivatePath(path string) bool {
	return path == "/admin" || path == "/portfolio" || path == "/micropub" ||
		hasPrefix(path, "/admin/") || hasPrefix(path, "/portfolio/")
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
