# Plan 09 - Sticky-header fixes, no drawer on /profile, work page reflow

Branch: `feat/redesign`. From milestone review 07 (`docs/reviews/07-m3b-nav-and-profile.md` -
read sections 2 and 3). This is the **layout-correctness** half of the M3b polish. Interaction
polish (focus + Escape, focus rescue, tap targets, scramble on hover devices) is the next plan;
the small-screen home intro comes after that.

## The cause behind most of this

Since plan 07, below 64rem the **document** is the scroller and a 52px sticky header covers its
top edge. Anything that sticks to, or scrolls to, the top of the viewport now lands under the
header. Fix it once with a shared header-height value, not per element.

## Step 0

Commit `docs/reviews/07-m3b-nav-and-profile.md` on its own: "Review the M3b nav and profile
milestone".

## Steps

1. **One header-height token.** Add `--header-h: 3.25rem` (52px - today's header `min-height` in
   `main.css`, "16 + 20 + 16") and make the header's `min-height` use it, so the two can't drift.
   Put it where the file's other layout tokens live. (Plan 08's token review will fold it into
   the agreed set - just keep it a single, named value.)
2. **Anchors and the skip link (below 64rem).** `html { scroll-padding-top: var(--header-h); }`
   below 64rem (a small extra gap is fine if it reads better). Covers goldmark heading IDs, the
   skip link to `#main`, and resume section links.
   *Check:* at 375, `/blog/<post>#<heading>` and the skip link land just below the header.
3. **Sticky elements clear the header (below 64rem).** Offset `top` by `--header-h` for
   `.work-detail__nav` (today `top: 16px`) and `.rail` (today `top: var(--space-4)`) wherever they
   stick to the document. At >= 64rem they stick inside `.shell__scroll` and stay as they are.
   *Check:* work list at 1000 and post rail at 900 stick below the header, not behind it.
4. **No drawer on `/profile`.** `Layout` skips the `aside` (drawer) when `p.Nav == "profile"`, so
   the resume appears once: no duplicate section IDs, no duplicate HTML. In `terminal.js`, the
   Profile link only intercepts the click when the drawer element exists - on `/profile` at
   desktop it simply behaves as a link to the current page (it's already `aria-current`).
   *Check:* `/profile` has zero duplicate IDs at every width (count them in the browser); the
   Profile link on `/profile` at 1280 doesn't throw or toggle anything.
5. **Work page reflows by its own width.** Move `.work-detail`'s column changes from the
   `@media (max-width: 60rem)` viewport query to `@container` queries on `.shell__primary`
   (already `container-type: inline-size`). Pick thresholds in the browser so the three columns
   only appear when the media column is genuinely usable, with a two-column step if it helps.
   This also removes the 60rem viewport breakpoint behind step 3's work-list case - recheck it.
   *Check:* `/work/{slug}` at 1024 with the drawer open has a usable media column; 1280 and 1512
   open and closed; 768 and 375 stacked cleanly.
6. **Profile summary test.** `TestProfilePage` asserts the description text starts with
   "Product designer". Assert instead that a description is present and equals
   `ui.ProfileSummary()`, so editing the resume copy never fails a handler test.
7. **CLAUDE.md layout note** (short - the full CLAUDE.md refresh is M4). Add:
   - one shell breakpoint at 64rem: desktop two-column shell + drawer above; below it a normal
     page where the document scrolls under a sticky header
   - **anything sticky or anchored must clear `--header-h` below 64rem**
   - the drawer is desktop only; `/profile` is the page; no drawer on `/profile`
   - content breakpoints at 46rem; work detail reflows with container queries
   - house style: short hyphens, never em dashes - including commit messages
   - note that the spacing scale is px-named today (`gap-24` = 24px) until plan 08's migration

## Check (whole plan)

- `go tool templ generate && go build ./... && go vet ./... && go test ./...`
- Screenshot matrix: `/`, `/work/{slug}`, `/blog`, a post, `/profile` at 375, 768, 1000, 1024,
  1280, 1512 - nothing sticks or lands under the header; nothing below the footer
- Desktop drawer unchanged (open on `/`, closed elsewhere, animation, `?profile=`)
- One light and one dark theme
- Stop uncommitted (after step 0) so Matt can look

## Out of scope

Focus into/out of the drawer + Escape, focus rescue on resize, 44px tap targets, theme-swatch hit
areas, scramble on touch (next plan); small-screen home intro; token migration (plan 08).

House style: short hyphens, never em dashes - in code comments, docs and commit messages.
