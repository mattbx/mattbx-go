package markdown

import (
	"strings"
	"testing"
)

func TestRenderBasics(t *testing.T) {
	got, err := Render("# Title\n\nSome **bold** text and a [link](/blog).")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"<h1", "Title", "<strong>bold</strong>", `href="/blog"`} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\ngot: %s", want, got)
		}
	}
}

// Sanitization is defence in depth: only /admin can author content, but a
// stored-XSS bug here would be permanent, so verify the dangerous shapes.
func TestRenderStripsDangerousMarkup(t *testing.T) {
	cases := map[string]string{
		"inline script":    "<script>alert(1)</script>",
		"event handler":    `<img src=x onerror="alert(1)">`,
		"javascript href":  `<a href="javascript:alert(1)">click</a>`,
		"iframe":           `<iframe src="https://evil.com"></iframe>`,
		"style attribute":  `<p style="position:fixed;top:0">hi</p>`,
		"object embed":     `<object data="evil.swf"></object>`,
		"form":             `<form action="https://evil.com"><input name=pw></form>`,
		"svg onload":       `<svg onload="alert(1)"></svg>`,
		"markdown js link": `[click](javascript:alert%281%29)`,
	}

	for name, src := range cases {
		out, err := Render(src)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		lower := strings.ToLower(out)
		for _, forbidden := range []string{"<script", "onerror", "onload", "javascript:", "<iframe", "style=", "<object", "<form"} {
			if strings.Contains(lower, forbidden) {
				t.Errorf("%s: %q survived sanitization\ngot: %s", name, forbidden, out)
			}
		}
	}
}

func TestRenderHighlightsCodeWithClasses(t *testing.T) {
	out, err := Render("```go\nfunc main() {}\n```")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "chroma") {
		t.Errorf("code block was not highlighted\ngot: %s", out)
	}
	if !strings.Contains(out, "class=") {
		t.Errorf("highlighting must survive sanitization as classes\ngot: %s", out)
	}
	// Inline styles would mean the sanitizer had to allow style="".
	if strings.Contains(out, "style=") {
		t.Errorf("expected class-based highlighting, got inline styles\n%s", out)
	}
}

func TestRenderGFMAndHeadingAnchors(t *testing.T) {
	out, err := Render("## A Heading\n\n| a | b |\n|---|---|\n| 1 | 2 |\n\n- [x] done\n")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"<table>", `id="a-heading"`} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot: %s", want, out)
		}
	}
}

func TestStyleSheetGenerates(t *testing.T) {
	css, err := StyleSheet()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(css, ".chroma") {
		t.Errorf("stylesheet does not define .chroma rules:\n%s", css[:min(200, len(css))])
	}
}
