package static

import (
	"regexp"
	"testing"
)

func TestURLFingerprintsEmbeddedAssets(t *testing.T) {
	want := regexp.MustCompile(`^/static/[\w./-]+\?v=[0-9a-f]{12}$`)
	for _, name := range []string{"main.css", "terminal.js"} {
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
