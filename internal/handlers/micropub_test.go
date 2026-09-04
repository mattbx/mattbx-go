package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/mattbx/mattbx-go/internal/db"
)

func micropubGet(h http.Handler, path, token string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func micropubPostForm(h http.Handler, token string, form url.Values) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, "/micropub", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func micropubPostJSON(h http.Handler, token string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	r := httptest.NewRequest(http.MethodPost, "/micropub", strings.NewReader(string(b)))
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func decodeJSON(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("response was not JSON: %v\nbody: %s", err, w.Body.String())
	}
	return m
}

// --- Auth -------------------------------------------------------------

func TestMicropubRejectsMissingAndWrongToken(t *testing.T) {
	_, h, _, _ := newTestServer(t)

	for name, w := range map[string]*httptest.ResponseRecorder{
		"no token":        micropubGet(h, "/micropub?q=config", ""),
		"wrong token":     micropubGet(h, "/micropub?q=config", "not-the-token"),
		"post, no token":  micropubPostForm(h, "", url.Values{"h": {"entry"}, "content": {"hi"}}),
		"post, wrong tok": micropubPostForm(h, "wrong", url.Values{"h": {"entry"}, "content": {"hi"}}),
	} {
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s: status = %d, want 401", name, w.Code)
		}
		body := decodeJSON(t, w)
		if body["error"] != "unauthorized" {
			t.Errorf("%s: error = %v, want \"unauthorized\"", name, body["error"])
		}
	}
}

func TestMicropubAcceptsCorrectToken(t *testing.T) {
	_, h, _, _ := newTestServer(t)
	w := micropubGet(h, "/micropub?q=config", micropubToken)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
}

// --- Create -------------------------------------------------------------

