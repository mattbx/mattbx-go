// Micropub (https://micropub.spec.indieweb.org/) lets any IndieWeb client —
// a phone shortcut, an app, a script — publish here without a site-specific
// integration. Auth is a single static bearer token rather than full
// IndieAuth: this is a single-owner site, so the OAuth dance that lets an
// arbitrary user authenticate against their own server buys nothing here.
// The same simplification is what dansim.work's own Micropub endpoint does.
//
// Deliberately unsupported: the media endpoint. There is no blob storage
// anywhere in this app (SQLite holds text; the volume isn't meant for
// arbitrary binary uploads), so photo posts are out of scope until that's a
// real decision, not a side effect of this endpoint.
//
// "delete" is implemented as unpublishing rather than a hard delete, and
// "undelete" reverses it. A Micropub client mis-click shouldn't be able to
// destroy content outright; this reuses the existing draft mechanism, which
// already has nowhere-but-admin visibility for unpublished posts.
package handlers

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mattbx/mattbx-go/internal/db"
	"github.com/mattbx/mattbx-go/internal/markdown"
	"github.com/mattbx/mattbx-go/internal/slug"
)

// mf2Post is the microformats2-json shape for an h-entry, both incoming
// (create/update) and outgoing (the ?q=source response). Every property is a
// JSON array per the mf2-json convention, even when single-valued.
type mf2Post struct {
	Type       []string         `json:"type,omitempty"`
	Properties map[string][]any `json:"properties,omitempty"`
	Action     string           `json:"action,omitempty"`
	URL        string           `json:"url,omitempty"`
	Replace    map[string][]any `json:"replace,omitempty"`
	Add        map[string][]any `json:"add,omitempty"`
	Delete     json.RawMessage  `json:"delete,omitempty"` // spec allows an array (drop props) or an object (remove values)
}

// micropubError writes the JSON error shape the spec (and dansim.work's live
// endpoint, confirmed by hand) both use.
func micropubError(w http.ResponseWriter, status int, code, description string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":             code,
		"error_description": description,
	})
}

// requireMicropubToken checks the bearer token and reports whether the
// request may proceed. Constant-time compare, matching the password checks
// in internal/auth.
func (s *Server) requireMicropubToken(w http.ResponseWriter, r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(auth, "Bearer ")
	if !ok || token == "" {
		micropubError(w, http.StatusUnauthorized, "unauthorized", "no access token supplied")
		return false
	}
	if subtle.ConstantTimeCompare([]byte(token), []byte(s.cfg.MicropubToken)) != 1 {
		micropubError(w, http.StatusUnauthorized, "unauthorized", "invalid access token")
		return false
	}
	return true
}

// handleMicropubQuery serves GET /micropub — client introspection (?q=config,
// ?q=source, ?q=category), used by clients to build their own posting UI and
// to fetch a post back for editing.
func (s *Server) handleMicropubQuery(w http.ResponseWriter, r *http.Request) {
	if !s.requireMicropubToken(w, r) {
		return
	}

	switch r.URL.Query().Get("q") {
	case "config":
		s.micropubConfig(w, r)
	case "source":
		s.micropubSource(w, r)
	case "category":
		s.micropubCategories(w, r)
	default:
		micropubError(w, http.StatusBadRequest, "invalid_request", "unsupported or missing ?q=")
	}
}

func (s *Server) micropubConfig(w http.ResponseWriter, r *http.Request) {
	// No media-endpoint key: there is nowhere to upload to yet.
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (s *Server) micropubCategories(w http.ResponseWriter, r *http.Request) {
	posts, err := s.posts.List(r.Context(), true, 0)
	if err != nil {
		s.log.Error("micropub categories", "err", err)
		micropubError(w, http.StatusInternalServerError, "internal", "could not list categories")
		return
	}
	seen := map[string]bool{}
	var out []string
	for _, p := range posts {
		for _, t := range p.TagList() {
			if !seen[t] {
				seen[t] = true
				out = append(out, t)
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"categories": out})
}

func (s *Server) micropubSource(w http.ResponseWriter, r *http.Request) {
	post, ok := s.postFromMicropubURL(w, r, r.URL.Query().Get("url"))
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, postToMF2(post))
}

// postFromMicropubURL resolves a Micropub "url" property (an absolute post
// URL, e.g. https://example.com/blog/my-slug) back to the stored post,
// writing an error response and returning ok=false on any failure.
func (s *Server) postFromMicropubURL(w http.ResponseWriter, r *http.Request, raw string) (*db.Post, bool) {
	if raw == "" {
		micropubError(w, http.StatusBadRequest, "invalid_request", "url is required")
		return nil, false
	}
	u, err := url.Parse(raw)
	if err != nil {
		micropubError(w, http.StatusBadRequest, "invalid_request", "url is not a valid URL")
		return nil, false
	}
	slugPart := strings.TrimPrefix(u.Path, "/blog/") // matches Post.PermalinkPath's shape
	if slugPart == u.Path || slugPart == "" {
		micropubError(w, http.StatusNotFound, "not_found", "no post at that url")
		return nil, false
	}
	post, err := s.posts.GetBySlug(r.Context(), slugPart, true)
	if err != nil {
		micropubError(w, http.StatusNotFound, "not_found", "no post at that url")
		return nil, false
	}
	return post, true
}

// postToMF2 renders a stored post back into mf2-json, the shape a client
// expects from ?q=source when it wants to load a post for editing.
func postToMF2(post *db.Post) mf2Post {
	props := map[string][]any{
		"content": {post.BodyMD},
	}
	if post.Title != "" {
		props["name"] = []any{post.Title}
	}
	if len(post.TagList()) > 0 {
		cats := make([]any, len(post.TagList()))
		for i, t := range post.TagList() {
			cats[i] = t
		}
		props["category"] = cats
	}
	status := "published"
	if !post.Published {
		status = "draft"
	}
	props["post-status"] = []any{status}
	props["url"] = []any{post.PermalinkPath()}

	return mf2Post{Type: []string{"h-entry"}, Properties: props}
}

// handleMicropubAction serves POST /micropub — create, update, delete, and
// undelete.
func (s *Server) handleMicropubAction(w http.ResponseWriter, r *http.Request) {
	if !s.requireMicropubToken(w, r) {
		return
	}

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		s.micropubJSON(w, r)
		return
	}
	s.micropubForm(w, r)
}

