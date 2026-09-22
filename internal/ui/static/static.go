// Package static embeds and serves the site's CSS, JavaScript and fonts.
//
// Assets ship inside the binary, so there is nothing to copy into the
// container and no external requests at runtime.
package static

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"mime"
	"net/http"
	"strings"
	"sync"
)

//go:embed *.css *.js *.svg fonts
var files embed.FS

func init() {
	// Go's built-in table has no woff2 entry and the runtime image ships no
	// mime.types, so register it explicitly.
	_ = mime.AddExtensionType(".woff2", "font/woff2")
}

// FS exposes the embedded assets.
func FS() fs.FS { return files }

var (
	hashOnce sync.Once
	hashes   map[string]string
)

// buildHashes fingerprints each asset once at startup so URLs can be cached
// forever and still change the moment the file does. It walks subdirectories
// so assets embedded later (fonts, media) are fingerprinted under their
// relative path, e.g. "fonts/x.woff2".
func buildHashes() {
	hashes = make(map[string]string)
	_ = fs.WalkDir(files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := files.ReadFile(path)
		if err != nil {
			return nil
		}
		sum := sha256.Sum256(b)
		hashes[path] = hex.EncodeToString(sum[:])[:12]
		return nil
	})
}

// URL returns the cache-busting path for an embedded asset, e.g.
// "/static/main.css?v=1a2b3c4d5e6f".
func URL(name string) string {
	hashOnce.Do(buildHashes)
	if v, ok := hashes[name]; ok {
		return "/static/" + name + "?v=" + v
	}
	return "/static/" + name
}

// Handler serves the embedded assets under /static/.
//
// The fingerprint in the URL makes the content immutable, so a long max-age is
// safe: a changed file gets a different URL.
func Handler() http.Handler {
	fileServer := http.FileServer(http.FS(files))
	return http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Query().Get("v") != "":
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		case strings.HasPrefix(r.URL.Path, "fonts/"):
			// Fonts are referenced from CSS, so there is no fingerprint to bust
			// the cache. They change rarely; a week keeps them warm.
			w.Header().Set("Cache-Control", "public, max-age=604800")
		default:
			w.Header().Set("Cache-Control", "public, max-age=300")
		}
		fileServer.ServeHTTP(w, r)
	}))
}
