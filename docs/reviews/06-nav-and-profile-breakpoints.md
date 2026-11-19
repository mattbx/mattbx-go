# Review 06 - site nav + profile drawer across breakpoints

Plan: `docs/plans/06-review-nav-and-profile-breakpoints.md`. Review only - no code changed.

Measured 8 Nov 2026 on `spike/noise-overlay` (the plan names `feat/redesign`; the tree had the
noise-overlay spike uncommitted, so home screenshots include it). Measured against the running
dev app on :8080 with Playwright, viewport height 900, dark default theme unless noted.
Screenshots: `docs/reviews/06-screens/{home,work,blog}-{375,768,1024,1280,1512}-{default,toggled}.png`
plus `work-{375,1280}-light.png`. "default" is the page as served; "toggled" is after one click on
the visible Profile control. The screenshots are untracked - commit or delete as you like.

---

## 1. Current state, measured

### How the shell behaves today

There are three layout modes, set by two unrelated breakpoints:

| Width | Shell | Profile open | Scrolling |
|---|---|---|---|
| < 48rem (768) | one column | **split screen:** primary on top (55%), profile pane underneath (`minmax(12rem, 45%)`) | primary pane scrolls as a whole, footer inside it; header scrolls away |
| 48rem - 60rem | one column | same split screen | header + footer pinned, main scrolls between them; profile pane scrolls separately |
| > 60rem (960) | grid `1fr | 390px` | side drawer, animated `--profile-w` (plan 02) | each column scrolls; header pinned |

`html` and `body` are `overflow: hidden` at every size, so the document itself never scrolls.

### Numbers

| Width | Primary column when open | Profile pane | Header (open -> closed) |
|---|---|---|---|
| 375 | 375 x 495 | 375 x 405, at y=495 | nav jumps x=288 -> 198 when Profile appears |
| 768 | 768 x 495 | 768 x 405, at y=495 | clock visible; same jump |
| 1024 | 634 wide (602 content) | 390 side | fits |
| 1280 | 890 (858 content) | 390 side | fits |
| 1512 | 1122 (1090 content) | 390 side | fits |

Hit areas: Work 29x16, Other 36x16, Profile 66x16, theme swatches 12x12 with 8px gaps. `--tap: 44px`
is defined in `main.css` and never used.

### Problems, in priority order

1. **The profile-below-footer bug is the 60rem stacked mode, not a stray rule.** Below 960px,
   opening the profile turns the shell into two stacked scroll panes. On phones the footer lives
   inside the top pane, so the resume shows up as a second pane under the footer
   (`work-375-toggled.png`). On home it's open on load, so first paint on a phone is a hero
   squeezed into 495px with half a resume underneath (`home-375-default.png`). The close control
   inside it still says "Profile ->", an arrow pointing right in a pane that's at the bottom.
2. **Home's "open by default" is decided on the server**, which can't know the viewport. Any fix
   for small screens has to be CSS (the open class must do nothing below the drawer breakpoint),
   not handler logic.
3. **Work detail breaks beside the drawer at 1024.** `.work-detail` goes to one column at a
   *viewport* width of 60rem, so at 1024 with the drawer open it keeps three columns in 602px. The
   media column is about 140px wide (`work-1024-toggled.png`). `.shell__primary` is already a
   container (`container-type: inline-size`), but this grid doesn't use container queries.
4. **Focus is dropped on every toggle.** Opening from the header button hides the button
   (`display: none`), and closing from the drawer makes the drawer `inert`. Both times focus falls
   to `<body>` (measured). Keyboard and screen reader users lose their place.
5. **No Escape.** Escape with the drawer open does nothing (measured).
6. **Tab order puts the drawer last.** On home at 1280 (open): skip link, wordmark, Work, Other,
   hero link, work rows, the five theme swatches, *then* Profile and the resume. On screen the
   drawer is top right, beside the nav.
7. **The document never scrolls.** Because `html`/`body` are `overflow: hidden`, mobile Safari
   doesn't collapse the address bar, tapping the status bar doesn't scroll to top, and the
   browser can't restore scroll position on Back. Below 48rem this all happens inside
   `.shell__primary`.
8. **Hit areas are small.** 16px-tall nav links pass WCAG 2.2 target size (2.5.8) only through the
   spacing exception. The theme swatches (12px, 8px apart) fail it. Below 44px is uncomfortable
   on touch either way.