// micropubForm handles application/x-www-form-urlencoded (and multipart
// without file parts) submissions — the classic Micropub request shape, and
// the one most simple clients and command-line tools use. Per spec, updates
// and deletes are JSON-only here: form-encoded update semantics (nested
// replace/add/delete of specific properties) don't have a clean form
// encoding, and every real client that does updates sends JSON anyway.
func (s *Server) micropubForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil && err != http.ErrNotMultipart {
		micropubError(w, http.StatusBadRequest, "invalid_request", "could not parse request body")
		return
	}
	if r.Form == nil {
		if err := r.ParseForm(); err != nil {
			micropubError(w, http.StatusBadRequest, "invalid_request", "could not parse request body")
			return
		}
	}

	switch r.FormValue("action") {
	case "", "create":
		// fallthrough to create below
	case "delete":
		s.micropubSetPublished(w, r, r.FormValue("url"), false)
		return
	case "undelete":
		s.micropubSetPublished(w, r, r.FormValue("url"), true)
		return
	case "update":
		micropubError(w, http.StatusBadRequest, "invalid_request", "updates must be sent as application/json")
		return
	default:
		micropubError(w, http.StatusBadRequest, "invalid_request", "unsupported action")
		return
	}

	h := r.FormValue("h")
	if h == "" {
		h = "entry"
	}
	if h != "entry" {
		micropubError(w, http.StatusBadRequest, "invalid_request", "only h=entry is supported")
		return
	}

	s.micropubCreate(w, r, micropubInput{
		Content:    r.FormValue("content"),
		Name:       r.FormValue("name"),
		Slug:       r.FormValue("mp-slug"),
		Categories: r.Form["category[]"],
		Status:     r.FormValue("post-status"),
	})
}

// micropubJSON handles the mf2-json request shape.
func (s *Server) micropubJSON(w http.ResponseWriter, r *http.Request) {
	var body mf2Post
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		micropubError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}

	switch body.Action {
	case "delete":
		s.micropubSetPublished(w, r, body.URL, false)
		return
	case "undelete":
		s.micropubSetPublished(w, r, body.URL, true)
		return
	case "update":
		s.micropubUpdate(w, r, body)
		return
	case "", "create":
		// fallthrough
	default:
		micropubError(w, http.StatusBadRequest, "invalid_request", "unsupported action")
		return
	}

	h := "entry"
	if len(body.Type) > 0 {
		h = strings.TrimPrefix(body.Type[0], "h-")
	}
	if h != "entry" {
		micropubError(w, http.StatusBadRequest, "invalid_request", "only h-entry is supported")
		return
	}

	s.micropubCreate(w, r, micropubInput{
		Content:    mf2String(body.Properties, "content"),
		Name:       mf2String(body.Properties, "name"),
		Slug:       mf2String(body.Properties, "mp-slug"),
		Categories: mf2Strings(body.Properties, "category"),
		Status:     mf2String(body.Properties, "post-status"),
	})
}

// micropubInput is the request normalized from either form or JSON encoding,
// so the actual create logic below doesn't care which one arrived.
type micropubInput struct {
	Content    string
	Name       string
	Slug       string
	Categories []string
	Status     string
}

