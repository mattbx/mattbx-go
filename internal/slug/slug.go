// Package slug derives URL-safe identifiers from titles.
package slug

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)
	trimDashes      = regexp.MustCompile(`^-+|-+$`)
)

// Make converts a title into a lowercase, hyphenated slug. Accented characters
// are folded to ASCII ("Café Notes" -> "cafe-notes") so URLs stay clean.
// Returns "" when nothing usable remains; callers should fall back to an ID.
func Make(title string) string {
	folded, _, err := transform.String(
		transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC),
		title,
	)
	if err != nil {
		folded = title
	}
	s := strings.ToLower(folded)
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	s = trimDashes.ReplaceAllString(s, "")

	const maxLen = 80
	if len(s) > maxLen {
		s = s[:maxLen]
		if i := strings.LastIndex(s, "-"); i > 0 {
			s = s[:i]
		}
	}
	return s
}
