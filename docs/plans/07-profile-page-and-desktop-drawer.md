# Plan 07 - Profile page on small screens, drawer on desktop only (fixes the below-footer bug)

Branch: **`feat/redesign`** (Matt is parking the noise-overlay spike first - check you're on
`feat/redesign` and the tree is clean apart from the docs listed in step 0 before starting).

Agreed with Matt 8 Nov, from review 06 (`docs/reviews/06-nav-and-profile-breakpoints.md` - read
sections 1, 2A, 2C and 4). This plan is the **bug-fix part** - review build steps 1-5 and 7.
Polish (focus/Escape, tap targets, work page container queries, scramble on touch) and the
small-screen home intro come in later plans. `docs/ia-decisions.md` is already updated for this.

## The intent

- **One breakpoint, 64rem (1024px), separates two modes.**
  - **>= 64rem:** today's desktop shell - two columns, the animated drawer (plan 02), per-route
    defaults (open on `/`, closed elsewhere), `?profile=1|0`
  - **< 64rem:** a normal stacked document. No drawer at all, whatever `profile-open` says. The
    page scrolls like any website, header sticky
- **Profile is a real page at `/profile`** at every width (the server can't know the width, so it
  must work everywhere - shareable, indexable, works without JS). Same resume, same source
  (`internal/ui/content/profile.md`)
- **One Profile control** in the nav: a link to `/profile`. On desktop with JS it opens the
  drawer instead; on small screens or without JS it just navigates
- Nav underline rules stay: Profile is underlined on `/profile`; the open drawer is never shown
  with an underline (it isn't a section)

## Step 0 - housekeeping

- Add `docs/reviews/*-screens/` to `.gitignore` (local screenshots, public repo)
- Commit `docs/plans/06-review-nav-and-profile-breakpoints.md`,
  `docs/reviews/06-nav-and-profile-breakpoints.md` and the updated `docs/ia-decisions.md` on
  their own: "Review nav and profile across breakpoints"

## Steps

1. **Extract `ProfileBody`** from `ProfilePanel` in `internal/ui/layout.templ`: intro, links row,
   sections. The drawer renders `ProfileBody`. No visual change.
   *Check:* `go test ./...`; drawer identical at 1280.
2. **Add `GET /profile`**, registered bare in `internal/handlers/router.go` (access control stays
   only there). New `ui.ProfilePage` renders `ProfileBody` as the page body inside the normal
   layout, with `Nav: "profile"` and the drawer closed. Sensible page title ("Profile") and a
   description from the resume summary.
   *Check:* handler test - 200, contains a resume section heading, Profile item marked current.
3. **`/about` → `/profile`** with a **301**. Delete `about.templ` (and anything only it used).
   Fix the stale `server.go` comment about `/about`.
   *Check:* test that `/about` 301s to `/profile`.
4. **One Profile control.** Replace the header launch `<button>` with
   `<a href="/profile" class="... profile-control profile-control--launch" data-profile-toggle ...>`
   and move it **inside** `<nav aria-label="Primary">` with Work and Other. In `terminal.js`,
   the click handler calls `preventDefault()` and toggles the drawer **only when**
   `matchMedia("(min-width: 64rem)").matches`; otherwise the link navigates. Keep
   `aria-expanded` / `aria-controls` on it (ARIA allows them on links). The close control inside
   the drawer chrome stays a button (desktop only). Keep the scramble attribute.
   *Check:* below 1024 it navigates to `/profile`; at 1024+ it toggles; with JS off it navigates
   at every width.
5. **Gate the drawer to >= 64rem in CSS** (`internal/ui/static/main.css`):
   - move `html.profile-open .shell`, the panel's visible/open state and the launch-control
     hiding into `@media (min-width: 64rem)`
   - below 64rem: `.profile-panel { display: none; }` and a single-column shell regardless of
     the class
   - **delete the 60rem stacked split block** (the cause of the bug)
   - keep plan 02's drawer motion working at >= 64rem
   *Check:* home at 375 and 768 has no profile pane; 1024+ unchanged; no layout shift on load at
   any width.
6. **Document scrolling below 64rem.** Change today's `47.999rem` "phone" block to apply below
   64rem; let `html` / `body` scroll there (lift `overflow: hidden`), make the header
   `position: sticky; top: 0`, keep the footer in normal flow at the end. Keep the pinned
   two-column shell at >= 64rem. Content breakpoints at `46rem` (ledger / leaf) stay as they are.
   *Check:* at 375 and 768 the whole page scrolls; header stays put; footer is last.

## Check (whole plan)

- `go tool templ generate && go build ./... && go vet ./... && go test ./...`
- Screenshot matrix (as in review 06): `/`, `/work/{slug}`, `/blog`, `/profile` at 375, 768,
  1024, 1280, 1512. Confirm:
  - **nothing ever renders below the footer**
  - < 1024: Profile is a nav link → `/profile`, underlined there; no drawer anywhere
  - >= 1024: drawer behaves exactly as before (open on `/`, closed elsewhere, animation,
    `?profile=`), Profile underlined only on `/profile`
  - resizing across 1024 with the drawer open doesn't break the layout
  - `/about` 301s to `/profile`
- One light theme and one dark theme
- Stop uncommitted (after step 0's docs commit) so Matt can look

## Out of scope (later plans)

Focus moving into/out of the drawer + Escape; focus rescue on resize; 44px tap targets and
theme-swatch hit areas; the `gap-6` question; work page container queries at 1024; scramble on
touch; the small-screen home intro; CLAUDE.md breakpoint notes.

House style: short hyphens, never em dashes - in code comments, docs and commit messages.
