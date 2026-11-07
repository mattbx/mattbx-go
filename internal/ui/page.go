// Package ui holds the templ components that render the site.
package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mattbx/mattbx-go/internal/db"
)

// Site copy — edit here to rebrand.
const (
	SiteName     = "Matt Barron"
	SiteMark     = "mattbx.co"
	SiteRole     = "Work & Other Things"
	SiteTagline  = "Work & Other Things"
	SiteHeroCode = "#224013"
	SiteContact  = "mailto:hello@mattbx.co"
	SiteLinkedIn = "https://www.linkedin.com/in/mattbarron"
	SiteGitHub   = "https://github.com/mattbx"
	SiteBio      = "I design and build digital products and things - from interfaces to the systems behind them. Based in Sydney AU. Available for work opportunities."
)

// ValidThemes are allowed data-theme cookie values.
var ValidThemes = map[string]bool{
	"default":    true,
	"light":      true,
	"green":      true,
	"dark-green": true,
	"blueish-2":  true,
}

// Page is per-request shell context from handlers.
type Page struct {
	Title       string // SiteName appended by the shell when set
	Description string
	Nav         string // "work" | "other" | "blog" | ""
	// IsAdmin is presentation only (toolbar). Access is middleware-gated.
	IsAdmin     bool
	BaseURL     string
	Path        string
	Theme       string // data-theme on <html>
	ProfileOpen bool
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

func (p Page) ThemeAttr() string {
	if ValidThemes[p.Theme] {
		return p.Theme
	}
	return "default"
}

func railDate(t time.Time) string    { return t.Format("2 Jan 2006") }
func machineDate(t time.Time) string { return t.Format("2006-01-02") }

// readingTime at ~220 wpm; floor 1 so nothing reads "0 min".
func readingTime(source string) string {
	words := len(strings.Fields(source))
	minutes := max(words/220, 1)
	return fmt.Sprintf("%d min", minutes)
}

func postCount(n int, singular, plural string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", singular)
	}
	return fmt.Sprintf("%d %s", n, plural)
}

func statusOf(published bool) string {
	if published {
		return "live"
	}
	return "draft"
}

func hasLinks(p *db.Project) bool { return p.LinkURL != "" || p.RepoURL != "" }

func editPostPath(id int64) string {
	return fmt.Sprintf("/admin/posts/%d/edit", id)
}

func editProjectPath(id int64) string {
	return fmt.Sprintf("/admin/projects/%d/edit", id)
}

func errorCode(status int) string { return fmt.Sprintf("Error %d", status) }

func statusClass(published bool) string {
	if published {
		return "text-fg"
	}
	return "text-dim"
}

func orderLabel(n int) string { return strconv.Itoa(n) }

// projectIndex formats sort_order as a zero-padded three-digit index.
func projectIndex(sortOrder int) string {
	if sortOrder < 0 {
		sortOrder = 0
	}
	return fmt.Sprintf("%03d", sortOrder)
}

// projectFormat returns the first tech tag, or empty.
func projectFormat(p *db.Project) string {
	tags := p.TechList()
	if len(tags) == 0 {
		return ""
	}
	return tags[0]
}

func workPath(slug string) string { return "/work/" + slug }

func deletePostPath(id int64) string    { return fmt.Sprintf("/admin/posts/%d/delete", id) }
func deleteProjectPath(id int64) string { return fmt.Sprintf("/admin/projects/%d/delete", id) }

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

// DisplayTitle falls back to a body snippet for untitled Micropub notes.
func DisplayTitle(post *db.Post) string {
	if post.Title != "" {
		return post.Title
	}
	return snippet(post.BodyMD, 60)
}

// snippet truncates on a word boundary; max is runes (not bytes) for UTF-8 safety.
func snippet(source string, max int) string {
	fields := strings.Fields(source)
	joined := strings.Join(fields, " ")
	runes := []rune(joined)
	if len(runes) <= max {
		return joined
	}
	cut := string(runes[:max])
	if i := strings.LastIndex(cut, " "); i > 0 {
		cut = cut[:i]
	}
	return cut + "…"
}
