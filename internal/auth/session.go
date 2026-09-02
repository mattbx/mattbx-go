// Package auth gates /admin and /portfolio behind separate passwords.
//
// Sessions are stateless: the cookie carries an expiry and an HMAC over
// (scope, expiry). There is no sessions table, so logins survive restarts and
// redeploys, and a stolen cookie cannot be re-scoped from portfolio to admin.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Scope identifies which area a session unlocks. It is part of the signed
// payload, so an admin cookie and a portfolio cookie are not interchangeable.
type Scope string

const (
	ScopeAdmin     Scope = "admin"
	ScopePortfolio Scope = "portfolio"
)

// Admin sessions are short: that cookie can rewrite the site. Portfolio
// sessions are long so people you share the password with aren't re-prompted.
const (
	adminTTL     = 24 * time.Hour
	portfolioTTL = 30 * 24 * time.Hour
)

func (s Scope) cookieName() string { return "mbx_" + string(s) }

func (s Scope) ttl() time.Duration {
	if s == ScopeAdmin {
		return adminTTL
	}
	return portfolioTTL
}

// LoginPath is where an unauthenticated request to this scope is sent.
func (s Scope) LoginPath() string {
	if s == ScopeAdmin {
		return "/admin/login"
	}
	return "/portfolio/login"
}

type Manager struct {
	secret   []byte
	password map[Scope]string
	secure   bool // false on plain-HTTP localhost, true everywhere else
}

func NewManager(secret []byte, adminPassword, portfolioPassword string, development bool) *Manager {
	return &Manager{
		secret: secret,
		password: map[Scope]string{
			ScopeAdmin:     adminPassword,
			ScopePortfolio: portfolioPassword,
		},
		secure: !development,
	}
}

// CheckPassword compares in constant time so a timing signal can't leak the
// password one byte at a time.
func (m *Manager) CheckPassword(s Scope, attempt string) bool {
	want, ok := m.password[s]
	if !ok || want == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(attempt), []byte(want)) == 1
}

func (m *Manager) sign(s Scope, expiry string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(s))
	mac.Write([]byte{'.'})
	mac.Write([]byte(expiry))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (m *Manager) mint(s Scope, expiry time.Time) string {
	exp := strconv.FormatInt(expiry.Unix(), 10)
	return exp + "." + m.sign(s, exp)
}

// valid reports whether value is a well-formed, correctly signed, unexpired
// token for this scope.
func (m *Manager) valid(s Scope, value string) bool {
	exp, sig, ok := strings.Cut(value, ".")
	if !ok {
		return false
	}
	if !hmac.Equal([]byte(sig), []byte(m.sign(s, exp))) {
		return false
	}
	unix, err := strconv.ParseInt(exp, 10, 64)
	if err != nil {
		return false
	}
	return time.Now().Before(time.Unix(unix, 0))
}

// Authenticated reports whether r carries a valid session for this scope.
func (m *Manager) Authenticated(r *http.Request, s Scope) bool {
	c, err := r.Cookie(s.cookieName())
	if err != nil {
		return false
	}
	return m.valid(s, c.Value)
}

func (m *Manager) SetSession(w http.ResponseWriter, s Scope) {
	ttl := s.ttl()
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieName(),
		Value:    m.mint(s, time.Now().Add(ttl)),
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (m *Manager) ClearSession(w http.ResponseWriter, s Scope) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieName(),
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// SafeNext sanitises a post-login redirect target. Only same-site absolute
// paths are allowed; anything else returns "" so we never bounce a visitor to
// an attacker-chosen host.
func SafeNext(raw string) string {
	if raw == "" || !strings.HasPrefix(raw, "/") {
		return ""
	}
	// "//evil.com" and "/\evil.com" are protocol-relative in most browsers.
	if strings.HasPrefix(raw, "//") || strings.HasPrefix(raw, "/\\") {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.IsAbs() || u.Host != "" {
		return ""
	}
	return u.String()
}
