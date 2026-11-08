# Review 07 - M3b milestone: nav + profile

Milestone review of the nav and profile work, checked against
`docs/reviews/06-nav-and-profile-breakpoints.md` (review 06) and
`docs/plans/07-profile-page-and-desktop-drawer.md` (plan 07). Review only - no code changed.

Measured 8 Nov 2026 on `feat/redesign` at `5148f2f`, using the dev app on :8080 with Playwright.
Commits in scope:

- `3e6123e` Add /profile page and limit the drawer to desktop (plan 07)
- `43908fe` Scramble the wordmark on hover like the nav links
- `5148f2f` Center the header clock between the wordmark and nav

Local screenshots: `docs/reviews/07-screens/` (gitignored).

---

## 1. Done

### Plan 07, step by step

| Step | Status | Evidence |
|---|---|---|
| 0. Housekeeping: gitignore screenshots, commit review docs | Done | `b4ba49d`, `d505d1e` |
| 1. Extract `ProfileBody` | Done | `layout.templ`; drawer unchanged at 1280 |
| 2. `GET /profile`, bare route, `ui.ProfilePage` | Done | 200 at all widths; `Nav: "profile"`; summary as meta description; tests `TestProfilePage`, `TestProfileNavLink` |
| 3. `/about` -> `/profile` (301), delete `about.templ` | Done | `TestAboutRedirectsToProfile`; `SiteBio` removed with it; `server.go` comment fixed |
| 4. One Profile control, inside `<nav>` | Done | `<a href="/profile" data-profile-toggle>`; toggles at >= 64rem, navigates below and without JS (both measured); `aria-expanded`/`aria-controls` only added by JS on desktop |
| 5. Drawer gated to >= 64rem, 60rem stacked split deleted | Done | `.profile-panel` is `display: none` below 64rem at 375 and 768 on every page; 1024+ unchanged |
| 6. Document scrolling below 64rem | Done | document scrolls, header sticky (`top: 0` after scrolling 1500px at 375), footer last |

Whole-plan checks: `go build`, `go vet` and `go test ./...` pass. The screenshot matrix (5 widths
x 4 pages) shows nothing below the footer. Resizing across 1024 with the drawer open recovers
cleanly. Dark and light themes look right. A production build (`ENV=production`, strict CSP)
works with no console errors.

### Review 06 problems, status

| # | Problem (review 06, section 1) | Status |
|---|---|---|
| 1 | Profile pane below the footer (60rem stacked mode) | **Fixed** - mode deleted |
| 2 | Home "open by default" decided server-side | **Fixed** - the open state only has CSS from 64rem |
| 3 | Work detail cramped beside the drawer at 1024 | Open - polish |
| 4 | Focus dropped on every toggle | Open - polish (still `<body>` after opening, measured) |
| 5 | No Escape | Open - polish (measured) |
| 6 | Drawer last in tab order | Open - polish (mitigated once focus moves into the drawer on open) |
| 7 | Document never scrolls | **Fixed below 64rem**; desktop keeps the pinned shell by design |
| 8 | Small hit areas | Open - polish (nav items 16px tall, wordmark 12px, swatches 12x12) |
| 9 | `gap-6` = 6px in the nav | **Fixed** - `gap-24`; all three items 24px apart at every width |
| 10 | Header shifts when the profile toggles | **Fixed on mobile** (no toggle there); desktop shift is expected |
| 11 | Profile outside the nav landmark | **Fixed** |
| 12 | Two "Profile" controls for one panel | Now a link plus the drawer's close button; fine |
| 13 | Scramble runs on tap | Open - polish; now applies to the wordmark too |
| 14 | No-JS profile | **Fixed** - the link navigates |
| 15 | Leftovers (`server.go` comment, `about.templ`, zero-width hidden drawer) | **Fixed** - the hidden drawer is `display: none` below 64rem |

### Beyond the plan

- Nav items spaced evenly at 24px (replaced a one-off 24px margin before Profile).
- Wordmark scrambles on hover, with no underline (it has no active state, per IA).
- Header clock (768px+) is centered between the wordmark and the nav. It's a direct child of
  the header, so `space-between` does the work; the gaps were measured equal at 768, 1024, 1280
  and 1512, with the drawer open and closed.

---

## 2. Broken or missed by plan 07

These are regressions or gaps the plan's checks didn't cover. Items 1-3 come from the same
cause: below 64rem the document is now the scroller and a 52px sticky header covers its top
edge. Anything that sticks or scrolls to the top of the viewport now lands under the header.

