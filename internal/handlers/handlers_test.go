package handlers

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattbx/mattbx-go/internal/config"
	"github.com/mattbx/mattbx-go/internal/db"
)

const (
	adminPassword     = "admin-secret"
	portfolioPassword = "portfolio-secret"
	micropubToken     = "micropub-test-token-0123456789ab" // >= 32 chars, config requires it
)

// newTestServer builds a server backed by a throwaway database.
func newTestServer(t *testing.T) (*Server, http.Handler, *db.PostStore, *db.ProjectStore) {
	t.Helper()

	sqlDB, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	cfg := &config.Config{
		Env:               "development",
		Port:              "0",
		BaseURL:           "https://example.test",
		DBPath:            ":memory:",
		AdminPassword:     adminPassword,
		PortfolioPassword: portfolioPassword,
		SessionSecret:     []byte("0123456789abcdef0123456789abcdef"),
		MicropubToken:     micropubToken,
	}

	s := New(cfg, sqlDB, slog.New(slog.NewTextHandler(io.Discard, nil)))
	return s, s.Routes(), db.NewPostStore(sqlDB), db.NewProjectStore(sqlDB)
}

// signIn performs a real login and returns the resulting session cookie.
func signIn(t *testing.T, h http.Handler, path, password string) *http.Cookie {
	t.Helper()

	form := url.Values{"password": {password}}
	r := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("login at %s: status %d, want 303\n%s", path, w.Code, w.Body.String())
	}
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("login at %s set no cookie", path)
	}
	return cookies[0]
}

func get(h http.Handler, path string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, path, nil)
	for _, c := range cookies {
		r.AddCookie(c)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func postForm(h http.Handler, path string, form url.Values, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		r.AddCookie(c)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

// --- Access control -------------------------------------------------------

func TestGatedRoutesRedirectAnonymousUsers(t *testing.T) {
	_, h, _, _ := newTestServer(t)

	for _, path := range []string{
		"/portfolio", "/portfolio/anything",
		"/admin", "/admin/posts/new", "/admin/projects/new",
		"/admin/posts/1/edit", "/admin/posts/1/delete",
	} {
		w := get(h, path)
		if w.Code != http.StatusSeeOther {
			t.Errorf("GET %s: status %d, want 303 redirect to a login page", path, w.Code)
			continue
		}
		if loc := w.Header().Get("Location"); !strings.Contains(loc, "login") {
			t.Errorf("GET %s redirected to %q, want a login page", path, loc)
		}
	}
}

// The two passwords must stay independent — this is the whole point of having
// separate scopes.
func TestPortfolioSessionCannotReachAdmin(t *testing.T) {
	_, h, _, _ := newTestServer(t)
	portfolio := signIn(t, h, "/portfolio/login", portfolioPassword)

	if w := get(h, "/portfolio", portfolio); w.Code != http.StatusOK {
		t.Fatalf("portfolio session cannot open /portfolio: %d", w.Code)
	}
	if w := get(h, "/admin", portfolio); w.Code != http.StatusSeeOther {
		t.Errorf("portfolio session reached /admin: status %d, want a redirect", w.Code)
	}
}

func TestAdminSessionCannotReachPortfolio(t *testing.T) {
	_, h, _, _ := newTestServer(t)
	admin := signIn(t, h, "/admin/login", adminPassword)

	if w := get(h, "/admin", admin); w.Code != http.StatusOK {
		t.Fatalf("admin session cannot open /admin: %d", w.Code)
	}
	// Admin is not implicitly a portfolio viewer; you sign in to each door.
	if w := get(h, "/portfolio", admin); w.Code != http.StatusSeeOther {
		t.Errorf("admin session reached /portfolio without its password: %d", w.Code)
	}
}

func TestWrongPasswordIsRejected(t *testing.T) {
	_, h, _, _ := newTestServer(t)

	w := postForm(h, "/admin/login", url.Values{"password": {"wrong"}})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
	if len(w.Result().Cookies()) != 0 {
		t.Error("a failed login set a session cookie")
	}
	// The portfolio password must not open the admin door either.
	w = postForm(h, "/admin/login", url.Values{"password": {portfolioPassword}})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("portfolio password on /admin/login: status %d, want 401", w.Code)
	}
}

func TestLoginIsRateLimited(t *testing.T) {
	_, h, _, _ := newTestServer(t)

	// The limiter allows 5 attempts per window.
	for i := range 5 {
		if w := postForm(h, "/admin/login", url.Values{"password": {"wrong"}}); w.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: status %d, want 401", i+1, w.Code)
		}
	}
	w := postForm(h, "/admin/login", url.Values{"password": {"wrong"}})
	if !strings.Contains(w.Body.String(), "Too many attempts") {
		t.Errorf("6th attempt was not throttled:\n%s", w.Body.String())
	}
	// Even the correct password is refused while throttled.
	w = postForm(h, "/admin/login", url.Values{"password": {adminPassword}})
	if w.Code == http.StatusSeeOther {
		t.Error("throttling was bypassed by supplying the correct password")
	}
}

