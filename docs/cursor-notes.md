# Cursor notes - Figma restyle (`feat/redesign`)

Session notes for the mattbx.co-final restyle. Not a product README.

## Visual direction

- Source: Figma [mattbx.co-final](https://www.figma.com/design/JfhTRsfprDuqpocHijFLXG/mattbx.co-final) (home open/closed, work detail, profile sidebar).
- Desktop-first Inter Variable + PP Fraktion Mono (Fraktion local/personal - not committed).
- Multi-theme via `html[data-theme]` (default, light, green, dark-green, blueish-2) - not `prefers-color-scheme`.
- Quiet shell: keep custom cursor; grain parked at opacity 0; corner crosses removed; CRT/phosphor not the dark look.

## What changed from the CRT / Jean Dawson pass

| Before | After |
|---|---|
| `/` = blog posts | `/` = public Work index |
| `/portfolio` password-gated | Public `/work/{slug}`; old URLs 301 |
| Fixed header: Posts / About / Portfolio | `mattbx.co` + clock + Work / Other + Profile |
| Mono hero + phosphor dark | Inter display stack + semantic theme tokens |
| About page | Profile drawer (390px) in the shell |
| Grain + corner crosses + CRT | Cursor only; grain off; corners deleted |

## Decisions locked

- Published projects are public; admin stays gated. Portfolio login routes left unused.
- Access control still only in `router.go`.
- Profile open by default on `/`; closed on `/work/…` and `/blog` (`?profile=1|0` overrides).
- Shell is a fluid 2-column grid: primary (header / scroll / footer) + profile aside; opening shrinks primary width.
- Headers pinned; primary and aside scroll independently.
- Footer pinned under the scroll pane from md up; below md it flows with content.
- Theme cookie + footer swatches; first paint from cookie in Go.
- `--cross-color` is cursor-only now (corners gone); mixes from `--text-primary` per theme.
- No media/upload pipeline, horse graphic, or full profile/resume sections this pass.

## Files touched (high level)

- Tokens / fonts: `internal/ui/tailwind/input.css`, `fonts.css`, `scripts/fetch-fonts.sh`, Inter `.woff2` + OFL.
- Shell / pages: `layout.templ`, `home.templ`, `portfolio.templ`, `about.templ`, `page.go`, `classes.go`.
- CSS / JS: `static/main.css`, `terminal.js`, `cursor.js`.
- Routes / handlers: `router.go`, `public.go`, `portfolio.go`, `server.go`, `admin.go`, `auth.go`, tests.
- Removed: `corners.templ`.
- Notes: `docs/cursor-notes.md`.

## Working

- Public Work home + detail; portfolio 301s; about → `/`.
- Profile column: gutter-aligned chrome, scramble on Profile control, independent scroll.
- Themes switch and persist; Inter loads; custom cursor visible again after `--cross-color` restore.
- Footer anchored on md+; below md rides with main content.
- Tests green for public work / redirects / profile defaults.

## Unfinished / fragile

- Profile body is bio + contact/LinkedIn only - Figma “Something” resume sections not filled.
- Work detail media slot is empty placeholder; no screenshots/uploads.
- PP Fraktion still drop-in on disk / Disco volume - stack falls back without it.
- Mobile is acceptable stack, not pixel-matched to Figma.
- Blog (“Other”) uses new shell/tokens only - no dedicated Figma frame.
- `SiteLinkedIn` in `page.go` may need a real URL check.
- CLAUDE.md still mentions corner crosses / older shell habits.

## Open questions

- Persist profile open/closed across navigations (cookie), or keep per-route defaults?
- How much real profile/resume copy vs placeholders?
- When to wire project media, and whether Micropub gets a media endpoint?
- Drop unused portfolio auth entirely in a later pass?
- Open a PR from `feat/redesign` when ready?
