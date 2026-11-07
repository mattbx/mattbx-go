# Plan 03 - Bring back the row hover fill on work rows

Branch: `feat/redesign`. Run **after plan 02** is reviewed and committed (both touch `main.css`).

## Goal

The keyframed row hover (each cell wipes in with a fill, text flips colour, wipes out on leave)
works on the posts list but was lost on the work rows in the restyle. Bring it back on:
- the home work list (`/`)
- the project nav on work detail (`/work/{slug}`) - same `workEntry` component, `compact` mode

Reuse the existing system as is - no new animation code.

## Why it broke

The CSS (`.hover-fill`, `.hover-fill-group` in `internal/ui/static/main.css`) and the arming
script (`initHoverFillArm` in `terminal.js`) are generic. `postEntry` in `blog.templ` still has
the hooks; `workEntry` in `home.templ` was rewritten without them.

## Steps

1. In `workEntry` (`internal/ui/home.templ`), mirror `postEntry`:
   - add `hover-fill-group` to the `<a class="entry__link">`
   - wrap the contents of each cell (index, title, ref, format) in `<span class="hover-fill">`
   - keep the draft badge as it is in `postEntry` (inside a filled cell is fine)
2. **Keep dim cells dim at rest.** `.hover-fill` sets `color: var(--color-fg)` and the
   `hover-fill-text-out` keyframe returns to `--color-fg`, which would brighten the ref and
   format cells. Add a resting-colour variable (e.g. `--hover-fill-rest`, default
   `var(--color-fg)`) used by the base rule and the text-out keyframe, and set it to the dim
   colour for those cells. Keep it to a small, generic change - posts rows must look identical
   to today.
3. Check the hover fill reads well against each theme (default, light, green, dark-green,
   blueish-2) - the fill uses `--color-fg` and the text flips to `--color-page`.
4. Optional, only if it falls out naturally: on work detail, the current project's row in the
   nav could show a quiet "current" state. Don't build anything new for it - note it for Matt.

## Check

- `go tool templ generate && go build ./... && go vet ./... && go test ./...`
- Home: hover a row - every cell wipes in together; leave - it wipes out; no flash on page load
- Fast sweeps up and down the list still overlap smoothly (that's the 0.5s duration's job)
- Keyboard: tab through rows - focus shows the fill
- Work detail nav: same behaviour in the narrow column, nothing clipped or wrapping oddly
- Ref and format cells stay dim at rest, flip on hover
- Posts list unchanged
- Reduced motion: no animation, hover still shows a clear state
- Stop uncommitted so Matt can look