9. **`gap-6` on the nav is 6px, not 24px.** `input.css` defines a px-named spacing scale
   (`--spacing-6: 0.375rem`), so `gap-6` gives 6px. "WORK OTHER" reads as one phrase and the two
   hit areas almost touch. Might be intended from Figma, but `site-header__end` uses 24px between
   the groups, so the spacing is uneven. Worth confirming.
10. **Header shifts when the profile toggles.** The launch button is removed (`display: none`),
    so the nav slides 90px at 375. On desktop that's expected (the column narrows); on mobile
    it's just a jump.
11. **Profile isn't in the nav landmark.** `<nav aria-label="Primary">` holds Work and Other; the
    Profile button sits outside it. A screen reader listing the nav hears two items.
12. **Two "Profile" buttons control one panel** (launch and close, both `aria-controls`,
    both `aria-expanded`). That works, but only one is ever visible, so it reads fine. Keep it in
    mind if the control changes.
13. **Scramble runs on tap.** `initScramble` checks reduced motion but not `(hover: hover)`, so on
    touch the label scrambles on tap, right before the page navigates. Magnetic is gated
    properly. Harmless but noisy.
14. **Doesn't work without JS.** The Profile control is a `<button>`; without JS it does nothing.
    `?profile=1` works, but nothing links to it.
15. **Small leftovers.** The `server.go` comment says `/about -> /?profile=1`, but `handleAbout`
    redirects to `/` with no parameter. `about.templ` has stale copy containing em dashes. When
    closed below 60rem, the hidden drawer body collapses to zero width and lays out at 14,600px
    tall. Invisible, but it costs layout work.

### What's fine

- Header fits at every measured width; no wrapping, even at 375 with the profile control showing.
  At 320 it still fits: wordmark 65 + Work 29 + Other 36 + Profile 66 + gaps is about 230px.
- Active underline follows the IA: Work underlined on `/work/ledger`, Other on `/blog`, nothing on
  home, wordmark never underlined.
- Closed drawer is `aria-hidden` + `inert`; open drawer is a labelled `<aside>`; both buttons use
  `aria-expanded`. The approach is right; only focus handling is missing.
- Global `:focus-visible` (2px `--color-fg`) reads in both dark and light (`work-375-light.png`).
- Reduced motion turns off the drawer transition.
- Clock is hidden below 48rem and fits at 768+.

---

## 2. Profile at each size - options

**Recommended drawer breakpoint: 64rem (1024).** At 1024 the open drawer leaves 602px of content.
That's enough for the hero (it scales with `cqi`) and the home work list. Work detail fits *once
it uses container queries* (see build step 6). That covers iPad landscape and every laptop, and
sends iPad portrait (768-834) and phones to the small-screen behaviour. If container queries for
work detail are out of scope, use **80rem (1280)** instead. At 858px of content the three columns
fit as they are, but a 1024-1279 laptop then loses the drawer.

### A. Drawer on desktop only; Profile is a page below the breakpoint (Matt's leaning)

- **>= 64rem:** unchanged drawer, motion and per-route defaults.
- **< 64rem:** the drawer `aside` is `display: none` and the grid is a single column whatever the
  `profile-open` class says. Profile is a normal nav item linking to **`/profile`**. On that page
  the item is underlined, the same as Work and Other.
- **URL / nav:** new public route `/profile`, registered bare. It renders the same resume as a page
  body at *every* width. It has to: the server can't know the width, so `/profile` on a desktop
  must still be a real page (shareable, indexable, works without JS). `/about` 301s to `/profile`.
- **Home "open by default":** stays server-side, but the CSS that applies it only exists at
  `min-width: 64rem`, so phones see a normal home page.
- **Resize across the breakpoint:** all CSS. Going wide with `profile-open` shows the drawer; going
  narrow hides it. If focus was inside the drawer when it gets hidden, move it to the Profile link
  (a `matchMedia` change listener, a few lines).
- **One control, not two (suggested):** make the header Profile control an
  `<a href="/profile" data-profile-toggle>`. At >= 64rem JS calls `preventDefault()` and toggles
  the drawer (with `aria-expanded`, which ARIA allows on links). Below 64rem, or without JS, it
  navigates. That's one element, no duplicated markup, and it fixes no-JS everywhere. It can move
  inside `<nav>`, which fixes issue 11.
- **Accessibility:** below 64rem nothing to manage, it's a page. On desktop add focus moving into
  the drawer on open and back to the trigger on close, plus Escape. Back works naturally on
  mobile. On desktop the drawer stays outside history, as today.
