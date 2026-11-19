# Plan 06 - REVIEW ONLY: site nav + profile drawer across breakpoints

Branch: `feat/redesign`. **This is a review and options task - no code changes.** Don't touch
the uncommitted work in the tree (`home.templ`, `main.css`, `static/noise.js`).

Write the findings to `docs/reviews/06-nav-and-profile-breakpoints.md`, then stop so Matt can
choose. A build plan follows from his choice.

## Why

Matt (8 Nov): the nav needs work, and the breakpoint variants were never designed. On small
screens it's broken:
- **Work / Other:** the "< Profile" control still shows with its arrow. Clicking it removes the
  item from the top nav and opens the profile **below all the content, including the footer**.
  Absolutely not intended.
- **Home:** same problem, but the profile is always open (home's default), so it sits under
  everything.
- The nav itself should be brought up to modern, good design standards at every size.

Matt's leaning (evaluate it seriously, but don't assume it): **the side drawer exists only on
desktop; on smaller screens Profile becomes a standard page in the nav.**

## Read first

- `docs/ia-decisions.md` - agreed IA and nav semantics (underline = current section only; home
  has no nav item; wordmark has no active state; `/work` 302s to the top project)
- `internal/ui/layout.templ` - header, nav, profile control, `ProfilePanel`
- `internal/ui/static/main.css` - shell grid, `.profile-panel`, breakpoints at `47.999rem`,
  `60rem`, `46rem` (x2); drawer motion via `@property --profile-w` (plan 02)
- `internal/ui/static/terminal.js` - `setProfileOpen`, toggle, `inert` / `aria-hidden`
- `internal/ui/profile.go` + `internal/ui/content/profile.md` - the resume content (one source)
- `internal/handlers/public.go` - `/about` currently 302s to `/` (leftover handler + `about.templ`
  exist and could be reused); per-route profile defaults and `?profile=1|0`
- Reference: https://8886592.cargo.site/ - on mobile its sidebar collapses to an "Information"
  link and columns stack (measured 7 Nov). Figma file mattbx.co-final is direction, not spec

## What to produce

1. **Current state, measured.** Screenshots at 375, 768, 1024, 1280 and 1512 wide, for `/`,
   a `/work/{slug}` page and `/blog`, drawer open and closed where it applies. List every
   problem you see: the profile-below-footer bug, header crowding, clock, tap targets, focus
   order, what screen readers get, scroll behaviour, anything else.
2. **Profile at each size - 2 to 4 options**, for example:
   - A. Drawer on desktop only; below a breakpoint, Profile is a normal page (`/profile` or
     reuse `/about`) and a normal nav item (Matt's leaning)
   - B. Drawer on desktop; full-screen overlay / sheet on small screens (same toggle)
   - C. Anything better you can justify
   For each: how it works at each breakpoint, URL and nav behaviour, how the home page's
   "open by default" works, what happens on resize across the breakpoint, accessibility (focus,
   `inert`, Escape, back button), effort, and risks. **Recommend the breakpoint** where the
   drawer should start - the main column needs room for the hero and work list beside a
   390px drawer.
3. **The nav itself - 2 to 3 options** at small, medium and large sizes. There are only three
   items (Work, Other, Profile) plus the wordmark and clock, so judge whether a menu button is
   needed at all or inline links are better. Cover: active state (underline semantics as
   agreed), tap target size, clock behaviour, pinned vs scrolling header, the scramble effect,
   focus styles, and how it reads in each theme.
4. **Recommendation:** one combined direction (nav + profile) with reasons, and an outline of
   the build steps it would need - small, ordered, each checkable. Flag anything that changes
   `docs/ia-decisions.md`.

## Constraints

- Go/templ, CSS-first, minimal first-party JS, no new libraries
- One content source for the profile (`content/profile.md`) whether it shows as drawer or page
- Keep the desktop drawer motion (plan 02) and the per-route defaults unless you argue to
  change them
- Access control stays only in `router.go` - any new public route is registered bare
- Reduced motion respected; no layout shift on load; works without JS where reasonable
- Figma is direction, not spec - improve on it
- House style: short hyphens, never em dashes, in the report and any commit messages

## Done when

`docs/reviews/06-nav-and-profile-breakpoints.md` exists with the measured current state,
profile options, nav options, a recommendation and a build outline. No code changed.
