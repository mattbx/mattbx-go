# Plan 02 - Gentler hero growth + drawer motion

Branch: `feat/redesign` (do not touch `main` - it deploys to production).
Follows plan 01 (hero sized in `cqi` against `.shell__primary`). No new libraries, no JS changes.

## Step 0 - commit plan 01 first (Matt approved it)

- Add `.playwright-mcp/` to `.gitignore` (local screenshots - public repo)
- Commit `internal/ui/static/main.css`, `internal/ui/tailwind/input.css` and
  `docs/plans/01-fluid-display-type.md` as one commit, e.g.
  "Size the hero against the primary column with container queries."
- Leave this plan file uncommitted until plan 02 is approved

## Part A - hero grows a third less above the Figma width

Today `--text-display: clamp(2rem, 9.31cqi, 100cqi)` is strictly proportional, so closing the
drawer at 1512 takes the hero from 101.5px to 137.8px. Matt wants that growth cut by a third.

Change the token in `internal/ui/tailwind/input.css`:

```css
--text-display: max(2rem, min(9.31cqi, calc(2.114rem + 6.21cqi)));
```

- Up to 1090px column content width (the Figma default view) `9.31cqi` is the smaller term, so
  everything at or below Figma size is unchanged - including mobile
- Above it, the shallower `2.114rem + 6.21cqi` line takes over (the two meet at 1090px)

Expected sizes:

| Viewport | Drawer | Before | After |
|---|---|---|---|
| 375 | - | 32px | 32px |
| 900 | - | 80.8px | 80.8px |
| 1440 | open | 94.8px | 94.8px |
| 1440 | closed | 131.1px | 121.2px |
| 1512 | open | 101.5px | 101.5px |
| 1512 | closed | 137.8px | 125.7px |
| 1920 | open | 139.5px | 126.8px |
| 1920 | closed | 175.8px | 151.0px |

## Part B - animate the drawer (CSS only)

The hero is sized off the primary column, so animating the column width animates the hero for
free. `terminal.js` already toggles `html.profile-open` - leave it alone.

All in `internal/ui/static/main.css`:

1. **One width token.** Add `--profile-width: 390px` (in `input.css` `@theme static` or at the top
   of the shell block - follow the file's existing pattern) and use it for the open value:
   `html.profile-open .shell { --profile-w: min(var(--profile-width), 100%); }`
2. **Make `--profile-w` animatable.** Register it:
   ```css
   @property --profile-w {
     syntax: "<length-percentage>";
     inherits: false;
     initial-value: 0px;
   }
   ```
   Then on `.shell`: `transition: --profile-w 400ms var(--ease-wipe);`
   (Fall back to transitioning `grid-template-columns` only if `@property` misbehaves.)
3. **Keep the drawer content from re-wrapping.** Give `.profile-panel__chrome` and
   `.profile-panel__body` a fixed inline size of
   `calc(var(--profile-width) - 2 * var(--spacing-gutter))` with `flex-shrink: 0` (the panel
   already has `overflow: hidden`), so the content slides into view instead of squashing.
4. **Don't hide the content until the close has finished.** `.profile-panel` flips
   `visibility` instantly today, which would blank it at the start of a close. Use the standard
   delay pattern:
   - `.profile-panel { transition: visibility 0s linear 400ms; }`
   - `html.profile-open .profile-panel { transition-delay: 0s; }`
   Keep `inert` / `aria-hidden` as they are (JS sets them immediately - that's fine).
5. **Reduced motion.** In the existing `@media (prefers-reduced-motion: reduce)` block, set
   `transition: none` on `.shell` and `.profile-panel`.
6. **Below 60rem** the drawer stacks instead of narrowing the column. Make sure the transition
   does nothing odd there; if it does, scope the transition to `(min-width: 60.001rem)`.

If 400ms / `--ease-wipe` feels wrong, say so in the report rather than tuning endlessly - Matt
will judge the feel.

## Check

- `go tool templ generate && go build ./... && go vet ./... && go test ./...`
- Measure the hero at every row of the table above (Playwright is fine)
- At 1512, toggle Profile a few times and record a short clip or before/mid/after screenshots:
  - the column and hero animate together over ~400ms, no jumps at start or end
  - drawer text slides in without re-wrapping; on close it stays visible until the end
  - **watch for steppy / jittery type** while it animates. If it's visibly steppy, report it -
    don't add GSAP or another library (a FLIP/transform approach is a separate decision)
- Reduced motion on (emulate it): toggle is instant
- Page load and back/forward: no animation on first paint
- Below 60rem: toggling still works, nothing animates strangely
- Stop after the checks, uncommitted, so Matt can look

## Out of scope

Grid/structure fluidity (M3), mobile drawer behaviour decisions, any new dependencies.