func TestLoginRedirectsToSafeNextOnly(t *testing.T) {
	_, h, _, _ := newTestServer(t)

	w := postForm(h, "/admin/login", url.Values{
		"password": {adminPassword},
		"next":     {"/admin/posts/new"},
	})
	if got := w.Header().Get("Location"); got != "/admin/posts/new" {
		t.Errorf("Location = %q, want /admin/posts/new", got)
	}

	w = postForm(h, "/admin/login", url.Values{
		"password": {adminPassword},
		"next":     {"https://evil.example.com/steal"},
	})
	if got := w.Header().Get("Location"); got != "/admin" {
		t.Errorf("offsite next: Location = %q, want /admin", got)
	}
}

// --- Drafts ---------------------------------------------------------------

func TestDraftsAreInvisibleToThePublic(t *testing.T) {
	_, h, posts, _ := newTestServer(t)
	ctx := context.Background()

	draft := &db.Post{Slug: "secret-draft", Title: "Secret Draft", BodyMD: "wip", BodyHTML: "<p>wip</p>"}
	if err := posts.Create(ctx, draft); err != nil {
		t.Fatal(err)
	}

	if w := get(h, "/blog/secret-draft"); w.Code != http.StatusNotFound {
		t.Errorf("anonymous GET of a draft: status %d, want 404", w.Code)
	}
	if w := get(h, "/blog"); strings.Contains(w.Body.String(), "Secret Draft") {
		t.Error("a draft title appeared on the public blog index")
	}
	if w := get(h, "/feed.xml"); strings.Contains(w.Body.String(), "Secret Draft") {
		t.Error("a draft appeared in the RSS feed")
	}

	// The author can see their own drafts.
	admin := signIn(t, h, "/admin/login", adminPassword)
	if w := get(h, "/blog/secret-draft", admin); w.Code != http.StatusOK {
		t.Errorf("admin GET of own draft: status %d, want 200", w.Code)
	}

	// Publishing makes it public.
	draft.Published = true
	if err := posts.Update(ctx, draft); err != nil {
		t.Fatal(err)
	}
	if w := get(h, "/blog/secret-draft"); w.Code != http.StatusOK {
		t.Errorf("published post: status %d, want 200", w.Code)
	}
}

// --- Authoring ------------------------------------------------------------

func TestCreatePostThroughTheAdminForm(t *testing.T) {
	_, h, posts, _ := newTestServer(t)
	admin := signIn(t, h, "/admin/login", adminPassword)

	w := postForm(h, "/admin/posts", url.Values{
		"title":     {"Hello, World!"},
		"summary":   {"A first post."},
		"body_md":   {"# Heading\n\n```go\nfunc main() {}\n```"},
		"published": {"1"},
	}, admin)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("create: status %d, want 303\n%s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Location"); got != "/blog/hello-world" {
		t.Errorf("Location = %q, want /blog/hello-world (slug derived from title)", got)
	}

	post, err := posts.GetBySlug(context.Background(), "hello-world", false)
	if err != nil {
		t.Fatalf("post was not stored: %v", err)
	}
	if !strings.Contains(post.BodyHTML, "chroma") {
		t.Error("body was not rendered with syntax highlighting at save time")
	}

	page := get(h, "/blog/hello-world").Body.String()
	if !strings.Contains(page, "Hello, World!") {
		t.Error("published post does not render its title")
	}
}

