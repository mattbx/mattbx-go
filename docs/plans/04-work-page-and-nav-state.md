# Plan 04 - A real Work page + correct nav active states

Branch: `feat/redesign`. Agreed with Matt 7 Nov. See `docs/ia-decisions.md` (updated with this).

## The intent (the build drifted from the Figma here)

- The nav underline means **"you are in this section"** - nothing else.
- **Home has no nav item.** The `mattbx.co` wordmark links home and never shows an active state.
- **Work** has its own page and is underlined whenever you're on it. **Other** stays as it is
  (already underlined on the blog list and posts).
- The work page's left column is titled **"Work"** - a plain title, not a link and not a back
  button. It lines up with the other column titles (client, project) as a row of headings.

## Steps

1. **Home stops claiming "work".** In `internal/handlers/public.go`, the home handler calls
   `s.page(r, "", ui.SiteTagline, "work")` - change the nav value to `""`. Home then shows no
   underlined nav item.
2. **Nav link.** In `internal/ui/layout.templ`, point the Work nav item at `/work` (it's `/` today).
   Keep `current(p.Nav == "work")` - work detail already passes `"work"`.
3. **`/work` opens the top project.** Add `GET /work` (and `GET /work/{$}` if needed for the
   trailing slash) in `internal/handlers/router.go`, registered bare (public) - access control
   stays only in `router.go`. The handler:
   - lists projects the same way the work detail nav does (`s.projects.List(ctx, s.isAdmin(r))` -
     ordered by `sort_order`, so "first" = the row shown as 01)
   - **302** (`http.StatusFound`, not 301 - the top project can change) to `/work/{slug}` of the
     first one
   - no projects: redirect to `/` (or render the existing empty state - pick the simpler)
4. **"Work" title, not "← Work" link.** In `internal/ui/portfolio.templ` (`WorkDetail`), replace
   `<a class="work-detail__back ..." href="/">← Work</a>` with a plain heading "Work" that uses the
   **same type style as the overview heading** next to it (today `font-ui text-section
   font-semibold tracking-tight`). Not a link. Use a sensible heading level for the page outline
   (the project title stays the page's `h1`). Remove `.work-detail__back` styles that become
   dead. Align the three column titles on one baseline if it's a small change; if it isn't, note
   it for the work-page styling pass rather than forcing it.
5. **Tests.** Add/adjust handler tests:
   - `GET /work` → 302 to `/work/{first slug}`; with no projects → the chosen fallback
   - home does not mark Work current; work detail does
6. Commit `docs/ia-decisions.md` (already updated in the working tree) with this change.

## Check

- `go tool templ generate && go build ./... && go vet ./... && go test ./...`
- In the browser:
  - Home: no nav item underlined; wordmark has no active state
  - Click **Work** in the nav → lands on the 01 project at its own URL, Work underlined, that row
    marked current in the list
  - Click a row on home → that project, Work underlined
  - Click another project in the work nav → that project, still Work underlined
  - Other: underlined on `/blog` and on a post, as before
  - The left column reads "Work" as a title (no arrow, not clickable)
- Stop uncommitted so Matt can look

## Out of scope

Work page styling/layout finesse (separate pass), client/format/image fields (IA build plan),
renaming `/blog`.
