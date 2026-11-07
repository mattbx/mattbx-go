package ui

import (
	"context"
	"strings"
	"testing"
)

func TestProfileParsesIntroAndSections(t *testing.T) {
	content, err := parseProfile(profileSource)
	if err != nil {
		t.Fatalf("parseProfile: %v", err)
	}
	if strings.TrimSpace(content.IntroHTML) == "" {
		t.Fatal("intro is empty")
	}
	want := []string{"Experience", "Earlier career", "Skills", "Selected clients"}
	if len(content.Sections) != len(want) {
		t.Fatalf("got %d sections, want %d", len(content.Sections), len(want))
	}
	for i, title := range want {
		if content.Sections[i].Title != title {
			t.Errorf("section %d: got %q, want %q", i, content.Sections[i].Title, title)
		}
		if strings.TrimSpace(content.Sections[i].HTML) == "" {
			t.Errorf("section %q has empty body", title)
		}
	}
}

// House style: short hyphens only. The renderer's Typographer would turn
// `--` into a dash, and em dashes are never typed directly.
func TestProfileSourceHouseStyle(t *testing.T) {
	if strings.Contains(profileSource, "—") {
		t.Error("profile.md contains an em dash")
	}
	if strings.Contains(profileSource, "--") {
		t.Error("profile.md contains `--`")
	}
}

func TestParseProfileFailsLoudly(t *testing.T) {
	if _, err := parseProfile(""); err == nil {
		t.Error("empty source: want error, got nil")
	}
	if _, err := parseProfile("## Only a section\n\nbody\n"); err == nil {
		t.Error("missing intro: want error, got nil")
	}
	if _, err := parseProfile("Intro only, no sections.\n"); err == nil {
		t.Error("missing sections: want error, got nil")
	}
}

func TestProfilePanelRendersSectionsAndLinks(t *testing.T) {
	var sb strings.Builder
	if err := ProfilePanel(Page{}).Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := sb.String()
	for _, want := range []string{
		"Experience",
		"Earlier career",
		"Skills",
		"Selected clients",
		SiteGitHub,
		"GitHub ↗",
		"Get in touch ↗",
		"LinkedIn ↗",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("drawer missing %q", want)
		}
	}
}
