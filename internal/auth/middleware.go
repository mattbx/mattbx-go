package auth

import (
	"net/http"
	"net/url"
)

// Require wraps h so that only requests carrying a valid session for scope get
// through. Everyone else is redirected to that scope's login page with a
// ?next= pointing back at what they asked for.
func (m *Manager) Require(scope Scope) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if m.Authenticated(r, scope) {
				h.ServeHTTP(w, r)
				return
			}
			target := scope.LoginPath()
			if next := SafeNext(r.URL.RequestURI()); next != "" {
				target += "?next=" + url.QueryEscape(next)
			}
			http.Redirect(w, r, target, http.StatusSeeOther)
		})
	}
}

// RequireFunc is Require for a bare handler func.
func (m *Manager) RequireFunc(scope Scope, h http.HandlerFunc) http.Handler {
	return m.Require(scope)(h)
}
