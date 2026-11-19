# Review - plan 10, typography in one file

Plan: [`docs/plans/10-typography-in-one-file.md`](../plans/10-typography-in-one-file.md).
Local screenshots: `docs/reviews/10-screens/` (gitignored) - `before/`, `after/`, `specimen/`.

## Result

All type now lives in `internal/ui/tailwind/typography.css`. `main.css` has no type values, and
templates use one `type-*` recipe per element. `/admin/type` (admin only) lists every role with
live computed values and renders a full post through the real post markup.

No visual change, checked two ways across 9 pages (home, work, blog, two posts, profile, admin,
admin post form, login gate) at 375 / 1024 / 1512 in the default and light themes (54 captures):

- **Pixels:** all 54 full-page screenshots identical to the before set (header clock masked).
- **Computed values:** every rendered element's font-family, size, weight, style, letter-spacing,
  line-height, text-transform, numeric variant, `::before`/`::after` values and box position
  identical before and after (0 differences).

The before/after runs used a scratch copy of the local database with one extra post covering every
prose element, since the existing posts are mostly one-liners.

## Tokens added in step 2

Values copied as-is. "Same as" marks an identical value, not a merge.

| Token | Value | Used by | Same as |
| --- | --- | --- | --- |
| `--text-document` (+ `--line-height` 1.65) | 1.0625rem | `body` | `--text-section`, h4, input, checkline, button sizes |
| `--text-display--font-weight` | 630 | `type-display` (was `.hero__display`) | |
| `--font-weight-nav-lock` | 700 | `.nav .is-locked::after` | wordmark, resume emphasis |
| `--leading-profile-links` | 1 | `.profile-panel__links` | nav / section / display line heights |
| `--font-weight-resume-heading` | 400 | `.profile-panel__heading` | |
| `--tracking-resume-heading` | normal | `.profile-panel__heading` | |
| `--font-weight-resume-emphasis` | 700 | `.profile-panel__section h3`, `strong` | wordmark |
| `--leading-ledger-row` | 20px | `.entry__link`, `type-ledger-index`, `type-ledger-cell` | `--text-meta--line-height` (1.25rem) |
| `--text-ledger-title` (+ lh 20px, ls -0.01em, fw 500) | 13px | `.entry__title`, `type-ledger-title` | `--text-project` / `--text-title` (0.8125rem) |
| `--text-rail` (+ lh 1.5) | 0.8125rem | `.rail` | was `--step--1`; 8 other 13px roles |
| `--text-rail-key` (+ ls 0.08em) | 0.6875rem | `.rail__key` | `--text-prose-th` |
| `--tracking-rail-key-inline` | 0 | `.rail__key` below 46rem | |
| `--text-prose-h1` (+ lh 1.15, ls -0.02em, fw 600) | 2rem | `.prose h1` | |
| `--text-prose-h2` (+ lh 1.25, ls -0.015em, fw 600) | 1.25rem | `.prose h2` | `--text-gate-title` |
| `--text-prose-h3` (+ lh 1.25, ls -0.015em, fw 600) | 1.125rem | `.prose h3` | companions = h2 |
| `--text-prose-h4` (+ lh 1.25, ls 0.06em, fw 600) | 1.0625rem | `.prose h4` | `--text-document` |
| `--text-prose-code` | 0.875em | `.prose code` | |
| `--text-prose-pre` (+ lh 1.55) | 0.8125rem | `.prose pre` | 13px group |
| `--text-prose-table` | 0.8125rem | `.prose table` | 13px group |
| `--text-prose-th` (+ ls 0.06em, fw 600) | 0.6875rem | `.prose th` | `--text-rail-key` |
| `--font-weight-prose-anchor` | 400 | `.prose :is(h2, h3):hover::before` | |
| `--text-field-label` (+ ls 0.06em) | 0.8125rem | `.field__label` | 13px group; ls = h4 / th |
| `--text-field-hint` (+ ls 0) | 0.8125rem | `.field__hint` | 13px group |
| `--text-input` | 1.0625rem | `.input`, `.textarea`, `.select` | `--text-document` |
| `--text-textarea` (+ lh 1.6) | 0.8125rem | `.textarea` | 13px group |
| `--text-checkline` | 1.0625rem | `.checkline` | `--text-document` |
| `--text-button` (+ lh 1, fw 550) | 1.0625rem | `.button` | `--text-document` |
| `--text-gate-mark` (+ ls 0.1em) | 0.8125rem | `.gate__mark` | 13px group |
| `--text-gate-title` (+ ls -0.02em, fw 600) | 1.25rem | `.gate__title` | `--text-prose-h2` |
| `--font-weight-rowlist-title` | 550 | `.rowlist__title` | `--text-button--font-weight` |
| `--text-rowlist-meta` | 0.8125rem | `.rowlist__meta` | 13px group |

From template utility clusters (now inside recipes):

| Token | Value | Recipe | Note |
| --- | --- | --- | --- |
| `--font-weight-wordmark` / `--tracking-wordmark` | 700 / 0em | `type-wordmark` | was `font-bold tracking-normal` |
| `--tracking-meta-field` | 0.025em | `type-meta-field` | was `tracking-wide` |
| `--tracking-eyebrow` | 0em | `type-eyebrow` | was `tracking-normal` |
| `--tracking-page-title` / `--font-weight-page-title` | -0.025em / 600 | `type-page-title` | was `tracking-tight font-semibold`; overrides `--text-section`'s -0.03em |
| `--tracking-badge` / `--leading-badge` | 0.05em / 1.5 | `type-badge` | was `tracking-wider leading-normal` |