- **Effort:** small to medium. One handler, one template (shares a `ProfileBody` component with the
  drawer), CSS that mostly deletes the 60rem stacked mode, about 20 lines of JS.
- **Risks:** `/profile` and the drawer both carry the resume. That's fine, it's one component and
  one markdown file, and the drawer markup is already on every page. The site gains a public
  page, which changes the IA (flagged below).

### B. Drawer on desktop; full-screen sheet on small screens (same toggle)

- **< 64rem:** the same button opens the `aside` as a full-viewport sheet over the page. It needs a
  focus trap, background scroll lock, Escape, a close control at the top, and (to make Back
  close it) a `pushState`/`popstate` pair. Native `<dialog>.showModal()` would give the trap and
  Escape for free, but the same element is an in-grid drawer on desktop, so you'd need two
  elements or a role switch on resize.
- **Home "open by default":** must be suppressed below 64rem, or phones get a sheet over the home
  page on load. Same CSS gating as A.
- **URL:** none. Can't be linked to or shared on mobile, no no-JS path, not indexable.
- **Resize:** an open sheet crossing to desktop has to turn into a drawer; trap and scroll lock
  have to be torn down. This is where bugs will live.
- **Effort:** medium to high, mostly JS. That works against "minimal first-party JS".
- **Risks:** iOS scroll-lock quirks, focus-trap edge cases, Back-button behaviour. More JS to test
  against the production CSP.

### C. A, plus a short intro on the small-screen home page

Same as A, but home below 64rem also renders the profile *intro* (title, location, the three
links) as a compact block after the hero, ending in "Full profile ->" linking to `/profile`.
That keeps the intent of "profile open on home" for phone visitors (they see who you are and how
to reach you) without the 3,000px resume. It's the Cargo pattern, but with a taste of the content
in place of a bare "Information" link.

- **Effort:** A + one small block. `IntroHTML` already exists separately in `profileContent`, so
  it's still one source.
- **Risk:** the intro appears twice on home at desktop sizes unless the block is hidden at >= 64rem
  (one CSS rule).

### Rejected: no drawer at all

Profile as a page at every size is the simplest option, but it throws away plan 02's drawer and
home's side-by-side layout, which is the most distinctive thing on desktop. Not recommended
unless the drawer becomes a maintenance burden.

---

## 3. The nav itself - options

Three items plus wordmark and clock. A menu button is not needed: at 320 the inline row fits
with room to spare (about 230px of about 288px).

### N1. Inline links at every size (recommended)

- **Small (< 48rem):** `mattbx.co` left, `WORK  OTHER  PROFILE` right. No clock. Each item gets a
  44px-tall hit area through block padding, with matching negative margin so the 52px header
  doesn't grow.
- **Medium (48 - 64rem):** add the clock, as today.
- **Large (>= 64rem):** as today: clock, Work, Other, Profile control with arrow; Profile moves
  into the drawer chrome when open.
- **Active state:** underline = current section, per IA. With option A, Profile is underlined on
  `/profile`, and never for the open drawer (the drawer isn't a section; its open state is shown
  by the drawer itself and `aria-expanded`).
- **Pinned vs scrolling:** below 64rem let the *document* scroll (lift `overflow: hidden` from
  `html`/`body`) and make the header `position: sticky`. 52px is cheap, it keeps navigation one
  tap away on long work pages, and document scrolling fixes the mobile Safari issues in 1.7. At
  >= 64rem keep the pinned two-column shell.
- **Scramble:** keep on hover and keyboard focus; gate to `(hover: hover)` so taps don't scramble.
- **Focus:** current global outline is fine in both themes. Check the outline doesn't clip under
  `overflow: hidden` on the header once padding grows.
- **Spacing:** decide whether the Work/Other gap is 6px (current, from the px-named scale) or
  matches the 24px group gap. With 44px hit areas, 6px visual spacing still gives separate
  targets.

### N2. Menu button on small screens

A "Menu" control opening a full-width panel with Work / Other / Profile, clock and contact links.
It looks familiar, but adds a tap and JS (or a `popover`) to reach three links that already fit.
Only worth it if the nav grows past about five items. Not recommended.

### N3. Bottom bar on small screens

Fixed bottom row with the three items, within thumb reach. Unusual for a portfolio; it competes
with the footer and the theme switcher and needs `safe-area-inset-bottom` handling. Not
recommended.

### Themes

