package static

import (
	"mime"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestURLFingerprintsEmbeddedAssets(t *testing.T) {
	want := regexp.MustCompile(`^/static/[\w./-]+\?v=[0-9a-f]{12}$`)
	for _, name := range []string{"main.css", "terminal.js", "fonts/README.txt"} {
		if got := URL(name); !want.MatchString(got) {
			t.Errorf("URL(%q) = %q, want a fingerprinted /static path", name, got)
		}
	}
}

func TestURLLeavesUnknownAssetsUnfingerprinted(t *testing.T) {
	if got, want := URL("missing.css"), "/static/missing.css"; got != want {
		t.Errorf("URL(missing.css) = %q, want %q", got, want)
	}
}

func TestHandlerCachesFontsForAWeek(t *testing.T) {
	rec := httptest.NewRecorder()
	Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/static/fonts/README.txt", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got, want := rec.Header().Get("Cache-Control"), "public, max-age=604800"; got != want {
		t.Errorf("Cache-Control = %q, want %q", got, want)
	}
}

func TestWoff2HasFontMIMEType(t *testing.T) {
	if got, want := mime.TypeByExtension(".woff2"), "font/woff2"; got != want {
		t.Errorf("TypeByExtension(.woff2) = %q, want %q", got, want)
	}
}
