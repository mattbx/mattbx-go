// Package markdown converts post and project bodies from Markdown to
// sanitized HTML.
//
// Rendering happens once at save time, not per request, so the public read
// path is a single query and a write. Sanitizing at save time also means
// nothing untrusted is ever stored.
package markdown

import (
	"bytes"
	"fmt"
	"regexp"
	"sync"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

// ChromaStyle is the syntax-highlighting theme. It is emitted as CSS classes
// (not inline styles), so the matching stylesheet must be generated into
// internal/ui/static/chroma.css — see `go generate ./internal/ui/...`.
const ChromaStyle = "github"

// cssClassPattern matches the short class tokens Chroma emits ("chroma", "kd",
// "nx", …) plus goldmark's own. Anything else is stripped.
var cssClassPattern = regexp.MustCompile(`^[a-zA-Z0-9\-_ ]+$`)

var (
	once      sync.Once
	converter goldmark.Markdown
	policy    *bluemonday.Policy
)

func initOnce() {
	once.Do(func() {
		converter = goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,         // tables, strikethrough, task lists, autolinks
				extension.Typographer, // smart quotes and dashes
				extension.Footnote,
				highlighting.NewHighlighting(
					highlighting.WithStyle(ChromaStyle),
					// Classes rather than inline styles: the sanitizer would
					// have to allow style="" otherwise, which is a far larger
					// attack surface than a class allowlist.
					highlighting.WithFormatOptions(chromahtml.WithClasses(true)),
				),
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(), // stable #anchors for deep links
			),
		)

		policy = bluemonday.UGCPolicy()
		policy.AllowAttrs("class").Matching(cssClassPattern).OnElements("span", "code", "pre", "div")
		policy.AllowAttrs("id").Matching(cssClassPattern).OnElements("h1", "h2", "h3", "h4", "h5", "h6", "section", "li", "sup")
		policy.AllowAttrs("href").Matching(regexp.MustCompile(`^#[a-zA-Z0-9\-_]+$`)).OnElements("a")
		policy.AllowAttrs("start", "reversed").OnElements("ol")
		policy.AllowAttrs("align").OnElements("td", "th")
		// Content is first-party (only /admin can author it), so nofollow adds
		// nothing — but external links should still not get window.opener.
		policy.RequireNoFollowOnLinks(false)
		policy.AddTargetBlankToFullyQualifiedLinks(true)
		policy.RequireNoReferrerOnFullyQualifiedLinks(true)
	})
}

// Render converts Markdown to sanitized HTML ready to embed in a page.
func Render(source string) (string, error) {
	initOnce()

	var buf bytes.Buffer
	if err := converter.Convert([]byte(source), &buf); err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}
	return policy.Sanitize(buf.String()), nil
}

// StyleSheet returns the CSS for ChromaStyle. Used by go:generate to write
// internal/ui/static/chroma.css so highlighting works without inline styles.
func StyleSheet() (string, error) {
	style := styles.Get(ChromaStyle)
	if style == nil {
		return "", fmt.Errorf("unknown chroma style %q", ChromaStyle)
	}
	var buf bytes.Buffer
	formatter := chromahtml.New(chromahtml.WithClasses(true))
	if err := formatter.WriteCSS(&buf, style); err != nil {
		return "", err
	}
	return buf.String(), nil
}
