package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testManager() *Manager {
	return NewManager([]byte("0123456789abcdef0123456789abcdef"), "admin-pw", "portfolio-pw", true)
}

// requestWith builds a request carrying the given cookie value for a scope.
func requestWith(s Scope, value string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: s.cookieName(), Value: value})
	return r
}

func TestSessionRoundTrip(t *testing.T) {
	m := testManager()
	w := httptest.NewRecorder()
	m.SetSession(w, ScopeAdmin)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("got %d cookies, want 1", len(cookies))
	}
	c := cookies[0]
	if !c.HttpOnly {
		t.Error("session cookie must be HttpOnly")
	}
	if c.SameSite != http.SameSiteLaxMode {
		t.Error("session cookie must be SameSite=Lax")
	}
	if !m.Authenticated(requestWith(ScopeAdmin, c.Value), ScopeAdmin) {
		t.Fatal("freshly minted cookie did not authenticate")
	}
}

func TestSecureFlagFollowsEnvironment(t *testing.T) {
	for _, tc := range []struct{ dev, wantSecure bool }{{true, false}, {false, true}} {
		m := NewManager([]byte("0123456789abcdef0123456789abcdef"), "a", "p", tc.dev)
		w := httptest.NewRecorder()
		m.SetSession(w, ScopeAdmin)
		if got := w.Result().Cookies()[0].Secure; got != tc.wantSecure {
			t.Errorf("development=%v: Secure=%v, want %v", tc.dev, got, tc.wantSecure)
		}
	}
}

func TestTamperedCookieRejected(t *testing.T) {
	m := testManager()
	w := httptest.NewRecorder()
	m.SetSession(w, ScopeAdmin)
	valid := w.Result().Cookies()[0].Value
	exp, sig, _ := strings.Cut(valid, ".")

	// Flipping one character of the signature must invalidate it.
	bad := []byte(sig)
	if bad[0] == 'A' {
		bad[0] = 'B'
	} else {
		bad[0] = 'A'
	}

	for name, value := range map[string]string{
		"tampered signature": exp + "." + string(bad),
		"extended expiry":    "99999999999." + sig,
		"no signature":       exp,
		"empty":              "",
		"garbage":            "not-a-token",
	} {
		if m.Authenticated(requestWith(ScopeAdmin, value), ScopeAdmin) {
			t.Errorf("%s was accepted", name)
		}
	}
}

func TestExpiredCookieRejected(t *testing.T) {
	m := testManager()
	expired := m.mint(ScopeAdmin, time.Now().Add(-time.Minute))
	if m.Authenticated(requestWith(ScopeAdmin, expired), ScopeAdmin) {
		t.Fatal("expired cookie was accepted")
	}
}

// The whole point of two scopes: a portfolio visitor must never reach /admin.
func TestScopesAreNotInterchangeable(t *testing.T) {
	m := testManager()
	w := httptest.NewRecorder()
	m.SetSession(w, ScopePortfolio)
	portfolioToken := w.Result().Cookies()[0].Value

	// Even replayed under the admin cookie name, the token must not validate,
	// because the scope is inside the signed payload.
	if m.Authenticated(requestWith(ScopeAdmin, portfolioToken), ScopeAdmin) {
		t.Fatal("a portfolio token authenticated as admin")
	}
}

func TestCheckPassword(t *testing.T) {
	m := testManager()
	if !m.CheckPassword(ScopeAdmin, "admin-pw") {
		t.Error("correct admin password rejected")
	}
	if m.CheckPassword(ScopeAdmin, "portfolio-pw") {
		t.Error("portfolio password unlocked admin")
	}
	if m.CheckPassword(ScopeAdmin, "") {
		t.Error("empty password accepted")
	}
}

