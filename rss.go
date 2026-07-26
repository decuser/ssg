// rss.go
package main

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type rssDocument struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Content string     `xml:"xmlns:content,attr"`
	Atom    string     `xml:"xmlns:atom,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title         string      `xml:"title"`
	Link          string      `xml:"link"`
	Description   string      `xml:"description"`
	Language      string      `xml:"language,omitempty"`
	LastBuildDate string      `xml:"lastBuildDate,omitempty"`
	AtomLink      rssAtomLink `xml:"atom:link"`
	Items         []rssItem   `xml:"item"`
}

type rssAtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type rssItem struct {
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	GUID        rssGUID  `xml:"guid"`
	PubDate     string   `xml:"pubDate"`
	Description string   `xml:"description"`
	Content     string   `xml:"content:encoded"`
	Categories  []string `xml:"category,omitempty"`
}

type rssGUID struct {
	IsPermaLink string `xml:"isPermaLink,attr"`
	Value       string `xml:",chardata"`
}

func renderRSS(site *Site, opts Options) error {
	base, err := normalizeBaseURL(opts.BaseURL)
	if err != nil {
		return err
	}

	feedURL := joinURL(base, "/feed.xml")
	items := make([]rssItem, 0, len(site.Posts))
	for _, post := range site.Posts {
		postURL := joinURL(base, "/posts/"+post.Slug+"/")
		items = append(items, rssItem{
			Title:       post.Title,
			Link:        postURL,
			GUID:        rssGUID{IsPermaLink: "true", Value: postURL},
			PubDate:     post.Date.Format(time.RFC1123Z),
			Description: prepareFeedHTML(post.Summary, base),
			Content:     prepareFeedHTML(post.Content, base),
			Categories:  post.Tags,
		})
	}

	lastBuildDate := ""
	if len(site.Posts) > 0 {
		lastBuildDate = site.Posts[0].Date.Format(time.RFC1123Z)
	}

	doc := rssDocument{
		Version: "2.0",
		Content: "http://purl.org/rss/1.0/modules/content/",
		Atom:    "http://www.w3.org/2005/Atom",
		Channel: rssChannel{
			Title:         opts.SiteTitle,
			Link:          base,
			Description:   opts.SiteDescription,
			Language:      opts.Language,
			LastBuildDate: lastBuildDate,
			AtomLink: rssAtomLink{
				Href: feedURL,
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Items: items,
		},
	}

	data, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	data = append([]byte(xml.Header), data...)
	data = append(data, '\n')

	return os.WriteFile(filepath.Join(opts.Output, "feed.xml"), data, 0644)
}

func normalizeBaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("-url is required to generate feed.xml")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid -url %q: use an absolute URL such as https://example.com", raw)
	}
	u.Path = strings.TrimRight(u.Path, "/")
	u.RawQuery = ""
	u.Fragment = ""
	return strings.TrimRight(u.String(), "/"), nil
}

func joinURL(base, path string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")
}

// Convert root-relative links in rendered post HTML to absolute links for feed readers.
func absolutizeHTML(s, base string) string {
	replacer := strings.NewReplacer(
		`href="/`, `href="`+strings.TrimRight(base, "/")+`/`,
		`src="/`, `src="`+strings.TrimRight(base, "/")+`/`,
		`href='/`, `href='`+strings.TrimRight(base, "/")+`/`,
		`src='/`, `src='`+strings.TrimRight(base, "/")+`/`,
	)
	return replacer.Replace(s)
}

var (
	feedBlockElements = regexp.MustCompile(`(?is)<(?:iframe|script|style|canvas)\b[^>]*>.*?</(?:iframe|script|style|canvas)\s*>`)
	feedEmptyElements = regexp.MustCompile(`(?is)<(?:iframe|embed|source|track|canvas)\b[^>]*?/?>`)
	feedWrapperTags   = regexp.MustCompile(`(?is)</?(?:object|video|audio|form)\b[^>]*>`)
	feedFormControls  = regexp.MustCompile(`(?is)<(?:input|button|select|textarea)\b[^>]*>.*?</(?:button|select|textarea)\s*>|<(?:input)\b[^>]*?/?>`)
)

// prepareFeedHTML removes browser-only embedded content that feed readers commonly
// render as blank space, while preserving ordinary fallback text and links.
func prepareFeedHTML(s, base string) string {
	s = feedBlockElements.ReplaceAllString(s, "")
	s = feedEmptyElements.ReplaceAllString(s, "")
	s = feedFormControls.ReplaceAllString(s, "")
	s = feedWrapperTags.ReplaceAllString(s, "")
	return absolutizeHTML(s, base)
}
