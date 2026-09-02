# mattbx-go

<https://github.com/mattbx/mattbx-go>

A personal blog and portfolio. One Go binary, SQLite on a volume, no build
pipeline and no JavaScript.

- **`/`** and **`/blog`** — public writing, with an RSS feed at `/feed.xml`
- **`/portfolio`** — password-gated work, shared selectively
- **`/admin`** — password-gated authoring for both

Blog and portfolio use **separate passwords**, so handing someone the portfolio
password never gives them write access.

## Stack

| Layer      | Choice                                         |
|------------|------------------------------------------------|
| Templates  | [templ](https://templ.guide) — compiled to Go  |
| Database   | SQLite via `modernc.org/sqlite` (pure Go)      |
| Markdown   | goldmark + Chroma, rendered once at save time  |
| Sanitizing | bluemonday, applied before anything is stored  |
| Routing    | stdlib `net/http` — no framework               |
| Hosting    | [Disco](https://disco.cloud/docs/) via git push |

## Running it locally

```bash
./scripts/local.sh
```

That's the whole setup. On first run it creates `.env` from `.env.example` and
generates a `SESSION_SECRET`, then serves <http://localhost:8383> with hot
reload — edit a `.templ`, `.go`, or `.css` file and the browser refreshes.

`templ` and `air` are pinned as Go tool dependencies in `go.mod`, so nothing
needs installing globally. Go 1.24+ is required for that.

Set your own passwords in `.env` before doing anything real:

```
ADMIN_PASSWORD=...
PORTFOLIO_PASSWORD=...
```

They must differ — the server refuses to start otherwise.

## Tests

```bash
go test ./...
```

The suite covers the parts worth protecting: session signing and scope
isolation, open-redirect rejection, login rate limiting, drafts staying
invisible to the public, and Markdown sanitization.

## Layout

```
cmd/web/            entry point, graceful shutdown
internal/config/    env loading, fail-fast validation
internal/db/        connection, embedded migrations, stores
internal/auth/      session cookies, middleware, rate limiting
internal/markdown/  Markdown -> sanitized HTML
internal/handlers/  routes; all access control lives in router.go
internal/ui/        templ components + embedded CSS
```

Generated `*_templ.go` files are **not** committed. `scripts/local.sh` and the
Dockerfile both run `templ generate`, so a fresh clone needs one of those (or a
bare `go tool templ generate`) before `go build` will work.

To restyle syntax highlighting, change the theme names in
`internal/ui/static/gen/main.go` and run `go generate ./internal/ui/...`.

Site name, role, and tagline are constants at the top of
`internal/ui/page.go`.

## Deploying

The repo carries a `Dockerfile` and a `disco.json`; Disco needs nothing else.

```bash
disco projects:add --name mattbx --domain your.domain --github mattbx/mattbx-go
```

Then set the secrets, which are deliberately absent from the image:

```bash
disco env:set ADMIN_PASSWORD=... PORTFOLIO_PASSWORD=... SESSION_SECRET=... BASE_URL=https://your.domain ENV=production
```

Generate the secret with `openssl rand -hex 32`. After that, `git push` deploys.

**The database lives on the `sqlite-data` volume mounted at `/data`.** Back it
up with:

```bash
disco volumes:export --project=mattbx --volume=sqlite-data > backup.tar.gz
```

### A note on migrations

Migrations run at startup and are tracked in `schema_migrations`. Disco does
zero-downtime deploys, which briefly runs the old and new containers against
the same volume, so **keep migrations additive** — the previous version has to
survive the new schema for a few seconds. For a destructive change, use Disco's
`hook:deploy:start:before` service instead, which aborts the deploy if the
migration fails.
