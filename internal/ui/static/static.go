// Package static embeds and serves the site's CSS.
//
// Assets ship inside the binary, so there is nothing to copy into the
// container and no external requests at runtime.
package static

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"net/http"
	"sync"
)

//go:embed *.css *.svg
var files embed.FS

// FS exposes the embedded assets.
func FS() fs.FS { return files }

var (
	hashOnce sync.Once
	hashes   map[string]string
)

// buildHashes fingerprints each asset once at startup so URLs can be cached
// forever and still change the moment the file does.
func buildHashes() {
	hashes = make(map[string]string)
	entries, err := files.ReadDir(".")
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := files.ReadFile(e.Name())
		if err != nil {
			continue
		}
		sum := sha256.Sum256(b)
		hashes[e.Name()] = hex.EncodeToString(sum[:])[:12]
	}
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
		if r.URL.Query().Get("v") != "" {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=300")
		}
		fileServer.ServeHTTP(w, r)
	}))
}