Case tokens (`--case-*`, 18 of them) sit in `:root` because text-transform has no Tailwind
namespace. All are `uppercase` except `--case-ledger-title`, `--case-field-hint` and
`--case-rail-key-inline` (`none`).

`--step-*` is gone from `main.css`.

### Duplicate groups worth merging first

- **13px (0.8125rem):** rail, prose pre, prose table, field label, field hint, textarea, gate mark,
  rowlist meta, ledger title (as 13px), plus the unused `--text-project` / `--text-title`.
- **17px (1.0625rem):** document, section, prose h4, input, checkline, button.
- **12px (0.75rem):** nav, meta, micro, body - four names for one size.
- **11px (0.6875rem):** rail key, prose th.
- **20px line height:** `--text-meta--line-height` (1.25rem), `--leading-ledger-row`,
  `--text-ledger-title--line-height`, `--text-project--line-height`, `--text-title--line-height`.
- **Uppercase tracking:** 0.06em (h4, th, field label), 0.08em (rail key), 0.1em (gate mark),
  0.05em (badge), 0.025em (meta field), 0 (eyebrow, wordmark, nav).

## Recipes (step 4)

15 recipes, each only for a role a template uses: `type-display`, `type-wordmark`, `type-nav`,
`type-meta-field`, `type-meta`, `type-meta-caps`, `type-eyebrow`, `type-page-title`, `type-lede`,
`type-resume`, `type-badge`, `type-notice`, `type-ledger-index`, `type-ledger-title`,
`type-ledger-cell`. `type-lede` and `type-resume` are identical; kept separate as two roles.

A recipe sets only what its old cluster set. `type-meta`, `type-meta-caps` and `type-notice` set
no letter-spacing, because `text-meta` never did - they inherit, as before.

## Decisions

- **`.hero__display` lost its type declarations** - `type-display` on the same element now
  carries them (identical values), so keeping both would be two sources for one value.
- **`.entry__title` keeps its type in `main.css`** even though work rows also get
  `type-ledger-title`. Blog rows have no recipe and rely on `.entry__title`'s letter-spacing
  surviving `font: inherit` (the shorthand doesn't reset letter-spacing). Removing it would shift
  blog titles from -0.01em to normal.
- **`.site-header__clock .text-meta` became `.site-header__clock .type-meta-field`**, since the
  clock's inner element no longer carries `text-meta`.
- **No step 3 exceptions were needed.** Left out of scope as not in the plan's property list:
  `font-style: italic` on blockquotes and `font-variant-numeric: tabular-nums` on the ledger index.
- **Specimen page:** the post's article markup was extracted from `BlogPost` into `postArticle` so the
  specimen renders the real thing. The sample body is a constant rendered once (`sync.OnceValues`),
  not per request. `Page.Scripts` loads `typespec.js` on that page only. Rows are listed in
  `typeGroups` in `typespec.templ`, and the page warns if the CSS has a `type-*` recipe with no
  row. A bottom section lists every type token the browser resolves, with identical values noted.

## Checks

- `go tool templ generate && go build ./... && go vet ./... && go test ./...` - pass. New tests:
  specimen is admin only (anonymous and portfolio sessions redirect), renders roles and the full
  post sample, and `typespec.js` loads on no other page.
- Step 3 grep of `main.css` - only `var(--...)` or `inherit`.
- `/admin/type` - 303 to login when signed out; works signed in. Editing `--text-nav` to
  0.875rem and rebuilding changed the nav row from 12px to 14px (edit reverted).
- `ENV=production` build - strict CSP on all pages including `/admin/type`, no console errors.

## Prose spacing and other oddities (reported, not fixed)

- **Paragraph spacing (known):** `.prose p { margin-block: 0; }` beats `.prose > * + *`, so
  consecutive paragraphs run together while list items keep `--space-2` between them. Visible in
  `before/post-rich-375-default.png`.
- **Heading space:** h2-h4 take `--space-6` (4rem) above and `--space-3` below, while paragraphs
  and lists get 0 between them (same specificity cause as above) - so the only vertical rhythm in a
  post is the large gap above each heading.
- **Hierarchy:** the post title (`type-page-title`, 17px) is the same size as body text
  (`--text-document`, 17px) and smaller than prose h2 (20px) and h1 (32px).
- **Code block colours in dark themes:** some Chroma tokens (`fmt.`, punctuation, the string) are
  near-invisible on `--color-surface` in the default theme.
- **Blog rows:** titles carry -0.01em tracking from `.entry__title` while the rest of the row has
  none (see Decisions).
- **Dead CSS:** `.entry__links` and `.nav .is-locked` have no markup; `prose--mono` and
  `work-detail__links` are classes with no CSS.
- **Tokens nothing reads** (marked "unused" in `typography.css`; editing them changes nothing):
  `--text-project` and companions, `--text-title` and companion, `--text-section--letter-spacing`
  (titles were always on `tracking-tight`, now `--tracking-page-title`), and
  `--text-micro--line-height` (badges were always on `leading-normal`, now `--leading-badge`).
