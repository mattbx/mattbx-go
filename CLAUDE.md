# mattbx-go

Personal blog + portfolio. One Go binary, templ for UI, SQLite on a volume,
self-hosted on [Disco](https://disco.cloud/docs/). No JS, no CSS framework,
no ORM.

## Commands

- `./scripts/local.sh` — the only command needed to run locally. Creates `.env`,
  generates `SESSION_SECRET`, serves http://localhost:8383 (app on :8080, Air
  proxy adds live reload).
- `go test ./...` — full suite.
- `go tool templ generate` — **required before `go build`** on a fresh clone.
- `go generate ./internal/ui/...` — regenerate `chroma.css` after changing the
  syntax-highlighting themes.

`templ` and `air` are pinned as tool dependencies in `go.mod`. Always invoke
them as `go tool templ` / `go tool air` — they are not installed globally.

## Hard constraints

These will break the deploy if violated:

- **Keep `CGO_ENABLED=0`.** The SQLite driver is `modernc.org/sqlite` (pure Go).
  Do not swap in `mattn/go-sqlite3` — it needs cgo and breaks the static build.
- **Keep migrations additive.** Disco's zero-downtime deploy briefly runs the
  old and new containers against the same volume, so the previous version must
  survive the new schema. Use `hook:deploy:start:before` in `disco.json` for
  anything destructive.
- **Keep the Alpine runtime base.** Disco's health check runs a shell command
  inside the container; distroless/scratch has no shell or `wget`.
- **Never put secrets in `disco.json`** — it is committed. Use `disco env:set`.

## Conventions

- **Access control lives only in `internal/handlers/router.go`.** A route is
  either registered bare (public) or wrapped in `requireAdmin`/`requirePortfolio`.
  Handlers never re-check permissions, so there is one place to audit.
- `Page.IsAdmin` is presentation only — it reveals drafts and edit links. It
  never grants access.
- **Markdown is rendered and sanitized at save time**, stored in `body_html`.
  The public read path is one query and no parsing. Never render user content
  at request time.
- `*_templ.go` is gitignored. Generated fresh by `scripts/local.sh` and the
  Dockerfile.
- CSS is hand-written and token-driven (`internal/ui/static/main.css`). Define
  colors on bare `:root`; the dark block only swaps values. Do not add Tailwind
  or any build step.
- Site name, role, and tagline are constants at the top of `internal/ui/page.go`.

## Gotchas found the hard way

- **Chroma emits CSS only for tokens a theme explicitly defines.** Light and
  dark rules must *both* be scoped in media queries, or tokens the dark theme
  leaves at default keep the light theme's foreground and become unreadable.
  See `internal/ui/static/gen/main.go`.
- **templ HTML-escapes apostrophes** to `&#39;`. Don't assert on raw
  apostrophes in tests — pick a substring without one.
- `modernc.org/sqlite` round-trips `time.Time` as RFC3339 text on `DATETIME`
  columns, and `NULL` into `sql.NullTime`. Pinned by `internal/db/driver_test.go`.
- Goldmark's `WithHardWraps()` turns every wrapped source line into a `<br>`.
  Wrong for prose; deliberately not enabled.
- `ADMIN_PASSWORD` and `PORTFOLIO_PASSWORD` must differ — startup refuses
  otherwise, so a portfolio visitor can never reach `/admin`.

## Milestone reviews

At the end of a major session or milestone, write a review entry into
[LEARNINGS.md](LEARNINGS.md): **what we did, what we learnt, issues, ideas.**
Newest entry first.

Trigger it either way:

- **Matt asks** — "let's do a milestone review", or similar.
- **Claude proposes it** when the moment fits: a feature landed with tests
  passing, a deploy is imminent, several non-obvious findings have piled up, or
  the session is clearly wrapping up.

Propose it in one line and accept a "no" — never interrupt work in progress to
suggest it, and never write an entry without saying so. Prefer specifics that
would save a future session real time (a wrong assumption, a subtle bug and its
signature, a decision and its reasoning) over a changelog of files touched;
`git log` already covers that. When a learning is a durable rule rather than a
one-off, promote it into the sections above as well.
