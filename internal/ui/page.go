// Package ui holds the templ components that render the site.
package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mattbx/mattbx-go/internal/db"
)

// Edit these in one place to rebrand the site.
const (
	SiteName    = "Matt Barron"
	SiteMark    = "mattbx"
	SiteRole    = "Software engineer"
	SiteTagline = "Writing about the software I build, and the work I've shipped."
)

// Page carries the per-request context the shell needs. Handlers build one and
// pass it into every top-level component.
type Page struct {
	Title string // page title; SiteName is appended by the shell

	// Description populates the meta description and og:description.
	Description string

	// Nav marks the active navigation item: "blog", "portfolio", or "".
	Nav string

	// IsAdmin controls whether the admin toolbar renders. It is presentation
	// only — every protected route is gated by middleware, never by this flag.
	IsAdmin bool

	// BaseURL is the site's public origin, used for absolute URLs.
	BaseURL string

	// Path is the current request path, used to build the canonical URL.
	Path string
}

func (p Page) DocumentTitle() string {
	if p.Title == "" {
		return SiteName
	}
	return p.Title + " · " + SiteName
}

func (p Page) CanonicalURL() string {
	return p.BaseURL + p.Path
}

func (p Page) Year() string { return time.Now().Format("2006") }

// railDate is the long form shown in the margin ("2 Sep 2026").
func railDate(t time.Time) string { return t.Format("2 Jan 2006") }

// machineDate is the value for <time datetime="…">.
func machineDate(t time.Time) string { return t.Format("2006-01-02") }

// readingTime estimates minutes from the Markdown source at 220 wpm, floored
// at one minute so nothing ever reads "0 min".
func readingTime(source string) string {
	words := len(strings.Fields(source))
	minutes := max(words/220, 1)
	return fmt.Sprintf("%d min", minutes)
}

// postCount renders a human count for the index eyebrow.
func postCount(n int, singular, plural string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", singular)
	}
	return fmt.Sprintf("%d %s", n, plural)
}

// statusOf describes whether an item is live or still a draft. Only ever shown
// to admins, since drafts are filtered out of public queries entirely.
func statusOf(published bool) string {
	if published {
		return "live"
	}
	return "draft"
}

// hasLinks reports whether a project has anything to link out to.
func hasLinks(p *db.Project) bool { return p.LinkURL != "" || p.RepoURL != "" }

// editPostPath and editProjectPath keep admin deep-links in one place.
func editPostPath(id int64) string {
	return fmt.Sprintf("/admin/posts/%d/edit", id)
}

func editProjectPath(id int64) string {
	return fmt.Sprintf("/admin/projects/%d/edit", id)
}

// errorCode formats an HTTP status for the error page's eyebrow.
func errorCode(status int) string { return fmt.Sprintf("Error %d", status) }

func statusClass(published bool) string {
	if published {
		return "badge--live"
	}
	return "badge--draft"
}

func orderLabel(n int) string { return strconv.Itoa(n) }

func deletePostPath(id int64) string    { return fmt.Sprintf("/admin/posts/%d/delete", id) }
func deleteProjectPath(id int64) string { return fmt.Sprintf("/admin/projects/%d/delete", id) }

// formEyebrow and formTitle keep the new/edit wording consistent.
func formEyebrow(isNew bool, noun string) string {
	if isNew {
		return "New " + noun
	}
	return "Editing " + noun
}

func formTitle(isNew bool, current, fallback string) string {
	if isNew || current == "" {
		return fallback
	}
	return current
}

// DisplayTitle is what's shown wherever a post needs a title-shaped label.
// Micropub notes are posted without one by convention (that's what makes them
// notes rather than articles), so this falls back to a truncated snippet of
// the body rather than showing an empty heading.
func DisplayTitle(post *db.Post) string {
	if post.Title != "" {
		return post.Title
	}
	return snippet(post.BodyMD, 60)
}

// snippet collapses whitespace/newlines and truncates on a word boundary.
func snippet(source string, max int) string {
	fields := strings.Fields(source)
	joined := strings.Join(fields, " ")
	if len(joined) <= max {
		return joined
	}
	cut := joined[:max]
	if i := strings.LastIndex(cut, " "); i > 0 {
		cut = cut[:i]
	}
	return cut + "…"
}
