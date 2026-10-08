# mattbx-go

Personal blog + portfolio. One Go binary, templ for UI, SQLite on a volume,
self-hosted on [Disco](https://disco.cloud/docs/). One small first-party script, Tailwind v4 via the standalone
CLI (no Node), no ORM.

Repo: <https://github.com/mattbx/mattbx-go>

## Commands

- `./scripts/local.sh` — the only command needed to run locally. Creates `.env`,
  generates `SESSION_SECRET`, serves http://localhost:8383 (app on :8080, Air
  proxy adds live reload).
- `go test ./...` — full suite.
- `go tool templ generate` — **required before `go build`** on a fresh clone.
- `go generate ./internal/ui/...` — regenerate `chroma.css` after changing the
  syntax-highlighting themes.

`templ` and `air` are pinned as tool dependencies in `go.mod`. Always invoke
them as `go tool templ` / `go tool air` — they are not installed globally. The
exception is Tailwind's standalone CLI (`brew install tailwindcss`); its output,
`internal/ui/static/tailwind.css`, is gitignored and built by
`scripts/local.sh` and the Dockerfile.

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
- **`main` is wired to production.** Disco deploys on push, with no staging
  step, so do the work on a branch and merge deliberately. The repo is public:
  anything committed is effectively permanent.

## Conventions

- **Access control lives only in `internal/handlers/router.go`.** A route is
  either registered bare (public) or wrapped in `requireAdmin`/`requirePortfolio`.
  Handlers never re-check permissions, so there is one place to audit.
  `/micropub` follows the same pattern with its own bearer-token middleware
  (`requireMicropubToken` in `micropub.go`) instead of a cookie scope, since
  it's a machine API for IndieWeb clients, not a browser session.
- `Page.IsAdmin` is presentation only — it reveals drafts and edit links. It
  never grants access.
- **Markdown is rendered and sanitized at save time**, stored in `body_html`.
  The public read path is one query and no parsing. Never render user content
  at request time.
- `*_templ.go` is gitignored. Generated fresh by `scripts/local.sh` and the
  Dockerfile.
- **Colors, type scale, and font stacks live in `internal/ui/tailwind/input.css`.**
  Role-based color tokens in `@theme static` (light defaults); dark overrides
  under `@layer theme` inside `@media (prefers-color-scheme: dark)`:
  `--color-page` (bg), `--color-fg` (terminal foreground — graphite light /
  phosphor dark), `--color-dim` (meta/secondary), `--color-line` (hairlines),
  `--color-surface` (sunk panels), `--color-danger` (destructive admin only).
  CRT source values (`--crt-void`, `--crt-phosphor`, …) stay outside `@theme`
  for the overlay layer.   Fluid type tokens (`--text-hero` … `--text-micro`) and gutters
  (`--spacing-gutter`, `--spacing-col`) are clamp/vw; tune via `:root`
  knobs `--fluid-frame` / `--fluid-scale` and per-step `--fluid-*-min|at|max`
  (preferred size in px at the design frame — no hand calc). Font stacks:
  `--font-ui` / `--font-prose` / `--font-mono`.
  `main.css` consumes with `var()` and must not redefine these on `:root`.
  Tailwind v4 is hybrid: shell/page chrome are utilities; prose, leaf/rail,
  ledger, forms, and gate stay in `main.css` until further migration. Preflight
  is not imported yet; unlayered `main.css` beats utilities — delete legacy
  rules when moving to utilities, never paper over with `!`. Shared utility
  clusters in `internal/ui/classes.go` must stay in `@source`. Opt-in motion:
  `.hover-underline`, `.hover-fill` (fg/page invert), `data-scramble`,
  `data-magnetic`, CRT overlay — class/`data-*` gated, not global. Site-wide
  film grain is `canvas.grain` + `terminal.js`, tuned via `:root` `--grain-*`
  (opacity/blend/density/scale/fps; light + dark defaults). Corner crosses
  (`@CornerCrosses()`) use `--cross-color` / `--cross-arm` / `--cross-arm-rest`.
  Motion timings: `--ease-wipe`, `--ease-underline`, `--dur-fast`, `--dur-fill`.
- **`chroma.css` is syntax highlighting only.** Generated from Chroma's
  `github` / `github-dark` styles by `go generate ./internal/ui/...`. It is a
  separate palette from the site tokens; do not fold it into `@theme`. Code
  block backgrounds are overridden in `main.css` via `.prose pre.chroma` to
  `--color-surface`.
- **The CSP is `script-src 'self'` in production only.** Inline scripts and
  inline `style=""` attributes are blocked there, and development sends no CSP
  (Air injects an inline reload script), so an effect can work locally and fail
  in prod. Check UI changes against a build running with `ENV=production`.
  Setting styles from JS through `el.style` is fine; only markup is restricted.
- Site name, role, and tagline are constants at the top of `internal/ui/page.go`.
- **Tailwind IntelliSense for `.templ`:** workspace `.vscode/settings.json` maps
  `templ` → `html`, points `tailwindCSS.experimental.configFile` at
  `internal/ui/tailwind/input.css`, and turns on string quick-suggestions. Needs
  the `a-h.templ` and Tailwind CSS IntelliSense extensions; standalone CLI is
  fine (no npm) on recent IntelliSense.

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
- **Global `a { color: … }` beats `text-*` utilities.** Unlayered link color
  in `main.css` wins over Tailwind. Chrome that should stay dim (footer,
  admin bar) needs a matching unlayered hook, not only a utility class.
- **Unknown or unsourced Tailwind classes fail silently.** A typo, a class
  only present in an unsourced Go string, or a theme token never defined in
  `@theme` produces no CSS and no build error — check the generated
  `tailwind.css` when something "doesn't apply."
- **Scramble + magnetic share a label node.** Put visible text in
  `[data-label]` inside `[data-magnetic]`; scramble targets that child so it
  does not wipe the wrapper DOM when rewriting `textContent`.
