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
	SiteName    = "Hi there, I'm Matt!"
	SiteMark    = "Mbx"
	SiteRole    = "UI/UX Engineer"
	SiteTagline = "I design and build digital products and things with my dog Pippi, in Sydney AU.*Available for work opportunities and tacos.*This is mostly placeholder content for now until I have something to say."
)

// Page is per-request shell context from handlers.
type Page struct {
	Title       string // SiteName appended by the shell when set
	Description string
	Nav         string // "blog" | "about" | "portfolio" | ""
	// IsAdmin is presentation only (toolbar). Access is middleware-gated.
	IsAdmin bool
	BaseURL string
	Path    string
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
