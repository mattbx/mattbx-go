package ui

import (
	_ "embed"
	"fmt"
	"strings"
	"sync"

	"github.com/mattbx/mattbx-go/internal/markdown"
)

// The profile drawer's resume lives in one markdown file, edited directly —
// no admin screens, no database. Everything before the first `## ` heading is
// the intro; each `## ` heading starts a section. Parsed and rendered once.
//
//go:embed content/profile.md
var profileSource string

// ProfileSection is one `## ` section of profile.md, rendered to HTML.
type ProfileSection struct {
	Title string
	HTML  string
}

type profileContent struct {
	IntroHTML string
	// Summary is the intro's longest paragraph as plain text (the bio, not
	// the title or location lines). Used as the /profile meta description.
	Summary  string
	Sections []ProfileSection
}

// ProfileSummary is the resume's summary paragraph, for page descriptions.
func ProfileSummary() string { return loadProfile().Summary }

// Force the parse at startup: a broken profile.md crashes the binary before
// it serves, so Disco's health check fails the new container and the old
// version keeps running — a bad edit can never 500 every page in production.
func init() { loadProfile() }

var loadProfile = sync.OnceValue(func() profileContent {
	content, err := parseProfile(profileSource)
	if err != nil {
		// A broken or empty file fails loudly here — any test that renders
		// the drawer (and the first request after a bad deploy) hits this,
		// never a silently empty panel.
		panic(fmt.Sprintf("ui: content/profile.md: %v", err))
	}
	return content
})

func parseProfile(source string) (profileContent, error) {
	var content profileContent

	lines := strings.Split(source, "\n")
	var part []string
	title := "" // empty while collecting the intro
	flush := func() error {
		body := strings.TrimSpace(strings.Join(part, "\n"))
		if body == "" {
			return fmt.Errorf("empty %s", partName(title))
		}
		html, err := markdown.Render(body)
		if err != nil {
			return fmt.Errorf("render %s: %w", partName(title), err)
		}
		if title == "" {
			content.IntroHTML = html
			content.Summary = longestParagraph(body)
		} else {
			content.Sections = append(content.Sections, ProfileSection{Title: title, HTML: html})
		}
		return nil
	}

	for _, line := range lines {
		if heading, ok := strings.CutPrefix(line, "## "); ok && !strings.HasPrefix(heading, "#") {
			if err := flush(); err != nil {
				return profileContent{}, err
			}
			title = strings.TrimSpace(heading)
			if title == "" {
				return profileContent{}, fmt.Errorf("section with empty title")
			}
			part = part[:0]
			continue
		}
		part = append(part, line)
	}
	if err := flush(); err != nil {
		return profileContent{}, err
	}
	if title == "" {
		return profileContent{}, fmt.Errorf("no `## ` sections found")
	}
	return content, nil
}

// longestParagraph returns the longest blank-line-separated paragraph,
// with its line breaks collapsed to spaces.
func longestParagraph(source string) string {
	longest := ""
	for _, para := range strings.Split(source, "\n\n") {
		para = strings.Join(strings.Fields(para), " ")
		if len(para) > len(longest) {
			longest = para
		}
	}
	return longest
}

func partName(title string) string {
	if title == "" {
		return "intro"
	}
	return fmt.Sprintf("section %q", title)
}

func profileSectionID(i int) string {
	return fmt.Sprintf("profile-section-%d", i)
}
