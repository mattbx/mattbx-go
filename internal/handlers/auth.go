package handlers

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/mattbx/mattbx-go/internal/auth"
	"github.com/mattbx/mattbx-go/internal/ui"
)

// gateCopy is the wording for each sign-in page. The portfolio gate is an
// invitation to someone you shared a password with; the admin gate is a door.
func gateCopy(scope auth.Scope) ui.GateView {
	if scope == auth.ScopeAdmin {
		return ui.GateView{
			Mark:    "Admin",
			Title:   "Sign in",
			Lede:    "Write access to posts and portfolio projects.",
			Action:  "/admin/login",
			Confirm: "Sign in",
		}
	}
	return ui.GateView{
		Mark:    "Private",
		Title:   "This work is shared privately",
		Lede:    "Enter the password I sent you and the portfolio will open. It stays unlocked on this device for 30 days.",
		Action:  "/portfolio/login",
		Confirm: "Open the portfolio",
	}
}

func (s *Server) handleAdminLoginForm(w http.ResponseWriter, r *http.Request) {
	s.showGate(w, r, auth.ScopeAdmin, "")
}

func (s *Server) handlePortfolioLoginForm(w http.ResponseWriter, r *http.Request) {
	s.showGate(w, r, auth.ScopePortfolio, "")
}

func (s *Server) showGate(w http.ResponseWriter, r *http.Request, scope auth.Scope, errMsg string) {
	// Already signed in? Go straight through rather than asking again.
	if errMsg == "" && s.sessions.Authenticated(r, scope) {
		http.Redirect(w, r, s.destination(r, scope), http.StatusSeeOther)
		return
	}

	v := gateCopy(scope)
	v.Next = auth.SafeNext(r.URL.Query().Get("next"))
	v.Error = errMsg

	status := http.StatusOK
	if errMsg != "" {
		status = http.StatusUnauthorized
	}

	nav := ""
	if scope == auth.ScopePortfolio {
		nav = "portfolio"
	}
	p := s.page(r, v.Title, "", nav)
	s.render(w, r, status, ui.Gate(p, v))
}

func (s *Server) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	s.attemptLogin(w, r, auth.ScopeAdmin)
}

func (s *Server) handlePortfolioLogin(w http.ResponseWriter, r *http.Request) {
	s.attemptLogin(w, r, auth.ScopePortfolio)
}

// attemptLogin verifies a password and, on success, mints a session for that
// scope only.
func (s *Server) attemptLogin(w http.ResponseWriter, r *http.Request, scope auth.Scope) {
	if err := r.ParseForm(); err != nil {
		s.showGate(w, r, scope, "That form didn't come through. Try again.")
		return
	}

	// Rate-limit per client and per scope so probing the portfolio password
	// can't lock anyone out of /admin.
	key := string(scope) + "|" + auth.ClientIP(r, !s.cfg.Development())

	if wait := s.logins.RetryAfter(key); wait > 0 {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", int(math.Ceil(wait.Seconds()))))
		s.showGate(w, r, scope, "Too many attempts. Try again in "+humanDuration(wait)+".")
		return
	}
	if !s.logins.Allow(key) {
		s.showGate(w, r, scope, "Too many attempts. Try again shortly.")
		return
	}

	if !s.sessions.CheckPassword(scope, r.PostFormValue("password")) {
		s.log.Warn("failed login", "scope", scope, "ip", auth.ClientIP(r, !s.cfg.Development()))
		s.showGate(w, r, scope, "That password isn't right.")
		return
	}

	s.logins.Reset(key)
	s.sessions.SetSession(w, scope)
	http.Redirect(w, r, s.destination(r, scope), http.StatusSeeOther)
}

// destination resolves where to land after signing in: the sanitised ?next=
// target if there is one, otherwise the scope's home.
func (s *Server) destination(r *http.Request, scope auth.Scope) string {
	if next := auth.SafeNext(r.FormValue("next")); next != "" {
		return next
	}
	if next := auth.SafeNext(r.URL.Query().Get("next")); next != "" {
		return next
	}
	if scope == auth.ScopeAdmin {
		return "/admin"
	}
	return "/portfolio"
}

func (s *Server) handleAdminLogout(w http.ResponseWriter, r *http.Request) {
	s.sessions.ClearSession(w, auth.ScopeAdmin)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handlePortfolioLogout(w http.ResponseWriter, r *http.Request) {
	s.sessions.ClearSession(w, auth.ScopePortfolio)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// humanDuration renders a wait as "3 minutes" or "45 seconds".
func humanDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(math.Ceil(d.Seconds())))
	}
	mins := int(math.Ceil(d.Minutes()))
	if mins == 1 {
		return "a minute"
	}
	return fmt.Sprintf("%d minutes", mins)
}
