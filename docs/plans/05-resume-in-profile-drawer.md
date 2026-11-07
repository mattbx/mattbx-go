# Plan 05 - Resume in the profile drawer

Branch: `feat/redesign`. Agreed with Matt 7 Nov. Figma: `profile sidebar` frame (direction, not spec).

## Goal

The profile drawer shows Matt's resume, split into clear sections like the mock: an intro, then
sections each with a thin rule, a small mono heading and the content. The content lives in one
markdown file that Matt edits like his vault note - no admin screens, no database.

## Content (already written - don't rewrite it)

`internal/ui/content/profile.md` - prepared in planning from Matt's resume draft (90% done;
an outcomes pass will land later, so the file will change).

Format the code relies on:
- Everything **before the first `## ` heading** is the intro: title line, location, summary,
  background line
- Each `## Heading` starts a section. Today: **Experience, Earlier career, Skills, Selected
  clients**. New `##` sections must just work, with no code change
- Inside sections: `### Company · domain`, a context paragraph, `**Role · years**`, bullets.
  Lines ending in `\` are hard line breaks (role line over its description)
- Plain markdown only - no HTML comments or raw HTML (the renderer omits raw HTML)
- House style: short hyphens, never em dashes. The renderer's Typographer turns `--` into dashes,
  so never type `--` in this file

## Steps

1. **Load and split once at startup.** In `internal/ui` (e.g. `profile.go`), `//go:embed
   content/profile.md`, split on `## ` headings into an intro + ordered sections
   (`Title`, rendered HTML), and render each part with the existing `internal/markdown.Render`
   (goldmark + bluemonday). Do it once (package init / `sync.Once`) - not per request. A broken or
   empty file should fail loudly in tests, not at request time.
2. **Drawer markup.** In `ProfilePanel` (`internal/ui/layout.templ`):
   - Intro: the rendered intro (title line, location, summary, background), then the links row
   - Links row: **Get in touch ↗ · LinkedIn ↗ · GitHub ↗** from Go constants - add `SiteGitHub =
     "https://github.com/mattbx"` in `page.go` next to `SiteContact` / `SiteLinkedIn`. Keep the
     current `LinkedIn` URL (Matt to confirm in review)
   - Then each section: `<section>` with a thin top rule, a mono heading (the `##` title) and the
     rendered body. Use proper heading levels and `aria-labelledby` so the drawer reads as an
     outline to screen readers
   - The drawer no longer shows `SiteBio`. Still fix the em dash in `SiteBio` (short hyphen) -
     it's used by the leftover `about.templ`
3. **Styling** (`main.css`, scoped under `.profile-panel` - do **not** reuse `.prose`, it runs on
   the old `--step-*` scale):
   - Follow the mock: generous space between sections, a 1px rule (`--color-line`) above each
     section, the heading in small mono **with normal tracking** (no wide-tracked tiny caps)
   - Body text stays the drawer's current mono body style; the title line, company names and
     role lines a step brighter (`--color-fg`) than body (`--color-dim`) for scanning
   - Bullets: a small square marker to match the work page bullets; tidy hanging indent
   - Use existing tokens only - no new sizes. The current hardcoded `12px` / `18px` in the
     profile styles can move to the matching tokens if it's a like-for-like swap; the full
     type lock-down is M3a
   - The panel's fixed inline size and drawer animation from plan 02 must keep working - content
     slides in, no re-wrapping
4. **Tests:**
   - `profile.md` parses into a non-empty intro plus sections titled Experience, Earlier career,
     Skills, Selected clients (in that order)
   - no `—` (em dash) and no `--` anywhere in `profile.md` - a house-style guard
   - the drawer renders the section headings and the GitHub link

## Check

- `go tool templ generate && go build ./... && go vet ./... && go test ./...`
- In the browser (home, drawer open):
  - Intro reads cleanly: title line, location, summary, background, then the three links
  - Each section has its rule, heading and content; Experience scans well - company, context,
    role/years, bullets
  - Drawer scrolls on its own; open/close animation still slides without re-wrapping
  - Check one light theme and one dark theme
  - Below 60rem (stacked drawer): readable, scrolls
- Screenshot the drawer top and the Experience section for Matt
- Stop uncommitted so Matt can look

## Out of scope

PDF resume, outcomes/numbers pass (content), type lock-down (M3a), editing via admin.
