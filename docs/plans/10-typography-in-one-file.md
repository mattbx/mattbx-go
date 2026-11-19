# Plan 10 - All typography in one file (`typography.css`), no visual change

Branch: `feat/redesign`. Agreed with Matt 8 Nov. Brought forward because Matt wants to rework the
type scale **by hand** - this plan only gathers today's type into one tidy, editable place. **Do
not redesign anything:** keep current role names and values, and make no visual change.

Plan 08 (token review) is on hold until this lands. Plan 09 may run before or after - if it has
landed, include its files in the before/after screenshots.

## The intent

- `internal/ui/tailwind/typography.css` is **the only place type is defined**: font families,
  every size, line height, letter spacing, weight and case
- It's imported from `input.css` the same way `fonts.css` is (`fonts.css` stays `@font-face` only)
- **`main.css` never contains a type value again** - layout and components only, referencing
  typography tokens via `var()`
- Templates name a **role** (`type-nav`), not a recipe (`font-mono text-nav uppercase
  tracking-normal`)

## Step 0 - before screenshots

Before changing anything, capture the screenshot matrix: `/`, `/work/{slug}`, `/blog`, a post,
`/profile` (if plan 09 hasn't removed the drawer there, fine), the admin dashboard - at 375, 1024
and 1512, in the default dark theme and light. Keep them for the after comparison (gitignored
location, e.g. `docs/reviews/10-screens/before/`).

## Steps

1. **Create `typography.css`** with a short header comment: what lives here, the rule that
   `main.css` holds no type values, and how to add a role. Import it in `input.css` right after
   `fonts.css`. **Move** (don't copy) the font families (`--font-mono`, `--font-ui`,
   `--font-prose`) and every `--text-*` token with its companions out of `input.css` into it,
   unchanged. Keep `@theme static` so `main.css` can `var()` everything.
2. **Bring every other type value in, as named tokens with identical values:**
   - the old CRT-era `--step-*` scale in `main.css` (prose headings, rails, etc.)
   - hardcoded sizes (`12px`, `13px`, `0.6875rem`...), weights (e.g. the hero's `630`), letter
     spacings, line heights in `main.css`
   - use Tailwind v4 namespaces so utilities stay standard: `--text-*` (+ `--line-height`,
     `--letter-spacing`, `--font-weight` companions), `--font-weight-*`, `--tracking-*`,
     `--leading-*` - check the current Tailwind v4 docs for exact names, not memory
   - name tokens by role where it's obvious (`--text-prose-h2`), otherwise by where they're used,
     and mark duplicates with a comment (e.g. `/* same as --text-nav */`). **Don't merge or
     rename existing roles** - that's Matt's job afterwards
   - delete `--step-*` from `main.css` once nothing uses it
3. **Strip type literals from `main.css`.** Every `font-size`, `font-weight`, `letter-spacing`,
   `line-height`, `font-family` and type-related `text-transform` uses a typography token
   (inherit / unitless resets like `line-height: 1` on a reset can stay if they're not design
   values - list any you leave and why).
   *Check:* a grep of `main.css` for those properties finds only `var(--...)`, `inherit`, or the
   listed exceptions.
4. **Role recipes.** For each type role the templates use, add a Tailwind v4 `@utility type-<role>`
   in `typography.css` that bundles family, size (with its companions), weight, tracking and case.
   Replace the purely typographic part of the clusters in templates and `classes.go` with the
   recipe class (leave colour, margin and layout utilities as they are). Update the comment at
   the top of `classes.go` to describe the convention. Only add recipes for roles actually used.
5. **Type specimen page (last - Matt may stop before this step).** `GET /admin/type`, wrapped in
   `requireAdmin` in `router.go`, `noindex`. Lists every role in `typography.css`: its name, the
   class/token to use, short and long sample text, and the **live computed values** (family,
   size, line height, tracking, weight, case) read in the browser so they're always true. Uses
   the normal layout so the theme switcher works on it. Any script is first-party and loaded
   only on that page (strict CSP in production). Purpose: Matt edits `typography.css`, refreshes,
   and sees the whole scale at once.
   **Include a full post sample in context**, rendered through the real post markup and `.prose`
   styles (not a copy): the post header (published date + read time, title, summary), a lead
   paragraph, h2 / h3 / h4, several paragraphs in a row, a paragraph followed by a list, an
   unordered and an ordered list, inline code, a fenced code block, a blockquote and a link. Blog
   type is where the hierarchy is most broken today (title smaller than body and h2), so Matt
   needs to tune it in context, not role by role.
6. **CLAUDE.md note** (short): type lives only in `internal/ui/tailwind/typography.css`;
   `main.css` holds no type values; templates use `type-*` recipes; `/admin/type` shows the scale.

## Check

- `go tool templ generate && go build ./... && go vet ./... && go test ./...`
- **After screenshots** of the same matrix - compare with step 0. Expect **no visual change**;
  if anything differs, explain each difference (or fix it)
- The grep from step 3
- `/admin/type` (if built) - 404/redirect when logged out, works logged in, values update after
  editing `typography.css` and rebuilding
- Production build (`ENV=production`) - no CSP errors
- Report: the list of tokens added in step 2 with where each was used, and the duplicates found -
  that list is Matt's starting point for the manual pass
- Stop uncommitted so Matt can look

## Known issue to report, not fix

Prose spacing bug (seen by Matt 8 Nov on a post at phone width): in `main.css`,
`.prose p { margin-block: 0; }` is more specific than `.prose > * + * { margin-top: ... }`, so a
paragraph that follows another block loses its top margin ("...under it." runs straight into the
next paragraph) while list items keep generous spacing. Changing it is a visual change, so
**leave it** - list it in the report with any other prose spacing oddities you notice. Matt will
fix it in his hands-on pass (or ask for a follow-up).

## Out of scope

Choosing the final roles and values (Matt, by hand), spacing / colour / motion and the px-named
spacing scale (plan 08), any visual change.

House style: short hyphens, never em dashes - in code comments, docs and commit messages.
