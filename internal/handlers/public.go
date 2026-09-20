package handlers

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"time"

	"github.com/mattbx/mattbx-go/internal/ui"
)

// homeRecentPosts caps how many entries the front page shows.
const homeRecentPosts = 5

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	// Admins see their own drafts inline; everyone else gets published only.
	posts, err := s.posts.List(r.Context(), s.isAdmin(r), homeRecentPosts)
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	p := s.page(r, "", ui.SiteTagline, "")
	s.render(w, r, http.StatusOK, ui.Home(p, posts))
}

func (s *Server) handleAbout(w http.ResponseWriter, r *http.Request) {
	p := s.page(r, "About", "A little about "+ui.SiteName+".", "about")
	s.render(w, r, http.StatusOK, ui.About(p))
}

func (s *Server) handleBlogIndex(w http.ResponseWriter, r *http.Request) {
	posts, err := s.posts.List(r.Context(), s.isAdmin(r), 0)
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	p := s.page(r, "Blog", "Notes on the software I build.", "blog")
	s.render(w, r, http.StatusOK, ui.BlogIndex(p, posts))
}

func (s *Server) handleBlogPost(w http.ResponseWriter, r *http.Request) {
	post, err := s.posts.GetBySlug(r.Context(), r.PathValue("slug"), s.isAdmin(r))
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}
	p := s.page(r, post.Title, post.Summary, "blog")
	s.render(w, r, http.StatusOK, ui.BlogPost(p, post))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Ping the database, not just the process: Disco swaps traffic to the new
	// container when this passes, and a server that can't reach SQLite is not
	// ready to serve.
	ctx, cancel := timeoutContext(r, 2*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		s.log.Error("health check failed", "err", err)
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("ok"))
}

func (s *Server) handleRobots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "User-agent: *\nDisallow: /admin\nDisallow: /portfolio\n\nSitemap: %s/feed.xml\n", s.cfg.BaseURL)
}

// --- RSS ------------------------------------------------------------------

type rss struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Atom    string     `xml:"xmlns:atom,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language"`
	AtomLink    atomLink  `xml:"atom:link"`
	Items       []rssItem `xml:"item"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
}

// handleFeed serves published posts only — drafts never appear in the feed,
// regardless of who is signed in.
func (s *Server) handleFeed(w http.ResponseWriter, r *http.Request) {
	posts, err := s.posts.List(r.Context(), false, 50)
	if err != nil {
		s.serverError(w, r, err)
		return
	}

	feed := rss{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		Channel: rssChannel{
			Title:       ui.SiteName,
			Link:        s.cfg.BaseURL,
			Description: ui.SiteTagline,
			Language:    "en",
			AtomLink: atomLink{
				Href: s.cfg.BaseURL + "/feed.xml",
				Rel:  "self",
				Type: "application/rss+xml",
			},
		},
	}

	for _, post := range posts {
		url := s.cfg.BaseURL + "/blog/" + post.Slug
		feed.Channel.Items = append(feed.Channel.Items, rssItem{
			Title:       ui.DisplayTitle(post),
			Link:        url,
			GUID:        url,
			PubDate:     post.Date().Format(time.RFC1123Z),
			Description: post.Summary,
		})
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Write([]byte(xml.Header))

	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(feed); err != nil {
		// Headers are already sent, so log and stop rather than write a 500.
		s.log.Error("encode feed", "err", err)
	}
}

// --- JSON Feed -------------------------------------------------------------

type jsonFeed struct {
	Version     string          `json:"version"`
	Title       string          `json:"title"`
	HomePageURL string          `json:"home_page_url"`
	FeedURL     string          `json:"feed_url"`
	Description string          `json:"description"`
	Language    string          `json:"language"`
	Authors     []jsonFeedAuthor `json:"authors"`
	Items       []jsonFeedItem  `json:"items"`
}

type jsonFeedAuthor struct {
	Name string `json:"name"`
}

type jsonFeedItem struct {
	ID            string   `json:"id"`
	URL           string   `json:"url"`
	Title         string   `json:"title,omitempty"`
	Summary       string   `json:"summary,omitempty"`
	ContentHTML   string   `json:"content_html"`
	DatePublished string   `json:"date_published"`
	DateModified  string   `json:"date_modified"`
	Tags          []string `json:"tags,omitempty"`
}

// feedContentHTML prepares stored HTML for syndication. Chroma escapes string
// literals as entities (&#34; etc.); feed readers are happier with the decoded
// characters. Go's JSON encoder also escapes < as \u003c by default — callers
// should set SetEscapeHTML(false) when writing the feed.
func feedContentHTML(s string) string {
	return html.UnescapeString(s)
}

// handleJSONFeed serves published posts only — drafts never appear in the feed,
// regardless of who is signed in.
func (s *Server) handleJSONFeed(w http.ResponseWriter, r *http.Request) {
	posts, err := s.posts.List(r.Context(), false, 50)
	if err != nil {
		s.serverError(w, r, err)
		return
	}

	feed := jsonFeed{
		Version:     "https://jsonfeed.org/version/1.1",
		Title:       ui.SiteName,
		HomePageURL: s.cfg.BaseURL + "/",
		FeedURL:     s.cfg.BaseURL + "/feed.json",
		Description: ui.SiteTagline,
		Language:    "en",
		Authors:     []jsonFeedAuthor{{Name: ui.SiteName}},
	}

	for _, post := range posts {
		url := s.cfg.BaseURL + post.PermalinkPath()
		item := jsonFeedItem{
			ID:            url,
			URL:           url,
			ContentHTML:   feedContentHTML(post.BodyHTML),
			DatePublished: post.Date().UTC().Format(time.RFC3339),
			DateModified:  post.UpdatedAt.UTC().Format(time.RFC3339),
			Tags:          post.TagList(),
		}
		if post.Title != "" {
			item.Title = post.Title
		}
		if post.Summary != "" {
			item.Summary = post.Summary
		}
		feed.Items = append(feed.Items, item)
	}

	w.Header().Set("Content-Type", "application/feed+json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(feed); err != nil {
		s.log.Error("encode json feed", "err", err)
	}
}
