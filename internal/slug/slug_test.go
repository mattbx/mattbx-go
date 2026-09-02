package slug

import "testing"

func TestMake(t *testing.T) {
	cases := map[string]string{
		"Hello World":            "hello-world",
		"  Trailing & leading  ": "trailing-leading",
		"Café Notes":             "cafe-notes",
		"Go 1.26: What's New?":   "go-1-26-what-s-new",
		"multiple---dashes":      "multiple-dashes",
		"C++ & Rust":             "c-rust",
		"!!!":                    "",
		"":                       "",
		"UPPER":                  "upper",
	}
	for in, want := range cases {
		if got := Make(in); got != want {
			t.Errorf("Make(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMakeTruncatesOnWordBoundary(t *testing.T) {
	long := "this is a very long title that goes on and on and should be truncated somewhere sensible rather than mid word"
	got := Make(long)
	if len(got) > 80 {
		t.Errorf("slug is %d chars, want <= 80: %q", len(got), got)
	}
	if got[len(got)-1] == '-' {
		t.Errorf("slug ends with a dash: %q", got)
	}
}
