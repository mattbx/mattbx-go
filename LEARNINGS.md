# Learnings

A running log of milestone reviews. **Newest first.**

Each entry covers four things: **what we did**, **what we learnt**, **issues**,
**ideas**. See [CLAUDE.md](CLAUDE.md#milestone-reviews) for when to write one.

The bar for an entry: would this save a future session real time? A wrong
assumption, a subtle bug and how it announced itself, a decision and why it
went that way. Not a list of files touched — `git log` has that.

---

## 2026-09-02 — Initial scaffold

Greenfield → a working, deploy-verified blog and portfolio.

### What we did

- Go 1.26 + templ + SQLite, stdlib `net/http` routing, no framework and no JS.
- Public blog with RSS; `/portfolio` and `/admin` behind **separate** passwords
  using signed stateless cookies (no sessions table, so logins survive deploys).
- Markdown posts and portfolio projects with full CRUD in `/admin`.
- Hand-written token-driven CSS ("marginalia" — a mono rail carrying dates and
  metadata beside the prose), light/dark, no build step.
- `scripts/local.sh` + Air hot reload; `Dockerfile` + `disco.json` for Disco.
- 36 tests, weighted toward access control and sanitization.

### What we learnt

- **Chroma only emits CSS for tokens a theme explicitly defines.** Generating
  light rules unscoped and dark rules inside a media query looks correct but
  isn't: any token `github-dark` leaves at default keeps `github`'s near-black
  foreground. Scope *both* themes. `prefers-color-scheme: light` also matches
  "no preference", so light stays the default.
- **`styles.Get()` returns a fallback, not `nil`,** for an unknown Chroma style
  name — so a nil check silently passes and you get the wrong theme. The
  generator now compares `style.Name` against what was asked for.
- **Verify driver behaviour before designing a schema around it.** A 30-second
  probe test confirmed `modernc.org/sqlite` round-trips `time.Time` as RFC3339
  and `NULL` into `sql.NullTime` — which avoided inventing Unix-integer columns
  and conversion helpers for a problem that didn't exist. The probe stayed as
  `internal/db/driver_test.go`.
- Multi-statement `Exec` works fine on this driver, so migration files can hold
  several statements without splitting them.
- Air's proxy handles browser reload on its own, so `templ generate --watch` is
  unnecessary — plain `templ generate` in the Air build command is enough.

### Issues

- **Air's live-reload script is inline**, so a strict CSP with `script-src
  'none'` silently kills hot reload. The CSP is therefore applied only when
  `ENV != development`. If reload ever breaks, check the CSP first.
- **Disco's health check runs a shell command inside the container**, which
  ruled out distroless/scratch. Alpine + busybox `wget` is the runtime base.
  Running as root also sidesteps ownership problems on a fresh Disco volume.
- Goldmark's `WithHardWraps()` was enabled at first — it turns every wrapped
  source line into a `<br>`, which is wrong for prose. Removed.
- A test failed on `isn't here` because **templ escapes apostrophes** to
  `&#39;`. The page was correct; the assertion was wrong.
- Browser automation `left_click` on the login button didn't submit the form
  (nothing reached the server). Not an app bug — curl and the handler tests
  both pass. Submit via JS when driving forms in the preview pane.

### Ideas

Not built, deliberately — worth considering later:

- **Draft preview links.** Drafts are currently visible only to a signed-in
  admin. A signed, expiring preview URL would let you share one for feedback
  without handing over the admin password.
- **Live Markdown preview** in the editor (a small POST that re-renders the
  pane). Deferred to keep the scaffold JS-free.
- **Tags for posts** — the rail was designed with a `rail__tags` slot already.
- **Backup automation.** `disco volumes:export` is manual today; a cron service
  in `disco.json` could push a dated dump somewhere off-box.
- **`sqlc`** if the hand-written SQL in `internal/db` grows past comfort. Not
  worth a second code generator yet.
- Sitemap. `robots.txt` currently points at the RSS feed, which is a stand-in.