func TestDuplicateSlugIsRejectedWithoutLosingInput(t *testing.T) {
	_, h, posts, _ := newTestServer(t)
	admin := signIn(t, h, "/admin/login", adminPassword)

	if err := posts.Create(context.Background(), &db.Post{Slug: "taken", Title: "Taken"}); err != nil {
		t.Fatal(err)
	}

	w := postForm(h, "/admin/posts", url.Values{
		"title":   {"Taken"},
		"body_md": {"some words I do not want to retype"},
	}, admin)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "already uses the slug") {
		t.Error("no explanation of the slug conflict")
	}
	if !strings.Contains(body, "some words I do not want to retype") {
		t.Error("the rejected form lost what the author had typed")
	}
}

func TestProjectLinksMustBeHTTP(t *testing.T) {
	_, h, _, _ := newTestServer(t)
	admin := signIn(t, h, "/admin/login", adminPassword)

	w := postForm(h, "/admin/projects", url.Values{
		"title":    {"Sneaky"},
		"link_url": {"javascript:alert(1)"},
	}, admin)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422 for a javascript: link", w.Code)
	}
}

func TestPortfolioProjectVisibleToPortfolioSession(t *testing.T) {
	_, h, _, projects := newTestServer(t)

	err := projects.Create(context.Background(), &db.Project{
		Slug: "ledger", Title: "Ledger", Summary: "A thing I built",
		Tech: "Go, SQLite", Published: true, BodyHTML: "<p>case study</p>",
	})
	if err != nil {
		t.Fatal(err)
	}

	portfolio := signIn(t, h, "/portfolio/login", portfolioPassword)

	body := get(h, "/portfolio", portfolio).Body.String()
	if !strings.Contains(body, "Ledger") {
		t.Error("portfolio index does not list the project")
	}

	w := get(h, "/portfolio/ledger", portfolio)
	if w.Code != http.StatusOK {
		t.Fatalf("project detail: status %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "case study") {
		t.Error("project detail does not render the body")
	}
}

// --- Infrastructure -------------------------------------------------------

func TestHealthzChecksTheDatabase(t *testing.T) {
	_, h, _, _ := newTestServer(t)

	w := get(h, "/healthz")
	if w.Code != http.StatusOK || w.Body.String() != "ok" {
		t.Errorf("healthz = %d %q, want 200 \"ok\"", w.Code, w.Body.String())
	}
}

func TestPrivatePathsAreNotIndexable(t *testing.T) {
	_, h, _, _ := newTestServer(t)

	w := get(h, "/portfolio")
	if got := w.Header().Get("X-Robots-Tag"); !strings.Contains(got, "noindex") {
		t.Errorf("X-Robots-Tag = %q, want noindex on /portfolio", got)
	}
	if got := w.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store on a gated path", got)
	}

	robots := get(h, "/robots.txt").Body.String()
	for _, want := range []string{"Disallow: /admin", "Disallow: /portfolio"} {
		if !strings.Contains(robots, want) {
			t.Errorf("robots.txt missing %q", want)
		}
	}
}

func TestUnknownPathRendersStyled404(t *testing.T) {
	_, h, _, _ := newTestServer(t)

	w := get(h, "/no/such/page")
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Back to the front page") {
		t.Error("404 did not render the styled error page")
	}
}

func TestFeedIsValidRSS(t *testing.T) {
	_, h, posts, _ := newTestServer(t)

	err := posts.Create(context.Background(), &db.Post{
		Slug: "first", Title: "First", Summary: "Hello", Published: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	w := get(h, "/feed.xml")
	body := w.Body.String()

	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "rss") {
		t.Errorf("Content-Type = %q, want an RSS type", ct)
	}
	for _, want := range []string{
		`<?xml version="1.0"`, "<rss", "<channel>",
		"https://example.test/blog/first", // absolute links, built from BaseURL
	} {
		if !strings.Contains(body, want) {
			t.Errorf("feed missing %q\n%s", want, body)
		}
	}
}