Nav colour is `--color-navigation`; the underline is `currentColor`; focus is `--color-fg`. All of
them follow the theme. Light reads well (`work-375-light.png`). The green and blue themes weren't
screenshotted in this pass; check them when the build lands.

---

## 4. Recommendation

**Option C (A plus the small-screen home intro) with nav N1, drawer from 64rem.**

- It fixes the reported bug by removing the mode that causes it (the 60rem stacked split), not by
  patching around it.
- Phones and tablets get a normal page that scrolls, can be linked and works without JS. Desktop
  keeps the drawer, its motion and the per-route defaults.
- One breakpoint (64rem, which is Tailwind `lg`) separates "desktop shell" from "stacked
  document". It replaces today's 47.999rem / 48rem / 60rem split. Content breakpoints (46rem for
  ledger and leaf) stay as they are.
- It needs the least JS of the options: one click handler with a media check, focus return,
  Escape.

### Build outline (small, ordered, each checkable)

1. **Extract `ProfileBody`** from `ProfilePanel` (intro, links, sections). Drawer renders it. Check:
   `go test ./...` passes, drawer looks identical at 1280.
2. **Add `GET /profile`** (bare in `router.go`), `ui.ProfilePage` using `ProfileBody`, `Nav:
   "profile"`, `ProfileOpen: false`. Check: handler test for 200, a resume section heading,
   `aria-current` on the Profile item.
3. **Point `/about` at `/profile`** (301) and delete `about.templ`. Fix the `server.go` comment.
   Check: test that `/about` redirects to `/profile`.
4. **One Profile control:** turn the header launch into `<a href="/profile" data-profile-toggle>`
   inside `<nav>`. In `terminal.js`, intercept the click only when
   `matchMedia("(min-width: 64rem)")` matches. Check: below 1024 it navigates; at 1024+ it toggles;
   with JS off it navigates at every width.
5. **Gate the drawer to >= 64rem in CSS:** move `html.profile-open .shell` and the panel's
   visible state into `@media (min-width: 64rem)`; `display: none` on `.profile-panel` below it;
   delete the 60rem stacked block. Check: home at 375 and 768 has no profile pane; at 1024+
   unchanged; no layout shift on load at any width.
6. **Work detail on container queries:** move `.work-detail`'s column changes from `@media` to
   `@container` on `.shell__primary`, with thresholds picked in the browser (about 50rem content
   for three columns). Check: `/work/ledger` at 1024 with the drawer open has a usable media column.
7. **Document scrolling below 64rem:** change the 47.999rem block to below 64rem, let
   `html`/`body` scroll there, make the header sticky, and keep the footer in flow. Check: on
   iOS Safari the address bar collapses on scroll; Back restores scroll position.
8. **Focus and Escape (desktop):** on open, focus the drawer's close control; on close, return
   focus to the header control; Escape closes. On a `matchMedia` change, if focus is inside a
   drawer that just got hidden, move it to the Profile link. Check: keyboard-only walk-through at
   1280; `document.activeElement` is never `<body>` after a toggle.
9. **Hit areas and spacing:** 44px-tall nav hit areas below 64rem; 24px hit areas on the theme
   swatches (padding or a pseudo-element, visual size unchanged); settle the `gap-6` question.
   Check: measure again at 375.
10. **Scramble on hover devices only:** gate `initScramble` on `(hover: hover)`. Check: tapping on
    a touch emulator doesn't scramble.
11. **Small-screen home intro (option C):** render `IntroHTML` + links + "Full profile ->" after the
    hero, hidden at >= 64rem. Check: 375 home shows it; 1280 home doesn't show it twice.
12. **Docs and verification:** update `docs/ia-decisions.md` and the CLAUDE.md breakpoint notes;
    repeat this review's screenshot matrix; check the result in a build running with
    `ENV=production` (CSP).

### Changes to `docs/ia-decisions.md`

- **Principles:** "Home, Work detail, Other, plus a Profile drawer. Nothing else public" becomes
  "... plus Profile (a drawer on desktop, a page at `/profile`)".
- **Pages table, Profile row:** "drawer, not a page" becomes "drawer from 64rem; `/profile` page at
  every size; drawer open by default on `/` at desktop sizes only".
- **Navigation:** add "Profile -> `/profile`; underlined on `/profile`. On desktop it opens the
  drawer, and the open drawer never shows an underline."
- **Still open -> Leftovers:** the `/about` handler is resolved (it redirects to `/profile`;
  template removed).
