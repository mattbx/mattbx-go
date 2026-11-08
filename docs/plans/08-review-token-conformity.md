# Plan 08 - REVIEW ONLY: token conformity + legacy clean-up

Branch: `feat/redesign`. **Review and proposal only - no code changes.** It can run while plan 07
is being built: note the commit you measured against, and if plan 07 changes land during your
review, re-check the files it touched before you finish.

Write the findings to `docs/reviews/08-token-conformity.md`, then stop so Matt can decide. A
migration plan follows from his choices.

## Why

The site has been through several builds (CRT terminal pass, the Cursor restyle, now Claude
Code) and the design tokens show it. Matt wants **one conformant, reusable set** - type,
spacing, colour, motion - that every page uses, with **standard Tailwind meanings**, and all
legacy from earlier builds removed.

Known already:
- **Spacing trap** (review 06): `internal/ui/tailwind/input.css` defines a px-named spacing
  scale (`--spacing-2: 0.125rem`, `--spacing-6: 0.375rem`, `--spacing-16: 1rem`...), so `gap-6`
  is **6px, not Tailwind's 24px**. Agents and humans will keep tripping on this
- **Two type systems:** Tailwind `--text-*` tokens (eight roles, only four distinct sizes;
  `text-project` and `text-title` unused in templates) plus the old CRT-era `--step-*` scale in
  `main.css` (~17 uses, mostly blog prose and older components)
- **Hardcoded values** in `main.css` (e.g. `12px`, `13px`, `0.6875rem`, theme swatch hex colours)
- `--tap: 44px` defined and never used
- Two colour naming layers (`--text-primary` / `--text-secondary` / ... per theme, plus
  `--color-page` / `--color-fg` / `--color-dim` / ... aliases)
- Uppercase in ~11 places plus `tracking-wide` / `tracking-wider` - check against the
  principle "no wide-tracked tiny uppercase labels"

## Read first

- `internal/ui/tailwind/input.css` - layers, imports (note: Tailwind `theme` + `utilities` only,
  **no Preflight**), `@theme static` tokens, per-theme colour blocks, `@source` paths
- `internal/ui/static/main.css` - unlayered, beats utilities; its own `:root` vars, the
  `--step-*` scale, component styles, `.prose`, breakpoints
- `internal/ui/classes.go` and the `*.templ` files - which utilities are actually used
- JS that reads CSS custom properties - **don't propose removing these without saying what
  replaces them:** `cursor.js` (`--cross-color`), `terminal.js` (`--grain-*`), `noise.js`
  (`--noise-*`, on a spike - may or may not land)
- `internal/ui/static/chroma.css` is generated (syntax highlighting) - note it, don't audit it
- `docs/ia-decisions.md` and `docs/reviews/06-nav-and-profile-breakpoints.md` (breakpoints:
  one shell breakpoint at 64rem; content breakpoints at 46rem)
- Tailwind v4 docs on `@theme`, the `--spacing` multiplier and namespace overrides - check the
  current docs (Context7 or the web), not memory
- Optional: if the Figma MCP is available, compare against the variables in file
  `mattbx.co-final` (key `JfhTRsfprDuqpocHijFLXG`). Figma is direction, not spec

## What to produce

1. **Inventory.** One table of every custom property / token: name, where defined, value (per
   theme for colours), number of uses and where, and a status - **keep / rename / merge /
   replace with Tailwind default / delete (legacy) / delete (unused)**.
2. **Hardcoded values.** Every literal size, colour, duration, easing, z-index, radius in
   `main.css`, templates and `classes.go` (file:line, value, the token it should use).
3. **Tailwind conformity.** Every place a token shadows a Tailwind default with a different
   meaning (the spacing scale first; check type, container, radius, breakpoint namespaces too),
   any arbitrary values, and a clear recommendation on what lives in utilities vs `main.css`.
4. **Proposed set,** kept small:
   - **Type:** the role list (fewer than eight), with size / line height / weight / tracking /
     case / font (Inter vs mono) per role, the fluid display role as built, and how post prose
     and the resume map onto it
   - **Spacing:** standard Tailwind scale; how the Figma-exact values (16px gutter, 24px column
     gap, 390px drawer...) are expressed - semantic tokens only where they earn it
   - **Colour:** one semantic naming layer across all five themes (page, fg, dim, muted, line,
     surface, navigation, danger...), contrast notes for each theme, and what the theme swatches
     should reference instead of hex
   - **Motion, breakpoints, radii, z-index:** as tokens if they're reused
5. **Legacy removal list:** what goes, why it's safe (no uses / replaced by), and any risk.
6. **Migration outline:** small, ordered steps, each checkable. Aim for **no visual change**
   except where you flag a deliberate tweak. Recommend a way to prove it (e.g. the review 06
   screenshot matrix before and after at 375 / 768 / 1024 / 1280 / 1512 in two themes). Flag
   anything that overlaps plan 07 or the M3b polish plan.
7. **Questions for Matt** - only real decisions (e.g. a size two roles would merge into, the
   Work/Other nav gap), each with your suggested answer.

## Constraints

- Go/templ, Tailwind v4 via the standalone CLI (no Node), CSS-first, no new libraries
- Keep Preflight off unless you argue for it with the cost spelled out
- Five themes via `html[data-theme]` must keep working, including first paint from the cookie
- Public repo - nothing private in the report
- House style: short hyphens, never em dashes, in the report and any commit messages

## Done when

`docs/reviews/08-token-conformity.md` exists with the inventory, hardcoded values, conformity
issues, a proposed set, the legacy list, a migration outline and questions. No code changed.
