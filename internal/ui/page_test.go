package ui

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/mattbx/mattbx-go/internal/db"
)

func TestDisplayTitlePrefersTitle(t *testing.T) {
	post := &db.Post{Title: "A Real Title", BodyMD: "body text"}
	if got := DisplayTitle(post); got != "A Real Title" {
		t.Errorf("DisplayTitle = %q, want the title", got)
	}
}

func TestDisplayTitleFallsBackToSnippetForNotes(t *testing.T) {
	post := &db.Post{Title: "", BodyMD: "just a quick note with no title at all"}
	got := DisplayTitle(post)
	if got == "" {
		t.Fatal("DisplayTitle returned empty for a note")
	}
	if !strings.HasPrefix(got, "just a quick note") {
		t.Errorf("DisplayTitle = %q, want it to start with the note's own words", got)
	}
}

// A CJK/emoji-heavy note with no spaces is exactly the shape that breaks a
// byte-index truncation: len() counts UTF-8 bytes, and multi-byte runes mean
// a byte cut can land mid-character. This must never produce invalid UTF-8,
// however long the unbroken run of non-ASCII text is.
func TestDisplayTitleTruncatesOnRuneNotByteBoundary(t *testing.T) {
	cases := map[string]string{
		"CJK, no spaces":   strings.Repeat("測試漢字內容", 20),
		"emoji, no spaces": strings.Repeat("🎉🎊🎈🎁🎀", 20),
		"accented, spaced": strings.Repeat("café résumé naïve ", 20),
		"mixed multi-byte": strings.Repeat("日本語とemojiと🎉mixed ", 20),
	}
	for name, body := range cases {
		post := &db.Post{BodyMD: body}
		got := DisplayTitle(post)
		if !utf8.ValidString(got) {
			t.Errorf("%s: DisplayTitle produced invalid UTF-8: %q", name, got)
		}
	}
}

func TestDisplayTitleShortNoteUnchanged(t *testing.T) {
	post := &db.Post{BodyMD: "short"}
	if got := DisplayTitle(post); got != "short" {
		t.Errorf("DisplayTitle = %q, want %q unchanged (under the truncation length)", got, "short")
	}
}