func TestMicropubCreateArticleViaForm(t *testing.T) {
	_, h, posts, _ := newTestServer(t)

	w := micropubPostForm(h, micropubToken, url.Values{
		"h":           {"entry"},
		"name":        {"Hello From Micropub"},
		"content":     {"Some **bold** content."},
		"category[]":  {"go", "indieweb"},
		"post-status": {"published"},
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", w.Code, w.Body.String())
	}
	loc := w.Header().Get("Location")
	if !strings.Contains(loc, "/blog/hello-from-micropub") {
		t.Fatalf("Location = %q, want a slug derived from the title", loc)
	}

	post, err := posts.GetBySlug(context.Background(), "hello-from-micropub", false)
	if err != nil {
		t.Fatalf("post was not stored: %v", err)
	}
	if !post.Published {
		t.Error("post-status=published did not publish the post")
	}
	if !strings.Contains(post.BodyHTML, "<strong>bold</strong>") {
		t.Error("content was not rendered as markdown")
	}
	if got := post.TagList(); len(got) != 2 || got[0] != "go" || got[1] != "indieweb" {
		t.Errorf("tags = %v, want [go indieweb]", got)
	}
}

func TestMicropubCreateNoteHasNoTitleAndAutoSlug(t *testing.T) {
	_, h, posts, _ := newTestServer(t)

	w := micropubPostForm(h, micropubToken, url.Values{
		"h":       {"entry"},
		"content": {"just a quick note, no title"},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", w.Code, w.Body.String())
	}
	loc := w.Header().Get("Location")
	slugPart := loc[strings.LastIndex(loc, "/")+1:]
	if !strings.HasPrefix(slugPart, "note-") {
		t.Errorf("auto slug = %q, want a note- prefix", slugPart)
	}

	post, err := posts.GetBySlug(context.Background(), slugPart, false)
	if err != nil {
		t.Fatalf("note was not stored: %v", err)
	}
	if post.Title != "" {
		t.Errorf("note got a title %q, want empty", post.Title)
	}
	// A note with no title defaults to published per spec (post-status
	// omitted means published, same as post-status=published explicitly).
	if !post.Published {
		t.Error("omitting post-status should default to published")
	}
}

func TestMicropubCreateViaJSON(t *testing.T) {
	_, h, posts, _ := newTestServer(t)

	w := micropubPostJSON(h, micropubToken, map[string]any{
		"type": []string{"h-entry"},
		"properties": map[string][]any{
			"name":        {"JSON Post"},
			"content":     {"hello from json"},
			"category":    {"go"},
			"post-status": {"draft"},
		},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", w.Code, w.Body.String())
	}

	post, err := posts.GetBySlug(context.Background(), "json-post", true)
	if err != nil {
		t.Fatalf("post was not stored: %v", err)
	}
	if post.Published {
		t.Error("post-status=draft should not publish")
	}
	// Drafts stay invisible to the public regardless of how they were authored.
	if _, err := posts.GetBySlug(context.Background(), "json-post", false); err != db.ErrNotFound {
		t.Errorf("draft leaked to public lookup: err = %v", err)
	}
}

func TestMicropubMpSlugOverridesTitle(t *testing.T) {
	_, h, posts, _ := newTestServer(t)

	w := micropubPostForm(h, micropubToken, url.Values{
		"h": {"entry"}, "name": {"A Title"}, "content": {"body"}, "mp-slug": {"custom-slug"},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
	if _, err := posts.GetBySlug(context.Background(), "custom-slug", false); err != nil {
		t.Errorf("mp-slug was not honored: %v", err)
	}
}

func TestMicropubDuplicateSlugGetsDeduped(t *testing.T) {
	_, h, posts, _ := newTestServer(t)

	for i := 0; i < 2; i++ {
		w := micropubPostForm(h, micropubToken, url.Values{
			"h": {"entry"}, "name": {"Same Title"}, "content": {"body"},
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("create %d: status = %d: %s", i, w.Code, w.Body.String())
		}
	}
	all, err := posts.List(context.Background(), true, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("got %d posts, want 2", len(all))
	}
	if all[0].Slug == all[1].Slug {
		t.Errorf("duplicate titles produced the same slug: %q", all[0].Slug)
	}
}

func TestMicropubRejectsUnsupportedType(t *testing.T) {
	_, h, _, _ := newTestServer(t)
	w := micropubPostForm(h, micropubToken, url.Values{"h": {"card"}, "content": {"x"}})
	if w.Code != http.StatusBadRequest {
		t.Errorf("h=card: status = %d, want 400", w.Code)
	}
}

func TestMicropubRejectsEmptyPost(t *testing.T) {
	_, h, _, _ := newTestServer(t)
	w := micropubPostForm(h, micropubToken, url.Values{"h": {"entry"}})
	if w.Code != http.StatusBadRequest {
		t.Errorf("no content or name: status = %d, want 400", w.Code)
	}
}

func TestMicropubRejectsMalformedJSON(t *testing.T) {
	_, h, _, _ := newTestServer(t)
	r := httptest.NewRequest(http.MethodPost, "/micropub", strings.NewReader("{not json"))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+micropubToken)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("malformed JSON: status = %d, want 400", w.Code)
	}
}

// --- Update / delete / undelete ------------------------------------------

func TestMicropubUpdateReplacesContentAndCategory(t *testing.T) {
	_, h, posts, _ := newTestServer(t)
	ctx := context.Background()

	orig := &db.Post{Slug: "to-update", Title: "Original", BodyMD: "old", Published: true, Tags: "a, b"}
	if err := posts.Create(ctx, orig); err != nil {
		t.Fatal(err)
	}

	w := micropubPostJSON(h, micropubToken, map[string]any{
		"action": "update",
		"url":    "https://example.test/blog/to-update",
		"replace": map[string][]any{
			"content":  {"new content"},
			"category": {"c", "d"},
		},
	})
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204: %s", w.Code, w.Body.String())
	}

	got, err := posts.GetBySlug(ctx, "to-update", false)
	if err != nil {
		t.Fatal(err)
	}
	if got.BodyMD != "new content" {
		t.Errorf("BodyMD = %q, want replaced", got.BodyMD)
	}
	if got.Title != "Original" {
		t.Errorf("Title = %q, an update must not change what wasn't in replace", got.Title)
	}
	if got.Slug != "to-update" {
		t.Error("update must never change the slug — that's the permalink")
	}
	if tags := got.TagList(); len(tags) != 2 || tags[0] != "c" {
		t.Errorf("tags = %v, want replaced to [c d]", tags)
	}
}

func TestMicropubDeleteUnpublishesRatherThanErasing(t *testing.T) {
	_, h, posts, _ := newTestServer(t)
	ctx := context.Background()

	post := &db.Post{Slug: "to-delete", Title: "Bye", BodyMD: "x", Published: true}
	if err := posts.Create(ctx, post); err != nil {
		t.Fatal(err)
	}

	w := micropubPostForm(h, micropubToken, url.Values{
		"action": {"delete"}, "url": {"https://example.test/blog/to-delete"},
	})
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: status = %d, want 204: %s", w.Code, w.Body.String())
	}

	// Gone from the public read path...
	if _, err := posts.GetBySlug(ctx, "to-delete", false); err != db.ErrNotFound {
		t.Errorf("deleted post still publicly visible: err = %v", err)
	}
	// ...but not actually erased.
	still, err := posts.GetBySlug(ctx, "to-delete", true)
	if err != nil {
		t.Fatalf("delete must not hard-delete the row: %v", err)
	}
	if still.Published {
		t.Error("row still marked published after delete")
	}

	// undelete reverses it.
	w = micropubPostForm(h, micropubToken, url.Values{
		"action": {"undelete"}, "url": {"https://example.test/blog/to-delete"},
	})
	if w.Code != http.StatusNoContent {
		t.Fatalf("undelete: status = %d, want 204: %s", w.Code, w.Body.String())
	}
	if _, err := posts.GetBySlug(ctx, "to-delete", false); err != nil {
		t.Errorf("undelete did not restore public visibility: %v", err)
	}
}

func TestMicropubUpdateUnknownURLNotFound(t *testing.T) {
	_, h, _, _ := newTestServer(t)
	w := micropubPostJSON(h, micropubToken, map[string]any{
		"action":  "update",
		"url":     "https://example.test/blog/does-not-exist",
		"replace": map[string][]any{"content": {"x"}},
	})
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// --- Query: source, config, category --------------------------------------

func TestMicropubSourceReturnsMF2JSON(t *testing.T) {
	_, h, posts, _ := newTestServer(t)
	ctx := context.Background()

	if err := posts.Create(ctx, &db.Post{
		Slug: "for-source", Title: "For Source", BodyMD: "content here",
		Published: true, Tags: "x, y",
	}); err != nil {
		t.Fatal(err)
	}

	w := micropubGet(h, "/micropub?q=source&url="+url.QueryEscape("https://example.test/blog/for-source"), micropubToken)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
	body := decodeJSON(t, w)
	props, ok := body["properties"].(map[string]any)
	if !ok {
		t.Fatalf("no properties in response: %v", body)
	}
	if name, _ := props["name"].([]any); len(name) == 0 || name[0] != "For Source" {
		t.Errorf("properties.name = %v, want [For Source]", props["name"])
	}
	if content, _ := props["content"].([]any); len(content) == 0 || content[0] != "content here" {
		t.Errorf("properties.content = %v", props["content"])
	}
}

func TestMicropubCategoryListsDistinctTags(t *testing.T) {
	_, h, posts, _ := newTestServer(t)
	ctx := context.Background()
	posts.Create(ctx, &db.Post{Slug: "p1", Title: "P1", Published: true, Tags: "go, indieweb"})
	posts.Create(ctx, &db.Post{Slug: "p2", Title: "P2", Published: true, Tags: "go, sqlite"})

	w := micropubGet(h, "/micropub?q=category", micropubToken)
	body := decodeJSON(t, w)
	cats, _ := body["categories"].([]any)
	seen := map[string]bool{}
	for _, c := range cats {
		seen[c.(string)] = true
	}
	for _, want := range []string{"go", "indieweb", "sqlite"} {
		if !seen[want] {
			t.Errorf("categories missing %q, got %v", want, cats)
		}
	}
}

// --- Discovery and headers -------------------------------------------------

func TestMicropubDiscoveryLinkOnHomepage(t *testing.T) {
	_, h, _, _ := newTestServer(t)
	body := get(h, "/").Body.String()
	if !strings.Contains(body, `rel="micropub"`) {
		t.Error("homepage does not advertise the micropub endpoint for client discovery")
	}
}

func TestMicropubPathIsNotIndexable(t *testing.T) {
	_, h, _, _ := newTestServer(t)
	w := micropubGet(h, "/micropub?q=config", micropubToken)
	if got := w.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
}

// --- Fixes from code review -------------------------------------------

func TestMicropubCategoryWithEmbeddedCommaDoesNotSplit(t *testing.T) {
	_, h, posts, _ := newTestServer(t)

	w := micropubPostForm(h, micropubToken, url.Values{
		"h": {"entry"}, "name": {"Comma Tag"}, "content": {"body"},
		"category[]": {"before, after", "second"},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
	post, err := posts.GetBySlug(context.Background(), "comma-tag", false)
	if err != nil {
		t.Fatal(err)
	}
	tags := post.TagList()
	if len(tags) != 2 {
		t.Fatalf("tags = %v, want exactly 2 (the embedded comma must not split a tag in two)", tags)
	}
	if tags[0] != "before after" || tags[1] != "second" {
		t.Errorf("tags = %v, want [\"before after\" \"second\"] (comma swapped for a space, no double space)", tags)
	}
}

func TestMicropubAcceptsBareCategoryFormField(t *testing.T) {
	_, h, posts, _ := newTestServer(t)

	w := micropubPostForm(h, micropubToken, url.Values{
		"h": {"entry"}, "name": {"Bare Category"}, "content": {"body"},
		"category": {"go"}, // no [] suffix
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
	post, err := posts.GetBySlug(context.Background(), "bare-category", false)
	if err != nil {
		t.Fatal(err)
	}
	if tags := post.TagList(); len(tags) != 1 || tags[0] != "go" {
		t.Errorf("tags = %v, want [go] from the bracket-less category field", tags)
	}
}

func TestMicropubSourceHonorsPropertiesFilter(t *testing.T) {
	_, h, posts, _ := newTestServer(t)
	ctx := context.Background()
	if err := posts.Create(ctx, &db.Post{
		Slug: "filtered", Title: "Filtered", BodyMD: "content", Published: true, Tags: "a",
	}); err != nil {
		t.Fatal(err)
	}

	q := url.Values{
		"q":            {"source"},
		"url":          {"https://example.test/blog/filtered"},
		"properties[]": {"content"},
	}
	w := micropubGet(h, "/micropub?"+q.Encode(), micropubToken)
	body := decodeJSON(t, w)
	props, _ := body["properties"].(map[string]any)

	if _, ok := props["content"]; !ok {
		t.Error("requested property \"content\" missing from filtered response")
	}
	if _, ok := props["name"]; ok {
		t.Error("unrequested property \"name\" leaked into a filtered response")
	}
	if _, ok := props["category"]; ok {
		t.Error("unrequested property \"category\" leaked into a filtered response")
	}
}

func TestMicropubUpdateDeleteAsArrayClearsCategory(t *testing.T) {
	_, h, posts, _ := newTestServer(t)
	ctx := context.Background()
	if err := posts.Create(ctx, &db.Post{
		Slug: "del1", Title: "Del1", Published: true, Tags: "a, b",
	}); err != nil {
		t.Fatal(err)
	}

	w := micropubPostJSON(h, micropubToken, map[string]any{
		"action": "update",
		"url":    "https://example.test/blog/del1",
		"delete": []string{"category"},
	})
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204: %s", w.Code, w.Body.String())
	}
	got, err := posts.GetBySlug(ctx, "del1", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.TagList()) != 0 {
		t.Errorf("tags = %v, want cleared by delete:[\"category\"]", got.TagList())
	}
}

func TestMicropubUpdateDeleteAsObjectRemovesSpecificValues(t *testing.T) {
	_, h, posts, _ := newTestServer(t)
	ctx := context.Background()
	if err := posts.Create(ctx, &db.Post{
		Slug: "del2", Title: "Del2", Published: true, Tags: "a, b, c",
	}); err != nil {
		t.Fatal(err)
	}

	w := micropubPostJSON(h, micropubToken, map[string]any{
		"action": "update",
		"url":    "https://example.test/blog/del2",
		"delete": map[string][]string{"category": {"b"}},
	})
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204: %s", w.Code, w.Body.String())
	}
	got, err := posts.GetBySlug(ctx, "del2", false)
	if err != nil {
		t.Fatal(err)
	}
	tags := got.TagList()
	if len(tags) != 2 || tags[0] != "a" || tags[1] != "c" {
		t.Errorf("tags = %v, want [a c] (only \"b\" removed)", tags)
	}
}

// A delete request naming a property we can't actually remove must fail
// loudly (400), not silently succeed (204) while doing nothing — a 204 there
// would tell the client its request worked when it didn't.
func TestMicropubUpdateDeleteUnsupportedPropertyRejected(t *testing.T) {
	_, h, posts, _ := newTestServer(t)
	ctx := context.Background()
	if err := posts.Create(ctx, &db.Post{
		Slug: "del3", Title: "Keep Me", BodyMD: "keep this content", Published: true,
	}); err != nil {
		t.Fatal(err)
	}

	for _, shape := range []any{
		[]string{"content"},
		map[string][]string{"name": {"x"}},
	} {
		w := micropubPostJSON(h, micropubToken, map[string]any{
			"action": "update",
			"url":    "https://example.test/blog/del3",
			"delete": shape,
		})
		if w.Code != http.StatusBadRequest {
			t.Errorf("delete %#v: status = %d, want 400", shape, w.Code)
		}
	}

	// And nothing was actually mutated by the rejected attempts.
	got, err := posts.GetBySlug(ctx, "del3", false)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Keep Me" || got.BodyMD != "keep this content" {
		t.Errorf("a rejected delete still mutated the post: title=%q body=%q", got.Title, got.BodyMD)
	}
}

// The auth check is now router-level middleware rather than an inline call
// in each handler; this confirms the handlers themselves are unreachable
// without it, i.e. the middleware genuinely gates them rather than just
// happening to run first.
func TestMicropubMiddlewareGatesBothRoutes(t *testing.T) {
	_, h, _, _ := newTestServer(t)

	get := micropubGet(h, "/micropub?q=config", "")
	if get.Code != http.StatusUnauthorized {
		t.Errorf("GET without token: status = %d, want 401", get.Code)
	}
	post := micropubPostForm(h, "", url.Values{"h": {"entry"}, "content": {"x"}})
	if post.Code != http.StatusUnauthorized {
		t.Errorf("POST without token: status = %d, want 401", post.Code)
	}
}
