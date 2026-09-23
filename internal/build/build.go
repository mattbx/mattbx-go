// Package build holds values stamped in at image-build time, so the running
// binary can report when it was built without a database lookup or any
// runtime configuration.
//
// Only the build date is stamped, not a git commit hash: .dockerignore
// deliberately excludes .git from the build context, so the Dockerfile has
// no commit to read. The date is computed inside the Dockerfile's own build
// step instead (see its final `go build` -ldflags), so nothing needs passing
// in from outside — it stays correct regardless of how the image gets built
// (Disco, a local `docker build`, CI, whatever comes later).
package build

// Date is overridden at build time via -ldflags. The default covers
// `go run`/air locally, where nothing sets it.
var Date = "dev build"
