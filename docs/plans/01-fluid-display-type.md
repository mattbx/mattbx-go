# Plan 01 - Fluid display type (container queries)

Branch: `feat/redesign` (do not touch `main` - it deploys to production).
Brief: Claude project "Brand stuff" > `website-brief.md`. Reference site: https://8886592.cargo.site/

## Goal

The home hero (`h1.hero__display`, `text-display`) scales with the **primary column**, not the
viewport. Opening the profile drawer narrows the column, so the hero shrinks with it; closing grows
it. No upper cap - it keeps scaling on big screens (Cargo-style). Body/UI type stays fixed `rem`.
No JavaScript.

## Why

Today `--text-display: clamp(2.5rem, 6.7vw, 6.346rem)` in `internal/ui/tailwind/input.css`:
- tracks the viewport, so the drawer opening/closing doesn't change it
- stops growing at ~1515px wide

## Reference numbers (Figma `mattbx.co-final`, frame "home + profile sidebar open - default view")

- Primary column 1122px, minus 16px gutters each side = **1090px content width**
- Hero size 6.346rem = **101.5px** at that width -> **101.5 / 1090 = 9.31cqi**
- Widest hero line ("Work & Other Things") is ~83% of the column at that size, so pure `cqi`
  always fits - no overflow at any width

## Steps (small, in order)

1. **Make the primary column a size container.** Add `container-type: inline-size` to
   `.shell__primary` in `internal/ui/static/main.css` (or Tailwind's `@container` utility on the
   element in `layout.templ` - pick one, not both). Note `cqi` resolves against the content box
   (inside the gutter padding), which is what the 1090px figure assumes.
2. **Retoken the display size** in `input.css` `@theme static`:
   - `--text-display: clamp(2rem, 9.31cqi, 100cqi);` (2rem floor only guards very narrow phones;
     the 100cqi max is effectively "no cap" but keeps clamp valid)
   - Remove `--text-hero` and its `--line-height` / `--letter-spacing` siblings - unused anywhere
     (grep `text-hero` to confirm first)
3. **Make the line overlap scale too.** `.hero__line + .hero__line { margin-top: var(--hero-overlap, -12px) }`
   is fixed px - change the fallback to `-0.12em` (12 / 101.5) so line spacing stays proportional.
4. **Leave everything else alone** - body, nav, meta, project tokens stay fixed `rem`.

## Check

- `go tool templ generate && go build ./... && go vet ./... && go test ./...`
- Run `./scripts/local.sh`, open http://localhost:8383 and check the hero at:
  - 1512 wide, profile open -> ~101px (should match Figma default view)
  - 1512 wide, profile closed (`?profile=0`) -> larger (~139px)
  - 900 and 1440 wide, open and closed - scales smoothly, no jumps
  - 375 wide (mobile) - three lines fit, no horizontal scroll
  - 1920+ wide - keeps growing; note how it feels (a max may come later)
- Toggle the Profile control on `/` and watch the hero resize live
- Screenshot open vs closed at 1512 for Matt

## Out of scope

Grid/structure fluidity, mobile body-size bump, work detail layout - separate plans.