func TestRequireRedirectsAnonymousUsers(t *testing.T) {
	m := testManager()
	guarded := m.Require(ScopePortfolio)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("secret"))
	}))

	w := httptest.NewRecorder()
	guarded.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/portfolio/thing", nil))

	res := w.Result()
	if res.StatusCode != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", res.StatusCode)
	}
	loc := res.Header.Get("Location")
	if !strings.HasPrefix(loc, "/portfolio/login") {
		t.Fatalf("Location = %q, want the portfolio login page", loc)
	}
	if !strings.Contains(loc, "next=") {
		t.Errorf("Location = %q, want a next= parameter", loc)
	}
	if strings.Contains(w.Body.String(), "secret") {
		t.Fatal("guarded handler body leaked to an anonymous request")
	}
}

func TestRequireAllowsValidSession(t *testing.T) {
	m := testManager()
	guarded := m.Require(ScopePortfolio)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("secret"))
	}))

	rec := httptest.NewRecorder()
	m.SetSession(rec, ScopePortfolio)

	r := httptest.NewRequest(http.MethodGet, "/portfolio", nil)
	r.AddCookie(rec.Result().Cookies()[0])

	w := httptest.NewRecorder()
	guarded.ServeHTTP(w, r)

	if w.Code != http.StatusOK || w.Body.String() != "secret" {
		t.Fatalf("valid session got %d %q", w.Code, w.Body.String())
	}
}

func TestSafeNextRejectsOffsiteTargets(t *testing.T) {
	for _, ok := range []string{"/blog", "/portfolio/thing?a=1", "/"} {
		if got := SafeNext(ok); got != ok {
			t.Errorf("SafeNext(%q) = %q, want it preserved", ok, got)
		}
	}
	for _, bad := range []string{
		"https://evil.com", "//evil.com", "/\\evil.com",
		"http://evil.com/x", "evil.com", "", "javascript:alert(1)",
	} {
		if got := SafeNext(bad); got != "" {
			t.Errorf("SafeNext(%q) = %q, want \"\" (open redirect)", bad, got)
		}
	}
}

func TestLimiterBlocksAfterMax(t *testing.T) {
	now := time.Now()
	l := NewLimiter(3, 15*time.Minute)
	l.nowFunc = func() time.Time { return now }

	for i := range 3 {
		if !l.Allow("1.2.3.4") {
			t.Fatalf("attempt %d blocked early", i+1)
		}
	}
	if l.Allow("1.2.3.4") {
		t.Fatal("4th attempt allowed past a limit of 3")
	}
	if l.RetryAfter("1.2.3.4") <= 0 {
		t.Error("RetryAfter should be positive while blocked")
	}
	if !l.Allow("5.6.7.8") {
		t.Error("a different client must not be affected")
	}

	// A successful login clears the count.
	l.Reset("1.2.3.4")
	if !l.Allow("1.2.3.4") {
		t.Error("Reset did not clear the attempt count")
	}

	// The window eventually expires.
	l2 := NewLimiter(1, time.Minute)
	l2.nowFunc = func() time.Time { return now }
	l2.Allow("9.9.9.9")
	if l2.Allow("9.9.9.9") {
		t.Fatal("limit not enforced")
	}
	l2.nowFunc = func() time.Time { return now.Add(2 * time.Minute) }
	if !l2.Allow("9.9.9.9") {
		t.Error("window did not expire")
	}
}

func TestClientIPBehindProxy(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:5555"
	r.Header.Set("X-Forwarded-For", "203.0.113.9")

	if got := ClientIP(r, true); got != "203.0.113.9" {
		t.Errorf("trusted proxy: got %q, want the forwarded client", got)
	}
	if got := ClientIP(r, false); got != "10.0.0.1" {
		t.Errorf("untrusted: got %q, want the peer address", got)
	}

	// A client forging the header only prepends; the proxy's own append wins.
	r.Header.Set("X-Forwarded-For", "1.1.1.1, 203.0.113.9")
	if got := ClientIP(r, true); got != "203.0.113.9" {
		t.Errorf("spoofed prefix: got %q, want the rightmost entry", got)
	}
}