1. **Work detail project list sticks under the header from 960 to 1023px.** `.work-detail__nav` is
   `position: sticky; top: 16px`. Between 60rem (where work detail gets its three columns) and
   64rem (where the pinned shell starts) it now sticks to the document, at 16px, behind the 52px
   header. Measured at 1000: list top 16, header bottom 52. Fine at 1280 (sticks at 68, inside
   `.shell__scroll`). Below 60rem it's one column and doesn't stick.
2. **Post rail sticks under the header from 736 to 1023px.** `.rail` is `position: sticky;
   top: var(--space-4)` (24px) above 46rem. Measured at 900: rail top 24, header bottom 52. Fine at
   1280 (76).
3. **Heading anchors and the skip link land under the header below 64rem.** Goldmark adds heading
   IDs (`WithAutoHeadingID`) for deep links. Measured at 375: `/blog/a-quick-example-post#why-this-matters`
   puts the heading at y=0, under the header (bottom at 52). `#main` from the skip link does the
   same. There's no `scroll-padding-top` anywhere.
4. **Duplicate IDs on `/profile`.** `Layout` always renders the drawer, so `/profile` contains
   `ProfileBody` twice: once in the page, once in the hidden `aside`. `id="profile-section-0"`
   appears twice (measured), and so does every section ID. That's invalid HTML, and
   `aria-labelledby` becomes ambiguous for the page's sections. It also ships the resume twice in
   that page's HTML.
5. **The meta description test is tied to the copy.** `TestProfilePage` checks the description
   starts with "Product designer". Rewording the summary in `profile.md` will fail a handler test.
   `ProfileSummary` picks the intro's longest paragraph; that's documented in `profile.go`, but it's
   a heuristic. If the title or location line ever outgrows the bio, the description silently
   changes.

Not regressions, but not checked: the green, dark-green and blueish-2 themes, and a real iOS
device (address bar collapse, Back scroll restore). Both were only reasoned about or emulated.

### Suggested fixes (for the polish plan)

- 1-3: one rule below 64rem, `html { scroll-padding-top: <header height>; }`, fixes anchors and the
  skip link. For the sticky elements, offset `top` by the header height below 64rem. A
  `--header-h` token (52px today, set from the header's `min-height`) would make both one value.
- 4: don't render the drawer on `/profile` (`Layout` skips the `aside` when `p.Nav == "profile"`),
  or give `ProfileBody` an ID prefix. Skipping the drawer is simpler and also removes the
  duplicate HTML.
- 5: assert that a description is present and comes from `ui.ProfileSummary()`, not on its text.

---

## 3. Left for the polish plan

Carried over from review 06 and plan 07's "out of scope", plus section 2 above. Ordered roughly
by user impact.

1. **Sticky offsets + `scroll-padding-top` below 64rem** (section 2, items 1-3).
2. **No drawer on `/profile`** (section 2, item 4).
3. **Focus and Escape (desktop):** focus the drawer's close control on open; return focus to the
   nav's Profile link on close; Escape closes. Measured today: focus falls to `<body>` after
   opening, and Escape does nothing.
4. **Focus rescue on resize:** if focus is inside the drawer when the viewport drops below 64rem,
   move it to the Profile link (`matchMedia` change listener; `terminal.js` already has
   `profileDesktop`). Measured today: focus falls to `<body>`.
5. **Work detail on container queries:** move `.work-detail`'s column changes to `@container`
   on `.shell__primary` so it reflows beside the drawer at 1024. This also removes the 60rem
   viewport breakpoint behind item 1 of section 2.
6. **Hit areas:** 44px-tall targets below 64rem for the wordmark and nav items (block padding
   plus negative margin, header height unchanged); 24px hit areas on the theme swatches with the
   visual size unchanged.
7. **Scramble only on hover devices:** gate `initScramble` on `(hover: hover)`; it now covers four
   header items, so taps scramble more often.
8. **Small-screen home intro** (review 06, option C): `IntroHTML` + links + "Full profile ->"
   after the hero, hidden at >= 64rem.
9. **Profile summary test** (section 2, item 5).
10. **Docs:** CLAUDE.md has no note of the 64rem shell breakpoint or the "below 64rem the
    document scrolls under a sticky header" rule. Add both, and the sticky-offset gotcha, so the
    next change doesn't repeat section 2.
11. **Verify** the remaining three themes and a real iOS device.

### Overlap with plan 08 (token review)

- The Work/Other gap question listed in plan 08 is settled (24px, `gap-24`).
- A `--header-h` token (item 1) and the 64rem / 46rem breakpoints belong in plan 08's breakpoint
  section.
- `gap-24` and the `mb-*` trap in the px-named spacing scale are the same issue plan 08 opens with.

### `docs/ia-decisions.md`

Matches what's built. When the polish plan lands, the home intro row (marked "later plan") will
need its marker removed. No other changes needed.
