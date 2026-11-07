# Site IA - agreed decisions

Source of truth for the site's information architecture, agreed with Matt in planning (7 Nov 2026).
Plans in `docs/plans/` build on this. If code and this file disagree, ask before changing either.

## Principles

- **Simple site.** Home, Work detail, Other, plus a Profile drawer. Nothing else public.
- **Figma is direction, not spec.** File: mattbx.co-final. Start from the mocks, improve on them in
  code wherever it helps; judge in the browser.
- **Reuse built interactions** (row hover fill, scramble, custom cursor) on every list and page -
  don't rebuild them per page.
- **House style:** short hyphens, never em dashes - in site copy, docs and commit messages.

## Pages

| Page | URL | Job |
|---|---|---|
| Home | `/` | Hero + the work list. Each row links to its project |
| Work detail | `/work/{slug}` | One project's info and images, with a left-hand nav listing all projects |
| Other | `/blog` today (URL undecided) | Writing and open-source work - **not yet defined, needs a rethink** |
| Profile | drawer, not a page | Bio + contact links; open by default on `/`, closed elsewhere |

## Work - agreed

- **One row = one project** (e.g. "Onboarding"), not one per client. Several projects for one
  client appear as separate rows - deliberate, it fills out the list.
- **Client** is its own field (e.g. "Tactiq.io"). It's the overview heading on work detail; the
  list row can combine client + project ("Tactiq Onboarding") so nothing is typed twice.
- **Images:** an ordered list per project. Any image may carry an optional small heading, which
  allows grouping later without a separate sections system. Layout is designed in the browser
  (the mock shows two small over one large, ~3:2).
- **Reference** (the "REF. V/GRQ-281" column): decorative, **generated from the slug** - no stored
  field. Free to play with symbols and formats; it doesn't need to be readable.
- **Format**: a real field - one short label per project. Confirmed labels: **Web app, Brand,
  Design system**. Full set to finalise (the mock also shows Engineering, Website, Marketing).
- **Work detail layout (from the mock):** left = "← Work" back link + the project list (same row
  as home); middle = client heading + description; right = project heading + images.
- Extras already in the data (summary, role, tech, live link, repo link) - which ones show is
  decided in the browser.
- Work rows (home and work-detail nav) use the same keyframed row hover as the posts list.

### Data changes this implies (projects)

- Add: `client`, `format`, and an ordered images relation (path, alt, optional heading, order)
- No field for reference - derive it from the slug
- Today the row's ref column shows `Role` and format shows the first `Tech` item - both to be
  replaced by the fields above
- Migrations stay additive (see CLAUDE.md - Disco runs old and new side by side)

## Still open

- **Other:** what it holds (writing, open source, experiments?), how it's organised, list vs
  detail, and its URL. Posts are one type today, told apart only by tags.
- **Format** label set.
- **Leftovers:** `/about` handler, `/portfolio/login` + logout routes, `PortfolioIndex` /
  `PortfolioProject` templates - keep or remove.
- Typography is a separate milestone (one locked type set) - not part of IA.