func (s *Server) micropubCreate(w http.ResponseWriter, r *http.Request, in micropubInput) {
	// mf2 content is sometimes an object ({"html": "..."} or similar) rather
	// than a plain string; mf2String already unwraps the common sub-keys, so
	// by the time we get here Content is the raw source text either way.
	if in.Content == "" && in.Name == "" {
		micropubError(w, http.StatusBadRequest, "invalid_request", "content or name is required")
		return
	}

	post := &db.Post{
		Title:     in.Name,
		BodyMD:    in.Content,
		Tags:      strings.Join(in.Categories, ", "),
		Published: in.Status != "draft",
	}

	html, err := markdown.Render(post.BodyMD)
	if err != nil {
		micropubError(w, http.StatusInternalServerError, "internal", "could not render content")
		return
	}
	post.BodyHTML = html

	postSlug, err := s.uniqueSlug(r.Context(), in.Slug, in.Name)
	if err != nil {
		micropubError(w, http.StatusInternalServerError, "internal", "could not generate a unique slug")
		return
	}
	post.Slug = postSlug

	if err := s.posts.Create(r.Context(), post); err != nil {
		s.log.Error("micropub create", "err", err)
		micropubError(w, http.StatusInternalServerError, "internal", "could not save post")
		return
	}

	w.Header().Set("Location", s.cfg.BaseURL+post.PermalinkPath())
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) micropubUpdate(w http.ResponseWriter, r *http.Request, body mf2Post) {
	post, ok := s.postFromMicropubURL(w, r, body.URL)
	if !ok {
		return
	}

	// The slug is the permalink; updates change content, never the URL.
	if v := mf2String(body.Replace, "content"); v != "" {
		post.BodyMD = v
		html, err := markdown.Render(post.BodyMD)
		if err != nil {
			micropubError(w, http.StatusInternalServerError, "internal", "could not render content")
			return
		}
		post.BodyHTML = html
	}
	if v := mf2String(body.Replace, "name"); v != "" {
		post.Title = v
	}
	if cats := mf2Strings(body.Replace, "category"); cats != nil {
		post.Tags = strings.Join(cats, ", ")
	}
	if v := mf2String(body.Replace, "post-status"); v != "" {
		post.Published = v != "draft"
	}
	if cats := mf2Strings(body.Add, "category"); len(cats) > 0 {
		merged := append(post.TagList(), cats...)
		post.Tags = strings.Join(merged, ", ")
	}

	if err := s.posts.Update(r.Context(), post); err != nil {
		s.log.Error("micropub update", "err", err)
		micropubError(w, http.StatusInternalServerError, "internal", "could not update post")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// micropubSetPublished implements delete (unpublish) and undelete
// (republish) — see the package doc comment for why this is a soft toggle
// rather than a real DELETE.
func (s *Server) micropubSetPublished(w http.ResponseWriter, r *http.Request, rawURL string, published bool) {
	post, ok := s.postFromMicropubURL(w, r, rawURL)
	if !ok {
		return
	}
	post.Published = published
	if err := s.posts.Update(r.Context(), post); err != nil {
		s.log.Error("micropub set-published", "err", err, "published", published)
		micropubError(w, http.StatusInternalServerError, "internal", "could not update post")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// uniqueSlug prefers an explicit mp-slug, falls back to deriving one from the
// title, and falls further back to a timestamp for title-less notes — then
// de-dupes against existing posts the same way the admin form does.
func (s *Server) uniqueSlug(ctx context.Context, explicit, title string) (string, error) {
	candidate := slug.Make(explicit)
	if candidate == "" {
		candidate = slug.Make(title)
	}
	if candidate == "" {
		candidate = "note-" + time.Now().UTC().Format("20060102-150405")
	}

	for attempt := 0; attempt < 20; attempt++ {
		try := candidate
		if attempt > 0 {
			try = fmt.Sprintf("%s-%d", candidate, attempt+1)
		}
		taken, err := s.posts.SlugTaken(ctx, try, 0)
		if err != nil {
			return "", err
		}
		if !taken {
			return try, nil
		}
	}
	return "", fmt.Errorf("could not find a unique slug after 20 attempts")
}

// mf2String reads a single-valued mf2-json property, unwrapping the common
// {"html": "..."} / {"markdown": "..."} / {"value": "..."} shapes a content
// object might arrive as. Plain strings pass through unchanged.
func mf2String(props map[string][]any, key string) string {
	vals, ok := props[key]
	if !ok || len(vals) == 0 {
		return ""
	}
	switch v := vals[0].(type) {
	case string:
		return v
	case map[string]any:
		for _, sub := range []string{"markdown", "html", "value"} {
			if s, ok := v[sub].(string); ok {
				return s
			}
		}
	}
	return ""
}

// mf2Strings reads a multi-valued mf2-json property (e.g. category[]).
// Returns nil (not an empty slice) when the key is absent, so callers can
// tell "not provided" apart from "provided as empty" where that distinction
// matters (see micropubUpdate's use with Add).
func mf2Strings(props map[string][]any, key string) []string {
	vals, ok := props[key]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		if s, ok := v.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
